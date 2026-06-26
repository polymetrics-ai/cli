// Package datadog implements the native pm Datadog source connector. It follows
// the declarative-HTTP template established by the stripe package: a thin package
// that composes the connsdk toolkit (Requester + RecordsAt extraction) with
// Datadog-specific stream definitions, endpoints, and pagination.
//
// Datadog authenticates with two headers, DD-API-KEY and DD-APPLICATION-KEY,
// rather than a single bearer token, so auth is applied via the Requester's
// DefaultHeaders. The connector is read-only: it exposes monitors, dashboards,
// users, SLOs, and downtimes. Like stripe, it self-registers via RegisterFactory
// in init(); the registryset package blank-imports this package in the
// production binary to run that side effect.
package datadog

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
	datadogDefaultBaseURL  = "https://api.datadoghq.com"
	datadogDefaultPageSize = 100
	datadogMaxPageSize     = 1000
	datadogUserAgent       = "polymetrics-go-cli"
	apiKeyHeader           = "DD-API-KEY"
	appKeyHeader           = "DD-APPLICATION-KEY"
)

func init() {
	connectors.RegisterFactory("datadog", New)
}

// New returns the Datadog connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Datadog source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "datadog" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "datadog",
		DisplayName:     "Datadog",
		IntegrationType: "api",
		Description:     "Reads Datadog monitors, dashboards, users, SLOs, and downtimes through the Datadog REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Datadog. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := datadogBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(datadogAPIKey(cfg)) == "" {
		return errors.New("datadog connector requires secret api_key")
	}
	if strings.TrimSpace(datadogAppKey(cfg)) == "" {
		return errors.New("datadog connector requires secret application_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the monitors list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"page": []string{"0"}, "page_size": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "api/v1/monitor", q, nil, nil); err != nil {
		return fmt.Errorf("check datadog: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: datadogStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "monitors"
	}
	endpoint, ok := datadogStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("datadog stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := datadogPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := datadogMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Datadog's per-endpoint pagination. The three styles
// (pageNone/pageV1/pageV2) are handled here on top of connsdk.Requester +
// connsdk.RecordsAt. A page is "full" when it returns exactly pageSize records;
// a shorter page ends the loop.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := pageQuery(endpoint.page, page, pageSize)
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read datadog %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode datadog %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Unpaginated endpoints return everything at once; stop after one page.
		if endpoint.page == pageNone {
			return nil
		}
		// A short (or empty) page means there are no more records.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// pageQuery builds the query params for the requested 0-based page under the
// given pagination style. pageNone returns no params.
func pageQuery(style pageStyle, page, pageSize int) url.Values {
	q := url.Values{}
	switch style {
	case pageV1:
		q.Set("page", strconv.Itoa(page))
		q.Set("page_size", strconv.Itoa(pageSize))
	case pageV2:
		q.Set("page[number]", strconv.Itoa(page))
		q.Set("page[size]", strconv.Itoa(pageSize))
	}
	return q
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise datadog credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":          fmt.Sprintf("Fixture %s %d", stream, i),
			"title":         fmt.Sprintf("Fixture %s %d", stream, i),
			"type":          "metric alert",
			"query":         "avg(last_5m):avg:system.cpu.user{*} > 90",
			"message":       "fixture monitor",
			"overall_state": "OK",
			"created":       "2026-01-01T00:00:00Z",
			"modified":      "2026-01-01T00:00:00Z",
			"created_at":    "2026-01-01T00:00:00Z",
			"modified_at":   "2026-01-01T00:00:00Z",
			"description":   "fixture record",
			"layout_type":   "ordered",
			"url":           "/dashboard/fixture",
			"author_handle": "fixture@example.com",
			"is_read_only":  false,
			"scope":         []any{"*"},
			"monitor_id":    int64(i),
			"active":        false,
			"disabled":      false,
			"start":         int64(0),
			"end":           int64(0),
			"priority":      int64(i),
			"attributes": map[string]any{
				"name":       fmt.Sprintf("Fixture User %d", i),
				"email":      fmt.Sprintf("fixture+%d@example.com", i),
				"handle":     fmt.Sprintf("fixture+%d@example.com", i),
				"status":     "Active",
				"disabled":   false,
				"verified":   true,
				"created_at": "2026-01-01T00:00:00Z",
			},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the two Datadog auth headers
// and the resolved base URL. The secrets only ever flow into DefaultHeaders;
// they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := datadogBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey := datadogAPIKey(cfg)
	appKey := datadogAppKey(cfg)
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("datadog connector requires secret api_key")
	}
	if strings.TrimSpace(appKey) == "" {
		return nil, errors.New("datadog connector requires secret application_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: datadogUserAgent,
		DefaultHeaders: map[string]string{
			apiKeyHeader: strings.TrimSpace(apiKey),
			appKeyHeader: strings.TrimSpace(appKey),
		},
	}, nil
}

func datadogAPIKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func datadogAppKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["application_key"]
}

// datadogBaseURL resolves and validates the base URL. The default is
// api.datadoghq.com; an explicit base_url override wins. Otherwise a "site"
// config value (e.g. datadoghq.eu, us3.datadoghq.com) selects the regional host.
// Any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func datadogBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if site := strings.TrimSpace(cfg.Config["site"]); site != "" {
			base = "https://api." + site
		} else {
			return datadogDefaultBaseURL, nil
		}
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("datadog config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("datadog config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("datadog config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func datadogPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return datadogDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("datadog config page_size must be an integer: %w", err)
	}
	if value < 1 || value > datadogMaxPageSize {
		return 0, fmt.Errorf("datadog config page_size must be between 1 and %d", datadogMaxPageSize)
	}
	return value, nil
}

func datadogMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("datadog config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("datadog config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Datadog is a read-only
// source connector; reverse-ETL writes to monitoring config are intentionally
// unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
