// Package datascope implements the native pm DataScope connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package composing the connsdk toolkit (Requester +
// APIKeyHeader auth + root-array extraction + offset pagination) with
// DataScope-specific stream definitions and endpoints.
//
// DataScope (app.mydatascope.com) is a mobile forms / field-data-collection
// platform. The REST API at https://www.mydatascope.com/api/external is
// read-only for this connector (no reverse-ETL writes), exposing locations,
// answers (form submissions), lists, and notifications.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package datascope

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	datascopeDefaultBaseURL  = "https://www.mydatascope.com/api/external"
	datascopeDefaultPageSize = 200
	datascopeMaxPageSize     = 200
	datascopeUserAgent       = "polymetrics-go-cli"
	// datascopeDatetimeLayout is the Go reference layout for DataScope's
	// "%d/%m/%Y %H:%M" datetime format used by start_date and the start/end
	// request window.
	datascopeDatetimeLayout = "02/01/2006 15:04"
)

func init() {
	connectors.RegisterFactory("datascope", New)
}

// New returns the DataScope connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm DataScope connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "datascope" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "datascope",
		DisplayName:     "DataScope",
		IntegrationType: "api",
		Description:     "Reads DataScope locations, form answers, lists, and notifications from the DataScope external REST API (read-only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to DataScope. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := datascopeBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(datascopeSecret(cfg)) == "" {
		return errors.New("datascope connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the locations list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"limit": []string{"1"}, "offset": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, "locations", query, nil, nil); err != nil {
		return fmt.Errorf("check datascope: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: datascopeStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a DataScope stream starts
// with an empty incremental cursor (full sync), which the start_date config can
// raise at read time.
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
		stream = "locations"
	}
	endpoint, ok := datascopeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("datascope stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := datascopePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := datascopeMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	if endpoint.windowed {
		start, errWindow := windowStart(req)
		if errWindow != nil {
			return errWindow
		}
		if start != "" {
			base.Set("start", start)
			base.Set("end", time.Now().UTC().Format(datascopeDatetimeLayout))
		}
	}

	paginator := &connsdk.OffsetPaginator{
		LimitParam:  "limit",
		OffsetParam: "offset",
		PageSize:    pageSize,
	}
	// DataScope returns a top-level JSON array (record selector field_path: []),
	// so the records path is the root. connsdk.Record is an alias for
	// map[string]any, which is exactly what the mappers consume.
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, "", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// Write is unsupported: DataScope is a read-only source connector (no
// reverse-ETL writes), so this satisfies the connectors.Connector interface by
// returning ErrUnsupportedOperation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise datascope credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		created := time.Date(2026, 1, 1, 0, int64ToMin(i), 0, 0, time.UTC).Format(datascopeDatetimeLayout)
		item := map[string]any{
			"id":              i,
			"form_answer_id":  i,
			"form_id":         100 + i,
			"name":            fmt.Sprintf("%s fixture %d", stream, i),
			"form_name":       fmt.Sprintf("Form %d", i),
			"form_state":      "completed",
			"code":            fmt.Sprintf("CODE-%d", i),
			"user_name":       fmt.Sprintf("Fixture User %d", i),
			"user_identifier": fmt.Sprintf("user+%d@example.com", i),
			"created_at":      created,
			"updated_at":      created,
			"latitude":        38.72 + float64(i)/1000,
			"longitude":       -9.13 - float64(i)/1000,
			"city":            "Lisbon",
			"country":         "PT",
			"region":          "Lisboa",
			"address":         fmt.Sprintf("%d Rua Fixture", i),
			"phone":           "+351000000000",
			"company_code":    "ACME",
			"company_name":    "Acme Field Services",
			"description":     fmt.Sprintf("Fixture %s %d", stream, i),
			"attribute1":      "a1",
			"attribute2":      "a2",
			"list_id":         200 + i,
			"account_id":      300 + i,
			"type":            "form_completed",
			"url":             fmt.Sprintf("https://www.mydatascope.com/fixtures/%d", i),
			"form_code":       fmt.Sprintf("F-%d", i),
			"user":            fmt.Sprintf("Fixture User %d", i),
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

// requester builds a connsdk.Requester wired with the raw api_key Authorization
// header and the resolved base URL. The secret only ever flows into
// connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := datascopeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := datascopeSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("datascope connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, ""),
		UserAgent: datascopeUserAgent,
	}, nil
}

// windowStart returns the start datetime (in DataScope's "%d/%m/%Y %H:%M"
// format) for the start/end request window, derived from the incremental cursor
// (if any) or else the start_date config. An empty result means no window.
func windowStart(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return normalizeDatascopeTime(cursor)
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	return normalizeDatascopeTime(startDate)
}

// normalizeDatascopeTime accepts either DataScope's native "%d/%m/%Y %H:%M"
// format or RFC3339 and returns the value formatted for DataScope's request
// window.
func normalizeDatascopeTime(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if t, err := time.Parse(datascopeDatetimeLayout, value); err == nil {
		return t.Format(datascopeDatetimeLayout), nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC().Format(datascopeDatetimeLayout), nil
	}
	return "", fmt.Errorf("datascope start_date must be %q or RFC3339, got %q", "dd/mm/YYYY HH:MM", value)
}

func datascopeSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// datascopeBaseURL resolves and validates the base URL. The default is
// www.mydatascope.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func datascopeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return datascopeDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("datascope config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("datascope config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("datascope config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func datascopePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return datascopeDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("datascope config page_size must be an integer: %w", err)
	}
	if value < 1 || value > datascopeMaxPageSize {
		return 0, fmt.Errorf("datascope config page_size must be between 1 and %d", datascopeMaxPageSize)
	}
	return value, nil
}

func datascopeMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("datascope config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("datascope config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// int64ToMin clamps a small loop index into a valid minute value for fixture
// timestamps.
func int64ToMin(i int) int {
	if i < 0 {
		return 0
	}
	return i % 60
}
