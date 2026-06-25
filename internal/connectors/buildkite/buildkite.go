// Package buildkite implements the native pm Buildkite connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package that composes the connsdk toolkit (Requester +
// Bearer auth + Link-header pagination + root-array extraction) with
// Buildkite-specific stream definitions and endpoints.
//
// Buildkite's REST API (https://buildkite.com/docs/apis/rest-api) authenticates
// with a Bearer API access token, paginates with RFC 5988 Link headers
// (rel="next") plus page/per_page query params, and returns each list endpoint
// as a top-level JSON array.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package buildkite

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
	buildkiteDefaultBaseURL  = "https://api.buildkite.com/v2"
	buildkiteDefaultPageSize = 100
	buildkiteMaxPageSize     = 100
	buildkiteUserAgent       = "polymetrics-go-cli"
	// buildkiteFixtureCreated is the deterministic created_at used by fixture
	// records (2026-01-01T00:00:00Z).
	buildkiteFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("buildkite", New)
}

// New returns the Buildkite connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Buildkite connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "buildkite" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "buildkite",
		DisplayName:     "Buildkite",
		IntegrationType: "api",
		Description:     "Reads Buildkite organizations, pipelines, builds, and agents through the Buildkite REST API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Buildkite.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := buildkiteBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(buildkiteSecret(cfg)) == "" {
		return errors.New("buildkite connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the organizations list confirms auth and connectivity
	// without mutating anything and without needing an organization slug.
	if err := r.DoJSON(ctx, http.MethodGet, "organizations", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check buildkite: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: buildkiteStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Buildkite stream starts
// with an empty incremental cursor (full refresh), which the start_date config
// can raise at read time.
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
		stream = "pipelines"
	}
	endpoint, ok := buildkiteStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("buildkite stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := endpointPath(endpoint, req.Config)
	if err != nil {
		return err
	}
	pageSize, err := buildkitePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := buildkiteMaxPages(req.Config)
	if err != nil {
		return err
	}

	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	if createdGTE := incrementalLowerBound(req); createdGTE != "" {
		// Buildkite builds support created_from; for other streams the param is
		// ignored harmlessly by the API but we only attach it to builds.
		if stream == "builds" {
			base.Set("created_from", createdGTE)
		}
	}

	// Buildkite paginates with RFC 5988 Link headers (rel="next"); records are a
	// top-level array, so the records path is the root (""). connsdk.Record is a
	// map[string]any alias, which is exactly the input mapRecord expects.
	paginator := &connsdk.LinkHeaderPaginator{FirstQuery: base}
	mapAndEmit := func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, path, base, paginator, "", maxPages, mapAndEmit); err != nil {
		return fmt.Errorf("read buildkite %s: %w", stream, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise buildkite credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"graphql_id":       fmt.Sprintf("gql_%s_%d", endpoint.resource, i),
			"slug":             fmt.Sprintf("%s-%d", strings.TrimSuffix(stream, "s"), i),
			"name":             fmt.Sprintf("Fixture %d", i),
			"created_at":       buildkiteFixtureCreated,
			"state":            "passed",
			"number":           int64(i),
			"branch":           "main",
			"commit":           fmt.Sprintf("%040d", i),
			"connection_state": "connected",
			"hostname":         fmt.Sprintf("agent-%d.example.com", i),
			"default_branch":   "main",
			"visibility":       "private",
			"connector":        "buildkite",
			"fixture":          true,
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := buildkiteBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := buildkiteSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("buildkite connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: buildkiteUserAgent,
	}, nil
}

// endpointPath resolves the request path for a stream, injecting the
// organization slug for organization-scoped streams.
func endpointPath(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if endpoint.scope == scopeTopLevel {
		return endpoint.resource, nil
	}
	org, err := organizationSlug(cfg)
	if err != nil {
		return "", err
	}
	return "organizations/" + org + "/" + endpoint.resource, nil
}

// organizationSlug returns the configured organization slug, validating it is a
// safe path segment.
func organizationSlug(cfg connectors.RuntimeConfig) (string, error) {
	org := strings.TrimSpace(cfg.Config["organization"])
	if org == "" {
		return "", errors.New("buildkite connector requires config organization for this stream")
	}
	if strings.ContainsAny(org, "/?#") {
		return "", fmt.Errorf("buildkite config organization %q must be a bare slug", org)
	}
	return url.PathEscape(org), nil
}

// incrementalLowerBound returns the RFC3339 lower bound for created_from,
// derived from the incremental cursor (if any) or else the start_date config. An
// empty result means no lower bound (full refresh).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func buildkiteSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// buildkiteBaseURL resolves and validates the base URL. The default is
// api.buildkite.com/v2; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func buildkiteBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return buildkiteDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("buildkite config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("buildkite config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("buildkite config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func buildkitePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return buildkiteDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("buildkite config page_size must be an integer: %w", err)
	}
	if value < 1 || value > buildkiteMaxPageSize {
		return 0, fmt.Errorf("buildkite config page_size must be between 1 and %d", buildkiteMaxPageSize)
	}
	return value, nil
}

func buildkiteMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("buildkite config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("buildkite config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported; Buildkite is exposed read-only. It satisfies the
// connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
