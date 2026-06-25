// Package onehms implements the native pm 100ms connector. It is a declarative-
// HTTP per-system connector built in the stripe template shape: a thin package
// that composes the connsdk toolkit (Requester + Bearer auth + RecordsAt
// extraction + cursor state) with 100ms-specific stream definitions and
// endpoints.
//
// The Go package is named onehms because a Go identifier cannot begin with a
// digit, but the directory, registry key, and connector Name() are all "100ms".
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
//
// 100ms exposes a REST API at https://api.100ms.live/v2. Every list endpoint
// returns {data:[...], last:"<cursor>"}; the next page is requested by passing
// the previous response's last value as ?start=<last>. An empty last means there
// are no further pages. Auth is a Bearer management token.
package onehms

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
	defaultBaseURL  = "https://api.100ms.live/v2"
	defaultPageSize = 100
	maxPageSize     = 100
	userAgent       = "polymetrics-go-cli"
	// fixtureCreated is the deterministic created_at used by fixture-mode records.
	fixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("100ms", New)
}

// New returns the 100ms connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm 100ms connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "100ms" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "100ms",
		DisplayName:     "100ms",
		IntegrationType: "api",
		Description:     "Reads 100ms rooms, sessions, recordings, and templates through the 100ms server-side REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to 100ms. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := baseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(managementToken(cfg)) == "" {
		return errors.New("100ms connector requires secret management_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the rooms list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "rooms", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check 100ms: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a 100ms stream starts with an
// empty incremental cursor (full refresh — the only sync mode 100ms supports).
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
		stream = "rooms"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("100ms stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	size, err := pageSize(req.Config)
	if err != nil {
		return err
	}
	limitPages, err := maxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, size, limitPages, emit)
}

// harvest drives 100ms's last/start cursor pagination. List responses are
// {data:[...], last:"<id>"}; the next page is requested with start=<last>. An
// empty last (or a page that returns no records) ends the walk. There is no
// connsdk paginator for this exact body-cursor shape, so the loop lives here,
// built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))

	start := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if start != "" {
			query.Set("start", start)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read 100ms %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode 100ms %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		last, err := connsdk.StringAt(resp.Body, "last")
		if err != nil {
			return fmt.Errorf("decode 100ms %s last cursor: %w", endpoint.resource, err)
		}
		last = strings.TrimSpace(last)
		// Stop on an empty cursor, no records, or a cursor that did not advance.
		if last == "" || len(records) == 0 || last == start {
			return nil
		}
		start = last
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise 100ms credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                   fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"name":                 fmt.Sprintf("Fixture %s %d", stream, i),
			"enabled":              true,
			"active":               i == 1,
			"description":          "fixture record",
			"customer_id":          "cust_fixture",
			"template_id":          "tmpl_fixture",
			"region":               "us",
			"large_room":           false,
			"max_duration_seconds": int64(3600),
			"room_id":              "rooms_fixture_1",
			"session_id":           "sessions_fixture_1",
			"status":               "completed",
			"size":                 int64(1024 * i),
			"duration":             int64(60 * i),
			"default":              i == 1,
			"created_at":           fixtureCreated,
			"updated_at":           fixtureCreated,
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

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The management token only ever flows into connsdk.Bearer; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := managementToken(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("100ms connector requires secret management_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(token),
		UserAgent: userAgent,
	}, nil
}

// Write satisfies the connectors.Connector interface. The 100ms connector is
// read-only: the server-side API exposes no safe reverse-ETL surface for pm, so
// writes are unsupported (Capabilities.Write is false).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func managementToken(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["management_token"]
}

// baseURL resolves and validates the base URL. The default is api.100ms.live/v2;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("100ms config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("100ms config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("100ms config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("100ms config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("100ms config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("100ms config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("100ms config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
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
