// Package sevenshifts implements the native pm 7shifts connector. It is a
// declarative-HTTP per-system connector built on the same shape as the stripe
// reference connector: a thin package that composes the connsdk toolkit
// (Requester + Bearer auth + RecordsAt extraction + cursor state) with
// 7shifts-specific stream definitions, endpoints, and pagination.
//
// The 7shifts REST API (v2) uses Bearer authentication on an access token and
// cursor pagination: list responses are {"data":[...],"meta":{"cursor":{"next":
// "<token>"}}} and the next page is requested with ?cursor=<token>. Most
// resources are nested under /v2/company/{company_id}/...; companies is the one
// top-level stream.
//
// The package self-registers with the connectors registry via RegisterFactory in
// init() under the bare system name "7shifts"; the registryset package
// blank-imports this package in the production binary to run that side effect.
// The Go package identifier is sevenshifts because an identifier may not begin
// with a digit, but the registry key, directory, and Name() are all "7shifts".
package sevenshifts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	defaultBaseURL  = "https://api.7shifts.com"
	defaultPageSize = 100
	maxPageSize     = 200
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("7shifts", New)
}

// New returns the 7shifts connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm 7shifts connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "7shifts" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "7shifts",
		DisplayName:     "7shifts",
		IntegrationType: "api",
		Description:     "Reads 7shifts companies, locations, departments, roles, users, shifts, and time punches through the 7shifts v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to 7shifts. In
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
	if strings.TrimSpace(secret(cfg)) == "" {
		return errors.New("7shifts connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the companies list confirms auth and connectivity
	// without mutating anything; it needs no company_id.
	if err := r.DoJSON(ctx, http.MethodGet, "/v2/companies", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check 7shifts: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams()}, nil
}

// Write satisfies the connectors.Connector interface. 7shifts is read-only for
// pm (Capabilities.Write=false), so writes are explicitly unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a 7shifts stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time via the modified_since filter.
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
		stream = "companies"
	}
	endpoint, ok := streamEndpoints[stream]
	if !ok {
		return fmt.Errorf("7shifts stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}

	companyID := strings.TrimSpace(req.Config.Config["company_id"])
	if endpoint.companyScoped && companyID == "" {
		return fmt.Errorf("7shifts stream %q requires config company_id", stream)
	}
	path := endpoint.pathFor(companyID)

	pageSize, err := pageSizeFor(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPagesFor(req.Config)
	if err != nil {
		return err
	}
	modifiedSince, err := incrementalLowerBound(req)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, path, endpoint.mapRecord, pageSize, maxPages, modifiedSince, emit)
}

// harvest drives 7shifts cursor pagination. List responses are
// {"data":[...],"meta":{"cursor":{"next":"<token>"}}}; the next page is requested
// with ?cursor=<token>, stopping when meta.cursor.next is null/absent. The loop
// lives here (built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt)
// because the stop condition is body-token-based rather than page-count-based.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, mapRecord func(map[string]any) connectors.Record, pageSize, maxPages int, modifiedSince string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if modifiedSince != "" {
		base.Set("modified_since", modifiedSince)
	}

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
			return fmt.Errorf("read 7shifts %s: %w", path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode 7shifts %s page: %w", path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "meta.cursor.next")
		if err != nil {
			return fmt.Errorf("decode 7shifts %s cursor: %w", path, err)
		}
		next = strings.TrimSpace(next)
		if next == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise 7shifts credential-free (mirrors the stripe
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	endpoint := streamEndpoints[stream]
	companyID := strings.TrimSpace(req.Config.Config["company_id"])
	if companyID == "" {
		companyID = "1"
	}
	companyNum, _ := strconv.Atoi(companyID)
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            i,
			"company_id":    companyNum,
			"location_id":   100 + i,
			"department_id": 200 + i,
			"role_id":       300 + i,
			"user_id":       400 + i,
			"name":          fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"first_name":    fmt.Sprintf("Fixture%d", i),
			"last_name":     "Example",
			"email":         fmt.Sprintf("fixture+%d@example.com", i),
			"type":          "employee",
			"active":        true,
			"approved":      true,
			"open":          false,
			"deleted":       false,
			"start":         "2026-01-01T09:00:00+00:00",
			"end":           "2026-01-01T17:00:00+00:00",
			"clocked_in":    "2026-01-01T09:01:00+00:00",
			"clocked_out":   "2026-01-01T17:02:00+00:00",
			"currency":      "USD",
			"country_code":  "US",
			"timezone":      "America/New_York",
			"created":       "2026-01-01T00:00:00+00:00",
			"modified":      fmt.Sprintf("2026-01-0%dT00:00:00+00:00", i),
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
// base URL. The secret only ever flows into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	token := secret(cfg)
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("7shifts connector requires secret access_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(token),
		UserAgent: userAgent,
	}, nil
}

// incrementalLowerBound returns the modified_since lower bound (YYYY-MM-DD),
// derived from the incremental cursor (if any) or else the start_date config. An
// empty result means no lower bound (full sync). 7shifts filters by date only.
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return toDate(cursor), nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	t, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		// Already a bare date or unexpected format: pass the date prefix through.
		return toDate(startDate), nil
	}
	return t.UTC().Format("2006-01-02"), nil
}

// toDate reduces an RFC3339 timestamp (or any value) to its YYYY-MM-DD prefix.
func toDate(value string) string {
	value = strings.TrimSpace(value)
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.UTC().Format("2006-01-02")
	}
	if len(value) >= 10 {
		return value[:10]
	}
	return value
}

func secret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}

// baseURL resolves and validates the base URL. The default is api.7shifts.com;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("7shifts config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("7shifts config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("7shifts config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pageSizeFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("7shifts config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("7shifts config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPagesFor(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("7shifts config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("7shifts config max_pages must be 0 for unlimited or a positive integer")
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
