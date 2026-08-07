package app

import (
	"bufio"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// warehouseTableRow is one materialized row plus the table file it came from.
type warehouseTableRow struct {
	File   string
	Record connectors.Record
}

// readWarehouseTableRows walks every materialized table under root and returns
// its rows. It deliberately walks rather than probing a known filename so the
// same assertion holds for the flat legacy layout and the per-connection
// layout that replaces it. Raw/WAL logs are excluded: they are the append-only
// source of truth, not the table a reader queries.
func readWarehouseTableRows(t *testing.T, root string) []warehouseTableRow {
	t.Helper()
	rows := make([]warehouseTableRow, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == legacyRawDirName || entry.Name() == warehouseWALDirName {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".jsonl") {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var record connectors.Record
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				return err
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			rows = append(rows, warehouseTableRow{File: rel, Record: record})
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatalf("walk warehouse %s: %v", root, err)
	}
	return rows
}

func warehouseRowIDs(rows []warehouseTableRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, toComparableString(row.Record["id"]))
	}
	sort.Strings(ids)
	return ids
}

// TestSecondConnectionDoesNotDestroyFirstConnectionRows is the regression proof
// for the silent cross-connection data loss: two connections with distinct
// credentials, both incremental_append_deduped, both materializing a table
// named "records". Before the fix the second sync rebuilt the shared final
// table from its own raw log and renamed it over the first connection's rows,
// with no error and exit status 0.
func TestSecondConnectionDoesNotDestroyFirstConnectionRows(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("multi_tenant_source", nil)
	a, warehouseDir := setupTwoConnectionWarehouseApp(t, source, "incremental_append_deduped")

	source.records = []connectors.Record{
		{"id": "a1", "name": "Acme Ada", "updated_at": "2026-08-06T00:00:00Z"},
		{"id": "a2", "name": "Acme Alan", "updated_at": "2026-08-06T00:00:01Z"},
	}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: "acme", Stream: "records", BatchSize: 10}); err != nil {
		t.Fatalf("acme RunETL() error = %v", err)
	}
	afterAcme := warehouseRowIDs(readWarehouseTableRows(t, warehouseDir))
	if len(afterAcme) != 2 {
		t.Fatalf("after acme sync rows = %v, want 2 acme rows", afterAcme)
	}

	source.records = []connectors.Record{
		{"id": "g1", "name": "Globex Grace", "updated_at": "2026-08-06T00:00:02Z"},
	}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: "globex", Stream: "records", BatchSize: 10}); err != nil {
		t.Fatalf("globex RunETL() error = %v", err)
	}

	rows := readWarehouseTableRows(t, warehouseDir)
	got := warehouseRowIDs(rows)
	for _, want := range []string{"a1", "a2", "g1"} {
		found := false
		for _, id := range got {
			if id == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("row %q is missing after the second connection synced; warehouse rows = %v", want, got)
		}
	}

	// One table file must never hold two connections' rows. Deduped modes lose
	// data outright; append modes interleave tenants into one table instead.
	// Both are the same defect, so both are asserted here.
	owner := map[string]string{"a1": "acme", "a2": "acme", "g1": "globex"}
	perFile := map[string]map[string]struct{}{}
	for _, row := range rows {
		id := toComparableString(row.Record["id"])
		connection, ok := owner[id]
		if !ok {
			t.Fatalf("unexpected row id %q in %s", id, row.File)
		}
		if perFile[row.File] == nil {
			perFile[row.File] = map[string]struct{}{}
		}
		perFile[row.File][connection] = struct{}{}
	}
	for file, connections := range perFile {
		if len(connections) > 1 {
			names := make([]string, 0, len(connections))
			for name := range connections {
				names = append(names, name)
			}
			sort.Strings(names)
			t.Fatalf("table %s holds rows from more than one connection: %v", file, names)
		}
	}
}

// setupTwoConnectionWarehouseApp builds one project with two connections that
// share a source connector and a destination table name but hold distinct
// credentials, which is the shape that reproduced the data loss.
func setupTwoConnectionWarehouseApp(t *testing.T, source *scriptedSyncSource, mode string) (*App, string) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatal(err)
	}
	a, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	registry := connectors.NewRegistry()
	registry.Register(source)
	a.registry = registry
	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "warehouse",
		Connector: "warehouse",
		Config:    map[string]string{"path": warehouseDir},
	}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"acme", "globex"} {
		if _, err := a.AddCredential(ctx, AddCredentialRequest{
			Name:      name + "-source",
			Connector: source.Name(),
			Secrets:   map[string]string{"token": name + "-token"},
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
			Name:        name,
			Source:      EndpointConfig{Connector: source.Name(), Credential: name + "-source"},
			Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
			Streams: map[string]StreamConfig{
				"records": {
					SyncMode:         mode,
					CursorField:      "updated_at",
					PrimaryKey:       []string{"id"},
					DestinationTable: "records",
				},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	return a, warehouseDir
}
