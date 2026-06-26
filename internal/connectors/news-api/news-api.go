// Package newsapi implements the native pm News API (newsapi.org) connector. It
// is a read-only declarative-HTTP per-system connector that composes the connsdk
// toolkit (Requester + X-Api-Key header auth + RecordsAt extraction) with News
// API stream definitions, endpoints, and page/pageSize pagination.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
//
// The directory is named "news-api" (with a hyphen) to match the bare system
// name; the Go package identifier is the sanitized "newsapi".
package newsapi

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
	defaultBaseURL  = "https://newsapi.org/v2"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("news-api", New)
}

// New returns the News API connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm News API connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "news-api" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "news-api",
		DisplayName:     "News API",
		IntegrationType: "api",
		Description:     "Reads articles and news sources from the News API (newsapi.org): the everything search, top headlines, and the sources directory.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the News API.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("news-api connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the sources directory confirms auth and connectivity
	// without any heavy query.
	if err := r.DoJSON(ctx, http.MethodGet, "top-headlines/sources", nil, nil, nil); err != nil {
		return fmt.Errorf("check news-api: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an article stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "everything"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("news-api stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	pages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	base, err := queryParams(stream, req)
	if err != nil {
		return err
	}

	if !endpoint.paginated {
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, base, nil)
		if err != nil {
			return fmt.Errorf("read news-api %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode news-api %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		return nil
	}

	// page/pageSize pagination: stop on a short page (the page-number paginator
	// convention) or when maxPages is reached.
	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "pageSize",
		StartPage: 1,
		PageSize:  size,
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, endpoint.recordsPath, pages, func(item connsdk.Record) error {
		return emit(endpoint.mapRecord(item))
	})
}

// Write is unsupported: news-api is a read-only source connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise news-api credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		if stream == "sources" {
			item = map[string]any{
				"id":          fmt.Sprintf("fixture-source-%d", i),
				"name":        fmt.Sprintf("Fixture Source %d", i),
				"description": "Deterministic fixture source.",
				"url":         fmt.Sprintf("https://example.com/source/%d", i),
				"category":    "technology",
				"language":    "en",
				"country":     "us",
			}
		} else {
			item = map[string]any{
				"source":      map[string]any{"id": fmt.Sprintf("fixture-source-%d", i), "name": fmt.Sprintf("Fixture Source %d", i)},
				"author":      fmt.Sprintf("Author %d", i),
				"title":       fmt.Sprintf("Fixture Headline %d", i),
				"description": "Deterministic fixture article.",
				"url":         fmt.Sprintf("https://example.com/article/%d", i),
				"urlToImage":  fmt.Sprintf("https://example.com/article/%d.jpg", i),
				"publishedAt": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
				"content":     fmt.Sprintf("Fixture content %d.", i),
			}
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-Api-Key header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := secret(cfg)
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("news-api connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-Api-Key", key, ""),
		UserAgent: userAgent,
	}, nil
}

// queryParams builds the per-stream request filters from config. The News API
// requires at least one of q/sources/domains for /everything, and country or
// category for /top-headlines; here we forward whatever the operator configured.
func queryParams(stream string, req connectors.ReadRequest) (url.Values, error) {
	cfg := req.Config.Config
	q := url.Values{}
	set := func(param, key string) {
		if v := strings.TrimSpace(cfg[key]); v != "" {
			q.Set(param, v)
		}
	}
	switch stream {
	case "everything":
		set("q", "search_query")
		set("searchIn", "search_in")
		set("sources", "sources")
		set("domains", "domains")
		set("excludeDomains", "exclude_domains")
		set("language", "language")
		set("sortBy", "sort_by")
		set("from", "start_date")
		set("to", "end_date")
		// Incremental cursor (publishedAt) raises the lower bound on resync.
		if cursor := connsdk.Cursor(req.State); cursor != "" {
			q.Set("from", cursor)
		}
	case "top_headlines":
		set("q", "search_query")
		set("country", "country")
		set("category", "category")
		set("sources", "sources")
		set("language", "language")
	case "sources":
		set("category", "category")
		set("language", "language")
		set("country", "country")
	}
	return q, nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// baseURL resolves and validates the base URL. The default is newsapi.org/v2; any
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("news-api config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("news-api config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("news-api config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("news-api config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("news-api config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("news-api config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("news-api config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
