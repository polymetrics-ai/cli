package warehouse

import (
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
	shared := filepath.Join(root, TablesDirName, "records.jsonl")
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
	if err := os.WriteFile(filepath.Join(root, "records.jsonl"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
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
	// it is named plainly rather than rendered as an empty entry.
	if err := os.WriteFile(filepath.Join(root, "records.jsonl"), []byte(`{"id":"direct"}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = FindTable(root, "records", "")
	if !errors.As(err, &ambiguous) {
		t.Fatalf("FindTable(with root table) error = %T %v, want *AmbiguousTableError", err, err)
	}
	if !strings.Contains(ambiguous.Error(), unattributedConnectionLabel) {
		t.Fatalf("ambiguity error %q does not name the unattributed table", ambiguous.Error())
	}
	if _, err := FindTable(root, "missing", ""); err == nil {
		t.Fatal("FindTable(missing table) error = nil, want rejection")
	}
}

func TestTablesSkipsDirectoriesWithoutAnOwnershipRecord(t *testing.T) {
	root := t.TempDir()
	orphan := filepath.Join(root, "ws_1", "hubspot", "conn_orphan", TablesDirName)
	if err := os.MkdirAll(orphan, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(orphan, "records.jsonl"), []byte("{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tables, err := Tables(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 0 {
		t.Fatalf("Tables() = %#v, want nothing readable without an ownership record", tables)
	}
}
