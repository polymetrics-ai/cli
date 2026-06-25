// Package cin7 implements the native pm Cin7 (Cin7 Core / DEAR Inventory)
// connector. It follows the stripe declarative-HTTP template: a thin package that
// composes the connsdk toolkit (Requester + dual API-key header auth +
// RecordsAt extraction) with Cin7-specific stream definitions, endpoints, and a
// PageIncrement pagination loop.
//
// Cin7 Core authenticates every request with two headers,
// api-auth-accountid and api-auth-applicationkey, and paginates list endpoints
// with page (1-based) + limit query parameters, returning each resource array
// under a per-stream envelope key (e.g. Products, CustomerList).
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. The connector is read-only:
// the upstream API is a full-refresh source with no safe reverse-ETL writes.
package cin7

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
	cin7DefaultBaseURL  = "https://inventory.dearsystems.com/externalapi/v2"
	cin7DefaultPageSize = 100
	cin7MaxPageSize     = 1000
	cin7UserAgent       = "polymetrics-go-cli"

	cin7AccountHeader = "api-auth-accountid"
	cin7KeyHeader     = "api-auth-applicationkey"
)

func init() {
	connectors.RegisterFactory("cin7", New)
}

// New returns the Cin7 connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Cin7 connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "cin7" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "cin7",
		DisplayName:     "Cin7",
		IntegrationType: "api",
		Description:     "Reads Cin7 Core (DEAR Inventory) products, customers, suppliers, sales, and purchases through the Cin7 Core REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Cin7. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := cin7BaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cin7AccountID(cfg)) == "" {
		return errors.New("cin7 connector requires config accountid")
	}
	if strings.TrimSpace(cin7Secret(cfg)) == "" {
		return errors.New("cin7 connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the product list confirms auth and connectivity without
	// mutating anything.
	q := url.Values{"page": []string{"1"}, "limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "product", q, nil, nil); err != nil {
		return fmt.Errorf("check cin7: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: cin7Streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "products"
	}
	endpoint, ok := cin7StreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("cin7 stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := cin7PageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := cin7MaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Cin7's 1-based page-increment pagination. List endpoints return
// {Total, Page, <RecordsPath>:[...]} and accept page + limit query params; the
// loop stops when a short page (fewer than pageSize records) is returned. The
// loop lives here, built on connsdk.Requester + connsdk.RecordsAt, because the
// per-stream records path makes a single shared Harvest awkward.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	for k, v := range endpoint.params {
		base.Set(k, v)
	}

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read cin7 %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode cin7 %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A page shorter than the requested size means we have reached the end.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise cin7 credential-free (mirrors stripe).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"ID":             fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"SaleID":         fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"SKU":            fmt.Sprintf("SKU-%d", i),
			"Name":           fmt.Sprintf("Fixture %s %d", stream, i),
			"Email":          fmt.Sprintf("fixture+%d@example.com", i),
			"Phone":          "555-0100",
			"Status":         "Active",
			"Currency":       "USD",
			"Category":       "Fixtures",
			"Brand":          "Acme",
			"Type":           "Stock",
			"UOM":            "ea",
			"PaymentTerm":    "Net 30",
			"TaxRule":        "Tax Exempt",
			"PriceTier1":     10 * i,
			"AverageCost":    5 * i,
			"OrderNumber":    fmt.Sprintf("SO-%d", i),
			"Customer":       "Fixture Customer",
			"CustomerID":     "cust_fixture_1",
			"Supplier":       "Fixture Supplier",
			"SupplierID":     "sup_fixture_1",
			"OrderStatus":    "AUTHORISED",
			"InvoiceStatus":  "PAID",
			"InvoiceAmount":  100 * i,
			"OrderDate":      "2026-01-01T00:00:00Z",
			"LastModifiedOn": "2026-01-01T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the dual Cin7 auth headers and
// the resolved base URL. The secret only ever flows into a request header; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := cin7BaseURL(cfg)
	if err != nil {
		return nil, err
	}
	account := strings.TrimSpace(cin7AccountID(cfg))
	if account == "" {
		return nil, errors.New("cin7 connector requires config accountid")
	}
	secret := strings.TrimSpace(cin7Secret(cfg))
	if secret == "" {
		return nil, errors.New("cin7 connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: cin7UserAgent,
		DefaultHeaders: map[string]string{
			cin7AccountHeader: account,
			cin7KeyHeader:     secret,
		},
	}, nil
}

func cin7AccountID(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	// Allow the secret store to also carry the account id, falling back to config.
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets["accountid"]); v != "" {
			return v
		}
	}
	return cfg.Config["accountid"]
}

func cin7Secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// cin7BaseURL resolves and validates the base URL. The default is the Cin7 Core
// external API; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func cin7BaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return cin7DefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("cin7 config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("cin7 config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("cin7 config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func cin7PageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return cin7DefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("cin7 config page_size must be an integer: %w", err)
	}
	if value < 1 || value > cin7MaxPageSize {
		return 0, fmt.Errorf("cin7 config page_size must be between 1 and %d", cin7MaxPageSize)
	}
	return value, nil
}

func cin7MaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("cin7 config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("cin7 config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
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

// firstField returns the first non-empty stringified value among keys.
func firstField(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := item[key]; ok && v != nil {
			if s, isStr := v.(string); isStr {
				if strings.TrimSpace(s) != "" {
					return v
				}
				continue
			}
			return v
		}
	}
	return nil
}

// Write satisfies the connectors.Connector interface but the Cin7 connector is
// read-only; reverse-ETL writes are not exposed.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}
