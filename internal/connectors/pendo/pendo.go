// Package pendo implements a conservative read-only native Pendo HTTP API
// connector. Pendo's reporting APIs vary by endpoint; this connector keeps to
// simple REST list endpoints and fixture-backed conformance without aggregation
// writes.
package pendo

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
	pendoDefaultBaseURL = "https://app.pendo.io/api/v1"
	pendoDefaultLimit   = 100
	pendoMaxLimit       = 500
	pendoUserAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("pendo", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "pendo" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "pendo",
		DisplayName:     "Pendo",
		IntegrationType: "api",
		Description:     "Reads Pendo visitors, accounts, pages, and features through REST list endpoints.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

type streamEndpoint struct {
	resource     string
	description  string
	fields       []connectors.Field
	cursorFields []string
}

var pendoStreamOrder = []string{"visitors", "accounts", "pages", "features"}

var pendoStreams = map[string]streamEndpoint{
	"visitors": {
		resource: "visitor", description: "Pendo visitors.", cursorFields: []string{"lastVisit"},
		fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "email", Type: "string"}, {Name: "accountId", Type: "string"}, {Name: "lastVisit", Type: "string"}},
	},
	"accounts": {
		resource: "account", description: "Pendo accounts.", cursorFields: []string{"lastVisit"},
		fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "lastVisit", Type: "string"}},
	},
	"pages": {
		resource: "page", description: "Pendo pages.", cursorFields: []string{"lastUpdated"},
		fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "lastUpdated", Type: "string"}},
	},
	"features": {
		resource: "feature", description: "Pendo features.", cursorFields: []string{"lastUpdated"},
		fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "lastUpdated", Type: "string"}},
	},
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
	if err := r.DoJSON(ctx, http.MethodGet, "visitor", url.Values{"limit": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check pendo: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	streams := make([]connectors.Stream, 0, len(pendoStreamOrder))
	for _, name := range pendoStreamOrder {
		endpoint := pendoStreams[name]
		streams = append(streams, connectors.Stream{Name: name, Description: endpoint.description, Fields: endpoint.fields, PrimaryKey: []string{"id"}, CursorFields: endpoint.cursorFields})
	}
	return connectors.Catalog{Connector: "pendo", Streams: streams}, nil
}

func (Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "visitors"
	}
	endpoint, ok := pendoStreams[stream]
	if !ok {
		return fmt.Errorf("pendo stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, req, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	limit, err := pageSize(req.Config, "limit", pendoDefaultLimit, pendoMaxLimit, "pendo")
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config, "pendo")
	if err != nil {
		return err
	}
	return harvestPages(ctx, r, endpoint, limit, maxPages, emit)
}

func harvestPages(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, limit, maxPages int, emit func(connectors.Record) error) error {
	pageToken := "1"
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		query := url.Values{"limit": []string{strconv.Itoa(limit)}, "page": []string{pageToken}}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read pendo %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode pendo %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(mapRecord(endpoint.fields, item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode pendo %s next token: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "email": fmt.Sprintf("fixture+%d@example.com", i), "accountId": "account_fixture", "name": fmt.Sprintf("Fixture %d", i), "lastVisit": "2026-01-01T00:00:00Z", "lastUpdated": "2026-01-01T00:00:00Z"}
		record := mapRecord(endpoint.fields, item)
		if cursor := req.State[connsdk.CursorStateKey]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, pendoDefaultBaseURL, "pendo")
	if err != nil {
		return nil, err
	}
	integrationKey := secret(cfg, "integration_key")
	if strings.TrimSpace(integrationKey) == "" {
		return nil, errors.New("pendo connector requires secret integration_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("x-pendo-integration-key", integrationKey, ""), UserAgent: pendoUserAgent}, nil
}

func mapRecord(fields []connectors.Field, item map[string]any) connectors.Record {
	record := connectors.Record{}
	for _, field := range fields {
		record[field.Name] = item[field.Name]
	}
	return record
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

func baseURL(cfg connectors.RuntimeConfig, fallback, connector string) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return fallback, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", connector, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", connector, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", connector)
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig, key string, fallback, max int, connector string) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config %s must be an integer: %w", connector, key, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s config %s must be between 1 and %d", connector, key, max)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig, connector string) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer, all, or unlimited: %w", connector, err)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s config max_pages must be 0 for unlimited or a positive integer", connector)
	}
	return value, nil
}
