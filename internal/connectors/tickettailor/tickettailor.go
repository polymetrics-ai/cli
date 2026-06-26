// Package tickettailor implements a read-only native Go connector for Ticket Tailor.
package tickettailor

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
	connectorName   = "tickettailor"
	defaultBaseURL  = "https://api.tickettailor.com/v1"
	defaultPageSize = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "Ticket Tailor", IntegrationType: "api", Description: "Reads events, orders, and issued tickets from the Ticket Tailor API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "events", url.Values{"page": {"1"}, "limit": {"1"}}, nil, nil); err != nil {
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
		stream = "events"
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
	return connsdk.Harvest(ctx, r, http.MethodGet, spec.path, nil, p, spec.recordsPath, max, func(rec connsdk.Record) error { return emit(connectors.Record(rec)) })
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
	if key == "" {
		return nil, errors.New("tickettailor connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(key, ""), UserAgent: userAgent}, nil
}

type streamSpec struct {
	name, description, path, recordsPath string
	fields                               []connectors.Field
}

var streamSpecs = map[string]streamSpec{
	"events":         {"events", "Ticket Tailor events.", "events", "data", fields("id", "name", "start_date", "end_date", "status")},
	"orders":         {"orders", "Ticket Tailor orders.", "orders", "data", fields("id", "event_id", "email", "total", "created_at")},
	"issued_tickets": {"issued_tickets", "Ticket Tailor issued tickets.", "issued_tickets", "data", fields("id", "event_id", "order_id", "ticket_type_id", "status")},
}

func streams() []connectors.Stream {
	order := []string{"events", "orders", "issued_tickets"}
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
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", spec.name, i), "name": fmt.Sprintf("Fixture %s %d", spec.name, i), "email": fmt.Sprintf("fixture+%d@example.com", i), "fixture": true}
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
