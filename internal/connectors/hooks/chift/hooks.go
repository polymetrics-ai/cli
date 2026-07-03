// Package chift is the Tier-2 escape hatch for the chift bundle: an
// AuthHook implementing Chift's session-token exchange (ports
// internal/connectors/chift/auth.go's sessionTokenAuth). Chift does not use
// a standard form-encoded OAuth2 client-credentials grant — a POST to
// <base_url>/token with a JSON body of {clientId, clientSecret, accountId}
// returns {access_token, expires_in}, which is then carried as a Bearer
// token on data requests. The engine's declarative
// "oauth2_client_credentials" auth mode always builds a form-encoded
// request with the standard grant_type/client_id/client_secret/scope field
// names (connsdk.OAuth2ClientCredentials) and cannot express Chift's JSON
// body or non-standard field names — a legitimate Tier-2 token-exchange-auth
// trigger (docs/migration/conventions.md §1).
package chift

import (
	"bytes"
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
	engine.RegisterHooks("chift", func() engine.Hooks { return New() })
}

// Hooks is the chift bundle's stateless Tier-2 hook set. Authenticator
// returns a fresh *sessionTokenAuth per call, which internally caches and
// refreshes its own minted token for the lifetime of that authenticator
// instance (mirrors legacy's per-Connector-call sessionTokenAuth lifecycle).
type Hooks struct{}

// New returns a fresh Hooks value.
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "chift" }

var (
	_ engine.Hooks    = (*Hooks)(nil)
	_ engine.AuthHook = (*Hooks)(nil)
)

// Authenticator builds a connsdk.Authenticator that mints and caches a
// Chift session access token (matches legacy's sessionTokenAuth): a POST to
// <base_url>/token with a JSON body of clientId/clientSecret/accountId
// (resolved from cfg.Secrets' client_id/client_secret/account_id — spec.json
// marks all three x-secret), returning {access_token, expires_in}. The
// resolved base URL comes from cfg.Config["base_url"], matching every other
// declarative field's config resolution (falls back to the spec.json
// default "https://api.chift.eu" when materializeConfigDefaults has already
// filled it in by the time hooks run).
func (h *Hooks) Authenticator(_ context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	creds := chiftCredentials{
		clientID:     strings.TrimSpace(cfg.Secrets["client_id"]),
		clientSecret: strings.TrimSpace(cfg.Secrets["client_secret"]),
		accountID:    strings.TrimSpace(cfg.Secrets["account_id"]),
	}
	if err := creds.validate(); err != nil {
		return nil, err
	}
	baseURL := strings.TrimSpace(cfg.Config["base_url"])
	if baseURL == "" {
		baseURL = "https://api.chift.eu"
	}
	return newSessionTokenAuth(baseURL, creds, nil), nil
}

// chiftCredentials carries the three required secrets for the token
// exchange.
type chiftCredentials struct {
	clientID     string
	clientSecret string
	accountID    string
}

func (cr chiftCredentials) validate() error {
	var missing []string
	if cr.clientID == "" {
		missing = append(missing, "client_id")
	}
	if cr.clientSecret == "" {
		missing = append(missing, "client_secret")
	}
	if cr.accountID == "" {
		missing = append(missing, "account_id")
	}
	if len(missing) > 0 {
		return fmt.Errorf("chift connector requires secret(s): %s", strings.Join(missing, ", "))
	}
	return nil
}

// sessionTokenAuth implements connsdk.Authenticator for Chift's session-token
// exchange, porting internal/connectors/chift/auth.go verbatim: the fetched
// token is cached and refreshed automatically shortly (60s) before expiry.
// Credentials and the token value are never logged.
type sessionTokenAuth struct {
	tokenURL string
	creds    chiftCredentials
	client   *http.Client
	// now is injectable for tests; defaults to time.Now.
	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func newSessionTokenAuth(baseURL string, creds chiftCredentials, client *http.Client) *sessionTokenAuth {
	return &sessionTokenAuth{
		tokenURL: strings.TrimRight(baseURL, "/") + "/token",
		creds:    creds,
		client:   client,
	}
}

func (a *sessionTokenAuth) clock() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *sessionTokenAuth) httpClient() *http.Client {
	if a.client != nil {
		return a.client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// Apply ensures a fresh token and sets the Authorization header.
func (a *sessionTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *sessionTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races.
	if a.token != "" && a.clock().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if err := a.creds.validate(); err != nil {
		return "", err
	}

	payload, err := json.Marshal(map[string]string{
		"clientId":     a.creds.clientID,
		"clientSecret": a.creds.clientSecret,
		"accountId":    a.creds.accountID,
	})
	if err != nil {
		return "", fmt.Errorf("chift token: encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("chift token: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.httpClient().Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("chift token: request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("chift token: endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("chift token: decode response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("chift token: response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.clock().Add(ttl)
	return a.token, nil
}
