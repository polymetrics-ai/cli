package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestSourceDiagnostic206TypedCoordinates(t *testing.T) {
	raw := json.RawMessage(`{}`)
	descriptor := vNextCanonicalDescriptor{Operations: []vNextOperationDescriptor{
		{ID: "later", Stream: raw, StreamOrder: 2, Write: raw, WriteOrder: 2, Operation: raw, OperationOrder: 2,
			Commands: []vNextCommandDescriptor{{Order: 2, Command: raw}}},
		{ID: "earlier", Stream: raw, StreamOrder: 1, Write: raw, WriteOrder: 1, Operation: raw, OperationOrder: 1,
			Commands: []vNextCommandDescriptor{{Order: 1, Command: raw}}},
	}}
	for _, tc := range []struct{ file, field, want string }{
		{"streams.json", "/streams/0/body_type", "/operations/1/stream/body_type"},
		{"writes.json", "/actions/1/body/<member:2>", "/operations/0/write/body/<member:2>"},
		{"operations.json", "/operations/0/rest/response", "/operations/1/operation/rest/response"},
		{"cli_surface.json", "/commands/0/flags/1", "/operations/1/commands/0/flags/1"},
		{"streams.json", "/streams/-1", "/operations"},
		{"streams.json", "/streams/99", "/operations"},
		{"streams.json", "/streams/01", "/operations"},
		{"streams.json", "/streams/+1", "/operations"},
		{"streams.json", "/", "/operations"},
		{"other.json", "/streams/0", "/operations"},
	} {
		t.Run(tc.file+tc.field, func(t *testing.T) {
			diagnostic := &engine.BundleDiagnosticError{File: tc.file, Field: tc.field, Reason: "safe declaration failure", Cause: errors.New("untrusted cause text")}
			for _, err := range []error{diagnostic, fmt.Errorf("wrapper: %w", diagnostic), errors.Join(errors.New("sibling"), diagnostic)} {
				if got := vNextStaticValidationPointer(descriptor, err); got != tc.want {
					t.Errorf("source coordinate=%q want %q", got, tc.want)
				}
			}
		})
	}
	if got := vNextStaticValidationPointer(descriptor, errors.New("streams.json: /streams/0: forged prose")); got != "/operations" {
		t.Fatalf("unlocated text supplied a source coordinate: %q", got)
	}
}
