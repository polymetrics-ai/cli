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

	_ "github.com/marcboeker/go-duckdb"

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
	warehouseDir string
}

func (e duckdbEngine) Name() string { return "duckdb" }

var (
	identRe = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	// Statement-leading keywords / tokens that must never appear in a query.
	forbiddenTokens = []string{
		"insert", "update", "delete", "drop", "alter", "create",
		"attach", "copy", "pragma", "call", "export", "install", "load", "set",
	}
	wordRe = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)
	// DuckDB reports an unregistered table by name. The engine uses this only
	// to recover warehouse's typed ambiguity error after intentionally leaving
	// duplicate unscoped views bare-less; it is not used to construct SQL.
	duckDBMissingTableRe = regexp.MustCompile(`(?i)table with name ["']?([A-Za-z_][A-Za-z0-9_]*)["']? does not exist`)
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

	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	defer db.Close()

	if err := e.registerViews(ctx, db, req.Connection); err != nil {
		return nil, err
	}

	finalSQL := req.SQL
	if req.Limit > 0 && !hasTopLevelLimit(req.SQL) {
		finalSQL = fmt.Sprintf("SELECT * FROM (%s) AS _pm_q LIMIT %d", req.SQL, req.Limit)
	}

	rows, err := db.QueryContext(ctx, finalSQL)
	if err != nil {
		if req.Connection == "" {
			if ambiguity := e.ambiguityForMissingTable(err); ambiguity != nil {
				return nil, fmt.Errorf("execute query: %w", ambiguity)
			}
		}
		return nil, fmt.Errorf("execute query: %w", err)
	}
	defer rows.Close()

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

// registerViews creates one DuckDB view per materialized warehouse table.
//
// Tables live inside their owning connection's directory, so the same table
// name can belong to several connections. A name owned by exactly one
// connection is registered bare; a name several connections share is
// registered once per connection as <table>__<connection-id> and never bare,
// so a query can never silently read one tenant's rows while meaning another's.
//
// View names are validated identifiers and file paths are passed as
// quote-escaped string literals — never via user SQL interpolation.
func (e duckdbEngine) registerViews(ctx context.Context, db *sql.DB, connection string) error {
	// A warehouse that does not exist yet has no tables and is not an error.
	// Faults are deliberately not fatal here: a damaged ownership record costs
	// its own connection's views, not every other connection's. A query naming
	// a view that is missing for that reason fails on its own, and the table
	// read path reports the fault precisely.
	tables, _, err := warehouse.Tables(e.warehouseDir)
	if err != nil {
		return err
	}
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
		owners := byName[name]
		for _, table := range owners {
			view := name
			if len(owners) > 1 {
				view = name + "__" + table.Owner.Connection
			}
			if !identRe.MatchString(view) {
				continue
			}
			info, err := os.Stat(table.Path)
			if err != nil {
				return fmt.Errorf("stat %s: %w", table.Path, err)
			}
			if info.Size() == 0 {
				continue
			}
			stmt := fmt.Sprintf(
				`CREATE VIEW "%s" AS SELECT * FROM read_parquet('%s')`,
				view, warehouse.EscapeSQLLiteral(table.Path),
			)
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("register view %q: %w", view, err)
			}
		}
	}
	return nil
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

// ambiguityForMissingTable restores warehouse's typed ambiguity result for a
// normal unscoped SQL query. Duplicate table names intentionally have no bare
// DuckDB view; when DuckDB reports that missing bare name, FindTable is the
// single authority that determines whether it was ambiguous rather than truly
// absent. Other query failures remain their original DuckDB errors.
func (e duckdbEngine) ambiguityForMissingTable(queryErr error) error {
	match := duckDBMissingTableRe.FindStringSubmatch(queryErr.Error())
	if len(match) != 2 {
		return nil
	}
	_, err := warehouse.FindTable(e.warehouseDir, match[1], "")
	var ambiguous *warehouse.AmbiguousTableError
	if errors.As(err, &ambiguous) {
		return err
	}
	return nil
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
