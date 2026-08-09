package mysql

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	gomysql "github.com/go-mysql-org/go-mysql/mysql"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const mysqlBinaryCharsetID = 63

const (
	mysqlCursorStateString byte = iota + 1
	mysqlCursorStateBytes
	mysqlCursorStateInt64
	mysqlCursorStateUint64
	mysqlCursorStateFloat64
	mysqlCursorStateBool
)

var mysqlCursorStatePrefix = []byte{0x00, 'p', 'm', ':', 'm', 'y', 's', 'q', 'l', ':', 'c', 'u', 'r', 's', 'o', 'r', ':', 0x01}

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

func (c Connector) ValidateCursorField(config connectors.RuntimeConfig, field string) error {
	field = strings.TrimSpace(field)
	configured := strings.TrimSpace(config.Config["cursor_field"])
	if field == "" || configured == "" || field != configured {
		return errors.New("mysql stream cursor field must match configured cursor_field")
	}
	if err := validateIdentifier(field); err != nil {
		return fmt.Errorf("mysql stream cursor field: %w", err)
	}
	return nil
}

func (c Connector) CursorStateFromRecord(record connectors.Record, field string) (connectors.OpaqueCursorState, error) {
	field = strings.TrimSpace(field)
	if field == "" {
		return connectors.OpaqueCursorState{}, errors.New("mysql cursor field is required")
	}
	value, ok := record[field]
	if !ok || value == nil {
		return connectors.OpaqueCursorState{}, fmt.Errorf("record is missing cursor field %q", field)
	}
	token, err := encodeMySQLCursorState(value)
	if err != nil {
		return connectors.OpaqueCursorState{}, err
	}
	return connectors.OpaqueCursorState{Token: token, Present: true}, nil
}

func (c Connector) CompareCursorStates(left, right connectors.OpaqueCursorState) (int, error) {
	leftValue, err := mysqlOpaqueCursorValue(left)
	if err != nil {
		return 0, err
	}
	rightValue, err := mysqlOpaqueCursorValue(right)
	if err != nil {
		return 0, err
	}
	switch left := leftValue.(type) {
	case []byte:
		right, ok := rightValue.([]byte)
		if !ok {
			return 0, errors.New("mysql cursor state types do not match")
		}
		return bytes.Compare(left, right), nil
	case int64:
		right, ok := rightValue.(int64)
		if !ok {
			return 0, errors.New("mysql cursor state types do not match")
		}
		switch {
		case left < right:
			return -1, nil
		case left > right:
			return 1, nil
		default:
			return 0, nil
		}
	case uint64:
		right, ok := rightValue.(uint64)
		if !ok {
			return 0, errors.New("mysql cursor state types do not match")
		}
		switch {
		case left < right:
			return -1, nil
		case left > right:
			return 1, nil
		default:
			return 0, nil
		}
	case float64:
		right, ok := rightValue.(float64)
		if !ok {
			return 0, errors.New("mysql cursor state types do not match")
		}
		if math.IsNaN(left) || math.IsNaN(right) {
			return 0, errors.New("mysql floating cursor state is not ordered")
		}
		switch {
		case left < right:
			return -1, nil
		case left > right:
			return 1, nil
		default:
			return 0, nil
		}
	case bool:
		right, ok := rightValue.(bool)
		if !ok {
			return 0, errors.New("mysql cursor state types do not match")
		}
		switch {
		case !left && right:
			return -1, nil
		case left && !right:
			return 1, nil
		default:
			return 0, nil
		}
	case string:
		if _, ok := rightValue.(string); !ok {
			return 0, errors.New("mysql cursor state types do not match")
		}
		return 0, connectors.ErrOpaqueCursorOrderUnavailable
	default:
		return 0, fmt.Errorf("mysql cursor state has unsupported type %T", leftValue)
	}
}

func encodeMySQLCursorState(value any) ([]byte, error) {
	token := append([]byte(nil), mysqlCursorStatePrefix...)
	appendUint64 := func(kind byte, value uint64) []byte {
		token = append(token, kind)
		encoded := make([]byte, 8)
		binary.BigEndian.PutUint64(encoded, value)
		return append(token, encoded...)
	}
	switch value := value.(type) {
	case string:
		return append(append(token, mysqlCursorStateString), value...), nil
	case []byte:
		return append(append(token, mysqlCursorStateBytes), value...), nil
	case int:
		return appendUint64(mysqlCursorStateInt64, uint64(value)), nil
	case int8:
		return appendUint64(mysqlCursorStateInt64, uint64(value)), nil
	case int16:
		return appendUint64(mysqlCursorStateInt64, uint64(value)), nil
	case int32:
		return appendUint64(mysqlCursorStateInt64, uint64(value)), nil
	case int64:
		return appendUint64(mysqlCursorStateInt64, uint64(value)), nil
	case uint:
		return appendUint64(mysqlCursorStateUint64, uint64(value)), nil
	case uint8:
		return appendUint64(mysqlCursorStateUint64, uint64(value)), nil
	case uint16:
		return appendUint64(mysqlCursorStateUint64, uint64(value)), nil
	case uint32:
		return appendUint64(mysqlCursorStateUint64, uint64(value)), nil
	case uint64:
		return appendUint64(mysqlCursorStateUint64, value), nil
	case float32:
		return appendUint64(mysqlCursorStateFloat64, math.Float64bits(float64(value))), nil
	case float64:
		return appendUint64(mysqlCursorStateFloat64, math.Float64bits(value)), nil
	case bool:
		if value {
			return append(token, mysqlCursorStateBool, 1), nil
		}
		return append(token, mysqlCursorStateBool, 0), nil
	default:
		return nil, fmt.Errorf("mysql cursor value has unsupported type %T", value)
	}
}

func readCursorState(req connectors.ReadRequest) (any, bool, error) {
	value, encoded, err := decodeMySQLCursorState(req.CursorState)
	if err != nil {
		return nil, false, err
	}
	if encoded {
		return value, true, nil
	}
	legacy, present := rawCursorState(req.State)
	if present {
		return legacy, true, nil
	}
	if req.CursorState.Present {
		return nil, false, errors.New("mysql cursor state is unrecognized")
	}
	return nil, false, nil
}

func decodeMySQLCursorState(state connectors.OpaqueCursorState) (any, bool, error) {
	if !state.Present || !bytes.HasPrefix(state.Token, mysqlCursorStatePrefix) {
		return nil, false, nil
	}
	if len(state.Token) == len(mysqlCursorStatePrefix) {
		return nil, true, errors.New("mysql cursor state is missing its type")
	}
	kind := state.Token[len(mysqlCursorStatePrefix)]
	payload := state.Token[len(mysqlCursorStatePrefix)+1:]
	switch kind {
	case mysqlCursorStateString:
		return string(payload), true, nil
	case mysqlCursorStateBytes:
		return append([]byte(nil), payload...), true, nil
	case mysqlCursorStateInt64:
		if len(payload) != 8 {
			return nil, true, errors.New("mysql signed cursor state is malformed")
		}
		return int64(binary.BigEndian.Uint64(payload)), true, nil
	case mysqlCursorStateUint64:
		if len(payload) != 8 {
			return nil, true, errors.New("mysql unsigned cursor state is malformed")
		}
		return binary.BigEndian.Uint64(payload), true, nil
	case mysqlCursorStateFloat64:
		if len(payload) != 8 {
			return nil, true, errors.New("mysql floating cursor state is malformed")
		}
		return math.Float64frombits(binary.BigEndian.Uint64(payload)), true, nil
	case mysqlCursorStateBool:
		if len(payload) != 1 || payload[0] > 1 {
			return nil, true, errors.New("mysql boolean cursor state is malformed")
		}
		return payload[0] == 1, true, nil
	default:
		return nil, true, errors.New("mysql cursor state type is unsupported")
	}
}

func mysqlOpaqueCursorValue(state connectors.OpaqueCursorState) (any, error) {
	value, encoded, err := decodeMySQLCursorState(state)
	if err != nil {
		return nil, err
	}
	if !encoded {
		return nil, errors.New("mysql cursor state is unrecognized")
	}
	return value, nil
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

	initialCursor, hasCursorState, err := readCursorState(req)
	if err != nil {
		return err
	}
	lastCursor := initialCursor
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
