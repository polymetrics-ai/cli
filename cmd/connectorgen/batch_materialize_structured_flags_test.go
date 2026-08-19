package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

// The generic materializer used to leave any action whose required value was
// a container off the command surface, despite commandrunner already having a
// declaration-bound structured JSON flag path. Keep the two layers aligned:
// materialization may expose one top-level JSON flag per required container,
// but it must never fabricate flags for that container's descendants.
func TestMaterializedWriteFlagsUseTopLevelStructuredJSONForRequiredContainers(t *testing.T) {
	action := engine.WriteAction{
		Name: "create_widget",
		RecordSchema: json.RawMessage(`{
			"type":"object",
			"additionalProperties":false,
			"required":["name","payload","items"],
			"properties":{
				"name":{"type":"string"},
				"payload":{"type":"object","additionalProperties":false,"required":["kind"],"properties":{"kind":{"type":"string"}}},
				"items":{"type":"array","items":{"type":"object","additionalProperties":false,"required":["id"],"properties":{"id":{"type":"integer"}}}}
			}
		}`),
	}

	flags, representable, err := materializedWriteFlags(action)
	if err != nil {
		t.Fatalf("materializedWriteFlags: %v", err)
	}
	if !representable {
		t.Fatal("representable = false, want declaration-bound structured flags")
	}
	got := map[string]struct {
		Type, MapsTo string
		Required     bool
	}{}
	for _, flag := range flags {
		got[flag.Name] = struct {
			Type, MapsTo string
			Required     bool
		}{flag.Type, flag.MapsTo, flag.Required}
	}
	want := map[string]struct {
		Type, MapsTo string
		Required     bool
	}{
		"name":    {Type: "string", MapsTo: "record.name", Required: true},
		"payload": {Type: "json", MapsTo: "record.payload", Required: true},
		"items":   {Type: "json", MapsTo: "record.items", Required: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("flags = %#v, want %#v", got, want)
	}
	for _, field := range []string{"payload", "items"} {
		if err := engine.ValidateStructuredJSONRecordField(action.RecordSchema, field); err != nil {
			t.Fatalf("structured field %q: %v", field, err)
		}
	}
}
