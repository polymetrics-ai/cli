// Package shopify implements a read-only Shopify REST Admin API connector. The
// REST Admin API is legacy for new public apps, but remains a documented surface
// for conservative read-only customer, order, and product extraction.
package shopify

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
	shopifyName            = "shopify"
	shopifyDefaultVersion  = "2026-01"
	shopifyDefaultPageSize = 250
	shopifyMaxPageSize     = 250
	shopifyUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("shopify", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return shopifyName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: shopifyName, DisplayName: "Shopify", IntegrationType: "api", Description: "Reads Shopify customers, orders, and products through the REST Admin API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if strings.TrimSpace(shopifyAccessToken(cfg)) == "" {
		return errors.New("shopify connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "shop.json", nil, nil, nil); err != nil {
		return fmt.Errorf("check shopify: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: shopifyStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	endpoint, ok := shopifyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("shopify stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := shopifyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := shopifyMaxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{"limit": []string{strconv.Itoa(pageSize)}}
	paginator := &connsdk.LinkHeaderPaginator{}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) readFixture(ctx context.Context, endpoint shopifyStreamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": i, "email": fmt.Sprintf("fixture%d@example.com", i), "title": fmt.Sprintf("Fixture Product %d", i), "name": fmt.Sprintf("#10%d", i), "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z", "total_price": "10.00"}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := shopifyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := shopifyAccessToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("shopify connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("X-Shopify-Access-Token", token, ""), UserAgent: shopifyUserAgent}, nil
}

type shopifyStreamEndpoint struct {
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var shopifyStreamEndpoints = map[string]shopifyStreamEndpoint{
	"customers": {resource: "customers.json", recordsPath: "customers", mapRecord: shopifyCustomerRecord},
	"orders":    {resource: "orders.json", recordsPath: "orders", mapRecord: shopifyOrderRecord},
	"products":  {resource: "products.json", recordsPath: "products", mapRecord: shopifyProductRecord},
}

func shopifyStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "customers", Description: "Shopify customers.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "email", Type: "string"}, {Name: "created_at", Type: "string"}, {Name: "updated_at", Type: "string"}}},
		{Name: "orders", Description: "Shopify orders.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "total_price", Type: "string"}, {Name: "created_at", Type: "string"}, {Name: "updated_at", Type: "string"}}},
		{Name: "products", Description: "Shopify products.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "title", Type: "string"}, {Name: "created_at", Type: "string"}, {Name: "updated_at", Type: "string"}}},
	}
}

func shopifyCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "email": item["email"], "created_at": item["created_at"], "updated_at": item["updated_at"]}
}

func shopifyOrderRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "email": item["email"], "total_price": item["total_price"], "created_at": item["created_at"], "updated_at": item["updated_at"]}
}

func shopifyProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "title": item["title"], "created_at": item["created_at"], "updated_at": item["updated_at"]}
}

func shopifyAccessToken(cfg connectors.RuntimeConfig) string { return cfg.Secrets["access_token"] }

func shopifyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		return validBaseURL(shopifyName, base, "")
	}
	shop := strings.TrimSpace(cfg.Config["shop"])
	if shop == "" {
		return "", errors.New("shopify connector requires config shop or base_url")
	}
	shop = strings.TrimPrefix(strings.TrimPrefix(shop, "https://"), "http://")
	shop = strings.Trim(shop, "/")
	if !strings.Contains(shop, ".") {
		shop += ".myshopify.com"
	}
	version := strings.Trim(strings.TrimSpace(cfg.Config["api_version"]), "/")
	if version == "" {
		version = shopifyDefaultVersion
	}
	return validBaseURL(shopifyName, "https://"+shop+"/admin/api/"+version, "")
}

func shopifyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(shopifyName, cfg.Config["page_size"], shopifyDefaultPageSize, 1, shopifyMaxPageSize)
}

func shopifyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	return maxPagesConfig(shopifyName, cfg.Config["max_pages"])
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func validBaseURL(name, raw, fallback string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", name, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(name, raw string, fallback, min, max int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s config value must be an integer: %w", name, err)
	}
	if value < min || value > max {
		return 0, fmt.Errorf("%s config value must be between %d and %d", name, min, max)
	}
	return value, nil
}

func maxPagesConfig(name, raw string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer, all, or unlimited: %w", name, err)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s config max_pages must be 0 for unlimited or a positive integer", name)
	}
	return value, nil
}
