package database_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

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
	plan := testDatabaseWritePlan(t, definition, control, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 2, 2, false)
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

	changedPlan := testDatabaseWritePlan(t, definition, control, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 3, 2, false)
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

func TestDatabaseWriteExecutorBindsEveryApprovalPlanDimensionBeforeMutation(t *testing.T) {
	definition := testDatabaseWriteDefinition(t)
	control := testDatabaseWriteControl(t, "orders", "stream-orders", 1)
	for _, tt := range []struct {
		name    string
		base    func() database.DatabaseWritePlan
		changed func() database.DatabaseWritePlan
	}{
		{
			name: "target identity",
			base: func() database.DatabaseWritePlan {
				return testDatabaseWritePlan(t, definition, control, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 2, 2, false)
			},
			changed: func() database.DatabaseWritePlan {
				return testDatabaseWritePlan(t, definition, testDatabaseWriteControl(t, "invoices", "stream-invoices", 1), synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 2, 2, false)
			},
		},
		{
			name: "schema fingerprint",
			base: func() database.DatabaseWritePlan {
				return testDatabaseWritePlan(t, definition, control, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 2, 2, false)
			},
			changed: func() database.DatabaseWritePlan {
				return testDatabaseWritePlan(t, definition, testDatabaseWriteControl(t, "orders", "stream-orders", 2), synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 2, 2, false)
			},
		},
		{
			name: "mode and canonical strategy",
			base: func() database.DatabaseWritePlan {
				return testDatabaseWritePlan(t, definition, control, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 2, 2, false)
			},
			changed: func() database.DatabaseWritePlan {
				return testDatabaseWritePlan(t, definition, control, synccontract.ModeFullOverwrite, connectors.ApplyStrategyReplace, nil, 2, 2, true)
			},
		},
		{
			name: "stable keys",
			base: func() database.DatabaseWritePlan {
				return testDatabaseWritePlan(t, definition, control, synccontract.ModeIncrementalUpsert, connectors.ApplyStrategyMerge, []string{"id"}, 2, 2, false)
			},
			changed: func() database.DatabaseWritePlan {
				return testDatabaseWritePlan(t, definition, control, synccontract.ModeIncrementalUpsert, connectors.ApplyStrategyMerge, []string{"external_id"}, 2, 2, false)
			},
		},
		{
			name: "record count",
			base: func() database.DatabaseWritePlan {
				return testDatabaseWritePlan(t, definition, control, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 2, 2, false)
			},
			changed: func() database.DatabaseWritePlan {
				return testDatabaseWritePlan(t, definition, control, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 3, 2, false)
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			driver := &databaseWriteDriverFake{atomicOverwrite: true}
			executor, ledgerStore := testDatabaseWriteExecutorWithStore(t, driver)
			base, changed := tt.base(), tt.changed()
			approval := testDatabaseWriteApproval(t, executor, base)
			if _, err := executor.Execute(context.Background(), changed, approval, testDatabaseWriteRecords(changed.RecordCount())); !errors.Is(err, database.ErrDatabaseWriteApprovalInvalid) {
				t.Fatalf("Execute(changed %s) error = %v, want approval mismatch", tt.name, err)
			}
			if driver.beginCalls != 0 || driver.batchCalls != 0 || driver.commitCalls != 0 || driver.rollbackCalls != 0 || ledgerStore.writeCount() != 0 {
				t.Fatalf("changed %s mutated target: begin/batch/commit/rollback/ledger = %d/%d/%d/%d/%d", tt.name, driver.beginCalls, driver.batchCalls, driver.commitCalls, driver.rollbackCalls, ledgerStore.writeCount())
			}
		})
	}

	// full_overwrite must carry its destructive effect. An attempt to remove it
	// is refused while constructing the sealed plan, before preview/session work.
	driver := &databaseWriteDriverFake{atomicOverwrite: true}
	executor, ledgerStore := testDatabaseWriteExecutorWithStore(t, driver)
	_, err := database.NewDatabaseWritePlan(context.Background(), database.DatabaseWritePlanRequest{
		Definition:  definition,
		Control:     control,
		Mode:        synccontract.ModeFullOverwrite,
		Strategy:    connectors.ApplyStrategyReplace,
		Mapping:     testDatabaseWriteMapping(t, "id", "id"),
		RecordCount: 2,
		BatchSize:   2,
		Destructive: false,
	})
	if !errors.Is(err, database.ErrDatabaseWritePlanInvalid) {
		t.Fatalf("NewDatabaseWritePlan(without overwrite effect) error = %v, want plan refusal", err)
	}
	if driver.beginCalls != 0 || driver.batchCalls != 0 || driver.commitCalls != 0 || driver.rollbackCalls != 0 || ledgerStore.writeCount() != 0 {
		t.Fatalf("removed destructive effect mutated target: begin/batch/commit/rollback/ledger = %d/%d/%d/%d/%d", driver.beginCalls, driver.batchCalls, driver.commitCalls, driver.rollbackCalls, ledgerStore.writeCount())
	}
	_ = executor // Executor construction is intentionally also side-effect free.
}

func TestDatabaseWriteExecutorConsumesApprovalBeforeOnePinnedBoundedSession(t *testing.T) {
	definition := testDatabaseWriteDefinition(t)
	plan := testDatabaseWritePlan(t, definition, testDatabaseWriteControl(t, "orders", "stream-orders", 1), synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 6, 2, false)
	driver := &databaseWriteDriverFake{atomicOverwrite: true}
	executor, ledgerStore := testDatabaseWriteExecutorWithStore(t, driver)
	preview, err := executor.Preview(context.Background(), plan)
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	approval, err := database.NewDatabaseWriteApproval(preview)
	if err != nil {
		t.Fatal(err)
	}
	driver.approval = approval

	result, err := executor.Execute(context.Background(), plan, approval, testDatabaseWriteRecords(6))
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Outcome() != database.CommitOutcomeCommitted {
		t.Fatalf("Execute() outcome = %q, want committed", result.Outcome())
	}
	if !driver.approvalConsumedAtBegin {
		t.Fatal("approval was not consumed before BeginDatabaseWrite")
	}
	if !approval.Consumed() {
		t.Fatal("approval was not marked consumed after accepted execution")
	}
	if driver.beginCalls != 1 || driver.commitCalls != 1 || driver.rollbackCalls != 0 {
		t.Fatalf("session lifecycle begin/commit/rollback = %d/%d/%d, want 1/1/0", driver.beginCalls, driver.commitCalls, driver.rollbackCalls)
	}
	if !reflect.DeepEqual(driver.batchSizes, []int{2, 2, 2}) {
		t.Fatalf("pinned session batch sizes = %v, want [2 2 2]", driver.batchSizes)
	}
	if driver.legacyWriteCalls != 0 {
		t.Fatalf("legacy Connector.Write calls = %d, want 0", driver.legacyWriteCalls)
	}
	if ledgerStore.writeCount() != 1 {
		t.Fatalf("delivery ledger writes = %d, want 1 after confirmed commit", ledgerStore.writeCount())
	}
	receipt, ok := result.Receipt()
	if !ok || receipt.DeliveryID() != "delivery-session-1" || receipt.CommittedAt().IsZero() {
		t.Fatalf("DeliveryReceiptV1 = (%#v, %t), want committed session evidence after ledger record", receipt, ok)
	}
	acknowledgement, err := result.DownstreamAcknowledgement()
	if err != nil {
		t.Fatalf("DownstreamAcknowledgement() error = %v", err)
	}
	if acknowledgement.Sink != "fixture-target-database" || acknowledgement.AcknowledgedAt.IsZero() {
		t.Fatalf("checkpoint acknowledgement = %#v, want durable target-database evidence", acknowledgement)
	}
}

func TestDatabaseWriteExecutorRollsBackFailuresAndCancellationWithoutCheckpointAuthority(t *testing.T) {
	tests := []struct {
		name    string
		driver  func(context.CancelFunc) *databaseWriteDriverFake
		want    error
		batches []int
	}{
		{
			name: "batch failure",
			driver: func(context.CancelFunc) *databaseWriteDriverFake {
				return &databaseWriteDriverFake{atomicOverwrite: true, applyErrorAt: 2}
			},
			want:    database.ErrDatabaseWriteBatchFailed,
			batches: []int{2, 2},
		},
		{
			name: "cancellation between batches",
			driver: func(cancel context.CancelFunc) *databaseWriteDriverFake {
				return &databaseWriteDriverFake{atomicOverwrite: true, cancelAfterBatch: 1, cancel: cancel}
			},
			want:    context.Canceled,
			batches: []int{2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			driver := tt.driver(cancel)
			executor, ledgerStore := testDatabaseWriteExecutorWithStore(t, driver)
			plan := testDatabaseWritePlan(t, testDatabaseWriteDefinition(t), testDatabaseWriteControl(t, "orders", "stream-orders", 1), synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 6, 2, false)
			approval := testDatabaseWriteApproval(t, executor, plan)

			result, err := executor.Execute(ctx, plan, approval, testDatabaseWriteRecords(6))
			if !errors.Is(err, tt.want) {
				t.Fatalf("Execute() error = %v, want %v", err, tt.want)
			}
			if result.Outcome() != database.CommitOutcomeRolledBack {
				t.Fatalf("Execute() outcome = %q, want rolled_back", result.Outcome())
			}
			if driver.beginCalls != 1 || driver.rollbackCalls != 1 || driver.commitCalls != 0 {
				t.Fatalf("failed lifecycle begin/rollback/commit = %d/%d/%d, want 1/1/0", driver.beginCalls, driver.rollbackCalls, driver.commitCalls)
			}
			if !reflect.DeepEqual(driver.batchSizes, tt.batches) {
				t.Fatalf("failed session batches = %v, want %v", driver.batchSizes, tt.batches)
			}
			if ledgerStore.writeCount() != 0 {
				t.Fatalf("ledger writes after failed execution = %d, want 0", ledgerStore.writeCount())
			}
			if _, err := result.DownstreamAcknowledgement(); !errors.Is(err, database.ErrDatabaseWriteReceiptUnavailable) {
				t.Fatalf("failed result acknowledgement error = %v, want receipt unavailable", err)
			}
		})
	}
}

func TestDatabaseWriteExecutorTreatsUnknownCommitAsTerminalWithoutRetryOrRollback(t *testing.T) {
	driver := &databaseWriteDriverFake{atomicOverwrite: true, commitOutcome: database.CommitOutcomeUnknown, commitError: errors.New("fixture connection ended during commit")}
	executor, ledgerStore := testDatabaseWriteExecutorWithStore(t, driver)
	plan := testDatabaseWritePlan(t, testDatabaseWriteDefinition(t), testDatabaseWriteControl(t, "orders", "stream-orders", 1), synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 2, 2, false)
	approval := testDatabaseWriteApproval(t, executor, plan)

	result, err := executor.Execute(context.Background(), plan, approval, testDatabaseWriteRecords(2))
	if !errors.Is(err, database.ErrDatabaseWriteCommitOutcomeUnknown) {
		t.Fatalf("Execute() error = %v, want unknown commit outcome", err)
	}
	if result.Outcome() != database.CommitOutcomeUnknown {
		t.Fatalf("Execute() outcome = %q, want unknown", result.Outcome())
	}
	if driver.commitCalls != 1 || driver.rollbackCalls != 0 || driver.batchCalls != 1 {
		t.Fatalf("unknown outcome commit/rollback/batches = %d/%d/%d, want 1/0/1", driver.commitCalls, driver.rollbackCalls, driver.batchCalls)
	}
	if ledgerStore.writeCount() != 0 {
		t.Fatalf("ledger writes after unknown commit = %d, want 0", ledgerStore.writeCount())
	}
	if _, err := result.DownstreamAcknowledgement(); !errors.Is(err, database.ErrDatabaseWriteReceiptUnavailable) {
		t.Fatalf("unknown result acknowledgement error = %v, want receipt unavailable", err)
	}
	_, replayErr := executor.Execute(context.Background(), plan, approval, testDatabaseWriteRecords(2))
	if !errors.Is(replayErr, database.ErrDatabaseWriteApprovalConsumed) {
		t.Fatalf("Execute() replay error = %v, want consumed approval", replayErr)
	}
	if driver.commitCalls != 1 || driver.rollbackCalls != 0 || driver.batchCalls != 1 {
		t.Fatalf("unknown outcome was retried or relabelled: commit/rollback/batches = %d/%d/%d", driver.commitCalls, driver.rollbackCalls, driver.batchCalls)
	}
}

func TestDatabaseWriteExecutorRequiresAtomicOverwriteAndPinsCanonicalStrategies(t *testing.T) {
	t.Run("non atomic overwrite refuses before mutation", func(t *testing.T) {
		driver := &databaseWriteDriverFake{atomicOverwrite: false}
		executor, ledgerStore := testDatabaseWriteExecutorWithStore(t, driver)
		plan := testDatabaseWritePlan(t, testDatabaseWriteDefinition(t), testDatabaseWriteControl(t, "orders", "stream-orders", 1), synccontract.ModeFullOverwrite, connectors.ApplyStrategyReplace, nil, 2, 1, true)
		approval := testDatabaseWriteApproval(t, executor, plan)
		if _, err := executor.Execute(context.Background(), plan, approval, testDatabaseWriteRecords(2)); !errors.Is(err, database.ErrDatabaseWritePlanInvalid) {
			t.Fatalf("Execute() error = %v, want atomic-overwrite refusal", err)
		}
		if driver.beginCalls != 0 || driver.batchCalls != 0 || driver.commitCalls != 0 || driver.rollbackCalls != 0 || ledgerStore.writeCount() != 0 {
			t.Fatalf("non-atomic overwrite mutated: lifecycle=%d/%d/%d/%d ledger=%d", driver.beginCalls, driver.batchCalls, driver.commitCalls, driver.rollbackCalls, ledgerStore.writeCount())
		}
	})

	t.Run("atomic overwrite publishes once after batches", func(t *testing.T) {
		driver := &databaseWriteDriverFake{atomicOverwrite: true}
		executor, _ := testDatabaseWriteExecutorWithStore(t, driver)
		plan := testDatabaseWritePlan(t, testDatabaseWriteDefinition(t), testDatabaseWriteControl(t, "orders", "stream-orders", 1), synccontract.ModeFullOverwrite, connectors.ApplyStrategyReplace, nil, 2, 1, true)
		if _, err := executor.Execute(context.Background(), plan, testDatabaseWriteApproval(t, executor, plan), testDatabaseWriteRecords(2)); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if !reflect.DeepEqual(driver.events, []string{"begin", "batch", "batch", "publish", "commit"}) {
			t.Fatalf("atomic overwrite events = %v, want batches then one publish and commit", driver.events)
		}
	})

	for _, tt := range []struct {
		name     string
		mode     synccontract.Mode
		strategy connectors.ApplyStrategy
		keys     []string
	}{
		{name: "append", mode: synccontract.ModeFullAppend, strategy: connectors.ApplyStrategyAppend},
		{name: "upsert", mode: synccontract.ModeIncrementalUpsert, strategy: connectors.ApplyStrategyMerge, keys: []string{"id"}},
		{name: "dedupe", mode: synccontract.ModeIncrementalDedupe, strategy: connectors.ApplyStrategyDedupe, keys: []string{"id"}},
	} {
		t.Run(tt.name+" uses one pinned canonical strategy", func(t *testing.T) {
			driver := &databaseWriteDriverFake{atomicOverwrite: true}
			executor, _ := testDatabaseWriteExecutorWithStore(t, driver)
			plan := testDatabaseWritePlan(t, testDatabaseWriteDefinition(t), testDatabaseWriteControl(t, "orders", "stream-orders", 1), tt.mode, tt.strategy, tt.keys, 1, 1, false)
			if _, err := executor.Execute(context.Background(), plan, testDatabaseWriteApproval(t, executor, plan), testDatabaseWriteRecords(1)); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if driver.beginCalls != 1 || driver.plan.Strategy() != tt.strategy || driver.legacyWriteCalls != 0 {
				t.Fatalf("strategy session begin/strategy/legacy = %d/%q/%d, want 1/%q/0", driver.beginCalls, driver.plan.Strategy(), driver.legacyWriteCalls, tt.strategy)
			}
		})
	}
}

func TestDatabaseWriteExecutorRequiresDefinitionCompatibleDriverAndRegistryAdmission(t *testing.T) {
	definition := testDatabaseWriteDefinition(t)
	plan := testDatabaseWritePlan(t, definition, testDatabaseWriteControl(t, "orders", "stream-orders", 1), synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 1, 1, false)

	t.Run("mismatched descriptor cannot preview or mutate", func(t *testing.T) {
		driver := &databaseWriteDriverFake{descriptor: database.DriverDescriptor{ID: "mysql", Protocol: "mysql-wire", APIVersion: 1}, atomicOverwrite: true}
		executor, ledgerStore := testDatabaseWriteExecutorWithStore(t, driver)
		if _, err := executor.Preview(context.Background(), plan); !errors.Is(err, database.ErrDatabaseWritePreviewUnavailable) {
			t.Fatalf("Preview() error = %v, want incompatible driver refusal", err)
		}
		if driver.beginCalls != 0 || driver.batchCalls != 0 || driver.commitCalls != 0 || driver.rollbackCalls != 0 || ledgerStore.writeCount() != 0 {
			t.Fatalf("incompatible driver mutated: begin/batch/commit/rollback/ledger = %d/%d/%d/%d/%d", driver.beginCalls, driver.batchCalls, driver.commitCalls, driver.rollbackCalls, ledgerStore.writeCount())
		}
	})

	t.Run("registry resolves only explicit write port", func(t *testing.T) {
		writeDriver := &databaseWriteDriverFake{atomicOverwrite: true}
		registry, err := database.NewDriverRegistry(writeDriver)
		if err != nil {
			t.Fatal(err)
		}
		resolved, err := registry.ResolveWriteDriver(context.Background(), definition)
		if err != nil {
			t.Fatalf("ResolveWriteDriver() error = %v", err)
		}
		if resolved != writeDriver {
			t.Fatalf("ResolveWriteDriver() = %#v, want registered fake", resolved)
		}

		descriptorOnly, err := database.NewDriverRegistry(declaredDriver{descriptor: postgresDriverDescriptor()})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := descriptorOnly.ResolveWriteDriver(context.Background(), definition); !errors.Is(err, database.ErrDatabaseWriteSessionUnavailable) {
			t.Fatalf("ResolveWriteDriver(descriptor only) error = %v, want write-port refusal", err)
		}
	})
}

func testDatabaseWriteExecutor(t *testing.T, driver *databaseWriteDriverFake) *database.DatabaseWriteExecutor {
	t.Helper()
	executor, _ := testDatabaseWriteExecutorWithStore(t, driver)
	return executor
}

func testDatabaseWriteExecutorWithStore(t *testing.T, driver *databaseWriteDriverFake) (*database.DatabaseWriteExecutor, *managedTargetDeliveryLedgerStoreFake) {
	t.Helper()
	store := newManagedTargetDeliveryLedgerStoreFake()
	ledger, err := database.NewManagedTargetDeliveryLedger(store)
	if err != nil {
		t.Fatal(err)
	}
	executor, err := database.NewDatabaseWriteExecutor(driver, ledger)
	if err != nil {
		t.Fatal(err)
	}
	return executor, store
}

func testDatabaseWriteDefinition(t *testing.T) database.Definition {
	t.Helper()
	return loadTestDefinition(t, strings.Replace(
		validDefinitionJSON,
		`"admitted_modes": []`,
		`"admitted_modes": ["full_append", "full_overwrite", "incremental_upsert", "incremental_dedupe"]`,
		1,
	))
}

func testDatabaseWriteApproval(t *testing.T, executor *database.DatabaseWriteExecutor, plan database.DatabaseWritePlan) *database.DatabaseWriteApproval {
	t.Helper()
	preview, err := executor.Preview(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	approval, err := database.NewDatabaseWriteApproval(preview)
	if err != nil {
		t.Fatal(err)
	}
	return approval
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
	mapping := testDatabaseWriteMapping(t, "id", "id")
	if len(keys) > 0 {
		mapping = testDatabaseWriteMapping(t, keys[0], keys[0])
	}
	plan, err := database.NewDatabaseWritePlan(context.Background(), database.DatabaseWritePlanRequest{
		Definition:  definition,
		Control:     control,
		Mode:        mode,
		Strategy:    strategy,
		Mapping:     mapping,
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

func testDatabaseWriteMapping(t *testing.T, source, target string) database.MappingContractV1 {
	t.Helper()
	logical, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	typePlan, err := database.CompileTypePlan(logical, logical)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := database.NewMappingContractV1([]database.MappingColumnV1{{
		Source: source,
		Target: target,
		Type:   typePlan,
	}})
	if err != nil {
		t.Fatal(err)
	}
	return mapping
}

func testDatabaseWriteRecords(count int) []connectors.Record {
	records := make([]connectors.Record, count)
	for index := range records {
		records[index] = connectors.Record{"id": index + 1}
	}
	return records
}

type databaseWriteDriverFake struct {
	descriptor              database.DriverDescriptor
	atomicOverwrite         bool
	approval                *database.DatabaseWriteApproval
	approvalConsumedAtBegin bool
	plan                    database.DatabaseWritePlan
	beginCalls              int
	batchCalls              int
	batchSizes              []int
	targetRows              map[string]bool
	commitCalls             int
	rollbackCalls           int
	publishCalls            int
	legacyWriteCalls        int
	applyErrorAt            int
	cancelAfterBatch        int
	cancel                  context.CancelFunc
	commitOutcome           database.CommitOutcome
	commitError             error
	events                  []string
}

func (d *databaseWriteDriverFake) DatabaseDriverDescriptor() database.DriverDescriptor {
	if d.descriptor != (database.DriverDescriptor{}) {
		return d.descriptor
	}
	return postgresDriverDescriptor()
}

func (d *databaseWriteDriverFake) DatabaseWriteCapabilities() database.DatabaseWriteCapabilities {
	return database.DatabaseWriteCapabilities{AtomicFullOverwrite: d.atomicOverwrite}
}

func (d *databaseWriteDriverFake) PreviewDatabaseWrite(_ context.Context, plan database.DatabaseWritePlan) (database.DatabaseWritePreview, error) {
	return database.NewDatabaseWritePreview(plan, "preview-1")
}

func (d *databaseWriteDriverFake) BeginDatabaseWrite(_ context.Context, plan database.DatabaseWritePlan) (database.WriteSession, error) {
	d.beginCalls++
	d.plan = plan
	d.events = append(d.events, "begin")
	if d.approval != nil {
		d.approvalConsumedAtBegin = d.approval.Consumed()
	}
	return d, nil
}

func (d *databaseWriteDriverFake) ApplyWriteBatch(_ context.Context, batch database.WriteBatch) error {
	d.batchCalls++
	records := batch.Records()
	tombstones := batch.Tombstones()
	d.batchSizes = append(d.batchSizes, len(records)+len(tombstones))
	d.events = append(d.events, "batch")
	if len(records)+len(tombstones) == 0 {
		return errors.New("fake batch must carry records or explicit tombstones")
	}
	if d.targetRows != nil {
		for _, tombstone := range tombstones {
			var key map[string]string
			if err := json.Unmarshal(tombstone.Key, &key); err != nil {
				return err
			}
			if id, ok := key["target_id"]; ok {
				delete(d.targetRows, id)
			}
		}
	}
	if d.cancelAfterBatch == d.batchCalls && d.cancel != nil {
		d.cancel()
	}
	if d.applyErrorAt == d.batchCalls {
		return errors.New("fixture batch failure")
	}
	return nil
}

func (d *databaseWriteDriverFake) PublishFullOverwrite(context.Context) error {
	d.publishCalls++
	d.events = append(d.events, "publish")
	return nil
}

func (d *databaseWriteDriverFake) CommitWrite(context.Context) (database.CommitOutcome, database.DeliveryReceiptV1, error) {
	d.commitCalls++
	d.events = append(d.events, "commit")
	outcome := d.commitOutcome
	if outcome == "" {
		outcome = database.CommitOutcomeCommitted
	}
	if outcome != database.CommitOutcomeCommitted {
		return outcome, database.DeliveryReceiptV1{}, d.commitError
	}
	receipt, err := database.NewDeliveryReceiptV1(d.plan, "delivery-session-1", time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC))
	if err != nil {
		return database.CommitOutcomeUnknown, database.DeliveryReceiptV1{}, err
	}
	return outcome, receipt, d.commitError
}

func (d *databaseWriteDriverFake) RollbackWrite(context.Context) error {
	d.rollbackCalls++
	d.events = append(d.events, "rollback")
	return nil
}

// Write is the legacy connector-shaped trap. DatabaseWriteExecutor is only
// given the narrower driver port, so any accidental per-record fallback would
// make this counter observable and fail its tests.
func (d *databaseWriteDriverFake) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	d.legacyWriteCalls++
	return connectors.WriteResult{}, errors.New("legacy connector write must not be used")
}
