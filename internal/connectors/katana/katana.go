// Package katana implements the native pm Katana (Katana MRP / Cloud Inventory)
// connector. It is a declarative-HTTP per-system connector modeled on the stripe
// reference package: a thin package that composes the connsdk toolkit (Requester
// + Bearer auth + RecordsAt extraction + page-number pagination) with
// Katana-specific stream definitions and endpoints.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
//
// Katana is a read-only source (the Airbyte source supports full_refresh only),
// so Capabilities.Write is false and there is no write.go.
package katana

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
	katanaDefaultBaseURL  = "https://api.katanamrp.com/v1"
	katanaDefaultPageSize = 50
	katanaMaxPageSize     = 250
	katanaUserAgent       = "polymetrics-go-cli"
	// katanaFixtureUpdated is the deterministic updated_at timestamp used by the
	// fixture-mode records.
	katanaFixtureUpdated = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("katana", New)
}

// New returns the Katana connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Katana connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "katana" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "katana",
		DisplayName:     "Katana",
		IntegrationType: "api",
		Description:     "Reads Katana MRP (Cloud Inventory) products, materials, variants, sales orders, and customers through the Katana REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Katana. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := katanaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(katanaSecret(cfg)) == "" {
		return errors.New("katana connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the products list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "products", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check katana: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: katanaStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Katana stream starts with
// an empty incremental cursor (full sync).
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
	endpoint, ok := katanaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("katana stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := katanaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := katanaMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Katana lists return {"data":[...]} and paginate by page number with a
	// limit; the last page is a short (or empty) page. PageNumberPaginator stops
	// once a page returns fewer than PageSize records, which matches that shape.
	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "limit",
		StartPage: 1,
		PageSize:  pageSize,
	}
	mapped := func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, url.Values{}, paginator, "data", maxPages, mapped); err != nil {
		return fmt.Errorf("read katana %s: %w", endpoint.resource, err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Katana is a read-only
// source (the upstream Airbyte source supports full_refresh only), so writes are
// unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise katana credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                     i,
			"name":                   fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"uom":                    "pcs",
			"category_name":          "Fixtures",
			"is_sellable":            true,
			"is_producible":          true,
			"is_purchasable":         false,
			"default_supplier_id":    nil,
			"additional_info":        "fixture",
			"sku":                    fmt.Sprintf("SKU-%d", i),
			"product_id":             i,
			"material_id":            nil,
			"sales_price":            10.0 * float64(i),
			"purchase_price":         5.0 * float64(i),
			"type":                   "product",
			"order_no":               fmt.Sprintf("SO-%d", i),
			"customer_id":            i,
			"status":                 "NOT_SHIPPED",
			"currency":               "USD",
			"total":                  100.0 * float64(i),
			"total_in_base_currency": 100.0 * float64(i),
			"order_created_date":     katanaFixtureUpdated,
			"delivery_date":          katanaFixtureUpdated,
			"email":                  fmt.Sprintf("fixture+%d@example.com", i),
			"phone":                  "+1-555-0100",
			"reference_id":           fmt.Sprintf("ref-%d", i),
			"category":               "wholesale",
			"created_at":             katanaFixtureUpdated,
			"updated_at":             katanaFixtureUpdated,
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
	base, err := katanaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := katanaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("katana connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: katanaUserAgent,
	}, nil
}

func katanaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// katanaBaseURL resolves and validates the base URL. The default is
// api.katanamrp.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func katanaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return katanaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("katana config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("katana config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("katana config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func katanaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return katanaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("katana config page_size must be an integer: %w", err)
	}
	if value < 1 || value > katanaMaxPageSize {
		return 0, fmt.Errorf("katana config page_size must be between 1 and %d", katanaMaxPageSize)
	}
	return value, nil
}

func katanaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("katana config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("katana config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
