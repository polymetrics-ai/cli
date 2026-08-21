package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/coordination"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

type rateParkingResumeContextKey struct{}
type authRepairContextKey struct{}

const rateParkingRearmInterruptedError = "rate-limit rearm attempt interrupted"

func isRateParkingResume(ctx context.Context) bool {
	_, resuming := rateParkingResumeRunID(ctx)
	return resuming
}

func rateParkingResumeRunID(ctx context.Context) (string, bool) {
	runID, ok := ctx.Value(rateParkingResumeContextKey{}).(string)
	return runID, ok && runID != ""
}

func isAuthRepair(ctx context.Context) bool {
	repair, _ := ctx.Value(authRepairContextKey{}).(bool)
	return repair
}

// parkRateLimitedRun is the application dispatch boundary: it accepts only a
// declaration-resolved opaque scope, a typed reset, and an already committed
// checkpoint. Every refusal returns before the parking store or run status is
// mutated.
func (a *App) parkRateLimitedRun(ctx context.Context, request etlModeDispatchRequest, result etlExecutionResult, runErr error) (Run, bool, error) {
	if a == nil || a.rateParking == nil {
		return Run{}, false, nil
	}
	resumedRunID, rearming := rateParkingResumeRunID(ctx)
	if rearming && request.runID == resumedRunID {
		return Run{}, true, errors.New("rate-limit resume attempt reused parked run ID")
	}
	if origin, tagged := synctransport.TransportExecutionOriginOf(runErr); tagged && origin != synctransport.TransportExecutionOriginSource {
		return Run{}, false, nil
	}
	var rateErr *connsdk.RateLimitError
	if !errors.As(runErr, &rateErr) || rateErr == nil || !rateErr.HasReset || rateErr.ResetAt.IsZero() {
		return Run{}, false, nil
	}
	resolver, ok := request.source.(connectors.RateLimitParkingScopeResolver)
	if !ok {
		return Run{}, false, nil
	}
	current, err := a.store.Load()
	if err != nil {
		return Run{}, true, err
	}
	stateKey := streamStateKey(request.connection.Name, request.streamName)
	streamState, found := current.StreamStates[stateKey]
	if !found || streamState.Checkpoint == nil || streamState.Checkpoint.CommittedAt == nil {
		return Run{}, false, nil
	}
	checkpoint := streamState.Checkpoint.Clone()
	if err := checkpoint.Validate(); err != nil {
		return Run{}, true, fmt.Errorf("park rate-limited run checkpoint: %w", err)
	}
	scope, err := resolver.RateLimitParkingScope(ctx, request.sourceRuntime, request.streamName, runErr)
	if err != nil {
		return Run{}, true, err
	}
	planID, err := declarativeTypedDestinationRateParkingPlanID(request)
	if err != nil {
		return Run{}, true, err
	}
	if planID != "" {
		if err := a.persistDeclarativeTypedDestinationRateParkingPlanID(request.runID, planID); err != nil {
			return Run{}, true, err
		}
	}
	if rearming {
		if _, err := engine.RearmRateLimitedRun(ctx, a.rateParking, resumedRunID, scope, checkpoint, runErr); err != nil {
			return Run{}, true, err
		}
	} else {
		if _, err := engine.ParkRateLimitedRun(ctx, a.rateParking, request.runID, scope, checkpoint, runErr); err != nil {
			return Run{}, true, err
		}
	}
	status := string(coordination.RateParkingOutcomeParkedRateLimit)
	runError := ""
	completedAt := time.Time{}
	if rearming {
		status = "failed"
		runError = coordination.ErrRateLimitRearmed.Error()
		completedAt = time.Now().UTC()
	}
	updated, err := a.updateState(func(current state) (state, error) {
		for index := range current.Runs {
			if current.Runs[index].ID != request.runID {
				continue
			}
			if rearming && current.Runs[index].Status != "running" {
				return current, fmt.Errorf("rate-limited resume attempt %q has status %q", request.runID, current.Runs[index].Status)
			}
			if !rearming && current.Runs[index].Status != "running" && current.Runs[index].Status != string(coordination.RateParkingOutcomeParkedRateLimit) {
				return current, fmt.Errorf("rate-limited run %q has terminal status %q", request.runID, current.Runs[index].Status)
			}
			if planID != "" && current.Runs[index].DeclarativeTypedDestinationPlanID != planID {
				return current, fmt.Errorf("rate-limited run %q declarative typed destination plan changed", request.runID)
			}
			current.Runs[index].Status = status
			current.Runs[index].RecordsRead = result.RecordsRead
			current.Runs[index].RecordsTransformed = result.RecordsTransformed
			current.Runs[index].RecordsLoaded = result.RecordsLoaded
			current.Runs[index].RecordsFailed = result.RecordsFailed
			current.Runs[index].BatchCount = result.BatchCount
			current.Runs[index].Checkpoint = cloneStringMap(result.Checkpoint)
			current.Runs[index].TransportPhaseMeasurement = cloneTransportPhaseMeasurement(result.TransportPhaseMeasurement)
			current.Runs[index].DestinationResults = cloneDestinationResults(result.DestinationResults)
			current.Runs[index].Error = runError
			current.Runs[index].CompletedAt = completedAt
			return current, nil
		}
		return current, fmt.Errorf("rate-limited run %q not found", request.runID)
	})
	if err != nil {
		return Run{}, true, err
	}
	for _, run := range updated.Runs {
		if run.ID == request.runID {
			if rearming {
				return run, true, coordination.ErrRateLimitRearmed
			}
			return run, true, coordination.ErrRateLimitParked
		}
	}
	return Run{}, true, fmt.Errorf("parked run %q was not stored", request.runID)
}

func declarativeTypedDestinationRateParkingPlanID(request etlModeDispatchRequest) (string, error) {
	descriptor, declared := connectors.DestinationTransportDescriptorOf(request.destination)
	if !declared || descriptor.Executor != declarativeTypedDestinationReference {
		return "", nil
	}
	if request.destinationApproval.PlanID == "" {
		return "", fmt.Errorf("declarative typed destination rate-limit parking requires an approval plan")
	}
	return request.destinationApproval.PlanID, nil
}

func (a *App) persistDeclarativeTypedDestinationRateParkingPlanID(runID, planID string) error {
	_, err := a.updateState(func(current state) (state, error) {
		for index := range current.Runs {
			if current.Runs[index].ID != runID {
				continue
			}
			if current.Runs[index].Status != "running" && current.Runs[index].Status != string(coordination.RateParkingOutcomeParkedRateLimit) {
				return current, fmt.Errorf("rate-limited run %q has terminal status %q", runID, current.Runs[index].Status)
			}
			if existing := current.Runs[index].DeclarativeTypedDestinationPlanID; existing != "" && existing != planID {
				return current, fmt.Errorf("rate-limited run %q declarative typed destination plan changed", runID)
			}
			current.Runs[index].DeclarativeTypedDestinationPlanID = planID
			return current, nil
		}
		return current, fmt.Errorf("rate-limited run %q not found", runID)
	})
	return err
}

func (a *App) resumeParkedRateLimitRun(ctx context.Context, parked coordination.ParkedRateLimitRun) error {
	loaded, err := a.store.Load()
	if err != nil {
		return err
	}
	if err := a.normalizeLoadedState(loaded, false); err != nil {
		return err
	}
	original, found := a.runByID(parked.RunID)
	if !found || original.Connection == "" || original.Stream == "" {
		return errors.New("parked run resume metadata is unavailable")
	}
	if original.Status == "resumed" {
		return nil
	}
	streamState, found := a.state.StreamStates[streamStateKey(original.Connection, original.Stream)]
	if !found || streamState.Checkpoint == nil || streamState.Checkpoint.CommittedAt == nil {
		return errors.New("parked run committed checkpoint is unavailable")
	}
	current := streamState.Checkpoint.Clone()
	if err := validateParkedCheckpointIdentity(current, parked.Checkpoint); err != nil {
		return err
	}
	linked, completed, err := a.reconcileRateParkingRearmAttempt(original)
	if err != nil {
		return err
	}
	if completed {
		return a.markParkedRateLimitRunResumed(parked.RunID, original.RateParkingRearmAttemptRunID)
	}
	if !transportCheckpointEqual(&current, &parked.Checkpoint) && !linked {
		return a.markParkedRateLimitRunResumed(parked.RunID, "")
	}
	attemptRunID, err := a.startParkedRateLimitRunRearm(ctx, original, current)
	if err != nil {
		return err
	}
	return a.markParkedRateLimitRunResumed(parked.RunID, attemptRunID)
}

func (a *App) startParkedRateLimitRunRearm(ctx context.Context, original Run, checkpoint synccontract.CheckpointEnvelope) (string, error) {
	attemptRunID, err := prefixedID("run")
	if err != nil {
		return "", err
	}
	if err := a.persistRateParkingRearmAttemptLink(original.ID, attemptRunID); err != nil {
		return "", err
	}
	request, err := parkedRateLimitRunETLRequest(original, checkpoint)
	if err != nil {
		return "", err
	}
	request.rateParkingRearmAttemptRunID = attemptRunID
	resumeCtx := context.WithValue(ctx, rateParkingResumeContextKey{}, original.ID)
	if _, err := a.RunETL(resumeCtx, request); err != nil {
		return "", classifyRateParkingResumeError(err)
	}
	return attemptRunID, nil
}

// classifyRateParkingResumeError keeps the coordinator connector-neutral while
// making a known-invalid, approval-bound authorization terminal for this exact
// parked scope. All other provider failures retain their retryable semantics.
func classifyRateParkingResumeError(err error) error {
	if err == nil {
		return nil
	}
	var expired *AuthorizationExpiredError
	if errors.As(err, &expired) {
		return coordination.NeedsReauthorization(err)
	}
	var revoked *AuthorizationRevokedError
	if errors.As(err, &revoked) {
		return coordination.NeedsReauthorization(err)
	}
	return err
}

func (a *App) persistRateParkingRearmAttemptLink(runID, attemptRunID string) error {
	_, err := a.updateState(func(current state) (state, error) {
		for index := range current.Runs {
			if current.Runs[index].ID != runID {
				continue
			}
			if current.Runs[index].Status != string(coordination.RateParkingOutcomeParkedRateLimit) && current.Runs[index].Status != "running" {
				return current, fmt.Errorf("parked run %q has status %q", runID, current.Runs[index].Status)
			}
			current.Runs[index].RateParkingRearmAttemptRunID = attemptRunID
			return current, nil
		}
		return current, fmt.Errorf("parked run %q not found", runID)
	})
	return err
}

func (a *App) reconcileRateParkingRearmAttempt(original Run) (bool, bool, error) {
	attemptRunID := original.RateParkingRearmAttemptRunID
	if attemptRunID == "" {
		return false, false, nil
	}
	attempt, found := a.runByID(attemptRunID)
	if !found {
		return true, false, nil
	}
	if attempt.ID == original.ID || attempt.Type != "etl" || attempt.Connection != original.Connection || attempt.Stream != original.Stream {
		return false, false, errors.New("rate-limit rearm attempt does not match the parked run")
	}
	switch attempt.Status {
	case "completed":
		return true, true, nil
	case "failed":
		return true, false, nil
	case "running":
		if err := a.failInterruptedRateParkingRearmAttempt(original.ID, attemptRunID); err != nil {
			return false, false, err
		}
		return true, false, nil
	default:
		return false, false, fmt.Errorf("rate-limit rearm attempt %q has status %q", attemptRunID, attempt.Status)
	}
}

func (a *App) failInterruptedRateParkingRearmAttempt(runID, attemptRunID string) error {
	_, err := a.updateState(func(current state) (state, error) {
		originalIndex := -1
		attemptIndex := -1
		for index := range current.Runs {
			switch current.Runs[index].ID {
			case runID:
				originalIndex = index
			case attemptRunID:
				attemptIndex = index
			}
		}
		if originalIndex < 0 || attemptIndex < 0 {
			return current, errors.New("rate-limit rearm attempt state is unavailable")
		}
		original := current.Runs[originalIndex]
		attempt := current.Runs[attemptIndex]
		if original.RateParkingRearmAttemptRunID != attemptRunID || attempt.ID == original.ID || attempt.Type != "etl" || attempt.Connection != original.Connection || attempt.Stream != original.Stream {
			return current, errors.New("rate-limit rearm attempt does not match the parked run")
		}
		if attempt.Status != "running" {
			return current, fmt.Errorf("rate-limit rearm attempt %q has status %q", attemptRunID, attempt.Status)
		}
		current.Runs[attemptIndex].Status = "failed"
		current.Runs[attemptIndex].Error = rateParkingRearmInterruptedError
		current.Runs[attemptIndex].CompletedAt = time.Now().UTC()
		return current, nil
	})
	return err
}

func (a *App) markParkedRateLimitRunResumed(runID, expectedAttemptRunID string) error {
	_, err := a.updateState(func(current state) (state, error) {
		for index := range current.Runs {
			if current.Runs[index].ID != runID {
				continue
			}
			if current.Runs[index].Status == "resumed" {
				return current, nil
			}
			if current.Runs[index].Status != string(coordination.RateParkingOutcomeParkedRateLimit) && current.Runs[index].Status != "running" {
				return current, fmt.Errorf("parked run %q has status %q", runID, current.Runs[index].Status)
			}
			if current.Runs[index].RateParkingRearmAttemptRunID != expectedAttemptRunID {
				return current, fmt.Errorf("parked run %q rearm attempt changed", runID)
			}
			current.Runs[index].Status = "resumed"
			current.Runs[index].CompletedAt = time.Now().UTC()
			current.Runs[index].RateParkingRearmAttemptRunID = ""
			return current, nil
		}
		return current, fmt.Errorf("parked run %q not found", runID)
	})
	return err
}

func parkedRateLimitRunETLRequest(original Run, checkpoint synccontract.CheckpointEnvelope) (RunETLRequest, error) {
	if original.BatchSize <= 0 {
		return RunETLRequest{}, errors.New("parked run effective batch size is unavailable")
	}
	resumeCheckpoint := checkpoint.Clone()
	request := RunETLRequest{Connection: original.Connection, Stream: original.Stream, BatchSize: original.BatchSize, rateParkingResumeCheckpoint: &resumeCheckpoint}
	if original.DeclarativeTypedDestinationPlanID != "" {
		request.DestinationApproval.PlanID = original.DeclarativeTypedDestinationPlanID
	}
	return request, nil
}

func (a *App) runByID(id string) (Run, bool) {
	for _, run := range a.state.Runs {
		if run.ID == id {
			return run, true
		}
	}
	return Run{}, false
}

func validateParkedCheckpointIdentity(current, parked synccontract.CheckpointEnvelope) error {
	if err := current.Validate(); err != nil {
		return fmt.Errorf("current parked-run checkpoint: %w", err)
	}
	if err := parked.Validate(); err != nil {
		return fmt.Errorf("durable parked-run checkpoint: %w", err)
	}
	if current.Source != parked.Source || current.Mechanism != parked.Mechanism ||
		current.SchemaVersion != parked.SchemaVersion || current.ProtocolVersion != parked.ProtocolVersion ||
		current.Dedupe.Kind != parked.Dedupe.Kind || !bytes.Equal(current.SourceGeneration, parked.SourceGeneration) {
		return errors.New("parked run checkpoint schema or source identity changed")
	}
	return nil
}
