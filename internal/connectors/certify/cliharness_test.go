package certify_test

import (
	"encoding/base64"
	"net/url"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

// TestHarnessRunInitAssertsEnvelopeKindAndExit drives the real in-process CLI
// (internal/cli.Run) via the harness against an ephemeral --root, per
// certification-design §A execution model and SPEC §1.6 cliharness.go scope.
func TestHarnessRunInitAssertsEnvelopeKindAndExit(t *testing.T) {
	root := t.TempDir()
	h := certify.NewHarness(root)

	res := h.Run("init", "--root", root, "--json")

	if res.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0; stderr=%s", res.ExitCode, res.Stderr)
	}
	if res.Kind != "InitResult" {
		t.Fatalf("Kind = %q, want InitResult; stdout=%s", res.Kind, res.Stdout)
	}
	if res.Envelope == nil {
		t.Fatalf("Envelope = nil, want parsed JSON envelope; stdout=%s", res.Stdout)
	}
	if _, ok := res.Envelope["project_dir"]; !ok {
		t.Errorf("Envelope missing project_dir: %+v", res.Envelope)
	}

	if err := h.MustKind(res, "InitResult", 0); err != nil {
		t.Errorf("MustKind() error = %v, want nil for matching kind/exit", err)
	}
}

// TestHarnessInjectsRootAutomaticallyWhenOmitted lets a stage author write
// bare CLI args (per stages_source.go's future usage) and still get project
// isolation without repeating --root on every call.
func TestHarnessInjectsRootAutomaticallyWhenOmitted(t *testing.T) {
	root := t.TempDir()
	h := certify.NewHarness(root)

	res := h.Run("init", "--json")

	if res.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0; stderr=%s", res.ExitCode, res.Stderr)
	}
	if res.Kind != "InitResult" {
		t.Fatalf("Kind = %q, want InitResult", res.Kind)
	}
}

// TestHarnessNonJSONRunHasNilEnvelope covers text-mode invocations (the
// harness must not fabricate a kind for commands run without --json).
func TestHarnessNonJSONRunHasNilEnvelope(t *testing.T) {
	root := t.TempDir()
	h := certify.NewHarness(root)

	res := h.Run("init", "--root", root)

	if res.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0; stderr=%s", res.ExitCode, res.Stderr)
	}
	if res.Envelope != nil {
		t.Errorf("Envelope = %+v, want nil for non-JSON run", res.Envelope)
	}
	if res.Kind != "" {
		t.Errorf("Kind = %q, want empty for non-JSON run", res.Kind)
	}
	if !strings.Contains(res.Stdout, "Initialized Polymetrics project") {
		t.Errorf("Stdout = %q, want human-readable init message", res.Stdout)
	}
}

// TestHarnessCapturesNonZeroExit proves the harness surfaces failing CLI
// invocations (unknown command -> usage error) instead of swallowing them.
func TestHarnessCapturesNonZeroExit(t *testing.T) {
	root := t.TempDir()
	h := certify.NewHarness(root)

	res := h.Run("definitely-not-a-command", "--json")

	if res.ExitCode == 0 {
		t.Fatalf("ExitCode = 0, want non-zero for unknown command; stdout=%s", res.Stdout)
	}
	if res.Kind != "Error" {
		t.Fatalf("Kind = %q, want Error; stdout=%s", res.Kind, res.Stdout)
	}
}

// TestMustKindMismatchIsTypedFailure asserts an envelope-kind mismatch
// surfaces as a typed *certify.KindMismatchError rather than a bare error,
// so stage implementations can distinguish "wrong shape" from other faults.
func TestMustKindMismatchIsTypedFailure(t *testing.T) {
	root := t.TempDir()
	h := certify.NewHarness(root)

	res := h.Run("init", "--root", root, "--json")

	err := h.MustKind(res, "SomethingElse", 0)
	if err == nil {
		t.Fatalf("MustKind() error = nil, want kind-mismatch error")
	}

	var mismatch *certify.KindMismatchError
	if !asKindMismatch(err, &mismatch) {
		t.Fatalf("MustKind() error = %v (%T), want *certify.KindMismatchError", err, err)
	}
	if mismatch.Want != "SomethingElse" || mismatch.Got != "InitResult" {
		t.Errorf("mismatch = %+v, want Want=SomethingElse Got=InitResult", mismatch)
	}
}

// TestMustKindExitMismatchIsTypedFailure covers the exit-code half of the
// assertion independent of kind.
func TestMustKindExitMismatchIsTypedFailure(t *testing.T) {
	root := t.TempDir()
	h := certify.NewHarness(root)

	res := h.Run("init", "--root", root, "--json")

	err := h.MustKind(res, "InitResult", 7)
	if err == nil {
		t.Fatalf("MustKind() error = nil, want exit-mismatch error")
	}
	var mismatch *certify.KindMismatchError
	if !asKindMismatch(err, &mismatch) {
		t.Fatalf("MustKind() error = %v (%T), want *certify.KindMismatchError", err, err)
	}
	if mismatch.WantExit != 7 || mismatch.GotExit != 0 {
		t.Errorf("mismatch = %+v, want WantExit=7 GotExit=0", mismatch)
	}
}

func asKindMismatch(err error, target **certify.KindMismatchError) bool {
	if m, ok := err.(*certify.KindMismatchError); ok {
		*target = m
		return true
	}
	return false
}

// TestScanForSecretsDetectsExactBase64AndURLEncodedForms is the secret-scan
// contract required by SPEC §1.6 / THREAT-MODEL §1 / §7: certify scans
// captured text for planted secret values in exact, base64, and
// URL-encoded forms.
func TestScanForSecretsDetectsExactBase64AndURLEncodedForms(t *testing.T) {
	secret := "sk_test_topsecret12345"
	b64 := base64.StdEncoding.EncodeToString([]byte(secret))
	urlEnc := url.QueryEscape("token=" + secret + "&x=1")

	tests := []struct {
		name string
		text string
	}{
		{"exact", "leaked in stdout: " + secret + " end"},
		{"base64", "authorization: Basic " + b64},
		{"urlencoded", "redirect?state=" + urlEnc},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hits := certify.ScanForSecrets(tc.text, []string{secret})
			if len(hits) == 0 {
				t.Fatalf("ScanForSecrets(%q) = empty, want at least one hit for planted secret", tc.name)
			}
		})
	}
}

// TestScanForSecretsDetectsJSONEscapedForm is the m4 regression test
// (SECURITY-REVIEW.md): a secret containing JSON-special characters (a
// double quote and a backslash), rendered inside a --json envelope, is
// escaped per RFC 8259 — this escaped form must still be detected, not just
// the exact/base64/URL-encoded forms.
func TestScanForSecretsDetectsJSONEscapedForm(t *testing.T) {
	secret := `sk_test_weird"quote\slash`
	// encoding/json would render this secret (as a JSON string value) with
	// the quote and backslash backslash-escaped: sk_test_weird\"quote\\slash
	jsonEscaped := `sk_test_weird\"quote\\slash`

	text := `{"token":"` + jsonEscaped + `"}`
	hits := certify.ScanForSecrets(text, []string{secret})
	if len(hits) == 0 {
		t.Fatalf("ScanForSecrets(JSON-escaped form) = empty, want a hit for secret %q in text %q", secret, text)
	}
}

func TestScanForSecretsCleanTextHasNoHits(t *testing.T) {
	secret := "sk_test_topsecret12345"
	hits := certify.ScanForSecrets("nothing sensitive here, just records=42", []string{secret})
	if len(hits) != 0 {
		t.Errorf("ScanForSecrets(clean) = %v, want no hits", hits)
	}
}

func TestScanForSecretsIgnoresEmptySecrets(t *testing.T) {
	hits := certify.ScanForSecrets("some text with empty secret entries", []string{"", "   "})
	if len(hits) != 0 {
		t.Errorf("ScanForSecrets(empty secrets) = %v, want no hits", hits)
	}
}

// TestHarnessRunRedactsArgvSecrets proves stages[].cli.argv_redacted never
// contains secret values, even when a stage command line embeds one
// (e.g. --config token=...). The harness must redact known secret values in
// the recorded argv string.
func TestHarnessRunRedactsArgvSecrets(t *testing.T) {
	root := t.TempDir()
	h := certify.NewHarness(root, certify.WithSecrets("super-secret-token"))

	res := h.Run("credentials", "add", "cert-sample", "--connector", "sample", "--root", root,
		"--config", "token=super-secret-token", "--json")

	if strings.Contains(res.ArgvRedacted, "super-secret-token") {
		t.Fatalf("ArgvRedacted = %q, must not contain secret value", res.ArgvRedacted)
	}
	if !strings.Contains(res.ArgvRedacted, "credentials") {
		t.Errorf("ArgvRedacted = %q, want it to retain non-secret argv content", res.ArgvRedacted)
	}
	// The stdout/stderr and envelope must also never carry the raw secret
	// value (defense in depth alongside argv redaction).
	if hits := certify.ScanForSecrets(res.Stdout, []string{"super-secret-token"}); len(hits) != 0 {
		t.Errorf("Stdout leaked secret: hits=%v stdout=%s", hits, res.Stdout)
	}
}
