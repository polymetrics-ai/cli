// Package omnisend implements the native pm Omnisend connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + APIKeyHeader auth + RecordsAt extraction + cursor pagination)
// with Omnisend-specific stream definitions and endpoints.
//
// Omnisend is a read-only source here: the public REST API is list-oriented and
// the catalog only advertises full_refresh syncs, so Capabilities.Write is
// false and no reverse-ETL write actions are exposed.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package omnisend

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
	omnisendDefaultBaseURL  = "https://api.omnisend.com/v3"
	omnisendDefaultPageSize = 100
	omnisendMaxPageSize     = 250
	omnisendUserAgent       = "polymetrics-go-cli"
	// omnisendAPIKeyHeader is the header Omnisend uses for API-key auth.
	omnisendAPIKeyHeader = "X-API-KEY"
)

func init() {
	connectors.RegisterFactory("omnisend", New)
}

// New returns the Omnisend connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Omnisend connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "omnisend" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "omnisend",
		DisplayName:     "Omnisend",
		IntegrationType: "api",
		Description:     "Reads Omnisend contacts, campaigns, carts, orders, and products through the Omnisend REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Omnisend. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := omnisendBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(omnisendSecret(cfg)) == "" {
		return errors.New("omnisend connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the contacts list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "contacts", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check omnisend: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: omnisendStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: Omnisend streams start with
// an empty cursor (full sync). Omnisend itself only supports full_refresh, but
// callers may still track a high-water mark over createdAt.
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
		stream = "contacts"
	}
	endpoint, ok := omnisendStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("omnisend stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := omnisendPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := omnisendMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Omnisend's cursor pagination. List responses are shaped
// {<resource>:[...], paging:{next:"<full url>"}}. The next page is requested by
// sending a GET to the full URL in paging.next; pagination stops when paging.next
// is null/absent. connsdk.Requester.resolveURL treats an absolute http(s) path as
// the request URL, so paging.next is passed straight through.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// First page: relative resource path with the page-size limit.
	path := endpoint.resource
	query := url.Values{}
	query.Set("limit", strconv.Itoa(pageSize))

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read omnisend %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode omnisend %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "paging.next")
		if err != nil {
			return fmt.Errorf("decode omnisend %s paging: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		// paging.next is a full URL that already carries the cursor and limit;
		// pass it through verbatim and clear the relative query so we do not
		// override the cursor.
		path = next
		query = nil
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise omnisend credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := fmt.Sprintf("%s_fixture_%d", endpoint.resource, i)
		item := map[string]any{
			endpoint.primaryKey: id,
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"firstName":         fmt.Sprintf("Fixture%d", i),
			"lastName":          "Example",
			"name":              fmt.Sprintf("Fixture Campaign %d", i),
			"subject":           "Fixture subject",
			"title":             fmt.Sprintf("Fixture Product %d", i),
			"description":       "Fixture description",
			"status":            "active",
			"type":              "regular",
			"currency":          "USD",
			"country":           "United States",
			"countryCode":       "US",
			"city":              "Springfield",
			"state":             "IL",
			"fromName":          "Fixture Store",
			"vendor":            "Fixture Vendor",
			"productUrl":        "https://example.com/p/" + id,
			"cartRecoveryUrl":   "https://example.com/cart/" + id,
			"contactID":         "contacts_fixture_1",
			"cartID":            "carts_fixture_1",
			"orderNumber":       int64(1000 + i),
			"cartSum":           int64(50 * i),
			"orderSum":          int64(100 * i),
			"subTotalSum":       int64(90 * i),
			"taxSum":            int64(10 * i),
			"shippingSum":       int64(5),
			"discountSum":       int64(0),
			"sent":              int64(1000 * i),
			"opened":            int64(400 * i),
			"clicked":           int64(100 * i),
			"bounced":           int64(2 * i),
			"unsubscribed":      int64(1 * i),
			"paymentStatus":     "paid",
			"fulfillmentStatus": "fulfilled",
			"createdAt":         fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updatedAt":         fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"startDate":         "2026-01-01T00:00:00Z",
			"endDate":           "2026-01-02T00:00:00Z",
			"segments":          []any{"vip"},
			"tags":              []any{"fixture"},
			"categoryIDs":       []any{"cat_1"},
			"variants":          []any{},
			"images":            []any{},
			"products":          []any{},
			"fixture":           true,
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

// requester builds a connsdk.Requester wired with X-API-KEY auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := omnisendBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := omnisendSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("omnisend connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(omnisendAPIKeyHeader, secret, ""),
		UserAgent: omnisendUserAgent,
	}, nil
}

func omnisendSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// omnisendBaseURL resolves and validates the base URL. The default is
// api.omnisend.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func omnisendBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return omnisendDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("omnisend config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("omnisend config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("omnisend config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func omnisendPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return omnisendDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("omnisend config page_size must be an integer: %w", err)
	}
	if value < 1 || value > omnisendMaxPageSize {
		return 0, fmt.Errorf("omnisend config page_size must be between 1 and %d", omnisendMaxPageSize)
	}
	return value, nil
}

func omnisendMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("omnisend config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("omnisend config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// Write satisfies the connectors.Connector interface. Omnisend is exposed as a
// read-only source here, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
