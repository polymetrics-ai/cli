package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/warehouse"
)

const (
	postgresTargetDatabaseIdentityKind = "postgres-database-oid"
	postgresNamespaceIdentityKind      = "postgres-namespace-oid"
	postgresRelationIdentityKind       = "postgres-relation-oid"
)

// AcquireManagedTargetLock holds PostgreSQL's native advisory lock on the
// derived namespace through the matching ReleaseManagedTargetLock call. A
// hash collision may conservatively serialize independent namespaces but can
// never permit two writers through the same namespace lock.
func (d *DatabaseDriver) AcquireManagedTargetLock(ctx context.Context, target database.ManagedTargetRef) (database.ManagedTargetLock, error) {
	if ctx == nil {
		return nil, errPostgresDatabaseDriverConnectionRequired
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if d == nil || d.conn == nil || d.connMu == nil {
		return nil, errPostgresDatabaseDriverConnectionRequired
	}
	d.connMu.Lock()
	defer d.connMu.Unlock()
	if _, err := d.conn.Exec(ctx, "SELECT pg_advisory_lock(hashtextextended($1, 0))", target.Namespace()); err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return nil, contextErr
		}
		return nil, errors.New("postgres managed target lock is unavailable")
	}
	return &postgresManagedTargetLock{conn: d.conn, connMu: d.connMu, namespace: target.Namespace()}, nil
}

type postgresManagedTargetLock struct {
	conn      *pgx.Conn
	connMu    *sync.Mutex
	namespace string
	once      sync.Once
}

func (l *postgresManagedTargetLock) ReleaseManagedTargetLock() {
	if l == nil || l.conn == nil || l.connMu == nil {
		return
	}
	l.once.Do(func() {
		// The shared lock port cannot return an error or accept a context. A
		// bounded cleanup context is therefore the narrowest way to avoid
		// leaking a session-level advisory lock after caller cancellation.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		l.connMu.Lock()
		defer l.connMu.Unlock()
		_, _ = l.conn.Exec(ctx, "SELECT pg_advisory_unlock(hashtextextended($1, 0))", l.namespace)
	})
}

// ObserveManagedTarget returns the server-observed database, namespace, and
// relation identities plus durable owner/control rows. It does no DDL and
// treats unreadable control state as opaque rather than exposing database
// errors at the shared provisioning boundary.
func (d *DatabaseDriver) ObserveManagedTarget(ctx context.Context, target database.ManagedTargetRef) (database.ManagedTargetObservation, error) {
	if ctx == nil {
		return database.ManagedTargetObservation{}, errPostgresDatabaseDriverConnectionRequired
	}
	if err := ctx.Err(); err != nil {
		return database.ManagedTargetObservation{}, err
	}
	if d == nil || d.conn == nil || d.connMu == nil {
		return database.ManagedTargetObservation{}, errPostgresDatabaseDriverConnectionRequired
	}
	d.connMu.Lock()
	defer d.connMu.Unlock()
	targetDatabase, err := d.targetDatabaseIdentity(ctx)
	if err != nil {
		return database.ManagedTargetObservation{}, err
	}

	var namespaceOID string
	err = d.conn.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_namespace WHERE nspname = $1", target.Namespace()).Scan(&namespaceOID)
	if errors.Is(err, pgx.ErrNoRows) {
		return database.ManagedTargetObservation{
			TargetDatabase:      targetDatabase,
			NamespaceOwnerState: database.ManagedTargetNamespaceOwnerAbsent,
			ControlState:        database.ManagedTargetControlAbsent,
		}, nil
	}
	if err != nil {
		return database.ManagedTargetObservation{}, postgresManagedTargetQueryError(ctx, err)
	}
	namespaceNative, err := database.NewNativeNamespaceIdentity(postgresNamespaceIdentityKind, namespaceOID)
	if err != nil {
		return database.ManagedTargetObservation{}, errors.New("postgres managed target namespace identity is invalid")
	}

	observation := database.ManagedTargetObservation{
		TargetDatabase:   targetDatabase,
		NamespacePresent: true,
		NamespaceNative:  namespaceNative,
	}
	namespaceState, namespaceOwner, err := d.loadNamespaceOwner(ctx, target, targetDatabase, namespaceNative)
	if err != nil {
		return database.ManagedTargetObservation{}, err
	}
	observation.NamespaceOwnerState = namespaceState
	observation.NamespaceOwnerRecord = namespaceOwner

	var relationOID string
	err = d.conn.QueryRow(ctx, `
		SELECT c.oid::text
		FROM pg_catalog.pg_class AS c
		WHERE c.relnamespace = $1::oid
		  AND c.relname = $2
		  AND c.relkind IN ('r', 'p')`, namespaceOID, target.Relation()).Scan(&relationOID)
	if errors.Is(err, pgx.ErrNoRows) {
		controlState, control, err := d.loadTargetControl(ctx, target, targetDatabase)
		if err != nil {
			return database.ManagedTargetObservation{}, err
		}
		observation.ControlState = controlState
		observation.ControlRecord = control
		return observation, nil
	}
	if err != nil {
		return database.ManagedTargetObservation{}, postgresManagedTargetQueryError(ctx, err)
	}
	relationNative := database.NativeRelationIdentity{Kind: postgresRelationIdentityKind, Value: relationOID}
	controlState, control, err := d.loadTargetControl(ctx, target, targetDatabase)
	if err != nil {
		return database.ManagedTargetObservation{}, err
	}
	observation.RelationPresent = true
	observation.NativeIdentity = relationNative
	observation.ControlState = controlState
	observation.ControlRecord = control
	if controlState == database.ManagedTargetControlPresent {
		observation.Schema = control.Schema()
		return observation, nil
	}
	if controlState == database.ManagedTargetControlUnreadable {
		return observation, nil
	}
	// A relation without a readable control row is always refused as unowned.
	// The nonzero marker exists only to satisfy the shared observation's
	// structural invariant; it is not a business schema mapping or a target
	// schema claim.
	marker := sha256.Sum256([]byte(fmt.Sprintf("postgres-unowned-relation-v1:%s:%s", namespaceOID, relationOID)))
	schema, err := database.NewManagedTargetSchema(1, database.SchemaFingerprint(marker))
	if err != nil {
		return database.ManagedTargetObservation{}, errors.New("postgres managed target unowned schema marker is invalid")
	}
	observation.Schema = schema
	return observation, nil
}

// CreateManagedTarget creates a mapped target and its private control records
// in one PostgreSQL transaction. The shared provisioning state machine has
// already asserted that no target exists, but this driver still derives every
// identifier from that sealed plan and writes no placeholder relation when its
// MappingContractV1 is absent or unsupported.
func (d *DatabaseDriver) CreateManagedTarget(ctx context.Context, plan database.ManagedTargetProvisioningPlan, owner database.TargetOwner) error {
	if ctx == nil {
		return errPostgresDatabaseDriverConnectionRequired
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if d == nil || d.conn == nil || d.connMu == nil {
		return errPostgresDatabaseDriverConnectionRequired
	}
	if !owner.Identity().SameIdentity(plan.Owner().Identity()) {
		return errPostgresTargetCreateFailed
	}
	mapping, ok := plan.Mapping()
	if !ok {
		return errPostgresTargetMappingRequired
	}
	columns, err := postgresManagedTargetColumns(mapping)
	if err != nil {
		return err
	}

	d.connMu.Lock()
	defer d.connMu.Unlock()
	if err := postgresPreflightDurability(ctx, d.conn); err != nil {
		return err
	}
	targetDatabase, err := d.targetDatabaseIdentity(ctx)
	if err != nil || targetDatabase.Kind() != plan.TargetDatabase().Kind() || targetDatabase.Value() != plan.TargetDatabase().Value() {
		return errPostgresTargetCreateFailed
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errPostgresTargetCreateFailed
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	if _, err := tx.Exec(ctx, "SET LOCAL synchronous_commit = 'on'"); err != nil {
		return errPostgresTargetCreateFailed
	}
	if err := postgresPreflightDurability(ctx, tx); err != nil {
		return err
	}
	if err := postgresCreateManagedTargetLayout(ctx, tx, plan, columns); err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		return errPostgresTargetCreateFailed
	}
	if err := tx.Commit(ctx); err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		return errPostgresTargetCreateFailed
	}
	return nil
}

type postgresManagedTargetColumn struct {
	name       string
	typeSQL    string
	nullable   bool
	collation  string
	constraint string
}

func postgresManagedTargetColumns(mapping database.MappingContractV1) ([]postgresManagedTargetColumn, error) {
	columns := mapping.Columns()
	if len(columns) == 0 {
		return nil, errPostgresTargetMappingRequired
	}
	result := make([]postgresManagedTargetColumn, 0, len(columns))
	for _, column := range columns {
		typeSQL, constraint, err := postgresManagedTargetType(column.Type.Target())
		if err != nil || validateIdentifier(column.Target) != nil {
			return nil, errPostgresTargetTypeUnsupported
		}
		result = append(result, postgresManagedTargetColumn{
			name:       column.Target,
			typeSQL:    typeSQL,
			nullable:   column.Nullable,
			collation:  column.Type.Target().Collation(),
			constraint: constraint,
		})
	}
	return result, nil
}

func postgresManagedTargetType(logical database.LogicalType) (string, string, error) {
	var typeSQL, constraint string
	switch logical.Kind() {
	case database.LogicalSignedInteger:
		switch logical.BitWidth() {
		case 8:
			typeSQL, constraint = "SMALLINT", "%s >= -128 AND %s <= 127"
		case 16:
			typeSQL = "SMALLINT"
		case 32:
			typeSQL = "INTEGER"
		case 64:
			typeSQL = "BIGINT"
		default:
			return "", "", errPostgresTargetTypeUnsupported
		}
	case database.LogicalUnsignedInteger:
		switch logical.BitWidth() {
		case 8:
			typeSQL, constraint = "NUMERIC(3,0)", "%s >= 0 AND %s <= 255"
		case 16:
			typeSQL, constraint = "NUMERIC(5,0)", "%s >= 0 AND %s <= 65535"
		case 32:
			typeSQL, constraint = "NUMERIC(10,0)", "%s >= 0 AND %s <= 4294967295"
		case 64:
			typeSQL, constraint = "NUMERIC(20,0)", "%s >= 0 AND %s <= 18446744073709551615"
		default:
			return "", "", errPostgresTargetTypeUnsupported
		}
	case database.LogicalDecimal:
		typeSQL = "NUMERIC(" + strconv.FormatUint(uint64(logical.Precision()), 10) + "," + strconv.FormatUint(uint64(logical.Scale()), 10) + ")"
	case database.LogicalFloat:
		switch logical.BitWidth() {
		case 32:
			typeSQL = "REAL"
		case 64:
			typeSQL = "DOUBLE PRECISION"
		default:
			return "", "", errPostgresTargetTypeUnsupported
		}
	case database.LogicalBoolean:
		typeSQL = "BOOLEAN"
	case database.LogicalString:
		if logical.MaxLength() == 0 {
			typeSQL = "TEXT"
		} else {
			typeSQL = "VARCHAR(" + strconv.FormatUint(uint64(logical.MaxLength()), 10) + ")"
		}
		if logical.Collation() != "" {
			if validateIdentifier(logical.Collation()) != nil {
				return "", "", errPostgresTargetTypeUnsupported
			}
			typeSQL += " COLLATE " + quoteIdentifier(logical.Collation())
		}
	case database.LogicalBinary:
		typeSQL = "BYTEA"
	case database.LogicalDate:
		typeSQL = "DATE"
	case database.LogicalTime:
		typeSQL = "TIME(" + strconv.FormatUint(uint64(logical.Precision()), 10) + ")"
		if logical.WithTimezone() {
			typeSQL += " WITH TIME ZONE"
		} else {
			typeSQL += " WITHOUT TIME ZONE"
		}
	case database.LogicalTimestamp:
		typeSQL = "TIMESTAMP(" + strconv.FormatUint(uint64(logical.Precision()), 10) + ")"
		if logical.WithTimezone() {
			typeSQL += " WITH TIME ZONE"
		} else {
			typeSQL += " WITHOUT TIME ZONE"
		}
	case database.LogicalUUID:
		typeSQL = "UUID"
	case database.LogicalJSON:
		typeSQL = "JSONB"
	default:
		return "", "", errPostgresTargetTypeUnsupported
	}
	return typeSQL, constraint, nil
}

func postgresCreateManagedTargetLayout(ctx context.Context, tx pgx.Tx, plan database.ManagedTargetProvisioningPlan, columns []postgresManagedTargetColumn) error {
	target := plan.Target()
	namespace := quoteIdentifier(target.Namespace())
	if _, err := tx.Exec(ctx, "CREATE SCHEMA "+namespace); err != nil {
		return err
	}
	ownerTable := postgresQualifiedControlTable(target.Namespace(), postgresNamespaceOwnerTable)
	controlTable := postgresQualifiedControlTable(target.Namespace(), postgresTargetControlTable)
	ledgerTable := postgresQualifiedControlTable(target.Namespace(), postgresDeliveryLedgerTable)
	if _, err := tx.Exec(ctx, `CREATE TABLE `+ownerTable+` (
		singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
		workspace_id TEXT NOT NULL,
		connector_id TEXT NOT NULL,
		connection_id TEXT NOT NULL,
		target_database_oid TEXT NOT NULL,
		namespace_oid TEXT NOT NULL
	)`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE `+controlTable+` (
		workspace_id TEXT NOT NULL,
		connector_id TEXT NOT NULL,
		connection_id TEXT NOT NULL,
		stream_id TEXT NOT NULL UNIQUE,
		relation_name TEXT NOT NULL PRIMARY KEY,
		target_database_oid TEXT NOT NULL,
		relation_oid TEXT NOT NULL,
		schema_version BIGINT NOT NULL CHECK (schema_version > 0),
		schema_fingerprint BYTEA NOT NULL CHECK (octet_length(schema_fingerprint) = 32)
	)`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `CREATE TABLE `+ledgerTable+` (
		stream_id TEXT NOT NULL PRIMARY KEY,
		relation_name TEXT NOT NULL,
		target_database_oid TEXT NOT NULL,
		delivery_id TEXT NOT NULL
	)`); err != nil {
		return err
	}
	columnDDL := make([]string, 0, len(columns))
	for _, column := range columns {
		definition := quoteIdentifier(column.name) + " " + column.typeSQL
		if !column.nullable {
			definition += " NOT NULL"
		}
		if column.constraint != "" {
			quoted := quoteIdentifier(column.name)
			definition += " CHECK (" + fmt.Sprintf(column.constraint, quoted, quoted) + ")"
		}
		columnDDL = append(columnDDL, definition)
	}
	relationTable := quoteIdentifier(target.Namespace()) + "." + quoteIdentifier(target.Relation())
	if _, err := tx.Exec(ctx, "CREATE TABLE "+relationTable+" ("+strings.Join(columnDDL, ", ")+")"); err != nil {
		return err
	}
	var namespaceOID, relationOID string
	if err := tx.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_namespace WHERE nspname = $1", target.Namespace()).Scan(&namespaceOID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `
		SELECT c.oid::text
		FROM pg_catalog.pg_class AS c
		JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2 AND c.relkind IN ('r', 'p')`, target.Namespace(), target.Relation()).Scan(&relationOID); err != nil {
		return err
	}
	identity := plan.Owner().Identity()
	if _, err := tx.Exec(ctx, "INSERT INTO "+ownerTable+" (workspace_id, connector_id, connection_id, target_database_oid, namespace_oid) VALUES ($1, $2, $3, $4, $5)", identity.WorkspaceID, identity.ConnectorID, identity.ConnectionID, plan.TargetDatabase().Value(), namespaceOID); err != nil {
		return err
	}
	fingerprint := plan.Schema().Fingerprint().Bytes()
	if _, err := tx.Exec(ctx, "INSERT INTO "+controlTable+" (workspace_id, connector_id, connection_id, stream_id, relation_name, target_database_oid, relation_oid, schema_version, schema_fingerprint) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)", identity.WorkspaceID, identity.ConnectorID, identity.ConnectionID, target.StreamID(), target.Relation(), plan.TargetDatabase().Value(), relationOID, plan.Schema().Version(), fingerprint); err != nil {
		return err
	}
	return nil
}

func (d *DatabaseDriver) targetDatabaseIdentity(ctx context.Context) (database.TargetDatabaseIdentity, error) {
	var oid string
	if err := d.conn.QueryRow(ctx, "SELECT oid::text FROM pg_catalog.pg_database WHERE datname = current_database()").Scan(&oid); err != nil {
		return database.TargetDatabaseIdentity{}, postgresManagedTargetQueryError(ctx, err)
	}
	identity, err := database.NewTargetDatabaseIdentity(postgresTargetDatabaseIdentityKind, oid)
	if err != nil {
		return database.TargetDatabaseIdentity{}, errors.New("postgres target database identity is invalid")
	}
	return identity, nil
}

func postgresManagedTargetQueryError(ctx context.Context, err error) error {
	if contextErr := ctx.Err(); contextErr != nil {
		return contextErr
	}
	return errors.New("postgres managed target query failed")
}

func postgresOwner(workspaceID, connectorID, connectionID string) (database.TargetOwner, error) {
	return database.NewTargetOwner(warehouse.ArtifactIdentity{
		WorkspaceID:  workspaceID,
		ConnectorID:  connectorID,
		ConnectionID: connectionID,
	})
}

var _ database.ManagedTargetProvisioningDriver = (*DatabaseDriver)(nil)
