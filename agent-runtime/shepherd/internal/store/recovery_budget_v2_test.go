package store

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/recovery"
)

func TestRecoveryReservationIsAtomicFencedAndIdempotentPerUnitAttempt(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for index := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			candidate := request
			candidate.ClaimToken = strings.Repeat(strconv.Itoa(index), 32)
			_, reserveErr := db.ReserveRecoveryAttempt(ctx, candidate)
			results <- reserveErr
		}()
	}
	wg.Wait()
	close(results)
	var successes int
	for reserveErr := range results {
		if reserveErr == nil {
			successes++
			continue
		}
		if !errors.Is(reserveErr, ErrRecoveryClaimFenced) {
			t.Fatalf("unexpected reserve error: %v", reserveErr)
		}
	}
	if successes != 1 {
		t.Fatalf("successful reservations=%d want 1", successes)
	}
	budget, err := db.GetRecoveryBudget(ctx, request.RecoveryBudgetKey)
	if err != nil {
		t.Fatal(err)
	}
	if budget.Attempts != 1 {
		t.Fatalf("budget attempts=%d want 1", budget.Attempts)
	}
}

func TestRecoveryCompletionRejectsMismatchedEvidenceAndClassForbiddenAction(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	if _, err := db.ReserveRecoveryAttempt(ctx, request); err != nil {
		t.Fatal(err)
	}
	mismatched := recoveryOutcomeFixture(request)
	mismatched.EvidenceHash = "sha256:" + strings.Repeat("0", 64)
	if err := db.CompleteRecoveryAttempt(ctx, mismatched); err == nil {
		t.Fatal("mismatched planner evidence was persisted")
	}
	if err := db.RejectRecoveryAttempt(ctx, request.RecoveryBudgetKey, request.UnitAttempt, request.ClaimToken,
		request.ControllerOwner, request.ControllerEpoch, recovery.ActionBlock, "mismatched evidence", request.Now); err != nil {
		t.Fatal(err)
	}
	silent := request
	silent.UnitAttempt = 2
	silent.ClaimToken = strings.Repeat("6", 32)
	silent.FailureClass = string(recovery.FailureSilentTool)
	forbiddenExhaustion := silent
	forbiddenExhaustion.ExhaustedAction = recovery.ActionFinalHumanGate
	if _, err := db.ReserveRecoveryAttempt(ctx, forbiddenExhaustion); err == nil {
		t.Fatal("class-forbidden exhaustion action was accepted")
	}
	if _, err := db.ReserveRecoveryAttempt(ctx, silent); err != nil {
		t.Fatal(err)
	}
	forbidden := recoveryOutcomeFixture(silent)
	forbidden.SelectedAction = recovery.ActionRetrySameUnit
	if err := db.CompleteRecoveryAttempt(ctx, forbidden); err == nil {
		t.Fatal("class-forbidden planner action was persisted")
	}
}

func TestRecoveryCompletionIsFencedAfterControllerOwnershipChanges(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	if _, err := db.ReserveRecoveryAttempt(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, request.DeliveryID, request.ControllerOwner, domain.RunReady); err != nil {
		t.Fatal(err)
	}
	if err := db.ReleaseLease(ctx, Lease{RunID: request.DeliveryID, Owner: request.ControllerOwner, Epoch: request.ControllerEpoch}); err != nil {
		t.Fatal(err)
	}
	replacementLease, err := db.AcquireLease(ctx, request.DeliveryID, "replacement-controller", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, request.DeliveryID, "replacement-controller"); err != nil {
		t.Fatal(err)
	}
	if err := db.CompleteRecoveryAttempt(ctx, recoveryOutcomeFixture(request)); !errors.Is(err, ErrRecoveryClaimFenced) {
		t.Fatalf("stale controller completion err=%v", err)
	}
	replacement := request
	replacement.UnitAttempt = 2
	replacement.ClaimToken = strings.Repeat("9", 32)
	replacement.ControllerOwner = "replacement-controller"
	replacement.ControllerEpoch = replacementLease.Epoch
	if _, err := db.ReserveRecoveryAttempt(ctx, replacement); err != nil {
		t.Fatalf("replacement controller could not recover orphaned reservation: %v", err)
	}
}

func TestRecoveryBudgetPersistsStructuredEvidenceBackoffAndExhaustion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "shepherd.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	first, err := db.ReserveRecoveryAttempt(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if first.Attempts != 1 || first.Backoff != time.Minute || !first.NextRetryAt.Equal(request.Now.Add(time.Minute)) {
		t.Fatalf("first reservation=%+v", first)
	}
	outcome := RecoveryOutcome{
		RecoveryBudgetKey:         request.RecoveryBudgetKey,
		UnitAttempt:               request.UnitAttempt,
		ClaimToken:                request.ClaimToken,
		ControllerOwner:           request.ControllerOwner,
		ControllerEpoch:           request.ControllerEpoch,
		PlannerRequestNonce:       strings.Repeat("1", 32),
		EvidenceHash:              request.FailureHash,
		AuthorityScopeHash:        "sha256:" + strings.Repeat("4", 64),
		PlannerEvidenceHash:       "sha256:" + strings.Repeat("2", 64),
		PlannerSessionID:          "019f5d4a-9fb4-7852-b640-d6fdf71bd3d9",
		PlannerSessionFingerprint: "sha256:" + strings.Repeat("3", 64),
		ObservedModel:             recovery.RequiredModel,
		Thinking:                  recovery.RequiredThinking,
		SelectedAction:            recovery.ActionRetryAfterBackoff,
		BoundedPlan:               []recovery.PlanStep{{Primitive: recovery.PrimitiveRetryFreshAttempt}},
		IssuedAt:                  request.Now,
		ExpiresAt:                 request.Now.Add(5 * time.Minute),
	}
	if err := db.CompleteRecoveryAttempt(ctx, outcome); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	stored, err := db.GetRecoveryAttempt(ctx, request.RecoveryBudgetKey, request.UnitAttempt)
	if err != nil {
		t.Fatal(err)
	}
	if stored.SelectedAction != outcome.SelectedAction || len(stored.BoundedPlan) != 1 || stored.PlannerSessionID != outcome.PlannerSessionID || stored.PlannerEvidenceHash != outcome.PlannerEvidenceHash || stored.EvidenceHash != outcome.EvidenceHash || stored.AuthorityScopeHash != outcome.AuthorityScopeHash {
		t.Fatalf("stored outcome=%+v", stored)
	}
	request.UnitAttempt = 2
	request.ClaimToken = strings.Repeat("b", 32)
	request.Now = request.Now.Add(time.Minute)
	second, err := db.ReserveRecoveryAttempt(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if second.Attempts != 2 || second.Backoff != 2*time.Minute || !second.NextRetryAt.Equal(request.Now.Add(2*time.Minute)) {
		t.Fatalf("second reservation=%+v", second)
	}
	if err := db.RejectRecoveryAttempt(ctx, request.RecoveryBudgetKey, request.UnitAttempt, request.ClaimToken, request.ControllerOwner, request.ControllerEpoch, recovery.ActionBlock, "planner evidence rejected", request.Now); err != nil {
		t.Fatal(err)
	}
	request.UnitAttempt = 3
	request.ClaimToken = strings.Repeat("c", 32)
	request.Now = request.Now.Add(time.Minute)
	exhausted, err := db.ReserveRecoveryAttempt(ctx, request)
	if !errors.Is(err, ErrRetryBudgetExhausted) || exhausted.ExhaustedAt.IsZero() || exhausted.SelectedAction != recovery.ActionAwaitDecision {
		t.Fatalf("exhausted=%+v err=%v", exhausted, err)
	}
	after, err := db.GetRecoveryBudget(ctx, request.RecoveryBudgetKey)
	if err != nil {
		t.Fatal(err)
	}
	if after.Attempts != 2 || after.ExhaustedAt.IsZero() || after.SelectedAction != recovery.ActionAwaitDecision {
		t.Fatalf("persisted exhausted budget=%+v", after)
	}
}

func TestRecoveryBudgetIsIndependentByFailureClassAndPolicyCannotChange(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	artifact := recoveryReservationFixture()
	if _, err := db.ReserveRecoveryAttempt(ctx, artifact); err != nil {
		t.Fatal(err)
	}
	interrupted := artifact
	interrupted.FailureClass = string(recovery.FailureInterrupted)
	interrupted.ClaimToken = strings.Repeat("d", 32)
	if budget, err := db.ReserveRecoveryAttempt(ctx, interrupted); err != nil || budget.Attempts != 1 {
		t.Fatalf("independent class budget=%+v err=%v", budget, err)
	}
	changed := artifact
	changed.UnitAttempt = 2
	changed.ClaimToken = strings.Repeat("e", 32)
	changed.MaxAttempts++
	if _, err := db.ReserveRecoveryAttempt(ctx, changed); err == nil {
		t.Fatal("recovery policy changed inside existing key")
	}
}

func TestRecoveryDispatchUsesLatestGlobalDecisionAcrossFailureClasses(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	first := recoveryReservationFixture()
	if _, err := db.ReserveRecoveryAttempt(ctx, first); err != nil {
		t.Fatal(err)
	}
	if err := db.CompleteRecoveryAttempt(ctx, recoveryOutcomeFixture(first)); err != nil {
		t.Fatal(err)
	}
	second := first
	second.FailureClass = string(recovery.FailureInterrupted)
	second.ClaimToken = strings.Repeat("7", 32)
	if _, err := db.ReserveRecoveryAttempt(ctx, second); err != nil {
		t.Fatal(err)
	}
	secondOutcome := recoveryOutcomeFixture(second)
	secondOutcome.PlannerRequestNonce = strings.Repeat("8", 32)
	secondOutcome.PlannerEvidenceHash = "sha256:" + strings.Repeat("8", 64)
	if err := db.CompleteRecoveryAttempt(ctx, secondOutcome); err != nil {
		t.Fatal(err)
	}
	claimed, err := db.ClaimRecoveryDispatch(ctx, second.DeliveryID, second.Generation, second.UnitID,
		second.HeadSHA, second.ControllerOwner, second.ControllerEpoch, second.Now.Add(time.Minute))
	if err != nil || claimed.FailureClass != second.FailureClass || claimed.Sequence <= 1 {
		t.Fatalf("latest cross-class dispatch=%+v err=%v", claimed, err)
	}
}

func TestFailedUnitWithoutRecoveryDecisionCannotRedispatchAfterCrash(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	unitKey := UnitAttemptKey{DeliveryID: request.DeliveryID, Generation: request.Generation,
		UnitID: request.UnitID, HeadSHA: request.HeadSHA}
	unitAttempt, err := db.BeginUnitAttempt(ctx, unitKey, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.FinishUnitAttemptFenced(ctx, unitKey, unitAttempt.Attempts, string(recovery.FailureDeadWorker),
		request.ControllerOwner, request.ControllerEpoch); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, request.ControllerOwner, request.ControllerEpoch, request.Now); !errors.Is(err, ErrRecoveryDecisionPending) {
		t.Fatalf("failed unit without recovery decision dispatch err=%v", err)
	}
}

func TestMutatingSkipDoesNotCreateMissingRecoveryDecision(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	unitKey := UnitAttemptKey{DeliveryID: request.DeliveryID, Generation: request.Generation,
		UnitID: request.UnitID, HeadSHA: request.HeadSHA}
	unitAttempt, err := db.BeginUnitAttempt(ctx, unitKey, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.FinishUnitAttemptFenced(ctx, unitKey, unitAttempt.Attempts, "mutating_skip",
		request.ControllerOwner, request.ControllerEpoch); err != nil {
		t.Fatal(err)
	}
	if attempt, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, request.ControllerOwner, request.ControllerEpoch, request.Now); err != nil || attempt.SelectedAction != "" {
		t.Fatalf("mutating skip recovery dispatch=%+v err=%v", attempt, err)
	}
	later, err := db.BeginUnitAttempt(ctx, unitKey, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.FinishUnitAttemptFenced(ctx, unitKey, later.Attempts, string(recovery.FailureDeadWorker),
		request.ControllerOwner, request.ControllerEpoch); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, request.ControllerOwner, request.ControllerEpoch, request.Now); !errors.Is(err, ErrRecoveryDecisionPending) {
		t.Fatalf("historical consumed decision masked later failure: %v", err)
	}
}

func TestRecoveryDispatchBlocksIncompleteOrTerminalDecision(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	if _, err := db.ReserveRecoveryAttempt(ctx, request); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, request.ControllerOwner, request.ControllerEpoch, request.Now); !errors.Is(err, ErrRecoveryDecisionPending) {
		t.Fatalf("incomplete recovery decision dispatch err=%v", err)
	}
	if err := db.CompleteRecoveryDecision(ctx, request.RecoveryBudgetKey, request.UnitAttempt,
		request.ClaimToken, request.ControllerOwner, request.ControllerEpoch, recovery.ActionBlock, request.Now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, request.ControllerOwner, request.ControllerEpoch, request.Now); !errors.Is(err, ErrRecoveryTerminal) {
		t.Fatalf("terminal recovery decision dispatch err=%v", err)
	}
}

func TestRecoveryDispatchHonorsPersistedBackoffAndClaimsOnce(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	request.FailureClass = string(recovery.FailureSilentTool)
	unitKey := UnitAttemptKey{DeliveryID: request.DeliveryID, Generation: request.Generation,
		UnitID: request.UnitID, HeadSHA: request.HeadSHA}
	initial, err := db.BeginUnitAttempt(ctx, unitKey, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.FinishUnitAttemptFenced(ctx, unitKey, initial.Attempts, string(recovery.FailureSilentTool),
		request.ControllerOwner, request.ControllerEpoch); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ReserveRecoveryAttempt(ctx, request); err != nil {
		t.Fatal(err)
	}
	outcome := recoveryOutcomeFixture(request)
	if err := db.CompleteRecoveryAttempt(ctx, outcome); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, request.DeliveryID, request.ControllerOwner, domain.RunReady); err != nil {
		t.Fatal(err)
	}
	if err := db.ReleaseLease(ctx, Lease{RunID: request.DeliveryID, Owner: request.ControllerOwner, Epoch: request.ControllerEpoch}); err != nil {
		t.Fatal(err)
	}
	nextLease, err := db.AcquireLease(ctx, request.DeliveryID, "next-execution", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, request.DeliveryID, "next-execution"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA, "next-execution", nextLease.Epoch, request.Now.Add(30*time.Second)); !errors.Is(err, ErrRecoveryBackoffPending) {
		t.Fatalf("early dispatch err=%v", err)
	}
	claimed, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA, "next-execution", nextLease.Epoch, request.Now.Add(time.Minute))
	if err != nil || claimed.DispatchedAt.IsZero() {
		t.Fatalf("claimed=%+v err=%v", claimed, err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA, "duplicate-execution", nextLease.Epoch, request.Now.Add(time.Minute)); !errors.Is(err, ErrRecoveryDispatchClaimed) {
		t.Fatalf("duplicate dispatch err=%v", err)
	}
	if _, err := db.db.ExecContext(ctx, `UPDATE delivery_runs SET state = ?, owner = '' WHERE delivery_id = ?`, domain.RunReady, request.DeliveryID); err != nil {
		t.Fatal(err)
	}
	if err := db.ReleaseLease(ctx, nextLease); err != nil {
		t.Fatal(err)
	}
	replacementLease, err := db.AcquireLease(ctx, request.DeliveryID, "replacement-execution", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, request.DeliveryID, "replacement-execution"); err != nil {
		t.Fatal(err)
	}
	reclaimed, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID, request.HeadSHA, "replacement-execution", replacementLease.Epoch, request.Now.Add(time.Minute))
	if err != nil || reclaimed.DispatchClaim != "replacement-execution" {
		t.Fatalf("reclaimed dispatch=%+v err=%v", reclaimed, err)
	}
	unitAttempt, err := db.BeginUnitAttempt(ctx, unitKey, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.FinishUnitAttemptFenced(ctx, unitKey, unitAttempt.Attempts, "success", "next-execution", nextLease.Epoch); err == nil {
		t.Fatal("stale controller disposed replacement recovery dispatch")
	}
	if err := db.FinishUnitAttemptFenced(ctx, unitKey, unitAttempt.Attempts, "success", "replacement-execution", replacementLease.Epoch); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, "replacement-execution", replacementLease.Epoch, request.Now.Add(time.Minute)); err != nil {
		t.Fatalf("consumed dispatch did not close cleanly: %v", err)
	}
	later, err := db.BeginUnitAttempt(ctx, unitKey, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.FinishUnitAttemptFenced(ctx, unitKey, later.Attempts, string(recovery.FailureDeadWorker),
		"replacement-execution", replacementLease.Epoch); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, "replacement-execution", replacementLease.Epoch, request.Now.Add(time.Minute)); !errors.Is(err, ErrRecoveryDecisionPending) {
		t.Fatalf("historical consumed dispatch masked later failure: %v", err)
	}
}

func TestUsedRecoveryDispatchCannotReplayAfterItsFreshAttemptFails(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	if _, err := db.ReserveRecoveryAttempt(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := db.CompleteRecoveryAttempt(ctx, recoveryOutcomeFixture(request)); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, request.DeliveryID, request.ControllerOwner, domain.RunReady); err != nil {
		t.Fatal(err)
	}
	if err := db.ReleaseLease(ctx, Lease{RunID: request.DeliveryID, Owner: request.ControllerOwner, Epoch: request.ControllerEpoch}); err != nil {
		t.Fatal(err)
	}
	nextLease, err := db.AcquireLease(ctx, request.DeliveryID, "next-execution", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, request.DeliveryID, "next-execution"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, "next-execution", nextLease.Epoch, request.Now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	unitKey := UnitAttemptKey{DeliveryID: request.DeliveryID, Generation: request.Generation,
		UnitID: request.UnitID, HeadSHA: request.HeadSHA}
	unitAttempt, err := db.BeginUnitAttempt(ctx, unitKey, 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.FinishUnitAttemptFenced(ctx, unitKey, unitAttempt.Attempts, string(recovery.FailureDeadWorker),
		"next-execution", nextLease.Epoch); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, "next-execution", nextLease.Epoch, request.Now.Add(time.Minute)); !errors.Is(err, ErrRecoveryDecisionPending) {
		t.Fatalf("used failed dispatch replay err=%v", err)
	}
}

func TestCrashReconciliationDisposesClaimedRecoveryDispatch(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	if _, err := db.ReserveRecoveryAttempt(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := db.CompleteRecoveryAttempt(ctx, recoveryOutcomeFixture(request)); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, request.DeliveryID, request.ControllerOwner, domain.RunReady); err != nil {
		t.Fatal(err)
	}
	if err := db.ReleaseLease(ctx, Lease{RunID: request.DeliveryID, Owner: request.ControllerOwner, Epoch: request.ControllerEpoch}); err != nil {
		t.Fatal(err)
	}
	crashedLease, err := db.AcquireLease(ctx, request.DeliveryID, "crashed-execution", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, request.DeliveryID, crashedLease.Owner); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, crashedLease.Owner, crashedLease.Epoch, request.Now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	unitKey := UnitAttemptKey{DeliveryID: request.DeliveryID, Generation: request.Generation,
		UnitID: request.UnitID, HeadSHA: request.HeadSHA}
	if _, err := db.BeginUnitAttempt(ctx, unitKey, 10); err != nil {
		t.Fatal(err)
	}
	recoveryLease, err := db.AcquireReconciliationLease(ctx, request.DeliveryID, "restart-execution", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.ReconcileInterruptedDelivery(ctx, recoveryLease, domain.RunReady); err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, request.DeliveryID, recoveryLease.Owner); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, recoveryLease.Owner, recoveryLease.Epoch, request.Now.Add(time.Minute)); !errors.Is(err, ErrRecoveryDecisionPending) {
		t.Fatalf("reconciled claimed dispatch replay err=%v", err)
	}
}

func TestRecoveryDispatchExpiresPlannerEvidenceIntoAwaitingDecision(t *testing.T) {
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
	startRecoveryController(t, ctx, db)
	request := recoveryReservationFixture()
	if _, err := db.ReserveRecoveryAttempt(ctx, request); err != nil {
		t.Fatal(err)
	}
	if err := db.CompleteRecoveryAttempt(ctx, recoveryOutcomeFixture(request)); err != nil {
		t.Fatal(err)
	}
	if err := db.FinishAttempt(ctx, request.DeliveryID, request.ControllerOwner, domain.RunReady); err != nil {
		t.Fatal(err)
	}
	if err := db.ReleaseLease(ctx, Lease{RunID: request.DeliveryID, Owner: request.ControllerOwner, Epoch: request.ControllerEpoch}); err != nil {
		t.Fatal(err)
	}
	nextLease, err := db.AcquireLease(ctx, request.DeliveryID, "next-execution", time.Now().UTC(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, request.DeliveryID, "next-execution"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ClaimRecoveryDispatch(ctx, request.DeliveryID, request.Generation, request.UnitID,
		request.HeadSHA, "next-execution", nextLease.Epoch, request.Now.Add(6*time.Minute)); !errors.Is(err, ErrRetryBudgetExhausted) {
		t.Fatalf("expired planner evidence dispatch err=%v", err)
	}
	budget, err := db.GetRecoveryBudget(ctx, request.RecoveryBudgetKey)
	if err != nil || budget.Status != "exhausted" || budget.SelectedAction != recovery.ActionAwaitDecision || budget.ExhaustedAt.IsZero() {
		t.Fatalf("expired recovery budget=%+v err=%v", budget, err)
	}
}

func TestLegacyRecoveryBudgetIsDurablyExhaustedOnMigration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "authority.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.EnsureDelivery(ctx, testDelivery("issue-389", 389)); err != nil {
		t.Fatal(err)
	}
	key := recoveryReservationFixture().RecoveryBudgetKey
	if _, err := db.db.ExecContext(ctx, `INSERT INTO recovery_budgets
		(delivery_id, generation, unit_id, head_sha, failure_class, attempts, max_attempts, backoff_ms,
		last_failure, recovery_plan, next_retry_at, exhausted_at, policy_version, updated_at)
		VALUES (?, ?, ?, ?, ?, 1, 3, 1000, 'legacy', 'static text', 1, 0, 0, 1)`, key.DeliveryID,
		key.Generation, key.UnitID, key.HeadSHA, key.FailureClass); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	budget, err := db.GetRecoveryBudget(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if budget.Status != "exhausted" || budget.SelectedAction != recovery.ActionAwaitDecision || budget.ExhaustedAt.IsZero() || budget.MaxAttempts != budget.Attempts {
		t.Fatalf("migrated legacy recovery budget=%+v", budget)
	}
	startRecoveryController(t, ctx, db)
	fixture := recoveryReservationFixture()
	if _, err := db.ClaimRecoveryDispatch(ctx, fixture.DeliveryID, fixture.Generation, fixture.UnitID,
		fixture.HeadSHA, fixture.ControllerOwner, fixture.ControllerEpoch, fixture.Now); !errors.Is(err, ErrRetryBudgetExhausted) {
		t.Fatalf("migrated legacy budget allowed dispatch: %v", err)
	}
	if _, err := db.ReserveRecoveryAttempt(ctx, fixture); !errors.Is(err, ErrRetryBudgetExhausted) {
		t.Fatalf("migrated legacy budget did not remain exhausted: %v", err)
	}
}

func TestStaticRecoveryTextIsRejected(t *testing.T) {
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
	key := recoveryReservationFixture().RecoveryBudgetKey
	if _, err := db.BeginRecoveryAttempt(ctx, key, 2, time.Second, "failure", "Sol/high recovery planning required before next retry", time.Now().UTC()); err == nil {
		t.Fatal("static recovery sentence was accepted as a structured plan")
	}
}

func startRecoveryController(t *testing.T, ctx context.Context, db *Store) {
	t.Helper()
	if _, err := db.AcquireLease(ctx, "issue-389", "recovery-controller", time.Now().UTC(), time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := db.BeginAttempt(ctx, "issue-389", "recovery-controller"); err != nil {
		t.Fatal(err)
	}
}

func recoveryReservationFixture() RecoveryReservation {
	return RecoveryReservation{
		RecoveryBudgetKey: RecoveryBudgetKey{
			DeliveryID:   "issue-389",
			Generation:   1,
			UnitID:       "execute-task/M001/S01/T01",
			HeadSHA:      strings.Repeat("a", 40),
			FailureClass: string(recovery.FailureArtifactMissing),
		},
		UnitAttempt:     1,
		ClaimToken:      strings.Repeat("a", 32),
		ControllerOwner: "recovery-controller",
		ControllerEpoch: 1,
		PolicyVersion:   recovery.PolicyVersion,
		MaxAttempts:     2,
		BaseBackoff:     time.Minute,
		MaxBackoff:      5 * time.Minute,
		FailureHash:     "sha256:" + strings.Repeat("f", 64),
		Diagnostic:      "artifact_missing",
		Reversible:      true,
		ExhaustedAction: recovery.ActionAwaitDecision,
		Now:             time.Unix(1_700_000_000, 0).UTC(),
	}
}

func recoveryOutcomeFixture(request RecoveryReservation) RecoveryOutcome {
	return RecoveryOutcome{
		RecoveryBudgetKey:         request.RecoveryBudgetKey,
		UnitAttempt:               request.UnitAttempt,
		ClaimToken:                request.ClaimToken,
		ControllerOwner:           request.ControllerOwner,
		ControllerEpoch:           request.ControllerEpoch,
		PlannerRequestNonce:       strings.Repeat("1", 32),
		EvidenceHash:              request.FailureHash,
		AuthorityScopeHash:        "sha256:" + strings.Repeat("4", 64),
		PlannerEvidenceHash:       "sha256:" + strings.Repeat("2", 64),
		PlannerSessionID:          "019f5d4a-9fb4-7852-b640-d6fdf71bd3d9",
		PlannerSessionFingerprint: "sha256:" + strings.Repeat("3", 64),
		ObservedModel:             recovery.RequiredModel,
		Thinking:                  recovery.RequiredThinking,
		SelectedAction:            recovery.ActionRetryAfterBackoff,
		BoundedPlan:               []recovery.PlanStep{{Primitive: recovery.PrimitiveRetryFreshAttempt}},
		IssuedAt:                  request.Now,
		ExpiresAt:                 request.Now.Add(5 * time.Minute),
	}
}
