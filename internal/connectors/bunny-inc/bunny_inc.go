// Package bunnyinc implements the native pm Bunny, Inc. connector (catalog slug
// source-bunny-inc). Bunny is a SaaS subscription-management / billing platform
// (bunny.com) exposing a single GraphQL endpoint per tenant subdomain.
//
// It follows the stripe declarative-HTTP template: a thin package composing the
// connsdk toolkit (Requester + Bearer auth) with Bunny-specific GraphQL stream
// definitions. The one structural difference from a REST source is pagination:
// Bunny uses GraphQL cursor pagination, so the page loop POSTs a query document
// with an {after} variable and reads data.<field>.pageInfo.{endCursor,hasNextPage}
// from the response body.
//
// Like stripe, it self-registers with the connectors registry via RegisterFactory
// in init(); the registryset package blank-imports this package in the production
// binary to run that side effect.
package bunnyinc

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
	registryName         = "bunny-inc"
	bunnyDefaultPageSize = 10
	bunnyMaxPageSize     = 100
	bunnyUserAgent       = "polymetrics-go-cli"
	bunnyGraphQLPath     = "graphql"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Bunny, Inc. connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Bunny, Inc. connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester.
	// Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Bunny, Inc.",
		IntegrationType: "api",
		Description:     "Reads Bunny subscription-billing data (accounts, contacts, invoices, payments, subscriptions) from the per-tenant Bunny GraphQL API.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Bunny. In
// fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := bunnyBaseURL(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(bunnySecret(cfg)) == "" {
		return errors.New("bunny-inc connector requires secret apikey")
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A single-record accounts query confirms auth and connectivity without
	// mutating anything.
	body := graphQLBody(fmt.Sprintf(accountsQuery, 1), "")
	if _, err := r.Do(ctx, http.MethodPost, bunnyGraphQLPath, nil, body); err != nil {
		return fmt.Errorf("check bunny-inc: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: bunnyStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a Bunny stream starts with an
// empty incremental cursor (full sync). Bunny only supports full_refresh upstream,
// but the cursor slot is kept for forward compatibility.
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
		stream = "accounts"
	}
	endpoint, ok := bunnyStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("bunny-inc stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := bunnyPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := bunnyMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// Write is unsupported: bunny-inc is a read-only source connector (no reverse-ETL
// writes are exposed by the Bunny GraphQL API surface this connector targets).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Bunny's GraphQL cursor pagination. Each page POSTs the stream
// query with an {after} variable; the response carries data.<field>.nodes (the
// records) and data.<field>.pageInfo.{endCursor,hasNextPage}. The next page uses
// endCursor as `after`; the loop stops when hasNextPage is false (or absent).
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	query := fmt.Sprintf(endpoint.query, pageSize)
	nodesPath := "data." + endpoint.gqlField + ".nodes"
	endCursorPath := "data." + endpoint.gqlField + ".pageInfo.endCursor"
	hasNextPath := "data." + endpoint.gqlField + ".pageInfo.hasNextPage"

	after := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodPost, bunnyGraphQLPath, nil, graphQLBody(query, after))
		if err != nil {
			return fmt.Errorf("read bunny-inc %s: %w", endpoint.gqlField, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, nodesPath)
		if err != nil {
			return fmt.Errorf("decode bunny-inc %s page: %w", endpoint.gqlField, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		hasNext, err := connsdk.StringAt(resp.Body, hasNextPath)
		if err != nil {
			return fmt.Errorf("decode bunny-inc %s pageInfo: %w", endpoint.gqlField, err)
		}
		endCursor, err := connsdk.StringAt(resp.Body, endCursorPath)
		if err != nil {
			return fmt.Errorf("decode bunny-inc %s endCursor: %w", endpoint.gqlField, err)
		}
		if hasNext != "true" || strings.TrimSpace(endCursor) == "" {
			return nil
		}
		after = endCursor
	}
	return nil
}

// graphQLBody builds the POST body for a Bunny GraphQL request: the query document
// plus an `after` variable (omitted when empty so the first page sends no cursor).
func graphQLBody(query, after string) map[string]any {
	variables := map[string]any{}
	if after != "" {
		variables["after"] = after
	}
	return map[string]any{"query": query, "variables": variables}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise bunny-inc credential-free (mirrors stripe).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":             fmt.Sprintf("%s_fixture_%d", endpoint.gqlField, i),
			"name":           fmt.Sprintf("Fixture %d", i),
			"code":           fmt.Sprintf("FIX-%d", i),
			"email":          fmt.Sprintf("fixture+%d@example.com", i),
			"accountId":      "accounts_fixture_1",
			"currencyId":     "USD",
			"amount":         100 * i,
			"amountDue":      100 * i,
			"amountPaid":     0,
			"payingStatus":   "active",
			"firstName":      fmt.Sprintf("Fix%d", i),
			"lastName":       "Ture",
			"period":         "monthly",
			"createdAt":      "2026-01-01T00:00:00Z",
			"updatedAt":      fmt.Sprintf("2026-01-0%dT00:00:00Z", i),
			"startDate":      "2026-01-01T00:00:00Z",
			"receivedAt":     "2026-01-01T00:00:00Z",
			"portalAccess":   true,
			"netPaymentDays": 30,
			"connector":      registryName,
			"fixture":        true,
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
	base, err := bunnyBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	secret := bunnySecret(cfg)
	if strings.TrimSpace(secret) == "" {
		return nil, errors.New("bunny-inc connector requires secret apikey")
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      connsdk.Bearer(secret),
		UserAgent: bunnyUserAgent,
	}, nil
}

func bunnySecret(cfg connectors.RuntimeConfig) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets["apikey"]
}

// bunnyBaseURL resolves and validates the base URL. A base_url override (used by
// tests and self-hosted proxies) takes precedence; otherwise the host is built
// from the required `subdomain` config as https://<subdomain>.bunny.com. Any
// override must be an absolute http(s) URL with a host to bound SSRF risk.
func bunnyBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	if base := strings.TrimSpace(cfg.Config["base_url"]); base != "" {
		parsed, err := url.Parse(base)
		if err != nil {
			return "", fmt.Errorf("bunny-inc config base_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("bunny-inc config base_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("bunny-inc config base_url must include a host")
		}
		return strings.TrimRight(base, "/"), nil
	}
	subdomain := strings.TrimSpace(cfg.Config["subdomain"])
	if subdomain == "" {
		return "", errors.New("bunny-inc connector requires config subdomain (or base_url)")
	}
	if !validSubdomain(subdomain) {
		return "", fmt.Errorf("bunny-inc config subdomain %q is invalid", subdomain)
	}
	return "https://" + subdomain + ".bunny.com", nil
}

// validSubdomain bounds the subdomain to DNS label characters so it cannot be used
// to inject a different host or path into the base URL.
func validSubdomain(s string) bool {
	if len(s) == 0 || len(s) > 63 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

func bunnyPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return bunnyDefaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bunny-inc config page_size must be an integer: %w", err)
	}
	if value < 1 || value > bunnyMaxPageSize {
		return 0, fmt.Errorf("bunny-inc config page_size must be between 1 and %d", bunnyMaxPageSize)
	}
	return value, nil
}

func bunnyMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("bunny-inc config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("bunny-inc config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
