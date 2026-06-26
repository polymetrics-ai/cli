// Package grafana implements the native pm Grafana connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction) with Grafana-specific stream definitions and endpoints.
//
// Grafana exposes a REST API on the user's instance (e.g.
// https://your-grafana.grafana.net) under /api/. Authentication uses a service
// account token or API key sent as Authorization: Bearer <api_key>. List
// endpoints return top-level JSON arrays; /api/search and /api/folders support
// page/limit pagination while the remaining list endpoints return everything in
// one response.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. The connector is read-only:
// Grafana writes (creating dashboards/datasources) are not safe reverse-ETL
// targets, so Capabilities.Write is false.
package grafana

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
	grafanaDefaultPageSize = 1000
	grafanaMaxPageSize     = 5000
	grafanaUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("grafana", New)
}

// New returns the Grafana connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Grafana connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "grafana" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "grafana",
		DisplayName:     "Grafana",
		IntegrationType: "api",
		Description:     "Reads Grafana dashboards, folders, data sources, organization users, and provisioned alert rules through the Grafana REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Grafana. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := grafanaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(grafanaSecret(cfg)) == "" {
		return errors.New("grafana connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the org confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "api/org", nil, nil, nil); err != nil {
		return fmt.Errorf("check grafana: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: grafanaStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "dashboards"
	}
	endpoint, ok := grafanaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("grafana stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := grafanaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := grafanaMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Grafana's read. Endpoints all return top-level JSON arrays.
// Paginated endpoints (/api/search) advance page=1,2,... with a fixed limit and
// stop when a short page (fewer than limit records) is returned. Non-paginated
// endpoints are fetched in a single request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	if !endpoint.paginated {
		_, err := c.readPage(ctx, r, endpoint, nil, emit)
		return err
	}

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("page", strconv.Itoa(page))
		count, err := c.readPage(ctx, r, endpoint, query, emit)
		if err != nil {
			return err
		}
		// A short (or empty) page means there are no more results.
		if count < pageSize {
			return nil
		}
	}
	return nil
}

// readPage performs one request and emits its records, returning the record
// count so the pagination loop can detect a short final page.
func (c Connector) readPage(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, query url.Values, emit func(connectors.Record) error) (int, error) {
	q := url.Values{}
	for k, v := range endpoint.fixedQuery {
		q.Set(k, v)
	}
	for k, vs := range query {
		for _, v := range vs {
			q.Set(k, v)
		}
	}
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, q, nil)
	if err != nil {
		return 0, fmt.Errorf("read grafana %s: %w", endpoint.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return 0, fmt.Errorf("decode grafana %s: %w", endpoint.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return 0, err
		}
	}
	return len(records), nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise grafana credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           int64(i),
			"uid":          fmt.Sprintf("%s_fixture_%d", stream, i),
			"orgId":        int64(1),
			"orgID":        int64(1),
			"userId":       int64(i),
			"title":        fmt.Sprintf("Fixture %s %d", stream, i),
			"name":         fmt.Sprintf("Fixture %s %d", stream, i),
			"url":          fmt.Sprintf("/d/fixture_%d", i),
			"type":         "dash-db",
			"tags":         []any{"fixture"},
			"isStarred":    false,
			"folderId":     int64(0),
			"folderUid":    "general",
			"folderTitle":  "General",
			"folderUID":    "general",
			"access":       "proxy",
			"isDefault":    i == 1,
			"readOnly":     false,
			"email":        fmt.Sprintf("fixture+%d@example.com", i),
			"login":        fmt.Sprintf("fixture_%d", i),
			"role":         "Viewer",
			"lastSeenAt":   "2026-01-01T00:00:00Z",
			"ruleGroup":    "fixture-group",
			"condition":    "A",
			"noDataState":  "NoData",
			"execErrState": "Error",
			"for":          "5m",
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
	base, err := grafanaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := grafanaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("grafana connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: grafanaUserAgent,
	}, nil
}

func grafanaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// grafanaBaseURL resolves and validates the base URL of the Grafana instance.
// The instance URL has no fixed default (it is per-tenant), so it is taken from
// the base_url override or the url config field. Any value must be an absolute
// https (or http for local test servers) URL with a host to bound SSRF risk.
func grafanaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = strings.TrimSpace(cfg.Config["url"])
	}
	if base == "" {
		return "", errors.New("grafana connector requires config url (Grafana instance URL) or base_url")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("grafana config url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("grafana config url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("grafana config url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func grafanaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return grafanaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("grafana config page_size must be an integer: %w", err)
	}
	if value < 1 || value > grafanaMaxPageSize {
		return 0, fmt.Errorf("grafana config page_size must be between 1 and %d", grafanaMaxPageSize)
	}
	return value, nil
}

func grafanaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("grafana config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("grafana config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Grafana is a read-only
// source in pm, so writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
