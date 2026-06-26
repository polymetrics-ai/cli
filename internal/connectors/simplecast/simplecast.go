// Package simplecast implements a read-only native Simplecast connector.
package simplecast

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
	simplecastDefaultBaseURL  = "https://api.simplecast.com"
	simplecastDefaultPageSize = 100
	simplecastMaxPageSize     = 500
	simplecastUserAgent       = "polymetrics-go-cli"
)

func init()                     { connectors.RegisterFactory("simplecast", New) }
func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "simplecast" }
func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "simplecast", DisplayName: "Simplecast", IntegrationType: "api", Description: "Reads Simplecast podcasts and episodes through the Simplecast REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if err := r.DoJSON(ctx, http.MethodGet, simplecastEndpoints["podcasts"].path, url.Values{"limit": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check simplecast: %w", err)
	}
	return nil
}
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams(simplecastEndpoints)}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "podcasts"
	}
	endpoint, ok := simplecastEndpoints[stream]
	if !ok {
		return fmt.Errorf("simplecast stream %q not found", stream)
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
	nextURL := ""
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := endpoint.path
		var q url.Values
		if nextURL == "" {
			q = url.Values{"page": []string{strconv.Itoa(page)}, "limit": []string{strconv.Itoa(pageSize)}}
		} else {
			path = nextURL
		}
		resp, err := r.Do(ctx, http.MethodGet, path, q, nil)
		if err != nil {
			return fmt.Errorf("read simplecast %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "collection")
		if err != nil {
			return fmt.Errorf("decode simplecast %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		nextURL, err = connsdk.StringAt(resp.Body, "pages.next.href")
		if err != nil {
			return fmt.Errorf("decode simplecast %s pages.next.href: %w", endpoint.path, err)
		}
		if strings.TrimSpace(nextURL) == "" {
			return nil
		}
	}
	return nil
}
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, simplecastDefaultBaseURL, "simplecast")
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(secret(cfg, "access_token"))
	if token == "" {
		return nil, errors.New("simplecast connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: simplecastUserAgent}, nil
}

type streamEndpoint struct {
	path, description string
	fields            []string
	mapRecord         func(map[string]any) connectors.Record
}

var simplecastEndpoints = map[string]streamEndpoint{
	"podcasts": {path: "podcasts", description: "Simplecast podcasts.", fields: []string{"id", "title", "status", "updated_at"}, mapRecord: simplecastRecord},
	"episodes": {path: "episodes", description: "Simplecast episodes.", fields: []string{"id", "title", "status", "updated_at"}, mapRecord: simplecastRecord},
}

func simplecastRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "title": first(item, "title", "name"), "status": item["status"], "updated_at": first(item, "updated_at", "published_at")}
}
func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := endpoint.mapRecord(map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "title": fmt.Sprintf("Fixture %s %d", stream, i), "status": "published", "updated_at": "2026-01-01T00:00:00Z"})
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}
func streams(endpoints map[string]streamEndpoint) []connectors.Stream {
	order := []string{"podcasts", "episodes"}
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
		return simplecastDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > simplecastMaxPageSize {
		return 0, fmt.Errorf("simplecast config page_size must be between 1 and %d", simplecastMaxPageSize)
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
		return 0, errors.New("simplecast config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
