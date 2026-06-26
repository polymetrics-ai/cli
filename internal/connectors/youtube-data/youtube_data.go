package youtubedata

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName  = "youtube-data"
	defaultBaseURL = "https://www.googleapis.com/youtube/v3"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamSpec struct {
	path       string
	recordPath string
	fields     []connectors.Field
	mapRecord  func(map[string]any) connectors.Record
}

var streams = map[string]streamSpec{
	"channels":  {path: "channels", recordPath: "items", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "view_count", Type: "string"}}, mapRecord: mapChannel},
	"videos":    {path: "videos", recordPath: "items", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "published_at", Type: "timestamp"}}, mapRecord: mapVideo},
	"playlists": {path: "playlists", recordPath: "items", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "published_at", Type: "timestamp"}}, mapRecord: mapVideo},
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "YouTube Data", IntegrationType: "api", Description: "Reads channels, videos, and playlists through the YouTube Data API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Secrets["api_key"]) == "" {
		return errors.New("youtube-data connector requires secret api_key")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	out := make([]connectors.Stream, 0, len(streams))
	for name, spec := range streams {
		out = append(out, connectors.Stream{Name: name, Description: "YouTube Data " + name + ".", PrimaryKey: []string{"id"}, Fields: spec.fields})
	}
	return connectors.Catalog{Connector: c.Name(), Streams: out}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "channels"
	}
	spec, ok := streams[stream]
	if !ok {
		return fmt.Errorf("youtube-data stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, stream, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("part", "snippet,statistics")
	if stream == "channels" {
		if ids := strings.TrimSpace(req.Config.Config["channel_ids"]); ids != "" {
			q.Set("id", ids)
		}
	} else if ids := strings.TrimSpace(req.Config.Config["ids"]); ids != "" {
		q.Set("id", ids)
	}
	resp, err := r.Do(ctx, http.MethodGet, spec.path, q, nil)
	if err != nil {
		return fmt.Errorf("read youtube-data %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, spec.recordPath)
	if err != nil {
		return err
	}
	for _, item := range records {
		if err := emit(spec.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := strings.TrimSpace(cfg.Secrets["api_key"])
	if key == "" {
		return nil, errors.New("youtube-data connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("key", key), UserAgent: userAgent}, nil
}

func mapChannel(item map[string]any) connectors.Record {
	snippet := object(item["snippet"])
	stats := object(item["statistics"])
	return connectors.Record{"id": item["id"], "title": snippet["title"], "view_count": stats["viewCount"]}
}

func mapVideo(item map[string]any) connectors.Record {
	snippet := object(item["snippet"])
	return connectors.Record{"id": item["id"], "title": snippet["title"], "published_at": snippet["publishedAt"]}
}

func emitFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	item := map[string]any{"id": stream + "_fixture_1", "snippet": map[string]any{"title": "Fixture " + stream, "publishedAt": "2026-01-01T00:00:00Z"}, "statistics": map[string]any{"viewCount": "42"}}
	rec := spec.mapRecord(item)
	rec["fixture"] = true
	return emit(rec)
}

func object(v any) map[string]any {
	if out, ok := v.(map[string]any); ok {
		return out
	}
	return map[string]any{}
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("youtube-data config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("youtube-data config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
