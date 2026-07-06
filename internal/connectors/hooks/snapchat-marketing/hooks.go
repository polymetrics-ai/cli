// Package snapchatmarketing implements the snapchat-marketing bundle's
// AuthHook (fresh migration, capability parity with
// internal/connectors/snapchat-marketing/auth.go's refreshTokenAuth): an
// OAuth 2.0 refresh-token-grant connsdk.Authenticator, porting legacy's
// refreshTokenAuth almost verbatim (~115 lines, well under the 300-line
// Tier-2 hook cap, conventions.md §1). Only one hook interface is
// implemented (AuthHook), well under the 2-interface cap.
//
// This is the identical shape to internal/connectors/hooks/strava's and
// hooks/gmail's pilot AuthHooks: the engine's built-in
// oauth2_client_credentials auth mode only performs a client-credentials
// grant, never grant_type=refresh_token, so a refresh-token exchange
// genuinely needs a Tier-2 hook (token-exchange auth is a listed legitimate
// Tier-2 trigger, conventions.md §1).
//
// Secret values (client_secret, the refresh token, cached access tokens)
// flow ONLY into the outgoing token-request form or the Authorization
// header; they are never logged and never appear in an error string
// (THREAT-MODEL.md Delta 2).
package snapchatmarketing

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
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("snapchat-marketing", func() engine.Hooks { return New() })
}

// Hooks is the snapchat-marketing hook set. It implements engine.AuthHook only.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now.
	Now func() time.Time
	// Client overrides the HTTP client used for the token exchange; nil uses
	// a default client with a 30s timeout (mirrors legacy's inline default,
	// auth.go:80-82).
	Client *http.Client
}

// New returns a fresh snapchat-marketing Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "snapchat-marketing" }

// Authenticator resolves the OAuth2 refresh-token-grant connsdk.Authenticator
// for spec (mode "custom", hook "snapchat-marketing"). Templated AuthSpec
// fields (token_url/client_id/client_secret/token) are interpolated against
// cfg here — buildCustomAuth passes spec through uninterpolated
// (engine/auth.go), since interpolation is mode-specific engine-side.
//
// spec.Token is interpreted as the refresh token (matching strava/gmail's
// identical convention): the AuthSpec has no dedicated "refresh_token"
// field, and Token is otherwise unused by the custom mode.
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error) {
	tokenURL, err := interpolateRequired(spec.TokenURL, "token_url", cfg)
	if err != nil {
		return nil, err
	}
	if err := validateHTTPURL(tokenURL, "token_url"); err != nil {
		return nil, err
	}

	clientID, err := interpolateRequired(spec.ClientID, "client_id", cfg)
	if err != nil {
		return nil, err
	}
	clientSecret, err := interpolateRequired(spec.ClientSecret, "client_secret", cfg)
	if err != nil {
		return nil, err
	}
	refreshToken, err := interpolateRequired(spec.Token, "refresh_token", cfg)
	if err != nil {
		return nil, err
	}

	return &refreshTokenAuth{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		client:       h.Client,
		now:          h.Now,
	}, nil
}

// interpolateRequired resolves tmpl via engine.Interpolate and wraps any
// error naming field for a caller-facing, secret-free message. An empty
// resolved value is also rejected (mirrors legacy's
// strings.TrimSpace(...) == "" checks, auth.go via requireSecrets).
func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return "", fmt.Errorf("snapchat-marketing oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("snapchat-marketing oauth: %s is required", field)
	}
	return val, nil
}

func authVars(cfg connectors.RuntimeConfig) engine.Vars {
	return engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// validateHTTPURL fails closed on anything but a well-formed http(s):// URL
// with a host, matching legacy's validateHTTPURL
// (snapchat_marketing.go:381-393), which accepts plain http (for local test
// servers) as well as https.
func validateHTTPURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("snapchat-marketing oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("snapchat-marketing oauth: %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("snapchat-marketing oauth: %s must include a host", field)
	}
	return nil
}

// refreshTokenAuth implements connsdk.Authenticator for Snapchat's OAuth 2.0
// refresh-token grant, mirroring legacy auth.go's refreshTokenAuth
// field-for-field and behavior-for-behavior: exchange client_id/
// client_secret/refresh_token for a short-lived bearer access token at
// tokenURL, cache it until 60s before its declared expiry (falling back to a
// 1-hour TTL when the response has no expires_in, matching legacy's
// auth.go:104-108 exactly), then set Authorization: Bearer <token> on each
// request. Secret values never flow anywhere except the outgoing
// token-request form or the Authorization header; they are never logged.
type refreshTokenAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	client       *http.Client

	// now is injectable for tests; defaults to time.Now.
	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *refreshTokenAuth) timeNow() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *refreshTokenAuth) httpClient() *http.Client {
	if a.client != nil {
		return a.client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Apply ensures a fresh access token and sets the Authorization header.
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
	// Refresh 60s before expiry to avoid edge races (legacy auth.go:58-60).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("snapchat-marketing oauth: token URL is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("snapchat-marketing oauth: refresh_token is required")
	}
	if strings.TrimSpace(a.clientID) == "" {
		return "", errors.New("snapchat-marketing oauth: client_id is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("refresh_token", a.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("snapchat-marketing oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("snapchat-marketing oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("snapchat-marketing oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("snapchat-marketing oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("snapchat-marketing oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	// Legacy default TTL is 1 hour when expires_in doesn't parse
	// (auth.go:105), matching gmail's default (not strava's 6-hour default).
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}
