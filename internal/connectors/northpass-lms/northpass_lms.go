// Package northpasslms implements the native pm Northpass LMS connector. It is a
// declarative-HTTP per-system connector built in the shape of the stripe
// reference: a thin package composing the connsdk toolkit (Requester +
// X-Api-Key auth + RecordsAt extraction) with Northpass-specific stream
// definitions and endpoints.
//
// The Northpass API (https://api.northpass.com/v2) is JSON:API-flavoured:
// list responses carry records under "data" (each {id,type,attributes}) and
// pagination via a "links" object whose "next" is an absolute URL. Auth is the
// X-Api-Key header. Only full-refresh reads are supported upstream, so this
// connector is read-only.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package northpasslms

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
	northpassDefaultBaseURL  = "https://api.northpass.com/v2"
	northpassDefaultPageSize = 100
	northpassMaxPageSize     = 1000
	northpassUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("northpass-lms", New)
}

// New returns the Northpass LMS connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Northpass LMS connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "northpass-lms" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "northpass-lms",
		DisplayName:     "Northpass LMS",
		IntegrationType: "api",
		Description:     "Reads Northpass LMS people, courses, course enrollments, and groups through the Northpass REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Northpass.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := northpassBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(northpassSecret(cfg)) == "" {
		return errors.New("northpass-lms connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the courses list confirms auth and connectivity.
	if err := r.DoJSON(ctx, http.MethodGet, "courses", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check northpass-lms: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: northpassStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "courses"
	}
	endpoint, ok := northpassStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("northpass-lms stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := northpassPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := northpassMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// Write satisfies connectors.Connector. Northpass LMS is read-only here (only
// full-refresh reads are supported), so writes are rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Northpass's JSON:API links.next pagination. List responses
// return {data:[...], links:{next:"<absolute url>"}}; the next page is the
// links.next URL verbatim. connsdk has no exact paginator for "absolute next URL
// in the body", so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	// First request: relative resource path with a page size.
	path := endpoint.resource
	query := url.Values{}
	query.Set("limit", strconv.Itoa(pageSize))

	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read northpass-lms %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode northpass-lms %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "links.next")
		if err != nil {
			return fmt.Errorf("decode northpass-lms %s links.next: %w", endpoint.resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		// links.next is an absolute URL carrying its own query (page=N&limit=..).
		// The Requester treats an http(s) path as absolute and uses it as-is, so
		// clear the merged query to avoid duplicating params.
		path = next
		query = url.Values{}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":   fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"type": endpoint.resource,
			"attributes": map[string]any{
				"name":         fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
				"slug":         fmt.Sprintf("fixture-%d", i),
				"email":        fmt.Sprintf("fixture+%d@example.com", i),
				"first_name":   "Fixture",
				"last_name":    strconv.Itoa(i),
				"status":       "active",
				"created_at":   "2026-01-01T00:00:00Z",
				"updated_at":   "2026-01-02T00:00:00Z",
				"completed_at": "2026-01-03T00:00:00Z",
				"percentage":   100,
				"learner_id":   fmt.Sprintf("person_fixture_%d", i),
				"course_id":    "course_fixture_1",
			},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with X-Api-Key auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := northpassBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := northpassSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("northpass-lms connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("X-Api-Key", secret, ""),
		UserAgent: northpassUserAgent,
	}, nil
}

func northpassSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// northpassBaseURL resolves and validates the base URL. The default is
// api.northpass.com/v2; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func northpassBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return northpassDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("northpass-lms config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("northpass-lms config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("northpass-lms config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func northpassPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return northpassDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("northpass-lms config page_size must be an integer: %w", err)
	}
	if value < 1 || value > northpassMaxPageSize {
		return 0, fmt.Errorf("northpass-lms config page_size must be between 1 and %d", northpassMaxPageSize)
	}
	return value, nil
}

func northpassMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("northpass-lms config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("northpass-lms config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
