// Package confluence implements the native pm Confluence connector. It is a
// declarative-HTTP per-system connector following the stripe template: a thin
// package composing the connsdk toolkit (Requester + Basic auth + RecordsAt
// extraction) with Confluence-specific stream definitions, endpoints, and the
// Confluence Cloud REST API v2 cursor pagination style.
//
// Like stripe and github it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// Auth is HTTP Basic with the Atlassian account email as the username and the
// API token as the password (per the Confluence Cloud REST API). The token is
// only ever passed into connsdk.Basic and is never logged.
package confluence

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
	// confluenceAPIPath is the v2 API path appended to the resolved site host.
	confluenceAPIPath         = "/wiki/api/v2"
	confluenceDefaultPageSize = 25
	confluenceMaxPageSize     = 250
	confluenceUserAgent       = "polymetrics-go-cli"
	// confluenceFixtureCreated is the deterministic createdAt timestamp used by
	// the fixture-mode records.
	confluenceFixtureCreated = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("confluence", New)
}

// New returns the Confluence connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Confluence connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "confluence" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "confluence",
		DisplayName:     "Confluence",
		IntegrationType: "api",
		Description:     "Reads Confluence Cloud spaces, pages, blog posts, labels, and attachments through the Confluence Cloud REST API v2.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Confluence.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := confluenceBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(confluenceEmail(cfg)) == "" {
		return errors.New("confluence connector requires config email")
	}
	if strings.TrimSpace(confluenceSecret(cfg)) == "" {
		return errors.New("confluence connector requires secret api_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the spaces list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "spaces", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check confluence: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: confluenceStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "spaces"
	}
	endpoint, ok := confluenceStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("confluence stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := confluencePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := confluenceMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// Write is unsupported: the Confluence connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Confluence v2 cursor pagination. List responses are shaped
// {results:[...], _links:{next:"/wiki/api/v2/<res>?cursor=<token>&limit=..."}}.
// The opaque cursor token is extracted from _links.next and supplied as the
// `cursor` query param on the next request. There is no body-token paginator in
// connsdk that parses a URL out of _links.next, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read confluence %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "results")
		if err != nil {
			return fmt.Errorf("decode confluence %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "_links.next")
		if err != nil {
			return fmt.Errorf("decode confluence %s next link: %w", endpoint.resource, err)
		}
		cursor = nextCursor(next)
		if cursor == "" {
			return nil
		}
	}
	return nil
}

// nextCursor extracts the opaque `cursor` query parameter from a Confluence
// _links.next value, which is a relative path like
// "/wiki/api/v2/spaces?cursor=<token>&limit=25". Returning just the token (not
// the path) keeps requests bounded to the configured host (SSRF-safe).
func nextCursor(next string) string {
	next = strings.TrimSpace(next)
	if next == "" {
		return ""
	}
	parsed, err := url.Parse(next)
	if err != nil {
		return ""
	}
	return parsed.Query().Get("cursor")
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise confluence credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":         fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"key":        fmt.Sprintf("FIX%d", i),
			"name":       fmt.Sprintf("Fixture %s %d", stream, i),
			"title":      fmt.Sprintf("Fixture %s %d", stream, i),
			"type":       "global",
			"status":     "current",
			"spaceId":    "spaces_fixture_1",
			"pageId":     "pages_fixture_1",
			"parentId":   "",
			"authorId":   "author_fixture_1",
			"homepageId": "pages_fixture_1",
			"mediaType":  "text/plain",
			"fileSize":   int64(1024 * i),
			"prefix":     "global",
			"createdAt":  confluenceFixtureCreated,
			"version":    map[string]any{"number": int64(i)},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Basic auth (email:api_token)
// and the resolved base URL. The secret only ever flows into connsdk.Basic; it
// is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := confluenceBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	email := strings.TrimSpace(confluenceEmail(cfg))
	if email == "" {
		return nil, errors.New("confluence connector requires config email")
	}
	secret := confluenceSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("confluence connector requires secret api_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(email, secret),
		UserAgent: confluenceUserAgent,
	}, nil
}

func confluenceSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_token"]
}

func confluenceEmail(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	return cfg.Config["email"]
}

// confluenceBaseURL resolves and validates the base URL. When base_url is set
// it is used verbatim (its path already includes the API prefix in tests/local
// servers); otherwise the URL is derived from domain_name as
// https://<domain_name>/wiki/api/v2. Any override must be an absolute https (or
// http for local test servers) URL with a host to bound SSRF risk.
func confluenceBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("confluence config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("confluence config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("confluence config base_url must include a host")
		}
		trimmed := strings.TrimRight(base, "/")
		// If the override has no API path, append the v2 prefix; if it already
		// points at the API root (or a test server root), use it as-is.
		if parsed.Path == "" || parsed.Path == "/" {
			return trimmed + confluenceAPIPath, nil
		}
		return trimmed, nil
	}

	domain := strings.TrimSpace(cfg.Config["domain_name"])
	if domain == "" {
		return "", errors.New("confluence connector requires config domain_name or base_url")
	}
	// domain_name is a bare host (e.g. mysite.atlassian.net); strip any scheme
	// the user mistakenly included and validate the result.
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.Trim(domain, "/")
	if domain == "" || strings.ContainsAny(domain, " /") {
		return "", fmt.Errorf("confluence config domain_name must be a bare host, got %q", cfg.Config["domain_name"])
	}
	return "https://" + domain + confluenceAPIPath, nil
}

func confluencePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return confluenceDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("confluence config page_size must be an integer: %w", err)
	}
	if value < 1 || value > confluenceMaxPageSize {
		return 0, fmt.Errorf("confluence config page_size must be between 1 and %d", confluenceMaxPageSize)
	}
	return value, nil
}

func confluenceMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("confluence config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("confluence config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
