package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunExitCodes(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{
			name:       "valid PR passes",
			args:       []string{"--title", "feat(agentic): add issue-first delivery system", "--body", "Closes #43"},
			wantCode:   0,
			wantStdout: "issueguard: ok (1 linked issue)",
		},
		{
			name:       "invalid PR is blocked",
			args:       []string{"--title", "add issue-first delivery system", "--body", "no issue"},
			wantCode:   1,
			wantStderr: "issueguard: blocked",
		},
		{
			name:       "missing body file is usage error",
			args:       []string{"--title", "feat(agentic): add issue-first delivery system", "--body-file", "missing.md"},
			wantCode:   2,
			wantStderr: "read PR body file:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			got := run(tt.args, &stdout, &stderr, func(string) string { return "" })
			if got != tt.wantCode {
				t.Fatalf("run() exit code = %d, want %d\nstdout: %s\nstderr: %s", got, tt.wantCode, stdout.String(), stderr.String())
			}
			if tt.wantStdout != "" && !strings.Contains(stdout.String(), tt.wantStdout) {
				t.Fatalf("stdout = %q, want substring %q", stdout.String(), tt.wantStdout)
			}
			if tt.wantStderr != "" && !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("stderr = %q, want substring %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestRunUsesFeatureManagerPlanningIssueFallback(t *testing.T) {
	planningRoot := t.TempDir()
	phaseDir := filepath.Join(planningRoot, "cli-pm-broker-profile-context-r1")
	if err := os.MkdirAll(phaseDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	plan := "Primary CLI issue: [#566](https://github.com/polymetrics-ai/cli/issues/566)\nParent PR: [#593](https://github.com/polymetrics-ai/cli/pull/593)\n"
	if err := os.WriteFile(filepath.Join(phaseDir, "PLAN.md"), []byte(plan), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	body := `## Intent

Add CLI-side PM Broker profile/context/domain foundation for Organization, Workspace, Environment, BrokerProfile, and runtime mode selection on branch fm/cli-pm-broker-profile-context-r1 targeting integration/pm-broker-production-program.

## What Changed

- Added metadata-only PM Broker context/domain support.
`
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	got := run(
		[]string{
			"--title", "feat(pmbroker): add metadata-only broker context foundation",
			"--body", body,
			"--head-ref", "fm/cli-pm-broker-profile-context-r1",
			"--planning-root", planningRoot,
		},
		&stdout,
		&stderr,
		func(string) string { return "" },
	)
	if got != 0 {
		t.Fatalf("run() exit code = %d, want 0\nstdout: %s\nstderr: %s", got, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "issueguard: ok (1 linked issue)") {
		t.Fatalf("stdout = %q, want linked issue success", stdout.String())
	}
}
