// Package instatus implements the native pm Instatus connector. It follows the
// stripe reference shape: a thin declarative-HTTP package that composes the
// connsdk toolkit (Requester + Bearer auth + RecordsAt extraction) with
// Instatus-specific stream definitions, endpoints, and page/per_page pagination.
//
// Instatus exposes a status-page REST API at https://api.instatus.com. The
// `pages` stream lists the workspace's status pages; components, incidents, and
// maintenances are parent-scoped under a specific page id (page_id config). The
// API only supports full-refresh syncs, so the connector is read-only with no
// incremental cursor.
//
// Like the other per-system connectors, it self-registers via RegisterFactory in
// init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package instatus

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
	instatusDefaultBaseURL  = "https://api.instatus.com"
	instatusDefaultPageSize = 50
	instatusMaxPageSize     = 100
	instatusUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("instatus", New)
}

// New returns the Instatus connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Instatus connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "instatus" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "instatus",
		DisplayName:     "Instatus",
		IntegrationType: "api",
		Description:     "Reads Instatus status pages, components, incidents, and maintenances through the Instatus REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Instatus. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := instatusBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(instatusSecret(cfg)) == "" {
		return errors.New("instatus connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the pages list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "v2/pages", url.Values{"page": []string{"1"}, "per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check instatus: %w", err)
	}
	return nil
}

// Write satisfies connectors.Connector. Instatus is read-only (full-refresh
// source only), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: instatusStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "pages"
	}
	endpoint, ok := instatusStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("instatus stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	path, err := instatusPath(endpoint, req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := instatusPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := instatusMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, path, endpoint, pageSize, maxPages, emit)
}

// harvest drives Instatus's page/per_page pagination. List endpoints return a
// top-level JSON array; the loop requests page+1 until a page shorter than
// per_page is returned (or maxPages is hit). connsdk has a PageNumberPaginator,
// but it is built around an object-with-records-path shape; Instatus arrays are
// extracted with RecordsAt at the root, so the small loop lives here.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read instatus %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode instatus %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// A page shorter than the requested size is the last page.
		if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise instatus credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"subdomain":    fmt.Sprintf("fixture-%d", i),
			"name":         fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"status":       "OPERATIONAL",
			"description":  "fixture record",
			"websiteUrl":   "https://example.com",
			"customDomain": "",
			"publicEmail":  fmt.Sprintf("fixture+%d@example.com", i),
			"language":     "en",
			"uniqueEmail":  fmt.Sprintf("cmp+%d@example.com", i),
			"showUptime":   true,
			"order":        i,
			"group":        nil,
			"started":      "2026-01-01T00:00:00Z",
			"resolved":     "2026-01-02T00:00:00Z",
			"start":        "2026-01-03T00:00:00Z",
			"duration":     60 * i,
			"autoStart":    true,
			"autoEnd":      true,
			"createdAt":    "2026-01-01T00:00:00Z",
			"updatedAt":    "2026-01-02T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := instatusBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := instatusSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("instatus connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: instatusUserAgent,
	}, nil
}

// instatusPath resolves the API path for a stream's endpoint. Parent-scoped
// streams (components/incidents/maintenances) require a page_id config value and
// build "/{version}/{page_id}/{resource}". Top-level streams use topLevelPath.
func instatusPath(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.scoped {
		return endpoint.topLevelPath, nil
	}
	pageID := strings.TrimSpace(cfg.Config["page_id"])
	if pageID == "" {
		return "", fmt.Errorf("instatus stream %q requires config page_id", endpoint.resource)
	}
	return fmt.Sprintf("%s/%s/%s", endpoint.version, url.PathEscape(pageID), endpoint.resource), nil
}

func instatusSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// instatusBaseURL resolves and validates the base URL. The default is
// api.instatus.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func instatusBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return instatusDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("instatus config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("instatus config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("instatus config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func instatusPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return instatusDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("instatus config page_size must be an integer: %w", err)
	}
	if value < 1 || value > instatusMaxPageSize {
		return 0, fmt.Errorf("instatus config page_size must be between 1 and %d", instatusMaxPageSize)
	}
	return value, nil
}

func instatusMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("instatus config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("instatus config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
