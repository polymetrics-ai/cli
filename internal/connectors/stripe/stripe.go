// Package stripe implements the native pm Stripe connector. It is the reference
// declarative-HTTP per-system connector: a thin package that composes the
// connsdk toolkit (Requester + Bearer auth + RecordsAt extraction + cursor
// state) with Stripe-specific stream definitions, endpoints, and write actions.
// Other HTTP connectors are expected to copy this shape.
//
// Like github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package stripe

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
	stripeDefaultBaseURL  = "https://api.stripe.com/v1"
	stripeDefaultPageSize = 100
	stripeMaxPageSize     = 100
	stripeUserAgent       = "polymetrics-go-cli"
	// stripeFixtureCreated is the deterministic `created` timestamp used by the
	// fixture-mode records (2026-01-01T00:00:00Z in unix seconds).
	stripeFixtureCreated int64 = 1767225600
)

func init() {
	connectors.RegisterFactory("stripe", New)
}

// New returns the Stripe connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Stripe connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "stripe" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "stripe",
		DisplayName:     "Stripe",
		IntegrationType: "api",
		Description:     "Reads Stripe customers, charges, invoices, subscriptions, and products, and writes approved reverse ETL customer actions through the Stripe REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: true},
	}
}

// Check verifies the connector is configured well enough to talk to Stripe. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := stripeBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(stripeSecret(cfg)) == "" {
		return errors.New("stripe connector requires secret client_secret")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the customers list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "customers", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check stripe: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: stripeStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Stripe stream starts with
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
	endpoint, ok := stripeStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("stripe stream %q not found", stream)
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
	pageSize, err := stripePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := stripeMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, lower, emit)
}

// harvest drives Stripe's id-cursor pagination. Stripe lists return
// {data:[...], has_more:bool}; the next page is requested with
// starting_after=<last object id>. There is no body token paginator in connsdk
// for this exact shape, so the loop lives here, built on connsdk.Requester +
// connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, createdGTE string, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if createdGTE != "" {
		base.Set("created[gte]", createdGTE)
	}

	startingAfter := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if startingAfter != "" {
			query.Set("starting_after", startingAfter)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read stripe %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode stripe %s page: %w", endpoint.resource, err)
		}
		lastID := ""
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			lastID = stringField(item, "id")
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasMore, err := connsdk.StringAt(resp.Body, "has_more")
		if err != nil {
			return fmt.Errorf("decode stripe %s has_more: %w", endpoint.resource, err)
		}
		if hasMore != "true" || lastID == "" {
			return nil
		}
		startingAfter = lastID
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise stripe credential-free (mirrors github's
// fixture intent and NativeCatalogConnector).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                   fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"object":               strings.TrimSuffix(stream, "s"),
			"created":              stripeFixtureCreated + int64(i),
			"livemode":             false,
			"currency":             "usd",
			"connector":            "stripe",
			"fixture":              true,
			"email":                fmt.Sprintf("fixture+%d@example.com", i),
			"name":                 fmt.Sprintf("Fixture %d", i),
			"status":               "active",
			"customer":             "cus_fixture_1",
			"amount":               int64(1000 * i),
			"total":                int64(1000 * i),
			"active":               true,
			"amount_due":           int64(1000 * i),
			"amount_paid":          int64(1000 * i),
			"updated":              stripeFixtureCreated + int64(i),
			"type":                 "service",
			"subscription":         "sub_fixture_1",
			"paid":                 true,
			"current_period_start": stripeFixtureCreated,
			"current_period_end":   stripeFixtureCreated + 86400,
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
// URL, and the optional Stripe-Account header. The secret only ever flows into
// connsdk.Bearer; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := stripeBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := stripeSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("stripe connector requires secret client_secret")
	}
	headers := map[string]string{}
	if account := strings.TrimSpace(cfg.Config["account_id"]); account != "" {
		headers["Stripe-Account"] = account
	}
	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           connsdk.Bearer(secret),
		UserAgent:      stripeUserAgent,
		DefaultHeaders: headers,
	}, nil
}

// incrementalLowerBound returns the unix-seconds lower bound for created[gte],
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
	t, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		return "", fmt.Errorf("stripe config start_date must be RFC3339: %w", err)
	}
	return strconv.FormatInt(t.Unix(), 10), nil
}

func stripeSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["client_secret"]
}

// stripeBaseURL resolves and validates the base URL. The default is
// api.stripe.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func stripeBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return stripeDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("stripe config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("stripe config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("stripe config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func stripePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return stripeDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("stripe config page_size must be an integer: %w", err)
	}
	if value < 1 || value > stripeMaxPageSize {
		return 0, fmt.Errorf("stripe config page_size must be between 1 and %d", stripeMaxPageSize)
	}
	return value, nil
}

func stripeMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("stripe config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("stripe config max_pages must be 0 for unlimited or a positive integer")
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

func stringField(item map[string]any, key string) string {
	switch v := item[key].(type) {
	case string:
		return v
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
