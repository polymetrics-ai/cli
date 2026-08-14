package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

// BootstrapCDCRequest is the PostgreSQL-only handover contract. Snapshot is
// called for bounded pages from the imported exported snapshot and must return
// only after its connection-owned materialization is durable. The coordinator
// has no target handle, raw SQL, or direct connector-to-connector route.
type BootstrapCDCRequest struct {
	Stream                     string
	Config                     connectors.RuntimeConfig
	BatchSize                  int
	DurableCheckpointCommitter connectors.DurableChangefeedCheckpointCommitter
	Snapshot                   func(context.Context, BootstrapSnapshotPage) error
}

// BootstrapSnapshotPage describes a bounded page from the exact PostgreSQL
// snapshot that the logical slot's initial LSN protects. The records and
// opaque barrier token are defensive copies owned by the receiver.
type BootstrapSnapshotPage struct {
	Records           []connectors.Record
	Source            synccontract.SourceIdentity
	SnapshotBarrier   synccontract.SnapshotBarrier
	SchemaFingerprint string
	// CandidateCheckpoint is the same source-owned initial barrier on every
	// page. A connection warehouse must durably stage every page before the
	// coordinator lets this checkpoint become resumable.
	CandidateCheckpoint synccontract.CheckpointEnvelope
	Final               bool
}

// BootstrapCDC joins one exported PostgreSQL snapshot to pgoutput-v2 at the
// slot's consistent point. Changes written after that point are retained by
// the slot while Snapshot materializes the pre-change state. The normal v2
// transaction machine then preserves its existing receipt -> checkpoint ->
// acknowledgement ordering for every post-barrier transaction.
func (c Connector) BootstrapCDC(ctx context.Context, request BootstrapCDCRequest, emit func(connectors.CDCEvent) error) error {
	if err := validateBootstrapCDCRequest(ctx, request, emit); err != nil {
		return err
	}
	conn, err := resolveConfig(request.Config)
	if err != nil {
		return err
	}
	publication, err := cdcPublication(request.Config)
	if err != nil {
		return err
	}
	replication, err := replicationConnection(ctx, conn)
	if err != nil {
		return err
	}
	defer closeReplicationConnection(replication)

	source, err := identifyCDCSource(ctx, replication, conn.database, conn.schema, request.Stream, publication)
	if err != nil {
		return err
	}
	if err := validateCDCPublicationStream(ctx, conn, source, publication); err != nil {
		return err
	}
	stage, err := newPostgresCDCTransactionStage(request.Config.ProjectDir, source)
	if err != nil {
		return err
	}
	barrier, exportedSnapshot, err := createPostgresBootstrapSlot(ctx, replication, conn, source)
	if err != nil {
		return err
	}

	source, initial, err := c.readBootstrapSnapshot(ctx, conn, source, barrier, exportedSnapshot, request)
	if err != nil {
		return err
	}
	if err := request.DurableCheckpointCommitter.CommitDurableChangefeedCheckpoint(ctx, initial); err != nil {
		return fmt.Errorf("postgres bootstrap: persist durable snapshot barrier: %w", err)
	}
	if err := startPostgresCDCReplication(ctx, replication, source.slotName, barrier, publication, &initial); err != nil {
		return err
	}
	if err := sendStandbyStatus(ctx, replication, barrier); err != nil {
		return err
	}
	return consumePGOutputV2LogicalReplication(ctx, replication, stage, source, barrier, barrier, connectors.CDCReadRequest{
		Stream:                     request.Stream,
		Config:                     request.Config,
		DurableCheckpointCommitter: request.DurableCheckpointCommitter,
	}, emit)
}

func validateBootstrapCDCRequest(ctx context.Context, request BootstrapCDCRequest, emit func(connectors.CDCEvent) error) error {
	if ctx == nil {
		return errors.New("postgres bootstrap requires a context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(request.Config) {
		return errors.New("postgres bootstrap requires a real PostgreSQL source; fixture mode is not a replication protocol")
	}
	if _, err := cdcStageProjectRoot(request.Config.ProjectDir); err != nil {
		return err
	}
	if strings.TrimSpace(request.Stream) == "" {
		return errors.New("postgres bootstrap requires a stream (schema.table)")
	}
	if request.BatchSize <= 0 {
		return errors.New("postgres bootstrap requires a positive batch size")
	}
	if request.Snapshot == nil {
		return errors.New("postgres bootstrap requires a durable snapshot receiver")
	}
	if emit == nil {
		return errors.New("postgres bootstrap requires a changefeed event callback")
	}
	if request.DurableCheckpointCommitter == nil {
		return errors.New("postgres bootstrap requires a durable checkpoint committer")
	}
	return nil
}

func createPostgresBootstrapSlot(ctx context.Context, replication *pgconn.PgConn, conn connConfig, source postgresCDCSource) (pglogrepl.LSN, string, error) {
	slot, found, err := replicationSlotStatus(ctx, conn, source.slotName)
	if err != nil {
		return 0, "", err
	}
	if found {
		if err := validateReplicationSlot(slot, conn.database); err != nil {
			return 0, "", err
		}
		return 0, "", synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "existing PostgreSQL replication slot requires a durable bootstrap checkpoint or explicit teardown")
	}
	created, err := pglogrepl.CreateReplicationSlot(ctx, replication, source.slotName, "pgoutput", pglogrepl.CreateReplicationSlotOptions{
		Mode:           pglogrepl.LogicalReplication,
		SnapshotAction: "EXPORT_SNAPSHOT",
	})
	if err != nil {
		return 0, "", fmt.Errorf("postgres bootstrap: create exported replication slot: %w", err)
	}
	if created.OutputPlugin != "pgoutput" {
		return 0, "", errors.New("postgres bootstrap: created replication slot did not select pgoutput")
	}
	barrier, err := pglogrepl.ParseLSN(created.ConsistentPoint)
	if err != nil || barrier == 0 {
		return 0, "", errors.New("postgres bootstrap: replication slot returned an invalid consistent point")
	}
	if !postgresExportedSnapshotName(created.SnapshotName) {
		return 0, "", errors.New("postgres bootstrap: replication slot returned an invalid exported snapshot name")
	}
	return barrier, created.SnapshotName, nil
}

func (c Connector) readBootstrapSnapshot(ctx context.Context, conn connConfig, source postgresCDCSource, barrier pglogrepl.LSN, exportedSnapshot string, request BootstrapCDCRequest) (result postgresCDCSource, initial synccontract.CheckpointEnvelope, err error) {
	resources, err := newTypedCatalogResources(c.databaseDefinition.Resources())
	if err != nil {
		return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, err
	}
	operationCtx, cancel, err := resources.operationContext(ctx)
	if err != nil {
		return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, err
	}
	defer cancel()

	pool, err := conn.openTypedCatalogPool(operationCtx, resources)
	if err != nil {
		return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, fmt.Errorf("postgres bootstrap: open snapshot pool: %w", err)
	}
	defer pool.Close()
	tx, err := pool.BeginTx(operationCtx, typedCatalogTransactionOptions())
	if err != nil {
		return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, fmt.Errorf("postgres bootstrap: begin imported snapshot: %w", err)
	}
	defer func() {
		rollbackErr := rollbackTypedCatalogSnapshot(operationCtx, tx, resources)
		if rollbackErr == nil {
			return
		}
		if err == nil {
			err = rollbackErr
			return
		}
		err = errors.Join(err, rollbackErr)
	}()
	if err := importPostgresSnapshot(operationCtx, tx, exportedSnapshot); err != nil {
		return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, err
	}
	relation, err := postgresSnapshotRelationRef(conn.database + "." + source.identity.ObjectScope)
	if err != nil {
		return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, errors.New("postgres bootstrap: source relation is invalid")
	}
	catalog, err := discoverTypedCatalogSnapshot(operationCtx, tx, conn.database, conn.schema, c.databaseDefinition)
	if err != nil {
		return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, fmt.Errorf("postgres bootstrap: discover snapshot catalog: %w", err)
	}
	pageSize, err := c.databaseDefinition.Resources().EffectivePageSize(request.BatchSize)
	if err != nil {
		return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, errors.New("postgres bootstrap: batch size is outside the declared resource bound")
	}
	plan, err := newPostgresSnapshotReadPlan(catalog, relation, pageSize)
	if err != nil {
		return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, err
	}
	bootstrap, err := newPostgresBootstrapBarrier(postgresCDCSource{
		identity:          source.identity,
		generation:        source.generation,
		slotName:          source.slotName,
		system:            source.system,
		publication:       source.publication,
		schemaFingerprint: plan.fingerprint.String(),
	}, barrier)
	if err != nil {
		return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, err
	}
	source.schemaFingerprint = plan.fingerprint.String()
	source.bootstrap = &bootstrap
	initial = postgresCDCCheckpointForLSNs(source, barrier, barrier, barrier, barrier)

	for page, after := 0, []any(nil); ; page++ {
		records, nextAfter, err := plan.readPage(operationCtx, tx, after)
		if err != nil {
			return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, err
		}
		final := len(records) < plan.pageSize
		if err := request.Snapshot(operationCtx, BootstrapSnapshotPage{
			Records:             cloneBootstrapRecords(records),
			Source:              source.identity,
			SnapshotBarrier:     synccontract.SnapshotBarrier{Kind: cdcSnapshotBarrierKind, Token: bootstrap.token()},
			SchemaFingerprint:   source.schemaFingerprint,
			CandidateCheckpoint: initial.Clone(),
			Final:               final,
		}); err != nil {
			return postgresCDCSource{}, synccontract.CheckpointEnvelope{}, fmt.Errorf("postgres bootstrap: persist snapshot page %d: %w", page, err)
		}
		if final {
			return source, initial, nil
		}
		after = nextAfter
	}
}

func importPostgresSnapshot(ctx context.Context, tx pgx.Tx, exportedSnapshot string) error {
	if !postgresExportedSnapshotName(exportedSnapshot) {
		return errors.New("postgres bootstrap: exported snapshot name is invalid")
	}
	if _, err := tx.Exec(ctx, "SET TRANSACTION SNAPSHOT '"+exportedSnapshot+"'"); err != nil {
		return errors.New("postgres bootstrap: import exported snapshot failed")
	}
	return nil
}

func postgresExportedSnapshotName(name string) bool {
	if len(name) == 0 || len(name) > 128 {
		return false
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

func cloneBootstrapRecords(records []connectors.Record) []connectors.Record {
	clones := make([]connectors.Record, len(records))
	for index, record := range records {
		clone := make(connectors.Record, len(record))
		for key, value := range record {
			clone[key] = cloneBootstrapValue(value)
		}
		clones[index] = clone
	}
	return clones
}

func cloneBootstrapValue(value any) any {
	switch typed := value.(type) {
	case []byte:
		return append([]byte(nil), typed...)
	case []any:
		clone := make([]any, len(typed))
		for index := range typed {
			clone[index] = cloneBootstrapValue(typed[index])
		}
		return clone
	case map[string]any:
		clone := make(map[string]any, len(typed))
		for key, nested := range typed {
			clone[key] = cloneBootstrapValue(nested)
		}
		return clone
	default:
		return typed
	}
}

func startPostgresCDCReplication(ctx context.Context, replication *pgconn.PgConn, slot string, start pglogrepl.LSN, publication string, checkpoint *synccontract.CheckpointEnvelope) error {
	if err := pglogrepl.StartReplication(ctx, replication, slot, start, pglogrepl.StartReplicationOptions{
		Mode: pglogrepl.LogicalReplication,
		PluginArgs: []string{
			"proto_version '2'",
			"streaming 'on'",
			"publication_names '" + publication + "'",
		},
	}); err != nil {
		return classifyCDCStartError(checkpoint, err)
	}
	return nil
}

func postgresCDCSchemaFingerprint(ctx context.Context, conn connConfig, source postgresCDCSource, definition database.Definition) (string, error) {
	resources, err := newTypedCatalogResources(definition.Resources())
	if err != nil {
		return "", err
	}
	operationCtx, cancel, err := resources.operationContext(ctx)
	if err != nil {
		return "", err
	}
	defer cancel()
	pool, err := conn.openTypedCatalogPool(operationCtx, resources)
	if err != nil {
		return "", errors.New("postgres CDC: open schema fingerprint pool failed")
	}
	defer pool.Close()
	catalog, err := discoverTypedCatalog(operationCtx, pool, conn.database, conn.schema, definition, resources)
	if err != nil {
		return "", fmt.Errorf("postgres CDC: discover schema fingerprint: %w", err)
	}
	relation, err := postgresSnapshotRelationRef(conn.database + "." + source.identity.ObjectScope)
	if err != nil {
		return "", errors.New("postgres CDC: source relation is invalid")
	}
	plan, err := newPostgresSnapshotReadPlan(catalog, relation, 1)
	if err != nil {
		return "", err
	}
	return plan.fingerprint.String(), nil
}
