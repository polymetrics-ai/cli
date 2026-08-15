package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/big"
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

func postgresInsertMappedHistoryRow(ctx context.Context, tx pgx.Tx, qualified string, columns []postgresManagedTargetColumn, args []any, validFrom time.Time) error {
	names, placeholders := postgresColumnNamesAndPlaceholders(columns, 1)
	names = append(names, quoteIdentifier(synccontract.HistoryValidFromColumn), quoteIdentifier(synccontract.HistoryValidToColumn), quoteIdentifier(synccontract.HistoryIsCurrentColumn))
	placeholders = append(placeholders, "$"+strconv.Itoa(len(args)+1), "$"+strconv.Itoa(len(args)+2), "$"+strconv.Itoa(len(args)+3))
	arguments := append(args, validFrom.UTC(), nil, true)
	_, err := tx.Exec(ctx, "INSERT INTO "+qualified+" ("+strings.Join(names, ", ")+") VALUES ("+strings.Join(placeholders, ", ")+")", arguments...)
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
	args, predicate, err := postgresKeyValuePredicate(keys, values, 1)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "DELETE FROM "+qualified+" WHERE "+predicate, args...)
	return err
}

func postgresKeyValuePredicate(keys []string, values map[string]any, firstPlaceholder int) ([]any, string, error) {
	predicates := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys))
	for index, key := range keys {
		value, exists := values[key]
		if !exists {
			return nil, "", errPostgresWriteValueInvalid
		}
		predicates = append(predicates, quoteIdentifier(key)+" IS NOT DISTINCT FROM $"+strconv.Itoa(firstPlaceholder+index))
		args = append(args, value)
	}
	return args, strings.Join(predicates, " AND "), nil
}

func postgresCloseHistoryKeyValues(ctx context.Context, tx pgx.Tx, qualified string, keys []string, values map[string]any, validTo time.Time) error {
	args, predicate, err := postgresKeyValuesPredicate(keys, values, 2)
	if err != nil {
		return err
	}
	arguments := append([]any{validTo.UTC()}, args...)
	_, err = tx.Exec(ctx, "UPDATE "+qualified+" SET "+quoteIdentifier(synccontract.HistoryValidToColumn)+" = $1, "+quoteIdentifier(synccontract.HistoryIsCurrentColumn)+" = FALSE WHERE "+quoteIdentifier(synccontract.HistoryIsCurrentColumn)+" = TRUE AND "+predicate, arguments...)
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

func postgresKeyValuesPredicate(keys []string, values map[string]any, firstPlaceholder int) ([]any, string, error) {
	predicates := make([]string, 0, len(keys))
	args := make([]any, 0, len(keys))
	for index, key := range keys {
		value, exists := values[key]
		if !exists {
			return nil, "", errPostgresWriteValueInvalid
		}
		predicates = append(predicates, quoteIdentifier(key)+" IS NOT DISTINCT FROM $"+strconv.Itoa(firstPlaceholder+index))
		args = append(args, value)
	}
	return args, strings.Join(predicates, " AND "), nil
}

func postgresMappedKeyDigest(keys []string, mapped connectors.Record, columns []postgresManagedTargetColumn) ([]byte, error) {
	args, _, err := postgresMappedKeyPredicate(keys, mapped, columns, 1)
	if err != nil {
		return nil, err
	}
	return postgresKeyDigest(keys, args, columns)
}

func postgresKeyValuesDigest(keys []string, values map[string]any, columns []postgresManagedTargetColumn) ([]byte, error) {
	args, _, err := postgresKeyValuesPredicate(keys, values, 1)
	if err != nil {
		return nil, err
	}
	return postgresKeyDigest(keys, args, columns)
}

func postgresKeyDigest(keys []string, values []any, columns []postgresManagedTargetColumn) ([]byte, error) {
	if len(keys) != len(values) {
		return nil, errPostgresWriteValueInvalid
	}
	byName := make(map[string]postgresManagedTargetColumn, len(columns))
	for _, column := range columns {
		byName[column.name] = column
	}
	canonical := make([]any, len(values))
	for index, key := range keys {
		column, found := byName[key]
		if !found {
			return nil, errPostgresWriteValueInvalid
		}
		value, err := postgresCanonicalKeyValue(values[index], column.typeSQL)
		if err != nil {
			return nil, err
		}
		canonical[index] = value
	}
	encoded, err := json.Marshal(canonical)
	if err != nil {
		return nil, errPostgresWriteValueInvalid
	}
	digest := sha256.Sum256(encoded)
	return append([]byte(nil), digest[:]...), nil
}

// postgresCanonicalKeyValue derives order-fence identity from the same typed
// PostgreSQL key semantics used by the delete predicate. Tombstone JSON and a
// mapped record can otherwise represent one numeric or temporal SQL value
// with different Go types, which would let an old replay bypass a tombstone.
func postgresCanonicalKeyValue(value any, typeSQL string) (any, error) {
	if value == nil {
		return nil, nil
	}
	switch {
	case strings.HasPrefix(typeSQL, "SMALLINT"), strings.HasPrefix(typeSQL, "INTEGER"), strings.HasPrefix(typeSQL, "BIGINT"), strings.HasPrefix(typeSQL, "NUMERIC"):
		text, err := postgresCanonicalDecimal(value)
		if err != nil {
			return nil, err
		}
		return "numeric:" + text, nil
	case strings.EqualFold(typeSQL, "REAL"):
		parsed, err := postgresCanonicalFloat(value, 32)
		if err != nil {
			return nil, err
		}
		return "real:" + parsed, nil
	case strings.EqualFold(typeSQL, "DOUBLE PRECISION"):
		parsed, err := postgresCanonicalFloat(value, 64)
		if err != nil {
			return nil, err
		}
		return "double:" + parsed, nil
	case strings.HasPrefix(typeSQL, "TIMESTAMP"):
		timestamp, ok := value.(time.Time)
		if !ok {
			return nil, errPostgresWriteValueInvalid
		}
		return "timestamp:" + timestamp.UTC().Format(time.RFC3339Nano), nil
	case strings.EqualFold(typeSQL, "DATE"):
		date, ok := value.(time.Time)
		if !ok {
			return nil, errPostgresWriteValueInvalid
		}
		return "date:" + date.Format("2006-01-02"), nil
	default:
		return value, nil
	}
}

func postgresCanonicalDecimal(value any) (string, error) {
	var text string
	switch typed := value.(type) {
	case string:
		text = typed
	case int64:
		text = strconv.FormatInt(typed, 10)
	case int32:
		text = strconv.FormatInt(int64(typed), 10)
	case int16:
		text = strconv.FormatInt(int64(typed), 10)
	case int8:
		text = strconv.FormatInt(int64(typed), 10)
	default:
		return "", errPostgresWriteValueInvalid
	}
	rational, ok := new(big.Rat).SetString(text)
	if !ok {
		return "", errPostgresWriteValueInvalid
	}
	return rational.RatString(), nil
}

func postgresCanonicalFloat(value any, bitSize int) (string, error) {
	var parsed float64
	switch typed := value.(type) {
	case string:
		result, err := strconv.ParseFloat(typed, bitSize)
		if err != nil || math.IsNaN(result) || math.IsInf(result, 0) {
			return "", errPostgresWriteValueInvalid
		}
		parsed = result
	case float32:
		parsed = float64(typed)
	case float64:
		parsed = typed
	default:
		return "", errPostgresWriteValueInvalid
	}
	if math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return "", errPostgresWriteValueInvalid
	}
	return strconv.FormatFloat(parsed, 'x', -1, bitSize), nil
}

func postgresOrderFenceAccepts(ctx context.Context, tx pgx.Tx, qualifiedFence, relation string, keyDigest []byte, position synccontract.CheckpointPosition) (bool, error) {
	if len(position.Primary) == 0 || len(position.TieBreaker) == 0 || len(keyDigest) != sha256.Size {
		return false, errPostgresWriteValueInvalid
	}
	var primary, tieBreaker []byte
	var deleted bool
	err := tx.QueryRow(ctx, "SELECT source_primary, source_tie_breaker, deleted FROM "+qualifiedFence+" WHERE relation_name = $1 AND key_digest = $2", relation, keyDigest).Scan(&primary, &tieBreaker, &deleted)
	if err == nil {
		stored := synccontract.CheckpointPosition{Primary: primary, TieBreaker: tieBreaker}
		return postgresSourcePositionAfter(position, stored), nil
	}
	if err == pgx.ErrNoRows {
		return true, nil
	}
	return false, err
}

func postgresStoreOrderFence(ctx context.Context, tx pgx.Tx, qualifiedFence, relation string, keyDigest []byte, position synccontract.CheckpointPosition, deleted bool) error {
	if len(position.Primary) == 0 || len(position.TieBreaker) == 0 || len(keyDigest) != sha256.Size {
		return errPostgresWriteValueInvalid
	}
	_, err := tx.Exec(ctx, `INSERT INTO `+qualifiedFence+` (relation_name, key_digest, source_primary, source_tie_breaker, deleted)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (relation_name, key_digest) DO UPDATE
		SET source_primary = EXCLUDED.source_primary,
		    source_tie_breaker = EXCLUDED.source_tie_breaker,
		    deleted = EXCLUDED.deleted`, relation, keyDigest, []byte(position.Primary), []byte(position.TieBreaker), deleted)
	return err
}

func postgresSourcePositionAfter(candidate, stored synccontract.CheckpointPosition) bool {
	if comparison := bytes.Compare(candidate.Primary, stored.Primary); comparison != 0 {
		return comparison > 0
	}
	return bytes.Compare(candidate.TieBreaker, stored.TieBreaker) > 0
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

func postgresMappedValuePredicate(columns []postgresManagedTargetColumn, firstPlaceholder int) string {
	predicates := make([]string, 0, len(columns))
	for index, column := range columns {
		predicates = append(predicates, quoteIdentifier(column.name)+" IS NOT DISTINCT FROM $"+strconv.Itoa(firstPlaceholder+index))
	}
	return strings.Join(predicates, " AND ")
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
	case strings.HasPrefix(typeSQL, "TIMESTAMP"):
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return nil, errPostgresWriteValueInvalid
		}
		return parsed, nil
	case strings.EqualFold(typeSQL, "DATE"):
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			return nil, errPostgresWriteValueInvalid
		}
		return parsed, nil
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
