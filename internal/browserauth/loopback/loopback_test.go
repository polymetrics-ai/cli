package loopback_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/browserauth/loopback"
)

// browserResult is what the simulated browser actually got back from the
// loopback listener. Tests assert on it rather than having the goroutine
// call t.Errorf directly: the goroutine outlives Login, so logging from it
// races with test completion and panics the whole test binary
// ("Log in goroutine after Test... has completed") instead of failing one
// test cleanly.
type browserResult struct {
	body string
	err  error
}

// fakeBrowser simulates the user completing the login in a real browser: it
// GETs the authorization URL against a fake provider auth endpoint, which
// redirects back to the loopback listener exactly like a browser would —
// no real browser or network access involved. It reads the callback
// response to completion, because the page the listener serves ("Signed
// in. You can close this tab") is the only thing the user ever sees; a
// listener that reports success to the CLI but drops the connection before
// that page lands leaves them staring at a browser error.
func fakeBrowserHittingRedirect(t *testing.T, wantError string) (func(string) error, <-chan browserResult) {
	t.Helper()
	results := make(chan browserResult, 1)
	open := func(rawURL string) error {
		go func() {
			u, err := url.Parse(rawURL)
			if err != nil {
				results <- browserResult{err: fmt.Errorf("parse authorization URL: %w", err)}
				return
			}
			redirectURI := u.Query().Get("redirect_uri")
			state := u.Query().Get("state")
			cb, err := url.Parse(redirectURI)
			if err != nil {
				results <- browserResult{err: fmt.Errorf("parse redirect_uri: %w", err)}
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
				results <- browserResult{err: fmt.Errorf("simulate browser redirect: %w", err)}
				return
			}
			defer func() { _ = resp.Body.Close() }()
			body, err := io.ReadAll(resp.Body)
			results <- browserResult{body: string(body), err: err}
		}()
		return nil
	}
	return open, results
}

// awaitCallbackPage returns what the simulated browser received, failing
// rather than hanging if the redirect never completed.
func awaitCallbackPage(t *testing.T, results <-chan browserResult) browserResult {
	t.Helper()
	select {
	case got := <-results:
		return got
	case <-time.After(10 * time.Second):
		t.Fatalf("simulated browser never finished reading the callback page")
		return browserResult{}
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

	openBrowser, browserResults := fakeBrowserHittingRedirect(t, "")
	flow, err := loopback.New(loopback.Config{
		AuthURL:     authServer.URL + "/authorize",
		TokenURL:    tokenServer.URL + "/token",
		ClientID:    "client-123",
		Scopes:      []string{"read", "write"},
		OpenBrowser: openBrowser,
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

	// The user's browser must actually receive the confirmation page. Login
	// returns as soon as the handler hands it the code, so tearing the
	// listener down at that moment would leave the tab on a connection
	// error even though the CLI succeeded.
	page := awaitCallbackPage(t, browserResults)
	if page.err != nil {
		t.Fatalf("browser never received the callback page: %v", page.err)
	}
	if !strings.Contains(page.body, "Signed in") {
		t.Fatalf("callback page = %q, want the signed-in confirmation", page.body)
	}
}

func TestLoopbackFlowProviderDeniesAuthorization(t *testing.T) {
	openBrowser, browserResults := fakeBrowserHittingRedirect(t, "access_denied")
	flow, err := loopback.New(loopback.Config{
		AuthURL:     "https://example.invalid/authorize",
		TokenURL:    "https://example.invalid/token",
		ClientID:    "client-123",
		OpenBrowser: openBrowser,
		Timeout:     5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = flow.Login(context.Background())
	if err == nil || !strings.Contains(err.Error(), "access_denied") {
		t.Fatalf("Login() error = %v, want access_denied", err)
	}

	// The denial path returns from Login immediately — no token exchange to
	// hold the listener open — so it is where a Close()d listener most
	// reliably beats the response out the door. The user must still get the
	// failure page, not a bare connection error.
	page := awaitCallbackPage(t, browserResults)
	if page.err != nil {
		t.Fatalf("browser never received the callback page: %v", page.err)
	}
	if !strings.Contains(page.body, "Sign-in failed") {
		t.Fatalf("callback page = %q, want the sign-in-failed page", page.body)
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

func TestLoopbackFlowIgnoresStateMismatch(t *testing.T) {
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
		Timeout: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = flow.Login(context.Background())
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("Login() error = %v, want timeout after ignored state mismatch", err)
	}
}

func TestLoopbackFlowIgnoresMismatchedErrorCallback(t *testing.T) {
	tokenServer := fakeTokenServer(t, "test-auth-code")
	defer tokenServer.Close()

	flow, err := loopback.New(loopback.Config{
		AuthURL:  "https://example.invalid/authorize",
		TokenURL: tokenServer.URL + "/token",
		ClientID: "client-123",
		OpenBrowser: func(rawURL string) error {
			go func() {
				u, err := url.Parse(rawURL)
				if err != nil {
					return
				}
				redirectURI := u.Query().Get("redirect_uri")
				state := u.Query().Get("state")
				callback, err := url.Parse(redirectURI)
				if err != nil {
					return
				}

				invalid := callback.Query()
				invalid.Set("state", "wrong-state")
				invalid.Set("error", "access_denied")
				callback.RawQuery = invalid.Encode()
				if resp, err := http.Get(callback.String()); err == nil {
					_ = resp.Body.Close()
				}

				valid := callback.Query()
				valid.Set("state", state)
				valid.Del("error")
				valid.Set("code", "test-auth-code")
				callback.RawQuery = valid.Encode()
				if resp, err := http.Get(callback.String()); err == nil {
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

	cred, err := flow.Login(context.Background())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if cred.OAuth == nil || cred.OAuth.AccessToken != "test-access-token" {
		t.Fatalf("Login() credential = %+v", cred)
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

func TestNewRejectsNonLoopbackRedirectHost(t *testing.T) {
	for _, host := range []string{"0.0.0.0", "192.0.2.25", "example.invalid", "localhost"} {
		t.Run(host, func(t *testing.T) {
			_, err := loopback.New(loopback.Config{
				AuthURL:      "https://example.invalid/authorize",
				TokenURL:     "https://example.invalid/token",
				ClientID:     "client-123",
				RedirectHost: host,
			})
			if err == nil {
				t.Fatalf("New() with RedirectHost %q: want loopback-host error, got nil", host)
			}
		})
	}
}

func TestNewAllowsLiteralLoopbackRedirectHost(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1"} {
		t.Run(host, func(t *testing.T) {
			if _, err := loopback.New(loopback.Config{
				AuthURL:      "https://example.invalid/authorize",
				TokenURL:     "https://example.invalid/token",
				ClientID:     "client-123",
				RedirectHost: host,
			}); err != nil {
				t.Fatalf("New() with RedirectHost %q: %v", host, err)
			}
		})
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
