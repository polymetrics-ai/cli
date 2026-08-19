// Package twenty keeps the Twenty CRM bundle's declared provider inventory
// aligned with the command runner's executable surface.
package twenty

import (
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

const twentyBundleName = "twenty"

// TestAllProviderOperationsHaveExecutableCommands is deliberately written
// before importing the recovered bundle. It pins the #277 REST inventory and
// proves every published command reaches the real runtime preflight rather
// than merely being listed in JSON.
func TestAllProviderOperationsHaveExecutableCommands(t *testing.T) {
	bundle, err := engine.Load(os.DirFS(".."), twentyBundleName)
	if err != nil {
		t.Fatalf("load %s bundle: %v", twentyBundleName, err)
	}
	if bundle.Surface == nil {
		t.Fatal("api_surface.json did not load")
	}
	if got := len(bundle.Surface.Endpoints); got != 168 {
		t.Fatalf("Twenty API-surface rows = %d, want 168", got)
	}

	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	surface := connector.CommandSurface()
	if surface == nil {
		t.Fatal("Twenty has no cli_surface.json")
	}
	if got := len(surface.Commands); got != 168 {
		t.Fatalf("Twenty CLI commands = %d, want 168", got)
	}

	wantIntents := map[string]int{
		"etl":          28,
		"direct_read":  28,
		"reverse_etl":  112,
	}
	gotIntents := make(map[string]int, len(wantIntents))
	for _, command := range surface.Commands {
		gotIntents[command.Intent]++
		if command.Availability != "implemented" {
			t.Errorf("%q availability = %q, want implemented", command.Path, command.Availability)
		}
		if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
			t.Errorf("Preflight(%q) = %v, want executable command", command.Path, err)
		}
	}
	for intent, want := range wantIntents {
		if got := gotIntents[intent]; got != want {
			t.Errorf("%s commands = %d, want %d", intent, got, want)
		}
	}
}
