package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
)

func TestDecisionRequestSurvivesRestartAndConsumesOnce(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	now := time.Unix(1_700_000_000, 0).UTC()
	request := DecisionRequest{
		RequestID: "decision-1", DeliveryID: "issue-389", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390,
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
	t.Cleanup(func() { _ = second.Close() })
	loaded, err := second.GetDecisionRequest(ctx, request.RequestID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.RequestID != request.RequestID || loaded.Options[0] != "retry" {
		t.Fatalf("loaded request=%+v", loaded)
	}
	lease, err := second.AcquireLease(ctx, request.DeliveryID, "publisher", now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := second.BeginAttempt(ctx, request.DeliveryID, lease.Owner); err != nil {
		t.Fatal(err)
	}
	if err := second.FinishAttempt(ctx, request.DeliveryID, lease.Owner, domain.RunAwaitingDecision); err != nil {
		t.Fatal(err)
	}
	if err := second.MarkDecisionRequestPublishedFenced(ctx, lease, request, 42, now); err != nil {
		t.Fatal(err)
	}
	if err := second.ApplyDecisionRequestAnswer(ctx, lease, request.RequestID, "retry", "karthik-sivadas",
		6113982, 43, request.Generation, request.HeadSHA, now, now); err != nil {
		t.Fatal(err)
	}
	consumed, err := second.GetDecisionRequest(ctx, request.RequestID)
	if err != nil {
		t.Fatal(err)
	}
	if consumed.Status != DecisionRequestConsumed || consumed.AcceptedAnswer != "retry" ||
		consumed.AcceptedActorID != 6113982 || consumed.AcceptedCommentID != 43 || consumed.ConsumedAt.IsZero() {
		t.Fatalf("consumed=%+v", consumed)
	}
	if laterRun, err := second.BeginAttempt(ctx, request.DeliveryID, lease.Owner); err != nil || laterRun.Generation != 2 {
		t.Fatalf("begin later attempt run=%+v: %v", laterRun, err)
	}
	if err := second.ApplyDecisionRequestAnswer(ctx, lease, request.RequestID, "retry", "karthik-sivadas",
		6113982, 43, request.Generation, request.HeadSHA, now, now); err != nil {
		t.Fatalf("idempotent decision replay failed during later attempt: %v", err)
	}
	run, err := second.GetDeliveryRun(ctx, request.DeliveryID)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := second.GetDecisionRequest(ctx, request.RequestID)
	if err != nil {
		t.Fatal(err)
	}
	if run.Generation != 2 || run.State != domain.RunRunning || !replayed.ConsumedAt.Equal(consumed.ConsumedAt) {
		t.Fatalf("consumed replay mutated state: run=%+v consumed_at=%s want=%s", run,
			replayed.ConsumedAt, consumed.ConsumedAt)
	}
}

func TestFencedResumeCancelsPriorGenerationUnansweredRequests(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	lease, err := db.AcquireLease(ctx, "issue-389", "resume-controller", now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, "issue-389", lease.Owner); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, "issue-389", lease.Owner, domain.RunAwaitingDecision); err != nil {
		t.Fatal(err)
	}
	request := DecisionRequest{
		RequestID: "decision-resume", DeliveryID: "issue-389", Repository: "polymetrics-ai/cli",
		Issue: 389, PullRequest: 390, UnitID: "execute-task/M001/S01/T01", Generation: 1,
		HeadSHA: strings.Repeat("a", 40), Kind: "retry_exhausted", Evidence: "bounded evidence",
		Options: []string{"retry", "stop"}, SafeDefault: "stop", ExpiresAt: now.Add(time.Hour),
		Status: DecisionRequestOpen,
	}
	if _, err := db.UpsertDecisionRequest(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := db.MarkDecisionRequestPublishedFenced(ctx, lease, request, 42, now); err != nil {
		t.Fatal(err)
	}
	if err := db.ResumeDeliveryFenced(ctx, lease, domain.HumanDecision{RunID: "issue-389", Generation: 1,
		ActorKind: domain.ActorHuman, Approved: true}, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	cancelled, err := db.GetDecisionRequest(ctx, request.RequestID)
	if err != nil {
		t.Fatal(err)
	}
	run, err := db.GetDeliveryRun(ctx, "issue-389")
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != DecisionRequestCancelled || run.Generation != 2 || run.State != domain.RunReady {
		t.Fatalf("resume did not settle request: request=%+v run=%+v", cancelled, run)
	}
}

func TestFencedResumeRejectsExpiredLeaseWithoutCancellingRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	lease, err := db.AcquireLease(ctx, "issue-389", "stale-resume", now, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, "issue-389", lease.Owner); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, "issue-389", lease.Owner, domain.RunAwaitingDecision); err != nil {
		t.Fatal(err)
	}
	request := DecisionRequest{RequestID: "decision-stale-resume", DeliveryID: "issue-389",
		Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390, UnitID: "plan-milestone/M001",
		Generation: 1, HeadSHA: strings.Repeat("a", 40), Kind: "retry_exhausted",
		Evidence: "bounded evidence", Options: []string{"retry", "stop"}, SafeDefault: "stop",
		ExpiresAt: now.Add(time.Hour), Status: DecisionRequestOpen}
	if _, err := db.UpsertDecisionRequest(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := db.ResumeDeliveryFenced(ctx, lease, domain.HumanDecision{RunID: "issue-389", Generation: 1,
		ActorKind: domain.ActorHuman, Approved: true}, now.Add(2*time.Second)); err == nil {
		t.Fatal("expired resume lease advanced delivery")
	}
	loaded, err := db.GetDecisionRequest(ctx, request.RequestID)
	if err != nil {
		t.Fatal(err)
	}
	run, err := db.GetDeliveryRun(ctx, "issue-389")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Status != DecisionRequestOpen || run.Generation != 1 || run.State != domain.RunAwaitingDecision {
		t.Fatalf("failed fenced resume mutated state: request=%+v run=%+v", loaded, run)
	}
}

func TestReplyCreatedBeforeExpiryAppliesAfterDelayedRecovery(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	lease, err := db.AcquireLease(ctx, "issue-389", "reply-controller", now, 3*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, "issue-389", lease.Owner); err != nil {
		t.Fatal(err)
	}
	request := DecisionRequest{
		RequestID: "decision-delayed", DeliveryID: "issue-389", Repository: "polymetrics-ai/cli",
		Issue: 389, PullRequest: 390, UnitID: "execute-task/M001/S01/T01", Generation: 1,
		HeadSHA: strings.Repeat("b", 40), Kind: "human_required", Evidence: "bounded evidence",
		Options: []string{"continue", "stop"}, SafeDefault: "stop", ExpiresAt: now.Add(time.Hour),
		Status: DecisionRequestOpen,
	}
	if _, err := db.UpsertDecisionRequest(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, request.DeliveryID, lease.Owner, domain.RunAwaitingDecision); err != nil {
		t.Fatal(err)
	}
	if err := db.MarkDecisionRequestPublishedFenced(ctx, lease, request, 42, now); err != nil {
		t.Fatal(err)
	}
	if err := db.ApplyDecisionRequestAnswer(ctx, lease, request.RequestID, "continue", "karthik-sivadas",
		6113982, 43, request.Generation, request.HeadSHA, now.Add(30*time.Minute), now.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	run, err := db.GetDeliveryRun(ctx, request.DeliveryID)
	if err != nil || run.State != domain.RunReady || run.Generation != 2 {
		t.Fatalf("delayed reply run=%+v err=%v", run, err)
	}
}

func TestStopReplyIsAtomicallyConsumedWhenDeliveryAlreadyBlocked(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	lease, err := db.AcquireLease(ctx, "issue-389", "reply-controller", now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, "issue-389", lease.Owner); err != nil {
		t.Fatal(err)
	}
	request := DecisionRequest{
		RequestID: "decision-stop", DeliveryID: "issue-389", Repository: "polymetrics-ai/cli",
		Issue: 389, PullRequest: 390, UnitID: "execute-task/M001/S01/T01", Generation: 1,
		HeadSHA: strings.Repeat("b", 40), Kind: "human_required", Evidence: "bounded evidence",
		Options: []string{"continue", "stop"}, SafeDefault: "stop", ExpiresAt: now.Add(time.Hour),
		Status: DecisionRequestOpen,
	}
	if _, err := db.UpsertDecisionRequest(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, request.DeliveryID, lease.Owner, domain.RunAwaitingDecision); err != nil {
		t.Fatal(err)
	}
	if err := db.MarkDecisionRequestPublishedFenced(ctx, lease, request, 42, now); err != nil {
		t.Fatal(err)
	}
	if err := db.BlockAwaitingDecision(ctx, request.DeliveryID, request.Generation); err != nil {
		t.Fatal(err)
	}
	if err := db.ApplyDecisionRequestAnswer(ctx, lease, request.RequestID, "stop", "karthik-sivadas",
		6113982, 43, request.Generation, request.HeadSHA, now, now); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.GetDecisionRequest(ctx, request.RequestID)
	if err != nil || loaded.Status != DecisionRequestConsumed || loaded.AcceptedCommentID != 43 {
		t.Fatalf("consumed stop request=%+v err=%v", loaded, err)
	}
}

func TestExpiredDecisionRequestAppliesSafeStopAndBlocksDelivery(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	lease, err := db.AcquireLease(ctx, "issue-389", "expiry-controller", now, 3*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, "issue-389", lease.Owner); err != nil {
		t.Fatal(err)
	}
	request := DecisionRequest{
		RequestID: "decision-expired", DeliveryID: "issue-389", Repository: "polymetrics-ai/cli",
		Issue: 389, PullRequest: 390, UnitID: "execute-task/M001/S01/T01", Generation: 1,
		HeadSHA: strings.Repeat("b", 40), Kind: "human_required", Evidence: "bounded evidence",
		Options: []string{"continue", "stop"}, SafeDefault: "stop", ExpiresAt: now.Add(time.Hour),
		Status: DecisionRequestOpen,
	}
	if _, err := db.UpsertDecisionRequest(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, "issue-389", lease.Owner, domain.RunAwaitingDecision); err != nil {
		t.Fatal(err)
	}
	if err := db.ExpireDecisionRequestAndBlock(ctx, lease, request.RequestID, now.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.GetDecisionRequest(ctx, request.RequestID)
	if err != nil || loaded.Status != DecisionRequestExpired {
		t.Fatalf("expired request=%+v err=%v", loaded, err)
	}
	run, err := db.GetDeliveryRun(ctx, "issue-389")
	if err != nil || run.State != domain.RunBlocked {
		t.Fatalf("blocked run=%+v err=%v", run, err)
	}
}

func TestDecisionRequestRejectsStaleGenerationHeadAndExpiry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "shepherd.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0).UTC()
	request := DecisionRequest{
		RequestID: "decision-2", DeliveryID: "issue-389", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390,
		UnitID: "execute-task/M001/S01/T01", Generation: 2, HeadSHA: strings.Repeat("b", 40),
		Kind: "human_required", Evidence: "bounded evidence", Options: []string{"continue", "stop"}, ExpiresAt: now.Add(time.Hour), Status: DecisionRequestOpen,
	}
	if _, err := db.UpsertDecisionRequest(ctx, request); err != nil {
		t.Fatal(err)
	}
	lease, err := db.AcquireLease(ctx, request.DeliveryID, "publisher", now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.MarkDecisionRequestPublishedFenced(ctx, lease, request, 42, now); err != nil {
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
			if err := db.ApplyDecisionRequestAnswer(ctx, lease, request.RequestID, "continue", "karthik-sivadas", 6113982, 43, test.generation, test.head, test.now, test.now); err == nil {
				t.Fatal("stale decision answer accepted")
			}
		})
	}
	loaded, err := db.GetDecisionRequest(ctx, request.RequestID)
	if err != nil || loaded.Status != DecisionRequestPublished || loaded.AcceptedAnswer != "" ||
		loaded.AcceptedCommentID != 0 {
		t.Fatalf("failed atomic answer mutated request=%+v err=%v", loaded, err)
	}
}
