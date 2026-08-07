package warehouse

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

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
