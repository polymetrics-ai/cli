// Package fleetio implements the native pm Fleetio connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, modeled on
// the stripe reference connector: a thin package that composes a connsdk
// Requester (Token + Account-Token auth headers, cursor pagination) with
// Fleetio-specific stream definitions, endpoints, and record mappers.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Fleetio's REST API (https://secure.fleetio.com/api/v1) authenticates with two
// headers: "Authorization: Token <api_key>" and "Account-Token: <account_token>".
// Index endpoints use cursor pagination: pass start_cursor (blank for the first
// page) and per_page; responses wrap the rows in a "records" array and expose
// "next_cursor". The connector is read-only.
package fleetio

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
	fleetioDefaultBaseURL  = "https://secure.fleetio.com/api/v1"
	fleetioDefaultPageSize = 100
	fleetioMaxPageSize     = 100
	fleetioUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("fleetio", New)
}

// New returns the Fleetio connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Fleetio connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "fleetio" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "fleetio",
		DisplayName:     "Fleetio",
		IntegrationType: "api",
		Description:     "Reads Fleetio fleet management data: vehicles, contacts, fuel entries, issues, and service entries through the Fleetio REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Fleetio. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := fleetioBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(fleetioAPIKey(cfg)) == "" {
		return errors.New("fleetio connector requires secret api_key")
	}
	if strings.TrimSpace(fleetioAccountToken(cfg)) == "" {
		return errors.New("fleetio connector requires secret account_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the vehicles list confirms auth and connectivity
	// without mutating anything.
	q := url.Values{"per_page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "vehicles", q, nil, nil); err != nil {
		return fmt.Errorf("check fleetio: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: fleetioStreams()}, nil
}

// Write is unsupported: the Fleetio connector is read-only. It satisfies the
// connectors.Connector interface while signaling that reverse ETL is not offered.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Fleetio stream starts with
// an empty cursor (Fleetio supports full refresh only).
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
		stream = "vehicles"
	}
	endpoint, ok := fleetioStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("fleetio stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := fleetioPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := fleetioMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Fleetio's cursor pagination. Index endpoints return
// {records:[...], next_cursor:"..."}; the next page is requested with
// start_cursor=<next_cursor>. The loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt, because connsdk's CursorPaginator does
// not stop on a null/absent token the same way and we want explicit page bounds.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))

	startCursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if startCursor != "" {
			query.Set("start_cursor", startCursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read fleetio %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "records")
		if err != nil {
			return fmt.Errorf("decode fleetio %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_cursor")
		if err != nil {
			return fmt.Errorf("decode fleetio %s next_cursor: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || len(records) == 0 {
			return nil
		}
		startCursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise fleetio credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  int64(i),
			"name":                fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"vin":                 fmt.Sprintf("VINFIXTURE%05d", i),
			"make":                "Acme",
			"model":               "FixtureLine",
			"year":                int64(2026),
			"license_plate":       fmt.Sprintf("FIX-%04d", i),
			"vehicle_status_name": "Active",
			"vehicle_type_name":   "Truck",
			"current_meter_value": fmt.Sprintf("%d", 1000*i),
			"first_name":          fmt.Sprintf("First%d", i),
			"last_name":           fmt.Sprintf("Last%d", i),
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"group_name":          "Operations",
			"technician":          true,
			"employee":            true,
			"vehicle_id":          int64(1),
			"vehicle_name":        "Truck 1",
			"date":                "2026-01-01T00:00:00Z",
			"us_gallons":          fmt.Sprintf("%d.0", 10*i),
			"cost":                fmt.Sprintf("%d.00", 40*i),
			"total_amount":        fmt.Sprintf("%d.00", 40*i),
			"meter_value":         fmt.Sprintf("%d", 1000*i),
			"is_sample":           true,
			"number":              int64(100 + i),
			"summary":             fmt.Sprintf("Issue %d", i),
			"description":         "Fixture issue description",
			"state":               "open",
			"due_date":            "2026-02-01T00:00:00Z",
			"resolved_at":         nil,
			"started_at":          "2026-01-01T00:00:00Z",
			"completed_at":        "2026-01-02T00:00:00Z",
			"labor_subtotal":      fmt.Sprintf("%d.00", 50*i),
			"parts_subtotal":      fmt.Sprintf("%d.00", 30*i),
			"archived_at":         nil,
			"created_at":          "2026-01-01T00:00:00Z",
			"updated_at":          fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
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

// requester builds a connsdk.Requester wired with the Fleetio auth headers, the
// resolved base URL, and the Account-Token header. The secrets only ever flow
// into the auth header / DefaultHeaders; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := fleetioBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey := fleetioAPIKey(cfg)
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("fleetio connector requires secret api_key")
	}
	accountToken := fleetioAccountToken(cfg)
	if strings.TrimSpace(accountToken) == "" {
		return nil, errors.New("fleetio connector requires secret account_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", apiKey, "Token "),
		UserAgent: fleetioUserAgent,
		DefaultHeaders: map[string]string{
			"Account-Token": strings.TrimSpace(accountToken),
		},
	}, nil
}

func fleetioAPIKey(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func fleetioAccountToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["account_token"]
}

// fleetioBaseURL resolves and validates the base URL. The default is
// secure.fleetio.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func fleetioBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return fleetioDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("fleetio config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("fleetio config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("fleetio config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fleetioPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return fleetioDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fleetio config page_size must be an integer: %w", err)
	}
	if value < 1 || value > fleetioMaxPageSize {
		return 0, fmt.Errorf("fleetio config page_size must be between 1 and %d", fleetioMaxPageSize)
	}
	return value, nil
}

func fleetioMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fleetio config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("fleetio config max_pages must be 0 for unlimited or a positive integer")
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
