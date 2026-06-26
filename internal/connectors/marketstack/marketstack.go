// Package marketstack implements the native pm Marketstack connector. It is a
// declarative-HTTP per-system connector in the shape of the stripe reference:
// a thin package that composes the connsdk toolkit (Requester + api-key-in-query
// auth + RecordsAt extraction + offset/limit pagination) with Marketstack-specific
// stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Marketstack serves read-only financial market data (no reverse-ETL writes make
// sense), so the connector is read-only: Capabilities.Write is false.
package marketstack

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
	marketstackDefaultBaseURL  = "https://api.marketstack.com/v1"
	marketstackDefaultPageSize = 100
	marketstackMaxPageSize     = 1000
	marketstackUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("marketstack", New)
}

// New returns the Marketstack connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Marketstack connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "marketstack" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "marketstack",
		DisplayName:     "Marketstack",
		IntegrationType: "api",
		Description:     "Reads Marketstack exchanges, tickers, end-of-day prices, splits, and dividends through the Marketstack REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Marketstack.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := marketstackBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(marketstackSecret(cfg)) == "" {
		return errors.New("marketstack connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the exchanges list confirms auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "exchanges", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check marketstack: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: marketstackStreams()}, nil
}

// Write is unsupported: Marketstack is a read-only market-data source. It
// satisfies the connectors.Connector interface while advertising Write=false in
// Metadata.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Marketstack stream starts
// with an empty incremental cursor (full sync), which the start_date config can
// raise at read time for the date-cursored streams.
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
		stream = "exchanges"
	}
	endpoint, ok := marketstackStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("marketstack stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := marketstackPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := marketstackMaxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{}
	if endpoint.acceptsSymbols {
		if symbols := strings.TrimSpace(req.Config.Config["symbols"]); symbols != "" {
			base.Set("symbols", symbols)
		}
		if lower := incrementalLowerBound(req); lower != "" {
			base.Set("date_from", lower)
		}
	}
	return c.harvest(ctx, r, endpoint, base, pageSize, maxPages, emit)
}

// harvest drives Marketstack's offset/limit pagination. Each list response is
// {pagination:{limit,offset,count,total}, data:[...]}. We request fixed-size
// pages, advancing offset until a short page (fewer records than the page size)
// is returned. Built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read marketstack %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode marketstack %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (or empty page) means we have reached the end.
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise marketstack credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"mic":          fmt.Sprintf("XFIX%d", i),
			"symbol":       fmt.Sprintf("FIX%d", i),
			"name":         fmt.Sprintf("Fixture %d", i),
			"acronym":      "FIX",
			"country":      "United States",
			"country_code": "US",
			"city":         "New York",
			"website":      "https://example.com",
			"has_eod":      true,
			"has_intraday": false,
			"date":         fmt.Sprintf("2026-01-0%dT00:00:00+0000", i),
			"exchange":     "XFIX",
			"open":         100.0 + float64(i),
			"high":         101.0 + float64(i),
			"low":          99.0 + float64(i),
			"close":        100.5 + float64(i),
			"volume":       1000.0 * float64(i),
			"adj_open":     100.0 + float64(i),
			"adj_high":     101.0 + float64(i),
			"adj_low":      99.0 + float64(i),
			"adj_close":    100.5 + float64(i),
			"adj_volume":   1000.0 * float64(i),
			"split_factor": 1.0,
			"dividend":     0.5 * float64(i),
			"currency":     map[string]any{"code": "USD", "name": "US Dollar", "symbol": "$"},
			"timezone":     map[string]any{"timezone": "America/New_York", "abbr": "EST"},
		}
		record := endpoint.mapRecord(item)
		record["fixture"] = true
		if cursor := connsdk.Cursor(req.State); cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with api-key-in-query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery (added
// to the request URL as access_key=...); it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := marketstackBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := marketstackSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("marketstack connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("access_key", secret),
		UserAgent: marketstackUserAgent,
	}, nil
}

// incrementalLowerBound returns the date_from lower bound, derived from the
// incremental cursor (if any) or else the start_date config. Marketstack expects
// YYYY-MM-DD; an RFC3339 start_date is truncated to its date portion. An empty
// result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return dateOnly(cursor)
	}
	if startDate := strings.TrimSpace(req.Config.Config["start_date"]); startDate != "" {
		return dateOnly(startDate)
	}
	return ""
}

func dateOnly(value string) string {
	value = strings.TrimSpace(value)
	if i := strings.IndexAny(value, "T "); i > 0 {
		return value[:i]
	}
	return value
}

func marketstackSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// marketstackBaseURL resolves and validates the base URL. The default is
// api.marketstack.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func marketstackBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return marketstackDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("marketstack config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("marketstack config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("marketstack config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func marketstackPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return marketstackDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("marketstack config page_size must be an integer: %w", err)
	}
	if value < 1 || value > marketstackMaxPageSize {
		return 0, fmt.Errorf("marketstack config page_size must be between 1 and %d", marketstackMaxPageSize)
	}
	return value, nil
}

func marketstackMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("marketstack config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("marketstack config max_pages must be 0 for unlimited or a positive integer")
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
