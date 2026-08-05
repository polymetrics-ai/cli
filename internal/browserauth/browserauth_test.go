package browserauth

import (
	"context"
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

type staticFlow struct {
	credential Credential
}

func (staticFlow) Name() string { return "test" }

func (f staticFlow) Login(context.Context) (Credential, error) { return f.credential, nil }
