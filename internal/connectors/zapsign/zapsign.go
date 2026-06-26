package zapsign

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
	connectorName  = "zapsign"
	defaultBaseURL = "https://api.zapsign.com.br/api/v1"
	userAgent      = "polymetrics-go-cli"
)

type streamSpec struct {
	path       string
	recordPath string
	fields     []connectors.Field
	mapRecord  func(map[string]any) connectors.Record
}

var streams = map[string]streamSpec{
	"documents": {path: "docs/", recordPath: "results", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "created_at", Type: "timestamp"}}, mapRecord: mapDocument},
	"signers":   {path: "signers/", recordPath: "results", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}}, mapRecord: mapSigner},
	"templates": {path: "templates/", recordPath: "results", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "timestamp"}}, mapRecord: mapTemplate},
}

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "ZapSign", IntegrationType: "api", Description: "Reads ZapSign documents, signers, and templates.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if strings.TrimSpace(cfg.Secrets["api_token"]) == "" {
		return errors.New("zapsign connector requires secret api_token")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	out := make([]connectors.Stream, 0, len(streams))
	for name, spec := range streams {
		out = append(out, connectors.Stream{Name: name, Description: "ZapSign " + name + ".", PrimaryKey: []string{"id"}, Fields: spec.fields})
	}
	return connectors.Catalog{Connector: c.Name(), Streams: out}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "documents"
	}
	spec, ok := streams[stream]
	if !ok {
		return fmt.Errorf("zapsign stream %q not found", stream)
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
		return fmt.Errorf("read zapsign %s: %w", stream, err)
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
	token := strings.TrimSpace(cfg.Secrets["api_token"])
	if token == "" {
		return nil, errors.New("zapsign connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", token, "Token "), UserAgent: userAgent}, nil
}

func mapDocument(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["token"], "name": item["name"], "status": item["status"], "created_at": item["created_at"]}
}

func mapSigner(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item["token"], item["id"]), "name": item["name"], "email": item["email"]}
}

func mapTemplate(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item["token"], item["id"]), "name": item["name"], "created_at": item["created_at"]}
}

func emitFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	item := map[string]any{"token": stream + "_fixture_1", "id": stream + "_fixture_1", "name": "Fixture " + stream, "status": "signed", "created_at": "2026-01-01T00:00:00Z", "email": "fixture@example.com"}
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
	if raw == "" {
		raw = defaultBaseURL
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("zapsign config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("zapsign config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
