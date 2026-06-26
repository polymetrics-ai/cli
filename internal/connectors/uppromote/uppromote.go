package uppromote

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
	defaultBaseURL = "https://api.uppromote.com"
	userAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("uppromote", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return "uppromote" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "uppromote",
		DisplayName:     "UpPromote",
		IntegrationType: "api",
		Description:     "Reads affiliates from the UpPromote API.",
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
		return errors.New("uppromote connector requires secret api_key")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{affiliatesStream()}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "affiliates"
	}
	if stream != "affiliates" {
		return fmt.Errorf("uppromote stream %q not found", stream)
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
	resp, err := r.Do(ctx, http.MethodGet, "api/affiliates", query, nil)
	if err != nil {
		return fmt.Errorf("read uppromote affiliates: %w", err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "affiliates")
	if err != nil {
		return fmt.Errorf("decode uppromote affiliates: %w", err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(affiliateRecord(item)); err != nil {
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
		return nil, errors.New("uppromote connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(apiKey), UserAgent: userAgent}, nil
}

func affiliatesStream() connectors.Stream {
	return connectors.Stream{
		Name:         "affiliates",
		Description:  "Affiliates registered in UpPromote.",
		PrimaryKey:   []string{"id"},
		CursorFields: []string{"created_at"},
		Fields: []connectors.Field{
			{Name: "id", Type: "string"},
			{Name: "email", Type: "string"},
			{Name: "created_at", Type: "timestamp"},
			{Name: "status", Type: "string"},
		},
	}
}

func affiliateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         stringValue(item["id"]),
		"email":      item["email"],
		"created_at": item["created_at"],
		"status":     item["status"],
	}
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"id": "aff_fixture_1", "email": "affiliate@example.com", "created_at": "2026-01-01T00:00:00Z", "status": "active"})
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("uppromote config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("uppromote config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("uppromote config base_url requires a host")
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
