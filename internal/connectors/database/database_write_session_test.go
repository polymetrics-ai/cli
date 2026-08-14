package database_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

func TestDatabaseWriteExecutorRefusesApprovalPlanMismatchBeforeSessionMutation(t *testing.T) {
	definition := loadTestDefinition(t, strings.Replace(
		validDefinitionJSON,
		`"admitted_modes": []`,
		`"admitted_modes": ["full_append", "full_overwrite", "incremental_upsert", "incremental_dedupe"]`,
		1,
	))
	control := testDatabaseWriteControl(t, "orders", "stream-orders", 1)
	plan := testDatabaseWritePlan(t, definition, control, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 2, false)
	driver := &databaseWriteDriverFake{atomicOverwrite: true}
	executor := testDatabaseWriteExecutor(t, driver)
	preview, err := executor.Preview(context.Background(), plan)
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	approval, err := database.NewDatabaseWriteApproval(preview)
	if err != nil {
		t.Fatalf("NewDatabaseWriteApproval() error = %v", err)
	}

	changedPlan := testDatabaseWritePlan(t, definition, control, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 3, false)
	_, err = executor.Execute(context.Background(), changedPlan, approval, testDatabaseWriteRecords(3))
	if !errors.Is(err, database.ErrDatabaseWriteApprovalInvalid) {
		t.Fatalf("Execute(changed plan) error = %v, want ErrDatabaseWriteApprovalInvalid", err)
	}
	if got := driver.beginCalls; got != 0 {
		t.Fatalf("driver begin calls after approval mismatch = %d, want 0", got)
	}
	if got := driver.batchCalls; got != 0 {
		t.Fatalf("driver batch calls after approval mismatch = %d, want 0", got)
	}
	if got := driver.commitCalls; got != 0 {
		t.Fatalf("driver commit calls after approval mismatch = %d, want 0", got)
	}
	if got := driver.rollbackCalls; got != 0 {
		t.Fatalf("driver rollback calls after approval mismatch = %d, want 0", got)
	}
}

func testDatabaseWriteExecutor(t *testing.T, driver *databaseWriteDriverFake) *database.DatabaseWriteExecutor {
	t.Helper()
	ledger, err := database.NewManagedTargetDeliveryLedger(newManagedTargetDeliveryLedgerStoreFake())
	if err != nil {
		t.Fatal(err)
	}
	executor, err := database.NewDatabaseWriteExecutor(driver, ledger)
	if err != nil {
		t.Fatal(err)
	}
	return executor
}

func testDatabaseWriteControl(t *testing.T, table, streamID string, schemaMarker byte) database.ManagedTargetControlRecord {
	t.Helper()
	identity := database.ConnectionIdentity{
		WorkspaceID:  "workspace-1",
		ConnectorID:  "source-database",
		ConnectionID: "source-connection-1",
	}
	artifact, err := warehouse.NewArtifactRef(identity, table)
	if err != nil {
		t.Fatal(err)
	}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal(err)
	}
	ref, err := database.NewManagedTargetRef(owner, artifact, streamID)
	if err != nil {
		t.Fatal(err)
	}
	return testManagedTargetControl(
		t,
		owner,
		ref,
		testTargetDatabase(t, "database-1"),
		database.NativeRelationIdentity{Kind: "fixture-native-id", Value: "relation-" + streamID},
		testManagedTargetSchema(t, 1, schemaMarker),
	)
}

func testDatabaseWritePlan(t *testing.T, definition database.Definition, control database.ManagedTargetControlRecord, mode synccontract.Mode, strategy connectors.ApplyStrategy, keys []string, records, batchSize int, destructive bool) database.DatabaseWritePlan {
	t.Helper()
	plan, err := database.NewDatabaseWritePlan(context.Background(), database.DatabaseWritePlanRequest{
		Definition:  definition,
		Control:     control,
		Mode:        mode,
		Strategy:    strategy,
		Keys:        keys,
		RecordCount: records,
		BatchSize:   batchSize,
		Destructive: destructive,
	})
	if err != nil {
		t.Fatalf("NewDatabaseWritePlan() error = %v", err)
	}
	return plan
}

func testDatabaseWriteRecords(count int) []connectors.Record {
	records := make([]connectors.Record, count)
	for index := range records {
		records[index] = connectors.Record{"id": index + 1}
	}
	return records
}

type databaseWriteDriverFake struct {
	atomicOverwrite bool
	beginCalls      int
	batchCalls      int
	commitCalls     int
	rollbackCalls   int
}

func (d *databaseWriteDriverFake) DatabaseWriteCapabilities() database.DatabaseWriteCapabilities {
	return database.DatabaseWriteCapabilities{AtomicFullOverwrite: d.atomicOverwrite}
}

func (d *databaseWriteDriverFake) PreviewDatabaseWrite(_ context.Context, plan database.DatabaseWritePlan) (database.DatabaseWritePreview, error) {
	return database.NewDatabaseWritePreview(plan, "preview-1")
}

func (d *databaseWriteDriverFake) BeginDatabaseWrite(_ context.Context, _ database.DatabaseWritePlan) (database.WriteSession, error) {
	d.beginCalls++
	return d, nil
}

func (d *databaseWriteDriverFake) ApplyWriteBatch(_ context.Context, batch database.WriteBatch) error {
	d.batchCalls++
	if len(batch.Records()) == 0 {
		return errors.New("fake batch must carry records")
	}
	return nil
}

func (d *databaseWriteDriverFake) PublishFullOverwrite(context.Context) error { return nil }

func (d *databaseWriteDriverFake) CommitWrite(context.Context) (database.CommitOutcome, database.DatabaseWriteReceipt, error) {
	d.commitCalls++
	return database.CommitOutcomeUnknown, database.DatabaseWriteReceipt{}, errors.New("not reached by approval mismatch")
}

func (d *databaseWriteDriverFake) RollbackWrite(context.Context) error {
	d.rollbackCalls++
	return nil
}
