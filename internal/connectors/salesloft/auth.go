package salesloft

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

// oauthRefreshAuth implements connsdk.Authenticator using the OAuth2
// refresh-token grant that Salesloft uses (grant_type=refresh_token). connsdk
// ships an OAuth2ClientCredentials authenticator, but Salesloft refreshes a
// long-lived refresh_token rather than using client_credentials, so this small
// authenticator lives in-package. It caches the access token and refreshes it
// before expiry. Secret values are never logged.
type oauthRefreshAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	// seedToken, when set, is used until it expires/refreshes so an already-valid
	// access_token from config is honored without an immediate network refresh.
	seedToken string
	client    *http.Client
	// now is injectable for tests; defaults to time.Now.
	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
	seeded  bool
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
	// Honor a config-provided seed access token exactly once, before any refresh.
	if !a.seeded && a.seedToken != "" {
		a.seeded = true
		a.token = a.seedToken
		a.expires = a.timeNow().Add(50 * time.Minute)
		return a.token, nil
	}

	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("salesloft oauth: token_url is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("refresh_token", a.refreshToken)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("salesloft oauth: build token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("salesloft oauth: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("salesloft oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken  string      `json:"access_token"`
		RefreshToken string      `json:"refresh_token"`
		TokenType    string      `json:"token_type"`
		ExpiresIn    json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("salesloft oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("salesloft oauth: token response missing access_token")
	}
	// Salesloft rotates the refresh token; keep the latest for subsequent refreshes.
	if strings.TrimSpace(out.RefreshToken) != "" {
		a.refreshToken = out.RefreshToken
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.timeNow().Add(ttl)
	return a.token, nil
}
