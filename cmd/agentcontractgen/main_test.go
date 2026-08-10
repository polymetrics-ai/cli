package main

import (
	"bytes"
	"context"
	"encoding/json"
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
			wantOutput: "Connector certification Shepherd gate",
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

func TestRunCertificationGateBlocksCurrentGitHubBaseline(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(context.Background(), []string{
		"certification-gate", "--root", root, "--connector", "github", "--transition", "integrate_sub_pr",
	}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("certification gate code = %d, want 1\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	var verdict struct {
		Decision string `json:"decision"`
		Failures []struct {
			ID string `json:"id"`
		} `json:"failures"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &verdict); err != nil {
		t.Fatalf("decode certification gate output: %v\n%s", err, stdout.String())
	}
	if verdict.Decision != "RETRY" {
		t.Fatalf("certification gate decision = %#v, want RETRY", verdict)
	}
	for _, failure := range verdict.Failures {
		if failure.ID == "capability/github/capability:check/live_evidence" {
			return
		}
	}
	t.Fatalf("certification gate failures = %#v, want GitHub live-evidence retry", verdict.Failures)
}
