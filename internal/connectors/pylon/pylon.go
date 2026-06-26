// Package pylon implements a read-only native connector for Pylon's public API.
package pylon

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
	defaultBaseURL  = "https://api.usepylon.com"
	defaultPageSize = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("pylon", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

var streamPaths = map[string]string{
	"issues":   "issues",
	"accounts": "accounts",
	"contacts": "contacts",
	"users":    "users",
	"messages": "messages",
}

func (Connector) Name() string { return "pylon" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "pylon",
		DisplayName:     "Pylon",
		IntegrationType: "api",
		Description:     "Reads Pylon issues, accounts, contacts, users, and messages through the Pylon API. Read-only.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "issues", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check pylon: %w", err)
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
		stream = "issues"
	}
	path, ok := streamPaths[stream]
	if !ok {
		return fmt.Errorf("pylon stream %q not found", stream)
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
	query := url.Values{"limit": []string{strconv.Itoa(pageSize)}}
	if start := strings.TrimSpace(req.Config.Config["start_date"]); start != "" {
		query.Set("updated_after", start)
	}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read pylon %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode pylon %s: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		nextCursor, err := connsdk.StringAt(resp.Body, "pagination.next_cursor")
		if err != nil {
			return fmt.Errorf("decode pylon next_cursor: %w", err)
		}
		if strings.TrimSpace(nextCursor) == "" {
			return nil
		}
		query.Set("cursor", nextCursor)
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
	token := secret(cfg, "api_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("pylon connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "name", Type: "string"}, {Name: "state", Type: "string"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "issues", Description: "Pylon issues.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
		{Name: "accounts", Description: "Pylon accounts.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
		{Name: "contacts", Description: "Pylon contacts.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
		{Name: "users", Description: "Pylon users.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "messages", Description: "Pylon messages.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
	}
}

func mapRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"title":      item["title"],
		"name":       first(item, "name", "full_name", "email"),
		"state":      first(item, "state", "status"),
		"email":      item["email"],
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
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_%d", stream, i), "title": fmt.Sprintf("Fixture %s %d", stream, i), "state": "open", "updated_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}); err != nil {
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
		return "", fmt.Errorf("pylon config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("pylon config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("pylon config base_url must include a host")
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
		return 0, errors.New("pylon config page_size must be between 1 and 200")
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
		return 0, errors.New("pylon config max_pages must be 0, all, unlimited, or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
