// Package openexchangerates implements the native pm Open Exchange Rates
// connector. It follows the declarative-HTTP template established by the stripe
// connector: a thin package that composes the connsdk toolkit (Requester +
// APIKeyQuery auth + RecordsAt/StringAt extraction) with Open Exchange Rates
// stream definitions and endpoints.
//
// The directory is named open-exchange-rates (hyphens preserved); the Go package
// identifier is openexchangerates and the registry key / Name() is the bare
// hyphenated system name "open-exchange-rates".
//
// The Open Exchange Rates API is a read-only foreign-exchange feed: there are no
// safe reverse-ETL writes, so the connector is read-only.
package openexchangerates

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
	oerDefaultBaseURL = "https://openexchangerates.org/api"
	oerUserAgent      = "polymetrics-go-cli"
	// oerMaxHistoricalDays bounds the historical date walk so a missing/empty
	// end_date plus a far-past start_date cannot fan out unbounded requests.
	oerMaxHistoricalDays = 366
	dateLayout           = "2006-01-02"
	// oerFixtureTimestamp is the deterministic timestamp used by fixture records
	// (2026-01-01T00:00:00Z in unix seconds).
	oerFixtureTimestamp int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("open-exchange-rates", New)
}

// New returns the Open Exchange Rates connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Open Exchange Rates connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
	// now is injectable for tests; defaults to time.Now. It bounds the historical
	// date walk's implicit end_date.
	now func() time.Time
}

func (Connector) Name() string { return "open-exchange-rates" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "open-exchange-rates",
		DisplayName:     "Open Exchange Rates",
		IntegrationType: "api",
		Description:     "Reads live, historical, and reference foreign-exchange rates from the Open Exchange Rates JSON API (read-only).",
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
	if _, err := oerBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(oerSecret(cfg)) == "" {
		return errors.New("open-exchange-rates connector requires secret app_id")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// usage.json confirms the app_id is valid without consuming a rate-data call
	// on most plans and mutates nothing.
	if err := r.DoJSON(ctx, http.MethodGet, "usage.json", nil, nil, nil); err != nil {
		return fmt.Errorf("check open-exchange-rates: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: oerStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an empty
// incremental cursor (the start_date config raises the lower bound at read time).
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
		stream = "latest"
	}
	switch stream {
	case "latest", "historical", "currencies", "usage":
	default:
		return fmt.Errorf("open-exchange-rates stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	switch stream {
	case "latest":
		return c.readLatest(ctx, r, emit)
	case "historical":
		return c.readHistorical(ctx, r, req, emit)
	case "currencies":
		return c.readCurrencies(ctx, r, emit)
	case "usage":
		return c.readUsage(ctx, r, emit)
	default:
		return fmt.Errorf("open-exchange-rates stream %q not found", stream)
	}
}

// Write is unsupported: Open Exchange Rates is a read-only feed. It satisfies the
// Connector interface and reports the standard unsupported-operation error.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) readLatest(ctx context.Context, r *connsdk.Requester, emit func(connectors.Record) error) error {
	payload, err := c.getObject(ctx, r, "latest.json")
	if err != nil {
		return fmt.Errorf("read open-exchange-rates latest: %w", err)
	}
	return rateRecords(payload, "", emit)
}

// readHistorical walks one request per calendar date from start_date to end_date
// (inclusive), hitting historical/{date}.json. This is the connector's
// multi-page path: the API has no body/link pagination, so date iteration is the
// page driver. The walk is bounded by oerMaxHistoricalDays.
func (c Connector) readHistorical(ctx context.Context, r *connsdk.Requester, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	start, end, err := c.historicalRange(req)
	if err != nil {
		return err
	}
	days := 0
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if days >= oerMaxHistoricalDays {
			return nil
		}
		days++
		date := day.Format(dateLayout)
		payload, err := c.getObject(ctx, r, "historical/"+date+".json")
		if err != nil {
			return fmt.Errorf("read open-exchange-rates historical %s: %w", date, err)
		}
		if err := rateRecords(payload, date, emit); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) readCurrencies(ctx context.Context, r *connsdk.Requester, emit func(connectors.Record) error) error {
	payload, err := c.getObject(ctx, r, "currencies.json")
	if err != nil {
		return fmt.Errorf("read open-exchange-rates currencies: %w", err)
	}
	return currencyRecords(payload, emit)
}

func (c Connector) readUsage(ctx context.Context, r *connsdk.Requester, emit func(connectors.Record) error) error {
	payload, err := c.getObject(ctx, r, "usage.json")
	if err != nil {
		return fmt.Errorf("read open-exchange-rates usage: %w", err)
	}
	return emit(usageRecord(payload))
}

// getObject fetches a JSON object endpoint. RecordsAt returns a single-element
// slice for a top-level object, which we unwrap to the underlying map.
func (c Connector) getObject(ctx context.Context, r *connsdk.Requester, path string) (map[string]any, error) {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if len(records) == 0 {
		return map[string]any{}, nil
	}
	return map[string]any(records[0]), nil
}

// historicalRange resolves the [start, end] date window for the historical walk.
// The lower bound comes from the incremental cursor (if any) else the start_date
// config; the upper bound comes from end_date config, else "today".
func (c Connector) historicalRange(req connectors.ReadRequest) (time.Time, time.Time, error) {
	startRaw := connsdk.Cursor(req.State)
	if startRaw == "" {
		startRaw = strings.TrimSpace(req.Config.Config["start_date"])
	}
	if startRaw == "" {
		return time.Time{}, time.Time{}, errors.New("open-exchange-rates historical stream requires config start_date (YYYY-MM-DD)")
	}
	start, err := time.Parse(dateLayout, startRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("open-exchange-rates start_date must be YYYY-MM-DD: %w", err)
	}

	endRaw := strings.TrimSpace(req.Config.Config["end_date"])
	var end time.Time
	if endRaw == "" {
		end = c.clock().UTC().Truncate(24 * time.Hour)
	} else {
		end, err = time.Parse(dateLayout, endRaw)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("open-exchange-rates end_date must be YYYY-MM-DD: %w", err)
		}
	}
	if end.Before(start) {
		end = start
	}
	return start, end, nil
}

func (c Connector) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	switch stream {
	case "currencies":
		payload := map[string]any{"EUR": "Euro", "GBP": "British Pound Sterling", "USD": "United States Dollar"}
		return currencyRecords(payload, emit)
	case "usage":
		payload := map[string]any{"status": int64(200), "data": map[string]any{
			"app_id": "oer_fixture",
			"status": "active",
			"plan":   map[string]any{"name": "Free"},
			"usage": map[string]any{
				"requests": int64(10), "requests_quota": int64(1000),
				"requests_remaining": int64(990), "days_elapsed": int64(1),
				"days_remaining": int64(29), "daily_average": int64(10),
			},
		}}
		return emit(usageRecord(payload))
	default:
		date := ""
		if stream == "historical" {
			date = "2026-01-01"
		}
		payload := map[string]any{
			"timestamp": oerFixtureTimestamp,
			"base":      "USD",
			"rates":     map[string]any{"EUR": 0.92, "GBP": 0.79, "JPY": 149.5},
		}
		return rateRecords(payload, date, emit)
	}
}

// requester builds a connsdk.Requester wired with app_id query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := oerBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := oerSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("open-exchange-rates connector requires secret app_id")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("app_id", secret),
		UserAgent: oerUserAgent,
	}, nil
}

func oerSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["app_id"]
}

// oerBaseURL resolves and validates the base URL. The default is
// openexchangerates.org/api; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func oerBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return oerDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("open-exchange-rates config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("open-exchange-rates config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("open-exchange-rates config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
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

// sortedKeys returns the keys of m in deterministic ascending order so flattened
// records (and tests) are stable.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
