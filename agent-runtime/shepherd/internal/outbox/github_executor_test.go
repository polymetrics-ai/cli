package outbox

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var allowTestExecution ExecutionFence = func(_ context.Context, claim ClaimedEffect, _ time.Time) (time.Time, error) {
	return claim.ExpiresAt, nil
}

type fakeCommentPort struct {
	comments      []Comment
	actor         string
	listErr       error
	onList        func()
	writeErr      error
	writeCreates  bool
	createCalls   int
	updateCalls   int
	nextCommentID int64
}

func (f *fakeCommentPort) Actor(context.Context) (string, error) {
	if f.actor == "" {
		return "shepherd-bot", nil
	}
	return f.actor, nil
}

func (f *fakeCommentPort) ListComments(context.Context, Target) ([]Comment, error) {
	if f.onList != nil {
		f.onList()
	}
	comments := append([]Comment(nil), f.comments...)
	actor, _ := f.Actor(context.Background())
	for index := range comments {
		if comments[index].Author == "" {
			comments[index].Author, comments[index].AuthorType = actor, "Bot"
		}
	}
	return comments, f.listErr
}

func (f *fakeCommentPort) CreateComment(_ context.Context, _ Target, body string) (Comment, error) {
	f.createCalls++
	id := f.nextCommentID
	if id == 0 {
		id = 77
	}
	actor, _ := f.Actor(context.Background())
	comment := Comment{ID: id, Body: body, Author: actor, AuthorType: "Bot"}
	if f.writeErr == nil || f.writeCreates {
		f.comments = append(f.comments, comment)
	}
	return comment, f.writeErr
}

func (f *fakeCommentPort) UpdateComment(_ context.Context, _ Target, id int64, body string) (Comment, error) {
	f.updateCalls++
	actor, _ := f.Actor(context.Background())
	comment := Comment{ID: id, Body: body, Author: actor, AuthorType: "Bot"}
	for index := range f.comments {
		if f.comments[index].ID == id && (f.writeErr == nil || f.writeCreates) {
			f.comments[index] = comment
		}
	}
	return comment, f.writeErr
}

func TestGitHubDispatcherReconcilesExactMarkerBeforeWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	_, authorization := testAuthorization(t, 4, 9, now)
	port := &fakeCommentPort{nextCommentID: 77}

	firstDB, err := Open(ctx, filepath.Join(t.TempDir(), "first.db"))
	if err != nil {
		t.Fatal(err)
	}
	first, _, err := firstDB.Enqueue(ctx, authorization, now)
	if err != nil {
		t.Fatal(err)
	}
	firstDispatcher := NewGitHubDispatcher(firstDB, port, allowTestExecution)
	result, err := firstDispatcher.Dispatch(ctx, authorization, authorization.Owner(), authorization.Epoch(), now, time.Minute)
	if err != nil || result.Code != ResultSent || result.ExternalID != 77 || port.createCalls != 1 {
		t.Fatalf("first dispatch result=%+v create=%d err=%v", result, port.createCalls, err)
	}
	if err := firstDB.Close(); err != nil {
		t.Fatal(err)
	}

	secondDB, err := Open(ctx, filepath.Join(t.TempDir(), "second.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = secondDB.Close() })
	if _, _, err := secondDB.Enqueue(ctx, authorization, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	secondDispatcher := NewGitHubDispatcher(secondDB, port, allowTestExecution)
	result, err = secondDispatcher.Dispatch(ctx, authorization, authorization.Owner(), authorization.Epoch(), now.Add(time.Second), time.Minute)
	if err != nil || result.Code != ResultReconciled || result.ExternalID != 77 || port.createCalls != 1 {
		t.Fatalf("reconciled dispatch result=%+v create=%d err=%v", result, port.createCalls, err)
	}
	reconciled, err := secondDB.Get(ctx, first.EffectID)
	if err != nil || reconciled.State != StateSent {
		t.Fatalf("reconciled effect=%+v err=%v", reconciled, err)
	}
}

func TestGitHubSummaryUpdateRequiresExactSentHistoryAndAdvancesRevision(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	controller, firstAuthorization := testAuthorization(t, 4, 9, now)
	database, err := Open(ctx, filepath.Join(t.TempDir(), "summary-update.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, _, err := database.Enqueue(ctx, firstAuthorization, now); err != nil {
		t.Fatal(err)
	}
	port := &fakeCommentPort{nextCommentID: 73}
	dispatcher := NewGitHubDispatcher(database, port, allowTestExecution)
	first, err := dispatcher.Dispatch(ctx, firstAuthorization, firstAuthorization.Owner(), firstAuthorization.Epoch(), now, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	secondIntent, err := NewSummaryIntent(Target{Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390},
		"issue-389", 4, strings.Repeat("a", 40), 2, summaryHash("new summary"), "new summary")
	if err != nil {
		t.Fatal(err)
	}
	secondAuthorization, err := controller.Authorize(ctx, secondIntent, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := database.Enqueue(ctx, secondAuthorization, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	second, err := dispatcher.Dispatch(ctx, secondAuthorization, secondAuthorization.Owner(),
		secondAuthorization.Epoch(), now.Add(time.Second), time.Minute)
	if err != nil || first.ExternalID != second.ExternalID || port.createCalls != 1 || port.updateCalls != 1 ||
		len(port.comments) != 1 || !strings.Contains(port.comments[0].Body, "new summary") {
		t.Fatalf("first=%+v second=%+v creates=%d updates=%d comments=%+v err=%v",
			first, second, port.createCalls, port.updateCalls, port.comments, err)
	}
}

func TestGitHubQuestionExecutorUsesBoundMarkerAndRequestIdentity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	facts := ControllerFacts{
		DeliveryID: "issue-389", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390,
		Generation: 4, HeadSHA: strings.Repeat("a", 40), Owner: "controller-1",
	}
	controller, facts := newTestController(t, facts, now)
	intent, err := NewQuestionIntent(Target{Repository: facts.Repository, Issue: facts.Issue, PullRequest: facts.PullRequest}, Question{
		RequestID: "decision-1", DeliveryID: facts.DeliveryID, UnitID: "execute-task/M001/S01/T01",
		Generation: facts.Generation, HeadSHA: facts.HeadSHA, Evidence: "retry budget exhausted",
		Options: []string{"retry", "stop"}, RecommendedOption: "retry", SafeDefault: "stop",
		ExpiresAt: now.Add(time.Hour), Mention: "karthik-sivadas",
	})
	if err != nil {
		t.Fatal(err)
	}
	authorization, err := controller.Authorize(ctx, intent, now)
	if err != nil {
		t.Fatal(err)
	}
	database, err := Open(ctx, filepath.Join(t.TempDir(), "question.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, _, err := database.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	port := &fakeCommentPort{nextCommentID: 91}
	result, err := NewGitHubDispatcher(database, port, allowTestExecution).Dispatch(ctx, authorization,
		authorization.Owner(), authorization.Epoch(), now, time.Minute)
	if err != nil || result.ExternalID != 91 || len(port.comments) != 1 {
		t.Fatalf("result=%+v comments=%+v err=%v", result, port.comments, err)
	}
	body := port.comments[0].Body
	for _, required := range []string{"<!-- shepherd-effect:decision-question:issue-389:decision-1:",
		"@karthik-sivadas", "Generation: `4`", facts.HeadSHA, "/shepherd decide decision-1 <option>"} {
		if !strings.Contains(body, required) {
			t.Fatalf("question body missing %q: %s", required, body)
		}
	}
}

func TestExpiredStartedQuestionReconcilesExactRemoteIdentityWithoutReplay(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	controller, facts := newTestController(t, ControllerFacts{
		DeliveryID: "issue-389", Repository: "polymetrics-ai/cli", Issue: 389, PullRequest: 390,
		Generation: 4, HeadSHA: strings.Repeat("a", 40), Owner: "controller-1",
	}, now)
	intent, err := NewQuestionIntent(Target{Repository: facts.Repository, Issue: facts.Issue, PullRequest: facts.PullRequest}, Question{
		RequestID: "decision-expired", DeliveryID: facts.DeliveryID, UnitID: "execute-task/M001/S01/T01",
		Generation: facts.Generation, HeadSHA: facts.HeadSHA, Evidence: "retry budget exhausted",
		Options: []string{"retry", "stop"}, RecommendedOption: "retry", SafeDefault: "stop",
		ExpiresAt: now.Add(time.Second), Mention: "karthik-sivadas",
	})
	if err != nil {
		t.Fatal(err)
	}
	authorization, err := controller.Authorize(ctx, intent, now)
	if err != nil {
		t.Fatal(err)
	}
	database, err := Open(ctx, filepath.Join(t.TempDir(), "expired-question.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, _, err := database.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	claim, err := database.Claim(ctx, authorization, authorization.Owner(), authorization.Epoch(), now, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.StartExecution(ctx, claim, now); err != nil {
		t.Fatal(err)
	}
	port := &fakeCommentPort{actor: "shepherd-bot"}
	body, err := newGitHubExecutor(port).Body(authorization.Record())
	if err != nil {
		t.Fatal(err)
	}
	port.comments = []Comment{{ID: 91, Body: body, Author: "shepherd-bot", AuthorType: "Bot"}}
	if _, err := database.RecoverExpiredClaims(ctx, controller, now.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	uncertain, err := database.Get(ctx, authorization.EffectID())
	if err != nil || uncertain.State != StateUncertain {
		t.Fatalf("expired started effect=%+v err=%v", uncertain, err)
	}
	recoveryAuthorization, err := controller.Reauthorize(ctx, uncertain, now.Add(2*time.Second))
	if err != nil {
		t.Fatalf("expired uncertain effect was not reconciliation-authorized: %v", err)
	}
	dispatcher := &Dispatcher{store: database, executor: newGitHubExecutor(port), fence: allowTestExecution,
		clock: func() time.Time { return now.Add(2 * time.Second) }}
	result, err := dispatcher.Reconcile(ctx, recoveryAuthorization, recoveryAuthorization.Owner(),
		recoveryAuthorization.Epoch(), now.Add(2*time.Second), time.Minute)
	if err != nil || result.ExternalID != 91 || port.createCalls != 0 || port.updateCalls != 0 {
		t.Fatalf("result=%+v creates=%d updates=%d err=%v", result, port.createCalls, port.updateCalls, err)
	}
}

func TestGitHubExecutorBlocksLegacyMarkerWithoutWriting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	_, authorization := testAuthorization(t, 4, 9, now)
	database, err := Open(ctx, filepath.Join(t.TempDir(), "legacy.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, _, err := database.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	port := &fakeCommentPort{comments: []Comment{{ID: 9, Body: "<!-- shepherd-decisions:issue-389 -->\n\nlegacy"}}}
	if _, err := NewGitHubDispatcher(database, port, allowTestExecution).Dispatch(ctx, authorization,
		authorization.Owner(), authorization.Epoch(), now, time.Minute); err == nil {
		t.Fatal("legacy marker was adopted or overwritten")
	}
	record, err := database.Get(ctx, authorization.EffectID())
	if err != nil || record.State != StateBlocked || port.createCalls != 0 || port.updateCalls != 0 {
		t.Fatalf("record=%+v creates=%d updates=%d err=%v", record, port.createCalls, port.updateCalls, err)
	}
}

func TestGitHubExecutorRejectsCopiedMarkerFromUnexpectedActor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	_, authorization := testAuthorization(t, 4, 9, now)
	database, err := Open(ctx, filepath.Join(t.TempDir(), "copied-marker.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, _, err := database.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	port := &fakeCommentPort{actor: "shepherd-bot"}
	body, err := newGitHubExecutor(port).Body(authorization.Record())
	if err != nil {
		t.Fatal(err)
	}
	port.comments = []Comment{{ID: 9, Body: body, Author: "mallory", AuthorType: "User"}}
	if _, err := NewGitHubDispatcher(database, port, allowTestExecution).Dispatch(ctx, authorization,
		authorization.Owner(), authorization.Epoch(), now, time.Minute); err == nil {
		t.Fatal("copied marker from an unexpected actor was reconciled")
	}
	record, err := database.Get(ctx, authorization.EffectID())
	if err != nil || record.State != StateBlocked || port.createCalls != 0 || port.updateCalls != 0 {
		t.Fatalf("record=%+v creates=%d updates=%d err=%v", record, port.createCalls, port.updateCalls, err)
	}
}

func TestGitHubExecutorIgnoresOversizedUnrelatedComment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	_, authorization := testAuthorization(t, 4, 9, now)
	database, err := Open(ctx, filepath.Join(t.TempDir(), "unrelated-large-comment.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, _, err := database.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	port := &fakeCommentPort{comments: []Comment{{ID: 9, Body: strings.Repeat("x", maxPayloadBytes+1),
		Author: "someone", AuthorType: "User"}}}
	if _, err := NewGitHubDispatcher(database, port, allowTestExecution).Dispatch(ctx, authorization,
		authorization.Owner(), authorization.Epoch(), now, time.Minute); err != nil {
		t.Fatal(err)
	}
	if port.createCalls != 1 || port.updateCalls != 0 {
		t.Fatalf("creates=%d updates=%d", port.createCalls, port.updateCalls)
	}
}

func TestGitHubDispatcherFailsClosedWithoutExecutionFence(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	_, authorization := testAuthorization(t, 4, 9, now)
	database, err := Open(ctx, filepath.Join(t.TempDir(), "missing-fence.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, _, err := database.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	port := &fakeCommentPort{}
	if _, err := NewGitHubDispatcher(database, port, nil).Dispatch(ctx, authorization,
		authorization.Owner(), authorization.Epoch(), now, time.Minute); !errors.Is(err, ErrFenced) {
		t.Fatalf("missing execution fence error=%v", err)
	}
	if port.createCalls != 0 || port.updateCalls != 0 {
		t.Fatalf("unfenced writes creates=%d updates=%d", port.createCalls, port.updateCalls)
	}
}

func TestGitHubDispatcherRevalidatesAuthorityAfterPlanningBeforeWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	_, authorization := testAuthorization(t, 4, 9, now)
	database, err := Open(ctx, filepath.Join(t.TempDir(), "fenced-plan.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, _, err := database.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	fenceCalls := 0
	fence := func(context.Context, ClaimedEffect, time.Time) (time.Time, error) {
		fenceCalls++
		if fenceCalls > 1 {
			return time.Time{}, ErrFenced
		}
		return now.Add(time.Minute), nil
	}
	port := &fakeCommentPort{}
	if _, err := NewGitHubDispatcher(database, port, fence).Dispatch(ctx, authorization,
		authorization.Owner(), authorization.Epoch(), now, time.Minute); !errors.Is(err, ErrFenced) {
		t.Fatalf("post-plan fence error=%v", err)
	}
	record, err := database.Get(ctx, authorization.EffectID())
	if err != nil || record.State != StateFailed || port.createCalls != 0 || port.updateCalls != 0 {
		t.Fatalf("record=%+v creates=%d updates=%d err=%v", record, port.createCalls, port.updateCalls, err)
	}
}

func TestGitHubDispatcherPersistsPostSendFenceLossAsUncertain(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	_, authorization := testAuthorization(t, 4, 9, now)
	database, err := Open(ctx, filepath.Join(t.TempDir(), "post-send-fence.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, _, err := database.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	fenceCalls := 0
	fence := func(context.Context, ClaimedEffect, time.Time) (time.Time, error) {
		fenceCalls++
		if fenceCalls > 2 {
			return time.Time{}, ErrFenced
		}
		return now.Add(time.Minute), nil
	}
	port := &fakeCommentPort{}
	if _, err := NewGitHubDispatcher(database, port, fence).Dispatch(ctx, authorization,
		authorization.Owner(), authorization.Epoch(), now, time.Minute); !errors.Is(err, ErrFenced) {
		t.Fatalf("post-send fence error=%v", err)
	}
	record, err := database.Get(ctx, authorization.EffectID())
	if err != nil || record.State != StateUncertain || port.createCalls != 1 {
		t.Fatalf("record=%+v creates=%d err=%v", record, port.createCalls, err)
	}
}

func TestGitHubDispatcherPersistsBoundedPreSendFailureAndRetryBudget(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	_, authorization := testAuthorization(t, 4, 9, now)
	database, err := Open(ctx, filepath.Join(t.TempDir(), "failed.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, _, err := database.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	dispatcher := NewGitHubDispatcher(database, &fakeCommentPort{listErr: errors.New("offline")}, allowTestExecution)
	if _, err := dispatcher.Dispatch(ctx, authorization, authorization.Owner(), authorization.Epoch(), now, time.Minute); err == nil {
		t.Fatal("pre-send listing failure reported success")
	}
	failed, err := database.Get(ctx, authorization.EffectID())
	if err != nil || failed.State != StateFailed || failed.ErrorCode != ErrorPreSend {
		t.Fatalf("failed=%+v err=%v", failed, err)
	}
	if events, eventErr := database.Events(ctx, authorization.EffectID()); eventErr != nil || !hasEvent(events, EventFailed) {
		t.Fatalf("failed telemetry events=%+v err=%v", events, eventErr)
	}
	if err := database.RetryFailed(ctx, authorization, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := dispatcher.Dispatch(ctx, authorization, authorization.Owner(), authorization.Epoch(), now.Add(2*time.Second), time.Minute); err == nil {
		t.Fatal("second pre-send listing failure reported success")
	}
	blocked, err := database.Get(ctx, authorization.EffectID())
	if err != nil || blocked.State != StateBlocked || blocked.ErrorCode != ErrorRetryExhausted {
		t.Fatalf("exhausted effect=%+v err=%v", blocked, err)
	}
	if events, eventErr := database.Events(ctx, authorization.EffectID()); eventErr != nil || !hasEvent(events, EventBlocked) {
		t.Fatalf("retry exhaustion telemetry=%+v err=%v", events, eventErr)
	}
	if err := database.RetryFailed(ctx, authorization, now.Add(3*time.Second)); !errors.Is(err, ErrTerminal) {
		t.Fatalf("terminal exhausted effect retry error=%v", err)
	}
}

func TestGitHubDispatcherBlocksMarkerConflictAndPostSendUncertainty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Now().UTC()
	_, authorization := testAuthorization(t, 4, 9, now)

	conflictDB, err := Open(ctx, filepath.Join(t.TempDir(), "conflict.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conflictDB.Close() })
	if _, _, err := conflictDB.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	seed := &fakeCommentPort{}
	body, err := newGitHubExecutor(seed).Body(authorization.Record())
	if err != nil {
		t.Fatal(err)
	}
	seed.comments = []Comment{{ID: 10, Body: body + "\nconflict"}, {ID: 11, Body: body}}
	_, err = NewGitHubDispatcher(conflictDB, seed, allowTestExecution).Dispatch(ctx, authorization, authorization.Owner(), authorization.Epoch(), now, time.Minute)
	if err == nil {
		t.Fatal("duplicate/conflicting marker ownership was accepted")
	}
	blocked, getErr := conflictDB.Get(ctx, authorization.EffectID())
	if getErr != nil || blocked.State != StateBlocked || seed.createCalls != 0 || seed.updateCalls != 0 {
		t.Fatalf("blocked=%+v creates=%d updates=%d getErr=%v dispatchErr=%v", blocked, seed.createCalls, seed.updateCalls, getErr, err)
	}
	if events, eventErr := conflictDB.Events(ctx, authorization.EffectID()); eventErr != nil || !hasEvent(events, EventBlocked) {
		t.Fatalf("blocked telemetry events=%+v err=%v", events, eventErr)
	}

	uncertainDB, err := Open(ctx, filepath.Join(t.TempDir(), "uncertain.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = uncertainDB.Close() })
	if _, _, err := uncertainDB.Enqueue(ctx, authorization, now); err != nil {
		t.Fatal(err)
	}
	ambiguous := &fakeCommentPort{writeErr: NewAmbiguousWriteError(errors.New("connection lost after send")), writeCreates: true, nextCommentID: 88}
	dispatcher := NewGitHubDispatcher(uncertainDB, ambiguous, allowTestExecution)
	if _, err := dispatcher.Dispatch(ctx, authorization, authorization.Owner(), authorization.Epoch(), now, time.Minute); err == nil {
		t.Fatal("ambiguous post-send failure reported success")
	}
	uncertain, getErr := uncertainDB.Get(ctx, authorization.EffectID())
	if getErr != nil || uncertain.State != StateUncertain || ambiguous.createCalls != 1 {
		t.Fatalf("uncertain=%+v creates=%d getErr=%v", uncertain, ambiguous.createCalls, getErr)
	}
	result, err := dispatcher.Reconcile(ctx, authorization, authorization.Owner(), authorization.Epoch(), now.Add(time.Minute), time.Minute)
	if err != nil || result.Code != ResultReconciled || result.ExternalID != 88 || ambiguous.createCalls != 1 {
		t.Fatalf("uncertain reconciliation result=%+v creates=%d err=%v", result, ambiguous.createCalls, err)
	}
	events, eventErr := uncertainDB.Events(ctx, authorization.EffectID())
	if eventErr != nil || !hasEvent(events, EventUncertain) || !hasEvent(events, EventReconciled) || !hasEvent(events, EventSent) {
		t.Fatalf("uncertainty telemetry events=%+v err=%v", events, eventErr)
	}
}
