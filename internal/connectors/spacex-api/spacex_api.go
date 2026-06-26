// Package spacexapi implements a read-only SpaceX API connector.
package spacexapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL = "https://api.spacexdata.com/v4"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("spacex-api", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "spacex-api" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "spacex-api", DisplayName: "SpaceX API", IntegrationType: "api", Description: "Reads public SpaceX launch, rocket, capsule, crew, payload, and Starlink data.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "company", nil, nil, nil); err != nil {
		return fmt.Errorf("check spacex-api: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "spacex-api", Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "launches"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("spacex-api stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resource := endpoint.resource
	if option := configValue(req.Config, "options"); stream == "launches" && option != "" {
		resource = "launches/" + strings.Trim(option, "/")
	}
	if id := configValue(req.Config, "id"); id != "" {
		resource = strings.TrimRight(resource, "/") + "/" + url.PathEscape(id)
	}
	return readRecords(ctx, r, resource, endpoint.recordsPath, emit)
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ resource, recordsPath string }

var streamEndpoints = map[string]streamEndpoint{
	"launches":   {"launches", ""},
	"rockets":    {"rockets", ""},
	"capsules":   {"capsules", ""},
	"crew":       {"crew", ""},
	"payloads":   {"payloads", ""},
	"starlink":   {"starlink", ""},
	"launchpads": {"launchpads", ""},
	"landpads":   {"landpads", ""},
	"ships":      {"ships", ""},
	"roadster":   {"roadster", ""},
	"company":    {"company", ""},
}

func streams() []connectors.Stream {
	names := []string{"launches", "rockets", "capsules", "crew", "payloads", "starlink", "launchpads", "landpads", "ships", "roadster", "company"}
	out := make([]connectors.Stream, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Stream{Name: name, Description: "SpaceX " + name + ".", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "date_utc", Type: "string"}}})
	}
	return out
}

func readRecords(ctx context.Context, r *connsdk.Requester, resource, recordsPath string, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, resource, nil, nil)
	if err != nil {
		return err
	}
	records, err := connsdk.RecordsAt(resp.Body, recordsPath)
	if err != nil {
		return err
	}
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record(rec)); err != nil {
			return err
		}
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s-%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "date_utc": fmt.Sprintf("2026-01-0%dT00:00:00.000Z", i)}); err != nil {
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
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedBaseURL("spacex-api", configValue(cfg, "base_url"), defaultBaseURL)
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
		return "", fmt.Errorf("%s config base_url must use http or https", connector)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", connector)
	}
	return strings.TrimRight(base, "/"), nil
}

func configValue(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Config == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Config[key])
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(configValue(cfg, "mode"), "fixture")
}
