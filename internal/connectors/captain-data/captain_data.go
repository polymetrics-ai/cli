// Package captaindata implements the native pm Captain Data connector. It is a
// declarative-HTTP per-system connector following the stripe template: a thin
// package that composes the connsdk toolkit (Requester + x-api-key header auth +
// RecordsAt extraction + cursor pagination) with Captain Data-specific stream
// definitions and endpoints.
//
// The Captain Data v3 API (https://api.captaindata.co/v3) authenticates with an
// x-api-key header and scopes every request to a project via the x-project-id
// header. Top-level list endpoints (workspace, workflows) return a JSON array at
// the root; the job_results endpoint returns {results:[...], paging:{next,
// have_next_page}} and is read with cursor pagination. The jobs and job_results
// streams are scoped by a parent uid supplied through config (workflow_uid /
// job_uid). The Captain Data source is read-only (full-refresh); it exposes no
// reverse-ETL writes, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package captaindata

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
	captainDataDefaultBaseURL = "https://api.captaindata.co/v3"
	captainDataUserAgent      = "polymetrics-go-cli"
	captainDataAPIKeyHeader   = "X-API-Key"
	captainDataProjectHeader  = "X-Project-Id"
	captainDataCursorParam    = "cursor"
)

func init() {
	connectors.RegisterFactory("captain-data", New)
}

// New returns the Captain Data connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Captain Data connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "captain-data" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "captain-data",
		DisplayName:     "Captain Data",
		IntegrationType: "api",
		Description:     "Reads Captain Data workspace, workflows, jobs, and job results through the Captain Data v3 REST API (read-only / full-refresh).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Captain Data.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := captainDataBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(captainDataSecret(cfg)) == "" {
		return errors.New("captain-data connector requires secret api_key")
	}
	if strings.TrimSpace(cfg.Config["project_uid"]) == "" {
		return errors.New("captain-data connector requires config project_uid")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the workspace endpoint confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "workspace", nil, nil, nil); err != nil {
		return fmt.Errorf("check captain-data: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: captainDataStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "workflows"
	}
	endpoint, ok := captainDataStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("captain-data stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	path, err := resolvePath(endpoint, req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := captainDataMaxPages(req.Config)
	if err != nil {
		return err
	}
	if endpoint.paginated {
		return c.harvest(ctx, r, path, endpoint, maxPages, emit)
	}
	return c.readSingle(ctx, r, path, endpoint, emit)
}

// readSingle reads a top-level (non-paginated) endpoint that returns a JSON array
// at recordsPath in a single request.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return fmt.Errorf("read captain-data %s: %w", path, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode captain-data %s: %w", path, err)
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

// harvest drives Captain Data's cursor pagination. Paginated endpoints return
// {results:[...], paging:{next, have_next_page}}; the next page is requested with
// cursor=<paging.next> until paging.have_next_page is false. The loop lives here,
// built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if cursor != "" {
			query.Set(captainDataCursorParam, cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read captain-data %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
		if err != nil {
			return fmt.Errorf("decode captain-data %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasNext, err := connsdk.StringAt(resp.Body, "paging.have_next_page")
		if err != nil {
			return fmt.Errorf("decode captain-data %s paging: %w", path, err)
		}
		if hasNext != "true" {
			return nil
		}
		next, err := connsdk.StringAt(resp.Body, "paging.next")
		if err != nil {
			return fmt.Errorf("decode captain-data %s cursor: %w", path, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise captain-data credential-free (mirrors the
// stripe fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"uid":          fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":         fmt.Sprintf("Fixture %s %d", stream, i),
			"plan":         "scale",
			"status":       "success",
			"workflow_uid": "workflows_fixture_1",
			"job_uid":      "jobs_fixture_1",
			"data":         map[string]any{"fixture": true, "index": i},
			"created_at":   "2026-01-01T00:00:00Z",
			"updated_at":   "2026-01-02T00:00:00Z",
			"ended_at":     "2026-01-01T01:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// resolvePath builds the request path for a stream, substituting the scope uid
// from config for scoped streams (jobs, job_results).
func resolvePath(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if endpoint.scopeParam == "" {
		return endpoint.resource, nil
	}
	scope := strings.TrimSpace(cfg.Config[endpoint.scopeParam])
	if scope == "" {
		return "", fmt.Errorf("captain-data stream requires config %s", endpoint.scopeParam)
	}
	return endpoint.pathFor(url.PathEscape(scope)), nil
}

// requester builds a connsdk.Requester wired with x-api-key auth, the x-project-id
// header, and the resolved base URL. The secret only ever flows into the
// authenticator and is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := captainDataBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := captainDataSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("captain-data connector requires secret api_key")
	}
	project := strings.TrimSpace(cfg.Config["project_uid"])
	if project == "" {
		return nil, errors.New("captain-data connector requires config project_uid")
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.APIKeyHeader(captainDataAPIKeyHeader, secret, ""),
		UserAgent:      captainDataUserAgent,
		DefaultHeaders: map[string]string{captainDataProjectHeader: project},
	}, nil
}

// Write satisfies the connectors.Connector interface. The Captain Data source is
// read-only (no approved reverse-ETL actions), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func captainDataSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// captainDataBaseURL resolves and validates the base URL. The default is
// api.captaindata.co/v3; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func captainDataBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return captainDataDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("captain-data config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("captain-data config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("captain-data config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func captainDataMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("captain-data config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("captain-data config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
