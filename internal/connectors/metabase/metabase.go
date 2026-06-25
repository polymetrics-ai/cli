// Package metabase implements the native pm Metabase connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit
// (Requester + X-Metabase-Session header auth + array/data extraction),
// composed with Metabase-specific stream definitions and endpoints. It copies
// the shape of the stripe reference connector.
//
// Like the other native connectors it self-registers with the connectors
// registry via RegisterFactory in init(); the registryset package blank-imports
// this package in the production binary to run that side effect.
//
// Metabase's REST API is read-only for the resources we expose and authenticates
// with a session token. Sessions are obtained by POSTing username/password to
// /session and are passed on subsequent requests in the X-Metabase-Session
// header. The connector accepts a session_token secret directly, or derives one
// from a username (config) + password (secret).
package metabase

import (
	"context"
	"encoding/json"
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
	metabaseDefaultPageSize = 50
	metabaseMaxPageSize     = 1000
	metabaseUserAgent       = "polymetrics-go-cli"
	metabaseSessionHeader   = "X-Metabase-Session"
)

func init() {
	connectors.RegisterFactory("metabase", New)
}

// New returns the Metabase connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Metabase connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "metabase" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "metabase",
		DisplayName:     "Metabase",
		IntegrationType: "api",
		Description:     "Reads Metabase cards, dashboards, collections, databases, and users through the Metabase REST API using session-token authentication.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Metabase. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := metabaseBaseURL(cfg); err != nil {
		return err
	}
	r, err := c.requester(ctx, cfg)
	if err != nil {
		return err
	}
	// A bounded read of the current user confirms the session is valid without
	// mutating anything. /user/current returns the authenticated user object.
	if err := r.DoJSON(ctx, http.MethodGet, "user/current", nil, nil, nil); err != nil {
		return fmt.Errorf("check metabase: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: metabaseStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "cards"
	}
	endpoint, ok := metabaseStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("metabase stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(ctx, req.Config)
	if err != nil {
		return err
	}
	pageSize, err := metabasePageSize(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, emit)
}

// listEnvelope models the two shapes Metabase list endpoints return: a bare
// top-level array, or an object wrapping the rows in "data" with paging totals.
// json.RawMessage lets us peek at the leading byte to decide which shape it is.
type listEnvelope struct {
	Data   []map[string]any `json:"data"`
	Total  *int             `json:"total"`
	Limit  *int             `json:"limit"`
	Offset *int             `json:"offset"`
}

// harvest reads a Metabase list endpoint. Most endpoints return a bare JSON
// array with no pagination; some (e.g. /api/user) return
// {"data":[...],"total":N,"limit":L,"offset":O} and must be paged by offset
// until every row is read. harvest handles both shapes from one loop.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize int, emit func(connectors.Record) error) error {
	offset := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
		if err != nil {
			return fmt.Errorf("read metabase %s: %w", endpoint.resource, err)
		}

		rows, total, wrapped, err := decodeList(resp.Body, endpoint.resource)
		if err != nil {
			return err
		}
		for _, item := range rows {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		// Bare-array endpoints return everything in one response.
		if !wrapped {
			return nil
		}
		// Data-wrapped endpoints page by offset until total rows are consumed,
		// or until a short/empty page signals the end.
		offset += len(rows)
		if len(rows) == 0 || (total >= 0 && offset >= total) {
			return nil
		}
		// Re-issue with explicit limit/offset for the next page.
		next := url.Values{}
		next.Set("limit", strconv.Itoa(pageSize))
		next.Set("offset", strconv.Itoa(offset))
		resp, err = r.Do(ctx, http.MethodGet, endpoint.resource, next, nil)
		if err != nil {
			return fmt.Errorf("read metabase %s: %w", endpoint.resource, err)
		}
		rows, total, _, err = decodeList(resp.Body, endpoint.resource)
		if err != nil {
			return err
		}
		for _, item := range rows {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		offset += len(rows)
		if len(rows) == 0 || (total >= 0 && offset >= total) {
			return nil
		}
	}
}

// decodeList parses a Metabase list response, returning its rows, the reported
// total (-1 when absent), and whether the payload was the data-wrapped shape.
func decodeList(body []byte, resource string) (rows []map[string]any, total int, wrapped bool, err error) {
	trimmed := strings.TrimSpace(string(body))
	total = -1
	if strings.HasPrefix(trimmed, "[") {
		var arr []map[string]any
		if derr := json.Unmarshal(body, &arr); derr != nil {
			return nil, -1, false, fmt.Errorf("decode metabase %s array: %w", resource, derr)
		}
		return arr, -1, false, nil
	}
	var env listEnvelope
	if derr := json.Unmarshal(body, &env); derr != nil {
		return nil, -1, false, fmt.Errorf("decode metabase %s envelope: %w", resource, derr)
	}
	if env.Total != nil {
		total = *env.Total
	}
	return env.Data, total, true, nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise metabase credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                i,
			"name":              fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"description":       "fixture record",
			"slug":              fmt.Sprintf("%s-fixture-%d", endpoint.resource, i),
			"collection_id":     1,
			"database_id":       1,
			"creator_id":        1,
			"personal_owner_id": nil,
			"query_type":        "query",
			"display":           "table",
			"engine":            "postgres",
			"location":          "/",
			"timezone":          "UTC",
			"archived":          false,
			"is_sample":         false,
			"is_on_demand":      false,
			"is_active":         true,
			"is_superuser":      false,
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"first_name":        fmt.Sprintf("Fixture%d", i),
			"last_name":         "User",
			"common_name":       fmt.Sprintf("Fixture%d User", i),
			"last_login":        "2026-01-01T00:00:00Z",
			"date_joined":       "2026-01-01T00:00:00Z",
			"created_at":        "2026-01-01T00:00:00Z",
			"updated_at":        "2026-01-01T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the resolved base URL and an
// X-Metabase-Session header carrying a session token. The token is obtained from
// the session_token secret directly, or minted from username+password. The
// secret only ever flows into headers; it is never logged.
func (c Connector) requester(ctx context.Context, cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := metabaseBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token, err := c.sessionToken(ctx, cfg, base)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(metabaseSessionHeader, token, ""),
		UserAgent: metabaseUserAgent,
	}, nil
}

// sessionToken resolves the session token: prefer the session_token secret; if
// absent, mint one by POSTing username (config) + password (secret) to /session.
func (c Connector) sessionToken(ctx context.Context, cfg connectors.RuntimeConfig, base string) (string, error) {
	if token := strings.TrimSpace(secret(cfg, "session_token")); token != "" {
		return token, nil
	}
	username := strings.TrimSpace(cfg.Config["username"])
	password := strings.TrimSpace(secret(cfg, "password"))
	if username == "" || password == "" {
		return "", errors.New("metabase connector requires a session_token secret, or a username config plus password secret")
	}
	// Mint a session. /session returns {"id":"<token>"}.
	loginReq := &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: metabaseUserAgent,
	}
	var out struct {
		ID string `json:"id"`
	}
	body := map[string]string{"username": username, "password": password}
	if err := loginReq.DoJSON(ctx, http.MethodPost, "session", nil, body, &out); err != nil {
		return "", fmt.Errorf("metabase session login: %w", err)
	}
	if strings.TrimSpace(out.ID) == "" {
		return "", errors.New("metabase session login returned an empty token")
	}
	return out.ID, nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// metabaseBaseURL resolves and validates the base URL from instance_api_url (or
// the base_url alias). Metabase requires an absolute https URL; http is also
// permitted so local test servers (httptest) work. The host must be present to
// bound SSRF risk. A trailing /api path segment is preserved as-is.
func metabaseBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["instance_api_url"])
	if base == "" {
		base = strings.TrimSpace(cfg.Config["base_url"])
	}
	if base == "" {
		return "", errors.New("metabase connector requires config instance_api_url")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("metabase config instance_api_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("metabase config instance_api_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("metabase config instance_api_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func metabasePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return metabaseDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("metabase config page_size must be an integer: %w", err)
	}
	if value < 1 || value > metabaseMaxPageSize {
		return 0, fmt.Errorf("metabase config page_size must be between 1 and %d", metabaseMaxPageSize)
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Metabase is exposed read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
