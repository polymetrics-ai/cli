package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

var errTransportStreamStateConflict = errors.New("transport stream state changed in another process")

var errTransportStreamWorkInProgress = errors.New("transport stream work is already owned by another process")
var errTransportStreamWorkFenceLost = errors.New("transport stream work fence was lost before I/O")

func hasDeclaredSyncTransport(source, destination connectors.Connector) bool {
	_, sourceDeclared := connectors.SyncTransportDescriptorOf(source)
	_, destinationDeclared := connectors.SyncTransportDescriptorOf(destination)
	// A transport is a closed pair: registry preflight needs a source and a
	// destination declaration. Requiring both here prevents a newly declared
	// primitive destination from rerouting an existing legacy source before that
	// source has declared and registered its own transport role. Invalid
	// two-sided declarations still reach preflight and fail closed there.
	return sourceDeclared && destinationDeclared
}

type transportRouteReason string

const (
	transportRouteDeclarationAbsent transportRouteReason = "declaration_absent"
	transportRouteDeclared          transportRouteReason = "declared_transport"
)

// shouldRunTransport keeps the closed issue-label transport opt-in at the
// persisted connection boundary. Its definition owns the admitted source
// executors, streams, and record mappings. Open installs the exact
// definition-owned composition so it can be preflighted, but that must not
// turn every existing JSON ETL connection into a transport run merely because
// its connector advertises a closed descriptor.
//
// Other externally declared transport pairs retain the normal descriptor-led
// dispatch. The walking slice participates only when the connection itself
// satisfies its fixed issues/source-issue/target-issue/label contract. The
// definition selects the allowed destination action for each declared mode;
// non-additive modes are additionally gated by the persisted connection.
func (a *App) shouldRunTransport(conn Connection, streamName string, mode SyncMode, source, destination connectors.Connector) bool {
	selected, _, _ := a.selectTransportRoute(conn, streamName, mode, source, destination)
	return selected
}

func (a *App) selectTransportRoute(conn Connection, streamName string, mode SyncMode, source, destination connectors.Connector) (bool, transportRouteReason, error) {
	action := conn.Streams[streamName].DestinationAction
	if !hasDeclaredSyncTransport(source, destination) {
		return false, transportRouteDeclarationAbsent, nil
	}
	if a == nil || a.transports == nil {
		return false, transportRouteDeclared, fmt.Errorf("closed transport registry is unavailable")
	}
	sourceDeclarative := isDeclarativeStreamTransportConnector(source)
	destinationDescriptor, destinationDeclared := connectors.DestinationTransportDescriptorOf(destination)
	destinationIssueLabel := destinationDeclared && destinationDescriptor.Executor == issueLabelDestinationReference
	if sourceDeclarative && isBoundedLocalWarehouseLegacyRoute(destination, mode.ContractMode) {
		// The local warehouse primitive owns a closed, bounded ordinary ETL
		// representation for these executable modes. Its two dedupe modes are
		// the only ones promoted to the registered transport executor below;
		// routing the remainder through that executor would either replace their
		// established semantics or reject a mode it does not declare.
		return false, transportRouteDeclarationAbsent, nil
	}
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source: source, Destination: destination, Stream: streamName, Mode: mode.ContractMode, DestinationAction: action,
	})
	if err != nil {
		return false, transportRouteDeclared, err
	}
	if destinationIssueLabel {
	} else if !sourceDeclarative {
		return true, transportRouteDeclared, nil
	} else {
		// A definition-selected source may pair with another definition-selected
		// destination. Its own eligible stream/mode and the destination strategy
		// remain enforced by registry preflight. Only the semantic managed-target
		// destination marker admits this cross-definition route, so legacy API-to-
		// warehouse connections are not diverted from their established ETL path.
		_, managedTarget := resolved.Destination.(synctransport.ManagedTargetApprovalDestination)
		_, definitionOwnedApproval := resolved.Destination.(synctransport.DefinitionOwnedApprovalDestination)
		if managedTarget || definitionOwnedApproval {
			return true, transportRouteDeclared, nil
		}
		materializer, localWarehouse := destination.(connectors.LocalWarehouseMaterializer)
		if localWarehouse && materializer.MaterializesLocalWarehouse() && isWarehouseDedupeContractMode(mode.ContractMode) {
			return true, transportRouteDeclared, nil
		}
		return false, transportRouteDeclared, &synctransport.DeclaredDestinationRouteError{
			Destination: destination.Name(), Executor: destinationDescriptor.Executor,
		}
	}
	transportConn, err := a.issueLabelTransportConnection(conn.ID)
	if err != nil {
		return false, transportRouteDeclared, err
	}
	configuredStream, _, err := issueLabelTransportStream(transportConn)
	if err != nil {
		return false, transportRouteDeclared, err
	}
	if streamName != configuredStream {
		return false, transportRouteDeclared, fmt.Errorf("declared issue-label transport stream %q does not match configured stream %q", streamName, configuredStream)
	}
	contract, err := a.issueLabelTransportContract(transportConn)
	if err != nil {
		return false, transportRouteDeclared, err
	}
	_, err = contract.actionForSyncMode(mode.ContractMode)
	if err != nil {
		return false, transportRouteDeclared, err
	}
	return true, transportRouteDeclared, nil
}

func isWarehouseDedupeContractMode(mode synccontract.Mode) bool {
	return mode == synccontract.ModeIncrementalDedupe || mode == synccontract.ModeIncrementalDedupeHistory
}

// isBoundedLocalWarehouseLegacyRoute recognizes only the local warehouse's
// declaration-owned durable materializer. It is intentionally structural: no
// connector-name branch or generic write capability can select this fallback.
func isBoundedLocalWarehouseLegacyRoute(destination connectors.Connector, mode synccontract.Mode) bool {
	if isWarehouseDedupeContractMode(mode) || !isLocalWarehouseDestination(destination) {
		return false
	}
	materializer, ok := destination.(connectors.LocalWarehouseMaterializer)
	return ok && materializer.MaterializesLocalWarehouse()
}

func isIssueLabelTransportConnector(connector connectors.Connector) bool {
	descriptor, ok := connectors.SyncTransportDescriptorOf(connector)
	if !ok || descriptor.Source == nil || descriptor.Destination == nil {
		return false
	}
	definition, ok := connectors.DefinitionOf(connector)
	if !ok {
		return false
	}
	contract, err := issueLabelTransportContractForDefinition(definition)
	if err != nil {
		return false
	}
	if !isDeclarativeStreamTransportConnector(connector) || descriptor.Destination.Executor != issueLabelDestinationReference {
		return false
	}
	wantModes := contract.modes()
	if len(descriptor.Destination.Modes) != len(wantModes) {
		return false
	}
	for i, mode := range wantModes {
		if descriptor.Destination.Modes[i] != mode || !transportContainsMode(descriptor.Source.Modes, mode) {
			return false
		}
	}
	wantActions := contract.destinationActionNames()
	if len(descriptor.Destination.EligibleActions) != len(wantActions) || len(descriptor.Destination.ApplyStrategies) != len(wantModes) {
		return false
	}
	strategies, err := contract.applyStrategies()
	if err != nil {
		return false
	}
	for i := range wantActions {
		if descriptor.Destination.EligibleActions[i] != wantActions[i] || descriptor.Destination.ApplyStrategies[i] != strategies[i] {
			return false
		}
	}
	return true
}

func isDeclarativeStreamTransportConnector(connector connectors.Connector) bool {
	descriptor, ok := connectors.SyncTransportDescriptorOf(connector)
	if !ok || descriptor.Source == nil || descriptor.Source.Executor != declarativeStreamSourceReference {
		return false
	}
	definition, ok := connectors.DefinitionOf(connector)
	if !ok {
		return false
	}
	return validateDeclarativeStreamEligibility(definition.Streams, descriptor.Source.EligibleStreams) == nil
}

func validateClosedTransportBatchSize(source, destination connectors.Connector, batchSize int) error {
	// The provider-mutating issue-label destination binds every source record
	// to one configured target issue and therefore remains singleton-only. The
	// same definition-selected source may deliver a bounded collection to a
	// semantic managed database destination, whose own executor provides the
	// batch transaction and replay contract.
	if isIssueLabelTransportConnector(destination) && batchSize != 1 {
		return errors.New("closed issue-label transport requires batch size 1")
	}
	if isDeclarativeStreamTransportConnector(source) && batchSize > issueCollectionTransportMaxRecords {
		return fmt.Errorf("bounded declarative collection batch size must not exceed %d", issueCollectionTransportMaxRecords)
	}
	return nil
}

// runTransportETL is the bounded bridge from persisted connection state to
// the transport-neutral orchestrator. It does not call legacy Connector.Read
// or Connector.Write, inject provider metadata, or select a generic `upsert`
// action. Real durable warehouse/apply adapters remain separate foundations;
// this seam therefore fails closed unless a stage and externally verified
// transports are registered.
func (a *App) runTransportETL(ctx context.Context, runID string, conn Connection, source connectors.Connector, sourceRuntime connectors.RuntimeConfig, destination connectors.Connector, destRuntime connectors.RuntimeConfig, sourceExpectation synccontract.ResumeExpectation, streamName string, stream StreamConfig, mode SyncMode, batchSize, maxInFlightBatches int, approval synctransport.DestinationApproval, rateParkingResumeCheckpoint *synccontract.CheckpointEnvelope) (etlExecutionResult, error) {
	emptyResult := etlExecutionResult{TransportPhaseMeasurement: &TransportPhaseMeasurement{}}
	if a.transports == nil {
		return emptyResult, fmt.Errorf("closed transport registry is unavailable")
	}
	if err := a.reconcileCommittedTransportStages(ctx); err != nil {
		return emptyResult, err
	}
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source: source, Destination: destination, Stream: streamName, Mode: mode.ContractMode, DestinationAction: stream.DestinationAction,
	})
	if err != nil {
		return emptyResult, err
	}
	_, requiresManagedTargetApproval := resolved.Destination.(synctransport.ManagedTargetApprovalDestination)
	_, requiresDefinitionOwnedApproval := resolved.Destination.(synctransport.DefinitionOwnedApprovalDestination)
	if err := validateClosedTransportBatchSize(source, destination, batchSize); err != nil {
		return emptyResult, err
	}
	if requiresManagedTargetApproval {
		var err error
		approval, err = a.authorizePostgresManagedTargetTransport(ctx, conn, streamName, approval, destRuntime)
		if err != nil {
			return emptyResult, err
		}
	}
	if requiresDefinitionOwnedApproval {
		var err error
		approval, err = a.authorizeDeclarativeTypedDestinationTransport(ctx, conn, streamName, approval, destRuntime)
		if err != nil {
			return emptyResult, err
		}
	}
	stateKey := streamStateKey(conn.Name, streamName)
	prior := a.state.StreamStates[stateKey]
	prior = cloneStreamState(prior)
	if prior.Checkpoint != nil {
		if err := validateStreamStateResume(prior, sourceExpectation); err != nil {
			return emptyResult, err
		}
	}
	if rateParkingResumeCheckpoint != nil && !transportCheckpointEqual(prior.Checkpoint, rateParkingResumeCheckpoint) {
		return emptyResult, errors.New("rate-limit resume checkpoint changed")
	}
	workLease, err := a.claimTransportWorkLease(ctx, stateKey, conn.Name, streamName, runID, sourceExpectation, mode.IsOverwrite())
	if err != nil {
		return emptyResult, err
	}
	prior = workLease.stateForTerminalRun()
	generationID := prior.GenerationID

	// One declaration-owned approval callback now gates every source/stage/
	// destination unit on the same durable stream fence before it consults the
	// existing authorization. This preserves the original authorization scope
	// while making a takeover reject stale effects without connector branches.
	originalAuthorize := approval.AuthorizeNextUnit
	approval.AuthorizeNextUnit = func(effectCtx context.Context) error {
		if err := workLease.renew(effectCtx); err != nil {
			return err
		}
		if originalAuthorize != nil {
			return originalAuthorize(effectCtx)
		}
		return nil
	}

	var committed *synccontract.CheckpointEnvelope
	transportRequest := synctransport.RunRequest{
		ConnectionID:       conn.ID,
		Generation:         generationID,
		Source:             source,
		SourceRuntime:      sourceRuntime,
		Destination:        destination,
		DestinationRuntime: destRuntime,
		DestinationBinding: synctransport.DestinationBinding{
			WorkspaceID:       a.state.WorkspaceID,
			SourceConnectorID: source.Name(),
			ConnectionID:      conn.ID,
			StreamID:          stream.StreamID,
			PrimaryKey:        append([]string(nil), stream.PrimaryKey...),
		},
		Stream:                    streamName,
		DestinationAction:         stream.DestinationAction,
		CursorField:               stream.CursorField,
		Mode:                      mode.ContractMode,
		BatchSize:                 batchSize,
		MaxInFlightBatches:        maxInFlightBatches,
		TransformPlanJSON:         stream.TransformPlan,
		TransformPlanHash:         stream.TransformPlanHash,
		FastSegments:              &connectionArrowSegmentStore{app: a, connectionID: conn.ID, stream: streamName},
		Resume:                    sourceExpectation,
		Checkpoint:                prior.Checkpoint,
		RateLimitResumeCheckpoint: rateParkingResumeCheckpoint,
		Approval:                  approval,
		Stage:                     a.transportStage,
		SourceAdmission:           workLease.renew,
		Commit: func(checkpoint synccontract.CheckpointEnvelope) error {
			interim := checkpoint.Clone()
			if interim.CommittedAt == nil {
				return fmt.Errorf("closed transport committed checkpoint is missing its acknowledgement timestamp")
			}
			if err := interim.ValidateResume(sourceExpectation); err != nil {
				return err
			}
			_, err := workLease.commitAfterAcknowledgement(ctx, interim)
			if err != nil {
				if errors.Is(err, errTransportStreamWorkFenceLost) {
					err = fmt.Errorf("%w: %w", errTransportStreamStateConflict, err)
				}
				return fmt.Errorf("persist acknowledged transport checkpoint: %w", err)
			}
			copy := interim.Clone()
			committed = &copy
			return nil
		},
	}
	transportResult, err := synctransport.NewOrchestrator(a.transports).Run(ctx, transportRequest)
	if err != nil {
		err = sanitizeRuntimeError(err, sourceRuntime, destRuntime)
	}
	transportMeasurement := transportPhaseMeasurement(transportResult)
	if committed == nil || committed.CommittedAt == nil {
		if err != nil {
			result := etlExecutionResult{
				RecordsRead: transportResult.RecordsRead, RecordsLoaded: transportResult.RecordsApplied,
				BatchCount: transportResult.Pages, TransportPhaseMeasurement: transportMeasurement,
				DestinationResults: cloneDestinationResults(transportResult.DestinationResults),
			}
			// A per-page run that already had a durable checkpoint can be a
			// rate-limit handoff. Keep its exact active lease until parking
			// atomically records the terminal run and releases that lease with the
			// same checkpoint. Every other uncommitted path has no durable handoff
			// evidence and must restore/delete its claim before terminal failure.
			if (mode.IsOverwrite() || prior.Checkpoint == nil) && !stateStoreCommitMayHaveSucceeded(err) {
				if releaseErr := workLease.abandonUncommitted(ctx); releaseErr != nil {
					if errors.Is(releaseErr, errTransportStreamWorkFenceLost) {
						releaseErr = fmt.Errorf("%w: %w", errTransportStreamStateConflict, releaseErr)
					}
					return result, errors.Join(err, releaseErr)
				}
			}
			return result, err
		}
		if transportResult.Pages == 0 && transportResult.RecordsRead == 0 && transportResult.RecordsApplied == 0 {
			if mode.IsOverwrite() && transportResult.EmptyPublication != nil {
				witness := *transportResult.EmptyPublication
				if err := witness.Validate(); err != nil {
					return emptyResult, fmt.Errorf("closed transport returned an invalid empty publication witness: %w", err)
				}
				if witness.Sink != destination.Name() {
					return emptyResult, fmt.Errorf("closed transport empty publication sink %q does not match destination %q", witness.Sink, destination.Name())
				}
				updated := workLease.stateForTerminalRun()
				updated.Connection = conn.Name
				updated.Stream = streamName
				updated.GenerationID = generationID
				updated.LastSuccessfulRunID = runID
				updated.RecordsLoaded = 0
				updated.UpdatedAt = witness.AcknowledgedAt
				deliveryReconciliation := &DeliveryReconciliation{
					State:            ETLRunStatusDeliveredReconciliationRequired,
					EmptyPublication: &witness,
				}
				if requiresManagedTargetApproval {
					deliveryReconciliation.PostgresManagedTargetPlanID = approval.PlanID
				}
				if requiresDefinitionOwnedApproval {
					deliveryReconciliation.DeclarativeTypedDestinationPlanID = approval.PlanID
				}
				result := etlExecutionResult{
					TransportPhaseMeasurement: transportMeasurement,
					DestinationResults:        cloneDestinationResults(transportResult.DestinationResults),
					DeliveryReconciliation:    deliveryReconciliation,
					PendingStreamState:        &pendingStreamState{Key: stateKey, State: updated},
				}
				result.Checkpoint = checkpointForResult(result, mode, stateKey, updated, "", false)
				// A marker records only that the exact declaration-owned approval
				// was consumed. The externally visible empty replacement, read-back
				// receipt, and terminal state were already sealed above, so a marker
				// failure must persist reconciliation rather than abandon and replay.
				if requiresManagedTargetApproval {
					if err := a.markPostgresManagedTargetPlanExecuted(approval.PlanID); err != nil {
						return result, synctransport.NewDeliveredReconciliationRequiredError(fmt.Errorf("mark PostgreSQL managed target plan executed: %w", err))
					}
				}
				if requiresDefinitionOwnedApproval {
					if err := a.markDeclarativeTypedDestinationPlanExecuted(approval.PlanID); err != nil {
						return result, synctransport.NewDeliveredReconciliationRequiredError(fmt.Errorf("mark declarative typed destination plan executed: %w", err))
					}
				}
				return result, nil
			}
			if mode.IsOverwrite() {
				// A full-overwrite that made no provider-visible publication cannot
				// be presented as a successful empty replacement. The only allowed
				// zero-result success is the sealed witness minted after publish and
				// read-back above.
				completionErr := fmt.Errorf("closed transport completed empty full-overwrite without a durable publication witness")
				if releaseErr := workLease.abandonUncommitted(ctx); releaseErr != nil {
					if errors.Is(releaseErr, errTransportStreamWorkFenceLost) {
						releaseErr = fmt.Errorf("%w: %w", errTransportStreamStateConflict, releaseErr)
					}
					return emptyResult, errors.Join(completionErr, releaseErr)
				}
				return emptyResult, completionErr
			}
			if requiresManagedTargetApproval {
				if err := a.markPostgresManagedTargetPlanExecuted(approval.PlanID); err != nil {
					return emptyResult, err
				}
			}
			if requiresDefinitionOwnedApproval {
				if err := a.markDeclarativeTypedDestinationPlanExecuted(approval.PlanID); err != nil {
					return emptyResult, err
				}
			}
			updated := workLease.stateForTerminalRun()
			updated.Connection = conn.Name
			updated.Stream = streamName
			updated.GenerationID = generationID
			updated.LastSuccessfulRunID = runID
			updated.RecordsLoaded = 0
			updated.UpdatedAt = time.Now().UTC()
			result := etlExecutionResult{
				PendingStreamState: &pendingStreamState{Key: stateKey, State: updated}, TransportPhaseMeasurement: transportMeasurement,
				DestinationResults: cloneDestinationResults(transportResult.DestinationResults),
			}
			result.Checkpoint = checkpointForResult(result, mode, stateKey, updated, "", false)
			return result, nil
		}
		completionErr := fmt.Errorf("closed transport completed without a durable committed checkpoint")
		if releaseErr := workLease.abandonUncommitted(ctx); releaseErr != nil {
			if errors.Is(releaseErr, errTransportStreamWorkFenceLost) {
				releaseErr = fmt.Errorf("%w: %w", errTransportStreamStateConflict, releaseErr)
			}
			return emptyResult, errors.Join(completionErr, releaseErr)
		}
		return emptyResult, completionErr
	}

	updated := workLease.stateForTerminalRun()
	updated.Connection = conn.Name
	updated.Stream = streamName
	updated.Checkpoint = committed
	updated.GenerationID = generationID
	updated.LastSuccessfulRunID = runID
	updated.RecordsLoaded = transportResult.RecordsApplied
	updated.UpdatedAt = *committed.CommittedAt
	var deliveryReconciliation *DeliveryReconciliation
	if transportResult.DeliveredReconciliationRequired {
		deliveryReconciliation = &DeliveryReconciliation{
			State:           ETLRunStatusDeliveredReconciliationRequired,
			StageRetirement: true,
		}
		// The provider action and checkpoint are durable, so an approval marker
		// must be repaired with the same persisted terminal evidence rather than
		// by re-entering the source/destination route.
		if requiresManagedTargetApproval {
			deliveryReconciliation.PostgresManagedTargetPlanID = approval.PlanID
		}
		if requiresDefinitionOwnedApproval {
			deliveryReconciliation.DeclarativeTypedDestinationPlanID = approval.PlanID
		}
	}
	result := etlExecutionResult{
		RecordsRead:               transportResult.RecordsRead,
		RecordsLoaded:             transportResult.RecordsApplied,
		BatchCount:                transportResult.Pages,
		TransportPhaseMeasurement: transportMeasurement,
		DestinationResults:        cloneDestinationResults(transportResult.DestinationResults),
		DeliveryReconciliation:    deliveryReconciliation,
		PendingStreamState: &pendingStreamState{
			Key:   stateKey,
			State: updated,
		},
	}
	result.Checkpoint = checkpointForResult(result, mode, stateKey, updated, "", false)
	if err != nil {
		return result, err
	}
	if requiresManagedTargetApproval {
		if err := a.markPostgresManagedTargetPlanExecuted(approval.PlanID); err != nil {
			result.DeliveryReconciliation = &DeliveryReconciliation{
				State:                       ETLRunStatusDeliveredReconciliationRequired,
				PostgresManagedTargetPlanID: approval.PlanID,
			}
			return result, synctransport.NewDeliveredReconciliationRequiredError(fmt.Errorf("mark PostgreSQL managed target plan executed: %w", err))
		}
	}
	if requiresDefinitionOwnedApproval {
		if err := a.markDeclarativeTypedDestinationPlanExecuted(approval.PlanID); err != nil {
			result.DeliveryReconciliation = &DeliveryReconciliation{
				State:                             ETLRunStatusDeliveredReconciliationRequired,
				DeclarativeTypedDestinationPlanID: approval.PlanID,
			}
			return result, synctransport.NewDeliveredReconciliationRequiredError(fmt.Errorf("mark declarative typed destination plan executed: %w", err))
		}
	}
	return result, nil
}

// reconcileCommittedTransportStages retires only previously committed,
// connection-owned worksets before the next closed transport can reach source
// I/O. Ordinary Open deliberately leaves an accepted receipt observable for
// recovery and certification evidence; a stage that needs eager disposal may
// separately opt into synctransport.RetirableWarehouseStage.
func (a *App) reconcileCommittedTransportStages(ctx context.Context) error {
	stage, ok := a.transportStage.(interface{ ReconcileCommitted(context.Context) error })
	if !ok {
		return nil
	}
	if err := stage.ReconcileCommitted(ctx); err != nil {
		return fmt.Errorf("reconcile committed transport stages: %w", err)
	}
	return nil
}

func transportPhaseMeasurement(result synctransport.Result) *TransportPhaseMeasurement {
	measurement := &TransportPhaseMeasurement{
		ExtractedRecords: result.RecordsRead, WarehouseParquetRecords: result.RecordsStaged, PostgreSQLAppliedRecords: result.RecordsApplied,
		ExtractElapsedNanos: result.ExtractElapsed.Nanoseconds(), WarehouseElapsedNanos: result.StageElapsed.Nanoseconds(), PostgreSQLElapsedNanos: result.ApplyElapsed.Nanoseconds(),
		SourceRecords: result.RecordsRead, TransformedRecords: result.RecordsStaged, CopyAppliedRecords: result.RecordsApplied,
		SourceLogicalBytes: result.SourceLogicalBytes, TransformedLogicalBytes: result.TransformedBytes, ParquetBytes: result.ParquetBytes,
		SourceReadElapsedNanos: result.ExtractElapsed.Nanoseconds(), TransformElapsedNanos: result.TransformElapsed.Nanoseconds(),
		ParquetCloseElapsedNanos: result.ParquetElapsed.Nanoseconds(), BinaryCOPYElapsedNanos: result.ApplyElapsed.Nanoseconds(),
		ReadBackElapsedNanos:             result.ReadBackElapsed.Nanoseconds(),
		IndexConstraintBuildElapsedNanos: result.IndexConstraintElapsed.Nanoseconds(),
		PublishReceiptElapsedNanos:       result.PublishElapsed.Nanoseconds(), CheckpointElapsedNanos: result.CheckpointElapsed.Nanoseconds(),
		CriticalPathElapsedNanos: result.WallElapsed.Nanoseconds(), PeakCreditBytes: result.PeakCreditBytes,
		ByteCreditWaitElapsedNanos: result.CreditWaitElapsed.Nanoseconds(),
	}
	if result.SourceLogicalBytes > 0 && result.WallElapsed > 0 {
		seconds := result.WallElapsed.Seconds()
		measurement.InputDecimalMBPerSecond = float64(result.SourceLogicalBytes) / 1_000_000 / seconds
		measurement.InputMiBPerSecond = float64(result.SourceLogicalBytes) / (1024 * 1024) / seconds
	}
	return measurement
}

func transportStreamStateEqual(left, right StreamState) bool {
	if left.Connection != right.Connection ||
		left.Stream != right.Stream ||
		left.GenerationID != right.GenerationID ||
		left.ActiveWorkID != right.ActiveWorkID ||
		left.ActiveWorkFence != right.ActiveWorkFence ||
		!transportOptionalTimeEqual(left.ActiveWorkLeaseUntil, right.ActiveWorkLeaseUntil) ||
		left.LastSuccessfulRunID != right.LastSuccessfulRunID ||
		left.RecordsLoaded != right.RecordsLoaded ||
		!left.UpdatedAt.Equal(right.UpdatedAt) {
		return false
	}
	return transportCheckpointEqual(left.Checkpoint, right.Checkpoint)
}

func transportCheckpointEqual(left, right *synccontract.CheckpointEnvelope) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	if left.StateVersion != right.StateVersion ||
		left.Source.Engine != right.Source.Engine ||
		left.Source.AccountOrCluster != right.Source.AccountOrCluster ||
		left.Source.ObjectScope != right.Source.ObjectScope ||
		left.Mechanism != right.Mechanism ||
		!transportSnapshotBarrierEqual(left.SnapshotBarrier, right.SnapshotBarrier) ||
		!transportCheckpointPositionEqual(left.Position, right.Position) ||
		!synccontract.ContinuationEqual(left.Continuation, right.Continuation) ||
		!transportOptionalBoolEqual(left.PositionObserved, right.PositionObserved) ||
		!transportOpaqueTokenEqual(left.SourceGeneration, right.SourceGeneration) ||
		left.SchemaVersion != right.SchemaVersion ||
		left.ProtocolVersion != right.ProtocolVersion ||
		left.Dedupe.Kind != right.Dedupe.Kind ||
		!transportOpaqueTokenEqual(left.Dedupe.Value, right.Dedupe.Value) ||
		left.DedupeWindow.Kind != right.DedupeWindow.Kind ||
		!transportOpaqueTokenEqual(left.DedupeWindow.Start, right.DedupeWindow.Start) ||
		!transportOpaqueTokenEqual(left.DedupeWindow.End, right.DedupeWindow.End) ||
		!left.ObservedAt.Equal(right.ObservedAt) ||
		!transportOptionalTimeEqual(left.CommittedAt, right.CommittedAt) ||
		(left.Partitions == nil) != (right.Partitions == nil) ||
		len(left.Partitions) != len(right.Partitions) {
		return false
	}
	for index := range left.Partitions {
		if !transportOpaqueTokenEqual(left.Partitions[index].Partition, right.Partitions[index].Partition) ||
			!transportCheckpointPositionEqual(left.Partitions[index].Position, right.Partitions[index].Position) {
			return false
		}
	}
	return true
}

func transportSnapshotBarrierEqual(left, right *synccontract.SnapshotBarrier) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Kind == right.Kind && transportOpaqueTokenEqual(left.Token, right.Token)
}

func transportCheckpointPositionEqual(left, right synccontract.CheckpointPosition) bool {
	return transportOpaqueTokenEqual(left.Primary, right.Primary) && transportOpaqueTokenEqual(left.TieBreaker, right.TieBreaker)
}

func transportOpaqueTokenEqual(left, right synccontract.OpaqueToken) bool {
	return (left == nil) == (right == nil) && bytes.Equal(left, right)
}

func transportOptionalBoolEqual(left, right *bool) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func transportOptionalTimeEqual(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}
