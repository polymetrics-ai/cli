// Package rentcast implements a read-only RentCast connector over property,
// listing, and market endpoints. RentCast list endpoints are modeled with
// offset/limit pagination and X-Api-Key authentication.
package rentcast

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
	rentcastDefaultBaseURL  = "https://api.rentcast.io/v1"
	rentcastDefaultPageSize = 100
	rentcastMaxPageSize     = 500
	rentcastUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("rentcast", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "rentcast" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "rentcast", DisplayName: "RentCast", IntegrationType: "api", Description: "Reads RentCast properties, sale listings, rental listings, and market data. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "properties", url.Values{"limit": []string{"1"}, "offset": []string{"0"}}, nil, nil); err != nil {
		return fmt.Errorf("check rentcast: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: rentcastStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "properties"
	}
	endpoint, ok := rentcastEndpoints[stream]
	if !ok {
		return fmt.Errorf("rentcast stream %q not found", stream)
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
	return c.harvest(ctx, r, endpoint, req.Config, pageSize, maxPages, emit)
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := rentcastFilters(cfg)
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read rentcast %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode rentcast %s: %w", endpoint.path, err)
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
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", endpoint.path, i), "formattedAddress": fmt.Sprintf("%d Fixture St", i), "propertyType": "Single Family", "price": int64(100000 * i), "rent": int64(1000 * i), "city": "Fixture City", "state": "CA", "zipCode": "90000", "lastSeenDate": "2026-01-01"}
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
		return nil, errors.New("rentcast connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("X-Api-Key", key, ""), UserAgent: rentcastUserAgent}, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path      string
	mapRecord func(map[string]any) connectors.Record
}

var rentcastEndpoints = map[string]streamEndpoint{
	"properties":       {path: "properties", mapRecord: propertyRecord},
	"sale_listings":    {path: "listings/sale", mapRecord: listingRecord},
	"rental_listings":  {path: "listings/rental/long-term", mapRecord: listingRecord},
	"markets":          {path: "markets", mapRecord: marketRecord},
	"value_estimates":  {path: "avm/value", mapRecord: estimateRecord},
	"rental_estimates": {path: "avm/rent/long-term", mapRecord: estimateRecord},
}

func rentcastStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "properties", Description: "RentCast property records.", PrimaryKey: []string{"id"}, CursorFields: []string{"last_seen_date"}, Fields: fields("id", "address", "property_type", "city", "state", "zip_code", "last_seen_date")},
		{Name: "sale_listings", Description: "RentCast sale listings.", PrimaryKey: []string{"id"}, CursorFields: []string{"last_seen_date"}, Fields: fields("id", "address", "property_type", "price", "last_seen_date")},
		{Name: "rental_listings", Description: "RentCast long-term rental listings.", PrimaryKey: []string{"id"}, CursorFields: []string{"last_seen_date"}, Fields: fields("id", "address", "property_type", "rent", "last_seen_date")},
		{Name: "markets", Description: "RentCast market records.", PrimaryKey: []string{"id"}, Fields: fields("id", "city", "state", "zip_code")},
		{Name: "value_estimates", Description: "RentCast value estimate records.", PrimaryKey: []string{"id"}, Fields: fields("id", "address", "price")},
		{Name: "rental_estimates", Description: "RentCast rental estimate records.", PrimaryKey: []string{"id"}, Fields: fields("id", "address", "rent")},
	}
}

func propertyRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "address": first(item, "formattedAddress", "address"), "property_type": item["propertyType"], "city": item["city"], "state": item["state"], "zip_code": item["zipCode"], "last_seen_date": first(item, "lastSeenDate", "updated_at")}
}
func listingRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "address": first(item, "formattedAddress", "address"), "property_type": item["propertyType"], "price": item["price"], "rent": item["rent"], "last_seen_date": first(item, "lastSeenDate", "updated_at")}
}
func marketRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "zipCode"), "city": item["city"], "state": item["state"], "zip_code": item["zipCode"]}
}
func estimateRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "formattedAddress", "address"), "address": first(item, "formattedAddress", "address"), "price": first(item, "price", "value"), "rent": first(item, "rent", "value")}
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

func rentcastFilters(cfg connectors.RuntimeConfig) url.Values {
	query := url.Values{}
	for _, key := range []string{"address", "city", "state", "zipCode", "propertyType"} {
		if value := strings.TrimSpace(cfg.Config[key]); value != "" {
			query.Set(key, value)
		}
	}
	return query
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
		return rentcastDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("rentcast config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("rentcast config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("rentcast config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return rentcastDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > rentcastMaxPageSize {
		return 0, fmt.Errorf("rentcast config page_size must be between 1 and %d", rentcastMaxPageSize)
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
		return 0, errors.New("rentcast config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
