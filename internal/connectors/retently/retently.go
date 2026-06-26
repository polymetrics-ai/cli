// Package retently implements a read-only Retently API connector.
package retently

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
	defaultBaseURL  = "https://app.retently.com/api/v2"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("retently", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "retently" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "retently", DisplayName: "Retently", IntegrationType: "api", Description: "Reads Retently customers, responses, surveys, and campaigns through the REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "customers", url.Values{"limit": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check retently: %w", err)
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
		return fmt.Errorf("retently stream %q not found", stream)
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
	for _, key := range []string{"updated_after", "created_after", "email", "campaign_id"} {
		if value := strings.TrimSpace(req.Config.Config[key]); value != "" {
			base.Set(key, value)
		}
	}
	p := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize: size}
	if err := harvest(ctx, r, ep, base, p, max, func(rec connsdk.Record) error { return emit(mapRecord(stream, rec)) }); err != nil {
		return fmt.Errorf("read retently %s: %w", ep.path, err)
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ path, recordsPath string }

var endpoints = map[string]streamEndpoint{
	"customers": {"customers", "data"},
	"responses": {"responses", "data"},
	"surveys":   {"surveys", "data"},
	"campaigns": {"campaigns", "data"},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "customers", Description: "Retently customers.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "email", "full_name", "company", "updated_at")},
		{Name: "responses", Description: "Retently survey responses.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields("id", "customer_id", "score", "comment", "created_at")},
		{Name: "surveys", Description: "Retently surveys.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "name", "type", "status", "updated_at")},
		{Name: "campaigns", Description: "Retently campaigns.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "name", "status", "updated_at")},
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
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "email": fmt.Sprintf("fixture+%d@example.com", i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "score": 10 - i, "updated_at": "2026-01-01T00:00:00Z"}); err != nil {
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
		return nil, errors.New("retently connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", key, "api_key=")}, nil
}

func recordsAt(body []byte, path string) ([]connsdk.Record, error) {
	paths := []string{path, "data", "items", "records", "results", ""}
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
		out["id"] = first(out, "id", "uuid", "email", "name")
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
		return "", fmt.Errorf("retently config base_url is invalid: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("retently config base_url must use http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return "", errors.New("retently config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt("retently", cfg.Config["page_size"], defaultPageSize, maxPageSize, "page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return optionalInt("retently", cfg.Config["max_pages"], "max_pages")
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
