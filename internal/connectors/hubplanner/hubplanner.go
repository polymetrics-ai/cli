// Package hubplanner implements the native pm Hubplanner connector. It is a
// declarative-HTTP per-system connector following the stripe reference shape: a
// thin package that composes the connsdk toolkit (Requester + raw API-key auth +
// top-level-array extraction + page/limit pagination) with Hubplanner-specific
// stream definitions and endpoints.
//
// Hubplanner is read-only here (the upstream API only supports full_refresh
// pulls of scheduling data), so no write actions are exposed.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// API reference: https://github.com/hubplanner/API
package hubplanner

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
	hubplannerDefaultBaseURL  = "https://api.hubplanner.com/v1"
	hubplannerDefaultPageSize = 200
	// hubplannerMaxPageSize is the documented maximum limit (1000).
	hubplannerMaxPageSize = 1000
	hubplannerUserAgent   = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("hubplanner", New)
}

// New returns the Hubplanner connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Hubplanner connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "hubplanner" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "hubplanner",
		DisplayName:     "Hubplanner",
		IntegrationType: "api",
		Description:     "Reads Hubplanner resources, projects, clients, events, holidays, bookings, and billing rates through the Hubplanner REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Hubplanner.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := hubplannerBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(hubplannerSecret(cfg)) == "" {
		return errors.New("hubplanner connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the resource list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"limit": []string{"1"}, "page": []string{"0"}}
	if err := r.DoJSON(ctx, http.MethodGet, "resource", q, nil, nil); err != nil {
		return fmt.Errorf("check hubplanner: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: hubplannerStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "resources"
	}
	endpoint, ok := hubplannerStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("hubplanner stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := hubplannerPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := hubplannerMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Hubplanner's page/limit pagination over a top-level JSON array.
// Pages are 0-indexed; the loop stops when a page returns fewer records than the
// requested limit (the documented termination signal) or maxPages is reached.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read hubplanner %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode hubplanner %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than the requested limit) is the last page.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Hubplanner is read-only
// (the upstream API exposes scheduling data we only pull), so writes are
// unsupported.
func (c Connector) Write(_ context.Context, _ connectors.WriteRequest, _ []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise hubplanner credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"_id":              fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"firstName":        fmt.Sprintf("Fixture%d", i),
			"lastName":         "Example",
			"name":             fmt.Sprintf("Fixture %s %d", stream, i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"phone":            "+10000000000",
			"status":           "STATUS_ACTIVE",
			"state":            "BOOKING_APPROVED",
			"role":             "member",
			"type":             "REGULAR",
			"note":             "fixture record",
			"createdDate":      "2026-01-01T00:00:00Z",
			"updatedDate":      "2026-01-02T00:00:00Z",
			"projectCode":      fmt.Sprintf("PRJ-%d", i),
			"budgetHours":      float64(100 * i),
			"budgetCashAmount": float64(1000 * i),
			"budgetCurrency":   "USD",
			"start":            "2026-01-01T00:00:00Z",
			"end":              "2026-01-08T00:00:00Z",
			"date":             "2026-12-25",
			"holidayGroup":     "default",
			"resource":         "resource_fixture_1",
			"project":          "project_fixture_1",
			"category":         "billable",
			"rate":             float64(150 * i),
			"currency":         "USD",
			"default":          i == 1,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with raw API-key auth and the
// resolved base URL. Hubplanner expects the API key directly in the
// Authorization header with no "Bearer" prefix. The secret only ever flows into
// connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := hubplannerBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := hubplannerSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("hubplanner connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, ""),
		UserAgent: hubplannerUserAgent,
	}, nil
}

func hubplannerSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// hubplannerBaseURL resolves and validates the base URL. The default is
// api.hubplanner.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func hubplannerBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return hubplannerDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("hubplanner config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("hubplanner config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("hubplanner config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func hubplannerPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return hubplannerDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("hubplanner config page_size must be an integer: %w", err)
	}
	if value < 1 || value > hubplannerMaxPageSize {
		return 0, fmt.Errorf("hubplanner config page_size must be between 1 and %d", hubplannerMaxPageSize)
	}
	return value, nil
}

func hubplannerMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("hubplanner config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("hubplanner config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
