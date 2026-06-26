// Package dbt implements the native pm dbt Cloud connector. It is a declarative
// HTTP per-system connector built on the connsdk toolkit, modeled on the stripe
// reference connector: a thin package that composes connsdk.Requester +
// APIKeyHeader ("Token <key>") auth + RecordsAt extraction with dbt Cloud
// Administrative API v2 account-scoped stream definitions and offset/limit
// pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The connector reads the dbt Cloud Administrative API v2
// (https://cloud.getdbt.com/api/v2). Auth is the dbt service-token convention:
// "Authorization: Token <api_key_2>". List endpoints are account-scoped under
// /accounts/{account_id}/<resource>/ and return {data:[...], extra:{pagination:
// {count,total_count}, filters:{limit,offset}}}.
package dbt

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
	dbtDefaultBaseURL  = "https://cloud.getdbt.com/api/v2"
	dbtDefaultPageSize = 100
	dbtMaxPageSize     = 100
	dbtUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("dbt", New)
}

// New returns the dbt connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm dbt Cloud connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "dbt" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "dbt",
		DisplayName:     "dbt Cloud",
		IntegrationType: "api",
		Description:     "Reads dbt Cloud projects, runs, repositories, users, and environments through the dbt Cloud Administrative API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to dbt Cloud. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := dbtBaseURL(cfg); err != nil {
		return err
	}
	accountID, err := dbtAccountID(cfg)
	if err != nil {
		return err
	}
	if strings.TrimSpace(dbtSecret(cfg)) == "" {
		return errors.New("dbt connector requires secret api_key_2")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the projects list confirms auth and connectivity
	// without mutating anything.
	path := fmt.Sprintf("accounts/%s/projects/", accountID)
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check dbt: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: dbtStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "projects"
	}
	endpoint, ok := dbtStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("dbt stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	accountID, err := dbtAccountID(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := dbtPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := dbtMaxPages(req.Config)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("accounts/%s/%s", accountID, endpoint.resource)
	return c.harvest(ctx, r, path, endpoint.mapRecord, pageSize, maxPages, emit)
}

// harvest drives dbt Cloud's offset/limit pagination. List responses are shaped
// {data:[...], extra:{pagination:{count,total_count}, filters:{limit,offset}}}.
// connsdk has no body-shape paginator for this exact total_count contract, so the
// loop lives here, built on connsdk.Requester + connsdk.RecordsAt + StringAt. It
// stops when a page is shorter than the page size, when the accumulated count
// reaches total_count, or when maxPages is reached.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, mapRecord func(map[string]any) connectors.Record, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	emitted := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read dbt %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode dbt %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
			emitted++
		}

		// Stop on a short page: fewer records than the requested limit means the
		// last page has been reached.
		if len(records) < pageSize {
			return nil
		}
		// Defensive stop: if the API reports a total_count we have already met or
		// exceeded, do not request further pages.
		if total := parseTotalCount(resp.Body); total > 0 && emitted >= total {
			return nil
		}
		// Guard against an endless loop when an empty page still claims to be full.
		if len(records) == 0 {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// parseTotalCount reads extra.pagination.total_count from a dbt list response.
// Returns 0 when absent or unparsable (which disables the total-count stop and
// lets the short-page check govern termination).
func parseTotalCount(body []byte) int {
	raw, err := connsdk.StringAt(body, "extra.pagination.total_count")
	if err != nil || raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise dbt credential-free (mirrors stripe's fixture
// intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                       i,
			"account_id":               1,
			"project_id":               1,
			"job_definition_id":        10,
			"environment_id":           5,
			"connection_id":            2,
			"repository_id":            3,
			"name":                     fmt.Sprintf("%s_fixture_%d", strings.TrimSuffix(stream, "s"), i),
			"description":              "fixture record",
			"state":                    1,
			"status":                   10,
			"status_humanized":         "Success",
			"is_complete":              true,
			"is_error":                 false,
			"is_cancelled":             false,
			"dbt_project_subdirectory": "",
			"remote_url":               "git@github.com:acme/analytics.git",
			"remote_backend":           "github",
			"git_clone_strategy":       "deploy_key",
			"email":                    fmt.Sprintf("fixture+%d@example.com", i),
			"first_name":               "Fixture",
			"last_name":                strconv.Itoa(i),
			"fullname":                 fmt.Sprintf("Fixture %d", i),
			"is_active":                true,
			"dbt_version":              "1.7.0",
			"type":                     "deployment",
			"use_custom_branch":        false,
			"custom_branch":            "",
			"started_at":               "2026-01-01T00:00:00Z",
			"finished_at":              "2026-01-01T00:05:00Z",
			"created_at":               "2026-01-01T00:00:00Z",
			"updated_at":               "2026-01-01T00:05:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with dbt "Token <key>" auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := dbtBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := dbtSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("dbt connector requires secret api_key_2")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, "Token "),
		UserAgent: dbtUserAgent,
	}, nil
}

func dbtSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key_2"]
}

// dbtAccountID resolves and validates the required account_id config.
func dbtAccountID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New("dbt connector requires config account_id")
	}
	id := strings.TrimSpace(cfg.Config["account_id"])
	if id == "" {
		return "", errors.New("dbt connector requires config account_id")
	}
	if _, err := strconv.Atoi(id); err != nil {
		return "", fmt.Errorf("dbt config account_id must be an integer: %w", err)
	}
	return id, nil
}

// dbtBaseURL resolves and validates the base URL. The default is
// cloud.getdbt.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func dbtBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return dbtDefaultBaseURL, nil
	}
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return dbtDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("dbt config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("dbt config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("dbt config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func dbtPageSize(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return dbtDefaultPageSize, nil
	}
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return dbtDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dbt config page_size must be an integer: %w", err)
	}
	if value < 1 || value > dbtMaxPageSize {
		return 0, fmt.Errorf("dbt config page_size must be between 1 and %d", dbtMaxPageSize)
	}
	return value, nil
}

func dbtMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return 0, nil
	}
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("dbt config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("dbt config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. The dbt Cloud connector is
// read-only (the Administrative API write surface is operational, not reverse
// ETL), so writes are unsupported.
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
