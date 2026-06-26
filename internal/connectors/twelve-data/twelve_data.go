// Package twelvedata implements a read-only native connector for Twelve Data APIs.
package twelvedata

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
	defaultBaseURL    = "https://api.twelvedata.com"
	defaultSymbol     = "AAPL"
	defaultInterval   = "1day"
	defaultOutputSize = 100
	maxOutputSize     = 5000
	userAgent         = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("twelve-data", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "twelve-data" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "twelve-data", DisplayName: "Twelve Data", IntegrationType: "api", Description: "Reads Twelve Data time series, quotes, stocks, and forex pair reference data.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "quote", url.Values{"symbol": []string{symbol(cfg)}}, nil, nil); err != nil {
		return fmt.Errorf("check twelve-data: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "time_series"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("twelve-data stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, spec, req.Config, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	query, err := streamQuery(stream, req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, spec.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read twelve-data %s: %w", spec.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
	if err != nil {
		return fmt.Errorf("decode twelve-data %s: %w", spec.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := spec.mapRecord(item)
		if spec.includeSymbol {
			rec["symbol"] = symbol(req.Config)
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readFixture(ctx context.Context, stream string, spec streamSpec, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"symbol": symbol(cfg), "datetime": fmt.Sprintf("2026-01-0%d", i), "open": "100", "close": "110", "volume": "1200", "name": fmt.Sprintf("Fixture %s %d", stream, i), "currency": "USD"}
		rec := spec.mapRecord(item)
		if spec.includeSymbol {
			rec["symbol"] = symbol(cfg)
		}
		rec["connector"] = "twelve-data"
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := strings.TrimSpace(cfg.Secrets["api_key"])
	if key == "" {
		return nil, errors.New("twelve-data connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("apikey", key), UserAgent: userAgent}, nil
}

func streamQuery(stream string, cfg connectors.RuntimeConfig) (url.Values, error) {
	query := url.Values{}
	switch stream {
	case "time_series":
		query.Set("symbol", symbol(cfg))
		query.Set("interval", stringDefault(cfg.Config["interval"], defaultInterval))
		outputSize, err := boundedInt(cfg.Config["output_size"], defaultOutputSize, maxOutputSize, "twelve-data config output_size")
		if err != nil {
			return nil, err
		}
		query.Set("outputsize", strconv.Itoa(outputSize))
	case "quote":
		query.Set("symbol", symbol(cfg))
	}
	return query, nil
}

func symbol(cfg connectors.RuntimeConfig) string {
	return stringDefault(cfg.Config["symbol"], defaultSymbol)
}
func stringDefault(raw, def string) string {
	if v := strings.TrimSpace(raw); v != "" {
		return v
	}
	return def
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("twelve-data config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("twelve-data config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("twelve-data config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
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
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamSpec struct {
	resource      string
	recordsPath   string
	includeSymbol bool
	mapRecord     func(map[string]any) connectors.Record
}

var streamSpecs = map[string]streamSpec{
	"time_series": {resource: "time_series", recordsPath: "values", includeSymbol: true, mapRecord: timeSeriesRecord},
	"quote":       {resource: "quote", recordsPath: ".", includeSymbol: true, mapRecord: quoteRecord},
	"stocks":      {resource: "stocks", recordsPath: "data", mapRecord: referenceRecord},
	"forex_pairs": {resource: "forex_pairs", recordsPath: "data", mapRecord: referenceRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "time_series", Description: "Twelve Data OHLCV time series for a symbol.", PrimaryKey: []string{"symbol", "datetime"}, CursorFields: []string{"datetime"}, Fields: []connectors.Field{{Name: "symbol", Type: "string"}, {Name: "datetime", Type: "string"}, {Name: "open", Type: "string"}, {Name: "close", Type: "string"}, {Name: "volume", Type: "string"}}},
		{Name: "quote", Description: "Twelve Data latest quote for a symbol.", PrimaryKey: []string{"symbol"}, Fields: []connectors.Field{{Name: "symbol", Type: "string"}, {Name: "name", Type: "string"}, {Name: "currency", Type: "string"}, {Name: "close", Type: "string"}}},
		{Name: "stocks", Description: "Twelve Data stock reference data.", PrimaryKey: []string{"symbol"}, Fields: referenceFields()},
		{Name: "forex_pairs", Description: "Twelve Data forex pair reference data.", PrimaryKey: []string{"symbol"}, Fields: referenceFields()},
	}
}

func referenceFields() []connectors.Field {
	return []connectors.Field{{Name: "symbol", Type: "string"}, {Name: "name", Type: "string"}, {Name: "currency", Type: "string"}}
}
func timeSeriesRecord(item map[string]any) connectors.Record {
	return connectors.Record{"datetime": item["datetime"], "open": item["open"], "high": item["high"], "low": item["low"], "close": item["close"], "volume": item["volume"]}
}
func quoteRecord(item map[string]any) connectors.Record {
	return connectors.Record{"symbol": item["symbol"], "name": item["name"], "currency": item["currency"], "close": item["close"]}
}
func referenceRecord(item map[string]any) connectors.Record {
	return connectors.Record{"symbol": item["symbol"], "name": item["name"], "currency": item["currency"]}
}
