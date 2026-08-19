package synccontract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestModeVocabularyIsClosed(t *testing.T) {
	want := []Mode{
		ModeFullOverwrite,
		ModeFullAppend,
		ModeIncrementalAppend,
		ModeIncrementalUpsert,
		ModeIncrementalDedupe,
		ModeIncrementalDedupeHistory,
		ModeChangeCapture,
	}
	if got := AllModes(); !reflect.DeepEqual(got, want) {
		t.Fatalf("AllModes() = %#v, want %#v", got, want)
	}
	for _, mode := range want {
		if err := mode.Validate(); err != nil {
			t.Fatalf("%q rejected: %v", mode, err)
		}
	}
	for _, mode := range []Mode{"full_refresh_append", "incremental_append_deduped", "", "incremental_merge"} {
		if err := mode.Validate(); err == nil {
			t.Fatalf("%q accepted outside closed vocabulary", mode)
		}
	}
}

func TestCheckpointEnvelopePreservesOpaqueTokensAndPartitionState(t *testing.T) {
	checkpoint := validCheckpoint()
	encoded, err := json.Marshal(checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte(`"partitions":"`)) {
		t.Fatalf("partitions collapsed to a scalar: %s", encoded)
	}

	var decoded CheckpointEnvelope
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatalf("decoded checkpoint invalid: %v", err)
	}
	if !bytes.Equal(decoded.SnapshotBarrier.Token, checkpoint.SnapshotBarrier.Token) {
		t.Fatalf("snapshot token changed: got %x want %x", decoded.SnapshotBarrier.Token, checkpoint.SnapshotBarrier.Token)
	}
	if !bytes.Equal(decoded.Position.Primary, checkpoint.Position.Primary) || !bytes.Equal(decoded.Position.TieBreaker, checkpoint.Position.TieBreaker) {
		t.Fatalf("global ordering token changed: got %#v want %#v", decoded.Position, checkpoint.Position)
	}
	if decoded.PositionObserved == nil || checkpoint.PositionObserved == nil || *decoded.PositionObserved != *checkpoint.PositionObserved {
		t.Fatalf("position observation changed: got %#v want %#v", decoded.PositionObserved, checkpoint.PositionObserved)
	}
	if !bytes.Equal(decoded.Partitions[0].Partition, checkpoint.Partitions[0].Partition) || !bytes.Equal(decoded.Partitions[0].Position.Primary, checkpoint.Partitions[0].Position.Primary) {
		t.Fatalf("partition token changed: got %#v want %#v", decoded.Partitions[0], checkpoint.Partitions[0])
	}
	if !bytes.Equal(decoded.Dedupe.Value, checkpoint.Dedupe.Value) {
		t.Fatalf("dedupe token changed: got %x want %x", decoded.Dedupe.Value, checkpoint.Dedupe.Value)
	}
	if !bytes.Equal(decoded.DedupeWindow.Start, checkpoint.DedupeWindow.Start) || !bytes.Equal(decoded.DedupeWindow.End, checkpoint.DedupeWindow.End) {
		t.Fatalf("dedupe window changed: got %#v want %#v", decoded.DedupeWindow, checkpoint.DedupeWindow)
	}

	clone := checkpoint.Clone()
	clone.Position.Primary[0] ^= 0xff
	clone.Partitions[0].Position.Primary[0] ^= 0xff
	clone.SnapshotBarrier.Token[0] ^= 0xff
	clone.SourceGeneration[0] ^= 0xff
	clone.Dedupe.Value[0] ^= 0xff
	clone.DedupeWindow.Start[0] ^= 0xff
	clone.DedupeWindow.End[0] ^= 0xff
	*clone.PositionObserved = false
	if bytes.Equal(clone.Position.Primary, checkpoint.Position.Primary) || bytes.Equal(clone.Partitions[0].Position.Primary, checkpoint.Partitions[0].Position.Primary) || bytes.Equal(clone.SnapshotBarrier.Token, checkpoint.SnapshotBarrier.Token) || bytes.Equal(clone.SourceGeneration, checkpoint.SourceGeneration) || bytes.Equal(clone.Dedupe.Value, checkpoint.Dedupe.Value) || bytes.Equal(clone.DedupeWindow.Start, checkpoint.DedupeWindow.Start) || bytes.Equal(clone.DedupeWindow.End, checkpoint.DedupeWindow.End) {
		t.Fatal("CheckpointEnvelope.Clone() aliases opaque token storage")
	}
	if !*checkpoint.PositionObserved {
		t.Fatal("CheckpointEnvelope.Clone() aliases position observation")
	}
}

func TestCheckpointEnvelopeRejectsDuplicatePartitionSlots(t *testing.T) {
	checkpoint := validCheckpoint()
	checkpoint.Partitions = append(checkpoint.Partitions, checkpoint.Partitions[0].Clone())
	if err := checkpoint.Validate(); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate partition state error = %v, want duplicate rejection", err)
	}
}

func TestCheckpointEnvelopeRequiresExplicitDedupeWindow(t *testing.T) {
	checkpoint := validCheckpoint()
	checkpoint.DedupeWindow = DedupeWindow{}
	if err := checkpoint.Validate(); err == nil || !strings.Contains(err.Error(), "dedupe window") {
		t.Fatalf("missing dedupe window error = %v, want explicit-window rejection", err)
	}
}

func TestCheckpointEnvelopeRejectsNilPartitionsOnResumeAfterClone(t *testing.T) {
	checkpoint := validCheckpoint()
	committedAt := checkpoint.ObservedAt.Add(time.Minute)
	checkpoint.CommittedAt = &committedAt
	checkpoint.Partitions = nil

	clone := checkpoint.Clone()
	if clone.Partitions != nil {
		t.Fatalf("CheckpointEnvelope.Clone() converted nil partitions to %#v", clone.Partitions)
	}
	err := clone.ValidateResume(ResumeExpectation{Source: clone.Source, SourceGeneration: clone.SourceGeneration})
	var recovery *RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != RecoveryOutcomeInvalidCheckpoint {
		t.Fatalf("ValidateResume() error = %T %v, want invalid checkpoint rebootstrap", err, err)
	}
}

func TestResumeOutcomesRequireExplicitRebootstrap(t *testing.T) {
	checkpoint := validCheckpoint()
	committedAt := checkpoint.ObservedAt.Add(time.Minute)
	checkpoint.CommittedAt = &committedAt
	before := checkpoint.Clone()
	expected := ResumeExpectation{Source: checkpoint.Source, SourceGeneration: checkpoint.SourceGeneration}
	if err := checkpoint.ValidateResume(expected); err != nil {
		t.Fatalf("matching state rejected: %v", err)
	}

	tests := []struct {
		name    string
		err     error
		outcome RecoveryOutcome
	}{
		{
			name:    "invalid checkpoint version",
			err:     CheckpointEnvelope{StateVersion: StateVersion + 1}.ValidateResume(expected),
			outcome: RecoveryOutcomeInvalidCheckpoint,
		},
		{
			name:    "retention gap",
			err:     RequireRebootstrap(RecoveryOutcomeRetentionGap, "provider retained no requested position"),
			outcome: RecoveryOutcomeRetentionGap,
		},
		{
			name:    "invalidated slot",
			err:     RequireRebootstrap(RecoveryOutcomeInvalidatedSlot, "slot dropped"),
			outcome: RecoveryOutcomeInvalidatedSlot,
		},
		{
			name:    "expired token",
			err:     RequireRebootstrap(RecoveryOutcomeExpiredToken, "resume token expired"),
			outcome: RecoveryOutcomeExpiredToken,
		},
		{
			name:    "source generation changed",
			err:     checkpoint.ValidateResume(ResumeExpectation{Source: checkpoint.Source, SourceGeneration: OpaqueToken("new-generation")}),
			outcome: RecoveryOutcomeSourceGenerationChanged,
		},
		{
			name:    "source identity incompatible",
			err:     checkpoint.ValidateResume(ResumeExpectation{Source: SourceIdentity{Engine: "postgres", AccountOrCluster: "other-cluster", ObjectScope: checkpoint.Source.ObjectScope}, SourceGeneration: checkpoint.SourceGeneration}),
			outcome: RecoveryOutcomeSourceIdentityIncompatible,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("error = nil")
			}
			var typed *RebootstrapRequiredError
			if !errors.As(tt.err, &typed) {
				t.Fatalf("error = %T %v, want RebootstrapRequiredError", tt.err, tt.err)
			}
			if typed.Outcome != tt.outcome {
				t.Fatalf("outcome = %q, want %q", typed.Outcome, tt.outcome)
			}
			if !errors.Is(tt.err, ErrRebootstrapRequired) {
				t.Fatalf("errors.Is(%v, ErrRebootstrapRequired) = false", tt.err)
			}
		})
	}
	if !reflect.DeepEqual(checkpoint, before) {
		t.Fatalf("resume validation mutated state: got %#v want %#v", checkpoint, before)
	}
}

func TestCommitAfterDownstreamAcknowledgement(t *testing.T) {
	candidate := validCheckpoint()
	commits := 0
	err := CommitAfterDownstreamAcknowledgement(candidate, DownstreamAcknowledgement{}, func(CheckpointEnvelope) error {
		commits++
		return nil
	})
	if !errors.Is(err, ErrDownstreamAcknowledgementRequired) {
		t.Fatalf("missing acknowledgement error = %v", err)
	}
	if commits != 0 {
		t.Fatalf("committed %d times before durable acknowledgement", commits)
	}

	acknowledgedAt := candidate.ObservedAt.Add(time.Minute)
	err = CommitAfterDownstreamAcknowledgement(candidate, DownstreamAcknowledgement{Sink: "warehouse", AcknowledgedAt: acknowledgedAt}, func(CheckpointEnvelope) error {
		commits++
		return nil
	})
	if !errors.Is(err, ErrDownstreamAcknowledgementRequired) {
		t.Fatalf("forged acknowledgement error = %v", err)
	}
	if commits != 0 {
		t.Fatalf("committed %d times with a forged acknowledgement", commits)
	}

	acknowledgement, err := NewDurableDownstreamAcknowledgement("warehouse", acknowledgedAt)
	if err != nil {
		t.Fatal(err)
	}
	var committed CheckpointEnvelope
	err = CommitAfterDownstreamAcknowledgement(candidate, acknowledgement, func(got CheckpointEnvelope) error {
		commits++
		committed = got
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if commits != 1 {
		t.Fatalf("commit count = %d, want 1", commits)
	}
	if committed.CommittedAt == nil || !committed.CommittedAt.Equal(acknowledgedAt) {
		t.Fatalf("committed_at = %v, want %s", committed.CommittedAt, acknowledgedAt)
	}
	if !committed.ObservedAt.Equal(candidate.ObservedAt) {
		t.Fatalf("observed_at changed: got %s want %s", committed.ObservedAt, candidate.ObservedAt)
	}
	if candidate.CommittedAt != nil {
		t.Fatal("commit mutated uncommitted candidate")
	}
}

func TestResumeRejectsObservedButUncommittedCheckpoint(t *testing.T) {
	checkpoint := validCheckpoint()
	err := checkpoint.ValidateResume(ResumeExpectation{Source: checkpoint.Source, SourceGeneration: checkpoint.SourceGeneration})
	var recovery *RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != RecoveryOutcomeInvalidCheckpoint {
		t.Fatalf("ValidateResume() error = %T %v, want invalid checkpoint rebootstrap", err, err)
	}
}

func TestTombstoneClosesHistoryWindowInsteadOfPhysicalDelete(t *testing.T) {
	beforeImage := json.RawMessage(`{"id":"42","email":"ada@example.test"}`)
	tombstone := Tombstone{
		Operation:   OperationDelete,
		EventID:     OpaqueToken{0xff, 0x01, 0x00},
		Key:         json.RawMessage(`{"id":"42"}`),
		DeleteImage: DeleteImageBefore,
		Before:      beforeImage,
		Position: CheckpointPosition{
			Primary:    OpaqueToken("00000010"),
			TieBreaker: OpaqueToken{0x00, 0xfe},
		},
	}
	if err := tombstone.Validate(); err != nil {
		t.Fatal(err)
	}
	closedAt := time.Date(2026, time.August, 6, 1, 2, 3, 0, time.UTC)
	mutation, err := CloseHistoryWindow(tombstone, closedAt)
	if err != nil {
		t.Fatal(err)
	}
	if mutation.Action != HistoryDeleteCloseValidityWindow || mutation.IsCurrent || !mutation.ValidTo.Equal(closedAt) {
		t.Fatalf("history mutation = %#v, want validity close", mutation)
	}
	if HistoryValidFromColumn != "_valid_from" || HistoryValidToColumn != "_valid_to" || HistoryIsCurrentColumn != "_is_current" {
		t.Fatalf("history columns = %q, %q, %q", HistoryValidFromColumn, HistoryValidToColumn, HistoryIsCurrentColumn)
	}
	if err := HistoryDeletePhysicalTargetDelete.Validate(); err == nil {
		t.Fatal("physical target delete accepted for history mode")
	}

	keyOnly := tombstone.Clone()
	keyOnly.DeleteImage = DeleteImageKeyOnly
	keyOnly.Before = nil
	if err := keyOnly.Validate(); err != nil {
		t.Fatalf("key-only tombstone rejected: %v", err)
	}
}

func TestTombstoneSeparatesUnavailableImagesAndSourceStateEvents(t *testing.T) {
	position := CheckpointPosition{Primary: OpaqueToken("00000010"), TieBreaker: OpaqueToken{0x00, 0xfe}}
	unavailable := Tombstone{
		Operation:   OperationDelete,
		EventID:     OpaqueToken{0xff, 0x02, 0x00},
		Key:         json.RawMessage(`{"id":"42"}`),
		DeleteImage: DeleteImageUnavailable,
		Position:    position,
	}
	if err := unavailable.Validate(); err != nil {
		t.Fatalf("unavailable-image tombstone rejected: %v", err)
	}
	if _, err := CloseHistoryWindow(unavailable, time.Date(2026, time.August, 6, 1, 2, 3, 0, time.UTC)); err != nil {
		t.Fatalf("history close rejected an unavailable-image row delete: %v", err)
	}

	for _, tombstone := range []Tombstone{
		{Operation: OperationTruncate, EventID: OpaqueToken("truncate-event"), Position: position},
		{Operation: OperationInvalidate, EventID: OpaqueToken("invalidate-event"), Position: position},
	} {
		if err := tombstone.Validate(); err != nil {
			t.Fatalf("%s tombstone rejected: %v", tombstone.Operation, err)
		}
		if _, err := CloseHistoryWindow(tombstone, time.Date(2026, time.August, 6, 1, 2, 3, 0, time.UTC)); err == nil {
			t.Fatalf("history close accepted source-level %s", tombstone.Operation)
		}
	}

	invalid := Tombstone{Operation: OperationTruncate, EventID: OpaqueToken("truncate-event"), Key: json.RawMessage(`{"id":"42"}`), Position: position}
	if err := invalid.Validate(); err == nil {
		t.Fatal("truncate tombstone accepted row-delete fields")
	}
	unavailable.Before = json.RawMessage(`{"id":"42"}`)
	if err := unavailable.Validate(); err == nil {
		t.Fatal("unavailable-image tombstone accepted a before image")
	}
}

func TestTombstoneRejectsNullRowFields(t *testing.T) {
	position := CheckpointPosition{Primary: OpaqueToken("00000010"), TieBreaker: OpaqueToken{0x00, 0xfe}}
	tests := []struct {
		name      string
		tombstone Tombstone
	}{
		{
			name: "null key",
			tombstone: Tombstone{
				Operation:   OperationDelete,
				EventID:     OpaqueToken("null-key"),
				Key:         json.RawMessage(" null "),
				DeleteImage: DeleteImageKeyOnly,
				Position:    position,
			},
		},
		{
			name: "null before image",
			tombstone: Tombstone{
				Operation:   OperationDelete,
				EventID:     OpaqueToken("null-before"),
				Key:         json.RawMessage(`{"id":"42"}`),
				DeleteImage: DeleteImageBefore,
				Before:      json.RawMessage("null"),
				Position:    position,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tombstone.Validate(); err == nil {
				t.Fatal("Tombstone.Validate() accepted null row field")
			}
		})
	}
}

func TestNativeContractNeedsRegisteredRunnableExecutorAndFixtureEvidence(t *testing.T) {
	contract := NativeCommandContract{
		ContractVersion: NativeCommandContractVersion,
		Protocol:        "postgres_wire",
		Command:         "logical_replication",
		Executor:        ExecutorReference{Kind: "native", ID: "postgres-logical"},
		Modes:           []Mode{ModeChangeCapture},
		Conformance:     RequiredConformanceEvidence(),
	}
	if err := contract.Validate(); err != nil {
		t.Fatalf("contract rejected: %v", err)
	}
	registry, err := NewNativeExecutorRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if registry.Admits(contract) {
		t.Fatal("contract executable without a registered native executor")
	}
	wrong := &fakeNativeExecutor{descriptor: NativeSyncExecutorDescriptor{
		Protocol: "postgres_wire",
		Command:  "logical_replication",
		Executor: ExecutorReference{Kind: "native", ID: "other"},
		Modes:    []Mode{ModeChangeCapture},
	}, evidence: RequiredConformanceEvidence()}
	if err := registry.Register(wrong); err != nil {
		t.Fatal(err)
	}
	if registry.Admits(contract) {
		t.Fatal("contract executable with a mismatched registered executor")
	}
	missingEvidence := &fakeNativeExecutor{descriptor: NativeSyncExecutorDescriptor{
		Protocol: contract.Protocol,
		Command:  contract.Command,
		Executor: contract.Executor,
		Modes:    contract.Modes,
	}}
	if err := registry.Register(missingEvidence); err == nil {
		t.Fatal("registry accepted a native executor without fixture evidence")
	}
	matching := &fakeNativeExecutor{descriptor: NativeSyncExecutorDescriptor{
		Protocol: contract.Protocol,
		Command:  contract.Command,
		Executor: contract.Executor,
		Modes:    contract.Modes,
	}, evidence: RequiredConformanceEvidence()}
	registry, err = NewNativeExecutorRegistry(matching)
	if err != nil {
		t.Fatal(err)
	}
	if !registry.Admits(contract) {
		t.Fatal("registry did not admit matching runnable executor and evidence")
	}
	if _, err := registry.Execute(context.Background(), contract, NativeSyncRequest{Mode: ModeChangeCapture}); err != nil {
		t.Fatal(err)
	}
	if matching.runCalls != 1 {
		t.Fatalf("native executor calls = %d, want 1", matching.runCalls)
	}
	for _, test := range []struct {
		field string
		value string
	}{
		{field: "protocol", value: "rest-v2"},
		{field: "protocol", value: "http-client"},
		{field: "command", value: "sql-run"},
		{field: "command", value: "shell-runner"},
		{field: "executor", value: "https-adapter"},
	} {
		invalid := contract
		switch test.field {
		case "protocol":
			invalid.Protocol = test.value
		case "command":
			invalid.Command = test.value
		case "executor":
			invalid.Executor.ID = test.value
		}
		if err := invalid.Validate(); err == nil {
			t.Fatalf("generic %s %q was accepted", test.field, test.value)
		}
	}

	encoded, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{`"api_surface"`, `"method"`, `"path"`, `"sql"`, `"http"`, `"shell"`} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Fatalf("native contract contains forbidden generic escape-hatch field %s: %s", forbidden, encoded)
		}
	}
}

func TestConformanceFixturesAreVersionedAndDefensivelyCopied(t *testing.T) {
	fixtures := ConformanceFixtures()
	if len(fixtures) == 0 {
		t.Fatal("ConformanceFixtures() returned no reusable fixtures")
	}
	ids := RequiredConformanceEvidence().FixtureIDs
	if len(ids) != len(fixtures) {
		t.Fatalf("evidence IDs = %d, fixtures = %d", len(ids), len(fixtures))
	}
	for _, expected := range []string{
		"change-insert",
		"change-update",
		"checkpoint-opaque-bytes",
		"partition-state-not-collapsed",
		"invalid-checkpoint-rebootstrap",
		"commit-after-durable-ack",
		"retention-gap-rebootstrap",
		"invalidated-slot-rebootstrap",
		"expired-token-rebootstrap",
		"generation-change-rebootstrap",
		"source-identity-rebootstrap",
		"history-delete-closes-window",
		"tombstone-key-only",
		"tombstone-before-image",
		"tombstone-unavailable-image",
		"source-truncate",
		"source-invalidate",
		"duplicate-replay-deduped",
		"snapshot-to-stream-handoff",
	} {
		if !containsFixtureID(ids, expected) {
			t.Fatalf("shared fixture corpus is missing %q", expected)
		}
	}
	fixtures[0].ID = "mutated"
	if ConformanceFixtures()[0].ID == "mutated" {
		t.Fatal("ConformanceFixtures() returned mutable shared state")
	}
}

func validCheckpoint() CheckpointEnvelope {
	positionObserved := true
	return CheckpointEnvelope{
		StateVersion: StateVersion,
		Source: SourceIdentity{
			Engine:           "postgres",
			AccountOrCluster: "cluster-a",
			ObjectScope:      "public.events",
		},
		Mechanism: "logical_replication",
		SnapshotBarrier: &SnapshotBarrier{
			Kind:  "exported_snapshot",
			Token: OpaqueToken{0xff, 0x00, 0x01, 0x80},
		},
		Position: CheckpointPosition{
			Primary:    OpaqueToken{0xff, 0x00, 0x02},
			TieBreaker: OpaqueToken{0x01, 0x00, 0xfe},
		},
		PositionObserved: &positionObserved,
		Partitions: []PartitionState{{
			Partition: OpaqueToken{0x00, 0xff, 0x02},
			Position: CheckpointPosition{
				Primary:    OpaqueToken{0x01, 0x02, 0x03},
				TieBreaker: OpaqueToken{0x04, 0x05, 0x06},
			},
		}},
		SourceGeneration: OpaqueToken{0x08, 0x00, 0xff},
		SchemaVersion:    "events-v4",
		ProtocolVersion:  "pgoutput-v1",
		Dedupe: DedupeIdentity{
			Kind:  "event_id",
			Value: OpaqueToken{0xff, 0x09, 0x00},
		},
		DedupeWindow: DedupeWindow{
			Kind:  "overlap",
			Start: OpaqueToken{0x01, 0x00, 0xff},
			End:   OpaqueToken{0x02, 0x00, 0xfe},
		},
		ObservedAt: time.Date(2026, time.August, 6, 0, 0, 0, 0, time.UTC),
	}
}

func containsFixtureID(ids []string, expected string) bool {
	for _, id := range ids {
		if id == expected {
			return true
		}
	}
	return false
}

type fakeNativeExecutor struct {
	descriptor NativeSyncExecutorDescriptor
	evidence   ConformanceEvidence
	runCalls   int
}

func (f *fakeNativeExecutor) NativeSyncExecutorDescriptor() NativeSyncExecutorDescriptor {
	return f.descriptor
}

func (f *fakeNativeExecutor) NativeSyncConformanceEvidence() ConformanceEvidence { return f.evidence }

func (f *fakeNativeExecutor) RunNativeSync(context.Context, NativeSyncRequest) (NativeSyncResult, error) {
	f.runCalls++
	return NativeSyncResult{}, nil
}
