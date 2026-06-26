// Package gnews implements the native pm GNews connector. It is a declarative-
// HTTP per-system connector built on the same shape as the stripe reference:
// a thin package that composes the connsdk toolkit (Requester + APIKeyQuery auth
// + RecordsAt extraction + page-number pagination) with GNews-specific stream
// definitions and endpoints.
//
// GNews (https://gnews.io) is a read-only news search API; there are no
// reverse-ETL writes that make sense, so the connector advertises Write=false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package gnews

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	gnewsDefaultBaseURL  = "https://gnews.io/api/v4"
	gnewsDefaultPageSize = 10  // GNews "max" default
	gnewsMaxPageSize     = 100 // GNews "max" ceiling
	gnewsUserAgent       = "polymetrics-go-cli"
	// gnewsAPIDateLayout is the timestamp format GNews accepts for from/to.
	gnewsAPIDateLayout = "2006-01-02T15:04:05Z"
)

func init() {
	connectors.RegisterFactory("gnews", New)
}

// New returns the GNews connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm GNews connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "gnews" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "gnews",
		DisplayName:     "GNews",
		IntegrationType: "api",
		Description:     "Reads GNews articles from the keyword search and top-headlines endpoints of the GNews REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to GNews. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gnewsBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(gnewsSecret(cfg)) == "" {
		return errors.New("gnews connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the search endpoint confirms auth and connectivity.
	// GNews search requires a query, so a harmless default is used when none is
	// configured.
	query := gnewsBaseQuery("search", cfg)
	query.Set("max", "1")
	if err := r.DoJSON(ctx, http.MethodGet, "search", query, nil, nil); err != nil {
		return fmt.Errorf("check gnews: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gnewsStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a GNews stream starts with
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
		stream = "search"
	}
	endpoint, ok := gnewsStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("gnews stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := gnewsPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gnewsMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := gnewsBaseQuery(stream, req.Config)
	base.Set("max", strconv.Itoa(pageSize))
	if from, err := incrementalLowerBound(req); err != nil {
		return err
	} else if from != "" {
		base.Set("from", from)
	}
	if to := gnewsConfigDate(req.Config, "end_date"); to != "" {
		base.Set("to", to)
	}

	// GNews paginates by ?page=N (1-based) with ?max=<page size>; a page shorter
	// than max signals the end. PageNumberPaginator + Harvest implement exactly
	// this shape, extracting records at the "articles" path.
	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		StartPage: 1,
		PageSize:  pageSize,
	}
	mapped := func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(map[string]any(rec)))
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, "articles", maxPages, mapped)
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise gnews credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"title":       fmt.Sprintf("Fixture article %d", i),
			"description": fmt.Sprintf("Fixture description %d", i),
			"content":     fmt.Sprintf("Fixture content %d", i),
			"url":         fmt.Sprintf("https://example.com/%s/%d", endpoint.resource, i),
			"image":       fmt.Sprintf("https://example.com/%s/%d.jpg", endpoint.resource, i),
			"publishedAt": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"lang":        "en",
			"source": map[string]any{
				"id":      "fixture-source",
				"name":    "Fixture News",
				"url":     "https://example.com",
				"country": "us",
			},
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "gnews"
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

// requester builds a connsdk.Requester wired with APIKeyQuery auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery as the
// apikey query parameter; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := gnewsBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := gnewsSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("gnews connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("apikey", secret),
		UserAgent: gnewsUserAgent,
	}, nil
}

// gnewsBaseQuery assembles the non-pagination filter params shared by every page
// of a stream from config: q/topic, lang, country, in, nullable, sortby.
func gnewsBaseQuery(stream string, cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	config := cfg.Config
	if config == nil {
		config = map[string]string{}
	}

	switch stream {
	case "top_headlines":
		if topic := strings.TrimSpace(config["top_headlines_topic"]); topic != "" {
			q.Set("topic", topic)
		}
		if query := firstNonEmpty(config["top_headlines_query"], config["query"]); query != "" {
			q.Set("q", query)
		}
	default: // search
		query := strings.TrimSpace(config["query"])
		if query == "" {
			// GNews search requires a query; default to a harmless broad term so
			// check/read still function when none is configured.
			query = "news"
		}
		q.Set("q", query)
	}

	if lang := strings.TrimSpace(config["language"]); lang != "" {
		q.Set("lang", lang)
	}
	if country := strings.TrimSpace(config["country"]); country != "" {
		q.Set("country", country)
	}
	if in := strings.TrimSpace(config["in"]); in != "" {
		q.Set("in", in)
	}
	if nullable := strings.TrimSpace(config["nullable"]); nullable != "" {
		q.Set("nullable", nullable)
	}
	if sortby := strings.TrimSpace(config["sortby"]); sortby != "" {
		q.Set("sortby", sortby)
	}
	return q
}

// incrementalLowerBound returns the GNews-formatted lower bound for the `from`
// filter, derived from the incremental cursor (if any) or else the start_date
// config. An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return normalizeGNewsDate(cursor)
	}
	return gnewsConfigDate(req.Config, "start_date"), nil
}

// gnewsConfigDate reads a config date field and renders it in the GNews API
// format. It accepts both the catalog "YYYY-MM-DD hh:mm:ss" form and RFC3339.
// Invalid values are dropped (returns "") rather than failing the read.
func gnewsConfigDate(cfg connectors.RuntimeConfig, key string) string {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return ""
	}
	out, err := normalizeGNewsDate(raw)
	if err != nil {
		return ""
	}
	return out
}

// normalizeGNewsDate parses a timestamp in one of the accepted layouts and
// renders it in the format GNews expects (RFC3339 UTC).
func normalizeGNewsDate(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	for _, layout := range []string{time.RFC3339, gnewsAPIDateLayout, "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC().Format(gnewsAPIDateLayout), nil
		}
	}
	return "", fmt.Errorf("gnews date %q is not a recognized format", raw)
}

func gnewsSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// gnewsBaseURL resolves and validates the base URL. The default is gnews.io; any
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func gnewsBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return gnewsDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("gnews config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("gnews config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("gnews config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func gnewsPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return gnewsDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gnews config page_size must be an integer: %w", err)
	}
	if value < 1 || value > gnewsMaxPageSize {
		return 0, fmt.Errorf("gnews config page_size must be between 1 and %d", gnewsMaxPageSize)
	}
	return value, nil
}

func gnewsMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gnews config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("gnews config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// Write is unsupported: GNews is a read-only news search API.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
