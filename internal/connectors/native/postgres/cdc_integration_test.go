package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/durability"
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
	if _, err := data.Exec(ctx, "CREATE PUBLICATION "+quoteIdentifier(publication)+" FOR TABLE "+quoteIdentifier(table)+" WITH (publish = 'insert, update, delete')"); err != nil {
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

	firstStore := newIntegrationChangefeedStore(t)
	firstTransaction, err := connectors.NewDurableChangefeedTransaction(firstStore, firstStore)
	if err != nil {
		t.Fatal("could not create PostgreSQL CDC durable transaction")
	}
	firstEvents := make(chan connectors.CDCEvent, 8)
	firstCtx, stopFirst := context.WithCancel(ctx)
	firstDone := startCDCRead(c, firstCtx, connectors.CDCReadRequest{
		Stream:             stream,
		Config:             cfg,
		DurableTransaction: firstTransaction,
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
	checkpoint := firstStore.wait(t, ctx)
	if err := checkpoint.Validate(); err != nil || checkpoint.CommittedAt == nil {
		t.Fatal("first PostgreSQL CDC checkpoint was not durably committed")
	}
	stopFirst()
	assertCDCStopped(t, firstDone)
	waitForCDCSlot(t, ctx, data, slot, false)

	secondStore := newIntegrationChangefeedStore(t)
	secondTransaction, err := connectors.NewDurableChangefeedTransaction(secondStore, secondStore)
	if err != nil {
		t.Fatal("could not create PostgreSQL CDC resume durable transaction")
	}
	secondEvents := make(chan connectors.CDCEvent, 4)
	secondCtx, stopSecond := context.WithCancel(ctx)
	defer stopSecond()
	resumeCheckpoint := checkpoint.Clone()
	secondDone := startCDCRead(c, secondCtx, connectors.CDCReadRequest{
		Stream:             stream,
		Config:             cfg,
		Checkpoint:         &resumeCheckpoint,
		DurableTransaction: secondTransaction,
	}, secondEvents)
	waitForCDCSlot(t, ctx, data, slot, true)
	const resumedValue = "Málaga 東京"
	if _, err := data.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, value) VALUES (2, $1)", resumedValue); err != nil {
		t.Fatal("could not insert PostgreSQL CDC resume record")
	}
	resumed := nextCDCEvent(t, ctx, secondEvents)
	if resumed.Operation != "insert" || resumed.Record["id"] != 2 {
		t.Fatal("CDC restart did not resume at the next source transaction")
	}
	value, ok := resumed.Record["value"].(string)
	if !ok || !bytes.Equal([]byte(value), []byte(resumedValue)) {
		t.Fatal("CDC restart did not preserve non-ASCII text exactly")
	}
	resumedCheckpoint := secondStore.wait(t, ctx)
	if string(resumedCheckpoint.Position.Primary) == string(checkpoint.Position.Primary) {
		t.Fatal("CDC restart did not advance the durable LSN")
	}
	assertDurableCDCEventValue(t, secondStore.eventsPath, resumedValue)
	if _, err := data.Exec(ctx, "ALTER PUBLICATION "+quoteIdentifier(publication)+" SET (publish = 'insert, update')"); err != nil {
		t.Fatal("could not change PostgreSQL CDC publication scope")
	}
	if _, err := data.Exec(ctx, "INSERT INTO "+quoteIdentifier(table)+" (id, value) VALUES (3, 'scope-drift')"); err != nil {
		t.Fatal("could not insert PostgreSQL CDC scope-drift record")
	}
	assertCDCRebootstrap(t, secondDone)
	select {
	case unexpected := <-secondStore.committed:
		t.Fatalf("CDC persisted a checkpoint after publication scope drift: %s", unexpected.Position.Primary)
	default:
	}
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

func assertCDCRebootstrap(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
			t.Fatalf("PostgreSQL CDC reader did not reject publication scope drift: %v", err)
		}
		var recovery *synccontract.RebootstrapRequiredError
		if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeSourceGenerationChanged {
			t.Fatalf("PostgreSQL CDC publication scope drift recovery = %#v, want source generation changed", recovery)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("PostgreSQL CDC reader did not reject publication scope drift promptly")
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

func assertDurableCDCEventValue(t *testing.T, path, want string) {
	t.Helper()
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("could not read durable PostgreSQL CDC event log")
	}
	var committed struct {
		Events []connectors.CDCEvent `json:"events"`
	}
	if err := json.Unmarshal(payload, &committed); err != nil {
		t.Fatal("could not decode durable PostgreSQL CDC event log")
	}
	if len(committed.Events) != 1 {
		t.Fatal("durable PostgreSQL CDC event log has an unexpected event count")
	}
	got, ok := committed.Events[0].Record["value"].(string)
	if !ok || !bytes.Equal([]byte(got), []byte(want)) {
		t.Fatal("durable PostgreSQL CDC event log did not preserve non-ASCII text exactly")
	}
}

type integrationChangefeedStore struct {
	directory      string
	eventsPath     string
	checkpointPath string
	committed      chan synccontract.CheckpointEnvelope
}

func newIntegrationChangefeedStore(t *testing.T) *integrationChangefeedStore {
	t.Helper()
	directory := t.TempDir()
	return &integrationChangefeedStore{
		directory:      directory,
		eventsPath:     filepath.Join(directory, "events.jsonl"),
		checkpointPath: filepath.Join(directory, "checkpoint.json"),
		committed:      make(chan synccontract.CheckpointEnvelope, 1),
	}
}

func (s *integrationChangefeedStore) Name() string { return "postgres_cdc_conformance" }

func (s *integrationChangefeedStore) CommitChangefeedTransaction(ctx context.Context, checkpoint synccontract.CheckpointEnvelope, events []connectors.CDCEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	payload, err := json.Marshal(struct {
		Checkpoint synccontract.CheckpointEnvelope `json:"checkpoint"`
		Events     []connectors.CDCEvent           `json:"events"`
	}{Checkpoint: checkpoint, Events: events})
	if err != nil {
		return fmt.Errorf("encode downstream transaction: %w", err)
	}
	file, err := os.OpenFile(s.eventsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open downstream transaction log: %w", err)
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("write downstream transaction log: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync downstream transaction log: %w", err)
	}
	return nil
}

func (s *integrationChangefeedStore) PersistDurableChangefeedCheckpoint(ctx context.Context, checkpoint synccontract.CheckpointEnvelope) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	payload, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("encode durable checkpoint: %w", err)
	}
	payload = append(payload, '\n')
	temporary, err := os.CreateTemp(s.directory, ".checkpoint-*")
	if err != nil {
		return fmt.Errorf("create checkpoint temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		if temporary != nil {
			_ = temporary.Close()
		}
		if err != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	if _, err := temporary.Write(payload); err != nil {
		return fmt.Errorf("write durable checkpoint: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync durable checkpoint: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close durable checkpoint: %w", err)
	}
	temporary = nil
	if err := os.Rename(temporaryPath, s.checkpointPath); err != nil {
		return fmt.Errorf("replace durable checkpoint: %w", err)
	}
	if err := durability.SyncDirectory(s.directory); err != nil {
		return fmt.Errorf("sync durable checkpoint directory: %w", err)
	}
	clone := checkpoint.Clone()
	select {
	case s.committed <- clone:
	default:
	}
	return nil
}

func (s *integrationChangefeedStore) wait(t *testing.T, ctx context.Context) synccontract.CheckpointEnvelope {
	t.Helper()
	select {
	case checkpoint := <-s.committed:
		return checkpoint
	case <-ctx.Done():
		t.Fatal("timed out waiting for a durably committed PostgreSQL CDC LSN")
		return synccontract.CheckpointEnvelope{}
	}
}
