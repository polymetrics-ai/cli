// Package lob implements the native pm Lob connector. It is a declarative-HTTP
// per-system connector built on the same shape as the stripe reference: a thin
// package that composes the connsdk toolkit (Requester + Basic auth +
// RecordsAt extraction + cursor state) with Lob-specific stream definitions,
// endpoints, and pagination.
//
// Lob is read-only (the upstream Airbyte source supports full_refresh only and
// the print/mail API has no safe reverse-ETL surface), so Capabilities.Write is
// false. It self-registers with the connectors registry via RegisterFactory in
// init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package lob

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
	lobDefaultBaseURL  = "https://api.lob.com/v1"
	lobDefaultPageSize = 50
	lobMaxPageSize     = 100
	lobUserAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory("lob", New)
}

// New returns the Lob connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Lob connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "lob" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "lob",
		DisplayName:     "Lob",
		IntegrationType: "api",
		Description:     "Reads Lob addresses, postcards, letters, checks, and bank accounts through the Lob print & mail REST API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Lob. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := lobBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(lobSecret(cfg)) == "" {
		return errors.New("lob connector requires secret api_key")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the addresses list confirms auth and connectivity
	// without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "addresses", url.Values{"limit": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check lob: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: lobStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Lob stream starts with an
// empty incremental cursor (full sync). Lob's list pagination is opaque
// (after-token), so this is a placeholder for stateful chaining at the harness
// level rather than an API-level filter.
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
		stream = "addresses"
	}
	endpoint, ok := lobStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("lob stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := lobPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := lobMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Lob's cursor pagination. Lob list responses are shaped
// {data:[...], next_url:"/<resource>?...&after=<token>"}; the next page is
// requested with after=<token> extracted from next_url. There is no body-token
// paginator in connsdk for the next_url shape, so the loop lives here, built on
// connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))

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
			return fmt.Errorf("read lob %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode lob %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		nextURL, err := connsdk.StringAt(resp.Body, "next_url")
		if err != nil {
			return fmt.Errorf("decode lob %s next_url: %w", endpoint.resource, err)
		}
		next := afterCursor(nextURL)
		if next == "" || len(records) == 0 {
			return nil
		}
		after = next
	}
	return nil
}

// afterCursor extracts the `after` query parameter from a Lob next_url value
// (e.g. "/postcards?limit=2&after=psc_2"). An empty/absent next_url or missing
// after token returns "" to signal pagination is exhausted.
func afterCursor(nextURL string) string {
	nextURL = strings.TrimSpace(nextURL)
	if nextURL == "" || strings.EqualFold(nextURL, "null") {
		return ""
	}
	parsed, err := url.Parse(nextURL)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Query().Get("after"))
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise lob credential-free (mirrors the stripe
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":            fmt.Sprintf("%s_fixture_%d", endpoint.resource, i),
			"object":        strings.TrimSuffix(stream, "s"),
			"description":   fmt.Sprintf("Fixture %s %d", stream, i),
			"name":          fmt.Sprintf("Fixture %d", i),
			"company":       "Example Co",
			"email":         fmt.Sprintf("fixture+%d@example.com", i),
			"phone":         "5555550100",
			"address_line1": "210 King St",
			"address_city":  "San Francisco",
			"address_state": "CA",
			"address_zip":   "94107",

			"address_country":        "US",
			"url":                    "https://lob-assets.example.com/fixture.pdf",
			"carrier":                "USPS",
			"status":                 "processed",
			"expected_delivery_date": "2026-01-10",
			"routing_number":         "322271627",
			"account_number":         "123456789",
			"account_type":           "company",
			"signatory":              "John Doe",
			"bank_name":              "Example Bank",
			"verified":               true,
			"date_created":           fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"date_modified":          fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"deleted":                false,
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

// requester builds a connsdk.Requester wired with Basic auth (API key as the
// username, blank password), the resolved base URL, and a user agent. The secret
// only ever flows into connsdk.Basic; it is never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := lobBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := lobSecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("lob connector requires secret api_key")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Basic(secret, ""),
		UserAgent: lobUserAgent,
	}, nil
}

func lobSecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["api_key"]
}

// lobBaseURL resolves and validates the base URL. The default is api.lob.com;
// any override must be an absolute https (or http for local test servers) URL
// with a host to bound SSRF risk.
func lobBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return lobDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("lob config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("lob config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("lob config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func lobPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		// Lob's Airbyte source exposes this as `limit`; accept either key.
		raw = strings.TrimSpace(cfg.Config["limit"])
	}
	if raw == "" {
		return lobDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("lob config page_size must be an integer: %w", err)
	}
	if value < 1 || value > lobMaxPageSize {
		return 0, fmt.Errorf("lob config page_size must be between 1 and %d", lobMaxPageSize)
	}
	return value, nil
}

func lobMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("lob config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("lob config max_pages must be 0 for unlimited or a positive integer")
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

// Write satisfies the connectors.Connector interface. Lob is read-only here, so
// every write is rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
