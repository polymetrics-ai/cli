package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// InitialState satisfies connectors.StatefulReader. A stream starts with an
// empty incremental cursor (full snapshot); subsequent reads advance the
// cursor stored under the conventional connsdk cursor key.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Read performs a snapshot SELECT over the stream's table, optionally
// filtered to rows whose cursor column is greater than the incremental
// lower bound carried in req.State (or the start cursor in config). Fixture
// mode emits canned rows.
func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn, err := resolveConfig(req.Config)
	if err != nil {
		return err
	}
	if req.Stream == "" {
		return errors.New("postgres read requires a stream (schema.table)")
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, req, emit)
	}

	cursorColumn := strings.TrimSpace(req.Config.Config["cursor_field"])
	lowerBound := connsdk.Cursor(req.State)
	limit, err := readLimit(req.Config)
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, conn.dsn())
	if err != nil {
		return fmt.Errorf("read postgres: open pool: %w", err)
	}
	defer pool.Close()

	return c.snapshot(ctx, pool, conn.schema, req.Stream, cursorColumn, lowerBound, limit, emit)
}

// snapshot builds and runs the SELECT for a stream. The table and cursor
// column are validated as plain identifiers and quoted, so they are never
// concatenated raw; the lower bound is passed as a bound parameter.
func (c Connector) snapshot(ctx context.Context, pool *pgxpool.Pool, schema, stream, cursorColumn, lowerBound string, limit int, emit func(connectors.Record) error) error {
	table, err := qualifyStream(schema, stream)
	if err != nil {
		return err
	}

	sql := "SELECT * FROM " + table
	var args []any
	if cursorColumn != "" {
		if err := validateIdentifier(cursorColumn); err != nil {
			return fmt.Errorf("read postgres: cursor_field: %w", err)
		}
		if lowerBound != "" {
			sql += " WHERE " + quoteIdentifier(cursorColumn) + " > $1"
			args = append(args, lowerBound)
		}
		sql += " ORDER BY " + quoteIdentifier(cursorColumn) + " ASC"
	}
	if limit > 0 {
		sql += " LIMIT " + strconv.Itoa(limit)
	}

	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("read postgres %s: %w", stream, err)
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	for rows.Next() {
		if err := ctx.Err(); err != nil {
			return err
		}
		values, err := rows.Values()
		if err != nil {
			return fmt.Errorf("read postgres %s: scan row: %w", stream, err)
		}
		record := make(connectors.Record, len(fieldDescs))
		for i, fd := range fieldDescs {
			record[string(fd.Name)] = values[i]
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read postgres %s: iterate: %w", stream, err)
	}
	return nil
}

// readFixture emits canned rows for a fixture stream without any network
// access. It applies the incremental cursor lower bound numerically so
// cursor behaviour is exercised credential-free.
func (c Connector) readFixture(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	rows, ok := fixtureRows(req.Stream)
	if !ok {
		return fmt.Errorf("postgres fixture stream %q not found", req.Stream)
	}
	var lower int64 = -1
	if cur := connsdk.Cursor(req.State); cur != "" {
		if n, err := strconv.ParseInt(cur, 10, 64); err == nil {
			lower = n
		}
	}
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if lower >= 0 && row.cursor <= lower {
			continue
		}
		if err := emit(copyRecord(row.record)); err != nil {
			return err
		}
	}
	return nil
}

// qualifyStream turns a stream name (either "table" or "schema.table") into
// a quoted, validated qualified identifier safe to embed in SQL.
func qualifyStream(defaultSchemaName, stream string) (string, error) {
	stream = strings.TrimSpace(stream)
	schemaName := defaultSchemaName
	table := stream
	if idx := strings.IndexByte(stream, '.'); idx >= 0 {
		schemaName = stream[:idx]
		table = stream[idx+1:]
	}
	if err := validateIdentifier(schemaName); err != nil {
		return "", fmt.Errorf("read postgres: schema: %w", err)
	}
	if err := validateIdentifier(table); err != nil {
		return "", fmt.Errorf("read postgres: table: %w", err)
	}
	return quoteIdentifier(schemaName) + "." + quoteIdentifier(table), nil
}

// copyRecord returns a shallow copy of a record so emitted fixture rows are
// not aliased between calls.
func copyRecord(in connectors.Record) connectors.Record {
	out := make(connectors.Record, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
