// Package searxng implements the native pm SearXNG connector. SearXNG
// (https://docs.searxng.org) is a self-hostable metasearch engine that exposes a
// JSON search API: GET {base_url}/search?q=<query>&format=json&pageno=<n>. It is
// the pm-native way to pull web/Reddit search results into the warehouse without
// any per-tool MCP server.
//
// The connector is a declarative-HTTP source built on the connsdk toolkit, the
// same shape as the gnews/stripe references: a thin package composing a Requester
// (no auth by default; optional Bearer for instances behind an auth proxy) with
// page-number pagination over the results[] array and SearXNG-specific stream
// definitions.
//
// SearXNG instances vary in which engines are installed, so the "reddit"
// convenience stream scopes the query with site:reddit.com (optionally a
// subreddit) rather than depending on a reddit engine being present. This keeps
// Reddit extraction working against any general instance.
//
// It is read-only (Write=false) and self-registers via RegisterFactory in init();
// registryset blank-imports this package in the production binary. SearXNG is not
// a catalog_data.json entry — it is a pm-native registry connector.
package searxng

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
	searxngDefaultPageSize = 10
	searxngMaxPageSize     = 100
	searxngDefaultMaxPages = 1 // SearXNG paging is fuzzy/engine-dependent; bound by default.
	searxngUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("searxng", New)
	// searxng has no catalog_data.json entry, so opt it into the live registry
	// the CLI uses.
	connectors.RegisterNativeLive("searxng")
}

// New returns the SearXNG connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm SearXNG connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "searxng" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "searxng",
		DisplayName:     "SearXNG",
		IntegrationType: "api",
		Description:     "Reads web and Reddit search results from a SearXNG metasearch instance's JSON API (format=json). Read-only. Requires base_url; no credentials by default.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to SearXNG. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	base, err := searxngBaseURL(cfg)
	if err != nil {
		return err
	}
	r := c.requester(base, cfg)
	q := url.Values{}
	q.Set("q", "polymetrics")
	q.Set("format", "json")
	if err := r.DoJSON(ctx, http.MethodGet, "search", q, nil, nil); err != nil {
		return fmt.Errorf("check searxng: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: searxngStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "search"
	}
	if _, ok := searxngStreamSet[stream]; !ok {
		return fmt.Errorf("searxng stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	base, err := searxngBaseURL(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := searxngPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := searxngMaxPages(req.Config)
	if err != nil {
		return err
	}
	query, err := searxngQuery(stream, req.Config)
	if err != nil {
		return err
	}

	r := c.requester(base, req.Config)
	baseQuery := searxngBaseQuery(query, req.Config)

	// SearXNG paginates by ?pageno=N (1-based). Result count per page is
	// engine-defined, so PageSize is used only as the short-page stop threshold;
	// no per-page size param is sent (SizeParam left empty).
	paginator := &connsdk.PageNumberPaginator{
		PageParam: "pageno",
		StartPage: 1,
		PageSize:  pageSize,
	}
	mapped := func(rec connsdk.Record) error {
		return emit(searxngResultRecord(stream, map[string]any(rec)))
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, "search", baseQuery, paginator, "results", maxPages, mapped)
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise searxng credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"url":           fmt.Sprintf("https://www.reddit.com/r/dataengineering/fixture/%d", i),
			"title":         fmt.Sprintf("Fixture result %d", i),
			"content":       fmt.Sprintf("Fixture content %d", i),
			"engine":        "fixture",
			"engines":       []any{"fixture"},
			"score":         float64(i),
			"category":      "general",
			"publishedDate": fmt.Sprintf("2026-01-0%dT00:00:00", i),
		}
		record := searxngResultRecord(stream, item)
		record["connector"] = "searxng"
		record["fixture"] = true
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester. SearXNG public instances are open, so
// auth is nil unless an optional api_key secret is supplied (for instances behind
// an auth proxy), in which case it is sent as a Bearer token.
func (c Connector) requester(base string, cfg connectors.RuntimeConfig) *connsdk.Requester {
	var auth connsdk.Authenticator
	if token := searxngSecret(cfg); strings.TrimSpace(token) != "" {
		auth = connsdk.Bearer(token)
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: searxngUserAgent,
	}
}

// searxngBaseQuery assembles the per-request params shared by every page:
// q, the mandatory format=json, and optional passthrough filters.
func searxngBaseQuery(query string, cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	q.Set("q", query)
	q.Set("format", "json")
	config := cfg.Config
	if config == nil {
		config = map[string]string{}
	}
	for cfgKey, param := range map[string]string{
		"categories": "categories",
		"engines":    "engines",
		"language":   "language",
		"time_range": "time_range",
		"safesearch": "safesearch",
	} {
		if v := strings.TrimSpace(config[cfgKey]); v != "" {
			q.Set(param, v)
		}
	}
	return q
}

// searxngQuery resolves the effective search query for a stream. The "reddit"
// stream scopes the query to reddit.com (optionally a subreddit) so any general
// instance returns Reddit results.
func searxngQuery(stream string, cfg connectors.RuntimeConfig) (string, error) {
	config := cfg.Config
	if config == nil {
		config = map[string]string{}
	}
	terms := strings.TrimSpace(config["query"])
	if stream == "reddit" {
		site := "site:reddit.com"
		if sub := strings.TrimSpace(config["subreddit"]); sub != "" {
			site = "site:reddit.com/r/" + strings.TrimPrefix(sub, "r/")
		}
		return strings.TrimSpace(site + " " + terms), nil
	}
	if terms == "" {
		return "", errors.New("searxng search stream requires config query")
	}
	return terms, nil
}

func searxngSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// searxngBaseURL resolves and validates the base URL. There is no default: a
// SearXNG instance must be named explicitly. Any value must be an absolute
// http/https URL with a host to bound SSRF risk.
func searxngBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return "", errors.New("searxng connector requires config base_url (your SearXNG instance)")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("searxng config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("searxng config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("searxng config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func searxngPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return searxngDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("searxng config page_size must be an integer: %w", err)
	}
	if value < 1 || value > searxngMaxPageSize {
		return 0, fmt.Errorf("searxng config page_size must be between 1 and %d", searxngMaxPageSize)
	}
	return value, nil
}

func searxngMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return searxngDefaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("searxng config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("searxng config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: SearXNG is a read-only search API.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
