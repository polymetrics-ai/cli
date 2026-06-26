// Package pivotaltracker implements a read-only Pivotal Tracker API v5 connector.
package pivotaltracker

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
	defaultBaseURL  = "https://www.pivotaltracker.com/services/v5"
	defaultPageSize = 100
	defaultMaxPages = 3
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("pivotal-tracker", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "pivotal-tracker" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "pivotal-tracker",
		DisplayName:     "Pivotal Tracker",
		IntegrationType: "api",
		Description:     "Reads Pivotal Tracker projects, stories, iterations, and epics through API v5.",
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
	if _, err := r.Do(ctx, http.MethodGet, "projects", url.Values{"limit": {"1"}, "offset": {"0"}}, nil); err != nil {
		return fmt.Errorf("check pivotal-tracker: %w", err)
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
		stream = "projects"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("pivotal-tracker stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, endpoint, emit)
	}
	path, err := endpoint.path(req.Config)
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
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		query := url.Values{"limit": {strconv.Itoa(pageSize)}, "offset": {strconv.Itoa(offset)}}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read pivotal-tracker %s: %w", stream, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, ".")
		if err != nil {
			return fmt.Errorf("decode pivotal-tracker %s: %w", stream, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
	projected bool
}

func (e streamEndpoint) path(cfg connectors.RuntimeConfig) (string, error) {
	if !e.projected {
		return e.resource, nil
	}
	projectID := strings.TrimSpace(cfg.Config["project_id"])
	if projectID == "" {
		return "", errors.New("pivotal-tracker stream requires config project_id")
	}
	return "projects/" + url.PathEscape(projectID) + "/" + e.resource, nil
}

var streamEndpoints = map[string]streamEndpoint{
	"projects":   {resource: "projects", mapRecord: projectRecord},
	"stories":    {resource: "stories", mapRecord: storyRecord, projected: true},
	"iterations": {resource: "iterations", mapRecord: iterationRecord, projected: true},
	"epics":      {resource: "epics", mapRecord: epicRecord, projected: true},
}

func streams() []connectors.Stream {
	common := []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "state", Type: "string"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "projects", Description: "Pivotal Tracker projects.", PrimaryKey: []string{"id"}, Fields: common},
		{Name: "stories", Description: "Stories in a configured Pivotal Tracker project.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: common},
		{Name: "iterations", Description: "Iterations in a configured Pivotal Tracker project.", PrimaryKey: []string{"id"}, Fields: common},
		{Name: "epics", Description: "Epics in a configured Pivotal Tracker project.", PrimaryKey: []string{"id"}, Fields: common},
	}
}

func projectRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "state": item["kind"], "updated_at": item["updated_at"]}
}

func storyRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "state": item["current_state"], "updated_at": item["updated_at"]}
}

func iterationRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["number"], "name": item["kind"], "state": item["team_strength"], "updated_at": item["finish"]}
}

func epicRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "state": item["label"], "updated_at": item["updated_at"]}
}

func readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": i, "number": i, "name": fmt.Sprintf("Fixture %d", i), "current_state": "started", "updated_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}
		if err := emit(endpoint.mapRecord(item)); err != nil {
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
	token := strings.TrimSpace(cfg.Secrets["api_token"])
	if token == "" {
		return nil, errors.New("pivotal-tracker connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("X-TrackerToken", token, ""), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("pivotal-tracker config base_url is invalid: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", errors.New("pivotal-tracker config base_url must be an absolute http or https URL")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "page_size", defaultPageSize, 1, 500)
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "max_pages", defaultMaxPages, 0, 10000)
}

func intConfig(cfg connectors.RuntimeConfig, key string, def, min, max int) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config[key]))
	if raw == "" {
		return def, nil
	}
	if key == "max_pages" && (raw == "all" || raw == "unlimited") {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min {
		return 0, fmt.Errorf("pivotal-tracker config %s must be an integer >= %d", key, min)
	}
	if max > 0 && value > max {
		return max, nil
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
