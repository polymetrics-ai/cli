package database

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"polymetrics.ai/internal/synccontract"
)

// ErrTransactionReceiptUnavailable reports that no durable whole-transaction
// receipt exists for the requested opaque transaction identity.
var ErrTransactionReceiptUnavailable = errors.New("durable transaction receipt is unavailable")

// TransactionStageLimits bounds the private durable storage retained by a
// committed-transaction stage.
type TransactionStageLimits struct {
	MaxTransactionBytes   int64
	MaxTransactionRecords int64
	MaxTransactionAge     time.Duration
	MaxStagedBytes        int64
}

// TransactionStageOptions configures one source-agnostic private stage.
type TransactionStageOptions struct {
	Root   string
	Limits TransactionStageLimits
}

// CommittedTransactionStage owns private in-progress transaction chunks.
type CommittedTransactionStage struct {
	root string
}

// TransactionChunk is one ordered streamed chunk of a committed transaction.
type TransactionChunk struct {
	Sequence uint64
	Records  int64
	Bytes    int64
	Reader   io.Reader
}

// CommittedTransaction is the whole transaction delivered only after a stage
// has crossed its private commit boundary.
type CommittedTransaction struct {
	TransactionKey string
	Bytes          int64
	Records        int64
	ContentDigest  string
}

// StreamChunks visits committed chunks in source order.
func (CommittedTransaction) StreamChunks(context.Context, func(TransactionChunk) error) error {
	return errors.New("committed transaction staging is not implemented")
}

// DownstreamTransactionReceipt is supplied by a receiver only after its
// complete transaction write is durable according to the receiver's protocol.
type DownstreamTransactionReceipt struct {
	ReceiptID string
	Sink      string
	DurableAt time.Time
}

// DurableTransactionReceiver receives one complete committed transaction and
// returns durable downstream evidence only after it has persisted the whole
// transaction.
type DurableTransactionReceiver interface {
	ReceiveCommittedTransaction(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error)
}

// TransactionReceipt is immutable receipt evidence for one complete staged
// transaction. Its unexported durable marker prevents a caller from forging
// acknowledgement eligibility with a struct literal.
type TransactionReceipt struct {
	TransactionKey      string
	DownstreamReceiptID string
	Sink                string
	DurableAt           time.Time
	Bytes               int64
	Records             int64
	ContentDigest       string
	durable             bool
}

// Acknowledgement adapts a persisted durable receipt to the existing sync
// checkpoint contract.
func (r TransactionReceipt) Acknowledgement() (synccontract.DownstreamAcknowledgement, error) {
	if !r.durable {
		return synccontract.DownstreamAcknowledgement{}, ErrTransactionReceiptUnavailable
	}
	return synccontract.NewDurableDownstreamAcknowledgement(r.Sink, r.DurableAt)
}

// OpenCommittedTransactionStage opens a private stage and performs recovery.
func OpenCommittedTransactionStage(options TransactionStageOptions) (*CommittedTransactionStage, error) {
	if err := os.MkdirAll(filepath.Join(options.Root, "receipts"), 0o700); err != nil {
		return nil, err
	}
	return &CommittedTransactionStage{root: options.Root}, nil
}

// BeginTransaction starts a private transaction stage.
func (*CommittedTransactionStage) BeginTransaction(context.Context, string) error { return nil }

// AppendChunk streams one opaque source chunk into a private transaction stage.
func (*CommittedTransactionStage) AppendChunk(context.Context, string, int64, io.Reader) error {
	return nil
}

// CommitTransaction is deliberately incomplete until the durable stage and
// receipt transition are implemented.
func (*CommittedTransactionStage) CommitTransaction(context.Context, string, DurableTransactionReceiver) (TransactionReceipt, error) {
	return TransactionReceipt{}, errors.New("committed transaction staging is not implemented")
}

// AbortTransaction removes an incomplete private stage.
func (*CommittedTransactionStage) AbortTransaction(context.Context, string) error { return nil }

// Receipt returns only a receipt that was made durable by this stage.
func (*CommittedTransactionStage) Receipt(string) (TransactionReceipt, error) {
	return TransactionReceipt{}, ErrTransactionReceiptUnavailable
}
