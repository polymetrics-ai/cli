// Package wordpress implements the native pm WordPress connector.
package wordpress

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("wordpress", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

type streamEndpoint struct {
	resource string
	fields   []connectors.Field
}

var streamEndpoints = map[string]streamEndpoint{
	"posts":      {resource: "wp-json/wp/v2/posts", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "date", Type: "timestamp"}, {Name: "slug", Type: "string"}, {Name: "status", Type: "string"}}},
	"pages":      {resource: "wp-json/wp/v2/pages", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "date", Type: "timestamp"}, {Name: "slug", Type: "string"}, {Name: "status", Type: "string"}}},
	"comments":   {resource: "wp-json/wp/v2/comments", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "date", Type: "timestamp"}, {Name: "post", Type: "integer"}, {Name: "status", Type: "string"}}},
	"media":      {resource: "wp-json/wp/v2/media", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "date", Type: "timestamp"}, {Name: "slug", Type: "string"}, {Name: "media_type", Type: "string"}}},
	"users":      {resource: "wp-json/wp/v2/users", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "slug", Type: "string"}}},
	"categories": {resource: "wp-json/wp/v2/categories", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "slug", Type: "string"}}},
	"tags":       {resource: "wp-json/wp/v2/tags", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "slug", Type: "string"}}},
}

func (Connector) Name() string { return "wordpress" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "wordpress",
		DisplayName:     "WordPress",
		IntegrationType: "api",
		Description:     "Reads public WordPress REST API resources such as posts, pages, comments, media, users, categories, and tags.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "wp-json/wp/v2/posts", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check wordpress: %w", err)
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
		stream = "posts"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("wordpress stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
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
	base := url.Values{}
	if after := strings.TrimSpace(req.Config.Config["start_date"]); after != "" {
		base.Set("after", after)
	}
	paginator := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, ".", maxPages, func(item connsdk.Record) error {
		return emit(connectors.Record(item))
	})
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	var auth connsdk.Authenticator
	username := strings.TrimSpace(secret(cfg, "username"))
	password := strings.TrimSpace(secret(cfg, "password"))
	if username != "" && password != "" {
		auth = connsdk.Basic(username, password)
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth, UserAgent: userAgent}, nil
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": i, "slug": fmt.Sprintf("fixture-%s-%d", stream, i), "date": "2026-01-01T00:00:00", "status": "publish"}); err != nil {
			return err
		}
	}
	return nil
}

func streams() []connectors.Stream {
	names := make([]string, 0, len(streamEndpoints))
	for name := range streamEndpoints {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]connectors.Stream, 0, len(streamEndpoints))
	for _, name := range names {
		endpoint := streamEndpoints[name]
		out = append(out, connectors.Stream{Name: name, Description: "WordPress " + strings.ReplaceAll(name, "_", " ") + ".", Fields: endpoint.fields, PrimaryKey: []string{"id"}, CursorFields: []string{"date"}})
	}
	return out
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		domain := strings.TrimSpace(cfg.Config["domain"])
		if domain == "" {
			return "", errors.New("wordpress connector requires config domain or base_url")
		}
		if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
			base = domain
		} else {
			base = "https://" + domain
		}
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("wordpress config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("wordpress config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("wordpress config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("wordpress config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("wordpress config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return 1, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("wordpress config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("wordpress config max_pages must be 0 for unlimited or a positive integer")
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
