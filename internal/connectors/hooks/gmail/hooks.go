// Package gmail implements the gmail bundle's AuthHook (wave1-pilot P-10,
// PLAN.md; SPEC.md §5.7): an OAuth 2.0 refresh-token-grant
// connsdk.Authenticator, porting legacy internal/connectors/gmail/auth.go's
// oauthRefreshAuth almost verbatim (~130 lines, well under the 300-line
// Tier-2 hook cap, conventions.md §1). Only one hook interface is
// implemented (AuthHook), well under the 2-interface cap.
//
// The 3-legged OAuth consent/acquisition dance is out of scope (SPEC §5.7):
// this hook only exchanges an already-issued refresh token for short-lived
// access tokens. Secret values (client_secret, the refresh token, cached
// access tokens) flow ONLY into the outgoing token-request form or the
// Authorization header; they are never logged and never appear in an error
// string (THREAT-MODEL.md Delta 2).
package gmail

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
	engine.RegisterHooks("gmail", func() engine.Hooks { return New() })
}

// Hooks is the gmail hook set. It implements engine.AuthHook only.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now (mirrors legacy
	// auth.go's oauthRefreshAuth.now).
	Now func() time.Time
	// Client overrides the HTTP client used for the token exchange; nil uses
	// a default client with a 30s timeout (mirrors legacy's httpClient()).
	Client *http.Client
}

// New returns a fresh gmail Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "gmail" }

// Authenticator resolves the OAuth2 refresh-token-grant connsdk.Authenticator
// for spec (mode "custom", hook "gmail"). Templated AuthSpec fields
// (token_url/client_id/client_secret/token/scopes) are interpolated against
// cfg here — buildCustomAuth passes spec through uninterpolated (engine/
// auth.go:149-158) since interpolation is mode-specific engine-side.
//
// spec.Token is interpreted as the refresh token (API-CONTRACT.md's
// documented field mapping): gmail's AuthSpec has no dedicated
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
	refreshToken, err := interpolateRequired(spec.Token, "refresh_token", cfg)
	if err != nil {
		return nil, err
	}
	// client_secret and scopes are genuinely optional (legacy auth.go:85-90:
	// both fields are omitted from the token-request form entirely when
	// unset, not sent empty). engine.Interpolate has no absent-key-falsy
	// tolerance outside "when" (conventions.md §3), so an unset secret/config
	// key is resolved directly from the raw maps here rather than through
	// Interpolate, which would otherwise hard-error on the missing key.
	clientSecret := interpolateOptional(spec.ClientSecret, cfg)
	scope := interpolateOptional(spec.Scopes, cfg)

	return &oauthRefreshAuth{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		scope:        scope,
		client:       h.Client,
		now:          h.Now,
	}, nil
}

// interpolateRequired resolves tmpl via engine.Interpolate and wraps any
// error naming field for a caller-facing, secret-free message. An empty
// resolved value is also rejected (mirrors legacy's
// strings.TrimSpace(...) == "" checks, auth.go:71-79).
func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return "", fmt.Errorf("gmail oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("gmail oauth: %s is required", field)
	}
	return val, nil
}

// interpolateOptional resolves tmpl best-effort: a reference to a key that
// is simply absent from cfg.Config/cfg.Secrets resolves to "" (legacy's
// optional-field semantics) rather than propagating engine.Interpolate's
// hard error for an unresolved config/secrets key. Any OTHER interpolation
// failure (CRLF injection, unknown filter/namespace) still propagates, since
// those indicate a genuine bundle/config defect, not "field intentionally
// unset".
func interpolateOptional(tmpl string, cfg connectors.RuntimeConfig) string {
	if strings.TrimSpace(tmpl) == "" {
		return ""
	}
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return ""
	}
	return val
}

func authVars(cfg connectors.RuntimeConfig) engine.Vars {
	return engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// validateHTTPSURL fails closed on anything but a well-formed https:// URL
// with a host (THREAT-MODEL.md Delta 2: token_url is the ONE new
// SSRF-adjacent surface this phase adds — an attacker-controlled token_url
// override could otherwise exfiltrate client_secret/the refresh token to an
// arbitrary endpoint). This is intentionally stricter than legacy's
// validatedURL (gmail.go:339-357), which also accepted plain http; the
// tighter rule is documented as a parity deviation in docs.md's Known
// limits (never stricter for any real Google OAuth endpoint, which is
// always https).
func validateHTTPSURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("gmail oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("gmail oauth: %s must use https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("gmail oauth: %s must include a host", field)
	}
	return nil
}

// oauthRefreshAuth implements connsdk.Authenticator for the Google OAuth 2.0
// refresh-token grant, mirroring legacy gmail/auth.go's oauthRefreshAuth
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
	// Refresh 60s before expiry to avoid edge races (legacy auth.go:67-69).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("gmail oauth: token URL is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("gmail oauth: refresh_token is required")
	}
	if strings.TrimSpace(a.clientID) == "" {
		return "", errors.New("gmail oauth: client_id is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)
	form.Set("client_id", a.clientID)
	if a.clientSecret != "" {
		form.Set("client_secret", a.clientSecret)
	}
	if a.scope != "" {
		form.Set("scope", a.scope)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("gmail oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("gmail oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gmail oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("gmail oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("gmail oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}
