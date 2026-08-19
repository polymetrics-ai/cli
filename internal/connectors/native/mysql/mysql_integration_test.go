//go:build databaseintegration

package mysql_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
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
	// mysqlIntegrationImageBytes is the pinned image's approximate on-disk
	// footprint, which is what the harness measures its pre-pull headroom
	// against. Measured for 8.4.11 on arm64.
	mysqlIntegrationImageBytes = 830 << 20
)

const (
	envEnabled                            = "POLYMETRICS_DATABASE_INTEGRATION"
	envContainerRuntime                   = "POLYMETRICS_CONTAINER_RUNTIME"
	envContainerEndpoint                  = "POLYMETRICS_CONTAINER_ENDPOINT"
	containerRuntimeConfigurationGuidance = "database integration requires POLYMETRICS_CONTAINER_RUNTIME=docker or podman and POLYMETRICS_CONTAINER_ENDPOINT=unix:///absolute/path/to/socket; no usable explicit local container runtime is configured"
)

var (
	errCollectedCDCEvents            = errors.New("test collected required mysql change events")
	errContainerRuntimeConfiguration = errors.New(containerRuntimeConfigurationGuidance)
)

func TestMySQLContainerHarness(t *testing.T) {
	// A skip here is loud on purpose. This test must never report success
	// without having connected to a real engine.
	if os.Getenv(envEnabled) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the MySQL Docker or Podman proof", envEnabled)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	containerRuntime := dbtest.Runtime(os.Getenv(envContainerRuntime))
	containerEndpoint := os.Getenv(envContainerEndpoint)
	harness, err := newMySQLContainerHarness(containerRuntime, containerEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Errorf("MySQL database test cleanup failed")
		}
		report := harness.Report()
		t.Logf("MySQL database test target image-store free bytes: before=%d after=%d", report.DiskFreeBefore, report.DiskFreeAfter)
	}()

	endpoint, err := harness.Start(ctx)
	if err != nil {
		// Harness errors contain only the safe operation stage and exit class;
		// preserve that reason so an opted-in live test never fails as an
		// unhelpful generic red result (and never exposes connection material).
		t.Fatalf("MySQL database container did not start: %v", err)
	}
	config := mysqlConfig(endpoint)
	connector := native.New()
	waitForMySQL(t, ctx, connector, config)
	seedMySQL(t, ctx, endpoint)

	catalog, err := connector.Catalog(ctx, config)
	if err != nil {
		t.Fatalf("MySQL connector schema discovery failed: %v", err)
	}
	stream := mysqlIntegrationDatabase + "." + mysqlIntegrationTable
	assertDiscoveredStream(t, catalog, stream)

	var full []connectors.Record
	if err := connector.Read(ctx, connectors.ReadRequest{Stream: stream, Config: config}, func(record connectors.Record) error {
		full = append(full, record)
		return nil
	}); err != nil {
		t.Fatalf("MySQL connector full read failed: %v", err)
	}
	assertRecordIDs(t, full, []string{"1", "2", "3", "4", "5"})

	var incremental []connectors.Record
	if err := connector.Read(ctx, connectors.ReadRequest{Stream: stream, Config: config, State: map[string]string{"cursor": "3"}}, func(record connectors.Record) error {
		incremental = append(incremental, record)
		return nil
	}); err != nil {
		t.Fatalf("MySQL connector incremental read failed: %v", err)
	}
	assertRecordIDs(t, incremental, []string{"4", "5"})

	caPath := filepath.Join(t.TempDir(), "mysql-ca.pem")
	if err := harness.CopyFileFromContainer(ctx, "/var/lib/mysql/ca.pem", caPath); err != nil {
		t.Fatal("could not copy the MySQL test certificate authority")
	}
	assertTransportSecurityIsEnforced(t, ctx, connector, endpoint, caPath)
	assertBinaryLogCDC(t, ctx, connector, config, endpoint, stream)
}

func newMySQLContainerHarness(containerRuntime dbtest.Runtime, containerEndpoint string) (*dbtest.Harness, error) {
	if containerRuntime == "" || containerEndpoint == "" {
		return nil, errContainerRuntimeConfiguration
	}
	harness, err := dbtest.New(dbtest.Config{
		Engine:             "mysql",
		ContainerRuntime:   containerRuntime,
		Image:              mysqlIntegrationImage,
		ContainerPort:      3306,
		DataVolumePath:     "/var/lib/mysql",
		ContainerEndpoint:  containerEndpoint,
		ExpectedImageBytes: mysqlIntegrationImageBytes,
		// Colima and similar local Docker VMs report /var/lib/docker inside
		// the daemon, not on the client host. dbtest refuses that path unless
		// this explicitly pre-cached, pinned probe can measure it in-daemon.
		DockerCapacityProbeImage: "docker.io/library/busybox:1.37.0",
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
		return nil, errContainerRuntimeConfiguration
	}
	return harness, nil
}

func TestNewMySQLContainerHarnessConfigurationGuidance(t *testing.T) {
	for _, tc := range []struct {
		name     string
		runtime  dbtest.Runtime
		endpoint string
		secret   string
	}{
		{name: "unknown runtime", runtime: "colima", endpoint: "unix:///tmp/mysql-dbtest.sock", secret: "colima"},
		{name: "unsafe endpoint", runtime: dbtest.RuntimeDocker, endpoint: "tcp://127.0.0.1:2375", secret: "tcp://127.0.0.1:2375"},
		{name: "runtime control character", runtime: "docker\t", endpoint: "unix:///tmp/mysql-dbtest.sock", secret: "docker\t"},
		{name: "endpoint trailing newline", runtime: dbtest.RuntimeDocker, endpoint: "unix:///tmp/mysql-dbtest.sock\n", secret: "unix:///tmp/mysql-dbtest.sock\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			harness, err := newMySQLContainerHarness(tc.runtime, tc.endpoint)
			if harness != nil {
				t.Fatal("newMySQLContainerHarness() returned a harness for invalid runtime configuration")
			}
			if !errors.Is(err, errContainerRuntimeConfiguration) {
				t.Fatalf("newMySQLContainerHarness() error = %v, want runtime configuration guidance", err)
			}
			if strings.Contains(err.Error(), tc.secret) {
				t.Fatalf("configuration guidance exposed supplied runtime input %q", tc.secret)
			}
			for _, required := range []string{
				envContainerRuntime + "=docker or podman",
				envContainerEndpoint + "=unix:///absolute/path/to/socket",
			} {
				if !strings.Contains(err.Error(), required) {
					t.Fatalf("configuration guidance %q does not name %q", err, required)
				}
			}
		})
	}
}

// assertTransportSecurityIsEnforced proves against the live server that the
// declared mode is what actually happens on the wire, rather than a spec
// field nothing reads.
func assertTransportSecurityIsEnforced(t *testing.T, ctx context.Context, connector native.Connector, endpoint dbtest.Endpoint, caPath string) {
	t.Helper()
	for _, tc := range []struct {
		mode        string
		wantConnect bool
		wantCipher  bool
		configure   func(*connectors.RuntimeConfig)
	}{
		{mode: "disabled", wantConnect: true},
		{mode: "preferred", wantConnect: true, wantCipher: true},
		{mode: "required", wantConnect: true, wantCipher: true},
		{mode: "verify-ca", wantConnect: true, wantCipher: true, configure: func(config *connectors.RuntimeConfig) {
			config.Config["sslrootcert"] = caPath
		}},
		{mode: "verify-identity", configure: func(config *connectors.RuntimeConfig) {
			config.Config["sslrootcert"] = caPath
			config.Config["sslservername"] = "invalid.mysql.test"
		}},
	} {
		t.Run("sslmode="+tc.mode, func(t *testing.T) {
			config := mysqlConfig(endpoint)
			config.Config["sslmode"] = tc.mode
			if tc.configure != nil {
				tc.configure(&config)
			}
			err := connector.Check(ctx, config)
			if !tc.wantConnect {
				if err == nil {
					t.Fatalf("Check() under sslmode %q succeeded against an untrusted self-signed certificate", tc.mode)
				}
				return
			}
			if err != nil {
				t.Fatalf("Check() under sslmode %q failed: %v", tc.mode, err)
			}
			if cipher := sessionCipher(t, ctx, config); (cipher != "") != tc.wantCipher {
				t.Fatalf("sslmode %q negotiated cipher %q, want encrypted=%t", tc.mode, cipher, tc.wantCipher)
			}
		})
	}
}

// sessionCipher asks the server what it actually negotiated. An empty
// Ssl_cipher means the session is plaintext.
func sessionCipher(t *testing.T, ctx context.Context, config connectors.RuntimeConfig) string {
	t.Helper()
	conn, err := native.DialForTest(ctx, config)
	if err != nil {
		t.Fatalf("could not reopen MySQL under sslmode %q: %v", config.Config["sslmode"], err)
	}
	defer func() { _ = conn.Close() }()
	result, err := conn.Execute("SHOW SESSION STATUS LIKE 'Ssl_cipher'")
	if err != nil {
		t.Fatal("could not read the negotiated MySQL session cipher")
	}
	defer result.Close()
	cipher, err := result.GetString(0, 1)
	if err != nil {
		t.Fatal("MySQL returned no session cipher status row")
	}
	return cipher
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
			t.Fatalf("could not seed deterministic MySQL test data: %v", err)
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
		if event.State["binlog_file"] == nil || event.State["binlog_pos"] == nil || event.State["binlog_row"] == nil || event.State["schema_fingerprint"] == nil {
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
	for _, state := range checkpoints.states {
		if state["schema_fingerprint"] == "" {
			t.Fatal("MySQL change capture committed a checkpoint without schema metadata")
		}
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
