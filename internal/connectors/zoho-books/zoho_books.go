// Package zohobooks implements a read-only native Zoho Books API connector.
package zohobooks

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
	connectorName  = "zoho-books"
	defaultBaseURL = "https://www.zohoapis.com/books/v3"
	userAgent      = "polymetrics-go-cli"
	fixtureTime    = "2026-01-01T00:00:00Z"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Zoho Books", IntegrationType: "api", Description: "Reads Zoho Books contacts, invoices, and items through the Zoho Books REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

type streamEndpoint struct {
	path        string
	recordsPath string
	desc        string
	idKeys      []string
	nameKeys    []string
	cursorKeys  []string
}

var streamOrder = []string{"contacts", "invoices", "items"}

var streams = map[string]streamEndpoint{
	"contacts": {path: "contacts", recordsPath: "contacts", desc: "Zoho Books contacts.", idKeys: []string{"contact_id", "id"}, nameKeys: []string{"contact_name", "customer_name", "name"}, cursorKeys: []string{"last_modified_time", "updated_time", "updated_at"}},
	"invoices": {path: "invoices", recordsPath: "invoices", desc: "Zoho Books invoices.", idKeys: []string{"invoice_id", "id"}, nameKeys: []string{"invoice_number", "number"}, cursorKeys: []string{"last_modified_time", "updated_time", "updated_at"}},
	"items":    {path: "items", recordsPath: "items", desc: "Zoho Books items.", idKeys: []string{"item_id", "id"}, nameKeys: []string{"name", "item_name"}, cursorKeys: []string{"last_modified_time", "updated_time", "updated_at"}},
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
	if _, err := r.Do(ctx, http.MethodGet, streams[streamOrder[0]].path, baseQuery(cfg, 1, 1), nil); err != nil {
		return fmt.Errorf("check %s: %w", connectorName, err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	out := make([]connectors.Stream, 0, len(streamOrder))
	for _, name := range streamOrder {
		ep := streams[name]
		out = append(out, connectors.Stream{Name: name, Description: ep.desc, Fields: catalogFields(), PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}})
	}
	return connectors.Catalog{Connector: connectorName, Streams: out}, nil
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
		stream = streamOrder[0]
	}
	ep, ok := streams[stream]
	if !ok {
		return fmt.Errorf("%s stream %q not found", connectorName, stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, ep, req, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	max, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return harvest(ctx, r, ep, req.Config, size, max, emit)
}

func harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, cfg connectors.RuntimeConfig, size, max int, emit func(connectors.Record) error) error {
	for page := 1; max == 0 || page <= max; page++ {
		resp, err := r.Do(ctx, http.MethodGet, ep.path, baseQuery(cfg, page, size), nil)
		if err != nil {
			return fmt.Errorf("read %s %s: %w", connectorName, ep.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, ep.recordsPath)
		if err != nil {
			return fmt.Errorf("decode %s %s: %w", connectorName, ep.path, err)
		}
		for _, rec := range records {
			if err := emit(mapRecord(ep, rec)); err != nil {
				return err
			}
		}
		if len(records) < size {
			return nil
		}
	}
	return nil
}

func readFixture(ctx context.Context, stream string, ep streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := mapRecord(ep, map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i), "status": "fixture", "updated_at": fixtureTime})
		if cursor := connsdk.Cursor(req.State); cursor != "" {
			rec["previous_cursor"] = cursor
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WriteResult{RecordsFailed: len(records)}, err
	}
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "access_token")
	if token == "" {
		return nil, errors.New("zoho-books connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", token, "Zoho-oauthtoken "), UserAgent: userAgent}, nil
}

func mapRecord(ep streamEndpoint, in map[string]any) connectors.Record {
	out := connectors.Record{}
	for k, v := range in {
		out[k] = v
	}
	if out["id"] == nil {
		out["id"] = firstValue(in, ep.idKeys)
	}
	if out["name"] == nil {
		out["name"] = firstValue(in, ep.nameKeys)
	}
	if out["updated_at"] == nil {
		out["updated_at"] = firstValue(in, ep.cursorKeys)
	}
	return out
}

func firstValue(in map[string]any, keys []string) any {
	for _, key := range keys {
		if value, ok := in[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func catalogFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "updated_at", Type: "timestamp"}}
}

func baseQuery(cfg connectors.RuntimeConfig, page, size int) url.Values {
	q := url.Values{"page": []string{strconv.Itoa(page)}, "per_page": []string{strconv.Itoa(size)}}
	if org := configValue(cfg, "organization_id"); org != "" {
		q.Set("organization_id", org)
	}
	return q
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(configValue(cfg, "mode"), "fixture")
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Secrets[name])
}

func configValue(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Config == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Config[name])
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := configValue(cfg, "base_url")
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url is invalid", connectorName)
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := configValue(cfg, "page_size")
	if raw == "" {
		return 200, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 200 {
		return 0, fmt.Errorf("%s config page_size must be between 1 and 200", connectorName)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.ToLower(configValue(cfg, "max_pages"))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s config max_pages must be a non-negative integer", connectorName)
	}
	return value, nil
}
