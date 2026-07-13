package gsd

import "testing"

func TestReconcileRejectsPrematureSuccess(t *testing.T) {
	t.Parallel()

	terminal, err := Reconcile("new-milestone", Result{Terminal: TerminalSuccess}, WorkflowSnapshot{
		Phase: "pre-planning", Next: NextDispatch{Action: "dispatch", UnitType: "discuss-milestone", UnitID: "M001"},
	})
	if terminal != TerminalBlocked || err == nil {
		t.Fatalf("terminal=%s err=%v", terminal, err)
	}
}

func TestReconcileAcceptsCompleteQuery(t *testing.T) {
	t.Parallel()

	terminal, err := Reconcile("auto", Result{Terminal: TerminalSuccess}, WorkflowSnapshot{Next: NextDispatch{Action: "complete"}})
	if terminal != TerminalSuccess || err != nil {
		t.Fatalf("terminal=%s err=%v", terminal, err)
	}
}
