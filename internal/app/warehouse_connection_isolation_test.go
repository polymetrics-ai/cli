package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/warehouse"
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
			if entry.Name() == warehouse.LegacyRawDirName || entry.Name() == warehouse.WALDirName {
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

// TestBothConnectionsKeepTheirOwnRowsAndAreReadableByName is the positive half
// of the isolation contract: each connection's table is intact, an unscoped
// read of the shared name is refused rather than answered from one tenant's
// file, and naming the connection resolves it.
func TestBothConnectionsKeepTheirOwnRowsAndAreReadableByName(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("multi_tenant_readback", nil)
	a, _ := setupTwoConnectionWarehouseApp(t, source, "incremental_append_deduped")

	source.records = []connectors.Record{
		{"id": "a1", "name": "Acme Ada", "updated_at": "2026-08-06T00:00:00Z"},
		{"id": "a2", "name": "Acme Alan", "updated_at": "2026-08-06T00:00:01Z"},
	}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: "acme", Stream: "records", BatchSize: 10}); err != nil {
		t.Fatal(err)
	}
	source.records = []connectors.Record{
		{"id": "g1", "name": "Globex Grace", "updated_at": "2026-08-06T00:00:02Z"},
	}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: "globex", Stream: "records", BatchSize: 10}); err != nil {
		t.Fatal(err)
	}

	_, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	var ambiguous *warehouse.AmbiguousTableError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("QueryTable(unscoped) error = %T %v, want *warehouse.AmbiguousTableError", err, err)
	}
	for _, want := range []string{"acme", "globex"} {
		if !strings.Contains(ambiguous.Error(), want) {
			t.Fatalf("ambiguity error %q does not name %q", ambiguous.Error(), want)
		}
	}

	for connection, wantIDs := range map[string][]string{
		"acme":   {"a1", "a2"},
		"globex": {"g1"},
	} {
		rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Connection: connection, Limit: 10})
		if err != nil {
			t.Fatalf("QueryTable(%s) error = %v", connection, err)
		}
		got := make([]string, 0, len(rows))
		for _, row := range rows {
			got = append(got, toComparableString(row["id"]))
		}
		sort.Strings(got)
		if !reflect.DeepEqual(got, wantIDs) {
			t.Fatalf("QueryTable(%s) ids = %v, want %v", connection, got, wantIDs)
		}
	}

	if _, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Connection: "nonexistent", Limit: 10}); err == nil {
		t.Fatal("QueryTable(unknown connection) error = nil, want rejection")
	}
}

// TestSyncRefusesWhenAnotherConnectionOwnsTheDirectory keeps the ownership
// assertion honest independently of directory nesting: if a connection ever
// resolves to a directory another connection owns, the run fails loudly rather
// than rebuilding that directory's tables from its own log.
func TestSyncRefusesWhenAnotherConnectionOwnsTheDirectory(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("ownership_guard", []connectors.Record{
		{"id": "a1", "name": "Acme Ada", "updated_at": "2026-08-06T00:00:00Z"},
	})
	a, warehouseDir := setupTwoConnectionWarehouseApp(t, source, "incremental_append_deduped")
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: "acme", Stream: "records", BatchSize: 10}); err != nil {
		t.Fatal(err)
	}

	// Rewrite the ownership record so the directory names a different
	// connection, which is the state a reintroduced shared path would produce.
	location, err := a.warehouseLocation(warehouseDir, mustFindConnection(t, a, "acme"))
	if err != nil {
		t.Fatal(err)
	}
	stolen := location.Owner
	stolen.Connection = "conn_someone_else"
	stolen.DisplayName = "globex"
	stolen.CreatedAt = time.Now().UTC()
	payload, err := json.Marshal(stolen)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(location.OwnerPath(), payload, 0o600); err != nil {
		t.Fatal(err)
	}

	before := readWarehouseTableRows(t, warehouseDir)
	_, err = a.RunETL(ctx, RunETLRequest{Connection: "acme", Stream: "records", BatchSize: 10})
	var ownership *warehouse.OwnershipError
	if !errors.As(err, &ownership) {
		t.Fatalf("RunETL() error = %T %v, want *warehouse.OwnershipError", err, err)
	}
	for _, want := range []string{"acme", "globex"} {
		if !strings.Contains(ownership.Error(), want) {
			t.Fatalf("ownership error %q does not name %q", ownership.Error(), want)
		}
	}
	if got := readWarehouseTableRows(t, warehouseDir); !reflect.DeepEqual(warehouseRowIDs(got), warehouseRowIDs(before)) {
		t.Fatalf("refused run still changed warehouse rows: %v -> %v", warehouseRowIDs(before), warehouseRowIDs(got))
	}
}

// TestSyncRefusesLegacyFlatWarehouse checks the one thing a legacy warehouse
// must do: stop, explain, and touch nothing. Which connection owns a flat
// table is unknowable, so pm neither guesses nor deletes.
func TestSyncRefusesLegacyFlatWarehouse(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("legacy_layout", []connectors.Record{
		{"id": "a1", "updated_at": "2026-08-06T00:00:00Z"},
	})
	a, warehouseDir := setupTwoConnectionWarehouseApp(t, source, "incremental_append_deduped")

	legacyRaw := filepath.Join(warehouseDir, warehouse.LegacyRawDirName)
	if err := os.MkdirAll(legacyRaw, 0o700); err != nil {
		t.Fatal(err)
	}
	legacyRawFile := filepath.Join(legacyRaw, "acme__records__records.jsonl")
	if err := os.WriteFile(legacyRawFile, []byte(`{"record":{"id":"old"}}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	legacyTable := filepath.Join(warehouseDir, "records.jsonl")
	if err := os.WriteFile(legacyTable, []byte(`{"id":"old"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := a.RunETL(ctx, RunETLRequest{Connection: "acme", Stream: "records", BatchSize: 10})
	var legacy *warehouse.LegacyLayoutError
	if !errors.As(err, &legacy) {
		t.Fatalf("RunETL() error = %T %v, want *warehouse.LegacyLayoutError", err, err)
	}
	if !strings.Contains(legacy.Error(), "Delete") || !strings.Contains(legacy.Error(), warehouseDir) {
		t.Fatalf("legacy error %q does not tell the operator what to do", legacy.Error())
	}
	for _, path := range []string{legacyRawFile, legacyTable} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("legacy file %s was removed by a refused run: %v", path, err)
		}
	}
}

// TestCreateConnectionRejectsAmbiguousNames pins the creation-time half of the
// fix. Coercing these names into a valid path is what let five distinct
// connections resolve to one warehouse file.
func TestCreateConnectionRejectsAmbiguousNames(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("name_validation", nil)
	a, _ := setupTwoConnectionWarehouseApp(t, source, "incremental_append_deduped")

	request := func(name string) CreateConnectionRequest {
		return CreateConnectionRequest{
			Name:        name,
			Source:      EndpointConfig{Connector: source.Name(), Credential: "acme-source"},
			Destination: EndpointConfig{Connector: "warehouse", Credential: "warehouse"},
			Streams: map[string]StreamConfig{
				"records": {SyncMode: "incremental_append_deduped", CursorField: "updated_at", PrimaryKey: []string{"id"}, DestinationTable: "records"},
			},
		}
	}
	for _, name := range []string{"acme.prod", "acme prod", "acme:prod", "acme/prod", "acme#prod", "..", " acme"} {
		if _, err := a.CreateConnection(ctx, request(name)); err == nil {
			t.Fatalf("CreateConnection(%q) error = nil, want rejection", name)
		}
	}
	// Case alone is too weak a distinction to separate two tenants' data.
	if _, err := a.CreateConnection(ctx, request("ACME")); err == nil {
		t.Fatal(`CreateConnection("ACME") error = nil, want rejection against existing "acme"`)
	}
	if _, err := a.CreateConnection(ctx, request("acme_staging")); err != nil {
		t.Fatalf("CreateConnection(acme_staging) error = %v, want acceptance", err)
	}
}

// TestConnectionIdentityIsOpaqueAndNotDerivedFromNameOrCredential keeps the
// path components free of anything a user typed or a credential carries.
func TestConnectionIdentityIsOpaqueAndNotDerivedFromNameOrCredential(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("identity_shape", []connectors.Record{
		{"id": "a1", "updated_at": "2026-08-06T00:00:00Z"},
	})
	a, warehouseDir := setupTwoConnectionWarehouseApp(t, source, "incremental_append_deduped")
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: "acme", Stream: "records", BatchSize: 10}); err != nil {
		t.Fatal(err)
	}
	conn := mustFindConnection(t, a, "acme")
	location, err := a.warehouseLocation(warehouseDir, conn)
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(warehouseDir, location.ConnectionDir)
	if err != nil {
		t.Fatal(err)
	}
	components := strings.Split(relative, string(filepath.Separator))
	if len(components) != 3 {
		t.Fatalf("connection directory %q has %d components, want workspace/connector/connection", relative, len(components))
	}
	if components[0] != a.state.WorkspaceID || components[2] != conn.ID {
		t.Fatalf("path components = %v, want workspace %q and connection %q", components, a.state.WorkspaceID, conn.ID)
	}
	for _, component := range components {
		if !warehouse.SafePathPart(component) {
			t.Fatalf("path component %q is not a safe path part", component)
		}
		if strings.Contains(component, "acme") || strings.Contains(component, "token") {
			t.Fatalf("path component %q leaks a display name or credential material", component)
		}
	}

	raw, err := os.ReadFile(location.OwnerPath())
	if err != nil {
		t.Fatalf("read ownership record: %v", err)
	}
	var owner warehouse.Owner
	if err := json.Unmarshal(raw, &owner); err != nil {
		t.Fatal(err)
	}
	if owner.DisplayName != "acme" || owner.Connection != conn.ID || owner.Workspace != a.state.WorkspaceID {
		t.Fatalf("ownership record = %#v, want acme/%s/%s", owner, conn.ID, a.state.WorkspaceID)
	}
	if strings.Contains(string(raw), "acme-token") {
		t.Fatal("ownership record contains credential material")
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
