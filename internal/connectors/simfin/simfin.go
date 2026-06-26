// Package simfin implements a read-only native SimFin connector.
package simfin

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
	simfinDefaultBaseURL  = "https://backend.simfin.com"
	simfinDefaultPageSize = 100
	simfinMaxPageSize     = 1000
	simfinUserAgent       = "polymetrics-go-cli"
)

func init()                     { connectors.RegisterFactory("simfin", New) }
func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "simfin" }
func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "simfin", DisplayName: "SimFin", IntegrationType: "api", Description: "Reads SimFin companies, statements, and markets through the SimFin REST API.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true}}
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
	if err := r.DoJSON(ctx, http.MethodGet, simfinEndpoints["companies"].path, url.Values{"limit": []string{"1"}, "page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check simfin: %w", err)
	}
	return nil
}
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams(simfinEndpoints)}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "companies"
	}
	endpoint, ok := simfinEndpoints[stream]
	if !ok {
		return fmt.Errorf("simfin stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, endpoint, emit)
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
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		q := url.Values{"limit": []string{strconv.Itoa(pageSize)}, "page": []string{strconv.Itoa(page)}}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, q, nil)
		if err != nil {
			return fmt.Errorf("read simfin %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode simfin %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, simfinDefaultBaseURL, "simfin")
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(secret(cfg, "api_key"))
	if token == "" {
		return nil, errors.New("simfin connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.APIKeyQuery("api-key", token), UserAgent: simfinUserAgent}, nil
}

type streamEndpoint struct {
	path, recordsPath, description string
	fields                         []string
	mapRecord                      func(map[string]any) connectors.Record
}

var simfinEndpoints = map[string]streamEndpoint{
	"companies":  {path: "api/v3/companies/list", recordsPath: "data", description: "SimFin company metadata.", fields: []string{"id", "name", "ticker"}, mapRecord: simfinRecord},
	"statements": {path: "api/v3/statements/list", recordsPath: "data", description: "SimFin statement metadata.", fields: []string{"id", "name", "ticker"}, mapRecord: simfinRecord},
	"markets":    {path: "api/v3/markets/list", recordsPath: "data", description: "SimFin markets.", fields: []string{"id", "name", "ticker"}, mapRecord: simfinRecord},
}

func simfinRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "simId", "companyId"), "name": first(item, "name", "companyName"), "ticker": item["ticker"], "updated_at": first(item, "updated_at", "fiscalPeriod")}
}
func readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := endpoint.mapRecord(map[string]any{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "ticker": fmt.Sprintf("FX%d", i)})
		rec["fixture"] = true
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}
func streams(endpoints map[string]streamEndpoint) []connectors.Stream {
	order := []string{"companies", "statements", "markets"}
	out := make([]connectors.Stream, 0, len(order))
	for _, name := range order {
		ep := endpoints[name]
		out = append(out, connectors.Stream{Name: name, Description: ep.description, PrimaryKey: []string{"id"}, Fields: fields(ep.fields...)})
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
func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}
func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
func baseURL(cfg connectors.RuntimeConfig, fallback, name string) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%s config base_url must use http or https", name)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(base, "/"), nil
}
func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["limit"])
	}
	if raw == "" {
		return simfinDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > simfinMaxPageSize {
		return 0, fmt.Errorf("simfin config page_size must be between 1 and %d", simfinMaxPageSize)
	}
	return value, nil
}
func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("simfin config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
