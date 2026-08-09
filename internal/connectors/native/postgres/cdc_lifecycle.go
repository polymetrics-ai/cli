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

type postgresCDCSource struct {
	identity   synccontract.SourceIdentity
	generation synccontract.OpaqueToken
	slotName   string
	system     pglogrepl.IdentifySystemResult
}

type postgresReplicationSlot struct {
	plugin       string
	database     string
	active       bool
	confirmedLSN string
	restartLSN   string
}

func replicationConnection(ctx context.Context, conn connConfig) (*pgconn.PgConn, error) {
	config, err := conn.replicationConfig()
	if err != nil {
		return nil, fmt.Errorf("postgres CDC: configure replication connection: %w", err)
	}
	if config.RuntimeParams == nil {
		config.RuntimeParams = make(map[string]string)
	}
	config.RuntimeParams["replication"] = "database"
	replication, err := pgconn.ConnectConfig(ctx, config)
	if err != nil {
		return nil, errors.New("postgres CDC: connect replication source failed")
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
		identity:   identity,
		generation: generation,
		slotName:   cdcSlotName(identity),
		system:     system,
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
	return nil
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
