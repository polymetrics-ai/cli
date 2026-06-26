// Package finnhub implements the native pm Finnhub connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit (Requester +
// APIKeyHeader auth + RecordsAt extraction), modeled on the stripe reference
// connector.
//
// Finnhub (https://finnhub.io) exposes free real-time stock, forex, and crypto
// market data. Its REST endpoints return top-level JSON arrays (or a single
// object for company_profile) with no list pagination, so the connector fans a
// stream out across the configured symbols/exchange instead of paginating a
// cursor.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package finnhub

import (
	"context"
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
	finnhubDefaultBaseURL = "https://finnhub.io/api/v1"
	finnhubUserAgent      = "polymetrics-go-cli"
	finnhubTokenHeader    = "X-Finnhub-Token"
	// finnhubFixtureDatetime is the deterministic epoch-seconds timestamp used by
	// the fixture-mode records (2026-01-01T00:00:00Z).
	finnhubFixtureDatetime int64 = 1767225600
	// defaultExchange and defaultNewsCategory mirror the catalog defaults.
	defaultExchange     = "US"
	defaultNewsCategory = "general"
)

func init() {
	connectors.RegisterFactory("finnhub", New)
}

// New returns the Finnhub connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Finnhub connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "finnhub" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "finnhub",
		DisplayName:     "Finnhub",
		IntegrationType: "api",
		Description:     "Reads Finnhub stock symbols, company and market news, company profiles, and analyst recommendations through the Finnhub REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Finnhub. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := finnhubBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(finnhubSecret(cfg)) == "" {
		return errors.New("finnhub connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of stock symbols confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"exchange": []string{finnhubExchange(cfg)}}
	if err := r.DoJSON(ctx, http.MethodGet, "stock/symbol", q, nil, nil); err != nil {
		return fmt.Errorf("check finnhub: %w", err)
	}
	return nil
}

// Write is unsupported: Finnhub is a read-only market-data API with no sensible
// reverse-ETL surface, so the connector advertises Write=false and rejects writes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: finnhubStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Finnhub stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "stock_symbols"
	}
	endpoint, ok := finnhubStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("finnhub stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, req, endpoint, emit)
}

// harvest fans the stream out across its request space. Finnhub has no list
// pagination: array endpoints return the full result in one body, so the
// connector's "pages" are the configured symbols (scopeSymbol) or a single
// exchange/global request (scopeExchange/scopeOnce).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, req connectors.ReadRequest, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	switch endpoint.scope {
	case scopeSymbol:
		symbols := finnhubSymbols(req.Config)
		if len(symbols) == 0 {
			return errors.New("finnhub config symbols is required for this stream")
		}
		from, to, err := dateWindow(req)
		if err != nil {
			return err
		}
		for _, symbol := range symbols {
			if err := ctx.Err(); err != nil {
				return err
			}
			q := url.Values{"symbol": []string{symbol}}
			if endpoint.dateWindow {
				q.Set("from", from)
				q.Set("to", to)
			}
			if err := c.fetchAndEmit(ctx, r, endpoint, q, symbol, emit); err != nil {
				return err
			}
		}
		return nil
	case scopeExchange:
		q := url.Values{"exchange": []string{finnhubExchange(req.Config)}}
		return c.fetchAndEmit(ctx, r, endpoint, q, "", emit)
	default: // scopeOnce
		q := url.Values{"category": []string{finnhubNewsCategory(req.Config)}}
		return c.fetchAndEmit(ctx, r, endpoint, q, "", emit)
	}
}

// fetchAndEmit issues one request, extracts the array (or single object) at the
// root, and emits each mapped record.
func (c Connector) fetchAndEmit(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, query url.Values, symbol string, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read finnhub %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode finnhub %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item, symbol)); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise finnhub credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	symbols := finnhubSymbols(req.Config)
	if len(symbols) == 0 {
		symbols = []string{"AAPL", "MSFT"}
	}
	for i, symbol := range symbols {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                   int64(i + 1),
			"symbol":               symbol,
			"displaySymbol":        symbol,
			"description":          fmt.Sprintf("Fixture %s", symbol),
			"type":                 "Common Stock",
			"currency":             "USD",
			"mic":                  "XNAS",
			"category":             "company",
			"datetime":             finnhubFixtureDatetime + int64(i),
			"headline":             fmt.Sprintf("Fixture headline %d", i+1),
			"source":               "Fixture",
			"summary":              "Deterministic fixture record.",
			"related":              symbol,
			"url":                  "https://example.com/fixture",
			"ticker":               symbol,
			"name":                 fmt.Sprintf("Fixture %s Inc", symbol),
			"country":              "US",
			"exchange":             defaultExchange,
			"marketCapitalization": int64(1000 * (i + 1)),
			"shareOutstanding":     int64(500 * (i + 1)),
			"finnhubIndustry":      "Technology",
			"period":               "2026-01-01",
			"buy":                  int64(10),
			"hold":                 int64(5),
			"sell":                 int64(1),
			"strongBuy":            int64(8),
			"strongSell":           int64(0),
		}
		record := endpoint.mapRecord(item, symbol)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the X-Finnhub-Token header auth
// and the resolved base URL. The secret only ever flows into the authenticator;
// it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := finnhubBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := finnhubSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("finnhub connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(finnhubTokenHeader, secret, ""),
		UserAgent: finnhubUserAgent,
	}, nil
}

// dateWindow returns the from/to (YYYY-MM-DD) window for company-news, derived
// from the incremental cursor or start_date config for the lower bound and today
// (UTC) for the upper bound. A missing lower bound defaults to 30 days ago.
func dateWindow(req connectors.ReadRequest) (string, string, error) {
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -30)

	if cursor := connsdk.Cursor(req.State); cursor != "" {
		if t, ok := parseEpochOrDate(cursor); ok {
			from = t
		}
	} else if startDate := strings.TrimSpace(req.Config.Config["start_date"]); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			return "", "", fmt.Errorf("finnhub config start_date must be RFC3339: %w", err)
		}
		from = t.UTC()
	}
	const layout = "2006-01-02"
	return from.Format(layout), to.Format(layout), nil
}

// parseEpochOrDate interprets a cursor as either epoch seconds (Finnhub news
// datetime) or an RFC3339 timestamp.
func parseEpochOrDate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC(), true
	}
	if secs, err := parseInt64(value); err == nil && secs > 0 {
		return time.Unix(secs, 0).UTC(), true
	}
	return time.Time{}, false
}

func finnhubSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// finnhubBaseURL resolves and validates the base URL. The default is
// finnhub.io/api/v1; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func finnhubBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return finnhubDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("finnhub config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("finnhub config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("finnhub config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// finnhubSymbols parses the comma-separated (or whitespace-separated) symbols
// config into a normalized, upper-cased list.
func finnhubSymbols(cfg connectors.RuntimeConfig) []string {
	raw := strings.TrimSpace(cfg.Config["symbols"])
	if raw == "" {
		return nil
	}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n' || r == '\t' || r == ';'
	})
	out := make([]string, 0, len(fields))
	seen := map[string]bool{}
	for _, f := range fields {
		s := strings.ToUpper(strings.TrimSpace(f))
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func finnhubExchange(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["exchange"]); v != "" {
		return strings.ToUpper(v)
	}
	return defaultExchange
}

func finnhubNewsCategory(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["market_news_category"]); v != "" {
		return strings.ToLower(v)
	}
	return defaultNewsCategory
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscan(strings.TrimSpace(s), &n)
	return n, err
}
