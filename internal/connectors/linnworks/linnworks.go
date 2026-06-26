// Package linnworks implements a conservative read-only Linnworks connector.
// Linnworks exposes several read operations as POST RPC endpoints; this package
// only calls known read-style endpoints and leaves Write unsupported.
package linnworks

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
	defaultBaseURL  = "https://eu-ext.linnworks.net/api"
	defaultPageSize = 100
	maxPageSize     = 200
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("linnworks", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "linnworks" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "linnworks", DisplayName: "Linnworks", IntegrationType: "api", Description: "Reads Linnworks inventory, open orders, warehouses, and suppliers through read-style API endpoints.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("linnworks connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	body := map[string]any{"entriesPerPage": 1, "pageNumber": 1}
	if err := r.DoJSON(ctx, http.MethodPost, "Inventory/GetInventoryItems", nil, body, nil); err != nil {
		return fmt.Errorf("check linnworks: %w", err)
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
		stream = "inventory_items"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("linnworks stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
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
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		body := map[string]any{"entriesPerPage": pageSize, "pageNumber": page}
		resp, err := r.Do(ctx, endpoint.method, endpoint.resource, nil, body)
		if err != nil {
			return fmt.Errorf("read linnworks %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode linnworks %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"ItemNumber": fmt.Sprintf("SKU-%d", i), "ItemTitle": fmt.Sprintf("Fixture %s %d", stream, i), "Quantity": i, "Name": fmt.Sprintf("Fixture %d", i), "pkStockItemId": fmt.Sprintf("item-%d", i)}
		rec := endpoint.mapRecord(item)
		rec["connector"] = "linnworks"
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
	token := secret(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("linnworks connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", token, ""), UserAgent: userAgent}, nil
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := cfg.Secrets["api_token"]; v != "" {
		return v
	}
	return cfg.Secrets["session_token"]
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("linnworks config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("linnworks config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("linnworks config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "linnworks config page_size")
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("linnworks config max_pages must be a non-negative integer: %w", err)
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
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
