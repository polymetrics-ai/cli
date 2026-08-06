package connectors

import (
	"strings"
	"testing"
)

func TestRenderCommandSurfaceCommandIncludesOperationMapping(t *testing.T) {
	line := renderCommandSurfaceCommand(CommandSurfaceCommand{
		Path:         "operations get_sources_by_target",
		Summary:      "Get sources by target",
		Intent:       "direct_read",
		Availability: "planned",
		Operation:    "zendesk-support.get_sources_by_target",
	})

	if !strings.Contains(line, "operation=zendesk-support.get_sources_by_target") {
		t.Fatalf("rendered command missing operation mapping: %s", line)
	}
}

func TestRenderCommandSurfaceFlagIncludesMapKey(t *testing.T) {
	line := renderCommandSurfaceFlag(CommandSurfaceFlag{
		Name:   "event-data-store-id",
		Type:   "string",
		MapsTo: "record.QueryParameterValues",
		MapKey: "$EventDataStoreId$",
	})

	if !strings.Contains(line, "map_key=$EventDataStoreId$") {
		t.Fatalf("rendered flag missing map key: %s", line)
	}
}

func TestEveryRegisteredConnectorHasGuideManualAndSkill(t *testing.T) {
	registry := NewRegistry()
	for _, meta := range registry.List() {
		connector, ok := registry.Get(meta.Name)
		if !ok {
			t.Fatalf("connector %s not found", meta.Name)
		}
		if err := ValidateConnectorGuide(connector); err != nil {
			t.Fatalf("ValidateConnectorGuide(%s) error = %v", meta.Name, err)
		}
		manual := RenderConnectorManual(connector)
		skill := RenderConnectorSkill(connector)
		if strings.Contains(manual, "{\n") {
			t.Fatalf("manual for %s should be human-readable, not raw JSON:\n%s", meta.Name, manual)
		}
		if strings.Contains(skill, "ghp_") || strings.Contains(skill, "secret-token") {
			t.Fatalf("skill for %s contains secret-like text:\n%s", meta.Name, skill)
		}
	}
}
