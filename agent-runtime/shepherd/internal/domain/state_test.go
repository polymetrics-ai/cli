package domain

import "testing"

func TestRunTransitionFailClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		from RunState
		to   RunState
		ok   bool
	}{
		{RunPlanned, RunRunning, true},
		{RunRunning, RunBlocked, true},
		{RunBlocked, RunRunning, false},
		{RunRunning, RunHumanGate, true},
		{RunHumanGate, RunComplete, false},
		{RunComplete, RunRunning, false},
		{RunFailed, RunComplete, false},
	}
	for _, tt := range tests {
		err := ValidateRunTransition(tt.from, tt.to)
		if (err == nil) != tt.ok {
			t.Fatalf("transition %s -> %s: got error %v, want ok=%v", tt.from, tt.to, err, tt.ok)
		}
	}
}

func TestBlockedResumeRequiresHumanDecisionAndGenerationBump(t *testing.T) {
	t.Parallel()

	decision := HumanDecision{RunID: "run-1", Generation: 4, ActorKind: ActorHuman, Approved: true}
	next, err := ResumeBlocked("run-1", 4, decision)
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if next != 5 {
		t.Fatalf("generation=%d want 5", next)
	}
	decision.ActorKind = ActorAgent
	if _, err := ResumeBlocked("run-1", 4, decision); err == nil {
		t.Fatal("expected agent decision to be rejected")
	}
}

func TestFailedResumeRequiresSameExplicitHumanDecision(t *testing.T) {
	t.Parallel()

	decision := HumanDecision{RunID: "run-1", Generation: 2, ActorKind: ActorHuman, Approved: true}
	next, err := ResumeStopped("run-1", 2, RunFailed, decision)
	if err != nil || next != 3 {
		t.Fatalf("resume failed delivery: generation=%d err=%v", next, err)
	}
	if _, err := ResumeStopped("run-1", 2, RunReady, decision); err == nil {
		t.Fatal("ready delivery accepted as explicitly resumable")
	}
}

func TestMergeCapabilityCannotBeGranted(t *testing.T) {
	t.Parallel()

	if _, err := NewGrant("run-1", "repo", 372, Capability("merge.main"), 1); err == nil {
		t.Fatal("expected merge-to-main capability to be rejected")
	}
}
