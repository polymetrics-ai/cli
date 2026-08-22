package app

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"polymetrics.ai/internal/safety"
	"polymetrics.ai/internal/synctransport"
)

// deliveredReconciliationFor returns the one already-durable transport result
// that must be repaired before this connection/stream can start any new work.
// It intentionally consults only persisted state; it does not resolve an
// endpoint, credential, operation, or provider route.
func (a *App) deliveredReconciliationFor(connection, stream string) (Run, bool) {
	for index := len(a.state.Runs) - 1; index >= 0; index-- {
		run := a.state.Runs[index]
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
	return a.persistDeliveredReconciliationRunWithCompleted(runID, result, runErr, acknowledged, nil)
}

func (a *App) persistCompletedDeliveredReconciliationRun(completed Run, runID string, result etlExecutionResult, runErr error, acknowledged *acknowledgedTransportCompletion) (Run, error) {
	return a.persistDeliveredReconciliationRunWithCompleted(runID, result, runErr, acknowledged, &completed)
}

func (a *App) persistDeliveredReconciliationRunWithCompleted(runID string, result etlExecutionResult, runErr error, acknowledged *acknowledgedTransportCompletion, completed *Run) (Run, error) {
	if result.PendingStreamState == nil || result.DeliveryReconciliation == nil || acknowledged == nil {
		return a.failRunWithResult(runID, result, runErr)
	}
	if completed != nil && (completed.ID != runID || completed.Status != "completed") {
		return Run{}, errors.Join(runErr, fmt.Errorf("acknowledged transport run %q is not an exact completed terminal result", runID))
	}
	expectedRevision := a.state.Revision
	completedAt := time.Now().UTC()
	transitionedInCallback := false
	updated, persistErr := a.updateState(func(current state) (state, error) {
		rebased := current.Revision != expectedRevision
		if rebased {
			expectedStreamState := acknowledged.state
			if completed != nil {
				expectedStreamState = result.PendingStreamState.State
			}
			currentStreamState, present := current.StreamStates[acknowledged.key]
			if !present || !transportStreamStateEqual(currentStreamState, expectedStreamState) {
				return current, fmt.Errorf("acknowledged transport stream state changed before delivered reconciliation: %w", errStateRevisionConflict)
			}
		}
		for i := range current.Runs {
			if current.Runs[i].ID != runID {
				continue
			}
			if completed != nil && !reflect.DeepEqual(current.Runs[i], *completed) {
				return current, fmt.Errorf("acknowledged transport run %q changed before delivered reconciliation: %w", runID, errStateRevisionConflict)
			}
			if completed == nil && current.Runs[i].Status != "running" {
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
			current.Runs[i].Error = safety.RedactErrorText(runErr.Error())
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
		current.StreamStates[result.PendingStreamState.Key] = cloneStreamState(result.PendingStreamState.State)
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
	if (postgresPlanID == "") != (reconciliation.PostgresManagedTargetPlanID == "") || (declarativePlanID == "") != (reconciliation.DeclarativeTypedDestinationPlanID == "") {
		return errors.New("delivered reconciliation plan identity is invalid")
	}
	if postgresPlanID != "" && declarativePlanID != "" {
		return errors.New("delivered reconciliation has conflicting declaration-owned marker identities")
	}
	if !reconciliation.StageRetirement && postgresPlanID == "" && declarativePlanID == "" && reconciliation.EmptyPublication == nil {
		return errors.New("delivered reconciliation has no repair action")
	}
	return nil
}
