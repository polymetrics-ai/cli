// Package toggl implements a read-only native Go connector for the Toggl Track API.
package toggl

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
	connectorName  = "toggl"
	defaultBaseURL = "https://api.track.toggl.com/api/v9"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Toggl", IntegrationType: "api", Description: "Reads time entries, projects, clients, and users from the Toggl Track API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "me", nil, nil, nil); err != nil {
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
		stream = "time_entries"
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
	path, err := spec.path(req.Config)
	if err != nil {
		return err
	}
	q := url.Values{}
	if stream == "time_entries" {
		if start := strings.TrimSpace(req.Config.Config["start_date"]); start != "" {
			q.Set("start_date", start)
		}
		if end := strings.TrimSpace(req.Config.Config["end_date"]); end != "" {
			q.Set("end_date", end)
		}
	}
	resp, err := r.Do(ctx, http.MethodGet, path, q, nil)
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
	token := secret(cfg, "api_token")
	if token == "" {
		return nil, errors.New("toggl connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(token, "api_token"), UserAgent: userAgent}, nil
}

type streamSpec struct {
	name, description string
	path              func(connectors.RuntimeConfig) (string, error)
	fields            []connectors.Field
}

var streamSpecs = map[string]streamSpec{
	"time_entries":       {"time_entries", "Toggl time entries for the authenticated user.", staticPath("me/time_entries"), fields("id", "workspace_id", "project_id", "description", "start", "stop", "duration")},
	"projects":           {"projects", "Toggl workspace projects.", workspacePath("projects"), fields("id", "workspace_id", "client_id", "name", "active")},
	"clients":            {"clients", "Toggl workspace clients.", workspacePath("clients"), fields("id", "workspace_id", "name", "active")},
	"workspace_users":    {"workspace_users", "Toggl workspace users.", workspacePath("users"), fields("id", "name", "email", "active")},
	"organization_users": {"organization_users", "Toggl organization users.", organizationPath("users"), fields("id", "name", "email", "active")},
}

func streams() []connectors.Stream {
	order := []string{"time_entries", "projects", "clients", "workspace_users", "organization_users"}
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
func staticPath(path string) func(connectors.RuntimeConfig) (string, error) {
	return func(connectors.RuntimeConfig) (string, error) { return path, nil }
}
func workspacePath(resource string) func(connectors.RuntimeConfig) (string, error) {
	return func(cfg connectors.RuntimeConfig) (string, error) {
		id := strings.TrimSpace(cfg.Config["workspace_id"])
		if id == "" {
			return "", errors.New("toggl connector requires config workspace_id")
		}
		return "workspaces/" + url.PathEscape(id) + "/" + resource, nil
	}
}
func organizationPath(resource string) func(connectors.RuntimeConfig) (string, error) {
	return func(cfg connectors.RuntimeConfig) (string, error) {
		id := strings.TrimSpace(cfg.Config["organization_id"])
		if id == "" {
			return "", errors.New("toggl connector requires config organization_id")
		}
		return "organizations/" + url.PathEscape(id) + "/" + resource, nil
	}
}
func readFixture(ctx context.Context, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", spec.name, i), "name": fmt.Sprintf("Fixture %s %d", spec.name, i), "description": fmt.Sprintf("Fixture %s %d", spec.name, i), "workspace_id": "fixture_workspace", "fixture": true}
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
func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets != nil {
		if value := strings.TrimSpace(cfg.Secrets[key]); value != "" {
			return value
		}
	}
	return strings.TrimSpace(cfg.Config[key])
}
