// Package jira implements the native pm Jira connector. It is a declarative-HTTP
// per-system connector built on the connsdk toolkit: a thin package that composes
// a connsdk Requester (HTTP Basic auth from the configured email + api_token
// secret) with Jira-specific stream definitions, endpoints, and the Jira REST v3
// offset pagination style (startAt/maxResults/total).
//
// It follows the stripe template shape. Like the other per-system connectors it
// self-registers with the connectors registry via RegisterFactory in init(); the
// registryset package blank-imports this package in the production binary to run
// that side effect.
//
// Jira is read-only: there is no reverse-ETL write path, so Capabilities.Write is
// false and Write returns ErrUnsupportedOperation.
package jira

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
	jiraAPIPrefix       = "rest/api/3"
	jiraDefaultPageSize = 50
	jiraMaxPageSize     = 100
	jiraUserAgent       = "polymetrics-go-cli"
	// jiraFixtureUpdated is the deterministic `updated` timestamp used by the
	// fixture-mode records.
	jiraFixtureUpdated = "2026-01-01T00:00:00.000+0000"
)

func init() {
	connectors.RegisterFactory("jira", New)
}

// New returns the Jira connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Jira connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "jira" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "jira",
		DisplayName:     "Jira",
		IntegrationType: "api",
		Description:     "Reads Jira issues, projects, and users through the Jira Cloud REST API v3 using HTTP Basic auth (email + API token). Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Jira. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := jiraBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(jiraEmail(cfg)) == "" {
		return errors.New("jira connector requires config email")
	}
	if strings.TrimSpace(jiraSecret(cfg)) == "" {
		return errors.New("jira connector requires secret credentials.api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the current user confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, jiraAPIPrefix+"/myself", nil, nil, nil); err != nil {
		return fmt.Errorf("check jira: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: jiraStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Jira stream starts with an
// empty incremental cursor (full sync).
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "issues"
	}
	endpoint, ok := jiraStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("jira stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := jiraPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := jiraMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Jira's offset pagination. Paginated Jira responses are
// {startAt, maxResults, total, <records>[]}; the next page is requested with
// startAt advanced by the number of records returned, stopping once
// startAt+len >= total (or a short/empty page is seen, which guards against a
// missing/zero total on array-shaped endpoints like users/search).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	startAt := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("startAt", strconv.Itoa(startAt))
		query.Set("maxResults", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, jiraAPIPrefix+"/"+endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read jira %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode jira %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		count := len(records)
		if count == 0 || count < pageSize {
			return nil
		}
		startAt += count
		if total, ok := pageTotal(resp.Body); ok && startAt >= total {
			return nil
		}
	}
	return nil
}

// pageTotal reads the offset envelope's `total` field, returning false when the
// endpoint does not report one (e.g. the bare-array users/search response).
func pageTotal(body []byte) (int, bool) {
	var env struct {
		Total *json.Number `json:"total"`
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&env); err != nil || env.Total == nil {
		return 0, false
	}
	n, err := env.Total.Int64()
	if err != nil {
		return 0, false
	}
	return int(n), true
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise jira credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := jiraStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		var item map[string]any
		switch stream {
		case "projects":
			item = map[string]any{
				"id":             fmt.Sprintf("%d", 10000+i),
				"key":            fmt.Sprintf("PROJ%d", i),
				"self":           fmt.Sprintf("https://fixture.atlassian.net/rest/api/3/project/%d", i),
				"name":           fmt.Sprintf("Fixture Project %d", i),
				"projectTypeKey": "software",
				"simplified":     false,
				"style":          "classic",
				"isPrivate":      false,
			}
		case "users":
			item = map[string]any{
				"accountId":    fmt.Sprintf("fixture-account-%d", i),
				"accountType":  "atlassian",
				"self":         fmt.Sprintf("https://fixture.atlassian.net/rest/api/3/user?accountId=fixture-account-%d", i),
				"displayName":  fmt.Sprintf("Fixture User %d", i),
				"emailAddress": fmt.Sprintf("fixture+%d@example.com", i),
				"active":       true,
			}
		default: // issues
			item = map[string]any{
				"id":   fmt.Sprintf("%d", 10000+i),
				"key":  fmt.Sprintf("FIX-%d", i),
				"self": fmt.Sprintf("https://fixture.atlassian.net/rest/api/3/issue/%d", i),
				"fields": map[string]any{
					"summary":   fmt.Sprintf("Fixture issue %d", i),
					"created":   jiraFixtureUpdated,
					"updated":   jiraFixtureUpdated,
					"status":    map[string]any{"name": "Open"},
					"issuetype": map[string]any{"name": "Task"},
					"priority":  map[string]any{"name": "Medium"},
					"assignee":  map[string]any{"displayName": fmt.Sprintf("Fixture User %d", i)},
					"reporter":  map[string]any{"displayName": "Fixture Reporter"},
					"project":   map[string]any{"key": "FIX"},
				},
			}
		}
		record := endpoint.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with HTTP Basic auth (email +
// api_token) and the resolved base URL. The api_token only ever flows into
// connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := jiraBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	email := strings.TrimSpace(jiraEmail(cfg))
	if email == "" {
		return nil, errors.New("jira connector requires config email")
	}
	token := strings.TrimSpace(jiraSecret(cfg))
	if token == "" {
		return nil, errors.New("jira connector requires secret credentials.api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(email, token),
		UserAgent: jiraUserAgent,
	}, nil
}

func jiraEmail(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["email"]
}

func jiraSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	// Primary key per the catalog; accept a couple of conventional fallbacks so
	// callers that flatten differently still resolve the token.
	for _, key := range []string{"credentials.api_token", "api_token", "credentials_api_token"} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// jiraBaseURL resolves and validates the base URL. The default is derived from
// the configured Jira `domain` (https://<domain>); an explicit base_url override
// must be an absolute https (or http for local test servers) URL with a host to
// bound SSRF risk.
func jiraBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		domain := strings.TrimSpace(cfg.Config["domain"])
		if domain == "" {
			return "", errors.New("jira connector requires config domain or base_url")
		}
		domain = strings.TrimPrefix(domain, "https://")
		domain = strings.TrimPrefix(domain, "http://")
		domain = strings.TrimRight(domain, "/")
		if domain == "" || strings.ContainsAny(domain, "/ ") {
			return "", fmt.Errorf("jira config domain %q is invalid (host only, no scheme or path)", cfg.Config["domain"])
		}
		return "https://" + domain, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("jira config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("jira config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("jira config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func jiraPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return jiraDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("jira config page_size must be an integer: %w", err)
	}
	if value < 1 || value > jiraMaxPageSize {
		return 0, fmt.Errorf("jira config page_size must be between 1 and %d", jiraMaxPageSize)
	}
	return value, nil
}

func jiraMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("jira config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("jira config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Jira is read-only in this
// connector, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
