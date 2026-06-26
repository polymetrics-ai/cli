// Package tremendous implements a read-only native connector for Tremendous APIs.
package tremendous

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
	defaultBaseURL  = "https://testflight.tremendous.com"
	defaultPageSize = 100
	maxPageSize     = 1000
	defaultMaxPages = 1
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("tremendous", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "tremendous" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "tremendous", DisplayName: "Tremendous", IntegrationType: "api", Description: "Reads Tremendous campaigns, orders, rewards, and funding sources through API list endpoints.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "api/v2/orders", url.Values{"limit": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check tremendous: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "orders"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("tremendous stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	return harvest(ctx, r, req.Config, spec, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func harvest(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, spec streamSpec, emit func(connectors.Record) error) error {
	pageSize, err := boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "tremendous config page_size")
	if err != nil {
		return err
	}
	maxPages, err := configuredMaxPages(cfg)
	if err != nil {
		return err
	}
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		query := url.Values{"limit": []string{strconv.Itoa(pageSize)}, "page": []string{strconv.Itoa(page)}}
		resp, err := r.Do(ctx, http.MethodGet, spec.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read tremendous %s: %w", spec.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, spec.recordsPath)
		if err != nil {
			return fmt.Errorf("decode tremendous %s: %w", spec.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(spec.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

func readFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "campaign_id": "campaign_fixture_1", "payment_status": "paid", "created_at": "2026-01-01T00:00:00Z"}
		rec := spec.mapRecord(item)
		rec["connector"] = "tremendous"
		rec["fixture"] = true
		if err := emit(rec); err != nil {
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
	token := strings.TrimSpace(cfg.Secrets["api_key"])
	if token == "" {
		return nil, errors.New("tremendous connector requires secret api_key")
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
		return "", fmt.Errorf("tremendous config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("tremendous config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("tremendous config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func configuredMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" {
		return defaultMaxPages, nil
	}
	if raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("tremendous config max_pages must be a non-negative integer: %w", err)
	}
	return value, nil
}
func boundedInt(raw string, def, max int, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s must be between 1 and %d", name, max)
	}
	return value, nil
}
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamSpec struct {
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamSpecs = map[string]streamSpec{
	"campaigns":       {resource: "api/v2/campaigns", recordsPath: "campaigns", mapRecord: namedRecord},
	"orders":          {resource: "api/v2/orders", recordsPath: "orders", mapRecord: orderRecord},
	"rewards":         {resource: "api/v2/rewards", recordsPath: "rewards", mapRecord: rewardRecord},
	"funding_sources": {resource: "api/v2/funding_sources", recordsPath: "funding_sources", mapRecord: namedRecord},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "string"}}
	return []connectors.Stream{
		{Name: "campaigns", Description: "Tremendous campaigns.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "orders", Description: "Tremendous orders.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "campaign_id", Type: "string"}, {Name: "payment_status", Type: "string"}, {Name: "created_at", Type: "string"}}},
		{Name: "rewards", Description: "Tremendous rewards.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "order_id", Type: "string"}, {Name: "status", Type: "string"}, {Name: "created_at", Type: "string"}}},
		{Name: "funding_sources", Description: "Tremendous funding sources.", PrimaryKey: []string{"id"}, Fields: fields},
	}
}

func namedRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "created_at": first(item, "created_at", "createdAt")}
}
func orderRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "campaign_id": first(item, "campaign_id", "campaignId"), "payment_status": first(item, "payment_status", "paymentStatus"), "created_at": first(item, "created_at", "createdAt")}
}
func rewardRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "order_id": first(item, "order_id", "orderId"), "status": item["status"], "created_at": first(item, "created_at", "createdAt")}
}
func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}
