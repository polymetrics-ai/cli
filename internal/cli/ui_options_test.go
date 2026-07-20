package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/flow"
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

func TestRunDashboardsActivateOnlyOnDualTTYAndBypassPlainPaths(t *testing.T) {
	stdinTTY := true
	stdoutTTY := true
	root := t.TempDir()
	initProject(t, root)
	flowDir := filepath.Join(root, ".polymetrics", "flows")
	if err := os.MkdirAll(flowDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeNativeFlowFixture(t, flowDir, "alpha")
	manifest := filepath.Join(flowDir, "alpha.json")

	var ttyOut, ttyErr bytes.Buffer
	ttyCode := RunWithOptions(
		[]string{"--root", root, "flow", "run", "--file", manifest, "--force"},
		&ttyOut,
		&ttyErr,
		RunOptions{Mode: ModeAuto, StdinIsTerminal: &stdinTTY, StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "xterm-256color"}},
	)
	if ttyCode != 0 {
		t.Fatalf("TTY flow dashboard code=%d stdout=%s stderr=%s", ttyCode, ttyOut.String(), ttyErr.String())
	}
	for _, want := range []string{"Flow alpha", "✓ score", "NORMAL · run", "ctrl+c cancel"} {
		if !strings.Contains(ttyOut.String(), want) {
			t.Fatalf("TTY flow dashboard missing %q:\n%s", want, ttyOut.String())
		}
	}

	plainOut, plainErr, plainCode := runAutoFlowDashboardCase(t, root, manifest, RunOptions{Mode: ModePlain}, "--force")
	if plainCode != 0 || plainErr != "" || plainOut != "Flow alpha: ok\n" {
		t.Fatalf("plain baseline code=%d stdout=%q stderr=%q", plainCode, plainOut, plainErr)
	}

	bypassCases := []struct {
		name string
		args []string
		opts RunOptions
	}{
		{name: "plain flag", args: []string{"--plain"}, opts: RunOptions{Mode: ModeAuto, StdinIsTerminal: &stdinTTY, StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "xterm-256color"}}},
		{name: "no input", args: []string{"--no-input"}, opts: RunOptions{Mode: ModeAuto, StdinIsTerminal: &stdinTTY, StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "xterm-256color"}}},
		{name: "ci", opts: RunOptions{Mode: ModeAuto, StdinIsTerminal: &stdinTTY, StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "xterm-256color", "CI": "1"}}},
		{name: "pm no tui", opts: RunOptions{Mode: ModeAuto, StdinIsTerminal: &stdinTTY, StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "xterm-256color", "PM_NO_TUI": "1"}}},
		{name: "term dumb", opts: RunOptions{Mode: ModeAuto, StdinIsTerminal: &stdinTTY, StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "dumb"}}},
		{name: "stdin piped", opts: RunOptions{Mode: ModeAuto, StdinIsTerminal: boolPtr(false), StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "xterm-256color"}}},
		{name: "stdout piped", opts: RunOptions{Mode: ModeAuto, StdinIsTerminal: &stdinTTY, StdoutIsTerminal: boolPtr(false), Env: map[string]string{"TERM": "xterm-256color"}}},
	}
	for _, tt := range bypassCases {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, code := runAutoFlowDashboardCase(t, root, manifest, tt.opts, append(tt.args, "--force")...)
			if code != plainCode || stdout != plainOut || stderr != plainErr {
				t.Fatalf("bypass differs from plain baseline\ncode got %d want %d\nstdout got %q want %q\nstderr got %q want %q", code, plainCode, stdout, plainOut, stderr, plainErr)
			}
			if containsANSI(stdout) || strings.Contains(stdout, "NORMAL · run") {
				t.Fatalf("bypass output used dashboard/ANSI: %q", stdout)
			}
		})
	}

	jsonOut, jsonErr, jsonCode := runAutoFlowDashboardCase(t, root, manifest, RunOptions{Mode: ModeAuto, StdinIsTerminal: &stdinTTY, StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "xterm-256color"}}, "--json", "--force")
	if jsonCode != 0 || jsonErr != "" || !strings.Contains(jsonOut, `"status": "ok"`) || strings.Contains(jsonOut, "NORMAL · run") || containsANSI(jsonOut) {
		t.Fatalf("json bypass failed code=%d stdout=%q stderr=%q", jsonCode, jsonOut, jsonErr)
	}
}

func TestFlowRunDashboardCancellationReachesEngineAndFinalFrame(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	app := &dashboardCancelFlowApp{started: make(chan struct{})}
	engine := &flow.Engine{
		Manifest: flow.FlowManifest{
			Version: 1,
			Name:    "cancel-dashboard",
			Steps: []flow.FlowStep{{
				ID:         "extract",
				Kind:       flow.KindSync,
				Connection: "sample",
				Streams:    []string{"customers"},
				Out:        []string{"customers"},
			}},
		},
		App:     app,
		LockDir: t.TempDir(),
	}
	manifest := engine.Manifest
	var stdout bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- flowRunDashboard(ctx, engine, manifest, &stdout, false)
	}()

	select {
	case <-app.started:
		cancel()
	case <-time.After(5 * time.Second):
		t.Fatal("dashboard flow did not start")
	}
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), context.Canceled.Error()) {
			t.Fatalf("flowRunDashboard error = %v, want cancellation", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("dashboard flow did not stop after cancellation")
	}
	for _, want := range []string{"Cancelled after extract", "Resume: pm flow run cancel-dashboard", "NORMAL · run"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("cancelled dashboard missing %q:\n%s", want, stdout.String())
		}
	}
}

type dashboardCancelFlowApp struct {
	started chan struct{}
}

func (a *dashboardCancelFlowApp) ETLRun(ctx context.Context, _ string, _ []string) (flow.ETLResult, error) {
	close(a.started)
	<-ctx.Done()
	return flow.ETLResult{}, ctx.Err()
}

func (*dashboardCancelFlowApp) QuerySQL(context.Context, string, int) ([]map[string]any, error) {
	return nil, nil
}

func (*dashboardCancelFlowApp) RLMRun(context.Context, flow.RLMRunRequest) (flow.RLMResult, error) {
	return flow.RLMResult{}, nil
}

func TestETLRunDashboardActivatesOnDualTTYAndNoInputBypasses(t *testing.T) {
	stdinTTY := true
	stdoutTTY := true
	root := setupETLProject(t, "full_refresh_overwrite")

	var stdout, stderr bytes.Buffer
	code := RunWithOptions(
		[]string{"--root", root, "etl", "run", "--connection", "sample-to-warehouse", "--stream", "customers", "--batch-size", "2"},
		&stdout,
		&stderr,
		RunOptions{Mode: ModeAuto, StdinIsTerminal: &stdinTTY, StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "xterm-256color"}},
	)
	if code != 0 {
		t.Fatalf("TTY ETL dashboard code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	for _, want := range []string{"ETL customers", "✓ customers", "3 read → 3 written", "NORMAL · run"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("TTY ETL dashboard missing %q:\n%s", want, stdout.String())
		}
	}

	var bypassOut, bypassErr bytes.Buffer
	bypassCode := RunWithOptions(
		[]string{"--no-input", "--root", root, "etl", "run", "--connection", "sample-to-warehouse", "--stream", "customers", "--batch-size", "2"},
		&bypassOut,
		&bypassErr,
		RunOptions{Mode: ModeAuto, StdinIsTerminal: &stdinTTY, StdoutIsTerminal: &stdoutTTY, Env: map[string]string{"TERM": "xterm-256color"}},
	)
	if bypassCode != 0 || !strings.HasPrefix(bypassOut.String(), "ETL run ") || !strings.Contains(bypassOut.String(), "completed: read=3 loaded=3 failed=0") {
		t.Fatalf("no-input ETL bypass failed code=%d stdout=%q stderr=%q", bypassCode, bypassOut.String(), bypassErr.String())
	}
	if containsANSI(bypassOut.String()) || strings.Contains(bypassOut.String(), "NORMAL · run") {
		t.Fatalf("no-input ETL bypass used dashboard/ANSI: %q", bypassOut.String())
	}
}

func runAutoFlowDashboardCase(t *testing.T, root, manifest string, opts RunOptions, extra ...string) (string, string, int) {
	t.Helper()
	args := []string{"--root", root, "flow", "run", "--file", manifest}
	args = append(args, extra...)
	var stdout, stderr bytes.Buffer
	code := RunWithOptions(args, &stdout, &stderr, opts)
	return stdout.String(), stderr.String(), code
}

func boolPtr(v bool) *bool { return &v }

func containsANSI(value string) bool { return strings.Contains(value, "\x1b[") }

func TestGlobalUIFlagsDocumentedInHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "root help", args: []string{"--help"}, want: []string{"--plain", "--no-input", "--progress ndjson", "Flow and ETL run dashboards", "3 validation error", "invalid UI/progress flag"}},
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
