package agentloop

import (
	"slices"
	"testing"
)

func TestSafetyStatusIsClosedWithoutEnableRoute(t *testing.T) {
	t.Setenv("POLYMETRICS_AGENT_LOOP_ENABLE", "1")

	status := CurrentSafetyStatus()
	if status.State != "closed" {
		t.Errorf("state = %q, want closed", status.State)
	}
	if status.RunEnabled {
		t.Error("run enabled = true, want false")
	}
	if status.ResumeEnabled {
		t.Error("resume enabled = true, want false")
	}
	if status.Code != "AUTO_LOOP_DISABLED_PHASE_0" {
		t.Errorf("code = %q, want AUTO_LOOP_DISABLED_PHASE_0", status.Code)
	}
	if status.ExitClass != "safety_disabled" {
		t.Errorf("exit class = %q, want safety_disabled", status.ExitClass)
	}
}

func TestTrackedEntrypointsReturnsDefensiveSortedCopy(t *testing.T) {
	t.Parallel()

	first := TrackedEntrypoints()
	if !slices.Equal(first, []string{"scripts/claude-auto-loop.sh", "scripts/pi-auto-loop.sh"}) {
		t.Fatalf("tracked entrypoints = %v", first)
	}
	first[0] = "mutated"
	second := TrackedEntrypoints()
	if second[0] != "scripts/claude-auto-loop.sh" {
		t.Fatalf("tracked entrypoints leaked internal slice: %v", second)
	}
}

func TestGuardDriverFailsClosed(t *testing.T) {
	t.Parallel()

	tracked := GuardDriver("scripts/claude-auto-loop.sh")
	if tracked.Code != "AUTO_LOOP_DISABLED_PHASE_0" || tracked.ExitCode != 78 {
		t.Fatalf("tracked guard = %+v, want disabled/78", tracked)
	}

	untracked := GuardDriver("scripts/new-auto-loop.sh")
	if untracked.Code != "AUTO_LOOP_ENTRYPOINT_UNTRACKED" || untracked.ExitCode != 64 {
		t.Fatalf("untracked guard = %+v, want untracked/64", untracked)
	}
}
