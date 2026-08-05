package driver_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/browserauth"
	"polymetrics.ai/internal/browserauth/driver"
)

// TestFlowYieldsRealBrowserSessionCredential proves the DoD claim "the
// package authenticates through a real browser and yields ... credential[s]"
// against an actual local Chrome/Chromium/Edge install — not a mock. It is
// opt-in (POLYMETRICS_BROWSER_INTEGRATION=1) and skipped by default, mirroring
// this repo's existing POLYMETRICS_INTEGRATION convention for runtime-backed
// tests (AGENTS.md "Verification"): CI and a plain `go test ./...` must not
// depend on a browser being installed, but the public Flow outcome is real and
// exercised here, not merely asserted.
func TestFlowYieldsRealBrowserSessionCredential(t *testing.T) {
	if os.Getenv("POLYMETRICS_BROWSER_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_BROWSER_INTEGRATION=1 to run against a real local browser")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "captured-by-real-browser", Path: "/"})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body>fake login page</body></html>"))
	}))
	defer server.Close()
	origin := "https://" + strings.TrimPrefix(server.URL, "http://")

	ctx := t.Context()
	flow, err := driver.NewFlow(driver.FlowConfig{
		Browser: driver.Config{
			LoginURL:        server.URL,
			RequiredCookies: []string{"session_token"},
			Headless:        true,
			Timeout:         30 * time.Second,
		},
		Origin: origin,
	})
	if err != nil {
		t.Fatalf("NewFlow() error = %v", err)
	}

	cred, err := browserauth.Login(ctx, flow)
	if err != nil {
		t.Fatalf("browserauth.Login() error = %v", err)
	}
	if cred.OAuth != nil || cred.Session == nil {
		t.Fatalf("credential = %+v, want exactly a session credential", cred)
	}
	if len(cred.Session.Cookies) != 1 || cred.Session.Cookies[0].Name != "session_token" || cred.Session.Cookies[0].Value != "captured-by-real-browser" {
		t.Fatalf("captured cookies = %+v, want exactly session_token=captured-by-real-browser", cred.Session.Cookies)
	}
	if cred.Session.FingerprintRef == "" {
		t.Fatal("session credential is missing the capturing browser fingerprint")
	}
	if cred.Session.Origin != origin {
		t.Fatalf("session origin = %q, want %q", cred.Session.Origin, origin)
	}
}
