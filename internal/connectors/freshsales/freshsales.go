// Package freshsales implements the native pm Freshsales connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, modeled on
// the stripe reference connector: a thin package that composes a connsdk
// Requester (Token-header auth) with Freshsales-specific stream definitions,
// view-based list endpoints, and page-number pagination.
//
// It self-registers with the connectors registry via RegisterFactory in init();
// the registryset package blank-imports this package in the production binary to
// run that side effect.
//
// The Freshsales (Freshworks CRM) REST API authenticates with an
// "Authorization: Token token=<api_key>" header, lists objects under
// "<resource>/view/<view_id>" returning {"<resource>":[...], "meta":{...}}, and
// paginates by a "page" query parameter bounded by meta.total_pages.
package freshsales

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
	freshsalesAPIPath     = "/crm/sales/api"
	freshsalesUserAgent   = "polymetrics-go-cli"
	freshsalesDefaultView = "0"
	// freshsalesFixtureUpdated is the deterministic updated_at used by fixture
	// records.
	freshsalesFixtureUpdated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("freshsales", New)
}

// New returns the Freshsales connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Freshsales connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "freshsales" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "freshsales",
		DisplayName:     "Freshsales",
		IntegrationType: "api",
		Description:     "Reads Freshsales (Freshworks CRM) contacts, sales accounts, deals, and leads through the Freshsales REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Freshsales.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := freshsalesBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(freshsalesSecret(cfg)) == "" {
		return errors.New("freshsales connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the first page of contacts confirms auth and
	// connectivity without mutating anything.
	endpoint := freshsalesStreamEndpoints["contacts"]
	path := listPath(endpoint.resource, viewID(cfg, "contacts"))
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check freshsales: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: freshsalesStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Freshsales stream starts
// with an empty incremental cursor (full sync). Freshsales only supports
// full_refresh upstream, so the cursor is informational.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Write is unsupported: the Freshsales connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "contacts"
	}
	endpoint, ok := freshsalesStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("freshsales stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := freshsalesMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, viewID(req.Config, stream), maxPages, emit)
}

// harvest drives Freshsales page-number pagination. List responses are shaped
// {"<recordsKey>":[...], "meta":{"total_pages":N,...}}; pages are requested with
// page=1..total_pages. The stop condition reads meta.total_pages from the body,
// falling back to "empty page" when the meta block is absent.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, view string, maxPages int, emit func(connectors.Record) error) error {
	path := listPath(endpoint.resource, view)
	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{"page": []string{strconv.Itoa(page)}}
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read freshsales %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode freshsales %s page: %w", endpoint.resource, err)
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
		totalPages := totalPagesFrom(resp.Body)
		if totalPages > 0 {
			if page >= totalPages {
				return nil
			}
			continue
		}
		// No meta.total_pages present: stop after a short/empty page is reached.
		// (records was non-empty here, so request the next page.)
	}
	return nil
}

// totalPagesFrom reads meta.total_pages from a Freshsales list response body,
// returning 0 when it is absent or unparsable.
func totalPagesFrom(body []byte) int {
	raw, err := connsdk.StringAt(body, "meta.total_pages")
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise freshsales credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                  int64(i),
			"first_name":          fmt.Sprintf("Fixture%d", i),
			"last_name":           "User",
			"display_name":        fmt.Sprintf("Fixture%d User", i),
			"email":               fmt.Sprintf("fixture+%d@example.com", i),
			"name":                fmt.Sprintf("%s Fixture %d", endpoint.resource, i),
			"company_name":        fmt.Sprintf("Fixture Co %d", i),
			"website":             "https://example.com",
			"phone":               "+1-555-0100",
			"mobile_number":       "+1-555-0101",
			"work_number":         "+1-555-0102",
			"job_title":           "Engineer",
			"city":                "Springfield",
			"country":             "US",
			"owner_id":            int64(1000),
			"amount":              float64(1000 * i),
			"currency_id":         int64(1),
			"deal_stage_id":       int64(2),
			"deal_pipeline_id":    int64(1),
			"sales_account_id":    int64(10),
			"probability":         int64(50),
			"lead_stage_id":       int64(1),
			"industry_type_id":    int64(3),
			"number_of_employees": int64(42),
			"annual_revenue":      float64(1000000),
			"expected_close":      freshsalesFixtureUpdated,
			"created_at":          freshsalesFixtureUpdated,
			"updated_at":          freshsalesFixtureUpdated,
		}
		record := endpoint.mapRecord(item)
		if cursor := connsdk.Cursor(req.State); cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Token-header auth and the
// resolved base URL. The secret only ever flows into connsdk.APIKeyHeader; it is
// never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := freshsalesBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := freshsalesSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("freshsales connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", secret, "Token token="),
		UserAgent: freshsalesUserAgent,
	}, nil
}

func freshsalesSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// freshsalesBaseURL resolves and validates the base URL. When base_url is set it
// is used directly (with the /crm/sales/api path appended if absent); otherwise
// it is derived from the required domain_name config (e.g.
// mydomain.myfreshworks.com). Any URL must be absolute http/https with a host to
// bound SSRF risk.
func freshsalesBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("freshsales config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("freshsales config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("freshsales config base_url must include a host")
		}
		trimmed := strings.TrimRight(base, "/")
		if !strings.Contains(parsed.Path, freshsalesAPIPath) {
			trimmed += freshsalesAPIPath
		}
		return trimmed, nil
	}

	domain := strings.TrimSpace(cfg.Config["domain_name"])
	if domain == "" {
		return "", errors.New("freshsales connector requires config domain_name (e.g. mydomain.myfreshworks.com)")
	}
	// Allow callers to pass either a bare host or a full URL for domain_name.
	host := domain
	if parsed, err := url.Parse(domain); err == nil && parsed.Host != "" {
		host = parsed.Host
	}
	host = strings.Trim(host, "/")
	if host == "" || strings.ContainsAny(host, " /") {
		return "", fmt.Errorf("freshsales config domain_name %q is invalid", domain)
	}
	return "https://" + host + freshsalesAPIPath, nil
}

// listPath builds the view-based list endpoint path for a resource.
func listPath(resource, view string) string {
	return resource + "/view/" + url.PathEscape(view)
}

// viewID resolves the Freshsales view id for a stream. A view id is required by
// the list endpoints; callers may override per-stream via config
// "<stream>_view_id" or globally via "view_id". Defaults to the "0" alias used
// for the default/all view.
func viewID(cfg connectors.RuntimeConfig, stream string) string {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config[stream+"_view_id"]); v != "" {
			return v
		}
		if v := strings.TrimSpace(cfg.Config["view_id"]); v != "" {
			return v
		}
	}
	return freshsalesDefaultView
}

func freshsalesMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("freshsales config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("freshsales config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
