package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
)

var (
	// ErrUnsupportedCatalogShape reports a catalog shape that cannot yet be
	// represented losslessly by the typed PostgreSQL source boundary. It carries
	// no identifier, configuration, or credential material.
	ErrUnsupportedCatalogShape = errors.New("postgres catalog contains an unsupported type or identifier shape")
	// ErrNoSupportedRelations means the configured schema contains no base
	// relation. The legacy Catalog projection can represent that as no streams;
	// the #4034 Catalog value deliberately requires at least one relation.
	ErrNoSupportedRelations = errors.New("postgres catalog contains no supported base tables")

	errTypedCatalogFixtureMode = errors.New("postgres typed catalog is unavailable in fixture mode")
)

const typedCatalogColumnsSQL = `
SELECT n.nspname,
       c.relname,
       c.oid,
       a.attname,
       a.attnum,
       NOT a.attnotnull,
       t.typname,
       t.typtype::text,
       t.typelem,
       a.atttypmod,
       COALESCE(coll.collname, '')
FROM pg_catalog.pg_class AS c
JOIN pg_catalog.pg_namespace AS n
  ON n.oid = c.relnamespace
JOIN pg_catalog.pg_attribute AS a
  ON a.attrelid = c.oid
JOIN pg_catalog.pg_type AS t
  ON t.oid = a.atttypid
LEFT JOIN pg_catalog.pg_collation AS coll
  ON coll.oid = a.attcollation
WHERE n.nspname = $1
  AND c.relkind IN ('r', 'p')
  AND a.attnum > 0
  AND NOT a.attisdropped
ORDER BY n.nspname, c.relname, a.attnum`

const typedCatalogKeysSQL = `
SELECT n.nspname,
       c.relname,
       con.conname,
       con.contype::text,
       a.attname,
       key_column.ordinality
FROM pg_catalog.pg_constraint AS con
JOIN pg_catalog.pg_class AS c
  ON c.oid = con.conrelid
JOIN pg_catalog.pg_namespace AS n
  ON n.oid = c.relnamespace
JOIN unnest(con.conkey) WITH ORDINALITY AS key_column(attnum, ordinality)
  ON true
JOIN pg_catalog.pg_attribute AS a
  ON a.attrelid = c.oid
 AND a.attnum = key_column.attnum
WHERE n.nspname = $1
  AND c.relkind IN ('r', 'p')
  AND con.contype IN ('p', 'u')
ORDER BY n.nspname, c.relname, con.conname, key_column.ordinality`

// TypedCatalog reads the configured PostgreSQL database's configured schema
// from pg_catalog and returns #4034's normalized catalog/fingerprint. It is a
// source discovery operation only: it sends no DDL, writes, or arbitrary SQL.
func (c Connector) TypedCatalog(ctx context.Context, cfg connectors.RuntimeConfig) (database.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return database.Catalog{}, err
	}
	conn, err := resolveConfig(cfg)
	if err != nil {
		return database.Catalog{}, err
	}
	if fixtureMode(cfg) {
		return database.Catalog{}, errTypedCatalogFixtureMode
	}
	if err := validateIdentifier(conn.schema); err != nil {
		return database.Catalog{}, ErrUnsupportedCatalogShape
	}
	if err := c.databaseDefinition.Validate(); err != nil {
		return database.Catalog{}, errors.New("postgres typed catalog definition is unavailable")
	}

	pool, err := conn.openPool(ctx)
	if err != nil {
		return database.Catalog{}, fmt.Errorf("catalog postgres: open pool: %w", err)
	}
	defer pool.Close()

	return discoverTypedCatalog(ctx, pool, conn.database, conn.schema, c.databaseDefinition)
}

type postgresCatalogColumn struct {
	schema       string
	relation     string
	relationOID  uint32
	column       string
	ordinal      int
	nullable     bool
	nativeName   string
	typeKind     string
	elementOID   uint32
	typeModifier int32
	collation    string
}

type typedCatalogRelationBuilder struct {
	relation database.Relation
}

// discoverTypedCatalog groups one ordered pg_catalog row set into the #4034
// catalog model, then adds ordered primary and unique constraint membership.
func discoverTypedCatalog(ctx context.Context, pool *pgxpool.Pool, databaseName, schema string, definition database.Definition) (database.Catalog, error) {
	rows, err := pool.Query(ctx, typedCatalogColumnsSQL, schema)
	if err != nil {
		return database.Catalog{}, fmt.Errorf("catalog postgres: query typed columns: %w", err)
	}
	defer rows.Close()

	byRelation := make(map[string]*typedCatalogRelationBuilder)
	order := make([]string, 0)
	for rows.Next() {
		var column postgresCatalogColumn
		if err := rows.Scan(
			&column.schema,
			&column.relation,
			&column.relationOID,
			&column.column,
			&column.ordinal,
			&column.nullable,
			&column.nativeName,
			&column.typeKind,
			&column.elementOID,
			&column.typeModifier,
			&column.collation,
		); err != nil {
			return database.Catalog{}, fmt.Errorf("catalog postgres: scan typed column: %w", err)
		}
		if column.schema != schema || validateIdentifier(column.schema) != nil || validateIdentifier(column.relation) != nil || validateIdentifier(column.column) != nil {
			return database.Catalog{}, ErrUnsupportedCatalogShape
		}
		relationKey := typedCatalogRelationKey(column.schema, column.relation)
		builder, found := byRelation[relationKey]
		if !found {
			ref := database.RelationRef{
				Schema: database.SchemaRef{
					Catalog: database.CatalogRef{Name: databaseName},
					Name:    column.schema,
				},
				Name: column.relation,
			}
			builder = &typedCatalogRelationBuilder{relation: database.Relation{
				Ref: ref,
				NativeIdentity: database.NativeRelationIdentity{
					Kind:  "oid",
					Value: strconv.FormatUint(uint64(column.relationOID), 10),
				},
			}}
			byRelation[relationKey] = builder
			order = append(order, relationKey)
		}
		nativeType, logicalType, err := postgresColumnType(definition, column)
		if err != nil {
			return database.Catalog{}, err
		}
		builder.relation.Columns = append(builder.relation.Columns, database.Column{
			Ref: database.ColumnRef{
				Relation: builder.relation.Ref,
				Name:     column.column,
			},
			Type:     logicalType,
			Nullable: column.nullable,
			Ordinal:  column.ordinal,
			Native:   &nativeType,
		})
	}
	if err := rows.Err(); err != nil {
		return database.Catalog{}, fmt.Errorf("catalog postgres: iterate typed columns: %w", err)
	}
	// Release the result's pooled connection before the independent key query;
	// discovery stays within the declared finite pool bound even when it is one.
	rows.Close()
	if len(order) == 0 {
		return database.Catalog{}, ErrNoSupportedRelations
	}

	if err := discoverTypedCatalogKeys(ctx, pool, schema, byRelation); err != nil {
		return database.Catalog{}, err
	}
	relations := make([]database.Relation, 0, len(order))
	for _, key := range order {
		relations = append(relations, byRelation[key].relation)
	}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: databaseName}, relations)
	if err != nil {
		// The foundation intentionally does not render discovered identifiers in
		// errors. Treat an unrepresentable catalog as a named safe rejection.
		return database.Catalog{}, ErrUnsupportedCatalogShape
	}
	return catalog, nil
}

func discoverTypedCatalogKeys(ctx context.Context, pool *pgxpool.Pool, schema string, relations map[string]*typedCatalogRelationBuilder) error {
	rows, err := pool.Query(ctx, typedCatalogKeysSQL, schema)
	if err != nil {
		return fmt.Errorf("catalog postgres: query typed keys: %w", err)
	}
	defer rows.Close()

	keyIndexes := make(map[string]int)
	for rows.Next() {
		var rowSchema, relationName, keyName, keyKind, columnName string
		var ordinal int
		if err := rows.Scan(&rowSchema, &relationName, &keyName, &keyKind, &columnName, &ordinal); err != nil {
			return fmt.Errorf("catalog postgres: scan typed key: %w", err)
		}
		if rowSchema != schema || validateIdentifier(rowSchema) != nil || validateIdentifier(relationName) != nil || validateIdentifier(columnName) != nil || ordinal <= 0 {
			return ErrUnsupportedCatalogShape
		}
		builder, found := relations[typedCatalogRelationKey(rowSchema, relationName)]
		if !found {
			return ErrUnsupportedCatalogShape
		}
		kind := database.KeyUnique
		switch keyKind {
		case "p":
			kind = database.KeyPrimary
		case "u":
		default:
			return ErrUnsupportedCatalogShape
		}
		keyID := typedCatalogRelationKey(rowSchema, relationName) + "\x00" + keyName
		index, found := keyIndexes[keyID]
		if !found {
			index = len(builder.relation.Keys)
			keyIndexes[keyID] = index
			builder.relation.Keys = append(builder.relation.Keys, database.Key{Name: keyName, Kind: kind})
		} else if builder.relation.Keys[index].Kind != kind {
			return ErrUnsupportedCatalogShape
		}
		key := &builder.relation.Keys[index]
		if ordinal != len(key.Columns)+1 {
			return ErrUnsupportedCatalogShape
		}
		key.Columns = append(key.Columns, database.ColumnRef{Relation: builder.relation.Ref, Name: columnName})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("catalog postgres: iterate typed keys: %w", err)
	}
	return nil
}

func typedCatalogRelationKey(schema, relation string) string {
	return schema + "\x00" + relation
}
