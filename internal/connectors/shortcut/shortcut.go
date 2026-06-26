// Package shortcut implements a read-only native Shortcut connector.
package shortcut

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
	shortcutDefaultBaseURL  = "https://api.app.shortcut.com"
	shortcutDefaultPageSize = 100
	shortcutMaxPageSize     = 500
	shortcutUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("shortcut", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "shortcut" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "shortcut", DisplayName: "Shortcut", IntegrationType: "api", Description: "Reads Shortcut stories, epics, projects, and iterations through the Shortcut REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if err := r.DoJSON(ctx, http.MethodGet, shortcutEndpoints["stories"].path, url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check shortcut: %w", err)
	}
	return nil
}
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams(shortcutEndpoints)}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "stories"
	}
	endpoint, ok := shortcutEndpoints[stream]
	if !ok {
		return fmt.Errorf("shortcut stream %q not found", stream)
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
	next := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		q := url.Values{"limit": []string{strconv.Itoa(pageSize)}}
		if next != "" {
			q.Set("next", next)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, q, nil)
		if err != nil {
			return fmt.Errorf("read shortcut %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode shortcut %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err = connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode shortcut %s next: %w", endpoint.path, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, shortcutDefaultBaseURL, "shortcut")
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(secret(cfg, "api_token"))
	if token == "" {
		return nil, errors.New("shortcut connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Shortcut-Token", token, ""), UserAgent: shortcutUserAgent}, nil
}

type streamEndpoint struct {
	path, recordsPath, description string
	fields                         []string
	mapRecord                      func(map[string]any) connectors.Record
}

var shortcutEndpoints = map[string]streamEndpoint{
	"stories":    {path: "api/v3/stories", recordsPath: "data", description: "Shortcut stories.", fields: []string{"id", "name", "state", "updated_at"}, mapRecord: shortcutRecord},
	"epics":      {path: "api/v3/epics", recordsPath: "data", description: "Shortcut epics.", fields: []string{"id", "name", "state", "updated_at"}, mapRecord: shortcutRecord},
	"projects":   {path: "api/v3/projects", recordsPath: "data", description: "Shortcut projects.", fields: []string{"id", "name", "state", "updated_at"}, mapRecord: shortcutRecord},
	"iterations": {path: "api/v3/iterations", recordsPath: "data", description: "Shortcut iterations.", fields: []string{"id", "name", "state", "updated_at"}, mapRecord: shortcutRecord},
}

func shortcutRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "state": first(item, "state", "workflow_state_id"), "updated_at": first(item, "updated_at", "updated_at_override")}
}
func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := endpoint.mapRecord(map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "state": "started", "updated_at": "2026-01-01T00:00:00Z"})
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}
func streams(endpoints map[string]streamEndpoint) []connectors.Stream {
	order := []string{"stories", "epics", "projects", "iterations"}
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
		return shortcutDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > shortcutMaxPageSize {
		return 0, fmt.Errorf("shortcut config page_size must be between 1 and %d", shortcutMaxPageSize)
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
		return 0, errors.New("shortcut config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
