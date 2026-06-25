// Package snapchatmarketing implements the native pm Snapchat Marketing (Ads
// API) connector. It follows the declarative-HTTP template established by the
// stripe package: a thin package composing the connsdk toolkit (Requester +
// OAuth2 refresh-token auth + JSON extraction + cursor pagination) with
// Snapchat-specific stream definitions and endpoint routing.
//
// The Snapchat Marketing API is read-only here (no reverse-ETL write actions are
// exposed) and authenticates with an OAuth2 refresh-token grant: the configured
// client_id/client_secret/refresh_token are exchanged at the token endpoint for
// a short-lived bearer access token applied to each request.
//
// Like stripe/github, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
package snapchatmarketing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics/internal/connectors"
	"polymetrics/internal/connectors/connsdk"
)

const (
	connectorName             = "snapchat-marketing"
	snapchatDefaultBaseURL    = "https://adsapi.snapchat.com/v1"
	snapchatDefaultTokenURL   = "https://accounts.snapchat.com/login/oauth2/access_token"
	snapchatUserAgent         = "polymetrics-go-cli"
	snapchatMaxPages          = 1000
	snapchatFixtureUpdatedAt  = "2026-01-01T00:00:00.000Z"
	snapchatFixtureCreatedAt  = "2025-12-31T00:00:00.000Z"
	snapchatPagingNextLinkKey = "paging.next_link"
)

func init() {
	connectors.RegisterFactory(connectorName, New)
}

// New returns the Snapchat Marketing connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Snapchat Marketing connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the OAuth2 token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            connectorName,
		DisplayName:     "Snapchat Marketing",
		IntegrationType: "api",
		Description:     "Reads Snapchat Marketing (Ads API) organizations, ad accounts, campaigns, ad squads, and ads via the OAuth2 refresh-token grant.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to the Snapchat
// Ads API. In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := snapchatBaseURL(cfg); err != nil {
		return err
	}
	if _, err := snapchatTokenURL(cfg); err != nil {
		return err
	}
	if err := requireSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg)
	if err != nil {
		return err
	}
	// A bounded read of the organizations list confirms auth (token exchange)
	// and connectivity without mutating anything.
	if _, err := r.Do(ctx, http.MethodGet, "organizations", url.Values{"limit": []string{"1"}}, nil); err != nil {
		return fmt.Errorf("check snapchat-marketing: %w", err)
	}
	return nil
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: snapchatStreams()}, nil
}

// InitialState satisfies connectors.StatefulReader: a stream starts with an empty
// incremental cursor (full refresh), which start_date config can raise later.
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
		stream = "organizations"
	}
	spec, ok := snapchatStreamSpecs[stream]
	if !ok {
		return fmt.Errorf("snapchat-marketing stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, spec, req, emit)
	}

	r, err := c.requester(req.Config)
	if err != nil {
		return err
	}
	paths, err := streamPaths(req.Config, spec)
	if err != nil {
		return err
	}
	for _, path := range paths {
		if err := c.harvest(ctx, r, path, spec, emit); err != nil {
			return err
		}
	}
	return nil
}

// Write is unsupported: the Snapchat Marketing connector is read-only. It exists
// to satisfy the connectors.Connector interface.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

// harvest drives Snapchat cursor pagination. List responses carry the next page
// under paging.next_link (an absolute URL); when absent the harvest stops. Each
// page holds an array under spec.arrayKey whose elements are envelopes wrapping
// the resource under spec.itemKey; unwrapEnvelopes flattens them.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, path string, spec streamSpec, emit func(connectors.Record) error) error {
	base := url.Values{}
	base.Set("limit", "50")

	nextURL := ""
	for page := 0; page < snapchatMaxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		reqPath := path
		var query url.Values
		if nextURL != "" {
			reqPath = nextURL
		} else {
			query = base
		}
		resp, err := r.Do(ctx, http.MethodGet, reqPath, query, nil)
		if err != nil {
			return fmt.Errorf("read snapchat-marketing %s: %w", spec.arrayKey, err)
		}
		records, err := unwrapEnvelopes(resp.Body, spec.arrayKey, spec.itemKey)
		if err != nil {
			return fmt.Errorf("decode snapchat-marketing %s page: %w", spec.arrayKey, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(spec.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, snapchatPagingNextLinkKey)
		if err != nil {
			return fmt.Errorf("decode snapchat-marketing %s paging: %w", spec.arrayKey, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		nextURL = next
	}
	return nil
}

// unwrapEnvelopes extracts the array at arrayKey from a Snapchat list response
// and unwraps each element's singular itemKey sub-object into a flat map. Snap
// returns e.g. {"campaigns":[{"sub_request_status":"SUCCESS","campaign":{...}}]}.
// Elements without the itemKey wrapper are passed through as-is.
func unwrapEnvelopes(body []byte, arrayKey, itemKey string) ([]map[string]any, error) {
	dec := json.NewDecoder(strings.NewReader(string(body)))
	dec.UseNumber()
	var root map[string]any
	if err := dec.Decode(&root); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	rawArr, ok := root[arrayKey].([]any)
	if !ok {
		return nil, nil
	}
	out := make([]map[string]any, 0, len(rawArr))
	for _, el := range rawArr {
		envelope, ok := el.(map[string]any)
		if !ok {
			continue
		}
		if inner, ok := envelope[itemKey].(map[string]any); ok {
			out = append(out, inner)
			continue
		}
		out = append(out, envelope)
	}
	return out, nil
}

// streamPaths resolves the list of endpoint paths to read for a stream, given
// the hierarchical Snapchat layout and configured organization/ad-account ids.
func streamPaths(cfg connectors.RuntimeConfig, spec streamSpec) ([]string, error) {
	switch spec.scope {
	case "organizations":
		return []string{"organizations"}, nil
	case "adaccounts":
		orgs := idList(cfg, "organization_ids")
		if len(orgs) == 0 {
			return nil, errors.New("snapchat-marketing stream adaccounts requires config organization_ids")
		}
		paths := make([]string, 0, len(orgs))
		for _, org := range orgs {
			paths = append(paths, "organizations/"+url.PathEscape(org)+"/adaccounts")
		}
		return paths, nil
	case "adaccount":
		accts := idList(cfg, "ad_account_ids")
		if len(accts) == 0 {
			return nil, fmt.Errorf("snapchat-marketing stream %s requires config ad_account_ids", spec.resource)
		}
		paths := make([]string, 0, len(accts))
		for _, acct := range accts {
			paths = append(paths, "adaccounts/"+url.PathEscape(acct)+"/"+spec.resource)
		}
		return paths, nil
	default:
		return nil, fmt.Errorf("snapchat-marketing unknown stream scope %q", spec.scope)
	}
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise the connector credential-free.
func (c Connector) readFixture(ctx context.Context, stream string, spec streamSpec, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                              fmt.Sprintf("%s_fixture_%d", spec.itemKey, i),
			"name":                            fmt.Sprintf("Fixture %s %d", spec.itemKey, i),
			"status":                          "ACTIVE",
			"type":                            "FIXTURE",
			"created_at":                      snapchatFixtureCreatedAt,
			"updated_at":                      snapchatFixtureUpdatedAt,
			"country":                         "US",
			"currency":                        "USD",
			"timezone":                        "America/Los_Angeles",
			"organization_id":                 "org_fixture_1",
			"ad_account_id":                   "ACC1",
			"campaign_id":                     "camp_fixture_1",
			"ad_squad_id":                     "squad_fixture_1",
			"creative_id":                     "creative_fixture_1",
			"objective":                       "WEB_CONVERSION",
			"optimization_goal":               "IMPRESSIONS",
			"billing_event":                   "IMPRESSION",
			"review_status":                   "APPROVED",
			"advertiser":                      "Fixture Advertiser",
			"address_line_1":                  "1 Fixture Way",
			"locality":                        "Santa Monica",
			"administrative_district_level_1": "CA",
			"postal_code":                     "90401",
			"start_time":                      snapchatFixtureCreatedAt,
			"end_time":                        snapchatFixtureUpdatedAt,
			"daily_budget_micro":              int64(1000000 * i),
			"lifetime_spend_cap_micro":        int64(5000000 * i),
			"bid_micro":                       int64(250000 * i),
		}
		record := spec.mapRecord(item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the OAuth2 refresh-token
// authenticator and the resolved base URL. Secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig) (*connsdk.Requester, error) {
	base, err := snapchatBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	tokenURL, err := snapchatTokenURL(cfg)
	if err != nil {
		return nil, err
	}
	if err := requireSecrets(cfg); err != nil {
		return nil, err
	}
	auth := &refreshTokenAuth{
		TokenURL:     tokenURL,
		ClientID:     secret(cfg, "client_id"),
		ClientSecret: secret(cfg, "client_secret"),
		RefreshToken: secret(cfg, "refresh_token"),
		Client:       c.Client,
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      auth,
		UserAgent: snapchatUserAgent,
	}, nil
}

func requireSecrets(cfg connectors.RuntimeConfig) error {
	for _, field := range []string{"client_id", "client_secret", "refresh_token"} {
		if strings.TrimSpace(secret(cfg, field)) == "" {
			return fmt.Errorf("snapchat-marketing connector requires secret %s", field)
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

// snapchatBaseURL resolves and validates the base URL. The default is
// adsapi.snapchat.com/v1; any override must be an absolute https (or http for
// local test servers) URL with a host to bound SSRF risk.
func snapchatBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return snapchatDefaultBaseURL, nil
	}
	if err := validateHTTPURL(base, "base_url"); err != nil {
		return "", err
	}
	return strings.TrimRight(base, "/"), nil
}

// snapchatTokenURL resolves and validates the OAuth2 token endpoint URL.
func snapchatTokenURL(cfg connectors.RuntimeConfig) (string, error) {
	tokenURL := strings.TrimSpace(cfg.Config["token_url"])
	if tokenURL == "" {
		return snapchatDefaultTokenURL, nil
	}
	if err := validateHTTPURL(tokenURL, "token_url"); err != nil {
		return "", err
	}
	return tokenURL, nil
}

func validateHTTPURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("snapchat-marketing config %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("snapchat-marketing config %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("snapchat-marketing config %s must include a host", field)
	}
	return nil
}

// idList parses a config value that is a comma-separated list of ids (e.g.
// ad_account_ids = "ACC1,ACC2"). Whitespace around ids is trimmed.
func idList(cfg connectors.RuntimeConfig, key string) []string {
	raw := strings.TrimSpace(cfg.Config[key])
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
