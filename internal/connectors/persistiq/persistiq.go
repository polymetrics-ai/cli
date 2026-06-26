// Package persistiq implements a read-only PersistIQ connector for the v1 list
// endpoints documented by the public API and Singer/dlt connector references.
// Authentication uses the X-API-KEY header; writes are not exposed.
package persistiq

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
	defaultBaseURL  = "https://api.persistiq.com"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init()                     { connectors.RegisterFactory("persistiq", New) }
func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "persistiq" }
func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "persistiq", DisplayName: "PersistIQ", IntegrationType: "api", Description: "Reads PersistIQ users, leads, campaigns, mailboxes, activities, and accounts through v1 REST endpoints.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "v1/users", url.Values{"page": []string{"1"}, "per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check persistiq: %w", err)
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
		stream = "leads"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("persistiq stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	p := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize: size}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, p, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error { return emit(connectors.Record(rec)) })
}
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ resource, recordsPath string }

var streamEndpoints = map[string]streamEndpoint{"leads": {"v1/leads", "leads"}, "users": {"v1/users", "users"}, "campaigns": {"v1/campaigns", "campaigns"}, "mailboxes": {"v1/mailboxes", "mailboxes"}, "activities": {"v1/activities", "activities"}, "accounts": {"v1/accounts", "accounts"}}

func streams() []connectors.Stream {
	names := []string{"leads", "users", "campaigns", "mailboxes", "activities", "accounts"}
	out := make([]connectors.Stream, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Stream{Name: name, Description: "PersistIQ " + name + " list endpoint.", PrimaryKey: []string{"id"}, Fields: commonFields()})
	}
	return out
}
func commonFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "status", Type: "string"}}
}
func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "email": fmt.Sprintf("fixture+%d@example.com", i), "status": "active"}); err != nil {
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
	key := secret(cfg, "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("persistiq connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("X-API-KEY", key, ""), UserAgent: userAgent}, nil
}
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedBaseURL("persistiq", cfg.Config["base_url"], defaultBaseURL)
}
func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt("persistiq", cfg.Config["page_size"], defaultPageSize, maxPageSize, "page_size")
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return optionalInt("persistiq", cfg.Config["max_pages"], "max_pages")
}
func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
func validatedBaseURL(connector, raw, def string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		return def, nil
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", connector, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", connector, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", connector)
	}
	return strings.TrimRight(base, "/"), nil
}
func boundedInt(connector, raw string, def, max int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v < 1 || v > max {
		return 0, fmt.Errorf("%s config %s must be an integer between 1 and %d", connector, name, max)
	}
	return v, nil
}
func optionalInt(connector, raw, name string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, fmt.Errorf("%s config %s must be 0, positive, all, or unlimited", connector, name)
	}
	return v, nil
}
