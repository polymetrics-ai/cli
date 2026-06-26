// Package float implements the native pm Float connector. It is a declarative-
// HTTP per-system connector built on the connsdk toolkit (Requester + Bearer
// auth + RecordsAt extraction) with Float-specific stream definitions,
// endpoints, and page-number pagination.
//
// Float's v3 REST API (https://developer.float.com/) returns top-level JSON
// arrays and paginates with page/per-page query params, reporting the total
// page count in the X-Pagination-Page-Count response header. The connector is
// read-only: Float's catalog declares full_refresh only and there is no obvious
// safe reverse-ETL write surface, so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package float

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
	floatDefaultBaseURL  = "https://api.float.com/v3"
	floatDefaultPageSize = 200
	floatMaxPageSize     = 200
	floatUserAgent       = "polymetrics-go-cli"
	floatPageCountHeader = "X-Pagination-Page-Count"
)

func init() {
	connectors.RegisterFactory("float", New)
}

// New returns the Float connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Float connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "float" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "float",
		DisplayName:     "Float",
		IntegrationType: "api",
		Description:     "Reads Float people, projects, clients, tasks, and departments through the Float v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Float. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := floatBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(floatSecret(cfg)) == "" {
		return errors.New("float connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the departments list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "departments", url.Values{"per-page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check float: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: floatStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "people"
	}
	endpoint, ok := floatStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("float stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := floatPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := floatMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// Write satisfies connectors.Connector. Float is read-only for reverse ETL in
// this connector, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Float's page-number pagination. List endpoints return a
// top-level JSON array and report the total page count in the
// X-Pagination-Page-Count header. We request page 1, read the header to learn
// the last page, and keep requesting until we reach it (or maxPages). This shape
// is not covered by connsdk's body-token/short-page paginators, so the loop
// lives here, built on connsdk.Requester + connsdk.RecordsAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per-page", strconv.Itoa(pageSize))

	totalPages := 0 // unknown until the first response
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if maxPages > 0 && page > maxPages {
			return nil
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read float %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode float %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}

		if totalPages == 0 {
			totalPages = parsePageCount(resp.Header.Get(floatPageCountHeader))
		}
		// Stop when we've consumed the last reported page, or when a page came
		// back empty (defensive: header missing or lying).
		if len(records) == 0 {
			return nil
		}
		if totalPages > 0 && page >= totalPages {
			return nil
		}
		if totalPages == 0 && len(records) < pageSize {
			// No page-count header and a short page: assume we're done.
			return nil
		}
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise float credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	idField := stream + "_id"
	switch stream {
	case "people":
		idField = "people_id"
	case "tasks":
		idField = "task_id"
	case "clients":
		idField = "client_id"
	case "departments":
		idField = "department_id"
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			idField:         i,
			"name":          fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"email":         fmt.Sprintf("fixture+%d@example.com", i),
			"client_id":     1,
			"project_id":    1,
			"department_id": 1,
			"active":        1,
			"billable":      1,
			"budget_total":  1000 * i,
			"created":       "2026-01-01T00:00:00Z",
			"modified":      "2026-01-02T00:00:00Z",
			"connector":     "float",
			"fixture":       true,
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "float"
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Bearer auth and the resolved
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := floatBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := floatSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("float connector requires secret access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: floatUserAgent,
	}, nil
}

func floatSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}

// floatBaseURL resolves and validates the base URL. The default is
// api.float.com; any override must be an absolute https (or http for local test
// servers) URL with a host to bound SSRF risk.
func floatBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return floatDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("float config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("float config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("float config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func floatPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return floatDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("float config page_size must be an integer: %w", err)
	}
	if value < 1 || value > floatMaxPageSize {
		return 0, fmt.Errorf("float config page_size must be between 1 and %d", floatMaxPageSize)
	}
	return value, nil
}

func floatMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("float config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("float config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func parsePageCount(header string) int {
	header = strings.TrimSpace(header)
	if header == "" {
		return 0
	}
	n, err := strconv.Atoi(header)
	if err != nil || n < 0 {
		return 0
	}
	return n
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
