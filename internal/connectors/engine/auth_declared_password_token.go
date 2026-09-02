package engine

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
	"sync"

	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/credential"
)

const declaredPasswordTokenMaxResponseBytes = 64 << 10

// declaredPasswordTokenAuthenticator is intentionally narrower than an OAuth
// password grant: it sends one fixed JSON object to one fixed HTTPS route,
// extracts only the top-level accessToken value, and never refreshes or replays.
type declaredPasswordTokenAuthenticator struct {
	tokenURL string
	email    string
	password string

	mu          sync.Mutex
	accessToken string
}

func buildDeclaredPasswordTokenAuthenticator(spec AuthSpec, vars Vars) (connsdk.Authenticator, error) {
	if strings.Contains(spec.TokenURL, "{{") || strings.Contains(spec.TokenURL, "}}") {
		return nil, errors.New("declared_password_token: token_url must be static")
	}
	tokenURL, err := url.Parse(spec.TokenURL)
	if err != nil || tokenURL.Scheme != "https" || tokenURL.Host == "" || tokenURL.RawQuery != "" || tokenURL.Fragment != "" {
		return nil, errors.New("declared_password_token: token_url must be one fixed HTTPS route")
	}
	email, err := interpolateCredential(spec.Username, vars)
	if err != nil {
		return nil, fmt.Errorf("declared_password_token: email: %w", err)
	}
	if err := credential.RequireAuthenticationValue("email", email); err != nil {
		return nil, fmt.Errorf("declared_password_token: %w", err)
	}
	password, err := interpolateCredential(spec.Password, vars)
	if err != nil {
		return nil, fmt.Errorf("declared_password_token: password: %w", err)
	}
	if err := credential.RequireAuthenticationValue("password", password); err != nil {
		return nil, fmt.Errorf("declared_password_token: %w", err)
	}
	return &declaredPasswordTokenAuthenticator{tokenURL: tokenURL.String(), email: email, password: password}, nil
}

func (a *declaredPasswordTokenAuthenticator) Apply(ctx context.Context, request *http.Request) error {
	token, err := a.token(ctx)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *declaredPasswordTokenAuthenticator) token(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.accessToken != "" {
		return a.accessToken, nil
	}
	body, err := json.Marshal(struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{Email: a.email, Password: a.password})
	if err != nil {
		return "", fmt.Errorf("declared_password_token: encode request")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("declared_password_token: build request")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("declared_password_token: request token")
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("declared_password_token: token endpoint returned %s", response.Status)
	}
	var payload struct {
		AccessToken string `json:"accessToken"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, declaredPasswordTokenMaxResponseBytes)).Decode(&payload); err != nil {
		return "", errors.New("declared_password_token: decode token response")
	}
	if err := credential.RequireAuthenticationValue("access token", payload.AccessToken); err != nil {
		return "", fmt.Errorf("declared_password_token: %w", err)
	}
	a.accessToken = payload.AccessToken
	return a.accessToken, nil
}
