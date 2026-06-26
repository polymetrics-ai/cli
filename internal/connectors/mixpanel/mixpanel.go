// Package mixpanel implements a conservative read-only Mixpanel Query API connector.
// It covers stable read endpoints and optional page tokens, not JQL/export jobs.
package mixpanel

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
	mixpanelName            = "mixpanel"
	mixpanelDefaultBaseURL  = "https://mixpanel.com/api/2.0"
	mixpanelDefaultPageSize = 1000
	mixpanelMaxPageSize     = 10000
	mixpanelUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("mixpanel", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return mixpanelName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: mixpanelName, DisplayName: "Mixpanel", IntegrationType: "api", Description: "Reads Mixpanel cohorts, annotations, and engage profiles through read-only Query API endpoints.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if err := mixpanelValidateAuth(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "cohorts/list", nil, nil, nil); err != nil {
		return fmt.Errorf("check mixpanel: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mixpanelStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "cohorts"
	}
	endpoint, ok := mixpanelStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mixpanel stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mixpanelPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mixpanelMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint mixpanelStreamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	pageValue := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if pageValue != "" {
			query.Set("page", pageValue)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read mixpanel %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode mixpanel %s: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next")
		if err != nil {
			return fmt.Errorf("decode mixpanel next page: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageValue = next
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint mixpanelStreamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": i, "name": fmt.Sprintf("Fixture %s %d", endpoint.resource, i), "count": i * 10, "$distinct_id": fmt.Sprintf("fixture-%d", i), "$email": fmt.Sprintf("fixture%d@example.com", i), "created": "2026-01-01T00:00:00Z"}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	if err := mixpanelValidateAuth(cfg); err != nil {
		return nil, err
	}
	base, err := mixpanelBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	username, password := mixpanelCredentials(cfg)
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(username, password), UserAgent: mixpanelUserAgent}, nil
}

type mixpanelStreamEndpoint struct {
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var mixpanelStreamEndpoints = map[string]mixpanelStreamEndpoint{
	"cohorts":     {resource: "cohorts/list", recordsPath: "cohorts", mapRecord: mixpanelCohortRecord},
	"annotations": {resource: "annotations", recordsPath: "annotations", mapRecord: mixpanelAnnotationRecord},
	"engage":      {resource: "engage", recordsPath: "results", mapRecord: mixpanelProfileRecord},
}

func mixpanelStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "cohorts", Description: "Mixpanel cohorts.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "count", Type: "integer"}}},
		{Name: "annotations", Description: "Mixpanel annotations.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "date", Type: "string"}, {Name: "description", Type: "string"}}},
		{Name: "engage", Description: "Mixpanel engage profile records.", PrimaryKey: []string{"distinct_id"}, Fields: []connectors.Field{{Name: "distinct_id", Type: "string"}, {Name: "email", Type: "string"}, {Name: "created", Type: "string"}}},
	}
}

func mixpanelCohortRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "count": item["count"]}
}

func mixpanelAnnotationRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "date": item["date"], "description": item["description"]}
}

func mixpanelProfileRecord(item map[string]any) connectors.Record {
	props, _ := item["$properties"].(map[string]any)
	return connectors.Record{"distinct_id": first(item["$distinct_id"], item["distinct_id"]), "email": first(props["$email"], item["$email"], item["email"]), "created": first(props["$created"], item["created"])}
}

func mixpanelValidateAuth(cfg connectors.RuntimeConfig) error {
	username, password := mixpanelCredentials(cfg)
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return errors.New("mixpanel connector requires config username and secret password")
	}
	return nil
}

func mixpanelCredentials(cfg connectors.RuntimeConfig) (string, string) {
	username := cfg.Config["username"]
	if username == "" {
		username = cfg.Secrets["username"]
	}
	password := cfg.Secrets["password"]
	if password == "" {
		password = cfg.Secrets["api_secret"]
	}
	return username, password
}

func mixpanelBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validBaseURL(mixpanelName, cfg.Config["base_url"], mixpanelDefaultBaseURL)
}

func mixpanelPageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(mixpanelName, cfg.Config["page_size"], mixpanelDefaultPageSize, 1, mixpanelMaxPageSize)
}

func mixpanelMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	return maxPagesConfig(mixpanelName, cfg.Config["max_pages"])
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func first(values ...any) any {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		if value != nil {
			return value
		}
	}
	return nil
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
