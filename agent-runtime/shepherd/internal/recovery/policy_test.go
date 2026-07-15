package recovery

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
)

func TestFailureClassSetIsExhaustiveAndUnknownFailsClosed(t *testing.T) {
	t.Parallel()
	want := []FailureClass{
		FailureRuntimeContractMismatch,
		FailureModelMismatch,
		FailureThinkingMismatch,
		FailureArtifactMissing,
		FailureArtifactInvalid,
		FailureStaleHead,
		FailureDirtyTree,
		FailureWriteScopeBreach,
		FailureDeadWorker,
		FailureSilentTool,
		FailureInterrupted,
		FailureGitHubPublishUncertain,
		FailureOutboxUncertain,
		FailureValidationFailed,
		FailureRatificationFailed,
		FailureRetryExhausted,
		FailureHumanRequired,
		FailureUnknown,
	}
	if got := AllFailureClasses(); !reflect.DeepEqual(got, want) {
		t.Fatalf("failure classes=%v want %v", got, want)
	}
	for _, class := range want {
		if parsed, err := ParseFailureClass(string(class)); err != nil || parsed != class {
			t.Fatalf("parse %q=%q err=%v", class, parsed, err)
		}
	}
	if _, err := ParseFailureClass("runtime_failure"); err == nil {
		t.Fatal("legacy broad runtime failure class was accepted")
	}
}

func TestFailurePolicyRoutesPlannerBlockAndAwaitDecision(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		failure   Failure
		planner   bool
		direct    Action
		exhausted Action
		allowed   []Action
	}{
		{name: "dead worker", failure: Failure{Class: FailureDeadWorker, Reversible: true}, planner: true, exhausted: ActionAwaitDecision, allowed: []Action{ActionRetrySameUnit, ActionRetryAfterBackoff, ActionRunRecoveryPlan, ActionAwaitDecision, ActionBlock}},
		{name: "irreversible dead worker", failure: Failure{Class: FailureDeadWorker}, direct: ActionBlock, exhausted: ActionBlock},
		{name: "silent tool", failure: Failure{Class: FailureSilentTool, Reversible: true}, planner: true, exhausted: ActionAwaitDecision, allowed: []Action{ActionRetryAfterBackoff, ActionRunRecoveryPlan, ActionAwaitDecision, ActionBlock}},
		{name: "artifact missing", failure: Failure{Class: FailureArtifactMissing, Reversible: true}, planner: true, exhausted: ActionAwaitDecision, allowed: []Action{ActionRetrySameUnit, ActionRetryAfterBackoff, ActionRunRecoveryPlan, ActionAwaitDecision, ActionBlock}},
		{name: "artifact invalid", failure: Failure{Class: FailureArtifactInvalid, Reversible: true}, planner: true, exhausted: ActionAwaitDecision, allowed: []Action{ActionRetrySameUnit, ActionRetryAfterBackoff, ActionRunRecoveryPlan, ActionAwaitDecision, ActionBlock}},
		{name: "interrupted", failure: Failure{Class: FailureInterrupted, Reversible: true}, planner: true, exhausted: ActionAwaitDecision, allowed: []Action{ActionRetrySameUnit, ActionRetryAfterBackoff, ActionRunRecoveryPlan, ActionAwaitDecision, ActionBlock}},
		{name: "reversible validation", failure: Failure{Class: FailureValidationFailed, Reversible: true}, planner: true, exhausted: ActionAwaitDecision, allowed: []Action{ActionRetrySameUnit, ActionRetryAfterBackoff, ActionRunRecoveryPlan, ActionAwaitDecision, ActionBlock}},
		{name: "irreversible validation", failure: Failure{Class: FailureValidationFailed}, direct: ActionBlock, exhausted: ActionBlock},
		{name: "model mismatch", failure: Failure{Class: FailureModelMismatch}, direct: ActionBlock, exhausted: ActionBlock},
		{name: "thinking mismatch", failure: Failure{Class: FailureThinkingMismatch}, direct: ActionBlock, exhausted: ActionBlock},
		{name: "scope breach", failure: Failure{Class: FailureWriteScopeBreach}, direct: ActionBlock, exhausted: ActionBlock},
		{name: "dirty tree", failure: Failure{Class: FailureDirtyTree}, direct: ActionBlock, exhausted: ActionBlock},
		{name: "stale head", failure: Failure{Class: FailureStaleHead}, direct: ActionBlock, exhausted: ActionBlock},
		{name: "ratification failed", failure: Failure{Class: FailureRatificationFailed}, direct: ActionBlock, exhausted: ActionBlock},
		{name: "github uncertain", failure: Failure{Class: FailureGitHubPublishUncertain}, direct: ActionBlock, exhausted: ActionBlock},
		{name: "outbox uncertain", failure: Failure{Class: FailureOutboxUncertain}, direct: ActionBlock, exhausted: ActionBlock},
		{name: "unknown", failure: Failure{Class: FailureUnknown}, direct: ActionAwaitDecision, exhausted: ActionAwaitDecision},
		{name: "retry exhausted", failure: Failure{Class: FailureRetryExhausted}, direct: ActionAwaitDecision, exhausted: ActionAwaitDecision},
		{name: "human required", failure: Failure{Class: FailureHumanRequired}, direct: ActionAwaitDecision, exhausted: ActionAwaitDecision},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			policy, err := PolicyFor(test.failure)
			if err != nil {
				t.Fatal(err)
			}
			if policy.PlannerEligible != test.planner || policy.DirectAction != test.direct || policy.ExhaustedAction != test.exhausted {
				t.Fatalf("policy=%+v", policy)
			}
			for _, action := range test.allowed {
				if !policy.Allows(action) {
					t.Fatalf("action %s not allowed by %+v", action, policy)
				}
			}
		})
	}
}

func TestClassifyUsesTypedFailuresAndFailsUnknownClosed(t *testing.T) {
	t.Parallel()
	typed := MarkFailure(FailureArtifactInvalid, true, errors.New("untrusted detail"))
	failure := Classify(gsd.Result{Terminal: gsd.TerminalError, Err: typed})
	if failure.Class != FailureArtifactInvalid || !failure.Reversible {
		t.Fatalf("typed classification=%+v", failure)
	}
	if got := Classify(gsd.Result{Terminal: gsd.TerminalTimeout, Err: context.DeadlineExceeded}); got.Class != FailureInterrupted {
		t.Fatalf("timeout classification=%+v", got)
	}
	if got := Classify(gsd.Result{Terminal: gsd.TerminalError, Err: errors.New("novel uncontrolled failure")}); got.Class != FailureUnknown || got.Reversible {
		t.Fatalf("unknown classification=%+v", got)
	}
	for _, diagnostic := range []string{"artifact missing", "silent tool", "dead worker", "process failed", "github uncertain"} {
		if got := Classify(gsd.Result{Terminal: gsd.TerminalError, Err: errors.New(diagnostic)}); got.Class != FailureUnknown || got.Reversible {
			t.Fatalf("untrusted diagnostic %q classified as %+v", diagnostic, got)
		}
	}
	if got := Classify(gsd.Result{Terminal: gsd.TerminalError, Err: errors.Join(
		MarkFailure(FailureArtifactMissing, true, errors.New("artifact")), errors.New("store failure"))}); got.Class != FailureUnknown {
		t.Fatalf("typed failure masked unknown sibling: %+v", got)
	}
	joined := errors.Join(MarkFailure(FailureArtifactMissing, true, errors.New("artifact")),
		MarkFailure(FailureDirtyTree, false, errors.New("dirty")))
	if got := Classify(gsd.Result{Terminal: gsd.TerminalError, ExitCode: 2, Err: joined}); got.Class != FailureDirtyTree || got.Reversible {
		t.Fatalf("unsafe joined classification=%+v", got)
	}
	if got := Classify(gsd.Result{Terminal: gsd.TerminalBlocked, ExitCode: 2, Err: errors.New("blocked")}); got.Class != FailureUnknown || got.Reversible {
		t.Fatalf("blocked exit classification=%+v", got)
	}
	if got := Classify(gsd.Result{Terminal: gsd.TerminalError, ExitCode: 2, Err: errors.New("untyped exit")}); got.Class != FailureUnknown || got.Reversible {
		t.Fatalf("untyped exit classification=%+v", got)
	}
	if got := Classify(gsd.Result{Terminal: gsd.TerminalError, ExitCode: 2, Err: gsd.ErrDeadWorker}); got.Class != FailureDeadWorker || !got.Reversible {
		t.Fatalf("typed dead-worker classification=%+v", got)
	}
	retryDecision := MarkDecision(Failure{Class: FailureArtifactMissing, Reversible: true}, ActionRetryAfterBackoff, errors.New("artifact"))
	joinedDecision := errors.Join(retryDecision, MarkFailure(FailureDirtyTree, false, errors.New("dirty")))
	if ShouldRetry(joinedDecision) {
		t.Fatal("retry decision masked joined unsafe failure")
	}
	if got := Classify(gsd.Result{Terminal: gsd.TerminalError, Err: joinedDecision}); got.Class != FailureDirtyTree {
		t.Fatalf("joined retry decision classification=%+v", got)
	}
	if ShouldRetry(errors.Join(retryDecision, errors.New("cleanup failure"))) {
		t.Fatal("retry decision masked joined unknown failure")
	}
	for _, sentinel := range []error{gsd.ErrSilentTool, gsd.ErrDeadWorker, context.DeadlineExceeded, gsd.ErrRuntimeContractMismatch} {
		if got := Classify(gsd.Result{Terminal: gsd.TerminalError, Err: errors.Join(sentinel, errors.New("unknown sibling"))}); got.Class != FailureUnknown {
			t.Fatalf("sentinel %v masked unknown sibling as %+v", sentinel, got)
		}
	}
}

func TestActionsAreStrictAndClassForbiddenActionsFail(t *testing.T) {
	t.Parallel()
	for _, action := range []Action{ActionRetrySameUnit, ActionRetryAfterBackoff, ActionRunRecoveryPlan, ActionAwaitDecision, ActionBlock, ActionFinalHumanGate} {
		if parsed, err := ParseAction(string(action)); err != nil || parsed != action {
			t.Fatalf("parse %q=%q err=%v", action, parsed, err)
		}
	}
	if _, err := ParseAction("execute_shell"); err == nil {
		t.Fatal("unknown recovery action accepted")
	}
	policy, err := PolicyFor(Failure{Class: FailureSilentTool, Reversible: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := policy.ValidateAction(ActionRetrySameUnit); err == nil {
		t.Fatal("class-forbidden retry_same_unit accepted for silent_tool")
	}
}
