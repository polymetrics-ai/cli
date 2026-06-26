// Package recurly implements a read-only Recurly v3 connector. It uses HTTP
// Basic auth with the API key as username and follows RFC 5988 Link pagination.
package recurly

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
	recurlyDefaultBaseURL  = "https://v3.recurly.com"
	recurlyDefaultPageSize = 200
	recurlyMaxPageSize     = 200
	recurlyUserAgent       = "polymetrics-go-cli"
	recurlyAccept          = "application/vnd.recurly.v2021-02-25"
)

func init() { connectors.RegisterFactory("recurly", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "recurly" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "recurly", DisplayName: "Recurly", IntegrationType: "api", Description: "Reads Recurly accounts, subscriptions, invoices, transactions, and plans through the v3 REST API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "accounts", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check recurly: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: recurlyStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "accounts"
	}
	endpoint, ok := recurlyEndpoints[stream]
	if !ok {
		return fmt.Errorf("recurly stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	paginator := &connsdk.LinkHeaderPaginator{FirstQuery: base}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.path, base, paginator, "data", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	}); err != nil {
		return fmt.Errorf("read recurly %s: %w", endpoint.path, err)
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", endpoint.path, i), "code": fmt.Sprintf("fixture_%d", i), "email": fmt.Sprintf("fixture+%d@example.com", i), "state": "active", "status": "active", "name": fmt.Sprintf("Fixture %d", i), "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z", "total": int64(100 * i)}
		if err := emit(endpoint.mapRecord(item)); err != nil {
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
	key := strings.TrimSpace(secret(cfg, "api_key"))
	if key == "" {
		return nil, errors.New("recurly connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(key, ""), UserAgent: recurlyUserAgent, Accept: recurlyAccept}, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path      string
	mapRecord func(map[string]any) connectors.Record
}

var recurlyEndpoints = map[string]streamEndpoint{
	"accounts":      {path: "accounts", mapRecord: accountRecord},
	"subscriptions": {path: "subscriptions", mapRecord: subscriptionRecord},
	"invoices":      {path: "invoices", mapRecord: invoiceRecord},
	"transactions":  {path: "transactions", mapRecord: transactionRecord},
	"plans":         {path: "plans", mapRecord: planRecord},
}

func recurlyStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "accounts", Description: "Recurly accounts.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "code", "email", "state", "created_at", "updated_at")},
		{Name: "subscriptions", Description: "Recurly subscriptions.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "account_id", "plan_id", "state", "created_at", "updated_at")},
		{Name: "invoices", Description: "Recurly invoices.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "account_id", "state", "total", "created_at")},
		{Name: "transactions", Description: "Recurly transactions.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields("id", "account_id", "status", "amount", "created_at")},
		{Name: "plans", Description: "Recurly plans.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "code", "name", "state", "updated_at")},
	}
}

func accountRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "code": item["code"], "email": item["email"], "state": first(item, "state", "status"), "created_at": item["created_at"], "updated_at": item["updated_at"]}
}
func subscriptionRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "account_id": nestedID(item["account"]), "plan_id": nestedID(item["plan"]), "state": first(item, "state", "status"), "created_at": item["created_at"], "updated_at": item["updated_at"]}
}
func invoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "account_id": nestedID(item["account"]), "state": first(item, "state", "status"), "total": item["total"], "created_at": item["created_at"]}
}
func transactionRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "account_id": nestedID(item["account"]), "status": first(item, "status", "state"), "amount": first(item, "amount", "total"), "created_at": item["created_at"]}
}
func planRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "code": item["code"], "name": item["name"], "state": first(item, "state", "status"), "updated_at": item["updated_at"]}
}

func nestedID(v any) any {
	if m, ok := v.(map[string]any); ok {
		return first(m, "id", "code")
	}
	return v
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return recurlyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("recurly config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("recurly config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("recurly config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return recurlyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > recurlyMaxPageSize {
		return 0, fmt.Errorf("recurly config page_size must be between 1 and %d", recurlyMaxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("recurly config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
