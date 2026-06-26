package zendesksunshine

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
	connectorName  = "zendesk-sunshine"
	defaultBaseURL = "https://example.zendesk.com/api/sunshine"
	userAgent      = "polymetrics-go-cli"
)

type streamSpec struct {
	path       string
	recordPath string
	fields     []connectors.Field
	mapRecord  func(map[string]any) connectors.Record
}

var streams = map[string]streamSpec{
	"object_types":  {path: "objects/types", recordPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "schema", Type: "object"}, {Name: "created_at", Type: "timestamp"}}, mapRecord: mapObjectType},
	"objects":       {path: "objects/records", recordPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "type", Type: "string"}, {Name: "attributes", Type: "object"}}, mapRecord: mapObject},
	"relationships": {path: "relationships/records", recordPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "type", Type: "string"}, {Name: "source", Type: "object"}, {Name: "target", Type: "object"}}, mapRecord: mapRelationship},
}

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Zendesk Sunshine", IntegrationType: "api", Description: "Reads Zendesk Sunshine object types, objects, and relationships.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if strings.TrimSpace(cfg.Config["email"]) == "" || strings.TrimSpace(cfg.Secrets["api_token"]) == "" {
		return errors.New("zendesk-sunshine connector requires config email and secret api_token")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	out := make([]connectors.Stream, 0, len(streams))
	for name, spec := range streams {
		out = append(out, connectors.Stream{Name: name, Description: "Zendesk Sunshine " + name + ".", PrimaryKey: []string{"id"}, Fields: spec.fields})
	}
	return connectors.Catalog{Connector: c.Name(), Streams: out}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "object_types"
	}
	spec, ok := streams[stream]
	if !ok {
		return fmt.Errorf("zendesk-sunshine stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, stream, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, spec.path, nil, nil)
	if err != nil {
		return fmt.Errorf("read zendesk-sunshine %s: %w", stream, err)
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
	email := strings.TrimSpace(cfg.Config["email"])
	token := strings.TrimSpace(cfg.Secrets["api_token"])
	if email == "" || token == "" {
		return nil, errors.New("zendesk-sunshine connector requires config email and secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(email+"/token", token), UserAgent: userAgent}, nil
}

func mapObjectType(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["key"], "schema": item["schema"], "created_at": item["created_at"]}
}

func mapObject(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item["id"], item["key"]), "type": item["type"], "attributes": item["attributes"]}
}

func mapRelationship(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item["id"], item["key"]), "type": item["type"], "source": item["source"], "target": item["target"]}
}

func emitFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	item := map[string]any{"id": stream + "_fixture_1", "key": stream + "_fixture_1", "type": "fixture", "schema": map[string]any{"properties": map[string]any{}}, "attributes": map[string]any{"name": "Fixture"}, "source": map[string]any{}, "target": map[string]any{}, "created_at": "2026-01-01T00:00:00Z"}
	rec := spec.mapRecord(item)
	rec["fixture"] = true
	return emit(rec)
}

func first(values ...any) any {
	for _, v := range values {
		if s := fmt.Sprint(v); v != nil && s != "" && s != "<nil>" {
			return v
		}
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" && strings.TrimSpace(cfg.Config["subdomain"]) != "" {
		raw = "https://" + strings.TrimSpace(cfg.Config["subdomain"]) + ".zendesk.com/api/sunshine"
	}
	if raw == "" {
		raw = defaultBaseURL
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("zendesk-sunshine config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("zendesk-sunshine config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
