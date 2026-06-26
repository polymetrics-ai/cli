// Package weatherstack implements the native pm Weatherstack connector.
package weatherstack

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
	defaultBaseURL = "https://api.weatherstack.com"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("weatherstack", New)
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
	"current":    {resource: "current", recordsPath: ".", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "location", Type: "object"}, {Name: "current", Type: "object"}}},
	"historical": {resource: "historical", recordsPath: ".", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "location", Type: "object"}, {Name: "historical", Type: "object"}}},
	"forecast":   {resource: "forecast", recordsPath: ".", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "location", Type: "object"}, {Name: "forecast", Type: "object"}}},
}

func (Connector) Name() string { return "weatherstack" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "weatherstack",
		DisplayName:     "Weatherstack",
		IntegrationType: "api",
		Description:     "Reads current, historical, and forecast weather data from Weatherstack.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "current", queryParams("current", cfg), nil, nil); err != nil {
		return fmt.Errorf("check weatherstack: %w", err)
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
		stream = "current"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("weatherstack stream %q not found", stream)
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
		return fmt.Errorf("read weatherstack %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode weatherstack %s: %w", stream, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if item["id"] == nil {
			item["id"] = fmt.Sprintf("%s:%s", stream, strings.TrimSpace(req.Config.Config["query"]))
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
	key := strings.TrimSpace(secret(cfg, "access_key"))
	if key == "" {
		return nil, errors.New("weatherstack connector requires secret access_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("access_key", key), UserAgent: userAgent}, nil
}

func queryParams(stream string, cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	if value := strings.TrimSpace(cfg.Config["query"]); value != "" {
		q.Set("query", value)
	}
	if stream == "historical" {
		if value := strings.TrimSpace(cfg.Config["historical_date"]); value != "" {
			q.Set("historical_date", value)
		}
	}
	if stream == "forecast" {
		if value := strings.TrimSpace(cfg.Config["forecast_days"]); value != "" {
			q.Set("forecast_days", value)
		}
	}
	return q
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "location": map[string]any{"name": "Fixture City"}, "current": map[string]any{"temperature": 20 + i}}); err != nil {
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
		out = append(out, connectors.Stream{Name: name, Description: "Weatherstack " + name + " weather.", Fields: endpoint.fields, PrimaryKey: []string{"id"}})
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
		return "", fmt.Errorf("weatherstack config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("weatherstack config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("weatherstack config base_url must include a host")
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
