//go:build databaseintegration

package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/native/dbtest"
	native "polymetrics.ai/internal/connectors/native/postgres"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

const (
	postgresCDCRouteImage      = "docker.io/library/postgres:16.10"
	postgresCDCRouteDatabase   = "pm_cdc_route"
	postgresCDCRouteUser       = "postgres"
	postgresCDCRouteImageBytes = 420 << 20
	// One pgoutput transaction below contains an insert, update, and delete.
	// The target page must accept that committed unit before its source LSN is
	// acknowledged; a smaller borrowed fixture limit would test only a local
	// harness refusal, not the live route.
	postgresCDCRouteTargetBatchLimit = 3
)

// TestPostgresCDCToManagedTargetHistoryRouteLive proves the missing R4
// composition: pgoutput is the source of the insert/update/delete events;
// the committed transaction becomes a durable local warehouse receipt and
// sealed workset before the keyed PostgreSQL history target is acknowledged.
// It deliberately reads the target through a separate normal PostgreSQL
// connection, rather than relying on a writer result.
func TestPostgresCDCToManagedTargetHistoryRouteLive(t *testing.T) {
	runPostgresCDCToManagedTargetHistoryRouteLive(t, postgresCDCRouteRestartPlan{name: "source and target", source: true, target: true})
}

// TestPostgresCDCToManagedTargetHistoryRouteRestartCounterfactualsLive makes
// the restart proof diagnosable. These cases distinguish a pgoutput resume
// defect from a target driver restart defect without changing the production
// timeout: an acknowledged transaction must progress promptly in every
// supported lifecycle.
func TestPostgresCDCToManagedTargetHistoryRouteRestartCounterfactualsLive(t *testing.T) {
	for _, plan := range []postgresCDCRouteRestartPlan{
		{name: "no restart"},
		{name: "source-only restart", source: true},
		{name: "target-only restart", target: true},
	} {
		t.Run(plan.name, func(t *testing.T) {
			runPostgresCDCToManagedTargetHistoryRouteLive(t, plan)
		})
	}
}

type postgresCDCRouteRestartPlan struct {
	name   string
	source bool
	target bool
}

func runPostgresCDCToManagedTargetHistoryRouteLive(t *testing.T, restart postgresCDCRouteRestartPlan) {
	t.Helper()
	if os.Getenv(postgresCatalogIntegrationEnabledEnv) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the PostgreSQL CDC-to-target proof", postgresCatalogIntegrationEnabledEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	harness, err := newPostgresCDCRouteContainerHarness(
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
			t.Error("PostgreSQL CDC route test cleanup failed")
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("start PostgreSQL CDC route container: %v", err)
	}

	config := postgresCDCRouteConfig(t, endpoint)
	connector := native.New()
	waitForPostgresCDCRoute(t, ctx, connector, config)
	source := openPostgresCDCRouteConnection(t, ctx, endpoint)
	defer func() { _ = source.Close(context.WithoutCancel(ctx)) }()
	assertPostgresCDCRoutePrerequisites(t, ctx, source)

	targetConnection := openPostgresCDCRouteConnection(t, ctx, endpoint)
	driver, err := native.NewDatabaseDriver(targetConnection)
	if err != nil {
		t.Fatal("construct live PostgreSQL history target driver")
	}
	fixture := newPostgresManagedTargetFixture(t, ctx, driver)
	provisioner, err := database.NewManagedTargetProvisioner(driver)
	if err != nil {
		t.Fatal("construct PostgreSQL history target provisioner")
	}
	control, err := provisioner.CreateOrAssert(ctx, fixture.plan)
	if err != nil {
		t.Fatal("create PostgreSQL history target")
	}
	definition := postgresManagedTargetWriteDefinition(t)

	worksetRoot := filepath.Join(config.ProjectDir, "cdc-route-worksets")
	receiver, err := newPostgresCDCRouteReceiver(ctx, worksetRoot, control, fixture.plan)
	if err != nil {
		t.Fatalf("construct CDC route receiver: %v", err)
	}
	polling, err := newPostgresCDCRoutePolling(driver, definition, control, fixture)
	if err != nil {
		t.Fatalf("construct CDC route history target: %v", err)
	}
	activeTarget := targetConnection
	defer func() { _ = activeTarget.Close(context.WithoutCancel(ctx)) }()

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	table := "pm_cdc_route_" + suffix
	publication := "pm_cdc_route_pub_" + suffix
	stream := "public." + table
	if _, err := source.Exec(ctx, "CREATE TABLE "+postgresCDCRouteIdentifier(table)+" (source_tenant text NOT NULL, source_id bigint NOT NULL, source_value text NOT NULL, PRIMARY KEY (source_tenant, source_id))"); err != nil {
		t.Fatal("create PostgreSQL CDC route source table")
	}
	defer func() {
		_, _ = source.Exec(context.Background(), "DROP TABLE IF EXISTS "+postgresCDCRouteIdentifier(table))
	}()
	if _, err := source.Exec(ctx, "CREATE PUBLICATION "+postgresCDCRouteIdentifier(publication)+" FOR TABLE "+postgresCDCRouteIdentifier(table)); err != nil {
		t.Fatal("create PostgreSQL CDC route publication")
	}
	defer func() {
		_, _ = source.Exec(context.Background(), "DROP PUBLICATION IF EXISTS "+postgresCDCRouteIdentifier(publication))
	}()
	config.Config["cdc_publication"] = publication

	slot, err := connector.CDCSlotName(ctx, config, stream)
	if err != nil {
		t.Fatal("derive PostgreSQL CDC route slot")
	}
	defer func() {
		teardownCtx, teardownCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer teardownCancel()
		if err := connector.TeardownCDC(teardownCtx, config, stream); err != nil {
			t.Error("tear down PostgreSQL CDC route slot")
		}
	}()

	firstCommitter := &postgresCDCRouteCommitter{receiver: receiver, polling: polling, acknowledgementSource: source, slot: slot, checkpoints: make(chan synccontract.CheckpointEnvelope, 2)}
	firstContext, stopFirst := context.WithCancel(ctx)
	firstDone := startPostgresCDCRouteRead(connector, firstContext, connectors.CDCReadRequest{
		Stream:                     stream,
		Config:                     config,
		TransactionReceiver:        receiver,
		DurableCheckpointCommitter: firstCommitter,
	})
	waitForPostgresCDCRouteSlot(t, ctx, source, slot, true)

	initial, err := source.Begin(ctx)
	if err != nil {
		t.Fatal("begin initial source transaction")
	}
	if _, err := initial.Exec(ctx, "INSERT INTO "+postgresCDCRouteIdentifier(table)+" (source_tenant, source_id, source_value) VALUES ('north', 1, 'v1'), ('south', 2, 'v1')"); err != nil {
		t.Fatal("insert initial CDC source rows")
	}
	if err := initial.Commit(ctx); err != nil {
		t.Fatal("commit initial source transaction")
	}
	firstCheckpoint := firstCommitter.wait(t, ctx, firstDone)
	assertPostgresCDCRouteAcknowledged(t, ctx, source, slot, firstCheckpoint)
	firstSlot := observePostgresCDCRouteSlot(t, ctx, source, slot)
	if !firstSlot.exists || !firstSlot.active {
		t.Fatalf("PostgreSQL CDC route first acknowledgement slot = %#v, want an existing active slot", firstSlot)
	}
	t.Logf("PostgreSQL CDC route after first acknowledgement: %s", firstSlot)
	assertPostgresCDCRouteStageReceipts(t, config.ProjectDir, 1)
	assertPostgresCDCRouteHistory(t, ctx, source, control, map[string]postgresCDCRouteHistoryExpectation{
		"north/1": {versions: []string{"v1"}, current: "v1"},
		"south/2": {versions: []string{"v1"}, current: "v1"},
	})

	activeCommitter := firstCommitter
	activeStop := stopFirst
	activeDone := firstDone
	if restart.source {
		stopFirst()
		assertPostgresCDCRouteStopped(t, firstDone)
		waitForPostgresCDCRouteSlot(t, ctx, source, slot, false)

		// The source reader must resume from the sealed receipt's checkpoint,
		// not a hand-constructed target event or a newer source position.
		secondCommitter := &postgresCDCRouteCommitter{receiver: receiver, polling: polling, acknowledgementSource: source, slot: slot, checkpoints: make(chan synccontract.CheckpointEnvelope, 2)}
		secondContext, stopSecond := context.WithCancel(ctx)
		resume := firstCheckpoint.Clone()
		secondDone := startPostgresCDCRouteRead(connector, secondContext, connectors.CDCReadRequest{
			Stream:                     stream,
			Config:                     config,
			Checkpoint:                 &resume,
			TransactionReceiver:        receiver,
			DurableCheckpointCommitter: secondCommitter,
		})
		activeCommitter, activeStop, activeDone = secondCommitter, stopSecond, secondDone
		waitForPostgresCDCRouteSlot(t, ctx, source, slot, true)
		resumedSlot := observePostgresCDCRouteSlot(t, ctx, source, slot)
		if !resumedSlot.exists || !resumedSlot.active || resumedSlot.confirmed < firstSlot.confirmed {
			t.Fatalf("PostgreSQL CDC route source restart slot = %#v, want active slot retaining acknowledged LSN %s", resumedSlot, firstSlot.confirmed)
		}
		t.Logf("PostgreSQL CDC route after source restart: %s", resumedSlot)
	}

	if restart.target {
		if err := activeTarget.Close(context.WithoutCancel(ctx)); err != nil {
			t.Fatal("close PostgreSQL history target connection for restart")
		}
		restartedTarget := openPostgresCDCRouteConnection(t, ctx, endpoint)
		restartedDriver, err := native.NewDatabaseDriver(restartedTarget)
		if err != nil {
			t.Fatal("reconstruct PostgreSQL history target driver")
		}
		restartedPolling, err := newPostgresCDCRoutePolling(restartedDriver, definition, control, fixture)
		if err != nil {
			t.Fatal("reconstruct CDC route history target")
		}
		activeTarget = restartedTarget
		activeCommitter.setPolling(restartedPolling)
		restartedSlot := observePostgresCDCRouteSlot(t, ctx, source, slot)
		if !restartedSlot.exists || !restartedSlot.active || restartedSlot.confirmed < firstSlot.confirmed {
			t.Fatalf("PostgreSQL CDC route target restart slot = %#v, want active slot retaining acknowledged LSN %s", restartedSlot, firstSlot.confirmed)
		}
		t.Logf("PostgreSQL CDC route after target restart: %s", restartedSlot)
	}

	changes, err := source.Begin(ctx)
	if err != nil {
		t.Fatal("begin insert/update/delete source transaction")
	}
	if _, err := changes.Exec(ctx, "INSERT INTO "+postgresCDCRouteIdentifier(table)+" (source_tenant, source_id, source_value) VALUES ('east', 3, 'new')"); err != nil {
		t.Fatal("insert live CDC source row")
	}
	if _, err := changes.Exec(ctx, "UPDATE "+postgresCDCRouteIdentifier(table)+" SET source_value = 'v2' WHERE source_tenant = 'north' AND source_id = 1"); err != nil {
		t.Fatal("update live CDC source row")
	}
	if _, err := changes.Exec(ctx, "DELETE FROM "+postgresCDCRouteIdentifier(table)+" WHERE source_tenant = 'south' AND source_id = 2"); err != nil {
		t.Fatal("delete live CDC source row")
	}
	if err := changes.Commit(ctx); err != nil {
		t.Fatal("commit live insert/update/delete source transaction")
	}
	secondCheckpoint := activeCommitter.wait(t, ctx, activeDone)
	if string(secondCheckpoint.Position.Primary) == string(firstCheckpoint.Position.Primary) {
		t.Fatal("PostgreSQL CDC route did not advance its durable checkpoint after the second real transaction")
	}
	assertPostgresCDCRouteAcknowledged(t, ctx, source, slot, secondCheckpoint)
	assertPostgresCDCRouteStageReceipts(t, config.ProjectDir, 2)
	last := receiver.lastTransaction()
	if last == nil || last.workset.Changes() != 2 || last.workset.TombstoneCount() != 1 || last.receipt.ID() == "" {
		t.Fatalf("live CDC route workset = %#v, want two source changes, one real delete tombstone, and a durable receipt", last)
	}
	assertPostgresCDCRouteHistory(t, ctx, source, control, map[string]postgresCDCRouteHistoryExpectation{
		"east/3":  {versions: []string{"new"}, current: "new"},
		"north/1": {versions: []string{"v1", "v2"}, current: "v2"},
		"south/2": {versions: []string{"v1"}, current: ""},
	})
	beforeReplay := postgresManagedTargetHistoryRows(t, ctx, source, control)

	activeStop()
	assertPostgresCDCRouteStopped(t, activeDone)
	waitForPostgresCDCRouteSlot(t, ctx, source, slot, false)
	if err := activeTarget.Close(context.WithoutCancel(ctx)); err != nil {
		t.Fatal("close PostgreSQL history target connection for replay")
	}

	// A fresh target connection replays the sealed workset. Independent target
	// read-back must remain unchanged; a duplicate source transaction may not
	// reopen a closed history window or create a second version.
	replayTarget := openPostgresCDCRouteConnection(t, ctx, endpoint)
	defer func() { _ = replayTarget.Close(context.WithoutCancel(ctx)) }()
	replayDriver, err := native.NewDatabaseDriver(replayTarget)
	if err != nil {
		t.Fatal("reconstruct PostgreSQL history target for replay")
	}
	replayPolling, err := newPostgresCDCRoutePolling(replayDriver, definition, control, fixture)
	if err != nil {
		t.Fatal("reconstruct CDC route history replay target")
	}
	if _, err := postgresCDCRouteApplyWorkset(ctx, replayPolling, receiver.mapping, last); err != nil {
		t.Fatalf("replay sealed CDC workset through fresh PostgreSQL target: %v", err)
	}
	if got := postgresManagedTargetHistoryRows(t, ctx, source, control); !samePostgresManagedTargetHistoryRows(got, beforeReplay) {
		t.Fatalf("replayed live CDC workset changed independent PostgreSQL target read-back: got=%#v want=%#v", got, beforeReplay)
	}
}

type postgresCDCRouteHistoryExpectation struct {
	versions []string
	current  string
}

func assertPostgresCDCRouteHistory(t *testing.T, ctx context.Context, source *pgx.Conn, control database.ManagedTargetControlRecord, want map[string]postgresCDCRouteHistoryExpectation) {
	t.Helper()
	rows := postgresManagedTargetHistoryRows(t, ctx, source, control)
	actual := make(map[string][]postgresManagedTargetHistoryRow)
	for _, row := range rows {
		key := fmt.Sprintf("%s/%d", row.tenant, row.id)
		actual[key] = append(actual[key], row)
	}
	if len(actual) != len(want) {
		t.Fatalf("independent history target keys = %#v, want %#v", actual, want)
	}
	for key, expectation := range want {
		versions, found := actual[key]
		if !found || len(versions) != len(expectation.versions) {
			t.Fatalf("independent history target %q = %#v, want %d version(s)", key, versions, len(expectation.versions))
		}
		current := ""
		for index, row := range versions {
			if row.value != expectation.versions[index] || row.validFrom.IsZero() {
				t.Fatalf("independent history target %q version %d = %#v, want value %q", key, index, row, expectation.versions[index])
			}
			if row.current {
				if row.validTo != nil || current != "" {
					t.Fatalf("independent history target %q has invalid current row %#v", key, row)
				}
				current = row.value
			} else if row.validTo == nil {
				t.Fatalf("independent history target %q retained an open non-current row %#v", key, row)
			}
		}
		if current != expectation.current {
			t.Fatalf("independent history target %q current = %q, want %q", key, current, expectation.current)
		}
	}
}

type postgresCDCRouteReceiver struct {
	root       string
	control    database.ManagedTargetControlRecord
	mapping    database.MappingContractV1
	sourceRows map[string]warehouse.Row
	baseline   string
	last       *postgresCDCRouteTransaction
}

type postgresCDCRouteTransaction struct {
	workset     database.ChangeDeliveryWorkset
	positions   map[string]synccontract.CheckpointPosition
	receipt     connectors.CDCTransactionReceipt
	warehouse   string
	transaction string
}

func newPostgresCDCRouteReceiver(ctx context.Context, root string, control database.ManagedTargetControlRecord, plan database.ManagedTargetProvisioningPlan) (*postgresCDCRouteReceiver, error) {
	mapping, ok := plan.Mapping()
	if !ok {
		return nil, errors.New("managed target fixture has no mapping")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	baseline := filepath.Join(root, "warehouse", "baseline.parquet")
	if err := warehouse.WriteTable(ctx, baseline, nil); err != nil {
		return nil, err
	}
	return &postgresCDCRouteReceiver{
		root:       root,
		control:    control,
		mapping:    mapping,
		sourceRows: make(map[string]warehouse.Row),
		baseline:   baseline,
	}, nil
}

func (r *postgresCDCRouteReceiver) ReceiveCDCTransaction(ctx context.Context, transaction connectors.CDCTransaction) (connectors.CDCTransactionReceipt, error) {
	if r == nil || transaction.ID() == "" {
		return connectors.CDCTransactionReceipt{}, errors.New("CDC route receiver has no transaction identity")
	}
	rows := clonePostgresCDCRouteRows(r.sourceRows)
	positions := make(map[string]synccontract.CheckpointPosition)
	tombstones := make([]synccontract.Tombstone, 0)
	var ordinal uint64
	if err := transaction.StreamEvents(ctx, func(event connectors.CDCEvent) error {
		ordinal++
		key, err := postgresCDCRouteRecordKey(event.Record)
		if err != nil {
			return err
		}
		switch event.Operation {
		case "insert", "update":
			position, err := postgresCDCRouteEventPosition(event, ordinal)
			if err != nil {
				return err
			}
			rows[key] = clonePostgresCDCRouteRow(event.Record)
			positions[key] = position
		case "delete":
			tombstone, err := native.CDCDeleteTombstone(event, []string{"source_tenant", "source_id"}, ordinal)
			if err != nil {
				return err
			}
			delete(rows, key)
			tombstones = append(tombstones, tombstone)
		default:
			return fmt.Errorf("unexpected real PostgreSQL CDC operation %q", event.Operation)
		}
		return nil
	}); err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	if int64(ordinal) != transaction.Records() {
		return connectors.CDCTransactionReceipt{}, fmt.Errorf("real CDC transaction streamed %d events, want declared %d", ordinal, transaction.Records())
	}

	digest := sha256.Sum256([]byte(transaction.ID()))
	transactionDigest := hex.EncodeToString(digest[:])
	warehousePath := filepath.Join(r.root, "warehouse", transactionDigest+".parquet")
	if err := warehouse.WriteTable(ctx, warehousePath, postgresCDCRouteSortedRows(rows)); err != nil {
		return connectors.CDCTransactionReceipt{}, fmt.Errorf("persist local warehouse receipt: %w", err)
	}
	workset, err := database.DeriveChangeDeliveryWorkset(ctx, database.ChangeDeliveryWorksetRequest{
		Control:          r.control,
		Keys:             []string{"source_tenant", "source_id"},
		SourceParquet:    warehousePath,
		BaselineParquet:  r.baseline,
		Tombstones:       tombstones,
		Root:             filepath.Join(r.root, "worksets"),
		MaxArtifactBytes: 1 << 20,
	})
	if err != nil {
		return connectors.CDCTransactionReceipt{}, fmt.Errorf("seal CDC warehouse workset: %w", err)
	}
	receipt, err := connectors.NewCDCTransactionReceipt("postgres-cdc-route-"+transactionDigest, "warehouse:postgres-cdc-route", time.Now().UTC())
	if err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	r.sourceRows = rows
	r.last = &postgresCDCRouteTransaction{workset: workset, positions: positions, receipt: receipt, warehouse: warehousePath, transaction: transaction.ID()}
	return receipt, nil
}

func (r *postgresCDCRouteReceiver) promote(transaction *postgresCDCRouteTransaction) error {
	if r == nil || r.last == nil || transaction == nil || r.last.transaction != transaction.transaction || r.last.warehouse != transaction.warehouse {
		return errors.New("CDC route cannot promote an unknown workset baseline")
	}
	r.baseline = transaction.warehouse
	return nil
}

func (r *postgresCDCRouteReceiver) RestoreCDCTransactionReceipt(_ context.Context, transactionID string, receipt connectors.CDCTransactionReceipt) error {
	if r == nil || r.last == nil || r.last.transaction != transactionID || r.last.receipt.ID() != receipt.ID() {
		return errors.New("CDC route durable receipt has no matching local workset")
	}
	return nil
}

func (r *postgresCDCRouteReceiver) lastTransaction() *postgresCDCRouteTransaction {
	if r == nil || r.last == nil {
		return nil
	}
	copy := *r.last
	copy.positions = make(map[string]synccontract.CheckpointPosition, len(r.last.positions))
	for key, position := range r.last.positions {
		copy.positions[key] = position.Clone()
	}
	return &copy
}

type postgresCDCRouteCommitter struct {
	mu                    sync.RWMutex
	receiver              *postgresCDCRouteReceiver
	polling               postgresManagedTargetPollingApply
	acknowledgementSource *pgx.Conn
	slot                  string
	checkpoints           chan synccontract.CheckpointEnvelope
}

func (c *postgresCDCRouteCommitter) CommitDurableChangefeedCheckpoint(ctx context.Context, candidate synccontract.CheckpointEnvelope) error {
	if c == nil || c.receiver == nil {
		return errors.New("CDC route checkpoint committer is unavailable")
	}
	transaction := c.receiver.lastTransaction()
	if transaction == nil || transaction.receipt.ID() == "" || transaction.workset.Identity() == "" {
		return errors.New("CDC route target application started before a durable warehouse receipt/workset")
	}
	if err := assertPostgresCDCRouteUnacknowledged(ctx, c.acknowledgementSource, c.slot, candidate); err != nil {
		return err
	}
	c.mu.RLock()
	polling := c.polling
	c.mu.RUnlock()
	acknowledgement, err := postgresCDCRouteApplyWorkset(ctx, polling, c.receiver.mapping, transaction)
	if err != nil {
		return err
	}
	if err := c.receiver.promote(transaction); err != nil {
		return err
	}
	return synccontract.CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(committed synccontract.CheckpointEnvelope) error {
		copy := committed.Clone()
		select {
		case c.checkpoints <- copy:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}

func (c *postgresCDCRouteCommitter) setPolling(polling postgresManagedTargetPollingApply) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.polling = polling
	c.mu.Unlock()
}

func (c *postgresCDCRouteCommitter) wait(t *testing.T, ctx context.Context, readerDone <-chan error) synccontract.CheckpointEnvelope {
	t.Helper()
	timer := time.NewTimer(15 * time.Second)
	defer timer.Stop()
	select {
	case checkpoint := <-c.checkpoints:
		return checkpoint
	case err := <-readerDone:
		t.Fatalf("PostgreSQL CDC route reader stopped before its durable checkpoint: %v", err)
		return synccontract.CheckpointEnvelope{}
	case <-timer.C:
		last := c.receiver.lastTransaction()
		state := observePostgresCDCRouteSlot(t, ctx, c.acknowledgementSource, c.slot)
		t.Fatalf("PostgreSQL CDC route checkpoint did not arrive within 15 seconds; slot=%s last_durable_transaction=%#v", state, last)
		return synccontract.CheckpointEnvelope{}
	case <-ctx.Done():
		t.Fatal("timed out waiting for a durably targeted PostgreSQL CDC checkpoint")
		return synccontract.CheckpointEnvelope{}
	}
}

func newPostgresCDCRoutePolling(driver *native.DatabaseDriver, definition database.Definition, control database.ManagedTargetControlRecord, fixture postgresManagedTargetFixture) (postgresManagedTargetPollingApply, error) {
	ledger, err := database.NewManagedTargetDeliveryLedger(driver)
	if err != nil {
		return postgresManagedTargetPollingApply{}, err
	}
	write, err := database.NewDatabaseWriteExecutor(driver, ledger)
	if err != nil {
		return postgresManagedTargetPollingApply{}, err
	}
	return newPostgresManagedTargetPollingApplyResult(write, definition, control, fixture)
}

func newPostgresManagedTargetPollingApplyResult(write *database.DatabaseWriteExecutor, definition database.Definition, control database.ManagedTargetControlRecord, fixture postgresManagedTargetFixture) (result postgresManagedTargetPollingApply, resultErr error) {
	// The existing fixture helper deliberately owns the declared PostgreSQL
	// source/target admission model. Convert its test fatal surface into a
	// direct setup call only at this one live route boundary.
	mapping, ok := fixture.plan.Mapping()
	if !ok {
		return result, errors.New("managed target fixture has no mapping")
	}
	sourceReference := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target_polling_source"}
	targetReference := connectors.TransportExecutorReference{Family: connectors.TransportExecutorFamilyNativeDatabase, ID: "postgres_managed_target_polling_target"}
	apply, err := engine.NewDatabasePollingApplyExecutor(engine.DatabasePollingApplyConfig{Reference: targetReference, Write: write, Definition: definition, Control: control, Mapping: mapping, BatchSize: postgresCDCRouteTargetBatchLimit})
	if err != nil {
		return result, err
	}
	registry := engine.NewPollingPreflightRegistry()
	if err := registry.RegisterSource(&postgresManagedTargetPollingSource{reference: sourceReference, definition: definition}); err != nil {
		return result, err
	}
	if err := registry.RegisterApply(apply); err != nil {
		return result, err
	}
	return postgresManagedTargetPollingApply{
		declaration: &connectors.PollingWatermarkDescriptor{
			Status: connectors.PollingWatermarkStatusImplemented,
			Source: connectors.PollingWatermarkSourceDescriptor{
				Executor: sourceReference, Object: connectors.PollingCatalogObjectSelector{Kind: connectors.PollingCatalogObjectRelation},
				Read:     connectors.PollingReadProtocol{Kind: connectors.PollingReadProtocolKeyset, MaxPageSize: 2, MaxPages: 2, MaxRequests: 2, StableTraversal: true, Predicate: connectors.PollingKeysetPredicateLexicographic},
				Snapshot: connectors.PollingSnapshotBarrier{Kind: connectors.PollingSnapshotBarrierTransaction}, Cursor: connectors.PollingCursor{Codec: connectors.PollingCursorCodecRFC3339Nano, Type: connectors.PollingCursorTypeTimestamp, Precision: "nanosecond"},
				Ordering: connectors.PollingOrderingTuple{Watermark: connectors.PollingOrderingField{CatalogField: "updated_at", Ascending: true}, TieBreaker: connectors.PollingOrderingField{CatalogField: "id", Ascending: true, Unique: true}},
				Mutation: connectors.PollingMutationPolicy{Mutable: true, CommitOrdered: true, BoundedOverlap: true}, Identity: connectors.PollingSourceIdentity{Engine: "postgres", AccountScope: "live-target", ObjectScope: "managed-target"},
				Schema: connectors.PollingSchemaCompatibilityExactFingerprint, DeleteVisibility: connectors.PollingDeleteVisibilityTombstone, SoftDeleteField: "deleted_at", SoftDeleteAdvancesCursor: true,
				Modes: []synccontract.Mode{synccontract.ModeFullOverwrite, synccontract.ModeFullAppend, synccontract.ModeIncrementalAppend, synccontract.ModeIncrementalUpsert, synccontract.ModeIncrementalDedupe, synccontract.ModeIncrementalDedupeHistory},
			},
			Target: connectors.PollingApplyDescriptor{Executor: targetReference, MaxBatchRecords: postgresCDCRouteTargetBatchLimit, MaxBatchBytes: 1 << 20, Staging: connectors.PollingStagingReplaceSupported, StableKeyMapping: []string{"tenant", "id"}, ConditionalOrderFence: true, Transaction: connectors.PollingTransactionRequired, PartialResult: connectors.PollingPartialResultRollback, RetrySafeCloseAndInsert: true, ValidityWindow: connectors.PollingValidityWindowSupported, Strategies: []connectors.PollingApplyStrategy{connectors.PollingApplyStrategyReplace, connectors.PollingApplyStrategyAppend, connectors.PollingApplyStrategyMerge, connectors.PollingApplyStrategyDedupe, connectors.PollingApplyStrategyDedupeHistory}},
		},
		object:   connectors.PollingCatalogObject{Kind: connectors.PollingCatalogObjectRelation, NameParts: []string{"public", "managed_target"}, Columns: []string{"tenant", "id", "updated_at", "deleted_at"}},
		registry: registry,
		apply:    apply,
	}, nil
}

func postgresCDCRouteApplyWorkset(ctx context.Context, polling postgresManagedTargetPollingApply, mapping database.MappingContractV1, transaction *postgresCDCRouteTransaction) (synccontract.DownstreamAcknowledgement, error) {
	if transaction == nil {
		return synccontract.DownstreamAcknowledgement{}, errors.New("CDC route workset transaction is absent")
	}
	resolved, err := engine.PollingPreflight(ctx, polling.registry, polling.declaration, polling.object, synccontract.ModeIncrementalDedupeHistory)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	page := engine.PollingApplyPage{}
	if err := transaction.workset.ReadDelta(ctx, func(row warehouse.Row) error {
		record := connectors.Record(row)
		key, err := postgresCDCRouteRecordKey(record)
		if err != nil {
			return err
		}
		position, found := transaction.positions[key]
		if !found {
			return fmt.Errorf("sealed CDC workset record %q has no real source position", key)
		}
		page.Records = append(page.Records, engine.PollingApplyRecord{Record: record, Position: position})
		return nil
	}); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	tombstones, err := transaction.workset.Tombstones(ctx)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	for index := range tombstones {
		tombstones[index], err = mapping.MapTombstone(tombstones[index], []string{"source_tenant", "source_id"})
		if err != nil {
			return synccontract.DownstreamAcknowledgement{}, err
		}
	}
	page.Tombstones = tombstones
	if int64(len(page.Records)) != transaction.workset.Changes() || int64(len(page.Tombstones)) != transaction.workset.TombstoneCount() {
		return synccontract.DownstreamAcknowledgement{}, errors.New("sealed CDC workset count does not match target page")
	}
	return engine.ApplyPollingPage(ctx, resolved, page)
}

func clonePostgresCDCRouteRows(rows map[string]warehouse.Row) map[string]warehouse.Row {
	clone := make(map[string]warehouse.Row, len(rows))
	for key, row := range rows {
		clone[key] = clonePostgresCDCRouteRow(row)
	}
	return clone
}

func clonePostgresCDCRouteRow(record connectors.Record) warehouse.Row {
	clone := make(warehouse.Row, len(record))
	for key, value := range record {
		clone[key] = value
	}
	return clone
}

func postgresCDCRouteSortedRows(rows map[string]warehouse.Row) []warehouse.Row {
	keys := make([]string, 0, len(rows))
	for key := range rows {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]warehouse.Row, 0, len(keys))
	for _, key := range keys {
		result = append(result, clonePostgresCDCRouteRow(rows[key]))
	}
	return result
}

func postgresCDCRouteRecordKey(record connectors.Record) (string, error) {
	tenant, ok := record["source_tenant"].(string)
	if !ok || tenant == "" {
		return "", errors.New("PostgreSQL CDC record has no source_tenant key")
	}
	id, found := record["source_id"]
	if !found {
		return "", errors.New("PostgreSQL CDC record has no source_id key")
	}
	return tenant + "/" + fmt.Sprint(id), nil
}

func postgresCDCRouteEventPosition(event connectors.CDCEvent, ordinal uint64) (synccontract.CheckpointPosition, error) {
	lsn, ok := event.State["lsn"].(string)
	if !ok || lsn == "" || ordinal == 0 {
		return synccontract.CheckpointPosition{}, errors.New("PostgreSQL CDC event lacks a real LSN position")
	}
	return synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken(lsn), TieBreaker: synccontract.OpaqueToken(strconv.FormatUint(ordinal, 10))}, nil
}

func newPostgresCDCRouteContainerHarness(runtime dbtest.Runtime, endpoint string) (*dbtest.Harness, error) {
	if runtime == "" || endpoint == "" {
		return nil, errors.New("database integration requires an explicit local docker or podman endpoint")
	}
	return dbtest.New(dbtest.Config{
		Engine:                   "postgres-cdc-route",
		ContainerRuntime:         runtime,
		Image:                    postgresCDCRouteImage,
		ContainerPort:            5432,
		DataVolumePath:           "/var/lib/postgresql/data",
		ContainerEndpoint:        endpoint,
		ExpectedImageBytes:       postgresCDCRouteImageBytes,
		DockerCapacityProbeImage: "docker.io/library/busybox:1.37.0",
		ContainerArgs:            []string{"--env", "POSTGRES_DB=" + postgresCDCRouteDatabase, "--env", "POSTGRES_USER=" + postgresCDCRouteUser, "--env", "POSTGRES_HOST_AUTH_METHOD=trust"},
		EngineArgs:               []string{"-c", "wal_level=logical", "-c", "max_replication_slots=8", "-c", "max_wal_senders=8", "-c", "logical_decoding_work_mem=64kB"},
	})
}

func postgresCDCRouteConfig(t *testing.T, endpoint dbtest.Endpoint) connectors.RuntimeConfig {
	t.Helper()
	return connectors.RuntimeConfig{ProjectDir: t.TempDir(), Config: map[string]string{"host": endpoint.Host, "port": strconv.Itoa(endpoint.Port), "database": postgresCDCRouteDatabase, "username": postgresCDCRouteUser, "sslmode": "disable"}, Secrets: map[string]string{"password": t.Name()}}
}

func openPostgresCDCRouteConnection(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint) *pgx.Conn {
	t.Helper()
	config, err := pgx.ParseConfig("sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	config.Host, config.Port, config.Database, config.User = endpoint.Host, uint16(endpoint.Port), postgresCDCRouteDatabase, postgresCDCRouteUser
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatal("open isolated PostgreSQL CDC route connection")
	}
	return connection
}

func waitForPostgresCDCRoute(t *testing.T, ctx context.Context, connector native.Connector, config connectors.RuntimeConfig) {
	t.Helper()
	for {
		if err := connector.Check(ctx, config); err == nil {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("PostgreSQL CDC route engine was not reachable before the integration deadline")
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func assertPostgresCDCRoutePrerequisites(t *testing.T, ctx context.Context, source *pgx.Conn) {
	t.Helper()
	var version int
	var walLevel, workMemory string
	if err := source.QueryRow(ctx, "SELECT current_setting('server_version_num')::int, current_setting('wal_level'), current_setting('logical_decoding_work_mem')").Scan(&version, &walLevel, &workMemory); err != nil || version < 140000 || walLevel != "logical" || workMemory != "64kB" {
		t.Fatalf("PostgreSQL CDC route prerequisites = version=%d wal_level=%q logical_decoding_work_mem=%q error=%v", version, walLevel, workMemory, err)
	}
}

func startPostgresCDCRouteRead(connector native.Connector, ctx context.Context, request connectors.CDCReadRequest) <-chan error {
	done := make(chan error, 1)
	go func() {
		done <- connector.ReadCDC(ctx, request, func(connectors.CDCEvent) error {
			return errors.New("PostgreSQL CDC route bypassed its durable transaction receiver")
		})
	}()
	return done
}

func waitForPostgresCDCRouteSlot(t *testing.T, ctx context.Context, source *pgx.Conn, slot string, active bool) {
	t.Helper()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for {
		var observed bool
		err := source.QueryRow(ctx, "SELECT active FROM pg_replication_slots WHERE slot_name = $1", slot).Scan(&observed)
		if err == nil && observed == active {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("timed out waiting for PostgreSQL CDC route slot state")
		case <-deadline.C:
			t.Fatalf("PostgreSQL CDC route slot %q did not reach active=%t", slot, active)
		case <-time.After(25 * time.Millisecond):
		}
	}
}

type postgresCDCRouteSlotState struct {
	exists     bool
	active     bool
	confirmed  pglogrepl.LSN
	restartLSN pglogrepl.LSN
}

func (s postgresCDCRouteSlotState) String() string {
	return fmt.Sprintf("exists=%t active=%t confirmed_flush_lsn=%s restart_lsn=%s", s.exists, s.active, s.confirmed, s.restartLSN)
}

// observePostgresCDCRouteSlot reads the server's independent record of the
// acknowledgement. It must be sampled after each restart; source re-entry
// must neither lose the slot nor advance confirmed_flush_lsn before a new
// receiver receipt and downstream target acknowledgement exist.
func observePostgresCDCRouteSlot(t *testing.T, ctx context.Context, source *pgx.Conn, slot string) postgresCDCRouteSlotState {
	t.Helper()
	var state postgresCDCRouteSlotState
	var confirmed, restart string
	err := source.QueryRow(ctx, "SELECT active, confirmed_flush_lsn::text, restart_lsn::text FROM pg_replication_slots WHERE slot_name = $1", slot).Scan(&state.active, &confirmed, &restart)
	if errors.Is(err, pgx.ErrNoRows) {
		return state
	}
	if err != nil {
		t.Fatalf("inspect PostgreSQL CDC route replication slot %q: %v", slot, err)
	}
	var parseErr error
	if state.confirmed, parseErr = pglogrepl.ParseLSN(confirmed); parseErr != nil {
		t.Fatalf("PostgreSQL CDC route slot %q confirmed_flush_lsn = %q: %v", slot, confirmed, parseErr)
	}
	if state.restartLSN, parseErr = pglogrepl.ParseLSN(restart); parseErr != nil {
		t.Fatalf("PostgreSQL CDC route slot %q restart_lsn = %q: %v", slot, restart, parseErr)
	}
	state.exists = true
	return state
}

func assertPostgresCDCRouteUnacknowledged(ctx context.Context, source *pgx.Conn, slot string, candidate synccontract.CheckpointEnvelope) error {
	if source == nil {
		return errors.New("PostgreSQL CDC route acknowledgement source is absent")
	}
	want, err := pglogrepl.ParseLSN(string(candidate.Position.Primary))
	if err != nil {
		return err
	}
	var text string
	if err := source.QueryRow(ctx, "SELECT confirmed_flush_lsn::text FROM pg_replication_slots WHERE slot_name = $1", slot).Scan(&text); err != nil {
		return err
	}
	got, err := pglogrepl.ParseLSN(text)
	if err != nil {
		return err
	}
	if got >= want {
		return fmt.Errorf("PostgreSQL CDC route acknowledged LSN %s before durable warehouse receipt/target acknowledgement %s", got, want)
	}
	return nil
}

func assertPostgresCDCRouteAcknowledged(t *testing.T, ctx context.Context, source *pgx.Conn, slot string, checkpoint synccontract.CheckpointEnvelope) {
	t.Helper()
	want, err := pglogrepl.ParseLSN(string(checkpoint.Position.Primary))
	if err != nil {
		t.Fatal("PostgreSQL CDC route checkpoint has an invalid LSN")
	}
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for {
		var text string
		if err := source.QueryRow(ctx, "SELECT confirmed_flush_lsn::text FROM pg_replication_slots WHERE slot_name = $1", slot).Scan(&text); err == nil {
			if got, parseErr := pglogrepl.ParseLSN(text); parseErr == nil && got >= want {
				return
			}
		}
		select {
		case <-ctx.Done():
			t.Fatal("timed out waiting for PostgreSQL CDC route acknowledgement")
		case <-deadline.C:
			t.Fatal("PostgreSQL CDC route did not acknowledge its durable checkpoint")
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func assertPostgresCDCRouteStageReceipts(t *testing.T, projectDir string, wantAtLeast int) {
	t.Helper()
	var receipts, manifests int
	err := filepath.WalkDir(filepath.Join(projectDir, "state", "postgres-cdc-stage"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		switch filepath.Base(filepath.Dir(path)) {
		case "receipts":
			receipts++
		case "transactions":
			manifests++
		}
		return nil
	})
	if err != nil || receipts < wantAtLeast || manifests != 0 {
		t.Fatalf("PostgreSQL CDC route stage receipts=%d manifests=%d error=%v, want receipts >= %d and no manifests", receipts, manifests, err, wantAtLeast)
	}
}

func assertPostgresCDCRouteStopped(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("PostgreSQL CDC route reader stop error = %v, want context cancellation", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("PostgreSQL CDC route reader did not stop promptly")
	}
}

func postgresCDCRouteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
