// Package newsdata implements the native pm NewsData.io connector. It follows
// the declarative-HTTP per-system connector shape established by the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// APIKeyQuery auth + RecordsAt extraction + body-token pagination) with
// NewsData.io-specific stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect. The connector is read-only: NewsData.io is a news search
// API with no reverse-ETL write surface.
package newsdata

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
	newsdataDefaultMaxPages = 5
	newsdataUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("newsdata", New)
}

// New returns the NewsData.io connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm NewsData.io connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "newsdata" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "newsdata",
		DisplayName:     "Newsdata",
		IntegrationType: "api",
		Description:     "Reads latest news, cryptocurrency news, and news sources from the NewsData.io REST API.",
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
		return errors.New("newsdata connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the latest endpoint confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "latest", url.Values{"size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check newsdata: %w", err)
	}
	return nil
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
		return fmt.Errorf("newsdata stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := newsdataMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, req.Config, maxPages, emit)
}

// harvest drives NewsData.io's nextPage token pagination. List responses are
// shaped {results:[...], nextPage:"<token>"|null}; the next page is requested
// with page=<token>. The sources endpoint is unpaginated (no nextPage), so the
// loop terminates after the first page for it.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, maxPages int, emit func(connectors.Record) error) error {
	base := newsdataFilters(cfg)

	page := ""
	for pageNum := 0; maxPages == 0 || pageNum < maxPages; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if page != "" {
			query.Set("page", page)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read newsdata %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode newsdata %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if !endpoint.paginated {
			return nil
		}
		next, err := connsdk.StringAt(resp.Body, "nextPage")
		if err != nil {
			return fmt.Errorf("decode newsdata %s nextPage: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		page = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise newsdata credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		if stream == "sources" {
			item = map[string]any{
				"id":       fmt.Sprintf("source_fixture_%d", i),
				"name":     fmt.Sprintf("Fixture Source %d", i),
				"url":      fmt.Sprintf("https://example.com/source%d", i),
				"icon":     "",
				"category": []any{"top"},
				"language": []any{"en"},
				"country":  []any{"united states"},
			}
		} else {
			item = map[string]any{
				"article_id":      fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
				"title":           fmt.Sprintf("Fixture Article %d", i),
				"link":            fmt.Sprintf("https://example.com/%s/%d", endpoint.resource, i),
				"description":     "Deterministic fixture article.",
				"content":         "Deterministic fixture content.",
				"pubDate":         fmt.Sprintf("2026-01-0%d 00:00:00", i),
				"source_id":       "fixture",
				"source_priority": int64(i),
				"language":        "english",
				"creator":         []any{"Fixture Author"},
				"keywords":        []any{"fixture"},
				"category":        []any{"top"},
				"country":         []any{"united states"},
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

// requester builds a connsdk.Requester wired with apikey-query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := newsdataBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := newsdataSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("newsdata connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("apikey", secret),
		UserAgent: newsdataUserAgent,
	}, nil
}

// newsdataFilters builds the common query params shared by every page from the
// connector config: comma-joined category/country/language/domain restrictions,
// the search query, and optional page size. Empty values are omitted.
func newsdataFilters(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	for cfgKey, param := range map[string]string{
		"category": "category",
		"country":  "country",
		"language": "language",
		"domain":   "domain",
	} {
		if v := strings.TrimSpace(cfg.Config[cfgKey]); v != "" {
			q.Set(param, v)
		}
	}
	if v := strings.TrimSpace(cfg.Config["query"]); v != "" {
		q.Set("q", v)
	}
	if v := strings.TrimSpace(cfg.Config["query_in_title"]); v != "" {
		q.Set("qInTitle", v)
	}
	if v := strings.TrimSpace(cfg.Config["size"]); v != "" {
		q.Set("size", v)
	}
	return q
}

func newsdataSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// newsdataBaseURL resolves and validates the base URL. The default is
// newsdata.io; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func newsdataBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return newsdataDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("newsdata config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("newsdata config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("newsdata config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// newsdataMaxPages bounds how many pages a read fetches. The default is small
// because NewsData.io free tiers cap requests per day; "all"/"unlimited"/0 lifts
// the bound.
func newsdataMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return newsdataDefaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("newsdata config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("newsdata config max_pages must be 0 for unlimited or a positive integer")
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

// Write satisfies the connectors.Connector interface. NewsData.io is a read-only
// news search API with no reverse-ETL surface, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
