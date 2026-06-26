// Package onepagecrm implements the native pm OnePageCRM connector. It is a
// declarative-HTTP per-system connector following the stripe reference shape: a
// thin package that composes the connsdk toolkit (Requester + Basic auth +
// RecordsAt extraction) with OnePageCRM-specific stream definitions, endpoints,
// and its nested record/pagination conventions.
//
// OnePageCRM's API lives at https://app.onepagecrm.com/api/v3/, authenticates
// with HTTP Basic auth (username = API user ID, password = API key), paginates
// with page/per_page and a max_page field in the response body, and wraps each
// list element under a singular key (e.g. {"contact": {...}}).
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package onepagecrm

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
	onepagecrmDefaultBaseURL  = "https://app.onepagecrm.com/api/v3"
	onepagecrmDefaultPageSize = 100
	onepagecrmMaxPageSize     = 100
	onepagecrmUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("onepagecrm", New)
}

// New returns the OnePageCRM connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm OnePageCRM connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "onepagecrm" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "onepagecrm",
		DisplayName:     "OnePageCRM",
		IntegrationType: "api",
		Description:     "Reads OnePageCRM contacts, deals, actions, companies, and users through the OnePageCRM REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to OnePageCRM.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := onepagecrmBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(onepagecrmUsername(cfg)) == "" {
		return errors.New("onepagecrm connector requires config username")
	}
	if strings.TrimSpace(onepagecrmPassword(cfg)) == "" {
		return errors.New("onepagecrm connector requires secret password")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "users", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check onepagecrm: %w", err)
	}
	return nil
}

// Write is unsupported: OnePageCRM is exposed as a read-only source connector.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: onepagecrmStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := onepagecrmStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("onepagecrm stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := onepagecrmPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := onepagecrmMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives OnePageCRM's page/per_page pagination. List responses include a
// data.page and data.max_page; the loop requests successive pages until page
// reaches max_page (or a short/empty page is returned). Each array element is
// unwrapped from its singular wrap key before mapping.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read onepagecrm %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.arrayPath)
		if err != nil {
			return fmt.Errorf("decode onepagecrm %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(unwrap(item, endpoint.wrapKey))); err != nil {
				return err
			}
		}

		// max_page from the body is authoritative when present: stop once the
		// current page reaches it. Otherwise fall back to a short-page heuristic
		// (fewer than a full page implies the last page).
		if maxPage, err := connsdk.StringAt(resp.Body, "data.max_page"); err == nil && maxPage != "" {
			mp, perr := strconv.Atoi(maxPage)
			if perr == nil {
				if page >= mp {
					return nil
				}
				continue
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise onepagecrm credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"first_name":          fmt.Sprintf("Fixture%d", i),
			"last_name":           "Example",
			"company_name":        "Fixture Co",
			"name":                fmt.Sprintf("%s fixture %d", stream, i),
			"job_title":           "Engineer",
			"status_id":           "status_1",
			"owner_id":            "user_1",
			"status":              "open",
			"stage":               "Lead",
			"amount":              float64(1000 * i),
			"currency":            "USD",
			"contact_id":          "contacts_fixture_1",
			"expected_close_date": "2026-01-01",
			"text":                "Follow up",
			"date":                "2026-01-01",
			"done":                false,
			"assignee_id":         "user_1",
			"phone":               "+1000000000",
			"url":                 "https://example.com",
			"description":         "fixture record",
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"role":                "member",
			"starred":             false,
			"created_at":          "2026-01-01T00:00:00Z",
			"updated_at":          "2026-01-01T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// unwrap pulls the object out of its singular wrap key. OnePageCRM list elements
// look like {"contact": {...}}; wrapKey is "contact". When the element is not
// wrapped (or wrapKey is empty) the element itself is returned.
func unwrap(item map[string]any, wrapKey string) map[string]any {
	if wrapKey == "" {
		return item
	}
	if inner, ok := item[wrapKey].(map[string]any); ok {
		return inner
	}
	return item
}

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The password secret only ever flows into connsdk.Basic; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := onepagecrmBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(onepagecrmUsername(cfg))
	if username == "" {
		return nil, errors.New("onepagecrm connector requires config username")
	}
	password := onepagecrmPassword(cfg)
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("onepagecrm connector requires secret password")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, password),
		UserAgent: onepagecrmUserAgent,
	}, nil
}

func onepagecrmUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func onepagecrmPassword(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["password"]
}

// onepagecrmBaseURL resolves and validates the base URL. The default is
// app.onepagecrm.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func onepagecrmBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return onepagecrmDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("onepagecrm config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("onepagecrm config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("onepagecrm config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func onepagecrmPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return onepagecrmDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("onepagecrm config page_size must be an integer: %w", err)
	}
	if value < 1 || value > onepagecrmMaxPageSize {
		return 0, fmt.Errorf("onepagecrm config page_size must be between 1 and %d", onepagecrmMaxPageSize)
	}
	return value, nil
}

func onepagecrmMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("onepagecrm config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("onepagecrm config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
