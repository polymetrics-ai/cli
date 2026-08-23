package synccontract

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// StateVersion is the current persisted checkpoint envelope version.
const StateVersion uint = 1

// OpaqueToken is provider-owned checkpoint data. It must only be copied or
// compared as bytes; callers must not parse, normalize, or reconstruct it.
// encoding/json represents a []byte as base64, which preserves every byte.
type OpaqueToken []byte

func cloneToken(token OpaqueToken) OpaqueToken {
	return append(OpaqueToken(nil), token...)
}

// SourceIdentity binds a state envelope to exactly one source scope.
type SourceIdentity struct {
	Engine           string `json:"engine"`
	AccountOrCluster string `json:"account_or_cluster"`
	ObjectScope      string `json:"object_scope"`
}

// Validate rejects a source identity that could match more than one source
// scope during checkpoint recovery.
func (s SourceIdentity) Validate() error {
	if strings.TrimSpace(s.Engine) == "" {
		return fmt.Errorf("source identity engine is required")
	}
	if strings.TrimSpace(s.AccountOrCluster) == "" {
		return fmt.Errorf("source identity account_or_cluster is required")
	}
	if strings.TrimSpace(s.ObjectScope) == "" {
		return fmt.Errorf("source identity object_scope is required")
	}
	return nil
}

func (s SourceIdentity) equal(other SourceIdentity) bool {
	return s.Engine == other.Engine &&
		s.AccountOrCluster == other.AccountOrCluster &&
		s.ObjectScope == other.ObjectScope
}

// SnapshotBarrier identifies the consistent point at which a snapshot joins
// an incremental/change stream. Its token is opaque provider data.
type SnapshotBarrier struct {
	Kind  string      `json:"kind"`
	Token OpaqueToken `json:"token"`
}

func (b SnapshotBarrier) clone() SnapshotBarrier {
	b.Token = cloneToken(b.Token)
	return b
}

func (b SnapshotBarrier) validate() error {
	if strings.TrimSpace(b.Kind) == "" {
		return fmt.Errorf("snapshot barrier kind is required")
	}
	if len(b.Token) == 0 {
		return fmt.Errorf("snapshot barrier token is required")
	}
	return nil
}

// CheckpointPosition preserves a primary provider position and its tie-breaker
// as separate opaque fields. They must never be joined into a scalar cursor.
type CheckpointPosition struct {
	Primary    OpaqueToken `json:"primary"`
	TieBreaker OpaqueToken `json:"tie_breaker"`
}

// Clone returns an independent copy of both opaque position tokens.
func (p CheckpointPosition) Clone() CheckpointPosition {
	p.Primary = cloneToken(p.Primary)
	p.TieBreaker = cloneToken(p.TieBreaker)
	return p
}

// validate permits an empty global position for a full snapshot. The position
// remains a structured field; mechanism-specific executors decide whether a
// provider has a resumable position at that point.
func (p CheckpointPosition) validate(_ string) error {
	return nil
}

func (p CheckpointPosition) validateOrdered(name string) error {
	if len(p.Primary) == 0 {
		return fmt.Errorf("%s primary checkpoint is required", name)
	}
	if len(p.TieBreaker) == 0 {
		return fmt.Errorf("%s tie-breaker is required", name)
	}
	return nil
}

// SourceContinuation carries an exact engine-produced continuation between
// durably acknowledged source runs. It has no provider-facing interpretation
// at this layer: consumers may copy it to the same source executor only.
type SourceContinuation struct {
	Kind  string      `json:"kind"`
	Token OpaqueToken `json:"token"`
}

// Clone returns an independent continuation or nil for an absent continuation.
func (c *SourceContinuation) Clone() *SourceContinuation {
	if c == nil {
		return nil
	}
	clone := *c
	clone.Token = cloneToken(c.Token)
	return &clone
}

// ContinuationEqual reports whether two optional continuations have the same
// presence, exact kind, and opaque token. Continuation is durable checkpoint
// state, so callers must not treat a changed continuation as an equal source
// position.
func ContinuationEqual(left, right *SourceContinuation) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Kind == right.Kind && bytes.Equal(left.Token, right.Token)
}

func (c SourceContinuation) validate() error {
	if strings.TrimSpace(c.Kind) == "" {
		return fmt.Errorf("source continuation kind is required")
	}
	if len(c.Token) == 0 || len(c.Token) > 4096 {
		return fmt.Errorf("source continuation token is required and must not exceed 4096 bytes")
	}
	return nil
}

// PartitionState preserves an independently resumable partition. It is a
// structured array entry rather than a delimited string or flattened map.
type PartitionState struct {
	Partition OpaqueToken        `json:"partition"`
	Position  CheckpointPosition `json:"position"`
}

// Clone returns an independent copy of a partition and its ordered position.
func (p PartitionState) Clone() PartitionState {
	p.Partition = cloneToken(p.Partition)
	p.Position = p.Position.Clone()
	return p
}

func (p PartitionState) validate() error {
	if len(p.Partition) == 0 {
		return fmt.Errorf("partition identifier is required")
	}
	return p.Position.validateOrdered("partition")
}

// DedupeIdentity records the source identity used to make at-least-once
// delivery idempotent. Its value remains opaque for the provider/executor.
type DedupeIdentity struct {
	Kind  string      `json:"kind"`
	Value OpaqueToken `json:"value,omitempty"`
}

// Clone returns an independent copy of the opaque dedupe identity.
func (d DedupeIdentity) Clone() DedupeIdentity {
	d.Value = cloneToken(d.Value)
	return d
}

func (d DedupeIdentity) validate() error {
	if strings.TrimSpace(d.Kind) == "" {
		return fmt.Errorf("dedupe identity kind is required")
	}
	return nil
}

// DedupeWindow records the explicit replay/overlap window in which the
// DedupeIdentity applies. Both bounds are opaque provider values: an executor
// may compare them only with its native protocol rules and must never flatten
// them into a scalar cursor.
type DedupeWindow struct {
	Kind  string      `json:"kind"`
	Start OpaqueToken `json:"start"`
	End   OpaqueToken `json:"end"`
}

// Clone returns an independent copy of the opaque replay-window bounds.
func (d DedupeWindow) Clone() DedupeWindow {
	d.Start = cloneToken(d.Start)
	d.End = cloneToken(d.End)
	return d
}

func (d DedupeWindow) validate() error {
	if strings.TrimSpace(d.Kind) == "" {
		return fmt.Errorf("dedupe window kind is required")
	}
	if len(d.Start) == 0 || len(d.End) == 0 {
		return fmt.Errorf("dedupe window bounds are required")
	}
	return nil
}

// CheckpointEnvelope is the only durable sync checkpoint shape. ObservedAt
// describes the source observation; CommittedAt is stamped only once a
// downstream durable acknowledgement permits persistence.
type CheckpointEnvelope struct {
	StateVersion    uint               `json:"state_version"`
	Source          SourceIdentity     `json:"source"`
	Mechanism       string             `json:"mechanism"`
	SnapshotBarrier *SnapshotBarrier   `json:"snapshot_barrier"`
	Position        CheckpointPosition `json:"position"`
	// Continuation is optional engine-owned state for a deliberately incomplete
	// bounded source scan. Its presence means this checkpoint is resumable but
	// must never be presented as source exhaustion.
	Continuation     *SourceContinuation `json:"continuation,omitempty"`
	PositionObserved *bool               `json:"position_observed,omitempty"`
	Partitions       []PartitionState    `json:"partitions"`
	SourceGeneration OpaqueToken         `json:"source_generation,omitempty"`
	SchemaVersion    string              `json:"schema_version"`
	ProtocolVersion  string              `json:"protocol_version"`
	Dedupe           DedupeIdentity      `json:"dedupe_identity"`
	DedupeWindow     DedupeWindow        `json:"dedupe_window"`
	ObservedAt       time.Time           `json:"observed_at"`
	CommittedAt      *time.Time          `json:"committed_at,omitempty"`
}

// Clone returns a defensive copy that preserves opaque bytes exactly.
func (c CheckpointEnvelope) Clone() CheckpointEnvelope {
	clone := c
	if c.SnapshotBarrier != nil {
		barrier := c.SnapshotBarrier.clone()
		clone.SnapshotBarrier = &barrier
	}
	clone.Position = c.Position.Clone()
	clone.Continuation = c.Continuation.Clone()
	if c.PositionObserved != nil {
		positionObserved := *c.PositionObserved
		clone.PositionObserved = &positionObserved
	}
	if c.Partitions != nil {
		clone.Partitions = make([]PartitionState, len(c.Partitions))
		for i := range c.Partitions {
			clone.Partitions[i] = c.Partitions[i].Clone()
		}
	}
	clone.SourceGeneration = cloneToken(c.SourceGeneration)
	clone.Dedupe = c.Dedupe.Clone()
	clone.DedupeWindow = c.DedupeWindow.Clone()
	if c.CommittedAt != nil {
		committedAt := *c.CommittedAt
		clone.CommittedAt = &committedAt
	}
	return clone
}

// Validate checks that a version-one envelope is structurally complete. It
// never converts provider data to strings or changes any stored bytes.
func (c CheckpointEnvelope) Validate() error {
	if c.StateVersion != StateVersion {
		return fmt.Errorf("unsupported checkpoint state version %d", c.StateVersion)
	}
	if err := c.Source.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(c.Mechanism) == "" {
		return fmt.Errorf("checkpoint mechanism is required")
	}
	if c.SnapshotBarrier == nil {
		return fmt.Errorf("snapshot barrier is required")
	}
	if err := c.SnapshotBarrier.validate(); err != nil {
		return err
	}
	if err := c.Position.validate("global"); err != nil {
		return err
	}
	if c.Continuation != nil {
		if err := c.Continuation.validate(); err != nil {
			return err
		}
	}
	if c.Partitions == nil {
		return fmt.Errorf("partition state must be an explicit array")
	}
	seenPartitions := make(map[string]struct{}, len(c.Partitions))
	for _, partition := range c.Partitions {
		if err := partition.validate(); err != nil {
			return err
		}
		key := string(partition.Partition)
		if _, exists := seenPartitions[key]; exists {
			return fmt.Errorf("duplicate partition state")
		}
		seenPartitions[key] = struct{}{}
	}
	if strings.TrimSpace(c.SchemaVersion) == "" {
		return fmt.Errorf("schema version is required")
	}
	if strings.TrimSpace(c.ProtocolVersion) == "" {
		return fmt.Errorf("protocol version is required")
	}
	if err := c.Dedupe.validate(); err != nil {
		return err
	}
	if err := c.DedupeWindow.validate(); err != nil {
		return err
	}
	if c.ObservedAt.IsZero() {
		return fmt.Errorf("observed timestamp is required")
	}
	if c.CommittedAt != nil && c.CommittedAt.Before(c.ObservedAt) {
		return fmt.Errorf("committed timestamp cannot precede observed timestamp")
	}
	return nil
}

// ResumeExpectation is the source identity the current executor has verified
// before it attempts to reuse a checkpoint.
type ResumeExpectation struct {
	Source           SourceIdentity
	SourceGeneration OpaqueToken
}

// ValidateResume refuses unsafe resumes without mutating c. An executor uses
// RequireRebootstrap directly for provider-reported retention, slot, and token
// invalidations; this method handles persisted-state compatibility checks.
func (c CheckpointEnvelope) ValidateResume(expected ResumeExpectation) error {
	if c.StateVersion != StateVersion {
		return RequireRebootstrap(RecoveryOutcomeInvalidCheckpoint, "checkpoint state version is not resumable")
	}
	if err := c.Validate(); err != nil {
		return RequireRebootstrap(RecoveryOutcomeInvalidCheckpoint, err.Error())
	}
	if c.CommittedAt == nil {
		return RequireRebootstrap(RecoveryOutcomeInvalidCheckpoint, "checkpoint has not been durably acknowledged")
	}
	if err := expected.Source.Validate(); err != nil {
		return RequireRebootstrap(RecoveryOutcomeInvalidCheckpoint, "resume source identity is incomplete")
	}
	if !c.Source.equal(expected.Source) {
		return RequireRebootstrap(RecoveryOutcomeSourceIdentityIncompatible, "checkpoint source identity does not match the requested stream")
	}
	if len(expected.SourceGeneration) > 0 && !bytes.Equal(c.SourceGeneration, expected.SourceGeneration) {
		return RequireRebootstrap(RecoveryOutcomeSourceGenerationChanged, "checkpoint source generation no longer matches")
	}
	return nil
}
