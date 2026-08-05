package driver

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod/lib/proto"

	"polymetrics.ai/internal/browserauth"
)

var _ browserauth.Flow = (*Flow)(nil)

func TestFlowLoginYieldsMinimumSessionCredential(t *testing.T) {
	flow, err := NewFlow(FlowConfig{
		Browser: Config{
			RequiredCookies: []string{"session", "csrf"},
			Timeout:         9 * time.Second,
		},
		Origin:          "https://provider.example",
		CSRFHeader:      "x-csrf-token",
		CSRFValueCookie: "csrf",
	})
	if err != nil {
		t.Fatalf("NewFlow() error = %v", err)
	}

	session := &stubSession{cookies: []browserauth.Cookie{
		{Name: "session", Value: "session-value"},
		{Name: "csrf", Value: "csrf-value"},
		{Name: "unrelated", Value: "must-not-leave-browser"},
	}}
	cred, err := flow.loginWithSession(context.Background(), session, Resolution{Version: "Chrome 131"})
	if err != nil {
		t.Fatalf("loginWithSession() error = %v", err)
	}
	if cred.OAuth != nil || cred.Session == nil {
		t.Fatalf("credential = %+v, want exactly a session credential", cred)
	}
	if got := cred.Session; got.Origin != "https://provider.example" || got.CSRFHeader != "x-csrf-token" || got.CSRFValue != "csrf-value" || got.FingerprintRef != "Chrome 131" || got.CapturedAt.IsZero() {
		t.Fatalf("session credential = %+v, want declared origin, CSRF metadata, fingerprint, and capture time", got)
	}
	if len(cred.Session.Cookies) != 2 || cred.Session.Cookies[0].Name != "session" || cred.Session.Cookies[1].Name != "csrf" {
		t.Fatalf("captured cookies = %+v, want only declared session and csrf cookies", cred.Session.Cookies)
	}
	if !session.closed {
		t.Fatal("controlled browser session was not closed after credential capture")
	}
	if session.waitTimeout != 9*time.Second {
		t.Fatalf("WaitFor timeout = %s, want configured browser timeout", session.waitTimeout)
	}
}

func TestFlowLoginClosesBrowserOnReadinessFailure(t *testing.T) {
	flow, err := NewFlow(FlowConfig{
		Browser: Config{RequiredCookies: []string{"session"}},
		Origin:  "https://provider.example",
	})
	if err != nil {
		t.Fatalf("NewFlow() error = %v", err)
	}

	session := &stubSession{waitErr: errors.New("not signed in")}
	_, err = flow.loginWithSession(context.Background(), session, Resolution{Version: "Chrome 131"})
	if err == nil || err.Error() != "not signed in" {
		t.Fatalf("loginWithSession() error = %v, want readiness error", err)
	}
	if !session.closed {
		t.Fatal("controlled browser session was not closed after readiness failure")
	}
}

func TestFlowLoginRejectsAmbiguousRequiredCookie(t *testing.T) {
	flow, err := NewFlow(FlowConfig{
		Browser: Config{RequiredCookies: []string{"session"}},
		Origin:  "https://provider.example",
	})
	if err != nil {
		t.Fatalf("NewFlow() error = %v", err)
	}

	session := &stubSession{cookies: []browserauth.Cookie{
		{Name: "session", Value: "first"},
		{Name: "session", Value: "second"},
	}}
	_, err = flow.loginWithSession(context.Background(), session, Resolution{Version: "Chrome 131"})
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("loginWithSession() error = %v, want ambiguous-cookie rejection", err)
	}
	if !session.closed {
		t.Fatal("controlled browser session was not closed after ambiguous cookie capture")
	}
}

func TestRodSessionQueriesConfiguredCookieOrigin(t *testing.T) {
	const origin = "https://provider.example"
	var gotURLs []string
	session := &rodSession{
		names:        map[string]bool{"session": true},
		cookieOrigin: origin,
		cookieReader: func(_ context.Context, urls []string) ([]*proto.NetworkCookie, error) {
			gotURLs = append([]string(nil), urls...)
			return []*proto.NetworkCookie{{Name: "session", Value: "value", Domain: "provider.example"}}, nil
		},
	}

	cookies, err := session.GetCookies(context.Background())
	if err != nil {
		t.Fatalf("GetCookies() error = %v", err)
	}
	if len(gotURLs) != 1 || gotURLs[0] != origin {
		t.Fatalf("GetCookies() query URLs = %v, want [%s]", gotURLs, origin)
	}
	if len(cookies) != 1 || cookies[0].Name != "session" {
		t.Fatalf("GetCookies() = %+v, want the required origin cookie", cookies)
	}
}

func TestNewFlowRejectsUnsafeOrIncompleteSessionMetadata(t *testing.T) {
	base := FlowConfig{
		Browser: Config{RequiredCookies: []string{"session", "csrf"}},
		Origin:  "https://provider.example",
	}

	for _, tc := range []struct {
		name string
		edit func(*FlowConfig)
	}{
		{
			name: "non-HTTPS origin",
			edit: func(cfg *FlowConfig) { cfg.Origin = "http://provider.example" },
		},
		{
			name: "origin path",
			edit: func(cfg *FlowConfig) { cfg.Origin = "https://provider.example/login" },
		},
		{
			name: "CSRF cookie not declared as a minimum cookie",
			edit: func(cfg *FlowConfig) {
				cfg.CSRFHeader = "x-csrf-token"
				cfg.CSRFValueCookie = "not-required"
			},
		},
		{
			name: "CSRF header without source cookie",
			edit: func(cfg *FlowConfig) { cfg.CSRFHeader = "x-csrf-token" },
		},
		{
			name: "duplicate required cookie",
			edit: func(cfg *FlowConfig) { cfg.Browser.RequiredCookies = []string{"session", "session"} },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base
			tc.edit(&cfg)
			if _, err := NewFlow(cfg); err == nil {
				t.Fatal("NewFlow() error = nil, want configuration rejection")
			}
		})
	}
}

type stubSession struct {
	cookies     []browserauth.Cookie
	waitErr     error
	waitTimeout time.Duration
	closed      bool
}

func (s *stubSession) Navigate(context.Context, string) error { return nil }

func (s *stubSession) WaitFor(_ context.Context, timeout time.Duration, ready func(Snapshot) bool) error {
	s.waitTimeout = timeout
	if s.waitErr != nil {
		return s.waitErr
	}
	cookies := make(map[string]string, len(s.cookies))
	for _, cookie := range s.cookies {
		cookies[cookie.Name] = cookie.Value
	}
	if !ready(Snapshot{Cookies: cookies}) {
		return errors.New("stub session did not become ready")
	}
	return nil
}

func (s *stubSession) GetCookies(context.Context) ([]browserauth.Cookie, error) {
	return append([]browserauth.Cookie(nil), s.cookies...), nil
}

func (s *stubSession) Close() error {
	s.closed = true
	return nil
}
