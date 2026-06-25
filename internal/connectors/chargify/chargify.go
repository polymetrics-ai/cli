// Package chargify implements the native pm Chargify (Maxio Advanced Billing)
// connector. It is a declarative-HTTP per-system connector built on the connsdk
// toolkit (Requester + Basic auth + a small page/per_page pagination loop) with
// Chargify-specific stream definitions, endpoints, and record mappers. It mirrors
// the stripe connector's shape.
//
// Chargify list endpoints (e.g. GET /customers.json) return a top-level JSON
// array whose elements each wrap the resource under a singular key, for example
// [{"customer": {...}}, {"customer": {...}}]. Pagination is page-number based
// via page (1-based) and per_page (max 200); a short page ends the read.
//
// Auth is HTTP Basic: by default the API key is the username and the literal
// string "x" is the password, matching Chargify's documented scheme. An explicit
// username (config) + password (secret) pair overrides that default.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package chargify

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
	chargifyDefaultPageSize = 100
	chargifyMaxPageSize     = 200
	chargifyUserAgent       = "polymetrics-go-cli"
	// chargifyBasicPassword is the conventional placeholder password Chargify
	// expects when authenticating with an API key as the Basic-auth username.
	chargifyBasicPassword = "x"
)

func init() {
	connectors.RegisterFactory("chargify", New)
}

// New returns the Chargify connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Chargify connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "chargify" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "chargify",
		DisplayName:     "Chargify",
		IntegrationType: "api",
		Description:     "Reads Chargify (Maxio Advanced Billing) customers, subscriptions, products, coupons, and transactions through the Chargify REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Chargify. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := chargifyBaseURL(cfg); err != nil {
		return err
	}
	if _, _, err := chargifyCredentials(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the customers list confirms auth and connectivity
	// without mutating anything.
	q := url.Values{"page": []string{"1"}, "per_page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "customers.json", q, nil, nil); err != nil {
		return fmt.Errorf("check chargify: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: chargifyStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Chargify stream starts with
// an empty incremental cursor (full sync). Chargify list endpoints do not accept
// an updated_at lower bound across all streams, so the cursor is tracked for
// resumability but the read remains a full refresh.
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
		stream = "customers"
	}
	endpoint, ok := chargifyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("chargify stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := chargifyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := chargifyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Chargify's page-number pagination. Each list response is a
// top-level array of single-key wrapper objects; a page shorter than pageSize
// (or empty) ends the read. The loop lives here because the records live at the
// JSON root and must be unwrapped by endpoint.wrapKey before mapping, which the
// generic connsdk paginators do not do.
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
			return fmt.Errorf("read chargify %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode chargify %s page: %w", endpoint.resource, err)
		}
		for _, wrapped := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			item := unwrap(wrapped, endpoint.wrapKey)
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// unwrap extracts the nested resource object from a Chargify list element. If the
// element is wrapped under wrapKey (e.g. {"customer": {...}}) the inner object is
// returned; otherwise the element is returned as-is so the connector tolerates
// unwrapped payloads too.
func unwrap(wrapped map[string]any, wrapKey string) map[string]any {
	if inner, ok := wrapped[wrapKey].(map[string]any); ok {
		return inner
	}
	return wrapped
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise chargify credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                        i,
			"first_name":                fmt.Sprintf("Fixture%d", i),
			"last_name":                 "Customer",
			"email":                     fmt.Sprintf("fixture+%d@example.com", i),
			"organization":              "Polymetrics",
			"reference":                 fmt.Sprintf("ref_%d", i),
			"phone":                     "555-0100",
			"country":                   "US",
			"name":                      fmt.Sprintf("%s Fixture %d", strings.TrimSuffix(stream, "s"), i),
			"handle":                    fmt.Sprintf("%s-fixture-%d", endpoint.wrapKey, i),
			"description":               "deterministic fixture record",
			"state":                     "active",
			"customer_id":               1,
			"product_id":                1,
			"product_family_id":         1,
			"subscription_id":           1,
			"balance_in_cents":          1000 * i,
			"total_revenue_in_cents":    1000 * i,
			"price_in_cents":            1000 * i,
			"amount_in_cents":           1000 * i,
			"percentage":                "10.0",
			"interval":                  1,
			"interval_unit":             "month",
			"code":                      fmt.Sprintf("COUPON%d", i),
			"transaction_type":          "payment",
			"kind":                      "charge",
			"success":                   true,
			"current_period_started_at": "2026-01-01T00:00:00Z",
			"current_period_ends_at":    "2026-02-01T00:00:00Z",
			"created_at":                "2026-01-01T00:00:00Z",
			"updated_at":                fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
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

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The credentials only ever flow into connsdk.Basic; they are never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := chargifyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username, password, err := chargifyCredentials(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, password),
		UserAgent: chargifyUserAgent,
	}, nil
}

// chargifyCredentials resolves the Basic-auth username/password pair. Preference
// order: an explicit username (config) + password (secret); otherwise the API key
// (secret) as the username with the literal "x" as the password.
func chargifyCredentials(cfg connectors.RuntimeConfig) (string, string, error) {
	username := strings.TrimSpace(cfg.Config["username"])
	password := secret(cfg, "password")
	if username != "" && password != "" {
		return username, password, nil
	}
	apiKey := secret(cfg, "api_key")
	if apiKey != "" {
		return apiKey, chargifyBasicPassword, nil
	}
	return "", "", errors.New("chargify connector requires secret api_key (or username + secret password)")
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Secrets[key])
}

// chargifyBaseURL resolves and validates the base URL. A base_url override takes
// precedence; otherwise it is derived from the domain/subdomain config. Any URL
// must be an absolute https (or http for local test servers) URL with a host to
// bound SSRF risk.
func chargifyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		return validateBaseURL(base)
	}
	domain := strings.TrimSpace(cfg.Config["domain"])
	if domain == "" {
		domain = strings.TrimSpace(cfg.Config["subdomain"])
		if domain != "" {
			domain = domain + ".chargify.com"
		}
	}
	if domain == "" {
		return "", errors.New("chargify config requires domain (or subdomain) or base_url")
	}
	if !strings.Contains(domain, "://") {
		domain = "https://" + domain
	}
	return validateBaseURL(domain)
}

func validateBaseURL(base string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("chargify config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("chargify config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("chargify config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func chargifyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return chargifyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("chargify config page_size must be an integer: %w", err)
	}
	if value < 1 || value > chargifyMaxPageSize {
		return 0, fmt.Errorf("chargify config page_size must be between 1 and %d", chargifyMaxPageSize)
	}
	return value, nil
}

func chargifyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("chargify config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("chargify config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: Chargify is exposed read-only (the upstream connector is
// a full-refresh source). The method satisfies the Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
