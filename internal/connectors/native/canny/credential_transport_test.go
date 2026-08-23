package canny

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/credential"
)

func TestCheckPreservesCredentialFormBytes(t *testing.T) {
	base := " " + strings.Repeat("c", 8192) + " "
	tests := []struct {
		name      string
		delimiter string
		retained  string
		escaped   string
	}{
		{name: "two terminal LFs", delimiter: "\n\n", retained: "\n", escaped: "%0A"},
		{name: "two terminal CRLFs", delimiter: "\r\n\r\n", retained: "\r\n", escaped: "%0D%0A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := credential.NormalizeStdin(base + tt.delimiter)
			wantValue := base + tt.retained
			assertCannyFingerprint(t, "normalized credential", value, wantValue)

			var calls atomic.Int32
			var body string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				payload, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("ReadAll() error = %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				body = string(payload)
				w.WriteHeader(http.StatusNoContent)
			}))
			defer srv.Close()

			err := (Connector{Client: srv.Client()}).Check(context.Background(), connectors.RuntimeConfig{
				Config:  map[string]string{"base_url": srv.URL},
				Secrets: map[string]string{"api_key": value},
			})
			if err != nil {
				t.Fatalf("Check() error = %v", err)
			}
			if calls.Load() != 1 {
				t.Fatalf("provider calls = %d, want 1", calls.Load())
			}
			wantBody := url.Values{"apiKey": []string{wantValue}}.Encode()
			assertCannyFingerprint(t, "form body", body, wantBody)
			if !strings.Contains(body, tt.escaped) || !strings.Contains(body, "+") {
				t.Fatal("form body did not percent-encode retained delimiter and spaces")
			}
		})
	}
}

func TestCheckRejectsTransportOnlyAPIKeyBeforeProvider(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "LF only", value: "\n"},
		{name: "CRLF only", value: "\r\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				calls.Add(1)
			}))
			defer srv.Close()

			err := (Connector{Client: srv.Client()}).Check(context.Background(), connectors.RuntimeConfig{
				Config:  map[string]string{"base_url": srv.URL},
				Secrets: map[string]string{"api_key": tt.value},
			})
			var empty *credential.EmptySecretError
			if !errors.As(err, &empty) {
				t.Fatalf("error type = %T, want EmptySecretError", err)
			}
			if calls.Load() != 0 {
				t.Fatalf("provider calls = %d, want 0", calls.Load())
			}
		})
	}
}

func assertCannyFingerprint(t *testing.T, label, got, want string) {
	t.Helper()
	gotHash := sha256.Sum256([]byte(got))
	wantHash := sha256.Sum256([]byte(want))
	if len(got) != len(want) || gotHash != wantHash {
		t.Fatalf("%s length/hash = %d/%x, want %d/%x", label, len(got), gotHash, len(want), wantHash)
	}
}
