package certify_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

// newTestServer returns an httptest.Server that echoes a fixed JSON body and
// records the Authorization header it received (for sanitizer assertions).
func newTestServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestRecordingTransportWritesCassette proves a RecordingTransport wraps a
// real http.RoundTripper, forwards the request, and persists a cassette
// entry (method, path, status, body) to the given directory (certification
// design §E: "Tier-1 record ... capturing sanitized golden cassettes").
func TestRecordingTransportWritesCassette(t *testing.T) {
	srv := newTestServer(t, `{"id":"1","name":"Ada"}`)
	dir := t.TempDir()

	rec := certify.NewRecordingTransport(http.DefaultTransport, dir, "etl_full_refresh_append")

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/customers", nil)
	resp, err := rec.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"id":"1","name":"Ada"}` {
		t.Errorf("response body = %q, want passthrough of server body", string(body))
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}

	if err := rec.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(dir, "etl_full_refresh_append"))
	if err != nil {
		t.Fatalf("ReadDir(stage dir) error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(cassette files) = %d, want 1: %v", len(entries), entries)
	}
	if entries[0].Name() != "0001.json" {
		t.Errorf("cassette file name = %q, want 0001.json", entries[0].Name())
	}
}

// TestRecordingTransportSequenceIncrementsPerStage proves repeat requests
// within one stage get sequential cassette files, and a different stage
// starts its own sequence at 1 (certification design §E: "matched by
// (method, path, per-stage sequence)").
func TestRecordingTransportSequenceIncrementsPerStage(t *testing.T) {
	srv := newTestServer(t, `{"ok":true}`)
	dir := t.TempDir()

	rec := certify.NewRecordingTransport(http.DefaultTransport, dir, "catalog_live")
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/x", nil)
		resp, err := rec.RoundTrip(req)
		if err != nil {
			t.Fatalf("RoundTrip() error = %v", err)
		}
		_ = resp.Body.Close()
	}
	if err := rec.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(dir, "catalog_live"))
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("len(entries) = %d, want 3", len(entries))
	}
	wantNames := []string{"0001.json", "0002.json", "0003.json"}
	for i, want := range wantNames {
		if entries[i].Name() != want {
			t.Errorf("entries[%d].Name() = %q, want %q", i, entries[i].Name(), want)
		}
	}
}

// TestRecordingTransportSanitizesAuthorizationHeader proves the sanitizer
// strips Authorization/Cookie/X-Api-Key-class headers at record time
// (certification design §E: "Sanitizer runs at record time: ...
// Authorization/Cookie/X-Api-Key-class headers").
func TestRecordingTransportSanitizesAuthorizationHeader(t *testing.T) {
	srv := newTestServer(t, `{"ok":true}`)
	dir := t.TempDir()

	rec := certify.NewRecordingTransport(http.DefaultTransport, dir, "credentials_test")
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/whoami", nil)
	req.Header.Set("Authorization", "Bearer sk_live_supersecrettoken")
	req.Header.Set("Cookie", "session=abc123def456")
	req.Header.Set("X-Api-Key", "apikey-secret-value")
	req.Header.Set("Accept", "application/json")

	resp, err := rec.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	_ = resp.Body.Close()
	if err := rec.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "credentials_test", "0001.json"))
	if err != nil {
		t.Fatalf("ReadFile(cassette) error = %v", err)
	}
	text := string(raw)
	for _, secret := range []string{"sk_live_supersecrettoken", "abc123def456", "apikey-secret-value"} {
		if certifyContains(text, secret) {
			t.Errorf("cassette contains unsanitized secret %q: %s", secret, text)
		}
	}
	if !certifyContains(text, "Accept") {
		t.Errorf("cassette dropped non-secret header Accept: %s", text)
	}
}

// TestRecordingTransportSanitizesSecretValuesInBody proves planted secret
// VALUES (not just header names) are scrubbed from request/response bodies
// at record time, per certification design §E "secret values (exact,
// base64, URL-encoded)".
func TestRecordingTransportSanitizesSecretValuesInBody(t *testing.T) {
	srv := newTestServer(t, `{"token":"sk_test_bodysecret","id":"1"}`)
	dir := t.TempDir()

	rec := certify.NewRecordingTransport(http.DefaultTransport, dir, "stage_x", certify.WithRecordSecrets("sk_test_bodysecret"))
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/x", nil)
	resp, err := rec.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	_ = resp.Body.Close()
	if err := rec.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	raw, _ := os.ReadFile(filepath.Join(dir, "stage_x", "0001.json"))
	if certifyContains(string(raw), "sk_test_bodysecret") {
		t.Errorf("cassette leaked planted secret value: %s", string(raw))
	}
}

// TestRecordingTransportPreservesResponseBodyAfterSanitizing proves the
// caller-visible response body is not mutated by the sanitizer even when the
// live body carried a secret value that must not reach the cassette file.
func TestRecordingTransportPreservesResponseBodyAfterSanitizing(t *testing.T) {
	srv := newTestServer(t, `{"token":"sk_test_bodysecret"}`)
	dir := t.TempDir()

	rec := certify.NewRecordingTransport(http.DefaultTransport, dir, "stage_y", certify.WithRecordSecrets("sk_test_bodysecret"))
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/x", nil)
	resp, err := rec.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != `{"token":"sk_test_bodysecret"}` {
		t.Errorf("caller-visible response body was mutated: %q", string(body))
	}
}

func certifyContains(haystack, needle string) bool {
	return len(needle) > 0 && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
