// Package amazonsellerpartner is the Tier-2 escape hatch for the
// amazon-seller-partner defs bundle: a Login with Amazon (LWA)
// refresh_token->access_token AuthHook (ports legacy's lwaAuthenticator from
// internal/connectors/amazon-seller-partner/auth.go). SP-API authenticates
// data requests with a short-lived access_token sent as the x-amz-access-token
// header (NOT Authorization: Bearer), exchanged from a long-lived
// refresh_token at the LWA token endpoint — this is the same class of
// token-exchange auth conventions.md §1 names for GitHub App's JWT->
// installation-token flow (see hooks/github/hooks.go), just a different grant
// shape (refresh_token) and a different destination header (x-amz-access-token
// instead of Authorization).
//
// Secret values (client secret, refresh token, access token) are never
// logged.
package amazonsellerpartner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

// defaultLWATokenURL matches the bundle's spec.json default for
// lwa_token_url; used only if that config key somehow resolves empty.
const defaultLWATokenURL = "https://api.amazon.com/auth/o2/token"

func init() {
	engine.RegisterHooks("amazon-seller-partner", func() engine.Hooks { return New() })
}

// Hooks is the amazon-seller-partner bundle's stateless Tier-2 hook set.
type Hooks struct {
	// Client overrides the HTTP client used for the LWA token exchange. Left
	// nil in production; injectable for tests.
	Client *http.Client
}

// New returns a fresh Hooks value.
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "amazon-seller-partner" }

var (
	_ engine.Hooks    = (*Hooks)(nil)
	_ engine.AuthHook = (*Hooks)(nil)
)

// Authenticator exchanges the configured LWA refresh_token for a short-lived
// access_token (matches legacy's lwaAuthenticator.accessToken) and returns a
// connsdk.APIKeyHeader authenticator that sets x-amz-access-token on every
// request. ctx is honored (a real network call). Unlike legacy's
// lwaAuthenticator, this performs no in-process caching across calls: the
// engine calls Authenticator once per Read/Check (engine/read.go's
// newRuntime/selectAuth), and that single returned Authenticator instance is
// reused for every page within that one call, so one Read still performs
// exactly one token exchange — matching legacy's own per-Read requester()
// construction. See docs.md's "Auth setup" section for the documented,
// non-data-changing deviation this implies for a long-lived process issuing
// many separate Read calls.
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	tokenURL := strings.TrimSpace(cfg.Config["lwa_token_url"])
	if tokenURL == "" {
		tokenURL = defaultLWATokenURL
	}
	clientID := strings.TrimSpace(cfg.Secrets["lwa_app_id"])
	if clientID == "" {
		return nil, errors.New("amazon-seller-partner: secret lwa_app_id is required")
	}
	clientSecret := strings.TrimSpace(cfg.Secrets["lwa_client_secret"])
	if clientSecret == "" {
		return nil, errors.New("amazon-seller-partner: secret lwa_client_secret is required")
	}
	refreshToken := strings.TrimSpace(cfg.Secrets["refresh_token"])
	if refreshToken == "" {
		return nil, errors.New("amazon-seller-partner: secret refresh_token is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("amazon-seller-partner: build LWA token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	httpReq.Header.Set("Accept", "application/json")

	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("amazon-seller-partner: LWA token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("amazon-seller-partner: LWA token endpoint returned status %d", resp.StatusCode)
	}

	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("amazon-seller-partner: decode LWA token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return nil, errors.New("amazon-seller-partner: LWA token response did not include access_token")
	}

	return connsdk.APIKeyHeader("x-amz-access-token", out.AccessToken, ""), nil
}
