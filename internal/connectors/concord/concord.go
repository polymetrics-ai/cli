// Package concord implements the native pm Concord connector. Concord is a
// contract lifecycle management platform; this is a declarative-HTTP per-system
// connector built on the connsdk toolkit (Requester + X-API-KEY auth +
// RecordsAt extraction + page-increment pagination), following the stripe
// reference connector's shape.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The Concord API is read-only here (full-refresh only upstream), so the
// connector exposes Check/Catalog/Read and sets Capabilities.Write=false.
package concord

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
	// concordDefaultBaseURL targets the production environment. The env config
	// ("uat" or "api") selects the host; base_url can override entirely (tests).
	concordDefaultEnv     = "api"
	concordBaseURLPattern = "https://%s.concordnow.com/api/rest/1"
	concordDefaultPage    = 100
	concordMaxPageSize    = 1000
	concordUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("concord", New)
}

// New returns the Concord connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Concord connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "concord" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "concord",
		DisplayName:     "Concord",
		IntegrationType: "api",
		Description:     "Reads Concord agreements, user organizations, folders, reports, and tags through the Concord REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Concord. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := concordBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(concordSecret(cfg)) == "" {
		return errors.New("concord connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// Listing the authenticated user's organizations confirms auth and
	// connectivity without mutating anything and without needing an org id.
	if err := r.DoJSON(ctx, http.MethodGet, "user/me/organizations", nil, nil, nil); err != nil {
		return fmt.Errorf("check concord: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: concordStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "user_organizations"
	}
	endpoint, ok := concordStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("concord stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resource := endpoint.resource
	if endpoint.orgScoped {
		orgID, err := concordOrgID(req.Config)
		if err != nil {
			return err
		}
		resource = strings.ReplaceAll(resource, "{org}", url.PathEscape(orgID))
	}
	pageSize, err := concordPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := concordMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, resource, endpoint.recordsPath, endpoint.mapRecord, pageSize, maxPages, emit)
}

// harvest drives Concord's page-increment pagination. Pages start at 0; the
// page size is sent as the "limit" request parameter and the page index as
// "page". A page returning fewer records than the requested size is the last
// one. connsdk has paginators but Concord's record path varies per stream
// (root array vs nested field), so a small in-package loop on
// connsdk.Requester + connsdk.RecordsAt is clearer here.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource, recordsPath string, mapRecord func(map[string]any) connectors.Record, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("limit", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read concord %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, recordsPath)
		if err != nil {
			return fmt.Errorf("decode concord %s page %d: %w", resource, page, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		// A short (or empty) page terminates pagination.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise concord credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"uid":            fmt.Sprintf("%s_fixture_%d", stream, i),
			"id":             int64(i),
			"name":           fmt.Sprintf("Fixture %s %d", stream, i),
			"title":          fmt.Sprintf("Fixture Agreement %d", i),
			"status":         "SIGNED",
			"stage":          "EXECUTED",
			"role":           "OWNER",
			"type":           "STANDARD",
			"color":          "#336699",
			"parentId":       int64(0),
			"organizationId": int64(42),
			"createdAt":      "2026-01-01T00:00:00Z",
			"updatedAt":      "2026-01-02T00:00:00Z",
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "concord"
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-API-KEY auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := concordBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := concordSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("concord connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-API-KEY", secret, ""),
		UserAgent: concordUserAgent,
	}, nil
}

func concordSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// concordOrgID resolves the organization id for org-scoped streams. It is
// required for agreements/folders/reports.
func concordOrgID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config != nil {
		if id := strings.TrimSpace(cfg.Config["organization_id"]); id != "" {
			return id, nil
		}
	}
	return "", errors.New("concord config organization_id is required for organization-scoped streams (agreements, folders, reports)")
}

// concordBaseURL resolves and validates the base URL. With no override it is
// derived from the env config ("uat" or "api", default "api"). Any base_url
// override must be an absolute http(s) URL with a host to bound SSRF risk.
func concordBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := ""
	if cfg.Config != nil {
		base = strings.TrimSpace(cfg.Config["base_url"])
	}
	if base == "" {
		env := concordDefaultEnv
		if cfg.Config != nil {
			if e := strings.TrimSpace(cfg.Config["env"]); e != "" {
				env = e
			}
		}
		if env != "uat" && env != "api" {
			return "", fmt.Errorf("concord config env must be \"uat\" or \"api\", got %q", env)
		}
		return fmt.Sprintf(concordBaseURLPattern, env), nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("concord config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("concord config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("concord config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func concordPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := ""
	if cfg.Config != nil {
		raw = strings.TrimSpace(cfg.Config["page_size"])
	}
	if raw == "" {
		return concordDefaultPage, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("concord config page_size must be an integer: %w", err)
	}
	if value < 1 || value > concordMaxPageSize {
		return 0, fmt.Errorf("concord config page_size must be between 1 and %d", concordMaxPageSize)
	}
	return value, nil
}

func concordMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := ""
	if cfg.Config != nil {
		raw = strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	}
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("concord config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("concord config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write satisfies the connectors.Connector interface. Concord is read-only in
// this connector, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
