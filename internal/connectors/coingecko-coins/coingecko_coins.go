// Package coingeckocoins implements the native pm CoinGecko Coins connector. It
// is a declarative-HTTP per-system connector built on the same shape as the
// stripe reference: a thin package that composes the connsdk toolkit (Requester
// + api-key header auth) with CoinGecko-specific stream definitions.
//
// CoinGecko's coin endpoints are not list-paginated; the connector is driven by
// config rather than by a record-list cursor:
//
//   - market_chart reads /coins/{coin_id}/market_chart?vs_currency&days, which
//     returns three parallel arrays of [unix_ms, value] pairs
//     (prices/market_caps/total_volumes). The read path zips them into one
//     record per timestamp.
//   - history reads /coins/{coin_id}/history?date=DD-MM-YYYY once per day from
//     start_date to end_date; the date param itself is the "page" cursor, so the
//     loop advances one day at a time and emits one snapshot record per date.
//   - coin reads /coins/{coin_id} for a single current metadata snapshot.
//
// Auth is an optional pro API key sent on the x-cg-pro-api-key header (CoinGecko
// allows unauthenticated calls against the public base URL; a key unlocks the
// pro base URL and higher limits). The key only ever flows into the connsdk
// authenticator and is never logged.
//
// Like github/stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package coingeckocoins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	// coingeckoPublicBaseURL is the free/unauthenticated base URL.
	coingeckoPublicBaseURL = "https://api.coingecko.com/api/v3"
	// coingeckoProBaseURL is used when a pro api_key is supplied.
	coingeckoProBaseURL = "https://pro-api.coingecko.com/api/v3"
	// coingeckoAPIKeyHeader carries the pro API key.
	coingeckoAPIKeyHeader = "x-cg-pro-api-key"
	coingeckoUserAgent    = "polymetrics-go-cli"
	// coingeckoDateLayout is CoinGecko's history date format (DD-MM-YYYY).
	coingeckoDateLayout = "02-01-2006"
	// coingeckoMaxHistoryDays bounds the per-day history loop to avoid an
	// unbounded request fan-out from a misconfigured date range.
	coingeckoMaxHistoryDays = 366
)

func init() {
	connectors.RegisterFactory("coingecko-coins", New)
}

// New returns the CoinGecko Coins connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm CoinGecko Coins connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "coingecko-coins" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "coingecko-coins",
		DisplayName:     "CoinGecko Coins",
		IntegrationType: "api",
		Description:     "Reads a coin's market chart time series, daily historical snapshots, and current metadata from the CoinGecko REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to CoinGecko.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// /ping is the canonical unauthenticated connectivity probe; it confirms the
	// base URL and (when supplied) the api_key without reading any coin data.
	if err := r.DoJSON(ctx, http.MethodGet, "ping", nil, nil, nil); err != nil {
		return fmt.Errorf("check coingecko-coins: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: coingeckoStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: every stream starts with an
// empty cursor (full sync); the date/time bounds come from config at read time.
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
		stream = "market_chart"
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	coinID, err := coinID(req.Config)
	if err != nil {
		return err
	}

	switch stream {
	case "market_chart":
		return c.readMarketChart(ctx, r, req.Config, coinID, emit)
	case "history":
		return c.readHistory(ctx, r, req, coinID, emit)
	case "coin":
		return c.readCoin(ctx, r, coinID, emit)
	default:
		return fmt.Errorf("coingecko-coins stream %q not found", stream)
	}
}

// readMarketChart reads /coins/{id}/market_chart and zips the three parallel
// [timestamp, value] arrays into one record per timestamp.
func (c Connector) readMarketChart(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, coinID string, emit func(connectors.Record) error) error {
	vsCurrency, err := vsCurrency(cfg)
	if err != nil {
		return err
	}
	query := url.Values{}
	query.Set("vs_currency", vsCurrency)
	query.Set("days", days(cfg))

	path := fmt.Sprintf("coins/%s/market_chart", url.PathEscape(coinID))
	resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return fmt.Errorf("read coingecko-coins market_chart: %w", err)
	}

	var body struct {
		Prices       [][]json.Number `json:"prices"`
		MarketCaps   [][]json.Number `json:"market_caps"`
		TotalVolumes [][]json.Number `json:"total_volumes"`
	}
	dec := json.NewDecoder(strings.NewReader(string(resp.Body)))
	dec.UseNumber()
	if err := dec.Decode(&body); err != nil {
		return fmt.Errorf("decode coingecko-coins market_chart: %w", err)
	}

	caps := indexByTimestamp(body.MarketCaps)
	vols := indexByTimestamp(body.TotalVolumes)
	for _, point := range body.Prices {
		if err := ctx.Err(); err != nil {
			return err
		}
		if len(point) < 2 {
			continue
		}
		ts := point[0]
		rec := marketChartRecord(coinID, vsCurrency, ts, point[1], caps[ts.String()], vols[ts.String()])
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// readHistory walks one /coins/{id}/history request per day from start_date to
// end_date (inclusive), emitting one snapshot record per date. The date query
// param is what advances the "pages".
func (c Connector) readHistory(ctx context.Context, r *connsdk.Requester, req connectors.ReadRequest, coinID string, emit func(connectors.Record) error) error {
	start, end, err := historyDateRange(req)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("coins/%s/history", url.PathEscape(coinID))
	days := 0
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if days >= coingeckoMaxHistoryDays {
			break
		}
		days++
		date := day.Format(coingeckoDateLayout)
		query := url.Values{}
		query.Set("date", date)
		query.Set("localization", "false")

		var item map[string]any
		if err := r.DoJSON(ctx, http.MethodGet, path, query, nil, &item); err != nil {
			return fmt.Errorf("read coingecko-coins history %s: %w", date, err)
		}
		if err := emit(historyRecord(coinID, date, item)); err != nil {
			return err
		}
	}
	return nil
}

// readCoin reads /coins/{id} for a single current metadata snapshot.
func (c Connector) readCoin(ctx context.Context, r *connsdk.Requester, coinID string, emit func(connectors.Record) error) error {
	path := fmt.Sprintf("coins/%s", url.PathEscape(coinID))
	query := url.Values{}
	query.Set("localization", "false")
	query.Set("tickers", "false")
	query.Set("community_data", "false")
	query.Set("developer_data", "false")

	var item map[string]any
	if err := r.DoJSON(ctx, http.MethodGet, path, query, nil, &item); err != nil {
		return fmt.Errorf("read coingecko-coins coin: %w", err)
	}
	return emit(coinRecord(item))
}

// Write satisfies the connectors.Connector interface. CoinGecko is a read-only
// market-data source, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise coingecko-coins credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	coin := strings.TrimSpace(req.Config.Config["coin_id"])
	if coin == "" {
		coin = "bitcoin"
	}
	vs := strings.TrimSpace(req.Config.Config["vs_currency"])
	if vs == "" {
		vs = "usd"
	}
	switch stream {
	case "market_chart", "":
		base := int64(1767225600000) // 2026-01-01T00:00:00Z in unix ms
		for i := 0; i < 2; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			ts := json.Number(fmt.Sprintf("%d", base+int64(i)*3600000))
			rec := marketChartRecord(coin, vs, ts,
				json.Number(fmt.Sprintf("%d", 42000+i)),
				json.Number(fmt.Sprintf("%d", 800000000000+i)),
				json.Number(fmt.Sprintf("%d", 25000000000+i)))
			if err := emit(rec); err != nil {
				return err
			}
		}
		return nil
	case "history":
		for i := 1; i <= 2; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			date := fmt.Sprintf("0%d-01-2026", i)
			item := map[string]any{
				"id": coin, "symbol": "btc", "name": "Bitcoin",
				"market_data": map[string]any{
					"current_price": map[string]any{vs: 42000 + i},
					"market_cap":    map[string]any{vs: 800000000000 + i},
					"total_volume":  map[string]any{vs: 25000000000 + i},
				},
			}
			if err := emit(historyRecord(coin, date, item)); err != nil {
				return err
			}
		}
		return nil
	case "coin":
		item := map[string]any{
			"id": coin, "symbol": "btc", "name": "Bitcoin",
			"market_cap_rank":   1,
			"hashing_algorithm": "SHA-256",
			"categories":        []any{"Cryptocurrency"},
			"market_data":       map[string]any{"current_price": map[string]any{vs: 42000}},
			"last_updated":      "2026-01-01T00:00:00.000Z",
		}
		return emit(coinRecord(item))
	default:
		return fmt.Errorf("coingecko-coins stream %q not found", stream)
	}
}

// requester builds a connsdk.Requester wired with the resolved base URL and the
// optional pro api-key header. The secret only ever flows into the connsdk
// authenticator; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := coingeckoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	var auth connsdk.Authenticator
	if key := apiKey(cfg); strings.TrimSpace(key) != "" {
		auth = connsdk.APIKeyHeader(coingeckoAPIKeyHeader, key, "")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: coingeckoUserAgent,
	}, nil
}

func apiKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// coingeckoBaseURL resolves and validates the base URL. With no override it
// defaults to the pro base when an api_key is present and the public base
// otherwise. Any override must be an absolute http/https URL with a host to
// bound SSRF risk.
func coingeckoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if strings.TrimSpace(apiKey(cfg)) != "" {
			return coingeckoProBaseURL, nil
		}
		return coingeckoPublicBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("coingecko-coins config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("coingecko-coins config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("coingecko-coins config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func coinID(cfg connectors.RuntimeConfig) (string, error) {
	id := strings.TrimSpace(cfg.Config["coin_id"])
	if id == "" {
		return "", errors.New("coingecko-coins config coin_id is required (e.g. bitcoin)")
	}
	return id, nil
}

func vsCurrency(cfg connectors.RuntimeConfig) (string, error) {
	vs := strings.TrimSpace(cfg.Config["vs_currency"])
	if vs == "" {
		return "", errors.New("coingecko-coins config vs_currency is required (e.g. usd)")
	}
	return vs, nil
}

func days(cfg connectors.RuntimeConfig) string {
	d := strings.TrimSpace(cfg.Config["days"])
	if d == "" {
		return "30"
	}
	return d
}

// historyDateRange parses the start_date (required) and end_date (optional,
// defaulting to start_date) config values in DD-MM-YYYY form into an inclusive
// [start, end] day range.
func historyDateRange(req connectors.ReadRequest) (time.Time, time.Time, error) {
	startRaw := strings.TrimSpace(req.Config.Config["start_date"])
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		startRaw = cursor
	}
	if startRaw == "" {
		return time.Time{}, time.Time{}, errors.New("coingecko-coins config start_date is required (DD-MM-YYYY) for the history stream")
	}
	start, err := time.Parse(coingeckoDateLayout, startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("coingecko-coins config start_date must be DD-MM-YYYY: %w", err)
	}
	endRaw := strings.TrimSpace(req.Config.Config["end_date"])
	if endRaw == "" {
		return start, start, nil
	}
	end, err := time.Parse(coingeckoDateLayout, endRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("coingecko-coins config end_date must be DD-MM-YYYY: %w", err)
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, errors.New("coingecko-coins config end_date must be on or after start_date")
	}
	return start, end, nil
}

// indexByTimestamp maps the first element (timestamp) of each [timestamp, value]
// pair to its second element, so the parallel market_chart arrays can be zipped
// by timestamp regardless of ordering or length mismatches.
func indexByTimestamp(pairs [][]json.Number) map[string]json.Number {
	out := make(map[string]json.Number, len(pairs))
	for _, pair := range pairs {
		if len(pair) < 2 {
			continue
		}
		out[pair[0].String()] = pair[1]
	}
	return out
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
