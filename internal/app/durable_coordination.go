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

func isRateParkingResume(ctx context.Context) bool {
	resuming, _ := ctx.Value(rateParkingResumeContextKey{}).(bool)
	return resuming
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
	if a == nil || a.rateParking == nil || isRateParkingResume(ctx) {
		return Run{}, false, nil
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
	if _, err := engine.ParkRateLimitedRun(ctx, a.rateParking, request.runID, scope, checkpoint, runErr); err != nil {
		return Run{}, true, err
	}
	updated, err := a.updateState(func(current state) (state, error) {
		for index := range current.Runs {
			if current.Runs[index].ID != request.runID {
				continue
			}
			if current.Runs[index].Status != "running" && current.Runs[index].Status != string(coordination.RateParkingOutcomeParkedRateLimit) {
				return current, fmt.Errorf("rate-limited run %q has terminal status %q", request.runID, current.Runs[index].Status)
			}
			current.Runs[index].Status = string(coordination.RateParkingOutcomeParkedRateLimit)
			current.Runs[index].RecordsRead = result.RecordsRead
			current.Runs[index].RecordsTransformed = result.RecordsTransformed
			current.Runs[index].RecordsLoaded = result.RecordsLoaded
			current.Runs[index].RecordsFailed = result.RecordsFailed
			current.Runs[index].BatchCount = result.BatchCount
			current.Runs[index].Checkpoint = cloneStringMap(result.Checkpoint)
			current.Runs[index].TransportPhaseMeasurement = cloneTransportPhaseMeasurement(result.TransportPhaseMeasurement)
			current.Runs[index].DestinationResults = cloneDestinationResults(result.DestinationResults)
			current.Runs[index].Error = ""
			current.Runs[index].CompletedAt = time.Time{}
			return current, nil
		}
		return current, fmt.Errorf("rate-limited run %q not found", request.runID)
	})
	if err != nil {
		return Run{}, true, err
	}
	for _, run := range updated.Runs {
		if run.ID == request.runID {
			return run, true, coordination.ErrRateLimitParked
		}
	}
	return Run{}, true, fmt.Errorf("parked run %q was not stored", request.runID)
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
	streamState, found := a.state.StreamStates[streamStateKey(original.Connection, original.Stream)]
	if !found || streamState.Checkpoint == nil || streamState.Checkpoint.CommittedAt == nil {
		return errors.New("parked run committed checkpoint is unavailable")
	}
	current := streamState.Checkpoint.Clone()
	if err := validateParkedCheckpointIdentity(current, parked.Checkpoint); err != nil {
		return err
	}
	if transportCheckpointEqual(&current, &parked.Checkpoint) {
		resumeCtx := context.WithValue(ctx, rateParkingResumeContextKey{}, true)
		if _, err := a.RunETL(resumeCtx, RunETLRequest{Connection: original.Connection, Stream: original.Stream}); err != nil {
			return err
		}
	}
	// If the checkpoint already advanced, the resumed run committed before a
	// crash between acknowledgement and parking completion. Marking the
	// original record resumed without executing again prevents replay.
	updated, err := a.updateState(func(current state) (state, error) {
		for index := range current.Runs {
			if current.Runs[index].ID != parked.RunID {
				continue
			}
			if current.Runs[index].Status == "resumed" {
				return current, nil
			}
			// The durable parking record is published before this status. A
			// process killed between those two commits can legitimately recover
			// with the original run still marked running.
			if current.Runs[index].Status != string(coordination.RateParkingOutcomeParkedRateLimit) && current.Runs[index].Status != "running" {
				return current, fmt.Errorf("parked run %q has status %q", parked.RunID, current.Runs[index].Status)
			}
			current.Runs[index].Status = "resumed"
			current.Runs[index].CompletedAt = time.Now().UTC()
			return current, nil
		}
		return current, fmt.Errorf("parked run %q not found", parked.RunID)
	})
	if err == nil {
		a.state = updated
	}
	return err
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
