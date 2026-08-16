package synctransport

import (
	"context"
	"fmt"
	"time"

	"polymetrics.ai/internal/synccontract"
)

// Orchestrator is the single transport-neutral execution path. It dispatches
// only after Registry.Preflight has resolved exact closed roles; it contains no
// API/database pairing branch.
type Orchestrator struct {
	registry *Registry
}

func NewOrchestrator(registry *Registry) *Orchestrator {
	return &Orchestrator{registry: registry}
}

func (o *Orchestrator) Run(ctx context.Context, request RunRequest) (Result, error) {
	if o == nil || o.registry == nil {
		return Result{}, fmt.Errorf("transport orchestrator registry is required")
	}
	if err := request.validateExecution(); err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	resolved, err := o.registry.Preflight(request.preflightRequest())
	if err != nil {
		return Result{}, err
	}
	if err := request.validateDispatchDependencies(); err != nil {
		return Result{}, err
	}

	plan, err := resolved.Destination.PlanDestination(ctx, cloneDestinationPlanRequest(DestinationPlanRequest{
		Connector:     request.Destination,
		Runtime:       request.DestinationRuntime,
		Source:        request.Source,
		SourceRuntime: request.SourceRuntime,
		Binding:       request.DestinationBinding,
		Stream:        request.Stream,
		Mode:          request.Mode,
		BatchSize:     request.BatchSize,
		ApplyStrategy: resolved.ApplyStrategy,
		Approval:      request.Approval,
	}))
	if err != nil {
		return Result{}, fmt.Errorf("plan destination transport: %w", err)
	}
	if plan.ApplyStrategy != resolved.ApplyStrategy {
		return Result{}, fmt.Errorf("destination transport plan did not preserve the descriptor-resolved apply strategy")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	result := Result{}
	err = resolved.Source.ReadTransport(ctx, cloneSourceRequest(SourceRequest{
		Connector:    request.Source,
		Runtime:      request.SourceRuntime,
		Stream:       request.Stream,
		CursorField:  request.CursorField,
		Mode:         request.Mode,
		BatchSize:    request.BatchSize,
		PrimaryKey:   request.DestinationBinding.PrimaryKey,
		Resume:       request.Resume,
		Checkpoint:   request.Checkpoint,
		UnitDeadline: request.unitDeadline(),
		RecordExtraction: func(elapsed time.Duration) {
			result.ExtractElapsed += elapsed
		},
	}), func(page SourcePage) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if len(page.Records) > request.BatchSize {
			return fmt.Errorf("source transport page has %d records, exceeding requested batch size %d", len(page.Records), request.BatchSize)
		}
		if err := page.CandidateCheckpoint.Validate(); err != nil {
			return fmt.Errorf("source transport candidate checkpoint: %w", err)
		}
		for _, tombstone := range page.Tombstones {
			if err := tombstone.Validate(); err != nil {
				return fmt.Errorf("source transport tombstone: %w", err)
			}
		}
		// Extraction has completed when the source gives us a valid bounded page,
		// independently of whether a later warehouse or destination unit fails.
		result.RecordsRead += len(page.Records)
		if request.Approval.AuthorizeNextUnit != nil {
			if err := request.Approval.AuthorizeNextUnit(ctx); err != nil {
				return fmt.Errorf("authorize transport unit: %w", err)
			}
		}

		// Retain the source-owned candidate for the final commit. The stage and
		// destination receive defensive payload copies and cannot replace it.
		candidate := page.CandidateCheckpoint.Clone()
		stagePage, err := cloneSourcePage(page)
		if err != nil {
			return fmt.Errorf("clone source transport page: %w", err)
		}
		stageStarted := time.Now()
		receipt, err := request.Stage.Stage(ctx, WarehouseStageRequest{
			ConnectionID:    request.ConnectionID,
			Generation:      request.Generation,
			SourceName:      request.Source.Name(),
			DestinationName: request.Destination.Name(),
			Stream:          request.Stream,
			Mode:            request.Mode,
			Page:            stagePage,
		})
		if err != nil {
			result.StageElapsed += time.Since(stageStarted)
			return fmt.Errorf("stage transport page: %w", err)
		}
		if err := receipt.Validate(); err != nil {
			return err
		}
		staged, err := request.Stage.Reopen(ctx, receipt)
		if err != nil {
			result.StageElapsed += time.Since(stageStarted)
			return fmt.Errorf("reopen staged transport receipt: %w", err)
		}
		result.StageElapsed += time.Since(stageStarted)
		if staged.ID == "" {
			return fmt.Errorf("warehouse stage reopened an empty workset ID")
		}
		if staged.ID != receipt.ID {
			return fmt.Errorf("warehouse stage reopened workset %q for receipt %q", staged.ID, receipt.ID)
		}
		// The re-opened workset is the independent durable Parquet boundary.
		result.RecordsStaged += len(staged.Records)
		if err := ctx.Err(); err != nil {
			return err
		}
		destinationWorkset, err := cloneWarehouseWorkset(staged)
		if err != nil {
			return fmt.Errorf("clone warehouse transport workset: %w", err)
		}
		destinationApplyRequest, err := cloneDestinationApplyRequest(DestinationApplyRequest{
			ConnectionID:  request.ConnectionID,
			Plan:          plan,
			Receipt:       receipt,
			Workset:       destinationWorkset,
			Runtime:       request.DestinationRuntime,
			Source:        request.Source,
			SourceRuntime: request.SourceRuntime,
			Binding:       request.DestinationBinding,
			Stream:        request.Stream,
			Mode:          request.Mode,
			BatchSize:     request.BatchSize,
			Approval:      request.Approval,
		})
		if err != nil {
			return fmt.Errorf("clone destination apply request: %w", err)
		}
		acknowledgement, err := func() (synccontract.DownstreamAcknowledgement, error) {
			applyCtx, cancelApply := context.WithTimeout(ctx, request.unitDeadline())
			defer cancelApply()

			applyStarted := time.Now()
			acknowledgement, err := resolved.Destination.ApplyDestination(applyCtx, destinationApplyRequest)
			result.ApplyElapsed += time.Since(applyStarted)
			if err != nil {
				return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("apply destination transport: %w", err)
			}
			// Apply acknowledged the bounded workset. Retain this count even if the
			// subsequent independent read-back refuses to advance its checkpoint.
			result.RecordsApplied += len(staged.Records)
			readBackRequest, err := cloneDestinationReadBackRequest(DestinationReadBackRequest{
				Plan:            plan,
				Workset:         destinationWorkset,
				Acknowledgement: acknowledgement,
				Runtime:         request.DestinationRuntime,
				Source:          request.Source,
				SourceRuntime:   request.SourceRuntime,
				Binding:         request.DestinationBinding,
				Stream:          request.Stream,
				Mode:            request.Mode,
			})
			if err != nil {
				return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("clone destination read-back request: %w", err)
			}
			readBackStarted := time.Now()
			if err := resolved.Destination.ReadBackDestination(applyCtx, readBackRequest); err != nil {
				result.ApplyElapsed += time.Since(readBackStarted)
				return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("read back destination transport receipt: %w", err)
			}
			result.ApplyElapsed += time.Since(readBackStarted)
			return acknowledgement, nil
		}()
		if err != nil {
			return err
		}
		if !page.DeferCheckpoint {
			if err := synccontract.CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(checkpoint synccontract.CheckpointEnvelope) error {
				if acknowledgement.Sink != request.Destination.Name() {
					return fmt.Errorf("durable downstream acknowledgement sink %q does not match destination %q", acknowledgement.Sink, request.Destination.Name())
				}
				return request.Commit(checkpoint)
			}); err != nil {
				return err
			}
		}

		committed := candidate.Clone()
		committedAt := acknowledgement.AcknowledgedAt
		committed.CommittedAt = &committedAt
		result.Pages++
		if !page.DeferCheckpoint {
			result.CommittedCheckpoint = &committed
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	if result.CommittedCheckpoint == nil {
		if result.Pages == 0 {
			if _, allowed := resolved.Source.(EmptyResultSource); allowed {
				return result, nil
			}
		}
		return result, fmt.Errorf("source transport completed without a checkpoint candidate")
	}
	return result, nil
}
