// Package publicapis implements a read-only native connector for the Public APIs
// directory API.
package publicapis

import (
	"bytes"
	"context"
	"encoding/json"
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
	defaultBaseURL  = "https://api.publicapis.org"
	defaultPageSize = 100
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("public-apis", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

func (Connector) Name() string { return "public-apis" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "public-apis",
		DisplayName:     "Public APIs",
		IntegrationType: "api",
		Description:     "Reads public API directory entries and categories. Read-only and credential-free.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "entries", url.Values{"limit": []string{"1"}, "offset": []string{"0"}}, nil, nil); err != nil {
		return fmt.Errorf("check public-apis: %w", err)
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
		stream = "entries"
	}
	if stream != "entries" && stream != "categories" {
		return fmt.Errorf("public-apis stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	if stream == "categories" {
		return c.readCategories(ctx, r, emit)
	}
	return c.readEntries(ctx, r, req.Config, emit)
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) readEntries(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	pageSize, err := pageSize(cfg)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(cfg)
	if err != nil {
		return err
	}
	for offset, page := 0, 0; maxPages == 0 || page < maxPages; offset, page = offset+pageSize, page+1 {
		resp, err := r.Do(ctx, http.MethodGet, "entries", url.Values{"limit": []string{strconv.Itoa(pageSize)}, "offset": []string{strconv.Itoa(offset)}}, nil)
		if err != nil {
			return fmt.Errorf("read public-apis entries: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "entries")
		if err != nil {
			return fmt.Errorf("decode public-apis entries: %w", err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapEntry(item)); err != nil {
				return err
			}
		}
		count, err := intAt(resp.Body, "count")
		if err != nil {
			return fmt.Errorf("decode public-apis count: %w", err)
		}
		if len(records) < pageSize || (count > 0 && offset+len(records) >= count) {
			return nil
		}
	}
	return nil
}

func (c Connector) readCategories(ctx context.Context, r *connsdk.Requester, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, "categories", nil, nil)
	if err != nil {
		return fmt.Errorf("read public-apis categories: %w", err)
	}
	var root any
	dec := json.NewDecoder(bytes.NewReader(resp.Body))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return fmt.Errorf("decode public-apis categories: %w", err)
	}
	values := stringSlice(root)
	if obj, ok := root.(map[string]any); ok {
		values = stringSlice(obj["categories"])
	}
	for _, category := range values {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": category, "category": category}); err != nil {
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

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "entries", Description: "Public API directory entries.", Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "api", Type: "string"}, {Name: "description", Type: "string"}, {Name: "category", Type: "string"}}, PrimaryKey: []string{"id"}},
		{Name: "categories", Description: "Public API directory categories.", Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "category", Type: "string"}}, PrimaryKey: []string{"id"}},
	}
}

func mapEntry(item map[string]any) connectors.Record {
	api := first(item, "API", "api")
	return connectors.Record{
		"id":          api,
		"api":         api,
		"description": first(item, "Description", "description"),
		"category":    first(item, "Category", "category"),
		"auth":        first(item, "Auth", "auth"),
		"https":       first(item, "HTTPS", "https"),
		"cors":        first(item, "Cors", "cors"),
		"link":        first(item, "Link", "link"),
	}
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if stream == "categories" {
			if err := emit(connectors.Record{"id": fmt.Sprintf("Fixture Category %d", i), "category": fmt.Sprintf("Fixture Category %d", i)}); err != nil {
				return err
			}
			continue
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("Fixture API %d", i), "api": fmt.Sprintf("Fixture API %d", i), "description": "Deterministic fixture record.", "category": "Fixture"}); err != nil {
			return err
		}
	}
	return nil
}

func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if v := item[key]; v != nil {
			return v
		}
	}
	return nil
}

func stringSlice(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func intAt(body []byte, path string) (int, error) {
	value, err := connsdk.StringAt(body, path)
	if err != nil || strings.TrimSpace(value) == "" {
		return 0, err
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		base = defaultBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("public-apis config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("public-apis config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("public-apis config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > 500 {
		return 0, errors.New("public-apis config page_size must be between 1 and 500")
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
		return 0, errors.New("public-apis config max_pages must be 0, all, unlimited, or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
