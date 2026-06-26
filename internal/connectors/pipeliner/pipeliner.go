// Package pipeliner implements a read-only Pipeliner CRM HTTP connector.
// Pipeliner deployments differ by account/space; this connector supports the
// documented REST entity list shape with Basic auth and bounded offset paging.
package pipeliner

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
	defaultBaseURL  = "https://api.pipelinersales.com/api/v100/rest"
	defaultPageSize = 100
	defaultMaxPages = 3
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("pipeliner", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "pipeliner" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "pipeliner",
		DisplayName:     "Pipeliner",
		IntegrationType: "api",
		Description:     "Reads Pipeliner CRM accounts, contacts, opportunities, and leads through the REST API.",
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
	spaceID := strings.TrimSpace(cfg.Config["space_id"])
	if spaceID == "" {
		return errors.New("pipeliner connector requires config space_id")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	path := "spaces/" + url.PathEscape(spaceID) + "/entities/Accounts"
	if _, err := r.Do(ctx, http.MethodGet, path, url.Values{"limit": {"1"}, "offset": {"0"}}, nil); err != nil {
		return fmt.Errorf("check pipeliner: %w", err)
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
		stream = "accounts"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("pipeliner stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, endpoint, emit)
	}
	spaceID := strings.TrimSpace(req.Config.Config["space_id"])
	if spaceID == "" {
		return errors.New("pipeliner connector requires config space_id")
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
	path := "spaces/" + url.PathEscape(spaceID) + "/entities/" + endpoint.resource
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		query := url.Values{"limit": {strconv.Itoa(pageSize)}, "offset": {strconv.Itoa(offset)}}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read pipeliner %s: %w", stream, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode pipeliner %s: %w", stream, err)
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
}

var streamEndpoints = map[string]streamEndpoint{
	"accounts":      {resource: "Accounts", mapRecord: entityRecord},
	"contacts":      {resource: "Contacts", mapRecord: entityRecord},
	"opportunities": {resource: "Opportunities", mapRecord: entityRecord},
	"leads":         {resource: "Leads", mapRecord: entityRecord},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "accounts", Description: "Pipeliner account entities.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields},
		{Name: "contacts", Description: "Pipeliner contact entities.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields},
		{Name: "opportunities", Description: "Pipeliner opportunity entities.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields},
		{Name: "leads", Description: "Pipeliner lead entities.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields},
	}
}

func entityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         first(item, "id", "Id", "ID"),
		"name":       first(item, "name", "Name"),
		"status":     first(item, "status", "Status"),
		"updated_at": first(item, "updated_at", "modified", "Modified", "UpdateDate"),
	}
}

func readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(map[string]any{"id": fmt.Sprintf("fixture-%d", i), "name": fmt.Sprintf("Fixture %d", i), "status": "active", "modified": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)})); err != nil {
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
	username := strings.TrimSpace(firstString(cfg.Secrets, "username", "api_username"))
	password := strings.TrimSpace(firstString(cfg.Secrets, "password", "api_password"))
	if username == "" || password == "" {
		return nil, errors.New("pipeliner connector requires secrets username and password")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(username, password), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("pipeliner config base_url is invalid: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", errors.New("pipeliner config base_url must be an absolute http or https URL")
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
		return 0, fmt.Errorf("pipeliner config %s must be an integer >= %d", key, min)
	}
	if max > 0 && value > max {
		return max, nil
	}
	return value, nil
}

func first(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return nil
}

func firstString(m map[string]string, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(m[key]); v != "" {
			return v
		}
	}
	return ""
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
