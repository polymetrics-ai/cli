// Package trello implements the native pm Trello connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit: a
// Requester wired with Trello's key/token query-param auth plus Trello-specific
// stream definitions, endpoints, and id-cursor pagination.
//
// Like stripe and github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// Trello is read-only here: its API mutations (creating cards/boards) are not a
// safe general reverse-ETL surface, so Capabilities.Write is false.
package trello

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	trelloDefaultBaseURL  = "https://api.trello.com/1"
	trelloDefaultPageSize = 100
	trelloMaxPageSize     = 1000
	trelloUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("trello", New)
}

// New returns the Trello connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Trello connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "trello" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "trello",
		DisplayName:     "Trello",
		IntegrationType: "api",
		Description:     "Reads Trello boards, lists, cards, checklists, and board actions through the Trello REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Trello. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := trelloBaseURL(cfg); err != nil {
		return err
	}
	key, token := trelloCredentials(cfg)
	if strings.TrimSpace(key) == "" || strings.TrimSpace(token) == "" {
		return errors.New("trello connector requires secrets key and token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the authenticated member confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "members/me", url.Values{"fields": []string{"id"}}, nil, nil); err != nil {
		return fmt.Errorf("check trello: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: trelloStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Trello stream starts with
// an empty incremental cursor (full sync), which start_date can raise.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Write is unsupported: Trello is exposed read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "boards"
	}
	endpoint, ok := trelloStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("trello stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := trelloPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := trelloMaxPages(req.Config)
	if err != nil {
		return err
	}

	if endpoint.scope == scopeBoards {
		return c.readBoards(ctx, r, endpoint, emit)
	}

	// Board-scoped streams: discover the boards, then read the sub-resource for
	// each. Board discovery reuses the boards endpoint mapper-free path.
	boardIDs, err := c.boardIDs(ctx, r, req.Config)
	if err != nil {
		return err
	}
	for _, boardID := range boardIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := c.readBoardResource(ctx, r, boardID, endpoint, pageSize, maxPages, emit); err != nil {
			return err
		}
	}
	return nil
}

// readBoards emits the boards stream itself, honouring configured board_ids when
// present and otherwise listing the member's boards.
func (c Connector) readBoards(ctx context.Context, r *requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	// boardIDs handled inline so each fetched board is mapped through the boards
	// record mapper.
	return c.eachBoardObject(ctx, r, func(item map[string]any) error {
		return emit(endpoint.mapRecord(item))
	})
}

// eachBoardObject fetches each configured board (by id) or every member board
// and invokes fn on the raw object.
func (c Connector) eachBoardObject(ctx context.Context, r *requester, fn func(map[string]any) error) error {
	ids := configuredBoardIDs(r.cfgBoardIDs)
	if len(ids) == 0 {
		resp, err := r.Do(ctx, http.MethodGet, "members/me/boards", nil, nil)
		if err != nil {
			return fmt.Errorf("list trello boards: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode trello boards: %w", err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := fn(item); err != nil {
				return err
			}
		}
		return nil
	}
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, "boards/"+url.PathEscape(id), nil, nil)
		if err != nil {
			return fmt.Errorf("get trello board %s: %w", id, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode trello board %s: %w", id, err)
		}
		for _, item := range records {
			if err := fn(item); err != nil {
				return err
			}
		}
	}
	return nil
}

// boardIDs returns the board IDs to iterate for board-scoped streams. It uses
// configured IDs when present, otherwise discovers the member's boards.
func (c Connector) boardIDs(ctx context.Context, r *requester, cfg connectors.RuntimeConfig) ([]string, error) {
	if ids := configuredBoardIDs(parseBoardIDs(cfg)); len(ids) > 0 {
		return ids, nil
	}
	var ids []string
	err := c.eachBoardObject(ctx, r, func(item map[string]any) error {
		if id := stringField(item, "id"); id != "" {
			ids = append(ids, id)
		}
		return nil
	})
	return ids, err
}

// readBoardResource reads one board-scoped sub-resource. Paginated resources
// (cards, actions) use Trello's id-cursor `before` paging; the rest return the
// full set in a single call.
func (c Connector) readBoardResource(ctx context.Context, r *requester, boardID string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := "boards/" + url.PathEscape(boardID) + "/" + endpoint.resource
	if !endpoint.paginated {
		resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
		if err != nil {
			return fmt.Errorf("read trello %s for board %s: %w", endpoint.resource, boardID, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode trello %s for board %s: %w", endpoint.resource, boardID, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(injectBoardID(item, boardID))); err != nil {
				return err
			}
		}
		return nil
	}

	before := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if before != "" {
			query.Set("before", before)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read trello %s for board %s: %w", endpoint.resource, boardID, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode trello %s for board %s: %w", endpoint.resource, boardID, err)
		}
		lastID := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			lastID = stringField(item, "id")
			if err := emit(endpoint.mapRecord(injectBoardID(item, boardID))); err != nil {
				return err
			}
		}
		// A short page (fewer than the requested limit) means we reached the end.
		if len(records) < pageSize || lastID == "" {
			return nil
		}
		before = lastID
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise trello credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":             fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"desc":             "fixture record",
			"closed":           false,
			"idBoard":          "board_fixture_1",
			"idOrganization":   "org_fixture_1",
			"idList":           "list_fixture_1",
			"idCard":           "card_fixture_1",
			"idMemberCreator":  "member_fixture_1",
			"type":             "createCard",
			"date":             "2026-01-01T00:00:00.000Z",
			"dateLastActivity": "2026-01-01T00:00:00.000Z",
			"url":              "https://trello.com/b/fixture",
			"shortUrl":         "https://trello.com/b/fix",
			"pos":              float64(i),
			"subscribed":       false,
			"dueComplete":      false,
			"connector":        "trello",
			"fixture":          true,
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

// requester builds a connsdk.Requester wired with Trello's key/token query-param
// auth and the resolved base URL. The secrets only ever flow into the
// authenticator query params; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*requester, error) {
	base, err := trelloBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	key, token := trelloCredentials(cfg)
	if strings.TrimSpace(key) == "" || strings.TrimSpace(token) == "" {
		return nil, errors.New("trello connector requires secrets key and token")
	}
	return &requester{
		Requester: connsdk.Requester{
			Client:    c.Client,
			BaseURL:   base,
			Auth:      keyTokenAuth(key, token),
			UserAgent: trelloUserAgent,
		},
		cfgBoardIDs: parseBoardIDs(cfg),
	}, nil
}

// requester wraps connsdk.Requester with the parsed board_ids config so the
// boards path can honour them without re-threading cfg through every call.
type requester struct {
	connsdk.Requester
	cfgBoardIDs []string
}

// keyTokenAuth applies Trello's two-parameter query auth (?key=..&token=..).
func keyTokenAuth(key, token string) connsdk.Authenticator {
	return connsdk.AuthFunc(func(_ context.Context, req *http.Request) error {
		q := req.URL.Query()
		q.Set("key", key)
		q.Set("token", token)
		req.URL.RawQuery = q.Encode()
		return nil
	})
}

func trelloCredentials(cfg connectors.RuntimeConfig) (key, token string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["key"], cfg.Secrets["token"]
}

// trelloBaseURL resolves and validates the base URL. The default is
// api.trello.com/1; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func trelloBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return trelloDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("trello config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("trello config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("trello config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func trelloPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return trelloDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("trello config page_size must be an integer: %w", err)
	}
	if value < 1 || value > trelloMaxPageSize {
		return 0, fmt.Errorf("trello config page_size must be between 1 and %d", trelloMaxPageSize)
	}
	return value, nil
}

func trelloMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("trello config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("trello config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// parseBoardIDs reads the optional board_ids config (comma or whitespace
// separated 24-char hex IDs).
func parseBoardIDs(cfg connectors.RuntimeConfig) []string {
	raw := strings.TrimSpace(cfg.Config["board_ids"])
	if raw == "" {
		return nil
	}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

func configuredBoardIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	return ids
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// injectBoardID ensures a board-scoped record carries its board id even when the
// upstream object omits it (Trello lists/checklists include idBoard already, but
// this is a cheap safety net).
func injectBoardID(item map[string]any, boardID string) map[string]any {
	if item == nil {
		return item
	}
	if v, ok := item["idBoard"]; !ok || v == nil || v == "" {
		item["idBoard"] = boardID
	}
	return item
}

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
