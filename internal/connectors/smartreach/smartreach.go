// Package smartreach implements a read-only SmartReach API connector.
package smartreach

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
	defaultBaseURL = "https://api.smartreach.io/api/v3"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("smartreach", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "smartreach" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "smartreach", DisplayName: "SmartReach", IntegrationType: "api", Description: "Reads SmartReach teams, campaigns, prospects, email settings, and do-not-contact records.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "teams", nil, nil, nil); err != nil {
		return fmt.Errorf("check smartreach: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "smartreach", Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "campaigns"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("smartreach stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return readRecords(ctx, r, endpoint.resource, endpoint.recordsPath, queryParams(stream, req.Config), emit)
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ resource, recordsPath string }

var streamEndpoints = map[string]streamEndpoint{
	"campaigns":      {"campaigns", "campaigns"},
	"prospects":      {"prospects", "prospects"},
	"teams":          {"teams", "teams"},
	"email_settings": {"email-settings", "email_settings"},
	"do_not_contact": {"do-not-contact", "dnc"},
}

func streams() []connectors.Stream {
	names := []string{"campaigns", "prospects", "teams", "email_settings", "do_not_contact"}
	out := make([]connectors.Stream, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Stream{Name: name, Description: "SmartReach " + name + ".", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "string"}}})
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
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s-%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "created_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}); err != nil {
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
		return nil, errors.New("smartreach connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("X-API-KEY", key, ""), UserAgent: userAgent}, nil
}

func queryParams(stream string, cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	if stream != "teams" {
		if team := firstConfig(cfg, "teamid", "team_id"); team != "" {
			q.Set("team_id", team)
		}
	}
	copyConfig(q, cfg, "older_than", "older_than")
	copyConfig(q, cfg, "newer_than", "newer_than")
	return q
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedBaseURL("smartreach", configValue(cfg, "base_url"), defaultBaseURL)
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

func firstConfig(cfg connectors.RuntimeConfig, names ...string) string {
	for _, name := range names {
		if v := configValue(cfg, name); v != "" {
			return v
		}
	}
	return ""
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

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(configValue(cfg, "mode"), "fixture")
}
