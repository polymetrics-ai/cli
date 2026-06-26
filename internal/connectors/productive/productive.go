// Package productive implements a read-only native connector for Productive's
// JSON:API-style REST API.
package productive

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
	defaultBaseURL  = "https://api.productive.io/api/v2"
	defaultPageSize = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("productive", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

var streamPaths = map[string]string{
	"projects":  "projects",
	"people":    "people",
	"companies": "companies",
	"tasks":     "tasks",
}

func (Connector) Name() string { return "productive" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "productive",
		DisplayName:     "Productive",
		IntegrationType: "api",
		Description:     "Reads Productive projects, people, companies, and tasks through the Productive API. Read-only.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "projects", url.Values{"page": []string{"1"}, "per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check productive: %w", err)
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
		stream = "projects"
	}
	path, ok := streamPaths[stream]
	if !ok {
		return fmt.Errorf("productive stream %q not found", stream)
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
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		query := url.Values{"page": []string{strconv.Itoa(page)}, "per_page": []string{strconv.Itoa(pageSize)}}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read productive %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode productive %s: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		totalPages, err := intAt(resp.Body, "meta.total_pages")
		if err != nil {
			return fmt.Errorf("decode productive pagination: %w", err)
		}
		if totalPages == 0 || page >= totalPages {
			return nil
		}
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
	token := secret(cfg, "api_key")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("productive connector requires secret api_key")
	}
	org := strings.TrimSpace(cfg.Config["organization_id"])
	if org == "" {
		return nil, errors.New("productive connector requires config organization_id")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-Auth-Token", token, ""),
		UserAgent: userAgent,
		DefaultHeaders: map[string]string{
			"X-Organization-Id": org,
		},
	}, nil
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "type", Type: "string"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "projects", Description: "Productive projects.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
		{Name: "people", Description: "Productive people.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
		{Name: "companies", Description: "Productive companies.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
		{Name: "tasks", Description: "Productive tasks.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
	}
}

func mapRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{"id": item["id"], "type": item["type"], "raw": item}
	if attrs, ok := item["attributes"].(map[string]any); ok {
		for key, value := range attrs {
			rec[key] = value
		}
	}
	if rec["name"] == nil {
		rec["name"] = first(item, "name", "title")
	}
	return rec
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%d", i), "type": stream, "name": fmt.Sprintf("Fixture %s %d", stream, i), "updated_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}); err != nil {
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

func intAt(body []byte, path string) (int, error) {
	value, err := connsdk.StringAt(body, path)
	if err != nil || strings.TrimSpace(value) == "" {
		return 0, err
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
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
		return "", fmt.Errorf("productive config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("productive config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("productive config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 200 {
		return 0, errors.New("productive config page_size must be between 1 and 200")
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
		return 0, errors.New("productive config max_pages must be 0, all, unlimited, or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
