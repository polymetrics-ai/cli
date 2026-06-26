// Package picqer implements a read-only Picqer WMS connector. Picqer uses HTTP
// Basic auth with the API key as username and offset pagination for list reads.
package picqer

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
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli (polymetrics.ai)"
)

func init()                     { connectors.RegisterFactory("picqer", New) }
func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "picqer" }
func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "picqer", DisplayName: "Picqer", IntegrationType: "api", Description: "Reads Picqer products, customers, orders, picklists, and warehouses through the Picqer REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "products", url.Values{"offset": []string{"0"}}, nil, nil); err != nil {
		return fmt.Errorf("check picqer: %w", err)
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
		stream = "products"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("picqer stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	p := &connsdk.OffsetPaginator{OffsetParam: "offset", PageSize: size}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, p, "", maxPages, func(rec connsdk.Record) error { return emit(mapRecord(rec)) })
}
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ resource string }

var streamEndpoints = map[string]streamEndpoint{"products": {"products"}, "customers": {"customers"}, "orders": {"orders"}, "picklists": {"picklists"}, "warehouses": {"warehouses"}, "suppliers": {"suppliers"}}

func streams() []connectors.Stream {
	names := []string{"products", "customers", "orders", "picklists", "warehouses", "suppliers"}
	out := make([]connectors.Stream, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Stream{Name: name, Description: "Picqer " + name + " list endpoint.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "created", Type: "string"}, {Name: "updated", Type: "string"}}})
	}
	return out
}
func mapRecord(rec connsdk.Record) connectors.Record {
	out := connectors.Record(rec)
	if out["id"] == nil {
		for _, k := range []string{"idproduct", "idcustomer", "idorder", "idpicklist", "idwarehouse", "idsupplier"} {
			if out[k] != nil {
				out["id"] = out[k]
				break
			}
		}
	}
	return out
}
func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": i, "name": fmt.Sprintf("Fixture %s %d", stream, i), "updated": "2026-01-01T00:00:00Z"}); err != nil {
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
	key := firstNonEmpty(secret(cfg, "api_key"), cfg.Config["username"])
	password := secret(cfg, "password")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("picqer connector requires secret api_key or config username")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(key, password), UserAgent: userAgent}, nil
}
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	if raw := strings.TrimSpace(cfg.Config["base_url"]); raw != "" {
		return validatedBaseURL("picqer", raw, "")
	}
	org := strings.TrimSpace(cfg.Config["organization_name"])
	if org == "" {
		return "", errors.New("picqer connector requires config organization_name or base_url")
	}
	return "https://" + url.PathEscape(org) + ".picqer.com/api/v1", nil
}
func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt("picqer", cfg.Config["page_size"], defaultPageSize, maxPageSize, "page_size")
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return optionalInt("picqer", cfg.Config["max_pages"], "max_pages")
}
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
func validatedBaseURL(connector, raw, def string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		return def, nil
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", connector, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", connector, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", connector)
	}
	return strings.TrimRight(base, "/"), nil
}
func boundedInt(connector, raw string, def, max int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v < 1 || v > max {
		return 0, fmt.Errorf("%s config %s must be an integer between 1 and %d", connector, name, max)
	}
	return v, nil
}
func optionalInt(connector, raw, name string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, fmt.Errorf("%s config %s must be 0, positive, all, or unlimited", connector, name)
	}
	return v, nil
}
