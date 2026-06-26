package segment

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName    = "segment"
	defaultBaseURL   = "https://api.segmentapis.com"
	defaultPageSize  = 100
	defaultMaxPages  = 1
	fixtureUpdatedAt = "2026-01-01T00:00:00Z"
	userAgent        = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Segment",
		IntegrationType: "api",
		Description:     "Reads Segment workspace, source, and destination metadata through the Segment Public API.",
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
	return r.DoJSON(ctx, http.MethodGet, "workspaces", nil, nil, nil)
}

func (Connector) Catalog(ctx context.Context, _ connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	stream := req.Stream
	if stream == "" {
		stream = "workspaces"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("segment stream %q not found", req.Stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, req.State, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := positiveInt(req.Config.Config["page_size"], defaultPageSize, 1, 1000, "page_size")
	if err != nil {
		return err
	}
	maxPages, err := parseMaxPages(req.Config.Config["max_pages"])
	if err != nil {
		return err
	}
	pager := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "page_size", StartPage: 1, PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, ep.resource, nil, pager, ep.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(connectors.Record(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	token := strings.TrimSpace(secret(cfg, "api_token"))
	if token == "" {
		return nil, errors.New("segment connector requires secret api_token")
	}
	base := baseURL(cfg, defaultBaseURL)
	if region := strings.TrimSpace(cfg.Config["region"]); region != "" && strings.TrimSpace(cfg.Config["base_url"]) == "" {
		base = "https://" + region + ".segmentapis.com"
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

type streamEndpoint struct{ resource, recordsPath string }

var streamEndpoints = map[string]streamEndpoint{
	"workspaces":   {resource: "workspaces", recordsPath: "workspaces"},
	"sources":      {resource: "sources", recordsPath: "sources"},
	"destinations": {resource: "destinations", recordsPath: "destinations"},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "slug", Type: "string"}, {Name: "updated_at", Type: "string"}}
	return []connectors.Stream{
		{Name: "workspaces", Description: "Segment workspaces.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "sources", Description: "Segment sources.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "destinations", Description: "Segment destinations.", PrimaryKey: []string{"id"}, Fields: fields},
	}
}

func readFixture(ctx context.Context, stream string, state map[string]string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "slug": fmt.Sprintf("fixture-%d", i), "updated_at": fixtureUpdatedAt}
		if cursor := connsdk.Cursor(state); cursor != "" {
			rec["previous_cursor"] = cursor
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

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

func baseURL(cfg connectors.RuntimeConfig, fallback string) string {
	if v := strings.TrimSpace(cfg.Config["base_url"]); v != "" {
		return strings.TrimRight(v, "/")
	}
	return fallback
}

func positiveInt(raw string, def, min, max int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < min || n > max {
		return 0, fmt.Errorf("%s must be between %d and %d", name, min, max)
	}
	return n, nil
}

func parseMaxPages(raw string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultMaxPages, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errors.New("max_pages must be a non-negative integer")
	}
	return n, nil
}
