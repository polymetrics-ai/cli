package loopback_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/browserauth/loopback"
)

// fakeBrowser simulates the user completing the login in a real browser: it
// GETs the authorization URL against a fake provider auth endpoint, which
// redirects back to the loopback listener exactly like a browser would —
// no real browser or network access involved.
func fakeBrowserHittingRedirect(t *testing.T, wantError string) func(string) error {
	t.Helper()
	return func(rawURL string) error {
		go func() {
			u, err := url.Parse(rawURL)
			if err != nil {
				t.Errorf("parse authorization URL: %v", err)
				return
			}
			redirectURI := u.Query().Get("redirect_uri")
			state := u.Query().Get("state")
			cb, err := url.Parse(redirectURI)
			if err != nil {
				t.Errorf("parse redirect_uri: %v", err)
				return
			}
			q := cb.Query()
			q.Set("state", state)
			if wantError != "" {
				q.Set("error", wantError)
				q.Set("error_description", "denied by test")
			} else {
				q.Set("code", "test-auth-code")
			}
			cb.RawQuery = q.Encode()
			resp, err := http.Get(cb.String())
			if err != nil {
				t.Errorf("simulate browser redirect: %v", err)
				return
			}
			_ = resp.Body.Close()
		}()
		return nil
	}
}

func fakeTokenServer(t *testing.T, wantCode string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse token form: %v", err)
		}
		if r.PostForm.Get("code") != wantCode {
			t.Fatalf("token exchange code = %q, want %q", r.PostForm.Get("code"), wantCode)
		}
		if r.PostForm.Get("code_verifier") == "" {
			t.Fatalf("token exchange missing code_verifier")
		}
		if r.PostForm.Get("grant_type") != "authorization_code" {
			t.Fatalf("token exchange grant_type = %q", r.PostForm.Get("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "test-access-token",
			"refresh_token": "test-refresh-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "read write",
		})
	}))
}

func TestLoopbackFlowSuccess(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("auth endpoint should never be directly requested by the flow itself: %s", r.URL)
	}))
	defer authServer.Close()
	tokenServer := fakeTokenServer(t, "test-auth-code")
	defer tokenServer.Close()

	flow, err := loopback.New(loopback.Config{
		AuthURL:     authServer.URL + "/authorize",
		TokenURL:    tokenServer.URL + "/token",
		ClientID:    "client-123",
		Scopes:      []string{"read", "write"},
		OpenBrowser: fakeBrowserHittingRedirect(t, ""),
		Timeout:     5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if flow.Name() != "loopback_pkce" {
		t.Fatalf("Name() = %q", flow.Name())
	}

	cred, err := flow.Login(context.Background())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if cred.OAuth == nil {
		t.Fatalf("Login() returned no OAuth credential")
	}
	if cred.OAuth.AccessToken != "test-access-token" {
		t.Fatalf("AccessToken = %q", cred.OAuth.AccessToken)
	}
	if cred.OAuth.RefreshToken != "test-refresh-token" {
		t.Fatalf("RefreshToken = %q", cred.OAuth.RefreshToken)
	}
	if cred.OAuth.Expired(time.Now()) {
		t.Fatalf("fresh credential reports Expired() = true")
	}
	if cred.Session != nil {
		t.Fatalf("loopback flow must never set Session, got %+v", cred.Session)
	}
}

func TestLoopbackFlowProviderDeniesAuthorization(t *testing.T) {
	flow, err := loopback.New(loopback.Config{
		AuthURL:     "https://example.invalid/authorize",
		TokenURL:    "https://example.invalid/token",
		ClientID:    "client-123",
		OpenBrowser: fakeBrowserHittingRedirect(t, "access_denied"),
		Timeout:     5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = flow.Login(context.Background())
	if err == nil || !strings.Contains(err.Error(), "access_denied") {
		t.Fatalf("Login() error = %v, want access_denied", err)
	}
}

func TestLoopbackFlowTimesOut(t *testing.T) {
	flow, err := loopback.New(loopback.Config{
		AuthURL:     "https://example.invalid/authorize",
		TokenURL:    "https://example.invalid/token",
		ClientID:    "client-123",
		OpenBrowser: func(string) error { return nil }, // nobody ever hits the callback
		Timeout:     50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = flow.Login(context.Background())
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("Login() error = %v, want timeout", err)
	}
}

func TestLoopbackFlowRejectsStateMismatch(t *testing.T) {
	flow, err := loopback.New(loopback.Config{
		AuthURL:  "https://example.invalid/authorize",
		TokenURL: "https://example.invalid/token",
		ClientID: "client-123",
		OpenBrowser: func(rawURL string) error {
			go func() {
				u, _ := url.Parse(rawURL)
				redirectURI := u.Query().Get("redirect_uri")
				cb, _ := url.Parse(redirectURI)
				q := cb.Query()
				q.Set("state", "wrong-state")
				q.Set("code", "test-auth-code")
				cb.RawQuery = q.Encode()
				resp, err := http.Get(cb.String())
				if err == nil {
					_ = resp.Body.Close()
				}
			}()
			return nil
		},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = flow.Login(context.Background())
	if err == nil || !strings.Contains(err.Error(), "state mismatch") {
		t.Fatalf("Login() error = %v, want state mismatch", err)
	}
}

func TestLoopbackFlowRequiresClientID(t *testing.T) {
	_, err := loopback.New(loopback.Config{
		AuthURL:  "https://example.invalid/authorize",
		TokenURL: "https://example.invalid/token",
	})
	if err == nil {
		t.Fatalf("New() without client_id: want error, got nil")
	}
}

func TestLoopbackFlowContextCancellation(t *testing.T) {
	flow, err := loopback.New(loopback.Config{
		AuthURL:     "https://example.invalid/authorize",
		TokenURL:    "https://example.invalid/token",
		ClientID:    "client-123",
		OpenBrowser: func(string) error { return nil },
		Timeout:     5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = flow.Login(ctx)
	if err == nil {
		t.Fatalf("Login() with cancelled context: want error, got nil")
	}
}
