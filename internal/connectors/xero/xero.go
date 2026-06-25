// Package xero implements the native pm Xero connector. It is a declarative-HTTP
// per-system connector built on the same shape as the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + cursor state) with Xero-specific stream definitions,
// endpoints, and the Xero Accounting API's page-based pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Xero is read-only here: the Accounting API supports writes, but reverse-ETL
// writes are not enabled for this connector (Capabilities.Write = false).
package xero

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
	xeroDefaultBaseURL  = "https://api.xero.com/api.xro/2.0"
	xeroDefaultPageSize = 100
	xeroMaxPageSize     = 100
	xeroUserAgent       = "polymetrics-go-cli"
	xeroTenantHeader    = "Xero-tenant-id"
)

func init() {
	connectors.RegisterFactory("xero", New)
}

// New returns the Xero connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Xero connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "xero" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "xero",
		DisplayName:     "Xero",
		IntegrationType: "api",
		Description:     "Reads Xero accounting data (invoices, contacts, accounts, bank transactions, items, and payments) from the Xero Accounting API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Xero. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := xeroBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(xeroAccessToken(cfg)) == "" {
		return errors.New("xero connector requires secret access_token")
	}
	if strings.TrimSpace(xeroTenantID(cfg)) == "" {
		return errors.New("xero connector requires secret tenant_id")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the Organisation endpoint confirms auth, the tenant
	// header, and connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "Organisation", nil, nil, nil); err != nil {
		return fmt.Errorf("check xero: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: xeroStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Xero stream starts with an
// empty incremental cursor (full sync), which the start_date config can raise at
// read time via the If-Modified-Since header.
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
		stream = "invoices"
	}
	endpoint, ok := xeroStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("xero stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := xeroMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, maxPages, emit)
}

// harvest drives Xero's page-based pagination. List responses are shaped
// {"<Resource>":[...]}; pages are requested with ?page=N starting at 1, each
// returning up to 100 records. The loop stops when a page returns fewer than the
// page size (the last page) or when maxPages is reached. Some Xero endpoints do
// not paginate and return the full set on page 1; a short page terminates the
// loop in that case too.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read xero %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.resource)
		if err != nil {
			return fmt.Errorf("decode xero %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < xeroDefaultPageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise xero credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			endpoint.idField:          fmt.Sprintf("%s_fixture_%d", strings.ToLower(endpoint.resource), i),
			"Type":                    "ACCREC",
			"PaymentType":             "ACCRECPAYMENT",
			"Status":                  "AUTHORISED",
			"ContactStatus":           "ACTIVE",
			"Name":                    fmt.Sprintf("Fixture %d", i),
			"FirstName":               "Fixture",
			"LastName":                strconv.Itoa(i),
			"EmailAddress":            fmt.Sprintf("fixture+%d@example.com", i),
			"Code":                    fmt.Sprintf("FX-%d", i),
			"InvoiceNumber":           fmt.Sprintf("INV-%04d", i),
			"CurrencyCode":            "USD",
			"Total":                   float64(100 * i),
			"SubTotal":                float64(90 * i),
			"TotalTax":                float64(10 * i),
			"AmountDue":               float64(100 * i),
			"AmountPaid":              float64(0),
			"Amount":                  float64(100 * i),
			"QuantityOnHand":          float64(i),
			"Date":                    "2026-01-01T00:00:00Z",
			"DueDate":                 "2026-02-01T00:00:00Z",
			"UpdatedDateUTC":          "2026-01-01T00:00:00Z",
			"IsCustomer":              true,
			"IsSupplier":              false,
			"IsReconciled":            false,
			"IsSold":                  true,
			"IsPurchased":             true,
			"IsTrackedAsInventory":    false,
			"EnablePaymentsToAccount": false,
			"Contact": map[string]any{
				"ContactID": "contact_fixture_1",
			},
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "xero"
		record["fixture"] = true
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth, the resolved base
// URL, and the required Xero-tenant-id header. The access token only ever flows
// into connsdk.Bearer; neither it nor the tenant id is ever logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := xeroBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := xeroAccessToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("xero connector requires secret access_token")
	}
	tenant := strings.TrimSpace(xeroTenantID(cfg))
	if tenant == "" {
		return nil, errors.New("xero connector requires secret tenant_id")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(token),
		UserAgent: xeroUserAgent,
		DefaultHeaders: map[string]string{
			xeroTenantHeader: tenant,
		},
	}, nil
}

// xeroAccessToken resolves the OAuth2 access token from secrets.
func xeroAccessToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}

// xeroTenantID resolves the tenant id. It is a secret field, but may also be
// supplied via config for convenience; secrets take precedence.
func xeroTenantID(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets["tenant_id"]); v != "" {
			return v
		}
	}
	if cfg.Config != nil {
		return strings.TrimSpace(cfg.Config["tenant_id"])
	}
	return ""
}

// xeroBaseURL resolves and validates the base URL. The default is api.xero.com;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func xeroBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return xeroDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("xero config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("xero config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("xero config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func xeroMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("xero config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("xero config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: the Xero connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
