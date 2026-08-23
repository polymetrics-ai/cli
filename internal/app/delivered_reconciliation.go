package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"polymetrics.ai/internal/safety"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// deliveredReconciliationFor returns the one already-durable transport result
// that must be repaired before this connection/stream can start any new work.
// It intentionally consults only persisted state; it does not resolve an
// endpoint, credential, operation, or provider route.
func (a *App) deliveredReconciliationFor(connection, stream string) (Run, bool) {
	return deliveredReconciliationForState(a.state, connection, stream)
}

func deliveredReconciliationForState(loaded state, connection, stream string) (Run, bool) {
	for index := len(loaded.Runs) - 1; index >= 0; index-- {
		run := loaded.Runs[index]
		if run.Connection == connection && run.Stream == stream && run.Status == ETLRunStatusDeliveredReconciliationRequired && run.DeliveryReconciliation != nil {
			return run, true
		}
	}
	return Run{}, false
}

// persistDeliveredReconciliationRun seals the exact provider receipt and
// acknowledged checkpoint as a terminal run before returning the local
// bookkeeping error. The terminal record is the anti-replay fence: future
// calls can repair only the declared cleanup named in it.
func (a *App) persistDeliveredReconciliationRun(runID string, result etlExecutionResult, runErr error, acknowledged *acknowledgedTransportCompletion) (Run, error) {
	if result.PendingStreamState == nil || result.DeliveryReconciliation == nil || acknowledged == nil {
		if runErr == nil {
			return Run{}, errors.New("delivered reconciliation is missing terminal transport state")
		}
		return a.failRunWithResult(runID, result, runErr)
	}
	expectedRevision := a.state.Revision
	completedAt := time.Now().UTC()
	pendingStreamState := cloneStreamState(result.PendingStreamState.State)
	nextFence, err := nextTransportWorkFence(pendingStreamState.ActiveWorkFence)
	if err != nil {
		return Run{}, err
	}
	pendingStreamState.ActiveWorkFence = nextFence
	transitionedInCallback := false
	updated, persistErr := a.updateState(func(current state) (state, error) {
		rebased := current.Revision != expectedRevision
		if rebased {
			currentStreamState, present := current.StreamStates[acknowledged.key]
			if !present || !transportStreamStateEqual(currentStreamState, acknowledged.state) {
				return current, fmt.Errorf("acknowledged transport stream state changed before delivered reconciliation: %w", errStateRevisionConflict)
			}
		}
		for i := range current.Runs {
			if current.Runs[i].ID != runID {
				continue
			}
			if current.Runs[i].Status != "running" {
				return current, fmt.Errorf("acknowledged transport run %q has status %q, want running before delivered reconciliation: %w", runID, current.Runs[i].Status, errStateRevisionConflict)
			}
			current.Runs[i].Status = ETLRunStatusDeliveredReconciliationRequired
			current.Runs[i].RecordsRead = result.RecordsRead
			current.Runs[i].RecordsTransformed = result.RecordsTransformed
			current.Runs[i].RecordsLoaded = result.RecordsLoaded
			current.Runs[i].RecordsFailed = result.RecordsFailed
			current.Runs[i].BatchCount = result.BatchCount
			current.Runs[i].Checkpoint = cloneStringMap(result.Checkpoint)
			current.Runs[i].TransportPhaseMeasurement = cloneTransportPhaseMeasurement(result.TransportPhaseMeasurement)
			current.Runs[i].DestinationResults = cloneDestinationResults(result.DestinationResults)
			current.Runs[i].DeliveryReconciliation = cloneDeliveryReconciliation(result.DeliveryReconciliation)
			if runErr == nil {
				current.Runs[i].Error = ""
			} else {
				current.Runs[i].Error = safety.RedactErrorText(runErr.Error())
			}
			current.Runs[i].CompletedAt = completedAt
			transitionedInCallback = true
			break
		}
		if !transitionedInCallback {
			return current, fmt.Errorf("acknowledged transport run %q not found before delivered reconciliation: %w", runID, errStateRevisionConflict)
		}
		if current.Checkpoints == nil {
			current.Checkpoints = map[string]map[string]string{}
		}
		current.Checkpoints[runID] = cloneStringMap(result.Checkpoint)
		if current.StreamStates == nil {
			current.StreamStates = map[string]StreamState{}
		}
		current.StreamStates[result.PendingStreamState.Key] = cloneStreamState(pendingStreamState)
		return current, nil
	})
	if persistErr != nil {
		if transitionedInCallback && stateStoreCommitMayHaveSucceeded(persistErr) {
			reloaded, reloadErr := a.reloadExactTerminalState(persistErr)
			if reloadErr == nil {
				if durableRun, terminalErr := terminalETLRunFromState(reloaded.State, runID); terminalErr == nil {
					return durableRun, errors.Join(runErr, fmt.Errorf("persist delivered reconciliation run: %w", persistErr))
				}
			}
		}
		return Run{}, errors.Join(runErr, fmt.Errorf("persist delivered reconciliation run: %w", persistErr))
	}
	for _, run := range updated.Runs {
		if run.ID == runID {
			return run, runErr
		}
	}
	return Run{}, errors.Join(runErr, fmt.Errorf("delivered reconciliation run %q was not stored", runID))
}

func (a *App) persistDeliveredReconciliationRunWithWorkLease(ctx context.Context, lease *transportWorkLease, runID string, result etlExecutionResult) (Run, error) {
	return a.persistLeaseOwnedDeliveredReconciliation(ctx, lease, runID, result, false)
}

func (a *App) persistPendingEmptyPublicationReadBackWithWorkLease(ctx context.Context, lease *transportWorkLease, runID string, result etlExecutionResult) (Run, error) {
	if result.DeliveryReconciliation == nil || result.DeliveryReconciliation.EmptyPublicationReadBackPending == nil || result.DeliveryReconciliation.EmptyPublication != nil {
		return Run{}, errors.New("lease-owned empty publication read-back receipt is invalid")
	}
	return a.persistLeaseOwnedDeliveredReconciliation(ctx, lease, runID, result, true)
}

func (a *App) persistLeaseOwnedDeliveredReconciliation(ctx context.Context, lease *transportWorkLease, runID string, result etlExecutionResult, retainLease bool) (Run, error) {
	if a == nil || lease == nil || lease.app != a || result.PendingStreamState == nil || result.DeliveryReconciliation == nil || result.PendingStreamState.Key != lease.key {
		return Run{}, errors.New("lease-owned delivered reconciliation is invalid")
	}
	persistCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), transportWorkLeaseDuration)
	defer cancel()
	if err := persistCtx.Err(); err != nil {
		return Run{}, err
	}
	lease.mu.Lock()
	defer lease.mu.Unlock()
	pendingStreamState := cloneStreamState(result.PendingStreamState.State)
	if pendingStreamState.ActiveWorkFence != lease.fence {
		return Run{}, errTransportStreamWorkFenceLost
	}
	if !retainLease {
		nextFence, err := nextTransportWorkFence(lease.fence)
		if err != nil {
			return Run{}, err
		}
		pendingStreamState.ActiveWorkID = ""
		pendingStreamState.ActiveWorkLeaseUntil = nil
		pendingStreamState.ActiveWorkFence = nextFence
	}
	completedAt := time.Now().UTC()
	transitionedInCallback := false
	updated, persistErr := a.updateState(func(current state) (state, error) {
		if err := persistCtx.Err(); err != nil {
			return current, err
		}
		now := transportWorkLeaseNow()
		actual, present := current.StreamStates[lease.key]
		if !present || actual.ActiveWorkID != lease.workID || actual.ActiveWorkFence != lease.fence || actual.ActiveWorkLeaseUntil == nil || !actual.ActiveWorkLeaseUntil.After(now) {
			return current, errTransportStreamWorkFenceLost
		}
		if retainLease {
			until := now.Add(transportWorkLeaseDuration)
			pendingStreamState.ActiveWorkID = lease.workID
			pendingStreamState.ActiveWorkFence = lease.fence
			pendingStreamState.ActiveWorkLeaseUntil = &until
		}
		for i := range current.Runs {
			if current.Runs[i].ID != runID {
				continue
			}
			if retainLease {
				if current.Runs[i].Status != "running" {
					return current, fmt.Errorf("lease-owned transport run %q has status %q, want running before empty publication read-back persistence: %w", runID, current.Runs[i].Status, errStateRevisionConflict)
				}
			} else if current.Runs[i].Status != "running" && !canFinalizePendingEmptyPublicationReadBack(current.Runs[i], result) {
				return current, fmt.Errorf("lease-owned transport run %q has status %q, want running or pending empty publication read-back before delivered reconciliation: %w", runID, current.Runs[i].Status, errStateRevisionConflict)
			}
			current.Runs[i].Status = ETLRunStatusDeliveredReconciliationRequired
			current.Runs[i].RecordsRead = result.RecordsRead
			current.Runs[i].RecordsTransformed = result.RecordsTransformed
			current.Runs[i].RecordsLoaded = result.RecordsLoaded
			current.Runs[i].RecordsFailed = result.RecordsFailed
			current.Runs[i].BatchCount = result.BatchCount
			current.Runs[i].Checkpoint = cloneStringMap(result.Checkpoint)
			current.Runs[i].TransportPhaseMeasurement = cloneTransportPhaseMeasurement(result.TransportPhaseMeasurement)
			current.Runs[i].DestinationResults = cloneDestinationResults(result.DestinationResults)
			current.Runs[i].DeliveryReconciliation = cloneDeliveryReconciliation(result.DeliveryReconciliation)
			current.Runs[i].Error = ""
			current.Runs[i].CompletedAt = completedAt
			transitionedInCallback = true
			break
		}
		if !transitionedInCallback {
			return current, fmt.Errorf("lease-owned transport run %q not found before delivered reconciliation: %w", runID, errStateRevisionConflict)
		}
		if current.Checkpoints == nil {
			current.Checkpoints = map[string]map[string]string{}
		}
		current.Checkpoints[runID] = cloneStringMap(result.Checkpoint)
		if current.StreamStates == nil {
			current.StreamStates = map[string]StreamState{}
		}
		current.StreamStates[lease.key] = cloneStreamState(pendingStreamState)
		return current, nil
	})
	if persistErr != nil {
		if transitionedInCallback && stateStoreCommitMayHaveSucceeded(persistErr) {
			reloaded, reloadErr := a.reloadExactTerminalState(persistErr)
			if reloadErr == nil {
				if durableRun, terminalErr := terminalETLRunFromState(reloaded.State, runID); terminalErr == nil {
					return durableRun, fmt.Errorf("persist lease-owned delivered reconciliation run: %w", persistErr)
				}
			}
		}
		return Run{}, fmt.Errorf("persist lease-owned delivered reconciliation run: %w", persistErr)
	}
	lease.state = cloneStreamState(pendingStreamState)
	for _, run := range updated.Runs {
		if run.ID == runID {
			return run, nil
		}
	}
	return Run{}, fmt.Errorf("lease-owned delivered reconciliation run %q was not stored", runID)
}

func canFinalizePendingEmptyPublicationReadBack(run Run, result etlExecutionResult) bool {
	if run.Status != ETLRunStatusDeliveredReconciliationRequired || run.DeliveryReconciliation == nil || run.DeliveryReconciliation.EmptyPublicationReadBackPending == nil || result.DeliveryReconciliation == nil || result.DeliveryReconciliation.EmptyPublication == nil || result.DeliveryReconciliation.EmptyPublicationReadBackPending != nil {
		return false
	}
	pending := run.DeliveryReconciliation.EmptyPublicationReadBackPending
	final := result.DeliveryReconciliation.EmptyPublication
	return pending.Witness.Sink == final.Sink && pending.Witness.AcknowledgedAt.Equal(final.AcknowledgedAt)
}

// reconcileDeliveredTransportRun repairs only the local, recorded aftermath of
// a delivery. It runs before endpoint resolution and therefore cannot replay
// source reads, destination writes, or a declaration-selected action.
func (a *App) reconcileDeliveredTransportRun(ctx context.Context, run Run) (Run, error) {
	if run.Status != ETLRunStatusDeliveredReconciliationRequired || run.DeliveryReconciliation == nil {
		return Run{}, fmt.Errorf("run %q is not a delivered reconciliation", run.ID)
	}
	reconciliation := cloneDeliveryReconciliation(run.DeliveryReconciliation)
	if err := validateDeliveryReconciliation(reconciliation); err != nil {
		return run, synctransport.NewDeliveredReconciliationRequiredError(err)
	}
	if reconciliation.EmptyPublicationReadBackPending != nil {
		readBack, err := a.reconcilePendingEmptyPublicationReadBack(ctx, run, *reconciliation.EmptyPublicationReadBackPending)
		if err != nil {
			return run, synctransport.NewDeliveredReconciliationRequiredError(err)
		}
		run = readBack
		reconciliation = cloneDeliveryReconciliation(run.DeliveryReconciliation)
		if err := validateDeliveryReconciliation(reconciliation); err != nil {
			return run, synctransport.NewDeliveredReconciliationRequiredError(err)
		}
	}
	if reconciliation.StageRetirement {
		if err := a.reconcileCommittedTransportStages(ctx); err != nil {
			return run, synctransport.NewDeliveredReconciliationRequiredError(err)
		}
	}
	if reconciliation.PostgresManagedTargetPlanID != "" {
		if err := a.markPostgresManagedTargetPlanExecuted(reconciliation.PostgresManagedTargetPlanID); err != nil {
			return run, synctransport.NewDeliveredReconciliationRequiredError(fmt.Errorf("mark PostgreSQL managed target plan executed: %w", err))
		}
	}
	if reconciliation.DeclarativeTypedDestinationPlanID != "" {
		if err := a.markDeclarativeTypedDestinationPlanExecuted(reconciliation.DeclarativeTypedDestinationPlanID); err != nil {
			return run, synctransport.NewDeliveredReconciliationRequiredError(fmt.Errorf("mark declarative typed destination plan executed: %w", err))
		}
	}

	expectedRevision := a.state.Revision
	completedAt := time.Now().UTC()
	updated, persistErr := a.updateState(func(current state) (state, error) {
		if current.Revision != expectedRevision {
			return current, errStateRevisionConflict
		}
		for i := range current.Runs {
			if current.Runs[i].ID != run.ID {
				continue
			}
			if current.Runs[i].Status != ETLRunStatusDeliveredReconciliationRequired || current.Runs[i].DeliveryReconciliation == nil {
				return current, fmt.Errorf("delivered reconciliation run %q changed before repair: %w", run.ID, errStateRevisionConflict)
			}
			current.Runs[i].Status = "completed"
			current.Runs[i].DeliveryReconciliation = nil
			current.Runs[i].Error = ""
			current.Runs[i].CompletedAt = completedAt
			return current, nil
		}
		return current, fmt.Errorf("delivered reconciliation run %q not found", run.ID)
	})
	if persistErr != nil {
		if stateStoreCommitMayHaveSucceeded(persistErr) {
			reloaded, reloadErr := a.reloadExactTerminalState(persistErr)
			if reloadErr == nil {
				if durableRun, terminalErr := terminalETLRunFromState(reloaded.State, run.ID); terminalErr == nil {
					return durableRun, persistErr
				}
			}
		}
		return Run{}, fmt.Errorf("persist delivered reconciliation repair: %w", persistErr)
	}
	for _, stored := range updated.Runs {
		if stored.ID == run.ID {
			return stored, nil
		}
	}
	return Run{}, fmt.Errorf("repaired run %q was not stored", run.ID)
}

func (a *App) reconcilePendingEmptyPublicationReadBack(ctx context.Context, run Run, receipt synctransport.EmptyPublicationReadBackReceipt) (Run, error) {
	if a == nil || a.transports == nil || receipt.Validate() != nil {
		return Run{}, errors.New("empty publication read-back reconciliation is invalid")
	}
	if err := a.admitPendingEmptyPublicationReadBack(run); err != nil {
		return Run{}, err
	}
	conn, found := a.findConnection(run.Connection)
	if !found {
		return Run{}, fmt.Errorf("connection %q is unavailable for empty publication read-back", run.Connection)
	}
	stream, found := conn.Streams[run.Stream]
	if !found {
		return Run{}, fmt.Errorf("stream %q is unavailable for empty publication read-back", run.Stream)
	}
	mode, err := ParseStreamSyncMode(stream)
	if err != nil || mode.ContractMode != synccontract.ModeFullOverwrite {
		return Run{}, fmt.Errorf("stream %q cannot recover an empty full-overwrite publication", run.Stream)
	}
	source, _, sourceRuntime, err := a.resolveEndpointWithCredential(ctx, conn.Source)
	if err != nil {
		return Run{}, err
	}
	destination, destinationRuntime, err := a.resolveEndpoint(ctx, conn.Destination)
	if err != nil {
		return Run{}, err
	}
	if destination.Name() != receipt.Witness.Sink {
		return Run{}, fmt.Errorf("empty publication read-back destination %q does not match receipt sink %q", destination.Name(), receipt.Witness.Sink)
	}
	started := time.Now()
	err = synctransport.NewOrchestrator(a.transports).ReadBackEmptyFullOverwrite(ctx, synctransport.EmptyPublicationReadBackRequest{
		Runtime:           destinationRuntime,
		Source:            source,
		SourceRuntime:     sourceRuntime,
		Destination:       destination,
		Binding:           synctransport.DestinationBinding{WorkspaceID: a.state.WorkspaceID, SourceConnectorID: source.Name(), ConnectionID: conn.ID, StreamID: stream.StreamID, PrimaryKey: append([]string(nil), stream.PrimaryKey...)},
		Stream:            run.Stream,
		DestinationAction: stream.DestinationAction,
		TransformPlanJSON: stream.TransformPlan,
		TransformPlanHash: stream.TransformPlanHash,
		Receipt:           receipt.Clone(),
	})
	if err != nil {
		return Run{}, sanitizeRuntimeError(err, sourceRuntime, destinationRuntime)
	}
	return a.recordEmptyPublicationReadBack(run, receipt, time.Since(started))
}

func (a *App) admitPendingEmptyPublicationReadBack(run Run) error {
	key := streamStateKey(run.Connection, run.Stream)
	streamState, present := a.state.StreamStates[key]
	if !present {
		return fmt.Errorf("empty publication stream state is unavailable: %w", errStateRevisionConflict)
	}
	if streamState.ActiveWorkID == "" {
		return nil
	}
	if streamState.ActiveWorkID != run.ID {
		return fmt.Errorf("empty publication work lease changed before read-back repair: %w", errStateRevisionConflict)
	}
	if streamState.ActiveWorkLeaseUntil != nil && streamState.ActiveWorkLeaseUntil.After(transportWorkLeaseNow()) {
		return errTransportStreamWorkInProgress
	}
	return nil
}

func (a *App) recordEmptyPublicationReadBack(run Run, receipt synctransport.EmptyPublicationReadBackReceipt, elapsed time.Duration) (Run, error) {
	if receipt.Validate() != nil {
		return Run{}, errors.New("empty publication read-back receipt is invalid")
	}
	final := cloneDeliveryReconciliation(run.DeliveryReconciliation)
	if final == nil || final.EmptyPublicationReadBackPending == nil || !emptyPublicationReadBackReceiptEqual(*final.EmptyPublicationReadBackPending, receipt) {
		return Run{}, errors.New("empty publication read-back reconciliation changed")
	}
	witness := receipt.Witness
	final.EmptyPublication = &witness
	final.EmptyPublicationReadBackPending = nil
	measurement := cloneTransportPhaseMeasurement(run.TransportPhaseMeasurement)
	if measurement == nil {
		measurement = &TransportPhaseMeasurement{}
	}
	measurement.ReadBackElapsedNanos += elapsed.Nanoseconds()
	key := streamStateKey(run.Connection, run.Stream)
	expectedRevision := a.state.Revision
	updated, persistErr := a.updateState(func(current state) (state, error) {
		if current.Revision != expectedRevision {
			return current, errStateRevisionConflict
		}
		stateValue, present := current.StreamStates[key]
		if !present {
			return current, fmt.Errorf("empty publication stream state is unavailable: %w", errStateRevisionConflict)
		}
		if stateValue.ActiveWorkID != "" {
			if stateValue.ActiveWorkID != run.ID {
				return current, fmt.Errorf("empty publication work lease changed before read-back repair: %w", errStateRevisionConflict)
			}
			if stateValue.ActiveWorkLeaseUntil != nil && stateValue.ActiveWorkLeaseUntil.After(transportWorkLeaseNow()) {
				return current, errTransportStreamWorkInProgress
			}
			nextFence, err := nextTransportWorkFence(stateValue.ActiveWorkFence)
			if err != nil {
				return current, err
			}
			stateValue.ActiveWorkID = ""
			stateValue.ActiveWorkLeaseUntil = nil
			stateValue.ActiveWorkFence = nextFence
		}
		stateValue.Connection = run.Connection
		stateValue.Stream = run.Stream
		stateValue.LastSuccessfulRunID = run.ID
		stateValue.RecordsLoaded = 0
		stateValue.UpdatedAt = witness.AcknowledgedAt
		for i := range current.Runs {
			if current.Runs[i].ID != run.ID {
				continue
			}
			if current.Runs[i].Status != ETLRunStatusDeliveredReconciliationRequired || current.Runs[i].DeliveryReconciliation == nil || current.Runs[i].DeliveryReconciliation.EmptyPublicationReadBackPending == nil || !emptyPublicationReadBackReceiptEqual(*current.Runs[i].DeliveryReconciliation.EmptyPublicationReadBackPending, receipt) {
				return current, fmt.Errorf("empty publication reconciliation changed before read-back repair: %w", errStateRevisionConflict)
			}
			current.Runs[i].TransportPhaseMeasurement = cloneTransportPhaseMeasurement(measurement)
			current.Runs[i].DeliveryReconciliation = cloneDeliveryReconciliation(final)
			current.StreamStates[key] = cloneStreamState(stateValue)
			return current, nil
		}
		return current, fmt.Errorf("empty publication reconciliation run %q not found", run.ID)
	})
	if persistErr != nil {
		if stateStoreCommitMayHaveSucceeded(persistErr) {
			reloaded, reloadErr := a.reloadExactTerminalState(persistErr)
			if reloadErr == nil {
				if durable, terminalErr := terminalETLRunFromState(reloaded.State, run.ID); terminalErr == nil {
					return durable, persistErr
				}
			}
		}
		return Run{}, fmt.Errorf("persist empty publication read-back repair: %w", persistErr)
	}
	for _, stored := range updated.Runs {
		if stored.ID == run.ID {
			return stored, nil
		}
	}
	return Run{}, fmt.Errorf("empty publication read-back repair %q was not stored", run.ID)
}

func emptyPublicationReadBackReceiptEqual(left, right synctransport.EmptyPublicationReadBackReceipt) bool {
	return left.Witness.Sink == right.Witness.Sink && left.Witness.AcknowledgedAt.Equal(right.Witness.AcknowledgedAt) && bytes.Equal(left.Output, right.Output)
}

func validateDeliveryReconciliation(reconciliation *DeliveryReconciliation) error {
	if reconciliation == nil || reconciliation.State != ETLRunStatusDeliveredReconciliationRequired {
		return errors.New("delivered reconciliation state is invalid")
	}
	postgresPlanID := strings.TrimSpace(reconciliation.PostgresManagedTargetPlanID)
	declarativePlanID := strings.TrimSpace(reconciliation.DeclarativeTypedDestinationPlanID)
	if reconciliation.EmptyPublication != nil {
		if err := reconciliation.EmptyPublication.Validate(); err != nil {
			return fmt.Errorf("delivered reconciliation empty publication witness is invalid: %w", err)
		}
	}
	if reconciliation.EmptyPublicationReadBackPending != nil {
		if err := reconciliation.EmptyPublicationReadBackPending.Validate(); err != nil {
			return fmt.Errorf("delivered reconciliation empty publication read-back receipt is invalid: %w", err)
		}
	}
	if reconciliation.EmptyPublication != nil && reconciliation.EmptyPublicationReadBackPending != nil {
		return errors.New("delivered reconciliation has both pending and completed empty publication receipts")
	}
	if (postgresPlanID == "") != (reconciliation.PostgresManagedTargetPlanID == "") || (declarativePlanID == "") != (reconciliation.DeclarativeTypedDestinationPlanID == "") {
		return errors.New("delivered reconciliation plan identity is invalid")
	}
	if postgresPlanID != "" && declarativePlanID != "" {
		return errors.New("delivered reconciliation has conflicting declaration-owned marker identities")
	}
	if !reconciliation.StageRetirement && postgresPlanID == "" && declarativePlanID == "" && reconciliation.EmptyPublication == nil && reconciliation.EmptyPublicationReadBackPending == nil {
		return errors.New("delivered reconciliation has no repair action")
	}
	return nil
}
