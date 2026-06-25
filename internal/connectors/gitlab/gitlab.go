// Package gitlab implements the native pm GitLab connector. It follows the
// declarative-HTTP per-system template established by the stripe connector: a
// thin package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + Link-header pagination + cursor state) with
// GitLab-specific stream definitions and endpoints.
//
// Like stripe/github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// GitLab REST API v4 notes:
//   - Base URL: https://gitlab.com/api/v4 (configurable for self-managed via
//     the api_url / base_url config).
//   - Auth: Authorization: Bearer <token>. A personal access token and an OAuth
//     access token both work as a bearer credential.
//   - Pagination: offset pagination with page/per_page query params and an
//     RFC 5988 Link header carrying rel="next" — handled by
//     connsdk.LinkHeaderPaginator + connsdk.Harvest.
//   - Collections are returned as top-level JSON arrays (records path "").
//
// This connector is read-only: GitLab has no obvious safe reverse-ETL write for
// the chosen streams, so Capabilities.Write is false and Write returns
// ErrUnsupportedOperation.
package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	gitlabDefaultBaseURL  = "https://gitlab.com/api/v4"
	gitlabDefaultPageSize = 50
	gitlabMaxPageSize     = 100
	gitlabUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("gitlab", New)
}

// New returns the GitLab connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm GitLab connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "gitlab" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "gitlab",
		DisplayName:     "GitLab",
		IntegrationType: "api",
		Description:     "Reads GitLab projects, groups, users, and issues through the GitLab REST API v4.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to GitLab. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gitlabBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(gitlabSecret(cfg)) == "" {
		return errors.New("gitlab connector requires secret credentials.access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the current-user endpoint confirms auth and connectivity
	// without listing any collection.
	if err := r.DoJSON(ctx, http.MethodGet, "user", nil, nil, nil); err != nil {
		return fmt.Errorf("check gitlab: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gitlabStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a GitLab stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "projects"
	}
	endpoint, ok := gitlabStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("gitlab stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := gitlabPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gitlabMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	// Best-effort incremental lower bound. GitLab list endpoints accept
	// updated_after / created_after depending on the resource; we apply the
	// start_date config as the matching filter for the stream when present.
	if lower := strings.TrimSpace(req.Config.Config["start_date"]); lower != "" {
		if param := gitlabSinceParam[stream]; param != "" {
			base.Set(param, lower)
		}
	}

	paginator := &connsdk.LinkHeaderPaginator{FirstQuery: url.Values{"page": []string{"1"}}}
	wrapped := func(item connsdk.Record) error {
		return emit(endpoint.mapRecord(item))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, endpoint.resource, base, paginator, "", maxPages, wrapped); err != nil {
		return fmt.Errorf("read gitlab %s: %w", endpoint.resource, err)
	}
	return nil
}

// gitlabSinceParam maps a stream to the GitLab query parameter that filters by a
// lower-bound timestamp, used to apply the start_date config incrementally.
var gitlabSinceParam = map[string]string{
	"projects": "last_activity_after",
	"groups":   "",
	"users":    "created_after",
	"issues":   "updated_after",
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise gitlab credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  int64(i),
			"iid":                 int64(i),
			"project_id":          int64(100),
			"name":                fmt.Sprintf("%s-fixture-%d", strings.TrimSuffix(stream, "s"), i),
			"username":            fmt.Sprintf("user_fixture_%d", i),
			"title":               fmt.Sprintf("Fixture %d", i),
			"path":                fmt.Sprintf("%s-%d", endpoint.resource, i),
			"path_with_namespace": fmt.Sprintf("acme/%s-%d", endpoint.resource, i),
			"full_path":           fmt.Sprintf("acme/%s-%d", endpoint.resource, i),
			"full_name":           fmt.Sprintf("Acme / %s %d", endpoint.resource, i),
			"description":         "fixture record",
			"default_branch":      "main",
			"visibility":          "private",
			"state":               "active",
			"web_url":             fmt.Sprintf("https://gitlab.com/%s/%d", endpoint.resource, i),
			"created_at":          fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updated_at":          fmt.Sprintf("2026-01-0%dT01:00:00Z", i),
			"last_activity_at":    fmt.Sprintf("2026-01-0%dT02:00:00Z", i),
			"closed_at":           nil,
			"archived":            false,
			"bot":                 false,
			"is_admin":            false,
			"star_count":          int64(i),
			"forks_count":         int64(0),
			"open_issues_count":   int64(i),
			"parent_id":           nil,
			"upvotes":             int64(0),
			"downvotes":           int64(0),
			"user_notes_count":    int64(i),
			"author":              map[string]any{"id": int64(900 + i)},
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

// Write is unsupported: the GitLab connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := gitlabBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := gitlabSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("gitlab connector requires secret credentials.access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: gitlabUserAgent,
	}, nil
}

// gitlabSecret resolves the access token from the configured secret fields. The
// access_token (personal access token or OAuth access token) is the bearer
// credential; the OAuth client_id/client_secret/refresh_token fields exist in
// the catalog for a refresh flow not required by this declarative read path.
func gitlabSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v := strings.TrimSpace(cfg.Secrets["credentials.access_token"]); v != "" {
		return v
	}
	// Tolerate a flattened secret key for convenience.
	return strings.TrimSpace(cfg.Secrets["access_token"])
}

// gitlabBaseURL resolves and validates the base URL. The default is
// gitlab.com/api/v4. An api_url config (bare host or full URL, per the GitLab
// catalog spec) or an explicit base_url override is normalized to an absolute
// https/http URL ending in /api/v4 to bound SSRF risk.
func gitlabBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		return validateBaseURL(base)
	}
	apiURL := strings.TrimSpace(cfg.Config["api_url"])
	if apiURL == "" {
		return gitlabDefaultBaseURL, nil
	}
	// api_url may be a bare host ("gitlab.com") or a full URL; normalize to an
	// absolute https URL and append the /api/v4 path segment.
	if !strings.Contains(apiURL, "://") {
		apiURL = "https://" + apiURL
	}
	normalized, err := validateBaseURL(apiURL)
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(normalized, "/api/v4") {
		normalized = normalized + "/api/v4"
	}
	return normalized, nil
}

// validateBaseURL enforces the SSRF guard: absolute http/https URL with a host.
func validateBaseURL(base string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("gitlab config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("gitlab config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("gitlab config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func gitlabPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return gitlabDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gitlab config page_size must be an integer: %w", err)
	}
	if value < 1 || value > gitlabMaxPageSize {
		return 0, fmt.Errorf("gitlab config page_size must be between 1 and %d", gitlabMaxPageSize)
	}
	return value, nil
}

func gitlabMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gitlab config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("gitlab config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
