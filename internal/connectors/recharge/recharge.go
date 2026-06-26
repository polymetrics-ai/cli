// Package recharge implements a read-only Recharge API connector.
package recharge

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
	rechargeName            = "recharge"
	rechargeDefaultBaseURL  = "https://api.rechargeapps.com"
	rechargeDefaultPageSize = 250
	rechargeMaxPageSize     = 250
	rechargeUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("recharge", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return rechargeName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: rechargeName, DisplayName: "Recharge", IntegrationType: "api", Description: "Reads Recharge customers, subscriptions, and orders through the Recharge REST API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if strings.TrimSpace(rechargeAccessToken(cfg)) == "" {
		return errors.New("recharge connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "customers", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check recharge: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: rechargeStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	endpoint, ok := rechargeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("recharge stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := rechargePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := rechargeMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint rechargeStreamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read recharge %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode recharge %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_cursor")
		if err != nil {
			return fmt.Errorf("decode recharge next_cursor: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint rechargeStreamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": i, "email": fmt.Sprintf("fixture%d@example.com", i), "status": "active", "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z", "address_id": i + 10, "customer_id": i + 100}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := rechargeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := rechargeAccessToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("recharge connector requires secret access_token")
	}
	version := strings.TrimSpace(cfg.Config["api_version"])
	if version == "" {
		version = "2021-11"
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("X-Recharge-Access-Token", token, ""), UserAgent: rechargeUserAgent, DefaultHeaders: map[string]string{"X-Recharge-Version": version}}, nil
}

type rechargeStreamEndpoint struct {
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var rechargeStreamEndpoints = map[string]rechargeStreamEndpoint{
	"customers":     {resource: "customers", recordsPath: "customers", mapRecord: rechargeCustomerRecord},
	"subscriptions": {resource: "subscriptions", recordsPath: "subscriptions", mapRecord: rechargeSubscriptionRecord},
	"orders":        {resource: "orders", recordsPath: "orders", mapRecord: rechargeOrderRecord},
}

func rechargeStreams() []connectors.Stream {
	common := []connectors.Field{{Name: "id", Type: "integer"}, {Name: "status", Type: "string"}, {Name: "created_at", Type: "string"}, {Name: "updated_at", Type: "string"}}
	return []connectors.Stream{
		{Name: "customers", Description: "Recharge customers.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "email", Type: "string"}, {Name: "created_at", Type: "string"}, {Name: "updated_at", Type: "string"}}},
		{Name: "subscriptions", Description: "Recharge subscriptions.", PrimaryKey: []string{"id"}, Fields: append(common, connectors.Field{Name: "customer_id", Type: "integer"})},
		{Name: "orders", Description: "Recharge orders.", PrimaryKey: []string{"id"}, Fields: append(common, connectors.Field{Name: "customer_id", Type: "integer"})},
	}
}

func rechargeCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "email": item["email"], "created_at": item["created_at"], "updated_at": item["updated_at"]}
}

func rechargeSubscriptionRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "status": item["status"], "customer_id": item["customer_id"], "created_at": item["created_at"], "updated_at": item["updated_at"]}
}

func rechargeOrderRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "status": item["status"], "customer_id": item["customer_id"], "created_at": item["created_at"], "updated_at": item["updated_at"]}
}

func rechargeAccessToken(cfg connectors.RuntimeConfig) string { return cfg.Secrets["access_token"] }

func rechargeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validBaseURL(rechargeName, cfg.Config["base_url"], rechargeDefaultBaseURL)
}

func rechargePageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(rechargeName, cfg.Config["page_size"], rechargeDefaultPageSize, 1, rechargeMaxPageSize)
}

func rechargeMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	return maxPagesConfig(rechargeName, cfg.Config["max_pages"])
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func validBaseURL(name, raw, fallback string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", name, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(name, raw string, fallback, min, max int) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s config value must be an integer: %w", name, err)
	}
	if value < min || value > max {
		return 0, fmt.Errorf("%s config value must be between %d and %d", name, min, max)
	}
	return value, nil
}

func maxPagesConfig(name, raw string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer, all, or unlimited: %w", name, err)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s config max_pages must be 0 for unlimited or a positive integer", name)
	}
	return value, nil
}
