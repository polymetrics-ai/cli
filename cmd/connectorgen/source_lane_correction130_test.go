package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// These owner regressions adapt sealed127's actual-loader counterexamples.
// The private reviewer overlay and its original receipts remain immutable.
func TestSourceLane130CollectionEffectiveObjects(t *testing.T) {
	type testCase struct{ name, schema, path, kind, code string }
	cases := []testCase{}
	for _, composition := range []string{"allOf", "oneOf", "anyOf"} {
		for _, depth := range []int{1, 2} {
			schema := fmt.Sprintf(`{"type":"object","%s":[{"$ref":"https://example.invalid/unresolved"}],"properties":{"records":{"type":"array","items":{"type":"object"}}}}`, composition)
			path := "/records"
			for i := 0; i < depth; i++ {
				schema = `{"type":"object","properties":{"outer":` + schema + `}}`
				path = "/outer" + path
			}
			cases = append(cases, testCase{fmt.Sprintf("%s_depth_%d", composition, depth), schema, path, "collection", "source_collection_interpretation_unresolved"})
		}
	}
	for _, typ := range []string{"array", "string", "number", "boolean"} {
		cases = append(cases, testCase{typ + "_single", fmt.Sprintf(`{"type":%q,"properties":{"metadata":{"type":"string"}},"items":{"type":"object"}}`, typ), "", "single_resource", "source_collection_interpretation_invalid"})
		cases = append(cases, testCase{typ + "_item", fmt.Sprintf(`{"type":"array","items":{"type":%q,"properties":{"id":{"type":"string"}}}}`, typ), "", "collection", "source_collection_interpretation_invalid"})
		cases = append(cases, testCase{typ + "_wrapper", fmt.Sprintf(`{"type":"object","properties":{"outer":{"type":%q,"properties":{"records":{"type":"array","items":{"type":"object"}}}}}}`, typ), "/outer/records", "collection", "source_collection_interpretation_invalid"})
	}
	for _, typ := range []string{`["object","null"]`, `null`, `42`} {
		cases = append(cases, testCase{"unsupported_type_" + typ, `{"type":` + typ + `,"properties":{"metadata":{"type":"string"}}}`, "", "single_resource", "source_collection_interpretation_unresolved"})
	}
	for _, object := range []string{`"type":"object",`, ""} {
		cases = append(cases, testCase{"positive_single_" + object, `{` + object + `"properties":{"id":{"type":"string"}}}`, "", "single_resource", ""})
		cases = append(cases, testCase{"positive_deeper_" + object, `{` + object + `"properties":{"outer":{` + object + `"properties":{"records":{"type":"array","items":{` + object + `"properties":{"id":{"type":"string"}}}}}}}}`, "/outer/records", "collection", ""})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, cohort, key, facts, ref := sourceCollection122Fixture(t, tc.schema, "")
			baseline, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary"), ResponseInterpretations: []sourceResponseInterpretation{{Kind: tc.kind, ResponseSchema: ref, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary")}}}
			if tc.kind == "collection" {
				a.ResponseInterpretations[0].RecordsPointer = &tc.path
			}
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
			if err != nil {
				t.Fatal(err)
			}
			if got.SourceTotals.Operations != 2 || got.SourceTotals.Cells != 14 {
				t.Fatal("lost actual retained membership")
			}
			want := "applicable"
			if tc.kind == "single_resource" {
				want = "not_applicable"
			}
			if tc.code != "" {
				want = "undetermined"
			}
			for i := range got.SourceOperations {
				row := &got.SourceOperations[i]
				if !row.Source.Observed || !reflect.DeepEqual(row.Source, baseline.SourceOperations[i].Source) {
					t.Fatal("retained source changed")
				}
				for j := range row.Lanes {
					cell := &row.Lanes[j]
					if row.Source.Key != key || cell.Lane != "etl" {
						if !reflect.DeepEqual(*cell, baseline.SourceOperations[i].Lanes[j]) {
							t.Errorf("unaffected source/lane changed: %+v/%s", row.Source.Key, cell.Lane)
						}
						continue
					}
					if cell.Applicability != want {
						t.Errorf("actual builder key=%+v lane=etl got=%s want=%s", key, cell.Applicability, want)
					}
					if cell.State == "implemented" || len(cell.References) != 0 || len(cell.ProofRefs) != 0 {
						t.Error("interpretation created executable proof")
					}
					found := tc.code == ""
					for _, d := range cell.Diagnostics {
						if d.Code == tc.code && d.Key == key && d.Pointer == ref.Pointer && reflect.DeepEqual(d.Lanes, []string{"etl"}) {
							found = true
						}
					}
					if !found {
						t.Errorf("missing exact scoped diagnostic %s: %+v", tc.code, cell.Diagnostics)
					}
					if tc.code == "" && got.Validation.Status != "valid" {
						t.Errorf("positive did not reach valid builder: %+v", got.Validation)
					}
					// Force producer and supplied candidate to agree on the bad claim.
					// The independent emitted-evidence path must still consult source.
					if tc.code != "" {
						cell.RuleID = "source_response_interpretation"
						cell.Applicability = "applicable"
						if tc.kind == "single_resource" {
							cell.Applicability = "not_applicable"
						}
					}
				}
			}
			if tc.code != "" {
				raw, _ := json.Marshal([]sourceSemanticAnnotation{a})
				found := false
				for _, d := range validateSourceLaneInterpretationEvidence(got, raw) {
					found = found || (d.Key == key && d.Code == "source_collection_authority_missing")
				}
				if !found {
					t.Error("independent emitted evidence accepted source-incompatible interpretation")
				}
			}
		})
	}
}
