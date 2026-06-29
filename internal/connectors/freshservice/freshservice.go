// Package freshservice implements the native pm Freshservice connector. It
// follows the declarative-HTTP per-system template established by the stripe and
// freshdesk packages: a thin package that composes the connsdk toolkit
// (Requester + HTTP Basic auth + plural-key extraction + page-number pagination
// + cursor state) with Freshservice-specific stream definitions, endpoints, and
// a fixture mode.
//
// Freshservice's API v2 authenticates with HTTP Basic where the username is the
// API key and the password is the literal "X". List endpoints wrap their record
// array in a plural key (e.g. {"tickets":[...]}) and paginate by page number
// (page + per_page, max 100); a short page signals the end.
//
// Like stripe and freshdesk, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package freshservice

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
	freshserviceDefaultPageSize = 100
	freshserviceMaxPageSize     = 100
	freshserviceUserAgent       = "polymetrics-go-cli"
	// freshserviceBasicPassword is the literal password Freshservice expects
	// alongside the API key (which is sent as the Basic-auth username).
	freshserviceBasicPassword = "X"
)

func init() {
	connectors.RegisterFactory("freshservice", New)
}

// New returns the Freshservice connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Freshservice connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "freshservice" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "freshservice",
		DisplayName:     "Freshservice",
		IntegrationType: "api",
		Description:     "Reads Freshservice tickets, agents, requesters, assets, and problems through the Freshservice REST API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to
// Freshservice. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := freshserviceBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(freshserviceSecret(cfg)) == "" {
		return errors.New("freshservice connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the agents list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "agents", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check freshservice: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: freshserviceStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Freshservice stream starts
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
		stream = "tickets"
	}
	endpoint, ok := freshserviceStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("freshservice stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := freshservicePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := freshserviceMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	if lower := strings.TrimSpace(req.Config.Config["start_date"]); lower != "" && stream == "tickets" {
		// Tickets support server-side filtering by updated_since; other streams
		// would reject the param so it is only applied to tickets.
		base.Set("updated_since", lower)
	}

	// Freshservice paginates by page number (page + per_page); a page shorter
	// than per_page signals the last page, which is exactly the contract of
	// connsdk's PageNumberPaginator. The record array is wrapped in a plural
	// key, supplied here as the records path.
	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "per_page",
		StartPage: 1,
		PageSize:  pageSize,
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, endpoint.listKey, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(map[string]any(rec)))
	}); err != nil {
		return fmt.Errorf("read freshservice %s: %w", endpoint.resource, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise freshservice credential-free (mirrors
// stripe's fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               int64(i),
			"display_id":       int64(100 + i),
			"subject":          fmt.Sprintf("Fixture %s %d", stream, i),
			"description_text": fmt.Sprintf("Fixture %s record %d", stream, i),
			"status":           int64(2),
			"priority":         int64(1),
			"impact":           int64(1),
			"source":           int64(2),
			"type":             "Incident",
			"first_name":       fmt.Sprintf("Fixture%d", i),
			"last_name":        "User",
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"primary_email":    fmt.Sprintf("fixture+%d@example.com", i),
			"job_title":        "Engineer",
			"name":             fmt.Sprintf("Fixture asset %d", i),
			"asset_type_id":    int64(7000),
			"asset_tag":        fmt.Sprintf("TAG-%d", i),
			"requester_id":     int64(1000 + i),
			"responder_id":     int64(2000 + i),
			"agent_id":         int64(2000 + i),
			"user_id":          int64(1000 + i),
			"group_id":         int64(3000),
			"department_id":    int64(4000),
			"department_ids":   []any{int64(4000)},
			"active":           true,
			"occasional":       false,
			"spam":             false,
			"due_by":           "2026-01-10T00:00:00Z",
			"created_at":       fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updated_at":       fmt.Sprintf("2026-01-0%dT12:00:00Z", i),
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

// requester builds a connsdk.Requester wired with HTTP Basic auth (api_key:X)
// and the resolved base URL. The secret only ever flows into connsdk.Basic; it
// is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := freshserviceBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := freshserviceSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("freshservice connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, freshserviceBasicPassword),
		UserAgent: freshserviceUserAgent,
	}, nil
}

func freshserviceSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// freshserviceBaseURL resolves and validates the base URL. Precedence:
//  1. an explicit base_url override (used by tests and proxies); or
//  2. the domain_name config (e.g. acme.freshservice.com) turned into
//     https://<domain>/api/v2.
//
// Any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func freshserviceBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("freshservice config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("freshservice config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("freshservice config base_url must include a host")
		}
		return strings.TrimRight(base, "/") + "/api/v2", nil
	}

	// The upstream catalog names this config field domain_name; accept the
	// shorter "domain" alias too for ergonomics.
	domain := strings.TrimSpace(cfg.Config["domain_name"])
	if domain == "" {
		domain = strings.TrimSpace(cfg.Config["domain"])
	}
	if domain == "" {
		return "", errors.New("freshservice connector requires config domain_name (e.g. acme.freshservice.com) or base_url")
	}
	// Strip any scheme the user may have included, then validate the host shape.
	domain = strings.TrimPrefix(strings.TrimPrefix(domain, "https://"), "http://")
	domain = strings.TrimRight(domain, "/")
	if domain == "" || strings.ContainsAny(domain, "/?#") {
		return "", fmt.Errorf("freshservice config domain_name must be a bare host, got %q", cfg.Config["domain_name"])
	}
	parsed, err := url.Parse("https://" + domain)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("freshservice config domain_name is invalid: %q", cfg.Config["domain_name"])
	}
	return "https://" + parsed.Host + "/api/v2", nil
}

func freshservicePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return freshserviceDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("freshservice config page_size must be an integer: %w", err)
	}
	if value < 1 || value > freshserviceMaxPageSize {
		return 0, fmt.Errorf("freshservice config page_size must be between 1 and %d", freshserviceMaxPageSize)
	}
	return value, nil
}

func freshserviceMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("freshservice config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("freshservice config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Freshservice is exposed read-only
// (Capabilities.Write=false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
