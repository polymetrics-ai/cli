// Package shipstation implements a read-only native ShipStation connector.
package shipstation

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
	shipstationDefaultBaseURL  = "https://ssapi.shipstation.com"
	shipstationDefaultPageSize = 100
	shipstationMaxPageSize     = 500
	shipstationUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("shipstation", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "shipstation" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "shipstation", DisplayName: "ShipStation", IntegrationType: "api", Description: "Reads ShipStation orders, shipments, products, and customers through the ShipStation REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	q := url.Values{"pageSize": []string{"1"}, "page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, shipstationEndpoints["orders"].path, q, nil, nil); err != nil {
		return fmt.Errorf("check shipstation: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams(shipstationEndpoints)}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "orders"
	}
	endpoint, ok := shipstationEndpoints[stream]
	if !ok {
		return fmt.Errorf("shipstation stream %q not found", stream)
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
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		q := url.Values{"page": []string{strconv.Itoa(page)}, "pageSize": []string{strconv.Itoa(pageSize)}}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, q, nil)
		if err != nil {
			return fmt.Errorf("read shipstation %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode shipstation %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, shipstationDefaultBaseURL, "shipstation")
	if err != nil {
		return nil, err
	}
	key := strings.TrimSpace(secret(cfg, "api_key"))
	apiSecret := strings.TrimSpace(secret(cfg, "api_secret"))
	if key == "" || apiSecret == "" {
		return nil, errors.New("shipstation connector requires secrets api_key and api_secret")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(key, apiSecret), UserAgent: shipstationUserAgent}, nil
}

type streamEndpoint struct {
	path        string
	recordsPath string
	description string
	fields      []string
	mapRecord   func(map[string]any) connectors.Record
}

var shipstationEndpoints = map[string]streamEndpoint{
	"orders":    {path: "orders", recordsPath: "orders", description: "ShipStation orders.", fields: []string{"id", "order_number", "status", "modified_at"}, mapRecord: orderRecord},
	"shipments": {path: "shipments", recordsPath: "shipments", description: "ShipStation shipments.", fields: []string{"id", "order_number", "status", "modified_at"}, mapRecord: shipmentRecord},
	"products":  {path: "products", recordsPath: "products", description: "ShipStation products.", fields: []string{"id", "sku", "name", "modified_at"}, mapRecord: productRecord},
	"customers": {path: "customers", recordsPath: "customers", description: "ShipStation customers.", fields: []string{"id", "name", "email", "modified_at"}, mapRecord: customerRecord},
}

func orderRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "orderId", "id"), "order_number": item["orderNumber"], "status": item["orderStatus"], "modified_at": first(item, "modifyDate", "updated_at")}
}
func shipmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "shipmentId", "id"), "order_number": item["orderNumber"], "status": item["shipmentStatus"], "modified_at": first(item, "modifyDate", "shipDate")}
}
func productRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "productId", "id"), "sku": item["sku"], "name": first(item, "name", "productName"), "modified_at": first(item, "modifyDate", "updated_at")}
}
func customerRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "customerId", "id"), "name": first(item, "name", "customerName"), "email": item["email"], "modified_at": first(item, "modifyDate", "updated_at")}
}

func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "orderId": fmt.Sprintf("%s_fixture_%d", stream, i), "orderNumber": fmt.Sprintf("FIX-%d", i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "modifyDate": "2026-01-01T00:00:00Z"}
		rec := endpoint.mapRecord(item)
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func streams(endpoints map[string]streamEndpoint) []connectors.Stream {
	order := []string{"orders", "shipments", "products", "customers"}
	out := make([]connectors.Stream, 0, len(order))
	for _, name := range order {
		ep := endpoints[name]
		out = append(out, connectors.Stream{Name: name, Description: ep.description, PrimaryKey: []string{"id"}, CursorFields: []string{"modified_at"}, Fields: fields(ep.fields...)})
	}
	return out
}

func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
}
func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
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

func baseURL(cfg connectors.RuntimeConfig, fallback, name string) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%s config base_url must use http or https", name)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["limit"])
	}
	if raw == "" {
		return shipstationDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > shipstationMaxPageSize {
		return 0, fmt.Errorf("shipstation config page_size must be between 1 and %d", shipstationMaxPageSize)
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
		return 0, errors.New("shipstation config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
