package mysql

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	gomysql "github.com/go-mysql-org/go-mysql/mysql"

	"polymetrics.ai/internal/connectors"
)

type mysqlExecutor interface {
	Execute(string, ...any) (*gomysql.Result, error)
}

// Catalog discovers the configured database's base tables, columns, and
// primary-key column order from information_schema.
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	conn, err := resolveConfig(cfg)
	if err != nil {
		return connectors.Catalog{}, err
	}
	db, err := conn.open(ctx)
	if err != nil {
		return connectors.Catalog{}, err
	}
	defer func() { _ = db.Close() }()

	result, err := db.Execute(`
SELECT c.table_name AS table_name, c.column_name AS column_name, c.data_type AS data_type, c.ordinal_position AS ordinal_position
FROM information_schema.columns c
JOIN information_schema.tables t
  ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE c.table_schema = ? AND t.table_type = 'BASE TABLE'
ORDER BY c.table_name, c.ordinal_position`, conn.database)
	if err != nil {
		return connectors.Catalog{}, fmt.Errorf("catalog mysql columns: %w", err)
	}
	columns, err := resultRecords(result)
	result.Close()
	if err != nil {
		return connectors.Catalog{}, err
	}

	pks, err := discoverPrimaryKeys(ctx, db, conn.database)
	if err != nil {
		return connectors.Catalog{}, err
	}
	cursorField := strings.TrimSpace(cfg.Config["cursor_field"])
	streams, err := catalogStreams(conn.database, columns, pks, cursorField)
	if err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: streams}, nil
}

func discoverPrimaryKeys(ctx context.Context, db mysqlExecutor, database string) (map[string][]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	result, err := db.Execute(`
SELECT k.table_name AS table_name, k.column_name AS column_name
FROM information_schema.table_constraints t
JOIN information_schema.key_column_usage k
  ON t.constraint_name = k.constraint_name
 AND t.table_schema = k.table_schema
 AND t.table_name = k.table_name
WHERE t.table_schema = ? AND t.constraint_type = 'PRIMARY KEY'
ORDER BY k.table_name, k.ordinal_position`, database)
	if err != nil {
		return nil, fmt.Errorf("catalog mysql primary keys: %w", err)
	}
	records, err := resultRecords(result)
	result.Close()
	if err != nil {
		return nil, err
	}
	pks := make(map[string][]string)
	for _, record := range records {
		table, tableOK := recordString(record["table_name"])
		column, columnOK := recordString(record["column_name"])
		if !tableOK || !columnOK {
			return nil, errors.New("catalog mysql primary keys returned invalid identifiers")
		}
		pks[table] = append(pks[table], column)
	}
	return pks, nil
}

func catalogStreams(database string, columns []connectors.Record, pks map[string][]string, cursorField string) ([]connectors.Stream, error) {
	byTable := make(map[string][]connectors.Field)
	cursorByTable := make(map[string]bool)
	for _, record := range columns {
		table, tableOK := recordString(record["table_name"])
		column, columnOK := recordString(record["column_name"])
		dataType, typeOK := recordString(record["data_type"])
		if !tableOK || !columnOK || !typeOK {
			return nil, errors.New("catalog mysql columns returned invalid metadata")
		}
		byTable[table] = append(byTable[table], connectors.Field{Name: column, Type: mysqlTypeToFieldType(dataType)})
		if column == cursorField {
			cursorByTable[table] = true
		}
	}
	streams := make([]connectors.Stream, 0, len(byTable))
	for table, fields := range byTable {
		stream := connectors.Stream{
			Name:        database + "." + table,
			Description: "MySQL table " + database + "." + table,
			Fields:      fields,
			PrimaryKey:  pks[table],
		}
		if cursorByTable[table] {
			stream.CursorFields = []string{cursorField}
		}
		streams = append(streams, stream)
	}
	sort.Slice(streams, func(i, j int) bool { return streams[i].Name < streams[j].Name })
	return streams, nil
}

func recordString(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case []byte:
		return string(typed), true
	default:
		return "", false
	}
}

func mysqlTypeToFieldType(dataType string) string {
	switch strings.ToLower(dataType) {
	case "tinyint", "smallint", "mediumint", "int", "integer", "bigint", "year":
		return "integer"
	case "decimal", "numeric", "float", "double", "real":
		return "number"
	case "bool", "boolean", "bit":
		return "boolean"
	case "date", "time", "datetime", "timestamp":
		return "timestamp"
	case "json":
		return "object"
	default:
		return "string"
	}
}
