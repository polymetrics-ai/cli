// Package alphavantage implements the native pm Alpha Vantage connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// api-key query auth) with Alpha Vantage-specific stream definitions.
//
// Alpha Vantage is a market-data API. Every request goes to a single /query
// endpoint; the desired data set is selected by a `function` query param
// (TIME_SERIES_DAILY, GLOBAL_QUOTE, ...). Auth is a single `apikey` query param.
// There is no pagination: the time-series endpoints return one JSON object keyed
// by date strings (compact = last 100 points, full = the entire history), and
// GLOBAL_QUOTE returns one flat object. Because the payload is a date-keyed
// object rather than an array, the read loop decodes the body here and flattens
// each dated entry into a record, injecting the symbol and date columns.
//
// On error or rate-limit the API still returns HTTP 200 with an "Error Message",
// "Note", or "Information" field; the connector surfaces those as errors so a
// throttled or invalid request does not look like an empty result.
//
// Like github/stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// The connector is read-only: Alpha Vantage exposes no reverse-ETL writes, so
// Capabilities.Write is false.
package alphavantage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	alphaVantageDefaultBaseURL  = "https://www.alphavantage.co"
	alphaVantageQueryPath       = "query"
	alphaVantageUserAgent       = "polymetrics-go-cli"
	alphaVantageDefaultSymbol   = "IBM"
	alphaVantageDefaultInterval = "1min"
)

func init() {
	connectors.RegisterFactory("alpha-vantage", New)
}

// New returns the Alpha Vantage connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Alpha Vantage connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "alpha-vantage" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "alpha-vantage",
		DisplayName:     "Alpha Vantage",
		IntegrationType: "api",
		Description:     "Reads Alpha Vantage daily, weekly, monthly, and intraday OHLCV time series plus the latest global quote for a configured stock symbol.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Alpha
// Vantage. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := alphaVantageBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(alphaVantageSecret(cfg)) == "" {
		return errors.New("alpha-vantage connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded GLOBAL_QUOTE read confirms auth and connectivity without pulling
	// a full history. A throttled or invalid response is surfaced as an error.
	query := url.Values{}
	query.Set("function", "GLOBAL_QUOTE")
	query.Set("symbol", alphaVantageSymbol(cfg))
	resp, err := r.Do(ctx, http.MethodGet, alphaVantageQueryPath, query, nil)
	if err != nil {
		return fmt.Errorf("check alpha-vantage: %w", err)
	}
	if err := apiError(resp.Body); err != nil {
		return fmt.Errorf("check alpha-vantage: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: alphaVantageStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an empty
// incremental cursor (full sync); subsequent runs carry the highest date seen.
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
		stream = "time_series_daily"
	}
	spec, ok := alphaVantageStreamSpecs[stream]
	if !ok {
		return fmt.Errorf("alpha-vantage stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, spec, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	symbol := alphaVantageSymbol(req.Config)
	query := url.Values{}
	query.Set("function", spec.function)
	query.Set("symbol", symbol)
	if spec.intraday {
		query.Set("interval", alphaVantageInterval(req.Config))
		if adjusted := strings.TrimSpace(req.Config.Config["adjusted"]); adjusted != "" {
			query.Set("adjusted", adjusted)
		}
	}
	if outputsize := strings.TrimSpace(req.Config.Config["outputsize"]); outputsize != "" {
		query.Set("outputsize", outputsize)
	}

	resp, err := r.Do(ctx, http.MethodGet, alphaVantageQueryPath, query, nil)
	if err != nil {
		return fmt.Errorf("read alpha-vantage %s: %w", stream, err)
	}
	if err := apiError(resp.Body); err != nil {
		return fmt.Errorf("read alpha-vantage %s: %w", stream, err)
	}

	return emitFromBody(resp.Body, spec, symbol, req, emit)
}

// emitFromBody decodes the Alpha Vantage response and flattens it into records.
// For quote streams it emits the single object; for time-series streams it
// iterates the date-keyed series object (resolving the intraday key from the
// interval), emitting entries in descending date order so the newest bar comes
// first (matching the API's own ordering).
func emitFromBody(body []byte, spec streamSpec, symbol string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	var payload map[string]json.RawMessage
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return fmt.Errorf("decode alpha-vantage response: %w", err)
	}

	if spec.seriesKey == "" && spec.objectKey != "" {
		raw, ok := payload[spec.objectKey]
		if !ok {
			return fmt.Errorf("alpha-vantage response missing %q", spec.objectKey)
		}
		var quote map[string]any
		if err := decodeRaw(raw, &quote); err != nil {
			return err
		}
		if len(quote) == 0 {
			return nil
		}
		return emit(globalQuoteRecord(quote))
	}

	seriesKey := spec.seriesKey
	if spec.intraday {
		seriesKey = fmt.Sprintf("Time Series (%s)", alphaVantageInterval(req.Config))
	}
	raw, ok := payload[seriesKey]
	if !ok {
		return fmt.Errorf("alpha-vantage response missing series %q", seriesKey)
	}
	var series map[string]map[string]any
	if err := decodeRaw(raw, &series); err != nil {
		return err
	}

	dates := make([]string, 0, len(series))
	for date := range series {
		dates = append(dates, date)
	}
	// Descending so the most recent bar is emitted first, mirroring the API.
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	lower := strings.TrimSpace(connsdk.Cursor(req.State))
	for _, date := range dates {
		if lower != "" && date <= lower {
			continue
		}
		if err := emit(ohlcvRecord(symbol, date, series[date])); err != nil {
			return err
		}
	}
	return nil
}

func decodeRaw(raw json.RawMessage, out any) error {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode alpha-vantage entry: %w", err)
	}
	return nil
}

// apiError inspects an Alpha Vantage 200 body for the error/throttle envelopes
// the API uses instead of HTTP error codes. It returns nil when none are present.
func apiError(body []byte) error {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(body, &envelope); err != nil {
		// A non-object body (or invalid JSON) is handled by the decode in
		// emitFromBody; here we only flag the known string-valued error fields.
		return nil
	}
	for _, key := range []string{"Error Message", "Note", "Information"} {
		if raw, ok := envelope[key]; ok {
			var msg string
			if err := json.Unmarshal(raw, &msg); err == nil && strings.TrimSpace(msg) != "" {
				return fmt.Errorf("alpha-vantage api %s: %s", key, msg)
			}
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise alpha-vantage credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, spec streamSpec, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	symbol := alphaVantageSymbol(req.Config)
	if spec.seriesKey == "" && spec.objectKey != "" {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := globalQuoteRecord(map[string]any{
			"01. symbol":             symbol,
			"02. open":               "100.0000",
			"03. high":               "110.0000",
			"04. low":                "99.0000",
			"05. price":              "105.0000",
			"06. volume":             "12345",
			"07. latest trading day": "2026-01-02",
			"08. previous close":     "104.0000",
			"09. change":             "1.0000",
			"10. change percent":     "0.9615%",
		})
		return emit(rec)
	}

	// Two deterministic dated bars per time-series stream.
	dates := []string{"2026-01-02", "2026-01-01"}
	for i, date := range dates {
		if err := ctx.Err(); err != nil {
			return err
		}
		entry := map[string]any{
			"1. open":   fmt.Sprintf("%d.0000", 100+i),
			"2. high":   fmt.Sprintf("%d.0000", 110+i),
			"3. low":    fmt.Sprintf("%d.0000", 99+i),
			"4. close":  fmt.Sprintf("%d.0000", 105+i),
			"5. volume": fmt.Sprintf("%d", 12345+i),
		}
		rec := ohlcvRecord(symbol, date, entry)
		if cursor := req.State["cursor"]; cursor != "" {
			rec["previous_cursor"] = cursor
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with api-key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := alphaVantageBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := alphaVantageSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("alpha-vantage connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("apikey", secret),
		UserAgent: alphaVantageUserAgent,
	}, nil
}

func alphaVantageSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func alphaVantageSymbol(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if sym := strings.TrimSpace(cfg.Config["symbol"]); sym != "" {
			return sym
		}
	}
	return alphaVantageDefaultSymbol
}

func alphaVantageInterval(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if iv := strings.TrimSpace(cfg.Config["interval"]); iv != "" {
			return iv
		}
	}
	return alphaVantageDefaultInterval
}

// alphaVantageBaseURL resolves and validates the base URL. The default is
// www.alphavantage.co; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func alphaVantageBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return alphaVantageDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("alpha-vantage config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("alpha-vantage config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("alpha-vantage config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Alpha Vantage is a read-only market-data API.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
