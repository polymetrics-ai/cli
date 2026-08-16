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

// OrderedPipelineUnsupportedError is a typed pre-I/O refusal when a caller
// asks for more than one in-flight batch but the selected fast-path endpoints
// have not both declared the ordered pipeline contract.
type OrderedPipelineUnsupportedError struct {
	Source      string
	Destination string
}

func (e *OrderedPipelineUnsupportedError) Error() string {
	if e == nil {
		return "ordered pipeline is not supported by the selected transport endpoints"
	}
	return fmt.Sprintf("ordered pipeline is not supported by source %q and destination %q", e.Source, e.Destination)
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
	if request.MaxInFlightBatches > 1 && (request.Mode != synccontract.ModeFullOverwrite || request.TransformPlanHash == "" || request.TransformPlanJSON == "" || !resolved.SourceDescriptor.OrderedPipeline || !resolved.DestinationDescriptor.OrderedPipeline) {
		return Result{}, &OrderedPipelineUnsupportedError{Source: request.Source.Name(), Destination: request.Destination.Name()}
	}

	plan, err := resolved.Destination.PlanDestination(ctx, cloneDestinationPlanRequest(DestinationPlanRequest{
		Connector:         request.Destination,
		Runtime:           request.DestinationRuntime,
		Source:            request.Source,
		SourceRuntime:     request.SourceRuntime,
		Binding:           request.DestinationBinding,
		Stream:            request.Stream,
		Mode:              request.Mode,
		BatchSize:         request.BatchSize,
		TransformPlanJSON: request.TransformPlanJSON,
		TransformPlanHash: request.TransformPlanHash,
		ApplyStrategy:     resolved.ApplyStrategy,
		Approval:          request.Approval,
	}))
	if err != nil {
		return Result{}, fmt.Errorf("plan destination transport: %w", err)
	}
	if plan.ApplyStrategy != resolved.ApplyStrategy {
		return Result{}, fmt.Errorf("destination transport plan did not preserve the descriptor-resolved apply strategy")
	}
	if plan.TransformPlanHash != request.TransformPlanHash {
		return Result{}, fmt.Errorf("destination transport plan did not preserve the transform plan hash")
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if request.Mode == synccontract.ModeFullOverwrite && request.TransformPlanHash != "" {
		if request.TransformPlanJSON == "" {
			return Result{}, ErrArrowFastPathInvalid
		}
		arrowSource, sourceOK := resolved.Source.(ArrowRangeExtractor)
		arrowDestination, destinationOK := resolved.Destination.(ArrowBulkDestination)
		if !sourceOK || !destinationOK {
			return Result{}, ErrArrowFastPathInvalid
		}
		return o.runArrowFullOverwrite(ctx, request, resolved, plan, arrowSource, arrowDestination)
	}
	if request.TransformPlanJSON != "" {
		return Result{}, ErrArrowFastPathInvalid
	}
	if err := request.validateLegacyDispatchDependencies(); err != nil {
		return Result{}, err
	}
	if request.Mode == synccontract.ModeFullOverwrite {
		fullOverwriteDestination, ok := resolved.Destination.(FullOverwriteDestination)
		if !ok {
			return Result{}, fmt.Errorf("destination transport does not implement run-scoped full_overwrite")
		}
		return o.runFullOverwrite(ctx, request, resolved, plan, fullOverwriteDestination)
	}

	result := Result{}
	pendingReceipts := make([]WarehouseReceipt, 0)
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
		pendingReceipts = append(pendingReceipts, receipt)
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
			if err := retireCommittedReceipts(ctx, request, pendingReceipts); err != nil {
				return err
			}
			pendingReceipts = pendingReceipts[:0]
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

// runFullOverwrite preserves the source → durable-stage → destination order
// while making publication a single run-scoped action. Normal transport modes
// deliberately retain their page acknowledgement behavior below; only replace
// requires one final receipt and one final checkpoint.
func (o *Orchestrator) runFullOverwrite(ctx context.Context, request RunRequest, resolved ResolvedTransport, plan DestinationPlan, destination FullOverwriteDestination) (result Result, err error) {
	fullRequest := FullOverwriteRunRequest{
		ConnectionID:  request.ConnectionID,
		Generation:    request.Generation,
		Plan:          plan,
		Runtime:       cloneRuntimeConfig(request.DestinationRuntime),
		Source:        request.Source,
		SourceRuntime: cloneRuntimeConfig(request.SourceRuntime),
		Binding: DestinationBinding{
			WorkspaceID:       request.DestinationBinding.WorkspaceID,
			SourceConnectorID: request.DestinationBinding.SourceConnectorID,
			ConnectionID:      request.DestinationBinding.ConnectionID,
			StreamID:          request.DestinationBinding.StreamID,
			PrimaryKey:        append([]string(nil), request.DestinationBinding.PrimaryKey...),
		},
		Stream:            request.Stream,
		Mode:              request.Mode,
		BatchSize:         request.BatchSize,
		TransformPlanJSON: request.TransformPlanJSON,
		TransformPlanHash: request.TransformPlanHash,
		Approval:          request.Approval,
	}
	session, err := destination.BeginFullOverwrite(ctx, fullRequest)
	if err != nil {
		return Result{}, fmt.Errorf("begin destination full-overwrite run: %w", err)
	}
	if isNilInterface(session) {
		return Result{}, fmt.Errorf("destination transport returned an empty full-overwrite run")
	}
	published := false
	defer func() {
		if published {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), request.unitDeadline())
		defer cancel()
		if abortErr := session.AbortFullOverwrite(cleanupCtx); abortErr != nil && err == nil {
			err = fmt.Errorf("abort destination full-overwrite run: %w", abortErr)
		}
	}()

	var lastCandidate *synccontract.CheckpointEnvelope
	receipts := make([]WarehouseReceipt, 0)
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
		if len(page.Tombstones) != 0 {
			return fmt.Errorf("full_overwrite source transport page has tombstones")
		}
		result.RecordsRead += len(page.Records)
		if request.Approval.AuthorizeNextUnit != nil {
			if err := request.Approval.AuthorizeNextUnit(ctx); err != nil {
				return fmt.Errorf("authorize transport unit: %w", err)
			}
		}

		candidate := page.CandidateCheckpoint.Clone()
		stagePage, err := cloneSourcePage(page)
		if err != nil {
			return fmt.Errorf("clone source transport page: %w", err)
		}
		stageStarted := time.Now()
		receipt, err := request.Stage.Stage(ctx, WarehouseStageRequest{
			ConnectionID: request.ConnectionID, Generation: request.Generation,
			SourceName: request.Source.Name(), DestinationName: request.Destination.Name(),
			Stream: request.Stream, Mode: request.Mode, Page: stagePage,
		})
		if err != nil {
			result.StageElapsed += time.Since(stageStarted)
			return fmt.Errorf("stage transport page: %w", err)
		}
		if err := receipt.Validate(); err != nil {
			return err
		}
		receipts = append(receipts, receipt)
		staged, err := request.Stage.Reopen(ctx, receipt)
		if err != nil {
			result.StageElapsed += time.Since(stageStarted)
			return fmt.Errorf("reopen staged transport receipt: %w", err)
		}
		result.StageElapsed += time.Since(stageStarted)
		if staged.ID == "" || staged.ID != receipt.ID {
			return fmt.Errorf("warehouse stage reopened workset %q for receipt %q", staged.ID, receipt.ID)
		}
		result.RecordsStaged += len(staged.Records)
		destinationWorkset, err := cloneWarehouseWorkset(staged)
		if err != nil {
			return fmt.Errorf("clone warehouse transport workset: %w", err)
		}
		applyRequest, err := cloneDestinationApplyRequest(DestinationApplyRequest{
			ConnectionID: request.ConnectionID, Plan: plan, Receipt: receipt, Workset: destinationWorkset,
			Runtime: request.DestinationRuntime, Source: request.Source, SourceRuntime: request.SourceRuntime,
			Binding: request.DestinationBinding, Stream: request.Stream, Mode: request.Mode,
			BatchSize: request.BatchSize, Approval: request.Approval,
		})
		if err != nil {
			return fmt.Errorf("clone destination apply request: %w", err)
		}
		applyCtx, cancelApply := context.WithTimeout(ctx, request.unitDeadline())
		applyStarted := time.Now()
		applyErr := session.ApplyFullOverwrite(applyCtx, applyRequest)
		result.ApplyElapsed += time.Since(applyStarted)
		cancelApply()
		if applyErr != nil {
			return fmt.Errorf("apply destination full-overwrite workset: %w", applyErr)
		}
		result.RecordsApplied += len(staged.Records)
		result.Pages++
		copy := candidate.Clone()
		lastCandidate = &copy
		return nil
	})
	if err != nil {
		return result, err
	}
	if result.Pages == 0 {
		if _, allowed := resolved.Source.(EmptyResultSource); !allowed {
			return result, fmt.Errorf("source transport completed without a full-overwrite checkpoint candidate")
		}
	}
	if request.Approval.AuthorizeNextUnit != nil {
		if err := request.Approval.AuthorizeNextUnit(ctx); err != nil {
			return result, fmt.Errorf("authorize full-overwrite publication: %w", err)
		}
	}
	publishCtx, cancelPublish := context.WithTimeout(ctx, request.unitDeadline())
	publishStarted := time.Now()
	acknowledgement, publishErr := session.PublishFullOverwrite(publishCtx, FullOverwritePublicationRequest{
		LastCheckpoint: lastCandidate, Pages: result.Pages, Records: result.RecordsApplied,
	})
	result.ApplyElapsed += time.Since(publishStarted)
	if publishErr == nil {
		publishErr = session.ReadBackFullOverwrite(publishCtx, acknowledgement)
	}
	cancelPublish()
	if publishErr != nil {
		return result, fmt.Errorf("publish destination full-overwrite run: %w", publishErr)
	}
	published = true
	if lastCandidate == nil {
		return result, nil
	}
	if err := synccontract.CommitAfterDownstreamAcknowledgement(*lastCandidate, acknowledgement, func(checkpoint synccontract.CheckpointEnvelope) error {
		if acknowledgement.Sink != request.Destination.Name() {
			return fmt.Errorf("durable downstream acknowledgement sink %q does not match destination %q", acknowledgement.Sink, request.Destination.Name())
		}
		return request.Commit(checkpoint)
	}); err != nil {
		return result, err
	}
	if err := retireCommittedReceipts(ctx, request, receipts); err != nil {
		return result, err
	}
	committed := lastCandidate.Clone()
	committedAt := acknowledgement.AcknowledgedAt
	committed.CommittedAt = &committedAt
	result.CommittedCheckpoint = &committed
	return result, nil
}

func retireCommittedReceipts(ctx context.Context, request RunRequest, receipts []WarehouseReceipt) error {
	stage, ok := request.Stage.(RetirableWarehouseStage)
	if !ok || len(receipts) == 0 {
		return nil
	}
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), request.unitDeadline())
	defer cancel()
	for _, receipt := range receipts {
		if err := stage.Retire(cleanupCtx, receipt); err != nil {
			return fmt.Errorf("retire committed warehouse stage receipt %q: %w", receipt.ID, err)
		}
	}
	return nil
}
