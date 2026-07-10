package connsdk

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

// Authenticator applies credentials to an outgoing request. Implementations must
// never log secret values.
type Authenticator interface {
	Apply(ctx context.Context, req *http.Request) error
}

// AuthFunc adapts a function to the Authenticator interface.
type AuthFunc func(ctx context.Context, req *http.Request) error

func (f AuthFunc) Apply(ctx context.Context, req *http.Request) error { return f(ctx, req) }

// staticHeader sets a fixed request header.
type staticHeader struct {
	name  string
	value string
}

func (a staticHeader) Apply(_ context.Context, req *http.Request) error {
	if a.value != "" {
		req.Header.Set(a.name, a.value)
	}
	return nil
}

// Bearer authenticates with an Authorization: Bearer <token> header.
func Bearer(token string) Authenticator {
	return staticHeader{name: "Authorization", value: "Bearer " + strings.TrimSpace(token)}
}

// APIKeyHeader authenticates with an arbitrary header (e.g. X-API-Key: <value>).
// An optional prefix is prepended to the value (e.g. "Token ").
func APIKeyHeader(header, value, prefix string) Authenticator {
	return staticHeader{name: header, value: prefix + strings.TrimSpace(value)}
}

// Basic authenticates with HTTP Basic auth.
func Basic(username, password string) Authenticator {
	creds := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return staticHeader{name: "Authorization", value: "Basic " + creds}
}

// apiKeyQuery authenticates by adding a query parameter to the request URL.
type apiKeyQuery struct {
	param string
	value string
}

func (a apiKeyQuery) Apply(_ context.Context, req *http.Request) error {
	q := req.URL.Query()
	q.Set(a.param, a.value)
	req.URL.RawQuery = q.Encode()
	return nil
}

// APIKeyQuery authenticates by adding ?param=value to every request.
func APIKeyQuery(param, value string) Authenticator {
	return apiKeyQuery{param: param, value: strings.TrimSpace(value)}
}

// OAuth2ClientCredentials fetches and caches a bearer token using the OAuth2
// client-credentials grant, refreshing automatically before expiry.
type OAuth2ClientCredentials struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scopes       []string
	// ExtraParams are added to the token request form (e.g. audience).
	ExtraParams url.Values
	// Client is used for the token request. Defaults to a 30s client.
	Client *http.Client
	// Now is injectable for tests. Defaults to time.Now.
	Now func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *OAuth2ClientCredentials) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

// Apply ensures a fresh token and sets the Authorization header.
func (a *OAuth2ClientCredentials) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *OAuth2ClientCredentials) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Refresh 60s before expiry to avoid edge races.
	if a.token != "" && a.now().Add(60*time.Second).Before(a.expires) {
		return a.token, nil
	}
	if strings.TrimSpace(a.TokenURL) == "" {
		return "", errors.New("oauth2: TokenURL is required")
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", a.ClientID)
	form.Set("client_secret", a.ClientSecret)
	if len(a.Scopes) > 0 {
		form.Set("scope", strings.Join(a.Scopes, " "))
	}
	for k, vs := range a.ExtraParams {
		for _, v := range vs {
			form.Add(k, v)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("oauth2: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("oauth2: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("oauth2: token endpoint returned %s", resp.Status)
	}

	var out struct {
		AccessToken string      `json:"access_token"`
		TokenType   string      `json:"token_type"`
		ExpiresIn   json.Number `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("oauth2: decode token response: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return "", errors.New("oauth2: token response missing access_token")
	}

	a.token = out.AccessToken
	ttl := 3600 * time.Second
	if secs, err := out.ExpiresIn.Int64(); err == nil && secs > 0 {
		ttl = time.Duration(secs) * time.Second
	}
	a.expires = a.now().Add(ttl)
	return a.token, nil
}
