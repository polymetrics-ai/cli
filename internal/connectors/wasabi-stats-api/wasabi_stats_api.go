// Package wasabistatsapi implements the native pm Wasabi Stats API connector.
package wasabistatsapi

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
	defaultBaseURL = "https://stats.wasabisys.com"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("wasabi-stats-api", New)
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
	"bucket_stats":  {resource: "v1/stats", recordsPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "bucket", Type: "string"}, {Name: "date", Type: "timestamp"}, {Name: "storage_bytes", Type: "integer"}}},
	"account_stats": {resource: "v1/accounts", recordsPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "date", Type: "timestamp"}, {Name: "storage_bytes", Type: "integer"}, {Name: "object_count", Type: "integer"}}},
}

func (Connector) Name() string { return "wasabi-stats-api" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "wasabi-stats-api",
		DisplayName:     "Wasabi Stats API",
		IntegrationType: "api",
		Description:     "Reads Wasabi account and bucket storage statistics from the Wasabi Stats API.",
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
	q := url.Values{"start_date": []string{strings.TrimSpace(cfg.Config["start_date"])}}
	if err := r.DoJSON(ctx, http.MethodGet, "v1/stats", q, nil, nil); err != nil {
		return fmt.Errorf("check wasabi-stats-api: %w", err)
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
		stream = "bucket_stats"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("wasabi-stats-api stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	q := url.Values{}
	if start := strings.TrimSpace(req.Config.Config["start_date"]); start != "" {
		q.Set("start_date", start)
	}
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, q, nil)
	if err != nil {
		return fmt.Errorf("read wasabi-stats-api %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode wasabi-stats-api %s: %w", stream, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if item["id"] == nil {
			item["id"] = firstNonEmpty(fmt.Sprint(item["bucket"]), fmt.Sprint(item["date"]), stream)
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
		return nil, errors.New("wasabi-stats-api connector requires secret api_key")
	}
	auth := connsdk.APIKeyHeader("Authorization", key, "Bearer ")
	if parts := strings.SplitN(key, ":", 2); len(parts) == 2 {
		auth = connsdk.Basic(parts[0], parts[1])
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth, UserAgent: userAgent}, nil
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "bucket": fmt.Sprintf("fixture-bucket-%d", i), "date": "2026-01-01T00:00:00Z", "storage_bytes": i * 1024}); err != nil {
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
		out = append(out, connectors.Stream{Name: name, Description: "Wasabi Stats API " + strings.ReplaceAll(name, "_", " ") + ".", Fields: endpoint.fields, PrimaryKey: []string{"id"}, CursorFields: []string{"date"}})
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
		return "", fmt.Errorf("wasabi-stats-api config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("wasabi-stats-api config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("wasabi-stats-api config base_url must include a host")
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}
