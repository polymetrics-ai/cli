// Package stockdata implements the native pm StockData connector.
package stockdata

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
	connectorName   = "stockdata"
	defaultBaseURL  = "https://api.stockdata.org/v1"
	defaultPageSize = 100
	maxPageSize     = 1000
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamEndpoint struct {
	resource     string
	recordsPath  string
	needsSymbols bool
	fields       []connectors.Field
	mapRecord    func(map[string]any) connectors.Record
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "StockData", IntegrationType: "api", Description: "Reads stockdata.org tickers, end-of-day prices, intraday prices, and news through the StockData API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	_, err = r.Do(ctx, http.MethodGet, "news/all", url.Values{"limit": []string{"1"}}, nil)
	if err != nil {
		return redactHTTPError(err, "check")
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "eod_prices"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("stockdata stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	base, err := baseQuery(endpoint, req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	paginator := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize: pageSize}
	err = connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
	if err != nil {
		return redactHTTPError(err, endpoint.resource)
	}
	return nil
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "api_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("stockdata connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("api_token", token), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("stockdata config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("stockdata config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("stockdata config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func baseQuery(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (url.Values, error) {
	q := url.Values{}
	if endpoint.needsSymbols {
		symbols := strings.TrimSpace(cfg.Config["symbols"])
		if symbols == "" {
			return nil, errors.New("stockdata stream requires config symbols")
		}
		q.Set("symbols", symbols)
	}
	if from := strings.TrimSpace(cfg.Config["date_from"]); from != "" {
		q.Set("date_from", from)
	}
	if to := strings.TrimSpace(cfg.Config["date_to"]); to != "" {
		q.Set("date_to", to)
	}
	return q, nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "stockdata config page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("stockdata config max_pages must be a non-negative integer: %w", err)
	}
	return value, nil
}

func boundedInt(raw string, def, max int, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s must be between 1 and %d", name, max)
	}
	return value, nil
}

func redactHTTPError(err error, resource string) error {
	var httpErr *connsdk.HTTPError
	if errors.As(err, &httpErr) {
		return fmt.Errorf("stockdata %s request returned http %d", resource, httpErr.Status)
	}
	return err
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"ticker": "AAPL", "symbol": "AAPL", "date": fmt.Sprintf("2026-01-0%d", i), "close": float64(100 + i), "title": fmt.Sprintf("Fixture %s %d", stream, i), "url": "https://example.com/fixture"}
		rec := endpoint.mapRecord(item)
		rec["connector"] = connectorName
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "tickers", Description: "StockData ticker symbols.", PrimaryKey: []string{"symbol"}, Fields: streamEndpoints["tickers"].fields},
		{Name: "eod_prices", Description: "StockData end-of-day prices for configured symbols.", PrimaryKey: []string{"ticker", "date"}, CursorFields: []string{"date"}, Fields: streamEndpoints["eod_prices"].fields},
		{Name: "intraday_prices", Description: "StockData intraday prices for configured symbols.", PrimaryKey: []string{"ticker", "date"}, CursorFields: []string{"date"}, Fields: streamEndpoints["intraday_prices"].fields},
		{Name: "news", Description: "StockData market news.", PrimaryKey: []string{"url"}, Fields: streamEndpoints["news"].fields},
	}
}

var streamEndpoints = map[string]streamEndpoint{
	"tickers":         {resource: "data/tickers", recordsPath: "data", fields: tickerFields(), mapRecord: copyRecord("symbol", "name", "exchange")},
	"eod_prices":      {resource: "data/eod", recordsPath: "data", needsSymbols: true, fields: priceFields(), mapRecord: copyRecord("ticker", "date", "close")},
	"intraday_prices": {resource: "data/intraday", recordsPath: "data", needsSymbols: true, fields: priceFields(), mapRecord: copyRecord("ticker", "date", "close")},
	"news":            {resource: "news/all", recordsPath: "data", fields: newsFields(), mapRecord: copyRecord("title", "url", "published_at")},
}

func tickerFields() []connectors.Field {
	return []connectors.Field{{Name: "symbol", Type: "string"}, {Name: "name", Type: "string"}, {Name: "exchange", Type: "string"}}
}

func priceFields() []connectors.Field {
	return []connectors.Field{{Name: "ticker", Type: "string"}, {Name: "date", Type: "timestamp"}, {Name: "close", Type: "number"}}
}

func newsFields() []connectors.Field {
	return []connectors.Field{{Name: "title", Type: "string"}, {Name: "url", Type: "string"}, {Name: "published_at", Type: "timestamp"}}
}

func copyRecord(keys ...string) func(map[string]any) connectors.Record {
	return func(item map[string]any) connectors.Record {
		rec := connectors.Record{}
		for _, key := range keys {
			rec[key] = item[key]
		}
		return rec
	}
}
