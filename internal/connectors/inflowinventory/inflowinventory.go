// Package inflowinventory implements the native pm inFlow Inventory connector.
// It follows the stripe declarative-HTTP template: a thin package composing the
// connsdk toolkit (Requester + API-key header auth + top-level array extraction)
// with inFlow-specific stream definitions, endpoints, and cursor-after
// pagination.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// inFlow Inventory is read-only here: the cloud API is primarily a system of
// record for inventory data, so this connector exposes Check/Catalog/Read and
// sets Capabilities.Write=false.
package inflowinventory

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
	inflowDefaultBaseURL  = "https://cloudapi.inflowinventory.com"
	inflowAcceptHeader    = "application/json;version=2024-03-12"
	inflowUserAgent       = "polymetrics-go-cli"
	inflowDefaultPageSize = 100
	inflowMaxPageSize     = 100
)

func init() {
	connectors.RegisterFactory("inflowinventory", New)
}

// New returns the inFlow Inventory connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm inFlow Inventory connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "inflowinventory" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "inflowinventory",
		DisplayName:     "inFlow Inventory",
		IntegrationType: "api",
		Description:     "Reads inFlow Inventory products, customers, vendors, sales orders, and categories through the inFlow cloud REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to inFlow. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := inflowBaseURL(cfg); err != nil {
		return err
	}
	companyID, err := inflowCompanyID(cfg)
	if err != nil {
		return err
	}
	if strings.TrimSpace(inflowSecret(cfg)) == "" {
		return errors.New("inflowinventory connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the categories list confirms auth and connectivity
	// without mutating anything.
	path := companyID + "/categories"
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"count": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check inflowinventory: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: inflowStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "products"
	}
	endpoint, ok := inflowStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("inflowinventory stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	companyID, err := inflowCompanyID(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := inflowPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := inflowMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, companyID, endpoint, pageSize, maxPages, emit)
}

// harvest drives inFlow's count/after cursor pagination. inFlow list endpoints
// return a top-level JSON array sorted by the resource id; the next page is
// requested with after=<last record id>. A page shorter than the requested
// count signals the end. There is no body token paginator in connsdk for this
// shape, so the loop lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, companyID string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := companyID + "/" + endpoint.resource
	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("count", strconv.Itoa(pageSize))
		if after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read inflowinventory %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode inflowinventory %s page: %w", endpoint.resource, err)
		}
		lastID := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			lastID = stringField(item, endpoint.primaryKey)
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A short page (or no advanceable cursor) means we are done.
		if len(records) < pageSize || lastID == "" {
			return nil
		}
		after = lastID
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise inflowinventory credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := fmt.Sprintf("%s_fixture_%d", endpoint.resource, i)
		item := map[string]any{
			endpoint.primaryKey:    id,
			"name":                 fmt.Sprintf("Fixture %s %d", stream, i),
			"sku":                  fmt.Sprintf("SKU-%d", i),
			"description":          "fixture record",
			"contactName":          fmt.Sprintf("Contact %d", i),
			"email":                fmt.Sprintf("fixture+%d@example.com", i),
			"phone":                "+1-555-0100",
			"isActive":             true,
			"itemType":             "StockedProduct",
			"currencyId":           "USD",
			"customerId":           "customers_fixture_1",
			"inventoryStatus":      "Available",
			"isCompleted":          false,
			"isDefault":            i == 1,
			"timestamp":            fmt.Sprintf("AAAAAAAAB%dk=", i),
			"lastModifiedDateTime": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
		}
		record := endpoint.mapRecord(item)
		// Surface a stable `id` so the conformance harness can key records
		// regardless of the per-stream primary-key name.
		record["id"] = id
		record["connector"] = "inflowinventory"
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with API-key header auth and the
// resolved base URL. inFlow expects the raw API key in the Authorization header
// (no Bearer prefix) plus a versioned Accept header. The secret only ever flows
// into connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := inflowBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := inflowSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("inflowinventory connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, ""),
		UserAgent: inflowUserAgent,
		Accept:    inflowAcceptHeader,
	}, nil
}

func inflowSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// inflowCompanyID resolves the required companyid config value, which inFlow
// embeds at the start of every resource path. It is validated to be a single
// safe path segment to keep it from escaping the intended host/path.
func inflowCompanyID(cfg connectors.RuntimeConfig) (string, error) {
	id := strings.TrimSpace(cfg.Config["companyid"])
	if id == "" {
		return "", errors.New("inflowinventory connector requires config companyid")
	}
	if strings.ContainsAny(id, "/?#") || strings.Contains(id, "..") {
		return "", fmt.Errorf("inflowinventory config companyid contains invalid characters: %q", id)
	}
	return id, nil
}

// inflowBaseURL resolves and validates the base URL. The default is
// cloudapi.inflowinventory.com; any override must be an absolute https (or http
// for local test servers) URL with a host to bound SSRF risk.
func inflowBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return inflowDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("inflowinventory config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("inflowinventory config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("inflowinventory config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func inflowPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return inflowDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("inflowinventory config page_size must be an integer: %w", err)
	}
	if value < 1 || value > inflowMaxPageSize {
		return 0, fmt.Errorf("inflowinventory config page_size must be between 1 and %d", inflowMaxPageSize)
	}
	return value, nil
}

func inflowMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("inflowinventory config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("inflowinventory config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Write is unsupported: inFlow Inventory is exposed read-only.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
