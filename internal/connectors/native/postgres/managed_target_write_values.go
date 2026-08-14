package postgres

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func postgresMappedArguments(columns []postgresManagedTargetColumn, mapped connectors.Record) ([]any, error) {
	args := make([]any, 0, len(columns))
	for _, column := range columns {
		value, exists := mapped[column.name]
		if !exists {
			return nil, errPostgresWriteValueInvalid
		}
		encoded, err := postgresEncodeMappedValue(value)
		if err != nil {
			return nil, err
		}
		args = append(args, encoded)
	}
	return args, nil
}

func postgresEncodeMappedValue(value any) (any, error) {
	switch typed := value.(type) {
	case nil:
		return nil, nil
	case int8:
		return int64(typed), nil
	case int16:
		return int64(typed), nil
	case int32:
		return int64(typed), nil
	case int64, bool, string:
		return typed, nil
	case uint8:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(typed), 10), nil
	case uint64:
		return strconv.FormatUint(typed, 10), nil
	case float32:
		if math.IsNaN(float64(typed)) || math.IsInf(float64(typed), 0) {
			return nil, errPostgresWriteValueInvalid
		}
		return typed, nil
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return nil, errPostgresWriteValueInvalid
		}
		return typed, nil
	case []byte:
		return append([]byte(nil), typed...), nil
	case json.RawMessage:
		if !json.Valid(typed) {
			return nil, errPostgresWriteValueInvalid
		}
		return string(typed), nil
	case [16]byte:
		return formatPostgresUUID(typed), nil
	case time.Time:
		return typed, nil
	default:
		return nil, errPostgresWriteValueInvalid
	}
}

func formatPostgresUUID(value [16]byte) string {
	encoded := hex.EncodeToString(value[:])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:]
}

func postgresInsertMappedRow(ctx context.Context, tx pgx.Tx, qualified string, columns []postgresManagedTargetColumn, args []any) error {
	names, placeholders := postgresColumnNamesAndPlaceholders(columns, 1)
	_, err := tx.Exec(ctx, "INSERT INTO "+qualified+" ("+strings.Join(names, ", ")+") VALUES ("+strings.Join(placeholders, ", ")+")", args...)
	return err
}

func postgresInsertMappedRowIfAbsent(ctx context.Context, tx pgx.Tx, qualified string, keys []string, columns []postgresManagedTargetColumn, mapped connectors.Record, args []any) error {
	names, placeholders := postgresColumnNamesAndPlaceholders(columns, 1)
	keyArgs, predicate, err := postgresMappedKeyPredicate(keys, mapped, columns, len(args)+1)
	if err != nil {
		return err
	}
	arguments := append(args, keyArgs...)
	_, err = tx.Exec(ctx, "INSERT INTO "+qualified+" ("+strings.Join(names, ", ")+") SELECT "+strings.Join(placeholders, ", ")+" WHERE NOT EXISTS (SELECT 1 FROM "+qualified+" WHERE "+predicate+")", arguments...)
	return err
}

func postgresDeleteMappedKeys(ctx context.Context, tx pgx.Tx, qualified string, keys []string, mapped connectors.Record, columns []postgresManagedTargetColumn) error {
	args, predicate, err := postgresMappedKeyPredicate(keys, mapped, columns, 1)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "DELETE FROM "+qualified+" WHERE "+predicate, args...)
	return err
}

func postgresDeleteKeyValues(ctx context.Context, tx pgx.Tx, qualified string, keys []string, values map[string]any) error {
	predicates := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys))
	for index, key := range keys {
		value, exists := values[key]
		if !exists {
			return errPostgresWriteValueInvalid
		}
		predicates = append(predicates, quoteIdentifier(key)+" IS NOT DISTINCT FROM $"+strconv.Itoa(index+1))
		args = append(args, value)
	}
	_, err := tx.Exec(ctx, "DELETE FROM "+qualified+" WHERE "+strings.Join(predicates, " AND "), args...)
	return err
}

func postgresMappedKeyPredicate(keys []string, mapped connectors.Record, columns []postgresManagedTargetColumn, firstPlaceholder int) ([]any, string, error) {
	byName := make(map[string]postgresManagedTargetColumn, len(columns))
	for _, column := range columns {
		byName[column.name] = column
	}
	predicates := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys))
	for index, key := range keys {
		if _, exists := byName[key]; !exists {
			return nil, "", errPostgresWriteValueInvalid
		}
		value, exists := mapped[key]
		if !exists {
			return nil, "", errPostgresWriteValueInvalid
		}
		encoded, err := postgresEncodeMappedValue(value)
		if err != nil {
			return nil, "", err
		}
		predicates = append(predicates, quoteIdentifier(key)+" IS NOT DISTINCT FROM $"+strconv.Itoa(firstPlaceholder+index))
		args = append(args, encoded)
	}
	return args, strings.Join(predicates, " AND "), nil
}

func postgresColumnNamesAndPlaceholders(columns []postgresManagedTargetColumn, firstPlaceholder int) ([]string, []string) {
	names := make([]string, 0, len(columns))
	placeholders := make([]string, 0, len(columns))
	for index, column := range columns {
		names = append(names, quoteIdentifier(column.name))
		placeholders = append(placeholders, "$"+strconv.Itoa(firstPlaceholder+index))
	}
	return names, placeholders
}

func postgresTombstoneKeyValues(tombstone synccontract.Tombstone, keys []string, columns []postgresManagedTargetColumn) (map[string]any, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(tombstone.Key, &raw); err != nil || len(raw) != len(keys) {
		return nil, errPostgresWriteValueInvalid
	}
	byName := make(map[string]postgresManagedTargetColumn, len(columns))
	for _, column := range columns {
		byName[column.name] = column
	}
	values := make(map[string]any, len(keys))
	for _, key := range keys {
		value, exists := raw[key]
		column, mapped := byName[key]
		if !exists || !mapped || (string(value) == "null" && !column.nullable) {
			return nil, errPostgresWriteValueInvalid
		}
		decoded, err := postgresTombstoneValue(value, column.typeSQL)
		if err != nil {
			return nil, err
		}
		values[key] = decoded
	}
	return values, nil
}

func postgresTombstoneValue(raw json.RawMessage, typeSQL string) (any, error) {
	if string(raw) == "null" {
		return nil, nil
	}
	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		return postgresEncodeTombstoneString(stringValue, typeSQL)
	}
	var boolean bool
	if err := json.Unmarshal(raw, &boolean); err == nil && strings.EqualFold(typeSQL, "BOOLEAN") {
		return boolean, nil
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	var number any
	if err := decoder.Decode(&number); err != nil {
		return nil, errPostgresWriteValueInvalid
	}
	jsonNumber, ok := number.(json.Number)
	if !ok {
		return nil, errPostgresWriteValueInvalid
	}
	return postgresEncodeTombstoneNumber(jsonNumber, typeSQL)
}

func postgresEncodeTombstoneString(value, typeSQL string) (any, error) {
	switch {
	case strings.HasPrefix(typeSQL, "SMALLINT"), strings.HasPrefix(typeSQL, "INTEGER"), strings.HasPrefix(typeSQL, "BIGINT"):
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, errPostgresWriteValueInvalid
		}
		return parsed, nil
	case strings.HasPrefix(typeSQL, "NUMERIC"), strings.EqualFold(typeSQL, "REAL"), strings.EqualFold(typeSQL, "DOUBLE PRECISION"):
		return value, nil
	case strings.EqualFold(typeSQL, "BYTEA"), strings.EqualFold(typeSQL, "JSONB"):
		return nil, errPostgresWriteValueInvalid
	default:
		return value, nil
	}
}

func postgresEncodeTombstoneNumber(value json.Number, typeSQL string) (any, error) {
	switch {
	case strings.HasPrefix(typeSQL, "SMALLINT"), strings.HasPrefix(typeSQL, "INTEGER"), strings.HasPrefix(typeSQL, "BIGINT"):
		parsed, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return nil, errPostgresWriteValueInvalid
		}
		return parsed, nil
	case strings.HasPrefix(typeSQL, "NUMERIC"), strings.EqualFold(typeSQL, "REAL"), strings.EqualFold(typeSQL, "DOUBLE PRECISION"):
		return string(value), nil
	default:
		return nil, errPostgresWriteValueInvalid
	}
}
