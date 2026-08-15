package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
)

var (
	// ErrUnsupportedCatalogShape reports a catalog shape that cannot yet be
	// represented losslessly by the typed PostgreSQL source boundary. It carries
	// no identifier, configuration, or credential material.
	ErrUnsupportedCatalogShape = errors.New("postgres catalog contains an unsupported relation, type, or identifier shape")
	errCatalogResourcePolicy   = errors.New("postgres typed catalog resource policy is invalid")
	// ErrSystemCatalogSchema reports a configured PostgreSQL-owned namespace.
	// It deliberately carries no schema, configuration, or credential material.
	ErrSystemCatalogSchema = errors.New("postgres catalog schema is reserved for PostgreSQL system objects")
	// ErrNoSupportedRelations means the configured schema contains no base
	// relation. The legacy Catalog projection can represent that as no streams;
	// the #4034 Catalog value deliberately requires at least one relation.
	ErrNoSupportedRelations = errors.New("postgres catalog contains no supported base tables")

	errTypedCatalogFixtureMode = errors.New("postgres typed catalog is unavailable in fixture mode")
)

const typedCatalogReadAuthorizationSQL = `
  AND pg_catalog.has_schema_privilege(n.oid, 'USAGE')
  AND (
    pg_catalog.has_table_privilege(c.oid, 'SELECT')
    OR NOT EXISTS (
      SELECT 1
      FROM pg_catalog.pg_attribute AS readable_attribute
      WHERE readable_attribute.attrelid = c.oid
        AND readable_attribute.attnum > 0
        AND NOT readable_attribute.attisdropped
        AND NOT pg_catalog.has_column_privilege(c.oid, readable_attribute.attnum, 'SELECT')
    )
  )`

const typedCatalogRelationsSQL = `
SELECT n.nspname,
       c.relname,
       c.oid
FROM pg_catalog.pg_class AS c
JOIN pg_catalog.pg_namespace AS n
  ON n.oid = c.relnamespace
WHERE n.nspname = $1
  AND c.relkind IN ('r', 'p')` + typedCatalogReadAuthorizationSQL + `
ORDER BY n.nspname, c.relname, c.oid`

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
  AND c.relkind IN ('r', 'p')` + typedCatalogReadAuthorizationSQL + `
  AND a.attnum > 0
  AND NOT a.attisdropped
ORDER BY n.nspname, c.relname, a.attnum`

const typedCatalogKeysSQL = `
SELECT n.nspname,
       c.relname,
       c.oid,
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
  AND c.relkind IN ('r', 'p')` + typedCatalogReadAuthorizationSQL + `
  AND con.contype IN ('p', 'u')
ORDER BY n.nspname, c.relname, con.conname, key_column.ordinality`

// TypedCatalog reads the configured PostgreSQL database's configured schema
// from pg_catalog and returns #4034's normalized catalog/fingerprint. It is a
// source discovery operation only: it sends no DDL, writes, or arbitrary SQL.
func (c Connector) TypedCatalog(ctx context.Context, cfg connectors.RuntimeConfig) (database.Catalog, error) {
	var result database.Catalog
	err := executeWithAuthenticationAdmission(ctx, cfg, func(admitted context.Context) error {
		var err error
		result, err = c.typedCatalog(admitted, cfg)
		return err
	})
	return result, err
}

func (c Connector) typedCatalog(ctx context.Context, cfg connectors.RuntimeConfig) (database.Catalog, error) {
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
	if isSystemCatalogSchema(conn.schema) {
		return database.Catalog{}, ErrSystemCatalogSchema
	}
	if err := c.databaseDefinition.Validate(); err != nil {
		return database.Catalog{}, errors.New("postgres typed catalog definition is unavailable")
	}
	resources, err := newTypedCatalogResources(c.databaseDefinition.Resources())
	if err != nil {
		return database.Catalog{}, err
	}
	operationCtx, cancel, err := resources.operationContext(ctx)
	if err != nil {
		return database.Catalog{}, err
	}
	defer cancel()

	pool, err := conn.openTypedCatalogPool(operationCtx, resources)
	if err != nil {
		return database.Catalog{}, fmt.Errorf("catalog postgres: open pool: %w", err)
	}
	defer pool.Close()

	return discoverTypedCatalog(operationCtx, pool, conn.database, conn.schema, c.databaseDefinition, resources)
}

func isSystemCatalogSchema(schema string) bool {
	switch schema {
	case "pg_catalog", "information_schema", "pg_toast":
		return true
	default:
		return strings.HasPrefix(schema, "pg_toast_") || strings.HasPrefix(schema, "pg_temp_")
	}
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

type typedCatalogResources struct {
	policy   database.ResourcePolicy
	poolSize int32
}

func newTypedCatalogResources(policy database.ResourcePolicy) (typedCatalogResources, error) {
	poolSize, err := policy.EffectivePoolSize(0)
	if err != nil {
		return typedCatalogResources{}, fmt.Errorf("%w: %v", errCatalogResourcePolicy, err)
	}
	return typedCatalogResources{policy: policy, poolSize: int32(poolSize)}, nil
}

func (resources typedCatalogResources) operationContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	operationCtx, cancel, err := resources.policy.WithOperationTimeout(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("%w: %v", errCatalogResourcePolicy, err)
	}
	return operationCtx, cancel, nil
}

func discoverTypedCatalog(ctx context.Context, pool *pgxpool.Pool, databaseName, schema string, definition database.Definition, resources typedCatalogResources) (database.Catalog, error) {
	tx, err := pool.BeginTx(ctx, typedCatalogTransactionOptions())
	if err != nil {
		return database.Catalog{}, fmt.Errorf("catalog postgres: begin typed catalog snapshot: %w", err)
	}
	catalog, discoverErr := discoverTypedCatalogSnapshot(ctx, tx, databaseName, schema, definition)
	rollbackErr := rollbackTypedCatalogSnapshot(ctx, tx, resources)
	if rollbackErr != nil {
		if discoverErr != nil {
			return database.Catalog{}, errors.Join(discoverErr, rollbackErr)
		}
		return database.Catalog{}, rollbackErr
	}
	if discoverErr != nil {
		return database.Catalog{}, discoverErr
	}
	return catalog, nil
}

func typedCatalogTransactionOptions() pgx.TxOptions {
	return pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}
}

func rollbackTypedCatalogSnapshot(ctx context.Context, tx pgx.Tx, resources typedCatalogResources) error {
	rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), resources.policy.ConnectTimeout)
	defer cancel()
	if err := tx.Rollback(rollbackCtx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return fmt.Errorf("catalog postgres: close typed catalog snapshot: %w", err)
	}
	return nil
}

func discoverTypedCatalogSnapshot(ctx context.Context, tx pgx.Tx, databaseName, schema string, definition database.Definition) (database.Catalog, error) {
	order, byOID, err := discoverTypedCatalogRelations(ctx, tx, databaseName, schema)
	if err != nil {
		return database.Catalog{}, err
	}
	if len(order) == 0 {
		return database.Catalog{}, ErrNoSupportedRelations
	}
	if err := discoverTypedCatalogColumns(ctx, tx, schema, definition, byOID); err != nil {
		return database.Catalog{}, err
	}
	for _, relationOID := range order {
		if len(byOID[relationOID].relation.Columns) == 0 {
			return database.Catalog{}, ErrUnsupportedCatalogShape
		}
	}
	if err := discoverTypedCatalogKeys(ctx, tx, schema, byOID); err != nil {
		return database.Catalog{}, err
	}

	relations := make([]database.Relation, 0, len(order))
	for _, relationOID := range order {
		relations = append(relations, byOID[relationOID].relation)
	}
	catalog, err := database.NewCatalog(database.CatalogRef{Name: databaseName}, relations)
	if err != nil {
		return database.Catalog{}, ErrUnsupportedCatalogShape
	}
	return catalog, nil
}

func discoverTypedCatalogRelations(ctx context.Context, tx pgx.Tx, databaseName, schema string) ([]uint32, map[uint32]*typedCatalogRelationBuilder, error) {
	rows, err := tx.Query(ctx, typedCatalogRelationsSQL, schema)
	if err != nil {
		return nil, nil, fmt.Errorf("catalog postgres: query typed relations: %w", err)
	}
	defer rows.Close()

	byOID := make(map[uint32]*typedCatalogRelationBuilder)
	order := make([]uint32, 0)
	for rows.Next() {
		var relationSchema, relationName string
		var relationOID uint32
		if err := rows.Scan(&relationSchema, &relationName, &relationOID); err != nil {
			return nil, nil, fmt.Errorf("catalog postgres: scan typed relation: %w", err)
		}
		if relationSchema != schema || relationOID == 0 || validateIdentifier(relationSchema) != nil || validateIdentifier(relationName) != nil {
			return nil, nil, ErrUnsupportedCatalogShape
		}
		if _, found := byOID[relationOID]; found {
			return nil, nil, ErrUnsupportedCatalogShape
		}
		ref := database.RelationRef{
			Schema: database.SchemaRef{
				Catalog: database.CatalogRef{Name: databaseName},
				Name:    relationSchema,
			},
			Name: relationName,
		}
		byOID[relationOID] = &typedCatalogRelationBuilder{relation: database.Relation{
			Ref: ref,
			NativeIdentity: database.NativeRelationIdentity{
				Kind:  "oid",
				Value: strconv.FormatUint(uint64(relationOID), 10),
			},
		}}
		order = append(order, relationOID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("catalog postgres: iterate typed relations: %w", err)
	}
	return order, byOID, nil
}

func discoverTypedCatalogColumns(ctx context.Context, tx pgx.Tx, schema string, definition database.Definition, relations map[uint32]*typedCatalogRelationBuilder) error {
	rows, err := tx.Query(ctx, typedCatalogColumnsSQL, schema)
	if err != nil {
		return fmt.Errorf("catalog postgres: query typed columns: %w", err)
	}
	defer rows.Close()

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
			return fmt.Errorf("catalog postgres: scan typed column: %w", err)
		}
		if column.schema != schema || column.relationOID == 0 || column.ordinal <= 0 || validateIdentifier(column.schema) != nil || validateIdentifier(column.relation) != nil || validateIdentifier(column.column) != nil {
			return ErrUnsupportedCatalogShape
		}
		builder, found := relations[column.relationOID]
		if !found || builder.relation.Ref.Schema.Name != column.schema || builder.relation.Ref.Name != column.relation {
			return ErrUnsupportedCatalogShape
		}
		nativeType, logicalType, err := postgresColumnType(definition, column)
		if err != nil {
			return err
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
		return fmt.Errorf("catalog postgres: iterate typed columns: %w", err)
	}
	return nil
}

func discoverTypedCatalogKeys(ctx context.Context, tx pgx.Tx, schema string, relations map[uint32]*typedCatalogRelationBuilder) error {
	rows, err := tx.Query(ctx, typedCatalogKeysSQL, schema)
	if err != nil {
		return fmt.Errorf("catalog postgres: query typed keys: %w", err)
	}
	defer rows.Close()

	keyIndexes := make(map[string]int)
	for rows.Next() {
		var rowSchema, relationName, keyName, keyKind, columnName string
		var relationOID uint32
		var ordinal int
		if err := rows.Scan(&rowSchema, &relationName, &relationOID, &keyName, &keyKind, &columnName, &ordinal); err != nil {
			return fmt.Errorf("catalog postgres: scan typed key: %w", err)
		}
		if rowSchema != schema || relationOID == 0 || validateIdentifier(rowSchema) != nil || validateIdentifier(relationName) != nil || validateIdentifier(columnName) != nil || ordinal <= 0 {
			return ErrUnsupportedCatalogShape
		}
		builder, found := relations[relationOID]
		if !found || builder.relation.Ref.Schema.Name != rowSchema || builder.relation.Ref.Name != relationName {
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
		keyID := strconv.FormatUint(uint64(relationOID), 10) + "\x00" + keyName
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
