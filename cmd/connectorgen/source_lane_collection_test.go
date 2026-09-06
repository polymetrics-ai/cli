package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSourceLane122RetainedCollectionBoundary(t *testing.T) {
	for _, tc := range []struct{ name, schema, want string }{
		{"unfamiliar envelope", `{"type":"object","properties":{"templates":{"type":"array","items":{"type":"object"}}}}`, "undetermined"},
		{"familiar metadata envelope", `{"type":"object","properties":{"data":{"type":"array","items":{"type":"object"}}}}`, "undetermined"},
		{"bare object", `{"type":"object"}`, "undetermined"},
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
