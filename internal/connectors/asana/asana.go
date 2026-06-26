// Package asana implements a read-only Asana API connector.
package asana

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
	asanaName            = "asana"
	asanaDefaultBaseURL  = "https://app.asana.com/api/1.0"
	asanaDefaultPageSize = 100
	asanaMaxPageSize     = 100
	asanaUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("asana", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return asanaName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: asanaName, DisplayName: "Asana", IntegrationType: "api", Description: "Reads Asana workspaces, projects, and tasks through the Asana v1 REST API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if strings.TrimSpace(asanaAccessToken(cfg)) == "" {
		return errors.New("asana connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "users/me", nil, nil, nil); err != nil {
		return fmt.Errorf("check asana: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: asanaStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "projects"
	}
	endpoint, ok := asanaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("asana stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := asanaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := asanaMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, req.Config, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint asanaStreamEndpoint, cfg connectors.RuntimeConfig, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := endpoint.resource
	query := asanaQuery(cfg, endpoint, pageSize)
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read asana %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode asana %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_page.uri")
		if err != nil {
			return fmt.Errorf("decode asana next_page.uri: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path = next
		query = nil
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint asanaStreamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"gid": fmt.Sprintf("fixture-%d", i), "name": fmt.Sprintf("Fixture %s %d", endpoint.resource, i), "resource_type": strings.TrimSuffix(endpoint.resource, "s"), "created_at": "2026-01-01T00:00:00Z", "modified_at": "2026-01-02T00:00:00Z", "completed": false}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := asanaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := asanaAccessToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("asana connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: asanaUserAgent}, nil
}

type asanaStreamEndpoint struct {
	resource  string
	optFields string
	mapRecord func(map[string]any) connectors.Record
}

var asanaStreamEndpoints = map[string]asanaStreamEndpoint{
	"workspaces": {resource: "workspaces", optFields: "gid,name,resource_type", mapRecord: asanaWorkspaceRecord},
	"projects":   {resource: "projects", optFields: "gid,name,resource_type,created_at,modified_at", mapRecord: asanaProjectRecord},
	"tasks":      {resource: "tasks", optFields: "gid,name,resource_type,created_at,modified_at,completed", mapRecord: asanaTaskRecord},
}

func asanaStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "workspaces", Description: "Asana workspaces.", PrimaryKey: []string{"gid"}, Fields: []connectors.Field{{Name: "gid", Type: "string"}, {Name: "name", Type: "string"}, {Name: "resource_type", Type: "string"}}},
		{Name: "projects", Description: "Asana projects.", PrimaryKey: []string{"gid"}, Fields: asanaObjectFields(false)},
		{Name: "tasks", Description: "Asana tasks.", PrimaryKey: []string{"gid"}, Fields: asanaObjectFields(true)},
	}
}

func asanaObjectFields(withCompleted bool) []connectors.Field {
	fields := []connectors.Field{{Name: "gid", Type: "string"}, {Name: "name", Type: "string"}, {Name: "resource_type", Type: "string"}, {Name: "created_at", Type: "string"}, {Name: "modified_at", Type: "string"}}
	if withCompleted {
		fields = append(fields, connectors.Field{Name: "completed", Type: "boolean"})
	}
	return fields
}

func asanaWorkspaceRecord(item map[string]any) connectors.Record {
	return connectors.Record{"gid": item["gid"], "name": item["name"], "resource_type": item["resource_type"]}
}

func asanaProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{"gid": item["gid"], "name": item["name"], "resource_type": item["resource_type"], "created_at": item["created_at"], "modified_at": item["modified_at"]}
}

func asanaTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{"gid": item["gid"], "name": item["name"], "resource_type": item["resource_type"], "created_at": item["created_at"], "modified_at": item["modified_at"], "completed": item["completed"]}
}

func asanaQuery(cfg connectors.RuntimeConfig, endpoint asanaStreamEndpoint, pageSize int) url.Values {
	query := url.Values{"limit": []string{strconv.Itoa(pageSize)}, "opt_fields": []string{endpoint.optFields}}
	if workspace := strings.TrimSpace(cfg.Config["workspace_id"]); workspace != "" && endpoint.resource != "workspaces" {
		query.Set("workspace", workspace)
	}
	if project := strings.TrimSpace(cfg.Config["project_id"]); project != "" && endpoint.resource == "tasks" {
		query.Set("project", project)
	}
	if assignee := strings.TrimSpace(cfg.Config["assignee"]); assignee != "" && endpoint.resource == "tasks" {
		query.Set("assignee", assignee)
	}
	return query
}

func asanaAccessToken(cfg connectors.RuntimeConfig) string { return cfg.Secrets["access_token"] }

func asanaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validBaseURL(asanaName, cfg.Config["base_url"], asanaDefaultBaseURL)
}

func asanaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(asanaName, cfg.Config["page_size"], asanaDefaultPageSize, 1, asanaMaxPageSize)
}

func asanaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	return maxPagesConfig(asanaName, cfg.Config["max_pages"])
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func validBaseURL(name, raw, fallback string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", name, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(name, raw string, fallback, min, max int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s config value must be an integer: %w", name, err)
	}
	if value < min || value > max {
		return 0, fmt.Errorf("%s config value must be between %d and %d", name, min, max)
	}
	return value, nil
}

func maxPagesConfig(name, raw string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer, all, or unlimited: %w", name, err)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s config max_pages must be 0 for unlimited or a positive integer", name)
	}
	return value, nil
}
