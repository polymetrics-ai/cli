// Package pokeapi implements a read-only connector for the public PokeAPI.
package pokeapi

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
	defaultBaseURL  = "https://pokeapi.co/api/v2"
	defaultPageSize = 100
	defaultMaxPages = 3
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("pokeapi", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "pokeapi" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "pokeapi",
		DisplayName:     "PokeAPI",
		IntegrationType: "api",
		Description:     "Reads Pokemon, types, abilities, and moves from the public PokeAPI.",
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
	if _, err := r.Do(ctx, http.MethodGet, "pokemon", url.Values{"limit": {"1"}, "offset": {"0"}}, nil); err != nil {
		return fmt.Errorf("check pokeapi: %w", err)
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
		stream = "pokemon"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("pokeapi stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, endpoint, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	p := &connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam: "offset", PageSize: pageSize}
	return connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, nil, p, "results", maxPages, func(item connsdk.Record) error {
		return emit(endpoint.mapRecord(item))
	})
}

func (Connector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"pokemon":   {resource: "pokemon", mapRecord: namedResourceRecord},
	"types":     {resource: "type", mapRecord: namedResourceRecord},
	"abilities": {resource: "ability", mapRecord: namedResourceRecord},
	"moves":     {resource: "move", mapRecord: namedResourceRecord},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "url", Type: "string"}}
	return []connectors.Stream{
		{Name: "pokemon", Description: "Pokemon resource list.", PrimaryKey: []string{"name"}, Fields: fields},
		{Name: "types", Description: "Pokemon type resource list.", PrimaryKey: []string{"name"}, Fields: fields},
		{Name: "abilities", Description: "Pokemon ability resource list.", PrimaryKey: []string{"name"}, Fields: fields},
		{Name: "moves", Description: "Pokemon move resource list.", PrimaryKey: []string{"name"}, Fields: fields},
	}
}

func namedResourceRecord(item map[string]any) connectors.Record {
	urlValue, _ := item["url"].(string)
	return connectors.Record{"id": idFromURL(urlValue), "name": item["name"], "url": urlValue}
}

func readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"name": fmt.Sprintf("fixture-%s-%d", endpoint.resource, i), "url": fmt.Sprintf("https://pokeapi.co/api/v2/%s/%d/", endpoint.resource, i)}
		if err := emit(endpoint.mapRecord(item)); err != nil {
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
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: userAgent}, nil
}

func idFromURL(raw string) string {
	trimmed := strings.Trim(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	return parts[len(parts)-1]
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("pokeapi config base_url is invalid: %w", err)
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return "", errors.New("pokeapi config base_url must be an absolute http or https URL")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "page_size", defaultPageSize, 1, 1000)
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	return intConfig(cfg, "max_pages", defaultMaxPages, 0, 10000)
}

func intConfig(cfg connectors.RuntimeConfig, key string, def, min, max int) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config[key]))
	if raw == "" {
		return def, nil
	}
	if key == "max_pages" && (raw == "all" || raw == "unlimited") {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min {
		return 0, fmt.Errorf("pokeapi config %s must be an integer >= %d", key, min)
	}
	if max > 0 && value > max {
		return max, nil
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
