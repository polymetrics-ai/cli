package vismaeconomic

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
	connectorName  = "visma-economic"
	defaultBaseURL = "https://restapi.e-conomic.com"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Visma e-conomic", IntegrationType: "api", Description: "Reads Visma e-conomic customers.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if strings.TrimSpace(cfg.Secrets["app_secret_token"]) == "" || strings.TrimSpace(cfg.Secrets["agreement_grant_token"]) == "" {
		return errors.New("visma-economic connector requires app_secret_token and agreement_grant_token secrets")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "customers", Description: "Visma e-conomic customers.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "currency", Type: "string"}}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	if stream != "customers" {
		return fmt.Errorf("visma-economic stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, "customers", nil, nil)
	if err != nil {
		return fmt.Errorf("read visma-economic customers: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "collection")
	if err != nil {
		return err
	}
	for _, item := range records {
		if err := emit(connectors.Record{"id": text(item["customerNumber"]), "name": item["name"], "currency": item["currency"]}); err != nil {
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
	app := strings.TrimSpace(cfg.Secrets["app_secret_token"])
	grant := strings.TrimSpace(cfg.Secrets["agreement_grant_token"])
	if app == "" || grant == "" {
		return nil, errors.New("visma-economic connector requires app_secret_token and agreement_grant_token secrets")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, DefaultHeaders: map[string]string{"X-AppSecretToken": app, "X-AgreementGrantToken": grant}, UserAgent: userAgent}, nil
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"id": "1", "name": "Fixture Customer", "currency": "DKK", "fixture": true})
}

func text(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
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
		return "", fmt.Errorf("visma-economic config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("visma-economic config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
