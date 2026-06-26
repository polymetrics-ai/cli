package vantage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL = "https://api.vantage.sh"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("vantage", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return "vantage" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "vantage",
		DisplayName:     "Vantage",
		IntegrationType: "api",
		Description:     "Reads costs from the Vantage API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := c.requester(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(secret(cfg, "access_token")) == "" {
		return errors.New("vantage connector requires secret access_token")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{costsStream()}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "costs"
	}
	if stream != "costs" {
		return fmt.Errorf("vantage stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, "v2/costs", nil, nil)
	if err != nil {
		return fmt.Errorf("read vantage costs: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "costs")
	if err != nil {
		return fmt.Errorf("decode vantage costs: %w", err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(costRecord(item)); err != nil {
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
	token := secret(cfg, "access_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("vantage connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func costsStream() connectors.Stream {
	return connectors.Stream{
		Name:        "costs",
		Description: "Cost records from Vantage.",
		PrimaryKey:  []string{"id"},
		Fields: []connectors.Field{
			{Name: "id", Type: "string"},
			{Name: "service", Type: "string"},
			{Name: "amount", Type: "string"},
			{Name: "date", Type: "string"},
		},
	}
}

func costRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": stringValue(item["id"]), "service": item["service"], "amount": item["amount"], "date": item["date"]}
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"id": "cost_fixture_1", "service": "compute", "amount": "12.34", "date": "2026-01-01"})
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("vantage config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("vantage config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("vantage config base_url requires a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func stringValue(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}
