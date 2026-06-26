// Package outbrainamplify implements a read-only connector for selected
// Outbrain Amplify API resources. The live API has both token and
// username/password auth modes; this port supports bearer tokens directly and
// falls back to HTTP Basic for local/proxy deployments that expose that mode.
package outbrainamplify

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
	defaultBaseURL  = "https://api.outbrain.com/amplify/v0.1"
	defaultPageSize = 100
	maxPageSize     = 500
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("outbrain-amplify", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "outbrain-amplify" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "outbrain-amplify", DisplayName: "Outbrain Amplify", IntegrationType: "api", Description: "Reads Outbrain Amplify marketers, campaigns, and performance reports.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	return r.DoJSON(ctx, http.MethodGet, "/marketers", url.Values{"limit": []string{"1"}}, nil, nil)
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "outbrain-amplify", Streams: streams()}, nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "marketers"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("outbrain-amplify stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, ep, req, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	max, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return harvest(ctx, r, ep, req.Config, size, max, emit)
}

func harvest(ctx context.Context, r *connsdk.Requester, ep streamEndpoint, cfg connectors.RuntimeConfig, size, max int, emit func(connectors.Record) error) error {
	offset := 0
	for page := 0; max == 0 || page < max; page++ {
		query := url.Values{"limit": []string{strconv.Itoa(size)}, "offset": []string{strconv.Itoa(offset)}}
		for _, key := range []string{"start_date", "end_date", "report_granularity", "conversion_count", "geo_location_breakdown"} {
			if value := strings.TrimSpace(cfg.Config[key]); value != "" {
				query.Set(key, value)
			}
		}
		resp, err := r.Do(ctx, http.MethodGet, ep.path(cfg), query, nil)
		if err != nil {
			return fmt.Errorf("read outbrain-amplify %s: %w", ep.path(cfg), err)
		}
		records, err := connsdk.RecordsAt(resp.Body, ep.recordsPath)
		if err != nil {
			return err
		}
		for _, item := range records {
			if err := emit(ep.mapRecord(map[string]any(item))); err != nil {
				return err
			}
		}
		offset += len(records)
		if len(records) == 0 || offset >= totalResults(resp.Body) {
			return nil
		}
	}
	return nil
}

func totalResults(body []byte) int {
	var env struct {
		TotalResults int `json:"totalResults"`
		Total        int `json:"total"`
	}
	_ = json.Unmarshal(body, &env)
	if env.TotalResults > 0 {
		return env.TotalResults
	}
	return env.Total
}

func readFixture(ctx context.Context, stream string, ep streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %d", i), "enabled": i%2 == 1, "status": "active", "created_at": "2026-01-01", "impressions": 100 * i, "clicks": 10 * i, "spend": 12.34}
		rec := ep.mapRecord(item)
		if cursor := req.State[connsdk.CursorStateKey]; cursor != "" {
			rec["previous_cursor"] = cursor
		}
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
	auth, err := authenticator(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth, UserAgent: userAgent}, nil
}

func authenticator(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	if token := secret(cfg, "access_token"); token != "" {
		return connsdk.Bearer(token), nil
	}
	username := strings.TrimSpace(cfg.Config["username"])
	password := secret(cfg, "password")
	if username != "" && password != "" {
		return connsdk.Basic(username, password), nil
	}
	return nil, errors.New("outbrain-amplify connector requires secret access_token or username/password")
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	return validateBaseURL("outbrain-amplify", base)
}

func validateBaseURL(name, raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", name, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(raw, "/"), nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Secrets[key])
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > maxPageSize {
		return 0, fmt.Errorf("outbrain-amplify config page_size must be between 1 and %d", maxPageSize)
	}
	return n, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.ToLower(strings.TrimSpace(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errors.New("outbrain-amplify config max_pages must be a non-negative integer")
	}
	return n, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamEndpoint struct {
	recordsPath string
	path        func(connectors.RuntimeConfig) string
	mapRecord   func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"marketers":           {recordsPath: "marketers", path: staticPath("/marketers"), mapRecord: record},
	"campaigns":           {recordsPath: "campaigns", path: campaignsPath, mapRecord: record},
	"performance_reports": {recordsPath: "results", path: reportsPath, mapRecord: record},
}

func staticPath(path string) func(connectors.RuntimeConfig) string {
	return func(connectors.RuntimeConfig) string { return path }
}

func campaignsPath(cfg connectors.RuntimeConfig) string {
	if marketer := strings.TrimSpace(cfg.Config["marketer_id"]); marketer != "" {
		return "/marketers/" + url.PathEscape(marketer) + "/campaigns"
	}
	return "/campaigns"
}

func reportsPath(cfg connectors.RuntimeConfig) string {
	if marketer := strings.TrimSpace(cfg.Config["marketer_id"]); marketer != "" {
		return "/marketers/" + url.PathEscape(marketer) + "/reports"
	}
	return "/reports"
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "enabled", Type: "boolean"}, {Name: "status", Type: "string"}, {Name: "created_at", Type: "string"}, {Name: "impressions", Type: "integer"}, {Name: "clicks", Type: "integer"}, {Name: "spend", Type: "number"}}
	return []connectors.Stream{{Name: "marketers", Description: "Outbrain marketers.", PrimaryKey: []string{"id"}, Fields: fields}, {Name: "campaigns", Description: "Outbrain campaigns.", PrimaryKey: []string{"id"}, Fields: fields}, {Name: "performance_reports", Description: "Outbrain performance report rows.", PrimaryKey: []string{"id"}, Fields: fields}}
}

func record(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "enabled": item["enabled"], "status": item["status"], "created_at": item["created_at"], "impressions": item["impressions"], "clicks": item["clicks"], "spend": item["spend"]}
}
