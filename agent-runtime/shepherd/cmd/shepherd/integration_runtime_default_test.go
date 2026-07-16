//go:build !integration

package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReleaseBuildIgnoresIntegrationEnvironment(t *testing.T) {
	t.Setenv("SHEPHERD_INTEGRATION_GSD_EXECUTOR", filepath.Join(t.TempDir(), "fake"))
	t.Setenv("SHEPHERD_INTEGRATION_CRASH_BOUNDARY", "final_gate_projected")
	t.Setenv("SHEPHERD_INTEGRATION_OUTBOX_CRASH", "post_enqueue")
	t.Setenv("SHEPHERD_INTEGRATION_SHORT_EFFECT_TTL", "1")
	t.Setenv("SHEPHERD_INTEGRATION_SHORT_DECISION_TTL", "1")
	t.Setenv("SHEPHERD_INTEGRATION_MUTATE_AFTER_VALIDATION", "1")
	t.Setenv("SHEPHERD_INTEGRATION_EXIT_AWAITING_DECISION", "1")
	if executor, err := integrationGSDExecutor(); err != nil || executor != "" {
		t.Fatalf("release executor=%q err=%v", executor, err)
	}
	if integrationEffectTTL() != effectClaimTTL {
		t.Fatalf("release effect TTL=%s", integrationEffectTTL())
	}
	if integrationDecisionTTL() != 24*time.Hour {
		t.Fatalf("release decision TTL=%s", integrationDecisionTTL())
	}
	if integrationExitAwaitingDecision() {
		t.Fatal("release build honored integration decision-exit seam")
	}
	workDir := t.TempDir()
	statePath := filepath.Join(workDir, ".gsd", "STATE.md")
	if err := os.MkdirAll(filepath.Dir(statePath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statePath, []byte("unchanged\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	integrationEffectEnqueuedBoundary()
	integrationRetryBoundary()
	integrationFinalGateBoundary()
	if err := integrationPostValidationBoundary(workDir); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(statePath)
	if err != nil || string(raw) != "unchanged\n" {
		t.Fatalf("release seam changed state: %q err=%v", raw, err)
	}
}
