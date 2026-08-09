// Package dockerhub implements the dockerhub bundle's AuthHook (conventions.md
// §1's Tier-2 hook table: "token-exchange auth"): a username+PAT-to-session-JWT
// exchange connsdk.Authenticator.
//
// Docker Hub's API v2 does not accept a Personal Access Token (PAT) directly as
// a static bearer token. Per Docker Hub's own OpenAPI document ("Authentication"
// tag description), a client must first POST the account username plus a
// password OR a PAT to POST /v2/users/login, which returns a short-lived
// session JWT in {"token": "<jwt>"}; that JWT is then sent as
// "Authorization: Bearer <jwt>" on every subsequent request. This is a
// non-OAuth2-shaped, JSON-bodied, no-client-credential exchange that the
// engine's declarative auth modes (bearer/basic/api_key_*/oauth2_*) cannot
// express (oauth2_client_credentials/oauth2_refresh_token always POST a
// form-encoded grant_type/client_id/client_secret body and expect an OAuth2
// token response shape; Docker Hub's login takes a plain JSON
// {username,password} body and returns only {token}, no client_id/secret, no
// expires_in). A custom AuthHook (mirroring akeneo's identical escape hatch for
// its own non-oauth2_client_credentials-shaped password-grant exchange)
// resolves that gap without inventing new engine dialect.
//
// The returned JWT has no `expires_in` in the login response, so this hook
// decodes the JWT payload's standard `exp` (Unix seconds) claim itself —
// signature verification is not needed since the token was just received
// directly from Docker Hub over TLS — and refreshes 60s before that expiry,
// mirroring akeneo's cache-and-refresh window. A JWT with no parseable `exp`
// claim falls back to a conservative 4-minute cache TTL.
//
// Secret values (the configured PAT/password, the cached session JWT) flow
// ONLY into the outgoing login request body or the Authorization header; they
// are never logged and never appear in an error string.
package dockerhub

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
	engine.RegisterHooks("dockerhub", func() engine.Hooks { return New() })
}

// Hooks is the dockerhub hook set. It implements engine.AuthHook only.
type Hooks struct {
	// Client overrides the HTTP client used for the token exchange; nil uses
	// a default client with a 30s timeout.
	Client *http.Client
	// Now is injectable for tests; nil uses time.Now.
	Now func() time.Time
}

var (
	_ engine.Hooks    = (*Hooks)(nil)
	_ engine.AuthHook = (*Hooks)(nil)
)

// New returns a fresh dockerhub Hooks value as engine.Hooks.
func New() engine.Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "dockerhub" }

// fallbackTokenTTL is used only when the session JWT's own `exp` claim cannot
// be parsed. Conservative on purpose: Docker Hub session tokens are
// documented as short-lived, and re-authenticating too early costs one extra
// HTTP round trip while re-authenticating too late (a stale-but-cached token
// on the wire) costs a failed request.
const fallbackTokenTTL = 4 * time.Minute

// Authenticator resolves the dockerhub connsdk.Authenticator for spec (mode
// "custom", hook "dockerhub"). Templated AuthSpec fields
// (token_url/username/password) are interpolated against cfg here —
// buildCustomAuth passes spec through uninterpolated (engine/auth.go) since
// interpolation is mode-specific engine-side.
//
// The returned authenticator is path-aware (dualAuth, below): Docker Hub's
// own OpenAPI document declares TWO distinct bearer security schemes —
// `bearerAuth` (the account session JWT this hook obtains via the
// username+PAT login exchange) for every non-SCIM endpoint, and
// `bearerSCIMAuth` (a separate, admin-console-issued token) for every
// /v2/scim/2.0/** endpoint. These are not interchangeable: nothing in
// Docker Hub's documentation states the account session JWT is accepted on
// SCIM routes, so this hook never attempts to reuse it there. The
// SCIM-scoped credential is read directly from cfg.Secrets["scim_bearer_token"]
// (bypassing template resolution deliberately — it is genuinely optional,
// and engine.Interpolate hard-errors on an undeclared secrets key, which
// would otherwise reject every non-SCIM command too when the operator has
// not configured SCIM at all) — a second, independently-configured secret,
// never derived from or substituted for docker_pat.
//
// The two credentials are also INDEPENDENTLY configurable. streams.json
// declares two custom-auth specs: one gated on secrets.docker_pat carrying
// the login-exchange fields, and one gated on secrets.scim_bearer_token
// carrying none of them (the `when` grammar has no OR, so one spec cannot
// express "either secret"). A spec with no login fields therefore means
// "SCIM-only connection": every SCIM command works with no docker_pat, and
// every non-SCIM command fails closed naming docker_pat rather than being
// sent unauthenticated.
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	scimToken := strings.TrimSpace(cfg.Secrets["scim_bearer_token"])
	auth := &dualAuth{scimToken: scimToken, scimPrefixes: scimPathPrefixes(cfg)}

	if strings.TrimSpace(spec.Username) == "" && strings.TrimSpace(spec.Password) == "" {
		if scimToken == "" {
			return nil, errors.New("dockerhub auth: docker_pat or scim_bearer_token is required")
		}
		return auth, nil
	}

	tokenURL, err := interpolateRequired(spec.TokenURL, "token_url", cfg)
	if err != nil {
		return nil, err
	}
	if err := validateHTTPSURL(tokenURL, "token_url"); err != nil {
		return nil, err
	}
	username, err := interpolateRequired(spec.Username, "docker_username", cfg)
	if err != nil {
		return nil, err
	}
	password, err := interpolateRequired(spec.Password, "docker_pat", cfg)
	if err != nil {
		return nil, err
	}

	auth.session = &sessionLoginAuth{
		tokenURL: tokenURL,
		username: username,
		password: password,
		client:   h.Client,
		now:      h.Now,
	}
	return auth, nil
}

// scimAPIPath is the SCIM 2.0 sub-path every bearerSCIMAuth endpoint sits
// under (dockerhub's api_surface.json /v2/scim/2.0/** rows), relative to the
// Docker Hub API root.
const (
	scimAPIPath      = "/scim/2.0/"
	dockerHubAPIRoot = "/v2"
)

func scimPathPrefixes(cfg connectors.RuntimeConfig) []string {
	baseURL := strings.TrimSpace(cfg.Config["base_url"])
	if baseURL == "" {
		return []string{dockerHubAPIRoot + scimAPIPath}
	}
	return []string{engine.RequestURLPathForBaseURL(dockerHubAPIRoot+scimAPIPath, baseURL, dockerHubAPIRoot)}
}

// dualAuth routes each outgoing request to one of two independent
// credentials by request path: the account session JWT (sessionLoginAuth)
// for everything else, or a static SCIM-scoped bearer token for
// /v2/scim/2.0/** requests. It never falls back from one credential to the
// other — a SCIM request with no scimToken configured fails closed with a
// named error rather than silently sending the account session JWT to an
// endpoint Docker Hub documents as requiring a different credential, and a
// non-SCIM request on a SCIM-only connection fails closed naming docker_pat
// rather than sending the SCIM token to a bearerAuth endpoint.
type dualAuth struct {
	session      *sessionLoginAuth
	scimToken    string
	scimPrefixes []string
}

func (a *dualAuth) Apply(ctx context.Context, req *http.Request) error {
	if a.isSCIMPath(req.URL.Path) {
		if strings.TrimSpace(a.scimToken) == "" {
			return errors.New("dockerhub auth: scim_bearer_token is required for /v2/scim/2.0/** commands")
		}
		req.Header.Set("Authorization", "Bearer "+a.scimToken)
		return nil
	}
	if a.session == nil {
		return errors.New("dockerhub auth: docker_pat is required for every non-SCIM command")
	}
	return a.session.Apply(ctx, req)
}

func (a *dualAuth) isSCIMPath(path string) bool {
	for _, prefix := range a.scimPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// interpolateRequired resolves tmpl via engine.Interpolate and wraps any error
// naming field for a caller-facing, secret-free message. An empty resolved
// value is also rejected.
func interpolateRequired(tmpl, field string, cfg connectors.RuntimeConfig) (string, error) {
	val, err := engine.Interpolate(tmpl, authVars(cfg))
	if err != nil {
		return "", fmt.Errorf("dockerhub auth: resolve %s: %w", field, err)
	}
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("dockerhub auth: %s is required", field)
	}
	return val, nil
}

func authVars(cfg connectors.RuntimeConfig) engine.Vars {
	return engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
}

// validateHTTPSURL fails closed on anything but a well-formed https:// URL
// with a host. Unlike akeneo (a customer-supplied PIM host, not necessarily
// https-only), Docker Hub's own API is only ever served over https, so this
// hook narrows to https-only the way gmail's hook narrows for its own
// well-known Google OAuth endpoint.
func validateHTTPSURL(raw, field string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("dockerhub auth: %s is invalid: %w", field, err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("dockerhub auth: %s must use https, got %q", field, parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("dockerhub auth: %s must include a host", field)
	}
	return nil
}

// sessionLoginAuth implements connsdk.Authenticator for Docker Hub's
// username+PAT-to-JWT login exchange: POST a JSON body {username, password}
// to tokenURL, cache the resulting session JWT until 60s before its own `exp`
// claim (or fallbackTokenTTL if `exp` cannot be parsed), then set
// Authorization: Bearer <jwt> on each request. Secret values never flow
// anywhere except the outgoing login request body or the Authorization
// header; they are never logged.
type sessionLoginAuth struct {
	tokenURL string
	username string
	password string
	client   *http.Client

	// now is injectable for tests; defaults to time.Now.
	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *sessionLoginAuth) timeNow() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *sessionLoginAuth) httpClient() *http.Client {
	if a.client != nil {
		return a.client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Apply ensures a fresh session JWT and sets the Authorization header.
func (a *sessionLoginAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.sessionToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *sessionLoginAuth) sessionToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races (mirrors akeneo's hook).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	bodyBytes, err := json.Marshal(map[string]string{
		"username": a.username,
		"password": a.password,
	})
	if err != nil {
		return "", fmt.Errorf("dockerhub auth: encode login request body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("dockerhub auth: build login request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	client := *a.httpClient()
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("dockerhub auth: login request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("dockerhub auth: login endpoint returned %s", resp.Status)
	}

	var out struct {
		Token string `json:"token"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("dockerhub auth: decode login response: %w", err)
	}
	if strings.TrimSpace(out.Token) == "" {
		return "", errors.New("dockerhub auth: login response missing token")
	}

	a.token = out.Token
	a.expires = a.timeNow().Add(jwtTTL(out.Token, a.timeNow()))
	return a.token, nil
}

// jwtTTL returns the remaining lifetime of a JWT (header.payload.signature)
// derived from its unvalidated `exp` (Unix seconds) claim, relative to now.
// Signature verification is deliberately skipped: this token was just
// received directly from Docker Hub over TLS, so the concern here is caching
// duration, not authenticity. Any parse failure (malformed JWT, missing/
// non-numeric exp, or an exp already in the past) falls back to
// fallbackTokenTTL rather than caching indefinitely or erroring the whole
// auth flow over a caching nicety.
func jwtTTL(token string, now time.Time) time.Duration {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fallbackTokenTTL
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fallbackTokenTTL
	}
	var claims struct {
		Exp json.Number `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return fallbackTokenTTL
	}
	expSeconds, err := claims.Exp.Int64()
	if err != nil {
		return fallbackTokenTTL
	}
	ttl := time.Unix(expSeconds, 0).Sub(now)
	if ttl <= 0 {
		return fallbackTokenTTL
	}
	return ttl
}
