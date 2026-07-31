package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"polymetrics.ai/internal/connectors"
)

const postgresTruncateConfirmPhrase = "truncate"

var postgresWriteActions = map[string]struct{}{
	"insert_row":     {},
	"update_row":     {},
	"upsert_row":     {},
	"delete_row":     {},
	"truncate_table": {},
}

type writeStatement struct {
	SQL  string
	Args []any
}

type writeField struct {
	Name  string
	Value any
}

// ValidateWrite validates the five bounded PostgreSQL reverse-ETL actions.
// It deliberately accepts only closed row/table schemas: table/schema/column
// identifiers plus typed scalar values. Raw SQL text, statement fragments,
// cascade/restart-identity options, and arbitrary payload fields are rejected.
func (c Connector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn, err := resolveConfig(req.Config)
	if err != nil {
		return err
	}
	if _, ok := postgresWriteActions[req.Action]; !ok {
		return fmt.Errorf("unsupported postgres write action %q", req.Action)
	}
	if len(records) == 0 {
		return errors.New("postgres write requires at least one record")
	}
	for i, record := range records {
		if _, err := buildWriteStatement(conn.schema, req.Action, record); err != nil {
			return fmt.Errorf("postgres write record %d: %w", i, err)
		}
	}
	return nil
}

// DryRunWrite returns a redacted preview. SQL templates include only quoted
// identifiers and parameter placeholders; bound values are never rendered.
func (c Connector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WritePreview{}, err
	}
	conn, _ := resolveConfig(req.Config)
	warnings := make([]string, 0, len(records)+1)
	warnings = append(warnings, "postgres reverse ETL writes require plan -> preview -> explicit approval -> execute")
	for i, record := range records {
		stmt, err := buildWriteStatement(conn.schema, req.Action, record)
		if err != nil {
			return connectors.WritePreview{}, fmt.Errorf("postgres write record %d: %w", i, err)
		}
		warnings = append(warnings, fmt.Sprintf("record %d SQL template: %s", i, stmt.SQL))
	}
	return connectors.WritePreview{RecordsStaged: len(records), Action: req.Action, Warnings: warnings}, nil
}

// Write executes a validated bounded row-DML/table action. Fixture mode still
// performs full validation and SQL-template construction, then reports the
// deterministic staged-record count without opening a database connection.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if err := c.ValidateWrite(ctx, req, records); err != nil {
		return connectors.WriteResult{}, err
	}
	conn, _ := resolveConfig(req.Config)
	if fixtureMode(req.Config) {
		return connectors.WriteResult{RecordsWritten: len(records)}, nil
	}

	pool, err := pgxpool.New(ctx, conn.dsn())
	if err != nil {
		return connectors.WriteResult{}, fmt.Errorf("write postgres: open pool: %w", err)
	}
	defer pool.Close()

	result := connectors.WriteResult{}
	for i, record := range records {
		if err := ctx.Err(); err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, err
		}
		stmt, err := buildWriteStatement(conn.schema, req.Action, record)
		if err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, fmt.Errorf("postgres write record %d: %w", i, err)
		}
		if _, err := pool.Exec(ctx, stmt.SQL, stmt.Args...); err != nil {
			result.RecordsFailed = len(records) - result.RecordsWritten
			return result, fmt.Errorf("write postgres record %d action %s: %w", i, req.Action, err)
		}
		result.RecordsWritten++
	}
	return result, nil
}

func buildWriteStatement(defaultSchemaName, action string, record connectors.Record) (writeStatement, error) {
	relation, err := relationFromRecord(defaultSchemaName, record)
	if err != nil {
		return writeStatement{}, err
	}

	switch action {
	case "insert_row":
		values, err := valuesFromRecord(record, true)
		if err != nil {
			return writeStatement{}, err
		}
		cols := make([]string, 0, len(values))
		placeholders := make([]string, 0, len(values))
		args := make([]any, 0, len(values))
		for i, field := range values {
			cols = append(cols, quoteIdentifier(field.Name))
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
			args = append(args, field.Value)
		}
		return writeStatement{
			SQL:  fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", relation, strings.Join(cols, ", "), strings.Join(placeholders, ", ")),
			Args: args,
		}, nil

	case "update_row":
		values, err := valuesFromRecord(record, true)
		if err != nil {
			return writeStatement{}, err
		}
		keys, err := keysFromRecord(record, true)
		if err != nil {
			return writeStatement{}, err
		}
		setParts := make([]string, 0, len(values))
		whereParts := make([]string, 0, len(keys))
		args := make([]any, 0, len(values)+len(keys))
		for i, field := range values {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", quoteIdentifier(field.Name), i+1))
			args = append(args, field.Value)
		}
		for i, field := range keys {
			whereParts = append(whereParts, fmt.Sprintf("%s = $%d", quoteIdentifier(field.Name), len(args)+i+1))
			args = append(args, field.Value)
		}
		return writeStatement{
			SQL:  fmt.Sprintf("UPDATE %s SET %s WHERE %s", relation, strings.Join(setParts, ", "), strings.Join(whereParts, " AND ")),
			Args: args,
		}, nil

	case "upsert_row":
		keys, err := keysFromRecord(record, true)
		if err != nil {
			return writeStatement{}, err
		}
		values, err := upsertValuesFromRecord(record, keys)
		if err != nil {
			return writeStatement{}, err
		}
		valueByName := make(map[string]writeField, len(values))
		for _, field := range values {
			valueByName[field.Name] = field
		}
		for _, key := range keys {
			if _, ok := valueByName[key.Name]; !ok {
				return writeStatement{}, fmt.Errorf("upsert_row values must include key column %q", key.Name)
			}
		}
		cols := make([]string, 0, len(values))
		placeholders := make([]string, 0, len(values))
		insertValues := make([]string, 0, len(values))
		args := make([]any, 0, len(values))
		for i, field := range values {
			quoted := quoteIdentifier(field.Name)
			cols = append(cols, quoted)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
			insertValues = append(insertValues, "source."+quoted)
			args = append(args, field.Value)
		}
		keyNames := make(map[string]bool, len(keys))
		onParts := make([]string, 0, len(keys))
		for _, field := range keys {
			keyNames[field.Name] = true
			quoted := quoteIdentifier(field.Name)
			onParts = append(onParts, "target."+quoted+" = source."+quoted)
		}
		updates := make([]string, 0, len(values))
		for _, field := range values {
			if keyNames[field.Name] {
				continue
			}
			quoted := quoteIdentifier(field.Name)
			updates = append(updates, quoted+" = source."+quoted)
		}
		sql := fmt.Sprintf("MERGE INTO %s AS target USING (VALUES (%s)) AS source (%s) ON %s", relation, strings.Join(placeholders, ", "), strings.Join(cols, ", "), strings.Join(onParts, " AND "))
		if len(updates) == 0 {
			sql += " WHEN MATCHED THEN DO NOTHING"
		} else {
			sql += " WHEN MATCHED THEN UPDATE SET " + strings.Join(updates, ", ")
		}
		sql += fmt.Sprintf(" WHEN NOT MATCHED THEN INSERT (%s) VALUES (%s)", strings.Join(cols, ", "), strings.Join(insertValues, ", "))
		return writeStatement{SQL: sql, Args: args}, nil

	case "delete_row":
		keys, err := keysFromRecord(record, true)
		if err != nil {
			return writeStatement{}, err
		}
		whereParts := make([]string, 0, len(keys))
		args := make([]any, 0, len(keys))
		for i, field := range keys {
			whereParts = append(whereParts, fmt.Sprintf("%s = $%d", quoteIdentifier(field.Name), i+1))
			args = append(args, field.Value)
		}
		return writeStatement{
			SQL:  fmt.Sprintf("DELETE FROM %s WHERE %s", relation, strings.Join(whereParts, " AND ")),
			Args: args,
		}, nil

	case "truncate_table":
		phrase, ok := stringFromRecord(record, "confirm_phrase")
		if !ok || strings.TrimSpace(strings.ToLower(phrase)) != postgresTruncateConfirmPhrase {
			return writeStatement{}, fmt.Errorf("truncate_table requires confirm_phrase %q", postgresTruncateConfirmPhrase)
		}
		if values, err := valuesFromRecord(record, false); err != nil {
			return writeStatement{}, err
		} else if len(values) > 0 {
			return writeStatement{}, errors.New("truncate_table does not accept values")
		}
		if keys, err := keysFromRecord(record, false); err != nil {
			return writeStatement{}, err
		} else if len(keys) > 0 {
			return writeStatement{}, errors.New("truncate_table does not accept keys")
		}
		return writeStatement{SQL: fmt.Sprintf("TRUNCATE TABLE ONLY %s", relation)}, nil
	default:
		return writeStatement{}, fmt.Errorf("unsupported postgres write action %q", action)
	}
}

func relationFromRecord(defaultSchemaName string, record connectors.Record) (string, error) {
	allowed := map[string]bool{
		"schema": true, "table": true, "values": true, "keys": true, "confirm_phrase": true,
		"value_column": true, "value_string": true, "value_int": true, "value_float": true, "value_bool": true, "value_null": true, "value_json": true,
		"key_column": true, "key_string": true, "key_int": true, "key_float": true, "key_bool": true,
	}
	for key := range record {
		if !allowed[key] {
			return "", fmt.Errorf("unsupported field %q", key)
		}
	}
	schemaName := defaultSchemaName
	if raw, ok := stringFromRecord(record, "schema"); ok && strings.TrimSpace(raw) != "" {
		schemaName = strings.TrimSpace(raw)
	}
	table, ok := stringFromRecord(record, "table")
	if !ok || strings.TrimSpace(table) == "" {
		return "", errors.New("table is required")
	}
	if strings.Contains(table, ".") {
		return "", errors.New("table must be a plain identifier; pass schema separately")
	}
	if err := validateIdentifier(schemaName); err != nil {
		return "", fmt.Errorf("schema: %w", err)
	}
	if err := validateIdentifier(table); err != nil {
		return "", fmt.Errorf("table: %w", err)
	}
	return quoteIdentifier(schemaName) + "." + quoteIdentifier(table), nil
}

func valuesFromRecord(record connectors.Record, required bool) ([]writeField, error) {
	canonical, hasCanonical, err := fieldsFromRecordIfPresent(record, "values")
	if err != nil {
		return nil, err
	}
	shortcut, hasShortcut, err := shortcutField(record, "value", required && !hasCanonical)
	if err != nil {
		return nil, err
	}
	if hasCanonical && hasShortcut {
		return nil, errors.New("values array and value_* shortcut fields are mutually exclusive")
	}
	if hasCanonical {
		return canonical, nil
	}
	if hasShortcut {
		return []writeField{shortcut}, nil
	}
	if required {
		return nil, errors.New("values is required")
	}
	return nil, nil
}

func keysFromRecord(record connectors.Record, required bool) ([]writeField, error) {
	canonical, hasCanonical, err := fieldsFromRecordIfPresent(record, "keys")
	if err != nil {
		return nil, err
	}
	shortcut, hasShortcut, err := shortcutField(record, "key", required && !hasCanonical)
	if err != nil {
		return nil, err
	}
	if hasCanonical && hasShortcut {
		return nil, errors.New("keys array and key_* shortcut fields are mutually exclusive")
	}
	if hasCanonical {
		return canonical, nil
	}
	if hasShortcut {
		return []writeField{shortcut}, nil
	}
	if required {
		return nil, errors.New("keys is required")
	}
	return nil, nil
}

func upsertValuesFromRecord(record connectors.Record, keys []writeField) ([]writeField, error) {
	canonical, hasCanonical, err := fieldsFromRecordIfPresent(record, "values")
	if err != nil {
		return nil, err
	}
	shortcut, hasShortcut, err := shortcutField(record, "value", false)
	if err != nil {
		return nil, err
	}
	if hasCanonical && hasShortcut {
		return nil, errors.New("values array and value_* shortcut fields are mutually exclusive")
	}
	if hasCanonical {
		return canonical, nil
	}
	values := append([]writeField(nil), keys...)
	if hasShortcut {
		for _, key := range keys {
			if key.Name == shortcut.Name {
				return nil, fmt.Errorf("upsert_row shortcut value column %q duplicates key column", shortcut.Name)
			}
		}
		values = append(values, shortcut)
	}
	return values, nil
}

func fieldsFromRecordIfPresent(record connectors.Record, fieldName string) ([]writeField, bool, error) {
	if _, ok := record[fieldName]; !ok {
		return nil, false, nil
	}
	fields, err := fieldsFromRecord(record, fieldName, true)
	return fields, true, err
}

func shortcutField(record connectors.Record, prefix string, required bool) (writeField, bool, error) {
	columnKey := prefix + "_column"
	column, hasColumn := stringFromRecord(record, columnKey)
	typedKeys := []string{prefix + "_string", prefix + "_int", prefix + "_float", prefix + "_bool", prefix + "_null", prefix + "_json"}
	present := make([]string, 0, 1)
	for _, key := range typedKeys {
		if _, ok := record[key]; ok {
			present = append(present, key)
		}
	}
	if !hasColumn && len(present) == 0 {
		if required {
			return writeField{}, false, fmt.Errorf("%s or %s shortcut fields are required", strings.TrimSuffix(prefix, "_"), prefix+"s")
		}
		return writeField{}, false, nil
	}
	if !hasColumn || strings.TrimSpace(column) == "" {
		return writeField{}, false, fmt.Errorf("%s is required when %s shortcut fields are used", columnKey, prefix)
	}
	if len(present) != 1 {
		sort.Strings(present)
		return writeField{}, false, fmt.Errorf("exactly one %s typed value field is required, got %v", prefix, present)
	}
	m := map[string]any{"name": column}
	typedKey := present[0]
	m["value"+strings.TrimPrefix(typedKey, prefix)] = record[typedKey]
	field, err := parseWriteField(m)
	if err != nil {
		return writeField{}, false, err
	}
	return field, true, nil
}

func fieldsFromRecord(record connectors.Record, fieldName string, required bool) ([]writeField, error) {
	raw, ok := record[fieldName]
	if !ok || raw == nil {
		if required {
			return nil, fmt.Errorf("%s is required", fieldName)
		}
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array", fieldName)
	}
	if len(items) == 0 {
		if required {
			return nil, fmt.Errorf("%s must contain at least one item", fieldName)
		}
		return nil, nil
	}
	fields := make([]writeField, 0, len(items))
	seen := make(map[string]bool, len(items))
	for i, item := range items {
		m, err := mapFromAny(item)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", fieldName, i, err)
		}
		field, err := parseWriteField(m)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", fieldName, i, err)
		}
		if seen[field.Name] {
			return nil, fmt.Errorf("%s[%d]: duplicate column %q", fieldName, i, field.Name)
		}
		seen[field.Name] = true
		fields = append(fields, field)
	}
	return fields, nil
}

func parseWriteField(m map[string]any) (writeField, error) {
	allowed := map[string]bool{
		"name":         true,
		"value_string": true,
		"value_int":    true,
		"value_float":  true,
		"value_bool":   true,
		"value_null":   true,
		"value_json":   true,
	}
	for key := range m {
		if !allowed[key] {
			return writeField{}, fmt.Errorf("unsupported field %q", key)
		}
	}
	rawName, ok := m["name"].(string)
	if !ok || strings.TrimSpace(rawName) == "" {
		return writeField{}, errors.New("name is required")
	}
	name := strings.TrimSpace(rawName)
	if err := validateIdentifier(name); err != nil {
		return writeField{}, fmt.Errorf("name: %w", err)
	}
	valueKeys := make([]string, 0, 1)
	for _, key := range []string{"value_string", "value_int", "value_float", "value_bool", "value_null", "value_json"} {
		if _, ok := m[key]; ok {
			valueKeys = append(valueKeys, key)
		}
	}
	if len(valueKeys) != 1 {
		sort.Strings(valueKeys)
		return writeField{}, fmt.Errorf("exactly one typed value field is required, got %v", valueKeys)
	}
	value, err := coerceTypedValue(valueKeys[0], m[valueKeys[0]])
	if err != nil {
		return writeField{}, err
	}
	return writeField{Name: name, Value: value}, nil
}

func coerceTypedValue(key string, value any) (any, error) {
	switch key {
	case "value_string":
		v, ok := value.(string)
		if !ok {
			return nil, errors.New("value_string must be a string")
		}
		return v, nil
	case "value_int":
		switch v := value.(type) {
		case int:
			return int64(v), nil
		case int32:
			return int64(v), nil
		case int64:
			return v, nil
		case float64:
			return int64FromFloat(v)
		case json.Number:
			return v.Int64()
		default:
			return nil, errors.New("value_int must be an integer")
		}
	case "value_float":
		switch v := value.(type) {
		case float32:
			return float64(v), nil
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case json.Number:
			return v.Float64()
		default:
			return nil, errors.New("value_float must be numeric")
		}
	case "value_bool":
		v, ok := value.(bool)
		if !ok {
			return nil, errors.New("value_bool must be a boolean")
		}
		return v, nil
	case "value_null":
		v, ok := value.(bool)
		if !ok || !v {
			return nil, errors.New("value_null must be true")
		}
		return nil, nil
	case "value_json":
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("value_json must be JSON-serializable: %w", err)
		}
		return string(encoded), nil
	default:
		return nil, fmt.Errorf("unsupported typed value field %q", key)
	}
}

func int64FromFloat(value float64) (int64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Trunc(value) != value {
		return 0, errors.New("value_int must be an integer")
	}
	parsed, err := strconv.ParseInt(strconv.FormatFloat(value, 'f', 0, 64), 10, 64)
	if err != nil {
		return 0, errors.New("value_int must fit int64")
	}
	return parsed, nil
}

func mapFromAny(value any) (map[string]any, error) {
	switch v := value.(type) {
	case map[string]any:
		return v, nil
	case connectors.Record:
		return map[string]any(v), nil
	default:
		return nil, fmt.Errorf("must be an object, got %T", value)
	}
}

func stringFromRecord(record connectors.Record, key string) (string, bool) {
	value, ok := record[key]
	if !ok || value == nil {
		return "", false
	}
	s, ok := value.(string)
	return s, ok
}
