// Package webflow implements the native pm Webflow connector.
package webflow

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
	defaultBaseURL = "https://api.webflow.com"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("webflow", New)
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
	"collections": {resource: siteResource("v2/sites/%s/collections"), recordsPath: "collections", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "displayName", Type: "string"}, {Name: "slug", Type: "string"}}},
	"pages":       {resource: siteResource("v2/sites/%s/pages"), recordsPath: "pages", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "slug", Type: "string"}}},
	"forms":       {resource: siteResource("v2/sites/%s/forms"), recordsPath: "forms", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "displayName", Type: "string"}, {Name: "createdOn", Type: "timestamp"}}},
}

func (Connector) Name() string { return "webflow" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "webflow",
		DisplayName:     "Webflow",
		IntegrationType: "api",
		Description:     "Reads Webflow site collections, pages, and forms using the Webflow Data API.",
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
	path, err := siteResource("v2/sites/%s/collections")(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, path, nil, nil, nil); err != nil {
		return fmt.Errorf("check webflow: %w", err)
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
		stream = "collections"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("webflow stream %q not found", stream)
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
		return fmt.Errorf("read webflow %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode webflow %s: %w", stream, err)
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
		return nil, errors.New("webflow connector requires secret api_key")
	}
	headers := map[string]string{}
	if version := strings.TrimSpace(cfg.Config["accept_version"]); version != "" {
		headers["Accept-Version"] = version
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent, DefaultHeaders: headers}, nil
}

func siteResource(format string) func(connectors.RuntimeConfig) (string, error) {
	return func(cfg connectors.RuntimeConfig) (string, error) {
		siteID := strings.TrimSpace(cfg.Config["site_id"])
		if siteID == "" {
			return "", errors.New("webflow connector requires config site_id")
		}
		return fmt.Sprintf(format, url.PathEscape(siteID)), nil
	}
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "displayName": fmt.Sprintf("Fixture %s %d", stream, i), "slug": fmt.Sprintf("fixture-%d", i)}); err != nil {
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
		out = append(out, connectors.Stream{Name: name, Description: "Webflow " + strings.ReplaceAll(name, "_", " ") + ".", Fields: endpoint.fields, PrimaryKey: []string{"id"}})
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
		return "", fmt.Errorf("webflow config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("webflow config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("webflow config base_url must include a host")
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
