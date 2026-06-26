package taboola

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
	defaultBaseURL  = "https://backstage.taboola.com"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("taboola", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return "taboola" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "taboola",
		DisplayName:     "Taboola",
		IntegrationType: "api",
		Description:     "Reads Taboola campaigns through the Backstage API. Read-only.",
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
	accountID, err := accountID(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("backstage/api/1.0/%s/campaigns", accountID)
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"page": []string{"1"}, "page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check taboola: %w", err)
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
		stream = "campaigns"
	}
	if stream != "campaigns" {
		return fmt.Errorf("taboola stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, emit)
	}
	accountID, err := accountID(req.Config)
	if err != nil {
		return err
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
	path := fmt.Sprintf("backstage/api/1.0/%s/campaigns", accountID)
	return connsdk.Harvest(ctx, r, http.MethodGet, path, nil, &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "page_size", StartPage: 1, PageSize: pageSize}, "results", maxPages, func(item connsdk.Record) error {
		return emit(mapCampaign(item))
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
	clientID := strings.TrimSpace(secretValue(cfg, "client_id"))
	clientSecret := strings.TrimSpace(secretValue(cfg, "client_secret"))
	if clientID == "" {
		return nil, errors.New("taboola connector requires secret client_id")
	}
	if clientSecret == "" {
		return nil, errors.New("taboola connector requires secret client_secret")
	}
	tokenURL := strings.TrimRight(base, "/") + "/backstage/oauth/token"
	if override := strings.TrimSpace(configValue(cfg, "token_url")); override != "" {
		tokenURL = override
	}
	auth := &connsdk.OAuth2ClientCredentials{TokenURL: tokenURL, ClientID: clientID, ClientSecret: clientSecret, Client: c.Client}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth, UserAgent: userAgent}, nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{{
		Name:         "campaigns",
		Description:  "Taboola campaigns.",
		PrimaryKey:   []string{"id"},
		CursorFields: []string{"created_at"},
		Fields: []connectors.Field{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "created_at", Type: "timestamp"},
		},
	}}
}

func mapCampaign(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"created_at": item["created_at"],
	}
}

func readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{
			"id":         fmt.Sprintf("campaign_fixture_%d", i),
			"name":       fmt.Sprintf("Fixture Campaign %d", i),
			"created_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"fixture":    true,
		}); err != nil {
			return err
		}
	}
	return nil
}

func accountID(cfg connectors.RuntimeConfig) (string, error) {
	id := strings.TrimSpace(configValue(cfg, "account_id"))
	if id == "" {
		return "", errors.New("taboola connector requires config account_id")
	}
	if strings.ContainsAny(id, "/?#") || strings.Contains(id, "..") {
		return "", fmt.Errorf("taboola config account_id contains invalid characters: %q", id)
	}
	return url.PathEscape(id), nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(configValue(cfg, "base_url"))
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("taboola config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("taboola config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("taboola config base_url must include a host")
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
		return 0, fmt.Errorf("taboola config page_size must be between 1 and %d", maxPageSize)
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
		return 0, errors.New("taboola config max_pages must be 0, all, unlimited, or a positive integer")
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
