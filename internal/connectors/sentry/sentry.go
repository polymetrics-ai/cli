// Package sentry implements the native pm Sentry source connector. It is a
// declarative-HTTP per-system connector modeled on the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction) with Sentry-specific stream definitions and endpoints.
//
// Sentry list endpoints return top-level JSON arrays and paginate with the
// RFC 5988 Link header, but with Sentry's own twist: a rel="next" link is always
// present and the results="true" attribute is what actually signals more data.
// connsdk.LinkHeaderPaginator follows rel="next" unconditionally, so the page
// loop lives in-package (harvest) to honor results="true".
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect. Read-only: Sentry source supports full_refresh
// only, so Capabilities.Write is false.
package sentry

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
	sentryDefaultHostname = "sentry.io"
	sentryAPIPrefix       = "api/0"
	sentryDefaultPageSize = 100
	sentryMaxPageSize     = 100
	sentryUserAgent       = "polymetrics-go-cli"
	// sentryFixtureDate is the deterministic ISO-8601 timestamp used by
	// fixture-mode records.
	sentryFixtureDate = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("sentry", New)
}

// New returns the Sentry connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Sentry source connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "sentry" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "sentry",
		DisplayName:     "Sentry",
		IntegrationType: "api",
		Description:     "Reads Sentry projects, issues, error events, and releases through the Sentry REST API (read-only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Sentry. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := sentryBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(sentrySecret(cfg)) == "" {
		return errors.New("sentry connector requires secret auth_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the projects list confirms auth and connectivity
	// without requiring org/project config or mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, sentryAPIPrefix+"/projects/", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check sentry: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. The Sentry source is
// read-only (full_refresh only), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: sentryStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "projects"
	}
	endpoint, ok := sentryStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("sentry stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	path, err := endpointPath(endpoint, req.Config)
	if err != nil {
		return err
	}
	pageSize, err := sentryPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := sentryMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, path, endpoint.mapRecord, pageSize, maxPages, emit)
}

// harvest drives Sentry's Link-header cursor pagination. Each list response is a
// top-level JSON array; the Link header carries a rel="next" entry whose
// results="true"/"false" attribute is the real "more pages" signal (a next link
// is always present). The loop is built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, mapRecord func(map[string]any) connectors.Record, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))

	cursor := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read sentry %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode sentry %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		next, more := nextCursor(resp.Header.Get("Link"))
		if !more || next == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise sentry credential-free (mirrors stripe's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	endpoint := sentryStreamEndpoints[stream]
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":           fmt.Sprintf("%s_fixture_%d", stream, i),
			"version":      fmt.Sprintf("%s_fixture_%d", stream, i),
			"shortId":      fmt.Sprintf("FIX-%d", i),
			"shortVersion": fmt.Sprintf("fix-%d", i),
			"slug":         fmt.Sprintf("fixture-%d", i),
			"name":         fmt.Sprintf("Fixture %d", i),
			"title":        fmt.Sprintf("Fixture issue %d", i),
			"message":      fmt.Sprintf("fixture message %d", i),
			"culprit":      "fixture.module in handler",
			"level":        "error",
			"status":       "unresolved",
			"type":         "error",
			"platform":     "python",
			"count":        strconv.Itoa(i),
			"userCount":    int64(i),
			"eventID":      fmt.Sprintf("evt_fixture_%d", i),
			"groupID":      fmt.Sprintf("grp_fixture_%d", i),
			"ref":          "main",
			"url":          "https://example.com/fixture",
			"isPublic":     false,
			"isBookmarked": false,
			"firstSeen":    sentryFixtureDate,
			"lastSeen":     sentryFixtureDate,
			"dateCreated":  sentryFixtureDate,
			"dateReleased": sentryFixtureDate,
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
	base, err := sentryBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := sentrySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("sentry connector requires secret auth_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: sentryUserAgent,
	}, nil
}

// endpointPath fills the endpoint's path template with the configured org and
// project slugs as its scope requires. The leading api/0 prefix is added for
// non-overridden base URLs by sentryBaseURL only when no explicit base_url is
// set; here every endpoint path carries the api/0 prefix so a custom base_url
// (e.g. an httptest server root) still hits /api/0/... .
func endpointPath(e streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	switch e.scope {
	case scopeGlobal:
		return sentryAPIPrefix + "/" + e.path, nil
	case scopeOrg:
		org, err := requireSlug(cfg, "organization")
		if err != nil {
			return "", err
		}
		return sentryAPIPrefix + "/" + fmt.Sprintf(e.path, org), nil
	case scopeProject:
		org, err := requireSlug(cfg, "organization")
		if err != nil {
			return "", err
		}
		project, err := requireSlug(cfg, "project")
		if err != nil {
			return "", err
		}
		return sentryAPIPrefix + "/" + fmt.Sprintf(e.path, org, project), nil
	default:
		return "", fmt.Errorf("sentry endpoint has unknown scope %d", e.scope)
	}
}

func requireSlug(cfg connectors.RuntimeConfig, key string) (string, error) {
	v := strings.TrimSpace(cfg.Config[key])
	if v == "" {
		return "", fmt.Errorf("sentry connector requires config %s", key)
	}
	return url.PathEscape(v), nil
}

func sentrySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["auth_token"]
}

// sentryBaseURL resolves and validates the base URL. With no base_url override it
// is built from the hostname config (default sentry.io) as https://<hostname>.
// Any explicit override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func sentryBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		hostname := strings.TrimSpace(cfg.Config["hostname"])
		if hostname == "" {
			hostname = sentryDefaultHostname
		}
		// hostname is a bare host; guard against scheme/path injection.
		if strings.ContainsAny(hostname, "/ ") || strings.Contains(hostname, "://") {
			return "", fmt.Errorf("sentry config hostname must be a bare host, got %q", hostname)
		}
		return "https://" + hostname, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("sentry config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("sentry config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("sentry config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func sentryPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return sentryDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("sentry config page_size must be an integer: %w", err)
	}
	if value < 1 || value > sentryMaxPageSize {
		return 0, fmt.Errorf("sentry config page_size must be between 1 and %d", sentryMaxPageSize)
	}
	return value, nil
}

func sentryMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("sentry config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("sentry config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// nextCursor parses a Sentry Link header and returns the rel="next" cursor along
// with whether that page actually has more results (results="true"). Sentry
// always emits a rel="next" entry, so the results attribute is authoritative.
func nextCursor(header string) (cursor string, more bool) {
	if header == "" {
		return "", false
	}
	for _, part := range strings.Split(header, ",") {
		segs := strings.Split(part, ";")
		if len(segs) < 2 {
			continue
		}
		urlPart := strings.TrimSpace(segs[0])
		if !strings.HasPrefix(urlPart, "<") || !strings.HasSuffix(urlPart, ">") {
			continue
		}
		isNext := false
		results := false
		cur := ""
		for _, attr := range segs[1:] {
			k, v := splitAttr(attr)
			switch k {
			case "rel":
				isNext = v == "next"
			case "results":
				results = v == "true"
			case "cursor":
				cur = v
			}
		}
		if !isNext {
			continue
		}
		if cur == "" {
			// Fall back to the cursor query param embedded in the URL.
			cur = cursorFromURL(urlPart[1 : len(urlPart)-1])
		}
		return cur, results
	}
	return "", false
}

// splitAttr parses a Link header attribute of the form key="value" or key=value.
func splitAttr(attr string) (string, string) {
	attr = strings.TrimSpace(attr)
	eq := strings.IndexByte(attr, '=')
	if eq < 0 {
		return strings.ToLower(attr), ""
	}
	key := strings.ToLower(strings.TrimSpace(attr[:eq]))
	val := strings.TrimSpace(attr[eq+1:])
	val = strings.Trim(val, `"`)
	return key, val
}

// cursorFromURL extracts the cursor query parameter from a next-page URL.
func cursorFromURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Query().Get("cursor")
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}
