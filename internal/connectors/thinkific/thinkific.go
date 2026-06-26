package thinkific

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
	connectorName  = "thinkific"
	defaultBaseURL = "https://api.thinkific.com"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Thinkific", IntegrationType: "api", Description: "Reads Thinkific courses through the public API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if strings.TrimSpace(cfg.Secrets["api_key"]) == "" || strings.TrimSpace(cfg.Config["subdomain"]) == "" {
		return errors.New("thinkific connector requires config subdomain and secret api_key")
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{{Name: "courses", Description: "Thinkific courses.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "slug", Type: "string"}, {Name: "created_at", Type: "timestamp"}}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "courses"
	}
	if stream != "courses" {
		return fmt.Errorf("thinkific stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return emitFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size := pageSize(req.Config, 100)
	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("limit", strconv.Itoa(size))
		resp, err := r.Do(ctx, http.MethodGet, "api/public/v1/courses", q, nil)
		if err != nil {
			return fmt.Errorf("read thinkific courses: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return err
		}
		for _, item := range records {
			if err := emit(connectors.Record{"id": item["id"], "name": item["name"], "slug": item["slug"], "created_at": item["created_at"]}); err != nil {
				return err
			}
		}
		if len(records) < size {
			return nil
		}
	}
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := strings.TrimSpace(cfg.Secrets["api_key"])
	subdomain := strings.TrimSpace(cfg.Config["subdomain"])
	if key == "" || subdomain == "" {
		return nil, errors.New("thinkific connector requires config subdomain and secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, DefaultHeaders: map[string]string{"X-Auth-API-Key": key, "X-Auth-Subdomain": subdomain}, UserAgent: userAgent}, nil
}

func emitFixture(ctx context.Context, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return emit(connectors.Record{"id": 1, "name": "Fixture Course", "slug": "fixture-course", "created_at": "2026-01-01T00:00:00Z", "fixture": true})
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func pageSize(cfg connectors.RuntimeConfig, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(cfg.Config["page_size"]))
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("thinkific config base_url is invalid: %w", err)
	}
	if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return "", errors.New("thinkific config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}
