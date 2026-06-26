package commcare

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
	defaultBaseURL  = "https://www.commcarehq.org"
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("commcare", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

var paths = map[string]string{"forms": "form", "cases": "case"}

func (Connector) Name() string { return "commcare" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "commcare", DisplayName: "CommCare", IntegrationType: "api", Description: "Reads CommCare forms and cases through the CommCare HQ API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, resource(cfg, "form"), url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check commcare: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "received_on", Type: "timestamp"}, {Name: "server_modified_on", Type: "timestamp"}}
	return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{
		{Name: "forms", Description: "CommCare forms.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"received_on"}},
		{Name: "cases", Description: "CommCare cases.", Fields: fields, PrimaryKey: []string{"id"}, CursorFields: []string{"server_modified_on"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "forms"
	}
	kind, ok := paths[stream]
	if !ok {
		return fmt.Errorf("commcare stream %q not found", stream)
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
	return readPages(ctx, r, resource(req.Config, kind), pageSize, maxPages, req.Config, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readPages(ctx context.Context, r *connsdk.Requester, first string, pageSize, maxPages int, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	path := first
	query := url.Values{"limit": []string{strconv.Itoa(pageSize)}, "offset": []string{"0"}}
	if appID := strings.TrimSpace(cfg.Config["app_id"]); appID != "" {
		query.Set("app_id", appID)
	}
	for i := 0; i < maxPages; i++ {
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read commcare %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "objects")
		if err != nil {
			return fmt.Errorf("decode commcare %s: %w", path, err)
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "meta.next")
		if err != nil || strings.TrimSpace(next) == "" {
			return err
		}
		path = next
		query = nil
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "received_on": fmt.Sprintf("2026-01-0%dT00:00:00Z", i), "server_modified_on": fmt.Sprintf("2026-01-0%dT00:00:00Z", i), "fixture": true}); err != nil {
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
	key := secret(cfg, "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("commcare connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", key, "ApiKey "), UserAgent: userAgent}, nil
}

func resource(cfg connectors.RuntimeConfig, kind string) string {
	space := strings.Trim(strings.TrimSpace(cfg.Config["project_space"]), "/")
	if space == "" {
		space = "project"
	}
	return "a/" + url.PathEscape(space) + "/api/v0.5/" + kind + "/"
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("commcare config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("commcare config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("commcare config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("commcare config %s must be a positive integer", key)
	}
	return value, nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
