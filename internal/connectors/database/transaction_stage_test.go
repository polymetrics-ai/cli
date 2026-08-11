package database

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	if receipt.TransactionKey == "" || receipt.ContentDigest == "" {
		t.Fatalf("receipt = %#v, want immutable whole-transaction identity and digest", receipt)
	}
	acknowledgement, err := receipt.Acknowledgement()
	if err != nil {
		t.Fatalf("receipt acknowledgement after commit error = %v", err)
	}
	if acknowledgement.Sink != "test-sink" || acknowledgement.AcknowledgedAt.IsZero() {
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
