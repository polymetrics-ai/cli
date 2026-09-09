package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Expectations are author-written136 states, never inferred by a production
// predicate, output manifest or comparator. Every case uses retained bytes.
type collection137Case struct {
	name, schema, contract, applicability, scopeCode string
}

func collection137Cases() []collection137Case {
	cases := []collection137Case{
		{"object_items_positive", `{"type":"array","items":{"type":"object","properties":{"id":{"type":"string"}}}}`, "", "applicable", ""},
		{"scalar_items_excluded", `{"type":"array","items":{"type":"string"}}`, "", "not_applicable", ""},
		{"nested_array_with_properties", `{"type":"array","items":{"type":"array","properties":{"id":{"type":"string"}},"items":{"type":"object"}}}`, "", "undetermined", "source_collection_scope_unknown"},
		{"mixed_item_type_with_properties", `{"type":"array","items":{"type":["object","null"],"properties":{"id":{"type":"string"}}}}`, "", "undetermined", "source_collection_scope_unknown"},
		{"prefix_scalar_tail_objects", `{"type":"array","prefixItems":[{"type":"string"}],"items":{"type":"object"}}`, "", "undetermined", "source_collection_scope_unknown"},
		{"object_unknown_descendant_positive", `{"type":"array","items":{"type":"object","properties":{"detail":{"$ref":"https://example.invalid/schema"}}}}`, "", "applicable", ""},
		{"closed_scalar_object", `{"type":"object","properties":{"id":{"type":"string"}},"additionalProperties":false}`, "", "not_applicable", ""},
		{"closed_unknown_child", `{"type":"object","properties":{"id":{}},"additionalProperties":false}`, "", "undetermined", "source_collection_interpretation_missing"},
		{"closed_array_override", `{"type":"array","properties":{"id":{"type":"string"}},"additionalProperties":false}`, "", "undetermined", "source_collection_scope_unknown"},
		{"closed_mixed_override", `{"type":["object","null"],"properties":{"id":{"type":"string"}},"additionalProperties":false}`, "", "undetermined", "source_collection_scope_unknown"},
		{"agreeing_array_alternatives", `{"oneOf":[{"type":"array","items":{"type":"object"}},{"type":"array","items":{"properties":{}}}]}`, "", "applicable", ""},
		{"agreeing_scalar_alternatives", `{"anyOf":[{"type":"string"},{"type":"number"}]}`, "", "not_applicable", ""},
		{"mixed_alternatives", `{"oneOf":[{"type":"array","items":{"type":"object"}},{"type":"string"}]}`, "", "undetermined", "source_collection_scope_unknown"},
		{"unknown_conjunct", `{"type":"array","items":{"type":"object"},"allOf":[{"$ref":"https://example.invalid/schema"}]}`, "", "undetermined", "source_collection_scope_unknown"},
		{"invalid_type_known_composition", `{"type":["object","null"],"allOf":[{"type":"array","items":{"type":"object"}}]}`, "", "undetermined", "source_collection_scope_unknown"},
		{"local_item_ref", `{"type":"array","items":{"$ref":"#/Record"}}`, `{"Record":{"type":"object"}}`, "applicable", ""},
		{"local_root_ref", `{"$ref":"#/Records","description":"annotation"}`, `{"Records":{"type":"array","items":{"type":"object"}}}`, "applicable", ""},
		{"escaped_ref", `{"type":"array","items":{"$ref":"#/a~1b~0c"}}`, `{"a/b~c":{"type":"object"},"equal":{"type":"object"}}`, "applicable", ""},
		{"missing_ref", `{"type":"array","items":{"$ref":"#/Missing"}}`, `{}`, "undetermined", "source_collection_scope_unknown"},
		{"external_ref", `{"type":"array","items":{"$ref":"https://example.invalid/schema"}}`, "", "undetermined", "source_collection_scope_unknown"},
		{"cyclic_ref", `{"type":"array","items":{"$ref":"#/Loop"}}`, `{"Loop":{"$ref":"#/Loop"}}`, "undetermined", "source_collection_scope_unknown"},
		{"invalid_escape", `{"type":"array","items":{"$ref":"#/a~2b"}}`, `{"a~2b":{"type":"object"}}`, "undetermined", "source_collection_scope_unknown"},
		{"structural_ref_sibling", `{"type":"array","items":{"$ref":"#/Record","properties":{}}}`, `{"Record":{"type":"object"}}`, "undetermined", "source_collection_scope_unknown"},
	}
	// Each row includes root and array-item obligations; scalars are positively
	// excluded while unsupported encodings and absent shape remain unknown.
	for _, tc := range []struct{ name, schema, itemWant, rootWant, scope string }{
		{"object_absent_properties", `{"type":"object"}`, "applicable", "undetermined", "source_collection_interpretation_missing"},
		{"object_empty_properties", `{"type":"object","properties":{}}`, "applicable", "undetermined", "source_collection_interpretation_missing"},
		{"typeless_empty_properties", `{"properties":{}}`, "applicable", "undetermined", "source_collection_interpretation_missing"},
		{"typeless_properties", `{"properties":{"id":{"type":"string"}}}`, "applicable", "undetermined", "source_collection_interpretation_missing"},
		{"unconstrained", `{}`, "undetermined", "undetermined", "source_collection_scope_unknown"},
		{"typeless_items", `{"items":{"type":"object"}}`, "undetermined", "undetermined", "source_collection_scope_unknown"},
		{"typeless_competing_items", `{"properties":{},"items":{"type":"object"}}`, "undetermined", "undetermined", "source_collection_scope_unknown"},
		{"typeless_competing_prefix", `{"properties":{},"prefixItems":[]}`, "undetermined", "undetermined", "source_collection_scope_unknown"},
	} {
		cases = append(cases, collection137Case{tc.name + "_root", tc.schema, "", tc.rootWant, tc.scope}, collection137Case{tc.name + "_item", `{"type":"array","items":` + tc.schema + `}`, "", tc.itemWant, map[bool]string{true: "source_collection_scope_unknown"}[tc.itemWant == "undetermined"]})
	}
	for _, typ := range []string{"string", "number", "integer", "boolean", "null"} {
		for _, location := range []string{"root", "item"} {
			schema := fmt.Sprintf(`{"type":%q,"properties":{"id":{"type":"string"}}}`, typ)
			if location == "item" {
				schema = `{"type":"array","items":` + schema + `}`
			}
			cases = append(cases, collection137Case{typ + "_properties_" + location, schema, "", "not_applicable", ""})
		}
	}
	for _, typ := range []string{`null`, `42`, `""`, `"unsupported"`, `["object","null"]`} {
		for _, location := range []string{"root", "item"} {
			schema := `{"type":` + typ + `,"properties":{}}`
			if location == "item" {
				schema = `{"type":"array","items":` + schema + `}`
			}
			cases = append(cases, collection137Case{"unsupported_" + typ + "_" + location, schema, "", "undetermined", "source_collection_scope_unknown"})
		}
	}
	for _, props := range []string{`null`, `[]`, `"wrong"`, `42`, `false`} {
		for _, typ := range []string{``, `"type":"object",`} {
			schema := `{` + typ + `"properties":` + props + `}`
			cases = append(cases, collection137Case{"malformed_properties_root_" + typ + props, schema, "", "undetermined", "source_collection_scope_unknown"}, collection137Case{"malformed_properties_item_" + typ + props, `{"type":"array","items":` + schema + `}`, "", "undetermined", "source_collection_scope_unknown"})
		}
	}
	for _, item := range []string{`null`, `true`, `42`, `"wrong"`, `[]`, `{}`} {
		cases = append(cases, collection137Case{"invalid_items_" + item, `{"type":"array","items":` + item + `}`, "", "undetermined", "source_collection_scope_unknown"})
	}
	cases = append(cases, collection137Case{"missing_items", `{"type":"array"}`, "", "undetermined", "source_collection_scope_unknown"})
	for _, prefix := range []string{`[]`, `[{"type":"string"}]`, `[{"type":"object"}]`, `null`, `false`, `42`, `"wrong"`} {
		for _, item := range []string{"object", "string"} {
			want, code := "undetermined", "source_collection_scope_unknown"
			if prefix == `[]` {
				code = ""
				want = "applicable"
				if item == "string" {
					want = "not_applicable"
				}
			}
			cases = append(cases, collection137Case{"prefix_" + prefix + "_tail_" + item, `{"type":"array","prefixItems":` + prefix + `,"items":{"type":"` + item + `"}}`, "", want, code})
		}
	}
	for _, tail := range []string{``, `,"items":false`} {
		cases = append(cases, collection137Case{"fixed_objects_" + tail, `{"type":"array","prefixItems":[{"type":"object"}]` + tail + `}`, "", "undetermined", "source_collection_scope_unknown"})
	}
	for _, composition := range []string{"allOf", "oneOf", "anyOf"} {
		cases = append(cases, collection137Case{"item_" + composition, `{"type":"array","items":{"type":"object","` + composition + `":[{"type":"object"}]}}`, "", "undetermined", "source_collection_scope_unknown"})
	}
	return cases
}

type collection137Violation struct {
	Key  sourceOperationKey
	Code string
}

func collection137Oracle(got sourceLaneManifest, wants [2]collection137Case) []collection137Violation {
	violations := []collection137Violation{}
	add := func(key sourceOperationKey, code string) {
		violations = append(violations, collection137Violation{key, code})
	}
	keys := [2]sourceOperationKey{{Connector: "fixture", Inventory: "primary", ID: "source.a"}, {Connector: "fixture", Inventory: "primary", ID: "source.b"}}
	lanes := [7]string{"direct_read", "direct_write", "binary_download", "binary_upload", "etl", "reverse_etl", "sync_transport"}
	if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 || got.SourceTotals.Operations != 2 {
		add(keys[0], "universe")
	}
	for i, row := range got.SourceOperations {
		if i >= len(keys) {
			add(row.Source.Key, "extra_key")
			continue
		}
		if row.Source.Key != keys[i] || !row.Source.Observed || row.Facts.Status == "unavailable" {
			add(keys[i], "source_frontier")
		}
		if len(row.Lanes) != len(lanes) {
			add(keys[i], "lane_count")
		}
		for j, c := range row.Lanes {
			if j >= len(lanes) || c.Lane != lanes[j] {
				add(keys[i], "lane_order")
			}
			if c.State == "implemented" || len(c.References) != 0 || len(c.ProofRefs) != 0 {
				add(keys[i], "fabricated_proof")
			}
			if c.Lane == "direct_read" && c.Applicability != "applicable" {
				add(keys[i], "direct_read")
			}
			if c.Lane != "etl" {
				continue
			}
			want := wants[i]
			state, rule := "mapped_unproven", "facts_unresolved"
			switch want.applicability {
			case "applicable":
				rule = "source_record_collection"
			case "not_applicable":
				state = "not_applicable"
				rule = "fixed_noncollection_read"
			}
			if c.Applicability != want.applicability || c.State != state || c.RuleID != rule {
				add(keys[i], "etl_state")
			}
			pointer := fmt.Sprintf("/rest/operations/%d/source_operation/responses/200/content/application~1json/schema", i)
			if want.scopeCode != "" && !collection137HasDiagnostic(c, keys[i], want.scopeCode, pointer, "deficit") {
				add(keys[i], "etl_scope")
			}
			if want.applicability == "undetermined" && !collection137HasDiagnostic(c, keys[i], "source_record_shape_unresolved", "", "deficit") {
				add(keys[i], "etl_unknown")
			}
		}
	}
	return violations
}
func collection137HasDiagnostic(c sourceLaneCell, key sourceOperationKey, code, pointer, severity string) bool {
	for _, d := range c.Diagnostics {
		if d.Key == key && d.Code == code && d.Pointer == pointer && reflect.DeepEqual(d.Lanes, []string{"etl"}) && d.Stage == "classification" && d.Severity == severity {
			return true
		}
	}
	return false
}
func TestSourceLane137CollectionInvariant(t *testing.T) {
	for _, tc := range collection137Cases() {
		t.Run(tc.name, func(t *testing.T) {
			root, cohort, _, _, _ := sourceCollection122Fixture(t, tc.schema, tc.contract)
			before := sourceBindingFixtureSnapshot(t, root)
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			if v := collection137Oracle(got, [2]collection137Case{tc, tc}); len(v) > 0 {
				t.Errorf("retained builder violations=%+v; actual ETL=%+v", v, got.SourceOperations[0].Lanes[4])
			}
			again, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			a, _ := json.Marshal(got)
			b, _ := json.Marshal(again)
			if string(a) != string(b) {
				t.Error("nondeterministic builder")
			}
			if !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, root)) {
				t.Error("retained inputs mutated")
			}
		})
	}
}
func TestSourceLane137CollectionOracleFalsification(t *testing.T) {
	tc := collection137Case{schema: `{"type":"array","items":{}}`, applicability: "undetermined", scopeCode: "source_collection_scope_unknown"}
	root, cohort, _, _, _ := sourceCollection122Fixture(t, tc.schema, "")
	got, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var control sourceLaneManifest
	if err := json.Unmarshal(raw, &control); err != nil {
		t.Fatal(err)
	}
	wants := [2]collection137Case{tc, tc}
	if v := collection137Oracle(control, wants); len(v) != 0 {
		t.Fatalf("genuine control clone: %+v", v)
	}
	control.SourceOperations[0].Lanes[4].Applicability = "applicable"
	control.SourceOperations[0].Lanes[4].RuleID = "source_record_collection"
	control.SourceOperations[0].Lanes[4].Diagnostics = nil
	summarizeSourceLaneManifest(&control)
	forgedRaw, err := json.Marshal(control)
	if err != nil {
		t.Fatal(err)
	}
	var comparison sourceLaneManifest
	if err := json.Unmarshal(forgedRaw, &comparison); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(control, comparison) {
		t.Fatal("forged comparator setup")
	}
	for _, candidate := range []sourceLaneManifest{control, comparison} {
		violations := collection137Oracle(candidate, wants)
		found := false
		for _, v := range violations {
			found = found || (v.Key == (sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "source.a"}) && v.Code == "etl_state")
		}
		if !found {
			t.Errorf("coherent forgery escaped independent exact-key state oracle: %+v", violations)
		}
	}
}

func TestSourceLane137CollectionInterpretationAgreement(t *testing.T) {
	type expectation struct{ name, schema, kind, path, want, code string }
	cases := []expectation{
		{"empty_prefix_objects", `{"type":"array","prefixItems":[],"items":{"type":"object"}}`, "collection", "", "applicable", ""},
		{"empty_prefix_scalars", `{"type":"array","prefixItems":[],"items":{"type":"string"}}`, "collection", "", "undetermined", "source_collection_interpretation_invalid"},
		{"prefix_tail", `{"type":"array","prefixItems":[{"type":"string"}],"items":{"type":"object"}}`, "collection", "", "undetermined", "source_collection_interpretation_unresolved"},
		{"empty_root_collection", `{}`, "collection", "", "undetermined", "source_collection_interpretation_unresolved"},
		{"empty_root_single", `{}`, "single_resource", "", "undetermined", "source_collection_interpretation_unresolved"},
		{"empty_items", `{"type":"array","items":{}}`, "collection", "", "undetermined", "source_collection_interpretation_unresolved"},
		{"malformed_prefix", `{"type":"array","prefixItems":null,"items":{"type":"object"}}`, "collection", "", "undetermined", "source_collection_interpretation_unresolved"},
	}
	for _, tc := range []struct{ name, object, code string }{
		{"typed", `{"type":"object"}`, ""},
		{"typeless", `{"properties":{}}`, ""},
		{"ambiguous", `{"properties":{},"items":{"type":"object"}}`, "source_collection_interpretation_unresolved"},
		{"mixed", `{"type":["object","null"],"properties":{}}`, "source_collection_interpretation_unresolved"},
		{"null_type", `{"type":null,"properties":{}}`, "source_collection_interpretation_unresolved"},
		{"invalid_properties", `{"type":"object","properties":[]}`, "source_collection_interpretation_unresolved"},
		{"null_properties", `{"type":"object","properties":null}`, "source_collection_interpretation_unresolved"},
		{"array", `{"type":"array","properties":{}}`, "source_collection_interpretation_invalid"},
		{"scalar", `{"type":"string","properties":{}}`, "source_collection_interpretation_invalid"},
		{"unknown_descendant", `{"type":"object","properties":{"detail":{"$ref":"https://example.invalid/schema"}}}`, ""},
	} {
		single, collection := "not_applicable", "applicable"
		if tc.code != "" {
			single, collection = "undetermined", "undetermined"
		}
		cases = append(cases, expectation{tc.name + "_single", tc.object, "single_resource", "", single, tc.code}, expectation{tc.name + "_items", `{"type":"array","items":` + tc.object + `}`, "collection", "", collection, tc.code})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, cohort, key, facts, ref := sourceCollection122Fixture(t, tc.schema, "")
			baseline, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			interpretation := sourceResponseInterpretation{Kind: tc.kind, ResponseSchema: ref, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary")}
			if tc.kind == "collection" {
				interpretation.RecordsPointer = &tc.path
			}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary"), ResponseInterpretations: []sourceResponseInterpretation{interpretation}}
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
			if err != nil {
				t.Fatal(err)
			}
			if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 {
				t.Fatal("retained two-key frontier missing")
			}
			for i, row := range got.SourceOperations {
				if !reflect.DeepEqual(row.Source, baseline.SourceOperations[i].Source) {
					t.Error("source identity changed")
				}
				for j, c := range row.Lanes {
					if row.Source.Key != key || c.Lane != "etl" {
						if !reflect.DeepEqual(c, baseline.SourceOperations[i].Lanes[j]) {
							t.Errorf("unrelated occurrence/lane changed %v/%s", row.Source.Key, c.Lane)
						}
						continue
					}
					state, rule := "mapped_unproven", "source_response_interpretation"
					if tc.want == "not_applicable" {
						state = "not_applicable"
					}
					if tc.want == "undetermined" {
						rule = "facts_unresolved"
					}
					if c.Applicability != tc.want || c.State != state || c.RuleID != rule {
						t.Errorf("interpreted ETL=%s/%s/%s want=%s/%s/%s", c.Applicability, c.State, c.RuleID, tc.want, state, rule)
					}
					if len(c.References) != 0 || len(c.ProofRefs) != 0 || c.State == "implemented" {
						t.Error("interpretation created proof")
					}
					if tc.code != "" {
						severity := "deficit"
						if tc.code == "source_collection_interpretation_invalid" {
							severity = "error"
						}
						if !collection137HasDiagnostic(c, key, tc.code, ref.Pointer, severity) {
							t.Errorf("missing exact interpretation diagnostic %s: %+v", tc.code, c.Diagnostics)
						}
					}
				}
			}
		})
	}
}

func TestSourceLane137CollectionAdmission(t *testing.T) {
	for _, negative := range []bool{false, true} {
		t.Run(fmt.Sprintf("unsupported_items_%t", negative), func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", func(lock *vNextSourceLock) {
				lock.Operations[0].Stream = json.RawMessage(`{"name":"widgets","path":"/widgets","records":{"path":"."},"schema":"schemas/widgets.json"}`)
			})
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				response := ownership132Response(doc)
				for name := range response {
					delete(response, name)
				}
				var record any
				if err := json.Unmarshal([]byte(sourceBindingRecord099F), &record); err != nil {
					t.Fatal(err)
				}
				response["type"] = "array"
				response["items"] = record
				if negative {
					response["prefixItems"] = []any{map[string]any{"type": "string"}}
				}
			})
			a.ResponseInterpretations = nil
			anchor := sourceBindingCitation099F(t, facts, a.IntendedBindings[0].SourceSchema.Pointer)
			a.IntendedBindings[0].SourceSchema = &anchor
			a.IntendedBindings[0].FieldMappings[0].Source = sourceBindingCitation099F(t, facts, anchor.Pointer+"/items")
			got, rebased := collection137AdmissionBuild(t, key, facts, a)
			cell := got.SourceOperations[0].Lanes[4]
			want, accepted := "applicable", 1
			if negative {
				want, accepted = "undetermined", 0
			}
			if cell.Applicability != want || len(cell.References) != accepted {
				t.Errorf("actual admitted stream frontier applicability=%s refs=%d want=%s/%d; diagnostics=%+v", cell.Applicability, len(cell.References), want, accepted, cell.Diagnostics)
			}
			if len(cell.IntendedBindings) != 1 || !sourceLaneTargetRefEqual(cell.IntendedBindings[0], rebased.IntendedBindings[0]) {
				t.Error("intended claim lost")
			}
			if negative {
				found := false
				for _, d := range cell.Diagnostics {
					found = found || (d.Key == key && d.Code == "target_applicability_conflict" && d.Stage == "reference" && d.Pointer == rebased.IntendedBindings[0].Artifact+"#"+rebased.IntendedBindings[0].Pointer && reflect.DeepEqual(d.Lanes, []string{"etl"}))
				}
				if !found {
					t.Error("did not reach exact applicability admission frontier")
				}
			}
			if cell.State == "implemented" || len(cell.ProofRefs) != 0 {
				t.Error("source/target agreement fabricated implementation proof")
			}
		})
	}
}

// Both occurrences share the actual admitted target and retained builder.
func collection137AdmissionBuild(t *testing.T, key sourceOperationKey, facts sourceFacts, a sourceSemanticAnnotation) (sourceLaneManifest, sourceSemanticAnnotation) {
	t.Helper()
	root := t.TempDir()
	for name, raw := range facts.bindings.Artifacts {
		proofWrite(t, root, name, raw)
	}
	for name, raw := range facts.bindings.Authoring {
		proofWrite(t, root, name, raw)
	}
	var doc map[string]any
	if err := json.Unmarshal(facts.Document, &doc); err != nil {
		t.Fatal(err)
	}
	ops := doc["rest"].(map[string]any)["operations"].([]any)
	siblingRaw, err := json.Marshal(ops[0])
	if err != nil {
		t.Fatal(err)
	}
	var sibling map[string]any
	if err := json.Unmarshal(siblingRaw, &sibling); err != nil {
		t.Fatal(err)
	}
	sibling["id"] = "retained.widgets.sibling"
	var record any
	if err := json.Unmarshal([]byte(sourceBindingRecord099F), &record); err != nil {
		t.Fatal(err)
	}
	schema := sibling["source_operation"].(map[string]any)["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)
	schema["schema"] = map[string]any{"type": "array", "items": record}
	doc["rest"].(map[string]any)["operations"] = append(ops, sibling)
	doc["schema_version"], doc["connector"], doc["counts"] = 2, "acme", map[string]any{"total": 2}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, root, "source.json", raw)
	cohort := sourceLaneCohort{SchemaVersion: 1, CohortID: "fixture", Inventories: []sourceInventoryAnchor{{Connector: "acme", Inventory: "primary", Class: "primary", Path: "source.json", SHA256: sourceBytesHash(raw), ExpectedIDs: []string{key.ID, "retained.widgets.sibling"}, ExpectedCount: 2}}}
	encoded, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	encoded = bytes.ReplaceAll(encoded, []byte(`"fixture:099F"`), []byte(`"acme:primary"`))
	if err := json.Unmarshal(encoded, &a); err != nil {
		t.Fatal(err)
	}
	inventory := loadRetainedSourceInventory(context.Background(), root, cohort)
	if len(inventory.Operations) != 2 || !inventory.Operations[1].Observed {
		t.Fatal("sibling inventory")
	}
	siblingFacts := normalizeSourceFacts(inventory.Operations[1], inventory.Documents[0], nil)
	b := sourceSemanticAnnotation{Key: inventory.Operations[1].Key, Citation: siblingFacts.Refs["summary"], Clause: sourceFactText(siblingFacts, "summary")}
	bRef := canonicalSourceLaneTargetRef(a.IntendedBindings[0])
	bRef.Lane = "direct_read"
	siblingAnchor := sourceBindingCitation099F(t, siblingFacts, "/rest/operations/1/source_operation/responses/200/content/application~1json/schema")
	siblingAnchor.DocumentID = "acme:primary"
	bRef.SourceSchema = &siblingAnchor
	bRef.FieldMappings[0].Source = sourceBindingCitation099F(t, siblingFacts, siblingAnchor.Pointer+"/items")
	bRef.FieldMappings[0].Source.DocumentID = "acme:primary"
	b.IntendedBindings = []sourceLaneTargetRef{bRef}
	before := sourceBindingFixtureSnapshot(t, root)
	got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 || got.SourceOperations[0].Source.Key != key || got.SourceOperations[1].Source.Key != b.Key {
		t.Fatal("admission exact keys/14 cells")
	}
	direct := requireSourceLane(t, got.SourceOperations[1].Lanes, "direct_read", "applicable")
	if len(direct.References) != 1 || !sourceLaneTargetRefEqual(direct.References[0], canonicalSourceLaneTargetRef(bRef)) {
		t.Errorf("valid unrelated direct reference lost: %+v", direct)
	}
	for _, row := range got.SourceOperations {
		for _, c := range row.Lanes {
			if c.State == "implemented" || len(c.ProofRefs) != 0 {
				t.Error("admission fabricated proof")
			}
		}
	}
	if !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, root)) {
		t.Error("admission mutated inputs")
	}
	return got, a
}

func TestSourceLane137CollectionScopesAndSemantics(t *testing.T) {
	for _, tc := range []struct{ name, method, first, second, status, want, rule, direct string }{
		{"collection_unknown_status", "GET", `{"type":"array","items":{"type":"object"}}`, `{}`, "201", "applicable", "source_record_collection", "applicable"},
		{"scalar_unknown_status", "GET", `{"type":"string"}`, `{}`, "201", "undetermined", "facts_unresolved", "applicable"},
		{"all_scalar_status", "GET", `{"type":"string"}`, `{"type":"number"}`, "201", "not_applicable", "fixed_noncollection_read", "applicable"},
		{"error_collection", "GET", `{"type":"string"}`, `{"type":"array","items":{"type":"object"}}`, "400", "not_applicable", "fixed_noncollection_read", "applicable"},
		{"mutation", "POST", `{"type":"array","items":{"type":"object"}}`, "", "", "not_applicable", "source_mutation", "not_applicable"},
		{"unknown_semantics", "CUSTOM", `{"type":"array","items":{"type":"object"}}`, "", "", "undetermined", "facts_unresolved", "undetermined"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, cohort, _, _, _ := sourceCollection122Fixture(t, tc.first, "")
			raw, err := os.ReadFile(filepath.Join(root, "source.json"))
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			for _, value := range doc["rest"].(map[string]any)["operations"].([]any) {
				row := value.(map[string]any)
				row["method"] = tc.method
				if tc.method == "POST" {
					row["source_operation"].(map[string]any)["summary"] = "Create a widget"
				}
				if tc.second != "" {
					var second any
					if err := json.Unmarshal([]byte(tc.second), &second); err != nil {
						t.Fatal(err)
					}
					row["source_operation"].(map[string]any)["responses"].(map[string]any)[tc.status] = map[string]any{"content": map[string]any{"application/json": map[string]any{"schema": second}}}
				}
			}
			raw, err = json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			proofWrite(t, root, "source.json", raw)
			cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 {
				t.Fatal("scope universe")
			}
			for i, row := range got.SourceOperations {
				wantKey := sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: []string{"source.a", "source.b"}[i]}
				if row.Source.Key != wantKey || !row.Source.Observed {
					t.Fatal("scope key")
				}
				requireSourceLane(t, row.Lanes, "direct_read", tc.direct)
				cell := requireSourceLane(t, row.Lanes, "etl", tc.want)
				if cell.RuleID != tc.rule {
					t.Errorf("scope rule=%s want=%s", cell.RuleID, tc.rule)
				}
				if tc.second == `{}` {
					p := fmt.Sprintf("/rest/operations/%d/source_operation/responses/201/content/application~1json/schema", i)
					if !collection137HasDiagnostic(cell, wantKey, "source_collection_scope_unknown", p, "deficit") {
						t.Errorf("lost distinct unknown scope: %+v", cell.Diagnostics)
					}
				}
				for _, c := range row.Lanes {
					if len(c.References) > 0 || len(c.ProofRefs) > 0 || c.State == "implemented" {
						t.Error("scope fabricated proof")
					}
				}
			}
		})
	}
}

func TestSourceLane137LocalShapeEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		typ       sourceLocalTypeState
		props     sourceLocalMemberState
		prefix    sourceLocalPrefixState
		object    sourceLocalObjectState
		coverage  sourceLocalArrayCoverage
	}{
		{"absent", `{}`, sourceLocalTypeAbsent, sourceLocalMemberAbsent, sourceLocalPrefixAbsent, sourceLocalObjectUnverified, sourceLocalArrayUnverified},
		{"explicit_null", `{"type":null}`, sourceLocalTypeUnsupported, sourceLocalMemberAbsent, sourceLocalPrefixAbsent, sourceLocalObjectUnverified, sourceLocalArrayUnverified},
		{"object", `{"type":"object"}`, sourceLocalTypeSupported, sourceLocalMemberAbsent, sourceLocalPrefixAbsent, sourceLocalObjectEstablished, sourceLocalArrayUnverified},
		{"object_null_properties", `{"type":"object","properties":null}`, sourceLocalTypeSupported, sourceLocalMemberMalformed, sourceLocalPrefixAbsent, sourceLocalObjectUnverified, sourceLocalArrayUnverified},
		{"inferred_object", `{"properties":{}}`, sourceLocalTypeAbsent, sourceLocalMemberValid, sourceLocalPrefixAbsent, sourceLocalObjectEstablished, sourceLocalArrayUnverified},
		{"scalar_properties", `{"type":"string","properties":{}}`, sourceLocalTypeSupported, sourceLocalMemberValid, sourceLocalPrefixAbsent, sourceLocalKnownNonobject, sourceLocalArrayUnverified},
		{"uniform", `{"type":"array","items":{}}`, sourceLocalTypeSupported, sourceLocalMemberAbsent, sourceLocalPrefixAbsent, sourceLocalKnownNonobject, sourceLocalArrayUniform},
		{"empty_prefix", `{"type":"array","prefixItems":[]}`, sourceLocalTypeSupported, sourceLocalMemberAbsent, sourceLocalPrefixEmpty, sourceLocalKnownNonobject, sourceLocalArrayUniform},
		{"prefix_tail", `{"type":"array","prefixItems":[{}]}`, sourceLocalTypeSupported, sourceLocalMemberAbsent, sourceLocalPrefixNonempty, sourceLocalKnownNonobject, sourceLocalArrayPrefixOrTail},
		{"malformed_prefix", `{"type":"array","prefixItems":null}`, sourceLocalTypeSupported, sourceLocalMemberAbsent, sourceLocalPrefixMalformed, sourceLocalKnownNonobject, sourceLocalArrayUnverified},
		{"composition", `{"type":"array","allOf":[]}`, sourceLocalTypeSupported, sourceLocalMemberAbsent, sourceLocalPrefixAbsent, sourceLocalObjectUnverified, sourceLocalArrayUnverified},
		{"no_reference_resolution", `{"$ref":"#/Object"}`, sourceLocalTypeAbsent, sourceLocalMemberAbsent, sourceLocalPrefixAbsent, sourceLocalObjectUnverified, sourceLocalArrayUnverified},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var node map[string]json.RawMessage
			if err := json.Unmarshal([]byte(tc.raw), &node); err != nil {
				t.Fatal(err)
			}
			before, err := json.Marshal(node)
			if err != nil {
				t.Fatal(err)
			}
			shape := sourceLocalShapeEvidence(node)
			if shape.TypeState != tc.typ || shape.PropertiesState != tc.props || shape.PrefixState != tc.prefix || shape.object() != tc.object || shape.arrayCoverage() != tc.coverage {
				t.Errorf("local state=%+v object=%v coverage=%v", shape, shape.object(), shape.arrayCoverage())
			}
			after, err := json.Marshal(node)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Error("local decoder mutated caller input")
			}
		})
	}
	// Existing resolution owns the only lookup and budget. The decoder cannot
	// follow this ref even though the real retained document makes it resolvable.
	_, _, _, facts, ref := sourceCollection122Fixture(t, `{"$ref":"#/Object"}`, `{"Object":{"type":"object"}}`)
	raw, err := sourceJSONPointer(facts.Document, ref.Pointer)
	if err != nil {
		t.Fatal(err)
	}
	var node map[string]json.RawMessage
	if err := json.Unmarshal(raw, &node); err != nil {
		t.Fatal(err)
	}
	if sourceLocalShapeEvidence(node).object() != sourceLocalObjectUnverified {
		t.Error("local decoder resolved a reference")
	}
	resolved, ok := sourceResolveObject(facts, raw, map[string]bool{}, 0)
	if !ok {
		t.Fatal("real retained resolver control")
	}
	before, err := json.Marshal(facts.analysis)
	if err != nil {
		t.Fatal(err)
	}
	if sourceLocalShapeEvidence(resolved).object() != sourceLocalObjectEstablished {
		t.Error("resolved local object lost")
	}
	after, err := json.Marshal(facts.analysis)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("local decoding incremented resolution budget/cache")
	}
}

func TestSourceLane137CollectionSuccessMedia(t *testing.T) {
	root, cohort, _, _, _ := sourceCollection122Fixture(t, `{"type":"array","items":{"type":"object"}}`, "")
	raw, err := os.ReadFile(filepath.Join(root, "source.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	for _, value := range doc["rest"].(map[string]any)["operations"].([]any) {
		content := value.(map[string]any)["source_operation"].(map[string]any)["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)
		content["application/vnd.unknown+json"] = map[string]any{"schema": map[string]any{}}
	}
	raw, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, root, "source.json", raw)
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	got, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	tc := collection137Case{applicability: "applicable"}
	if v := collection137Oracle(got, [2]collection137Case{tc, tc}); len(v) > 0 {
		t.Errorf("distinct media source oracle: %+v", v)
	}
	for i, row := range got.SourceOperations {
		pointer := fmt.Sprintf("/rest/operations/%d/source_operation/responses/200/content/application~1vnd.unknown+json/schema", i)
		if !collection137HasDiagnostic(row.Lanes[4], row.Source.Key, "source_collection_scope_unknown", pointer, "deficit") {
			t.Errorf("lost exact unknown media occurrence: %+v", row.Lanes[4].Diagnostics)
		}
	}
}
