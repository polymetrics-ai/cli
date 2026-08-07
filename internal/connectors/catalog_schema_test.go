package connectors

import (
	"encoding/json"
	"testing"
)

func TestStreamFromSchemaPreservesStaticBundleShape(t *testing.T) {
	raw := json.RawMessage(`{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"title": "widgets",
		"type": "object",
		"x-primary-key": ["id"],
		"x-cursor-field": "updated_at",
		"properties": {
			"id": {"type": "string"},
			"active": {"type": ["boolean", "null"]},
			"updated_at": {"type": "string", "format": "date-time"}
		}
	}`)

	stream, err := StreamFromSchema("widgets", "Widget records", raw)
	if err != nil {
		t.Fatalf("StreamFromSchema: %v", err)
	}
	if string(stream.Schema) != string(raw) {
		t.Fatalf("schema = %s, want verbatim %s", stream.Schema, raw)
	}
	if got, want := stream.PrimaryKey, []string{"id"}; !sameStringSlice(got, want) {
		t.Fatalf("primary key = %v, want %v", got, want)
	}
	if got, want := stream.CursorFields, []string{"updated_at"}; !sameStringSlice(got, want) {
		t.Fatalf("cursor fields = %v, want %v", got, want)
	}
	if got, want := stream.Fields, []Field{
		{Name: "active", Type: "boolean"},
		{Name: "id", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}; !sameFields(got, want) {
		t.Fatalf("fields = %#v, want %#v", got, want)
	}
}

func sameStringSlice(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func sameFields(got, want []Field) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
