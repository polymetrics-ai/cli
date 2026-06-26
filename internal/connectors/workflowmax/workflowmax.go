// Package workflowmax implements a conservative read-only native WorkflowMax connector.
package workflowmax

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
	defaultBaseURL  = "https://api.workflowmax2.com"
	defaultPageSize = 100
	maxPageSize     = 500
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("workflowmax", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "workflowmax" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "workflowmax", DisplayName: "WorkflowMax", IntegrationType: "api", Description: "Reads WorkflowMax jobs, clients, and contacts through REST API endpoints. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

type streamEndpoint struct {
	resource     string
	recordsPath  string
	description  string
	fields       []connectors.Field
	cursorFields []string
}

var streamOrder = []string{"jobs", "clients", "contacts"}

var streamEndpoints = map[string]streamEndpoint{
	"jobs":     {resource: "jobs", recordsPath: "data", description: "WorkflowMax jobs.", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "updated_at", Type: "string"}}, cursorFields: []string{"updated_at"}},
	"clients":  {resource: "clients", recordsPath: "data", description: "WorkflowMax clients.", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "updated_at", Type: "string"}}, cursorFields: []string{"updated_at"}},
	"contacts": {resource: "contacts", recordsPath: "data", description: "WorkflowMax contacts.", fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "updated_at", Type: "string"}}, cursorFields: []string{"updated_at"}},
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
	if err := r.DoJSON(ctx, http.MethodGet, streamEndpoints["jobs"].resource, url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check workflowmax: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	streams := make([]connectors.Stream, 0, len(streamOrder))
	for _, name := range streamOrder {
		endpoint := streamEndpoints[name]
		streams = append(streams, connectors.Stream{Name: name, Description: endpoint.description, Fields: endpoint.fields, PrimaryKey: []string{"id"}, CursorFields: endpoint.cursorFields})
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "jobs"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("workflowmax stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, c.Name(), stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := boundedInt(req.Config.Config["page_size"], defaultPageSize, maxPageSize, "workflowmax config page_size")
	if err != nil {
		return err
	}
	maxPages, err := readMaxPages(req.Config, "workflowmax")
	if err != nil {
		return err
	}
	p := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "page_size", StartPage: 1, PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, p, endpoint.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(connectors.Record(rec))
	})
}

func readFixture(ctx context.Context, connector, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "updated_at": "2026-01-01T00:00:00Z", "connector": connector, "stream": stream, "fixture": true}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, defaultBaseURL, "workflowmax")
	if err != nil {
		return nil, err
	}
	token := secret(cfg, "access_token")
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("workflowmax connector requires secret access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

func baseURL(cfg connectors.RuntimeConfig, fallback, connector string) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", connector, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https", connector)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", connector)
	}
	return strings.TrimRight(base, "/"), nil
}

func boundedInt(raw string, fallback, max int, label string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", label, err)
	}
	if value < 1 || value > max {
		return 0, fmt.Errorf("%s must be between 1 and %d", label, max)
	}
	return value, nil
}

func readMaxPages(cfg connectors.RuntimeConfig, connector string) (int, error) {
	raw := strings.TrimSpace(cfg.Config["max_pages"])
	if raw == "" {
		return 1, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s config max_pages must be an integer: %w", connector, err)
	}
	if value < 1 {
		return 0, fmt.Errorf("%s config max_pages must be at least 1", connector)
	}
	return value, nil
}
