// Package twiliotaskrouter implements a read-only native connector for Twilio TaskRouter APIs.
package twiliotaskrouter

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
	defaultBaseURL  = "https://taskrouter.twilio.com"
	defaultPageSize = 50
	maxPageSize     = 1000
	defaultMaxPages = 1
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("twilio-taskrouter", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "twilio-taskrouter" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "twilio-taskrouter", DisplayName: "Twilio TaskRouter", IntegrationType: "api", Description: "Reads Twilio TaskRouter workers, tasks, activities, task queues, and workflows for a workspace.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	path, err := workspaceResource(cfg, streamSpecs["workers"].resource)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"PageSize": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check twilio-taskrouter: %w", err)
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
		stream = "workers"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("twilio-taskrouter stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := workspaceResource(req.Config, spec.resource)
	if err != nil {
		return err
	}
	return harvest(ctx, r, req.Config, path, spec, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func harvest(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, path string, spec streamSpec, emit func(connectors.Record) error) error {
	pageSize, err := boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "twilio-taskrouter config page_size")
	if err != nil {
		return err
	}
	maxPages, err := configuredMaxPages(cfg)
	if err != nil {
		return err
	}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, url.Values{"PageSize": []string{strconv.Itoa(pageSize)}}, nil)
		if err != nil {
			return fmt.Errorf("read twilio-taskrouter %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode twilio-taskrouter %s: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(spec.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

func readFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"sid": fmt.Sprintf("%s_fixture_%d", strings.ToUpper(stream[:2]), i), "friendly_name": fmt.Sprintf("Fixture %s %d", stream, i), "activity_name": "Available", "assignment_status": "pending", "available": true}
		rec := spec.mapRecord(item)
		rec["connector"] = "twilio-taskrouter"
		rec["fixture"] = true
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
	sid := strings.TrimSpace(cfg.Secrets["account_sid"])
	token := strings.TrimSpace(cfg.Secrets["auth_token"])
	if sid == "" || token == "" {
		return nil, errors.New("twilio-taskrouter connector requires secrets account_sid and auth_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(sid, token), UserAgent: userAgent}, nil
}

func workspaceResource(cfg connectors.RuntimeConfig, resource string) (string, error) {
	workspace := strings.TrimSpace(cfg.Config["workspace_sid"])
	if workspace == "" {
		return "", errors.New("twilio-taskrouter config workspace_sid is required")
	}
	return fmt.Sprintf("v1/Workspaces/%s/%s", url.PathEscape(workspace), resource), nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("twilio-taskrouter config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("twilio-taskrouter config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("twilio-taskrouter config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func configuredMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return defaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("twilio-taskrouter config max_pages must be a non-negative integer: %w", err)
	}
	return value, nil
}
func boundedInt(raw string, def, max int, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s must be between 1 and %d", name, max)
	}
	return value, nil
}
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamSpec struct {
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamSpecs = map[string]streamSpec{
	"workers":     {resource: "Workers", recordsPath: "workers", mapRecord: workerRecord},
	"tasks":       {resource: "Tasks", recordsPath: "tasks", mapRecord: taskRecord},
	"activities":  {resource: "Activities", recordsPath: "activities", mapRecord: namedRecord},
	"task_queues": {resource: "TaskQueues", recordsPath: "task_queues", mapRecord: namedRecord},
	"workflows":   {resource: "Workflows", recordsPath: "workflows", mapRecord: namedRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "workers", Description: "TaskRouter workers.", PrimaryKey: []string{"sid"}, Fields: []connectors.Field{{Name: "sid", Type: "string"}, {Name: "friendly_name", Type: "string"}, {Name: "activity_name", Type: "string"}, {Name: "available", Type: "boolean"}}},
		{Name: "tasks", Description: "TaskRouter tasks.", PrimaryKey: []string{"sid"}, Fields: []connectors.Field{{Name: "sid", Type: "string"}, {Name: "assignment_status", Type: "string"}, {Name: "workflow_sid", Type: "string"}}},
		{Name: "activities", Description: "TaskRouter activities.", PrimaryKey: []string{"sid"}, Fields: namedFields()},
		{Name: "task_queues", Description: "TaskRouter task queues.", PrimaryKey: []string{"sid"}, Fields: namedFields()},
		{Name: "workflows", Description: "TaskRouter workflows.", PrimaryKey: []string{"sid"}, Fields: namedFields()},
	}
}

func namedFields() []connectors.Field {
	return []connectors.Field{{Name: "sid", Type: "string"}, {Name: "friendly_name", Type: "string"}}
}
func workerRecord(item map[string]any) connectors.Record {
	return connectors.Record{"sid": item["sid"], "friendly_name": item["friendly_name"], "activity_name": item["activity_name"], "available": item["available"]}
}
func taskRecord(item map[string]any) connectors.Record {
	return connectors.Record{"sid": item["sid"], "assignment_status": item["assignment_status"], "workflow_sid": item["workflow_sid"]}
}
func namedRecord(item map[string]any) connectors.Record {
	return connectors.Record{"sid": item["sid"], "friendly_name": item["friendly_name"]}
}
