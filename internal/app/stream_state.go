package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// UnmarshalJSON upgrades the former scalar cursor only into a deliberately
// non-resumable version-zero envelope. It preserves the old bytes for a
// diagnostic/rebootstrap workflow, but no execution path may use it as a new
// scalar cursor or silently replace it with a full scan.
func (s *StreamState) UnmarshalJSON(data []byte) error {
	type streamStateWire struct {
		Connection                         string                           `json:"connection"`
		Stream                             string                           `json:"stream"`
		Checkpoint                         *synccontract.CheckpointEnvelope `json:"checkpoint"`
		CommittedTransportReceipts         []TransportReceiptCommit         `json:"committed_transport_receipts"`
		TransportReceiptAssociationVersion uint                             `json:"transport_receipt_association_version"`
		LegacyCommittedTransportCheckpoint *synccontract.CheckpointEnvelope `json:"legacy_committed_transport_checkpoint"`
		LegacyCursor                       *string                          `json:"cursor"`
		GenerationID                       int64                            `json:"generation_id"`
		ActiveWorkID                       string                           `json:"active_work_id"`
		ActiveWorkFence                    int64                            `json:"active_work_fence"`
		ActiveWorkLeaseUntil               *time.Time                       `json:"active_work_lease_until"`
		LastSuccessfulRunID                string                           `json:"last_successful_run_id"`
		RecordsLoaded                      int                              `json:"records_loaded"`
		UpdatedAt                          time.Time                        `json:"updated_at"`
	}
	var wire streamStateWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.Checkpoint != nil && wire.LegacyCursor != nil {
		return fmt.Errorf("stream state cannot contain both checkpoint envelope and legacy cursor")
	}
	committedReceipts, err := cloneTransportReceiptCommits(wire.CommittedTransportReceipts)
	if err != nil {
		return fmt.Errorf("stream state committed transport receipts: %w", err)
	}
	*s = StreamState{
		Connection:                         wire.Connection,
		Stream:                             wire.Stream,
		CommittedTransportReceipts:         committedReceipts,
		TransportReceiptAssociationVersion: wire.TransportReceiptAssociationVersion,
		GenerationID:                       wire.GenerationID,
		ActiveWorkID:                       wire.ActiveWorkID,
		ActiveWorkFence:                    wire.ActiveWorkFence,
		LastSuccessfulRunID:                wire.LastSuccessfulRunID,
		RecordsLoaded:                      wire.RecordsLoaded,
		UpdatedAt:                          wire.UpdatedAt,
	}
	if wire.Checkpoint != nil {
		checkpoint := wire.Checkpoint.Clone()
		s.Checkpoint = &checkpoint
	}
	if wire.LegacyCommittedTransportCheckpoint != nil {
		checkpoint := wire.LegacyCommittedTransportCheckpoint.Clone()
		if checkpoint.CommittedAt != nil {
			return fmt.Errorf("legacy committed transport checkpoint must not include a commit timestamp")
		}
		s.LegacyCommittedTransportCheckpoint = &checkpoint
	}
	if wire.ActiveWorkLeaseUntil != nil {
		until := wire.ActiveWorkLeaseUntil.UTC()
		s.ActiveWorkLeaseUntil = &until
	}
	if wire.LegacyCursor != nil {
		s.Checkpoint = &synccontract.CheckpointEnvelope{
			StateVersion: 0,
			Position: synccontract.CheckpointPosition{
				Primary: synccontract.OpaqueToken([]byte(*wire.LegacyCursor)),
			},
		}
	}
	return nil
}

func cloneStreamState(state StreamState) StreamState {
	clone := state
	clone.CommittedTransportReceipts = append([]TransportReceiptCommit(nil), state.CommittedTransportReceipts...)
	if state.Checkpoint != nil {
		checkpoint := state.Checkpoint.Clone()
		clone.Checkpoint = &checkpoint
	}
	if state.LegacyCommittedTransportCheckpoint != nil {
		checkpoint := state.LegacyCommittedTransportCheckpoint.Clone()
		clone.LegacyCommittedTransportCheckpoint = &checkpoint
	}
	if state.ActiveWorkLeaseUntil != nil {
		until := state.ActiveWorkLeaseUntil.UTC()
		clone.ActiveWorkLeaseUntil = &until
	}
	return clone
}

const transportReceiptAssociationVersion uint = 1

func migrateLegacyTransportReceiptAssociation(streamState *StreamState) bool {
	if streamState == nil || streamState.TransportReceiptAssociationVersion >= transportReceiptAssociationVersion {
		return false
	}
	streamState.TransportReceiptAssociationVersion = transportReceiptAssociationVersion
	if streamState.LegacyCommittedTransportCheckpoint == nil && len(streamState.CommittedTransportReceipts) == 0 && streamState.Checkpoint != nil && streamState.Checkpoint.CommittedAt != nil {
		checkpoint := streamState.Checkpoint.Clone()
		checkpoint.CommittedAt = nil
		streamState.LegacyCommittedTransportCheckpoint = &checkpoint
	}
	return true
}

func migrateLegacyTransportReceiptAssociations(value *state) bool {
	if value == nil || len(value.StreamStates) == 0 {
		return false
	}
	changed := false
	for key, streamState := range value.StreamStates {
		streamState = cloneStreamState(streamState)
		if !migrateLegacyTransportReceiptAssociation(&streamState) {
			continue
		}
		value.StreamStates[key] = streamState
		changed = true
	}
	return changed
}

func transportReceiptCommitFromWarehouseReceipt(receipt synctransport.WarehouseReceipt) (TransportReceiptCommit, error) {
	if err := receipt.Validate(); err != nil {
		return TransportReceiptCommit{}, err
	}
	return TransportReceiptCommit{
		ReceiptID:        receipt.ID,
		Owner:            receipt.Owner,
		Generation:       receipt.Generation,
		Stream:           receipt.Stream,
		Mode:             receipt.Mode,
		CheckpointSHA256: receipt.CheckpointSHA256,
		TombstonesSHA256: receipt.TombstonesSHA256,
		ManifestSHA256:   receipt.ManifestSHA256,
		ContentSHA256:    receipt.ContentSHA256,
		ParquetSHA256:    receipt.ParquetSHA256,
		Records:          receipt.Records,
		Tombstones:       receipt.Tombstones,
	}, nil
}

func (commit TransportReceiptCommit) warehouseReceipt() synctransport.WarehouseReceipt {
	return synctransport.WarehouseReceipt{
		ID:               commit.ReceiptID,
		Owner:            commit.Owner,
		Generation:       commit.Generation,
		Stream:           commit.Stream,
		Mode:             commit.Mode,
		CheckpointSHA256: commit.CheckpointSHA256,
		TombstonesSHA256: commit.TombstonesSHA256,
		ManifestSHA256:   commit.ManifestSHA256,
		ContentSHA256:    commit.ContentSHA256,
		ParquetSHA256:    commit.ParquetSHA256,
		Records:          commit.Records,
		Tombstones:       commit.Tombstones,
	}
}

func (commit TransportReceiptCommit) matchesWarehouseReceipt(receipt synctransport.WarehouseReceipt) bool {
	candidate, err := transportReceiptCommitFromWarehouseReceipt(receipt)
	return err == nil && commit == candidate
}

func cloneTransportReceiptCommits(commits []TransportReceiptCommit) ([]TransportReceiptCommit, error) {
	if len(commits) == 0 {
		return nil, nil
	}
	clone := make([]TransportReceiptCommit, 0, len(commits))
	seen := make(map[TransportReceiptCommit]struct{}, len(commits))
	for _, commit := range commits {
		if _, err := transportReceiptCommitFromWarehouseReceipt(commit.warehouseReceipt()); err != nil {
			return nil, err
		}
		if _, present := seen[commit]; present {
			return nil, fmt.Errorf("duplicate receipt %q", commit.ReceiptID)
		}
		seen[commit] = struct{}{}
		clone = append(clone, commit)
	}
	return clone, nil
}

func appendTransportReceiptCommits(current, additions []TransportReceiptCommit) ([]TransportReceiptCommit, error) {
	combined, err := cloneTransportReceiptCommits(current)
	if err != nil {
		return nil, err
	}
	if len(additions) == 0 {
		return combined, nil
	}
	seen := make(map[TransportReceiptCommit]struct{}, len(combined)+len(additions))
	for _, commit := range combined {
		seen[commit] = struct{}{}
	}
	for _, commit := range additions {
		if _, err := transportReceiptCommitFromWarehouseReceipt(commit.warehouseReceipt()); err != nil {
			return nil, err
		}
		if _, present := seen[commit]; present {
			continue
		}
		seen[commit] = struct{}{}
		combined = append(combined, commit)
	}
	return combined, nil
}

func (state StreamState) hasCommittedTransportReceipt(receipt synctransport.WarehouseReceipt) bool {
	for _, commit := range state.CommittedTransportReceipts {
		if commit.matchesWarehouseReceipt(receipt) {
			return true
		}
	}
	return false
}

func (state StreamState) hasLegacyCommittedTransportCheckpoint(candidate synccontract.CheckpointEnvelope) bool {
	if state.LegacyCommittedTransportCheckpoint != nil {
		return reflect.DeepEqual(candidate, *state.LegacyCommittedTransportCheckpoint)
	}
	if state.TransportReceiptAssociationVersion != 0 || len(state.CommittedTransportReceipts) != 0 || state.Checkpoint == nil || state.Checkpoint.CommittedAt == nil {
		return false
	}
	checkpoint := state.Checkpoint.Clone()
	checkpoint.CommittedAt = nil
	return reflect.DeepEqual(candidate, checkpoint)
}

func removeTransportReceiptCommit(commits []TransportReceiptCommit, receipt synctransport.WarehouseReceipt) ([]TransportReceiptCommit, bool) {
	remaining := make([]TransportReceiptCommit, 0, len(commits))
	removed := false
	for _, commit := range commits {
		if commit.matchesWarehouseReceipt(receipt) {
			removed = true
			continue
		}
		remaining = append(remaining, commit)
	}
	if len(remaining) == 0 {
		return nil, removed
	}
	return remaining, removed
}

func streamStateCursor(state StreamState) (string, bool) {
	if !streamStateHasSourcePosition(state) {
		return "", false
	}
	return string(state.Checkpoint.Position.Primary), true
}

func streamStateHasSourcePosition(state StreamState) bool {
	if state.Checkpoint == nil {
		return false
	}
	if state.Checkpoint.PositionObserved != nil {
		return *state.Checkpoint.PositionObserved
	}
	return len(state.Checkpoint.Position.Primary) != 0
}

func streamReadState(state StreamState, generationID int64) map[string]string {
	readState := map[string]string{
		"generation_id": strconv.FormatInt(generationID, 10),
	}
	if cursor, present := streamStateCursor(state); present {
		readState["cursor"] = cursor
	}
	return readState
}

func streamReadCursorState(state StreamState) connectors.OpaqueCursorState {
	if !streamStateHasSourcePosition(state) {
		return connectors.OpaqueCursorState{}
	}
	return connectors.OpaqueCursorState{
		Token:   append([]byte(nil), state.Checkpoint.Position.Primary...),
		Present: true,
	}
}

type streamCursorTracker struct {
	sourceOrdered  connectors.SourceOrderedCursorReader
	prior          string
	priorObserved  bool
	nextLegacy     string
	nextOpaque     synccontract.OpaqueToken
	nextObserved   bool
	report         string
	reportObserved bool
}

func newStreamCursorTracker(state StreamState, source connectors.Connector, config connectors.RuntimeConfig, field string, mode SourceSyncMode) (streamCursorTracker, error) {
	tracker := streamCursorTracker{}
	if mode == SourceSyncIncremental {
		prior, observed := streamStateCursor(state)
		tracker.prior = prior
		tracker.priorObserved = observed
		tracker.nextLegacy = prior
		tracker.nextObserved = observed
		if observed {
			tracker.nextOpaque = append(synccontract.OpaqueToken(nil), state.Checkpoint.Position.Primary...)
		}
	}
	if strings.TrimSpace(field) == "" {
		return tracker, nil
	}
	if sourceOrdered, ok := source.(connectors.SourceOrderedCursorReader); ok {
		if err := sourceOrdered.ValidateCursorField(config, field); err != nil {
			return streamCursorTracker{}, fmt.Errorf("validate source-ordered cursor field: %w", err)
		}
		tracker.sourceOrdered = sourceOrdered
	}
	return tracker, nil
}

func (t streamCursorTracker) legacyLowerBound() (string, bool) {
	if t.sourceOrdered != nil {
		return "", false
	}
	return t.prior, t.priorObserved
}

func (t *streamCursorTracker) observe(record connectors.Record, field string, mode SourceSyncMode) (string, connectors.OpaqueCursorState, bool, error) {
	cursor, err := recordCursor(record, field)
	if err != nil {
		return "", connectors.OpaqueCursorState{}, false, err
	}
	if t.sourceOrdered != nil {
		state, err := t.sourceOrdered.CursorStateFromRecord(record, field)
		if err != nil {
			return "", connectors.OpaqueCursorState{}, false, err
		}
		if !state.Present {
			return "", connectors.OpaqueCursorState{}, false, fmt.Errorf("source-ordered cursor reader did not provide a cursor state")
		}
		t.nextOpaque = append(t.nextOpaque[:0], state.Token...)
		t.nextObserved = true
		t.report = cursor
		t.reportObserved = true
		return cursor, state, true, nil
	}
	if mode == SourceSyncIncremental && t.priorObserved && compareCursor(cursor, t.prior) < 0 {
		return cursor, connectors.OpaqueCursorState{}, false, nil
	}
	if !t.nextObserved || compareCursor(cursor, t.nextLegacy) > 0 {
		t.nextLegacy = cursor
		t.nextObserved = true
	}
	return cursor, connectors.OpaqueCursorState{}, true, nil
}

func (t streamCursorTracker) readRequest(stream string, config connectors.RuntimeConfig, state StreamState, generationID int64, mode SourceSyncMode) connectors.ReadRequest {
	readConfig := config
	readConfig.Config = cloneStringMap(config.Config)
	request := connectors.ReadRequest{
		Stream: stream,
		Config: readConfig,
		State: map[string]string{
			"generation_id": strconv.FormatInt(generationID, 10),
		},
	}
	if mode != SourceSyncIncremental {
		return request
	}
	request.State = streamReadState(state, generationID)
	request.CursorState = streamReadCursorState(state)
	if cursor, present := t.legacyLowerBound(); present {
		request.Config.Config["since"] = cursor
	}
	return request
}

func (t streamCursorTracker) checkpoint() (synccontract.OpaqueToken, bool) {
	if t.sourceOrdered != nil {
		return append(synccontract.OpaqueToken(nil), t.nextOpaque...), t.nextObserved
	}
	return synccontract.OpaqueToken([]byte(t.nextLegacy)), t.nextObserved
}

func (t streamCursorTracker) reportedCursor() (string, bool) {
	if t.sourceOrdered != nil {
		return t.report, t.reportObserved
	}
	return t.nextLegacy, t.nextObserved
}

func streamSourceIdentity(source connectors.Connector, credential CredentialMeta, streamName string) synccontract.SourceIdentity {
	return synccontract.SourceIdentity{
		Engine:           source.Name(),
		AccountOrCluster: credential.ID,
		ObjectScope:      streamName,
	}
}

func streamSourceGeneration(source connectors.Connector, credential CredentialMeta, runtime connectors.RuntimeConfig, streamName string) synccontract.OpaqueToken {
	keys := make([]string, 0, len(runtime.Config))
	for key := range runtime.Config {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var canonical strings.Builder
	for _, value := range []string{source.Name(), credential.ID, streamName} {
		canonical.WriteString(strconv.Quote(value))
		canonical.WriteByte('\n')
	}
	for _, key := range keys {
		canonical.WriteString(strconv.Quote(key))
		canonical.WriteByte('\n')
		canonical.WriteString(strconv.Quote(runtime.Config[key]))
		canonical.WriteByte('\n')
	}
	digest := sha256.Sum256([]byte(canonical.String()))
	return append(synccontract.OpaqueToken(nil), digest[:]...)
}

func streamResumeExpectation(source connectors.Connector, credential CredentialMeta, runtime connectors.RuntimeConfig, streamName string) synccontract.ResumeExpectation {
	return synccontract.ResumeExpectation{
		Source:           streamSourceIdentity(source, credential, streamName),
		SourceGeneration: streamSourceGeneration(source, credential, runtime, streamName),
	}
}

func validateStreamStateResume(state StreamState, expected synccontract.ResumeExpectation) error {
	if state.Checkpoint == nil {
		return nil
	}
	return state.Checkpoint.ValidateResume(expected)
}

func legacyCheckpointEnvelope(source synccontract.ResumeExpectation, stream StreamConfig, runID string, cursor synccontract.OpaqueToken, positionObserved bool, observedAt time.Time) synccontract.CheckpointEnvelope {
	dedupeIdentity, _ := json.Marshal(stream.PrimaryKey)
	return synccontract.CheckpointEnvelope{
		StateVersion:    synccontract.StateVersion,
		Source:          source.Source,
		Mechanism:       "legacy_scalar_cursor",
		SnapshotBarrier: &synccontract.SnapshotBarrier{Kind: "legacy_run", Token: synccontract.OpaqueToken([]byte(runID))},
		Position: synccontract.CheckpointPosition{
			Primary:    append(synccontract.OpaqueToken(nil), cursor...),
			TieBreaker: synccontract.OpaqueToken([]byte(runID)),
		},
		PositionObserved: &positionObserved,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: source.SourceGeneration,
		SchemaVersion:    "legacy-app-v1",
		ProtocolVersion:  "connector-read-v1",
		Dedupe: synccontract.DedupeIdentity{
			Kind:  "legacy_primary_key",
			Value: synccontract.OpaqueToken(dedupeIdentity),
		},
		DedupeWindow: synccontract.DedupeWindow{
			Kind:  "legacy_run",
			Start: synccontract.OpaqueToken([]byte(runID)),
			End:   synccontract.OpaqueToken([]byte(runID)),
		},
		ObservedAt: observedAt,
	}
}

func committedLegacyStreamState(conn Connection, source synccontract.ResumeExpectation, streamName string, stream StreamConfig, runID string, cursor synccontract.OpaqueToken, positionObserved bool, generationID int64, recordsLoaded int, observedAt time.Time, acknowledgement synccontract.DownstreamAcknowledgement) (StreamState, error) {
	candidate := legacyCheckpointEnvelope(source, stream, runID, cursor, positionObserved, observedAt)
	var committed synccontract.CheckpointEnvelope
	if err := synccontract.CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(checkpoint synccontract.CheckpointEnvelope) error {
		committed = checkpoint
		return nil
	}); err != nil {
		return StreamState{}, err
	}
	return StreamState{
		Connection:          conn.Name,
		Stream:              streamName,
		Checkpoint:          &committed,
		GenerationID:        generationID,
		LastSuccessfulRunID: runID,
		RecordsLoaded:       recordsLoaded,
		UpdatedAt:           acknowledgement.AcknowledgedAt,
	}, nil
}
