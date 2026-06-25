// Package clockodo implements the native pm Clockodo connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + custom-header auth + RecordsAt/StringAt extraction) with
// Clockodo-specific stream definitions, endpoints, and page-based pagination.
//
// Clockodo authenticates with three custom request headers rather than a bearer
// token: X-ClockodoApiUser (the account email), X-ClockodoApiKey (the API key),
// and X-Clockodo-External-Application (an application;contact identifier). The
// connector is read-only.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package clockodo

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
	clockodoDefaultBaseURL = "https://my.clockodo.com/api"
	clockodoUserAgent      = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("clockodo", New)
}

// New returns the Clockodo connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Clockodo connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "clockodo" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "clockodo",
		DisplayName:     "Clockodo",
		IntegrationType: "api",
		Description:     "Reads Clockodo customers, projects, services, and users through the Clockodo REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Clockodo. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "v2/users", nil, nil, nil); err != nil {
		return fmt.Errorf("check clockodo: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: clockodoStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "customers"
	}
	endpoint, ok := clockodoStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("clockodo stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := clockodoMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, maxPages, emit)
}

// Write is required by the connectors.Connector interface but Clockodo is
// read-only here (Capabilities.Write=false); it always reports the operation as
// unsupported.
func (c Connector) Write(ctx context.Context, _ connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.WriteResult{}, err
	}
	return connectors.WriteResult{RecordsFailed: len(records)}, fmt.Errorf("clockodo connector is read-only: %w", connectors.ErrUnsupportedOperation)
}

// harvest drives Clockodo's page-based pagination. Paginated list endpoints
// return {paging:{current_page,count_pages,...}, <recordsKey>:[...]}; the next
// page is requested with page=<n+1> until current_page reaches count_pages.
// Non-paginated endpoints are read in a single request. There is no body-token
// paginator in connsdk for this exact shape, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, maxPages int, emit func(connectors.Record) error) error {
	page := 1
	for fetched := 0; maxPages == 0 || fetched < maxPages; fetched++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if endpoint.paginated {
			query.Set("page", strconv.Itoa(page))
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read clockodo %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.recordsKey)
		if err != nil {
			return fmt.Errorf("decode clockodo %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if !endpoint.paginated {
			return nil
		}
		current, countPages := clockodoPaging(resp.Body)
		if countPages <= 0 || current >= countPages {
			return nil
		}
		if len(records) == 0 {
			return nil
		}
		page = current + 1
	}
	return nil
}

// clockodoPaging reads current_page and count_pages from the response body's
// `paging` object. Either may be absent (returning 0), which the harvest loop
// treats as "no further pages".
func clockodoPaging(body []byte) (current, countPages int) {
	cur, _ := connsdk.StringAt(body, "paging.current_page")
	cnt, _ := connsdk.StringAt(body, "paging.count_pages")
	current, _ = strconv.Atoi(strings.TrimSpace(cur))
	countPages, _ = strconv.Atoi(strings.TrimSpace(cnt))
	return current, countPages
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise clockodo credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":               i,
			"name":             fmt.Sprintf("Fixture %s %d", strings.TrimSuffix(stream, "s"), i),
			"number":           strconv.Itoa(1000 + i),
			"active":           true,
			"billable_default": true,
			"note":             "",
			"color":            i,
			"customers_id":     1,
			"budget_money":     nil,
			"budget_is_hours":  false,
			"completed":        false,
			"deadline":         nil,
			"email":            fmt.Sprintf("fixture+%d@example.com", i),
			"role":             "co-worker",
			"language":         "en",
			"timezone":         "Europe/Berlin",
			"teams_id":         nil,
		}
		record := endpoint.mapRecord(item)
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with Clockodo's custom-header auth,
// the resolved base URL, and the required external-application identifier. The
// api_key secret only ever flows into the X-ClockodoApiKey header; it is never
// logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := clockodoBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey := clockodoSecret(cfg)
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("clockodo connector requires secret api_key")
	}
	email := strings.TrimSpace(cfg.Config["email_address"])
	if email == "" {
		return nil, errors.New("clockodo connector requires config email_address")
	}
	externalApp := strings.TrimSpace(cfg.Config["external_application"])
	if externalApp == "" {
		return nil, errors.New("clockodo connector requires config external_application")
	}
	headers := map[string]string{
		"X-ClockodoApiUser":               email,
		"X-ClockodoApiKey":                apiKey,
		"X-Clockodo-External-Application": externalApp,
	}
	if lang := strings.TrimSpace(cfg.Config["language"]); lang != "" {
		headers["Accept-Language"] = lang
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		UserAgent:      clockodoUserAgent,
		DefaultHeaders: headers,
	}, nil
}

func clockodoSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// clockodoBaseURL resolves and validates the base URL. The default is
// my.clockodo.com/api; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func clockodoBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return clockodoDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("clockodo config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("clockodo config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("clockodo config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func clockodoMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("clockodo config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("clockodo config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
