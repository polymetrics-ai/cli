// Package watchmode implements the native pm Watchmode connector.
package watchmode

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
	defaultBaseURL = "https://api.watchmode.com"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("watchmode", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

type streamEndpoint struct {
	resource    string
	recordsPath string
	fields      []connectors.Field
}

var streamEndpoints = map[string]streamEndpoint{
	"search":  {resource: "v1/search/", recordsPath: "title_results", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "type", Type: "string"}, {Name: "year", Type: "integer"}}},
	"sources": {resource: "v1/sources/", recordsPath: ".", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "type", Type: "string"}, {Name: "region", Type: "string"}}},
}

func (Connector) Name() string { return "watchmode" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "watchmode",
		DisplayName:     "Watchmode",
		IntegrationType: "api",
		Description:     "Reads Watchmode search results and streaming source metadata.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "v1/sources/", nil, nil, nil); err != nil {
		return fmt.Errorf("check watchmode: %w", err)
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
		stream = "search"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("watchmode stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, queryParams(stream, req.Config), nil)
	if err != nil {
		return fmt.Errorf("read watchmode %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode watchmode %s: %w", stream, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
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
	key := strings.TrimSpace(secret(cfg, "api_key"))
	if key == "" {
		return nil, errors.New("watchmode connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("apiKey", key), UserAgent: userAgent}, nil
}

func queryParams(stream string, cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	if stream == "search" {
		value := strings.TrimSpace(cfg.Config["search_val"])
		if value == "" {
			value = "Terminator"
		}
		q.Set("search_value", value)
		q.Set("search_field", "name")
	}
	if start := strings.TrimSpace(cfg.Config["start_date"]); start != "" {
		q.Set("start_date", start)
	}
	return q
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": i, "name": fmt.Sprintf("Fixture Title %d", i), "type": "movie", "stream": stream}); err != nil {
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
		out = append(out, connectors.Stream{Name: name, Description: "Watchmode " + strings.ReplaceAll(name, "_", " ") + ".", Fields: endpoint.fields, PrimaryKey: []string{"id"}})
	}
	return out
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("watchmode config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("watchmode config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("watchmode config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
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
