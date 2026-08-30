package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

func TestDecodeAsanaEventPageRequiresProviderSyncToken(t *testing.T) {
	page, err := decodeAsanaEventPage(map[string]any{
		"data": []any{map[string]any{
			"action":   " DELETED ",
			"resource": map[string]any{"gid": " 123 ", "resource_type": " Task "},
		}},
		"sync":     " next-token ",
		"has_more": true,
	})
	if err != nil {
		t.Fatalf("decodeAsanaEventPage() error = %v", err)
	}
	if page.Sync != "next-token" || !page.HasMore || len(page.Data) != 1 {
		t.Fatalf("decoded page = %+v", page)
	}
	if got := page.Data[0]; got.Action != "deleted" || got.Resource.GID != "123" || got.Resource.ResourceType != "task" {
		t.Fatalf("normalized event = %+v", got)
	}

	if _, err := decodeAsanaEventPage(map[string]any{"data": []any{}}); err == nil {
		t.Fatal("decodeAsanaEventPage() accepted a response without provider sync state")
	}
}

func TestAsanaEventTombstoneUsesDeletedButNeverRemoved(t *testing.T) {
	position := synccontract.CheckpointPosition{
		Primary:    synccontract.OpaqueToken("next-token"),
		TieBreaker: synccontract.OpaqueToken("window-digest"),
	}
	deleted := asanaEvent{Action: "deleted", CreatedAt: "2026-08-28T10:00:00Z", Resource: asanaResource{GID: "123", ResourceType: "task"}}
	tombstone, ok, err := asanaEventTombstone(deleted, "gid", position)
	if err != nil {
		t.Fatalf("asanaEventTombstone(deleted) error = %v", err)
	}
	if !ok {
		t.Fatal("asanaEventTombstone(deleted) did not emit a tombstone")
	}
	if got, want := string(tombstone.Key), `{"gid":"123"}`; got != want {
		t.Fatalf("tombstone key = %s, want %s", got, want)
	}
	if !bytes.Equal(tombstone.Position.Primary, position.Primary) || !bytes.Equal(tombstone.Position.TieBreaker, position.TieBreaker) {
		t.Fatalf("tombstone position = %+v, want %+v", tombstone.Position, position)
	}

	removed := deleted
	removed.Action = "removed"
	if _, ok, err := asanaEventTombstone(removed, "gid", position); err != nil || ok {
		t.Fatalf("asanaEventTombstone(removed) = ok %t, err %v; relationship removal must not become a delete", ok, err)
	}
}

func TestAsanaEventTombstoneIdentityIsStableWithinTokenWindow(t *testing.T) {
	position := synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("token"), TieBreaker: synccontract.OpaqueToken("digest")}
	event := asanaEvent{Action: "deleted", CreatedAt: "2026-08-28T10:00:00Z", Resource: asanaResource{GID: "123", ResourceType: "task"}}
	first, ok, err := asanaEventTombstone(event, "gid", position)
	if err != nil || !ok {
		t.Fatalf("first tombstone = ok %t, err %v", ok, err)
	}
	second, ok, err := asanaEventTombstone(event, "gid", position)
	if err != nil || !ok {
		t.Fatalf("second tombstone = ok %t, err %v", ok, err)
	}
	if !bytes.Equal(first.EventID, second.EventID) {
		t.Fatalf("event identity changed across replay: %q != %q", first.EventID, second.EventID)
	}
}

func TestAsanaEventCurrentStateIsIndependentOfProviderEventOrder(t *testing.T) {
	request := asanaEventTestRequest(nil)
	events := []asanaEvent{
		{Action: "changed", CreatedAt: "2026-08-28T10:00:02Z", Resource: asanaResource{GID: "task-1", ResourceType: "task"}},
		{Action: "deleted", CreatedAt: "2026-08-28T10:00:01Z", Resource: asanaResource{GID: "task-2", ResourceType: "task"}},
		{Action: "added", CreatedAt: "2026-08-28T10:00:00Z", Resource: asanaResource{GID: "task-1", ResourceType: "task"}},
	}
	reversed := append([]asanaEvent(nil), events...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	firstCheckpoint, err := asanaEventCheckpoint(request.Resume, request.Stream, "start-token", "final-token", events)
	if err != nil {
		t.Fatal(err)
	}
	secondCheckpoint, err := asanaEventCheckpoint(request.Resume, request.Stream, "start-token", "final-token", reversed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstCheckpoint.Position.TieBreaker, secondCheckpoint.Position.TieBreaker) || !bytes.Equal(firstCheckpoint.Dedupe.Value, secondCheckpoint.Dedupe.Value) {
		t.Fatalf("token-window identity depends on provider event order: first=%+v second=%+v", firstCheckpoint, secondCheckpoint)
	}

	connector := &asanaEventTestConnector{operation: func(request connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
		if request.Operation != "get_task" {
			return connectors.DirectReadResult{}, fmt.Errorf("unexpected operation %q", request.Operation)
		}
		if request.PathParams["task_gid"] == "task-1" {
			return connectors.DirectReadResult{Status: 200, Body: map[string]any{"data": map[string]any{"gid": "task-1", "name": "current"}}}, nil
		}
		return connectors.DirectReadResult{Status: 404}, errors.New("provider response status 404")
	}}
	request.Connector = connector
	firstRecords, firstTombstones, err := hydrateAsanaEventWindow(context.Background(), connector, request, asanaEventStreamBindings["tasks"], events, firstCheckpoint.Position)
	if err != nil {
		t.Fatal(err)
	}
	secondRecords, secondTombstones, err := hydrateAsanaEventWindow(context.Background(), connector, request, asanaEventStreamBindings["tasks"], reversed, secondCheckpoint.Position)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(firstRecords, secondRecords) || !reflect.DeepEqual(firstTombstones, secondTombstones) {
		t.Fatalf("coalesced current state depends on event order: first=%+v/%+v second=%+v/%+v", firstRecords, firstTombstones, secondRecords, secondTombstones)
	}
}

func TestAsanaIncrementalAppendWarehouseRowsUseProviderTokenAndPreserveDeleteMarker(t *testing.T) {
	position := synccontract.CheckpointPosition{Primary: synccontract.OpaqueToken("provider-token"), TieBreaker: synccontract.OpaqueToken("window-digest")}
	tombstone, ok, err := asanaEventTombstone(asanaEvent{Action: "deleted", Resource: asanaResource{GID: "task-2", ResourceType: "task"}}, "gid", position)
	if err != nil || !ok {
		t.Fatalf("tombstone = ok %t, err %v", ok, err)
	}
	raw, err := localWarehouseTransportRawRecords(
		synctransport.WarehouseReceipt{ID: "stage-1", Generation: 1},
		synctransport.WarehouseWorkset{
			Records:             []connectors.Record{{"gid": "task-1", "updated_at": "row-cursor"}},
			Tombstones:          []synccontract.Tombstone{tombstone},
			CandidateCheckpoint: synccontract.CheckpointEnvelope{Position: position},
		},
		StreamConfig{CursorField: "updated_at", PrimaryKey: []string{"gid"}},
		connectors.DestinationApplyStrategy{Mode: synccontract.ModeIncrementalAppend, Strategy: connectors.ApplyStrategyAppend},
	)
	if err != nil {
		t.Fatalf("localWarehouseTransportRawRecords() error = %v", err)
	}
	if len(raw) != 2 || raw[0].Cursor != "provider-token" || raw[1].Cursor != "provider-token" {
		t.Fatalf("append raw cursors = %+v, want provider token on record and tombstone", raw)
	}
	if deleted, _ := raw[1].Record["_polymetrics_deleted"].(bool); !raw[1].Deleted || !deleted {
		t.Fatalf("append tombstone row = %+v, want durable changelog delete marker", raw[1])
	}
	if _, found := raw[0].Record["_polymetrics_deleted"]; found {
		t.Fatalf("provider record was annotated by the source path: %+v", raw[0].Record)
	}
}

type asanaEventTestConnector struct {
	operation func(connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error)
	snapshot  []connectors.Record
}

func (*asanaEventTestConnector) Name() string { return "asana" }
func (*asanaEventTestConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "asana", IntegrationType: "api", Capabilities: connectors.Capabilities{Read: true}}
}
func (*asanaEventTestConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }
func (*asanaEventTestConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: "asana", Streams: []connectors.Stream{{Name: "tasks", PrimaryKey: []string{"gid"}}}}, nil
}
func (c *asanaEventTestConnector) Read(_ context.Context, _ connectors.ReadRequest, emit func(connectors.Record) error) error {
	for _, record := range c.snapshot {
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}
func (*asanaEventTestConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
func (c *asanaEventTestConnector) OperationDirectRead(_ context.Context, request connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	return c.operation(request)
}
func (*asanaEventTestConnector) SyncTransportDescriptor() *connectors.SyncTransportDescriptor {
	return &connectors.SyncTransportDescriptor{Source: &connectors.SourceTransportDescriptor{
		Executor:        asanaEventSourceReference,
		EligibleStreams: []string{"tasks"},
		Modes:           []synccontract.Mode{synccontract.ModeIncrementalAppend, synccontract.ModeIncrementalUpsert, synccontract.ModeIncrementalDedupe},
		Delivery: connectors.DeliveryGuarantees{
			Idempotency: connectors.DeliveryIdempotencyKeyed,
			Ordering:    connectors.DeliveryOrderingWindowCoalesced,
			Deletes:     connectors.DeliveryDeletesTombstone,
		},
		Conformance: asanaEventSourceConformance,
	}}
}

func asanaEventTestRequest(connector connectors.Connector) synctransport.SourceRequest {
	return synctransport.SourceRequest{
		Connector: connector,
		Runtime: connectors.RuntimeConfig{
			Config: map[string]string{"project_id": "project-1"},
		},
		Stream:     "tasks",
		Mode:       synccontract.ModeIncrementalDedupe,
		BatchSize:  2,
		PrimaryKey: []string{"gid"},
		Resume: synccontract.ResumeExpectation{
			Source:           synccontract.SourceIdentity{Engine: "asana", AccountOrCluster: "fixture", ObjectScope: "tasks/project-1"},
			SourceGeneration: synccontract.OpaqueToken("generation-1"),
		},
	}
}

func TestAsanaEventSourceBootstrapTokenPrecedesExhaustiveSnapshot(t *testing.T) {
	var calls []string
	connector := &asanaEventTestConnector{
		operation: func(request connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
			calls = append(calls, request.Operation)
			if request.Operation != "get_events" || request.Query["resource"] != "project-1" || request.Query["sync"] != "" {
				t.Fatalf("bootstrap request = %+v", request)
			}
			return connectors.DirectReadResult{Status: 412, Receipt: &connectors.ProviderResponseReceipt{Status: 412, Body: map[string]any{"sync": "bootstrap-token"}}}, errors.New("provider response status 412")
		},
		snapshot: []connectors.Record{
			{"gid": "task-1", "name": "one", "resource_type": "task"},
			{"gid": "task-2", "name": "two", "resource_type": "task"},
			{"gid": "task-3", "name": "three", "resource_type": "task"},
		},
	}
	request := asanaEventTestRequest(connector)
	var pages []synctransport.SourcePage
	err := (&asanaEventSourceExecutor{}).ReadTransport(context.Background(), request, func(page synctransport.SourcePage) error {
		if len(calls) == 1 {
			calls = append(calls, "snapshot")
		}
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadTransport(bootstrap) error = %v", err)
	}
	if len(calls) != 2 || calls[0] != "get_events" || calls[1] != "snapshot" {
		t.Fatalf("provider calls = %v, want token acquisition before snapshot", calls)
	}
	if len(pages) != 2 || len(pages[0].Records) != 2 || len(pages[1].Records) != 1 {
		t.Fatalf("snapshot pages = %+v", pages)
	}
	if !pages[0].DeferCheckpoint || pages[1].DeferCheckpoint {
		t.Fatalf("snapshot checkpoint deferral = [%t %t], want [true false]", pages[0].DeferCheckpoint, pages[1].DeferCheckpoint)
	}
	for index, page := range pages {
		if got := string(page.CandidateCheckpoint.Position.Primary); got != "bootstrap-token" {
			t.Fatalf("page %d checkpoint token = %q", index, got)
		}
	}
}

func TestAsanaEventSourceExpiredTokenRebootstrapsBeforeSnapshot(t *testing.T) {
	connector := &asanaEventTestConnector{
		operation: func(request connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
			if request.Operation != "get_events" || request.Query["sync"] != "expired-token" {
				t.Fatalf("expired-token request = %+v", request)
			}
			return connectors.DirectReadResult{Status: 412, Receipt: &connectors.ProviderResponseReceipt{Status: 412, Body: map[string]any{"sync": "reset-token"}}}, errors.New("provider response status 412")
		},
		snapshot: []connectors.Record{{"gid": "task-1", "name": "current"}},
	}
	request := asanaEventTestRequest(connector)
	checkpoint, err := asanaEventCheckpoint(request.Resume, request.Stream, "expired-token", "expired-token", nil)
	if err != nil {
		t.Fatal(err)
	}
	committedAt := checkpoint.ObservedAt.Add(time.Millisecond)
	checkpoint.CommittedAt = &committedAt
	request.Checkpoint = &checkpoint

	var pages []synctransport.SourcePage
	err = (&asanaEventSourceExecutor{}).ReadTransport(context.Background(), request, func(page synctransport.SourcePage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadTransport(expired token) error = %v", err)
	}
	if len(pages) != 1 || len(pages[0].Records) != 1 || pages[0].DeferCheckpoint {
		t.Fatalf("reset snapshot pages = %+v", pages)
	}
	if got := string(pages[0].CandidateCheckpoint.Position.Primary); got != "reset-token" {
		t.Fatalf("reset checkpoint token = %q, want provider 412 token", got)
	}
}

func TestAsanaEventSourceCoalescesCompleteWindowBeforeHydration(t *testing.T) {
	connector := &asanaEventTestConnector{operation: func(request connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
		switch request.Operation {
		case "get_events":
			switch request.Query["sync"] {
			case "start-token":
				return connectors.DirectReadResult{Status: 200, Body: map[string]any{"data": []any{
					map[string]any{"action": "changed", "resource": map[string]any{"gid": "task-1", "resource_type": "task"}},
					map[string]any{"action": "deleted", "resource": map[string]any{"gid": "task-2", "resource_type": "task"}},
					map[string]any{"action": "removed", "resource": map[string]any{"gid": "task-3", "resource_type": "task"}},
				}, "sync": "middle-token", "has_more": true}}, nil
			case "middle-token":
				return connectors.DirectReadResult{Status: 200, Body: map[string]any{"data": []any{
					map[string]any{"action": "changed", "resource": map[string]any{"gid": "task-1", "resource_type": "task"}},
				}, "sync": "final-token", "has_more": false}}, nil
			default:
				return connectors.DirectReadResult{}, fmt.Errorf("unexpected Events sync token %q", request.Query["sync"])
			}
		case "get_task":
			switch request.PathParams["task_gid"] {
			case "task-1":
				return connectors.DirectReadResult{Status: 200, Body: map[string]any{"data": map[string]any{"gid": "task-1", "name": "current", "resource_type": "task"}}}, nil
			case "task-2", "task-3":
				return connectors.DirectReadResult{Status: 404}, errors.New("provider response status 404")
			}
		default:
			return connectors.DirectReadResult{}, fmt.Errorf("unexpected operation %q", request.Operation)
		}
		return connectors.DirectReadResult{}, errors.New("unreachable operation")
	}}

	request := asanaEventTestRequest(connector)
	checkpoint, err := asanaEventCheckpoint(request.Resume, request.Stream, "start-token", "start-token", nil)
	if err != nil {
		t.Fatalf("asanaEventCheckpoint() error = %v", err)
	}
	committedAt := checkpoint.ObservedAt.Add(time.Millisecond)
	checkpoint.CommittedAt = &committedAt
	request.Checkpoint = &checkpoint

	var pages []synctransport.SourcePage
	err = (&asanaEventSourceExecutor{}).ReadTransport(context.Background(), request, func(page synctransport.SourcePage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("ReadTransport(window) error = %v", err)
	}
	if len(pages) != 1 || len(pages[0].Records) != 1 || len(pages[0].Tombstones) != 1 {
		t.Fatalf("coalesced pages = %+v, want one current task and one delete", pages)
	}
	if got := pages[0].Records[0]["gid"]; got != "task-1" {
		t.Fatalf("hydrated record gid = %v", got)
	}
	if got := string(pages[0].Tombstones[0].Key); got != `{"gid":"task-2"}` {
		t.Fatalf("tombstone key = %s; removed task-3 must not be present", got)
	}
	if got := string(pages[0].CandidateCheckpoint.Position.Primary); got != "final-token" {
		t.Fatalf("committed candidate = %q, want exhausted window token", got)
	}
}

func TestAsanaEventSourceRefusesPartialTokenWindow(t *testing.T) {
	connector := &asanaEventTestConnector{operation: func(request connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
		return connectors.DirectReadResult{Status: 200, Body: map[string]any{"data": []any{}, "sync": request.Query["sync"], "has_more": true}}, nil
	}}
	request := asanaEventTestRequest(connector)
	checkpoint, err := asanaEventCheckpoint(request.Resume, request.Stream, "stalled-token", "stalled-token", nil)
	if err != nil {
		t.Fatal(err)
	}
	committedAt := checkpoint.ObservedAt.Add(time.Millisecond)
	checkpoint.CommittedAt = &committedAt
	request.Checkpoint = &checkpoint
	emitted := 0
	err = (&asanaEventSourceExecutor{}).ReadTransport(context.Background(), request, func(synctransport.SourcePage) error {
		emitted++
		return nil
	})
	if err == nil || emitted != 0 {
		t.Fatalf("partial window = emitted %d, error %v; want pre-emission refusal", emitted, err)
	}
}
