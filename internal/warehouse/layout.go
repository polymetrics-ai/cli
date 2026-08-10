// Package warehouse defines the on-disk layout of the local warehouse.
//
// Every materialization is nested under the identity that produced it:
//
//	<root>/<workspace-id>/<connector>/<connection-id>/
//	    owner.json           identity of the connection owning this directory
//	    wal/<stream>.jsonl     append-only log; the source of truth, stays JSONL
//	    tables/<table>.parquet derived materialization, rewritten wholesale
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
	// TableFileExt is the extension of a materialized table. A table is one
	// Parquet file, never a directory of parts: a directory cannot be renamed
	// into place over an existing one, so swapping it opens a window in which a
	// reader sees no table at all while its rows sit intact on disk. Parts were
	// measured and bought no read or write parallelism at our scale.
	TableFileExt = ".parquet"
	// LegacyTableFileExt is the extension materialized tables had before the
	// Parquet switch. It is only ever read, as evidence that a warehouse
	// predates this format.
	LegacyTableFileExt = ".jsonl"
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

// Identity returns the shared structural owner triple for this warehouse
// region. DisplayName is deliberately absent: renaming a connection does not
// change which data is its own.
func (o Owner) Identity() ArtifactIdentity {
	return ArtifactIdentity{
		WorkspaceID:  o.Workspace,
		ConnectorID:  o.Connector,
		ConnectionID: o.Connection,
	}
}

// SameIdentity reports whether two owner records describe the same connection
// through the shared artifact identity rule.
func (o Owner) SameIdentity(other Owner) bool {
	return o.Identity().SameIdentity(other.Identity())
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

// LegacyTableFormatError reports a warehouse whose tables were materialized as
// JSONL, before Parquet became the table format. It is not read and not
// deleted: reading it would work today and be silently stale the moment a sync
// writes the Parquet table beside it, leaving two files for one table name with
// no way for a reader to tell which is current. The write-ahead log is
// untouched, so re-running the sync rebuilds the table losslessly.
type LegacyTableFormatError struct {
	Files []string
}

func (e *LegacyTableFormatError) Error() string {
	if e == nil || len(e.Files) == 0 {
		return "warehouse table format is unavailable"
	}
	return fmt.Sprintf(
		"warehouse tables are stored as Parquet, but %d table(s) are still JSONL (%s). "+
			"These predate the Parquet format and are not read, because a sync would write the Parquet table "+
			"beside them and leave two files for one table name. "+
			"Delete them and re-run the owning connection's sync to rebuild each table from its write-ahead log, "+
			"which is untouched; pm will not rewrite or delete warehouse data for you",
		len(e.Files), strings.Join(e.Files, ", "),
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

// TablePath is the derived materialization for one table: a single Parquet
// file, never a directory of parts. See TableFileExt for why.
func (l Location) TablePath(table string) (string, error) {
	component, err := PathComponent("table", table)
	if err != nil {
		return "", err
	}
	return filepath.Join(l.ConnectionDir, TablesDirName, component+TableFileExt), nil
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

// CheckLegacyTableFormat refuses a warehouse whose tables were materialized as
// JSONL, before Parquet became the table format.
//
// It is a refusal rather than a silent skip for the same reason a damaged
// ownership record is: the rows are on disk, so reporting the table as absent
// would be a wrong answer. It is a refusal rather than a read because a sync
// writes `<table>.parquet` beside the JSONL and never removes it — two files
// for one table name, with nothing on disk saying which is current. Nothing is
// deleted on the operator's behalf; the write-ahead log beside these files is
// untouched, so a re-run rebuilds each table losslessly.
//
// Root-level tables are inspected too. They are the unattributed direct writes
// no connection owns, and a stale one there is missed by exactly the same
// Parquet-only glob, so ignoring it would report an absent table just as
// wrongly.
func CheckLegacyTableFormat(root string) error {
	patterns := []string{
		filepath.Join(root, "*", "*", "*", TablesDirName, "*"+LegacyTableFileExt),
		filepath.Join(root, "*"+LegacyTableFileExt),
	}
	files := make([]string, 0)
	for _, pattern := range patterns {
		stale, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("scan warehouse tables: %w", err)
		}
		for _, path := range stale {
			if info, err := os.Stat(path); err != nil || info.IsDir() {
				continue
			}
			files = append(files, path)
		}
	}
	if len(files) == 0 {
		return nil
	}
	sort.Strings(files)
	return &LegacyTableFormatError{Files: files}
}

// Table is one materialized table located in the nested layout.
type Table struct {
	Path       string
	Name       string
	Connection string
	Owner      Owner
}

// Fault is a connection directory whose ownership record exists but cannot be
// read. It is deliberately NOT an error returned from the scan: the whole point
// of nesting per connection is that one connection's problem stays that
// connection's problem, so a single damaged record must not deny reads of
// tables owned by connections whose records are healthy. Callers surface a
// fault when it could explain what they failed to find, or when it could hide
// a second holder of what they did find, and otherwise carry on.
type Fault struct {
	Dir  string
	File string
	Err  error
	// Tables names the tables materialized in the damaged directory. A table
	// file is named after the table, so which names a directory holds is
	// readable without its ownership record — and it has to be, because those
	// names are candidates a lookup would otherwise never learn about.
	Tables []string
}

// Holds reports whether the damaged directory materializes a table of this
// name, and so could be a candidate a lookup cannot see.
func (f Fault) Holds(table string) bool {
	for _, name := range f.Tables {
		if name == table {
			return true
		}
	}
	return false
}

func (f Fault) Error() string {
	return fmt.Sprintf("warehouse directory %s has an unreadable ownership record at %s: %v", f.Dir, f.File, f.Err)
}

func (f Fault) Unwrap() error { return f.Err }

// FaultError reports that a read could not be answered while one or more
// ownership records were unreadable. It names every damaged record and the
// recovery step, so the operator is never told a table does not exist when a
// damaged record is the reason it was not found.
type FaultError struct {
	Table  string
	Faults []Fault
	// Undecided reports that the table did resolve to exactly one healthy
	// connection and was refused anyway, because a damaged directory
	// materializes the same name and so may be the connection the read meant.
	// An unreadable ownership record must never be what decides which
	// connection's rows an unscoped read returns: dropping a hidden candidate
	// turns an ambiguous question into a confidently wrong answer, which is
	// the failure this layout exists to remove.
	Undecided bool
}

func (e *FaultError) Error() string {
	if e == nil || len(e.Faults) == 0 {
		return "warehouse ownership records are unreadable"
	}
	records := make([]string, 0, len(e.Faults))
	for _, fault := range e.Faults {
		records = append(records, fault.File)
	}
	subject := "the warehouse"
	if e.Table != "" {
		subject = fmt.Sprintf("warehouse table %q", e.Table)
	}
	recovery := "Restore or remove the damaged record and re-run the connection's sync to rewrite it; " +
		"pm will not rewrite or delete warehouse data for you"
	if e.Undecided {
		return fmt.Sprintf(
			"%s is materialized by one connection with a healthy ownership record and by %d whose record is unreadable (%s), "+
				"so which connection an unscoped read means cannot be decided. "+
				"Name the connection to read one of them now. %s",
			subject, len(e.Faults), strings.Join(records, ", "), recovery,
		)
	}
	return fmt.Sprintf(
		"%s could not be resolved because %d ownership record(s) are unreadable (%s). "+
			"Tables owned by connections with healthy records are unaffected. %s",
		subject, len(e.Faults), strings.Join(records, ", "), recovery,
	)
}

// FaultsError builds the error a caller returns when it could not answer and
// unreadable ownership records could be the reason. It returns nil when there
// are no faults, so callers can use it unconditionally. Every fault is
// relevant when a lookup came back empty, because any of them could be the
// directory the missing table lived in.
func FaultsError(table string, faults []Fault) error {
	if len(faults) == 0 {
		return nil
	}
	return &FaultError{Table: table, Faults: faults}
}

// Tables lists every materialized table under root, sorted by table name then
// owning connection, alongside any connection directories whose ownership
// record could not be read.
//
// A connection directory with no ownership record at all is skipped: an unowned
// directory is not a table anyone may read as if it belonged to them. A record
// that exists but cannot be read is a Fault rather than a skip, because the
// write path already refuses exactly that case; silently dropping the rows here
// would tell the operator the table does not exist while it sits intact on disk.
//
// A fault is returned beside the healthy tables rather than in place of them.
// Aborting the scan would turn one damaged file into a total outage that also
// denied every healthy connection's tables, which is the opposite of what
// nesting per connection is for: one connection's problem stays that
// connection's problem.
func Tables(root string) ([]Table, []Fault, error) {
	if err := CheckLegacyLayout(root); err != nil {
		return nil, nil, err
	}
	if err := CheckLegacyTableFormat(root); err != nil {
		return nil, nil, err
	}
	out := make([]Table, 0)
	faults := make([]Fault, 0)
	// <root>/<workspace>/<connector>/<connection>/tables/<table>.parquet
	matches, err := filepath.Glob(filepath.Join(root, "*", "*", "*", TablesDirName, "*"+TableFileExt))
	if err != nil {
		return nil, nil, fmt.Errorf("scan warehouse tables: %w", err)
	}
	faulted := make(map[string]int)
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(path), TableFileExt)
		connectionDir := filepath.Dir(filepath.Dir(path))
		owner, err := readOwner(connectionDir)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			// One fault per damaged directory, however many tables it holds,
			// but every one of those tables is recorded on it: a name a lookup
			// cannot see is exactly what makes an answer unsafe to give.
			index, seen := faulted[connectionDir]
			if !seen {
				index = len(faults)
				faulted[connectionDir] = index
				faults = append(faults, Fault{
					Dir:  connectionDir,
					File: filepath.Join(connectionDir, OwnerFileName),
					Err:  err,
				})
			}
			faults[index].Tables = append(faults[index].Tables, name)
			continue
		}
		out = append(out, Table{
			Path:       path,
			Name:       name,
			Connection: owner.DisplayName,
			Owner:      owner,
		})
	}
	// Root-level tables are the unattributed ones: written straight through the
	// connector Write surface, or seeded by hand. They carry no connection
	// identity, and none is invented for them — attributing these rows to a
	// connection that never produced them is the very confusion this layout
	// removes. A sync never writes here.
	unattributed, err := filepath.Glob(filepath.Join(root, "*"+TableFileExt))
	if err != nil {
		return nil, nil, fmt.Errorf("scan warehouse root tables: %w", err)
	}
	for _, path := range unattributed {
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			continue
		}
		out = append(out, Table{
			Path: path,
			Name: strings.TrimSuffix(filepath.Base(path), TableFileExt),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Connection < out[j].Connection
	})
	sort.Slice(faults, func(i, j int) bool { return faults[i].Dir < faults[j].Dir })
	return out, faults, nil
}

// FindTable resolves one table by name, optionally scoped to a connection
// display name or to UnattributedConnection for the root-level tables no
// connection owns. It reports ambiguity rather than picking a winner, and
// treats a table it cannot see behind a damaged ownership record as ambiguity
// too.
func FindTable(root, table, connection string) (Table, error) {
	tables, faults, err := Tables(root)
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
		// One match is only an unambiguous answer if nothing could have been
		// missed. The empty selector means "any connection", so it is decided
		// by every directory under the root — including one whose ownership
		// record cannot be read, whose tables Tables() therefore had to drop.
		// A damaged record that hides a competing candidate would otherwise
		// silently turn an ambiguous question into one tenant's rows.
		//
		// Only the damaged directories that hold a table of this name can be
		// that candidate, and a table file is named after its table, so this
		// stays precise: a read that genuinely does not depend on a damaged
		// directory is still answered. A named connection selector is answered
		// too, because it says which connection it means, and the unattributed
		// selector reaches only root-level files, which no connection
		// directory can hold.
		if connection == "" {
			if hidden := faultsHolding(table, faults); len(hidden) > 0 {
				return Table{}, &FaultError{Table: table, Faults: hidden, Undecided: true}
			}
		}
		return matches[0], nil
	case 0:
		// A damaged ownership record could be exactly why nothing matched, so
		// it is reported instead of claiming the table does not exist. Telling
		// an operator a table is absent while its rows sit on disk is the
		// silent-absence failure this layout exists to remove.
		if err := FaultsError(table, faults); err != nil {
			return Table{}, err
		}
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

// faultsHolding returns the damaged directories that materialize a table of
// this name, and so could each be a candidate a lookup cannot see.
func faultsHolding(table string, faults []Fault) []Fault {
	hidden := make([]Fault, 0, len(faults))
	for _, fault := range faults {
		if fault.Holds(table) {
			hidden = append(hidden, fault)
		}
	}
	return hidden
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

// readOwner decodes a connection directory's ownership record, and accepts one
// only if it says which connection it belongs to. A record missing any part of
// its identity is damaged, not unowned: an empty DisplayName would otherwise
// reach the read path as the sentinel for a root-level table owned by nobody,
// and answer to UnattributedConnection. ValidateConnectionName reserves that
// selector against every real connection name; this reserves it against every
// record that cannot name one.
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
	for _, field := range []struct{ name, value string }{
		{"workspace", owner.Workspace},
		{"connector", owner.Connector},
		{"connection", owner.Connection},
		{"display_name", owner.DisplayName},
	} {
		if field.value == "" {
			return Owner{}, fmt.Errorf("warehouse ownership record has no %s", field.name)
		}
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
