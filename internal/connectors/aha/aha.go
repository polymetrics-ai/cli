// Package aha implements the native pm Aha! source connector. It follows the
// declarative-HTTP template established by the stripe package: a thin package that
// composes the connsdk toolkit (Requester + Bearer auth + RecordsAt extraction +
// cursor state) with Aha!-specific stream definitions, endpoints, and the Aha!
// page-number pagination envelope.
//
// Aha! is a planning/roadmapping product; this connector is read-only (the source
// catalog supports full_refresh only and there is no obvious safe reverse-ETL
// write surface), so Capabilities.Write is false and there is no write.go.
//
// Like stripe/github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package in
// the production binary to run that side effect.
package aha

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
	ahaDefaultBaseURL  = "https://secure.aha.io"
	ahaAPIPrefix       = "api/v1"
	ahaDefaultPageSize = 30
	ahaMaxPageSize     = 200
	ahaUserAgent       = "polymetrics-go-cli"
	// ahaFixtureCreated/Updated are deterministic ISO-8601 timestamps used by the
	// fixture-mode records.
	ahaFixtureCreated = "2026-01-01T00:00:00Z"
	ahaFixtureUpdated = "2026-01-02T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("aha", New)
}

// New returns the Aha! connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Aha! connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "aha" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "aha",
		DisplayName:     "Aha!",
		IntegrationType: "api",
		Description:     "Reads Aha! features, products, ideas, releases, initiatives, and goals through the Aha! REST API (read-only).",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Aha!. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := ahaBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(ahaSecret(cfg)) == "" {
		return errors.New("aha connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the products list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, ahaAPIPrefix+"/products", url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check aha: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: ahaStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Aha! is a read-only source
// (Capabilities.Write is false), so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: an Aha! stream starts with an
// empty incremental cursor (full sync).
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
		stream = "features"
	}
	endpoint, ok := ahaStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("aha stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := ahaPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := ahaMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Aha!'s page-number pagination. Aha! lists return
// {"<resource>":[...],"pagination":{"total_records":N,"total_pages":N,"current_page":N}}.
// We request page=1,2,... until current_page >= total_pages (or a page returns no
// records). connsdk's PageNumberPaginator stops on a short page, which would
// under-read a final exactly-full page; reading the pagination object is the
// faithful Aha! contract, so the loop lives here on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	path := ahaAPIPrefix + "/" + endpoint.resource
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("page", strconv.Itoa(page))
		query.Set("per_page", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read aha %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode aha %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) == 0 {
			return nil
		}
		totalPages := intAt(resp.Body, "pagination.total_pages")
		currentPage := intAt(resp.Body, "pagination.current_page")
		if currentPage == 0 {
			currentPage = page
		}
		// Stop once we have consumed the last page. If total_pages is absent
		// (0), fall back to stopping on a short page.
		if totalPages > 0 {
			if currentPage >= totalPages {
				return nil
			}
		} else if len(records) < pageSize {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise aha credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"reference_num":    fmt.Sprintf("FIX-%d", i),
			"reference_prefix": "FIX",
			"name":             fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"created_at":       ahaFixtureCreated,
			"updated_at":       ahaFixtureUpdated,
			"start_date":       "2026-01-01",
			"due_date":         "2026-02-01",
			"release_date":     "2026-03-01",
			"released":         false,
			"product_line":     false,
			"score":            float64(10 * i),
			"votes":            i,
			"workflow_status":  map[string]any{"id": "1", "name": "In progress"},
			"url":              fmt.Sprintf("https://secure.aha.io/%s/%d", endpoint.resource, i),
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

// requester builds a connsdk.Requester wired with Bearer auth (Aha! API keys are
// passed exactly like OAuth bearer tokens) and the resolved base URL. The secret
// only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := ahaBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := ahaSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("aha connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: ahaUserAgent,
	}, nil
}

func ahaSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// ahaBaseURL resolves and validates the base URL. The default is secure.aha.io;
// because Aha! is account-scoped (<company>.aha.io), the base_url override is the
// primary way to point at an account. Any override must be an absolute https (or
// http for local test servers) URL with a host to bound SSRF risk.
func ahaBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		// Allow the Aha!-conventional "url" config field (account domain) as an
		// alias for base_url.
		base = strings.TrimSpace(cfg.Config["url"])
	}
	if base == "" {
		return ahaDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("aha config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("aha config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("aha config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func ahaPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["per_page"])
	if raw == "" {
		raw = strings.TrimSpace(cfg.Config["page_size"])
	}
	if raw == "" {
		return ahaDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("aha config per_page must be an integer: %w", err)
	}
	if value < 1 || value > ahaMaxPageSize {
		return 0, fmt.Errorf("aha config per_page must be between 1 and %d", ahaMaxPageSize)
	}
	return value, nil
}

func ahaMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("aha config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("aha config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

// intAt reads an integer value at a dotted path in a JSON body, returning 0 when
// the value is absent or non-numeric.
func intAt(body []byte, path string) int {
	s, err := connsdk.StringAt(body, path)
	if err != nil {
		return 0
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	// Tolerate float-formatted integers (e.g. "2.0").
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(f)
	}
	return 0
}
