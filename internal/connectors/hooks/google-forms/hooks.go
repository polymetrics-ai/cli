// Package googleforms implements the google-forms bundle's AuthHook + CheckHook
// (docs.md "Auth setup"): an OAuth 2.0 refresh-token-grant connsdk.Authenticator,
// porting legacy internal/connectors/google-forms/googleforms.go's
// oauthRefreshAuth almost verbatim (mirrors hooks/gmail/hooks.go, the identical
// Google OAuth2 refresh-token-grant shape). Only two hook interfaces are
// implemented (AuthHook, CheckHook), well under the Tier-2 cap.
//
// The 3-legged OAuth consent/acquisition dance is out of scope: this hook only
// exchanges an already-issued refresh token for short-lived access tokens.
// Secret values (client_secret, the refresh token, cached access tokens) flow
// ONLY into the outgoing token-request form or the Authorization header; they
// are never logged and never appear in an error string.
package googleforms

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
	engine.RegisterHooks("google-forms", func() engine.Hooks { return New() })
}

// Hooks is the google-forms hook set. It implements engine.AuthHook and
// engine.CheckHook.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now.
	Now func() time.Time
	// Client overrides the HTTP client used for the token exchange and the
	// Check request; nil uses a default client with a 30s timeout.
	Client *http.Client
}

// New returns a fresh google-forms Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "google-forms" }

// Authenticator resolves the OAuth2 refresh-token-grant connsdk.Authenticator
// for spec (mode "custom", hook "google-forms"). Templated AuthSpec fields
// are interpolated against cfg here -- buildCustomAuth passes spec through
// uninterpolated (engine/auth.go), since interpolation is mode-specific
// engine-side.
//
// spec.Token is interpreted as the refresh token (mirrors gmail's identical
// AuthSpec field mapping): google-forms' AuthSpec has no dedicated
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
	// client_secret is genuinely optional (legacy googleforms.go's
	// oauthRefreshAuth always sends it, but some Google OAuth client types,
	// e.g. installed-app/native clients, issue refresh tokens that don't
	// require one at token-refresh time -- matching gmail's identical
	// optional-client_secret handling). engine.Interpolate has no
	// absent-key-falsy tolerance outside "when", so an unset secret is
	// resolved directly from the raw map here rather than through
	// Interpolate, which would otherwise hard-error on the missing key.
	clientSecret := interpolateOptional(spec.ClientSecret, cfg)

	return &oauthRefreshAuth{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		client:       h.Client,
		now:          h.Now,
	}, nil
}

// Check implements engine.CheckHook: legacy's Check
// (googleforms.go:74-100) reads the FIRST configured form_id
// (formIDs[0]) as a bounded metadata read confirming auth/connectivity.
// The declarative check dialect has no "take the first element of a
// comma-separated config value" primitive, so this is expressed as a hook
// instead of a bare "{{ config.form_id }}" path template (which would
// break for any multi-form_id config).
func (h *Hooks) Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *engine.Runtime) (bool, error) {
	ids := splitFormIDs(cfg.Config["form_id"])
	if len(ids) == 0 {
		return true, errors.New("google-forms check: config form_id is required (one or more form IDs)")
	}
	path := "/forms/" + url.PathEscape(ids[0])
	if err := rt.Requester.DoJSON(ctx, http.MethodGet, path, nil, nil, nil); err != nil {
		return true, fmt.Errorf("check google-forms: %w", err)
	}
	return true, nil
}

// splitFormIDs splits a comma/space/newline-separated form_id config value,
// ported verbatim from legacy's splitFormIDs.
func splitFormIDs(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == ' ' || r == '\t' || r == '\r'
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

// interpolateRequired resolves tmpl via engine.Interpolate and wraps any
// error naming field for a caller-facing, secret-free message. An empty
// resolved value is also rejected.
func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return "", fmt.Errorf("google-forms oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("google-forms oauth: %s is required", field)
	}
	return val, nil
}

// interpolateOptional resolves tmpl best-effort: ANY engine.Interpolate
// error (an absent config/secrets key -- the intended case -- but also a
// CRLF-injecting resolved value or an unknown filter/namespace reference)
// resolves to "" rather than propagating. Verified benign at this
// function's only call site (spec.ClientSecret, an optional-when-empty
// OAuth token-request POST-form value, never a header or path) -- mirrors
// hooks/gmail/hooks.go's identical helper and documented tolerance.
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
// with a host (the one new SSRF-adjacent surface this hook adds -- an
// attacker-controlled token_url override could otherwise exfiltrate
// client_secret/the refresh token to an arbitrary endpoint). This is
// intentionally stricter than legacy's resolveHTTPURL (googleforms.go),
// which also accepted plain http for both base_url and token_url; the
// tighter rule is documented as a parity deviation in docs.md's Known
// limits (never stricter for any real Google OAuth endpoint, which is
// always https) -- mirrors gmail's identical hook-side tightening.
func validateHTTPSURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("google-forms oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("google-forms oauth: %s must use https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("google-forms oauth: %s must include a host", field)
	}
	return nil
}

// oauthRefreshAuth implements connsdk.Authenticator for the Google OAuth 2.0
// refresh-token grant, mirroring legacy googleforms.go's oauthRefreshAuth
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
	// Refresh 60s before expiry to avoid edge races (legacy
	// googleforms.go's identical 60s guard).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("google-forms oauth: token URL is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("google-forms oauth: refresh_token is required")
	}
	if strings.TrimSpace(a.clientID) == "" {
		return "", errors.New("google-forms oauth: client_id is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.clientID)
	if a.clientSecret != "" {
		form.Set("client_secret", a.clientSecret)
	}
	form.Set("refresh_token", a.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("google-forms oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("google-forms oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("google-forms oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("google-forms oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("google-forms oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}
