// Package jamfpro is the Tier-2 escape hatch for the jamf-pro defs bundle: a
// single AuthHook porting legacy's Basic-credential token exchange
// (internal/connectors/jamf-pro/jamf_pro.go's fetchToken, read-only
// reference).
//
// Jamf Pro's modern API authenticates in two steps: POST HTTP Basic
// credentials (username/password, no request body) to /v1/auth/token,
// returning a custom JSON shape {"token":..., "expires":...}; that token is
// then sent as Authorization: Bearer on every subsequent request. This is a
// genuine token-exchange auth scheme — a named legitimate Tier-2 AuthHook
// trigger (conventions.md §1's hook table: "token-exchange auth (GitHub App
// JWT->installation token)") — mirroring github's own AuthHook shape
// (Basic/JWT in, Bearer out) exactly, just with a simpler exchange request.
// The engine's only two built-in auth modes that could conceivably fit
// ("basic" and "oauth2_client_credentials") do not: "basic" sends Basic on
// every request rather than exchanging it once for a Bearer token, and
// "oauth2_client_credentials" always POSTs a
// grant_type/client_id/client_secret/scope form body (engine/auth.go's
// buildOAuth2ClientCredentials) — Jamf Pro's exchange sends no body at all
// and uses HTTP Basic instead of form-encoded client credentials. Neither
// mode can be coerced into this shape without changing the wire request,
// so a custom AuthHook is the correct, minimal escape hatch (well under the
// ~300-line soft target, one hook interface).
package jamfpro

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("jamf-pro", func() engine.Hooks { return New() })
}

// Hooks implements engine.AuthHook for the jamf-pro bundle. Its only state is
// test-injection overrides (mirrors hooks/keka's Hooks shape); every method
// is otherwise a pure function of its arguments (the returned Authenticator
// carries its own token cache, mirroring legacy's per-Read fetchToken reuse
// across pages, extended here to a time-based cache like keka/github).
type Hooks struct {
	// Client overrides the HTTP client used for the token exchange; nil uses
	// http.DefaultClient.
	Client *http.Client
	// Now is injectable for tests; nil uses time.Now.
	Now func() time.Time
}

// New returns a fresh jamf-pro Hooks value.
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "jamf-pro" }

var (
	_ engine.Hooks    = (*Hooks)(nil)
	_ engine.AuthHook = (*Hooks)(nil)
)

// Authenticator builds a token-caching Bearer authenticator using Jamf
// Pro's Basic-credential token exchange (matches legacy's fetchToken
// exactly: POST Basic(username,password) to /v1/auth/token with no request
// body, decode {"token":...} from the JSON response, use it as
// Authorization: Bearer on every subsequent request). ctx is honored so a
// caller cancellation aborts an in-flight token fetch (F8-equivalent: the
// real caller context is threaded through, never context.Background()).
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.Config["base_url"]), "/")
	if baseURL == "" {
		return nil, errors.New("jamf-pro connector requires config base_url")
	}
	username := strings.TrimSpace(cfg.Config["username"])
	if username == "" {
		return nil, errors.New("jamf-pro connector requires config username")
	}
	password := ""
	if cfg.Secrets != nil {
		password = cfg.Secrets["password"]
	}
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("jamf-pro connector requires secret password")
	}

	return &tokenAuth{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		Client:   h.Client,
		Now:      h.Now,
	}, nil
}

// tokenAuth implements connsdk.Authenticator using Jamf Pro's Basic-credential
// token exchange. It is a direct port of legacy's fetchToken
// (internal/connectors/jamf-pro/jamf_pro.go): the fetched bearer token is
// cached until shortly before its declared expiry; the username/password and
// the token itself are never logged.
type tokenAuth struct {
	BaseURL  string
	Username string
	Password string
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

// jamfTokenPath mirrors the legacy connector's jamfTokenPath constant.
const jamfTokenPath = "/v1/auth/token"

func (a *tokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races (matches keka/github's
	// own early-refresh margin).
	if a.token != "" && a.now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.BaseURL+jamfTokenPath, nil)
	if err != nil {
		return "", fmt.Errorf("jamf-pro token exchange: build request: %w", err)
	}
	req.SetBasicAuth(a.Username, a.Password)
	req.Header.Set("Accept", "application/json")

	client := a.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("jamf-pro token exchange: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("jamf-pro token exchange returned status %d", resp.StatusCode)
	}

	var out struct {
		Token   string `json:"token"`
		Expires string `json:"expires"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("jamf-pro token exchange: decode response: %w", err)
	}
	if strings.TrimSpace(out.Token) == "" {
		return "", errors.New("jamf-pro token exchange returned an empty token")
	}

	a.token = out.Token
	// expires is an RFC3339-ish timestamp in Jamf Pro's real response
	// ("2099-01-01T00:00:00.000Z"); parse it when present, falling back to a
	// conservative 1-hour TTL (matches legacy's own behavior of not parsing
	// "expires" at all and simply re-fetching once per Read/Check call —
	// caching here is strictly an efficiency improvement, never a
	// correctness change, since a stale-but-still-valid token is still
	// accepted by Jamf Pro and an expired one fails loudly on the next
	// request, exactly like a fresh per-call fetch would surface).
	ttl := time.Hour
	for _, layout := range []string{"2006-01-02T15:04:05.000Z", time.RFC3339} {
		if parsed, err := time.Parse(layout, out.Expires); err == nil {
			if d := time.Until(parsed); d > 0 {
				ttl = d
			}
			break
		}
	}
	a.expires = a.now().Add(ttl)
	return a.token, nil
}
