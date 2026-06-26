// Package justcall implements the native pm JustCall connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit (Requester
// + APIKeyHeader auth + RecordsAt extraction) plus JustCall-specific stream
// definitions, endpoints, and page-increment pagination. It mirrors the stripe
// reference connector's shape.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
package justcall

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
	justcallDefaultBaseURL  = "https://api.justcall.io"
	justcallDefaultPageSize = 100
	justcallMaxPageSize     = 100
	justcallUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("justcall", New)
}

// New returns the JustCall connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm JustCall connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "justcall" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "justcall",
		DisplayName:     "JustCall",
		IntegrationType: "api",
		Description:     "Reads JustCall users, call logs, SMS, contacts, and phone numbers through the JustCall REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to JustCall. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := justcallBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(justcallSecret(cfg)) == "" {
		return errors.New("justcall connector requires secret api_key_2")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"page": []string{"0"}, "per_page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "v2.1/users", q, nil, nil); err != nil {
		return fmt.Errorf("check justcall: %w", err)
	}
	return nil
}

// Write is unsupported: JustCall is a read-only source here. The method exists
// only to satisfy connectors.Connector; Capabilities.Write is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: justcallStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "calls"
	}
	endpoint, ok := justcallStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("justcall stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := justcallPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := justcallMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, req, pageSize, maxPages, emit)
}

// harvest drives JustCall's page-increment pagination. JustCall list endpoints
// return {data:[...]}; pages are numbered from 0 via the `page` query parameter
// with `per_page` as the page size. There is no body token, so the loop stops on
// the first short page (or after one request for non-paginated endpoints). The
// v1 list endpoints use POST with an empty JSON body; the v2.1 endpoints use GET.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, req connectors.ReadRequest, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	if lower := incrementalLowerBound(endpoint, req); lower != "" {
		base.Set("from_datetime", lower)
	}

	method := endpoint.method
	if method == "" {
		method = http.MethodGet
	}

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if endpoint.paginated {
			query.Set("page", strconv.Itoa(page))
			query.Set("per_page", strconv.Itoa(pageSize))
		}

		var resp *connsdk.Response
		var err error
		if method == http.MethodPost {
			// v1 list endpoints take a JSON body; an empty object is accepted
			// and lists all records subject to the page params.
			resp, err = r.Do(ctx, http.MethodPost, endpoint.resource, query, map[string]any{})
		} else {
			resp, err = r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		}
		if err != nil {
			return fmt.Errorf("read justcall %s: %w", endpoint.resource, err)
		}

		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode justcall %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		// Non-paginated endpoints return everything in one response; stop after
		// the first request. Paginated endpoints stop on a short/empty page.
		if !endpoint.paginated || len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise justcall credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                 fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":               fmt.Sprintf("Fixture %d", i),
			"email":              fmt.Sprintf("fixture+%d@example.com", i),
			"role":               "agent",
			"call_date":          "2026-01-0" + strconv.Itoa(i),
			"call_time":          "10:00:00",
			"sms_date":           "2026-01-0" + strconv.Itoa(i),
			"sms_time":           "10:00:00",
			"direction":          "Outgoing",
			"delivery_status":    "delivered",
			"agent_id":           "agent_fixture_1",
			"agent_name":         fmt.Sprintf("Fixture %d", i),
			"agent_email":        fmt.Sprintf("fixture+%d@example.com", i),
			"contact_name":       fmt.Sprintf("Contact %d", i),
			"contact_number":     "+15555550000",
			"call_duration":      "60",
			"cost_incurred":      "0",
			"justcall_number":    "+15555551111",
			"justcall_line_name": "Fixture Line",
			"firstname":          fmt.Sprintf("Fixture %d", i),
			"lastname":           "Contact",
			"phone":              "+15555550000",
			"company":            "Example",
			"friendly_name":      "Fixture Line",
			"custom_name":        "Fixture Line",
			"capabilities":       "voice,sms",
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

// requester builds a connsdk.Requester wired with JustCall's API-key header auth
// and the resolved base URL. The secret only ever flows into APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := justcallBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := justcallSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("justcall connector requires secret api_key_2")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, ""),
		UserAgent: justcallUserAgent,
	}, nil
}

// incrementalLowerBound returns the from_datetime lower bound for cursor-bearing
// streams (calls, sms), derived from the incremental cursor (if any) or else the
// start_date config. Streams without a cursor return "".
func incrementalLowerBound(endpoint streamEndpoint, req connectors.ReadRequest) string {
	if endpoint.resource != "v2.1/calls" && endpoint.resource != "v2.1/texts" {
		return ""
	}
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func justcallSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key_2"]
}

// justcallBaseURL resolves and validates the base URL. The default is
// api.justcall.io; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func justcallBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return justcallDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("justcall config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("justcall config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("justcall config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func justcallPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return justcallDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("justcall config page_size must be an integer: %w", err)
	}
	if value < 1 || value > justcallMaxPageSize {
		return 0, fmt.Errorf("justcall config page_size must be between 1 and %d", justcallMaxPageSize)
	}
	return value, nil
}

func justcallMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("justcall config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("justcall config max_pages must be 0 for unlimited or a positive integer")
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
