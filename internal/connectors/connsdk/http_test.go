package connsdk

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func noSleep(_ context.Context, _ time.Duration) error { return nil }

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
	if strings.Contains(httpErr.Error(), "nope") {
		t.Fatalf("HTTPError surfaced response body")
	}
}

func TestHTTPErrorErrorRedactsURLQueryAndBody(t *testing.T) {
	err := (&HTTPError{Status: http.StatusUnauthorized, URL: "https://user:secret-token@api.example.test/items?api_key=secret-token#frag", Body: `{"error":"secret-token denied"}`}).Error()
	for _, leaked := range []string{"secret-token", "api_key=", "denied", "user:", "#frag"} {
		if strings.Contains(err, leaked) {
			t.Fatalf("HTTPError leaked %q in %q", leaked, err)
		}
	}
	if strings.Contains(err, "{\"error\"") {
		t.Fatalf("HTTPError surfaced remote response body: %q", err)
	}
	if !strings.Contains(err, "http 401") || !strings.Contains(err, "https://api.example.test/items") || !strings.Contains(err, "Unauthorized") {
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
		defer f.Close()
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
