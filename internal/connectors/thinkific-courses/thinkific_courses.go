// Package thinkificcourses implements a read-only native Go connector for the
// Thinkific public courses API.
package thinkificcourses

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
	connectorName   = "thinkific-courses"
	defaultBaseURL  = "https://api.thinkific.com/api/public/v1"
	defaultPageSize = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Thinkific Courses",
		IntegrationType: "api",
		Description:     "Reads courses, chapters, lessons, and enrollments from the Thinkific public API.",
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
	q := url.Values{"page": {"1"}, "limit": {"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "courses", q, nil, nil); err != nil {
		return fmt.Errorf("check %s: %w", connectorName, err)
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
		stream = "courses"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("%s stream %q not found", connectorName, stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, spec, emit)
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
	p := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize: size}
	return connsdk.Harvest(ctx, r, http.MethodGet, spec.path, nil, p, spec.recordsPath, max, func(rec connsdk.Record) error {
		return emit(connectors.Record(rec))
	})
}

func (Connector) Write(ctx context.Context, _ connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WriteResult{}, err
	}
	return connectors.WriteResult{RecordsFailed: len(records)}, fmt.Errorf("%s connector is read-only: %w", connectorName, connectors.ErrUnsupportedOperation)
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := secret(cfg, "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("thinkific-courses connector requires secret api_key")
	}
	subdomain := firstConfig(cfg, "X-Auth-Subdomain", "x_auth_subdomain", "subdomain")
	if subdomain == "" {
		return nil, errors.New("thinkific-courses connector requires config X-Auth-Subdomain")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		UserAgent: userAgent,
		DefaultHeaders: map[string]string{
			"X-Auth-API-Key":   key,
			"X-Auth-Subdomain": subdomain,
		},
	}, nil
}

type streamSpec struct {
	name        string
	description string
	path        string
	recordsPath string
	fields      []connectors.Field
}

var streamSpecs = map[string]streamSpec{
	"courses":     {name: "courses", description: "Thinkific courses.", path: "courses", recordsPath: "items", fields: commonFields("id", "name", "slug", "description", "created_at", "updated_at")},
	"chapters":    {name: "chapters", description: "Thinkific course chapters.", path: "chapters", recordsPath: "items", fields: commonFields("id", "course_id", "name", "position", "created_at", "updated_at")},
	"lessons":     {name: "lessons", description: "Thinkific course lessons.", path: "lessons", recordsPath: "items", fields: commonFields("id", "chapter_id", "course_id", "name", "position", "created_at", "updated_at")},
	"enrollments": {name: "enrollments", description: "Thinkific course enrollments.", path: "enrollments", recordsPath: "items", fields: commonFields("id", "course_id", "user_id", "percentage_completed", "activated_at", "completed_at", "updated_at")},
}

func streams() []connectors.Stream {
	order := []string{"courses", "chapters", "lessons", "enrollments"}
	out := make([]connectors.Stream, 0, len(order))
	for _, name := range order {
		s := streamSpecs[name]
		out = append(out, connectors.Stream{Name: s.name, Description: s.description, Fields: s.fields, PrimaryKey: []string{"id"}})
	}
	return out
}

func commonFields(names ...string) []connectors.Field {
	fields := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		fields = append(fields, connectors.Field{Name: name, Type: "string"})
	}
	return fields
}

func readFixture(ctx context.Context, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{
			"id":         fmt.Sprintf("%s_fixture_%d", spec.name, i),
			"name":       fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(spec.name, "s"), i),
			"created_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updated_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"fixture":    true,
		}
		for _, field := range spec.fields {
			if _, ok := rec[field.Name]; !ok {
				rec[field.Name] = fmt.Sprintf("%s_%d", field.Name, i)
			}
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	raw := strings.TrimSpace(cfg.Config["base_url"])
	if raw == "" {
		raw = defaultBaseURL
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid %s base_url", connectorName)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("invalid %s base_url scheme %q", connectorName, u.Scheme)
	}
	return strings.TrimRight(raw, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	value := strings.TrimSpace(cfg.Config["page_size"])
	if value == "" {
		return defaultPageSize, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s page_size must be a positive integer", connectorName)
	}
	return n, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	value := strings.TrimSpace(cfg.Config["max_pages"])
	if value == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%s max_pages must be a non-negative integer", connectorName)
	}
	return n, nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets != nil {
		if value := strings.TrimSpace(cfg.Secrets[key]); value != "" {
			return value
		}
	}
	return strings.TrimSpace(cfg.Config[key])
}

func firstConfig(cfg connectors.RuntimeConfig, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(cfg.Config[key]); value != "" {
			return value
		}
	}
	return ""
}
