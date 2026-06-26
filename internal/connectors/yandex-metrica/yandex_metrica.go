// Package yandexmetrica implements a read-only Yandex Metrica Statistics API
// connector using /stat/v1/data with offset/limit pagination.
package yandexmetrica

import (
	"context"
	"encoding/json"
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
	connectorName  = "yandex-metrica"
	defaultBaseURL = "https://api-metrika.yandex.net"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Yandex Metrica", IntegrationType: "api", Description: "Reads Yandex Metrica statistics rows. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	q, err := query(cfg, 1, 1)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "/stat/v1/data", q, nil, nil); err != nil {
		return fmt.Errorf("check yandex-metrica: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{{Name: "traffic", Description: "Yandex Metrica statistics data for configured dimensions and metrics.", Fields: []connectors.Field{{Name: "dimension_1_name", Type: "string"}, {Name: "dimension_1_id", Type: "string"}, {Name: "metric_1", Type: "number"}, {Name: "dimensions", Type: "object"}, {Name: "metrics", Type: "object"}}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "traffic"
	}
	if stream != "traffic" {
		return fmt.Errorf("yandex-metrica stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	limit := intConfig(req.Config.Config["limit"], 100)
	emitted := 0
	for offset := 1; ; offset += limit {
		if err := ctx.Err(); err != nil {
			return err
		}
		q, err := query(req.Config, limit, offset)
		if err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, "/stat/v1/data", q, nil)
		if err != nil {
			return fmt.Errorf("read yandex-metrica traffic: %w", err)
		}
		items, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return err
		}
		for _, item := range items {
			emitted++
			if err := emit(rowRecord(item)); err != nil {
				return err
			}
		}
		total, _ := connsdk.StringAt(resp.Body, "total_rows")
		if total == "" || emitted >= atoi(total) || len(items) == 0 {
			return nil
		}
	}
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Config["counter_id"]) == "" {
		return nil, errors.New("yandex-metrica connector requires config counter_id")
	}
	token := strings.TrimSpace(cfg.Secrets["auth_token"])
	if token == "" {
		return nil, errors.New("yandex-metrica connector requires secret auth_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func query(cfg connectors.RuntimeConfig, limit, offset int) (url.Values, error) {
	start := strings.TrimSpace(cfg.Config["start_date"])
	if start == "" {
		return nil, errors.New("yandex-metrica connector requires config start_date")
	}
	q := url.Values{}
	q.Set("ids", strings.TrimSpace(cfg.Config["counter_id"]))
	q.Set("date1", start)
	if end := strings.TrimSpace(cfg.Config["end_date"]); end != "" {
		q.Set("date2", end)
	}
	q.Set("dimensions", first(cfg.Config["dimensions"], "ym:s:lastTrafficSource"))
	q.Set("metrics", first(cfg.Config["metrics"], "ym:s:visits"))
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	return q, nil
}

func rowRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{"dimensions": item["dimensions"], "metrics": item["metrics"]}
	if dims, ok := item["dimensions"].([]any); ok {
		for i, d := range dims {
			if m, ok := d.(map[string]any); ok {
				rec[fmt.Sprintf("dimension_%d_name", i+1)] = m["name"]
				rec[fmt.Sprintf("dimension_%d_id", i+1)] = m["id"]
			}
		}
	}
	if metrics, ok := item["metrics"].([]any); ok {
		for i, v := range metrics {
			rec[fmt.Sprintf("metric_%d", i+1)] = normalize(v)
		}
	}
	return rec
}

func readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for _, rec := range []connectors.Record{{"dimension_1_name": "Google", "dimension_1_id": "google", "metric_1": 10.0, "fixture": true}, {"dimension_1_name": "Direct", "dimension_1_id": "direct", "metric_1": 3.0, "fixture": true}} {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func normalize(v any) any {
	if n, ok := v.(json.Number); ok {
		if f, err := n.Float64(); err == nil {
			return f
		}
	}
	return v
}

func intConfig(raw string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(raw)); err == nil && n > 0 {
		return n
	}
	return def
}

func atoi(raw string) int { n, _ := strconv.Atoi(strings.TrimSpace(raw)); return n }

func first(raw, def string) string {
	if strings.TrimSpace(raw) == "" {
		return def
	}
	return strings.TrimSpace(raw)
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("yandex-metrica config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
