package warehouse

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
)

// TestWriteTablePreservesConnectorJSONStrings pins the warehouse boundary:
// provider JSON strings remain strings even when their bytes resemble a SQL
// DATE or TIMESTAMP. Reverse ETL validates the reopened row against the
// connector's request schema, so DuckDB type inference must not silently turn
// a documented string field into time.Time inside a nested request object.
func TestWriteTablePreservesConnectorJSONStrings(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "connector_json_strings"+TableFileExt)
	wantData := map[string]any{
		"date":           "2026-01-07",
		"timestamp":      "2026-01-07T03:04:05Z",
		"malformed_date": "2026-13-40",
		"plain":          "not-a-date",
		"number":         int64(7),
		"fraction":       1.25,
		"enabled":        true,
		"nullable":       nil,
		"items": []any{
			map[string]any{"when": "2026-01-08T04:05:06Z", "count": int64(2)},
		},
	}
	if err := WriteTable(ctx, path, []Row{{"id": "fixture", "data": wantData}}); err != nil {
		t.Fatalf("WriteTable() error = %v", err)
	}

	var got Row
	if err := ReadTable(ctx, path, func(row Row) error {
		got = row
		return nil
	}); err != nil {
		t.Fatalf("ReadTable() error = %v", err)
	}
	gotData, ok := got["data"].(map[string]any)
	if !ok {
		t.Fatalf("reopened data = %T(%v), want JSON object", got["data"], got["data"])
	}
	if !reflect.DeepEqual(gotData, wantData) {
		t.Fatalf("reopened data = %#v, want exact JSON shape %#v", gotData, wantData)
	}
}

func TestWriteTablePreservesAllEmptyObjectBatches(t *testing.T) {
	ctx := context.Background()
	for _, test := range []struct {
		name string
		rows []Row
	}{
		{
			name: "nested request objects",
			rows: []Row{
				{"id": "one", "data": map[string]any{}},
				{"id": "two", "data": map[string]any{}},
			},
		},
		{
			name: "top-level connector records",
			rows: []Row{{}, {}},
		},
		{
			name: "provider field matching physical sentinel",
			rows: []Row{{emptyObjectPhysicalColumn: true}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "empty_objects"+TableFileExt)
			if err := WriteTable(ctx, path, test.rows); err != nil {
				t.Fatalf("WriteTable() error = %v", err)
			}
			var got []Row
			if err := ReadTable(ctx, path, func(row Row) error {
				got = append(got, row)
				return nil
			}); err != nil {
				t.Fatalf("ReadTable() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.rows) {
				t.Fatalf("reopened rows = %#v, want exact JSON objects %#v", got, test.rows)
			}
		})
	}
}

func TestReadAllEmptyObjectTableStopsOnEmitterError(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "empty_objects"+TableFileExt)
	if err := WriteTable(ctx, path, []Row{{}, {}, {}}); err != nil {
		t.Fatalf("WriteTable() error = %v", err)
	}
	wantErr := errors.New("stop after first empty object")
	emitted := 0
	err := ReadTable(ctx, path, func(row Row) error {
		emitted++
		if len(row) != 0 {
			t.Fatalf("reopened row = %#v, want empty object", row)
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) || emitted != 1 {
		t.Fatalf("ReadTable() = error %v after %d rows, want emitter error after one row", err, emitted)
	}
}

// TestWriteTableKeepsAColumnThatFirstAppearsLate is the regression proof for a
// silent column drop. The table is built by handing rows to a JSON reader that
// infers the schema, and that inference samples a bounded prefix by default —
// 20,480 rows. A connector whose field is sparse, or whose payload grew a field
// partway through a backfill, would have that column inferred away and every
// one of its values written as if it had never existed: no error, no warning,
// and a table that reads back cleanly while missing data that is sitting in the
// write-ahead log beside it.
//
// The row count here is deliberately past that default rather than a round
// number, so the test fails if the bound is reintroduced at any size.
func TestWriteTableKeepsAColumnThatFirstAppearsLate(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "late_column"+TableFileExt)

	const rows = 25000
	const lateRow = 24999
	out := make([]Row, 0, rows)
	for i := 0; i < rows; i++ {
		row := Row{"id": fmt.Sprintf("r%05d", i)}
		if i == lateRow {
			row["late_field"] = "only here"
		}
		out = append(out, row)
	}
	if err := WriteTable(ctx, path, out); err != nil {
		t.Fatalf("WriteTable() error = %v", err)
	}

	var seen int
	var lateValue any
	var haveColumn bool
	if err := ReadTable(ctx, path, func(row Row) error {
		seen++
		if value, ok := row["late_field"]; ok {
			haveColumn = true
			if value != nil {
				lateValue = value
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("ReadTable() error = %v", err)
	}

	if seen != rows {
		t.Fatalf("read %d rows, want %d", seen, rows)
	}
	if !haveColumn {
		t.Fatalf("column %q was dropped from the materialized table; it first appears at row %d", "late_field", lateRow)
	}
	if lateValue != "only here" {
		t.Fatalf("late column value = %v, want %q", lateValue, "only here")
	}
}

// TestWriteTableKeepsAColumnWhoseTypeWidensLate covers the same inference bound
// from the other side: a column that looks like one type in the sampled prefix
// and turns out to hold another later. Losing the row, or the value, would be
// the same silent wrong answer as dropping the column.
func TestWriteTableKeepsAColumnWhoseTypeWidensLate(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "widening"+TableFileExt)

	const rows = 25000
	out := make([]Row, 0, rows)
	for i := 0; i < rows; i++ {
		row := Row{"id": fmt.Sprintf("r%05d", i), "value": i}
		if i == rows-1 {
			row["value"] = "not a number"
		}
		out = append(out, row)
	}
	if err := WriteTable(ctx, path, out); err != nil {
		t.Fatalf("WriteTable() error = %v", err)
	}

	var seen int
	var last any
	if err := ReadTable(ctx, path, func(row Row) error {
		seen++
		last = row["value"]
		return nil
	}); err != nil {
		t.Fatalf("ReadTable() error = %v", err)
	}
	if seen != rows {
		t.Fatalf("read %d rows, want %d", seen, rows)
	}
	if fmt.Sprint(last) != "not a number" {
		t.Fatalf("widened value = %v, want the string it was written as", last)
	}
}

// TestTablePathsSurviveShellAndSQLMetacharactersInTheRoot covers the one part of
// a table path pm does not generate: the warehouse root, which the operator
// supplies with `--config path=...` and which SafePathPart never sees. Every
// Windows path contains backslashes, and a root can legitimately contain a
// quote or a space on any platform. Those reach a SQL string literal, so this
// pins that they are carried through as data rather than interpreted.
func TestTablePathsSurviveShellAndSQLMetacharactersInTheRoot(t *testing.T) {
	ctx := context.Background()
	for _, dir := range []string{
		`back\slash`,     // every Windows path; also a legal macOS/Linux name
		`back\the\table`, // \t, the escape sequence most likely to be mangled
		`quo'te`,         // closes the SQL literal if not escaped
		`with space`,
	} {
		t.Run(dir, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), dir)
			path := filepath.Join(root, "records"+TableFileExt)
			want := Row{"id": "r1", "name": "Ada"}
			if err := WriteTable(ctx, path, []Row{want}); err != nil {
				t.Fatalf("WriteTable() error = %v", err)
			}
			got := make([]Row, 0, 1)
			if err := ReadTable(ctx, path, func(row Row) error {
				got = append(got, row)
				return nil
			}); err != nil {
				t.Fatalf("ReadTable() error = %v", err)
			}
			if len(got) != 1 || got[0]["id"] != "r1" || got[0]["name"] != "Ada" {
				t.Fatalf("round trip through root %q returned %v, want %v", dir, got, want)
			}
		})
	}
}
