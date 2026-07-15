package supervisor

import (
	"testing"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
)

func TestDecideDispatchesOnlyCanonicalUnit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		snapshot gsd.WorkflowSnapshot
		command  string
		unit     string
	}{
		{name: "planning", snapshot: gsd.WorkflowSnapshot{Next: gsd.NextDispatch{Action: "dispatch", UnitType: "plan-milestone", UnitID: "M001"}}, command: "plan-milestone", unit: "plan-milestone/M001"},
		{name: "discussion maps to targeted discuss", snapshot: gsd.WorkflowSnapshot{MilestoneID: "M001", Next: gsd.NextDispatch{Action: "dispatch", UnitType: "discuss-milestone", UnitID: "M001"}}, command: "discuss", unit: "discuss-milestone/M001"},
		{name: "execution", snapshot: gsd.WorkflowSnapshot{Next: gsd.NextDispatch{Action: "dispatch", UnitType: "execute-task", UnitID: "M001/S01/T01"}}, command: "execute-task", unit: "execute-task/M001/S01/T01"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			decision, err := DecideWithRegistry(tc.snapshot, testSupervisorRegistry())
			if err != nil {
				t.Fatalf("decide: %v", err)
			}
			if decision.Kind != DecisionDispatch || decision.Command != tc.command || decision.Unit != tc.unit {
				t.Fatalf("decision=%+v", decision)
			}
		})
	}
}

func TestDecideRejectsGenericOrUnknownDispatch(t *testing.T) {
	t.Parallel()
	for _, unitType := range []string{"auto", "new-milestone", "resume", ""} {
		unitType := unitType
		t.Run(unitType, func(t *testing.T) {
			t.Parallel()
			if _, err := DecideWithRegistry(gsd.WorkflowSnapshot{Next: gsd.NextDispatch{Action: "dispatch", UnitType: unitType}}, testSupervisorRegistry()); err == nil {
				t.Fatal("expected unsupported canonical dispatch to fail closed")
			}
		})
	}
}

func TestDecideStopsAtFinalHumanGate(t *testing.T) {
	t.Parallel()
	decision, err := DecideWithRegistry(gsd.WorkflowSnapshot{Phase: "complete", Next: gsd.NextDispatch{Action: "stop"}}, testSupervisorRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if decision.Kind != DecisionFinalGate || decision.Reason == "" {
		t.Fatalf("decision=%+v", decision)
	}
}

func testSupervisorRegistry() gsd.UnitRegistry {
	return gsd.UnitRegistry{Units: map[string]gsd.UnitMetadata{
		"plan-milestone":    {UnitType: "plan-milestone", PhaseChain: []string{"planning"}},
		"discuss-milestone": {UnitType: "discuss-milestone", PhaseChain: []string{"discuss", "planning"}},
		"execute-task":      {UnitType: "execute-task", PhaseChain: []string{"execution"}},
	}}
}

func TestDecideStopsUnsafeSkipForHumanDecision(t *testing.T) {
	t.Parallel()
	decision, err := DecideWithRegistry(gsd.WorkflowSnapshot{Phase: "executing", Next: gsd.NextDispatch{Action: "skip", UnitType: "execute-task", UnitID: "M001/S01/T01"}}, testSupervisorRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if decision.Kind != DecisionBlocked || decision.Reason == "" {
		t.Fatalf("decision=%+v", decision)
	}
}
