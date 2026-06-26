package sharetribe

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
	connectorName    = "sharetribe"
	defaultBaseURL   = "https://flex-api.sharetribe.com/v1"
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
		DisplayName:     "Sharetribe",
		IntegrationType: "api",
		Description:     "Reads Sharetribe listings, users, transactions, and events through the Integration API.",
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
	return r.DoJSON(ctx, http.MethodGet, "integration_api/listings/query", nil, nil, nil)
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
		stream = "listings"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("sharetribe stream %q not found", req.Stream)
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
	pager := &connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1, PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, ep.resource, nil, pager, ep.recordsPath, maxPages, func(rec connsdk.Record) error {
		return emit(connectors.Record(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	token := strings.TrimSpace(secret(cfg, "oauth_access_token"))
	if token == "" {
		return nil, errors.New("sharetribe connector requires secret oauth_access_token")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: baseURL(cfg, defaultBaseURL), Auth: connsdk.Bearer(token), UserAgent: userAgent}, nil
}

type streamEndpoint struct{ resource, recordsPath string }

var streamEndpoints = map[string]streamEndpoint{
	"listings":     {resource: "integration_api/listings/query", recordsPath: "data"},
	"users":        {resource: "integration_api/users/query", recordsPath: "data"},
	"transactions": {resource: "integration_api/transactions/query", recordsPath: "data"},
	"events":       {resource: "integration_api/events/query", recordsPath: "data"},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "type", Type: "string"}, {Name: "attributes", Type: "object"}, {Name: "updated_at", Type: "string"}}
	return []connectors.Stream{
		{Name: "listings", Description: "Sharetribe listings.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "users", Description: "Sharetribe users.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "transactions", Description: "Sharetribe transactions.", PrimaryKey: []string{"id"}, Fields: fields},
		{Name: "events", Description: "Sharetribe events.", PrimaryKey: []string{"id"}, Fields: fields},
	}
}

func readFixture(ctx context.Context, stream string, state map[string]string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_%d", stream, i), "type": strings.TrimSuffix(stream, "s"), "attributes": map[string]any{"title": fmt.Sprintf("Fixture %s %d", stream, i)}, "updated_at": fixtureUpdatedAt}
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
