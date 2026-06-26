// Package appsflyer implements a read-only AppsFlyer Pull API connector for raw
// CSV export reports. The API does not expose writes through this connector.
package appsflyer

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName  = "appsflyer"
	defaultBaseURL = "https://hq1.appsflyer.com"
	userAgent      = "polymetrics-go-cli"
)

var streamPaths = map[string]string{"installs_report": "installs_report", "in_app_events_report": "in_app_events_report"}

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "AppsFlyer", IntegrationType: "api", Description: "Reads AppsFlyer raw-data CSV export reports. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	_, err = r.Do(ctx, http.MethodGet, reportPath(cfg, "installs_report"), reportQuery(cfg), nil)
	if err != nil {
		return fmt.Errorf("check appsflyer: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	fields := []connectors.Field{{Name: "appsflyer_id", Type: "string"}, {Name: "event_time", Type: "string"}, {Name: "media_source", Type: "string"}, {Name: "campaign", Type: "string"}}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "installs_report", Description: "AppsFlyer raw installs report.", Fields: fields},
		{Name: "in_app_events_report", Description: "AppsFlyer raw in-app events report.", Fields: fields},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "installs_report"
	}
	if _, ok := streamPaths[stream]; !ok {
		return fmt.Errorf("appsflyer stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, reportPath(req.Config, stream), reportQuery(req.Config), nil)
	if err != nil {
		return fmt.Errorf("read appsflyer %s: %w", stream, err)
	}
	return emitCSV(ctx, resp.Body, emit)
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Config["app_id"]) == "" {
		return nil, errors.New("appsflyer connector requires config app_id")
	}
	token := strings.TrimSpace(cfg.Secrets["api_token"])
	if token == "" {
		return nil, errors.New("appsflyer connector requires secret api_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent, Accept: "text/csv,application/json"}, nil
}

func reportPath(cfg connectors.RuntimeConfig, stream string) string {
	return "/api/raw-data/export/app/" + url.PathEscape(strings.TrimSpace(cfg.Config["app_id"])) + "/" + streamPaths[stream] + "/v5"
}

func reportQuery(cfg connectors.RuntimeConfig) url.Values {
	q := url.Values{}
	from := strings.TrimSpace(cfg.Config["start_date"])
	to := strings.TrimSpace(first(cfg.Config["end_date"], cfg.Config["start_date"]))
	if from != "" {
		q.Set("from", firstDate(from))
	}
	if to != "" {
		q.Set("to", firstDate(to))
	}
	if tz := strings.TrimSpace(cfg.Config["timezone"]); tz != "" {
		q.Set("timezone", tz)
	}
	return q
}

func emitCSV(ctx context.Context, body []byte, emit func(connectors.Record) error) error {
	r := csv.NewReader(bytes.NewReader(body))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("decode appsflyer csv: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}
	headers := make([]string, len(rows[0]))
	for i, h := range rows[0] {
		headers[i] = snake(h)
	}
	for _, row := range rows[1:] {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{}
		for i, v := range row {
			if i < len(headers) && headers[i] != "" {
				rec[headers[i]] = v
			}
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"appsflyer_id": fmt.Sprintf("af_fixture_%d", i), "event_time": fmt.Sprintf("2026-01-%02d 00:00:00", i), "media_source": "fixture", "campaign": stream, "fixture": true}); err != nil {
			return err
		}
	}
	return nil
}

func snake(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastUnderscore = false
		} else if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	out := strings.Trim(b.String(), "_")
	return strings.ReplaceAll(out, "apps_flyer", "appsflyer")
}

func firstDate(raw string) string {
	if i := strings.IndexByte(raw, ' '); i > 0 {
		return raw[:i]
	}
	return raw
}

func first(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("appsflyer config base_url must be an absolute http(s) URL")
	}
	return strings.TrimRight(raw, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
