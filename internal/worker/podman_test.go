package worker

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	"polymetrics.ai/internal/rlm"
)

// fakeRunnerCmd returns a NewCmd that simulates the container: it writes
// out/output.ndjson + manifest.json into the JobDir and exits 0. With exitCode
// != 0 it writes nothing and exits with that code (and the given stderr).
func fakeRunnerCmd(rows int, exitCode int, stderr string) func(context.Context, rlm.AgentRequest, string) *exec.Cmd {
	return func(ctx context.Context, req rlm.AgentRequest, name string) *exec.Cmd {
		outDir := filepath.Join(req.JobDir, "out")
		_ = os.MkdirAll(outDir, 0o700)
		script := ""
		if exitCode == 0 {
			// Write rows + manifest, then exit 0.
			outFile := filepath.Join(outDir, "output.ndjson")
			manFile := filepath.Join(outDir, "manifest.json")
			body := ""
			for i := 0; i < rows; i++ {
				body += `{"_polymetrics_raw_id":"r` + itoa(i) + `","_rlm_score":0.5}` + "\n"
			}
			man, _ := json.Marshal(manifest{ExpectedCount: rows, RecordsRead: rows})
			script = "cat > '" + outFile + "' <<'EOF'\n" + body + "EOF\n" +
				"cat > '" + manFile + "' <<'EOF'\n" + string(man) + "\nEOF\n" +
				"exit 0"
		} else {
			script = "echo '" + stderr + "' >&2; exit " + itoa(exitCode)
		}
		return exec.CommandContext(ctx, "sh", "-c", script)
	}
}

func TestRunPodman_Success(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	acts := &PodmanActivities{
		NewCmd:            fakeRunnerCmd(3, 0, ""),
		Reap:              func(string) {},
		Heartbeat:         func(context.Context) {},
		HeartbeatInterval: time.Millisecond,
	}
	env.RegisterActivity(acts)

	jobDir := t.TempDir()
	val, err := env.ExecuteActivity(acts.RunPodman, rlm.AgentRequest{Fingerprint: "fp1", JobDir: jobDir})
	if err != nil {
		t.Fatalf("activity error: %v", err)
	}
	var res rlm.AgentResult
	if err := val.Get(&res); err != nil {
		t.Fatalf("get result: %v", err)
	}
	if res.RecordsScored != 3 {
		t.Fatalf("RecordsScored = %d, want 3", res.RecordsScored)
	}
}

func TestRunPodman_HeartbeatFiresWithoutStdout(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	var beats int32
	acts := &PodmanActivities{
		// A command that produces NO stdout for ~80ms then writes output.
		NewCmd: func(ctx context.Context, req rlm.AgentRequest, name string) *exec.Cmd {
			outDir := filepath.Join(req.JobDir, "out")
			_ = os.MkdirAll(outDir, 0o700)
			man, _ := json.Marshal(manifest{ExpectedCount: 0, RecordsRead: 0})
			script := "sleep 0.08; : > '" + filepath.Join(outDir, "output.ndjson") + "'; cat > '" +
				filepath.Join(outDir, "manifest.json") + "' <<'EOF'\n" + string(man) + "\nEOF\nexit 0"
			return exec.CommandContext(ctx, "sh", "-c", script)
		},
		Reap:              func(string) {},
		Heartbeat:         func(context.Context) { atomic.AddInt32(&beats, 1) },
		HeartbeatInterval: 10 * time.Millisecond,
	}
	env.RegisterActivity(acts)

	jobDir := t.TempDir()
	if _, err := env.ExecuteActivity(acts.RunPodman, rlm.AgentRequest{Fingerprint: "fp2", JobDir: jobDir}); err != nil {
		t.Fatalf("activity error: %v", err)
	}
	if atomic.LoadInt32(&beats) == 0 {
		t.Fatal("wall-clock heartbeat never fired during a silent run")
	}
}

func TestRunPodman_ReflectionExhaustedNonRetryable(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()
	acts := &PodmanActivities{
		NewCmd:            fakeRunnerCmd(0, 3, "could not satisfy validation"),
		Reap:              func(string) {},
		Heartbeat:         func(context.Context) {},
		HeartbeatInterval: time.Millisecond,
	}
	env.RegisterActivity(acts)

	jobDir := t.TempDir()
	_, err := env.ExecuteActivity(acts.RunPodman, rlm.AgentRequest{Fingerprint: "fp3", JobDir: jobDir})
	if err == nil {
		t.Fatal("want error for exit 3")
	}
}

func TestClassifyExit(t *testing.T) {
	cases := []struct {
		code     int
		stderr   string
		wantType string // "" => retryable (plain error, not an ApplicationError type)
	}{
		{125, "Error: no such image foo", errBadAnalysisRequest},
		{137, "killed", errOOMKilled},
		{3, "validation exhausted", errReflectionExhausted},
		{4, "connection refused", errLLMUnreachable},
		{1, "some transient error", ""},
	}
	for _, tc := range cases {
		cmd := exec.Command("sh", "-c", "exit "+itoa(tc.code))
		runErr := cmd.Run()
		got := classifyExit(runErr, tc.stderr)
		if got == nil {
			t.Fatalf("exit %d: want error", tc.code)
		}
		var appErr *temporal.ApplicationError
		isApp := errors.As(got, &appErr)
		if tc.wantType == "" {
			if isApp && appErr.NonRetryable() {
				t.Errorf("exit %d should be retryable, got non-retryable %v", tc.code, got)
			}
			continue
		}
		if !isApp {
			t.Errorf("exit %d: want ApplicationError, got %T", tc.code, got)
			continue
		}
		if appErr.Type() != tc.wantType {
			t.Errorf("exit %d: type = %q, want %q", tc.code, appErr.Type(), tc.wantType)
		}
		if !appErr.NonRetryable() {
			t.Errorf("exit %d: %q should be non-retryable", tc.code, tc.wantType)
		}
	}
}
