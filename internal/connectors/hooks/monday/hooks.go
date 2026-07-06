// Package monday is the wave1-pilot Tier-2 StreamHook (+ CheckHook) for the
// monday defs bundle: monday.com's single GraphQL endpoint (POST
// https://api.monday.com/v2) carries pagination state INSIDE the request
// body, which StreamSpec.Body (engine/bundle.go) cannot express — the
// declarative read path never sends a body (engine/read.go:142 always passes
// nil). This ports monday.go/streams.go's GraphQL query construction,
// in-body pagination, and record mapping verbatim via StreamHook.ReadStream
// (all 5 streams) and CheckHook.Check (legacy's `query { me { id } }` check,
// monday.go:90-96), both reusing rt.Requester (the engine's already-built
// *connsdk.Requester: base URL/auth/headers already resolved).
package monday

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// mirror monday.go's identically named constants.
const (
	mondaySafetyMaxPages  = 10000
	mondayDefaultPageSize = 50
)

func init() {
	engine.RegisterHooks("monday", func() engine.Hooks { return Hooks{} })
}

// Hooks is monday's Tier-2 hook set: StreamHook (all reads) + CheckHook.
type Hooks struct{}

func (Hooks) ConnectorName() string { return "monday" }

// pageStreamSpec describes one page-number-paginated GraphQL root field
// (ported from monday/streams.go's streamSpec/pageStreamSpecs).
type pageStreamSpec struct {
	root        string
	selection   string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

// pageStreamSpecs is the routing table for page-number paginated streams.
var pageStreamSpecs = map[string]pageStreamSpec{
	"boards": {root: "boards", selection: "id name state board_kind description type updated_at workspace_id", recordsPath: "boards", mapRecord: boardRecord},
	"users":  {root: "users", selection: "id name email enabled is_admin is_guest is_pending created_at", recordsPath: "users", mapRecord: userRecord},
	"teams":  {root: "teams", selection: "id name picture_url", recordsPath: "teams", mapRecord: teamRecord},
	"tags":   {root: "tags", selection: "id name color", recordsPath: "tags", mapRecord: tagRecord},
}

// itemSelection is the GraphQL field selection for an item, shared by the
// boards { items_page } and next_items_page queries (monday/streams.go).
const itemSelection = "id name state created_at updated_at group { id title } board { id name }"

// ReadStream implements engine.StreamHook, handling every declared stream
// (boards, items, users, teams, tags) with handled=true always — the
// declarative streams.json fallback is a structural "shadow" path exercised
// only by conformance's dynamic checks (Hooks=nil there), never here (docs.md
// "Known limits").
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	name := stream.Name
	if name == "" {
		name = "boards"
	}

	pageSize := mondayPageSize(req.Config)
	maxPages := mondayMaxPages(req.Config)

	if name == "items" {
		return true, h.readItems(ctx, rt.Requester, pageSize, maxPages, emit)
	}
	spec, ok := pageStreamSpecs[name]
	if !ok {
		return false, nil
	}
	return true, h.readPaged(ctx, rt.Requester, spec, pageSize, maxPages, emit)
}

// Check implements engine.CheckHook: a bounded `query { me { id } }` confirms
// auth and connectivity without reading data (monday.go:90-96).
func (h Hooks) Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *engine.Runtime) (bool, error) {
	if _, err := execute(ctx, rt.Requester, `query { me { id } }`); err != nil {
		return true, fmt.Errorf("check monday: %w", err)
	}
	return true, nil
}

// readPaged drives page-number pagination for boards/users/teams/tags (ported
// from monday.go's readPaged): a short page (fewer records than page size)
// signals the end.
func (h Hooks) readPaged(ctx context.Context, r *connsdk.Requester, spec pageStreamSpec, pageSize, maxPages int, emit func(connectors.Record) error) error {
	limit := maxPages
	if limit <= 0 {
		limit = mondaySafetyMaxPages
	}
	for page := 1; page <= limit; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := fmt.Sprintf("query { %s (limit: %d, page: %d) { %s } }", spec.root, pageSize, page, spec.selection)
		body, err := execute(ctx, r, query)
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

// readItems drives cursor-based item pagination (ported from monday.go's
// readItems): the first request fetches data.boards[].items_page{cursor,
// items}; subsequent requests follow the cursor via next_items_page until null.
func (h Hooks) readItems(ctx context.Context, r *connsdk.Requester, pageSize, maxPages int, emit func(connectors.Record) error) error {
	firstQuery := fmt.Sprintf("query { boards (limit: %d) { id items_page (limit: %d) { cursor items { %s } } } }", pageSize, pageSize, itemSelection)
	body, err := execute(ctx, r, firstQuery)
	if err != nil {
		return fmt.Errorf("read monday items: %w", err)
	}

	cursor, err := emitItemsFromBoards(ctx, body, emit)
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
		body, err := execute(ctx, r, query)
		if err != nil {
			return fmt.Errorf("read monday items cursor page %d: %w", page, err)
		}
		if _, err := emitItems(ctx, body, "data.next_items_page.items", emit); err != nil {
			return err
		}
		cursor, _ = connsdk.StringAt(body, "data.next_items_page.cursor")
	}
	return nil
}

// emitItemsFromBoards emits items embedded under data.boards[].items_page and
// returns the first non-empty items_page cursor to continue from.
func emitItemsFromBoards(ctx context.Context, body []byte, emit func(connectors.Record) error) (string, error) {
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
				obj, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				if err := ctx.Err(); err != nil {
					return "", err
				}
				if err := emit(itemRecord(obj)); err != nil {
					return "", err
				}
			}
		}
		if cursor == "" {
			cursor = stringField(page, "cursor")
		}
	}
	return cursor, nil
}

// emitItems emits items found at the given dotted path.
func emitItems(ctx context.Context, body []byte, path string, emit func(connectors.Record) error) (int, error) {
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

// execute POSTs a GraphQL query and returns the raw response body. monday
// returns HTTP 200 even for GraphQL errors (a top-level "errors" array),
// checked here so a malformed query is never silently treated as empty.
func execute(ctx context.Context, r *connsdk.Requester, query string) ([]byte, error) {
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

// graphQLError returns the first GraphQL error message, or "".
func graphQLError(body []byte) string {
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

// mondayPageSize/mondayMaxPages resolve config.page_size/config.max_pages;
// unlike monday.go's helpers these never error: an invalid page_size falls
// back to mondayDefaultPageSize, and ""/"all"/"unlimited"/invalid max_pages
// means unbounded (0).
func mondayPageSize(cfg connectors.RuntimeConfig) int {
	if v, ok := parsePositiveInt(cfg.Config["page_size"]); ok {
		return v
	}
	return mondayDefaultPageSize
}

func mondayMaxPages(cfg connectors.RuntimeConfig) int {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0
	}
	if v, ok := parsePositiveInt(raw); ok {
		return v
	}
	return 0
}

func parsePositiveInt(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	var value int
	if _, err := fmt.Sscanf(raw, "%d", &value); err != nil || value < 1 {
		return 0, false
	}
	return value, true
}

// stringField coerces a map value to a string: monday returns numeric ids as
// JSON numbers in some shapes and strings in others.
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

// --- record mapping (ported from monday/streams.go) ---

func boardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id": stringField(item, "id"), "name": item["name"], "state": item["state"],
		"board_kind": item["board_kind"], "description": item["description"], "type": item["type"],
		"updated_at": item["updated_at"], "workspace_id": stringField(item, "workspace_id"),
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id": stringField(item, "id"), "name": item["name"], "email": item["email"],
		"enabled": item["enabled"], "is_admin": item["is_admin"], "is_guest": item["is_guest"],
		"is_pending": item["is_pending"], "created_at": item["created_at"],
	}
}

func teamRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": stringField(item, "id"), "name": item["name"], "picture_url": item["picture_url"]}
}

func tagRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": stringField(item, "id"), "name": item["name"], "color": item["color"]}
}

// itemRecord flattens a monday item, hoisting nested group/board objects into
// flat group_*/board_* columns.
func itemRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id": stringField(item, "id"), "name": item["name"], "state": item["state"],
		"created_at": item["created_at"], "updated_at": item["updated_at"],
	}
	if group, ok := item["group"].(map[string]any); ok {
		rec["group_id"] = stringField(group, "id")
		rec["group_title"] = group["title"]
	}
	if board, ok := item["board"].(map[string]any); ok {
		rec["board_id"] = stringField(board, "id")
		rec["board_name"] = board["name"]
	}
	return rec
}
