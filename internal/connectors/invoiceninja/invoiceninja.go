// Package invoiceninja implements the native pm Invoice Ninja source connector.
// It follows the declarative-HTTP template established by the stripe connector: a
// thin package composing the connsdk toolkit (Requester + APIKeyHeader auth +
// RecordsAt extraction + a small page-number pagination loop) with Invoice
// Ninja-specific stream definitions and endpoints.
//
// Invoice Ninja v5 REST API specifics (https://api-docs.invoicing.co/):
//   - base URL: https://invoicing.co/api/v1/
//   - auth: X-API-TOKEN: <api_key> request header
//   - pagination: ?page=<n>&per_page=<size>, page numbers start at 1; a page
//     shorter than per_page is the last page (PageIncrement semantics)
//   - list responses wrap records in {"data":[...], "meta":{...}}
//
// All streams are full-refresh only (the API exposes no incremental cursor), so
// this connector is read-only with no write actions.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package invoiceninja

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
	invoiceNinjaDefaultBaseURL  = "https://invoicing.co/api/v1"
	invoiceNinjaDefaultPageSize = 100
	invoiceNinjaMaxPageSize     = 1000
	invoiceNinjaUserAgent       = "polymetrics-go-cli"
	invoiceNinjaTokenHeader     = "X-API-TOKEN"
)

func init() {
	connectors.RegisterFactory("invoiceninja", New)
}

// New returns the Invoice Ninja connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Invoice Ninja source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "invoiceninja" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "invoiceninja",
		DisplayName:     "Invoice Ninja",
		IntegrationType: "api",
		Description:     "Reads Invoice Ninja clients, invoices, products, payments, and quotes through the Invoice Ninja v5 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Invoice
// Ninja. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := invoiceNinjaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(invoiceNinjaSecret(cfg)) == "" {
		return errors.New("invoiceninja connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the clients list confirms auth and connectivity without
	// mutating anything.
	query := url.Values{"per_page": []string{"1"}, "page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "clients", query, nil, nil); err != nil {
		return fmt.Errorf("check invoiceninja: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Invoice Ninja is exposed
// as a read-only source (no reverse-ETL allow-list), so writes are rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: invoiceNinjaStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "clients"
	}
	endpoint, ok := invoiceNinjaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("invoiceninja stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := invoiceNinjaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := invoiceNinjaMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Invoice Ninja's page-number pagination. Lists return
// {data:[...], meta:{...}}; pages are 1-indexed via ?page=<n>&per_page=<size>.
// A page that returns fewer than per_page records is the last page. The mapper
// is applied per record before emit.
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
			return fmt.Errorf("read invoiceninja %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode invoiceninja %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (fewer than the requested size) is the last page.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise invoiceninja credential-free (mirrors the
// stripe fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                    fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":                  fmt.Sprintf("Fixture %d", i),
			"display_name":          fmt.Sprintf("Fixture %d", i),
			"number":                fmt.Sprintf("000%d", i),
			"product_key":           fmt.Sprintf("SKU-%d", i),
			"notes":                 "fixture product",
			"client_id":             "client_fixture_1",
			"status_id":             "1",
			"balance":               float64(10 * i),
			"amount":                float64(100 * i),
			"applied":               float64(100 * i),
			"refunded":              float64(0),
			"paid_to_date":          float64(0),
			"price":                 float64(25 * i),
			"cost":                  float64(10 * i),
			"quantity":              float64(1),
			"tax_name1":             "VAT",
			"tax_rate1":             float64(20),
			"transaction_reference": fmt.Sprintf("ref_%d", i),
			"currency_id":           "1",
			"vat_number":            "",
			"phone":                 "",
			"website":               "",
			"date":                  "2026-01-01",
			"due_date":              "2026-01-31",
			"valid_until":           "2026-02-15",
			"is_deleted":            false,
			"created_at":            int64(1767225600 + i),
			"updated_at":            int64(1767225600 + i),
			"archived_at":           int64(0),
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-API-TOKEN header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := invoiceNinjaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := invoiceNinjaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("invoiceninja connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(invoiceNinjaTokenHeader, secret, ""),
		UserAgent: invoiceNinjaUserAgent,
	}, nil
}

func invoiceNinjaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// invoiceNinjaBaseURL resolves and validates the base URL. The default is
// invoicing.co; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func invoiceNinjaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return invoiceNinjaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invoiceninja config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("invoiceninja config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("invoiceninja config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func invoiceNinjaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return invoiceNinjaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invoiceninja config page_size must be an integer: %w", err)
	}
	if value < 1 || value > invoiceNinjaMaxPageSize {
		return 0, fmt.Errorf("invoiceninja config page_size must be between 1 and %d", invoiceNinjaMaxPageSize)
	}
	return value, nil
}

func invoiceNinjaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invoiceninja config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("invoiceninja config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
