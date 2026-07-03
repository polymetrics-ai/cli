// Package zohobigin implements the zoho-bigin bundle's AuthHook
// (docs/migration/quarantine.json's zoho-bigin AUTH_COMPLEX entry): an
// OAuth 2.0 refresh-token-grant connsdk.Authenticator, porting legacy
// internal/connectors/zoho-bigin/zoho_bigin.go's refreshToken almost
// verbatim. Only one hook interface is implemented (AuthHook), well under
// the 2-interface cap and well under the 300-line Tier-2 soft target.
//
// This hook COPIES the gmail hook's pattern (internal/connectors/hooks/
// gmail/hooks.go) per the migration brief: https-only token_url, in-memory
// token caching until 60s before declared expiry, and secret-free errors.
// Secret values (client_secret, the refresh token, cached access tokens)
// flow ONLY into the outgoing token-request form or the Authorization
// header; they are never logged and never appear in an error string.
package zohobigin

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
	engine.RegisterHooks("zoho-bigin", func() engine.Hooks { return New() })
}

// Hooks is the zoho-bigin hook set. It implements engine.AuthHook only.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now.
	Now func() time.Time
	// Client overrides the HTTP client used for the token exchange; nil uses
	// a default client with a 30s timeout (mirrors legacy's inline default).
	Client *http.Client
}

// New returns a fresh zoho-bigin Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "zoho-bigin" }

// Authenticator resolves the OAuth2 refresh-token-grant connsdk.Authenticator
// for spec (mode "custom", hook "zoho-bigin"). Templated AuthSpec fields are
// interpolated against cfg here — buildCustomAuth passes spec through
// uninterpolated (engine/auth.go), since interpolation is mode-specific
// engine-side.
//
// spec.Token is interpreted as the refresh token (mirrors gmail's
// documented field mapping): zoho-bigin's AuthSpec has no dedicated
// "refresh_token" field, and Token is otherwise unused by the custom mode.
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error) {
	tokenURL, err := interpolateRequired(spec.TokenURL, "token_url", cfg)
	if err != nil {
		return nil, err
	}
	if err := validateHTTPSURL(tokenURL, "token_url"); err != nil {
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

	return &oauthRefreshAuth{
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
// resolved value is also rejected (mirrors legacy's requireOAuth check on
// client_id/client_secret/client_refresh_token).
func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return "", fmt.Errorf("zoho-bigin oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("zoho-bigin oauth: %s is required", field)
	}
	return val, nil
}

func authVars(cfg connectors.RuntimeConfig) engine.Vars {
	return engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// validateHTTPSURL fails closed on anything but a well-formed https:// URL
// with a host — an attacker-controlled token_url override could otherwise
// exfiltrate client_secret/the refresh token to an arbitrary endpoint. This
// is intentionally stricter than legacy's validateURL (zoho_bigin.go:233-241),
// which also accepted plain http; the tighter rule is documented as a parity
// deviation in docs.md's Known limits (mirrors gmail's identical deviation;
// never stricter for any real Zoho OAuth endpoint, which is always https).
func validateHTTPSURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("zoho-bigin oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("zoho-bigin oauth: %s must use https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("zoho-bigin oauth: %s must include a host", field)
	}
	return nil
}

// oauthRefreshAuth implements connsdk.Authenticator for the Zoho OAuth 2.0
// refresh-token grant, mirroring legacy zoho_bigin.go's refreshToken
// field-for-field and behavior-for-behavior: exchange the refresh token for
// a short-lived access token at tokenURL, cache it until 60s before its
// declared expiry, then set Authorization: Zoho-oauthtoken <token> on each
// request (Zoho's own header scheme, distinct from a plain Bearer prefix).
// Secret values never flow anywhere except the outgoing token-request form
// or the Authorization header; they are never logged.
type oauthRefreshAuth struct {
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
	req.Header.Set("Authorization", "Zoho-oauthtoken "+token)
	return nil
}

func (a *oauthRefreshAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races (mirrors gmail's hook).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("zoho-bigin oauth: token URL is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("zoho-bigin oauth: refresh_token is required")
	}
	if strings.TrimSpace(a.clientID) == "" {
		return "", errors.New("zoho-bigin oauth: client_id is required")
	}
	if strings.TrimSpace(a.clientSecret) == "" {
		return "", errors.New("zoho-bigin oauth: client_secret is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("refresh_token", a.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("zoho-bigin oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("zoho-bigin oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("zoho-bigin oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("zoho-bigin oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("zoho-bigin oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}
