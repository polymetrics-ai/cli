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

func TestRenderCommandSurfaceCommandRendersRepeatableAndTextExportFlags(t *testing.T) {
	line := renderCommandSurfaceCommand(CommandSurfaceCommand{
		Path:   "audit export",
		Intent: "text_export",
		Flags:  []CommandSurfaceFlag{{Name: "header-x-mode", Repeatable: true}},
	})
	if !strings.Contains(line, "--header-x-mode (repeatable)") {
		t.Fatalf("rendered command did not mark repeatable flag: %s", line)
	}
	for _, flag := range BinaryDownloadFlags() {
		if !strings.Contains(line, "--"+flag.Name) {
			t.Fatalf("text export guide did not render --%s: %s", flag.Name, line)
		}
	}
}

func TestConfigSectionRendersConditionalSecretRequirement(t *testing.T) {
	section := configSection(Manifest{SecretFields: []SecretField{{
		Name:         "password",
		RequiredWhen: "mode is not fixture",
	}}})
	if len(section.Lines) != 1 || section.Lines[0] != "password (secret) (required when mode is not fixture)" {
		t.Fatalf("configSection() = %#v, want conditional password requirement", section.Lines)
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
