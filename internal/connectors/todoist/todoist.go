// Package todoist implements a read-only native Go connector for the Todoist REST API.
package todoist

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName  = "todoist"
	defaultBaseURL = "https://api.todoist.com/rest/v2"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Todoist", IntegrationType: "api", Description: "Reads projects, sections, tasks, and comments from the Todoist REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "projects", nil, nil, nil); err != nil {
		return fmt.Errorf("check %s: %w", connectorName, err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "tasks"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("%s stream %q not found", connectorName, stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	q := url.Values{}
	if stream == "comments" {
		if projectID := strings.TrimSpace(req.Config.Config["project_id"]); projectID != "" {
			q.Set("project_id", projectID)
		}
		if taskID := strings.TrimSpace(req.Config.Config["task_id"]); taskID != "" {
			q.Set("task_id", taskID)
		}
	}
	resp, err := r.Do(ctx, http.MethodGet, spec.path, q, nil)
	if err != nil {
		return fmt.Errorf("read %s %s: %w", connectorName, stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, ".")
	if err != nil {
		return err
	}
	for _, rec := range records {
		if err := emit(connectors.Record(rec)); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, _ connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WriteResult{}, err
	}
	return connectors.WriteResult{RecordsFailed: len(records)}, fmt.Errorf("%s connector is read-only: %w", connectorName, connectors.ErrUnsupportedOperation)
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := firstSecret(cfg, "token", "bearer_token")
	if token == "" {
		return nil, errors.New("todoist connector requires secret token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

type streamSpec struct {
	name, description, path string
	fields                  []connectors.Field
}

var streamSpecs = map[string]streamSpec{
	"projects": {"projects", "Todoist projects.", "projects", fields("id", "name", "color", "is_favorite")},
	"sections": {"sections", "Todoist sections.", "sections", fields("id", "project_id", "name", "order")},
	"tasks":    {"tasks", "Todoist tasks.", "tasks", fields("id", "project_id", "section_id", "content", "description", "is_completed", "created_at", "due")},
	"comments": {"comments", "Todoist comments for configured project_id or task_id.", "comments", fields("id", "project_id", "task_id", "content", "posted_at")},
}

func streams() []connectors.Stream {
	order := []string{"projects", "sections", "tasks", "comments"}
	out := make([]connectors.Stream, 0, len(order))
	for _, name := range order {
		s := streamSpecs[name]
		out = append(out, connectors.Stream{Name: s.name, Description: s.description, Fields: s.fields, PrimaryKey: []string{"id"}})
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
func readFixture(ctx context.Context, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", spec.name, i), "content": fmt.Sprintf("Fixture %s %d", spec.name, i), "name": fmt.Sprintf("Fixture %s %d", spec.name, i), "fixture": true}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid %s base_url", connectorName)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("invalid %s base_url scheme %q", connectorName, u.Scheme)
	}
	return strings.TrimRight(raw, "/"), nil
}
func firstSecret(cfg connectors.RuntimeConfig, keys ...string) string {
	for _, key := range keys {
		if cfg.Secrets != nil {
			if value := strings.TrimSpace(cfg.Secrets[key]); value != "" {
				return value
			}
		}
		if value := strings.TrimSpace(cfg.Config[key]); value != "" {
			return value
		}
	}
	return ""
}
