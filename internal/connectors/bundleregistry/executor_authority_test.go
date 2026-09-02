package bundleregistry

import (
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestEveryImplementedCommandHasProductionRuntimeSurface(t *testing.T) {
	bundles, err := engine.LoadAll(defs.FS)
	if err != nil {
		t.Fatalf("LoadAll(defs): %v", err)
	}
	registry := New()

	for _, bundle := range bundles {
		if bundle.CLISurface == nil {
			continue
		}
		declared := make(map[string]string, len(bundle.CLISurface.Commands))
		for _, command := range bundle.CLISurface.Commands {
			if command.Availability == "implemented" {
				declared[command.Path] = command.Availability
			}
		}
		if len(declared) == 0 {
			continue
		}

		connector, ok := registry.Get(bundle.Name)
		if !ok {
			t.Errorf("implemented commands for %q have no production connector", bundle.Name)
			continue
		}
		surfaceProvider, ok := connector.(connectors.CommandSurfaceProvider)
		if !ok || surfaceProvider.CommandSurface() == nil {
			t.Errorf("implemented commands for %q have no production command surface: %T", bundle.Name, connector)
			continue
		}
		actual := make(map[string]string, len(surfaceProvider.CommandSurface().Commands))
		for _, command := range surfaceProvider.CommandSurface().Commands {
			actual[command.Path] = command.Availability
		}
		for path, availability := range declared {
			if got, ok := actual[path]; !ok || got != availability {
				t.Errorf("implemented command %q for %q has production availability %q, want %q", path, bundle.Name, got, availability)
			}
		}
	}
}

func TestRegistryRejectsDuplicateConnectorNames(t *testing.T) {
	registry := connectors.NewEmptyRegistry()
	if err := registry.Register(connectors.Sample{}); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if err := registry.Register(connectors.Sample{}); err == nil {
		t.Fatal("Register() silently accepted a duplicate connector name")
	}
}
