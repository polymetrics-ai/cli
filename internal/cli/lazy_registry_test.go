package cli

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestDynamicConnectorCommandsUseLazyMetadata(t *testing.T) {
	var loads atomic.Int32
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{
		Name:            "github",
		DisplayName:     "GitHub",
		IntegrationType: "api",
	}}, func(_ context.Context, name string) (connectors.Connector, error) {
		loads.Add(1)
		return engine.New(engine.Bundle{
			Name: name,
			Metadata: engine.Metadata{
				Name:            name,
				DisplayName:     "GitHub",
				IntegrationType: "api",
			},
			CLISurface: &engine.CLISurface{
				Usage:   "pm github <command>",
				Tagline: "Work with GitHub.",
			},
		}, nil), nil
	}, connectors.CommandSummary{Connector: "github", Usage: "pm github <command>", Tagline: "Work with GitHub."})
	if err != nil {
		t.Fatal(err)
	}

	section := dynamicConnectorCommandsSection(registry)
	if !strings.Contains(section, "pm github <command> - GitHub: Work with GitHub.") {
		t.Fatalf("dynamic connector command section = %q", section)
	}
	if got := loads.Load(); got != 0 {
		t.Fatalf("dynamicConnectorCommandsSection() resolved %d bundles, want 0", got)
	}
}
