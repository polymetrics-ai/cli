// Package k6cloud implements the native pm k6 Cloud connector. It is a
// declarative-HTTP per-system connector that composes the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction) with k6 Cloud-specific
// stream definitions and endpoints, mirroring the stripe reference connector.
//
// The package directory and registry key are "k6-cloud" (with a hyphen); the Go
// package identifier is "k6cloud". It self-registers with the connectors
// registry via RegisterFactory in init(); the registryset package blank-imports
// this package in the production binary to run that side effect.
//
// k6 Cloud exposes three full-refresh streams: organizations, projects (a
// substream read per organization), and k6-tests (page-incremented). There are
// no obvious safe reverse-ETL writes for load-test resources, so the connector
// is read-only (Capabilities.Write=false).
package k6cloud

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
	k6DefaultBaseURL  = "https://api.k6.io"
	k6DefaultPageSize = 32
	k6MaxPageSize     = 100
	k6UserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("k6-cloud", New)
}

// New returns the k6 Cloud connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm k6 Cloud connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "k6-cloud" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "k6-cloud",
		DisplayName:     "k6 Cloud",
		IntegrationType: "api",
		Description:     "Reads k6 Cloud organizations, projects, and load tests through the k6 Cloud REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to k6 Cloud. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := k6BaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(k6Secret(cfg)) == "" {
		return errors.New("k6-cloud connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the organizations list confirms auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "/v3/organizations", nil, nil, nil); err != nil {
		return fmt.Errorf("check k6-cloud: %w", err)
	}
	return nil
}

// Write is unsupported: k6 Cloud load-test resources have no obvious safe
// reverse-ETL writes, so the connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: k6Streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "organizations"
	}
	spec, ok := k6StreamSpecs[stream]
	if !ok {
		return fmt.Errorf("k6-cloud stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, spec, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := k6PageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := k6MaxPages(req.Config)
	if err != nil {
		return err
	}

	if spec.perOrganization {
		return c.readPerOrganization(ctx, r, spec, pageSize, maxPages, emit)
	}
	return c.harvest(ctx, r, spec.resource, spec, pageSize, maxPages, emit)
}

// readPerOrganization reads the projects substream: it first lists organization
// ids, then harvests projects for each organization.
func (c Connector) readPerOrganization(ctx context.Context, r *connsdk.Requester, spec streamSpec, pageSize, maxPages int, emit func(connectors.Record) error) error {
	orgSpec := k6StreamSpecs["organizations"]
	orgIDs, err := c.collectOrganizationIDs(ctx, r, orgSpec)
	if err != nil {
		return err
	}
	for _, id := range orgIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := fmt.Sprintf(spec.resource, id)
		if err := c.harvest(ctx, r, path, spec, pageSize, maxPages, emit); err != nil {
			return err
		}
	}
	return nil
}

// collectOrganizationIDs reads the organizations list and returns each id.
func (c Connector) collectOrganizationIDs(ctx context.Context, r *connsdk.Requester, orgSpec streamSpec) ([]int64, error) {
	resp, err := r.Do(ctx, http.MethodGet, orgSpec.resource, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("read k6-cloud organizations: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, orgSpec.recordsPath)
	if err != nil {
		return nil, fmt.Errorf("decode k6-cloud organizations: %w", err)
	}
	ids := make([]int64, 0, len(records))
	for _, item := range records {
		if id, ok := intField(item, "id"); ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// harvest drives k6 Cloud's page-increment pagination. The API returns
// {<recordsPath>:[...]}; a page that is full (== pageSize records) implies there
// may be a next page, requested with page=N+1. A short page ends the loop.
// Non-paginated streams emit the single response and return.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, spec streamSpec, pageSize, maxPages int, emit func(connectors.Record) error) error {
	page := 1
	for pageNum := 0; maxPages == 0 || pageNum < maxPages; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if spec.paginated {
			query.Set("page", strconv.Itoa(page))
			query.Set("page_size", strconv.Itoa(pageSize))
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read k6-cloud %s: %w", spec.recordsPath, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode k6-cloud %s page: %w", spec.recordsPath, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(spec.mapRecord(item)); err != nil {
				return err
			}
		}
		if !spec.paginated || len(records) < pageSize {
			return nil
		}
		page++
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise k6-cloud credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               int64(i),
			"name":             fmt.Sprintf("%s-fixture-%d", stream, i),
			"description":      "fixture record",
			"owner_id":         int64(1),
			"organization_id":  int64(1),
			"project_id":       int64(1),
			"user_id":          int64(1),
			"last_test_run_id": fmt.Sprintf("run_%d", i),
			"test_run_ids":     []any{fmt.Sprintf("run_%d", i)},
			"script":           "export default function () {}",
			"billing_email":    fmt.Sprintf("fixture+%d@example.com", i),
			"billing_country":  "US",
			"is_default":       i == 1,
			"is_saml_org":      false,
			"created":          "2026-01-01T00:00:00Z",
			"updated":          "2026-01-02T00:00:00Z",
		}
		if err := emit(spec.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := k6BaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := k6Secret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("k6-cloud connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: k6UserAgent,
	}, nil
}

func k6Secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// k6BaseURL resolves and validates the base URL. The default is api.k6.io; any
// override must be an absolute https (or http for local test servers) URL with a
// host to bound SSRF risk.
func k6BaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return k6DefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("k6-cloud config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("k6-cloud config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("k6-cloud config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func k6PageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return k6DefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("k6-cloud config page_size must be an integer: %w", err)
	}
	if value < 1 || value > k6MaxPageSize {
		return 0, fmt.Errorf("k6-cloud config page_size must be between 1 and %d", k6MaxPageSize)
	}
	return value, nil
}

func k6MaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("k6-cloud config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("k6-cloud config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// intField extracts an integer id from a decoded JSON object. k6 ids decode as
// json.Number (the connsdk decoder uses UseNumber); plain ints/floats are also
// handled for robustness.
func intField(item map[string]any, key string) (int64, bool) {
	switch v := item[key].(type) {
	case nil:
		return 0, false
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		// json.Number implements fmt.Stringer; parse its decimal form.
		s := fmt.Sprintf("%v", v)
		n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	}
}
