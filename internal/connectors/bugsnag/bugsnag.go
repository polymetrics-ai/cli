// Package bugsnag implements the native pm Bugsnag connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// ApiKey "token" auth + root-array extraction + Link-header pagination) with
// Bugsnag-specific stream definitions and endpoints.
//
// Bugsnag resources are hierarchical (organizations -> projects -> errors /
// events / releases). Child streams resolve their required parent id from
// config (organization_id / project_id) and otherwise auto-discover the parent
// via the Data Access API.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package bugsnag

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
	bugsnagDefaultBaseURL  = "https://api.bugsnag.com"
	bugsnagDefaultPageSize = 100
	bugsnagMaxPageSize     = 100
	bugsnagUserAgent       = "polymetrics-go-cli"
	// bugsnagAPIVersion is the mandatory X-Version header value for the v2
	// Data Access API.
	bugsnagAPIVersion = "2"
)

func init() {
	connectors.RegisterFactory("bugsnag", New)
}

// New returns the Bugsnag connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Bugsnag connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk
	// Requester. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "bugsnag" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "bugsnag",
		DisplayName:     "Bugsnag",
		IntegrationType: "api",
		Description:     "Reads Bugsnag organizations, projects, collaborators, errors, events, and releases through the Bugsnag Data Access API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Bugsnag. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := bugsnagBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(bugsnagSecret(cfg)) == "" {
		return errors.New("bugsnag connector requires secret auth_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing the user's organizations confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "user/organizations", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check bugsnag: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: bugsnagStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Bugsnag stream starts with
// an empty incremental cursor (full sync). Bugsnag's free tier only guarantees
// full-refresh, so this is a conservative default.
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
		stream = "organizations"
	}
	endpoint, ok := bugsnagStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("bugsnag stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := bugsnagPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := bugsnagMaxPages(req.Config)
	if err != nil {
		return err
	}

	parentIDs, err := c.resolveParents(ctx, r, endpoint.scope, req.Config, pageSize, maxPages)
	if err != nil {
		return err
	}
	for _, parentID := range parentIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := endpoint.pathFor(parentID)
		if err := c.harvest(ctx, r, path, endpoint.mapRecord, pageSize, maxPages, emit); err != nil {
			return err
		}
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Bugsnag is a read-only
// (extract) source; reverse ETL is not supported, so writes are rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// resolveParents returns the list of parent ids whose endpoints must be read to
// fully cover a stream. For scopeRoot it returns a single empty id (the endpoint
// is fixed). For scoped streams it prefers the configured id(s) and otherwise
// auto-discovers them from the parent endpoint(s).
func (c Connector) resolveParents(ctx context.Context, r *connsdk.Requester, s scope, cfg connectors.RuntimeConfig, pageSize, maxPages int) ([]string, error) {
	switch s {
	case scopeRoot:
		return []string{""}, nil
	case scopeOrganization:
		if ids := configIDs(cfg, "organization_id"); len(ids) > 0 {
			return ids, nil
		}
		return c.discoverOrganizationIDs(ctx, r, pageSize, maxPages)
	case scopeProject:
		if ids := configIDs(cfg, "project_id"); len(ids) > 0 {
			return ids, nil
		}
		return c.discoverProjectIDs(ctx, r, cfg, pageSize, maxPages)
	default:
		return nil, fmt.Errorf("bugsnag: unknown stream scope %d", s)
	}
}

// discoverOrganizationIDs lists the user's organizations and returns their ids.
func (c Connector) discoverOrganizationIDs(ctx context.Context, r *connsdk.Requester, pageSize, maxPages int) ([]string, error) {
	var ids []string
	err := c.harvest(ctx, r, "user/organizations", func(item map[string]any) connectors.Record {
		return connectors.Record(item)
	}, pageSize, maxPages, func(rec connectors.Record) error {
		if id := stringField(rec, "id"); id != "" {
			ids = append(ids, id)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("bugsnag: discover organizations: %w", err)
	}
	if len(ids) == 0 {
		return nil, errors.New("bugsnag: no organizations found for auth_token; set config organization_id")
	}
	return ids, nil
}

// discoverProjectIDs lists projects across the user's organizations and returns
// their ids. It is used when a project-scoped stream is read without an explicit
// project_id config.
func (c Connector) discoverProjectIDs(ctx context.Context, r *connsdk.Requester, cfg connectors.RuntimeConfig, pageSize, maxPages int) ([]string, error) {
	orgIDs := configIDs(cfg, "organization_id")
	if len(orgIDs) == 0 {
		discovered, err := c.discoverOrganizationIDs(ctx, r, pageSize, maxPages)
		if err != nil {
			return nil, err
		}
		orgIDs = discovered
	}
	var ids []string
	for _, org := range orgIDs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		path := "organizations/" + org + "/projects"
		err := c.harvest(ctx, r, path, func(item map[string]any) connectors.Record {
			return connectors.Record(item)
		}, pageSize, maxPages, func(rec connectors.Record) error {
			if id := stringField(rec, "id"); id != "" {
				ids = append(ids, id)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("bugsnag: discover projects for organization %s: %w", org, err)
		}
	}
	if len(ids) == 0 {
		return nil, errors.New("bugsnag: no projects found; set config project_id")
	}
	return ids, nil
}

// harvest drives Bugsnag's Link-header (RFC 5988 rel="next") pagination over a
// root-array endpoint, mapping and emitting each record. It is built on
// connsdk.Harvest + connsdk.LinkHeaderPaginator + RecordsAt(root).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, mapRecord func(map[string]any) connectors.Record, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	paginator := &connsdk.LinkHeaderPaginator{}
	// connsdk.Harvest's emit callback uses connsdk.Record (map[string]any);
	// bridge it to the stream's mapper + connectors.Record emitter.
	return connsdk.Harvest(ctx, r, http.MethodGet, path, base, paginator, "", maxPages, func(rec connsdk.Record) error {
		return emit(mapRecord(map[string]any(rec)))
	})
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise bugsnag credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                      fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":                    fmt.Sprintf("Fixture %s %d", stream, i),
			"slug":                    fmt.Sprintf("fixture-%d", i),
			"organization_id":         "org_fixture_1",
			"project_id":              "proj_fixture_1",
			"error_id":                "err_fixture_1",
			"release_group_id":        "rg_fixture_1",
			"api_key":                 "fixturekey",
			"email":                   fmt.Sprintf("fixture+%d@example.com", i),
			"error_class":             "RuntimeError",
			"message":                 "fixture error",
			"context":                 "GET /fixture",
			"severity":                "error",
			"original_severity":       "error",
			"status":                  "open",
			"type":                    "rails",
			"language":                "ruby",
			"is_admin":                false,
			"unhandled":               true,
			"is_full_report":          true,
			"two_factor_enabled":      false,
			"pending_invitation":      false,
			"auto_upgrade":            false,
			"open_error_count":        int64(i),
			"for_review_error_count":  int64(0),
			"collaborators_count":     int64(1),
			"events":                  int64(10 * i),
			"comment_count":           int64(0),
			"errors_seen_count":       int64(5 * i),
			"errors_introduced_count": int64(i),
			"app_version":             fmt.Sprintf("1.0.%d", i),
			"app_version_code":        strconv.Itoa(i),
			"app_bundle_version":      fmt.Sprintf("1.0.%d", i),
			"build_label":             fmt.Sprintf("build-%d", i),
			"release_stage":           "production",
			"release_source":          "api",
			"first_seen":              "2026-01-01T00:00:00.000Z",
			"last_seen":               "2026-01-02T00:00:00.000Z",
			"received_at":             "2026-01-02T00:00:00.000Z",
			"created_at":              "2026-01-01T00:00:00.000Z",
			"updated_at":              "2026-01-02T00:00:00.000Z",
			"last_request_at":         "2026-01-02T00:00:00.000Z",
			"release_time":            "2026-01-01T00:00:00.000Z",
			"connector":               "bugsnag",
			"fixture":                 true,
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

// requester builds a connsdk.Requester wired with Bugsnag "token" ApiKey auth,
// the resolved base URL, and the mandatory X-Version header. The secret only
// ever flows into connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := bugsnagBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := bugsnagSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("bugsnag connector requires secret auth_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, "token "),
		UserAgent: bugsnagUserAgent,
		DefaultHeaders: map[string]string{
			"X-Version": bugsnagAPIVersion,
		},
	}, nil
}

func bugsnagSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["auth_token"]
}

// bugsnagBaseURL resolves and validates the base URL. The default is
// api.bugsnag.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func bugsnagBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return bugsnagDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("bugsnag config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("bugsnag config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("bugsnag config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func bugsnagPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return bugsnagDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bugsnag config page_size must be an integer: %w", err)
	}
	if value < 1 || value > bugsnagMaxPageSize {
		return 0, fmt.Errorf("bugsnag config page_size must be between 1 and %d", bugsnagMaxPageSize)
	}
	return value, nil
}

func bugsnagMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bugsnag config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("bugsnag config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// configIDs returns the comma-separated id list under key (e.g. project_id),
// trimmed and de-empties. Supports a single id or a comma-separated list so a
// caller can scope multiple parents.
func configIDs(cfg connectors.RuntimeConfig, key string) []string {
	if cfg.Config == nil {
		return nil
	}
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if id := strings.TrimSpace(part); id != "" {
			out = append(out, id)
		}
	}
	return out
}

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
