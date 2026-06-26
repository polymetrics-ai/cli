// Package zohodesk implements a read-only native Zoho Desk API connector.
package zohodesk

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
	connectorName  = "zoho-desk"
	defaultBaseURL = "https://desk.zoho.com/api/v1"
	userAgent      = "polymetrics-go-cli"
	fixtureTime    = "2026-01-01T00:00:00Z"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Zoho Desk", IntegrationType: "api", Description: "Reads Zoho Desk tickets, contacts, and accounts through the Zoho Desk REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

type streamEndpoint struct {
	path        string
	recordsPath string
	desc        string
	idKeys      []string
	nameKeys    []string
	cursorKeys  []string
}

var streamOrder = []string{"tickets", "contacts", "accounts"}

var streams = map[string]streamEndpoint{
	"tickets":  {path: "tickets", recordsPath: "data", desc: "Zoho Desk tickets.", idKeys: []string{"id"}, nameKeys: []string{"subject", "ticketNumber", "name"}, cursorKeys: []string{"modifiedTime", "updated_at"}},
	"contacts": {path: "contacts", recordsPath: "data", desc: "Zoho Desk contacts.", idKeys: []string{"id"}, nameKeys: []string{"lastName", "firstName", "name"}, cursorKeys: []string{"modifiedTime", "updated_at"}},
	"accounts": {path: "accounts", recordsPath: "data", desc: "Zoho Desk accounts.", idKeys: []string{"id"}, nameKeys: []string{"accountName", "name"}, cursorKeys: []string{"modifiedTime", "updated_at"}},
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
	if _, err := r.Do(ctx, http.MethodGet, streams[streamOrder[0]].path, pageQuery(1, 1), nil); err != nil {
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
	return harvest(ctx, r, ep, size, max, emit)
}

func harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, size, max int, emit func(connectors.Record) error) error {
	for page := 1; max == 0 || page <= max; page++ {
		resp, err := r.Do(ctx, http.MethodGet, ep.path, pageQuery(page, size), nil)
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
		rec := mapRecord(ep, map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i), "subject": "Fixture", "updated_at": fixtureTime})
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
		return nil, errors.New("zoho-desk connector requires secret access_token")
	}
	headers := map[string]string{}
	if orgID := configValue(cfg, "org_id"); orgID != "" {
		headers["orgId"] = orgID
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", token, "Zoho-oauthtoken "), UserAgent: userAgent, DefaultHeaders: headers}, nil
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
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "subject", Type: "string"}, {Name: "updated_at", Type: "timestamp"}}
}

func pageQuery(page, size int) url.Values {
	return url.Values{"from": []string{strconv.Itoa((page - 1) * size)}, "limit": []string{strconv.Itoa(size)}, "page": []string{strconv.Itoa(page)}, "per_page": []string{strconv.Itoa(size)}}
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
		return 100, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 100 {
		return 0, fmt.Errorf("%s config page_size must be between 1 and 100", connectorName)
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
