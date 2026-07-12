package connectors

import (
	"strings"
	"testing"
)

func TestCommandSurfaceRenderingNormalizesStructuredMetadata(t *testing.T) {
	flag := renderCommandSurfaceFlag(CommandSurfaceFlag{
		Name:    "credential",
		Type:    "string",
		Summary: "Use a saved Freshchat connector credential.",
		MapsTo:  "credential",
	})
	if want := "--credential (string): Use a saved Freshchat connector credential; maps_to=credential"; flag != want {
		t.Fatalf("renderCommandSurfaceFlag() = %q, want %q", flag, want)
	}

	cmd := renderCommandSurfaceCommand(CommandSurfaceCommand{
		Path:         "file upload",
		Summary:      "Upload a Freshchat file",
		Intent:       "direct_write",
		Availability: "unsupported_local",
		Notes:        "Typed file_upload operation metadata declares the endpoint; execution remains blocked.",
	})
	if strings.Contains(cmd, "availability=unsupported_local unsupported local workflow") {
		t.Fatalf("renderCommandSurfaceCommand() duplicated unsupported workflow metadata: %q", cmd)
	}
	if !strings.Contains(cmd, "[intent=direct_write availability=unsupported_local]") {
		t.Fatalf("renderCommandSurfaceCommand() missing structured availability metadata: %q", cmd)
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
