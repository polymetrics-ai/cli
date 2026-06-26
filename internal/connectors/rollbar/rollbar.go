package rollbar

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
	connectorName   = "rollbar"
	defaultBaseURL  = "https://api.rollbar.com"
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamSpec struct {
	path        string
	recordsPath string
	desc        string
	fields      []connectors.Field
}

var streams = map[string]streamSpec{
	"items": {
		path:        "api/1/items/",
		recordsPath: "result.items",
		desc:        "Rollbar error items.",
		fields:      []connectors.Field{{Name: "id", Type: "integer"}, {Name: "title", Type: "string"}, {Name: "environment", Type: "string"}},
	},
	"projects": {
		path:        "api/1/projects",
		recordsPath: "result",
		desc:        "Rollbar projects visible to the token.",
		fields:      []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}},
	},
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Rollbar", IntegrationType: "api", Description: "Reads Rollbar projects and error items through the Rollbar API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, streams["items"].path, url.Values{"page": []string{"1"}, "per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check rollbar: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "items", Description: streams["items"].desc, Fields: streams["items"].fields, PrimaryKey: []string{"id"}},
		{Name: "projects", Description: streams["projects"].desc, Fields: streams["projects"].fields, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "items"
	}
	spec, ok := streams[stream]
	if !ok {
		return fmt.Errorf("rollbar stream %q not found", stream)
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
	return readPages(ctx, r, spec, pageSize, maxPages, emit)
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readPages(ctx context.Context, r *connsdk.Requester, spec streamSpec, pageSize, maxPages int, emit func(connectors.Record) error) error {
	page := 1
	for i := 0; maxPages == 0 || i < maxPages; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{"page": []string{strconv.Itoa(page)}, "per_page": []string{strconv.Itoa(pageSize)}}
		resp, err := r.Do(ctx, http.MethodGet, spec.path, query, nil)
		if err != nil {
			return fmt.Errorf("read rollbar %s: %w", spec.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode rollbar %s: %w", spec.path, err)
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		totalPages := intAt(resp.Body, "result.total_pages")
		currentPage := intAt(resp.Body, "result.page")
		if currentPage == 0 {
			currentPage = page
		}
		if len(records) == 0 || (totalPages > 0 && currentPage >= totalPages) {
			return nil
		}
		page = currentPage + 1
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": i, "title": fmt.Sprintf("Fixture %s %d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "fixture": true}); err != nil {
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
	token := strings.TrimSpace(secret(cfg, "access_token"))
	if token == "" {
		return nil, errors.New("rollbar connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("X-Rollbar-Access-Token", token, ""), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("rollbar config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("rollbar config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("rollbar config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("rollbar config %s must be a non-negative integer", key)
	}
	return value, nil
}

func intAt(body []byte, path string) int {
	value, _ := connsdk.StringAt(body, path)
	parsed, _ := strconv.Atoi(strings.TrimSpace(value))
	return parsed
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
