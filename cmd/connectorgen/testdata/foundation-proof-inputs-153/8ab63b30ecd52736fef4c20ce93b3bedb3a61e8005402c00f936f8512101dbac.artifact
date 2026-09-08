package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestConnectorAuthenticationAdmissionMarksOnlyDeclaredProviderFailures(t *testing.T) {
	for _, test := range []struct {
		name         string
		status       int
		rules        []ErrorRule
		wantVerified bool
	}{
		{name: "declared 401", status: http.StatusUnauthorized, rules: []ErrorRule{{Status: http.StatusUnauthorized, Hint: "credential is invalid"}}, wantVerified: true},
		{name: "declared auth class", status: http.StatusForbidden, rules: []ErrorRule{{Status: http.StatusForbidden, Class: "auth_failed"}}, wantVerified: true},
		{name: "generic 401 is not enough", status: http.StatusUnauthorized, wantVerified: false},
		{name: "unmatched declared 401", status: http.StatusUnauthorized, rules: []ErrorRule{{Status: http.StatusUnauthorized, MatchBody: "revoked"}}, wantVerified: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var sends int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				sends++
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(`{"error":"invalid"}`))
			}))
			t.Cleanup(server.Close)
			admission := &capturingAuthenticationAdmission{}
			connector := New(Bundle{Name: "auth-fixture", HTTP: HTTPBase{
				URL: server.URL, Check: &RequestSpec{Method: http.MethodGet, Path: "/check"}, ErrorMap: test.rules,
			}}, nil)
			err := connector.Check(context.Background(), connectors.RuntimeConfig{AuthenticationAdmission: admission})
			if err == nil {
				t.Fatal("provider refusal unexpectedly succeeded")
			}
			if admission.calls != 1 || sends != 1 {
				t.Fatalf("admission calls=%d provider sends=%d, want one each", admission.calls, sends)
			}
			if admission.verified != test.wantVerified {
				t.Fatalf("verified classification=%t, want %t for %v", admission.verified, test.wantVerified, err)
			}
		})
	}
}

type capturingAuthenticationAdmission struct {
	calls    int
	verified bool
}

func (a *capturingAuthenticationAdmission) Execute(ctx context.Context, operation func(context.Context) error) error {
	a.calls++
	err := operation(ctx)
	a.verified = connectors.IsVerifiedAuthenticationFailure(err)
	return err
}
