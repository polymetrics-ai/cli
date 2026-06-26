// Package zenloop implements a conservative read-only Zenloop API connector.
// It covers the common list endpoints with page/per_page pagination; write and
// mutation endpoints are intentionally not exposed.
package zenloop

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
	connectorName  = "zenloop"
	defaultBaseURL = "https://api.zenloop.com/v1"
	userAgent      = "polymetrics-go-cli"
)

var paths = map[string]string{"answers": "/answers", "surveys": "/surveys", "survey_groups": "/survey_groups"}

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Zenloop", IntegrationType: "api", Description: "Reads Zenloop answers, surveys, and survey groups. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	q := url.Values{"page": []string{"1"}, "per_page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, paths["answers"], q, nil, nil); err != nil {
		return fmt.Errorf("check zenloop: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "score", Type: "number"}, {Name: "comment", Type: "string"}, {Name: "inserted_at", Type: "timestamp"}, {Name: "survey_id", Type: "string"}}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "answers", Description: "Zenloop answers.", PrimaryKey: []string{"id"}, CursorFields: []string{"inserted_at"}, Fields: fields},
		{Name: "surveys", Description: "Zenloop surveys.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "timestamp"}}},
		{Name: "survey_groups", Description: "Zenloop survey groups.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "answers"
	}
	path, ok := paths[stream]
	if !ok {
		return fmt.Errorf("zenloop stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize := intConfig(req.Config.Config["page_size"], 100)
	for page := 1; ; page++ {
		q := baseQuery(req.Config)
		q.Set("page", strconv.Itoa(page))
		q.Set("per_page", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, path, q, nil)
		if err != nil {
			return fmt.Errorf("read zenloop %s: %w", stream, err)
		}
		items, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(record(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "meta.next_page")
		if err != nil {
			return err
		}
		if strings.TrimSpace(next) == "" || len(items) == 0 {
			return nil
		}
		if n, err := strconv.Atoi(next); err == nil {
			page = n - 1
		}
	}
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(cfg.Secrets["api_token"])
	if token == "" {
		return nil, errors.New("zenloop connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func baseQuery(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	for _, key := range []string{"date_from", "survey_id", "survey_group_id"} {
		if v := first(cfg.Config[key], cfg.Secrets[key]); v != "" {
			q.Set(key, v)
		}
	}
	return q
}

func record(item map[string]any) connectors.Record {
	out := connectors.Record{}
	for k, v := range item {
		if k == "survey" {
			if m, ok := v.(map[string]any); ok {
				out["survey_id"] = m["id"]
			}
		}
		out[k] = normalize(v)
	}
	return out
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "fixture": true}
		if stream == "answers" {
			rec["score"] = float64(10 - i)
			rec["comment"] = "fixture"
			rec["inserted_at"] = fmt.Sprintf("2026-01-%02dT00:00:00Z", i)
			rec["survey_id"] = "survey_fixture"
		} else {
			rec["name"] = fmt.Sprintf("Fixture %d", i)
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func normalize(v any) any {
	if n, ok := v.(json.Number); ok {
		if f, err := n.Float64(); err == nil {
			return f
		}
	}
	return v
}

func intConfig(raw string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil && n > 0 {
		return n
	}
	return def
}

func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("zenloop config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
