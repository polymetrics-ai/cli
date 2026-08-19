package database_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/database"
)

func TestFullOverwriteReceiptV1IsDeterministicAndPayloadFree(t *testing.T) {
	publishedAt := time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC)
	first, err := database.NewFullOverwriteReceiptV1(receiptHash("plan"), receiptHash("checkpoint"), receiptHash("content"), 42, publishedAt)
	if err != nil {
		t.Fatal(err)
	}
	second, err := database.NewFullOverwriteReceiptV1(receiptHash("plan"), receiptHash("checkpoint"), receiptHash("content"), 42, publishedAt.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || first.Records != 42 || first.PublishedAt != publishedAt {
		t.Fatalf("receipt = %#v; deterministic evidence or actual publish time was lost", first)
	}
}

func TestFullOverwriteReceiptV1RefusesMalformedHashesBeforeStore(t *testing.T) {
	if _, err := database.NewFullOverwriteReceiptV1("not-a-hash", receiptHash("checkpoint"), receiptHash("content"), 1, time.Now()); !errors.Is(err, database.ErrFullOverwriteReceiptInvalid) {
		t.Fatalf("NewFullOverwriteReceiptV1(malformed) error = %T %v, want ErrFullOverwriteReceiptInvalid", err, err)
	}
}

func TestFullOverwriteReceiptV1RejectsNegativeRecordCount(t *testing.T) {
	if _, err := database.NewFullOverwriteReceiptV1(receiptHash("plan"), receiptHash("checkpoint"), receiptHash("content"), -1, time.Now()); !errors.Is(err, database.ErrFullOverwriteReceiptInvalid) {
		t.Fatalf("NewFullOverwriteReceiptV1(negative records) error = %T %v, want ErrFullOverwriteReceiptInvalid", err, err)
	}
}

func receiptHash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
