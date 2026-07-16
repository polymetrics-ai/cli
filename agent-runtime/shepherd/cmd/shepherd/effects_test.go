package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	decisionlog "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/decision"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
	shepherdgit "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/git"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/outbox"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/recovery"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
)

func TestOutboxOpenFailureAfterAttemptStartFinalizesDelivery(t *testing.T) {
	ctx := context.Background()
	_, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	previousOpen := externalEffectStoreOpen
	calls := 0
	externalEffectStoreOpen = func(openCtx context.Context, path string) (*outbox.Store, error) {
		calls++
		if calls > 1 {
			return nil, errors.New("injected outbox open failure")
		}
		return outbox.Open(openCtx, path)
	}
	t.Cleanup(func() { externalEffectStoreOpen = previousOpen })
	if err := runSupervise(ctx, runner, config, testUnitRegistry(), 389, contextPath, false,
		"shepherd", "fake runtime"); err == nil {
		t.Fatal("outbox open failure reported success")
	}
	authorityStore, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = authorityStore.Close() })
	run, err := authorityStore.GetDeliveryRun(ctx, "issue-389")
	if err != nil || run.State != domain.RunFailed {
		t.Fatalf("delivery run=%+v err=%v", run, err)
	}
}

func TestExternalEffectControllerRejectsMutableConfigTargetDrift(t *testing.T) {
	ctx := context.Background()
	authorityStore, config, governance := setupRecoveryGovernanceTest(t, recovery.Failure{Class: recovery.FailureDeadWorker, Reversible: true})
	config.Repository = "polymetrics-ai/rebound"
	if controller, err := openExternalEffectController(ctx, authorityStore, governance.Lease, config,
		governance.DeliveryID, governance.Issue, governance.Generation, governance.HeadSHA); err == nil {
		_ = controller.Close()
		t.Fatal("mutable repository drift was accepted")
	}
	config.Repository = "polymetrics-ai/cli"
	config.PullRequest++
	if controller, err := openExternalEffectController(ctx, authorityStore, governance.Lease, config,
		governance.DeliveryID, governance.Issue, governance.Generation, governance.HeadSHA); err == nil {
		_ = controller.Close()
		t.Fatal("mutable pull request drift was accepted")
	}
}

func TestExternalEffectControllerRejectsStaleAuthorityLeaseBeforeEnqueue(t *testing.T) {
	authorityStore, config, governance := setupRecoveryGovernanceTest(t, recovery.Failure{Class: recovery.FailureDeadWorker, Reversible: true})
	controller, err := openExternalEffectController(context.Background(), authorityStore, governance.Lease,
		config, governance.DeliveryID, governance.Issue, governance.Generation, governance.HeadSHA)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = controller.Close() })
	if _, err := authorityStore.AcquireLease(context.Background(), governance.DeliveryID, "new-controller",
		time.Now().UTC().Add(2*time.Minute), time.Minute); err != nil {
		t.Fatal(err)
	}
	summary := "summary"
	digest := sha256.Sum256([]byte(summary))
	_, err = controller.RequestSummary(context.Background(), decisionlog.Snapshot{
		DeliveryID: governance.DeliveryID, Revision: 1,
		LedgerHash: "sha256:" + hex.EncodeToString(digest[:]), Summary: summary,
	})
	if err == nil {
		t.Fatal("stale authority lease authorized an external effect")
	}
	records, listErr := controller.store.ListDelivery(context.Background(), governance.DeliveryID)
	if listErr != nil || len(records) != 0 {
		t.Fatalf("stale authority persisted effects=%+v err=%v", records, listErr)
	}
}

func TestExternalEffectExecutionRejectsCanonicalHeadDrift(t *testing.T) {
	ctx := context.Background()
	authorityStore, config, governance := setupRecoveryGovernanceTest(t, recovery.Failure{Class: recovery.FailureDeadWorker, Reversible: true})
	controller, err := openExternalEffectController(ctx, authorityStore, governance.Lease, config,
		governance.DeliveryID, governance.Issue, governance.Generation, governance.HeadSHA)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = controller.Close() })
	if err := os.WriteFile(filepath.Join(config.WorkDir, "head-drift.txt"), []byte("drift\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitForTest(t, config.WorkDir, "add", "head-drift.txt")
	runGitForTest(t, config.WorkDir, "commit", "-m", "test: advance canonical head")
	summary := "summary"
	digest := sha256.Sum256([]byte(summary))
	if _, err := controller.RequestSummary(ctx, decisionlog.Snapshot{
		DeliveryID: governance.DeliveryID, Revision: 1,
		LedgerHash: "sha256:" + hex.EncodeToString(digest[:]), Summary: summary,
	}); !errors.Is(err, outbox.ErrFenced) {
		t.Fatalf("stale-head external effect error=%v", err)
	}
}

func TestDecisionSummaryProjectionSurvivesHeadAndEpochRolloverWithoutDuplicateEffect(t *testing.T) {
	ctx := context.Background()
	authorityStore, config, governance := setupRecoveryGovernanceTest(t, recovery.Failure{Class: recovery.FailureDeadWorker, Reversible: true})
	first, err := openExternalEffectController(ctx, authorityStore, governance.Lease, config,
		governance.DeliveryID, governance.Issue, governance.Generation, governance.HeadSHA)
	if err != nil {
		t.Fatal(err)
	}
	summary := "## Shepherd decisions\n\n| Decision |\n|---|\n| continue |\n"
	digest := sha256.Sum256([]byte(summary))
	snapshot := decisionlog.Snapshot{DeliveryID: governance.DeliveryID, Revision: 1,
		LedgerHash: "sha256:" + hex.EncodeToString(digest[:]), Summary: summary}
	firstResult, err := first.RequestSummary(ctx, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if err := authorityStore.ReleaseLease(ctx, governance.Lease); err != nil {
		t.Fatal(err)
	}
	if err := authorityStore.FinishAttempt(ctx, governance.DeliveryID, governance.ExecutionID, domain.RunFailed); err != nil {
		t.Fatal(err)
	}
	if err := authorityStore.ResumeDelivery(ctx, domain.HumanDecision{
		RunID: governance.DeliveryID, Generation: governance.Generation,
		ActorKind: domain.ActorHuman, Approved: true,
	}); err != nil {
		t.Fatal(err)
	}
	secondLease, err := authorityStore.AcquireLease(ctx, governance.DeliveryID, "rollover-controller",
		time.Now().UTC(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(config.WorkDir, "rollover.txt"), []byte("rollover\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGitForTest(t, config.WorkDir, "add", "rollover.txt")
	runGitForTest(t, config.WorkDir, "commit", "-m", "test: roll canonical head")
	canonical, err := shepherdgit.Inspect(ctx, config.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	second, err := openExternalEffectController(ctx, authorityStore, secondLease, config,
		governance.DeliveryID, governance.Issue, governance.Generation+1, canonical.HeadSHA)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })
	secondResult, err := second.RequestSummary(ctx, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if firstResult != secondResult {
		t.Fatalf("rollover changed projection result: first=%+v second=%+v", firstResult, secondResult)
	}
	records, err := second.store.ListDelivery(ctx, governance.DeliveryID)
	if err != nil || len(records) != 1 || records[0].State != outbox.StateSent {
		t.Fatalf("rollover effects=%+v err=%v", records, err)
	}
}

func TestStartupRecoversExpiredQuestionClaimBeforeProjection(t *testing.T) {
	ctx := context.Background()
	authorityStore, config, governance := setupRecoveryGovernanceTest(t, recovery.Failure{Class: recovery.FailureDeadWorker, Reversible: true})
	controller, err := openExternalEffectController(ctx, authorityStore, governance.Lease, config,
		governance.DeliveryID, governance.Issue, governance.Generation, governance.HeadSHA)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = controller.Close() })
	now := time.Now().UTC()
	request, err := authorityStore.UpsertDecisionRequest(ctx, store.DecisionRequest{
		RequestID: "decision-expired-claim", DeliveryID: governance.DeliveryID,
		Repository: config.Repository, Issue: governance.Issue, PullRequest: config.PullRequest,
		UnitID: governance.UnitID, Generation: governance.Generation, HeadSHA: governance.HeadSHA,
		Kind: "retry_exhausted", Evidence: "retry budget exhausted", Options: []string{"retry", "stop"},
		RecommendedOption: "retry", SafeDefault: "stop", ExpiresAt: now.Add(time.Hour),
		Status: store.DecisionRequestOpen,
	})
	if err != nil {
		t.Fatal(err)
	}
	intent, err := decisionQuestionIntent(request)
	if err != nil {
		t.Fatal(err)
	}
	authorization, err := controller.policy.Authorize(ctx, intent, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := controller.store.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	if _, err := controller.store.Claim(ctx, authorization, authorization.Owner(), authorization.Epoch(),
		now, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := authorityStore.FinishAttempt(ctx, governance.DeliveryID, governance.ExecutionID,
		domain.RunAwaitingDecision); err != nil {
		t.Fatal(err)
	}
	controller.now = func() time.Time { return now.Add(2 * time.Millisecond) }
	if err := reconcileExternalEffects(ctx, authorityStore, controller, config.StateDir,
		governance.DeliveryID); err != nil {
		t.Fatal(err)
	}
	record, err := controller.store.Get(ctx, intent.EffectID())
	if err != nil || record.State != outbox.StateSent {
		t.Fatalf("recovered question effect=%+v err=%v", record, err)
	}
	projected, err := authorityStore.GetDecisionRequest(ctx, request.RequestID)
	if err != nil || projected.Status != store.DecisionRequestPublished || projected.GitHubCommentID <= 0 {
		t.Fatalf("projected question=%+v err=%v", projected, err)
	}
}

func TestExpiredPendingQuestionConvergesToCancelledEffectAndSafeStop(t *testing.T) {
	ctx := context.Background()
	authorityStore, config, governance := setupRecoveryGovernanceTest(t, recovery.Failure{Class: recovery.FailureDeadWorker, Reversible: true})
	controller, err := openExternalEffectController(ctx, authorityStore, governance.Lease, config,
		governance.DeliveryID, governance.Issue, governance.Generation, governance.HeadSHA)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = controller.Close() })
	now := time.Now().UTC()
	request, err := authorityStore.UpsertDecisionRequest(ctx, store.DecisionRequest{
		RequestID: "decision-expired-pending", DeliveryID: governance.DeliveryID,
		Repository: config.Repository, Issue: governance.Issue, PullRequest: config.PullRequest,
		UnitID: governance.UnitID, Generation: governance.Generation, HeadSHA: governance.HeadSHA,
		Kind: "retry_exhausted", Evidence: "retry budget exhausted", Options: []string{"retry", "stop"},
		RecommendedOption: "retry", SafeDefault: "stop", ExpiresAt: now.Add(time.Second),
		Status: store.DecisionRequestOpen,
	})
	if err != nil {
		t.Fatal(err)
	}
	intent, err := decisionQuestionIntent(request)
	if err != nil {
		t.Fatal(err)
	}
	authorization, err := controller.policy.Authorize(ctx, intent, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := controller.store.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	if err := authorityStore.FinishAttempt(ctx, governance.DeliveryID, governance.ExecutionID,
		domain.RunAwaitingDecision); err != nil {
		t.Fatal(err)
	}
	controller.now = func() time.Time { return now.Add(2 * time.Second) }
	if err := reconcileExternalEffects(ctx, authorityStore, controller, config.StateDir,
		governance.DeliveryID); err != nil {
		t.Fatal(err)
	}
	record, err := controller.store.Get(ctx, intent.EffectID())
	if err != nil || record.State != outbox.StateCancelled || record.ErrorCode != outbox.ErrorEffectExpired {
		t.Fatalf("expired effect=%+v err=%v", record, err)
	}
	run, err := authorityStore.GetDeliveryRun(ctx, governance.DeliveryID)
	if err != nil || run.State != domain.RunBlocked {
		t.Fatalf("delivery run=%+v err=%v", run, err)
	}
}

func TestStartupEstablishesLatestLedgerRevisionBeforeRecoveringOlderSummary(t *testing.T) {
	ctx := context.Background()
	authorityStore, config, governance := setupRecoveryGovernanceTest(t, recovery.Failure{Class: recovery.FailureDeadWorker, Reversible: true})
	controller, err := openExternalEffectController(ctx, authorityStore, governance.Lease, config,
		governance.DeliveryID, governance.Issue, governance.Generation, governance.HeadSHA)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = controller.Close() })
	decisionStore, err := decisionlog.Open(filepath.Join(config.StateDir, "decisions"))
	if err != nil {
		t.Fatal(err)
	}
	appendRecord := func(id, answer string, at time.Time) {
		t.Helper()
		if err := decisionStore.Append(ctx, decisionlog.Record{
			ID: id, DeliveryID: governance.DeliveryID, ExecutionID: governance.ExecutionID,
			UnitID: governance.UnitID, QuestionID: "scope-" + id, Question: "Continue?",
			Answer: answer, Actor: decisionlog.ActorHuman, Basis: "bounded approval", At: at,
		}); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	appendRecord("decision-old", "continue", now)
	oldSnapshot, err := decisionStore.Snapshot(governance.DeliveryID)
	if err != nil {
		t.Fatal(err)
	}
	oldIntent, err := outbox.NewSummaryIntent(outbox.Target{Repository: config.Repository,
		Issue: governance.Issue, PullRequest: config.PullRequest}, governance.DeliveryID,
		governance.Generation, governance.HeadSHA, oldSnapshot.Revision, oldSnapshot.LedgerHash,
		oldSnapshot.Summary)
	if err != nil {
		t.Fatal(err)
	}
	oldAuthorization, err := controller.policy.Authorize(ctx, oldIntent, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := controller.store.Enqueue(ctx, oldAuthorization, now); err != nil {
		t.Fatal(err)
	}
	appendRecord("decision-new", "stop", now.Add(time.Second))
	if err := decisionStore.Close(); err != nil {
		t.Fatal(err)
	}
	if err := reconcileExternalEffects(ctx, authorityStore, controller, config.StateDir,
		governance.DeliveryID); err != nil {
		t.Fatal(err)
	}
	oldRecord, err := controller.store.Get(ctx, oldIntent.EffectID())
	if err != nil || oldRecord.State != outbox.StateCancelled || oldRecord.ErrorCode != outbox.ErrorStaleRevision {
		t.Fatalf("old summary=%+v err=%v", oldRecord, err)
	}
}

func TestStartupReconcilesDecisionLedgerAndQuestionProjectionCrashBoundaries(t *testing.T) {
	if testing.Short() {
		t.Skip("fake supervise setup exercises git-backed startup reconciliation")
	}
	ctx := context.Background()
	_, contextPath, config, runner := setupFakeSuperviseRuntime(t)
	restoreValidator := installFakeIndependentValidator(t, &fakeIndependentValidator{
		result: validFakeValidationResult("openai-codex/gpt-5.6-sol", "PROCEED"),
	})
	defer restoreValidator()
	if err := runSupervise(ctx, runner, config, testUnitRegistry(), 389, contextPath, false, "shepherd", "fake runtime"); err != nil {
		t.Fatal(err)
	}

	authorityStore, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = authorityStore.Close() })
	run, err := authorityStore.GetDeliveryRun(ctx, "issue-389")
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := shepherdgit.Inspect(ctx, config.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := authorityStore.AcquireReconciliationLease(ctx, "issue-389", "effect-reconcile-test",
		time.Now().UTC(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = authorityStore.ReleaseLease(context.Background(), lease) })
	effects, err := openExternalEffectController(ctx, authorityStore, lease, config, "issue-389", 389,
		run.Generation, canonical.HeadSHA)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = effects.Close() })

	decisionStore, err := decisionlog.Open(filepath.Join(config.StateDir, "decisions"))
	if err != nil {
		t.Fatal(err)
	}
	if err := decisionStore.Append(ctx, decisionlog.Record{
		ID: "decision-ledger-gap", DeliveryID: "issue-389", ExecutionID: "execution-gap",
		UnitID: "execute-task/M001/S01/T01", QuestionID: "scope", Question: "Continue?",
		Answer: "continue", Actor: decisionlog.ActorHuman, Basis: "test human approval", At: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	if err := decisionStore.Close(); err != nil {
		t.Fatal(err)
	}

	request, err := authorityStore.UpsertDecisionRequest(ctx, store.DecisionRequest{
		RequestID: "decision-projection-gap", DeliveryID: "issue-389", Repository: config.Repository,
		Issue: 389, PullRequest: config.PullRequest, UnitID: "execute-task/M001/S01/T01", Generation: run.Generation,
		HeadSHA: canonical.HeadSHA, Kind: "retry_exhausted", Evidence: "retry budget exhausted",
		Options: []string{"retry", "stop"}, RecommendedOption: "retry", SafeDefault: "stop",
		ExpiresAt: time.Now().UTC().Add(time.Hour), Status: store.DecisionRequestOpen,
	})
	if err != nil {
		t.Fatal(err)
	}
	questionResult, err := effects.RequestQuestion(ctx, request)
	if err != nil || questionResult.ExternalID <= 0 {
		t.Fatalf("question effect result=%+v err=%v", questionResult, err)
	}
	// Simulate a crash after durable outbox success and before projecting the
	// external comment ID into the authority decision request.
	if err := reconcileExternalEffects(ctx, authorityStore, effects, config.StateDir, "issue-389"); err != nil {
		t.Fatal(err)
	}
	projected, err := authorityStore.GetDecisionRequest(ctx, request.RequestID)
	if err != nil || projected.Status != store.DecisionRequestPublished || projected.GitHubCommentID != questionResult.ExternalID {
		t.Fatalf("projected request=%+v err=%v", projected, err)
	}

	outboxStore, err := outbox.Open(ctx, filepath.Join(config.StateDir, "outbox.db"))
	if err != nil {
		t.Fatal(err)
	}
	records, err := outboxStore.ListDelivery(ctx, "issue-389")
	if err != nil {
		t.Fatal(err)
	}
	if err := outboxStore.Close(); err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("effects=%+v, want one summary and one question", records)
	}
	for _, record := range records {
		if record.State != outbox.StateSent || record.Result.ExternalID <= 0 ||
			(record.Kind != outbox.KindDecisionSummary && record.Kind != outbox.KindDecisionQuestion) {
			t.Fatalf("invalid reconciled effect=%+v", record)
		}
		if strings.Contains(strings.ToLower(string(record.Payload)), "token=") {
			t.Fatalf("effect payload contains secret-shaped data: %s", record.Payload)
		}
	}
	if err := reconcileExternalEffects(ctx, authorityStore, effects, config.StateDir, "issue-389"); err != nil {
		t.Fatalf("idempotent reconciliation: %v", err)
	}
}
