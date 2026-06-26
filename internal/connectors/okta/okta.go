// Package okta implements a read-only Okta connector over the Okta Management
// APIs. It supports API-token auth directly and bearer-token auth when callers
// provide an OAuth access token; private-key OAuth is intentionally not
// implemented because JWT signing/key handling would broaden the dependency and
// secret surface for this native port.
package okta

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
	oktaDefaultPageSize = 200
	oktaMaxPageSize     = 200
	oktaUserAgent       = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory("okta", New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

func (Connector) Name() string { return "okta" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "okta",
		DisplayName:     "Okta",
		IntegrationType: "api",
		Description:     "Reads Okta users, groups, and system log events from the Okta REST APIs.",
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
	return r.DoJSON(ctx, http.MethodGet, "/api/v1/users", url.Values{"limit": []string{"1"}}, nil, nil)
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: "okta", Streams: streams()}, nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "users"
	}
	ep, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("okta stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, ep, req, emit)
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	limit, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	max, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	query := url.Values{"limit": []string{strconv.Itoa(limit)}}
	if ep.sinceParam != "" {
		if since := lowerBound(req); since != "" {
			query.Set(ep.sinceParam, since)
		}
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, ep.path, query, &connsdk.LinkHeaderPaginator{}, ".", max, func(rec connsdk.Record) error {
		return emit(ep.mapRecord(map[string]any(rec)))
	})
}

func readFixture(ctx context.Context, stream string, ep streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":        fmt.Sprintf("%s_fixture_%d", stream, i),
			"status":    "ACTIVE",
			"created":   "2026-01-01T00:00:00Z",
			"lastLogin": "2026-01-02T00:00:00Z",
			"profile": map[string]any{
				"email":       fmt.Sprintf("fixture+%d@example.com", i),
				"login":       fmt.Sprintf("fixture+%d@example.com", i),
				"displayName": fmt.Sprintf("Fixture %d", i),
				"firstName":   "Fixture",
				"lastName":    strconv.Itoa(i),
			},
		}
		rec := ep.mapRecord(item)
		if cursor := req.State[connsdk.CursorStateKey]; cursor != "" {
			rec["previous_cursor"] = cursor
		}
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
	auth, err := authenticator(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{Client: c.Client, BaseURL: base, Auth: auth, UserAgent: oktaUserAgent}, nil
}

func authenticator(cfg connectors.RuntimeConfig) (connsdk.Authenticator, error) {
	if token := secret(cfg, "api_token"); token != "" {
		return connsdk.APIKeyHeader("Authorization", token, "SSWS "), nil
	}
	if token := secret(cfg, "access_token"); token != "" {
		return connsdk.Bearer(token), nil
	}
	return nil, errors.New("okta connector requires secret api_token or access_token")
}

func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		return validateBaseURL("okta", base)
	}
	domain := strings.TrimSpace(cfg.Config["domain"])
	if domain == "" {
		return "", errors.New("okta connector requires config domain or base_url")
	}
	if !strings.Contains(domain, "://") {
		domain = "https://" + domain
	}
	return validateBaseURL("okta", domain)
}

func validateBaseURL(name, raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("%s config base_url is invalid: %w", name, err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("%s config base_url must use http or https, got %q", name, u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("%s config base_url must include a host", name)
	}
	return strings.TrimRight(raw, "/"), nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return strings.TrimSpace(cfg.Secrets[key])
}

func lowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return oktaDefaultPageSize, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > oktaMaxPageSize {
		return 0, fmt.Errorf("okta config page_size must be between 1 and %d", oktaMaxPageSize)
	}
	return n, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.ToLower(strings.TrimSpace(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errors.New("okta config max_pages must be a non-negative integer")
	}
	return n, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

type streamEndpoint struct {
	path       string
	sinceParam string
	mapRecord  func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"users":       {path: "/api/v1/users", mapRecord: userRecord},
	"groups":      {path: "/api/v1/groups", mapRecord: groupRecord},
	"system_logs": {path: "/api/v1/logs", sinceParam: "since", mapRecord: logRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "users", Description: "Okta users.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "status", Type: "string"}, {Name: "email", Type: "string"}, {Name: "login", Type: "string"}, {Name: "created", Type: "timestamp"}, {Name: "last_login", Type: "timestamp"}}},
		{Name: "groups", Description: "Okta groups.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "description", Type: "string"}, {Name: "created", Type: "timestamp"}}},
		{Name: "system_logs", Description: "Okta system log events.", PrimaryKey: []string{"uuid"}, CursorFields: []string{"published"}, Fields: []connectors.Field{{Name: "uuid", Type: "string"}, {Name: "published", Type: "timestamp"}, {Name: "event_type", Type: "string"}, {Name: "display_message", Type: "string"}}},
	}
}

func userRecord(item map[string]any) connectors.Record {
	profile := object(item, "profile")
	return connectors.Record{"id": item["id"], "status": item["status"], "email": profile["email"], "login": profile["login"], "created": item["created"], "last_login": item["lastLogin"]}
}

func groupRecord(item map[string]any) connectors.Record {
	profile := object(item, "profile")
	return connectors.Record{"id": item["id"], "name": profile["name"], "description": profile["description"], "created": item["created"]}
}

func logRecord(item map[string]any) connectors.Record {
	return connectors.Record{"uuid": item["uuid"], "published": item["published"], "event_type": item["eventType"], "display_message": item["displayMessage"]}
}

func object(item map[string]any, key string) map[string]any {
	if v, ok := item[key].(map[string]any); ok {
		return v
	}
	return map[string]any{}
}
