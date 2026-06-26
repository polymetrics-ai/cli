// Package perk implements a read-only Perk/TravelPerk API connector. The public
// docs currently span Perk and TravelPerk hosts; this package conservatively
// reads documented TravelPerk list endpoints using Authorization: ApiKey.
package perk

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
	defaultBaseURL  = "https://api.travelperk.com"
	defaultPageSize = 50
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init()                     { connectors.RegisterFactory("perk", New) }
func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "perk" }
func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "perk", DisplayName: "Perk", IntegrationType: "api", Description: "Reads Perk/TravelPerk trips and invoices through read-only REST list endpoints.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "trips", url.Values{"limit": []string{"1"}, "offset": []string{"0"}}, nil, nil); err != nil {
		return fmt.Errorf("check perk: %w", err)
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
		stream = "trips"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("perk stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{}
	if start := firstNonEmpty(req.State["cursor"], req.Config.Config["start_date"]); start != "" {
		base.Set(endpoint.startParam, start)
	}
	p := &connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam: "offset", PageSize: size}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, p, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error { return emit(connectors.Record(rec)) })
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ resource, recordsPath, startParam string }

var streamEndpoints = map[string]streamEndpoint{"trips": {resource: "trips", recordsPath: "trips", startParam: "modified_gte"}, "invoices": {resource: "invoices", recordsPath: "invoices", startParam: "issuing_date_gte"}}

func streams() []connectors.Stream {
	return []connectors.Stream{{Name: "trips", Description: "Perk trips listed by modified time.", PrimaryKey: []string{"id"}, CursorFields: []string{"modified"}, Fields: commonFields()}, {Name: "invoices", Description: "Perk invoices.", PrimaryKey: []string{"serial_number"}, CursorFields: []string{"issuing_date"}, Fields: []connectors.Field{{Name: "serial_number", Type: "string"}, {Name: "status", Type: "string"}, {Name: "issuing_date", Type: "date"}, {Name: "total", Type: "string"}}}}
}
func commonFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "string"}, {Name: "trip_name", Type: "string"}, {Name: "status", Type: "string"}, {Name: "modified", Type: "timestamp"}}
}
func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "trip_name": fmt.Sprintf("Fixture %s %d", stream, i), "modified": "2026-01-01T00:00:00Z", "serial_number": fmt.Sprintf("INV-%d", i)}); err != nil {
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
	key := secret(cfg, "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("perk connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyHeader("Authorization", key, "ApiKey "), DefaultHeaders: map[string]string{"Api-Version": "1"}, UserAgent: userAgent}, nil
}
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validatedBaseURL("perk", cfg.Config["base_url"], defaultBaseURL)
}
func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt("perk", cfg.Config["page_size"], defaultPageSize, maxPageSize, "page_size")
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return optionalInt("perk", cfg.Config["max_pages"], "max_pages")
}
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
func validatedBaseURL(connector, raw, def string) (string, error) {
	base := strings.TrimSpace(raw)
	if base == "" {
		return def, nil
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", connector, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", connector, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", connector)
	}
	return strings.TrimRight(base, "/"), nil
}
func boundedInt(connector, raw string, def, max int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v < 1 || v > max {
		return 0, fmt.Errorf("%s config %s must be an integer between 1 and %d", connector, name, max)
	}
	return v, nil
}
func optionalInt(connector, raw, name string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, fmt.Errorf("%s config %s must be 0, positive, all, or unlimited", connector, name)
	}
	return v, nil
}
