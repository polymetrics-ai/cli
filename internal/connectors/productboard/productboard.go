// Package productboard implements a conservative read-only connector for the
// Productboard public API. It follows cursor/next-link pagination and avoids all
// mutating endpoints.
package productboard

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
	defaultBaseURL  = "https://api.productboard.com"
	defaultPageSize = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("productboard", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

var streamPaths = map[string]string{
	"features":   "features",
	"notes":      "notes",
	"components": "components",
	"products":   "products",
}

func (Connector) Name() string { return "productboard" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "productboard",
		DisplayName:     "Productboard",
		IntegrationType: "api",
		Description:     "Reads Productboard features, notes, components, and products through the public API. Read-only.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "features", url.Values{"page": []string{"1"}, "limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check productboard: %w", err)
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
		stream = "features"
	}
	path, ok := streamPaths[stream]
	if !ok {
		return fmt.Errorf("productboard stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
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
	next := path
	first := true
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var query url.Values
		if first {
			query = url.Values{"page": []string{"1"}, "limit": []string{strconv.Itoa(pageSize)}}
			if start := strings.TrimSpace(req.Config.Config["start_date"]); start != "" {
				query.Set("updated_since", start)
			}
			first = false
		}
		resp, err := r.Do(ctx, http.MethodGet, next, query, nil)
		if err != nil {
			return fmt.Errorf("read productboard %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode productboard %s: %w", path, err)
		}
		for _, item := range records {
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		nextURL, err := connsdk.StringAt(resp.Body, "links.next")
		if err != nil {
			return fmt.Errorf("decode productboard next: %w", err)
		}
		if strings.TrimSpace(nextURL) == "" {
			return nil
		}
		next = nextURL
	}
	return nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "access_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("productboard connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "features", Description: "Productboard features.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
		{Name: "notes", Description: "Productboard notes.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
		{Name: "components", Description: "Productboard components.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "products", Description: "Productboard products.", Fields: fields, PrimaryKey: []string{"id"}},
	}
}

func mapRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       first(item, "name", "title"),
		"title":      item["title"],
		"status":     item["status"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
		"raw":        item,
	}
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "updated_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}); err != nil {
			return err
		}
	}
	return nil
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
		return "", fmt.Errorf("productboard config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("productboard config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("productboard config base_url must include a host")
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
		return 0, errors.New("productboard config page_size must be between 1 and 100")
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
		return 0, errors.New("productboard config max_pages must be 0, all, unlimited, or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
