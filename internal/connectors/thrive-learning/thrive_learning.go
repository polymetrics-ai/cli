// Package thrivelearning implements a read-only native Go connector for Thrive Learning.
package thrivelearning

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
	connectorName   = "thrive-learning"
	defaultBaseURL  = "https://api.thrivelearning.com"
	defaultPageSize = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Thrive Learning", IntegrationType: "api", Description: "Reads users, content, and learning completions from the Thrive Learning API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if start := strings.TrimSpace(cfg.Config["start_date"]); start != "" {
		q.Set("updated_since", start)
	}
	if err := r.DoJSON(ctx, http.MethodGet, "users", q, nil, nil); err != nil {
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
		stream = "users"
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
	q := url.Values{}
	if start := strings.TrimSpace(req.Config.Config["start_date"]); start != "" {
		q.Set("updated_since", start)
	}
	p := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize: size}
	return connsdk.Harvest(ctx, r, http.MethodGet, spec.path, q, p, spec.recordsPath, max, func(rec connsdk.Record) error {
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
	username := strings.TrimSpace(cfg.Config["username"])
	if username == "" {
		return nil, errors.New("thrive-learning connector requires config username")
	}
	password := secret(cfg, "password")
	if password == "" {
		return nil, errors.New("thrive-learning connector requires secret password")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(username, password), UserAgent: userAgent}, nil
}

type streamSpec struct {
	name, description, path, recordsPath string
	fields                               []connectors.Field
}

var streamSpecs = map[string]streamSpec{
	"users":       {"users", "Thrive Learning users.", "users", "items", fields("id", "email", "name", "created_at", "updated_at")},
	"content":     {"content", "Thrive Learning learning content.", "content", "items", fields("id", "title", "type", "created_at", "updated_at")},
	"completions": {"completions", "Thrive Learning content completions.", "completions", "items", fields("id", "user_id", "content_id", "completed_at", "updated_at")},
}

func streams() []connectors.Stream {
	order := []string{"users", "content", "completions"}
	out := make([]connectors.Stream, 0, len(order))
	for _, name := range order {
		s := streamSpecs[name]
		out = append(out, connectors.Stream{Name: s.name, Description: s.description, Fields: s.fields, PrimaryKey: []string{"id"}})
	}
	return out
}

func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
}

func readFixture(ctx context.Context, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", spec.name, i), "name": fmt.Sprintf("Fixture %s %d", spec.name, i), "title": fmt.Sprintf("Fixture %s %d", spec.name, i), "email": fmt.Sprintf("fixture+%d@example.com", i), "fixture": true}
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
