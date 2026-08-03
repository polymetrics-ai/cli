package driver_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"polymetrics.ai/internal/browserauth/driver"
)

// TestSessionCapturesRealBrowserCookies proves the DoD claim "the package
// authenticates through a real browser and yields ... credential[s]" against
// an actual local Chrome/Chromium/Edge install — not a mock. It is opt-in
// (POLYMETRICS_BROWSER_INTEGRATION=1) and skipped by default, mirroring this
// repo's existing POLYMETRICS_INTEGRATION convention for runtime-backed
// tests (AGENTS.md "Verification"): CI and a plain `go test ./...` must not
// depend on a browser being installed, but the capability itself is real and
// exercised here, not merely asserted.
func TestSessionCapturesRealBrowserCookies(t *testing.T) {
	if os.Getenv("POLYMETRICS_BROWSER_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_BROWSER_INTEGRATION=1 to run against a real local browser")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "captured-by-real-browser", Path: "/"})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body>fake login page</body></html>"))
	}))
	defer server.Close()

	ctx := t.Context()
	session, err := driver.New(ctx, driver.Config{
		LoginURL:        server.URL,
		RequiredCookies: []string{"session_token"},
		Headless:        true,
		Timeout:         30 * time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = session.Close() }()

	// No explicit Navigate: New navigates to Config.LoginURL itself.
	if err := session.WaitFor(ctx, 10*time.Second, func(snap driver.Snapshot) bool {
		return snap.Cookies["session_token"] != ""
	}); err != nil {
		t.Fatalf("WaitFor() error = %v", err)
	}

	cookies, err := session.GetCookies(ctx)
	if err != nil {
		t.Fatalf("GetCookies() error = %v", err)
	}
	if len(cookies) != 1 || cookies[0].Name != "session_token" || cookies[0].Value != "captured-by-real-browser" {
		t.Fatalf("GetCookies() = %+v, want exactly session_token=captured-by-real-browser", cookies)
	}
}
