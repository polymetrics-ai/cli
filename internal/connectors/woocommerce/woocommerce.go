// Package woocommerce implements the native pm WooCommerce connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit (Requester +
// HTTP Basic auth + top-level-array extraction + page-number pagination), modeled
// on the stripe reference connector.
//
// WooCommerce exposes the WordPress-hosted WooCommerce REST API (wc/v3) at
// https://<shop>/wp-json/wc/v3. Authentication over HTTPS is HTTP Basic with the
// consumer key as username and consumer secret as password. List endpoints return
// a top-level JSON array and paginate with page/per_page, reporting the total page
// count in the X-WP-TotalPages response header.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package woocommerce

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
	woocommerceAPIPath      = "/wp-json/wc/v3"
	woocommerceDefaultPage  = 10
	woocommerceMaxPageSize  = 100
	woocommerceUserAgent    = "polymetrics-go-cli"
	woocommerceFixtureMonth = "2026-01-0" // suffixed with the 1-based fixture index
)

func init() {
	connectors.RegisterFactory("woocommerce", New)
}

// New returns the WooCommerce connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm WooCommerce connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "woocommerce" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "woocommerce",
		DisplayName:     "WooCommerce",
		IntegrationType: "api",
		Description:     "Reads WooCommerce orders, products, customers, and coupons through the WooCommerce REST API (wc/v3).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to WooCommerce.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := woocommerceBaseURL(cfg); err != nil {
		return err
	}
	key, secret := woocommerceSecrets(cfg)
	if strings.TrimSpace(key) == "" || strings.TrimSpace(secret) == "" {
		return errors.New("woocommerce connector requires secrets api_key and api_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the orders list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "orders", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check woocommerce: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: woocommerceStreams()}, nil
}

// Write is unsupported: this connector is read-only (no reverse-ETL writes). It
// satisfies the connectors.Connector interface by returning the shared
// unsupported-operation error.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a WooCommerce stream starts
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
		stream = "orders"
	}
	endpoint, ok := woocommerceStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("woocommerce stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	lower, err := incrementalLowerBound(req)
	if err != nil {
		return err
	}
	pageSize, err := woocommercePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := woocommerceMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives WooCommerce's WordPress page-number pagination. Each list
// response is a top-level JSON array and the total page count is reported in the
// X-WP-TotalPages header. The loop stops when the reported total pages is reached,
// when a short/empty page is returned, or when maxPages is hit. It is built on
// connsdk.Requester + connsdk.RecordsAt rather than a generic paginator because
// the stop condition reads a response header.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, modifiedAfter string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	base.Set("order", "asc")
	base.Set("orderby", "id")
	if modifiedAfter != "" {
		// modified_after filters by date_modified_gmt for incremental syncs;
		// after is included as a fallback for resources that ignore it.
		base.Set("modified_after", modifiedAfter)
		base.Set("after", modifiedAfter)
	}

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read woocommerce %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode woocommerce %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		totalPages := parseTotalPages(resp.Header.Get("X-WP-TotalPages"))
		// Stop when the header says we have read every page, or when this page
		// was short/empty (no header present, or fewer records than requested).
		if totalPages > 0 {
			if page >= totalPages {
				return nil
			}
			continue
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise woocommerce credential-free (mirrors the
// stripe fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		ts := woocommerceFixtureMonth + strconv.Itoa(i) + "T00:00:00"
		item := map[string]any{
			"id":                 i,
			"number":             strconv.Itoa(1000 + i),
			"status":             "processing",
			"currency":           "USD",
			"total":              fmt.Sprintf("%d.00", 10*i),
			"total_tax":          "0.00",
			"customer_id":        i,
			"payment_method":     "stripe",
			"name":               fmt.Sprintf("Fixture Product %d", i),
			"slug":               fmt.Sprintf("fixture-product-%d", i),
			"type":               "simple",
			"sku":                fmt.Sprintf("FIX-%d", i),
			"price":              fmt.Sprintf("%d.00", 10*i),
			"regular_price":      fmt.Sprintf("%d.00", 10*i),
			"sale_price":         "",
			"stock_status":       "instock",
			"stock_quantity":     100 * i,
			"total_sales":        i,
			"email":              fmt.Sprintf("fixture+%d@example.com", i),
			"first_name":         "Fixture",
			"last_name":          strconv.Itoa(i),
			"username":           fmt.Sprintf("fixture%d", i),
			"role":               "customer",
			"is_paying_customer": true,
			"code":               fmt.Sprintf("FIXTURE%d", i),
			"discount_type":      "percent",
			"amount":             "10",
			"usage_count":        i,
			"usage_limit":        100,
			"date_expires":       "",
			"date_created":       ts,
			"date_created_gmt":   ts,
			"date_modified":      ts,
			"date_modified_gmt":  ts,
			"date_paid":          ts,
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

// requester builds a connsdk.Requester wired with HTTP Basic auth (consumer key
// as username, consumer secret as password) and the resolved base URL. The
// secrets only ever flow into connsdk.Basic; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := woocommerceBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	key, secret := woocommerceSecrets(cfg)
	if strings.TrimSpace(key) == "" || strings.TrimSpace(secret) == "" {
		return nil, errors.New("woocommerce connector requires secrets api_key and api_secret")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(strings.TrimSpace(key), strings.TrimSpace(secret)),
		UserAgent: woocommerceUserAgent,
	}, nil
}

// incrementalLowerBound returns the ISO8601 lower bound for the modified_after
// filter, derived from the incremental cursor (if any) or else the start_date
// config. An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	// WooCommerce accepts ISO8601; a bare YYYY-MM-DD is widened to midnight UTC.
	if t, err := time.Parse("2006-01-02", startDate); err == nil {
		return t.UTC().Format("2006-01-02T15:04:05"), nil
	}
	if t, err := time.Parse(time.RFC3339, startDate); err == nil {
		return t.UTC().Format("2006-01-02T15:04:05"), nil
	}
	return "", fmt.Errorf("woocommerce config start_date must be YYYY-MM-DD or RFC3339, got %q", startDate)
}

func woocommerceSecrets(cfg connectors.RuntimeConfig) (string, string) {
	if cfg.Secrets == nil {
		return "", ""
	}
	return cfg.Secrets["api_key"], cfg.Secrets["api_secret"]
}

// woocommerceBaseURL resolves and validates the base URL. An explicit base_url
// config override (used by tests and self-hosted gateways) wins; otherwise it is
// derived from the shop config as https://<shop>/wp-json/wc/v3. Any base must be
// an absolute https (or http for local test servers) URL with a host to bound
// SSRF risk.
func woocommerceBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if override := strings.TrimSpace(cfg.Config["base_url"]); override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("woocommerce config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("woocommerce config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("woocommerce config base_url must include a host")
		}
		base := strings.TrimRight(override, "/")
		// Append the wc/v3 API path unless the override already includes it.
		if !strings.Contains(base, "/wp-json/") {
			base += woocommerceAPIPath
		}
		return base, nil
	}

	shop := strings.TrimSpace(cfg.Config["shop"])
	if shop == "" {
		return "", errors.New("woocommerce connector requires config shop (e.g. EXAMPLE.com) or base_url")
	}
	// shop is a bare host like "example.com"; reject schemes/paths that would
	// open an SSRF vector or break URL construction.
	host := strings.TrimPrefix(strings.TrimPrefix(shop, "https://"), "http://")
	host = strings.TrimRight(host, "/")
	if host == "" || strings.ContainsAny(host, "/?#@") {
		return "", fmt.Errorf("woocommerce config shop must be a bare host (e.g. EXAMPLE.com), got %q", shop)
	}
	parsed, err := url.Parse("https://" + host)
	if err != nil || parsed.Host == "" {
		return "", fmt.Errorf("woocommerce config shop is not a valid host: %q", shop)
	}
	return "https://" + host + woocommerceAPIPath, nil
}

func woocommercePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return woocommerceDefaultPage, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("woocommerce config page_size must be an integer: %w", err)
	}
	if value < 1 || value > woocommerceMaxPageSize {
		return 0, fmt.Errorf("woocommerce config page_size must be between 1 and %d", woocommerceMaxPageSize)
	}
	return value, nil
}

func woocommerceMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("woocommerce config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("woocommerce config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// parseTotalPages parses the X-WP-TotalPages header; an absent or invalid value
// yields 0, which signals the harvest loop to fall back to short-page detection.
func parseTotalPages(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0
	}
	return n
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
