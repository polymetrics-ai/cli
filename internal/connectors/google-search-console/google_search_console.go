// Package googlesearchconsole implements the native pm Google Search Console
// connector. It follows the declarative-HTTP template established by the stripe
// connector: a thin package composing the connsdk toolkit (Requester + Bearer
// auth + RecordsAt extraction + cursor state) with Search Console-specific stream
// definitions, endpoints, and pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The connector is read-only: the Search Console API has no safe reverse-ETL
// write surface, so Capabilities.Write is false and Write returns
// ErrUnsupportedOperation.
package googlesearchconsole

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	gscName            = "google-search-console"
	gscDefaultBaseURL  = "https://www.googleapis.com/webmasters/v3"
	gscDefaultPageSize = 25000
	gscMaxPageSize     = 25000
	gscUserAgent       = "polymetrics-go-cli"
	// gscFixtureDate is the deterministic date used by fixture-mode records.
	gscFixtureDate = "2026-01-01"
)

func init() {
	connectors.RegisterFactory(gscName, New)
}

// New returns the Google Search Console connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Google Search Console connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return gscName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            gscName,
		DisplayName:     "Google Search Console",
		IntegrationType: "api",
		Description:     "Reads Google Search Console sites, sitemaps, and Search Analytics performance reports (by date, query, page, country, and device) through the Search Console v3 REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Search
// Console. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gscBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(gscAccessToken(cfg)) == "" {
		return errors.New("google-search-console connector requires secret authorization.access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the sites list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "sites", nil, nil, nil); err != nil {
		return fmt.Errorf("check google-search-console: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gscStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: search-analytics streams start
// with an empty incremental cursor (the date the data is keyed on), which the
// start_date config raises at read time.
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
		stream = "sites"
	}
	def, ok := gscStreamDefs[stream]
	if !ok {
		return fmt.Errorf("google-search-console stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, def, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	switch def.kind {
	case metaStream:
		return c.readMeta(ctx, r, def, req, emit)
	case analyticsStream:
		return c.readAnalytics(ctx, r, def, req, emit)
	default:
		return fmt.Errorf("google-search-console stream %q has unknown kind", stream)
	}
}

// readMeta reads a GET list endpoint. The sites stream is account-scoped; sitemaps
// are read per configured site URL. Neither endpoint paginates.
func (c Connector) readMeta(ctx context.Context, r *connsdk.Requester, def streamDef, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if !def.perSite {
		resp, err := r.Do(ctx, http.MethodGet, def.resource, nil, nil)
		if err != nil {
			return fmt.Errorf("read google-search-console %s: %w", def.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, def.recordsPath)
		if err != nil {
			return fmt.Errorf("decode google-search-console %s: %w", def.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(def.mapMeta(item)); err != nil {
				return err
			}
		}
		return nil
	}

	for _, site := range siteURLs(req.Config) {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := "sites/" + url.PathEscape(site) + "/" + def.resource
		resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
		if err != nil {
			return fmt.Errorf("read google-search-console %s for %s: %w", def.resource, site, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, def.recordsPath)
		if err != nil {
			return fmt.Errorf("decode google-search-console %s: %w", def.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			rec := def.mapMeta(item)
			rec["site_url"] = site
			if err := emit(rec); err != nil {
				return err
			}
		}
	}
	return nil
}

// analyticsRequest is the JSON body for a searchAnalytics/query call.
type analyticsRequest struct {
	StartDate  string   `json:"startDate"`
	EndDate    string   `json:"endDate"`
	Dimensions []string `json:"dimensions"`
	SearchType string   `json:"type,omitempty"`
	DataState  string   `json:"dataState,omitempty"`
	RowLimit   int      `json:"rowLimit"`
	StartRow   int      `json:"startRow"`
}

// readAnalytics drives the searchAnalytics/query endpoint. The API returns up to
// rowLimit rows; the next page is requested by advancing startRow by the number
// of rows received, until a short (or empty) page is returned. It loops over each
// configured site URL.
func (c Connector) readAnalytics(ctx context.Context, r *connsdk.Requester, def streamDef, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	startDate, endDate, err := analyticsDateRange(req)
	if err != nil {
		return err
	}
	pageSize, err := gscPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gscMaxPages(req.Config)
	if err != nil {
		return err
	}
	searchType := gscSearchType(req.Config)
	dataState := strings.TrimSpace(req.Config.Config["data_state"])

	for _, site := range siteURLs(req.Config) {
		path := "sites/" + url.PathEscape(site) + "/searchAnalytics/query"
		startRow := 0
		for page := 0; maxPages == 0 || page < maxPages; page++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			body := analyticsRequest{
				StartDate:  startDate,
				EndDate:    endDate,
				Dimensions: def.dimensions,
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
				rec := analyticsRecord(site, searchType, def.dimensions, row)
				if err := emit(rec); err != nil {
					return err
				}
			}
			if len(rows) < pageSize {
				break
			}
			startRow += len(rows)
		}
	}
	return nil
}

// analyticsRecord flattens a single searchAnalytics row. The row's keys array is
// positionally aligned with the requested dimensions; metric fields (clicks,
// impressions, ctr, position) come from the row body.
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

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, def streamDef, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	site := "https://example.com/"
	if urls := siteURLs(req.Config); len(urls) > 0 {
		site = urls[0]
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var rec connectors.Record
		switch def.kind {
		case metaStream:
			if def.resource == "sites" {
				rec = connectors.Record{
					"site_url":         fmt.Sprintf("https://fixture-%d.example.com/", i),
					"permission_level": "siteOwner",
				}
			} else {
				rec = connectors.Record{
					"site_url":          site,
					"path":              fmt.Sprintf("https://example.com/sitemap-%d.xml", i),
					"last_submitted":    gscFixtureDate + "T00:00:00Z",
					"last_downloaded":   gscFixtureDate + "T01:00:00Z",
					"is_pending":        false,
					"is_sitemaps_index": false,
					"type":              "sitemap",
				}
			}
		default: // analyticsStream
			rec = connectors.Record{
				"site_url":    site,
				"search_type": gscSearchType(req.Config),
				"date":        fmt.Sprintf("2026-01-%02d", i),
				"clicks":      10 * i,
				"impressions": 100 * i,
				"ctr":         0.1,
				"position":    float64(i) + 1.5,
			}
			for _, dim := range def.dimensions {
				if dim == "date" {
					continue
				}
				rec[dim] = fmt.Sprintf("fixture_%s_%d", dim, i)
			}
		}
		if cursor := req.State["cursor"]; cursor != "" {
			rec["previous_cursor"] = cursor
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// Write is unsupported: the connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with Bearer auth (the OAuth access
// token) and the resolved base URL. The secret only ever flows into connsdk.Bearer;
// it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := gscBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := gscAccessToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("google-search-console connector requires secret authorization.access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(token),
		UserAgent: gscUserAgent,
	}, nil
}

// analyticsDateRange resolves the [startDate, endDate] window for a search
// analytics query. The incremental cursor (a date) or start_date config sets the
// lower bound; end_date config or today sets the upper bound.
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

// siteURLs returns the configured site URL property list. It accepts a comma- or
// newline-separated site_urls config value (the catalog declares it an array).
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

func gscAccessToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Secrets["authorization.access_token"]); v != "" {
		return v
	}
	// Tolerate a bare key for convenience in tooling that flattens the path.
	return strings.TrimSpace(cfg.Secrets["access_token"])
}

// gscBaseURL resolves and validates the base URL. The default is
// www.googleapis.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func gscBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return gscDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("google-search-console config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("google-search-console config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("google-search-console config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func gscPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return gscDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-search-console config page_size must be an integer: %w", err)
	}
	if value < 1 || value > gscMaxPageSize {
		return 0, fmt.Errorf("google-search-console config page_size must be between 1 and %d", gscMaxPageSize)
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
		return 0, errors.New("google-search-console config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
