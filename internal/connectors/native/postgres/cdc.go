package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

var _ connectors.ChangefeedExecutor = Connector{}

// ChangefeedExecutorDescriptor is the runtime half of the PostgreSQL bundle's
// logical-replication declaration. The existing capability projection compares
// it exactly with changefeed.json before advertising CDC.
func (c Connector) ChangefeedExecutorDescriptor() connectors.ChangefeedExecutorDescriptor {
	return connectors.ChangefeedExecutorDescriptor{
		Status:    connectors.ChangefeedStatusImplemented,
		Mechanism: connectors.ChangefeedMechanismLogicalReplication,
		Executor:  connectors.ChangefeedExecutorRef{Kind: "native", ID: cdcExecutorID},
		Checkpoint: connectors.ChangefeedCheckpoint{
			Kind:        "lsn",
			Keys:        []string{"lsn"},
			CommitAfter: "downstream_ack",
			OnInvalid:   "resnapshot_required",
		},
	}
}

// ReadCDC consumes one PostgreSQL logical-replication stream until ctx is
// cancelled. Every transaction's events must be accepted before its full,
// source-bound checkpoint is durably committed; only then does the connector
// acknowledge the LSN to PostgreSQL. That order intentionally yields
// at-least-once replay instead of a skipped transaction after a crash.
func (c Connector) ReadCDC(ctx context.Context, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(req.Config) {
		return errors.New("postgres CDC requires a real PostgreSQL source; fixture mode is not a replication protocol")
	}
	if emit == nil {
		return errors.New("postgres CDC requires an event callback")
	}
	if req.DurableCheckpointCommitter == nil {
		return errors.New("postgres CDC requires a durable checkpoint committer")
	}
	if len(req.State) != 0 {
		return errors.New("postgres CDC rejects legacy scalar state; supply a durable checkpoint envelope")
	}
	if strings.TrimSpace(req.Stream) == "" {
		return errors.New("postgres CDC requires a stream (schema.table)")
	}

	conn, err := resolveConfig(req.Config)
	if err != nil {
		return err
	}
	publication, err := cdcPublication(req.Config)
	if err != nil {
		return err
	}
	replication, err := replicationConnection(ctx, conn)
	if err != nil {
		return err
	}
	defer closeReplicationConnection(replication)

	source, err := identifyCDCSource(ctx, replication, conn.database, conn.schema, req.Stream, publication)
	if err != nil {
		return err
	}
	if err := validateCDCResume(req.Checkpoint, source); err != nil {
		return err
	}
	barrier, err := ensureReplicationSlot(ctx, replication, conn, source)
	if err != nil {
		return err
	}
	start, err := cdcStartLSN(req.Checkpoint, barrier)
	if err != nil {
		return err
	}

	if err := pglogrepl.StartReplication(ctx, replication, source.slotName, start, pglogrepl.StartReplicationOptions{
		Mode: pglogrepl.LogicalReplication,
		PluginArgs: []string{
			"proto_version '1'",
			"publication_names '" + publication + "'",
		},
	}); err != nil {
		return classifyCDCStartError(req.Checkpoint, err)
	}

	return consumeLogicalReplication(ctx, replication, source, barrier, req, emit)
}

func classifyCDCStartError(checkpoint *synccontract.CheckpointEnvelope, err error) error {
	var server *pgconn.PgError
	if checkpoint != nil && errors.As(err, &server) {
		switch server.Code {
		case "58P01": // undefined_file: requested WAL is no longer retained.
			return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeRetentionGap, "PostgreSQL no longer retains the requested replication position")
		case "55000": // object_not_in_prerequisite_state: invalid replication slot.
			return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidatedSlot, "PostgreSQL rejected the persisted replication slot")
		}
	}
	return fmt.Errorf("postgres CDC: start logical replication: %w", err)
}

func validateCDCResume(checkpoint *synccontract.CheckpointEnvelope, source postgresCDCSource) error {
	if checkpoint == nil {
		return nil
	}
	if checkpoint.Mechanism != cdcChangefeedMechanism {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "checkpoint mechanism is not PostgreSQL logical replication")
	}
	return checkpoint.ValidateResume(synccontract.ResumeExpectation{
		Source:           source.identity,
		SourceGeneration: source.generation,
	})
}

func cdcStartLSN(checkpoint *synccontract.CheckpointEnvelope, barrier pglogrepl.LSN) (pglogrepl.LSN, error) {
	if checkpoint == nil {
		return barrier, nil
	}
	if len(checkpoint.Position.Primary) == 0 {
		return 0, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "checkpoint LSN is missing")
	}
	lsn, err := pglogrepl.ParseLSN(string(checkpoint.Position.Primary))
	if err != nil || lsn < barrier {
		return 0, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "checkpoint LSN is not retained by the derived replication slot")
	}
	return lsn, nil
}

func consumeLogicalReplication(ctx context.Context, replication *pgconn.PgConn, source postgresCDCSource, barrier pglogrepl.LSN, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	decoder := newPGOutputDecoder()
	lastDurable := pglogrepl.LSN(0)
	inTransaction := false

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		raw, err := replication.ReceiveMessage(ctx)
		if err != nil {
			return fmt.Errorf("postgres CDC: receive replication message: %w", err)
		}
		copyData, ok := raw.(*pgproto3.CopyData)
		if !ok {
			if serverErr, ok := raw.(*pgproto3.ErrorResponse); ok {
				return fmt.Errorf("postgres CDC: replication server error: %s", serverErr.Message)
			}
			return fmt.Errorf("postgres CDC: unexpected replication message %T", raw)
		}
		if len(copyData.Data) == 0 {
			return errors.New("postgres CDC: received empty replication copy data")
		}

		switch copyData.Data[0] {
		case pglogrepl.PrimaryKeepaliveMessageByteID:
			keepalive, err := pglogrepl.ParsePrimaryKeepaliveMessage(copyData.Data[1:])
			if err != nil {
				return fmt.Errorf("postgres CDC: parse primary keepalive: %w", err)
			}
			if keepalive.ReplyRequested {
				if err := sendStandbyStatus(ctx, replication, lastDurable); err != nil {
					return err
				}
			}
		case pglogrepl.XLogDataByteID:
			xlog, err := pglogrepl.ParseXLogData(copyData.Data[1:])
			if err != nil {
				return fmt.Errorf("postgres CDC: parse xlog data: %w", err)
			}
			logical, err := pglogrepl.Parse(xlog.WALData)
			if err != nil {
				return fmt.Errorf("postgres CDC: parse pgoutput message: %w", err)
			}
			switch message := logical.(type) {
			case *pglogrepl.BeginMessage:
				if inTransaction {
					return errors.New("postgres CDC: nested pgoutput transaction")
				}
				inTransaction = true
			case *pglogrepl.CommitMessage:
				if !inTransaction {
					return errors.New("postgres CDC: commit without a pgoutput transaction")
				}
				candidate := postgresCDCCheckpoint(source, barrier, lastDurable, message)
				if err := req.DurableCheckpointCommitter.CommitDurableChangefeedCheckpoint(ctx, candidate); err != nil {
					return fmt.Errorf("postgres CDC: persist durable checkpoint: %w", err)
				}
				if err := sendStandbyStatus(ctx, replication, message.TransactionEndLSN); err != nil {
					return err
				}
				lastDurable = message.TransactionEndLSN
				inTransaction = false
			default:
				if !inTransaction {
					return errors.New("postgres CDC: pgoutput data arrived outside a transaction")
				}
				events, err := decoder.decode(xlog.WALData, "")
				if err != nil {
					return fmt.Errorf("postgres CDC: decode pgoutput: %w", err)
				}
				for _, event := range events {
					if err := emit(event); err != nil {
						return err
					}
				}
			}
		default:
			return fmt.Errorf("postgres CDC: unsupported replication copy data type %q", copyData.Data[0])
		}
	}
}

func postgresCDCCheckpoint(source postgresCDCSource, barrier, previous pglogrepl.LSN, message *pglogrepl.CommitMessage) synccontract.CheckpointEnvelope {
	start := barrier
	if previous != 0 {
		start = previous
	}
	position := message.TransactionEndLSN.String()
	return synccontract.CheckpointEnvelope{
		StateVersion: synccontract.StateVersion,
		Source:       source.identity,
		Mechanism:    cdcChangefeedMechanism,
		SnapshotBarrier: &synccontract.SnapshotBarrier{
			Kind:  cdcSnapshotBarrierKind,
			Token: synccontract.OpaqueToken([]byte(barrier.String())),
		},
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken([]byte(position)),
			TieBreaker: synccontract.OpaqueToken([]byte(message.CommitLSN.String())),
		},
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: append(synccontract.OpaqueToken(nil), source.generation...),
		SchemaVersion:    cdcCheckpointSchema,
		ProtocolVersion:  cdcProtocolVersion,
		Dedupe: synccontract.DedupeIdentity{
			Kind:  "postgres_transaction_end_lsn",
			Value: synccontract.OpaqueToken([]byte(position)),
		},
		DedupeWindow: synccontract.DedupeWindow{
			Kind:  "postgres_lsn_interval",
			Start: synccontract.OpaqueToken([]byte(start.String())),
			End:   synccontract.OpaqueToken([]byte(position)),
		},
		ObservedAt: time.Now().UTC(),
	}
}

func sendStandbyStatus(ctx context.Context, replication *pgconn.PgConn, position pglogrepl.LSN) error {
	if err := pglogrepl.SendStandbyStatusUpdate(ctx, replication, pglogrepl.StandbyStatusUpdate{
		WALWritePosition: position,
	}); err != nil {
		return fmt.Errorf("postgres CDC: acknowledge durable LSN: %w", err)
	}
	return nil
}
