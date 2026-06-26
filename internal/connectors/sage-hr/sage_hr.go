package sagehr

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
	connectorName  = "sage-hr"
	defaultBaseURL = "https://api.sage.hr/v1"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamSpec struct {
	path string
	desc string
}

var streams = map[string]streamSpec{
	"employees":        {path: "employees", desc: "Sage HR employees."},
	"teams":            {path: "teams", desc: "Sage HR teams."},
	"timeoff_requests": {path: "timeoff/requests", desc: "Sage HR time off requests."},
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Sage HR", IntegrationType: "api", Description: "Reads Sage HR employees, teams, and time off requests through the Sage HR API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, streams["employees"].path, nil, nil, nil); err != nil {
		return fmt.Errorf("check sage-hr: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "id", Type: "integer"}, {Name: "first_name", Type: "string"}, {Name: "last_name", Type: "string"}, {Name: "name", Type: "string"}}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "employees", Description: streams["employees"].desc, Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "teams", Description: streams["teams"].desc, Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "timeoff_requests", Description: streams["timeoff_requests"].desc, Fields: fields, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "employees"
	}
	spec, ok := streams[stream]
	if !ok {
		return fmt.Errorf("sage-hr stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, spec.path, nil, nil)
	if err != nil {
		return fmt.Errorf("read sage-hr %s: %w", stream, err)
	}
	records, err := recordsAtAny(resp.Body, "data", "")
	if err != nil {
		return fmt.Errorf("decode sage-hr %s: %w", stream, err)
	}
	for _, rec := range records {
		if err := emit(connectors.Record(rec)); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func recordsAtAny(body []byte, paths ...string) ([]map[string]any, error) {
	for _, path := range paths {
		records, err := connsdk.RecordsAt(body, path)
		if err != nil || len(records) > 0 {
			return records, err
		}
	}
	return nil, nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": i, "first_name": fmt.Sprintf("Fixture %d", i), "last_name": stream, "name": fmt.Sprintf("Fixture %s %d", stream, i), "fixture": true}); err != nil {
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
	token := strings.TrimSpace(secret(cfg, "api_key"))
	if token == "" {
		return nil, errors.New("sage-hr connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("X-Auth-Token", token, ""), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("sage-hr config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("sage-hr config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("sage-hr config base_url must include a host")
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
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
