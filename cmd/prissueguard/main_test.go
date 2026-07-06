package main

import (
	"bytes"
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
