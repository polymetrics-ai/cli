// Package googleanalyticsdataapi implements the native pm Google Analytics 4
// (GA4) Data API connector. It is a declarative-HTTP per-system connector built
// on the same shape as the stripe reference: a thin package that composes the
// connsdk toolkit (Requester + Bearer/OAuth2 access-token auth + JSON extraction)
// with GA4-specific report definitions and the runReport endpoint.
//
// GA4 has no fixed REST resources; reporting is a POST runReport call that takes
// a dimension x metric query and returns rows. Each published "stream" is a
// canned report spec (see streams.go); a row is flattened to a record by
// projecting dimensionHeaders/metricHeaders onto the row's
// dimensionValues/metricValues.
//
// Like github/stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// The connector is read-only: the GA4 Data API exposes no safe reverse-ETL
// writes, so Capabilities.Write is false.
package googleanalyticsdataapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	connectorName        = "google-analytics-data-api"
	gaDefaultBaseURL     = "https://analyticsdata.googleapis.com"
	gaAPIVersion         = "v1beta"
	gaDefaultPageSize    = 10000
	gaMaxPageSize        = 250000
	gaUserAgent          = "polymetrics-go-cli"
	gaDefaultStartDate   = "30daysAgo"
	gaDefaultEndDate     = "today"
	gaFixturePropertyID  = "000000000"
	gaFixtureRecordCount = 2
)

func init() {
	connectors.RegisterFactory(connectorName, New)
}

// New returns the GA4 Data API connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Google Analytics 4 Data API connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Google Analytics 4 (GA4)",
		IntegrationType: "api",
		Description:     "Reads Google Analytics 4 reports (active users, traffic sources, devices, pages) from the Analytics Data API runReport endpoint. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the GA4 Data
// API. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(gaSecret(cfg)) == "" {
		return errors.New("google-analytics-data-api connector requires an OAuth2 access token (secret credentials.access_token)")
	}
	property, err := gaPropertyID(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded one-row runReport confirms auth, the property id, and
	// connectivity without mutating anything.
	spec := gaReports["daily_active_users"]
	body := buildReportBody(spec, cfg, 0, 1)
	path := reportPath(property)
	if err := r.DoJSON(ctx, http.MethodPost, path, nil, body, nil); err != nil {
		return fmt.Errorf("check google-analytics-data-api: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gaStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The GA4 Data API is a
// reporting (read) API with no safe reverse-ETL writes, so writes are
// unsupported and Capabilities.Write is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a GA4 stream starts with an
// empty cursor; the supported_sync_modes are full_refresh, but date-dimensioned
// reports can carry a "date" cursor the start_date config raises at read time.
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
		stream = "daily_active_users"
	}
	spec, ok := gaReports[stream]
	if !ok {
		return fmt.Errorf("google-analytics-data-api stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, spec, req, emit)
	}

	property, err := gaPropertyID(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := gaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gaMaxPages(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, spec, property, req.Config, pageSize, maxPages, emit)
}

// harvest drives GA4 offset/limit pagination. runReport returns
// {dimensionHeaders, metricHeaders, rows, rowCount}; each page is requested with
// offset advanced by limit until offset >= rowCount (or rows run out). The loop
// lives here because the offset paginator is body-driven and report-specific.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, spec reportSpec, property string, cfg connectors.RuntimeConfig, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := reportPath(property)
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := buildReportBody(spec, cfg, offset, pageSize)
		resp, err := r.Do(ctx, http.MethodPost, path, nil, body)
		if err != nil {
			return fmt.Errorf("read google-analytics-data-api %s: %w", spec.name, err)
		}
		report, err := decodeReport(resp.Body)
		if err != nil {
			return fmt.Errorf("decode google-analytics-data-api %s page: %w", spec.name, err)
		}
		for _, row := range report.Rows {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRow(property, report, row)); err != nil {
				return err
			}
		}
		emitted := offset + len(report.Rows)
		// Stop when we've consumed all reported rows or the page came back
		// short (defensive against a missing/zero rowCount).
		if len(report.Rows) == 0 || (report.RowCount > 0 && emitted >= report.RowCount) || len(report.Rows) < pageSize {
			return nil
		}
		offset = emitted
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors
// stripe's fixture intent).
func (c Connector) readFixture(ctx context.Context, spec reportSpec, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 0; i < gaFixtureRecordCount; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := connectors.Record{
			"property_id": gaFixturePropertyID,
			"connector":   connectorName,
			"fixture":     true,
		}
		for di, dim := range spec.dimensions {
			if dim == "date" {
				record[dim] = fmt.Sprintf("2026010%d", i+1)
				continue
			}
			record[dim] = fmt.Sprintf("%s_fixture_%d", dim, i+1)
			_ = di
		}
		for mi, metric := range spec.metrics {
			record[metric] = strconv.Itoa((i + 1) * (mi + 1) * 10)
		}
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth (the GA4 OAuth2
// access token), the resolved base URL, and a JSON content type. The secret only
// ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := gaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := gaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("google-analytics-data-api connector requires an OAuth2 access token (secret credentials.access_token)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: gaUserAgent,
	}, nil
}

// reportPath builds the runReport endpoint path for a property id, e.g.
// "v1beta/properties/123456:runReport".
func reportPath(property string) string {
	return fmt.Sprintf("%s/properties/%s:runReport", gaAPIVersion, property)
}

// buildReportBody constructs the runReport request body for a report spec at the
// given offset/limit, applying the configured date range.
func buildReportBody(spec reportSpec, cfg connectors.RuntimeConfig, offset, limit int) map[string]any {
	dims := make([]map[string]string, 0, len(spec.dimensions))
	for _, d := range spec.dimensions {
		dims = append(dims, map[string]string{"name": d})
	}
	metrics := make([]map[string]string, 0, len(spec.metrics))
	for _, m := range spec.metrics {
		metrics = append(metrics, map[string]string{"name": m})
	}
	start, end := gaDateRange(cfg)
	return map[string]any{
		"dimensions":    dims,
		"metrics":       metrics,
		"dateRanges":    []map[string]string{{"startDate": start, "endDate": end}},
		"offset":        offset,
		"limit":         limit,
		"keepEmptyRows": false,
	}
}

func gaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	// The catalog declares the secret under nested credentials.* keys; accept the
	// access_token under both the dotted and bare forms (and a couple of common
	// aliases) so callers can flatten the vault however they like.
	for _, key := range []string{
		"credentials.access_token",
		"access_token",
		"credentials.api_key",
		"api_key",
	} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// gaPropertyID resolves the single GA4 property id to read. The catalog's
// property_ids config is a list; this connector reads the first id (callers can
// run one stream-read per property). The id may be given bare ("123456") or
// prefixed ("properties/123456").
func gaPropertyID(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(firstNonEmpty(cfg.Config["property_ids"], cfg.Config["property_id"]))
	if raw == "" {
		return "", errors.New("google-analytics-data-api connector requires config property_ids")
	}
	// property_ids may be a comma/space separated list; take the first entry.
	for _, sep := range []string{",", " ", "\n"} {
		if i := strings.IndexAny(raw, sep); i >= 0 {
			raw = raw[:i]
			break
		}
	}
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "properties/")
	if raw == "" {
		return "", errors.New("google-analytics-data-api config property_ids is empty")
	}
	for _, r := range raw {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("google-analytics-data-api property id must be numeric, got %q", raw)
		}
	}
	return raw, nil
}

// gaDateRange resolves the report date range from config, defaulting to the last
// 30 days. GA4 accepts YYYY-MM-DD or relative tokens (NdaysAgo, today,
// yesterday), which are passed through unchanged.
func gaDateRange(cfg connectors.RuntimeConfig) (string, string) {
	start := strings.TrimSpace(firstNonEmpty(cfg.Config["date_ranges_start_date"], cfg.Config["start_date"]))
	if start == "" {
		start = gaDefaultStartDate
	}
	end := strings.TrimSpace(firstNonEmpty(cfg.Config["date_ranges_end_date"], cfg.Config["end_date"]))
	if end == "" {
		end = gaDefaultEndDate
	}
	return start, end
}

// gaBaseURL resolves and validates the base URL. The default is
// analyticsdata.googleapis.com; any override must be an absolute https (or http
// for local test servers) URL with a host to bound SSRF risk.
func gaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return gaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("google-analytics-data-api config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("google-analytics-data-api config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("google-analytics-data-api config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func gaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return gaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-analytics-data-api config page_size must be an integer: %w", err)
	}
	if value < 1 || value > gaMaxPageSize {
		return 0, fmt.Errorf("google-analytics-data-api config page_size must be between 1 and %d", gaMaxPageSize)
	}
	return value, nil
}

func gaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("google-analytics-data-api config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("google-analytics-data-api config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
