// Package pinterest implements the pinterest bundle's AuthHook (fresh
// migration, capability parity with
// internal/connectors/pinterest/pinterest.go's refreshTokenAuth): an OAuth
// 2.0 refresh-token-grant connsdk.Authenticator, porting legacy's
// refreshTokenAuth almost verbatim (well under the 300-line Tier-2 hook cap,
// conventions.md §1). Only one hook interface is implemented (AuthHook),
// well under the 2-interface cap.
//
// This is the identical shape to internal/connectors/hooks/strava's pilot
// AuthHook: the engine's declarative oauth2_client_credentials auth mode
// only performs a client-credentials grant, never grant_type=refresh_token,
// so a refresh-token exchange genuinely needs a Tier-2 hook (token-exchange
// auth is a listed legitimate Tier-2 trigger, conventions.md §1).
//
// Unlike strava, Pinterest's token endpoint authenticates the CLIENT via
// HTTP Basic (client_id/client_secret in the Authorization header), not
// form-encoded client_id/client_secret fields — this is legacy's exact
// wire shape (pinterest.go's accessToken: httpReq.SetBasicAuth(clientID,
// clientSecret), form body carries only grant_type+refresh_token) and is
// preserved here rather than normalized to strava's shape.
//
// Secret values (client_id, client_secret, the refresh token, cached access
// tokens) flow ONLY into the outgoing token-request Basic-auth header/form
// or the data-request Authorization header; they are never logged and never
// appear in an error string (THREAT-MODEL.md Delta 2).
package pinterest

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
	engine.RegisterHooks("pinterest", func() engine.Hooks { return New() })
}

// Hooks is the pinterest hook set. It implements engine.AuthHook only.
type Hooks struct {
	// Now is injectable for tests; nil uses time.Now.
	Now func() time.Time
	// Client overrides the HTTP client used for the token exchange; nil uses
	// a default client with a 30s timeout (mirrors legacy's inline default,
	// pinterest.go:365-368).
	Client *http.Client
}

// New returns a fresh pinterest Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "pinterest" }

// Authenticator resolves the OAuth2 refresh-token-grant connsdk.Authenticator
// for spec (mode "custom", hook "pinterest"). Templated AuthSpec fields
// (token_url/client_id/client_secret/token) are interpolated against cfg here
// — buildCustomAuth passes spec through uninterpolated (engine/auth.go),
// since interpolation is mode-specific engine-side.
//
// spec.Token is interpreted as the refresh token (matching strava/gmail's
// identical convention): pinterest's AuthSpec has no dedicated
// "refresh_token" field, and Token is otherwise unused by the custom mode.
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
// strings.TrimSpace(...) == "" checks).
func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return "", fmt.Errorf("pinterest oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("pinterest oauth: %s is required", field)
	}
	return val, nil
}

func authVars(cfg connectors.RuntimeConfig) engine.Vars {
	return engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// validateHTTPURL fails closed on anything but a well-formed http(s):// URL
// with a host, matching legacy's pinterestBaseURL-style validation, which
// accepts plain http (for local test servers) as well as https.
func validateHTTPURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("pinterest oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("pinterest oauth: %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("pinterest oauth: %s must include a host", field)
	}
	return nil
}

// refreshTokenAuth implements connsdk.Authenticator for Pinterest's OAuth 2.0
// refresh-token grant, mirroring legacy pinterest.go's refreshTokenAuth
// field-for-field and behavior-for-behavior: exchange the refresh_token for a
// short-lived bearer access token at tokenURL, authenticating the CLIENT via
// HTTP Basic (client_id/client_secret), with a form body carrying only
// grant_type=refresh_token and refresh_token — cache the token until 60s
// before its declared expiry (falling back to a 1-hour TTL when the response
// has no positive expires_in, matching legacy's decodeTokenResponse), then
// set Authorization: Bearer <token> on each request. Secret values never
// flow anywhere except the outgoing token-request Basic-auth header/form or
// the Authorization header; they are never logged.
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
	// Refresh 60s before expiry to avoid edge races (matches legacy's
	// pinterest.go accessToken: a.now().Add(60*time.Second).Before(a.expires)).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("pinterest oauth: token URL is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("pinterest oauth: refresh_token is required")
	}
	if strings.TrimSpace(a.clientID) == "" {
		return "", errors.New("pinterest oauth: client_id is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("pinterest oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.SetBasicAuth(a.clientID, a.clientSecret)

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("pinterest oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("pinterest oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("pinterest oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("pinterest oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	// Legacy default TTL is 1 hour when expires_in is absent/non-positive
	// (pinterest.go's decodeTokenResponse), matching Pinterest's default
	// access-token lifetime.
	a.expires = a.timeNow().Add(time.Hour)
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		a.expires = a.timeNow().Add(time.Duration(secs) * time.Second)
	}
	return a.token, nil
}
