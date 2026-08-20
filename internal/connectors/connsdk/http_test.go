package connsdk

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/transportpolicy"
)

func noSleep(_ context.Context, _ time.Duration) error { return nil }

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type requestWriteFailureConn struct {
	net.Conn
	fail func([]byte) bool
}

func (c requestWriteFailureConn) Write(p []byte) (int, error) {
	if c.fail(p) {
		return 0, errors.New("injected request write failure")
	}
	return c.Conn.Write(p)
}

type advancingReadCloser struct {
	io.ReadCloser
	advance func()
}

func (r *advancingReadCloser) Read(p []byte) (int, error) {
	if r.advance != nil {
		r.advance()
		r.advance = nil
	}
	return r.ReadCloser.Read(p)
}

type partialResponseReadCloser struct {
	body   []byte
	err    error
	closed bool
}

func (r *partialResponseReadCloser) Read(p []byte) (int, error) {
	if len(r.body) == 0 {
		return 0, r.err
	}
	n := copy(p, r.body)
	r.body = r.body[n:]
	return n, r.err
}

func (r *partialResponseReadCloser) Close() error {
	r.closed = true
	return nil
}

func primeHTTPConnection(t *testing.T, client *http.Client, baseURL string) {
	t.Helper()
	resp, err := client.Get(baseURL + "/prime")
	if err != nil {
		t.Fatalf("prime connection: %v", err)
	}
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		_ = resp.Body.Close()
		t.Fatalf("drain prime response: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close prime response: %v", err)
	}
}

func TestRequesterDoJSONDecodesSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id": 7, "name": "ada"}`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	var out struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := r.DoJSON(context.Background(), http.MethodGet, "/thing", nil, nil, &out); err != nil {
		t.Fatalf("DoJSON error = %v", err)
	}
	if out.ID != 7 || out.Name != "ada" {
		t.Fatalf("decoded = %+v", out)
	}
}

func TestRequesterRetriesOn429ThenSucceeds(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	resp, err := r.Do(context.Background(), http.MethodGet, "/x", nil, nil)
	if err != nil {
		t.Fatalf("Do error = %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d", resp.Status)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("calls = %d, want 2", got)
	}
}

func TestRequesterRetainsPartialResponseOnBodyReadError(t *testing.T) {
	readErr := errors.New("injected response read failure")
	body := &partialResponseReadCloser{body: []byte("provider response prefix"), err: readErr}
	r := &Requester{
		BaseURL:        "https://example.invalid",
		DisableRetries: true,
		Client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusAccepted,
				Header:     http.Header{"X-Provider-Receipt": {"receipt-one", "receipt-two"}},
				Body:       body,
				Request:    req,
			}, nil
		})},
	}

	response, err := r.Do(context.Background(), http.MethodPost, "/writes", nil, map[string]string{"id": "one"})
	if !errors.Is(err, readErr) {
		t.Fatalf("Do error = %v, want response read failure", err)
	}
	if response == nil || response.Status != http.StatusAccepted || string(response.Body) != "provider response prefix" {
		t.Fatalf("response = %#v, want captured partial provider response", response)
	}
	if got := response.Header.Values("X-Provider-Receipt"); !slices.Equal(got, []string{"receipt-one", "receipt-two"}) {
		t.Fatalf("provider receipts = %#v, want both values", got)
	}
	if !body.closed {
		t.Fatal("response body was not closed after the read failure")
	}
}

func TestRequesterHonorsProviderRetryAfterBeyondFallbackCap(t *testing.T) {
	var calls int32
	now := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "90")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	var waits []time.Duration
	var jitterCalls int
	r := &Requester{
		BaseURL:    srv.URL,
		MaxRetries: 1,
		MaxBackoff: 30 * time.Second,
		Now:        func() time.Time { return now },
		Jitter: func(time.Duration) time.Duration {
			jitterCalls++
			return 0
		},
		Sleep: func(_ context.Context, d time.Duration) error {
			waits = append(waits, d)
			return nil
		},
	}
	if _, err := r.Do(context.Background(), http.MethodGet, "/x", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(waits) != 1 {
		t.Fatalf("waits = %v, want one wait", waits)
	}
	if got, want := waits[0], 90*time.Second; got != want {
		t.Fatalf("provider Retry-After wait = %v, want exact %v", got, want)
	}
	if jitterCalls != 0 {
		t.Fatalf("provider Retry-After invoked fallback jitter %d times", jitterCalls)
	}
	if got, want := atomic.LoadInt32(&calls), int32(2); got != want {
		t.Fatalf("HTTP calls = %d, want retry cap of %d", got, want)
	}
}

func TestRequesterWaitsOnlyUntilProviderResetAfterBodyDrain(t *testing.T) {
	receivedAt := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	now := receivedAt
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.Header().Set("Retry-After", "90")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"limited"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	baseTransport := http.DefaultTransport.(*http.Transport).Clone()
	t.Cleanup(baseTransport.CloseIdleConnections)
	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp, err := baseTransport.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body = &advancingReadCloser{
				ReadCloser: resp.Body,
				advance: func() {
					now = now.Add(60 * time.Second)
				},
			}
		}
		return resp, nil
	})}

	var waits []time.Duration
	r := &Requester{
		BaseURL:    srv.URL,
		Client:     client,
		MaxRetries: 1,
		Now:        func() time.Time { return now },
		Sleep: func(_ context.Context, d time.Duration) error {
			waits = append(waits, d)
			return nil
		},
	}
	if _, err := r.Do(context.Background(), http.MethodGet, "/limited", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got, want := waits, []time.Duration{30 * time.Second}; !slices.Equal(got, want) {
		t.Fatalf("waits = %v, want %v", got, want)
	}
	if got, want := atomic.LoadInt32(&calls), int32(2); got != want {
		t.Fatalf("HTTP calls = %d, want %d", got, want)
	}
}

type rateLimitAdmissionFunc func(context.Context, RateLimitRequest) error

func (f rateLimitAdmissionFunc) Admit(ctx context.Context, request RateLimitRequest) error {
	return f(ctx, request)
}

type rateLimitObserverFunc func(context.Context, RateLimitObservation)

func (f rateLimitObserverFunc) Observe(ctx context.Context, observation RateLimitObservation) {
	f(ctx, observation)
}

type rateLimitEventSinkFunc func(RateLimitEvent)

func (f rateLimitEventSinkFunc) RecordRateLimitEvent(event RateLimitEvent) {
	f(event)
}

func TestRequesterRateLimitEventsRecordDeadlineCutoffBeforeSend(t *testing.T) {
	var sends int32
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		atomic.AddInt32(&sends, 1)
	}))
	defer srv.Close()

	var events []RateLimitEvent
	r := &Requester{
		BaseURL: srv.URL,
		Admission: rateLimitAdmissionFunc(func(ctx context.Context, _ RateLimitRequest) error {
			<-ctx.Done()
			return ctx.Err()
		}),
		RateLimitEvents: rateLimitEventSinkFunc(func(event RateLimitEvent) {
			events = append(events, event)
		}),
	}
	r.RateLimitAdmissionTimeout = 10 * time.Millisecond
	_, err := r.Do(context.Background(), http.MethodGet, "/deadline", nil, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Do() error = %v, want deadline exceeded", err)
	}
	if got, want := atomic.LoadInt32(&sends), int32(0); got != want {
		t.Fatalf("provider sends = %d, want %d when admission reaches its deadline", got, want)
	}
	if got, want := len(events), 2; got != want {
		t.Fatalf("rate-limit events = %+v, want wait and not_sent", events)
	}
	if got := events[0]; got.Type != RateLimitEventWait || got.Method != http.MethodGet || got.Attempt != 1 || got.DurationMS <= 0 {
		t.Fatalf("wait event = %+v, want first GET wait with positive duration", got)
	}
	if got := events[1]; got.Type != RateLimitEventNotSent || got.Method != http.MethodGet || got.Attempt != 1 || got.Reason != "deadline_cutoff" {
		t.Fatalf("not-sent event = %+v, want first GET deadline cutoff", got)
	}
}

func TestRequesterRateLimitEventsRecordAttemptAndProviderReset(t *testing.T) {
	now := time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)
	reset := now.Add(time.Hour)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	var events []RateLimitEvent
	r := &Requester{
		BaseURL: srv.URL,
		Now:     func() time.Time { return now },
		RateLimitEvents: rateLimitEventSinkFunc(func(event RateLimitEvent) {
			events = append(events, event)
		}),
	}
	if _, err := r.Do(context.Background(), http.MethodGet, "/reset", nil, nil); err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	if got, want := len(events), 2; got != want {
		t.Fatalf("rate-limit events = %+v, want attempt and reset", events)
	}
	if got := events[0]; got.Type != RateLimitEventAttempt || got.Method != http.MethodGet || got.Attempt != 1 {
		t.Fatalf("attempt event = %+v, want first GET attempt", got)
	}
	if got := events[1]; got.Type != RateLimitEventReset || got.Method != http.MethodGet || got.Attempt != 1 || !got.ResetAt.Equal(reset) {
		t.Fatalf("reset event = %+v, want provider reset %s", got, reset)
	}
}

func TestRequesterFallbackRetryUsesBoundedFullJitter(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	var jitterCaps []time.Duration
	var waits []time.Duration
	r := &Requester{
		BaseURL:     srv.URL,
		MaxRetries:  1,
		BaseBackoff: time.Second,
		MaxBackoff:  30 * time.Second,
		Jitter: func(cap time.Duration) time.Duration {
			jitterCaps = append(jitterCaps, cap)
			return cap + time.Second // requester must clamp an untrusted implementation.
		},
		Sleep: func(_ context.Context, d time.Duration) error {
			waits = append(waits, d)
			return nil
		},
	}
	if _, err := r.Do(context.Background(), http.MethodGet, "/x", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got, want := jitterCaps, []time.Duration{time.Second}; !slices.Equal(got, want) {
		t.Fatalf("jitter caps = %v, want %v", got, want)
	}
	if got, want := waits, []time.Duration{time.Second}; !slices.Equal(got, want) {
		t.Fatalf("fallback waits = %v, want bounded full-jitter wait %v", got, want)
	}
}

func TestRequesterAdmissionPreventsInitialLogicalSend(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		atomic.AddInt32(&hits, 1)
	}))
	defer srv.Close()

	ctx := context.Background()
	var admissions []RateLimitRequest
	r := &Requester{
		BaseURL: srv.URL,
		Admission: rateLimitAdmissionFunc(func(ctx context.Context, request RateLimitRequest) error {
			admissions = append(admissions, request)
			return context.Canceled
		}),
	}
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "json",
			run: func() error {
				_, err := r.Do(ctx, http.MethodPost, "/json", nil, map[string]string{"name": "widget"})
				return err
			},
		},
		{
			name: "form",
			run: func() error {
				_, err := r.DoForm(ctx, http.MethodPost, "/form", nil, url.Values{"name": {"widget"}})
				return err
			},
		},
		{
			name: "multipart",
			run: func() error {
				_, err := r.DoMultipart(ctx, http.MethodPost, "/multipart", nil, MultipartForm{Fields: map[string]string{"name": "widget"}})
				return err
			},
		},
		{
			name: "stream",
			run: func() error {
				_, err := r.DoStream(ctx, http.MethodGet, "/stream", nil, StreamOptions{})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); !errors.Is(err, context.Canceled) {
				t.Fatalf("request error = %v, want context cancellation", err)
			}
		})
	}
	if got, want := atomic.LoadInt32(&hits), int32(0); got != want {
		t.Fatalf("server hits = %d, want %d", got, want)
	}
	if got, want := len(admissions), len(tests); got != want {
		t.Fatalf("admissions = %d, want %d", got, want)
	}
	for _, admission := range admissions {
		if admission.Attempt != 1 {
			t.Fatalf("admission attempt = %d, want 1", admission.Attempt)
		}
	}
}

func TestRequesterAdmissionHonorsCallerCancellationBeforeLogicalSend(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		atomic.AddInt32(&hits, 1)
	}))
	defer srv.Close()

	started := make(chan struct{})
	r := &Requester{
		BaseURL: srv.URL,
		Admission: rateLimitAdmissionFunc(func(ctx context.Context, _ RateLimitRequest) error {
			close(started)
			<-ctx.Done()
			return ctx.Err()
		}),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errs := make(chan error, 1)
	go func() {
		_, err := r.Do(ctx, http.MethodGet, "/wait", nil, nil)
		errs <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("admission did not receive request context")
	}
	cancel()
	select {
	case err := <-errs:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("request error = %v, want context cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("request did not return when admission context was canceled")
	}
	if got, want := atomic.LoadInt32(&hits), int32(0); got != want {
		t.Fatalf("server hits = %d, want %d", got, want)
	}
}

func TestRequesterAdmitsEachLogicalRetryAttempt(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	var admissions []RateLimitRequest
	r := &Requester{
		BaseURL:    srv.URL,
		MaxRetries: 1,
		Admission: rateLimitAdmissionFunc(func(_ context.Context, request RateLimitRequest) error {
			admissions = append(admissions, request)
			return nil
		}),
		Jitter: func(time.Duration) time.Duration { return 0 },
		Sleep:  noSleep,
	}
	if _, err := r.Do(context.Background(), http.MethodGet, "/retry", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got, want := admissions, []RateLimitRequest{{Method: http.MethodGet, Attempt: 1}, {Method: http.MethodGet, Attempt: 2}}; !slices.Equal(got, want) {
		t.Fatalf("admissions = %+v, want %+v", got, want)
	}
}

func TestRequesterAdmitsPermittedRedirectHop(t *testing.T) {
	var finalHits int32
	mux := http.NewServeMux()
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
	})
	mux.HandleFunc("/final", func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&finalHits, 1)
		w.Header().Set("RateLimit-Remaining", "99")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	var admissions []RateLimitRequest
	var observations []RateLimitObservation
	r := &Requester{
		BaseURL:        srv.URL,
		DisableRetries: true,
		Admission: rateLimitAdmissionFunc(func(_ context.Context, request RateLimitRequest) error {
			admissions = append(admissions, request)
			return nil
		}),
		Observer: rateLimitObserverFunc(func(_ context.Context, observation RateLimitObservation) {
			observations = append(observations, observation)
		}),
	}

	if _, err := r.Do(context.Background(), http.MethodGet, "/redirect", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got, want := admissions, []RateLimitRequest{{Method: http.MethodGet, Attempt: 1}, {Method: http.MethodGet, Attempt: 2}}; !slices.Equal(got, want) {
		t.Fatalf("admissions = %+v, want %+v", got, want)
	}
	if got, want := atomic.LoadInt32(&finalHits), int32(1); got != want {
		t.Fatalf("final hits = %d, want %d", got, want)
	}
	if got, want := len(observations), 1; got != want {
		t.Fatalf("observations = %d, want %d", got, want)
	}
	if got, want := observations[0].Attempt, 2; got != want {
		t.Fatalf("observation attempt = %d, want %d", got, want)
	}
}

func TestRequesterReturnsTypedRateLimitErrorAndObservation(t *testing.T) {
	const fixtureSecret = "fixture-rate-limit-secret"
	now := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "90")
		w.Header().Set("RateLimit-Limit", "100")
		w.Header().Set("RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"` + fixtureSecret + `"}`))
	}))
	defer srv.Close()

	var observations []RateLimitObservation
	r := &Requester{
		BaseURL:        srv.URL,
		DisableRetries: true,
		Now:            func() time.Time { return now },
		Observer: rateLimitObserverFunc(func(_ context.Context, observation RateLimitObservation) {
			observations = append(observations, observation)
		}),
	}
	_, err := r.Do(context.Background(), http.MethodGet, "/limited", nil, nil)
	if err == nil {
		t.Fatal("Do: expected rate-limit error")
	}
	var rateLimitErr *RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("error type = %T, want *RateLimitError", err)
	}
	if got, want := rateLimitErr.ResetAt, now.Add(90*time.Second); !got.Equal(want) {
		t.Fatalf("RateLimitError reset = %v, want %v", got, want)
	}
	if !rateLimitErr.HasReset || !rateLimitErr.HasRetryAfter {
		t.Fatalf("RateLimitError reset presence = %+v, want parsed provider timing", rateLimitErr)
	}
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || httpErr.Status != http.StatusTooManyRequests {
		t.Fatalf("wrapped HTTPError = %v, want status 429", httpErr)
	}
	if strings.Contains(err.Error(), fixtureSecret) {
		t.Fatalf("RateLimitError output leaked fixture secret: %q", err)
	}
	if got, want := len(observations), 1; got != want {
		t.Fatalf("observations = %d, want %d", got, want)
	}
	observation := observations[0]
	if observation.Source != RateLimitObservationSourceRetryAfter || !observation.Attempted || observation.Attempt != 1 || observation.Status != http.StatusTooManyRequests {
		t.Fatalf("observation = %+v, want attempted Retry-After 429", observation)
	}
	if got, want := observation.ResetAt, now.Add(90*time.Second); !got.Equal(want) {
		t.Fatalf("observation reset = %v, want %v", got, want)
	}
	if got, want := observation.RetryAfter, 90*time.Second; got != want {
		t.Fatalf("observation RetryAfter = %v, want %v", got, want)
	}
	if got, want := observation.Limit, int64(100); got != want {
		t.Fatalf("observation Limit = %d, want %d", got, want)
	}
	if !observation.HasLimit {
		t.Fatal("observation did not mark RateLimit-Limit present")
	}
	if got, want := observation.Remaining, int64(0); got != want {
		t.Fatalf("observation Remaining = %d, want %d", got, want)
	}
	if !observation.HasRemaining {
		t.Fatal("observation did not preserve remaining=0 as present")
	}
	if output := fmt.Sprintf("%+v", observation); strings.Contains(output, fixtureSecret) {
		t.Fatalf("observation output leaked fixture secret: %q", output)
	}
}

func TestRequesterSnapshotsRateLimitTimingBeforeDrainingBody(t *testing.T) {
	receivedAt := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	now := receivedAt
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "90")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"limited"}`))
	}))
	defer srv.Close()

	baseTransport := http.DefaultTransport.(*http.Transport).Clone()
	t.Cleanup(baseTransport.CloseIdleConnections)
	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		resp, err := baseTransport.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		resp.Body = &advancingReadCloser{
			ReadCloser: resp.Body,
			advance: func() {
				now = now.Add(60 * time.Second)
			},
		}
		return resp, nil
	})}

	var observations []RateLimitObservation
	r := &Requester{
		BaseURL:        srv.URL,
		Client:         client,
		DisableRetries: true,
		Now:            func() time.Time { return now },
		Observer: rateLimitObserverFunc(func(_ context.Context, observation RateLimitObservation) {
			observations = append(observations, observation)
		}),
	}
	_, err := r.Do(context.Background(), http.MethodGet, "/limited", nil, nil)
	var rateLimitErr *RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("error type = %T, want *RateLimitError", err)
	}
	wantReset := receivedAt.Add(90 * time.Second)
	if got := rateLimitErr.ResetAt; !got.Equal(wantReset) {
		t.Fatalf("RateLimitError reset = %v, want %v", got, wantReset)
	}
	if got, want := now, receivedAt.Add(60*time.Second); !got.Equal(want) {
		t.Fatalf("clock after body read = %v, want %v", got, want)
	}
	if got, want := len(observations), 1; got != want {
		t.Fatalf("observations = %d, want %d", got, want)
	}
	if got := observations[0].ResetAt; !got.Equal(wantReset) {
		t.Fatalf("observation reset = %v, want %v", got, wantReset)
	}
}

// net/http treats a request carrying Idempotency-Key as replayable after some
// transport failures. DisableRetries is the no-retry contract used by
// rest_write, so it must remove that implicit retry signal as well as its own
// retry loop and 401 refresh path.
func TestRequesterDisableRetriesMakesMutationNonReplayable(t *testing.T) {
	var sawGetBody bool
	var sawIdempotencyKey string
	var sawClose bool
	r := &Requester{
		BaseURL:        "https://example.invalid",
		DisableRetries: true,
		DefaultHeaders: map[string]string{"Idempotency-Key": "configured-key"},
		Client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			sawGetBody = req.GetBody != nil
			sawIdempotencyKey = req.Header.Get("Idempotency-Key")
			sawClose = req.Close
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			}, nil
		})},
	}

	if _, err := r.Do(context.Background(), http.MethodPost, "/widgets", nil, map[string]string{"name": "widget"}); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if sawGetBody {
		t.Fatal("DisableRetries left Request.GetBody available for transport replay")
	}
	if sawIdempotencyKey != "" {
		t.Fatalf("DisableRetries left Idempotency-Key=%q on a no-retry request", sawIdempotencyKey)
	}
	if !sawClose {
		t.Fatal("DisableRetries did not require a fresh connection for a mutation")
	}
}

func TestNoReplayClientKeepsHTTPNegotiationAvailable(t *testing.T) {
	client := noReplayClient(&http.Client{Transport: http.DefaultTransport})
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("strict transport type = %T, want *http.Transport", client.Transport)
	}
	if !transport.DisableKeepAlives {
		t.Fatal("strict transport must keep one-use connections enabled")
	}
	if transport.Protocols != nil {
		t.Fatal("strict transport forced an HTTP protocol instead of preserving normal negotiation")
	}
}

func TestNoReplayClientALPNMatchesConfiguredProtocol(t *testing.T) {
	var handlerProtocol string
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		handlerProtocol = req.Proto
		w.WriteHeader(http.StatusOK)
	}))
	srv.EnableHTTP2 = true
	srv.StartTLS()
	defer srv.Close()

	base := srv.Client()
	transport, ok := base.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("test transport type = %T, want *http.Transport", base.Transport)
	}
	transport = transport.Clone()
	tlsConfig := transport.TLSClientConfig.Clone()
	tlsConfig.NextProtos = []string{"h2", "http/1.1"}
	transport.TLSClientConfig = tlsConfig
	base.Transport = transport

	strictClient := noReplayClient(base)
	strictTransport, ok := strictClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("strict transport type = %T, want *http.Transport", strictClient.Transport)
	}
	wantALPN := "http/1.1"
	wantHandlerProtocol := "HTTP/1.1"
	if strictTransport.Protocols == nil || strictTransport.Protocols.HTTP2() {
		wantALPN = "h2"
		wantHandlerProtocol = "HTTP/2.0"
	}

	requester := &Requester{BaseURL: srv.URL, Client: strictClient, DisableRetries: true}
	var negotiatedProtocol string
	ctx := httptrace.WithClientTrace(context.Background(), &httptrace.ClientTrace{
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err == nil {
				negotiatedProtocol = state.NegotiatedProtocol
			}
		},
	})
	_, err := requester.Do(ctx, http.MethodPost, "/mutate", nil, map[string]string{"name": "widget"})
	if got, want := negotiatedProtocol, wantALPN; got != want {
		t.Fatalf("strict TLS negotiated protocol = %q, want %q", got, want)
	}
	if err != nil {
		t.Fatalf("strict mutation: %v", err)
	}
	if got, want := handlerProtocol, wantHandlerProtocol; got != want {
		t.Fatalf("handler protocol = %q, want %q", got, want)
	}
}

func TestRequesterDisableRetriesPreventsBodylessMutationTransportReplay(t *testing.T) {
	var mutationHits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/mutate" {
			atomic.AddInt32(&mutationHits, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	var injectedFailures int32
	dialer := net.Dialer{}
	transport := &http.Transport{
		MaxIdleConnsPerHost: 1,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			return requestWriteFailureConn{Conn: conn, fail: func(p []byte) bool {
				return strings.HasPrefix(string(p), "DELETE /mutate ") && atomic.CompareAndSwapInt32(&injectedFailures, 0, 1)
			}}, nil
		},
	}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport}
	primeHTTPConnection(t, client, srv.URL)

	r := &Requester{BaseURL: srv.URL, Client: client, DisableRetries: true}
	if _, err := r.Do(context.Background(), http.MethodDelete, "/mutate", nil, nil); err == nil {
		t.Fatal("Do: expected transport error")
	}
	if got, want := atomic.LoadInt32(&injectedFailures), int32(1); got != want {
		t.Fatalf("injected failures = %d, want %d", got, want)
	}
	if got := atomic.LoadInt32(&mutationHits); got != 0 {
		t.Fatalf("mutation hits = %d, want no transport replay", got)
	}
}

// The requestBody field is typed as io.Reader, but the concrete *bytes.Reader
// must survive the interface handoff to http.NewRequest so ordinary requests
// remain replayable when Go detects a stale pooled connection before a write.
func TestRequesterKeepsJSONBodyReplayableBeforeNoReplayPolicy(t *testing.T) {
	var sawGetBody bool
	r := &Requester{
		BaseURL: "https://example.invalid",
		Client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			sawGetBody = req.GetBody != nil
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			}, nil
		})},
	}

	if _, err := r.Do(context.Background(), http.MethodPost, "/mutate", nil, map[string]string{"name": "widget"}); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if !sawGetBody {
		t.Fatal("JSON request lost Request.GetBody before the no-replay policy ran")
	}
}

func TestRequesterReplaysReplayableJSONPostAfterStaleIdleWriteFailure(t *testing.T) {
	var mutationHits, injectedFailures, dials int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/mutate" {
			atomic.AddInt32(&mutationHits, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dialer := net.Dialer{}
	transport := &http.Transport{
		MaxIdleConnsPerHost: 1,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			dial := atomic.AddInt32(&dials, 1)
			return requestWriteFailureConn{Conn: conn, fail: func(p []byte) bool {
				return dial == 1 && strings.HasPrefix(string(p), "POST /mutate ") && atomic.CompareAndSwapInt32(&injectedFailures, 0, 1)
			}}, nil
		},
	}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport}
	primeHTTPConnection(t, client, srv.URL)

	r := &Requester{BaseURL: srv.URL, Client: client}
	resp, err := r.Do(context.Background(), http.MethodPost, "/mutate", nil, map[string]string{"name": "widget"})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Status, http.StatusOK)
	}
	if got, want := atomic.LoadInt32(&injectedFailures), int32(1); got != want {
		t.Fatalf("injected failures = %d, want %d", got, want)
	}
	if got := atomic.LoadInt32(&dials); got < 2 {
		t.Fatalf("dials = %d, want a fresh connection after the stale idle failure", got)
	}
	if got, want := atomic.LoadInt32(&mutationHits), int32(1); got != want {
		t.Fatalf("mutation hits = %d, want %d after safe replay", got, want)
	}
}

func TestRequesterStrictMutationAvoidsStaleIdleConnectionReplay(t *testing.T) {
	var mutationHits, injectedFailures, dials int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/mutate" {
			atomic.AddInt32(&mutationHits, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dialer := net.Dialer{}
	transport := &http.Transport{
		MaxIdleConnsPerHost: 1,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			dial := atomic.AddInt32(&dials, 1)
			return requestWriteFailureConn{Conn: conn, fail: func(p []byte) bool {
				return dial == 1 && strings.HasPrefix(string(p), "POST /mutate ") && atomic.CompareAndSwapInt32(&injectedFailures, 0, 1)
			}}, nil
		},
	}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport}
	primeHTTPConnection(t, client, srv.URL)

	r := &Requester{BaseURL: srv.URL, Client: client, DisableRetries: true}
	resp, err := r.Do(context.Background(), http.MethodPost, "/mutate", nil, map[string]string{"name": "widget"})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Status, http.StatusOK)
	}
	if got := atomic.LoadInt32(&injectedFailures); got != 0 {
		t.Fatalf("strict mutation reused the stale idle connection; injected failures = %d", got)
	}
	if got := atomic.LoadInt32(&dials); got < 2 {
		t.Fatalf("dials = %d, want the strict mutation to use a fresh connection", got)
	}
	if got, want := atomic.LoadInt32(&mutationHits), int32(1); got != want {
		t.Fatalf("mutation hits = %d, want %d (exactly one dispatch)", got, want)
	}
}

func TestRequesterDisableRetriesRejectsMutationRedirect(t *testing.T) {
	for _, status := range []int{http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var initialHits int32
			var targetHits int32
			mux := http.NewServeMux()
			mux.HandleFunc("/initial", func(w http.ResponseWriter, req *http.Request) {
				atomic.AddInt32(&initialHits, 1)
				http.Redirect(w, req, "/target", status)
			})
			mux.HandleFunc("/target", func(w http.ResponseWriter, _ *http.Request) {
				atomic.AddInt32(&targetHits, 1)
				w.WriteHeader(http.StatusOK)
			})
			srv := httptest.NewServer(mux)
			defer srv.Close()

			r := &Requester{BaseURL: srv.URL, DisableRetries: true}
			_, err := r.Do(context.Background(), http.MethodPost, "/initial", nil, nil)
			if !errors.Is(err, transportpolicy.ErrRedirectRefused) {
				t.Fatalf("Do error = %v, want redirect refusal", err)
			}
			if got, want := atomic.LoadInt32(&initialHits), int32(1); got != want {
				t.Fatalf("initial hits = %d, want %d", got, want)
			}
			if got := atomic.LoadInt32(&targetHits); got != 0 {
				t.Fatalf("target hits = %d, want no redirected mutation", got)
			}
		})
	}
}

func TestRequesterAdmitsReplayableReadOncePerLogicalAttempt(t *testing.T) {
	var readHits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/read" {
			atomic.AddInt32(&readHits, 1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	var readWrites int32
	var injectedFailures int32
	dialer := net.Dialer{}
	transport := &http.Transport{
		MaxIdleConnsPerHost: 1,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			return requestWriteFailureConn{Conn: conn, fail: func(p []byte) bool {
				if !strings.HasPrefix(string(p), "GET /read ") {
					return false
				}
				atomic.AddInt32(&readWrites, 1)
				return atomic.CompareAndSwapInt32(&injectedFailures, 0, 1)
			}}, nil
		},
	}
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport}
	primeHTTPConnection(t, client, srv.URL)

	var admissions []RateLimitRequest
	r := &Requester{
		BaseURL: srv.URL,
		Client:  client,
		Sleep:   noSleep,
		Admission: rateLimitAdmissionFunc(func(_ context.Context, request RateLimitRequest) error {
			admissions = append(admissions, request)
			return nil
		}),
	}
	if _, err := r.Do(context.Background(), http.MethodGet, "/read", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got, want := admissions, []RateLimitRequest{{Method: http.MethodGet, Attempt: 1}}; !slices.Equal(got, want) {
		t.Fatalf("admissions = %+v, want %+v", got, want)
	}
	if got, want := atomic.LoadInt32(&injectedFailures), int32(1); got != want {
		t.Fatalf("injected failures = %d, want %d", got, want)
	}
	if got, want := atomic.LoadInt32(&readWrites), int32(2); got != want {
		t.Fatalf("read writes = %d, want %d after replay", got, want)
	}
	if got, want := atomic.LoadInt32(&readHits), int32(1); got != want {
		t.Fatalf("read hits = %d, want %d", got, want)
	}
}

func TestRequesterDoLimitedCapsCapturedBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"message":"0123456789"}`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	resp, err := r.DoLimited(context.Background(), http.MethodGet, "/x", nil, nil, 8)
	if err != nil {
		t.Fatalf("DoLimited error = %v", err)
	}
	if got, want := len(resp.Body), 9; got != want {
		t.Fatalf("captured body bytes = %d, want %d", got, want)
	}
}

func TestRequesterRetriesOn503(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep, MaxRetries: 5}
	if _, err := r.Do(context.Background(), http.MethodGet, "/x", nil, nil); err != nil {
		t.Fatalf("Do error = %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}

func TestRequesterReturnsHTTPErrorOn404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"nope"}`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	_, err := r.Do(context.Background(), http.MethodGet, "/missing", nil, nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	if httpErr.Status != http.StatusNotFound {
		t.Fatalf("status = %d", httpErr.Status)
	}
}

func TestHTTPErrorErrorRedactsURLQueryAndBody(t *testing.T) {
	err := (&HTTPError{Status: http.StatusUnauthorized, URL: "https://api.example.test/items?api_key=secret-token", Body: `{"error":"secret-token denied"}`}).Error()
	for _, leaked := range []string{"secret-token", "api_key=", "denied"} {
		if strings.Contains(err, leaked) {
			t.Fatalf("HTTPError leaked %q in %q", leaked, err)
		}
	}
	if !strings.Contains(err, "http 401") || !strings.Contains(err, "https://api.example.test/items") {
		t.Fatalf("HTTPError lost useful context: %q", err)
	}
}

func TestRequesterDoJSONDecodeErrorDoesNotIncludeRequestURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"broken"`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Auth: APIKeyQuery("api_key", "secret-token"), Sleep: noSleep}
	var out map[string]any
	err := r.DoJSON(context.Background(), http.MethodGet, "/items", nil, nil, &out)
	if err == nil {
		t.Fatal("expected decode error")
	}
	for _, leaked := range []string{srv.URL, "secret-token", "api_key"} {
		if strings.Contains(err.Error(), leaked) {
			t.Fatalf("decode error leaked %q in %q", leaked, err.Error())
		}
	}
}

func TestRequesterDoesNotRetry4xx(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	if _, err := r.Do(context.Background(), http.MethodGet, "/x", nil, nil); err == nil {
		t.Fatal("expected error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("calls = %d, want 1 (no retry on 400)", got)
	}
}

func TestRequesterHonorsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	if _, err := r.Do(ctx, http.MethodGet, "/x", nil, nil); err == nil {
		t.Fatal("expected context error")
	}
}

func TestRequesterDoFormEncodesBodyAndAuth(t *testing.T) {
	var sawContentType, sawAuth, sawBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawContentType = r.Header.Get("Content-Type")
		sawAuth = r.Header.Get("Authorization")
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		sawBody = r.PostForm.Get("email")
		_, _ = w.Write([]byte(`{"id":"cus_1"}`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Auth: Bearer("sk_test_1"), Sleep: noSleep}
	form := map[string][]string{"email": {"a@example.com"}, "name": {"Ada"}}
	resp, err := r.DoForm(context.Background(), http.MethodPost, "/customers", nil, form)
	if err != nil {
		t.Fatalf("DoForm error = %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d", resp.Status)
	}
	if sawContentType != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type = %q", sawContentType)
	}
	if sawAuth != "Bearer sk_test_1" {
		t.Fatalf("Authorization = %q", sawAuth)
	}
	if sawBody != "a@example.com" {
		t.Fatalf("form email = %q", sawBody)
	}
}

func TestRequesterDoMultipartEncodesFileAndAuth(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/payload.txt"
	if err := os.WriteFile(filePath, []byte("hello multipart"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	var sawAuth, sawField, sawFile string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
			t.Fatalf("Content-Type = %q, want multipart boundary", r.Header.Get("Content-Type"))
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		sawField = r.MultipartForm.Value["source"][0]
		fh := r.MultipartForm.File["mediaFile"][0]
		f, err := fh.Open()
		if err != nil {
			t.Fatalf("Open part: %v", err)
		}
		defer func() { _ = f.Close() }()
		raw, _ := io.ReadAll(f)
		sawFile = string(raw)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Auth: Bearer("test-token"), Sleep: noSleep}
	resp, err := r.DoMultipart(context.Background(), http.MethodPut, "/upload", nil, MultipartForm{
		Fields: map[string]string{"source": "recorder"},
		Files:  []MultipartFile{{FieldName: "mediaFile", Path: filePath, ContentType: "text/plain", MaxBytes: 1024}},
	})
	if err != nil {
		t.Fatalf("DoMultipart error = %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("status = %d", resp.Status)
	}
	if sawAuth != "Bearer test-token" || sawField != "recorder" || sawFile != "hello multipart" {
		t.Fatalf("auth=%q field=%q file=%q", sawAuth, sawField, sawFile)
	}
}

func TestRequesterDoMultipartLimitedBoundsCapturedResponse(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/payload.txt"
	if err := os.WriteFile(filePath, []byte("bounded multipart request"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		_, _ = w.Write([]byte("this response is deliberately longer than the declared operation cap"))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	resp, err := r.DoMultipartLimited(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{FieldName: "mediaFile", Path: filePath, MaxBytes: 1024}},
	}, 4)
	if err != nil {
		t.Fatalf("DoMultipartLimited: %v", err)
	}
	if got := len(resp.Body); got != 5 {
		t.Fatalf("captured response bytes = %d, want max_bytes + one = 5", got)
	}
}

func TestRequesterDoMultipartPreservesEarlyTerminalResponse(t *testing.T) {
	now := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	r := &Requester{
		BaseURL:        "https://example.invalid",
		DisableRetries: true,
		Now:            func() time.Time { return now },
		Client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			// An HTTP server may reject an upload before consuming its complete
			// streamed body. Closing it here models that transport outcome without
			// relying on scheduling between an httptest server and the pipe writer.
			if err := req.Body.Close(); err != nil {
				t.Fatalf("close streamed request body: %v", err)
			}
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": {"90"}},
				Body:       io.NopCloser(strings.NewReader(`{"error":"provider rejected upload"}`)),
				Request:    req,
			}, nil
		})},
	}

	_, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Fields: map[string]string{"message": "fixture upload"},
	})
	if err == nil {
		t.Fatal("DoMultipart error = nil, want the early provider response")
	}
	var rateLimitErr *RateLimitError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("DoMultipart error = %T %v, want *RateLimitError", err, err)
	}
	if rateLimitErr.HTTPError == nil || rateLimitErr.HTTPError.Body != `{"error":"provider rejected upload"}` {
		t.Fatalf("HTTP error = %#v, want the complete provider response", rateLimitErr.HTTPError)
	}
	if !rateLimitErr.HasRetryAfter || !rateLimitErr.ResetAt.Equal(now.Add(90*time.Second)) {
		t.Fatalf("rate-limit error = %#v, want parsed Retry-After reset", rateLimitErr)
	}
}

func TestRequesterDoMultipartRetainsTerminalResponseOnProducerCleanupError(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/payload.txt"
	if err := os.WriteFile(filePath, []byte("payload"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	providerBody := `{"accepted":true,"credential":"provider-returned-token"}`
	r := &Requester{
		BaseURL:        "https://example.invalid",
		DisableRetries: true,
		Sleep:          noSleep,
		Auth: AuthFunc(func(context.Context, *http.Request) error {
			return os.Remove(filePath)
		}),
		Client: &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			_, readErr := io.Copy(io.Discard, req.Body)
			if closeErr := req.Body.Close(); readErr == nil && closeErr != nil {
				readErr = closeErr
			}
			if readErr == nil {
				t.Fatal("multipart producer unexpectedly completed")
			}
			return &http.Response{
				StatusCode: http.StatusAccepted,
				Header:     http.Header{"X-Provider-Receipt": {"receipt-one", "receipt-two"}},
				Body:       io.NopCloser(strings.NewReader(providerBody)),
				Request:    req,
			}, nil
		})},
	}

	response, err := r.DoMultipartLimited(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Fields: map[string]string{"hold": strings.Repeat("x", 1<<20)},
		Files:  []MultipartFile{{FieldName: "media", Path: filePath, MaxBytes: 1024}},
	}, 1024)
	if err == nil || !strings.Contains(err.Error(), "send request body") {
		t.Fatalf("DoMultipartLimited error = %v, want producer cleanup diagnostic", err)
	}
	if strings.Contains(err.Error(), providerBody) {
		t.Fatalf("cleanup diagnostic leaked provider body: %q", err)
	}
	if response == nil {
		t.Fatal("DoMultipartLimited dropped terminal provider response")
	}
	if response.Status != http.StatusAccepted || string(response.Body) != providerBody {
		t.Fatalf("terminal response = %#v, want accepted provider response", response)
	}
	if got := response.Header.Values("X-Provider-Receipt"); !slices.Equal(got, []string{"receipt-one", "receipt-two"}) {
		t.Fatalf("terminal receipt = %#v, want both provider values", got)
	}
}

func TestRequesterDoMultipartRetriesWithReopenedFile(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/payload.txt"
	if err := os.WriteFile(filePath, []byte("retry payload"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := atomic.AddInt32(&calls, 1)
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		fh := r.MultipartForm.File["mediaFile"][0]
		f, err := fh.Open()
		if err != nil {
			t.Fatalf("Open part: %v", err)
		}
		raw, _ := io.ReadAll(f)
		_ = f.Close()
		if string(raw) != "retry payload" {
			t.Fatalf("attempt %d file body = %q", call, string(raw))
		}
		if call == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep, MaxRetries: 1}
	_, err := r.DoMultipart(context.Background(), http.MethodPut, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{FieldName: "mediaFile", Path: filePath, MaxBytes: 1024}},
	})
	if err != nil {
		t.Fatalf("DoMultipart error = %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Fatalf("calls = %d, want 2", got)
	}
}

func TestRequesterDoMultipartRejectsGrowthAfterPreflightValidation(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/payload.txt"
	if err := os.WriteFile(filePath, []byte("1234"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := &Requester{
		BaseURL: srv.URL,
		Sleep:   noSleep,
		Auth: AuthFunc(func(context.Context, *http.Request) error {
			return os.WriteFile(filePath, []byte("1234567890"), 0o600)
		}),
	}
	_, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{FieldName: "mediaFile", Path: filePath, MaxBytes: 4}},
	})
	if err == nil {
		t.Fatal("DoMultipart error = nil, want stream-time max-bytes rejection")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Fatalf("DoMultipart error = %q, want too large", err.Error())
	}
}

func TestRequesterDoMultipartRejectsChangedApprovedContentBeforeSend(t *testing.T) {
	filePath := t.TempDir() + "/media.txt"
	if err := os.WriteFile(filePath, []byte("evil"), 0o600); err != nil {
		t.Fatal(err)
	}
	approved := sha256.Sum256([]byte("safe"))
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	r := &Requester{BaseURL: server.URL, MaxRetries: 1, Sleep: noSleep}
	_, err := r.DoMultipart(context.Background(), http.MethodPost, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{
			FieldName:      "mediaFile",
			Path:           filePath,
			MaxBytes:       4,
			ExpectedSHA256: hex.EncodeToString(approved[:]),
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "changed since approval") {
		t.Fatalf("DoMultipart error = %v, want approved-content mismatch", err)
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("HTTP calls = %d, want zero before approved content is verified", got)
	}
}

func TestRequesterDoMultipartRejectsTooLargeFileBeforeSend(t *testing.T) {
	dir := t.TempDir()
	filePath := dir + "/payload.txt"
	if err := os.WriteFile(filePath, []byte("too large"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits++ }))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	_, err := r.DoMultipart(context.Background(), http.MethodPut, "/upload", nil, MultipartForm{
		Files: []MultipartFile{{FieldName: "mediaFile", Path: filePath, MaxBytes: 4}},
	})
	if err == nil {
		t.Fatalf("DoMultipart: want too-large error")
	}
	if hits != 0 {
		t.Fatalf("server hits = %d, want 0", hits)
	}
}

func TestRequesterDoFormNoBodySendsNoContentType(t *testing.T) {
	var sawContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawContentType = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	r := &Requester{BaseURL: srv.URL, Sleep: noSleep}
	if _, err := r.DoForm(context.Background(), http.MethodPost, "/x", nil, nil); err != nil {
		t.Fatalf("DoForm error = %v", err)
	}
	if sawContentType != "" {
		t.Fatalf("Content-Type = %q, want empty for bodyless form post", sawContentType)
	}
}

func TestParseRetryAfterSeconds(t *testing.T) {
	d, ok := parseRetryAfter("5")
	if !ok || d != 5*time.Second {
		t.Fatalf("parseRetryAfter(5) = %v, %v", d, ok)
	}
	if _, ok := parseRetryAfter(""); ok {
		t.Fatal("empty Retry-After should not parse")
	}
}

func TestParseRetryAfterAtPreservesProviderDate(t *testing.T) {
	now := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	reset := now.Add(90 * time.Second)
	delay, gotReset, ok := parseRetryAfterAt(reset.Format(http.TimeFormat), now)
	if !ok {
		t.Fatal("parseRetryAfterAt: want HTTP-date to parse")
	}
	if got, want := delay, 90*time.Second; got != want {
		t.Fatalf("delay = %v, want %v", got, want)
	}
	if !gotReset.Equal(reset) {
		t.Fatalf("reset = %v, want exact provider date %v", gotReset, reset)
	}
}

func TestRateLimitObservationParsesStandardBudgetHeaders(t *testing.T) {
	now := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	header := make(http.Header)
	header.Set("RateLimit-Limit", "100")
	header.Set("RateLimit-Remaining", "0")
	header.Set("RateLimit-Reset", "90")
	observation, ok := rateLimitObservation(http.StatusOK, header, 1, now, "")
	if !ok {
		t.Fatal("rateLimitObservation: want standard headers to be observed")
	}
	if got, want := observation.Source, RateLimitObservationSourceHeaders; got != want {
		t.Fatalf("source = %q, want %q", got, want)
	}
	if !observation.HasLimit || observation.Limit != 100 || !observation.HasRemaining || observation.Remaining != 0 {
		t.Fatalf("budget observation = %+v, want limit=100 remaining=0", observation)
	}
	if !observation.HasReset || !observation.ResetAt.Equal(now.Add(90*time.Second)) {
		t.Fatalf("reset observation = %+v, want %v", observation, now.Add(90*time.Second))
	}
}

func TestRateLimitObservationParsesDeclaredActualCost(t *testing.T) {
	now := time.Date(2026, time.August, 6, 12, 0, 0, 0, time.UTC)
	header := make(http.Header)
	header.Set("X-Actual-Cost", "2.5")
	observation, ok := rateLimitObservation(http.StatusOK, header, 1, now, "X-Actual-Cost")
	if !ok || !observation.HasCost || observation.Cost != 2.5 {
		t.Fatalf("actual-cost observation = %+v, want typed cost 2.5", observation)
	}
	if _, ok := rateLimitObservation(http.StatusOK, header, 1, now, "X-Other-Cost"); ok {
		t.Fatal("undeclared response header produced a rate-limit observation")
	}
}
