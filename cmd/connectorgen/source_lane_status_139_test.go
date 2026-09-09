package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceLane139RetainedResponseStatus(t *testing.T) {
	statuses := []struct {
		key     string
		success bool
	}{
		{"200", true}, {"299", true}, {"2XX", true},
		{"2invalid", false}, {"20", false}, {"2000", false}, {"2ab", false},
		{"2", false}, {"2X0", false}, {"2xx", false}, {"20/0", false},
		{"default", false}, {"400", false}, {"4XX", false},
	}
	for _, status := range statuses {
		for _, shape := range []struct{ name, media, schema, etl, binary string }{
			{"collection", "application/json", `{"type":"array","items":{"type":"object"}}`, "applicable", "not_applicable"},
			{"scalar", "application/json", `{"type":"string"}`, "not_applicable", "not_applicable"},
			{"binary", "application/octet-stream", `{"type":"string","format":"binary"}`, "not_applicable", "applicable"},
		} {
			t.Run(status.key+"/"+shape.name, func(t *testing.T) {
				root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
				raw, err := os.ReadFile(filepath.Join(root, "source.json"))
				if err != nil {
					t.Fatal(err)
				}
				var doc map[string]any
				if err := json.Unmarshal(raw, &doc); err != nil {
					t.Fatal(err)
				}
				for _, item := range doc["rest"].(map[string]any)["operations"].([]any) {
					item.(map[string]any)["source_operation"] = map[string]any{
						"summary":   "Read response",
						"responses": map[string]any{status.key: map[string]any{"content": map[string]any{shape.media: map[string]any{"schema": json.RawMessage(shape.schema)}}}},
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
					t.Fatal("lost source universe")
				}
				for i, row := range got.SourceOperations {
					wantKey := sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: []string{"source.a", "source.b"}[i]}
					if row.Source.Key != wantKey || !row.Source.Observed || len(row.Lanes) != 7 {
						t.Fatal("lost exact observed key or lanes")
					}
					wantETL, wantBinary := "undetermined", "undetermined"
					if status.success {
						wantETL, wantBinary = shape.etl, shape.binary
					}
					requireSourceLane(t, row.Lanes, "direct_read", "applicable")
					requireSourceLane(t, row.Lanes, "etl", wantETL)
					requireSourceLane(t, row.Lanes, "binary_download", wantBinary)
					for _, cell := range row.Lanes {
						if cell.State == "implemented" || len(cell.References) != 0 || len(cell.ProofRefs) != 0 {
							t.Fatal("source shape invented executable proof")
						}
					}
				}
				if findings := validateSourceLaneManifest(got, got); len(findings) != 0 {
					t.Fatalf("valid emitted manifest rejected: %+v", findings)
				}
				if !status.success {
					// The validator must reject a forged positive even if the caller
					// supplies the same forged object as its expected manifest.
					for i := range got.SourceOperations {
						for j := range got.SourceOperations[i].Lanes {
							c := &got.SourceOperations[i].Lanes[j]
							if c.Lane == "etl" {
								c.Applicability = "applicable"
								c.RuleID = "source_response_interpretation"
							}
						}
					}
					if findings := validateSourceLaneManifest(got, got); len(findings) == 0 {
						t.Fatal("forged collection authority accepted")
					}
				}
			})
		}
	}
}

func TestSourceLane139StatusSiblingCoverage(t *testing.T) {
	for _, tc := range []struct {
		status  string
		unknown bool
	}{
		{"2invalid", true}, {"20", true}, {"2000", true}, {"2ab", true}, {"20/0", true},
		{"299", true}, {"2XX", true}, {"400", false}, {"4XX", false}, {"default", false},
	} {
		t.Run(tc.status, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "envelope", nil)
			before := sourceBindingOutcome099F(t, key, facts, a, 1)
			if len(before.Diagnostics) != 0 {
				t.Fatal("complete admitted positive required")
			}
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				op["responses"].(map[string]any)[tc.status] = map[string]any{"$ref": "https://provider.invalid/unresolved"}
			})
			cell := sourceBindingOutcome099F(t, key, facts, a, 1)
			if len(cell.References) != 1 || !sourceLaneTargetRefEqual(cell.References[0], before.References[0]) {
				t.Fatal("valid local200 binding lost")
			}
			pointer := "/rest/operations/0/source_operation/responses/" + escapeSourcePointer(tc.status)
			found := false
			for _, d := range cell.Diagnostics {
				if d.Code == "target_required_scope_unverified" && d.Pointer == pointer && d.Key == key && d.Severity == "deficit" {
					found = true
				}
			}
			if found != tc.unknown {
				t.Errorf("coverage uncertainty=%v want%v: %+v", found, tc.unknown, cell.Diagnostics)
			}
			if len(cell.ProofRefs) != 0 || cell.State == "implemented" {
				t.Fatal("unproven binding promoted")
			}
		})
	}
}

func TestSourceLane139NoBodyStatus(t *testing.T) {
	for _, tc := range []struct {
		name, responses string
		noBody          bool
	}{
		{"204", `{"204":{}}`, true}, {"205", `{"205":{}}`, true},
		{"204 and400", `{"204":{},"400":{}}`, true},
		{"204 and default", `{"204":{},"default":{}}`, true},
		{"absent", `null`, false}, {"empty", `{}`, false}, {"default", `{"default":{}}`, false},
		{"200", `{"200":{}}`, false}, {"299", `{"299":{}}`, false}, {"range", `{"2XX":{}}`, false},
		{"short sibling", `{"204":{},"20":{}}`, false},
		{"long sibling", `{"204":{},"2000":{}}`, false},
		{"nondigit sibling", `{"204":{},"2ab":{}}`, false},
		{"invalid sibling", `{"204":{},"2invalid":{}}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, facts, a := sourceBindingFixture099F(t, "body", nil)
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				op["responses"] = json.RawMessage(tc.responses)
			})
			// Retained normalization feeds the same no-body predicate used by
			// actual admitted write contract checking, alongside its classification.
			if got := sourceLaneNoResponseBody(facts); got != tc.noBody {
				t.Errorf("no-body=%v want%v", got, tc.noBody)
			}
			cells := classifySourceLanes(key, facts, &a)
			if len(cells) != 7 {
				t.Fatal("lost mutation lanes")
			}
			for _, cell := range cells {
				if cell.State == "implemented" || len(cell.ProofRefs) != 0 {
					t.Fatal("no-body manufactured proof")
				}
			}
		})
	}
}

// These are later controls added after initial same-test GREEN. The actual
// retained citations authorize the 200/299/2XX counterparts, while even a
// complete forged citation set cannot authorize a malformed status scope.
func TestSourceLane139StatusInterpretationAuthority(t *testing.T) {
	for _, tc := range []struct {
		status string
		valid  bool
	}{
		{"200", true}, {"299", true}, {"2XX", true},
		{"2ab", false}, {"20", false}, {"2000", false}, {"2invalid", false},
	} {
		t.Run(tc.status, func(t *testing.T) {
			key, facts, annotation := sourceBindingFixture099F(t, "envelope", nil)
			facts = sourceBindingRepin099F(t, key, facts, func(doc map[string]any) {
				op := doc["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)
				responses := op["responses"].(map[string]any)
				response := responses["200"]
				delete(responses, "200")
				responses[tc.status] = response
			})
			raw, err := json.Marshal(annotation)
			if err != nil {
				t.Fatal(err)
			}
			raw = []byte(strings.ReplaceAll(string(raw), "/responses/200/", "/responses/"+tc.status+"/"))
			if err := decodeStrictJSON(raw, &annotation); err != nil {
				t.Fatal(err)
			}
			kind, refs, _ := applySourceResponseInterpretations(key, facts, "read", annotation.ResponseInterpretations)
			if (kind == sourceCollection) != tc.valid {
				t.Errorf("typed collection=%v want%v", kind, tc.valid)
			}
			// Supply the entire genuinely cited interpretation independently of
			// the builder's acceptance. This exercises validator source authority.
			if !tc.valid {
				for _, interp := range annotation.ResponseInterpretations {
					refs = append(refs, interp.ResponseSchema, interp.Citation)
				}
			}
			manifest := sourceLaneManifest{
				Documents: []retainedSourceDocument{{ID: "fixture:099F", Payload: facts.Document}},
				SourceOperations: []sourceLaneManifestRow{{
					Source: retainedSourceOperation{Key: key, Observed: true, DocumentID: "fixture:099F", Pointer: "/rest/operations/0"},
					Lanes:  []sourceLaneCell{{Lane: "etl", RuleID: "source_response_interpretation", Applicability: "applicable", FactRefs: refs}},
				}},
			}
			annotations, err := json.Marshal([]sourceSemanticAnnotation{annotation})
			if err != nil {
				t.Fatal(err)
			}
			issues := validateSourceLaneInterpretationEvidence(manifest, annotations)
			if (len(issues) == 0) != tc.valid {
				t.Errorf("independent authority accepted=%v want%v: %+v", len(issues) == 0, tc.valid, issues)
			}
		})
	}
}
