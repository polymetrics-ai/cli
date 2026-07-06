package basecamp

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
)

// oauthRefreshAuth implements connsdk.Authenticator using the Basecamp /
// 37signals Launchpad OAuth2 refresh grant. Unlike a standard OAuth2
// refresh_token grant (and unlike connsdk.OAuth2ClientCredentials), Launchpad's
// token endpoint expects `type=refresh` form parameters rather than
// `grant_type=refresh_token`, so this small authenticator lives in-package. It
// exchanges the long-lived refresh token for a short-lived access token, caches
// it, and refreshes before expiry. Secret values are never logged.
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

	// Honor a still-valid cached token (refresh 60s before expiry).
	if a.token != "" && a.timeNow().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}

	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("basecamp oauth: token_url is required")
	}

	// Launchpad's refresh grant: type=refresh, not grant_type=refresh_token.
	form := url.Values{}
	form.Set("type", "refresh")
	form.Set("refresh_token", a.refreshToken)
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("basecamp oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("basecamp oauth: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("basecamp oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("basecamp oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("basecamp oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	// Basecamp access tokens currently have a 2-week lifetime; honor the
	// returned expires_in but default conservatively if absent.
	ttl := 14 * 24 * time.Hour
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}
