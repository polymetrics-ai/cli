// Package exchangerates implements the native pm Exchange Rates connector for
// the exchangeratesapi.io REST API (catalog slug source-exchange-rates).
//
// It follows the declarative-HTTP shape of the stripe reference connector: a thin
// package composing the connsdk toolkit (Requester + APIKeyQuery auth + RecordsAt
// extraction) with Exchange-Rates-specific stream definitions and endpoints.
//
// The API has no list pagination; instead the core exchange_rates stream advances
// one ISO date at a time over the [start_date, end_date] window, which is this
// connector's natural unit of progress. It is read-only: the upstream is a pure
// data API with no reverse-ETL writes.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect. The bare system name "exchange-rates" is the
// registry key and Connector.Name; the Go package is named exchangerates because
// identifiers cannot contain a hyphen.
package exchangerates

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName            = "exchange-rates"
	exchangeRatesDefaultBase = "https://api.exchangeratesapi.io/v1"
	exchangeRatesUserAgent   = "polymetrics-go-cli"
	// maxDateWindow bounds how many days a single exchange_rates read will fetch
	// to avoid an unbounded sequence of requests when end_date is far ahead.
	maxDateWindow = 366
	isoDateLayout = "2006-01-02"
)

func init() {
	connectors.RegisterFactory(connectorName, New)
}

// New returns the Exchange Rates connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Exchange Rates connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Exchange Rates API",
		IntegrationType: "api",
		Description:     "Reads daily, latest, and supported-symbol foreign-exchange rates from the exchangeratesapi.io REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the API. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(accessKey(cfg)) == "" {
		return errors.New("exchange-rates connector requires secret access_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the symbols endpoint confirms auth and connectivity
	// without depending on any date being present.
	if err := r.DoJSON(ctx, http.MethodGet, "symbols", nil, nil, nil); err != nil {
		return fmt.Errorf("check exchange-rates: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: exchangeRatesStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: the exchange_rates stream
// starts with an empty incremental cursor, which the start_date config raises at
// read time.
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
		stream = streamExchangeRates
	}
	switch stream {
	case streamExchangeRates, streamLatest, streamSymbols:
	default:
		return fmt.Errorf("exchange-rates stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	switch stream {
	case streamSymbols:
		return c.readSymbols(ctx, r, emit)
	case streamLatest:
		return c.readLatest(ctx, r, req.Config, emit)
	default:
		return c.readExchangeRates(ctx, r, req, emit)
	}
}

// readExchangeRates iterates one ISO date at a time from the lower bound through
// end_date (or today), requesting GET /<YYYY-MM-DD> for each. This is the API's
// form of pagination: there is no list cursor, so progress is the date sequence.
func (c Connector) readExchangeRates(ctx context.Context, r *connsdk.Requester, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	start, err := lowerBoundDate(req)
	if err != nil {
		return err
	}
	end, err := endDate(req.Config)
	if err != nil {
		return err
	}
	if end.Before(start) {
		return nil
	}
	query := baseQuery(req.Config)
	ignoreWeekends := boolConfig(req.Config, "ignore_weekends", true)

	for day, count := start, 0; !day.After(end) && count < maxDateWindow; day = day.AddDate(0, 0, 1) {
		if err := ctx.Err(); err != nil {
			return err
		}
		count++
		if ignoreWeekends {
			switch day.Weekday() {
			case time.Saturday, time.Sunday:
				continue
			}
		}
		path := day.Format(isoDateLayout)
		var payload map[string]any
		if err := r.DoJSON(ctx, http.MethodGet, path, cloneValues(query), nil, &payload); err != nil {
			return fmt.Errorf("read exchange-rates %s: %w", path, err)
		}
		if err := apiError(payload); err != nil {
			return fmt.Errorf("read exchange-rates %s: %w", path, err)
		}
		if payload["date"] == nil {
			payload["date"] = path
		}
		if err := emit(rateRecord(payload)); err != nil {
			return err
		}
	}
	return nil
}

// readLatest fetches the single most-recent rates record.
func (c Connector) readLatest(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	var payload map[string]any
	if err := r.DoJSON(ctx, http.MethodGet, "latest", baseQuery(cfg), nil, &payload); err != nil {
		return fmt.Errorf("read exchange-rates latest: %w", err)
	}
	if err := apiError(payload); err != nil {
		return fmt.Errorf("read exchange-rates latest: %w", err)
	}
	return emit(rateRecord(payload))
}

// readSymbols fetches the supported-currency map and emits one record per code,
// in deterministic (sorted) order.
func (c Connector) readSymbols(ctx context.Context, r *connsdk.Requester, emit func(connectors.Record) error) error {
	var payload struct {
		Success bool           `json:"success"`
		Error   any            `json:"error"`
		Symbols map[string]any `json:"symbols"`
	}
	if err := r.DoJSON(ctx, http.MethodGet, "symbols", nil, nil, &payload); err != nil {
		return fmt.Errorf("read exchange-rates symbols: %w", err)
	}
	if payload.Error != nil {
		return fmt.Errorf("read exchange-rates symbols: %v", payload.Error)
	}
	codes := make([]string, 0, len(payload.Symbols))
	for code := range payload.Symbols {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	for _, code := range codes {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(symbolRecord(code, payload.Symbols[code])); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	switch stream {
	case streamSymbols:
		for _, code := range []string{"EUR", "USD"} {
			if err := ctx.Err(); err != nil {
				return err
			}
			name := map[string]string{"EUR": "Euro", "USD": "United States Dollar"}[code]
			if err := emit(symbolRecord(code, name)); err != nil {
				return err
			}
		}
		return nil
	case streamLatest:
		return emit(fixtureRate("2026-01-02"))
	default:
		for _, date := range []string{"2026-01-01", "2026-01-02"} {
			if err := ctx.Err(); err != nil {
				return err
			}
			record := fixtureRate(date)
			if cursor := req.State["cursor"]; cursor != "" {
				record["previous_cursor"] = cursor
			}
			if err := emit(record); err != nil {
				return err
			}
		}
		return nil
	}
}

func fixtureRate(date string) connectors.Record {
	return rateRecord(map[string]any{
		"success":    true,
		"historical": true,
		"date":       date,
		"timestamp":  int64(1767225600),
		"base":       "EUR",
		"rates":      map[string]any{"USD": 1.05, "GBP": 0.85, "JPY": 160.0},
	})
}

// requester builds a connsdk.Requester wired with the access_key query-param auth
// and the resolved base URL. The secret only ever flows into connsdk.APIKeyQuery;
// it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := accessKey(cfg)
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("exchange-rates connector requires secret access_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("access_key", key),
		UserAgent: exchangeRatesUserAgent,
	}, nil
}

// baseQuery builds the per-request query params shared by rate endpoints (the
// optional source base currency).
func baseQuery(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	if base := strings.TrimSpace(cfg.Config["base"]); base != "" {
		q.Set("base", base)
	}
	return q
}

// lowerBoundDate returns the first date to fetch, derived from the incremental
// cursor (if any) or else the start_date config.
func lowerBoundDate(req connectors.ReadRequest) (time.Time, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		t, err := time.Parse(isoDateLayout, cursor)
		if err != nil {
			return time.Time{}, fmt.Errorf("exchange-rates cursor must be YYYY-MM-DD: %w", err)
		}
		// Resume from the day after the last synced date.
		return t.AddDate(0, 0, 1), nil
	}
	start := strings.TrimSpace(req.Config.Config["start_date"])
	if start == "" {
		return time.Time{}, errors.New("exchange-rates connector requires config start_date (YYYY-MM-DD)")
	}
	t, err := time.Parse(isoDateLayout, start)
	if err != nil {
		return time.Time{}, fmt.Errorf("exchange-rates config start_date must be YYYY-MM-DD: %w", err)
	}
	return t, nil
}

// endDate returns the inclusive upper bound for the date window: the end_date
// config if present, otherwise today (UTC).
func endDate(cfg connectors.RuntimeConfig) (time.Time, error) {
	end := strings.TrimSpace(cfg.Config["end_date"])
	if end == "" {
		now := time.Now().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	}
	t, err := time.Parse(isoDateLayout, end)
	if err != nil {
		return time.Time{}, fmt.Errorf("exchange-rates config end_date must be YYYY-MM-DD: %w", err)
	}
	return t, nil
}

func accessKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_key"]
}

// baseURL resolves and validates the base URL. The default is
// api.exchangeratesapi.io; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return exchangeRatesDefaultBase, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("exchange-rates config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("exchange-rates config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("exchange-rates config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// apiError surfaces the API's {"success":false,"error":{...}} envelope as a Go
// error so a failed day does not look like an empty success.
func apiError(payload map[string]any) error {
	if payload == nil {
		return nil
	}
	if errVal, ok := payload["error"]; ok && errVal != nil {
		return fmt.Errorf("api error: %v", errVal)
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func boolConfig(cfg connectors.RuntimeConfig, key string, def bool) bool {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config[key]))
	switch raw {
	case "":
		return def
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return def
	}
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

// Write satisfies the Connector interface. Exchange Rates is a read-only data API
// with no reverse-ETL surface, so writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
