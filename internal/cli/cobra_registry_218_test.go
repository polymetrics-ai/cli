package cli

import (
	"bytes"
	"context"
	"github.com/spf13/cobra"
	"io"
	"polymetrics.ai/internal/connectors"
	"strings"
	"testing"
)

func TestCobraRegistryConstruction218(t *testing.T) {
	supplied := appRegistry()
	for _, tc := range []struct {
		name string
		want int
		run  func(cobraRegistrySource)
	}{
		{"legacy", 0, func(s cobraRegistrySource) {
			s.newLegacyCommand(context.Background(), "", io.Discard, false, cobraLegacyCommand{name: "version"})
		}},
		{"help", 0, func(s cobraRegistrySource) { s.setManualHelp(&cobra.Command{}, "", io.Discard, false) }},
		{"tree", 0, func(s cobraRegistrySource) {
			o := defaultAppOpeners()
			o.registry = s.registry
			o.registryFallback = s.fallback
			newRootCmd(context.Background(), testRouterConfig(t.TempDir(), false), io.Discard, io.Discard, o)
		}},
		{"standalone", 1, func(s cobraRegistrySource) {
			o := defaultAppOpeners()
			o.registryFallback = s.fallback
			newRootCmd(context.Background(), testRouterConfig(t.TempDir(), false), io.Discard, io.Discard, o)
		}},
		{"run", 0, func(s cobraRegistrySource) {
			o := defaultAppOpeners()
			o.registry = s.registry
			o.registryFallback = s.fallback
			if code := run([]string{"--root", t.TempDir(), "version"}, io.Discard, io.Discard, o); code != 0 {
				t.Fatalf("run exit=%d", code)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			count := 0
			source := cobraRegistrySource{registry: supplied, fallback: func() *connectors.Registry { count++; return appRegistry() }}
			tc.run(source)
			t.Logf("actual fallback constructor calls=%d expected=%d", count, tc.want)
			if count != tc.want {
				t.Fatalf("discarded registry constructors=%d want=%d", count, tc.want)
			}
		})
	}
}

func TestCobraPublicConsecutiveInvocations218(t *testing.T) {
	root := t.TempDir()
	for _, args := range [][]string{{"version"}, {"help"}, {"--json", "version"}, {"connectors", "--help"}} {
		var out, err bytes.Buffer
		if code := Run(append([]string{"--root", root}, args...), &out, &err); code != 0 {
			t.Fatalf("public invocation failed: code=%d", code)
		}
		if strings.TrimSpace(out.String()) == "" {
			t.Fatal("public invocation lost output")
		}
	}
}

func TestCobraSuppliedRegistryIdentity218(t *testing.T) {
	acquired := 0
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{Name: "oracle218", DisplayName: "Oracle218"}}, func(context.Context, string) (connectors.Connector, error) { acquired++; return nil, nil }, connectors.CommandSummary{Connector: "oracle218", Usage: "oracle218 inspect", Tagline: "literal218"})
	if err != nil {
		t.Fatal(err)
	}
	openers := defaultAppOpeners()
	openers.registry = registry
	var out bytes.Buffer
	root := newRootCmd(context.Background(), testRouterConfig(t.TempDir(), false), &out, io.Discard, openers)
	if err := executeRootCmd(root, []string{"help"}); err != nil {
		t.Fatal(err)
	}
	setManualHelp(root, "", &out, false, registry)
	if err := root.Help(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "oracle218") || acquired != 0 {
		t.Fatalf("supplied metadata lost or laziness broken; acquisitions=%d", acquired)
	}
}
