package ruddr

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
	connectorName   = "ruddr"
	defaultBaseURL  = "https://api.ruddr.io"
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamSpec struct {
	resource string
	desc     string
}

var streams = map[string]streamSpec{
	"clients":      {resource: "clients", desc: "Ruddr clients."},
	"projects":     {resource: "projects", desc: "Ruddr projects."},
	"time_entries": {resource: "time_entries", desc: "Ruddr time entries."},
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Ruddr", IntegrationType: "api", Description: "Reads Ruddr clients, projects, and time entries through the Ruddr API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	path, err := pathFor(cfg, streams["projects"])
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"page": []string{"1"}, "page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check ruddr: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "project_id", Type: "string"}, {Name: "hours", Type: "number"}}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "clients", Description: streams["clients"].desc, Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "projects", Description: streams["projects"].desc, Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "time_entries", Description: streams["time_entries"].desc, Fields: fields, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "time_entries"
	}
	spec, ok := streams[stream]
	if !ok {
		return fmt.Errorf("ruddr stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := pathFor(req.Config, spec)
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
	return readPages(ctx, r, path, pageSize, maxPages, emit)
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readPages(ctx context.Context, r *connsdk.Requester, firstPath string, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := firstPath
	query := url.Values{"page": []string{"1"}, "page_size": []string{strconv.Itoa(pageSize)}}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read ruddr %s: %w", firstPath, err)
		}
		records, err := recordsAtAny(resp.Body, "results", "data", "")
		if err != nil {
			return fmt.Errorf("decode ruddr %s: %w", firstPath, err)
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		next, err := firstStringAt(resp.Body, "next", "links.next")
		if err != nil {
			return err
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path = next
		query = nil
	}
	return nil
}

func recordsAtAny(body []byte, paths ...string) ([]map[string]any, error) {
	for _, path := range paths {
		records, err := connsdk.RecordsAt(body, path)
		if err != nil || len(records) > 0 {
			return records, err
		}
	}
	return nil, nil
}

func firstStringAt(body []byte, paths ...string) (string, error) {
	for _, path := range paths {
		value, err := connsdk.StringAt(body, path)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(value) != "" {
			return value, nil
		}
	}
	return "", nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "project_id": "fixture_project", "hours": float64(i), "fixture": true}); err != nil {
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
	token := strings.TrimSpace(secret(cfg, "api_key"))
	if token == "" {
		return nil, errors.New("ruddr connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func pathFor(cfg connectors.RuntimeConfig, spec streamSpec) (string, error) {
	workspace := strings.Trim(strings.TrimSpace(cfg.Config["workspace_id"]), "/")
	if workspace == "" {
		return "", errors.New("ruddr connector requires config workspace_id")
	}
	return "api/workspaces/" + url.PathEscape(workspace) + "/" + spec.resource, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("ruddr config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("ruddr config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("ruddr config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	if strings.EqualFold(raw, "all") || strings.EqualFold(raw, "unlimited") {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("ruddr config %s must be a non-negative integer", key)
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
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
