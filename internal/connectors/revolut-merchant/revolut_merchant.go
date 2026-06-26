// Package revolutmerchant implements a read-only Revolut Merchant API connector.
package revolutmerchant

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
	defaultBaseURL  = "https://merchant.revolut.com/api/1.0"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("revolut-merchant", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "revolut-merchant" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "revolut-merchant", DisplayName: "Revolut Merchant", IntegrationType: "api", Description: "Reads Revolut Merchant orders, customers, settlements, and payment links through the REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "orders", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check revolut-merchant: %w", err)
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
	ep, ok := endpoints[stream]
	if !ok {
		return fmt.Errorf("revolut-merchant stream %q not found", stream)
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
	base := url.Values{"limit": []string{strconv.Itoa(size)}}
	for _, key := range []string{"from_created_date", "to_created_date", "customer_id", "state"} {
		if value := strings.TrimSpace(req.Config.Config[key]); value != "" {
			base.Set(key, value)
		}
	}
	p := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize: size}
	if err := harvest(ctx, r, ep, base, p, max, func(rec connsdk.Record) error { return emit(mapRecord(stream, rec)) }); err != nil {
		return fmt.Errorf("read revolut-merchant %s: %w", ep.path, err)
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ path, recordsPath string }

var endpoints = map[string]streamEndpoint{
	"orders":        {"orders", ""},
	"customers":     {"customers", ""},
	"settlements":   {"settlements", ""},
	"payment_links": {"payment-links", ""},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "orders", Description: "Revolut Merchant orders.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields("id", "state", "amount", "currency", "created_at")},
		{Name: "customers", Description: "Revolut Merchant customers.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields("id", "email", "full_name", "created_at")},
		{Name: "settlements", Description: "Revolut Merchant settlements.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields("id", "amount", "currency", "created_at")},
		{Name: "payment_links", Description: "Revolut Merchant payment links.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields("id", "state", "amount", "currency", "created_at")},
	}
}

func harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, base url.Values, p connsdk.Paginator, max int, emit func(connsdk.Record) error) error {
	page := p.Start()
	for pageNum := 0; page != nil; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if max > 0 && pageNum >= max {
			return nil
		}
		resp, err := r.Do(ctx, http.MethodGet, ep.path, mergeValues(base, page.Query), nil)
		if err != nil {
			return err
		}
		records, err := recordsAt(resp.Body, ep.recordsPath)
		if err != nil {
			return err
		}
		for _, rec := range records {
			if err := emit(rec); err != nil {
				return err
			}
		}
		page = p.Next(resp, len(records))
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "state": "completed", "amount": 1000 * i, "currency": "USD", "created_at": "2026-01-01T00:00:00Z"}); err != nil {
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
	key := strings.TrimSpace(secret(cfg, "api_key"))
	if key == "" {
		return nil, errors.New("revolut-merchant connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent, DefaultHeaders: map[string]string{"Revolut-Api-Version": "2023-09-01"}}, nil
}

func recordsAt(body []byte, path string) ([]connsdk.Record, error) {
	paths := []string{path, "data", "items", "orders", "customers", "settlements", "payment_links", ""}
	seen := map[string]bool{}
	for _, candidate := range paths {
		if seen[candidate] {
			continue
		}
		seen[candidate] = true
		records, err := connsdk.RecordsAt(body, candidate)
		if err != nil || len(records) > 0 {
			return records, err
		}
	}
	return nil, nil
}

func mapRecord(stream string, rec connsdk.Record) connectors.Record {
	out := connectors.Record{}
	for k, v := range rec {
		out[k] = v
	}
	if out["id"] == nil {
		out["id"] = first(out, "id", "uuid", "email", "reference")
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

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("revolut-merchant config base_url is invalid: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("revolut-merchant config base_url must use http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return "", errors.New("revolut-merchant config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt("revolut-merchant", cfg.Config["page_size"], defaultPageSize, maxPageSize, "page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return optionalInt("revolut-merchant", cfg.Config["max_pages"], "max_pages")
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

func mergeValues(base, extra url.Values) url.Values {
	out := url.Values{}
	for k, vs := range base {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	for k, vs := range extra {
		out.Del(k)
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}
