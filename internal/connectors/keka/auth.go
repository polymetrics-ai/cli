package keka

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

// kekaTokenAuth implements the Keka OAuth2 client-credentials token exchange. It
// differs from connsdk.OAuth2ClientCredentials because Keka uses a custom
// grant_type ("kekaapi") and requires the api_key form field, neither of which
// the generic helper sends. The fetched bearer token is cached until shortly
// before expiry. Secret values (client_secret, api_key, the token itself) are
// never logged.
type kekaTokenAuth struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	APIKey       string
	GrantType    string
	Scope        string
	// Client is used for the token request. Defaults to a 30s client.
	Client *http.Client
	// Now is injectable for tests. Defaults to time.Now.
	Now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *kekaTokenAuth) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

// Apply ensures a fresh token and sets the Authorization header.
func (a *kekaTokenAuth) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *kekaTokenAuth) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races.
	if a.token != "" && a.now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.TokenURL) == "" {
		return "", errors.New("keka oauth: token_url is required")
	}

	form := url.Values{}
	form.Set("grant_type", valueOr(a.GrantType, kekaDefaultGrantType))
	form.Set("scope", valueOr(a.Scope, kekaDefaultScope))
	form.Set("client_id", a.ClientID)
	form.Set("client_secret", a.ClientSecret)
	if strings.TrimSpace(a.APIKey) != "" {
		form.Set("api_key", a.APIKey)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("keka oauth: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("keka oauth: token request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("keka oauth: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("keka oauth: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("keka oauth: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.now().Add(ttl)
	return a.token, nil
}
