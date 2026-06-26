// Package avni implements a read-only native connector for Avni-style HTTP API
// exports. It uses basic authentication, bounded page traversal, deterministic
// fixture mode, and self-registration with the connector registry.
package avni

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
	defaultBaseURL  = "https://app.avniproject.org"
	defaultPageSize = 100
	defaultMaxPages = 100
	userAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("avni", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

type streamEndpoint struct {
	name        string
	resource    string
	recordsPath string
	primaryKey  []string
	fields      []connectors.Field
}

var streamEndpoints = map[string]streamEndpoint{
	"subjects": {
		name:        "subjects",
		resource:    "api/subjects",
		recordsPath: "items",
		primaryKey:  []string{"id"},
		fields: []connectors.Field{
			{Name: "id", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "updated_at", Type: "timestamp"},
		},
	},
	"encounters": {
		name:        "encounters",
		resource:    "api/encounters",
		recordsPath: "items",
		primaryKey:  []string{"id"},
		fields: []connectors.Field{
			{Name: "id", Type: "string"},
			{Name: "subject_id", Type: "string"},
			{Name: "encounter_type", Type: "string"},
			{Name: "updated_at", Type: "timestamp"},
		},
	},
}

func (Connector) Name() string { return "avni" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "avni",
		DisplayName:     "Avni",
		IntegrationType: "api",
		Description:     "Reads Avni subjects and encounters through a read-only HTTP API.",
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
	if strings.TrimSpace(cfg.Config["username"]) == "" {
		return errors.New("avni connector requires config username")
	}
	if strings.TrimSpace(secret(cfg, "password")) == "" {
		return errors.New("avni connector requires secret password")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "api/subjects", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check avni: %w", err)
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
		stream = "subjects"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("avni stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := intConfig(req.Config, "page_size", defaultPageSize)
	if err != nil {
		return err
	}
	maxPages, err := intConfig(req.Config, "max_pages", defaultMaxPages)
	if err != nil {
		return err
	}
	return readPaged(ctx, r, endpoint, pageSize, maxPages, req.Config, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func readPaged(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	page := "1"
	for i := 0; i < maxPages; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{"page": []string{page}, "page_size": []string{strconv.Itoa(pageSize)}}
		if start := strings.TrimSpace(cfg.Config["start_date"]); start != "" {
			query.Set("start_date", start)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read avni %s: %w", endpoint.name, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode avni %s: %w", endpoint.name, err)
		}
		for _, rec := range records {
			if err := emit(connectors.Record(rec)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_page")
		if err != nil {
			return fmt.Errorf("decode avni next_page: %w", err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		page = next
	}
	return nil
}

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		rec := connectors.Record{"id": fmt.Sprintf("%s_fixture_%d", stream, i), "updated_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i), "fixture": true}
		if stream == "subjects" {
			rec["name"] = fmt.Sprintf("Fixture Subject %d", i)
		} else {
			rec["subject_id"] = "subjects_fixture_1"
			rec["encounter_type"] = "fixture"
		}
		if err := emit(rec); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, defaultBaseURL)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Basic(cfg.Config["username"], secret(cfg, "password")), UserAgent: userAgent}, nil
}

func streams() []connectors.Stream {
	order := []string{"subjects", "encounters"}
	out := make([]connectors.Stream, 0, len(order))
	for _, name := range order {
		e := streamEndpoints[name]
		out = append(out, connectors.Stream{Name: e.name, Description: "Reads Avni " + e.name + ".", Fields: e.fields, PrimaryKey: e.primaryKey, CursorFields: []string{"updated_at"}})
	}
	return out
}

func baseURL(cfg connectors.RuntimeConfig, fallback string) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = fallback
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("avni config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("avni config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("avni config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func intConfig(cfg connectors.RuntimeConfig, key string, fallback int) (int, error) {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("avni config %s must be a positive integer", key)
	}
	return value, nil
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
