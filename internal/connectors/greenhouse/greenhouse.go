// Package greenhouse implements the native pm Greenhouse (Harvest API) connector.
// It is a declarative-HTTP per-system connector built on the same shape as the
// stripe reference connector: a thin package that composes the connsdk toolkit
// (Requester + Basic auth + RecordsAt extraction + Link-header pagination) with
// Greenhouse-specific stream definitions and endpoints.
//
// Greenhouse Harvest authenticates with HTTP Basic auth where the API token is
// the username and the password is blank, lists return top-level JSON arrays,
// and pagination follows the RFC-5988 Link header (rel="next"). It is read-only;
// no reverse-ETL write actions are exposed.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package greenhouse

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	greenhouseDefaultBaseURL  = "https://harvest.greenhouse.io/v1"
	greenhouseDefaultPageSize = 100
	greenhouseMaxPageSize     = 500
	greenhouseUserAgent       = "polymetrics-go-cli"
	// greenhouseFixtureUpdated is the deterministic updated_at used by fixture
	// records (2026-01-01T00:00:00Z).
	greenhouseFixtureUpdated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("greenhouse", New)
}

// New returns the Greenhouse connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Greenhouse connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "greenhouse" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "greenhouse",
		DisplayName:     "Greenhouse",
		IntegrationType: "api",
		Description:     "Reads Greenhouse candidates, applications, jobs, offers, and users through the Greenhouse Harvest REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Greenhouse.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := greenhouseBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(greenhouseSecret(cfg)) == "" {
		return errors.New("greenhouse connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the jobs list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "jobs", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check greenhouse: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: greenhouseStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Greenhouse stream starts
// with an empty incremental cursor (full sync). Greenhouse only supports
// full_refresh upstream, so the cursor is informational.
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
		stream = "candidates"
	}
	endpoint, ok := greenhouseStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("greenhouse stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := greenhousePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := greenhouseMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))

	paginator := &connsdk.LinkHeaderPaginator{}
	// Greenhouse list endpoints return a top-level JSON array, so the records
	// path is the root ("").
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, "", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	}); err != nil {
		return fmt.Errorf("read greenhouse %s: %w", endpoint.resource, err)
	}
	return nil
}

// Write is unsupported: the Greenhouse connector is read-only. It satisfies the
// connectors.Connector interface but always returns ErrUnsupportedOperation, and
// Metadata reports Write=false so reverse-ETL is never attempted against it.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise greenhouse credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                    int64(i),
			"first_name":            fmt.Sprintf("Fixture%d", i),
			"last_name":             "Candidate",
			"name":                  fmt.Sprintf("Fixture %s %d", stream, i),
			"company":               "Example Inc",
			"title":                 "Engineer",
			"status":                "active",
			"candidate_id":          int64(100 + i),
			"application_id":        int64(200 + i),
			"requisition_id":        fmt.Sprintf("REQ-%d", i),
			"primary_email_address": fmt.Sprintf("fixture+%d@example.com", i),
			"disabled":              false,
			"site_admin":            false,
			"is_private":            false,
			"confidential":          false,
			"version":               int64(1),
			"source_id":             int64(7),
			"employee_id":           fmt.Sprintf("E%d", i),
			"created_at":            greenhouseFixtureUpdated,
			"updated_at":            greenhouseFixtureUpdated,
			"last_activity":         greenhouseFixtureUpdated,
			"last_activity_at":      greenhouseFixtureUpdated,
			"applied_at":            greenhouseFixtureUpdated,
			"opened_at":             greenhouseFixtureUpdated,
			"starts_at":             greenhouseFixtureUpdated,
			"sent_at":               greenhouseFixtureUpdated,
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

// requester builds a connsdk.Requester wired with Basic auth (API token as
// username, blank password), the resolved base URL. The secret only ever flows
// into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := greenhouseBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := greenhouseSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("greenhouse connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: greenhouseUserAgent,
	}, nil
}

func greenhouseSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// greenhouseBaseURL resolves and validates the base URL. The default is
// harvest.greenhouse.io; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func greenhouseBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return greenhouseDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("greenhouse config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("greenhouse config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("greenhouse config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func greenhousePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return greenhouseDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("greenhouse config page_size must be an integer: %w", err)
	}
	if value < 1 || value > greenhouseMaxPageSize {
		return 0, fmt.Errorf("greenhouse config page_size must be between 1 and %d", greenhouseMaxPageSize)
	}
	return value, nil
}

func greenhouseMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("greenhouse config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("greenhouse config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
