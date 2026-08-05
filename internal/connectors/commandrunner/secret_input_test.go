package commandrunner

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestResolveSecretInputsUsesNonInlineSourcesAndNeverLeaksOnApplyFailure(t *testing.T) {
	const secret = "synthetic-secret-value-must-not-escape"

	request := SecretInputRequest{
		InputMode: "env_or_stdin",
		Fields: []SecretInputField{
			{Path: "body.token", Type: "string"},
		},
		FromEnv: []string{"body.token=PM_TEST_SECRET_INPUT"},
		LookupEnv: func(name string) (string, bool) {
			if name != "PM_TEST_SECRET_INPUT" {
				t.Fatalf("LookupEnv name = %q, want declared source name", name)
			}
			return secret, true
		},
		Apply: func(path, value string) error {
			if path != "body.token" || value != secret {
				t.Fatal("Apply did not receive the declared typed secret input")
			}
			return fmt.Errorf("provider rejected %s", value)
		},
	}

	if strings.Contains(strings.Join(request.FromEnv, ","), secret) {
		t.Fatal("secret reached the argv-derived source-reference representation")
	}
	err := ResolveSecretInputs(request)
	if err == nil {
		t.Fatal("ResolveSecretInputs error = nil, want sanitized apply failure")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("ResolveSecretInputs error leaked the secret")
	}
}

func TestResolveSecretInputsRejectsUnsafeSourcesWithoutReadingValues(t *testing.T) {
	const secret = "synthetic-secret-value-must-not-escape"

	tests := []struct {
		name    string
		request SecretInputRequest
	}{
		{
			name: "inline mode",
			request: SecretInputRequest{
				InputMode: "inline",
				Fields:    []SecretInputField{{Path: "body.password", Type: "string"}},
				FromEnv:   []string{"body.password=PM_TEST_SECRET_INPUT"},
				LookupEnv: func(string) (string, bool) { return secret, true },
			},
		},
		{
			name: "undeclared target",
			request: SecretInputRequest{
				InputMode: "env_or_stdin",
				Fields:    []SecretInputField{{Path: "body.password", Type: "string"}},
				FromEnv:   []string{"body.other=PM_TEST_SECRET_INPUT"},
				LookupEnv: func(string) (string, bool) { return secret, true },
			},
		},
		{
			name: "source lookup failure echoes value",
			request: SecretInputRequest{
				InputMode: "env_or_stdin",
				Fields:    []SecretInputField{{Path: "body.password", Type: "string"}},
				FromEnv:   []string{"body.password=PM_TEST_SECRET_INPUT"},
				LookupEnv: func(string) (string, bool) { return "", false },
				Apply:     func(string, string) error { return errors.New(secret) },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ResolveSecretInputs(tt.request)
			if err == nil {
				t.Fatal("ResolveSecretInputs error = nil, want rejection")
			}
			if strings.Contains(err.Error(), secret) {
				t.Fatal("ResolveSecretInputs error leaked the secret")
			}
		})
	}
}

func TestResolveSecretInputsOnlyAllowsDeclaredStringBodyTargets(t *testing.T) {
	called := false
	err := ResolveSecretInputs(SecretInputRequest{
		InputMode: "env_or_stdin",
		Fields: []SecretInputField{
			{Path: "config.password", Type: "string"},
		},
		FromEnv: []string{"config.password=PM_TEST_SECRET_INPUT"},
		LookupEnv: func(string) (string, bool) {
			return "synthetic-secret-value-must-not-escape", true
		},
		Apply: func(string, string) error {
			called = true
			return nil
		},
	})
	if err == nil {
		t.Fatal("ResolveSecretInputs error = nil, want non-body target rejection")
	}
	if called {
		t.Fatal("Apply was called for a non-body secret target")
	}
}

func TestResolveSecretInputsNeverLeaksAStdinReadFailure(t *testing.T) {
	const secret = "synthetic-secret-value-must-not-escape"
	err := ResolveSecretInputs(SecretInputRequest{
		InputMode:   "env_or_stdin",
		Fields:      []SecretInputField{{Path: "body.password", Type: "string", Required: true}},
		StdinFields: []string{"body.password"},
		Stdin:       failingSecretReader{err: errors.New("stdin contained " + secret)},
		Apply:       func(string, string) error { t.Fatal("Apply called after stdin failure"); return nil },
	})
	if err == nil {
		t.Fatal("ResolveSecretInputs error = nil, want stdin read failure")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatal("ResolveSecretInputs stdin error leaked the secret")
	}
}

type failingSecretReader struct{ err error }

func (r failingSecretReader) Read([]byte) (int, error) { return 0, r.err }

var _ io.Reader = failingSecretReader{}
