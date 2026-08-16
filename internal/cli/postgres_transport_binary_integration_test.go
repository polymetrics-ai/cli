//go:build databaseintegration

package cli

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	pmapp "polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/native/dbtest"
	"polymetrics.ai/internal/synctransport"
	"polymetrics.ai/internal/warehouse"
)

const (
	postgresTransportImage    = "docker.io/library/postgres:16.10"
	postgresTransportSourceDB = "pm_transport_source"
	postgresTransportTargetDB = "pm_transport_target"
	postgresTransportUser     = "pm_transport"

	githubCommitTransportMaxPagesEnv = "POLYMETRICS_GITHUB_COMMITS_MAX_PAGES"
	githubCommitTransportPageSize    = 100
	githubCommitTransportMinimumRows = 99345

	postgresFastPathProofLogicalBytes = int64(5_000_000_000)
	postgresFastPathPayloadBytes      = 1 << 20
	postgresFastPathRows              = 5120
)

var errGitHubCommitScaleMaxPages = errors.New("GitHub commit scale max pages must be unlimited or a positive integer")

type githubCommitTransportScale struct {
	MaxPages     string
	ExpectedRows int
	MinimumRows  int
}

func githubCommitTransportScaleConfig(raw string) (githubCommitTransportScale, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" || raw == "unlimited" {
		return githubCommitTransportScale{MaxPages: "unlimited", MinimumRows: githubCommitTransportMinimumRows}, nil
	}
	pages, err := strconv.Atoi(raw)
	if err != nil || pages <= 0 || pages > int(^uint(0)>>1)/githubCommitTransportPageSize {
		return githubCommitTransportScale{}, errGitHubCommitScaleMaxPages
	}
	return githubCommitTransportScale{MaxPages: strconv.Itoa(pages), ExpectedRows: pages * githubCommitTransportPageSize}, nil
}

func TestGitHubCommitTransportScaleConfigDefaultFullCertification(t *testing.T) {
	config, err := githubCommitTransportScaleConfig("")
	if err != nil {
		t.Fatalf("default GitHub commit scale config: %v", err)
	}
	if config.MaxPages != "unlimited" || config.ExpectedRows != 0 || config.MinimumRows != 99345 {
		t.Fatalf("default GitHub commit scale config = %#v, want unlimited with 99,345 minimum rows", config)
	}
}

func TestGitHubCommitTransportScaleConfigRejectsInvalidPagesBeforeHarness(t *testing.T) {
	for _, raw := range []string{"0", "-1", "not-a-number"} {
		t.Run(raw, func(t *testing.T) {
			_, err := githubCommitTransportScaleConfig(raw)
			if !errors.Is(err, errGitHubCommitScaleMaxPages) {
				t.Fatalf("githubCommitTransportScaleConfig(%q) error = %T %v, want errGitHubCommitScaleMaxPages before harness startup", raw, err, err)
			}
		})
	}
}

func TestGitHubCommitTransportScaleConfigBoundedPagesProduceExactRecordCounts(t *testing.T) {
	for _, testCase := range []struct {
		name         string
		raw          string
		wantMaxPages string
		wantRows     int
	}{
		{name: "single GitHub page", raw: "1", wantMaxPages: "1", wantRows: 100},
		{name: "ninety thousand commits", raw: "900", wantMaxPages: "900", wantRows: 90000},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			config, err := githubCommitTransportScaleConfig(testCase.raw)
			if err != nil {
				t.Fatalf("githubCommitTransportScaleConfig(%q): %v", testCase.raw, err)
			}
			if config.MaxPages != testCase.wantMaxPages || config.ExpectedRows != testCase.wantRows || config.MinimumRows != 0 {
				t.Fatalf("githubCommitTransportScaleConfig(%q) = %#v, want pages=%q exact_rows=%d", testCase.raw, config, testCase.wantMaxPages, testCase.wantRows)
			}
		})
	}
}

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

// TestPMBinaryPostgresFullOverwriteRetainsEverySourcePage characterizes the
// shipped full-overwrite route with two bounded source pages. Its target query
// is deliberately the final-content assertion: a per-page TRUNCATE leaves
// only the second page and must fail this test rather than be mistaken for a
// successful run.
func TestPMBinaryPostgresFullOverwriteRetainsEverySourcePage(t *testing.T) {
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
			t.Errorf("close full-overwrite PostgreSQL harness: %v", err)
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("start full-overwrite PostgreSQL harness: %v", err)
	}
	admin := waitForPostgresTransport(t, ctx, endpoint, postgresTransportSourceDB, postgresTransportUser)
	defer func() { _ = admin.Close(context.Background()) }()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+postgresTransportTargetDB); err != nil {
		t.Fatalf("create full-overwrite target database: %v", err)
	}
	if _, err := admin.Exec(ctx, `CREATE TABLE public.overwrite_events (id bigint PRIMARY KEY, sequence bigint NOT NULL, label text NOT NULL)`); err != nil {
		t.Fatalf("create full-overwrite source table: %v", err)
	}
	if _, err := admin.Exec(ctx, `INSERT INTO public.overwrite_events (id, sequence, label) VALUES (1, 10, 'page-one-a'), (2, 20, 'page-one-b'), (3, 30, 'page-two')`); err != nil {
		t.Fatalf("seed full-overwrite source rows: %v", err)
	}
	target := waitForPostgresTransport(t, ctx, endpoint, postgresTransportTargetDB, postgresTransportUser)
	defer func() { _ = target.Close(context.Background()) }()

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	mustPostgresTransportPM(t, binary, "", "init", "--root", root, "--json")
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-source", postgresTransportSourceDB)
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-target", postgresTransportTargetDB)
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "postgres-full-overwrite",
		"--source", "postgres:pg-source", "--destination", "postgres:pg-target",
		"--stream", "public.overwrite_events", "--sync-mode", "full_overwrite",
		"--cursor", "sequence", "--primary-key", "id", "--table", "overwrite_events",
		"--root", root, "--json")

	run := runApprovedPostgresTransportBinary(t, binary, root, "postgres-full-overwrite", "public.overwrite_events", 2).Run
	if run.Status != "completed" || run.RecordsRead != 3 || run.RecordsLoaded != 3 {
		t.Fatalf("two-page full-overwrite binary run = %#v, want completed 3-row transfer", run)
	}
	schema, relation, count := postgresTransportTargetRelation(t, ctx, target)
	if count != 3 || postgresTransportFullOverwriteReceiptCount(t, ctx, target, schema) != 1 {
		t.Fatalf("two-page full-overwrite target = %s.%s rows=%d, want all 3 rows and one durable receipt", schema, relation, count)
	}
	qualified := pgx.Identifier{schema, relation}.Sanitize()
	var ids []int64
	if err := target.QueryRow(ctx, "SELECT array_agg(id ORDER BY id) FROM "+qualified).Scan(&ids); err != nil {
		t.Fatalf("read two-page full-overwrite target rows: %v", err)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Fatalf("two-page full-overwrite target IDs = %v, want every source page [1 2 3]", ids)
	}
}

// TestPMBinaryPostgresTransformedFullOverwriteUsesArrowCOPY proves the actual
// shipped transformed route from CLI construction, rather than wiring native
// source/destination objects in a test. It requires two source ranges, a
// realistic typed projection/filter, durable Parquet segments, PostgreSQL's
// binary COPY apply, one publish receipt, and checkpoint-after-readback.
func TestPMBinaryPostgresTransformedFullOverwriteUsesArrowCOPY(t *testing.T) {
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
			t.Errorf("close transformed full-overwrite PostgreSQL harness: %v", err)
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("start transformed full-overwrite PostgreSQL harness: %v", err)
	}
	admin := waitForPostgresTransport(t, ctx, endpoint, postgresTransportSourceDB, postgresTransportUser)
	defer func() { _ = admin.Close(context.Background()) }()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+postgresTransportTargetDB); err != nil {
		t.Fatalf("create transformed full-overwrite target database: %v", err)
	}
	if _, err := admin.Exec(ctx, `CREATE TABLE public.transformed_events (id bigint PRIMARY KEY, sequence bigint NOT NULL, amount bigint NOT NULL, status text NOT NULL, updated_at timestamptz NOT NULL)`); err != nil {
		t.Fatalf("create transformed full-overwrite source table: %v", err)
	}
	if _, err := admin.Exec(ctx, `INSERT INTO public.transformed_events (id, sequence, amount, status, updated_at) VALUES (1, 10, 11, 'new', '2026-08-01T10:00:00Z'), (2, 20, 22, 'skip', '2026-08-02T10:00:00Z'), (3, 30, 33, 'done', '2026-08-03T10:00:00Z'), (4, 40, 44, 'skip', '2026-08-04T10:00:00Z')`); err != nil {
		t.Fatalf("seed transformed full-overwrite source rows: %v", err)
	}
	target := waitForPostgresTransport(t, ctx, endpoint, postgresTransportTargetDB, postgresTransportUser)
	defer func() { _ = target.Close(context.Background()) }()

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	mustPostgresTransportPM(t, binary, "", "init", "--root", root, "--json")
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-source", postgresTransportSourceDB)
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-target", postgresTransportTargetDB)
	transformFile := filepath.Join(t.TempDir(), "transformed-events.json")
	if err := os.WriteFile(transformFile, []byte(`{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"},{"expr":{"date":"updated_at"},"target":"event_date","type":"date"},{"expr":{"cast":{"multiply":["amount",100]}},"target":"amount_cents","type":"int64","rounding":"exact"},{"expr":{"upper":"status"},"target":"status","type":"string"}],"where":{"not_equal":[{"mod":["id",2]},0]}}`), 0o600); err != nil {
		t.Fatalf("write closed transform fixture: %v", err)
	}
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "postgres-transformed-full-overwrite",
		"--source", "postgres:pg-source", "--destination", "postgres:pg-target",
		"--stream", "public.transformed_events", "--sync-mode", "full_overwrite",
		"--cursor", "sequence", "--primary-key", "id", "--table", "transformed_events",
		"--transform-file", transformFile, "--root", root, "--json")

	run := runApprovedPostgresTransportBinary(t, binary, root, "postgres-transformed-full-overwrite", "public.transformed_events", 2).Run
	if run.Status != "completed" || run.RecordsRead != 4 || run.RecordsLoaded != 2 {
		t.Fatalf("transformed two-page full-overwrite binary run = %#v, want completed source=4 transformed=2", run)
	}
	if measurement := run.TransportPhaseMeasurement; measurement == nil || measurement.SourceLogicalBytes <= 0 || measurement.TransformElapsedNanos <= 0 || measurement.ParquetCloseElapsedNanos <= 0 || measurement.BinaryCOPYElapsedNanos <= 0 || measurement.PublishReceiptElapsedNanos <= 0 || measurement.CheckpointElapsedNanos <= 0 || measurement.CriticalPathElapsedNanos <= 0 {
		t.Fatalf("transformed run phase measurement = %#v, want durable logical-byte and all phase counters", measurement)
	}
	schema, relation, count := postgresTransportTargetRelation(t, ctx, target)
	if count != 2 {
		t.Fatalf("transformed target = %s.%s rows=%d, want filtered transformed rows", schema, relation, count)
	}
	qualified := pgx.Identifier{schema, relation}.Sanitize()
	rows, err := target.Query(ctx, "SELECT event_id, event_date::text, amount_cents, status FROM "+qualified+" ORDER BY event_id")
	if err != nil {
		t.Fatalf("read transformed target rows: %v", err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var id, cents int64
		var date string
		var status string
		if err := rows.Scan(&id, &date, &cents, &status); err != nil {
			t.Fatal(err)
		}
		got = append(got, strconv.FormatInt(id, 10)+"/"+date+"/"+strconv.FormatInt(cents, 10)+"/"+status)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if want := []string{"1/2026-08-01/1100/NEW", "3/2026-08-03/3300/DONE"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("transformed durable target rows = %v, want %v", got, want)
	}
	if receipts := postgresTransportFullOverwriteReceiptCount(t, ctx, target, schema); receipts != 1 {
		t.Fatalf("full-overwrite receipt count = %d, want one receipt in publish transaction", receipts)
	}
}

// TestPMBinaryPostgresTransformedFullOverwriteFiveGigabyteProof is the
// captain's opt-in acceptance harness. Its repeated payload is intentionally
// 5+ GB of logical Arrow source bytes while remaining compact on PostgreSQL
// and Zstd storage, so it measures extraction, closed typed transformation,
// Parquet close/fsync, binary COPY, receipt publication, and checkpoint—not a
// host disk exhaustion. Invoke only after reclaiming dangling container images:
//
// POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_PG_FASTPATH_5GB=1 \
// POLYMETRICS_CONTAINER_RUNTIME=docker \
// POLYMETRICS_CONTAINER_ENDPOINT=unix:///.../docker.sock \
//
//	go test -tags=databaseintegration -count=1 -timeout 45m -v ./internal/cli \
//	  -run '^TestPMBinaryPostgresTransformedFullOverwriteFiveGigabyteProof$'
//
// "Input" is source logical bytes: the pre-transform Arrow buffers for the
// projected source fields. It excludes Parquet, pgwire, target storage, and
// checkpoint bytes. Individual phase intervals can overlap; gate rates use
// critical_path_elapsed_ns, the end-to-end wall clock.
func TestPMBinaryPostgresTransformedFullOverwriteFiveGigabyteProof(t *testing.T) {
	if os.Getenv("POLYMETRICS_DATABASE_INTEGRATION") != "1" || os.Getenv("POLYMETRICS_PG_FASTPATH_5GB") != "1" {
		t.Skip("5 GB PostgreSQL fast-path proof is opt-in")
	}
	if available := postgresFastPathAvailableBytes(t); available < 3<<30 {
		t.Fatalf("hard stop: host free space %d below 3 GiB safety reserve", available)
	}
	proof := postgresFastPathProofReport{Status: "starting", InputBytesDefinition: "pre-transform projected source Arrow buffer bytes; excludes Parquet, pgwire, target storage, and checkpoint bytes", GateDecimalMBPerSecond: 200, GateCriticalPathNanos: (25 * time.Second).Nanoseconds()}
	defer func() { postgresFastPathWriteProofReport(t, proof) }()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
	defer cancel()
	harness := newPostgresTransportHarness(t)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Errorf("close 5 GB PostgreSQL proof harness: %v", err)
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("start 5 GB PostgreSQL proof harness: %v", err)
	}
	admin := waitForPostgresTransport(t, ctx, endpoint, postgresTransportSourceDB, postgresTransportUser)
	defer func() { _ = admin.Close(context.Background()) }()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+postgresTransportTargetDB); err != nil {
		t.Fatalf("create 5 GB proof target database: %v", err)
	}
	if _, err := admin.Exec(ctx, `CREATE TABLE public.fastpath_events (id bigint PRIMARY KEY, sequence bigint NOT NULL, amount bigint NOT NULL, status text NOT NULL, updated_at timestamptz NOT NULL, payload text NOT NULL)`); err != nil {
		t.Fatalf("create 5 GB proof source table: %v", err)
	}
	if _, err := admin.Exec(ctx, fmt.Sprintf(`INSERT INTO public.fastpath_events (id, sequence, amount, status, updated_at, payload)
SELECT value, value * 10, value, CASE WHEN value %% 10 = 0 THEN 'skip' ELSE 'active' END,
       '2026-08-01T00:00:00Z'::timestamptz + value * interval '1 second',
       repeat('x', %d)
FROM generate_series(1, %d) AS value`, postgresFastPathPayloadBytes, postgresFastPathRows)); err != nil {
		t.Fatalf("seed 5 GB logical proof source: %v", err)
	}
	proof.Status = "source_seeded"

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	mustPostgresTransportPM(t, binary, "", "init", "--root", root, "--json")
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-source", postgresTransportSourceDB)
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-target", postgresTransportTargetDB)

	identity := postgresFastPathProofPlan(t, `{"version":1,"select":[{"source":"id","target":"id","type":"int64"},{"source":"sequence","target":"sequence","type":"int64"},{"source":"amount","target":"amount","type":"int64"},{"source":"status","target":"status","type":"string"},{"source":"updated_at","target":"updated_at","type":"timestamp"},{"source":"payload","target":"payload","type":"string"}]}`)
	realistic := postgresFastPathProofPlan(t, `{"version":1,"select":[{"source":"id","target":"event_id","type":"int64"},{"source":"sequence","target":"sequence","type":"int64"},{"expr":{"date":"updated_at"},"target":"event_date","type":"date"},{"expr":{"cast":{"multiply":["amount",100]}},"target":"amount_cents","type":"int64","rounding":"exact"},{"expr":{"upper":"status"},"target":"status","type":"string"},{"source":"payload","target":"payload","type":"string"}],"where":{"not_equal":[{"mod":["id",10]},0]}}`)

	identityRun, err := postgresFastPathProofRun(ctx, binary, root, "postgres-fastpath-identity", identity, "fastpath_identity")
	if err != nil {
		proof.Failure = err.Error()
		t.Fatal(err)
	}
	proof.Identity = identityRun
	proof.Status = "identity_completed"
	t.Logf("FASTPATH_PROOF identity input_bytes=%d decimal_mb_s=%.2f mib_s=%.2f wall_ns=%d source_ns=%d transform_ns=%d parquet_ns=%d copy_ns=%d index_constraint_ns=%d publish_ns=%d checkpoint_ns=%d peak_host_disk_used=%d min_host_free=%d", identityRun.Measurement.SourceLogicalBytes, identityRun.Measurement.InputDecimalMBPerSecond, identityRun.Measurement.InputMiBPerSecond, identityRun.Measurement.CriticalPathElapsedNanos, identityRun.Measurement.SourceReadElapsedNanos, identityRun.Measurement.TransformElapsedNanos, identityRun.Measurement.ParquetCloseElapsedNanos, identityRun.Measurement.BinaryCOPYElapsedNanos, identityRun.Measurement.IndexConstraintBuildElapsedNanos, identityRun.Measurement.PublishReceiptElapsedNanos, identityRun.Measurement.CheckpointElapsedNanos, identityRun.PeakDiskUsed, identityRun.MinimumFree)
	realisticRun, err := postgresFastPathProofRun(ctx, binary, root, "postgres-fastpath-realistic", realistic, "fastpath_realistic")
	if err != nil {
		proof.Failure = err.Error()
		t.Fatal(err)
	}
	proof.Realistic = realisticRun
	proof.Status = "realistic_completed"
	if realisticRun.Measurement.SourceLogicalBytes < postgresFastPathProofLogicalBytes {
		proof.Status = "gate_missed"
		proof.Failure = fmt.Sprintf("realistic input logical bytes %d below required %d", realisticRun.Measurement.SourceLogicalBytes, postgresFastPathProofLogicalBytes)
		t.Fatalf("realistic input logical bytes = %d, want at least %d", realisticRun.Measurement.SourceLogicalBytes, postgresFastPathProofLogicalBytes)
	}
	if realisticRun.Measurement.CriticalPathElapsedNanos > (25*time.Second).Nanoseconds() || realisticRun.Measurement.InputDecimalMBPerSecond < 200 {
		proof.Status = "gate_missed"
		proof.Failure = fmt.Sprintf("realistic 5 GB gate missed: %.2f decimal MB/s, %.2f MiB/s, wall_ns=%d", realisticRun.Measurement.InputDecimalMBPerSecond, realisticRun.Measurement.InputMiBPerSecond, realisticRun.Measurement.CriticalPathElapsedNanos)
		t.Fatalf("realistic 5 GB gate missed: input_bytes=%d wall_ns=%d decimal_mb_s=%.2f mib_s=%.2f; phase source=%d transform=%d parquet=%d copy=%d publish=%d checkpoint=%d",
			realisticRun.Measurement.SourceLogicalBytes, realisticRun.Measurement.CriticalPathElapsedNanos, realisticRun.Measurement.InputDecimalMBPerSecond, realisticRun.Measurement.InputMiBPerSecond,
			realisticRun.Measurement.SourceReadElapsedNanos, realisticRun.Measurement.TransformElapsedNanos, realisticRun.Measurement.ParquetCloseElapsedNanos, realisticRun.Measurement.BinaryCOPYElapsedNanos, realisticRun.Measurement.PublishReceiptElapsedNanos, realisticRun.Measurement.CheckpointElapsedNanos)
	}
	t.Logf("FASTPATH_PROOF realistic input_bytes=%d decimal_mb_s=%.2f mib_s=%.2f wall_ns=%d source_ns=%d transform_ns=%d parquet_ns=%d copy_ns=%d index_constraint_ns=%d publish_ns=%d checkpoint_ns=%d peak_host_disk_used=%d min_host_free=%d", realisticRun.Measurement.SourceLogicalBytes, realisticRun.Measurement.InputDecimalMBPerSecond, realisticRun.Measurement.InputMiBPerSecond, realisticRun.Measurement.CriticalPathElapsedNanos, realisticRun.Measurement.SourceReadElapsedNanos, realisticRun.Measurement.TransformElapsedNanos, realisticRun.Measurement.ParquetCloseElapsedNanos, realisticRun.Measurement.BinaryCOPYElapsedNanos, realisticRun.Measurement.IndexConstraintBuildElapsedNanos, realisticRun.Measurement.PublishReceiptElapsedNanos, realisticRun.Measurement.CheckpointElapsedNanos, realisticRun.PeakDiskUsed, realisticRun.MinimumFree)
	proof.Status = "passed"
}

type postgresFastPathProofResult struct {
	Measurement  *pmapp.TransportPhaseMeasurement `json:"measurement"`
	PeakDiskUsed int64                            `json:"peak_host_disk_used_bytes"`
	MinimumFree  int64                            `json:"minimum_host_free_bytes"`
}

// postgresFastPathProofReport is written only when the opt-in acceptance
// command sets POLYMETRICS_PG_FASTPATH_5GB_REPORT. It makes the two mappings'
// phase results reviewable without relying on a terminal scrollback, while the
// production run itself still persists its measurement before cleanup.
type postgresFastPathProofReport struct {
	Status                 string                      `json:"status"`
	Failure                string                      `json:"failure,omitempty"`
	InputBytesDefinition   string                      `json:"input_bytes_definition"`
	GateDecimalMBPerSecond float64                     `json:"gate_decimal_mb_per_second"`
	GateCriticalPathNanos  int64                       `json:"gate_critical_path_ns"`
	Identity               postgresFastPathProofResult `json:"identity"`
	Realistic              postgresFastPathProofResult `json:"realistic"`
}

func postgresFastPathWriteProofReport(t *testing.T, report postgresFastPathProofReport) {
	t.Helper()
	path := strings.TrimSpace(os.Getenv("POLYMETRICS_PG_FASTPATH_5GB_REPORT"))
	if path == "" {
		return
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("encode 5 GB fast-path proof report: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create 5 GB fast-path proof report directory: %v", err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		t.Fatalf("write 5 GB fast-path proof report: %v", err)
	}
}

func postgresFastPathProofPlan(t *testing.T, raw string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "transform.json")
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatalf("write fast-path proof transform: %v", err)
	}
	return path
}

func postgresFastPathProofRun(ctx context.Context, binary, root, connection, transformFile, table string) (postgresFastPathProofResult, error) {
	var result postgresFastPathProofResult
	if output, err := runTransportPM(binary, "",
		"connections", "create", connection,
		"--source", "postgres:pg-source", "--destination", "postgres:pg-target",
		"--stream", "public.fastpath_events", "--sync-mode", "full_overwrite",
		"--cursor", "sequence", "--primary-key", "id", "--table", table,
		"--transform-file", transformFile, "--root", root, "--json"); err != nil {
		return result, fmt.Errorf("create %s fast-path connection: %w", connection, err)
	} else if strings.TrimSpace(output) == "" {
		return result, fmt.Errorf("create %s fast-path connection produced no output", connection)
	}
	planOutput, err := runTransportPM(binary, "", "etl", "transport", "postgres-managed-target", "plan", "--connection", connection, "--stream", "public.fastpath_events", "--root", root, "--json")
	if err != nil {
		return result, fmt.Errorf("plan %s fast-path run: %w", connection, err)
	}
	var planned struct {
		Plan struct {
			ID string `json:"id"`
		} `json:"plan"`
	}
	if err := json.Unmarshal([]byte(planOutput), &planned); err != nil || planned.Plan.ID == "" {
		return result, fmt.Errorf("decode %s fast-path plan", connection)
	}
	preview, err := runTransportPM(binary, "", "etl", "transport", "postgres-managed-target", "preview", planned.Plan.ID, "--root", root)
	if err != nil {
		return result, fmt.Errorf("preview %s fast-path run: %w", connection, err)
	}
	const marker = "Approval token: "
	index := strings.Index(preview, marker)
	if index < 0 {
		return result, fmt.Errorf("preview %s fast-path run issued no approval token", connection)
	}
	token := strings.TrimSpace(strings.SplitN(preview[index+len(marker):], "\n", 2)[0])
	if token == "" {
		return result, fmt.Errorf("preview %s fast-path run issued an empty approval token", connection)
	}
	monitor, err := newPostgresFastPathDiskMonitor()
	if err != nil {
		return result, fmt.Errorf("monitor %s fast-path run disk: %w", connection, err)
	}
	runOutput, err := runPostgresTransportApproval(binary, root, connection, "public.fastpath_events", 128, approvedPostgresTransportRun{PlanID: planned.Plan.ID, Token: token})
	peakUsed, minimumFree, monitorErr := monitor.Stop()
	if monitorErr != nil {
		return result, fmt.Errorf("monitor %s fast-path run disk: %w", connection, monitorErr)
	}
	if minimumFree < 3<<30 {
		return result, fmt.Errorf("hard stop: host free space %d fell below 3 GiB safety reserve", minimumFree)
	}
	if err != nil {
		return result, fmt.Errorf("run %s fast-path transport: %w", connection, err)
	}
	var envelope struct {
		Run postgresTransportRun `json:"run"`
	}
	if err := json.Unmarshal([]byte(runOutput), &envelope); err != nil {
		return result, fmt.Errorf("decode %s fast-path run", connection)
	}
	approved := envelope.Run
	if approved.Status != "completed" || approved.RecordsRead != postgresFastPathRows || approved.TransportPhaseMeasurement == nil {
		return result, fmt.Errorf("%s fast-path run was not completed with a durable measurement", connection)
	}
	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("%s fast-path proof context ended: %w", connection, err)
	}
	return postgresFastPathProofResult{Measurement: approved.TransportPhaseMeasurement, PeakDiskUsed: peakUsed, MinimumFree: minimumFree}, nil
}

type postgresFastPathDiskMonitor struct {
	stop    chan struct{}
	done    chan struct{}
	mu      sync.Mutex
	total   int64
	minimum int64
	err     error
	once    sync.Once
}

func newPostgresFastPathDiskMonitor() (*postgresFastPathDiskMonitor, error) {
	total, available, err := postgresFastPathDiskSpace()
	if err != nil {
		return nil, err
	}
	monitor := &postgresFastPathDiskMonitor{stop: make(chan struct{}), done: make(chan struct{}), total: total, minimum: available}
	go func() {
		defer close(monitor.done)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-monitor.stop:
				return
			case <-ticker.C:
				_, available, err := postgresFastPathDiskSpace()
				monitor.mu.Lock()
				if err != nil && monitor.err == nil {
					monitor.err = err
				}
				if err == nil && available < monitor.minimum {
					monitor.minimum = available
				}
				monitor.mu.Unlock()
			}
		}
	}()
	return monitor, nil
}

func (m *postgresFastPathDiskMonitor) Stop() (int64, int64, error) {
	m.once.Do(func() {
		close(m.stop)
		<-m.done
	})
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.total - m.minimum, m.minimum, m.err
}

func postgresFastPathAvailableBytes(t *testing.T) int64 {
	t.Helper()
	_, available, err := postgresFastPathDiskSpace()
	if err != nil {
		t.Fatalf("measure host disk for fast-path proof: %v", err)
	}
	return available
}

func postgresFastPathDiskSpace() (int64, int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(".", &stat); err != nil {
		return 0, 0, err
	}
	total := int64(uint64(stat.Blocks) * uint64(stat.Bsize))
	available := int64(uint64(stat.Bavail) * uint64(stat.Bsize))
	return total, available, nil
}

// TestPMBinaryExecutesAuthenticatedGitHubWarehousePostgres proves the
// representative API-to-database route required by #3982 and the merge gate
// on PR #4167. A built pm extracts exactly fifty issues from rails/rails using
// a credential supplied only through pm credentials add, materializes the
// bounded page in connection-owned Parquet, and dispatches the registered
// PostgreSQL managed target. Fresh filesystem and PostgreSQL reads establish
// the row count and types independently of the run result held in memory.
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
	defer func() {
		if target != nil {
			_ = target.Close(context.Background())
		}
	}()

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "github-project")
	mustPostgresTransportPM(t, binary, "", "init", "--root", root, "--json")
	mustPostgresTransportPM(t, binary, "",
		"credentials", "add", "github-live", "--connector", "github",
		"--config", "owner=rails", "--config", "repo=rails", "--config", "auth_type=token",
		"--config", "rate_limit_account=authenticated-transport-proof",
		"--from-env", "token=POLYMETRICS_GITHUB_TOKEN", "--root", root, "--json")
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-target", postgresTransportTargetDB)
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "github-to-postgres",
		"--source", "github:github-live", "--destination", "postgres:pg-target",
		"--stream", "issues", "--sync-mode", "incremental_upsert",
		"--cursor", "updated_at", "--primary-key", "node_id", "--table", "issues",
		"--root", root, "--json")

	// A real, persisted but unpreviewed plan supplies no mutation authority.
	// Both the shipped command and a fresh app.Open production composition must
	// refuse before the API source, warehouse checkpoint, or target is touched.
	unapproved := planPostgresTransportApproval(t, binary, root, "github-to-postgres", "issues")
	beforeUnapprovedCounts := postgresTransportBusinessCounts(t, ctx, target)
	beforeUnapprovedState := postgresTransportStreamStates(t, root)
	unapprovedOutput, unapprovedErr := runTransportPM(binary, "",
		"etl", "run", "--connection", "github-to-postgres", "--stream", "issues",
		"--batch-size", "50", "--root", root, "--json")
	if unapprovedErr == nil || !strings.Contains(unapprovedOutput, pmapp.ErrPostgresManagedTargetApprovalRequired.Error()) {
		t.Fatalf("unapproved binary run = (%v, %s), want managed-target approval refusal", unapprovedErr, unapprovedOutput)
	}
	unapprovedApp, err := pmapp.Open(root)
	if err != nil {
		t.Fatalf("open production composition for typed unapproved refusal: %v", err)
	}
	_, err = unapprovedApp.RunETL(ctx, pmapp.RunETLRequest{
		Connection: "github-to-postgres", Stream: "issues", BatchSize: 50,
		DestinationApproval: synctransport.DestinationApproval{PlanID: unapproved.PlanID},
	})
	if !errors.Is(err, pmapp.ErrPostgresManagedTargetApprovalRequired) {
		t.Fatalf("unapproved app.Open run error = %T %v, want ErrPostgresManagedTargetApprovalRequired", err, err)
	}
	if after := postgresTransportBusinessCounts(t, ctx, target); !samePostgresTransportCounts(after, beforeUnapprovedCounts) {
		t.Fatalf("unapproved plan changed target rows: before=%v after=%v", beforeUnapprovedCounts, after)
	}
	if after := postgresTransportStreamStates(t, root); after != beforeUnapprovedState {
		t.Fatalf("unapproved plan advanced checkpoint: before=%s after=%s", beforeUnapprovedState, after)
	}

	assertCanceledGitHubManagedTargetIsStateFree(t, ctx, binary, root, target)

	approved := runApprovedPostgresTransportBinary(t, binary, root, "github-to-postgres", "issues", 50)
	run := approved.Run
	if run.Status != "completed" || run.RecordsRead != 50 || run.RecordsLoaded != 50 {
		t.Fatalf("authenticated rails/rails binary run = %#v, want exact 50-row transfer", run)
	}
	assertPostgresTransportWarehouse(t, root)
	warehouseCount := postgresTransportWarehouseIssueRows(t, root, "rails/rails")
	if warehouseCount != 50 {
		t.Fatalf("independent rails/rails Parquet row count = %d, want 50", warehouseCount)
	}
	checkpointAfterSuccess := postgresTransportStreamStates(t, root)
	if checkpointAfterSuccess == beforeUnapprovedState || !strings.Contains(checkpointAfterSuccess, "github-to-postgres") || !strings.Contains(checkpointAfterSuccess, "issues") {
		t.Fatalf("authenticated rails/rails run did not independently advance its checkpoint: %s", checkpointAfterSuccess)
	}

	// Reconnect after the pm process exits so the proof queries PostgreSQL
	// independently rather than relying on driver state from the write path.
	_ = target.Close(context.Background())
	target = waitForPostgresTransport(t, ctx, endpoint, postgresTransportTargetDB, postgresTransportUser)
	schema, relation, count, delivery := postgresTransportTargetState(t, ctx, target)
	if count != warehouseCount || delivery == "" {
		t.Fatalf("GitHub managed target = %s.%s rows=%d delivery_present=%t, want %d durable rows", schema, relation, count, delivery != "", warehouseCount)
	}
	qualified := pgx.Identifier{schema, relation}.Sanitize()
	var numberType, nodeIDType, labelsType, updatedAtType, lockedType string
	if err := target.QueryRow(ctx, "SELECT pg_typeof(number)::text, pg_typeof(node_id)::text, pg_typeof(labels)::text, pg_typeof(updated_at)::text, pg_typeof(locked)::text FROM "+qualified+" LIMIT 1").Scan(&numberType, &nodeIDType, &labelsType, &updatedAtType, &lockedType); err != nil {
		t.Fatalf("read authenticated GitHub target types: %v", err)
	}
	if numberType != "bigint" || nodeIDType != "text" || labelsType != "jsonb" || updatedAtType != "text" || lockedType != "boolean" {
		t.Fatalf("authenticated GitHub target types = number:%s node_id:%s labels:%s updated_at:%s locked:%s", numberType, nodeIDType, labelsType, updatedAtType, lockedType)
	}
	t.Logf("observed rails/rails rows: warehouse_parquet=%d postgres=%d; PostgreSQL types number=%s node_id=%s labels=%s updated_at=%s locked=%s", warehouseCount, count, numberType, nodeIDType, labelsType, updatedAtType, lockedType)
	beforeUnattendedCounts := postgresTransportBusinessCounts(t, ctx, target)

	// The first human token mints a durable, shape-bound authorization. A
	// same-shape run must be reachable through a fresh shipped binary without
	// placing that single-use token back into stdin or argv.
	unattendedOutput, unattendedErr := runTransportPM(binary, "",
		"etl", "run", "--connection", "github-to-postgres", "--stream", "issues",
		"--batch-size", "50", "--approval-plan", approved.PlanID, "--confirm", "destructive",
		"--root", root, "--json")
	if unattendedErr != nil {
		t.Fatalf("same-shape unattended GitHub transport = (%v, %s), want completed durable-authorization run", unattendedErr, unattendedOutput)
	}
	var unattended struct {
		Run postgresTransportRun `json:"run"`
	}
	if err := json.Unmarshal([]byte(unattendedOutput), &unattended); err != nil {
		t.Fatalf("decode unattended GitHub transport run: %v output=%s", err, unattendedOutput)
	}
	if unattended.Run.Status != "completed" || unattended.Run.RecordsRead != 0 || unattended.Run.RecordsLoaded != 0 {
		t.Fatalf("same-shape unattended GitHub transport run = %#v, want completed zero-new-row resume", unattended.Run)
	}
	if after := postgresTransportBusinessCounts(t, ctx, target); !samePostgresTransportCounts(after, beforeUnattendedCounts) {
		t.Fatalf("same-shape unattended transport changed target rows: before=%v after=%v", beforeUnattendedCounts, after)
	}

	beforeReplayCounts := postgresTransportBusinessCounts(t, ctx, target)
	beforeReplayState := postgresTransportStreamStates(t, root)
	replayOutput, replayErr := runPostgresTransportApproval(binary, root, "github-to-postgres", "issues", 50, approved)
	if replayErr == nil || !strings.Contains(replayOutput, "authorization token") || !strings.Contains(replayOutput, "already been consumed") {
		t.Fatalf("consumed GitHub approval replay = (%v, %s), want typed replay refusal", replayErr, replayOutput)
	}
	if strings.Contains(replayOutput, approved.Token) {
		t.Fatal("consumed GitHub approval token appeared in refusal output")
	}
	replayApp, err := pmapp.Open(root)
	if err != nil {
		t.Fatalf("open production composition for typed replay refusal: %v", err)
	}
	_, err = replayApp.RunETL(ctx, pmapp.RunETLRequest{
		Connection: "github-to-postgres", Stream: "issues", BatchSize: 50,
		DestinationApproval: postgresTransportAppApproval(approved),
	})
	var typedReplay *pmapp.AuthorizationTokenReplayError
	if !errors.As(err, &typedReplay) {
		t.Fatalf("consumed app.Open replay error = %T %v, want AuthorizationTokenReplayError", err, err)
	}
	if after := postgresTransportBusinessCounts(t, ctx, target); !samePostgresTransportCounts(after, beforeReplayCounts) {
		t.Fatalf("consumed replay changed target rows: before=%v after=%v", beforeReplayCounts, after)
	}
	if after := postgresTransportStreamStates(t, root); after != beforeReplayState {
		t.Fatalf("consumed replay advanced checkpoint: before=%s after=%s", beforeReplayState, after)
	}
}

// TestPMBinaryExecutesAuthenticatedGitHubCommitsWarehousePostgres is the
// decisive #4171 production proof. It intentionally traverses every declared
// GitHub page with max_pages=unlimited and counts both durable hops
// independently; a one-page default would therefore fail far below the
// 99,345-row certification reference that exposed the admission defect.
func TestPMBinaryExecutesAuthenticatedGitHubCommitsWarehousePostgres(t *testing.T) {
	if os.Getenv("POLYMETRICS_DATABASE_INTEGRATION") != "1" || os.Getenv("POLYMETRICS_GITHUB_INTEGRATION") != "1" {
		t.Skip("authenticated GitHub commits-to-PostgreSQL integration is opt-in")
	}
	if os.Getenv("POLYMETRICS_GITHUB_TOKEN") == "" {
		t.Fatal("POLYMETRICS_GITHUB_INTEGRATION=1 requires POLYMETRICS_GITHUB_TOKEN")
	}
	scale, err := githubCommitTransportScaleConfig(os.Getenv(githubCommitTransportMaxPagesEnv))
	if err != nil {
		t.Fatalf("%s: %v", githubCommitTransportMaxPagesEnv, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	harness := newPostgresTransportHarness(t)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Errorf("close GitHub commits-to-PostgreSQL harness: %v", err)
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("start GitHub commits-to-PostgreSQL harness: %v", err)
	}
	admin := waitForPostgresTransport(t, ctx, endpoint, postgresTransportSourceDB, postgresTransportUser)
	defer func() { _ = admin.Close(context.Background()) }()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+postgresTransportTargetDB); err != nil {
		t.Fatalf("create GitHub commits target database: %v", err)
	}
	target := waitForPostgresTransport(t, ctx, endpoint, postgresTransportTargetDB, postgresTransportUser)
	defer func() { _ = target.Close(context.Background()) }()

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "github-commits-project")
	mustPostgresTransportPM(t, binary, "", "init", "--root", root, "--json")
	mustPostgresTransportPM(t, binary, "",
		"credentials", "add", "github-commits-live", "--connector", "github",
		"--config", "owner=rails", "--config", "repo=rails", "--config", "auth_type=token",
		"--config", "rate_limit_account=authenticated-commits-transport-proof",
		"--config", "max_pages="+scale.MaxPages,
		"--from-env", "token=POLYMETRICS_GITHUB_TOKEN", "--root", root, "--json")
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-target", postgresTransportTargetDB)
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "github-commits-to-postgres",
		"--source", "github:github-commits-live", "--destination", "postgres:pg-target",
		"--stream", "commits", "--sync-mode", "incremental_upsert",
		"--cursor", "commit_committer_date", "--primary-key", "sha", "--table", "commits",
		"--root", root, "--json")
	// Run this defer before the harness and TempDir cleanup defers. A terminal
	// failure must retain its independently named durable counts in the test
	// log; cleanup previously erased the only evidence of where a long run got.
	defer logDurableGitHubCommitTransportPhaseCounts(t, root, "github-commits-to-postgres", "commits")

	approved := runApprovedPostgresTransportBinary(t, binary, root, "github-commits-to-postgres", "commits", 1000)
	run := approved.Run
	if run.Status != "completed" || run.RecordsRead != run.RecordsLoaded {
		t.Fatalf("authenticated rails/rails commits binary run = %#v, want an exact completed transfer", run)
	}
	if scale.ExpectedRows != 0 && run.RecordsRead != scale.ExpectedRows {
		t.Fatalf("authenticated rails/rails commits extracted %d rows with max_pages=%s, want exact %d", run.RecordsRead, scale.MaxPages, scale.ExpectedRows)
	}
	if scale.MinimumRows != 0 && run.RecordsRead < scale.MinimumRows {
		t.Fatalf("authenticated rails/rails commits extracted %d rows, below the %d-row defect reference; max_pages=%s may not have reached the wire", run.RecordsRead, scale.MinimumRows, scale.MaxPages)
	}
	if measurement := run.TransportPhaseMeasurement; measurement == nil || measurement.ExtractedRecords != run.RecordsRead || measurement.WarehouseParquetRecords != run.RecordsRead || measurement.PostgreSQLAppliedRecords != run.RecordsLoaded {
		t.Fatalf("completed GitHub commit phase measurement = %#v, want durable exact extraction/Parquet/PostgreSQL counts", measurement)
	}
	assertPostgresTransportWarehouse(t, root)
	warehouseCount := postgresTransportWarehouseCommitRows(t, root, "rails/rails")
	if warehouseCount != run.RecordsRead {
		t.Fatalf("independent rails/rails commits Parquet rows = %d, want all %d extracted rows", warehouseCount, run.RecordsRead)
	}

	_ = target.Close(context.Background())
	target = waitForPostgresTransport(t, ctx, endpoint, postgresTransportTargetDB, postgresTransportUser)
	schema, relation, targetCount, delivery := postgresTransportTargetState(t, ctx, target)
	if targetCount != run.RecordsRead || delivery == "" {
		t.Fatalf("GitHub commits managed target = %s.%s rows=%d delivery_present=%t, want all %d extracted rows", schema, relation, targetCount, delivery != "", run.RecordsRead)
	}
	t.Logf("observed rails/rails commits with max_pages=%s: extracted=%d warehouse_parquet=%d postgres=%d expected_rows=%d minimum_rows=%d", scale.MaxPages, run.RecordsRead, warehouseCount, targetCount, scale.ExpectedRows, scale.MinimumRows)
}

type postgresTransportRun struct {
	Status                    string                           `json:"status"`
	RecordsRead               int                              `json:"records_read"`
	RecordsLoaded             int                              `json:"records_loaded"`
	TransportPhaseMeasurement *pmapp.TransportPhaseMeasurement `json:"transport_phase_measurement,omitempty"`
}

// logDurableGitHubCommitTransportPhaseCounts reopens persisted state rather
// than trusting the child process's success claim. It intentionally runs while
// the test's project and database harness still exist, on both success and
// failure exits, and logs only counts and elapsed durations.
func logDurableGitHubCommitTransportPhaseCounts(t *testing.T, root, connection, stream string) {
	t.Helper()
	instance, err := pmapp.Open(root)
	if err != nil {
		t.Errorf("reopen GitHub commit project for durable phase counts before cleanup: %v", err)
		return
	}
	stateBytes, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Errorf("read GitHub commit durable state before cleanup: %v", err)
		return
	}
	var persisted struct {
		Runs []pmapp.Run `json:"runs"`
	}
	if err := json.Unmarshal(stateBytes, &persisted); err != nil {
		t.Errorf("decode GitHub commit durable state before cleanup: %v", err)
		return
	}
	for i := len(persisted.Runs) - 1; i >= 0; i-- {
		run := persisted.Runs[i]
		if run.Connection != connection || run.Stream != stream {
			continue
		}
		// Use the separately reopened App snapshot for the reported values; the
		// direct JSON decode above identifies the terminal run even after the
		// child command itself failed before it could print a run envelope.
		reopened, err := instance.GetRun(run.ID)
		if err != nil {
			t.Errorf("read durable GitHub commit run %s before cleanup: %v", run.ID, err)
			return
		}
		measurement := reopened.TransportPhaseMeasurement
		if measurement == nil {
			t.Errorf("durable GitHub commit run %s has no phase measurement before cleanup", run.ID)
			return
		}
		t.Logf("durable rails/rails commits phase counts before cleanup: status=%s extracted=%d warehouse_parquet=%d postgres=%d extract_elapsed_ns=%d warehouse_elapsed_ns=%d postgresql_elapsed_ns=%d", reopened.Status, measurement.ExtractedRecords, measurement.WarehouseParquetRecords, measurement.PostgreSQLAppliedRecords, measurement.ExtractElapsedNanos, measurement.WarehouseElapsedNanos, measurement.PostgreSQLElapsedNanos)
		return
	}
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
	planned := planPostgresTransportApproval(t, binary, root, connection, stream)
	preview := mustPostgresTransportPM(t, binary, "",
		"etl", "transport", "postgres-managed-target", "preview", planned.PlanID,
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
	return approvedPostgresTransportRun{PlanID: planned.PlanID, Token: token}
}

func planPostgresTransportApproval(t *testing.T, binary, root, connection, stream string) approvedPostgresTransportRun {
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
	return approvedPostgresTransportRun{PlanID: planned.Plan.ID}
}

func postgresTransportAppApproval(approval approvedPostgresTransportRun) synctransport.DestinationApproval {
	return synctransport.DestinationApproval{
		PlanID:        approval.PlanID,
		ApprovalToken: approval.Token,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	}
}

func assertCanceledGitHubManagedTargetIsStateFree(t *testing.T, ctx context.Context, binary, root string, target *pgx.Conn) {
	t.Helper()
	requestStarted := make(chan struct{}, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/repos/rails/rails/issues" {
			http.NotFound(response, request)
			return
		}
		select {
		case requestStarted <- struct{}{}:
		default:
		}
		select {
		case <-request.Context().Done():
		case <-release:
		}
	}))
	defer server.Close()
	defer close(release)

	mustPostgresTransportPM(t, binary, "",
		"credentials", "add", "github-cancel", "--connector", "github",
		"--config", "owner=rails", "--config", "repo=rails", "--config", "auth_type=token",
		"--config", "base_url="+server.URL,
		"--config", "rate_limit_account=authenticated-transport-cancel",
		"--from-env", "token=POLYMETRICS_GITHUB_TOKEN", "--root", root, "--json")
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "github-cancel-to-postgres",
		"--source", "github:github-cancel", "--destination", "postgres:pg-target",
		"--stream", "issues", "--sync-mode", "incremental_upsert",
		"--cursor", "updated_at", "--primary-key", "node_id", "--table", "canceled_issues",
		"--root", root, "--json")
	approval := preparePostgresTransportApproval(t, binary, root, "github-cancel-to-postgres", "issues")
	beforeCounts := postgresTransportBusinessCounts(t, ctx, target)
	beforeState := postgresTransportStreamStates(t, root)
	composed, err := pmapp.Open(root)
	if err != nil {
		t.Fatalf("open production composition for cancellation: %v", err)
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, runErr := composed.RunETL(runCtx, pmapp.RunETLRequest{
			Connection: "github-cancel-to-postgres", Stream: "issues", BatchSize: 50,
			DestinationApproval: postgresTransportAppApproval(approval),
		})
		done <- runErr
	}()
	select {
	case <-requestStarted:
		cancel()
	case runErr := <-done:
		t.Fatalf("cancellation run exited before its provider request was in flight: %v", runErr)
	case <-time.After(30 * time.Second):
		t.Fatal("cancellation run did not reach its in-flight provider request")
	}
	select {
	case runErr := <-done:
		if !errors.Is(runErr, context.Canceled) {
			t.Fatalf("in-flight cancellation error = %T %v, want context.Canceled", runErr, runErr)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("in-flight cancellation did not return")
	}
	if after := postgresTransportBusinessCounts(t, ctx, target); !samePostgresTransportCounts(after, beforeCounts) {
		t.Fatalf("in-flight cancellation changed target rows: before=%v after=%v", beforeCounts, after)
	}
	if after := postgresTransportStreamStates(t, root); after != beforeState {
		t.Fatalf("in-flight cancellation advanced checkpoint: before=%s after=%s", beforeState, after)
	}
}

func postgresTransportWarehouseIssueRows(t *testing.T, root, repository string) int {
	t.Helper()
	var tablePath string
	err := filepath.Walk(filepath.Join(root, ".polymetrics", "warehouse"), func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || filepath.Ext(info.Name()) != ".parquet" || !strings.HasPrefix(info.Name(), "transport-stage_") || filepath.Base(filepath.Dir(path)) != "tables" {
			return nil
		}
		if tablePath != "" {
			t.Fatalf("found more than one connection-owned issues Parquet table: %s and %s", tablePath, path)
		}
		tablePath = path
		return nil
	})
	if err != nil {
		t.Fatalf("discover connection-owned issues Parquet: %v", err)
	}
	if tablePath == "" {
		t.Fatal("connection-owned issues Parquet table was not materialized")
	}
	count := 0
	if err := warehouse.ReadTable(context.Background(), tablePath, func(row warehouse.Row) error {
		count++
		if row["repository"] != repository {
			t.Fatalf("Parquet issue %d repository = %#v, want %q", count, row["repository"], repository)
		}
		if nodeID, ok := row["node_id"].(string); !ok || strings.TrimSpace(nodeID) == "" {
			t.Fatalf("Parquet issue %d node_id = %#v, want non-empty text", count, row["node_id"])
		}
		return nil
	}); err != nil {
		t.Fatalf("independently read connection-owned issues Parquet: %v", err)
	}
	return count
}

func postgresTransportWarehouseCommitRows(t *testing.T, root, repository string) int {
	t.Helper()
	var tablePaths []string
	err := filepath.Walk(filepath.Join(root, ".polymetrics", "warehouse"), func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || filepath.Ext(info.Name()) != ".parquet" || !strings.HasPrefix(info.Name(), "transport-stage_") || filepath.Base(filepath.Dir(path)) != "tables" {
			return nil
		}
		tablePaths = append(tablePaths, path)
		return nil
	})
	if err != nil {
		t.Fatalf("discover connection-owned commits Parquet: %v", err)
	}
	if len(tablePaths) == 0 {
		t.Fatal("connection-owned commits Parquet table was not materialized")
	}
	count := 0
	for _, tablePath := range tablePaths {
		if err := warehouse.ReadTable(context.Background(), tablePath, func(row warehouse.Row) error {
			count++
			if row["repository"] != repository {
				t.Fatalf("Parquet commit %d repository = %#v, want %q", count, row["repository"], repository)
			}
			if sha, ok := row["sha"].(string); !ok || strings.TrimSpace(sha) == "" {
				t.Fatalf("Parquet commit %d sha = %#v, want non-empty text", count, row["sha"])
			}
			return nil
		}); err != nil {
			t.Fatalf("independently read connection-owned commits Parquet %q: %v", tablePath, err)
		}
	}
	return count
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
	schema, relation, count := postgresTransportTargetRelation(t, ctx, target)
	ledger := pgx.Identifier{schema, "__polymetrics_delivery_ledger"}.Sanitize()
	var delivery string
	if err := target.QueryRow(ctx, "SELECT delivery_id FROM "+ledger+" LIMIT 1").Scan(&delivery); err != nil {
		t.Fatalf("read managed target delivery receipt: %v", err)
	}
	return schema, relation, count, delivery
}

func postgresTransportTargetRelation(t *testing.T, ctx context.Context, target *pgx.Conn) (string, string, int) {
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
	return schema, relation, count
}

func postgresTransportFullOverwriteReceiptCount(t *testing.T, ctx context.Context, target *pgx.Conn, schema string) int {
	t.Helper()
	receiptTable := pgx.Identifier{schema, "__polymetrics_full_overwrite_receipt"}.Sanitize()
	var receipts int
	if err := target.QueryRow(ctx, "SELECT count(*) FROM "+receiptTable).Scan(&receipts); err != nil {
		t.Fatalf("read full-overwrite receipt count: %v", err)
	}
	return receipts
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
