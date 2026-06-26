// Package judgemereviews implements the native pm Judge.me Reviews connector. It
// is a declarative-HTTP per-system connector built on the same shape as the
// stripe reference: a thin package that composes the connsdk toolkit (Requester +
// query-param API-key auth + RecordsAt extraction) with Judge.me-specific stream
// definitions, endpoints, and pagination.
//
// The Judge.me API authenticates with two query parameters on every request:
// api_token (the secret api_key) and shop_domain (the *.myshopify.com store).
// List endpoints return {"<resource>":[...]} and paginate with page/per_page.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package judgemereviews

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
	registryName     = "judge-me-reviews"
	defaultBaseURL   = "https://judge.me/api/v1"
	defaultPageSize  = 100
	maxPageSize      = 100
	defaultMaxPages  = 1000
	userAgent        = "polymetrics-go-cli"
	fixtureCreatedAt = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Judge.me Reviews connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Judge.me Reviews connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Judge.me Reviews",
		IntegrationType: "api",
		Description:     "Reads Judge.me reviews, products, and widgets for a Shopify shop through the Judge.me REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Judge.me. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("judge-me-reviews connector requires secret api_key")
	}
	if strings.TrimSpace(shopDomain(cfg)) == "" {
		return errors.New("judge-me-reviews connector requires config shop_domain")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the reviews list confirms auth and connectivity without
	// mutating anything.
	query := url.Values{"per_page": []string{"1"}, "page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "reviews", query, nil, nil); err != nil {
		return fmt.Errorf("check judge-me-reviews: %w", err)
	}
	return nil
}

// Write is unsupported: Judge.me Reviews is a read-only source. The method
// exists only to satisfy connectors.Connector; Capabilities.Write is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "reviews"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("judge-me-reviews stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Judge.me's page-number pagination. List endpoints return
// {"<resource>":[...]}; the next page is requested by incrementing page until a
// short (or empty) page is returned. The auth query params (api_token,
// shop_domain) are supplied by the Requester's Authenticator, so they are not
// repeated here.
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
			return fmt.Errorf("read judge-me-reviews %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode judge-me-reviews %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A page shorter than the requested size means the last page was reached.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  int64(i),
			"title":               fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"body":                "Fixture review body",
			"rating":              int64(5),
			"product_external_id": fmt.Sprintf("prod_%d", i),
			"external_id":         fmt.Sprintf("prod_%d", i),
			"handle":              fmt.Sprintf("fixture-product-%d", i),
			"url":                 fmt.Sprintf("https://example.myshopify.com/products/fixture-%d", i),
			"name":                fmt.Sprintf("Fixture Widget %d", i),
			"widget_type":         "review_widget",
			"status":              "active",
			"source":              "web",
			"curated":             "ok",
			"published":           true,
			"hidden":              false,
			"verified":            "buyer",
			"created_at":          fixtureCreatedAt,
			"updated_at":          fixtureCreatedAt,
			"reviewer": map[string]any{
				"id":    int64(100 + i),
				"name":  fmt.Sprintf("Fixture Reviewer %d", i),
				"email": fmt.Sprintf("reviewer+%d@example.com", i),
			},
		}
		record := endpoint.mapRecord(item)
		record["connector"] = registryName
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with query-param API-key auth and
// the resolved base URL. The secret only ever flows into the Authenticator; it is
// never logged. shop_domain is also carried as a query param on every request.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("judge-me-reviews connector requires secret api_key")
	}
	shop := strings.TrimSpace(shopDomain(cfg))
	if shop == "" {
		return nil, errors.New("judge-me-reviews connector requires config shop_domain")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      queryAuth{token: token, shop: shop},
		UserAgent: userAgent,
	}, nil
}

// queryAuth applies Judge.me's two required query parameters (api_token and
// shop_domain) to every outgoing request. It composes the same way connsdk's
// APIKeyQuery does, but for two params at once. It never logs the secret value.
type queryAuth struct {
	token string
	shop  string
}

func (a queryAuth) Apply(_ context.Context, req *http.Request) error {
	q := req.URL.Query()
	q.Set("api_token", a.token)
	q.Set("shop_domain", a.shop)
	req.URL.RawQuery = q.Encode()
	return nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

func shopDomain(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["shop_domain"]
}

// baseURL resolves and validates the base URL. The default is judge.me/api/v1;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("judge-me-reviews config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("judge-me-reviews config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("judge-me-reviews config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("judge-me-reviews config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("judge-me-reviews config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return defaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("judge-me-reviews config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("judge-me-reviews config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
