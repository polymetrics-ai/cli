package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	committedReceiptCommit, err := transportReceiptCommitFromWarehouseReceipt(committedReceipt)
	if err != nil {
		t.Fatalf("bind committed transport receipt: %v", err)
	}
	committedCheckpoint := checkpoint.Clone()
	committedCheckpoint.ObservedAt = committedCheckpoint.ObservedAt.Add(time.Second)
	committedAt := committedCheckpoint.ObservedAt.Add(time.Second)
	committedCheckpoint.CommittedAt = &committedAt
	if _, err := fixture.app.updateState(func(current state) (state, error) {
		if current.StreamStates == nil {
			current.StreamStates = make(map[string]StreamState)
		}
		current.StreamStates[streamStateKey(connection.Name, "issues")] = StreamState{
			Connection: connection.Name, Stream: "issues", Checkpoint: &committedCheckpoint, CommittedTransportReceipts: []TransportReceiptCommit{committedReceiptCommit}, GenerationID: 1, UpdatedAt: committedAt,
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
	// finalization; the next generic execution reconciles it before source I/O.
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

func TestDeferredReconciliationKeepsUncommittedRepeatedFullAppendReceipt(t *testing.T) {
	ctx := context.Background()
	fixture := newIssueLabelWarehouseStageFixture(t)
	stage, ok := fixture.app.transportStage.(*connectionWarehouseStage)
	if !ok {
		t.Fatalf("transport stage = %T, want connectionWarehouseStage", fixture.app.transportStage)
	}
	checkpoint := issueLabelTransportDurabilityCheckpoint()
	firstPage := synctransport.SourcePage{
		Records:             []connectors.Record{{"id": "repeated-stage"}},
		CandidateCheckpoint: checkpoint,
	}
	committedReceipt := stageIssueLabelWarehousePage(t, ctx, fixture, firstPage)
	connection := fixture.app.state.Connections[0]
	committedReceiptCommit, err := transportReceiptCommitFromWarehouseReceipt(committedReceipt)
	if err != nil {
		t.Fatalf("bind committed transport receipt: %v", err)
	}
	committedCheckpoint := checkpoint.Clone()
	committedAt := committedCheckpoint.ObservedAt.Add(time.Second)
	committedCheckpoint.CommittedAt = &committedAt
	if _, err := fixture.app.updateState(func(current state) (state, error) {
		if current.StreamStates == nil {
			current.StreamStates = make(map[string]StreamState)
		}
		current.StreamStates[streamStateKey(connection.Name, "issues")] = StreamState{
			Connection: connection.Name, Stream: "issues", Checkpoint: &committedCheckpoint, CommittedTransportReceipts: []TransportReceiptCommit{committedReceiptCommit}, GenerationID: 1, UpdatedAt: committedAt,
		}
		return current, nil
	}); err != nil {
		t.Fatalf("persist first committed workset: %v", err)
	}
	repeatedCheckpoint := checkpoint.Clone()
	repeatedCheckpoint.ObservedAt = repeatedCheckpoint.ObservedAt.Add(2 * time.Second)
	repeatedPage := firstPage
	repeatedPage.CandidateCheckpoint = repeatedCheckpoint
	uncommittedReceipt := stageIssueLabelWarehousePage(t, ctx, fixture, repeatedPage)
	if uncommittedReceipt.ID == committedReceipt.ID {
		t.Fatalf("repeated full-append receipt = %q, want distinct workset", uncommittedReceipt.ID)
	}
	committedArtifact, err := stage.artifactFor(connection, committedReceipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	uncommittedArtifact, err := stage.artifactFor(connection, uncommittedReceipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := Open(fixture.app.root)
	if err != nil {
		t.Fatalf("open after repeated full-append worksets: %v", err)
	}
	if err := fresh.reconcileCommittedTransportStages(ctx); err != nil {
		t.Fatalf("reconcile repeated full-append worksets: %v", err)
	}
	if _, err := os.Stat(committedArtifact.manifestPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("committed artifact stat error = %v, want removal", err)
	}
	if _, err := os.Stat(uncommittedArtifact.manifestPath); err != nil {
		t.Fatalf("uncommitted artifact stat error = %v, want retention", err)
	}
	freshStage, ok := fresh.transportStage.(*connectionWarehouseStage)
	if !ok {
		t.Fatalf("fresh transport stage = %T, want connectionWarehouseStage", fresh.transportStage)
	}
	recovered, err := freshStage.Stage(ctx, synctransport.WarehouseStageRequest{
		ConnectionID: fixture.connectionID, Generation: 1, SourceName: "github", DestinationName: "github",
		Stream: "issues", Mode: synccontract.ModeFullAppend, Page: repeatedPage,
	})
	if err != nil {
		t.Fatalf("recover repeated uncommitted workset: %v", err)
	}
	if recovered != uncommittedReceipt {
		t.Fatalf("recovered receipt = %#v, want %#v", recovered, uncommittedReceipt)
	}
}

func TestDeferredReconciliationMigratesLegacyReceiptBeforeRepeatedFullAppendRecovery(t *testing.T) {
	ctx := context.Background()
	fixture := newIssueLabelWarehouseStageFixture(t)
	stage, ok := fixture.app.transportStage.(*connectionWarehouseStage)
	if !ok {
		t.Fatalf("transport stage = %T, want connectionWarehouseStage", fixture.app.transportStage)
	}
	checkpoint := issueLabelTransportDurabilityCheckpoint()
	firstPage := synctransport.SourcePage{
		Records:             []connectors.Record{{"id": "legacy-repeated-stage"}},
		CandidateCheckpoint: checkpoint,
	}
	legacyReceipt := stageIssueLabelWarehousePage(t, ctx, fixture, firstPage)
	connection := fixture.app.state.Connections[0]
	committedCheckpoint := checkpoint.Clone()
	committedAt := committedCheckpoint.ObservedAt.Add(time.Second)
	committedCheckpoint.CommittedAt = &committedAt
	legacyState, err := fixture.app.store.Load()
	if err != nil {
		t.Fatalf("load pre-association state: %v", err)
	}
	if legacyState.StreamStates == nil {
		legacyState.StreamStates = make(map[string]StreamState)
	}
	legacyState.StreamStates[streamStateKey(connection.Name, "issues")] = StreamState{
		Connection: connection.Name, Stream: "issues", Checkpoint: &committedCheckpoint, GenerationID: 1, UpdatedAt: committedAt,
	}
	if err := fixture.app.store.Save(legacyState); err != nil {
		t.Fatalf("persist pre-association state: %v", err)
	}
	fresh, err := Open(fixture.app.root)
	if err != nil {
		t.Fatalf("open legacy state: %v", err)
	}
	streamState := fresh.state.StreamStates[streamStateKey(connection.Name, "issues")]
	if streamState.TransportReceiptAssociationVersion != transportReceiptAssociationVersion || streamState.LegacyCommittedTransportCheckpoint == nil {
		t.Fatalf("migrated legacy state = %#v, want exact legacy checkpoint marker", streamState)
	}
	freshStage, ok := fresh.transportStage.(*connectionWarehouseStage)
	if !ok {
		t.Fatalf("fresh transport stage = %T, want connectionWarehouseStage", fresh.transportStage)
	}
	repeatedCheckpoint := checkpoint.Clone()
	repeatedCheckpoint.ObservedAt = repeatedCheckpoint.ObservedAt.Add(time.Second)
	repeatedPage := firstPage
	repeatedPage.CandidateCheckpoint = repeatedCheckpoint
	repeatedReceipt, err := freshStage.Stage(ctx, synctransport.WarehouseStageRequest{
		ConnectionID: fixture.connectionID, Generation: 1, SourceName: "github", DestinationName: "github",
		Stream: "issues", Mode: synccontract.ModeFullAppend, Page: repeatedPage,
	})
	if err != nil {
		t.Fatalf("stage repeated legacy workset: %v", err)
	}
	if repeatedReceipt.ID == legacyReceipt.ID {
		t.Fatalf("repeated legacy receipt = %q, want a distinct workset", repeatedReceipt.ID)
	}
	if err := fresh.reconcileCommittedTransportStages(ctx); err != nil {
		t.Fatalf("reconcile migrated legacy receipt: %v", err)
	}
	legacyArtifact, err := stage.artifactFor(connection, legacyReceipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(legacyArtifact.manifestPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("legacy committed artifact stat error = %v, want removal", err)
	}
	repeatedArtifact, err := freshStage.artifactFor(connection, repeatedReceipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(repeatedArtifact.manifestPath); err != nil {
		t.Fatalf("repeated uncommitted artifact stat error = %v, want retention", err)
	}
	recovered, err := freshStage.Stage(ctx, synctransport.WarehouseStageRequest{
		ConnectionID: fixture.connectionID, Generation: 1, SourceName: "github", DestinationName: "github",
		Stream: "issues", Mode: synccontract.ModeFullAppend, Page: repeatedPage,
	})
	if err != nil {
		t.Fatalf("recover repeated legacy workset: %v", err)
	}
	if recovered != repeatedReceipt {
		t.Fatalf("recovered repeated receipt = %#v, want %#v", recovered, repeatedReceipt)
	}
	streamState = fresh.state.StreamStates[streamStateKey(connection.Name, "issues")]
	if streamState.LegacyCommittedTransportCheckpoint != nil || len(streamState.CommittedTransportReceipts) != 0 {
		t.Fatalf("reconciled legacy state = %#v, want consumed marker and no active receipt association", streamState)
	}
}

func TestDeferredReconciliationDefersReceiptWithLiveTransportWorkLease(t *testing.T) {
	ctx := context.Background()
	fixture := newIssueLabelWarehouseStageFixture(t)
	stage, ok := fixture.app.transportStage.(*connectionWarehouseStage)
	if !ok {
		t.Fatalf("transport stage = %T, want connectionWarehouseStage", fixture.app.transportStage)
	}
	checkpoint := issueLabelTransportDurabilityCheckpoint()
	receipt := stageIssueLabelWarehousePage(t, ctx, fixture, synctransport.SourcePage{
		Records:             []connectors.Record{{"id": "live-stage"}},
		CandidateCheckpoint: checkpoint,
	})
	connection := fixture.app.state.Connections[0]
	commit, err := transportReceiptCommitFromWarehouseReceipt(receipt)
	if err != nil {
		t.Fatalf("bind live receipt: %v", err)
	}
	committedCheckpoint := checkpoint.Clone()
	committedAt := committedCheckpoint.ObservedAt.Add(time.Second)
	committedCheckpoint.CommittedAt = &committedAt
	leaseUntil := time.Now().UTC().Add(time.Minute)
	workID := "running-acknowledged-work"
	if _, err := fixture.app.updateState(func(current state) (state, error) {
		if current.StreamStates == nil {
			current.StreamStates = make(map[string]StreamState)
		}
		current.Runs = append(current.Runs, Run{ID: workID, Status: "running"})
		current.StreamStates[streamStateKey(connection.Name, "issues")] = StreamState{
			Connection: connection.Name, Stream: "issues", Checkpoint: &committedCheckpoint,
			CommittedTransportReceipts: []TransportReceiptCommit{commit}, TransportReceiptAssociationVersion: transportReceiptAssociationVersion,
			GenerationID: 1, ActiveWorkID: workID, ActiveWorkFence: 1, ActiveWorkLeaseUntil: &leaseUntil, UpdatedAt: committedAt,
		}
		return current, nil
	}); err != nil {
		t.Fatalf("persist live acknowledged receipt: %v", err)
	}
	fresh, err := Open(fixture.app.root)
	if err != nil {
		t.Fatalf("open live acknowledged receipt: %v", err)
	}
	artifact, err := stage.artifactFor(connection, receipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := fresh.reconcileCommittedTransportStages(ctx); err != nil {
		t.Fatalf("reconcile live acknowledged receipt: %v", err)
	}
	if _, err := os.Stat(artifact.manifestPath); err != nil {
		t.Fatalf("live receipt artifact stat error = %v, want retention", err)
	}
	streamState := fresh.state.StreamStates[streamStateKey(connection.Name, "issues")]
	if len(streamState.CommittedTransportReceipts) != 1 || streamState.ActiveWorkID != workID {
		t.Fatalf("live reconciled stream state = %#v, want retained receipt and work lease", streamState)
	}
	if _, err := fresh.updateState(func(current state) (state, error) {
		for index := range current.Runs {
			if current.Runs[index].ID == workID {
				current.Runs[index].Status = "completed"
			}
		}
		streamState := current.StreamStates[streamStateKey(connection.Name, "issues")]
		streamState.ActiveWorkID = ""
		streamState.ActiveWorkLeaseUntil = nil
		current.StreamStates[streamStateKey(connection.Name, "issues")] = streamState
		return current, nil
	}); err != nil {
		t.Fatalf("complete live acknowledged work: %v", err)
	}
	if err := fresh.reconcileCommittedTransportStages(ctx); err != nil {
		t.Fatalf("reconcile terminal acknowledged receipt: %v", err)
	}
	if _, err := os.Stat(artifact.manifestPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("terminal receipt artifact stat error = %v, want removal", err)
	}
	streamState = fresh.state.StreamStates[streamStateKey(connection.Name, "issues")]
	if len(streamState.CommittedTransportReceipts) != 0 {
		t.Fatalf("terminal reconciled stream state = %#v, want cleared receipt association", streamState)
	}
}

func TestConnectionWarehouseStageRefusesUnreadableReceiptOnRetry(t *testing.T) {
	ctx := context.Background()
	fixture := newIssueLabelWarehouseStageFixture(t)
	stage, ok := fixture.app.transportStage.(*connectionWarehouseStage)
	if !ok {
		t.Fatalf("transport stage = %T, want connectionWarehouseStage", fixture.app.transportStage)
	}
	request := synctransport.WarehouseStageRequest{
		ConnectionID:    fixture.connectionID,
		Generation:      1,
		SourceName:      "github",
		DestinationName: "github",
		Stream:          "issues",
		Mode:            synccontract.ModeFullAppend,
		Page: synctransport.SourcePage{
			Records:             []connectors.Record{{"id": "unreadable-stage"}},
			CandidateCheckpoint: issueLabelTransportDurabilityCheckpoint(),
		},
	}
	receipt, err := stage.Stage(ctx, request)
	if err != nil {
		t.Fatalf("stage durable receipt: %v", err)
	}
	connection := fixture.app.state.Connections[0]
	artifact, err := stage.artifactFor(connection, receipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact.manifestPath, []byte("{"), 0o600); err != nil {
		t.Fatalf("corrupt staged receipt: %v", err)
	}
	if _, err := stage.Stage(ctx, request); err == nil || !strings.Contains(err.Error(), "read staged receipt") {
		t.Fatalf("retry Stage error = %v, want unreadable durable receipt refusal", err)
	}
	entries, err := os.ReadDir(filepath.Dir(artifact.manifestPath))
	if err != nil {
		t.Fatalf("read staged receipt directory: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(artifact.manifestPath) {
		t.Fatalf("staged receipt entries = %#v, want only original unreadable receipt", entries)
	}
}
