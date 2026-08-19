package connsdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// countingRefresher is a minimal Authenticator+AuthRefresher fake: it applies a
// generation-numbered bearer credential and bumps the generation on refresh.
type countingRefresher struct {
	applied   int32
	refreshed int32
	failWith  error
}

func (c *countingRefresher) Apply(_ context.Context, req *http.Request) error {
	atomic.AddInt32(&c.applied, 1)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer gen-%d", atomic.LoadInt32(&c.refreshed)))
	return nil
}

func (c *countingRefresher) RefreshAuth(_ context.Context, _ *http.Request) error {
	if c.failWith != nil {
		atomic.AddInt32(&c.refreshed, 1)
		return c.failWith
	}
	atomic.AddInt32(&c.refreshed, 1)
	return nil
}

// plainAuth implements only Authenticator, never AuthRefresher.
type plainAuth struct{ applied int32 }

func (p *plainAuth) Apply(_ context.Context, req *http.Request) error {
	atomic.AddInt32(&p.applied, 1)
	req.Header.Set("Authorization", "Bearer static")
	return nil
}

// TestRequesterRefreshesAuthOnceOn401AndRetries proves a 401 triggers a
// refresh and one retry, and that the retry carries the refreshed credential
// (issue #3706). Expiry-based renewal alone cannot see an out-of-band
// revocation; the cached token still looks valid.
func TestRequesterRefreshesAuthOnceOn401AndRetries(t *testing.T) {
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Header.Get("Authorization"))
		if len(seen) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	auth := &countingRefresher{}
	r := &Requester{BaseURL: srv.URL, Auth: auth, Sleep: noSleep}

	resp, err := r.Do(context.Background(), http.MethodGet, "/thing", nil, nil)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.Status)
	}
	if got := atomic.LoadInt32(&auth.refreshed); got != 1 {
		t.Fatalf("refreshes = %d, want 1", got)
	}
	if len(seen) != 2 {
		t.Fatalf("upstream attempts = %d, want 2", len(seen))
	}
	if seen[0] != "Bearer gen-0" || seen[1] != "Bearer gen-1" {
		t.Fatalf("credentials seen upstream = %v, want the retry to carry the refreshed one", seen)
	}
}

func TestRequesterReauthKeepsLogicalAttemptsMonotonic(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("RateLimit-Remaining", "99")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	var admissions []RateLimitRequest
	var observations []RateLimitObservation
	r := &Requester{
		BaseURL: srv.URL,
		Auth:    &countingRefresher{},
		Admission: rateLimitAdmissionFunc(func(_ context.Context, request RateLimitRequest) error {
			admissions = append(admissions, request)
			return nil
		}),
		Observer: rateLimitObserverFunc(func(_ context.Context, observation RateLimitObservation) {
			observations = append(observations, observation)
		}),
	}

	if _, err := r.Do(context.Background(), http.MethodGet, "/thing", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got, want := len(admissions), 2; got != want {
		t.Fatalf("admissions = %d, want %d", got, want)
	}
	if admissions[0].Attempt != 1 || admissions[1].Attempt != 2 {
		t.Fatalf("admission attempts = [%d %d], want [1 2]", admissions[0].Attempt, admissions[1].Attempt)
	}
	if got, want := len(observations), 2; got != want {
		t.Fatalf("observations = %d, want %d", got, want)
	}
	if observations[0].Attempt != 1 || observations[1].Attempt != 2 {
		t.Fatalf("observation attempts = [%d %d], want [1 2]", observations[0].Attempt, observations[1].Attempt)
	}
}

func TestRequesterDisableRetriesSuppressesAuthReplay(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	auth := &countingRefresher{}
	r := &Requester{BaseURL: srv.URL, Auth: auth, DisableRetries: true, Sleep: noSleep}

	_, err := r.Do(context.Background(), http.MethodPost, "/thing", nil, map[string]any{"name": "fixture"})
	if err == nil {
		t.Fatal("Do() error = nil, want terminal 401")
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("upstream attempts = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&auth.refreshed); got != 0 {
		t.Fatalf("refreshes = %d, want 0", got)
	}
}

// TestRequesterRefreshesAtMostOncePerRequestOnPersistent401 is the termination
// proof (issue #3706). A provider that keeps returning 401 — revoked grant,
// wrong client, disabled app — must not be hammered.
func TestRequesterRefreshesAtMostOncePerRequestOnPersistent401(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
	}))
	defer srv.Close()

	auth := &countingRefresher{}
	r := &Requester{BaseURL: srv.URL, Auth: auth, Sleep: noSleep}

	_, err := r.Do(context.Background(), http.MethodGet, "/thing", nil, nil)
	if err == nil {
		t.Fatalf("Do() error = nil, want a terminal 401")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusUnauthorized {
		t.Fatalf("Do() error = %v, want a terminal *HTTPError with status 401", err)
	}
	if got := atomic.LoadInt32(&auth.refreshed); got != 1 {
		t.Fatalf("refreshes = %d, want exactly 1 per request", got)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Fatalf("upstream attempts = %d, want 2 (original + one reauth retry)", got)
	}
}

// TestRequesterRefreshFailureDoesNotRetryAgain proves the once-per-request
// guard is set BEFORE the refresh is attempted, so a refresh that itself errors
// cannot produce a second one (issue #3706).
func TestRequesterRefreshFailureDoesNotRetryAgain(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	auth := &countingRefresher{failWith: errors.New("token endpoint down")}
	r := &Requester{BaseURL: srv.URL, Auth: auth, Sleep: noSleep}

	_, err := r.Do(context.Background(), http.MethodGet, "/thing", nil, nil)
	if err == nil {
		t.Fatalf("Do() error = nil, want an error")
	}
	if got := atomic.LoadInt32(&auth.refreshed); got != 1 {
		t.Fatalf("refreshes = %d, want exactly 1 even though it failed", got)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("upstream attempts = %d, want 1 (a failed refresh must not retry)", got)
	}
}

// TestRequesterReauthRetriesEvenWithNoRetryBudget proves the reauth retry does
// not come out of the transient-failure budget: a MaxRetries:0 requester must
// still get its one post-refresh attempt.
func TestRequesterReauthRetriesEvenWithNoRetryBudget(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	auth := &countingRefresher{}
	r := &Requester{BaseURL: srv.URL, Auth: auth, MaxRetries: 1, Sleep: noSleep}

	if _, err := r.Do(context.Background(), http.MethodGet, "/thing", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Fatalf("upstream attempts = %d, want 2", got)
	}
}

// A 401-triggered credential refresh is a retry even though it does not spend
// MaxRetries. Non-idempotent rest_write dispatches set DisableRetries, so this
// regression guard proves an expired/revoked credential cannot cause a second
// mutation attempt through the auth-refresh path.
func TestRequesterDisableRetriesSuppressesAuthRefresh(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
	}))
	defer srv.Close()

	auth := &countingRefresher{}
	r := &Requester{BaseURL: srv.URL, Auth: auth, DisableRetries: true, Sleep: noSleep}

	_, err := r.Do(context.Background(), http.MethodPost, "/mutate", nil, map[string]string{"name": "widget"})
	if err == nil {
		t.Fatal("Do() error = nil, want a terminal 401")
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusUnauthorized {
		t.Fatalf("Do() error = %v, want a terminal *HTTPError with status 401", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("upstream attempts = %d, want exactly 1", got)
	}
	if got := atomic.LoadInt32(&auth.refreshed); got != 0 {
		t.Fatalf("refreshes = %d, want 0 when retries are disabled", got)
	}
}

// TestRequesterDoesNotRefreshWhenAuthenticatorIsNotARefresher is the
// additive/opt-in proof: every existing auth mode — none, bearer, basic, the
// two api-key modes, oauth2_client_credentials, custom — behaves byte-for-byte
// as it does today on a 401.
func TestRequesterDoesNotRefreshWhenAuthenticatorIsNotARefresher(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	auth := &plainAuth{}
	r := &Requester{BaseURL: srv.URL, Auth: auth, Sleep: noSleep}

	_, err := r.Do(context.Background(), http.MethodGet, "/thing", nil, nil)
	if err == nil {
		t.Fatalf("Do() error = nil, want a terminal 401")
	}
	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("upstream attempts = %d, want 1 (no reauth for a plain Authenticator)", got)
	}
	if got := atomic.LoadInt32(&auth.applied); got != 1 {
		t.Fatalf("auth applied %d times, want 1", got)
	}
}

// TestOAuth2RefreshTokenRefreshAuthReExchanges wires the real authenticator
// into a real Requester: a revoked access token 401s, the refresher exchanges
// once, and the retry succeeds with the new credential.
func TestOAuth2RefreshTokenRefreshAuthReExchanges(t *testing.T) {
	ts := newTokenServer(t, func(call int, _ map[string]string, w http.ResponseWriter) {
		fmt.Fprintf(w, `{"access_token":"AT-%d","token_type":"Bearer","expires_in":3600}`, call)
	})

	var seen []string
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		seen = append(seen, got)
		// AT-1 was revoked out of band; AT-2 works.
		if got == "Bearer AT-1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer api.Close()

	auth := &OAuth2RefreshToken{TokenURL: ts.URL, RefreshToken: "rt"}
	r := &Requester{BaseURL: api.URL, Auth: auth, Sleep: noSleep}

	if _, err := r.Do(context.Background(), http.MethodGet, "/thing", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(seen) != 2 || seen[0] != "Bearer AT-1" || seen[1] != "Bearer AT-2" {
		t.Fatalf("credentials seen upstream = %v, want [Bearer AT-1 Bearer AT-2]", seen)
	}
	if got := ts.callCount(); got != 2 {
		t.Fatalf("token endpoint calls = %d, want 2 (initial + one reauth)", got)
	}
}

// TestOAuth2RefreshTokenConcurrentRefreshAuthCollapses proves concurrent 401
// refreshes for the SAME stale credential collapse to one exchange (issue
// #3707): a caller whose credential has already been replaced returns without
// exchanging, and its retry picks up the fresher token.
func TestOAuth2RefreshTokenConcurrentRefreshAuthCollapses(t *testing.T) {
	ts := newTokenServer(t, func(call int, _ map[string]string, w http.ResponseWriter) {
		fmt.Fprintf(w, `{"access_token":"AT-%d","token_type":"Bearer","expires_in":3600}`, call)
	})

	auth := &OAuth2RefreshToken{TokenURL: ts.URL, RefreshToken: "rt"}

	// Seed the cache so every goroutine below starts from the same token.
	seed := newReq(t)
	if err := auth.Apply(context.Background(), seed); err != nil {
		t.Fatalf("seed Apply: %v", err)
	}
	if got := seed.Header.Get("Authorization"); got != "Bearer AT-1" {
		t.Fatalf("seed Authorization = %q", got)
	}

	const callers = 8
	done := make(chan error, callers)
	for i := 0; i < callers; i++ {
		go func() {
			stale, err := http.NewRequest(http.MethodGet, "https://example.com/x", nil)
			if err != nil {
				done <- err
				return
			}
			stale.Header.Set("Authorization", "Bearer AT-1")
			done <- auth.RefreshAuth(context.Background(), stale)
		}()
	}
	for i := 0; i < callers; i++ {
		if err := <-done; err != nil {
			t.Fatalf("RefreshAuth: %v", err)
		}
	}

	if got := ts.callCount(); got != 2 {
		t.Fatalf("token endpoint calls = %d, want 2 (seed + exactly one collapsed refresh)", got)
	}
	req := newReq(t)
	if err := auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply after refresh: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer AT-2" {
		t.Fatalf("Authorization after collapsed refresh = %q, want Bearer AT-2", got)
	}
}
