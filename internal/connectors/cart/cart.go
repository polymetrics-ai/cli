package cart

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
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("cart", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

var resources = map[string]string{"orders": "orders", "customers": "customers", "products": "products"}

func (Connector) Name() string { return "cart" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "cart", DisplayName: "Cart.com", IntegrationType: "api", Description: "Reads Cart.com orders, customers, and products through a read-only REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "orders", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check cart: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "order_number", Type: "string"}, {Name: "updated_at", Type: "timestamp"}}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{
		{Name: "orders", Description: "Cart.com orders.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "customers", Description: "Cart.com customers.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "products", Description: "Cart.com products.", Fields: fields, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "orders"
	}
	resource, ok := resources[stream]
	if !ok {
		return fmt.Errorf("cart stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := intConfig(req.Config, "page_size", defaultPageSize)
	if err != nil {
		return err
	}
	maxPages, err := intConfig(req.Config, "max_pages", defaultMaxPages)
	if err != nil {
		return err
	}
	return readPaged(ctx, r, resource, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readPaged(ctx context.Context, r *connsdk.Requester, resource string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	page := "1"
	for i := 0; i < maxPages; i++ {
		query := url.Values{"page": []string{page}, "page_size": []string{strconv.Itoa(pageSize)}}
		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read cart %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode cart %s: %w", resource, err)
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "meta.next_page")
		if err != nil || strings.TrimSpace(next) == "" {
			return err
		}
		page = next
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "order_number": fmt.Sprintf("F-%04d", i), "updated_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i), "fixture": true}); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "access_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("cart connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		store := strings.TrimSpace(cfg.Config["store_name"])
		if store == "" {
			return "", errors.New("cart connector requires config base_url or store_name")
		}
		base = "https://" + store + "/api/v1"
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("cart config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("cart config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("cart config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("cart config %s must be a positive integer", key)
	}
	return value, nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
