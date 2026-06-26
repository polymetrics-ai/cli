// Package firehydrant implements the native pm FireHydrant connector. It follows
// the declarative-HTTP per-system connector shape pioneered by the stripe
// package: a thin package that composes the connsdk toolkit (Requester + Bearer
// auth + RecordsAt extraction) with FireHydrant-specific stream definitions,
// endpoints, and pagination.
//
// FireHydrant is an incident-management/reliability platform; this connector is
// read-only (full-refresh) over a core set of catalog and incident streams. It
// self-registers with the connectors registry via RegisterFactory in init(); the
// registryset package blank-imports this package in the production binary to run
// that side effect.
package firehydrant

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
	firehydrantDefaultBaseURL  = "https://api.firehydrant.io/v1"
	firehydrantDefaultPageSize = 50
	firehydrantMaxPageSize     = 200
	firehydrantUserAgent       = "polymetrics-go-cli"
	// firehydrantFixtureCreated is the deterministic created_at timestamp used by
	// the fixture-mode records.
	firehydrantFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("firehydrant", New)
}

// New returns the FireHydrant connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm FireHydrant connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "firehydrant" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "firehydrant",
		DisplayName:     "FireHydrant",
		IntegrationType: "api",
		Description:     "Reads FireHydrant incidents, services, teams, environments, and functionalities through the FireHydrant REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to FireHydrant.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := firehydrantBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(firehydrantSecret(cfg)) == "" {
		return errors.New("firehydrant connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the environments list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "environments", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check firehydrant: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: firehydrantStreams()}, nil
}

// Write is unsupported: FireHydrant is exposed as a read-only source connector.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "incidents"
	}
	endpoint, ok := firehydrantStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("firehydrant stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := firehydrantPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := firehydrantMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives FireHydrant's page-number pagination. List responses are shaped
// {data:[...], pagination:{page, next, prev, ...}}; pagination.next holds the next
// page number (or null when exhausted). There is no body-token paginator in
// connsdk for this exact shape, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))

	page := ""
	for pageNum := 0; maxPages == 0 || pageNum < maxPages; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if page != "" {
			query.Set("page", page)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read firehydrant %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode firehydrant %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "pagination.next")
		if err != nil {
			return fmt.Errorf("decode firehydrant %s pagination: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		// A null/empty/zero "next" means there are no further pages.
		if next == "" || next == "0" {
			return nil
		}
		page = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise firehydrant credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":              fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"number":            int64(i),
			"description":       "fixture record",
			"summary":           "fixture summary",
			"slug":              fmt.Sprintf("fixture-%d", i),
			"current_milestone": "started",
			"severity":          "SEV1",
			"priority":          "P1",
			"service_tier":      int64(1),
			"created_at":        firehydrantFixtureCreated,
			"updated_at":        firehydrantFixtureCreated,
			"started_at":        firehydrantFixtureCreated,
			"resolved_at":       nil,
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
	base, err := firehydrantBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := firehydrantSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("firehydrant connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: firehydrantUserAgent,
	}, nil
}

func firehydrantSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

// firehydrantBaseURL resolves and validates the base URL. The default is
// api.firehydrant.io; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func firehydrantBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return firehydrantDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("firehydrant config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("firehydrant config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("firehydrant config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func firehydrantPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return firehydrantDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("firehydrant config page_size must be an integer: %w", err)
	}
	if value < 1 || value > firehydrantMaxPageSize {
		return 0, fmt.Errorf("firehydrant config page_size must be between 1 and %d", firehydrantMaxPageSize)
	}
	return value, nil
}

func firehydrantMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("firehydrant config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("firehydrant config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}
