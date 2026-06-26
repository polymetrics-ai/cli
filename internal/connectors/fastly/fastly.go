// Package fastly implements the native pm Fastly connector. It follows the
// declarative-HTTP per-system connector template (see internal/connectors/stripe):
// a thin package that composes the connsdk toolkit (Requester + Fastly-Key API
// header auth + RecordsAt extraction) with Fastly-specific stream definitions and
// endpoints.
//
// Fastly's API authenticates with a "Fastly-Key: <token>" header. List endpoints
// such as /service return a top-level JSON array and accept page/per_page
// pagination; singleton endpoints such as /current_user return a single object.
// This connector is read-only — the Fastly API has no obvious safe reverse-ETL
// write surface, so Capabilities.Write is false.
//
// Like the other per-system connectors it self-registers with the connectors
// registry via RegisterFactory in init(); the registryset package blank-imports
// this package in the production binary to run that side effect.
package fastly

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
	fastlyDefaultBaseURL  = "https://api.fastly.com"
	fastlyDefaultPageSize = 100
	fastlyMaxPageSize     = 100
	fastlyAuthHeader      = "Fastly-Key"
	fastlyUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("fastly", New)
}

// New returns the Fastly connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Fastly connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "fastly" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "fastly",
		DisplayName:     "Fastly",
		IntegrationType: "api",
		Description:     "Reads Fastly services, the current user, the current customer (account), and datacenters through the Fastly REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Fastly. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := fastlyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(fastlySecret(cfg)) == "" {
		return errors.New("fastly connector requires secret fastly_api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Reading the current user confirms the token authenticates without mutating
	// or scanning customer data.
	if err := r.DoJSON(ctx, http.MethodGet, "current_user", nil, nil, nil); err != nil {
		return fmt.Errorf("check fastly: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: fastlyStreams()}, nil
}

// Write satisfies the connectors.Connector interface. The Fastly connector is
// read-only (no safe reverse-ETL surface), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "services"
	}
	endpoint, ok := fastlyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("fastly stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if !endpoint.paginated {
		return c.readSingle(ctx, r, endpoint, emit)
	}
	pageSize, err := fastlyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := fastlyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// readSingle reads a singleton endpoint (e.g. /current_user) that returns one
// JSON object. RecordsAt at the root wraps the object into a one-element set.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read fastly %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode fastly %s: %w", endpoint.resource, err)
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

// harvest drives Fastly's page/per_page pagination over a top-level JSON array.
// Fastly list endpoints return a bare array with no envelope or has_more flag, so
// the loop stops when a page returns fewer records than per_page (the standard
// short-page signal, mirrored by connsdk.OffsetPaginator).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read fastly %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode fastly %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short (or empty) page means there are no more pages.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise fastly credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	count := 2
	if !endpoint.paginated {
		count = 1 // singleton streams have exactly one record
	}
	for i := 1; i <= count; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                      fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"code":                    fmt.Sprintf("DC%d", i),
			"name":                    fmt.Sprintf("Fixture %s %d", stream, i),
			"login":                   fmt.Sprintf("fixture+%d@example.com", i),
			"role":                    "engineer",
			"customer_id":             "cus_fixture_1",
			"owner_id":                "usr_fixture_1",
			"type":                    "vcl",
			"version":                 int64(i),
			"group":                   "US",
			"shield":                  "fixture-shield",
			"pricing_plan":            "fixture",
			"paused":                  false,
			"locked":                  false,
			"two_factor_auth_enabled": false,
			"can_stream_syslog":       true,
			"has_account_panel":       true,
			"created_at":              "2026-01-01T00:00:00Z",
			"updated_at":              fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"connector":               "fastly",
			"fixture":                 true,
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Fastly-Key API-header auth and
// the resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it
// is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := fastlyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := fastlySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("fastly connector requires secret fastly_api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(fastlyAuthHeader, secret, ""),
		UserAgent: fastlyUserAgent,
	}, nil
}

func fastlySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["fastly_api_token"]
}

// fastlyBaseURL resolves and validates the base URL. The default is
// api.fastly.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func fastlyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return fastlyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("fastly config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("fastly config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("fastly config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fastlyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return fastlyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fastly config page_size must be an integer: %w", err)
	}
	if value < 1 || value > fastlyMaxPageSize {
		return 0, fmt.Errorf("fastly config page_size must be between 1 and %d", fastlyMaxPageSize)
	}
	return value, nil
}

func fastlyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fastly config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("fastly config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
