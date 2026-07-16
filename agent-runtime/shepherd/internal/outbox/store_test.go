package outbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
	authoritystore "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
)

func TestStorePersistsImmutableEffectAndEnforcesFencedStateMachine(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "outbox.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Unix(1_700_000_000, 0).UTC()
	controller, authorized := testAuthorization(t, 4, 9, now)
	record, inserted, err := db.Enqueue(ctx, authorized, now)
	if err != nil || !inserted {
		t.Fatalf("enqueue inserted=%t err=%v", inserted, err)
	}
	duplicate, inserted, err := db.Enqueue(ctx, authorized, now.Add(time.Second))
	if err != nil || inserted || duplicate.EffectID != record.EffectID || duplicate.PayloadHash != record.PayloadHash {
		t.Fatalf("idempotent enqueue record=%+v inserted=%t err=%v", duplicate, inserted, err)
	}
	conflictingIntent, err := NewSummaryIntent(record.Target, record.DeliveryID, record.Generation,
		record.HeadSHA, record.Revision, summaryHash("different payload"), "different payload")
	if err != nil {
		t.Fatal(err)
	}
	conflictingAuthorization, err := controller.Authorize(ctx, conflictingIntent, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if conflictingAuthorization.EffectID() != authorized.EffectID() {
		t.Fatal("semantic idempotency identity changed with payload bytes")
	}
	if _, _, err := db.Enqueue(ctx, conflictingAuthorization, now.Add(time.Second)); err == nil {
		t.Fatal("same semantic idempotency key accepted different payload bytes")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	reopened, err := db.Get(ctx, record.EffectID)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.State != StatePending || reopened.PayloadHash != record.PayloadHash ||
		string(reopened.Payload) != string(record.Payload) || reopened.IdempotencyKey != record.IdempotencyKey {
		t.Fatalf("reopen changed effect: before=%+v after=%+v", record, reopened)
	}

	claim, err := db.Claim(ctx, authorized, authorized.Owner(), authorized.Epoch(), now.Add(2*time.Second), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Claim(ctx, authorized, authorized.Owner(), authorized.Epoch(), now.Add(3*time.Second), time.Minute); err == nil {
		t.Fatal("effect was claimed twice")
	}
	if err := db.StartExecution(ctx, claim, now.Add(4*time.Second)); err != nil {
		t.Fatal(err)
	}
	stale := claim
	stale.ControllerEpoch++
	if err := db.MarkSent(ctx, stale, Result{Code: ResultSent, ExternalID: 77, ExternalActor: "shepherd-bot"}, now.Add(5*time.Second)); !errors.Is(err, ErrFenced) {
		t.Fatalf("stale claim result error=%v, want fenced", err)
	}
	if err := db.MarkSent(ctx, claim, Result{Code: ResultSent, ExternalID: 77, ExternalActor: "shepherd-bot"}, now.Add(5*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := db.MarkUncertain(ctx, claim, ErrorPostSendAmbiguous, now.Add(6*time.Second)); !errors.Is(err, ErrTerminal) {
		t.Fatalf("sent effect changed terminal state: %v", err)
	}
	terminal, err := db.Get(ctx, record.EffectID)
	if err != nil || terminal.State != StateSent || terminal.Result.ExternalID != 77 {
		t.Fatalf("terminal=%+v err=%v", terminal, err)
	}
	events, err := db.Events(ctx, record.EffectID)
	if err != nil {
		t.Fatal(err)
	}
	wantEvents := []EventKind{EventRequested, EventAuthorized, EventEnqueued, EventClaimed, EventExecutionStarted, EventSent}
	if len(events) != len(wantEvents) {
		t.Fatalf("events=%+v, want exact lifecycle %v", events, wantEvents)
	}
	for index, required := range wantEvents {
		if events[index].Kind != required || events[index].Sequence != int64(index+1) {
			t.Fatalf("event[%d]=%+v, want kind=%s sequence=%d", index, events[index], required, index+1)
		}
	}

	otherController, _ := newTestController(t, ControllerFacts{
		DeliveryID: "issue-389", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390,
		Generation: 4, HeadSHA: strings.Repeat("a", 40), Owner: "controller-2",
	}, now.Add(7*time.Second))
	if _, err := otherController.Reauthorize(ctx, reopened, now.Add(7*time.Second)); err != nil {
		t.Fatalf("immutable effect could not receive a new fenced grant: %v", err)
	}
}

func TestStoreSerializesSummarySlotAndCancelsOnlyUnstartedEffects(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "outbox.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Unix(1_700_000_000, 0).UTC()
	controller, older := testAuthorization(t, 4, 9, now)
	if _, _, err := database.Enqueue(ctx, older, now); err != nil {
		t.Fatal(err)
	}
	claim, err := database.Claim(ctx, older, older.Owner(), older.Epoch(), now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	newerIntent, err := NewSummaryIntent(Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390},
		"issue-389", 4, strings.Repeat("a", 40), 2, summaryHash("newer"), "newer")
	if err != nil {
		t.Fatal(err)
	}
	newer, err := controller.Authorize(ctx, newerIntent, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := database.Enqueue(ctx, newer, now.Add(time.Second)); !errors.Is(err, ErrFenced) {
		t.Fatalf("newer summary was durably enqueued during an active slot write: %v", err)
	}
	if err := database.StartExecution(ctx, claim, now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := database.MarkSent(ctx, claim, Result{Code: ResultSent, ExternalID: 42,
		ExternalActor: "shepherd-bot"}, now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := database.Enqueue(ctx, newer, now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := database.Cancel(ctx, newer, now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	cancelled, err := database.Get(ctx, newer.EffectID())
	if err != nil || cancelled.State != StateCancelled {
		t.Fatalf("cancelled=%+v err=%v", cancelled, err)
	}
	if _, err := database.Claim(ctx, newer, newer.Owner(), newer.Epoch(), now.Add(4*time.Second), time.Minute); !errors.Is(err, ErrTerminal) {
		t.Fatalf("cancelled effect was claimable: %v", err)
	}
}

func TestStoreFailedOlderSummaryIsSupersededWithoutRetry(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "failed-summary.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Now().UTC()
	controller, older := testAuthorization(t, 4, 9, now)
	if _, _, err := database.Enqueue(ctx, older, now); err != nil {
		t.Fatal(err)
	}
	olderClaim, err := database.Claim(ctx, older, older.Owner(), older.Epoch(), now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.MarkFailed(ctx, olderClaim, ErrorPreSend, now); err != nil {
		t.Fatal(err)
	}
	newerIntent, err := NewSummaryIntent(Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390},
		"issue-389", 4, strings.Repeat("a", 40), 2, summaryHash("newer sent"), "newer sent")
	if err != nil {
		t.Fatal(err)
	}
	newer, err := controller.Authorize(ctx, newerIntent, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := database.Enqueue(ctx, newer, now); err != nil {
		t.Fatal(err)
	}
	newerClaim, err := database.Claim(ctx, newer, newer.Owner(), newer.Epoch(), now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.StartExecution(ctx, newerClaim, now); err != nil {
		t.Fatal(err)
	}
	if err := database.MarkSent(ctx, newerClaim, Result{Code: ResultSent, ExternalID: 42,
		ExternalActor: "shepherd-bot"}, now); err != nil {
		t.Fatal(err)
	}
	if err := database.RetryFailed(ctx, older, now.Add(time.Second)); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("older failed summary retry=%v", err)
	}
	superseded, err := database.Get(ctx, older.EffectID())
	if err != nil || superseded.State != StateCancelled || superseded.RetryCount != 0 {
		t.Fatalf("superseded failed summary=%+v err=%v", superseded, err)
	}
}

func TestStoreUncertainSummaryFencesNewerRevision(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "uncertain-summary.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Now().UTC()
	controller, older := testAuthorization(t, 4, 9, now)
	if _, _, err := database.Enqueue(ctx, older, now); err != nil {
		t.Fatal(err)
	}
	claim, err := database.Claim(ctx, older, older.Owner(), older.Epoch(), now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.StartExecution(ctx, claim, now); err != nil {
		t.Fatal(err)
	}
	if err := database.MarkUncertain(ctx, claim, ErrorPostSendAmbiguous, now); err != nil {
		t.Fatal(err)
	}
	newerIntent, err := NewSummaryIntent(Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390},
		"issue-389", 4, strings.Repeat("a", 40), 2, summaryHash("newer"), "newer")
	if err != nil {
		t.Fatal(err)
	}
	newer, err := controller.Authorize(ctx, newerIntent, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := database.Enqueue(ctx, newer, now); !errors.Is(err, ErrFenced) {
		t.Fatalf("new summary bypassed unresolved uncertainty: %v", err)
	}
}

func TestStoreHigherBlockedSummaryStillFencesOlderPendingRevision(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "blocked-summary.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Now().UTC()
	controller, older := testAuthorization(t, 4, 9, now)
	newerIntent, err := NewSummaryIntent(Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390},
		"issue-389", 4, strings.Repeat("a", 40), 2, summaryHash("newer blocked"), "newer blocked")
	if err != nil {
		t.Fatal(err)
	}
	newer, err := controller.Authorize(ctx, newerIntent, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := database.Enqueue(ctx, newer, now); err != nil {
		t.Fatal(err)
	}
	claim, err := database.Claim(ctx, newer, newer.Owner(), newer.Epoch(), now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.MarkBlocked(ctx, claim, ErrorRetryExhausted, now); err != nil {
		t.Fatal(err)
	}
	if _, _, err := database.Enqueue(ctx, older, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Claim(ctx, older, older.Owner(), older.Epoch(), now.Add(time.Second),
		time.Minute); !errors.Is(err, ErrStaleRevision) {
		t.Fatalf("older summary bypassed higher blocked revision: %v", err)
	}
	superseded, err := database.Get(ctx, older.EffectID())
	if err != nil || superseded.State != StateCancelled || superseded.ErrorCode != ErrorStaleRevision {
		t.Fatalf("superseded summary=%+v err=%v", superseded, err)
	}
}

func TestStoreReconcilesStaleUnstartedAndStartedEffectsWithoutOldAuthority(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "outbox.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Now().UTC()
	_, pendingAuthorization := testAuthorization(t, 1, 1, now)
	pending, _, err := database.Enqueue(ctx, pendingAuthorization, now)
	if err != nil {
		t.Fatal(err)
	}
	currentController, _ := newTestController(t, ControllerFacts{
		DeliveryID: "issue-389", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390,
		Generation: 2, HeadSHA: strings.Repeat("c", 40), Owner: "current-controller",
	}, now.Add(2*time.Second))
	if err := database.ReconcileStale(ctx, currentController, now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	cancelled, err := database.Get(ctx, pending.EffectID)
	if err != nil || cancelled.State != StateCancelled || cancelled.ErrorCode != ErrorChangedTarget {
		t.Fatalf("cancelled stale effect=%+v err=%v", cancelled, err)
	}

	startedController, _ := testAuthorization(t, 1, 1, now.Add(3*time.Second))
	blockedIntent, err := NewSummaryIntent(Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390},
		"issue-389", 1, strings.Repeat("a", 40), 2, summaryHash("blocked"), "blocked")
	if err != nil {
		t.Fatal(err)
	}
	blockedAuthorization, err := startedController.Authorize(ctx, blockedIntent, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	blocked, _, err := database.Enqueue(ctx, blockedAuthorization, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	blockedClaim, err := database.Claim(ctx, blockedAuthorization, blockedAuthorization.Owner(),
		blockedAuthorization.Epoch(), now.Add(3*time.Second), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.MarkBlocked(ctx, blockedClaim, ErrorRetryExhausted, now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := database.ReconcileStale(ctx, currentController, now.Add(4*time.Second)); err != nil {
		t.Fatalf("definitely-unsent stale blocker blocked current authority: %v", err)
	}
	blocked, err = database.Get(ctx, blocked.EffectID)
	if err != nil || blocked.State != StateBlocked {
		t.Fatalf("stale blocked effect=%+v err=%v", blocked, err)
	}

	startedIntent, err := NewSummaryIntent(Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390},
		"issue-389", 1, strings.Repeat("a", 40), 3, summaryHash("started"), "started")
	if err != nil {
		t.Fatal(err)
	}
	startedAuthorization, err := startedController.Authorize(ctx, startedIntent, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	started, _, err := database.Enqueue(ctx, startedAuthorization, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	claim, err := database.Claim(ctx, startedAuthorization, startedAuthorization.Owner(),
		startedAuthorization.Epoch(), now.Add(3*time.Second), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.StartExecution(ctx, claim, now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := database.ReconcileStale(ctx, currentController, now.Add(5*time.Second)); !errors.Is(err, ErrStaleUncertain) {
		t.Fatalf("stale started reconciliation error=%v", err)
	}
	uncertain, err := database.Get(ctx, started.EffectID)
	if err != nil || uncertain.State != StateUncertain || uncertain.ErrorCode != ErrorChangedTarget {
		t.Fatalf("uncertain stale effect=%+v err=%v", uncertain, err)
	}
}

func TestStoreRejectsUnsafeExecutionPhaseTransitions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	database, err := Open(ctx, filepath.Join(t.TempDir(), "transition-matrix.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	now := time.Now().UTC()
	_, authorization := testAuthorization(t, 2, 5, now)
	if _, _, err := database.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	claim, err := database.Claim(ctx, authorization, authorization.Owner(), authorization.Epoch(), now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.MarkUncertain(ctx, claim, ErrorPostSendAmbiguous, now); err == nil {
		t.Fatal("unstarted claim became uncertain")
	}
	if err := database.StartExecution(ctx, claim, now); err != nil {
		t.Fatal(err)
	}
	if err := database.MarkBlocked(ctx, claim, ErrorMarkerConflict, now); err == nil {
		t.Fatal("execution-started claim became definitely blocked")
	}
	if err := database.MarkUncertain(ctx, claim, ErrorPostSendAmbiguous, now); err != nil {
		t.Fatal(err)
	}
}

func TestStoreRecoversOnlyExpiredUnstartedClaimAndNeverReplaysUncertain(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "outbox.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	now := time.Unix(1_700_000_000, 0).UTC()
	controller, authorized := testAuthorization(t, 2, 5, now)
	record, _, err := db.Enqueue(ctx, authorized, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Claim(ctx, authorized, authorized.Owner(), authorized.Epoch(), now, time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RecoverExpiredClaims(ctx, controller, now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	recovered, err := db.Get(ctx, record.EffectID)
	if err != nil || recovered.State != StatePending {
		t.Fatalf("unstarted claim recovery=%+v err=%v", recovered, err)
	}
	if events, err := db.Events(ctx, record.EffectID); err != nil || !hasEvent(events, EventClaimRecovered) {
		t.Fatalf("claim recovery event missing: events=%+v err=%v", events, err)
	}

	claim, err := db.Claim(ctx, authorized, authorized.Owner(), authorized.Epoch(), now.Add(3*time.Second), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.StartExecution(ctx, claim, now.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RecoverExpiredClaims(ctx, controller, now.Add(5*time.Second)); err != nil {
		t.Fatal(err)
	}
	uncertain, err := db.Get(ctx, record.EffectID)
	if err != nil || uncertain.State != StateUncertain || uncertain.ErrorCode != ErrorClaimExpired {
		t.Fatalf("started claim recovery=%+v err=%v", uncertain, err)
	}
	if _, err := db.Claim(ctx, authorized, authorized.Owner(), authorized.Epoch(), now.Add(6*time.Second), time.Second); !errors.Is(err, ErrTerminal) {
		t.Fatalf("uncertain effect was replayable: %v", err)
	}
}

func newTestController(t *testing.T, facts ControllerFacts, now time.Time) (*Controller, ControllerFacts) {
	t.Helper()
	authority, err := authoritystore.Open(context.Background(), filepath.Join(t.TempDir(), "authority.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = authority.Close() })
	root := t.TempDir()
	if err := authority.EnsureDelivery(context.Background(), authoritystore.Delivery{
		ID: facts.DeliveryID, Issue: facts.Issue, ParentIssue: facts.Issue,
		Repository: facts.Repository, PullRequest: facts.PullRequest, WorkDir: root,
		ContextHash: "sha256:" + strings.Repeat("a", 64), Branch: "test/" + facts.DeliveryID,
		BaseBranch: "main", GSDProjectRoot: root, InitialHead: facts.HeadSHA, GSDVersion: "1.11.0",
	}); err != nil {
		t.Fatal(err)
	}
	for generation := int64(1); generation < facts.Generation; generation++ {
		owner := fmt.Sprintf("generation-%d", generation)
		if _, err := authority.BeginAttempt(context.Background(), facts.DeliveryID, owner); err != nil {
			t.Fatal(err)
		}
		if err := authority.FinishAttempt(context.Background(), facts.DeliveryID, owner, domain.RunFailed); err != nil {
			t.Fatal(err)
		}
		if err := authority.ResumeDelivery(context.Background(), domain.HumanDecision{
			RunID: facts.DeliveryID, Generation: generation, ActorKind: domain.ActorHuman, Approved: true,
		}); err != nil {
			t.Fatal(err)
		}
	}
	lease, err := authority.AcquireLease(context.Background(), facts.DeliveryID, facts.Owner, now, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	facts.Epoch = lease.Epoch
	controller, err := NewController(context.Background(), authority, lease, facts, now)
	if err != nil {
		t.Fatal(err)
	}
	return controller, facts
}

func testAuthorization(t *testing.T, generation, _ int64, now time.Time) (*Controller, Authorization) {
	t.Helper()
	facts := ControllerFacts{
		DeliveryID: "issue-389", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390,
		Generation: generation, HeadSHA: strings.Repeat("a", 40), Owner: "controller-1",
	}
	controller, facts := newTestController(t, facts, now)
	intent, err := NewSummaryIntent(Target{Repository: facts.Repository, Issue: facts.Issue, PullRequest: facts.PullRequest},
		facts.DeliveryID, facts.Generation, facts.HeadSHA, 1, summaryHash("summary"), "summary")
	if err != nil {
		t.Fatal(err)
	}
	authorized, err := controller.Authorize(context.Background(), intent, now)
	if err != nil {
		t.Fatal(err)
	}
	return controller, authorized
}

func summaryHash(summary string) string {
	digest := sha256.Sum256([]byte(summary))
	return "sha256:" + hex.EncodeToString(digest[:])
}

func hasEvent(events []Event, kind EventKind) bool {
	for _, event := range events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}
