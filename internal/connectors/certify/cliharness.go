package certify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// cliRunFunc matches internal/cli.Run's signature. certify cannot import
// internal/cli directly: internal/cli/certify_cli.go wires `pm connectors
// certify` back into this package, and Go forbids the resulting import
// cycle (cli -> certify -> cli). Instead, cmd/pm/main.go — which already
// imports both packages — calls SetCLIRunFunc(cli.Run) once at process
// startup, and every in-process CLI invocation in this package (Harness.Run,
// and transitively every certify stage) goes through the cliRun package
// variable below. Tests that never call SetCLIRunFunc get a clear panic
// naming the missing wiring rather than a nil-pointer dereference.
type cliRunFunc func(args []string, stdout, stderr io.Writer) int

// CLIInvocationOptions carries invocation-local safety configuration across
// the certify -> cmd/pm -> internal/cli import-cycle boundary.
type CLIInvocationOptions struct {
	CrontabFile string
}

type cliRunContextFunc func(ctx context.Context, args []string, stdout, stderr io.Writer, opts CLIInvocationOptions) int

var (
	cliRun        cliRunFunc
	cliRunContext cliRunContextFunc
)

// SetCLIRunFunc registers the real internal/cli.Run entrypoint (or a test
// double) as this package's in-process CLI driver. Production callers
// (cmd/pm/main.go) call this exactly once, before any certify.Runner /
// certify.RunBatch / pm connectors certify invocation. Test packages that
// exercise Harness/Runner directly (this package's own *_test.go files) rely
// on TestMain having already called it once for the whole test binary.
func SetCLIRunFunc(run func(args []string, stdout, stderr io.Writer) int) {
	cliRun = run
}

// SetCLIRunContextFunc registers the context-aware in-process entrypoint.
// SetCLIRunFunc remains supported for compatibility with existing embedders.
func SetCLIRunContextFunc(run func(context.Context, []string, io.Writer, io.Writer, CLIInvocationOptions) int) {
	cliRunContext = run
}

// Harness drives the real CLI surface in-process (the cliRun function
// registered via SetCLIRunFunc) against an ephemeral project root, per
// certification design §A "Execution model": every stage is an in-process
// cli.Run([..., "--root", dir, "--json"], ...) call whose exit code and
// envelope kind are asserted.
type Harness struct {
	root    string
	ctx     context.Context
	secrets []string
}

// HarnessOption configures optional Harness behavior.
type HarnessOption func(*Harness)

// WithSecrets registers known secret values that must never appear verbatim
// in recorded argv, stdout, or stderr (THREAT-MODEL.md §1/§7). Values are
// used both for argv redaction and are available to callers via
// ScanForSecrets against captured output.
func WithContext(ctx context.Context) HarnessOption {
	return func(h *Harness) {
		if ctx != nil {
			h.ctx = ctx
		}
	}
}

func WithSecrets(secrets ...string) HarnessOption {
	return func(h *Harness) {
		h.secrets = append(h.secrets, secrets...)
	}
}

// NewHarness builds a Harness bound to root, an ephemeral project directory
// (e.g. os.MkdirTemp) that every Run call is isolated to.
func NewHarness(root string, opts ...HarnessOption) *Harness {
	h := &Harness{root: root, ctx: context.Background()}
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
// pass bare command args, executes the registered cliRun (see
// SetCLIRunFunc) in-process, captures stdout/stderr, parses the --json
// envelope when present, and computes a secret-redacted argv string for
// report recording. Panics with a descriptive message if SetCLIRunFunc was
// never called — this is a wiring bug, not a runtime condition callers
// should handle.
func (h *Harness) Run(args ...string) CLIResult {
	return h.RunWithOptions(CLIInvocationOptions{}, args...)
}

// RunWithOptions executes one context-aware invocation with safety options
// that belong to this invocation rather than process-global environment.
func (h *Harness) RunWithOptions(opts CLIInvocationOptions, args ...string) CLIResult {
	if cliRunContext == nil && cliRun == nil {
		panic("certify: CLI runner is nil — register the in-process CLI before driving the harness (see cmd/pm/main.go)")
	}
	full := h.withRoot(args)

	var stdout, stderr bytes.Buffer
	if err := h.ctx.Err(); err != nil {
		return CLIResult{ExitCode: 1, Stderr: err.Error(), ArgvRedacted: redactArgv(full, h.secrets)}
	}
	exitCode := 0
	if cliRunContext != nil {
		exitCode = cliRunContext(h.ctx, full, &stdout, &stderr, opts)
	} else if opts.CrontabFile != "" {
		exitCode = 1
		fmt.Fprint(&stderr, "certify: context-aware CLI runner required for crontab confinement")
	} else {
		exitCode = cliRun(full, &stdout, &stderr)
	}

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

// redactArgv joins args into a report-safe command line. Sensitive flag
// operands are redacted by meaning, even when their generated value was not
// known when the Harness was constructed (notably reverse approval tokens).
func redactArgv(args []string, secrets []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, "pm")
	redactNext := ""
	for _, arg := range args {
		if redactNext != "" {
			parts = append(parts, redactArgOperand(redactNext, arg, secrets))
			redactNext = ""
			continue
		}
		name, value, assigned := strings.Cut(arg, "=")
		if isSensitiveArgFlag(name) {
			if assigned {
				parts = append(parts, name+"=***")
			} else {
				parts = append(parts, name)
				redactNext = name
			}
			continue
		}
		if name == "--config" {
			if assigned {
				parts = append(parts, name+"="+redactConfigOperand(value, secrets))
			} else {
				parts = append(parts, name)
				redactNext = name
			}
			continue
		}
		parts = append(parts, redactSecretsInText(arg, secrets))
	}
	return strings.Join(parts, " ")
}

func redactArgOperand(flag, value string, secrets []string) string {
	if flag == "--config" {
		return redactConfigOperand(value, secrets)
	}
	return "***"
}

func redactConfigOperand(value string, secrets []string) string {
	field, _, assigned := strings.Cut(value, "=")
	if assigned && sensitiveFieldName(field) {
		return field + "=***"
	}
	return redactSecretsInText(value, secrets)
}

func isSensitiveArgFlag(name string) bool {
	switch name {
	case "--approve", "--value", "--token", "--password", "--private-key", "--client-secret":
		return true
	default:
		return false
	}
}

func redactSecretsInText(text string, secrets []string) string {
	for _, secret := range secrets {
		if strings.TrimSpace(secret) == "" {
			continue
		}
		for _, form := range secretForms(secret) {
			if form != "" {
				text = strings.ReplaceAll(text, form, "***")
			}
		}
	}
	return text
}

// SecretHit is opaque detection metadata. It deliberately contains no
// matched value or reversible encoding of that value.
type SecretHit struct {
	SecretIndex int `json:"secret_index"`
}

// ScanForSecrets reports opaque metadata for configured values found in text.
// Callers may persist counts/indexes, never the matched credential material.
func ScanForSecrets(text string, secrets []string) []SecretHit {
	var hits []SecretHit
	for i, secret := range secrets {
		if strings.TrimSpace(secret) == "" {
			continue
		}
		if containsSecretForm(text, secret) {
			hits = append(hits, SecretHit{SecretIndex: i})
		}
	}
	return hits
}

func containsSecretForm(text, secret string) bool {
	for _, form := range secretForms(secret) {
		if form != "" && strings.Contains(text, form) {
			return true
		}
	}
	return false
}

func secretForms(secret string) []string {
	return []string{
		secret,
		base64.StdEncoding.EncodeToString([]byte(secret)),
		base64.RawStdEncoding.EncodeToString([]byte(secret)),
		url.QueryEscape(secret),
		(&url.URL{Path: secret}).EscapedPath(),
		jsonEscapedForm(secret),
	}
}

// jsonEscapedForm renders secret the way encoding/json would render it as a
// JSON string VALUE (quotes/backslashes/control characters escaped per RFC
// 8259), with the wrapping quotes stripped — so a secret containing `"`,
// `\`, or control characters that survives verbatim inside a --json
// envelope's string value is still detected (m4, SECURITY-REVIEW.md).
func jsonEscapedForm(secret string) string {
	raw, err := json.Marshal(secret)
	if err != nil {
		return secret
	}
	// json.Marshal of a string always yields a quoted JSON string
	// ("..."); strip exactly the leading/trailing quote bytes it added.
	s := string(raw)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
