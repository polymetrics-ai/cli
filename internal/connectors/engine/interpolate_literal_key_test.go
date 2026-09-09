package engine

import (
	"strings"
	"testing"
)

func TestResolveRecordPathValueUsesLiteralDottedKey(t *testing.T) {
	value, err := resolveRecordPathValue(map[string]any{"1. open": "100.25"}, strings.Split("1. open", "."))
	if err != nil {
		t.Fatalf("resolveRecordPathValue() error = %v", err)
	}
	if value != "100.25" {
		t.Fatalf("resolveRecordPathValue() = %#v, want literal dotted field value", value)
	}
}
