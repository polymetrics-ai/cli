// Package amazonads implements the native pm Amazon Ads connector. It follows
// the declarative-HTTP shape established by the stripe connector: a thin package
// that composes the connsdk toolkit (Requester + extraction helpers) with
// Amazon-Ads-specific stream definitions, endpoints, and a Login with Amazon
// refresh_token authenticator.
//
// Amazon Ads has two unusual traits versus a vanilla bearer API:
//   - Auth is a Login with Amazon OAuth2 refresh_token grant: the long-lived
//     refresh_token plus client_id/client_secret are exchanged for a 1h access
//     token, which is sent as Authorization: Bearer. The client_id is ALSO sent
//     verbatim in the Amazon-Advertising-API-ClientId header.
//   - Most entity endpoints are scoped to a single advertising profile via the
//     Amazon-Advertising-API-Scope header; the profiles endpoint itself is not.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect. The registry key is the bare
// system name "amazon-ads" even though the Go package identifier is amazonads.
package amazonads

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
	registryName       = "amazon-ads"
	defaultRegion      = "NA"
	defaultPageSize    = 100
	maxPageSize        = 100
	amazonAdsUserAgent = "polymetrics-go-cli"
)

// regionBaseURL maps an Amazon Ads region to its API base URL.
var regionBaseURL = map[string]string{
	"NA": "https://advertising-api.amazon.com",
	"EU": "https://advertising-api-eu.amazon.com",
	"FE": "https://advertising-api-fe.amazon.com",
}

// regionTokenURL maps an Amazon Ads region to its Login with Amazon token URL.
var regionTokenURL = map[string]string{
	"NA": "https://api.amazon.com/auth/o2/token",
	"EU": "https://api.amazon.co.uk/auth/o2/token",
	"FE": "https://api.amazon.co.jp/auth/o2/token",
}

func init() {
	connectors.RegisterFactory(registryName, New)
}

// New returns the Amazon Ads connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Amazon Ads connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return registryName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            registryName,
		DisplayName:     "Amazon Ads",
		IntegrationType: "api",
		Description:     "Reads Amazon Advertising profiles, Sponsored Products campaigns, ad groups, keywords, and portfolios via the Amazon Ads API using a Login with Amazon refresh-token grant. Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Amazon Ads.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := amazonAdsBaseURL(cfg); err != nil {
		return err
	}
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg, false)
	if err != nil {
		return err
	}
	// A bounded read of the profiles list confirms auth and connectivity
	// (it forces the token exchange) without mutating anything.
	query := url.Values{"startIndex": []string{"0"}, "count": []string{"1"}}
	if err := r.DoJSON(ctx, http.MethodGet, "v2/profiles", query, nil, nil); err != nil {
		return fmt.Errorf("check amazon-ads: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: amazonAdsStreams()}, nil
}

// Write satisfies the connectors.Connector interface. Amazon Ads is read-only
// in this connector (Capabilities.Write=false): the Amazon Ads write surface is
// not exposed as reverse-ETL, so any write request is rejected.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsFailed: len(records)}, errors.New("amazon-ads connector is read-only; writes are not supported")
}

// InitialState satisfies connectors.StatefulReader. Amazon Ads v2 entity
// endpoints are full-refresh, so the initial state is just the stream marker.
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
		stream = "profiles"
	}
	endpoint, ok := amazonAdsStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("amazon-ads stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, emit)
	}

	r, err := c.requester(req.Config, endpoint.scoped)
	if err != nil {
		return err
	}
	pageSize, err := amazonAdsPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := amazonAdsMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, endpoint, pageSize, maxPages, emit)
}

// harvest drives Amazon Ads v2 offset pagination. The v2 entity endpoints accept
// startIndex (offset) and count (page size) and return a top-level JSON array.
// A page shorter than count signals the end. This loop lives in-package because
// the records live at the JSON root (recordsPath "") rather than under a key.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, pageSize, maxPages int, emit func(connectors.Record) error) error {
	startIndex := 0
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		query.Set("startIndex", strconv.Itoa(startIndex))
		query.Set("count", strconv.Itoa(pageSize))

		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read amazon-ads %s: %w", endpoint.resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "")
		if err != nil {
			return fmt.Errorf("decode amazon-ads %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		if len(records) < pageSize {
			return nil
		}
		startIndex += pageSize
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise amazon-ads credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := int64(1000 + i)
		item := map[string]any{
			"profileId":            id,
			"campaignId":           id,
			"adGroupId":            id,
			"portfolioId":          id,
			"keywordId":            id,
			"name":                 fmt.Sprintf("%s fixture %d", strings.TrimSuffix(stream, "s"), i),
			"campaignType":         "sponsoredProducts",
			"targetingType":        "manual",
			"state":                "enabled",
			"dailyBudget":          float64(10 * i),
			"defaultBid":           float64(i),
			"bid":                  float64(i),
			"startDate":            "20260101",
			"endDate":              "",
			"premiumBidAdjustment": true,
			"inBudget":             true,
			"countryCode":          "US",
			"currencyCode":         "USD",
			"timezone":             "America/Los_Angeles",
			"keywordText":          fmt.Sprintf("fixture keyword %d", i),
			"matchType":            "exact",
			"accountInfo": map[string]any{
				"marketplaceStringId": "ATVPDKIKX0DER",
				"type":                "seller",
				"name":                fmt.Sprintf("Fixture Account %d", i),
				"id":                  fmt.Sprintf("ENTITY%d", i),
			},
		}
		if err := emit(endpoint.mapRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the Login with Amazon
// refresh_token authenticator, the resolved base URL, the ClientId header, and
// (for scoped streams) the profile scope header. Secrets only ever flow into the
// authenticator and headers; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig, scoped bool) (*connsdk.Requester, error) {
	base, err := amazonAdsBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}
	clientID := secret(cfg, "client_id")

	headers := map[string]string{
		"Amazon-Advertising-API-ClientId": clientID,
	}
	if scoped {
		profileID := strings.TrimSpace(cfg.Config["profile_id"])
		if profileID == "" {
			return nil, errors.New("amazon-ads config profile_id is required for profile-scoped streams")
		}
		headers["Amazon-Advertising-API-Scope"] = profileID
	}

	auth := &refreshTokenAuth{
		tokenURL:     amazonAdsTokenURL(cfg),
		clientID:     clientID,
		clientSecret: secret(cfg, "client_secret"),
		refreshToken: secret(cfg, "refresh_token"),
		client:       c.Client,
	}

	return &connsdk.Requester{
		Client:         c.Client,
		BaseURL:        base,
		Auth:           auth,
		UserAgent:      amazonAdsUserAgent,
		DefaultHeaders: headers,
	}, nil
}

func requireSecrets(cfg connectors.RuntimeConfig) error {
	for _, field := range []string{"client_id", "client_secret", "refresh_token"} {
		if strings.TrimSpace(secret(cfg, field)) == "" {
			return fmt.Errorf("amazon-ads connector requires secret %s", field)
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

// region returns the validated region code (NA/EU/FE), defaulting to NA.
func region(cfg connectors.RuntimeConfig) string {
	r := strings.ToUpper(strings.TrimSpace(cfg.Config["region"]))
	if r == "" {
		return defaultRegion
	}
	if _, ok := regionBaseURL[r]; ok {
		return r
	}
	return defaultRegion
}

// amazonAdsBaseURL resolves and validates the base URL. The default is derived
// from the configured region; any explicit override must be an absolute https
// (or http for local test servers) URL with a host to bound SSRF risk.
func amazonAdsBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	override := strings.TrimSpace(cfg.Config["base_url"])
	if override == "" {
		return regionBaseURL[region(cfg)], nil
	}
	parsed, err := url.Parse(override)
	if err != nil {
		return "", fmt.Errorf("amazon-ads config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("amazon-ads config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("amazon-ads config base_url must include a host")
	}
	return strings.TrimRight(override, "/"), nil
}

// amazonAdsTokenURL resolves the Login with Amazon token endpoint, honoring a
// token_url override (used by tests) and otherwise deriving it from the region.
func amazonAdsTokenURL(cfg connectors.RuntimeConfig) string {
	if override := strings.TrimSpace(cfg.Config["token_url"]); override != "" {
		return override
	}
	return regionTokenURL[region(cfg)]
}

func amazonAdsPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("amazon-ads config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("amazon-ads config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func amazonAdsMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("amazon-ads config max_pages must be an integer, all, or unlimited: %w", err)
	}
	if value < 0 {
		return 0, errors.New("amazon-ads config max_pages must be 0 for unlimited or a positive integer")
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
