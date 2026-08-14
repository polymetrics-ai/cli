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
	statestore "polymetrics.ai/internal/state"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
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

func TestLegacyStateReloadRetainsSyncModeCompatibilityAfterReversePlanLookup(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("legacy_mode_reload", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	created, connection := setupSyncModeApp(t, source, "incremental_append")

	// This is a pre-contract project state with coordination metadata that is
	// already current. Opening it applies the legacy mode adapter only in
	// memory, which is exactly what a later raw reload used to discard.
	legacy := created.state
	legacy.SyncModeCompatibilityVersion = 0
	for index := range legacy.Connections {
		if legacy.Connections[index].Name != connection {
			continue
		}
		stream := legacy.Connections[index].Streams["records"]
		stream.LegacyCompatibility = false
		legacy.Connections[index].Streams["records"] = stream
	}
	if legacy.CoordinationSalt == "" || len(legacy.CredentialBindings) == 0 {
		t.Fatal("test setup did not retain current credential coordination metadata")
	}
	if err := created.store.Save(legacy); err != nil {
		t.Fatalf("persist legacy-shaped state: %v", err)
	}

	reopened, err := Open(created.root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	reopened.registry = created.registry
	conn, ok := reopened.findConnection(connection)
	if !ok || !conn.Streams["records"].LegacyCompatibility {
		t.Fatalf("Open() did not normalize persisted legacy stream: %#v", conn)
	}

	_, err = reopened.RunReverseETL(ctx, RunReverseETLRequest{PlanID: "unknown-plan"})
	if err == nil || !strings.Contains(err.Error(), `reverse plan "unknown-plan" not found`) {
		t.Fatalf("RunReverseETL(unknown plan) error = %v, want missing-plan error", err)
	}
	if _, err := reopened.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("RunETL() after unknown reverse plan error = %v", err)
	}
}

func TestStreamReadStateDistinguishesAbsentAndExplicitEmptyCursor(t *testing.T) {
	initial := streamReadState(StreamState{}, 7)
	if initial["generation_id"] != "7" {
		t.Fatalf("initial read state = %#v, want generation", initial)
	}
	if _, present := initial["cursor"]; present {
		t.Fatalf("initial read state = %#v, want no cursor", initial)
	}

	positionUnobserved := false
	unobserved := streamReadState(StreamState{Checkpoint: &synccontract.CheckpointEnvelope{
		PositionObserved: &positionUnobserved,
	}}, 8)
	if _, present := unobserved["cursor"]; present {
		t.Fatalf("unobserved checkpoint read state = %#v, want no cursor", unobserved)
	}

	positionObserved := true
	empty := streamReadState(StreamState{Checkpoint: &synccontract.CheckpointEnvelope{
		Position:         synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken{}},
		PositionObserved: &positionObserved,
	}}, 9)
	if cursor, present := empty["cursor"]; !present || cursor != "" {
		t.Fatalf("empty checkpoint read state = %#v, want explicit empty cursor", empty)
	}

	whitespace := streamReadState(StreamState{Checkpoint: &synccontract.CheckpointEnvelope{
		Position:         synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("  ")},
		PositionObserved: &positionObserved,
	}}, 10)
	if cursor := whitespace["cursor"]; cursor != "  " {
		t.Fatalf("whitespace checkpoint cursor = %q, want preserved value", cursor)
	}
}

func TestIncrementalRunKeepsNoPositionDistinctFromObservedEmptyCursor(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("empty_cursor_state", nil)
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	stateKey := streamStateKey(connection, "records")

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("RunETL(empty): %v", err)
	}
	first := a.state.StreamStates[stateKey]
	if first.Checkpoint == nil || first.Checkpoint.PositionObserved == nil || *first.Checkpoint.PositionObserved {
		t.Fatalf("empty run checkpoint = %#v, want an explicitly unobserved position", first.Checkpoint)
	}
	if _, present := source.requests[0].State["cursor"]; present {
		t.Fatalf("initial source state = %#v, want no cursor", source.requests[0].State)
	}

	source.records = []connectors.Record{{"id": "empty", "updated_at": ""}}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("RunETL(observed empty cursor): %v", err)
	}
	if _, present := source.requests[1].State["cursor"]; present {
		t.Fatalf("source state after an unobserved checkpoint = %#v, want no cursor", source.requests[1].State)
	}
	second := a.state.StreamStates[stateKey]
	if second.Checkpoint == nil || second.Checkpoint.PositionObserved == nil || !*second.Checkpoint.PositionObserved {
		t.Fatalf("empty cursor checkpoint = %#v, want an observed position", second.Checkpoint)
	}
	if got := string(second.Checkpoint.Position.Primary); got != "" {
		t.Fatalf("empty cursor checkpoint = %q, want preserved empty value", got)
	}

	encoded, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	var reloaded StreamState
	if err := json.Unmarshal(encoded, &reloaded); err != nil {
		t.Fatal(err)
	}
	a.state.StreamStates[stateKey] = reloaded
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("RunETL(resumed empty cursor): %v", err)
	}
	if cursor, present := source.requests[2].State["cursor"]; !present || cursor != "" {
		t.Fatalf("source state after an observed empty cursor = %#v, want explicit empty cursor", source.requests[2].State)
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
	if len(source.requests) != 1 {
		t.Fatalf("initial source requests = %#v, want one", source.requests)
	}
	if _, present := source.requests[0].State["cursor"]; present {
		t.Fatalf("initial source state = %#v, want no cursor", source.requests[0].State)
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

func TestSourceOrderedOpaqueCursorResumesWithoutLossOrReplayAcrossDestinations(t *testing.T) {
	for _, destinationKind := range []string{"warehouse", "connector"} {
		t.Run(destinationKind, func(t *testing.T) {
			ctx := context.Background()
			source := newOrderedOpaqueCursorSource("ordered_opaque_cursor_"+destinationKind, []connectors.Record{
				{"id": "first", "updated_at": []byte{0x00}},
				{"id": "second", "updated_at": []byte{0xff}},
			})
			a, connection := setupSyncModeApp(t, source, "incremental_append")
			stateKey := streamStateKey(connection, "records")

			var connectorDestination *batchDestination
			run := func(runID string) error {
				if destinationKind == "warehouse" {
					_, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10})
					return err
				}
				conn, ok := a.findConnection(connection)
				if !ok {
					return errors.New("connection missing")
				}
				mode, err := ParseStreamSyncMode(conn.Streams["records"])
				if err != nil {
					return err
				}
				resolved, credential, runtime, err := a.resolveEndpointWithCredential(ctx, conn.Source)
				if err != nil {
					return err
				}
				if connectorDestination == nil {
					connectorDestination = &batchDestination{}
				}
				expectation := streamResumeExpectation(resolved, credential, runtime, "records")
				result, err := a.runConnectorETL(ctx, runID, conn, resolved, runtime, connectorDestination, connectors.RuntimeConfig{}, expectation, "records", conn.Streams["records"], mode, 10)
				if err != nil {
					return err
				}
				if result.PendingStreamState == nil {
					return errors.New("connector ETL did not return pending stream state")
				}
				a.state.StreamStates[stateKey] = result.PendingStreamState.State
				return nil
			}

			if err := run("opaque_first"); err != nil {
				t.Fatalf("first ETL run: %v", err)
			}
			first := a.state.StreamStates[stateKey]
			if first.Checkpoint == nil || !bytes.Equal(first.Checkpoint.Position.Primary, []byte{0xff}) {
				t.Fatalf("first checkpoint cursor = %#v, want exact binary 0xff", first.Checkpoint)
			}
			encoded, err := json.Marshal(first)
			if err != nil {
				t.Fatal(err)
			}
			var reloaded StreamState
			if err := json.Unmarshal(encoded, &reloaded); err != nil {
				t.Fatal(err)
			}
			a.state.StreamStates[stateKey] = reloaded

			source.records = append(source.records, connectors.Record{"id": "third", "updated_at": []byte{0xff, 0x00}})
			if err := run("opaque_second"); err != nil {
				t.Fatalf("second ETL run: %v", err)
			}
			if len(source.requests) != 2 {
				t.Fatalf("source requests = %d, want 2", len(source.requests))
			}
			resumed := source.requests[1]
			if !resumed.CursorState.Present || !bytes.Equal(resumed.CursorState.Token, []byte{0xff}) {
				t.Fatalf("resumed opaque cursor = %#v, want exact binary 0xff", resumed.CursorState)
			}
			if _, present := resumed.Config.Config["since"]; present {
				t.Fatalf("source-ordered resume config = %#v, want no reconstructed lower bound", resumed.Config.Config)
			}
			if got := strings.Join(source.emitted, ","); got != "first,second,third" {
				t.Fatalf("source emitted = %q, want no replay and no skipped third row", got)
			}
			final := a.state.StreamStates[stateKey]
			if final.Checkpoint == nil || !bytes.Equal(final.Checkpoint.Position.Primary, []byte{0xff, 0x00}) {
				t.Fatalf("final checkpoint cursor = %#v, want exact binary 0xff00", final.Checkpoint)
			}
			if destinationKind == "warehouse" {
				rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
				if err != nil {
					t.Fatal(err)
				}
				if len(rows) != 3 {
					t.Fatalf("warehouse rows = %d, want 3", len(rows))
				}
			} else if !reflect.DeepEqual(connectorDestination.batches, []int{2, 1}) {
				t.Fatalf("connector batches = %#v, want [2 1]", connectorDestination.batches)
			}
		})
	}
}

func TestSourceOrderedCursorFieldMismatchRefusesBothDestinationPaths(t *testing.T) {
	ctx := context.Background()
	source := newBoundOrderedOpaqueCursorSource("bound_opaque_cursor", []connectors.Record{{
		"id": "first", "updated_at": []byte{0x01},
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	for i := range a.state.Connections {
		if a.state.Connections[i].Name == connection {
			a.state.Connections[i].Source.Config = map[string]string{"cursor_field": "sequence"}
		}
	}

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err == nil || !strings.Contains(err.Error(), "cursor field") {
		t.Fatalf("RunETL() error = %v, want cursor-field mismatch", err)
	}
	if len(source.requests) != 0 {
		t.Fatalf("warehouse path read requests = %#v, want none", source.requests)
	}

	conn, ok := a.findConnection(connection)
	if !ok {
		t.Fatal("connection missing")
	}
	mode, err := ParseStreamSyncMode(conn.Streams["records"])
	if err != nil {
		t.Fatal(err)
	}
	resolved, credential, runtime, err := a.resolveEndpointWithCredential(ctx, conn.Source)
	if err != nil {
		t.Fatal(err)
	}
	expectation := streamResumeExpectation(resolved, credential, runtime, "records")
	_, err = a.runConnectorETL(ctx, "mismatch", conn, resolved, runtime, &batchDestination{}, connectors.RuntimeConfig{}, expectation, "records", conn.Streams["records"], mode, 1)
	if err == nil || !strings.Contains(err.Error(), "cursor field") {
		t.Fatalf("runConnectorETL() error = %v, want cursor-field mismatch", err)
	}
	if len(source.requests) != 0 {
		t.Fatalf("connector path read requests = %#v, want none", source.requests)
	}
}

func TestSourceOrderedFullRefreshStartsWithoutResumeAcrossDestinations(t *testing.T) {
	for _, destinationKind := range []string{"warehouse", "connector"} {
		t.Run(destinationKind, func(t *testing.T) {
			ctx := context.Background()
			source := newOrderedOpaqueCursorSource("full_refresh_opaque_"+destinationKind, []connectors.Record{{
				"id": "old", "name": "old", "updated_at": []byte{0xff},
			}})
			a, connection := setupSyncModeApp(t, source, "full_refresh_overwrite")
			stateKey := streamStateKey(connection, "records")
			destination := &batchDestination{}
			run := func(runID string) error {
				if destinationKind == "warehouse" {
					_, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10})
					return err
				}
				conn, ok := a.findConnection(connection)
				if !ok {
					return errors.New("connection missing")
				}
				mode, err := ParseStreamSyncMode(conn.Streams["records"])
				if err != nil {
					return err
				}
				resolved, credential, runtime, err := a.resolveEndpointWithCredential(ctx, conn.Source)
				if err != nil {
					return err
				}
				expectation := streamResumeExpectation(resolved, credential, runtime, "records")
				result, err := a.runConnectorETL(ctx, runID, conn, resolved, runtime, destination, connectors.RuntimeConfig{}, expectation, "records", conn.Streams["records"], mode, 10)
				if err != nil {
					return err
				}
				if result.PendingStreamState == nil {
					return errors.New("connector ETL did not return pending stream state")
				}
				a.state.StreamStates[stateKey] = result.PendingStreamState.State
				return nil
			}

			if err := run("full_first"); err != nil {
				t.Fatalf("first ETL run: %v", err)
			}
			source.records = []connectors.Record{{"id": "fresh", "name": "fresh", "updated_at": []byte{0x00}}}
			if err := run("full_second"); err != nil {
				t.Fatalf("second ETL run: %v", err)
			}
			if len(source.requests) != 2 {
				t.Fatalf("source requests = %d, want 2", len(source.requests))
			}
			second := source.requests[1]
			if second.CursorState.Present {
				t.Fatalf("full refresh cursor state = %#v, want absent", second.CursorState)
			}
			if _, present := second.State["cursor"]; present {
				t.Fatalf("full refresh state = %#v, want no cursor", second.State)
			}
			if _, present := second.Config.Config["since"]; present {
				t.Fatalf("full refresh config = %#v, want no lower bound", second.Config.Config)
			}
			if got := strings.Join(source.emitted, ","); got != "old,fresh" {
				t.Fatalf("source emitted = %q, want old,fresh", got)
			}
			state := a.state.StreamStates[stateKey]
			if state.Checkpoint == nil || !bytes.Equal(state.Checkpoint.Position.Primary, []byte{0x00}) {
				t.Fatalf("full refresh checkpoint = %#v, want exact binary 0x00", state.Checkpoint)
			}
			if destinationKind == "warehouse" {
				rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
				if err != nil {
					t.Fatal(err)
				}
				byID := rowsByID(rows)
				if len(byID) != 1 || byID["fresh"]["name"] != "fresh" {
					t.Fatalf("full refresh warehouse rows = %#v, want fresh only", rows)
				}
			} else if !reflect.DeepEqual(destination.batches, []int{1, 1}) {
				t.Fatalf("connector batches = %#v, want [1 1]", destination.batches)
			}
		})
	}
}

func TestSourceOrderedBinaryCursorResumesIncrementalAppend(t *testing.T) {
	ctx := context.Background()
	source := newOrderedOpaqueCursorSource("dedupe_opaque_cursor", []connectors.Record{
		{"id": "same", "name": "old", "updated_at": []byte{0x00}},
		{"id": "same", "name": "latest", "updated_at": []byte{0xff}},
	})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	stateKey := streamStateKey(connection, "records")
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10}); err != nil {
		t.Fatalf("first RunETL(): %v", err)
	}
	rows, err := a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	byID := rowsByID(rows)
	if len(rows) != 2 || len(byID) != 1 || byID["same"]["name"] != "latest" {
		t.Fatalf("first incremental rows = %#v, want both source records and latest map value", rows)
	}
	state := a.state.StreamStates[stateKey]
	if state.Checkpoint == nil || !bytes.Equal(state.Checkpoint.Position.Primary, []byte{0xff}) {
		t.Fatalf("first checkpoint = %#v, want exact binary 0xff", state.Checkpoint)
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	var reloaded StreamState
	if err := json.Unmarshal(encoded, &reloaded); err != nil {
		t.Fatal(err)
	}
	a.state.StreamStates[stateKey] = reloaded

	source.records = append(source.records, connectors.Record{"id": "later", "name": "later", "updated_at": []byte{0xff, 0x00}})
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 10}); err != nil {
		t.Fatalf("second RunETL(): %v", err)
	}
	if len(source.requests) != 2 {
		t.Fatalf("source requests = %d, want 2", len(source.requests))
	}
	if resumed := source.requests[1].CursorState; !resumed.Present || !bytes.Equal(resumed.Token, []byte{0xff}) {
		t.Fatalf("resumed cursor = %#v, want exact binary 0xff", resumed)
	}
	if got := strings.Join(source.emitted, ","); got != "same,same,later" {
		t.Fatalf("source emitted = %q, want no replay", got)
	}
	rows, err = a.QueryTable(ctx, QueryTableRequest{Table: "records", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	byID = rowsByID(rows)
	if len(rows) != 3 || len(byID) != 2 || byID["same"]["name"] != "latest" || byID["later"]["name"] != "later" {
		t.Fatalf("second incremental rows = %#v, want appended latest and later records", rows)
	}
	state = a.state.StreamStates[stateKey]
	if state.Checkpoint == nil || !bytes.Equal(state.Checkpoint.Position.Primary, []byte{0xff, 0x00}) {
		t.Fatalf("final checkpoint = %#v, want exact binary 0xff00", state.Checkpoint)
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
	destination, destinationRuntime, err := a.resolveEndpoint(ctx, conn.Destination)
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
	result, err := a.runWarehouseETL(ctx, run.ID, conn, sourceConnector, sourceRuntime, destination, destinationRuntime, expectation, "records", conn.Streams["records"], mode, 1)
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

func TestSyncLocalWarehouseDirectoryUsesDurableCommit(t *testing.T) {
	if err := syncLocalWarehouseDirectory(t.TempDir()); err != nil {
		t.Fatalf("syncLocalWarehouseDirectory() error = %v", err)
	}
}

func TestSyncLocalWarehouseDirectoryReportsCommitFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing")
	if err := syncLocalWarehouseDirectory(path); err == nil {
		t.Fatal("syncLocalWarehouseDirectory() error = nil, want missing directory error")
	}
}

func expectedLocalWarehouseDirectorySyncChain(t *testing.T, dir string) []string {
	t.Helper()
	path, err := filepath.Abs(dir)
	if err != nil {
		t.Fatalf("resolve warehouse directory: %v", err)
	}
	chain := make([]string, 0, 8)
	for {
		chain = append(chain, filepath.Clean(path))
		parent := filepath.Dir(path)
		if parent == path {
			return chain
		}
		path = parent
	}
}

func mustFindConnection(t *testing.T, a *App, name string) Connection {
	t.Helper()
	conn, ok := a.findConnection(name)
	if !ok {
		t.Fatalf("connection %q not found", name)
	}
	return conn
}

// expectedWarehouseRunSyncChains is the full sequence a warehouse run emits:
// the tables directory chain, then the wal directory chain. Both walk to the
// filesystem root, so every ancestor a materialized row depends on — including
// the connection directory holding its ownership record — is durable before
// the checkpoint is acknowledged.
func expectedWarehouseRunSyncChains(t *testing.T, a *App, connection, warehouseDir string) []string {
	t.Helper()
	location, err := a.warehouseLocation(warehouseDir, mustFindConnection(t, a, connection))
	if err != nil {
		t.Fatalf("resolve warehouse location: %v", err)
	}
	chains := expectedLocalWarehouseDirectorySyncChain(t, filepath.Join(location.ConnectionDir, warehouse.TablesDirName))
	return append(chains, expectedLocalWarehouseDirectorySyncChain(t, filepath.Join(location.ConnectionDir, warehouse.WALDirName))...)
}

func setLocalWarehousePath(t *testing.T, a *App, path string) {
	t.Helper()
	for index := range a.state.Credentials {
		if a.state.Credentials[index].Name != "warehouse" {
			continue
		}
		a.state.Credentials[index].Config = map[string]string{
			"path":                path,
			"allow_external_path": "true",
		}
		return
	}
	t.Fatal("warehouse credential not found")
}

func TestRunWarehouseETLSyncsNewDirectoryParentChainBeforeAcknowledgement(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("warehouse_directory_parent_chain", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	warehouseDir := filepath.Join(t.TempDir(), "new", "warehouse", "root")
	if _, err := os.Stat(filepath.Dir(warehouseDir)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("test setup warehouse parent stat error = %v, want missing parent", err)
	}
	setLocalWarehousePath(t, a, warehouseDir)

	originalSync := syncLocalWarehouseDirectoryCommit
	var synced []string
	syncLocalWarehouseDirectoryCommit = func(dir string) error {
		synced = append(synced, filepath.Clean(dir))
		return nil
	}
	t.Cleanup(func() { syncLocalWarehouseDirectoryCommit = originalSync })

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("RunETL() error = %v", err)
	}
	if state := a.state.StreamStates[streamStateKey(connection, "records")]; state.Checkpoint == nil {
		t.Fatal("successful run did not acknowledge a checkpoint")
	}

	wantSyncs := expectedWarehouseRunSyncChains(t, a, connection, warehouseDir)
	if !reflect.DeepEqual(synced, wantSyncs) {
		t.Fatalf("warehouse directory sync calls = %#v, want %#v", synced, wantSyncs)
	}
}

func TestRunWarehouseETLSyncsExternalWarehouseChainForEachWriter(t *testing.T) {
	ctx := context.Background()
	warehouseDir := filepath.Join(t.TempDir(), "shared", "warehouse")
	sourceA := newScriptedSyncSource("warehouse_directory_chain_a", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connectionA := setupSyncModeApp(t, sourceA, "incremental_append")
	sourceB := newScriptedSyncSource("warehouse_directory_chain_b", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	b, connectionB := setupSyncModeApp(t, sourceB, "incremental_append")
	setLocalWarehousePath(t, a, warehouseDir)
	setLocalWarehousePath(t, b, warehouseDir)

	originalSync := syncLocalWarehouseDirectoryCommit
	var synced []string
	syncLocalWarehouseDirectoryCommit = func(dir string) error {
		synced = append(synced, filepath.Clean(dir))
		return nil
	}
	t.Cleanup(func() {
		syncLocalWarehouseDirectoryCommit = originalSync
	})

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connectionA, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("first RunETL() error = %v", err)
	}
	wantSyncsA := expectedWarehouseRunSyncChains(t, a, connectionA, warehouseDir)
	if !reflect.DeepEqual(synced, wantSyncsA) {
		t.Fatalf("first warehouse directory sync calls = %#v, want %#v", synced, wantSyncsA)
	}
	locationA, err := a.warehouseLocation(warehouseDir, mustFindConnection(t, a, connectionA))
	if err != nil {
		t.Fatalf("resolve first warehouse location: %v", err)
	}
	if _, err := os.Stat(filepath.Join(locationA.ConnectionDir, warehouse.WALDirName)); err != nil {
		t.Fatalf("first writer did not create wal directory: %v", err)
	}

	// The second writer shares the warehouse root but owns a different
	// directory, so it syncs its own chain rather than reusing the first
	// writer's.
	synced = nil
	if _, err := b.RunETL(ctx, RunETLRequest{Connection: connectionB, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("second RunETL() error = %v", err)
	}
	wantSyncsB := expectedWarehouseRunSyncChains(t, b, connectionB, warehouseDir)
	if !reflect.DeepEqual(synced, wantSyncsB) {
		t.Fatalf("second warehouse directory sync calls = %#v, want %#v", synced, wantSyncsB)
	}
	if state := b.state.StreamStates[streamStateKey(connectionB, "records")]; state.Checkpoint == nil {
		t.Fatal("second writer did not acknowledge a checkpoint")
	}
}

func TestRunWarehouseETLResyncsDirectoryChainAfterSyncFailure(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("warehouse_directory_retry", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	warehouseDir := filepath.Join(t.TempDir(), "new", "warehouse", "root")
	setLocalWarehousePath(t, a, warehouseDir)
	wantSyncs := expectedWarehouseRunSyncChains(t, a, connection, warehouseDir)
	failingDir := filepath.Clean(filepath.Dir(warehouseDir))
	failureIndex := -1
	for index, dir := range wantSyncs {
		if dir == failingDir {
			failureIndex = index
			break
		}
	}
	if failureIndex < 0 {
		t.Fatalf("sync chain %#v does not include %q", wantSyncs, failingDir)
	}

	originalSync := syncLocalWarehouseDirectoryCommit
	var failedSyncs []string
	syncLocalWarehouseDirectoryCommit = func(dir string) error {
		dir = filepath.Clean(dir)
		failedSyncs = append(failedSyncs, dir)
		if dir == failingDir {
			return errors.New("injected warehouse directory sync failure")
		}
		return nil
	}
	t.Cleanup(func() { syncLocalWarehouseDirectoryCommit = originalSync })

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err == nil {
		t.Fatal("RunETL() error = nil after directory sync failure")
	}
	if !reflect.DeepEqual(failedSyncs, wantSyncs[:failureIndex+1]) {
		t.Fatalf("failed warehouse directory sync calls = %#v, want %#v", failedSyncs, wantSyncs[:failureIndex+1])
	}
	if state := a.state.StreamStates[streamStateKey(connection, "records")]; state.Checkpoint != nil {
		t.Fatalf("checkpoint acknowledged after directory sync failure: %#v", state.Checkpoint)
	}
	location, err := a.warehouseLocation(warehouseDir, mustFindConnection(t, a, connection))
	if err != nil {
		t.Fatalf("resolve warehouse location: %v", err)
	}
	if _, err := os.Stat(filepath.Join(location.ConnectionDir, warehouse.WALDirName)); err != nil {
		t.Fatalf("failed run did not leave directory setup for retry: %v", err)
	}

	var retriedSyncs []string
	syncLocalWarehouseDirectoryCommit = func(dir string) error {
		retriedSyncs = append(retriedSyncs, filepath.Clean(dir))
		return nil
	}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("retry RunETL() error = %v", err)
	}
	if !reflect.DeepEqual(retriedSyncs, wantSyncs) {
		t.Fatalf("retry warehouse directory sync calls = %#v, want %#v", retriedSyncs, wantSyncs)
	}
	if state := a.state.StreamStates[streamStateKey(connection, "records")]; state.Checkpoint == nil {
		t.Fatal("retry did not acknowledge a checkpoint")
	}
}

func TestRunETLPublishesCommittedCheckpointAfterStateUnlockFailure(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("post_commit_unlock_failure", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	a.store.Locker = &postCommitUnlockFailureLocker{failAt: 2}

	_, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
	var outcome *statestore.CommitOutcomeError
	if !errors.As(err, &outcome) {
		t.Fatalf("RunETL() error = %T %v, want CommitOutcomeError", err, err)
	}
	if outcome.Outcome != statestore.CommitOutcomeCommitted {
		t.Fatalf("commit outcome = %q, want committed", outcome.Outcome)
	}
	stateKey := streamStateKey(connection, "records")
	memoryState := a.state.StreamStates[stateKey]
	if memoryState.Checkpoint == nil {
		t.Fatal("in-memory checkpoint was not published after committed state save")
	}

	reopened, err := Open(a.root)
	if err != nil {
		t.Fatal(err)
	}
	diskState := reopened.state.StreamStates[stateKey]
	if diskState.Checkpoint == nil || !reflect.DeepEqual(*diskState.Checkpoint, *memoryState.Checkpoint) {
		t.Fatalf("persisted checkpoint = %#v, want %#v", diskState.Checkpoint, memoryState.Checkpoint)
	}

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("RunETL() after committed state save error = %v", err)
	}
	if len(source.requests) != 2 {
		t.Fatalf("source reads = %d, want 2", len(source.requests))
	}
	if got := source.requests[1].State["cursor"]; got != "2026-08-06T00:00:00Z" {
		t.Fatalf("resumed cursor = %q, want committed checkpoint cursor", got)
	}
}

func TestWarehouseETLUsesResolvedDestinationName(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("credential_resolved_warehouse", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	for i := range a.state.Connections {
		if a.state.Connections[i].Name != connection {
			continue
		}
		destination := a.state.Connections[i].Destination
		destination.Connector = ""
		a.state.Connections[i].Destination = destination
	}

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("RunETL() error = %v", err)
	}
	if state := a.state.StreamStates[streamStateKey(connection, "records")]; state.Checkpoint == nil {
		t.Fatal("warehouse run did not commit a checkpoint")
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

func TestStreamStateAllowsRefreshRotationButRejectsSourceConfigChange(t *testing.T) {
	ctx := context.Background()
	source := newScriptedSyncSource("refresh_rotation_resume", []connectors.Record{{
		"id": "1", "updated_at": "2026-08-06T00:00:00Z",
	}})
	a, connection := setupSyncModeApp(t, source, "incremental_append")
	credential, ok := a.findCredential("source")
	if !ok {
		t.Fatal("source credential missing")
	}
	if err := a.vault.Put(ctx, credential.ID, map[string]string{"refresh_token": "before"}); err != nil {
		t.Fatal(err)
	}
	rotated := false
	source.onRead = func(ctx context.Context, req connectors.ReadRequest) error {
		if rotated {
			return nil
		}
		rotated = true
		if req.Config.SecretStore == nil {
			return errors.New("source secret store is unavailable")
		}
		return req.Config.SecretStore.PutSecret(ctx, "refresh_token", "after")
	}

	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1}); err != nil {
		t.Fatalf("RunETL() after refresh rotation error = %v", err)
	}
	if len(source.requests) != 2 {
		t.Fatalf("source reads after rotation = %d, want 2", len(source.requests))
	}
	if got := source.requests[1].State["cursor"]; got != "2026-08-06T00:00:00Z" {
		t.Fatalf("resumed cursor after rotation = %q, want prior cursor", got)
	}

	for i := range a.state.Connections {
		if a.state.Connections[i].Name != connection {
			continue
		}
		sourceConfig := a.state.Connections[i].Source
		sourceConfig.Config = map[string]string{"account": "changed"}
		a.state.Connections[i].Source = sourceConfig
	}
	readCount := len(source.requests)
	_, err := a.RunETL(ctx, RunETLRequest{Connection: connection, Stream: "records", BatchSize: 1})
	var recovery *synccontract.RebootstrapRequiredError
	if !errors.As(err, &recovery) || recovery.Outcome != synccontract.RecoveryOutcomeSourceGenerationChanged {
		t.Fatalf("RunETL() error = %T %v, want source generation rebootstrap", err, err)
	}
	if len(source.requests) != readCount {
		t.Fatalf("source read after stable source configuration changed: %#v", source.requests)
	}
}

func TestConnectorETLRequiresCompleteDestinationWritesBeforeDurabilityAcknowledgement(t *testing.T) {
	tests := []struct {
		name        string
		mode        string
		records     []connectors.Record
		batchSize   int
		writeResult connectors.WriteResult
	}{
		{
			name:      "short batch without reported failures",
			mode:      "incremental_append",
			batchSize: 2,
			records: []connectors.Record{
				{"id": "1", "updated_at": "2026-08-06T00:00:00Z"},
				{"id": "2", "updated_at": "2026-08-07T00:00:00Z"},
			},
			writeResult: connectors.WriteResult{RecordsWritten: 1},
		},
		{
			name:        "failed empty overwrite setup",
			mode:        "full_refresh_overwrite",
			batchSize:   1,
			writeResult: connectors.WriteResult{RecordsFailed: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			source := newScriptedSyncSource("incomplete_write_source", tt.records)
			a, connection := setupSyncModeApp(t, source, tt.mode)
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
			destination := &incompleteWriteDestination{
				scriptedSyncSource: newScriptedSyncSource("incomplete_write_destination", nil),
				result:             tt.writeResult,
			}
			result, err := a.runConnectorETL(ctx, "run_incomplete", conn, source, sourceRuntime, destination, connectors.RuntimeConfig{}, expectation, "records", conn.Streams["records"], mode, tt.batchSize)
			if err == nil {
				t.Fatal("runConnectorETL() error = nil after incomplete destination result")
			}
			if destination.acknowledgements != 0 {
				t.Fatalf("durability acknowledgement count = %d, want 0 after incomplete write", destination.acknowledgements)
			}
			if result.PendingStreamState != nil || result.Checkpoint != nil {
				t.Fatalf("checkpoint candidate = %#v, want none after incomplete write", result.PendingStreamState)
			}
		})
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

type incompleteWriteDestination struct {
	*scriptedSyncSource
	result           connectors.WriteResult
	acknowledgements int
}

func (d *incompleteWriteDestination) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return d.result, nil
}

func (d *incompleteWriteDestination) AcknowledgeETLDurability(_ context.Context, _ string) (synccontract.DownstreamAcknowledgement, error) {
	d.acknowledgements++
	return synccontract.NewDurableDownstreamAcknowledgement(d.Name(), time.Now().UTC())
}

type postCommitUnlockFailureLocker struct {
	unlocks int
	failAt  int
}

func (l *postCommitUnlockFailureLocker) Lock() (func() error, error) {
	return func() error {
		l.unlocks++
		if l.unlocks == l.failAt {
			return errors.New("unlock failed")
		}
		return nil
	}, nil
}
