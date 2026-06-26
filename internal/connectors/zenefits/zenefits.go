package zenefits

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
	connectorName  = "zenefits"
	defaultBaseURL = "https://api.zenefits.com/core"
	userAgent      = "polymetrics-go-cli"
)

type streamSpec struct {
	path       string
	recordPath string
	fields     []connectors.Field
	mapRecord  func(map[string]any) connectors.Record
}

var streams = map[string]streamSpec{
	"people":      {path: "people", recordPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "first_name", Type: "string"}, {Name: "last_name", Type: "string"}, {Name: "status", Type: "string"}}, mapRecord: mapPerson},
	"companies":   {path: "companies", recordPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}, mapRecord: mapCompany},
	"departments": {path: "departments", recordPath: "data", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}, mapRecord: mapCompany},
}

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Zenefits", IntegrationType: "api", Description: "Reads Zenefits people, companies, and departments.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if strings.TrimSpace(cfg.Secrets["token"]) == "" {
		return errors.New("zenefits connector requires secret token")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	out := make([]connectors.Stream, 0, len(streams))
	for name, spec := range streams {
		out = append(out, connectors.Stream{Name: name, Description: "Zenefits " + name + ".", PrimaryKey: []string{"id"}, Fields: spec.fields})
	}
	return connectors.Catalog{Connector: c.Name(), Streams: out}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "people"
	}
	spec, ok := streams[stream]
	if !ok {
		return fmt.Errorf("zenefits stream %q not found", stream)
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
		return fmt.Errorf("read zenefits %s: %w", stream, err)
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
	token := strings.TrimSpace(cfg.Secrets["token"])
	if token == "" {
		return nil, errors.New("zenefits connector requires secret token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func mapPerson(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "first_name": item["first_name"], "last_name": item["last_name"], "status": item["status"]}
}

func mapCompany(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"]}
}

func emitFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	item := map[string]any{"id": stream + "_fixture_1", "first_name": "Fixture", "last_name": "Person", "status": "active", "name": "Fixture " + stream}
	rec := spec.mapRecord(item)
	rec["fixture"] = true
	return emit(rec)
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
		return "", fmt.Errorf("zenefits config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("zenefits config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
