package postgres

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

// TestLogicalReplicationResumesAndCleansSlot is deliberately an opt-in live
// conformance test. It exercises PostgreSQL's wire replication protocol and
// pgoutput rather than mocking either one. The harness supplies individual
// connection fields through its environment; it never constructs or reports a
// connection string.
func TestLogicalReplicationResumesAndCleansSlot(t *testing.T) {
	cfg := postgresCDCIntegrationConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	conn, err := resolveConfig(cfg)
	if err != nil {
		t.Fatal("CDC integration configuration is invalid")
	}
	data, err := pgx.Connect(ctx, conn.dsn())
	if err != nil {
		t.Fatal("could not connect to PostgreSQL CDC integration source")
	}
	defer func() { _ = data.Close(ctx) }()

	var walLevel string
	if err := data.QueryRow(ctx, "SHOW wal_level").Scan(&walLevel); err != nil || walLevel != "logical" {
		t.Fatal("PostgreSQL CDC integration source must set wal_level=logical")
	}

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	table := "pm_cdc_it_" + suffix
	publication := "pm_cdc_pub_" + suffix
	stream := "public." + table
	if err := validateIdentifier(table); err != nil {
		t.Fatal("generated CDC integration table name is invalid")
	}
	if err := validateIdentifier(publication); err != nil {
		t.Fatal("generated CDC integration publication name is invalid")
	}
	cfg.Config["cdc_publication"] = publication

	if _, err := data.Exec(ctx, "CREATE TABLE "+quoteIdentifier(table)+" (id integer PRIMARY KEY, value text NOT NULL)"); err != nil {
		t.Fatal("could not create PostgreSQL CDC integration table")
	}
	defer func() {
		_, _ = data.Exec(context.Background(), "DROP TABLE IF EXISTS "+quoteIdentifier(table))
	}()
	if _, err := data.Exec(ctx, "CREATE PUBLICATION "+quoteIdentifier(publication)+" FOR TABLE "+quoteIdentifier(table)); err != nil {
		t.Fatal("could not create PostgreSQL CDC integration publication")
	}
	defer func() {
		_, _ = data.Exec(context.Background(), "DROP PUBLICATION IF EXISTS "+quoteIdentifier(publication))
	}()

	c := New()
	slot, err := c.CDCSlotName(ctx, cfg, stream)
	if err != nil {
		t.Fatal("could not derive PostgreSQL CDC slot name")
	}
	defer func() {
		teardownCtx, teardownCancel := context.WithTimeout(context.Background(), cdcSlotCleanupTimeout)
		defer teardownCancel()
		_ = c.TeardownCDC(teardownCtx, cfg, stream)
	}()

	firstCommitter := newIntegrationCheckpointCommitter()
	firstEvents := make(chan connectors.CDCEvent, 8)
	firstCtx, stopFirst := context.WithCancel(ctx)
	firstDone := startCDCRead(c, firstCtx, connectors.CDCReadRequest{
		Stream:                     stream,
		Config:                     cfg,
		DurableCheckpointCommitter: firstCommitter,
	}, firstEvents)
	waitForCDCSlot(t, ctx, data, slot, true)

	tx, err := data.Begin(ctx)
	if err != nil {
		t.Fatal("could not start PostgreSQL CDC integration transaction")
	}
	if _, err := tx.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, value) VALUES (1, 'first')"); err != nil {
		t.Fatal("could not insert PostgreSQL CDC integration record")
	}
	if _, err := tx.Exec(ctx, "UPDATE "+quoteIdentifier(table)+" SET value = 'updated' WHERE id = 1"); err != nil {
		t.Fatal("could not update PostgreSQL CDC integration record")
	}
	if _, err := tx.Exec(ctx, "DELETE FROM "+quoteIdentifier(table)+" WHERE id = 1"); err != nil {
		t.Fatal("could not delete PostgreSQL CDC integration record")
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal("could not commit PostgreSQL CDC integration transaction")
	}

	assertCDCOperations(t, ctx, firstEvents, []string{"insert", "update", "delete"})
	checkpoint := firstCommitter.wait(t, ctx)
	if err := checkpoint.Validate(); err != nil || checkpoint.CommittedAt == nil {
		t.Fatal("first PostgreSQL CDC checkpoint was not durably committed")
	}
	stopFirst()
	assertCDCStopped(t, firstDone)
	waitForCDCSlot(t, ctx, data, slot, false)

	secondCommitter := newIntegrationCheckpointCommitter()
	secondEvents := make(chan connectors.CDCEvent, 4)
	secondCtx, stopSecond := context.WithCancel(ctx)
	resumeCheckpoint := checkpoint.Clone()
	secondDone := startCDCRead(c, secondCtx, connectors.CDCReadRequest{
		Stream:                     stream,
		Config:                     cfg,
		Checkpoint:                 &resumeCheckpoint,
		DurableCheckpointCommitter: secondCommitter,
	}, secondEvents)
	waitForCDCSlot(t, ctx, data, slot, true)
	if _, err := data.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, value) VALUES (2, 'resumed')"); err != nil {
		t.Fatal("could not insert PostgreSQL CDC resume record")
	}
	resumed := nextCDCEvent(t, ctx, secondEvents)
	if resumed.Operation != "insert" || resumed.Record["id"] != 2 {
		t.Fatal("CDC restart did not resume at the next source transaction")
	}
	resumedCheckpoint := secondCommitter.wait(t, ctx)
	if string(resumedCheckpoint.Position.Primary) == string(checkpoint.Position.Primary) {
		t.Fatal("CDC restart did not advance the durable LSN")
	}
	stopSecond()
	assertCDCStopped(t, secondDone)
	waitForCDCSlot(t, ctx, data, slot, false)

	if err := c.TeardownCDC(ctx, cfg, stream); err != nil {
		t.Fatal("could not tear down PostgreSQL CDC replication slot")
	}
	assertCDCSlotRemoved(t, ctx, data, slot)
}

func postgresCDCIntegrationConfig(t *testing.T) connectors.RuntimeConfig {
	t.Helper()
	if os.Getenv("POLYMETRICS_INTEGRATION") != "1" {
		t.Skip("set POLYMETRICS_INTEGRATION=1 to run the live PostgreSQL CDC conformance test")
	}
	lookup := func(name string) string {
		value := strings.TrimSpace(os.Getenv(name))
		if value == "" {
			t.Skip("PostgreSQL CDC integration environment is incomplete")
		}
		return value
	}
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"host":     lookup("POLYMETRICS_POSTGRES_CDC_HOST"),
			"port":     lookup("POLYMETRICS_POSTGRES_CDC_PORT"),
			"database": lookup("POLYMETRICS_POSTGRES_CDC_DATABASE"),
			"username": lookup("POLYMETRICS_POSTGRES_CDC_USERNAME"),
			"sslmode":  "disable",
		},
		Secrets: map[string]string{"password": lookup("POLYMETRICS_POSTGRES_CDC_PASSWORD")},
	}
}

func startCDCRead(c Connector, ctx context.Context, req connectors.CDCReadRequest, events chan<- connectors.CDCEvent) <-chan error {
	done := make(chan error, 1)
	go func() {
		done <- c.ReadCDC(ctx, req, func(event connectors.CDCEvent) error {
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

func assertCDCOperations(t *testing.T, ctx context.Context, events <-chan connectors.CDCEvent, want []string) {
	t.Helper()
	for _, operation := range want {
		event := nextCDCEvent(t, ctx, events)
		if event.Operation != operation {
			t.Fatal("PostgreSQL CDC event ordering did not follow the committed source transaction")
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

func assertCDCStopped(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal("PostgreSQL CDC reader did not stop when its context was cancelled")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("PostgreSQL CDC reader did not stop promptly")
	}
}

func waitForCDCSlot(t *testing.T, ctx context.Context, data *pgx.Conn, slot string, active bool) {
	t.Helper()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for {
		var observed bool
		err := data.QueryRow(ctx, "SELECT active FROM pg_replication_slots WHERE slot_name = $1", slot).Scan(&observed)
		if err == nil && observed == active {
			return
		}
		select {
		case <-deadline.C:
			t.Fatal("PostgreSQL CDC replication slot did not reach its expected lifecycle state")
		case <-ctx.Done():
			t.Fatal("timed out waiting for PostgreSQL CDC replication slot lifecycle state")
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func assertCDCSlotRemoved(t *testing.T, ctx context.Context, data *pgx.Conn, slot string) {
	t.Helper()
	var exists bool
	if err := data.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_replication_slots WHERE slot_name = $1)", slot).Scan(&exists); err != nil || exists {
		t.Fatal("PostgreSQL CDC teardown left a replication slot behind")
	}
}

type integrationCheckpointCommitter struct {
	committed chan synccontract.CheckpointEnvelope
}

func newIntegrationCheckpointCommitter() *integrationCheckpointCommitter {
	return &integrationCheckpointCommitter{committed: make(chan synccontract.CheckpointEnvelope, 1)}
}

func (c *integrationCheckpointCommitter) CommitDurableChangefeedCheckpoint(_ context.Context, candidate synccontract.CheckpointEnvelope) error {
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement("postgres_cdc_conformance", time.Now().UTC())
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

func (c *integrationCheckpointCommitter) wait(t *testing.T, ctx context.Context) synccontract.CheckpointEnvelope {
	t.Helper()
	select {
	case checkpoint := <-c.committed:
		return checkpoint
	case <-ctx.Done():
		t.Fatal("timed out waiting for a durably committed PostgreSQL CDC LSN")
		return synccontract.CheckpointEnvelope{}
	}
}
