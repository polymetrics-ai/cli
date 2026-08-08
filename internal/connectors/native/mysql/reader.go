package mysql

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	gomysql "github.com/go-mysql-org/go-mysql/mysql"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const mysqlBinaryCharsetID = 63

// InitialState returns the stream state without a cursor. The connector writes
// no state itself; the caller advances it only after downstream work.
func (c Connector) InitialState(ctx context.Context, stream string, _ connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return map[string]string{"stream": stream}, nil
}

func rawCursorState(state map[string]string) (string, bool) {
	if state == nil {
		return "", false
	}
	value, present := state[connsdk.CursorStateKey]
	return value, present
}

// Read runs a bounded snapshot or incremental read. Snapshot pages use a
// single-column primary key, while cursor pages use the cursor and primary key
// together so each page has a complete deterministic order.
func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(req.Stream) == "" {
		return errors.New("mysql read requires a stream (table or database.table)")
	}
	conn, err := resolveConfig(req.Config)
	if err != nil {
		return err
	}
	database, table, err := qualifyStream(conn.database, req.Stream)
	if err != nil {
		return err
	}
	cursorField := strings.TrimSpace(req.Config.Config["cursor_field"])
	if cursorField != "" {
		if err := validateIdentifier(cursorField); err != nil {
			return fmt.Errorf("mysql read cursor_field: %w", err)
		}
	}
	limit, err := readLimit(req.Config)
	if err != nil {
		return err
	}
	if req.Limit > 0 && (limit == 0 || req.Limit < limit) {
		limit = req.Limit
	}
	pageSize, err := pageSize(req.Config, limit)
	if err != nil {
		return err
	}

	db, err := conn.open(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	primaryKeys, err := discoverPrimaryKeys(ctx, db, database, table)
	if err != nil {
		return err
	}
	primaryKey, err := singleColumnPrimaryKey(primaryKeys[table])
	if err != nil {
		return err
	}
	if cursorField != "" {
		if err := requireUniqueNonNullableCursorField(ctx, db, database, table, cursorField); err != nil {
			return err
		}
	}

	initialCursor, hasCursorState := rawCursorState(req.State)
	lastCursor := any(initialCursor)
	var lastPrimaryKey any
	resume := cursorField != "" && hasCursorState
	hasPageBoundary := false
	emitted := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		remaining := pageSize
		if limit > 0 && limit-emitted < remaining {
			remaining = limit - emitted
		}
		if remaining <= 0 {
			return nil
		}

		query, args := snapshotQuery(database, table, cursorField, primaryKey, lastCursor, lastPrimaryKey, resume, hasPageBoundary, remaining)
		result, err := db.Execute(query, args...)
		if err != nil {
			return fmt.Errorf("read mysql table: %w", err)
		}
		records, err := resultRecords(result)
		result.Close()
		if err != nil {
			return err
		}
		for _, record := range records {
			if err := emit(record); err != nil {
				return err
			}
			emitted++
		}
		if len(records) < remaining {
			return nil
		}
		lastRecord := records[len(records)-1]
		if cursorField != "" {
			cursor, ok := lastRecord[cursorField]
			if !ok || cursor == nil {
				return errors.New("mysql read cursor_field is missing from a result")
			}
			lastCursor = copyReadBoundaryValue(cursor)
		}
		value, ok := lastRecord[primaryKey]
		if !ok || value == nil {
			return errors.New("mysql read primary key is missing from a result")
		}
		lastPrimaryKey = copyReadBoundaryValue(value)
		hasPageBoundary = true
	}
}

func pageSize(cfg connectors.RuntimeConfig, limit int) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		if limit > 0 && limit < 1000 {
			return limit, nil
		}
		return 1000, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, errors.New("mysql config page_size must be a positive integer")
	}
	return value, nil
}

func singleColumnPrimaryKey(primaryKey []string) (string, error) {
	if len(primaryKey) != 1 || validateIdentifier(primaryKey[0]) != nil {
		return "", errors.New("mysql read requires a single-column primary key for complete paging")
	}
	return primaryKey[0], nil
}

func requireUniqueNonNullableCursorField(ctx context.Context, db mysqlExecutor, database, table, cursorField string) error {
	result, err := db.Execute(`
SELECT s.column_name AS column_name
FROM information_schema.statistics s
JOIN information_schema.columns c
  ON c.table_schema = s.table_schema
 AND c.table_name = s.table_name
 AND c.column_name = s.column_name
WHERE s.table_schema = ?
  AND s.table_name = ?
  AND s.column_name = ?
  AND s.non_unique = 0
  AND s.seq_in_index = 1
  AND c.is_nullable = 'NO'
  AND NOT EXISTS (
    SELECT 1
    FROM information_schema.statistics later_part
    WHERE later_part.table_schema = s.table_schema
      AND later_part.table_name = s.table_name
      AND later_part.index_name = s.index_name
      AND later_part.seq_in_index > 1
  )
LIMIT 1`, database, table, cursorField)
	if err != nil {
		// See cataloger.go: TRAILING is a MySQL reserved word, and swallowing
		// the server's error hid that for a whole integration run.
		return fmt.Errorf("read mysql cursor metadata: %w", err)
	}
	records, err := resultRecords(result)
	result.Close()
	if err != nil {
		return err
	}
	return validateUniqueCursorFieldRecords(records, cursorField)
}

func validateUniqueCursorFieldRecords(records []connectors.Record, cursorField string) error {
	if len(records) != 1 {
		return errors.New("mysql read cursor_field must reference a non-null single-column primary or unique key")
	}
	name, ok := recordString(records[0]["column_name"])
	if !ok || name != cursorField {
		return errors.New("mysql read cursor metadata is invalid")
	}
	return nil
}

func snapshotQuery(database, table, cursorField, primaryKey string, lowerCursor, lowerPrimaryKey any, resume, hasPageBoundary bool, limit int) (string, []any) {
	query := "SELECT * FROM " + quoteIdentifier(database) + "." + quoteIdentifier(table)
	if cursorField == "" {
		if hasPageBoundary {
			query += " WHERE " + quoteIdentifier(primaryKey) + " > ?"
			return query + " ORDER BY " + quoteIdentifier(primaryKey) + " ASC LIMIT " + strconv.Itoa(limit), []any{lowerPrimaryKey}
		}
		return query + " ORDER BY " + quoteIdentifier(primaryKey) + " ASC LIMIT " + strconv.Itoa(limit), nil
	}
	if hasPageBoundary {
		query += " WHERE (" + quoteIdentifier(cursorField) + " > ? OR (" + quoteIdentifier(cursorField) + " = ? AND " + quoteIdentifier(primaryKey) + " > ?))"
		query += " ORDER BY " + quoteIdentifier(cursorField) + " ASC, " + quoteIdentifier(primaryKey) + " ASC"
		return query + " LIMIT " + strconv.Itoa(limit), []any{lowerCursor, lowerCursor, lowerPrimaryKey}
	}
	if resume {
		query += " WHERE " + quoteIdentifier(cursorField) + " > ?"
		query += " ORDER BY " + quoteIdentifier(cursorField) + " ASC, " + quoteIdentifier(primaryKey) + " ASC"
		return query + " LIMIT " + strconv.Itoa(limit), []any{lowerCursor}
	}
	return query + " ORDER BY " + quoteIdentifier(cursorField) + " ASC, " + quoteIdentifier(primaryKey) + " ASC LIMIT " + strconv.Itoa(limit), nil
}

func copyReadBoundaryValue(value any) any {
	if bytes, ok := value.([]byte); ok {
		return append([]byte(nil), bytes...)
	}
	return value
}

func resultRecords(result *gomysql.Result) ([]connectors.Record, error) {
	if result == nil || result.Resultset == nil {
		return nil, errors.New("mysql read returned no result set")
	}
	records := make([]connectors.Record, 0, len(result.Values))
	for _, row := range result.Values {
		if len(row) != len(result.Fields) {
			return nil, errors.New("mysql read returned malformed row")
		}
		record := make(connectors.Record, len(row))
		for idx, field := range result.Fields {
			if field == nil || len(field.Name) == 0 {
				return nil, errors.New("mysql read returned unnamed column")
			}
			record[string(field.Name)] = projectReadValue(field, row[idx].Value())
		}
		records = append(records, record)
	}
	return records, nil
}

func projectReadValue(field *gomysql.Field, value any) any {
	bytes, ok := value.([]byte)
	if !ok {
		return value
	}
	if isTextReadField(field) {
		return string(bytes)
	}
	return append([]byte(nil), bytes...)
}

func isTextReadField(field *gomysql.Field) bool {
	if field == nil {
		return false
	}
	switch field.Type {
	case gomysql.MYSQL_TYPE_JSON,
		gomysql.MYSQL_TYPE_DECIMAL,
		gomysql.MYSQL_TYPE_NEWDECIMAL,
		gomysql.MYSQL_TYPE_ENUM,
		gomysql.MYSQL_TYPE_SET,
		gomysql.MYSQL_TYPE_DATE,
		gomysql.MYSQL_TYPE_NEWDATE,
		gomysql.MYSQL_TYPE_TIME,
		gomysql.MYSQL_TYPE_TIME2,
		gomysql.MYSQL_TYPE_DATETIME,
		gomysql.MYSQL_TYPE_DATETIME2,
		gomysql.MYSQL_TYPE_TIMESTAMP,
		gomysql.MYSQL_TYPE_TIMESTAMP2:
		return true
	case gomysql.MYSQL_TYPE_VARCHAR,
		gomysql.MYSQL_TYPE_VAR_STRING,
		gomysql.MYSQL_TYPE_STRING,
		gomysql.MYSQL_TYPE_TINY_BLOB,
		gomysql.MYSQL_TYPE_MEDIUM_BLOB,
		gomysql.MYSQL_TYPE_LONG_BLOB,
		gomysql.MYSQL_TYPE_BLOB:
		return field.Flag&gomysql.BINARY_FLAG == 0 && field.Charset != mysqlBinaryCharsetID
	default:
		return false
	}
}

func recordCursor(value any) string {
	switch typed := value.(type) {
	case []byte:
		return string(typed)
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}
