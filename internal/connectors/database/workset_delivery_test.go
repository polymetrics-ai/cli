package database_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/warehouse"
)

// TestChangeDeliveryPlanSealsWorksetTargetBindings proves the delivery bridge
// cannot substitute a workset into a different managed target. The workset is
// a real immutable Parquet artifact rather than a synthetic row slice.
func TestChangeDeliveryPlanSealsWorksetTargetBindings(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	workset, err := database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if err != nil {
		t.Fatalf("DeriveChangeDeliveryWorkset() error = %v", err)
	}

	plan, err := database.NewChangeDeliveryPlan(context.Background(), database.ChangeDeliveryPlanRequest{
		Definition: testDatabaseWriteDefinition(t),
		Workset:    workset,
		Control:    request.Control,
		Mapping:    testChangeDeliveryMapping(t),
		BatchSize:  2,
	})
	if err != nil {
		t.Fatalf("NewChangeDeliveryPlan() error = %v", err)
	}
	if got, want := plan.WorksetIdentity(), workset.Identity(); got != want {
		t.Fatalf("sealed workset identity = %q, want %q", got, want)
	}
	if got, want := plan.RecordCount(), int(workset.Changes()); got != want {
		t.Fatalf("sealed change count = %d, want immutable workset delta count %d", got, want)
	}
	if got, want := plan.TombstoneCount(), int(workset.TombstoneCount()); got != want {
		t.Fatalf("sealed tombstone count = %d, want immutable workset tombstone count %d", got, want)
	}
	if plan.PlanHash() == "" {
		t.Fatal("sealed delivery plan did not expose its opaque immutable hash")
	}

	replacedControl, err := database.NewManagedTargetControlRecord(
		request.Control.Owner(),
		request.Control.Target(),
		request.Control.TargetDatabase(),
		database.NativeRelationIdentity{Kind: "fixture-native-id", Value: "relation-orders-replaced"},
		request.Control.Schema(),
	)
	if err != nil {
		t.Fatal(err)
	}
	replacedPlan, err := database.NewChangeDeliveryPlan(context.Background(), database.ChangeDeliveryPlanRequest{
		Definition: testDatabaseWriteDefinition(t),
		Workset:    workset,
		Control:    replacedControl,
		Mapping:    testChangeDeliveryMapping(t),
		BatchSize:  2,
	})
	if err != nil {
		t.Fatalf("NewChangeDeliveryPlan(replaced OID) error = %v", err)
	}
	if replacedPlan.PlanHash() == plan.PlanHash() {
		t.Fatal("native relation OID change reused the original delivery plan hash")
	}

	changedTarget := testChangeDeliveryWorksetControl(t, "destination-2", 1)
	_, err = database.NewChangeDeliveryPlan(context.Background(), database.ChangeDeliveryPlanRequest{
		Definition: testDatabaseWriteDefinition(t),
		Workset:    workset,
		Control:    changedTarget,
		Mapping:    testChangeDeliveryMapping(t),
		BatchSize:  2,
	})
	if !errors.Is(err, database.ErrChangeDeliveryPlanInvalid) {
		t.Fatalf("NewChangeDeliveryPlan(changed destination) error = %v, want immutable-target refusal", err)
	}
}

// TestChangeDeliveryExecutorPersistsCandidateBaselineAfterReceipt proves the
// baseline is a durable, destination-keyed copy of the sealed candidate only
// after the target executor returns a ledger-backed receipt.
func TestChangeDeliveryExecutorPersistsCandidateBaselineAfterReceipt(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	workset, err := database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if err != nil {
		t.Fatalf("DeriveChangeDeliveryWorkset() error = %v", err)
	}
	store, err := database.NewFileChangeDeliveryBaselineStore(t.TempDir(), 1<<20)
	if err != nil {
		t.Fatalf("NewFileChangeDeliveryBaselineStore() error = %v", err)
	}
	driver := &databaseWriteDriverFake{atomicOverwrite: true}
	writeExecutor := testDatabaseWriteExecutor(t, driver)
	executor, err := database.NewChangeDeliveryExecutor(writeExecutor, store)
	if err != nil {
		t.Fatalf("NewChangeDeliveryExecutor() error = %v", err)
	}
	plan := testChangeDeliveryPlan(t, workset, request.Control)
	preview, err := executor.Preview(context.Background(), plan)
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	approval, err := database.NewChangeDeliveryApproval(preview)
	if err != nil {
		t.Fatalf("NewChangeDeliveryApproval() error = %v", err)
	}
	result, err := executor.Execute(context.Background(), plan, approval)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, ok := result.Receipt(); !ok {
		t.Fatal("successful delivery did not expose receipt after baseline persistence")
	}
	key, err := database.NewManagedTargetDeliveryLedgerKey(request.Control)
	if err != nil {
		t.Fatal(err)
	}
	baseline, found, err := store.Lookup(context.Background(), key)
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if !found {
		t.Fatal("confirmed delivery did not persist a destination baseline")
	}
	if got, want := baseline.WorksetIdentity(), workset.Identity(); got != want {
		t.Fatalf("persisted baseline workset = %q, want %q", got, want)
	}
	if got, want := baseline.DeliveryID(), "delivery-session-1"; got != want {
		t.Fatalf("persisted baseline delivery receipt = %q, want %q", got, want)
	}
	var rows int
	if err := baseline.ReadCandidateBaseline(context.Background(), func(warehouse.Row) error {
		rows++
		return nil
	}); err != nil {
		t.Fatalf("ReadCandidateBaseline() error = %v", err)
	}
	if got, want := rows, 4; got != want {
		t.Fatalf("persisted candidate baseline rows = %d, want immutable projection count %d", got, want)
	}
}

// TestChangeDeliveryExecutorRetainsBaselineOnUncertainOrFailedDelivery proves
// neither a failed apply, an unknown commit, nor a missing durable receipt can
// overwrite a prior per-destination baseline. Each case retains the immutable
// workset identity required for an explicit replay decision.
func TestChangeDeliveryExecutorRetainsBaselineOnUncertainOrFailedDelivery(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	workset, err := database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if err != nil {
		t.Fatalf("DeriveChangeDeliveryWorkset() error = %v", err)
	}
	plan := testChangeDeliveryPlan(t, workset, request.Control)
	for _, tt := range []struct {
		name      string
		newWrite  func(*testing.T) *database.DatabaseWriteExecutor
		wantError error
	}{
		{
			name: "batch failure",
			newWrite: func(t *testing.T) *database.DatabaseWriteExecutor {
				return testDatabaseWriteExecutor(t, &databaseWriteDriverFake{atomicOverwrite: true, applyErrorAt: 1})
			},
			wantError: database.ErrDatabaseWriteBatchFailed,
		},
		{
			name: "unknown commit",
			newWrite: func(t *testing.T) *database.DatabaseWriteExecutor {
				return testDatabaseWriteExecutor(t, &databaseWriteDriverFake{atomicOverwrite: true, commitOutcome: database.CommitOutcomeUnknown})
			},
			wantError: database.ErrChangeDeliveryReplayRequired,
		},
		{
			name: "ledger receipt failure",
			newWrite: func(t *testing.T) *database.DatabaseWriteExecutor {
				driver := &databaseWriteDriverFake{atomicOverwrite: true}
				ledger, err := database.NewManagedTargetDeliveryLedger(changeDeliveryLedgerFailureStore{})
				if err != nil {
					t.Fatal(err)
				}
				executor, err := database.NewDatabaseWriteExecutor(driver, ledger)
				if err != nil {
					t.Fatal(err)
				}
				return executor
			},
			wantError: database.ErrDatabaseWriteReceiptUnavailable,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			baseline := &changeDeliveryBaselineStoreFake{worksetIdentity: "prior-workset", deliveryID: "prior-receipt"}
			executor, err := database.NewChangeDeliveryExecutor(tt.newWrite(t), baseline)
			if err != nil {
				t.Fatal(err)
			}
			preview, err := executor.Preview(context.Background(), plan)
			if err != nil {
				t.Fatalf("Preview() error = %v", err)
			}
			approval, err := database.NewChangeDeliveryApproval(preview)
			if err != nil {
				t.Fatal(err)
			}
			result, err := executor.Execute(context.Background(), plan, approval)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("Execute() error = %v, want %v", err, tt.wantError)
			}
			if got, want := result.WorksetIdentity(), workset.Identity(); got != want {
				t.Fatalf("replay workset identity = %q, want %q", got, want)
			}
			if baseline.writes != 0 || baseline.worksetIdentity != "prior-workset" || baseline.deliveryID != "prior-receipt" {
				t.Fatalf("failure advanced prior baseline: writes/workset/receipt = %d/%q/%q", baseline.writes, baseline.worksetIdentity, baseline.deliveryID)
			}
			if _, ok := result.Receipt(); ok {
				t.Fatal("failed or uncertain delivery exposed baseline-backed receipt")
			}
		})
	}
}

// TestFileChangeDeliveryBaselineStoreSeparatesConcurrentDestinations proves
// the baseline store derives its address from the complete owner/target key.
// Concurrent successful deliveries for distinct source connections cannot
// overwrite or read back each other's immutable candidate baseline.
func TestFileChangeDeliveryBaselineStoreSeparatesConcurrentDestinations(t *testing.T) {
	store, err := database.NewFileChangeDeliveryBaselineStore(t.TempDir(), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	firstRequest := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	secondRequest := testChangeDeliveryWorksetRequest(t, "destination-2", 1, []string{"tenant_id", "id"})
	secondRequest.Control = testChangeDeliveryWorksetControlForOwner(t, database.ConnectionIdentity{
		WorkspaceID:  "workspace-2",
		ConnectorID:  "postgres",
		ConnectionID: "source-connection-2",
	}, "destination-2")
	firstWorkset, err := database.DeriveChangeDeliveryWorkset(context.Background(), firstRequest)
	if err != nil {
		t.Fatal(err)
	}
	secondWorkset, err := database.DeriveChangeDeliveryWorkset(context.Background(), secondRequest)
	if err != nil {
		t.Fatal(err)
	}
	if firstWorkset.Identity() == secondWorkset.Identity() {
		t.Fatal("different workspace/connection destinations reused an immutable workset identity")
	}

	type deliveryCase struct {
		workset  database.ChangeDeliveryWorkset
		control  database.ManagedTargetControlRecord
		executor *database.ChangeDeliveryExecutor
		plan     database.ChangeDeliveryPlan
		approval *database.ChangeDeliveryApproval
	}
	cases := make([]deliveryCase, 0, 2)
	for _, input := range []struct {
		workset database.ChangeDeliveryWorkset
		control database.ManagedTargetControlRecord
	}{{workset: firstWorkset, control: firstRequest.Control}, {workset: secondWorkset, control: secondRequest.Control}} {
		writeExecutor := testDatabaseWriteExecutor(t, &databaseWriteDriverFake{atomicOverwrite: true})
		executor, err := database.NewChangeDeliveryExecutor(writeExecutor, store)
		if err != nil {
			t.Fatal(err)
		}
		plan := testChangeDeliveryPlan(t, input.workset, input.control)
		preview, err := executor.Preview(context.Background(), plan)
		if err != nil {
			t.Fatal(err)
		}
		approval, err := database.NewChangeDeliveryApproval(preview)
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, deliveryCase{workset: input.workset, control: input.control, executor: executor, plan: plan, approval: approval})
	}
	start := make(chan struct{})
	errs := make(chan error, len(cases))
	for _, tc := range cases {
		tc := tc
		go func() {
			<-start
			_, err := tc.executor.Execute(context.Background(), tc.plan, tc.approval)
			errs <- err
		}()
	}
	close(start)
	for range cases {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent delivery error = %v", err)
		}
	}
	for _, tc := range cases {
		key, err := database.NewManagedTargetDeliveryLedgerKey(tc.control)
		if err != nil {
			t.Fatal(err)
		}
		baseline, found, err := store.Lookup(context.Background(), key)
		if err != nil || !found || baseline.WorksetIdentity() != tc.workset.Identity() {
			t.Fatalf("destination baseline = (%q, %t, %v), want workset %q", baseline.WorksetIdentity(), found, err, tc.workset.Identity())
		}
	}
}

func testChangeDeliveryWorksetControlForOwner(t *testing.T, identity database.ConnectionIdentity, destination string) database.ManagedTargetControlRecord {
	t.Helper()
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := warehouse.NewArtifactRef(identity, "orders")
	if err != nil {
		t.Fatal(err)
	}
	target, err := database.NewManagedTargetRef(owner, artifact, "stream-orders")
	if err != nil {
		t.Fatal(err)
	}
	databaseIdentity, err := database.NewTargetDatabaseIdentity("fixture-target", destination)
	if err != nil {
		t.Fatal(err)
	}
	var fingerprint database.SchemaFingerprint
	copy(fingerprint[:], []byte(fmt.Sprintf("schema-%s", destination)))
	schema, err := database.NewManagedTargetSchema(1, fingerprint)
	if err != nil {
		t.Fatal(err)
	}
	control, err := database.NewManagedTargetControlRecord(owner, target, databaseIdentity, database.NativeRelationIdentity{Kind: "fixture-native-id", Value: "relation-orders-" + destination}, schema)
	if err != nil {
		t.Fatal(err)
	}
	return control
}

// TestChangeDeliveryExecutorRefusesStaleWorksetApprovalBeforeSessionMutation
// asserts the plan's workset binding is part of approval authority. Matching
// counts are intentionally insufficient: this second workset has a different
// immutable content address but the same managed target and mapping.
func TestChangeDeliveryExecutorRefusesStaleWorksetApprovalBeforeSessionMutation(t *testing.T) {
	request := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	workset, err := database.DeriveChangeDeliveryWorkset(context.Background(), request)
	if err != nil {
		t.Fatalf("DeriveChangeDeliveryWorkset(first) error = %v", err)
	}
	changedRequest := testChangeDeliveryWorksetRequest(t, "destination-1", 1, []string{"tenant_id", "id"})
	if err := warehouse.WriteTable(context.Background(), changedRequest.SourceParquet, []warehouse.Row{
		{"tenant_id": "north", "id": 1, "value": "revised", "nullable": nil},
		{"tenant_id": "north", "id": 2, "value": "same", "nullable": nil},
	}); err != nil {
		t.Fatalf("write changed source parquet: %v", err)
	}
	changedWorkset, err := database.DeriveChangeDeliveryWorkset(context.Background(), changedRequest)
	if err != nil {
		t.Fatalf("DeriveChangeDeliveryWorkset(changed) error = %v", err)
	}
	if changedWorkset.Identity() == workset.Identity() {
		t.Fatal("changed immutable workset reused the original identity")
	}

	firstPlan := testChangeDeliveryPlan(t, workset, request.Control)
	changedPlan := testChangeDeliveryPlan(t, changedWorkset, changedRequest.Control)
	driver := &databaseWriteDriverFake{atomicOverwrite: true}
	writeExecutor, ledger := testDatabaseWriteExecutorWithStore(t, driver)
	baseline := &changeDeliveryBaselineStoreFake{}
	executor, err := database.NewChangeDeliveryExecutor(writeExecutor, baseline)
	if err != nil {
		t.Fatalf("NewChangeDeliveryExecutor() error = %v", err)
	}
	preview, err := executor.Preview(context.Background(), firstPlan)
	if err != nil {
		t.Fatalf("Preview() error = %v", err)
	}
	approval, err := database.NewChangeDeliveryApproval(preview)
	if err != nil {
		t.Fatalf("NewChangeDeliveryApproval() error = %v", err)
	}
	if _, err := executor.Execute(context.Background(), changedPlan, approval); !errors.Is(err, database.ErrChangeDeliveryApprovalInvalid) {
		t.Fatalf("Execute(stale workset approval) error = %v, want immutable-workset approval refusal", err)
	}
	if driver.beginCalls != 0 || driver.batchCalls != 0 || driver.commitCalls != 0 || driver.rollbackCalls != 0 || ledger.writeCount() != 0 || baseline.writes != 0 {
		t.Fatalf("stale approval mutated target or baselines: begin/batch/commit/rollback/ledger/baseline = %d/%d/%d/%d/%d/%d", driver.beginCalls, driver.batchCalls, driver.commitCalls, driver.rollbackCalls, ledger.writeCount(), baseline.writes)
	}
}

func testChangeDeliveryPlan(t *testing.T, workset database.ChangeDeliveryWorkset, control database.ManagedTargetControlRecord) database.ChangeDeliveryPlan {
	t.Helper()
	plan, err := database.NewChangeDeliveryPlan(context.Background(), database.ChangeDeliveryPlanRequest{
		Definition: testDatabaseWriteDefinition(t),
		Workset:    workset,
		Control:    control,
		Mapping:    testChangeDeliveryMapping(t),
		BatchSize:  2,
	})
	if err != nil {
		t.Fatalf("NewChangeDeliveryPlan() error = %v", err)
	}
	return plan
}

type changeDeliveryBaselineStoreFake struct {
	writes          int
	worksetIdentity string
	deliveryID      string
}

func (s *changeDeliveryBaselineStoreFake) RecordChangeDeliveryBaseline(_ context.Context, _ database.ManagedTargetDeliveryLedgerKey, workset database.ChangeDeliveryWorkset, receipt database.DeliveryReceiptV1) error {
	s.writes++
	s.worksetIdentity = workset.Identity()
	s.deliveryID = receipt.DeliveryID()
	return nil
}

type changeDeliveryLedgerFailureStore struct{}

func (changeDeliveryLedgerFailureStore) LoadManagedTargetDelivery(context.Context, database.ManagedTargetDeliveryLedgerKey) (database.ManagedTargetDeliveryRecord, bool, error) {
	return database.ManagedTargetDeliveryRecord{}, false, nil
}

func (changeDeliveryLedgerFailureStore) StoreManagedTargetDelivery(context.Context, database.ManagedTargetDeliveryLedgerKey, database.ManagedTargetDeliveryRecord) error {
	return errors.New("fixture ledger persistence failure")
}

func testChangeDeliveryMapping(t *testing.T) database.MappingContractV1 {
	t.Helper()
	text, err := database.NewString(0, "")
	if err != nil {
		t.Fatal(err)
	}
	integer, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	textPlan, err := database.CompileTypePlan(text, text)
	if err != nil {
		t.Fatal(err)
	}
	integerPlan, err := database.CompileTypePlan(integer, integer)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := database.NewMappingContractV1([]database.MappingColumnV1{
		{Source: "tenant_id", Target: "tenant_id", Type: textPlan},
		{Source: "id", Target: "id", Type: integerPlan},
		{Source: "value", Target: "value", Type: textPlan},
		{Source: "nullable", Target: "nullable", Type: textPlan, Nullable: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	return mapping
}
