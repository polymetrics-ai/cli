package connsdk

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/transportpolicy"
)

// customHeaderAuth models the 71 connector definitions that authenticate with
// a NON-Authorization custom header. Go strips only Authorization,
// WWW-Authenticate and Cookie across cross-domain redirects, so this is
// precisely the credential class the redirect policy has to handle itself.
func customHeaderAuth(name, value string) AuthFunc {
	return func(_ context.Context, req *http.Request) error {
		req.Header.Set(name, value)
		return nil
	}
}

func drain(t *testing.T, rc io.ReadCloser) string {
	t.Helper()
	defer func() { _ = rc.Close() }()
	raw, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read stream: %v", err)
	}
	return string(raw)
}

// TestDoStreamReturnsOpenBody: DoStream hands back an OPEN body rather than a
// buffered []byte, which is the whole reason it exists — Response.Body is
// capped at 64 MiB, so a declared 100 MiB max_bytes cannot be satisfied
// through Do/DoLimited.
func TestDoStreamReturnsOpenBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("PDFBYTES"))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL}
	resp, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{})
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d", resp.Status)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/pdf" {
		t.Fatalf("content type = %q", got)
	}
	if got := drain(t, resp.Body); got != "PDFBYTES" {
		t.Fatalf("body = %q", got)
	}
}

// TestDoStreamSendsDeclaredAcceptMediaType keeps fixed binary representations
// declaration-owned. Callers cannot provide headers, but an operation such as
// GitHub's pull-request diff endpoint can select its documented media type.
func TestDoStreamSendsDeclaredAcceptMediaType(t *testing.T) {
	const mediaType = "application/vnd.github.diff"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != mediaType {
			t.Errorf("Accept = %q, want %q", got, mediaType)
		}
		_, _ = w.Write([]byte("diff --git a/a b/a\n"))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL}
	resp, err := r.DoStream(context.Background(), http.MethodGet, "/diff", nil, StreamOptions{Accept: mediaType})
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	if got := drain(t, resp.Body); !strings.HasPrefix(got, "diff --git") {
		t.Fatalf("body = %q", got)
	}
}

// TestDoStreamRejectsCrossOriginRedirect: fail closed by default. A download
// endpoint redirecting to a CDN must not silently proceed.
func TestDoStreamRejectsCrossOriginRedirect(t *testing.T) {
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("cdn-bytes"))
	}))
	defer cdn.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, cdn.URL+"/blob", http.StatusFound)
	}))
	defer origin.Close()

	r := &Requester{BaseURL: origin.URL, Auth: customHeaderAuth("X-API-Key", "supersecret")}
	_, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{})
	if err == nil {
		t.Fatal("cross-origin redirect must be refused by default")
	}
	if !strings.Contains(err.Error(), "cross-host") {
		t.Fatalf("error should name the cross-host guard, got: %v", err)
	}
}

// TestDoStreamStripsCustomAuthHeaderCrossOrigin is the core security
// assertion: when a cross-host hop IS explicitly permitted, the connector
// credential must not travel with it. Go would forward X-API-Key; the policy
// must remove it.
func TestDoStreamStripsCustomAuthHeaderCrossOrigin(t *testing.T) {
	var cdnSawAuth, cdnSawDefault, cdnSawCookie atomic.Value
	cdnSawAuth.Store("")
	cdnSawDefault.Store("")
	cdnSawCookie.Store("")
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cdnSawAuth.Store(r.Header.Get("X-API-Key"))
		cdnSawDefault.Store(r.Header.Get("X-Tenant-Token"))
		cdnSawCookie.Store(r.Header.Get("Authorization"))
		_, _ = w.Write([]byte("cdn-bytes"))
	}))
	defer cdn.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, cdn.URL+"/blob", http.StatusFound)
	}))
	defer origin.Close()

	r := &Requester{
		BaseURL:        origin.URL,
		Auth:           customHeaderAuth("X-API-Key", "supersecret"),
		DefaultHeaders: map[string]string{"X-Tenant-Token": "tenantsecret"},
	}
	resp, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{AllowCrossHost: true})
	if err != nil {
		t.Fatalf("DoStream with AllowCrossHost: %v", err)
	}
	if got := drain(t, resp.Body); got != "cdn-bytes" {
		t.Fatalf("body = %q", got)
	}
	if v := cdnSawAuth.Load().(string); v != "" {
		t.Fatalf("custom auth header leaked to cross-origin host: %q", v)
	}
	if v := cdnSawDefault.Load().(string); v != "" {
		t.Fatalf("default header leaked to cross-origin host: %q", v)
	}
	if v := cdnSawCookie.Load().(string); v != "" {
		t.Fatalf("authorization leaked to cross-origin host: %q", v)
	}
}

// TestDoStreamKeepsAuthSameOrigin: a same-origin redirect is normal and must
// still carry credentials, otherwise ordinary paginated/redirecting APIs break.
func TestDoStreamKeepsAuthSameOrigin(t *testing.T) {
	var sawKey atomic.Value
	sawKey.Store("")
	mux := http.NewServeMux()
	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/blob", http.StatusFound)
	})
	mux.HandleFunc("/blob", func(w http.ResponseWriter, r *http.Request) {
		sawKey.Store(r.Header.Get("X-API-Key"))
		_, _ = w.Write([]byte("same-origin-bytes"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Auth: customHeaderAuth("X-API-Key", "supersecret")}
	resp, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{})
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	if got := drain(t, resp.Body); got != "same-origin-bytes" {
		t.Fatalf("body = %q", got)
	}
	if v := sawKey.Load().(string); v != "supersecret" {
		t.Fatalf("same-origin redirect must keep credentials, got %q", v)
	}
}

func TestDoStreamAdmitsPermittedRedirectHops(t *testing.T) {
	t.Run("same origin", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/blob", http.StatusFound)
		})
		mux.HandleFunc("/blob", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("same-origin-bytes"))
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		assertStreamRedirectAdmissions(t, srv.URL, StreamOptions{})
	})

	t.Run("allowed host", func(t *testing.T) {
		cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("cdn-bytes"))
		}))
		defer cdn.Close()
		origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, cdn.URL+"/blob", http.StatusFound)
		}))
		defer origin.Close()

		assertStreamRedirectAdmissions(t, origin.URL, StreamOptions{AllowedHosts: []string{strings.TrimPrefix(cdn.URL, "http://")}})
	})
}

func assertStreamRedirectAdmissions(t *testing.T, baseURL string, opts StreamOptions) {
	t.Helper()
	var admissions []RateLimitRequest
	r := &Requester{
		BaseURL:        baseURL,
		DisableRetries: true,
		Admission: rateLimitAdmissionFunc(func(_ context.Context, request RateLimitRequest) error {
			admissions = append(admissions, request)
			return nil
		}),
	}

	resp, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, opts)
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	_ = drain(t, resp.Body)
	if got, want := admissions, []RateLimitRequest{{Method: http.MethodGet, Attempt: 1}, {Method: http.MethodGet, Attempt: 2}}; !slices.Equal(got, want) {
		t.Fatalf("admissions = %+v, want %+v", got, want)
	}
}

// TestDoStreamAllowedHostsPermitsNamedHost: an explicit per-operation
// allowlist entry permits exactly that host, and still sends no credentials.
func TestDoStreamAllowedHostsPermitsNamedHost(t *testing.T) {
	var sawKey atomic.Value
	sawKey.Store("")
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawKey.Store(r.Header.Get("X-API-Key"))
		_, _ = w.Write([]byte("cdn-bytes"))
	}))
	defer cdn.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, cdn.URL+"/blob", http.StatusFound)
	}))
	defer origin.Close()

	cdnHost := strings.TrimPrefix(cdn.URL, "http://")
	r := &Requester{BaseURL: origin.URL, Auth: customHeaderAuth("X-API-Key", "supersecret")}

	resp, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{AllowedHosts: []string{cdnHost}})
	if err != nil {
		t.Fatalf("allowlisted host must be permitted: %v", err)
	}
	if got := drain(t, resp.Body); got != "cdn-bytes" {
		t.Fatalf("body = %q", got)
	}
	if v := sawKey.Load().(string); v != "" {
		t.Fatalf("allowlisted cross-origin host must still receive no credentials, got %q", v)
	}

	// A host NOT on the list stays refused.
	if _, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{AllowedHosts: []string{"example.invalid:443"}}); err == nil {
		t.Fatal("non-allowlisted host must stay refused")
	}
}

// TestDoStreamRetryDiscardsPartialBody: doWithBody retries the whole request
// up to 5 times. A retry after partial bytes must restart from zero — the
// caller must never observe two attempts concatenated.
func TestDoStreamRetryDiscardsPartialBody(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("PARTIAL-GARBAGE"))
			return
		}
		_, _ = w.Write([]byte("CLEAN"))
	}))
	defer srv.Close()

	req := &Requester{
		BaseURL: srv.URL,
		Sleep:   func(context.Context, time.Duration) error { return nil },
	}
	resp, err := req.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{})
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	got := drain(t, resp.Body)
	if got != "CLEAN" {
		t.Fatalf("retry must restart from zero, got %q", got)
	}
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d, want 2", attempts.Load())
	}
}

func TestDoStreamReturnsTypedRateLimitError(t *testing.T) {
	now := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "90")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"limited"}`))
	}))
	defer srv.Close()

	r := &Requester{
		BaseURL:        srv.URL,
		DisableRetries: true,
		Now:            func() time.Time { return now },
	}
	_, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{})
	if err == nil {
		t.Fatal("DoStream: want rate-limit error")
	}
	var rateLimitErr *RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("error type = %T, want *RateLimitError", err)
	}
	if got, want := rateLimitErr.ResetAt, now.Add(90*time.Second); !got.Equal(want) {
		t.Fatalf("reset = %v, want %v", got, want)
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusTooManyRequests {
		t.Fatalf("wrapped HTTPError = %v, want 429", httpErr)
	}
}

func TestDoStreamDisableRetriesPreventsMutationTransportReplay(t *testing.T) {
	var mutationHits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/mutate" {
			atomic.AddInt32(&mutationHits, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	var injectedFailures int32
	var idempotencyHeaders int32
	dialer := net.Dialer{}
	transport := &http.Transport{
		MaxIdleConnsPerHost: 1,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			return requestWriteFailureConn{Conn: conn, fail: func(p []byte) bool {
				if !strings.HasPrefix(string(p), "POST /mutate ") {
					return false
				}
				if strings.Contains(string(p), "\r\nIdempotency-Key:") {
					atomic.AddInt32(&idempotencyHeaders, 1)
				}
				return atomic.CompareAndSwapInt32(&injectedFailures, 0, 1)
			}}, nil
		},
	}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport}
	primeHTTPConnection(t, client, srv.URL)

	r := &Requester{
		BaseURL:        srv.URL,
		Client:         client,
		DisableRetries: true,
		DefaultHeaders: map[string]string{"Idempotency-Key": "configured-key"},
	}
	resp, err := r.DoStream(context.Background(), http.MethodPost, "/mutate", nil, StreamOptions{})
	if err == nil {
		if resp != nil {
			_ = resp.Body.Close()
		}
		t.Fatal("DoStream: expected transport error")
	}
	if got, want := atomic.LoadInt32(&injectedFailures), int32(1); got != want {
		t.Fatalf("injected failures = %d, want %d", got, want)
	}
	if got := atomic.LoadInt32(&idempotencyHeaders); got != 0 {
		t.Fatalf("idempotency headers = %d, want none", got)
	}
	if got := atomic.LoadInt32(&mutationHits); got != 0 {
		t.Fatalf("mutation hits = %d, want no transport replay", got)
	}
}

func TestDoStreamDisableRetriesRejectsMutationRedirect(t *testing.T) {
	var initialHits int32
	var targetHits int32
	mux := http.NewServeMux()
	mux.HandleFunc("/initial", func(w http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&initialHits, 1)
		http.Redirect(w, req, "/target", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/target", func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&targetHits, 1)
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, DisableRetries: true}
	resp, err := r.DoStream(context.Background(), http.MethodPost, "/initial", nil, StreamOptions{})
	if resp != nil {
		_ = resp.Body.Close()
		t.Fatal("DoStream returned a response after refusing redirect")
	}
	if !errors.Is(err, transportpolicy.ErrRedirectRefused) {
		t.Fatalf("DoStream error = %v, want redirect refusal", err)
	}
	if got, want := atomic.LoadInt32(&initialHits), int32(1); got != want {
		t.Fatalf("initial hits = %d, want %d", got, want)
	}
	if got := atomic.LoadInt32(&targetHits); got != 0 {
		t.Fatalf("target hits = %d, want no redirected mutation", got)
	}
}

// TestDoStreamDoesNotMutateSharedClient: the redirect policy must be set on a
// clone. Mutating the shared client would silently apply this policy to every
// other request the connector makes.
func TestDoStreamDoesNotMutateSharedClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	shared := &http.Client{Timeout: 90 * time.Second}
	r := &Requester{BaseURL: srv.URL, Client: shared}
	resp, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{})
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	_ = drain(t, resp.Body)
	if shared.CheckRedirect != nil {
		t.Fatal("DoStream must not install CheckRedirect on the shared client")
	}
	if shared.Timeout != 90*time.Second {
		t.Fatal("DoStream must not mutate the shared client timeout")
	}
}

// TestDoStreamSlowProgressingDownloadOutlivesClientTimeout: http.Client.Timeout
// is a wall-clock deadline that starts at Do() and keeps running while the
// caller reads the returned body. A slow-but-progressing download that takes
// longer than the configured timeout must still deliver the full payload —
// liveness for a streamed body is the stall watchdog's job, not a wall-clock
// bound that doubles as a bandwidth requirement.
func TestDoStreamSlowProgressingDownloadOutlivesClientTimeout(t *testing.T) {
	payload := strings.Repeat("0123456789abcdef", 128) // 2 KiB
	const chunk = 256
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		fl, ok := w.(http.Flusher)
		if !ok {
			t.Error("test server must support flushing")
			return
		}
		for i := 0; i < len(payload); i += chunk {
			end := i + chunk
			if end > len(payload) {
				end = len(payload)
			}
			_, _ = w.Write([]byte(payload[i:end]))
			fl.Flush()
			time.Sleep(60 * time.Millisecond)
		}
	}))
	defer srv.Close()

	shared := &http.Client{Timeout: 300 * time.Millisecond}
	r := &Requester{BaseURL: srv.URL, Client: shared}
	resp, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{})
	if err != nil {
		t.Fatalf("DoStream: %v", err)
	}
	got := drain(t, resp.Body)
	if got != payload {
		t.Fatalf("slow-but-progressing download truncated: got %d bytes, want %d", len(got), len(payload))
	}
	if shared.Timeout != 300*time.Millisecond {
		t.Fatalf("DoStream mutated the shared client timeout: %v", shared.Timeout)
	}
}

// TestDoStreamHTTPErrorDoesNotLeakOpenBody: a terminal 4xx must return an
// *HTTPError with the body already closed, not a dangling reader.
func TestDoStreamHTTPErrorClosesBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"nope"}`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL}
	resp, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{})
	if err == nil {
		t.Fatal("want error for 404")
	}
	if resp != nil {
		t.Fatalf("response must be nil on error, got %+v", resp)
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusNotFound {
		t.Fatalf("want *HTTPError 404, got %v", err)
	}
}

// TestDoStreamRejectsRedirectLoop bounds the hop count so a hostile or buggy
// API cannot redirect forever.
func TestDoStreamRejectsRedirectLoop(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/next", http.StatusFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL}
	if _, err := r.DoStream(context.Background(), http.MethodGet, "/file", nil, StreamOptions{}); err == nil {
		t.Fatal("want redirect-loop refusal")
	}
}
