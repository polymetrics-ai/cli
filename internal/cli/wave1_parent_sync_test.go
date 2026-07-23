package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestWave1ParentSyncPreservesGongAndCobraRouting(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantKind    string
		wantCommand string
		wantText    string
	}{
		{
			name:        "dynamic Gong help uses JSON envelope",
			args:        []string{"help", "gong", "--json"},
			wantKind:    "CommandManual",
			wantCommand: "gong",
			wantText:    "calls transcript",
		},
		{
			name:     "native connectors inspect keeps Gong manifest",
			args:     []string{"connectors", "inspect", "gong", "--json"},
			wantKind: "Connector",
			wantText: `"name": "gong"`,
		},
		{
			name:     "native version remains routable after dynamic connector",
			args:     []string{"version", "--json"},
			wantKind: "Version",
			wantText: `"version": "dev"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tt.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d, stderr = %s", tt.args, code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("Run(%v) stderr = %q, want empty", tt.args, stderr.String())
			}
			var envelope struct {
				Kind    string `json:"kind"`
				Command string `json:"command"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
				t.Fatalf("Run(%v) output is not one JSON envelope: %v\n%s", tt.args, err, stdout.String())
			}
			if envelope.Kind != tt.wantKind {
				t.Fatalf("Run(%v) kind = %q, want %q", tt.args, envelope.Kind, tt.wantKind)
			}
			if tt.wantCommand != "" && envelope.Command != tt.wantCommand {
				t.Fatalf("Run(%v) command = %q, want %q", tt.args, envelope.Command, tt.wantCommand)
			}
			if !strings.Contains(stdout.String(), tt.wantText) {
				t.Fatalf("Run(%v) output missing %q:\n%s", tt.args, tt.wantText, stdout.String())
			}
		})
	}
}
