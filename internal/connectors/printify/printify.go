// Package printify implements a read-only native connector for Printify's public
// REST API. Mutating product/order endpoints intentionally remain unsupported.
package printify

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
	defaultBaseURL  = "https://api.printify.com/v1"
	defaultPageSize = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("printify", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

type streamSpec struct {
	pathTemplate string
	recordsPath  string
	requiresShop bool
	paginated    bool
}

var streamSpecs = map[string]streamSpec{
	"shops":           {pathTemplate: "shops.json", recordsPath: ""},
	"products":        {pathTemplate: "shops/%s/products.json", recordsPath: "data", requiresShop: true, paginated: true},
	"orders":          {pathTemplate: "shops/%s/orders.json", recordsPath: "data", requiresShop: true, paginated: true},
	"blueprints":      {pathTemplate: "catalog/blueprints.json", recordsPath: ""},
	"print_providers": {pathTemplate: "catalog/print_providers.json", recordsPath: ""},
}

func (Connector) Name() string { return "printify" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "printify",
		DisplayName:     "Printify",
		IntegrationType: "api",
		Description:     "Reads Printify shops, catalog resources, products, and orders through the Printify REST API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
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
	if err := r.DoJSON(ctx, http.MethodGet, "shops.json", nil, nil, nil); err != nil {
		return fmt.Errorf("check printify: %w", err)
	}
	return nil
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
		stream = "shops"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("printify stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}
	path, err := resourcePath(req.Config, spec)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if spec.paginated {
		return c.harvestPaginated(ctx, r, path, spec.recordsPath, req.Config, emit)
	}
	return c.readSinglePage(ctx, r, path, spec.recordsPath, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvestPaginated(ctx context.Context, r *connsdk.Requester, path, recordsPath string, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	pageSize, err := pageSize(cfg)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(cfg)
	if err != nil {
		return err
	}
	next := path
	first := true
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var query url.Values
		if first {
			query = url.Values{"page": []string{"1"}, "limit": []string{strconv.Itoa(pageSize)}}
			first = false
		}
		resp, err := r.Do(ctx, http.MethodGet, next, query, nil)
		if err != nil {
			return fmt.Errorf("read printify %s: %w", path, err)
		}
		if err := emitRecords(ctx, resp.Body, recordsPath, emit); err != nil {
			return err
		}
		nextURL, err := connsdk.StringAt(resp.Body, "next_page_url")
		if err != nil {
			return fmt.Errorf("decode printify next_page_url: %w", err)
		}
		if strings.TrimSpace(nextURL) == "" {
			return nil
		}
		next = nextURL
	}
	return nil
}

func (c Connector) readSinglePage(ctx context.Context, r *connsdk.Requester, path, recordsPath string, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read printify %s: %w", path, err)
	}
	return emitRecords(ctx, resp.Body, recordsPath, emit)
}

func emitRecords(ctx context.Context, body []byte, recordsPath string, emit func(connectors.Record) error) error {
	records, err := connsdk.RecordsAt(body, recordsPath)
	if err != nil {
		return fmt.Errorf("decode printify records: %w", err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(mapRecord(item)); err != nil {
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
	token := secret(cfg, "api_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("printify connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "status", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "shops", Description: "Printify shops.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "products", Description: "Products in a configured Printify shop.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
		{Name: "orders", Description: "Orders in a configured Printify shop.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
		{Name: "blueprints", Description: "Printify catalog blueprints.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "print_providers", Description: "Printify catalog print providers.", Fields: fields, PrimaryKey: []string{"id"}},
	}
}

func mapRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"title":         first(item, "title", "name"),
		"sales_channel": item["sales_channel"],
		"status":        item["status"],
		"visible":       item["visible"],
		"created_at":    item["created_at"],
		"updated_at":    item["updated_at"],
		"raw":           item,
	}
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": i, "title": fmt.Sprintf("Fixture %s %d", stream, i), "status": "fixture", "updated_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}); err != nil {
			return err
		}
	}
	return nil
}

func resourcePath(cfg connectors.RuntimeConfig, spec streamSpec) (string, error) {
	if !spec.requiresShop {
		return spec.pathTemplate, nil
	}
	shopID := strings.TrimSpace(cfg.Config["shop_id"])
	if shopID == "" {
		return "", errors.New("printify connector requires config shop_id for this stream")
	}
	return fmt.Sprintf(spec.pathTemplate, url.PathEscape(shopID)), nil
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("printify config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("printify config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("printify config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 100 {
		return 0, errors.New("printify config page_size must be between 1 and 100")
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("printify config max_pages must be 0, all, unlimited, or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
