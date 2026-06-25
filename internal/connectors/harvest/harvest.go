// Package harvest implements the native pm Harvest connector. It follows the
// stripe reference shape for declarative-HTTP per-system connectors: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + cursor state) with Harvest-specific stream definitions,
// endpoints, and pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Harvest API v2 reference: https://help.getharvest.com/api-v2/. Auth is a
// personal access token (Bearer) plus the Harvest-Account-Id header; list
// endpoints return {"<resource>":[...], "page":N, "next_page":N|null, ...} and
// are walked by page number until next_page is null. Read-only: Harvest writes
// are not exposed for reverse ETL here.
package harvest

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
	harvestDefaultBaseURL  = "https://api.harvestapp.com/v2"
	harvestDefaultPageSize = 100
	harvestMaxPageSize     = 2000
	harvestUserAgent       = "polymetrics-go-cli"
	// harvestFixtureUpdated is the deterministic updated_at timestamp used by
	// fixture-mode records.
	harvestFixtureUpdated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("harvest", New)
}

// New returns the Harvest connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Harvest connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "harvest" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "harvest",
		DisplayName:     "Harvest",
		IntegrationType: "api",
		Description:     "Reads Harvest clients, projects, tasks, users, and time entries through the Harvest v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Harvest. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := harvestBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(harvestToken(cfg)) == "" {
		return errors.New("harvest connector requires secret credentials.api_token")
	}
	if strings.TrimSpace(harvestAccountID(cfg)) == "" {
		return errors.New("harvest connector requires account_id")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the clients list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "clients", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check harvest: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: harvestStreams()}, nil
}

// Write is unsupported: this connector is read-only (Capabilities.Write=false).
// The method exists to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Harvest stream starts with
// an empty incremental cursor (full sync), which the replication_start_date
// config can raise at read time.
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
		stream = "clients"
	}
	endpoint, ok := harvestStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("harvest stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	lower, err := incrementalLowerBound(req)
	if err != nil {
		return err
	}
	pageSize, err := harvestPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := harvestMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives Harvest's page-number pagination. List responses carry the
// records under endpoint.recordKey plus a top-level next_page field (the number
// of the next page, or null when exhausted). The loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt, because the records
// key varies per stream and the stop condition reads a body field.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, updatedSince string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	if updatedSince != "" {
		base.Set("updated_since", updatedSince)
	}

	page := 1
	for pageNum := 0; maxPages == 0 || pageNum < maxPages; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read harvest %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordKey)
		if err != nil {
			return fmt.Errorf("decode harvest %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		nextPage, err := connsdk.StringAt(resp.Body, "next_page")
		if err != nil {
			return fmt.Errorf("decode harvest %s next_page: %w", endpoint.resource, err)
		}
		next, ok := parsePage(nextPage)
		if !ok || next <= page {
			return nil
		}
		page = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise harvest credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  int64(i),
			"name":                fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"is_active":           true,
			"is_billable":         true,
			"billable":            true,
			"is_billed":           false,
			"is_running":          false,
			"is_admin":            false,
			"is_default":          false,
			"address":             "123 Fixture St.",
			"statement_key":       fmt.Sprintf("key_%d", i),
			"currency":            "USD",
			"code":                fmt.Sprintf("CODE-%d", i),
			"budget":              float64(1000 * i),
			"first_name":          fmt.Sprintf("First%d", i),
			"last_name":           fmt.Sprintf("Last%d", i),
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"timezone":            "UTC",
			"billable_by_default": true,
			"default_hourly_rate": float64(100 * i),
			"spent_date":          "2026-01-01",
			"hours":               float64(i),
			"notes":               fmt.Sprintf("fixture note %d", i),
			"client":              map[string]any{"id": int64(100 + i), "name": fmt.Sprintf("Client %d", i)},
			"project":             map[string]any{"id": int64(200 + i), "name": fmt.Sprintf("Project %d", i)},
			"user":                map[string]any{"id": int64(300 + i), "name": fmt.Sprintf("User %d", i)},
			"task":                map[string]any{"id": int64(400 + i), "name": fmt.Sprintf("Task %d", i)},
			"created_at":          harvestFixtureUpdated,
			"updated_at":          harvestFixtureUpdated,
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

// requester builds a connsdk.Requester wired with Bearer auth, the resolved base
// URL, and the required Harvest-Account-Id header. The secret only ever flows
// into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := harvestBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := harvestToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("harvest connector requires secret credentials.api_token")
	}
	account := strings.TrimSpace(harvestAccountID(cfg))
	if account == "" {
		return nil, errors.New("harvest connector requires account_id")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(token),
		UserAgent: harvestUserAgent,
		DefaultHeaders: map[string]string{
			"Harvest-Account-Id": account,
		},
	}, nil
}

// incrementalLowerBound returns the RFC3339 updated_since lower bound, derived
// from the incremental cursor (if any) or else the replication_start_date
// config. An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["replication_start_date"])
	if startDate == "" {
		startDate = strings.TrimSpace(req.Config.Config["start_date"])
	}
	if startDate == "" {
		return "", nil
	}
	if _, err := time.Parse(time.RFC3339, startDate); err != nil {
		return "", fmt.Errorf("harvest config replication_start_date must be RFC3339: %w", err)
	}
	return startDate, nil
}

// harvestToken resolves the Harvest personal access token from secrets. It
// accepts both the dotted catalog key and a flat fallback.
func harvestToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Secrets["credentials.api_token"]); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Secrets["api_token"])
}

// harvestAccountID resolves the Harvest account id. It is a secret in the
// catalog but is sent as a request header, so it is read from secrets first and
// from config as a fallback for local/test wiring.
func harvestAccountID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets["account_id"]); v != "" {
			return v
		}
	}
	if cfg.Config != nil {
		return strings.TrimSpace(cfg.Config["account_id"])
	}
	return ""
}

// harvestBaseURL resolves and validates the base URL. The default is
// api.harvestapp.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func harvestBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return harvestDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("harvest config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("harvest config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("harvest config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func harvestPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return harvestDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("harvest config page_size must be an integer: %w", err)
	}
	if value < 1 || value > harvestMaxPageSize {
		return 0, fmt.Errorf("harvest config page_size must be between 1 and %d", harvestMaxPageSize)
	}
	return value, nil
}

func harvestMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("harvest config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("harvest config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// parsePage parses the next_page body field. Harvest renders it as an integer
// page number or null; "" and "null" both mean no further pages.
func parsePage(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if value == "" || value == "null" {
		return 0, false
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
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
