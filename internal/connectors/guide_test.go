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

func TestMechanismSectionRendersWebGovernanceMetadata(t *testing.T) {
	section := mechanismSection(Manifest{Metadata: Metadata{Mechanism: &MechanismSpec{
		Kind:                      MechanismWebSession,
		SanctionedByProvider:      false,
		OptInRequired:             true,
		UpstreamPin:               &MechanismUpstreamPin{Repo: "https://github.com/example/reference", SHA: "abc123", VerifiedAt: "2026-08-03"},
		BreakageReviewCadenceDays: 30,
		DisabledReason:            "upstream contract changed; awaiting review",
	}}})
	lines := strings.Join(section.Lines, "\n")
	for _, want := range []string{
		"UNOFFICIAL, experimental",
		"upstream_pin: https://github.com/example/reference@abc123 (verified_at: 2026-08-03)",
		"breakage_review_cadence_days: 30",
		"disabled_reason: upstream contract changed; awaiting review",
	} {
		if !strings.Contains(lines, want) {
			t.Fatalf("Mechanism section = %q, want %q", lines, want)
		}
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
