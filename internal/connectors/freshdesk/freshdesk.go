// Package freshdesk implements the native pm Freshdesk connector. It follows the
// declarative-HTTP per-system template established by the stripe package: a thin
// package that composes the connsdk toolkit (Requester + HTTP Basic auth +
// top-level-array extraction + Link-header pagination + cursor state) with
// Freshdesk-specific stream definitions, endpoints, and a fixture mode.
//
// Freshdesk's API authenticates with HTTP Basic where the username is the API
// key and the password is the literal "X"; list endpoints return a top-level
// JSON array and advertise the next page through an RFC 5988 Link header.
//
// Like stripe and github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package freshdesk

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
	freshdeskDefaultPageSize = 100
	freshdeskMaxPageSize     = 100
	freshdeskUserAgent       = "polymetrics-go-cli"
	// freshdeskBasicPassword is the literal password Freshdesk expects alongside
	// the API key (which is sent as the Basic-auth username).
	freshdeskBasicPassword = "X"
)

func init() {
	connectors.RegisterFactory("freshdesk", New)
}

// New returns the Freshdesk connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Freshdesk connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "freshdesk" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "freshdesk",
		DisplayName:     "Freshdesk",
		IntegrationType: "api",
		Description:     "Reads Freshdesk tickets, contacts, companies, agents, and groups through the Freshdesk REST API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Freshdesk. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := freshdeskBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(freshdeskSecret(cfg)) == "" {
		return errors.New("freshdesk connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the agents list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "agents", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check freshdesk: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: freshdeskStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Freshdesk stream starts
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
	endpoint, ok := freshdeskStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("freshdesk stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := freshdeskPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := freshdeskMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	if lower := strings.TrimSpace(req.Config.Config["start_date"]); lower != "" {
		// Tickets support server-side filtering by updated_since; other streams
		// ignore the param harmlessly.
		if stream == "tickets" {
			base.Set("updated_since", lower)
		}
	}

	// Freshdesk advertises the next page via an RFC 5988 Link header, which
	// connsdk's LinkHeaderPaginator follows directly.
	paginator := &connsdk.LinkHeaderPaginator{FirstQuery: base}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, "", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(map[string]any(rec)))
	}); err != nil {
		return fmt.Errorf("read freshdesk %s: %w", endpoint.resource, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise freshdesk credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           int64(i),
			"subject":      fmt.Sprintf("Fixture ticket %d", i),
			"type":         "Question",
			"status":       int64(2),
			"priority":     int64(1),
			"source":       int64(2),
			"requester_id": int64(1000 + i),
			"responder_id": int64(2000 + i),
			"group_id":     int64(3000),
			"company_id":   int64(4000),
			"spam":         false,
			"name":         fmt.Sprintf("Fixture %d", i),
			"email":        fmt.Sprintf("fixture+%d@example.com", i),
			"phone":        "",
			"mobile":       "",
			"active":       true,
			"available":    true,
			"occasional":   false,
			"ticket_scope": int64(1),
			"description":  fmt.Sprintf("Fixture %s record %d", stream, i),
			"note":         "",
			"due_by":       "2026-01-10T00:00:00Z",
			"created_at":   fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updated_at":   fmt.Sprintf("2026-01-0%dT12:00:00Z", i),
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

// requester builds a connsdk.Requester wired with HTTP Basic auth (api_key:X) and
// the resolved base URL. The secret only ever flows into connsdk.Basic; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := freshdeskBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := freshdeskSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("freshdesk connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, freshdeskBasicPassword),
		UserAgent: freshdeskUserAgent,
	}, nil
}

func freshdeskSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// freshdeskBaseURL resolves and validates the base URL. Precedence:
//  1. an explicit base_url override (used by tests and proxies); or
//  2. the domain config (e.g. acme.freshdesk.com) turned into
//     https://<domain>/api/v2.
//
// Any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func freshdeskBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("freshdesk config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("freshdesk config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("freshdesk config base_url must include a host")
		}
		return strings.TrimRight(base, "/") + "/api/v2", nil
	}

	domain := strings.TrimSpace(cfg.Config["domain"])
	if domain == "" {
		return "", errors.New("freshdesk connector requires config domain (e.g. acme.freshdesk.com) or base_url")
	}
	// Strip any scheme the user may have included, then validate the host shape.
	domain = strings.TrimPrefix(strings.TrimPrefix(domain, "https://"), "http://")
	domain = strings.TrimRight(domain, "/")
	if domain == "" || strings.ContainsAny(domain, "/?#") {
		return "", fmt.Errorf("freshdesk config domain must be a bare host, got %q", cfg.Config["domain"])
	}
	parsed, err := url.Parse("https://" + domain)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("freshdesk config domain is invalid: %q", cfg.Config["domain"])
	}
	return "https://" + parsed.Host + "/api/v2", nil
}

func freshdeskPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return freshdeskDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("freshdesk config page_size must be an integer: %w", err)
	}
	if value < 1 || value > freshdeskMaxPageSize {
		return 0, fmt.Errorf("freshdesk config page_size must be between 1 and %d", freshdeskMaxPageSize)
	}
	return value, nil
}

func freshdeskMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("freshdesk config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("freshdesk config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Freshdesk is exposed read-only (Capabilities.Write=false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
