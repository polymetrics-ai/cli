// Package akeneo implements the native pm Akeneo PIM connector. It follows the
// stripe declarative-HTTP template: a thin package composing the connsdk toolkit
// (Requester + a custom OAuth2 password-grant Authenticator + _embedded.items
// extraction + _links.next.href pagination) with Akeneo-specific stream
// definitions and endpoints.
//
// Akeneo is read-only here: the PIM has no obviously-safe reverse-ETL write
// surface, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package akeneo

import (
	"bytes"
	"context"
	"encoding/base64"
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
	akeneoTokenPath       = "api/oauth/v1/token"
	akeneoRESTPrefix      = "api/rest/v1"
	akeneoDefaultPageSize = 100
	akeneoMaxPageSize     = 100
	akeneoUserAgent       = "polymetrics-go-cli"
	// akeneoFixtureUpdated is the deterministic `updated` timestamp used by the
	// fixture-mode records.
	akeneoFixtureUpdated = "2026-01-01T00:00:00+00:00"
)

func init() {
	connectors.RegisterFactory("akeneo", New)
}

// New returns the Akeneo connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Akeneo connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "akeneo" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "akeneo",
		DisplayName:     "Akeneo",
		IntegrationType: "api",
		Description:     "Reads Akeneo PIM products, categories, families, attributes, and channels through the Akeneo REST API (OAuth2 password grant).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Akeneo. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := akeneoBaseURL(cfg); err != nil {
		return err
	}
	if err := requireLiveConfig(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the channels list confirms the token exchange and
	// connectivity without mutating anything.
	if _, err := r.Do(ctx, http.MethodGet, akeneoRESTPrefix+"/channels", url.Values{"limit": []string{"1"}}, nil); err != nil {
		return fmt.Errorf("check akeneo: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: akeneoStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "products"
	}
	endpoint, ok := akeneoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("akeneo stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := akeneoPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := akeneoMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Akeneo's HAL pagination. List responses are
// {_embedded:{items:[...]}, _links:{next:{href:"<absolute url>"}}}; the next page
// is requested by following _links.next.href verbatim. There is no link-in-body
// paginator in connsdk for this exact shape, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// The first request is relative; subsequent requests follow the absolute
	// next href, which the connsdk Requester treats as-is.
	path := akeneoRESTPrefix + "/" + endpoint.resource
	query := url.Values{}
	query.Set("limit", strconv.Itoa(pageSize))

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read akeneo %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "_embedded.items")
		if err != nil {
			return fmt.Errorf("decode akeneo %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "_links.next.href")
		if err != nil {
			return fmt.Errorf("decode akeneo %s next link: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		// Follow the absolute next href verbatim; it already carries the page and
		// limit, so the relative query is no longer needed.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise akeneo credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		code := fmt.Sprintf("%s_fixture_%d", endpoint.resource, i)
		item := map[string]any{
			"code":               code,
			"identifier":         code,
			"uuid":               fmt.Sprintf("00000000-0000-0000-0000-00000000000%d", i),
			"enabled":            true,
			"family":             "fixture_family",
			"parent":             nil,
			"categories":         []any{"fixture_category"},
			"groups":             []any{},
			"values":             map[string]any{},
			"created":            akeneoFixtureUpdated,
			"updated":            akeneoFixtureUpdated,
			"labels":             map[string]any{"en_US": fmt.Sprintf("Fixture %d", i)},
			"attribute_as_label": "name",
			"attribute_as_image": nil,
			"attributes":         []any{"name", "description"},
			"type":               "pim_catalog_text",
			"group":              "marketing",
			"localizable":        false,
			"scopable":           false,
			"currencies":         []any{"USD"},
			"locales":            []any{"en_US"},
			"category_tree":      "master",
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

// requester builds a connsdk.Requester wired with the Akeneo OAuth2 password-grant
// authenticator and the resolved base URL. Secrets only ever flow into the token
// exchange; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := akeneoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireLiveConfig(cfg); err != nil {
		return nil, err
	}
	auth := &passwordGrantAuth{
		tokenURL:     base + "/" + akeneoTokenPath,
		clientID:     strings.TrimSpace(cfg.Config["client_id"]),
		clientSecret: strings.TrimSpace(akeneoSecret(cfg)),
		username:     strings.TrimSpace(cfg.Config["api_username"]),
		password:     akeneoPassword(cfg),
		client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: akeneoUserAgent,
	}, nil
}

// passwordGrantAuth implements connsdk.Authenticator for Akeneo's OAuth2
// password grant. It POSTs a JSON body to the token endpoint with the
// client_id:secret pair as Basic auth, caches the access token, and refreshes it
// before expiry. It never logs secret values.
type passwordGrantAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	username     string
	password     string
	client       *http.Client

	mu      sync.Mutex
	token   string
	expires time.Time
	now     func() time.Time
}

func (a *passwordGrantAuth) timeNow() time.Time {
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
	// Refresh 60s before expiry to avoid edge races.
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}

	bodyBytes, err := json.Marshal(map[string]string{
		"grant_type": "password",
		"username":   a.username,
		"password":   a.password,
	})
	if err != nil {
		return "", fmt.Errorf("akeneo token: encode body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("akeneo token: build request: %w", err)
	}
	creds := base64.StdEncoding.EncodeToString([]byte(a.clientID + ":" + a.clientSecret))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("akeneo token: request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("akeneo token: endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("akeneo token: decode response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("akeneo token: response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}

// requireLiveConfig validates the non-secret and secret fields needed for a live
// token exchange are present. It never logs the values.
func requireLiveConfig(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(cfg.Config["client_id"]) == "" {
		return errors.New("akeneo connector requires config client_id")
	}
	if strings.TrimSpace(cfg.Config["api_username"]) == "" {
		return errors.New("akeneo connector requires config api_username")
	}
	if strings.TrimSpace(akeneoSecret(cfg)) == "" {
		return errors.New("akeneo connector requires secret secret (OAuth client secret)")
	}
	if strings.TrimSpace(akeneoPassword(cfg)) == "" {
		return errors.New("akeneo connector requires secret password (API user password)")
	}
	return nil
}

func akeneoSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["secret"]
}

func akeneoPassword(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// akeneoBaseURL resolves and validates the Akeneo host URL. The host is the
// connection-specific Akeneo cloud URL (e.g.
// https://example.trial.akeneo.cloud); a base_url override is also honored for
// tests/proxies. Any value must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func akeneoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = strings.TrimSpace(cfg.Config["host"])
	}
	if base == "" {
		return "", errors.New("akeneo connector requires config host")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("akeneo config host is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("akeneo config host must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("akeneo config host must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func akeneoPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return akeneoDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("akeneo config page_size must be an integer: %w", err)
	}
	if value < 1 || value > akeneoMaxPageSize {
		return 0, fmt.Errorf("akeneo config page_size must be between 1 and %d", akeneoMaxPageSize)
	}
	return value, nil
}

func akeneoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("akeneo config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("akeneo config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Akeneo is read-only in this
// connector, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
