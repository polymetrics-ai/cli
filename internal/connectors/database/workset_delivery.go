package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

var (
	// ErrChangeDeliveryPlanInvalid refuses a workset delivery whose immutable
	// workset, asserted managed target, mapping, or key bindings do not agree.
	// It is returned before preview or session work.
	ErrChangeDeliveryPlanInvalid = errors.New("change delivery plan is invalid")
	// ErrChangeDeliveryApprovalInvalid refuses an approval produced for a
	// different immutable workset delivery plan before a target session opens.
	ErrChangeDeliveryApprovalInvalid = errors.New("change delivery approval does not match the sealed plan")
	// ErrChangeDeliveryApprovalConsumed refuses a second use of one accepted
	// delivery approval before another target session can begin.
	ErrChangeDeliveryApprovalConsumed = errors.New("change delivery approval is already consumed")
	// ErrChangeDeliveryBaselineUnavailable means target receipt/ledger evidence
	// exists but the candidate baseline could not become durable. It preserves
	// replay evidence and must never claim downstream acknowledgement.
	ErrChangeDeliveryBaselineUnavailable = errors.New("change delivery baseline is unavailable")
	// ErrChangeDeliveryReplayRequired makes an unknown target commit explicit:
	// the immutable workset identity is retained, but this executor never
	// retries or promotes a candidate baseline blindly.
	ErrChangeDeliveryReplayRequired = errors.New("change delivery replay is required")
)

// ChangeDeliveryPlanRequest contains the only inputs accepted to turn a
// sealed Parquet workset into a keyed database write plan. It deliberately
// accepts no relation text, SQL, target connection, or caller-supplied record
// counts: those are derived from the workset and asserted control record.
type ChangeDeliveryPlanRequest struct {
	Definition Definition
	Workset    ChangeDeliveryWorkset
	Control    ManagedTargetControlRecord
	Mapping    MappingContractV1
	BatchSize  int
}

// ChangeDeliveryPlan seals the immutable workset identity/content together
// with the existing database write plan. Its fields remain private so an
// approval can never be reused with another workset that merely has matching
// row counts.
type ChangeDeliveryPlan struct {
	workset ChangeDeliveryWorkset
	write   DatabaseWritePlan
	hash    string
}

// NewChangeDeliveryPlan validates an immutable workset against the exact
// managed target before deriving the only supported outbound mode:
// incremental_upsert. Source-key names from the workset are translated through
// MappingContractV1 to the target-key names consumed by the write driver.
func NewChangeDeliveryPlan(ctx context.Context, request ChangeDeliveryPlanRequest) (ChangeDeliveryPlan, error) {
	if ctx == nil {
		return ChangeDeliveryPlan{}, ErrChangeDeliveryPlanInvalid
	}
	if err := ctx.Err(); err != nil {
		return ChangeDeliveryPlan{}, err
	}
	if request.Control.validate() != nil || request.Mapping.validate() != nil || request.Workset.manifest.validate() != nil || request.Workset.dir == "" {
		return ChangeDeliveryPlan{}, ErrChangeDeliveryPlanInvalid
	}

	// Reopen and hash every immutable artifact at the boundary. The opaque
	// value never exports its path, but a changed on-disk artifact must still be
	// rejected before its content becomes a target mutation.
	workset, err := openChangeDeliveryWorkset(ctx, request.Workset.dir, request.Workset.manifest.Identity, request.Workset.manifest.MaxArtifactBytes)
	if err != nil {
		if errors.Is(err, ErrChangeDeliveryWorksetInvalid) {
			return ChangeDeliveryPlan{}, ErrChangeDeliveryPlanInvalid
		}
		return ChangeDeliveryPlan{}, ErrChangeDeliveryWorksetUnavailable
	}
	if !worksetMatchesControl(workset, request.Control) {
		return ChangeDeliveryPlan{}, ErrChangeDeliveryPlanInvalid
	}
	targetKeys, err := mappedChangeDeliveryKeys(workset.manifest.Keys, request.Mapping)
	if err != nil {
		return ChangeDeliveryPlan{}, ErrChangeDeliveryPlanInvalid
	}
	if workset.Changes() > int64(maxChangeDeliveryPlanCount) || workset.TombstoneCount() > int64(maxChangeDeliveryPlanCount) {
		return ChangeDeliveryPlan{}, ErrChangeDeliveryPlanInvalid
	}
	write, err := NewDatabaseWritePlan(ctx, DatabaseWritePlanRequest{
		Definition:     request.Definition,
		Control:        request.Control,
		Mode:           synccontract.ModeIncrementalUpsert,
		Strategy:       connectors.ApplyStrategyMerge,
		Mapping:        request.Mapping,
		Keys:           targetKeys,
		RecordCount:    int(workset.Changes()),
		TombstoneCount: int(workset.TombstoneCount()),
		BatchSize:      request.BatchSize,
		Destructive:    false,
	})
	if err != nil {
		return ChangeDeliveryPlan{}, ErrChangeDeliveryPlanInvalid
	}
	return ChangeDeliveryPlan{workset: workset, write: write, hash: changeDeliveryPlanHash(workset, write)}, nil
}

const maxChangeDeliveryPlanCount = int(^uint(0) >> 1)

func worksetMatchesControl(workset ChangeDeliveryWorkset, control ManagedTargetControlRecord) bool {
	if workset.manifest.validate() != nil || control.validate() != nil {
		return false
	}
	key, err := NewManagedTargetDeliveryLedgerKey(control)
	if err != nil || changeDeliveryBindingFromLedgerKey(key) != workset.manifest.Binding {
		return false
	}
	schema := control.Schema()
	return schema.Version() == workset.manifest.SchemaVersion && schema.Fingerprint().String() == workset.manifest.SchemaFingerprint
}

func mappedChangeDeliveryKeys(sourceKeys []string, mapping MappingContractV1) ([]string, error) {
	if mapping.validate() != nil || len(sourceKeys) == 0 {
		return nil, ErrChangeDeliveryPlanInvalid
	}
	columns := mapping.Columns()
	targetBySource := make(map[string]string, len(columns))
	for _, column := range columns {
		targetBySource[column.Source] = column.Target
	}
	targetKeys := make([]string, len(sourceKeys))
	for index, sourceKey := range sourceKeys {
		targetKey, found := targetBySource[sourceKey]
		if !found {
			return nil, ErrChangeDeliveryPlanInvalid
		}
		targetKeys[index] = targetKey
	}
	return normalizeDatabaseWriteKeys(targetKeys)
}

func (p ChangeDeliveryPlan) validate() error {
	if p.workset.manifest.validate() != nil || p.workset.dir == "" || p.write.validate() != nil || p.hash == "" || p.hash != changeDeliveryPlanHash(p.workset, p.write) || !worksetMatchesControl(p.workset, p.write.Control()) || p.write.Mode() != synccontract.ModeIncrementalUpsert || p.write.Strategy() != connectors.ApplyStrategyMerge || int64(p.write.RecordCount()) != p.workset.Changes() || int64(p.write.TombstoneCount()) != p.workset.TombstoneCount() {
		return ErrChangeDeliveryPlanInvalid
	}
	keys, err := mappedChangeDeliveryKeys(p.workset.manifest.Keys, p.write.Mapping())
	if err != nil || len(keys) != len(p.write.Keys()) {
		return ErrChangeDeliveryPlanInvalid
	}
	for index, key := range keys {
		if p.write.Keys()[index] != key {
			return ErrChangeDeliveryPlanInvalid
		}
	}
	return nil
}

func (p ChangeDeliveryPlan) matches(other ChangeDeliveryPlan) bool {
	return p.validate() == nil && other.validate() == nil && p.workset.Identity() == other.workset.Identity() && p.workset.ContentSHA256() == other.workset.ContentSHA256() && p.write.matches(other.write)
}

// WorksetIdentity returns the sealed immutable workset address. It is safe to
// use as replay/audit evidence, but it cannot expose the workset filesystem
// location or candidate baseline data.
func (p ChangeDeliveryPlan) WorksetIdentity() string { return p.workset.Identity() }

// PlanHash returns the opaque digest binding the workset, asserted owner,
// target database/relation identity, schema, mapped ordered keys, and bounded
// effects. It changes when a relation OID, schema, mapping, key order, or
// workset changes, and can therefore be safely used in preview/audit output.
func (p ChangeDeliveryPlan) PlanHash() string { return p.hash }

// RecordCount returns the exact sealed delta count that the only permitted
// target session must apply.
func (p ChangeDeliveryPlan) RecordCount() int { return p.write.RecordCount() }

// TombstoneCount returns the exact sealed count of explicit delete events.
// It never includes physically absent rows.
func (p ChangeDeliveryPlan) TombstoneCount() int { return p.write.TombstoneCount() }

func changeDeliveryPlanHash(workset ChangeDeliveryWorkset, write DatabaseWritePlan) string {
	control := write.Control()
	owner := control.Owner().Identity()
	target := control.TargetDatabase()
	native := control.NativeIdentity()
	schema := control.Schema()
	components := []string{
		workset.Identity(),
		workset.ContentSHA256(),
		owner.WorkspaceID,
		owner.ConnectorID,
		owner.ConnectionID,
		target.Kind(),
		target.Value(),
		control.Target().StreamID(),
		control.Target().Namespace(),
		control.Target().Relation(),
		native.Kind,
		native.Value,
		strconv.FormatUint(uint64(schema.Version()), 10),
		schema.Fingerprint().String(),
		string(write.Mode()),
		string(write.Strategy()),
		strconv.Itoa(write.RecordCount()),
		strconv.Itoa(write.TombstoneCount()),
		strconv.Itoa(write.BatchSize()),
	}
	components = append(components, write.Keys()...)
	for _, column := range write.Mapping().Columns() {
		components = append(components, column.Source, column.Target, strconv.FormatBool(column.Nullable), changeDeliveryLogicalTypeHashComponent(column.Type.Source()), changeDeliveryLogicalTypeHashComponent(column.Type.Target()))
	}
	return changeDeliveryHash("polymetrics-change-delivery-plan-v1", components...)
}

func changeDeliveryLogicalTypeHashComponent(logical LogicalType) string {
	components := []string{
		string(logical.Kind()),
		strconv.FormatUint(uint64(logical.BitWidth()), 10),
		strconv.FormatUint(uint64(logical.Precision()), 10),
		strconv.FormatUint(uint64(logical.Scale()), 10),
		strconv.FormatUint(uint64(logical.MaxLength()), 10),
		logical.Collation(),
		strconv.FormatBool(logical.WithTimezone()),
	}
	engine, name, options := logical.OpaqueNativeDetails()
	components = append(components, engine, name)
	components = append(components, options...)
	if element := logical.Element(); element != nil {
		components = append(components, changeDeliveryLogicalTypeHashComponent(*element))
	}
	return changeDeliveryHash("polymetrics-change-delivery-logical-type-v1", components...)
}

// ChangeDeliveryBaselineStore persists a candidate baseline only after a
// ledger-backed DeliveryReceiptV1 is available. The narrow port accepts sealed
// values only: callers cannot supply an arbitrary path, key, or receipt.
type ChangeDeliveryBaselineStore interface {
	RecordChangeDeliveryBaseline(context.Context, ManagedTargetDeliveryLedgerKey, ChangeDeliveryWorkset, DeliveryReceiptV1) error
}

// ChangeDeliveryPreview is native preview evidence for one complete immutable
// workset delivery plan. It deliberately exposes no relation, SQL, or target
// connection details.
type ChangeDeliveryPreview struct {
	plan    ChangeDeliveryPlan
	preview DatabaseWritePreview
}

func (p ChangeDeliveryPreview) validate() error {
	if p.plan.validate() != nil || !p.preview.matches(p.plan.write) {
		return ErrChangeDeliveryApprovalInvalid
	}
	return nil
}

type changeDeliveryApprovalUse struct {
	consumed atomic.Bool
}

// ChangeDeliveryApproval is one-shot approval for both native preview evidence
// and the exact immutable workset. DatabaseWriteApproval alone cannot bind a
// workset because two worksets can have the same target and record counts.
type ChangeDeliveryApproval struct {
	plan  ChangeDeliveryPlan
	write *DatabaseWriteApproval
	use   *changeDeliveryApprovalUse
}

// NewChangeDeliveryApproval creates an approval from one successful immutable
// workset preview. It has no reset operation and no caller-controlled token.
func NewChangeDeliveryApproval(preview ChangeDeliveryPreview) (*ChangeDeliveryApproval, error) {
	if preview.validate() != nil {
		return nil, ErrChangeDeliveryApprovalInvalid
	}
	write, err := NewDatabaseWriteApproval(preview.preview)
	if err != nil {
		return nil, ErrChangeDeliveryApprovalInvalid
	}
	return &ChangeDeliveryApproval{plan: preview.plan, write: write, use: &changeDeliveryApprovalUse{}}, nil
}

func (a *ChangeDeliveryApproval) consume(plan ChangeDeliveryPlan) error {
	if a == nil || a.use == nil || a.write == nil || !a.plan.matches(plan) {
		return ErrChangeDeliveryApprovalInvalid
	}
	if !a.use.consumed.CompareAndSwap(false, true) {
		return ErrChangeDeliveryApprovalConsumed
	}
	return nil
}

// Consumed reports whether this workset delivery approval was admitted. It is
// observability only and cannot reset the one-shot authority.
func (a *ChangeDeliveryApproval) Consumed() bool {
	return a != nil && a.use != nil && a.use.consumed.Load()
}

// ChangeDeliveryExecutor owns the one-engine bridge from an immutable workset
// to the shared database write executor and receipt-bound baseline store.
type ChangeDeliveryExecutor struct {
	write     *DatabaseWriteExecutor
	baselines ChangeDeliveryBaselineStore
}

// NewChangeDeliveryExecutor creates a controller around the existing shared
// write executor and a durable per-destination baseline store.
func NewChangeDeliveryExecutor(write *DatabaseWriteExecutor, baselines ChangeDeliveryBaselineStore) (*ChangeDeliveryExecutor, error) {
	if write == nil || isNilInterface(baselines) {
		return nil, ErrChangeDeliveryPlanInvalid
	}
	return &ChangeDeliveryExecutor{write: write, baselines: baselines}, nil
}

// Preview obtains native evidence for the complete immutable workset plan.
// A changed workset must obtain a distinct preview and approval.
func (e *ChangeDeliveryExecutor) Preview(ctx context.Context, plan ChangeDeliveryPlan) (ChangeDeliveryPreview, error) {
	if ctx == nil || plan.validate() != nil || e == nil || e.write == nil || isNilInterface(e.baselines) {
		return ChangeDeliveryPreview{}, ErrChangeDeliveryPlanInvalid
	}
	preview, err := e.write.Preview(ctx, plan.write)
	if err != nil {
		return ChangeDeliveryPreview{}, err
	}
	return ChangeDeliveryPreview{plan: plan, preview: preview}, nil
}

// ChangeDeliveryResult is receipt/baseline outcome evidence. It does not
// expose an acknowledgement until both the target ledger and the matching
// candidate baseline have persisted.
type ChangeDeliveryResult struct {
	plan     ChangeDeliveryPlan
	write    DatabaseWriteResult
	baseline bool
}

// Outcome returns the known target transaction outcome.
func (r ChangeDeliveryResult) Outcome() CommitOutcome { return r.write.Outcome() }

// WorksetIdentity returns the immutable artifact callers must retain when a
// replay is required.
func (r ChangeDeliveryResult) WorksetIdentity() string { return r.plan.WorksetIdentity() }

// Receipt returns target durability evidence only after the candidate baseline
// has durably recorded against the same destination ledger key.
func (r ChangeDeliveryResult) Receipt() (DeliveryReceiptV1, bool) {
	if !r.baseline {
		return DeliveryReceiptV1{}, false
	}
	return r.write.Receipt()
}

// DownstreamAcknowledgement remains unavailable while baseline persistence is
// incomplete, even if the target commit itself succeeded.
func (r ChangeDeliveryResult) DownstreamAcknowledgement() (synccontract.DownstreamAcknowledgement, error) {
	if !r.baseline {
		return synccontract.DownstreamAcknowledgement{}, ErrChangeDeliveryBaselineUnavailable
	}
	return r.write.DownstreamAcknowledgement()
}

// Execute consumes one exact workset approval, materializes only its sealed
// keyed delta and explicit tombstones, and promotes the candidate baseline only
// after the shared executor confirms commit and durable target-ledger receipt.
func (e *ChangeDeliveryExecutor) Execute(ctx context.Context, plan ChangeDeliveryPlan, approval *ChangeDeliveryApproval) (ChangeDeliveryResult, error) {
	result := ChangeDeliveryResult{plan: plan}
	if ctx == nil || plan.validate() != nil || e == nil || e.write == nil || isNilInterface(e.baselines) {
		return result, ErrChangeDeliveryPlanInvalid
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	input, workset, err := changeDeliveryWriteInput(ctx, plan)
	if err != nil {
		return result, err
	}
	if err := approval.consume(plan); err != nil {
		return result, err
	}
	writeResult, err := e.write.ExecuteInput(ctx, plan.write, approval.write, input)
	result.write = writeResult
	if err != nil {
		if errors.Is(err, ErrDatabaseWriteCommitOutcomeUnknown) {
			return result, fmt.Errorf("%w: %w", ErrChangeDeliveryReplayRequired, err)
		}
		return result, err
	}
	receipt, ok := writeResult.Receipt()
	if !ok {
		return result, ErrChangeDeliveryBaselineUnavailable
	}
	key, err := NewManagedTargetDeliveryLedgerKey(plan.write.Control())
	if err != nil {
		return result, ErrChangeDeliveryBaselineUnavailable
	}
	// Once the target commit and ledger are durable, a caller cancellation must
	// not strand the successful candidate baseline. The store still reports its
	// own failure explicitly and no acknowledgement is then issued.
	if err := e.baselines.RecordChangeDeliveryBaseline(context.WithoutCancel(ctx), key, workset, receipt); err != nil {
		return result, ErrChangeDeliveryBaselineUnavailable
	}
	result.baseline = true
	return result, nil
}

func changeDeliveryWriteInput(ctx context.Context, plan ChangeDeliveryPlan) (DatabaseWriteInput, ChangeDeliveryWorkset, error) {
	workset, err := openChangeDeliveryWorkset(ctx, plan.workset.dir, plan.workset.Identity(), plan.workset.manifest.MaxArtifactBytes)
	if err != nil {
		return DatabaseWriteInput{}, ChangeDeliveryWorkset{}, ErrChangeDeliveryPlanInvalid
	}
	if workset.Identity() != plan.workset.Identity() || workset.ContentSHA256() != plan.workset.ContentSHA256() || !worksetMatchesControl(workset, plan.write.Control()) {
		return DatabaseWriteInput{}, ChangeDeliveryWorkset{}, ErrChangeDeliveryPlanInvalid
	}
	records := make([]connectors.Record, 0, workset.Changes())
	if err := workset.ReadDelta(ctx, func(row warehouse.Row) error {
		record := make(connectors.Record, len(row))
		for key, value := range row {
			record[key] = value
		}
		records = append(records, record)
		return nil
	}); err != nil {
		return DatabaseWriteInput{}, ChangeDeliveryWorkset{}, err
	}
	tombstones, err := workset.Tombstones(ctx)
	if err != nil {
		return DatabaseWriteInput{}, ChangeDeliveryWorkset{}, err
	}
	tombstones, err = mapChangeDeliveryTombstones(tombstones, workset.manifest.Keys, plan.write.Mapping())
	if err != nil {
		return DatabaseWriteInput{}, ChangeDeliveryWorkset{}, ErrChangeDeliveryPlanInvalid
	}
	if int64(len(records)) != workset.Changes() || int64(len(tombstones)) != workset.TombstoneCount() {
		return DatabaseWriteInput{}, ChangeDeliveryWorkset{}, ErrChangeDeliveryPlanInvalid
	}
	envelope, err := NewTombstoneEnvelope(tombstones)
	if err != nil {
		return DatabaseWriteInput{}, ChangeDeliveryWorkset{}, ErrChangeDeliveryPlanInvalid
	}
	input, err := NewDatabaseWriteInput(records, envelope)
	if err != nil {
		return DatabaseWriteInput{}, ChangeDeliveryWorkset{}, ErrChangeDeliveryPlanInvalid
	}
	if input.validateFor(plan.write) != nil {
		return DatabaseWriteInput{}, ChangeDeliveryWorkset{}, ErrChangeDeliveryPlanInvalid
	}
	return input, workset, nil
}

// mapChangeDeliveryTombstones changes only the sealed key vocabulary from the
// workset's source fields to MappingContractV1's target fields. It neither
// adds deletes nor infers them from absent rows; every returned tombstone is a
// clone of an explicit workset event.
func mapChangeDeliveryTombstones(tombstones []synccontract.Tombstone, sourceKeys []string, mapping MappingContractV1) ([]synccontract.Tombstone, error) {
	if len(tombstones) == 0 {
		return nil, nil
	}
	if len(sourceKeys) == 0 || mapping.validate() != nil {
		return nil, ErrChangeDeliveryPlanInvalid
	}
	targetBySource := make(map[string]string, len(mapping.Columns()))
	for _, column := range mapping.Columns() {
		targetBySource[column.Source] = column.Target
	}
	mapped := make([]synccontract.Tombstone, len(tombstones))
	for index, tombstone := range tombstones {
		if tombstone.Operation != synccontract.OperationDelete || tombstone.Validate() != nil {
			return nil, ErrChangeDeliveryPlanInvalid
		}
		var sourceValues map[string]json.RawMessage
		if err := json.Unmarshal(tombstone.Key, &sourceValues); err != nil || len(sourceValues) != len(sourceKeys) {
			return nil, ErrChangeDeliveryPlanInvalid
		}
		targetValues := make(map[string]json.RawMessage, len(sourceKeys))
		for _, sourceKey := range sourceKeys {
			value, found := sourceValues[sourceKey]
			targetKey, mappedKey := targetBySource[sourceKey]
			if !found || !mappedKey || targetKey == "" {
				return nil, ErrChangeDeliveryPlanInvalid
			}
			targetValues[targetKey] = append(json.RawMessage(nil), value...)
		}
		encoded, err := json.Marshal(targetValues)
		if err != nil {
			return nil, ErrChangeDeliveryPlanInvalid
		}
		mapped[index] = tombstone.Clone()
		mapped[index].Key = encoded
	}
	return mapped, nil
}
