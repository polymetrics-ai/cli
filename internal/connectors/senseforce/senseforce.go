package senseforce

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
	connectorName    = "senseforce"
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
		DisplayName:     "Senseforce",
		IntegrationType: "api",
		Description:     "Reads records from a configured Senseforce dataset through the Senseforce API.",
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
	r, resource, err := c.requesterAndResource(cfg)
	if err != nil {
		return err
	}
	return r.DoJSON(ctx, http.MethodGet, resource, url.Values{"page_size": []string{"1"}}, nil, nil)
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
		stream = "records"
	}
	if stream != "records" {
		return fmt.Errorf("senseforce stream %q not found", req.Stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, req.State, emit)
	}
	r, resource, err := c.requesterAndResource(req.Config)
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
	return connsdk.Harvest(ctx, r, http.MethodGet, resource, nil, pager, "data", maxPages, func(rec connsdk.Record) error {
		return emit(connectors.Record(rec))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requesterAndResource(cfg connectors.RuntimeConfig) (*connsdk.Requester, string, error) {
	token := strings.TrimSpace(secret(cfg, "access_token"))
	if token == "" {
		return nil, "", errors.New("senseforce connector requires secret access_token")
	}
	base := strings.TrimSpace(cfg.Config["backend_url"])
	if base == "" {
		return nil, "", errors.New("senseforce connector requires config backend_url")
	}
	datasetID := strings.TrimSpace(cfg.Config["dataset_id"])
	if datasetID == "" {
		return nil, "", errors.New("senseforce connector requires config dataset_id")
	}
	resource := "api/v1/datasets/" + url.PathEscape(datasetID) + "/records"
	return &connsdk.Requester{Client: c.Client, BaseURL: strings.TrimRight(base, "/"), Auth: connsdk.Bearer(token), UserAgent: userAgent}, resource, nil
}

func streams() []connectors.Stream {
	return []connectors.Stream{{Name: "records", Description: "Rows from the configured Senseforce dataset.", PrimaryKey: []string{"id"}, CursorFields: []string{"Timestamp"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "Timestamp", Type: "string"}, {Name: "value", Type: "number"}}}}
}

func readFixture(ctx context.Context, state map[string]string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("row_%d", i), "Timestamp": fixtureUpdatedAt, "value": i}
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
