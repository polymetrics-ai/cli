// Package blogger implements the blogger bundle's AuthHook: an OAuth 2.0
// refresh-token-grant connsdk.Authenticator, porting legacy
// internal/connectors/blogger/blogger.go's refreshTokenAuth almost verbatim
// (~130 lines, well under the 300-line Tier-2 hook cap, conventions.md §1).
// Only one hook interface is implemented (AuthHook), well under the
// 2-interface cap.
//
// The 3-legged OAuth consent/acquisition dance is out of scope: this hook
// only exchanges an already-issued refresh token for short-lived access
// tokens. Secret values (client_secret, the refresh token, cached access
// tokens) flow ONLY into the outgoing token-request form or the
// Authorization header; they are never logged and never appear in an error
// string (THREAT-MODEL.md).
package blogger

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
	engine.RegisterHooks("blogger", func() engine.Hooks { return New() })
}

// Hooks is the blogger hook set. It implements engine.AuthHook only.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now (mirrors legacy
	// refreshTokenAuth.now).
	Now func() time.Time
	// Client overrides the HTTP client used for the token exchange; nil uses
	// a default client with a 30s timeout (mirrors legacy's httpClient()
	// pattern in bing-ads/gmail's ports).
	Client *http.Client
}

// New returns a fresh blogger Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "blogger" }

// Authenticator resolves the OAuth2 refresh-token-grant connsdk.Authenticator
// for spec (mode "custom", hook "blogger"). Templated AuthSpec fields
// (token_url/client_id/client_secret/token) are interpolated against cfg
// here — buildCustomAuth passes spec through uninterpolated (engine/
// auth.go), since interpolation is mode-specific engine-side.
//
// spec.Token is interpreted as the refresh token (matching gmail's identical
// field-mapping convention): blogger's AuthSpec has no dedicated
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
// resolved value is also rejected (mirrors legacy's
// strings.TrimSpace(...) == "" checks, blogger.go:289).
func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return "", fmt.Errorf("blogger oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("blogger oauth: %s is required", field)
	}
	return val, nil
}

func authVars(cfg connectors.RuntimeConfig) engine.Vars {
	return engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// validateHTTPSURL fails closed on anything but a well-formed https:// URL
// with a host: token_url is the one SSRF-adjacent surface this hook
// introduces — an attacker-controlled token_url override could otherwise
// exfiltrate client_secret/the refresh token to an arbitrary endpoint. This
// is intentionally stricter than legacy's bloggerTokenURL (blogger.go:
// 435-443), which performed no scheme validation on an override at all; the
// tighter rule is documented as a parity deviation in docs.md's Known
// limits (never stricter for any real Google OAuth endpoint, which is
// always https).
func validateHTTPSURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("blogger oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("blogger oauth: %s must use https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("blogger oauth: %s must include a host", field)
	}
	return nil
}

// oauthRefreshAuth implements connsdk.Authenticator for the Google OAuth 2.0
// refresh-token grant, mirroring legacy blogger.go's refreshTokenAuth
// field-for-field and behavior-for-behavior: exchange the refresh token for
// a short-lived access token at tokenURL, cache it until 60s before its
// declared expiry, then set Authorization: Bearer <token> on each request.
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
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *oauthRefreshAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Reuse a cached token until 60s before expiry to avoid edge races
	// (legacy blogger.go:338-341).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("blogger oauth: token URL is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("blogger oauth: refresh_token is required")
	}
	if strings.TrimSpace(a.clientID) == "" {
		return "", errors.New("blogger oauth: client_id is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("refresh_token", a.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("blogger oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("blogger oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("blogger oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("blogger oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("blogger oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}
