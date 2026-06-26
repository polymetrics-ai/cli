// Package qualaroo implements a conservative read-only Qualaroo HTTP connector.
// It covers the stable API v1 list endpoints used for survey definitions
// (nudges) and responses. Qualaroo accounts can expose additional per-survey
// response paths; this connector intentionally keeps the no-deps surface to the
// documented list-style endpoints and allows base_url overrides for tests/proxies.
package qualaroo

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
	qualarooDefaultBaseURL  = "https://api.qualaroo.com/api/v1"
	qualarooDefaultPageSize = 100
	qualarooMaxPageSize     = 500
	qualarooUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("qualaroo", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "qualaroo" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "qualaroo",
		DisplayName:     "Qualaroo",
		IntegrationType: "api",
		Description:     "Reads Qualaroo nudges and response records through the Qualaroo API. Read-only.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "nudges", url.Values{"page": []string{"1"}, "per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check qualaroo: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: qualarooStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "nudges"
	}
	endpoint, ok := qualarooEndpoints[stream]
	if !ok {
		return fmt.Errorf("qualaroo stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
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

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	page := 1
	for fetched := 0; maxPages == 0 || fetched < maxPages; fetched++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read qualaroo %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode qualaroo %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "pagination.next_page")
		if err != nil {
			return fmt.Errorf("decode qualaroo %s next_page: %w", endpoint.path, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		nextPage, err := strconv.Atoi(next)
		if err != nil || nextPage <= page {
			return nil
		}
		page = nextPage
	}
	return nil
}

func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": fmt.Sprintf("%s_fixture_%d", endpoint.path, i), "name": fmt.Sprintf("Fixture %s %d", endpoint.path, i), "status": "active", "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z", "email": fmt.Sprintf("fixture+%d@example.com", i)}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg, qualarooDefaultBaseURL, true)
	if err != nil {
		return nil, err
	}
	secret := strings.TrimSpace(secret(cfg, "api_key"))
	if secret == "" {
		return nil, errors.New("qualaroo connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, UserAgent: qualarooUserAgent, Auth: connsdk.AuthFunc(func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", `Token token="`+secret+`"`)
		return nil
	})}, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path        string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var qualarooEndpoints = map[string]streamEndpoint{
	"nudges":    {path: "nudges", recordsPath: "nudges", mapRecord: nudgeRecord},
	"responses": {path: "responses", recordsPath: "responses", mapRecord: responseRecord},
}

func qualarooStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "nudges", Description: "Qualaroo survey nudges.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: baseFields("id", "name", "status", "created_at", "updated_at")},
		{Name: "responses", Description: "Qualaroo response records when exposed by the API account.", PrimaryKey: []string{"id"}, CursorFields: []string{"created_at"}, Fields: baseFields("id", "nudge_id", "email", "created_at", "updated_at")},
	}
}

func nudgeRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id"), "name": first(item, "name", "title"), "status": first(item, "status"), "created_at": first(item, "created_at"), "updated_at": first(item, "updated_at")}
}

func responseRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": first(item, "id", "response_id"), "nudge_id": first(item, "nudge_id", "survey_id"), "email": first(item, "email"), "created_at": first(item, "created_at"), "updated_at": first(item, "updated_at")}
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func baseFields(names ...string) []connectors.Field {
	out := make([]connectors.Field, 0, len(names))
	for _, name := range names {
		out = append(out, connectors.Field{Name: name, Type: "string"})
	}
	return out
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

func baseURL(cfg connectors.RuntimeConfig, def string, appendAPIV1 bool) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return def, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("qualaroo config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("qualaroo config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("qualaroo config base_url must include a host")
	}
	base = strings.TrimRight(base, "/")
	if appendAPIV1 && !strings.HasSuffix(base, "/api/v1") {
		base += "/api/v1"
	}
	return base, nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return qualarooDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > qualarooMaxPageSize {
		return 0, fmt.Errorf("qualaroo config page_size must be between 1 and %d", qualarooMaxPageSize)
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
		return 0, errors.New("qualaroo config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
