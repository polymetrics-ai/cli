package app

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestDeferredReconciliationRetiresOnlyCommittedConnectionOwnedTransportStages(t *testing.T) {
	ctx := context.Background()
	fixture := newIssueLabelWarehouseStageFixture(t)
	stage, ok := fixture.app.transportStage.(*connectionWarehouseStage)
	if !ok {
		t.Fatalf("transport stage = %T, want connectionWarehouseStage", fixture.app.transportStage)
	}
	checkpoint := issueLabelTransportDurabilityCheckpoint()
	committedReceipt := stageIssueLabelWarehousePage(t, ctx, fixture, synctransport.SourcePage{
		Records:             []connectors.Record{{"id": "committed-stage"}},
		CandidateCheckpoint: checkpoint,
	})
	activeCheckpoint := checkpoint.Clone()
	activeCheckpoint.Position.Primary = synccontract.OpaqueToken("next")
	activeCheckpoint.Position.TieBreaker = synccontract.OpaqueToken("next")
	activeReceipt := stageIssueLabelWarehousePage(t, ctx, fixture, synctransport.SourcePage{
		Records:             []connectors.Record{{"id": "active-stage"}},
		CandidateCheckpoint: activeCheckpoint,
	})
	connection := fixture.app.state.Connections[0]
	committedCheckpoint := checkpoint.Clone()
	committedAt := time.Now().UTC()
	committedCheckpoint.CommittedAt = &committedAt
	if _, err := fixture.app.updateState(func(current state) (state, error) {
		if current.StreamStates == nil {
			current.StreamStates = make(map[string]StreamState)
		}
		current.StreamStates[streamStateKey(connection.Name, "issues")] = StreamState{
			Connection: connection.Name, Stream: "issues", Checkpoint: &committedCheckpoint, GenerationID: 1, UpdatedAt: committedAt,
		}
		return current, nil
	}); err != nil {
		t.Fatalf("persist committed stage checkpoint: %v", err)
	}
	committedArtifact, err := stage.artifactFor(connection, committedReceipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	activeArtifact, err := stage.artifactFor(connection, activeReceipt.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Opening a fresh process models a kill immediately after the checkpoint
	// rename: the checkpoint is durable, while its transient receipt was never
	// retired by the killed process. Open retains the evidence for recovery and
	// certification; the next generic execution reconciles it before source I/O.
	fresh, err := Open(fixture.app.root)
	if err != nil {
		t.Fatalf("open after committed-stage interruption: %v", err)
	}
	if fresh == nil {
		t.Fatal("Open() returned a nil app")
	}
	for _, path := range []string{committedArtifact.manifestPath, committedArtifact.walPath, committedArtifact.parquetPath, activeArtifact.manifestPath, activeArtifact.walPath, activeArtifact.parquetPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("fresh Open removed transient artifact %q: %v", path, err)
		}
	}
	if err := fresh.reconcileCommittedTransportStages(ctx); err != nil {
		t.Fatalf("reconcile committed transport stages: %v", err)
	}
	for _, path := range []string{committedArtifact.manifestPath, committedArtifact.walPath, committedArtifact.parquetPath} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("committed transient artifact %q stat error = %v, want removal", path, err)
		}
	}
	for _, path := range []string{activeArtifact.manifestPath, activeArtifact.walPath, activeArtifact.parquetPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("active transient artifact %q stat error = %v, want retention", path, err)
		}
	}
}
