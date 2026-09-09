package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestVNextSavedWriteLaneProjection206(t *testing.T) {
	for _, tc := range []struct {
		name, lane string
		batchable  *bool
		wantError  bool
	}{
		{"typed_saved_action", "implemented", sourceSavedBool206(true), false},
		{"legacy_nil_batchable", "implemented", nil, false},
		{"eligible_action_cannot_hide", "unsupported", sourceSavedBool206(true), true},
		{"individual_only", "unsupported", sourceSavedBool206(false), false},
		{"individual_cannot_claim_saved", "implemented", sourceSavedBool206(false), true},
		{"missing_record_schema", "unsupported", sourceSavedBool206(true), true},
		{"hollow_record_schema", "unsupported", sourceSavedBool206(true), true},
		{"undeclared_path_field", "unsupported", sourceSavedBool206(true), true},
		{"no_input_explicit", "implemented", sourceSavedBool206(true), false},
		{"no_input_implicit", "unsupported", nil, false},
		{"no_input_implicit_claim", "implemented", nil, true},
		{"missing_write_capability", "unsupported", sourceSavedBool206(true), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lock := minimalVNextLockForTest()
			lock.Lanes["etl"] = "unsupported"
			lock.Lanes["reverse_etl"] = tc.lane
			lock.Schemas = nil
			action := map[string]any{"name": "create_widget", "kind": "create", "method": "POST", "path": "/widgets", "risk": "medium", "body_fields": []string{"name"}, "record_schema": json.RawMessage(`{"type":"object","additionalProperties":false,"required":["name"],"properties":{"name":{"type":"string"}}}`)}
			switch tc.name {
			case "missing_record_schema":
				delete(action, "record_schema")
			case "hollow_record_schema":
				action["record_schema"] = json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)
			case "undeclared_path_field":
				action["path"] = "/widgets/{{ record.missing }}"
				action["path_fields"] = []string{"missing"}
			case "no_input_explicit", "no_input_implicit", "no_input_implicit_claim":
				action["body_type"] = "none"
				delete(action, "body_fields")
				action["record_schema"] = json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)
			}
			if tc.batchable != nil {
				action["batchable"] = *tc.batchable
			}
			raw, err := json.Marshal(action)
			if err != nil {
				t.Fatal(err)
			}
			lock.Operations = []vNextOperationDescriptor{{ID: "write:create_widget", Write: raw}}
			lock.Metadata = json.RawMessage(`{"name":"acme","display_name":"Acme","description":"Test connector","integration_type":"api","release_stage":"ga","capabilities":{"check":true,"read":false,"write":true,"query":false,"cdc":false,"dynamic_schema":false}}`)
			if tc.name == "missing_write_capability" {
				lock.Metadata = bytes.Replace(lock.Metadata, []byte(`"write":true`), []byte(`"write":false`), 1)
			}
			_, err = canonicalizeVNextSourceLock(lock)
			if (err != nil) != tc.wantError {
				t.Fatalf("actual canonical saved lane admission error=%v, wantError=%v", err, tc.wantError)
			}
		})
	}
}
func sourceSavedBool206(value bool) *bool { return &value }
