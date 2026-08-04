package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCheckAndRender(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantOutput string
	}{
		{
			name:       "check",
			args:       []string{"check", "--root", root},
			wantCode:   0,
			wantOutput: "canonical contract and registered projections are current",
		},
		{
			name:       "render connector",
			args:       []string{"render", "--root", root, "--role", "pm-connector-worker"},
			wantCode:   0,
			wantOutput: "source_policy_map",
		},
		{
			name:       "unknown role",
			args:       []string{"render", "--root", root, "--role", "orchestrator"},
			wantCode:   1,
			wantOutput: "unknown role",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			code := run(context.Background(), test.args, &stdout, &stderr)
			combined := stdout.String() + stderr.String()
			if code != test.wantCode {
				t.Fatalf("run() code = %d, want %d\n%s", code, test.wantCode, combined)
			}
			if !strings.Contains(combined, test.wantOutput) {
				t.Fatalf("run() output = %q, want substring %q", combined, test.wantOutput)
			}
		})
	}
}
