// Package repairshopr implements a read-only RepairShopr API connector.
package repairshopr

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
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("repairshopr", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "repairshopr" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "repairshopr", DisplayName: "RepairShopr", IntegrationType: "api", Description: "Reads RepairShopr customers, tickets, invoices, estimates, and assets through the REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "customers", url.Values{"per_page": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check repairshopr: %w", err)
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
		stream = "customers"
	}
	ep, ok := endpoints[stream]
	if !ok {
		return fmt.Errorf("repairshopr stream %q not found", stream)
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
	max, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{}
	for _, key := range []string{"created_after", "updated_after", "query"} {
		if value := strings.TrimSpace(req.Config.Config[key]); value != "" {
			base.Set(key, value)
		}
	}
	p := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize: size}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, ep.path, base, p, ep.recordsPath, max, func(rec connsdk.Record) error {
		return emit(mapRecord(stream, rec))
	}); err != nil {
		return fmt.Errorf("read repairshopr %s: %w", ep.path, err)
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ path, recordsPath string }

var endpoints = map[string]streamEndpoint{
	"customers": {"customers", "customers"},
	"tickets":   {"tickets", "tickets"},
	"invoices":  {"invoices", "invoices"},
	"estimates": {"estimates", "estimates"},
	"assets":    {"customer_assets", "assets"},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "customers", Description: "RepairShopr customers.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "name", "email", "phone", "created_at", "updated_at")},
		{Name: "tickets", Description: "RepairShopr tickets.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "number", "subject", "status", "customer_id", "updated_at")},
		{Name: "invoices", Description: "RepairShopr invoices.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "number", "customer_id", "total", "status", "updated_at")},
		{Name: "estimates", Description: "RepairShopr estimates.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "number", "customer_id", "total", "status", "updated_at")},
		{Name: "assets", Description: "RepairShopr customer assets.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "customer_id", "name", "serial_number", "updated_at")},
	}
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "status": "active", "updated_at": "2026-01-01T00:00:00Z"}); err != nil {
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
	token := strings.TrimSpace(secret(cfg, "api_token"))
	if token == "" {
		return nil, errors.New("repairshopr connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	if raw := strings.TrimSpace(cfg.Config["base_url"]); raw != "" {
		return validatedBaseURL("repairshopr", raw)
	}
	subdomain := strings.TrimSpace(cfg.Config["subdomain"])
	if subdomain == "" {
		return "", errors.New("repairshopr connector requires config base_url or subdomain")
	}
	return "https://" + url.PathEscape(subdomain) + ".repairshopr.com/api/v1", nil
}

func mapRecord(stream string, rec connsdk.Record) connectors.Record {
	out := connectors.Record{}
	for k, v := range rec {
		out[k] = v
	}
	if out["id"] == nil {
		out["id"] = first(out, "id", "uuid", "number")
	}
	out["stream"] = stream
	return out
}

func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
}

func first(record connectors.Record, keys ...string) any {
	for _, key := range keys {
		if v := record[key]; v != nil {
			return v
		}
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func validatedBaseURL(connector, raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", connector, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", connector, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", connector)
	}
	return strings.TrimRight(raw, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt("repairshopr", cfg.Config["page_size"], defaultPageSize, maxPageSize, "page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return optionalInt("repairshopr", cfg.Config["max_pages"], "max_pages")
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
