package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

const (
	cdcSlotPrefix          = "pm_cdc_"
	cdcSnapshotBarrierKind = "postgres_logical_slot"
	cdcProtocolVersion     = "pgoutput-v1"
	cdcCheckpointSchema    = "postgres-cdc-v1"
	cdcChangefeedMechanism = "logical_replication"
	cdcExecutorID          = "postgres_logical_replication"
	cdcSlotCleanupTimeout  = 5 * time.Second
)

var (
	ErrCDCSlotActive                   = errors.New("postgres CDC replication slot is active")
	errCDCFixtureMode                  = errors.New("postgres CDC requires a real PostgreSQL source; fixture mode is not a replication protocol")
	errCDCRelationHasDescendants       = errors.New("postgres CDC does not support a selected relation with descendant tables")
	errCDCRelationNotPublished         = errors.New("postgres CDC selected relation is not included in the configured publication")
	errCDCPublicationMissingDML        = errors.New("postgres CDC requires a publication that publishes insert, update, and delete changes")
	errCDCPublicationPublishesTruncate = errors.New("postgres CDC requires a publication without TRUNCATE events")
)

type postgresCDCSource struct {
	identity   synccontract.SourceIdentity
	generation synccontract.OpaqueToken
	slotName   string
	schema     string
	table      string
	system     pglogrepl.IdentifySystemResult
}

type postgresCDCPublicationScope struct {
	publicationOID     string
	publicationVersion string
	relationOID        string
	membershipVersion  string
	hasDescendants     bool
	published          bool
	publishesInsert    bool
	publishesUpdate    bool
	publishesDelete    bool
	publishesTruncate  bool
}

type postgresReplicationSlot struct {
	plugin       string
	database     string
	active       bool
	confirmedLSN string
	restartLSN   string
}

type postgresCDCSlotSetup struct {
	barrier pglogrepl.LSN
	created bool
}

func replicationConnection(ctx context.Context, conn connConfig) (*pgconn.PgConn, error) {
	config, err := pgconn.ParseConfig(conn.dsn())
	if err != nil {
		return nil, fmt.Errorf("postgres CDC: configure replication connection: %w", err)
	}
	if config.RuntimeParams == nil {
		config.RuntimeParams = make(map[string]string)
	}
	config.RuntimeParams["replication"] = "database"
	replication, err := pgconn.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("postgres CDC: connect replication: %w", err)
	}
	return replication, nil
}

func closeReplicationConnection(conn *pgconn.PgConn) {
	if conn == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), cdcSlotCleanupTimeout)
	defer cancel()
	_ = conn.Close(ctx)
}

func closeCDCDataConnection(conn *pgx.Conn) {
	if conn == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), cdcSlotCleanupTimeout)
	defer cancel()
	_ = conn.Close(ctx)
}

func identifyCDCSource(ctx context.Context, replication *pgconn.PgConn, database, schema, stream string, publication string) (postgresCDCSource, error) {
	system, err := pglogrepl.IdentifySystem(ctx, replication)
	if err != nil {
		return postgresCDCSource{}, fmt.Errorf("postgres CDC: identify source: %w", err)
	}
	if strings.TrimSpace(system.SystemID) == "" || strings.TrimSpace(system.DBName) == "" {
		return postgresCDCSource{}, errors.New("postgres CDC: source identity is incomplete")
	}
	if system.DBName != database {
		return postgresCDCSource{}, errors.New("postgres CDC: replication source database does not match configured database")
	}
	canonicalStream, err := canonicalCDCStream(schema, stream)
	if err != nil {
		return postgresCDCSource{}, fmt.Errorf("postgres CDC: stream: %w", err)
	}
	identity := synccontract.SourceIdentity{
		Engine:           "postgres",
		AccountOrCluster: system.SystemID + ":" + system.DBName,
		ObjectScope:      canonicalStream,
	}
	if err := identity.Validate(); err != nil {
		return postgresCDCSource{}, fmt.Errorf("postgres CDC: source identity: %w", err)
	}
	sourceSchema, sourceTable, _ := strings.Cut(canonicalStream, ".")
	generation := synccontract.OpaqueToken([]byte(strconv.FormatInt(int64(system.Timeline), 10) + "\n" + publication))
	return postgresCDCSource{
		identity:   identity,
		generation: generation,
		slotName:   cdcSlotName(identity),
		schema:     sourceSchema,
		table:      sourceTable,
		system:     system,
	}, nil
}

func (source postgresCDCSource) withPublicationScope(scope postgresCDCPublicationScope) postgresCDCSource {
	fingerprint := scope.fingerprint()
	generation := make(synccontract.OpaqueToken, 0, len(source.generation)+1+len(fingerprint))
	generation = append(generation, source.generation...)
	generation = append(generation, '\n')
	generation = append(generation, fingerprint...)
	source.generation = generation
	return source
}

// canonicalCDCStream turns a caller stream into the fully qualified relation
// name that is bound into a checkpoint source identity. This makes a default
// schema explicit, so a checkpoint cannot move between similarly named tables
// in different schemas.
func canonicalCDCStream(schema, stream string) (string, error) {
	stream = strings.TrimSpace(stream)
	defaultSchemaName := strings.TrimSpace(schema)
	if defaultSchemaName == "" {
		defaultSchemaName = defaultSchema
	}
	table := stream
	if index := strings.IndexByte(stream, '.'); index >= 0 {
		defaultSchemaName = stream[:index]
		table = stream[index+1:]
	}
	if err := validateIdentifier(defaultSchemaName); err != nil {
		return "", fmt.Errorf("schema: %w", err)
	}
	if err := validateIdentifier(table); err != nil {
		return "", fmt.Errorf("table: %w", err)
	}
	return defaultSchemaName + "." + table, nil
}

func cdcSlotName(identity synccontract.SourceIdentity) string {
	sum := sha256.Sum256([]byte(identity.Engine + "\x00" + identity.AccountOrCluster + "\x00" + identity.ObjectScope))
	return cdcSlotPrefix + hex.EncodeToString(sum[:16])
}

func cdcPublication(cfg connectors.RuntimeConfig) (string, error) {
	publication := strings.TrimSpace(cfg.Config["cdc_publication"])
	if publication == "" {
		return "", errors.New("postgres CDC requires config cdc_publication")
	}
	if err := validateIdentifier(publication); err != nil {
		return "", fmt.Errorf("postgres CDC: cdc_publication: %w", err)
	}
	return publication, nil
}

func requireRealCDCSource(cfg connectors.RuntimeConfig) error {
	if fixtureMode(cfg) {
		return errCDCFixtureMode
	}
	return nil
}

func replicationSlotStatus(ctx context.Context, conn connConfig, slotName string) (postgresReplicationSlot, bool, error) {
	data, err := pgx.Connect(ctx, conn.dsn())
	if err != nil {
		return postgresReplicationSlot{}, false, fmt.Errorf("postgres CDC: connect slot inspection: %w", err)
	}
	defer func() { _ = data.Close(ctx) }()

	var slot postgresReplicationSlot
	err = data.QueryRow(ctx, `
	SELECT COALESCE(plugin, ''), COALESCE(database, ''), active,
	       COALESCE(confirmed_flush_lsn::text, ''), COALESCE(restart_lsn::text, '')
FROM pg_replication_slots
WHERE slot_name = $1`, slotName).Scan(&slot.plugin, &slot.database, &slot.active, &slot.confirmedLSN, &slot.restartLSN)
	if errors.Is(err, pgx.ErrNoRows) {
		return postgresReplicationSlot{}, false, nil
	}
	if err != nil {
		return postgresReplicationSlot{}, false, fmt.Errorf("postgres CDC: inspect replication slot: %w", err)
	}
	return slot, true, nil
}

func validateReplicationSlot(slot postgresReplicationSlot, database string) error {
	if slot.plugin != "pgoutput" {
		return errors.New("postgres CDC: derived replication slot uses an incompatible output plugin")
	}
	if slot.database != database {
		return errors.New("postgres CDC: derived replication slot belongs to another database")
	}
	if slot.active {
		return fmt.Errorf("%w: derived replication slot is already active", ErrCDCSlotActive)
	}
	return nil
}

func slotBarrier(slot postgresReplicationSlot, fallback pglogrepl.LSN) (pglogrepl.LSN, error) {
	for _, raw := range []string{slot.confirmedLSN, slot.restartLSN} {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		lsn, err := pglogrepl.ParseLSN(raw)
		if err != nil {
			return 0, errors.New("postgres CDC: replication slot has an invalid retained position")
		}
		return lsn, nil
	}
	if fallback == 0 {
		return 0, errors.New("postgres CDC: replication slot has no usable start position")
	}
	return fallback, nil
}

func openCDCRelationScopeConnection(ctx context.Context, conn connConfig) (*pgx.Conn, error) {
	data, err := pgx.Connect(ctx, conn.dsn())
	if err != nil {
		return nil, fmt.Errorf("postgres CDC: connect relation scope inspection: %w", err)
	}
	return data, nil
}

func inspectCDCRelationScope(ctx context.Context, data *pgx.Conn, source postgresCDCSource, publication string) (postgresCDCPublicationScope, error) {
	if data == nil {
		return postgresCDCPublicationScope{}, errors.New("postgres CDC: relation scope inspection connection is required")
	}
	var scope postgresCDCPublicationScope
	err := data.QueryRow(ctx, `
	SELECT
		COALESCE((
			SELECT publication.oid::text
			FROM pg_publication publication
			WHERE publication.pubname = $3
		), ''),
		COALESCE((
			SELECT publication.xmin::text
			FROM pg_publication publication
			WHERE publication.pubname = $3
		), ''),
		COALESCE((
			SELECT relation.oid::text
			FROM pg_class relation
			JOIN pg_namespace relation_schema ON relation_schema.oid = relation.relnamespace
			WHERE relation_schema.nspname = $1 AND relation.relname = $2
		), ''),
		COALESCE((
			SELECT publication_relation.xmin::text
			FROM pg_publication_rel publication_relation
			JOIN pg_publication publication ON publication.oid = publication_relation.prpubid
			JOIN pg_class relation ON relation.oid = publication_relation.prrelid
			JOIN pg_namespace relation_schema ON relation_schema.oid = relation.relnamespace
			WHERE publication.pubname = $3 AND relation_schema.nspname = $1 AND relation.relname = $2
		), ''),
		EXISTS (
			SELECT 1
			FROM pg_inherits inheritance
			JOIN pg_class parent ON parent.oid = inheritance.inhparent
			JOIN pg_namespace parent_schema ON parent_schema.oid = parent.relnamespace
			WHERE parent_schema.nspname = $1 AND parent.relname = $2
		),
		EXISTS (
			SELECT 1
			FROM pg_publication_tables
			WHERE pubname = $3 AND schemaname = $1 AND tablename = $2
		),
		COALESCE((
			SELECT pubinsert
			FROM pg_publication
			WHERE pubname = $3
		), false),
		COALESCE((
			SELECT pubupdate
			FROM pg_publication
			WHERE pubname = $3
		), false),
		COALESCE((
			SELECT pubdelete
			FROM pg_publication
			WHERE pubname = $3
		), false),
		COALESCE((
			SELECT pubtruncate
			FROM pg_publication
			WHERE pubname = $3
		), false)`, source.schema, source.table, publication).Scan(
		&scope.publicationOID,
		&scope.publicationVersion,
		&scope.relationOID,
		&scope.membershipVersion,
		&scope.hasDescendants,
		&scope.published,
		&scope.publishesInsert,
		&scope.publishesUpdate,
		&scope.publishesDelete,
		&scope.publishesTruncate,
	)
	if err != nil {
		return postgresCDCPublicationScope{}, fmt.Errorf("postgres CDC: inspect selected relation scope: %w", err)
	}
	return scope, nil
}

func (scope postgresCDCPublicationScope) fingerprint() string {
	parts := []string{
		scope.publicationOID,
		scope.publicationVersion,
		scope.relationOID,
		scope.membershipVersion,
		strconv.FormatBool(scope.hasDescendants),
		strconv.FormatBool(scope.published),
		strconv.FormatBool(scope.publishesInsert),
		strconv.FormatBool(scope.publishesUpdate),
		strconv.FormatBool(scope.publishesDelete),
		strconv.FormatBool(scope.publishesTruncate),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}

func (scope postgresCDCPublicationScope) validate() error {
	if err := validateCDCPublicationScope(scope.published); err != nil {
		return err
	}
	if err := validateCDCRelationHierarchy(scope.hasDescendants); err != nil {
		return err
	}
	if err := validateCDCPublicationDML(scope.publishesInsert, scope.publishesUpdate, scope.publishesDelete); err != nil {
		return err
	}
	return validateCDCPublicationTruncate(scope.publishesTruncate)
}

func validateCDCPublicationScopeChange(expected, current postgresCDCPublicationScope) error {
	if expected.fingerprint() != current.fingerprint() {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeSourceGenerationChanged, "PostgreSQL publication or selected relation scope changed")
	}
	return current.validate()
}

func validateCurrentCDCRelationScope(ctx context.Context, data *pgx.Conn, source postgresCDCSource, publication string, expected postgresCDCPublicationScope) error {
	current, err := inspectCDCRelationScope(ctx, data, source, publication)
	if err != nil {
		return err
	}
	return validateCDCPublicationScopeChange(expected, current)
}

func validateCDCPublicationScope(published bool) error {
	if !published {
		return errCDCRelationNotPublished
	}
	return nil
}

func validateCDCRelationHierarchy(hasDescendants bool) error {
	if hasDescendants {
		return errCDCRelationHasDescendants
	}
	return nil
}

func validateCDCPublicationDML(publishesInsert, publishesUpdate, publishesDelete bool) error {
	if !publishesInsert || !publishesUpdate || !publishesDelete {
		return errCDCPublicationMissingDML
	}
	return nil
}

func validateCDCPublicationTruncate(publishesTruncate bool) error {
	if publishesTruncate {
		return errCDCPublicationPublishesTruncate
	}
	return nil
}

func ensureReplicationSlot(ctx context.Context, replication *pgconn.PgConn, conn connConfig, source postgresCDCSource) (postgresCDCSlotSetup, error) {
	slot, found, err := replicationSlotStatus(ctx, conn, source.slotName)
	if err != nil {
		return postgresCDCSlotSetup{}, err
	}
	if found {
		if err := validateReplicationSlot(slot, conn.database); err != nil {
			return postgresCDCSlotSetup{}, err
		}
		barrier, err := slotBarrier(slot, source.system.XLogPos)
		return postgresCDCSlotSetup{barrier: barrier}, err
	}

	// NOEXPORT_SNAPSHOT is accepted by PostgreSQL 12+ and avoids leaving an
	// exported snapshot around. A CDC run reads only changes after the slot's
	// consistent point; initial snapshotting is an explicit, separate sync.
	created, createErr := pglogrepl.CreateReplicationSlot(ctx, replication, source.slotName, "pgoutput", pglogrepl.CreateReplicationSlotOptions{Mode: pglogrepl.LogicalReplication, SnapshotAction: "NOEXPORT_SNAPSHOT"})
	if createErr != nil {
		// A concurrent process may have created this derived slot after the
		// inspection. Re-read it and accept only the same safe shape.
		slot, found, err = replicationSlotStatus(ctx, conn, source.slotName)
		if err != nil || !found {
			return postgresCDCSlotSetup{}, fmt.Errorf("postgres CDC: create replication slot: %w", createErr)
		}
		if err := validateReplicationSlot(slot, conn.database); err != nil {
			return postgresCDCSlotSetup{}, err
		}
		barrier, err := slotBarrier(slot, source.system.XLogPos)
		return postgresCDCSlotSetup{barrier: barrier}, err
	}
	if created.OutputPlugin != "pgoutput" {
		return postgresCDCSlotSetup{}, errors.New("postgres CDC: created replication slot did not select pgoutput")
	}
	barrier, err := pglogrepl.ParseLSN(created.ConsistentPoint)
	if err != nil {
		return postgresCDCSlotSetup{}, errors.New("postgres CDC: created replication slot returned an invalid consistent point")
	}
	return postgresCDCSlotSetup{barrier: barrier, created: true}, nil
}

// CDCSlotName returns the source-bound, connector-owned replication slot name
// for a stream. It is provided for lifecycle inspection and never accepts a
// caller-selected slot name.
func (c Connector) CDCSlotName(ctx context.Context, cfg connectors.RuntimeConfig, stream string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := requireRealCDCSource(cfg); err != nil {
		return "", err
	}
	conn, err := resolveConfig(cfg)
	if err != nil {
		return "", err
	}
	publication, err := cdcPublication(cfg)
	if err != nil {
		return "", err
	}
	replication, err := replicationConnection(ctx, conn)
	if err != nil {
		return "", err
	}
	defer closeReplicationConnection(replication)
	source, err := identifyCDCSource(ctx, replication, conn.database, conn.schema, stream, publication)
	if err != nil {
		return "", err
	}
	return source.slotName, nil
}

// TeardownCDC drops only the connector-derived inactive slot for stream. It
// is an explicit lifecycle operation so a persistent slot survives a normal
// process restart yet cannot be accidentally left behind after a teardown.
func (c Connector) TeardownCDC(ctx context.Context, cfg connectors.RuntimeConfig, stream string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := requireRealCDCSource(cfg); err != nil {
		return err
	}
	conn, err := resolveConfig(cfg)
	if err != nil {
		return err
	}
	publication, err := cdcPublication(cfg)
	if err != nil {
		return err
	}
	replication, err := replicationConnection(ctx, conn)
	if err != nil {
		return err
	}
	defer closeReplicationConnection(replication)
	source, err := identifyCDCSource(ctx, replication, conn.database, conn.schema, stream, publication)
	if err != nil {
		return err
	}
	slot, found, err := replicationSlotStatus(ctx, conn, source.slotName)
	if err != nil || !found {
		return err
	}
	if err := validateReplicationSlot(slot, conn.database); err != nil {
		return err
	}
	if err := pglogrepl.DropReplicationSlot(ctx, replication, source.slotName, pglogrepl.DropReplicationSlotOptions{}); err != nil {
		return classifyCDCSlotDropError(err)
	}
	return nil
}

func classifyCDCSlotDropError(err error) error {
	var server *pgconn.PgError
	if errors.As(err, &server) && server.Code == "55006" {
		return fmt.Errorf("%w: PostgreSQL refused to drop a slot that became active", ErrCDCSlotActive)
	}
	return fmt.Errorf("postgres CDC: drop replication slot: %w", err)
}
