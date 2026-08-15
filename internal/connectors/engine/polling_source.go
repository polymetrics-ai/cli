package engine

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

const (
	pollingSourceCheckpointMechanism = "polling_watermark"
	pollingSourceProtocolVersion     = "polling-source-v1"
)

// PollingSourceRunner is the I/O half of a source that has already passed
// PollingPreflight. Its request is intentionally catalog-bound and keyset-only:
// neither definitions nor callers can supply a query, URL, method, or command.
// The native runner validates cursor precision and strict ordering because the
// opaque tuple must not be parsed or ordered by the connector-neutral engine.
type PollingSourceRunner interface {
	PollingPreflightSourceExecutor
	PollingSourceRuntimeState(context.Context, connectors.PollingCatalogObject) (PollingSourceRuntimeState, error)
	FetchPollingSourcePage(context.Context, PollingSourcePageRequest) (PollingSourcePage, error)
	ValidatePollingSourcePageTraversal(context.Context, *synccontract.CheckpointPosition, PollingSourcePage) error
}

// PollingSourceRuntimeState is the provider state copied into every candidate
// checkpoint. The registered native runner obtains it from its own closed
// protocol; the engine never accepts it from a user invocation.
type PollingSourceRuntimeState struct {
	SourceGeneration synccontract.OpaqueToken
	SchemaVersion    string
	SnapshotBarrier  synccontract.SnapshotBarrier
	Partitions       []synccontract.PartitionState
	Dedupe           synccontract.DedupeIdentity
	DedupeWindow     synccontract.DedupeWindow
}

// Clone returns an independent runtime-state copy without changing opaque
// tokens into display strings or scalar cursors.
func (s PollingSourceRuntimeState) Clone() PollingSourceRuntimeState {
	clone := s
	clone.SourceGeneration = append(synccontract.OpaqueToken(nil), s.SourceGeneration...)
	clone.SnapshotBarrier.Token = append(synccontract.OpaqueToken(nil), s.SnapshotBarrier.Token...)
	if s.Partitions != nil {
		clone.Partitions = make([]synccontract.PartitionState, len(s.Partitions))
		for index := range s.Partitions {
			clone.Partitions[index] = s.Partitions[index].Clone()
		}
	}
	clone.Dedupe = s.Dedupe.Clone()
	clone.DedupeWindow = s.DedupeWindow.Clone()
	return clone
}

// PollingSourceRequestBudget is consumed by the registered runner before each
// physical provider request. A runner cannot turn one sync invocation into an
// unbounded scan just because its database driver follows continuation work.
type PollingSourceRequestBudget struct {
	mu        sync.Mutex
	remaining int
	used      int
}

func newPollingSourceRequestBudget(limit int) *PollingSourceRequestBudget {
	return &PollingSourceRequestBudget{remaining: limit}
}

// Consume reserves one permitted provider request.
func (b *PollingSourceRequestBudget) Consume(ctx context.Context) error {
	if b == nil {
		return fmt.Errorf("polling source request budget is required")
	}
	if ctx == nil {
		return fmt.Errorf("polling source request context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.remaining <= 0 {
		return fmt.Errorf("polling source request budget exhausted")
	}
	b.remaining--
	b.used++
	return nil
}

func (b *PollingSourceRequestBudget) usedRequests() int {
	if b == nil {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.used
}

// PollingSourcePageRequest is the only native read input exposed after
// preflight. After is the exact last *durably committed* tuple, never a tuple
// from a page that has merely been fetched or attempted downstream.
type PollingSourcePageRequest struct {
	Object          connectors.PollingCatalogObject
	After           *synccontract.CheckpointPosition
	PageSize        int
	SnapshotBarrier synccontract.SnapshotBarrier
	RequestBudget   *PollingSourceRequestBudget
}

func (r PollingSourcePageRequest) clone() PollingSourcePageRequest {
	clone := r
	clone.Object = clonePollingCatalogObject(r.Object)
	if r.After != nil {
		after := r.After.Clone()
		clone.After = &after
	}
	clone.SnapshotBarrier.Token = append(synccontract.OpaqueToken(nil), r.SnapshotBarrier.Token...)
	return clone
}

// PollingSourceItem is one page member. A member is exactly one record or one
// soft-delete tombstone, and its position is the complete watermark/tie tuple
// the next page must resume after.
type PollingSourceItem struct {
	Record    connectors.Record
	Tombstone *synccontract.Tombstone
	Position  synccontract.CheckpointPosition
}

func (i PollingSourceItem) Clone() PollingSourceItem {
	clone := i
	clone.Position = i.Position.Clone()
	if i.Tombstone != nil {
		tombstone := i.Tombstone.Clone()
		clone.Tombstone = &tombstone
	}
	return clone
}

// PollingSourcePage is a single bounded keyset page. More is a native-runner
// assertion that another page exists behind this page's last tuple.
type PollingSourcePage struct {
	Items      []PollingSourceItem
	More       bool
	ObservedAt time.Time
}

func (p PollingSourcePage) Clone() PollingSourcePage {
	clone := p
	if p.Items != nil {
		clone.Items = make([]PollingSourceItem, len(p.Items))
		for index := range p.Items {
			clone.Items[index] = p.Items[index].Clone()
		}
	}
	return clone
}

// PollingSourceExecutor converts one preflight-admitted native polling runner
// into the shared transport source role. It owns page sequencing only; the
// transport orchestrator owns durable destination acknowledgement and storage.
type PollingSourceExecutor struct {
	mode        synccontract.Mode
	declaration *connectors.PollingWatermarkDescriptor
	object      connectors.PollingCatalogObject
	runner      PollingSourceRunner
}

var _ synctransport.SourceExecutor = (*PollingSourceExecutor)(nil)

// NewPollingSourceExecutor accepts only a successful PollingPreflight result.
// It deliberately type-asserts the registered preflight executor instead of
// repeating preflight's definition, registration, or corpus-admission rules.
func NewPollingSourceExecutor(resolved ResolvedPollingWatermark) (*PollingSourceExecutor, error) {
	if resolved.Declaration == nil {
		return nil, fmt.Errorf("resolved polling watermark declaration is required")
	}
	if err := resolved.Mode.Validate(); err != nil {
		return nil, err
	}
	runner, ok := resolved.Source.(PollingSourceRunner)
	if !ok || isNilPollingPreflightExecutor(runner) {
		return nil, fmt.Errorf("resolved polling source executor %q does not implement page-safe polling reads", resolved.Declaration.Source.Executor.ID)
	}
	if runner.PollingSourceExecutorReference() != resolved.Declaration.Source.Executor {
		return nil, fmt.Errorf("resolved polling source executor does not match the declaration")
	}
	return &PollingSourceExecutor{
		mode:        resolved.Mode,
		declaration: resolved.Declaration.Clone(),
		object:      clonePollingCatalogObject(resolved.Object),
		runner:      runner,
	}, nil
}

// TransportExecutorReference exposes the exact registered native source.
func (e *PollingSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	if e == nil || e.declaration == nil {
		return connectors.TransportExecutorReference{}
	}
	return e.declaration.Source.Executor
}

// ReadTransport fetches, emits, and only then advances the in-memory keyset
// tuple. emit returns only after the transport path has staged, durably
// applied, read back, and persisted the candidate checkpoint; an error leaves
// the next request on the prior committed tuple for safe replay.
func (e *PollingSourceExecutor) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	if e == nil || e.declaration == nil || isNilPollingPreflightExecutor(e.runner) {
		return fmt.Errorf("polling source executor is required")
	}
	if ctx == nil {
		return fmt.Errorf("polling source context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if emit == nil {
		return fmt.Errorf("polling source page emitter is required")
	}
	if request.Mode != e.mode {
		return fmt.Errorf("polling source request mode %q does not match preflight mode %q", request.Mode, e.mode)
	}
	pageSize, err := e.effectivePageSize(request.BatchSize)
	if err != nil {
		return err
	}

	state, err := e.runner.PollingSourceRuntimeState(ctx, clonePollingCatalogObject(e.object))
	if err != nil {
		return fmt.Errorf("read polling source runtime state: %w", err)
	}
	state = state.Clone()
	expected := e.sourceExpectation(state)
	if err := validatePollingSourceRuntimeState(e.declaration, expected.Source, state); err != nil {
		return err
	}
	if err := validatePollingSourceRequestResume(request.Resume, expected); err != nil {
		return err
	}

	var after *synccontract.CheckpointPosition
	if request.Checkpoint != nil {
		checkpoint := request.Checkpoint.Clone()
		if err := e.validateResume(checkpoint, expected, state); err != nil {
			return err
		}
		position := checkpoint.Position.Clone()
		after = &position
	}

	budget := newPollingSourceRequestBudget(e.declaration.Source.Read.MaxRequests)
	for pageNumber := 0; ; pageNumber++ {
		if pageNumber >= e.declaration.Source.Read.MaxPages {
			return fmt.Errorf("polling source reached declared maximum of %d pages before completion", e.declaration.Source.Read.MaxPages)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		usedBefore := budget.usedRequests()
		page, err := e.runner.FetchPollingSourcePage(ctx, PollingSourcePageRequest{
			Object:          clonePollingCatalogObject(e.object),
			After:           clonePollingSourcePosition(after),
			PageSize:        pageSize,
			SnapshotBarrier: clonePollingSnapshotBarrier(state.SnapshotBarrier),
			RequestBudget:   budget,
		}.clone())
		if err != nil {
			return fmt.Errorf("fetch polling source page: %w", err)
		}
		if budget.usedRequests() == usedBefore {
			return fmt.Errorf("polling source runner returned without consuming a bounded provider request")
		}
		if err := e.runner.ValidatePollingSourcePageTraversal(ctx, clonePollingSourcePosition(after), page.Clone()); err != nil {
			return fmt.Errorf("validate polling source page traversal: %w", err)
		}
		if err := e.validatePage(page, after, pageSize); err != nil {
			return err
		}
		if len(page.Items) == 0 {
			return nil
		}

		candidate, records, tombstones, err := e.candidateForPage(page, state)
		if err != nil {
			return err
		}
		if err := emit(synctransport.SourcePage{
			Records:             records,
			Tombstones:          tombstones,
			CandidateCheckpoint: candidate,
		}); err != nil {
			return fmt.Errorf("deliver polling source page: %w", err)
		}

		// This assignment is deliberately after emit. A failed stage, apply,
		// read-back, acknowledgement, or checkpoint save leaves after on the
		// prior durable tuple and makes the page replayable.
		next := candidate.Position.Clone()
		after = &next
		if !page.More {
			return nil
		}
	}
}

func (e *PollingSourceExecutor) effectivePageSize(batchSize int) (int, error) {
	if batchSize <= 0 {
		return 0, fmt.Errorf("polling source batch size must be positive")
	}
	limit := batchSize
	if e.declaration.Source.Read.MaxPageSize < limit {
		limit = e.declaration.Source.Read.MaxPageSize
	}
	if e.declaration.Target.MaxBatchRecords < limit {
		limit = e.declaration.Target.MaxBatchRecords
	}
	if limit <= 0 {
		return 0, fmt.Errorf("polling source resolved page size must be positive")
	}
	return limit, nil
}

func (e *PollingSourceExecutor) sourceExpectation(state PollingSourceRuntimeState) synccontract.ResumeExpectation {
	identity := e.declaration.Source.Identity
	return synccontract.ResumeExpectation{
		Source: synccontract.SourceIdentity{
			Engine:           identity.Engine,
			AccountOrCluster: identity.AccountScope,
			ObjectScope:      identity.ObjectScope,
		},
		SourceGeneration: append(synccontract.OpaqueToken(nil), state.SourceGeneration...),
	}
}

func validatePollingSourceRuntimeState(declaration *connectors.PollingWatermarkDescriptor, source synccontract.SourceIdentity, state PollingSourceRuntimeState) error {
	if len(state.SourceGeneration) == 0 {
		return fmt.Errorf("polling source runtime generation is required")
	}
	if state.SnapshotBarrier.Kind != string(declaration.Source.Snapshot.Kind) {
		return fmt.Errorf("polling source runtime snapshot barrier does not match the declaration")
	}
	observed := true
	probe := synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           source,
		Mechanism:        pollingSourceCheckpointMechanism,
		SnapshotBarrier:  pollingSourceBarrierPointer(state.SnapshotBarrier),
		Position:         synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("state-validation"), TieBreaker: synccontract.OpaqueToken("state-validation")},
		PositionObserved: &observed,
		Partitions:       clonePollingPartitions(state.Partitions),
		SourceGeneration: append(synccontract.OpaqueToken(nil), state.SourceGeneration...),
		SchemaVersion:    state.SchemaVersion,
		ProtocolVersion:  pollingSourceProtocolVersion,
		Dedupe:           state.Dedupe.Clone(),
		DedupeWindow:     state.DedupeWindow.Clone(),
		ObservedAt:       time.Unix(1, 0).UTC(),
	}
	if err := probe.Validate(); err != nil {
		return fmt.Errorf("polling source runtime state: %w", err)
	}
	return nil
}

func validatePollingSourceRequestResume(actual, expected synccontract.ResumeExpectation) error {
	if actual.Source != expected.Source {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeSourceIdentityIncompatible, "polling source request resume expectation does not match preflight source identity")
	}
	if !bytes.Equal(actual.SourceGeneration, expected.SourceGeneration) {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeSourceGenerationChanged, "polling source request generation does not match the current native source")
	}
	return nil
}

func (e *PollingSourceExecutor) validateResume(checkpoint synccontract.CheckpointEnvelope, expected synccontract.ResumeExpectation, state PollingSourceRuntimeState) error {
	if err := checkpoint.ValidateResume(expected); err != nil {
		return err
	}
	if checkpoint.Mechanism != pollingSourceCheckpointMechanism {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "checkpoint mechanism is not polling_watermark")
	}
	if checkpoint.ProtocolVersion != pollingSourceProtocolVersion {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "checkpoint protocol version is not resumable by this polling source")
	}
	if checkpoint.SchemaVersion != state.SchemaVersion {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "checkpoint schema fingerprint no longer matches")
	}
	if checkpoint.SnapshotBarrier == nil || !equalPollingSnapshotBarrier(*checkpoint.SnapshotBarrier, state.SnapshotBarrier) {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "checkpoint snapshot barrier no longer matches")
	}
	if err := e.validatePosition(checkpoint.Position); err != nil {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "checkpoint cursor tuple is invalid")
	}
	return nil
}

func (e *PollingSourceExecutor) validatePage(page PollingSourcePage, after *synccontract.CheckpointPosition, pageSize int) error {
	if len(page.Items) > pageSize {
		return fmt.Errorf("polling source returned %d records, exceeding bounded page size %d", len(page.Items), pageSize)
	}
	if len(page.Items) == 0 {
		if page.More {
			return fmt.Errorf("polling source returned an empty page while claiming more rows")
		}
		return nil
	}
	if page.ObservedAt.IsZero() {
		return fmt.Errorf("polling source page observation time is required")
	}
	seen := make([]synccontract.CheckpointPosition, 0, len(page.Items))
	previous := clonePollingSourcePosition(after)
	for index, item := range page.Items {
		if (item.Record == nil) == (item.Tombstone == nil) {
			return fmt.Errorf("polling source page item %d must contain exactly one record or tombstone", index)
		}
		if err := e.validatePosition(item.Position); err != nil {
			return fmt.Errorf("polling source page item %d position: %w", index, err)
		}
		if item.Tombstone != nil {
			if e.declaration.Source.DeleteVisibility != connectors.PollingDeleteVisibilityTombstone {
				return fmt.Errorf("polling source emitted a tombstone although deletes are declared hard-delete-invisible")
			}
			if err := item.Tombstone.Validate(); err != nil {
				return fmt.Errorf("polling source page item %d tombstone: %w", index, err)
			}
			if !equalPollingPosition(item.Tombstone.Position, item.Position) {
				return fmt.Errorf("polling source page item %d tombstone position does not match its page tuple", index)
			}
		}
		if containsPollingPosition(seen, item.Position) || (previous != nil && equalPollingPosition(*previous, item.Position)) {
			return fmt.Errorf("polling source page item %d does not advance the complete keyset tuple", index)
		}
		seen = append(seen, item.Position.Clone())
		position := item.Position.Clone()
		previous = &position
	}
	return nil
}

func (e *PollingSourceExecutor) validatePosition(position synccontract.CheckpointPosition) error {
	if len(position.Primary) == 0 || len(position.TieBreaker) == 0 {
		return fmt.Errorf("watermark and unique tie-breaker checkpoint tokens are required")
	}
	return nil
}

func (e *PollingSourceExecutor) candidateForPage(page PollingSourcePage, state PollingSourceRuntimeState) (synccontract.CheckpointEnvelope, []connectors.Record, []synccontract.Tombstone, error) {
	last := page.Items[len(page.Items)-1].Position.Clone()
	observed := true
	identity := e.sourceExpectation(state).Source
	candidate := synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           identity,
		Mechanism:        pollingSourceCheckpointMechanism,
		SnapshotBarrier:  pollingSourceBarrierPointer(state.SnapshotBarrier),
		Position:         last,
		PositionObserved: &observed,
		Partitions:       clonePollingPartitions(state.Partitions),
		SourceGeneration: append(synccontract.OpaqueToken(nil), state.SourceGeneration...),
		SchemaVersion:    state.SchemaVersion,
		ProtocolVersion:  pollingSourceProtocolVersion,
		Dedupe:           state.Dedupe.Clone(),
		DedupeWindow:     state.DedupeWindow.Clone(),
		ObservedAt:       page.ObservedAt,
	}
	if err := candidate.Validate(); err != nil {
		return synccontract.CheckpointEnvelope{}, nil, nil, fmt.Errorf("polling source candidate checkpoint: %w", err)
	}
	records := make([]connectors.Record, 0, len(page.Items))
	tombstones := make([]synccontract.Tombstone, 0, len(page.Items))
	for _, item := range page.Items {
		if item.Record != nil {
			records = append(records, item.Record)
			continue
		}
		tombstones = append(tombstones, item.Tombstone.Clone())
	}
	return candidate, records, tombstones, nil
}

func clonePollingSourcePosition(position *synccontract.CheckpointPosition) *synccontract.CheckpointPosition {
	if position == nil {
		return nil
	}
	clone := position.Clone()
	return &clone
}

func clonePollingSnapshotBarrier(barrier synccontract.SnapshotBarrier) synccontract.SnapshotBarrier {
	barrier.Token = append(synccontract.OpaqueToken(nil), barrier.Token...)
	return barrier
}

func pollingSourceBarrierPointer(barrier synccontract.SnapshotBarrier) *synccontract.SnapshotBarrier {
	clone := clonePollingSnapshotBarrier(barrier)
	return &clone
}

func clonePollingPartitions(partitions []synccontract.PartitionState) []synccontract.PartitionState {
	if partitions == nil {
		return nil
	}
	clone := make([]synccontract.PartitionState, len(partitions))
	for index := range partitions {
		clone[index] = partitions[index].Clone()
	}
	return clone
}

func equalPollingPosition(left, right synccontract.CheckpointPosition) bool {
	return bytes.Equal(left.Primary, right.Primary) && bytes.Equal(left.TieBreaker, right.TieBreaker)
}

func containsPollingPosition(positions []synccontract.CheckpointPosition, candidate synccontract.CheckpointPosition) bool {
	for _, position := range positions {
		if equalPollingPosition(position, candidate) {
			return true
		}
	}
	return false
}

func equalPollingSnapshotBarrier(left, right synccontract.SnapshotBarrier) bool {
	return left.Kind == right.Kind && bytes.Equal(left.Token, right.Token)
}
