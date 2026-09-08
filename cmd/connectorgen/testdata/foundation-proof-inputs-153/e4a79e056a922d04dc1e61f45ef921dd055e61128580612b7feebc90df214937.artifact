package database

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

var ErrFullOverwriteReceiptInvalid = errors.New("database full-overwrite receipt is invalid")

// FullOverwriteReceiptV1 is connector-neutral durable evidence for a single
// published replacement. It intentionally contains content identities rather
// than a database handle, SQL, source record, or checkpoint token. Native bulk
// applies persist and reconcile this exact contract in their own transaction.
type FullOverwriteReceiptV1 struct {
	ID             string
	PlanHash       string
	CheckpointHash string
	ContentHash    string
	Records        int64
	PublishedAt    time.Time
}

func NewFullOverwriteReceiptV1(planHash, checkpointHash, contentHash string, records int64, publishedAt time.Time) (FullOverwriteReceiptV1, error) {
	if !validFullOverwriteHash(planHash) || !validFullOverwriteHash(checkpointHash) || !validFullOverwriteHash(contentHash) || records < 0 || publishedAt.IsZero() {
		return FullOverwriteReceiptV1{}, ErrFullOverwriteReceiptInvalid
	}
	digest := sha256.Sum256([]byte("full-overwrite-receipt-v1\x00" + planHash + "\x00" + checkpointHash + "\x00" + contentHash))
	receipt := FullOverwriteReceiptV1{
		ID:             "replace-" + hex.EncodeToString(digest[:]),
		PlanHash:       planHash,
		CheckpointHash: checkpointHash,
		ContentHash:    contentHash,
		Records:        records,
		PublishedAt:    publishedAt.UTC(),
	}
	if err := receipt.Validate(); err != nil {
		return FullOverwriteReceiptV1{}, err
	}
	return receipt, nil
}

func (r FullOverwriteReceiptV1) Validate() error {
	if !strings.HasPrefix(r.ID, "replace-") || len(r.ID) != len("replace-")+sha256.Size*2 || !validFullOverwriteHash(r.PlanHash) || !validFullOverwriteHash(r.CheckpointHash) || !validFullOverwriteHash(r.ContentHash) || r.Records < 0 || r.PublishedAt.IsZero() {
		return ErrFullOverwriteReceiptInvalid
	}
	return nil
}

func validFullOverwriteHash(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
