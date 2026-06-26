// Package whiskyhunter implements the native pm Whisky Hunter connector.
package whiskyhunter

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
	defaultBaseURL = "https://whiskyhunter.net"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("whisky-hunter", New)
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
	"auctions":     {resource: "api/auctions_data/", recordsPath: ".", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "dt", Type: "timestamp"}, {Name: "winning_bid", Type: "number"}}},
	"distilleries": {resource: "api/distilleries_info/", recordsPath: ".", fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "country", Type: "string"}}},
}

func (Connector) Name() string { return "whisky-hunter" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "whisky-hunter",
		DisplayName:     "Whisky Hunter",
		IntegrationType: "api",
		Description:     "Reads public Whisky Hunter auction and distillery data.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "api/auctions_data/", nil, nil, nil); err != nil {
		return fmt.Errorf("check whisky-hunter: %w", err)
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
		stream = "auctions"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("whisky-hunter stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read whisky-hunter %s: %w", stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode whisky-hunter %s: %w", stream, err)
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
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: userAgent}, nil
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": i, "dt": "2026-01-01", "winning_bid": 100 + i, "stream": stream}); err != nil {
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
		out = append(out, connectors.Stream{Name: name, Description: "Whisky Hunter " + strings.ReplaceAll(name, "_", " ") + ".", Fields: endpoint.fields, PrimaryKey: []string{"id"}})
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
		return "", fmt.Errorf("whisky-hunter config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("whisky-hunter config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("whisky-hunter config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
