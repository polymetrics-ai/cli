package zapiersupportedstorage

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
	connectorName  = "zapier-supported-storage"
	defaultBaseURL = "https://store.zapier.com"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Zapier Supported Storage", IntegrationType: "api", Description: "Reads Zapier Storage records.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if strings.TrimSpace(cfg.Secrets["secret"]) == "" {
		return errors.New("zapier-supported-storage connector requires secret")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "records", Description: "Zapier Storage records.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "key", Type: "string"}, {Name: "value", Type: "string"}, {Name: "updated_at", Type: "timestamp"}}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "records"
	}
	if stream != "records" {
		return fmt.Errorf("zapier-supported-storage stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, "api/records", nil, nil)
	if err != nil {
		return fmt.Errorf("read zapier-supported-storage records: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "records")
	if err != nil {
		return err
	}
	for _, item := range records {
		if err := emit(connectors.Record{"id": item["id"], "key": item["key"], "value": item["value"], "updated_at": item["updated_at"]}); err != nil {
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
	secret := strings.TrimSpace(cfg.Secrets["secret"])
	if secret == "" {
		return nil, errors.New("zapier-supported-storage connector requires secret")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("secret", secret), UserAgent: userAgent}, nil
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"id": "key_fixture_1", "key": "key_fixture_1", "value": "fixture value", "updated_at": "2026-01-01T00:00:00Z", "fixture": true})
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
		return "", fmt.Errorf("zapier-supported-storage config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("zapier-supported-storage config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
