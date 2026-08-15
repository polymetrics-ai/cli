package postgres

import (
	"context"
	"encoding/binary"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pglogrepl"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

func TestPGOutputV2StreamAbortPublishesCheckpointsAndAcknowledgesNothing(t *testing.T) {
	t.Parallel()

	probe := newCDCv2Probe()
	machine := newTestCDCv2Machine(t, probe)
	const xid = uint32(71)
	if err := machine.Handle(context.Background(), streamStartFrame(xid, true), pglogrepl.LSN(0x10)); err != nil {
		t.Fatalf("Handle(StreamStart) error = %v", err)
	}
	if err := machine.Handle(context.Background(), originMessage(0x01, "upstream"), pglogrepl.LSN(0x10)); err != nil {
		t.Fatalf("Handle(Origin) error = %v", err)
	}
	if err := machine.Handle(context.Background(), streamRelationFrame(xid, relationMessage(testRelationID, "public", "users", testColumn{name: "id", typeID: 23})), pglogrepl.LSN(0x11)); err != nil {
		t.Fatalf("Handle(Relation) error = %v", err)
	}
	if err := machine.Handle(context.Background(), streamFrame(xid, typeMessage(23, "pg_catalog", "int4")), pglogrepl.LSN(0x11)); err != nil {
		t.Fatalf("Handle(Type) error = %v", err)
	}
	if err := machine.Handle(context.Background(), streamDMLFrame(xid, insertMessage(testRelationID, textField("7"))), pglogrepl.LSN(0x12)); err != nil {
		t.Fatalf("Handle(Insert) error = %v", err)
	}
	if err := machine.Handle(context.Background(), streamStopFrame(), pglogrepl.LSN(0x13)); err != nil {
		t.Fatalf("Handle(StreamStop) error = %v", err)
	}
	if got := probe.order; len(got) != 0 {
		t.Fatalf("delivery before stream abort = %#v, want no emitted event, durable receipt, checkpoint, or acknowledgement", got)
	}
	if err := machine.Handle(context.Background(), streamAbortFrame(xid), pglogrepl.LSN(0x14)); err != nil {
		t.Fatalf("Handle(StreamAbort) error = %v", err)
	}
	if got := probe.order; len(got) != 0 {
		t.Fatalf("delivery after stream abort = %#v, want no emitted event, durable receipt, checkpoint, or acknowledgement", got)
	}
}

func TestPGOutputV2StreamCommitReceiptsBeforeCheckpointAndAcknowledgement(t *testing.T) {
	t.Parallel()

	probe := newCDCv2Probe()
	machine := newTestCDCv2Machine(t, probe)
	const xid = uint32(72)
	for _, frame := range [][]byte{
		streamStartFrame(xid, true),
		streamRelationFrame(xid, relationMessage(testRelationID, "public", "users", testColumn{name: "id", typeID: 23})),
		streamDMLFrame(xid, insertMessage(testRelationID, textField("8"))),
		streamStopFrame(),
	} {
		if err := machine.Handle(context.Background(), frame, pglogrepl.LSN(0x20)); err != nil {
			t.Fatalf("Handle(pre-commit frame) error = %v", err)
		}
	}
	if got := probe.order; len(got) != 0 {
		t.Fatalf("delivery before StreamCommit = %#v, want no events, receipt, checkpoint, or acknowledgement", got)
	}
	transaction, ok := machine.transactions[xid]
	if !ok {
		t.Fatal("streamed transaction was not staged before StreamCommit")
	}
	probe.transactionID = transaction.id
	if err := machine.Handle(context.Background(), streamCommitFrame(xid, pglogrepl.LSN(0x30)), pglogrepl.LSN(0x30)); err != nil {
		t.Fatalf("Handle(StreamCommit) error = %v", err)
	}
	if got, want := probe.order, []string{"emit", "receipt", "checkpoint", "ack"}; !sameStrings(got, want) {
		t.Fatalf("StreamCommit ordering = %#v, want %#v", got, want)
	}
	if got := probe.events; len(got) != 1 || got[0].Operation != "insert" || got[0].Record["id"] != 8 {
		t.Fatalf("committed events = %#v, want one ordered insert", got)
	}
}

func TestPGOutputV2NonStreamBeginCommitsOnlyAtCommitBoundary(t *testing.T) {
	t.Parallel()

	probe := newCDCv2Probe()
	machine := newTestCDCv2Machine(t, probe)
	const xid = uint32(73)
	for _, frame := range [][]byte{
		beginFrame(xid),
		relationMessage(testRelationID, "public", "users", testColumn{name: "id", typeID: 23}),
		insertMessage(testRelationID, textField("9")),
	} {
		if err := machine.Handle(context.Background(), frame, pglogrepl.LSN(0x40)); err != nil {
			t.Fatalf("Handle(pre-commit frame) error = %v", err)
		}
	}
	if got := probe.order; len(got) != 0 {
		t.Fatalf("delivery before Commit = %#v, want no events, receipt, checkpoint, or acknowledgement", got)
	}
	probe.transactionID = machine.transactions[xid].id
	if err := machine.Handle(context.Background(), commitFrame(pglogrepl.LSN(0x41), pglogrepl.LSN(0x42)), pglogrepl.LSN(0x42)); err != nil {
		t.Fatalf("Handle(Commit) error = %v", err)
	}
	if got, want := probe.order, []string{"emit", "receipt", "checkpoint", "ack"}; !sameStrings(got, want) {
		t.Fatalf("Commit ordering = %#v, want %#v", got, want)
	}
}

func TestPGOutputV2RejectsStreamFrameForAnotherTransaction(t *testing.T) {
	t.Parallel()

	probe := newCDCv2Probe()
	machine := newTestCDCv2Machine(t, probe)
	const activeXID = uint32(75)
	if err := machine.Handle(context.Background(), streamStartFrame(activeXID, true), pglogrepl.LSN(0x50)); err != nil {
		t.Fatalf("Handle(StreamStart) error = %v", err)
	}
	if err := machine.Handle(context.Background(), streamRelationFrame(activeXID, relationMessage(testRelationID, "public", "users", testColumn{name: "id", typeID: 23})), pglogrepl.LSN(0x51)); err != nil {
		t.Fatalf("Handle(Relation) error = %v", err)
	}
	err := machine.Handle(context.Background(), streamDMLFrame(activeXID+1, insertMessage(testRelationID, textField("10"))), pglogrepl.LSN(0x52))
	if err == nil || !strings.Contains(err.Error(), "does not match the active stream") {
		t.Fatalf("Handle(mismatched streamed DML) = %v, want stream transaction mismatch", err)
	}
	if got := probe.order; len(got) != 0 {
		t.Fatalf("mismatched streamed DML caused downstream activity: %#v", got)
	}
}

func TestCDCStageRecoveryRequiresTypedRebootstrapForReceiptlessTransaction(t *testing.T) {
	t.Parallel()

	source := postgresCDCSource{identity: synccontract.SourceIdentity{
		Engine: "postgres", AccountOrCluster: "system:database", ObjectScope: "public.users",
	}}
	projectDir := t.TempDir()
	stage, err := newPostgresCDCTransactionStage(projectDir, source)
	if err != nil {
		t.Fatalf("newPostgresCDCTransactionStage() error = %v", err)
	}
	transactionID := cdcTransactionID(source, 74, pglogrepl.LSN(0x50))
	if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	if err := stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("receiptless")); err != nil {
		t.Fatalf("AppendChunk() error = %v", err)
	}
	_, err = stage.CommitTransaction(context.Background(), transactionID, cdcStageReceiverFunc(func(context.Context, database.CommittedTransaction) (database.DownstreamTransactionReceipt, error) {
		return database.DownstreamTransactionReceipt{}, errors.New("downstream unavailable")
	}))
	if err == nil {
		t.Fatal("CommitTransaction() error = nil, want receipt-less staged transaction")
	}
	_, err = newPostgresCDCTransactionStage(projectDir, source)
	if !errors.Is(err, synccontract.ErrRebootstrapRequired) {
		t.Fatalf("newPostgresCDCTransactionStage(recovery) = %v, want typed rebootstrap", err)
	}
}

func TestCDCTransactionIdentityIncludesWALPosition(t *testing.T) {
	t.Parallel()

	source := postgresCDCSource{identity: synccontract.SourceIdentity{
		Engine: "postgres", AccountOrCluster: "system:database", ObjectScope: "public.users",
	}}
	if first, second := cdcTransactionID(source, 76, pglogrepl.LSN(0x60)), cdcTransactionID(source, 76, pglogrepl.LSN(0x61)); first == second {
		t.Fatal("CDC transaction identity must distinguish WAL positions for a reused PostgreSQL XID")
	}
}

type cdcV2Probe struct {
	order         []string
	events        []connectors.CDCEvent
	stage         *database.CommittedTransactionStage
	transactionID string
}

func newCDCv2Probe() *cdcV2Probe { return &cdcV2Probe{} }

func newTestCDCv2Machine(t *testing.T, probe *cdcV2Probe) *pgoutputV2TransactionMachine {
	t.Helper()
	source := postgresCDCSource{
		identity:   synccontract.SourceIdentity{Engine: "postgres", AccountOrCluster: "system:database", ObjectScope: "public.users"},
		generation: synccontract.OpaqueToken("timeline"),
	}
	stage, err := database.OpenCommittedTransactionStage(database.TransactionStageOptions{
		Root: t.TempDir(),
		Limits: database.TransactionStageLimits{
			MaxTransactionBytes:   1 << 20,
			MaxTransactionRecords: 16,
			MaxTransactionAge:     time.Minute,
			MaxStagedBytes:        1 << 20,
			MaxStagedTransactions: 4,
		},
	})
	if err != nil {
		t.Fatalf("OpenCommittedTransactionStage() error = %v", err)
	}
	probe.stage = stage
	return newPGOutputV2TransactionMachine(stage, source, pglogrepl.LSN(0x01), pglogrepl.LSN(0x01), connectors.CDCReadRequest{
		DurableCheckpointCommitter: cdcCommitterFunc(func(_ context.Context, candidate synccontract.CheckpointEnvelope) error {
			if probe.transactionID == "" {
				return errors.New("checkpoint observed without a transaction receipt target")
			}
			if _, err := probe.stage.Receipt(probe.transactionID); err != nil {
				return errors.New("checkpoint observed before durable transaction receipt")
			}
			probe.order = append(probe.order, "checkpoint")
			return nil
		}),
	}, func(event connectors.CDCEvent) error {
		probe.order = append(probe.order, "emit")
		probe.events = append(probe.events, event)
		return nil
	}, func(_ context.Context, position pglogrepl.LSN) error {
		if position == 0 {
			return errors.New("acknowledgement position is missing")
		}
		probe.order = append(probe.order, "ack")
		return nil
	}, func() { probe.order = append(probe.order, "receipt") })
}

type cdcCommitterFunc func(context.Context, synccontract.CheckpointEnvelope) error

func (f cdcCommitterFunc) CommitDurableChangefeedCheckpoint(ctx context.Context, candidate synccontract.CheckpointEnvelope) error {
	return f(ctx, candidate)
}

type cdcStageReceiverFunc func(context.Context, database.CommittedTransaction) (database.DownstreamTransactionReceipt, error)

func (f cdcStageReceiverFunc) ReceiveCommittedTransaction(ctx context.Context, transaction database.CommittedTransaction) (database.DownstreamTransactionReceipt, error) {
	return f(ctx, transaction)
}

func streamStartFrame(xid uint32, first bool) []byte {
	frame := append([]byte{'S'}, uint32Frame(xid)...)
	if first {
		return append(frame, 1)
	}
	return append(frame, 0)
}

func streamStopFrame() []byte { return []byte{'E'} }

func streamAbortFrame(xid uint32) []byte {
	frame := append([]byte{'A'}, uint32Frame(xid)...)
	return append(frame, uint32Frame(xid)...)
}

func streamCommitFrame(xid uint32, lsn pglogrepl.LSN) []byte {
	frame := append([]byte{'c'}, uint32Frame(xid)...)
	frame = append(frame, 0)
	frame = append(frame, uint64Frame(uint64(lsn))...)
	frame = append(frame, uint64Frame(uint64(lsn))...)
	return append(frame, uint64Frame(0)...)
}

func beginFrame(xid uint32) []byte {
	frame := append([]byte{'B'}, uint64Frame(0x40)...)
	frame = append(frame, uint64Frame(0)...)
	return append(frame, uint32Frame(xid)...)
}

func commitFrame(commitLSN, endLSN pglogrepl.LSN) []byte {
	frame := append([]byte{'C', 0}, uint64Frame(uint64(commitLSN))...)
	frame = append(frame, uint64Frame(uint64(endLSN))...)
	return append(frame, uint64Frame(0)...)
}

func streamRelationFrame(xid uint32, frame []byte) []byte { return streamFrame(xid, frame) }
func streamDMLFrame(xid uint32, frame []byte) []byte      { return streamFrame(xid, frame) }

func streamFrame(xid uint32, frame []byte) []byte {
	out := make([]byte, 0, len(frame)+4)
	out = append(out, frame[0])
	out = append(out, uint32Frame(xid)...)
	return append(out, frame[1:]...)
}

func uint32Frame(value uint32) []byte {
	frame := make([]byte, 4)
	binary.BigEndian.PutUint32(frame, value)
	return frame
}

func uint64Frame(value uint64) []byte {
	frame := make([]byte, 8)
	binary.BigEndian.PutUint64(frame, value)
	return frame
}
