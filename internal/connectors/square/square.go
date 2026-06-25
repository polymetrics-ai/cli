// Package square implements the native pm Square connector. It is a
// declarative-HTTP per-system connector built on the connsdk toolkit (Requester
// + Bearer auth + RecordsAt extraction + cursor pagination) with Square-specific
// stream definitions and endpoints. It mirrors the stripe reference connector.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// Square is read-only here: the upstream Airbyte source supports full_refresh
// only and there is no obviously safe reverse-ETL write surface, so
// Capabilities.Write is false.
package square

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
	squareDefaultBaseURL  = "https://connect.squareup.com/v2"
	squareSandboxBaseURL  = "https://connect.squareupsandbox.com/v2"
	squareDefaultPageSize = 100
	squareMaxPageSize     = 100
	squareUserAgent       = "polymetrics-go-cli"
	// squareAPIVersion pins the Square-Version header. Square is a dated API; a
	// pinned version keeps the response shape stable.
	squareAPIVersion = "2024-01-18"
	// squareFixtureCreated is the deterministic created_at used by fixture-mode
	// records (2026-01-01T00:00:00Z).
	squareFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("square", New)
}

// New returns the Square connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Square connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "square" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "square",
		DisplayName:     "Square",
		IntegrationType: "api",
		Description:     "Reads Square payments, refunds, customers, and locations through the Square Connect v2 REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Square. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := squareBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(squareSecret(cfg)) == "" {
		return errors.New("square connector requires secret credentials.api_key (or an OAuth access token)")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of locations confirms auth and connectivity without
	// mutating anything; locations is a small, always-available list.
	if err := r.DoJSON(ctx, http.MethodGet, "locations", nil, nil, nil); err != nil {
		return fmt.Errorf("check square: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. Square is read-only in
// this connector (Capabilities.Write is false), so any write is rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: squareStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Square stream starts with
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
		stream = "payments"
	}
	endpoint, ok := squareStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("square stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	lower, err := incrementalLowerBound(req)
	if err != nil {
		return err
	}
	pageSize, err := squarePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := squareMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives Square's cursor pagination. Square list endpoints return
// {<arrayKey>:[...], cursor:"..."}; the next page is requested with
// cursor=<token>. An empty/absent cursor ends the walk. The connsdk
// CursorPaginator assumes one fixed array path, so this small in-package loop
// keeps the array key per-stream while reusing connsdk.Requester + RecordsAt +
// StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, beginTime string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if endpoint.timeParam != "" && beginTime != "" {
		base.Set(endpoint.timeParam, beginTime)
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
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read square %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.arrayKey)
		if err != nil {
			return fmt.Errorf("decode square %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "cursor")
		if err != nil {
			return fmt.Errorf("decode square %s cursor: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise square credential-free (mirrors stripe).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":              fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"created_at":      squareFixtureCreated,
			"updated_at":      squareFixtureCreated,
			"status":          "COMPLETED",
			"source_type":     "CARD",
			"location_id":     "L_fixture_1",
			"order_id":        "O_fixture_1",
			"payment_id":      "pay_fixture_1",
			"receipt_number":  fmt.Sprintf("R%d", i),
			"reason":          "fixture refund",
			"amount_money":    map[string]any{"amount": int64(1000 * i), "currency": "USD"},
			"total_money":     map[string]any{"amount": int64(1000 * i), "currency": "USD"},
			"processing_fee":  []any{},
			"given_name":      fmt.Sprintf("Fixture%d", i),
			"family_name":     "User",
			"email_address":   fmt.Sprintf("fixture+%d@example.com", i),
			"phone_number":    "+15555550100",
			"company_name":    "Fixture Co",
			"reference_id":    fmt.Sprintf("ref_%d", i),
			"creation_source": "THIRD_PARTY",
			"name":            fmt.Sprintf("Fixture Location %d", i),
			"country":         "US",
			"currency":        "USD",
			"type":            "PHYSICAL",
			"merchant_id":     "M_fixture_1",
			"timezone":        "UTC",
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

// requester builds a connsdk.Requester wired with Bearer auth, the resolved base
// URL, and the Square-Version header. The secret only ever flows into
// connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := squareBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := squareSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("square connector requires secret credentials.api_key (or an OAuth access token)")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: squareUserAgent,
		DefaultHeaders: map[string]string{
			"Square-Version": squareAPIVersion,
		},
	}, nil
}

// incrementalLowerBound returns the RFC3339 lower bound for the time filter,
// derived from the incremental cursor (if any) or else the start_date config
// (a YYYY-MM-DD date, per the Square catalog). An empty result means no lower
// bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	// Accept both a bare date (YYYY-MM-DD) and a full RFC3339 timestamp.
	if t, err := time.Parse("2006-01-02", startDate); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	if t, err := time.Parse(time.RFC3339, startDate); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	return "", fmt.Errorf("square config start_date must be YYYY-MM-DD or RFC3339, got %q", startDate)
}

// squareSecret resolves the Square access token. The Square catalog flattens the
// authentication oneOf into dotted secret keys; the API-key flow uses
// credentials.api_key, the OAuth flow ultimately yields an access token. We
// accept the first non-empty of the known keys so either auth shape works as a
// raw bearer token.
func squareSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	for _, key := range []string{
		"credentials.api_key",
		"api_key",
		"access_token",
		"credentials.access_token",
	} {
		if v := strings.TrimSpace(cfg.Secrets[key]); v != "" {
			return v
		}
	}
	return ""
}

// squareBaseURL resolves and validates the base URL. The default is the Square
// production host; is_sandbox=true selects the sandbox host. Any explicit
// base_url override must be an absolute https (or http for local test servers)
// URL with a host to bound SSRF risk.
func squareBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		if isSandbox(cfg) {
			return squareSandboxBaseURL, nil
		}
		return squareDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("square config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("square config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("square config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func isSandbox(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(cfg.Config["is_sandbox"])) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

func squarePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return squareDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("square config page_size must be an integer: %w", err)
	}
	if value < 1 || value > squareMaxPageSize {
		return 0, fmt.Errorf("square config page_size must be between 1 and %d", squareMaxPageSize)
	}
	return value, nil
}

func squareMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("square config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("square config max_pages must be 0 for unlimited or a positive integer")
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
