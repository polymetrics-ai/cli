// Package salesloft implements the salesloft bundle's AuthHook (fresh
// migration, capability parity with
// internal/connectors/salesloft/auth.go's oauthRefreshAuth): an OAuth 2.0
// refresh-token-grant connsdk.Authenticator, porting legacy's
// oauthRefreshAuth almost verbatim (well under the 300-line Tier-2 hook cap,
// conventions.md §1). Only one hook interface is implemented (AuthHook),
// well under the 2-interface cap.
//
// This is the identical shape to internal/connectors/hooks/strava's pilot
// AuthHook: the engine's declarative oauth2_client_credentials auth mode
// only performs a client-credentials grant, never grant_type=refresh_token,
// so a refresh-token exchange genuinely needs a Tier-2 hook (token-exchange
// auth is a listed legitimate Tier-2 trigger, conventions.md §1). This hook
// is only reached when secrets.api_key is unset AND secrets.refresh_token IS
// set (streams.json's base.auth candidate-list ordering: api_key Bearer is
// tried first, matching legacy's authenticator's own precedence — API key
// wins when both are configured).
//
// Salesloft additionally rotates the refresh token on every exchange
// (legacy auth.go:113-116) and honors a pre-existing access_token secret
// once, before any network refresh call (legacy auth.go:63-69, the
// "seedToken" behavior) — both are preserved here. The seed/rotation values
// are read directly from cfg.Secrets rather than via a dedicated AuthSpec
// field (AuthSpec has no "seed_token" concept; Token already carries the
// refresh token per the strava/gmail convention), exactly like other hooks
// reading a non-AuthSpec config/secret value directly off cfg.
//
// Secret values (client_id, client_secret, the refresh token, any seed/
// rotated access tokens) flow ONLY into the outgoing token-request form or
// the Authorization header; they are never logged and never appear in an
// error string (THREAT-MODEL.md Delta 2).
package salesloft

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
	engine.RegisterHooks("salesloft", func() engine.Hooks { return New() })
}

// Hooks is the salesloft hook set. It implements engine.AuthHook only.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now.
	Now func() time.Time
	// Client overrides the HTTP client used for the token exchange; nil uses
	// a default client with a 30s timeout (mirrors legacy's inline default).
	Client *http.Client
}

// New returns a fresh salesloft Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "salesloft" }

// Authenticator resolves the OAuth2 refresh-token-grant connsdk.Authenticator
// for spec (mode "custom", hook "salesloft"). Templated AuthSpec fields
// (token_url/client_id/client_secret/token) are interpolated against cfg here
// — buildCustomAuth passes spec through uninterpolated (engine/auth.go),
// since interpolation is mode-specific engine-side.
//
// spec.Token is interpreted as the refresh token (matching strava/gmail's
// identical convention): salesloft's AuthSpec has no dedicated
// "refresh_token" field, and Token is otherwise unused by the custom mode.
// The optional access_token seed is read directly from cfg.Secrets (not
// templated through AuthSpec) since it is not one of AuthSpec's fields.
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
		seedToken:    strings.TrimSpace(salesloftSecret(cfg, "access_token")),
		client:       h.Client,
		now:          h.Now,
	}, nil
}

// salesloftSecret resolves a secret. It accepts both the flat key (e.g.
// "access_token") and the dotted catalog form (e.g. "credentials.access_token")
// since some deployments declare secret fields under a credentials object.
func salesloftSecret(cfg connectors.RuntimeConfig, key string) string {
	if cfg.Secrets == nil {
		return ""
	}
	if v, ok := cfg.Secrets[key]; ok {
		return v
	}
	return cfg.Secrets["credentials."+key]
}

// interpolateRequired resolves tmpl via engine.Interpolate and wraps any
// error naming field for a caller-facing, secret-free message. An empty
// resolved value is also rejected (mirrors legacy's
// strings.TrimSpace(...) == "" checks).
func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return "", fmt.Errorf("salesloft oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("salesloft oauth: %s is required", field)
	}
	return val, nil
}

func authVars(cfg connectors.RuntimeConfig) engine.Vars {
	return engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// validateHTTPURL fails closed on anything but a well-formed http(s):// URL
// with a host, matching legacy's validateURL, which accepts plain http (for
// local test servers) as well as https.
func validateHTTPURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("salesloft oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("salesloft oauth: %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("salesloft oauth: %s must include a host", field)
	}
	return nil
}

// refreshTokenAuth implements connsdk.Authenticator for Salesloft's OAuth 2.0
// refresh-token grant, mirroring legacy auth.go's oauthRefreshAuth
// field-for-field and behavior-for-behavior: exchange
// client_id/client_secret/refresh_token (as form fields — Salesloft's token
// endpoint does NOT use HTTP Basic client auth, unlike Pinterest's) for a
// short-lived bearer access token at tokenURL, cache it until 60s before its
// declared expiry (falling back to a 1-hour TTL when the response has no
// positive expires_in), honor a pre-existing access_token seed exactly once
// before any network refresh, rotate the refresh token whenever the response
// carries a new one, then set Authorization: Bearer <token> on each request.
// Secret values never flow anywhere except the outgoing token-request form
// or the Authorization header; they are never logged.
type refreshTokenAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	// seedToken, when set, is used until it expires/refreshes so an
	// already-valid access_token from config is honored without an
	// immediate network refresh (legacy auth.go:63-69).
	seedToken string
	client    *http.Client

	// now is injectable for tests; defaults to time.Now.
	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
	seeded  bool
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

	// Honor a still-valid cached token (refresh 60s before expiry).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	// Honor a config-provided seed access token exactly once, before any
	// refresh (legacy auth.go:63-69).
	if !a.seeded && a.seedToken != "" {
		a.seeded = true
		a.token = a.seedToken
		a.expires = a.timeNow().Add(50 * time.Minute)
		return a.token, nil
	}

	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("salesloft oauth: token_url is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("refresh_token", a.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("salesloft oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("salesloft oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("salesloft oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken  string      `json:"access_token"`
		RefreshToken string      `json:"refresh_token"`
		TokenType    string      `json:"token_type"`
		ExpiresIn    json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("salesloft oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("salesloft oauth: token response missing access_token")
	}
	// Salesloft rotates the refresh token; keep the latest for subsequent
	// refreshes (legacy auth.go:113-116).
	if strings.TrimSpace(out.RefreshToken) != "" {
		a.refreshToken = out.RefreshToken
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}
