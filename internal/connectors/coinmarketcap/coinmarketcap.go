// Package coinmarketcap implements the native pm CoinMarketCap connector. It is
// a declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package composing the connsdk toolkit (Requester +
// APIKeyHeader auth + RecordsAt extraction) with CoinMarketCap-specific stream
// definitions and endpoints.
//
// CoinMarketCap exposes the Pro API at https://pro-api.coinmarketcap.com,
// authenticates via the X-CMC_PRO_API_KEY header, and wraps every payload in
// {status:{...}, data:...}. List endpoints (map, listings/latest, categories,
// fiat/map) paginate with a 1-based `start` cursor plus a `limit`; a page
// shorter than `limit` ends the read. global-metrics returns a single object.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect. CoinMarketCap is a read-only market-data source, so the
// connector advertises no write capability.
package coinmarketcap

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
	coinmarketcapDefaultBaseURL  = "https://pro-api.coinmarketcap.com"
	coinmarketcapAuthHeader      = "X-CMC_PRO_API_KEY"
	coinmarketcapDefaultPageSize = 100
	coinmarketcapMaxPageSize     = 5000
	coinmarketcapUserAgent       = "polymetrics-go-cli"
	coinmarketcapDefaultStream   = "listings_latest"
)

func init() {
	connectors.RegisterFactory("coinmarketcap", New)
}

// New returns the CoinMarketCap connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm CoinMarketCap connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "coinmarketcap" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "coinmarketcap",
		DisplayName:     "CoinMarketCap",
		IntegrationType: "api",
		Description:     "Reads CoinMarketCap cryptocurrency map, latest market listings, categories, fiat currencies, and global metrics through the CoinMarketCap Pro API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to
// CoinMarketCap. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := coinmarketcapBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(coinmarketcapSecret(cfg)) == "" {
		return errors.New("coinmarketcap connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the fiat map confirms auth and connectivity cheaply
	// (it is a small, low-credit-cost endpoint) without mutating anything.
	q := url.Values{"limit": []string{"1"}, "start": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "v1/fiat/map", q, nil, nil); err != nil {
		return fmt.Errorf("check coinmarketcap: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: coinmarketcapStreams()}, nil
}

// Write is unsupported: CoinMarketCap is a read-only market-data source, so the
// connector satisfies the interface but rejects writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = coinmarketcapDefaultStream
	}
	endpoint, ok := coinmarketcapStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("coinmarketcap stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if !endpoint.paginated {
		return c.readSingle(ctx, r, endpoint, emit)
	}
	pageSize, err := coinmarketcapPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := coinmarketcapMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// readSingle handles endpoints whose `data` is a single object (global-metrics).
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read coinmarketcap %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "data")
	if err != nil {
		return fmt.Errorf("decode coinmarketcap %s: %w", endpoint.resource, err)
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

// harvest drives CoinMarketCap's 1-based start/limit pagination. Each page is
// requested with start=<1-based index> and limit=<pageSize>; a page returning
// fewer than limit records is the last one. There is no body token for this
// shape, so the loop lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	start := 1
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("start", strconv.Itoa(start))
		query.Set("limit", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read coinmarketcap %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode coinmarketcap %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		start += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise coinmarketcap credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                      i,
			"name":                    fmt.Sprintf("Fixture Coin %d", i),
			"symbol":                  fmt.Sprintf("FIX%d", i),
			"slug":                    fmt.Sprintf("fixture-coin-%d", i),
			"rank":                    i,
			"cmc_rank":                i,
			"is_active":               1,
			"num_market_pairs":        100 * i,
			"circulating_supply":      1000000 * i,
			"total_supply":            2000000 * i,
			"max_supply":              21000000,
			"first_historical_data":   "2013-04-28T18:47:21.000Z",
			"last_historical_data":    "2026-01-01T00:00:00.000Z",
			"last_updated":            "2026-01-01T00:00:00.000Z",
			"title":                   fmt.Sprintf("Fixture Category %d", i),
			"description":             "Deterministic fixture record.",
			"num_tokens":              10 * i,
			"market_cap":              1234567.0 * float64(i),
			"volume":                  7654.0 * float64(i),
			"sign":                    "$",
			"active_cryptocurrencies": 9000 + i,
			"total_cryptocurrencies":  12000 + i,
			"active_market_pairs":     80000 + i,
			"active_exchanges":        500 + i,
			"total_exchanges":         9000 + i,
			"btc_dominance":           48.5,
			"eth_dominance":           18.2,
			"quote": map[string]any{
				"USD": map[string]any{"price": 100.5 * float64(i), "market_cap": 1234567.0 * float64(i)},
			},
			"connector": "coinmarketcap",
			"fixture":   true,
		}
		// categories use a string id in the real API; honor that for the
		// categories stream so the fixture mirrors production shape.
		if stream == "categories" {
			item["id"] = fmt.Sprintf("cat_fixture_%d", i)
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the X-CMC_PRO_API_KEY header
// auth and the resolved base URL. The secret only ever flows into
// connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := coinmarketcapBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := coinmarketcapSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("coinmarketcap connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(coinmarketcapAuthHeader, secret, ""),
		UserAgent: coinmarketcapUserAgent,
		DefaultHeaders: map[string]string{
			"Accept-Encoding": "deflate, gzip",
		},
	}, nil
}

func coinmarketcapSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// coinmarketcapBaseURL resolves and validates the base URL. The default is
// pro-api.coinmarketcap.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func coinmarketcapBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return coinmarketcapDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("coinmarketcap config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("coinmarketcap config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("coinmarketcap config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func coinmarketcapPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return coinmarketcapDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("coinmarketcap config page_size must be an integer: %w", err)
	}
	if value < 1 || value > coinmarketcapMaxPageSize {
		return 0, fmt.Errorf("coinmarketcap config page_size must be between 1 and %d", coinmarketcapMaxPageSize)
	}
	return value, nil
}

func coinmarketcapMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("coinmarketcap config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("coinmarketcap config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
