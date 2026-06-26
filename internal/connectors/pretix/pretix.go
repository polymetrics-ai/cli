// Package pretix implements a read-only native connector for the pretix REST API.
package pretix

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
	defaultBaseURL  = "https://pretix.eu/api/v1"
	defaultPageSize = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("pretix", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

type streamSpec struct {
	pathTemplate  string
	requiresOrg   bool
	requiresEvent bool
}

var streamSpecs = map[string]streamSpec{
	"organizers": {pathTemplate: "organizers/"},
	"events":     {pathTemplate: "organizers/%s/events/", requiresOrg: true},
	"items":      {pathTemplate: "organizers/%s/events/%s/items/", requiresOrg: true, requiresEvent: true},
	"orders":     {pathTemplate: "organizers/%s/events/%s/orders/", requiresOrg: true, requiresEvent: true},
}

func (Connector) Name() string { return "pretix" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "pretix",
		DisplayName:     "Pretix",
		IntegrationType: "api",
		Description:     "Reads pretix organizers, events, items, and orders through the pretix REST API. Read-only.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "organizers/", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check pretix: %w", err)
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
		stream = "events"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("pretix stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}
	path, err := resourcePath(req.Config, spec)
	if err != nil {
		return err
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
	return c.harvest(ctx, r, path, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	next := path
	first := true
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var query url.Values
		if first {
			query = url.Values{"page": []string{"1"}, "page_size": []string{strconv.Itoa(pageSize)}}
			first = false
		}
		resp, err := r.Do(ctx, http.MethodGet, next, query, nil)
		if err != nil {
			return fmt.Errorf("read pretix %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode pretix %s: %w", path, err)
		}
		for _, item := range records {
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		nextURL, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode pretix next: %w", err)
		}
		if strings.TrimSpace(nextURL) == "" {
			return nil
		}
		next = nextURL
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "api_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("pretix connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", token, "Token "), UserAgent: userAgent}, nil
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "slug", Type: "string"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "organizers", Description: "pretix organizers.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "events", Description: "pretix events for a configured organizer.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "items", Description: "pretix ticket items for a configured organizer and event.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "orders", Description: "pretix orders for a configured organizer and event.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}},
	}
}

func mapRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         first(item, "slug", "code", "id", "ID"),
		"slug":       item["slug"],
		"code":       item["code"],
		"name":       item["name"],
		"created_at": first(item, "created", "created_at"),
		"updated_at": first(item, "modified", "updated_at", "date_from"),
		"raw":        item,
	}
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s-%d", stream, i), "slug": fmt.Sprintf("fixture-%d", i), "name": map[string]any{"en": fmt.Sprintf("Fixture %s %d", stream, i)}, "updated_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}); err != nil {
			return err
		}
	}
	return nil
}

func resourcePath(cfg connectors.RuntimeConfig, spec streamSpec) (string, error) {
	if spec.requiresOrg {
		org := strings.TrimSpace(cfg.Config["organizer"])
		if org == "" {
			return "", errors.New("pretix connector requires config organizer for this stream")
		}
		if spec.requiresEvent {
			event := strings.TrimSpace(cfg.Config["event"])
			if event == "" {
				return "", errors.New("pretix connector requires config event for this stream")
			}
			return fmt.Sprintf(spec.pathTemplate, url.PathEscape(org), url.PathEscape(event)), nil
		}
		return fmt.Sprintf(spec.pathTemplate, url.PathEscape(org)), nil
	}
	return spec.pathTemplate, nil
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
		return "", fmt.Errorf("pretix config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("pretix config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("pretix config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 500 {
		return 0, errors.New("pretix config page_size must be between 1 and 500")
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
		return 0, errors.New("pretix config max_pages must be 0, all, unlimited, or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
