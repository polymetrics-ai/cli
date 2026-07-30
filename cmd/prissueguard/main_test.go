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
			name:       "no-mistakes delivery record passes",
			args:       []string{"--title", "ci: add dry-run Homebrew tap notification", "--body", noMistakesDeliveryBody()},
			wantCode:   0,
			wantStdout: "issueguard: ok (explicit no-mistakes delivery record)",
		},
		{
			name:       "valid PR with explicit issue wording passes",
			args:       []string{"--title", "feat(connectors): genericize repository read policies", "--body", "Implement the focused connector-boundary Issue B migration on branch refactor/connector-engine-policy-migration: remove GitHub-specific shared runtime policy names."},
			wantCode:   0,
			wantStdout: "issueguard: ok (explicit issue wording)",
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

func noMistakesDeliveryBody() string {
	return strings.Join([]string{
		"## Intent",
		"",
		"Implement the CLI-side least-privilege Homebrew tap notification.",
		"",
		"## What Changed",
		"",
		"- Added a least-privilege dry-run notification.",
		"",
		"## Testing",
		"",
		"Targeted validation passed.",
		"",
		"## Pipeline",
		"",
		"Updates from [git push no-mistakes](https://github.com/kunchenguid/no-mistakes)",
		"",
	}, "\n")
}
