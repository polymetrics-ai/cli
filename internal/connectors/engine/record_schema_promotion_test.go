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
