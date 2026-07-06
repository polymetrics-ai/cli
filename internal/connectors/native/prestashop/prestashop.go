// Package prestashop implements the native pm PrestaShop connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + HTTP Basic auth +
// RecordsAt extraction + cursor state) with PrestaShop-specific stream
// definitions, endpoints, and pagination.
//
// The PrestaShop Webservice authenticates with HTTP Basic where the access key
// is the username and the password is empty. Responses default to XML; this
// connector requests JSON via output_format=JSON and full field expansion via
// display=full. List responses are wrapped in an object keyed by the resource
// name (e.g. {"customers":[...]}). Pagination uses limit=<offset>,<count> and
// incremental reads filter on date_upd.
package prestashop

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
	prestashopAPIPath      = "api"
	prestashopDefaultPage  = 100
	prestashopMaxPageSize  = 1000
	prestashopUserAgent    = "polymetrics-go-cli"
	prestashopFixtureDate  = "2026-01-01 00:00:00"
	prestashopMaxDateBound = "2100-01-01 00:00:00"
)

// New returns the PrestaShop connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm PrestaShop connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "prestashop" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "prestashop",
		DisplayName:     "PrestaShop",
		IntegrationType: "api",
		Description:     "Reads PrestaShop customers, orders, products, addresses, and carts through the PrestaShop Webservice REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to PrestaShop.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := prestashopBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(prestashopSecret(cfg)) == "" {
		return errors.New("prestashop connector requires secret access_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the customers list confirms auth and connectivity
	// without mutating anything.
	query := url.Values{}
	query.Set("output_format", "JSON")
	query.Set("limit", "1")
	if err := r.DoJSON(ctx, http.MethodGet, "customers", query, nil, nil); err != nil {
		return fmt.Errorf("check prestashop: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: prestashopStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a PrestaShop stream starts
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
	endpoint, ok := prestashopStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("prestashop stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := prestashopPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := prestashopMaxPages(req.Config)
	if err != nil {
		return err
	}
	lower, err := incrementalLowerBound(req)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives PrestaShop's limit=<offset>,<count> pagination. List responses
// are wrapped in an object keyed by the resource name (e.g. {"customers":[...]}),
// so records are extracted at that path. A page shorter than the requested
// count signals the end of the stream. The loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, dateLower string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("output_format", "JSON")
	base.Set("display", "full")
	if dateLower != "" {
		base.Set("date", "1")
		base.Set("filter[date_upd]", fmt.Sprintf("[%s,%s]", dateLower, prestashopMaxDateBound))
		base.Set("sort", "[date_upd_ASC]")
	}

	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("limit", fmt.Sprintf("%d,%d", offset, pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read prestashop %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.resource)
		if err != nil {
			return fmt.Errorf("decode prestashop %s page: %w", endpoint.resource, err)
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
// conformance harness can exercise prestashop credential-free (mirrors the
// stripe fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  i,
			"id_customer":         i,
			"id_default_group":    1,
			"id_lang":             1,
			"id_country":          1,
			"id_state":            0,
			"id_currency":         1,
			"id_carrier":          1,
			"id_manufacturer":     0,
			"id_supplier":         0,
			"id_category_default": 2,
			"id_address_delivery": i,
			"id_address_invoice":  i,
			"current_state":       2,
			"firstname":           fmt.Sprintf("Fixture%d", i),
			"lastname":            "Example",
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"company":             "Example Co",
			"city":                "Paris",
			"postcode":            "75001",
			"phone":               "0102030405",
			"reference":           fmt.Sprintf("REF%05d", i),
			"payment":             "Bank wire",
			"total_paid":          fmt.Sprintf("%d.00", 100*i),
			"total_paid_real":     fmt.Sprintf("%d.00", 100*i),
			"price":               fmt.Sprintf("%d.99", 10*i),
			"quantity":            i,
			"active":              "1",
			"valid":               "1",
			"newsletter":          "0",
			"date_add":            prestashopFixtureDate,
			"date_upd":            prestashopFixtureDate,
			"fixture":             true,
			"connector":           "prestashop",
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

// Write is unsupported: PrestaShop is read-only for this connector. It satisfies
// the Connector interface but performs no mutation.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with HTTP Basic auth (access key as
// username, empty password) and the resolved base URL. The secret only ever
// flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := prestashopBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := prestashopSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("prestashop connector requires secret access_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: prestashopUserAgent,
	}, nil
}

// incrementalLowerBound returns the date_upd lower bound for the date filter,
// derived from the incremental cursor (if any) or else the start_date config.
// An empty result means no lower bound (full sync). PrestaShop expects datetime
// strings of the form "YYYY-MM-DD HH:MM:SS"; a bare start_date (YYYY-MM-DD) is
// expanded to the start of that day.
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return normalizeDateTime(cursor), nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	return normalizeDateTime(startDate), nil
}

// normalizeDateTime expands a bare date (YYYY-MM-DD) to a full PrestaShop
// datetime, and passes through values that already look like datetimes.
func normalizeDateTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.Format("2006-01-02") + " 00:00:00"
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC().Format("2006-01-02 15:04:05")
	}
	return value
}

func prestashopSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_key"]
}

// prestashopBaseURL resolves and validates the shop URL, appending the /api
// path. The shop URL comes from config "base_url" (or "url"); any value must be
// an absolute https (or http for local test servers) URL with a host to bound
// SSRF risk.
func prestashopBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = strings.TrimSpace(cfg.Config["url"])
	}
	if base == "" {
		return "", errors.New("prestashop connector requires config base_url (shop URL)")
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("prestashop config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("prestashop config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("prestashop config base_url must include a host")
	}
	trimmed := strings.TrimRight(base, "/")
	// If the URL already targets the webservice (.../api), keep it as-is.
	if strings.HasSuffix(trimmed, "/"+prestashopAPIPath) {
		return trimmed, nil
	}
	return trimmed + "/" + prestashopAPIPath, nil
}

func prestashopPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return prestashopDefaultPage, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("prestashop config page_size must be an integer: %w", err)
	}
	if value < 1 || value > prestashopMaxPageSize {
		return 0, fmt.Errorf("prestashop config page_size must be between 1 and %d", prestashopMaxPageSize)
	}
	return value, nil
}

func prestashopMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("prestashop config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("prestashop config max_pages must be 0 for unlimited or a positive integer")
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
