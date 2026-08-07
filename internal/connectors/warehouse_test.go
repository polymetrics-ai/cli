package connectors

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/warehouse"
)

// TestWarehouseCatalogNamesEveryHolderOfATableName keeps the catalog agreeing
// with the error the very next read of that name raises. A name held by a
// connection and by an unattributed root-level file at once was described as
// if the connection owned it alone, while the read that followed correctly
// refused and named both.
func TestWarehouseCatalogNamesEveryHolderOfATableName(t *testing.T) {
	root := t.TempDir()
	cfg := RuntimeConfig{Config: map[string]string{"path": root}}

	location, err := warehouse.LocationFor(root, "ws_1", "hubspot", "conn_a", "acme")
	if err != nil {
		t.Fatal(err)
	}
	if err := location.EnsureOwnership(); err != nil {
		t.Fatal(err)
	}
	owned, err := location.TablePath("records")
	if err != nil {
		t.Fatal(err)
	}
	writeWarehouseTableFixture(t, owned, warehouse.Row{"id": "a1"})
	writeWarehouseTableFixture(t, filepath.Join(root, "records"+warehouse.TableFileExt), warehouse.Row{"id": "direct"})
	writeWarehouseTableFixture(t, filepath.Join(root, "seeded"+warehouse.TableFileExt), warehouse.Row{"id": "s1"})

	catalog, err := Warehouse{}.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog() error = %v", err)
	}
	described := make(map[string]string, len(catalog.Streams))
	for _, stream := range catalog.Streams {
		described[stream.Name] = stream.Description
	}
	shared, ok := described["records"]
	if !ok {
		t.Fatalf("Catalog() streams = %#v, want the shared table listed", catalog.Streams)
	}
	for _, want := range []string{"acme", warehouse.UnattributedConnection} {
		if !strings.Contains(shared, want) {
			t.Fatalf("catalog description %q does not name %q, which the next read of this table names", shared, want)
		}
	}
	// A table only the root holds is owned by no connection, so no ownership
	// is claimed for it and an unscoped read of it still resolves.
	if got := described["seeded"]; strings.Contains(got, "connection") {
		t.Fatalf("catalog description %q claims a connection owns an unattributed table", got)
	}
}

// TestWarehouseValidateWriteRefusesAnUnwritableTableName pins the refusal at
// the moment a caller can still act on it. A reverse plan is approved before
// it writes and cannot be edited afterwards, so a name rejected only by Write
// would leave an approved plan that can never run.
func TestWarehouseValidateWriteRefusesAnUnwritableTableName(t *testing.T) {
	ctx := context.Background()
	cfg := RuntimeConfig{Config: map[string]string{"path": t.TempDir()}}
	for _, table := range []string{"my table", "../escape", ".", "sub/table", ""} {
		err := Warehouse{}.ValidateWrite(ctx, WriteRequest{Stream: "", Table: table, Config: cfg}, nil)
		if err == nil {
			t.Fatalf("ValidateWrite(%q) error = nil, want rejection", table)
		}
		if !strings.Contains(err.Error(), "use only letters, digits") {
			t.Fatalf("rejection %q does not say which characters are acceptable", err)
		}
	}
	if err := (Warehouse{}).ValidateWrite(ctx, WriteRequest{Stream: "records", Config: cfg}, nil); err != nil {
		t.Fatalf("ValidateWrite(stream name) error = %v, want acceptance", err)
	}
	if err := (Warehouse{}).ValidateWrite(ctx, WriteRequest{Stream: "records", Table: "records_copy", Config: cfg}, nil); err != nil {
		t.Fatalf("ValidateWrite(usable table) error = %v, want acceptance", err)
	}
}

// TestWarehouseValidateWriteAgreesWithWrite is the drift guard. ValidateWrite
// exists only to predict Write, so a second copy of the table resolution or of
// the name rule would let the predictor diverge from the thing it predicts and
// send an approved reverse plan into a write that refuses it — the failure
// ValidateWrite was added to prevent.
func TestWarehouseValidateWriteAgreesWithWrite(t *testing.T) {
	ctx := context.Background()
	for _, req := range []WriteRequest{
		{Stream: "records"},
		{Stream: "records", Table: "records_copy"},
		{Stream: "records", Table: "my table"},
		{Stream: "my stream"},
		{Stream: "records", Table: "../escape"},
		{Stream: "records", Table: "."},
		{Stream: "", Table: ""},
		{Stream: "sub/stream", Table: ""},
		// A warehouse still holding a pre-Parquet table refuses every write.
		// The predictor has to know that too, or a reverse plan approved
		// against such a warehouse could never run.
		{Stream: "records", Config: legacyFormatWarehouseConfig(t)},
	} {
		if req.Config.Config == nil {
			req.Config = RuntimeConfig{Config: map[string]string{"path": t.TempDir()}}
		}
		validated := (Warehouse{}).ValidateWrite(ctx, req, nil)
		_, written := (Warehouse{}).Write(ctx, req, []Record{{"id": "a1"}})
		if (validated == nil) != (written == nil) {
			t.Fatalf("ValidateWrite(stream %q, table %q) = %v but Write = %v; the predictor and the write disagree", req.Stream, req.Table, validated, written)
		}
		if validated != nil && validated.Error() != written.Error() {
			t.Fatalf("ValidateWrite refusal %q does not match the Write refusal %q", validated, written)
		}
	}
}

// writeWarehouseTableFixture materializes a table the way a sync would, so a
// connector test exercises the real on-disk format rather than a hand-rolled
// stand-in that could drift from it.
func writeWarehouseTableFixture(t *testing.T, path string, rows ...warehouse.Row) {
	t.Helper()
	if err := warehouse.WriteTable(context.Background(), path, rows); err != nil {
		t.Fatalf("write table fixture %s: %v", path, err)
	}
}

// legacyFormatWarehouseConfig points at a warehouse still holding a table in
// the pre-Parquet format.
func legacyFormatWarehouseConfig(t *testing.T) RuntimeConfig {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "left_over"+warehouse.LegacyTableFileExt), []byte(`{"id":"legacy-1"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return RuntimeConfig{Config: map[string]string{"path": root}}
}
