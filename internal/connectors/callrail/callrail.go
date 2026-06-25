// Package callrail implements the native pm CallRail connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit, following
// the stripe reference shape: a thin package that composes a connsdk Requester
// (Token-header auth + RecordsAt extraction + page-number pagination) with
// CallRail-specific stream definitions and endpoints.
//
// CallRail is read-only here (call tracking analytics has no obvious safe
// reverse-ETL writes), so Capabilities.Write is false.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package callrail

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
	callrailDefaultBaseURL  = "https://api.callrail.com/v3"
	callrailDefaultPageSize = 100
	callrailMaxPageSize     = 250
	callrailUserAgent       = "polymetrics-go-cli"
	// callrailFixtureTime is the deterministic timestamp used by fixture records.
	callrailFixtureTime = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("callrail", New)
}

// New returns the CallRail connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm CallRail connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "callrail" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "callrail",
		DisplayName:     "CallRail",
		IntegrationType: "api",
		Description:     "Reads CallRail calls, companies, users, and text messages through the CallRail v3 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to CallRail. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := callrailBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(callrailAPIKey(cfg)) == "" {
		return errors.New("callrail connector requires secret api_key")
	}
	account, err := callrailAccountID(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the companies list confirms auth and connectivity
	// without mutating anything.
	path := accountPath(account, "companies.json")
	if err := r.DoJSON(ctx, http.MethodGet, path, url.Values{"per_page": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check callrail: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: callrailStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a CallRail stream starts with
// an empty incremental cursor (full sync), which the start_date config can raise
// at read time.
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
		stream = "calls"
	}
	endpoint, ok := callrailStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("callrail stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	account, err := callrailAccountID(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := callrailPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := callrailMaxPages(req.Config)
	if err != nil {
		return err
	}
	startDate, err := startDateParam(req)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, account, endpoint, pageSize, maxPages, startDate, emit)
}

// harvest drives CallRail's page-number pagination. List responses are
// {page, per_page, total_pages, total_records, <arrayKey>:[...]}; pages are
// requested with page=1,2,... and the loop stops once page >= total_pages. This
// shape (records under a per-stream key, page count in the body) does not match a
// connsdk Paginator exactly, so the loop lives here on top of connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, account string, endpoint streamEndpoint, pageSize, maxPages int, startDate string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("per_page", strconv.Itoa(pageSize))
	if startDate != "" {
		base.Set("start_date", startDate)
	}
	path := accountPath(account, endpoint.resource)

	for page := 1; maxPages == 0 || page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))
		resp, err := r.Do(ctx, http.MethodGet, path, query, nil)
		if err != nil {
			return fmt.Errorf("read callrail %s: %w", endpoint.arrayKey, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.arrayKey)
		if err != nil {
			return fmt.Errorf("decode callrail %s page: %w", endpoint.arrayKey, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		totalPages, err := connsdk.StringAt(resp.Body, "total_pages")
		if err != nil {
			return fmt.Errorf("decode callrail %s total_pages: %w", endpoint.arrayKey, err)
		}
		total, convErr := strconv.Atoi(strings.TrimSpace(totalPages))
		// Stop when we have reached the last page, when the count is unknown, or
		// when a page came back empty (defensive against off-by-one totals).
		if convErr != nil || total <= page || len(records) == 0 {
			return nil
		}
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise callrail credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                    fmt.Sprintf("%s_fixture_%d", strings.TrimSuffix(endpoint.resource, ".json"), i),
			"answered":              true,
			"direction":             "inbound",
			"duration":              int64(30 * i),
			"start_time":            callrailFixtureTime,
			"created_at":            callrailFixtureTime,
			"last_message_at":       callrailFixtureTime,
			"customer_name":         fmt.Sprintf("Fixture %d", i),
			"customer_phone_number": fmt.Sprintf("+1555000000%d", i),
			"tracking_phone_number": "+15555550000",
			"name":                  fmt.Sprintf("Fixture %d", i),
			"email":                 fmt.Sprintf("fixture+%d@example.com", i),
			"company_id":            "CMP_fixture_1",
			"status":                "active",
			"state":                 "active",
			"role":                  "admin",
			"time_zone":             "America/New_York",
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

// requester builds a connsdk.Requester wired with CallRail Token-header auth and
// the resolved base URL. The api_key only ever flows into the Authorization
// header value; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := callrailBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	apiKey := strings.TrimSpace(callrailAPIKey(cfg))
	if apiKey == "" {
		return nil, errors.New("callrail connector requires secret api_key")
	}
	// CallRail expects: Authorization: Token token="API_KEY".
	authValue := fmt.Sprintf("Token token=%q", apiKey)
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.APIKeyHeader("Authorization", authValue, ""),
		UserAgent: callrailUserAgent,
	}, nil
}

// startDateParam derives the start_date query value (YYYY-MM-DD) from the
// incremental cursor (if any, taking its date component) or else the start_date
// config. An empty result means no lower bound.
func startDateParam(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return dateOnly(cursor), nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		// Accept full RFC3339 too, narrowing to the date component.
		if _, rerr := time.Parse(time.RFC3339, startDate); rerr != nil {
			return "", fmt.Errorf("callrail config start_date must be YYYY-MM-DD: %w", err)
		}
		return dateOnly(startDate), nil
	}
	return startDate, nil
}

func dateOnly(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 10 {
		return value[:10]
	}
	return value
}

func callrailAPIKey(cfg connectors.RuntimeConfig) string {
	if v := secretOrConfig(cfg, "api_key"); v != "" {
		return v
	}
	return ""
}

// callrailAccountID resolves the required account id from Secrets or Config.
func callrailAccountID(cfg connectors.RuntimeConfig) (string, error) {
	if v := secretOrConfig(cfg, "account_id"); v != "" {
		return v, nil
	}
	return "", errors.New("callrail connector requires account_id (config or secret)")
}

// secretOrConfig returns a value preferring Secrets, falling back to Config.
func secretOrConfig(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets != nil {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config[key]); v != "" {
			return v
		}
	}
	return ""
}

// accountPath builds the account-scoped resource path, e.g. a/123/calls.json.
func accountPath(account, resource string) string {
	return "a/" + url.PathEscape(account) + "/" + resource
}

// callrailBaseURL resolves and validates the base URL. The default is
// api.callrail.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func callrailBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return callrailDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("callrail config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("callrail config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("callrail config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func callrailPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return callrailDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("callrail config page_size must be an integer: %w", err)
	}
	if value < 1 || value > callrailMaxPageSize {
		return 0, fmt.Errorf("callrail config page_size must be between 1 and %d", callrailMaxPageSize)
	}
	return value, nil
}

func callrailMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("callrail config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("callrail config max_pages must be 0 for unlimited or a positive integer")
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

// Write is unsupported: CallRail is a read-only analytics source here.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
