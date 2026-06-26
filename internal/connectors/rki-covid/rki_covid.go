// Package rkicovid implements a read-only connector for public RKI COVID data.
package rkicovid

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
	defaultBaseURL  = "https://api.corona-zahlen.org"
	defaultPageSize = 100
	maxPageSize     = 1000
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("rki-covid", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "rki-covid" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "rki-covid", DisplayName: "RKI COVID", IntegrationType: "api", Description: "Reads public Germany COVID case, state, district, and history data derived from RKI reports.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "germany", nil, nil, nil); err != nil {
		return fmt.Errorf("check rki-covid: %w", err)
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
		stream = "germany"
	}
	ep, ok := endpoints[stream]
	if !ok {
		return fmt.Errorf("rki-covid stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	query := url.Values{}
	if days := strings.TrimSpace(req.Config.Config["days"]); days != "" {
		query.Set("days", days)
	}
	if _, err := pageSize(req.Config); err != nil {
		return err
	}
	resp, err := r.Do(ctx, http.MethodGet, ep.path, query, nil)
	if err != nil {
		return fmt.Errorf("read rki-covid %s: %w", ep.path, err)
	}
	records, err := recordsAt(resp.Body, ep.recordsPath)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}
	for _, rec := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(mapRecord(stream, rec)); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct{ path, recordsPath string }

var endpoints = map[string]streamEndpoint{
	"germany":        {"germany", ""},
	"states":         {"states", "data"},
	"districts":      {"districts", "data"},
	"cases_history":  {"history/cases", "data"},
	"deaths_history": {"history/deaths", "data"},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "germany", Description: "Germany-wide current COVID metrics.", PrimaryKey: []string{"id"}, Fields: fields("id", "cases", "deaths", "recovered", "week_incidence")},
		{Name: "states", Description: "COVID metrics by German state.", PrimaryKey: []string{"id"}, Fields: fields("id", "name", "cases", "deaths", "week_incidence")},
		{Name: "districts", Description: "COVID metrics by German district.", PrimaryKey: []string{"id"}, Fields: fields("id", "name", "county", "cases", "deaths")},
		{Name: "cases_history", Description: "Germany COVID cases history.", PrimaryKey: []string{"id"}, CursorFields: []string{"date"}, Fields: fields("id", "date", "cases")},
		{Name: "deaths_history", Description: "Germany COVID deaths history.", PrimaryKey: []string{"id"}, CursorFields: []string{"date"}, Fields: fields("id", "date", "deaths")},
	}
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "name": fmt.Sprintf("Fixture %s %d", stream, i), "cases": 100 * i, "deaths": i, "date": "2026-01-01"}); err != nil {
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

func recordsAt(body []byte, path string) ([]connsdk.Record, error) {
	paths := []string{path, "data", "items", "records", ""}
	seen := map[string]bool{}
	for _, candidate := range paths {
		if seen[candidate] {
			continue
		}
		seen[candidate] = true
		records, err := connsdk.RecordsAt(body, candidate)
		if err != nil || len(records) > 0 {
			return records, err
		}
	}
	return nil, nil
}

func mapRecord(stream string, rec connsdk.Record) connectors.Record {
	out := connectors.Record{}
	for k, v := range rec {
		out[k] = v
	}
	if out["id"] == nil {
		out["id"] = first(out, "id", "ags", "abbreviation", "name", "date")
	}
	if out["id"] == nil {
		out["id"] = stream
	}
	out["stream"] = stream
	return out
}

func fields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
}

func first(record connectors.Record, keys ...string) any {
	for _, key := range keys {
		if v := record[key]; v != nil {
			return v
		}
	}
	return nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("rki-covid config base_url is invalid: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("rki-covid config base_url must use http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return "", errors.New("rki-covid config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	return boundedInt("rki-covid", cfg.Config["page_size"], defaultPageSize, maxPageSize, "page_size")
}

func boundedInt(connector, raw string, def, max int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return def, nil
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || v < 1 || v > max {
		return 0, fmt.Errorf("%s config %s must be an integer between 1 and %d", connector, name, max)
	}
	return v, nil
}
