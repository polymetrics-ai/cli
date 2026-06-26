// Package adjust implements a conservative read-only Adjust Report Service API
// connector. It reads the documented reports-service/report endpoint and flattens
// dimension and metric objects into records; it does not expose Adjust write APIs.
package adjust

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
	connectorName  = "adjust"
	defaultBaseURL = "https://automate.adjust.com"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Adjust", IntegrationType: "api", Description: "Reads Adjust report-service report rows. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	q := reportQuery(cfg)
	q.Set("limit", "1")
	_, err = r.Do(ctx, http.MethodGet, "/reports-service/report", q, nil)
	if err != nil {
		return fmt.Errorf("check adjust: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{{Name: "reports", Description: "Adjust Report Service rows for configured dimensions and metrics.", Fields: []connectors.Field{{Name: "date", Type: "string"}, {Name: "country", Type: "string"}, {Name: "app", Type: "string"}, {Name: "installs", Type: "number"}, {Name: "clicks", Type: "number"}, {Name: "cost", Type: "number"}}}}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "reports"
	}
	if stream != "reports" {
		return fmt.Errorf("adjust stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	q := reportQuery(req.Config)
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		q.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, "/reports-service/report", q, nil)
		if err != nil {
			return fmt.Errorf("read adjust reports: %w", err)
		}
		items, err := connsdk.RecordsAt(resp.Body, "rows")
		if err != nil {
			return fmt.Errorf("decode adjust reports: %w", err)
		}
		for _, item := range items {
			if err := emit(reportRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_page")
		if err != nil {
			return err
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		if n, err := strconv.Atoi(next); err == nil {
			page = n - 1
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
	token := strings.TrimSpace(cfg.Secrets["api_token"])
	if token == "" {
		return nil, errors.New("adjust connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func reportQuery(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	q.Set("dimensions", csvOrDefault(cfg.Config["dimensions"], "country"))
	q.Set("metrics", csvOrDefault(cfg.Config["metrics"], "installs"))
	start := first(cfg.Config["ingest_start"], cfg.Config["start_date"])
	end := first(cfg.Config["end_date"], cfg.Config["until"], start)
	if start != "" && end != "" {
		q.Set("date_period", start+":"+end)
	}
	if add := strings.TrimSpace(cfg.Config["additional_metrics"]); add != "" {
		q.Set("additional_metrics", add)
	}
	return q
}

func reportRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{}
	for k, v := range item {
		switch k {
		case "dimensions", "metrics":
			if m, ok := v.(map[string]any); ok {
				for mk, mv := range m {
					rec[mk] = normalize(mv)
				}
			}
		default:
			rec[k] = normalize(v)
		}
	}
	return rec
}

func readFixture(ctx context.Context, emit func(connectors.Record) error) error {
	for _, rec := range []connectors.Record{{"country": "US", "installs": 10.0, "fixture": true}, {"country": "DE", "installs": 3.0, "fixture": true}} {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("adjust config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func csvOrDefault(raw, def string) string {
	if strings.TrimSpace(raw) == "" {
		return def
	}
	return strings.Join(strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ' ' || r == '\n' || r == '\t' }), ",")
}

func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func normalize(v any) any {
	if n, ok := v.(json.Number); ok {
		if f, err := n.Float64(); err == nil {
			return f
		}
	}
	return v
}
