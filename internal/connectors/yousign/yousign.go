// Package yousign implements a read-only native Yousign API connector.
package yousign

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName  = "yousign"
	defaultBaseURL = "https://api.yousign.app/v3"
	userAgent      = "polymetrics-go-cli"
	fixtureTime    = "2026-01-01T00:00:00Z"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Yousign", IntegrationType: "api", Description: "Reads Yousign signature requests, contacts, and documents through the Yousign REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

type streamEndpoint struct {
	path        string
	recordsPath string
	desc        string
	idKeys      []string
	nameKeys    []string
	cursorKeys  []string
}

var streamOrder = []string{"signature_requests", "contacts", "documents"}

var streamEndpoints = map[string]streamEndpoint{
	"signature_requests": {path: "signature_requests", recordsPath: "data", desc: "Yousign signature requests.", idKeys: []string{"id"}, nameKeys: []string{"name"}, cursorKeys: []string{"created_at", "updated_at"}},
	"contacts":           {path: "contacts", recordsPath: "data", desc: "Yousign contacts.", idKeys: []string{"id"}, nameKeys: []string{"name", "email"}, cursorKeys: []string{"created_at", "updated_at"}},
	"documents":          {path: "documents", recordsPath: "data", desc: "Yousign documents.", idKeys: []string{"id"}, nameKeys: []string{"name", "filename"}, cursorKeys: []string{"created_at", "updated_at"}},
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
	if _, err := r.Do(ctx, http.MethodGet, streamEndpoints[streamOrder[0]].path, url.Values{"limit": []string{"1"}}, nil); err != nil {
		return fmt.Errorf("check %s: %w", connectorName, err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	streams := make([]connectors.Stream, 0, len(streamOrder))
	for _, name := range streamOrder {
		ep := streamEndpoints[name]
		streams = append(streams, connectors.Stream{Name: name, Description: ep.desc, Fields: catalogFields(), PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}})
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = streamOrder[0]
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("%s stream %q not found", connectorName, stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, ep, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, ep.path, baseQuery(req.Config), nil)
	if err != nil {
		return fmt.Errorf("read %s %s: %w", connectorName, stream, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, ep.recordsPath)
	if err != nil {
		return fmt.Errorf("decode %s %s: %w", connectorName, stream, err)
	}
	for _, rec := range records {
		if err := emit(mapRecord(ep, rec)); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "api_key")
	if token == "" {
		return nil, errors.New("yousign connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func readFixture(ctx context.Context, stream string, ep streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := mapRecord(ep, map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i), "status": "fixture", "created_at": fixtureTime})
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
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

func baseQuery(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	if limit := configValue(cfg, "limit"); limit != "" {
		q.Set("limit", limit)
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
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https", connectorName)
	}
	return strings.TrimRight(base, "/"), nil
}
