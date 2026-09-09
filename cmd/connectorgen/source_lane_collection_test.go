package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSourceLane122RetainedCollectionBoundary(t *testing.T) {
	for _, tc := range []struct{ name, schema, want string }{
		{"unfamiliar envelope", `{"type":"object","properties":{"templates":{"type":"array","items":{"type":"object"}}}}`, "undetermined"},
		{"familiar metadata envelope", `{"type":"object","properties":{"data":{"type":"array","items":{"type":"object"}}}}`, "undetermined"},
		{"bare object", `{"type":"object"}`, "undetermined"},
		{"malformed closed properties", `{"type":"object","additionalProperties":false,"properties":[]}`, "undetermined"},
		{"null closed properties", `{"type":"object","additionalProperties":false,"properties":null}`, "undetermined"},
		{"agreeing array alternatives", `{"oneOf":[{"type":"array","items":{"type":"object"}},{"type":"array","items":{"type":"object","properties":{"id":{"type":"string"}}}}]}`, "applicable"},
		{"agreeing scalar alternatives", `{"anyOf":[{"type":"string"},{"type":"number"}]}`, "not_applicable"},
		{"closed nested metadata", `{"type":"object","additionalProperties":false,"properties":{"meta":{"type":"object","additionalProperties":false,"properties":{"count":{"type":"integer"}}}}}`, "not_applicable"},
		{"closed scalar object", `{"type":"object","additionalProperties":false,"properties":{"id":{"type":"string"}}}`, "not_applicable"},
		{"root records unknown field", `{"type":"array","items":{"type":"object","properties":{"detail":{"$ref":"https://example.invalid/schema"}}}}`, "applicable"},
		{"mixed alternatives", `{"oneOf":[{"type":"array","items":{"type":"object"}},{"type":"string"}]}`, "undetermined"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
			var schema any
			if err := json.Unmarshal([]byte(tc.schema), &schema); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(filepath.Join(root, "source.json"))
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err = json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			for _, r := range doc["rest"].(map[string]any)["operations"].([]any) {
				r.(map[string]any)["source_operation"] = map[string]any{"summary": "Read response", "responses": map[string]any{"200": map[string]any{"content": map[string]any{"application/json": map[string]any{"schema": schema}}}}}
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
				t.Fatal("builder lost source rows/cells")
			}
			for _, r := range got.SourceOperations {
				if !r.Source.Observed || r.Facts.Status == "unavailable" {
					t.Fatal("fixture did not reach source classification")
				}
				requireSourceLane(t, r.Lanes, "direct_read", "applicable")
				c := requireSourceLane(t, r.Lanes, "etl", tc.want)
				if c.State == "implemented" || len(c.References) != 0 || len(c.ProofRefs) != 0 {
					t.Fatal("cardinality created executable evidence")
				}
			}
		})
	}
}

func TestSourceLane122ReviewedCollectionSeeds(t *testing.T) {
	for _, tc := range []struct{ connector, id, schemaPointer, schemaHash, records, support, clause string }{
		{"notion", "notion.rest.list-data-source-templates", "/rest/operations/20/source_operation/responses/200/content/application~1json/schema", "d0bb6e883cbb6b3319c80548204e679444c5760ba168d367b5f96cc5b32a9e6b", "/templates", "summary", "List templates in a data source"},
		{"jira", "jira.rest.getForgeAppPropertyKeys", "/rest/operations/612/source_operation/responses/200/content/application~1json/schema", "bc67b792846dc2f7af0c72295878bcf3d11b1666b3169afe83cb921aafa878e5", "/keys", "description", "Returns all property keys for the Forge app."},
		{"circleci", "circleci.rest.getOrgSummaryData", "/rest/operations/2/source_operation/responses/200/content/application~1json/schema", "bebfac3af76e397550408de4732d94ad8e4ca0c2a1fd05f4cc154c0782c24001", "/org_project_data", "summary", "Get summary metrics with trends for the entire org, and for each project."},
	} {
		t.Run(tc.connector, func(t *testing.T) {
			row, doc := retainedFactFixture(t, tc.connector, tc.id)
			facts := normalizeSourceFacts(row, doc, nil)
			var archived struct {
				Rest struct {
					Operations []struct {
						ID string `json:"id"`
					} `json:"operations"`
				} `json:"rest"`
			}
			if err := json.Unmarshal(doc.Payload, &archived); err != nil {
				t.Fatal(err)
			}
			ids := []string{}
			for _, r := range archived.Rest.Operations {
				ids = append(ids, r.ID)
			}
			root := t.TempDir()
			proofWrite(t, root, doc.Path, doc.Payload)
			cohort := sourceLaneCohort{SchemaVersion: 1, CohortID: "fixture", Inventories: []sourceInventoryAnchor{{Connector: tc.connector, Inventory: "primary", Class: "primary", Path: doc.Path, SHA256: doc.RetainedFileSHA256, ExpectedIDs: ids, ExpectedCount: len(ids)}}}
			for _, present := range []bool{false, true} {
				var annotations []sourceSemanticAnnotation
				if present {
					raw, err := json.Marshal(map[string]any{"key": row.Key, "citation": facts.Refs["summary"], "clause": sourceFactText(facts, "summary"), "response_interpretations": []any{map[string]any{"kind": "collection", "response_schema": sourceFactRef{DocumentID: doc.ID, Pointer: tc.schemaPointer, ValueSHA256: tc.schemaHash}, "records_pointer": tc.records, "citation": facts.Refs[tc.support], "clause": tc.clause}}})
					if err != nil {
						t.Fatal(err)
					}
					// This exercises the existing typed builder input. Before the optional
					// member exists it is ignored; the reached classification must still fail
					// the independently expected source collection, rather than a compiler.
					var a sourceSemanticAnnotation
					if err = json.Unmarshal(raw, &a); err != nil {
						t.Fatal(err)
					}
					annotations = []sourceSemanticAnnotation{a}
				}
				got, err := buildSourceLaneManifest(context.Background(), root, cohort, annotations)
				if err != nil {
					t.Fatal(err)
				}
				if len(got.SourceOperations) != len(ids) || got.SourceTotals.Cells != 7*len(ids) {
					t.Fatal("source universe changed")
				}
				found := false
				for _, r := range got.SourceOperations {
					if r.Source.Key == row.Key {
						found = true
						if !r.Source.Observed || r.Facts.Status == "unavailable" {
							t.Fatal("retained seed did not reach classifier")
						}
						want := "undetermined"
						if present {
							want = "applicable"
						}
						for _, c := range r.Lanes {
							if c.Lane == "etl" {
								if !present && !proofHasDiagnostic(c.Diagnostics, "source_collection_interpretation_missing", "deficit") {
									t.Errorf("%s missing explicit source interpretation diagnostic", tc.id)
								}
								if c.Applicability != want {
									t.Errorf("%s interpretation=%v ETL=%s want %s", tc.id, present, c.Applicability, want)
								}
								if c.State == "implemented" || len(c.References) != 0 || len(c.ProofRefs) != 0 {
									t.Fatal("source interpretation fabricated execution")
								}
							}
						}
					}
				}
				if !found {
					t.Fatal("exact reviewed source key lost")
				}
			}
		})
	}
}

func TestSourceLane122InterpretationContracts(t *testing.T) {
	for _, which := range []string{"collection", "single resource", "unknown kind", "missing pointer", "null pointer", "single with pointer", "duplicate", "conflict", "wrong hash", "wrong document", "wrong scope", "wrong instance path", "schema as instance", "scalar selected", "collection mutation", "root array single"} {
		t.Run(which, func(t *testing.T) {
			schema := map[string]any{"type": "object", "properties": map[string]any{"data": map[string]any{"type": "array", "items": map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}}}, "count": map[string]any{"type": "integer"}}}
			method, summary := "GET", "List widget records"
			if which == "single resource" || which == "single with pointer" {
				summary = "Read one resource with its history"
			}
			if which == "collection mutation" {
				method, summary = "POST", "Create widget records"
			}
			if which == "root array single" {
				schema = map[string]any{"type": "array", "items": map[string]any{"type": "object"}}
			}
			rows := []any{}
			for _, id := range []string{"source.a", "source.b"} {
				rows = append(rows, map[string]any{"id": id, "protocol": "rest", "method": method, "path": "/widgets", "source_operation": map[string]any{"summary": summary, "responses": map[string]any{"200": map[string]any{"description": summary, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}}}})
			}
			raw, err := json.Marshal(map[string]any{"schema_version": 2, "connector": "fixture", "counts": map[string]any{"total": 2}, "rest": map[string]any{"operations": rows}})
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			proofWrite(t, root, "source.json", raw)
			cohort := sourceLaneCohort{SchemaVersion: 1, CohortID: "fixture", Inventories: []sourceInventoryAnchor{{Connector: "fixture", Inventory: "primary", Class: "primary", Path: "source.json", SHA256: sourceBytesHash(raw), ExpectedIDs: []string{"source.a", "source.b"}, ExpectedCount: 2}}}
			inventory := loadRetainedSourceInventory(context.Background(), root, cohort)
			if len(inventory.Documents) != 1 || len(inventory.Operations) != 2 || !inventory.Operations[0].Observed {
				t.Fatal("fixture did not pass retained loader")
			}
			row := inventory.Operations[0]
			facts := normalizeSourceFacts(row, inventory.Documents[0], nil)
			schemaRaw, _ := json.Marshal(schema)
			ref := sourceFactRef{DocumentID: "fixture:primary", Pointer: "/rest/operations/0/source_operation/responses/200/content/application~1json/schema", ValueSHA256: sourceBytesHash(schemaRaw)}
			interpretation := map[string]any{"kind": "collection", "response_schema": ref, "records_pointer": "/data", "citation": facts.Refs["summary"], "clause": summary}
			switch which {
			case "single resource", "root array single":
				interpretation["kind"] = "single_resource"
				delete(interpretation, "records_pointer")
			case "single with pointer":
				interpretation["kind"] = "single_resource"
			case "unknown kind":
				interpretation["kind"] = "anything"
			case "missing pointer":
				delete(interpretation, "records_pointer")
			case "null pointer":
				interpretation["records_pointer"] = nil
			case "wrong hash":
				ref.ValueSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
				interpretation["response_schema"] = ref
			case "wrong document":
				ref.DocumentID = "other:primary"
				interpretation["response_schema"] = ref
			case "wrong scope":
				ref.Pointer = "/rest/operations/1/source_operation/responses/200/content/application~1json/schema"
				interpretation["response_schema"] = ref
			case "wrong instance path":
				interpretation["records_pointer"] = "/other"
			case "schema as instance":
				interpretation["records_pointer"] = "/properties/data/items"
			case "scalar selected":
				interpretation["records_pointer"] = "/count"
			}
			interpretations := []any{interpretation}
			if which == "duplicate" {
				interpretations = append(interpretations, interpretation)
			}
			if which == "conflict" {
				interpretations = append(interpretations, map[string]any{"kind": "single_resource", "response_schema": ref, "citation": facts.Refs["summary"], "clause": summary})
			}
			encoded, err := json.Marshal(map[string]any{"key": row.Key, "citation": facts.Refs["summary"], "clause": summary, "response_interpretations": interpretations})
			if err != nil {
				t.Fatal(err)
			}
			var annotation sourceSemanticAnnotation
			if err = json.Unmarshal(encoded, &annotation); err != nil {
				t.Fatal(err)
			}
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{annotation})
			if err != nil {
				t.Fatal(err)
			}
			if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 {
				t.Fatal("interpretation discarded source membership")
			}
			valid := which == "collection" || which == "single resource"
			if (got.Validation.Status == "valid") != valid {
				t.Errorf("%s validation=%s want valid=%v", which, got.Validation.Status, valid)
			}
			for _, r := range got.SourceOperations {
				if r.Source.Key == row.Key {
					if valid {
						want := "applicable"
						if which == "single resource" {
							want = "not_applicable"
						}
						requireSourceLane(t, r.Lanes, "etl", want)
					}
					expectedRead := "applicable"
					if method == "POST" {
						expectedRead = "not_applicable"
					}
					requireSourceLane(t, r.Lanes, "direct_read", expectedRead)
					for _, c := range r.Lanes {
						if c.State == "implemented" || len(c.ProofRefs) != 0 {
							t.Fatal("interpretation fabricated behavior proof")
						}
					}
				}
			}
		})
	}
}

func sourceCollection122Fixture(t *testing.T, schemaJSON, contractJSON string) (string, sourceLaneCohort, sourceOperationKey, sourceFacts, sourceFactRef) {
	t.Helper()
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	raw, err := os.ReadFile(filepath.Join(root, "source.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	var schema any
	if err = json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		t.Fatal(err)
	}
	for _, row := range doc["rest"].(map[string]any)["operations"].([]any) {
		row.(map[string]any)["source_operation"] = map[string]any{"summary": "List primary and secondary records", "responses": map[string]any{"200": map[string]any{"content": map[string]any{"application/json": map[string]any{"schema": schema}}}}}
	}
	if contractJSON != "" {
		var contract any
		if err = json.Unmarshal([]byte(contractJSON), &contract); err != nil {
			t.Fatal(err)
		}
		doc["source_contract"] = contract
	}
	raw, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, root, "source.json", raw)
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	inventory := loadRetainedSourceInventory(context.Background(), root, cohort)
	if len(inventory.Documents) != 1 || len(inventory.Operations) != 2 || !inventory.Operations[0].Observed {
		t.Fatal("lineage fixture did not pass source loading")
	}
	row := inventory.Operations[0]
	facts := normalizeSourceFacts(row, inventory.Documents[0], nil)
	schemaRaw, _ := json.Marshal(schema)
	ref := sourceFactRef{DocumentID: "fixture:primary", Pointer: "/rest/operations/0/source_operation/responses/200/content/application~1json/schema", ValueSHA256: sourceBytesHash(schemaRaw)}
	return root, cohort, row.Key, facts, ref
}

func TestSourceLane122InterpretationLineage(t *testing.T) {
	for _, tc := range []struct {
		name, schema, contract, path, kind, want string
		valid                                    bool
	}{
		{"deeper", `{"type":"object","properties":{"outer":{"type":"object","properties":{"records":{"type":"array","items":{"type":"object"}}}}}}`, "", "/outer/records", "collection", "applicable", true},
		{"escaped", `{"type":"object","properties":{"a/b~c":{"type":"array","items":{"type":"object"}}}}`, "", "/a~1b~0c", "collection", "applicable", true},
		{"local schema", `{"$ref":"#/components/schemas/Envelope"}`, `{"components":{"schemas":{"Envelope":{"type":"object","properties":{"data":{"type":"array","items":{"type":"object"}}}}}}}`, "/data", "collection", "applicable", true},
		{"external", `{"$ref":"https://example.invalid/schema"}`, "", "/data", "collection", "undetermined", true},
		{"cycle", `{"$ref":"#/components/schemas/Cycle"}`, `{"components":{"schemas":{"Cycle":{"$ref":"#/components/schemas/Cycle"}}}}`, "/data", "collection", "undetermined", true},
		{"scalar wrapper", `{"type":"string","properties":{"data":{"type":"array","items":{"type":"object"}}}}`, "", "/data", "collection", "undetermined", false},
		{"single unknown conjunction", `{"type":"object","allOf":[{"$ref":"https://example.invalid/schema"}]}`, "", "", "single_resource", "undetermined", true},
		{"unrelated support", `{"type":"object","properties":{"data":{"type":"array","items":{"type":"object"}},"count":{"type":"integer","description":"Count of metadata values"}}}`, "", "/data", "collection", "undetermined", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, cohort, key, facts, ref := sourceCollection122Fixture(t, tc.schema, tc.contract)
			interpretation := sourceResponseInterpretation{Kind: tc.kind, ResponseSchema: ref, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary")}
			if tc.kind == "collection" {
				path := tc.path
				interpretation.RecordsPointer = &path
			}
			if tc.name == "unrelated support" {
				pointer := ref.Pointer + "/properties/count/description"
				value := json.RawMessage(`"Count of metadata values"`)
				interpretation.Citation = sourceFactRef{DocumentID: ref.DocumentID, Pointer: pointer, ValueSHA256: sourceBytesHash(value)}
				interpretation.Clause = "Count of metadata values"
			}
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary"), ResponseInterpretations: []sourceResponseInterpretation{interpretation}}
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
			if err != nil {
				t.Fatal(err)
			}
			if (got.Validation.Status == "valid") != tc.valid {
				t.Errorf("lineage validation=%s want valid=%v", got.Validation.Status, tc.valid)
			}
			for _, r := range got.SourceOperations {
				if r.Source.Key == key {
					requireSourceLane(t, r.Lanes, "etl", tc.want)
					requireSourceLane(t, r.Lanes, "direct_read", "applicable")
				}
			}
		})
	}
}

func TestSourceLane122InterpretationOrder(t *testing.T) {
	root, cohort, key, facts, ref := sourceCollection122Fixture(t, `{"type":"object","properties":{"data":{"type":"array","items":{"type":"object"}},"other":{"type":"array","items":{"type":"object"}}}}`, "")
	a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary")}
	for _, path := range []string{"/data", "/other"} {
		p := path
		a.ResponseInterpretations = append(a.ResponseInterpretations, sourceResponseInterpretation{Kind: "collection", ResponseSchema: ref, RecordsPointer: &p, Citation: a.Citation, Clause: a.Clause})
	}
	first, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
	if err != nil {
		t.Fatal(err)
	}
	a.ResponseInterpretations[0], a.ResponseInterpretations[1] = a.ResponseInterpretations[1], a.ResponseInterpretations[0]
	second, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
	if err != nil {
		t.Fatal(err)
	}
	left, _ := json.Marshal(first)
	right, _ := json.Marshal(second)
	if string(left) != string(right) {
		t.Fatal("interpretation permutation changes complete manifest bytes")
	}
}

func TestSourceLane122IndependentInterpretationEvidence(t *testing.T) {
	for _, which := range []string{"wrong literal hash", "omitted required citation", "unbacked interpreted claim"} {
		t.Run(which, func(t *testing.T) {
			schema := `{"type":"object","properties":{"data":{"type":"array","items":{"type":"object"}}}}`
			if which == "unbacked interpreted claim" {
				schema = `{"type":"array","items":{"type":"object"}}`
			}
			root, cohort, key, facts, ref := sourceCollection122Fixture(t, schema, "")
			path := "/data"
			a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary"), ResponseInterpretations: []sourceResponseInterpretation{{Kind: "collection", ResponseSchema: ref, RecordsPointer: &path, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary")}}}
			annotations := []sourceSemanticAnnotation{a}
			if which == "unbacked interpreted claim" {
				annotations = nil
			}
			expected, err := buildSourceLaneManifest(context.Background(), root, cohort, annotations)
			if err != nil {
				t.Fatal(err)
			}
			if expected.Validation.Status != "valid" {
				t.Fatal("independent evidence fixture was not valid")
			}
			for i := range expected.SourceOperations {
				if expected.SourceOperations[i].Source.Key == key {
					for j := range expected.SourceOperations[i].Lanes {
						c := &expected.SourceOperations[i].Lanes[j]
						if c.Lane != "etl" {
							continue
						}
						if c.Applicability != "applicable" {
							t.Fatal("fixture did not reach collection")
						}
						switch which {
						case "wrong literal hash":
							for k := range c.FactRefs {
								if c.FactRefs[k].Pointer == ref.Pointer {
									c.FactRefs[k].ValueSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
								}
							}
						case "omitted required citation":
							c.FactRefs = nil
						case "unbacked interpreted claim":
							c.RuleID = "source_response_interpretation"
						}
					}
				}
			}
			// Both producer-shaped values agree on the same wrong claim. The validator
			// must use retained documents and original authoring input, not that agreement.
			raw, _ := json.Marshal(expected)
			var candidate sourceLaneManifest
			if err = json.Unmarshal(raw, &candidate); err != nil {
				t.Fatal(err)
			}
			found := false
			for _, d := range validateSourceLaneManifest(candidate, expected) {
				if d.Code == "source_collection_citation_invalid" || d.Code == "source_collection_citation_missing" || d.Code == "source_collection_authority_missing" {
					found = true
				}
			}
			if !found {
				t.Fatal("independent validator trusted matching incorrect produced interpretation evidence")
			}
		})
	}
}

func TestSourceLane122ResponseOccurrenceCustody(t *testing.T) {
	for _, mode := range []string{"local response", "request clone", "error clone", "component clone", "foreign operation", "unknown success sibling", "unknown media sibling"} {
		t.Run(mode, func(t *testing.T) {
			root, cohort, _, _, _ := sourceCollection122Fixture(t, `{"type":"object","properties":{"records":{"type":"array","items":{"type":"object"}}}}`, "")
			raw, err := os.ReadFile(filepath.Join(root, "source.json"))
			if err != nil {
				t.Fatal(err)
			}
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
			responses := op["responses"].(map[string]any)
			response := responses["200"].(map[string]any)
			schema := response["content"].(map[string]any)["application/json"].(map[string]any)["schema"]
			pointer := "/rest/operations/0/source_operation/responses/200/content/application~1json/schema"
			unknownPointer := ""
			switch mode {
			case "local response":
				doc["source_contract"] = map[string]any{"components": map[string]any{"responses": map[string]any{"Page": response}}}
				responses["200"] = map[string]any{"$ref": "#/components/responses/Page"}
				pointer = "/source_contract/components/responses/Page/content/application~1json/schema"
			case "request clone":
				op["requestBody"] = response
				pointer = "/rest/operations/0/source_operation/requestBody/content/application~1json/schema"
			case "error clone":
				responses["400"] = response
				pointer = "/rest/operations/0/source_operation/responses/400/content/application~1json/schema"
			case "component clone":
				doc["source_contract"] = map[string]any{"components": map[string]any{"schemas": map[string]any{"Clone": schema}}}
				pointer = "/source_contract/components/schemas/Clone"
			case "foreign operation":
				pointer = "/rest/operations/1/source_operation/responses/200/content/application~1json/schema"
			case "unknown success sibling":
				responses["201"] = map[string]any{"$ref": "https://example.invalid/response"}
				unknownPointer = "/rest/operations/0/source_operation/responses/201"
			case "unknown media sibling":
				response["content"].(map[string]any)["application/xml"] = map[string]any{"schema": map[string]any{"$ref": "https://example.invalid/schema"}}
				unknownPointer = "/rest/operations/0/source_operation/responses/200/content/application~1xml/schema"
			}
			raw, err = json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			proofWrite(t, root, "source.json", raw)
			cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
			inventory := loadRetainedSourceInventory(context.Background(), root, cohort)
			if len(inventory.Operations) != 2 || !inventory.Operations[0].Observed {
				t.Fatal("source occurrence fixture not reached")
			}
			row := inventory.Operations[0]
			facts := normalizeSourceFacts(row, inventory.Documents[0], nil)
			literal, err := json.Marshal(schema)
			if err != nil {
				t.Fatal(err)
			}
			path := "/records"
			a := sourceSemanticAnnotation{Key: row.Key, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary"), ResponseInterpretations: []sourceResponseInterpretation{{Kind: "collection", ResponseSchema: sourceFactRef{DocumentID: "fixture:primary", Pointer: pointer, ValueSHA256: sourceBytesHash(literal)}, RecordsPointer: &path, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary")}}}
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
			if err != nil {
				t.Fatal(err)
			}
			valid := mode == "local response" || unknownPointer != ""
			if (got.Validation.Status == "valid") != valid {
				t.Fatalf("validation=%s want valid=%v", got.Validation.Status, valid)
			}
			if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 {
				t.Fatal("source membership changed")
			}
			for _, result := range got.SourceOperations {
				if result.Source.Key != row.Key {
					continue
				}
				want := "undetermined"
				if valid {
					want = "applicable"
				}
				cell := requireSourceLane(t, result.Lanes, "etl", want)
				requireSourceLane(t, result.Lanes, "direct_read", "applicable")
				if valid {
					found := false
					for _, ref := range cell.FactRefs {
						found = found || ref == a.ResponseInterpretations[0].ResponseSchema
					}
					if !found {
						t.Fatal("lost literal successful response occurrence")
					}
				}
				if unknownPointer != "" {
					found := false
					for _, d := range cell.Diagnostics {
						found = found || (d.Code == "source_collection_scope_unknown" && d.Pointer == unknownPointer && d.Severity == "deficit")
					}
					if !found {
						t.Fatalf("lost unknown sibling at %s: %+v", unknownPointer, cell.Diagnostics)
					}
				}
				if cell.State == "implemented" || len(cell.ProofRefs) != 0 || len(cell.References) != 0 {
					t.Fatal("interpretation promoted executable proof")
				}
			}
		})
	}
}

func TestSourceLane122CLIInterpretationReader(t *testing.T) {
	root, cohort, key, facts, ref := sourceCollection122Fixture(t, `{"type":"object","properties":{"records":{"type":"array","items":{"type":"object"}}}}`, "")
	path := "/records"
	a := sourceSemanticAnnotation{Key: key, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary"), ResponseInterpretations: []sourceResponseInterpretation{{Kind: "collection", ResponseSchema: ref, RecordsPointer: &path, Citation: facts.Refs["summary"], Clause: sourceFactText(facts, "summary")}}}
	cohortRaw, err := json.Marshal(cohort)
	if err != nil {
		t.Fatal(err)
	}
	proofWrite(t, root, "data/connector-canon/batch1-source-lane-cohort.json", cohortRaw)
	proofDocument(t, root, []sourceLaneProofRecord{})
	for _, invalid := range []bool{false, true} {
		raw, err := json.Marshal(map[string]any{"schema_version": 1, "annotations": []sourceSemanticAnnotation{a}})
		if err != nil {
			t.Fatal(err)
		}
		if invalid {
			raw = bytes.Replace(raw, []byte(`"records_pointer":"/records"`), []byte(`"records_pointer":null`), 1)
		}
		proofWrite(t, root, "data/connector-canon/batch1-source-lane-annotations.json", raw)
		before := sourceBindingFixtureSnapshot(t, root)
		var out, diagnostics bytes.Buffer
		code := runContext(context.Background(), []string{"source-lanes", "--repo", root}, &out, &diagnostics)
		wantCode := 0
		if invalid {
			wantCode = 1
		}
		if code != wantCode {
			t.Fatalf("reader code%d want%d: %s", code, wantCode, &diagnostics)
		}
		var manifest sourceLaneManifest
		if err := json.Unmarshal(out.Bytes(), &manifest); err != nil {
			t.Fatal(err)
		}
		if len(manifest.SourceOperations) != 2 || manifest.SourceTotals.Cells != 14 {
			t.Fatal("reader lost source rows")
		}
		for _, row := range manifest.SourceOperations {
			if row.Source.Key != key {
				continue
			}
			want := "applicable"
			if invalid {
				want = "undetermined"
			}
			cell := requireSourceLane(t, row.Lanes, "etl", want)
			if invalid && !proofHasDiagnostic(cell.Diagnostics, "source_collection_interpretation_invalid", "error") {
				t.Fatal("reader lost invalid interpretation diagnosis")
			}
			requireSourceLane(t, row.Lanes, "direct_read", "applicable")
		}
		if !reflect.DeepEqual(before, sourceBindingFixtureSnapshot(t, root)) {
			t.Fatal("read-only CLI changed fixture")
		}
	}
}

func TestSourceLane122ReferenceConjunctionRefusal(t *testing.T) {
	for _, tc := range []struct{ name, schema, want string }{
		{"annotation sibling", `{"$ref":"#/components/schemas/Records","description":"Retained records"}`, "applicable"},
		{"conflicting scalar sibling", `{"$ref":"#/components/schemas/Records","type":"string"}`, "undetermined"},
		{"unproved constraint sibling", `{"$ref":"#/components/schemas/Records","not":{"type":"array"}}`, "undetermined"},
		{"conflicting item sibling", `{"type":"array","items":{"$ref":"#/components/schemas/Record","type":"string"}}`, "undetermined"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, cohort, key, _, _ := sourceCollection122Fixture(t, tc.schema, `{"components":{"schemas":{"Records":{"type":"array","items":{"type":"object"}},"Record":{"type":"object"}}}}`)
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range got.SourceOperations {
				if row.Source.Key == key {
					requireSourceLane(t, row.Lanes, "etl", tc.want)
					requireSourceLane(t, row.Lanes, "direct_read", "applicable")
				}
			}
		})
	}
}
