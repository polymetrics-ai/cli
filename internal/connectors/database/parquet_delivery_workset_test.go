package database_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

func TestDeriveChangeDeliveryWorksetImmutableIdentity(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})

	first, err := database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if err != nil {
		t.Fatalf("first DeriveChangeDeliveryWorkset() error = %v", err)
	}
	second, err := database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if err != nil {
		t.Fatalf("second DeriveChangeDeliveryWorkset() error = %v", err)
	}
	if got, want := []byte(first.Identity()), []byte(second.Identity()); !bytes.Equal(got, want) {
		t.Fatalf("identical derivations have distinct identity bytes: first=%x second=%x", got, want)
	}
	if first.ContentSHA256() != second.ContentSHA256() {
		t.Fatalf("identical derivations have distinct content hash: first=%q second=%q", first.ContentSHA256(), second.ContentSHA256())
	}

	changedDestination := request
	changedDestination.Control = testChangeDeliveryWorksetControl(t, "destination-2", 1)
	destinationWorkset, err := database.DeriveChangeDeliveryWorkset(context.Background(), changedDestination)
	if err != nil {
		t.Fatalf("destination-mutated DeriveChangeDeliveryWorkset() error = %v", err)
	}
	if destinationWorkset.Identity() == first.Identity() {
		t.Fatal("changing the asserted destination reused the prior workset identity")
	}

	changedSchema := request
	changedSchema.Control = testChangeDeliveryWorksetControl(t, "destination-1", 2)
	schemaWorkset, err := database.DeriveChangeDeliveryWorkset(context.Background(), changedSchema)
	if err != nil {
		t.Fatalf("schema-mutated DeriveChangeDeliveryWorkset() error = %v", err)
	}
	if schemaWorkset.Identity() == first.Identity() {
		t.Fatal("changing the asserted schema reused the prior workset identity")
	}

	changedKeys := request
	changedKeys.Keys = []string{"id", "tenant_id"}
	keyWorkset, err := database.DeriveChangeDeliveryWorkset(context.Background(), changedKeys)
	if err != nil {
		t.Fatalf("key-mutated DeriveChangeDeliveryWorkset() error = %v", err)
	}
	if keyWorkset.Identity() == first.Identity() {
		t.Fatal("changing the ordered key binding reused the prior workset identity")
	}

	renamedProvenance := request
	renamedProvenance.Control = testChangeDeliveryWorksetControlForArtifact(t, "destination-1", 1, "orders-renamed")
	renamedWorkset, err := database.DeriveChangeDeliveryWorkset(context.Background(), renamedProvenance)
	if err != nil {
		t.Fatalf("renamed-provenance DeriveChangeDeliveryWorkset() error = %v", err)
	}
	if got, want := renamedWorkset.Identity(), first.Identity(); got != want {
		t.Fatalf("renaming mutable source artifact provenance changed workset identity: got %q want %q", got, want)
	}

	before := readWorksetProjectionRows(t, first)
	beforeHash := first.ContentSHA256()
	if err := warehouse.WriteTable(context.Background(), request.SourceParquet, []warehouse.Row{{
		"tenant_id": "north",
		"id":        1,
		"value":     "replacement",
	}}); err != nil {
		t.Fatalf("replace source parquet after derivation: %v", err)
	}
	after := readWorksetProjectionRows(t, first)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("existing workset projection changed after source replacement: got %#v want %#v", after, before)
	}
	if got := first.ContentSHA256(); got != beforeHash {
		t.Fatalf("existing workset content hash changed after source replacement: got %q want %q", got, beforeHash)
	}
}

func TestDeriveChangeDeliveryWorksetDerivesRealParquetDeltaAndExplicitTombstones(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	baselineBefore, err := os.ReadFile(request.BaselineParquet)
	if err != nil {
		t.Fatal(err)
	}
	request.Tombstones = []synccontract.Tombstone{testChangeDeliveryTombstone()}

	workset, err := database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if err != nil {
		t.Fatalf("DeriveChangeDeliveryWorkset() error = %v", err)
	}
	if got, want := workset.Records(), int64(4); got != want {
		t.Fatalf("projection records = %d, want %d", got, want)
	}
	if got, want := workset.Changes(), int64(3); got != want {
		t.Fatalf("delta records = %d, want %d", got, want)
	}
	if got, want := workset.TombstoneCount(), int64(1); got != want {
		t.Fatalf("tombstone records = %d, want %d", got, want)
	}

	delta := readWorksetDeltaRows(t, workset)
	for _, key := range []struct {
		tenant string
		id     string
	}{
		{tenant: "north", id: "1"}, // updated
		{tenant: "south", id: "1"}, // inserted
		{tenant: "east", id: "1"},  // string "7" differs from numeric 7
	} {
		if !hasWorksetRow(delta, key.tenant, key.id) {
			t.Fatalf("delta rows %#v omit expected %s/%s change", delta, key.tenant, key.id)
		}
	}
	if hasWorksetRow(delta, "north", "2") {
		t.Fatalf("unchanged row north/2 was emitted in delta: %#v", delta)
	}

	tombstones, err := workset.Tombstones(context.Background())
	if err != nil {
		t.Fatalf("Tombstones() error = %v", err)
	}
	if !reflect.DeepEqual(tombstones, request.Tombstones) {
		t.Fatalf("sealed tombstones = %#v, want only explicit input %#v", tombstones, request.Tombstones)
	}
	for _, tombstone := range tombstones {
		if bytes.Contains(tombstone.Key, []byte(`"gone"`)) {
			t.Fatalf("physical baseline absence became a tombstone: %#v", tombstones)
		}
	}

	baselineAfter, err := os.ReadFile(request.BaselineParquet)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(baselineAfter, baselineBefore) {
		t.Fatal("derivation mutated the supplied baseline instead of producing a separate candidate")
	}
	if got, want := len(readWorksetCandidateBaselineRows(t, workset)), 4; got != want {
		t.Fatalf("candidate baseline rows = %d, want source projection count %d", got, want)
	}
}

func TestDeriveChangeDeliveryWorksetRefusesWithoutPublishing(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*testing.T, *database.ChangeDeliveryWorksetRequest)
	}{
		{
			name: "duplicate composite source key",
			mutate: func(t *testing.T, request *database.ChangeDeliveryWorksetRequest) {
				t.Helper()
				if err := warehouse.WriteTable(context.Background(), request.SourceParquet, []warehouse.Row{
					{"tenant_id": "north", "id": 1, "value": "first"},
					{"tenant_id": "north", "id": 1, "value": "second"},
				}); err != nil {
					t.Fatalf("write duplicate-key source fixture: %v", err)
				}
			},
		},
		{
			name: "null composite source key",
			mutate: func(t *testing.T, request *database.ChangeDeliveryWorksetRequest) {
				t.Helper()
				if err := warehouse.WriteTable(context.Background(), request.SourceParquet, []warehouse.Row{
					{"tenant_id": nil, "id": 1, "value": "missing-key"},
				}); err != nil {
					t.Fatalf("write null-key source fixture: %v", err)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
			baselineBefore, err := os.ReadFile(request.BaselineParquet)
			if err != nil {
				t.Fatal(err)
			}
			test.mutate(t, &request)
			_, err = database.DeriveChangeDeliveryWorkset(context.Background(), request)
			if !errors.Is(err, database.ErrChangeDeliveryWorksetInvalid) {
				t.Fatalf("DeriveChangeDeliveryWorkset() error = %v, want invalid-key refusal", err)
			}
			assertNoPublishedWorkset(t, request.Root)
			baselineAfter, err := os.ReadFile(request.BaselineParquet)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(baselineAfter, baselineBefore) {
				t.Fatal("invalid derivation mutated the supplied baseline")
			}
		})
	}
}

func TestDeriveChangeDeliveryWorksetCancellationDoesNotCreateArtifacts(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := database.DeriveChangeDeliveryWorkset(ctx, request)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("DeriveChangeDeliveryWorkset() error = %v, want canceled context", err)
	}
	if _, err := os.Stat(request.Root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("canceled derivation created artifact root %q: stat error = %v", request.Root, err)
	}
}

func TestDeriveChangeDeliveryWorksetCancellationCleansStagingArtifacts(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	ctx := &cancelOnErrCallContext{cancelAt: 3, done: make(chan struct{})}
	_, err := database.DeriveChangeDeliveryWorkset(ctx, request)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("DeriveChangeDeliveryWorkset() error = %v, want cancellation during staged copy", err)
	}
	entries, err := os.ReadDir(request.Root)
	if err != nil {
		t.Fatalf("read workset root after cancellation: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("canceled derivation left staged artifacts in %q: %#v", request.Root, entries)
	}
}

func TestDeriveChangeDeliveryWorksetRejectsArtifactBeyondDeclaredBound(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	baselineBefore, err := os.ReadFile(request.BaselineParquet)
	if err != nil {
		t.Fatal(err)
	}
	request.MaxArtifactBytes = 1
	_, err = database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if !errors.Is(err, database.ErrChangeDeliveryWorksetInvalid) {
		t.Fatalf("DeriveChangeDeliveryWorkset() error = %v, want bounded-artifact refusal", err)
	}
	if _, err := os.Stat(request.Root); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("over-bound derivation created artifact root %q: stat error = %v", request.Root, err)
	}
	baselineAfter, err := os.ReadFile(request.BaselineParquet)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(baselineAfter, baselineBefore) {
		t.Fatal("over-bound derivation mutated the supplied baseline")
	}
}

func TestDeriveChangeDeliveryWorksetRejectsCorruptReuseWithoutReplacingIt(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	baselineBefore, err := os.ReadFile(request.BaselineParquet)
	if err != nil {
		t.Fatal(err)
	}
	workset, err := database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if err != nil {
		t.Fatalf("initial DeriveChangeDeliveryWorkset() error = %v", err)
	}
	manifestPath := filepath.Join(request.Root, workset.Identity(), "manifest.json")
	corrupt := []byte("not a workset manifest")
	if err := os.WriteFile(manifestPath, corrupt, 0o600); err != nil {
		t.Fatalf("corrupt manifest fixture: %v", err)
	}

	_, err = database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if !errors.Is(err, database.ErrChangeDeliveryWorksetUnavailable) {
		t.Fatalf("DeriveChangeDeliveryWorkset() error = %v, want corrupt-artifact refusal", err)
	}
	after, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, corrupt) {
		t.Fatal("corrupt immutable artifact was replaced during reuse attempt")
	}
	baselineAfter, err := os.ReadFile(request.BaselineParquet)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(baselineAfter, baselineBefore) {
		t.Fatal("corrupt-reuse refusal mutated the supplied baseline")
	}
}

func TestDeriveChangeDeliveryWorksetAcceptsWarehouseEmptyParquet(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	if err := warehouse.WriteTable(context.Background(), request.SourceParquet, nil); err != nil {
		t.Fatalf("write empty source parquet: %v", err)
	}
	if err := warehouse.WriteTable(context.Background(), request.BaselineParquet, nil); err != nil {
		t.Fatalf("write empty baseline parquet: %v", err)
	}
	baselineBefore, err := os.ReadFile(request.BaselineParquet)
	if err != nil {
		t.Fatal(err)
	}

	workset, err := database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if err != nil {
		t.Fatalf("DeriveChangeDeliveryWorkset() error = %v", err)
	}
	if workset.Records() != 0 || workset.Changes() != 0 || workset.TombstoneCount() != 0 {
		t.Fatalf("empty workset counts = records:%d changes:%d tombstones:%d, want all zero", workset.Records(), workset.Changes(), workset.TombstoneCount())
	}
	if got := len(readWorksetProjectionRows(t, workset)); got != 0 {
		t.Fatalf("empty projection rows = %d, want 0", got)
	}
	if got := len(readWorksetDeltaRows(t, workset)); got != 0 {
		t.Fatalf("empty delta rows = %d, want 0", got)
	}
	if got := len(readWorksetCandidateBaselineRows(t, workset)); got != 0 {
		t.Fatalf("empty candidate baseline rows = %d, want 0", got)
	}
	baselineAfter, err := os.ReadFile(request.BaselineParquet)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(baselineAfter, baselineBefore) {
		t.Fatal("empty derivation mutated the supplied empty baseline")
	}
}

func testChangeDeliveryWorksetRequest(t *testing.T, destination string, schemaVersion uint, keys []string) database.ChangeDeliveryWorksetRequest {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "source.parquet")
	baseline := filepath.Join(root, "baseline.parquet")
	if err := warehouse.WriteTable(context.Background(), source, []warehouse.Row{
		{"tenant_id": "north", "id": 1, "value": "updated", "nullable": nil},
		{"tenant_id": "north", "id": 2, "value": "same", "nullable": nil},
		{"tenant_id": "south", "id": 1, "value": "new", "nullable": true},
		{"tenant_id": "east", "id": 1, "value": "7", "nullable": nil},
	}); err != nil {
		t.Fatalf("write source parquet: %v", err)
	}
	if err := warehouse.WriteTable(context.Background(), baseline, []warehouse.Row{
		{"tenant_id": "north", "id": 1, "value": "old", "nullable": nil},
		{"tenant_id": "north", "id": 2, "value": "same", "nullable": nil},
		{"tenant_id": "gone", "id": 9, "value": "physical absence", "nullable": false},
		{"tenant_id": "east", "id": 1, "value": 7, "nullable": nil},
	}); err != nil {
		t.Fatalf("write baseline parquet: %v", err)
	}
	return database.ChangeDeliveryWorksetRequest{
		Control:          testChangeDeliveryWorksetControl(t, destination, schemaVersion),
		Keys:             append([]string(nil), keys...),
		SourceParquet:    source,
		BaselineParquet:  baseline,
		Root:             filepath.Join(root, "worksets"),
		MaxArtifactBytes: 1 << 20,
	}
}

func testChangeDeliveryWorksetControl(t *testing.T, destination string, schemaVersion uint) database.ManagedTargetControlRecord {
	t.Helper()
	return testChangeDeliveryWorksetControlForArtifact(t, destination, schemaVersion, "orders")
}

func testChangeDeliveryWorksetControlForArtifact(t *testing.T, destination string, schemaVersion uint, artifactTable string) database.ManagedTargetControlRecord {
	t.Helper()
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "postgres",
		ConnectionID: "source-connection-1",
	}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := warehouse.NewArtifactRef(identity, artifactTable)
	if err != nil {
		t.Fatal(err)
	}
	target, err := database.NewManagedTargetRef(owner, artifact, "stream-orders")
	if err != nil {
		t.Fatal(err)
	}
	return testManagedTargetControl(
		t,
		owner,
		target,
		testTargetDatabase(t, destination),
		database.NativeRelationIdentity{Kind: "fixture-native-id", Value: "relation-orders"},
		testManagedTargetSchema(t, schemaVersion, byte(schemaVersion)),
	)
}

func testChangeDeliveryTombstone() synccontract.Tombstone {
	return synccontract.Tombstone{
		Operation:   synccontract.OperationDelete,
		EventID:     synccontract.OpaqueToken("event-1"),
		Key:         json.RawMessage(`{"tenant_id":"removed","id":4}`),
		DeleteImage: synccontract.DeleteImageKeyOnly,
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken("00000011"),
			TieBreaker: synccontract.OpaqueToken("00000001"),
		},
	}
}

func readWorksetProjectionRows(t *testing.T, workset database.ChangeDeliveryWorkset) []warehouse.Row {
	t.Helper()
	var rows []warehouse.Row
	if err := workset.ReadProjection(context.Background(), func(row warehouse.Row) error {
		rows = append(rows, row)
		return nil
	}); err != nil {
		t.Fatalf("read workset projection: %v", err)
	}
	return rows
}

func readWorksetDeltaRows(t *testing.T, workset database.ChangeDeliveryWorkset) []warehouse.Row {
	t.Helper()
	var rows []warehouse.Row
	if err := workset.ReadDelta(context.Background(), func(row warehouse.Row) error {
		rows = append(rows, row)
		return nil
	}); err != nil {
		t.Fatalf("read workset delta: %v", err)
	}
	return rows
}

func readWorksetCandidateBaselineRows(t *testing.T, workset database.ChangeDeliveryWorkset) []warehouse.Row {
	t.Helper()
	var rows []warehouse.Row
	if err := workset.ReadCandidateBaseline(context.Background(), func(row warehouse.Row) error {
		rows = append(rows, row)
		return nil
	}); err != nil {
		t.Fatalf("read candidate baseline: %v", err)
	}
	return rows
}

func hasWorksetRow(rows []warehouse.Row, tenant, id string) bool {
	for _, row := range rows {
		if row["tenant_id"] == tenant && fmt.Sprint(row["id"]) == id {
			return true
		}
	}
	return false
}

func assertNoPublishedWorkset(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read workset root %q: %v", root, err)
	}
	if len(entries) != 0 {
		t.Fatalf("invalid derivation published artifact(s) in %q: %#v", root, entries)
	}
}
