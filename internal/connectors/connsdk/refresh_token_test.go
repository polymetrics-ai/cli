package connsdk

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/credential"
)

// tokenServer is a scriptable OAuth2 token endpoint. Each request increments
// calls and is answered by respond, which receives the 1-based call number and
// the parsed form so a test can assert on the grant it received.
type tokenServer struct {
	*httptest.Server

	calls   int32
	mu      sync.Mutex
	grants  []string
	respond func(call int, form map[string]string, w http.ResponseWriter)
}

func newTokenServer(t *testing.T, respond func(call int, form map[string]string, w http.ResponseWriter)) *tokenServer {
	t.Helper()
	ts := &tokenServer{respond: respond}
	ts.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := int(atomic.AddInt32(&ts.calls, 1))
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		form := map[string]string{}
		for k := range r.PostForm {
			form[k] = r.PostForm.Get(k)
		}
		ts.mu.Lock()
		ts.grants = append(ts.grants, form["refresh_token"])
		ts.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		ts.respond(call, form, w)
	}))
	t.Cleanup(ts.Close)
	return ts
}

func (ts *tokenServer) callCount() int { return int(atomic.LoadInt32(&ts.calls)) }

func (ts *tokenServer) presentedGrants() []string {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return append([]string(nil), ts.grants...)
}

// TestOAuth2RefreshTokenFirstExchangeSetsBearer is the first-exchange proof
// (issue #3703): a refresh-token grant is POSTed to the declared token URL and
// the resulting access token reaches the request as a bearer credential.
func TestOAuth2RefreshTokenFirstExchangeSetsBearer(t *testing.T) {
	var gotForm map[string]string
	ts := newTokenServer(t, func(_ int, form map[string]string, w http.ResponseWriter) {
		gotForm = form
		_, _ = w.Write([]byte(`{"access_token":"AT-1","token_type":"Bearer","expires_in":3600}`))
	})

	auth := &OAuth2RefreshToken{
		TokenURL:     ts.URL,
		ClientID:     "cid",
		ClientSecret: "csecret",
		RefreshToken: "rt-original",
	}

	req := newReq(t)
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer AT-1" {
		t.Fatalf("Authorization = %q, want %q", got, "Bearer AT-1")
	}
	if gotForm["grant_type"] != "refresh_token" {
		t.Fatalf("grant_type = %q, want refresh_token", gotForm["grant_type"])
	}
	if gotForm["refresh_token"] != "rt-original" {
		t.Fatalf("refresh_token form param = %q, want rt-original", gotForm["refresh_token"])
	}
	if gotForm["client_id"] != "cid" || gotForm["client_secret"] != "csecret" {
		t.Fatalf("client credentials not sent: %+v", gotForm)
	}
}

func TestOAuth2RefreshTokenBasicClientAuthentication(t *testing.T) {
	var (
		form      map[string]string
		clientID  string
		clientKey string
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var ok bool
		clientID, clientKey, ok = request.BasicAuth()
		if !ok {
			t.Fatal("token request omitted HTTP Basic client authentication")
		}
		if err := request.ParseForm(); err != nil {
			t.Fatalf("parse token form: %v", err)
		}
		form = map[string]string{}
		for key := range request.PostForm {
			form[key] = request.PostForm.Get(key)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"AT-basic","expires_in":3600}`))
	}))
	t.Cleanup(server.Close)

	auth := &OAuth2RefreshToken{
		TokenURL:             server.URL,
		ClientID:             "client-id",
		ClientSecret:         "client-secret",
		ClientAuthentication: "basic",
		RefreshToken:         "refresh-token",
	}
	request := newReq(t)
	if err := auth.Apply(context.Background(), request); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if clientID != "client-id" || clientKey != "client-secret" {
		t.Fatalf("BasicAuth() = (%q, %q), want declared client credentials", clientID, clientKey)
	}
	if form["grant_type"] != "refresh_token" || form["refresh_token"] != "refresh-token" {
		t.Fatalf("token form = %#v, want declared refresh-token grant", form)
	}
	if form["client_id"] != "" || form["client_secret"] != "" {
		t.Fatalf("token form included Basic client credentials: %#v", form)
	}
	if got := request.Header.Get("Authorization"); got != "Bearer AT-basic" {
		t.Fatalf("Authorization = %q, want exchanged bearer token", got)
	}
}

func TestOAuth2RefreshTokenRejectsUnknownClientAuthenticationBeforeExchange(t *testing.T) {
	server := newTokenServer(t, func(_ int, _ map[string]string, w http.ResponseWriter) {
		t.Fatal("unsupported client authentication reached token endpoint")
		w.WriteHeader(http.StatusInternalServerError)
	})
	auth := &OAuth2RefreshToken{
		TokenURL:             server.URL,
		ClientAuthentication: "unsupported",
		RefreshToken:         "refresh-token",
	}
	if err := auth.Apply(context.Background(), newReq(t)); err == nil {
		t.Fatal("Apply() accepted unsupported client authentication")
	}
	if got := server.callCount(); got != 0 {
		t.Fatalf("token exchange calls = %d, want zero", got)
	}
}

func TestOAuth2RefreshTokenPreservesFormCredentialBytes(t *testing.T) {
	refreshToken := "refresh-token-canary\n"
	clientID := "client-id-canary\n"
	clientSecret := strings.Repeat("client-secret-canary-", 1024) + "\n"
	var gotForm map[string]string
	ts := newTokenServer(t, func(_ int, form map[string]string, w http.ResponseWriter) {
		gotForm = form
		_, _ = w.Write([]byte(`{"access_token":"access-token","expires_in":3600}`))
	})

	auth := &OAuth2RefreshToken{
		TokenURL:     ts.URL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
	}
	if err := auth.Apply(context.Background(), newReq(t)); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	assertCredentialBytes(t, gotForm["refresh_token"], refreshToken)
	assertCredentialBytes(t, gotForm["client_id"], clientID)
	assertCredentialBytes(t, gotForm["client_secret"], clientSecret)
}

func TestOAuth2RefreshTokenRejectsEmptyRequiredTokenBeforeExchange(t *testing.T) {
	for _, tt := range []struct {
		name  string
		token string
	}{
		{name: "empty"},
		{name: "LF-only", token: "\n"},
		{name: "CRLF-only", token: "\r\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ts := newTokenServer(t, func(_ int, _ map[string]string, w http.ResponseWriter) {
				w.WriteHeader(http.StatusInternalServerError)
			})
			auth := &OAuth2RefreshToken{TokenURL: ts.URL, RefreshToken: tt.token}
			req := newReq(t)
			err := auth.Apply(context.Background(), req)
			if err == nil {
				t.Fatal("Apply() accepted an empty refresh token")
			}
			var empty *credential.EmptySecretError
			if !errors.As(err, &empty) {
				t.Fatalf("Apply() error is not typed empty-secret classification: %T", err)
			}
			if got := atomic.LoadInt32(&ts.calls); got != 0 {
				t.Fatalf("token endpoint calls = %d, want 0", got)
			}
			if got := req.Header.Get("Authorization"); got != "" {
				t.Fatal("Authorization header emitted for empty refresh token")
			}
		})
	}
}

func TestOAuth2RefreshTokenRejectsHeaderControlAccessToken(t *testing.T) {
	ts := newTokenServer(t, func(_ int, _ map[string]string, w http.ResponseWriter) {
		_, _ = w.Write([]byte(`{"access_token":"access-token\ncanary","expires_in":3600}`))
	})
	auth := &OAuth2RefreshToken{TokenURL: ts.URL, RefreshToken: "refresh-token-canary"}
	req := newReq(t)
	err := auth.Apply(context.Background(), req)
	var invalid *credential.InvalidSecretValueError
	if !errors.As(err, &invalid) {
		t.Fatalf("Apply() error is not typed invalid-secret classification: %T", err)
	}
	if got := req.Header.Get("Authorization"); got != "" {
		t.Fatal("Authorization header emitted for prohibited access-token bytes")
	}
}

// TestOAuth2RefreshTokenReusesAccessTokenBeforeExpiry proves the token is
// cached rather than re-exchanged per request (issue #3704).
func TestOAuth2RefreshTokenReusesAccessTokenBeforeExpiry(t *testing.T) {
	ts := newTokenServer(t, func(call int, _ map[string]string, w http.ResponseWriter) {
		fmt.Fprintf(w, `{"access_token":"AT-%d","token_type":"Bearer","expires_in":3600}`, call)
	})

	now := time.Unix(1_700_000_000, 0)
	auth := &OAuth2RefreshToken{
		TokenURL:     ts.URL,
		RefreshToken: "rt",
		Now:          func() time.Time { return now },
	}

	for i := 0; i < 5; i++ {
		req := newReq(t)
		if err := auth.Apply(context.Background(), req); err != nil {
			t.Fatalf("Apply #%d: %v", i, err)
		}
		if got := req.Header.Get("Authorization"); got != "Bearer AT-1" {
			t.Fatalf("Apply #%d Authorization = %q, want Bearer AT-1", i, got)
		}
	}
	if got := ts.callCount(); got != 1 {
		t.Fatalf("token endpoint calls = %d, want 1 (cached until near expiry)", got)
	}
}

// TestOAuth2RefreshTokenRefreshesAtExpiry proves the cached token is dropped
// once the clock passes the renewal point (issue #3704).
func TestOAuth2RefreshTokenRefreshesAtExpiry(t *testing.T) {
	ts := newTokenServer(t, func(call int, _ map[string]string, w http.ResponseWriter) {
		fmt.Fprintf(w, `{"access_token":"AT-%d","token_type":"Bearer","expires_in":3600}`, call)
	})

	now := time.Unix(1_700_000_000, 0)
	auth := &OAuth2RefreshToken{
		TokenURL:     ts.URL,
		RefreshToken: "rt",
		Now:          func() time.Time { return now },
	}

	req := newReq(t)
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("first Apply: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer AT-1" {
		t.Fatalf("first Authorization = %q", got)
	}

	// Past the 3600s lifetime.
	now = now.Add(3601 * time.Second)
	req2 := newReq(t)
	if err := auth.Apply(context.Background(), req2); err != nil {
		t.Fatalf("second Apply: %v", err)
	}
	if got := req2.Header.Get("Authorization"); got != "Bearer AT-2" {
		t.Fatalf("second Authorization = %q, want Bearer AT-2 (re-exchanged at expiry)", got)
	}
	if got := ts.callCount(); got != 2 {
		t.Fatalf("token endpoint calls = %d, want 2", got)
	}
}

// TestOAuth2RefreshTokenRenewsWithinSafetyMargin proves renewal happens
// slightly BEFORE the provider's stated deadline, so a token cannot expire in
// flight (issue #3704).
func TestOAuth2RefreshTokenRenewsWithinSafetyMargin(t *testing.T) {
	ts := newTokenServer(t, func(call int, _ map[string]string, w http.ResponseWriter) {
		fmt.Fprintf(w, `{"access_token":"AT-%d","token_type":"Bearer","expires_in":3600}`, call)
	})

	now := time.Unix(1_700_000_000, 0)
	auth := &OAuth2RefreshToken{
		TokenURL:     ts.URL,
		RefreshToken: "rt",
		Now:          func() time.Time { return now },
	}

	req := newReq(t)
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("first Apply: %v", err)
	}

	// Inside the token's stated lifetime, but inside the 60s safety margin.
	now = now.Add(3570 * time.Second)
	req2 := newReq(t)
	if err := auth.Apply(context.Background(), req2); err != nil {
		t.Fatalf("second Apply: %v", err)
	}
	if got := ts.callCount(); got != 2 {
		t.Fatalf("token endpoint calls = %d, want 2 (renewed inside the safety margin)", got)
	}
}

// TestOAuth2RefreshTokenMissingExpiresInUsesShortFallbackTTL proves a token
// response with no usable expires_in is treated as SHORT-lived, never as
// "never expires" (issue #3704). OAuth2ClientCredentials assumes 3600s in this
// case; for a user-context token that guess is unsafe.
func TestOAuth2RefreshTokenMissingExpiresInUsesShortFallbackTTL(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
	}{
		{"absent", `{"access_token":"AT-%d","token_type":"Bearer"}`},
		{"zero", `{"access_token":"AT-%d","token_type":"Bearer","expires_in":0}`},
		{"negative", `{"access_token":"AT-%d","token_type":"Bearer","expires_in":-1}`},
		{"unparseable", `{"access_token":"AT-%d","token_type":"Bearer","expires_in":"soon"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := newTokenServer(t, func(call int, _ map[string]string, w http.ResponseWriter) {
				fmt.Fprintf(w, tc.body, call)
			})
			now := time.Unix(1_700_000_000, 0)
			auth := &OAuth2RefreshToken{
				TokenURL:     ts.URL,
				RefreshToken: "rt",
				Now:          func() time.Time { return now },
			}

			req := newReq(t)
			if err := auth.Apply(context.Background(), req); err != nil {
				t.Fatalf("first Apply: %v", err)
			}
			// Well inside an hour, but past a conservative short life.
			now = now.Add(10 * time.Minute)
			req2 := newReq(t)
			if err := auth.Apply(context.Background(), req2); err != nil {
				t.Fatalf("second Apply: %v", err)
			}
			if got := ts.callCount(); got != 2 {
				t.Fatalf("token endpoint calls = %d, want 2 (missing expires_in must not mean 'never expires')", got)
			}
		})
	}
}

// TestOAuth2RefreshTokenShortLifetimeStillCaches proves the safety margin is
// clamped to half the token lifetime. Without the clamp, a provider handing out
// a lifetime shorter than the margin would re-exchange on EVERY request —
// exactly the token-endpoint hammering this mode exists to avoid.
func TestOAuth2RefreshTokenShortLifetimeStillCaches(t *testing.T) {
	ts := newTokenServer(t, func(call int, _ map[string]string, w http.ResponseWriter) {
		fmt.Fprintf(w, `{"access_token":"AT-%d","token_type":"Bearer","expires_in":10}`, call)
	})

	now := time.Unix(1_700_000_000, 0)
	auth := &OAuth2RefreshToken{
		TokenURL:     ts.URL,
		RefreshToken: "rt",
		Now:          func() time.Time { return now },
	}

	for i := 0; i < 4; i++ {
		req := newReq(t)
		if err := auth.Apply(context.Background(), req); err != nil {
			t.Fatalf("Apply #%d: %v", i, err)
		}
	}
	if got := ts.callCount(); got != 1 {
		t.Fatalf("token endpoint calls = %d, want 1 (a lifetime shorter than the margin must still cache)", got)
	}

	// Half of 10s is the clamped margin, so renewal is due at +5s.
	now = now.Add(6 * time.Second)
	req := newReq(t)
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply after clamped margin: %v", err)
	}
	if got := ts.callCount(); got != 2 {
		t.Fatalf("token endpoint calls = %d, want 2 after the clamped renewal point", got)
	}
}

// TestOAuth2RefreshTokenRotatedTokenUsedOnNextExchange proves a provider that
// rotates its refresh token is honoured: the SECOND exchange presents the
// rotated value, not the original (issue #3705).
func TestOAuth2RefreshTokenRotatedTokenUsedOnNextExchange(t *testing.T) {
	ts := newTokenServer(t, func(call int, _ map[string]string, w http.ResponseWriter) {
		fmt.Fprintf(w, `{"access_token":"AT-%d","refresh_token":"rt-rotated-%d","expires_in":3600}`, call, call)
	})

	now := time.Unix(1_700_000_000, 0)
	auth := &OAuth2RefreshToken{
		TokenURL:     ts.URL,
		RefreshToken: "rt-original",
		Now:          func() time.Time { return now },
	}

	if err := auth.Apply(context.Background(), newReq(t)); err != nil {
		t.Fatalf("first Apply: %v", err)
	}
	now = now.Add(2 * time.Hour)
	if err := auth.Apply(context.Background(), newReq(t)); err != nil {
		t.Fatalf("second Apply: %v", err)
	}

	grants := ts.presentedGrants()
	want := []string{"rt-original", "rt-rotated-1"}
	if len(grants) != len(want) {
		t.Fatalf("presented grants = %v, want %v", grants, want)
	}
	for i := range want {
		if grants[i] != want[i] {
			t.Fatalf("exchange %d presented refresh_token %q, want %q", i+1, grants[i], want[i])
		}
	}
}

// TestOAuth2RefreshTokenRotationCallbackInvoked proves the rotated value is
// handed to the caller for persistence, and only when it actually changed
// (issue #3705).
func TestOAuth2RefreshTokenRotationCallbackInvoked(t *testing.T) {
	ts := newTokenServer(t, func(call int, _ map[string]string, w http.ResponseWriter) {
		if call == 1 {
			fmt.Fprintf(w, `{"access_token":"AT-1","refresh_token":"rt-rotated","expires_in":3600}`)
			return
		}
		// Second exchange returns the SAME refresh token: no rotation.
		fmt.Fprintf(w, `{"access_token":"AT-2","refresh_token":"rt-rotated","expires_in":3600}`)
	})

	var mu sync.Mutex
	var rotations []string
	now := time.Unix(1_700_000_000, 0)
	auth := &OAuth2RefreshToken{
		TokenURL:     ts.URL,
		RefreshToken: "rt-original",
		Now:          func() time.Time { return now },
		OnRefreshTokenRotated: func(_ context.Context, token string) error {
			mu.Lock()
			defer mu.Unlock()
			rotations = append(rotations, token)
			return nil
		},
	}

	if err := auth.Apply(context.Background(), newReq(t)); err != nil {
		t.Fatalf("first Apply: %v", err)
	}
	now = now.Add(2 * time.Hour)
	if err := auth.Apply(context.Background(), newReq(t)); err != nil {
		t.Fatalf("second Apply: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(rotations) != 1 || rotations[0] != "rt-rotated" {
		t.Fatalf("rotations = %v, want exactly one %q (unchanged token must not re-persist)", rotations, "rt-rotated")
	}
}

// TestOAuth2RefreshTokenRotationPersistFailureSurfaces proves a store failure
// is reported rather than silently swallowed: a rotated token that was not
// persisted breaks the NEXT run, so the run that caused it must fail loudly.
func TestOAuth2RefreshTokenRotationPersistFailureSurfaces(t *testing.T) {
	ts := newTokenServer(t, func(_ int, _ map[string]string, w http.ResponseWriter) {
		fmt.Fprint(w, `{"access_token":"AT-1","refresh_token":"rt-rotated","expires_in":3600}`)
	})

	auth := &OAuth2RefreshToken{
		TokenURL:     ts.URL,
		RefreshToken: "rt-original",
		OnRefreshTokenRotated: func(_ context.Context, _ string) error {
			return fmt.Errorf("vault write failed")
		},
	}

	err := auth.Apply(context.Background(), newReq(t))
	if err == nil {
		t.Fatalf("Apply() error = nil, want a surfaced persistence failure")
	}
	if !strings.Contains(err.Error(), "vault write failed") {
		t.Fatalf("Apply() error = %v, want it to name the persistence failure", err)
	}
	if strings.Contains(err.Error(), "rt-rotated") || strings.Contains(err.Error(), "rt-original") {
		t.Fatalf("persistence failure leaked a refresh token: %v", err)
	}
}

// TestOAuth2RefreshTokenConcurrentApplyCausesExactlyOneExchange is the
// concurrency proof (issue #3707). Several streams share one authenticator;
// they must not all exchange. Run under -race.
func TestOAuth2RefreshTokenConcurrentApplyCausesExactlyOneExchange(t *testing.T) {
	release := make(chan struct{})
	ts := newTokenServer(t, func(_ int, _ map[string]string, w http.ResponseWriter) {
		// Hold the first exchange open long enough that every other
		// goroutine is definitely inside Apply before it completes.
		<-release
		fmt.Fprint(w, `{"access_token":"AT-shared","token_type":"Bearer","expires_in":3600}`)
	})

	auth := &OAuth2RefreshToken{TokenURL: ts.URL, RefreshToken: "rt"}

	const callers = 16
	var wg sync.WaitGroup
	started := make(chan struct{}, callers)
	got := make([]string, callers)
	errs := make([]error, callers)
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			started <- struct{}{}
			req := newReq(t)
			if err := auth.Apply(context.Background(), req); err != nil {
				errs[i] = err
				return
			}
			got[i] = req.Header.Get("Authorization")
		}(i)
	}
	for i := 0; i < callers; i++ {
		<-started
	}
	close(release)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("caller %d: Apply: %v", i, err)
		}
	}
	for i, h := range got {
		if h != "Bearer AT-shared" {
			t.Fatalf("caller %d Authorization = %q, want the shared token", i, h)
		}
	}
	if calls := ts.callCount(); calls != 1 {
		t.Fatalf("token endpoint calls = %d, want exactly 1 for %d concurrent callers", calls, callers)
	}
}

// TestOAuth2RefreshTokenErrorsNeverLeakCredentials is the redaction proof for
// the parent issue's standing constraint: the refresh token, client secret and
// access token are all secrets and must reach no error string, including from a
// token endpoint that deliberately echoes them back.
func TestOAuth2RefreshTokenErrorsNeverLeakCredentials(t *testing.T) {
	const (
		refreshToken = "rt-SUPERSECRET-refresh"
		clientSecret = "cs-SUPERSECRET-client"
		accessToken  = "at-SUPERSECRET-access"
	)
	secrets := []string{refreshToken, clientSecret, accessToken}

	// A closed listener gives a deterministic transport failure whose
	// url.Error text carries the request URL.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	deadURL := "http://" + ln.Addr().String() + "/token?client_secret=" + clientSecret
	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	cases := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		auth    func(tokenURL string) *OAuth2RefreshToken
	}{
		{
			name: "provider 4xx echoing every credential back in the body",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, `{"error":"invalid_grant","error_description":"refresh_token=%s client_secret=%s access_token=%s"}`,
					refreshToken, clientSecret, accessToken)
			},
		},
		{
			name: "malformed token response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintf(w, `{"access_token": %s`, refreshToken)
			},
		},
		{
			name: "token response missing access_token",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintf(w, `{"token_type":"Bearer","refresh_token":%q}`, refreshToken)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(tc.handler))
			defer srv.Close()
			auth := &OAuth2RefreshToken{
				TokenURL:     srv.URL,
				ClientID:     "cid",
				ClientSecret: clientSecret,
				RefreshToken: refreshToken,
			}
			err := auth.Apply(context.Background(), newReq(t))
			if err == nil {
				t.Fatalf("Apply() error = nil, want an error")
			}
			assertNoSecrets(t, err.Error(), secrets)
		})
	}

	t.Run("missing token url", func(t *testing.T) {
		auth := &OAuth2RefreshToken{ClientSecret: clientSecret, RefreshToken: refreshToken}
		err := auth.Apply(context.Background(), newReq(t))
		if err == nil {
			t.Fatalf("Apply() error = nil, want an error for a missing token URL")
		}
		assertNoSecrets(t, err.Error(), secrets)
	})

	t.Run("missing refresh token", func(t *testing.T) {
		auth := &OAuth2RefreshToken{TokenURL: "https://example.invalid/token", ClientSecret: clientSecret}
		err := auth.Apply(context.Background(), newReq(t))
		if err == nil {
			t.Fatalf("Apply() error = nil, want an error for a missing refresh token")
		}
		assertNoSecrets(t, err.Error(), secrets)
	})

	t.Run("transport failure against a dead endpoint", func(t *testing.T) {
		auth := &OAuth2RefreshToken{
			TokenURL:     deadURL,
			ClientSecret: clientSecret,
			RefreshToken: refreshToken,
			Client:       &http.Client{Timeout: 2 * time.Second},
		}
		err := auth.Apply(context.Background(), newReq(t))
		if err == nil {
			t.Fatalf("Apply() error = nil, want a transport error")
		}
		assertNoSecrets(t, err.Error(), secrets)
	})
}

func assertNoSecrets(t *testing.T, text string, secrets []string) {
	t.Helper()
	for _, s := range secrets {
		if strings.Contains(text, s) {
			t.Fatalf("error text leaked a credential (%q): %s", s, text)
		}
	}
}
