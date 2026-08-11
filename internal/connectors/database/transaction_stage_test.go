package database

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	stage, err := OpenCommittedTransactionStage(TransactionStageOptions{
		Root: root,
		Limits: TransactionStageLimits{
			MaxTransactionBytes:   1 << 20,
			MaxTransactionRecords: 10,
			MaxTransactionAge:     time.Minute,
			MaxStagedBytes:        1 << 20,
			MaxStagedTransactions: 16,
		},
	})
	if err != nil {
		t.Fatalf("OpenCommittedTransactionStage() error = %v", err)
	}

	const transactionID = "source-transaction-1"
	if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	if err := stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("first")); err != nil {
		t.Fatalf("AppendChunk(first) error = %v", err)
	}
	if err := stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("second")); err != nil {
		t.Fatalf("AppendChunk(second) error = %v", err)
	}

	receiver := &recordingTransactionReceiver{}
	if len(receiver.transactions) != 0 {
		t.Fatalf("receiver transactions before commit = %d, want 0", len(receiver.transactions))
	}
	if _, err := stage.Receipt(transactionID); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Fatalf("Receipt() before commit error = %v, want ErrTransactionReceiptUnavailable", err)
	}
	if entries, err := os.ReadDir(filepath.Join(root, "receipts")); err != nil {
		t.Fatalf("ReadDir(receipts) before commit error = %v", err)
	} else if len(entries) != 0 {
		t.Fatalf("receipt artifacts before commit = %d, want 0", len(entries))
	}

	receipt, err := stage.CommitTransaction(context.Background(), transactionID, receiver)
	if err != nil {
		t.Fatalf("CommitTransaction() error = %v", err)
	}
	if got := len(receiver.transactions); got != 1 {
		t.Fatalf("receiver transactions after commit = %d, want 1", got)
	}
	if got, want := receiver.transactions[0].chunks, []string{"first", "second"}; !sameStrings(got, want) {
		t.Fatalf("receiver chunk order = %q, want %q", got, want)
	}
	if receipt.TransactionKey() == "" || receipt.ContentDigest() == "" {
		t.Fatalf("receipt = %#v, want immutable whole-transaction identity and digest", receipt)
	}
	if receipt.DownstreamReceiptID() != "receiver-receipt-1" || receipt.Bytes() != int64(len("first")+len("second")) || receipt.Records() != 2 {
		t.Fatalf("receipt metadata = %#v, want durable whole-transaction metadata", receipt)
	}
	acknowledgement, err := receipt.Acknowledgement()
	if err != nil {
		t.Fatalf("receipt acknowledgement after commit error = %v", err)
	}
	if acknowledgement.Sink != receipt.Sink() || !acknowledgement.AcknowledgedAt.Equal(receipt.DurableAt()) {
		t.Fatalf("receipt acknowledgement = %#v, want durable test-sink acknowledgement", acknowledgement)
	}
	if entries, err := os.ReadDir(filepath.Join(root, "receipts")); err != nil {
		t.Fatalf("ReadDir(receipts) after commit error = %v", err)
	} else if len(entries) != 1 {
		t.Fatalf("receipt artifacts after commit = %d, want 1 durable receipt", len(entries))
	}
	if recovered, err := stage.Receipt(transactionID); err != nil {
		t.Fatalf("Receipt() after commit error = %v", err)
	} else if recovered != receipt {
		t.Fatalf("Receipt() after commit = %#v, want %#v", recovered, receipt)
	}
	reopened := newTestTransactionStage(t, root, testTransactionStageLimits())
	if recovered, err := reopened.Receipt(transactionID); err != nil {
		t.Fatalf("Receipt() from reopened durable stage error = %v", err)
	} else if recovered != receipt {
		t.Fatalf("Receipt() from reopened durable stage = %#v, want persisted %#v", recovered, receipt)
	}
}

type recordedTransaction struct {
	chunks []string
}

type recordingTransactionReceiver struct {
	transactions []recordedTransaction
}

func (r *recordingTransactionReceiver) ReceiveCommittedTransaction(ctx context.Context, transaction CommittedTransaction) (DownstreamTransactionReceipt, error) {
	var chunks []string
	if err := transaction.StreamChunks(ctx, func(chunk TransactionChunk) error {
		payload, err := io.ReadAll(chunk.Reader)
		if err != nil {
			return err
		}
		chunks = append(chunks, string(payload))
		return nil
	}); err != nil {
		return DownstreamTransactionReceipt{}, err
	}
	r.transactions = append(r.transactions, recordedTransaction{chunks: chunks})
	return DownstreamTransactionReceipt{
		ReceiptID: "receiver-receipt-1",
		Sink:      "test-sink",
		DurableAt: time.Now().UTC(),
	}, nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func TestCommittedTransactionStageAbortRemovesPrivateChunks(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const transactionID = "abort-me"
	if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	if err := stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("private")); err != nil {
		t.Fatalf("AppendChunk() error = %v", err)
	}
	if err := stage.AbortTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("AbortTransaction() error = %v", err)
	}
	if _, err := stage.Receipt(transactionID); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Fatalf("Receipt() after abort error = %v, want ErrTransactionReceiptUnavailable", err)
	}
	assertPrivateStageCounts(t, root, 0, 0)
	assertNoTemporaryStageArtifacts(t, root)

	if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("BeginTransaction() after abort error = %v, want clean identity reuse", err)
	}
	if err := stage.AbortTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("AbortTransaction() after reuse error = %v", err)
	}
}

func TestCommittedTransactionStageDiscardIntentDoesNotBlockReusedIdentity(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const transactionID = "discard-then-reuse"
	stageCompleteChunk(t, stage, transactionID)
	if err := stage.AbortTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("AbortTransaction() error = %v", err)
	}
	stageCompleteChunk(t, stage, transactionID)
	if _, err := stage.CommitTransaction(context.Background(), transactionID, transactionReceiverFunc(func(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error) {
		return DownstreamTransactionReceipt{}, errors.New("injected receiver failure")
	})); err == nil {
		t.Fatal("CommitTransaction() error = nil, want sealed receipt-less transaction")
	}

	recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
	if got := recovered.PendingTransactions(); len(got) != 1 {
		t.Fatalf("PendingTransactions() after identity reuse = %#v, want one recovery-held transaction", got)
	}
	if err := recovered.AdmitRecoveredTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("AdmitRecoveredTransaction() error = %v", err)
	}
	receiver := &recordingTransactionReceiver{}
	if _, err := recovered.CommitTransaction(context.Background(), transactionID, receiver); err != nil {
		t.Fatalf("CommitTransaction() after identity reuse admission error = %v", err)
	}
	if got := len(receiver.transactions); got != 1 {
		t.Fatalf("receiver transactions after identity reuse admission = %d, want 1", got)
	}
}

func TestCommittedTransactionStageRejectsQuotasAndLeavesNoResidue(t *testing.T) {
	t.Run("transaction bytes", func(t *testing.T) {
		limits := testTransactionStageLimits()
		limits.MaxTransactionBytes = 5
		limits.MaxStagedBytes = 5
		stage := newTestTransactionStage(t, t.TempDir(), limits)
		const transactionID = "byte-quota"
		if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
			t.Fatalf("BeginTransaction() error = %v", err)
		}
		assertTransactionStageLimit(t, stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("too-big")), TransactionStageLimitTransactionBytes)
		assertTransactionStageRejectedWithoutResidue(t, stage, transactionID)
	})

	t.Run("transaction records", func(t *testing.T) {
		limits := testTransactionStageLimits()
		limits.MaxTransactionRecords = 1
		stage := newTestTransactionStage(t, t.TempDir(), limits)
		const transactionID = "record-quota"
		if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
			t.Fatalf("BeginTransaction() error = %v", err)
		}
		if err := stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("first")); err != nil {
			t.Fatalf("AppendChunk(first) error = %v", err)
		}
		assertTransactionStageLimit(t, stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("second")), TransactionStageLimitTransactionRecords)
		assertTransactionStageRejectedWithoutResidue(t, stage, transactionID)
	})

	t.Run("transaction age", func(t *testing.T) {
		root := t.TempDir()
		limits := testTransactionStageLimits()
		limits.MaxTransactionAge = time.Second
		now := time.Date(2026, time.August, 11, 0, 0, 0, 0, time.UTC)
		stage, err := openCommittedTransactionStage(TransactionStageOptions{Root: root, Limits: limits}, defaultTransactionStageStorage(), func() time.Time {
			return now
		})
		if err != nil {
			t.Fatalf("openCommittedTransactionStage() error = %v", err)
		}
		const transactionID = "time-quota"
		if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
			t.Fatalf("BeginTransaction() error = %v", err)
		}
		now = now.Add(2 * time.Second)
		assertTransactionStageLimit(t, stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("late")), TransactionStageLimitTransactionAge)
		assertTransactionStageRejectedWithoutResidue(t, stage, transactionID)
	})
}

func TestCommittedTransactionStageRetainsRecoveredBytesAgainstRootQuota(t *testing.T) {
	root := t.TempDir()
	limits := testTransactionStageLimits()
	limits.MaxTransactionBytes = 4
	limits.MaxStagedBytes = 4
	stage := newTestTransactionStage(t, root, limits)
	if err := stage.BeginTransaction(context.Background(), "sealed"); err != nil {
		t.Fatalf("BeginTransaction(sealed) error = %v", err)
	}
	if err := stage.AppendChunk(context.Background(), "sealed", 1, strings.NewReader("four")); err != nil {
		t.Fatalf("AppendChunk(sealed) error = %v", err)
	}
	if _, err := stage.CommitTransaction(context.Background(), "sealed", transactionReceiverFunc(func(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error) {
		return DownstreamTransactionReceipt{}, errors.New("receiver unavailable")
	})); err == nil {
		t.Fatal("CommitTransaction(sealed) error = nil, want receiver failure")
	}

	recovered := newTestTransactionStage(t, root, limits)
	if got := recovered.PendingTransactions(); len(got) != 1 || got[0].Bytes != 4 {
		t.Fatalf("PendingTransactions() = %#v, want sealed four-byte stage", got)
	}
	if err := recovered.BeginTransaction(context.Background(), "new"); err != nil {
		t.Fatalf("BeginTransaction(new) error = %v", err)
	}
	assertTransactionStageLimit(t, recovered.AppendChunk(context.Background(), "new", 1, strings.NewReader("x")), TransactionStageLimitStagedBytes)
	if got := recovered.PendingTransactions(); len(got) != 1 || got[0].TransactionKey == "" {
		t.Fatalf("PendingTransactions() after root refusal = %#v, want retained sealed transaction", got)
	}
}

func TestCommittedTransactionStageBoundsControlSlotsBeforeDurableBegin(t *testing.T) {
	root := t.TempDir()
	limits := testTransactionStageLimits()
	limits.MaxStagedTransactions = 1
	stage := newTestTransactionStage(t, root, limits)
	if err := stage.BeginTransaction(context.Background(), "control-slot-first"); err != nil {
		t.Fatalf("BeginTransaction(first) error = %v", err)
	}
	assertTransactionStageLimit(t, stage.BeginTransaction(context.Background(), "control-slot-second"), TransactionStageLimitStagedTransactions)
	assertPrivateStageCounts(t, root, 1, 0)
	if err := stage.AbortTransaction(context.Background(), "control-slot-first"); err != nil {
		t.Fatalf("AbortTransaction(first) error = %v", err)
	}
	if err := stage.BeginTransaction(context.Background(), "control-slot-second"); err != nil {
		t.Fatalf("BeginTransaction(second) after retirement error = %v", err)
	}
}

func TestCommittedTransactionStageSaturatesUntrustedRecordQuotaDiagnostics(t *testing.T) {
	root := t.TempDir()
	limits := testTransactionStageLimits()
	limits.MaxTransactionRecords = 1
	stage := newTestTransactionStage(t, root, limits)
	const transactionID = "record-overflow"
	if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	err := stage.AppendChunk(context.Background(), transactionID, transactionStageMaximumInt64, strings.NewReader("one"))
	assertTransactionStageLimit(t, err, TransactionStageLimitTransactionRecords)
	var limitErr *TransactionStageLimitExceeded
	if !errors.As(err, &limitErr) {
		t.Fatalf("error = %v, want *TransactionStageLimitExceeded", err)
	}
	if limitErr.Observed != transactionStageMaximumInt64 {
		t.Fatalf("limit error observed = %d, want saturated %d", limitErr.Observed, transactionStageMaximumInt64)
	}
	assertTransactionStageRejectedWithoutResidue(t, stage, transactionID)
}

func TestCommittedTransactionStageStreamsWithABoundedBuffer(t *testing.T) {
	root := t.TempDir()
	payloadBytes := transactionStageBufferSize*3 + 17
	limits := testTransactionStageLimits()
	limits.MaxTransactionBytes = int64(payloadBytes + 1)
	limits.MaxStagedBytes = int64(payloadBytes + 1)
	stage := newTestTransactionStage(t, root, limits)
	if err := stage.BeginTransaction(context.Background(), "bounded-reader"); err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	reader := &boundedTransactionReader{remaining: payloadBytes}
	if err := stage.AppendChunk(context.Background(), "bounded-reader", 1, reader); err != nil {
		t.Fatalf("AppendChunk() error = %v", err)
	}
	if reader.maxRequested > transactionStageBufferSize {
		t.Fatalf("AppendChunk() read buffer = %d, want at most %d", reader.maxRequested, transactionStageBufferSize)
	}
	if err := stage.AbortTransaction(context.Background(), "bounded-reader"); err != nil {
		t.Fatalf("AbortTransaction() error = %v", err)
	}
	assertPrivateStageCounts(t, root, 0, 0)
}

func TestCommittedTransactionStageCancellationLeavesNoFinalChunkOrReceipt(t *testing.T) {
	t.Run("during append", func(t *testing.T) {
		root := t.TempDir()
		stage := newTestTransactionStage(t, root, testTransactionStageLimits())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := stage.BeginTransaction(ctx, "cancel-append"); err != nil {
			t.Fatalf("BeginTransaction() error = %v", err)
		}
		err := stage.AppendChunk(ctx, "cancel-append", 1, &cancelAfterFirstRead{cancel: cancel})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("AppendChunk() error = %v, want context.Canceled", err)
		}
		assertPrivateStageCounts(t, root, 0, 0)
		assertNoTemporaryStageArtifacts(t, root)
	})

	t.Run("after receiver delivery before receipt", func(t *testing.T) {
		root := t.TempDir()
		stage := newTestTransactionStage(t, root, testTransactionStageLimits())
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := stage.BeginTransaction(ctx, "cancel-commit"); err != nil {
			t.Fatalf("BeginTransaction() error = %v", err)
		}
		if err := stage.AppendChunk(ctx, "cancel-commit", 1, strings.NewReader("complete")); err != nil {
			t.Fatalf("AppendChunk() error = %v", err)
		}
		_, err := stage.CommitTransaction(ctx, "cancel-commit", transactionReceiverFunc(func(ctx context.Context, transaction CommittedTransaction) (DownstreamTransactionReceipt, error) {
			if err := transaction.StreamChunks(ctx, func(chunk TransactionChunk) error {
				_, err := io.ReadAll(chunk.Reader)
				return err
			}); err != nil {
				return DownstreamTransactionReceipt{}, err
			}
			cancel()
			return DownstreamTransactionReceipt{ReceiptID: "cancelled", Sink: "test-sink", DurableAt: time.Now().UTC()}, nil
		}))
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("CommitTransaction() error = %v, want context.Canceled", err)
		}
		assertPrivateStageCounts(t, root, 0, 0)
		assertNoTemporaryStageArtifacts(t, root)
	})
}

func TestCommittedTransactionStageRestartRecoveryRetainsOnlySealedReceiptlessWork(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const transactionID = "recover-sealed"
	if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	if err := stage.AppendChunk(context.Background(), transactionID, 2, strings.NewReader("recoverable")); err != nil {
		t.Fatalf("AppendChunk() error = %v", err)
	}
	if _, err := stage.CommitTransaction(context.Background(), transactionID, transactionReceiverFunc(func(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error) {
		return DownstreamTransactionReceipt{}, errors.New("receiver unavailable")
	})); err == nil {
		t.Fatal("CommitTransaction() error = nil, want receiver failure")
	}

	recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
	if got := recovered.PendingTransactions(); len(got) != 1 || got[0].Records != 2 || got[0].Bytes != int64(len("recoverable")) {
		t.Fatalf("PendingTransactions() = %#v, want one sealed receipt-less transaction", got)
	}
	if _, err := recovered.Receipt(transactionID); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Fatalf("Receipt() before recovered commit error = %v, want unavailable", err)
	}
	receiver := &recordingTransactionReceiver{}
	if _, err := recovered.CommitTransaction(context.Background(), transactionID, receiver); !errors.Is(err, ErrTransactionStageRecoveryRequired) {
		t.Fatalf("CommitTransaction(recovered) before admission error = %v, want ErrTransactionStageRecoveryRequired", err)
	}
	if got := len(receiver.transactions); got != 0 {
		t.Fatalf("receiver transactions before recovery admission = %d, want 0", got)
	}
	if err := recovered.AdmitRecoveredTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("AdmitRecoveredTransaction() error = %v", err)
	}
	receipt, err := recovered.CommitTransaction(context.Background(), transactionID, receiver)
	if err != nil {
		t.Fatalf("CommitTransaction(recovered) error = %v", err)
	}
	if _, err := receipt.Acknowledgement(); err != nil {
		t.Fatalf("recovered receipt acknowledgement error = %v", err)
	}
	if got := len(receiver.transactions); got != 1 || !sameStrings(receiver.transactions[0].chunks, []string{"recoverable"}) {
		t.Fatalf("recovered receiver transactions = %#v, want one complete transaction", receiver.transactions)
	}
	assertPrivateStageCounts(t, root, 0, 1)
	assertNoTemporaryStageArtifacts(t, root)
}

func TestCommittedTransactionStageRecoveryCleansIncompleteAndOrphanState(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	if err := stage.BeginTransaction(context.Background(), "incomplete"); err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	if err := stage.AppendChunk(context.Background(), "incomplete", 1, strings.NewReader("private")); err != nil {
		t.Fatalf("AppendChunk() error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "transactions", ".orphan"), 0o700); err != nil {
		t.Fatalf("MkdirAll(orphan) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "transactions", ".orphan", "partial"), []byte("partial"), 0o600); err != nil {
		t.Fatalf("WriteFile(orphan) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "receipts", ".stage.tmp-leftover"), []byte("partial"), 0o600); err != nil {
		t.Fatalf("WriteFile(receipt temporary) error = %v", err)
	}

	recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
	if got := recovered.PendingTransactions(); len(got) != 0 {
		t.Fatalf("PendingTransactions() = %#v, want active/orphan state discarded", got)
	}
	if _, err := recovered.Receipt("incomplete"); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Fatalf("Receipt(incomplete) error = %v, want unavailable", err)
	}
	assertPrivateStageCounts(t, root, 0, 0)
	assertNoTemporaryStageArtifacts(t, root)
}

func TestCommittedTransactionStageRejectsReceiverThatDoesNotConsumeWholeTransaction(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	if err := stage.BeginTransaction(context.Background(), "unconsumed"); err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	if err := stage.AppendChunk(context.Background(), "unconsumed", 1, strings.NewReader("must-read")); err != nil {
		t.Fatalf("AppendChunk() error = %v", err)
	}
	_, err := stage.CommitTransaction(context.Background(), "unconsumed", transactionReceiverFunc(func(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error) {
		return DownstreamTransactionReceipt{ReceiptID: "false-success", Sink: "test-sink", DurableAt: time.Now().UTC()}, nil
	}))
	if err == nil {
		t.Fatal("CommitTransaction() error = nil, want refusal for unconsumed whole transaction")
	}
	if _, err := stage.Receipt("unconsumed"); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Fatalf("Receipt() after unconsumed receiver error = %v, want unavailable", err)
	}
	if got := stage.PendingTransactions(); len(got) != 1 {
		t.Fatalf("PendingTransactions() = %#v, want sealed retryable transaction", got)
	}
	assertPrivateStageCounts(t, root, 1, 0)
}

func TestTransactionReceiptCannotForgeAcknowledgementEligibility(t *testing.T) {
	forged := TransactionReceipt{}
	if _, err := forged.Acknowledgement(); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Fatalf("forged receipt Acknowledgement() error = %v, want ErrTransactionReceiptUnavailable", err)
	}
}

func TestCommittedTransactionStageOpaqueIdentityNeverBecomesPath(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	identities := []string{"../escape/one", "../../escape/two", "opaque\x00control", "line\nbreak"}
	for _, identity := range identities {
		if err := stage.BeginTransaction(context.Background(), identity); err != nil {
			t.Fatalf("BeginTransaction(%q) error = %v", identity, err)
		}
		if err := stage.AppendChunk(context.Background(), identity, 1, strings.NewReader("safe")); err != nil {
			t.Fatalf("AppendChunk(%q) error = %v", identity, err)
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, "transactions"))
	if err != nil {
		t.Fatalf("ReadDir(transactions) error = %v", err)
	}
	if len(entries) != len(identities) {
		t.Fatalf("transaction directory entries = %d, want %d", len(entries), len(identities))
	}
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || !validTransactionStageKey(entry.Name()) {
			t.Fatalf("transaction directory entry = %q, want hashed private directory", entry.Name())
		}
		if _, exists := seen[entry.Name()]; exists {
			t.Fatalf("transaction directory entry %q aliases another opaque identity", entry.Name())
		}
		seen[entry.Name()] = struct{}{}
		if relative, err := filepath.Rel(root, filepath.Join(root, "transactions", entry.Name())); err != nil || strings.HasPrefix(relative, "..") {
			t.Fatalf("transaction directory relative path = (%q, %v), want root-contained", relative, err)
		}
	}
	for _, identity := range identities {
		if err := stage.AbortTransaction(context.Background(), identity); err != nil {
			t.Fatalf("AbortTransaction(%q) error = %v", identity, err)
		}
	}
	assertPrivateStageCounts(t, root, 0, 0)
}

func TestCommittedTransactionStageConcurrentTransactionsStayIsolated(t *testing.T) {
	root := t.TempDir()
	limits := testTransactionStageLimits()
	limits.MaxStagedBytes = 8 << 20
	stage := newTestTransactionStage(t, root, limits)
	receiver := &synchronizedTransactionReceiver{transactions: make(map[string][]string)}

	const workers = 12
	errorsByWorker := make(chan error, workers)
	var workersDone sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		worker := worker
		workersDone.Add(1)
		go func() {
			defer workersDone.Done()
			transactionID := fmt.Sprintf("concurrent-%02d", worker)
			payload := fmt.Sprintf("payload-%02d", worker)
			if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
				errorsByWorker <- err
				return
			}
			if err := stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader(payload)); err != nil {
				errorsByWorker <- err
				return
			}
			if _, err := stage.CommitTransaction(context.Background(), transactionID, receiver); err != nil {
				errorsByWorker <- err
			}
		}()
	}
	workersDone.Wait()
	close(errorsByWorker)
	for err := range errorsByWorker {
		t.Errorf("concurrent transaction error = %v", err)
	}
	if t.Failed() {
		return
	}
	if got := receiver.snapshot(); len(got) != workers {
		t.Fatalf("receiver transactions = %#v, want %d isolated transactions", got, workers)
	} else {
		for worker := 0; worker < workers; worker++ {
			key, err := transactionStageKey(fmt.Sprintf("concurrent-%02d", worker))
			if err != nil {
				t.Fatalf("transactionStageKey() error = %v", err)
			}
			if chunks := got[key]; !sameStrings(chunks, []string{fmt.Sprintf("payload-%02d", worker)}) {
				t.Fatalf("receiver chunks for worker %d = %q, want isolated payload", worker, chunks)
			}
		}
	}
	assertPrivateStageCounts(t, root, 0, workers)
	assertNoTemporaryStageArtifacts(t, root)
}

type transactionReceiverFunc func(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error)

func (f transactionReceiverFunc) ReceiveCommittedTransaction(ctx context.Context, transaction CommittedTransaction) (DownstreamTransactionReceipt, error) {
	return f(ctx, transaction)
}

type boundedTransactionReader struct {
	remaining    int
	maxRequested int
}

func (r *boundedTransactionReader) Read(payload []byte) (int, error) {
	if len(payload) > r.maxRequested {
		r.maxRequested = len(payload)
	}
	if r.remaining == 0 {
		return 0, io.EOF
	}
	read := len(payload)
	if read > r.remaining {
		read = r.remaining
	}
	for index := 0; index < read; index++ {
		payload[index] = 'x'
	}
	r.remaining -= read
	return read, nil
}

type cancelAfterFirstRead struct {
	cancel func()
	sent   bool
}

func (r *cancelAfterFirstRead) Read(payload []byte) (int, error) {
	if r.sent {
		return 0, io.EOF
	}
	r.sent = true
	copy(payload, "partial")
	r.cancel()
	return len("partial"), nil
}

type synchronizedTransactionReceiver struct {
	mu           sync.Mutex
	transactions map[string][]string
}

func (r *synchronizedTransactionReceiver) ReceiveCommittedTransaction(ctx context.Context, transaction CommittedTransaction) (DownstreamTransactionReceipt, error) {
	chunks := make([]string, 0)
	if err := transaction.StreamChunks(ctx, func(chunk TransactionChunk) error {
		payload, err := io.ReadAll(chunk.Reader)
		if err != nil {
			return err
		}
		chunks = append(chunks, string(payload))
		return nil
	}); err != nil {
		return DownstreamTransactionReceipt{}, err
	}
	r.mu.Lock()
	r.transactions[transaction.TransactionKey] = append([]string(nil), chunks...)
	r.mu.Unlock()
	return DownstreamTransactionReceipt{
		ReceiptID: "concurrent-" + transaction.TransactionKey,
		Sink:      "test-sink",
		DurableAt: time.Now().UTC(),
	}, nil
}

func (r *synchronizedTransactionReceiver) snapshot() map[string][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := make(map[string][]string, len(r.transactions))
	for key, chunks := range r.transactions {
		clone[key] = append([]string(nil), chunks...)
	}
	return clone
}

func testTransactionStageLimits() TransactionStageLimits {
	return TransactionStageLimits{
		MaxTransactionBytes:   1 << 20,
		MaxTransactionRecords: 100,
		MaxTransactionAge:     time.Minute,
		MaxStagedBytes:        2 << 20,
		MaxStagedTransactions: 64,
	}
}

func newTestTransactionStage(t *testing.T, root string, limits TransactionStageLimits) *CommittedTransactionStage {
	t.Helper()
	stage, err := OpenCommittedTransactionStage(TransactionStageOptions{Root: root, Limits: limits})
	if err != nil {
		t.Fatalf("OpenCommittedTransactionStage() error = %v", err)
	}
	return stage
}

func assertTransactionStageLimit(t *testing.T, err error, want TransactionStageLimit) {
	t.Helper()
	if !errors.Is(err, ErrTransactionStageLimitExceeded) {
		t.Fatalf("error = %v, want ErrTransactionStageLimitExceeded", err)
	}
	var limitErr *TransactionStageLimitExceeded
	if !errors.As(err, &limitErr) {
		t.Fatalf("error = %v, want *TransactionStageLimitExceeded", err)
	}
	if limitErr.Limit != want {
		t.Fatalf("limit error = %#v, want limit %q", limitErr, want)
	}
}

func assertTransactionStageCleanupRequired(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, ErrTransactionStageCleanupRequired) {
		t.Fatalf("error = %v, want ErrTransactionStageCleanupRequired", err)
	}
	var cleanupErr *TransactionStageCleanupError
	if !errors.As(err, &cleanupErr) {
		t.Fatalf("error = %v, want *TransactionStageCleanupError", err)
	}
}

func assertTransactionStageRejectedWithoutResidue(t *testing.T, stage *CommittedTransactionStage, transactionID string) {
	t.Helper()
	if _, err := stage.Receipt(transactionID); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Fatalf("Receipt() after quota refusal error = %v, want unavailable", err)
	}
	assertPrivateStageCounts(t, stage.root, 0, 0)
	assertNoTemporaryStageArtifacts(t, stage.root)
}

func assertPrivateStageCounts(t *testing.T, root string, wantTransactions, wantReceipts int) {
	t.Helper()
	for _, check := range []struct {
		directory string
		want      int
	}{
		{directory: filepath.Join(root, "transactions"), want: wantTransactions},
		{directory: filepath.Join(root, "receipts"), want: wantReceipts},
	} {
		entries, err := os.ReadDir(check.directory)
		if err != nil {
			t.Fatalf("ReadDir(%s) error = %v", check.directory, err)
		}
		if len(entries) != check.want {
			t.Fatalf("private stage entries in %s = %d, want %d (%v)", check.directory, len(entries), check.want, entries)
		}
	}
}

func assertDiscardControlFinalCount(t *testing.T, root string, want int) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "discards"))
	if err != nil {
		t.Fatalf("ReadDir(discards) error = %v", err)
	}
	var got int
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".json") {
			got++
		}
	}
	if got != want {
		t.Fatalf("discard control finals = %d, want %d (%v)", got, want, entries)
	}
}

func assertNoTemporaryStageArtifacts(t *testing.T, root string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(entry.Name(), ".tmp") || entry.Name() == "partial" {
			return fmt.Errorf("temporary stage artifact remains at %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
