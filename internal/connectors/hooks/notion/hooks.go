// Package notion implements the notion bundle's Tier-2 StreamHook (wave2
// quarantine repair, docs/migration/quarantine.json's "notion" ENGINE_GAP
// entry): the databases/pages streams require POST /search with the
// pagination cursor (start_cursor) and object filter injected into the JSON
// request body on every page. The engine's declarative read path
// (engine/read.go's readDeclarative) always issues rt.Requester.Do(ctx,
// method, path, query, nil) -- the body argument is hardcoded nil on every
// declarative read; StreamSpec.Body (engine/bundle.go's
// json:"body,omitempty" field, commented "POST-body streams") is declared
// but never read anywhere in read.go -- dead/unwired. This ports legacy
// internal/connectors/notion/notion.go's harvest loop verbatim (~140 lines,
// well under the 300-line Tier-2 soft target, docs/migration/conventions.md
// §1). Only one hook interface is implemented (StreamHook); auth is fully
// declarative (bearer in streams.json) and needs no AuthHook at all.
package notion

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	defaultPageSize = 100
	maxPageSize     = 100
)

func init() {
	engine.RegisterHooks("notion", func() engine.Hooks { return Hooks{} })
}

// Hooks is the notion bundle's Tier-2 hook set: StreamHook only. It has no
// state of its own; every method is a pure function of its arguments.
type Hooks struct{}

func (Hooks) ConnectorName() string { return "notion" }

// streamRoute mirrors legacy's notionStreamEndpoints routing table
// (notion/streams.go): the API resource, HTTP method, and object filter
// (empty for GET endpoints) each stream reads from.
type streamRoute struct {
	resource     string
	method       string
	searchObject string
}

var streamRoutes = map[string]streamRoute{
	"databases": {resource: "/search", method: http.MethodPost, searchObject: "database"},
	"pages":     {resource: "/search", method: http.MethodPost, searchObject: "page"},
	"users":     {resource: "/users", method: http.MethodGet},
}

// ReadStream drives Notion's start_cursor pagination for every stream this
// bundle declares. handled is always true for a recognized stream name; an
// unrecognized name returns handled=false so the declarative fallback stays
// an honest path per the Hooks interface contract (should not happen for a
// correctly authored bundle).
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}

	name := stream.Name
	if name == "" {
		name = "databases"
	}
	route, ok := streamRoutes[name]
	if !ok {
		return false, nil
	}

	schema := rt.Bundle.Schemas[name]
	if schema == nil {
		return false, nil
	}
	props := schema.Properties()

	pageSize, err := pageSizeFor(req.Config)
	if err != nil {
		return true, err
	}
	maxPages := maxPagesFor(req.Config)

	return true, h.harvest(ctx, rt.Requester, route, pageSize, maxPages, props, emit)
}

// harvest drives Notion's start_cursor pagination. Every list response is
// {results:[...], next_cursor:string|null, has_more:bool}. POST /search
// carries the cursor and object filter in the request body; GET /users
// carries the cursor as the start_cursor query parameter. Byte-for-byte
// port of legacy's harvest (notion.go:146-205).
func (h Hooks) harvest(ctx context.Context, r *connsdk.Requester, route streamRoute, pageSize, maxPages int, props []string, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		var (
			resp *connsdk.Response
			err  error
		)
		if route.method == http.MethodPost {
			body := map[string]any{"page_size": pageSize}
			if route.searchObject != "" {
				body["filter"] = map[string]any{"property": "object", "value": route.searchObject}
			}
			if cursor != "" {
				body["start_cursor"] = cursor
			}
			resp, err = r.Do(ctx, http.MethodPost, route.resource, nil, body)
		} else {
			query := url.Values{}
			query.Set("page_size", strconv.Itoa(pageSize))
			if cursor != "" {
				query.Set("start_cursor", cursor)
			}
			resp, err = r.Do(ctx, http.MethodGet, route.resource, query, nil)
		}
		if err != nil {
			return fmt.Errorf("notion: read %s: %w", route.resource, err)
		}

		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("notion: decode %s page: %w", route.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(connectors.Record(projectBySchema(item, props))); err != nil {
				return err
			}
		}

		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return fmt.Errorf("notion: decode %s has_more: %w", route.resource, err)
		}
		next, err := connsdk.StringAt(resp.Body, "next_cursor")
		if err != nil {
			return fmt.Errorf("notion: decode %s next_cursor: %w", route.resource, err)
		}
		if hasMore != "true" || strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// projectBySchema keeps only the schema-declared properties from raw,
// matching conventions.md §2's schema-as-projection rule -- notion's
// schemas are derived field-for-field from legacy's mapRecord functions,
// all of which use the raw API field name verbatim (no renames), so a
// plain exact-key-match copy is sufficient (no graph-style rename table
// needed here).
func projectBySchema(raw map[string]any, props []string) map[string]any {
	out := make(map[string]any, len(props))
	for _, name := range props {
		if v, ok := raw[name]; ok {
			out[name] = v
		}
	}
	return out
}

// pageSizeFor mirrors legacy's notionPageSize (notion.go:303-316).
func pageSizeFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("notion config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("notion config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

// maxPagesFor mirrors legacy's notionMaxPages (notion.go:318-331):
// permissive parse, never errors -- an empty/"all"/"unlimited"/malformed/
// negative value means unbounded (0).
func maxPagesFor(cfg connectors.RuntimeConfig) int {
	raw := strings.ToLower(strings.TrimSpace(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
