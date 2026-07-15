package store

import (
	"context"
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
