// Package spotlercrm implements the native pm Spotler CRM connector.
package spotlercrm

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
	connectorName   = "spotlercrm"
	defaultBaseURL  = "https://api.spotlercrm.com/api/v1"
	defaultPageSize = 100
	maxPageSize     = 500
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
	return connectors.Metadata{Name: connectorName, DisplayName: "Spotler CRM", IntegrationType: "api", Description: "Reads Spotler CRM contacts, accounts, opportunities, and tasks through the Spotler CRM API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "contacts", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check spotlercrm: %w", err)
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
		stream = "contacts"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("spotlercrm stream %q not found", stream)
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
	paginator := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize: pageSize}
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
		return nil, errors.New("spotlercrm connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("X-API-Key", key, ""), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("spotlercrm config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("spotlercrm config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("spotlercrm config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "spotlercrm config page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("spotlercrm config max_pages must be a non-negative integer: %w", err)
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
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "email": fmt.Sprintf("fixture+%d@example.com", i), "firstName": "Fixture", "lastName": strconv.Itoa(i), "status": "open"}
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
		{Name: "contacts", Description: "Spotler CRM contacts.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["contacts"].fields},
		{Name: "accounts", Description: "Spotler CRM accounts.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["accounts"].fields},
		{Name: "opportunities", Description: "Spotler CRM opportunities.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["opportunities"].fields},
		{Name: "tasks", Description: "Spotler CRM tasks.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["tasks"].fields},
	}
}

var streamEndpoints = map[string]streamEndpoint{
	"contacts":      {resource: "contacts", recordsPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "email", Type: "string"}, {Name: "firstName", Type: "string"}, {Name: "lastName", Type: "string"}}, mapRecord: copyRecord("id", "email", "firstName", "lastName")},
	"accounts":      {resource: "accounts", recordsPath: "data", fields: accountFields(), mapRecord: copyRecord("id", "name", "status")},
	"opportunities": {resource: "opportunities", recordsPath: "data", fields: accountFields(), mapRecord: copyRecord("id", "name", "status")},
	"tasks":         {resource: "tasks", recordsPath: "data", fields: accountFields(), mapRecord: copyRecord("id", "name", "status")},
}

func accountFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}}
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
