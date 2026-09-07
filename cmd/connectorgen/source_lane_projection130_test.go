package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestSourceLane130PhysicalProjectionBuilder(t *testing.T) {
	for _, mode := range []string{"unique", "composition", "unresolved", "depth", "budget", "additionalProperties", "malformed_tuple"} {
		for _, direct := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/direct=%t", mode, direct), func(t *testing.T) {
				key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					var record any
					if err := json.Unmarshal([]byte(sourceBindingRecord099F), &record); err != nil {
						t.Fatal(err)
					}
					doc["source_contract"] = map[string]any{"components": map[string]any{"schemas": map[string]any{"Record": record}}}
					op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
					schema := op["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
					props := schema["properties"].(map[string]any)
					props["data"].(map[string]any)["items"] = map[string]any{"$ref": "#/components/schemas/Record"}
					switch mode {
					case "additionalProperties":
						props["other"] = map[string]any{"type": "object", "additionalProperties": map[string]any{"$ref": "#/components/schemas/Record"}}
					case "malformed_tuple":
						props["other"] = map[string]any{"type": "array", "prefixItems": map[string]any{"$ref": "#/components/schemas/Record"}}
					case "composition":
						props["other"] = map[string]any{"anyOf": []any{map[string]any{"$ref": "#/components/schemas/Record"}, map[string]any{"type": "null"}}}
					case "unresolved":
						props["other"] = map[string]any{"$ref": "https://example.invalid/unresolved"}
					case "depth":
						components := doc["source_contract"].(map[string]any)["components"].(map[string]any)["schemas"].(map[string]any)
						for i := 0; i < 129; i++ {
							next := fmt.Sprintf("Link%d", i+1)
							if i == 128 {
								next = "Record"
							}
							components[fmt.Sprintf("Link%d", i)] = map[string]any{"$ref": "#/components/schemas/" + next}
						}
						props["other"] = map[string]any{"$ref": "#/components/schemas/Link0"}
					case "budget":
						for i := 0; i < 4100; i++ {
							props[fmt.Sprintf("z%04d", i)] = map[string]any{"type": "string"}
						}
						props["zzsecond"] = map[string]any{"$ref": "#/components/schemas/Record"}
					}
				})
				anchor := sourceBindingCitation099F(t, facts, a.IntendedBindings[0].SourceSchema.Pointer)
				a.IntendedBindings[0].SourceSchema = &anchor
				pointer := "/source_contract/components/schemas/Record"
				if direct {
					pointer = anchor.Pointer + "/properties/data/items"
				}
				a.IntendedBindings[0].FieldMappings[0].Source = sourceBindingCitation099F(t, facts, pointer)
				a.ResponseInterpretations[0].ResponseSchema = anchor
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
				encoded, _ := json.Marshal(a)
				encoded = bytes.ReplaceAll(encoded, []byte(`"fixture:099F"`), []byte(`"acme:primary"`))
				if err := json.Unmarshal(encoded, &a); err != nil {
					t.Fatal(err)
				}
				before := sourceBindingFixtureSnapshot(t, root)
				got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
				if err != nil {
					t.Fatal(err)
				}
				if mode == "budget" {
					again, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
					if err != nil {
						t.Fatal(err)
					}
					firstBytes, _ := json.Marshal(got)
					secondBytes, _ := json.Marshal(again)
					if !bytes.Equal(firstBytes, secondBytes) {
						t.Fatal("same-byte budget search changed complete report")
					}
				}
				if got.SourceTotals.Operations != 1 || got.SourceTotals.Cells != 7 || got.SourceOperations[0].Source.Key != key || !got.SourceOperations[0].Source.Observed {
					t.Fatal("did not reach exact retained builder")
				}
				if !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, root)) {
					t.Fatal("builder modified retained/admitted artifacts")
				}
				row := got.SourceOperations[0]
				requireSourceLane(t, row.Lanes, "direct_read", "applicable")
				cell := requireSourceLane(t, row.Lanes, "etl", "applicable")
				blocked := mode != "unique" && !direct
				want := 1
				if blocked {
					want = 0
				}
				if len(cell.References) != want {
					t.Errorf("actual admitted builder accepted=%d want=%d: %+v", len(cell.References), want, cell.Diagnostics)
				}
				if blocked {
					found := false
					for _, d := range cell.Diagnostics {
						found = found || (d.Key == key && d.Stage == "reference" && d.Pointer == pointer && (d.Code == "source_schema_unverified" || d.Code == "source_projection_ambiguous"))
					}
					if !found {
						t.Errorf("missing exact reference refusal: %+v", cell.Diagnostics)
					}
				}
				// Reviewed input at the reducer boundary is independently supplied;
				// this fixture never claims actual provider/ETL certification.
				record := sourceLaneProofRecord{ID: "130-reviewed-boundary", Key: key, Lane: "etl", Targets: []sourceLaneTargetRef{canonicalSourceLaneTargetRef(a.IntendedBindings[0])}}
				inputs := sourceLaneProofInputs{byCell: map[sourceLaneProofCell]sourceLaneProofRecord{{key, "etl"}: record}}
				assessed := requireSourceLane(t, assessSourceLaneProof(key, row.Lanes, inputs), "etl", "applicable")
				if (assessed.State == "implemented") == blocked {
					t.Errorf("proof boundary ignored source ownership: %+v", assessed)
				}
				if blocked && len(assessed.ProofRefs) != 0 {
					t.Error("unverified binding acquired proof")
				}
			})
		}
	}
}

func TestSourceLane130BodyProjectionConsumer(t *testing.T) {
	for _, composition := range []bool{false, true} {
		t.Run(fmt.Sprint(composition), func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "body", nil)
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				schema := op["requestBody"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
				props := schema["properties"].(map[string]any)
				doc["source_contract"] = map[string]any{"components": map[string]any{"schemas": map[string]any{"Data": props["data"]}}}
				props["data"] = map[string]any{"$ref": "#/components/schemas/Data"}
				if composition {
					props["other"] = map[string]any{"oneOf": []any{map[string]any{"$ref": "#/components/schemas/Data"}}}
				}
			})
			anchor := sourceBindingCitation099F(t, facts, a.IntendedBindings[0].SourceSchema.Pointer)
			a.IntendedBindings[0].SourceSchema = &anchor
			a.IntendedBindings[0].FieldMappings[1].Source = sourceBindingCitation099F(t, facts, "/source_contract/components/schemas/Data")
			want := 1
			codes := []string{}
			if composition {
				want = 0
				codes = []string{"source_schema_unverified"}
			}
			sourceBindingOutcome099F(t, key, facts, a, want, codes...)
		})
	}
}

func TestSourceLane130GraphQLProjectionConsumer(t *testing.T) {
	// This is the real typed variable-placement consumer; it runs only after
	// canonical admission and engine bundle loading, with independently supplied
	// source variable and actual loaded target variable coordinates.
	for _, composition := range []bool{false, true} {
		t.Run(fmt.Sprint(composition), func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "body", func(lock *vNextSourceLock) {
				schema := `{"type":"object","properties":{"id":{"type":"string"}},"required":["id"],"additionalProperties":false}`
				lock.Schemas["schemas/request.json"] = json.RawMessage(schema)
				lock.Operations[0].Write = json.RawMessage(`{"name":"create_widget","kind":"create","method":"POST","path":"/graphql","body_type":"graphql","graphql":{"document":"mutation Change($id: String!) { change(id: $id) { id } }","operation_name":"Change","variables":{"id":"{{ record.id }}"}},"record_schema":` + schema + `,"risk":"low"}`)
			})
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				doc["source_contract"] = map[string]any{"components": map[string]any{"schemas": map[string]any{"ID": map[string]any{"type": "string"}}}}
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				schema := map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"$ref": "#/components/schemas/ID"}}, "required": []any{"id"}}
				if composition {
					schema["anyOf"] = []any{map[string]any{"properties": map[string]any{"id": map[string]any{"$ref": "#/components/schemas/ID"}}}}
				}
				op["variables_schema"] = schema
			})
			anchor := sourceBindingCitation099F(t, facts, "/rest/operations/0/source_operation/variables_schema")
			ref := a.IntendedBindings[0]
			ref.SourceSchema = &anchor
			pointer := "/properties/id"
			ref.FieldMappings = []sourceLaneFieldMapping{{Source: sourceBindingCitation099F(t, facts, "/source_contract/components/schemas/ID"), Target: sourceLaneFieldTarget{Kind: sourceLaneFieldSchema, Pointer: &pointer}}}
			var action engine.WriteAction
			var writes struct {
				Actions []engine.WriteAction `json:"actions"`
			}
			if err := json.Unmarshal(facts.bindings.Artifacts[ref.Artifact], &writes); err != nil || len(writes.Actions) != 1 {
				t.Fatal("actual admitted write missing", err)
			}
			action = writes.Actions[0]
			if action.GraphQL == nil {
				t.Fatal("actual GraphQL consumer absent")
			}
			issues := sourceLaneGraphQLVariablesContract(facts, ref, *action.GraphQL)
			if composition && len(issues) == 0 {
				t.Error("actual GraphQL variable consumer accepted incomplete physical ownership")
			}
			if !composition && len(issues) != 0 {
				t.Fatalf("unique positive did not reach variable placement: %+v", issues)
			}
		})
	}
}

func TestSourceLane130NestedDefinitionProjection(t *testing.T) {
	for _, duplicate := range []bool{false, true} {
		for _, occurrence := range []bool{false, true} {
			t.Run(fmt.Sprintf("duplicate=%t/occurrence=%t", duplicate, occurrence), func(t *testing.T) {
				schema := map[string]any{
					"type": "object", "$defs": map[string]any{"Record": map[string]any{"type": "object"}},
					"properties": map[string]any{"data": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/$defs/Record"}}},
				}
				if duplicate {
					schema["properties"].(map[string]any)["other"] = map[string]any{"$ref": "#/$defs/Record"}
				}
				raw, err := json.Marshal(schema)
				if err != nil {
					t.Fatal(err)
				}
				pointer := "/$defs/Record"
				if occurrence {
					pointer = "/properties/data/items"
				}
				projection, code := sourceLaneSchemaProjection(sourceFacts{Document: raw}, sourceFactRef{Pointer: ""}, pointer)
				if duplicate && !occurrence {
					if code != "source_projection_ambiguous" {
						t.Errorf("physical nested definition did not retain ambiguity: %s", code)
					}
				} else if code != "" || !reflect.DeepEqual(projection.Path, []string{"data", "[]"}) {
					t.Errorf("valid local use-site projection lost: code=%s path=%v", code, projection.Path)
				}
				// The actual effective-target projection consumer uses the same
				// raw-root/local-ref shape; definition storage is not a coordinate.
				target, targetCode := sourceLaneTargetProjection(raw, pointer)
				if duplicate && !occurrence {
					if targetCode == "" {
						t.Error("ambiguous target definition accepted")
					}
				} else if targetCode != "" || !reflect.DeepEqual(target.Path, []string{"data", "[]"}) {
					t.Errorf("effective target local projection lost: code=%s path=%v", targetCode, target.Path)
				}
			})
		}
	}
}
