// Package plausible implements a read-only Plausible Analytics API connector.
package plausible

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
	defaultBaseURL  = "https://plausible.io/api/v1"
	defaultPageSize = 100
	defaultMaxPages = 3
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("plausible", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "plausible" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "plausible",
		DisplayName:     "Plausible",
		IntegrationType: "api",
		Description:     "Reads Plausible Analytics sites and stats reports through the Stats API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
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
	if _, err := r.Do(ctx, http.MethodGet, "sites", nil, nil); err != nil {
		return fmt.Errorf("check plausible: %w", err)
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
		stream = "sites"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("plausible stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, endpoint, emit)
	}
	if endpoint.requiresSite && strings.TrimSpace(req.Config.Config["site_id"]) == "" {
		return errors.New("plausible stats streams require config site_id")
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if endpoint.paginated {
		return c.readPaginated(ctx, r, endpoint, req.Config, emit)
	}
	resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query(req.Config, endpoint), nil)
	if err != nil {
		return fmt.Errorf("read plausible %s: %w", stream, err)
	}
	_, err = emitRecords(ctx, resp.Body, endpoint, emit)
	return err
}

func (c Connector) readPaginated(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	pageSize, err := pageSize(cfg)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(cfg)
	if err != nil {
		return err
	}
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		q := query(cfg, endpoint)
		q.Set("limit", strconv.Itoa(pageSize))
		q.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, q, nil)
		if err != nil {
			return fmt.Errorf("read plausible %s: %w", endpoint.path, err)
		}
		count, err := emitRecords(ctx, resp.Body, endpoint, emit)
		if err != nil {
			return err
		}
		if count < pageSize {
			return nil
		}
	}
	return nil
}

func emitRecords(ctx context.Context, body []byte, endpoint streamEndpoint, emit func(connectors.Record) error) (int, error) {
	records, err := connsdk.RecordsAt(body, endpoint.recordsPath)
	if err != nil {
		return 0, fmt.Errorf("decode plausible %s: %w", endpoint.path, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return 0, err
		}
	}
	return len(records), nil
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path         string
	recordsPath  string
	requiresSite bool
	paginated    bool
	mapRecord    func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"sites":      {path: "sites", recordsPath: "sites", mapRecord: siteRecord},
	"aggregate":  {path: "stats/aggregate", recordsPath: "results", requiresSite: true, mapRecord: aggregateRecord},
	"timeseries": {path: "stats/timeseries", recordsPath: "results", requiresSite: true, mapRecord: timeseriesRecord},
	"breakdown":  {path: "stats/breakdown", recordsPath: "results", requiresSite: true, paginated: true, mapRecord: breakdownRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "sites", Description: "Plausible sites available to the token.", PrimaryKey: []string{"site_id"}, Fields: []connectors.Field{{Name: "site_id", Type: "string"}, {Name: "domain", Type: "string"}}},
		{Name: "aggregate", Description: "Aggregate Plausible metrics for a site and period.", PrimaryKey: []string{"site_id"}, Fields: statFields()},
		{Name: "timeseries", Description: "Time-series Plausible metrics for a site and period.", PrimaryKey: []string{"date"}, Fields: append([]connectors.Field{{Name: "date", Type: "string"}}, statFields()...)},
		{Name: "breakdown", Description: "Breakdown Plausible metrics by a configured property.", PrimaryKey: []string{"property_value"}, Fields: append([]connectors.Field{{Name: "property_value", Type: "string"}}, statFields()...)},
	}
}

func statFields() []connectors.Field {
	return []connectors.Field{{Name: "site_id", Type: "string"}, {Name: "visitors", Type: "number"}, {Name: "visits", Type: "number"}, {Name: "pageviews", Type: "number"}, {Name: "events", Type: "number"}, {Name: "bounce_rate", Type: "number"}, {Name: "visit_duration", Type: "number"}}
}

func siteRecord(item map[string]any) connectors.Record {
	return connectors.Record{"site_id": first(item, "site_id", "domain"), "domain": first(item, "domain", "site_id")}
}

func aggregateRecord(item map[string]any) connectors.Record { return metricRecord(item, nil) }

func timeseriesRecord(item map[string]any) connectors.Record {
	rec := metricRecord(item, nil)
	rec["date"] = item["date"]
	return rec
}

func breakdownRecord(item map[string]any) connectors.Record {
	rec := metricRecord(item, nil)
	rec["property_value"] = first(item, "page", "source", "referrer", "utm_campaign", "country", "region", "city", "browser", "os", "device")
	return rec
}

func metricRecord(item map[string]any, base connectors.Record) connectors.Record {
	if base == nil {
		base = connectors.Record{}
	}
	for _, key := range []string{"site_id", "visitors", "visits", "pageviews", "events", "bounce_rate", "visit_duration"} {
		base[key] = metricValue(item[key])
	}
	return base
}

func metricValue(v any) any {
	if obj, ok := v.(map[string]any); ok {
		if value, ok := obj["value"]; ok {
			return value
		}
	}
	return v
}

func readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"site_id": "example.com", "domain": "example.com", "date": fmt.Sprintf("2026-01-0%d", i), "page": fmt.Sprintf("/fixture-%d", i), "visitors": i, "visits": i + 1, "pageviews": i + 2, "events": i + 3}
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
	token := strings.TrimSpace(cfg.Secrets["api_token"])
	if token == "" {
		return nil, errors.New("plausible connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func query(cfg connectors.RuntimeConfig, endpoint streamEndpoint) url.Values {
	q := url.Values{}
	if endpoint.requiresSite {
		q.Set("site_id", strings.TrimSpace(cfg.Config["site_id"]))
		q.Set("period", valueOrDefault(cfg.Config["period"], "30d"))
	}
	if endpoint.path == "stats/breakdown" {
		q.Set("property", valueOrDefault(cfg.Config["property"], "event:page"))
	}
	for _, key := range []string{"date", "metrics", "filters", "compare"} {
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			q.Set(key, v)
		}
	}
	return q
}

func valueOrDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return strings.TrimSpace(v)
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("plausible config base_url is invalid: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", errors.New("plausible config base_url must be an absolute http or https URL")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "page_size", defaultPageSize, 1, 1000)
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "max_pages", defaultMaxPages, 0, 10000)
}

func intConfig(cfg connectors.RuntimeConfig, key string, def, min, max int) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config[key]))
	if raw == "" {
		return def, nil
	}
	if key == "max_pages" && (raw == "all" || raw == "unlimited") {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min {
		return 0, fmt.Errorf("plausible config %s must be an integer >= %d", key, min)
	}
	if max > 0 && value > max {
		return max, nil
	}
	return value, nil
}

func first(m map[string]any, keys ...string) any {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
