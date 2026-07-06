package postgres

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"polymetrics.ai/internal/connectors"
)

// Catalog discovers tables and columns. Fixture mode returns canned streams
// (fixtureStreams) so unit tests and the conformance harness never need a
// live database.
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	conn, err := resolveConfig(cfg)
	if err != nil {
		return connectors.Catalog{}, err
	}
	if fixtureMode(cfg) {
		return connectors.Catalog{Connector: c.Name(), Streams: fixtureStreams()}, nil
	}

	pool, err := pgxpool.New(ctx, conn.dsn())
	if err != nil {
		return connectors.Catalog{}, fmt.Errorf("catalog postgres: open pool: %w", err)
	}
	defer pool.Close()

	streams, err := discoverStreams(ctx, pool, conn.schema)
	if err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams}, nil
}

// discoverStreams reads information_schema.tables/columns for the target
// schema and builds one Stream per base table, with Fields typed from each
// column's data_type and a primary key derived from table_constraints when
// present.
func discoverStreams(ctx context.Context, pool *pgxpool.Pool, schema string) ([]connectors.Stream, error) {
	const colsSQL = `
SELECT c.table_name, c.column_name, c.data_type, c.ordinal_position
FROM information_schema.columns c
JOIN information_schema.tables t
  ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE c.table_schema = $1 AND t.table_type = 'BASE TABLE'
ORDER BY c.table_name, c.ordinal_position`

	rows, err := pool.Query(ctx, colsSQL, schema)
	if err != nil {
		return nil, fmt.Errorf("catalog postgres: query columns: %w", err)
	}
	defer rows.Close()

	byTable := map[string][]connectors.Field{}
	var order []string
	for rows.Next() {
		var table, column, dataType string
		var pos int
		if err := rows.Scan(&table, &column, &dataType, &pos); err != nil {
			return nil, fmt.Errorf("catalog postgres: scan column: %w", err)
		}
		if _, seen := byTable[table]; !seen {
			order = append(order, table)
		}
		byTable[table] = append(byTable[table], connectors.Field{Name: column, Type: pgTypeToFieldType(dataType)})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog postgres: iterate columns: %w", err)
	}

	pks, err := discoverPrimaryKeys(ctx, pool, schema)
	if err != nil {
		return nil, err
	}

	streams := make([]connectors.Stream, 0, len(order))
	for _, table := range order {
		qualified := schema + "." + table
		streams = append(streams, connectors.Stream{
			Name:        qualified,
			Description: "PostgreSQL table " + qualified,
			Fields:      byTable[table],
			PrimaryKey:  pks[table],
		})
	}
	sort.Slice(streams, func(i, j int) bool { return streams[i].Name < streams[j].Name })
	return streams, nil
}

// discoverPrimaryKeys returns table_name -> ordered primary key columns for
// the schema, used to populate Stream.PrimaryKey during discovery.
func discoverPrimaryKeys(ctx context.Context, pool *pgxpool.Pool, schema string) (map[string][]string, error) {
	const pkSQL = `
SELECT tc.table_name, kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name
 AND tc.table_schema = kcu.table_schema
WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_schema = $1
ORDER BY tc.table_name, kcu.ordinal_position`

	rows, err := pool.Query(ctx, pkSQL, schema)
	if err != nil {
		return nil, fmt.Errorf("catalog postgres: query primary keys: %w", err)
	}
	defer rows.Close()
	out := map[string][]string{}
	for rows.Next() {
		var table, column string
		if err := rows.Scan(&table, &column); err != nil {
			return nil, fmt.Errorf("catalog postgres: scan primary key: %w", err)
		}
		out[table] = append(out[table], column)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("catalog postgres: iterate primary keys: %w", err)
	}
	return out, nil
}

// pgTypeToFieldType maps a PostgreSQL information_schema.columns.data_type
// value to the connector's coarse field type vocabulary. Unknown types fall
// back to "string" so discovery never fails on an exotic type.
func pgTypeToFieldType(dataType string) string {
	switch dataType {
	case "smallint", "integer", "bigint", "smallserial", "serial", "bigserial":
		return "integer"
	case "numeric", "decimal", "real", "double precision", "money":
		return "number"
	case "boolean":
		return "boolean"
	case "timestamp without time zone", "timestamp with time zone", "date", "time without time zone", "time with time zone":
		return "timestamp"
	case "json", "jsonb":
		return "object"
	case "ARRAY":
		return "array"
	default:
		// character varying, text, uuid, bytea, inet, etc.
		return "string"
	}
}

// fixtureStreams is the canned catalog returned in fixture mode so the
// conformance harness and unit tests can run without a live database. It
// mirrors the shape a real information_schema discovery would produce
// (ported verbatim from legacy internal/connectors/postgres/streams.go).
func fixtureStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "public.users",
			Description:  "Fixture users table (mode=fixture canned stream).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields: []connectors.Field{
				{Name: "id", Type: "integer"},
				{Name: "email", Type: "string"},
				{Name: "full_name", Type: "string"},
				{Name: "is_active", Type: "boolean"},
				{Name: "updated_at", Type: "timestamp"},
			},
		},
		{
			Name:         "public.orders",
			Description:  "Fixture orders table (mode=fixture canned stream).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields: []connectors.Field{
				{Name: "id", Type: "integer"},
				{Name: "user_id", Type: "integer"},
				{Name: "amount_cents", Type: "integer"},
				{Name: "status", Type: "string"},
				{Name: "updated_at", Type: "timestamp"},
			},
		},
	}
}

// fixtureRow pairs a canned record with a numeric cursor value used to
// emulate incremental cursor filtering deterministically in fixture mode.
type fixtureRow struct {
	cursor int64
	record connectors.Record
}

// fixtureRows returns deterministic canned rows for a fixture stream. The
// cursor column (updated_at here, surfaced as a numeric string for ordering
// in tests) is monotonically increasing so the incremental lower bound can
// filter them. Returns (rows, true) when the stream is known.
func fixtureRows(stream string) ([]fixtureRow, bool) {
	switch stream {
	case "public.users":
		return []fixtureRow{
			{cursor: 1000, record: connectors.Record{"id": 1, "email": "ada@example.com", "full_name": "Ada Lovelace", "is_active": true, "updated_at": "1000"}},
			{cursor: 2000, record: connectors.Record{"id": 2, "email": "grace@example.com", "full_name": "Grace Hopper", "is_active": true, "updated_at": "2000"}},
			{cursor: 3000, record: connectors.Record{"id": 3, "email": "katherine@example.com", "full_name": "Katherine Johnson", "is_active": false, "updated_at": "3000"}},
		}, true
	case "public.orders":
		return []fixtureRow{
			{cursor: 1500, record: connectors.Record{"id": 10, "user_id": 1, "amount_cents": 4999, "status": "paid", "updated_at": "1500"}},
			{cursor: 2500, record: connectors.Record{"id": 11, "user_id": 2, "amount_cents": 12000, "status": "pending", "updated_at": "2500"}},
		}, true
	default:
		return nil, false
	}
}
