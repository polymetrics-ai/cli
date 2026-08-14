//go:build databaseintegration

package postgres

import (
	"context"
	"errors"
	"os"
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
	done := make(chan error, 1)
	go func() {
		done <- connector.ReadCDC(ctx, req, func(event connectors.CDCEvent) error {
			select {
			case events <- event:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	return done
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
