package database

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"polymetrics.ai/internal/warehouse"
)

const (
	changeDeliveryBaselineDomain      = "polymetrics-change-delivery-baseline-v1"
	changeDeliveryBaselineCurrentFile = "current.json"
	changeDeliveryBaselineFileSuffix  = ".parquet"
	changeDeliveryBaselineManifestMax = int64(1 << 20)
)

// ChangeDeliveryBaseline is immutable receipt-bound evidence for one local
// per-destination candidate baseline. It exposes readers and opaque audit
// identifiers, never its filesystem path.
type ChangeDeliveryBaseline struct {
	dir    string
	record changeDeliveryBaselineRecord
	limit  int64
}

// WorksetIdentity returns the immutable artifact that produced this baseline.
func (b ChangeDeliveryBaseline) WorksetIdentity() string { return b.record.WorksetIdentity }

// ContentSHA256 returns the immutable source workset content binding.
func (b ChangeDeliveryBaseline) ContentSHA256() string { return b.record.ContentSHA256 }

// DeliveryID returns the opaque durable target receipt identifier that
// authorized this baseline promotion.
func (b ChangeDeliveryBaseline) DeliveryID() string { return b.record.DeliveryID }

// CommittedAt returns the confirmed target commit time in UTC.
func (b ChangeDeliveryBaseline) CommittedAt() time.Time { return b.record.CommittedAt }

// ReadCandidateBaseline reads the isolated persisted candidate without
// exposing a mutable artifact path.
func (b ChangeDeliveryBaseline) ReadCandidateBaseline(ctx context.Context, emit func(warehouse.Row) error) error {
	if ctx == nil || emit == nil || b.dir == "" || b.record.validate() != nil || !validChangeDeliveryArtifactLimit(b.limit) {
		return ErrChangeDeliveryBaselineUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	path := filepath.Join(b.dir, b.record.WorksetIdentity+changeDeliveryBaselineFileSuffix)
	hash, err := hashChangeDeliveryFile(ctx, path, b.limit)
	if err != nil || hash != b.record.CandidateBaselineSHA256 {
		return ErrChangeDeliveryBaselineUnavailable
	}
	if err := warehouse.ReadTable(ctx, path, emit); err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	return nil
}

type changeDeliveryBaselineRecord struct {
	Version                 int       `json:"version"`
	WorksetIdentity         string    `json:"workset_identity"`
	ContentSHA256           string    `json:"content_sha256"`
	CandidateBaselineSHA256 string    `json:"candidate_baseline_sha256"`
	DeliveryID              string    `json:"delivery_id"`
	CommittedAt             time.Time `json:"committed_at"`
}

func (r changeDeliveryBaselineRecord) validate() error {
	if r.Version != changeDeliveryWorksetFormatVersion || !validChangeDeliverySHA256(r.WorksetIdentity) || !validChangeDeliverySHA256(r.ContentSHA256) || !validChangeDeliverySHA256(r.CandidateBaselineSHA256) || !validOpaqueIdentityComponent(r.DeliveryID) || r.CommittedAt.IsZero() {
		return ErrChangeDeliveryBaselineUnavailable
	}
	return nil
}

// FileChangeDeliveryBaselineStore is a local durable, destination-keyed
// candidate-baseline store. Its directory names are derived hashes of the
// complete managed-target ledger key, so no workspace/connector/connection
// identifier can traverse or alias another destination's state.
type FileChangeDeliveryBaselineStore struct {
	root  string
	limit int64
	mu    sync.Mutex
}

// NewFileChangeDeliveryBaselineStore creates a store rooted at an explicitly
// caller-owned directory. The same finite artifact limit that bounded the
// workset also bounds every baseline copy and read.
func NewFileChangeDeliveryBaselineStore(root string, limit int64) (*FileChangeDeliveryBaselineStore, error) {
	if root == "" || !validChangeDeliveryArtifactLimit(limit) {
		return nil, ErrChangeDeliveryBaselineUnavailable
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, ErrChangeDeliveryBaselineUnavailable
	}
	return &FileChangeDeliveryBaselineStore{root: root, limit: limit}, nil
}

// RecordChangeDeliveryBaseline copies one sealed candidate and atomically
// points this exact destination at it. It is callable only after a confirmed
// target receipt; callers cannot manufacture a record for an arbitrary path.
func (s *FileChangeDeliveryBaselineStore) RecordChangeDeliveryBaseline(ctx context.Context, key ManagedTargetDeliveryLedgerKey, workset ChangeDeliveryWorkset, receipt DeliveryReceiptV1) error {
	if ctx == nil || s == nil || s.root == "" || !validChangeDeliveryArtifactLimit(s.limit) || key.validate() != nil || workset.manifest.validate() != nil || workset.dir == "" {
		return ErrChangeDeliveryBaselineUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if receipt.validateFor(receipt.plan) != nil || !worksetMatchesControl(workset, receipt.plan.Control()) {
		return ErrChangeDeliveryBaselineUnavailable
	}
	wantKey, err := NewManagedTargetDeliveryLedgerKey(receipt.plan.Control())
	if err != nil || wantKey != key {
		return ErrChangeDeliveryBaselineUnavailable
	}
	// Verify the source again before copying; the local baseline must never
	// become durable evidence for an artifact whose sealed bytes changed.
	workset, err = openChangeDeliveryWorkset(ctx, workset.dir, workset.Identity(), s.limit)
	if err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	dir := s.destinationDir(key)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	candidatePath := filepath.Join(dir, workset.Identity()+changeDeliveryBaselineFileSuffix)
	if err := s.ensureCandidate(ctx, workset, candidatePath); err != nil {
		return err
	}
	record := changeDeliveryBaselineRecord{
		Version:                 changeDeliveryWorksetFormatVersion,
		WorksetIdentity:         workset.Identity(),
		ContentSHA256:           workset.ContentSHA256(),
		CandidateBaselineSHA256: workset.manifest.CandidateBaselineSHA256,
		DeliveryID:              receipt.DeliveryID(),
		CommittedAt:             receipt.CommittedAt(),
	}
	if err := record.validate(); err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	if err := writeChangeDeliveryBaselineRecord(dir, record); err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	return nil
}

// Lookup reads and validates the current baseline for exactly one immutable
// destination ledger key. A corrupt pointer or artifact fails closed.
func (s *FileChangeDeliveryBaselineStore) Lookup(ctx context.Context, key ManagedTargetDeliveryLedgerKey) (ChangeDeliveryBaseline, bool, error) {
	if ctx == nil || s == nil || s.root == "" || !validChangeDeliveryArtifactLimit(s.limit) || key.validate() != nil {
		return ChangeDeliveryBaseline{}, false, ErrChangeDeliveryBaselineUnavailable
	}
	if err := ctx.Err(); err != nil {
		return ChangeDeliveryBaseline{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	dir := s.destinationDir(key)
	bytes, err := os.ReadFile(filepath.Join(dir, changeDeliveryBaselineCurrentFile))
	if errors.Is(err, os.ErrNotExist) {
		return ChangeDeliveryBaseline{}, false, nil
	}
	if err != nil || int64(len(bytes)) > changeDeliveryBaselineManifestMax {
		return ChangeDeliveryBaseline{}, false, ErrChangeDeliveryBaselineUnavailable
	}
	var record changeDeliveryBaselineRecord
	if err := json.Unmarshal(bytes, &record); err != nil || record.validate() != nil {
		return ChangeDeliveryBaseline{}, false, ErrChangeDeliveryBaselineUnavailable
	}
	baseline := ChangeDeliveryBaseline{dir: dir, record: record, limit: s.limit}
	if err := baseline.ReadCandidateBaseline(ctx, func(warehouse.Row) error { return nil }); err != nil {
		return ChangeDeliveryBaseline{}, false, err
	}
	return baseline, true, nil
}

func (s *FileChangeDeliveryBaselineStore) destinationDir(key ManagedTargetDeliveryLedgerKey) string {
	owner := key.Owner().Identity()
	target := key.TargetDatabase()
	digest := changeDeliveryHash(changeDeliveryBaselineDomain,
		owner.WorkspaceID,
		owner.ConnectorID,
		owner.ConnectionID,
		target.Kind(),
		target.Value(),
		key.StreamID(),
		key.Namespace(),
		key.Relation(),
	)
	return filepath.Join(s.root, digest)
}

func (s *FileChangeDeliveryBaselineStore) ensureCandidate(ctx context.Context, workset ChangeDeliveryWorkset, candidatePath string) error {
	if existing, err := hashChangeDeliveryFile(ctx, candidatePath, s.limit); err == nil {
		if existing == workset.manifest.CandidateBaselineSHA256 {
			return nil
		}
		return ErrChangeDeliveryBaselineUnavailable
	} else if !errors.Is(err, ErrChangeDeliveryWorksetUnavailable) {
		return ErrChangeDeliveryBaselineUnavailable
	}
	temporary, err := os.CreateTemp(filepath.Dir(candidatePath), ".candidate-")
	if err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	temporaryPath := temporary.Name()
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return ErrChangeDeliveryBaselineUnavailable
	}
	defer func() { _ = os.Remove(temporaryPath) }()
	// copyChangeDeliveryFile owns O_EXCL creation, so reserve a unique name
	// then remove the empty placeholder before handing it to that bounded copy.
	if err := os.Remove(temporaryPath); err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	hash, err := copyChangeDeliveryFile(ctx, filepath.Join(workset.dir, changeDeliveryBaselineFile), temporaryPath, s.limit)
	if err != nil || hash != workset.manifest.CandidateBaselineSHA256 {
		return ErrChangeDeliveryBaselineUnavailable
	}
	file, err := os.OpenFile(temporaryPath, os.O_WRONLY, 0)
	if err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return ErrChangeDeliveryBaselineUnavailable
	}
	if err := file.Close(); err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	if err := os.Rename(temporaryPath, candidatePath); err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	return nil
}

func writeChangeDeliveryBaselineRecord(dir string, record changeDeliveryBaselineRecord) error {
	bytes, err := json.Marshal(record)
	if err != nil || int64(len(bytes)) > changeDeliveryBaselineManifestMax {
		return ErrChangeDeliveryBaselineUnavailable
	}
	temporary, err := os.CreateTemp(dir, ".current-")
	if err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return ErrChangeDeliveryBaselineUnavailable
	}
	if _, err := temporary.Write(bytes); err != nil {
		_ = temporary.Close()
		return ErrChangeDeliveryBaselineUnavailable
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return ErrChangeDeliveryBaselineUnavailable
	}
	if err := temporary.Close(); err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	if err := os.Rename(temporaryPath, filepath.Join(dir, changeDeliveryBaselineCurrentFile)); err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	directory, err := os.Open(dir)
	if err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	defer func() { _ = directory.Close() }()
	if err := directory.Sync(); err != nil {
		return ErrChangeDeliveryBaselineUnavailable
	}
	return nil
}

var _ ChangeDeliveryBaselineStore = (*FileChangeDeliveryBaselineStore)(nil)
