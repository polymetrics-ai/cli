// Package squarespace implements the native pm Squarespace connector.
package squarespace

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
	connectorName   = "squarespace"
	defaultBaseURL  = "https://api.squarespace.com/1.0"
	defaultPageSize = 100
	maxPageSize     = 200
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamEndpoint struct {
	resource    string
	recordsPath string
	fields      []connectors.Field
	mapRecord   func(map[string]any) connectors.Record
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Squarespace", IntegrationType: "api", Description: "Reads Squarespace orders, products, inventory items, and profiles through the Squarespace API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "commerce/orders", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check squarespace: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "orders"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("squarespace stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, emit)
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
	first := url.Values{"limit": []string{strconv.Itoa(pageSize)}}
	paginator := &connsdk.CursorPaginator{CursorParam: "cursor", TokenPath: "pagination.nextPageCursor", FirstQuery: first}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := secret(cfg, "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("squarespace connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("squarespace config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("squarespace config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("squarespace config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "squarespace config page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("squarespace config max_pages must be a non-negative integer: %w", err)
	}
	return value, nil
}

func boundedInt(raw string, def, max int, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s must be between 1 and %d", name, max)
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

func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "orderNumber": fmt.Sprintf("10%02d", i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "createdOn": "2026-01-01T00:00:00Z", "modifiedOn": "2026-01-02T00:00:00Z", "sku": fmt.Sprintf("SKU-%d", i), "quantity": i}
		rec := endpoint.mapRecord(item)
		rec["connector"] = connectorName
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "orders", Description: "Squarespace Commerce orders.", PrimaryKey: []string{"id"}, CursorFields: []string{"modifiedOn"}, Fields: streamEndpoints["orders"].fields},
		{Name: "products", Description: "Squarespace Commerce products.", PrimaryKey: []string{"id"}, CursorFields: []string{"modifiedOn"}, Fields: streamEndpoints["products"].fields},
		{Name: "inventory", Description: "Squarespace Commerce inventory items.", PrimaryKey: []string{"sku"}, Fields: streamEndpoints["inventory"].fields},
		{Name: "profiles", Description: "Squarespace profiles.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["profiles"].fields},
	}
}

var streamEndpoints = map[string]streamEndpoint{
	"orders":    {resource: "commerce/orders", recordsPath: "result", fields: orderFields(), mapRecord: copyRecord("id", "orderNumber", "createdOn", "modifiedOn")},
	"products":  {resource: "commerce/products", recordsPath: "result", fields: productFields(), mapRecord: copyRecord("id", "name", "createdOn", "modifiedOn")},
	"inventory": {resource: "commerce/inventory", recordsPath: "result", fields: inventoryFields(), mapRecord: copyRecord("sku", "quantity", "modifiedOn")},
	"profiles":  {resource: "profiles", recordsPath: "profiles", fields: productFields(), mapRecord: copyRecord("id", "name", "createdOn", "modifiedOn")},
}

func orderFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "orderNumber", Type: "string"}, {Name: "createdOn", Type: "timestamp"}, {Name: "modifiedOn", Type: "timestamp"}}
}

func productFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "createdOn", Type: "timestamp"}, {Name: "modifiedOn", Type: "timestamp"}}
}

func inventoryFields() []connectors.Field {
	return []connectors.Field{{Name: "sku", Type: "string"}, {Name: "quantity", Type: "integer"}, {Name: "modifiedOn", Type: "timestamp"}}
}

func copyRecord(keys ...string) func(map[string]any) connectors.Record {
	return func(item map[string]any) connectors.Record {
		rec := connectors.Record{}
		for _, key := range keys {
			rec[key] = item[key]
		}
		return rec
	}
}
