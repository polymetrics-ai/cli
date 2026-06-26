// Package shippo implements a read-only native Shippo connector.
package shippo

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
	shippoDefaultBaseURL  = "https://api.goshippo.com"
	shippoDefaultPageSize = 100
	shippoMaxPageSize     = 200
	shippoUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("shippo", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "shippo" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "shippo", DisplayName: "Shippo", IntegrationType: "api", Description: "Reads Shippo addresses, parcels, shipments, and transactions through the Shippo REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	q := url.Values{"results": []string{"1"}, "page": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, shippoEndpoints["addresses"].path, q, nil, nil); err != nil {
		return fmt.Errorf("check shippo: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams(shippoEndpoints)}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "addresses"
	}
	endpoint, ok := shippoEndpoints[stream]
	if !ok {
		return fmt.Errorf("shippo stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, emit)
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

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	nextURL := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := endpoint.path
		query := url.Values{"results": []string{strconv.Itoa(pageSize)}}
		if nextURL == "" {
			query.Set("page", "1")
		} else {
			path = nextURL
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read shippo %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode shippo %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode shippo %s next: %w", endpoint.path, err)
		}
		if strings.TrimSpace(next) == "" || len(records) == 0 {
			return nil
		}
		nextURL = next
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, shippoDefaultBaseURL, "shippo")
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(secret(cfg, "api_token"))
	if token == "" {
		return nil, errors.New("shippo connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", token, "ShippoToken "), UserAgent: shippoUserAgent}, nil
}

type streamEndpoint struct {
	path        string
	description string
	fields      []string
	mapRecord   func(map[string]any) connectors.Record
}

var shippoEndpoints = map[string]streamEndpoint{
	"addresses":    {path: "addresses", description: "Shippo address book records.", fields: []string{"id", "name", "email", "updated_at"}, mapRecord: shippoRecord},
	"parcels":      {path: "parcels", description: "Shippo parcel templates and parcel records.", fields: []string{"id", "name", "status", "updated_at"}, mapRecord: shippoRecord},
	"shipments":    {path: "shipments", description: "Shippo shipments.", fields: []string{"id", "status", "updated_at"}, mapRecord: shippoRecord},
	"transactions": {path: "transactions", description: "Shippo label transactions.", fields: []string{"id", "status", "updated_at"}, mapRecord: shippoRecord},
}

func shippoRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "object_id", "id"), "name": first(item, "name", "object_owner"), "email": item["email"], "status": item["status"], "updated_at": first(item, "updated_at", "object_updated")}
}

func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := endpoint.mapRecord(map[string]any{"object_id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "status": "ok", "updated_at": "2026-01-01T00:00:00Z"})
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func streams(endpoints map[string]streamEndpoint) []connectors.Stream {
	order := []string{"addresses", "parcels", "shipments", "transactions"}
	out := make([]connectors.Stream, 0, len(order))
	for _, name := range order {
		ep := endpoints[name]
		out = append(out, connectors.Stream{Name: name, Description: ep.description, PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields(ep.fields...)})
	}
	return out
}

func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func baseURL(cfg connectors.RuntimeConfig, fallback, name string) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%s config base_url must use http or https", name)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["limit"])
	}
	if raw == "" {
		return shippoDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > shippoMaxPageSize {
		return 0, fmt.Errorf("shippo config page_size must be between 1 and %d", shippoMaxPageSize)
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
		return 0, errors.New("shippo config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
