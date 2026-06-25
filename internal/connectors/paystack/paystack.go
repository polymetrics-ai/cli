// Package paystack implements the native pm Paystack connector. It follows the
// declarative-HTTP per-system connector template established by the stripe
// package: a thin package that composes the connsdk toolkit (Requester + Bearer
// auth + RecordsAt extraction + cursor state) with Paystack-specific stream
// definitions, endpoints, and pagination.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// Paystack is a read-only source here: the API has no obviously-safe reverse-ETL
// write actions for the core streams, so Capabilities.Write is false.
package paystack

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
	paystackDefaultBaseURL  = "https://api.paystack.co"
	paystackDefaultPageSize = 100
	paystackMaxPageSize     = 100
	paystackUserAgent       = "polymetrics-go-cli"
	// paystackFixtureCreated is the deterministic createdAt used by fixture-mode
	// records (2026-01-01T00:00:00Z).
	paystackFixtureCreated = "2026-01-01T00:00:00.000Z"
)

func init() {
	connectors.RegisterFactory("paystack", New)
}

// New returns the Paystack connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Paystack connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "paystack" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "paystack",
		DisplayName:     "Paystack",
		IntegrationType: "api",
		Description:     "Reads Paystack customers, transactions, subscriptions, invoices, and disputes through the Paystack REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Paystack. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := paystackBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(paystackSecret(cfg)) == "" {
		return errors.New("paystack connector requires secret secret_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the customer list confirms auth and connectivity without
	// mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "customer", url.Values{"perPage": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check paystack: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: paystackStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Paystack is a read-only
// source here; reverse-ETL writes are not exposed, so this always reports the
// operation as unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// InitialState satisfies connectors.StatefulReader: a Paystack stream starts with
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
		stream = "customers"
	}
	endpoint, ok := paystackStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("paystack stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	from, err := incrementalLowerBound(req)
	if err != nil {
		return err
	}
	pageSize, err := paystackPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := paystackMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, from, emit)
}

// harvest drives Paystack's page-number pagination. List endpoints return
// {status, message, data:[...], meta:{..., next}}; the next page number lives at
// meta.next (or null when exhausted). The loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt, because the stop
// signal is a body field rather than a Link header.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, from string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("perPage", strconv.Itoa(pageSize))
	if from != "" {
		base.Set("from", from)
	}

	page := 1
	for pageNum := 0; maxPages == 0 || pageNum < maxPages; pageNum++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		query.Set("page", strconv.Itoa(page))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read paystack %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode paystack %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		// meta.next is the next page number; empty/null/0 means no more pages.
		next, err := connsdk.StringAt(resp.Body, "meta.next")
		if err != nil {
			return fmt.Errorf("decode paystack %s meta.next: %w", endpoint.resource, err)
		}
		nextPage, ok := parseNextPage(next)
		if !ok {
			// Fall back to a short-page stop when meta.next is absent: a page
			// shorter than the requested size is the last page.
			if len(records) < pageSize {
				return nil
			}
			nextPage = page + 1
		}
		if nextPage <= page {
			return nil
		}
		page = nextPage
	}
	return nil
}

// parseNextPage interprets a meta.next value as the next page number. Paystack
// returns either an integer page number or null. A blank, "null", or non-positive
// value means there is no next page.
func parseNextPage(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "null") {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise paystack credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                int64(i),
			"customer_code":     fmt.Sprintf("CUS_fixture_%d", i),
			"subscription_code": fmt.Sprintf("SUB_fixture_%d", i),
			"request_code":      fmt.Sprintf("PRQ_fixture_%d", i),
			"reference":         fmt.Sprintf("%s_ref_%d", endpoint.resource, i),
			"email":             fmt.Sprintf("fixture+%d@example.com", i),
			"first_name":        "Fixture",
			"last_name":         fmt.Sprintf("User%d", i),
			"phone":             "+10000000000",
			"domain":            "test",
			"risk_action":       "default",
			"amount":            int64(1000 * i),
			"refund_amount":     int64(1000 * i),
			"currency":          "NGN",
			"status":            "success",
			"resolution":        "merchant-accepted",
			"category":          "general",
			"channel":           "card",
			"gateway_response":  "Successful",
			"paid":              true,
			"due_date":          paystackFixtureCreated,
			"due_at":            paystackFixtureCreated,
			"next_payment_date": paystackFixtureCreated,
			"email_token":       fmt.Sprintf("tok_fixture_%d", i),
			"paid_at":           paystackFixtureCreated,
			"createdAt":         paystackFixtureCreated,
			"updatedAt":         paystackFixtureCreated,
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
	base, err := paystackBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := paystackSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("paystack connector requires secret secret_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: paystackUserAgent,
	}, nil
}

// incrementalLowerBound returns the RFC3339 lower bound for the `from` filter,
// derived from the incremental cursor (if any) or else the start_date config.
// An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) (string, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor, nil
	}
	startDate := strings.TrimSpace(req.Config.Config["start_date"])
	if startDate == "" {
		return "", nil
	}
	if _, err := time.Parse(time.RFC3339, startDate); err != nil {
		return "", fmt.Errorf("paystack config start_date must be RFC3339: %w", err)
	}
	return startDate, nil
}

func paystackSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["secret_key"]
}

// paystackBaseURL resolves and validates the base URL. The default is
// api.paystack.co; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func paystackBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return paystackDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("paystack config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("paystack config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("paystack config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func paystackPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return paystackDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("paystack config page_size must be an integer: %w", err)
	}
	if value < 1 || value > paystackMaxPageSize {
		return 0, fmt.Errorf("paystack config page_size must be between 1 and %d", paystackMaxPageSize)
	}
	return value, nil
}

func paystackMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("paystack config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("paystack config max_pages must be 0 for unlimited or a positive integer")
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
