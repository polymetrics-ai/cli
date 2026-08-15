//go:build databaseintegration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/native/dbtest"
	"polymetrics.ai/internal/synccontract"
)

const (
	postgresCDCIntegrationImage      = "docker.io/library/postgres:16.10"
	postgresCDCIntegrationDatabase   = "pm_cdc"
	postgresCDCIntegrationUser       = "postgres"
	postgresCDCIntegrationImageBytes = 420 << 20

	postgresCDCIntegrationEnabledEnv  = "POLYMETRICS_DATABASE_INTEGRATION"
	postgresCDCIntegrationRuntimeEnv  = "POLYMETRICS_CONTAINER_RUNTIME"
	postgresCDCIntegrationEndpointEnv = "POLYMETRICS_CONTAINER_ENDPOINT"
)

var errPostgresCDCContainerRuntime = errors.New("database integration requires POLYMETRICS_CONTAINER_RUNTIME=docker or podman and POLYMETRICS_CONTAINER_ENDPOINT=unix:///absolute/path/to/socket; no usable explicit local container runtime is configured")

// TestPostgresPGOutputV2ContainerHarness proves the executable PostgreSQL 14+
// path against a real server. The transaction exceeds logical_decoding_work_mem
// so pgoutput has to use streamed v2 frames before its StreamCommit boundary.
func TestPostgresPGOutputV2ContainerHarness(t *testing.T) {
	if os.Getenv(postgresCDCIntegrationEnabledEnv) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the PostgreSQL pgoutput v2 proof", postgresCDCIntegrationEnabledEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	harness, err := newPostgresCDCContainerHarness(
		dbtest.Runtime(strings.TrimSpace(os.Getenv(postgresCDCIntegrationRuntimeEnv))),
		strings.TrimSpace(os.Getenv(postgresCDCIntegrationEndpointEnv)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Error("PostgreSQL CDC database test cleanup failed")
		}
		report := harness.Report()
		t.Logf("PostgreSQL CDC target image-store free bytes: before=%d after=%d", report.DiskFreeBefore, report.DiskFreeAfter)
	}()

	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("PostgreSQL CDC database container did not start: %v", err)
	}
	config := postgresCDCIntegrationConfig(t, endpoint)
	connector := New()
	waitForPostgresCDC(t, ctx, connector, config)

	source := openPostgresCDCSource(t, ctx, endpoint)
	defer func() { _ = source.Close(ctx) }()
	assertPostgresCDCPrerequisites(t, ctx, source)

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	table := "pm_cdc_stream_" + suffix
	publication := "pm_cdc_publication_" + suffix
	stream := "public." + table
	if err := validateIdentifier(table); err != nil {
		t.Fatal("generated PostgreSQL CDC table identifier is invalid")
	}
	if err := validateIdentifier(publication); err != nil {
		t.Fatal("generated PostgreSQL CDC publication identifier is invalid")
	}
	if _, err := source.Exec(ctx, "CREATE TABLE "+quoteIdentifier(table)+" (id integer PRIMARY KEY, payload text NOT NULL)"); err != nil {
		t.Fatal("could not create PostgreSQL CDC integration table")
	}
	defer func() { _, _ = source.Exec(context.Background(), "DROP TABLE IF EXISTS "+quoteIdentifier(table)) }()
	if _, err := source.Exec(ctx, "CREATE PUBLICATION "+quoteIdentifier(publication)+" FOR TABLE "+quoteIdentifier(table)); err != nil {
		t.Fatal("could not create PostgreSQL CDC integration publication")
	}
	defer func() {
		_, _ = source.Exec(context.Background(), "DROP PUBLICATION IF EXISTS "+quoteIdentifier(publication))
	}()
	config.Config["cdc_publication"] = publication

	slot, err := connector.CDCSlotName(ctx, config, stream)
	if err != nil {
		t.Fatal("could not derive PostgreSQL CDC slot name")
	}
	defer func() {
		teardownCtx, teardownCancel := context.WithTimeout(context.Background(), cdcSlotCleanupTimeout)
		defer teardownCancel()
		if err := connector.TeardownCDC(teardownCtx, config, stream); err != nil {
			t.Errorf("could not tear down PostgreSQL CDC replication slot")
		}
	}()

	firstCommitter := newCDCIntegrationCommitter()
	firstEvents := make(chan connectors.CDCEvent, 256)
	firstCtx, stopFirst := context.WithCancel(ctx)
	firstDone := startCDCRead(connector, firstCtx, connectors.CDCReadRequest{
		Stream:                     stream,
		Config:                     config,
		DurableCheckpointCommitter: firstCommitter,
	}, firstEvents)
	waitForCDCSlotState(t, ctx, source, slot, true)

	ackBeforeAbort := cdcConfirmedLSN(t, ctx, source, slot)
	aborted, err := source.Begin(ctx)
	if err != nil {
		t.Fatal("could not begin PostgreSQL CDC abort transaction")
	}
	if _, err := aborted.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, payload) VALUES (1, 'aborted')"); err != nil {
		t.Fatal("could not write PostgreSQL CDC abort record")
	}
	if err := aborted.Rollback(ctx); err != nil {
		t.Fatal("could not abort PostgreSQL CDC transaction")
	}
	assertNoCDCActivity(t, firstEvents, firstCommitter, 400*time.Millisecond)
	if got := cdcConfirmedLSN(t, ctx, source, slot); got != ackBeforeAbort {
		t.Fatalf("aborted transaction advanced acknowledged LSN from %s to %s", ackBeforeAbort, got)
	}

	streamed, err := source.Begin(ctx)
	if err != nil {
		t.Fatal("could not begin streamed PostgreSQL CDC transaction")
	}
	payload := strings.Repeat("x", 8192)
	for id := 2; id < 130; id++ {
		if _, err := streamed.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, payload) VALUES ($1, $2)", id, payload); err != nil {
			t.Fatal("could not write streamed PostgreSQL CDC record")
		}
	}
	// Logical decoding spills this large transaction and may send streamed
	// chunks before commit. Nothing may escape our private stage yet.
	assertNoCDCActivity(t, firstEvents, firstCommitter, 400*time.Millisecond)
	if err := streamed.Commit(ctx); err != nil {
		t.Fatal("could not commit streamed PostgreSQL CDC transaction")
	}
	assertCDCInsertCount(t, ctx, firstEvents, 128)
	checkpoint := firstCommitter.wait(t, ctx)
	if err := checkpoint.Validate(); err != nil || checkpoint.CommittedAt == nil {
		t.Fatal("PostgreSQL CDC committed an invalid durable checkpoint")
	}
	assertCDCAcknowledgedCheckpoint(t, ctx, source, slot, checkpoint)

	stopFirst()
	assertCDCStopped(t, firstDone)
	waitForCDCSlotState(t, ctx, source, slot, false)

	secondCommitter := newCDCIntegrationCommitter()
	secondEvents := make(chan connectors.CDCEvent, 4)
	secondCtx, stopSecond := context.WithCancel(ctx)
	resume := checkpoint.Clone()
	secondDone := startCDCRead(connector, secondCtx, connectors.CDCReadRequest{
		Stream:                     stream,
		Config:                     config,
		Checkpoint:                 &resume,
		DurableCheckpointCommitter: secondCommitter,
	}, secondEvents)
	waitForCDCSlotState(t, ctx, source, slot, true)
	if _, err := source.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, payload) VALUES (130, 'resumed')"); err != nil {
		t.Fatal("could not write PostgreSQL CDC restart record")
	}
	resumed := nextCDCEvent(t, ctx, secondEvents)
	if resumed.Operation != "insert" || resumed.Record["id"] != 130 {
		t.Fatalf("PostgreSQL CDC restart event = %#v, want the next committed insert", resumed)
	}
	resumedCheckpoint := secondCommitter.wait(t, ctx)
	if string(resumedCheckpoint.Position.Primary) == string(checkpoint.Position.Primary) {
		t.Fatal("PostgreSQL CDC restart did not advance the durable checkpoint")
	}
	stopSecond()
	assertCDCStopped(t, secondDone)
	waitForCDCSlotState(t, ctx, source, slot, false)

	if err := connector.TeardownCDC(ctx, config, stream); err != nil {
		t.Fatal("could not tear down PostgreSQL CDC replication slot")
	}
	assertCDCSlotRemoved(t, ctx, source, slot)
}

// TestPostgresGapFreeBootstrapContainerHarness holds delivery of an exported
// snapshot while a real transaction changes the same rows. The test only
// passes when the pre-barrier rows arrive in the snapshot and every committed
// during-/after-barrier mutation arrives once through pgoutput-v2.
func TestPostgresGapFreeBootstrapContainerHarness(t *testing.T) {
	if os.Getenv(postgresCDCIntegrationEnabledEnv) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the PostgreSQL gap-free bootstrap proof", postgresCDCIntegrationEnabledEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	harness, err := newPostgresCDCContainerHarness(
		dbtest.Runtime(strings.TrimSpace(os.Getenv(postgresCDCIntegrationRuntimeEnv))),
		strings.TrimSpace(os.Getenv(postgresCDCIntegrationEndpointEnv)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Error("PostgreSQL bootstrap database test cleanup failed")
		}
		report := harness.Report()
		t.Logf("PostgreSQL bootstrap target image-store free bytes: before=%d after=%d", report.DiskFreeBefore, report.DiskFreeAfter)
	}()

	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("PostgreSQL bootstrap database container did not start: %v", err)
	}
	config := postgresCDCIntegrationConfig(t, endpoint)
	connector := New()
	waitForPostgresCDC(t, ctx, connector, config)

	source := openPostgresCDCSource(t, ctx, endpoint)
	defer func() { _ = source.Close(ctx) }()
	assertPostgresCDCPrerequisites(t, ctx, source)

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	table := "pm_bootstrap_stream_" + suffix
	publication := "pm_bootstrap_publication_" + suffix
	stream := "public." + table
	if err := validateIdentifier(table); err != nil {
		t.Fatal("generated PostgreSQL bootstrap table identifier is invalid")
	}
	if err := validateIdentifier(publication); err != nil {
		t.Fatal("generated PostgreSQL bootstrap publication identifier is invalid")
	}
	if _, err := source.Exec(ctx, "CREATE TABLE "+quoteIdentifier(table)+" (id integer PRIMARY KEY, payload text NOT NULL)"); err != nil {
		t.Fatal("could not create PostgreSQL bootstrap integration table")
	}
	defer func() { _, _ = source.Exec(context.Background(), "DROP TABLE IF EXISTS "+quoteIdentifier(table)) }()
	if _, err := source.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, payload) VALUES (1, 'before-update'), (2, 'before-delete')"); err != nil {
		t.Fatal("could not seed PostgreSQL bootstrap rows")
	}
	if _, err := source.Exec(ctx, "CREATE PUBLICATION "+quoteIdentifier(publication)+" FOR TABLE "+quoteIdentifier(table)); err != nil {
		t.Fatal("could not create PostgreSQL bootstrap publication")
	}
	defer func() {
		_, _ = source.Exec(context.Background(), "DROP PUBLICATION IF EXISTS "+quoteIdentifier(publication))
	}()
	config.Config["cdc_publication"] = publication

	slot, err := connector.CDCSlotName(ctx, config, stream)
	if err != nil {
		t.Fatal("could not derive PostgreSQL bootstrap slot name")
	}
	defer func() {
		teardownCtx, teardownCancel := context.WithTimeout(context.Background(), cdcSlotCleanupTimeout)
		defer teardownCancel()
		if err := connector.TeardownCDC(teardownCtx, config, stream); err != nil {
			t.Errorf("could not tear down PostgreSQL bootstrap replication slot: %v", err)
		}
	}()

	committer := newCDCIntegrationCommitter()
	events := make(chan connectors.CDCEvent, 16)
	snapshotSeen := make(chan struct{})
	releaseSnapshot := make(chan struct{})
	observation := postgresBootstrapObservation{rows: make(map[int]string), snapshotCounts: make(map[int]int), changeCounts: make(map[string]int)}
	bootstrapCtx, stopBootstrap := context.WithCancel(ctx)
	defer stopBootstrap()
	done := make(chan error, 1)
	go func() {
		done <- connector.BootstrapCDC(bootstrapCtx, BootstrapCDCRequest{
			Stream:                     stream,
			Config:                     config,
			BatchSize:                  8,
			DurableCheckpointCommitter: committer,
			Snapshot: func(callbackCtx context.Context, page BootstrapSnapshotPage) error {
				if err := observation.applySnapshot(page); err != nil {
					return err
				}
				if !page.Final {
					return errors.New("bootstrap test expected the seeded snapshot in one bounded page")
				}
				close(snapshotSeen)
				select {
				case <-releaseSnapshot:
					return nil
				case <-callbackCtx.Done():
					return callbackCtx.Err()
				}
			},
		}, func(event connectors.CDCEvent) error {
			select {
			case events <- event:
				return nil
			case <-bootstrapCtx.Done():
				return bootstrapCtx.Err()
			}
		})
	}()

	select {
	case <-snapshotSeen:
	case err := <-done:
		t.Fatalf("PostgreSQL bootstrap failed before snapshot handover: %v", err)
	case <-ctx.Done():
		t.Fatal("PostgreSQL bootstrap did not materialize its exported snapshot")
	}
	assertCDCSlotExists(t, ctx, source, slot)

	during, err := source.Begin(ctx)
	if err != nil {
		t.Fatal("could not begin PostgreSQL concurrent bootstrap transaction")
	}
	if _, err := during.Exec(ctx, "UPDATE "+quoteIdentifier(table)+" SET payload = 'during-update' WHERE id = 1"); err != nil {
		t.Fatal("could not update PostgreSQL row during blocked bootstrap snapshot")
	}
	if _, err := during.Exec(ctx, "DELETE FROM "+quoteIdentifier(table)+" WHERE id = 2"); err != nil {
		t.Fatal("could not delete PostgreSQL row during blocked bootstrap snapshot")
	}
	if _, err := during.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, payload) VALUES (3, 'during-insert')"); err != nil {
		t.Fatal("could not insert PostgreSQL row during blocked bootstrap snapshot")
	}
	if err := during.Commit(ctx); err != nil {
		t.Fatal("could not commit PostgreSQL concurrent bootstrap transaction")
	}
	close(releaseSnapshot)

	initial := committer.wait(t, ctx)
	assertPostgresBootstrapCheckpoint(t, initial, stream, publication)
	assertCDCAcknowledgedCheckpoint(t, ctx, source, slot, initial)
	waitForCDCSlotState(t, ctx, source, slot, true)

	for range 3 {
		observation.applyChange(t, nextCDCEvent(t, ctx, events))
	}
	duringCheckpoint := committer.wait(t, ctx)
	assertCDCAcknowledgedCheckpoint(t, ctx, source, slot, duringCheckpoint)
	if observation.changeCounts["update:1"] != 1 || observation.changeCounts["delete:2"] != 1 || observation.changeCounts["insert:3"] != 1 {
		t.Fatalf("PostgreSQL concurrent bootstrap changes = %#v, want one update/delete/insert at the change boundary", observation.changeCounts)
	}

	if _, err := source.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, payload) VALUES (4, 'after-bootstrap')"); err != nil {
		t.Fatal("could not insert PostgreSQL row after bootstrap handover")
	}
	observation.applyChange(t, nextCDCEvent(t, ctx, events))
	afterCheckpoint := committer.wait(t, ctx)
	assertCDCAcknowledgedCheckpoint(t, ctx, source, slot, afterCheckpoint)
	if observation.changeCounts["insert:4"] != 1 {
		t.Fatalf("PostgreSQL after-bootstrap change count = %#v, want one insert", observation.changeCounts)
	}
	assertPostgresCDCStageReceipts(t, config.ProjectDir, 2)

	actual := readPostgresBootstrapRows(t, ctx, source, table)
	want := map[int]string{1: "during-update", 3: "during-insert", 4: "after-bootstrap"}
	if !reflect.DeepEqual(observation.rows, want) {
		t.Fatalf("combined PostgreSQL snapshot plus changefeed rows = %#v, want %#v", observation.rows, want)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("PostgreSQL source final rows = %#v, want %#v", actual, want)
	}
	if observation.snapshotCounts[1] != 1 || observation.snapshotCounts[2] != 1 || len(observation.snapshotCounts) != 2 {
		t.Fatalf("PostgreSQL bootstrap snapshot rows = %#v, want each pre-barrier row exactly once", observation.snapshotCounts)
	}

	stopBootstrap()
	assertCDCStopped(t, done)
	waitForCDCSlotState(t, ctx, source, slot, false)
}

// TestPostgresBootstrapSnapshotFailureRequiresExplicitRebootstrap proves a
// failure before the initial durable checkpoint leaves the created slot
// unacknowledged and cannot be retried as though it had a safe boundary.
func TestPostgresBootstrapSnapshotFailureRequiresExplicitRebootstrap(t *testing.T) {
	if os.Getenv(postgresCDCIntegrationEnabledEnv) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the PostgreSQL bootstrap failure proof", postgresCDCIntegrationEnabledEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	harness, err := newPostgresCDCContainerHarness(
		dbtest.Runtime(strings.TrimSpace(os.Getenv(postgresCDCIntegrationRuntimeEnv))),
		strings.TrimSpace(os.Getenv(postgresCDCIntegrationEndpointEnv)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Error("PostgreSQL bootstrap failure database test cleanup failed")
		}
	}()

	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("PostgreSQL bootstrap failure database container did not start: %v", err)
	}
	config := postgresCDCIntegrationConfig(t, endpoint)
	connector := New()
	waitForPostgresCDC(t, ctx, connector, config)
	source := openPostgresCDCSource(t, ctx, endpoint)
	defer func() { _ = source.Close(ctx) }()

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	table := "pm_bootstrap_failure_" + suffix
	publication := "pm_bootstrap_failure_pub_" + suffix
	stream := "public." + table
	if _, err := source.Exec(ctx, "CREATE TABLE "+quoteIdentifier(table)+" (id integer PRIMARY KEY, payload text NOT NULL)"); err != nil {
		t.Fatal("could not create PostgreSQL bootstrap failure integration table")
	}
	defer func() { _, _ = source.Exec(context.Background(), "DROP TABLE IF EXISTS "+quoteIdentifier(table)) }()
	if _, err := source.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, payload) VALUES (1, 'before-failure')"); err != nil {
		t.Fatal("could not seed PostgreSQL bootstrap failure row")
	}
	if _, err := source.Exec(ctx, "CREATE PUBLICATION "+quoteIdentifier(publication)+" FOR TABLE "+quoteIdentifier(table)); err != nil {
		t.Fatal("could not create PostgreSQL bootstrap failure publication")
	}
	defer func() {
		_, _ = source.Exec(context.Background(), "DROP PUBLICATION IF EXISTS "+quoteIdentifier(publication))
	}()
	config.Config["cdc_publication"] = publication

	slot, err := connector.CDCSlotName(ctx, config, stream)
	if err != nil {
		t.Fatal("could not derive PostgreSQL bootstrap failure slot name")
	}
	defer func() {
		teardownCtx, teardownCancel := context.WithTimeout(context.Background(), cdcSlotCleanupTimeout)
		defer teardownCancel()
		if err := connector.TeardownCDC(teardownCtx, config, stream); err != nil {
			t.Errorf("could not tear down PostgreSQL bootstrap failure replication slot: %v", err)
		}
	}()

	committer := newCDCIntegrationCommitter()
	wantSnapshotFailure := errors.New("snapshot warehouse persistence failed")
	err = connector.BootstrapCDC(ctx, BootstrapCDCRequest{
		Stream:                     stream,
		Config:                     config,
		BatchSize:                  8,
		DurableCheckpointCommitter: committer,
		Snapshot: func(context.Context, BootstrapSnapshotPage) error {
			return wantSnapshotFailure
		},
	}, func(connectors.CDCEvent) error { return nil })
	if !errors.Is(err, wantSnapshotFailure) {
		t.Fatalf("BootstrapCDC(snapshot failure) = %v, want the durable snapshot failure", err)
	}
	assertNoBootstrapCheckpoint(t, committer)
	assertCDCSlotExists(t, ctx, source, slot)
	assertCDCSlotNotStreaming(t, ctx, source, slot)

	snapshotCalled := false
	err = connector.BootstrapCDC(ctx, BootstrapCDCRequest{
		Stream:                     stream,
		Config:                     config,
		BatchSize:                  8,
		DurableCheckpointCommitter: committer,
		Snapshot: func(context.Context, BootstrapSnapshotPage) error {
			snapshotCalled = true
			return nil
		},
	}, func(connectors.CDCEvent) error { return nil })
	if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
		t.Fatalf("BootstrapCDC(uncheckpointed slot) = %v, want explicit rebootstrap", err)
	}
	if snapshotCalled {
		t.Fatal("BootstrapCDC reused an uncheckpointed slot and called snapshot delivery")
	}
	assertNoBootstrapCheckpoint(t, committer)
	assertCDCSlotNotStreaming(t, ctx, source, slot)

	if err := connector.TeardownCDC(ctx, config, stream); err != nil {
		t.Fatal("could not explicitly tear down the snapshot-failed PostgreSQL bootstrap slot")
	}
	assertCDCSlotRemoved(t, ctx, source, slot)

	checkpointFailure := errors.New("durable bootstrap checkpoint persistence failed")
	failingCommitter := &bootstrapFailingCommitter{err: checkpointFailure}
	snapshotCalls := 0
	err = connector.BootstrapCDC(ctx, BootstrapCDCRequest{
		Stream:                     stream,
		Config:                     config,
		BatchSize:                  8,
		DurableCheckpointCommitter: failingCommitter,
		Snapshot: func(context.Context, BootstrapSnapshotPage) error {
			snapshotCalls++
			return nil
		},
	}, func(connectors.CDCEvent) error { return nil })
	if !errors.Is(err, checkpointFailure) || snapshotCalls == 0 || failingCommitter.calls != 1 {
		t.Fatalf("BootstrapCDC(checkpoint failure) = %v, snapshot calls=%d, checkpoint calls=%d; want a snapshot then one failed durable checkpoint", err, snapshotCalls, failingCommitter.calls)
	}
	assertCDCSlotExists(t, ctx, source, slot)
	assertCDCSlotNotStreaming(t, ctx, source, slot)

	snapshotCalled = false
	err = connector.BootstrapCDC(ctx, BootstrapCDCRequest{
		Stream:                     stream,
		Config:                     config,
		BatchSize:                  8,
		DurableCheckpointCommitter: committer,
		Snapshot: func(context.Context, BootstrapSnapshotPage) error {
			snapshotCalled = true
			return nil
		},
	}, func(connectors.CDCEvent) error { return nil })
	if !errors.Is(err, synccontract.ErrRebootstrapRequired) || snapshotCalled {
		t.Fatalf("BootstrapCDC(checkpoint-failed slot) = %v, snapshot called=%t; want explicit rebootstrap before snapshot reuse", err, snapshotCalled)
	}
}

type postgresBootstrapObservation struct {
	rows           map[int]string
	snapshotCounts map[int]int
	changeCounts   map[string]int
}

type bootstrapFailingCommitter struct {
	err   error
	calls int
}

func (c *bootstrapFailingCommitter) CommitDurableChangefeedCheckpoint(context.Context, synccontract.CheckpointEnvelope) error {
	c.calls++
	return c.err
}

func (o *postgresBootstrapObservation) applySnapshot(page BootstrapSnapshotPage) error {
	if page.Source.Engine != "postgres" || page.Source.ObjectScope == "" || page.SnapshotBarrier.Kind != cdcSnapshotBarrierKind || page.SchemaFingerprint == "" {
		return errors.New("bootstrap snapshot page did not bind a PostgreSQL source, barrier, and schema fingerprint")
	}
	if err := page.CandidateCheckpoint.Validate(); err != nil {
		return fmt.Errorf("bootstrap snapshot page has no stageable candidate checkpoint: %w", err)
	}
	barrier, ok, err := postgresBootstrapBarrierFromCheckpoint(&page.CandidateCheckpoint)
	if err != nil || !ok || barrier.Relation != page.Source.ObjectScope || barrier.SchemaFingerprint != page.SchemaFingerprint || string(page.SnapshotBarrier.Token) != string(page.CandidateCheckpoint.SnapshotBarrier.Token) {
		return errors.New("bootstrap snapshot page candidate does not bind its durable barrier and source observation")
	}
	for _, record := range page.Records {
		id, payload, err := postgresBootstrapRecord(record)
		if err != nil {
			return err
		}
		o.snapshotCounts[id]++
		if o.snapshotCounts[id] != 1 {
			return fmt.Errorf("bootstrap snapshot delivered id %d more than once", id)
		}
		o.rows[id] = payload
	}
	return nil
}

func (o *postgresBootstrapObservation) applyChange(t *testing.T, event connectors.CDCEvent) {
	t.Helper()
	id, err := postgresBootstrapRecordID(event.Record)
	if err != nil {
		t.Fatalf("PostgreSQL CDC event has no usable primary key: %#v (%v)", event, err)
	}
	key := event.Operation + ":" + strconv.Itoa(id)
	o.changeCounts[key]++
	if o.changeCounts[key] != 1 {
		t.Fatalf("PostgreSQL CDC delivered duplicate change %q: %#v", key, event)
	}
	switch event.Operation {
	case "insert", "update":
		_, payload, err := postgresBootstrapRecord(event.Record)
		if err != nil {
			t.Fatalf("PostgreSQL CDC %s event cannot update combined state: %#v (%v)", event.Operation, event, err)
		}
		o.rows[id] = payload
	case "delete":
		delete(o.rows, id)
	default:
		t.Fatalf("PostgreSQL CDC bootstrap event operation = %q, want insert, update, or explicit delete", event.Operation)
	}
}

func postgresBootstrapRecord(record connectors.Record) (int, string, error) {
	id, err := postgresBootstrapRecordID(record)
	if err != nil {
		return 0, "", err
	}
	payload, ok := record["payload"].(string)
	if !ok {
		return 0, "", errors.New("record payload is not a string")
	}
	return id, payload, nil
}

func postgresBootstrapRecordID(record connectors.Record) (int, error) {
	value, ok := record["id"]
	if !ok {
		return 0, errors.New("record id is missing")
	}
	switch id := value.(type) {
	case int:
		return id, nil
	case int8:
		return int(id), nil
	case int16:
		return int(id), nil
	case int32:
		return int(id), nil
	case int64:
		return int(id), nil
	default:
		return 0, fmt.Errorf("record id has unsupported type %T", value)
	}
}

func assertPostgresBootstrapCheckpoint(t *testing.T, checkpoint synccontract.CheckpointEnvelope, stream, publication string) {
	t.Helper()
	if err := checkpoint.Validate(); err != nil || checkpoint.CommittedAt == nil {
		t.Fatalf("PostgreSQL bootstrap committed an invalid initial durable checkpoint: %v", err)
	}
	barrier, ok, err := postgresBootstrapBarrierFromCheckpoint(&checkpoint)
	if err != nil || !ok || barrier.Relation != stream || barrier.Publication != publication || barrier.SystemID == "" || barrier.Timeline <= 0 || barrier.SchemaFingerprint == "" {
		t.Fatalf("PostgreSQL bootstrap checkpoint barrier = (%#v, %t, %v), want exact source/publication/schema metadata", barrier, ok, err)
	}
	position, err := pglogrepl.ParseLSN(string(checkpoint.Position.Primary))
	if err != nil || position.String() != barrier.InitialLSN {
		t.Fatalf("PostgreSQL bootstrap checkpoint LSN = %q, barrier = %q", checkpoint.Position.Primary, barrier.InitialLSN)
	}
}

func assertCDCSlotExists(t *testing.T, ctx context.Context, source *pgx.Conn, slot string) {
	t.Helper()
	var exists bool
	if err := source.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = $1)", slot).Scan(&exists); err != nil || !exists {
		t.Fatalf("PostgreSQL bootstrap did not retain its slot while snapshot delivery was blocked: exists=%t err=%v", exists, err)
	}
}

func assertNoBootstrapCheckpoint(t *testing.T, committer *cdcIntegrationCommitter) {
	t.Helper()
	timer := time.NewTimer(250 * time.Millisecond)
	defer timer.Stop()
	select {
	case checkpoint := <-committer.committed:
		t.Fatalf("PostgreSQL bootstrap advanced a durable checkpoint before snapshot persistence: %#v", checkpoint)
	case <-timer.C:
	}
}

func assertCDCSlotNotStreaming(t *testing.T, ctx context.Context, source *pgx.Conn, slot string) {
	t.Helper()
	var active bool
	if err := source.QueryRow(ctx, "SELECT active FROM pg_replication_slots WHERE slot_name = $1", slot).Scan(&active); err != nil {
		t.Fatal("could not inspect PostgreSQL bootstrap slot replication state")
	}
	if active {
		t.Fatal("PostgreSQL bootstrap began replication before a durable checkpoint existed")
	}
}

func assertPostgresCDCStageReceipts(t *testing.T, projectDir string, wantAtLeast int) {
	t.Helper()
	var receipts, manifests int
	err := filepath.WalkDir(filepath.Join(projectDir, "state", cdcStageDirectory), func(path string, entry fs.DirEntry, walkErr error) error {
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
	if err != nil {
		t.Fatalf("could not inspect PostgreSQL CDC transaction-stage artifacts: %v", err)
	}
	if receipts < wantAtLeast {
		t.Fatalf("PostgreSQL CDC durable receipt files = %d, want at least %d", receipts, wantAtLeast)
	}
	if manifests != 0 {
		t.Fatalf("PostgreSQL CDC left private transaction manifests after durable receipt = %d", manifests)
	}
}

func readPostgresBootstrapRows(t *testing.T, ctx context.Context, source *pgx.Conn, table string) map[int]string {
	t.Helper()
	rows, err := source.Query(ctx, "SELECT id, payload FROM "+quoteIdentifier(table)+" ORDER BY id")
	if err != nil {
		t.Fatal("could not read final PostgreSQL bootstrap rows")
	}
	defer rows.Close()
	actual := make(map[int]string)
	for rows.Next() {
		var id int
		var payload string
		if err := rows.Scan(&id, &payload); err != nil {
			t.Fatal("could not scan final PostgreSQL bootstrap row")
		}
		actual[id] = payload
	}
	if err := rows.Err(); err != nil {
		t.Fatal("could not finish reading final PostgreSQL bootstrap rows")
	}
	return actual
}

func newPostgresCDCContainerHarness(runtime dbtest.Runtime, endpoint string) (*dbtest.Harness, error) {
	if runtime == "" || endpoint == "" {
		return nil, errPostgresCDCContainerRuntime
	}
	harness, err := dbtest.New(dbtest.Config{
		Engine:                   "postgres-cdc",
		ContainerRuntime:         runtime,
		Image:                    postgresCDCIntegrationImage,
		ContainerPort:            5432,
		DataVolumePath:           "/var/lib/postgresql/data",
		ContainerEndpoint:        endpoint,
		ExpectedImageBytes:       postgresCDCIntegrationImageBytes,
		DockerCapacityProbeImage: "docker.io/library/busybox:1.37.0",
		ContainerArgs: []string{
			"--env", "POSTGRES_DB=" + postgresCDCIntegrationDatabase,
			"--env", "POSTGRES_USER=" + postgresCDCIntegrationUser,
			"--env", "POSTGRES_HOST_AUTH_METHOD=trust",
		},
		EngineArgs: []string{
			"-c", "wal_level=logical",
			"-c", "max_replication_slots=8",
			"-c", "max_wal_senders=8",
			"-c", "logical_decoding_work_mem=64kB",
		},
	})
	if err != nil {
		return nil, errPostgresCDCContainerRuntime
	}
	return harness, nil
}

func postgresCDCIntegrationConfig(t *testing.T, endpoint dbtest.Endpoint) connectors.RuntimeConfig {
	t.Helper()
	return connectors.RuntimeConfig{
		ProjectDir: t.TempDir(),
		Config: map[string]string{
			"host":     endpoint.Host,
			"port":     strconv.Itoa(endpoint.Port),
			"database": postgresCDCIntegrationDatabase,
			"username": postgresCDCIntegrationUser,
			"sslmode":  "disable",
		},
		// Trust authentication ignores this generated, non-secret test value;
		// live connector validation still requires a nonempty password field.
		Secrets: map[string]string{"password": t.Name()},
	}
}

func waitForPostgresCDC(t *testing.T, ctx context.Context, connector Connector, config connectors.RuntimeConfig) {
	t.Helper()
	for {
		if err := connector.Check(ctx, config); err == nil {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("PostgreSQL CDC engine was not reachable before the integration deadline")
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func openPostgresCDCSource(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint) *pgx.Conn {
	t.Helper()
	config, err := pgx.ParseConfig("sslmode=disable")
	if err != nil {
		t.Fatal("could not configure the isolated PostgreSQL CDC source")
	}
	config.Host = endpoint.Host
	config.Port = uint16(endpoint.Port)
	config.Database = postgresCDCIntegrationDatabase
	config.User = postgresCDCIntegrationUser
	source, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatal("could not open the isolated PostgreSQL CDC source")
	}
	return source
}

func assertPostgresCDCPrerequisites(t *testing.T, ctx context.Context, source *pgx.Conn) {
	t.Helper()
	var version int
	var walLevel, workMem string
	if err := source.QueryRow(ctx, "SELECT current_setting('server_version_num')::int, current_setting('wal_level'), current_setting('logical_decoding_work_mem')").Scan(&version, &walLevel, &workMem); err != nil {
		t.Fatal("could not inspect PostgreSQL CDC prerequisites")
	}
	if version < 140000 || walLevel != "logical" || workMem != "64kB" {
		t.Fatalf("PostgreSQL CDC prerequisites = version=%d wal_level=%q logical_decoding_work_mem=%q", version, walLevel, workMem)
	}
}

func startCDCRead(connector Connector, ctx context.Context, req connectors.CDCReadRequest, events chan<- connectors.CDCEvent) <-chan error {
	receiver := cdcIntegrationTransactionReceiver{events: events}
	req.TransactionReceiver = receiver
	done := make(chan error, 1)
	go func() {
		done <- connector.ReadCDC(ctx, req, func(connectors.CDCEvent) error {
			return errors.New("PostgreSQL live CDC bypassed its durable transaction receiver")
		})
	}()
	return done
}

type cdcIntegrationTransactionReceiver struct {
	events chan<- connectors.CDCEvent
}

func (r cdcIntegrationTransactionReceiver) ReceiveCDCTransaction(ctx context.Context, transaction connectors.CDCTransaction) (connectors.CDCTransactionReceipt, error) {
	if err := transaction.StreamEvents(ctx, func(event connectors.CDCEvent) error {
		select {
		case r.events <- event:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}); err != nil {
		return connectors.CDCTransactionReceipt{}, err
	}
	digest := sha256.Sum256([]byte(transaction.ID()))
	return connectors.NewCDCTransactionReceipt("postgres-live-"+hex.EncodeToString(digest[:]), "warehouse:postgres-cdc-integration", time.Now().UTC())
}

func assertNoCDCActivity(t *testing.T, events <-chan connectors.CDCEvent, committer *cdcIntegrationCommitter, duration time.Duration) {
	t.Helper()
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case event := <-events:
		t.Fatalf("PostgreSQL CDC delivered an event before StreamCommit or after StreamAbort: %#v", event)
	case checkpoint := <-committer.committed:
		t.Fatalf("PostgreSQL CDC checkpointed before StreamCommit or after StreamAbort: %#v", checkpoint)
	case <-timer.C:
	}
}

func assertCDCInsertCount(t *testing.T, ctx context.Context, events <-chan connectors.CDCEvent, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		event := nextCDCEvent(t, ctx, events)
		if event.Operation != "insert" {
			t.Fatalf("PostgreSQL CDC event %d operation = %q, want insert", i, event.Operation)
		}
	}
}

func nextCDCEvent(t *testing.T, ctx context.Context, events <-chan connectors.CDCEvent) connectors.CDCEvent {
	t.Helper()
	select {
	case event := <-events:
		return event
	case <-ctx.Done():
		t.Fatal("timed out waiting for PostgreSQL CDC event")
		return connectors.CDCEvent{}
	}
}

func waitForCDCSlotState(t *testing.T, ctx context.Context, source *pgx.Conn, slot string, active bool) {
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
		case <-deadline.C:
			t.Fatalf("PostgreSQL CDC replication slot %q did not reach active=%t", slot, active)
		case <-ctx.Done():
			t.Fatal("timed out waiting for PostgreSQL CDC replication slot lifecycle state")
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func cdcConfirmedLSN(t *testing.T, ctx context.Context, source *pgx.Conn, slot string) pglogrepl.LSN {
	t.Helper()
	var text string
	if err := source.QueryRow(ctx, "SELECT confirmed_flush_lsn::text FROM pg_replication_slots WHERE slot_name = $1", slot).Scan(&text); err != nil {
		t.Fatal("could not inspect PostgreSQL CDC acknowledged LSN")
	}
	position, err := pglogrepl.ParseLSN(text)
	if err != nil {
		t.Fatal("PostgreSQL CDC slot reported an invalid acknowledged LSN")
	}
	return position
}

func assertCDCAcknowledgedCheckpoint(t *testing.T, ctx context.Context, source *pgx.Conn, slot string, checkpoint synccontract.CheckpointEnvelope) {
	t.Helper()
	want, err := pglogrepl.ParseLSN(string(checkpoint.Position.Primary))
	if err != nil {
		t.Fatal("PostgreSQL CDC checkpoint had an invalid LSN")
	}
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for {
		if got := cdcConfirmedLSN(t, ctx, source, slot); got >= want {
			return
		}
		select {
		case <-deadline.C:
			t.Fatal("PostgreSQL CDC did not acknowledge the LSN after durable checkpoint persistence")
		case <-ctx.Done():
			t.Fatal("timed out waiting for PostgreSQL CDC acknowledgement")
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func assertCDCStopped(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("PostgreSQL CDC reader did not stop when its context was cancelled: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("PostgreSQL CDC reader did not stop promptly")
	}
}

func assertCDCSlotRemoved(t *testing.T, ctx context.Context, source *pgx.Conn, slot string) {
	t.Helper()
	var exists bool
	if err := source.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = $1)", slot).Scan(&exists); err != nil || exists {
		t.Fatal("PostgreSQL CDC teardown left a replication slot behind")
	}
}

type cdcIntegrationCommitter struct {
	committed chan synccontract.CheckpointEnvelope
}

func newCDCIntegrationCommitter() *cdcIntegrationCommitter {
	return &cdcIntegrationCommitter{committed: make(chan synccontract.CheckpointEnvelope, 4)}
}

func (c *cdcIntegrationCommitter) CommitDurableChangefeedCheckpoint(_ context.Context, candidate synccontract.CheckpointEnvelope) error {
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement("postgres_pgoutput_v2_integration", time.Now().UTC())
	if err != nil {
		return err
	}
	return synccontract.CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(committed synccontract.CheckpointEnvelope) error {
		clone := committed.Clone()
		select {
		case c.committed <- clone:
		default:
		}
		return nil
	})
}

func (c *cdcIntegrationCommitter) wait(t *testing.T, ctx context.Context) synccontract.CheckpointEnvelope {
	t.Helper()
	select {
	case checkpoint := <-c.committed:
		return checkpoint
	case <-ctx.Done():
		t.Fatal("timed out waiting for a durably committed PostgreSQL CDC LSN")
		return synccontract.CheckpointEnvelope{}
	}
}
