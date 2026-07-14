package cli

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestConnectorHelpResolverSupportsArbitraryDepth(t *testing.T) {
	surface := &connectors.CommandSurface{
		Tagline: "Work with Acme safely.",
		Usage:   "pm acme <topic> <leaf> [flags]",
		Groups: []connectors.CommandSurfaceGroup{
			{ID: "work", Title: "Work", Commands: []string{"alpha"}},
		},
		GlobalFlags: []connectors.CommandSurfaceFlag{
			{Name: "json", Type: "boolean", Summary: "Write JSON output."},
		},
		Commands: []connectors.CommandSurfaceCommand{
			{Path: "alpha beta gamma", Summary: "Read gamma.", Intent: "etl", Availability: "implemented", Stream: "gamma"},
			{Path: "alpha delta", Summary: "Create delta.", Intent: "reverse_etl", Availability: "implemented", Write: "create_delta", Approval: "Requires plan, preview, explicit approval, then execute."},
			{Path: "alpha help", Summary: "A declared command whose leaf is help."},
		},
	}

	tests := []struct {
		name        string
		args        []string
		wantHandled bool
		wantCommand string
		wantText    []string
		wantErr     bool
	}{
		{
			name:        "namespace",
			wantHandled: true,
			wantCommand: "acme",
			wantText:    []string{"NAME", "pm acme", "Work", "alpha beta gamma", "Requires plan, preview, explicit approval, then execute."},
		},
		{
			name:        "deep prefix",
			args:        []string{"alpha", "beta"},
			wantHandled: true,
			wantCommand: "acme alpha beta",
			wantText:    []string{"pm acme alpha beta", "alpha beta gamma", "stream=gamma"},
		},
		{
			name:        "exact command executes",
			args:        []string{"alpha", "beta", "gamma"},
			wantHandled: false,
		},
		{
			name:        "declared help leaf executes",
			args:        []string{"alpha", "help"},
			wantHandled: false,
		},
		{
			name:        "declared help leaf has explicit help",
			args:        []string{"alpha", "help", "--help"},
			wantHandled: true,
			wantCommand: "acme alpha help",
			wantText:    []string{"A declared command whose leaf is help."},
		},
		{
			name:        "explicit leaf help",
			args:        []string{"alpha", "beta", "gamma", "--help"},
			wantHandled: true,
			wantCommand: "acme alpha beta gamma",
			wantText:    []string{"SYNOPSIS", "pm acme alpha beta gamma [flags]", "Read gamma.", "stream=gamma"},
		},
		{
			name:        "prefix help alias",
			args:        []string{"help", "alpha", "beta"},
			wantHandled: true,
			wantCommand: "acme alpha beta",
		},
		{
			name:    "unknown prefix",
			args:    []string{"alpha", "unknown", "--help"},
			wantErr: true,
		},
		{
			name:    "unknown action without help",
			args:    []string{"alpha", "unknown"},
			wantErr: true,
		},
		{
			name:    "prefix rejects execution flags",
			args:    []string{"alpha", "--limit", "10"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveConnectorHelp("acme", surface, tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveConnectorHelp() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got.handled != tt.wantHandled {
				t.Fatalf("handled = %v, want %v", got.handled, tt.wantHandled)
			}
			if got.command != tt.wantCommand {
				t.Fatalf("command = %q, want %q", got.command, tt.wantCommand)
			}
			for _, want := range tt.wantText {
				if !strings.Contains(got.manual, want) {
					t.Fatalf("manual missing %q:\n%s", want, got.manual)
				}
			}
		})
	}
}

func TestConnectorHelpPreservesDeclaredOrderAndLeafMetadata(t *testing.T) {
	surface := &connectors.CommandSurface{
		Tagline: "Ordered commands.",
		Usage:   "pm order <topic> <leaf>",
		Groups: []connectors.CommandSurfaceGroup{
			{ID: "second", Title: "Second Group", Commands: []string{"zeta"}},
			{ID: "first", Title: "First Group", Commands: []string{"alpha"}},
		},
		Commands: []connectors.CommandSurfaceCommand{
			{Path: "alpha list", Summary: "List alpha."},
			{Path: "zeta create", Summary: "Create zeta.", Operation: "zeta.create", OutputPolicy: "Redact secrets.", Risk: "external mutation", Approval: "Requires approval.", Examples: []string{"pm order zeta create --name example"}},
		},
	}

	namespace, err := renderConnectorHelpManual("order", surface, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Index(namespace, "Second Group") > strings.Index(namespace, "First Group") {
		t.Fatalf("namespace did not preserve group order:\n%s", namespace)
	}
	if strings.Index(namespace, "zeta create") > strings.Index(namespace, "alpha list") {
		t.Fatalf("namespace did not preserve grouped command order:\n%s", namespace)
	}

	leaf, err := renderConnectorHelpManual("order", surface, []string{"zeta", "create"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"operation=zeta.create", "OUTPUT POLICY", "Redact secrets.", "RISK", "external mutation", "APPROVAL", "Requires approval.", "EXAMPLES", "pm order zeta create --name example"} {
		if !strings.Contains(leaf, want) {
			t.Fatalf("leaf manual missing %q:\n%s", want, leaf)
		}
	}
}
