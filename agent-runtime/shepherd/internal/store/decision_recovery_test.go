package store

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDecisionRequestSurvivesRestartAndConsumesOnce(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	now := time.Unix(1_700_000_000, 0).UTC()
	request := DecisionRequest{
		RequestID: "decision-1", DeliveryID: "issue-389", Issue: 389, PullRequest: 391,
		UnitID: "plan-milestone/M001", Generation: 1, HeadSHA: strings.Repeat("a", 40),
		Kind: "retry_exhausted", Evidence: "bounded evidence", Options: []string{"retry", "stop"},
		RecommendedOption: "retry", SafeDefault: "stop", ExpiresAt: now.Add(time.Hour), Status: DecisionRequestOpen,
	}
	first, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	if _, err := first.UpsertDecisionRequest(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	loaded, err := second.GetDecisionRequest(ctx, request.RequestID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.RequestID != request.RequestID || loaded.Options[0] != "retry" {
		t.Fatalf("loaded request=%+v", loaded)
	}
	if err := second.MarkDecisionRequestPublished(ctx, request.RequestID, 42); err != nil {
		t.Fatal(err)
	}
	if err := second.AcceptDecisionRequestAnswer(ctx, request.RequestID, "retry", "karthik-sivadas", request.Generation, request.HeadSHA, now); err != nil {
		t.Fatal(err)
	}
	consumed, err := second.ConsumeDecisionRequest(ctx, request.RequestID)
	if err != nil {
		t.Fatal(err)
	}
	if consumed.Status != DecisionRequestConsumed || consumed.AcceptedAnswer != "retry" || consumed.ConsumedAt.IsZero() {
		t.Fatalf("consumed=%+v", consumed)
	}
	if _, err := second.ConsumeDecisionRequest(ctx, request.RequestID); err == nil {
		t.Fatal("duplicate decision consumption succeeded")
	}
}

func TestDecisionRequestRejectsStaleGenerationHeadAndExpiry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0).UTC()
	request := DecisionRequest{
		RequestID: "decision-2", DeliveryID: "issue-389", Issue: 389, PullRequest: 391,
		UnitID: "execute-task/M001/S01/T01", Generation: 2, HeadSHA: strings.Repeat("b", 40),
		Kind: "human_required", Evidence: "bounded evidence", Options: []string{"continue", "stop"}, ExpiresAt: now.Add(time.Hour), Status: DecisionRequestOpen,
	}
	if _, err := db.UpsertDecisionRequest(ctx, request); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name       string
		generation int64
		head       string
		now        time.Time
	}{
		{name: "stale generation", generation: 1, head: request.HeadSHA, now: now},
		{name: "stale head", generation: 2, head: strings.Repeat("c", 40), now: now},
		{name: "expired", generation: 2, head: request.HeadSHA, now: now.Add(2 * time.Hour)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := db.AcceptDecisionRequestAnswer(ctx, request.RequestID, "continue", "karthik-sivadas", test.generation, test.head, test.now); err == nil {
				t.Fatal("stale decision answer accepted")
			}
		})
	}
}

func TestRecoveryBudgetIsPerFailureClassAndSurvivesRestart(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	head := strings.Repeat("d", 40)
	first, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0).UTC()
	artifactKey := RecoveryBudgetKey{DeliveryID: "issue-389", Generation: 1, UnitID: "execute-task/M001/S01/T01", HeadSHA: head, FailureClass: "artifact_missing"}
	modelKey := artifactKey
	modelKey.FailureClass = "model_mismatch"
	budget, err := first.BeginRecoveryAttempt(ctx, artifactKey, 2, time.Minute, "missing artifact", "recreate artifact", now)
	if err != nil || budget.Attempts != 1 || budget.NextRetryAt.Sub(now) != time.Minute {
		t.Fatalf("artifact budget=%+v err=%v", budget, err)
	}
	other, err := first.BeginRecoveryAttempt(ctx, modelKey, 1, 0, "model drift", "", now)
	if err != nil || other.Attempts != 1 {
		t.Fatalf("model budget=%+v err=%v", other, err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	budget, err = second.BeginRecoveryAttempt(ctx, artifactKey, 2, time.Minute, "missing again", "recreate again", now.Add(time.Minute))
	if err != nil || budget.Attempts != 2 {
		t.Fatalf("reopened artifact budget=%+v err=%v", budget, err)
	}
	if _, err := second.BeginRecoveryAttempt(ctx, artifactKey, 2, time.Minute, "missing third", "", now.Add(2*time.Minute)); !errors.Is(err, ErrRetryBudgetExhausted) {
		t.Fatalf("error=%v, want recovery budget exhaustion", err)
	}
}
