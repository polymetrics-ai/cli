// Package coinapi implements the native pm CoinAPI connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// X-CoinAPI-Key header auth + top-level-array extraction) with CoinAPI-specific
// stream definitions and endpoints.
//
// CoinAPI is a read-only market-data API. Metadata streams (symbols, exchanges,
// assets) are GET /v1/<resource> calls that return a top-level JSON array.
// Historical series (ohlcv_historical_data, trades_historical_data) are
// symbol-scoped GET /v1/<resource>/<symbol_id>/history calls that return a
// top-level array and are paginated by limit + advancing time_start past the
// last record's time cursor. Auth is a single X-CoinAPI-Key request header.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package coinapi

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
	coinAPIProductionBaseURL = "https://rest.coinapi.io"
	coinAPISandboxBaseURL    = "https://rest-sandbox.coinapi.io"
	coinAPIDefaultLimit      = 100
	coinAPIMaxLimit          = 100000
	coinAPIUserAgent         = "polymetrics-go-cli"
	coinAPIAuthHeader        = "X-CoinAPI-Key"
	// coinAPIFixtureStart is the deterministic candle start used by fixture mode.
	coinAPIFixtureStart = "2026-01-01T00:00:00.0000000Z"
)

func init() {
	connectors.RegisterFactory("coin-api", New)
}

// New returns the CoinAPI connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm CoinAPI connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "coin-api" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "coin-api",
		DisplayName:     "Coin API",
		IntegrationType: "api",
		Description:     "Reads CoinAPI market data: symbols, exchanges, assets, and historical OHLCV and trades for a configured symbol via the CoinAPI REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

// Check verifies the connector is configured well enough to talk to CoinAPI. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := coinAPIBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(coinAPISecret(cfg)) == "" {
		return errors.New("coin-api connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the exchanges list confirms auth and connectivity
	// without depending on symbol/period config.
	if err := r.DoJSON(ctx, http.MethodGet, "v1/exchanges", nil, nil, nil); err != nil {
		return fmt.Errorf("check coin-api: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: coinAPIStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a historical stream starts
// with an empty incremental cursor (full sync from start_date), which read time
// raises via the stored cursor.
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
		stream = "symbols"
	}
	endpoint, ok := coinAPIStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("coin-api stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if endpoint.kind == kindMetadata {
		return c.readMetadata(ctx, r, endpoint, emit)
	}
	return c.readTimeseries(ctx, r, endpoint, req, emit)
}

// readMetadata reads a single top-level-array reference endpoint.
func (c Connector) readMetadata(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, "v1/"+endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read coin-api %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode coin-api %s: %w", endpoint.resource, err)
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

// readTimeseries drives CoinAPI's limit + time_start pagination over a
// symbol-scoped /history endpoint. CoinAPI history returns a top-level array
// ordered by time; the next page is requested with time_start advanced past the
// last record's time cursor. A page shorter than the limit ends the loop. The
// connsdk paginators do not cover the "advance a time param from a body field"
// shape, so the loop lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) readTimeseries(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	symbolID := strings.TrimSpace(req.Config.Config["symbol_id"])
	if symbolID == "" {
		return errors.New("coin-api config symbol_id is required for historical streams")
	}
	period := strings.TrimSpace(req.Config.Config["period"])
	limit, err := coinAPILimit(req.Config)
	if err != nil {
		return err
	}
	timeStart, err := timeseriesLowerBound(req)
	if err != nil {
		return err
	}
	timeEnd := strings.TrimSpace(req.Config.Config["end_date"])

	path := fmt.Sprintf("v1/%s/%s/history", endpoint.resource, url.PathEscape(symbolID))
	base := url.Values{}
	base.Set("limit", strconv.Itoa(limit))
	if endpoint.resource == "ohlcv" {
		if period == "" {
			return errors.New("coin-api config period is required for ohlcv_historical_data")
		}
		base.Set("period_id", period)
	}
	if timeEnd != "" {
		base.Set("time_end", timeEnd)
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if timeStart != "" {
			query.Set("time_start", timeStart)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read coin-api %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode coin-api %s page: %w", endpoint.resource, err)
		}
		lastCursor := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			rec := endpoint.mapRecord(item)
			rec["symbol_id"] = symbolID
			if endpoint.resource == "ohlcv" {
				rec["period_id"] = period
			}
			if cv := stringField(item, endpoint.cursorField); cv != "" {
				lastCursor = cv
			}
			if err := emit(rec); err != nil {
				return err
			}
		}
		// Terminate on a short page or when the cursor cannot advance.
		if len(records) < limit || lastCursor == "" || lastCursor == timeStart {
			return nil
		}
		timeStart = lastCursor
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise coin-api credential-free (mirrors stripe).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 0; i < 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"symbol_id":          "BITSTAMP_SPOT_BTC_USD",
			"exchange_id":        "BITSTAMP",
			"symbol_type":        "SPOT",
			"asset_id_base":      "BTC",
			"asset_id_quote":     "USD",
			"asset_id":           fmt.Sprintf("FIX%d", i+1),
			"name":               fmt.Sprintf("Fixture %d", i+1),
			"website":            "https://example.com",
			"type_is_crypto":     int64(1),
			"data_symbols_count": int64(10 + i),
			"price_usd":          float64(20000 + i),
			"data_start":         "2020-01-01",
			"data_end":           "2026-01-01",
			"time_period_start":  coinAPIFixtureStart,
			"time_period_end":    coinAPIFixtureStart,
			"time_exchange":      coinAPIFixtureStart,
			"time_coinapi":       coinAPIFixtureStart,
			"uuid":               fmt.Sprintf("00000000-0000-0000-0000-00000000000%d", i+1),
			"price_open":         float64(29000 + i),
			"price_high":         float64(29500 + i),
			"price_low":          float64(28500 + i),
			"price_close":        float64(29300 + i),
			"volume_traded":      float64(i) + 1.5,
			"trades_count":       int64(1000 + i),
			"price":              float64(29100 + i),
			"size":               float64(i) + 0.25,
			"taker_side":         "BUY",
		}
		rec := endpoint.mapRecord(item)
		if endpoint.kind == kindTimeseries {
			rec["symbol_id"] = "BITSTAMP_SPOT_BTC_USD"
			if endpoint.resource == "ohlcv" {
				rec["period_id"] = "1DAY"
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

// requester builds a connsdk.Requester wired with the X-CoinAPI-Key header and
// the resolved base URL. The secret only ever flows into the auth header; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := coinAPIBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := coinAPISecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("coin-api connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(coinAPIAuthHeader, secret, ""),
		UserAgent: coinAPIUserAgent,
	}, nil
}

// timeseriesLowerBound returns the ISO-8601 time_start lower bound, derived from
// the incremental cursor (if any) or else the start_date config.
func timeseriesLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	return startDate, nil
}

func coinAPISecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// coinAPIBaseURL resolves and validates the base URL. An explicit base_url
// override wins (validated for scheme+host to bound SSRF risk); otherwise the
// environment config selects the production or sandbox CoinAPI host.
func coinAPIBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if override := strings.TrimSpace(cfg.Config["base_url"]); override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("coin-api config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("coin-api config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("coin-api config base_url must include a host")
		}
		return strings.TrimRight(override, "/"), nil
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Config["environment"])) {
	case "", "production":
		return coinAPIProductionBaseURL, nil
	case "sandbox":
		return coinAPISandboxBaseURL, nil
	default:
		return "", fmt.Errorf("coin-api config environment must be production or sandbox, got %q", cfg.Config["environment"])
	}
}

func coinAPILimit(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["limit"])
	if raw == "" {
		return coinAPIDefaultLimit, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("coin-api config limit must be an integer: %w", err)
	}
	if value < 1 || value > coinAPIMaxLimit {
		return 0, fmt.Errorf("coin-api config limit must be between 1 and %d", coinAPIMaxLimit)
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

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Write is unsupported: CoinAPI is a read-only market-data source.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
