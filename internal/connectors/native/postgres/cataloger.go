package postgres

import (
	"context"
	"errors"

	"polymetrics.ai/internal/connectors"
)

// Catalog is the legacy compatibility projection of TypedCatalog. Its live
// table/field metadata is always derived from the one typed PostgreSQL catalog
// result; fixture mode is deliberately the only canned test-only exception.
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	var result connectors.Catalog
	err := executeWithAuthenticationAdmission(ctx, cfg, func(admitted context.Context) error {
		var err error
		result, err = c.catalog(admitted, cfg)
		return err
	})
	return result, err
}

func (c Connector) catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	if fixtureMode(cfg) {
		// Keep fixture mode's existing configuration validation and static rows
		// isolated from the shipping live discovery path.
		if _, err := resolveConfig(cfg); err != nil {
			return connectors.Catalog{}, err
		}
		return connectors.Catalog{Connector: c.Name(), Streams: fixtureStreams()}, nil
	}

	typed, err := c.typedCatalog(ctx, cfg)
	if err != nil {
		if errors.Is(err, ErrNoSupportedRelations) {
			return connectors.Catalog{Connector: c.Name(), Streams: []connectors.Stream{}}, nil
		}
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: legacyStreamsFromTypedCatalog(typed)}, nil
}

// fixtureStreams is the canned catalog returned in fixture mode so the
// fixture tests and unit tests can run without a live database. It is
// deliberately confined to mode=fixture and is never used by live pg_catalog
// discovery (ported verbatim from legacy internal/connectors/postgres/streams.go).
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
