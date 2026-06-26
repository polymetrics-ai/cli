// Package surveysparrow implements the native pm SurveySparrow connector.
package surveysparrow

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
	connectorName   = "survey-sparrow"
	defaultBaseURL  = "https://api.surveysparrow.com/v3"
	defaultPageSize = 100
	maxPageSize     = 500
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamEndpoint struct {
	resource    string
	recordsPath string
	needsSurvey bool
	fields      []connectors.Field
	mapRecord   func(map[string]any) connectors.Record
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "SurveySparrow", IntegrationType: "api", Description: "Reads SurveySparrow surveys, contacts, responses, and questions through the SurveySparrow API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "surveys", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check survey-sparrow: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "surveys"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("survey-sparrow stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, emit)
	}
	resource, err := resolveResource(endpoint, req.Config)
	if err != nil {
		return err
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
	paginator := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, resource, nil, paginator, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "access_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("survey-sparrow connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("survey-sparrow config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("survey-sparrow config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("survey-sparrow config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.needsSurvey {
		return endpoint.resource, nil
	}
	surveyID := strings.TrimSpace(cfg.Config["survey_id"])
	if surveyID == "" {
		return "", fmt.Errorf("survey-sparrow stream requires config survey_id for path %q", endpoint.resource)
	}
	return strings.ReplaceAll(endpoint.resource, "{survey_id}", url.PathEscape(surveyID)), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "survey-sparrow config page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("survey-sparrow config max_pages must be a non-negative integer: %w", err)
	}
	return value, nil
}

func boundedInt(raw string, def, max int, name string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s must be between 1 and %d", name, max)
	}
	return value, nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": i, "name": fmt.Sprintf("Fixture %s %d", stream, i), "survey_type": "ClassicForm", "email": fmt.Sprintf("fixture+%d@example.com", i), "completed_time": "2026-01-01T00:00:00Z", "question": "How satisfied are you?"}
		rec := endpoint.mapRecord(item)
		rec["connector"] = connectorName
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "surveys", Description: "SurveySparrow surveys.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["surveys"].fields},
		{Name: "contacts", Description: "SurveySparrow contacts.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["contacts"].fields},
		{Name: "responses", Description: "SurveySparrow responses.", PrimaryKey: []string{"id"}, CursorFields: []string{"completed_time"}, Fields: streamEndpoints["responses"].fields},
		{Name: "questions", Description: "SurveySparrow questions for a survey_id.", PrimaryKey: []string{"id"}, Fields: streamEndpoints["questions"].fields},
	}
}

var streamEndpoints = map[string]streamEndpoint{
	"surveys":   {resource: "surveys", recordsPath: "data", fields: surveyFields(), mapRecord: copyRecord("id", "name", "survey_type")},
	"contacts":  {resource: "contacts", recordsPath: "data", fields: contactFields(), mapRecord: copyRecord("id", "email", "name")},
	"responses": {resource: "responses", recordsPath: "data", fields: responseFields(), mapRecord: copyRecord("id", "completed_time", "survey_id")},
	"questions": {resource: "surveys/{survey_id}/questions", recordsPath: "data", needsSurvey: true, fields: questionFields(), mapRecord: copyRecord("id", "question", "survey_id")},
}

func surveyFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "survey_type", Type: "string"}}
}

func contactFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "integer"}, {Name: "email", Type: "string"}, {Name: "name", Type: "string"}}
}

func responseFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "integer"}, {Name: "completed_time", Type: "timestamp"}, {Name: "survey_id", Type: "integer"}}
}

func questionFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "integer"}, {Name: "question", Type: "string"}, {Name: "survey_id", Type: "integer"}}
}

func copyRecord(keys ...string) func(map[string]any) connectors.Record {
	return func(item map[string]any) connectors.Record {
		rec := connectors.Record{}
		for _, key := range keys {
			rec[key] = item[key]
		}
		return rec
	}
}
