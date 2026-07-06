// Package yahoofinanceprice implements a read-only Yahoo Finance chart-price connector.
package yahoofinanceprice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL = "https://query1.finance.yahoo.com"
	userAgent      = "polymetrics-go-cli"
)

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "yahoo-finance-price" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "yahoo-finance-price", DisplayName: "Yahoo Finance Price", IntegrationType: "api", Description: "Reads public Yahoo Finance chart prices and flattens them into OHLCV records. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

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
	_, err = r.Do(ctx, http.MethodGet, chartPath(symbol(cfg)), chartQuery(cfg), nil)
	if err != nil {
		return fmt.Errorf("check yahoo-finance-price: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "symbol", Type: "string"}, {Name: "currency", Type: "string"}, {Name: "timestamp", Type: "number"}, {Name: "open", Type: "number"}, {Name: "high", Type: "number"}, {Name: "low", Type: "number"}, {Name: "close", Type: "number"}, {Name: "volume", Type: "number"}, {Name: "adjclose", Type: "number"}}
	stream := connectors.Stream{Name: "prices", Description: "Yahoo Finance chart OHLCV prices for the configured symbol.", Fields: fields, PrimaryKey: []string{"symbol", "timestamp"}, CursorFields: []string{"timestamp"}}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{stream}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "prices"
	}
	if stream != "prices" {
		return fmt.Errorf("yahoo-finance-price stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, chartPath(symbol(req.Config)), chartQuery(req.Config), nil)
	if err != nil {
		return fmt.Errorf("read yahoo-finance-price: %w", err)
	}
	records, err := flattenChart(resp.Body)
	if err != nil {
		return err
	}
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"symbol": "FIXTURE", "currency": "USD", "timestamp": int64(1767225600 + i*86400), "open": float64(i), "high": float64(i) + 0.5, "low": float64(i) - 0.5, "close": float64(i) + 0.25, "volume": float64(i * 100), "adjclose": float64(i) + 0.2, "fixture": true}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

type chartResponse struct {
	Chart struct {
		Result []chartResult `json:"result"`
		Error  any           `json:"error"`
	} `json:"chart"`
}

type chartResult struct {
	Meta       map[string]any `json:"meta"`
	Timestamp  []int64        `json:"timestamp"`
	Indicators struct {
		Quote    []map[string][]any `json:"quote"`
		AdjClose []map[string][]any `json:"adjclose"`
	} `json:"indicators"`
}

func flattenChart(body []byte) ([]connectors.Record, error) {
	var payload chartResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode yahoo-finance-price chart: %w", err)
	}
	if payload.Chart.Error != nil {
		return nil, errors.New("yahoo-finance-price chart response returned an error")
	}
	if len(payload.Chart.Result) == 0 {
		return nil, nil
	}
	result := payload.Chart.Result[0]
	quote := map[string][]any{}
	if len(result.Indicators.Quote) > 0 {
		quote = result.Indicators.Quote[0]
	}
	adj := map[string][]any{}
	if len(result.Indicators.AdjClose) > 0 {
		adj = result.Indicators.AdjClose[0]
	}
	out := make([]connectors.Record, 0, len(result.Timestamp))
	for i, ts := range result.Timestamp {
		out = append(out, connectors.Record{
			"symbol":    stringValue(result.Meta["symbol"]),
			"currency":  stringValue(result.Meta["currency"]),
			"timestamp": ts,
			"open":      numberAt(quote["open"], i),
			"high":      numberAt(quote["high"], i),
			"low":       numberAt(quote["low"], i),
			"close":     numberAt(quote["close"], i),
			"volume":    numberAt(quote["volume"], i),
			"adjclose":  numberAt(adj["adjclose"], i),
		})
	}
	return out, nil
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func numberAt(values []any, index int) any {
	if index < 0 || index >= len(values) || values[index] == nil {
		return nil
	}
	return values[index]
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, defaultBaseURL, "yahoo-finance-price")
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: userAgent}, nil
}

func chartPath(symbol string) string {
	return "v8/finance/chart/" + url.PathEscape(symbol)
}

func chartQuery(cfg connectors.RuntimeConfig) url.Values {
	query := url.Values{}
	rangeValue := strings.TrimSpace(cfg.Config["range"])
	if rangeValue == "" {
		rangeValue = "1mo"
	}
	interval := strings.TrimSpace(cfg.Config["interval"])
	if interval == "" {
		interval = "1d"
	}
	query.Set("range", rangeValue)
	query.Set("interval", interval)
	return query
}

func symbol(cfg connectors.RuntimeConfig) string {
	value := strings.TrimSpace(cfg.Config["symbol"])
	if value == "" {
		return "AAPL"
	}
	return strings.ToUpper(value)
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func baseURL(cfg connectors.RuntimeConfig, fallback, connector string) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", connector, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https", connector)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", connector)
	}
	return strings.TrimRight(base, "/"), nil
}
