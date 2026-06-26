package sendowl

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
	connectorName    = "sendowl"
	defaultBaseURL   = "https://www.sendowl.com/api/v1"
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
		DisplayName:     "SendOwl",
		IntegrationType: "api",
		Description:     "Reads SendOwl orders, products, and subscriptions through the SendOwl API.",
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
	return r.DoJSON(ctx, http.MethodGet, "orders", nil, nil, nil)
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
		stream = "orders"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("sendowl stream %q not found", req.Stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, req.State, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := positiveInt(req.Config.Config["page_size"], defaultPageSize, 1, 500, "page_size")
	if err != nil {
		return err
	}
	maxPages, err := parseMaxPages(req.Config.Config["max_pages"])
	if err != nil {
		return err
	}
	pager := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, ep.resource, nil, pager, ep.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(connectors.Record(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	username := strings.TrimSpace(cfg.Config["username"])
	password := strings.TrimSpace(secret(cfg, "password"))
	if username == "" {
		return nil, errors.New("sendowl connector requires config username")
	}
	if password == "" {
		return nil, errors.New("sendowl connector requires secret password")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: baseURL(cfg, defaultBaseURL), Auth: connsdk.Basic(username, password), UserAgent: userAgent}, nil
}

type streamEndpoint struct{ resource, recordsPath string }

var streamEndpoints = map[string]streamEndpoint{
	"orders":        {resource: "orders", recordsPath: ""},
	"products":      {resource: "products", recordsPath: ""},
	"subscriptions": {resource: "subscriptions", recordsPath: ""},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "string"}}
	return []connectors.Stream{
		{Name: "orders", Description: "SendOwl orders.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: append(fields, connectors.Field{Name: "buyer_email", Type: "string"})},
		{Name: "products", Description: "SendOwl products.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "subscriptions", Description: "SendOwl subscriptions.", PrimaryKey: []string{"id"}, Fields: fields},
	}
}

func readFixture(ctx context.Context, stream string, state map[string]string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": i, "name": fmt.Sprintf("Fixture %s %d", stream, i), "buyer_email": fmt.Sprintf("buyer+%d@example.com", i), "created_at": fixtureUpdatedAt}
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
