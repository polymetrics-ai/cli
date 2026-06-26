// Package incidentio implements the native pm incident.io connector. It is a
// declarative-HTTP per-system connector following the stripe reference shape: a
// thin package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + cursor pagination) with incident.io-specific stream
// definitions, endpoints, and record mappers.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// incident.io is read-only here: the public API is primarily a read surface for
// incidents, severities, roles, users, and follow-ups, so the connector exposes
// Read without any reverse-ETL write actions.
package incidentio

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
	defaultBaseURL  = "https://api.incident.io"
	defaultPageSize = 100
	maxPageSize     = 250
	userAgent       = "polymetrics-go-cli"
	connectorName   = "incident-io"
)

func init() {
	connectors.RegisterFactory(connectorName, New)
}

// New returns the incident.io connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm incident.io connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Incident.io",
		IntegrationType: "api",
		Description:     "Reads incident.io incidents, severities, incident roles, users, and follow-ups through the incident.io REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to incident.io.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("incident-io connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of severities (a small, always-present list) confirms auth
	// and connectivity without reading large incident volumes.
	if err := r.DoJSON(ctx, http.MethodGet, "v1/severities", nil, nil, nil); err != nil {
		return fmt.Errorf("check incident-io: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an
// empty incremental cursor (full sync). incident.io only supports full_refresh
// upstream, so the cursor is informational.
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
		stream = "incidents"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("incident-io stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	pages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, size, pages, emit)
}

// harvest drives incident.io's cursor pagination. List responses are shaped
// {<recordsKey>:[...], pagination_meta:{after:"..."}}; the next page is requested
// with after=<pagination_meta.after>. When pagination_meta.after is empty or
// missing, the list is exhausted. Non-paginated v1 endpoints simply return a
// single page (no pagination_meta.after) so the loop stops after one request.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	if endpoint.paginated {
		base.Set("page_size", strconv.Itoa(pageSize))
	}

	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if endpoint.paginated && after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read incident-io %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode incident-io %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if !endpoint.paginated {
			return nil
		}
		next, err := connsdk.StringAt(resp.Body, "pagination_meta.after")
		if err != nil {
			return fmt.Errorf("decode incident-io %s pagination_meta: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		after = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", endpoint.recordsKey, i),
			"reference":     fmt.Sprintf("INC-%d", i),
			"name":          fmt.Sprintf("Fixture incident %d", i),
			"summary":       "deterministic fixture record",
			"mode":          "test",
			"visibility":    "public",
			"description":   "deterministic fixture record",
			"role_type":     "custom",
			"shortform":     fmt.Sprintf("role%d", i),
			"instructions":  "fixture",
			"rank":          int64(i),
			"email":         fmt.Sprintf("fixture+%d@example.com", i),
			"slack_user_id": fmt.Sprintf("U%05d", i),
			"role":          "responder",
			"incident_id":   "incidents_fixture_1",
			"title":         fmt.Sprintf("Fixture follow-up %d", i),
			"status":        "outstanding",
			"created_at":    fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updated_at":    fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"severity":      map[string]any{"id": "sev_fixture_1", "name": "Major"},
			"incident_status": map[string]any{
				"id": "status_fixture_1", "name": "Triage", "category": "active",
			},
			"base_role": map[string]any{"id": "role_fixture_1", "name": "Owner"},
			"assignee":  map[string]any{"id": "user_fixture_1", "name": "Fixture User"},
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

// Write satisfies the connectors.Connector interface. incident.io is read-only
// in this connector (Capabilities.Write is false), so any write request is
// rejected rather than silently dropped.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey := secret(cfg)
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("incident-io connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(apiKey),
		UserAgent: userAgent,
	}, nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// baseURL resolves and validates the base URL. The default is api.incident.io;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("incident-io config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("incident-io config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("incident-io config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("incident-io config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("incident-io config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("incident-io config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("incident-io config max_pages must be 0 for unlimited or a positive integer")
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
