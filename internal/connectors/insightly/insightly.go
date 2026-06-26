// Package insightly implements the native pm Insightly connector. It is a
// declarative-HTTP per-system connector built from the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + HTTP Basic auth + RecordsAt extraction + skip/top pagination)
// with Insightly-specific stream definitions and endpoints.
//
// Insightly's API uses HTTP Basic auth where the API token is the username and
// the password is blank, and offset (skip/top) pagination over a top-level JSON
// array. The connector is read-only.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package insightly

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
	insightlyDefaultPod      = "na1"
	insightlyDefaultPageSize = 100
	insightlyMaxPageSize     = 500
	insightlyUserAgent       = "polymetrics-go-cli"
	// insightlyFixtureUpdated is the deterministic DATE_UPDATED_UTC used by the
	// fixture-mode records.
	insightlyFixtureUpdated = "2026-01-01 00:00:00"
)

func init() {
	connectors.RegisterFactory("insightly", New)
}

// New returns the Insightly connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Insightly connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "insightly" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "insightly",
		DisplayName:     "Insightly",
		IntegrationType: "api",
		Description:     "Reads Insightly CRM contacts, organisations, opportunities, leads, projects, and tasks through the Insightly REST API v3.1.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Insightly. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := insightlyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(insightlyToken(cfg)) == "" {
		return errors.New("insightly connector requires secret token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the Contacts list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{"top": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "Contacts", query, nil, nil); err != nil {
		return fmt.Errorf("check insightly: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: insightlyStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: an Insightly stream starts
// with an empty incremental cursor (full sync), which the start_date config can
// raise at read time.
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
		stream = "contacts"
	}
	endpoint, ok := insightlyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("insightly stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := insightlyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := insightlyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Insightly's skip/top offset pagination. Insightly list endpoints
// return a top-level JSON array; the next page is requested by advancing skip by
// the page size. A page shorter than the requested size signals the end. There is
// no body token paginator in connsdk for this exact shape, so the loop lives here,
// built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	skip := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("top", strconv.Itoa(pageSize))
		query.Set("skip", strconv.Itoa(skip))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read insightly %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode insightly %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than requested) means there are no more records.
		if len(records) < pageSize {
			return nil
		}
		skip += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise insightly credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			endpoint.idField:    int64(i),
			"FIRST_NAME":        fmt.Sprintf("Fixture%d", i),
			"LAST_NAME":         "Example",
			"EMAIL_ADDRESS":     fmt.Sprintf("fixture+%d@example.com", i),
			"EMAIL":             fmt.Sprintf("fixture+%d@example.com", i),
			"PHONE":             "+1-555-0100",
			"TITLE":             "Fixture Task",
			"ORGANISATION_ID":   int64(100 + i),
			"ORGANISATION_NAME": fmt.Sprintf("Fixture Org %d", i),
			"OPPORTUNITY_NAME":  fmt.Sprintf("Fixture Opportunity %d", i),
			"OPPORTUNITY_STATE": "OPEN",
			"OPPORTUNITY_VALUE": float64(1000 * i),
			"PROBABILITY":       int64(50),
			"BID_CURRENCY":      "USD",
			"PROJECT_NAME":      fmt.Sprintf("Fixture Project %d", i),
			"STATUS":            "IN PROGRESS",
			"PRIORITY":          int64(2),
			"COMPLETED":         false,
			"DUE_DATE":          "2026-02-01 00:00:00",
			"CONVERTED":         false,
			"LEAD_STATUS_ID":    int64(1),
			"LEAD_SOURCE_ID":    int64(1),
			"PIPELINE_ID":       int64(1),
			"STAGE_ID":          int64(1),
			"OWNER_USER_ID":     int64(1),
			"WEBSITE":           "https://example.com",
			"DATE_CREATED_UTC":  insightlyFixtureUpdated,
			"DATE_UPDATED_UTC":  insightlyFixtureUpdated,
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

// requester builds a connsdk.Requester wired with HTTP Basic auth (token as
// username, blank password) and the resolved base URL. The secret only ever flows
// into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := insightlyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := insightlyToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("insightly connector requires secret token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(token, ""),
		UserAgent: insightlyUserAgent,
	}, nil
}

func insightlyToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["token"]
}

// insightlyBaseURL resolves and validates the base URL. By default it is derived
// from the configured pod (e.g. na1 -> https://api.na1.insightly.com/v3.1). Any
// explicit base_url override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func insightlyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		pod := strings.TrimSpace(cfg.Config["pod"])
		if pod == "" {
			pod = insightlyDefaultPod
		}
		if err := validatePod(pod); err != nil {
			return "", err
		}
		return fmt.Sprintf("https://api.%s.insightly.com/v3.1", pod), nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("insightly config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("insightly config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("insightly config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// validatePod ensures the pod identifier is a simple alphanumeric token so it
// cannot inject extra host/path segments into the derived base URL.
func validatePod(pod string) error {
	for _, r := range pod {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		default:
			return fmt.Errorf("insightly config pod must be alphanumeric, got %q", pod)
		}
	}
	return nil
}

func insightlyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return insightlyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("insightly config page_size must be an integer: %w", err)
	}
	if value < 1 || value > insightlyMaxPageSize {
		return 0, fmt.Errorf("insightly config page_size must be between 1 and %d", insightlyMaxPageSize)
	}
	return value, nil
}

func insightlyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("insightly config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("insightly config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: the Insightly connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
