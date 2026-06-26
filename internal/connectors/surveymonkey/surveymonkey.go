// Package surveymonkey implements a read-only SurveyMonkey API connector.
package surveymonkey

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
	surveyMonkeyName            = "surveymonkey"
	surveyMonkeyDefaultBaseURL  = "https://api.surveymonkey.com/v3"
	surveyMonkeyDefaultPageSize = 100
	surveyMonkeyMaxPageSize     = 1000
	surveyMonkeyUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("surveymonkey", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return surveyMonkeyName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: surveyMonkeyName, DisplayName: "SurveyMonkey", IntegrationType: "api", Description: "Reads SurveyMonkey surveys, collectors, and bulk survey responses through the v3 REST API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if strings.TrimSpace(surveyMonkeyAccessToken(cfg)) == "" {
		return errors.New("surveymonkey connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "users/me", nil, nil, nil); err != nil {
		return fmt.Errorf("check surveymonkey: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: surveyMonkeyStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "surveys"
	}
	endpoint, ok := surveyMonkeyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("surveymonkey stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := surveyMonkeyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := surveyMonkeyMaxPages(req.Config)
	if err != nil {
		return err
	}
	resource, err := surveyMonkeyResource(req.Config, endpoint)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, resource, endpoint, pageSize, maxPages, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, endpoint surveyMonkeyStreamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := resource
	query := url.Values{"per_page": []string{strconv.Itoa(pageSize)}}
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read surveymonkey %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode surveymonkey %s: %w", resource, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "links.next")
		if err != nil {
			return fmt.Errorf("decode surveymonkey next link: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		path = next
		query = nil
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint surveyMonkeyStreamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("fixture-%d", i), "title": fmt.Sprintf("Fixture Survey %d", i), "name": fmt.Sprintf("Fixture Collector %d", i), "date_created": "2026-01-01T00:00:00Z", "date_modified": "2026-01-02T00:00:00Z", "analyze_url": "https://example.invalid"}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := surveyMonkeyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := surveyMonkeyAccessToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("surveymonkey connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: surveyMonkeyUserAgent}, nil
}

type surveyMonkeyStreamEndpoint struct {
	resource     string
	surveyScoped bool
	mapRecord    func(map[string]any) connectors.Record
}

var surveyMonkeyStreamEndpoints = map[string]surveyMonkeyStreamEndpoint{
	"surveys":          {resource: "surveys", mapRecord: surveyMonkeySurveyRecord},
	"collectors":       {resource: "collectors", surveyScoped: true, mapRecord: surveyMonkeyCollectorRecord},
	"survey_responses": {resource: "responses/bulk", surveyScoped: true, mapRecord: surveyMonkeyResponseRecord},
}

func surveyMonkeyStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "surveys", Description: "SurveyMonkey surveys.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "date_created", Type: "string"}, {Name: "date_modified", Type: "string"}}},
		{Name: "collectors", Description: "Collectors for the configured survey_id.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "date_created", Type: "string"}, {Name: "date_modified", Type: "string"}}},
		{Name: "survey_responses", Description: "Bulk responses for the configured survey_id.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "date_created", Type: "string"}, {Name: "date_modified", Type: "string"}, {Name: "analyze_url", Type: "string"}}},
	}
}

func surveyMonkeySurveyRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "title": item["title"], "date_created": item["date_created"], "date_modified": item["date_modified"]}
}

func surveyMonkeyCollectorRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "date_created": item["date_created"], "date_modified": item["date_modified"]}
}

func surveyMonkeyResponseRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "date_created": item["date_created"], "date_modified": item["date_modified"], "analyze_url": item["analyze_url"]}
}

func surveyMonkeyResource(cfg connectors.RuntimeConfig, endpoint surveyMonkeyStreamEndpoint) (string, error) {
	if !endpoint.surveyScoped {
		return endpoint.resource, nil
	}
	surveyID := strings.TrimSpace(cfg.Config["survey_id"])
	if surveyID == "" {
		return "", errors.New("surveymonkey connector requires config survey_id for this stream")
	}
	return "surveys/" + url.PathEscape(surveyID) + "/" + endpoint.resource, nil
}

func surveyMonkeyAccessToken(cfg connectors.RuntimeConfig) string { return cfg.Secrets["access_token"] }

func surveyMonkeyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return validBaseURL(surveyMonkeyName, cfg.Config["base_url"], surveyMonkeyDefaultBaseURL)
}

func surveyMonkeyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(surveyMonkeyName, cfg.Config["page_size"], surveyMonkeyDefaultPageSize, 1, surveyMonkeyMaxPageSize)
}

func surveyMonkeyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	return maxPagesConfig(surveyMonkeyName, cfg.Config["max_pages"])
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
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
