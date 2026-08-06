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

// InitialState returns the conventional opaque cursor state. The connector
// writes no state itself; the caller advances it only after downstream work.
func (c Connector) InitialState(ctx context.Context, stream string, _ connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Read runs a bounded snapshot or incremental read. A configured cursor field
// makes paging deterministic: each page continues strictly after its last
// emitted cursor, so the Docker/Colima proof can exercise more than one real query.
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

	last := connsdk.Cursor(req.State)
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

		query, args := snapshotQuery(database, table, cursorField, last, remaining)
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
		if len(records) < remaining || cursorField == "" {
			return nil
		}
		cursor, ok := records[len(records)-1][cursorField]
		if !ok || cursor == nil {
			return fmt.Errorf("mysql read cursor_field %q missing from result", cursorField)
		}
		last = recordCursor(cursor)
		if last == "" {
			return fmt.Errorf("mysql read cursor_field %q produced an empty cursor", cursorField)
		}
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

func snapshotQuery(database, table, cursorField, lowerBound string, limit int) (string, []any) {
	query := "SELECT * FROM " + quoteIdentifier(database) + "." + quoteIdentifier(table)
	if cursorField != "" {
		query += " WHERE " + quoteIdentifier(cursorField) + " > ? ORDER BY " + quoteIdentifier(cursorField) + " ASC"
		return query + " LIMIT " + strconv.Itoa(limit), []any{lowerBound}
	}
	return query + " LIMIT " + strconv.Itoa(limit), nil
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
			record[string(field.Name)] = row[idx].Value()
		}
		records = append(records, record)
	}
	return records, nil
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
