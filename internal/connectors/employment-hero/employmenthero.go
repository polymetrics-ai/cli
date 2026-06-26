// Package employmenthero implements the native pm Employment Hero connector. It
// is a declarative-HTTP per-system connector built on the stripe template: a
// thin package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction at data.items + page-index pagination) with Employment
// Hero specific stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect. The registry key is "employment-hero" (matching the
// catalog slug source-employment-hero); the Go package is "employmenthero"
// because identifiers cannot contain hyphens.
//
// The Employment Hero API is full-refresh only and read-only: there are no safe
// reverse-ETL writes to expose, so Capabilities.Write is false.
package employmenthero

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
	// registryName is the connector's registry key and catalog slug target.
	registryName = "employment-hero"

	employmentHeroDefaultBaseURL  = "https://api.employmenthero.com/api/v1"
	employmentHeroDefaultPageSize = 100
	employmentHeroMaxPageSize     = 100
	employmentHeroUserAgent       = "polymetrics-go-cli"
	// recordsPath is the dotted JSON path to the records array in list responses.
	recordsPath = "data.items"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Employment Hero connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Employment Hero connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Employment Hero",
		IntegrationType: "api",
		Description:     "Reads Employment Hero organisations, employees, leave requests, and teams through the Employment Hero REST API. Read-only (the API is full-refresh only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Employment
// Hero. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := employmentHeroBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(employmentHeroSecret(cfg)) == "" {
		return errors.New("employment-hero connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the organisations list confirms auth and connectivity
	// without mutating anything.
	q := url.Values{"items_per_page": []string{"1"}, "page_index": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "organisations", q, nil, nil); err != nil {
		return fmt.Errorf("check employment-hero: %w", err)
	}
	return nil
}

// Write is not supported: the Employment Hero API exposes no safe reverse-ETL
// writes, so the connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: employmentHeroStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "organisations"
	}
	endpoint, ok := employmentHeroStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("employment-hero stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	orgID := ""
	if endpoint.orgScoped {
		var err error
		orgID, err = organizationID(req.Config)
		if err != nil {
			return err
		}
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := employmentHeroPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := employmentHeroMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint.path(orgID), endpoint.mapRecord, pageSize, maxPages, emit)
}

// harvest drives Employment Hero's page-index pagination. List responses wrap the
// records in {data:{items:[...], total_pages:N}}; pages are requested with
// page_index=1..total_pages and items_per_page=pageSize. The loop stops at
// total_pages, on a short/empty page, or at maxPages. It is built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, mapRecord func(map[string]any) connectors.Record, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page_index", strconv.Itoa(page))
		query.Set("items_per_page", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read employment-hero %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return fmt.Errorf("decode employment-hero %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		// Stop on a short page (fewer than requested) to handle endpoints that
		// omit total_pages, then defer to total_pages when present.
		if len(records) < pageSize {
			return nil
		}
		totalPages := intAt(resp.Body, "data.total_pages")
		if totalPages > 0 && page >= totalPages {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                     fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":                   fmt.Sprintf("Fixture %s %d", stream, i),
			"country":                "AU",
			"phone":                  "+61000000000",
			"logo_url":               "https://example.com/logo.png",
			"first_name":             fmt.Sprintf("First%d", i),
			"middle_name":            "",
			"last_name":              fmt.Sprintf("Last%d", i),
			"known_as":               fmt.Sprintf("First%d", i),
			"title":                  "Mr",
			"job_title":              "Engineer",
			"role":                   "employee",
			"account_email":          fmt.Sprintf("fixture+%d@example.com", i),
			"company_email":          fmt.Sprintf("fixture+%d@example.com", i),
			"personal_email":         fmt.Sprintf("personal+%d@example.com", i),
			"company_mobile":         "+61000000001",
			"personal_mobile_number": "+61000000002",
			"gender":                 "unspecified",
			"date_of_birth":          "1990-01-01",
			"location":               "Sydney",
			"start_date":             "2026-01-01",
			"employing_entity":       "Fixture Pty Ltd",
			"primary_manager":        "mgr_fixture_1",
			"employee_id":            "employees_fixture_1",
			"leave_category_name":    "Annual Leave",
			"status":                 "approved",
			"end_date":               "2026-01-05",
			"total_hours":            float64(8 * i),
			"leave_balance_amount":   float64(40 * i),
			"comment":                "fixture leave request",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := employmentHeroBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := employmentHeroSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("employment-hero connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: employmentHeroUserAgent,
	}, nil
}

// organizationID resolves the organisation id for an org-scoped stream from the
// config. It accepts organization_id (single) or the first entry of the
// comma-separated organization_configids list (mirrors the upstream catalog).
func organizationID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config != nil {
		if id := strings.TrimSpace(cfg.Config["organization_id"]); id != "" {
			return id, nil
		}
		if list := strings.TrimSpace(cfg.Config["organization_configids"]); list != "" {
			for _, part := range strings.Split(list, ",") {
				if id := strings.TrimSpace(part); id != "" {
					return id, nil
				}
			}
		}
	}
	return "", errors.New("employment-hero org-scoped stream requires config organization_id (or organization_configids); read the organisations stream first to discover ids")
}

func employmentHeroSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// employmentHeroBaseURL resolves and validates the base URL. The default is
// api.employmenthero.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func employmentHeroBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return employmentHeroDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("employment-hero config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("employment-hero config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("employment-hero config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func employmentHeroPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["items_per_page"])
	if raw == "" {
		return employmentHeroDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("employment-hero config items_per_page must be an integer: %w", err)
	}
	if value < 1 || value > employmentHeroMaxPageSize {
		return 0, fmt.Errorf("employment-hero config items_per_page must be between 1 and %d", employmentHeroMaxPageSize)
	}
	return value, nil
}

func employmentHeroMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("employment-hero config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("employment-hero config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// intAt reads an integer at a dotted JSON path, returning 0 when absent or
// non-numeric.
func intAt(body []byte, path string) int {
	s, err := connsdk.StringAt(body, path)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}
