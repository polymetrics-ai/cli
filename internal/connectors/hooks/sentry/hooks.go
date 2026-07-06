// Package sentry implements the Tier-2 StreamHook for the sentry pilot
// migration (PLAN.md P-5, SPEC.md §5.3 resolution ladder rung 2).
//
// Sentry's list endpoints paginate via an RFC 5988 Link header with a twist
// legacy hand-rolls precisely (internal/connectors/sentry/sentry.go:7-9,
// 144-152, read-only reference): a rel="next" Link entry is ALWAYS present,
// even on the truly-last page, and the real "more pages" signal is that
// entry's results="true"/"false" attribute. The engine's declarative
// link_header pagination type has no knowledge of results= at all
// (engine/paginate.go's linkHeaderPaginator follows rel="next"
// unconditionally) — see defs/sentry/docs.md's "Streams notes" for the full
// ladder-rejection evidence (conformance's pagination_terminates check hard
// -fails on the extra trailing request a Tier-1 paginator would always
// issue). This hook ports legacy's harvest/nextCursor logic exactly: same
// request shape (per_page + cursor query params), same stop condition
// (results="false", or no next link at all).
package sentry

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
	engine.RegisterHooks("sentry", func() engine.Hooks { return New() })
}

// Hooks implements engine.StreamHook for the sentry bundle. It has no
// state of its own; every method is a pure function of its arguments.
type Hooks struct{}

// New returns a fresh sentry Hooks value (StreamHook implementation).
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "sentry" }

// ReadStream drives Sentry's Link-header cursor pagination for every stream
// this bundle declares — the declarative fallback (streams.json's
// pagination.type: none) is never taken in production; handled is always
// true here except for a stream name this hook does not recognize (should
// not happen for a correctly-authored bundle, but returning handled=false
// rather than panicking keeps the declarative path as an honest fallback
// per the Hooks interface contract).
func (h *Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}

	schema := rt.Bundle.Schemas[stream.Name]
	if schema == nil {
		return false, nil
	}

	path, err := resolveStreamPath(stream, req.Config)
	if err != nil {
		return true, err
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

	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))

	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return true, err
		}
		query := cloneValues(base)
		if cursor != "" {
			query.Set("cursor", cursor)
		}

		resp, err := rt.Requester.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return true, fmt.Errorf("sentry: read %s: %w", stream.Name, err)
		}

		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return true, fmt.Errorf("sentry: decode %s page: %w", stream.Name, err)
		}
		for _, raw := range records {
			if err := ctx.Err(); err != nil {
				return true, err
			}
			if err := emit(connectors.Record(projectBySchema(raw, props))); err != nil {
				return true, err
			}
		}

		next, more := nextCursor(resp.Header.Get("Link"))
		if !more || next == "" {
			return true, nil
		}
		cursor = next
	}
	return true, nil
}

// resolveStreamPath builds the request path for stream, matching legacy's
// endpointPath (sentry/sentry.go:248-276) exactly: projects is
// org/project-independent, issues/events require organization+project,
// releases requires organization only.
func resolveStreamPath(stream engine.StreamSpec, cfg connectors.RuntimeConfig) (string, error) {
	switch stream.Name {
	case "projects":
		return "/api/0/projects/", nil
	case "issues", "events":
		org, err := requireSlug(cfg, "organization")
		if err != nil {
			return "", err
		}
		project, err := requireSlug(cfg, "project")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("/api/0/projects/%s/%s/%s/", org, project, stream.Name), nil
	case "releases":
		org, err := requireSlug(cfg, "organization")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("/api/0/organizations/%s/releases/", org), nil
	default:
		return "", fmt.Errorf("sentry: stream %q has no known endpoint", stream.Name)
	}
}

func requireSlug(cfg connectors.RuntimeConfig, key string) (string, error) {
	v := strings.TrimSpace(cfg.Config[key])
	if v == "" {
		return "", fmt.Errorf("sentry connector requires config %s", key)
	}
	return url.PathEscape(v), nil
}

// pageSizeFor mirrors legacy's sentryPageSize (sentry/sentry.go:323-336).
func pageSizeFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("sentry config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("sentry config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

// maxPagesFor mirrors legacy's sentryMaxPages (sentry/sentry.go:338-351).
func maxPagesFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("sentry config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, fmt.Errorf("sentry config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// projectBySchema keeps only the schema-declared properties from raw,
// matching conventions.md §2's schema-as-projection rule — sentry's schemas
// are derived field-for-field from legacy's mapRecord functions, all of
// which use the raw API field name verbatim (no renames), so no
// computed_fields equivalent is needed here.
func projectBySchema(raw map[string]any, props []string) map[string]any {
	out := make(map[string]any, len(props))
	for _, name := range props {
		if v, ok := raw[name]; ok {
			out[name] = v
		}
	}
	return out
}

// nextCursor parses a Sentry Link header and returns the rel="next" cursor
// along with whether that page actually has more results (results="true").
// Byte-for-byte port of legacy's nextCursor (sentry/sentry.go:353-393):
// Sentry always emits a rel="next" entry, so the results attribute is
// authoritative, not the mere presence of the link.
func nextCursor(header string) (cursor string, more bool) {
	if header == "" {
		return "", false
	}
	for _, part := range strings.Split(header, ",") {
		segs := strings.Split(part, ";")
		if len(segs) < 2 {
			continue
		}
		urlPart := strings.TrimSpace(segs[0])
		if !strings.HasPrefix(urlPart, "<") || !strings.HasSuffix(urlPart, ">") {
			continue
		}
		isNext := false
		results := false
		cur := ""
		for _, attr := range segs[1:] {
			k, v := splitAttr(attr)
			switch k {
			case "rel":
				isNext = v == "next"
			case "results":
				results = v == "true"
			case "cursor":
				cur = v
			}
		}
		if !isNext {
			continue
		}
		if cur == "" {
			cur = cursorFromURL(urlPart[1 : len(urlPart)-1])
		}
		return cur, results
	}
	return "", false
}

// splitAttr parses a Link header attribute of the form key="value" or
// key=value. Byte-for-byte port of legacy's splitAttr.
func splitAttr(attr string) (string, string) {
	attr = strings.TrimSpace(attr)
	eq := strings.IndexByte(attr, '=')
	if eq < 0 {
		return strings.ToLower(attr), ""
	}
	key := strings.ToLower(strings.TrimSpace(attr[:eq]))
	val := strings.TrimSpace(attr[eq+1:])
	val = strings.Trim(val, `"`)
	return key, val
}

// cursorFromURL extracts the cursor query parameter from a next-page URL.
// Byte-for-byte port of legacy's cursorFromURL.
func cursorFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Query().Get("cursor")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}
