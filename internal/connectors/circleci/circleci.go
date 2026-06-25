// Package circleci implements the native pm CircleCI connector. It is a
// declarative-HTTP per-system connector that composes the connsdk toolkit
// (Requester + Circle-Token header auth + items[]/next_page_token pagination)
// with CircleCI-specific stream definitions and endpoints. It mirrors the stripe
// reference connector's shape.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
//
// The connector is read-only: CircleCI's v2 API write surface (trigger pipeline,
// cancel workflow, approve job) mutates CI state and is not a safe reverse-ETL
// target, so Write is unsupported and Capabilities.Write is false.
package circleci

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
	circleciDefaultBaseURL  = "https://circleci.com/api/v2"
	circleciDefaultPageSize = 100
	circleciUserAgent       = "polymetrics-go-cli"
	// circleciTokenHeader is CircleCI's documented personal API token header.
	circleciTokenHeader = "Circle-Token"
)

var (
	errMissingProjectSlug = errors.New("circleci stream requires config project_slug (e.g. gh/org/repo)")
	errMissingPipelineID  = errors.New("circleci stream requires config pipeline_id")
	errMissingWorkflowID  = errors.New("circleci stream requires config workflow_id")
)

func init() {
	connectors.RegisterFactory("circleci", New)
}

// New returns the CircleCI connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm CircleCI connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "circleci" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "circleci",
		DisplayName:     "CircleCI",
		IntegrationType: "api",
		Description:     "Reads CircleCI projects, pipelines, workflows, and jobs through the CircleCI v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to CircleCI. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := circleciBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(circleciSecret(cfg)) == "" {
		return errors.New("circleci connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// /me confirms the token authenticates without depending on project config.
	if err := r.DoJSON(ctx, http.MethodGet, "me", nil, nil, nil); err != nil {
		return fmt.Errorf("check circleci: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: circleciStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a CircleCI stream starts with
// an empty incremental cursor (full sync), which start_date config can raise at
// read time.
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
	endpoint, ok := circleciStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("circleci stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := endpoint.resolvePath(pathParamsFrom(req.Config))
	if err != nil {
		return err
	}
	maxPages, err := circleciMaxPages(req.Config)
	if err != nil {
		return err
	}

	if endpoint.shape == shapeObject {
		return c.readObject(ctx, r, path, endpoint, emit)
	}
	return c.harvest(ctx, r, path, endpoint, maxPages, emit)
}

// readObject reads a single-object endpoint (e.g. project metadata) and emits one
// mapped record.
func (c Connector) readObject(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read circleci %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, "")
	if err != nil {
		return fmt.Errorf("decode circleci %s: %w", path, err)
	}
	for _, item := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// harvest drives CircleCI's items[]/next_page_token pagination. List endpoints
// return {"items":[...],"next_page_token":"..."}; the next page is requested with
// page-token=<token>. The loop is built on connsdk.Requester + connsdk.RecordsAt
// + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	pageToken := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if pageToken != "" {
			query.Set("page-token", pageToken)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read circleci %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode circleci %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_page_token")
		if err != nil {
			return fmt.Errorf("decode circleci %s next_page_token: %w", path, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		pageToken = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise circleci credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := circleciStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", stream, i),
			"slug":              "gh/fixture-org/fixture-repo",
			"name":              fmt.Sprintf("fixture-%s-%d", stream, i),
			"organization_name": "fixture-org",
			"organization_slug": "gh/fixture-org",
			"organization_id":   "org_fixture",
			"vcs_url":           "https://github.com/fixture-org/fixture-repo",
			"vcs_info":          map[string]any{"default_branch": "main"},
			"number":            int64(i),
			"pipeline_number":   int64(i),
			"job_number":        int64(i),
			"project_slug":      "gh/fixture-org/fixture-repo",
			"pipeline_id":       "pipeline_fixture_1",
			"state":             "created",
			"status":            "success",
			"type":              "build",
			"created_at":        fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"updated_at":        fmt.Sprintf("2026-01-0%dT01:00:00Z", i),
			"started_at":        fmt.Sprintf("2026-01-0%dT00:05:00Z", i),
			"stopped_at":        fmt.Sprintf("2026-01-0%dT00:10:00Z", i),
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

// requester builds a connsdk.Requester wired with Circle-Token header auth and
// the resolved base URL. The secret only ever flows into the auth header; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := circleciBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := circleciSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("circleci connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader(circleciTokenHeader, secret, ""),
		UserAgent: circleciUserAgent,
	}, nil
}

// pathParamsFrom derives the endpoint path parameters from the runtime config.
// project_slug is the canonical CircleCI project identifier (e.g. gh/org/repo);
// org_id/project_id are accepted as alternative config but slug is preferred.
func pathParamsFrom(cfg connectors.RuntimeConfig) pathParams {
	return pathParams{
		projectSlug: strings.TrimSpace(cfg.Config["project_slug"]),
		pipelineID:  strings.TrimSpace(cfg.Config["pipeline_id"]),
		workflowID:  strings.TrimSpace(cfg.Config["workflow_id"]),
	}
}

func circleciSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// circleciBaseURL resolves and validates the base URL. The default is
// circleci.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func circleciBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return circleciDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("circleci config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("circleci config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("circleci config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func circleciMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("circleci config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("circleci config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: CircleCI's mutation endpoints affect live CI state and
// are not safe reverse-ETL targets.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
