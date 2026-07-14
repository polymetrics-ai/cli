package gsd

import (
	"errors"
	"testing"
)

func TestReconcileAcceptsOneUnitProgress(t *testing.T) {
	t.Parallel()

	terminal, err := Reconcile("new-milestone", Result{Terminal: TerminalSuccess}, WorkflowSnapshot{}, WorkflowSnapshot{
		Phase: "pre-planning", Next: NextDispatch{Action: "dispatch", UnitType: "discuss-milestone", UnitID: "M001"},
	})
	if terminal != TerminalSuccess || err != nil {
		t.Fatalf("terminal=%s err=%v", terminal, err)
	}
}

func TestReconcileAcceptsTargetedDiscussUnit(t *testing.T) {
	t.Parallel()

	terminal, err := Reconcile("discuss", Result{Terminal: TerminalSuccess}, WorkflowSnapshot{}, WorkflowSnapshot{
		Phase: "pre-planning", Next: NextDispatch{Action: "dispatch", UnitType: "plan-milestone", UnitID: "M001"},
	})
	if terminal != TerminalSuccess || err != nil {
		t.Fatalf("terminal=%s err=%v", terminal, err)
	}
}

func TestReconcileAcceptsCompleteQuery(t *testing.T) {
	t.Parallel()

	terminal, err := Reconcile("auto", Result{Terminal: TerminalSuccess}, WorkflowSnapshot{}, WorkflowSnapshot{Phase: "complete", Next: NextDispatch{Action: "stop"}})
	if terminal != TerminalSuccess || err != nil {
		t.Fatalf("terminal=%s err=%v", terminal, err)
	}
}

func TestReconcileRejectsFalseGreenCanonicalUnit(t *testing.T) {
	t.Parallel()
	before := WorkflowSnapshot{Phase: "pre-planning", Next: NextDispatch{Action: "dispatch", UnitType: "research-milestone", UnitID: "M001"}}
	terminal, err := Reconcile("research-milestone", Result{Terminal: TerminalSuccess}, before, before)
	if terminal != TerminalError || err == nil {
		t.Fatalf("terminal=%s err=%v", terminal, err)
	}
	after := WorkflowSnapshot{Phase: "pre-planning", Next: NextDispatch{Action: "dispatch", UnitType: "plan-milestone", UnitID: "M001"}}
	terminal, err = Reconcile("research-milestone", Result{Terminal: TerminalSuccess}, before, after)
	if terminal != TerminalSuccess || err != nil {
		t.Fatalf("advanced terminal=%s err=%v", terminal, err)
	}
}

func TestReconcileClassifiesMutatingSkipAsARecoverableFence(t *testing.T) {
	t.Parallel()
	before := WorkflowSnapshot{Phase: "planning", Next: NextDispatch{Action: "dispatch", UnitType: "plan-slice", UnitID: "M001/S02"}}
	after := WorkflowSnapshot{Phase: "executing", Next: NextDispatch{Action: "skip", UnitType: "execute-task", UnitID: "M001/S02/T01"}}
	terminal, err := Reconcile("plan-slice", Result{Terminal: TerminalSuccess}, before, after)
	if terminal != TerminalBlocked || !errors.Is(err, ErrMutatingSkip) {
		t.Fatalf("terminal=%s err=%v", terminal, err)
	}
}
