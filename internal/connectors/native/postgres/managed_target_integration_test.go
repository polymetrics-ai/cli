//go:build databaseintegration

package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/native/dbtest"
	native "polymetrics.ai/internal/connectors/native/postgres"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

const postgresManagedTargetRestrictedUser = "pm_managed_target_no_control"

// TestPostgresManagedTargetWorksetDeliveryLive proves the immutable Parquet
// workset consumer against the managed PostgreSQL driver. It observes target
// rows after every delivery: source absence retains a prior row while only an
// explicit tombstone removes it. The mapping deliberately renames source keys
// to ensure tombstones traverse MappingContractV1 before PostgreSQL applies
// their target-key predicates.
func TestPostgresManagedTargetWorksetDeliveryLive(t *testing.T) {
	if os.Getenv(postgresCatalogIntegrationEnabledEnv) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the PostgreSQL Docker or Podman proof", postgresCatalogIntegrationEnabledEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	harness, err := newPostgresCatalogContainerHarness(
		dbtest.Runtime(strings.TrimSpace(os.Getenv(postgresCatalogIntegrationRuntimeEnv))),
		strings.TrimSpace(os.Getenv(postgresCatalogIntegrationEndpointEnv)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Error("PostgreSQL database test cleanup failed")
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatal("PostgreSQL database container did not start")
	}
	waitForPostgresCatalog(t, ctx, native.New(), postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema))
	source := openPostgresCatalogSource(t, ctx, endpoint)
	defer func() { _ = source.Close(context.WithoutCancel(ctx)) }()
	driver, err := native.NewDatabaseDriver(source)
	if err != nil {
		t.Fatal("could not construct PostgreSQL managed target driver")
	}
	fixture := newPostgresManagedTargetFixture(t, ctx, driver)
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal("could not construct managed target provisioner")
	}
	control, err := provisioner.CreateOrAssert(ctx, fixture.plan)
	if err != nil {
		t.Fatal("could not create PostgreSQL managed workset target")
	}
	ledger, err := database.NewManagedTargetDeliveryLedger(driver)
	if err != nil {
		t.Fatal("could not construct managed target delivery ledger")
	}
	writeExecutor, err := database.NewDatabaseWriteExecutor(driver, ledger)
	if err != nil {
		t.Fatal("could not construct PostgreSQL workset write executor")
	}
	baselineStore, err := database.NewFileChangeDeliveryBaselineStore(t.TempDir(), 1<<20)
	if err != nil {
		t.Fatal("could not construct durable local baseline store")
	}
	delivery, err := database.NewChangeDeliveryExecutor(writeExecutor, baselineStore)
	if err != nil {
		t.Fatal("could not construct PostgreSQL workset delivery executor")
	}
	root := t.TempDir()

	firstSource := []warehouse.Row{
		{"source_tenant": "north", "source_id": int64(1), "source_value": "old"},
		{"source_tenant": "retain", "source_id": int64(9), "source_value": "keep"},
	}
	emptyBaseline := writePostgresWorksetParquet(t, ctx, root, "empty-baseline", nil)
	firstWorkset := derivePostgresWorkset(t, ctx, root, control, firstSource, emptyBaseline, nil)
	firstReceipt := executePostgresWorksetDelivery(t, ctx, delivery, fixture, control, firstWorkset)
	assertPostgresManagedTargetRows(t, ctx, source, control, []postgresManagedTargetRow{
		{tenant: "north", id: 1, value: "old"},
		{tenant: "retain", id: 9, value: "keep"},
	})
	assertPostgresManagedTargetReceiptPersisted(t, ctx, driver, control, firstReceipt)

	firstBaseline := lookupPostgresWorksetBaseline(t, ctx, baselineStore, control)
	secondBaseline := writePostgresWorksetParquet(t, ctx, root, "second-baseline", readPostgresWorksetBaselineRows(t, ctx, firstBaseline))
	secondSource := []warehouse.Row{
		{"source_tenant": "north", "source_id": int64(1), "source_value": "updated"},
		{"source_tenant": "south", "source_id": int64(2), "source_value": "new"},
	}
	secondWorkset := derivePostgresWorkset(t, ctx, root, control, secondSource, secondBaseline, nil)
	secondReceipt := executePostgresWorksetDelivery(t, ctx, delivery, fixture, control, secondWorkset)
	// retain/9 is physically absent from this projection and baseline delta, but
	// stays in PostgreSQL because the workset has no explicit tombstone for it.
	assertPostgresManagedTargetRows(t, ctx, source, control, []postgresManagedTargetRow{
		{tenant: "north", id: 1, value: "updated"},
		{tenant: "retain", id: 9, value: "keep"},
		{tenant: "south", id: 2, value: "new"},
	})
	assertPostgresManagedTargetReceiptPersisted(t, ctx, driver, control, secondReceipt)

	secondBaselineRecord := lookupPostgresWorksetBaseline(t, ctx, baselineStore, control)
	thirdBaseline := writePostgresWorksetParquet(t, ctx, root, "third-baseline", readPostgresWorksetBaselineRows(t, ctx, secondBaselineRecord))
	deleteRetained := synccontract.Tombstone{
		Operation:   synccontract.OperationDelete,
		EventID:     synccontract.OpaqueToken("postgres-workset-delete-retain-9"),
		Key:         json.RawMessage(`{"source_tenant":"retain","source_id":9}`),
		DeleteImage: synccontract.DeleteImageKeyOnly,
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken("00000022"),
			TieBreaker: synccontract.OpaqueToken("00000001"),
		},
	}
	thirdWorkset := derivePostgresWorkset(t, ctx, root, control, secondSource, thirdBaseline, []synccontract.Tombstone{deleteRetained})
	thirdReceipt := executePostgresWorksetDelivery(t, ctx, delivery, fixture, control, thirdWorkset)
	assertPostgresManagedTargetRows(t, ctx, source, control, []postgresManagedTargetRow{
		{tenant: "north", id: 1, value: "updated"},
		{tenant: "south", id: 2, value: "new"},
	})
	assertPostgresManagedTargetReceiptPersisted(t, ctx, driver, control, thirdReceipt)
	finalBaseline := lookupPostgresWorksetBaseline(t, ctx, baselineStore, control)
	if got, want := finalBaseline.WorksetIdentity(), thirdWorkset.Identity(); got != want {
		t.Fatalf("durable workset baseline identity = %q, want final immutable workset %q", got, want)
	}
	if got, want := finalBaseline.DeliveryID(), thirdReceipt.DeliveryID(); got != want {
		t.Fatalf("durable workset baseline receipt = %q, want PostgreSQL receipt %q", got, want)
	}
}

// TestPostgresManagedTargetIncrementalDedupeHistoryLive proves the only
// supported history route from its first version through a process restart.
// It observes the retained closed row, exact validity boundary, replay
// stability, durable receipt, and soft-delete state rather than treating a
// successful call as evidence.
func TestPostgresManagedTargetIncrementalDedupeHistoryLive(t *testing.T) {
	if os.Getenv(postgresCatalogIntegrationEnabledEnv) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the PostgreSQL Docker or Podman proof", postgresCatalogIntegrationEnabledEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	harness, err := newPostgresCatalogContainerHarness(
		dbtest.Runtime(strings.TrimSpace(os.Getenv(postgresCatalogIntegrationRuntimeEnv))),
		strings.TrimSpace(os.Getenv(postgresCatalogIntegrationEndpointEnv)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Error("PostgreSQL database test cleanup failed")
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatal("PostgreSQL database container did not start")
	}
	waitForPostgresCatalog(t, ctx, native.New(), postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema))
	source := openPostgresCatalogSource(t, ctx, endpoint)
	driver, err := native.NewDatabaseDriver(source)
	if err != nil {
		t.Fatal("could not construct PostgreSQL history target driver")
	}
	fixture := newPostgresManagedTargetFixture(t, ctx, driver)
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal("could not construct PostgreSQL history target provisioner")
	}
	control, err := provisioner.CreateOrAssert(ctx, fixture.plan)
	if err != nil {
		t.Fatal("could not create PostgreSQL history target")
	}
	ledger, err := database.NewManagedTargetDeliveryLedger(driver)
	if err != nil {
		t.Fatal("could not construct PostgreSQL history delivery ledger")
	}
	executor, err := database.NewDatabaseWriteExecutor(driver, ledger)
	if err != nil {
		t.Fatal("could not construct PostgreSQL history write executor")
	}
	definition := postgresManagedTargetWriteDefinition(t)
	keys := []string{"tenant", "id"}
	first := connectors.Record{"source_tenant": "north", "source_id": int64(1), "source_value": "v1"}
	second := connectors.Record{"source_tenant": "north", "source_id": int64(1), "source_value": "v2"}

	firstPlan := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupeHistory, connectors.ApplyStrategyDedupeHistory, keys, 1, 0, false)
	firstReceipt := postgresExecuteManagedTargetWrite(t, ctx, executor, firstPlan, []connectors.Record{first}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetReceiptPersisted(t, ctx, driver, control, firstReceipt)
	initialRows := postgresManagedTargetHistoryRows(t, ctx, source, control)
	if len(initialRows) != 1 || initialRows[0].value != "v1" || !initialRows[0].current || initialRows[0].validTo != nil || initialRows[0].validFrom.IsZero() {
		t.Fatalf("first history version state = %#v, want one open current v1 row", initialRows)
	}

	secondPlan := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupeHistory, connectors.ApplyStrategyDedupeHistory, keys, 1, 0, false)
	secondReceipt := postgresExecuteManagedTargetWrite(t, ctx, executor, secondPlan, []connectors.Record{second}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetReceiptPersisted(t, ctx, driver, control, secondReceipt)
	versionedRows := postgresManagedTargetHistoryRows(t, ctx, source, control)
	if len(versionedRows) != 2 || versionedRows[0].value != "v1" || versionedRows[0].current || versionedRows[0].validTo == nil || versionedRows[1].value != "v2" || !versionedRows[1].current || versionedRows[1].validTo != nil || !versionedRows[0].validTo.Equal(versionedRows[1].validFrom) {
		t.Fatalf("version transition rows = %#v, want closed v1 joined exactly to open current v2", versionedRows)
	}

	// A fresh native connection and executor prove replay/close behavior after
	// recovery, not merely through the first driver's in-process state.
	if err := source.Close(context.WithoutCancel(ctx)); err != nil {
		t.Fatal("could not close initial PostgreSQL history driver connection")
	}
	restartedSource := openPostgresCatalogSource(t, ctx, endpoint)
	defer func() { _ = restartedSource.Close(context.WithoutCancel(ctx)) }()
	restartedDriver, err := native.NewDatabaseDriver(restartedSource)
	if err != nil {
		t.Fatal("could not reconstruct PostgreSQL history target driver")
	}
	restartedLedger, err := database.NewManagedTargetDeliveryLedger(restartedDriver)
	if err != nil {
		t.Fatal("could not reconstruct PostgreSQL history ledger")
	}
	restartedExecutor, err := database.NewDatabaseWriteExecutor(restartedDriver, restartedLedger)
	if err != nil {
		t.Fatal("could not reconstruct PostgreSQL history executor")
	}
	replayPlan := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupeHistory, connectors.ApplyStrategyDedupeHistory, keys, 2, 0, false)
	replayReceipt := postgresExecuteManagedTargetWrite(t, ctx, restartedExecutor, replayPlan, []connectors.Record{first, second}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetReceiptPersisted(t, ctx, restartedDriver, control, replayReceipt)
	if got := postgresManagedTargetHistoryRows(t, ctx, restartedSource, control); !samePostgresManagedTargetHistoryRows(got, versionedRows) {
		t.Fatalf("late/replay delivery changed durable history: got=%#v want=%#v", got, versionedRows)
	}

	tombstone := synccontract.Tombstone{
		Operation:   synccontract.OperationDelete,
		EventID:     synccontract.OpaqueToken("postgres-history-delete-north-1"),
		Key:         json.RawMessage(`{"tenant":"north","id":1}`),
		DeleteImage: synccontract.DeleteImageKeyOnly,
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken("00000051"),
			TieBreaker: synccontract.OpaqueToken("00000001"),
		},
	}
	envelope, err := database.NewTombstoneEnvelope([]synccontract.Tombstone{tombstone})
	if err != nil {
		t.Fatal("could not construct PostgreSQL history soft-delete envelope")
	}
	closePlan := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupeHistory, connectors.ApplyStrategyDedupeHistory, keys, 0, 1, false)
	closeReceipt := postgresExecuteManagedTargetWrite(t, ctx, restartedExecutor, closePlan, nil, envelope)
	assertPostgresManagedTargetReceiptPersisted(t, ctx, restartedDriver, control, closeReceipt)
	closedRows := postgresManagedTargetHistoryRows(t, ctx, restartedSource, control)
	if len(closedRows) != 2 || closedRows[0].current || closedRows[0].validTo == nil || closedRows[1].current || closedRows[1].validTo == nil || versionedRows[0].validTo == nil || !closedRows[0].validTo.Equal(*versionedRows[0].validTo) || !closedRows[1].validTo.After(closedRows[1].validFrom) {
		t.Fatalf("history soft-delete rows = %#v, want both versions retained and v2 validity closed", closedRows)
	}
}

func derivePostgresWorkset(t *testing.T, ctx context.Context, root string, control database.ManagedTargetControlRecord, sourceRows []warehouse.Row, baselinePath string, tombstones []synccontract.Tombstone) database.ChangeDeliveryWorkset {
	t.Helper()
	sourcePath := writePostgresWorksetParquet(t, ctx, root, "source", sourceRows)
	workset, err := database.DeriveChangeDeliveryWorkset(ctx, database.ChangeDeliveryWorksetRequest{
		Control:          control,
		Keys:             []string{"source_tenant", "source_id"},
		SourceParquet:    sourcePath,
		BaselineParquet:  baselinePath,
		Tombstones:       tombstones,
		Root:             root,
		MaxArtifactBytes: 1 << 20,
	})
	if err != nil {
		t.Fatalf("DeriveChangeDeliveryWorkset() error = %v", err)
	}
	return workset
}

func writePostgresWorksetParquet(t *testing.T, ctx context.Context, root, name string, rows []warehouse.Row) string {
	t.Helper()
	path := filepath.Join(root, name+".parquet")
	if err := warehouse.WriteTable(ctx, path, rows); err != nil {
		t.Fatalf("write workset Parquet %q: %v", name, err)
	}
	return path
}

func executePostgresWorksetDelivery(t *testing.T, ctx context.Context, delivery *database.ChangeDeliveryExecutor, fixture postgresManagedTargetFixture, control database.ManagedTargetControlRecord, workset database.ChangeDeliveryWorkset) database.DeliveryReceiptV1 {
	t.Helper()
	mapping, ok := fixture.plan.Mapping()
	if !ok {
		t.Fatal("managed PostgreSQL fixture lost its mapping")
	}
	plan, err := database.NewChangeDeliveryPlan(ctx, database.ChangeDeliveryPlanRequest{
		Definition: postgresManagedTargetWriteDefinition(t),
		Workset:    workset,
		Control:    control,
		Mapping:    mapping,
		BatchSize:  2,
	})
	if err != nil {
		t.Fatalf("NewChangeDeliveryPlan() error = %v", err)
	}
	preview, err := delivery.Preview(ctx, plan)
	if err != nil {
		t.Fatalf("ChangeDeliveryExecutor.Preview() error = %v", err)
	}
	approval, err := database.NewChangeDeliveryApproval(preview)
	if err != nil {
		t.Fatalf("NewChangeDeliveryApproval() error = %v", err)
	}
	result, err := delivery.Execute(ctx, plan, approval)
	if err != nil {
		t.Fatalf("ChangeDeliveryExecutor.Execute() error = %v", err)
	}
	receipt, ok := result.Receipt()
	if !ok || receipt.DeliveryID() == "" {
		t.Fatalf("workset delivery did not return receipt after baseline persistence: %#v", result)
	}
	return receipt
}

func lookupPostgresWorksetBaseline(t *testing.T, ctx context.Context, store *database.FileChangeDeliveryBaselineStore, control database.ManagedTargetControlRecord) database.ChangeDeliveryBaseline {
	t.Helper()
	key, err := database.NewManagedTargetDeliveryLedgerKey(control)
	if err != nil {
		t.Fatal(err)
	}
	baseline, found, err := store.Lookup(ctx, key)
	if err != nil || !found {
		t.Fatalf("Lookup(workset baseline) = (%#v, %t, %v), want durable candidate", baseline, found, err)
	}
	return baseline
}

func readPostgresWorksetBaselineRows(t *testing.T, ctx context.Context, baseline database.ChangeDeliveryBaseline) []warehouse.Row {
	t.Helper()
	var rows []warehouse.Row
	if err := baseline.ReadCandidateBaseline(ctx, func(row warehouse.Row) error {
		rows = append(rows, row)
		return nil
	}); err != nil {
		t.Fatalf("ReadCandidateBaseline() error = %v", err)
	}
	return rows
}

// TestPostgresManagedTargetDriverLiveControlAssertions proves PostgreSQL's
// native identity and control-record observer against a server that it did not
// initialize. Every refusal compares durable state before and after the call,
// so a returned error alone can never satisfy this regression.
func TestPostgresManagedTargetDriverLiveControlAssertions(t *testing.T) {
	if os.Getenv(postgresCatalogIntegrationEnabledEnv) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the PostgreSQL Docker or Podman proof", postgresCatalogIntegrationEnabledEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	harness, err := newPostgresCatalogContainerHarness(
		dbtest.Runtime(strings.TrimSpace(os.Getenv(postgresCatalogIntegrationRuntimeEnv))),
		strings.TrimSpace(os.Getenv(postgresCatalogIntegrationEndpointEnv)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Error("PostgreSQL database test cleanup failed")
		}
	}()

	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatal("PostgreSQL database container did not start")
	}
	waitForPostgresCatalog(t, ctx, native.New(), postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema))
	source := openPostgresCatalogSource(t, ctx, endpoint)
	defer func() { _ = source.Close(ctx) }()

	driver, err := native.NewDatabaseDriver(source)
	if err != nil {
		t.Fatal("could not construct PostgreSQL managed target driver")
	}
	assertPostgresManagedTargetDurability(t, ctx, source, driver)

	fixture := newPostgresManagedTargetFixture(t, ctx, driver)
	assertPostgresManagedTargetMappingFence(t, ctx, driver, fixture)
	assertPostgresUnsupportedMappingRefusal(t, ctx, driver, fixture)
	assertPostgresManagedTargetMissingOwnerRefusal(t, ctx, source, driver, fixture)
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal("could not construct managed target provisioner")
	}
	control, err := provisioner.CreateOrAssert(ctx, fixture.plan)
	if err != nil {
		t.Fatalf("first PostgreSQL managed target create failed: %v", err)
	}
	assertPostgresManagedTargetCreated(t, ctx, source, fixture, control)

	beforeExact := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	control, err = provisioner.CreateOrAssert(ctx, fixture.plan)
	if err != nil {
		t.Fatalf("exact PostgreSQL managed target assertion failed: %v", err)
	}
	if control.NativeIdentity().Value != beforeExact.relationOID || control.Schema() != fixture.schema {
		t.Fatal("exact PostgreSQL managed target assertion did not return the observed control identity")
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, beforeExact)

	ledgerKey, err := database.NewManagedTargetDeliveryLedgerKey(control)
	if err != nil {
		t.Fatal("could not derive PostgreSQL managed target ledger key")
	}
	if _, found, err := driver.LoadManagedTargetDelivery(ctx, ledgerKey); err != nil || found {
		t.Fatal("new PostgreSQL managed target ledger was not observably empty")
	}
	receipt, err := database.NewManagedTargetDeliveryRecord("delivery-managed-target-1")
	if err != nil {
		t.Fatal("could not create PostgreSQL managed target delivery record")
	}
	if err := driver.StoreManagedTargetDelivery(ctx, ledgerKey, receipt); err != nil {
		t.Fatal("could not persist PostgreSQL managed target delivery record")
	}
	stored, found, err := driver.LoadManagedTargetDelivery(ctx, ledgerKey)
	if err != nil || !found || stored.DeliveryID() != receipt.DeliveryID() {
		t.Fatal("PostgreSQL managed target ledger did not return the committed delivery record")
	}
	afterLedger := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if afterLedger.ledgerRows != beforeExact.ledgerRows+1 {
		t.Fatal("PostgreSQL managed target ledger store did not create exactly one durable row")
	}
	assertPostgresManagedTargetWriteModes(t, ctx, source, driver, control, fixture)

	assertPostgresManagedTargetSchemaDriftRefusal(t, ctx, source, provisioner, fixture)
	assertPostgresManagedTargetForeignOwnerRefusal(t, ctx, source, provisioner, fixture)
	assertPostgresManagedTargetControlCollisionRefusal(t, ctx, source, provisioner, fixture)
	assertPostgresManagedTargetPermissionRefusal(t, ctx, endpoint, source, fixture)
	assertPostgresManagedTargetUnknownCommit(t, ctx, endpoint, driver, control, fixture)
	assertPostgresManagedTargetOIDReplacementRefusal(t, ctx, source, provisioner, fixture)
}

func assertPostgresManagedTargetMappingFence(t *testing.T, ctx context.Context, driver *native.DatabaseDriver, fixture postgresManagedTargetFixture) {
	t.Helper()
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal("could not construct PostgreSQL managed target provisioner for mapping fence")
	}
	if _, err := provisioner.CreateOrAssert(ctx, fixture.unmappedPlan); !errors.Is(err, database.ErrManagedTargetUnverifiable) {
		t.Fatalf("PostgreSQL incomplete mapping create error = %v, want ErrManagedTargetUnverifiable", err)
	}
	after, err := driver.ObserveManagedTarget(ctx, fixture.target)
	if err != nil {
		t.Fatal("could not observe PostgreSQL target after incomplete mapping refusal")
	}
	if after.NamespacePresent || after.RelationPresent || after.ControlState != database.ManagedTargetControlAbsent {
		t.Fatal("PostgreSQL incomplete mapping refusal created namespace, relation, or control state")
	}
}

func assertPostgresUnsupportedMappingRefusal(t *testing.T, ctx context.Context, driver *native.DatabaseDriver, fixture postgresManagedTargetFixture) {
	t.Helper()
	element, err := database.NewString(0, "")
	if err != nil {
		t.Fatal(err)
	}
	array, err := database.NewArray(element)
	if err != nil {
		t.Fatal(err)
	}
	typePlan, err := database.CompileTypePlan(array, array)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := database.NewMappingContractV1([]database.MappingColumnV1{{
		Source:   "source_array",
		Target:   "target_array",
		Type:     typePlan,
		Nullable: false,
	}})
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := warehouse.NewArtifactRef(fixture.owner.Identity(), "managed_target_unsupported_array")
	if err != nil {
		t.Fatal(err)
	}
	target, err := database.NewManagedTargetRef(fixture.owner, artifact, "managed-target-unsupported-array-stream")
	if err != nil {
		t.Fatal(err)
	}
	fingerprint := sha256.Sum256([]byte("postgres-managed-target-unsupported-array-v1"))
	schema, err := database.NewManagedTargetSchema(1, database.SchemaFingerprint(fingerprint))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := database.NewManagedTargetProvisioningPlan(fixture.owner, target, fixture.targetDatabase, schema, mapping)
	if err != nil {
		t.Fatal(err)
	}
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provisioner.CreateOrAssert(ctx, plan); !errors.Is(err, database.ErrManagedTargetUnverifiable) {
		t.Fatalf("PostgreSQL unsupported mapped DDL error = %v, want ErrManagedTargetUnverifiable", err)
	}
	after, err := driver.ObserveManagedTarget(ctx, target)
	if err != nil || after.NamespacePresent || after.RelationPresent || after.ControlState != database.ManagedTargetControlAbsent {
		t.Fatalf("PostgreSQL unsupported mapped DDL mutated target state: observation=%#v error=%v", after, err)
	}
}

func assertPostgresManagedTargetMissingOwnerRefusal(t *testing.T, ctx context.Context, source *pgx.Conn, driver *native.DatabaseDriver, fixture postgresManagedTargetFixture) {
	t.Helper()
	identity := fixture.owner.Identity()
	identity.ConnectionID = "managed-target-missing-owner-connection"
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := warehouse.NewArtifactRef(identity, "managed_target_missing_owner")
	if err != nil {
		t.Fatal(err)
	}
	target, err := database.NewManagedTargetRef(owner, artifact, "managed-target-missing-owner-stream")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := source.Exec(ctx, "CREATE SCHEMA "+postgresManagedTargetQualifiedName(target.Namespace())); err != nil {
		t.Fatal("could not create independently unowned PostgreSQL namespace")
	}
	fingerprint := sha256.Sum256([]byte("postgres-managed-target-missing-owner-v1"))
	schema, err := database.NewManagedTargetSchema(1, database.SchemaFingerprint(fingerprint))
	if err != nil {
		t.Fatal(err)
	}
	mapping, ok := fixture.plan.Mapping()
	if !ok {
		t.Fatal("fixture lost shared mapping for missing-owner refusal")
	}
	plan, err := database.NewManagedTargetProvisioningPlan(owner, target, fixture.targetDatabase, schema, mapping)
	if err != nil {
		t.Fatal(err)
	}
	before, err := driver.ObserveManagedTarget(ctx, target)
	if err != nil || !before.NamespacePresent || before.NamespaceOwnerState != database.ManagedTargetNamespaceOwnerAbsent || before.RelationPresent || before.ControlState != database.ManagedTargetControlAbsent {
		t.Fatalf("PostgreSQL independent missing-owner fixture = (%#v, %v), want unowned namespace with no target", before, err)
	}
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provisioner.CreateOrAssert(ctx, plan); !errors.Is(err, database.ErrManagedTargetNamespaceOwnerMissing) {
		t.Fatalf("PostgreSQL missing owner error = %v, want ErrManagedTargetNamespaceOwnerMissing", err)
	}
	after, err := driver.ObserveManagedTarget(ctx, target)
	if err != nil || after != before {
		t.Fatalf("PostgreSQL missing owner refusal mutated durable state: got=(%#v, %v) want=%#v", after, err, before)
	}
}

type postgresManagedTargetFixture struct {
	owner          database.TargetOwner
	target         database.ManagedTargetRef
	targetDatabase database.TargetDatabaseIdentity
	schema         database.ManagedTargetSchema
	plan           database.ManagedTargetProvisioningPlan
	unmappedPlan   database.ManagedTargetProvisioningPlan
}

func newPostgresManagedTargetFixture(t *testing.T, ctx context.Context, driver *native.DatabaseDriver) postgresManagedTargetFixture {
	t.Helper()
	identity := database.ConnectionIdentity{
		WorkspaceID:  "managed-target-workspace",
		ConnectorID:  "postgres",
		ConnectionID: "managed-target-connection",
	}
	artifact, err := warehouse.NewArtifactRef(identity, "managed_target_orders")
	if err != nil {
		t.Fatal("could not create managed target source artifact")
	}
	owner, err := database.NewTargetOwner(identity)
	if err != nil {
		t.Fatal("could not create managed target owner")
	}
	target, err := database.NewManagedTargetRef(owner, artifact, "managed-target-stream")
	if err != nil {
		t.Fatal("could not create managed target reference")
	}
	absent, err := driver.ObserveManagedTarget(ctx, target)
	if err != nil {
		t.Fatal("could not observe absent PostgreSQL managed target")
	}
	if absent.NamespacePresent || absent.RelationPresent || absent.TargetDatabase.Value() == "" {
		t.Fatal("PostgreSQL absent managed target observation did not expose only the live database identity")
	}
	fingerprint := sha256.Sum256([]byte("postgres-managed-target-live-schema-v1"))
	schema, err := database.NewManagedTargetSchema(1, database.SchemaFingerprint(fingerprint))
	if err != nil {
		t.Fatal("could not create managed target schema assertion")
	}
	mapping := postgresManagedTargetMapping(t)
	plan, err := database.NewManagedTargetProvisioningPlan(owner, target, absent.TargetDatabase, schema, mapping)
	if err != nil {
		t.Fatal("could not create managed target provisioning plan")
	}
	unmappedPlan, err := database.NewManagedTargetProvisioningPlan(owner, target, absent.TargetDatabase, schema)
	if err != nil {
		t.Fatal("could not create unmapped managed target provisioning plan")
	}
	return postgresManagedTargetFixture{
		owner:          owner,
		target:         target,
		targetDatabase: absent.TargetDatabase,
		schema:         schema,
		plan:           plan,
		unmappedPlan:   unmappedPlan,
	}
}

func postgresManagedTargetMapping(t *testing.T) database.MappingContractV1 {
	t.Helper()
	integer, err := database.NewSignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	integerPlan, err := database.CompileTypePlan(integer, integer)
	if err != nil {
		t.Fatal(err)
	}
	text, err := database.NewString(0, "")
	if err != nil {
		t.Fatal(err)
	}
	textPlan, err := database.CompileTypePlan(text, text)
	if err != nil {
		t.Fatal(err)
	}
	mapping, err := database.NewMappingContractV1([]database.MappingColumnV1{
		{
			Source:   "source_tenant",
			Target:   "tenant",
			Type:     textPlan,
			Nullable: false,
		},
		{
			Source:   "source_id",
			Target:   "id",
			Type:     integerPlan,
			Nullable: false,
		},
		{
			Source:   "source_value",
			Target:   "value",
			Type:     textPlan,
			Nullable: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return mapping
}

func assertPostgresManagedTargetCreated(t *testing.T, ctx context.Context, source *pgx.Conn, fixture postgresManagedTargetFixture, control database.ManagedTargetControlRecord) {
	t.Helper()
	state := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if state.namespaceOID == "" || state.relationOID == "" || state.ownerRows != 1 || state.controlRows != 1 || state.ledgerRows != 0 {
		t.Fatalf("PostgreSQL first create did not persist exactly one namespace/control target: %+v", state)
	}
	if control.NativeIdentity().Value != state.relationOID || control.Schema() != fixture.schema || state.ownerConnectionID != fixture.owner.Identity().ConnectionID || state.streamID != fixture.target.StreamID() {
		t.Fatalf("PostgreSQL first create control does not match durable target state: control=%#v state=%+v", control, state)
	}
	rows, err := source.Query(ctx, `
		SELECT a.attname, pg_catalog.format_type(a.atttypid, a.atttypmod), a.attnotnull
		FROM pg_catalog.pg_attribute AS a
		JOIN pg_catalog.pg_class AS c ON c.oid = a.attrelid
		JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2 AND a.attnum > 0 AND NOT a.attisdropped
		ORDER BY a.attnum`, fixture.target.Namespace(), fixture.target.Relation())
	if err != nil {
		t.Fatalf("could not inspect PostgreSQL first-create target columns: %v", err)
	}
	defer rows.Close()
	type targetColumn struct {
		name    string
		typeSQL string
		notNull bool
	}
	got := make([]targetColumn, 0, 3)
	for rows.Next() {
		var column targetColumn
		if err := rows.Scan(&column.name, &column.typeSQL, &column.notNull); err != nil {
			t.Fatal("could not scan PostgreSQL first-create target column")
		}
		got = append(got, column)
	}
	want := []targetColumn{
		{name: "tenant", typeSQL: "text", notNull: true},
		{name: "id", typeSQL: "bigint", notNull: true},
		{name: "value", typeSQL: "text", notNull: false},
		{name: synccontract.HistoryValidFromColumn, typeSQL: "timestamp with time zone", notNull: false},
		{name: synccontract.HistoryValidToColumn, typeSQL: "timestamp with time zone", notNull: false},
		{name: synccontract.HistoryIsCurrentColumn, typeSQL: "boolean", notNull: false},
	}
	if err := rows.Err(); err != nil || len(got) != len(want) {
		t.Fatalf("PostgreSQL first-create target columns = %#v (%v), want %#v", got, err, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("PostgreSQL first-create target column[%d] = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func assertPostgresManagedTargetWriteModes(t *testing.T, ctx context.Context, source *pgx.Conn, driver *native.DatabaseDriver, control database.ManagedTargetControlRecord, fixture postgresManagedTargetFixture) {
	t.Helper()
	definition := postgresManagedTargetWriteDefinition(t)
	ledger, err := database.NewManagedTargetDeliveryLedger(driver)
	if err != nil {
		t.Fatal("could not construct PostgreSQL managed target delivery ledger")
	}
	executor, err := database.NewDatabaseWriteExecutor(driver, ledger)
	if err != nil {
		t.Fatal("could not construct PostgreSQL mapped write executor")
	}

	fullAppend := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 2, 0, false)
	fullAppendReceipt := postgresExecuteManagedTargetWrite(t, ctx, executor, fullAppend, []connectors.Record{
		{"source_tenant": "tenant-a", "source_id": int64(1), "source_value": "first"},
		{"source_tenant": "tenant-b", "source_id": int64(1), "source_value": "second"},
	}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetReceiptPersisted(t, ctx, driver, control, fullAppendReceipt)
	assertPostgresManagedTargetRows(t, ctx, source, control, []postgresManagedTargetRow{
		{tenant: "tenant-a", id: 1, value: "first"},
		{tenant: "tenant-b", id: 1, value: "second"},
	})

	incrementalAppend := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalAppend, connectors.ApplyStrategyAppend, nil, 1, 0, false)
	postgresExecuteManagedTargetWrite(t, ctx, executor, incrementalAppend, []connectors.Record{{"source_tenant": "tenant-a", "source_id": int64(2), "source_value": "append"}}, database.TombstoneEnvelope{})

	dedupe := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupe, connectors.ApplyStrategyDedupe, []string{"tenant", "id"}, 2, 0, false)
	postgresExecuteManagedTargetWrite(t, ctx, executor, dedupe, []connectors.Record{
		{"source_tenant": "tenant-a", "source_id": int64(1), "source_value": "must-not-replace"},
		{"source_tenant": "tenant-a", "source_id": int64(3), "source_value": "deduped-new"},
	}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetRows(t, ctx, source, control, []postgresManagedTargetRow{
		{tenant: "tenant-a", id: 1, value: "first"},
		{tenant: "tenant-a", id: 2, value: "append"},
		{tenant: "tenant-a", id: 3, value: "deduped-new"},
		{tenant: "tenant-b", id: 1, value: "second"},
	})

	upsert := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalUpsert, connectors.ApplyStrategyMerge, []string{"tenant", "id"}, 2, 0, false)
	postgresExecuteManagedTargetWrite(t, ctx, executor, upsert, []connectors.Record{
		{"source_tenant": "tenant-a", "source_id": int64(1), "source_value": "upserted"},
		{"source_tenant": "tenant-b", "source_id": int64(1), "source_value": "upserted-second"},
	}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetRows(t, ctx, source, control, []postgresManagedTargetRow{
		{tenant: "tenant-a", id: 1, value: "upserted"},
		{tenant: "tenant-a", id: 2, value: "append"},
		{tenant: "tenant-a", id: 3, value: "deduped-new"},
		{tenant: "tenant-b", id: 1, value: "upserted-second"},
	})

	// An ordinary keyed batch updates only records it contains. The absent
	// tenant-b row is observed after commit to prove physical absence is not a
	// hidden delete signal.
	absentDoesNotDelete := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalUpsert, connectors.ApplyStrategyMerge, []string{"tenant", "id"}, 1, 0, false)
	postgresExecuteManagedTargetWrite(t, ctx, executor, absentDoesNotDelete, []connectors.Record{{"source_tenant": "tenant-a", "source_id": int64(1), "source_value": "upserted-again"}}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetRows(t, ctx, source, control, []postgresManagedTargetRow{
		{tenant: "tenant-a", id: 1, value: "upserted-again"},
		{tenant: "tenant-a", id: 2, value: "append"},
		{tenant: "tenant-a", id: 3, value: "deduped-new"},
		{tenant: "tenant-b", id: 1, value: "upserted-second"},
	})

	tombstone := synccontract.Tombstone{
		Operation:   synccontract.OperationDelete,
		EventID:     synccontract.OpaqueToken("postgres-live-delete-tenant-b-1"),
		Key:         json.RawMessage(`{"tenant":"tenant-b","id":1}`),
		DeleteImage: synccontract.DeleteImageKeyOnly,
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken("00000001"),
			TieBreaker: synccontract.OpaqueToken("00000001"),
		},
	}
	envelope, err := database.NewTombstoneEnvelope([]synccontract.Tombstone{tombstone})
	if err != nil {
		t.Fatal("could not construct explicit PostgreSQL tombstone envelope")
	}
	explicitDelete := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalUpsert, connectors.ApplyStrategyMerge, []string{"tenant", "id"}, 0, 1, false)
	postgresExecuteManagedTargetWrite(t, ctx, executor, explicitDelete, nil, envelope)
	assertPostgresManagedTargetRows(t, ctx, source, control, []postgresManagedTargetRow{
		{tenant: "tenant-a", id: 1, value: "upserted-again"},
		{tenant: "tenant-a", id: 2, value: "append"},
		{tenant: "tenant-a", id: 3, value: "deduped-new"},
	})

	assertPostgresManagedTargetOrderFenceAndHistory(t, ctx, source, executor, definition, control, fixture)

	assertPostgresManagedTargetStatementFailureRollback(t, ctx, source, driver, executor, definition, control, fixture)
	assertPostgresManagedTargetCancellationRollback(t, ctx, source, driver, definition, control, fixture)

	beforeFailedOverwrite := postgresManagedTargetRows(t, ctx, source, control)
	failedOverwrite := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeFullOverwrite, connectors.ApplyStrategyReplace, nil, 1, 0, true)
	badInput, err := database.NewDatabaseWriteInput([]connectors.Record{{"source_tenant": "tenant-c", "source_id": int32(9), "source_value": "wrong-logical-type"}}, database.TombstoneEnvelope{})
	if err != nil {
		t.Fatal("could not construct PostgreSQL failed-overwrite input")
	}
	preview, err := executor.Preview(ctx, failedOverwrite)
	if err != nil {
		t.Fatal("could not preview PostgreSQL failed-overwrite plan")
	}
	approval, err := database.NewDatabaseWriteApproval(preview)
	if err != nil {
		t.Fatal("could not approve PostgreSQL failed-overwrite plan")
	}
	if _, err := executor.ExecuteInput(ctx, failedOverwrite, approval, badInput); !errors.Is(err, database.ErrDatabaseWriteBatchFailed) {
		t.Fatalf("PostgreSQL failed atomic overwrite error = %v, want ErrDatabaseWriteBatchFailed", err)
	}
	if got := postgresManagedTargetRows(t, ctx, source, control); !samePostgresManagedTargetRows(got, beforeFailedOverwrite) {
		t.Fatalf("PostgreSQL failed atomic overwrite changed durable rows: got=%#v want=%#v", got, beforeFailedOverwrite)
	}

	fullOverwrite := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeFullOverwrite, connectors.ApplyStrategyReplace, nil, 1, 0, true)
	fullOverwriteReceipt := postgresExecuteManagedTargetWrite(t, ctx, executor, fullOverwrite, []connectors.Record{{"source_tenant": "tenant-final", "source_id": int64(9), "source_value": "replaced"}}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetReceiptPersisted(t, ctx, driver, control, fullOverwriteReceipt)
	assertPostgresManagedTargetRows(t, ctx, source, control, []postgresManagedTargetRow{{tenant: "tenant-final", id: 9, value: "replaced"}})
}

func assertPostgresManagedTargetStatementFailureRollback(t *testing.T, ctx context.Context, source *pgx.Conn, driver *native.DatabaseDriver, executor *database.DatabaseWriteExecutor, definition database.Definition, control database.ManagedTargetControlRecord, fixture postgresManagedTargetFixture) {
	t.Helper()
	qualified := postgresManagedTargetQualifiedName(control.Target().Namespace(), control.Target().Relation())
	const constraint = "pmmt_test_statement_refusal"
	if _, err := source.Exec(ctx, "ALTER TABLE "+qualified+" ADD CONSTRAINT "+postgresManagedTargetQualifiedName(constraint)+" CHECK (value <> 'statement-refusal')"); err != nil {
		t.Fatal("could not inject PostgreSQL mapped statement failure")
	}
	defer func() {
		if _, err := source.Exec(context.WithoutCancel(ctx), "ALTER TABLE "+qualified+" DROP CONSTRAINT "+postgresManagedTargetQualifiedName(constraint)); err != nil {
			t.Error("could not remove PostgreSQL mapped statement-failure constraint")
		}
	}()
	before := postgresManagedTargetRows(t, ctx, source, control)
	beforeDeliveryID := postgresManagedTargetDeliveryID(t, ctx, driver, control)
	plan := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeFullOverwrite, connectors.ApplyStrategyReplace, nil, 1, 0, true)
	input, err := database.NewDatabaseWriteInput([]connectors.Record{{"source_tenant": "tenant-statement", "source_id": int64(8), "source_value": "statement-refusal"}}, database.TombstoneEnvelope{})
	if err != nil {
		t.Fatal("could not construct PostgreSQL statement-failure input")
	}
	preview, err := executor.Preview(ctx, plan)
	if err != nil {
		t.Fatal("could not preview PostgreSQL statement-failure plan")
	}
	approval, err := database.NewDatabaseWriteApproval(preview)
	if err != nil {
		t.Fatal("could not approve PostgreSQL statement-failure plan")
	}
	if _, err := executor.ExecuteInput(ctx, plan, approval, input); !errors.Is(err, database.ErrDatabaseWriteBatchFailed) {
		t.Fatalf("PostgreSQL statement failure error = %v, want ErrDatabaseWriteBatchFailed", err)
	}
	if got := postgresManagedTargetRows(t, ctx, source, control); !samePostgresManagedTargetRows(got, before) {
		t.Fatalf("PostgreSQL statement failure changed durable rows: got=%#v want=%#v", got, before)
	}
	if got := postgresManagedTargetDeliveryID(t, ctx, driver, control); got != beforeDeliveryID {
		t.Fatalf("PostgreSQL statement failure changed durable delivery evidence: got=%q want=%q", got, beforeDeliveryID)
	}
}

// assertPostgresManagedTargetOrderFenceAndHistory proves live PostgreSQL
// state for the source-ordered strategies. Each stale replay is deliberately
// applied after a newer durable event; a no-op implementation would leave the
// asserted rows, tombstones, and validity windows in different states.
func assertPostgresManagedTargetOrderFenceAndHistory(t *testing.T, ctx context.Context, source *pgx.Conn, executor *database.DatabaseWriteExecutor, definition database.Definition, control database.ManagedTargetControlRecord, fixture postgresManagedTargetFixture) {
	t.Helper()
	keys := []string{"tenant", "id"}
	polling := newPostgresManagedTargetPollingApply(t, executor, definition, control, fixture)
	refusalResolved, err := engine.PollingPreflight(ctx, polling.registry, polling.declaration, polling.object, synccontract.ModeIncrementalUpsert)
	if err != nil {
		t.Fatalf("could not preflight oversized PostgreSQL polling page: %v", err)
	}
	beforeRefusal := postgresManagedTargetRows(t, ctx, source, control)
	oversized := engine.PollingApplyPage{Records: []engine.PollingApplyRecord{
		{Record: connectors.Record{"source_tenant": "refusal", "source_id": int64(1), "source_value": "one"}, Position: postgresManagedTargetPosition(10)},
		{Record: connectors.Record{"source_tenant": "refusal", "source_id": int64(2), "source_value": "two"}, Position: postgresManagedTargetPosition(11)},
		{Record: connectors.Record{"source_tenant": "refusal", "source_id": int64(3), "source_value": "three"}, Position: postgresManagedTargetPosition(12)},
	}}
	if _, err := engine.ApplyPollingPage(ctx, refusalResolved, oversized); err == nil {
		t.Fatal("oversized PostgreSQL polling page applied, want pre-mutation refusal")
	}
	if got := postgresManagedTargetRows(t, ctx, source, control); !samePostgresManagedTargetRows(got, beforeRefusal) {
		t.Fatalf("oversized PostgreSQL polling page changed target rows: got=%#v want=%#v", got, beforeRefusal)
	}

	postgresApplyManagedTargetPollingPage(t, ctx, polling, synccontract.ModeIncrementalUpsert, []database.OrderedRecord{{
		Record:   connectors.Record{"source_tenant": "fenced-upsert", "source_id": int64(1), "source_value": "newest"},
		Position: postgresManagedTargetPosition(20),
	}}, database.TombstoneEnvelope{})
	// A physically absent unrelated key is not a deletion instruction.
	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalUpsert, connectors.ApplyStrategyMerge, keys, 1, 0, false, true), []database.OrderedRecord{{
		Record:   connectors.Record{"source_tenant": "fenced-unrelated", "source_id": int64(2), "source_value": "other"},
		Position: postgresManagedTargetPosition(21),
	}}, database.TombstoneEnvelope{})
	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalUpsert, connectors.ApplyStrategyMerge, keys, 1, 0, false, true), []database.OrderedRecord{{
		Record:   connectors.Record{"source_tenant": "fenced-upsert", "source_id": int64(1), "source_value": "stale"},
		Position: postgresManagedTargetPosition(19),
	}}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetValue(t, ctx, source, control, "fenced-upsert", 1, "newest", 1)

	upsertDelete := postgresManagedTargetTombstone(t, "fenced-upsert", 1, 22)
	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalUpsert, connectors.ApplyStrategyMerge, keys, 0, 1, false, true), nil, postgresManagedTargetTombstoneEnvelope(t, upsertDelete))
	assertPostgresManagedTargetValue(t, ctx, source, control, "fenced-upsert", 1, "", 0)
	// The older record cannot resurrect a physically deleted row because the
	// explicit tombstone advanced the persisted source-order fence.
	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalUpsert, connectors.ApplyStrategyMerge, keys, 1, 0, false, true), []database.OrderedRecord{{
		Record:   connectors.Record{"source_tenant": "fenced-upsert", "source_id": int64(1), "source_value": "late-replay"},
		Position: postgresManagedTargetPosition(20),
	}}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetValue(t, ctx, source, control, "fenced-upsert", 1, "", 0)

	dedupe := postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupe, connectors.ApplyStrategyDedupe, keys, 1, 0, false, true)
	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, dedupe, []database.OrderedRecord{{
		Record:   connectors.Record{"source_tenant": "fenced-dedupe", "source_id": int64(1), "source_value": "first"},
		Position: postgresManagedTargetPosition(30),
	}}, database.TombstoneEnvelope{})
	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupe, connectors.ApplyStrategyDedupe, keys, 1, 0, false, true), []database.OrderedRecord{{
		Record:   connectors.Record{"source_tenant": "fenced-dedupe", "source_id": int64(1), "source_value": "duplicate"},
		Position: postgresManagedTargetPosition(31),
	}}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetValue(t, ctx, source, control, "fenced-dedupe", 1, "first", 1)
	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupe, connectors.ApplyStrategyDedupe, keys, 0, 1, false, true), nil, postgresManagedTargetTombstoneEnvelope(t, postgresManagedTargetTombstone(t, "fenced-dedupe", 1, 32)))
	assertPostgresManagedTargetValue(t, ctx, source, control, "fenced-dedupe", 1, "", 0)

	postgresApplyManagedTargetPollingPage(t, ctx, polling, synccontract.ModeIncrementalDedupeHistory, []database.OrderedRecord{{
		Record:   connectors.Record{"source_tenant": "history", "source_id": int64(1), "source_value": "first"},
		Position: postgresManagedTargetPosition(40),
	}}, database.TombstoneEnvelope{})
	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupeHistory, connectors.ApplyStrategyDedupeHistory, keys, 1, 0, false, true), []database.OrderedRecord{{
		Record:   connectors.Record{"source_tenant": "history", "source_id": int64(1), "source_value": "second"},
		Position: postgresManagedTargetPosition(41),
	}}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetHistory(t, ctx, source, control, "history", 1, []postgresManagedTargetHistoryRow{{value: "first", open: false, current: false}, {value: "second", open: true, current: true}})
	// A later page omits history/1 entirely. Both retained versions prove that
	// omission is not a close or physical-delete operation.
	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupeHistory, connectors.ApplyStrategyDedupeHistory, keys, 1, 0, false, true), []database.OrderedRecord{{
		Record:   connectors.Record{"source_tenant": "history-other", "source_id": int64(2), "source_value": "other"},
		Position: postgresManagedTargetPosition(42),
	}}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetHistory(t, ctx, source, control, "history", 1, []postgresManagedTargetHistoryRow{{value: "first", open: false, current: false}, {value: "second", open: true, current: true}})

	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupeHistory, connectors.ApplyStrategyDedupeHistory, keys, 0, 1, false, true), nil, postgresManagedTargetTombstoneEnvelope(t, postgresManagedTargetTombstone(t, "history", 1, 43)))
	assertPostgresManagedTargetHistory(t, ctx, source, control, "history", 1, []postgresManagedTargetHistoryRow{{value: "first", open: false, current: false}, {value: "second", open: false, current: false}})
	// A stale record after the tombstone cannot create another current history
	// row, so the explicit close remains the observable target state.
	postgresExecuteManagedTargetOrderedWrite(t, ctx, executor, postgresManagedTargetFencedWritePlan(t, definition, control, fixture, synccontract.ModeIncrementalDedupeHistory, connectors.ApplyStrategyDedupeHistory, keys, 1, 0, false, true), []database.OrderedRecord{{
		Record:   connectors.Record{"source_tenant": "history", "source_id": int64(1), "source_value": "late-replay"},
		Position: postgresManagedTargetPosition(41),
	}}, database.TombstoneEnvelope{})
	assertPostgresManagedTargetHistory(t, ctx, source, control, "history", 1, []postgresManagedTargetHistoryRow{{value: "first", open: false, current: false}, {value: "second", open: false, current: false}})
}

func assertPostgresManagedTargetCancellationRollback(t *testing.T, ctx context.Context, source *pgx.Conn, driver *native.DatabaseDriver, definition database.Definition, control database.ManagedTargetControlRecord, fixture postgresManagedTargetFixture) {
	t.Helper()
	ledger, err := database.NewManagedTargetDeliveryLedger(driver)
	if err != nil {
		t.Fatal("could not construct PostgreSQL cancellation ledger")
	}
	cancelledCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	wrapped := &postgresAfterApplyDriver{driver: driver, afterApply: cancel}
	executor, err := database.NewDatabaseWriteExecutor(wrapped, ledger)
	if err != nil {
		t.Fatal("could not construct PostgreSQL cancellation executor")
	}
	before := postgresManagedTargetRows(t, ctx, source, control)
	beforeDeliveryID := postgresManagedTargetDeliveryID(t, ctx, driver, control)
	plan := postgresManagedTargetWritePlan(t, definition, control, fixture, synccontract.ModeFullOverwrite, connectors.ApplyStrategyReplace, nil, 1, 0, true)
	input, err := database.NewDatabaseWriteInput([]connectors.Record{{"source_tenant": "tenant-cancel", "source_id": int64(8), "source_value": "cancelled"}}, database.TombstoneEnvelope{})
	if err != nil {
		t.Fatal("could not construct PostgreSQL cancellation input")
	}
	preview, err := executor.Preview(ctx, plan)
	if err != nil {
		t.Fatal("could not preview PostgreSQL cancellation plan")
	}
	approval, err := database.NewDatabaseWriteApproval(preview)
	if err != nil {
		t.Fatal("could not approve PostgreSQL cancellation plan")
	}
	if _, err := executor.ExecuteInput(cancelledCtx, plan, approval, input); !errors.Is(err, context.Canceled) {
		t.Fatalf("PostgreSQL cancellation error = %v, want context.Canceled", err)
	}
	if wrapped.beginCalls != 1 {
		t.Fatalf("PostgreSQL cancellation begin calls = %d, want one pinned session", wrapped.beginCalls)
	}
	if got := postgresManagedTargetRows(t, ctx, source, control); !samePostgresManagedTargetRows(got, before) {
		t.Fatalf("PostgreSQL cancelled write changed durable rows: got=%#v want=%#v", got, before)
	}
	if got := postgresManagedTargetDeliveryID(t, ctx, driver, control); got != beforeDeliveryID {
		t.Fatalf("PostgreSQL cancelled write changed durable delivery evidence: got=%q want=%q", got, beforeDeliveryID)
	}
}

func assertPostgresManagedTargetUnknownCommit(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint, ledgerDriver *native.DatabaseDriver, control database.ManagedTargetControlRecord, fixture postgresManagedTargetFixture) {
	t.Helper()
	beforeDeliveryID := postgresManagedTargetDeliveryID(t, ctx, ledgerDriver, control)
	writeConn := openPostgresCatalogSource(t, ctx, endpoint)
	defer func() { _ = writeConn.Close(context.WithoutCancel(ctx)) }()
	driver, err := native.NewDatabaseDriver(writeConn)
	if err != nil {
		t.Fatal("could not construct PostgreSQL disconnect-injection driver")
	}
	ledger, err := database.NewManagedTargetDeliveryLedger(ledgerDriver)
	if err != nil {
		t.Fatal("could not construct PostgreSQL unknown-commit ledger")
	}
	wrapped := &postgresAfterApplyDriver{driver: driver, afterApply: func() {
		_ = writeConn.Close(context.WithoutCancel(ctx))
	}}
	executor, err := database.NewDatabaseWriteExecutor(wrapped, ledger)
	if err != nil {
		t.Fatal("could not construct PostgreSQL disconnect-injection executor")
	}
	plan := postgresManagedTargetWritePlan(t, postgresManagedTargetWriteDefinition(t), control, fixture, synccontract.ModeFullAppend, connectors.ApplyStrategyAppend, nil, 1, 0, false)
	input, err := database.NewDatabaseWriteInput([]connectors.Record{{"source_tenant": "tenant-unknown", "source_id": int64(99), "source_value": "commit-outcome-unknown"}}, database.TombstoneEnvelope{})
	if err != nil {
		t.Fatal("could not construct PostgreSQL unknown-commit input")
	}
	preview, err := executor.Preview(ctx, plan)
	if err != nil {
		t.Fatal("could not preview PostgreSQL unknown-commit plan")
	}
	approval, err := database.NewDatabaseWriteApproval(preview)
	if err != nil {
		t.Fatal("could not approve PostgreSQL unknown-commit plan")
	}
	result, err := executor.ExecuteInput(ctx, plan, approval, input)
	if !errors.Is(err, database.ErrDatabaseWriteCommitOutcomeUnknown) || result.Outcome() != database.CommitOutcomeUnknown {
		t.Fatalf("PostgreSQL disconnect commit outcome = (%s, %v), want unknown outcome without acknowledgement", result.Outcome(), err)
	}
	if wrapped.beginCalls != 1 || !writeConn.IsClosed() {
		t.Fatalf("PostgreSQL unknown commit retry/disconnect evidence = begin_calls=%d closed=%t, want one session and a closed real connection", wrapped.beginCalls, writeConn.IsClosed())
	}
	if got := postgresManagedTargetDeliveryID(t, ctx, ledgerDriver, control); got != beforeDeliveryID {
		t.Fatalf("PostgreSQL unknown commit changed durable delivery evidence: got=%q want=%q", got, beforeDeliveryID)
	}
}

func postgresManagedTargetWriteDefinition(t *testing.T) database.Definition {
	t.Helper()
	bundle, err := engine.Load(defs.FS, "postgres")
	if err != nil || bundle.Database == nil {
		t.Fatal("could not load PostgreSQL admitted write-mode definition")
	}
	return *bundle.Database
}

func postgresManagedTargetWritePlan(t *testing.T, definition database.Definition, control database.ManagedTargetControlRecord, fixture postgresManagedTargetFixture, mode synccontract.Mode, strategy connectors.ApplyStrategy, keys []string, records, tombstones int, destructive bool) database.DatabaseWritePlan {
	t.Helper()
	return postgresManagedTargetFencedWritePlan(t, definition, control, fixture, mode, strategy, keys, records, tombstones, destructive, false)
}

func postgresManagedTargetFencedWritePlan(t *testing.T, definition database.Definition, control database.ManagedTargetControlRecord, fixture postgresManagedTargetFixture, mode synccontract.Mode, strategy connectors.ApplyStrategy, keys []string, records, tombstones int, destructive, conditionalOrderFence bool) database.DatabaseWritePlan {
	t.Helper()
	mapping, ok := fixture.plan.Mapping()
	if !ok {
		t.Fatal("mapped PostgreSQL target plan lost its shared mapping")
	}
	request := database.DatabaseWritePlanRequest{
		Definition:            definition,
		Control:               control,
		Mode:                  mode,
		Strategy:              strategy,
		Mapping:               mapping,
		Keys:                  keys,
		RecordCount:           records,
		TombstoneCount:        tombstones,
		BatchSize:             2,
		Destructive:           destructive,
		ConditionalOrderFence: conditionalOrderFence,
	}
	if mode == synccontract.ModeIncrementalDedupeHistory {
		request.HistoryRoute = database.DatabaseWriteHistoryRoute{Source: definition.Driver(), Destination: definition.Driver()}
	}
	plan, err := database.NewDatabaseWritePlan(context.Background(), request)
	if err != nil {
		t.Fatalf("could not create PostgreSQL %s mapped write plan: %v", mode, err)
	}
	return plan
}

func postgresExecuteManagedTargetWrite(t *testing.T, ctx context.Context, executor *database.DatabaseWriteExecutor, plan database.DatabaseWritePlan, records []connectors.Record, tombstones database.TombstoneEnvelope) database.DeliveryReceiptV1 {
	t.Helper()
	input, err := database.NewDatabaseWriteInput(records, tombstones)
	if err != nil {
		t.Fatal("could not construct bounded PostgreSQL write input")
	}
	return postgresExecuteManagedTargetInput(t, ctx, executor, plan, input)
}

func postgresExecuteManagedTargetOrderedWrite(t *testing.T, ctx context.Context, executor *database.DatabaseWriteExecutor, plan database.DatabaseWritePlan, records []database.OrderedRecord, tombstones database.TombstoneEnvelope) database.DeliveryReceiptV1 {
	t.Helper()
	input, err := database.NewOrderedDatabaseWriteInput(records, tombstones)
	if err != nil {
		t.Fatal("could not construct source-ordered PostgreSQL write input")
	}
	return postgresExecuteManagedTargetInput(t, ctx, executor, plan, input)
}

func postgresExecuteManagedTargetInput(t *testing.T, ctx context.Context, executor *database.DatabaseWriteExecutor, plan database.DatabaseWritePlan, input database.DatabaseWriteInput) database.DeliveryReceiptV1 {
	t.Helper()
	preview, err := executor.Preview(ctx, plan)
	if err != nil {
		t.Fatalf("PostgreSQL %s preview failed: %v", plan.Mode(), err)
	}
	approval, err := database.NewDatabaseWriteApproval(preview)
	if err != nil {
		t.Fatalf("PostgreSQL %s approval failed: %v", plan.Mode(), err)
	}
	result, err := executor.ExecuteInput(ctx, plan, approval, input)
	if err != nil {
		t.Fatalf("PostgreSQL %s execution failed: %v", plan.Mode(), err)
	}
	receipt, ok := result.Receipt()
	if !ok || receipt.DeliveryID() == "" || receipt.CommittedAt().IsZero() {
		t.Fatalf("PostgreSQL %s did not return durable target receipt: %#v", plan.Mode(), result)
	}
	return receipt
}

type postgresManagedTargetPollingApply struct {
	declaration *connectors.PollingWatermarkDescriptor
	object      connectors.PollingCatalogObject
	registry    *engine.PollingPreflightRegistry
	apply       *engine.DatabasePollingApplyExecutor
}

func newPostgresManagedTargetPollingApply(t *testing.T, write *database.DatabaseWriteExecutor, definition database.Definition, control database.ManagedTargetControlRecord, fixture postgresManagedTargetFixture) postgresManagedTargetPollingApply {
	t.Helper()
	mapping, ok := fixture.plan.Mapping()
	if !ok {
		t.Fatal("mapped PostgreSQL target plan lost its polling mapping")
	}
	sourceReference := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target_polling_source"}
	targetReference := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target_polling_target"}
	apply, err := engine.NewDatabasePollingApplyExecutor(engine.DatabasePollingApplyConfig{
		Reference:  targetReference,
		Evidence:   engine.RequiredPollingWatermarkConformanceEvidence(),
		Write:      write,
		Definition: definition,
		Control:    control,
		Mapping:    mapping,
		BatchSize:  2,
	})
	if err != nil {
		t.Fatalf("could not construct PostgreSQL native polling target: %v", err)
	}
	registry := engine.NewPollingPreflightRegistry()
	if err := registry.RegisterSource(&postgresManagedTargetPollingSource{reference: sourceReference, evidence: engine.RequiredPollingWatermarkConformanceEvidence()}); err != nil {
		t.Fatalf("could not register PostgreSQL polling source fixture: %v", err)
	}
	if err := registry.RegisterApply(apply); err != nil {
		t.Fatalf("could not register PostgreSQL native polling target: %v", err)
	}
	return postgresManagedTargetPollingApply{
		declaration: &connectors.PollingWatermarkDescriptor{
			Status: connectors.PollingWatermarkStatusImplemented,
			Source: connectors.PollingWatermarkSourceDescriptor{
				Executor: sourceReference,
				Object:   connectors.PollingCatalogObjectSelector{Kind: connectors.PollingCatalogObjectRelation},
				Read: connectors.PollingReadProtocol{
					Kind: connectors.PollingReadProtocolKeyset, MaxPageSize: 2, MaxPages: 2, MaxRequests: 2,
					StableTraversal: true, Predicate: connectors.PollingKeysetPredicateLexicographic,
				},
				Snapshot:         connectors.PollingSnapshotBarrier{Kind: connectors.PollingSnapshotBarrierTransaction},
				Cursor:           connectors.PollingCursor{Codec: connectors.PollingCursorCodecRFC3339Nano, Type: connectors.PollingCursorTypeTimestamp, Precision: "nanosecond"},
				Ordering:         connectors.PollingOrderingTuple{Watermark: connectors.PollingOrderingField{CatalogField: "updated_at", Ascending: true}, TieBreaker: connectors.PollingOrderingField{CatalogField: "id", Ascending: true, Unique: true}},
				Mutation:         connectors.PollingMutationPolicy{Mutable: true, CommitOrdered: true, BoundedOverlap: true},
				Identity:         connectors.PollingSourceIdentity{Engine: "postgres", AccountScope: "live-target", ObjectScope: "managed-target"},
				Schema:           connectors.PollingSchemaCompatibilityExactFingerprint,
				DeleteVisibility: connectors.PollingDeleteVisibilityTombstone,
				SoftDeleteField:  "deleted_at", SoftDeleteAdvancesCursor: true,
				Modes: []synccontract.Mode{
					synccontract.ModeFullOverwrite, synccontract.ModeFullAppend, synccontract.ModeIncrementalAppend,
					synccontract.ModeIncrementalUpsert, synccontract.ModeIncrementalDedupe, synccontract.ModeIncrementalDedupeHistory,
				},
			},
			Target: connectors.PollingApplyDescriptor{
				Executor: targetReference, MaxBatchRecords: 2, MaxBatchBytes: 1 << 20,
				Staging: connectors.PollingStagingReplaceSupported, StableKeyMapping: []string{"tenant", "id"}, ConditionalOrderFence: true,
				Transaction: connectors.PollingTransactionRequired, PartialResult: connectors.PollingPartialResultRollback,
				RetrySafeCloseAndInsert: true, ValidityWindow: connectors.PollingValidityWindowSupported,
				Strategies: []connectors.PollingApplyStrategy{
					connectors.PollingApplyStrategyReplace, connectors.PollingApplyStrategyAppend, connectors.PollingApplyStrategyMerge,
					connectors.PollingApplyStrategyDedupe, connectors.PollingApplyStrategyDedupeHistory,
				},
			},
		},
		object:   connectors.PollingCatalogObject{Kind: connectors.PollingCatalogObjectRelation, NameParts: []string{"public", "managed_target"}, Columns: []string{"tenant", "id", "updated_at", "deleted_at"}},
		registry: registry,
		apply:    apply,
	}
}

func postgresApplyManagedTargetPollingPage(t *testing.T, ctx context.Context, polling postgresManagedTargetPollingApply, mode synccontract.Mode, records []database.OrderedRecord, tombstones database.TombstoneEnvelope) {
	t.Helper()
	resolved, err := engine.PollingPreflight(ctx, polling.registry, polling.declaration, polling.object, mode)
	if err != nil {
		t.Fatalf("PostgreSQL native polling preflight for %s failed: %v", mode, err)
	}
	page := engine.PollingApplyPage{Records: make([]engine.PollingApplyRecord, len(records)), Tombstones: tombstones.Tombstones()}
	for index, record := range records {
		page.Records[index] = engine.PollingApplyRecord{Record: record.Record, Position: record.Position}
	}
	acknowledgement, err := engine.ApplyPollingPage(ctx, resolved, page)
	if err != nil || acknowledgement.Sink == "" || acknowledgement.AcknowledgedAt.IsZero() {
		t.Fatalf("PostgreSQL native polling apply for %s = (%#v, %v), want durable acknowledgement", mode, acknowledgement, err)
	}
}

type postgresManagedTargetPollingSource struct {
	reference connectors.TransportExecutorReference
	evidence  engine.PollingWatermarkConformanceEvidence
}

func (s *postgresManagedTargetPollingSource) PollingSourceExecutorReference() connectors.TransportExecutorReference {
	return s.reference
}

func (s *postgresManagedTargetPollingSource) PollingSourceConformanceEvidence() engine.PollingWatermarkConformanceEvidence {
	return s.evidence
}

func postgresManagedTargetPosition(sequence byte) synccontract.CheckpointPosition {
	return synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken{sequence}, TieBreaker: synccontract.OpaqueToken{1}}
}

func postgresManagedTargetTombstone(t *testing.T, tenant string, id int64, sequence byte) synccontract.Tombstone {
	t.Helper()
	key, err := json.Marshal(map[string]any{"tenant": tenant, "id": id})
	if err != nil {
		t.Fatal("could not encode PostgreSQL tombstone key")
	}
	return synccontract.Tombstone{
		Operation:   synccontract.OperationDelete,
		EventID:     synccontract.OpaqueToken{sequence, 0xff},
		Key:         key,
		DeleteImage: synccontract.DeleteImageKeyOnly,
		Position:    postgresManagedTargetPosition(sequence),
	}
}

func postgresManagedTargetTombstoneEnvelope(t *testing.T, tombstone synccontract.Tombstone) database.TombstoneEnvelope {
	t.Helper()
	envelope, err := database.NewTombstoneEnvelope([]synccontract.Tombstone{tombstone})
	if err != nil {
		t.Fatal("could not construct PostgreSQL source-ordered tombstone envelope")
	}
	return envelope
}

func assertPostgresManagedTargetValue(t *testing.T, ctx context.Context, source *pgx.Conn, control database.ManagedTargetControlRecord, tenant string, id int64, want string, wantRows int) {
	t.Helper()
	rows, err := source.Query(ctx, "SELECT value FROM "+postgresManagedTargetQualifiedName(control.Target().Namespace(), control.Target().Relation())+" WHERE tenant = $1 AND id = $2", tenant, id)
	if err != nil {
		t.Fatal("could not query PostgreSQL fenced target value")
	}
	defer rows.Close()
	values := make([]string, 0, wantRows)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			t.Fatal("could not scan PostgreSQL fenced target value")
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil || len(values) != wantRows || (wantRows == 1 && values[0] != want) {
		t.Fatalf("PostgreSQL fenced target value %q/%d = %#v (%v), want %q with %d row(s)", tenant, id, values, err, want, wantRows)
	}
}

func assertPostgresManagedTargetHistory(t *testing.T, ctx context.Context, source *pgx.Conn, control database.ManagedTargetControlRecord, tenant string, id int64, want []postgresManagedTargetHistoryRow) {
	t.Helper()
	rows, err := source.Query(ctx, "SELECT value, "+synccontract.HistoryValidFromColumn+" IS NOT NULL, "+synccontract.HistoryValidToColumn+" IS NULL, "+synccontract.HistoryIsCurrentColumn+" FROM "+postgresManagedTargetQualifiedName(control.Target().Namespace(), control.Target().Relation())+" WHERE tenant = $1 AND id = $2 AND "+synccontract.HistoryValidFromColumn+" IS NOT NULL ORDER BY value", tenant, id)
	if err != nil {
		t.Fatal("could not query PostgreSQL history target rows")
	}
	defer rows.Close()
	got := make([]postgresManagedTargetHistoryRow, 0, len(want))
	for rows.Next() {
		var row postgresManagedTargetHistoryRow
		var hasValidFrom bool
		if err := rows.Scan(&row.value, &hasValidFrom, &row.open, &row.current); err != nil {
			t.Fatal("could not scan PostgreSQL history target row")
		}
		if !hasValidFrom {
			t.Fatal("PostgreSQL history row has no validity start")
		}
		got = append(got, row)
	}
	if err := rows.Err(); err != nil || len(got) != len(want) {
		t.Fatalf("PostgreSQL history rows = %#v (%v), want %#v", got, err, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("PostgreSQL history row[%d] = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func assertPostgresManagedTargetReceiptPersisted(t *testing.T, ctx context.Context, driver *native.DatabaseDriver, control database.ManagedTargetControlRecord, receipt database.DeliveryReceiptV1) {
	t.Helper()
	if got := postgresManagedTargetDeliveryID(t, ctx, driver, control); got != receipt.DeliveryID() {
		t.Fatalf("PostgreSQL session receipt was not persisted in the delivery ledger: got=%q want=%q", got, receipt.DeliveryID())
	}
}

func postgresManagedTargetDeliveryID(t *testing.T, ctx context.Context, driver *native.DatabaseDriver, control database.ManagedTargetControlRecord) string {
	t.Helper()
	key, err := database.NewManagedTargetDeliveryLedgerKey(control)
	if err != nil {
		t.Fatal("could not derive PostgreSQL delivery ledger key")
	}
	record, found, err := driver.LoadManagedTargetDelivery(ctx, key)
	if err != nil || !found || record.DeliveryID() == "" {
		t.Fatalf("PostgreSQL durable delivery evidence was not readable: record=%#v found=%t error=%v", record, found, err)
	}
	return record.DeliveryID()
}

type postgresManagedTargetRow struct {
	tenant string
	id     int64
	value  string
}

type postgresManagedTargetHistoryRow struct {
	tenant    string
	id        int64
	value     string
	validFrom time.Time
	validTo   *time.Time
	open      bool
	current   bool
}

func postgresManagedTargetRows(t *testing.T, ctx context.Context, source *pgx.Conn, control database.ManagedTargetControlRecord) []postgresManagedTargetRow {
	t.Helper()
	rows, err := source.Query(ctx, "SELECT tenant, id, value, pg_typeof(id)::text, pg_typeof(tenant)::text FROM "+postgresManagedTargetQualifiedName(control.Target().Namespace(), control.Target().Relation())+" ORDER BY tenant, id")
	if err != nil {
		t.Fatal("could not read live PostgreSQL managed target rows")
	}
	defer rows.Close()
	result := make([]postgresManagedTargetRow, 0)
	for rows.Next() {
		var row postgresManagedTargetRow
		var idType, tenantType string
		if err := rows.Scan(&row.tenant, &row.id, &row.value, &idType, &tenantType); err != nil || idType != "bigint" || tenantType != "text" {
			t.Fatalf("PostgreSQL managed target typed row scan = (%#v, id_type=%q, tenant_type=%q, %v), want bigint/text", row, idType, tenantType, err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal("could not finish reading live PostgreSQL managed target rows")
	}
	return result
}

func assertPostgresManagedTargetRows(t *testing.T, ctx context.Context, source *pgx.Conn, control database.ManagedTargetControlRecord, want []postgresManagedTargetRow) {
	t.Helper()
	if got := postgresManagedTargetRows(t, ctx, source, control); !samePostgresManagedTargetRows(got, want) {
		t.Fatalf("PostgreSQL managed target durable rows = %#v, want %#v", got, want)
	}
}

func samePostgresManagedTargetRows(left, right []postgresManagedTargetRow) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func postgresManagedTargetHistoryRows(t *testing.T, ctx context.Context, source *pgx.Conn, control database.ManagedTargetControlRecord) []postgresManagedTargetHistoryRow {
	t.Helper()
	rows, err := source.Query(ctx,
		"SELECT tenant, id, value, "+
			postgresManagedTargetQualifiedName("_valid_from")+", "+
			postgresManagedTargetQualifiedName("_valid_to")+", "+
			postgresManagedTargetQualifiedName("_is_current")+
			" FROM "+postgresManagedTargetQualifiedName(control.Target().Namespace(), control.Target().Relation())+
			" ORDER BY tenant, id, "+postgresManagedTargetQualifiedName("_valid_from")+", value",
	)
	if err != nil {
		t.Fatal("could not read live PostgreSQL managed history rows")
	}
	defer rows.Close()
	result := make([]postgresManagedTargetHistoryRow, 0)
	for rows.Next() {
		var row postgresManagedTargetHistoryRow
		if err := rows.Scan(&row.tenant, &row.id, &row.value, &row.validFrom, &row.validTo, &row.current); err != nil {
			t.Fatalf("could not scan live PostgreSQL managed history row: %v", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal("could not finish reading live PostgreSQL managed history rows")
	}
	return result
}

func samePostgresManagedTargetHistoryRows(left, right []postgresManagedTargetHistoryRow) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].tenant != right[index].tenant || left[index].id != right[index].id || left[index].value != right[index].value || !left[index].validFrom.Equal(right[index].validFrom) || left[index].current != right[index].current {
			return false
		}
		switch {
		case left[index].validTo == nil && right[index].validTo == nil:
		case left[index].validTo != nil && right[index].validTo != nil && left[index].validTo.Equal(*right[index].validTo):
		default:
			return false
		}
	}
	return true
}

// postgresAfterApplyDriver is a narrow failure-injection seam around a real
// PostgreSQL driver. The wrapped session and transaction are still native; the
// callback only makes cancellation/disconnect timing deterministic after a
// server-side batch mutation and before CommitWrite.
type postgresAfterApplyDriver struct {
	driver     *native.DatabaseDriver
	afterApply func()
	beginCalls int
}

func (d *postgresAfterApplyDriver) DatabaseDriverDescriptor() database.DriverDescriptor {
	return d.driver.DatabaseDriverDescriptor()
}

func (d *postgresAfterApplyDriver) DatabaseWriteCapabilities() database.DatabaseWriteCapabilities {
	return d.driver.DatabaseWriteCapabilities()
}

func (d *postgresAfterApplyDriver) PreviewDatabaseWrite(ctx context.Context, plan database.DatabaseWritePlan) (database.DatabaseWritePreview, error) {
	return d.driver.PreviewDatabaseWrite(ctx, plan)
}

func (d *postgresAfterApplyDriver) BeginDatabaseWrite(ctx context.Context, plan database.DatabaseWritePlan) (database.WriteSession, error) {
	d.beginCalls++
	session, err := d.driver.BeginDatabaseWrite(ctx, plan)
	if err != nil {
		return nil, err
	}
	return &postgresAfterApplySession{WriteSession: session, afterApply: d.afterApply}, nil
}

type postgresAfterApplySession struct {
	database.WriteSession
	afterApply func()
	once       sync.Once
}

func (s *postgresAfterApplySession) ApplyWriteBatch(ctx context.Context, batch database.WriteBatch) error {
	if err := s.WriteSession.ApplyWriteBatch(ctx, batch); err != nil {
		return err
	}
	if s.afterApply != nil {
		s.once.Do(s.afterApply)
	}
	return nil
}

type postgresManagedTargetState struct {
	namespaceOID      string
	relationOID       string
	ownerRows         int
	controlRows       int
	ledgerRows        int
	ownerConnectionID string
	streamID          string
	fingerprint       string
}

func observePostgresManagedTargetState(t *testing.T, ctx context.Context, source *pgx.Conn, target database.ManagedTargetRef) postgresManagedTargetState {
	t.Helper()
	ownerTable := postgresManagedTargetQualifiedName(target.Namespace(), "__polymetrics_namespace_owner")
	controlTable := postgresManagedTargetQualifiedName(target.Namespace(), "__polymetrics_target_control")
	ledgerTable := postgresManagedTargetQualifiedName(target.Namespace(), "__polymetrics_delivery_ledger")
	var state postgresManagedTargetState
	if err := source.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_namespace WHERE nspname = $1", target.Namespace()).Scan(&state.namespaceOID); err != nil {
		t.Fatal("could not observe PostgreSQL managed target namespace state")
	}
	if err := source.QueryRow(ctx, "SELECT c.oid::text FROM pg_catalog.pg_class AS c JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace WHERE n.nspname = $1 AND c.relname = $2", target.Namespace(), target.Relation()).Scan(&state.relationOID); err != nil {
		t.Fatal("could not observe PostgreSQL managed target relation state")
	}
	if err := source.QueryRow(ctx, "SELECT count(*) FROM "+ownerTable).Scan(&state.ownerRows); err != nil {
		t.Fatal("could not observe PostgreSQL managed target owner state")
	}
	if err := source.QueryRow(ctx, "SELECT count(*) FROM "+controlTable).Scan(&state.controlRows); err != nil {
		t.Fatal("could not observe PostgreSQL managed target control count")
	}
	if err := source.QueryRow(ctx, "SELECT count(*) FROM "+ledgerTable).Scan(&state.ledgerRows); err != nil {
		t.Fatal("could not observe PostgreSQL managed target ledger count")
	}
	if err := source.QueryRow(ctx, "SELECT connection_id FROM "+ownerTable).Scan(&state.ownerConnectionID); err != nil {
		t.Fatal("could not observe PostgreSQL managed target owner row")
	}
	if err := source.QueryRow(ctx, "SELECT stream_id, encode(schema_fingerprint, 'hex') FROM "+controlTable+" WHERE relation_name = $1", target.Relation()).Scan(&state.streamID, &state.fingerprint); err != nil {
		t.Fatal("could not observe PostgreSQL managed target control row")
	}
	return state
}

func assertPostgresManagedTargetState(t *testing.T, ctx context.Context, source *pgx.Conn, target database.ManagedTargetRef, want postgresManagedTargetState) {
	t.Helper()
	if got := observePostgresManagedTargetState(t, ctx, source, target); got != want {
		t.Fatalf("PostgreSQL managed target durable state changed during refusal: got=%+v want=%+v", got, want)
	}
}

func assertPostgresManagedTargetOIDReplacementRefusal(t *testing.T, ctx context.Context, source *pgx.Conn, provisioner *database.ManagedTargetProvisioner, fixture postgresManagedTargetFixture) {
	t.Helper()
	relation := postgresManagedTargetQualifiedName(fixture.target.Namespace(), fixture.target.Relation())
	if _, err := source.Exec(ctx, "DROP TABLE "+relation); err != nil {
		t.Fatal("could not replace the PostgreSQL managed target relation")
	}
	if _, err := source.Exec(ctx, "CREATE TABLE "+relation+" (id bigint NOT NULL PRIMARY KEY)"); err != nil {
		t.Fatal("could not create same-named PostgreSQL replacement relation")
	}
	before := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetReplaced) {
		t.Fatalf("PostgreSQL OID replacement error = %v, want ErrManagedTargetReplaced", err)
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, before)
}

func assertPostgresManagedTargetSchemaDriftRefusal(t *testing.T, ctx context.Context, source *pgx.Conn, provisioner *database.ManagedTargetProvisioner, fixture postgresManagedTargetFixture) {
	t.Helper()
	controlTable := postgresManagedTargetQualifiedName(fixture.target.Namespace(), "__polymetrics_target_control")
	fingerprint := sha256.Sum256([]byte("postgres-managed-target-live-schema-v2"))
	if _, err := source.Exec(ctx, "UPDATE "+controlTable+" SET schema_fingerprint = $1 WHERE relation_name = $2", fingerprint[:], fixture.target.Relation()); err != nil {
		t.Fatal("could not create PostgreSQL managed target schema drift")
	}
	before := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetSchemaDrift) {
		t.Fatalf("PostgreSQL schema drift error = %v, want ErrManagedTargetSchemaDrift", err)
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, before)
	original := fixture.schema.Fingerprint()
	if _, err := source.Exec(ctx, "UPDATE "+controlTable+" SET schema_fingerprint = $1 WHERE relation_name = $2", original[:], fixture.target.Relation()); err != nil {
		t.Fatal("could not restore PostgreSQL managed target schema assertion")
	}
}

func assertPostgresManagedTargetForeignOwnerRefusal(t *testing.T, ctx context.Context, source *pgx.Conn, provisioner *database.ManagedTargetProvisioner, fixture postgresManagedTargetFixture) {
	t.Helper()
	ownerTable := postgresManagedTargetQualifiedName(fixture.target.Namespace(), "__polymetrics_namespace_owner")
	if _, err := source.Exec(ctx, "UPDATE "+ownerTable+" SET connection_id = $1", "foreign-managed-target-connection"); err != nil {
		t.Fatal("could not create a foreign PostgreSQL namespace owner record")
	}
	before := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetUnverifiable) {
		t.Fatalf("PostgreSQL foreign owner error = %v, want fail-closed ErrManagedTargetUnverifiable", err)
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, before)
	identity := fixture.owner.Identity()
	if _, err := source.Exec(ctx, "UPDATE "+ownerTable+" SET connection_id = $1", identity.ConnectionID); err != nil {
		t.Fatal("could not restore PostgreSQL managed target namespace owner")
	}
}

func assertPostgresManagedTargetControlCollisionRefusal(t *testing.T, ctx context.Context, source *pgx.Conn, provisioner *database.ManagedTargetProvisioner, fixture postgresManagedTargetFixture) {
	t.Helper()
	controlTable := postgresManagedTargetQualifiedName(fixture.target.Namespace(), "__polymetrics_target_control")
	if _, err := source.Exec(ctx, "UPDATE "+controlTable+" SET stream_id = $1 WHERE relation_name = $2", "other-managed-target-stream", fixture.target.Relation()); err != nil {
		t.Fatal("could not create a PostgreSQL managed target control collision")
	}
	before := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetUnverifiable) {
		t.Fatalf("PostgreSQL control collision error = %v, want fail-closed ErrManagedTargetUnverifiable", err)
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, before)
	if _, err := source.Exec(ctx, "UPDATE "+controlTable+" SET stream_id = $1 WHERE relation_name = $2", fixture.target.StreamID(), fixture.target.Relation()); err != nil {
		t.Fatal("could not restore PostgreSQL managed target control record")
	}
}

func assertPostgresManagedTargetPermissionRefusal(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint, source *pgx.Conn, fixture postgresManagedTargetFixture) {
	t.Helper()
	if _, err := source.Exec(ctx, "CREATE ROLE "+postgresManagedTargetRestrictedUser+" LOGIN"); err != nil {
		t.Fatal("could not create PostgreSQL managed target restricted role")
	}
	if _, err := source.Exec(ctx, "GRANT CONNECT ON DATABASE "+postgresManagedTargetQualifiedName(postgresCatalogIntegrationDatabase)+" TO "+postgresManagedTargetQualifiedName(postgresManagedTargetRestrictedUser)); err != nil {
		t.Fatal("could not grant PostgreSQL managed target restricted database access")
	}
	if _, err := source.Exec(ctx, "GRANT USAGE ON SCHEMA "+postgresManagedTargetQualifiedName(fixture.target.Namespace())+" TO "+postgresManagedTargetQualifiedName(postgresManagedTargetRestrictedUser)); err != nil {
		t.Fatal("could not grant PostgreSQL managed target restricted schema usage")
	}
	restricted := openPostgresCatalogUser(t, ctx, endpoint, postgresManagedTargetRestrictedUser)
	defer func() { _ = restricted.Close(ctx) }()
	driver, err := native.NewDatabaseDriver(restricted)
	if err != nil {
		t.Fatal("could not construct restricted PostgreSQL managed target driver")
	}
	var databaseOID, namespaceOID string
	if err := restricted.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_database WHERE datname = current_database()").Scan(&databaseOID); err != nil || databaseOID == "" {
		t.Fatal("restricted PostgreSQL session could not observe its destination database identity")
	}
	if err := restricted.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_namespace WHERE nspname = $1", fixture.target.Namespace()).Scan(&namespaceOID); err != nil || namespaceOID == "" {
		t.Fatal("restricted PostgreSQL session could not observe the derived namespace identity")
	}
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal("could not construct restricted PostgreSQL managed target provisioner")
	}
	observation, err := driver.ObserveManagedTarget(ctx, fixture.target)
	if err != nil {
		t.Fatal("restricted PostgreSQL managed target observation did not reach the private control boundary")
	}
	if observation.NamespaceOwnerState != database.ManagedTargetNamespaceOwnerUnreadable || observation.ControlState != database.ManagedTargetControlUnreadable {
		t.Fatalf("restricted PostgreSQL managed target observation = namespace_owner=%d control=%d, want both unreadable", observation.NamespaceOwnerState, observation.ControlState)
	}
	before := observePostgresManagedTargetState(t, ctx, source, fixture.target)
	if _, err := provisioner.CreateOrAssert(ctx, fixture.plan); !errors.Is(err, database.ErrManagedTargetOwnerUnreadable) {
		t.Fatalf("PostgreSQL managed target permission error = %v, want ErrManagedTargetOwnerUnreadable", err)
	}
	assertPostgresManagedTargetState(t, ctx, source, fixture.target, before)
}

func assertPostgresManagedTargetDurability(t *testing.T, ctx context.Context, source *pgx.Conn, driver *native.DatabaseDriver) {
	t.Helper()
	if err := driver.PreflightDurability(ctx); err != nil {
		t.Fatalf("PostgreSQL default durability preflight failed: %v", err)
	}
	if _, err := source.Exec(ctx, "SET synchronous_commit = off"); err != nil {
		t.Fatal("could not set PostgreSQL session durability test value")
	}
	var setting string
	if err := source.QueryRow(ctx, "SELECT current_setting('synchronous_commit')").Scan(&setting); err != nil || setting != "off" {
		t.Fatal("PostgreSQL session did not apply unsafe durability test setting")
	}
	if err := driver.PreflightDurability(ctx); err == nil {
		t.Fatal("PostgreSQL durability preflight accepted synchronous_commit=off")
	}
	if _, err := source.Exec(ctx, "SET synchronous_commit = on"); err != nil {
		t.Fatal("could not restore PostgreSQL session durability test value")
	}
	if err := driver.PreflightDurability(ctx); err != nil {
		t.Fatal("PostgreSQL durability preflight did not observe restored synchronous_commit=on")
	}
}

func openPostgresCatalogUser(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint, user string) *pgx.Conn {
	t.Helper()
	config, err := pgx.ParseConfig("sslmode=disable")
	if err != nil {
		t.Fatal("could not configure the isolated PostgreSQL restricted source")
	}
	config.Host = endpoint.Host
	config.Port = uint16(endpoint.Port)
	config.Database = postgresCatalogIntegrationDatabase
	config.User = user
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatal("could not open the isolated PostgreSQL restricted source")
	}
	return conn
}

func postgresManagedTargetQualifiedName(parts ...string) string {
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		quoted = append(quoted, `"`+strings.ReplaceAll(part, `"`, `""`)+`"`)
	}
	return strings.Join(quoted, ".")
}
