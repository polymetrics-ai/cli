package servicenow

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
	connectorName    = "service-now"
	defaultPageSize  = 100
	defaultMaxPages  = 1
	fixtureUpdatedAt = "2026-01-01 00:00:00"
	userAgent        = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "ServiceNow",
		IntegrationType: "api",
		Description:     "Reads ServiceNow table data through the ServiceNow Table API.",
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
	return r.DoJSON(ctx, http.MethodGet, "api/now/table/sys_user", nil, nil, nil)
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
		stream = "incidents"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("service-now stream %q not found", req.Stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, req.State, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := positiveInt(req.Config.Config["page_size"], defaultPageSize, 1, 10000, "page_size")
	if err != nil {
		return err
	}
	maxPages, err := parseMaxPages(req.Config.Config["max_pages"])
	if err != nil {
		return err
	}
	pager := &connsdk.OffsetPaginator{LimitParam: "sysparm_limit", OffsetParam: "sysparm_offset", PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, ep.resource, nil, pager, "result", maxPages, func(rec connsdk.Record) error {
		return emit(connectors.Record(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return nil, errors.New("service-now connector requires config base_url")
	}
	username := strings.TrimSpace(cfg.Config["username"])
	password := strings.TrimSpace(secret(cfg, "password"))
	if username == "" {
		return nil, errors.New("service-now connector requires config username")
	}
	if password == "" {
		return nil, errors.New("service-now connector requires secret password")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: strings.TrimRight(base, "/"), Auth: connsdk.Basic(username, password), UserAgent: userAgent}, nil
}

type streamEndpoint struct{ resource string }

var streamEndpoints = map[string]streamEndpoint{
	"incidents": {resource: "api/now/table/incident"},
	"users":     {resource: "api/now/table/sys_user"},
	"groups":    {resource: "api/now/table/sys_user_group"},
}

func streams() []connectors.Stream {
	common := []connectors.Field{{Name: "sys_id", Type: "string"}, {Name: "number", Type: "string"}, {Name: "name", Type: "string"}, {Name: "updated_on", Type: "string"}}
	return []connectors.Stream{
		{Name: "incidents", Description: "ServiceNow incident table rows.", PrimaryKey: []string{"sys_id"}, CursorFields: []string{"updated_on"}, Fields: common},
		{Name: "users", Description: "ServiceNow user table rows.", PrimaryKey: []string{"sys_id"}, CursorFields: []string{"updated_on"}, Fields: common},
		{Name: "groups", Description: "ServiceNow user group table rows.", PrimaryKey: []string{"sys_id"}, CursorFields: []string{"updated_on"}, Fields: common},
	}
}

func readFixture(ctx context.Context, stream string, state map[string]string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"sys_id": fmt.Sprintf("%s_%d", stream, i), "number": fmt.Sprintf("INC%03d", i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "updated_on": fixtureUpdatedAt}
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
