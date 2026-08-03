// Package reddit is the Tier-2 escape hatch: a refresh_token->access_token
// exchange AuthHook. Reddit's client_credentials grant (already supported
// declaratively by the engine's oauth2_client_credentials auth mode) mints
// an "Application Only" token that Reddit's own docs state can never act on
// behalf of a user (https://github.com/reddit-archive/reddit/wiki/OAuth2#application-only-oauth) --
// no user-context endpoint (moderation, private messages, votes, saves, or
// anything scoped to a specific account) is reachable with it. The
// refresh_token grant is the only way to keep a durable, user-context
// bearer token alive past Reddit's 1-hour access-token expiry, and the
// engine has no built-in support for it, so this hook ports the exchange
// (mirrors hooks/github's JWT->installation-token pattern).
package reddit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("reddit", func() engine.Hooks { return New() })
}

// Hooks is the reddit bundle's stateless Tier-2 hook set.
type Hooks struct{}

// New returns a fresh Hooks value.
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "reddit" }

var (
	_ engine.Hooks    = (*Hooks)(nil)
	_ engine.AuthHook = (*Hooks)(nil)
)

const defaultTokenURL = "https://www.reddit.com/api/v1/access_token"

// Authenticator exchanges the configured refresh_token for a fresh
// access_token via POST {token_url} grant_type=refresh_token (Reddit OAuth2
// "Refreshing the token": https://github.com/reddit-archive/reddit/wiki/OAuth2#refreshing-the-token),
// then returns a Bearer authenticator wrapping it. ctx is honored (a real
// network call). Uncached, matching hooks/github's own re-mint-on-every-call
// behavior -- this hook is invoked once per Read/Write/Check call
// (engine/read.go's newRuntime), so a fresh token is minted at the start of
// every sync/command rather than being reused past its 1-hour lifetime.
func (h *Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, _ engine.AuthSpec) (connsdk.Authenticator, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	refreshToken := strings.TrimSpace(cfg.Secrets["refresh_token"])
	if refreshToken == "" {
		return nil, errors.New("reddit: refresh_token exchange requires secrets.refresh_token")
	}
	clientID := strings.TrimSpace(cfg.Secrets["client_id"])
	if clientID == "" {
		return nil, errors.New("reddit: refresh_token exchange requires secrets.client_id")
	}
	// client_secret may legitimately be empty: Reddit's installed
	// (non-confidential) app type has no secret and documents sending an
	// empty string as the HTTP Basic password in that case.
	clientSecret := cfg.Secrets["client_secret"]

	tokenURL := strings.TrimSpace(cfg.Config["token_url"])
	if tokenURL == "" {
		tokenURL = defaultTokenURL
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader([]byte(form.Encode())))
	if err != nil {
		return nil, fmt.Errorf("reddit: build refresh_token request: %w", err)
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent(cfg))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reddit: refresh_token exchange: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("reddit: refresh_token exchange returned status %d", resp.StatusCode)
	}
	if readErr != nil {
		return nil, fmt.Errorf("reddit: read refresh_token response: %w", readErr)
	}
	var out struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("reddit: decode refresh_token response: %w", err)
	}
	if out.Error != "" {
		return nil, fmt.Errorf("reddit: refresh_token exchange rejected: %s", out.Error)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return nil, errors.New("reddit: refresh_token response did not include access_token")
	}
	return connsdk.Bearer(out.AccessToken), nil
}

// userAgent mirrors streams.json's declarative User-Agent template
// (<platform>:<app ID>:<version> (by /u/<reddit_username>)) for the token
// exchange request itself, which is not routed through the declarative
// header pipeline. Reddit rate-limits non-conforming User-Agents on every
// endpoint, including the token endpoint.
func userAgent(cfg connectors.RuntimeConfig) string {
	username := strings.TrimSpace(cfg.Config["reddit_username"])
	if username == "" {
		username = "unknown"
	}
	return "go:ai.polymetrics.cli:v1 (by /u/" + username + ")"
}
