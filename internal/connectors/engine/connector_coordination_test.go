package engine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestConnectorSourceBoundStreamDriftPrecedesAuthenticationAdmission(t *testing.T) {
	for _, test := range []struct {
		name string
		read func(*Connector, context.Context, connectors.ReadRequest, func(connectors.Record) error) error
	}{
		{name: "Read", read: (*Connector).Read},
		{name: "ReadWithOutcome", read: (*Connector).ReadWithOutcome},
	} {
		t.Run(test.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
			t.Cleanup(server.Close)
			bundle := Bundle{
				HTTP: HTTPBase{URL: server.URL, Pagination: &PaginationSpec{Type: "next_url", NextURLPath: "next_page.uri"}},
				Operations: []OperationSpec{{
					ID: "get_workspaces", Kind: "stream_etl",
					SourceOperation: &SourceOperationBinding{ID: "asana.rest.getWorkspaces", Method: http.MethodGet, Path: "/workspaces"},
					Composite:       &CompositeOperationSpec{Steps: []string{"stream:workspaces"}},
				}},
				Streams: []StreamSpec{{Name: "workspaces", Path: "/users", Records: RecordsSpec{Path: "data"}, SchemaRef: "schemas/workspaces.json"}},
			}
			admission := &capturingAuthenticationAdmission{}
			err := test.read(New(bundle, nil), context.Background(), connectors.ReadRequest{
				Stream: "workspaces", Config: connectors.RuntimeConfig{AuthenticationAdmission: admission},
			}, func(connectors.Record) error { return nil })
			if err == nil || !strings.Contains(err.Error(), "does not match locked source endpoint") {
				t.Fatalf("source-bound stream drift = %v, want structural refusal", err)
			}
			if admission.calls != 0 || requests != 0 {
				t.Fatalf("auth admissions/requester calls = %d/%d, want 0/0", admission.calls, requests)
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
