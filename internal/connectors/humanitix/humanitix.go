// Package humanitix implements the native pm Humanitix connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, mirroring
// the stripe reference connector: a thin package that composes a Requester +
// APIKeyHeader auth + DpathExtractor-style record extraction + page-increment
// pagination with Humanitix-specific stream definitions and endpoints.
//
// Humanitix is a ticketing platform. The public API
// (https://api.humanitix.com/v1/) authenticates with an x-api-key header and
// paginates list endpoints with page/pageSize query parameters (1-based page
// increment). This connector is read-only: the Humanitix public API exposes no
// safe reverse-ETL writes, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package humanitix

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
	humanitixDefaultBaseURL  = "https://api.humanitix.com/v1"
	humanitixDefaultPageSize = 100
	humanitixMaxPageSize     = 100
	humanitixUserAgent       = "polymetrics-go-cli"
	// humanitixFixtureUpdated is the deterministic updatedAt used by fixture mode.
	humanitixFixtureUpdated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("humanitix", New)
}

// New returns the Humanitix connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Humanitix connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "humanitix" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "humanitix",
		DisplayName:     "Humanitix",
		IntegrationType: "api",
		Description:     "Reads Humanitix events, orders, tickets, and tags through the Humanitix public REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Humanitix.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := humanitixBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(humanitixSecret(cfg)) == "" {
		return errors.New("humanitix connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the events list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "events", url.Values{"page": []string{"1"}, "pageSize": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check humanitix: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: humanitixStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Humanitix stream starts
// with an empty incremental cursor (full sync), which the since config can raise
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
		stream = "events"
	}
	endpoint, ok := humanitixStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("humanitix stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := humanitixPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := humanitixMaxPages(req.Config)
	if err != nil {
		return err
	}

	path := endpoint.resource
	if endpoint.substream {
		eventID := strings.TrimSpace(req.Config.Config["event_id"])
		if eventID == "" {
			return fmt.Errorf("humanitix stream %q requires config event_id", stream)
		}
		path = strings.ReplaceAll(path, "{eventid}", url.PathEscape(eventID))
	}

	base := url.Values{}
	// Humanitix events support a `since` filter (ISO timestamp) for incremental
	// reads; carry it from the cursor or the since config when present.
	if since := incrementalLowerBound(req); since != "" {
		base.Set("since", since)
	}

	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "pageSize",
		StartPage: 1,
		PageSize:  pageSize,
	}

	return connsdk.Harvest(ctx, r, http.MethodGet, path, base, paginator, endpoint.recordsField, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// Write satisfies the connectors.Connector interface. Humanitix is read-only in
// pm (no safe reverse-ETL writes), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise humanitix credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"_id":             fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":            fmt.Sprintf("Fixture %s %d", stream, i),
			"slug":            fmt.Sprintf("fixture-%s-%d", stream, i),
			"currency":        "AUD",
			"location":        "AU",
			"startDate":       humanitixFixtureUpdated,
			"endDate":         humanitixFixtureUpdated,
			"organiserId":     "org_fixture_1",
			"userId":          "user_fixture_1",
			"public":          true,
			"published":       true,
			"markedAsSoldOut": false,
			"eventId":         "events_fixture_1",
			"eventDateId":     "date_fixture_1",
			"orderId":         "orders_fixture_1",
			"orderName":       fmt.Sprintf("ORDER-%d", i),
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"firstName":       "Fixture",
			"lastName":        fmt.Sprintf("User%d", i),
			"mobile":          "",
			"status":          "complete",
			"financialStatus": "paid",
			"ticketTypeId":    "tt_fixture_1",
			"ticketTypeName":  "General Admission",
			"total":           float64(100 * i),
			"price":           float64(50 * i),
			"number":          float64(i),
			"isDonation":      false,
			"manualOrder":     false,
			"completedAt":     humanitixFixtureUpdated,
			"createdAt":       humanitixFixtureUpdated,
			"updatedAt":       humanitixFixtureUpdated,
		}
		record := endpoint.mapRecord(item)
		if cursor := connsdk.Cursor(req.State); cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with x-api-key header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := humanitixBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := humanitixSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("humanitix connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("x-api-key", secret, ""),
		UserAgent: humanitixUserAgent,
	}, nil
}

// incrementalLowerBound returns the `since` lower bound, derived from the
// incremental cursor (if any) or else the since config. An empty result means no
// lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["since"])
}

func humanitixSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// humanitixBaseURL resolves and validates the base URL. The default is
// api.humanitix.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func humanitixBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return humanitixDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("humanitix config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("humanitix config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("humanitix config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func humanitixPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return humanitixDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("humanitix config page_size must be an integer: %w", err)
	}
	if value < 1 || value > humanitixMaxPageSize {
		return 0, fmt.Errorf("humanitix config page_size must be between 1 and %d", humanitixMaxPageSize)
	}
	return value, nil
}

func humanitixMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("humanitix config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("humanitix config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
