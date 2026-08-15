package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pglogrepl"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

const (
	cdcStageDirectory                   = "postgres-cdc-stage"
	cdcStageMaxTransactionBytes   int64 = 64 << 20
	cdcStageMaxTransactionRecords int64 = 100_000
	cdcStageMaxStagedBytes        int64 = 256 << 20
	cdcStageMaxStagedTransactions int64 = 16
	cdcStageMaxTransactionAge           = 15 * time.Minute
	cdcTransactionSink                  = "postgres_cdc_event_callback"
)

var errCDCProjectDirectory = errors.New("postgres CDC requires an owned project directory for crash-recoverable transaction staging")

type cdcV2Acknowledger func(context.Context, pglogrepl.LSN) error

type pgoutputV2TransactionMachine struct {
	stage        *database.CommittedTransactionStage
	source       postgresCDCSource
	barrier      pglogrepl.LSN
	lastDurable  pglogrepl.LSN
	req          connectors.CDCReadRequest
	emit         func(connectors.CDCEvent) error
	acknowledge  cdcV2Acknowledger
	decoder      *pgoutputDecoder
	transactions map[uint32]cdcV2Transaction
	normalXID    uint32
	streamXID    uint32
	onReceipt    func()
}

type cdcV2Transaction struct {
	id       string
	streamed bool
	replayed bool
}

func newPostgresCDCTransactionStage(projectDir string, source postgresCDCSource) (*database.CommittedTransactionStage, error) {
	root, err := cdcStageProjectRoot(projectDir)
	if err != nil {
		return nil, err
	}
	stageRoot := filepath.Join(root, "state", cdcStageDirectory, cdcStageSourceKey(source))
	stage, err := database.OpenCommittedTransactionStage(database.TransactionStageOptions{
		Root: stageRoot,
		Limits: database.TransactionStageLimits{
			MaxTransactionBytes:   cdcStageMaxTransactionBytes,
			MaxTransactionRecords: cdcStageMaxTransactionRecords,
			MaxTransactionAge:     cdcStageMaxTransactionAge,
			MaxStagedBytes:        cdcStageMaxStagedBytes,
			MaxStagedTransactions: cdcStageMaxStagedTransactions,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("postgres CDC: open transaction stage: %w", err)
	}
	if pending := stage.PendingTransactions(); len(pending) != 0 {
		return nil, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "PostgreSQL CDC found receipt-less staged transactions; explicit rebootstrap is required")
	}
	return stage, nil
}

func cdcStageProjectRoot(projectDir string) (string, error) {
	if strings.TrimSpace(projectDir) == "" {
		return "", errCDCProjectDirectory
	}
	root, err := filepath.Abs(projectDir)
	if err != nil {
		return "", fmt.Errorf("postgres CDC: resolve project staging root: %w", err)
	}
	if filepath.Clean(root) == string(filepath.Separator) {
		return "", errCDCProjectDirectory
	}
	return root, nil
}

func cdcStageSourceKey(source postgresCDCSource) string {
	sum := sha256.Sum256([]byte(source.identity.Engine + "\x00" + source.identity.AccountOrCluster + "\x00" + source.identity.ObjectScope))
	return hex.EncodeToString(sum[:])
}

func newPGOutputV2TransactionMachine(stage *database.CommittedTransactionStage, source postgresCDCSource, barrier, start pglogrepl.LSN, req connectors.CDCReadRequest, emit func(connectors.CDCEvent) error, acknowledge cdcV2Acknowledger, onReceipt func()) *pgoutputV2TransactionMachine {
	return &pgoutputV2TransactionMachine{
		stage:        stage,
		source:       source,
		barrier:      barrier,
		lastDurable:  start,
		req:          req,
		emit:         emit,
		acknowledge:  acknowledge,
		decoder:      newPGOutputDecoderForRelation(source.identity.ObjectScope),
		transactions: make(map[uint32]cdcV2Transaction),
		onReceipt:    onReceipt,
	}
}

func (m *pgoutputV2TransactionMachine) Handle(ctx context.Context, frame []byte, walStart pglogrepl.LSN) error {
	if m == nil || m.stage == nil || m.emit == nil || m.acknowledge == nil {
		return errors.New("postgres CDC v2 transaction machine is unavailable")
	}
	if len(frame) == 0 {
		return errors.New("postgres CDC: received empty pgoutput v2 frame")
	}
	inStream := m.streamXID != 0
	logical, err := pglogrepl.ParseV2(frame, inStream)
	if err != nil {
		return fmt.Errorf("postgres CDC: parse pgoutput v2 message: %w", err)
	}

	switch message := logical.(type) {
	case *pglogrepl.BeginMessage:
		if inStream || m.normalXID != 0 || len(m.transactions) != 0 {
			return errors.New("postgres CDC: begin arrived while another transaction is active")
		}
		if err := m.begin(ctx, message.Xid, message.FinalLSN, false); err != nil {
			return err
		}
		m.normalXID = message.Xid
		return nil
	case *pglogrepl.StreamStartMessageV2:
		if inStream {
			return errors.New("postgres CDC: nested pgoutput stream start")
		}
		if message.FirstSegment != 0 && message.FirstSegment != 1 {
			return errors.New("postgres CDC: stream start has an invalid first-segment marker")
		}
		if message.FirstSegment == 1 {
			if err := m.begin(ctx, message.Xid, walStart, true); err != nil {
				return err
			}
		} else if transaction, ok := m.transactions[message.Xid]; !ok || !transaction.streamed {
			return errors.New("postgres CDC: continuation stream segment has no staged transaction")
		}
		m.streamXID = message.Xid
		return nil
	case *pglogrepl.StreamStopMessageV2:
		if !inStream {
			return errors.New("postgres CDC: stream stop arrived outside a streamed transaction")
		}
		m.streamXID = 0
		return nil
	case *pglogrepl.StreamAbortMessageV2:
		if inStream {
			return errors.New("postgres CDC: stream abort arrived before stream stop")
		}
		return m.abort(ctx, message.Xid)
	case *pglogrepl.StreamCommitMessageV2:
		if inStream {
			return errors.New("postgres CDC: stream commit arrived before stream stop")
		}
		return m.commit(ctx, message.Xid, message.CommitLSN, message.TransactionEndLSN)
	case *pglogrepl.CommitMessage:
		if inStream {
			return errors.New("postgres CDC: non-stream commit arrived inside a stream")
		}
		if m.normalXID == 0 {
			return errors.New("postgres CDC: commit without a pgoutput transaction")
		}
		return m.commit(ctx, m.normalXID, message.CommitLSN, message.TransactionEndLSN)
	case *pglogrepl.RelationMessageV2:
		return m.decodeMetadata(frame, inStream, walStart)
	case *pglogrepl.TypeMessageV2:
		return m.decodeMetadata(frame, inStream, walStart)
	case *pglogrepl.OriginMessage:
		return m.decodeMetadata(frame, false, walStart)
	case *pglogrepl.LogicalDecodingMessageV2:
		return m.validateStreamFrameXID(frame, inStream)
	case *pglogrepl.InsertMessageV2, *pglogrepl.UpdateMessageV2, *pglogrepl.DeleteMessageV2:
		return m.stageEvents(ctx, frame, inStream, walStart)
	case *pglogrepl.TruncateMessageV2:
		if err := m.validateStreamFrameXID(frame, inStream); err != nil {
			return err
		}
		transactionID, err := m.activeTransactionID(inStream)
		if err != nil {
			return err
		}
		events, err := m.decoder.truncate(message.RelationIDs, walStart.String())
		if err != nil {
			return fmt.Errorf("postgres CDC: decode pgoutput truncate: %w", err)
		}
		return m.appendEvents(ctx, transactionID, events)
	default:
		return fmt.Errorf("postgres CDC: unsupported pgoutput v2 frame %T", logical)
	}
}

func (m *pgoutputV2TransactionMachine) begin(ctx context.Context, xid uint32, transactionLSN pglogrepl.LSN, streamed bool) error {
	if _, exists := m.transactions[xid]; exists {
		return errors.New("postgres CDC: duplicate transaction start")
	}
	if transactionLSN == 0 {
		return errors.New("postgres CDC: transaction start is missing its WAL position")
	}
	transaction := cdcV2Transaction{id: cdcTransactionID(m.source, xid, transactionLSN), streamed: streamed}
	err := m.stage.BeginTransaction(ctx, transaction.id)
	if errors.Is(err, database.ErrTransactionStageAlreadyCommitted) {
		transaction.replayed = true
	} else if err != nil {
		return fmt.Errorf("postgres CDC: begin streamed transaction stage: %w", err)
	}
	m.transactions[xid] = transaction
	return nil
}

func (m *pgoutputV2TransactionMachine) abort(ctx context.Context, xid uint32) error {
	transaction, exists := m.transactions[xid]
	if !exists || !transaction.streamed {
		return errors.New("postgres CDC: stream abort has no staged transaction")
	}
	delete(m.transactions, xid)
	if transaction.replayed {
		return errors.New("postgres CDC: stream abort conflicts with an already durable receipt")
	}
	if err := m.stage.AbortTransaction(ctx, transaction.id); err != nil {
		return fmt.Errorf("postgres CDC: abort streamed transaction stage: %w", err)
	}
	return nil
}

func (m *pgoutputV2TransactionMachine) commit(ctx context.Context, xid uint32, commitLSN, endLSN pglogrepl.LSN) error {
	transaction, exists := m.transactions[xid]
	if !exists {
		return errors.New("postgres CDC: commit has no staged transaction")
	}
	if endLSN == 0 || commitLSN == 0 {
		return errors.New("postgres CDC: commit is missing its LSN")
	}
	if !transaction.replayed {
		receiver := postgresCDCTransactionReceiver{emit: m.emit}
		if _, err := m.stage.CommitTransaction(ctx, transaction.id, receiver); err != nil {
			return fmt.Errorf("postgres CDC: receive durable transaction: %w", err)
		}
		if _, err := m.stage.Receipt(transaction.id); err != nil {
			return fmt.Errorf("postgres CDC: verify durable transaction receipt: %w", err)
		}
		if m.onReceipt != nil {
			m.onReceipt()
		}
	}

	candidate := postgresCDCCheckpointForLSNs(m.source, m.barrier, m.lastDurable, commitLSN, endLSN)
	if err := m.req.DurableCheckpointCommitter.CommitDurableChangefeedCheckpoint(ctx, candidate); err != nil {
		return fmt.Errorf("postgres CDC: persist durable checkpoint: %w", err)
	}
	if err := m.acknowledge(ctx, endLSN); err != nil {
		return err
	}
	m.lastDurable = endLSN
	delete(m.transactions, xid)
	if m.normalXID == xid {
		m.normalXID = 0
	}
	return nil
}

func (m *pgoutputV2TransactionMachine) decodeMetadata(frame []byte, inStream bool, walStart pglogrepl.LSN) error {
	if err := m.validateStreamFrameXID(frame, inStream); err != nil {
		return err
	}
	payload, err := pgoutputV2Payload(frame, inStream)
	if err != nil {
		return err
	}
	if _, err := m.decoder.decode(payload, walStart.String()); err != nil {
		return fmt.Errorf("postgres CDC: decode pgoutput metadata: %w", err)
	}
	return nil
}

func (m *pgoutputV2TransactionMachine) stageEvents(ctx context.Context, frame []byte, inStream bool, walStart pglogrepl.LSN) error {
	if err := m.validateStreamFrameXID(frame, inStream); err != nil {
		return err
	}
	transactionID, err := m.activeTransactionID(inStream)
	if err != nil {
		return err
	}
	payload, err := pgoutputV2Payload(frame, inStream)
	if err != nil {
		return err
	}
	events, err := m.decoder.decode(payload, walStart.String())
	if err != nil {
		return fmt.Errorf("postgres CDC: decode pgoutput DML: %w", err)
	}
	return m.appendEvents(ctx, transactionID, events)
}

func (m *pgoutputV2TransactionMachine) validateStreamFrameXID(frame []byte, inStream bool) error {
	if !inStream {
		return nil
	}
	if len(frame) < 5 {
		return errors.New("postgres CDC: streamed pgoutput frame is missing its transaction ID")
	}
	if got := binary.BigEndian.Uint32(frame[1:5]); got != m.streamXID {
		return errors.New("postgres CDC: streamed pgoutput frame transaction ID does not match the active stream")
	}
	return nil
}

func (m *pgoutputV2TransactionMachine) activeTransactionID(inStream bool) (string, error) {
	if inStream {
		transaction, exists := m.transactions[m.streamXID]
		if !exists || !transaction.streamed {
			return "", errors.New("postgres CDC: streamed DML has no active transaction")
		}
		return transaction.id, nil
	}
	if m.normalXID == 0 {
		return "", errors.New("postgres CDC: DML arrived outside a transaction")
	}
	transaction, exists := m.transactions[m.normalXID]
	if !exists || transaction.streamed {
		return "", errors.New("postgres CDC: non-stream DML has no active transaction")
	}
	return transaction.id, nil
}

func (m *pgoutputV2TransactionMachine) appendEvents(ctx context.Context, transactionID string, events []connectors.CDCEvent) error {
	if len(events) == 0 {
		return nil
	}
	transaction, exists := m.transactionForID(transactionID)
	if !exists {
		return errors.New("postgres CDC: transaction stage is unavailable")
	}
	if transaction.replayed {
		return nil
	}
	var payload bytes.Buffer
	if err := gob.NewEncoder(&payload).Encode(events); err != nil {
		return fmt.Errorf("postgres CDC: encode staged events: %w", err)
	}
	if err := m.stage.AppendChunk(ctx, transactionID, int64(len(events)), &payload); err != nil {
		return fmt.Errorf("postgres CDC: append streamed transaction chunk: %w", err)
	}
	return nil
}

func (m *pgoutputV2TransactionMachine) transactionForID(transactionID string) (cdcV2Transaction, bool) {
	for _, transaction := range m.transactions {
		if transaction.id == transactionID {
			return transaction, true
		}
	}
	return cdcV2Transaction{}, false
}

func pgoutputV2Payload(frame []byte, inStream bool) ([]byte, error) {
	if len(frame) == 0 {
		return nil, errors.New("postgres CDC: pgoutput v2 frame is empty")
	}
	if !inStream {
		return frame, nil
	}
	switch frame[0] {
	case 'M', 'R', 'Y', 'I', 'U', 'D', 'T':
		if len(frame) < 5 {
			return nil, errors.New("postgres CDC: streamed pgoutput frame is missing its transaction ID")
		}
		payload := make([]byte, 1, len(frame)-4)
		payload[0] = frame[0]
		return append(payload, frame[5:]...), nil
	default:
		return frame, nil
	}
}

func cdcTransactionID(source postgresCDCSource, xid uint32, transactionLSN pglogrepl.LSN) string {
	return source.identity.Engine + "\x00" + source.identity.AccountOrCluster + "\x00" + source.identity.ObjectScope + "\x00" + fmt.Sprintf("%d", xid) + "\x00" + transactionLSN.String()
}

type postgresCDCTransactionReceiver struct {
	emit func(connectors.CDCEvent) error
}

func (r postgresCDCTransactionReceiver) ReceiveCommittedTransaction(ctx context.Context, transaction database.CommittedTransaction) (database.DownstreamTransactionReceipt, error) {
	if r.emit == nil {
		return database.DownstreamTransactionReceipt{}, errors.New("postgres CDC event callback is required")
	}
	var observed int64
	if err := transaction.StreamChunks(ctx, func(chunk database.TransactionChunk) error {
		var events []connectors.CDCEvent
		decoder := gob.NewDecoder(chunk.Reader)
		if err := decoder.Decode(&events); err != nil {
			return fmt.Errorf("decode staged transaction chunk: %w", err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			if err == nil {
				return errors.New("staged transaction chunk contains trailing values")
			}
			return fmt.Errorf("decode staged transaction chunk trailing data: %w", err)
		}
		if int64(len(events)) != chunk.Records {
			return errors.New("staged transaction chunk record count does not match its contents")
		}
		for _, event := range events {
			if err := r.emit(event); err != nil {
				return err
			}
			observed++
		}
		return nil
	}); err != nil {
		return database.DownstreamTransactionReceipt{}, err
	}
	if observed != transaction.Records {
		return database.DownstreamTransactionReceipt{}, errors.New("staged transaction record count does not match its receipt")
	}
	return database.DownstreamTransactionReceipt{
		ReceiptID: "postgres-cdc-" + transaction.ContentDigest,
		Sink:      cdcTransactionSink,
		DurableAt: time.Now().UTC(),
	}, nil
}
