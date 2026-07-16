package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	decisionlog "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/decision"
	shepherdgit "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/git"
	shepherdgithub "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/github"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/outbox"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/workspace"
)

const effectClaimTTL = 2 * time.Minute

var externalEffectStoreOpen = outbox.Open

var externalEffectDispatcherFactory = func(database *outbox.Store, fence outbox.ExecutionFence) *outbox.Dispatcher {
	return shepherdgithub.NewCLIOutboxDispatcher(database, fence)
}

type externalEffectRequester interface {
	RequestSummary(context.Context, decisionlog.Snapshot) (outbox.Result, error)
	RequestQuestion(context.Context, store.DecisionRequest) (outbox.Result, error)
}

type externalEffectController struct {
	policy           *outbox.Controller
	facts            outbox.ControllerFacts
	authority        *store.Store
	lease            store.Lease
	store            *outbox.Store
	dispatcher       *outbox.Dispatcher
	workspaceManager *workspace.Manager
	now              func() time.Time
}

func openExternalEffectController(ctx context.Context, authorityStore *store.Store, lease store.Lease,
	config fileConfig, deliveryID string, issue int, generation int64, headSHA string) (*externalEffectController, error) {
	if authorityStore == nil || lease.RunID != deliveryID {
		return nil, errors.New("authority store and matching delivery lease are required")
	}
	if err := authorityStore.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
		return nil, err
	}
	delivery, err := authorityStore.GetDelivery(ctx, deliveryID)
	if err != nil {
		return nil, err
	}
	resolvedDeliveryRoot, deliveryRootErr := filepath.EvalSymlinks(delivery.WorkDir)
	resolvedConfigRoot, configRootErr := filepath.EvalSymlinks(config.WorkDir)
	if deliveryRootErr != nil || configRootErr != nil || resolvedDeliveryRoot != resolvedConfigRoot {
		return nil, errors.New("external effect work_dir no longer resolves to the immutable delivery root")
	}
	if delivery.Issue != issue || delivery.Repository != config.Repository ||
		delivery.PullRequest != config.PullRequest || delivery.WorkDir != config.WorkDir {
		return nil, errors.New("external effect config does not match the immutable delivery target")
	}
	facts := outbox.ControllerFacts{
		DeliveryID: deliveryID, Repository: delivery.Repository, Issue: delivery.Issue,
		PullRequest: delivery.PullRequest, Generation: generation, HeadSHA: headSHA,
		Owner: lease.Owner, Epoch: lease.Epoch,
	}
	policy, err := outbox.NewController(ctx, authorityStore, lease, facts, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("external effect policy: %w", err)
	}
	database, err := externalEffectStoreOpen(ctx, filepath.Join(config.StateDir, "outbox.db"))
	if err != nil {
		return nil, err
	}
	executionFence := func(fenceCtx context.Context, claim outbox.ClaimedEffect, now time.Time) (time.Time, error) {
		if claim.ControllerOwner != lease.Owner || claim.ControllerEpoch != lease.Epoch {
			return time.Time{}, outbox.ErrFenced
		}
		if err := policy.Revalidate(fenceCtx, now); err != nil {
			return time.Time{}, err
		}
		canonical, err := shepherdgit.Inspect(fenceCtx, delivery.WorkDir)
		if err != nil || canonical.Branch != delivery.Branch || canonical.HeadSHA != facts.HeadSHA {
			return time.Time{}, outbox.ErrFenced
		}
		return lease.ExpiresAt, nil
	}
	return &externalEffectController{
		policy: policy, facts: facts, authority: authorityStore, lease: lease, store: database,
		dispatcher: externalEffectDispatcherFactory(database, executionFence),
		now:        func() time.Time { return time.Now().UTC() },
	}, nil
}

func (c *externalEffectController) Close() error {
	if c == nil || c.store == nil {
		return nil
	}
	return c.store.Close()
}

func withStandaloneExternalEffectController(ctx context.Context, config fileConfig, issue int,
	operation func(*externalEffectController) error) (returnErr error) {
	if operation == nil {
		return errors.New("external effect operation is required")
	}
	deliveryID := deliveryID(issue)
	authorityStore, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, authorityStore.Close()) }()
	delivery, err := authorityStore.GetDelivery(ctx, deliveryID)
	if err != nil {
		return err
	}
	if delivery.Issue != issue || delivery.Branch == "" || delivery.Repository != config.Repository ||
		delivery.PullRequest != config.PullRequest || delivery.WorkDir != config.WorkDir {
		return errors.New("external effect config does not match the durable delivery target/worktree")
	}
	manager, repositoryLock, err := lockDeliveryWorkspace(config)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, repositoryLock.Close()) }()
	run, err := authorityStore.GetDeliveryRun(ctx, deliveryID)
	if err != nil {
		return err
	}
	snapshot, err := shepherdgit.Inspect(ctx, config.WorkDir)
	if err != nil {
		return err
	}
	if snapshot.Branch != delivery.Branch || snapshot.HeadSHA == "" {
		return errors.New("external effect requires the authorized canonical branch and exact head")
	}
	owner := fmt.Sprintf("effect-%d", time.Now().UTC().UnixNano())
	lease, err := authorityStore.AcquireLease(ctx, deliveryID, owner, time.Now().UTC(), effectClaimTTL)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, authorityStore.ReleaseLease(context.Background(), lease)) }()
	fencedRun, err := authorityStore.GetDeliveryRun(ctx, deliveryID)
	if err != nil || fencedRun.Generation != run.Generation {
		return errors.Join(errors.New("delivery generation changed during external effect admission"), err)
	}
	if err := repositoryLock.Check(); err != nil {
		return err
	}
	controller, err := openExternalEffectController(ctx, authorityStore, lease, config, deliveryID, issue,
		fencedRun.Generation, snapshot.HeadSHA)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, controller.Close()) }()
	controller.workspaceManager = manager
	return operation(controller)
}

func (c *externalEffectController) RequestSummary(ctx context.Context, snapshot decisionlog.Snapshot) (outbox.Result, error) {
	if c == nil || c.policy == nil {
		return outbox.Result{}, errors.New("external effect controller is required")
	}
	intent, err := outbox.NewSummaryIntent(outbox.Target{
		Repository: c.policyFactsRepository(), Issue: c.policyFactsIssue(), PullRequest: c.policyFactsPullRequest(),
	}, snapshot.DeliveryID, c.policyFactsGeneration(), c.policyFactsHead(), snapshot.Revision,
		snapshot.LedgerHash, snapshot.Summary)
	if err != nil {
		return outbox.Result{}, err
	}
	now := c.now()
	if err := c.policy.Revalidate(ctx, now); err != nil {
		return outbox.Result{}, err
	}
	projected, found, err := c.store.FindSentSummary(ctx, snapshot.DeliveryID, intent.Target(),
		snapshot.Revision, snapshot.LedgerHash)
	if err != nil {
		return outbox.Result{}, err
	}
	if found {
		return projected.Result, nil
	}
	return c.request(ctx, intent)
}

func (c *externalEffectController) RequestQuestion(ctx context.Context, request store.DecisionRequest) (outbox.Result, error) {
	if c == nil || c.policy == nil {
		return outbox.Result{}, errors.New("external effect controller is required")
	}
	intent, err := decisionQuestionIntent(request)
	if err != nil {
		return outbox.Result{}, err
	}
	return c.request(ctx, intent)
}

func (c *externalEffectController) VerifyPublishedQuestion(ctx context.Context, request store.DecisionRequest) error {
	intent, err := decisionQuestionIntent(request)
	if err != nil {
		return err
	}
	authorization, err := c.policy.Authorize(ctx, intent, c.now())
	if err != nil {
		return err
	}
	record, err := c.store.Get(ctx, authorization.EffectID())
	if err != nil {
		return fmt.Errorf("published decision question has no durable outbox effect: %w", err)
	}
	if record.State != outbox.StateSent || record.Result.ExternalID != request.GitHubCommentID ||
		record.Result.ExternalActor == "" {
		return errors.New("published decision question is not bound to a sent outbox result")
	}
	return nil
}

func decisionQuestionIntent(request store.DecisionRequest) (outbox.Intent, error) {
	return outbox.NewQuestionIntent(outbox.Target{
		Repository: request.Repository, Issue: request.Issue, PullRequest: request.PullRequest,
	}, outbox.Question{
		RequestID: request.RequestID, DeliveryID: request.DeliveryID, UnitID: request.UnitID,
		Generation: request.Generation, HeadSHA: request.HeadSHA, Evidence: request.Evidence,
		Options: request.Options, RecommendedOption: request.RecommendedOption, SafeDefault: request.SafeDefault,
		ExpiresAt: request.ExpiresAt, Mention: "karthik-sivadas",
	})
}

func (c *externalEffectController) request(ctx context.Context, intent outbox.Intent) (outbox.Result, error) {
	if c == nil || c.policy == nil || c.authority == nil || c.store == nil || c.dispatcher == nil {
		return outbox.Result{}, errors.New("external effect controller is required")
	}
	now := c.now()
	if err := c.authority.CheckLease(ctx, c.lease, now); err != nil {
		return outbox.Result{}, fmt.Errorf("external effect lease: %w", err)
	}
	authorization, err := c.policy.Authorize(ctx, intent, now)
	if err != nil {
		return outbox.Result{}, err
	}
	record, _, err := c.store.Enqueue(ctx, authorization, now)
	if err != nil {
		return outbox.Result{}, fmt.Errorf("enqueue authorized external effect: %w", err)
	}
	integrationEffectEnqueuedBoundary()
	return c.dispatchRecord(ctx, authorization, record, now)
}

func (c *externalEffectController) dispatchRecord(ctx context.Context, authorization outbox.Authorization,
	record outbox.EffectRecord, now time.Time) (outbox.Result, error) {
	if err := c.authority.CheckLease(ctx, c.lease, now); err != nil {
		return outbox.Result{}, fmt.Errorf("external effect execution lease: %w", err)
	}
	switch record.State {
	case outbox.StateSent:
		return record.Result, nil
	case outbox.StatePending:
		return c.dispatcher.Dispatch(ctx, authorization, c.lease.Owner, c.lease.Epoch, now, integrationEffectTTL())
	case outbox.StateFailed:
		if err := c.store.RetryFailed(ctx, authorization, now); err != nil {
			return outbox.Result{}, err
		}
		return c.dispatcher.Dispatch(ctx, authorization, c.lease.Owner, c.lease.Epoch, now, integrationEffectTTL())
	case outbox.StateUncertain:
		return c.dispatcher.Reconcile(ctx, authorization, c.lease.Owner, c.lease.Epoch, now, integrationEffectTTL())
	case outbox.StateClaimed:
		return outbox.Result{}, errors.New("external effect is already claimed and requires fenced recovery")
	case outbox.StateBlocked, outbox.StateCancelled:
		return outbox.Result{}, fmt.Errorf("external effect is terminal in state %q", record.State)
	default:
		return outbox.Result{}, fmt.Errorf("external effect has unknown state %q", record.State)
	}
}

func reconcileExternalEffects(ctx context.Context, authorityStore *store.Store,
	controller *externalEffectController, stateDir, deliveryID string) error {
	if authorityStore == nil || controller == nil {
		return errors.New("authority and external effect controllers are required")
	}
	legacyEffects, err := authorityStore.HasLegacyExternalEffects(ctx)
	if err != nil {
		return err
	}
	if legacyEffects {
		return store.ErrLegacyExternalEffects
	}
	preReconcileRequests, err := authorityStore.ListOpenDecisionRequests(ctx, deliveryID)
	if err != nil {
		return err
	}
	for _, request := range preReconcileRequests {
		if request.Status != store.DecisionRequestOpen || request.HeadSHA == controller.facts.HeadSHA {
			continue
		}
		recoveredPromotion, err := authorityStore.IsRecoveredPromotionDecision(ctx, request)
		if err != nil || !recoveredPromotion {
			if err != nil {
				return err
			}
			continue
		}
		intent, err := decisionQuestionIntent(request)
		if err != nil {
			return err
		}
		record, err := controller.store.Get(ctx, intent.EffectID())
		if err == nil && record.State == outbox.StateSent {
			if err := authorityStore.MarkDecisionRequestPublishedFenced(ctx, controller.lease, request,
				record.Result.ExternalID, controller.now()); err != nil {
				return err
			}
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	if err := authorityStore.CancelStaleDecisionRequests(ctx, controller.lease, deliveryID,
		controller.facts.Repository, controller.facts.Issue, controller.facts.PullRequest,
		controller.facts.Generation, controller.facts.HeadSHA, controller.now()); err != nil {
		return fmt.Errorf("cancel stale decision requests: %w", err)
	}
	requests, err := authorityStore.ListOpenDecisionRequests(ctx, deliveryID)
	if err != nil {
		return err
	}
	for _, request := range requests {
		if !request.ExpiresAt.After(controller.now()) {
			intent, intentErr := decisionQuestionIntent(request)
			var effectErr error
			projectedSent := false
			if intentErr == nil {
				if record, getErr := controller.store.Get(ctx, intent.EffectID()); getErr == nil {
					if record.State == outbox.StateSent {
						effectErr = authorityStore.MarkDecisionRequestPublishedFenced(ctx, controller.lease,
							request, record.Result.ExternalID, controller.now())
						projectedSent = effectErr == nil
					} else {
						effectErr = controller.store.CancelExpiredQuestion(ctx, controller.policy,
							intent.EffectID(), controller.now())
						if errors.Is(effectErr, outbox.ErrStaleUncertain) {
							uncertain, readErr := controller.store.Get(ctx, intent.EffectID())
							if readErr == nil && uncertain.State == outbox.StateUncertain {
								result, reconcileErr := controller.reconcileUncertainRecord(ctx, uncertain,
									controller.now())
								if reconcileErr == nil {
									effectErr = authorityStore.MarkDecisionRequestPublishedFenced(ctx,
										controller.lease, request, result.ExternalID, controller.now())
									projectedSent = effectErr == nil
								} else {
									effectErr = errors.Join(effectErr, reconcileErr)
								}
							} else if readErr != nil {
								effectErr = errors.Join(effectErr, readErr)
							}
						}
					}
				} else if !errors.Is(getErr, sql.ErrNoRows) {
					effectErr = getErr
				}
			} else {
				effectErr = intentErr
			}
			if projectedSent {
				continue
			}
			if errors.Is(effectErr, outbox.ErrStaleUncertain) {
				return fmt.Errorf("reconcile expired uncertain decision-question effect %s before safe-stop: %w",
					request.RequestID, effectErr)
			}
			if err := authorityStore.ExpireDecisionRequestAndBlock(ctx, controller.lease, request.RequestID,
				controller.now()); err != nil {
				return fmt.Errorf("expire decision request %s: %w", request.RequestID, err)
			}
			if effectErr != nil {
				return fmt.Errorf("settle expired decision-question effect %s: %w", request.RequestID, effectErr)
			}
		}
	}
	if err := controller.RecoverClaims(ctx); err != nil {
		return err
	}
	requests, err = authorityStore.ListOpenDecisionRequests(ctx, deliveryID)
	if err != nil {
		return err
	}
	for _, request := range requests {
		if !request.ExpiresAt.After(controller.now()) {
			continue
		}
		recoveredPromotion, err := authorityStore.IsRecoveredPromotionDecision(ctx, request)
		if err != nil {
			return err
		}
		if recoveredPromotion && request.Status == store.DecisionRequestOpen {
			intent, err := decisionQuestionIntent(request)
			if err != nil {
				return err
			}
			record, getErr := controller.store.Get(ctx, intent.EffectID())
			safeToCancel := errors.Is(getErr, sql.ErrNoRows)
			if getErr == nil {
				switch record.State {
				case outbox.StatePending, outbox.StateFailed:
					authorization, err := controller.policy.Authorize(ctx, intent, controller.now())
					if err != nil {
						return err
					}
					if _, _, err := controller.store.Enqueue(ctx, authorization, controller.now()); err != nil {
						return err
					}
					if err := controller.store.Cancel(ctx, authorization, controller.now()); err != nil {
						return err
					}
					safeToCancel = true
				case outbox.StateCancelled, outbox.StateBlocked:
					safeToCancel = true
				}
			} else if !errors.Is(getErr, sql.ErrNoRows) {
				return getErr
			}
			if safeToCancel {
				if err := authorityStore.CancelUnpublishedRecoveredPromotionDecision(ctx, controller.lease,
					request, controller.now()); err != nil {
					return err
				}
				continue
			}
		}
		if recoveredPromotion && request.Status == store.DecisionRequestPublished && request.GitHubCommentID > 0 {
			continue
		}
		if request.Status == store.DecisionRequestPublished && request.GitHubCommentID > 0 {
			if err := controller.VerifyPublishedQuestion(ctx, request); err != nil {
				return fmt.Errorf("verify published decision question %s: %w", request.RequestID, err)
			}
			continue
		}
		result, err := controller.RequestQuestion(ctx, request)
		if err != nil {
			return fmt.Errorf("reconcile decision-question effect %s: %w", request.RequestID, err)
		}
		if err := authorityStore.MarkDecisionRequestPublishedFenced(ctx, controller.lease, request,
			result.ExternalID, controller.now()); err != nil {
			return fmt.Errorf("project decision-question result %s: %w", request.RequestID, err)
		}
	}
	if err := controller.RecoverUncertain(ctx); err != nil {
		return err
	}
	decisionStore, err := decisionlog.Open(filepath.Join(stateDir, "decisions"))
	if err != nil {
		return err
	}
	snapshot, snapshotErr := decisionStore.Snapshot(deliveryID)
	closeErr := decisionStore.Close()
	if snapshotErr != nil || closeErr != nil {
		return errors.Join(snapshotErr, closeErr)
	}
	if snapshot.Revision > 0 {
		if _, err := controller.RequestSummary(ctx, snapshot); err != nil {
			return fmt.Errorf("reconcile decision-summary effect: %w", err)
		}
	}
	if err := controller.Recover(ctx); err != nil {
		return err
	}
	return nil
}

func (c *externalEffectController) RecoverClaims(ctx context.Context) error {
	now := c.now()
	if err := c.policy.Revalidate(ctx, now); err != nil {
		return fmt.Errorf("external effect recovery lease: %w", err)
	}
	if err := c.store.ReconcileStale(ctx, c.policy, now); err != nil {
		return fmt.Errorf("reconcile stale external effects: %w", err)
	}
	if _, err := c.store.RecoverExpiredClaims(ctx, c.policy, now); err != nil {
		return fmt.Errorf("recover expired effect claims: %w", err)
	}
	return nil
}

func (c *externalEffectController) reconcileUncertainRecord(ctx context.Context, record outbox.EffectRecord,
	now time.Time) (outbox.Result, error) {
	authorization, err := c.policy.Reauthorize(ctx, record, now)
	if err != nil {
		return outbox.Result{}, err
	}
	if _, _, err := c.store.Enqueue(ctx, authorization, now); err != nil {
		return outbox.Result{}, err
	}
	return c.dispatchRecord(ctx, authorization, record, now)
}

func (c *externalEffectController) RecoverUncertain(ctx context.Context) error {
	now := c.now()
	if err := c.policy.Revalidate(ctx, now); err != nil {
		return err
	}
	records, err := c.store.ListDelivery(ctx, c.policyFactsDelivery(), outbox.StateUncertain)
	if err != nil {
		return err
	}
	for _, record := range records {
		if _, err := c.reconcileUncertainRecord(ctx, record, now); err != nil {
			return fmt.Errorf("reconcile uncertain durable effect %s: %w", record.EffectID, err)
		}
	}
	return nil
}

func (c *externalEffectController) Recover(ctx context.Context) error {
	if err := c.RecoverClaims(ctx); err != nil {
		return err
	}
	now := c.now()
	records, err := c.store.ListDelivery(ctx, c.policyFactsDelivery(), outbox.StatePending,
		outbox.StateFailed, outbox.StateUncertain, outbox.StateClaimed, outbox.StateBlocked)
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.State == outbox.StateBlocked {
			if record.Kind == outbox.KindDecisionQuestion {
				request, err := c.authority.GetDecisionRequest(ctx, record.SourceID)
				if err == nil && (request.Status == store.DecisionRequestCancelled ||
					request.Status == store.DecisionRequestExpired) {
					continue
				}
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return err
				}
			}
			if record.Generation != c.facts.Generation || record.HeadSHA != c.facts.HeadSHA ||
				record.Target.Repository != c.facts.Repository || record.Target.Issue != c.facts.Issue ||
				record.Target.PullRequest != c.facts.PullRequest {
				continue
			}
			if record.Kind == outbox.KindDecisionSummary &&
				(record.ErrorCode == outbox.ErrorStaleRevision || record.ErrorCode == outbox.ErrorRetryExhausted) {
				higher, err := c.store.HasHigherSummaryRevision(ctx, record)
				if err != nil {
					return err
				}
				if higher {
					continue
				}
			}
			return fmt.Errorf("durable external effect %s is blocked: %s", record.EffectID, record.ErrorCode)
		}
		authorization, err := c.policy.Reauthorize(ctx, record, now)
		if err != nil {
			return fmt.Errorf("reauthorize durable effect %s: %w", record.EffectID, err)
		}
		if _, _, err := c.store.Enqueue(ctx, authorization, now); err != nil {
			return fmt.Errorf("persist recovered effect authorization %s: %w", record.EffectID, err)
		}
		if _, err := c.dispatchRecord(ctx, authorization, record, now); err != nil {
			if errors.Is(err, outbox.ErrStaleRevision) {
				continue
			}
			return fmt.Errorf("recover durable effect %s: %w", record.EffectID, err)
		}
	}
	return nil
}

// Facts are intentionally exposed only through these controller-local helpers;
// enqueue and execution code cannot create or widen grants.
func (c *externalEffectController) policyFactsDelivery() string   { return c.facts.DeliveryID }
func (c *externalEffectController) policyFactsRepository() string { return c.facts.Repository }
func (c *externalEffectController) policyFactsIssue() int         { return c.facts.Issue }
func (c *externalEffectController) policyFactsPullRequest() int   { return c.facts.PullRequest }
func (c *externalEffectController) policyFactsGeneration() int64  { return c.facts.Generation }
func (c *externalEffectController) policyFactsHead() string       { return c.facts.HeadSHA }
