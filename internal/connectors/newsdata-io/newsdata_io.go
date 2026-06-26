// Package newsdataio implements the native pm NewsData.io connector. It follows
// the declarative-HTTP template established by the stripe connector: a thin
// package that composes the connsdk toolkit (Requester + apikey query auth +
// RecordsAt extraction + nextPage body-token pagination) with NewsData.io stream
// definitions and endpoints.
//
// NewsData.io is a read-only news feed API, so the connector exposes Check,
// Catalog, and Read only (no reverse-ETL writes). It self-registers with the
// connectors registry via RegisterFactory in init(); the registryset package
// blank-imports this package in the production binary to run that side effect.
package newsdataio

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
	newsdataDefaultBaseURL  = "https://newsdata.io/api/1"
	newsdataDefaultPageSize = 10
	newsdataMaxPageSize     = 50
	newsdataUserAgent       = "polymetrics-go-cli"
	// newsdataFixturePubDate is the deterministic pubDate used by fixture records.
	newsdataFixturePubDate = "2026-01-01 00:00:00"
)

func init() {
	connectors.RegisterFactory("newsdata-io", New)
}

// New returns the NewsData.io connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm NewsData.io connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "newsdata-io" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "newsdata-io",
		DisplayName:     "NewsData.io",
		IntegrationType: "api",
		Description:     "Reads latest, crypto, and archived news articles plus available news sources from the NewsData.io REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to NewsData.io.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := newsdataBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(newsdataSecret(cfg)) == "" {
		return errors.New("newsdata-io connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the latest endpoint confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "latest", url.Values{"size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check newsdata-io: %w", err)
	}
	return nil
}

// Write is unsupported: NewsData.io is a read-only news feed with no sensible
// reverse-ETL target, so the connector advertises Write=false and returns the
// shared unsupported-operation error.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: newsdataStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "latest"
	}
	endpoint, ok := newsdataStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("newsdata-io stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := newsdataPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := newsdataMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, stream, endpoint, pageSize, maxPages, req.Config, emit)
}

// harvest drives NewsData.io's nextPage body-token pagination. Responses look like
// {status, totalResults, results:[...], nextPage:"token"}; the next page is
// requested with page=<nextPage>. The sources endpoint is not paginated. The loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, stream string, endpoint streamEndpoint, pageSize, maxPages int, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	base := newsdataQueryFilters(cfg)
	if stream != "sources" && pageSize > 0 {
		base.Set("size", strconv.Itoa(pageSize))
	}

	nextPage := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if nextPage != "" {
			query.Set("page", nextPage)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read newsdata-io %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode newsdata-io %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// The sources endpoint is a single page; only article streams paginate.
		if stream == "sources" {
			return nil
		}
		token, err := connsdk.StringAt(resp.Body, "nextPage")
		if err != nil {
			return fmt.Errorf("decode newsdata-io %s nextPage: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(token) == "" {
			return nil
		}
		nextPage = token
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise newsdata-io credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		if stream == "sources" {
			item = map[string]any{
				"id":          fmt.Sprintf("source_fixture_%d", i),
				"name":        fmt.Sprintf("Fixture Source %d", i),
				"url":         fmt.Sprintf("https://example.com/source/%d", i),
				"icon":        "https://example.com/icon.png",
				"description": "Fixture news source.",
				"priority":    int64(i),
				"language":    []any{"english"},
				"category":    []any{"top"},
				"country":     []any{"united states of america"},
			}
		} else {
			item = map[string]any{
				"article_id":  fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
				"title":       fmt.Sprintf("Fixture Article %d", i),
				"link":        fmt.Sprintf("https://example.com/article/%d", i),
				"description": "Fixture article description.",
				"content":     "Fixture article content.",
				"pubDate":     newsdataFixturePubDate,
				"image_url":   "https://example.com/image.png",
				"source_id":   "fixture_source",
				"source_name": "Fixture Source",
				"source_url":  "https://example.com",
				"language":    "english",
				"creator":     []any{"Fixture Author"},
				"keywords":    []any{"fixture", "news"},
				"category":    []any{"top"},
				"country":     []any{"united states of america"},
				"duplicate":   false,
			}
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

// requester builds a connsdk.Requester wired with apikey query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := newsdataBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := newsdataSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("newsdata-io connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("apikey", secret),
		UserAgent: newsdataUserAgent,
	}, nil
}

// newsdataQueryFilters builds the optional article filter query from config.
// These map directly to NewsData.io's documented query parameters. Comma-joined
// list values (categories, countries, languages, domains) are passed through.
func newsdataQueryFilters(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	if cfg.Config == nil {
		return q
	}
	setIf := func(param, key string) {
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			q.Set(param, v)
		}
	}
	setIf("q", "search_query")
	setIf("category", "categories")
	setIf("country", "countries")
	setIf("language", "languages")
	setIf("domain", "domains")
	// Archive endpoint accepts from_date/to_date; pass start/end through when set.
	setIf("from_date", "start_date")
	setIf("to_date", "end_date")
	return q
}

func newsdataSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// newsdataBaseURL resolves and validates the base URL. The default is
// newsdata.io/api/1; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func newsdataBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return newsdataDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("newsdata-io config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("newsdata-io config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("newsdata-io config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func newsdataPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return newsdataDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("newsdata-io config page_size must be an integer: %w", err)
	}
	if value < 1 || value > newsdataMaxPageSize {
		return 0, fmt.Errorf("newsdata-io config page_size must be between 1 and %d", newsdataMaxPageSize)
	}
	return value, nil
}

func newsdataMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		// NewsData.io paginates indefinitely over a large corpus; bound the
		// default to a sane number of pages so a full sync terminates.
		return 10, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("newsdata-io config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("newsdata-io config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
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
