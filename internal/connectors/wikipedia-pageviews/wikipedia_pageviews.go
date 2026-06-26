// Package wikipediapageviews implements the native pm Wikipedia Pageviews connector.
package wikipediapageviews

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL = "https://wikimedia.org"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("wikipedia-pageviews", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

type streamEndpoint struct {
	resource    func(connectors.RuntimeConfig) (string, error)
	recordsPath string
	fields      []connectors.Field
}

var streamEndpoints = map[string]streamEndpoint{
	"pageviews":    {resource: pageviewsPath, recordsPath: "items", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "project", Type: "string"}, {Name: "article", Type: "string"}, {Name: "timestamp", Type: "string"}, {Name: "views", Type: "integer"}}},
	"top_articles": {resource: topArticlesPath, recordsPath: "items", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "country", Type: "string"}, {Name: "articles", Type: "array"}}},
}

func (Connector) Name() string { return "wikipedia-pageviews" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "wikipedia-pageviews",
		DisplayName:     "Wikipedia Pageviews",
		IntegrationType: "api",
		Description:     "Reads Wikimedia pageview metrics for articles and top-article reports.",
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
	path, err := pageviewsPath(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, path, nil, nil, nil); err != nil {
		return fmt.Errorf("check wikipedia-pageviews: %w", err)
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
		stream = "pageviews"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("wikipedia-pageviews stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}
	path, err := endpoint.resource(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read wikipedia-pageviews %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode wikipedia-pageviews %s: %w", stream, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if item["id"] == nil {
			item["id"] = recordID(item, req.Config)
		}
		if err := emit(connectors.Record(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: userAgent}, nil
}

func pageviewsPath(cfg connectors.RuntimeConfig) (string, error) {
	project, err := required(cfg, "project")
	if err != nil {
		return "", err
	}
	access, err := required(cfg, "access")
	if err != nil {
		return "", err
	}
	agent, err := required(cfg, "agent")
	if err != nil {
		return "", err
	}
	article, err := required(cfg, "article")
	if err != nil {
		return "", err
	}
	start, err := required(cfg, "start")
	if err != nil {
		return "", err
	}
	end, err := required(cfg, "end")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("api/rest_v1/metrics/pageviews/per-article/%s/%s/%s/%s/daily/%s/%s", url.PathEscape(project), url.PathEscape(access), url.PathEscape(agent), url.PathEscape(article), url.PathEscape(start), url.PathEscape(end)), nil
}

func topArticlesPath(cfg connectors.RuntimeConfig) (string, error) {
	project, err := required(cfg, "project")
	if err != nil {
		return "", err
	}
	access, err := required(cfg, "access")
	if err != nil {
		return "", err
	}
	country, err := required(cfg, "country")
	if err != nil {
		return "", err
	}
	start, err := required(cfg, "start")
	if err != nil {
		return "", err
	}
	date := start
	if len(date) >= 8 {
		date = date[:8]
	}
	if len(date) != 8 {
		return "", errors.New("wikipedia-pageviews config start must be YYYYMMDD or YYYYMMDDHH")
	}
	return fmt.Sprintf("api/rest_v1/metrics/pageviews/top-per-country/%s/%s/%s/%s/%s/%s", url.PathEscape(project), url.PathEscape(country), url.PathEscape(access), date[:4], date[4:6], date[6:8]), nil
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "project": "en.wikipedia.org", "article": "Fixture", "timestamp": fmt.Sprintf("2026010%d00", i), "views": 100 + i}); err != nil {
			return err
		}
	}
	return nil
}

func recordID(item map[string]any, cfg connectors.RuntimeConfig) string {
	parts := []string{fmt.Sprint(item["project"]), fmt.Sprint(item["article"]), fmt.Sprint(item["timestamp"])}
	if parts[0] == "<nil>" {
		parts[0] = cfg.Config["project"]
	}
	return strings.Join(parts, ":")
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
		out = append(out, connectors.Stream{Name: name, Description: "Wikimedia " + strings.ReplaceAll(name, "_", " ") + " metrics.", Fields: endpoint.fields, PrimaryKey: []string{"id"}, CursorFields: []string{"timestamp"}})
	}
	return out
}

func required(cfg connectors.RuntimeConfig, key string) (string, error) {
	value := strings.TrimSpace(cfg.Config[key])
	if value == "" {
		return "", fmt.Errorf("wikipedia-pageviews connector requires config %s", key)
	}
	return value, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("wikipedia-pageviews config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("wikipedia-pageviews config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("wikipedia-pageviews config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
