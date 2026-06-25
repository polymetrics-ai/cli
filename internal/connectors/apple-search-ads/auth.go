package applesearchads

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

// clientCredentialsAuth implements the OAuth2 client-credentials grant used by
// the Apple Search Ads API. It exchanges client_id/client_secret for a
// short-lived access token at Apple's token endpoint, caches it, and refreshes
// it shortly before expiry. It satisfies connsdk.Authenticator.
//
// Secret values (client_secret, the access token) are never logged; only opaque
// errors and HTTP statuses are surfaced.
type clientCredentialsAuth struct {
	tokenURL     string
	clientID     string
	clientSecret string
	scope        string
	// client is the HTTP client for the token exchange. Defaults to a 30s client.
	client *http.Client
	// now is injectable for tests. Defaults to time.Now.
	now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

// Apply ensures a fresh access token and sets the Authorization header.
func (a *clientCredentialsAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *clientCredentialsAuth) clock() time.Time {
	if a.now != nil {
		return a.now()
	}
	return time.Now()
}

func (a *clientCredentialsAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Refresh 60s before expiry to avoid edge races.
	if a.token != "" && a.clock().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.tokenURL) == "" {
		return "", errors.New("apple-search-ads oauth: token endpoint is required")
	}
	if strings.TrimSpace(a.clientID) == "" || strings.TrimSpace(a.clientSecret) == "" {
		return "", errors.New("apple-search-ads oauth: client_id and client_secret are required")
	}

	// Apple's default token endpoint carries grant_type and scope as query
	// params, but the grant also accepts them in the form body. Send them in the
	// body so the exchange works regardless of whether the configured endpoint
	// already includes the query params.
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	scope := strings.TrimSpace(a.scope)
	if scope == "" {
		scope = "searchadsorg"
	}
	form.Set("scope", scope)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("apple-search-ads oauth: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("apple-search-ads oauth: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("apple-search-ads oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("apple-search-ads oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("apple-search-ads oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.clock().Add(ttl)
	return a.token, nil
}
