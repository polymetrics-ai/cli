// Package pennylane implements a conservative read-only Pennylane API connector.
// It covers the documented v2 list endpoints that share Bearer auth and
// cursor-based pagination ({items, next_cursor}). Writes are intentionally not
// exposed because Pennylane's write surface mutates accounting data.
package pennylane

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
	defaultBaseURL  = "https://app.pennylane.com/api/external/v2"
	defaultPageSize = 50
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("pennylane", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "pennylane" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "pennylane",
		DisplayName:     "Pennylane",
		IntegrationType: "api",
		Description:     "Reads Pennylane v2 customers, customer invoices, suppliers, products, and categories through the REST API.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "customers", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check pennylane: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("pennylane stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	limit, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	base := url.Values{"limit": []string{strconv.Itoa(limit)}}
	if filter := strings.TrimSpace(req.Config.Config["filter"]); filter != "" {
		base.Set("filter", filter)
	}
	if sort := strings.TrimSpace(req.Config.Config["sort"]); sort != "" {
		base.Set("sort", sort)
	}
	p := &connsdk.CursorPaginator{CursorParam: "cursor", TokenPath: "next_cursor"}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, p, "items", maxPages, func(rec connsdk.Record) error {
		return emit(connectors.Record(rec))
	})
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ resource string }

var streamEndpoints = map[string]streamEndpoint{
	"customers":         {resource: "customers"},
	"customer_invoices": {resource: "customer_invoices"},
	"suppliers":         {resource: "suppliers"},
	"products":          {resource: "products"},
	"categories":        {resource: "categories"},
}

func streams() []connectors.Stream {
	names := []string{"customers", "customer_invoices", "suppliers", "products", "categories"}
	out := make([]connectors.Stream, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Stream{
			Name:         name,
			Description:  "Pennylane " + strings.ReplaceAll(name, "_", " ") + " list endpoint.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       commonFields(),
		})
	}
	return out
}

func commonFields() []connectors.Field {
	return []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": i, "name": fmt.Sprintf("Fixture %s %d", stream, i), "updated_at": "2026-01-01T00:00:00Z"}); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	key := secret(cfg, "api_key")
	if strings.TrimSpace(key) == "" {
		return nil, errors.New("pennylane connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: userAgent}, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("pennylane config base_url is invalid: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("pennylane config base_url must use http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return "", errors.New("pennylane config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt(cfg.Config["page_size"], defaultPageSize, maxPageSize, "page_size")
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return optionalInt(cfg.Config["max_pages"], "max_pages")
}

func boundedInt(raw string, def, max int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v < 1 || v > max {
		return 0, fmt.Errorf("pennylane config %s must be an integer between 1 and %d", name, max)
	}
	return v, nil
}

func optionalInt(raw, name string) (int, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return 0, fmt.Errorf("pennylane config %s must be 0, positive, all, or unlimited", name)
	}
	return v, nil
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
