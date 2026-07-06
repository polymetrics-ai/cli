// Package microsoftlists implements the Tier-2 StreamHook for the
// microsoft-lists quarantine-repair migration (docs/migration/
// quarantine.json's microsoft-lists ENGINE_GAP entry, conventions.md §1's
// Tier-2 hooks table).
//
// Microsoft Graph's list pagination is "@odata.nextLink" — a JSON key
// containing a literal dot — carrying the next page's FULL ABSOLUTE URL
// (with its own $skiptoken cursor already embedded). Legacy hand-rolls this
// exactly (internal/connectors/microsoft-lists/microsoft-lists.go's
// harvest/nextLink, read-only reference): GET the resource (with any
// per-stream extra query merged in), extract value[], decode the top-level
// "@odata.nextLink" string directly (not via a dotted-path helper — the
// key's own literal dot makes dotted-path addressing ambiguous), and if
// non-empty, re-request that exact URL verbatim. The engine's declarative
// next_url pagination type reads its cursor via connsdk.StringAt's
// dotted-path parser, which splits on "." and therefore cannot address a
// literal key containing a dot — this is the confirmed ENGINE_GAP this hook
// resolves without an engine change (see defs/microsoft-lists/docs.md's
// Streams notes for the full reasoning).
//
// Only one hook interface is implemented (StreamHook), well under the
// 2-interface Tier-2 cap; auth stays fully declarative
// (oauth2_client_credentials in streams.json, dual when-gated candidates
// mirroring sharepoint-lists-enterprise/microsoft-teams/microsoft-entra-id).
//
// Line-count note (conventions.md §1's ~300-line soft target): this file
// runs to ~300 lines because it carries 4 distinct stream shapes in one
// routing table (graphEndpoints) — a site-scoped stream (lists) plus 3
// list-scoped streams needing config.list_id validation, one of which
// (list_items) also merges a per-stream extra query param ($expand=fields)
// — and 2 of the 4 mapRecord cases flatten a nested facet (list.template,
// contentType.id) rather than a flat rename. This is the same class of
// "several distinct per-stream shapes in one hook" justification the
// mini-wave sizing guidance calls out (e.g. monday's 2 pagination shapes +
// 5 record mappers); well under the 400-line hard ceiling and still a
// single hook interface.
package microsoftlists

import (
	"context"
	"encoding/json"
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
	maxPageSize     = 200
)

// graphEndpoint describes one stream's Graph resource shape: the path
// segment (after /sites/{site_id}/) to append, relative to a possibly
// list-scoped resource, whether it requires config.list_id, and any extra
// query params merged into every page request.
type graphEndpoint struct {
	resourceFor func(listID string) string
	needsListID bool
	extraQuery  map[string]string
}

// graphEndpoints is the per-stream routing table — the exact routing table
// legacy's streamEndpoints declares (microsoft-lists/streams.go).
var graphEndpoints = map[string]graphEndpoint{
	"lists": {
		resourceFor: func(string) string { return "/lists" },
	},
	"list_items": {
		resourceFor: func(listID string) string { return "/lists/" + listID + "/items" },
		needsListID: true,
		extraQuery:  map[string]string{"$expand": "fields"},
	},
	"columns": {
		resourceFor: func(listID string) string { return "/lists/" + listID + "/columns" },
		needsListID: true,
	},
	"content_types": {
		resourceFor: func(listID string) string { return "/lists/" + listID + "/contentTypes" },
		needsListID: true,
	},
}

func init() {
	engine.RegisterHooks("microsoft-lists", func() engine.Hooks { return New() })
}

// Hooks implements engine.StreamHook for the microsoft-lists bundle. It has
// no state of its own; every method is a pure function of its arguments.
type Hooks struct{}

// New returns a fresh microsoft-lists Hooks value (StreamHook
// implementation).
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "microsoft-lists" }

// ReadStream drives Microsoft Graph's @odata.nextLink pagination for every
// stream this bundle declares. handled is false only for a stream name this
// hook does not recognize (should not happen for a correctly-authored
// bundle; returning handled=false rather than panicking keeps the
// declarative path as an honest fallback per the Hooks interface contract).
func (h *Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}

	endpoint, ok := graphEndpoints[stream.Name]
	if !ok {
		return false, nil
	}
	schema := rt.Bundle.Schemas[stream.Name]
	if schema == nil {
		return false, nil
	}

	siteID := strings.TrimSpace(req.Config.Config["site_id"])
	if siteID == "" {
		return true, fmt.Errorf("microsoft-lists connector requires config site_id")
	}
	listID := ""
	if endpoint.needsListID {
		listID = strings.TrimSpace(req.Config.Config["list_id"])
		if listID == "" {
			return true, fmt.Errorf("microsoft-lists stream %q requires config list_id", stream.Name)
		}
	}

	pageSize, err := pageSizeFor(req.Config)
	if err != nil {
		return true, err
	}
	maxPages, err := maxPagesFor(req.Config)
	if err != nil {
		return true, err
	}

	props := schema.Properties()
	path := "/sites/" + url.PathEscape(siteID) + endpoint.resourceFor(url.PathEscape(listID))
	query := url.Values{"$top": []string{strconv.Itoa(pageSize)}}
	for k, v := range endpoint.extraQuery {
		query.Set(k, v)
	}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return true, err
		}

		resp, err := rt.Requester.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return true, fmt.Errorf("microsoft-lists: read %s: %w", stream.Name, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "value")
		if err != nil {
			return true, fmt.Errorf("microsoft-lists: decode %s page: %w", stream.Name, err)
		}
		for _, raw := range records {
			if err := ctx.Err(); err != nil {
				return true, err
			}
			mapped := mapRecord(stream.Name, raw)
			if err := emit(connectors.Record(projectBySchema(mapped, props))); err != nil {
				return true, err
			}
		}

		next, err := nextLink(resp.Body)
		if err != nil {
			return true, fmt.Errorf("microsoft-lists: decode %s nextLink: %w", stream.Name, err)
		}
		if strings.TrimSpace(next) == "" {
			return true, nil
		}
		// nextLink is an absolute URL that already carries the cursor and any
		// page-size/extra-query hints; subsequent pages must not re-merge query.
		path = next
		query = nil
	}
	return true, nil
}

// nextLink extracts the Microsoft Graph "@odata.nextLink" absolute URL from
// a collection response body. The key contains a literal dot, so the
// engine's dotted-path helpers cannot select it; decode the top-level
// object directly, exactly like legacy's own nextLink helper.
func nextLink(body []byte) (string, error) {
	var envelope struct {
		NextLink string `json:"@odata.nextLink"`
	}
	if len(body) == 0 {
		return "", nil
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", fmt.Errorf("decode graph envelope: %w", err)
	}
	return strings.TrimSpace(envelope.NextLink), nil
}

// pageSizeFor mirrors legacy's graphPageSize (microsoft-lists.go).
func pageSizeFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("microsoft-lists config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("microsoft-lists config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

// maxPagesFor mirrors legacy's graphMaxPages (microsoft-lists.go).
func maxPagesFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("microsoft-lists config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, fmt.Errorf("microsoft-lists config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// mapRecord renames a raw Graph object's camelCase fields to the bundle's
// snake_case schema fields, byte-for-byte matching legacy's per-stream
// mapRecord functions (microsoft-lists/streams.go), including the two
// nested-facet flattenings (list.template -> list_template,
// contentType.id -> content_type_id).
func mapRecord(stream string, raw map[string]any) map[string]any {
	switch stream {
	case "lists":
		out := map[string]any{
			"id":                      raw["id"],
			"name":                    raw["name"],
			"display_name":            raw["displayName"],
			"description":             raw["description"],
			"web_url":                 raw["webUrl"],
			"created_date_time":       raw["createdDateTime"],
			"last_modified_date_time": raw["lastModifiedDateTime"],
			"etag":                    raw["eTag"],
		}
		if info, ok := raw["list"].(map[string]any); ok {
			out["list_template"] = info["template"]
		}
		return out
	case "list_items":
		return map[string]any{
			"id":                      raw["id"],
			"content_type_id":         contentTypeID(raw),
			"web_url":                 raw["webUrl"],
			"created_date_time":       raw["createdDateTime"],
			"last_modified_date_time": raw["lastModifiedDateTime"],
			"etag":                    raw["eTag"],
			"fields":                  raw["fields"],
		}
	case "columns":
		return map[string]any{
			"id":           raw["id"],
			"name":         raw["name"],
			"display_name": raw["displayName"],
			"description":  raw["description"],
			"column_group": raw["columnGroup"],
			"required":     raw["required"],
			"read_only":    raw["readOnly"],
			"hidden":       raw["hidden"],
			"indexed":      raw["indexed"],
		}
	case "content_types":
		return map[string]any{
			"id":          raw["id"],
			"name":        raw["name"],
			"description": raw["description"],
			"group":       raw["group"],
			"hidden":      raw["hidden"],
			"read_only":   raw["readOnly"],
			"sealed":      raw["sealed"],
		}
	default:
		return raw
	}
}

// contentTypeID pulls the nested contentType.id off a list item, if present.
func contentTypeID(item map[string]any) any {
	if ct, ok := item["contentType"].(map[string]any); ok {
		return ct["id"]
	}
	return nil
}

// projectBySchema keeps only the schema-declared properties from mapped,
// matching conventions.md §2's schema-as-projection rule.
func projectBySchema(mapped map[string]any, props []string) map[string]any {
	out := make(map[string]any, len(props))
	for _, name := range props {
		if v, ok := mapped[name]; ok {
			out[name] = v
		}
	}
	return out
}
