// Package adobecommercemagento implements the native pm Adobe Commerce (Magento)
// connector. It is a declarative-HTTP per-system connector following the stripe
// template: a thin package that composes the connsdk toolkit (Requester + Bearer
// auth + RecordsAt extraction + cursor state) with Magento-specific stream
// definitions, endpoints, and searchCriteria pagination.
//
// The Adobe Commerce REST API authenticates with an Integration Access Token sent
// as Authorization: Bearer <api_key>. List endpoints live under
// https://<store_host>/rest/<api_version>/<resource> and return
// {"items":[...],"total_count":N,"search_criteria":{...}}. Pagination is driven by
// the 1-based searchCriteria[currentPage] param alongside searchCriteria[pageSize];
// the read loop walks pages until the accumulated record count reaches total_count
// (or a short page is returned). The Magento source is read-only (full refresh /
// incremental on updated_at); it exposes no reverse-ETL writes, so
// Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package adobecommercemagento

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	magentoDefaultAPIVersion = "V1"
	magentoDefaultPageSize   = 100
	magentoMaxPageSize       = 300
	magentoUserAgent         = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("adobe-commerce-magento", New)
}

// New returns the Adobe Commerce (Magento) connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Adobe Commerce (Magento) connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "adobe-commerce-magento" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "adobe-commerce-magento",
		DisplayName:     "Adobe Commerce (Magento)",
		IntegrationType: "api",
		Description:     "Reads Adobe Commerce (Magento) products, orders, customers, categories, and invoices through the Magento REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Magento. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := magentoBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(magentoSecret(cfg)) == "" {
		return errors.New("adobe-commerce-magento connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the products list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{}
	query.Set("searchCriteria[pageSize]", "1")
	query.Set("searchCriteria[currentPage]", "1")
	if err := r.DoJSON(ctx, http.MethodGet, "products", query, nil, nil); err != nil {
		return fmt.Errorf("check adobe-commerce-magento: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: magentoStreams()}, nil
}

// Write is not supported: the Magento source is read-only. It exists to satisfy
// the connectors.Connector interface and always reports the operation as
// unsupported (Capabilities.Write is false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Magento stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
	endpoint, ok := magentoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("adobe-commerce-magento stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := magentoPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := magentoMaxPages(req.Config)
	if err != nil {
		return err
	}
	lower := incrementalLowerBound(req)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives Magento's searchCriteria pagination. Magento list endpoints
// return {"items":[...],"total_count":N,...}; the next page is requested by
// incrementing the 1-based searchCriteria[currentPage]. The loop stops when the
// accumulated record count reaches total_count, a short page is returned, or
// maxPages is hit. There is no body-token paginator in connsdk for this exact
// shape, so the loop lives here, built on connsdk.Requester + connsdk.RecordsAt +
// connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, updatedGT string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("searchCriteria[pageSize]", strconv.Itoa(pageSize))
	if updatedGT != "" {
		// Filter on updated_at > cursor (incremental). filter_groups are ANDed,
		// filters within a group are ORed; a single filter is fine here.
		base.Set("searchCriteria[filter_groups][0][filters][0][field]", "updated_at")
		base.Set("searchCriteria[filter_groups][0][filters][0][value]", updatedGT)
		base.Set("searchCriteria[filter_groups][0][filters][0][condition_type]", "gt")
	}

	seen := 0
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("searchCriteria[currentPage]", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read adobe-commerce-magento %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode adobe-commerce-magento %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		seen += len(records)

		// A short page means we have exhausted the result set.
		if len(records) < pageSize || len(records) == 0 {
			return nil
		}
		// total_count, when present, gives a definitive stop condition.
		total := magentoTotalCount(resp.Body)
		if total > 0 && seen >= total {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free (mirrors
// stripe's fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  int64(i),
			"entity_id":           int64(i),
			"order_id":            int64(i),
			"sku":                 fmt.Sprintf("%s-SKU-%d", stream, i),
			"increment_id":        fmt.Sprintf("%09d", 1000+i),
			"name":                fmt.Sprintf("Fixture %s %d", stream, i),
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"firstname":           "Fixture",
			"lastname":            fmt.Sprintf("User%d", i),
			"price":               9.99 * float64(i),
			"grand_total":         19.99 * float64(i),
			"base_grand_total":    19.99 * float64(i),
			"order_currency_code": "USD",
			"customer_id":         int64(i),
			"customer_email":      fmt.Sprintf("fixture+%d@example.com", i),
			"group_id":            int64(1),
			"store_id":            int64(1),
			"website_id":          int64(1),
			"parent_id":           int64(2),
			"is_active":           true,
			"position":            int64(i),
			"level":               int64(2),
			"product_count":       int64(i),
			"status":              1,
			"state":               1,
			"visibility":          int64(4),
			"type_id":             "simple",
			"attribute_set_id":    int64(4),
			"weight":              1.0,
			"created_at":          "2026-01-01 00:00:00",
			"updated_at":          fmt.Sprintf("2026-01-0%d 00:00:00", i),
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
	base, err := magentoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := magentoSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("adobe-commerce-magento connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: magentoUserAgent,
	}, nil
}

// incrementalLowerBound returns the updated_at lower bound for the gt filter,
// derived from the incremental cursor (if any) or else the start_date config. An
// empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func magentoSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// magentoBaseURL resolves and validates the base URL of the form
// https://<store_host>/rest/<api_version>. An explicit base_url override (used by
// tests and self-hosted gateways) wins; otherwise store_host + api_version build
// it. Any URL must be an absolute https (or http for local test servers) URL with
// a host to bound SSRF risk.
func magentoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if override := strings.TrimSpace(cfg.Config["base_url"]); override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("adobe-commerce-magento config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("adobe-commerce-magento config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("adobe-commerce-magento config base_url must include a host")
		}
		return strings.TrimRight(override, "/") + "/rest/" + magentoAPIVersion(cfg), nil
	}

	host := strings.TrimSpace(cfg.Config["store_host"])
	if host == "" {
		return "", errors.New("adobe-commerce-magento config requires store_host (e.g. magento.mystore.com) or base_url")
	}
	// store_host may be given with or without a scheme; normalize to https.
	if !strings.Contains(host, "://") {
		host = "https://" + host
	}
	parsed, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("adobe-commerce-magento config store_host is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("adobe-commerce-magento config store_host must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("adobe-commerce-magento config store_host must include a host")
	}
	return strings.TrimRight(parsed.Scheme+"://"+parsed.Host, "/") + "/rest/" + magentoAPIVersion(cfg), nil
}

func magentoAPIVersion(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["api_version"]); v != "" {
		return v
	}
	return magentoDefaultAPIVersion
}

func magentoPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return magentoDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("adobe-commerce-magento config page_size must be an integer: %w", err)
	}
	if value < 1 || value > magentoMaxPageSize {
		return 0, fmt.Errorf("adobe-commerce-magento config page_size must be between 1 and %d", magentoMaxPageSize)
	}
	return value, nil
}

func magentoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("adobe-commerce-magento config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("adobe-commerce-magento config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// magentoTotalCount reads the integer total_count from a list response body. It
// returns 0 when absent or unparseable, which makes the short-page check the sole
// stop condition.
func magentoTotalCount(body []byte) int {
	raw, err := connsdk.StringAt(body, "total_count")
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
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
