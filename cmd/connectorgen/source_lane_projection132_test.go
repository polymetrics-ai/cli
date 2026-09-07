package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

// These assertions extend the immutable131 diagnostics. Original names map from
// TestCP12Ownership131* to TestSourceLaneOwnership131* for the full source suite.
func ownership132Diagnostic(t *testing.T, key sourceOperationKey, cell sourceLaneCell, ref sourceLaneTargetRef, mode string) {
	t.Helper()
	code, severity := "", "deficit"
	switch mode {
	case "root_allOf", "array_anyOf", "malformed_required", "physical_unsupported_scalar_ref":
		code = "source_schema_unverified"
	case "root_scalar":
		code = "source_binding_scope_mismatch"
		severity = "error"
	case "items_with_prefix":
		code = "target_record_projection_unverified"
	}
	if code != "" {
		ownership132CellDiagnostic(t, key, cell, ref.FieldMappings[0].Source.Pointer, code, severity)
	} else if len(cell.References) == 1 && !sourceLaneTargetRefEqual(cell.References[0], canonicalSourceLaneTargetRef(ref)) {
		t.Error("valid reference identity changed")
	}
}
func ownership132CellDiagnostic(t *testing.T, key sourceOperationKey, cell sourceLaneCell, pointer, code, severity string) {
	t.Helper()
	found := false
	for _, d := range cell.Diagnostics {
		if d.Code == code && d.Pointer == pointer && d.Stage == "reference" && d.Key == key && reflect.DeepEqual(d.Lanes, []string{cell.Lane}) && d.Severity == severity {
			found = true
		}
	}
	if !found {
		t.Errorf("missing exact %s/%s at %s: %+v", code, severity, pointer, cell.Diagnostics)
	}
}
func ownership132RequireDiagnostic(t *testing.T, key sourceOperationKey, got sourceLaneManifest, a sourceSemanticAnnotation, code, severity string) {
	t.Helper()
	ref := a.IntendedBindings[0]
	cell := requireSourceLane(t, got.SourceOperations[0].Lanes, ref.Lane, "applicable")
	ownership132CellDiagnostic(t, key, cell, ref.FieldMappings[0].Source.Pointer, code, severity)
	if code == "source_schema_unverified" {
		for _, d := range cell.Diagnostics {
			if d.Code == "target_graphql_variable_mismatch" {
				t.Error("unknown lineage relabeled as known GraphQL mismatch")
			}
		}
	}
}
func ownership132Response(doc map[string]any) map[string]any {
	return doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
}
func TestSourceLane132StorageBuilder(t *testing.T) {
	for _, registry := range []string{"$defs", "definitions", "components"} {
		for _, mode := range []string{"unique", "duplicate", "literal", "unused_unknown", "malformed_registry", "malformed_selected", "root_ref"} {
			t.Run(registry+"/"+mode, func(t *testing.T) {
				key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
				anchorPointer := a.IntendedBindings[0].SourceSchema.Pointer
				selected := anchorPointer + "/" + registry + "/Record"
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					schema := ownership132Response(doc)
					var envelope map[string]any
					encoded, _ := json.Marshal(schema)
					if err := json.Unmarshal(encoded, &envelope); err != nil {
						t.Fatal(err)
					}
					var record any
					if err := json.Unmarshal([]byte(sourceBindingRecord099F), &record); err != nil {
						t.Fatal(err)
					}
					storage := map[string]any{"Record": record}
					contract := map[string]any{"Envelope": envelope}
					selected = "/source_contract/Envelope/" + registry + "/Record"
					local := "#/Envelope/" + registry + "/Record"
					if registry == "components" {
						contract["components"] = map[string]any{"schemas": storage}
						selected = "/source_contract/components/schemas/Record"
						local = "#/components/schemas/Record"
					} else {
						envelope[registry] = storage
					}
					props := envelope["properties"].(map[string]any)
					props["data"].(map[string]any)["items"] = map[string]any{"$ref": local}
					if mode == "duplicate" {
						props["other"] = map[string]any{"$ref": local}
					}
					if mode == "unused_unknown" {
						storage["Unused"] = map[string]any{"$dynamicRef": "https://example.invalid/unknown"}
					}
					if mode == "malformed_selected" {
						storage["Record"] = []any{record}
					}
					if mode == "malformed_registry" {
						props["data"].(map[string]any)["items"] = record
						envelope["$defs"] = []any{record}
						selected = "/source_contract/Envelope/properties/data/items"
					}
					if mode == "literal" {
						selected = "/source_contract/Envelope/properties/data/items"
					}
					for k := range schema {
						delete(schema, k)
					}
					schema["$ref"] = "#/Envelope"
					if mode == "root_ref" {
						contract["Outer"] = map[string]any{"$ref": "#/Envelope"}
						schema["$ref"] = "#/Outer"
					}
					doc["source_contract"] = contract
				})
				a.ResponseInterpretations = nil
				ref := &a.IntendedBindings[0]
				ref.Lane = "direct_read"
				anchor := sourceBindingCitation099F(t, facts, anchorPointer)
				ref.SourceSchema = &anchor
				ref.FieldMappings[0].Source = sourceBindingCitation099F(t, facts, selected)
				got, rebased := ownership131Build(t, key, facts, a)
				blocked := mode == "duplicate" || mode == "malformed_registry" || mode == "malformed_selected"
				ownership131Check(t, key, got, rebased, blocked)
				if blocked {
					code := "source_schema_unverified"
					severity := "deficit"
					if mode == "duplicate" {
						code = "source_projection_ambiguous"
						severity = "error"
					}
					ownership132RequireDiagnostic(t, key, got, rebased, code, severity)
				}
			})
		}
	}
}

func TestSourceLane132PointersAndConsumers(t *testing.T) {
	for _, prefix := range []string{"", "/source_contract"} {
		for _, tc := range []struct {
			name, root, wanted, code string
			path                     []string
		}{
			{"slash", `{"type":"object","properties":{"a/b":{"type":"string"}},"required":["a/b"]}`, "/properties/a~1b", "", []string{"a/b"}},
			{"tilde", `{"type":"object","properties":{"x~y":{"type":"string"}}}`, "/properties/x~0y", "", []string{"x~y"}},
			{"empty", `{"type":"object","properties":{"":{"type":"string"}},"required":[""]}`, "/properties/", "", []string{""}},
			{"fixed", `{"type":"array","prefixItems":[{"type":"string"}]}`, "/prefixItems/0", "", []string{"[0]"}},
			{"fixed_alias", `{"type":"array","prefixItems":[{"type":"string"}]}`, "/prefixItems/00", "target_schema_pointer_mismatch", nil},
			{"fixed_absent", `{"type":"array","prefixItems":[{"type":"string"}]}`, "/prefixItems/1", "target_schema_pointer_mismatch", nil},
			{"target_unknown", `{"type":"object","allOf":[{}],"properties":{"id":{"type":"string"}}}`, "/properties/id", "target_schema_unverified", nil},
			{"target_scalar", `{"type":"string","properties":{"id":{"type":"string"}}}`, "/properties/id", "target_schema_pointer_mismatch", nil},
			{"bad_escape", `{"type":"object","properties":{"R~2":{"type":"string"}}}`, "/properties/R~2", "target_schema_pointer_mismatch", nil},
		} {
			t.Run(prefix+"/"+tc.name, func(t *testing.T) {
				raw := json.RawMessage(tc.root)
				actual, code := sourceLaneTargetProjection(raw, tc.wanted)
				if code != tc.code {
					t.Errorf("target consumer code=%s want=%s", code, tc.code)
				}
				if code == "" && !reflect.DeepEqual(actual.Path, tc.path) {
					t.Errorf("actual coordinate=%v want=%v", actual.Path, tc.path)
				}
				doc := raw
				if prefix != "" {
					doc = json.RawMessage(`{"source_contract":` + tc.root + `}`)
				}
				if tc.code == "" {
					p, c := sourceLaneSchemaProjection(sourceFacts{Document: doc, RefPrefix: prefix}, sourceFactRef{Pointer: prefix}, prefix+tc.wanted)
					if c != "" || !reflect.DeepEqual(p.Path, tc.path) {
						t.Errorf("source=%+v code=%s", p, c)
					}
				}
			})
		}
	}
	for _, prefix := range []string{"", "/source_contract"} {
		for _, edge := range []string{"#/defs/R~02", "#/defs/R~2", "#/defs/missing", "https://example.invalid/schema", "#/defs/Cycle"} {
			t.Run(prefix+edge, func(t *testing.T) {
				root := `{"defs":{"R~2":{"type":"string"},"Cycle":{"$ref":"#/defs/Cycle"}},"schema":{"$ref":` + fmt.Sprintf("%q", edge) + `}}`
				doc := json.RawMessage(root)
				if prefix != "" {
					doc = json.RawMessage(`{"source_contract":` + root + `}`)
				}
				facts := sourceFacts{Document: doc, RefPrefix: prefix}
				raw := json.RawMessage(`{"$ref":` + fmt.Sprintf("%q", edge) + `}`)
				_, ok := sourceResolveObject(facts, raw, map[string]bool{}, 0)
				if ok != (edge == "#/defs/R~02") {
					t.Errorf("resolver accepted=%t edge=%s", ok, edge)
				}
				p, c := sourceLaneSchemaProjection(facts, sourceFactRef{Pointer: prefix + "/schema"}, prefix+"/schema")
				if (c == "") != (edge == "#/defs/R~02") {
					t.Errorf("selected root=%+v code=%s", p, c)
				}
			})
		}
	}
	for _, schema := range []string{
		`{"type":"array","prefixItems":[{"type":"string"}],"items":` + sourceBindingRecord099F + `}`,
		`{"type":"array","prefixItems":[],"items":` + sourceBindingRecord099F + `}`,
	} {
		t.Run("record/"+schema, func(t *testing.T) {
			_, code := sourceLaneRecordCoordinate(sourceFacts{Document: json.RawMessage(schema)}, json.RawMessage(schema), engine.StreamSpec{})
			want := ""
			if !strings.Contains(schema, `"prefixItems":[]`) {
				want = "target_record_projection_unverified"
			}
			if code != want {
				t.Errorf("record consumer=%s want=%s", code, want)
			}
		})
	}
}

func TestSourceLane132SiblingIdentityAndSeverity(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		for _, materialized := range []bool{false, true} {
			t.Run(fmt.Sprintf("reverse=%t/materialized=%t", reverse, materialized), func(t *testing.T) {
				key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
				facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
					schema := ownership132Response(doc)
					schema["properties"].(map[string]any)["other"] = map[string]any{"type": "array", "anyOf": []any{map[string]any{}}, "items": json.RawMessage(sourceBindingRecord099F)}
				})
				anchor := sourceBindingCitation099F(t, facts, a.IntendedBindings[0].SourceSchema.Pointer)
				good := a.IntendedBindings[0]
				good.Lane = "direct_read"
				good.SourceSchema = &anchor
				bad := good
				bad.Kind = "schema"
				bad.ID = "schemas/widgets.json"
				bad.Artifact = "internal/connectors/defs/acme/" + bad.ID
				bad.ArtifactSHA256 = sourceBytesHash(facts.bindings.Artifacts[bad.Artifact])
				bad.Pointer = ""
				bad.CanonicalPointer = "/operations/0/schema_refs/record"
				bad.FieldMappings = append([]sourceLaneFieldMapping{}, good.FieldMappings...)
				bad.FieldMappings[0].Source = sourceBindingCitation099F(t, facts, anchor.Pointer+"/properties/other/items")
				a.ResponseInterpretations = nil
				a.IntendedBindings = []sourceLaneTargetRef{good, bad}
				if reverse {
					a.IntendedBindings = []sourceLaneTargetRef{bad, good}
				}
				if materialized {
					a.MaterializedBindings = a.IntendedBindings
					a.IntendedBindings = nil
				}
				got, rebased := ownership131Build(t, key, facts, a)
				refs := rebased.IntendedBindings
				if materialized {
					refs = rebased.MaterializedBindings
				}
				expectedGood, expectedBad := refs[0], refs[1]
				if reverse {
					expectedGood, expectedBad = refs[1], refs[0]
				}
				cell := requireSourceLane(t, got.SourceOperations[0].Lanes, "direct_read", "applicable")
				if len(cell.References) != 1 || !sourceLaneTargetRefEqual(cell.References[0], canonicalSourceLaneTargetRef(expectedGood)) {
					t.Errorf("exact good sibling lost or bad retained: %+v", cell)
				}
				severity := "deficit"
				if materialized {
					severity = "error"
				}
				ownership132CellDiagnostic(t, key, cell, expectedBad.FieldMappings[0].Source.Pointer, "source_schema_unverified", severity)
				record := sourceLaneProofRecord{ID: "132-independent-sibling-proof", Key: key, Lane: "direct_read", Targets: []sourceLaneTargetRef{canonicalSourceLaneTargetRef(expectedGood)}}
				result := assessSourceLaneProof(key, got.SourceOperations[0].Lanes, sourceLaneProofInputs{byCell: map[sourceLaneProofCell]sourceLaneProofRecord{{key, "direct_read"}: record}})
				assessed := requireSourceLane(t, result, "direct_read", "applicable")
				if assessed.State == "implemented" || len(assessed.ProofRefs) > 0 {
					t.Error("affected lane promoted despite bad sibling")
				}
				baseline := assessSourceLaneProof(key, got.SourceOperations[0].Lanes, sourceLaneProofInputs{})
				for i, c := range baseline {
					if c.Lane != "direct_read" && !reflect.DeepEqual(c, result[i]) {
						t.Errorf("unrelated lane %s changed", c.Lane)
					}
				}
			})
		}
	}
}

func TestSourceLane132SelectedDirectUnknownSibling(t *testing.T) {
	for _, keyword := range []string{"$dynamicRef", "$recursiveRef", "$id", "unknownApplicator"} {
		t.Run(keyword, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				schema := ownership132Response(doc)
				schema["properties"].(map[string]any)["other"] = map[string]any{keyword: "#/Record"}
			})
			anchor := sourceBindingCitation099F(t, facts, a.IntendedBindings[0].SourceSchema.Pointer)
			a.IntendedBindings[0].SourceSchema = &anchor
			a.IntendedBindings[0].Lane = "direct_read"
			a.ResponseInterpretations = nil
			got, rebased := ownership131Build(t, key, facts, a)
			ownership131Check(t, key, got, rebased, false)
		})
	}
}

func TestSourceLane132GraphQLDelegation(t *testing.T) {
	for _, kind := range []string{"write", "command", "schema", "canonical_operation"} {
		for _, mode := range []string{"positive", "composed", "name", "template"} {
			composition := mode == "composed"
			t.Run(kind+"/"+mode, func(t *testing.T) {
				const document = "mutation Change($id: String!) { change(id: $id) { id } }"
				key, facts, a := sourceBindingFixture099F(t, "body", func(lock *vNextSourceLock) {
					schema := `{"type":"object","properties":{"id":{"type":"string"}},"required":["id"],"additionalProperties":false}`
					lock.Schemas["schemas/request.json"] = json.RawMessage(schema)
					lock.Lanes["direct_write"] = "implemented"
					lock.CLI = json.RawMessage(`{"usage":"pm acme <command>","tagline":"Acme commands"}`)
					lock.Operations[0].Commands = []vNextCommandDescriptor{{Command: json.RawMessage(`{"path":"widgets create","summary":"Create widget","intent":"direct_write","availability":"implemented","write":"create_widget","output_policy":"json_redacted","flags":[]}`)}}
					lock.Operations[0].Write = json.RawMessage(`{"name":"create_widget","kind":"create","method":"POST","path":"/graphql","body_type":"graphql","graphql":{"document":"` + document + `","operation_name":"Change","variables":{"id":"{{ record.id }}"}},"record_schema":` + schema + `,"risk":"low"}`)
					if mode == "template" {
						lock.Operations[0].Write = json.RawMessage(strings.ReplaceAll(string(lock.Operations[0].Write), "record.id", "record.other"))
					}
					if mode == "name" {
						lock.Operations[0].Write = json.RawMessage(strings.ReplaceAll(string(lock.Operations[0].Write), `"variables":{"id":`, `"variables":{"other":`))
					}
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
				switch kind {
				case "command":
					ref.Kind = kind
					ref.ID = "widgets create"
					ref.Artifact = "internal/connectors/defs/acme/cli_surface.json"
					ref.Pointer = "/commands/0"
					ref.ArtifactSHA256 = sourceBytesHash(facts.bindings.Artifacts[ref.Artifact])
					ref.CanonicalPointer = "/operations/0/commands/0"
				case "schema":
					ref.Kind = kind
					ref.ID = "schemas/request.json"
					ref.Artifact = "internal/connectors/defs/acme/" + ref.ID
					ref.Pointer = ""
					ref.ArtifactSHA256 = sourceBytesHash(facts.bindings.Artifacts[ref.Artifact])
					ref.CanonicalPointer = "/operations/0/schema_refs/request"
				case "canonical_operation":
					ref.Kind = kind
					ref.ID = ref.CanonicalID
					ref.Artifact = "internal/connectors/defs/acme/source.lock.json"
					ref.ArtifactSHA256 = sourceBytesHash(facts.bindings.Authoring[ref.Artifact])
					ref.Pointer = "/operations/0"
					ref.CanonicalPointer = "/operations/0"
				}
				got, rebased := ownership131Build(t, key, facts, a)
				ownership131Check(t, key, got, rebased, mode != "positive")
				if mode == "name" || mode == "template" {
					cell := requireSourceLane(t, got.SourceOperations[0].Lanes, "direct_write", "applicable")
					diagnosticPointer := rebased.IntendedBindings[0].Artifact + "#" + rebased.IntendedBindings[0].Pointer
					if kind == "schema" || kind == "canonical_operation" {
						diagnosticPointer = "internal/connectors/defs/acme/writes.json#/actions/0"
					}
					ownership132CellDiagnostic(t, key, cell, diagnosticPointer, "target_graphql_variable_mismatch", "error")
				}
				if composition {
					ownership132RequireDiagnostic(t, key, got, rebased, "source_schema_unverified", "deficit")
				}
			})
		}
	}
}

// Post-initial-GREEN controls of the new state carrier and shared cached lookup;
// these are later-edge evidence, never the original131 or augmented matrix RED.
func TestSourceLane132CheckedStateBoundaries(t *testing.T) {
	for _, raw := range []string{
		`{"type":"object","required":[null],"properties":{"":{"type":"string"}}}`,
		`{"type":"object","required":["id","id"],"properties":{"id":{"type":"string"}}}`,
		`{"type":"object","additionalProperties": null ,"properties":{"id":{"type":"string"}}}`,
		`{"items":{"type":"string"}}`,
	} {
		t.Run(raw, func(t *testing.T) {
			pointer := "/properties/id"
			if strings.Contains(raw, `[null]`) {
				pointer = "/properties/"
			}
			if strings.HasPrefix(raw, `{"items"`) {
				pointer = "/items"
			}
			_, code := sourceLaneSchemaProjection(sourceFacts{Document: json.RawMessage(raw)}, sourceFactRef{Pointer: ""}, pointer)
			if code != "source_schema_unverified" {
				t.Errorf("unsupported ancestor code=%s", code)
			}
		})
	}
	for _, cached := range []bool{false, true} {
		for _, pointer := range []string{"", "/a~1b/0", "/a~1b/00", "/R~02", "/R~2"} {
			t.Run(fmt.Sprintf("cached=%t%s", cached, pointer), func(t *testing.T) {
				raw := json.RawMessage(`{"a/b":[{"type":"string"}],"R~2":{"type":"integer"}}`)
				facts := sourceFacts{Document: raw}
				if cached {
					facts.analysis = &sourceShapeAnalysis{}
				}
				value, err := sourceAnalysisPointer(facts, pointer)
				want := pointer != "/a~1b/00" && pointer != "/R~2"
				if (err == nil) != want {
					t.Errorf("pointer=%s accepted=%t want=%t", pointer, err == nil, want)
				}
				if want {
					expected, e := sourceJSONPointer(raw, pointer)
					if e != nil {
						t.Fatal(e)
					}
					a, _ := canonicalSourceJSON(value)
					b, _ := canonicalSourceJSON(expected)
					if string(a) != string(b) {
						t.Error("cached/ref raw lookup selected different bytes")
					}
				}
			})
		}
	}
	for _, count := range []int{127, 129} {
		t.Run(fmt.Sprintf("selected-root-chain/%d", count), func(t *testing.T) {
			defs := map[string]any{}
			for i := 0; i < count; i++ {
				defs[fmt.Sprint(i)] = map[string]any{"$ref": fmt.Sprintf("#/defs/%d", i+1)}
			}
			defs[fmt.Sprint(count)] = map[string]any{"type": "string"}
			raw, _ := json.Marshal(map[string]any{"defs": defs, "schema": map[string]any{"$ref": "#/defs/0"}})
			facts := sourceFacts{Document: raw}
			first, code := sourceLaneSchemaProjection(facts, sourceFactRef{Pointer: "/schema"}, "/schema")
			again, c := sourceLaneSchemaProjection(facts, sourceFactRef{Pointer: "/schema"}, "/schema")
			if c != code || !reflect.DeepEqual(first, again) {
				t.Fatal("identical bounded input changed lineage result")
			}
			want := ""
			if count == 129 {
				want = "source_schema_unverified"
			}
			if code != want {
				t.Errorf("selected root depth code=%s want=%s", code, want)
			}
		})
	}
}
