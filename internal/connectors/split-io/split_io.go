// Package splitio implements the native pm Split.io connector.
package splitio

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
	connectorName   = "split-io"
	defaultBaseURL  = "https://api.split.io"
	defaultPageSize = 100
	maxPageSize     = 1000
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamEndpoint struct {
	resource       string
	recordsPath    string
	primaryKey     []string
	cursorFields   []string
	fields         []connectors.Field
	needsWorkspace bool
	mapRecord      func(map[string]any) connectors.Record
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Split.io",
		IntegrationType: "api",
		Description:     "Reads Split.io workspaces, environments, feature flags, and segments through the Split Admin API.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "internal/api/v2/workspaces", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check split-io: %w", err)
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
		stream = "workspaces"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("split-io stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, emit)
	}
	resource, err := resolveResource(endpoint, req.Config)
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
	paginator := &connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam: "offset", PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, resource, nil, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := secret(cfg, "api_key")
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("split-io connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(secret), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("split-io config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("split-io config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("split-io config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.needsWorkspace {
		return endpoint.resource, nil
	}
	workspaceID := strings.TrimSpace(cfg.Config["workspace_id"])
	if workspaceID == "" {
		return "", fmt.Errorf("split-io stream requires config workspace_id for path %q", endpoint.resource)
	}
	return strings.ReplaceAll(endpoint.resource, "{workspace_id}", url.PathEscape(workspaceID)), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "split-io config page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("split-io config max_pages must be a non-negative integer: %w", err)
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

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":        fmt.Sprintf("Fixture %s %d", stream, i),
			"status":      "active",
			"trafficType": "user",
			"environment": "production",
			"updatedAt":   "2026-01-01T00:00:00Z",
		}
		rec := endpoint.mapRecord(item)
		rec["connector"] = connectorName
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "workspaces", Description: "Split.io workspaces.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["workspaces"].fields},
		{Name: "environments", Description: "Split.io environments for a workspace_id.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["environments"].fields},
		{Name: "splits", Description: "Split.io feature flags for a workspace_id.", PrimaryKey: []string{"id"}, CursorFields: []string{"updatedAt"}, Fields: streamEndpoints["splits"].fields},
		{Name: "segments", Description: "Split.io segments for a workspace_id.", PrimaryKey: []string{"id"}, CursorFields: []string{"updatedAt"}, Fields: streamEndpoints["segments"].fields},
	}
}

var streamEndpoints = map[string]streamEndpoint{
	"workspaces":   {resource: "internal/api/v2/workspaces", recordsPath: "objects", fields: commonFields(), mapRecord: copyRecord("id", "name", "status")},
	"environments": {resource: "internal/api/v2/environments/ws/{workspace_id}", recordsPath: "objects", needsWorkspace: true, fields: commonFields(), mapRecord: copyRecord("id", "name", "status")},
	"splits":       {resource: "internal/api/v2/splits/ws/{workspace_id}", recordsPath: "objects", needsWorkspace: true, fields: splitFields(), mapRecord: copyRecord("id", "name", "trafficType", "environment", "status", "updatedAt")},
	"segments":     {resource: "internal/api/v2/segments/ws/{workspace_id}", recordsPath: "objects", needsWorkspace: true, fields: commonFields(), mapRecord: copyRecord("id", "name", "status", "updatedAt")},
}

func commonFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "updatedAt", Type: "timestamp"}}
}

func splitFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "trafficType", Type: "string"}, {Name: "environment", Type: "string"}, {Name: "status", Type: "string"}, {Name: "updatedAt", Type: "timestamp"}}
}

func copyRecord(keys ...string) func(map[string]any) connectors.Record {
	return func(item map[string]any) connectors.Record {
		rec := connectors.Record{}
		for _, key := range keys {
			rec[key] = item[key]
		}
		return rec
	}
}
