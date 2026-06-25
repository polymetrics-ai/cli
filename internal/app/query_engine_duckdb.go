//go:build duckdb

package app

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"polymetrics/internal/connectors"
)

// newSQLEngine returns the DuckDB-backed analytical engine (built only with
// -tags duckdb). It queries the JSONL warehouse files via in-memory views.
func newSQLEngine(a *App) sqlQueryEngine {
	return duckdbEngine{warehouseDir: filepath.Join(a.projectDir, "warehouse")}
}

// duckdbEngine runs read-only analytical SQL over the JSONL warehouse using an
// in-memory DuckDB instance. It is stateless: each query opens a fresh
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

func (e duckdbEngine) QuerySQL(ctx context.Context, query string, limit int) ([]connectors.Record, error) {
	if err := validateSelectOnly(query); err != nil {
		return nil, err
	}

	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("open duckdb: %w", err)
	}
	defer db.Close()

	if err := e.registerViews(ctx, db); err != nil {
		return nil, err
	}

	finalSQL := query
	if limit > 0 && !hasTopLevelLimit(query) {
		finalSQL = fmt.Sprintf("SELECT * FROM (%s) AS _pm_q LIMIT %d", query, limit)
	}

	rows, err := db.QueryContext(ctx, finalSQL)
	if err != nil {
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
			rec[col] = normalizeValue(holders[i])
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rows: %w", err)
	}
	return out, nil
}

// registerViews creates one DuckDB view per *.jsonl file in the warehouse dir,
// named after the file (sans extension). View names are validated identifiers
// and file paths are passed as quote-escaped string literals — never via user
// SQL interpolation.
func (e duckdbEngine) registerViews(ctx context.Context, db *sql.DB) error {
	entries, err := os.ReadDir(e.warehouseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read warehouse dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		table := strings.TrimSuffix(name, ".jsonl")
		if !identRe.MatchString(table) {
			continue
		}
		abs := filepath.Join(e.warehouseDir, name)
		info, err := os.Stat(abs)
		if err != nil {
			return fmt.Errorf("stat %s: %w", abs, err)
		}
		if info.Size() == 0 {
			continue
		}
		stmt := fmt.Sprintf(
			`CREATE VIEW "%s" AS SELECT * FROM read_ndjson_auto('%s')`,
			table, escapeSQLLiteral(abs),
		)
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("register view %q: %w", table, err)
		}
	}
	return nil
}

// escapeSQLLiteral escapes single quotes for safe inclusion inside a single
// quoted SQL string literal.
func escapeSQLLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
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

// normalizeValue coerces DuckDB/driver values into the same Go types the JSONL
// engine yields, so rows are interchangeable across engines. DuckDB returns
// wide integers (HUGEINT, e.g. from SUM) as *big.Int; collapse those to int64
// when they fit, else float64.
func normalizeValue(v any) any {
	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339)
	case *big.Int:
		if val == nil {
			return nil
		}
		if val.IsInt64() {
			return val.Int64()
		}
		f := new(big.Float).SetInt(val)
		out, _ := f.Float64()
		return out
	case big.Int:
		return normalizeValue(&val)
	case uint64:
		if val <= math.MaxInt64 {
			return int64(val)
		}
		return float64(val)
	default:
		return val
	}
}
