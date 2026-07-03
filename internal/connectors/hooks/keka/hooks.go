// Package keka is the Tier-2 escape hatch for the keka bundle: a single
// AuthHook porting legacy's custom OAuth2 client-credentials-shaped token
// exchange (internal/connectors/keka/auth.go's kekaTokenAuth, read-only
// reference).
//
// Keka's token endpoint requires a non-standard grant_type value
// ("kekaapi", not the RFC 6749 "client_credentials" the engine's declarative
// oauth2_client_credentials auth mode hard-codes — engine/connsdk/auth.go's
// accessToken always sends "grant_type":"client_credentials" and has no
// override) plus an api_key form field alongside client_id/client_secret.
// The declarative auth[].extra_params dialect (conventions.md §3) can ADD a
// form field but cannot REPLACE grant_type (connsdk.OAuth2ClientCredentials.
// ExtraParams is applied via form.Add after grant_type is already form.Set
// to the hard-coded value, so declaring extra_params.grant_type would send
// BOTH values on the wire, not override it) — a genuine ENGINE_GAP, not a
// Tier-1-expressible shape. Token-exchange auth with a non-standard grant is
// a named legitimate Tier-2 trigger (conventions.md §1's AuthHook row); this
// single hook interface, well under the ~300-line soft target, is the
// correct escape hatch rather than stretching the declarative dialect.
package keka

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

const (
	defaultGrantType = "kekaapi"
	defaultScope     = "kekaapi"
	defaultTokenURL  = "https://login.keka.com/connect/token"
)

func init() {
	engine.RegisterHooks("keka", func() engine.Hooks { return New() })
}

// Hooks implements engine.AuthHook for the keka bundle. Its only state is
// test-injection overrides (mirrors hooks/gmail's Hooks shape); every method
// is otherwise a pure function of its arguments (the returned Authenticator
// carries its own token cache, mirroring legacy's kekaTokenAuth).
type Hooks struct {
	// Client overrides the HTTP client used for the token exchange; nil uses
	// http.DefaultClient (mirrors legacy's kekaTokenAuth.Client fallback).
	Client *http.Client
	// Now is injectable for tests; nil uses time.Now (mirrors legacy
	// auth.go's kekaTokenAuth.now).
	Now func() time.Time
}

// New returns a fresh keka Hooks value.
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "keka" }

var (
	_ engine.Hooks    = (*Hooks)(nil)
	_ engine.AuthHook = (*Hooks)(nil)
)

// Authenticator builds a token-caching OAuth2 authenticator using Keka's
// custom grant_type/api_key token exchange (matches legacy's kekaTokenAuth
// exactly: same form fields, same 60s-early refresh, same 3600s fallback
// TTL when the token response omits expires_in). ctx is honored so a caller
// cancellation aborts an in-flight token fetch (F8-equivalent: the real
// caller context is threaded through, never context.Background()).
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	tokenURL := strings.TrimSpace(cfg.Config["token_url"])
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}
	clientID := strings.TrimSpace(cfg.Config["client_id"])
	if clientID == "" {
		return nil, errors.New("keka connector requires config client_id")
	}
	clientSecret := ""
	if cfg.Secrets != nil {
		clientSecret = cfg.Secrets["client_secret"]
	}
	if strings.TrimSpace(clientSecret) == "" {
		return nil, errors.New("keka connector requires secret client_secret")
	}
	apiKey := ""
	if cfg.Secrets != nil {
		apiKey = cfg.Secrets["api_key"]
	}
	if strings.TrimSpace(apiKey) == "" {
		// api_key is x-secret in spec.json but not required; fall back to a
		// plain config value for callers that (against the marker's intent)
		// still pass it unencrypted, matching legacy's cfg.Config["api_key"]
		// read exactly.
		apiKey = cfg.Config["api_key"]
	}
	grantType := valueOr(cfg.Config["grant_type"], defaultGrantType)
	scope := valueOr(cfg.Config["scope"], defaultScope)

	return &tokenAuth{
		TokenURL:     tokenURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		APIKey:       apiKey,
		GrantType:    grantType,
		Scope:        scope,
		Client:       h.Client,
		Now:          h.Now,
	}, nil
}

func valueOr(raw, fallback string) string {
	if v := strings.TrimSpace(raw); v != "" {
		return v
	}
	return fallback
}

// tokenAuth implements connsdk.Authenticator using Keka's custom OAuth2
// client-credentials-shaped exchange (grant_type=kekaapi + api_key form
// field). It is a direct port of legacy's kekaTokenAuth
// (internal/connectors/keka/auth.go): the fetched bearer token is cached
// until shortly before expiry; secret values (client_secret, api_key, the
// token itself) are never logged.
type tokenAuth struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	APIKey       string
	GrantType    string
	Scope        string
	// Client is used for the token request. Defaults to http.DefaultClient.
	Client *http.Client
	// Now is injectable for tests. Defaults to time.Now.
	Now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *tokenAuth) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

// Apply ensures a fresh token and sets the Authorization header.
func (a *tokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *tokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races (matches legacy).
	if a.token != "" && a.now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.TokenURL) == "" {
		return "", errors.New("keka oauth: token_url is required")
	}

	form := url.Values{}
	form.Set("grant_type", valueOr(a.GrantType, defaultGrantType))
	form.Set("scope", valueOr(a.Scope, defaultScope))
	form.Set("client_id", a.ClientID)
	form.Set("client_secret", a.ClientSecret)
	if strings.TrimSpace(a.APIKey) != "" {
		form.Set("api_key", a.APIKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("keka oauth: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("keka oauth: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("keka oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("keka oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("keka oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.now().Add(ttl)
	return a.token, nil
}
