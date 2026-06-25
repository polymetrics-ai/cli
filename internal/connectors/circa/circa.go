// Package circa implements the native pm Circa connector. It follows the
// declarative-HTTP template established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction + cursor state) with Circa-specific stream definitions, endpoints,
// and pagination.
//
// Circa's REST API (https://app.circa.co/api/v1) authenticates with a bearer
// API key, returns list payloads shaped as {data:[...]}, and paginates by an
// incrementing `page` query parameter starting at 1. events/contacts/companies
// support incremental sync via updated_at[min]/updated_at[max]; teams is a full
// refresh. The connector is read-only.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package circa

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	circaDefaultBaseURL  = "https://app.circa.co/api/v1"
	circaDefaultPageSize = 25
	circaMaxPageSize     = 100
	circaUserAgent       = "polymetrics-go-cli"
	// circaFixtureUpdated is the deterministic updated_at base used by the
	// fixture-mode records.
	circaFixtureUpdated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("circa", New)
}

// New returns the Circa connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Circa connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "circa" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "circa",
		DisplayName:     "Circa",
		IntegrationType: "api",
		Description:     "Reads Circa events, contacts, companies, and teams through the Circa REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Circa. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := circaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(circaSecret(cfg)) == "" {
		return errors.New("circa connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the teams list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "teams", q, nil, nil); err != nil {
		return fmt.Errorf("check circa: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: circaStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Circa stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
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
	endpoint, ok := circaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("circa stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := circaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := circaMaxPages(req.Config)
	if err != nil {
		return err
	}
	lower := ""
	if endpoint.incremental {
		if lower, err = incrementalLowerBound(req); err != nil {
			return err
		}
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives Circa's page-increment pagination. Circa lists return
// {data:[...]}; pages start at 1 and advance while a full page (len == pageSize)
// is returned. A short or empty page ends the loop. There is no body-token
// paginator in connsdk for this exact shape, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, updatedAtMin string, emit func(connectors.Record) error) error {
	base := url.Values{}
	if updatedAtMin != "" {
		base.Set("updated_at[min]", updatedAtMin)
	}

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read circa %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode circa %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than pageSize records) means the last page was
		// reached. An empty page also terminates.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise circa credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":           fmt.Sprintf("Fixture %s %d", stream, i),
			"status":         "live",
			"website":        "https://example.com",
			"brief_url":      "https://example.com/brief",
			"time_zone":      "UTC",
			"paid_total":     float64(100 * i),
			"actual_total":   float64(120 * i),
			"planned_total":  float64(150 * i),
			"email":          fmt.Sprintf("fixture+%d@example.com", i),
			"first_name":     "Fixture",
			"last_name":      fmt.Sprintf("User%d", i),
			"company":        map[string]any{"id": "co_fixture_1", "name": "Fixture Co"},
			"email_opt_in":   true,
			"sync_status":    map[string]any{"status": "synced"},
			"created_method": "api",
			"updated_method": "api",
			"created_by":     map[string]any{"id": "u_fixture_1", "email": "owner@example.com"},
			"created_at":     circaFixtureUpdated,
			"updated_at":     circaFixtureUpdated,
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := circaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := circaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("circa connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: circaUserAgent,
	}, nil
}

// incrementalLowerBound returns the RFC3339 lower bound for updated_at[min],
// derived from the incremental cursor (if any) or else the start_date config.
// An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	if _, err := time.Parse(time.RFC3339, startDate); err != nil {
		return "", fmt.Errorf("circa config start_date must be RFC3339: %w", err)
	}
	return startDate, nil
}

func circaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// circaBaseURL resolves and validates the base URL. The default is
// app.circa.co; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func circaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return circaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("circa config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("circa config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("circa config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func circaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return circaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("circa config page_size must be an integer: %w", err)
	}
	if value < 1 || value > circaMaxPageSize {
		return 0, fmt.Errorf("circa config page_size must be between 1 and %d", circaMaxPageSize)
	}
	return value, nil
}

func circaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("circa config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("circa config max_pages must be 0 for unlimited or a positive integer")
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

// Write satisfies the connectors.Connector interface. Circa is read-only, so
// writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
