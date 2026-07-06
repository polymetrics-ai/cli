package bingads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	// defaultCustomerBaseURL is the production Customer Management REST root.
	defaultCustomerBaseURL = "https://clientcenter.api.bingads.microsoft.com/CustomerManagement/v13"
	// defaultCampaignBaseURL is the production Campaign Management REST root.
	defaultCampaignBaseURL = "https://campaign.api.bingads.microsoft.com/CampaignManagement/v13"
	// defaultTokenURLTemplate is the Microsoft identity platform token endpoint;
	// %s is the tenant id (default "common").
	defaultTokenURLTemplate = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
	// defaultScope is the OAuth scope required for Microsoft Advertising access.
	defaultScope = "https://ads.microsoft.com/msads.manage offline_access"

	userAgent       = "polymetrics-go-cli"
	defaultTenantID = "common"
)

// HTTPDoer is the minimal HTTP client surface this package depends on,
// satisfied by *http.Client. Overridable for tests.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// serviceKind selects which Bing Ads REST service (and therefore which base
// URL and header set) a stream targets. Mirrors legacy streams.go's
// serviceKind.
type serviceKind int

const (
	serviceCustomer serviceKind = iota
	serviceCampaign
)

// fixtureMode reports whether cfg requests the credential-free fixture path
// (mirrors legacy's identically named helper).
func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func secret(cfg connectors.RuntimeConfig, name string) string {
	if cfg.Secrets == nil {
		return ""
	}
	return cfg.Secrets[name]
}

// validateSecrets enforces the required credential set. Values are never
// logged (mirrors legacy bing_ads.go:252-259).
func validateSecrets(cfg connectors.RuntimeConfig) error {
	for _, name := range []string{"developer_token", "client_id", "refresh_token"} {
		if strings.TrimSpace(secret(cfg, name)) == "" {
			return fmt.Errorf("bing-ads connector requires secret %s", name)
		}
	}
	return nil
}

// customerBaseURL/campaignBaseURL resolve and validate the two REST service
// base URLs (mirrors legacy bing_ads.go:295-303).
func customerBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return resolveBaseURL(cfg, "base_url", defaultCustomerBaseURL)
}

func campaignBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	return resolveBaseURL(cfg, "campaign_base_url", defaultCampaignBaseURL)
}

// resolveBaseURL reads an override from cfg.Config[key] (or returns def) and
// validates scheme+host to bound SSRF risk (mirrors legacy, bing_ads.go:
// 305-323).
func resolveBaseURL(cfg connectors.RuntimeConfig, key, def string) (string, error) {
	base := strings.TrimSpace(cfg.Config[key])
	if base == "" {
		return def, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("bing-ads config %s is invalid: %w", key, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("bing-ads config %s must use http or https, got %q", key, parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("bing-ads config %s must include a host", key)
	}
	return strings.TrimRight(base, "/"), nil
}

// tokenURL resolves the OAuth token endpoint. A token_url override wins;
// otherwise the Microsoft identity platform endpoint for the configured
// tenant is used. Overrides are scheme/host validated (mirrors legacy,
// bing_ads.go:325-350).
func tokenURL(cfg connectors.RuntimeConfig) (string, error) {
	if override := strings.TrimSpace(cfg.Config["token_url"]); override != "" {
		parsed, err := url.Parse(override)
		if err != nil {
			return "", fmt.Errorf("bing-ads config token_url is invalid: %w", err)
		}
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return "", fmt.Errorf("bing-ads config token_url must use http or https, got %q", parsed.Scheme)
		}
		if parsed.Host == "" {
			return "", errors.New("bing-ads config token_url must include a host")
		}
		return override, nil
	}
	tenant := strings.TrimSpace(secret(cfg, "tenant_id"))
	if tenant == "" {
		tenant = strings.TrimSpace(cfg.Config["tenant_id"])
	}
	if tenant == "" {
		tenant = defaultTenantID
	}
	return fmt.Sprintf(defaultTokenURLTemplate, url.PathEscape(tenant)), nil
}

// customerAccountID resolves CustomerAccountId, falling back to account_id
// (mirrors legacy, bing_ads.go:288-293).
func customerAccountID(cfg connectors.RuntimeConfig) string {
	if v := strings.TrimSpace(cfg.Config["customer_account_id"]); v != "" {
		return v
	}
	return strings.TrimSpace(cfg.Config["account_id"])
}

// accountIDList parses the comma-separated account_ids config into a slice,
// falling back to a single account_id/customer_account_id when set (mirrors
// legacy's accountIDList, bing_ads.go:270-285).
func accountIDList(cfg connectors.RuntimeConfig) []string {
	raw := strings.TrimSpace(cfg.Config["account_ids"])
	if raw == "" {
		if single := customerAccountID(cfg); single != "" {
			return []string{single}
		}
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if id := strings.TrimSpace(part); id != "" {
			out = append(out, id)
		}
	}
	return out
}

// requester builds a connsdk.Requester wired with the Microsoft OAuth
// refresh-token authenticator, the resolved base URL for the stream's
// service, the DeveloperToken header, and (for campaign-scoped streams) the
// CustomerId and CustomerAccountId headers. Secrets only ever flow into the
// authenticator and headers; they are never logged (mirrors legacy
// bing_ads.go:195-249).
func (c Connector) requester(cfg connectors.RuntimeConfig, kind serviceKind) (*connsdk.Requester, error) {
	if err := validateSecrets(cfg); err != nil {
		return nil, err
	}

	var base string
	var err error
	switch kind {
	case serviceCampaign:
		base, err = campaignBaseURL(cfg)
	default:
		base, err = customerBaseURL(cfg)
	}
	if err != nil {
		return nil, err
	}

	tokenEndpoint, err := tokenURL(cfg)
	if err != nil {
		return nil, err
	}

	var httpClient *http.Client
	if c.Client != nil {
		if hc, ok := c.Client.(*http.Client); ok {
			httpClient = hc
		}
	}

	auth := &oauthRefreshAuth{
		tokenURL:     tokenEndpoint,
		clientID:     secret(cfg, "client_id"),
		clientSecret: secret(cfg, "client_secret"),
		refreshToken: secret(cfg, "refresh_token"),
		client:       httpClient,
	}

	headers := map[string]string{
		"DeveloperToken": secret(cfg, "developer_token"),
	}
	if kind == serviceCampaign {
		if v := strings.TrimSpace(cfg.Config["customer_id"]); v != "" {
			headers["CustomerId"] = v
		}
		if v := customerAccountID(cfg); v != "" {
			headers["CustomerAccountId"] = v
		}
	}

	return &connsdk.Requester{
		Client:         httpClient,
		BaseURL:        base,
		Auth:           auth,
		UserAgent:      userAgent,
		DefaultHeaders: headers,
	}, nil
}

// Check verifies the connector is configured well enough to talk to Bing
// Ads. In fixture mode it short-circuits without a network call (mirrors
// legacy bing_ads.go:78-104).
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := customerBaseURL(cfg); err != nil {
		return err
	}
	if err := validateSecrets(cfg); err != nil {
		return err
	}
	r, err := c.requester(cfg, serviceCustomer)
	if err != nil {
		return err
	}
	ep := streamRoutes["accounts"]
	if _, err := r.Do(ctx, http.MethodPost, ep.path, nil, ep.body(cfg, "")); err != nil {
		return fmt.Errorf("check bing-ads: %w", err)
	}
	return nil
}

// oauthRefreshAuth implements connsdk.Authenticator for the Microsoft
// identity platform refresh-token grant, ported field-for-field from legacy
// auth.go's identically named type.
type oauthRefreshAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	scope        string
	client       *http.Client

	// now is injectable for tests; defaults to time.Now.
	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *oauthRefreshAuth) timeNow() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *oauthRefreshAuth) httpClient() *http.Client {
	if a.client != nil {
		return a.client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Apply ensures a fresh access token and sets the Authorization header.
func (a *oauthRefreshAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *oauthRefreshAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races.
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("bing-ads oauth: token URL is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("bing-ads oauth: refresh_token is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)
	form.Set("client_id", a.clientID)
	if a.clientSecret != "" {
		form.Set("client_secret", a.clientSecret)
	}
	scope := a.scope
	if scope == "" {
		scope = defaultScope
	}
	form.Set("scope", scope)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("bing-ads oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("bing-ads oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("bing-ads oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("bing-ads oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("bing-ads oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}
