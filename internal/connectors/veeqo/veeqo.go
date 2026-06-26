package veeqo

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
	defaultBaseURL = "https://api.veeqo.com"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("veeqo", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return "veeqo" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "veeqo",
		DisplayName:     "Veeqo",
		IntegrationType: "api",
		Description:     "Reads orders from the Veeqo API.",
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
	if strings.TrimSpace(secret(cfg, "api_key")) == "" {
		return errors.New("veeqo connector requires secret api_key")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{ordersStream()}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "orders"
	}
	if stream != "orders" {
		return fmt.Errorf("veeqo stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	query := url.Values{}
	if start := strings.TrimSpace(req.Config.Config["start_date"]); start != "" {
		query.Set("start_date", start)
	}
	resp, err := r.Do(ctx, http.MethodGet, "orders", query, nil)
	if err != nil {
		return fmt.Errorf("read veeqo orders: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode veeqo orders: %w", err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(orderRecord(item)); err != nil {
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
	apiKey := secret(cfg, "api_key")
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("veeqo connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("x-api-key", apiKey, ""), UserAgent: userAgent}, nil
}

func ordersStream() connectors.Stream {
	return connectors.Stream{
		Name:         "orders",
		Description:  "Orders in Veeqo.",
		PrimaryKey:   []string{"id"},
		CursorFields: []string{"created_at"},
		Fields: []connectors.Field{
			{Name: "id", Type: "string"},
			{Name: "number", Type: "string"},
			{Name: "status", Type: "string"},
			{Name: "created_at", Type: "timestamp"},
		},
	}
}

func orderRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": stringValue(item["id"]), "number": item["number"], "status": item["status"], "created_at": item["created_at"]}
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"id": "1001", "number": "SO-1001", "status": "shipped", "created_at": "2026-01-01T00:00:00Z"})
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("veeqo config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("veeqo config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("veeqo config base_url requires a host")
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
