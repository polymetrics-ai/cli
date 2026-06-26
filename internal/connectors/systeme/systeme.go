package systeme

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL  = "https://api.systeme.io/api"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("systeme", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return "systeme" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "systeme",
		DisplayName:     "Systeme",
		IntegrationType: "api",
		Description:     "Reads Systeme.io contacts through the Systeme.io API. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
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
	if err := r.DoJSON(ctx, http.MethodGet, "contacts", url.Values{"page": []string{"1"}, "limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check systeme: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := strings.TrimSpace(req.Stream)
	if stream == "" {
		stream = "contacts"
	}
	if stream != "contacts" {
		return fmt.Errorf("systeme stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, "contacts", nil, &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize: pageSize}, "items", maxPages, func(item connsdk.Record) error {
		return emit(mapContact(item))
	})
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := strings.TrimSpace(secretValue(cfg, "api_key"))
	if secret == "" {
		return nil, errors.New("systeme connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("X-API-Key", secret, ""), UserAgent: userAgent}, nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{{
		Name:         "contacts",
		Description:  "Systeme.io contacts.",
		PrimaryKey:   []string{"id"},
		CursorFields: []string{"created_at"},
		Fields: []connectors.Field{
			{Name: "id", Type: "string"},
			{Name: "email", Type: "string"},
			{Name: "created_at", Type: "timestamp"},
		},
	}}
}

func mapContact(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"email":      item["email"],
		"created_at": first(item, "created_at", "createdAt"),
	}
}

func readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{
			"id":         fmt.Sprintf("contact_fixture_%d", i),
			"email":      fmt.Sprintf("fixture+%d@example.com", i),
			"created_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"fixture":    true,
		}); err != nil {
			return err
		}
	}
	return nil
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(configValue(cfg, "base_url"))
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("systeme config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("systeme config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("systeme config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(configValue(cfg, "page_size"))
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("systeme config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(configValue(cfg, "max_pages")))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("systeme config max_pages must be 0, all, unlimited, or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(configValue(cfg, "mode")), "fixture")
}

func configValue(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config[key]
}

func secretValue(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}
