package outbox

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Comment struct {
	ID         int64
	Body       string
	Author     string
	AuthorType string
}

type CommentPort interface {
	Actor(context.Context) (string, error)
	ListComments(context.Context, Target) ([]Comment, error)
	CreateComment(context.Context, Target, string) (Comment, error)
	UpdateComment(context.Context, Target, int64, string) (Comment, error)
}

type Action string

const (
	actionCreate Action = "create"
	actionUpdate Action = "update"
	actionExact  Action = "exact"
	actionBlock  Action = "block"
)

type Plan struct {
	Action        Action
	ExternalID    int64
	ExternalActor string
	ErrorCode     ErrorCode
}

type executor interface {
	Plan(context.Context, EffectRecord, []EffectRecord) (Plan, error)
	Send(context.Context, EffectRecord, Plan) (Result, error)
}

type githubExecutor struct {
	port CommentPort
}

func newGitHubExecutor(port CommentPort) *githubExecutor { return &githubExecutor{port: port} }

func (e *githubExecutor) Body(record EffectRecord) (string, error) {
	if e == nil || e.port == nil {
		return "", errors.New("GitHub comment port is required")
	}
	intent, err := DecodeIntent(record.Kind, record.Payload)
	if err != nil || matchIntent(record, intent) != nil {
		return "", errors.New("effect record is not bound to a valid immutable payload")
	}
	var body string
	switch record.Kind {
	case KindDecisionSummary:
		var payload SummaryPayload
		if err := decodeStrict(record.Payload, &payload); err != nil {
			return "", err
		}
		body = summaryMarker(record) + "\n\n" + strings.TrimSpace(payload.Summary) + "\n"
	case KindDecisionQuestion:
		var payload QuestionPayload
		if err := decodeStrict(record.Payload, &payload); err != nil {
			return "", err
		}
		body = questionMarker(record) + "\n\n" + renderQuestion(payload) + "\n"
	default:
		return "", fmt.Errorf("effect kind %q has no GitHub executor", record.Kind)
	}
	if len(body) > maxPayloadBytes {
		return "", errors.New("rendered GitHub comment exceeds the bounded size")
	}
	return body, nil
}

func (e *githubExecutor) Plan(ctx context.Context, record EffectRecord, history []EffectRecord) (Plan, error) {
	body, err := e.Body(record)
	if err != nil {
		return Plan{}, err
	}
	actor, err := e.port.Actor(ctx)
	if err != nil || !safeID(actor) {
		return Plan{}, errors.New("bounded GitHub writer identity is required")
	}
	comments, err := e.port.ListComments(ctx, record.Target)
	if err != nil {
		return Plan{}, fmt.Errorf("list bounded GitHub comments: %w", err)
	}
	if len(comments) > 10_000 {
		return Plan{}, errors.New("GitHub comment listing exceeds the bounded count")
	}
	slot := markerSlot(record)
	var owned []Comment
	for _, comment := range comments {
		claimsLegacy := strings.HasPrefix(comment.Body, legacyMarker(record))
		claimsCurrent := strings.HasPrefix(comment.Body, slot)
		if !claimsLegacy && !claimsCurrent {
			continue
		}
		if comment.ID <= 0 || len(comment.Body) > maxPayloadBytes {
			return Plan{Action: actionBlock, ErrorCode: ErrorMarkerConflict}, nil
		}
		if (claimsLegacy || claimsCurrent) && (comment.Author != actor ||
			(comment.AuthorType != "User" && comment.AuthorType != "Bot")) {
			return Plan{Action: actionBlock, ExternalID: comment.ID, ErrorCode: ErrorMarkerConflict}, nil
		}
		if claimsLegacy {
			return Plan{Action: actionBlock, ExternalID: comment.ID, ExternalActor: actor, ErrorCode: ErrorMarkerConflict}, nil
		}
		if claimsCurrent {
			owned = append(owned, comment)
		}
	}
	if len(owned) > 1 {
		return Plan{Action: actionBlock, ExternalActor: actor, ErrorCode: ErrorMarkerConflict}, nil
	}
	if len(owned) == 0 {
		return Plan{Action: actionCreate, ExternalActor: actor}, nil
	}
	comment := owned[0]
	if comment.Body == body {
		return Plan{Action: actionExact, ExternalID: comment.ID, ExternalActor: actor}, nil
	}
	if record.Kind == KindDecisionQuestion {
		return Plan{Action: actionBlock, ExternalID: comment.ID, ExternalActor: actor, ErrorCode: ErrorMarkerConflict}, nil
	}
	revision, ok := parseSummaryRevision(comment.Body, record.DeliveryID)
	if !ok || revision >= record.Revision {
		code := ErrorMarkerConflict
		if ok && revision > record.Revision {
			code = ErrorStaleRevision
		}
		return Plan{Action: actionBlock, ExternalID: comment.ID, ExternalActor: actor, ErrorCode: code}, nil
	}
	for _, previous := range history {
		if previous.Kind != KindDecisionSummary || previous.DeliveryID != record.DeliveryID ||
			previous.Target != record.Target || previous.Revision != revision || previous.State != StateSent ||
			previous.Result.ExternalID != comment.ID {
			continue
		}
		previousBody, err := e.Body(previous)
		if err == nil && previousBody == comment.Body && previous.Result.ExternalActor == actor {
			return Plan{Action: actionUpdate, ExternalID: comment.ID, ExternalActor: actor}, nil
		}
	}
	return Plan{Action: actionBlock, ExternalID: comment.ID, ExternalActor: actor, ErrorCode: ErrorMarkerConflict}, nil
}

func (e *githubExecutor) Send(ctx context.Context, record EffectRecord, plan Plan) (Result, error) {
	body, err := e.Body(record)
	if err != nil {
		return Result{}, err
	}
	switch plan.Action {
	case actionCreate:
		written, err := e.port.CreateComment(ctx, record.Target, body)
		if err != nil {
			return Result{}, err
		}
		if !matchesWrittenComment(written, 0, plan.ExternalActor, body) {
			return Result{}, NewAmbiguousWriteError(errors.New("GitHub create did not return the bound comment identity"))
		}
		return Result{Code: ResultSent, ExternalID: written.ID, ExternalActor: written.Author}, nil
	case actionUpdate:
		if plan.ExternalID <= 0 {
			return Result{}, errors.New("GitHub update requires an existing comment identity")
		}
		written, err := e.port.UpdateComment(ctx, record.Target, plan.ExternalID, body)
		if err != nil {
			return Result{}, err
		}
		if !matchesWrittenComment(written, plan.ExternalID, plan.ExternalActor, body) {
			return Result{}, NewAmbiguousWriteError(errors.New("GitHub update did not return the bound comment identity"))
		}
		return Result{Code: ResultSent, ExternalID: written.ID, ExternalActor: written.Author}, nil
	default:
		return Result{}, errors.New("GitHub send plan is not writable")
	}
}

func matchesWrittenComment(comment Comment, expectedID int64, actor, body string) bool {
	return comment.ID > 0 && (expectedID == 0 || comment.ID == expectedID) && comment.Author == actor &&
		(comment.AuthorType == "User" || comment.AuthorType == "Bot") && comment.Body == body
}

type ambiguousWriteError struct{ cause error }

func (e ambiguousWriteError) Error() string   { return "external write completion is ambiguous" }
func (e ambiguousWriteError) Unwrap() error   { return e.cause }
func (e ambiguousWriteError) Ambiguous() bool { return true }

func NewAmbiguousWriteError(cause error) error {
	if cause == nil {
		cause = errors.New("external write completion is unknown")
	}
	return ambiguousWriteError{cause: cause}
}

func isAmbiguousWrite(err error) bool {
	var ambiguous interface{ Ambiguous() bool }
	return errors.As(err, &ambiguous) && ambiguous.Ambiguous()
}

type ExecutionFence func(context.Context, ClaimedEffect, time.Time) (time.Time, error)

type Dispatcher struct {
	store    *Store
	executor executor
	clock    func() time.Time
	fence    ExecutionFence
}

func NewGitHubDispatcher(store *Store, port CommentPort, fence ExecutionFence) *Dispatcher {
	return &Dispatcher{
		store: store, executor: newGitHubExecutor(port),
		clock: func() time.Time { return time.Now().UTC() }, fence: fence,
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, authorization Authorization, owner string, epoch int64, now time.Time, ttl time.Duration) (Result, error) {
	if d == nil || d.store == nil || d.executor == nil || d.clock == nil {
		return Result{}, errors.New("outbox store, executor, and clock are required")
	}
	claim, err := d.store.Claim(ctx, authorization, owner, epoch, now, ttl)
	if err != nil {
		return Result{}, err
	}
	planCtx, cancelPlan, _, err := d.fencedContext(ctx, claim)
	if err != nil {
		markErr := d.store.MarkFailed(ctx, claim, ErrorPreSend, d.clock())
		return Result{}, errors.Join(err, markErr)
	}
	history, historyErr := d.store.ListDelivery(planCtx, claim.DeliveryID, StateSent)
	if historyErr != nil {
		cancelPlan()
		markErr := d.store.MarkFailed(ctx, claim, ErrorPreSend, d.clock())
		return Result{}, errors.Join(historyErr, markErr)
	}
	plan, err := d.executor.Plan(planCtx, claim.EffectRecord, history)
	cancelPlan()
	if err != nil {
		markErr := d.store.MarkFailed(ctx, claim, ErrorPreSend, d.clock())
		return Result{}, errors.Join(err, markErr)
	}
	writeCtx, cancelWrite, executionNow, err := d.fencedContext(ctx, claim)
	if err != nil {
		markErr := d.store.MarkFailed(ctx, claim, ErrorPreSend, executionNow)
		return Result{}, errors.Join(err, markErr)
	}
	defer cancelWrite()
	switch plan.Action {
	case actionExact:
		result := Result{Code: ResultReconciled, ExternalID: plan.ExternalID, ExternalActor: plan.ExternalActor}
		return result, d.store.MarkSent(ctx, claim, result, executionNow)
	case actionBlock:
		markErr := d.store.MarkBlocked(ctx, claim, plan.ErrorCode, executionNow)
		return Result{}, errors.Join(errors.New("GitHub comment reconciliation blocked the effect"), markErr)
	case actionCreate, actionUpdate:
	default:
		markErr := d.store.MarkBlocked(ctx, claim, ErrorUnsupported, executionNow)
		return Result{}, errors.Join(errors.New("unsupported executor plan"), markErr)
	}
	if err := d.store.StartExecution(ctx, claim, executionNow); err != nil {
		code := ErrorPreSend
		if errors.Is(err, ErrStaleRevision) {
			code = ErrorStaleRevision
			return Result{}, errors.Join(err, d.store.MarkBlocked(ctx, claim, code, executionNow))
		}
		return Result{}, errors.Join(err, d.store.MarkFailed(ctx, claim, code, executionNow))
	}
	result, err := d.executor.Send(writeCtx, claim.EffectRecord, plan)
	settledAt := d.clock()
	if err != nil {
		// Every error returned after invoking the write port is ambiguous. The
		// marker reconciliation path, not blind retry, is the only recovery.
		if !isAmbiguousWrite(err) {
			err = NewAmbiguousWriteError(err)
		}
		markErr := d.store.MarkUncertain(ctx, claim, ErrorPostSendAmbiguous, settledAt)
		return Result{}, errors.Join(err, markErr)
	}
	_, cancelFinal, finalAt, err := d.fencedContext(ctx, claim)
	cancelFinal()
	if err != nil {
		markErr := d.store.MarkUncertain(ctx, claim, ErrorPostSendAmbiguous, settledAt)
		return Result{}, errors.Join(err, markErr)
	}
	if err := d.store.MarkSent(ctx, claim, result, finalAt); err != nil {
		return Result{}, err
	}
	return result, nil
}

func (d *Dispatcher) Reconcile(ctx context.Context, authorization Authorization, owner string, epoch int64, now time.Time, ttl time.Duration) (Result, error) {
	if d == nil || d.store == nil || d.executor == nil || d.clock == nil || owner != authorization.Owner() || epoch != authorization.Epoch() || ttl <= 0 {
		return Result{}, ErrFenced
	}
	record, err := d.store.Get(ctx, authorization.EffectID())
	if err != nil {
		return Result{}, err
	}
	if record.State != StateUncertain {
		return Result{}, errors.New("only an uncertain effect may use reconciliation-only dispatch")
	}
	claim := ClaimedEffect{EffectRecord: record, GrantID: authorization.GrantID(), ControllerOwner: owner,
		ControllerEpoch: epoch, ClaimedAt: now, ExpiresAt: now.Add(ttl)}
	planCtx, cancel, _, err := d.reconciliationContext(ctx, claim)
	if err != nil {
		return Result{}, err
	}
	defer cancel()
	history, err := d.store.ListDelivery(planCtx, record.DeliveryID, StateSent)
	if err != nil {
		return Result{}, err
	}
	plan, err := d.executor.Plan(planCtx, record, history)
	if err != nil {
		return Result{}, err
	}
	_, cancelFinal, finalAt, err := d.reconciliationContext(ctx, claim)
	if err != nil {
		return Result{}, err
	}
	defer cancelFinal()
	if plan.Action != actionExact || plan.ExternalID <= 0 {
		return Result{}, errors.New("uncertain effect has no exact remote identity; write remains blocked")
	}
	result := Result{Code: ResultReconciled, ExternalID: plan.ExternalID, ExternalActor: plan.ExternalActor}
	if err := d.store.MarkReconciled(ctx, authorization, result, finalAt); err != nil {
		return Result{}, err
	}
	return result, nil
}

func (d *Dispatcher) fencedContext(ctx context.Context, claim ClaimedEffect) (context.Context, context.CancelFunc, time.Time, error) {
	return d.boundedContext(ctx, claim, true)
}

func (d *Dispatcher) reconciliationContext(ctx context.Context, claim ClaimedEffect) (context.Context, context.CancelFunc, time.Time, error) {
	return d.boundedContext(ctx, claim, false)
}

func (d *Dispatcher) boundedContext(ctx context.Context, claim ClaimedEffect,
	enforceQuestionDeadline bool) (context.Context, context.CancelFunc, time.Time, error) {
	now := d.clock().UTC()
	if d.fence == nil {
		return nil, func() {}, now, ErrFenced
	}
	deadline := claim.ExpiresAt
	if questionDeadline, ok := effectExpiry(claim.EffectRecord); enforceQuestionDeadline && ok && questionDeadline.Before(deadline) {
		deadline = questionDeadline
	}
	authorityDeadline, err := d.fence(ctx, claim, now)
	if err != nil {
		return nil, func() {}, now, err
	}
	if authorityDeadline.IsZero() {
		return nil, func() {}, now, errors.New("execution fence returned no authority deadline")
	}
	if authorityDeadline.Before(deadline) {
		deadline = authorityDeadline
	}
	if !now.Before(deadline) {
		return nil, func() {}, now, ErrFenced
	}
	bounded, cancel := context.WithDeadline(ctx, deadline)
	return bounded, cancel, now, nil
}

func effectExpiry(record EffectRecord) (time.Time, bool) {
	if record.Kind != KindDecisionQuestion {
		return time.Time{}, false
	}
	var payload QuestionPayload
	if err := decodeStrict(record.Payload, &payload); err != nil {
		return time.Time{}, false
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, payload.ExpiresAt)
	return expiresAt, err == nil
}

func summaryMarker(record EffectRecord) string {
	return markerSlot(record) + strconv.FormatInt(record.Revision, 10) + ":" +
		strings.TrimPrefix(record.PayloadHash, "sha256:") + ":" +
		strings.TrimPrefix(record.IdempotencyKey, "sha256:") + " -->"
}

func questionMarker(record EffectRecord) string {
	return markerSlot(record) + strings.TrimPrefix(record.PayloadHash, "sha256:") + ":" +
		strings.TrimPrefix(record.IdempotencyKey, "sha256:") + " -->"
}

func legacyMarker(record EffectRecord) string {
	switch record.Kind {
	case KindDecisionSummary:
		return "<!-- shepherd-decisions:" + record.DeliveryID + " -->"
	case KindDecisionQuestion:
		return "<!-- shepherd-question:" + record.DeliveryID + ":" + record.SourceID + " -->"
	default:
		return "<!-- shepherd-unsupported -->"
	}
}

func markerSlot(record EffectRecord) string {
	switch record.Kind {
	case KindDecisionSummary:
		return "<!-- shepherd-effect:decision-summary:" + record.DeliveryID + ":"
	case KindDecisionQuestion:
		return "<!-- shepherd-effect:decision-question:" + record.DeliveryID + ":" + record.SourceID + ":"
	default:
		return "<!-- shepherd-effect:unsupported:"
	}
}

func parseSummaryRevision(body, deliveryID string) (int64, bool) {
	line, _, _ := strings.Cut(body, "\n")
	prefix := "<!-- shepherd-effect:decision-summary:" + deliveryID + ":"
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, " -->") {
		return 0, false
	}
	fields := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, prefix), " -->"), ":")
	if len(fields) != 3 || len(fields[1]) != 64 || len(fields[2]) != 64 ||
		!validLowerHex(fields[1]) || !validLowerHex(fields[2]) {
		return 0, false
	}
	revision, err := strconv.ParseInt(fields[0], 10, 64)
	return revision, err == nil && revision > 0
}

func renderQuestion(payload QuestionPayload) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "@%s Shepherd needs a decision.\n\n", payload.Mention)
	fmt.Fprintf(&builder, "- Request: `%s`\n", payload.RequestID)
	fmt.Fprintf(&builder, "- Issue: #%d\n", payload.Issue)
	fmt.Fprintf(&builder, "- PR: #%d\n", payload.PullRequest)
	fmt.Fprintf(&builder, "- Unit: `%s`\n", payload.UnitID)
	fmt.Fprintf(&builder, "- Generation: `%d`\n", payload.Generation)
	fmt.Fprintf(&builder, "- Head: `%s`\n", payload.HeadSHA)
	fmt.Fprintf(&builder, "- Evidence: %s\n", payload.Evidence)
	if payload.Recommended != "" {
		fmt.Fprintf(&builder, "- Recommended option: `%s`\n", payload.Recommended)
	}
	if payload.SafeDefault != "" {
		fmt.Fprintf(&builder, "- Safe default at expiry: `%s`\n", payload.SafeDefault)
	}
	fmt.Fprintf(&builder, "- Expires: `%s`\n\n", payload.ExpiresAt)
	builder.WriteString("Options:\n")
	for _, option := range payload.Options {
		fmt.Fprintf(&builder, "- `%s`\n", option)
	}
	fmt.Fprintf(&builder, "\nReply exactly with: `/shepherd decide %s <option>`", payload.RequestID)
	return builder.String()
}
