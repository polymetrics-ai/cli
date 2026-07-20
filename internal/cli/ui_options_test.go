package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWithOptionsPlainMatchesRun(t *testing.T) {
	args := []string{"version", "--json"}

	var runStdout, runStderr bytes.Buffer
	runCode := Run(args, &runStdout, &runStderr)

	var optsStdout, optsStderr bytes.Buffer
	optsCode := RunWithOptions(args, &optsStdout, &optsStderr, RunOptions{Mode: ModePlain})

	if optsCode != runCode || optsStdout.String() != runStdout.String() || optsStderr.String() != runStderr.String() {
		t.Fatalf("RunWithOptions plain mismatch\nRun code=%d stdout=%q stderr=%q\nRunWithOptions code=%d stdout=%q stderr=%q", runCode, runStdout.String(), runStderr.String(), optsCode, optsStdout.String(), optsStderr.String())
	}
}

func TestGlobalUIFlagsAreStrippedBeforeDispatch(t *testing.T) {
	stdoutTTY := true
	var stdout, stderr bytes.Buffer
	code := RunWithOptions(
		[]string{"--plain", "--no-input", "version", "--json"},
		&stdout,
		&stderr,
		RunOptions{Mode: ModeAuto, StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "xterm-256color"}},
	)
	if code != 0 {
		t.Fatalf("RunWithOptions with global UI flags code = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"kind": "Version"`) {
		t.Fatalf("stdout missing version envelope after stripping global UI flags:\n%s", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestInvocationEnvCapturesColorControls(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	t.Setenv("NO_COLOR", "0")
	t.Setenv("CLICOLOR", "0")

	got := invocationEnv(nil)
	for _, key := range []string{"TERM", "NO_COLOR", "CLICOLOR"} {
		if got[key] == "" {
			t.Fatalf("invocationEnv missing %s in %#v", key, got)
		}
	}
	if got["NO_COLOR"] != "0" || got["CLICOLOR"] != "0" {
		t.Fatalf("invocationEnv color controls = %#v, want NO_COLOR=0 and CLICOLOR=0", got)
	}
}

func TestProgressFlagRejectsUnsupportedValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--json", "--progress", "pretty", "version"}, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("Run invalid progress code = %d, want 3\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}

	var got struct {
		Kind  string `json:"kind"`
		Error struct {
			Category string `json:"category"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not JSON error: %v\n%s", err, stdout.String())
	}
	if got.Kind != "Error" || got.Error.Category != "validation" || !strings.Contains(got.Error.Message, "--progress") {
		t.Fatalf("unexpected error envelope: %+v\nstdout=%s", got, stdout.String())
	}
	if !strings.Contains(stderr.String(), "invalid --progress") {
		t.Fatalf("stderr = %q, want progress validation diagnostic", stderr.String())
	}
}

func TestProgressNDJSONWritesOnlyStderr(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)

	flowDir := t.TempDir()
	spec := `{
		"name": "lead-score",
		"features": [
			{"name": "email", "weight": 1.0, "score_if_set": 1.0}
		]
	}`
	if err := os.WriteFile(filepath.Join(flowDir, "lead-score.json"), []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	manifest := `{
		"version": 1,
		"name": "progress-flow",
		"steps": [
			{
				"id": "score",
				"kind": "rlm",
				"spec": "lead-score.json",
				"mode": "fixture",
				"in": [],
				"out": ["lead_scores"]
			}
		]
	}`
	manifestPath := filepath.Join(flowDir, "flow.json")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "--progress", "ndjson", "flow", "run", "--file", manifestPath}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run progress flow code = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}

	dec := json.NewDecoder(bytes.NewReader(stdout.Bytes()))
	var result map[string]any
	if err := dec.Decode(&result); err != nil {
		t.Fatalf("decode stdout result: %v\n%s", err, stdout.String())
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		t.Fatalf("stdout has more than one JSON value: err=%v extra=%v stdout=%q", err, extra, stdout.String())
	}
	if result["status"] != "ok" {
		t.Fatalf("flow result status = %v, want ok\nstdout=%s", result["status"], stdout.String())
	}
	if strings.Contains(stdout.String(), `"scope":"flow"`) || strings.Contains(stdout.String(), `"kind":"started"`) {
		t.Fatalf("stdout contains progress event data, want final envelope only:\n%s", stdout.String())
	}

	lines := strings.Split(strings.TrimSpace(stderr.String()), "\n")
	if len(lines) < 2 || strings.TrimSpace(stderr.String()) == "" {
		t.Fatalf("stderr progress lines = %d, want multiple NDJSON events\nstderr=%q", len(lines), stderr.String())
	}
	for i, line := range lines {
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("stderr line %d is not NDJSON: %v\nline=%q\nstderr=%s", i, err, line, stderr.String())
		}
		if event["kind"] == "" || event["scope"] == "" {
			t.Fatalf("stderr line %d missing kind/scope: %#v", i, event)
		}
	}
}

func TestProgressNDJSONFailureDocumentsMixedStderr(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)

	flowDir := t.TempDir()
	manifest := `{
		"version": 1,
		"name": "failing-progress-flow",
		"steps": [
			{
				"id": "score",
				"kind": "rlm",
				"spec": "missing-spec.json",
				"mode": "fixture",
				"in": [],
				"out": ["lead_scores"]
			}
		]
	}`
	manifestPath := filepath.Join(flowDir, "flow.json")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "--progress", "ndjson", "flow", "run", "--file", manifestPath}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("Run failing progress flow code = 0, want non-zero\nstdout=%s\nstderr=%s", stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"kind": "Error"`) {
		t.Fatalf("stdout missing final JSON error envelope:\n%s", stdout.String())
	}

	var sawNDJSON, sawDiagnostic bool
	for _, line := range strings.Split(strings.TrimSpace(stderr.String()), "\n") {
		if strings.HasPrefix(line, "error:") {
			sawDiagnostic = true
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("stderr line is neither NDJSON progress nor error diagnostic: %v\nline=%q\nstderr=%s", err, line, stderr.String())
		}
		if event["kind"] != "" && event["scope"] == "flow" {
			sawNDJSON = true
		}
	}
	if !sawNDJSON || !sawDiagnostic {
		t.Fatalf("stderr should be mixed on failure: sawNDJSON=%t sawDiagnostic=%t stderr=%q", sawNDJSON, sawDiagnostic, stderr.String())
	}
}

func TestGlobalUIFlagsDocumentedInHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "root help", args: []string{"--help"}, want: []string{"--plain", "--no-input", "--progress ndjson", "Future TTY renderers", "3 validation error", "invalid UI/progress flag"}},
		{name: "config help", args: []string{"help", "config"}, want: []string{"--plain", "--no-input", "--progress ndjson", "invalid UI/progress flag"}},
		{name: "etl help", args: []string{"etl", "--help"}, want: []string{"--progress ndjson", "stderr", "stderr may also include the final error diagnostic", "3 validation error", "invalid UI/progress flag"}},
		{name: "flow help", args: []string{"flow", "--help"}, want: []string{"--progress ndjson", "stderr", "stderr may also include the final error diagnostic", "3 validation error", "invalid UI/progress flag"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tt.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d\nstdout=%s\nstderr=%s", tt.args, code, stdout.String(), stderr.String())
			}
			for _, want := range tt.want {
				if !strings.Contains(stdout.String(), want) {
					t.Fatalf("help output missing %q:\n%s", want, stdout.String())
				}
			}
		})
	}
}
