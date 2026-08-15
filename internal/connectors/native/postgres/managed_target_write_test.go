package postgres

import (
	"encoding/json"
	"math"
	"testing"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

func TestPostgresHistoryModeRequiresManagedTableLock(t *testing.T) {
	if !postgresWriteModeRequiresTableLock(synccontract.ModeIncrementalDedupeHistory) {
		t.Fatal("incremental_dedupe_history did not require the managed-table lock")
	}
}

func TestPostgresManagedTargetTypeMappingAndValueEncoding(t *testing.T) {
	int8Type, err := database.NewSignedInteger(8)
	if err != nil {
		t.Fatal(err)
	}
	uint64Type, err := database.NewUnsignedInteger(64)
	if err != nil {
		t.Fatal(err)
	}
	decimal, err := database.NewDecimal(12, 3)
	if err != nil {
		t.Fatal(err)
	}
	float64Type, err := database.NewFloat(64)
	if err != nil {
		t.Fatal(err)
	}
	text, err := database.NewString(42, "")
	if err != nil {
		t.Fatal(err)
	}
	timezoneTime, err := database.NewTime(3, true)
	if err != nil {
		t.Fatal(err)
	}
	timestamp, err := database.NewTimestamp(6, false)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name      string
		logical   database.LogicalType
		wantType  string
		wantCheck string
	}{
		{name: "int8", logical: int8Type, wantType: "SMALLINT", wantCheck: "%s >= -128 AND %s <= 127"},
		{name: "uint64", logical: uint64Type, wantType: "NUMERIC(20,0)", wantCheck: "%s >= 0 AND %s <= 18446744073709551615"},
		{name: "decimal", logical: decimal, wantType: "NUMERIC(12,3)"},
		{name: "float64", logical: float64Type, wantType: "DOUBLE PRECISION"},
		{name: "boolean", logical: database.NewBoolean(), wantType: "BOOLEAN"},
		{name: "string", logical: text, wantType: "VARCHAR(42)"},
		{name: "binary", logical: mustPostgresManagedTargetBinary(t), wantType: "BYTEA"},
		{name: "date", logical: database.NewDate(), wantType: "DATE"},
		{name: "time with timezone", logical: timezoneTime, wantType: "TIME(3) WITH TIME ZONE"},
		{name: "timestamp without timezone", logical: timestamp, wantType: "TIMESTAMP(6) WITHOUT TIME ZONE"},
		{name: "uuid", logical: database.NewUUID(), wantType: "UUID"},
		{name: "json", logical: database.NewJSON(), wantType: "JSONB"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotType, gotCheck, err := postgresManagedTargetType(tc.logical)
			if err != nil || gotType != tc.wantType || gotCheck != tc.wantCheck {
				t.Fatalf("postgresManagedTargetType() = (%q, %q, %v), want (%q, %q, nil)", gotType, gotCheck, err, tc.wantType, tc.wantCheck)
			}
		})
	}
	array, err := database.NewArray(text)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := postgresManagedTargetType(array); err == nil {
		t.Fatal("postgresManagedTargetType(array) succeeded, want closed unsupported-type refusal")
	}

	if got, err := postgresEncodeMappedValue(^uint64(0)); err != nil || got != "18446744073709551615" {
		t.Fatalf("postgresEncodeMappedValue(uint64 max) = (%#v, %v), want decimal string", got, err)
	}
	if got, err := postgresEncodeMappedValue([16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}); err != nil || got != "00010203-0405-0607-0809-0a0b0c0d0e0f" {
		t.Fatalf("postgresEncodeMappedValue(uuid) = (%#v, %v), want canonical UUID string", got, err)
	}
	if got, err := postgresEncodeMappedValue(json.RawMessage(`{"ok":true}`)); err != nil || got != `{"ok":true}` {
		t.Fatalf("postgresEncodeMappedValue(json) = (%#v, %v), want valid JSON string", got, err)
	}
	if _, err := postgresEncodeMappedValue(math.Inf(1)); err == nil {
		t.Fatal("postgresEncodeMappedValue(+Inf) succeeded, want non-finite float refusal")
	}
	if _, err := postgresEncodeMappedValue(json.RawMessage(`{`)); err == nil {
		t.Fatal("postgresEncodeMappedValue(invalid JSON) succeeded, want refusal")
	}
	if _, err := postgresTombstoneKeyValues(synccontract.Tombstone{Key: json.RawMessage(`{"id":null}`)}, []string{"id"}, []postgresManagedTargetColumn{{name: "id", typeSQL: "BIGINT", nullable: false}}); err == nil {
		t.Fatal("postgresTombstoneKeyValues(null non-null key) succeeded, want refusal")
	}
}

func TestPostgresOrderFenceKeyDigestUsesTypedSQLKeySemantics(t *testing.T) {
	columns := []postgresManagedTargetColumn{
		{name: "id", typeSQL: "BIGINT", nullable: false},
		{name: "ratio", typeSQL: "DOUBLE PRECISION", nullable: false},
	}
	fromMappedRecord, err := postgresKeyDigest([]string{"id", "ratio"}, []any{int64(7), float64(1)}, columns)
	if err != nil {
		t.Fatalf("postgresKeyDigest(mapped values) error = %v", err)
	}
	fromTombstone, err := postgresKeyDigest([]string{"id", "ratio"}, []any{int64(7), "1.0"}, columns)
	if err != nil {
		t.Fatalf("postgresKeyDigest(tombstone values) error = %v", err)
	}
	if !samePostgresBytes(fromMappedRecord, fromTombstone) {
		t.Fatalf("typed order-fence key digest drifted between equivalent mapped/tombstone values: mapped=%x tombstone=%x", fromMappedRecord, fromTombstone)
	}
	if _, err := postgresKeyDigest([]string{"id"}, []any{"not-an-integer"}, columns); err == nil {
		t.Fatal("postgresKeyDigest accepted a key value outside the mapped PostgreSQL type")
	}
}

func mustPostgresManagedTargetBinary(t *testing.T) database.LogicalType {
	t.Helper()
	logical, err := database.NewBinary(0)
	if err != nil {
		t.Fatal(err)
	}
	return logical
}
