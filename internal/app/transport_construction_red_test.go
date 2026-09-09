package app

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// These tests deliberately name the durable-stage contract that #4081 needs.
// They are committed RED before its production composition root and adapter.
func TestOpenInstallsIssueLabelWarehouseMediatedTransport(t *testing.T) {
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.ensureTransportRegistry(); err != nil {
		t.Fatal(err)
	}

	github, ok := a.registry.Get("github")
	if !ok {
		t.Fatal("GitHub connector is not registered")
	}
	if a.transportStage == nil {
		t.Fatal("Open() left the GitHub warehouse-mediated transport stage nil")
	}
	if _, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source:      github,
		Destination: github,
		Stream:      "issues",
		Mode:        synccontract.ModeFullAppend,
	}); err != nil {
		t.Fatalf("GitHub transport preflight = %v, want explicit accepted construction", err)
	}
}

func TestIssueLabelWarehouseStageReopensDurableReceiptAfterSourceReferencesAreDiscarded(t *testing.T) {
	ctx := context.Background()
	fixture := newIssueLabelWarehouseStageFixture(t)
	page := synctransport.SourcePage{Records: []connectors.Record{{
		"id":    "issue-4081-source",
		"title": "durable GitHub issue fixture",
	}}}
	receipt := stageIssueLabelWarehousePage(t, ctx, fixture, page)

	// No source-owned page or record remains available to the reopen call.
	page.Records[0]["title"] = "mutated source alias"
	page.Records = nil
	page = synctransport.SourcePage{}
	runtime.GC()

	first, err := fixture.app.transportStage.Reopen(ctx, receipt)
	if err != nil {
		t.Fatalf("Reopen() = %v", err)
	}
	if got := first.Records; len(got) != 1 || got[0]["title"] != "durable GitHub issue fixture" {
		t.Fatalf("first reopened records = %#v, want the durable source value", got)
	}
	first.Records[0]["title"] = "caller-mutated reopen alias"

	second, err := fixture.app.transportStage.Reopen(ctx, receipt)
	if err != nil {
		t.Fatalf("second Reopen() = %v", err)
	}
	if got := second.Records; len(got) != 1 || got[0]["title"] != "durable GitHub issue fixture" {
		t.Fatalf("second reopened records = %#v, want immutable durable source value", got)
	}
}

func TestIssueLabelWarehouseStageRejectsTamperedReceipt(t *testing.T) {
	ctx := context.Background()
	fixture := newIssueLabelWarehouseStageFixture(t)
	receipt := stageIssueLabelWarehousePage(t, ctx, fixture, synctransport.SourcePage{Records: []connectors.Record{{
		"id": "issue-4081-tamper",
	}}})

	for _, tt := range []struct {
		name   string
		mutate func(synctransport.WarehouseReceipt) synctransport.WarehouseReceipt
	}{
		{
			name: "owner",
			mutate: func(receipt synctransport.WarehouseReceipt) synctransport.WarehouseReceipt {
				receipt.Owner = "another-connection-owner"
				return receipt
			},
		},
		{
			name: "generation",
			mutate: func(receipt synctransport.WarehouseReceipt) synctransport.WarehouseReceipt {
				receipt.Generation++
				return receipt
			},
		},
		{
			name: "manifest hash",
			mutate: func(receipt synctransport.WarehouseReceipt) synctransport.WarehouseReceipt {
				receipt.ManifestSHA256 = "00"
				return receipt
			},
		},
		{
			name: "content hash",
			mutate: func(receipt synctransport.WarehouseReceipt) synctransport.WarehouseReceipt {
				receipt.ContentSHA256 = "00"
				return receipt
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			workset, err := fixture.app.transportStage.Reopen(ctx, tt.mutate(receipt))
			if err == nil {
				t.Fatalf("Reopen() workset = %#v, want tampered receipt rejection", workset)
			}
			if len(workset.Records) != 0 {
				t.Fatalf("Reopen() leaked records after %s tampering: %#v", tt.name, workset.Records)
			}
		})
	}
}

func TestIssueLabelWarehouseStageFreshOpenPreservesImmutableCheckpointAndTombstones(t *testing.T) {
	ctx := context.Background()
	fixture := newIssueLabelWarehouseStageFixture(t)
	checkpoint := issueLabelTransportDurabilityCheckpoint()
	tombstone := issueLabelTransportDurabilityTombstone()
	page := synctransport.SourcePage{
		Records: []connectors.Record{{
			"id":     "issue-4081-durable",
			"number": 100,
			"title":  "persisted before fresh open",
		}},
		Tombstones:          []synccontract.Tombstone{tombstone},
		CandidateCheckpoint: checkpoint,
	}
	receipt := stageIssueLabelWarehousePage(t, ctx, fixture, page)
	emptyReceipt := stageIssueLabelWarehousePage(t, ctx, fixture, synctransport.SourcePage{
		Records:             []connectors.Record{{"id": "issue-4081-empty", "number": 101}},
		Tombstones:          []synccontract.Tombstone{},
		CandidateCheckpoint: checkpoint,
	})

	// No source-owned records, tombstones, or checkpoint slices survive the
	// stage boundary. A fresh App must resolve solely from the artifact handle.
	page.Records[0]["title"] = "mutated source alias"
	page.Tombstones[0].EventID[0] = 'X'
	page.CandidateCheckpoint.Partitions = append(page.CandidateCheckpoint.Partitions, synccontract.PartitionState{
		Partition: synccontract.OpaqueToken("caller"),
		Position:  synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("1"), TieBreaker: synccontract.OpaqueToken("1")},
	})
	page = synctransport.SourcePage{}
	runtime.GC()

	fresh, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	first, err := fresh.transportStage.Reopen(ctx, receipt)
	if err != nil {
		t.Fatalf("fresh Reopen() = %v", err)
	}
	if len(first.Records) != 1 || first.Records[0]["title"] != "persisted before fresh open" {
		t.Fatalf("fresh reopened records = %#v, want persisted source record", first.Records)
	}
	if len(first.Tombstones) != 1 || string(first.Tombstones[0].EventID) != "delete-4081" {
		t.Fatalf("fresh reopened tombstones = %#v, want durable tombstone", first.Tombstones)
	}
	if first.CandidateCheckpoint.Partitions == nil || len(first.CandidateCheckpoint.Partitions) != 0 {
		t.Fatalf("fresh reopened checkpoint partitions = %#v, want explicit empty array", first.CandidateCheckpoint.Partitions)
	}

	// Mutating a reopened value cannot alter the parquet/manifest artifact that
	// the next independent App.Open resolves.
	first.Records[0]["title"] = "caller-mutated reopened record"
	first.Tombstones[0].EventID[0] = 'Y'
	first.Tombstones[0].Key[2] = 'X'
	first.CandidateCheckpoint.Partitions = append(first.CandidateCheckpoint.Partitions, synccontract.PartitionState{
		Partition: synccontract.OpaqueToken("caller"),
		Position:  synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("2"), TieBreaker: synccontract.OpaqueToken("2")},
	})

	freshAgain, err := Open(fixture.app.root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := freshAgain.transportStage.Reopen(ctx, receipt)
	if err != nil {
		t.Fatalf("second fresh Reopen() = %v", err)
	}
	if len(second.Records) != 1 || second.Records[0]["title"] != "persisted before fresh open" {
		t.Fatalf("second fresh reopened records = %#v, want immutable persisted source record", second.Records)
	}
	if len(second.Tombstones) != 1 || string(second.Tombstones[0].EventID) != "delete-4081" || string(second.Tombstones[0].Key) != `{"id":"issue-4081-durable"}` {
		t.Fatalf("second fresh reopened tombstones = %#v, want immutable durable tombstone", second.Tombstones)
	}
	if second.CandidateCheckpoint.Partitions == nil || len(second.CandidateCheckpoint.Partitions) != 0 {
		t.Fatalf("second fresh checkpoint partitions = %#v, want immutable explicit empty array", second.CandidateCheckpoint.Partitions)
	}
	empty, err := freshAgain.transportStage.Reopen(ctx, emptyReceipt)
	if err != nil {
		t.Fatalf("empty-slice Reopen() = %v", err)
	}
	if empty.Tombstones == nil || len(empty.Tombstones) != 0 {
		t.Fatalf("empty tombstones = %#v, want explicit non-nil empty array", empty.Tombstones)
	}
	if empty.CandidateCheckpoint.Partitions == nil || len(empty.CandidateCheckpoint.Partitions) != 0 {
		t.Fatalf("empty checkpoint partitions = %#v, want explicit non-nil empty array", empty.CandidateCheckpoint.Partitions)
	}
}

func TestIssueLabelWarehouseStageReopenRejectsActualArtifactCorruptionWithoutPayload(t *testing.T) {
	ctx := context.Background()
	for _, tt := range []struct {
		name string
		path func(connectionWarehouseStageArtifact) string
	}{
		{name: "manifest", path: func(artifact connectionWarehouseStageArtifact) string { return artifact.manifestPath }},
		{name: "wal", path: func(artifact connectionWarehouseStageArtifact) string { return artifact.walPath }},
		{name: "parquet", path: func(artifact connectionWarehouseStageArtifact) string { return artifact.parquetPath }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newIssueLabelWarehouseStageFixture(t)
			receipt := stageIssueLabelWarehousePage(t, ctx, fixture, synctransport.SourcePage{
				Records:             []connectors.Record{{"id": "issue-4081-" + tt.name, "number": 100}},
				CandidateCheckpoint: issueLabelTransportDurabilityCheckpoint(),
			})
			stage, ok := fixture.app.transportStage.(*connectionWarehouseStage)
			if !ok {
				t.Fatalf("transport stage = %T, want connectionWarehouseStage", fixture.app.transportStage)
			}
			conn, ok := fixture.app.findConnectionByID(fixture.connectionID)
			if !ok {
				t.Fatal("staged connection is unavailable")
			}
			artifact, err := stage.artifactFor(conn, receipt.ID)
			if err != nil {
				t.Fatal(err)
			}
			path := tt.path(artifact)
			bytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, append(bytes, '\n'), 0o600); err != nil {
				t.Fatal(err)
			}

			fresh, err := Open(fixture.app.root)
			if err != nil {
				t.Fatal(err)
			}
			workset, err := fresh.transportStage.Reopen(ctx, receipt)
			if err == nil {
				t.Fatalf("Reopen() workset = %#v, want %s corruption rejection", workset, tt.name)
			}
			if !reflect.DeepEqual(workset, synctransport.WarehouseWorkset{}) {
				t.Fatalf("Reopen() leaked payload after %s corruption: %#v", tt.name, workset)
			}
		})
	}
}

func issueLabelTransportDurabilityCheckpoint() synccontract.CheckpointEnvelope {
	positionObserved := true
	return synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           synccontract.SourceIdentity{Engine: "github", AccountOrCluster: "durability-fixture", ObjectScope: "issues"},
		Mechanism:        "github_durable_stage_fixture",
		SnapshotBarrier:  &synccontract.SnapshotBarrier{Kind: "issues-page", Token: synccontract.OpaqueToken("barrier-4081")},
		Position:         synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("100"), TieBreaker: synccontract.OpaqueToken("100")},
		PositionObserved: &positionObserved,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: synccontract.OpaqueToken("generation-4081"),
		SchemaVersion:    "github-issues-v1",
		ProtocolVersion:  "engine-read-v1",
		Dedupe:           synccontract.DedupeIdentity{Kind: "github_issue", Value: synccontract.OpaqueToken("issue-4081-durable")},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "github_issue_page", Start: synccontract.OpaqueToken("start"), End: synccontract.OpaqueToken("end")},
		ObservedAt:       time.Date(2026, time.August, 13, 0, 0, 0, 0, time.UTC),
	}
}

func issueLabelTransportDurabilityTombstone() synccontract.Tombstone {
	return synccontract.Tombstone{
		Operation:   synccontract.OperationDelete,
		EventID:     synccontract.OpaqueToken("delete-4081"),
		Key:         json.RawMessage(`{"id":"issue-4081-durable"}`),
		DeleteImage: synccontract.DeleteImageBefore,
		Before:      json.RawMessage(`{"id":"issue-4081-durable","title":"before"}`),
		Position:    synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("101"), TieBreaker: synccontract.OpaqueToken("101")},
	}
}

type issueLabelWarehouseStageFixture struct {
	app          *App
	connectionID string
}

func newIssueLabelWarehouseStageFixture(t *testing.T) issueLabelWarehouseStageFixture {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "github-source", Connector: "github"}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "github-destination", Connector: "github"}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "github-issues-label-demo",
		Source:      EndpointConfig{Connector: "github", Credential: "github-source"},
		Destination: EndpointConfig{Connector: "github", Credential: "github-destination"},
		Streams: map[string]StreamConfig{
			"issues": {SyncMode: string(synccontract.ModeFullAppend), DestinationTable: "issues"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if len(a.state.Connections) != 1 || a.state.Connections[0].ID == "" {
		t.Fatalf("created GitHub connection = %#v, want one connection with an opaque ID", a.state.Connections)
	}
	return issueLabelWarehouseStageFixture{app: a, connectionID: a.state.Connections[0].ID}
}

func stageIssueLabelWarehousePage(t *testing.T, ctx context.Context, fixture issueLabelWarehouseStageFixture, page synctransport.SourcePage) synctransport.WarehouseReceipt {
	t.Helper()
	receipt, err := fixture.app.transportStage.Stage(ctx, synctransport.WarehouseStageRequest{
		ConnectionID:    fixture.connectionID,
		Generation:      1,
		SourceName:      "github",
		DestinationName: "github",
		Stream:          "issues",
		Mode:            synccontract.ModeFullAppend,
		Page:            page,
	})
	if err != nil {
		t.Fatalf("Stage() = %v", err)
	}
	return receipt
}
