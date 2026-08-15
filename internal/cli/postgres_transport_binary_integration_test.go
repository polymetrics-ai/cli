//go:build databaseintegration

package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors/native/dbtest"
)

const (
	postgresTransportImage    = "docker.io/library/postgres:16.10"
	postgresTransportSourceDB = "pm_transport_source"
	postgresTransportTargetDB = "pm_transport_target"
	postgresTransportUser     = "pm_transport"
)

// TestPMBinaryExecutesPostgresWarehousePostgres proves the shipped composition
// root rather than a hand-built driver: fresh pm processes consume a real
// PostgreSQL source, durable connection-owned WAL/Parquet/manifest artifacts,
// the registered managed-target destination, and a separate real PostgreSQL
// target database. Replaying the same source snapshot through a fresh approved
// run leaves both business rows and the acknowledged delivery ID unchanged.
func TestPMBinaryExecutesPostgresWarehousePostgres(t *testing.T) {
	if os.Getenv("POLYMETRICS_DATABASE_INTEGRATION") != "1" {
		t.Skip("database integration is opt-in")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	harness := newPostgresTransportHarness(t)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Errorf("close PostgreSQL transport harness: %v", err)
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("start PostgreSQL transport harness: %v", err)
	}
	admin := waitForPostgresTransport(t, ctx, endpoint, postgresTransportSourceDB, postgresTransportUser)
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+postgresTransportTargetDB); err != nil {
		t.Fatalf("create isolated target database: %v", err)
	}
	if _, err := admin.Exec(ctx, `CREATE TABLE public.events (id bigint PRIMARY KEY, sequence bigint NOT NULL, label text NOT NULL)`); err != nil {
		t.Fatalf("create live source table: %v", err)
	}
	if _, err := admin.Exec(ctx, `INSERT INTO public.events (id, sequence, label) SELECT value, value * 10, 'event-' || value::text FROM generate_series(1, 1001) AS value`); err != nil {
		t.Fatalf("seed live source rows: %v", err)
	}
	if _, err := admin.Exec(ctx, `CREATE TABLE public.bootstrap_events (id bigint PRIMARY KEY, sequence bigint NOT NULL, label text NOT NULL)`); err != nil {
		t.Fatalf("create live bootstrap source table: %v", err)
	}
	if _, err := admin.Exec(ctx, `INSERT INTO public.bootstrap_events VALUES (1, 10, 'bootstrap')`); err != nil {
		t.Fatalf("seed live bootstrap source row: %v", err)
	}
	if _, err := admin.Exec(ctx, `CREATE PUBLICATION pm_transport_pub FOR TABLE public.bootstrap_events`); err != nil {
		t.Fatalf("create live bootstrap publication: %v", err)
	}
	if _, err := admin.Exec(ctx, `CREATE TABLE public.empty_events (id bigint PRIMARY KEY, sequence bigint NOT NULL, label text NOT NULL)`); err != nil {
		t.Fatalf("create empty source table: %v", err)
	}
	if _, err := admin.Exec(ctx, `CREATE ROLE pm_transport_denied LOGIN`); err != nil {
		t.Fatalf("create restricted target role: %v", err)
	}
	defer func() { _ = admin.Close(context.Background()) }()
	target := waitForPostgresTransport(t, ctx, endpoint, postgresTransportTargetDB, postgresTransportUser)
	defer func() { _ = target.Close(context.Background()) }()

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	mustPostgresTransportPM(t, binary, "", "init", "--root", root, "--json")
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-source", postgresTransportSourceDB)
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-bootstrap", postgresTransportSourceDB,
		"transport_bootstrap=true", "cdc_publication=pm_transport_pub")
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-target", postgresTransportTargetDB)
	addPostgresTransportCredentialForUser(t, binary, root, endpoint, "pg-target-denied", postgresTransportTargetDB, "pm_transport_denied")
	addPostgresTransportCredentialForUser(t, binary, root, endpoint, "pg-target-missing-user", postgresTransportTargetDB, "pm_transport_missing_user")
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "postgres-to-postgres",
		"--source", "postgres:pg-source", "--destination", "postgres:pg-target",
		"--stream", "public.events", "--sync-mode", "incremental_upsert",
		"--cursor", "sequence", "--primary-key", "id", "--table", "events",
		"--root", root, "--json")

	firstApproved := runApprovedPostgresTransportBinary(t, binary, root, "postgres-to-postgres", "public.events", 1000)
	firstRun := firstApproved.Run
	if firstRun.RecordsRead != 1001 || firstRun.RecordsLoaded != 1001 || firstRun.Status != "completed" {
		t.Fatalf("first PostgreSQL binary run = %#v, want exact 1001-row completed transfer", firstRun)
	}
	schema, relation, count, delivery := postgresTransportTargetState(t, ctx, target)
	if count != 1001 || delivery == "" {
		t.Fatalf("managed target state = schema %q relation %q rows %d delivery %q, want 1001 durable rows and receipt", schema, relation, count, delivery)
	}
	assertPostgresTransportWarehouse(t, root)
	checkpointBeforeTokenReplay := postgresTransportStreamStates(t, root)
	replayOutput, replayErr := runTransportPM(binary, firstApproved.Token+"\n",
		"etl", "run", "--connection", "postgres-to-postgres", "--stream", "public.events",
		"--batch-size", "1000", "--approval-plan", firstApproved.PlanID,
		"--approval-token-stdin", "--confirm", "destructive", "--root", root, "--json")
	if replayErr == nil {
		t.Fatalf("consumed approval token replay unexpectedly succeeded: %s", replayOutput)
	}
	if strings.Contains(replayOutput, firstApproved.Token) {
		t.Fatal("consumed approval token appeared in refusal output")
	}
	_, _, refusedReplayCount, refusedReplayDelivery := postgresTransportTargetState(t, ctx, target)
	if refusedReplayCount != count || refusedReplayDelivery != delivery {
		t.Fatalf("consumed-token refusal changed target: rows %d->%d delivery %q->%q", count, refusedReplayCount, delivery, refusedReplayDelivery)
	}
	if checkpointAfterTokenReplay := postgresTransportStreamStates(t, root); checkpointAfterTokenReplay != checkpointBeforeTokenReplay {
		t.Fatalf("consumed-token refusal advanced checkpoint: before=%s after=%s", checkpointBeforeTokenReplay, checkpointAfterTokenReplay)
	}

	secondRun := runApprovedPostgresTransportBinary(t, binary, root, "postgres-to-postgres", "public.events", 1000).Run
	if secondRun.RecordsRead != 1001 || secondRun.RecordsLoaded != 1001 || secondRun.Status != "completed" {
		t.Fatalf("replay PostgreSQL binary run = %#v, want exact completed replay", secondRun)
	}
	_, _, replayCount, replayDelivery := postgresTransportTargetState(t, ctx, target)
	if replayCount != count || replayDelivery != delivery {
		t.Fatalf("acknowledged replay changed target: rows %d->%d delivery %q->%q", count, replayCount, delivery, replayDelivery)
	}

	schemaDriftApproval := preparePostgresTransportApproval(t, binary, root, "postgres-to-postgres", "public.events")
	checkpointBeforeSchemaDrift := postgresTransportStreamStates(t, root)
	if _, err := admin.Exec(ctx, `ALTER TABLE public.events ADD COLUMN drifted text`); err != nil {
		t.Fatalf("inject live source schema drift: %v", err)
	}
	driftOutput, driftErr := runPostgresTransportApproval(binary, root, "postgres-to-postgres", "public.events", 1000, schemaDriftApproval)
	if driftErr == nil || !strings.Contains(driftOutput, "PostgreSQL managed target approval is stale") {
		t.Fatalf("schema drift run = (%v, %s), want typed drift refusal", driftErr, driftOutput)
	}
	_, _, driftCount, driftDelivery := postgresTransportTargetState(t, ctx, target)
	if driftCount != count || driftDelivery != delivery {
		t.Fatalf("schema-drift refusal changed target: rows %d->%d delivery %q->%q", count, driftCount, delivery, driftDelivery)
	}
	if checkpointAfterDrift := postgresTransportStreamStates(t, root); checkpointAfterDrift != checkpointBeforeSchemaDrift {
		t.Fatalf("schema-drift refusal advanced checkpoint: before=%s after=%s", checkpointBeforeSchemaDrift, checkpointAfterDrift)
	}
	if _, err := admin.Exec(ctx, `ALTER TABLE public.events DROP COLUMN drifted`); err != nil {
		t.Fatalf("restore live source schema after drift refusal: %v", err)
	}

	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "postgres-bootstrap",
		"--source", "postgres:pg-bootstrap", "--destination", "postgres:pg-target",
		"--stream", "public.bootstrap_events", "--sync-mode", "incremental_upsert",
		"--cursor", "sequence", "--primary-key", "id", "--table", "bootstrap_events",
		"--root", root, "--json")
	bootstrapApproval := preparePostgresTransportApproval(t, binary, root, "postgres-bootstrap", "public.bootstrap_events")
	bootstrapProcess := startPostgresTransportApproval(t, binary, root, "postgres-bootstrap", "public.bootstrap_events", 1000, bootstrapApproval)
	defer bootstrapProcess.stop()
	waitForPostgresTransportCondition(t, bootstrapProcess, func() bool {
		counts := postgresTransportBusinessCounts(t, ctx, target)
		return len(counts) == 2 && counts[0] == 1 && counts[1] == 1001 && strings.Contains(postgresTransportStreamStates(t, root), `"logical_replication"`)
	}, "bootstrap snapshot row and durable logical-replication checkpoint")
	bootstrapBarrierState := postgresTransportStreamStates(t, root)
	if _, err := admin.Exec(ctx, `BEGIN; UPDATE public.bootstrap_events SET sequence = 20, label = 'post-barrier-update' WHERE id = 1; INSERT INTO public.bootstrap_events VALUES (2, 30, 'post-barrier-insert'); DELETE FROM public.bootstrap_events WHERE id = 1; COMMIT`); err != nil {
		t.Fatalf("write post-barrier transaction: %v", err)
	}
	waitForPostgresTransportCondition(t, bootstrapProcess, func() bool {
		return postgresTransportLabelCount(t, ctx, target, "post-barrier-insert") == 1 && postgresTransportLabelCount(t, ctx, target, "bootstrap") == 0 && postgresTransportStreamStates(t, root) != bootstrapBarrierState
	}, "post-barrier transaction in target and an advanced LSN checkpoint")
	postBarrierState := postgresTransportStreamStates(t, root)
	bootstrapProcess.killAndWait(t)

	resumeApproval := preparePostgresTransportApproval(t, binary, root, "postgres-bootstrap", "public.bootstrap_events")
	resumeProcess := startPostgresTransportApproval(t, binary, root, "postgres-bootstrap", "public.bootstrap_events", 1000, resumeApproval)
	defer resumeProcess.stop()
	if _, err := admin.Exec(ctx, `INSERT INTO public.bootstrap_events VALUES (3, 40, 'resumed-after-process-death')`); err != nil {
		t.Fatalf("write resumed change after process death: %v", err)
	}
	waitForPostgresTransportCondition(t, resumeProcess, func() bool {
		return postgresTransportLabelCount(t, ctx, target, "resumed-after-process-death") == 1 && postgresTransportStreamStates(t, root) != postBarrierState
	}, "resumed CDC row and checkpoint after process restart")
	resumeProcess.killAndWait(t)
	counts := postgresTransportBusinessCounts(t, ctx, target)
	if len(counts) != 2 || counts[0] != 2 || counts[1] != 1001 {
		t.Fatalf("managed target business row counts after bootstrap/resume = %v, want [2 1001]", counts)
	}

	assertPostgresTransportDestinationRefusal(t, ctx, binary, root, target, "postgres-auth-refusal", "pg-target-missing-user", "postgres managed target transport is unavailable")
	assertPostgresTransportDestinationRefusal(t, ctx, binary, root, target, "postgres-permission-refusal", "pg-target-denied", "database managed target state cannot be proven")

	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "postgres-empty",
		"--source", "postgres:pg-source", "--destination", "postgres:pg-target",
		"--stream", "public.empty_events", "--sync-mode", "incremental_upsert",
		"--cursor", "sequence", "--primary-key", "id", "--table", "empty_events",
		"--root", root, "--json")
	emptyRun := runApprovedPostgresTransportBinary(t, binary, root, "postgres-empty", "public.empty_events", 1000).Run
	if emptyRun.RecordsRead != 0 || emptyRun.RecordsLoaded != 0 || emptyRun.Status != "completed" {
		t.Fatalf("empty PostgreSQL binary run = %#v, want completed zero-row boundary", emptyRun)
	}
	if emptyCounts := postgresTransportBusinessCounts(t, ctx, target); len(emptyCounts) != 2 || emptyCounts[0] != 2 || emptyCounts[1] != 1001 {
		t.Fatalf("empty PostgreSQL run changed target business rows: %v", emptyCounts)
	}
	if emptyStates := postgresTransportStreamStates(t, root); !strings.Contains(emptyStates, "postgres-empty") || !strings.Contains(emptyStates, "public.empty_events") {
		t.Fatalf("empty PostgreSQL run did not durably advance its checkpoint: %s", emptyStates)
	}
}

// TestPMBinaryExecutesAuthenticatedGitHubWarehousePostgres proves the second
// production destination route required by #3982. The source call reaches the
// real GitHub API with a credential supplied only through pm credentials add;
// the test never logs or serializes that value. It is a distinct opt-in from
// the database harness so ordinary database-integration CI does not silently
// substitute public or fake GitHub access when the credential is unavailable.
func TestPMBinaryExecutesAuthenticatedGitHubWarehousePostgres(t *testing.T) {
	if os.Getenv("POLYMETRICS_DATABASE_INTEGRATION") != "1" || os.Getenv("POLYMETRICS_GITHUB_INTEGRATION") != "1" {
		t.Skip("authenticated GitHub-to-PostgreSQL integration is opt-in")
	}
	if os.Getenv("POLYMETRICS_GITHUB_TOKEN") == "" {
		t.Fatal("POLYMETRICS_GITHUB_INTEGRATION=1 requires POLYMETRICS_GITHUB_TOKEN")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	harness := newPostgresTransportHarness(t)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Errorf("close GitHub-to-PostgreSQL harness: %v", err)
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("start GitHub-to-PostgreSQL harness: %v", err)
	}
	admin := waitForPostgresTransport(t, ctx, endpoint, postgresTransportSourceDB, postgresTransportUser)
	defer func() { _ = admin.Close(context.Background()) }()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+postgresTransportTargetDB); err != nil {
		t.Fatalf("create GitHub target database: %v", err)
	}
	target := waitForPostgresTransport(t, ctx, endpoint, postgresTransportTargetDB, postgresTransportUser)
	defer func() { _ = target.Close(context.Background()) }()

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "github-project")
	mustPostgresTransportPM(t, binary, "", "init", "--root", root, "--json")
	mustPostgresTransportPM(t, binary, "",
		"credentials", "add", "github-live", "--connector", "github",
		"--config", "owner=polymetrics-ai", "--config", "repo=cli", "--config", "auth_type=token",
		"--config", "rate_limit_account=authenticated-transport-proof",
		"--from-env", "token=POLYMETRICS_GITHUB_TOKEN", "--root", root, "--json")
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-target", postgresTransportTargetDB)
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "github-to-postgres",
		"--source", "github:github-live", "--destination", "postgres:pg-target",
		"--stream", "issues", "--sync-mode", "full_overwrite",
		"--cursor", "updated_at", "--primary-key", "node_id", "--table", "issues",
		"--source-config", "transport_source_issue_number=4163",
		"--root", root, "--json")

	run := runApprovedPostgresTransportBinary(t, binary, root, "github-to-postgres", "issues", 1).Run
	if run.Status != "completed" || run.RecordsRead != 1 || run.RecordsLoaded != 1 {
		t.Fatalf("authenticated GitHub binary run = %#v, want exact one-row transfer", run)
	}
	schema, relation, count, delivery := postgresTransportTargetState(t, ctx, target)
	if count != 1 || delivery == "" {
		t.Fatalf("GitHub managed target = %s.%s rows=%d delivery=%q, want one durable row", schema, relation, count, delivery)
	}
	qualified := pgx.Identifier{schema, relation}.Sanitize()
	var number int64
	var nodeID, labels string
	if err := target.QueryRow(ctx, "SELECT number, node_id, labels::text FROM "+qualified).Scan(&number, &nodeID, &labels); err != nil {
		t.Fatalf("read authenticated GitHub target row: %v", err)
	}
	if number != 4163 || strings.TrimSpace(nodeID) == "" || !json.Valid([]byte(labels)) {
		t.Fatalf("authenticated GitHub target row number=%d node_id_present=%t labels_json=%t", number, nodeID != "", json.Valid([]byte(labels)))
	}
	assertPostgresTransportWarehouse(t, root)
	if states := postgresTransportStreamStates(t, root); !strings.Contains(states, "github-to-postgres") || !strings.Contains(states, "issues") {
		t.Fatalf("authenticated GitHub run did not publish its acknowledged checkpoint: %s", states)
	}
}

type postgresTransportRun struct {
	Status        string `json:"status"`
	RecordsRead   int    `json:"records_read"`
	RecordsLoaded int    `json:"records_loaded"`
}

type approvedPostgresTransportRun struct {
	Run    postgresTransportRun
	PlanID string
	Token  string
}

func runApprovedPostgresTransportBinary(t *testing.T, binary, root, connection, stream string, batchSize int) approvedPostgresTransportRun {
	t.Helper()
	approval := preparePostgresTransportApproval(t, binary, root, connection, stream)
	runOutput, err := runPostgresTransportApproval(binary, root, connection, stream, batchSize, approval)
	if err != nil {
		t.Fatalf("approved PostgreSQL transport run failed: %v\n%s", err, runOutput)
	}
	var envelope struct {
		Run postgresTransportRun `json:"run"`
	}
	if err := json.Unmarshal([]byte(runOutput), &envelope); err != nil {
		t.Fatalf("decode PostgreSQL transport run: %v output=%s", err, runOutput)
	}
	return approvedPostgresTransportRun{Run: envelope.Run, PlanID: approval.PlanID, Token: approval.Token}
}

func preparePostgresTransportApproval(t *testing.T, binary, root, connection, stream string) approvedPostgresTransportRun {
	t.Helper()
	planOutput := mustPostgresTransportPM(t, binary, "",
		"etl", "transport", "postgres-managed-target", "plan",
		"--connection", connection, "--stream", stream, "--root", root, "--json")
	var planned struct {
		Plan struct {
			ID string `json:"id"`
		} `json:"plan"`
	}
	if err := json.Unmarshal([]byte(planOutput), &planned); err != nil || planned.Plan.ID == "" {
		t.Fatalf("decode PostgreSQL transport plan: %v output=%s", err, planOutput)
	}
	preview := mustPostgresTransportPM(t, binary, "",
		"etl", "transport", "postgres-managed-target", "preview", planned.Plan.ID,
		"--root", root)
	const marker = "Approval token: "
	index := strings.Index(preview, marker)
	if index < 0 {
		t.Fatal("PostgreSQL transport human preview did not issue an approval token")
	}
	token := strings.TrimSpace(strings.SplitN(preview[index+len(marker):], "\n", 2)[0])
	if token == "" {
		t.Fatal("PostgreSQL transport approval token was empty")
	}
	return approvedPostgresTransportRun{PlanID: planned.Plan.ID, Token: token}
}

func runPostgresTransportApproval(binary, root, connection, stream string, batchSize int, approval approvedPostgresTransportRun) (string, error) {
	return runTransportPM(binary, approval.Token+"\n",
		"etl", "run", "--connection", connection, "--stream", stream,
		"--batch-size", strconv.Itoa(batchSize), "--approval-plan", approval.PlanID,
		"--approval-token-stdin", "--confirm", "destructive", "--root", root, "--json")
}

type postgresTransportProcess struct {
	command *exec.Cmd
	output  lockedPostgresTransportBuffer
	done    chan error
	stopped bool
}

type lockedPostgresTransportBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (b *lockedPostgresTransportBuffer) Write(value []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.Write(value)
}

func (b *lockedPostgresTransportBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
}

func startPostgresTransportApproval(t *testing.T, binary, root, connection, stream string, batchSize int, approval approvedPostgresTransportRun) *postgresTransportProcess {
	t.Helper()
	process := &postgresTransportProcess{done: make(chan error, 1)}
	process.command = exec.Command(binary,
		"etl", "run", "--connection", connection, "--stream", stream,
		"--batch-size", strconv.Itoa(batchSize), "--approval-plan", approval.PlanID,
		"--approval-token-stdin", "--confirm", "destructive", "--root", root, "--json")
	process.command.Stdin = strings.NewReader(approval.Token + "\n")
	process.command.Stdout = &process.output
	process.command.Stderr = &process.output
	if err := process.command.Start(); err != nil {
		t.Fatalf("start PostgreSQL transport process: %v", err)
	}
	go func() { process.done <- process.command.Wait() }()
	return process
}

func (p *postgresTransportProcess) killAndWait(t *testing.T) {
	t.Helper()
	if p == nil || p.stopped {
		return
	}
	if err := p.command.Process.Kill(); err != nil {
		t.Fatalf("kill PostgreSQL transport process: %v", err)
	}
	if err := <-p.done; err == nil {
		t.Fatalf("killed PostgreSQL transport process reported success: %s", p.output.String())
	}
	p.stopped = true
}

func (p *postgresTransportProcess) stop() {
	if p == nil || p.stopped {
		return
	}
	_ = p.command.Process.Kill()
	<-p.done
	p.stopped = true
}

func waitForPostgresTransportCondition(t *testing.T, process *postgresTransportProcess, condition func() bool, description string) {
	t.Helper()
	deadline := time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-process.done:
			process.stopped = true
			t.Fatalf("PostgreSQL transport process exited before %s: %v\n%s", description, err, process.output.String())
		default:
		}
		if condition() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s; process output=%s", description, process.output.String())
}

func postgresTransportStreamStates(t *testing.T, root string) string {
	t.Helper()
	stateBytes, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read project state: %v", err)
	}
	var persisted struct {
		StreamStates json.RawMessage `json:"stream_states"`
	}
	if err := json.Unmarshal(stateBytes, &persisted); err != nil {
		t.Fatalf("decode project stream states: %v", err)
	}
	return string(persisted.StreamStates)
}

func newPostgresTransportHarness(t *testing.T) *dbtest.Harness {
	t.Helper()
	harness, err := dbtest.New(dbtest.Config{
		Engine: "postgres-transport", Image: postgresTransportImage,
		ContainerPort: 5432, DataVolumePath: "/var/lib/postgresql/data",
		ContainerRuntime:   dbtest.Runtime(strings.TrimSpace(os.Getenv("POLYMETRICS_CONTAINER_RUNTIME"))),
		ContainerEndpoint:  strings.TrimSpace(os.Getenv("POLYMETRICS_CONTAINER_ENDPOINT")),
		ExpectedImageBytes: 420 << 20, DockerCapacityProbeImage: "docker.io/library/busybox:1.37.0",
		ContainerArgs: []string{
			"--env", "POSTGRES_DB=" + postgresTransportSourceDB,
			"--env", "POSTGRES_USER=" + postgresTransportUser,
			"--env", "POSTGRES_HOST_AUTH_METHOD=trust",
		},
		EngineArgs: []string{"-c", "wal_level=logical", "-c", "max_replication_slots=8", "-c", "max_wal_senders=8"},
	})
	if err != nil {
		t.Fatalf("configure PostgreSQL transport harness: %v", err)
	}
	return harness
}

func waitForPostgresTransport(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint, database, user string) *pgx.Conn {
	t.Helper()
	for {
		config, err := pgx.ParseConfig("sslmode=disable")
		if err != nil {
			t.Fatal(err)
		}
		config.Host, config.Port, config.Database, config.User = endpoint.Host, uint16(endpoint.Port), database, user
		conn, err := pgx.ConnectConfig(ctx, config)
		if err == nil {
			return conn
		}
		select {
		case <-ctx.Done():
			t.Fatal("PostgreSQL transport database was not reachable")
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func addPostgresTransportCredential(t *testing.T, binary, root string, endpoint dbtest.Endpoint, name, database string, extraConfig ...string) {
	t.Helper()
	addPostgresTransportCredentialForUser(t, binary, root, endpoint, name, database, postgresTransportUser, extraConfig...)
}

func addPostgresTransportCredentialForUser(t *testing.T, binary, root string, endpoint dbtest.Endpoint, name, database, username string, extraConfig ...string) {
	t.Helper()
	args := []string{
		"credentials", "add", name, "--connector", "postgres",
		"--config", "host=" + endpoint.Host, "--config", "port=" + strconv.Itoa(endpoint.Port),
		"--config", "database=" + database, "--config", "username=" + username,
		"--config", "schema=public", "--config", "sslmode=disable",
	}
	for _, config := range extraConfig {
		args = append(args, "--config", config)
	}
	args = append(args, "--value-stdin", "password", "--root", root, "--json")
	mustPostgresTransportPM(t, binary, postgresTransportFixturePassword(t)+"\n", args...)
}

func postgresTransportFixturePassword(t *testing.T) string {
	t.Helper()
	var value [32]byte
	if _, err := rand.Read(value[:]); err != nil {
		t.Fatalf("generate PostgreSQL transport fixture password: %v", err)
	}
	return hex.EncodeToString(value[:])
}

func assertPostgresTransportDestinationRefusal(t *testing.T, ctx context.Context, binary, root string, target *pgx.Conn, connection, targetCredential, wantError string) {
	t.Helper()
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", connection,
		"--source", "postgres:pg-source", "--destination", "postgres:"+targetCredential,
		"--stream", "public.bootstrap_events", "--sync-mode", "incremental_upsert",
		"--cursor", "sequence", "--primary-key", "id", "--table", connection,
		"--root", root, "--json")
	approval := preparePostgresTransportApproval(t, binary, root, connection, "public.bootstrap_events")
	beforeCounts := postgresTransportBusinessCounts(t, ctx, target)
	beforeStates := postgresTransportStreamStates(t, root)
	output, err := runPostgresTransportApproval(binary, root, connection, "public.bootstrap_events", 1000, approval)
	if err == nil || !strings.Contains(output, wantError) {
		t.Fatalf("destination refusal %q = (%v, %s), want typed refusal containing %q", connection, err, output, wantError)
	}
	if afterCounts := postgresTransportBusinessCounts(t, ctx, target); !samePostgresTransportCounts(afterCounts, beforeCounts) {
		t.Fatalf("destination refusal %q changed target rows: before=%v after=%v", connection, beforeCounts, afterCounts)
	}
	if afterStates := postgresTransportStreamStates(t, root); afterStates != beforeStates {
		t.Fatalf("destination refusal %q advanced checkpoint: before=%s after=%s", connection, beforeStates, afterStates)
	}
}

func samePostgresTransportCounts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func postgresTransportBusinessCounts(t *testing.T, ctx context.Context, target *pgx.Conn) []int {
	t.Helper()
	rows, err := target.Query(ctx, `SELECT n.nspname, c.relname
FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname LIKE 'pm_%' AND c.relkind IN ('r','p') AND c.relname NOT LIKE '\_\_%' ESCAPE '\'
ORDER BY n.nspname, c.relname`)
	if err != nil {
		t.Fatalf("discover managed business relations: %v", err)
	}
	var relations [][2]string
	for rows.Next() {
		var schema, relation string
		if err := rows.Scan(&schema, &relation); err != nil {
			t.Fatal(err)
		}
		relations = append(relations, [2]string{schema, relation})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	var counts []int
	for _, relation := range relations {
		var count int
		if err := target.QueryRow(ctx, "SELECT count(*) FROM "+pgx.Identifier{relation[0], relation[1]}.Sanitize()).Scan(&count); err != nil {
			t.Fatal(err)
		}
		counts = append(counts, count)
	}
	sort.Ints(counts)
	return counts
}

func postgresTransportLabelCount(t *testing.T, ctx context.Context, target *pgx.Conn, label string) int {
	t.Helper()
	rows, err := target.Query(ctx, `SELECT n.nspname, c.relname
	FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
	JOIN pg_catalog.pg_attribute a ON a.attrelid = c.oid AND a.attname = 'label' AND NOT a.attisdropped
	WHERE n.nspname LIKE 'pm_%' AND c.relkind IN ('r','p') AND c.relname NOT LIKE '\_\_%' ESCAPE '\'
	ORDER BY n.nspname, c.relname`)
	if err != nil {
		t.Fatalf("discover managed label relations: %v", err)
	}
	var relations [][2]string
	for rows.Next() {
		var schema, relation string
		if err := rows.Scan(&schema, &relation); err != nil {
			t.Fatal(err)
		}
		relations = append(relations, [2]string{schema, relation})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	total := 0
	for _, relation := range relations {
		var count int
		query := "SELECT count(*) FROM " + pgx.Identifier{relation[0], relation[1]}.Sanitize() + " WHERE label = $1"
		if err := target.QueryRow(ctx, query, label).Scan(&count); err != nil {
			t.Fatal(err)
		}
		total += count
	}
	return total
}

func mustPostgresTransportPM(t *testing.T, binary, stdin string, args ...string) string {
	t.Helper()
	output, err := runTransportPM(binary, stdin, args...)
	if err != nil {
		t.Fatalf("pm %s failed: %v\n%s", transportCommandName(args), err, output)
	}
	return output
}

func postgresTransportTargetState(t *testing.T, ctx context.Context, target *pgx.Conn) (string, string, int, string) {
	t.Helper()
	var schema, relation string
	err := target.QueryRow(ctx, `SELECT n.nspname, c.relname
FROM pg_catalog.pg_class c JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname LIKE 'pm_%' AND c.relkind IN ('r','p') AND c.relname NOT LIKE '\_\_%' ESCAPE '\'
ORDER BY n.nspname, c.relname LIMIT 1`).Scan(&schema, &relation)
	if err != nil {
		t.Fatalf("discover managed target relation: %v", err)
	}
	qualified := pgx.Identifier{schema, relation}.Sanitize()
	var count int
	if err := target.QueryRow(ctx, "SELECT count(*) FROM "+qualified).Scan(&count); err != nil {
		t.Fatalf("count managed target rows: %v", err)
	}
	ledger := pgx.Identifier{schema, "__polymetrics_delivery_ledger"}.Sanitize()
	var delivery string
	if err := target.QueryRow(ctx, "SELECT delivery_id FROM "+ledger+" LIMIT 1").Scan(&delivery); err != nil {
		t.Fatalf("read managed target delivery receipt: %v", err)
	}
	return schema, relation, count, delivery
}

func assertPostgresTransportWarehouse(t *testing.T, root string) {
	t.Helper()
	want := map[string]bool{".jsonl": false, ".parquet": false, ".json": false}
	err := filepath.Walk(filepath.Join(root, ".polymetrics", "warehouse"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if _, ok := want[filepath.Ext(path)]; ok {
				want[filepath.Ext(path)] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inspect connection warehouse: %v", err)
	}
	for extension, found := range want {
		if !found {
			t.Fatalf("connection warehouse has no %s artifact", extension)
		}
	}
}
