// Package tplcentral implements a read-only native connector for 3PL Central / Extensiv WMS APIs.
package tplcentral

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
	defaultBaseURL  = "https://api.3plcentral.com"
	defaultPageSize = 100
	maxPageSize     = 1000
	defaultMaxPages = 1
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("tplcentral", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "tplcentral" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "tplcentral",
		DisplayName:     "3PL Central",
		IntegrationType: "api",
		Description:     "Reads 3PL Central / Extensiv WMS customers, orders, items, and shipments through API list endpoints.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "orders", url.Values{"limit": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check tplcentral: %w", err)
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
		stream = "orders"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("tplcentral stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return harvest(ctx, r, req.Config, spec, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func harvest(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, spec streamSpec, emit func(connectors.Record) error) error {
	pageSize, err := pageSize(cfg)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(cfg)
	if err != nil {
		return err
	}
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{"limit": []string{strconv.Itoa(pageSize)}, "page": []string{strconv.Itoa(page)}}
		resp, err := r.Do(ctx, http.MethodGet, spec.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read tplcentral %s: %w", spec.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode tplcentral %s: %w", spec.resource, err)
		}
		for _, item := range records {
			if err := emit(spec.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

func readFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":          fmt.Sprintf("Fixture %s %d", stream, i),
			"reference_num": fmt.Sprintf("REF-%d", i),
			"sku":           fmt.Sprintf("SKU-%d", i),
			"status":        "active",
			"created_at":    "2026-01-01T00:00:00Z",
		}
		rec := spec.mapRecord(item)
		rec["connector"] = "tplcentral"
		rec["fixture"] = true
		if err := emit(rec); err != nil {
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
	token := strings.TrimSpace(cfg.Secrets["access_token"])
	if token == "" {
		return nil, errors.New("tplcentral connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("tplcentral config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("tplcentral config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("tplcentral config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "tplcentral config page_size")
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
	if err != nil || value < 0 {
		return 0, fmt.Errorf("tplcentral config max_pages must be a non-negative integer: %w", err)
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

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamSpec struct {
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamSpecs = map[string]streamSpec{
	"customers": {resource: "customers", recordsPath: "data", mapRecord: customerRecord},
	"orders":    {resource: "orders", recordsPath: "data", mapRecord: orderRecord},
	"items":     {resource: "items", recordsPath: "data", mapRecord: itemRecord},
	"shipments": {resource: "shipments", recordsPath: "data", mapRecord: shipmentRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "customers", Description: "3PL Central customer accounts.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "string"}}},
		{Name: "orders", Description: "3PL Central warehouse orders.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "reference_num", Type: "string"}, {Name: "status", Type: "string"}, {Name: "created_at", Type: "string"}}},
		{Name: "items", Description: "3PL Central item master records.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "sku", Type: "string"}, {Name: "name", Type: "string"}}},
		{Name: "shipments", Description: "3PL Central shipment records.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "order_id", Type: "string"}, {Name: "status", Type: "string"}, {Name: "created_at", Type: "string"}}},
	}
}

func customerRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "customer_id"), "name": first(item, "name", "company_name"), "created_at": first(item, "created_at", "createdAt")}
}

func orderRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "order_id"), "reference_num": first(item, "reference_num", "referenceNum"), "status": item["status"], "created_at": first(item, "created_at", "createdAt")}
}

func itemRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "item_id"), "sku": first(item, "sku", "SKU"), "name": item["name"]}
}

func shipmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "shipment_id"), "order_id": first(item, "order_id", "orderId"), "status": item["status"], "created_at": first(item, "created_at", "createdAt")}
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}
