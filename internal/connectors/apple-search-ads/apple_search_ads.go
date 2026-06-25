// Package applesearchads implements the native pm Apple Search Ads connector. It
// follows the declarative-HTTP shape established by the stripe connector: a thin
// package that composes the connsdk toolkit (Requester + extraction helpers)
// with Apple-Search-Ads-specific stream definitions, endpoints, and an OAuth2
// client-credentials authenticator.
//
// Apple Search Ads has two traits versus a vanilla bearer API:
//   - Auth is an OAuth2 client_credentials grant: client_id/client_secret are
//     exchanged at Apple's token endpoint for a short-lived access token, which
//     is sent as Authorization: Bearer.
//   - Every Campaign Management API request is scoped to an organization via the
//     X-AP-Context: orgId=<org_id> header.
//
// Pagination uses Apple's {data:[...], pagination:{totalResults,startIndex,
// itemsPerPage}} envelope with offset/limit. Campaigns are listed with GET and
// query params; ad groups, keywords, and ads are read org-wide via the POST
// .../find endpoints whose selector body carries {pagination:{offset,limit}}.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. The registry key is the bare
// system name "apple-search-ads" even though the Go package identifier is
// applesearchads. The connector is read-only (Capabilities.Write=false): Apple's
// write surface is not exposed as reverse-ETL.
package applesearchads

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	registryName = "apple-search-ads"

	defaultBaseURL  = "https://api.searchads.apple.com/api/v5"
	defaultTokenURL = "https://appleid.apple.com/auth/oauth2/token?grant_type=client_credentials&scope=searchadsorg"
	defaultScope    = "searchadsorg"

	defaultPageSize = 1000
	maxPageSize     = 1000
	userAgent       = "polymetrics-go-cli"
)

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Apple Search Ads connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Apple Search Ads connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Apple Ads",
		IntegrationType: "api",
		Description:     "Reads Apple Search Ads campaigns, ad groups, targeting keywords, and ads via the Apple Search Ads Campaign Management API using an OAuth2 client-credentials grant scoped to an organization. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Apple Search
// Ads. In fixture mode it short-circuits without a network call.
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
	if _, err := orgID(cfg); err != nil {
		return err
	}
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the campaigns list confirms auth and connectivity (it
	// forces the token exchange) without mutating anything.
	query := url.Values{"offset": []string{"0"}, "limit": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "campaigns", query, nil, nil); err != nil {
		return fmt.Errorf("check apple-search-ads: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: appleStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Apple Search Ads is
// read-only in this connector (Capabilities.Write=false): the Apple write
// surface is not exposed as reverse-ETL, so any write request is rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, errors.New("apple-search-ads connector is read-only; writes are not supported")
}

// InitialState satisfies connectors.StatefulReader. The supported sync mode is
// full_refresh, so the initial state is just the stream marker.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]string{"stream": stream}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "campaigns"
	}
	endpoint, ok := appleStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("apple-search-ads stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	pageSize, err := pageSizeFor(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := maxPagesFor(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Apple Search Ads offset pagination over the {data, pagination}
// envelope. Campaigns paginate with offset/limit query params; the .../find
// streams paginate with a {pagination:{offset,limit}} selector body. The loop
// stops when a short page is returned, when the running total reaches
// pagination.totalResults, or when maxPages is reached.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	offset := 0
	seen := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		var (
			resp *connsdk.Response
			err  error
		)
		if endpoint.usesFind {
			body := map[string]any{
				"pagination": map[string]any{
					"offset": offset,
					"limit":  pageSize,
				},
			}
			resp, err = r.Do(ctx, http.MethodPost, endpoint.resource, nil, body)
		} else {
			query := url.Values{}
			query.Set("offset", strconv.Itoa(offset))
			query.Set("limit", strconv.Itoa(pageSize))
			resp, err = r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		}
		if err != nil {
			return fmt.Errorf("read apple-search-ads %s: %w", endpoint.resource, err)
		}

		records, err := connsdk.RecordsAt(resp.Body, "data")
		if err != nil {
			return fmt.Errorf("decode apple-search-ads %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		seen += len(records)

		// Stop on a short page (fewer than requested) — there is no further data.
		if len(records) < pageSize || len(records) == 0 {
			return nil
		}
		// Stop once we have collected the reported total, when present.
		if total, ok := totalResults(resp.Body); ok && seen >= total {
			return nil
		}
		offset += pageSize
	}
	return nil
}

// totalResults reads pagination.totalResults from a response body, if present.
func totalResults(body []byte) (int, bool) {
	raw, err := connsdk.StringAt(body, "pagination.totalResults")
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0, false
	}
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, false
	}
	return v, true
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise apple-search-ads credential-free (mirrors
// stripe's fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := int64(1000 + i)
		item := map[string]any{
			"id":               id,
			"orgId":            int64(123456),
			"campaignId":       int64(1000),
			"adGroupId":        int64(2000),
			"creativeId":       int64(3000),
			"name":             fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"status":           "ENABLED",
			"servingStatus":    "RUNNING",
			"displayStatus":    "RUNNING",
			"adChannelType":    "SEARCH",
			"billingEvent":     "TAPS",
			"pricingModel":     "CPC",
			"creativeType":     "DEFAULT",
			"matchType":        "EXACT",
			"text":             fmt.Sprintf("fixture keyword %d", i),
			"creationTime":     "2026-01-01T00:00:00.000Z",
			"modificationTime": fmt.Sprintf("2026-01-0%dT00:00:00.000Z", i),
			"startTime":        "2026-01-01T00:00:00.000Z",
			"endTime":          "",
			"deleted":          false,
			"dailyBudgetAmount": map[string]any{
				"amount":   fmt.Sprintf("%d.00", 100*i),
				"currency": "USD",
			},
			"budgetAmount": map[string]any{
				"amount":   fmt.Sprintf("%d.00", 1000*i),
				"currency": "USD",
			},
			"defaultBidAmount": map[string]any{
				"amount":   fmt.Sprintf("%d.50", i),
				"currency": "USD",
			},
			"bidAmount": map[string]any{
				"amount":   fmt.Sprintf("%d.25", i),
				"currency": "USD",
			},
			"countriesOrRegions": []any{"US"},
			"supplySources":      []any{"APPSTORE_SEARCH_RESULTS"},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the OAuth2 client-credentials
// authenticator, the resolved base URL, and the X-AP-Context org header. Secrets
// only ever flow into the authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := baseURL(cfg)
	if err != nil {
		return nil, err
	}
	org, err := orgID(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}

	auth := &clientCredentialsAuth{
		tokenURL:     tokenURL(cfg),
		clientID:     secret(cfg, "client_id"),
		clientSecret: secret(cfg, "client_secret"),
		scope:        defaultScope,
		client:       c.Client,
	}

	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: userAgent,
		DefaultHeaders: map[string]string{
			"X-AP-Context": "orgId=" + org,
		},
	}, nil
}

func requireSecrets(cfg connectors.RuntimeConfig) error {
	for _, field := range []string{"client_id", "client_secret"} {
		if strings.TrimSpace(secret(cfg, field)) == "" {
			return fmt.Errorf("apple-search-ads connector requires secret %s", field)
		}
	}
	return nil
}

func secret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[key]
}

// orgID resolves and validates the organization id (sent in X-AP-Context). It is
// required for any non-fixture read.
func orgID(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return "", errors.New("apple-search-ads config org_id is required")
	}
	org := strings.TrimSpace(cfg.Config["org_id"])
	if org == "" {
		return "", errors.New("apple-search-ads config org_id is required")
	}
	if _, err := strconv.ParseInt(org, 10, 64); err != nil {
		return "", fmt.Errorf("apple-search-ads config org_id must be an integer: %w", err)
	}
	return org, nil
}

// tokenURL resolves the OAuth2 token endpoint, honoring the token_refresh_endpoint
// override (the catalog config field) and otherwise using Apple's default.
func tokenURL(cfg connectors.RuntimeConfig) string {
	if cfg.Config != nil {
		if override := strings.TrimSpace(cfg.Config["token_refresh_endpoint"]); override != "" {
			return override
		}
	}
	return defaultTokenURL
}

// baseURL resolves and validates the base URL. The default is
// api.searchads.apple.com; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func baseURL(cfg connectors.RuntimeConfig) (string, error) {
	if cfg.Config == nil {
		return defaultBaseURL, nil
	}
	override := strings.TrimSpace(cfg.Config["base_url"])
	if override == "" {
		return defaultBaseURL, nil
	}
	parsed, err := url.Parse(override)
	if err != nil {
		return "", fmt.Errorf("apple-search-ads config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("apple-search-ads config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("apple-search-ads config base_url must include a host")
	}
	return strings.TrimRight(override, "/"), nil
}

func pageSizeFor(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return defaultPageSize, nil
	}
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("apple-search-ads config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("apple-search-ads config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPagesFor(cfg connectors.RuntimeConfig) (int, error) {
	if cfg.Config == nil {
		return 0, nil
	}
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("apple-search-ads config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("apple-search-ads config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
