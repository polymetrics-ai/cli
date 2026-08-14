package database

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

var (
	// ErrDatabaseWritePlanInvalid refuses an unsealed, incompatible, or
	// resource-unbounded database apply plan before any driver interaction.
	ErrDatabaseWritePlanInvalid = errors.New("database write plan is invalid")
	// ErrDatabaseWriteApprovalInvalid refuses an approval whose preview does
	// not bind the exact sealed plan offered for execution.
	ErrDatabaseWriteApprovalInvalid = errors.New("database write approval does not match the sealed plan")
	// ErrDatabaseWriteApprovalConsumed refuses a second use of an accepted
	// approval. Consumption happens before a session can mutate a target.
	ErrDatabaseWriteApprovalConsumed = errors.New("database write approval is already consumed")
	// ErrDatabaseWritePreviewUnavailable hides driver-specific preview errors.
	ErrDatabaseWritePreviewUnavailable = errors.New("database write preview is unavailable")
	// ErrDatabaseWriteSessionUnavailable hides driver-specific session-open
	// errors and rejects a typed-nil returned session.
	ErrDatabaseWriteSessionUnavailable = errors.New("database write session is unavailable")
	// ErrDatabaseWriteBatchFailed hides arbitrary driver record/database detail
	// after the executor has rolled back the entire session.
	ErrDatabaseWriteBatchFailed = errors.New("database write batch failed")
	// ErrDatabaseWriteRollbackFailed reports that a session could not prove its
	// rollback after a pre-commit execution failure.
	ErrDatabaseWriteRollbackFailed = errors.New("database write rollback failed")
	// ErrDatabaseWriteReceiptUnavailable prevents a checkpoint acknowledgement
	// when no durable, ledger-recorded commit receipt exists.
	ErrDatabaseWriteReceiptUnavailable = errors.New("database write durable receipt is unavailable")
	// ErrDatabaseWriteCommitOutcomeUnknown is terminal for one execution. It
	// deliberately does not claim rollback and must never trigger a blind retry.
	ErrDatabaseWriteCommitOutcomeUnknown = errors.New("database write commit outcome is unknown")
	// ErrDatabaseWriteRolledBack reports an explicit driver-confirmed rollback
	// outcome. It is distinct from an unknown commit outcome.
	ErrDatabaseWriteRolledBack = errors.New("database write was rolled back")
)

// CommitOutcome is the only session-commit state vocabulary. Unknown means a
// driver cannot prove whether the target committed; it is never retried or
// normalized to rolled back by this shared layer.
type CommitOutcome string

const (
	CommitOutcomeCommitted  CommitOutcome = "committed"
	CommitOutcomeRolledBack CommitOutcome = "rolled_back"
	CommitOutcomeUnknown    CommitOutcome = "unknown"
)

func (o CommitOutcome) valid() bool {
	return o == CommitOutcomeCommitted || o == CommitOutcomeRolledBack || o == CommitOutcomeUnknown
}

// DatabaseWriteCapabilities contains the finite native guarantees that are
// needed before the shared layer opens a write session. It is intentionally
// not a capability bit exposed by a connector manifest.
type DatabaseWriteCapabilities struct {
	AtomicFullOverwrite bool
}

// DatabaseWritePlanRequest contains the only authority accepted for a shared
// database write. It names neither SQL nor a relation/connection string:
// control is an asserted managed target and Definition supplies the declared
// resource bound and admitted closed mode.
type DatabaseWritePlanRequest struct {
	Definition  Definition
	Control     ManagedTargetControlRecord
	Mode        synccontract.Mode
	Strategy    connectors.ApplyStrategy
	Keys        []string
	RecordCount int
	BatchSize   int
	Destructive bool
}

// DatabaseWritePlan is an immutable, non-executing target application plan.
// Its fields stay private so approval can compare the complete original
// binding rather than a caller-mutable projection.
type DatabaseWritePlan struct {
	driver      DriverDeclaration
	control     ManagedTargetControlRecord
	mode        synccontract.Mode
	strategy    connectors.ApplyStrategy
	keys        []string
	recordCount int
	batchSize   int
	destructive bool
}

// NewDatabaseWritePlan seals a typed target/mode/key/count/effects contract
// before preview or session opening. The declared definition must admit the
// requested mode and supplies the finite batch bound.
func NewDatabaseWritePlan(ctx context.Context, request DatabaseWritePlanRequest) (DatabaseWritePlan, error) {
	if ctx == nil {
		return DatabaseWritePlan{}, ErrDatabaseWritePlanInvalid
	}
	if err := ctx.Err(); err != nil {
		return DatabaseWritePlan{}, err
	}
	if err := request.Definition.Validate(); err != nil || request.Control.validate() != nil || request.Mode.Validate() != nil || request.Strategy.Validate() != nil {
		return DatabaseWritePlan{}, ErrDatabaseWritePlanInvalid
	}
	if !definitionAdmitsMode(request.Definition.AdmittedModes(), request.Mode) || canonicalDatabaseWriteStrategy(request.Mode) != request.Strategy {
		return DatabaseWritePlan{}, ErrDatabaseWritePlanInvalid
	}
	if request.RecordCount <= 0 {
		return DatabaseWritePlan{}, ErrDatabaseWritePlanInvalid
	}
	batchSize, err := request.Definition.Resources().EffectiveBatchSize(request.BatchSize)
	if err != nil {
		return DatabaseWritePlan{}, ErrDatabaseWritePlanInvalid
	}
	keys, err := normalizeDatabaseWriteKeys(request.Keys)
	if err != nil {
		return DatabaseWritePlan{}, ErrDatabaseWritePlanInvalid
	}
	if databaseWriteModeRequiresKeys(request.Mode) != (len(keys) > 0) {
		return DatabaseWritePlan{}, ErrDatabaseWritePlanInvalid
	}
	if (request.Mode == synccontract.ModeFullOverwrite) != request.Destructive {
		return DatabaseWritePlan{}, ErrDatabaseWritePlanInvalid
	}
	if err := ctx.Err(); err != nil {
		return DatabaseWritePlan{}, err
	}
	return DatabaseWritePlan{
		driver:      request.Definition.Driver(),
		control:     request.Control,
		mode:        request.Mode,
		strategy:    request.Strategy,
		keys:        keys,
		recordCount: request.RecordCount,
		batchSize:   batchSize,
		destructive: request.Destructive,
	}, nil
}

func (p DatabaseWritePlan) validate() error {
	if p.driver.validate() != nil || p.control.validate() != nil || p.mode.Validate() != nil || p.strategy.Validate() != nil || p.recordCount <= 0 || p.batchSize <= 0 || canonicalDatabaseWriteStrategy(p.mode) != p.strategy {
		return ErrDatabaseWritePlanInvalid
	}
	keys, err := normalizeDatabaseWriteKeys(p.keys)
	if err != nil || len(keys) != len(p.keys) || databaseWriteModeRequiresKeys(p.mode) != (len(keys) > 0) || (p.mode == synccontract.ModeFullOverwrite) != p.destructive {
		return ErrDatabaseWritePlanInvalid
	}
	for index := range keys {
		if keys[index] != p.keys[index] {
			return ErrDatabaseWritePlanInvalid
		}
	}
	return nil
}

func (p DatabaseWritePlan) matchesDriver(driver DatabaseWriteDriver) bool {
	return !isNilInterface(driver) && driver.DatabaseDriverDescriptor().declaration() == p.driver
}

// Control returns the asserted target record bound into this plan.
func (p DatabaseWritePlan) Control() ManagedTargetControlRecord { return p.control }

// Mode returns the closed synchronization mode selected for this plan.
func (p DatabaseWritePlan) Mode() synccontract.Mode { return p.mode }

// Strategy returns the canonical closed database apply strategy for Mode.
func (p DatabaseWritePlan) Strategy() connectors.ApplyStrategy { return p.strategy }

// Keys returns an independent ordered projection of the stable key mapping.
func (p DatabaseWritePlan) Keys() []string { return append([]string(nil), p.keys...) }

// RecordCount returns the exact approved bounded workset count.
func (p DatabaseWritePlan) RecordCount() int { return p.recordCount }

// BatchSize returns the finite maximum batch size derived from database.json.
func (p DatabaseWritePlan) BatchSize() int { return p.batchSize }

// Destructive reports whether the approved plan carries the required
// destructive effect. Only full_overwrite is destructive in this layer.
func (p DatabaseWritePlan) Destructive() bool { return p.destructive }

func (p DatabaseWritePlan) matches(other DatabaseWritePlan) bool {
	if p.driver != other.driver || p.mode != other.mode || p.strategy != other.strategy || p.recordCount != other.recordCount || p.batchSize != other.batchSize || p.destructive != other.destructive || len(p.keys) != len(other.keys) || !sameManagedTargetControl(p.control, other.control) {
		return false
	}
	for index := range p.keys {
		if p.keys[index] != other.keys[index] {
			return false
		}
	}
	return true
}

func definitionAdmitsMode(modes []synccontract.Mode, mode synccontract.Mode) bool {
	for _, declared := range modes {
		if declared == mode {
			return true
		}
	}
	return false
}

func canonicalDatabaseWriteStrategy(mode synccontract.Mode) connectors.ApplyStrategy {
	switch mode {
	case synccontract.ModeFullAppend, synccontract.ModeIncrementalAppend:
		return connectors.ApplyStrategyAppend
	case synccontract.ModeFullOverwrite:
		return connectors.ApplyStrategyReplace
	case synccontract.ModeIncrementalUpsert:
		return connectors.ApplyStrategyMerge
	case synccontract.ModeIncrementalDedupe:
		return connectors.ApplyStrategyDedupe
	case synccontract.ModeIncrementalDedupeHistory:
		return connectors.ApplyStrategyDedupeHistory
	case synccontract.ModeChangeCapture:
		return connectors.ApplyStrategyChangeApply
	default:
		return ""
	}
}

func databaseWriteModeRequiresKeys(mode synccontract.Mode) bool {
	return mode == synccontract.ModeIncrementalUpsert || mode == synccontract.ModeIncrementalDedupe || mode == synccontract.ModeIncrementalDedupeHistory
}

func normalizeDatabaseWriteKeys(keys []string) ([]string, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	clone := append([]string(nil), keys...)
	seen := make(map[string]struct{}, len(clone))
	for _, key := range clone {
		if validateIdentifierComponent(key) != nil {
			return nil, ErrDatabaseWritePlanInvalid
		}
		if _, exists := seen[key]; exists {
			return nil, ErrDatabaseWritePlanInvalid
		}
		seen[key] = struct{}{}
	}
	return clone, nil
}

func sameManagedTargetControl(left, right ManagedTargetControlRecord) bool {
	leftTarget, rightTarget := left.Target(), right.Target()
	return left.Owner().Identity().SameIdentity(right.Owner().Identity()) &&
		leftTarget.Owner().Identity().SameIdentity(rightTarget.Owner().Identity()) &&
		leftTarget.StreamID() == rightTarget.StreamID() &&
		leftTarget.Namespace() == rightTarget.Namespace() &&
		leftTarget.Relation() == rightTarget.Relation() &&
		left.TargetDatabase().sameIdentity(right.TargetDatabase()) &&
		left.NativeIdentity() == right.NativeIdentity() &&
		left.Schema().sameSchema(right.Schema())
}

// DatabaseWritePreview is a driver-observed preview of one sealed plan. The
// preview ID is opaque evidence and never a target name, query, or credential.
type DatabaseWritePreview struct {
	plan      DatabaseWritePlan
	previewID string
}

// NewDatabaseWritePreview permits a native driver to return only opaque,
// plan-bound preview evidence. It cannot bind an approval to another target.
func NewDatabaseWritePreview(plan DatabaseWritePlan, previewID string) (DatabaseWritePreview, error) {
	if plan.validate() != nil || !validOpaqueIdentityComponent(previewID) {
		return DatabaseWritePreview{}, ErrDatabaseWritePreviewUnavailable
	}
	return DatabaseWritePreview{plan: plan, previewID: previewID}, nil
}

func (p DatabaseWritePreview) validate() error {
	if p.plan.validate() != nil || !validOpaqueIdentityComponent(p.previewID) {
		return ErrDatabaseWritePreviewUnavailable
	}
	return nil
}

// PreviewID returns opaque preview evidence suitable for audit correlation.
func (p DatabaseWritePreview) PreviewID() string { return p.previewID }

func (p DatabaseWritePreview) matches(plan DatabaseWritePlan) bool {
	return p.validate() == nil && plan.validate() == nil && p.plan.matches(plan)
}

type databaseWriteApprovalUse struct {
	consumed atomic.Bool
}

// DatabaseWriteApproval is a one-shot in-memory admission produced only from
// a completed database preview. It contains no user approval token or raw
// credential, and it is consumed before a session is opened.
type DatabaseWriteApproval struct {
	preview DatabaseWritePreview
	use     *databaseWriteApprovalUse
}

// NewDatabaseWriteApproval derives a one-shot approval from a valid preview.
func NewDatabaseWriteApproval(preview DatabaseWritePreview) (*DatabaseWriteApproval, error) {
	if preview.validate() != nil {
		return nil, ErrDatabaseWriteApprovalInvalid
	}
	return &DatabaseWriteApproval{preview: preview, use: &databaseWriteApprovalUse{}}, nil
}

func (a *DatabaseWriteApproval) consume(plan DatabaseWritePlan) error {
	if a == nil || a.use == nil || !a.preview.matches(plan) {
		return ErrDatabaseWriteApprovalInvalid
	}
	if !a.use.consumed.CompareAndSwap(false, true) {
		return ErrDatabaseWriteApprovalConsumed
	}
	return nil
}

// Consumed reports whether this one-shot approval has been admitted for a
// session. It is observability only; it cannot reset or otherwise alter the
// approval state.
func (a *DatabaseWriteApproval) Consumed() bool {
	return a != nil && a.use != nil && a.use.consumed.Load()
}

// WriteBatch is an executor-created bounded payload for one pinned session.
// It provides no caller-selected mode, relation, SQL, or per-record commit.
type WriteBatch struct {
	sequence uint64
	records  []connectors.Record
}

func newWriteBatch(sequence uint64, records []connectors.Record) (WriteBatch, error) {
	if sequence == 0 || len(records) == 0 {
		return WriteBatch{}, ErrDatabaseWritePlanInvalid
	}
	return WriteBatch{sequence: sequence, records: cloneDatabaseWriteRecords(records)}, nil
}

func (b WriteBatch) validate(limit int) error {
	if b.sequence == 0 || len(b.records) == 0 || len(b.records) > limit {
		return ErrDatabaseWritePlanInvalid
	}
	return nil
}

// Sequence returns the one-based session-local batch order.
func (b WriteBatch) Sequence() uint64 { return b.sequence }

// Records returns an independent top-level record projection for this call.
func (b WriteBatch) Records() []connectors.Record { return cloneDatabaseWriteRecords(b.records) }

func cloneDatabaseWriteRecords(records []connectors.Record) []connectors.Record {
	clone := make([]connectors.Record, len(records))
	for index, record := range records {
		clone[index] = make(connectors.Record, len(record))
		for key, value := range record {
			clone[index][key] = value
		}
	}
	return clone
}

// DatabaseWriteReceipt is target durability evidence from a confirmed commit.
// It is plan-bound and becomes checkpoint authority only after the managed
// target delivery ledger records its opaque delivery identifier.
type DatabaseWriteReceipt struct {
	plan        DatabaseWritePlan
	delivery    ManagedTargetDeliveryRecord
	committedAt time.Time
}

// NewDatabaseWriteReceipt constructs target durability evidence for exactly
// one sealed plan. Native drivers must call it only after their own transaction
// protocol has confirmed durable commit.
func NewDatabaseWriteReceipt(plan DatabaseWritePlan, deliveryID string, committedAt time.Time) (DatabaseWriteReceipt, error) {
	delivery, err := NewManagedTargetDeliveryRecord(deliveryID)
	if err != nil || plan.validate() != nil || committedAt.IsZero() {
		return DatabaseWriteReceipt{}, ErrDatabaseWriteReceiptUnavailable
	}
	return DatabaseWriteReceipt{plan: plan, delivery: delivery, committedAt: committedAt.UTC()}, nil
}

func (r DatabaseWriteReceipt) validateFor(plan DatabaseWritePlan) error {
	if plan.validate() != nil || r.plan.validate() != nil || !r.plan.matches(plan) || r.delivery.validate() != nil || r.committedAt.IsZero() {
		return ErrDatabaseWriteReceiptUnavailable
	}
	return nil
}

// DeliveryID returns the opaque target durable-delivery identifier.
func (r DatabaseWriteReceipt) DeliveryID() string { return r.delivery.DeliveryID() }

// CommittedAt returns the confirmed native commit time in UTC.
func (r DatabaseWriteReceipt) CommittedAt() time.Time { return r.committedAt }

// DatabaseWriteDriver is the native-driver port consumed by the shared layer.
// It has no SQL strings, arbitrary relation, or per-record write operation.
type DatabaseWriteDriver interface {
	Driver
	DatabaseWriteCapabilities() DatabaseWriteCapabilities
	PreviewDatabaseWrite(context.Context, DatabaseWritePlan) (DatabaseWritePreview, error)
	BeginDatabaseWrite(context.Context, DatabaseWritePlan) (WriteSession, error)
}

// WriteSession is one native pinned transaction. ApplyWriteBatch is the only
// mutation operation; CommitWrite reports certainty explicitly and rollback is
// owned by this exact pinned session.
type WriteSession interface {
	ApplyWriteBatch(context.Context, WriteBatch) error
	PublishFullOverwrite(context.Context) error
	CommitWrite(context.Context) (CommitOutcome, DatabaseWriteReceipt, error)
	RollbackWrite(context.Context) error
}

// DatabaseWriteResult records what the executor can prove. An acknowledgement
// is intentionally unavailable until a committed target receipt is also
// persisted by the managed-target delivery ledger.
type DatabaseWriteResult struct {
	outcome CommitOutcome
	receipt DatabaseWriteReceipt
	ledger  bool
}

// Outcome returns the final known transaction state.
func (r DatabaseWriteResult) Outcome() CommitOutcome { return r.outcome }

// Receipt returns durable target evidence only for a confirmed commit.
func (r DatabaseWriteResult) Receipt() (DatabaseWriteReceipt, bool) {
	if r.outcome != CommitOutcomeCommitted || !r.ledger {
		return DatabaseWriteReceipt{}, false
	}
	return r.receipt, true
}

// DownstreamAcknowledgement returns the existing checkpoint authority only
// after confirmed commit and durable ledger evidence. Callers still hand it to
// synccontract.CommitAfterDownstreamAcknowledgement; this package never
// advances a source checkpoint itself.
func (r DatabaseWriteResult) DownstreamAcknowledgement() (synccontract.DownstreamAcknowledgement, error) {
	receipt, ok := r.Receipt()
	if !ok {
		return synccontract.DownstreamAcknowledgement{}, ErrDatabaseWriteReceiptUnavailable
	}
	return synccontract.NewDurableDownstreamAcknowledgement(receipt.plan.Control().TargetDatabase().Kind(), receipt.CommittedAt())
}

// DatabaseWriteExecutor coordinates preview, one-shot approval, one pinned
// session, receipt recording, and checkpoint eligibility. It is driver-neutral
// and intentionally distinct from CommittedTransactionStage.
type DatabaseWriteExecutor struct {
	driver DatabaseWriteDriver
	ledger *ManagedTargetDeliveryLedger
}

// NewDatabaseWriteExecutor creates the shared executor around a concrete
// native driver port and the durable managed-target delivery ledger.
func NewDatabaseWriteExecutor(driver DatabaseWriteDriver, ledger *ManagedTargetDeliveryLedger) (*DatabaseWriteExecutor, error) {
	if isNilInterface(driver) || ledger == nil || isNilInterface(ledger.store) {
		return nil, ErrDatabaseWriteSessionUnavailable
	}
	return &DatabaseWriteExecutor{driver: driver, ledger: ledger}, nil
}

// Preview obtains native, opaque preview evidence for one sealed plan. A
// driver cannot return evidence for a different plan and have it approved.
func (e *DatabaseWriteExecutor) Preview(ctx context.Context, plan DatabaseWritePlan) (DatabaseWritePreview, error) {
	if ctx == nil || plan.validate() != nil || e == nil || isNilInterface(e.driver) || !plan.matchesDriver(e.driver) || e.ledger == nil {
		return DatabaseWritePreview{}, ErrDatabaseWritePreviewUnavailable
	}
	if err := ctx.Err(); err != nil {
		return DatabaseWritePreview{}, err
	}
	preview, err := e.driver.PreviewDatabaseWrite(ctx, plan)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return DatabaseWritePreview{}, contextErr
		}
		return DatabaseWritePreview{}, ErrDatabaseWritePreviewUnavailable
	}
	if !preview.matches(plan) {
		return DatabaseWritePreview{}, ErrDatabaseWritePreviewUnavailable
	}
	return preview, nil
}

// Execute consumes approval before it opens one session, runs the plan's
// bounded batches, and records durable target evidence only after confirmed
// commit. Neither a batch failure/cancellation nor unknown commit can produce
// a checkpoint acknowledgement.
func (e *DatabaseWriteExecutor) Execute(ctx context.Context, plan DatabaseWritePlan, approval *DatabaseWriteApproval, records []connectors.Record) (DatabaseWriteResult, error) {
	if ctx == nil || plan.validate() != nil || e == nil || isNilInterface(e.driver) || !plan.matchesDriver(e.driver) || e.ledger == nil || isNilInterface(e.ledger.store) || len(records) != plan.RecordCount() {
		return DatabaseWriteResult{}, ErrDatabaseWritePlanInvalid
	}
	if err := ctx.Err(); err != nil {
		return DatabaseWriteResult{}, err
	}
	if plan.Mode() == synccontract.ModeFullOverwrite && !e.driver.DatabaseWriteCapabilities().AtomicFullOverwrite {
		return DatabaseWriteResult{}, ErrDatabaseWritePlanInvalid
	}
	if err := approval.consume(plan); err != nil {
		return DatabaseWriteResult{}, err
	}
	session, err := e.driver.BeginDatabaseWrite(ctx, plan)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return DatabaseWriteResult{}, contextErr
		}
		return DatabaseWriteResult{}, ErrDatabaseWriteSessionUnavailable
	}
	if isNilInterface(session) {
		return DatabaseWriteResult{}, ErrDatabaseWriteSessionUnavailable
	}

	rollback := func(cause error) (DatabaseWriteResult, error) {
		result := DatabaseWriteResult{outcome: CommitOutcomeRolledBack}
		if err := session.RollbackWrite(context.WithoutCancel(ctx)); err != nil {
			return result, ErrDatabaseWriteRollbackFailed
		}
		return result, cause
	}

	for start, sequence := 0, uint64(1); start < len(records); start, sequence = start+plan.BatchSize(), sequence+1 {
		if err := ctx.Err(); err != nil {
			return rollback(err)
		}
		end := start + plan.BatchSize()
		if end > len(records) {
			end = len(records)
		}
		batch, err := newWriteBatch(sequence, records[start:end])
		if err != nil || batch.validate(plan.BatchSize()) != nil {
			return rollback(ErrDatabaseWritePlanInvalid)
		}
		if err := session.ApplyWriteBatch(ctx, batch); err != nil {
			if contextErr := ctx.Err(); contextErr != nil {
				return rollback(contextErr)
			}
			return rollback(ErrDatabaseWriteBatchFailed)
		}
	}
	if err := ctx.Err(); err != nil {
		return rollback(err)
	}
	if plan.Mode() == synccontract.ModeFullOverwrite {
		if err := session.PublishFullOverwrite(ctx); err != nil {
			if contextErr := ctx.Err(); contextErr != nil {
				return rollback(contextErr)
			}
			return rollback(ErrDatabaseWriteBatchFailed)
		}
	}
	if err := ctx.Err(); err != nil {
		return rollback(err)
	}

	outcome, receipt, commitErr := session.CommitWrite(ctx)
	// After CommitWrite returns, the driver may have durably committed even
	// when it returned an error. Never issue a compensating rollback or retry.
	if outcome == CommitOutcomeUnknown || !outcome.valid() || (outcome == CommitOutcomeCommitted && commitErr != nil) {
		return DatabaseWriteResult{outcome: CommitOutcomeUnknown}, ErrDatabaseWriteCommitOutcomeUnknown
	}
	if outcome == CommitOutcomeRolledBack {
		return DatabaseWriteResult{outcome: CommitOutcomeRolledBack}, ErrDatabaseWriteRolledBack
	}
	if receipt.validateFor(plan) != nil {
		return DatabaseWriteResult{outcome: CommitOutcomeUnknown}, ErrDatabaseWriteCommitOutcomeUnknown
	}

	key, err := NewManagedTargetDeliveryLedgerKey(plan.Control())
	if err != nil {
		return DatabaseWriteResult{outcome: CommitOutcomeCommitted, receipt: receipt}, ErrDatabaseWriteReceiptUnavailable
	}
	if err := e.ledger.Record(context.WithoutCancel(ctx), key, receipt.delivery); err != nil {
		return DatabaseWriteResult{outcome: CommitOutcomeCommitted, receipt: receipt}, ErrDatabaseWriteReceiptUnavailable
	}
	return DatabaseWriteResult{outcome: CommitOutcomeCommitted, receipt: receipt, ledger: true}, nil
}
