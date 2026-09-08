package engine

import (
	"encoding/json"
	"testing"

	"polymetrics.ai/internal/connectors"
)

type savedPreparedHook206 struct{ claims bool }

func (h savedPreparedHook206) ConnectorName() string               { return "fixture" }
func (h savedPreparedHook206) HandlesWriteAction(WriteAction) bool { return h.claims }
func (h savedPreparedHook206) PrepareWrite(WriteAction, []connectors.Record) (PreparedWriteHookPlan, bool, error) {
	panic("preflight invoked executable hook")
}

type savedRecordHook206 struct{}

func (savedRecordHook206) ConnectorName() string { return "fixture" }
func (savedRecordHook206) MapWriteRecord(WriteAction, connectors.Record) (connectors.Record, bool, error) {
	panic("preflight invoked record hook")
}

type savedBareHook206 struct{}

func (savedBareHook206) ConnectorName() string { return "fixture" }

func TestSavedPreflightSelectedHookContracts206(t *testing.T) {
	for _, tc := range []struct {
		name                     string
		hook                     Hooks
		missingSchema, wantError bool
	}{
		{"prepared_claimed", savedPreparedHook206{true}, false, false},
		{"prepared_unclaimed", savedPreparedHook206{false}, false, true},
		{"record_mapper", savedRecordHook206{}, false, false},
		{"bare_name", savedBareHook206{}, false, true},
		{"absent", nil, false, true},
		{"prepared_without_schema", savedPreparedHook206{true}, true, true},
		{"mapper_without_schema", savedRecordHook206{}, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			action := WriteAction{Name: "create_widget", Method: "POST", Path: "/widgets", Hook: "fixture", BodyType: "json", BodyFields: []string{"name"}, RecordSchema: json.RawMessage(`{"type":"object","required":["name"],"properties":{"name":{"type":"string"}},"additionalProperties":false}`)}
			if tc.missingSchema {
				action.RecordSchema = nil
			}
			bundle := Bundle{Name: "fixture", Writes: []WriteAction{action}}
			bundle.Metadata.Capabilities.Write = true
			err := New(bundle, tc.hook).PreflightSavedWriteAction(action.Name)
			if (err != nil) != tc.wantError {
				t.Fatalf("loaded selected preflight error=%v wantError=%v", err, tc.wantError)
			}
		})
	}
}
