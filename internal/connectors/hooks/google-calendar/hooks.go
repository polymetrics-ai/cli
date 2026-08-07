// Package googlecalendar provides Google Calendar bundle hooks.
package googlecalendar

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

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("google-calendar", func() engine.Hooks { return New() })
}

// Hooks implements the custom OAuth2 refresh-token AuthHook required by
// Google Calendar's declarative HTTP bundle. It intentionally does not
// override Check or ReadStream; fixture replay and live reads use the engine's
// normal declarative request path.
type Hooks struct {
	Client *http.Client
	Now    func() time.Time
}

// New returns a hook set for google-calendar.
func New() engine.Hooks { return Hooks{} }

func (Hooks) ConnectorName() string { return "google-calendar" }

func (h Hooks) Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error) {
	if fixtureAuth(cfg) {
		return connsdk.AuthFunc(func(context.Context, *http.Request) error { return nil }), nil
	}

	vars := engine.Vars{Config: cfg.Config, Secrets: cfg.Secrets}
	tokenURL, err := engine.Interpolate(spec.TokenURL, vars)
	if err != nil {
		return nil, fmt.Errorf("google-calendar oauth: token_url: %w", err)
	}
	clientID, err := engine.Interpolate(spec.ClientID, vars)
	if err != nil {
		return nil, fmt.Errorf("google-calendar oauth: client_id: %w", err)
	}
	clientSecret, err := engine.Interpolate(spec.ClientSecret, vars)
	if err != nil {
		return nil, fmt.Errorf("google-calendar oauth: client_secret: %w", err)
	}
	refreshToken, err := engine.Interpolate(spec.Token, vars)
	if err != nil {
		return nil, fmt.Errorf("google-calendar oauth: refresh token: %w", err)
	}

	if err := validateTokenEndpoint(tokenURL); err != nil {
		return nil, err
	}
	if strings.TrimSpace(clientID) == "" {
		return nil, errors.New("google-calendar oauth: client_id is required")
	}
	if strings.TrimSpace(clientSecret) == "" {
		return nil, errors.New("google-calendar oauth: client_secret is required")
	}
	if strings.TrimSpace(refreshToken) == "" {
		return nil, errors.New("google-calendar oauth: refresh token is required")
	}

	return &refreshTokenAuthenticator{
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		client:       h.Client,
		now:          h.Now,
	}, nil
}

func fixtureAuth(cfg connectors.RuntimeConfig) bool {
	return cfg.ProjectDir == "__polymetrics_conformance_fixture__"
}

func validateTokenEndpoint(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("google-calendar oauth: token_url must be an absolute URL")
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme == "http" && loopbackHost(u.Hostname()) {
		return nil
	}
	return fmt.Errorf("google-calendar oauth: token_url must use https")
}

func loopbackHost(host string) bool {
	switch strings.ToLower(host) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

type refreshTokenAuthenticator struct {
	tokenURL     string
	clientID     string
	clientSecret string
	refreshToken string
	client       *http.Client
	now          func() time.Time

	mu      sync.Mutex
	token   string
	expires time.Time
}

func (a *refreshTokenAuthenticator) Apply(ctx context.Context, req *http.Request) error {
	token, err := a.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *refreshTokenAuthenticator) accessToken(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now
	if a.now != nil {
		now = a.now
	}
	if a.token != "" && now().Add(time.Minute).Before(a.expires) {
		return a.token, nil
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", a.clientID)
	form.Set("client_secret", a.clientSecret)
	form.Set("refresh_token", a.refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("google-calendar oauth: build refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("google-calendar oauth: refresh request failed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
	}
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&payload); err != nil {
		return "", fmt.Errorf("google-calendar oauth: parse refresh response: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		if payload.Error != "" {
			return "", fmt.Errorf("google-calendar oauth: refresh failed: %s", payload.Error)
		}
		return "", fmt.Errorf("google-calendar oauth: refresh failed with status %d", res.StatusCode)
	}
	if strings.TrimSpace(payload.AccessToken) == "" {
		return "", errors.New("google-calendar oauth: refresh response missing access token")
	}
	if payload.ExpiresIn <= 0 {
		payload.ExpiresIn = 3600
	}

	a.token = payload.AccessToken
	a.expires = now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	return a.token, nil
}
