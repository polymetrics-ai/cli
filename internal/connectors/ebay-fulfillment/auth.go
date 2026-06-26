package ebayfulfillment

import (
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
)

// refreshTokenAuth applies eBay OAuth2 access tokens to outgoing requests,
// exchanging the configured refresh token for a short-lived access token via the
// identity token endpoint and caching it until shortly before expiry.
//
// eBay's token endpoint authenticates the request with HTTP Basic
// (client_id:client_secret) and takes grant_type=refresh_token in the form body.
// Secrets (refresh token, client secret, access token) are never logged.
type refreshTokenAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	scopes       string
	client       *http.Client

	// Now is injectable for tests. Defaults to time.Now.
	Now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *refreshTokenAuth) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

// Apply ensures a fresh access token and sets the Authorization header.
func (a *refreshTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *refreshTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Reuse the cached token until 60s before expiry to avoid edge races.
	if a.token != "" && a.now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("ebay-fulfillment: refresh_token_endpoint is required")
	}
	if strings.TrimSpace(a.refreshToken) == "" {
		return "", errors.New("ebay-fulfillment: refresh_token is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", a.refreshToken)
	if a.scopes != "" {
		form.Set("scope", a.scopes)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("ebay-fulfillment: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	// eBay authenticates the token call with HTTP Basic client_id:client_secret.
	if a.clientID != "" || a.clientSecret != "" {
		creds := base64.StdEncoding.EncodeToString([]byte(a.clientID + ":" + a.clientSecret))
		req.Header.Set("Authorization", "Basic "+creds)
	}

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ebay-fulfillment: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ebay-fulfillment: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("ebay-fulfillment: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("ebay-fulfillment: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 7200 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.now().Add(ttl)
	return a.token, nil
}
