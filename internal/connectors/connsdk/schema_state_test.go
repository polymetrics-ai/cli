package connsdk

import (
	"encoding/json"
	"testing"
)

func TestInferType(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, "null"},
		{true, "boolean"},
		{json.Number("42"), "integer"},
		{json.Number("4.2"), "number"},
		{"hello", "string"},
		{"2026-06-25T10:00:00Z", "timestamp"},
		{"2026-06-25", "timestamp"},
		{map[string]any{"a": 1}, "object"},
		{[]any{1, 2}, "array"},
	}
	for _, c := range cases {
		if got := InferType(c.in); got != c.want {
			t.Errorf("InferType(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestInferFieldsSortedAndTyped(t *testing.T) {
	rec := Record{
		"name":       "ada",
		"id":         json.Number("1"),
		"active":     true,
		"updated_at": "2026-06-25T10:00:00Z",
	}
	fields := InferFields(rec)
	if len(fields) != 4 {
		t.Fatalf("len = %d, want 4", len(fields))
	}
	if fields[0].Name != "active" || fields[1].Name != "id" {
		t.Fatalf("not sorted: %+v", fields)
	}
	byName := map[string]string{}
	for _, f := range fields {
		byName[f.Name] = f.Type
	}
	if byName["id"] != "integer" || byName["updated_at"] != "timestamp" || byName["active"] != "boolean" {
		t.Fatalf("types = %+v", byName)
	}
}

func TestMaxCursor(t *testing.T) {
	cases := []struct{ a, b, want string }{
		{"", "5", "5"},
		{"5", "", "5"},
		{"3", "10", "10"}, // numeric
		{"10", "3", "10"}, // numeric
		{"2026-01-01T00:00:00Z", "2026-06-01T00:00:00Z", "2026-06-01T00:00:00Z"}, // time
		{"2026-06-01", "2026-01-01", "2026-06-01"},                               // date
		{"abc", "abd", "abd"}, // lexicographic
	}
	for _, c := range cases {
		if got := MaxCursor(c.a, c.b); got != c.want {
			t.Errorf("MaxCursor(%q,%q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

func TestCursorRoundTrip(t *testing.T) {
	state := map[string]string{"snapshot_completed": "true"}
	next := WithCursor(state, "2026-06-25T00:00:00Z")
	if Cursor(next) != "2026-06-25T00:00:00Z" {
		t.Fatalf("Cursor = %q", Cursor(next))
	}
	if _, ok := state["cursor"]; ok {
		t.Fatal("WithCursor mutated input")
	}
	if next["snapshot_completed"] != "true" {
		t.Fatal("WithCursor dropped existing keys")
	}
}
