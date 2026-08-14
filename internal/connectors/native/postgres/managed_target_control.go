package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"polymetrics.ai/internal/connectors/database"
)

const (
	postgresNamespaceOwnerTable = "__polymetrics_namespace_owner"
	postgresTargetControlTable  = "__polymetrics_target_control"
	postgresDeliveryLedgerTable = "__polymetrics_delivery_ledger"
)

func (d *DatabaseDriver) loadNamespaceOwner(ctx context.Context, target database.ManagedTargetRef, targetDatabase database.TargetDatabaseIdentity, native database.NativeNamespaceIdentity) (database.ManagedTargetNamespaceOwnerState, database.ManagedTargetNamespaceOwnerRecord, error) {
	query := fmt.Sprintf(`
		SELECT workspace_id, connector_id, connection_id, target_database_oid, namespace_oid
		FROM %s
		LIMIT 2`, postgresQualifiedControlTable(target.Namespace(), postgresNamespaceOwnerTable))
	rows, err := d.conn.Query(ctx, query)
	if err != nil {
		if postgresUndefinedTable(err) {
			return database.ManagedTargetNamespaceOwnerAbsent, database.ManagedTargetNamespaceOwnerRecord{}, nil
		}
		if postgresPermissionDenied(err) {
			return database.ManagedTargetNamespaceOwnerUnreadable, database.ManagedTargetNamespaceOwnerRecord{}, nil
		}
		return database.ManagedTargetNamespaceOwnerUnknown, database.ManagedTargetNamespaceOwnerRecord{}, postgresManagedTargetQueryError(ctx, err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return database.ManagedTargetNamespaceOwnerUnknown, database.ManagedTargetNamespaceOwnerRecord{}, postgresManagedTargetQueryError(ctx, err)
		}
		return database.ManagedTargetNamespaceOwnerAbsent, database.ManagedTargetNamespaceOwnerRecord{}, nil
	}
	var workspaceID, connectorID, connectionID, databaseOID, namespaceOID string
	if err := rows.Scan(&workspaceID, &connectorID, &connectionID, &databaseOID, &namespaceOID); err != nil || rows.Next() {
		return database.ManagedTargetNamespaceOwnerUnknown, database.ManagedTargetNamespaceOwnerRecord{}, errors.New("postgres managed target namespace owner record is invalid")
	}
	if err := rows.Err(); err != nil {
		return database.ManagedTargetNamespaceOwnerUnknown, database.ManagedTargetNamespaceOwnerRecord{}, postgresManagedTargetQueryError(ctx, err)
	}
	owner, err := postgresOwner(workspaceID, connectorID, connectionID)
	if err != nil || databaseOID != targetDatabase.Value() || namespaceOID != native.Value() {
		return database.ManagedTargetNamespaceOwnerUnknown, database.ManagedTargetNamespaceOwnerRecord{}, errors.New("postgres managed target namespace owner record is invalid")
	}
	record, err := database.NewManagedTargetNamespaceOwnerRecord(owner, target, targetDatabase, native)
	if err != nil {
		return database.ManagedTargetNamespaceOwnerUnknown, database.ManagedTargetNamespaceOwnerRecord{}, errors.New("postgres managed target namespace owner record is invalid")
	}
	return database.ManagedTargetNamespaceOwnerPresent, record, nil
}

func (d *DatabaseDriver) loadTargetControl(ctx context.Context, target database.ManagedTargetRef, targetDatabase database.TargetDatabaseIdentity) (database.ManagedTargetControlState, database.ManagedTargetControlRecord, error) {
	query := fmt.Sprintf(`
		SELECT workspace_id, connector_id, connection_id, stream_id, target_database_oid,
		       relation_oid, schema_version, schema_fingerprint
		FROM %s
		WHERE relation_name = $1`, postgresQualifiedControlTable(target.Namespace(), postgresTargetControlTable))
	var workspaceID, connectorID, connectionID, streamID, databaseOID, relationOID string
	var schemaVersion int64
	var fingerprintBytes []byte
	err := d.conn.QueryRow(ctx, query, target.Relation()).Scan(
		&workspaceID, &connectorID, &connectionID, &streamID, &databaseOID,
		&relationOID, &schemaVersion, &fingerprintBytes,
	)
	if errors.Is(err, pgx.ErrNoRows) || postgresUndefinedTable(err) {
		return database.ManagedTargetControlAbsent, database.ManagedTargetControlRecord{}, nil
	}
	if postgresPermissionDenied(err) {
		return database.ManagedTargetControlUnreadable, database.ManagedTargetControlRecord{}, nil
	}
	if err != nil {
		return database.ManagedTargetControlUnknown, database.ManagedTargetControlRecord{}, postgresManagedTargetQueryError(ctx, err)
	}
	owner, err := postgresOwner(workspaceID, connectorID, connectionID)
	if err != nil || streamID != target.StreamID() || databaseOID != targetDatabase.Value() {
		return database.ManagedTargetControlUnknown, database.ManagedTargetControlRecord{}, errors.New("postgres managed target control record is invalid")
	}
	var fingerprint database.SchemaFingerprint
	if len(fingerprintBytes) != len(fingerprint) {
		return database.ManagedTargetControlUnknown, database.ManagedTargetControlRecord{}, errors.New("postgres managed target control record is invalid")
	}
	copy(fingerprint[:], fingerprintBytes)
	if schemaVersion <= 0 {
		return database.ManagedTargetControlUnknown, database.ManagedTargetControlRecord{}, errors.New("postgres managed target control record is invalid")
	}
	schema, err := database.NewManagedTargetSchema(uint(schemaVersion), fingerprint)
	if err != nil {
		return database.ManagedTargetControlUnknown, database.ManagedTargetControlRecord{}, errors.New("postgres managed target control record is invalid")
	}
	control, err := database.NewManagedTargetControlRecord(owner, target, targetDatabase, database.NativeRelationIdentity{
		Kind: postgresRelationIdentityKind, Value: relationOID,
	}, schema)
	if err != nil {
		return database.ManagedTargetControlUnknown, database.ManagedTargetControlRecord{}, errors.New("postgres managed target control record is invalid")
	}
	return database.ManagedTargetControlPresent, control, nil
}

// LoadManagedTargetDelivery reads one opaque delivery record from the private
// namespace-local control ledger. Missing control storage is an error rather
// than an opportunity to create or adopt a target on the receipt path.
func (d *DatabaseDriver) LoadManagedTargetDelivery(ctx context.Context, key database.ManagedTargetDeliveryLedgerKey) (database.ManagedTargetDeliveryRecord, bool, error) {
	if ctx == nil {
		return database.ManagedTargetDeliveryRecord{}, false, errPostgresDatabaseDriverConnectionRequired
	}
	if err := ctx.Err(); err != nil {
		return database.ManagedTargetDeliveryRecord{}, false, err
	}
	if d == nil || d.conn == nil || d.connMu == nil {
		return database.ManagedTargetDeliveryRecord{}, false, errPostgresDatabaseDriverConnectionRequired
	}
	d.connMu.Lock()
	defer d.connMu.Unlock()
	query := fmt.Sprintf(`SELECT delivery_id FROM %s WHERE stream_id = $1 AND relation_name = $2 AND target_database_oid = $3`, postgresQualifiedControlTable(key.Namespace(), postgresDeliveryLedgerTable))
	var deliveryID string
	err := d.conn.QueryRow(ctx, query, key.StreamID(), key.Relation(), key.TargetDatabase().Value()).Scan(&deliveryID)
	if errors.Is(err, pgx.ErrNoRows) {
		return database.ManagedTargetDeliveryRecord{}, false, nil
	}
	if err != nil {
		return database.ManagedTargetDeliveryRecord{}, false, postgresManagedTargetQueryError(ctx, err)
	}
	record, err := database.NewManagedTargetDeliveryRecord(deliveryID)
	if err != nil {
		return database.ManagedTargetDeliveryRecord{}, false, errors.New("postgres managed target delivery record is invalid")
	}
	return record, true, nil
}

// StoreManagedTargetDelivery persists one opaque receipt only after the
// mapped future write session has confirmed commit. This method does not
// perform target DDL or accept a relation supplied outside the sealed key.
func (d *DatabaseDriver) StoreManagedTargetDelivery(ctx context.Context, key database.ManagedTargetDeliveryLedgerKey, record database.ManagedTargetDeliveryRecord) error {
	if ctx == nil {
		return errPostgresDatabaseDriverConnectionRequired
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if d == nil || d.conn == nil || d.connMu == nil {
		return errPostgresDatabaseDriverConnectionRequired
	}
	d.connMu.Lock()
	defer d.connMu.Unlock()
	query := fmt.Sprintf(`
		INSERT INTO %s (stream_id, relation_name, target_database_oid, delivery_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (stream_id) DO UPDATE
		SET relation_name = EXCLUDED.relation_name,
		    target_database_oid = EXCLUDED.target_database_oid,
		    delivery_id = EXCLUDED.delivery_id`, postgresQualifiedControlTable(key.Namespace(), postgresDeliveryLedgerTable))
	if _, err := d.conn.Exec(ctx, query, key.StreamID(), key.Relation(), key.TargetDatabase().Value(), record.DeliveryID()); err != nil {
		return postgresManagedTargetQueryError(ctx, err)
	}
	return nil
}

func postgresQualifiedControlTable(namespace, table string) string {
	return quoteIdentifier(namespace) + "." + quoteIdentifier(table)
}

func postgresUndefinedTable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42P01"
}

func postgresPermissionDenied(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "42501"
}

var _ database.ManagedTargetDeliveryLedgerStore = (*DatabaseDriver)(nil)
