// Package statsig implements the native pm Statsig connector.
package statsig

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
	connectorName   = "statsig"
	defaultBaseURL  = "https://statsigapi.net/console/v1"
	defaultPageSize = 100
	maxPageSize     = 1000
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamEndpoint struct {
	resource    string
	recordsPath string
	fields      []connectors.Field
	mapRecord   func(map[string]any) connectors.Record
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Statsig", IntegrationType: "api", Description: "Reads Statsig feature gates, dynamic configs, experiments, and segments through the Statsig Console API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "gates", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check statsig: %w", err)
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
		stream = "feature_gates"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("statsig stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, emit)
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
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
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
	key := secret(cfg, "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("statsig connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("STATSIG-API-KEY", key, ""), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("statsig config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("statsig config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("statsig config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "statsig config page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("statsig config max_pages must be a non-negative integer: %w", err)
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
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "description": "fixture record", "status": "active", "isEnabled": true}
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
		{Name: "feature_gates", Description: "Statsig feature gates.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["feature_gates"].fields},
		{Name: "dynamic_configs", Description: "Statsig dynamic configs.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["dynamic_configs"].fields},
		{Name: "experiments", Description: "Statsig experiments.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["experiments"].fields},
		{Name: "segments", Description: "Statsig segments.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["segments"].fields},
	}
}

var streamEndpoints = map[string]streamEndpoint{
	"feature_gates":   {resource: "gates", recordsPath: "data", fields: commonFields(), mapRecord: copyRecord("id", "name", "description", "status", "isEnabled")},
	"dynamic_configs": {resource: "dynamic_configs", recordsPath: "data", fields: commonFields(), mapRecord: copyRecord("id", "name", "description", "status", "isEnabled")},
	"experiments":     {resource: "experiments", recordsPath: "data", fields: commonFields(), mapRecord: copyRecord("id", "name", "description", "status", "isEnabled")},
	"segments":        {resource: "segments", recordsPath: "data", fields: commonFields(), mapRecord: copyRecord("id", "name", "description", "status", "isEnabled")},
}

func commonFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "description", Type: "string"}, {Name: "status", Type: "string"}, {Name: "isEnabled", Type: "boolean"}}
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
