// Package acuityscheduling implements the native pm Acuity Scheduling connector.
// It is a declarative-HTTP per-system connector following the stripe/close-com
// template: a thin package that composes the connsdk toolkit (Requester + HTTP
// Basic auth + RecordsAt extraction) with Acuity-specific stream definitions,
// endpoints, and page-number pagination.
//
// The Acuity Scheduling API uses HTTP Basic auth with the account User ID as the
// username and the API key as the password, list endpoints under
// https://acuityscheduling.com/api/v1 that return a JSON array at the root, and
// max/page pagination on /appointments (the other list endpoints return a single
// full array). The Acuity source is read-only (full-refresh); it exposes no
// reverse-ETL writes, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package acuityscheduling

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
	registryName          = "acuity-scheduling"
	acuityDefaultBaseURL  = "https://acuityscheduling.com/api/v1"
	acuityDefaultPageSize = 100
	acuityMaxPageSize     = 100
	acuityUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Acuity Scheduling connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Acuity Scheduling connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Acuity Scheduling",
		IntegrationType: "api",
		Description:     "Reads Acuity Scheduling appointments, clients, appointment types, calendars, and forms through the Acuity REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Acuity. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := acuityBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(acuityUsername(cfg)) == "" {
		return errors.New("acuity-scheduling connector requires config username")
	}
	if strings.TrimSpace(acuitySecret(cfg)) == "" {
		return errors.New("acuity-scheduling connector requires secret password")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the calendars list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "calendars", nil, nil, nil); err != nil {
		return fmt.Errorf("check acuity-scheduling: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: acuityStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "appointments"
	}
	endpoint, ok := acuityStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("acuity-scheduling stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := acuityPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := acuityMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Acuity's max/page pagination. Acuity list endpoints return a
// JSON array at the root; paginated endpoints (appointments) advance `page`
// (1-indexed) until a short or empty page is returned. Non-paginated endpoints
// issue a single request. The loop lives here (rather than connsdk.Harvest) so
// the single-request endpoints do not get spurious page params.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	page := 1
	for iter := 0; maxPages == 0 || iter < maxPages; iter++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("max", strconv.Itoa(pageSize))
			query.Set("page", strconv.Itoa(page))
		}

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read acuity-scheduling %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode acuity-scheduling %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// Non-paginated endpoints return everything in one request; a short or
		// empty page on a paginated endpoint signals the end of the data.
		if !endpoint.paginated || len(records) < pageSize || len(records) == 0 {
			return nil
		}
		page++
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise acuity-scheduling credential-free (mirrors
// stripe's fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                int64(1000 + i),
			"firstName":         "Fixture",
			"lastName":          fmt.Sprintf("%d", i),
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"phone":             "555-0100",
			"datetime":          fmt.Sprintf("2026-01-0%dT09:00:00-0800", i),
			"endTime":           fmt.Sprintf("2026-01-0%dT09:30:00-0800", i),
			"date":              fmt.Sprintf("January %d, 2026", i),
			"time":              "9:00am",
			"type":              "Intro Session",
			"appointmentTypeID": int64(42),
			"calendar":          "Main Calendar",
			"calendarID":        int64(7),
			"duration":          "30",
			"price":             "0.00",
			"paid":              "no",
			"amountPaid":        "0.00",
			"canceled":          false,
			"datetimeCreated":   "2026-01-01T00:00:00-0800",
			"name":              fmt.Sprintf("Fixture %d", i),
			"active":            true,
			"description":       "Fixture record",
			"category":          "General",
			"color":             "#000000",
			"private":           false,
			"replyTo":           fmt.Sprintf("fixture+%d@example.com", i),
			"location":          "Online",
			"timezone":          "America/Los_Angeles",
			"hidden":            false,
		}
		record := endpoint.mapRecord(item)
		record["connector"] = registryName
		record["stream"] = stream
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth (account User
// ID as username, API key as password) and the resolved base URL. The secret
// only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := acuityBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := acuityUsername(cfg)
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("acuity-scheduling connector requires config username")
	}
	secret := acuitySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("acuity-scheduling connector requires secret password")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, secret),
		UserAgent: acuityUserAgent,
	}, nil
}

func acuityUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func acuitySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// acuityBaseURL resolves and validates the base URL. The default is
// acuityscheduling.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func acuityBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return acuityDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("acuity-scheduling config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("acuity-scheduling config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("acuity-scheduling config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func acuityPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return acuityDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("acuity-scheduling config page_size must be an integer: %w", err)
	}
	if value < 1 || value > acuityMaxPageSize {
		return 0, fmt.Errorf("acuity-scheduling config page_size must be between 1 and %d", acuityMaxPageSize)
	}
	return value, nil
}

func acuityMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("acuity-scheduling config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("acuity-scheduling config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write satisfies the connectors.Connector interface. The Acuity source is
// read-only (no approved reverse-ETL actions), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
