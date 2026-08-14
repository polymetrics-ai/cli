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
	cdcProtocolVersion     = "pgoutput-v2"
	cdcCheckpointSchema    = "postgres-cdc-v2"
	cdcChangefeedMechanism = "logical_replication"
	cdcExecutorID          = "postgres_logical_replication"
	cdcClientEncoding      = "UTF8"
	cdcSlotCleanupTimeout  = 5 * time.Second
)

var (
	errCDCClientEncoding         = errors.New("postgres CDC replication connection did not negotiate UTF-8 client encoding")
	errCDCReplicaIdentityDefault = errors.New("postgres CDC requires a primary key when REPLICA IDENTITY is DEFAULT")
	errCDCReplicaIdentityFull    = errors.New("postgres CDC does not support REPLICA IDENTITY FULL")
	errCDCReplicaIdentityIndex   = errors.New("postgres CDC does not support REPLICA IDENTITY USING INDEX")
	errCDCReplicaIdentityNothing = errors.New("postgres CDC does not support REPLICA IDENTITY NOTHING")
	errCDCReplicaIdentityUnknown = errors.New("postgres CDC encountered an unknown replica identity mode")
)

type postgresCDCSource struct {
	identity          synccontract.SourceIdentity
	generation        synccontract.OpaqueToken
	slotName          string
	system            pglogrepl.IdentifySystemResult
	publication       string
	schemaFingerprint string
	bootstrap         *postgresBootstrapBarrier
}

type postgresReplicationSlot struct {
	plugin       string
	database     string
	active       bool
	confirmedLSN string
	restartLSN   string
}

func replicationConnection(ctx context.Context, conn connConfig) (*pgconn.PgConn, error) {
	if err := preflightReplicationServer(ctx, conn); err != nil {
		return nil, err
	}
	config, err := conn.replicationConfig()
	if err != nil {
		return nil, fmt.Errorf("postgres CDC: configure replication connection: %w", err)
	}
	configureCDCReplicationConfig(config)
	replication, err := pgconn.ConnectConfig(ctx, config)
	if err != nil {
		return nil, errors.New("postgres CDC: connect replication source failed")
	}
	return replication, nil
}

var (
	errCDCWALLevel         = errors.New("postgres CDC requires wal_level=logical; self-managed PostgreSQL needs a restart after changing it, RDS/Aurora need rds.logical_replication=1 in a parameter group and reboot, and Cloud SQL needs cloudsql.logical_decoding enabled (Azure uses its logical-decoding equivalent)")
	errCDCReplicationSlots = errors.New("postgres CDC requires max_replication_slots > 0; increase it in the server parameter group and restart PostgreSQL")
	errCDCWALSenders       = errors.New("postgres CDC requires max_wal_senders > 0; increase it in the server parameter group and restart PostgreSQL")
	errCDCReplicationRole  = errors.New("postgres CDC requires the connecting role to have REPLICATION; grant the role REPLICATION or use a dedicated replication role")
	errCDCServerVersion    = errors.New("postgres CDC requires PostgreSQL 14 or newer for pgoutput streamed transactions")
)

func preflightReplicationServer(ctx context.Context, conn connConfig) error {
	cfg, err := conn.dataConfig()
	if err != nil {
		return fmt.Errorf("postgres CDC server preflight: %w", err)
	}
	db, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return errors.New("postgres CDC server preflight: connect failed")
	}
	defer func() { _ = db.Close(context.Background()) }()
	var walLevel string
	var serverVersion int
	var slots, senders int
	var replication bool
	if err := db.QueryRow(ctx, "SELECT current_setting('server_version_num')::int, current_setting('wal_level'), current_setting('max_replication_slots')::int, current_setting('max_wal_senders')::int, (SELECT rolreplication FROM pg_roles WHERE rolname = current_user)").Scan(&serverVersion, &walLevel, &slots, &senders, &replication); err != nil {
		return errors.New("postgres CDC server preflight: unable to inspect version, wal_level, replication slots, WAL senders, and role attributes")
	}
	if err := validateCDCServerVersion(serverVersion); err != nil {
		return err
	}
	if !strings.EqualFold(walLevel, "logical") {
		return fmt.Errorf("%w (server reports %q)", errCDCWALLevel, walLevel)
	}
	if slots <= 0 {
		return fmt.Errorf("%w (server reports %d)", errCDCReplicationSlots, slots)
	}
	if senders <= 0 {
		return fmt.Errorf("%w (server reports %d)", errCDCWALSenders, senders)
	}
	if !replication {
		return errCDCReplicationRole
	}
	return nil
}

func validateCDCServerVersion(serverVersion int) error {
	if serverVersion < 140000 {
		return fmt.Errorf("%w (server reports %d)", errCDCServerVersion, serverVersion)
	}
	return nil
}

func configureCDCReplicationConfig(config *pgconn.Config) {
	if config.RuntimeParams == nil {
		config.RuntimeParams = make(map[string]string)
	}
	config.RuntimeParams["replication"] = "database"
	config.RuntimeParams["client_encoding"] = cdcClientEncoding
	previousValidateConnect := config.ValidateConnect
	config.ValidateConnect = func(ctx context.Context, replication *pgconn.PgConn) error {
		if err := validateCDCClientEncoding(replication.ParameterStatus("client_encoding")); err != nil {
			return err
		}
		if previousValidateConnect != nil {
			return previousValidateConnect(ctx, replication)
		}
		return nil
	}
}

func validateCDCClientEncoding(encoding string) error {
	if !strings.EqualFold(strings.TrimSpace(encoding), cdcClientEncoding) {
		return errCDCClientEncoding
	}
	return nil
}

func closeReplicationConnection(conn *pgconn.PgConn) {
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
	generation := synccontract.OpaqueToken([]byte(strconv.FormatInt(int64(system.Timeline), 10) + "\n" + publication))
	return postgresCDCSource{
		identity:    identity,
		generation:  generation,
		slotName:    cdcSlotName(identity),
		system:      system,
		publication: publication,
	}, nil
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

func replicationSlotStatus(ctx context.Context, conn connConfig, slotName string) (postgresReplicationSlot, bool, error) {
	data, err := cdcDataConnection(ctx, conn)
	if err != nil {
		return postgresReplicationSlot{}, false, err
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

func cdcDataConnection(ctx context.Context, conn connConfig) (*pgx.Conn, error) {
	config, err := conn.dataConfig()
	if err != nil {
		return nil, fmt.Errorf("postgres CDC: configure data connection: %w", err)
	}
	data, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, errors.New("postgres CDC: connect source failed")
	}
	return data, nil
}

// validateCDCPublicationStream proves the pre-existing publication actually
// includes the requested relation before a WAL-retaining slot is created. A
// publication can contain several tables, so the decoder still filters every
// row frame to this same source-bound relation after startup.
func validateCDCPublicationStream(ctx context.Context, conn connConfig, source postgresCDCSource, publication string) error {
	schema, table, found := strings.Cut(source.identity.ObjectScope, ".")
	if !found || schema == "" || table == "" {
		return errors.New("postgres CDC: source identity has an invalid stream scope")
	}
	data, err := cdcDataConnection(ctx, conn)
	if err != nil {
		return err
	}
	defer func() { _ = data.Close(ctx) }()

	var included bool
	err = data.QueryRow(ctx, `
	SELECT EXISTS (
		SELECT 1
		FROM pg_publication_tables
		WHERE pubname = $1 AND schemaname = $2 AND tablename = $3
	)`, publication, schema, table).Scan(&included)
	if err != nil {
		return fmt.Errorf("postgres CDC: inspect publication membership: %w", err)
	}
	if !included {
		return errors.New("postgres CDC: publication does not include the requested stream")
	}
	var replicaIdentity string
	var hasPrimaryKey bool
	err = data.QueryRow(ctx, `
	SELECT relation.relreplident::text,
	       EXISTS (
		   SELECT 1 FROM pg_index primary_index
		   WHERE primary_index.indrelid = relation.oid AND primary_index.indisprimary
	       )
	FROM pg_class relation
	JOIN pg_namespace relation_schema ON relation_schema.oid = relation.relnamespace
	WHERE relation_schema.nspname = $1 AND relation.relname = $2`, schema, table).Scan(&replicaIdentity, &hasPrimaryKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.New("postgres CDC: requested stream relation does not exist")
	}
	if err != nil {
		return fmt.Errorf("postgres CDC: inspect replica identity: %w", err)
	}
	if err := validateCDCReplicaIdentity(replicaIdentity, hasPrimaryKey); err != nil {
		return err
	}
	return nil
}

// validateCDCReplicaIdentity admits only the pgoutput shape the current
// decoder can map without guessing. Other normal PostgreSQL modes are rejected
// before slot creation with a stable, named reason rather than silently
// misidentifying updates or deletes.
func validateCDCReplicaIdentity(replicaIdentity string, hasPrimaryKey bool) error {
	switch replicaIdentity {
	case "d":
		if !hasPrimaryKey {
			return errCDCReplicaIdentityDefault
		}
		return nil
	case "f":
		return errCDCReplicaIdentityFull
	case "i":
		return errCDCReplicaIdentityIndex
	case "n":
		return errCDCReplicaIdentityNothing
	default:
		return fmt.Errorf("%w: %q", errCDCReplicaIdentityUnknown, replicaIdentity)
	}
}

func validateReplicationSlot(slot postgresReplicationSlot, database string) error {
	if slot.plugin != "pgoutput" {
		return errors.New("postgres CDC: derived replication slot uses an incompatible output plugin")
	}
	if slot.database != database {
		return errors.New("postgres CDC: derived replication slot belongs to another database")
	}
	if slot.active {
		return errors.New("postgres CDC: derived replication slot is already active")
	}
	return nil
}

// slotRetentionLSN returns the oldest position PostgreSQL still guarantees
// this slot can replay. confirmed_flush_lsn is deliberately not used: it is
// an acknowledgement high-water mark, and using it as a recovery lower bound
// could skip a transaction whose checkpoint was not durably retained.
func slotRetentionLSN(slot postgresReplicationSlot) (pglogrepl.LSN, error) {
	if strings.TrimSpace(slot.restartLSN) == "" {
		return 0, errors.New("postgres CDC: replication slot has no usable retained position")
	}
	lsn, err := pglogrepl.ParseLSN(slot.restartLSN)
	if err != nil {
		return 0, errors.New("postgres CDC: replication slot has an invalid retained position")
	}
	return lsn, nil
}

// ensureReplicationSlot creates a connector-owned logical slot or returns
// its retained replay boundary. The bool is true only for a slot created by
// this call; callers use it to refuse an uncheckpointed existing slot rather
// than silently skipping prior source transactions.
func ensureReplicationSlot(ctx context.Context, replication *pgconn.PgConn, conn connConfig, source postgresCDCSource) (pglogrepl.LSN, bool, error) {
	slot, found, err := replicationSlotStatus(ctx, conn, source.slotName)
	if err != nil {
		return 0, false, err
	}
	if found {
		if err := validateReplicationSlot(slot, conn.database); err != nil {
			return 0, false, err
		}
		retained, err := slotRetentionLSN(slot)
		if err != nil {
			return 0, false, err
		}
		return retained, false, nil
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
			return 0, false, fmt.Errorf("postgres CDC: create replication slot: %w", createErr)
		}
		if err := validateReplicationSlot(slot, conn.database); err != nil {
			return 0, false, err
		}
		retained, err := slotRetentionLSN(slot)
		if err != nil {
			return 0, false, err
		}
		return retained, false, nil
	}
	if created.OutputPlugin != "pgoutput" {
		return 0, false, errors.New("postgres CDC: created replication slot did not select pgoutput")
	}
	barrier, err := pglogrepl.ParseLSN(created.ConsistentPoint)
	if err != nil {
		return 0, false, errors.New("postgres CDC: created replication slot returned an invalid consistent point")
	}
	return barrier, true, nil
}

// CDCSlotName returns the source-bound, connector-owned replication slot name
// for a stream. It is provided for lifecycle inspection and never accepts a
// caller-selected slot name.
func (c Connector) CDCSlotName(ctx context.Context, cfg connectors.RuntimeConfig, stream string) (string, error) {
	if err := ctx.Err(); err != nil {
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
	// Do not wait for an activity race to clear: waiting could drop a slot a
	// concurrent reader started after validation. An active slot is refused,
	// preserving deterministic cleanup without tearing down another reader.
	if err := pglogrepl.DropReplicationSlot(ctx, replication, source.slotName, pglogrepl.DropReplicationSlotOptions{Wait: false}); err != nil {
		return fmt.Errorf("postgres CDC: drop replication slot: %w", err)
	}
	return nil
}
