package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
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

// CreateManagedTarget deliberately refuses before DDL while the shared mapping
// contract is absent. Creating a placeholder business relation would make a
// false schema assertion and force the future mapped writer to auto-evolve it,
// which both the issue and managed-target contract forbid.
func (d *DatabaseDriver) CreateManagedTarget(ctx context.Context, _ database.ManagedTargetProvisioningPlan, _ database.TargetOwner) error {
	if ctx == nil {
		return errPostgresDatabaseDriverConnectionRequired
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if d == nil || d.conn == nil || d.connMu == nil {
		return errPostgresDatabaseDriverConnectionRequired
	}
	return errPostgresTargetLayoutUnavailable
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
