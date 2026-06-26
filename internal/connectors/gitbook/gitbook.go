// Package gitbook implements the native pm GitBook connector. It follows the
// stripe reference shape: a thin package that composes the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction + cursor pagination) with
// GitBook-specific stream definitions and endpoints.
//
// GitBook's HTTP API is read-only for our purposes (documentation content,
// users, organizations, members), so this connector exposes no write actions.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package gitbook

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
	gitbookDefaultBaseURL  = "https://api.gitbook.com/v1"
	gitbookDefaultPageSize = 50
	gitbookMaxPageSize     = 100
	gitbookUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("gitbook", New)
}

// New returns the GitBook connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm GitBook connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "gitbook" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "gitbook",
		DisplayName:     "GitBook",
		IntegrationType: "api",
		Description:     "Reads GitBook users, organizations, organization members, and space content (pages) through the GitBook REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to GitBook. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gitbookBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(gitbookSecret(cfg)) == "" {
		return errors.New("gitbook connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A read of the authenticated user confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "user", nil, nil, nil); err != nil {
		return fmt.Errorf("check gitbook: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gitbookStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "organizations"
	}
	endpoint, ok := gitbookStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("gitbook stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	resource, err := c.resolveResource(endpoint.resource, req.Config)
	if err != nil {
		return err
	}
	if !endpoint.list {
		return c.readSingle(ctx, r, resource, endpoint, emit)
	}
	pageSize, err := gitbookPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gitbookMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, resource, endpoint, pageSize, maxPages, emit)
}

// Write is unsupported: GitBook content is not a reverse-ETL target in this
// connector, so it is read-only (Metadata reports Write=false).
func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// readSingle reads a non-paginated single-object endpoint (e.g. /user) and emits
// exactly one record.
func (c Connector) readSingle(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	resp, err := r.Do(ctx, http.MethodGet, resource, nil, nil)
	if err != nil {
		return fmt.Errorf("read gitbook %s: %w", resource, err)
	}
	records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsPath)
	if err != nil {
		return fmt.Errorf("decode gitbook %s: %w", resource, err)
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

// harvest drives GitBook cursor pagination. List endpoints return
// {"items":[...],"next":{"page":"<cursor>"}}; the next page is requested with
// ?page=<cursor>. We build on connsdk.Harvest + connsdk.CursorPaginator, which
// reads the next token at next.page and re-issues the request with it.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	paginator := &connsdk.CursorPaginator{
		CursorParam: "page",
		TokenPath:   "next.page",
	}
	wrapped := func(item connsdk.Record) error {
		return emit(endpoint.mapRecord(map[string]any(item)))
	}
	if err := connsdk.Harvest(ctx, r, http.MethodGet, resource, base, paginator, endpoint.recordsPath, maxPages, wrapped); err != nil {
		return fmt.Errorf("read gitbook %s: %w", resource, err)
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise gitbook credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":          fmt.Sprintf("%s_fixture_%d", stream, i),
			"displayName": fmt.Sprintf("Fixture %d", i),
			"title":       fmt.Sprintf("Fixture %s %d", stream, i),
			"email":       fmt.Sprintf("fixture+%d@example.com", i),
			"type":        "fixture",
			"kind":        "document",
			"slug":        fmt.Sprintf("fixture-%d", i),
			"path":        fmt.Sprintf("fixture/%d", i),
			"role":        "member",
			"createdAt":   "2026-01-01T00:00:00Z",
			"connector":   "gitbook",
			"fixture":     true,
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
	base, err := gitbookBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := gitbookSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("gitbook connector requires secret access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: gitbookUserAgent,
	}, nil
}

// resolveResource substitutes the {org} and {space} tokens in an endpoint path
// with the configured organization_id / space_id. It returns an error if a
// required identifier is missing.
func (c Connector) resolveResource(resource string, cfg connectors.RuntimeConfig) (string, error) {
	if strings.Contains(resource, "{space}") {
		space := strings.TrimSpace(cfg.Config["space_id"])
		if space == "" {
			return "", errors.New("gitbook config space_id is required for this stream")
		}
		resource = strings.ReplaceAll(resource, "{space}", url.PathEscape(space))
	}
	if strings.Contains(resource, "{org}") {
		org := strings.TrimSpace(cfg.Config["organization_id"])
		if org == "" {
			return "", errors.New("gitbook config organization_id is required for this stream")
		}
		resource = strings.ReplaceAll(resource, "{org}", url.PathEscape(org))
	}
	return resource, nil
}

func gitbookSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}

// gitbookBaseURL resolves and validates the base URL. The default is
// api.gitbook.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func gitbookBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return gitbookDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("gitbook config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("gitbook config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("gitbook config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func gitbookPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return gitbookDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gitbook config page_size must be an integer: %w", err)
	}
	if value < 1 || value > gitbookMaxPageSize {
		return 0, fmt.Errorf("gitbook config page_size must be between 1 and %d", gitbookMaxPageSize)
	}
	return value, nil
}

func gitbookMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gitbook config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("gitbook config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
