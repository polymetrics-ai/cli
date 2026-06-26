// Package giphy implements the native pm Giphy connector. It is a declarative-HTTP
// per-system connector following the stripe template: a thin package composing
// the connsdk toolkit (Requester + APIKeyQuery auth + RecordsAt extraction +
// offset/limit pagination) with Giphy-specific stream definitions and endpoints.
//
// The Giphy API is read-only search/trending media discovery; there is no
// sensible reverse-ETL write target, so the connector is read-only
// (Capabilities.Write=false).
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package giphy

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
	giphyDefaultBaseURL  = "https://api.giphy.com/v1"
	giphyDefaultPageSize = 25
	giphyMaxPageSize     = 50 // Giphy caps limit at 50 per request.
	giphyUserAgent       = "polymetrics-go-cli"
	// giphyFixtureImport is the deterministic import_datetime used by fixture
	// records.
	giphyFixtureImport = "2026-01-01 00:00:00"
)

func init() {
	connectors.RegisterFactory("giphy", New)
}

// New returns the Giphy connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Giphy connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "giphy" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "giphy",
		DisplayName:     "Giphy",
		IntegrationType: "api",
		Description:     "Reads GIFs, stickers, and clips from the Giphy search and trending REST endpoints. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Giphy. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := giphyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(giphySecret(cfg)) == "" {
		return errors.New("giphy connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of trending GIFs confirms the api_key and connectivity
	// without requiring a search query.
	query := url.Values{"limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "gifs/trending", query, nil, nil); err != nil {
		return fmt.Errorf("check giphy: %w", err)
	}
	return nil
}

// Write is unsupported: the Giphy API is a read-only search/trending source with
// no sensible reverse-ETL target. The connector advertises Write=false, and this
// method satisfies the connectors.Connector interface by rejecting all writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: giphyStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "trending_gifs"
	}
	endpoint, ok := giphyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("giphy stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := giphyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := giphyMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if rating := strings.TrimSpace(req.Config.Config["rating"]); rating != "" {
		base.Set("rating", rating)
	}
	if endpoint.needsQuery {
		q := streamQuery(req.Config, endpoint)
		if q == "" {
			return fmt.Errorf("giphy stream %q requires a search query (config %q or query)", stream, endpoint.queryConfigKey)
		}
		base.Set("q", q)
	}

	return c.harvest(ctx, r, endpoint, base, pageSize, maxPages, emit)
}

// harvest drives Giphy's offset/limit pagination. List responses carry a
// pagination object {total_count, count, offset}; the next page is requested by
// advancing offset until the records returned exhaust total_count or a short
// page is returned. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read giphy %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode giphy %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		// Stop on a short page: fewer records than requested means no more data.
		if len(records) < pageSize {
			return nil
		}
		// Stop when we have consumed everything the server reported.
		offset += len(records)
		total := intAt(resp.Body, "pagination.total_count")
		if total > 0 && offset >= total {
			return nil
		}
		if len(records) == 0 {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise giphy credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	kind := "gif"
	if strings.HasPrefix(stream, "sticker") {
		kind = "sticker"
	} else if strings.HasPrefix(stream, "clip") {
		kind = "video"
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", stream, i),
			"type":              kind,
			"slug":              fmt.Sprintf("fixture-%s-%d", stream, i),
			"url":               fmt.Sprintf("https://giphy.com/gifs/%s_fixture_%d", stream, i),
			"bitly_url":         fmt.Sprintf("https://gph.is/fixture-%d", i),
			"embed_url":         fmt.Sprintf("https://giphy.com/embed/%s_fixture_%d", stream, i),
			"title":             fmt.Sprintf("Fixture %s %d", kind, i),
			"rating":            "g",
			"username":          "polymetrics-fixture",
			"source":            "https://example.com",
			"source_tld":        "example.com",
			"content_url":       "",
			"import_datetime":   giphyFixtureImport,
			"trending_datetime": "0000-00-00 00:00:00",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with api_key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := giphyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := giphySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("giphy connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("api_key", secret),
		UserAgent: giphyUserAgent,
	}, nil
}

// streamQuery resolves the search query for a stream, preferring the
// stream-specific config key (e.g. query_for_gif) and falling back to the
// generic "query" key.
func streamQuery(cfg connectors.RuntimeConfig, endpoint streamEndpoint) string {
	if endpoint.queryConfigKey != "" {
		if q := strings.TrimSpace(cfg.Config[endpoint.queryConfigKey]); q != "" {
			return q
		}
	}
	return strings.TrimSpace(cfg.Config["query"])
}

func giphySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// giphyBaseURL resolves and validates the base URL. The default is
// api.giphy.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func giphyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return giphyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("giphy config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("giphy config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("giphy config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func giphyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return giphyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("giphy config page_size must be an integer: %w", err)
	}
	if value < 1 || value > giphyMaxPageSize {
		return 0, fmt.Errorf("giphy config page_size must be between 1 and %d", giphyMaxPageSize)
	}
	return value, nil
}

func giphyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("giphy config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("giphy config max_pages must be 0 for unlimited or a positive integer")
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

// intAt reads an integer from the response body at a dotted path, returning 0
// when absent or unparseable.
func intAt(body []byte, path string) int {
	s, err := connsdk.StringAt(body, path)
	if err != nil || s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
