// Package polygonstockapi implements a read-only Polygon.io stock API connector.
package polygonstockapi

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
	defaultBaseURL  = "https://api.polygon.io"
	defaultPageSize = 100
	defaultMaxPages = 3
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("polygon-stock-api", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "polygon-stock-api" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "polygon-stock-api",
		DisplayName:     "Polygon Stock API",
		IntegrationType: "api",
		Description:     "Reads Polygon.io stock tickers, dividends, and splits through reference endpoints.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
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
	if _, err := r.Do(ctx, http.MethodGet, "v3/reference/tickers", url.Values{"limit": {"1"}}, nil); err != nil {
		return fmt.Errorf("check polygon-stock-api: %w", err)
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
		stream = "tickers"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("polygon-stock-api stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
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
	path := endpoint.path
	query := polygonQuery(req.Config, endpoint, pageSize)
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read polygon-stock-api %s: %w", stream, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode polygon-stock-api %s: %w", stream, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_url")
		if err != nil {
			return fmt.Errorf("decode polygon-stock-api next_url: %w", err)
		}
		if strings.TrimSpace(next) == "" || len(records) < pageSize {
			return nil
		}
		path = next
		query = nil
	}
	return nil
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path      string
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"tickers":   {path: "v3/reference/tickers", mapRecord: tickerRecord},
	"dividends": {path: "v3/reference/dividends", mapRecord: dividendRecord},
	"splits":    {path: "v3/reference/splits", mapRecord: splitRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "tickers", Description: "Polygon.io stock ticker reference records.", PrimaryKey: []string{"ticker"}, Fields: tickerFields()},
		{Name: "dividends", Description: "Polygon.io dividend reference records.", PrimaryKey: []string{"id"}, CursorFields: []string{"ex_dividend_date"}, Fields: dividendFields()},
		{Name: "splits", Description: "Polygon.io stock split reference records.", PrimaryKey: []string{"id"}, CursorFields: []string{"execution_date"}, Fields: splitFields()},
	}
}

func tickerFields() []connectors.Field {
	return []connectors.Field{{Name: "ticker", Type: "string"}, {Name: "name", Type: "string"}, {Name: "market", Type: "string"}, {Name: "locale", Type: "string"}, {Name: "primary_exchange", Type: "string"}, {Name: "currency_name", Type: "string"}, {Name: "active", Type: "boolean"}}
}

func dividendFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "ticker", Type: "string"}, {Name: "ex_dividend_date", Type: "string"}, {Name: "cash_amount", Type: "number"}, {Name: "currency", Type: "string"}}
}

func splitFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "ticker", Type: "string"}, {Name: "execution_date", Type: "string"}, {Name: "split_from", Type: "number"}, {Name: "split_to", Type: "number"}}
}

func tickerRecord(item map[string]any) connectors.Record {
	return connectors.Record{"ticker": item["ticker"], "name": item["name"], "market": item["market"], "locale": item["locale"], "primary_exchange": item["primary_exchange"], "currency_name": item["currency_name"], "active": item["active"]}
}

func dividendRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "cash_amount"), "ticker": item["ticker"], "ex_dividend_date": item["ex_dividend_date"], "cash_amount": item["cash_amount"], "currency": first(item, "currency", "currency_name")}
}

func splitRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "execution_date"), "ticker": item["ticker"], "execution_date": item["execution_date"], "split_from": item["split_from"], "split_to": item["split_to"]}
}

func readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("fixture-%d", i), "ticker": fmt.Sprintf("FIX%d", i), "name": fmt.Sprintf("Fixture %d", i), "market": "stocks", "locale": "us", "ex_dividend_date": fmt.Sprintf("2026-01-0%d", i), "execution_date": fmt.Sprintf("2026-02-0%d", i), "cash_amount": float64(i), "split_from": 1, "split_to": 2}
		if err := emit(endpoint.mapRecord(item)); err != nil {
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
		return nil, errors.New("polygon-stock-api connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent}, nil
}

func polygonQuery(cfg connectors.RuntimeConfig, endpoint streamEndpoint, pageSize int) url.Values {
	q := url.Values{"limit": {strconv.Itoa(pageSize)}}
	for _, key := range []string{"ticker", "market", "locale", "type", "active", "sort", "order", "ex_dividend_date", "execution_date"} {
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			q.Set(key, v)
		}
	}
	if endpoint.path == "v3/reference/tickers" && q.Get("active") == "" {
		q.Set("active", "true")
	}
	return q
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("polygon-stock-api config base_url is invalid: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", errors.New("polygon-stock-api config base_url must be an absolute http or https URL")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "page_size", defaultPageSize, 1, 1000)
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "max_pages", defaultMaxPages, 0, 10000)
}

func intConfig(cfg connectors.RuntimeConfig, key string, def, min, max int) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config[key]))
	if raw == "" {
		return def, nil
	}
	if key == "max_pages" && (raw == "all" || raw == "unlimited") {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min {
		return 0, fmt.Errorf("polygon-stock-api config %s must be an integer >= %d", key, min)
	}
	if max > 0 && value > max {
		return max, nil
	}
	return value, nil
}

func first(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
