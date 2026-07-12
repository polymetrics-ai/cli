package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHelpAndUnknownCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         []string
		expectedCode int
		stdout       string
		stderr       string
	}{
		{name: "bare command is help", expectedCode: 0, stdout: "loopctl safety"},
		{name: "help", args: []string{"--help"}, expectedCode: 0, stdout: "loopctl replay"},
		{name: "unknown", args: []string{"unknown"}, expectedCode: 64, stderr: "unknown command"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := run(tt.args, &stdout, &stderr)
			if code != tt.expectedCode {
				t.Fatalf("exit code = %d, want %d; stdout=%q stderr=%q", code, tt.expectedCode, stdout.String(), stderr.String())
			}
			if tt.stdout != "" && !strings.Contains(stdout.String(), tt.stdout) {
				t.Errorf("stdout = %q, want substring %q", stdout.String(), tt.stdout)
			}
			if tt.stderr != "" && !strings.Contains(stderr.String(), tt.stderr) {
				t.Errorf("stderr = %q, want substring %q", stderr.String(), tt.stderr)
			}
		})
	}
}

func TestRunSafetyCommands(t *testing.T) {
	t.Parallel()

	t.Run("status", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := run([]string{"safety", "status", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
		}
		var result struct {
			State         string `json:"state"`
			RunEnabled    bool   `json:"run_enabled"`
			ResumeEnabled bool   `json:"resume_enabled"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatalf("json.Unmarshal(): %v; output=%q", err, stdout.String())
		}
		if result.State != "closed" || result.RunEnabled || result.ResumeEnabled {
			t.Fatalf("status = %+v, want closed/disabled", result)
		}
	})

	t.Run("entrypoints", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := run([]string{"safety", "entrypoints", "--json"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
		}
		if !strings.Contains(stdout.String(), "scripts/claude-auto-loop.sh") || !strings.Contains(stdout.String(), "scripts/pi-auto-loop.sh") {
			t.Fatalf("entrypoint output = %q", stdout.String())
		}
	})

	t.Run("guard tracked driver", func(t *testing.T) {
		t.Parallel()
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		code := run([]string{"safety", "guard-driver", "scripts/claude-auto-loop.sh", "--json"}, &stdout, &stderr)
		if code != 78 {
			t.Fatalf("exit code = %d, want 78", code)
		}
		if stdout.Len() != 0 {
			t.Fatalf("stdout = %q, want empty", stdout.String())
		}
		if !strings.Contains(stderr.String(), "AUTO_LOOP_DISABLED_PHASE_0") {
			t.Fatalf("stderr = %q, want typed denial", stderr.String())
		}
	})
}

func TestRunReplay(t *testing.T) {
	t.Parallel()

	fixturePath := filepath.Join("..", "..", "internal", "agentloop", "testdata", "incidents", "dead_worker.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"replay", fixturePath, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"violation_code":"WORKER_LIVENESS_LOST"`) {
		t.Fatalf("stdout = %q, want replay result", stdout.String())
	}

	badPath := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(badPath, []byte(`{"unknown":true}`), 0o600); err != nil {
		t.Fatalf("os.WriteFile(): %v", err)
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"replay", badPath, "--json"}, &stdout, &stderr)
	if code != 65 {
		t.Fatalf("bad fixture exit code = %d, want 65; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "FIXTURE_UNKNOWN_FIELD") {
		t.Fatalf("bad fixture stderr = %q, want validation code", stderr.String())
	}
}
