// Package mendeley implements the native pm Mendeley connector. It follows the
// declarative-HTTP per-system connector shape (see internal/connectors/stripe):
// a thin package that composes the connsdk toolkit (Requester + OAuth2
// refresh-token auth + RecordsAt extraction + Link-header pagination) with
// Mendeley-specific stream definitions and endpoints.
//
// Mendeley is read-only here: the public reverse-ETL surface is not exposed, so
// Capabilities.Write is false.
package mendeley

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
	mendeleyDefaultBaseURL  = "https://api.mendeley.com"
	mendeleyDefaultTokenURL = "https://api.mendeley.com/oauth/token"
	mendeleyDefaultPageSize = 100
	mendeleyMaxPageSize     = 500
	mendeleyUserAgent       = "polymetrics-go-cli"
	// mendeleyFixtureModified is the deterministic timestamp used by fixture-mode
	// records.
	mendeleyFixtureModified = "2026-01-01T00:00:00.000Z"
)

// New returns the Mendeley connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mendeley connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mendeley" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mendeley",
		DisplayName:     "Mendeley",
		IntegrationType: "api",
		Description:     "Reads documents, folders, groups, and annotations from the Mendeley reference manager REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mendeley. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mendeleyBaseURL(cfg); err != nil {
		return err
	}
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg, mendeleyStreamEndpoints["documents"].accept)
	if err != nil {
		return err
	}
	// A bounded read of the documents list confirms the refresh-token grant,
	// auth, and connectivity without mutating anything.
	if _, err := r.Do(ctx, http.MethodGet, "documents", url.Values{"limit": []string{"1"}}, nil); err != nil {
		return fmt.Errorf("check mendeley: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mendeleyStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Mendeley stream starts with
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
		stream = "documents"
	}
	endpoint, ok := mendeleyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mendeley stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config, endpoint.accept)
	if err != nil {
		return err
	}
	pageSize, err := mendeleyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mendeleyMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	base.Set("order", "asc")
	if lower := incrementalLowerBound(req); lower != "" && endpointSupportsModifiedSince(stream) {
		base.Set("modified_since", lower)
	}

	// Mendeley list endpoints return a top-level JSON array and paginate via the
	// RFC 5988 Link header rel="next" (marker-based). connsdk's
	// LinkHeaderPaginator + Harvest cover this exactly.
	paginator := &connsdk.LinkHeaderPaginator{}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, "", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mendeley credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"title":             fmt.Sprintf("Fixture %s %d", stream, i),
			"name":              fmt.Sprintf("Fixture %s %d", stream, i),
			"type":              "journal",
			"text":              fmt.Sprintf("fixture annotation %d", i),
			"source":            "Polymetrics Journal",
			"year":              json.Number("2026"),
			"abstract":          "Deterministic fixture record.",
			"description":       "Deterministic fixture record.",
			"profile_id":        "profile_fixture_1",
			"group_id":          "group_fixture_1",
			"document_id":       "documents_fixture_1",
			"parent_id":         "",
			"access_level":      "private",
			"role":              "owner",
			"owning_profile_id": "profile_fixture_1",
			"privacy_level":     "private",
			"filehash":          fmt.Sprintf("hash_%d", i),
			"webpage":           "https://www.mendeley.com",
			"created":           mendeleyFixtureModified,
			"modified":          mendeleyFixtureModified,
			"last_modified":     mendeleyFixtureModified,
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

// requester builds a connsdk.Requester wired with OAuth2 refresh-token auth, the
// resolved base URL, and the resource-specific Accept media type. Secrets only
// ever flow into the authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig, accept string) (*connsdk.Requester, error) {
	base, err := mendeleyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}
	tokenURL, err := mendeleyTokenURL(cfg)
	if err != nil {
		return nil, err
	}
	auth := &refreshTokenAuth{
		TokenURL:     tokenURL,
		ClientID:     secret(cfg, "client_id"),
		ClientSecret: secret(cfg, "client_secret"),
		RefreshToken: secret(cfg, "client_refresh_token"),
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: mendeleyUserAgent,
		Accept:    accept,
	}, nil
}

// refreshTokenAuth implements the OAuth2 refresh-token grant against the Mendeley
// token endpoint, caching the access token and refreshing before expiry. It never
// logs secret values.
type refreshTokenAuth struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	RefreshToken string
	Client       *http.Client
	// Now is injectable for tests; defaults to time.Now.
	Now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *refreshTokenAuth) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
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
	// Refresh 60s before expiry to avoid edge races.
	if a.token != "" && a.now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.TokenURL) == "" {
		return "", errors.New("mendeley: token_url is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.RefreshToken)
	form.Set("client_id", a.ClientID)
	form.Set("client_secret", a.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("mendeley: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("mendeley: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("mendeley: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("mendeley: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("mendeley: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.now().Add(ttl)
	return a.token, nil
}

// incrementalLowerBound returns the RFC3339 lower bound for modified_since,
// derived from the incremental cursor (if any) or else the start_date config. An
// empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

// endpointSupportsModifiedSince reports whether a stream accepts the
// modified_since filter. groups do not.
func endpointSupportsModifiedSince(stream string) bool {
	switch stream {
	case "documents", "annotations", "files":
		return true
	default:
		return false
	}
}

func requireSecrets(cfg connectors.RuntimeConfig) error {
	for _, name := range []string{"client_id", "client_secret", "client_refresh_token"} {
		if strings.TrimSpace(secret(cfg, name)) == "" {
			return fmt.Errorf("mendeley connector requires secret %s", name)
		}
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

// mendeleyBaseURL resolves and validates the base URL. The default is
// api.mendeley.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func mendeleyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedURL(cfg.Config["base_url"], mendeleyDefaultBaseURL, "base_url")
}

// mendeleyTokenURL resolves and validates the OAuth2 token endpoint URL.
func mendeleyTokenURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedURL(cfg.Config["token_url"], mendeleyDefaultTokenURL, "token_url")
}

func validatedURL(raw, fallback, field string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("mendeley config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mendeley config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("mendeley config %s must include a host", field)
	}
	return strings.TrimRight(raw, "/"), nil
}

func mendeleyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mendeleyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mendeley config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mendeleyMaxPageSize {
		return 0, fmt.Errorf("mendeley config page_size must be between 1 and %d", mendeleyMaxPageSize)
	}
	return value, nil
}

func mendeleyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mendeley config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mendeley config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Mendeley is read-only in
// pm, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
