package engine

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
	"polymetrics.ai/internal/synccontract"
)

func TestParkRateLimitedRun_PersistsOnlyTypedAuthoritativeEvidence(t *testing.T) {
	resetAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	checkpoint := engineParkingCheckpoint(resetAt.Add(-time.Second))
	parking := &recordingRateParkingCoordinator{}
	rateLimited := fmt.Errorf("read source: %w", &connsdk.RateLimitError{
		Source:   connsdk.RateLimitObservationSourceRetryAfter,
		ResetAt:  resetAt,
		HasReset: true,
	})

	if _, err := ParkRateLimitedRun(context.Background(), parking, "run-engine-3867", connectors.RateLimitScopeKey("scope-engine"), checkpoint, rateLimited); err != nil {
		t.Fatalf("ParkRateLimitedRun() error = %v", err)
	}
	if parking.calls != 1 {
		t.Fatalf("parking mutations = %d, want 1", parking.calls)
	}
	if parking.request.RunID != "run-engine-3867" || parking.request.Reason != connsdk.RateLimitObservationSourceRetryAfter || !parking.request.ResetAt.Equal(resetAt) {
		t.Fatalf("parking request = %#v, want actual retry_after reset evidence", parking.request)
	}
	if parking.request.Checkpoint.CommittedAt == nil || !parking.request.Checkpoint.CommittedAt.Equal(*checkpoint.CommittedAt) {
		t.Fatalf("parking checkpoint = %#v, want committed checkpoint %#v", parking.request.Checkpoint, checkpoint)
	}
}

func TestParkRateLimitedRun_RefusesMissingResetAndNonRateErrorsWithoutMutation(t *testing.T) {
	checkpoint := engineParkingCheckpoint(time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC))
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "rate limit without reset",
			err:  &connsdk.RateLimitError{Source: connsdk.RateLimitObservationSourceHTTP429, HasReset: false},
		},
		{
			name: "ordinary provider error",
			err:  errors.New("provider unavailable"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parking := &recordingRateParkingCoordinator{}
			_, err := ParkRateLimitedRun(context.Background(), parking, "run-refusal", connectors.RateLimitScopeKey("scope-refusal"), checkpoint, tt.err)
			if err == nil {
				t.Fatal("ParkRateLimitedRun() error = nil, want refusal")
			}
			if parking.calls != 0 {
				t.Fatalf("parking mutations = %d, want 0 on refusal", parking.calls)
			}
		})
	}
}

type recordingRateParkingCoordinator struct {
	calls   int
	request coordination.RateParkingRequest
}

func (c *recordingRateParkingCoordinator) Park(_ context.Context, request coordination.RateParkingRequest) (coordination.ParkedRateLimitRun, error) {
	c.calls++
	c.request = request
	return coordination.ParkedRateLimitRun{}, nil
}

func engineParkingCheckpoint(observedAt time.Time) synccontract.CheckpointEnvelope {
	committedAt := observedAt.Add(time.Second)
	return synccontract.CheckpointEnvelope{
		StateVersion: synccontract.StateVersion,
		Source: synccontract.SourceIdentity{
			Engine:           "engine-fixture",
			AccountOrCluster: "engine-account",
			ObjectScope:      "records",
		},
		Mechanism:       "engine-fixture",
		SnapshotBarrier: &synccontract.SnapshotBarrier{Kind: "fixture", Token: synccontract.OpaqueToken("barrier")},
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken("primary"),
			TieBreaker: synccontract.OpaqueToken("tie-breaker"),
		},
		Partitions:      []synccontract.PartitionState{},
		SchemaVersion:   "v1",
		ProtocolVersion: "v1",
		Dedupe:          synccontract.DedupeIdentity{Kind: "fixture", Value: synccontract.OpaqueToken("id")},
		DedupeWindow:    synccontract.DedupeWindow{Kind: "fixture", Start: synccontract.OpaqueToken("start"), End: synccontract.OpaqueToken("end")},
		ObservedAt:      observedAt,
		CommittedAt:     &committedAt,
	}
}
