// Package jobnimbus implements the native pm JobNimbus connector. It is a
// declarative-HTTP per-system connector following the stripe reference shape: a
// thin package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + offset pagination) with JobNimbus-specific stream
// definitions and endpoints.
//
// JobNimbus exposes a read-only CRM REST API at https://app.jobnimbus.com/api1.
// List endpoints return {results:[...]} (contacts/jobs/tasks), {activity:[...]}
// (activities), or {files:[...]} (files), and are paged with an offset `from`
// query parameter plus a `size` page size. All objects carry a string `jnid`
// primary key.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package jobnimbus

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
	jobnimbusDefaultBaseURL  = "https://app.jobnimbus.com/api1"
	jobnimbusDefaultPageSize = 1000
	jobnimbusMaxPageSize     = 1000
	jobnimbusUserAgent       = "polymetrics-go-cli"
	// jobnimbusFixtureCreated is the deterministic epoch-seconds timestamp used
	// by fixture-mode records (2026-01-01T00:00:00Z).
	jobnimbusFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("jobnimbus", New)
}

// New returns the JobNimbus connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm JobNimbus connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "jobnimbus" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "jobnimbus",
		DisplayName:     "JobNimbus",
		IntegrationType: "api",
		Description:     "Reads JobNimbus CRM contacts, jobs, tasks, activities, and files through the JobNimbus REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to JobNimbus.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := jobnimbusBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(jobnimbusSecret(cfg)) == "" {
		return errors.New("jobnimbus connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the contacts list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"size": []string{"1"}, "from": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, "contacts", query, nil, nil); err != nil {
		return fmt.Errorf("check jobnimbus: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: jobnimbusStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: JobNimbus supports only
// full-refresh upstream, so a stream starts with an empty cursor.
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
		stream = "contacts"
	}
	endpoint, ok := jobnimbusStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("jobnimbus stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := jobnimbusPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := jobnimbusMaxPages(req.Config)
	if err != nil {
		return err
	}

	// JobNimbus pages with an offset `from` parameter and `size` page size,
	// stopping when a short page is returned. connsdk.OffsetPaginator models
	// exactly this; the per-stream recordsPath handles the inconsistent
	// envelope key (results / activity / files).
	paginator := &connsdk.OffsetPaginator{
		LimitParam:  "size",
		OffsetParam: "from",
		PageSize:    pageSize,
	}
	mapped := func(raw connsdk.Record) error {
		return emit(endpoint.mapRecord(raw))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, paginator, endpoint.recordsPath, maxPages, mapped); err != nil {
		return fmt.Errorf("read jobnimbus %s: %w", endpoint.resource, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise jobnimbus credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"jnid":               fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"type":               strings.TrimSuffix(stream, "s"),
			"display_name":       fmt.Sprintf("Fixture %d", i),
			"first_name":         "Fixture",
			"last_name":          strconv.Itoa(i),
			"company":            "Example Co",
			"email":              fmt.Sprintf("fixture+%d@example.com", i),
			"name":               fmt.Sprintf("Fixture Job %d", i),
			"number":             strconv.Itoa(1000 + i),
			"title":              fmt.Sprintf("Fixture Task %d", i),
			"note":               "fixture activity note",
			"filename":           fmt.Sprintf("fixture_%d.pdf", i),
			"content_type":       "application/pdf",
			"status_name":        "Active",
			"record_type_name":   stream,
			"customer":           "cust_fixture_1",
			"created_by_name":    "Fixture User",
			"sales_rep_name":     "Fixture Rep",
			"source":             "fixture",
			"priority":           int64(i),
			"size":               int64(1024 * i),
			"md5":                "00000000000000000000000000000000",
			"is_active":          true,
			"is_archived":        false,
			"is_completed":       false,
			"is_status_change":   false,
			"date_created":       jobnimbusFixtureCreated + int64(i),
			"date_updated":       jobnimbusFixtureCreated + int64(i),
			"date_status_change": jobnimbusFixtureCreated + int64(i),
			"date_file_created":  jobnimbusFixtureCreated + int64(i),
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

// requester builds a connsdk.Requester wired with Bearer auth (the JobNimbus
// api_key) and the resolved base URL. The secret only ever flows into
// connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := jobnimbusBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := jobnimbusSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("jobnimbus connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: jobnimbusUserAgent,
	}, nil
}

func jobnimbusSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// jobnimbusBaseURL resolves and validates the base URL. The default is
// app.jobnimbus.com/api1; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func jobnimbusBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return jobnimbusDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("jobnimbus config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("jobnimbus config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("jobnimbus config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func jobnimbusPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return jobnimbusDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("jobnimbus config page_size must be an integer: %w", err)
	}
	if value < 1 || value > jobnimbusMaxPageSize {
		return 0, fmt.Errorf("jobnimbus config page_size must be between 1 and %d", jobnimbusMaxPageSize)
	}
	return value, nil
}

func jobnimbusMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("jobnimbus config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("jobnimbus config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: JobNimbus is exposed read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
