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
	assertTransactionStageCleanupRequired(t, func() error {
		_, err := stage.CommitTransaction(context.Background(), transactionID, &recordingTransactionReceiver{})
		return err
	}())
	if got := stage.PendingTransactions(); len(got) != 0 {
		t.Fatalf("PendingTransactions() = %#v, want no terminal failed append", got)
	}
	assertPrivateStageCounts(t, root, 1, 0)
	if err := stage.AbortTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("AbortTransaction() retry error = %v", err)
	}
	if err := stage.ReconcileDiscardControls(context.Background()); err != nil {
		t.Fatalf("ReconcileDiscardControls() error = %v", err)
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

func TestCommittedTransactionStageDiscardIntentWriteFailureAndCleanupFailureHaltsRecovery(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const transactionID = "discard-intent-write-and-cleanup-failure"
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
	writeFault := &transactionStageFault{
		operation: "write",
		match: func(path string) bool {
			return isTransactionStageDiscardControlTemporary(stage, key, path)
		},
		err: errors.New("injected discard-control write failure"),
	}
	cleanupFault := &transactionStageFault{
		operation: "remove_all",
		match: func(path string) bool {
			return path == stage.transactionDirectory(key)
		},
		err: errors.New("injected discard cleanup failure"),
	}
	installTransactionStageFaults(stage, writeFault, cleanupFault)
	if err := stage.AbortTransaction(context.Background(), transactionID); err == nil {
		t.Fatal("AbortTransaction() error = nil, want discard-control and cleanup failure")
	}
	if !writeFault.fired() || !cleanupFault.fired() {
		t.Fatalf("discard faults fired = (%t, %t), want both", writeFault.fired(), cleanupFault.fired())
	}

	recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
	if got := recovered.PendingTransactions(); len(got) != 1 {
		t.Errorf("PendingTransactions() after dual failure = %#v, want one recovery-held sealed transaction", got)
	}
	receiver := &recordingTransactionReceiver{}
	receipt, commitErr := recovered.CommitTransaction(context.Background(), transactionID, receiver)
	if !errors.Is(commitErr, ErrTransactionStageRecoveryRequired) {
		t.Errorf("CommitTransaction() after dual failure error = %v, want ErrTransactionStageRecoveryRequired", commitErr)
	}
	if got := len(receiver.transactions); got != 0 {
		t.Errorf("receiver transactions after dual failure = %d, want 0", got)
	}
	if _, err := recovered.Receipt(transactionID); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Errorf("Receipt() after recovery-held commit error = %v, want unavailable", err)
	}
	if _, err := receipt.Acknowledgement(); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Errorf("receipt acknowledgement after recovery-held commit error = %v, want unavailable", err)
	}
}

func TestCommittedTransactionStageDiscardIntentWriteFailureWithDurableCleanupNeverRecovers(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const transactionID = "discard-intent-write-failure-cleanup-success"
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
	writeFault := &transactionStageFault{
		operation: "write",
		match: func(path string) bool {
			return isTransactionStageDiscardControlTemporary(stage, key, path)
		},
		err: errors.New("injected discard-control write failure"),
	}
	installTransactionStageFault(stage, writeFault)
	if err := stage.AbortTransaction(context.Background(), transactionID); err == nil {
		t.Fatal("AbortTransaction() error = nil, want discard-control failure")
	}
	if !writeFault.fired() {
		t.Fatal("discard-control write fault was not exercised")
	}
	assertPrivateStageCounts(t, root, 0, 0)

	recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
	if got := recovered.PendingTransactions(); len(got) != 0 {
		t.Fatalf("PendingTransactions() after durable cleanup = %#v, want none", got)
	}
	receiver := &recordingTransactionReceiver{}
	if _, err := recovered.CommitTransaction(context.Background(), transactionID, receiver); !errors.Is(err, ErrTransactionStageNotFound) {
		t.Fatalf("CommitTransaction() after durable cleanup error = %v, want ErrTransactionStageNotFound", err)
	}
	if got := len(receiver.transactions); got != 0 {
		t.Fatalf("receiver transactions after durable cleanup = %d, want 0", got)
	}
}

func TestCommittedTransactionStageDiscardIntentParentSyncFailureRetainsExternalMarker(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const transactionID = "discard-intent-parent-sync-failure"
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
	intentSyncFault := &transactionStageFault{
		operation: "directory_sync",
		match: func(path string) bool {
			return path == stage.discardsDirectory()
		},
		err: errors.New("injected discard intent parent sync failure"),
	}
	cleanupFault := &transactionStageFault{
		operation: "remove_all",
		match: func(path string) bool {
			return path == stage.transactionDirectory(key)
		},
		err: errors.New("injected discard cleanup failure"),
	}
	installTransactionStageFaults(stage, intentSyncFault, cleanupFault)
	if err := stage.AbortTransaction(context.Background(), transactionID); err == nil {
		t.Fatal("AbortTransaction() error = nil, want intent sync and cleanup failure")
	}
	if !intentSyncFault.fired() || !cleanupFault.fired() {
		t.Fatalf("discard faults fired = (%t, %t), want both", intentSyncFault.fired(), cleanupFault.fired())
	}
	stage.mu.Lock()
	entry := stage.entries[key]
	stage.mu.Unlock()
	if entry == nil {
		t.Fatal("discarded entry = nil, want terminal cleanup retry state")
	}
	marker := stage.discardIntentPath(key, entry.manifest.instanceID())
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("Stat(discard intent) error = %v, want renamed marker retained", err)
	}

	recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
	if got := recovered.PendingTransactions(); len(got) != 0 {
		t.Fatalf("PendingTransactions() after marker recovery = %#v, want none", got)
	}
	receiver := &recordingTransactionReceiver{}
	if _, err := recovered.CommitTransaction(context.Background(), transactionID, receiver); !errors.Is(err, ErrTransactionStageNotFound) {
		t.Fatalf("CommitTransaction() after marker recovery error = %v, want ErrTransactionStageNotFound", err)
	}
	if got := len(receiver.transactions); got != 0 {
		t.Fatalf("receiver transactions after marker recovery = %d, want 0", got)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("Stat(discard intent) after recovery error = %v, want retired marker", err)
	}
}

func TestCommittedTransactionStageDiscardCleanupParentSyncFailureNeverAdmitsRecovery(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const transactionID = "discard-cleanup-parent-sync-failure"
	stageCompleteChunk(t, stage, transactionID)
	if _, err := stage.CommitTransaction(context.Background(), transactionID, transactionReceiverFunc(func(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error) {
		return DownstreamTransactionReceipt{}, errors.New("injected receiver failure")
	})); err == nil {
		t.Fatal("CommitTransaction() error = nil, want sealed receipt-less transaction")
	}
	cleanupSyncFault := &transactionStageFault{
		operation: "directory_sync",
		match: func(path string) bool {
			return path == stage.transactionsDirectory()
		},
		err: errors.New("injected discard cleanup parent sync failure"),
	}
	installTransactionStageFault(stage, cleanupSyncFault)
	if err := stage.AbortTransaction(context.Background(), transactionID); err == nil {
		t.Fatal("AbortTransaction() error = nil, want cleanup parent sync failure")
	}
	if !cleanupSyncFault.fired() {
		t.Fatal("discard cleanup parent sync fault was not exercised")
	}
	if got := stage.PendingTransactions(); len(got) != 0 {
		t.Fatalf("PendingTransactions() after indeterminate cleanup = %#v, want none", got)
	}

	recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
	if got := recovered.PendingTransactions(); len(got) != 0 {
		t.Fatalf("PendingTransactions() after indeterminate cleanup recovery = %#v, want none", got)
	}
	receiver := &recordingTransactionReceiver{}
	if _, err := recovered.CommitTransaction(context.Background(), transactionID, receiver); !errors.Is(err, ErrTransactionStageNotFound) {
		t.Fatalf("CommitTransaction() after indeterminate cleanup recovery error = %v, want ErrTransactionStageNotFound", err)
	}
	if got := len(receiver.transactions); got != 0 {
		t.Fatalf("receiver transactions after indeterminate cleanup recovery = %d, want 0", got)
	}
}

func TestCommittedTransactionStageDiscardControlRetentionIsBounded(t *testing.T) {
	root := t.TempDir()
	limits := testTransactionStageLimits()
	limits.MaxTransactionBytes = 1
	limits.MaxStagedBytes = 1
	limits.MaxStagedTransactions = 1
	stage := newTestTransactionStage(t, root, limits)

	for _, transactionID := range []string{"discard-control-retention-1", "discard-control-retention-2", "discard-control-retention-3"} {
		if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
			t.Fatalf("BeginTransaction(%q) error = %v", transactionID, err)
		}
		if err := stage.AbortTransaction(context.Background(), transactionID); err != nil {
			t.Fatalf("AbortTransaction(%q) error = %v", transactionID, err)
		}
	}

	_ = newTestTransactionStage(t, root, limits)
	assertDiscardControlFinalCount(t, root, 0)
}

func TestCommittedTransactionStageRecoveryReapsOnlyOwnedDiscardTemps(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	temporary, err := os.CreateTemp(stage.discardsDirectory(), ".stage.tmp-*")
	if err != nil {
		t.Fatalf("CreateTemp(discard control) error = %v", err)
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		t.Fatalf("Close(discard control temporary) error = %v", err)
	}
	unrelatedPath := filepath.Join(stage.discardsDirectory(), "operator-note")
	if err := os.WriteFile(unrelatedPath, []byte("preserve"), 0o600); err != nil {
		t.Fatalf("WriteFile(unrelated discard artifact) error = %v", err)
	}

	recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
	if _, err := os.Stat(temporaryPath); !os.IsNotExist(err) {
		t.Fatalf("Stat(owned discard temporary) error = %v, want absence", err)
	}
	if _, err := os.Stat(unrelatedPath); err != nil {
		t.Fatalf("Stat(unrelated discard artifact) error = %v, want preserved", err)
	}
	assertTransactionStageCleanupRequired(t, recovered.BeginTransaction(context.Background(), "discard-control-unrelated-artifact"))
	assertTransactionStageCleanupRequired(t, recovered.ReconcileDiscardControls(context.Background()))
}

func TestCommittedTransactionStageRecoveryPreservesUnexpectedDiscardControls(t *testing.T) {
	for _, tt := range []struct {
		name   string
		create func(*testing.T, *CommittedTransactionStage) string
	}{
		{
			name: "corrupt final",
			create: func(t *testing.T, stage *CommittedTransactionStage) string {
				t.Helper()
				path := filepath.Join(stage.discardsDirectory(), strings.Repeat("a", transactionStageKeyBytes)+"-"+strings.Repeat("b", transactionStageKeyBytes)+".json")
				if err := os.WriteFile(path, []byte("not-json"), 0o600); err != nil {
					t.Fatalf("WriteFile(corrupt final) error = %v", err)
				}
				return path
			},
		},
		{
			name: "oversized final",
			create: func(t *testing.T, stage *CommittedTransactionStage) string {
				t.Helper()
				key := strings.Repeat("a", transactionStageKeyBytes)
				instanceID := strings.Repeat("b", transactionStageKeyBytes)
				path := filepath.Join(stage.discardsDirectory(), key+"-"+instanceID+".json")
				payload := `{"version":1,"transaction_key":"` + key + `","instance_id":"` + instanceID + `","state":"discarded"}` + strings.Repeat(" ", transactionStageDiscardControlMaximumBytes)
				if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
					t.Fatalf("WriteFile(oversized final) error = %v", err)
				}
				return path
			},
		},
		{
			name: "temporary lookalike",
			create: func(t *testing.T, stage *CommittedTransactionStage) string {
				t.Helper()
				path := filepath.Join(stage.discardsDirectory(), ".stage.tmp-12345678901")
				if err := os.WriteFile(path, []byte("preserve"), 0o600); err != nil {
					t.Fatalf("WriteFile(temporary lookalike) error = %v", err)
				}
				return path
			},
		},
		{
			name: "directory",
			create: func(t *testing.T, stage *CommittedTransactionStage) string {
				t.Helper()
				path := filepath.Join(stage.discardsDirectory(), "operator-directory")
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatalf("Mkdir(discard artifact) error = %v", err)
				}
				return path
			},
		},
		{
			name: "symlink",
			create: func(t *testing.T, stage *CommittedTransactionStage) string {
				t.Helper()
				target := filepath.Join(stage.discardsDirectory(), "operator-target")
				if err := os.WriteFile(target, []byte("preserve"), 0o600); err != nil {
					t.Fatalf("WriteFile(symlink target) error = %v", err)
				}
				path := filepath.Join(stage.discardsDirectory(), "operator-link")
				if err := os.Symlink(target, path); err != nil {
					t.Fatalf("Symlink(discard artifact) error = %v", err)
				}
				return path
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			stage := newTestTransactionStage(t, root, testTransactionStageLimits())
			path := tt.create(t, stage)
			recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
			assertTransactionStageCleanupRequired(t, recovered.BeginTransaction(context.Background(), "unexpected-discard-artifact"))
			if _, err := os.Lstat(path); err != nil {
				t.Fatalf("Lstat(discard artifact) error = %v, want preserved", err)
			}
		})
	}
}

func TestCommittedTransactionStageDiscardRetirementCrashMatrix(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const transactionID = "discard-retirement-crash"
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
	cleanupFault := &transactionStageFault{
		operation: "remove_all",
		match: func(path string) bool {
			return path == stage.transactionDirectory(key)
		},
		err: errors.New("injected discard transaction cleanup failure"),
	}
	installTransactionStageFault(stage, cleanupFault)
	if err := stage.AbortTransaction(context.Background(), transactionID); err == nil {
		t.Fatal("AbortTransaction() error = nil, want cleanup failure")
	}
	if !cleanupFault.fired() {
		t.Fatal("discard transaction cleanup fault was not exercised")
	}

	_ = newTestTransactionStage(t, root, testTransactionStageLimits())
	assertPrivateStageCounts(t, root, 0, 0)
	assertDiscardControlFinalCount(t, root, 0)
}

func TestCommittedTransactionStageDiscardFinalRetirementFailuresPoisonRoot(t *testing.T) {
	for _, tt := range []struct {
		name      string
		operation string
		match     func(*CommittedTransactionStage, string, string) bool
	}{
		{
			name:      "final remove",
			operation: "remove",
			match: func(_ *CommittedTransactionStage, marker, path string) bool {
				return path == marker
			},
		},
		{
			name:      "final directory sync",
			operation: "directory_sync",
			match: func(stage *CommittedTransactionStage, _ string, path string) bool {
				return path == stage.discardsDirectory()
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			stage := newTestTransactionStage(t, root, testTransactionStageLimits())
			transactionID := "discard-final-retirement-" + strings.ReplaceAll(tt.name, " ", "-")
			if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
				t.Fatalf("BeginTransaction() error = %v", err)
			}
			key, err := transactionStageKey(transactionID)
			if err != nil {
				t.Fatalf("transactionStageKey() error = %v", err)
			}
			stage.mu.Lock()
			marker := stage.discardIntentPath(key, stage.entries[key].manifest.instanceID())
			stage.mu.Unlock()
			fault := &transactionStageFault{
				operation: tt.operation,
				match: func(path string) bool {
					return tt.match(stage, marker, path)
				},
				err: errors.New("injected discard final retirement failure"),
			}
			installTransactionStageFault(stage, fault)
			abortErr := stage.AbortTransaction(context.Background(), transactionID)
			assertTransactionStageCleanupRequired(t, abortErr)
			if !fault.fired() {
				t.Fatal("discard final retirement fault was not exercised")
			}
			assertTransactionStageCleanupRequired(t, stage.BeginTransaction(context.Background(), transactionID+"-new"))
			if err := stage.ReconcileDiscardControls(context.Background()); err != nil {
				t.Fatalf("ReconcileDiscardControls() error = %v", err)
			}
			if _, err := os.Stat(marker); !os.IsNotExist(err) {
				t.Fatalf("Stat(discard intent) error = %v, want retired marker", err)
			}
			if err := stage.BeginTransaction(context.Background(), transactionID+"-new"); err != nil {
				t.Fatalf("BeginTransaction() after reconciliation error = %v", err)
			}
		})
	}
}

func TestCommittedTransactionStageDiscardControlCleanupFailurePoisonsRoot(t *testing.T) {
	root := t.TempDir()
	stage := newTestTransactionStage(t, root, testTransactionStageLimits())
	const doomedTransactionID = "discard-control-poison-doomed"
	const committingTransactionID = "discard-control-poison-commit"
	const recoveredTransactionID = "discard-control-poison-recovered"
	if err := stage.BeginTransaction(context.Background(), doomedTransactionID); err != nil {
		t.Fatalf("BeginTransaction(doomed) error = %v", err)
	}
	stageCompleteChunk(t, stage, committingTransactionID)
	stageCompleteChunk(t, stage, recoveredTransactionID)
	recoveredKey, err := transactionStageKey(recoveredTransactionID)
	if err != nil {
		t.Fatalf("transactionStageKey(recovered) error = %v", err)
	}
	stage.mu.Lock()
	stage.entries[recoveredKey].manifest.State = transactionStageStateSealed
	stage.entries[recoveredKey].status = stageStatusRecoveryHeld
	stage.mu.Unlock()
	doomedKey, err := transactionStageKey(doomedTransactionID)
	if err != nil {
		t.Fatalf("transactionStageKey(doomed) error = %v", err)
	}
	cleanupFault := &transactionStageFault{
		operation: "remove_all",
		match: func(path string) bool {
			return path == stage.transactionDirectory(doomedKey)
		},
		err: errors.New("injected discard control cleanup failure"),
	}
	installTransactionStageFault(stage, cleanupFault)
	if err := stage.AbortTransaction(context.Background(), doomedTransactionID); err == nil {
		t.Fatal("AbortTransaction(doomed) error = nil, want cleanup failure")
	}
	if !cleanupFault.fired() {
		t.Fatal("discard control cleanup fault was not exercised")
	}

	beginErr := stage.BeginTransaction(context.Background(), "discard-control-poison-new")
	assertTransactionStageCleanupRequired(t, beginErr)
	appendErr := stage.AppendChunk(context.Background(), committingTransactionID, 1, strings.NewReader("more"))
	assertTransactionStageCleanupRequired(t, appendErr)
	_, commitErr := stage.CommitTransaction(context.Background(), committingTransactionID, &recordingTransactionReceiver{})
	assertTransactionStageCleanupRequired(t, commitErr)
	admitErr := stage.AdmitRecoveredTransaction(context.Background(), recoveredTransactionID)
	assertTransactionStageCleanupRequired(t, admitErr)

	if err := stage.AbortTransaction(context.Background(), doomedTransactionID); err != nil {
		t.Fatalf("AbortTransaction(doomed) cleanup retry error = %v", err)
	}
	if err := stage.ReconcileDiscardControls(context.Background()); err != nil {
		t.Fatalf("ReconcileDiscardControls() error = %v", err)
	}
	if err := stage.BeginTransaction(context.Background(), "discard-control-poison-new"); err != nil {
		t.Fatalf("BeginTransaction() after reconciliation error = %v", err)
	}
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

func TestCommittedTransactionStageDiscardControlTempCleanupRetainsSlotUntilReconciled(t *testing.T) {
	root := t.TempDir()
	limits := testTransactionStageLimits()
	limits.MaxStagedTransactions = 1
	stage := newTestTransactionStage(t, root, limits)
	const firstTransactionID = "discard-temp-retains-slot-first"
	key, err := transactionStageKey(firstTransactionID)
	if err != nil {
		t.Fatalf("transactionStageKey() error = %v", err)
	}
	writeFault := &transactionStageFault{
		operation: "write",
		match: func(path string) bool {
			return isTransactionStageDiscardControlTemporary(stage, key, path)
		},
		err: errors.New("injected discard-control temporary write failure"),
	}
	removeFault := &transactionStageFault{
		operation: "remove",
		match: func(path string) bool {
			return isTransactionStageDiscardControlTemporary(stage, key, path)
		},
		err: errors.New("injected discard-control temporary removal failure"),
	}
	syncRecorder := &transactionStageFault{
		operation: "directory_sync",
		match: func(path string) bool {
			return path == stage.discardsDirectory()
		},
	}
	if err := stage.BeginTransaction(context.Background(), firstTransactionID); err != nil {
		t.Fatalf("BeginTransaction(first) error = %v", err)
	}
	installTransactionStageFaults(stage, writeFault, removeFault, syncRecorder)
	assertTransactionStageCleanupRequired(t, stage.AbortTransaction(context.Background(), firstTransactionID))
	if !writeFault.fired() || !removeFault.fired() {
		t.Fatalf("discard temporary faults fired = (%t, %t), want both", writeFault.fired(), removeFault.fired())
	}
	assertTransactionStageRootPoisoned(t, stage)
	assertTransactionStageControlState(t, stage, firstTransactionID, transactionStageControlTemporary)
	assertTransactionStageControlCount(t, stage, 1)
	assertDiscardControlArtifactCounts(t, root, 1, 0)

	for _, transactionID := range []string{"discard-temp-retains-slot-second", "discard-temp-retains-slot-third"} {
		assertTransactionStageCleanupRequired(t, stage.BeginTransaction(context.Background(), transactionID))
		assertTransactionStageControlCount(t, stage, 1)
		assertDiscardControlArtifactCounts(t, root, 1, 0)
	}

	syncCallsBeforeReconcile := syncRecorder.callCount()
	if err := stage.ReconcileDiscardControls(context.Background()); err != nil {
		t.Fatalf("ReconcileDiscardControls() error = %v", err)
	}
	if got := syncRecorder.callCount(); got <= syncCallsBeforeReconcile {
		t.Fatalf("discard directory sync calls after reconciliation = %d, want more than %d", got, syncCallsBeforeReconcile)
	}
	assertTransactionStageRootClean(t, stage)
	assertTransactionStageControlCount(t, stage, 0)
	assertDiscardControlArtifactCounts(t, root, 0, 0)
	if err := stage.BeginTransaction(context.Background(), "discard-temp-retains-slot-after-reconcile"); err != nil {
		t.Fatalf("BeginTransaction() after reconciliation error = %v", err)
	}
	if err := stage.AbortTransaction(context.Background(), "discard-temp-retains-slot-after-reconcile"); err != nil {
		t.Fatalf("AbortTransaction() after reconciliation error = %v", err)
	}
}

func TestCommittedTransactionStageDiscardControlTempRemovalRequiresDirectorySync(t *testing.T) {
	root := t.TempDir()
	limits := testTransactionStageLimits()
	limits.MaxStagedTransactions = 1
	stage := newTestTransactionStage(t, root, limits)
	const transactionID = "discard-temp-removal-needs-directory-sync"
	key, err := transactionStageKey(transactionID)
	if err != nil {
		t.Fatalf("transactionStageKey() error = %v", err)
	}
	writeFault := &transactionStageFault{
		operation: "write",
		match: func(path string) bool {
			return isTransactionStageDiscardControlTemporary(stage, key, path)
		},
		err: errors.New("injected discard-control temporary write failure"),
	}
	syncFault := &transactionStageFault{
		operation: "directory_sync",
		match: func(path string) bool {
			return path == stage.discardsDirectory()
		},
		err: errors.New("injected discard-control temporary cleanup sync failure"),
	}
	if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
		t.Fatalf("BeginTransaction() error = %v", err)
	}
	installTransactionStageFaults(stage, writeFault, syncFault)
	assertTransactionStageCleanupRequired(t, stage.AbortTransaction(context.Background(), transactionID))
	if !writeFault.fired() || !syncFault.fired() {
		t.Fatalf("discard temporary faults fired = (%t, %t), want both", writeFault.fired(), syncFault.fired())
	}
	assertTransactionStageRootPoisoned(t, stage)
	assertTransactionStageControlState(t, stage, transactionID, transactionStageControlTemporary)
	assertTransactionStageControlCount(t, stage, 1)
	assertDiscardControlArtifactCounts(t, root, 0, 0)
	assertTransactionStageCleanupRequired(t, stage.BeginTransaction(context.Background(), "discard-temp-removal-needs-directory-sync-next"))

	if err := stage.ReconcileDiscardControls(context.Background()); err != nil {
		t.Fatalf("ReconcileDiscardControls() error = %v", err)
	}
	assertTransactionStageRootClean(t, stage)
	assertTransactionStageControlCount(t, stage, 0)
	if err := stage.BeginTransaction(context.Background(), "discard-temp-removal-needs-directory-sync-after"); err != nil {
		t.Fatalf("BeginTransaction() after reconciliation error = %v", err)
	}
	if err := stage.AbortTransaction(context.Background(), "discard-temp-removal-needs-directory-sync-after"); err != nil {
		t.Fatalf("AbortTransaction() after reconciliation error = %v", err)
	}
}

func TestCommittedTransactionStageRecoveredOverCapacityCannotClearPoisonOrDeliver(t *testing.T) {
	root := t.TempDir()
	creationLimits := testTransactionStageLimits()
	creationLimits.MaxStagedTransactions = 2
	created := newTestTransactionStage(t, root, creationLimits)
	transactionIDs := []string{"recovered-over-capacity-first", "recovered-over-capacity-second"}
	for _, transactionID := range transactionIDs {
		sealReceiptlessTransactionStage(t, created, transactionID)
	}

	restrictedLimits := testTransactionStageLimits()
	restrictedLimits.MaxStagedTransactions = 1
	recovered := newTestTransactionStage(t, root, restrictedLimits)
	assertRecoveredTransactionStageControlMapping(t, recovered, transactionIDs, 1)
	assertTransactionStageRootPoisoned(t, recovered)
	for attempt := 0; attempt < 3; attempt++ {
		assertTransactionStageCleanupRequired(t, recovered.ReconcileDiscardControls(context.Background()))
		assertRecoveredTransactionStageControlMapping(t, recovered, transactionIDs, 1)
	}

	unreservedTransactionID := unreservedRecoveredTransactionID(t, recovered, transactionIDs)
	assertTransactionStageCleanupRequired(t, recovered.AdmitRecoveredTransaction(context.Background(), unreservedTransactionID))
	receiver := &recordingTransactionReceiver{}
	receipt, commitErr := recovered.CommitTransaction(context.Background(), unreservedTransactionID, receiver)
	assertTransactionStageCleanupRequired(t, commitErr)
	assertTransactionStageNoDeliveryEvidence(t, recovered, unreservedTransactionID, receipt, receiver)

	widened := newTestTransactionStage(t, root, creationLimits)
	assertTransactionStageRootClean(t, widened)
	assertRecoveredTransactionStageControlMapping(t, widened, transactionIDs, 0)
	for _, transactionID := range transactionIDs {
		receiver := &recordingTransactionReceiver{}
		if _, err := widened.CommitTransaction(context.Background(), transactionID, receiver); !errors.Is(err, ErrTransactionStageRecoveryRequired) {
			t.Fatalf("CommitTransaction(%q) before explicit admission error = %v, want ErrTransactionStageRecoveryRequired", transactionID, err)
		}
		if got := len(receiver.transactions); got != 0 {
			t.Fatalf("receiver calls for %q before explicit admission = %d, want 0", transactionID, got)
		}
	}
}

func TestCommittedTransactionStageRecoveredControlReservationRestartMatrix(t *testing.T) {
	tests := []struct {
		name                string
		transactionIDs      []string
		creationLimit       int64
		reopenLimits        []int64
		wantMissingControls []int
		receiptBacked       bool
	}{
		{
			name:                "one to one",
			transactionIDs:      []string{"recovered-restart-one"},
			creationLimit:       1,
			reopenLimits:        []int64{1},
			wantMissingControls: []int{0},
		},
		{
			name:                "two to two",
			transactionIDs:      []string{"recovered-restart-two-first", "recovered-restart-two-second"},
			creationLimit:       2,
			reopenLimits:        []int64{2},
			wantMissingControls: []int{0},
		},
		{
			name:                "two to one",
			transactionIDs:      []string{"recovered-restart-over-capacity-first", "recovered-restart-over-capacity-second"},
			creationLimit:       2,
			reopenLimits:        []int64{1},
			wantMissingControls: []int{1},
		},
		{
			name:                "two to one to two",
			transactionIDs:      []string{"recovered-restart-expand-first", "recovered-restart-expand-second"},
			creationLimit:       2,
			reopenLimits:        []int64{1, 2},
			wantMissingControls: []int{1, 0},
		},
		{
			name:                "receipt backed residue",
			transactionIDs:      []string{"recovered-restart-receiptless"},
			creationLimit:       2,
			reopenLimits:        []int64{1},
			wantMissingControls: []int{0},
			receiptBacked:       true,
		},
		{
			name:                "reversed transaction key order",
			transactionIDs:      []string{"recovered-restart-order-second", "recovered-restart-order-first"},
			creationLimit:       2,
			reopenLimits:        []int64{2},
			wantMissingControls: []int{0},
		},
		{
			name:                "repeated reopen and reconcile",
			transactionIDs:      []string{"recovered-restart-repeat-first", "recovered-restart-repeat-second"},
			creationLimit:       2,
			reopenLimits:        []int64{2, 2, 2},
			wantMissingControls: []int{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			creationLimits := testTransactionStageLimits()
			creationLimits.MaxStagedTransactions = tt.creationLimit
			created := newTestTransactionStage(t, root, creationLimits)
			const receiptBackedTransactionID = "recovered-restart-durable-receipt"
			if tt.receiptBacked {
				stageCompleteChunk(t, created, receiptBackedTransactionID)
				if _, err := created.CommitTransaction(context.Background(), receiptBackedTransactionID, &recordingTransactionReceiver{}); err != nil {
					t.Fatalf("CommitTransaction(receipt backed) error = %v", err)
				}
			}
			for _, transactionID := range tt.transactionIDs {
				sealReceiptlessTransactionStage(t, created, transactionID)
			}

			for index, maximum := range tt.reopenLimits {
				limits := testTransactionStageLimits()
				limits.MaxStagedTransactions = maximum
				recovered := newTestTransactionStage(t, root, limits)
				wantMissing := tt.wantMissingControls[index]
				assertRecoveredTransactionStageControlMapping(t, recovered, tt.transactionIDs, wantMissing)
				if wantMissing != 0 {
					assertTransactionStageRootPoisoned(t, recovered)
					for attempt := 0; attempt < 2; attempt++ {
						assertTransactionStageCleanupRequired(t, recovered.ReconcileDiscardControls(context.Background()))
						assertRecoveredTransactionStageControlMapping(t, recovered, tt.transactionIDs, wantMissing)
					}
				} else {
					assertTransactionStageRootClean(t, recovered)
					if err := recovered.ReconcileDiscardControls(context.Background()); err != nil {
						t.Fatalf("ReconcileDiscardControls() error = %v", err)
					}
				}
				if tt.receiptBacked {
					receipt, err := recovered.Receipt(receiptBackedTransactionID)
					if err != nil {
						t.Fatalf("Receipt(receipt backed) error = %v", err)
					}
					if _, err := receipt.Acknowledgement(); err != nil {
						t.Fatalf("receipt-backed acknowledgement error = %v", err)
					}
				}
			}
		})
	}
}

func TestCommittedTransactionStageDiscardControlTemporaryCrashMatrix(t *testing.T) {
	tests := []struct {
		name          string
		faults        func(*CommittedTransactionStage, string) []*transactionStageFault
		wantCleanup   bool
		wantTemporary int
		wantFinal     int
		wantControl   bool
		wantState     transactionStageControlState
	}{
		{
			name: "temporary create",
			faults: func(stage *CommittedTransactionStage, _ string) []*transactionStageFault {
				return []*transactionStageFault{{
					operation: "create_temp",
					match: func(path string) bool {
						return path == stage.discardsDirectory()
					},
					err: errors.New("injected discard-control temporary creation failure"),
				}}
			},
		},
		{
			name: "temporary write",
			faults: func(stage *CommittedTransactionStage, key string) []*transactionStageFault {
				return discardControlTemporaryFaults(stage, key, "write", errors.New("injected discard-control temporary write failure"))
			},
		},
		{
			name: "temporary file sync",
			faults: func(stage *CommittedTransactionStage, key string) []*transactionStageFault {
				return discardControlTemporaryFaults(stage, key, "file_sync", errors.New("injected discard-control temporary file sync failure"))
			},
		},
		{
			name: "temporary close",
			faults: func(stage *CommittedTransactionStage, key string) []*transactionStageFault {
				return discardControlTemporaryFaults(stage, key, "close", errors.New("injected discard-control temporary close failure"))
			},
		},
		{
			name: "temporary rename",
			faults: func(stage *CommittedTransactionStage, key string) []*transactionStageFault {
				return []*transactionStageFault{{
					operation: "rename",
					match: func(path string) bool {
						return path == stage.discardIntentPath(key, stageEntryInstanceID(t, stage, key))
					},
					err: errors.New("injected discard-control temporary rename failure"),
				}}
			},
		},
		{
			name: "temporary remove",
			faults: func(stage *CommittedTransactionStage, key string) []*transactionStageFault {
				return []*transactionStageFault{
					discardControlTemporaryFault(stage, key, "write", errors.New("injected discard-control temporary write failure")),
					discardControlTemporaryFault(stage, key, "remove", errors.New("injected discard-control temporary remove failure")),
				}
			},
			wantCleanup:   true,
			wantTemporary: 1,
			wantControl:   true,
			wantState:     transactionStageControlTemporary,
		},
		{
			name: "temporary parent sync",
			faults: func(stage *CommittedTransactionStage, key string) []*transactionStageFault {
				return []*transactionStageFault{
					discardControlTemporaryFault(stage, key, "write", errors.New("injected discard-control temporary write failure")),
					{
						operation: "directory_sync",
						match: func(path string) bool {
							return path == stage.discardsDirectory()
						},
						err: errors.New("injected discard-control temporary parent sync failure"),
					},
				}
			},
			wantCleanup: true,
			wantControl: true,
			wantState:   transactionStageControlTemporary,
		},
		{
			name: "transaction remove",
			faults: func(stage *CommittedTransactionStage, key string) []*transactionStageFault {
				return []*transactionStageFault{{
					operation: "remove_all",
					match: func(path string) bool {
						return path == stage.transactionDirectory(key)
					},
					err: errors.New("injected transaction discard cleanup removal failure"),
				}}
			},
			wantCleanup: true,
			wantFinal:   1,
			wantControl: true,
			wantState:   transactionStageControlFinal,
		},
		{
			name: "transaction parent sync",
			faults: func(stage *CommittedTransactionStage, key string) []*transactionStageFault {
				return []*transactionStageFault{{
					operation: "directory_sync",
					match: func(path string) bool {
						return path == stage.transactionsDirectory()
					},
					err: errors.New("injected transaction discard cleanup parent sync failure"),
				}}
			},
			wantCleanup: true,
			wantFinal:   1,
			wantControl: true,
			wantState:   transactionStageControlFinal,
		},
		{
			name: "final remove",
			faults: func(stage *CommittedTransactionStage, key string) []*transactionStageFault {
				return []*transactionStageFault{{
					operation: "remove",
					match: func(path string) bool {
						return path == stage.discardIntentPath(key, stageEntryInstanceID(t, stage, key))
					},
					err: errors.New("injected discard-control final removal failure"),
				}}
			},
			wantCleanup: true,
			wantFinal:   1,
			wantControl: true,
			wantState:   transactionStageControlFinal,
		},
		{
			name: "final parent sync",
			faults: func(stage *CommittedTransactionStage, _ string) []*transactionStageFault {
				return []*transactionStageFault{{
					operation: "directory_sync",
					skip:      1,
					match: func(path string) bool {
						return path == stage.discardsDirectory()
					},
					err: errors.New("injected discard-control final parent sync failure"),
				}}
			},
			wantCleanup: true,
			wantControl: true,
			wantState:   transactionStageControlFinal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			stage := newTestTransactionStage(t, root, testTransactionStageLimits())
			transactionID := "discard-control-temporary-crash-" + strings.ReplaceAll(tt.name, " ", "-")
			key, err := transactionStageKey(transactionID)
			if err != nil {
				t.Fatalf("transactionStageKey() error = %v", err)
			}
			if err := stage.BeginTransaction(context.Background(), transactionID); err != nil {
				t.Fatalf("BeginTransaction() error = %v", err)
			}
			faults := tt.faults(stage, key)
			installTransactionStageFaults(stage, faults...)
			abortErr := stage.AbortTransaction(context.Background(), transactionID)
			if abortErr == nil {
				t.Fatal("AbortTransaction() error = nil, want injected failure")
			}
			for _, fault := range faults {
				if !fault.fired() {
					t.Fatalf("fault %+v was not exercised", fault)
				}
			}
			if tt.wantCleanup {
				assertTransactionStageCleanupRequired(t, abortErr)
				assertTransactionStageRootPoisoned(t, stage)
			} else {
				if errors.Is(abortErr, ErrTransactionStageCleanupRequired) {
					t.Fatalf("AbortTransaction() error = %v, want non-cleanup injected failure", abortErr)
				}
				assertTransactionStageRootClean(t, stage)
			}
			assertDiscardControlArtifactCounts(t, root, tt.wantTemporary, tt.wantFinal)
			if tt.wantControl {
				assertTransactionStageControlState(t, stage, transactionID, tt.wantState)
				assertTransactionStageControlCount(t, stage, 1)
			} else {
				assertTransactionStageControlCount(t, stage, 0)
			}

			receiver := &recordingTransactionReceiver{}
			receipt, commitErr := stage.CommitTransaction(context.Background(), transactionID, receiver)
			if tt.wantCleanup {
				assertTransactionStageCleanupRequired(t, commitErr)
			} else if !errors.Is(commitErr, ErrTransactionStageNotFound) {
				t.Fatalf("CommitTransaction() after durable discard error = %v, want ErrTransactionStageNotFound", commitErr)
			}
			assertTransactionStageNoDeliveryEvidence(t, stage, transactionID, receipt, receiver)

			recovered := newTestTransactionStage(t, root, testTransactionStageLimits())
			restartedReceiver := &recordingTransactionReceiver{}
			restartedReceipt, restartedErr := recovered.CommitTransaction(context.Background(), transactionID, restartedReceiver)
			if !errors.Is(restartedErr, ErrTransactionStageNotFound) {
				t.Fatalf("CommitTransaction() after restart error = %v, want ErrTransactionStageNotFound", restartedErr)
			}
			assertTransactionStageNoDeliveryEvidence(t, recovered, transactionID, restartedReceipt, restartedReceiver)
		})
	}
}

func sealReceiptlessTransactionStage(t *testing.T, stage *CommittedTransactionStage, transactionID string) {
	t.Helper()
	stageCompleteChunk(t, stage, transactionID)
	if _, err := stage.CommitTransaction(context.Background(), transactionID, transactionReceiverFunc(func(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error) {
		return DownstreamTransactionReceipt{}, errors.New("injected receiver failure")
	})); err == nil {
		t.Fatalf("CommitTransaction(%q) error = nil, want sealed receipt-less transaction", transactionID)
	}
}

func assertTransactionStageRootPoisoned(t *testing.T, stage *CommittedTransactionStage) {
	t.Helper()
	stage.mu.Lock()
	cleanupErr := stage.cleanupErr
	stage.mu.Unlock()
	if cleanupErr == nil {
		t.Fatal("transaction stage root cleanup error = nil, want poison")
	}
}

func assertTransactionStageRootClean(t *testing.T, stage *CommittedTransactionStage) {
	t.Helper()
	stage.mu.Lock()
	cleanupErr := stage.cleanupErr
	stage.mu.Unlock()
	if cleanupErr != nil {
		t.Fatalf("transaction stage root cleanup error = %v, want nil", cleanupErr)
	}
}

func assertTransactionStageControlCount(t *testing.T, stage *CommittedTransactionStage, want int) {
	t.Helper()
	stage.mu.Lock()
	got := len(stage.controls)
	stage.mu.Unlock()
	if got != want {
		t.Fatalf("transaction stage control count = %d, want %d", got, want)
	}
}

func assertTransactionStageControlState(t *testing.T, stage *CommittedTransactionStage, transactionID string, want transactionStageControlState) {
	t.Helper()
	key, err := transactionStageKey(transactionID)
	if err != nil {
		t.Fatalf("transactionStageKey() error = %v", err)
	}
	stage.mu.Lock()
	entry := stage.entries[key]
	var control *transactionStageControl
	if entry != nil {
		control = stage.controls[transactionStageControlKey(key, entry.manifest.instanceID())]
	}
	stage.mu.Unlock()
	if entry == nil {
		t.Fatalf("transaction stage entry for %q = nil, want retained entry", transactionID)
	}
	if control == nil {
		t.Fatalf("transaction stage control for %q = nil, want state %v", transactionID, want)
	}
	if control.state != want {
		t.Fatalf("transaction stage control state for %q = %v, want %v", transactionID, control.state, want)
	}
}

func assertDiscardControlArtifactCounts(t *testing.T, root string, wantTemporary, wantFinal int) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "discards"))
	if err != nil {
		t.Fatalf("ReadDir(discards) error = %v", err)
	}
	var temporary, final int
	for _, entry := range entries {
		if isOwnedTransactionStageDiscardTemporary(entry.Name()) {
			temporary++
		}
		if strings.HasSuffix(entry.Name(), ".json") {
			final++
		}
	}
	if temporary != wantTemporary || final != wantFinal {
		t.Fatalf("discard control artifacts = %d temporary, %d final; want %d temporary, %d final (%v)", temporary, final, wantTemporary, wantFinal, entries)
	}
}

func assertRecoveredTransactionStageControlMapping(t *testing.T, stage *CommittedTransactionStage, transactionIDs []string, wantMissing int) {
	t.Helper()
	stage.mu.Lock()
	defer stage.mu.Unlock()
	if got := len(stage.entries); got != len(transactionIDs) {
		t.Fatalf("recovered entries = %d, want %d", got, len(transactionIDs))
	}
	matchedControls := make(map[string]bool, len(transactionIDs))
	missing := 0
	for _, transactionID := range transactionIDs {
		key, err := transactionStageKey(transactionID)
		if err != nil {
			t.Fatalf("transactionStageKey(%q) error = %v", transactionID, err)
		}
		entry := stage.entries[key]
		if entry == nil {
			t.Fatalf("recovered entry for %q = nil", transactionID)
		}
		if entry.status != stageStatusRecoveryHeld || entry.manifest.State != transactionStageStateSealed {
			t.Fatalf("recovered entry for %q = %#v, want sealed recovery-held entry", transactionID, entry)
		}
		controlKey := transactionStageControlKey(key, entry.manifest.instanceID())
		control := stage.controls[controlKey]
		if control == nil {
			missing++
			continue
		}
		if control.state != transactionStageControlReserved || control.transactionKey != key || control.instanceID != entry.manifest.instanceID() {
			t.Fatalf("recovered control for %q = %#v, want exact reserved generation", transactionID, control)
		}
		matchedControls[controlKey] = true
	}
	if missing != wantMissing {
		t.Fatalf("recovered entries without exact reserved controls = %d, want %d", missing, wantMissing)
	}
	if got, want := len(stage.controls), len(transactionIDs)-wantMissing; got != want {
		t.Fatalf("recovered control count = %d, want %d", got, want)
	}
	for controlKey := range stage.controls {
		if !matchedControls[controlKey] {
			t.Fatalf("unexpected recovered control %q", controlKey)
		}
	}
}

func unreservedRecoveredTransactionID(t *testing.T, stage *CommittedTransactionStage, transactionIDs []string) string {
	t.Helper()
	stage.mu.Lock()
	defer stage.mu.Unlock()
	for _, transactionID := range transactionIDs {
		key, err := transactionStageKey(transactionID)
		if err != nil {
			t.Fatalf("transactionStageKey(%q) error = %v", transactionID, err)
		}
		entry := stage.entries[key]
		if entry != nil && stage.controls[transactionStageControlKey(key, entry.manifest.instanceID())] == nil {
			return transactionID
		}
	}
	t.Fatal("unreserved recovered transaction = none, want one")
	return ""
}

func assertTransactionStageNoDeliveryEvidence(t *testing.T, stage *CommittedTransactionStage, transactionID string, receipt TransactionReceipt, receiver *recordingTransactionReceiver) {
	t.Helper()
	if got := len(receiver.transactions); got != 0 {
		t.Fatalf("receiver calls for %q = %d, want 0", transactionID, got)
	}
	key, err := transactionStageKey(transactionID)
	if err != nil {
		t.Fatalf("transactionStageKey() error = %v", err)
	}
	if _, err := os.Stat(stage.receiptPath(key)); !os.IsNotExist(err) {
		t.Fatalf("Stat(receipt path) error = %v, want absence", err)
	}
	if _, err := stage.Receipt(transactionID); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Fatalf("Receipt(%q) error = %v, want unavailable", transactionID, err)
	}
	if _, err := receipt.Acknowledgement(); !errors.Is(err, ErrTransactionReceiptUnavailable) {
		t.Fatalf("receipt acknowledgement for %q error = %v, want unavailable", transactionID, err)
	}
}

func discardControlTemporaryFaults(stage *CommittedTransactionStage, key, operation string, err error) []*transactionStageFault {
	return []*transactionStageFault{discardControlTemporaryFault(stage, key, operation, err)}
}

func discardControlTemporaryFault(stage *CommittedTransactionStage, key, operation string, err error) *transactionStageFault {
	return &transactionStageFault{
		operation: operation,
		match: func(path string) bool {
			return isTransactionStageDiscardControlTemporary(stage, key, path)
		},
		err: err,
	}
}

func stageEntryInstanceID(t *testing.T, stage *CommittedTransactionStage, key string) string {
	t.Helper()
	stage.mu.Lock()
	entry := stage.entries[key]
	stage.mu.Unlock()
	if entry == nil {
		t.Fatalf("transaction stage entry for key %q = nil", key)
	}
	return entry.manifest.instanceID()
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
	calls     int
	wasFired  bool
}

type transactionStageFaultApplier interface {
	apply(string, string) error
}

type transactionStageFaultSet struct {
	faults []*transactionStageFault
}

func (s transactionStageFaultSet) apply(operation, path string) error {
	for _, fault := range s.faults {
		if err := fault.apply(operation, path); err != nil {
			return err
		}
	}
	return nil
}

func (f *transactionStageFault) apply(operation, path string) error {
	if f == nil || operation != f.operation || f.match == nil || !f.match(path) {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
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

func (f *transactionStageFault) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func installTransactionStageFault(stage *CommittedTransactionStage, fault *transactionStageFault) {
	installTransactionStageFaults(stage, fault)
}

func installTransactionStageFaults(stage *CommittedTransactionStage, faults ...*transactionStageFault) {
	installTransactionStageFaultApplier(stage, transactionStageFaultSet{faults: faults})
}

func installTransactionStageFaultApplier(stage *CommittedTransactionStage, fault transactionStageFaultApplier) {
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
	stage.storage.remove = func(path string) error {
		if err := fault.apply("remove", path); err != nil {
			return err
		}
		return base.remove(path)
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
	fault transactionStageFaultApplier
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

func (f *faultedTransactionStageFile) Close() error {
	if err := f.fault.apply("close", f.Name()); err != nil {
		return err
	}
	return f.transactionStageFile.Close()
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

func isTransactionStageDiscardControlTemporary(stage *CommittedTransactionStage, key, path string) bool {
	return isTransactionStageStateTemporary(stage, key, path) ||
		(strings.Contains(filepath.ToSlash(path), "/discards/") && isOwnedTransactionStageDiscardTemporary(filepath.Base(path)))
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
