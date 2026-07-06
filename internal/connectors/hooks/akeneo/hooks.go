// Package akeneo implements the akeneo bundle's AuthHook (conventions.md
// §1's Tier-2 hook table: "token-exchange auth"): an OAuth2 password-grant
// connsdk.Authenticator, porting legacy internal/connectors/akeneo/akeneo.go's
// passwordGrantAuth almost verbatim (well under the 300-line Tier-2 hook
// soft target; conventions.md §1). Only one hook interface is implemented
// (AuthHook).
//
// Akeneo's modern API authenticates in two steps: POST HTTP Basic
// client_id:secret plus a JSON body {grant_type:password, username,
// password} to /api/oauth/v1/token, returning {access_token, expires_in,
// token_type}; that access token is then sent as Authorization: Bearer on
// every subsequent request. This connector was previously quarantined
// (docs/migration/quarantine.json, blocker AUTH_COMPLEX) because the
// engine's only declarative token-exchange auth mode
// ("oauth2_client_credentials") always POSTs a form-encoded
// grant_type/client_id/client_secret/scope body (engine/auth.go's
// buildOAuth2ClientCredentials) — Akeneo's password grant needs a JSON
// body carrying username/password instead, which that mode cannot express.
// A custom AuthHook (mirroring gmail's/jamf-pro's identical escape hatch
// for their own non-oauth2_client_credentials-shaped exchanges) resolves
// that blocker without inventing new engine dialect.
//
// Secret values (the client secret, the API user password, cached access
// tokens) flow ONLY into the outgoing token-request Basic header/JSON body
// or the Authorization header; they are never logged and never appear in
// an error string.
package akeneo

import (
	"bytes"
	"context"
	"encoding/base64"
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
	engine.RegisterHooks("akeneo", func() engine.Hooks { return New() })
}

// Hooks is the akeneo hook set. It implements engine.AuthHook only.
type Hooks struct {
	// Client overrides the HTTP client used for the token exchange; nil uses
	// a default client with a 30s timeout (mirrors legacy's inline client
	// construction, akeneo.go:314-317).
	Client *http.Client
	// Now is injectable for tests; nil uses time.Now (mirrors legacy
	// passwordGrantAuth.now).
	Now func() time.Time
}

// New returns a fresh akeneo Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

var (
	_ engine.Hooks    = (*Hooks)(nil)
	_ engine.AuthHook = (*Hooks)(nil)
)

func (h *Hooks) ConnectorName() string { return "akeneo" }

// Authenticator resolves the OAuth2 password-grant connsdk.Authenticator for
// spec (mode "custom", hook "akeneo"). Templated AuthSpec fields
// (token_url/client_id/client_secret/username/password) are interpolated
// against cfg here — buildCustomAuth passes spec through uninterpolated
// (engine/auth.go) since interpolation is mode-specific engine-side.
//
// spec.Username/spec.Password carry the Akeneo API user's username/password
// (API-CONTRACT.md's documented field mapping): akeneo's AuthSpec has no
// dedicated "api_username"/"api_password" fields, and AuthSpec.Username/
// Password are otherwise unused by the custom mode.
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

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
	clientSecret, err := interpolateRequired(spec.ClientSecret, "secret", cfg)
	if err != nil {
		return nil, err
	}
	username, err := interpolateRequired(spec.Username, "api_username", cfg)
	if err != nil {
		return nil, err
	}
	password, err := interpolateRequired(spec.Password, "password", cfg)
	if err != nil {
		return nil, err
	}

	return &passwordGrantAuth{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		username:     username,
		password:     password,
		client:       h.Client,
		now:          h.Now,
	}, nil
}

// interpolateRequired resolves tmpl via engine.Interpolate and wraps any
// error naming field for a caller-facing, secret-free message. An empty
// resolved value is also rejected (mirrors legacy's requireLiveConfig
// blank-value checks, akeneo.go:352-366).
func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return "", fmt.Errorf("akeneo oauth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("akeneo oauth: %s is required", field)
	}
	return val, nil
}

func authVars(cfg connectors.RuntimeConfig) engine.Vars {
	return engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// validateHTTPURL fails closed on anything but a well-formed http(s):// URL
// with a host, mirroring legacy's akeneoBaseURL guard (akeneo.go:387-406):
// legacy accepts both http and https (its own test suite exercises a plain
// http httptest.Server for both the token endpoint and resource requests),
// so unlike gmail's hook (which narrows to https-only for the unrelated
// Google OAuth endpoint) this hook keeps legacy's exact scheme tolerance —
// narrowing it here would be an undocumented, unnecessary parity deviation
// for a connector whose only production deployments use a customer-supplied
// PIM host that is not necessarily on a well-known https-only domain.
func validateHTTPURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("akeneo oauth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("akeneo oauth: %s must use http or https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("akeneo oauth: %s must include a host", field)
	}
	return nil
}

// passwordGrantAuth implements connsdk.Authenticator for Akeneo's OAuth2
// password grant, mirroring legacy akeneo.go's passwordGrantAuth
// field-for-field and behavior-for-behavior: POST a JSON body
// {grant_type:password, username, password} with HTTP Basic
// client_id:secret to tokenURL, cache the resulting access token until 60s
// before its declared expiry, then set Authorization: Bearer <token> on
// each request. Secret values never flow anywhere except the outgoing
// token-request Basic header/JSON body or the Authorization header; they
// are never logged.
type passwordGrantAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	username     string
	password     string
	client       *http.Client

	// now is injectable for tests; defaults to time.Now.
	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *passwordGrantAuth) timeNow() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *passwordGrantAuth) httpClient() *http.Client {
	if a.client != nil {
		return a.client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Apply ensures a fresh access token and sets the Authorization header.
func (a *passwordGrantAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *passwordGrantAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races (legacy akeneo.go:291-293).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("akeneo oauth: token URL is required")
	}
	if strings.TrimSpace(a.username) == "" {
		return "", errors.New("akeneo oauth: api_username is required")
	}
	if strings.TrimSpace(a.password) == "" {
		return "", errors.New("akeneo oauth: password is required")
	}

	bodyBytes, err := json.Marshal(map[string]string{
		"grant_type": "password",
		"username":   a.username,
		"password":   a.password,
	})
	if err != nil {
		return "", fmt.Errorf("akeneo oauth: encode token request body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("akeneo oauth: build token request: %w", err)
	}
	creds := base64.StdEncoding.EncodeToString([]byte(a.clientID + ":" + a.clientSecret))
	httpReq.Header.Set("Authorization", "Basic "+creds)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("akeneo oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("akeneo oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("akeneo oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("akeneo oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}
