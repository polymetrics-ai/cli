package app

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestRunETLStopsWhenInitialRunCannotPersist(t *testing.T) {
	a := &App{state: state{
		Connections: []Connection{{
			Name:        "source_to_warehouse",
			Source:      EndpointConfig{Connector: "source", Credential: "source"},
			Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
			Streams: map[string]StreamConfig{
				"records": {SyncMode: "full_refresh_overwrite", DestinationTable: "records"},
			},
		}},
		Checkpoints: map[string]map[string]string{},
	}}
	run, err := a.RunETL(context.Background(), RunETLRequest{Connection: "source_to_warehouse", Stream: "records"})
	if err == nil {
		t.Fatal("RunETL error = nil, want initial persistence failure")
	}
	if run.ID != "" {
		t.Fatalf("RunETL returned run ID %q after failed initial persistence", run.ID)
	}
	if len(a.state.Runs) != 0 {
		t.Fatalf("in-memory runs = %+v, want no unpersisted run", a.state.Runs)
	}
}

func TestRunETLStopsOnInitialRunRevisionConflict(t *testing.T) {
	store := newStateStore(filepath.Join(t.TempDir(), "state.json"))
	if err := store.Save(state{Revision: 1}); err != nil {
		t.Fatalf("persist newer state: %v", err)
	}
	a := &App{
		store: store,
		state: state{
			Connections: []Connection{{
				Name:        "source_to_warehouse",
				Source:      EndpointConfig{Connector: "source", Credential: "source"},
				Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
				Streams: map[string]StreamConfig{
					"records": {SyncMode: "full_refresh_overwrite", DestinationTable: "records"},
				},
			}},
			Checkpoints: map[string]map[string]string{},
		},
	}

	run, err := a.RunETL(context.Background(), RunETLRequest{Connection: "source_to_warehouse", Stream: "records"})
	if !errors.Is(err, errStateRevisionConflict) {
		t.Fatalf("RunETL error = %v, want state revision conflict", err)
	}
	if run.ID != "" {
		t.Fatalf("RunETL returned run ID %q after initial revision conflict", run.ID)
	}
	if a.state.Revision != 1 || len(a.state.Runs) != 0 {
		t.Fatalf("in-memory state after revision conflict = %+v", a.state)
	}
}

func TestCompleteRunReturnsCompletedRunOnPersistenceFailure(t *testing.T) {
	a := &App{state: state{
		Runs:        []Run{{ID: "run_persistence_failure", Type: "etl", Status: "running"}},
		Checkpoints: map[string]map[string]string{},
	}}
	result := etlExecutionResult{
		RecordsRead:        2,
		RecordsTransformed: 2,
		RecordsLoaded:      2,
		BatchCount:         1,
		Checkpoint:         map[string]string{"cursor": "unbounded-cursor"},
		PendingStreamState: &pendingStreamState{Key: "source_to_warehouse/records", State: StreamState{}},
		RateLimit: connectors.RateLimitSummary{Connectors: []connectors.RateLimitConnectorSummary{{
			Connector:   "sample",
			Declaration: connectors.RateLimitDeclarationUndeclared,
		}}},
	}
	run, err := a.completeRun("run_persistence_failure", result)
	if !IsCompletedRunPersistenceError(err) {
		t.Fatalf("completeRun error = %v, want persistence error", err)
	}
	if run.ID != "run_persistence_failure" || run.Status != "completed" || run.RecordsRead != 2 || run.RecordsLoaded != 2 || len(run.RateLimit.Connectors) != 1 {
		t.Fatalf("completeRun returned run = %+v", run)
	}
}

func TestCompleteRunRejectsMissingRun(t *testing.T) {
	a := &App{state: state{Checkpoints: map[string]map[string]string{}}}
	run, err := a.completeRun("run_missing", etlExecutionResult{
		Checkpoint:         map[string]string{"cursor": "unbounded-cursor"},
		PendingStreamState: &pendingStreamState{Key: "source_to_warehouse/records", State: StreamState{}},
	})
	if !IsCompletedRunPersistenceError(err) {
		t.Fatalf("completeRun error = %v, want persistence error", err)
	}
	if run.ID != "" {
		t.Fatalf("completeRun returned run = %+v, want empty", run)
	}
	if len(a.state.Checkpoints) != 0 {
		t.Fatalf("checkpoints = %+v, want none", a.state.Checkpoints)
	}
}
