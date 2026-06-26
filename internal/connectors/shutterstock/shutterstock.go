// Package shutterstock implements a read-only native Shutterstock connector.
package shutterstock

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
	shutterstockDefaultBaseURL  = "https://api.shutterstock.com"
	shutterstockDefaultPageSize = 100
	shutterstockMaxPageSize     = 500
	shutterstockUserAgent       = "polymetrics-go-cli"
)

func init()                     { connectors.RegisterFactory("shutterstock", New) }
func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "shutterstock" }
func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "shutterstock", DisplayName: "Shutterstock", IntegrationType: "api", Description: "Reads Shutterstock image, video, and audio search metadata through the Shutterstock REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if err := r.DoJSON(ctx, http.MethodGet, shutterstockEndpoints["images"].path, url.Values{"page": []string{"1"}, "per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check shutterstock: %w", err)
	}
	return nil
}
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams(shutterstockEndpoints)}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "images"
	}
	endpoint, ok := shutterstockEndpoints[stream]
	if !ok {
		return fmt.Errorf("shutterstock stream %q not found", stream)
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
	return c.harvest(ctx, r, endpoint, req.Config, pageSize, maxPages, emit)
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		q := filters(cfg)
		q.Set("page", strconv.Itoa(page))
		q.Set("per_page", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, q, nil)
		if err != nil {
			return fmt.Errorf("read shutterstock %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode shutterstock %s: %w", endpoint.path, err)
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
	base, err := baseURL(cfg, shutterstockDefaultBaseURL, "shutterstock")
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(secret(cfg, "access_token"))
	if token == "" {
		return nil, errors.New("shutterstock connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: shutterstockUserAgent}, nil
}

type streamEndpoint struct {
	path, description string
	fields            []string
	mapRecord         func(map[string]any) connectors.Record
}

var shutterstockEndpoints = map[string]streamEndpoint{
	"images": {path: "v2/images/search", description: "Shutterstock image search results.", fields: []string{"id", "description", "media_type", "updated_at"}, mapRecord: shutterstockRecord},
	"videos": {path: "v2/videos/search", description: "Shutterstock video search results.", fields: []string{"id", "description", "media_type", "updated_at"}, mapRecord: shutterstockRecord},
	"audio":  {path: "v2/audio/search", description: "Shutterstock audio search results.", fields: []string{"id", "description", "media_type", "updated_at"}, mapRecord: shutterstockRecord},
}

func shutterstockRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "description": first(item, "description", "title"), "media_type": first(item, "media_type", "asset_type"), "updated_at": first(item, "updated_time", "updated_at")}
}
func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := endpoint.mapRecord(map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "description": fmt.Sprintf("Fixture %s %d", stream, i), "media_type": stream, "updated_time": "2026-01-01T00:00:00Z"})
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}
func streams(endpoints map[string]streamEndpoint) []connectors.Stream {
	order := []string{"images", "videos", "audio"}
	out := make([]connectors.Stream, 0, len(order))
	for _, name := range order {
		ep := endpoints[name]
		out = append(out, connectors.Stream{Name: name, Description: ep.description, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields(ep.fields...)})
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
func filters(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	for _, key := range []string{"query", "sort", "orientation", "category"} {
		if value := strings.TrimSpace(cfg.Config[key]); value != "" {
			q.Set(key, value)
		}
	}
	return q
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
		raw = strings.TrimSpace(cfg.Config["per_page"])
	}
	if raw == "" {
		return shutterstockDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > shutterstockMaxPageSize {
		return 0, fmt.Errorf("shutterstock config page_size must be between 1 and %d", shutterstockMaxPageSize)
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
		return 0, errors.New("shutterstock config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
