// Package mantle implements the native pm Mantle connector. Mantle
// (api.heymantle.com) is a billing/usage platform for Shopify apps; this
// connector reads its customers and subscriptions streams over the declarative
// HTTP toolkit.
//
// It follows the stripe reference shape: a thin package that composes the connsdk
// toolkit (Requester + Bearer auth + RecordsAt extraction + cursor state) with
// Mantle-specific stream definitions, endpoints, and pagination. Mantle paginates
// with a top-level cursor token plus a hasNextPage boolean rather than Stripe's
// has_more/starting_after, so the page loop lives in this package.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package mantle

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
	mantleDefaultBaseURL  = "https://api.heymantle.com"
	mantleDefaultPageSize = 500
	mantleMaxPageSize     = 10000
	mantleUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("mantle", New)
}

// New returns the Mantle connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Mantle connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "mantle" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "mantle",
		DisplayName:     "Mantle",
		IntegrationType: "api",
		Description:     "Reads Mantle customers and subscriptions through the Mantle (heymantle.com) REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Mantle. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := mantleBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(mantleSecret(cfg)) == "" {
		return errors.New("mantle connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the customers list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "v1/customers", url.Values{"take": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check mantle: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: mantleStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Mantle stream starts with an
// empty incremental cursor (full sync), which the start_date config can raise at
// read time.
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
	endpoint, ok := mantleStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("mantle stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := mantlePageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := mantleMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Mantle's cursor pagination. List endpoints return
// {<selector>:[...], cursor:"<token>", hasNextPage:bool}; the next page is
// requested with cursor=<token>. connsdk's CursorPaginator does not support the
// hasNextPage stop-condition shape, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("take", strconv.Itoa(pageSize))

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
			return fmt.Errorf("read mantle %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.selector)
		if err != nil {
			return fmt.Errorf("decode mantle %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasNext, err := connsdk.StringAt(resp.Body, "hasNextPage")
		if err != nil {
			return fmt.Errorf("decode mantle %s hasNextPage: %w", endpoint.resource, err)
		}
		nextCursor, err := connsdk.StringAt(resp.Body, "cursor")
		if err != nil {
			return fmt.Errorf("decode mantle %s cursor: %w", endpoint.resource, err)
		}
		if hasNext != "true" || strings.TrimSpace(nextCursor) == "" {
			return nil
		}
		cursor = nextCursor
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise mantle credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                    fmt.Sprintf("%s_fixture_%d", stream, i),
			"name":                  fmt.Sprintf("Fixture %d", i),
			"email":                 fmt.Sprintf("fixture+%d@example.com", i),
			"domain":                "example.com",
			"industry":              "software",
			"countryCode":           "US",
			"test":                  true,
			"active":                true,
			"total":                 float64(99 * i),
			"subtotal":              float64(90 * i),
			"last30Revenue":         float64(100 * i),
			"lifetimeValue":         float64(1000 * i),
			"averageMonthlyRevenue": float64(80 * i),
			"createdAt":             "2026-01-01T00:00:00.000Z",
			"updatedAt":             "2026-01-01T00:00:00+00:00",
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
	base, err := mantleBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := mantleSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("mantle connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: mantleUserAgent,
	}, nil
}

func mantleSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// mantleBaseURL resolves and validates the base URL. The default is
// api.heymantle.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func mantleBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return mantleDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("mantle config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("mantle config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("mantle config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func mantlePageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return mantleDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mantle config page_size must be an integer: %w", err)
	}
	if value < 1 || value > mantleMaxPageSize {
		return 0, fmt.Errorf("mantle config page_size must be between 1 and %d", mantleMaxPageSize)
	}
	return value, nil
}

func mantleMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("mantle config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("mantle config max_pages must be 0 for unlimited or a positive integer")
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

// Write is unsupported: the Mantle source connector is read-only.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
