// Package clockify implements the native pm Clockify source connector. It is a
// thin declarative-HTTP package that composes the connsdk toolkit (Requester +
// X-Api-Key auth + page-number pagination over top-level JSON arrays) with
// Clockify-specific stream definitions and endpoints. It follows the stripe
// reference connector's shape.
//
// Like stripe and github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package clockify

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
	clockifyDefaultBaseURL  = "https://api.clockify.me/api"
	clockifyDefaultPageSize = 50
	clockifyMaxPageSize     = 200
	clockifyUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("clockify", New)
}

// New returns the Clockify connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Clockify connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "clockify" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "clockify",
		DisplayName:     "Clockify",
		IntegrationType: "api",
		Description:     "Reads Clockify workspaces, clients, projects, tags, and users through the Clockify REST API v1.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Clockify. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := clockifyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(clockifySecret(cfg)) == "" {
		return errors.New("clockify connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the top-level workspaces list confirms auth and
	// connectivity without needing a workspace_id and without mutating anything.
	q := url.Values{"page": []string{"1"}, "page-size": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "v1/workspaces", q, nil, nil); err != nil {
		return fmt.Errorf("check clockify: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: clockifyStreams()}, nil
}

// Write is unsupported: Clockify is a read-only source connector for pm.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "workspaces"
	}
	endpoint, ok := clockifyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("clockify stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := clockifyPath(req.Config, endpoint)
	if err != nil {
		return err
	}
	pageSize, err := clockifyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := clockifyMaxPages(req.Config)
	if err != nil {
		return err
	}

	// Clockify lists return a top-level JSON array (no envelope) and use
	// 1-indexed page-number pagination with a `pageSize` param. PageNumberPaginator
	// stops when a page returns fewer than pageSize records, which matches
	// Clockify's behavior.
	paginator := &connsdk.PageNumberPaginator{
		PageParam: "page",
		SizeParam: "page-size",
		StartPage: 1,
		PageSize:  pageSize,
	}
	mapped := func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(map[string]any(rec)))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, path, nil, paginator, "", maxPages, mapped); err != nil {
		return fmt.Errorf("read clockify %s: %w", stream, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise clockify credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                      fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":                    fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"workspaceId":             "ws_fixture_1",
			"email":                   fmt.Sprintf("fixture+%d@example.com", i),
			"address":                 "123 Fixture St",
			"note":                    "fixture",
			"archived":                false,
			"clientId":                "cl_fixture_1",
			"clientName":              "Fixture Client",
			"color":                   "#FF0000",
			"billable":                true,
			"public":                  false,
			"duration":                "PT0S",
			"activeWorkspace":         "ws_fixture_1",
			"defaultWorkspace":        "ws_fixture_1",
			"status":                  "ACTIVE",
			"profilePicture":          "",
			"imageUrl":                "",
			"featureSubscriptionType": "FREE",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-Api-Key auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := clockifyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := clockifySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("clockify connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-Api-Key", secret, ""),
		UserAgent: clockifyUserAgent,
	}, nil
}

// clockifyPath builds the request path for a stream. Workspace-scoped streams
// require a workspace_id from config.
func clockifyPath(cfg connectors.RuntimeConfig, endpoint streamEndpoint) (string, error) {
	if !endpoint.workspaceScoped {
		return "v1/workspaces", nil
	}
	workspaceID := strings.TrimSpace(cfg.Config["workspace_id"])
	if workspaceID == "" {
		return "", errors.New("clockify connector requires config workspace_id for this stream")
	}
	return "v1/workspaces/" + url.PathEscape(workspaceID) + "/" + endpoint.resource, nil
}

func clockifySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// clockifyBaseURL resolves and validates the base URL. The default is
// api.clockify.me/api; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func clockifyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		// Honor the catalog's api_url field as an alternate override name.
		base = strings.TrimSpace(cfg.Config["api_url"])
	}
	if base == "" {
		return clockifyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("clockify config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("clockify config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("clockify config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func clockifyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return clockifyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("clockify config page_size must be an integer: %w", err)
	}
	if value < 1 || value > clockifyMaxPageSize {
		return 0, fmt.Errorf("clockify config page_size must be between 1 and %d", clockifyMaxPageSize)
	}
	return value, nil
}

func clockifyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("clockify config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("clockify config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
