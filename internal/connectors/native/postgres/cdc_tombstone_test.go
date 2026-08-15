package postgres

import (
	"encoding/json"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

func TestCDCDeleteTombstoneConvertsPGOutputDeleteToExplicitSourceKeys(t *testing.T) {
	decoder := newPGOutputDecoder()
	if _, err := decoder.decode(relationMessage(testRelationID, "public", "users", testColumn{name: "tenant", typeID: 25}, testColumn{name: "id", typeID: 23}, testColumn{name: "value", typeID: 25}), ""); err != nil {
		t.Fatalf("decode relation: %v", err)
	}
	events, err := decoder.decode(deleteMessage(testRelationID, 'K', textField("retain"), textField("9")), "0/16B6C50")
	if err != nil || len(events) != 1 {
		t.Fatalf("decode delete = (%#v, %v), want one pgoutput delete", events, err)
	}
	tombstone, err := CDCDeleteTombstone(events[0], []string{"tenant", "id"})
	if err != nil {
		t.Fatalf("CDCDeleteTombstone() error = %v", err)
	}
	if err := tombstone.Validate(); err != nil {
		t.Fatalf("CDCDeleteTombstone() returned invalid tombstone: %v", err)
	}
	if tombstone.Operation != synccontract.OperationDelete || tombstone.DeleteImage != synccontract.DeleteImageKeyOnly || string(tombstone.Position.Primary) != "0/16B6C50" || len(tombstone.EventID) == 0 || len(tombstone.Position.TieBreaker) == 0 {
		t.Fatalf("CDC delete tombstone metadata = %#v, want explicit source-ordered delete evidence", tombstone)
	}
	var key map[string]any
	if err := json.Unmarshal(tombstone.Key, &key); err != nil {
		t.Fatalf("CDC delete tombstone key = %q, want JSON: %v", tombstone.Key, err)
	}
	if got, want := key, map[string]any{"tenant": "retain", "id": float64(9)}; !reflect.DeepEqual(got, want) {
		t.Fatalf("CDC delete tombstone key = %#v, want %#v", got, want)
	}
	if duplicate, err := CDCDeleteTombstone(events[0], []string{"tenant", "id"}); err != nil || !reflect.DeepEqual(duplicate, tombstone) {
		t.Fatalf("CDCDeleteTombstone() replay = (%#v, %v), want the same explicit tombstone %#v", duplicate, err, tombstone)
	}
}

func TestCDCDeleteTombstoneRefusesNonDeleteAndIncompleteSourceKeys(t *testing.T) {
	deleteEvent := connectors.CDCEvent{
		Operation: "delete",
		Record:    connectors.Record{"tenant": "retain", "id": int64(9)},
		State:     connectors.Record{"lsn": "0/16B6C50"},
	}
	for _, event := range []connectors.CDCEvent{
		{Operation: "insert", Record: deleteEvent.Record, State: deleteEvent.State},
		{Operation: "delete", Record: connectors.Record{"tenant": "retain"}, State: deleteEvent.State},
		{Operation: "delete", Record: deleteEvent.Record, State: connectors.Record{}},
	} {
		if tombstone, err := CDCDeleteTombstone(event, []string{"tenant", "id"}); err == nil || tombstone.Key != nil || len(tombstone.EventID) != 0 {
			t.Fatalf("CDCDeleteTombstone(%#v) = (%#v, %v), want no envelope and a pre-apply refusal", event, tombstone, err)
		}
	}
}
