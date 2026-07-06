// Package basecamp implements the native pm Basecamp connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package composing the connsdk toolkit (Requester + OAuth2
// refresh auth + RecordsAt extraction + Link-header pagination + cursor state)
// with Basecamp 3 stream definitions, endpoints, and mappers.
//
// Basecamp 3 API specifics:
//   - Base URL: https://3.basecampapi.com/<account_id>
//   - Auth: Authorization: Bearer <access_token>, where the access token is
//     obtained by exchanging the refresh token at the Launchpad token endpoint
//     (https://launchpad.37signals.com/authorization/token) using type=refresh.
//   - List responses are top-level JSON arrays ([...]).
//   - Pagination: RFC5988 Link header with rel="next" carrying an absolute URL
//     to the next page; a missing rel="next" signals the final page.
//   - A descriptive User-Agent (with contact) is mandatory.
package basecamp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	basecampDefaultAPIRoot  = "https://3.basecampapi.com"
	basecampDefaultTokenURL = "https://launchpad.37signals.com/authorization/token"
	basecampUserAgent       = "polymetrics-go-cli (https://polymetrics.dev)"
	// basecampFixtureTimestamp is the deterministic timestamp used by fixture
	// records (2026-01-01T00:00:00Z).
	basecampFixtureTimestamp = "2026-01-01T00:00:00Z"
)

// New returns the Basecamp connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Basecamp connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "basecamp" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "basecamp",
		DisplayName:     "Basecamp",
		IntegrationType: "api",
		Description:     "Reads Basecamp 3 projects, people, and account activity events through the Basecamp REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Basecamp. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := basecampBaseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the projects list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "projects.json", nil, nil, nil); err != nil {
		return fmt.Errorf("check basecamp: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: basecampStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Basecamp stream starts with
// an empty incremental cursor (full sync). Basecamp's top-level list endpoints do
// not support server-side updated_at filtering, so the cursor is advisory.
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
		stream = "projects"
	}
	endpoint, ok := basecampStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("basecamp stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := basecampMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, maxPages, emit)
}

// harvest drives Basecamp's RFC5988 Link-header pagination. Each list response is
// a top-level JSON array; the next page (if any) is the absolute URL carried in
// the Link header's rel="next". connsdk.LinkHeaderPaginator parses that header,
// and RecordsAt with the root path ("") yields the array elements. The per-stream
// mapper is applied to each, so the loop lives here rather than using the generic
// connsdk.Harvest (which has no per-record hook).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	paginator := &connsdk.LinkHeaderPaginator{}
	page := paginator.Start()
	for pageNum := 0; page != nil; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && pageNum >= maxPages {
			return nil
		}

		reqPath := endpoint.resource
		if page.URL != "" {
			reqPath = page.URL
		}
		resp, err := r.Do(ctx, http.MethodGet, reqPath, page.Query, nil)
		if err != nil {
			return fmt.Errorf("read basecamp %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode basecamp %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		page = paginator.Next(resp, len(records))
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise basecamp credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              i,
			"recording_id":    1000 + i,
			"status":          "active",
			"name":            fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"description":     "Fixture record",
			"purpose":         "topic",
			"bookmark_url":    fmt.Sprintf("https://3.basecampapi.com/9999999/my/bookmarks/%d.json", i),
			"url":             fmt.Sprintf("https://3.basecampapi.com/9999999/projects/%d.json", i),
			"app_url":         fmt.Sprintf("https://3.basecamp.com/9999999/projects/%d", i),
			"email_address":   fmt.Sprintf("fixture+%d@example.com", i),
			"title":           "Engineer",
			"admin":           i == 1,
			"owner":           i == 1,
			"client":          false,
			"personable_type": "User",
			"time_zone":       "America/Chicago",
			"action":          "created",
			"kind":            "todo_created",
			"summary":         fmt.Sprintf("Fixture event %d", i),
			"created_at":      basecampFixtureTimestamp,
			"updated_at":      basecampFixtureTimestamp,
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "basecamp"
		record["fixture"] = true
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with OAuth refresh auth and the
// resolved base URL. Secrets only ever flow into the authenticator; they are
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := basecampBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	auth, err := c.authenticator(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: basecampUserAgent,
	}, nil
}

// authenticator builds the OAuth2 refresh-token authenticator from the configured
// secrets. A standalone access_token secret (no refresh) is also accepted as a
// Bearer token for convenience.
func (c Connector) authenticator(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	clientID := basecampSecret(cfg, "client_id")
	clientSecret := basecampSecret(cfg, "client_secret")
	refreshToken := basecampSecret(cfg, "client_refresh_token_2")
	if strings.TrimSpace(refreshToken) == "" {
		// Allow the alternate flat name some configs use.
		refreshToken = basecampSecret(cfg, "refresh_token")
	}

	if strings.TrimSpace(refreshToken) != "" && strings.TrimSpace(clientID) != "" && strings.TrimSpace(clientSecret) != "" {
		return &oauthRefreshAuth{
			tokenURL:     basecampTokenURL(cfg),
			clientID:     clientID,
			clientSecret: clientSecret,
			refreshToken: refreshToken,
			client:       c.Client,
		}, nil
	}
	if accessToken := basecampSecret(cfg, "access_token"); strings.TrimSpace(accessToken) != "" {
		return connsdk.Bearer(accessToken), nil
	}
	return nil, errors.New("basecamp connector requires secrets client_id, client_secret, and client_refresh_token_2 (or an access_token)")
}

// basecampSecret resolves a secret. It accepts both the flat key (e.g.
// "client_id") and the dotted catalog form (e.g. "credentials.client_id").
func basecampSecret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v, ok := cfg.Secrets[key]; ok {
		return v
	}
	return cfg.Secrets["credentials."+key]
}

// basecampBaseURL resolves and validates the base URL, which must include the
// account id path segment. A base_url override is used verbatim (after
// validation); otherwise the default API root is joined with account_id. Any URL
// must be an absolute https (or http for local test servers) URL with a host to
// bound SSRF risk.
func basecampBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	accountID := strings.TrimSpace(cfg.Config["account_id"])
	if base != "" {
		if err := validateURL(base, "base_url"); err != nil {
			return "", err
		}
		base = strings.TrimRight(base, "/")
		// If the override does not already carry the account segment and an
		// account_id is configured, append it so resource paths resolve.
		if accountID != "" && !strings.HasSuffix(base, "/"+accountID) {
			base = base + "/" + accountID
		}
		return base, nil
	}
	if accountID == "" {
		return "", errors.New("basecamp connector requires config account_id")
	}
	return basecampDefaultAPIRoot + "/" + accountID, nil
}

// basecampTokenURL resolves the OAuth token endpoint, validating any override.
func basecampTokenURL(cfg connectors.RuntimeConfig) string {
	tokenURL := strings.TrimSpace(cfg.Config["token_url"])
	if tokenURL == "" {
		return basecampDefaultTokenURL
	}
	if err := validateURL(tokenURL, "token_url"); err != nil {
		// Fall back to the default rather than returning a partial URL; the token
		// request will surface a clear error if the default is wrong.
		return basecampDefaultTokenURL
	}
	return tokenURL
}

func validateURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("basecamp config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("basecamp config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("basecamp config %s must include a host", field)
	}
	return nil
}

func basecampMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("basecamp config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("basecamp config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write is unsupported: Basecamp is read-only in this connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
