package postgres

import "polymetrics/internal/connectors"

// pgTypeToFieldType maps a PostgreSQL information_schema.columns.data_type value
// to the connector's coarse field type vocabulary. Unknown types fall back to
// "string" so discovery never fails on an exotic type.
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
// conformance harness and unit tests can run without a live database. It mirrors
// the shape a real information_schema discovery would produce.
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

// fixtureRows returns deterministic canned rows for a fixture stream. The cursor
// column (updated_at here, surfaced as a numeric string for ordering in tests)
// is monotonically increasing so the incremental lower bound can filter them.
// Returns (rows, true) when the stream is known.
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

// fixtureRow pairs a canned record with a numeric cursor value used to emulate
// incremental cursor filtering deterministically in fixture mode.
type fixtureRow struct {
	cursor int64
	record connectors.Record
}
