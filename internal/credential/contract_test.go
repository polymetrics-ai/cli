package credential

import (
	"crypto/sha256"
	"errors"
	"testing"
)

func TestNormalizeStdinRemovesOnlyOneDocumentedTerminalDelimiter(t *testing.T) {
	canary := "credential-transport-canary"
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "LF", input: canary + "\n", want: canary},
		{name: "CRLF", input: canary + "\r\n", want: canary},
		{name: "extra LF remains", input: canary + "\n\n", want: canary + "\n"},
		{name: "extra CRLF remains", input: canary + "\r\n\r\n", want: canary + "\r\n"},
		{name: "bare CR remains", input: canary + "\r", want: canary + "\r"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeStdin(tt.input)
			if gotLength, wantLength := len(got), len(tt.want); gotLength != wantLength {
				t.Fatalf("normalized length = %d, want %d", gotLength, wantLength)
			}
			if gotHash, wantHash := sha256.Sum256([]byte(got)), sha256.Sum256([]byte(tt.want)); gotHash != wantHash {
				t.Fatalf("normalized SHA-256 = %x, want %x", gotHash, wantHash)
			}
		})
	}
}

func TestCredentialValueRequirementsRejectOnlyTransportOnlyValues(t *testing.T) {
	checks := []struct {
		name    string
		require func(string, string) error
	}{
		{name: "persistent", require: RequirePersistentValue},
		{name: "authentication", require: RequireAuthenticationValue},
	}
	invalid := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "LF only", value: "\n"},
		{name: "CRLF only", value: "\r\n"},
	}
	accepted := []struct {
		name  string
		value string
	}{
		{name: "two LFs", value: "\n\n"},
		{name: "two CRLFs", value: "\r\n\r\n"},
		{name: "whitespace bytes", value: " \t "},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			for _, tt := range invalid {
				t.Run(tt.name, func(t *testing.T) {
					err := check.require("token", tt.value)
					var empty *EmptySecretError
					if !errors.As(err, &empty) {
						t.Fatalf("requirement error type = %T, want EmptySecretError", err)
					}
				})
			}
			for _, tt := range accepted {
				t.Run(tt.name, func(t *testing.T) {
					if err := check.require("token", tt.value); err != nil {
						t.Fatalf("requirement error = %v, want nil", err)
					}
				})
			}
		})
	}
}
