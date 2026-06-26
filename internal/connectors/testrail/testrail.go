package testrail

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
	connectorName  = "testrail"
	defaultBaseURL = "https://example.testrail.io"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "TestRail", IntegrationType: "api", Description: "Reads TestRail projects through the TestRail API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if strings.TrimSpace(cfg.Config["username"]) == "" || strings.TrimSpace(cfg.Secrets["password"]) == "" {
		return errors.New("testrail connector requires config username and secret password")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "projects", Description: "TestRail projects.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "announcement", Type: "string"}, {Name: "is_completed", Type: "boolean"}}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "projects"
	}
	if stream != "projects" {
		return fmt.Errorf("testrail stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, "index.php?/api/v2/get_projects", nil, nil)
	if err != nil {
		return fmt.Errorf("read testrail projects: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, ".")
	if err != nil {
		return err
	}
	for _, item := range records {
		if err := emit(connectors.Record{"id": item["id"], "name": item["name"], "announcement": item["announcement"], "is_completed": item["is_completed"]}); err != nil {
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
	username := strings.TrimSpace(cfg.Config["username"])
	password := strings.TrimSpace(cfg.Secrets["password"])
	if username == "" || password == "" {
		return nil, errors.New("testrail connector requires config username and secret password")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(username, password), UserAgent: userAgent}, nil
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"id": 1, "name": "Fixture Project", "is_completed": false, "fixture": true})
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
		return "", fmt.Errorf("testrail config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("testrail config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
