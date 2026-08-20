package engine

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInspectRecordSchemaExpandsUnionArmsBeforeMeasuring(t *testing.T) {
	const schema = `{
		"oneOf": [
			{"type":"object","properties":{"ticket":{"type":"object"}},"additionalProperties":false},
			{"type":"object","properties":{"tickets":{"type":"array"}},"additionalProperties":false}
		],
		"additionalProperties": false
	}`
	for _, keyword := range []string{"oneOf", "anyOf"} {
		t.Run(keyword, func(t *testing.T) {
			shape, err := InspectRecordSchema(json.RawMessage(strings.Replace(schema, "oneOf", keyword, 1)))
			if err != nil {
				t.Fatalf("InspectRecordSchema: %v", err)
			}
			if shape.RootUnion != keyword || shape.ArmCount != 2 || shape.SubstantiveArmCount != 2 {
				t.Fatalf("shape = %+v, want two substantive %s arms", shape, keyword)
			}
			candidate := `{"oneOf":[{"type":"object","properties":{"ticket":{"type":"object"}}},{"type":"object","properties":{"tickets":{"type":"array"}}}]}`
			if err := ValidatePromotableRecordSchema(json.RawMessage(strings.Replace(candidate, "oneOf", keyword, 1))); err == nil || !strings.Contains(err.Error(), "separate named write action") {
				t.Fatalf("ValidatePromotableRecordSchema union error = %v, want named-action guidance", err)
			}
		})
	}
}

func TestInspectRecordSchemaPreservesWrapperFieldsWhileExpandingArms(t *testing.T) {
	shape, err := InspectRecordSchema(json.RawMessage(`{
		"type":"object",
		"required":["account_id"],
		"properties":{"account_id":{"type":"integer"}},
		"oneOf":[
			{"required":["ticket"]},
			{"required":["tickets"]}
		],
		"additionalProperties":false
	}`))
	if err != nil {
		t.Fatalf("InspectRecordSchema: %v", err)
	}
	if shape.RootUnion != "oneOf" || shape.ArmCount != 2 || shape.SubstantiveArmCount != 2 {
		t.Fatalf("shape = %+v, want two expanded arms with wrapper fields preserved", shape)
	}
}

func TestMergeRecordSchemaRequiredPreservesStableUnionOrder(t *testing.T) {
	merged, ok := mergeRecordSchemaRequired(
		json.RawMessage(`[
			"account_id",
			"shared"
		]`),
		json.RawMessage(`[
			"ticket",
			"account_id",
			"batch",
			"ticket"
		]`),
	)
	if !ok {
		t.Fatal("mergeRecordSchemaRequired returned false")
	}
	var got []string
	if err := json.Unmarshal(merged, &got); err != nil {
		t.Fatalf("unmarshal merged required fields: %v", err)
	}
	want := []string{"account_id", "shared", "ticket", "batch"}
	if len(got) != len(want) {
		t.Fatalf("merged required = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("merged required = %v, want %v", got, want)
		}
	}
}

func TestMergeRecordSchemaRequiredKeepsAnEmptyRequiredArray(t *testing.T) {
	merged, ok := mergeRecordSchemaRequired(json.RawMessage(`[]`), json.RawMessage(`[]`))
	if !ok {
		t.Fatal("mergeRecordSchemaRequired returned false")
	}
	if got, want := string(merged), `[]`; got != want {
		t.Fatalf("merged required = %s, want %s", got, want)
	}
}

func TestValidatePromotableRecordSchemaRejectsOnlyEmptyObject(t *testing.T) {
	err := ValidatePromotableRecordSchema(json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`))
	if err == nil || !strings.Contains(err.Error(), "only an empty object") {
		t.Fatalf("ValidatePromotableRecordSchema error = %v, want hollow-schema rejection", err)
	}
}

func TestValidatePromotableRecordSchemaAllowsClosedNamedFields(t *testing.T) {
	err := ValidatePromotableRecordSchema(json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"additionalProperties":false}`))
	if err != nil {
		t.Fatalf("ValidatePromotableRecordSchema: %v", err)
	}
}

func TestPreflightWriteActionAllowsOnlyDeclaredNoInputEmptyRecord(t *testing.T) {
	empty := json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)
	tests := []struct {
		name    string
		action  WriteAction
		wantErr bool
	}{
		{
			name: "configuration-bound no-body operation",
			action: WriteAction{
				Name: "delete_configured_repo", Method: "DELETE", Path: "/repos/{{ config.owner }}/{{ config.repo }}",
				BodyType: "none", RecordSchema: empty,
			},
		},
		{
			name: "hollow JSON operation",
			action: WriteAction{
				Name: "collapsed_provider_body", Method: "POST", Path: "/widgets", RecordSchema: empty,
			},
			wantErr: true,
		},
		{
			name: "undeclared record-bound path",
			action: WriteAction{
				Name: "missing_path_schema", Method: "DELETE", Path: "/widgets/{{ record.id }}",
				BodyType: "none", RecordSchema: empty,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connector := New(Bundle{Name: "widgets", Writes: []WriteAction{tt.action}}, nil)
			err := connector.PreflightWriteAction(tt.action.Name)
			if (err != nil) != tt.wantErr {
				t.Fatalf("PreflightWriteAction() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRecordSchemaFieldMappingRequiresExactCompleteFields(t *testing.T) {
	schema := json.RawMessage(`{
		"type": "object",
		"required": ["targetId", "http"],
		"additionalProperties": false,
		"properties": {
			"targetId": {"type": "string"},
			"http": {"type": "string"}
		}
	}`)
	for _, testCase := range []struct {
		name    string
		fields  []string
		wantErr string
	}{
		{name: "exact provider fields", fields: []string{"targetId", "http"}},
		{name: "whitespace is not normalized", fields: []string{" targetId ", "http"}, wantErr: "not declared"},
		{name: "unknown field", fields: []string{"target_id", "http"}, wantErr: "not declared"},
		{name: "incomplete mapping", fields: []string{"targetId"}, wantErr: "required field \"http\" is not mapped"},
		{name: "duplicate mapping", fields: []string{"targetId", "targetId", "http"}, wantErr: "duplicates field"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			err := ValidateRecordSchemaFieldMapping(schema, testCase.fields)
			if testCase.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateRecordSchemaFieldMapping: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), testCase.wantErr) {
				t.Fatalf("ValidateRecordSchemaFieldMapping error = %v, want %q", err, testCase.wantErr)
			}
		})
	}
}
