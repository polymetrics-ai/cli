// Package posthog implements a read-only PostHog REST API connector for the
// modern /api/projects/{project_id}/events/ and /persons/ endpoints.
package posthog

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
	connectorName  = "posthog"
	defaultBaseURL = "https://app.posthog.com"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "PostHog", IntegrationType: "api", Description: "Reads PostHog events and persons via the PostHog REST API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	q := url.Values{"limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, projectPath(cfg, "events"), q, nil, nil); err != nil {
		return fmt.Errorf("check posthog: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "events", Description: "PostHog events for a project.", PrimaryKey: []string{"id"}, CursorFields: []string{"timestamp"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "event", Type: "string"}, {Name: "timestamp", Type: "timestamp"}, {Name: "distinct_id", Type: "string"}, {Name: "properties", Type: "object"}}},
		{Name: "persons", Description: "PostHog persons for a project.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "distinct_id", Type: "string"}, {Name: "properties", Type: "object"}, {Name: "created_at", Type: "timestamp"}}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "events"
	}
	if stream != "events" && stream != "persons" {
		return fmt.Errorf("posthog stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("limit", intConfig(req.Config, "page_size", 100))
	if stream == "events" && strings.TrimSpace(req.Config.Config["start_date"]) != "" {
		q.Set("after", strings.TrimSpace(req.Config.Config["start_date"]))
	}
	path := projectPath(req.Config, stream)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, q, nil)
		if err != nil {
			return fmt.Errorf("read posthog %s: %w", stream, err)
		}
		items, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode posthog %s: %w", stream, err)
		}
		for _, item := range items {
			if err := emit(record(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return err
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path = next
		q = nil
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
	if strings.TrimSpace(cfg.Config["project_id"]) == "" {
		return nil, errors.New("posthog connector requires config project_id")
	}
	token := strings.TrimSpace(cfg.Secrets["api_key"])
	if token == "" {
		return nil, errors.New("posthog connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func projectPath(cfg connectors.RuntimeConfig, stream string) string {
	return "/api/projects/" + url.PathEscape(strings.TrimSpace(cfg.Config["project_id"])) + "/" + stream + "/"
}

func record(item map[string]any) connectors.Record {
	out := connectors.Record{}
	for k, v := range item {
		if n, ok := v.(json.Number); ok {
			if f, err := n.Float64(); err == nil {
				out[k] = f
				continue
			}
		}
		out[k] = v
	}
	return out
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "fixture": true}
		if stream == "events" {
			rec["event"] = "fixture_event"
			rec["timestamp"] = fmt.Sprintf("2026-01-%02dT00:00:00Z", i)
			rec["distinct_id"] = fmt.Sprintf("user_%d", i)
		} else {
			rec["distinct_id"] = fmt.Sprintf("user_%d", i)
			rec["created_at"] = fmt.Sprintf("2026-01-%02dT00:00:00Z", i)
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("posthog config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, def int) string {
	if n, err := strconv.Atoi(strings.TrimSpace(cfg.Config[key])); err == nil && n > 0 {
		return strconv.Itoa(n)
	}
	return strconv.Itoa(def)
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
