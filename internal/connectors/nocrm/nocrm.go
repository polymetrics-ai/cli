// Package nocrm implements the native pm noCRM.io connector. It follows the
// declarative-HTTP per-system connector shape established by the stripe package:
// a thin package that composes the connsdk toolkit (Requester + X-API-KEY auth +
// RecordsAt extraction) with noCRM-specific stream definitions and endpoints.
//
// noCRM's REST API is read-oriented for analytics use; this connector is
// read-only (full_refresh) and exposes a fixture mode so the conformance harness
// can run it without live credentials.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package nocrm

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
	nocrmDefaultBaseURL  = "https://api.nocrm.io/api/v2"
	nocrmDefaultPageSize = 100
	nocrmMaxPageSize     = 100
	nocrmUserAgent       = "polymetrics-go-cli"
	nocrmTotalCountHdr   = "X-TOTAL-COUNT"
)

func init() {
	connectors.RegisterFactory("nocrm", New)
}

// New returns the noCRM connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm noCRM connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "nocrm" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "nocrm",
		DisplayName:     "noCRM",
		IntegrationType: "api",
		Description:     "Reads noCRM.io leads, pipelines, users, teams, and prospecting lists through the noCRM REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to noCRM. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := nocrmBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(nocrmSecret(cfg)) == "" {
		return errors.New("nocrm connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the pipelines list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "pipelines", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check nocrm: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: nocrmStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "leads"
	}
	endpoint, ok := nocrmStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("nocrm stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := nocrmPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := nocrmMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives noCRM's offset/limit pagination. List endpoints return a
// top-level JSON array and an X-TOTAL-COUNT response header. The next page is
// requested with offset += limit. The loop stops when a short page is returned,
// when the running count reaches X-TOTAL-COUNT, or when maxPages is hit. There is
// no body-token paginator in connsdk for this header-driven shape, so the loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	seen := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(offset))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read nocrm %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode nocrm %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		seen += len(records)

		// Stop on a short page (fewer than the requested limit).
		if len(records) < pageSize {
			return nil
		}
		// Stop once we have observed every record the server advertises.
		if total, ok := parseTotalCount(resp.Header.Get(nocrmTotalCountHdr)); ok && seen >= total {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise nocrm credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              i,
			"title":           fmt.Sprintf("%s fixture %d", endpoint.resource, i),
			"name":            fmt.Sprintf("%s fixture %d", endpoint.resource, i),
			"status":          "todo",
			"step":            "New",
			"step_id":         1,
			"pipeline":        "Default",
			"pipeline_id":     1,
			"amount":          1000 * i,
			"currency":        "USD",
			"probability":     50,
			"user_id":         1,
			"team_id":         1,
			"position":        i,
			"default":         i == 1,
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"firstname":       "Fixture",
			"lastname":        fmt.Sprintf("User %d", i),
			"admin":           i == 1,
			"active":          true,
			"archived":        false,
			"prospects_count": 10 * i,
			"created_at":      "2026-01-01T00:00:00Z",
			"updated_at":      "2026-01-02T00:00:00Z",
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-API-KEY auth and the resolved
// base URL. The secret only ever flows into connsdk.APIKeyHeader; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := nocrmBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := nocrmSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("nocrm connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-API-KEY", secret, ""),
		UserAgent: nocrmUserAgent,
	}, nil
}

func nocrmSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// nocrmBaseURL resolves and validates the base URL. The default targets the
// noCRM API; in production the per-account subdomain base URL is supplied via the
// base_url config (e.g. https://yourcompany.nocrm.io/api/v2). Any value must be an
// absolute https (or http for local test servers) URL with a host to bound SSRF
// risk. When a subdomain is configured without an explicit base_url, the URL is
// derived from it.
func nocrmBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if sub := nocrmSubdomain(cfg); sub != "" {
			return fmt.Sprintf("https://%s.nocrm.io/api/v2", sub), nil
		}
		return nocrmDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("nocrm config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("nocrm config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("nocrm config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

// nocrmSubdomain returns a sanitized subdomain (letters, digits, hyphens) from
// config, guarding against injection into the derived URL host.
func nocrmSubdomain(cfg connectors.RuntimeConfig) string {
	if cfg.Config == nil {
		return ""
	}
	raw := strings.TrimSpace(cfg.Config["subdomain"])
	if raw == "" {
		return ""
	}
	for _, r := range raw {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-'
		if !ok {
			return ""
		}
	}
	return raw
}

func nocrmPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return nocrmDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nocrm config page_size must be an integer: %w", err)
	}
	if value < 1 || value > nocrmMaxPageSize {
		return 0, fmt.Errorf("nocrm config page_size must be between 1 and %d", nocrmMaxPageSize)
	}
	return value, nil
}

func nocrmMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("nocrm config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("nocrm config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

// parseTotalCount parses the X-TOTAL-COUNT header into an int. A missing or
// malformed value returns ok=false so the caller falls back to short-page
// detection.
func parseTotalCount(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// Write is unsupported: noCRM is exposed read-only for analytics ingestion.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
