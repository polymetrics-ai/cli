package warehouse

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateConnectionNameRejectsNamesThatFoldTogether pins the exact
// collision sets measured against the removed lossy name folding, where `.`,
// `/`, ` ` and `:` all became `_` and every other character was dropped. Each
// group below collapsed to a single warehouse file, so every member after the
// first is data loss waiting for a second sync.
func TestValidateConnectionNameRejectsNamesThatFoldTogether(t *testing.T) {
	collisions := [][]string{
		{"acme.prod", "acme prod", "acme:prod", "acme/prod", "acme_prod"},
		{"acme#prod", "acme%prod", "acmeprod"},
	}
	for _, group := range collisions {
		accepted := make([]string, 0, len(group))
		for _, name := range group {
			if err := ValidateConnectionName(name); err == nil {
				accepted = append(accepted, name)
			}
		}
		if len(accepted) > 1 {
			t.Fatalf("names %v all remain valid; they folded to one warehouse path: %v", group, accepted)
		}
	}
}

func TestValidateConnectionName(t *testing.T) {
	valid := []string{"acme", "acme_prod", "acme-prod", "a", "A1", "0acme", strings.Repeat("a", MaxConnectionNameLength)}
	for _, name := range valid {
		if err := ValidateConnectionName(name); err != nil {
			t.Fatalf("ValidateConnectionName(%q) error = %v, want nil", name, err)
		}
	}
	invalid := []string{
		"", "   ", " acme", "acme ", "acme.prod", "acme prod", "acme:prod", "acme/prod",
		"acme#prod", "acme%prod", "..", "../escape", "_acme", "-acme", "acmé",
		strings.Repeat("a", MaxConnectionNameLength+1),
	}
	for _, name := range invalid {
		if err := ValidateConnectionName(name); err == nil {
			t.Fatalf("ValidateConnectionName(%q) error = nil, want rejection", name)
		}
	}
}

func TestSafePathPartRejectsRatherThanRewrites(t *testing.T) {
	// "." is rejected on its own: filepath.Join collapses it, so it would
	// resolve to its parent rather than to a directory of its own.
	for _, value := range []string{"", ".", "..", "a/b", "a b", "a:b", "a#b", "..hidden/../x", strings.Repeat("a", 257)} {
		if SafePathPart(value) {
			t.Fatalf("SafePathPart(%q) = true, want rejection", value)
		}
	}
	for _, value := range []string{"records", "my.table", "my-table", "my_table", "conn_ab12", "ws_00ff"} {
		if !SafePathPart(value) {
			t.Fatalf("SafePathPart(%q) = false, want acceptance", value)
		}
	}
}

func TestOwnerIdentityUsesTheSharedArtifactTriple(t *testing.T) {
	owner := Owner{
		Workspace:   "ws_1",
		Connector:   "postgres",
		Connection:  "conn_1",
		DisplayName: "before-rename",
	}
	want := ArtifactIdentity{
		WorkspaceID:  "ws_1",
		ConnectorID:  "postgres",
		ConnectionID: "conn_1",
	}
	if got := owner.Identity(); got != want {
		t.Fatalf("Owner.Identity() = %#v, want %#v", got, want)
	}

	renamed := owner
	renamed.DisplayName = "after-rename"
	if !owner.SameIdentity(renamed) {
		t.Fatal("Owner.SameIdentity() treated a display-name change as an ownership change")
	}
	differentConnection := owner
	differentConnection.Connection = "conn_2"
	if owner.SameIdentity(differentConnection) {
		t.Fatal("Owner.SameIdentity() accepted a different connection ID")
	}
	differentConnector := owner
	differentConnector.Connector = "mysql"
	if owner.SameIdentity(differentConnector) {
		t.Fatal("Owner.SameIdentity() accepted a different connector ID")
	}
}

func TestLocationForRejectsUnsafeIdentityComponents(t *testing.T) {
	root := t.TempDir()
	cases := []struct {
		name         string
		workspace    string
		connector    string
		connectionID string
	}{
		{"workspace escape", "../evil", "hubspot", "conn_1"},
		{"connector escape", "ws_1", "..", "conn_1"},
		{"connection escape", "ws_1", "hubspot", "../other"},
		{"connection separator", "ws_1", "hubspot", "a/b"},
		{"empty connection", "ws_1", "hubspot", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := LocationFor(root, tc.workspace, tc.connector, tc.connectionID, "display"); err == nil {
				t.Fatal("LocationFor() error = nil, want rejection")
			}
		})
	}
}

func TestEnsureOwnershipRefusesAnotherConnectionsDirectory(t *testing.T) {
	root := t.TempDir()
	first, err := LocationFor(root, "ws_1", "hubspot", "conn_first", "acme")
	if err != nil {
		t.Fatal(err)
	}
	if err := first.EnsureOwnership(); err != nil {
		t.Fatalf("first EnsureOwnership() error = %v", err)
	}
	// Re-establishing ownership for the same connection is a no-op.
	if err := first.EnsureOwnership(); err != nil {
		t.Fatalf("repeat EnsureOwnership() error = %v", err)
	}

	// The same connection under a new display name keeps its directory and
	// updates the record, so a read scoped by connection name still resolves.
	renamed := first
	renamed.Owner.DisplayName = "acme-renamed"
	if err := renamed.EnsureOwnership(); err != nil {
		t.Fatalf("renamed EnsureOwnership() error = %v", err)
	}
	stored, err := readOwner(first.ConnectionDir)
	if err != nil {
		t.Fatal(err)
	}
	if stored.DisplayName != "acme-renamed" || stored.Connection != "conn_first" {
		t.Fatalf("ownership record after rename = %#v", stored)
	}

	// A second connection landing on the same directory is the regression this
	// guard exists for: it must fail loudly rather than take the directory over.
	hijack := first
	hijack.Owner.Connection = "conn_second"
	hijack.Owner.DisplayName = "globex"
	err = hijack.EnsureOwnership()
	var ownership *OwnershipError
	if !errors.As(err, &ownership) {
		t.Fatalf("EnsureOwnership() error = %T %v, want *OwnershipError", err, err)
	}
	for _, want := range []string{"acme", "globex", "conn_first", "conn_second"} {
		if !strings.Contains(ownership.Error(), want) {
			t.Fatalf("error %q does not name %q", ownership.Error(), want)
		}
	}
}

func TestEnsureOwnershipRefusesUnreadableRecord(t *testing.T) {
	root := t.TempDir()
	location, err := LocationFor(root, "ws_1", "hubspot", "conn_first", "acme")
	if err != nil {
		t.Fatal(err)
	}
	if err := location.EnsureOwnership(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(location.OwnerPath(), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	var ownership *OwnershipError
	if err := location.EnsureOwnership(); !errors.As(err, &ownership) {
		t.Fatalf("EnsureOwnership() error = %T %v, want *OwnershipError", err, err)
	}
}

// TestAssertOwnedTableCatchesAReintroducedSharedPath is the defence-in-depth
// check. Directory nesting already makes a shared table path unrepresentable;
// this proves that if a future change ever produced one anyway, the write is
// refused instead of silently overwriting another connection's rows.
func TestAssertOwnedTableCatchesAReintroducedSharedPath(t *testing.T) {
	root := t.TempDir()
	location, err := LocationFor(root, "ws_1", "hubspot", "conn_first", "acme")
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
	if err := location.AssertOwnedTable(owned, "records"); err != nil {
		t.Fatalf("AssertOwnedTable(own table) error = %v", err)
	}

	// The shape a regression would take: a table path back at the shared root,
	// outside any connection's directory.
	shared := filepath.Join(root, TablesDirName, "records"+TableFileExt)
	err = location.AssertOwnedTable(shared, "records")
	var ownership *OwnershipError
	if !errors.As(err, &ownership) {
		t.Fatalf("AssertOwnedTable(shared path) error = %T %v, want *OwnershipError", err, err)
	}
	if !ownership.MissingOwner {
		t.Fatalf("ownership error = %#v, want a missing-owner refusal", ownership)
	}
}

func TestCheckLegacyLayoutRefusesTheRemovedFlatLayout(t *testing.T) {
	root := t.TempDir()
	if err := CheckLegacyLayout(root); err != nil {
		t.Fatalf("CheckLegacyLayout(empty) error = %v, want nil", err)
	}
	// A root-level table on its own is an unattributed direct write, not the
	// shared-table layout, so it is left alone.
	writeTableFixture(t, filepath.Join(root, "records"+TableFileExt), Row{"id": "direct"})
	if err := CheckLegacyLayout(root); err != nil {
		t.Fatalf("CheckLegacyLayout(root table) error = %v, want nil", err)
	}

	if err := os.MkdirAll(filepath.Join(root, LegacyRawDirName), 0o700); err != nil {
		t.Fatal(err)
	}
	err := CheckLegacyLayout(root)
	var legacy *LegacyLayoutError
	if !errors.As(err, &legacy) {
		t.Fatalf("CheckLegacyLayout() error = %T %v, want *LegacyLayoutError", err, err)
	}
	message := legacy.Error()
	for _, want := range []string{root, LegacyRawDirName, "Delete", "will not rewrite or delete"} {
		if !strings.Contains(message, want) {
			t.Fatalf("legacy error %q does not mention %q", message, want)
		}
	}
}

func TestFindTableReportsAmbiguityInsteadOfPickingAWinner(t *testing.T) {
	root := t.TempDir()
	for _, tenant := range []struct{ id, name string }{{"conn_a", "acme"}, {"conn_b", "globex"}} {
		location, err := LocationFor(root, "ws_1", "hubspot", tenant.id, tenant.name)
		if err != nil {
			t.Fatal(err)
		}
		if err := location.EnsureOwnership(); err != nil {
			t.Fatal(err)
		}
		path, err := location.TablePath("records")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(`{"id":"`+tenant.id+`"}`+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	_, err := FindTable(root, "records", "")
	var ambiguous *AmbiguousTableError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("FindTable(unscoped) error = %T %v, want *AmbiguousTableError", err, err)
	}
	if len(ambiguous.Connections) != 2 {
		t.Fatalf("ambiguous connections = %v, want both tenants", ambiguous.Connections)
	}

	for _, tenant := range []struct{ name, id string }{{"acme", "conn_a"}, {"globex", "conn_b"}} {
		table, err := FindTable(root, "records", tenant.name)
		if err != nil {
			t.Fatalf("FindTable(%q) error = %v", tenant.name, err)
		}
		if table.Owner.Connection != tenant.id {
			t.Fatalf("FindTable(%q) owner = %q, want %q", tenant.name, table.Owner.Connection, tenant.id)
		}
	}
	if _, err := FindTable(root, "records", "unknown"); err == nil {
		t.Fatal("FindTable(unknown connection) error = nil, want rejection")
	}

	// A root-level table shares the namespace but has no owning connection, so
	// it is named by the selector that reaches it rather than rendered as an
	// empty entry — an entry no value could select would be a false promise.
	writeTableFixture(t, filepath.Join(root, "records"+TableFileExt), Row{"id": "direct"})
	_, err = FindTable(root, "records", "")
	if !errors.As(err, &ambiguous) {
		t.Fatalf("FindTable(with root table) error = %T %v, want *AmbiguousTableError", err, err)
	}
	if !strings.Contains(ambiguous.Error(), UnattributedConnection) {
		t.Fatalf("ambiguity error %q does not name the unattributed table", ambiguous.Error())
	}
	unattributed, err := FindTable(root, "records", UnattributedConnection)
	if err != nil {
		t.Fatalf("FindTable(%q) error = %v, want the root-level table", UnattributedConnection, err)
	}
	if unattributed.Path != filepath.Join(root, "records"+TableFileExt) {
		t.Fatalf("FindTable(%q) path = %q, want the root-level table", UnattributedConnection, unattributed.Path)
	}
	if err := ValidateConnectionName(UnattributedConnection); err == nil {
		t.Fatalf("ValidateConnectionName(%q) error = nil, want the selector reserved against real connections", UnattributedConnection)
	}
	if _, err := FindTable(root, "missing", ""); err == nil {
		t.Fatal("FindTable(missing table) error = nil, want rejection")
	}
	if _, err := FindTable(root, "missing", UnattributedConnection); err == nil {
		t.Fatalf("FindTable(missing, %q) error = nil, want rejection", UnattributedConnection)
	}
}

// A missing ownership record and an unreadable one are different facts. The
// write path already refuses both loudly; a read that answered "no such table"
// while the rows sat intact on disk would send the operator to re-sync data
// that already exists.
func TestTablesReportsAnUnreadableOwnershipRecord(t *testing.T) {
	for _, tc := range []struct {
		name    string
		record  string
		wantErr string
	}{
		{name: "corrupt", record: "{not json", wantErr: "decode warehouse ownership record"},
		{name: "unsupported version", record: `{"version":2}`, wantErr: "version 2 is unsupported"},
		{name: "no identity", record: `{"version":1}`, wantErr: "has no workspace"},
		{
			name:    "no display name",
			record:  `{"version":1,"workspace":"ws_1","connector":"hubspot","connection":"conn_a"}`,
			wantErr: "has no display_name",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			connectionDir := filepath.Join(root, "ws_1", "hubspot", "conn_a")
			tables := filepath.Join(connectionDir, TablesDirName)
			if err := os.MkdirAll(tables, 0o700); err != nil {
				t.Fatal(err)
			}
			writeTableFixture(t, filepath.Join(tables, "records"+TableFileExt), Row{"id": "r1"})
			if err := os.WriteFile(filepath.Join(connectionDir, OwnerFileName), []byte(tc.record), 0o600); err != nil {
				t.Fatal(err)
			}
			_, faults, err := Tables(root)
			if err != nil {
				t.Fatalf("Tables() error = %v, want the fault reported beside the healthy tables", err)
			}
			if len(faults) != 1 {
				t.Fatalf("Tables() faults = %#v, want the unreadable ownership record reported", faults)
			}
			for _, want := range []string{connectionDir, OwnerFileName, tc.wantErr} {
				if !strings.Contains(faults[0].Error(), want) {
					t.Fatalf("fault %q does not mention %q", faults[0], want)
				}
			}
			// The fact still reaches the operator: a lookup that cannot be
			// answered surfaces it rather than claiming the table is absent.
			_, err = FindTable(root, "records", "")
			var faulted *FaultError
			if !errors.As(err, &faulted) {
				t.Fatalf("FindTable() error = %T %v, want *FaultError", err, err)
			}
			for _, want := range []string{OwnerFileName, "Restore or remove"} {
				if !strings.Contains(faulted.Error(), want) {
					t.Fatalf("FindTable() error %q does not mention %q", faulted, want)
				}
			}
			if strings.Contains(faulted.Error(), "run a sync to materialize it") {
				t.Fatalf("FindTable() reported the table as absent despite a damaged record: %q", faulted)
			}
		})
	}
}

// TestOneDamagedOwnershipRecordDoesNotDenyHealthyConnections is the isolation
// contract. Nesting per connection exists so that one connection's problem
// stays that connection's problem; a guard whose blast radius is the whole
// project has the wrong shape, however loudly it fails.
func TestOneDamagedOwnershipRecordDoesNotDenyHealthyConnections(t *testing.T) {
	root := t.TempDir()
	healthy, err := LocationFor(root, "ws_1", "hubspot", "conn_healthy", "acme")
	if err != nil {
		t.Fatal(err)
	}
	if err := healthy.EnsureOwnership(); err != nil {
		t.Fatal(err)
	}
	healthyTable, err := healthy.TablePath("acme_records")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(healthyTable), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(healthyTable, []byte(`{"id":"a1"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// A second connection's record is damaged after its table was written.
	damaged, err := LocationFor(root, "ws_1", "hubspot", "conn_damaged", "globex")
	if err != nil {
		t.Fatal(err)
	}
	if err := damaged.EnsureOwnership(); err != nil {
		t.Fatal(err)
	}
	damagedTable, err := damaged.TablePath("globex_records")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(damagedTable), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(damagedTable, []byte(`{"id":"g1"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(damaged.OwnerPath(), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	tables, faults, err := Tables(root)
	if err != nil {
		t.Fatalf("Tables() error = %v, want the healthy connection still listed", err)
	}
	if len(tables) != 1 || tables[0].Connection != "acme" || tables[0].Name != "acme_records" {
		t.Fatalf("Tables() = %#v, want only the healthy connection's table", tables)
	}
	if len(faults) != 1 || faults[0].Dir != damaged.ConnectionDir {
		t.Fatalf("Tables() faults = %#v, want the damaged directory", faults)
	}

	// The healthy connection reads normally: it does not depend on the damaged
	// directory, so the damage costs it nothing.
	found, err := FindTable(root, "acme_records", "acme")
	if err != nil {
		t.Fatalf("FindTable(healthy) error = %v, want the read to succeed", err)
	}
	if found.Path != healthyTable {
		t.Fatalf("FindTable(healthy) path = %q, want %q", found.Path, healthyTable)
	}
	if _, err := FindTable(root, "acme_records", ""); err != nil {
		t.Fatalf("FindTable(healthy, unscoped) error = %v, want the read to succeed", err)
	}

	// The damaged connection's own table surfaces the fault instead.
	var faulted *FaultError
	if _, err := FindTable(root, "globex_records", ""); !errors.As(err, &faulted) {
		t.Fatalf("FindTable(damaged) error = %T %v, want *FaultError", err, err)
	}
	if !strings.Contains(faulted.Error(), damaged.OwnerPath()) {
		t.Fatalf("fault error %q does not name the damaged record", faulted)
	}
	if !strings.Contains(faulted.Error(), "healthy records are unaffected") {
		t.Fatalf("fault error %q does not say the blast radius is contained", faulted)
	}
}

// TestOwnershipRecordWithoutADisplayNameCannotPassAsUnattributed closes the
// record side of the reservation ValidateConnectionName closes on the name
// side. The empty connection is the sentinel for a root-level table owned by
// nobody, so a record that cannot say which connection it belongs to must be a
// fault rather than a value that answers to the unattributed selector.
func TestOwnershipRecordWithoutADisplayNameCannotPassAsUnattributed(t *testing.T) {
	root := t.TempDir()
	connectionDir := filepath.Join(root, "ws_1", "hubspot", "conn_a")
	tables := filepath.Join(connectionDir, TablesDirName)
	if err := os.MkdirAll(tables, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTableFixture(t, filepath.Join(tables, "records"+TableFileExt), Row{"id": "a1"})
	record := `{"version":1,"workspace":"ws_1","connector":"hubspot","connection":"conn_a"}`
	if err := os.WriteFile(filepath.Join(connectionDir, OwnerFileName), []byte(record), 0o600); err != nil {
		t.Fatal(err)
	}

	listed, faults, err := Tables(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 0 {
		t.Fatalf("Tables() = %#v, want a connection-owned table not listed as anyone's", listed)
	}
	if len(faults) != 1 || !faults[0].Holds("records") {
		t.Fatalf("Tables() faults = %#v, want the record reported as damaged", faults)
	}

	var faulted *FaultError
	if _, err := FindTable(root, "records", UnattributedConnection); !errors.As(err, &faulted) {
		t.Fatalf("FindTable(records, %q) error = %T %v, want the damaged record reported rather than answered", UnattributedConnection, err, err)
	}

	// A genuinely unattributed table of the same name is still selectable, and
	// it does not inherit the damaged directory's rows.
	writeTableFixture(t, filepath.Join(root, "records"+TableFileExt), Row{"id": "direct"})
	found, err := FindTable(root, "records", UnattributedConnection)
	if err != nil {
		t.Fatalf("FindTable(records, %q) error = %v, want the root-level table", UnattributedConnection, err)
	}
	if found.Path != filepath.Join(root, "records"+TableFileExt) {
		t.Fatalf("FindTable(records, %q) path = %q, want the root-level table", UnattributedConnection, found.Path)
	}
}

// TestDamagedRecordCannotDecideWhichConnectionAnUnscopedReadReturns is the
// other half of the isolation contract. Making a fault non-fatal so healthy
// connections keep reading is right, but the dropped directory is still a
// candidate: if it holds the same table name, an unscoped read that answered
// from the single healthy match would let a corrupted ownership file decide
// which tenant's rows the operator receives. That is a confident wrong answer
// to an ambiguous question, which is the defect this whole layout removes.
func TestDamagedRecordCannotDecideWhichConnectionAnUnscopedReadReturns(t *testing.T) {
	root := t.TempDir()
	write := func(location Location, table, id string) string {
		t.Helper()
		path, err := location.TablePath(table)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(`{"id":"`+id+`"}`+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}

	healthy, err := LocationFor(root, "ws_1", "hubspot", "conn_healthy", "acme")
	if err != nil {
		t.Fatal(err)
	}
	if err := healthy.EnsureOwnership(); err != nil {
		t.Fatal(err)
	}
	healthyTable := write(healthy, "records", "a1")
	write(healthy, "acme_only", "a2")

	damaged, err := LocationFor(root, "ws_1", "hubspot", "conn_damaged", "globex")
	if err != nil {
		t.Fatal(err)
	}
	if err := damaged.EnsureOwnership(); err != nil {
		t.Fatal(err)
	}
	write(damaged, "records", "g1")
	write(damaged, "globex_only", "g2")
	if err := os.WriteFile(damaged.OwnerPath(), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	// The damaged directory reports the table names it holds, because a name a
	// lookup cannot see is exactly what makes an answer unsafe to give.
	_, faults, err := Tables(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(faults) != 1 {
		t.Fatalf("Tables() faults = %#v, want the damaged directory", faults)
	}
	if !faults[0].Holds("records") || !faults[0].Holds("globex_only") {
		t.Fatalf("fault tables = %v, want both tables the damaged directory holds", faults[0].Tables)
	}
	if faults[0].Holds("acme_only") {
		t.Fatalf("fault tables = %v, want only the damaged directory's own tables", faults[0].Tables)
	}

	resolver, err := NewTableResolver(root)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := resolver.Tables()
	if len(snapshot) == 0 {
		t.Fatal("resolver snapshot has no healthy tables")
	}
	snapshot[0].Name = "changed"

	_, err = resolver.Find("records", "")
	var faulted *FaultError
	if !errors.As(err, &faulted) {
		t.Fatalf("TableResolver.Find(records, unscoped) error = %T %v, want the read refused rather than answered from one tenant", err, err)
	}
	if !faulted.Undecided {
		t.Fatalf("FindTable(records, unscoped) error = %q, want it reported as undecided rather than absent", faulted)
	}
	for _, want := range []string{"records", damaged.OwnerPath(), "Name the connection", "Restore or remove"} {
		if !strings.Contains(faulted.Error(), want) {
			t.Fatalf("undecided fault error %q does not mention %q", faulted, want)
		}
	}
	if strings.Contains(faulted.Error(), "--") {
		t.Fatalf("undecided fault error %q names a flag no caller promised", faulted)
	}

	// Isolation is intact: naming the connection answers, and a table the
	// damaged directory does not hold is answered unscoped as before.
	found, err := resolver.Find("records", "acme")
	if err != nil {
		t.Fatalf("TableResolver.Find(records, acme) error = %v, want the scoped read to succeed", err)
	}
	if found.Path != healthyTable {
		t.Fatalf("FindTable(records, acme) path = %q, want %q", found.Path, healthyTable)
	}
	if _, err := resolver.Find("acme_only", ""); err != nil {
		t.Fatalf("TableResolver.Find(acme_only, unscoped) error = %v, want a read that does not depend on the damaged directory to succeed", err)
	}

	// A root-level file the damaged directory cannot hold stays selectable.
	writeTableFixture(t, filepath.Join(root, "direct"+TableFileExt), Row{"id": "d1"})
	if _, err := resolver.Find("direct", UnattributedConnection); err == nil {
		t.Fatal("TableResolver.Find(direct) error = nil, want the pre-existing snapshot unchanged")
	}
	if _, err := FindTable(root, "direct", UnattributedConnection); err != nil {
		t.Fatalf("FindTable(direct, %q) error = %v, want the unattributed read to succeed", UnattributedConnection, err)
	}
}

// internal/warehouse cannot know which command is running, so it must not name
// a flag the caller may not accept. The surface that raised the error supplies
// the remedy it can actually honour.
func TestAmbiguityRemedyComesFromTheCallingSurface(t *testing.T) {
	err := error(&AmbiguousTableError{Table: "records", Connections: []string{"acme", ""}})
	if strings.Contains(err.Error(), "--connection") {
		t.Fatalf("default ambiguity error %q names a flag no caller promised", err)
	}
	if !strings.Contains(err.Error(), UnattributedConnection) {
		t.Fatalf("default ambiguity error %q does not name the unattributed table", err)
	}
	if got := WithAmbiguityRemedy(err, "re-create the plan"); !strings.Contains(got.Error(), "re-create the plan") {
		t.Fatalf("WithAmbiguityRemedy() error = %q, want the caller's remedy", got)
	}
	other := errors.New("unrelated")
	if got := WithAmbiguityRemedy(other, "re-create the plan"); got != other {
		t.Fatalf("WithAmbiguityRemedy(unrelated) = %v, want it returned untouched", got)
	}
}

func TestTablesSkipsDirectoriesWithoutAnOwnershipRecord(t *testing.T) {
	root := t.TempDir()
	orphan := filepath.Join(root, "ws_1", "hubspot", "conn_orphan", TablesDirName)
	if err := os.MkdirAll(orphan, 0o700); err != nil {
		t.Fatal(err)
	}
	writeTableFixture(t, filepath.Join(orphan, "records"+TableFileExt), Row{"id": "r1"})
	tables, faults, err := Tables(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 0 {
		t.Fatalf("Tables() = %#v, want nothing readable without an ownership record", tables)
	}
	// A wholly absent record is a skip, not a fault: an unowned directory is
	// not damaged, it is simply not anyone's to read.
	if len(faults) != 0 {
		t.Fatalf("Tables() faults = %#v, want a missing record treated as a skip", faults)
	}
}

// writeTableFixture materializes a table the way a sync would, so a layout test
// exercises the real on-disk format rather than a hand-rolled stand-in that
// could drift from it.
func writeTableFixture(t *testing.T, path string, rows ...Row) {
	t.Helper()
	if err := WriteTable(context.Background(), path, rows); err != nil {
		t.Fatalf("write table fixture %s: %v", path, err)
	}
}
