package connectors

import (
	"strings"
	"testing"
)

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
