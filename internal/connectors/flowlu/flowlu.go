// Package flowlu implements the native pm Flowlu connector. It is a declarative
// HTTP per-system connector built on the connsdk toolkit: a thin package that
// composes connsdk.Requester + api_key query authentication + response.items
// extraction with Flowlu-specific stream definitions and endpoints.
//
// It mirrors the stripe reference connector's shape. Like the other per-system
// connectors it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the
// production binary to run that side effect.
//
// Flowlu API shape (https://www.flowlu.com/api/):
//   - base URL: https://{company}.flowlu.com/api/v1/module
//   - auth: api_key query parameter on every request
//   - list endpoints return {"response":{"total_result":N,"page":P,"count":C,"items":[...]}}
//   - pagination: page=1.. with a count page size, stopping on an empty items page
//
// The connector is read-only: Flowlu's write surface is not exposed for reverse
// ETL here, so Capabilities.Write is false and Write returns ErrUnsupportedOperation.
package flowlu

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
	flowluDefaultPageSize = 100
	flowluMaxPageSize     = 100
	flowluUserAgent       = "polymetrics-go-cli"
	// flowluRecordsPath is the dotted path to the records array in list responses.
	flowluRecordsPath = "response.items"
	// flowluFixtureDate is the deterministic timestamp used by fixture-mode records.
	flowluFixtureDate = "2026-01-01 00:00:00"
)

func init() {
	connectors.RegisterFactory("flowlu", New)
}

// New returns the Flowlu connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Flowlu connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "flowlu" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "flowlu",
		DisplayName:     "Flowlu",
		IntegrationType: "api",
		Description:     "Reads Flowlu CRM accounts, leads, tasks, projects, invoices, and agile issues through the Flowlu REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Flowlu. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := flowluBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(flowluSecret(cfg)) == "" {
		return errors.New("flowlu connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the accounts list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"page": []string{"1"}, "count": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "crm/account/list", query, nil, nil); err != nil {
		return fmt.Errorf("check flowlu: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: flowluStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Flowlu stream starts with
// an empty cursor (full refresh, the only sync mode Flowlu's list API supports
// reliably).
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
		stream = "accounts"
	}
	endpoint, ok := flowluStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("flowlu stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := flowluPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := flowluMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Flowlu's page-number pagination. List endpoints return
// {response:{items:[...]}}; pages are requested with page=1.. and a count page
// size, and the loop stops when a page returns no items. The loop lives here
// (rather than connsdk.Harvest) so it can read records at the nested
// response.items path and stop on an empty page.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("count", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read flowlu %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, flowluRecordsPath)
		if err != nil {
			return fmt.Errorf("decode flowlu %s page: %w", endpoint.resource, err)
		}
		if len(records) == 0 {
			return nil
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than the requested size) is the last page.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise flowlu credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                int64(i),
			"name":              fmt.Sprintf("%s fixture %d", stream, i),
			"title":             fmt.Sprintf("Fixture %d", i),
			"description":       "fixture record",
			"first_name":        fmt.Sprintf("First%d", i),
			"last_name":         fmt.Sprintf("Last%d", i),
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"phone":             "+10000000000",
			"type":              int64(1),
			"active":            int64(1),
			"owner_id":          int64(1),
			"stage_id":          int64(1),
			"pipeline_id":       int64(1),
			"budget":            "1000.00",
			"priority":          int64(2),
			"workflow_stage_id": int64(1),
			"responsible_id":    int64(1),
			"manager_id":        int64(1),
			"deadline":          flowluFixtureDate,
			"invoice_number":    fmt.Sprintf("INV-%d", i),
			"customer_id":       int64(1),
			"total_amount":      "100.00",
			"currency_id":       int64(1),
			"invoice_status":    int64(1),
			"invoice_date":      flowluFixtureDate,
			"project_id":        int64(1),
			"sprint_id":         int64(1),
			"created_date":      flowluFixtureDate,
			"updated_date":      flowluFixtureDate,
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

// Write is unsupported: the Flowlu connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with api_key query auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyQuery; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := flowluBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := flowluSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("flowlu connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyQuery("api_key", secret),
		UserAgent: flowluUserAgent,
	}, nil
}

func flowluSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// flowluBaseURL resolves and validates the base URL. When base_url is not
// overridden it is derived from the required company subdomain as
// https://{company}.flowlu.com/api/v1/module. Any override must be an absolute
// https (or http for local test servers) URL with a host to bound SSRF risk.
func flowluBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		company := strings.TrimSpace(cfg.Config["company"])
		if company == "" {
			return "", errors.New("flowlu connector requires config company (subdomain) or base_url")
		}
		if !validCompany(company) {
			return "", fmt.Errorf("flowlu config company %q must be a bare subdomain", company)
		}
		return fmt.Sprintf("https://%s.flowlu.com/api/v1/module", company), nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("flowlu config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("flowlu config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("flowlu config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// validCompany checks that a company subdomain contains only characters valid in
// a DNS label, so it cannot smuggle a path/host into the derived URL.
func validCompany(company string) bool {
	for _, r := range company {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func flowluPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return flowluDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("flowlu config page_size must be an integer: %w", err)
	}
	if value < 1 || value > flowluMaxPageSize {
		return 0, fmt.Errorf("flowlu config page_size must be between 1 and %d", flowluMaxPageSize)
	}
	return value, nil
}

func flowluMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("flowlu config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("flowlu config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
