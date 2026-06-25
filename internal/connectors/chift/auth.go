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
)

// sessionTokenAuth implements connsdk.Authenticator for Chift's session-token
// exchange. Chift does not use a standard form-encoded OAuth2 client-credentials
// grant: instead a POST to <base>/token with a JSON body of
// {clientId, clientSecret, accountId} returns {access_token, expires_in}, which
// is then sent as a Bearer token on data requests.
//
// The fetched token is cached and refreshed automatically shortly before
// expiry. Credentials and the token value are never logged.
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
	defer resp.Body.Close()
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
