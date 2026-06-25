// Package brex implements the native pm Brex connector. It is a declarative-HTTP
// per-system connector built on the same shape as the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + Bearer auth +
// RecordsAt extraction + cursor pagination) with Brex-specific stream
// definitions, endpoints, and record mappers.
//
// Brex is a financial API (cards, transactions, expenses) with no safe
// reverse-ETL writes, so this connector is read-only (Capabilities.Write =
// false).
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package brex

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	brexDefaultBaseURL  = "https://platform.brexapis.com"
	brexDefaultPageSize = 100
	brexMaxPageSize     = 100
	brexUserAgent       = "polymetrics-go-cli"
	// brexFixtureDate is the deterministic date stamp used by fixture-mode
	// records.
	brexFixtureDate = "2026-01-01"
)

func init() {
	connectors.RegisterFactory("brex", New)
}

// New returns the Brex connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Brex connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "brex" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "brex",
		DisplayName:     "Brex",
		IntegrationType: "api",
		Description:     "Reads Brex transactions, users, expenses, vendors, and budgets through the Brex platform REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Brex. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := brexBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(brexSecret(cfg)) == "" {
		return errors.New("brex connector requires secret user_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the users list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "/v2/users", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check brex: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: brexStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Brex stream starts with an
// empty incremental cursor (full sync), which the start_date config can raise at
// read time for streams that support a datetime cursor.
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
		stream = "transactions"
	}
	endpoint, ok := brexStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("brex stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := brexPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := brexMaxPages(req.Config)
	if err != nil {
		return err
	}
	startParam, startValue := incrementalStart(stream, req)
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, startParam, startValue, emit)
}

// harvest drives Brex's cursor pagination. Brex lists return
// {items:[...], next_cursor:"..."}; the next page is requested with
// cursor=<next_cursor>, stopping when next_cursor is null/empty. The loop lives
// here, built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, startParam, startValue string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if startParam != "" && startValue != "" {
		base.Set(startParam, startValue)
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
		resp, err := r.Do(ctx, http.MethodGet, endpoint.path, query, nil)
		if err != nil {
			return fmt.Errorf("read brex %s: %w", endpoint.path, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode brex %s page: %w", endpoint.path, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "next_cursor")
		if err != nil {
			return fmt.Errorf("decode brex %s next_cursor: %w", endpoint.path, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise brex credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	pkField := "id"
	if stream == "budgets" {
		pkField = "budget_id"
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			pkField:           fmt.Sprintf("%s_fixture_%d", stream, i),
			"card_id":         "card_fixture_1",
			"description":     fmt.Sprintf("Fixture %s %d", stream, i),
			"amount":          map[string]any{"amount": 1000 * i, "currency": "USD"},
			"original_amount": map[string]any{"amount": 1000 * i, "currency": "USD"},
			"limit":           map[string]any{"amount": 100000, "currency": "USD"},
			"first_name":      "Fixture",
			"last_name":       fmt.Sprintf("User%d", i),
			"email":           fmt.Sprintf("fixture+%d@example.com", i),
			"company_name":    fmt.Sprintf("Vendor %d", i),
			"name":            fmt.Sprintf("Budget %d", i),
			"status":          "ACTIVE",
			"category":        "SOFTWARE",
			"type":            "PURCHASE",
			"period_type":     "MONTHLY",
			"purchased_at":    brexFixtureDate + "T00:00:00Z",
			"updated_at":      brexFixtureDate + "T00:00:00Z",
			"posted_at_date":  brexFixtureDate,
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "brex"
		record["fixture"] = true
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
	base, err := brexBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := brexSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("brex connector requires secret user_token")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: brexUserAgent,
	}, nil
}

// incrementalStart returns the Brex request parameter and value used to bound a
// stream's start time. Only transactions and expenses support a datetime cursor;
// other streams return ("", ""). The value comes from the incremental cursor (if
// any) or else the start_date config.
func incrementalStart(stream string, req connectors.ReadRequest) (string, string) {
	var param string
	switch stream {
	case "transactions":
		param = "posted_at_start"
	case "expenses":
		param = "purchased_at_start"
	default:
		return "", ""
	}
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return param, cursor
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return param, ""
	}
	// Normalize RFC3339 to the date/datetime form Brex expects; pass through on
	// parse failure so a caller-provided value is still honored.
	if t, err := time.Parse(time.RFC3339, startDate); err == nil {
		return param, t.UTC().Format("2006-01-02T15:04:05")
	}
	return param, startDate
}

func brexSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["user_token"]
}

// brexBaseURL resolves and validates the base URL. The default is
// platform.brexapis.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func brexBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return brexDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("brex config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("brex config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("brex config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func brexPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return brexDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("brex config page_size must be an integer: %w", err)
	}
	if value < 1 || value > brexMaxPageSize {
		return 0, fmt.Errorf("brex config page_size must be between 1 and %d", brexMaxPageSize)
	}
	return value, nil
}

func brexMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("brex config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("brex config max_pages must be 0 for unlimited or a positive integer")
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

// Write is unsupported: Brex is a read-only financial source in pm.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
