// Package railz implements a read-only Railz REST connector over core list
// endpoints. It accepts a pre-issued bearer access token; OAuth token exchange is
// intentionally left to callers so the package stays dependency-free.
package railz

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
	railzDefaultBaseURL  = "https://api.railz.ai/v1"
	railzDefaultPageSize = 100
	railzMaxPageSize     = 500
	railzUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("railz", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "railz" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "railz", DisplayName: "Railz", IntegrationType: "api", Description: "Reads Railz businesses, connections, customers, invoices, and bills through the Railz REST API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "businesses", url.Values{"limit": []string{"1"}, "offset": []string{"0"}}, nil, nil); err != nil {
		return fmt.Errorf("check railz: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: railzStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "businesses"
	}
	endpoint, ok := railzEndpoints[stream]
	if !ok {
		return fmt.Errorf("railz stream %q not found", stream)
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
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read railz %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode railz %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		offset += pageSize
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", endpoint.path, i), "uuid": fmt.Sprintf("%s_fixture_%d", endpoint.path, i), "businessUuid": fmt.Sprintf("biz_%d", i), "businessName": fmt.Sprintf("Fixture Business %d", i), "name": fmt.Sprintf("Fixture %d", i), "status": "active", "createdAt": "2026-01-01T00:00:00Z"}
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
	token := strings.TrimSpace(firstSecret(cfg, "access_token", "api_key"))
	if token == "" {
		return nil, errors.New("railz connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: railzUserAgent}, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path      string
	mapRecord func(map[string]any) connectors.Record
}

var railzEndpoints = map[string]streamEndpoint{
	"businesses":  {path: "businesses", mapRecord: businessRecord},
	"connections": {path: "connections", mapRecord: basicRecord},
	"customers":   {path: "customers", mapRecord: customerRecord},
	"invoices":    {path: "invoices", mapRecord: invoiceRecord},
	"bills":       {path: "bills", mapRecord: invoiceRecord},
}

func railzStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "businesses", Description: "Railz businesses.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields("id", "name", "status", "created_at")},
		{Name: "connections", Description: "Railz data connections.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: fields("id", "business_id", "status", "created_at")},
		{Name: "customers", Description: "Railz accounting customers.", PrimaryKey: []string{"id"}, Fields: fields("id", "business_id", "name", "email")},
		{Name: "invoices", Description: "Railz invoices.", PrimaryKey: []string{"id"}, Fields: fields("id", "business_id", "customer_id", "total_amount", "status")},
		{Name: "bills", Description: "Railz bills.", PrimaryKey: []string{"id"}, Fields: fields("id", "business_id", "vendor_id", "total_amount", "status")},
	}
}

func businessRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "businessUuid", "uuid", "id"), "name": first(item, "businessName", "name"), "status": first(item, "status"), "created_at": first(item, "createdAt", "created_at")}
}
func basicRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "connectionUuid", "uuid", "id"), "business_id": first(item, "businessUuid", "business_id"), "status": first(item, "status"), "created_at": first(item, "createdAt", "created_at")}
}
func customerRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "customerUuid", "uuid", "id"), "business_id": first(item, "businessUuid", "business_id"), "name": first(item, "name", "displayName"), "email": first(item, "email")}
}
func invoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "invoiceUuid", "billUuid", "uuid", "id"), "business_id": first(item, "businessUuid", "business_id"), "customer_id": first(item, "customerUuid", "customer_id"), "vendor_id": first(item, "vendorUuid", "vendor_id"), "total_amount": first(item, "totalAmount", "amount"), "status": first(item, "status")}
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

func firstSecret(cfg connectors.RuntimeConfig, keys ...string) string {
	for _, key := range keys {
		if cfg.Secrets != nil && strings.TrimSpace(cfg.Secrets[key]) != "" {
			return cfg.Secrets[key]
		}
	}
	return ""
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return railzDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("railz config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("railz config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("railz config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return railzDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > railzMaxPageSize {
		return 0, fmt.Errorf("railz config page_size must be between 1 and %d", railzMaxPageSize)
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
		return 0, errors.New("railz config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
