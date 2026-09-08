package engine

import (
	"context"
	"errors"
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
	"polymetrics.ai/internal/synccontract"
)

// RateLimitParkingCoordinator is the narrow parking consumer used by the
// engine. It is defined here because the engine owns typed provider-error
// classification; the coordination package remains provider-neutral.
type RateLimitParkingCoordinator interface {
	Park(context.Context, coordination.RateParkingRequest) (coordination.ParkedRateLimitRun, error)
}

type RateLimitParkingRearmCoordinator interface {
	Rearm(context.Context, coordination.RateParkingRequest) (coordination.ParkedRateLimitRun, error)
}

// ParkRateLimitedRun converts only a typed rate-limit error with authoritative
// parsed reset evidence into durable parking state. Generic failures and 429s
// without a reset time perform zero parking mutations.
func ParkRateLimitedRun(ctx context.Context, coordinator RateLimitParkingCoordinator, runID string, scope connectors.RateLimitScopeKey, checkpoint synccontract.CheckpointEnvelope, runErr error) (coordination.ParkedRateLimitRun, error) {
	if err := ctx.Err(); err != nil {
		return coordination.ParkedRateLimitRun{}, err
	}
	if coordinator == nil {
		return coordination.ParkedRateLimitRun{}, errors.New("rate parking coordinator is unavailable")
	}
	request, err := rateLimitParkingRequest(runID, scope, checkpoint, runErr)
	if err != nil {
		return coordination.ParkedRateLimitRun{}, err
	}
	parked, err := coordinator.Park(ctx, request)
	if err != nil {
		return coordination.ParkedRateLimitRun{}, fmt.Errorf("park rate-limited run: %w", err)
	}
	return parked, nil
}

func RearmRateLimitedRun(ctx context.Context, coordinator RateLimitParkingRearmCoordinator, runID string, scope connectors.RateLimitScopeKey, checkpoint synccontract.CheckpointEnvelope, runErr error) (coordination.ParkedRateLimitRun, error) {
	if err := ctx.Err(); err != nil {
		return coordination.ParkedRateLimitRun{}, err
	}
	if coordinator == nil {
		return coordination.ParkedRateLimitRun{}, errors.New("rate parking coordinator is unavailable")
	}
	request, err := rateLimitParkingRequest(runID, scope, checkpoint, runErr)
	if err != nil {
		return coordination.ParkedRateLimitRun{}, err
	}
	parked, err := coordinator.Rearm(ctx, request)
	if err != nil {
		return coordination.ParkedRateLimitRun{}, fmt.Errorf("rearm rate-limited run: %w", err)
	}
	return parked, nil
}

func rateLimitParkingRequest(runID string, scope connectors.RateLimitScopeKey, checkpoint synccontract.CheckpointEnvelope, runErr error) (coordination.RateParkingRequest, error) {
	var rateErr *connsdk.RateLimitError
	if !errors.As(runErr, &rateErr) || rateErr == nil || !rateErr.HasReset || rateErr.ResetAt.IsZero() {
		return coordination.RateParkingRequest{}, errors.New("rate-limit parking requires authoritative reset evidence")
	}
	if !rateParkingReasonValid(rateErr.Source) {
		return coordination.RateParkingRequest{}, errors.New("rate-limit parking requires a known rate-limit reason")
	}
	return coordination.RateParkingRequest{
		RunID:      runID,
		Scope:      scope,
		Checkpoint: checkpoint.Clone(),
		ResetAt:    rateErr.ResetAt.UTC(),
		Reason:     rateErr.Source,
	}, nil
}

func rateParkingReasonValid(reason connsdk.RateLimitObservationSource) bool {
	switch reason {
	case connsdk.RateLimitObservationSourceRetryAfter,
		connsdk.RateLimitObservationSourceHeaders,
		connsdk.RateLimitObservationSourceBody,
		connsdk.RateLimitObservationSourceHTTP429:
		return true
	default:
		return false
	}
}
