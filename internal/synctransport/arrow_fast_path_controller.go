package synctransport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/apache/arrow-go/v18/arrow"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

// runArrowFullOverwrite is the source/destination-neutral high-throughput
// controller. Connector adapters own only range extraction and native bulk
// apply. The controller owns Arrow/DuckDB transformation, immutable Parquet
// segment admission, byte credits, one run receipt and checkpoint-after-
// reconciliation sequencing.
func (o *Orchestrator) runArrowFullOverwrite(ctx context.Context, request RunRequest, resolved ResolvedTransport, plan DestinationPlan, source ArrowRangeExtractor, destination ArrowBulkDestination) (result Result, err error) {
	if request.MaxInFlightBatches > 1 {
		return o.runArrowFullOverwritePipelined(ctx, request, resolved, plan, source, destination)
	}
	started := time.Now()
	defer func() { result.WallElapsed = time.Since(started) }()
	if request.Mode != synccontract.ModeFullOverwrite || source == nil || destination == nil || isNilInterface(request.FastSegments) || request.TransformPlanHash == "" || request.TransformPlanJSON == "" || plan.TransformPlanHash != request.TransformPlanHash {
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

	session, err := destination.BeginArrowFullOverwrite(ctx, ArrowFullOverwriteRunRequest{
		ConnectionID: request.ConnectionID, Generation: request.Generation, Plan: plan,
		Runtime: cloneRuntimeConfig(request.DestinationRuntime), Source: request.Source,
		SourceRuntime: cloneRuntimeConfig(request.SourceRuntime),
		Binding:       DestinationBinding{WorkspaceID: request.DestinationBinding.WorkspaceID, SourceConnectorID: request.DestinationBinding.SourceConnectorID, ConnectionID: request.DestinationBinding.ConnectionID, StreamID: request.DestinationBinding.StreamID, PrimaryKey: append([]string(nil), request.DestinationBinding.PrimaryKey...)},
		Stream:        request.Stream, BatchSize: request.BatchSize, TransformPlanJSON: request.TransformPlanJSON,
		TransformPlanHash: request.TransformPlanHash, Approval: request.Approval,
	})
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

	var lastCandidate *synccontract.CheckpointEnvelope
	segments := make([]FastSegmentReceipt, 0)
	err = source.ExtractArrowRanges(ctx, cloneArrowExtractRequest(ArrowExtractRequest{
		Connector: request.Source, Runtime: request.SourceRuntime, Stream: request.Stream, CursorField: request.CursorField,
		PrimaryKey: request.DestinationBinding.PrimaryKey, Resume: request.Resume, Checkpoint: sourceCheckpointForMode(request.Mode, request.Checkpoint),
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
			defer lease.Release()
		}
		transformStarted := time.Now()
		output, err := transformer.Transform(ctx, batch.Record)
		result.TransformElapsed += time.Since(transformStarted)
		if err != nil {
			return fmt.Errorf("transform Arrow source range: %w", err)
		}
		defer output.Release()
		candidate := batch.CandidateCheckpoint.Clone()
		lastCandidate = &candidate
		result.ExtractElapsed += batch.ExtractElapsed
		result.RecordsRead += int(batch.SourceRows)
		result.SourceLogicalBytes += batch.SourceLogicalBytes
		result.Pages++
		for output.Next() {
			record := output.Record()
			segmentID, err := newArrowSegmentID()
			if err != nil {
				return err
			}
			parquetStarted := time.Now()
			receipt, err := request.FastSegments.StoreArrowSegment(ctx, FastSegmentWriteRequest{
				ConnectionID: request.ConnectionID, Generation: request.Generation, Stream: request.Stream, SegmentID: segmentID,
				TransformPlanHash: request.TransformPlanHash, SourceLogicalBytes: batch.SourceLogicalBytes, SourceRows: batch.SourceRows, Record: record,
			})
			result.ParquetElapsed += time.Since(parquetStarted)
			if err != nil || validateFastSegmentReceipt(receipt, request.TransformPlanHash) != nil {
				if err == nil {
					err = ErrArrowFastPathInvalid
				}
				return fmt.Errorf("close Arrow Parquet segment: %w", err)
			}
			applyCtx, cancelApply := context.WithTimeout(ctx, request.unitDeadline())
			applyStarted := time.Now()
			applyErr := session.ApplyArrowSegment(applyCtx, ArrowBulkApplyRequest{
				ConnectionID: request.ConnectionID, Plan: plan, Segment: receipt, Record: record,
				Runtime: request.DestinationRuntime, Source: request.Source, SourceRuntime: request.SourceRuntime,
				Binding: request.DestinationBinding, Stream: request.Stream, BatchSize: request.BatchSize, Approval: request.Approval,
			})
			result.ApplyElapsed += time.Since(applyStarted)
			cancelApply()
			if applyErr != nil {
				return fmt.Errorf("bulk apply Arrow segment: %w", tagTransportExecutionError(TransportExecutionOriginDestination, applyErr))
			}
			if reporter, ok := session.(ArrowBulkPhaseReporter); ok {
				result.IndexConstraintElapsed = reporter.ArrowBulkPhaseMeasurement().IndexConstraintBuildElapsed
			}
			segments = append(segments, receipt)
			result.RecordsStaged += int(receipt.TransformedRows)
			result.RecordsApplied += int(receipt.TransformedRows)
			result.TransformedBytes += receipt.TransformedBytes
			result.ParquetBytes += receipt.ParquetBytes
		}
		if err := output.Err(); err != nil {
			return fmt.Errorf("iterate transformed Arrow range: %w", err)
		}
		return nil
	})
	if err != nil {
		return result, tagTransportExecutionError(TransportExecutionOriginSource, err)
	}
	if lastCandidate == nil {
		if _, allowed := resolved.Source.(EmptyResultSource); !allowed {
			return result, fmt.Errorf("Arrow source transport completed without a checkpoint candidate")
		}
	}
	if request.Approval.AuthorizeNextUnit != nil {
		if err := request.Approval.AuthorizeNextUnit(ctx); err != nil {
			return result, fmt.Errorf("authorize Arrow full-overwrite publication: %w", err)
		}
	}
	publishCtx, cancelPublish := context.WithTimeout(ctx, request.unitDeadline())
	publishStarted := time.Now()
	acknowledgement, publishErr := session.PublishArrowFullOverwrite(publishCtx, ArrowFullOverwritePublicationRequest{
		LastCheckpoint: lastCandidate, Segments: append([]FastSegmentReceipt(nil), segments...),
		SourceLogicalBytes: result.SourceLogicalBytes, SourceRows: int64(result.RecordsRead),
		TransformedRows: int64(result.RecordsApplied), TransformedBytes: result.TransformedBytes,
	})
	result.PublishElapsed += time.Since(publishStarted)
	if publishErr == nil {
		readBackStarted := time.Now()
		publishErr = session.ReadBackArrowFullOverwrite(publishCtx, acknowledgement)
		result.PublishElapsed += time.Since(readBackStarted)
	}
	cancelPublish()
	if publishErr != nil {
		return result, fmt.Errorf("publish Arrow full-overwrite run: %w", tagTransportExecutionError(TransportExecutionOriginDestination, publishErr))
	}
	if reporter, ok := session.(ArrowBulkPhaseReporter); ok {
		result.IndexConstraintElapsed = reporter.ArrowBulkPhaseMeasurement().IndexConstraintBuildElapsed
	}
	published = true
	if lastCandidate == nil {
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

func cloneArrowExtractRequest(request ArrowExtractRequest) ArrowExtractRequest {
	clone := request
	clone.Runtime = cloneRuntimeConfig(request.Runtime)
	clone.PrimaryKey = append([]string(nil), request.PrimaryKey...)
	clone.Resume.SourceGeneration = append(synccontract.OpaqueToken(nil), request.Resume.SourceGeneration...)
	if request.Checkpoint != nil {
		checkpoint := request.Checkpoint.Clone()
		clone.Checkpoint = &checkpoint
	}
	return clone
}

func validateArrowSourceBatch(batch ArrowSourceBatch, maximumRecords int) error {
	if batch.Record == nil || batch.SourceLogicalBytes < 0 || batch.SourceRows < batch.Record.NumRows() || batch.SourceRows > int64(maximumRecords) || batch.ExtractElapsed < 0 {
		return ErrArrowFastPathInvalid
	}
	if err := batch.CandidateCheckpoint.Validate(); err != nil {
		return fmt.Errorf("Arrow source candidate checkpoint: %w", err)
	}
	return nil
}

func validateFastSegmentReceipt(receipt FastSegmentReceipt, transformHash string) error {
	if receipt.ID == "" || receipt.SchemaHash == "" || receipt.ContentSHA256 == "" || receipt.ParquetSHA256 == "" || receipt.TransformPlanHash != transformHash || receipt.SourceLogicalBytes < 0 || receipt.SourceRows < 0 || receipt.TransformedRows < 0 || receipt.TransformedBytes < 0 || receipt.ParquetBytes < 1 {
		return ErrArrowFastPathInvalid
	}
	return nil
}

func arrowRecordBytes(record arrow.Record) int64 {
	var total uint64
	for index := 0; index < int(record.NumCols()); index++ {
		total += record.Column(index).Data().SizeInBytes()
	}
	if total > uint64(^uint64(0)>>1) {
		return int64(^uint64(0) >> 1)
	}
	return int64(total)
}

func newArrowSegmentID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", ErrArrowFastPathInvalid
	}
	return "segment-" + hex.EncodeToString(value[:]), nil
}
