package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPMBinaryExecutesPostgresFixturePollingResume exercises the compiled pm
// entry point and App.Open composition, rather than hand-constructing the
// PostgreSQL runner. Fixture mode makes the reach proof hermetic while still
// proving that the definition-selected source invokes #3858 and persists its
// tuple checkpoint through the ordinary warehouse transport route.
func TestPMBinaryExecutesPostgresFixturePollingResume(t *testing.T) {
	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	run := func(args ...string) string {
		t.Helper()
		output, err := runTransportPM(binary, "", args...)
		if err != nil {
			t.Fatalf("pm %s failed: %v\n%s", transportCommandName(args), err, output)
		}
		return output
	}

	run("init", "--root", root, "--json")
	run(
		"credentials", "add", "postgres-fixture", "--connector", "postgres",
		"--config", "mode=fixture", "--config", "host=fixture.internal",
		"--config", "database=analytics", "--config", "username=reader", "--config", "sslmode=require",
		"--root", root, "--json",
	)
	run(
		"credentials", "add", "warehouse-local", "--connector", "warehouse",
		"--config", "path="+filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json",
	)
	run(
		"connections", "create", "postgres-fixture-polling",
		"--source", "postgres:postgres-fixture", "--destination", "warehouse:warehouse-local",
		"--stream", "public.users", "--sync-mode", "incremental_dedupe",
		"--cursor", "updated_at", "--primary-key", "id", "--table", "fixture_users",
		"--root", root, "--json",
	)

	first := decodePostgresPollingRun(t, run("etl", "run", "--connection", "postgres-fixture-polling", "--stream", "public.users", "--batch-size", "2", "--root", root, "--json"))
	if first.Status != "completed" || first.RecordsRead != 3 || first.RecordsLoaded != 3 || first.BatchCount != 2 {
		t.Fatalf("first fixture polling run = %#v, want completed two-page three-row transport", first)
	}
	state := readPostgresPollingState(t, root)
	if !strings.Contains(state, `"mechanism": "polling_watermark"`) {
		t.Fatalf("first fixture polling state did not persist the shared checkpoint: %s", state)
	}
	if strings.Contains(state, "postgres_bounded_full_snapshot") {
		t.Fatalf("first fixture polling state retained private snapshot checkpoint: %s", state)
	}

	second := decodePostgresPollingRun(t, run("etl", "run", "--connection", "postgres-fixture-polling", "--stream", "public.users", "--batch-size", "2", "--root", root, "--json"))
	if second.Status != "completed" || second.RecordsRead != 0 || second.RecordsLoaded != 0 || second.BatchCount != 1 {
		t.Fatalf("resumed fixture polling run = %#v, want one acknowledged zero-row page", second)
	}
}

// TestPMBinaryRefusesPostgresFixturePollingUnknownStreamCursorBeforePageRead
// proves the command-side catalog-validation path: an invalid per-stream
// cursor cannot open a page, create a warehouse artifact, or mutate a
// checkpoint. (Connection creation fills an omitted cursor from the source
// catalog; the lower-level source test separately covers that pre-I/O
// omission refusal.)
func TestPMBinaryRefusesPostgresFixturePollingUnknownStreamCursorBeforePageRead(t *testing.T) {
	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	run := func(args ...string) {
		t.Helper()
		if output, err := runTransportPM(binary, "", args...); err != nil {
			t.Fatalf("pm %s failed: %v\n%s", transportCommandName(args), err, output)
		}
	}
	run("init", "--root", root, "--json")
	run("credentials", "add", "postgres-fixture", "--connector", "postgres", "--config", "mode=fixture", "--config", "host=fixture.internal", "--config", "database=analytics", "--config", "username=reader", "--root", root, "--json")
	run("credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path="+filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json")

	run(
		"connections", "create", "postgres-fixture-missing-cursor",
		"--source", "postgres:postgres-fixture", "--destination", "warehouse:warehouse-local",
		"--stream", "public.users", "--sync-mode", "incremental_dedupe", "--cursor", "not_a_column", "--primary-key", "id", "--table", "fixture_users",
		"--root", root, "--json",
	)
	output, err := runTransportPM(binary, "", "etl", "run", "--connection", "postgres-fixture-missing-cursor", "--stream", "public.users", "--batch-size", "2", "--root", root, "--json")
	if err == nil || !strings.Contains(output, "postgres polling cursor field is absent from the selected relation") {
		t.Fatalf("unknown PostgreSQL stream cursor run = (%v, %s), want pre-page catalog refusal", err, output)
	}
	state := readPostgresPollingState(t, root)
	if strings.Contains(state, "polling_watermark") || strings.Contains(state, `"postgres-fixture-missing-cursor:public.users": {`) {
		t.Fatalf("unknown-cursor refusal wrote a source page or checkpoint: %s", state)
	}
}

type postgresPollingRun struct {
	Status        string `json:"status"`
	RecordsRead   int    `json:"records_read"`
	RecordsLoaded int    `json:"records_loaded"`
	BatchCount    int    `json:"batch_count"`
}

func decodePostgresPollingRun(t *testing.T, output string) postgresPollingRun {
	t.Helper()
	var envelope struct {
		Kind string             `json:"kind"`
		Run  postgresPollingRun `json:"run"`
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode PostgreSQL polling binary output: %v\n%s", err, output)
	}
	if envelope.Kind != "ETLRun" {
		t.Fatalf("PostgreSQL polling binary output kind = %q, want ETLRun", envelope.Kind)
	}
	return envelope.Run
}

func readPostgresPollingState(t *testing.T, root string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read PostgreSQL polling project state: %v", err)
	}
	return string(contents)
}
