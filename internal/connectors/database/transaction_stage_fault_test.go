package database

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
)

func TestCommittedTransactionStageBeginAppendAndSealFaultBoundariesCleanIncompleteState(t *testing.T) {
	tests := []struct {
		name      string
		phase     transactionStageFaultPhase
		operation string
		skip      int
		match     func(*CommittedTransactionStage, string, string) bool
		err       error
	}{
		{
			name:      "begin directory creation disk full",
			phase:     transactionStageFaultBegin,
			operation: "mkdir_all",
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.chunksDirectory(key)
			},
			err: syscall.ENOSPC,
		},
		{
			name:      "begin state write",
			phase:     transactionStageFaultBegin,
			operation: "write",
			match:     isTransactionStageStateTemporary,
			err:       errors.New("injected begin state write failure"),
		},
		{
			name:      "begin state file sync",
			phase:     transactionStageFaultBegin,
			operation: "file_sync",
			match:     isTransactionStageStateTemporary,
			err:       errors.New("injected begin state sync failure"),
		},
		{
			name:      "begin state rename",
			phase:     transactionStageFaultBegin,
			operation: "rename",
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.manifestPath(key)
			},
			err: errors.New("injected begin state rename failure"),
		},
		{
			name:      "begin state parent sync",
			phase:     transactionStageFaultBegin,
			operation: "directory_sync",
			skip:      1, // createStageDirectory syncs this directory before state.json exists.
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.transactionDirectory(key)
			},
			err: errors.New("injected begin state directory sync failure"),
		},
		{
			name:      "append chunk write disk full",
			phase:     transactionStageFaultAppend,
			operation: "write",
			match:     isTransactionStageChunkTemporary,
			err:       syscall.ENOSPC,
		},
		{
			name:      "append chunk file sync",
			phase:     transactionStageFaultAppend,
			operation: "file_sync",
			match:     isTransactionStageChunkTemporary,
			err:       errors.New("injected chunk sync failure"),
		},
		{
			name:      "append chunk rename",
			phase:     transactionStageFaultAppend,
			operation: "rename",
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.chunkPath(key, 0)
			},
			err: errors.New("injected chunk rename failure"),
		},
		{
			name:      "append chunk parent sync",
			phase:     transactionStageFaultAppend,
			operation: "directory_sync",
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.chunksDirectory(key)
			},
			err: errors.New("injected chunk directory sync failure"),
		},
		{
			name:      "append manifest write",
			phase:     transactionStageFaultAppend,
			operation: "write",
			match:     isTransactionStageStateTemporary,
			err:       errors.New("injected append manifest write failure"),
		},
		{
			name:      "append manifest file sync",
			phase:     transactionStageFaultAppend,
			operation: "file_sync",
			match:     isTransactionStageStateTemporary,
			err:       errors.New("injected append manifest sync failure"),
		},
		{
			name:      "append manifest rename",
			phase:     transactionStageFaultAppend,
			operation: "rename",
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.manifestPath(key)
			},
			err: errors.New("injected append manifest rename failure"),
		},
		{
			name:      "append manifest parent sync",
			phase:     transactionStageFaultAppend,
			operation: "directory_sync",
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.transactionDirectory(key)
			},
			err: errors.New("injected append manifest directory sync failure"),
		},
		{
			name:      "seal state write",
			phase:     transactionStageFaultSeal,
			operation: "write",
			match:     isTransactionStageStateTemporary,
			err:       errors.New("injected seal state write failure"),
		},
		{
			name:      "seal state file sync",
			phase:     transactionStageFaultSeal,
			operation: "file_sync",
			match:     isTransactionStageStateTemporary,
			err:       errors.New("injected seal state sync failure"),
		},
		{
			name:      "seal state rename",
			phase:     transactionStageFaultSeal,
			operation: "rename",
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.manifestPath(key)
			},
			err: errors.New("injected seal state rename failure"),
		},
		{
			name:      "seal state parent sync",
			phase:     transactionStageFaultSeal,
			operation: "directory_sync",
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.transactionDirectory(key)
			},
			err: errors.New("injected seal state directory sync failure"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			stage := newTestTransactionStage(t, root, testTransactionStageLimits())
			const transactionID = "fault-boundary"
			key, err := transactionStageKey(transactionID)
			if err != nil {
				t.Fatalf("transactionStageKey() error = %v", err)
			}
			fault := &transactionStageFault{operation: tt.operation, skip: tt.skip, match: func(path string) bool {
				return tt.match(stage, key, path)
			}, err: tt.err}

			var operationErr error
			switch tt.phase {
			case transactionStageFaultBegin:
				installTransactionStageFault(stage, fault)
				operationErr = stage.BeginTransaction(context.Background(), transactionID)
			case transactionStageFaultAppend:
				if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
					t.Fatalf("BeginTransaction() error = %v", err)
				}
				installTransactionStageFault(stage, fault)
				operationErr = stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("chunk"))
			case transactionStageFaultSeal:
				if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
					t.Fatalf("BeginTransaction() error = %v", err)
				}
				if err := stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("chunk")); err != nil {
					t.Fatalf("AppendChunk() error = %v", err)
				}
				installTransactionStageFault(stage, fault)
				receiver := &recordingTransactionReceiver{}
				_, operationErr = stage.CommitTransaction(context.Background(), transactionID, receiver)
				if got := len(receiver.transactions); got != 0 {
					t.Fatalf("receiver transactions after seal durability fault = %d, want 0", got)
				}
			default:
				t.Fatalf("unknown fault phase %d", tt.phase)
			}

			if operationErr == nil {
				t.Fatal("faulted transaction stage operation error = nil, want refusal")
			}
			if !fault.fired() {
				t.Fatalf("fault %+v was not exercised", tt)
			}
			if _, err := stage.Receipt(transactionID); !errors.Is(err, ErrTransactionReceiptUnavailable) {
				t.Fatalf("Receipt() after durability fault error = %v, want unavailable", err)
			}
			assertPrivateStageCounts(t, root, 0, 0)
			assertNoTemporaryStageArtifacts(t, root)
			recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
			if got := recovered.PendingTransactions(); len(got) != 0 {
				t.Fatalf("PendingTransactions() after crash-boundary recovery = %#v, want none", got)
			}
			assertPrivateStageCounts(t, root, 0, 0)
		})
	}
}

func TestCommittedTransactionStageReceiverAndReceiptFaultsRemainReceiptless(t *testing.T) {
	t.Run("receiver failure leaves sealed retryable transaction", func(t *testing.T) {
		root := t.TempDir()
		stage := newTestTransactionStage(t, root, testTransactionStageLimits())
		const transactionID = "receiver-failure"
		stageCompleteChunk(t, stage, transactionID)
		_, err := stage.CommitTransaction(context.Background(), transactionID, transactionReceiverFunc(func(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error) {
			return DownstreamTransactionReceipt{}, errors.New("injected receiver failure")
		}))
		if err == nil {
			t.Fatal("CommitTransaction() error = nil, want receiver failure")
		}
		assertReceiptlessSealedStage(t, root, stage, transactionID)
	})

	tests := []struct {
		name      string
		operation string
		match     func(*CommittedTransactionStage, string, string) bool
		err       error
	}{
		{
			name:      "receipt write disk full",
			operation: "write",
			match:     isTransactionStageReceiptTemporary,
			err:       syscall.ENOSPC,
		},
		{
			name:      "receipt file sync",
			operation: "file_sync",
			match:     isTransactionStageReceiptTemporary,
			err:       errors.New("injected receipt file sync failure"),
		},
		{
			name:      "receipt rename",
			operation: "rename",
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.receiptPath(key)
			},
			err: errors.New("injected receipt rename failure"),
		},
		{
			name:      "receipt parent sync",
			operation: "directory_sync",
			match: func(stage *CommittedTransactionStage, key, path string) bool {
				return path == stage.receiptsDirectory()
			},
			err: errors.New("injected receipt directory sync failure"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			stage := newTestTransactionStage(t, root, testTransactionStageLimits())
			const transactionID = "receipt-failure"
			stageCompleteChunk(t, stage, transactionID)
			key, err := transactionStageKey(transactionID)
			if err != nil {
				t.Fatalf("transactionStageKey() error = %v", err)
			}
			fault := &transactionStageFault{operation: tt.operation, match: func(path string) bool {
				return tt.match(stage, key, path)
			}, err: tt.err}
			installTransactionStageFault(stage, fault)
			receiver := &recordingTransactionReceiver{}
			if _, err := stage.CommitTransaction(context.Background(), transactionID, receiver); err == nil {
				t.Fatal("CommitTransaction() error = nil, want receipt durability failure")
			}
			if !fault.fired() {
				t.Fatalf("receipt fault %+v was not exercised", tt)
			}
			if got := len(receiver.transactions); got != 1 {
				t.Fatalf("receiver transactions = %d, want one whole transaction before receipt persistence", got)
			}
			assertReceiptlessSealedStage(t, root, stage, transactionID)

			recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
			if got := recovered.PendingTransactions(); len(got) != 1 {
				t.Fatalf("PendingTransactions() after receipt fault recovery = %#v, want one sealed retry", got)
			}
			if _, err := recovered.Receipt(transactionID); !errors.Is(err, ErrTransactionReceiptUnavailable) {
				t.Fatalf("recovered Receipt() error = %v, want unavailable", err)
			}
		})
	}
}

func TestCommittedTransactionStageFailedAppendCleanupStaysTerminalUntilAbortRetries(t *testing.T) {
	root := t.TempDir()
	limits := testTransactionStageLimits()
	limits.MaxTransactionRecords = 1
	stage := newTestTransactionStage(t, root, limits)
	const transactionID = "failed-append-cleanup"
	stageCompleteChunk(t, stage, transactionID)
	key, err := transactionStageKey(transactionID)
	if err != nil {
		t.Fatalf("transactionStageKey() error = %v", err)
	}
	fault := &transactionStageFault{
		operation: "remove_all",
		match: func(path string) bool {
			return path == stage.transactionDirectory(key)
		},
		err: errors.New("injected failed-append cleanup failure"),
	}
	installTransactionStageFault(stage, fault)
	assertTransactionStageLimit(t, stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("partial")), TransactionStageLimitTransactionRecords)
	if !fault.fired() {
		t.Fatal("failed-append cleanup fault was not exercised")
	}
	if _, err := stage.CommitTransaction(context.Background(), transactionID, &recordingTransactionReceiver{}); !errors.Is(err, ErrTransactionStageInProgress) {
		t.Fatalf("CommitTransaction() error = %v, want ErrTransactionStageInProgress", err)
	}
	if got := stage.PendingTransactions(); len(got) != 0 {
		t.Fatalf("PendingTransactions() = %#v, want no terminal failed append", got)
	}
	assertPrivateStageCounts(t, root, 1, 0)
	if err := stage.AbortTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("AbortTransaction() retry error = %v", err)
	}
	assertPrivateStageCounts(t, root, 0, 0)
}

func TestCommittedTransactionStageDiscardedSealedStageNeverRecoversOrDelivers(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const transactionID = "discarded-sealed-recovery"
	stageCompleteChunk(t, stage, transactionID)
	if _, err := stage.CommitTransaction(context.Background(), transactionID, transactionReceiverFunc(func(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error) {
		return DownstreamTransactionReceipt{}, errors.New("injected receiver failure")
	})); err == nil {
		t.Fatal("CommitTransaction() error = nil, want sealed receipt-less transaction")
	}
	key, err := transactionStageKey(transactionID)
	if err != nil {
		t.Fatalf("transactionStageKey() error = %v", err)
	}
	fault := &transactionStageFault{
		operation: "remove_all",
		match: func(path string) bool {
			return path == stage.transactionDirectory(key)
		},
		err: errors.New("injected sealed-discard cleanup failure"),
	}
	installTransactionStageFault(stage, fault)
	abortCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := stage.AbortTransaction(abortCtx, transactionID); err == nil {
		t.Fatal("AbortTransaction() error = nil, want cleanup failure")
	}
	if !fault.fired() {
		t.Fatal("sealed-discard cleanup fault was not exercised")
	}

	recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
	if got := recovered.PendingTransactions(); len(got) != 0 {
		t.Fatalf("PendingTransactions() after discarded recovery = %#v, want none", got)
	}
	receiver := &recordingTransactionReceiver{}
	if _, err := recovered.CommitTransaction(context.Background(), transactionID, receiver); !errors.Is(err, ErrTransactionStageNotFound) {
		t.Fatalf("CommitTransaction() after discarded recovery error = %v, want ErrTransactionStageNotFound", err)
	}
	if got := len(receiver.transactions); got != 0 {
		t.Fatalf("receiver transactions after discarded recovery = %d, want 0", got)
	}
	assertPrivateStageCounts(t, root, 0, 0)
	assertNoTemporaryStageArtifacts(t, root)
}

func TestCommittedTransactionStagePostReceiptCleanupFailurePreservesOnlyDurableReceipt(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const transactionID = "cleanup-failure"
	stageCompleteChunk(t, stage, transactionID)
	key, err := transactionStageKey(transactionID)
	if err != nil {
		t.Fatalf("transactionStageKey() error = %v", err)
	}
	fault := &transactionStageFault{
		operation: "remove_all",
		match: func(path string) bool {
			return path == stage.transactionDirectory(key)
		},
		err: errors.New("injected post-receipt cleanup failure"),
	}
	installTransactionStageFault(stage, fault)
	receipt, err := stage.CommitTransaction(context.Background(), transactionID, &recordingTransactionReceiver{})
	if err != nil {
		t.Fatalf("CommitTransaction() error = %v, want durable receipt despite cleanup retry requirement", err)
	}
	if !fault.fired() {
		t.Fatal("post-receipt cleanup fault was not exercised")
	}
	if _, err := receipt.Acknowledgement(); err != nil {
		t.Fatalf("receipt acknowledgement after cleanup failure error = %v", err)
	}
	assertPrivateStageCounts(t, root, 1, 1)

	recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
	if got := recovered.PendingTransactions(); len(got) != 0 {
		t.Fatalf("PendingTransactions() after receipt-backed cleanup recovery = %#v, want none", got)
	}
	if _, err := recovered.Receipt(transactionID); err != nil {
		t.Fatalf("Receipt() after cleanup recovery error = %v", err)
	}
	assertPrivateStageCounts(t, root, 0, 1)
	assertNoTemporaryStageArtifacts(t, root)
}

type transactionStageFaultPhase uint8

const (
	transactionStageFaultBegin transactionStageFaultPhase = iota
	transactionStageFaultAppend
	transactionStageFaultSeal
)

type transactionStageFault struct {
	mu        sync.Mutex
	operation string
	skip      int
	match     func(string) bool
	err       error
	wasFired  bool
}

func (f *transactionStageFault) apply(operation, path string) error {
	if f == nil || operation != f.operation || f.match == nil || !f.match(path) {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.skip > 0 {
		f.skip--
		return nil
	}
	if f.wasFired {
		return nil
	}
	f.wasFired = true
	return f.err
}

func (f *transactionStageFault) fired() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.wasFired
}

func installTransactionStageFault(stage *CommittedTransactionStage, fault *transactionStageFault) {
	base := stage.storage
	stage.storage.mkdirAll = func(path string, mode os.FileMode) error {
		if err := fault.apply("mkdir_all", path); err != nil {
			return err
		}
		return base.mkdirAll(path, mode)
	}
	stage.storage.createTemp = func(directory, pattern string) (transactionStageFile, error) {
		if err := fault.apply("create_temp", directory); err != nil {
			return nil, err
		}
		file, err := base.createTemp(directory, pattern)
		if err != nil {
			return nil, err
		}
		return &faultedTransactionStageFile{transactionStageFile: file, fault: fault}, nil
	}
	stage.storage.rename = func(from, to string) error {
		if err := fault.apply("rename", to); err != nil {
			return err
		}
		return base.rename(from, to)
	}
	stage.storage.removeAll = func(path string) error {
		if err := fault.apply("remove_all", path); err != nil {
			return err
		}
		return base.removeAll(path)
	}
	stage.storage.syncDirectory = func(path string) error {
		if err := fault.apply("directory_sync", path); err != nil {
			return err
		}
		return base.syncDirectory(path)
	}
}

type faultedTransactionStageFile struct {
	transactionStageFile
	fault *transactionStageFault
}

func (f *faultedTransactionStageFile) Write(payload []byte) (int, error) {
	if err := f.fault.apply("write", f.Name()); err != nil {
		return 0, err
	}
	return f.transactionStageFile.Write(payload)
}

func (f *faultedTransactionStageFile) Sync() error {
	if err := f.fault.apply("file_sync", f.Name()); err != nil {
		return err
	}
	return f.transactionStageFile.Sync()
}

func isTransactionStageChunkTemporary(_ *CommittedTransactionStage, _ string, path string) bool {
	return strings.Contains(filepath.ToSlash(path), "/chunks/.chunk.tmp-")
}

func isTransactionStageStateTemporary(_ *CommittedTransactionStage, _ string, path string) bool {
	return strings.Contains(filepath.ToSlash(path), "/transactions/") && !strings.Contains(filepath.ToSlash(path), "/chunks/") && strings.Contains(filepath.Base(path), ".stage.tmp-")
}

func isTransactionStageReceiptTemporary(_ *CommittedTransactionStage, _ string, path string) bool {
	return strings.Contains(filepath.ToSlash(path), "/receipts/") && strings.Contains(filepath.Base(path), ".stage.tmp-")
}

func stageCompleteChunk(t *testing.T, stage *CommittedTransactionStage, transactionID string) {
	t.Helper()
	if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	if err := stage.AppendChunk(context.Background(), transactionID, 1, strings.NewReader("whole")); err != nil {
		t.Fatalf("AppendChunk() error = %v", err)
	}
}

func assertReceiptlessSealedStage(t *testing.T, root string, stage *CommittedTransactionStage, transactionID string) {
	t.Helper()
	if _, err := stage.Receipt(transactionID); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Fatalf("Receipt() after failure error = %v, want unavailable", err)
	}
	if got := stage.PendingTransactions(); len(got) != 1 {
		t.Fatalf("PendingTransactions() = %#v, want one sealed receipt-less transaction", got)
	}
	assertPrivateStageCounts(t, root, 1, 0)
	assertNoTemporaryStageArtifacts(t, root)
}
