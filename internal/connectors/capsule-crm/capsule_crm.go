// Package capsulecrm implements the native pm Capsule CRM connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit (Requester +
// Bearer auth + RecordsAt extraction + page-number pagination + cursor state),
// following the stripe reference connector's shape.
//
// Capsule CRM v2 (https://developer.capsulecrm.com/) wraps each list endpoint's
// records under a top-level key matching the resource (e.g. {"parties":[...]})
// and paginates with ?page=N&perPage=M (1-based, default 50, max 100), signalling
// the end of data with a short page. This connector is read-only: the API has no
// obviously-safe reverse-ETL surface to allow-list, so Capabilities.Write=false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package capsulecrm

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
	capsuleDefaultBaseURL  = "https://api.capsulecrm.com/api/v2"
	capsuleDefaultPageSize = 50
	capsuleMaxPageSize     = 100
	capsuleUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("capsule-crm", New)
}

// New returns the Capsule CRM connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Capsule CRM connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "capsule-crm" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "capsule-crm",
		DisplayName:     "Capsule CRM",
		IntegrationType: "api",
		Description:     "Reads Capsule CRM parties, opportunities, cases, tasks, and users through the Capsule v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Capsule. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := capsuleBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(capsuleSecret(cfg)) == "" {
		return errors.New("capsule-crm connector requires secret bearer_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the parties list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"perPage": []string{"1"}, "page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "parties", q, nil, nil); err != nil {
		return fmt.Errorf("check capsule-crm: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: capsuleStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Capsule stream starts with
// an empty incremental cursor (full sync). The supported sync mode upstream is
// full_refresh; the cursor field is retained for forward compatibility.
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
		stream = "parties"
	}
	endpoint, ok := capsuleStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("capsule-crm stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := capsulePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := capsuleMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Capsule paginates with 1-based page + perPage and stops on a short page,
	// which is exactly connsdk.PageNumberPaginator's contract.
	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "perPage",
		StartPage: 1,
		PageSize:  pageSize,
	}
	base := url.Values{}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, endpoint.recordsKey, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// Write satisfies the connectors.Connector interface. Capsule CRM is exposed
// read-only (no allow-listed reverse-ETL actions), so writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise capsule-crm credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               i,
			"type":             "person",
			"firstName":        fmt.Sprintf("Fixture%d", i),
			"lastName":         "Example",
			"organisationName": fmt.Sprintf("Fixture Org %d", i),
			"name":             fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"description":      "fixture record",
			"status":           "OPEN",
			"username":         fmt.Sprintf("fixture%d", i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"createdAt":        "2026-01-01T00:00:00Z",
			"updatedAt":        fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"value":            map[string]any{"amount": 1000 * i, "currency": "USD"},
			"milestone":        map[string]any{"id": 1, "name": "Lead"},
			"party":            map[string]any{"id": 1},
			"category":         map[string]any{"id": 2},
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
	base, err := capsuleBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := capsuleSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("capsule-crm connector requires secret bearer_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: capsuleUserAgent,
	}, nil
}

func capsuleSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["bearer_token"]
}

// capsuleBaseURL resolves and validates the base URL. The default is
// api.capsulecrm.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func capsuleBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return capsuleDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("capsule-crm config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("capsule-crm config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("capsule-crm config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func capsulePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return capsuleDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("capsule-crm config page_size must be an integer: %w", err)
	}
	if value < 1 || value > capsuleMaxPageSize {
		return 0, fmt.Errorf("capsule-crm config page_size must be between 1 and %d", capsuleMaxPageSize)
	}
	return value, nil
}

func capsuleMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("capsule-crm config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("capsule-crm config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
