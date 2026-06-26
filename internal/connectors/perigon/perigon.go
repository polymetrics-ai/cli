// Package perigon implements a read-only Perigon News API connector. Perigon
// authenticates with the documented apiKey query parameter; this connector keeps
// support narrow to the core list endpoints used for ETL.
package perigon

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
	defaultBaseURL  = "https://api.perigon.io"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("perigon", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "perigon" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "perigon", DisplayName: "Perigon", IntegrationType: "api", Description: "Reads Perigon news articles and stories through the Perigon REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "v1/articles/all", url.Values{"size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check perigon: %w", err)
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
		stream = "articles"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("perigon stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, endpoint, emit)
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
	base := url.Values{}
	if q := strings.TrimSpace(req.Config.Config["query"]); q != "" {
		base.Set("q", q)
	}
	if start := firstNonEmpty(req.State["cursor"], req.Config.Config["start_date"]); start != "" {
		base.Set("from", start)
	}
	p := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "size", StartPage: 1, PageSize: size}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, p, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(map[string]any(rec)))
	})
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	resource, recordsPath string
	mapRecord             func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"articles": {resource: "v1/articles/all", recordsPath: "articles", mapRecord: articleRecord},
	"stories":  {resource: "v1/stories/all", recordsPath: "stories", mapRecord: passthroughRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "articles", Description: "Perigon article search results.", PrimaryKey: []string{"article_id"}, CursorFields: []string{"pub_date"}, Fields: []connectors.Field{{Name: "article_id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "pub_date", Type: "timestamp"}, {Name: "url", Type: "string"}}},
		{Name: "stories", Description: "Perigon story clusters.", PrimaryKey: []string{"id"}, Fields: commonFields()},
	}
}

func articleRecord(item map[string]any) connectors.Record {
	return connectors.Record{"article_id": firstAny(item, "articleId", "id"), "title": item["title"], "pub_date": firstAny(item, "pubDate", "publishedAt"), "url": item["url"], "source": item["source"]}
}

func passthroughRecord(item map[string]any) connectors.Record { return connectors.Record(item) }

func readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("fixture-%d", i), "articleId": fmt.Sprintf("article-%d", i), "title": fmt.Sprintf("Fixture Article %d", i), "pubDate": "2026-01-01T00:00:00Z", "url": fmt.Sprintf("https://example.com/%d", i)}
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
	key := secret(cfg, "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("perigon connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("apiKey", key), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedBaseURL("perigon", cfg.Config["base_url"], defaultBaseURL)
}
func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt("perigon", cfg.Config["page_size"], defaultPageSize, maxPageSize, "page_size")
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return optionalInt("perigon", cfg.Config["max_pages"], "max_pages")
}
func commonFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
}
func firstAny(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if v := m[k]; v != nil {
			return v
		}
	}
	return nil
}
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
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
