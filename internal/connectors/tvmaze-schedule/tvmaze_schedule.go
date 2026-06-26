// Package tvmazeschedule implements a read-only native connector for TVmaze schedule APIs.
package tvmazeschedule

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	defaultBaseURL = "https://api.tvmaze.com"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("tvmaze-schedule", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "tvmaze-schedule" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "tvmaze-schedule", DisplayName: "TVmaze Schedule", IntegrationType: "api", Description: "Reads public TVmaze broadcast and web schedules without credentials.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
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
	if err := r.DoJSON(ctx, http.MethodGet, "schedule", url.Values{"country": []string{"US"}}, nil, nil); err != nil {
		return fmt.Errorf("check tvmaze-schedule: %w", err)
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
		stream = "schedule"
	}
	spec, ok := streamSpecs[stream]
	if !ok {
		return fmt.Errorf("tvmaze-schedule stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, spec, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	query := url.Values{}
	if v := strings.TrimSpace(req.Config.Config["country"]); v != "" {
		query.Set("country", v)
	}
	if v := strings.TrimSpace(req.Config.Config["date"]); v != "" {
		query.Set("date", v)
	}
	resp, err := r.Do(ctx, http.MethodGet, spec.resource, query, nil)
	if err != nil {
		return fmt.Errorf("read tvmaze-schedule %s: %w", spec.resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, ".")
	if err != nil {
		return fmt.Errorf("decode tvmaze-schedule %s: %w", spec.resource, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(spec.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readFixture(ctx context.Context, stream string, spec streamSpec, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": float64(100 + i), "name": fmt.Sprintf("Fixture episode %d", i), "airdate": "2026-01-01", "airtime": "20:00", "show": map[string]any{"id": float64(i), "name": "Fixture Show"}}
		rec := spec.mapRecord(item)
		rec["connector"] = "tvmaze-schedule"
		rec["fixture"] = true
		if err := emit(rec); err != nil {
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

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("tvmaze-schedule config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("tvmaze-schedule config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("tvmaze-schedule config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamSpec struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

var streamSpecs = map[string]streamSpec{
	"schedule":     {resource: "schedule", mapRecord: episodeRecord},
	"web_schedule": {resource: "web/schedule", mapRecord: episodeRecord},
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "airdate", Type: "string"}, {Name: "airtime", Type: "string"}, {Name: "show_id", Type: "integer"}, {Name: "show_name", Type: "string"}}
	return []connectors.Stream{
		{Name: "schedule", Description: "TVmaze broadcast schedule episodes.", PrimaryKey: []string{"id"}, CursorFields: []string{"airdate"}, Fields: fields},
		{Name: "web_schedule", Description: "TVmaze web schedule episodes.", PrimaryKey: []string{"id"}, CursorFields: []string{"airdate"}, Fields: fields},
	}
}

func episodeRecord(item map[string]any) connectors.Record {
	show, _ := item["show"].(map[string]any)
	return connectors.Record{"id": item["id"], "name": item["name"], "airdate": item["airdate"], "airtime": item["airtime"], "show_id": show["id"], "show_name": show["name"]}
}
