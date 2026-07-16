package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCobraRouterShellBuildsFreshHiddenWrapperTree(t *testing.T) {
	first := newRootCmd(context.Background(), ".", io.Discard, io.Discard, false)
	second := newRootCmd(context.Background(), ".", io.Discard, io.Discard, false)
	if first == second {
		t.Fatal("newRootCmd returned the same command tree instance")
	}
	for _, root := range []*cobra.Command{first, second} {
		t.Run(root.CommandPath(), func(t *testing.T) {
			if !root.DisableFlagParsing {
				t.Fatal("root command must keep legacy global parsing and connector flag passthrough")
			}
			if !root.SilenceErrors || !root.SilenceUsage {
				t.Fatal("cobra errors/usages must be silenced so writeError remains the sole reporter")
			}
			for _, name := range []string{"help", "man", "connectors", "credentials", "catalog", "etl", "query", "reverse", "agent", "runtime", "flow", "perf", "docs", "skills", "version", "rlm", "schedule"} {
				if got := findCobraCommand(root, name); got == nil {
					t.Fatalf("missing top-level cobra wrapper %q", name)
				}
			}
			for _, name := range []string{"extract", "worker"} {
				got := findCobraCommand(root, name)
				if got == nil {
					t.Fatalf("missing hidden top-level cobra wrapper %q", name)
				}
				if !got.Hidden {
					t.Fatalf("%s wrapper must remain hidden", name)
				}
				if !got.DisableFlagParsing {
					t.Fatalf("%s wrapper must keep DisableFlagParsing", name)
				}
			}
		})
	}
}

func TestCobraRouterShellPreservesLegacyHelpInterceptionForFallback(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "unknown command help", args: []string{"nosuch", "--help", "--json"}, want: `"message": "help topic \"nosuch\" not found"`},
		{name: "dynamic connector help", args: []string{"github", "help", "--json"}, want: `"message": "help topic \"github\" not found"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tt.args, &stdout, &stderr)
			if code != 1 {
				t.Fatalf("Run(%v) code = %d, want 1; stdout=%s stderr=%s", tt.args, code, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), tt.want) {
				t.Fatalf("stdout missing %q:\n%s", tt.want, stdout.String())
			}
			if strings.Contains(stderr.String(), "unknown command") || strings.Contains(stderr.String(), "missing connector command path") {
				t.Fatalf("fallback help was routed as command execution: stderr=%s", stderr.String())
			}
		})
	}
}

func findCobraCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
