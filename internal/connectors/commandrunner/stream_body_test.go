package commandrunner

import (
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestStreamOverridesAdmitsDeclaredBodyFlags(t *testing.T) {
	command := connectors.CommandSurfaceCommand{
		Path:   "candidates list",
		Intent: "etl",
		Flags: []connectors.CommandSurfaceFlag{{
			Name: "created-after", Type: "string", MapsTo: "body.createdAfter",
		}},
	}
	_, values, err := streamOverrides(command, connectors.RuntimeConfig{}, map[string][]string{"created-after": {"2026-01-01"}})
	if err != nil {
		t.Fatalf("streamOverrides: %v", err)
	}
	if values["createdAfter"] != "2026-01-01" || len(values) != 1 {
		t.Fatalf("declared body values = %#v, want closed createdAfter binding", values)
	}
}
