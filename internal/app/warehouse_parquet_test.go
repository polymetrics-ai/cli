package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	_ "github.com/marcboeker/go-duckdb"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/warehouse"
)

// readParquetFile reads a materialized table with a reader that knows nothing
// about pm. A test that read it back through the same code that wrote it would
// prove only self-consistency; this proves the file on disk is really Parquet
// and really holds the rows.
func readParquetFile(t *testing.T, path string) []map[string]any {
	t.Helper()
	db, err := sql.Open("duckdb", "")
	if err != nil {
		t.Fatalf("open duckdb: %v", err)
	}
	defer func() { _ = db.Close() }()
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM read_parquet('%s')", strings.ReplaceAll(path, "'", "''")))
	if err != nil {
		t.Fatalf("read_parquet(%s): %v", path, err)
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("columns: %v", err)
	}
	out := make([]map[string]any, 0)
	for rows.Next() {
		holders := make([]any, len(cols))
		targets := make([]any, len(cols))
		for i := range holders {
			targets[i] = &holders[i]
		}
		if err := rows.Scan(targets...); err != nil {
			t.Fatalf("scan: %v", err)
		}
		record := make(map[string]any, len(cols))
		for i, col := range cols {
			record[col] = holders[i]
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate: %v", err)
	}
	return out
}

func parquetRowIDs(rows []map[string]any) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, toComparableString(row["id"]))
	}
	sort.Strings(ids)
	return ids
}

// locateTable resolves the one materialized table a reader would read.
func locateTable(t *testing.T, root, table string) warehouse.Table {
	t.Helper()
	located, err := warehouse.FindTable(root, table, "")
	if err != nil {
		t.Fatalf("FindTable(%s): %v", table, err)
	}
	return located
}

// TestWarehouseMaterializesTablesAsParquet is the primary contract: a sync
// leaves behind a real Parquet file holding the synced rows, at the path the
// per-connection layout resolves to. It asserts on the records that come back
// out of the file, never on the sync's exit status — a status check would have
// passed against both defects #3901 fixed.
func TestWarehouseMaterializesTablesAsParquet(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("parquet_materialization", []connectors.Record{
		{"id": "r1", "name": "Ada", "status": "active", "updated_at": "2026-08-06T00:00:00Z"},
		{"id": "r2", "name": "Alan", "status": "churned", "updated_at": "2026-08-06T00:00:01Z"},
	})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10}); err != nil {
		t.Fatalf("RunETL() error = %v", err)
	}

	root := a.warehouseRoot()
	located := locateTable(t, root, "records")
	if filepath.Ext(located.Path) != warehouse.TableFileExt {
		t.Fatalf("materialized table %s has extension %q, want %q", located.Path, filepath.Ext(located.Path), warehouse.TableFileExt)
	}
	info, err := os.Stat(located.Path)
	if err != nil {
		t.Fatalf("stat table: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("materialized table %s is a directory; a table is one Parquet file", located.Path)
	}

	got := parquetRowIDs(readParquetFile(t, located.Path))
	if want := []string{"r1", "r2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("parquet rows = %v, want %v", got, want)
	}

	// The WAL is the source of truth and must stay newline-delimited JSON: it
	// is opened O_APPEND and fsynced per batch, and Parquet cannot be appended
	// to once closed. This is what makes the table format switchable at all.
	walMatches, err := filepath.Glob(filepath.Join(filepath.Dir(filepath.Dir(located.Path)), warehouse.WALDirName, "*"))
	if err != nil {
		t.Fatalf("glob wal: %v", err)
	}
	if len(walMatches) == 0 {
		t.Fatal("no write-ahead log was written")
	}
	for _, path := range walMatches {
		if filepath.Ext(path) != ".jsonl" {
			t.Fatalf("write-ahead log %s is not JSONL; the WAL must stay appendable", path)
		}
	}
}

// TestTablePathIsASingleParquetFile pins the settled layout decision at the
// layer that owns it. Table-as-directory was measured and rejected: it buys no
// read or write parallelism at our scale and cannot be swapped into place
// atomically, which opens a window where a reader sees no table at all while
// its rows sit on disk.
func TestTablePathIsASingleParquetFile(t *testing.T) {
	location, err := warehouse.LocationFor(t.TempDir(), "ws1", "sample", "conn1", "acme")
	if err != nil {
		t.Fatalf("LocationFor() error = %v", err)
	}
	path, err := location.TablePath("records")
	if err != nil {
		t.Fatalf("TablePath() error = %v", err)
	}
	want := filepath.Join(location.ConnectionDir, warehouse.TablesDirName, "records"+warehouse.TableFileExt)
	if path != want {
		t.Fatalf("TablePath() = %q, want %q", path, want)
	}
}

// TestAppendModeRebuildsFromWALWithoutDuplicating covers the one write path
// that was not already wholesale. Append modes streamed into the table
// O_APPEND; Parquet cannot be appended to, so they now rebuild from the WAL.
// The risk that introduces is duplication, so that is what is asserted.
func TestAppendModeRebuildsFromWALWithoutDuplicating(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("parquet_append_rebuild", nil)
	a, connection := setupSyncModeApp(t, source, "incremental_append")

	source.records = []connectors.Record{
		{"id": "r1", "name": "Ada", "updated_at": "2026-08-06T00:00:00Z"},
	}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10}); err != nil {
		t.Fatalf("first RunETL() error = %v", err)
	}
	source.records = []connectors.Record{
		{"id": "r2", "name": "Alan", "updated_at": "2026-08-06T00:00:02Z"},
	}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10}); err != nil {
		t.Fatalf("second RunETL() error = %v", err)
	}

	located := locateTable(t, a.warehouseRoot(), "records")
	got := parquetRowIDs(readParquetFile(t, located.Path))
	if want := []string{"r1", "r2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("append-mode rows after two syncs = %v, want %v exactly once each", got, want)
	}
}

// TestZeroRowSyncMaterializesAReadableEmptyTable keeps the pre-Parquet
// behaviour of a sync that loads nothing: the table still exists, is still
// listed, and reads back as zero rows. Making it vanish instead would tell an
// operator the table does not exist.
func TestZeroRowSyncMaterializesAReadableEmptyTable(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("parquet_zero_rows", nil)
	a, connection := setupSyncModeApp(t, source, "full_refresh_overwrite")
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10}); err != nil {
		t.Fatalf("RunETL() error = %v", err)
	}

	located := locateTable(t, a.warehouseRoot(), "records")
	if filepath.Ext(located.Path) != warehouse.TableFileExt {
		t.Fatalf("empty table %s has extension %q, want %q", located.Path, filepath.Ext(located.Path), warehouse.TableFileExt)
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatalf("QueryTable() error = %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("QueryTable() rows = %d, want 0", len(rows))
	}
}

// TestPreParquetJSONLTableIsRefusedAndLeftOnDisk states what happens to a
// warehouse written by an unreleased pre-Parquet build. Reading the JSONL
// would work today and be silently stale the moment a sync writes the Parquet
// beside it: two files, one table name, one of them wrong. So it is detected
// and refused by name, and nothing is deleted on the operator's behalf.
func TestPreParquetJSONLTableIsRefusedAndLeftOnDisk(t *testing.T) {
	root := t.TempDir()
	location, err := warehouse.LocationFor(root, "ws1", "sample", "conn1", "acme")
	if err != nil {
		t.Fatalf("LocationFor() error = %v", err)
	}
	if err := location.EnsureOwnership(); err != nil {
		t.Fatalf("EnsureOwnership() error = %v", err)
	}
	stale := filepath.Join(location.ConnectionDir, warehouse.TablesDirName, "records.jsonl")
	if err := os.MkdirAll(filepath.Dir(stale), 0o700); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(connectors.Record{"id": "r1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stale, append(payload, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, err = warehouse.Tables(root)
	var legacy *warehouse.LegacyTableFormatError
	if !errors.As(err, &legacy) {
		t.Fatalf("Tables() error = %T %v, want *warehouse.LegacyTableFormatError", err, err)
	}
	if !strings.Contains(legacy.Error(), stale) {
		t.Fatalf("refusal %q does not name the file it found", legacy.Error())
	}
	if _, err := os.Stat(stale); err != nil {
		t.Fatalf("the pre-Parquet table was not left on disk: %v", err)
	}
}

// TestSyncRefusesAPreParquetWarehouseInsteadOfReportingSuccess is the write half
// of the same contract, and it exists because the read half alone was not
// enough: a sync into a warehouse whose reads are refused reported records
// loaded and exit status 0, leaving an operator told the sync worked and told
// the table cannot be read. pm does not write into a warehouse it will not
// read. The legacy flat layout is already refused here for the same reason.
func TestSyncRefusesAPreParquetWarehouseInsteadOfReportingSuccess(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("parquet_write_refusal", []connectors.Record{
		{"id": "r1", "updated_at": "2026-08-06T00:00:00Z"},
	})
	a, connection := setupSyncModeApp(t, source, "full_refresh_overwrite")
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10}); err != nil {
		t.Fatalf("first RunETL() error = %v", err)
	}
	located := locateTable(t, a.warehouseRoot(), "records")

	stale := filepath.Join(filepath.Dir(located.Path), "left_over"+warehouse.LegacyTableFileExt)
	if err := os.WriteFile(stale, []byte(`{"id":"legacy-1"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(stale)
	if err != nil {
		t.Fatal(err)
	}

	_, err = a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10})
	var legacy *warehouse.LegacyTableFormatError
	if !errors.As(err, &legacy) {
		t.Fatalf("RunETL() into a pre-Parquet warehouse error = %T %v, want *warehouse.LegacyTableFormatError", err, err)
	}
	after, err := os.ReadFile(stale)
	if err != nil {
		t.Fatalf("the pre-Parquet table was not left on disk: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("the pre-Parquet table was rewritten: %q -> %q", before, after)
	}
}
