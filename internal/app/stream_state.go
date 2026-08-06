package app

import (
	"encoding/json"
	"fmt"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

// UnmarshalJSON upgrades the former scalar cursor only into a deliberately
// non-resumable version-zero envelope. It preserves the old bytes for a
// diagnostic/rebootstrap workflow, but no execution path may use it as a new
// scalar cursor or silently replace it with a full scan.
func (s *StreamState) UnmarshalJSON(data []byte) error {
	type streamStateWire struct {
		Connection          string                           `json:"connection"`
		Stream              string                           `json:"stream"`
		Checkpoint          *synccontract.CheckpointEnvelope `json:"checkpoint"`
		LegacyCursor        *string                          `json:"cursor"`
		GenerationID        int64                            `json:"generation_id"`
		LastSuccessfulRunID string                           `json:"last_successful_run_id"`
		RecordsLoaded       int                              `json:"records_loaded"`
		UpdatedAt           time.Time                        `json:"updated_at"`
	}
	var wire streamStateWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.Checkpoint != nil && wire.LegacyCursor != nil {
		return fmt.Errorf("stream state cannot contain both checkpoint envelope and legacy cursor")
	}
	*s = StreamState{
		Connection:          wire.Connection,
		Stream:              wire.Stream,
		GenerationID:        wire.GenerationID,
		LastSuccessfulRunID: wire.LastSuccessfulRunID,
		RecordsLoaded:       wire.RecordsLoaded,
		UpdatedAt:           wire.UpdatedAt,
	}
	if wire.Checkpoint != nil {
		checkpoint := wire.Checkpoint.Clone()
		s.Checkpoint = &checkpoint
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

func streamStateCursor(state StreamState) string {
	if state.Checkpoint == nil {
		return ""
	}
	// This adapter is only for the pre-contract connector ReadRequest shape.
	// Native sync executors consume OpaqueToken directly from the envelope.
	return string(state.Checkpoint.Position.Primary)
}

func streamSourceIdentity(source connectors.Connector, credential CredentialMeta, streamName string) synccontract.SourceIdentity {
	return synccontract.SourceIdentity{
		Engine:           source.Name(),
		AccountOrCluster: credential.ID,
		ObjectScope:      streamName,
	}
}

func streamSourceGeneration(runtime connectors.RuntimeConfig) synccontract.OpaqueToken {
	generation := make(synccontract.OpaqueToken, 0, len(runtime.CredentialRevision)+len(runtime.ConfigurationDigest)+1)
	generation = append(generation, runtime.CredentialRevision...)
	generation = append(generation, 0)
	generation = append(generation, runtime.ConfigurationDigest...)
	return generation
}

func streamResumeExpectation(source connectors.Connector, credential CredentialMeta, runtime connectors.RuntimeConfig, streamName string) synccontract.ResumeExpectation {
	return synccontract.ResumeExpectation{
		Source:           streamSourceIdentity(source, credential, streamName),
		SourceGeneration: streamSourceGeneration(runtime),
	}
}

func validateStreamStateResume(state StreamState, expected synccontract.ResumeExpectation) error {
	if state.Checkpoint == nil {
		return nil
	}
	return state.Checkpoint.ValidateResume(expected)
}

func legacyCheckpointEnvelope(source synccontract.ResumeExpectation, stream StreamConfig, runID, cursor string, observedAt time.Time) synccontract.CheckpointEnvelope {
	dedupeIdentity, _ := json.Marshal(stream.PrimaryKey)
	return synccontract.CheckpointEnvelope{
		StateVersion:    synccontract.StateVersion,
		Source:          source.Source,
		Mechanism:       "legacy_scalar_cursor",
		SnapshotBarrier: &synccontract.SnapshotBarrier{Kind: "legacy_run", Token: synccontract.OpaqueToken([]byte(runID))},
		Position: synccontract.CheckpointPosition{
			Primary:    synccontract.OpaqueToken([]byte(cursor)),
			TieBreaker: synccontract.OpaqueToken([]byte(runID)),
		},
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

func committedLegacyStreamState(conn Connection, source synccontract.ResumeExpectation, streamName string, stream StreamConfig, runID, cursor string, generationID int64, recordsLoaded int, observedAt time.Time, acknowledgement synccontract.DownstreamAcknowledgement) (StreamState, error) {
	candidate := legacyCheckpointEnvelope(source, stream, runID, cursor, observedAt)
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
