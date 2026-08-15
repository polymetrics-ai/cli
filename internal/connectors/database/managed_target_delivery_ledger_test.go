package database_test

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/warehouse"
)

func TestManagedTargetDeliveryLedgerRenameAndRestart(t *testing.T) {
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "source-database",
		ConnectionID: "source-connection-1",
	}
	artifact, err := warehouse.NewArtifactRef(identity, "orders")
	if err != nil {
		t.Fatal(err)
	}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal(err)
	}
	targetDatabase := testTargetDatabase(t, "database-1")
	ref, err := database.NewManagedTargetRef(owner, artifact, "stream-orders")
	if err != nil {
		t.Fatal(err)
	}
	control := testManagedTargetControl(
		t,
		owner,
		ref,
		targetDatabase,
		database.NativeRelationIdentity{Kind: "fixture-native-id", Value: "relation-orders"},
		testManagedTargetSchema(t, 1, 1),
	)
	key, err := database.NewManagedTargetDeliveryLedgerKey(control)
	if err != nil {
		t.Fatal(err)
	}
	record, err := database.NewManagedTargetDeliveryRecord("delivery-orders-1")
	if err != nil {
		t.Fatal(err)
	}
	store := newManagedTargetDeliveryLedgerStoreFake()
	ledger, err := database.NewManagedTargetDeliveryLedger(store)
	if err != nil {
		t.Fatal(err)
	}
	if err := ledger.Record(context.Background(), key, record); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if got := store.writeCount(); got != 1 {
		t.Fatalf("durable store writes = %d, want 1", got)
	}

	// The source artifact table is mutable provenance only. Reopening after a
	// rename must use the same immutable StreamID relation and find the record
	// written before restart.
	renamedArtifact, err := warehouse.NewArtifactRef(identity, "orders-renamed")
	if err != nil {
		t.Fatal(err)
	}
	renamedRef, err := database.NewManagedTargetRef(owner, renamedArtifact, ref.StreamID())
	if err != nil {
		t.Fatal(err)
	}
	renamedControl := testManagedTargetControl(
		t,
		owner,
		renamedRef,
		targetDatabase,
		control.NativeIdentity(),
		control.Schema(),
	)
	renamedKey, err := database.NewManagedTargetDeliveryLedgerKey(renamedControl)
	if err != nil {
		t.Fatal(err)
	}
	if renamedKey.StreamID() != key.StreamID() || renamedKey.Relation() != key.Relation() || renamedKey.Namespace() != key.Namespace() {
		t.Fatalf("renamed ledger key = %+v, want same immutable target address as %+v", renamedKey, key)
	}

	restartedLedger, err := database.NewManagedTargetDeliveryLedger(store)
	if err != nil {
		t.Fatal(err)
	}
	got, found, err := restartedLedger.Lookup(context.Background(), renamedKey)
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if !found {
		t.Fatal("Lookup() found = false after rename and restart, want durable delivery record")
	}
	if got.DeliveryID() != record.DeliveryID() {
		t.Fatalf("Lookup() delivery ID = %q, want %q", got.DeliveryID(), record.DeliveryID())
	}
	if got := store.readCount(); got != 1 {
		t.Fatalf("durable store reads = %d, want 1", got)
	}
}

func TestManagedTargetDeliveryLedgerSeparatesRelations(t *testing.T) {
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "source-database",
		ConnectionID: "source-connection-1",
	}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal(err)
	}
	targetDatabase := testTargetDatabase(t, "database-1")
	firstKey := testManagedTargetDeliveryLedgerKey(t, owner, identity, "orders", "stream-orders", targetDatabase, "relation-orders")
	secondKey := testManagedTargetDeliveryLedgerKey(t, owner, identity, "invoices", "stream-invoices", targetDatabase, "relation-invoices")
	if firstKey.Namespace() != secondKey.Namespace() {
		t.Fatalf("same-owner ledger namespaces = %q and %q, want shared namespace", firstKey.Namespace(), secondKey.Namespace())
	}
	if firstKey.Relation() == secondKey.Relation() || firstKey.StreamID() == secondKey.StreamID() {
		t.Fatalf("sibling ledger identity collapsed: first=%+v second=%+v", firstKey, secondKey)
	}

	store := newManagedTargetDeliveryLedgerStoreFake()
	ledger, err := database.NewManagedTargetDeliveryLedger(store)
	if err != nil {
		t.Fatal(err)
	}
	first, err := database.NewManagedTargetDeliveryRecord("delivery-orders-1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := database.NewManagedTargetDeliveryRecord("delivery-invoices-1")
	if err != nil {
		t.Fatal(err)
	}
	firstReplacement, err := database.NewManagedTargetDeliveryRecord("delivery-orders-2")
	if err != nil {
		t.Fatal(err)
	}
	for _, write := range []struct {
		key    database.ManagedTargetDeliveryLedgerKey
		record database.ManagedTargetDeliveryRecord
	}{
		{key: firstKey, record: first},
		{key: secondKey, record: second},
		{key: firstKey, record: firstReplacement},
	} {
		if err := ledger.Record(context.Background(), write.key, write.record); err != nil {
			t.Fatalf("Record(%q) error = %v", write.record.DeliveryID(), err)
		}
	}
	if got := store.writeCount(); got != 3 {
		t.Fatalf("durable store writes = %d, want 3", got)
	}
	if got := store.entryCount(); got != 2 {
		t.Fatalf("durable store entries = %d, want 2 independent relations", got)
	}

	gotFirst, foundFirst, err := ledger.Lookup(context.Background(), firstKey)
	if err != nil {
		t.Fatalf("Lookup(first) error = %v", err)
	}
	gotSecond, foundSecond, err := ledger.Lookup(context.Background(), secondKey)
	if err != nil {
		t.Fatalf("Lookup(second) error = %v", err)
	}
	if !foundFirst || !foundSecond {
		t.Fatalf("independent ledger lookups found = %t, %t; want both records", foundFirst, foundSecond)
	}
	if gotFirst.DeliveryID() != firstReplacement.DeliveryID() {
		t.Fatalf("first ledger record = %q, want replacement %q", gotFirst.DeliveryID(), firstReplacement.DeliveryID())
	}
	if gotSecond.DeliveryID() != second.DeliveryID() {
		t.Fatalf("second ledger record changed to %q, want %q", gotSecond.DeliveryID(), second.DeliveryID())
	}
}

func TestManagedTargetDeliveryLedgerRejectsInvalidKeyBeforeStoreMutation(t *testing.T) {
	store := newManagedTargetDeliveryLedgerStoreFake()
	ledger, err := database.NewManagedTargetDeliveryLedger(store)
	if err != nil {
		t.Fatal(err)
	}
	record, err := database.NewManagedTargetDeliveryRecord("delivery-orders-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := ledger.Record(context.Background(), database.ManagedTargetDeliveryLedgerKey{}, record); err == nil {
		t.Fatal("Record() error = nil, want invalid key refusal")
	}
	if got := store.writeCount(); got != 0 {
		t.Fatalf("durable store writes after invalid key = %d, want 0", got)
	}
}

func TestManagedTargetDeliveryLedgerKeyBindsOwnerAndTargetDatabase(t *testing.T) {
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "source-database",
		ConnectionID: "source-connection-1",
	}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal(err)
	}
	targetDatabase := testTargetDatabase(t, "database-1")
	key := testManagedTargetDeliveryLedgerKey(t, owner, identity, "orders", "stream-orders", targetDatabase, "relation-orders")
	if !key.Owner().Identity().SameIdentity(owner.Identity()) {
		t.Fatalf("ledger owner = %#v, want asserted owner %#v", key.Owner(), owner)
	}
	if got := key.TargetDatabase(); got.Kind() != targetDatabase.Kind() || got.Value() != targetDatabase.Value() {
		t.Fatalf("ledger target database = %#v, want %#v", got, targetDatabase)
	}

	store := newManagedTargetDeliveryLedgerStoreFake()
	ledger, err := database.NewManagedTargetDeliveryLedger(store)
	if err != nil {
		t.Fatal(err)
	}
	record, err := database.NewManagedTargetDeliveryRecord("delivery-orders-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := ledger.Record(context.Background(), key, record); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	foreignIdentity := identity
	foreignIdentity.ConnectionID = "source-connection-2"
	foreignOwner, err := database.NewTargetOwner(foreignIdentity)
	if err != nil {
		t.Fatal(err)
	}
	foreignKey := testManagedTargetDeliveryLedgerKey(t, foreignOwner, foreignIdentity, "orders", "stream-orders", targetDatabase, "relation-orders")
	movedKey := testManagedTargetDeliveryLedgerKey(t, owner, identity, "orders", "stream-orders", testTargetDatabase(t, "database-2"), "relation-orders")
	for _, other := range []database.ManagedTargetDeliveryLedgerKey{foreignKey, movedKey} {
		_, found, err := ledger.Lookup(context.Background(), other)
		if err != nil {
			t.Fatalf("Lookup(%+v) error = %v", other, err)
		}
		if found {
			t.Fatalf("Lookup(%+v) found a record owned by a different source or target database", other)
		}
	}
	if got := store.entryCount(); got != 1 {
		t.Fatalf("durable store entries = %d, want owner/database isolation to retain one record", got)
	}
}

func testManagedTargetDeliveryLedgerKey(t *testing.T, owner database.TargetOwner, identity database.ConnectionIdentity, table, streamID string, targetDatabase database.TargetDatabaseIdentity, nativeValue string) database.ManagedTargetDeliveryLedgerKey {
	t.Helper()
	artifact, err := warehouse.NewArtifactRef(identity, table)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := database.NewManagedTargetRef(owner, artifact, streamID)
	if err != nil {
		t.Fatal(err)
	}
	control := testManagedTargetControl(
		t,
		owner,
		ref,
		targetDatabase,
		database.NativeRelationIdentity{Kind: "fixture-native-id", Value: nativeValue},
		testManagedTargetSchema(t, 1, 1),
	)
	key, err := database.NewManagedTargetDeliveryLedgerKey(control)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

type managedTargetDeliveryLedgerStoreFake struct {
	records map[database.ManagedTargetDeliveryLedgerKey]database.ManagedTargetDeliveryRecord
	writes  int
	reads   int
}

func newManagedTargetDeliveryLedgerStoreFake() *managedTargetDeliveryLedgerStoreFake {
	return &managedTargetDeliveryLedgerStoreFake{records: make(map[database.ManagedTargetDeliveryLedgerKey]database.ManagedTargetDeliveryRecord)}
}

func (s *managedTargetDeliveryLedgerStoreFake) LoadManagedTargetDelivery(_ context.Context, key database.ManagedTargetDeliveryLedgerKey) (database.ManagedTargetDeliveryRecord, bool, error) {
	s.reads++
	record, found := s.records[key]
	return record, found, nil
}

func (s *managedTargetDeliveryLedgerStoreFake) StoreManagedTargetDelivery(_ context.Context, key database.ManagedTargetDeliveryLedgerKey, record database.ManagedTargetDeliveryRecord) error {
	s.writes++
	s.records[key] = record
	return nil
}

func (s *managedTargetDeliveryLedgerStoreFake) writeCount() int { return s.writes }

func (s *managedTargetDeliveryLedgerStoreFake) readCount() int { return s.reads }

func (s *managedTargetDeliveryLedgerStoreFake) entryCount() int { return len(s.records) }
