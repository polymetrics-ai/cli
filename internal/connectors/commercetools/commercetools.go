package commercetools

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
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("commercetools", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

var resources = map[string]string{"customers": "customers", "orders": "orders", "products": "products"}

func (Connector) Name() string { return "commercetools" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "commercetools", DisplayName: "commercetools", IntegrationType: "api", Description: "Reads commercetools customers, orders, and products through the HTTP API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, projectPath(cfg, "customers"), url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check commercetools: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "version", Type: "integer"}, {Name: "createdAt", Type: "timestamp"}}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{
		{Name: "customers", Description: "commercetools customers.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "orders", Description: "commercetools orders.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "products", Description: "commercetools products.", Fields: fields, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	resource, ok := resources[stream]
	if !ok {
		return fmt.Errorf("commercetools stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := intConfig(req.Config, "page_size", defaultPageSize)
	if err != nil {
		return err
	}
	maxPages, err := intConfig(req.Config, "max_pages", defaultMaxPages)
	if err != nil {
		return err
	}
	return readOffset(ctx, r, projectPath(req.Config, resource), pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readOffset(ctx context.Context, r *connsdk.Requester, resource string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; page < maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, resource, url.Values{"limit": []string{strconv.Itoa(pageSize)}, "offset": []string{strconv.Itoa(offset)}}, nil)
		if err != nil {
			return fmt.Errorf("read commercetools %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode commercetools %s: %w", resource, err)
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
		if total, ok := intAt(resp.Body, "total"); ok && offset >= total {
			return nil
		}
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "version": i, "createdAt": fmt.Sprintf("2026-01-0%dT00:00:00Z", i), "fixture": true}); err != nil {
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
	if err := requireCredentials(cfg); err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: &connsdk.OAuth2ClientCredentials{TokenURL: tokenURL(cfg), ClientID: secretOrConfig(cfg, "client_id"), ClientSecret: secretOrConfig(cfg, "client_secret"), Client: c.Client}, UserAgent: userAgent}, nil
}

func projectPath(cfg connectors.RuntimeConfig, resource string) string {
	project := strings.TrimSpace(cfg.Config["project_key"])
	if project == "" {
		project = "project"
	}
	return url.PathEscape(project) + "/" + resource
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		region := strings.TrimSpace(cfg.Config["region"])
		host := strings.TrimSpace(cfg.Config["host"])
		if region == "" || host == "" {
			return "", errors.New("commercetools connector requires config base_url or host and region")
		}
		base = "https://api." + region + "." + host + ".commercetools.com"
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("commercetools config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("commercetools config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("commercetools config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func tokenURL(cfg connectors.RuntimeConfig) string {
	if raw := strings.TrimSpace(cfg.Config["token_url"]); raw != "" {
		return raw
	}
	region := strings.TrimSpace(cfg.Config["region"])
	host := strings.TrimSpace(cfg.Config["host"])
	if region == "" || host == "" {
		return ""
	}
	return "https://auth." + region + "." + host + ".commercetools.com/oauth/token"
}

func requireCredentials(cfg connectors.RuntimeConfig) error {
	if strings.TrimSpace(secretOrConfig(cfg, "client_id")) == "" {
		return errors.New("commercetools connector requires secret client_id")
	}
	if strings.TrimSpace(secretOrConfig(cfg, "client_secret")) == "" {
		return errors.New("commercetools connector requires secret client_secret")
	}
	if strings.TrimSpace(tokenURL(cfg)) == "" {
		return errors.New("commercetools connector requires config token_url or host and region")
	}
	return nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("commercetools config %s must be a positive integer", key)
	}
	return value, nil
}

func intAt(body []byte, path string) (int, bool) {
	raw, err := connsdk.StringAt(body, path)
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0, false
	}
	value, err := strconv.Atoi(raw)
	return value, err == nil
}

func secretOrConfig(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets != nil && cfg.Secrets[key] != "" {
		return cfg.Secrets[key]
	}
	return cfg.Config[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
