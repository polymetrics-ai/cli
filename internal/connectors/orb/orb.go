// Package orb implements a read-only connector for Orb's REST API. It uses the
// documented bearer-token auth and cursor pagination used by Orb list endpoints.
package orb

import (
	"context"
	"encoding/json"
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
	defaultBaseURL  = "https://api.withorb.com/v1"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("orb", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "orb" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "orb", DisplayName: "Orb", IntegrationType: "api", Description: "Reads Orb customers, subscriptions, plans, and invoices.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	return r.DoJSON(ctx, http.MethodGet, "/customers", url.Values{"limit": []string{"1"}}, nil, nil)
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "orb", Streams: streams()}, nil
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
		stream = "customers"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("orb stream %q not found", stream)
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
	return harvest(ctx, r, ep, req, size, max, emit)
}

func harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, req connectors.ReadRequest, size, max int, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; max == 0 || page < max; page++ {
		query := url.Values{"limit": []string{strconv.Itoa(size)}}
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		if lower := lowerBound(req); lower != "" {
			query.Set("created_at[gte]", lower)
		}
		resp, err := r.Do(ctx, http.MethodGet, ep.path, query, nil)
		if err != nil {
			return fmt.Errorf("read orb %s: %w", ep.path, err)
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
		next := nextCursor(resp.Body)
		if strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

func nextCursor(body []byte) string {
	var env struct {
		Pagination struct {
			NextCursor string `json:"next_cursor"`
			HasMore    bool   `json:"has_more"`
		} `json:"pagination_metadata"`
	}
	_ = json.Unmarshal(body, &env)
	if !env.Pagination.HasMore {
		return ""
	}
	return env.Pagination.NextCursor
}

func readFixture(ctx context.Context, stream string, ep streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %d", i), "email": fmt.Sprintf("fixture+%d@example.com", i), "status": "active", "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z", "customer_id": "cus_fixture", "plan_id": "plan_fixture", "amount": 1000 * i, "currency": "USD"}
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
	key := secret(cfg, "api_key")
	if key == "" {
		return nil, errors.New("orb connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	return validateBaseURL("orb", base)
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

func lowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxPageSize {
		return 0, fmt.Errorf("orb config page_size must be between 1 and %d", maxPageSize)
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
		return 0, errors.New("orb config max_pages must be a non-negative integer")
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

var streamEndpoints = map[string]streamEndpoint{"customers": {path: "/customers", mapRecord: record}, "subscriptions": {path: "/subscriptions", mapRecord: record}, "plans": {path: "/plans", mapRecord: record}, "invoices": {path: "/invoices", mapRecord: record}}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "status", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}, {Name: "customer_id", Type: "string"}, {Name: "plan_id", Type: "string"}, {Name: "amount", Type: "integer"}, {Name: "currency", Type: "string"}}
	return []connectors.Stream{{Name: "customers", Description: "Orb customers.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields}, {Name: "subscriptions", Description: "Orb subscriptions.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields}, {Name: "plans", Description: "Orb plans.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields}, {Name: "invoices", Description: "Orb invoices.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields}}
}

func record(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "email": item["email"], "status": item["status"], "created_at": item["created_at"], "updated_at": item["updated_at"], "customer_id": item["customer_id"], "plan_id": item["plan_id"], "amount": item["amount"], "currency": item["currency"]}
}
