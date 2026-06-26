// Package getlago implements the native pm Lago (getlago) connector. It follows
// the declarative-HTTP template established by the stripe package: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction) with Lago-specific stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// Lago's source is read-only (full refresh only), so Write is unsupported and the
// Write capability is false.
package getlago

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
	getlagoDefaultBaseURL  = "https://api.getlago.com/api/v1"
	getlagoDefaultPageSize = 100
	getlagoMaxPageSize     = 100
	getlagoUserAgent       = "polymetrics-go-cli"
	// getlagoFixtureCreated is the deterministic created_at timestamp used by the
	// fixture-mode records.
	getlagoFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("getlago", New)
}

// New returns the Lago connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Lago connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "getlago" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "getlago",
		DisplayName:     "Lago",
		IntegrationType: "api",
		Description:     "Reads Lago customers, invoices, subscriptions, plans, and billable metrics through the Lago REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Lago. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := getlagoBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(getlagoSecret(cfg)) == "" {
		return errors.New("getlago connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the customers list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "customers", url.Values{"per_page": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check getlago: %w", err)
	}
	return nil
}

// Write is unsupported: the Lago source connector is read-only (full refresh
// only). The Write capability is reported false in Metadata.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: getlagoStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	endpoint, ok := getlagoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("getlago stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := getlagoPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := getlagoMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Lago's page-number pagination. Lago lists return
// {<resource>:[...], meta:{current_page, next_page, total_pages, ...}}; the next
// page is requested with page=<meta.next_page> until next_page is null. There is
// no connsdk paginator that reads a numeric next-page token out of the body for
// this exact shape, so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))

	page := 1
	for pageNum := 0; maxPages == 0 || pageNum < maxPages; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read getlago %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode getlago %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		nextPage, err := connsdk.StringAt(resp.Body, "meta.next_page")
		if err != nil {
			return fmt.Errorf("decode getlago %s meta.next_page: %w", endpoint.resource, err)
		}
		next := strings.TrimSpace(nextPage)
		if next == "" || next == "null" {
			return nil
		}
		parsed, err := strconv.Atoi(next)
		if err != nil || parsed <= page {
			return nil
		}
		page = parsed
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise getlago credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"lago_id":              fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"sequential_id":        int64(i),
			"external_id":          fmt.Sprintf("ext_%d", i),
			"external_customer_id": fmt.Sprintf("ext_cust_%d", i),
			"lago_customer_id":     "cust_fixture_1",
			"slug":                 fmt.Sprintf("%s-%d", stream, i),
			"name":                 fmt.Sprintf("Fixture %d", i),
			"email":                fmt.Sprintf("fixture+%d@example.com", i),
			"currency":             "USD",
			"amount_currency":      "USD",
			"country":              "US",
			"customer_type":        "company",
			"number":               fmt.Sprintf("INV-%03d", i),
			"issuing_date":         "2026-01-01",
			"status":               "finalized",
			"payment_status":       "succeeded",
			"invoice_type":         "subscription",
			"fees_amount_cents":    int64(1000 * i),
			"taxes_amount_cents":   int64(100 * i),
			"total_amount_cents":   int64(1100 * i),
			"plan_code":            "starter",
			"code":                 fmt.Sprintf("code_%d", i),
			"interval":             "monthly",
			"amount_cents":         int64(1000 * i),
			"pay_in_advance":       true,
			"trial_period":         float64(0),
			"aggregation_type":     "count_agg",
			"field_name":           "",
			"recurring":            false,
			"billing_time":         "calendar",
			"started_at":           getlagoFixtureCreated,
			"terminated_at":        nil,
			"created_at":           getlagoFixtureCreated,
			"updated_at":           getlagoFixtureCreated,
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
	base, err := getlagoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := getlagoSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("getlago connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: getlagoUserAgent,
	}, nil
}

func getlagoSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// getlagoBaseURL resolves and validates the base URL. The default is
// api.getlago.com; any override (config api_url, or base_url alias) must be an
// absolute https (or http for local test servers) URL with a host to bound SSRF
// risk.
func getlagoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["api_url"])
	if base == "" {
		base = strings.TrimSpace(cfg.Config["base_url"])
	}
	if base == "" {
		return getlagoDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("getlago config api_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("getlago config api_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("getlago config api_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func getlagoPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return getlagoDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("getlago config page_size must be an integer: %w", err)
	}
	if value < 1 || value > getlagoMaxPageSize {
		return 0, fmt.Errorf("getlago config page_size must be between 1 and %d", getlagoMaxPageSize)
	}
	return value, nil
}

func getlagoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("getlago config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("getlago config max_pages must be 0 for unlimited or a positive integer")
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
