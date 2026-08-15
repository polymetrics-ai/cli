package postgres

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

func (d *DatabaseDriver) assertManagedTargetForWrite(ctx context.Context, querier postgresManagedTargetQuerier, plan database.DatabaseWritePlan) error {
	control := plan.Control()
	target := control.Target()
	if target.Namespace() == "" || target.Relation() == "" || len(plan.Mapping().Columns()) == 0 {
		return errPostgresWriteTargetUnverified
	}
	targetDatabase, err := postgresTargetDatabaseIdentity(ctx, querier)
	if err != nil || targetDatabase.Kind() != control.TargetDatabase().Kind() || targetDatabase.Value() != control.TargetDatabase().Value() {
		return errPostgresWriteTargetUnverified
	}
	var namespaceOID, relationOID string
	if err := querier.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_namespace WHERE nspname = $1", target.Namespace()).Scan(&namespaceOID); err != nil || namespaceOID == "" {
		return errPostgresWriteTargetUnverified
	}
	if err := querier.QueryRow(ctx, `
		SELECT c.oid::text
		FROM pg_catalog.pg_class AS c
		JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2 AND c.relkind IN ('r', 'p')`, target.Namespace(), target.Relation()).Scan(&relationOID); err != nil || relationOID != control.NativeIdentity().Value {
		return errPostgresWriteTargetUnverified
	}
	if err := postgresAssertNamespaceOwner(ctx, querier, control, namespaceOID); err != nil {
		return errPostgresWriteTargetUnverified
	}
	if err := postgresAssertTargetControl(ctx, querier, control, relationOID); err != nil {
		return errPostgresWriteTargetUnverified
	}
	if err := postgresAssertMappedRelation(ctx, querier, target, plan.Mapping(), plan.Mode() == synccontract.ModeIncrementalDedupeHistory); err != nil {
		return err
	}
	if plan.ConditionalOrderFence() {
		return postgresAssertOrderFence(ctx, querier, target.Namespace())
	}
	return nil
}

type postgresManagedTargetQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func postgresTargetDatabaseIdentity(ctx context.Context, querier postgresManagedTargetQuerier) (database.TargetDatabaseIdentity, error) {
	var oid string
	if err := querier.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_database WHERE datname = current_database()").Scan(&oid); err != nil {
		return database.TargetDatabaseIdentity{}, err
	}
	return database.NewTargetDatabaseIdentity(postgresTargetDatabaseIdentityKind, oid)
}

func postgresAssertNamespaceOwner(ctx context.Context, querier postgresManagedTargetQuerier, control database.ManagedTargetControlRecord, namespaceOID string) error {
	query := `SELECT workspace_id, connector_id, connection_id, target_database_oid, namespace_oid FROM ` + postgresQualifiedControlTable(control.Target().Namespace(), postgresNamespaceOwnerTable) + ` LIMIT 2`
	rows, err := querier.Query(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	if !rows.Next() {
		return errPostgresWriteTargetUnverified
	}
	var workspaceID, connectorID, connectionID, databaseOID, storedNamespaceOID string
	if err := rows.Scan(&workspaceID, &connectorID, &connectionID, &databaseOID, &storedNamespaceOID); err != nil || rows.Next() || rows.Err() != nil {
		return errPostgresWriteTargetUnverified
	}
	identity := control.Owner().Identity()
	if workspaceID != identity.WorkspaceID || connectorID != identity.ConnectorID || connectionID != identity.ConnectionID || databaseOID != control.TargetDatabase().Value() || storedNamespaceOID != namespaceOID {
		return errPostgresWriteTargetUnverified
	}
	return nil
}

func postgresAssertTargetControl(ctx context.Context, querier postgresManagedTargetQuerier, control database.ManagedTargetControlRecord, relationOID string) error {
	query := `SELECT workspace_id, connector_id, connection_id, stream_id, target_database_oid, relation_oid, schema_version, schema_fingerprint FROM ` + postgresQualifiedControlTable(control.Target().Namespace(), postgresTargetControlTable) + ` WHERE relation_name = $1`
	var workspaceID, connectorID, connectionID, streamID, databaseOID, storedRelationOID string
	var version int64
	var fingerprint []byte
	if err := querier.QueryRow(ctx, query, control.Target().Relation()).Scan(&workspaceID, &connectorID, &connectionID, &streamID, &databaseOID, &storedRelationOID, &version, &fingerprint); err != nil {
		return err
	}
	identity := control.Owner().Identity()
	wantFingerprint := control.Schema().Fingerprint().Bytes()
	if workspaceID != identity.WorkspaceID || connectorID != identity.ConnectorID || connectionID != identity.ConnectionID || streamID != control.Target().StreamID() || databaseOID != control.TargetDatabase().Value() || storedRelationOID != relationOID || storedRelationOID != control.NativeIdentity().Value || version != int64(control.Schema().Version()) || !samePostgresBytes(fingerprint, wantFingerprint) {
		return errPostgresWriteTargetUnverified
	}
	return nil
}

func postgresAssertMappedRelation(ctx context.Context, querier postgresManagedTargetQuerier, target database.ManagedTargetRef, mapping database.MappingContractV1, allowHistoryColumns bool) error {
	columns, err := postgresManagedTargetColumns(mapping)
	if err != nil {
		return err
	}
	rows, err := querier.Query(ctx, `
		SELECT a.attname, a.attnotnull, pg_catalog.format_type(a.atttypid, a.atttypmod), COALESCE(coll.collname, '')
		FROM pg_catalog.pg_attribute AS a
		JOIN pg_catalog.pg_class AS c ON c.oid = a.attrelid
		JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
		LEFT JOIN pg_catalog.pg_collation AS coll ON coll.oid = a.attcollation
		WHERE n.nspname = $1 AND c.relname = $2 AND a.attnum > 0 AND NOT a.attisdropped
		ORDER BY a.attnum`, target.Namespace(), target.Relation())
	if err != nil {
		return err
	}
	defer rows.Close()
	columns = append(columns, postgresManagedTargetSystemColumns()...)
	for _, expected := range columns {
		if !rows.Next() {
			return errPostgresWriteTargetUnverified
		}
		var name, typeName, collation string
		var notNull bool
		if err := rows.Scan(&name, &notNull, &typeName, &collation); err != nil || name != expected.name || notNull == expected.nullable || (expected.collation != "" && collation != expected.collation) || !postgresEquivalentColumnType(expected.typeSQL, typeName) {
			return errPostgresWriteTargetUnverified
		}
	}
	if !allowHistoryColumns {
		if rows.Next() || rows.Err() != nil {
			return errPostgresWriteTargetUnverified
		}
		return nil
	}
	for index, expected := range postgresManagedTargetHistoryColumns() {
		if !rows.Next() {
			// A newly provisioned managed target has only its mapped business
			// columns. BeginDatabaseWrite adds the complete history layout inside
			// its transaction, before the first history row is visible.
			if index == 0 {
				return rows.Err()
			}
			return errPostgresWriteTargetUnverified
		}
		var name, typeName, collation string
		var notNull bool
		if err := rows.Scan(&name, &notNull, &typeName, &collation); err != nil || name != expected.name || notNull == expected.nullable || collation != "" || !postgresEquivalentColumnType(expected.typeSQL, typeName) {
			return errPostgresWriteTargetUnverified
		}
	}
	if rows.Next() || rows.Err() != nil {
		return errPostgresWriteTargetUnverified
	}
	return nil
}

func postgresAssertOrderFence(ctx context.Context, querier postgresManagedTargetQuerier, namespace string) error {
	rows, err := querier.Query(ctx, `SELECT relation_name, key_digest, source_primary, source_tie_breaker, deleted FROM `+postgresQualifiedControlTable(namespace, postgresOrderFenceTable)+` LIMIT 1`)
	if err != nil {
		return errPostgresWriteTargetUnverified
	}
	rows.Close()
	return rows.Err()
}

func postgresEquivalentColumnType(expected, observed string) bool {
	normalize := func(value string) string {
		value = strings.ToLower(strings.TrimSpace(value))
		value = strings.ReplaceAll(value, "varchar", "character varying")
		value = strings.ReplaceAll(value, "without time zone", "")
		value = strings.Join(strings.Fields(value), " ")
		value = strings.ReplaceAll(value, ", ", ",")
		return value
	}
	return normalize(expected) == normalize(observed)
}

func samePostgresBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
