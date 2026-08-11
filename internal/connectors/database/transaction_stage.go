package database

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"polymetrics.ai/internal/durability"
	"polymetrics.ai/internal/synccontract"
)

const (
	transactionStageFormatVersion = 1
	transactionStageBufferSize    = 32 * 1024
	transactionStageKeyBytes      = sha256.Size * 2
	transactionStageMaximumInt64  = int64(^uint64(0) >> 1)
)

var (
	// ErrTransactionReceiptUnavailable reports that no durable
	// whole-transaction receipt exists for the requested opaque transaction
	// identity.
	ErrTransactionReceiptUnavailable = errors.New("durable transaction receipt is unavailable")

	// ErrTransactionStageNotFound reports an absent private transaction stage.
	ErrTransactionStageNotFound = errors.New("transaction stage is unavailable")

	// ErrTransactionStageInProgress prevents interleaved transaction
	// boundaries from making source order ambiguous.
	ErrTransactionStageInProgress = errors.New("transaction stage operation is already in progress")

	// ErrTransactionStageRecoveryRequired reports sealed work that needs an
	// explicit recovery admission before it can be delivered.
	ErrTransactionStageRecoveryRequired = errors.New("transaction stage recovery admission is required")

	// ErrTransactionStageAlreadyCommitted prevents a completed receipt from
	// being replaced or an opaque transaction identity from being re-used.
	ErrTransactionStageAlreadyCommitted = errors.New("transaction stage already has a durable receipt")

	// ErrTransactionStageLimitExceeded is the sentinel parent for a named
	// finite resource refusal.
	ErrTransactionStageLimitExceeded = errors.New("transaction stage resource limit exceeded")
)

// TransactionStageLimit identifies the finite resource whose boundary was
// crossed. The names are stable because callers need to distinguish a replay
// from an operator-capacity refusal without inspecting error text.
type TransactionStageLimit string

const (
	TransactionStageLimitTransactionBytes   TransactionStageLimit = "transaction_bytes"
	TransactionStageLimitTransactionRecords TransactionStageLimit = "transaction_records"
	TransactionStageLimitTransactionAge     TransactionStageLimit = "transaction_age"
	TransactionStageLimitStagedBytes        TransactionStageLimit = "staged_bytes"
)

// TransactionStageLimitExceeded is a typed, fail-closed quota refusal. For
// TransactionStageLimitTransactionAge, Maximum and Observed are nanoseconds.
type TransactionStageLimitExceeded struct {
	Limit    TransactionStageLimit
	Maximum  int64
	Observed int64
}

func (e *TransactionStageLimitExceeded) Error() string {
	if e == nil {
		return ErrTransactionStageLimitExceeded.Error()
	}
	return fmt.Sprintf("%s: %s observed %d exceeds maximum %d", ErrTransactionStageLimitExceeded, e.Limit, e.Observed, e.Maximum)
}

func (e *TransactionStageLimitExceeded) Unwrap() error { return ErrTransactionStageLimitExceeded }

// TransactionStageLimits bounds both one private transaction and the retained
// root. All limits are required so an unconfigured stage cannot become an
// unbounded spool.
type TransactionStageLimits struct {
	MaxTransactionBytes   int64
	MaxTransactionRecords int64
	MaxTransactionAge     time.Duration
	MaxStagedBytes        int64
}

func (l TransactionStageLimits) validate() error {
	if l.MaxTransactionBytes <= 0 || l.MaxTransactionRecords <= 0 || l.MaxTransactionAge <= 0 || l.MaxStagedBytes <= 0 {
		return errors.New("transaction stage limits must be positive and finite")
	}
	if l.MaxStagedBytes < l.MaxTransactionBytes {
		return errors.New("transaction stage root byte limit must cover one transaction")
	}
	return nil
}

// TransactionStageOptions configures one source-agnostic private stage. Root
// is controlled by the caller; opaque transaction IDs never become paths below
// that root.
type TransactionStageOptions struct {
	Root   string
	Limits TransactionStageLimits
}

// TransactionChunk is one ordered streamed chunk of a committed transaction.
// Reader is valid only during the StreamChunks callback that supplied it.
type TransactionChunk struct {
	Sequence uint64
	Records  int64
	Bytes    int64
	Reader   io.Reader
}

// CommittedTransaction is the whole transaction delivered only after a stage
// has crossed its private commit boundary. TransactionKey is a deterministic
// digest of the opaque provider identity, never the provider value itself.
type CommittedTransaction struct {
	TransactionKey string
	Bytes          int64
	Records        int64
	ContentDigest  string
	stream         func(context.Context, func(TransactionChunk) error) error
	delivery       *transactionStageDelivery
}

// StreamChunks visits complete chunks in append order. A receiver must consume
// each chunk reader completely before its callback returns; otherwise it has
// not processed the whole committed transaction and cannot obtain a receipt.
func (t CommittedTransaction) StreamChunks(ctx context.Context, visit func(TransactionChunk) error) error {
	if ctx == nil {
		return errors.New("transaction stream context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if visit == nil {
		return errors.New("transaction chunk visitor is required")
	}
	if t.stream == nil {
		return errors.New("committed transaction stream is unavailable")
	}
	if t.delivery == nil || !t.delivery.start() {
		return errors.New("committed transaction stream is unavailable")
	}
	err := t.stream(ctx, visit)
	t.delivery.finish(err == nil)
	return err
}

// DownstreamTransactionReceipt is supplied by a receiver only after its
// complete transaction write is durable according to the receiver's protocol.
// The stage persists a richer immutable receipt around this downstream fact.
type DownstreamTransactionReceipt struct {
	ReceiptID string
	Sink      string
	DurableAt time.Time
}

func (r DownstreamTransactionReceipt) validate() error {
	if strings.TrimSpace(r.ReceiptID) == "" || strings.TrimSpace(r.Sink) == "" || r.DurableAt.IsZero() {
		return errors.New("durable downstream transaction receipt is incomplete")
	}
	if len(r.ReceiptID) > 1024 || len(r.Sink) > 1024 {
		return errors.New("durable downstream transaction receipt is invalid")
	}
	return nil
}

// DurableTransactionReceiver receives one complete committed transaction and
// returns durable downstream evidence only after it has persisted the whole
// transaction. It is deliberately not a generic destination writer.
type DurableTransactionReceiver interface {
	ReceiveCommittedTransaction(context.Context, CommittedTransaction) (DownstreamTransactionReceipt, error)
}

// TransactionReceipt is immutable receipt evidence for one complete staged
// transaction.
type TransactionReceipt struct {
	payload transactionReceiptPayload
}

type transactionReceiptPayload struct {
	transactionKey      string
	downstreamReceiptID string
	sink                string
	durableAt           time.Time
	bytes               int64
	records             int64
	contentDigest       string
	durable             bool
}

func newTransactionReceipt(transactionKey, downstreamReceiptID, sink string, durableAt time.Time, bytes, records int64, contentDigest string) TransactionReceipt {
	return TransactionReceipt{payload: transactionReceiptPayload{
		transactionKey:      transactionKey,
		downstreamReceiptID: downstreamReceiptID,
		sink:                sink,
		durableAt:           durableAt.UTC(),
		bytes:               bytes,
		records:             records,
		contentDigest:       contentDigest,
		durable:             true,
	}}
}

// TransactionKey returns the opaque-safe transaction identity.
func (r TransactionReceipt) TransactionKey() string { return r.payload.transactionKey }

// DownstreamReceiptID returns the receiver's durable receipt identifier.
func (r TransactionReceipt) DownstreamReceiptID() string { return r.payload.downstreamReceiptID }

// Sink returns the receiver that durably stored the transaction.
func (r TransactionReceipt) Sink() string { return r.payload.sink }

// DurableAt returns when the receiver made the transaction durable.
func (r TransactionReceipt) DurableAt() time.Time { return r.payload.durableAt }

// Bytes returns the complete transaction byte count.
func (r TransactionReceipt) Bytes() int64 { return r.payload.bytes }

// Records returns the complete transaction record count.
func (r TransactionReceipt) Records() int64 { return r.payload.records }

// ContentDigest returns the complete transaction content digest.
func (r TransactionReceipt) ContentDigest() string { return r.payload.contentDigest }

// Acknowledgement adapts a persisted durable receipt to the existing sync
// checkpoint contract. A local stage, receiver call, or manually built value
// cannot produce this acknowledgement.
func (r TransactionReceipt) Acknowledgement() (synccontract.DownstreamAcknowledgement, error) {
	if !r.payload.durable {
		return synccontract.DownstreamAcknowledgement{}, ErrTransactionReceiptUnavailable
	}
	return synccontract.NewDurableDownstreamAcknowledgement(r.payload.sink, r.payload.durableAt)
}

// PendingTransaction is sealed receipt-less work found in a stage root. It is
// intentionally keyed by the opaque-safe digest and remains acknowledgement
// ineligible until CommitTransaction stores a valid durable receipt.
type PendingTransaction struct {
	TransactionKey string
	Bytes          int64
	Records        int64
	ContentDigest  string
	CreatedAt      time.Time
}

type transactionStageStatus uint8

const (
	stageStatusCreating transactionStageStatus = iota
	stageStatusActive
	stageStatusAppending
	stageStatusSealing
	stageStatusSealed
	stageStatusRecoveryHeld
	stageStatusCommitting
	stageStatusDiscarding
	stageStatusDiscardFailed
	stageStatusReceiptPersisted
)

type transactionStageEntry struct {
	manifest transactionStageManifest
	status   transactionStageStatus
}

type transactionStageDelivery struct {
	mu        sync.Mutex
	started   bool
	completed bool
}

func (d *transactionStageDelivery) start() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.started {
		return false
	}
	d.started = true
	return true
}

func (d *transactionStageDelivery) finish(completed bool) {
	d.mu.Lock()
	d.completed = completed
	d.mu.Unlock()
}

func (d *transactionStageDelivery) complete() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.started && d.completed
}

type transactionStageManifest struct {
	Version        int                     `json:"version"`
	TransactionKey string                  `json:"transaction_key"`
	InstanceID     string                  `json:"instance_id"`
	State          string                  `json:"state"`
	CreatedAt      time.Time               `json:"created_at"`
	Chunks         []transactionStageChunk `json:"chunks"`
	Bytes          int64                   `json:"bytes"`
	Records        int64                   `json:"records"`
	ContentDigest  string                  `json:"content_digest"`
}

type transactionStageChunk struct {
	Sequence      uint64 `json:"sequence"`
	File          string `json:"file"`
	Bytes         int64  `json:"bytes"`
	Records       int64  `json:"records"`
	ContentDigest string `json:"content_digest"`
}

type storedTransactionReceipt struct {
	Version             int       `json:"version"`
	TransactionKey      string    `json:"transaction_key"`
	DownstreamReceiptID string    `json:"downstream_receipt_id"`
	Sink                string    `json:"sink"`
	DurableAt           time.Time `json:"durable_at"`
	Bytes               int64     `json:"bytes"`
	Records             int64     `json:"records"`
	ContentDigest       string    `json:"content_digest"`
}

type transactionStageDiscardIntent struct {
	Version        int    `json:"version"`
	TransactionKey string `json:"transaction_key"`
	InstanceID     string `json:"instance_id"`
	State          string `json:"state"`
}

type transactionStageDurabilityOutcome uint8

const (
	transactionStageDurabilityNotApplied transactionStageDurabilityOutcome = iota
	transactionStageDurabilityDurable
	transactionStageDurabilityIndeterminate
)

func (o transactionStageDurabilityOutcome) String() string {
	switch o {
	case transactionStageDurabilityNotApplied:
		return "not-applied"
	case transactionStageDurabilityDurable:
		return "durable"
	case transactionStageDurabilityIndeterminate:
		return "indeterminate"
	default:
		return "unknown"
	}
}

// transactionStageFile is narrow enough to fault-inject write and fsync
// boundaries in-package without giving production callers filesystem knobs.
type transactionStageFile interface {
	io.Reader
	io.Writer
	io.Closer
	Sync() error
	Name() string
}

type transactionStageStorage struct {
	mkdirAll      func(string, fs.FileMode) error
	createTemp    func(string, string) (transactionStageFile, error)
	open          func(string) (transactionStageFile, error)
	readFile      func(string) ([]byte, error)
	readDir       func(string) ([]os.DirEntry, error)
	rename        func(string, string) error
	remove        func(string) error
	removeAll     func(string) error
	syncDirectory func(string) error
}

func defaultTransactionStageStorage() transactionStageStorage {
	return transactionStageStorage{
		mkdirAll: os.MkdirAll,
		createTemp: func(directory, pattern string) (transactionStageFile, error) {
			return os.CreateTemp(directory, pattern)
		},
		open: func(path string) (transactionStageFile, error) {
			return os.Open(path)
		},
		readFile:      os.ReadFile,
		readDir:       os.ReadDir,
		rename:        os.Rename,
		remove:        os.Remove,
		removeAll:     os.RemoveAll,
		syncDirectory: durability.SyncDirectory,
	}
}

// CommittedTransactionStage owns private in-progress chunks and durable
// whole-transaction receipts. It has no database, source checkpoint, or
// destination-DML authority.
type CommittedTransactionStage struct {
	root    string
	limits  TransactionStageLimits
	now     func() time.Time
	storage transactionStageStorage

	mu            sync.Mutex
	entries       map[string]*transactionStageEntry
	receipts      map[string]TransactionReceipt
	stagedBytes   int64
	inFlightBytes int64
}

// OpenCommittedTransactionStage opens a private stage and removes only
// incomplete/orphan data from a prior process. Bare sealed receipt-less work
// requires explicit recovery admission before delivery.
func OpenCommittedTransactionStage(options TransactionStageOptions) (*CommittedTransactionStage, error) {
	return openCommittedTransactionStage(options, defaultTransactionStageStorage(), func() time.Time {
		return time.Now().UTC()
	})
}

func openCommittedTransactionStage(options TransactionStageOptions, storage transactionStageStorage, now func() time.Time) (*CommittedTransactionStage, error) {
	if strings.TrimSpace(options.Root) == "" {
		return nil, errors.New("transaction stage root is required")
	}
	if err := options.Limits.validate(); err != nil {
		return nil, err
	}
	if now == nil {
		return nil, errors.New("transaction stage clock is required")
	}
	root, err := filepath.Abs(options.Root)
	if err != nil {
		return nil, fmt.Errorf("resolve transaction stage root: %w", err)
	}
	stage := &CommittedTransactionStage{
		root:     filepath.Clean(root),
		limits:   options.Limits,
		now:      now,
		storage:  storage,
		entries:  make(map[string]*transactionStageEntry),
		receipts: make(map[string]TransactionReceipt),
	}
	if err := stage.recover(); err != nil {
		return nil, err
	}
	return stage, nil
}

// BeginTransaction starts one private staged transaction. It does not call a
// receiver or construct acknowledgement evidence.
func (s *CommittedTransactionStage) BeginTransaction(ctx context.Context, transactionID string) error {
	if err := requireTransactionStageContext(ctx); err != nil {
		return err
	}
	key, err := transactionStageKey(transactionID)
	if err != nil {
		return err
	}
	createdAt, err := s.currentTime()
	if err != nil {
		return err
	}
	instanceID, err := newTransactionStageInstanceID()
	if err != nil {
		return err
	}
	manifest := transactionStageManifest{
		Version:        transactionStageFormatVersion,
		TransactionKey: key,
		InstanceID:     instanceID,
		State:          transactionStageStateActive,
		CreatedAt:      createdAt,
		ContentDigest:  transactionDigestSeed(),
	}

	s.mu.Lock()
	if _, exists := s.receipts[key]; exists {
		s.mu.Unlock()
		return ErrTransactionStageAlreadyCommitted
	}
	if _, exists := s.entries[key]; exists {
		s.mu.Unlock()
		return ErrTransactionStageInProgress
	}
	entry := &transactionStageEntry{manifest: manifest, status: stageStatusCreating}
	s.entries[key] = entry
	s.mu.Unlock()

	if err := s.createStageDirectory(key); err != nil {
		return s.failBegin(key, entry, err)
	}
	if err := s.writeManifest(ctx, manifest); err != nil {
		return s.failBegin(key, entry, err)
	}

	s.mu.Lock()
	if s.entries[key] == entry {
		entry.status = stageStatusActive
	}
	s.mu.Unlock()
	return nil
}

// AppendChunk streams one opaque source chunk into private storage. A stream
// failure, cancellation, or named quota refusal discards the incomplete
// transaction rather than retaining a partial final chunk.
func (s *CommittedTransactionStage) AppendChunk(ctx context.Context, transactionID string, records int64, reader io.Reader) error {
	if err := requireTransactionStageContext(ctx); err != nil {
		return err
	}
	if reader == nil {
		return errors.New("transaction chunk reader is required")
	}
	if records <= 0 {
		return errors.New("transaction chunk record count must be positive")
	}
	key, err := transactionStageKey(transactionID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	entry, exists := s.entries[key]
	if !exists {
		s.mu.Unlock()
		return ErrTransactionStageNotFound
	}
	if entry.status != stageStatusActive {
		s.mu.Unlock()
		return ErrTransactionStageInProgress
	}
	if err := s.ageLimitExceeded(entry.manifest); err != nil {
		entry.status = stageStatusAppending
		s.mu.Unlock()
		return s.failAppend(ctx, key, entry, 0, err)
	}
	observedRecords, withinRecordLimit := transactionStageLimitedAdd(entry.manifest.Records, records, s.limits.MaxTransactionRecords)
	if !withinRecordLimit {
		entry.status = stageStatusAppending
		s.mu.Unlock()
		return s.failAppend(ctx, key, entry, 0, &TransactionStageLimitExceeded{
			Limit:    TransactionStageLimitTransactionRecords,
			Maximum:  s.limits.MaxTransactionRecords,
			Observed: transactionStageSaturatingAdd(entry.manifest.Records, records),
		})
	}
	entry.status = stageStatusAppending
	manifest := cloneTransactionStageManifest(entry.manifest)
	s.mu.Unlock()

	chunk, reserved, err := s.writeChunk(ctx, key, manifest, records, reader)
	if err != nil {
		return s.failAppend(ctx, key, entry, reserved, err)
	}
	next := cloneTransactionStageManifest(manifest)
	next.Chunks = append(next.Chunks, chunk)
	next.Bytes, _ = transactionStageLimitedAdd(manifest.Bytes, chunk.Bytes, s.limits.MaxTransactionBytes)
	next.Records = observedRecords
	next.ContentDigest = nextTransactionDigest(next.ContentDigest, chunk)
	if err := s.writeManifest(ctx, next); err != nil {
		return s.failAppend(ctx, key, entry, reserved, err)
	}

	s.mu.Lock()
	if s.entries[key] != entry || entry.status != stageStatusAppending {
		s.mu.Unlock()
		return s.failAppend(ctx, key, entry, reserved, ErrTransactionStageInProgress)
	}
	s.releaseReservationLocked(reserved)
	s.stagedBytes += chunk.Bytes
	entry.manifest = next
	entry.status = stageStatusActive
	s.mu.Unlock()
	return nil
}

// CommitTransaction seals private chunks, verifies their complete ordered
// representation before receiver delivery, and returns eligibility only after
// its immutable receipt has been durably persisted.
func (s *CommittedTransactionStage) CommitTransaction(ctx context.Context, transactionID string, receiver DurableTransactionReceiver) (TransactionReceipt, error) {
	if ctx == nil {
		return TransactionReceipt{}, errors.New("transaction stage context is required")
	}
	key, err := transactionStageKey(transactionID)
	if err != nil {
		return TransactionReceipt{}, err
	}

	s.mu.Lock()
	if receipt, exists := s.receipts[key]; exists {
		s.mu.Unlock()
		return receipt, nil
	}
	entry, exists := s.entries[key]
	if !exists {
		s.mu.Unlock()
		return TransactionReceipt{}, ErrTransactionStageNotFound
	}
	if entry.status == stageStatusRecoveryHeld {
		s.mu.Unlock()
		return TransactionReceipt{}, ErrTransactionStageRecoveryRequired
	}
	if entry.status != stageStatusActive && entry.status != stageStatusSealed {
		s.mu.Unlock()
		return TransactionReceipt{}, ErrTransactionStageInProgress
	}
	if err := ctx.Err(); err != nil {
		entry.status = stageStatusDiscarding
		s.mu.Unlock()
		return TransactionReceipt{}, s.discardEntry(ctx, key, entry, err)
	}
	if err := s.ageLimitExceeded(entry.manifest); err != nil {
		entry.status = stageStatusDiscarding
		s.mu.Unlock()
		return TransactionReceipt{}, s.discardEntry(ctx, key, entry, err)
	}
	if receiver == nil {
		s.mu.Unlock()
		return TransactionReceipt{}, errors.New("durable transaction receiver is required")
	}

	manifest := cloneTransactionStageManifest(entry.manifest)
	needsSeal := manifest.State == transactionStageStateActive
	if needsSeal {
		entry.status = stageStatusSealing
	} else {
		entry.status = stageStatusCommitting
	}
	s.mu.Unlock()

	if needsSeal {
		manifest.State = transactionStageStateSealed
		if err := s.writeManifest(ctx, manifest); err != nil {
			return TransactionReceipt{}, s.discardEntry(ctx, key, entry, err)
		}
		s.mu.Lock()
		if s.entries[key] != entry || entry.status != stageStatusSealing {
			s.mu.Unlock()
			return TransactionReceipt{}, s.discardEntry(ctx, key, entry, ErrTransactionStageInProgress)
		}
		entry.manifest = cloneTransactionStageManifest(manifest)
		entry.status = stageStatusCommitting
		s.mu.Unlock()
	}

	if err := s.verifyManifestFiles(ctx, manifest); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return TransactionReceipt{}, s.discardEntry(ctx, key, entry, err)
		}
		s.finishCommitAsSealed(key, entry)
		return TransactionReceipt{}, fmt.Errorf("verify sealed transaction stage: %w", err)
	}

	delivery := &transactionStageDelivery{}
	downstreamReceipt, err := receiver.ReceiveCommittedTransaction(ctx, s.committedTransaction(manifest, delivery))
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return TransactionReceipt{}, s.discardEntry(ctx, key, entry, ctxErr)
		}
		s.finishCommitAsSealed(key, entry)
		return TransactionReceipt{}, fmt.Errorf("receive committed transaction: %w", err)
	}
	if !delivery.complete() {
		s.finishCommitAsSealed(key, entry)
		return TransactionReceipt{}, errors.New("durable transaction receiver did not consume the complete transaction")
	}
	if err := ctx.Err(); err != nil {
		return TransactionReceipt{}, s.discardEntry(ctx, key, entry, err)
	}
	if err := downstreamReceipt.validate(); err != nil {
		s.finishCommitAsSealed(key, entry)
		return TransactionReceipt{}, err
	}

	receipt := newTransactionReceipt(
		manifest.TransactionKey,
		downstreamReceipt.ReceiptID,
		downstreamReceipt.Sink,
		downstreamReceipt.DurableAt,
		manifest.Bytes,
		manifest.Records,
		manifest.ContentDigest,
	)
	if err := s.persistReceipt(ctx, receipt); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return TransactionReceipt{}, s.discardEntry(ctx, key, entry, ctxErr)
		}
		s.finishCommitAsSealed(key, entry)
		return TransactionReceipt{}, err
	}

	s.mu.Lock()
	s.receipts[key] = receipt
	if s.entries[key] == entry {
		entry.status = stageStatusReceiptPersisted
	}
	s.mu.Unlock()

	// Receipt durability is the acknowledgement boundary. A cleanup failure is
	// deliberately recoverable and must not make that already-durable receipt
	// look ineligible; restart recovery removes the sealed residue.
	if err := s.removeStageDirectory(key); err == nil {
		s.mu.Lock()
		if s.entries[key] == entry {
			s.subtractStagedBytesLocked(entry.manifest.Bytes)
			delete(s.entries, key)
		}
		s.mu.Unlock()
	}
	return receipt, nil
}

// AbortTransaction removes a receipt-less transaction. It deliberately
// completes cleanup even when the caller's context is cancelled.
func (s *CommittedTransactionStage) AbortTransaction(ctx context.Context, transactionID string) error {
	key, err := transactionStageKey(transactionID)
	if err != nil {
		return err
	}
	s.mu.Lock()
	if _, exists := s.receipts[key]; exists {
		s.mu.Unlock()
		return ErrTransactionStageAlreadyCommitted
	}
	entry, exists := s.entries[key]
	if !exists {
		s.mu.Unlock()
		return ErrTransactionStageNotFound
	}
	if entry.status != stageStatusActive && entry.status != stageStatusSealed && entry.status != stageStatusRecoveryHeld && entry.status != stageStatusDiscardFailed {
		s.mu.Unlock()
		return ErrTransactionStageInProgress
	}
	entry.status = stageStatusDiscarding
	s.mu.Unlock()
	return s.discardEntry(ctx, key, entry, nil)
}

// AdmitRecoveredTransaction permits a verified sealed stage recovered from a
// prior process to proceed through CommitTransaction.
func (s *CommittedTransactionStage) AdmitRecoveredTransaction(ctx context.Context, transactionID string) error {
	if err := requireTransactionStageContext(ctx); err != nil {
		return err
	}
	key, err := transactionStageKey(transactionID)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.receipts[key]; exists {
		return ErrTransactionStageAlreadyCommitted
	}
	entry, exists := s.entries[key]
	if !exists {
		return ErrTransactionStageNotFound
	}
	if entry.status == stageStatusSealed {
		return nil
	}
	if entry.status != stageStatusRecoveryHeld {
		return ErrTransactionStageInProgress
	}
	entry.status = stageStatusSealed
	return nil
}

// Receipt returns only a receipt that passed the durable receipt transition.
// The transaction ID is transformed to the same private stage key used at
// begin time, so it never becomes a filename or stored payload.
func (s *CommittedTransactionStage) Receipt(transactionID string) (TransactionReceipt, error) {
	key, err := transactionStageKey(transactionID)
	if err != nil {
		return TransactionReceipt{}, err
	}
	s.mu.Lock()
	receipt, exists := s.receipts[key]
	s.mu.Unlock()
	if !exists {
		return TransactionReceipt{}, ErrTransactionReceiptUnavailable
	}
	return receipt, nil
}

// PendingTransactions reports sealed receipt-less transactions discovered in
// this process or at startup. It never reports active private chunks.
func (s *CommittedTransactionStage) PendingTransactions() []PendingTransaction {
	s.mu.Lock()
	pending := make([]PendingTransaction, 0)
	for _, entry := range s.entries {
		if (entry.status != stageStatusSealed && entry.status != stageStatusRecoveryHeld) || entry.manifest.State != transactionStageStateSealed {
			continue
		}
		pending = append(pending, PendingTransaction{
			TransactionKey: entry.manifest.TransactionKey,
			Bytes:          entry.manifest.Bytes,
			Records:        entry.manifest.Records,
			ContentDigest:  entry.manifest.ContentDigest,
			CreatedAt:      entry.manifest.CreatedAt,
		})
	}
	s.mu.Unlock()
	sort.Slice(pending, func(left, right int) bool {
		return pending[left].TransactionKey < pending[right].TransactionKey
	})
	return pending
}

func (s *CommittedTransactionStage) recover() error {
	if err := s.ensureLayout(); err != nil {
		return err
	}
	if err := s.recoverReceipts(); err != nil {
		return err
	}
	entries, err := s.storage.readDir(s.transactionsDirectory())
	if err != nil {
		return fmt.Errorf("read transaction stage directory: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(s.transactionsDirectory(), name)
		if !entry.IsDir() || !validTransactionStageKey(name) {
			if err := s.removeOrphan(path, s.transactionsDirectory()); err != nil {
				return err
			}
			continue
		}
		if _, hasReceipt := s.receipts[name]; hasReceipt {
			if err := s.removeStageDirectory(name); err != nil {
				return fmt.Errorf("clean receipt-backed transaction stage: %w", err)
			}
			continue
		}
		manifest, err := s.readManifest(name)
		if err == nil && manifest.State == transactionStageStateDiscarded {
			if removeErr := s.removeStageDirectory(name); removeErr != nil {
				return fmt.Errorf("clean discarded transaction stage: %w", removeErr)
			}
			continue
		}
		if err != nil {
			if removeErr := s.removeStageDirectory(name); removeErr != nil {
				return fmt.Errorf("clean incomplete transaction stage: %w", removeErr)
			}
			continue
		}
		discarded, err := s.readDiscardIntent(manifest)
		if err != nil {
			return fmt.Errorf("read transaction discard intent: %w", err)
		}
		if discarded {
			if removeErr := s.removeStageDirectory(name); removeErr != nil {
				return fmt.Errorf("clean discarded transaction stage: %w", removeErr)
			}
			continue
		}
		if manifest.State != transactionStageStateSealed || s.verifyManifestFiles(context.Background(), manifest) != nil {
			if removeErr := s.removeStageDirectory(name); removeErr != nil {
				return fmt.Errorf("clean incomplete transaction stage: %w", removeErr)
			}
			continue
		}
		s.entries[name] = &transactionStageEntry{manifest: manifest, status: stageStatusRecoveryHeld}
		if stagedBytes, withinLimit := transactionStageLimitedAdd(s.stagedBytes, manifest.Bytes, s.limits.MaxStagedBytes); withinLimit {
			s.stagedBytes = stagedBytes
		} else {
			// A lower limit on restart must never let historical sealed data
			// bypass the new bound. Saturating at the configured limit blocks
			// new appends until the retained work is committed or aborted.
			s.stagedBytes = s.limits.MaxStagedBytes
		}
	}
	return nil
}

func (s *CommittedTransactionStage) recoverReceipts() error {
	entries, err := s.storage.readDir(s.receiptsDirectory())
	if err != nil {
		return fmt.Errorf("read transaction receipt directory: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		key := strings.TrimSuffix(name, ".json")
		path := filepath.Join(s.receiptsDirectory(), name)
		if entry.IsDir() || !strings.HasSuffix(name, ".json") || !validTransactionStageKey(key) {
			if err := s.removeOrphan(path, s.receiptsDirectory()); err != nil {
				return err
			}
			continue
		}
		receipt, err := s.readReceipt(key)
		if err != nil {
			return fmt.Errorf("read durable transaction receipt: %w", err)
		}
		s.receipts[key] = receipt
	}
	return nil
}

func (s *CommittedTransactionStage) ensureLayout() error {
	for _, directory := range []string{s.root, s.transactionsDirectory(), s.receiptsDirectory(), s.discardsDirectory()} {
		if err := s.storage.mkdirAll(directory, 0o700); err != nil {
			return fmt.Errorf("create transaction stage directory: %w", err)
		}
		if err := s.storage.syncDirectory(directory); err != nil {
			return fmt.Errorf("sync transaction stage directory: %w", err)
		}
		if directory != s.root {
			if err := s.storage.syncDirectory(s.rootDirectory()); err != nil {
				return fmt.Errorf("sync transaction stage root: %w", err)
			}
		}
	}
	return nil
}

func (s *CommittedTransactionStage) createStageDirectory(key string) error {
	if err := s.storage.mkdirAll(s.chunksDirectory(key), 0o700); err != nil {
		return fmt.Errorf("create private transaction stage: %w", err)
	}
	if err := s.storage.syncDirectory(s.transactionDirectory(key)); err != nil {
		return fmt.Errorf("sync private transaction stage: %w", err)
	}
	if err := s.storage.syncDirectory(s.transactionsDirectory()); err != nil {
		return fmt.Errorf("sync transaction stage parent: %w", err)
	}
	return nil
}

func (s *CommittedTransactionStage) failBegin(key string, entry *transactionStageEntry, cause error) error {
	cleanupErr := s.removeStageDirectory(key)
	s.mu.Lock()
	if s.entries[key] == entry {
		delete(s.entries, key)
	}
	s.mu.Unlock()
	return withTransactionStageCleanupError(cause, cleanupErr)
}

func (s *CommittedTransactionStage) writeChunk(ctx context.Context, key string, manifest transactionStageManifest, records int64, reader io.Reader) (chunk transactionStageChunk, reserved int64, err error) {
	if err := s.storage.mkdirAll(s.chunksDirectory(key), 0o700); err != nil {
		return transactionStageChunk{}, 0, fmt.Errorf("create transaction chunk directory: %w", err)
	}
	temporary, err := s.storage.createTemp(s.chunksDirectory(key), ".chunk.tmp-*")
	if err != nil {
		return transactionStageChunk{}, 0, fmt.Errorf("create temporary transaction chunk: %w", err)
	}
	temporaryPath := temporary.Name()
	closed := false
	finalized := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		if err != nil {
			_ = s.storage.remove(temporaryPath)
			if finalized {
				_ = s.storage.remove(s.chunkPath(key, uint64(len(manifest.Chunks))))
				_ = s.storage.syncDirectory(s.chunksDirectory(key))
			}
		}
	}()

	digest := sha256.New()
	buffer := make([]byte, transactionStageBufferSize)
	var copied int64
	for {
		if err := requireTransactionStageContext(ctx); err != nil {
			return transactionStageChunk{}, reserved, err
		}
		if err := s.ageLimitExceeded(manifest); err != nil {
			return transactionStageChunk{}, reserved, err
		}
		read, readErr := reader.Read(buffer)
		if read > 0 {
			totalBeforeChunk, withinLimit := transactionStageLimitedAdd(manifest.Bytes, copied, s.limits.MaxTransactionBytes)
			if !withinLimit {
				return transactionStageChunk{}, reserved, &TransactionStageLimitExceeded{
					Limit:    TransactionStageLimitTransactionBytes,
					Maximum:  s.limits.MaxTransactionBytes,
					Observed: transactionStageSaturatingAdd(manifest.Bytes, copied),
				}
			}
			nextTotal, withinLimit := transactionStageLimitedAdd(totalBeforeChunk, int64(read), s.limits.MaxTransactionBytes)
			if !withinLimit {
				return transactionStageChunk{}, reserved, &TransactionStageLimitExceeded{
					Limit:    TransactionStageLimitTransactionBytes,
					Maximum:  s.limits.MaxTransactionBytes,
					Observed: transactionStageSaturatingAdd(totalBeforeChunk, int64(read)),
				}
			}
			if err := s.reserveBytes(int64(read)); err != nil {
				return transactionStageChunk{}, reserved, err
			}
			reserved += int64(read)
			if err := writeTransactionStageBytes(temporary, buffer[:read]); err != nil {
				return transactionStageChunk{}, reserved, fmt.Errorf("write temporary transaction chunk: %w", err)
			}
			if _, err := digest.Write(buffer[:read]); err != nil {
				return transactionStageChunk{}, reserved, fmt.Errorf("digest transaction chunk: %w", err)
			}
			copied = nextTotal - manifest.Bytes
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return transactionStageChunk{}, reserved, fmt.Errorf("read transaction chunk: %w", readErr)
		}
		if read == 0 {
			return transactionStageChunk{}, reserved, io.ErrNoProgress
		}
	}
	if err := requireTransactionStageContext(ctx); err != nil {
		return transactionStageChunk{}, reserved, err
	}
	if err := s.ageLimitExceeded(manifest); err != nil {
		return transactionStageChunk{}, reserved, err
	}
	if err := temporary.Sync(); err != nil {
		return transactionStageChunk{}, reserved, fmt.Errorf("sync temporary transaction chunk: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return transactionStageChunk{}, reserved, fmt.Errorf("close temporary transaction chunk: %w", err)
	}
	closed = true
	sequence := uint64(len(manifest.Chunks))
	finalPath := s.chunkPath(key, sequence)
	if err := s.storage.rename(temporaryPath, finalPath); err != nil {
		return transactionStageChunk{}, reserved, fmt.Errorf("rename durable transaction chunk: %w", err)
	}
	finalized = true
	if err := s.storage.syncDirectory(s.chunksDirectory(key)); err != nil {
		return transactionStageChunk{}, reserved, fmt.Errorf("sync transaction chunk directory: %w", err)
	}
	return transactionStageChunk{
		Sequence:      sequence,
		File:          transactionStageChunkFile(sequence),
		Bytes:         copied,
		Records:       records,
		ContentDigest: hex.EncodeToString(digest.Sum(nil)),
	}, reserved, nil
}

func (s *CommittedTransactionStage) reserveBytes(bytes int64) error {
	if bytes <= 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	used, withinLimit := transactionStageLimitedAdd(s.stagedBytes, s.inFlightBytes, s.limits.MaxStagedBytes)
	if !withinLimit {
		return &TransactionStageLimitExceeded{
			Limit:    TransactionStageLimitStagedBytes,
			Maximum:  s.limits.MaxStagedBytes,
			Observed: transactionStageSaturatingAdd(s.stagedBytes, s.inFlightBytes),
		}
	}
	if _, withinLimit := transactionStageLimitedAdd(used, bytes, s.limits.MaxStagedBytes); !withinLimit {
		return &TransactionStageLimitExceeded{
			Limit:    TransactionStageLimitStagedBytes,
			Maximum:  s.limits.MaxStagedBytes,
			Observed: transactionStageSaturatingAdd(used, bytes),
		}
	}
	s.inFlightBytes += bytes
	return nil
}

func (s *CommittedTransactionStage) failAppend(ctx context.Context, key string, entry *transactionStageEntry, reserved int64, cause error) error {
	s.mu.Lock()
	s.releaseReservationLocked(reserved)
	if s.entries[key] == entry {
		entry.status = stageStatusDiscarding
	}
	s.mu.Unlock()
	return s.discardEntry(ctx, key, entry, cause)
}

func (s *CommittedTransactionStage) discardEntry(ctx context.Context, key string, entry *transactionStageEntry, cause error) error {
	intentOutcome, intentErr := s.persistDiscardIntent(transactionStageCleanupContext(ctx), entry.manifest)
	cleanupOutcome, cleanupErr := s.removeStageDirectoryWithOutcome(key)
	s.mu.Lock()
	if s.entries[key] == entry {
		if cleanupOutcome == transactionStageDurabilityDurable {
			s.subtractStagedBytesLocked(entry.manifest.Bytes)
			delete(s.entries, key)
		} else {
			entry.status = stageStatusDiscardFailed
		}
	}
	s.mu.Unlock()
	return withTransactionStageCleanupError(cause, transactionStageDiscardError(intentOutcome, cleanupOutcome, errors.Join(intentErr, cleanupErr)))
}

func (s *CommittedTransactionStage) finishCommitAsSealed(key string, entry *transactionStageEntry) {
	s.mu.Lock()
	if s.entries[key] == entry {
		entry.status = stageStatusSealed
		entry.manifest.State = transactionStageStateSealed
	}
	s.mu.Unlock()
}

func (s *CommittedTransactionStage) committedTransaction(manifest transactionStageManifest, delivery *transactionStageDelivery) CommittedTransaction {
	return CommittedTransaction{
		TransactionKey: manifest.TransactionKey,
		Bytes:          manifest.Bytes,
		Records:        manifest.Records,
		ContentDigest:  manifest.ContentDigest,
		delivery:       delivery,
		stream: func(ctx context.Context, visitor func(TransactionChunk) error) error {
			return s.streamManifest(ctx, manifest, visitor)
		},
	}
}

func (s *CommittedTransactionStage) streamManifest(ctx context.Context, manifest transactionStageManifest, visitor func(TransactionChunk) error) error {
	for _, chunk := range manifest.Chunks {
		if err := requireTransactionStageContext(ctx); err != nil {
			return err
		}
		file, err := s.storage.open(filepath.Join(s.chunksDirectory(manifest.TransactionKey), chunk.File))
		if err != nil {
			return fmt.Errorf("open staged transaction chunk: %w", err)
		}
		limited := &io.LimitedReader{R: file, N: chunk.Bytes}
		digest := sha256.New()
		reader := &transactionStageContextReader{ctx: ctx, reader: io.TeeReader(limited, digest)}
		visitErr := visitor(TransactionChunk{
			Sequence: chunk.Sequence,
			Records:  chunk.Records,
			Bytes:    chunk.Bytes,
			Reader:   reader,
		})
		closeErr := file.Close()
		if visitErr != nil {
			return visitErr
		}
		if closeErr != nil {
			return fmt.Errorf("close staged transaction chunk: %w", closeErr)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if limited.N != 0 {
			return errors.New("durable transaction receiver did not consume a complete chunk")
		}
		if hex.EncodeToString(digest.Sum(nil)) != chunk.ContentDigest {
			return errors.New("staged transaction chunk changed before receipt")
		}
	}
	return nil
}

func (s *CommittedTransactionStage) verifyManifestFiles(ctx context.Context, manifest transactionStageManifest) error {
	for _, chunk := range manifest.Chunks {
		if err := requireTransactionStageContext(ctx); err != nil {
			return err
		}
		file, err := s.storage.open(filepath.Join(s.chunksDirectory(manifest.TransactionKey), chunk.File))
		if err != nil {
			return err
		}
		digest := sha256.New()
		buffer := make([]byte, transactionStageBufferSize)
		var copied int64
		for {
			if err := requireTransactionStageContext(ctx); err != nil {
				_ = file.Close()
				return err
			}
			read, readErr := file.Read(buffer)
			if read > 0 {
				copied += int64(read)
				if copied > chunk.Bytes {
					_ = file.Close()
					return errors.New("staged transaction chunk has unexpected bytes")
				}
				if _, err := digest.Write(buffer[:read]); err != nil {
					_ = file.Close()
					return err
				}
			}
			if errors.Is(readErr, io.EOF) {
				break
			}
			if readErr != nil {
				_ = file.Close()
				return readErr
			}
			if read == 0 {
				_ = file.Close()
				return io.ErrNoProgress
			}
		}
		if err := file.Close(); err != nil {
			return err
		}
		if copied != chunk.Bytes || hex.EncodeToString(digest.Sum(nil)) != chunk.ContentDigest {
			return errors.New("staged transaction chunk failed integrity verification")
		}
	}
	return nil
}

func (s *CommittedTransactionStage) persistReceipt(ctx context.Context, receipt TransactionReceipt) error {
	if err := requireTransactionStageContext(ctx); err != nil {
		return err
	}
	receiptPayload := receipt.payload
	path := s.receiptPath(receiptPayload.transactionKey)
	if existing, err := s.storage.readFile(path); err == nil {
		var stored storedTransactionReceipt
		if decodeTransactionStageJSON(existing, &stored) == nil && stored.toReceipt(receiptPayload.transactionKey) == receipt {
			return nil
		}
		return ErrTransactionStageAlreadyCommitted
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("inspect durable transaction receipt: %w", err)
	}
	stored := storedTransactionReceipt{
		Version:             transactionStageFormatVersion,
		TransactionKey:      receiptPayload.transactionKey,
		DownstreamReceiptID: receiptPayload.downstreamReceiptID,
		Sink:                receiptPayload.sink,
		DurableAt:           receiptPayload.durableAt.UTC(),
		Bytes:               receiptPayload.bytes,
		Records:             receiptPayload.records,
		ContentDigest:       receiptPayload.contentDigest,
	}
	payload, err := json.Marshal(stored)
	if err != nil {
		return fmt.Errorf("encode durable transaction receipt: %w", err)
	}
	payload = append(payload, '\n')
	if err := s.atomicWrite(ctx, path, payload, false); err != nil {
		return fmt.Errorf("persist durable transaction receipt: %w", err)
	}
	return nil
}

func (s *CommittedTransactionStage) readReceipt(key string) (TransactionReceipt, error) {
	payload, err := s.storage.readFile(s.receiptPath(key))
	if err != nil {
		return TransactionReceipt{}, err
	}
	var stored storedTransactionReceipt
	if err := decodeTransactionStageJSON(payload, &stored); err != nil {
		return TransactionReceipt{}, fmt.Errorf("decode durable transaction receipt: %w", err)
	}
	receipt := stored.toReceipt(key)
	if !receipt.payload.durable {
		return TransactionReceipt{}, errors.New("durable transaction receipt is invalid")
	}
	return receipt, nil
}

func (r storedTransactionReceipt) toReceipt(expectedKey string) TransactionReceipt {
	if r.Version != transactionStageFormatVersion || r.TransactionKey != expectedKey || !validTransactionStageKey(r.TransactionKey) ||
		strings.TrimSpace(r.DownstreamReceiptID) == "" || strings.TrimSpace(r.Sink) == "" || r.DurableAt.IsZero() ||
		r.Bytes < 0 || r.Records < 0 || !validTransactionStageDigest(r.ContentDigest) {
		return TransactionReceipt{}
	}
	return newTransactionReceipt(
		r.TransactionKey,
		r.DownstreamReceiptID,
		r.Sink,
		r.DurableAt,
		r.Bytes,
		r.Records,
		r.ContentDigest,
	)
}

func (s *CommittedTransactionStage) writeManifest(ctx context.Context, manifest transactionStageManifest) error {
	if err := manifest.validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("encode transaction stage state: %w", err)
	}
	payload = append(payload, '\n')
	if err := s.atomicWrite(ctx, s.manifestPath(manifest.TransactionKey), payload, true); err != nil {
		return fmt.Errorf("persist transaction stage state: %w", err)
	}
	return nil
}

func (s *CommittedTransactionStage) readManifest(key string) (transactionStageManifest, error) {
	payload, err := s.storage.readFile(s.manifestPath(key))
	if err != nil {
		return transactionStageManifest{}, err
	}
	var manifest transactionStageManifest
	if err := decodeTransactionStageJSON(payload, &manifest); err != nil {
		return transactionStageManifest{}, err
	}
	if manifest.TransactionKey != key {
		return transactionStageManifest{}, errors.New("transaction stage state does not match its directory")
	}
	if err := manifest.validate(); err != nil {
		return transactionStageManifest{}, err
	}
	return manifest, nil
}

func (s *CommittedTransactionStage) persistDiscardIntent(ctx context.Context, manifest transactionStageManifest) (transactionStageDurabilityOutcome, error) {
	if err := requireTransactionStageContext(ctx); err != nil {
		return transactionStageDurabilityNotApplied, err
	}
	if discarded, err := s.readDiscardIntent(manifest); err != nil {
		return transactionStageDurabilityIndeterminate, err
	} else if discarded {
		if err := s.storage.syncDirectory(s.rootDirectory()); err != nil {
			return transactionStageDurabilityIndeterminate, fmt.Errorf("sync transaction stage root: %w", err)
		}
		if err := s.storage.syncDirectory(s.discardsDirectory()); err != nil {
			return transactionStageDurabilityIndeterminate, fmt.Errorf("sync transaction discard directory: %w", err)
		}
		return transactionStageDurabilityDurable, nil
	}
	if err := s.storage.mkdirAll(s.discardsDirectory(), 0o700); err != nil {
		return transactionStageDurabilityNotApplied, fmt.Errorf("create transaction discard directory: %w", err)
	}
	if err := s.storage.syncDirectory(s.rootDirectory()); err != nil {
		return transactionStageDurabilityNotApplied, fmt.Errorf("sync transaction stage root: %w", err)
	}
	intent := transactionStageDiscardIntent{
		Version:        transactionStageFormatVersion,
		TransactionKey: manifest.TransactionKey,
		InstanceID:     manifest.instanceID(),
		State:          transactionStageStateDiscarded,
	}
	payload, err := json.Marshal(intent)
	if err != nil {
		return transactionStageDurabilityNotApplied, fmt.Errorf("encode transaction discard intent: %w", err)
	}
	payload = append(payload, '\n')
	outcome, err := s.atomicWriteWithOutcome(ctx, s.discardIntentPath(intent.TransactionKey, intent.InstanceID), payload, false, false)
	if !errors.Is(err, ErrTransactionStageAlreadyCommitted) {
		if err != nil {
			return outcome, fmt.Errorf("persist transaction discard intent: %w", err)
		}
		return outcome, nil
	}
	if discarded, readErr := s.readDiscardIntent(manifest); readErr != nil {
		return transactionStageDurabilityIndeterminate, readErr
	} else if !discarded {
		return transactionStageDurabilityIndeterminate, errors.New("transaction discard intent disappeared")
	}
	if err := s.storage.syncDirectory(s.discardsDirectory()); err != nil {
		return transactionStageDurabilityIndeterminate, fmt.Errorf("sync transaction discard directory: %w", err)
	}
	return transactionStageDurabilityDurable, nil
}

func (s *CommittedTransactionStage) readDiscardIntent(manifest transactionStageManifest) (bool, error) {
	payload, err := s.storage.readFile(s.discardIntentPath(manifest.TransactionKey, manifest.instanceID()))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read transaction discard intent: %w", err)
	}
	var intent transactionStageDiscardIntent
	if err := decodeTransactionStageJSON(payload, &intent); err != nil {
		return false, fmt.Errorf("decode transaction discard intent: %w", err)
	}
	if err := intent.validate(manifest); err != nil {
		return false, err
	}
	return true, nil
}

func (s *CommittedTransactionStage) atomicWrite(ctx context.Context, path string, payload []byte, replace bool) error {
	_, err := s.atomicWriteWithOutcome(ctx, path, payload, replace, true)
	return err
}

func (s *CommittedTransactionStage) atomicWriteWithOutcome(ctx context.Context, path string, payload []byte, replace, removeRenamedOnError bool) (outcome transactionStageDurabilityOutcome, err error) {
	outcome = transactionStageDurabilityNotApplied
	if err := requireTransactionStageContext(ctx); err != nil {
		return outcome, err
	}
	directory := filepath.Dir(path)
	if err := s.storage.mkdirAll(directory, 0o700); err != nil {
		return outcome, fmt.Errorf("create transaction stage state directory: %w", err)
	}
	temporary, err := s.storage.createTemp(directory, ".stage.tmp-*")
	if err != nil {
		return outcome, fmt.Errorf("create temporary transaction stage state: %w", err)
	}
	temporaryPath := temporary.Name()
	closed := false
	renamed := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		if err != nil {
			_ = s.storage.remove(temporaryPath)
			if renamed && removeRenamedOnError {
				_ = s.storage.remove(path)
				_ = s.storage.syncDirectory(directory)
			}
		}
	}()
	if !replace {
		if _, readErr := s.storage.readFile(path); readErr == nil {
			return outcome, ErrTransactionStageAlreadyCommitted
		} else if !errors.Is(readErr, fs.ErrNotExist) {
			return outcome, fmt.Errorf("inspect transaction stage state: %w", readErr)
		}
	}
	if err := writeTransactionStageBytes(temporary, payload); err != nil {
		return outcome, fmt.Errorf("write temporary transaction stage state: %w", err)
	}
	if err := requireTransactionStageContext(ctx); err != nil {
		return outcome, err
	}
	if err := temporary.Sync(); err != nil {
		return outcome, fmt.Errorf("sync temporary transaction stage state: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return outcome, fmt.Errorf("close temporary transaction stage state: %w", err)
	}
	closed = true
	if err := requireTransactionStageContext(ctx); err != nil {
		return outcome, err
	}
	if err := s.storage.rename(temporaryPath, path); err != nil {
		return outcome, fmt.Errorf("rename durable transaction stage state: %w", err)
	}
	renamed = true
	outcome = transactionStageDurabilityIndeterminate
	if err := requireTransactionStageContext(ctx); err != nil {
		return outcome, err
	}
	if err := s.storage.syncDirectory(directory); err != nil {
		return outcome, fmt.Errorf("sync transaction stage state directory: %w", err)
	}
	return transactionStageDurabilityDurable, nil
}

func (m transactionStageManifest) validate() error {
	if m.Version != transactionStageFormatVersion || !validTransactionStageKey(m.TransactionKey) ||
		(m.State != transactionStageStateActive && m.State != transactionStageStateSealed && m.State != transactionStageStateDiscarded) || m.CreatedAt.IsZero() ||
		m.Bytes < 0 || m.Records < 0 || !validTransactionStageDigest(m.ContentDigest) || (m.InstanceID != "" && !validTransactionStageInstanceID(m.InstanceID)) {
		return errors.New("transaction stage state is invalid")
	}
	var bytesTotal, recordsTotal int64
	digest := transactionDigestSeed()
	for index, chunk := range m.Chunks {
		if chunk.Sequence != uint64(index) || chunk.File != transactionStageChunkFile(chunk.Sequence) || chunk.Bytes < 0 || chunk.Records <= 0 ||
			!validTransactionStageDigest(chunk.ContentDigest) || exceedsLimit(bytesTotal, chunk.Bytes, m.Bytes) || exceedsLimit(recordsTotal, chunk.Records, m.Records) {
			return errors.New("transaction stage chunk state is invalid")
		}
		bytesTotal += chunk.Bytes
		recordsTotal += chunk.Records
		digest = nextTransactionDigest(digest, chunk)
	}
	if bytesTotal != m.Bytes || recordsTotal != m.Records || digest != m.ContentDigest {
		return errors.New("transaction stage totals are invalid")
	}
	return nil
}

func (m transactionStageManifest) instanceID() string {
	if m.InstanceID != "" {
		return m.InstanceID
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte("polymetrics-transaction-stage-instance-v1\x00"))
	_, _ = hash.Write([]byte(m.TransactionKey))
	_, _ = hash.Write([]byte("\x00"))
	_, _ = hash.Write([]byte(m.CreatedAt.UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(hash.Sum(nil))
}

func (i transactionStageDiscardIntent) validate(manifest transactionStageManifest) error {
	if i.Version != transactionStageFormatVersion || i.TransactionKey != manifest.TransactionKey || i.InstanceID != manifest.instanceID() ||
		i.State != transactionStageStateDiscarded || !validTransactionStageKey(i.TransactionKey) || !validTransactionStageInstanceID(i.InstanceID) {
		return errors.New("transaction discard intent is invalid")
	}
	return nil
}

const (
	transactionStageStateActive    = "active"
	transactionStageStateSealed    = "sealed"
	transactionStageStateDiscarded = "discarded"
)

func transactionStageKey(transactionID string) (string, error) {
	if transactionID == "" || len(transactionID) > 4096 || !utf8.ValidString(transactionID) {
		return "", errors.New("transaction stage identity is invalid")
	}
	sum := sha256.Sum256(append([]byte("polymetrics-transaction-stage-v1\x00"), []byte(transactionID)...))
	return hex.EncodeToString(sum[:]), nil
}

func newTransactionStageInstanceID() (string, error) {
	var random [sha256.Size]byte
	if _, err := cryptorand.Read(random[:]); err != nil {
		return "", fmt.Errorf("generate transaction stage instance identity: %w", err)
	}
	return hex.EncodeToString(random[:]), nil
}

func validTransactionStageKey(key string) bool {
	if len(key) != transactionStageKeyBytes {
		return false
	}
	_, err := hex.DecodeString(key)
	return err == nil
}

func validTransactionStageInstanceID(instanceID string) bool {
	return validTransactionStageKey(instanceID)
}

func validTransactionStageDigest(digest string) bool {
	return validTransactionStageKey(digest)
}

func transactionDigestSeed() string {
	sum := sha256.Sum256([]byte("polymetrics-committed-transaction-v1"))
	return hex.EncodeToString(sum[:])
}

func nextTransactionDigest(previous string, chunk transactionStageChunk) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte("polymetrics-committed-transaction-chunk-v1\x00"))
	previousBytes, _ := hex.DecodeString(previous)
	_, _ = hash.Write(previousBytes)
	var encoded [24]byte
	binary.BigEndian.PutUint64(encoded[0:8], chunk.Sequence)
	binary.BigEndian.PutUint64(encoded[8:16], uint64(chunk.Records))
	binary.BigEndian.PutUint64(encoded[16:24], uint64(chunk.Bytes))
	_, _ = hash.Write(encoded[:])
	chunkBytes, _ := hex.DecodeString(chunk.ContentDigest)
	_, _ = hash.Write(chunkBytes)
	return hex.EncodeToString(hash.Sum(nil))
}

func transactionStageChunkFile(sequence uint64) string {
	return fmt.Sprintf("chunk-%020d.bin", sequence)
}

func cloneTransactionStageManifest(manifest transactionStageManifest) transactionStageManifest {
	clone := manifest
	clone.Chunks = append([]transactionStageChunk(nil), manifest.Chunks...)
	return clone
}

func requireTransactionStageContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("transaction stage context is required")
	}
	return ctx.Err()
}

func transactionStageCleanupContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}

func (s *CommittedTransactionStage) currentTime() (time.Time, error) {
	now := s.now()
	if now.IsZero() {
		return time.Time{}, errors.New("transaction stage clock returned zero time")
	}
	return now.UTC(), nil
}

func (s *CommittedTransactionStage) ageLimitExceeded(manifest transactionStageManifest) error {
	now, err := s.currentTime()
	if err != nil {
		return err
	}
	elapsed := now.Sub(manifest.CreatedAt)
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > s.limits.MaxTransactionAge {
		return &TransactionStageLimitExceeded{
			Limit:    TransactionStageLimitTransactionAge,
			Maximum:  int64(s.limits.MaxTransactionAge),
			Observed: int64(elapsed),
		}
	}
	return nil
}

func exceedsLimit(current, addition, maximum int64) bool {
	_, withinLimit := transactionStageLimitedAdd(current, addition, maximum)
	return !withinLimit
}

func transactionStageLimitedAdd(current, addition, maximum int64) (int64, bool) {
	if current < 0 || addition < 0 || maximum < 0 || current > maximum || addition > maximum-current {
		return 0, false
	}
	return current + addition, true
}

func transactionStageSaturatingAdd(current, addition int64) int64 {
	if current < 0 || addition < 0 || current > transactionStageMaximumInt64-addition {
		return transactionStageMaximumInt64
	}
	return current + addition
}

func writeTransactionStageBytes(writer io.Writer, payload []byte) error {
	for len(payload) > 0 {
		written, err := writer.Write(payload)
		if written > 0 {
			payload = payload[written:]
		}
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}

func (s *CommittedTransactionStage) releaseReservationLocked(bytes int64) {
	if bytes <= 0 {
		return
	}
	if bytes >= s.inFlightBytes {
		s.inFlightBytes = 0
		return
	}
	s.inFlightBytes -= bytes
}

func (s *CommittedTransactionStage) subtractStagedBytesLocked(bytes int64) {
	if bytes <= 0 {
		return
	}
	if bytes >= s.stagedBytes {
		s.stagedBytes = 0
		return
	}
	s.stagedBytes -= bytes
}

func withTransactionStageCleanupError(cause, cleanupErr error) error {
	if cause == nil {
		return cleanupErr
	}
	if cleanupErr == nil {
		return cause
	}
	return fmt.Errorf("%w: clean transaction stage: %v", cause, cleanupErr)
}

func transactionStageDiscardError(intentOutcome, cleanupOutcome transactionStageDurabilityOutcome, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("discard transaction stage (%s intent, %s cleanup): %w", intentOutcome, cleanupOutcome, err)
}

func (s *CommittedTransactionStage) removeStageDirectory(key string) error {
	return s.removeOrphan(s.transactionDirectory(key), s.transactionsDirectory())
}

func (s *CommittedTransactionStage) removeStageDirectoryWithOutcome(key string) (transactionStageDurabilityOutcome, error) {
	if err := s.storage.removeAll(s.transactionDirectory(key)); err != nil {
		return transactionStageDurabilityIndeterminate, fmt.Errorf("remove incomplete transaction stage artifact: %w", err)
	}
	if err := s.storage.syncDirectory(s.transactionsDirectory()); err != nil {
		return transactionStageDurabilityIndeterminate, fmt.Errorf("sync transaction stage cleanup: %w", err)
	}
	return transactionStageDurabilityDurable, nil
}

func (s *CommittedTransactionStage) removeOrphan(path, parent string) error {
	if err := s.storage.removeAll(path); err != nil {
		return fmt.Errorf("remove incomplete transaction stage artifact: %w", err)
	}
	if err := s.storage.syncDirectory(parent); err != nil {
		return fmt.Errorf("sync transaction stage cleanup: %w", err)
	}
	return nil
}

func (s *CommittedTransactionStage) rootDirectory() string { return s.root }

func (s *CommittedTransactionStage) transactionsDirectory() string {
	return filepath.Join(s.rootDirectory(), "transactions")
}

func (s *CommittedTransactionStage) receiptsDirectory() string {
	return filepath.Join(s.rootDirectory(), "receipts")
}

func (s *CommittedTransactionStage) discardsDirectory() string {
	return filepath.Join(s.rootDirectory(), "discards")
}

func (s *CommittedTransactionStage) transactionDirectory(key string) string {
	return filepath.Join(s.transactionsDirectory(), key)
}

func (s *CommittedTransactionStage) chunksDirectory(key string) string {
	return filepath.Join(s.transactionDirectory(key), "chunks")
}

func (s *CommittedTransactionStage) manifestPath(key string) string {
	return filepath.Join(s.transactionDirectory(key), "state.json")
}

func (s *CommittedTransactionStage) chunkPath(key string, sequence uint64) string {
	return filepath.Join(s.chunksDirectory(key), transactionStageChunkFile(sequence))
}

func (s *CommittedTransactionStage) receiptPath(key string) string {
	return filepath.Join(s.receiptsDirectory(), key+".json")
}

func (s *CommittedTransactionStage) discardIntentPath(key, instanceID string) string {
	return filepath.Join(s.discardsDirectory(), key+"-"+instanceID+".json")
}

func decodeTransactionStageJSON(payload []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("transaction stage state has multiple JSON values")
		}
		return err
	}
	return nil
}

type transactionStageContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *transactionStageContextReader) Read(payload []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(payload)
}
