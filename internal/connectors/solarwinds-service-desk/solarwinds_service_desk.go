// Package solarwindsservicedesk implements a read-only SolarWinds Service Desk API connector.
package solarwindsservicedesk

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
	defaultBaseURL = "https://api.samanage.com"
	acceptHeader   = "application/vnd.samanage.v1.1+json"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("solarwinds-service-desk", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "solarwinds-service-desk" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "solarwinds-service-desk", DisplayName: "SolarWinds Service Desk", IntegrationType: "api", Description: "Reads SolarWinds Service Desk incidents, users, departments, categories, problems, and changes.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "incidents.json", nil, nil, nil); err != nil {
		return fmt.Errorf("check solarwinds-service-desk: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "solarwinds-service-desk", Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "incidents"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("solarwinds-service-desk stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return readRecords(ctx, r, endpoint.resource, endpoint.recordsPath, queryParams(req.Config), emit)
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ resource, recordsPath string }

var streamEndpoints = map[string]streamEndpoint{
	"incidents":   {"incidents.json", ""},
	"users":       {"users.json", ""},
	"departments": {"departments.json", ""},
	"categories":  {"categories.json", ""},
	"problems":    {"problems.json", ""},
	"changes":     {"changes.json", ""},
}

func streams() []connectors.Stream {
	names := []string{"incidents", "users", "departments", "categories", "problems", "changes"}
	out := make([]connectors.Stream, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Stream{Name: name, Description: "SolarWinds Service Desk " + name + ".", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "string"}, {Name: "updated_at", Type: "string"}}})
	}
	return out
}

func readRecords(ctx context.Context, r *connsdk.Requester, resource, recordsPath string, q url.Values, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, resource, q, nil)
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
		if err := emit(connectors.Record{"id": i, "name": fmt.Sprintf("Fixture %s %d", stream, i), "created_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}); err != nil {
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
	key := firstSecret(cfg, "api_key_2", "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("solarwinds-service-desk connector requires secret api_key_2")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent, Accept: acceptHeader}, nil
}

func queryParams(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	copyConfig(q, cfg, "start_date", "updated_after")
	copyConfig(q, cfg, "page", "page")
	copyConfig(q, cfg, "per_page", "per_page")
	return q
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedBaseURL("solarwinds-service-desk", configValue(cfg, "base_url"), defaultBaseURL)
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

func copyConfig(q url.Values, cfg connectors.RuntimeConfig, from, to string) {
	if v := configValue(cfg, from); v != "" {
		q.Set(to, v)
	}
}

func firstSecret(cfg connectors.RuntimeConfig, names ...string) string {
	if cfg.Secrets == nil {
		return ""
	}
	for _, name := range names {
		if v := cfg.Secrets[name]; strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(configValue(cfg, "mode"), "fixture")
}
