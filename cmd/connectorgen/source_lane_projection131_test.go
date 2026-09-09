package main

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSourceLaneOwnership131SelectedAncestor(t *testing.T) {
	for _, mode := range []string{"positive", "unrelated_unknown", "root_allOf", "array_anyOf", "root_scalar", "malformed_required"} {
		t.Run(mode, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				schema := op["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
				switch mode {
				case "unrelated_unknown":
					schema["properties"].(map[string]any)["other"] = map[string]any{"$ref": "https://example.invalid/unresolved"}
				case "root_allOf":
					schema["allOf"] = []any{map[string]any{"$ref": "https://example.invalid/unresolved"}}
				case "array_anyOf":
					schema["properties"].(map[string]any)["data"].(map[string]any)["anyOf"] = []any{map[string]any{"$ref": "https://example.invalid/unresolved"}}
				case "root_scalar":
					schema["type"] = "string"
				case "malformed_required":
					schema["required"] = "data"
				}
			})
			a.IntendedBindings[0].Lane = "direct_read"
			a.ResponseInterpretations = nil
			anchor := sourceBindingCitation099F(t, facts, a.IntendedBindings[0].SourceSchema.Pointer)
			a.IntendedBindings[0].SourceSchema = &anchor
			pointer := anchor.Pointer + "/properties/data/items"
			a.IntendedBindings[0].FieldMappings[0].Source = sourceBindingCitation099F(t, facts, pointer)
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
			doc["schema_version"], doc["connector"], doc["counts"] = 2, "acme", map[string]any{"total": 1}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			proofWrite(t, root, "source.json", raw)
			cohort := sourceLaneCohort{SchemaVersion: 1, CohortID: "fixture", Inventories: []sourceInventoryAnchor{{Connector: "acme", Inventory: "primary", Class: "primary", Path: "source.json", SHA256: sourceBytesHash(raw), ExpectedIDs: []string{key.ID}, ExpectedCount: 1}}}
			encoded, err := json.Marshal(a)
			if err != nil {
				t.Fatal(err)
			}
			encoded = bytes.ReplaceAll(encoded, []byte(`"fixture:099F"`), []byte(`"acme:primary"`))
			if err := json.Unmarshal(encoded, &a); err != nil {
				t.Fatal(err)
			}
			before := sourceBindingFixtureSnapshot(t, root)
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
			if err != nil {
				t.Fatal(err)
			}
			if got.SourceTotals.Operations != 1 || got.SourceTotals.Cells != 7 || got.SourceOperations[0].Source.Key != key || !got.SourceOperations[0].Source.Observed {
				t.Fatal("actual retained builder frontier missing")
			}
			if !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, root)) {
				t.Fatal("retained/admitted source mutated")
			}
			row := got.SourceOperations[0]
			cell := requireSourceLane(t, row.Lanes, "direct_read", "applicable")
			blocked := mode != "positive" && mode != "unrelated_unknown"
			want := 1
			if blocked {
				want = 0
			}
			t.Logf("reached actual admitted builder: mode=%s accepted=%d diagnostics=%+v", mode, len(cell.References), cell.Diagnostics)
			ownership132Diagnostic(t, key, cell, a.IntendedBindings[0], mode)
			if len(cell.References) != want {
				t.Errorf("selected ancestor insufficiently established: accepted=%d want=%d", len(cell.References), want)
			}
			record := sourceLaneProofRecord{ID: "131-independent-reducer-boundary", Key: key, Lane: "direct_read", Targets: []sourceLaneTargetRef{canonicalSourceLaneTargetRef(a.IntendedBindings[0])}}
			inputs := sourceLaneProofInputs{byCell: map[sourceLaneProofCell]sourceLaneProofRecord{{key, "direct_read"}: record}}
			assessed := requireSourceLane(t, assessSourceLaneProof(key, row.Lanes, inputs), "direct_read", "applicable")
			if blocked && (assessed.State == "implemented" || len(assessed.ProofRefs) != 0) {
				t.Errorf("incomplete selected-ancestor binding promoted by independent reducer input: state=%s proofs=%v", assessed.State, assessed.ProofRefs)
			}
		})
	}
}

func TestSourceLaneOwnership131ScalarEdgeAndTuple(t *testing.T) {
	for _, mode := range []string{"physical_unique", "physical_unsupported_scalar_ref", "items_with_prefix", "prefix_empty"} {
		t.Run(mode, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				schema := op["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
				if mode == "physical_unique" || mode == "physical_unsupported_scalar_ref" {
					var record any
					if err := json.Unmarshal([]byte(sourceBindingRecord099F), &record); err != nil {
						t.Fatal(err)
					}
					doc["source_contract"] = map[string]any{"components": map[string]any{"schemas": map[string]any{"Record": record}}}
					schema["properties"].(map[string]any)["data"].(map[string]any)["items"] = map[string]any{"$ref": "#/components/schemas/Record"}
					if mode == "physical_unsupported_scalar_ref" {
						schema["properties"].(map[string]any)["other"] = map[string]any{"$dynamicRef": "#/components/schemas/Record"}
					}
				} else {
					prefix := []any{}
					if mode == "items_with_prefix" {
						prefix = []any{map[string]any{"type": "string"}}
					}
					schema["properties"].(map[string]any)["data"].(map[string]any)["prefixItems"] = prefix
				}
			})
			a.IntendedBindings[0].Lane = "direct_read"
			a.ResponseInterpretations = nil
			anchor := sourceBindingCitation099F(t, facts, a.IntendedBindings[0].SourceSchema.Pointer)
			a.IntendedBindings[0].SourceSchema = &anchor
			pointer := anchor.Pointer + "/properties/data/items"
			if mode == "physical_unique" || mode == "physical_unsupported_scalar_ref" {
				pointer = "/source_contract/components/schemas/Record"
			}
			a.IntendedBindings[0].FieldMappings[0].Source = sourceBindingCitation099F(t, facts, pointer)
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
			doc["schema_version"], doc["connector"], doc["counts"] = 2, "acme", map[string]any{"total": 1}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			proofWrite(t, root, "source.json", raw)
			cohort := sourceLaneCohort{SchemaVersion: 1, CohortID: "fixture", Inventories: []sourceInventoryAnchor{{Connector: "acme", Inventory: "primary", Class: "primary", Path: "source.json", SHA256: sourceBytesHash(raw), ExpectedIDs: []string{key.ID}, ExpectedCount: 1}}}
			encoded, err := json.Marshal(a)
			if err != nil {
				t.Fatal(err)
			}
			encoded = bytes.ReplaceAll(encoded, []byte(`"fixture:099F"`), []byte(`"acme:primary"`))
			if err := json.Unmarshal(encoded, &a); err != nil {
				t.Fatal(err)
			}
			before := sourceBindingFixtureSnapshot(t, root)
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
			if err != nil {
				t.Fatal(err)
			}
			if got.SourceTotals.Operations != 1 || got.SourceTotals.Cells != 7 || got.SourceOperations[0].Source.Key != key || !got.SourceOperations[0].Source.Observed {
				t.Fatal("actual retained builder frontier missing")
			}
			if !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, root)) {
				t.Fatal("retained/admitted source mutated")
			}
			row := got.SourceOperations[0]
			cell := requireSourceLane(t, row.Lanes, "direct_read", "applicable")
			blocked := mode == "physical_unsupported_scalar_ref" || mode == "items_with_prefix"
			want := 1
			if blocked {
				want = 0
			}
			t.Logf("reached actual admitted builder: mode=%s accepted=%d diagnostics=%+v", mode, len(cell.References), cell.Diagnostics)
			ownership132Diagnostic(t, key, cell, a.IntendedBindings[0], mode)
			if len(cell.References) != want {
				t.Errorf("selected ancestor insufficiently established: accepted=%d want=%d", len(cell.References), want)
			}
			record := sourceLaneProofRecord{ID: "131-independent-reducer-boundary", Key: key, Lane: "direct_read", Targets: []sourceLaneTargetRef{canonicalSourceLaneTargetRef(a.IntendedBindings[0])}}
			inputs := sourceLaneProofInputs{byCell: map[sourceLaneProofCell]sourceLaneProofRecord{{key, "direct_read"}: record}}
			assessed := requireSourceLane(t, assessSourceLaneProof(key, row.Lanes, inputs), "direct_read", "applicable")
			if blocked && (assessed.State == "implemented" || len(assessed.ProofRefs) != 0) {
				t.Errorf("incomplete selected-ancestor binding promoted by independent reducer input: state=%s proofs=%v", assessed.State, assessed.ProofRefs)
			}
		})
	}
}

func ownership131Build(t *testing.T, key sourceOperationKey, facts sourceFacts, a sourceSemanticAnnotation) (sourceLaneManifest, sourceSemanticAnnotation) {
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
	doc["schema_version"], doc["connector"], doc["counts"] = 2, "acme", map[string]any{"total": 1}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, root, "source.json", raw)
	cohort := sourceLaneCohort{SchemaVersion: 1, CohortID: "fixture", Inventories: []sourceInventoryAnchor{{Connector: "acme", Inventory: "primary", Class: "primary", Path: "source.json", SHA256: sourceBytesHash(raw), ExpectedIDs: []string{key.ID}, ExpectedCount: 1}}}
	encoded, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	encoded = bytes.ReplaceAll(encoded, []byte(`"fixture:099F"`), []byte(`"acme:primary"`))
	if err := json.Unmarshal(encoded, &a); err != nil {
		t.Fatal(err)
	}
	before := sourceBindingFixtureSnapshot(t, root)
	got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceTotals.Operations != 1 || got.SourceTotals.Cells != 7 || got.SourceOperations[0].Source.Key != key || !got.SourceOperations[0].Source.Observed {
		t.Fatal("retained admitted builder frontier absent")
	}
	if !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, root)) {
		t.Fatal("retained/admitted fixture mutated")
	}
	again, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
	if err != nil {
		t.Fatal(err)
	}
	first, _ := json.Marshal(got)
	second, _ := json.Marshal(again)
	if !bytes.Equal(first, second) {
		t.Fatal("identical input changed complete builder report")
	}
	if !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, root)) {
		t.Fatal("repeat mutated input")
	}
	return got, a
}
func ownership131Check(t *testing.T, key sourceOperationKey, got sourceLaneManifest, a sourceSemanticAnnotation, blocked bool) {
	t.Helper()
	lane := a.IntendedBindings[0].Lane
	cell := requireSourceLane(t, got.SourceOperations[0].Lanes, lane, "applicable")
	want := 1
	if blocked {
		want = 0
	}
	t.Logf("actual builder lane=%s accepted=%d diagnostics=%+v", lane, len(cell.References), cell.Diagnostics)
	if len(cell.References) != want {
		t.Errorf("unsupported selected lineage accepted=%d want=%d", len(cell.References), want)
	}
	record := sourceLaneProofRecord{ID: "131-independent-graph-ref-reducer", Key: key, Lane: lane, Targets: []sourceLaneTargetRef{canonicalSourceLaneTargetRef(a.IntendedBindings[0])}}
	inputs := sourceLaneProofInputs{byCell: map[sourceLaneProofCell]sourceLaneProofRecord{{key, lane}: record}}
	assessed := requireSourceLane(t, assessSourceLaneProof(key, got.SourceOperations[0].Lanes, inputs), lane, "applicable")
	if !blocked && len(cell.References) == 1 && !sourceLaneTargetRefEqual(cell.References[0], canonicalSourceLaneTargetRef(a.IntendedBindings[0])) {
		t.Error("accepted reference identity changed")
	}
	if blocked && (assessed.State == "implemented" || len(assessed.ProofRefs) != 0) {
		t.Errorf("unsupported lineage promoted: state=%s proofs=%v", assessed.State, assessed.ProofRefs)
	}
}
func TestSourceLaneOwnership131GraphQLRoot(t *testing.T) {
	for _, composition := range []bool{false, true} {
		t.Run(map[bool]string{false: "positive", true: "composed_ancestor"}[composition], func(t *testing.T) {
			const document = "mutation Change($id: String!) { change(id: $id) { id } }"
			key, facts, a := sourceBindingFixture099F(t, "body", func(lock *vNextSourceLock) {
				schema := `{"type":"object","properties":{"id":{"type":"string"}},"required":["id"],"additionalProperties":false}`
				lock.Schemas["schemas/request.json"] = json.RawMessage(schema)
				lock.Operations[0].Write = json.RawMessage(`{"name":"create_widget","kind":"create","method":"POST","path":"/graphql","body_type":"graphql","graphql":{"document":"` + document + `","operation_name":"Change","variables":{"id":"{{ record.id }}"}},"record_schema":` + schema + `,"risk":"low"}`)
			})
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				row := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)
				row["path"] = "/graphql"
				op := row["source_operation"].(map[string]any)
				delete(op, "parameters")
				delete(op, "requestBody")
				op["operation_name"], op["document"] = "Change", document
				schema := map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}, "required": []any{"id"}, "additionalProperties": false}
				if composition {
					schema["allOf"] = []any{map[string]any{"$ref": "https://example.invalid/unresolved"}}
				}
				op["variables_schema"] = schema
			})
			base := "/rest/operations/0/source_operation"
			anchor := sourceBindingCitation099F(t, facts, base+"/variables_schema")
			opName := sourceBindingCitation099F(t, facts, base+"/operation_name")
			docRef := sourceBindingCitation099F(t, facts, base+"/document")
			a.GraphQL = &sourceLaneGraphQLRefs{OperationName: &opName, Document: &docRef, RequestSchema: &anchor}
			ref := &a.IntendedBindings[0]
			ref.SourceSchema = &anchor
			pointer := "/properties/id"
			ref.FieldMappings = []sourceLaneFieldMapping{{Source: sourceBindingCitation099F(t, facts, anchor.Pointer+"/properties/id"), Target: sourceLaneFieldTarget{Kind: sourceLaneFieldSchema, Pointer: &pointer}}}
			got, rebased := ownership131Build(t, key, facts, a)
			ownership131Check(t, key, got, rebased, composition)
			if composition {
				ownership132RequireDiagnostic(t, key, got, rebased, "source_schema_unverified", "deficit")
			}
		})
	}
}
func TestSourceLaneOwnership131LocalRefEscape(t *testing.T) {
	for _, invalid := range []bool{false, true} {
		t.Run(map[bool]string{false: "valid_tilde", true: "invalid_tilde"}[invalid], func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				var record any
				if err := json.Unmarshal([]byte(sourceBindingRecord099F), &record); err != nil {
					t.Fatal(err)
				}
				doc["source_contract"] = map[string]any{"components": map[string]any{"schemas": map[string]any{"R~2": record}}}
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				schema := op["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
				edge := "#/components/schemas/R~02"
				if invalid {
					edge = "#/components/schemas/R~2"
				}
				schema["properties"].(map[string]any)["data"].(map[string]any)["items"] = map[string]any{"$ref": edge}
			})
			a.ResponseInterpretations = nil
			a.IntendedBindings[0].Lane = "direct_read"
			ref := &a.IntendedBindings[0]
			anchor := sourceBindingCitation099F(t, facts, ref.SourceSchema.Pointer)
			ref.SourceSchema = &anchor
			ref.FieldMappings[0].Source = sourceBindingCitation099F(t, facts, anchor.Pointer+"/properties/data/items")
			got, rebased := ownership131Build(t, key, facts, a)
			ownership131Check(t, key, got, rebased, invalid)
			if invalid {
				ownership132RequireDiagnostic(t, key, got, rebased, "source_schema_unverified", "deficit")
			}
		})
	}
}
func TestSourceLaneOwnership131SnapshotOracleControl(t *testing.T) {
	root := t.TempDir()
	expected := `{"expected":1}`
	proofWrite(t, root, "witness.json", []byte(expected))
	before := sourceBindingFixtureSnapshot(t, root)
	if before["witness.json"] != expected {
		t.Fatal("oracle did not retain independently supplied bytes")
	}
	proofWrite(t, root, "witness.json", []byte(`{"expected":2}`))
	after := sourceBindingFixtureSnapshot(t, root)
	if reflect.DeepEqual(before, after) {
		t.Fatal("snapshot oracle accepted deliberately wrong readable content")
	}
	proofWrite(t, root, "witness.json", []byte(expected))
	if !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, root)) {
		t.Fatal("fixture-owned restoration did not restore retained bytes")
	}
}
