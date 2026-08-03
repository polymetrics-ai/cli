// Package device implements the OAuth 2.0 Device Authorization Grant
// (RFC 8628) for hosts with no local browser to open a loopback redirect
// against — a headless server, a container, an SSH session. The user is
// shown a short code and a verification URL to open on any other device;
// this package polls the token endpoint until they finish there. It never
// sees or asks for a password, matching loopback's boundary exactly.
package device

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"polymetrics.ai/internal/browserauth"
)

// Config configures one device-flow login.
type Config struct {
	// DeviceAuthURL and TokenURL are the provider's RFC 8628 endpoints.
	DeviceAuthURL string
	TokenURL      string

	ClientID     string
	ClientSecret string // optional
	Scopes       []string

	// OnUserCode is invoked once with the user_code and verification_uri
	// (and verification_uri_complete, when the provider sends one) so the
	// caller can display them. Required — without it the user has no way to
	// complete the login.
	OnUserCode func(userCode, verificationURI, verificationURIComplete string)

	// Timeout bounds the whole polling loop. Defaults to 10 minutes if the
	// provider's own expires_in is absent or larger.
	Timeout time.Duration

	HTTPClient *http.Client
}

// Flow is a browserauth.Flow that runs one device-authorization login.
type Flow struct {
	cfg Config
}

func New(cfg Config) (*Flow, error) {
	if strings.TrimSpace(cfg.DeviceAuthURL) == "" {
		return nil, errors.New("device: device_auth_url is required")
	}
	if strings.TrimSpace(cfg.TokenURL) == "" {
		return nil, errors.New("device: token_url is required")
	}
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, errors.New("device: client_id is required")
	}
	if cfg.OnUserCode == nil {
		return nil, errors.New("device: on_user_code callback is required")
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Flow{cfg: cfg}, nil
}

func (f *Flow) Name() string { return "device_code" }

type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int64  `json:"expires_in"`
	Interval                int64  `json:"interval"`
}

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// RFC 8628 §3.5 error codes for the token polling loop.
const (
	errAuthorizationPending = "authorization_pending"
	errSlowDown             = "slow_down"
	errAccessDenied         = "access_denied"
	errExpiredToken         = "expired_token"
)

func (f *Flow) Login(ctx context.Context) (browserauth.Credential, error) {
	auth, err := f.requestDeviceAuth(ctx)
	if err != nil {
		return browserauth.Credential{}, err
	}

	f.cfg.OnUserCode(auth.UserCode, auth.VerificationURI, auth.VerificationURIComplete)

	timeout := f.cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	if auth.ExpiresIn > 0 {
		providerWindow := time.Duration(auth.ExpiresIn) * time.Second
		if providerWindow < timeout {
			timeout = providerWindow
		}
	}
	interval := time.Duration(auth.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	for {
		ticker := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			ticker.Stop()
			return browserauth.Credential{}, ctx.Err()
		case <-deadline.C:
			ticker.Stop()
			return browserauth.Credential{}, errors.New("device: timed out waiting for user to complete verification")
		case <-ticker.C:
		}

		token, pollErr := f.pollToken(ctx, auth.DeviceCode)
		switch {
		case pollErr == nil:
			return browserauth.Credential{OAuth: token}, nil
		case errors.Is(pollErr, errAuthorizationPendingSentinel):
			continue
		case errors.Is(pollErr, errSlowDownSentinel):
			interval += 5 * time.Second
			continue
		default:
			return browserauth.Credential{}, pollErr
		}
	}
}

func (f *Flow) requestDeviceAuth(ctx context.Context) (*deviceAuthResponse, error) {
	form := url.Values{}
	form.Set("client_id", f.cfg.ClientID)
	if len(f.cfg.Scopes) > 0 {
		form.Set("scope", strings.Join(f.cfg.Scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.cfg.DeviceAuthURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("device: build device authorization request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := f.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device: device authorization request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("device: read device authorization response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device: device authorization request failed: status %d: %s", resp.StatusCode, string(body))
	}

	var parsed deviceAuthResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("device: decode device authorization response: %w", err)
	}
	if parsed.DeviceCode == "" || parsed.UserCode == "" {
		return nil, errors.New("device: device authorization response missing device_code or user_code")
	}
	return &parsed, nil
}

var (
	errAuthorizationPendingSentinel = errors.New(errAuthorizationPending)
	errSlowDownSentinel             = errors.New(errSlowDown)
)

func (f *Flow) pollToken(ctx context.Context, deviceCode string) (*browserauth.OAuthCredential, error) {
	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	form.Set("device_code", deviceCode)
	form.Set("client_id", f.cfg.ClientID)
	if f.cfg.ClientSecret != "" {
		form.Set("client_secret", f.cfg.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("device: build token poll request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := f.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device: token poll request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("device: read token poll response: %w", err)
	}

	var parsed tokenResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("device: decode token poll response (status %d): %w", resp.StatusCode, err)
	}

	switch parsed.Error {
	case "":
		// fall through to success handling below
	case errAuthorizationPending:
		return nil, errAuthorizationPendingSentinel
	case errSlowDown:
		return nil, errSlowDownSentinel
	case errAccessDenied:
		return nil, errors.New("device: user denied authorization")
	case errExpiredToken:
		return nil, errors.New("device: device code expired before the user completed verification")
	default:
		return nil, fmt.Errorf("device: token poll failed: %s: %s", parsed.Error, parsed.ErrorDescription)
	}

	if parsed.AccessToken == "" {
		return nil, errors.New("device: token poll response had no access_token")
	}

	cred := &browserauth.OAuthCredential{
		AccessToken:  parsed.AccessToken,
		RefreshToken: parsed.RefreshToken,
		TokenType:    parsed.TokenType,
		TokenURL:     f.cfg.TokenURL,
		ClientID:     f.cfg.ClientID,
		ClientSecret: f.cfg.ClientSecret,
	}
	if parsed.ExpiresIn > 0 {
		cred.ExpiresAt = time.Now().Add(time.Duration(parsed.ExpiresIn) * time.Second)
	}
	if parsed.Scope != "" {
		cred.Scopes = strings.Fields(parsed.Scope)
	} else {
		cred.Scopes = f.cfg.Scopes
	}
	return cred, nil
}
