// Package oura implements a read-only connector for the Oura API v2 user
// collection endpoints. It uses bearer-token auth and next_token pagination.
package oura

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL  = "https://api.ouraring.com/v2/usercollection"
	defaultPageSize = 100
	maxPageSize     = 200
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("oura", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "oura" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "oura", DisplayName: "Oura", IntegrationType: "api", Description: "Reads Oura personal info and daily readiness, sleep, and activity data.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	return r.DoJSON(ctx, http.MethodGet, "/personal_info", nil, nil, nil)
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "oura", Streams: streams()}, nil
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
		stream = "daily_sleep"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("oura stream %q not found", stream)
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
	return harvest(ctx, r, ep, req.Config, size, max, emit)
}

func harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, cfg connectors.RuntimeConfig, size, max int, emit func(connectors.Record) error) error {
	token := ""
	for page := 0; max == 0 || page < max; page++ {
		query := dateQuery(cfg)
		query.Set("page_size", strconv.Itoa(size))
		if token != "" {
			query.Set("next_token", token)
		}
		resp, err := r.Do(ctx, http.MethodGet, ep.path, query, nil)
		if err != nil {
			return fmt.Errorf("read oura %s: %w", ep.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, ep.recordsPath)
		if err != nil {
			return err
		}
		for _, item := range records {
			if err := emit(ep.mapRecord(map[string]any(item))); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_token")
		if err != nil {
			return err
		}
		if strings.TrimSpace(next) == "" || ep.recordsPath == "." {
			return nil
		}
		token = next
	}
	return nil
}

func readFixture(ctx context.Context, stream string, ep streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "day": fmt.Sprintf("2026-01-0%d", i), "score": 80 + i, "timestamp": "2026-01-01T00:00:00Z", "email": fmt.Sprintf("fixture+%d@example.com", i), "age": 40, "weight": 75.5, "height": 180.0, "biological_sex": "unknown"}
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
		return nil, errors.New("oura connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	return validateBaseURL("oura", base)
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

func dateQuery(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	if start := dateOnly(cfg.Config["start_datetime"]); start != "" {
		q.Set("start_date", start)
	}
	if end := dateOnly(cfg.Config["end_datetime"]); end != "" {
		q.Set("end_date", end)
	}
	return q
}

func dateOnly(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.Format("2006-01-02")
	}
	if len(value) >= len("2006-01-02") {
		return value[:len("2006-01-02")]
	}
	return value
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxPageSize {
		return 0, fmt.Errorf("oura config page_size must be between 1 and %d", maxPageSize)
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
		return 0, errors.New("oura config max_pages must be a non-negative integer")
	}
	return n, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamEndpoint struct {
	path        string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"personal_info":   {path: "/personal_info", recordsPath: ".", mapRecord: personalRecord},
	"daily_sleep":     {path: "/daily_sleep", recordsPath: "data", mapRecord: dailyRecord},
	"daily_activity":  {path: "/daily_activity", recordsPath: "data", mapRecord: dailyRecord},
	"daily_readiness": {path: "/daily_readiness", recordsPath: "data", mapRecord: dailyRecord},
}

func streams() []connectors.Stream {
	dailyFields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "day", Type: "string"}, {Name: "score", Type: "integer"}, {Name: "timestamp", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "personal_info", Description: "Oura member profile details.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "email", Type: "string"}, {Name: "age", Type: "integer"}, {Name: "weight", Type: "number"}, {Name: "height", Type: "number"}, {Name: "biological_sex", Type: "string"}}},
		{Name: "daily_sleep", Description: "Oura daily sleep summaries.", PrimaryKey: []string{"id"}, CursorFields: []string{"day"}, Fields: dailyFields},
		{Name: "daily_activity", Description: "Oura daily activity summaries.", PrimaryKey: []string{"id"}, CursorFields: []string{"day"}, Fields: dailyFields},
		{Name: "daily_readiness", Description: "Oura daily readiness summaries.", PrimaryKey: []string{"id"}, CursorFields: []string{"day"}, Fields: dailyFields},
	}
}

func dailyRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "day": item["day"], "score": item["score"], "timestamp": item["timestamp"]}
}

func personalRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "email": item["email"], "age": item["age"], "weight": item["weight"], "height": item["height"], "biological_sex": item["biological_sex"]}
}
