// Package fastbill implements the native pm FastBill connector. It follows the
// declarative-HTTP shape established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + Basic auth + RecordsAt
// extraction) with FastBill-specific stream definitions and the FastBill JSON
// SERVICE envelope.
//
// FastBill exposes a single endpoint (.../api/1.0/api.php) that takes a JSON
// POST body with a SERVICE field (e.g. "customer.get") plus LIMIT/OFFSET, and
// returns records nested under RESPONSE.<COLLECTION> (e.g. RESPONSE.CUSTOMERS).
// Authentication is HTTP Basic using the account e-mail (username) and the
// API key as the password. The API is full-refresh only, so the connector is
// read-only with no incremental cursor.
package fastbill

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
	fastbillDefaultBaseURL  = "https://my.fastbill.com/api/1.0/api.php"
	fastbillDefaultPageSize = 100
	fastbillMaxPageSize     = 100
	fastbillUserAgent       = "polymetrics-go-cli"
)

// New returns the FastBill connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm FastBill connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "fastbill" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "fastbill",
		DisplayName:     "FastBill",
		IntegrationType: "api",
		Description:     "Reads FastBill customers, invoices, products, recurring invoices, and revenues through the FastBill JSON API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to FastBill. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := fastbillBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(fastbillUsername(cfg)) == "" {
		return errors.New("fastbill connector requires config username")
	}
	if strings.TrimSpace(fastbillSecret(cfg)) == "" {
		return errors.New("fastbill connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded customer.get confirms auth and connectivity without mutating.
	body := serviceBody("customer.get", 1, 0)
	if err := r.DoJSON(ctx, http.MethodPost, "", nil, body, nil); err != nil {
		return fmt.Errorf("check fastbill: %w", err)
	}
	return nil
}

// Write is unsupported: FastBill is read-only for reverse ETL purposes.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: fastbillStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	endpoint, ok := fastbillStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("fastbill stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := fastbillPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := fastbillMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives FastBill's LIMIT/OFFSET pagination. Each request POSTs a JSON
// SERVICE envelope; records come back under RESPONSE.<COLLECTION>. A page that
// returns fewer than pageSize records ends the loop. There is no body-token
// paginator in connsdk for this POST+envelope shape, so the loop lives here,
// built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	recordsPath := "RESPONSE." + endpoint.collection
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := serviceBody(endpoint.service, pageSize, offset)
		resp, err := r.Do(ctx, http.MethodPost, "", nil, body)
		if err != nil {
			return fmt.Errorf("read fastbill %s: %w", endpoint.service, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return fmt.Errorf("decode fastbill %s page: %w", endpoint.service, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise fastbill credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := strconv.Itoa(i)
		item := map[string]any{
			"CUSTOMER_ID":     id,
			"CUSTOMER_NUMBER": "K-" + id,
			"CUSTOMER_TYPE":   "business",
			"ORGANIZATION":    fmt.Sprintf("Fixture Org %d", i),
			"FIRST_NAME":      "Fixture",
			"LAST_NAME":       fmt.Sprintf("User %d", i),
			"EMAIL":           fmt.Sprintf("fixture+%d@example.com", i),
			"PHONE":           "555-0100",
			"COUNTRY_CODE":    "DE",
			"CURRENCY_CODE":   "EUR",
			"CREATED":         "2026-01-01 00:00:00",
			"INVOICE_ID":      id,
			"INVOICE_NUMBER":  "RE-" + id,
			"TYPE":            "outgoing",
			"INVOICE_DATE":    "2026-01-01",
			"DUE_DATE":        "2026-01-15",
			"TOTAL":           fmt.Sprintf("%d.00", 100*i),
			"VAT_TOTAL":       fmt.Sprintf("%d.00", 19*i),
			"SUB_TOTAL":       fmt.Sprintf("%d.00", 81*i),
			"IS_CANCELED":     "0",
			"ARTICLE_NUMBER":  "ART-" + id,
			"TITLE":           fmt.Sprintf("Fixture Product %d", i),
			"DESCRIPTION":     "Deterministic fixture product",
			"UNIT_PRICE":      fmt.Sprintf("%d.00", 10*i),
			"VAT_PERCENT":     "19.00",
			"IS_GREEDY":       "0",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Basic auth and the resolved
// base URL. The secret only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := fastbillBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username := strings.TrimSpace(fastbillUsername(cfg))
	if username == "" {
		return nil, errors.New("fastbill connector requires config username")
	}
	secret := fastbillSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("fastbill connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(username, secret),
		UserAgent: fastbillUserAgent,
	}, nil
}

// serviceBody builds the FastBill JSON request envelope.
func serviceBody(service string, limit, offset int) map[string]any {
	return map[string]any{
		"SERVICE": service,
		"LIMIT":   limit,
		"OFFSET":  offset,
	}
}

func fastbillUsername(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["username"]
}

func fastbillSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// fastbillBaseURL resolves and validates the base URL. The default is
// my.fastbill.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func fastbillBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return fastbillDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("fastbill config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("fastbill config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("fastbill config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fastbillPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return fastbillDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fastbill config page_size must be an integer: %w", err)
	}
	if value < 1 || value > fastbillMaxPageSize {
		return 0, fmt.Errorf("fastbill config page_size must be between 1 and %d", fastbillMaxPageSize)
	}
	return value, nil
}

func fastbillMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("fastbill config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("fastbill config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
