// Package freshchat implements the native pm Freshchat connector. It is a
// declarative-HTTP per-system connector modelled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + page-number pagination) with Freshchat-specific stream
// definitions and endpoints.
//
// Freshchat is a customer-messaging product. Its v2 REST API is read-oriented for
// the resources we expose (agents, users, groups, channels, roles); this
// connector is read-only and sets Capabilities.Write=false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package freshchat

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
	freshchatDefaultPageSize = 50
	freshchatMaxPageSize     = 100
	freshchatUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("freshchat", New)
}

// New returns the Freshchat connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Freshchat connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "freshchat" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "freshchat",
		DisplayName:     "Freshchat",
		IntegrationType: "api",
		Description:     "Reads Freshchat agents, users, groups, channels, and roles through the Freshchat v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Freshchat. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := freshchatBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(freshchatSecret(cfg)) == "" {
		return errors.New("freshchat connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the agents list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"page": []string{"1"}, "items_per_page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "agents", q, nil, nil); err != nil {
		return fmt.Errorf("check freshchat: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: freshchatStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Freshchat stream starts
// with an empty incremental cursor (full sync).
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
		stream = "agents"
	}
	endpoint, ok := freshchatStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("freshchat stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := freshchatPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := freshchatMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Freshchat list endpoints page via ?page=N&items_per_page=M and return the
	// array under a per-resource wrapper key. PageNumberPaginator stops once a
	// short page (< pageSize) is returned.
	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "items_per_page",
		StartPage: 1,
		PageSize:  pageSize,
	}
	base := url.Values{}
	wrap := func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, endpoint.wrapper, maxPages, wrap); err != nil {
		return fmt.Errorf("read freshchat %s: %w", endpoint.resource, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise freshchat credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"email":        fmt.Sprintf("fixture+%d@example.com", i),
			"first_name":   fmt.Sprintf("Fixture%d", i),
			"last_name":    "User",
			"phone":        "",
			"name":         fmt.Sprintf("Fixture %s %d", stream, i),
			"description":  "fixture record",
			"role":         "agent",
			"routing_type": "INTELLIGENT",
			"locale":       "en",
			"enabled":      true,
			"public":       true,
			"created_time": "2026-01-01T00:00:00.000Z",
			"updated_time": fmt.Sprintf("2026-01-0%dT00:00:00.000Z", i),
		}
		record := endpoint.mapRecord(item)
		record["fixture"] = true
		record["connector"] = "freshchat"
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
	base, err := freshchatBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := freshchatSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("freshchat connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: freshchatUserAgent,
	}, nil
}

func freshchatSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// freshchatBaseURL resolves and validates the base URL. With no explicit
// base_url override it is derived from account_name as
// https://<account_name>.freshchat.com/v2. Any override must be an absolute
// https (or http for local test servers) URL with a host to bound SSRF risk.
func freshchatBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		account := strings.TrimSpace(cfg.Config["account_name"])
		if account == "" {
			return "", errors.New("freshchat connector requires config account_name or base_url")
		}
		if strings.ContainsAny(account, "/:. ") {
			return "", fmt.Errorf("freshchat config account_name %q must be a bare subdomain", account)
		}
		return fmt.Sprintf("https://%s.freshchat.com/v2", account), nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("freshchat config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("freshchat config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("freshchat config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func freshchatPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return freshchatDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("freshchat config page_size must be an integer: %w", err)
	}
	if value < 1 || value > freshchatMaxPageSize {
		return 0, fmt.Errorf("freshchat config page_size must be between 1 and %d", freshchatMaxPageSize)
	}
	return value, nil
}

func freshchatMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("freshchat config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("freshchat config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is not supported: Freshchat is exposed read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
