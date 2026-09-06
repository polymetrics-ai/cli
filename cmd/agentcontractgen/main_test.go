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
		{name: "check", args: []string{"check", "--root", root}, wantCode: 0, wantOutput: "canonical contract and registered projections are current"},
		{name: "render connector", args: []string{"render", "--root", root, "--role", "pm-connector-worker"}, wantCode: 0, wantOutput: "Connector overlay"},
		{name: "unknown role", args: []string{"render", "--root", root, "--role", "orchestrator"}, wantCode: 1, wantOutput: "unknown role"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), test.args, &stdout, &stderr)
			combined := stdout.String() + stderr.String()
			if code != test.wantCode || !strings.Contains(combined, test.wantOutput) {
				t.Fatalf("run() code/output = %d/%q, want %d containing %q", code, combined, test.wantCode, test.wantOutput)
			}
		})
	}
}

func TestRemovedCertificationGateIsUnknown(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"certification-gate"}, &stdout, &stderr); code != 2 {
		t.Fatalf("removed certification-gate code = %d, want usage error", code)
	}
}
