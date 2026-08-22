package synctransport

import (
	"context"
	"fmt"
	"time"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

// runArrowFullOverwritePipelined keeps the full-overwrite publish barrier
// outside a bounded producer/consumer pair. The source owns an Arrow record
// until its emit callback returns; the queue receives one retained reference,
// and the single consumer releases it after transform, segment storage and
// ordered COPY complete. Consequently the next source range can overlap one
// COPY without allowing unordered apply or an early checkpoint.
func (o *Orchestrator) runArrowFullOverwritePipelined(ctx context.Context, request RunRequest, resolved ResolvedTransport, plan DestinationPlan, source ArrowRangeExtractor, destination ArrowBulkDestination) (result Result, err error) {
	started := time.Now()
	defer func() { result.WallElapsed = time.Since(started) }()
	if request.MaxInFlightBatches < 2 || request.Mode != synccontract.ModeFullOverwrite || source == nil || destination == nil || isNilInterface(request.FastSegments) || request.TransformPlanHash == "" || request.TransformPlanJSON == "" || plan.TransformPlanHash != request.TransformPlanHash {
		return result, ErrArrowFastPathInvalid
	}
	transform, parseErr := database.ParseTransformPlanV1([]byte(request.TransformPlanJSON))
	if parseErr != nil || transform.Hash() != request.TransformPlanHash {
		return result, ErrArrowFastPathInvalid
	}
	transformer, transformEngineErr := database.NewArrowTransformer(ctx, transform)
	if transformEngineErr != nil {
		return result, fmt.Errorf("open Arrow transform engine: %w", transformEngineErr)
	}
	defer func() {
		if closeErr := transformer.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close Arrow transform engine: %w", closeErr)
		}
	}()
	capacity := request.ByteCreditCapacity
	if capacity == 0 {
		capacity = DefaultByteCreditCapacity
	}
	credits, creditErr := NewByteCreditController(capacity)
	if creditErr != nil {
		return result, ErrArrowFastPathInvalid
	}
	defer func() {
		snapshot := credits.Snapshot()
		result.PeakCreditBytes = snapshot.Peak
		result.CreditWaitElapsed = time.Duration(snapshot.WaitNanos)
	}()

	beginCtx, cancelBegin := transportUnitContext(ctx, request.unitDeadline())
	if err := authorizeDestinationEffect(beginCtx, request.Approval, "Arrow full-overwrite begin"); err != nil {
		cancelBegin()
		return result, err
	}
	session, err := destination.BeginArrowFullOverwrite(beginCtx, cloneArrowFullOverwriteRunRequest(ArrowFullOverwriteRunRequest{
		ConnectionID: request.ConnectionID, Generation: request.Generation, Plan: plan,
		Runtime: request.DestinationRuntime, Source: request.Source,
		SourceRuntime: request.SourceRuntime, Binding: request.DestinationBinding,
		Stream: request.Stream, BatchSize: request.BatchSize, TransformPlanJSON: request.TransformPlanJSON,
		TransformPlanHash: request.TransformPlanHash, Approval: request.Approval,
	}))
	cancelBegin()
	if err != nil || isNilInterface(session) {
		if err == nil {
			err = ErrArrowFastPathInvalid
		}
		return result, fmt.Errorf("begin Arrow full-overwrite run: %w", tagTransportExecutionError(TransportExecutionOriginDestination, err))
	}
	published := false
	defer func() {
		if published {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), request.unitDeadline())
		defer cancel()
		if abortErr := session.AbortArrowFullOverwrite(cleanupCtx); abortErr != nil && err == nil {
			err = fmt.Errorf("abort Arrow full-overwrite run: %w", tagTransportExecutionError(TransportExecutionOriginDestination, abortErr))
		}
	}()

	pipelineCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	// The consumer's active unit and a source record blocked in its synchronous
	// emit callback both count toward the bound. Reserving those two positions
	// leaves only depth-2 buffered slots: at depth two, page N+1 can overlap
	// COPY of N, but its callback cannot return to trigger page N+2 extraction
	// until COPY advances. This prevents a full queue plus an extra retained
	// callback record from exceeding the requested in-flight count.
	work := make(chan arrowPipelineBatch, request.MaxInFlightBatches-2)
	producerDone := make(chan arrowPipelineProducerResult, 1)
	go func() {
		defer close(work)
		producerDone <- produceArrowPipeline(pipelineCtx, request, source, credits, work)
	}()

	segments := make([]FastSegmentReceipt, 0)
	var lastCandidate *synccontract.CheckpointEnvelope
	consumerErr := consumeArrowPipeline(pipelineCtx, request, plan, session, transformer, work, &result, &segments, &lastCandidate)
	if consumerErr != nil {
		cancel()
	}
	producer := <-producerDone
	// Extraction is producer-owned so the concurrent producer never mutates the
	// consumer's Result. Preserve every completed source measurement even when
	// a later consumer operation fails; this is the same partial-result
	// accounting the serial controller exposes.
	result.ExtractElapsed += producer.extractElapsed
	result.RecordsRead += producer.recordsRead
	result.SourceLogicalBytes += producer.sourceLogicalBytes
	result.Pages += producer.pages
	if consumerErr != nil {
		drainArrowPipeline(work)
		return result, consumerErr
	}
	if producer.err != nil {
		drainArrowPipeline(work)
		return result, producer.err
	}
	if lastCandidate == nil {
		if _, allowed := resolved.Source.(EmptyResultSource); !allowed {
			return result, fmt.Errorf("Arrow source transport completed without a checkpoint candidate")
		}
	}
	if err := authorizeDestinationEffect(ctx, request.Approval, "Arrow full-overwrite publication"); err != nil {
		return result, err
	}
	publishCtx, cancelPublish := transportUnitContext(ctx, request.unitDeadline())
	publishStarted := time.Now()
	acknowledgement, publishErr := session.PublishArrowFullOverwrite(publishCtx, ArrowFullOverwritePublicationRequest{
		LastCheckpoint: lastCandidate, Segments: append([]FastSegmentReceipt(nil), segments...),
		SourceLogicalBytes: result.SourceLogicalBytes, SourceRows: int64(result.RecordsRead),
		TransformedRows: int64(result.RecordsApplied), TransformedBytes: result.TransformedBytes,
	})
	collectDestinationResult(&result, acknowledgement, publishErr)
	result.PublishElapsed += time.Since(publishStarted)
	cancelPublish()
	if publishErr == nil {
		published = true
		if lastCandidate == nil {
			result.WallElapsed = time.Since(started)
			if err := handoffEmptyPublicationReadBackPending(ctx, request, &result, acknowledgement, request.Destination.Name()); err != nil {
				return result, err
			}
		}
		readBackCtx, cancelReadBack := transportUnitContext(ctx, request.unitDeadline())
		readBackStarted := time.Now()
		if request.ReadBackAdmission != nil {
			publishErr = request.ReadBackAdmission(readBackCtx)
		}
		if publishErr == nil {
			publishErr = session.ReadBackArrowFullOverwrite(readBackCtx, acknowledgement)
		}
		result.ReadBackElapsed += time.Since(readBackStarted)
		cancelReadBack()
	}
	if publishErr != nil {
		return result, fmt.Errorf("publish Arrow full-overwrite run: %w", tagTransportExecutionError(TransportExecutionOriginDestination, publishErr))
	}
	if reporter, ok := session.(ArrowBulkPhaseReporter); ok {
		result.IndexConstraintElapsed = reporter.ArrowBulkPhaseMeasurement().IndexConstraintBuildElapsed
	}
	if lastCandidate == nil {
		result.WallElapsed = time.Since(started)
		if err := handoffEmptyPublication(ctx, request, &result, acknowledgement, request.Destination.Name()); err != nil {
			return result, err
		}
		return result, nil
	}
	checkpointStarted := time.Now()
	if err := synccontract.CommitAfterDownstreamAcknowledgement(*lastCandidate, acknowledgement, func(checkpoint synccontract.CheckpointEnvelope) error {
		if acknowledgement.Sink != request.Destination.Name() {
			return fmt.Errorf("durable downstream acknowledgement sink %q does not match destination %q", acknowledgement.Sink, request.Destination.Name())
		}
		return request.Commit(checkpoint)
	}); err != nil {
		result.CheckpointElapsed += time.Since(checkpointStarted)
		return result, err
	}
	result.CheckpointElapsed += time.Since(checkpointStarted)
	committed := lastCandidate.Clone()
	committedAt := acknowledgement.AcknowledgedAt
	committed.CommittedAt = &committedAt
	result.CommittedCheckpoint = &committed
	return result, nil
}

type arrowPipelineBatch struct {
	batch ArrowSourceBatch
	lease *ByteCreditLease
}

func (b arrowPipelineBatch) release() {
	if b.batch.Record != nil {
		b.batch.Record.Release()
	}
	if b.lease != nil {
		b.lease.Release()
	}
}

type arrowPipelineProducerResult struct {
	extractElapsed     time.Duration
	recordsRead        int
	sourceLogicalBytes int64
	pages              int
	err                error
}

func produceArrowPipeline(ctx context.Context, request RunRequest, source ArrowRangeExtractor, credits *ByteCreditController, work chan<- arrowPipelineBatch) (result arrowPipelineProducerResult) {
	if err := authorizeTransportSource(ctx, request); err != nil {
		result.err = err
		return result
	}
	result.err = source.ExtractArrowRanges(ctx, cloneArrowExtractRequest(ArrowExtractRequest{
		Connector: request.Source, Runtime: request.SourceRuntime, Stream: request.Stream, CursorField: request.CursorField,
		PrimaryKey: request.DestinationBinding.PrimaryKey, Resume: request.Resume, Checkpoint: sourceCheckpointForMode(request.Mode, request.Checkpoint, request.RateLimitResumeCheckpoint),
		BatchSize: request.BatchSize, UnitDeadline: request.unitDeadline(), TransformPlanJSON: request.TransformPlanJSON, TransformHash: request.TransformPlanHash,
	}), func(batch ArrowSourceBatch) (callbackErr error) {
		defer func() {
			callbackErr = tagTransportExecutionError(TransportExecutionOriginInternal, callbackErr)
		}()
		if err := validateArrowSourceBatch(batch, request.BatchSize); err != nil {
			return err
		}
		if request.Approval.AuthorizeNextUnit != nil {
			if err := request.Approval.AuthorizeNextUnit(ctx); err != nil {
				return fmt.Errorf("authorize Arrow transport unit: %w", err)
			}
		}
		creditBytes := batch.SourceLogicalBytes
		if creditBytes == 0 {
			creditBytes = arrowRecordBytes(batch.Record)
		}
		var lease *ByteCreditLease
		if creditBytes > 0 {
			var err error
			lease, err = credits.Acquire(ctx, creditBytes)
			if err != nil {
				return fmt.Errorf("admit Arrow source range: %w", err)
			}
		}
		batch.Record.Retain()
		admitted := arrowPipelineBatch{batch: batch, lease: lease}
		select {
		case work <- admitted:
			result.extractElapsed += batch.ExtractElapsed
			result.recordsRead += int(batch.SourceRows)
			result.sourceLogicalBytes += batch.SourceLogicalBytes
			result.pages++
			return nil
		case <-ctx.Done():
			admitted.release()
			return ctx.Err()
		}
	})
	if result.err != nil {
		result.err = tagTransportExecutionError(TransportExecutionOriginSource, result.err)
	}
	if result.err != nil && ctx.Err() != nil {
		return result
	}
	return result
}

func consumeArrowPipeline(ctx context.Context, request RunRequest, plan DestinationPlan, session ArrowFullOverwriteRun, transformer *database.ArrowTransformer, work <-chan arrowPipelineBatch, result *Result, segments *[]FastSegmentReceipt, lastCandidate **synccontract.CheckpointEnvelope) (err error) {
	defer func() {
		err = tagTransportExecutionError(TransportExecutionOriginInternal, err)
	}()
	for admitted := range work {
		err := func() error {
			defer admitted.release()
			if err := ctx.Err(); err != nil {
				return err
			}
			transformStarted := time.Now()
			output, err := transformer.Transform(ctx, admitted.batch.Record)
			result.TransformElapsed += time.Since(transformStarted)
			if err != nil {
				return fmt.Errorf("transform Arrow source range: %w", err)
			}
			defer output.Release()
			for output.Next() {
				record := output.Record()
				segmentID, err := newArrowSegmentID()
				if err != nil {
					return err
				}
				parquetStarted := time.Now()
				receipt, err := request.FastSegments.StoreArrowSegment(ctx, FastSegmentWriteRequest{
					ConnectionID: request.ConnectionID, Generation: request.Generation, Stream: request.Stream, SegmentID: segmentID,
					TransformPlanHash: request.TransformPlanHash, SourceLogicalBytes: admitted.batch.SourceLogicalBytes, SourceRows: admitted.batch.SourceRows, Record: record,
				})
				result.ParquetElapsed += time.Since(parquetStarted)
				if err != nil || validateFastSegmentReceipt(receipt, request.TransformPlanHash) != nil {
					if err == nil {
						err = ErrArrowFastPathInvalid
					}
					return fmt.Errorf("close Arrow Parquet segment: %w", err)
				}
				applyCtx, cancelApply := context.WithTimeout(ctx, request.unitDeadline())
				if err := authorizeDestinationEffect(applyCtx, request.Approval, "Arrow segment apply"); err != nil {
					cancelApply()
					return err
				}
				applyStarted := time.Now()
				applyErr := session.ApplyArrowSegment(applyCtx, cloneArrowBulkApplyRequest(ArrowBulkApplyRequest{
					ConnectionID: request.ConnectionID, Plan: plan, Segment: receipt, Record: record,
					Runtime: request.DestinationRuntime, Source: request.Source, SourceRuntime: request.SourceRuntime,
					Binding: request.DestinationBinding, Stream: request.Stream, BatchSize: request.BatchSize, Approval: request.Approval,
				}))
				result.ApplyElapsed += time.Since(applyStarted)
				cancelApply()
				if applyErr != nil {
					return fmt.Errorf("bulk apply Arrow segment: %w", tagTransportExecutionError(TransportExecutionOriginDestination, applyErr))
				}
				if reporter, ok := session.(ArrowBulkPhaseReporter); ok {
					result.IndexConstraintElapsed = reporter.ArrowBulkPhaseMeasurement().IndexConstraintBuildElapsed
				}
				*segments = append(*segments, receipt)
				result.RecordsStaged += int(receipt.TransformedRows)
				result.RecordsApplied += int(receipt.TransformedRows)
				result.TransformedBytes += receipt.TransformedBytes
				result.ParquetBytes += receipt.ParquetBytes
			}
			if err := output.Err(); err != nil {
				return fmt.Errorf("iterate transformed Arrow range: %w", err)
			}
			candidate := admitted.batch.CandidateCheckpoint.Clone()
			*lastCandidate = &candidate
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func drainArrowPipeline(work <-chan arrowPipelineBatch) {
	for admitted := range work {
		admitted.release()
	}
}
