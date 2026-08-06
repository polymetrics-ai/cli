package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

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

func TestRunETLStopsWhenInitialRunStateCannotPersist(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("initial_run_state_failure", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	beforeMemory, err := json.Marshal(a.state)
	if err != nil {
		t.Fatal(err)
	}
	beforeDisk, err := os.ReadFile(a.statePath)
	if err != nil {
		t.Fatal(err)
	}
	originalStore := a.store
	a.store.Path = t.TempDir()
	t.Cleanup(func() { a.store = originalStore })

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err == nil {
		t.Fatal("RunETL() error = nil when initial run state cannot persist")
	}
	if len(source.requests) != 0 {
		t.Fatalf("source read despite initial run-state failure: %#v", source.requests)
	}
	afterMemory, err := json.Marshal(a.state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterMemory, beforeMemory) {
		t.Fatalf("in-memory state changed after initial save failure: got %s want %s", afterMemory, beforeMemory)
	}
	afterDisk, err := os.ReadFile(a.statePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterDisk, beforeDisk) {
		t.Fatalf("persisted state changed after initial save failure: got %s want %s", afterDisk, beforeDisk)
	}
}

func TestCompleteRunPublishesPendingStreamStateOnlyAfterStateSave(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("pending_stream_state", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	conn, ok := a.findConnection(connection)
	if !ok {
		t.Fatal("connection missing")
	}
	mode, err := ParseStreamSyncMode(conn.Streams["records"])
	if err != nil {
		t.Fatal(err)
	}
	sourceConnector, sourceCredential, sourceRuntime, err := a.resolveEndpointWithCredential(ctx, conn.Source)
	if err != nil {
		t.Fatal(err)
	}
	_, destinationRuntime, err := a.resolveEndpoint(ctx, conn.Destination)
	if err != nil {
		t.Fatal(err)
	}
	run, err := a.beginRun(Run{ID: "run_pending_state", Type: "etl", Connection: connection, Stream: "records", Status: "running", StartedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	beforeMemory, err := json.Marshal(a.state)
	if err != nil {
		t.Fatal(err)
	}
	beforeDisk, err := os.ReadFile(a.statePath)
	if err != nil {
		t.Fatal(err)
	}
	expectation := streamResumeExpectation(sourceConnector, sourceCredential, sourceRuntime, "records")
	result, err := a.runWarehouseETL(ctx, run.ID, conn, sourceConnector, sourceRuntime, destinationRuntime, expectation, "records", conn.Streams["records"], mode, 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.PendingStreamState == nil {
		t.Fatal("runWarehouseETL() returned no pending stream state")
	}
	afterRunnerMemory, err := json.Marshal(a.state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterRunnerMemory, beforeMemory) {
		t.Fatalf("runner published pending state before persistence: got %s want %s", afterRunnerMemory, beforeMemory)
	}
	originalStore := a.store
	a.store.Path = t.TempDir()
	if _, err := a.completeRun(run.ID, result); err == nil {
		t.Fatal("completeRun() error = nil when state save fails")
	}
	afterMemory, err := json.Marshal(a.state)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterMemory, beforeMemory) {
		t.Fatalf("in-memory state changed after completion save failure: got %s want %s", afterMemory, beforeMemory)
	}
	afterDisk, err := os.ReadFile(a.statePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterDisk, beforeDisk) {
		t.Fatalf("persisted state changed after completion save failure: got %s want %s", afterDisk, beforeDisk)
	}
	a.store = originalStore
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	if got := source.requests[len(source.requests)-1].State["cursor"]; got != "" {
		t.Fatalf("later run resumed from unpublished checkpoint %q", got)
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

func TestNewIncrementalAppendCannotReadWithoutNativeExecutor(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("new_incremental_without_executor", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	for i := range a.state.Connections {
		if a.state.Connections[i].Name != connection {
			continue
		}
		stream := a.state.Connections[i].Streams["records"]
		stream.LegacyCompatibility = false
		a.state.Connections[i].Streams["records"] = stream
	}

	_, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
	var unavailable *synccontract.ModeNotExecutableError
	if !errors.As(err, &unavailable) || unavailable.Mode != synccontract.ModeIncrementalAppend {
		t.Fatalf("RunETL() error = %T %v, want incremental native admission failure", err, err)
	}
	if len(source.requests) != 0 {
		t.Fatalf("source read despite unavailable incremental contract mode: %#v", source.requests)
	}
}

func TestUncommittedCheckpointRequiresRebootstrapBeforeRead(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("uncommitted_checkpoint", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	stateKey := streamStateKey(connection, "records")
	state := a.state.StreamStates[stateKey]
	checkpoint := state.Checkpoint.Clone()
	checkpoint.CommittedAt = nil
	state.Checkpoint = &checkpoint
	a.state.StreamStates[stateKey] = state
	readCount := len(source.requests)

	_, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeInvalidCheckpoint {
		t.Fatalf("RunETL() error = %T %v, want invalid checkpoint rebootstrap", err, err)
	}
	if len(source.requests) != readCount {
		t.Fatalf("source read with observed-only checkpoint: %#v", source.requests)
	}
}

func TestStreamStateRejectsRecreatedCredentialAliasBeforeRead(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("stable_credential_identity", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	stateKey := streamStateKey(connection, "records")
	before := a.state.StreamStates[stateKey].Checkpoint.Clone()
	if before.Source.AccountOrCluster == "source" {
		t.Fatal("checkpoint source identity retained the mutable credential alias")
	}
	if err := a.RemoveCredential(ctx, "source"); err != nil {
		t.Fatal(err)
	}
	if _, err := a.AddCredential(ctx, AddCredentialRequest{Name: "source", Connector: source.Name()}); err != nil {
		t.Fatal(err)
	}
	readCount := len(source.requests)

	_, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeSourceIdentityIncompatible {
		t.Fatalf("RunETL() error = %T %v, want source identity rebootstrap", err, err)
	}
	if len(source.requests) != readCount {
		t.Fatalf("source read after credential identity changed: %#v", source.requests)
	}
	after := a.state.StreamStates[stateKey].Checkpoint
	if after == nil || !reflect.DeepEqual(*after, before) {
		t.Fatalf("checkpoint changed after credential identity recovery: got %#v want %#v", after, before)
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
	mode, err := ParseStreamSyncMode(conn.Streams["records"])
	if err != nil {
		t.Fatal(err)
	}
	_, sourceCredential, sourceRuntime, err := a.resolveEndpointWithCredential(ctx, conn.Source)
	if err != nil {
		t.Fatal(err)
	}
	expectation := streamResumeExpectation(source, sourceCredential, sourceRuntime, "records")
	destination := partialWriteDestination{scriptedSyncSource: newScriptedSyncSource("partial_write_destination", nil)}
	_, err = a.runConnectorETL(ctx, "run_partial", conn, source, sourceRuntime, &destination, connectors.RuntimeConfig{}, expectation, "records", conn.Streams["records"], mode, 1)
	if err == nil {
		t.Fatal("runConnectorETL() error = nil after partial destination result")
	}
	if destination.acknowledgements != 0 {
		t.Fatalf("durability acknowledgement count = %d, want 0 after partial write", destination.acknowledgements)
	}
	if state := a.state.StreamStates[streamStateKey(connection, "records")]; state.Checkpoint != nil {
		t.Fatalf("checkpoint committed after partial destination result: %#v", state.Checkpoint)
	}
}

func TestConnectorETLRefusesDestinationWithoutDurableAcknowledgementBeforeRead(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("undurable_connector_source", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	conn, ok := a.findConnection(connection)
	if !ok {
		t.Fatal("connection missing")
	}
	mode, err := ParseStreamSyncMode(conn.Streams["records"])
	if err != nil {
		t.Fatal(err)
	}
	_, sourceCredential, sourceRuntime, err := a.resolveEndpointWithCredential(ctx, conn.Source)
	if err != nil {
		t.Fatal(err)
	}
	expectation := streamResumeExpectation(source, sourceCredential, sourceRuntime, "records")
	destination := newScriptedSyncSource("undurable_connector_destination", nil)
	_, err = a.runConnectorETL(ctx, "run_undurable", conn, source, sourceRuntime, destination, connectors.RuntimeConfig{}, expectation, "records", conn.Streams["records"], mode, 1)
	var admission *synccontract.DestinationDurabilityAdmissionError
	if !errors.As(err, &admission) {
		t.Fatalf("runConnectorETL() error = %T %v, want typed durability admission failure", err, err)
	}
	if !errors.Is(err, synccontract.ErrDurableETLDestinationRequired) {
		t.Fatalf("runConnectorETL() error = %v, want durable destination admission error", err)
	}
	if admission.Destination != destination.Name() {
		t.Fatalf("admission destination = %q, want %q", admission.Destination, destination.Name())
	}
	if !strings.Contains(err.Error(), "migrate this connection") {
		t.Fatalf("durability admission error lacks migration guidance: %v", err)
	}
	if len(source.requests) != 0 {
		t.Fatalf("source read before generic destination durability admission: %#v", source.requests)
	}
}

func TestRunETLRefusesOutboxWithoutDurableAcknowledgementBeforeRead(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("undurable_outbox_source", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, _ := setupSyncModeApp(t, source, "incremental_append")
	outboxDir := filepath.Join(a.projectDir, "outbox")
	if _, err := a.AddCredential(ctx, AddCredentialRequest{
		Name:      "outbox",
		Connector: "outbox",
		Config:    map[string]string{"path": outboxDir},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.CreateConnection(ctx, CreateConnectionRequest{
		Name:        "records_to_outbox",
		Source:      EndpointConfig{Connector: source.Name(), Credential: "source"},
		Destination: EndpointConfig{Connector: "outbox", Credential: "outbox"},
		Streams: map[string]StreamConfig{
			"records": {
				SyncMode:         "incremental_append",
				CursorField:      "updated_at",
				DestinationTable: "records",
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	_, err := a.RunETL(ctx, RunETLRequest{Connection: "records_to_outbox", Stream: "records", BatchSize: 1})
	var admission *synccontract.DestinationDurabilityAdmissionError
	if !errors.As(err, &admission) {
		t.Fatalf("RunETL() error = %T %v, want typed durability admission failure", err, err)
	}
	if admission.Destination != "outbox" {
		t.Fatalf("admission destination = %q, want outbox", admission.Destination)
	}
	if !strings.Contains(err.Error(), "migrate this connection") {
		t.Fatalf("durability admission error lacks migration guidance: %v", err)
	}
	if len(source.requests) != 0 {
		t.Fatalf("source read before outbox durability admission: %#v", source.requests)
	}
	if _, statErr := os.Stat(filepath.Join(outboxDir, "records.jsonl")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("outbox write before durability admission: %v", statErr)
	}
}

type partialWriteDestination struct {
	*scriptedSyncSource
	acknowledgements int
}

func (d *partialWriteDestination) Write(_ context.Context, _ connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{RecordsWritten: len(records) - 1, RecordsFailed: 1}, nil
}

func (d *partialWriteDestination) AcknowledgeETLDurability(_ context.Context, _ string) (synccontract.DownstreamAcknowledgement, error) {
	d.acknowledgements++
	return synccontract.NewDurableDownstreamAcknowledgement(d.Name(), time.Now().UTC())
}
