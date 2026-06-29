// Package uptick implements the native pm Uptick connector. Uptick is field
// service management software (https://developer.uptick.com/). This is a
// declarative-HTTP per-system connector built on the connsdk toolkit, modeled on
// the stripe reference connector: a thin package that composes a Requester with
// Uptick-specific OAuth2 password-grant auth, stream definitions, endpoints, and
// links.next cursor pagination.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
//
// The connector is read-only: the Uptick upstream source supports only
// full_refresh / incremental reads, so Capabilities.Write is false.
package uptick

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	// uptickAPIVersion is the Uptick REST API version segment used by the
	// upstream upstream manifest.
	uptickAPIVersion = "v2.14"
	// uptickTokenPath is the OAuth2 password-grant token endpoint, relative to
	// base_url.
	uptickTokenPath       = "/api/oauth2/token/"
	uptickDefaultPageSize = 100
	uptickMaxPageSize     = 200
	uptickUserAgent       = "polymetrics-go-cli"
	// uptickMaxPagesGuard bounds the links.next pagination loop so a misbehaving
	// or cyclic API cannot spin forever when max_pages is unset.
	uptickMaxPagesGuard = 10000
	// uptickFixtureUpdated is the deterministic `updated` timestamp used by
	// fixture-mode records.
	uptickFixtureUpdated = "2026-01-01T00:00:00.000000Z"
)

func init() {
	connectors.RegisterFactory("uptick", New)
}

// New returns the Uptick connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Uptick connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "uptick" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "uptick",
		DisplayName:     "Uptick",
		IntegrationType: "api",
		Description:     "Reads Uptick field service management data (tasks, clients, properties, invoices, assets) through the Uptick REST API using OAuth2 password-grant auth.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Uptick. In
// fixture mode it short-circuits without a network call. Otherwise it validates
// config and performs a bounded read of the clients stream to confirm the OAuth
// token exchange and connectivity.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := uptickBaseURL(cfg); err != nil {
		return err
	}
	if err := requireCreds(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	query := url.Values{}
	query.Set("page_size", "1")
	endpoint := uptickStreamEndpoints["clients"]
	if endpoint.fields != "" {
		query.Set("fields["+endpoint.sparseType+"]", endpoint.fields)
	}
	if _, err := r.Do(ctx, http.MethodGet, resourcePath(endpoint.resource), query, nil); err != nil {
		return fmt.Errorf("check uptick: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: uptickStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Uptick is read-only for pm
// (the upstream source supports only reads), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: an Uptick stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "clients"
	}
	endpoint, ok := uptickStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("uptick stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	if _, err := uptickBaseURL(req.Config); err != nil {
		return err
	}
	if err := requireCreds(req.Config); err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := uptickPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := uptickMaxPages(req.Config)
	if err != nil {
		return err
	}
	updatedSince := incrementalLowerBound(req)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, updatedSince, emit)
}

// harvest drives Uptick's links.next cursor pagination. List responses are shaped
// {data:[...], links:{next:"<absolute url>"|null}}; the next page is fetched by
// requesting that absolute URL. The Requester treats an http(s) path as absolute,
// so the loop simply feeds links.next back in until it is empty.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, updatedSince string, emit func(connectors.Record) error) error {
	path := resourcePath(endpoint.resource)
	query := url.Values{}
	query.Set("ordering", "-updated")
	query.Set("page_size", strconv.Itoa(pageSize))
	if endpoint.fields != "" {
		query.Set("fields["+endpoint.sparseType+"]", endpoint.fields)
	}
	if updatedSince != "" {
		query.Set("updatedsince", updatedSince)
	}

	guard := maxPages
	if guard <= 0 || guard > uptickMaxPagesGuard {
		guard = uptickMaxPagesGuard
	}

	for page := 0; page < guard; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read uptick %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode uptick %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "links.next")
		if err != nil {
			return fmt.Errorf("decode uptick %s links.next: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		// links.next is an absolute URL carrying its own query (page cursor +
		// the filters we set). Switch the request to it and drop the seed query.
		path = next
		query = nil
		if maxPages > 0 && page+1 >= maxPages {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise uptick credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               int64(i),
			"created":          uptickFixtureUpdated,
			"updated":          uptickFixtureUpdated,
			"deleted":          nil,
			"is_active":        true,
			"ref":              fmt.Sprintf("%s-%d", endpoint.resource, i),
			"name":             fmt.Sprintf("Fixture %s %d", endpoint.resource, i),
			"description":      fmt.Sprintf("fixture %s record %d", endpoint.resource, i),
			"status":           "active",
			"address":          fmt.Sprintf("%d Example St", i),
			"contact_name":     fmt.Sprintf("Fixture Contact %d", i),
			"contact_email":    fmt.Sprintf("fixture+%d@example.com", i),
			"contact_phone_bh": "+61000000000",
			"notes":            "fixture",
			"due":              uptickFixtureUpdated,
			"priority":         "normal",
			"client":           "1",
			"property":         "1",
			"timezone":         "Australia/Sydney",
			"coords":           "0,0",
			"number":           fmt.Sprintf("INV-%d", i),
			"currency":         "AUD",
			"date":             "2026-01-01",
			"due_date":         "2026-01-31",
			"subtotal":         "100.00",
			"gst":              "10.00",
			"total":            "110.00",
			"is_overdue":       false,
			"is_sent":          true,
			"task":             "1",
			"uptick_ref":       fmt.Sprintf("U-%d", i),
			"label":            fmt.Sprintf("Asset %d", i),
			"location":         "Roof",
			"make":             "Acme",
			"model":            "X1",
			"size":             "Large",
			"barcode":          fmt.Sprintf("BC-%d", i),
			"serviced_date":    "2026-01-01",
			"type":             "1",
			"variant":          "1",
		}
		record := endpoint.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the OAuth2 password-grant
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := uptickBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth := &passwordGrantAuth{
		tokenURL:     base + uptickTokenPath,
		clientID:     secret(cfg, "client_id"),
		clientSecret: secret(cfg, "client_secret"),
		username:     strings.TrimSpace(cfg.Config["username"]),
		password:     secret(cfg, "password"),
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: uptickUserAgent,
	}, nil
}

// passwordGrantAuth implements the OAuth2 Resource Owner Password Credentials
// grant Uptick requires (client_id/client_secret + username/password). It fetches
// and caches the access token, refreshing before expiry. connsdk's
// OAuth2ClientCredentials only covers the client-credentials grant, so this small
// in-package authenticator carries the username/password fields. It never logs
// secret values.
type passwordGrantAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	username     string
	password     string
	client       *http.Client

	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *passwordGrantAuth) clock() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *passwordGrantAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *passwordGrantAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && a.clock().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("uptick oauth: token URL is required")
	}

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("username", a.username)
	form.Set("password", a.password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("uptick oauth: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("uptick oauth: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("uptick oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("uptick oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("uptick oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.clock().Add(ttl)
	return a.token, nil
}

// incrementalLowerBound returns the updatedsince lower bound, derived from the
// incremental cursor (if any) or else the start_date config. Empty means no lower
// bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

// resourcePath builds the resource path relative to base_url, e.g.
// "api/v2.14/clients/".
func resourcePath(resource string) string {
	return "api/" + uptickAPIVersion + "/" + resource + "/"
}

func requireCreds(cfg connectors.RuntimeConfig) error {
	for _, field := range []string{"client_id", "client_secret", "password"} {
		if strings.TrimSpace(secret(cfg, field)) == "" {
			return fmt.Errorf("uptick connector requires secret %s", field)
		}
	}
	if strings.TrimSpace(cfg.Config["username"]) == "" {
		return errors.New("uptick connector requires config username")
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// uptickBaseURL resolves and validates the base URL. base_url is required (an
// Uptick instance is per-tenant, e.g. https://demo-fire.onuptick.com). Any value
// must be an absolute https (or http for local test servers) URL with a host, to
// bound SSRF risk.
func uptickBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return "", errors.New("uptick config base_url is required (e.g. https://demo-fire.onuptick.com)")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("uptick config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("uptick config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("uptick config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func uptickPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return uptickDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("uptick config page_size must be an integer: %w", err)
	}
	if value < 1 || value > uptickMaxPageSize {
		return 0, fmt.Errorf("uptick config page_size must be between 1 and %d", uptickMaxPageSize)
	}
	return value, nil
}

func uptickMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("uptick config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("uptick config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
