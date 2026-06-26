// Package lightspeedretail implements the native pm Lightspeed Retail
// (X-Series) connector. It is a declarative-HTTP per-system connector following
// the stripe/acuity-scheduling template: a thin package that composes the
// connsdk toolkit (Requester + Bearer auth + RecordsAt extraction) with
// Lightspeed-specific stream definitions, endpoints, and version-cursor
// pagination.
//
// The Lightspeed Retail X-Series API is hosted per retailer at
// https://<subdomain>.retail.lightspeed.app and authenticates with a personal
// token / OAuth access token via Authorization: Bearer <api_key>. The 2.0 list
// endpoints (api/2.0/products, api/2.0/customers, ...) return a JSON object with
// a "data" array and a "version" object; pages are walked by passing
// after=<version.max> until version.max is empty/null. The Lightspeed source is
// read-only (full-refresh); it exposes no reverse-ETL writes, so
// Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package lightspeedretail

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
	registryName             = "lightspeed-retail"
	lightspeedHostSuffix     = ".retail.lightspeed.app"
	lightspeedDefaultPageSiz = 100
	lightspeedMaxPageSize    = 200
	lightspeedUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Lightspeed Retail connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Lightspeed Retail connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Lightspeed Retail",
		IntegrationType: "api",
		Description:     "Reads Lightspeed Retail (X-Series) products, customers, sales, outlets, and registers through the Lightspeed REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Lightspeed.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := lightspeedBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(lightspeedSecret(cfg)) == "" {
		return errors.New("lightspeed-retail connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the outlets list confirms auth and connectivity without
	// mutating anything (outlets is a small, always-present resource).
	q := url.Values{"page_size": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "api/2.0/outlets", q, nil, nil); err != nil {
		return fmt.Errorf("check lightspeed-retail: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: lightspeedStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Lightspeed stream starts
// with an empty version cursor (full sync); the read path raises it as pages are
// walked.
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
		stream = "products"
	}
	endpoint, ok := lightspeedStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("lightspeed-retail stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := lightspeedPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := lightspeedMaxPages(req.Config)
	if err != nil {
		return err
	}
	after := connsdk.Cursor(req.State)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, after, emit)
}

// harvest drives Lightspeed's version-cursor pagination. Each 2.0 list response
// is {data:[...], version:{min, max}}; the next page is requested with
// after=<version.max> until version.max is empty/null. This exact shape (cursor
// lives in the body under version.max, parameter is "after") has no off-the-shelf
// connsdk paginator, so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, after string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("page_size", strconv.Itoa(pageSize))

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read lightspeed-retail %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode lightspeed-retail %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		nextMax, err := connsdk.StringAt(resp.Body, "version.max")
		if err != nil {
			return fmt.Errorf("decode lightspeed-retail %s version.max: %w", endpoint.resource, err)
		}
		nextMax = strings.TrimSpace(nextMax)
		// An empty/null/zero version.max means there are no further pages.
		if nextMax == "" || nextMax == "0" || len(records) == 0 {
			return nil
		}
		// Guard against a server that keeps returning the same cursor.
		if nextMax == after {
			return nil
		}
		after = nextMax
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise lightspeed-retail credential-free (mirrors
// stripe's fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  fmt.Sprintf("%s_fixture_%d", stream, i),
			"version":             int64(1000 + i),
			"name":                fmt.Sprintf("Fixture %d", i),
			"handle":              fmt.Sprintf("fixture-%d", i),
			"sku":                 fmt.Sprintf("SKU-%d", i),
			"description":         "Fixture record",
			"brand_id":            "brand_fixture_1",
			"supplier_id":         "supplier_fixture_1",
			"product_category":    "General",
			"price_including_tax": float64(10 * i),
			"price_excluding_tax": float64(9 * i),
			"supply_price":        float64(5 * i),
			"has_variants":        false,
			"is_active":           true,
			"is_composite":        false,
			"customer_code":       fmt.Sprintf("C-%d", i),
			"customer_group_id":   "group_fixture_1",
			"balance":             float64(0),
			"loyalty_balance":     float64(0),
			"year_to_date":        float64(100 * i),
			"do_not_email":        false,
			"enable_loyalty":      true,
			"invoice_number":      fmt.Sprintf("INV-%d", i),
			"customer_id":         "customer_fixture_1",
			"status":              "CLOSED",
			"register_id":         "register_fixture_1",
			"user_id":             "user_fixture_1",
			"total_price":         float64(10 * i),
			"total_tax":           float64(1 * i),
			"sale_date":           "2026-01-01T00:00:00Z",
			"currency":            "USD",
			"currency_symbol":     "$",
			"default_tax_id":      "tax_fixture_1",
			"display_prices":      "inclusive",
			"time_zone":           "America/New_York",
			"outlet_id":           "outlet_fixture_1",
			"invoice_prefix":      "INV",
			"invoice_sequence":    int64(i),
			"is_open":             true,
			"email_receipt":       true,
			"print_receipt":       false,
			"created_at":          "2026-01-01T00:00:00Z",
			"updated_at":          "2026-01-02T00:00:00Z",
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := lightspeedBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := lightspeedSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("lightspeed-retail connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: lightspeedUserAgent,
	}, nil
}

func lightspeedSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// lightspeedBaseURL resolves and validates the base URL. An explicit base_url
// override wins (validated to bound SSRF). Otherwise the host is derived from the
// required subdomain config as https://<subdomain>.retail.lightspeed.app.
func lightspeedBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("lightspeed-retail config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("lightspeed-retail config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("lightspeed-retail config base_url must include a host")
		}
		return strings.TrimRight(base, "/"), nil
	}

	subdomain := strings.TrimSpace(cfg.Config["subdomain"])
	if subdomain == "" {
		return "", errors.New("lightspeed-retail connector requires config subdomain (or base_url)")
	}
	if !validSubdomain(subdomain) {
		return "", fmt.Errorf("lightspeed-retail config subdomain %q is invalid", subdomain)
	}
	return "https://" + subdomain + lightspeedHostSuffix, nil
}

// validSubdomain ensures the subdomain is a single DNS label (letters, digits,
// hyphens) so it cannot be used to inject a different host into the base URL.
func validSubdomain(sub string) bool {
	if len(sub) == 0 || len(sub) > 63 {
		return false
	}
	for _, r := range sub {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return sub[0] != '-' && sub[len(sub)-1] != '-'
}

func lightspeedPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return lightspeedDefaultPageSiz, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("lightspeed-retail config page_size must be an integer: %w", err)
	}
	if value < 1 || value > lightspeedMaxPageSize {
		return 0, fmt.Errorf("lightspeed-retail config page_size must be between 1 and %d", lightspeedMaxPageSize)
	}
	return value, nil
}

func lightspeedMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("lightspeed-retail config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("lightspeed-retail config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write satisfies the connectors.Connector interface. The Lightspeed source is
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

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}
