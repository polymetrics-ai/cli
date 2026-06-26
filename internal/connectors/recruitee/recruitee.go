// Package recruitee implements a read-only Recruitee REST connector over core
// ATS list endpoints. It requires a company_id so paths stay allow-listed under
// /c/{company_id}/ rather than accepting arbitrary request paths.
package recruitee

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
	recruiteeDefaultBaseURL  = "https://api.recruitee.com"
	recruiteeDefaultPageSize = 100
	recruiteeMaxPageSize     = 100
	recruiteeUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("recruitee", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "recruitee" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "recruitee", DisplayName: "Recruitee", IntegrationType: "api", Description: "Reads Recruitee offers, candidates, departments, sources, and tags through the Recruitee REST API. Read-only.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, companyID, err := c.requester(cfg)
	if err != nil {
		return err
	}
	if err := r.DoJSON(ctx, http.MethodGet, "c/"+companyID+"/offers", url.Values{"page": []string{"1"}, "limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check recruitee: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: recruiteeStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "offers"
	}
	endpoint, ok := recruiteeEndpoints[stream]
	if !ok {
		return fmt.Errorf("recruitee stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, emit)
	}
	r, companyID, err := c.requester(req.Config)
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
	return c.harvest(ctx, r, companyID, endpoint, pageSize, maxPages, emit)
}

func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, companyID string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("limit", strconv.Itoa(pageSize))
		resp, err := r.Do(ctx, http.MethodGet, "c/"+companyID+"/"+endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read recruitee %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode recruitee %s: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
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

func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{"id": i, "title": fmt.Sprintf("Fixture %s %d", endpoint.path, i), "name": fmt.Sprintf("Fixture %d", i), "status": "published", "email": fmt.Sprintf("fixture+%d@example.com", i), "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z"}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, string, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, "", err
	}
	companyID := cleanSegment(strings.TrimSpace(cfg.Config["company_id"]))
	if companyID == "" {
		return nil, "", errors.New("recruitee connector requires config company_id")
	}
	key := strings.TrimSpace(secret(cfg, "api_key"))
	if key == "" {
		return nil, "", errors.New("recruitee connector requires secret api_key")
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: connsdk.Bearer(key), UserAgent: recruiteeUserAgent}, companyID, nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

type streamEndpoint struct {
	path        string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var recruiteeEndpoints = map[string]streamEndpoint{
	"offers":      {path: "offers", recordsPath: "offers", mapRecord: offerRecord},
	"candidates":  {path: "candidates", recordsPath: "candidates", mapRecord: candidateRecord},
	"departments": {path: "departments", recordsPath: "departments", mapRecord: namedRecord},
	"sources":     {path: "sources", recordsPath: "sources", mapRecord: namedRecord},
	"tags":        {path: "tags", recordsPath: "tags", mapRecord: namedRecord},
}

func recruiteeStreams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "offers", Description: "Recruitee job offers.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "title", "status", "created_at", "updated_at")},
		{Name: "candidates", Description: "Recruitee candidates.", PrimaryKey: []string{"id"}, CursorFields: []string{"updated_at"}, Fields: fields("id", "name", "email", "created_at", "updated_at")},
		{Name: "departments", Description: "Recruitee departments.", PrimaryKey: []string{"id"}, Fields: fields("id", "name")},
		{Name: "sources", Description: "Recruitee candidate sources.", PrimaryKey: []string{"id"}, Fields: fields("id", "name")},
		{Name: "tags", Description: "Recruitee tags.", PrimaryKey: []string{"id"}, Fields: fields("id", "name")},
	}
}

func offerRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "title": item["title"], "status": item["status"], "created_at": item["created_at"], "updated_at": item["updated_at"]}
}
func candidateRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": first(item, "name", "full_name"), "email": item["email"], "created_at": item["created_at"], "updated_at": item["updated_at"]}
}
func namedRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": first(item, "name", "title")}
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func fields(names ...string) []connectors.Field {
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

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return recruiteeDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("recruitee config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("recruitee config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("recruitee config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func cleanSegment(value string) string {
	if value == "" || strings.ContainsAny(value, "/?#") || strings.Contains(value, "..") {
		return ""
	}
	return value
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return recruiteeDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > recruiteeMaxPageSize {
		return 0, fmt.Errorf("recruitee config page_size must be between 1 and %d", recruiteeMaxPageSize)
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
		return 0, errors.New("recruitee config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}
