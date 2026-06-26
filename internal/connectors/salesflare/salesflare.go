package salesflare

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
	connectorName   = "salesflare"
	defaultBaseURL  = "https://api.salesflare.com"
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamSpec struct {
	path string
	desc string
}

var streams = map[string]streamSpec{
	"accounts":      {path: "accounts", desc: "Salesflare accounts."},
	"contacts":      {path: "contacts", desc: "Salesflare contacts."},
	"opportunities": {path: "opportunities", desc: "Salesflare opportunities."},
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Salesflare", IntegrationType: "api", Description: "Reads Salesflare accounts, contacts, and opportunities through the Salesflare API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, streams["accounts"].path, url.Values{"page": []string{"1"}, "limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check salesflare: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "accounts", Description: streams["accounts"].desc, Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "contacts", Description: streams["contacts"].desc, Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "opportunities", Description: streams["opportunities"].desc, Fields: fields, PrimaryKey: []string{"id"}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "accounts"
	}
	spec, ok := streams[stream]
	if !ok {
		return fmt.Errorf("salesflare stream %q not found", stream)
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
	query := url.Values{"page": []string{"1"}, "limit": []string{strconv.Itoa(pageSize)}}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read salesflare %s: %w", spec.path, err)
		}
		records, err := recordsAtAny(resp.Body, "data", "")
		if err != nil {
			return fmt.Errorf("decode salesflare %s: %w", spec.path, err)
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		next, err := firstStringAt(resp.Body, "pagination.next_page", "next_page", "next")
		if err != nil {
			return err
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		if strings.HasPrefix(next, "http://") || strings.HasPrefix(next, "https://") || strings.HasPrefix(next, "/") {
			path = next
			query = nil
			continue
		}
		query.Set("page", next)
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
		if err := emit(connectors.Record{"id": i, "name": fmt.Sprintf("Fixture %s %d", stream, i), "fixture": true}); err != nil {
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
		return nil, errors.New("salesflare connector requires secret api_key")
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
		return "", fmt.Errorf("salesflare config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("salesflare config base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("salesflare config base_url must include a host")
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
		return 0, fmt.Errorf("salesflare config %s must be a non-negative integer", key)
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
