// Package gocardless implements the native pm GoCardless source connector. It is
// a declarative-HTTP per-system connector built on the same shape as the stripe
// reference: a thin package composing the connsdk toolkit (Requester + Bearer
// auth + RecordsAt extraction + cursor state) with GoCardless-specific stream
// definitions and endpoints.
//
// GoCardless is read-only here (full-refresh source, matching the upstream
// Airbyte source-gocardless): no reverse-ETL writes are exposed.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the
// production binary to run that side effect.
package gocardless

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
	gocardlessLiveBaseURL    = "https://api.gocardless.com"
	gocardlessSandboxBaseURL = "https://api-sandbox.gocardless.com"
	// gocardlessDefaultVersion is the version header used when the config omits
	// gocardless_version. It is a known-good released API version date.
	gocardlessDefaultVersion  = "2015-07-06"
	gocardlessDefaultPageSize = 50
	gocardlessMaxPageSize     = 500
	gocardlessUserAgent       = "polymetrics-go-cli"
	// gocardlessFixtureCreated is the deterministic created_at base used by the
	// fixture-mode records.
	gocardlessFixtureCreated = "2026-01-01T00:00:00Z"
)

func init() {
	connectors.RegisterFactory("gocardless", New)
}

// New returns the GoCardless connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm GoCardless connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "gocardless" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "gocardless",
		DisplayName:     "GoCardless",
		IntegrationType: "api",
		Description:     "Reads GoCardless payments, mandates, payouts, and refunds through the GoCardless REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to GoCardless.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := gocardlessBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(gocardlessSecret(cfg)) == "" {
		return errors.New("gocardless connector requires secret access_token")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the payments list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "payments", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check gocardless: %w", err)
	}
	return nil
}

// Write satisfies the connectors.Connector interface. GoCardless is a read-only
// full-refresh source here, so writes are unsupported.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: gocardlessStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a GoCardless stream starts
// with an empty incremental cursor (full sync); start_date can raise it.
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
	endpoint, ok := gocardlessStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("gocardless stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	createdGT := incrementalLowerBound(req)
	pageSize, err := gocardlessPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := gocardlessMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, createdGT, emit)
}

// harvest drives GoCardless's cursor pagination. List responses are shaped
// {"<resource>":[...], "meta":{"cursors":{"after":"<id>","before":...}}}; the
// next page is requested with after=<meta.cursors.after> until that cursor is
// null/empty. The records array is keyed by the resource name. connsdk has no
// single paginator for this exact shape (body cursor + resource-keyed array), so
// the loop lives here, built on connsdk.Requester + RecordsAt + StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, createdGT string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if createdGT != "" {
		base.Set("created_at[gt]", createdGT)
	}

	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if after != "" {
			query.Set("after", after)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read gocardless %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.resource)
		if err != nil {
			return fmt.Errorf("decode gocardless %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "meta.cursors.after")
		if err != nil {
			return fmt.Errorf("decode gocardless %s cursor: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		after = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise gocardless credential-free.
func (c Connector) readFixture(ctx context.Context, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	prefix := strings.ToUpper(endpoint.resource[:2])
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                        fmt.Sprintf("%s_fixture_%d", prefix, i),
			"created_at":                gocardlessFixtureCreated,
			"charge_date":               "2026-01-02",
			"arrival_date":              "2026-01-03",
			"amount":                    1000 * i,
			"amount_refunded":           0,
			"deducted_fees":             20 * i,
			"currency":                  "GBP",
			"status":                    "confirmed",
			"description":               fmt.Sprintf("Fixture %s %d", endpoint.resource, i),
			"reference":                 fmt.Sprintf("REF-%d", i),
			"scheme":                    "bacs",
			"payout_type":               "merchant",
			"next_possible_charge_date": "2026-01-10",
			"payments_require_approval": false,
			"connector":                 "gocardless",
			"fixture":                   true,
			"links": map[string]any{
				"mandate":               "MD_fixture_1",
				"payout":                "PO_fixture_1",
				"payment":               "PM_fixture_1",
				"creditor":              "CR_fixture_1",
				"customer_bank_account": "BA_fixture_1",
				"creditor_bank_account": "BA_fixture_2",
			},
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
// URL, and the required GoCardless-Version header. The secret only ever flows
// into connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := gocardlessBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := gocardlessSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("gocardless connector requires secret access_token")
	}
	headers := map[string]string{
		"GoCardless-Version": gocardlessVersion(cfg),
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.Bearer(secret),
		UserAgent:      gocardlessUserAgent,
		DefaultHeaders: headers,
	}, nil
}

// incrementalLowerBound returns the RFC3339 lower bound for created_at[gt],
// derived from the incremental cursor (if any) or else the start_date config.
// An empty result means no lower bound (full sync).
func incrementalLowerBound(req connectors.ReadRequest) string {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		return cursor
	}
	return strings.TrimSpace(req.Config.Config["start_date"])
}

func gocardlessSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["access_token"]
}

// gocardlessVersion resolves the GoCardless-Version header value.
func gocardlessVersion(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if v := strings.TrimSpace(cfg.Config["gocardless_version"]); v != "" {
			return v
		}
	}
	return gocardlessDefaultVersion
}

// gocardlessBaseURL resolves and validates the base URL. With no override it
// selects the live or sandbox host from gocardless_environment (default
// sandbox). Any base_url override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func gocardlessBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		switch strings.ToLower(strings.TrimSpace(cfg.Config["gocardless_environment"])) {
		case "live":
			return gocardlessLiveBaseURL, nil
		default:
			return gocardlessSandboxBaseURL, nil
		}
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("gocardless config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("gocardless config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("gocardless config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func gocardlessPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return gocardlessDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gocardless config page_size must be an integer: %w", err)
	}
	if value < 1 || value > gocardlessMaxPageSize {
		return 0, fmt.Errorf("gocardless config page_size must be between 1 and %d", gocardlessMaxPageSize)
	}
	return value, nil
}

func gocardlessMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("gocardless config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("gocardless config max_pages must be 0 for unlimited or a positive integer")
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
