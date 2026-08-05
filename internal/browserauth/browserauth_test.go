package browserauth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestLoginRejectsInvalidCredentialShape(t *testing.T) {
	for _, credential := range []Credential{
		{},
		{OAuth: &OAuthCredential{AccessToken: "token"}, Session: &SessionCredential{Cookies: []Cookie{{Name: "session", Value: "value"}}}},
	} {
		_, err := Login(context.Background(), staticFlow{credential: credential})
		if err == nil || !strings.Contains(err.Error(), "exactly one") {
			t.Fatalf("Login() error = %v, want exactly-one-outcome rejection", err)
		}
	}
}

func TestCredentialNeedsReauthentication(t *testing.T) {
	now := time.Date(2026, time.August, 5, 12, 0, 0, 0, time.UTC)
	soon := now.Add(20 * time.Second)
	later := now.Add(31 * time.Second)

	for _, tc := range []struct {
		name string
		cred Credential
		want bool
	}{
		{
			name: "expired OAuth token",
			cred: Credential{OAuth: &OAuthCredential{ExpiresAt: soon}},
			want: true,
		},
		{
			name: "fresh OAuth token",
			cred: Credential{OAuth: &OAuthCredential{ExpiresAt: later}},
			want: false,
		},
		{
			name: "session nearing expiry",
			cred: Credential{Session: &SessionCredential{ExpiresHint: &soon}},
			want: true,
		},
		{
			name: "session expiry unknown",
			cred: Credential{Session: &SessionCredential{}},
			want: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cred.NeedsReauthentication(now); got != tc.want {
				t.Fatalf("NeedsReauthentication(%s) = %t, want %t", now, got, tc.want)
			}
		})
	}
}

func TestSafeOAuthErrorCode(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
		want string
	}{
		{name: "standard OAuth code", raw: "access_denied", want: "access_denied"},
		{name: "untrusted provider value", raw: "redaction-sentinel", want: "provider_error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SafeOAuthErrorCode(tc.raw); got != tc.want {
				t.Fatalf("SafeOAuthErrorCode(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestSafeOAuthTransportError(t *testing.T) {
	const marker = "redaction-sentinel"
	err := SafeOAuthTransportError(errors.New(marker))
	if err == nil {
		t.Fatal("SafeOAuthTransportError() error = nil, want safe error")
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatalf("SafeOAuthTransportError() exposed transport detail: %v", err)
	}
	if err.Error() != "oauth transport request failed" {
		t.Fatalf("SafeOAuthTransportError() error = %v, want generic transport error", err)
	}
}

type staticFlow struct {
	credential Credential
}

func (staticFlow) Name() string { return "test" }

func (f staticFlow) Login(context.Context) (Credential, error) { return f.credential, nil }
