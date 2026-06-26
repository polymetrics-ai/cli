// Package oveit implements a conservative read-only connector for documented
// Oveit API resources. The public Airbyte spec exposes email/password auth; this
// native port sends those credentials as HTTP Basic auth and limits itself to
// read-only collection endpoints.
package oveit

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
	defaultBaseURL  = "https://oveit.com/api"
	defaultPageSize = 100
	maxPageSize     = 500
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("oveit", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "oveit" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "oveit", DisplayName: "Oveit", IntegrationType: "api", Description: "Reads Oveit events, orders, and attendees.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	return r.DoJSON(ctx, http.MethodGet, "/events", url.Values{"per_page": []string{"1"}}, nil, nil)
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "oveit", Streams: streams()}, nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "events"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("oveit stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, ep, req, emit)
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
	return harvest(ctx, r, ep, size, max, emit)
}

func harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, size, max int, emit func(connectors.Record) error) error {
	page := "1"
	for pageNum := 0; max == 0 || pageNum < max; pageNum++ {
		query := url.Values{"page": []string{page}, "per_page": []string{strconv.Itoa(size)}}
		resp, err := r.Do(ctx, http.MethodGet, ep.path, query, nil)
		if err != nil {
			return fmt.Errorf("read oveit %s: %w", ep.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return err
		}
		for _, item := range records {
			if err := emit(ep.mapRecord(map[string]any(item))); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_page")
		if err != nil {
			return err
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		page = next
	}
	return nil
}

func readFixture(ctx context.Context, stream string, ep streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %d", i), "email": fmt.Sprintf("fixture+%d@example.com", i), "status": "active", "starts_at": "2026-01-01T00:00:00Z", "created_at": "2026-01-01T00:00:00Z", "total": 1000 * i}
		rec := ep.mapRecord(item)
		if cursor := req.State[connsdk.CursorStateKey]; cursor != "" {
			rec["previous_cursor"] = cursor
		}
		if err := emit(rec); err != nil {
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
	email := strings.TrimSpace(cfg.Config["email"])
	password := secret(cfg, "password")
	if email == "" || password == "" {
		return nil, errors.New("oveit connector requires config email and secret password")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(email, password), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	return validateBaseURL("oveit", base)
}

func validateBaseURL(name, raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", name, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(raw, "/"), nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Secrets[key])
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxPageSize {
		return 0, fmt.Errorf("oveit config page_size must be between 1 and %d", maxPageSize)
	}
	return n, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.ToLower(strings.TrimSpace(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errors.New("oveit config max_pages must be a non-negative integer")
	}
	return n, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamEndpoint struct {
	path      string
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{"events": {path: "/events", mapRecord: record}, "orders": {path: "/orders", mapRecord: record}, "attendees": {path: "/attendees", mapRecord: record}}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "status", Type: "string"}, {Name: "starts_at", Type: "timestamp"}, {Name: "created_at", Type: "timestamp"}, {Name: "total", Type: "integer"}}
	return []connectors.Stream{{Name: "events", Description: "Oveit events.", PrimaryKey: []string{"id"}, Fields: fields}, {Name: "orders", Description: "Oveit orders.", PrimaryKey: []string{"id"}, Fields: fields}, {Name: "attendees", Description: "Oveit attendees.", PrimaryKey: []string{"id"}, Fields: fields}}
}

func record(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "email": item["email"], "status": item["status"], "starts_at": item["starts_at"], "created_at": item["created_at"], "total": item["total"]}
}
