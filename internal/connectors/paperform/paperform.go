// Package paperform implements a read-only native Paperform HTTP API connector.
package paperform

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
	paperformDefaultBaseURL = "https://api.paperform.co/v1"
	paperformDefaultLimit   = 100
	paperformMaxLimit       = 100
	paperformUserAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("paperform", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "paperform" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "paperform",
		DisplayName:     "Paperform",
		IntegrationType: "api",
		Description:     "Reads Paperform forms and form submissions through the Paperform REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

type streamEndpoint struct {
	resource     string
	description  string
	fields       []connectors.Field
	cursorFields []string
}

var paperformStreamOrder = []string{"forms", "submissions"}

var paperformStreams = map[string]streamEndpoint{
	"forms": {
		resource: "forms", description: "Paperform forms.", cursorFields: []string{"created_at"},
		fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "slug", Type: "string"}, {Name: "created_at", Type: "string"}, {Name: "updated_at", Type: "string"}},
	},
	"submissions": {
		resource: "forms/{form_id}/submissions", description: "Paperform submissions for the configured form_id.", cursorFields: []string{"created_at"},
		fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "form_id", Type: "string"}, {Name: "created_at", Type: "string"}, {Name: "updated_at", Type: "string"}, {Name: "data", Type: "object"}},
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
	if err := r.DoJSON(ctx, http.MethodGet, "forms", url.Values{"limit": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check paperform: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	streams := make([]connectors.Stream, 0, len(paperformStreamOrder))
	for _, name := range paperformStreamOrder {
		endpoint := paperformStreams[name]
		streams = append(streams, connectors.Stream{Name: name, Description: endpoint.description, Fields: endpoint.fields, PrimaryKey: []string{"id"}, CursorFields: endpoint.cursorFields})
	}
	return connectors.Catalog{Connector: "paperform", Streams: streams}, nil
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
		stream = "forms"
	}
	endpoint, ok := paperformStreams[stream]
	if !ok {
		return fmt.Errorf("paperform stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, req, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resource, err := paperformResource(endpoint, req.Config)
	if err != nil {
		return err
	}
	limit, err := pageSize(req.Config, "limit", paperformDefaultLimit, paperformMaxLimit, "paperform")
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config, "paperform")
	if err != nil {
		return err
	}
	return harvestPages(ctx, r, resource, endpoint, limit, maxPages, emit)
}

func harvestPages(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, limit, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		query := url.Values{"limit": []string{strconv.Itoa(limit)}, "page": []string{strconv.Itoa(page)}}
		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read paperform %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode paperform %s page: %w", resource, err)
		}
		for _, item := range records {
			if err := emit(mapRecord(endpoint.fields, item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return fmt.Errorf("decode paperform %s has_more: %w", resource, err)
		}
		if hasMore != "true" || len(records) == 0 {
			return nil
		}
	}
	return nil
}

func paperformResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !strings.Contains(endpoint.resource, "{form_id}") {
		return endpoint.resource, nil
	}
	formID := strings.TrimSpace(cfg.Config["form_id"])
	if formID == "" {
		return "", errors.New("paperform submissions stream requires config form_id")
	}
	return strings.ReplaceAll(endpoint.resource, "{form_id}", url.PathEscape(formID)), nil
}

func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "form_id": "form_fixture", "title": fmt.Sprintf("Fixture %d", i), "slug": fmt.Sprintf("fixture-%d", i), "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z", "data": map[string]any{"fixture": true}}
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
	base, err := baseURL(cfg, paperformDefaultBaseURL, "paperform")
	if err != nil {
		return nil, err
	}
	apiKey := secret(cfg, "api_key")
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("paperform connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(apiKey), UserAgent: paperformUserAgent}, nil
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
