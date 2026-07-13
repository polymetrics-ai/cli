package contract

import (
	"strings"
	"testing"
)

func TestDispatchRequiresFourFieldContract(t *testing.T) {
	t.Parallel()

	valid := Dispatch{
		Objective:    "Implement issue #375",
		OutputFormat: "Committed code plus a compact handoff",
		ToolGuidance: []string{"Use apply_patch", "Run focused tests"},
		Tools:        []Tool{ToolRead, ToolEdit, ToolTest},
		Boundaries:   []string{"Do not merge", "Do not print secrets"},
		WriteScope:   []string{"agent-runtime/shepherd/**"},
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid dispatch rejected: %v", err)
	}

	tests := []Dispatch{
		{OutputFormat: valid.OutputFormat, ToolGuidance: valid.ToolGuidance, Tools: valid.Tools, Boundaries: valid.Boundaries, WriteScope: valid.WriteScope},
		{Objective: valid.Objective, ToolGuidance: valid.ToolGuidance, Tools: valid.Tools, Boundaries: valid.Boundaries, WriteScope: valid.WriteScope},
		{Objective: valid.Objective, OutputFormat: valid.OutputFormat, Tools: valid.Tools, Boundaries: valid.Boundaries, WriteScope: valid.WriteScope},
		{Objective: valid.Objective, OutputFormat: valid.OutputFormat, ToolGuidance: valid.ToolGuidance, Tools: valid.Tools, WriteScope: valid.WriteScope},
	}
	for i, dispatch := range tests {
		if err := dispatch.Validate(); err == nil {
			t.Fatalf("case %d: expected missing contract field to fail", i)
		}
	}
}

func TestHandoffIsBounded(t *testing.T) {
	t.Parallel()

	handoff := strings.Repeat("line\n", 41)
	if err := ValidateHandoff(handoff); err == nil {
		t.Fatal("expected handoff over 40 lines to fail")
	}
}

func TestDispatchRejectsForbiddenTools(t *testing.T) {
	t.Parallel()

	dispatch := Dispatch{
		Objective:    "Publish changes",
		OutputFormat: "handoff",
		ToolGuidance: []string{"publish result"},
		Tools:        []Tool{Tool("github.merge")},
		Boundaries:   []string{"do not merge"},
		WriteScope:   []string{"docs/**"},
	}
	if err := dispatch.Validate(); err == nil {
		t.Fatal("expected direct merge guidance to fail")
	}
}
