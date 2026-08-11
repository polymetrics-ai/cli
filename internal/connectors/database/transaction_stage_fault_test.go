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
