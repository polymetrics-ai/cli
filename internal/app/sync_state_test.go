package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestLegacyScalarStreamStateRequiresRebootstrapBeforeRead(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("legacy_scalar_state", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	// A full refresh is the dangerous case: invalid legacy state must not be
	// cleared and converted into an unrequested scan just because this run's
	// selected mode would otherwise read a fresh snapshot.
	a, connection := setupSyncModeApp(t, source, "full_refresh_overwrite")

	var legacy StreamState
	if err := json.Unmarshal([]byte(`{"connection":"records_to_warehouse","stream":"records","cursor":"opaque-legacy-cursor","generation_id":3}`), &legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.Checkpoint == nil || legacy.Checkpoint.StateVersion != 0 {
		t.Fatalf("legacy state = %#v, want version-zero envelope", legacy)
	}
	a.state.StreamStates[streamStateKey(connection, "records")] = legacy

	_, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) {
		t.Fatalf("RunETL() error = %T %v, want typed rebootstrap outcome", err, err)
	}
	if recovery.Outcome != synccontract.RecoveryOutcomeInvalidCheckpoint {
		t.Fatalf("recovery outcome = %q, want invalid checkpoint", recovery.Outcome)
	}
	if len(source.requests) != 0 {
		t.Fatalf("source read after legacy state required rebootstrap: %#v", source.requests)
	}
	if got := a.state.StreamStates[streamStateKey(connection, "records")]; got.Checkpoint == nil || got.Checkpoint.StateVersion != 0 {
		t.Fatalf("legacy state was cleared or replaced: %#v", got)
	}
}

func TestIncrementalRunStoresCommittedStateEnvelopeAfterDownstreamSuccess(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("envelope_commit", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	state := a.state.StreamStates[streamStateKey(connection, "records")]
	if state.Checkpoint == nil {
		t.Fatal("successful run stored no checkpoint envelope")
	}
	if got := string(state.Checkpoint.Position.Primary); got != "2026-08-06T00:00:00Z" {
		t.Fatalf("primary checkpoint = %q", got)
	}
	if state.Checkpoint.CommittedAt == nil || state.Checkpoint.ObservedAt.IsZero() {
		t.Fatalf("checkpoint timestamps = %#v, want observed and committed", state.Checkpoint)
	}
	if state.Checkpoint.CommittedAt.Before(state.Checkpoint.ObservedAt) {
		t.Fatalf("checkpoint committed before observation: %#v", state.Checkpoint)
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encoded, []byte(`"cursor"`)) {
		t.Fatalf("stream state still persisted a scalar cursor: %s", encoded)
	}
	before := state.Checkpoint.Clone()

	source.records = []connectors.Record{{"id": "2", "updated_at": "2026-08-07T00:00:00Z"}}
	source.failAfter = 1
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err == nil {
		t.Fatal("RunETL(failing incremental) error = nil")
	}
	after := a.state.StreamStates[streamStateKey(connection, "records")].Checkpoint
	if after == nil || !reflect.DeepEqual(*after, before) {
		t.Fatalf("checkpoint advanced after unsuccessful downstream work: got %#v want %#v", after, before)
	}
}

func TestContractModeCannotReadWithoutNativeExecutor(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("contract_mode_without_executor", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "full_append")

	_, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
	var unavailable *synccontract.ModeNotExecutableError
	if !errors.As(err, &unavailable) {
		t.Fatalf("RunETL() error = %T %v, want ModeNotExecutableError", err, err)
	}
	if unavailable.Mode != synccontract.ModeFullAppend {
		t.Fatalf("unavailable mode = %q, want %q", unavailable.Mode, synccontract.ModeFullAppend)
	}
	if len(source.requests) != 0 {
		t.Fatalf("source read despite unavailable mode: %#v", source.requests)
	}
}

func TestConnectorETLDoesNotCommitAfterPartialDestinationResult(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("partial_write_source", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	conn, ok := a.findConnection(connection)
	if !ok {
		t.Fatal("connection missing")
	}
	mode, err := ParseSyncMode("incremental_append")
	if err != nil {
		t.Fatal(err)
	}
	destination := partialWriteDestination{scriptedSyncSource: newScriptedSyncSource("partial_write_destination", nil)}
	_, err = a.runConnectorETL(ctx, "run_partial", conn, source, connectors.RuntimeConfig{}, &destination, connectors.RuntimeConfig{}, "records", conn.Streams["records"], mode, 1)
	if err == nil {
		t.Fatal("runConnectorETL() error = nil after partial destination result")
	}
	if state := a.state.StreamStates[streamStateKey(connection, "records")]; state.Checkpoint != nil {
		t.Fatalf("checkpoint committed after partial destination result: %#v", state.Checkpoint)
	}
}

type partialWriteDestination struct {
	*scriptedSyncSource
}

func (d *partialWriteDestination) Write(_ context.Context, _ connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsWritten: len(records) - 1, RecordsFailed: 1}, nil
}
