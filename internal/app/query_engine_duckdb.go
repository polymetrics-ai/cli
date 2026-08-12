package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	duckdb "github.com/marcboeker/go-duckdb"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/warehouse"
)

// newSQLEngine returns the DuckDB-backed analytical engine. It is the only
// engine: an optional one would mean two builds that write different table
// formats, which is install-time drift rather than a build option.
func newSQLEngine(a *App) sqlQueryEngine {
	return duckdbEngine{warehouseDir: filepath.Join(a.projectDir, "warehouse")}
}

// duckdbEngine runs read-only analytical SQL over the Parquet warehouse using
// an in-memory DuckDB instance. It is stateless: each query opens a fresh
// connection and registers per-query views over the current warehouse files.
type duckdbEngine struct {
	warehouseDir     string
	newTableResolver func(string) (*warehouse.TableResolver, error)
}

func (e duckdbEngine) Name() string { return "duckdb" }

var (
	// Statement-leading keywords / tokens that must never appear in a query.
	forbiddenTokens = []string{
		"insert", "update", "delete", "drop", "alter", "create",
		"attach", "copy", "pragma", "call", "export", "install", "load", "set",
	}
	wordRe = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)
)

// validateSelectOnly enforces that sql is a single read-only statement. It must
// begin with SELECT or WITH, must not chain statements via ';', and must not
// contain any DDL/DML/side-effecting whole-word tokens.
func validateSelectOnly(sql string) error {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return fmt.Errorf("empty query is not allowed")
	}
	if strings.Contains(trimmed, ";") {
		return fmt.Errorf("statement chaining (';') is not allowed")
	}
	lower := strings.ToLower(trimmed)
	if !strings.HasPrefix(lower, "select") && !strings.HasPrefix(lower, "with") {
		return fmt.Errorf("only SELECT/WITH queries are allowed")
	}
	forbidden := make(map[string]struct{}, len(forbiddenTokens))
	for _, tok := range forbiddenTokens {
		forbidden[tok] = struct{}{}
	}
	for _, word := range wordRe.FindAllString(lower, -1) {
		if _, bad := forbidden[word]; bad {
			return fmt.Errorf("token %q is not allowed in a read-only query", word)
		}
	}
	return nil
}

func (e duckdbEngine) QuerySQL(ctx context.Context, req QuerySQLRequest) ([]connectors.Record, error) {
	if err := validateSelectOnly(req.SQL); err != nil {
		return nil, err
	}
	resolver, err := e.tableResolver()
	if err != nil {
		return nil, err
	}
	policy := newQueryViewPolicy(req, resolver)

	db, lookupError, err := e.openScopedDuckDB(req.Connection, resolver, policy)
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := e.registerViews(ctx, db, req.Connection, resolver, policy); err != nil {
		return nil, err
	}

	finalSQL := req.SQL
	if req.Limit > 0 && !hasTopLevelLimit(req.SQL) {
		finalSQL = fmt.Sprintf("SELECT * FROM (%s) AS _pm_q LIMIT %d", req.SQL, req.Limit)
	}

	rows, err := db.QueryContext(ctx, finalSQL)
	if err != nil {
		if lookupErr := lookupError(); lookupErr != nil {
			return nil, fmt.Errorf("execute query: %w", lookupErr)
		}
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read columns: %w", err)
	}

	out := make([]connectors.Record, 0)
	for rows.Next() {
		holders := make([]any, len(cols))
		targets := make([]any, len(cols))
		for i := range holders {
			targets[i] = &holders[i]
		}
		if err := rows.Scan(targets...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		rec := make(connectors.Record, len(cols))
		for i, col := range cols {
			rec[col] = warehouse.NormalizeValue(holders[i])
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}
	return out, nil
}

func (e duckdbEngine) registerViews(
	ctx context.Context,
	db *sql.DB,
	connection string,
	resolver *warehouse.TableResolver,
	policy queryViewPolicy,
) error {
	tables := resolver.Tables()
	byName := make(map[string][]warehouse.Table, len(tables))
	for _, table := range tables {
		if !querySelectsConnection(connection, table.Connection) {
			continue
		}
		byName[table.Name] = append(byName[table.Name], table)
	}
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if policy.blocksBareView(name) {
			if connection == "" && policy.registersCanonicalAliases(name) {
				if err := registerGeneratedOwnerAliases(ctx, db, name, byName[name], policy); err != nil {
					return err
				}
			}
			continue
		}
		table, findErr := resolver.Find(name, connection)
		if findErr == nil {
			if err := registerWarehouseView(ctx, db, name, table); err != nil {
				return err
			}
			continue
		}
		if connection != "" {
			continue
		}
		var ambiguous *warehouse.AmbiguousTableError
		if !errors.As(findErr, &ambiguous) {
			continue
		}
		if policy.blocksGeneratedAliases(name) {
			continue
		}
		if err := registerGeneratedOwnerAliases(ctx, db, name, byName[name], policy); err != nil {
			return err
		}
	}
	return nil
}

func registerGeneratedOwnerAliases(
	ctx context.Context,
	db *sql.DB,
	name string,
	tables []warehouse.Table,
	policy queryViewPolicy,
) error {
	for _, table := range tables {
		alias := generatedOwnerAlias(name, table)
		if !policy.allowsGeneratedAlias(alias) {
			continue
		}
		if err := registerWarehouseView(ctx, db, alias, table); err != nil {
			return err
		}
	}
	return nil
}

func (e duckdbEngine) tableResolver() (*warehouse.TableResolver, error) {
	if e.newTableResolver != nil {
		return e.newTableResolver(e.warehouseDir)
	}
	return warehouse.NewTableResolver(e.warehouseDir)
}

func (e duckdbEngine) openScopedDuckDB(
	connection string,
	resolver *warehouse.TableResolver,
	policy queryViewPolicy,
) (*sql.DB, func() error, error) {
	connector, err := duckdb.NewConnector("", nil)
	if err != nil {
		return nil, nil, err
	}
	var lookupMu sync.Mutex
	var firstLookupErr error
	duckdb.RegisterReplacementScan(connector, func(tableName string) (string, []any, error) {
		table, err := policy.find(resolver, tableName, connection)
		if err != nil {
			lookupMu.Lock()
			if firstLookupErr == nil {
				firstLookupErr = err
			}
			lookupMu.Unlock()
			return "", nil, err
		}
		return "read_parquet", []any{table.Path}, nil
	})
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(1)
	return db, func() error {
		lookupMu.Lock()
		defer lookupMu.Unlock()
		return firstLookupErr
	}, nil
}

type queryViewPolicy struct {
	ambiguousBaseNameByKey             map[string]string
	caseEquivalentUniqueAmbiguityByKey map[string]tableAmbiguity
	faultedCanonicalTableByKey         map[string]faultedCanonicalTable
	generatedAliasBaseName             map[string]string
	collidingGeneratedAliases          map[string]struct{}
	sameOwnerDestinationCollisionByKey map[string]warehouseDestinationCollision
	sameOwnerGeneratedAliasByKey       map[string]warehouseDestinationCollision
	unscopedFlow                       bool
}

// tableAmbiguity is a canonical DuckDB identifier collision derived from a
// single immutable resolver snapshot. Its original table spelling and
// connection list are retained so the flow boundary can still give the caller
// a typed warehouse error rather than a DuckDB catalog implementation detail.
type tableAmbiguity struct {
	table       string
	connections []string
}

func (a tableAmbiguity) err() error {
	return &warehouse.AmbiguousTableError{
		Table:       a.table,
		Connections: append([]string(nil), a.connections...),
	}
}

type faultedCanonicalTable struct {
	faults []warehouse.Fault
}

func (f faultedCanonicalTable) err(table string) error {
	faults := make([]warehouse.Fault, len(f.faults))
	for i, fault := range f.faults {
		faults[i] = fault
		faults[i].Tables = append([]string(nil), fault.Tables...)
	}
	return &warehouse.FaultError{Table: table, Faults: faults, Undecided: true}
}

func newQueryViewPolicy(req QuerySQLRequest, resolver *warehouse.TableResolver) queryViewPolicy {
	tables := resolver.Tables()
	byName := make(map[string][]warehouse.Table, len(tables))
	catalogNames := make(map[string]struct{}, len(tables))
	for _, table := range tables {
		byName[table.Name] = append(byName[table.Name], table)
		if querySelectsConnection(req.Connection, table.Connection) {
			catalogNames[duckDBIdentifierKey(table.Name)] = struct{}{}
		}
	}
	policy := queryViewPolicy{
		caseEquivalentUniqueAmbiguityByKey: make(map[string]tableAmbiguity),
		collidingGeneratedAliases:          make(map[string]struct{}),
		sameOwnerDestinationCollisionByKey: make(map[string]warehouseDestinationCollision),
		sameOwnerGeneratedAliasByKey:       make(map[string]warehouseDestinationCollision),
	}
	policy.unscopedFlow = req.Connection == "" && req.Origin == QuerySQLOriginFlow
	for _, collision := range req.sameOwnerCaseEquivalentDestinationCollisions {
		if req.Connection != "" && req.Connection != collision.connection {
			continue
		}
		if _, exists := policy.sameOwnerDestinationCollisionByKey[collision.key]; !exists {
			policy.sameOwnerDestinationCollisionByKey[collision.key] = collision
		}
		for _, table := range collision.tables {
			aliasKey := duckDBIdentifierKey(table + "__" + collision.connectionID)
			// A resolver-visible real table wins over an invented generated
			// alias. The snapshot is the authority, not a reserved namespace.
			if _, exists := catalogNames[aliasKey]; exists {
				continue
			}
			if _, exists := policy.sameOwnerGeneratedAliasByKey[aliasKey]; !exists {
				policy.sameOwnerGeneratedAliasByKey[aliasKey] = collision
			}
		}
	}
	// A connection-scoped query still needs the same-owner policy above, but
	// all existing cross-owner/alias rules remain unscoped-only.
	if req.Connection != "" {
		return policy
	}
	if policy.unscopedFlow {
		policy.ambiguousBaseNameByKey = make(map[string]string)
		policy.generatedAliasBaseName = make(map[string]string)
	}
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)
	canonicalGroups := make(map[string]canonicalTableGroup, len(names))
	for _, name := range names {
		nameTables := byName[name]
		key := duckDBIdentifierKey(name)
		group := canonicalGroups[key]
		group.names = append(group.names, name)
		group.tables = append(group.tables, nameTables...)
		_, findErr := resolver.Find(name, "")
		exactAmbiguous := isAmbiguousTableError(findErr)
		if exactAmbiguous {
			group.hasExactAmbiguity = true
			if policy.unscopedFlow {
				current, exists := policy.ambiguousBaseNameByKey[key]
				if !exists || name == key || (current != key && name < current) {
					policy.ambiguousBaseNameByKey[key] = name
				}
			}
		}
		canonicalGroups[key] = group

		for _, table := range nameTables {
			alias := generatedOwnerAlias(name, table)
			aliasKey := duckDBIdentifierKey(alias)
			if _, exists := catalogNames[aliasKey]; exists {
				policy.collidingGeneratedAliases[aliasKey] = struct{}{}
			}
			if exactAmbiguous && policy.unscopedFlow {
				policy.generatedAliasBaseName[aliasKey] = name
			}
		}
	}
	for key, group := range canonicalGroups {
		if len(group.names) < 2 || group.hasExactAmbiguity {
			continue
		}
		policy.caseEquivalentUniqueAmbiguityByKey[key] = group.ambiguity(key)
	}
	if policy.unscopedFlow {
		policy.faultedCanonicalTableByKey = faultedCanonicalTables(resolver.Faults(), canonicalGroups)
	}
	return policy
}

type canonicalTableGroup struct {
	names             []string
	tables            []warehouse.Table
	hasExactAmbiguity bool
}

func (g canonicalTableGroup) ambiguity(key string) tableAmbiguity {
	table := g.names[0]
	for _, name := range g.names {
		if name == key {
			table = name
			break
		}
	}
	connections := make([]string, 0, len(g.tables))
	seenConnections := make(map[string]struct{}, len(g.tables))
	for _, table := range g.tables {
		if _, exists := seenConnections[table.Connection]; exists {
			continue
		}
		seenConnections[table.Connection] = struct{}{}
		connections = append(connections, table.Connection)
	}
	sort.Strings(connections)
	return tableAmbiguity{table: table, connections: connections}
}

func faultedCanonicalTables(faults []warehouse.Fault, groups map[string]canonicalTableGroup) map[string]faultedCanonicalTable {
	knownNamesByKey := make(map[string]map[string]struct{}, len(groups))
	for key, group := range groups {
		names := make(map[string]struct{}, len(group.names))
		for _, name := range group.names {
			names[name] = struct{}{}
		}
		knownNamesByKey[key] = names
	}

	out := make(map[string]faultedCanonicalTable)
	for _, fault := range faults {
		seenKeys := make(map[string]struct{})
		for _, name := range fault.Tables {
			key := duckDBIdentifierKey(name)
			knownNames, exists := knownNamesByKey[key]
			if !exists {
				continue
			}
			if _, exact := knownNames[name]; exact {
				continue
			}
			if _, seen := seenKeys[key]; seen {
				continue
			}
			seenKeys[key] = struct{}{}
			collision := out[key]
			collision.faults = append(collision.faults, fault)
			out[key] = collision
		}
	}
	return out
}

func isAmbiguousTableError(err error) bool {
	var ambiguous *warehouse.AmbiguousTableError
	return errors.As(err, &ambiguous)
}

func generatedOwnerAlias(name string, table warehouse.Table) string {
	return name + "__" + table.Owner.Connection
}

func (p queryViewPolicy) blocksGeneratedAliases(name string) bool {
	if _, ok := p.sameOwnerDestinationCollisionByKey[duckDBIdentifierKey(name)]; ok {
		return true
	}
	_, ok := p.ambiguousBaseNameByKey[duckDBIdentifierKey(name)]
	return ok
}

func (p queryViewPolicy) blocksBareView(name string) bool {
	key := duckDBIdentifierKey(name)
	if _, ok := p.sameOwnerDestinationCollisionByKey[key]; ok {
		return true
	}
	if _, ok := p.faultedCanonicalTableByKey[key]; ok {
		return true
	}
	if _, ok := p.sameOwnerGeneratedAliasByKey[key]; ok {
		return true
	}
	if _, ok := p.ambiguousBaseNameByKey[key]; ok {
		return true
	}
	if _, ok := p.caseEquivalentUniqueAmbiguityByKey[key]; ok {
		return true
	}
	_, ok := p.generatedAliasBaseName[key]
	return ok
}

func (p queryViewPolicy) registersCanonicalAliases(name string) bool {
	if p.unscopedFlow {
		return false
	}
	if _, ok := p.sameOwnerDestinationCollisionByKey[duckDBIdentifierKey(name)]; ok {
		return false
	}
	_, ok := p.caseEquivalentUniqueAmbiguityByKey[duckDBIdentifierKey(name)]
	return ok
}

func (p queryViewPolicy) allowsGeneratedAlias(name string) bool {
	if _, blocked := p.sameOwnerGeneratedAliasByKey[duckDBIdentifierKey(name)]; blocked {
		return false
	}
	_, collides := p.collidingGeneratedAliases[duckDBIdentifierKey(name)]
	return !collides
}

func (p queryViewPolicy) find(resolver *warehouse.TableResolver, tableName, connection string) (warehouse.Table, error) {
	key := duckDBIdentifierKey(tableName)
	if collision, ok := p.sameOwnerDestinationCollisionByKey[key]; ok {
		return warehouse.Table{}, collision.err()
	}
	if faulted, ok := p.faultedCanonicalTableByKey[key]; ok {
		return warehouse.Table{}, faulted.err(tableName)
	}
	if collision, ok := p.sameOwnerGeneratedAliasByKey[key]; ok {
		return warehouse.Table{}, collision.err()
	}
	if baseName, ok := p.ambiguousBaseNameByKey[key]; ok {
		return resolver.Find(baseName, "")
	}
	if ambiguity, ok := p.caseEquivalentUniqueAmbiguityByKey[key]; ok && p.unscopedFlow {
		return warehouse.Table{}, ambiguity.err()
	}
	if baseName, ok := p.generatedAliasBaseName[key]; ok {
		return resolver.Find(baseName, "")
	}
	return resolver.Find(tableName, connection)
}

func registerWarehouseView(ctx context.Context, db *sql.DB, view string, table warehouse.Table) error {
	info, err := os.Stat(table.Path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", table.Path, err)
	}
	if info.Size() == 0 {
		return nil
	}
	stmt := fmt.Sprintf(
		"CREATE VIEW %s AS SELECT * FROM read_parquet('%s')",
		quoteDuckDBIdentifier(view), warehouse.EscapeSQLLiteral(table.Path),
	)
	if _, err := db.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("register view %q: %w", view, err)
	}
	return nil
}

func quoteDuckDBIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

// querySelectsConnection mirrors the warehouse selector contract while
// registering the views available to one analytical query. The authoritative
// ownership lookup remains warehouse.FindTable; this filter only prevents a
// selected query from registering another connection's files in DuckDB.
func querySelectsConnection(requested, owner string) bool {
	switch requested {
	case "":
		return true
	case warehouse.UnattributedConnection:
		return owner == ""
	default:
		return owner == requested
	}
}

// hasTopLevelLimit reports whether the query already contains a LIMIT clause at
// the top level (outside any parentheses).
func hasTopLevelLimit(query string) bool {
	depth := 0
	lower := strings.ToLower(query)
	for i := 0; i < len(lower); i++ {
		switch lower[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		}
		if depth == 0 && strings.HasPrefix(lower[i:], "limit") {
			before := i == 0 || !isWordByte(lower[i-1])
			after := i+5 >= len(lower) || !isWordByte(lower[i+5])
			if before && after {
				return true
			}
		}
	}
	return false
}

func isWordByte(b byte) bool {
	return b == '_' ||
		(b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9')
}
