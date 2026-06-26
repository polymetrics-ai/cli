// Package pexelsapi implements a read-only Pexels API connector. Pexels uses an
// Authorization header containing the raw API key and paginates with next_page.
package pexelsapi

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
	defaultBaseURL  = "https://api.pexels.com"
	defaultPageSize = 40
	maxPageSize     = 80
	userAgent       = "polymetrics-go-cli"
)

func init()                     { connectors.RegisterFactory("pexels-api", New) }
func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "pexels-api" }
func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "pexels-api", DisplayName: "Pexels API", IntegrationType: "api", Description: "Reads Pexels photo and video search results through the Pexels REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "v1/search", url.Values{"query": []string{"people"}, "per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check pexels-api: %w", err)
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
		stream = "photos"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("pexels-api stream %q not found", stream)
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
	query := url.Values{"page": []string{"1"}, "per_page": []string{strconv.Itoa(size)}}
	for _, key := range []string{"query", "orientation", "size", "color", "locale"} {
		if v := strings.TrimSpace(req.Config.Config[key]); v != "" {
			query.Set(key, v)
		}
	}
	if query.Get("query") == "" && endpoint.requiresQuery {
		query.Set("query", "people")
	}
	path := endpoint.resource
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read pexels-api %s: %w", stream, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode pexels-api %s: %w", stream, err)
		}
		for _, rec := range records {
			if err := emit(endpoint.mapRecord(map[string]any(rec))); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_page")
		if err != nil {
			return fmt.Errorf("decode pexels-api next_page: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path, query = next, nil
	}
	return nil
}
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	resource, recordsPath string
	requiresQuery         bool
	mapRecord             func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{"photos": {"v1/search", "photos", true, photoRecord}, "curated_photos": {"v1/curated", "photos", false, photoRecord}, "videos": {"v1/videos/search", "videos", true, videoRecord}, "popular_videos": {"v1/videos/popular", "videos", false, videoRecord}}

func streams() []connectors.Stream {
	return []connectors.Stream{{Name: "photos", Description: "Pexels photo search results.", PrimaryKey: []string{"id"}, Fields: photoFields()}, {Name: "curated_photos", Description: "Pexels curated photos.", PrimaryKey: []string{"id"}, Fields: photoFields()}, {Name: "videos", Description: "Pexels video search results.", PrimaryKey: []string{"id"}, Fields: videoFields()}, {Name: "popular_videos", Description: "Pexels popular videos.", PrimaryKey: []string{"id"}, Fields: videoFields()}}
}
func photoFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "integer"}, {Name: "url", Type: "string"}, {Name: "photographer", Type: "string"}, {Name: "photographer_url", Type: "string"}, {Name: "src", Type: "object"}, {Name: "alt", Type: "string"}}
}
func videoFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "integer"}, {Name: "url", Type: "string"}, {Name: "image", Type: "string"}, {Name: "duration", Type: "integer"}, {Name: "user", Type: "object"}}
}
func photoRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "url": item["url"], "photographer": item["photographer"], "photographer_url": item["photographer_url"], "src": item["src"], "alt": item["alt"]}
}
func videoRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "url": item["url"], "image": item["image"], "duration": item["duration"], "user": item["user"]}
}
func readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": i, "url": fmt.Sprintf("https://example.com/%d", i), "photographer": "Fixture Photographer", "src": map[string]any{"original": "https://example.com/image.jpg"}, "alt": "fixture", "image": "https://example.com/video.jpg", "duration": 30, "user": map[string]any{"name": "Fixture"}}
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
		return nil, errors.New("pexels-api connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", key, ""), UserAgent: userAgent}, nil
}
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedBaseURL("pexels-api", cfg.Config["base_url"], defaultBaseURL)
}
func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt("pexels-api", cfg.Config["page_size"], defaultPageSize, maxPageSize, "page_size")
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return optionalInt("pexels-api", cfg.Config["max_pages"], "max_pages")
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
