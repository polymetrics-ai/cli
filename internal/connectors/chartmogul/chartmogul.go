// Package chartmogul implements the native pm ChartMogul connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + HTTP Basic auth +
// RecordsAt extraction + cursor pagination) with ChartMogul-specific stream
// definitions and endpoints.
//
// ChartMogul authenticates with HTTP Basic, sending the API key as the username
// and an empty password. List endpoints (customers, activities) paginate with a
// body cursor + has_more flag over an entries[] array; metric endpoints return a
// single entries[] page. The connector is read-only.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package chartmogul

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	chartmogulDefaultBaseURL  = "https://api.chartmogul.com/v1"
	chartmogulDefaultPageSize = 200
	chartmogulMaxPageSize     = 200
	chartmogulUserAgent       = "polymetrics-go-cli"
	// chartmogulMetricsEndDate bounds the metrics window when none is supplied.
	chartmogulFixtureDate = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("chartmogul", New)
}

// New returns the ChartMogul connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm ChartMogul connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "chartmogul" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "chartmogul",
		DisplayName:     "ChartMogul",
		IntegrationType: "api",
		Description:     "Reads ChartMogul customers, subscription activities, customer-count metrics, and account details through the ChartMogul REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to ChartMogul.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := chartmogulBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(chartmogulSecret(cfg)) == "" {
		return errors.New("chartmogul connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// /account is a cheap, read-only endpoint that confirms auth + connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "account", nil, nil, nil); err != nil {
		return fmt.Errorf("check chartmogul: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: chartmogulStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader. ChartMogul list endpoints do
// not support incremental filtering by update time, so each stream starts with an
// empty cursor (full refresh), which the start_date config can raise at read time
// for the activities and metrics streams.
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
		stream = "customers"
	}
	endpoint, ok := chartmogulStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("chartmogul stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := chartmogulMaxPages(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := chartmogulPageSize(req.Config)
	if err != nil {
		return err
	}

	switch endpoint.pagination {
	case pageObject, pageSingle:
		return c.readSingle(ctx, r, stream, endpoint, req, emit)
	default:
		return c.harvest(ctx, r, stream, endpoint, req, pageSize, maxPages, emit)
	}
}

// harvest drives ChartMogul's cursor pagination. List responses are
// {entries:[...], has_more:bool, cursor:"..."}; the next page is requested with
// cursor=<cursor>. There is no body-cursor paginator in connsdk for this exact
// has_more shape, so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, stream string, endpoint streamEndpoint, req connectors.ReadRequest, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	// The activities stream supports a start-date lower bound.
	if stream == "activities" {
		if lower := startDate(req); lower != "" {
			base.Set("start-date", lower)
		}
	}

	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read chartmogul %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "entries")
		if err != nil {
			return fmt.Errorf("decode chartmogul %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return fmt.Errorf("decode chartmogul %s has_more: %w", endpoint.resource, err)
		}
		nextCursor, err := connsdk.StringAt(resp.Body, "cursor")
		if err != nil {
			return fmt.Errorf("decode chartmogul %s cursor: %w", endpoint.resource, err)
		}
		if hasMore != "true" || strings.TrimSpace(nextCursor) == "" {
			return nil
		}
		cursor = nextCursor
	}
	return nil
}

// readSingle reads endpoints that return all rows in one response: metrics
// (entries[]) and account (a single object). RecordsAt handles both an entries[]
// array and a single object.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	query := url.Values{}
	recordsPath := ""
	if endpoint.pagination == pageSingle {
		recordsPath = "entries"
		// Metrics endpoints require a start-date/end-date window.
		start := startDate(req)
		if start == "" {
			start = chartmogulFixtureDate
		}
		query.Set("start-date", metricsDate(start))
		query.Set("end-date", metricsDate(time.Now().UTC().Format(time.RFC3339)))
		if endpoint.metricsInterval != "" {
			query.Set("interval", endpoint.metricsInterval)
		}
	}
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read chartmogul %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, recordsPath)
	if err != nil {
		return fmt.Errorf("decode chartmogul %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise chartmogul credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"uuid":                  fmt.Sprintf("%s_fixture_%d", stream, i),
			"external_id":           fmt.Sprintf("ext_%d", i),
			"name":                  fmt.Sprintf("Fixture %d", i),
			"email":                 fmt.Sprintf("fixture+%d@example.com", i),
			"status":                "Active",
			"customer-since":        chartmogulFixtureDate,
			"company":               "Fixture Co",
			"country":               "US",
			"city":                  "San Francisco",
			"currency":              "USD",
			"mrr":                   int64(1000 * i),
			"arr":                   int64(12000 * i),
			"billing-system-type":   "Stripe",
			"type":                  "new_biz",
			"date":                  chartmogulFixtureDate,
			"activity-mrr":          int64(1000 * i),
			"activity-mrr-movement": int64(1000 * i),
			"activity-arr":          int64(12000 * i),
			"description":           "purchased subscription",
			"customer-uuid":         "customers_fixture_1",
			"customer-name":         "Fixture 1",
			"customer-external-id":  "ext_1",
			"customers":             int64(10 + i),
			"percentage-change":     0.0,
			"time_zone":             "UTC",
			"week_start_on":         "monday",
		}
		// The metrics stream keys on date, so make fixture dates distinct.
		if stream == "customer_count" {
			item["date"] = fmt.Sprintf("2026-0%d-01", i)
		}
		record := endpoint.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth (api_key as
// username, empty password) and the resolved base URL. The secret only ever flows
// into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := chartmogulBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := chartmogulSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("chartmogul connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: chartmogulUserAgent,
	}, nil
}

func chartmogulSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// chartmogulBaseURL resolves and validates the base URL. The default is
// api.chartmogul.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func chartmogulBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return chartmogulDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("chartmogul config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("chartmogul config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("chartmogul config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func chartmogulPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return chartmogulDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("chartmogul config page_size must be an integer: %w", err)
	}
	if value < 1 || value > chartmogulMaxPageSize {
		return 0, fmt.Errorf("chartmogul config page_size must be between 1 and %d", chartmogulMaxPageSize)
	}
	return value, nil
}

func chartmogulMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("chartmogul config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("chartmogul config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// startDate returns the configured start_date (RFC3339), preferring an
// incremental cursor when present.
func startDate(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

// metricsDate reduces an RFC3339 timestamp to the YYYY-MM-DD form ChartMogul
// metrics endpoints expect. Non-timestamp values are returned as-is.
func metricsDate(value string) string {
	value = strings.TrimSpace(value)
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC().Format("2006-01-02")
	}
	if len(value) >= 10 {
		return value[:10]
	}
	return value
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the Connector interface. ChartMogul is read-only in pm; writes
// are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
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
