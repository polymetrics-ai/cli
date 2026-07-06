// Package freeagentconnector implements the native pm FreeAgent connector. It is
// a declarative-HTTP per-system connector built on the connsdk toolkit: a
// refresh-token OAuth2 authenticator (FreeAgent's token endpoint uses HTTP Basic
// auth for the client credentials plus a refresh_token grant) feeding a Bearer
// requester, FreeAgent's page/per_page pagination, and per-stream record mappers.
//
// name "free-agent-connector"; the Go package name drops the hyphens.
package freeagentconnector

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
	registryName       = "free-agent-connector"
	defaultBaseURL     = "https://api.freeagent.com/v2"
	tokenEndpointPath  = "token_endpoint"
	defaultPageSize    = 100
	maxPageSize        = 100
	userAgent          = "polymetrics-go-cli"
	fixtureUpdatedAt   = "2026-01-01T00:00:00Z"
	fixtureCreatedAt   = "2025-12-01T00:00:00Z"
	tokenRefreshWindow = 60 * time.Second
)

// New returns the FreeAgent connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm FreeAgent connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "FreeAgent",
		IntegrationType: "api",
		Description:     "Reads FreeAgent contacts, invoices, bills, projects, and tasks through the FreeAgent v2 REST API using OAuth2 refresh-token authentication.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to FreeAgent. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the contacts list confirms the refresh-token exchange,
	// auth, and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "contacts", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check free-agent: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streamDefs()}, nil
}

// InitialState satisfies connectors.StatefulReader: a FreeAgent stream starts with
// an empty incremental cursor (full sync), which the updated_since config can
// raise at read time.
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
		stream = "contacts"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("free-agent stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := resolvePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := resolveMaxPages(req.Config)
	if err != nil {
		return err
	}
	updatedSince := incrementalLowerBound(req)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, updatedSince, emit)
}

// Write is unsupported: FreeAgent is read-only for this connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives FreeAgent's page/per_page pagination. List responses are shaped
// {"<resource>":[...]}; the loop stops on a short or empty page. The loop lives
// here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, updatedSince string, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))
		if updatedSince != "" {
			query.Set("updated_since", updatedSince)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read free-agent %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode free-agent %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"url":               fmt.Sprintf("https://api.freeagent.com/v2/%s/%d", endpoint.resource, i),
			"first_name":        fmt.Sprintf("Fixture%d", i),
			"last_name":         "Example",
			"organisation_name": fmt.Sprintf("Fixture Org %d", i),
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"phone_number":      "0000000000",
			"status":            "Active",
			"account_balance":   "0.0",
			"reference":         fmt.Sprintf("REF-%d", i),
			"contact":           "https://api.freeagent.com/v2/contacts/1",
			"dated_on":          "2026-01-01",
			"due_on":            "2026-02-01",
			"currency":          "GBP",
			"net_value":         fmt.Sprintf("%d.00", 100*i),
			"total_value":       fmt.Sprintf("%d.00", 120*i),
			"due_value":         fmt.Sprintf("%d.00", 120*i),
			"name":              fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"budget":            "1000",
			"budget_units":      "Hours",
			"project":           "https://api.freeagent.com/v2/projects/1",
			"is_billable":       true,
			"billing_rate":      "50.0",
			"billing_period":    "hour",
			"created_at":        fixtureCreatedAt,
			"updated_at":        fixtureUpdatedAt,
		}
		record := endpoint.mapRecord(item)
		record["connector"] = registryName
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the FreeAgent refresh-token
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}
	auth := &refreshTokenAuth{
		tokenURL:     base + "/" + tokenEndpointPath,
		clientID:     cfg.Secrets["client_id"],
		clientSecret: cfg.Secrets["client_secret"],
		refreshToken: cfg.Secrets["client_refresh_token_2"],
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: userAgent,
	}, nil
}

// refreshTokenAuth implements connsdk.Authenticator for FreeAgent's OAuth2
// refresh-token grant. The token endpoint authenticates the client with HTTP
// Basic auth (client_id:client_secret) and takes grant_type=refresh_token plus
// the refresh_token in the form body. The resulting access token is cached and
// refreshed before expiry; it is applied as Authorization: Bearer.
type refreshTokenAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	client       *http.Client

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *refreshTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *refreshTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && time.Now().Add(tokenRefreshWindow).Before(a.expires) {
		return a.token, nil
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("free-agent: build token request: %w", err)
	}
	httpReq.SetBasicAuth(a.clientID, a.clientSecret)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("free-agent: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("free-agent: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("free-agent: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("free-agent: token response missing access_token")
	}

	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.token = out.AccessToken
	a.expires = time.Now().Add(ttl)
	return a.token, nil
}

// incrementalLowerBound returns the RFC3339 updated_since value derived from the
// incremental cursor (if any) or else the updated_since config. An empty result
// means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["updated_since"])
}

func requireSecrets(cfg connectors.RuntimeConfig) error {
	if cfg.Secrets == nil {
		return errors.New("free-agent connector requires secrets client_id, client_secret, client_refresh_token_2")
	}
	for _, key := range []string{"client_id", "client_secret", "client_refresh_token_2"} {
		if strings.TrimSpace(cfg.Secrets[key]) == "" {
			return fmt.Errorf("free-agent connector requires secret %s", key)
		}
	}
	return nil
}

// baseURL resolves and validates the base URL. The default is api.freeagent.com;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("free-agent config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("free-agent config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("free-agent config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func resolvePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("free-agent config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("free-agent config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func resolveMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("free-agent config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("free-agent config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
