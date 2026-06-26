// Package unleash implements a read-only native connector for Unleash APIs.
package unleash

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
	defaultBaseURL  = "https://app.unleash-hosted.com"
	defaultProject  = "default"
	defaultPageSize = 100
	maxPageSize     = 1000
	defaultMaxPages = 1
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("unleash", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "unleash" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "unleash", DisplayName: "Unleash", IntegrationType: "api", Description: "Reads Unleash projects, feature toggles, environments, and segments through admin API list endpoints.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	path := projectResource(cfg, streamSpecs["features"].resource)
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check unleash: %w", err)
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
		stream = "features"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("unleash stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path := spec.resource
	if spec.projectScoped {
		path = projectResource(req.Config, spec.resource)
	}
	return harvest(ctx, r, req.Config, path, spec, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func harvest(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, path string, spec streamSpec, emit func(connectors.Record) error) error {
	pageSize, err := boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "unleash config page_size")
	if err != nil {
		return err
	}
	maxPages, err := configuredMaxPages(cfg)
	if err != nil {
		return err
	}
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		query := url.Values{"limit": []string{strconv.Itoa(pageSize)}, "offset": []string{strconv.Itoa((page - 1) * pageSize)}}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read unleash %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode unleash %s: %w", path, err)
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
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("fixture_%s_%d", stream, i), "project": defaultProject, "enabled": true, "type": "release"}
		rec := spec.mapRecord(item)
		rec["connector"] = "unleash"
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
	token := strings.TrimSpace(cfg.Secrets["api_token"])
	if token == "" {
		return nil, errors.New("unleash connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func projectResource(cfg connectors.RuntimeConfig, suffix string) string {
	project := strings.TrimSpace(cfg.Config["project_id"])
	if project == "" {
		project = defaultProject
	}
	return "api/admin/projects/" + url.PathEscape(project) + "/" + suffix
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("unleash config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("unleash config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("unleash config base_url must include a host")
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
		return 0, fmt.Errorf("unleash config max_pages must be a non-negative integer: %w", err)
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
	resource      string
	recordsPath   string
	projectScoped bool
	mapRecord     func(map[string]any) connectors.Record
}

var streamSpecs = map[string]streamSpec{
	"projects":     {resource: "api/admin/projects", recordsPath: "projects", mapRecord: projectRecord},
	"features":     {resource: "features", recordsPath: "features", projectScoped: true, mapRecord: featureRecord},
	"environments": {resource: "api/admin/environments", recordsPath: "environments", mapRecord: namedRecord},
	"segments":     {resource: "api/admin/segments", recordsPath: "segments", mapRecord: namedRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "projects", Description: "Unleash projects.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}},
		{Name: "features", Description: "Unleash feature toggles for a project.", PrimaryKey: []string{"name"}, Fields: []connectors.Field{{Name: "name", Type: "string"}, {Name: "project", Type: "string"}, {Name: "enabled", Type: "boolean"}, {Name: "type", Type: "string"}}},
		{Name: "environments", Description: "Unleash environments.", PrimaryKey: []string{"name"}, Fields: namedFields()},
		{Name: "segments", Description: "Unleash segments.", PrimaryKey: []string{"id"}, Fields: namedFields()},
	}
}

func namedFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}
}
func projectRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "name"), "name": item["name"]}
}
func featureRecord(item map[string]any) connectors.Record {
	return connectors.Record{"name": item["name"], "project": item["project"], "enabled": item["enabled"], "type": item["type"]}
}
func namedRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "name"), "name": item["name"]}
}
func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}
