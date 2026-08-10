package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStageValidatedJSONBytesLeavesDestinationUntouchedForInvalidCandidate(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "ledger.json")
	const original = `{"preserve":"evidence"}` + "\n"
	if err := os.WriteFile(destination, []byte(original), 0o644); err != nil {
		t.Fatalf("seed destination: %v", err)
	}

	if _, err := stageValidatedJSONBytes(destination, []byte(`{"broken":`)); err == nil {
		t.Fatal("stage invalid JSON error = nil, want refusal before replacement")
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read destination after invalid candidate: %v", err)
	}
	if string(got) != original {
		t.Fatalf("destination = %q after invalid candidate, want unchanged %q", got, original)
	}
}

func TestAssertCountsRequiresRateLimitFileConservation(t *testing.T) {
	counts := map[string]int{
		"already_complete":           5,
		"recovered_pilot":            4,
		"recovered_seven":            7,
		"newly_materialized":         189,
		"materialized_total":         205,
		"retry_pending":              221,
		"genuinely_blocked":          0,
		"remaining":                  221,
		"rate_limits_declared":       3,
		"rate_limits_unknown":        422,
		"rate_limits_not_applicable": 1,
		"rate_limits_file_total":     426,
		"target_total":               426,
	}
	if err := assertCounts(counts); err != nil {
		t.Fatalf("valid rate-limit conservation: %v", err)
	}
	counts["rate_limits_file_total"] = 425
	if err := assertCounts(counts); err == nil || !strings.Contains(err.Error(), "rate-limit file total") {
		t.Fatalf("missing rate-limit file error = %v, want conservation failure", err)
	}
}

func TestFinalBlockedAttemptRequiresExplicitTerminalOutcome(t *testing.T) {
	retry := BatchEvent{State: "retry_pending", Evidence: "batches/primary.json", Stage: "artifact_inventory_unknown", Reason: "primary source did not yield a complete inventory"}
	if _, ok := finalBlockedAttempt([]BatchEvent{retry}); ok {
		t.Fatal("retry_pending event resolved as genuinely blocked")
	}

	blocked := BatchEvent{State: "genuinely_blocked", Evidence: "batches/final-exhaustion.json", Route: "retained official sources exhausted", Stage: "official_source_exhausted", Reason: "all retained static official routes were attempted"}
	got, ok := finalBlockedAttempt([]BatchEvent{retry, blocked})
	if !ok {
		t.Fatal("explicit genuinely_blocked event was not selected")
	}
	if got != blocked {
		t.Fatalf("final blocked event = %+v, want %+v", got, blocked)
	}

	malformed := blocked
	malformed.Stage = "artifact_inventory_unknown"
	if _, ok := finalBlockedAttempt([]BatchEvent{malformed}); ok {
		t.Fatal("non-terminal genuinely_blocked event resolved the queue")
	}
	if _, ok := finalBlockedAttempt([]BatchEvent{blocked, retry}); ok {
		t.Fatal("terminal outcome superseded by a later retry resolved the queue")
	}
}
