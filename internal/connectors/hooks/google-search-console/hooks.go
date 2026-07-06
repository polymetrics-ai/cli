// Package googlesearchconsole implements the google-search-console bundle's
// StreamHook (docs.md "Overview"): the Search Console v3 `searchAnalytics.query`
// endpoint is a POST whose JSON request body carries
// startDate/endDate/dimensions/type/dataState/rowLimit/startRow, and whose
// pagination state (startRow, advanced by the number of rows returned each
// page) lives INSIDE that body -- internal/connectors/engine/bundle.go's
// StreamSpec.Body field exists but internal/connectors/engine/read.go's
// declarative read path never sends a request body at all (the final
// argument to rt.Requester.Do is always a literal nil), so this read cannot
// be expressed in streams.json alone. This ports legacy
// internal/connectors/google-search-console's readAnalytics/analyticsRecord
// (google_search_console.go:207-285) verbatim, reusing rt.Requester (the
// engine's already-built HTTP client/auth/base-URL plumbing) exactly as the
// declarative path itself would.
//
// sites is NOT handled here -- it is a fully declarative GET read (see
// streams.json) and this StreamHook returns handled=false for it, falling
// back to the engine's own declarative read path. sitemaps is handled here
// because legacy accepts either site_urls or the single-site site_url fallback
// and stringifies warning/error count fields; fan_out config_key cannot
// express that fallback.
package googlesearchconsole

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("google-search-console", func() engine.Hooks { return New() })
}

// New returns a fresh google-search-console Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

// Hooks is the google-search-console hook set. It implements
// engine.StreamHook only.
type Hooks struct{}

func (h *Hooks) ConnectorName() string { return "google-search-console" }

// analyticsDimensions is the per-stream fixed one-dimension searchAnalytics
// dimension set, mirroring legacy streams.go's gscStreamDefs routing table.
var analyticsDimensions = map[string][]string{
	"search_analytics_by_date":    {"date"},
	"search_analytics_by_country": {"country"},
	"search_analytics_by_device":  {"device"},
	"search_analytics_by_page":    {"page"},
	"search_analytics_by_query":   {"query"},
}

// ReadStream implements engine.StreamHook. It handles sitemaps plus every
// search_analytics_by_* stream (handled=true); every other stream returns
// handled=false so the engine's declarative read path runs instead.
func (h *Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	if stream.Name == "sitemaps" {
		if err := h.readSitemaps(ctx, rt.Requester, req.Config, emit); err != nil {
			return true, err
		}
		return true, nil
	}

	dims, ok := analyticsDimensions[stream.Name]
	if !ok {
		return false, nil
	}

	startDate, endDate, err := analyticsDateRange(req)
	if err != nil {
		return true, err
	}
	pageSize, err := gscPageSize(req.Config)
	if err != nil {
		return true, err
	}
	maxPages, err := gscMaxPages(req.Config)
	if err != nil {
		return true, err
	}
	searchType := gscSearchType(req.Config)
	dataState := strings.TrimSpace(req.Config.Config["data_state"])

	for _, site := range siteURLs(req.Config) {
		if err := ctx.Err(); err != nil {
			return true, err
		}
		if err := h.readAnalyticsForSite(ctx, rt.Requester, site, dims, startDate, endDate, searchType, dataState, pageSize, maxPages, emit); err != nil {
			return true, err
		}
	}
	return true, nil
}

func (h *Hooks) readSitemaps(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	for _, site := range siteURLs(cfg) {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := "/sites/" + url.PathEscape(site) + "/sitemaps"
		resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
		if err != nil {
			return fmt.Errorf("read google-search-console sitemaps for %s: %w", site, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "sitemap")
		if err != nil {
			return fmt.Errorf("decode google-search-console sitemaps: %w", err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(sitemapRecord(site, item)); err != nil {
				return err
			}
		}
	}
	return nil
}

func sitemapRecord(site string, item map[string]any) connectors.Record {
	return connectors.Record{
		"site_url":          site,
		"path":              item["path"],
		"last_submitted":    item["lastSubmitted"],
		"last_downloaded":   item["lastDownloaded"],
		"is_pending":        item["isPending"],
		"is_sitemaps_index": item["isSitemapsIndex"],
		"type":              item["type"],
		"warnings":          stringify(item["warnings"]),
		"errors":            stringify(item["errors"]),
	}
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		return fmt.Sprintf("%v", t)
	}
}

// analyticsRequestBody is the JSON body for a searchAnalytics/query call,
// ported verbatim from legacy's analyticsRequest struct.
type analyticsRequestBody struct {
	StartDate  string   `json:"startDate"`
	EndDate    string   `json:"endDate"`
	Dimensions []string `json:"dimensions"`
	SearchType string   `json:"type,omitempty"`
	DataState  string   `json:"dataState,omitempty"`
	RowLimit   int      `json:"rowLimit"`
	StartRow   int      `json:"startRow"`
}

// readAnalyticsForSite drives one site's searchAnalytics/query pagination
// loop, ported verbatim from legacy's readAnalytics
// (google_search_console.go:207-263): the next page is requested by
// advancing startRow by the number of rows received, until a short (or
// empty) page is returned, or maxPages (0 = unbounded) is reached.
func (h *Hooks) readAnalyticsForSite(ctx context.Context, r *connsdk.Requester, site string, dims []string, startDate, endDate, searchType, dataState string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := "/sites/" + url.PathEscape(site) + "/searchAnalytics/query"
	startRow := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := analyticsRequestBody{
			StartDate:  startDate,
			EndDate:    endDate,
			Dimensions: dims,
			SearchType: searchType,
			DataState:  dataState,
			RowLimit:   pageSize,
			StartRow:   startRow,
		}
		resp, err := r.Do(ctx, http.MethodPost, path, nil, body)
		if err != nil {
			return fmt.Errorf("read google-search-console searchAnalytics for %s: %w", site, err)
		}
		rows, err := connsdk.RecordsAt(resp.Body, "rows")
		if err != nil {
			return fmt.Errorf("decode google-search-console searchAnalytics: %w", err)
		}
		for _, row := range rows {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(analyticsRecord(site, searchType, dims, row)); err != nil {
				return err
			}
		}
		if len(rows) < pageSize {
			return nil
		}
		startRow += len(rows)
	}
	return nil
}

// analyticsRecord flattens a single searchAnalytics row, ported verbatim
// from legacy's analyticsRecord (streams.go): the row's keys array is
// positionally aligned with the requested dimensions; metric fields
// (clicks, impressions, ctr, position) come from the row body.
func analyticsRecord(site, searchType string, dimensions []string, row map[string]any) connectors.Record {
	rec := connectors.Record{
		"site_url":    site,
		"search_type": searchType,
	}
	if keys, ok := row["keys"].([]any); ok {
		for i, dim := range dimensions {
			if i < len(keys) {
				rec[dim] = keys[i]
			}
		}
	}
	rec["clicks"] = row["clicks"]
	rec["impressions"] = row["impressions"]
	rec["ctr"] = row["ctr"]
	rec["position"] = row["position"]
	return rec
}

// analyticsDateRange resolves the [startDate, endDate] window for a search
// analytics query, ported verbatim from legacy's analyticsDateRange: the
// incremental cursor (a previously-synced date) or start_date config sets
// the lower bound; end_date config or today (UTC) sets the upper bound.
func analyticsDateRange(req connectors.ReadRequest) (string, string, error) {
	start := connsdk.Cursor(req.State)
	if start == "" {
		start = strings.TrimSpace(req.Config.Config["start_date"])
	}
	if start == "" {
		start = "2021-01-01"
	}
	if _, err := time.Parse("2006-01-02", start); err != nil {
		return "", "", fmt.Errorf("google-search-console start_date must be YYYY-MM-DD: %w", err)
	}
	end := strings.TrimSpace(req.Config.Config["end_date"])
	if end == "" {
		end = time.Now().UTC().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", end); err != nil {
		return "", "", fmt.Errorf("google-search-console end_date must be YYYY-MM-DD: %w", err)
	}
	return start, end, nil
}

// siteURLs returns the configured site URL property list (comma- or
// newline-separated), matching the fan_out config_key split every
// declarative stream in this bundle already uses.
func siteURLs(cfg connectors.RuntimeConfig) []string {
	raw := strings.TrimSpace(cfg.Config["site_urls"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["site_url"])
	}
	if raw == "" {
		return nil
	}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if v := strings.TrimSpace(f); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func gscSearchType(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["search_type"]); v != "" {
		return v
	}
	return "web"
}

func gscPageSize(cfg connectors.RuntimeConfig) (int, error) {
	const defaultPageSize = 25000
	const maxPageSize = 25000
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-search-console config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("google-search-console config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func gscMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-search-console config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, fmt.Errorf("google-search-console config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}
