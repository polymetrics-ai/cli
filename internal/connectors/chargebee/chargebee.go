// Package chargebee implements the native pm Chargebee source connector. It is a
// declarative-HTTP per-system connector built on the stripe template: a thin
// package that composes the connsdk toolkit (Requester + HTTP Basic auth +
// RecordsAt extraction + cursor state) with Chargebee-specific stream
// definitions, endpoints, and pagination.
//
// Chargebee's REST API is rooted at https://{site}.chargebee.com/api/v2,
// authenticates with HTTP Basic (the site API key as the username and an empty
// password), and paginates list endpoints with limit + offset where each
// response carries a top-level "list" array and an opaque "next_offset" token.
// Each list element wraps its resource in a single key (e.g. {"customer": {...}}).
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect. This connector is read-only.
package chargebee

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	chargebeeDefaultPageSize = 100
	chargebeeMaxPageSize     = 100
	chargebeeUserAgent       = "polymetrics-go-cli"
	// chargebeeFixtureCreated is the deterministic created_at/updated_at timestamp
	// used by fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	chargebeeFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("chargebee", New)
}

// New returns the Chargebee connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Chargebee source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "chargebee" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "chargebee",
		DisplayName:     "Chargebee",
		IntegrationType: "api",
		Description:     "Reads Chargebee customers, subscriptions, invoices, plans, and items through the Chargebee v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Chargebee. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := chargebeeBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(chargebeeSecret(cfg)) == "" {
		return errors.New("chargebee connector requires secret site_api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the customers list confirms auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "customers", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check chargebee: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: chargebeeStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Chargebee stream starts
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
		stream = "customers"
	}
	endpoint, ok := chargebeeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("chargebee stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := chargebeePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := chargebeeMaxPages(req.Config)
	if err != nil {
		return err
	}
	lower := incrementalLowerBound(req)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives Chargebee's offset/next_offset pagination. List responses are
// shaped {"list":[{<envelope>:{...}}, ...], "next_offset":"<token>"}; the next
// page is requested with offset=<next_offset>. There is no body-token paginator
// in connsdk for this exact (per-item envelope) shape, so the loop lives here,
// built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, updatedAfter string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if updatedAfter != "" {
		// Chargebee filters incremental reads with updated_at[after]=<unix>.
		base.Set("updated_at[after]", updatedAfter)
		base.Set("sort_by[asc]", "updated_at")
	}

	offset := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if offset != "" {
			query.Set("offset", offset)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read chargebee %s: %w", endpoint.resource, err)
		}
		items, err := connsdk.RecordsAt(resp.Body, "list")
		if err != nil {
			return fmt.Errorf("decode chargebee %s page: %w", endpoint.resource, err)
		}
		for _, wrapper := range items {
			if err := ctx.Err(); err != nil {
				return err
			}
			obj := unwrap(wrapper, endpoint.envelope)
			if obj == nil {
				continue
			}
			if err := emit(endpoint.mapRecord(obj)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_offset")
		if err != nil {
			return fmt.Errorf("decode chargebee %s next_offset: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		offset = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise chargebee credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                   fmt.Sprintf("%s_fixture_%d", endpoint.envelope, i),
			"first_name":           fmt.Sprintf("Fixture%d", i),
			"last_name":            "User",
			"email":                fmt.Sprintf("fixture+%d@example.com", i),
			"company":              "Example Inc",
			"phone":                "+15555550100",
			"auto_collection":      "on",
			"net_term_days":        int64(0),
			"taxability":           "taxable",
			"customer_id":          "customer_fixture_1",
			"subscription_id":      "subscription_fixture_1",
			"plan_id":              "plan_fixture_1",
			"status":               "active",
			"currency_code":        "USD",
			"plan_quantity":        int64(1),
			"plan_amount":          int64(1000 * i),
			"current_term_start":   chargebeeFixtureCreated,
			"current_term_end":     chargebeeFixtureCreated + 2592000,
			"started_at":           chargebeeFixtureCreated,
			"total":                int64(1000 * i),
			"amount_paid":          int64(1000 * i),
			"amount_due":           int64(0),
			"date":                 chargebeeFixtureCreated,
			"due_date":             chargebeeFixtureCreated + 86400,
			"paid_at":              chargebeeFixtureCreated + 3600,
			"name":                 fmt.Sprintf("Fixture %d", i),
			"invoice_name":         fmt.Sprintf("Fixture Plan %d", i),
			"price":                int64(1000 * i),
			"period":               int64(1),
			"period_unit":          "month",
			"type":                 "plan",
			"item_family_id":       "family_fixture_1",
			"is_shippable":         false,
			"enabled_for_checkout": true,
			"created_at":           chargebeeFixtureCreated + int64(i),
			"updated_at":           chargebeeFixtureCreated + int64(i),
			"deleted":              false,
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

// Write satisfies the connectors.Connector interface. Chargebee is exposed as a
// read-only source connector, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with HTTP Basic auth and the
// resolved base URL. Chargebee uses the site API key as the Basic username with
// an empty password; the secret only ever flows into connsdk.Basic and is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := chargebeeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := chargebeeSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("chargebee connector requires secret site_api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: chargebeeUserAgent,
	}, nil
}

// incrementalLowerBound returns the unix-seconds lower bound for
// updated_at[after], derived from the incremental cursor (if any) or else the
// start_date config. An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return ""
	}
	// Chargebee timestamps are unix seconds. Accept either a raw unix value or an
	// RFC3339 timestamp for start_date.
	if _, err := strconv.ParseInt(startDate, 10, 64); err == nil {
		return startDate
	}
	if t, err := parseRFC3339(startDate); err == nil {
		return strconv.FormatInt(t, 10)
	}
	return ""
}

// parseRFC3339 parses an RFC3339 timestamp and returns it as unix seconds.
func parseRFC3339(value string) (int64, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func chargebeeSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["site_api_key"]
}

// chargebeeBaseURL resolves and validates the base URL. A base_url override (used
// in tests) wins; otherwise it is derived from the required site config field as
// https://{site}.chargebee.com/api/v2. Any override must be an absolute https (or
// http for local test servers) URL with a host to bound SSRF risk.
func chargebeeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		site := strings.TrimSpace(cfg.Config["site"])
		if site == "" {
			return "", errors.New("chargebee connector requires config site (or base_url)")
		}
		if !validSite(site) {
			return "", fmt.Errorf("chargebee config site %q is invalid", site)
		}
		return fmt.Sprintf("https://%s.chargebee.com/api/v2", site), nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("chargebee config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("chargebee config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("chargebee config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// validSite restricts the site prefix to the characters Chargebee allows in a
// site name (letters, digits, hyphen), guarding the derived host against
// injection.
func validSite(site string) bool {
	if len(site) > 64 {
		return false
	}
	for _, r := range site {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func chargebeePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return chargebeeDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("chargebee config page_size must be an integer: %w", err)
	}
	if value < 1 || value > chargebeeMaxPageSize {
		return 0, fmt.Errorf("chargebee config page_size must be between 1 and %d", chargebeeMaxPageSize)
	}
	return value, nil
}

func chargebeeMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("chargebee config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("chargebee config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// unwrap extracts the resource object from a Chargebee list element. Each element
// is shaped {<envelope>: {...}}; if the wrapper is absent the element is treated
// as the object itself (defensive fallback).
func unwrap(wrapper map[string]any, envelope string) map[string]any {
	if wrapper == nil {
		return nil
	}
	if inner, ok := wrapper[envelope]; ok {
		if obj, ok := inner.(map[string]any); ok {
			return obj
		}
		return nil
	}
	return wrapper
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
