// Package monday implements the native pm monday.com source connector. It follows
// the declarative-HTTP shape of the stripe reference connector, but talks to a
// GraphQL API rather than REST: a single POST endpoint (https://api.monday.com/v2)
// whose body carries a GraphQL query and whose response wraps records under a
// top-level "data" object.
//
// It composes the connsdk toolkit (Requester + a raw-token Authorization
// authenticator + RecordsAt extraction) with monday-specific stream definitions,
// GraphQL field selections, and pagination. boards/users/teams/tags use simple
// page-number pagination; items use monday's cursor-based next_items_page model.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect. monday is read-only (no reverse-ETL writes), so
// Capabilities.Write is false.
package monday

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
	mondayDefaultBaseURL  = "https://api.monday.com/v2"
	mondayDefaultPageSize = 50
	mondayMaxPageSize     = 500
	mondayUserAgent       = "polymetrics-go-cli"
	// mondayFixtureUpdated is the deterministic updated_at used by fixture-mode
	// records (an RFC3339 timestamp).
	mondayFixtureUpdated = "2026-01-01T00:00:00Z"
	// mondaySafetyMaxPages bounds an otherwise-unbounded paginated read so a
	// runaway loop cannot spin forever when max_pages is left unset.
	mondaySafetyMaxPages = 10000
)

func init() {
	connectors.RegisterFactory("monday", New)
}

// New returns the monday.com connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm monday.com source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "monday" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "monday",
		DisplayName:     "Monday",
		IntegrationType: "api",
		Description:     "Reads monday.com boards, items, users, teams, and tags through the monday.com GraphQL API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to monday.com.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mondayBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mondaySecret(cfg)) == "" {
		return errors.New("monday connector requires an api_token or access_token secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded `me` query confirms auth and connectivity without reading data.
	// execute already surfaces any GraphQL/HTTP error (including a bad token),
	// so a successful return is sufficient to confirm the credentials work.
	if _, err := c.execute(ctx, r, `query { me { id } }`); err != nil {
		return fmt.Errorf("check monday: %w", err)
	}
	return nil
}

// Write is unsupported: monday is a read-only source connector. It satisfies the
// connectors.Connector interface but always reports the operation as unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mondayStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "boards"
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mondayPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mondayMaxPages(req.Config)
	if err != nil {
		return err
	}

	if stream == "items" {
		return c.readItems(ctx, r, pageSize, maxPages, emit)
	}
	spec, ok := pageStreamSpecs[stream]
	if !ok {
		return fmt.Errorf("monday stream %q not found", stream)
	}
	return c.readPaged(ctx, r, spec, pageSize, maxPages, emit)
}

// readPaged drives monday's page-number pagination for boards/users/teams/tags.
// monday returns a plain array under data.<root>; a short page (fewer records
// than the page size) signals the end.
func (c Connector) readPaged(ctx context.Context, r *connsdk.Requester, spec streamSpec, pageSize, maxPages int, emit func(connectors.Record) error) error {
	limit := maxPages
	if limit <= 0 {
		limit = mondaySafetyMaxPages
	}
	for page := 1; page <= limit; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := fmt.Sprintf("query { %s (limit: %d, page: %d) { %s } }", spec.root, pageSize, page, spec.selection)
		body, err := c.execute(ctx, r, query)
		if err != nil {
			return fmt.Errorf("read monday %s page %d: %w", spec.root, page, err)
		}
		records, err := connsdk.RecordsAt(body, "data."+spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode monday %s page %d: %w", spec.root, page, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(spec.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readItems drives monday's cursor-based item pagination. The first request
// fetches data.boards[].items_page{cursor, items}; subsequent requests follow
// the cursor through the next_items_page root field until the cursor is null.
func (c Connector) readItems(ctx context.Context, r *connsdk.Requester, pageSize, maxPages int, emit func(connectors.Record) error) error {
	firstQuery := fmt.Sprintf("query { boards (limit: %d) { id items_page (limit: %d) { cursor items { %s } } } }", pageSize, pageSize, itemSelection)
	body, err := c.execute(ctx, r, firstQuery)
	if err != nil {
		return fmt.Errorf("read monday items: %w", err)
	}

	cursor, err := c.emitItemsFromBoards(ctx, body, emit)
	if err != nil {
		return err
	}

	limit := maxPages
	if limit <= 0 {
		limit = mondaySafetyMaxPages
	}
	for page := 1; cursor != "" && page <= limit; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := fmt.Sprintf("query { next_items_page (limit: %d, cursor: %q) { cursor items { %s } } }", pageSize, cursor, itemSelection)
		body, err := c.execute(ctx, r, query)
		if err != nil {
			return fmt.Errorf("read monday items cursor page %d: %w", page, err)
		}
		if _, err := c.emitItems(ctx, body, "data.next_items_page.items", emit); err != nil {
			return err
		}
		cursor, _ = connsdk.StringAt(body, "data.next_items_page.cursor")
	}
	return nil
}

// emitItemsFromBoards emits the items embedded under data.boards[].items_page and
// returns the first non-empty items_page cursor to continue pagination from.
func (c Connector) emitItemsFromBoards(ctx context.Context, body []byte, emit func(connectors.Record) error) (string, error) {
	boards, err := connsdk.RecordsAt(body, "data.boards")
	if err != nil {
		return "", fmt.Errorf("decode monday items boards: %w", err)
	}
	cursor := ""
	for _, board := range boards {
		page, ok := board["items_page"].(map[string]any)
		if !ok {
			continue
		}
		if items, ok := page["items"].([]any); ok {
			for _, raw := range items {
				if obj, ok := raw.(map[string]any); ok {
					if err := ctx.Err(); err != nil {
						return "", err
					}
					if err := emit(itemRecord(obj)); err != nil {
						return "", err
					}
				}
			}
		}
		if cursor == "" {
			cursor = stringField(page, "cursor")
		}
	}
	return cursor, nil
}

// emitItems emits items found at the given dotted path and reports how many were
// emitted.
func (c Connector) emitItems(ctx context.Context, body []byte, path string, emit func(connectors.Record) error) (int, error) {
	records, err := connsdk.RecordsAt(body, path)
	if err != nil {
		return 0, fmt.Errorf("decode monday items page: %w", err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := emit(itemRecord(item)); err != nil {
			return 0, err
		}
	}
	return len(records), nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise monday credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if stream != "items" {
		if _, ok := pageStreamSpecs[stream]; !ok {
			return fmt.Errorf("monday stream %q not found", stream)
		}
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var record connectors.Record
		switch stream {
		case "items":
			record = itemRecord(map[string]any{
				"id":         fmt.Sprintf("item_fixture_%d", i),
				"name":       fmt.Sprintf("Fixture Item %d", i),
				"state":      "active",
				"created_at": mondayFixtureUpdated,
				"updated_at": mondayFixtureUpdated,
				"group":      map[string]any{"id": "grp_1", "title": "Group One"},
				"board":      map[string]any{"id": "1", "name": "Fixture Board"},
			})
		case "boards":
			record = boardRecord(map[string]any{
				"id":           fmt.Sprintf("%d", i),
				"name":         fmt.Sprintf("Fixture Board %d", i),
				"state":        "active",
				"board_kind":   "public",
				"description":  "fixture board",
				"type":         "board",
				"updated_at":   mondayFixtureUpdated,
				"workspace_id": "ws_1",
			})
		case "users":
			record = userRecord(map[string]any{
				"id":         fmt.Sprintf("u%d", i),
				"name":       fmt.Sprintf("Fixture User %d", i),
				"email":      fmt.Sprintf("fixture+%d@example.com", i),
				"enabled":    true,
				"is_admin":   false,
				"is_guest":   false,
				"is_pending": false,
				"created_at": mondayFixtureUpdated,
			})
		case "teams":
			record = teamRecord(map[string]any{
				"id":          fmt.Sprintf("t%d", i),
				"name":        fmt.Sprintf("Fixture Team %d", i),
				"picture_url": "",
			})
		case "tags":
			record = tagRecord(map[string]any{
				"id":    fmt.Sprintf("tag%d", i),
				"name":  fmt.Sprintf("fixture-tag-%d", i),
				"color": "blue",
			})
		}
		record["connector"] = "monday"
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// execute POSTs a GraphQL query and returns the raw response body. The request
// body is the standard {"query": "..."} envelope. monday returns HTTP 200 even
// for GraphQL errors, surfacing them under a top-level "errors" array, which is
// checked here so a malformed query is not silently treated as an empty read.
func (c Connector) execute(ctx context.Context, r *connsdk.Requester, query string) ([]byte, error) {
	payload := map[string]any{"query": query}
	resp, err := r.Do(ctx, http.MethodPost, "", nil, payload)
	if err != nil {
		return nil, err
	}
	if errMsg := graphQLError(resp.Body); errMsg != "" {
		return nil, fmt.Errorf("monday graphql error: %s", errMsg)
	}
	return resp.Body, nil
}

// graphQLError returns the first GraphQL error message if the response carries a
// non-empty top-level "errors" array, or "" otherwise. monday returns HTTP 200
// even for query errors, so this is the only way to detect them. connsdk's
// dotted-path helpers do not index into arrays, so the body is decoded directly.
func graphQLError(body []byte) string {
	// monday's legacy error envelope uses a flat error_message string.
	if msg, _ := connsdk.StringAt(body, "error_message"); msg != "" {
		return msg
	}
	var envelope struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return ""
	}
	if len(envelope.Errors) > 0 && strings.TrimSpace(envelope.Errors[0].Message) != "" {
		return envelope.Errors[0].Message
	}
	return ""
}

// requester builds a connsdk.Requester wired with monday's raw-token
// Authorization header and the API-Version header. The secret only ever flows
// into the authenticator; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := mondayBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := mondaySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("monday connector requires an api_token or access_token secret")
	}
	headers := map[string]string{}
	if version := strings.TrimSpace(cfg.Config["api_version"]); version != "" {
		headers["API-Version"] = version
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, ""),
		UserAgent: mondayUserAgent,
		// monday surfaces both its REST-style HTTP errors and a complexity-based
		// rate limit (HTTP 429), both of which connsdk's default retry covers.
		DefaultHeaders: headers,
	}, nil
}

// mondaySecret resolves the monday token. monday accepts either a personal API
// token (api_token) or an OAuth access token (access_token). The catalog nests
// these under "credentials", so both the dotted and bare key forms are accepted.
func mondaySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	for _, key := range []string{
		"credentials.api_token",
		"credentials.access_token",
		"api_token",
		"access_token",
	} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// mondayBaseURL resolves and validates the base URL. The default is
// api.monday.com/v2; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func mondayBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mondayDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("monday config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("monday config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("monday config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mondayPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mondayDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("monday config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mondayMaxPageSize {
		return 0, fmt.Errorf("monday config page_size must be between 1 and %d", mondayMaxPageSize)
	}
	return value, nil
}

func mondayMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("monday config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("monday config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// stringField coerces a map value to a string. monday returns numeric ids as
// JSON numbers in some shapes and strings in others; this normalizes them.
func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
