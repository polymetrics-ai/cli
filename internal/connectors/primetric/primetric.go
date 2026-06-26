// Package primetric implements a conservative read-only connector for the
// Primetric REST API using OAuth2 client credentials and documented paginated
// list-style resources.
package primetric

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
	defaultBaseURL  = "https://api.primetric.com/api/v1"
	defaultTokenURL = "https://api.primetric.com/oauth/token"
	defaultPageSize = 50
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("primetric", New)
}

func New() connectors.Connector { return Connector{} }

type Connector struct {
	Client *http.Client
}

var streamPaths = map[string]string{
	"employees": "employees",
	"projects":  "projects",
	"clients":   "clients",
	"roles":     "roles",
}

func (Connector) Name() string { return "primetric" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "primetric",
		DisplayName:     "Primetric",
		IntegrationType: "api",
		Description:     "Reads Primetric employees, projects, clients, and roles through OAuth-authenticated REST list endpoints. Read-only.",
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
	if err := r.DoJSON(ctx, http.MethodGet, "employees", url.Values{"page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check primetric: %w", err)
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
		stream = "employees"
	}
	path, ok := streamPaths[stream]
	if !ok {
		return fmt.Errorf("primetric stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		resp, err := r.Do(ctx, http.MethodGet, path, url.Values{"page": []string{strconv.Itoa(page)}}, nil)
		if err != nil {
			return fmt.Errorf("read primetric %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode primetric %s: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		totalPages, err := intAt(resp.Body, "meta.total_pages")
		if err != nil {
			return fmt.Errorf("decode primetric pagination: %w", err)
		}
		if totalPages == 0 || page >= totalPages {
			return nil
		}
	}
	return nil
}

func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	tokenURL, err := tokenURL(cfg)
	if err != nil {
		return nil, err
	}
	clientID := secret(cfg, "client_id")
	clientSecret := secret(cfg, "client_secret")
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return nil, errors.New("primetric connector requires secrets client_id and client_secret")
	}
	return &connsdk.Requester{
		Client:  c.Client,
		BaseURL: base,
		Auth: &connsdk.OAuth2ClientCredentials{
			TokenURL:     tokenURL,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"read"},
			Client:       c.Client,
		},
		UserAgent: userAgent,
	}, nil
}

func streams() []connectors.Stream {
	fields := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "created_at", Type: "timestamp"}, {Name: "updated_at", Type: "timestamp"}}
	return []connectors.Stream{
		{Name: "employees", Description: "Primetric employees.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "projects", Description: "Primetric projects.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "clients", Description: "Primetric clients.", Fields: fields, PrimaryKey: []string{"id"}},
		{Name: "roles", Description: "Primetric roles.", Fields: fields, PrimaryKey: []string{"id"}},
	}
}

func mapRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":         item["id"],
		"name":       first(item, "name", "full_name"),
		"first_name": item["first_name"],
		"last_name":  item["last_name"],
		"email":      item["email"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
		"raw":        item,
	}
	if rec["name"] == nil {
		firstName, _ := item["first_name"].(string)
		lastName, _ := item["last_name"].(string)
		rec["name"] = strings.TrimSpace(firstName + " " + lastName)
	}
	return rec
}

func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": i, "name": fmt.Sprintf("Fixture %s %d", stream, i), "email": fmt.Sprintf("fixture+%d@example.com", i), "updated_at": fmt.Sprintf("2026-01-0%dT00:00:00Z", i)}); err != nil {
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

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	return absoluteURL(cfg.Config["base_url"], defaultBaseURL, "base_url")
}

func tokenURL(cfg connectors.RuntimeConfig) (string, error) {
	return absoluteURL(cfg.Config["token_url"], defaultTokenURL, "token_url")
}

func absoluteURL(value, fallback, name string) (string, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		raw = fallback
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("primetric config %s is invalid: %w", name, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("primetric config %s must use http or https, got %q", name, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("primetric config %s must include a host", name)
	}
	return strings.TrimRight(raw, "/"), nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("primetric config max_pages must be 0, all, unlimited, or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

var _ = defaultPageSize
