//go:build databaseintegration

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors/native/dbtest"
	"polymetrics.ai/internal/warehouse"
)

const (
	postgresCDCBinaryImage      = "docker.io/library/postgres:16.10"
	postgresCDCBinaryDatabase   = "pm_cli_cdc"
	postgresCDCBinaryUser       = "postgres"
	postgresCDCBinaryImageBytes = 420 << 20
)

// TestPMBinaryDispatchesPostgresChangeCaptureToWarehouse proves the shipped
// executable call chain: cmd/pm -> cli.Run -> App.RunETL -> PostgreSQL ReadCDC
// -> the connection-owned warehouse. The source transaction is committed only
// after the fresh binary has an active pgoutput slot, and success requires the
// exact row, full checkpoint, native stage receipt, and acknowledged source LSN.
func TestPMBinaryDispatchesPostgresChangeCaptureToWarehouse(t *testing.T) {
	if os.Getenv("POLYMETRICS_DATABASE_INTEGRATION") != "1" {
		t.Skip("database integration skipped: set POLYMETRICS_DATABASE_INTEGRATION=1 to run the PostgreSQL binary dispatch proof")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	harness, err := dbtest.New(dbtest.Config{
		Engine:                   "postgres-cdc-binary",
		ContainerRuntime:         dbtest.Runtime(strings.TrimSpace(os.Getenv("POLYMETRICS_CONTAINER_RUNTIME"))),
		Image:                    postgresCDCBinaryImage,
		ContainerPort:            5432,
		DataVolumePath:           "/var/lib/postgresql/data",
		ContainerEndpoint:        strings.TrimSpace(os.Getenv("POLYMETRICS_CONTAINER_ENDPOINT")),
		ExpectedImageBytes:       postgresCDCBinaryImageBytes,
		DockerCapacityProbeImage: "docker.io/library/busybox:1.37.0",
		ContainerArgs: []string{
			"--env", "POSTGRES_DB=" + postgresCDCBinaryDatabase,
			"--env", "POSTGRES_USER=" + postgresCDCBinaryUser,
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
		t.Fatal("PostgreSQL binary database test requires an explicit usable local container runtime")
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Error("PostgreSQL binary database test cleanup failed")
		}
	}()

	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("PostgreSQL binary database container did not start: %v", err)
	}
	source := openPostgresBinarySource(t, ctx, endpoint)
	defer func() { _ = source.Close(ctx) }()

	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	table := "pm_cli_cdc_" + suffix
	publication := "pm_cli_pub_" + suffix
	stream := "public." + table
	if _, err := source.Exec(ctx, "CREATE TABLE "+pgx.Identifier{table}.Sanitize()+" (id integer PRIMARY KEY, payload text NOT NULL)"); err != nil {
		t.Fatal("could not create PostgreSQL binary CDC table")
	}
	if _, err := source.Exec(ctx, "CREATE PUBLICATION "+pgx.Identifier{publication}.Sanitize()+" FOR TABLE "+pgx.Identifier{table}.Sanitize()); err != nil {
		t.Fatal("could not create PostgreSQL binary CDC publication")
	}

	binary := buildTransportPM(t)
	sha, size := transportBinaryIdentity(t, binary)
	t.Logf("fresh PostgreSQL CDC pm binary sha256=%s size_bytes=%d", sha, size)
	root := filepath.Join(t.TempDir(), "project")
	t.Setenv("PM_POSTGRES_CDC_TEST_PASSWORD", "local-integration-test-value")
	mustRunPostgresBinary(t, binary, "init", "--root", root, "--json")
	mustRunPostgresBinary(t, binary,
		"credentials", "add", "postgres-source",
		"--connector", "postgres",
		"--config", "host="+endpoint.Host,
		"--config", "port="+strconv.Itoa(endpoint.Port),
		"--config", "database="+postgresCDCBinaryDatabase,
		"--config", "username="+postgresCDCBinaryUser,
		"--config", "sslmode=disable",
		"--config", "cdc_publication="+publication,
		"--from-env", "password=PM_POSTGRES_CDC_TEST_PASSWORD",
		"--root", root,
		"--json",
	)
	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	mustRunPostgresBinary(t, binary,
		"credentials", "add", "warehouse-local",
		"--connector", "warehouse",
		"--config", "path="+warehouseDir,
		"--root", root,
		"--json",
	)
	connectionOutput := mustRunPostgresBinary(t, binary,
		"connections", "create", "postgres-cdc-binary",
		"--source", "postgres:postgres-source",
		"--destination", "warehouse:warehouse-local",
		"--stream", stream,
		"--sync-mode", "change_capture",
		"--primary-key", "id",
		"--table", "postgres_cdc_rows",
		"--root", root,
		"--json",
	)
	connectionID := transportConnectionIDFromOutput(t, connectionOutput)

	var stdout, stderr bytes.Buffer
	command := exec.Command(binary,
		"etl", "run",
		"--connection", "postgres-cdc-binary",
		"--stream", stream,
		"--batch-size", "1",
		"--root", root,
		"--json",
	)
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		t.Fatalf("start fresh pm PostgreSQL CDC process: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	stopped := false
	defer func() {
		if stopped {
			return
		}
		stopPostgresBinaryProcess(t, command, done)
	}()

	waitForPostgresBinaryCondition(t, ctx, done, &stdout, &stderr, "active PostgreSQL logical-replication slot", func() bool {
		var active bool
		return source.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_replication_slots WHERE active AND plugin = 'pgoutput')").Scan(&active) == nil && active
	})
	if _, err := source.Exec(ctx, "INSERT INTO "+pgx.Identifier{table}.Sanitize()+" (id, payload) VALUES (901, 'binary-dispatch')"); err != nil {
		t.Fatal("could not commit PostgreSQL binary CDC transaction")
	}
	var committedLSN string
	if err := source.QueryRow(ctx, "SELECT pg_current_wal_lsn()::text").Scan(&committedLSN); err != nil || committedLSN == "" {
		t.Fatalf("could not read PostgreSQL binary CDC transaction LSN: %v", err)
	}

	receiptObserved := false
	acknowledgedAfterReceipt := false
	waitForPostgresBinaryCondition(t, ctx, done, &stdout, &stderr, "binary CDC warehouse row, checkpoint, bounded stage receipt, and source acknowledgement", func() bool {
		receipt := postgresBinaryBoundedStageReceiptPersisted(root)
		if receipt {
			receiptObserved = true
		}
		acknowledged := postgresBinarySourceAcknowledgedAtOrAfter(ctx, source, committedLSN)
		if acknowledged && !receiptObserved {
			t.Fatal("PostgreSQL source acknowledgement became observable before the connection-owned warehouse receipt")
		}
		if acknowledged && receiptObserved {
			acknowledgedAfterReceipt = true
		}
		return postgresBinaryWarehouseHasRow(ctx, warehouseDir, 901) &&
			postgresBinaryCheckpointPersisted(root, "postgres-cdc-binary", stream) &&
			receipt && acknowledgedAfterReceipt
	})

	stopPostgresBinaryProcess(t, command, done)
	stopped = true
	if !postgresBinaryWarehouseHasRow(ctx, warehouseDir, 901) || !postgresBinaryCheckpointPersisted(root, "postgres-cdc-binary", stream) || !postgresBinaryBoundedStageReceiptPersisted(root) || !acknowledgedAfterReceipt {
		t.Fatal("fresh pm binary lost its observable PostgreSQL CDC state during shutdown")
	}
	t.Logf("postgres_binary_cdc_evidence={\"binary_sha256\":%q,\"connection_id\":%q,\"source_id\":901,\"warehouse_row\":true,\"checkpoint\":true,\"stage_receipt\":true,\"source_lsn_acknowledged\":true}", sha, connectionID)
}

func openPostgresBinarySource(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint) *pgx.Conn {
	t.Helper()
	config, err := pgx.ParseConfig("sslmode=disable")
	if err != nil {
		t.Fatal("could not configure PostgreSQL binary CDC source")
	}
	config.Host = endpoint.Host
	config.Port = uint16(endpoint.Port)
	config.Database = postgresCDCBinaryDatabase
	config.User = postgresCDCBinaryUser
	for {
		connection, err := pgx.ConnectConfig(ctx, config)
		if err == nil {
			return connection
		}
		select {
		case <-ctx.Done():
			t.Fatal("PostgreSQL binary CDC engine was not reachable before the integration deadline")
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func mustRunPostgresBinary(t *testing.T, binary string, args ...string) string {
	t.Helper()
	command := exec.Command(binary, args...)
	command.Env = os.Environ()
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("fresh pm %s failed: %v\n%s", transportCommandName(args), err, output)
	}
	return string(output)
}

func waitForPostgresBinaryCondition(t *testing.T, ctx context.Context, done <-chan error, stdout, stderr *bytes.Buffer, description string, condition func() bool) {
	t.Helper()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if condition() {
			return
		}
		select {
		case err := <-done:
			t.Fatalf("fresh pm exited before %s: %v\nstdout=%s\nstderr=%s", description, err, stdout.String(), stderr.String())
		case <-ctx.Done():
			t.Fatalf("timed out waiting for %s", description)
		case <-ticker.C:
		}
	}
}

func stopPostgresBinaryProcess(t *testing.T, command *exec.Cmd, done <-chan error) {
	t.Helper()
	if command.Process == nil || (command.ProcessState != nil && command.ProcessState.Exited()) {
		return
	}
	_ = command.Process.Signal(os.Interrupt)
	select {
	case <-done:
		return
	case <-time.After(10 * time.Second):
		if err := command.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			t.Errorf("stop fresh pm PostgreSQL CDC process: %v", err)
		}
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("fresh pm PostgreSQL CDC process did not exit after kill")
		}
	}
}

func postgresBinaryWarehouseHasRow(ctx context.Context, warehouseRoot string, wantID int) bool {
	table := ""
	_ = filepath.WalkDir(warehouseRoot, func(path string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && filepath.Base(path) == "postgres_cdc_rows.parquet" {
			table = path
			return fs.SkipAll
		}
		return nil
	})
	if table == "" {
		return false
	}
	found := false
	if err := warehouse.ReadTable(ctx, table, func(row warehouse.Row) error {
		if fmt.Sprint(row["id"]) == strconv.Itoa(wantID) && row["payload"] == "binary-dispatch" {
			found = true
		}
		return nil
	}); err != nil {
		return false
	}
	return found
}

func postgresBinaryCheckpointPersisted(root, connection, stream string) bool {
	contents, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		return false
	}
	var persisted struct {
		StreamStates map[string]struct {
			Connection string `json:"connection"`
			Stream     string `json:"stream"`
			Checkpoint struct {
				Position struct {
					Primary string `json:"primary"`
				} `json:"position"`
				CommittedAt *time.Time `json:"committed_at"`
			} `json:"checkpoint"`
		} `json:"stream_states"`
	}
	if json.Unmarshal(contents, &persisted) != nil {
		return false
	}
	for _, state := range persisted.StreamStates {
		if state.Connection == connection && state.Stream == stream && state.Checkpoint.Position.Primary != "" && state.Checkpoint.CommittedAt != nil {
			return true
		}
	}
	return false
}

func postgresBinaryBoundedStageReceiptPersisted(root string) bool {
	receipts, transactions := 0, 0
	_ = filepath.WalkDir(filepath.Join(root, ".polymetrics", "state", "postgres-cdc-stage"), func(path string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && filepath.Ext(path) == ".json" {
			switch filepath.Base(filepath.Dir(path)) {
			case "receipts":
				receipts++
			case "transactions":
				transactions++
			}
		}
		return nil
	})
	return receipts > 0 && transactions == 0
}

func postgresBinarySourceAcknowledgedAtOrAfter(ctx context.Context, source *pgx.Conn, wantLSN string) bool {
	var acknowledged bool
	err := source.QueryRow(ctx, "SELECT COALESCE(confirmed_flush_lsn >= $1::pg_lsn, false) FROM pg_replication_slots WHERE active AND plugin = 'pgoutput' LIMIT 1", wantLSN).Scan(&acknowledged)
	return err == nil && acknowledged
}
