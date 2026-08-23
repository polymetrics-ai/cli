package safetyculture

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/credential"
)

func TestRequesterPreservesHeaderSafeCredentialBytes(t *testing.T) {
	token := " " + strings.Repeat("s", 8192) + " "
	requester, err := (Connector{}).requester(connectors.RuntimeConfig{
		Config:  map[string]string{"base_url": "https://example.invalid"},
		Secrets: map[string]string{"access_token": token},
	})
	if err != nil {
		t.Fatalf("requester() error = %v", err)
	}
	req, err := http.NewRequest(http.MethodGet, "https://example.invalid/audits", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if err := requester.Auth.Apply(context.Background(), req); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	assertSafetyFingerprint(t, "Authorization", req.Header.Get("Authorization"), "Bearer "+token)
}

func TestCheckRejectsInvalidCredentialBeforeProvider(t *testing.T) {
	marker := "safetyculture-" + strings.Repeat("x", 256)
	tests := []struct {
		name        string
		value       string
		wantEmpty   bool
		containsRaw bool
	}{
		{name: "empty", value: "", wantEmpty: true},
		{name: "LF only", value: "\n", wantEmpty: true},
		{name: "CRLF only", value: "\r\n", wantEmpty: true},
		{name: "retained LF after stdin delimiter", value: credential.NormalizeStdin(marker + "\n\n"), containsRaw: true},
		{name: "retained CRLF after stdin delimiter", value: credential.NormalizeStdin(marker + "\r\n\r\n"), containsRaw: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				calls.Add(1)
			}))
			defer srv.Close()

			err := (Connector{Client: srv.Client()}).Check(context.Background(), connectors.RuntimeConfig{
				Config:  map[string]string{"base_url": srv.URL},
				Secrets: map[string]string{"access_token": tt.value},
			})
			if err == nil {
				t.Fatal("Check() error = nil")
			}
			if calls.Load() != 0 {
				t.Fatalf("provider calls = %d, want 0", calls.Load())
			}
			var empty *credential.EmptySecretError
			var invalid *credential.InvalidSecretValueError
			if tt.wantEmpty {
				if !errors.As(err, &empty) {
					t.Fatalf("error type = %T, want EmptySecretError", err)
				}
			} else if !errors.As(err, &invalid) {
				t.Fatalf("error type = %T, want InvalidSecretValueError", err)
			}
			if tt.containsRaw && strings.Contains(err.Error(), marker) {
				t.Fatal("credential diagnostic exposed the canary")
			}
		})
	}
}

func assertSafetyFingerprint(t *testing.T, label, got, want string) {
	t.Helper()
	gotHash := sha256.Sum256([]byte(got))
	wantHash := sha256.Sum256([]byte(want))
	if len(got) != len(want) || gotHash != wantHash {
		t.Fatalf("%s length/hash = %d/%x, want %d/%x", label, len(got), gotHash, len(want), wantHash)
	}
}
