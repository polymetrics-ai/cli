package safetyculture

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
	connectorName   = "safetyculture"
	defaultBaseURL  = "https://api.safetyculture.io"
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamSpec struct {
	path        string
	recordsPath string
	desc        string
}

var streams = map[string]streamSpec{
	"audits":    {path: "audits", recordsPath: "audits", desc: "SafetyCulture audits."},
	"templates": {path: "templates", recordsPath: "templates", desc: "SafetyCulture templates."},
	"users":     {path: "users", recordsPath: "users", desc: "SafetyCulture users."},
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "SafetyCulture", IntegrationType: "api", Description: "Reads SafetyCulture audits, templates, and users through the SafetyCulture API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, streams["audits"].path, url.Values{"page": []string{"1"}, "page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check safetyculture: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "modified_at", Type: "timestamp"}}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "audits", Description: streams["audits"].desc, Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "templates", Description: streams["templates"].desc, Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "users", Description: streams["users"].desc, Fields: fields, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "audits"
	}
	spec, ok := streams[stream]
	if !ok {
		return fmt.Errorf("safetyculture stream %q not found", stream)
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
	path := spec.path
	query := url.Values{"page": []string{"1"}, "page_size": []string{strconv.Itoa(pageSize)}}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read safetyculture %s: %w", spec.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode safetyculture %s: %w", spec.path, err)
		}
		for _, rec := range records {
			if err := emit(normalize(rec)); err != nil {
				return err
			}
		}
		next, err := firstStringAt(resp.Body, "next_page", "links.next")
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

func normalize(in map[string]any) connectors.Record {
	out := connectors.Record(in)
	if out["id"] == nil {
		for _, key := range []string{"audit_id", "template_id", "user_id"} {
			if value := out[key]; value != nil {
				out["id"] = value
				break
			}
		}
	}
	return out
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
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "fixture": true}); err != nil {
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
		return nil, errors.New("safetyculture connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("safetyculture config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("safetyculture config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("safetyculture config base_url must include a host")
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
		return 0, fmt.Errorf("safetyculture config %s must be a non-negative integer", key)
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
