//go:build databaseintegration

package mysql_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-mysql-org/go-mysql/client"
	gomysql "github.com/go-mysql-org/go-mysql/mysql"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/native/dbtest"
	native "polymetrics.ai/internal/connectors/native/mysql"
)

const (
	mysqlIntegrationImage    = "docker.io/library/mysql:8.4.11"
	mysqlIntegrationDatabase = "pm_harness"
	mysqlIntegrationTable    = "events"
	mysqlReplicationServerID = "731001"
)

var errCollectedCDCEvents = errors.New("test collected required mysql change events")

func TestMySQLContainerHarness(t *testing.T) {
	if os.Getenv("POLYMETRICS_DATABASE_INTEGRATION") != "1" {
		t.Skip("database integration skipped: set POLYMETRICS_DATABASE_INTEGRATION=1 to start the MySQL Docker/Colima proof")
	}
	dockerContext := strings.TrimSpace(os.Getenv("DOCKER_CONTEXT"))
	if dockerContext == "" {
		t.Skip("database integration skipped: DOCKER_CONTEXT must name the explicit Docker/Colima context")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	harness, err := dbtest.New(dbtest.Config{
		Engine:         "mysql",
		Image:          mysqlIntegrationImage,
		ContainerPort:  3306,
		DataVolumePath: "/var/lib/mysql",
		DockerContext:  dockerContext,
		ResetColima:    os.Getenv("POLYMETRICS_DATABASE_RESET_COLIMA") == "1",
		ColimaProfile:  "default",
		ContainerArgs: []string{
			"--env", "MYSQL_ALLOW_EMPTY_PASSWORD=yes",
			"--env", "MYSQL_ROOT_HOST=%",
			"--env", "MYSQL_DATABASE=" + mysqlIntegrationDatabase,
		},
		EngineArgs: []string{
			"--log-bin=mysql-bin",
			"--binlog-format=ROW",
			"--binlog-row-image=FULL",
			"--binlog-row-metadata=FULL",
			"--server-id=731000",
		},
	})
	if err != nil {
		t.Fatal("could not configure the MySQL database test harness")
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Errorf("MySQL database test cleanup failed")
		}
		report := harness.Report()
		t.Logf("MySQL database test disk free bytes: before=%d after=%d colima_reset=%t", report.DiskFreeBefore, report.DiskFreeAfter, report.ColimaReset)
		if !report.ColimaReset {
			t.Log("MySQL database test did not reset Colima; set POLYMETRICS_DATABASE_RESET_COLIMA=1 to reclaim host disk")
		}
		if report.ColimaReset && report.DiskFreeAfter+(128<<20) < report.DiskFreeBefore {
			t.Errorf("MySQL database test reclaimed insufficient disk space")
		}
	}()

	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatal("MySQL database container did not start")
	}
	config := mysqlConfig(endpoint)
	connector := native.New()
	waitForMySQL(t, ctx, connector, config)
	seedMySQL(t, ctx, endpoint)

	catalog, err := connector.Catalog(ctx, config)
	if err != nil {
		t.Fatal("MySQL connector schema discovery failed")
	}
	stream := mysqlIntegrationDatabase + "." + mysqlIntegrationTable
	assertDiscoveredStream(t, catalog, stream)

	var full []connectors.Record
	if err := connector.Read(ctx, connectors.ReadRequest{Stream: stream, Config: config}, func(record connectors.Record) error {
		full = append(full, record)
		return nil
	}); err != nil {
		t.Fatal("MySQL connector full read failed")
	}
	assertRecordIDs(t, full, []string{"1", "2", "3", "4", "5"})

	var incremental []connectors.Record
	if err := connector.Read(ctx, connectors.ReadRequest{Stream: stream, Config: config, State: map[string]string{"cursor": "3"}}, func(record connectors.Record) error {
		incremental = append(incremental, record)
		return nil
	}); err != nil {
		t.Fatal("MySQL connector incremental read failed")
	}
	assertRecordIDs(t, incremental, []string{"4", "5"})

	assertBinaryLogCDC(t, ctx, connector, config, endpoint, stream)
}

func mysqlConfig(endpoint dbtest.Endpoint) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{Config: map[string]string{
		"host":                  endpoint.Host,
		"port":                  strconv.Itoa(endpoint.Port),
		"database":              mysqlIntegrationDatabase,
		"username":              gomysql.DEFAULT_USER,
		"cursor_field":          "sequence",
		"page_size":             "2",
		"read_limit":            "10",
		"replication_server_id": mysqlReplicationServerID,
	}}
}

func waitForMySQL(t *testing.T, ctx context.Context, connector native.Connector, config connectors.RuntimeConfig) {
	t.Helper()
	for {
		if err := connector.Check(ctx, config); err == nil {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("MySQL engine was not reachable before the integration deadline")
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func seedMySQL(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint) {
	t.Helper()
	db, err := client.ConnectWithContext(ctx, net.JoinHostPort(endpoint.Host, strconv.Itoa(endpoint.Port)), gomysql.DEFAULT_USER, "", mysqlIntegrationDatabase, 10*time.Second)
	if err != nil {
		t.Fatal("could not open the isolated MySQL test database")
	}
	defer func() { _ = db.Close() }()
	statements := []string{
		"CREATE TABLE events (id BIGINT PRIMARY KEY, sequence BIGINT NOT NULL UNIQUE, label VARCHAR(64) NOT NULL)",
		"INSERT INTO events (id, sequence, label) VALUES (1, 1, 'alpha'), (2, 2, 'bravo'), (3, 3, 'charlie'), (4, 4, 'delta'), (5, 5, 'echo')",
	}
	for _, statement := range statements {
		if _, err := db.Execute(statement); err != nil {
			t.Fatal("could not seed deterministic MySQL test data")
		}
	}
}

func assertDiscoveredStream(t *testing.T, catalog connectors.Catalog, name string) {
	t.Helper()
	for _, stream := range catalog.Streams {
		if stream.Name != name {
			continue
		}
		if len(stream.PrimaryKey) != 1 || stream.PrimaryKey[0] != "id" {
			t.Fatal("MySQL discovery returned an incorrect primary key")
		}
		if len(stream.CursorFields) != 1 || stream.CursorFields[0] != "sequence" {
			t.Fatal("MySQL discovery returned no configured incremental cursor")
		}
		return
	}
	t.Fatal("MySQL discovery omitted the seeded stream")
}

func assertRecordIDs(t *testing.T, records []connectors.Record, want []string) {
	t.Helper()
	if len(records) != len(want) {
		t.Fatalf("MySQL connector returned %d records, want %d", len(records), len(want))
	}
	got := make([]string, 0, len(records))
	for _, record := range records {
		got = append(got, recordText(record["id"]))
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("MySQL record identifiers = %v, want %v", got, want)
	}
}

func assertBinaryLogCDC(t *testing.T, ctx context.Context, connector native.Connector, config connectors.RuntimeConfig, endpoint dbtest.Endpoint, stream string) {
	t.Helper()
	type checkpointRecorder struct{ states []map[string]string }
	var checkpoints checkpointRecorder
	commit := checkpointCommitterFunc(func(_ context.Context, state map[string]string) error {
		copy := make(map[string]string, len(state))
		for key, value := range state {
			copy[key] = value
		}
		checkpoints.states = append(checkpoints.states, copy)
		return nil
	})

	events := make([]connectors.CDCEvent, 0, 4)
	done := make(chan error, 1)
	go func() {
		done <- connector.ReadCDC(ctx, connectors.CDCReadRequest{Stream: stream, Config: config, CheckpointCommitter: commit}, func(event connectors.CDCEvent) error {
			events = append(events, event)
			if len(events) == 4 {
				return errCollectedCDCEvents
			}
			return nil
		})
	}()
	waitForCDCRegistration(t, ctx, endpoint)
	db, err := client.ConnectWithContext(ctx, net.JoinHostPort(endpoint.Host, strconv.Itoa(endpoint.Port)), gomysql.DEFAULT_USER, "", mysqlIntegrationDatabase, 10*time.Second)
	if err != nil {
		t.Fatal("could not open MySQL to generate change events")
	}
	defer func() { _ = db.Close() }()
	for _, statement := range []string{
		"INSERT INTO events (id, sequence, label) VALUES (6, 6, 'foxtrot'), (7, 7, 'hotel')",
		"UPDATE events SET label = 'golf' WHERE id = 6",
		"DELETE FROM events WHERE id = 6",
	} {
		if _, err := db.Execute(statement); err != nil {
			t.Fatal("could not generate a MySQL binary-log change event")
		}
	}
	select {
	case err := <-done:
		if !errors.Is(err, errCollectedCDCEvents) {
			var serverErr *gomysql.MyError
			if errors.As(err, &serverErr) {
				t.Fatalf("MySQL connector change capture failed with server error code %d", serverErr.Code)
			}
			t.Fatal("MySQL connector change capture did not return after the required real events")
		}
	case <-ctx.Done():
		t.Fatal("MySQL connector change capture did not receive real row events before the deadline")
	}
	operations := make([]string, 0, len(events))
	identifiers := make([]string, 0, len(events))
	for _, event := range events {
		operations = append(operations, event.Operation)
		identifiers = append(identifiers, recordText(event.Record["id"]))
		if event.State["binlog_file"] == nil || event.State["binlog_pos"] == nil || event.State["binlog_row"] == nil {
			t.Fatal("MySQL connector change capture returned an incomplete real record or checkpoint state")
		}
	}
	if strings.Join(operations, ",") != "insert,insert,update,delete" {
		t.Fatalf("MySQL change operations = %v, want insert, insert, update, delete", operations)
	}
	if strings.Join(identifiers, ",") != "6,7,6,6" {
		t.Fatalf("MySQL change record identifiers = %v, want 6, 7, 6, 6", identifiers)
	}
	if events[0].State["binlog_pos"] != events[1].State["binlog_pos"] || events[0].State["binlog_row"] == events[1].State["binlog_row"] {
		t.Fatal("MySQL change capture did not distinguish rows from one binary-log event")
	}
	if len(checkpoints.states) < 2 {
		t.Fatal("MySQL change capture did not commit acknowledged binary-log positions")
	}
}

func waitForCDCRegistration(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint) {
	t.Helper()
	db, err := client.ConnectWithContext(ctx, net.JoinHostPort(endpoint.Host, strconv.Itoa(endpoint.Port)), gomysql.DEFAULT_USER, "", mysqlIntegrationDatabase, 10*time.Second)
	if err != nil {
		t.Fatal("could not inspect MySQL replication registration")
	}
	defer func() { _ = db.Close() }()
	for {
		result, err := db.Execute("SELECT COUNT(*) FROM information_schema.processlist WHERE COMMAND IN ('Binlog Dump', 'Binlog Dump GTID')")
		if err == nil && result != nil && result.Resultset != nil {
			count, countErr := result.GetInt(0, 0)
			result.Close()
			if countErr == nil && count > 0 {
				return
			}
		} else if result != nil {
			result.Close()
		}
		select {
		case <-ctx.Done():
			t.Fatal("MySQL replication client did not register before generating change events")
		case <-time.After(50 * time.Millisecond):
		}
	}
}

type checkpointCommitterFunc func(context.Context, map[string]string) error

func (fn checkpointCommitterFunc) CommitChangefeedCheckpoint(ctx context.Context, state map[string]string) error {
	return fn(ctx, state)
}

func recordText(value any) string {
	switch typed := value.(type) {
	case []byte:
		return string(typed)
	default:
		return fmt.Sprint(typed)
	}
}
