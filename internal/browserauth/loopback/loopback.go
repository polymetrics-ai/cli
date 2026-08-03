// Package loopback implements OAuth 2.0 authorization-code + PKCE (RFC
// 7636) via the system's default browser and a short-lived HTTP listener on
// 127.0.0.1 — the same pattern `gh auth login`, `gcloud auth login`, and
// `aws sso login` use. It is the chosen mechanism for the official
// connectors' interactive login (report §3.3): zero new browser-automation
// dependency, and the user authenticates in the browser they already trust
// with their own password manager and 2FA — this package never sees, asks
// for, or stores a password.
package loopback

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/browserauth"
)

// Config configures one authorization-code+PKCE login.
type Config struct {
	// AuthURL and TokenURL are the provider's OAuth 2.0 endpoints.
	AuthURL  string
	TokenURL string

	ClientID     string
	ClientSecret string // optional; PKCE public clients may omit it
	Scopes       []string

	// RedirectHost/RedirectPath build the loopback redirect_uri. RedirectHost
	// defaults to "127.0.0.1" (RFC 8252 §7.3 prefers the literal loopback IP
	// over "localhost" to avoid DNS rebinding and OS name-resolution
	// quirks). RedirectPath defaults to "/callback". The port is always
	// chosen by the OS from an ephemeral free port — never fixed — so two
	// concurrent logins, or a stale process holding a port, can't collide.
	RedirectHost string
	RedirectPath string

	// ExtraAuthParams are added to the authorization request verbatim
	// (e.g. a provider-specific "access_type=offline" or "prompt=consent").
	ExtraAuthParams map[string]string

	// OpenBrowser is called with the fully-built authorization URL.
	// Defaults to the OS's default-browser opener. Tests supply their own to
	// simulate the user's browser hitting the redirect without a real one.
	OpenBrowser func(rawURL string) error

	// OnAuthorizationURL, when set, is always invoked with the
	// authorization URL — regardless of whether OpenBrowser succeeds — so a
	// caller (e.g. `pm auth login`) can print "open this URL" as a fallback.
	OnAuthorizationURL func(rawURL string)

	// Timeout bounds how long the loopback listener waits for the user to
	// complete the browser login. Defaults to 5 minutes.
	Timeout time.Duration

	HTTPClient *http.Client
}

// Flow is a browserauth.Flow that runs one authorization-code+PKCE login.
type Flow struct {
	cfg Config
}

// New validates cfg and returns a Flow ready to Login.
func New(cfg Config) (*Flow, error) {
	if strings.TrimSpace(cfg.AuthURL) == "" {
		return nil, errors.New("loopback: auth_url is required")
	}
	if strings.TrimSpace(cfg.TokenURL) == "" {
		return nil, errors.New("loopback: token_url is required")
	}
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, errors.New("loopback: client_id is required")
	}
	if cfg.RedirectHost == "" {
		cfg.RedirectHost = "127.0.0.1"
	}
	if cfg.RedirectPath == "" {
		cfg.RedirectPath = "/callback"
	}
	if !strings.HasPrefix(cfg.RedirectPath, "/") {
		cfg.RedirectPath = "/" + cfg.RedirectPath
	}
	if cfg.OpenBrowser == nil {
		cfg.OpenBrowser = defaultOpenBrowser
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Minute
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Flow{cfg: cfg}, nil
}

func (f *Flow) Name() string { return "loopback_pkce" }

// Login runs the interactive authorization-code+PKCE flow: bind a loopback
// listener, open the browser (or hand the caller the URL), wait for the
// redirect, and exchange the returned code for a token.
func (f *Flow) Login(ctx context.Context) (browserauth.Credential, error) {
	listener, err := net.Listen("tcp", net.JoinHostPort(f.cfg.RedirectHost, "0"))
	if err != nil {
		return browserauth.Credential{}, fmt.Errorf("loopback: bind redirect listener: %w", err)
	}
	defer func() { _ = listener.Close() }()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://%s/%s", net.JoinHostPort(f.cfg.RedirectHost, strconv.Itoa(port)), strings.TrimPrefix(f.cfg.RedirectPath, "/"))

	verifier, challenge, err := generatePKCE()
	if err != nil {
		return browserauth.Credential{}, fmt.Errorf("loopback: generate PKCE challenge: %w", err)
	}
	state, err := randomURLSafe(24)
	if err != nil {
		return browserauth.Credential{}, fmt.Errorf("loopback: generate state: %w", err)
	}

	authURL, err := f.buildAuthURL(redirectURI, state, challenge)
	if err != nil {
		return browserauth.Credential{}, err
	}

	resultCh := make(chan result, 1)
	server := &http.Server{Handler: f.callbackHandler(f.cfg.RedirectPath, state, resultCh)}
	serveErrCh := make(chan error, 1)
	go func() { serveErrCh <- server.Serve(listener) }()
	defer func() { _ = server.Close() }()

	if f.cfg.OnAuthorizationURL != nil {
		f.cfg.OnAuthorizationURL(authURL)
	}
	if err := f.cfg.OpenBrowser(authURL); err != nil && f.cfg.OnAuthorizationURL == nil {
		// No fallback channel was given and the browser couldn't be opened —
		// surface the URL as the error so the caller isn't left guessing.
		return browserauth.Credential{}, fmt.Errorf("loopback: open browser: %w (visit manually: %s)", err, authURL)
	}

	timer := time.NewTimer(f.cfg.Timeout)
	defer timer.Stop()

	var code string
	select {
	case <-ctx.Done():
		return browserauth.Credential{}, ctx.Err()
	case <-timer.C:
		return browserauth.Credential{}, errors.New("loopback: timed out waiting for browser login")
	case err := <-serveErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return browserauth.Credential{}, fmt.Errorf("loopback: redirect listener: %w", err)
		}
		return browserauth.Credential{}, errors.New("loopback: redirect listener stopped before receiving a callback")
	case r := <-resultCh:
		if r.err != nil {
			return browserauth.Credential{}, r.err
		}
		code = r.code
	}

	token, err := f.exchangeCode(ctx, code, verifier, redirectURI)
	if err != nil {
		return browserauth.Credential{}, err
	}
	return browserauth.Credential{OAuth: token}, nil
}

func (f *Flow) buildAuthURL(redirectURI, state, challenge string) (string, error) {
	u, err := url.Parse(f.cfg.AuthURL)
	if err != nil {
		return "", fmt.Errorf("loopback: parse auth_url: %w", err)
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", f.cfg.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	if len(f.cfg.Scopes) > 0 {
		q.Set("scope", strings.Join(f.cfg.Scopes, " "))
	}
	for k, v := range f.cfg.ExtraAuthParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// callbackHandler serves exactly the redirect path once: it verifies state,
// extracts code (or the provider's error), shows the user a plain
// confirmation page, and reports the result on resultCh. Every other path
// (including a second hit of the callback path) gets a 404 — the listener
// is single-use by design.
func (f *Flow) callbackHandler(path, wantState string, resultCh chan<- result) http.Handler {
	mux := http.NewServeMux()
	var once sync.Once
	var handled bool
	var mu sync.Mutex
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		alreadyHandled := handled
		mu.Unlock()
		if alreadyHandled {
			http.NotFound(w, r)
			return
		}

		once.Do(func() {
			mu.Lock()
			handled = true
			mu.Unlock()

			q := r.URL.Query()
			switch {
			case q.Get("error") != "":
				errCode, desc := q.Get("error"), q.Get("error_description")
				writeCallbackPage(w, false)
				resultCh <- result{err: fmt.Errorf("loopback: authorization denied: %s: %s", errCode, desc)}
			case q.Get("state") != wantState:
				writeCallbackPage(w, false)
				resultCh <- result{err: errors.New("loopback: state mismatch on redirect (possible CSRF)")}
			case q.Get("code") == "":
				writeCallbackPage(w, false)
				resultCh <- result{err: errors.New("loopback: redirect had no authorization code")}
			default:
				writeCallbackPage(w, true)
				resultCh <- result{code: q.Get("code")}
			}
		})
	})
	return mux
}

type result struct {
	code string
	err  error
}

func writeCallbackPage(w http.ResponseWriter, ok bool) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if ok {
		_, _ = io.WriteString(w, "<html><body>Signed in. You can close this tab and return to the terminal.</body></html>")
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	_, _ = io.WriteString(w, "<html><body>Sign-in failed. Return to the terminal for details.</body></html>")
}

// tokenResponse is the RFC 6749 §5.1 access token response shape, plus the
// error shape (§5.2) tolerated in the same struct since providers vary in
// which fields they send alongside an error.
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int64  `json:"expires_in"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (f *Flow) exchangeCode(ctx context.Context, code, verifier, redirectURI string) (*browserauth.OAuthCredential, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", f.cfg.ClientID)
	form.Set("code_verifier", verifier)
	if f.cfg.ClientSecret != "" {
		form.Set("client_secret", f.cfg.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("loopback: build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := f.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("loopback: token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("loopback: read token response: %w", err)
	}

	var parsed tokenResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("loopback: decode token response (status %d): %w", resp.StatusCode, err)
	}
	if parsed.Error != "" {
		return nil, fmt.Errorf("loopback: token exchange failed: %s: %s", parsed.Error, parsed.ErrorDescription)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("loopback: token exchange failed: status %d", resp.StatusCode)
	}
	if parsed.AccessToken == "" {
		return nil, errors.New("loopback: token response had no access_token")
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

func generatePKCE() (verifier, challenge string, err error) {
	verifier, err = randomURLSafe(32)
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

func randomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func defaultOpenBrowser(rawURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	return cmd.Start()
}
