// Package warehouse defines the on-disk layout of the local JSONL warehouse.
//
// Every materialization is nested under the identity that produced it:
//
//	<root>/<workspace-id>/<connector>/<connection-id>/
//	    owner.json           identity of the connection owning this directory
//	    wal/<stream>.jsonl   append-only log; the source of truth
//	    tables/<table>.jsonl derived materialization, rewritten wholesale
//
// Two connections cannot produce the same table path because they never share
// a parent directory, so collision is unrepresentable rather than merely
// unlikely. There is no naming function left to get wrong.
//
// Every path component is either an opaque generated identifier or a value
// SafePathPart accepts verbatim. Nothing is ever rewritten to fit a path:
// rewriting is what let five distinct connection names collapse onto one file.
package warehouse

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	// OwnerFileName is the ownership assertion at a connection directory root.
	OwnerFileName = "owner.json"
	// OwnerVersion is the current ownership document version.
	OwnerVersion = 1
	// WALDirName holds a connection's append-only stream logs.
	WALDirName = "wal"
	// TablesDirName holds a connection's derived table materializations.
	TablesDirName = "tables"
	// LegacyRawDirName is the removed flat layout's shared raw log directory.
	// It is only ever read, as evidence that a warehouse predates this layout.
	LegacyRawDirName = "_pm_raw"
	// MaxConnectionNameLength bounds a connection name so it stays usable in
	// errors, state keys, and CLI output.
	MaxConnectionNameLength = 128
	// UnattributedConnection selects the root-level tables that no connection
	// owns: those written straight through the connector Write surface, or
	// seeded by hand. It is never given a real connection identity, but it does
	// need a selector, because the empty string already means "any connection"
	// and so can never name it. ValidateConnectionName rejects a leading '_',
	// so no real connection can ever answer to this value.
	UnattributedConnection = "_unattributed"
)

// SafePathPart reports whether value is safe to use verbatim as a single path
// component. It rejects `.`, `..` and any character outside [A-Za-z0-9._-]
// rather than silently rewriting them, which is the property the whole layout
// rests on. `.` is rejected on its own because filepath.Join collapses it, so
// it would silently resolve to its parent instead of to a directory of its own.
// It is the single implementation shared by warehouse paths and the catalog
// storage introduced in #3892; do not restate this rule elsewhere.
func SafePathPart(value string) bool {
	if value == "" || value == "." || len(value) > 256 || strings.Contains(value, "..") {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

// ValidateConnectionName rejects connection names that cannot be told apart
// once punctuation is folded. `acme.prod`, `acme prod`, `acme:prod`,
// `acme/prod` and `acme_prod` were all accepted before and all resolved to the
// same warehouse file; `:` additionally collides with the connection:stream
// sync-state key separator. Names are rejected at creation rather than coerced
// into something valid, because coercion is what made two distinct connections
// indistinguishable in the first place.
func ValidateConnectionName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("connection name is required")
	}
	if name != strings.TrimSpace(name) {
		return fmt.Errorf("connection name %q must not begin or end with whitespace", name)
	}
	if len(name) > MaxConnectionNameLength {
		return fmt.Errorf("connection name is %d characters; the maximum is %d", len(name), MaxConnectionNameLength)
	}
	for index, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			continue
		case (r == '-' || r == '_') && index > 0:
			continue
		case index == 0:
			return fmt.Errorf("connection name %q must start with a letter or digit", name)
		}
		return fmt.Errorf("connection name %q contains %q: use only letters, digits, '-' and '_' so two connections can never resolve to the same warehouse table", name, string(r))
	}
	return nil
}

// PathComponent returns value when SafePathPart accepts it, and an explanatory
// error when it does not.
func PathComponent(kind, value string) (string, error) {
	if !SafePathPart(value) {
		return "", fmt.Errorf("warehouse %s %q cannot be used as a path component: use only letters, digits, '.', '-' and '_'", kind, value)
	}
	return value, nil
}

// Owner is the ownership assertion stored at a connection directory's root.
// It records the opaque identity triple that determines the directory path,
// plus the human-readable connection name — kept here precisely because
// lossiness is harmless in a value that never becomes a path.
type Owner struct {
	Version     int       `json:"version"`
	Workspace   string    `json:"workspace"`
	Connector   string    `json:"connector"`
	Connection  string    `json:"connection"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// SameIdentity reports whether two owner records describe the same connection.
// DisplayName is deliberately excluded: renaming a connection does not change
// which data is its own.
func (o Owner) SameIdentity(other Owner) bool {
	return o.Workspace == other.Workspace &&
		o.Connector == other.Connector &&
		o.Connection == other.Connection
}

// OwnershipError reports that a warehouse directory is owned by a different
// connection than the one attempting to write it. It is always fatal:
// overwriting is the silent data loss this layout exists to prevent.
type OwnershipError struct {
	Dir            string
	Table          string
	OwnerName      string
	OwnerID        string
	WriterName     string
	WriterID       string
	MissingOwner   bool
	UnreadableFile string
}

func (e *OwnershipError) Error() string {
	if e == nil {
		return "warehouse ownership is unavailable"
	}
	subject := "warehouse directory " + e.Dir
	if e.Table != "" {
		subject = fmt.Sprintf("warehouse table %q in %s", e.Table, e.Dir)
	}
	switch {
	case e.UnreadableFile != "":
		return fmt.Sprintf("%s has an unreadable ownership record at %s; refusing to write it on behalf of connection %q", subject, e.UnreadableFile, e.WriterName)
	case e.MissingOwner:
		return fmt.Sprintf("%s has no ownership record; refusing to write it on behalf of connection %q", subject, e.WriterName)
	default:
		return fmt.Sprintf("%s is owned by connection %q (%s); connection %q (%s) must not write it", subject, e.OwnerName, e.OwnerID, e.WriterName, e.WriterID)
	}
}

// LegacyLayoutError reports a warehouse written by the removed flat layout, in
// which every connection shared one table file. Those tables carry no
// ownership record, so which connection owns a given table is unknowable — and
// in deduped modes the honest answer may be "the last one to sync, and the
// others' rows are already gone". pm refuses and tells the operator what to do
// rather than guessing an owner or deleting anything itself.
type LegacyLayoutError struct {
	Dir      string
	Evidence string
}

func (e *LegacyLayoutError) Error() string {
	if e == nil {
		return "warehouse layout is unavailable"
	}
	return fmt.Sprintf(
		"warehouse %s uses the removed flat layout (found %s), in which every connection shared one table file. "+
			"Tables are now stored per connection under <workspace-id>/<connector>/<connection-id>/%s/. "+
			"Delete %s and re-run your syncs to rebuild it; pm will not rewrite or delete warehouse data for you",
		e.Dir, e.Evidence, TablesDirName, e.Dir,
	)
}

// AmbiguousTableError reports that more than one connection materializes a
// table of the same name, so a read must say which one it means.
type AmbiguousTableError struct {
	Table       string
	Connections []string
	// Remedy states what the operator can do on the surface that raised this
	// error. It is filled in by that surface rather than hardcoded here,
	// because this package cannot know which command is running and an error
	// naming a flag the command does not accept is worse than one naming none.
	// The default says only what is true everywhere.
	Remedy string
}

func (e *AmbiguousTableError) Error() string {
	if e == nil {
		return "warehouse table is ambiguous"
	}
	// A table written straight through the connector Write surface has no
	// owning connection. Naming it by the selector that reaches it is better
	// than printing an empty entry, and better than quietly dropping it from
	// the count.
	named := make([]string, 0, len(e.Connections))
	for _, connection := range e.Connections {
		if connection == "" {
			named = append(named, UnattributedConnection)
			continue
		}
		named = append(named, connection)
	}
	remedy := e.Remedy
	if remedy == "" {
		remedy = "the read must name the connection it means"
	}
	return fmt.Sprintf(
		"table %q is materialized by %d connections (%s); %s",
		e.Table, len(named), strings.Join(named, ", "), remedy,
	)
}

// WithAmbiguityRemedy tells a table-ambiguity error what the calling surface
// can actually offer, and returns err untouched when it is not one. Callers use
// it so the advice matches the command the operator just ran.
func WithAmbiguityRemedy(err error, remedy string) error {
	var ambiguous *AmbiguousTableError
	if errors.As(err, &ambiguous) {
		ambiguous.Remedy = remedy
	}
	return err
}

// Location is one connection's private region of a warehouse root.
type Location struct {
	Root          string
	ConnectionDir string
	Owner         Owner
}

// LocationFor resolves the directory a connection owns inside root. Every
// component is validated before it reaches filepath.Join, so an invalid
// identity fails here rather than escaping into a neighbour's directory.
func LocationFor(root, workspaceID, connector, connectionID, displayName string) (Location, error) {
	workspace, err := PathComponent("workspace id", workspaceID)
	if err != nil {
		return Location{}, err
	}
	connectorPart, err := PathComponent("connector", connector)
	if err != nil {
		return Location{}, err
	}
	connection, err := PathComponent("connection id", connectionID)
	if err != nil {
		return Location{}, err
	}
	return Location{
		Root:          root,
		ConnectionDir: filepath.Join(root, workspace, connectorPart, connection),
		Owner: Owner{
			Version:     OwnerVersion,
			Workspace:   workspace,
			Connector:   connectorPart,
			Connection:  connection,
			DisplayName: displayName,
		},
	}, nil
}

// OwnerPath is the location's ownership record path.
func (l Location) OwnerPath() string {
	return filepath.Join(l.ConnectionDir, OwnerFileName)
}

// WALPath is the append-only log for one stream.
func (l Location) WALPath(stream string) (string, error) {
	component, err := PathComponent("stream", stream)
	if err != nil {
		return "", err
	}
	return filepath.Join(l.ConnectionDir, WALDirName, component+".jsonl"), nil
}

// TablePath is the derived materialization for one table.
func (l Location) TablePath(table string) (string, error) {
	component, err := PathComponent("table", table)
	if err != nil {
		return "", err
	}
	return filepath.Join(l.ConnectionDir, TablesDirName, component+".jsonl"), nil
}

// EnsureOwnership creates the connection directory and its ownership record,
// or verifies that an existing record names this connection. A mismatch is a
// hard error, never a silent overwrite.
//
// The record's own bytes are fsynced before it is renamed into place, but its
// directory entry becomes durable with the connection directory chain the
// caller syncs before acknowledging a run — every wal/ and tables/ ancestor
// chain passes through this directory. A run that dies in between simply
// rewrites the record next time; no row is ever misattributed by its absence.
func (l Location) EnsureOwnership() error {
	if err := os.MkdirAll(l.ConnectionDir, 0o700); err != nil {
		return fmt.Errorf("create warehouse connection directory: %w", err)
	}
	stored, err := readOwner(l.ConnectionDir)
	if errors.Is(err, os.ErrNotExist) {
		owner := l.Owner
		owner.CreatedAt = time.Now().UTC()
		if err := writeJSONAtomic(l.OwnerPath(), owner); err != nil {
			return fmt.Errorf("write warehouse ownership record: %w", err)
		}
		return nil
	}
	if err != nil {
		return &OwnershipError{
			Dir:            l.ConnectionDir,
			WriterName:     l.Owner.DisplayName,
			WriterID:       l.Owner.Connection,
			UnreadableFile: l.OwnerPath(),
		}
	}
	if !stored.SameIdentity(l.Owner) {
		return &OwnershipError{
			Dir:        l.ConnectionDir,
			OwnerName:  stored.DisplayName,
			OwnerID:    stored.Connection,
			WriterName: l.Owner.DisplayName,
			WriterID:   l.Owner.Connection,
		}
	}
	if stored.DisplayName != l.Owner.DisplayName {
		// Same connection under a new name. The directory does not move,
		// because it was never keyed on the name; only the record catches up,
		// so reads that scope by connection name keep resolving.
		stored.DisplayName = l.Owner.DisplayName
		if err := writeJSONAtomic(l.OwnerPath(), stored); err != nil {
			return fmt.Errorf("update warehouse ownership record: %w", err)
		}
	}
	return nil
}

// AssertOwnedTable re-derives ownership from the table path itself rather than
// from the Location that produced it. Directory nesting already makes a shared
// table path unrepresentable; this is the independent second check that fails
// loudly if a future change ever reintroduces one, instead of silently
// overwriting another connection's rows.
func (l Location) AssertOwnedTable(path, table string) error {
	dir := filepath.Dir(filepath.Dir(path))
	stored, err := readOwner(dir)
	if errors.Is(err, os.ErrNotExist) {
		return &OwnershipError{
			Dir:          dir,
			Table:        table,
			WriterName:   l.Owner.DisplayName,
			WriterID:     l.Owner.Connection,
			MissingOwner: true,
		}
	}
	if err != nil {
		return &OwnershipError{
			Dir:            dir,
			Table:          table,
			WriterName:     l.Owner.DisplayName,
			WriterID:       l.Owner.Connection,
			UnreadableFile: filepath.Join(dir, OwnerFileName),
		}
	}
	if !stored.SameIdentity(l.Owner) {
		return &OwnershipError{
			Dir:        dir,
			Table:      table,
			OwnerName:  stored.DisplayName,
			OwnerID:    stored.Connection,
			WriterName: l.Owner.DisplayName,
			WriterID:   l.Owner.Connection,
		}
	}
	return nil
}

// CheckLegacyLayout refuses a warehouse root written by the removed flat
// layout. Only the old materializer created _pm_raw/, so its presence is
// unambiguous evidence rather than a heuristic — and it is always present in a
// warehouse that layout ever synced, because the raw log was written on every
// run. A root-level table with no _pm_raw beside it was therefore not produced
// by the shared-table layout, so it is left alone rather than condemned.
func CheckLegacyLayout(root string) error {
	legacy := filepath.Join(root, LegacyRawDirName)
	info, err := os.Stat(legacy)
	if err != nil || !info.IsDir() {
		return nil
	}
	return &LegacyLayoutError{Dir: root, Evidence: legacy}
}

// Table is one materialized table located in the nested layout.
type Table struct {
	Path       string
	Name       string
	Connection string
	Owner      Owner
}

// Tables lists every materialized table under root, sorted by table name then
// owning connection. A connection directory with no ownership record at all is
// skipped: an unowned directory is not a table anyone may read as if it
// belonged to them. A record that exists but cannot be read is reported
// instead, because the write path already refuses exactly that case; silently
// dropping the rows here would tell the operator the table does not exist
// while it sits intact on disk.
func Tables(root string) ([]Table, error) {
	if err := CheckLegacyLayout(root); err != nil {
		return nil, err
	}
	out := make([]Table, 0)
	// <root>/<workspace>/<connector>/<connection>/tables/<table>.jsonl
	matches, err := filepath.Glob(filepath.Join(root, "*", "*", "*", TablesDirName, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("scan warehouse tables: %w", err)
	}
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		connectionDir := filepath.Dir(filepath.Dir(path))
		owner, err := readOwner(connectionDir)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("warehouse directory %s has an unreadable ownership record at %s: %w", connectionDir, filepath.Join(connectionDir, OwnerFileName), err)
		}
		out = append(out, Table{
			Path:       path,
			Name:       strings.TrimSuffix(filepath.Base(path), ".jsonl"),
			Connection: owner.DisplayName,
			Owner:      owner,
		})
	}
	// Root-level tables are the unattributed ones: written straight through the
	// connector Write surface, or seeded by hand. They carry no connection
	// identity, and none is invented for them — attributing these rows to a
	// connection that never produced them is the very confusion this layout
	// removes. A sync never writes here.
	unattributed, err := filepath.Glob(filepath.Join(root, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("scan warehouse root tables: %w", err)
	}
	for _, path := range unattributed {
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			continue
		}
		out = append(out, Table{
			Path: path,
			Name: strings.TrimSuffix(filepath.Base(path), ".jsonl"),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Connection < out[j].Connection
	})
	return out, nil
}

// FindTable resolves one table by name, optionally scoped to a connection
// display name or to UnattributedConnection for the root-level tables no
// connection owns. It reports ambiguity rather than picking a winner.
func FindTable(root, table, connection string) (Table, error) {
	tables, err := Tables(root)
	if err != nil {
		return Table{}, err
	}
	matches := make([]Table, 0, 1)
	for _, candidate := range tables {
		if candidate.Name != table {
			continue
		}
		if !selects(connection, candidate.Connection) {
			continue
		}
		matches = append(matches, candidate)
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		switch connection {
		case "":
			return Table{}, fmt.Errorf("no warehouse table %q; run a sync to materialize it", table)
		case UnattributedConnection:
			return Table{}, fmt.Errorf("no unattributed warehouse table %q at the warehouse root", table)
		default:
			return Table{}, fmt.Errorf("connection %q has no warehouse table %q", connection, table)
		}
	default:
		names := make([]string, 0, len(matches))
		for _, match := range matches {
			names = append(names, match.Connection)
		}
		return Table{}, &AmbiguousTableError{Table: table, Connections: names}
	}
}

// selects reports whether a requested connection selector matches a candidate
// table's owning connection. The empty selector means "any", so it can never
// name the unattributed tables; UnattributedConnection is the selector that
// does.
func selects(requested, owner string) bool {
	switch requested {
	case "":
		return true
	case UnattributedConnection:
		return owner == ""
	default:
		return owner == requested
	}
}

func readOwner(connectionDir string) (Owner, error) {
	raw, err := os.ReadFile(filepath.Join(connectionDir, OwnerFileName))
	if err != nil {
		return Owner{}, err
	}
	var owner Owner
	if err := json.Unmarshal(raw, &owner); err != nil {
		return Owner{}, fmt.Errorf("decode warehouse ownership record: %w", err)
	}
	if owner.Version != OwnerVersion {
		return Owner{}, fmt.Errorf("warehouse ownership record version %d is unsupported", owner.Version)
	}
	return owner, nil
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	payload = append(payload, '\n')
	temp, err := os.CreateTemp(filepath.Dir(path), ".owner.tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}
	tempPath := temp.Name()
	defer func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}()
	if _, err := temp.Write(payload); err != nil {
		return fmt.Errorf("write temporary file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("sync temporary file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace file: %w", err)
	}
	return nil
}
