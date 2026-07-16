package cli

import (
	"context"
	"io"
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

func findCobraCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
