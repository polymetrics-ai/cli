// Package amazonads is the Tier-2 escape hatch for the amazon-ads defs
// bundle: a Login with Amazon (LWA) refresh_token->access_token AuthHook
// (ports legacy's refreshTokenAuth from
// internal/connectors/amazon-ads/auth.go), plus the conditional
// Amazon-Advertising-API-Scope header legacy attaches only to
// profile-scoped streams (internal/connectors/amazon-ads/streams.go's
// streamEndpoint.scoped table: every stream except profiles).
//
// Amazon Ads sends the exchanged access_token as Authorization: Bearer,
// same destination header the declarative oauth2_client_credentials mode
// already knows how to write — but the GRANT is refresh_token, not
// client_credentials, which the engine dialect has no mode for (the exact
// token-exchange auth Tier-2 trigger conventions.md §1 names, ported here
// like amazon-seller-partner's LWA hook, and github's JWT->installation
// token hook). Unlike amazon-seller-partner (destination header
// x-amz-access-token) this connector's destination is the ordinary
// Authorization header, so this hook additionally has to fold in the
// conditional Scope header itself: the engine's declarative
// base.headers is resolved exactly ONCE per Runtime (read.go's
// newRuntime), shared across whichever single stream a Read call targets,
// with no per-stream header override in the dialect (StreamSpec has no
// Headers field) — so a header that must be present for 4 streams and
// absent for a 5th cannot be expressed as a declarative base.headers entry.
// The returned Authenticator's Apply runs LAST in the request pipeline
// (connsdk/http.go's do: applyHeaders then Auth.Apply), after
// DefaultHeaders are already set, and receives the fully-built
// *http.Request — so it can inspect req.URL.Path to decide whether this
// particular request is the (unscoped) profiles endpoint, mirroring
// legacy's per-endpoint scoped flag exactly, without a 2nd hook interface.
//
// Secret values (client_secret, refresh_token, the access token) are never
// logged.
package amazonads

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

// defaultTokenURL matches the bundle's spec.json default for token_url;
// used only if that config key somehow resolves empty.
const defaultTokenURL = "https://api.amazon.com/auth/o2/token"

// unscopedPath is the one Amazon Ads v2 entity endpoint that is NOT scoped
// to a profile (it enumerates the profiles/scopes themselves) — matches
// legacy's amazonAdsStreamEndpoints["profiles"].scoped == false.
const unscopedPath = "v2/profiles"

func init() {
	engine.RegisterHooks("amazon-ads", func() engine.Hooks { return New() })
}

// Hooks is the amazon-ads bundle's stateless Tier-2 hook set.
type Hooks struct {
	// Client overrides the HTTP client used for the LWA token exchange. Left
	// nil in production; injectable for tests.
	Client *http.Client
}

// New returns a fresh Hooks value.
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "amazon-ads" }

var (
	_ engine.Hooks    = (*Hooks)(nil)
	_ engine.AuthHook = (*Hooks)(nil)
)

// Authenticator exchanges the configured LWA refresh_token for a
// short-lived access_token (matches legacy's refreshTokenAuth.accessToken)
// and returns an Authenticator that sets Authorization: Bearer <token> on
// every request, plus Amazon-Advertising-API-Scope: <profile_id> on every
// request EXCEPT the unscoped profiles endpoint — matching legacy's
// requester(cfg, endpoint.scoped) gate. ctx is honored (a real network
// call). Unlike legacy's refreshTokenAuth, this performs no in-process
// caching across calls: the engine calls Authenticator once per Read/Check
// (engine/read.go's newRuntime/selectAuth), and that single returned
// Authenticator instance is reused for every page within that one call, so
// one Read still performs exactly one token exchange — matching legacy's
// own per-Read requester() construction. See docs.md's "Auth setup"
// section for the documented, non-data-changing deviation this implies for
// a long-lived process issuing many separate Read calls.
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	tokenURL := strings.TrimSpace(cfg.Config["token_url"])
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}
	clientID := strings.TrimSpace(cfg.Secrets["client_id"])
	if clientID == "" {
		return nil, errors.New("amazon-ads: secret client_id is required")
	}
	clientSecret := strings.TrimSpace(cfg.Secrets["client_secret"])
	if clientSecret == "" {
		return nil, errors.New("amazon-ads: secret client_secret is required")
	}
	refreshToken := strings.TrimSpace(cfg.Secrets["refresh_token"])
	if refreshToken == "" {
		return nil, errors.New("amazon-ads: secret refresh_token is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("amazon-ads: build LWA token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("amazon-ads: LWA token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("amazon-ads: LWA token endpoint returned status %d", resp.StatusCode)
	}

	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("amazon-ads: decode LWA token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return nil, errors.New("amazon-ads: LWA token response did not include access_token")
	}

	return &scopedBearer{
		accessToken: out.AccessToken,
		profileID:   strings.TrimSpace(cfg.Config["profile_id"]),
	}, nil
}

// scopedBearer sets Authorization: Bearer <token> on every request, plus
// Amazon-Advertising-API-Scope: <profileID> on every request EXCEPT the
// unscoped profiles endpoint (matches legacy streams.go's per-endpoint
// scoped flag). It satisfies connsdk.Authenticator.
type scopedBearer struct {
	accessToken string
	profileID   string
}

func (a *scopedBearer) Apply(_ context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.accessToken)
	if strings.HasSuffix(strings.TrimRight(req.URL.Path, "/"), unscopedPath) {
		return nil
	}
	// Every other v2 entity endpoint is profile-scoped (matches legacy's
	// requester(cfg, scoped=true) gate): a missing profile_id is a hard
	// error here, never a silently-unscoped request.
	if a.profileID == "" {
		return errors.New("amazon-ads config profile_id is required for profile-scoped streams")
	}
	req.Header.Set("Amazon-Advertising-API-Scope", a.profileID)
	return nil
}
