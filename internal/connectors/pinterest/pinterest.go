// Package pinterest implements the native pm Pinterest connector. It follows the
// declarative-HTTP template established by the stripe package: a thin package
// that composes the connsdk toolkit (Requester + RecordsAt extraction) with
// Pinterest-specific stream definitions, endpoints, and an OAuth2 refresh-token
// authenticator.
//
// Pinterest API v5 (https://api.pinterest.com/v5) authenticates with OAuth2.
// Open-source clients hold a client_id, client_secret, and long-lived
// refresh_token; the connector exchanges the refresh token for a short-lived
// access token at /oauth/token (refresh_token grant, HTTP Basic client auth)
// and sends that access token as a Bearer header on data requests. List
// endpoints share the shape {"items":[...],"bookmark":"..."} and are paginated
// by passing ?bookmark=<token> until the bookmark is empty.
//
// Like stripe, it self-registers with the connectors registry via
// RegisterFactory in init(); the registryset package blank-imports this package
// in the production binary to run that side effect.
//
// This connector is read-only: the upstream upstream source supports only
// full_refresh syncs and no reverse ETL, so Capabilities.Write is false and
// there is no write.go.
package pinterest

import (
	"context"
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
	pinterestDefaultBaseURL = "https://api.pinterest.com/v5"
	pinterestTokenPath      = "oauth/token"
	pinterestUserAgent      = "polymetrics-go-cli"
	pinterestDefaultMaxPage = 0 // 0 = unlimited
)

func init() {
	connectors.RegisterFactory("pinterest", New)
}

// New returns the Pinterest connector as a connectors.Connector.
func New() connectors.Connector { return Connector{} }

// Connector is the native pm Pinterest connector.
type Connector struct {
	// Client overrides the HTTP client used by the underlying connsdk Requester
	// and the token exchange. Left nil in production; injectable for tests.
	Client *http.Client
}

func (Connector) Name() string { return "pinterest" }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{
		Name:            "pinterest",
		DisplayName:     "Pinterest",
		IntegrationType: "api",
		Description:     "Reads Pinterest ad accounts, boards, campaigns, ad groups, and audiences through the Pinterest API v5 (OAuth2 refresh-token auth). Read-only.",
		Capabilities:    connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false},
	}
}

// Check verifies the connector is configured well enough to talk to Pinterest.
// In fixture mode it short-circuits without a network call.
func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := pinterestBaseURL(cfg); err != nil {
		return err
	}
	creds, err := pinterestCredentials(cfg)
	if err != nil {
		return err
	}
	r, err := c.requester(cfg, creds)
	if err != nil {
		return err
	}
	// A bounded read of the ad_accounts list confirms the OAuth exchange and
	// connectivity without mutating anything.
	if err := r.DoJSON(ctx, http.MethodGet, "ad_accounts", url.Values{"page_size": []string{"1"}}, nil, nil); err != nil {
		return fmt.Errorf("check pinterest: %w", err)
	}
	return nil
}

// Write is unsupported: this connector is read-only. It satisfies the
// connectors.Connector interface while reporting that writes are not available.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: pinterestStreams()}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "ad_accounts"
	}
	endpoint, ok := pinterestStreamEndpoints[stream]
	if !ok {
		return fmt.Errorf("pinterest stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, endpoint, req, emit)
	}

	resourcePath, err := resolveResourcePath(endpoint, req.Config)
	if err != nil {
		return err
	}
	creds, err := pinterestCredentials(req.Config)
	if err != nil {
		return err
	}
	r, err := c.requester(req.Config, creds)
	if err != nil {
		return err
	}
	pageSize, err := pinterestPageSize(req.Config)
	if err != nil {
		return err
	}
	maxPages, err := pinterestMaxPages(req.Config)
	if err != nil {
		return err
	}
	return c.harvest(ctx, r, resourcePath, pageSize, maxPages, endpoint.mapRecord, emit)
}

// harvest drives Pinterest's bookmark-cursor pagination. List responses are
// {"items":[...],"bookmark":"..."}; the next page is requested with
// bookmark=<token>. There is no body-token paginator in connsdk that reads the
// bookmark from this exact shape and stops on null, so the loop lives here,
// built on connsdk.Requester + connsdk.RecordsAt + connsdk.StringAt.
func (c Connector) harvest(ctx context.Context, r *connsdk.Requester, resource string, pageSize, maxPages int, mapRecord func(map[string]any) connectors.Record, emit func(connectors.Record) error) error {
	bookmark := ""
	for page := 0; maxPages == 0 || page < maxPages; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := url.Values{}
		if pageSize > 0 {
			query.Set("page_size", fmt.Sprintf("%d", pageSize))
		}
		if bookmark != "" {
			query.Set("bookmark", bookmark)
		}
		resp, err := r.Do(ctx, http.MethodGet, resource, query, nil)
		if err != nil {
			return fmt.Errorf("read pinterest %s: %w", resource, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "items")
		if err != nil {
			return fmt.Errorf("decode pinterest %s page: %w", resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "bookmark")
		if err != nil {
			return fmt.Errorf("decode pinterest %s bookmark: %w", resource, err)
		}
		next = strings.TrimSpace(next)
		if next == "" || next == "null" {
			return nil
		}
		bookmark = next
	}
	return nil
}

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise pinterest credential-free (mirrors stripe's
// fixture intent).
func (c Connector) readFixture(ctx context.Context, stream string, endpoint streamEndpoint, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	accountID := strings.TrimSpace(req.Config.Config["account_id"])
	if accountID == "" {
		accountID = "act_fixture"
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		item := map[string]any{
			"id":                fmt.Sprintf("%s_fixture_%d", stream, i),
			"ad_account_id":     accountID,
			"campaign_id":       "camp_fixture_1",
			"name":              fmt.Sprintf("Fixture %s %d", stream, i),
			"description":       "fixture record",
			"status":            "ACTIVE",
			"privacy":           "PUBLIC",
			"objective_type":    "AWARENESS",
			"audience_type":     "CUSTOMER_LIST",
			"country":           "US",
			"currency":          "USD",
			"owner":             map[string]any{"username": "fixture_user"},
			"pin_count":         int64(10 * i),
			"follower_count":    int64(5 * i),
			"size":              int64(100 * i),
			"created_at":        "2026-01-01T00:00:00",
			"created_time":      int64(1767225600 + i),
			"updated_time":      int64(1767225600 + i),
			"created_timestamp": int64(1767225600 + i),
		}
		record := endpoint.mapRecord(item)
		record["connector"] = "pinterest"
		record["fixture"] = true
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

// requester builds a connsdk.Requester wired with the OAuth2 refresh-token
// authenticator and the resolved base URL. The secrets only ever flow into the
// authenticator; they are never logged.
func (c Connector) requester(cfg connectors.RuntimeConfig, creds pinterestCreds) (*connsdk.Requester, error) {
	base, err := pinterestBaseURL(cfg)
	if err != nil {
		return nil, err
	}
	return &connsdk.Requester{
		Client:    c.Client,
		BaseURL:   base,
		Auth:      newRefreshTokenAuth(base, creds, c.Client),
		UserAgent: pinterestUserAgent,
	}, nil
}

// resolveResourcePath substitutes the configured account id into account-scoped
// resource templates and validates it is present.
func resolveResourcePath(endpoint streamEndpoint, cfg connectors.RuntimeConfig) (string, error) {
	if !endpoint.accountScoped {
		return endpoint.resource, nil
	}
	accountID := strings.TrimSpace(cfg.Config["account_id"])
	if accountID == "" {
		return "", errors.New("pinterest account-scoped stream requires config account_id")
	}
	return fmt.Sprintf(endpoint.resource, url.PathEscape(accountID)), nil
}

// pinterestCreds holds the OAuth credentials resolved from Secrets. They are
// never logged or printed.
type pinterestCreds struct {
	clientID     string
	clientSecret string
	refreshToken string
}

// pinterestCredentials resolves the three OAuth secrets from cfg.Secrets. The
// catalog secret keys are dotted ("credentials.client_id"); flat fallbacks are
// also accepted for convenience.
func pinterestCredentials(cfg connectors.RuntimeConfig) (pinterestCreds, error) {
	get := func(dotted, flat string) string {
		if cfg.Secrets == nil {
			return ""
		}
		if v := strings.TrimSpace(cfg.Secrets[dotted]); v != "" {
			return v
		}
		return strings.TrimSpace(cfg.Secrets[flat])
	}
	creds := pinterestCreds{
		clientID:     get("credentials.client_id", "client_id"),
		clientSecret: get("credentials.client_secret", "client_secret"),
		refreshToken: get("credentials.refresh_token", "refresh_token"),
	}
	var missing []string
	if creds.clientID == "" {
		missing = append(missing, "credentials.client_id")
	}
	if creds.clientSecret == "" {
		missing = append(missing, "credentials.client_secret")
	}
	if creds.refreshToken == "" {
		missing = append(missing, "credentials.refresh_token")
	}
	if len(missing) > 0 {
		return pinterestCreds{}, fmt.Errorf("pinterest connector requires secrets: %s", strings.Join(missing, ", "))
	}
	return creds, nil
}

// refreshTokenAuth implements connsdk.Authenticator using the OAuth2
// refresh-token grant. It exchanges the refresh token for an access token at
// the token endpoint (HTTP Basic client auth) and caches it until shortly
// before expiry. Secrets are never logged.
type refreshTokenAuth struct {
	tokenURL string
	creds    pinterestCreds
	client   *http.Client
	now      func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func newRefreshTokenAuth(baseURL string, creds pinterestCreds, client *http.Client) *refreshTokenAuth {
	return &refreshTokenAuth{
		tokenURL: strings.TrimRight(baseURL, "/") + "/" + pinterestTokenPath,
		creds:    creds,
		client:   client,
		now:      time.Now,
	}
}

func (a *refreshTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *refreshTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && a.now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.creds.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("pinterest oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.SetBasicAuth(a.creds.clientID, a.creds.clientSecret)

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("pinterest oauth: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("pinterest oauth: token endpoint returned %s", resp.Status)
	}

	token, ttl, err := decodeTokenResponse(resp)
	if err != nil {
		return "", err
	}
	a.token = token
	a.expires = a.now().Add(ttl)
	return a.token, nil
}

// pinterestBaseURL resolves and validates the base URL. The default is
// api.pinterest.com; any override must be an absolute https (or http for local
// test servers) URL with a host to bound SSRF risk.
func pinterestBaseURL(cfg connectors.RuntimeConfig) (string, error) {
	base := strings.TrimSpace(cfg.Config["base_url"])
	if base == "" {
		return pinterestDefaultBaseURL, nil
	}
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("pinterest config base_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("pinterest config base_url must use http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return "", errors.New("pinterest config base_url must include a host")
	}
	return strings.TrimRight(base, "/"), nil
}

func pinterestPageSize(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return 0, nil // let API default apply
	}
	value, err := parsePositiveInt(raw)
	if err != nil {
		return 0, fmt.Errorf("pinterest config page_size must be a positive integer: %w", err)
	}
	if value > 250 {
		return 0, errors.New("pinterest config page_size must be between 1 and 250")
	}
	return value, nil
}

func pinterestMaxPages(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return pinterestDefaultMaxPage, nil
	}
	value, err := parsePositiveInt(raw)
	if err != nil {
		return 0, fmt.Errorf("pinterest config max_pages must be a positive integer, all, or unlimited: %w", err)
	}
	return value, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	if cfg.Config == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
