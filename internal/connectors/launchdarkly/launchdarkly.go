// Package launchdarkly implements the native pm LaunchDarkly connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, modeled on
// the stripe reference: a thin package that composes connsdk (Requester +
// APIKeyHeader auth + RecordsAt "items" extraction + offset/limit pagination)
// with LaunchDarkly-specific stream definitions and endpoints.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// LaunchDarkly's REST API authenticates with the raw access token placed
// directly in the Authorization header (no "Bearer" prefix). List endpoints wrap
// rows under "items" and paginate with offset/limit query parameters.
package launchdarkly

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
	launchdarklyDefaultBaseURL  = "https://app.launchdarkly.com/api/v2"
	launchdarklyDefaultPageSize = 20
	launchdarklyMaxPageSize     = 100
	launchdarklyUserAgent       = "polymetrics-go-cli"
	// launchdarklyFixtureDate is the deterministic millisecond timestamp used by
	// fixture-mode records (2026-01-01T00:00:00Z in unix milliseconds).
	launchdarklyFixtureDate int64 = 1767225600000
)

func init() {
	connectors.RegisterFactory("launchdarkly", New)
}

// New returns the LaunchDarkly connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm LaunchDarkly connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "launchdarkly" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "launchdarkly",
		DisplayName:     "LaunchDarkly",
		IntegrationType: "api",
		Description:     "Reads LaunchDarkly projects, members, audit log entries, feature flags, and environments through the LaunchDarkly REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to LaunchDarkly.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := launchdarklyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(launchdarklySecret(cfg)) == "" {
		return errors.New("launchdarkly connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the projects list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "projects", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check launchdarkly: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: launchdarklyStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "projects"
	}
	endpoint, ok := launchdarklyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("launchdarkly stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, req, emit)
	}

	resource, err := resolveResource(endpoint, req.Config)
	if err != nil {
		return err
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := launchdarklyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := launchdarklyMaxPages(req.Config)
	if err != nil {
		return err
	}

	paginator := &connsdk.OffsetPaginator{
		LimitParam:  "limit",
		OffsetParam: "offset",
		PageSize:    pageSize,
	}
	return connsdk.Harvest(ctx, r, http.MethodGet, resource, nil, paginator, "items", maxPages, func(rec connsdk.Record) error {
		return emit(endpoint.mapRecord(rec))
	})
}

// Write is unsupported: LaunchDarkly is exposed as a read-only source here.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise launchdarkly credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	resource := strings.ReplaceAll(endpoint.resource, "/{project_key}", "")
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"_id":              fmt.Sprintf("%s_fixture_%d", endpoint.primaryKey, i),
			"key":              fmt.Sprintf("%s-fixture-%d", resource, i),
			"name":             fmt.Sprintf("Fixture %d", i),
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"firstName":        "Fixture",
			"lastName":         strconv.Itoa(i),
			"role":             "reader",
			"_pendingInvite":   false,
			"date":             launchdarklyFixtureDate + int64(i),
			"kind":             "fixture",
			"description":      fmt.Sprintf("fixture %s record %d", resource, i),
			"shortDescription": "fixture",
			"creationDate":     launchdarklyFixtureDate + int64(i),
			"temporary":        false,
			"color":            "417505",
			"defaultTtl":       0,
			"tags":             []any{"fixture"},
		}
		record := endpoint.mapRecord(item)
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with LaunchDarkly's raw-token
// Authorization header (no Bearer prefix) and the resolved base URL. The secret
// only ever flows into connsdk.APIKeyHeader; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := launchdarklyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := launchdarklySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("launchdarkly connector requires secret access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, ""),
		UserAgent: launchdarklyUserAgent,
	}, nil
}

// resolveResource substitutes the project_key config into project-scoped paths.
func resolveResource(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.needsProject {
		return endpoint.resource, nil
	}
	projectKey := strings.TrimSpace(cfg.Config["project_key"])
	if projectKey == "" {
		return "", fmt.Errorf("launchdarkly stream requires config project_key for path %q", endpoint.resource)
	}
	return strings.ReplaceAll(endpoint.resource, "{project_key}", url.PathEscape(projectKey)), nil
}

func launchdarklySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}

// launchdarklyBaseURL resolves and validates the base URL. The default is
// app.launchdarkly.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func launchdarklyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return launchdarklyDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("launchdarkly config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("launchdarkly config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("launchdarkly config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func launchdarklyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return launchdarklyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("launchdarkly config page_size must be an integer: %w", err)
	}
	if value < 1 || value > launchdarklyMaxPageSize {
		return 0, fmt.Errorf("launchdarkly config page_size must be between 1 and %d", launchdarklyMaxPageSize)
	}
	return value, nil
}

func launchdarklyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("launchdarkly config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("launchdarkly config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
