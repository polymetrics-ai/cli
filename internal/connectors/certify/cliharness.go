package certify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"polymetrics.ai/internal/cli"
)

// Harness drives the real CLI surface in-process (cli.Run) against an
// ephemeral project root, per certification design §A "Execution model":
// every stage is an in-process cli.Run([..., "--root", dir, "--json"], ...)
// call whose exit code and envelope kind are asserted.
type Harness struct {
	root    string
	secrets []string
}

// HarnessOption configures optional Harness behavior.
type HarnessOption func(*Harness)

// WithSecrets registers known secret values that must never appear verbatim
// in recorded argv, stdout, or stderr (THREAT-MODEL.md §1/§7). Values are
// used both for argv redaction and are available to callers via
// ScanForSecrets against captured output.
func WithSecrets(secrets ...string) HarnessOption {
	return func(h *Harness) {
		h.secrets = append(h.secrets, secrets...)
	}
}

// NewHarness builds a Harness bound to root, an ephemeral project directory
// (e.g. os.MkdirTemp) that every Run call is isolated to.
func NewHarness(root string, opts ...HarnessOption) *Harness {
	h := &Harness{root: root}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// CLIResult captures one in-process CLI invocation: its exit code, raw
// stdout/stderr, the parsed --json envelope (nil for text-mode runs), the
// envelope's "kind" field (empty when absent/non-JSON), and a redacted argv
// string suitable for embedding in a report's stages[].cli.argv_redacted.
type CLIResult struct {
	ExitCode     int
	Stdout       string
	Stderr       string
	Envelope     map[string]any
	Kind         string
	ArgvRedacted string
}

// Run injects --root (when the caller has not supplied one) so stages can
// pass bare command args, executes cli.Run in-process, captures
// stdout/stderr, parses the --json envelope when present, and computes a
// secret-redacted argv string for report recording.
func (h *Harness) Run(args ...string) CLIResult {
	full := h.withRoot(args)

	var stdout, stderr bytes.Buffer
	exitCode := cli.Run(full, &stdout, &stderr)

	res := CLIResult{
		ExitCode:     exitCode,
		Stdout:       stdout.String(),
		Stderr:       stderr.String(),
		ArgvRedacted: redactArgv(full, h.secrets),
	}

	if hasJSONFlag(full) {
		if env, kind, ok := parseEnvelope(res.Stdout); ok {
			res.Envelope = env
			res.Kind = kind
		}
	}

	return res
}

func (h *Harness) withRoot(args []string) []string {
	for _, a := range args {
		if a == "--root" || strings.HasPrefix(a, "--root=") {
			return args
		}
	}
	full := make([]string, 0, len(args)+2)
	full = append(full, args...)
	full = append(full, "--root", h.root)
	return full
}

func hasJSONFlag(args []string) bool {
	for _, a := range args {
		if a == "--json" {
			return true
		}
	}
	return false
}

// parseEnvelope decodes the last top-level JSON object found in stdout (the
// CLI always emits exactly one envelope per invocation, but tolerating
// leading log noise keeps this robust) and extracts its "kind" field.
func parseEnvelope(stdout string) (env map[string]any, kind string, ok bool) {
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" {
		return nil, "", false
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil, "", false
	}
	k, _ := parsed["kind"].(string)
	return parsed, k, true
}

// KindMismatchError is returned by MustKind when the observed CLIResult does
// not match the expected envelope kind and/or exit code. Callers can inspect
// Want/Got/WantExit/GotExit to build a typed stage failure rather than
// pattern-matching an error string.
type KindMismatchError struct {
	Want, Got         string
	WantExit, GotExit int
}

func (e *KindMismatchError) Error() string {
	return fmt.Sprintf("certify: cli result mismatch: kind got=%q want=%q, exit got=%d want=%d",
		e.Got, e.Want, e.GotExit, e.WantExit)
}

// MustKind asserts that res carries the expected envelope kind and exit
// code, returning a *KindMismatchError (a typed failure per SPEC §1.6 —
// "envelope kind mismatch -> typed failure") when either does not match.
func (h *Harness) MustKind(res CLIResult, kind string, exit int) error {
	if res.Kind != kind || res.ExitCode != exit {
		return &KindMismatchError{Want: kind, Got: res.Kind, WantExit: exit, GotExit: res.ExitCode}
	}
	return nil
}

// redactArgv joins args into a single "pm ..." command line with any known
// secret value replaced by "***", per THREAT-MODEL.md §7 ("argv recorded in
// reports is redacted").
func redactArgv(args []string, secrets []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, "pm")
	for _, a := range args {
		parts = append(parts, redactSecretsInText(a, secrets))
	}
	return strings.Join(parts, " ")
}

func redactSecretsInText(text string, secrets []string) string {
	for _, s := range secrets {
		if s == "" {
			continue
		}
		text = strings.ReplaceAll(text, s, "***")
	}
	return text
}

// ScanForSecrets reports which of the given secret values appear in text in
// exact, base64-encoded, or URL-encoded form (THREAT-MODEL.md §1/§7:
// "Certify harness scans ALL captured stdout/stderr + the report for
// planted secret values in exact, base64, and URL-encoded forms"). Empty or
// whitespace-only secrets are ignored. The returned slice names which
// secrets were found (deduplicated), not every match location.
func ScanForSecrets(text string, secrets []string) []string {
	var hits []string
	for _, secret := range secrets {
		trimmed := strings.TrimSpace(secret)
		if trimmed == "" {
			continue
		}
		if containsSecretForm(text, secret) {
			hits = append(hits, secret)
		}
	}
	return hits
}

func containsSecretForm(text, secret string) bool {
	if strings.Contains(text, secret) {
		return true
	}
	if strings.Contains(text, base64.StdEncoding.EncodeToString([]byte(secret))) {
		return true
	}
	if strings.Contains(text, base64.RawStdEncoding.EncodeToString([]byte(secret))) {
		return true
	}
	if strings.Contains(text, url.QueryEscape(secret)) {
		return true
	}
	if strings.Contains(text, (&url.URL{Path: secret}).EscapedPath()) {
		return true
	}
	return false
}
