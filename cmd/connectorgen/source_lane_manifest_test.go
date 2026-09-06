package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSourceLaneManifestAnnotationMembership(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	inventory := loadRetainedSourceInventory(context.Background(), root, cohort)
	row := inventory.Operations[0]
	facts := normalizeSourceFacts(row, inventory.Documents[0], nil)
	// Empty semantics carries no unsupported interpretation; it is a genuine
	// intended-binding annotation anchored to the retained method value.
	a := sourceSemanticAnnotation{Key: row.Key, Citation: facts.Refs["method"], Clause: "get"}
	for _, tc := range []struct {
		name        string
		annotations []sourceSemanticAnnotation
		invalid     bool
	}{
		{"valid", []sourceSemanticAnnotation{a}, false},
		{"duplicate", []sourceSemanticAnnotation{a, a}, true},
		{"unknown source", []sourceSemanticAnnotation{{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "source.other"}, Citation: a.Citation, Clause: a.Clause}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, tc.annotations)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 || got.SourceOperations[0].Source.Key.ID != "source.a" || got.SourceOperations[1].Source.Key.ID != "source.b" {
				t.Fatal("annotation changed independently anchored source rows/cells")
			}
			if (got.Validation.Status == "invalid") != tc.invalid {
				t.Fatalf("annotation membership invalid=%v; validation=%+v diagnostics=%+v", tc.invalid, got.Validation, got.Diagnostics)
			}
			for _, row := range got.SourceOperations {
				if len(row.Lanes) != 7 {
					t.Fatal("annotation failure removed a lane")
				}
			}
			encoded, err := json.Marshal(got)
			if err != nil || len(encoded) == 0 {
				t.Fatalf("complete diagnostic manifest not serializable: %v", err)
			}
		})
	}
}

func TestSourceLaneManifestRetainedCorpus(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	var cohort sourceLaneCohort
	if err := decodeStrictJSON(read("data/connector-canon/batch1-source-lane-cohort.json"), &cohort); err != nil {
		t.Fatal(err)
	}
	var annotations struct {
		SchemaVersion int                        `json:"schema_version"`
		Annotations   []sourceSemanticAnnotation `json:"annotations"`
	}
	if err := decodeStrictJSON(read("data/connector-canon/batch1-source-lane-annotations.json"), &annotations); err != nil {
		t.Fatal(err)
	}
	got, err := buildSourceLaneManifest(context.Background(), root, cohort, annotations.Annotations)
	if err != nil {
		t.Fatal(err)
	}
	if got.SourceTotals.Primary != 4341 || got.SourceTotals.Supplement != 2 || got.SourceTotals.Cells != 30401 || len(got.SourceOperations) != 4343 {
		t.Fatalf("source census drift: %+v", got.SourceTotals)
	}
	var expectedFixture struct {
		Keys []sourceOperationKey `json:"keys"`
	}
	if err := json.Unmarshal(read("cmd/connectorgen/testdata/source_lanes/batch1-expected-ids.json"), &expectedFixture); err != nil {
		t.Fatal(err)
	}
	expectedKeys := map[sourceOperationKey]bool{}
	for _, key := range expectedFixture.Keys {
		if expectedKeys[key] {
			t.Fatalf("independent fixture duplicate: %+v", key)
		}
		expectedKeys[key] = true
	}
	if len(expectedKeys) != 4343 {
		t.Fatalf("independent fixture has %d keys", len(expectedKeys))
	}
	observed := map[sourceOperationKey]bool{}
	for _, row := range got.SourceOperations {
		if observed[row.Source.Key] || !row.Source.Observed {
			t.Fatalf("source duplicate or unobserved: %+v", row.Source.Key)
		}
		observed[row.Source.Key] = true
		if !expectedKeys[row.Source.Key] {
			t.Errorf("unexpected source outside independent retained census: %+v", row.Source.Key)
		}
		if len(row.Lanes) != 7 {
			t.Fatalf("source lost lanes: %+v", row.Source.Key)
		}
		for i, lane := range sourceLaneNames() {
			if row.Lanes[i].Lane != lane {
				t.Fatalf("lane order drift: %+v", row.Source.Key)
			}
		}
		if row.Facts.Status == "unavailable" {
			t.Errorf("retained source facts lost: %+v %v", row.Source.Key, row.Facts.Diagnostics)
		}
	}
	for key := range expectedKeys {
		if !observed[key] {
			t.Errorf("independent expected source missing: %+v", key)
		}
	}
	if got.Validation.Status != "valid" {
		t.Fatalf("source-valid corpus failed: %+v", got.Validation)
	}
	t.Logf("primary=%d supplements=%d cells=%d documents=%d deficits=%d", got.SourceTotals.Primary, got.SourceTotals.Supplement, got.SourceTotals.Cells, len(got.Documents), got.Validation.Deficits)
}

func TestSourceLaneManifestNormalizationFailure(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	raw, err := os.ReadFile(filepath.Join(root, "source.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	rows := doc["rest"].(map[string]any)["operations"].([]any)
	delete(rows[0].(map[string]any), "source_operation")
	raw, err = json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	got, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 || got.SourceOperations[0].Source.Key.ID != "source.a" || got.SourceOperations[1].Source.Key.ID != "source.b" {
		t.Fatal("normalization failure removed anchored identities")
	}
	if got.SourceOperations[0].Facts.Status != "unavailable" || got.SourceOperations[1].Facts.Status != "available" {
		t.Fatal("actual normalizer boundary or unaffected sibling was lost")
	}
	reached := false
	for _, d := range got.Diagnostics {
		if d.Stage == "normalization" && d.Key.ID == "source.a" && d.Code == "source_operation_not_retained" && d.Severity == "error" {
			reached = true
		}
	}
	if !reached || got.Validation.Status != "invalid" {
		t.Fatalf("actual normalization error hidden in success report: %+v %+v", got.Validation, got.Diagnostics)
	}
}

func TestSourceLaneManifestBindingObservation(t *testing.T) {
	for _, which := range []string{"present target", "wrong existing target", "malformed canonical input"} {
		t.Run(which, func(t *testing.T) {
			root, _, _ := sourceBindingRepositoryFixture(t)
			route := "/widgets"
			if which == "wrong existing target" {
				route = "/other"
			}
			raw, err := json.Marshal(map[string]any{"schema_version": 2, "connector": "acme", "counts": map[string]int{"total": 2}, "rest": map[string]any{"operations": []any{
				map[string]any{"id": "source.a", "protocol": "rest", "method": "GET", "path": route, "source_operation": map[string]any{"summary": "Get widgets", "responses": map[string]any{"200": map[string]any{"description": "Unresolved response"}}}},
				map[string]any{"id": "source.b", "protocol": "rest", "method": "GET", "path": "/widgets", "source_operation": map[string]any{"summary": "Get widgets", "responses": map[string]any{"200": map[string]any{"description": "Unresolved response"}}}},
			}}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "source.json"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			cohort := sourceLaneCohort{SchemaVersion: 1, CohortID: "fixture", Inventories: []sourceInventoryAnchor{{Connector: "acme", Inventory: "primary", Class: "primary", Path: "source.json", SHA256: sourceBytesHash(raw), ExpectedIDs: []string{"source.a", "source.b"}, ExpectedCount: 2}}}
			inventory := loadRetainedSourceInventory(context.Background(), root, cohort)
			row := inventory.Operations[0]
			facts := normalizeSourceFacts(row, inventory.Documents[0], nil)
			a := sourceSemanticAnnotation{Key: row.Key, Citation: facts.Refs["summary"], Clause: "Get widgets", IntendedBindings: []sourceLaneTargetRef{{Kind: "operation", Connector: "acme", ID: "widgets.get", Lane: "direct_read", Artifact: "internal/connectors/defs/acme/operations.json"}}}
			if which == "malformed canonical input" {
				if err := os.WriteFile(filepath.Join(root, "internal/connectors/defs/acme/source.lock.json"), []byte(`{"schema_version":4,"operations":`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			before := sourceBindingFixtureSnapshot(t, root)
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, []sourceSemanticAnnotation{a})
			if err != nil {
				t.Fatal(err)
			}
			if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 {
				t.Fatal("binding observation changed source universe")
			}
			want := "target_contract_unverified"
			stage := "reference"
			severity := "deficit"
			if which == "wrong existing target" {
				want = "target_semantics_mismatch"
				severity = "error"
			}
			if which == "malformed canonical input" {
				want = "canonical_input_invalid"
				stage = "canonical_import"
			}
			found := false
			for _, d := range got.Diagnostics {
				if d.Key == row.Key && d.Code == want && d.Stage == stage && d.Severity == severity {
					found = true
				}
			}
			if !found {
				t.Fatalf("actual %s observation not joined to source: %+v", which, got.Diagnostics)
			}
			if (got.Validation.Status == "invalid") != (which == "wrong existing target") {
				t.Fatalf("unproven availability and false semantics conflated: %+v", got.Validation)
			}
			if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
				t.Fatal("manifest binding observation mutated fixture")
			}
		})
	}
}

func TestSourceLaneManifestProofFailurePreservation(t *testing.T) {
	for _, which := range []string{"empty proof", "malformed proof", "orphan assertion"} {
		t.Run(which, func(t *testing.T) {
			root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
			proofDocument(t, root, []sourceLaneProofRecord{})
			if which == "malformed proof" {
				proofWrite(t, root, sourceLaneProofPath, []byte(`{"schema_version":1,"records":`))
			}
			if which == "orphan assertion" {
				proofDocument(t, root, []sourceLaneProofRecord{{ID: "orphan", Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "source.other"}, Lane: "direct_read", ClaimCurrent: true}})
			}
			before := sourceBindingFixtureSnapshot(t, root)
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 {
				t.Fatal("proof failure removed source rows/cells")
			}
			if (got.Validation.Status == "invalid") != (which != "empty proof") {
				t.Fatalf("invalid proof silently disappeared: %+v %+v", got.Validation, got.Diagnostics)
			}
			if which == "orphan assertion" {
				found := false
				for _, d := range got.Diagnostics {
					if d.Key.ID == "source.other" && d.Stage == "proof" && d.Severity == "error" {
						found = true
					}
				}
				if !found {
					t.Fatal("orphan claimed proof was suppressed by anchored iteration")
				}
			}
			for _, row := range got.SourceOperations {
				if row.Source.Key.ID != "source.a" && row.Source.Key.ID != "source.b" {
					t.Fatal("proof supplied source membership")
				}
				if row.Lanes[0].State != "mapped_unproven" || !proofHasDiagnostic(row.Lanes[0].Diagnostics, "proof_unavailable", "deficit") {
					t.Fatal("empty/invalid proof did not retain explicit unproven lane")
				}
			}
			if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
				t.Fatal("proof phase mutated fixture")
			}
		})
	}
}

func TestSourceLaneManifestArrayEncoding(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	got, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Rows []struct {
			Lanes []map[string]json.RawMessage `json:"lanes"`
		} `json:"source_operations"`
	}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	for rowIndex, row := range decoded.Rows {
		for laneIndex, lane := range row.Lanes {
			for _, field := range []string{"fact_refs", "intended_bindings", "references", "proof_refs", "owner_refs", "gap_refs", "diagnostics"} {
				value := lane[field]
				if len(value) == 0 || value[0] != '[' {
					t.Errorf("row%d lane%d %s must be JSON array, got%s", rowIndex, laneIndex, field, value)
				}
			}
		}
	}
}

func TestSourceLaneManifestExactValidation(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	expected, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	copyManifest := func() sourceLaneManifest {
		t.Helper()
		raw, err := json.Marshal(expected)
		if err != nil {
			t.Fatal(err)
		}
		var result sourceLaneManifest
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	for _, tc := range []struct {
		name    string
		change  func(*sourceLaneManifest)
		invalid bool
	}{
		{"valid", func(*sourceLaneManifest) {}, false},
		{"same count wrong source", func(m *sourceLaneManifest) { m.SourceOperations[0].Source.Key.ID = "source.other" }, true},
		{"wrong normalized path", func(m *sourceLaneManifest) { m.SourceOperations[0].Facts.Path = "/other" }, true},
		{"lost retained fact", func(m *sourceLaneManifest) { delete(m.SourceOperations[0].Facts.Groups, "responses") }, true},
		{"substituted source bytes", func(m *sourceLaneManifest) { m.Documents[0].Payload = json.RawMessage(`{"schema_version":2}`) }, true},
		{"false implementation", func(m *sourceLaneManifest) {
			m.SourceOperations[0].Lanes[0].State = "implemented"
			m.SourceOperations[0].Lanes[0].ProofRefs = []string{"invented"}
		}, true},
		{"suppressed diagnostic", func(m *sourceLaneManifest) { m.Diagnostics = []sourceLaneDiagnostic{}; m.Validation.Deficits = 0 }, true},
		{"wrong summary", func(m *sourceLaneManifest) { m.LaneSummary[0].Primary["implemented"] = 2 }, true},
		{"reduced self-consistent set", func(m *sourceLaneManifest) {
			m.SourceOperations = m.SourceOperations[:1]
			m.SourceTotals.Primary = 1
			m.SourceTotals.Operations = 1
			m.SourceTotals.Cells = 7
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := copyManifest()
			tc.change(&candidate)
			got := validateSourceLaneManifest(candidate, expected)
			if (len(got) > 0) != tc.invalid {
				t.Fatalf("semantic validator invalid=%v diagnostics=%+v", tc.invalid, got)
			}
		})
	}
}

func TestSourceLaneManifestDocumentReference(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	original, err := os.ReadFile(filepath.Join(root, "source.json"))
	if err != nil {
		t.Fatal(err)
	}
	var input map[string]any
	if err := json.Unmarshal(original, &input); err != nil {
		t.Fatal(err)
	}
	first := input["rest"].(map[string]any)["operations"].([]any)[0].(map[string]any)
	first["opaque_provider_extension"] = map[string]any{"values": []any{json.Number("9007199254740993"), true, nil}}
	original, err = json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "source.json"), original, 0600); err != nil {
		t.Fatal(err)
	}
	cohort.Inventories[0].SHA256 = sourceBytesHash(original)
	result, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var restored sourceLaneManifest
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	row := restored.SourceOperations[0]
	if row.Source.DocumentID != restored.Documents[0].ID || row.Source.Pointer != "/rest/operations/0" {
		t.Fatal("source document association lost")
	}
	actual, err := sourceJSONPointer(restored.Documents[0].Payload, row.Source.Pointer+"/opaque_provider_extension/values/0")
	if err != nil || string(actual) != "9007199254740993" {
		t.Fatalf("opaque provider facts lost or rounded: %s %v", actual, err)
	}
	var decoded struct {
		Rows []struct {
			Source map[string]json.RawMessage `json:"source"`
		} `json:"source_operations"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, duplicated := decoded.Rows[0].Source["source_node"]; duplicated {
		t.Fatal("report repeats the full provider node instead of its existing immutable document reference")
	}
}

func TestSourceLaneManifestObservedCounts(t *testing.T) {
	for _, missing := range []bool{false, true} {
		name := "retained"
		if missing {
			name = "missing"
		}
		t.Run(name, func(t *testing.T) {
			root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
			if missing {
				if err := os.Remove(filepath.Join(root, "source.json")); err != nil {
					t.Fatal(err)
				}
			}
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var actual struct {
				Totals map[string]int `json:"source_totals"`
			}
			if err := json.Unmarshal(raw, &actual); err != nil {
				t.Fatal(err)
			}
			expected := 2
			if missing {
				expected = 0
			}
			for field, want := range map[string]int{"primary": 2, "supplement": 0, "operations": 2, "cells": 14, "observed_primary": expected, "observed_supplement": 0, "observed_operations": expected} {
				value, present := actual.Totals[field]
				if !present || value != want {
					t.Errorf("%s present=%v value=%d want=%d", field, present, value, want)
				}
			}
			if len(got.SourceOperations) != 2 || got.SourceOperations[0].Source.Key.ID != "source.a" || got.SourceOperations[1].Source.Key.ID != "source.b" {
				t.Fatal("observed availability changed anchored keys")
			}
		})
	}
}

func TestSourceLaneManifestReachedNormalizationFailures(t *testing.T) {
	for _, cancelAfterFirst := range []bool{false, true} {
		name := "post-normalization error"
		if cancelAfterFirst {
			name = "post-normalization cancellation"
		}
		t.Run(name, func(t *testing.T) {
			root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			before := sourceBindingFixtureSnapshot(t, root)
			var witnessed []string
			result, err := buildSourceLaneManifestObserved(ctx, root, cohort, nil, func(key sourceOperationKey, facts sourceFacts) error {
				if facts.Status != "available" || facts.Method != "GET" || facts.Path != "/items" {
					t.Fatalf("observer did not follow actual normalization: %+v", facts)
				}
				witnessed = append(witnessed, key.ID)
				if key.ID == "source.a" {
					if cancelAfterFirst {
						cancel()
						return nil
					}
					return errors.New("actual post-normalization cut")
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(result.SourceOperations) != 2 || result.SourceTotals.Cells != 14 || result.SourceOperations[0].Source.Key.ID != "source.a" || result.SourceOperations[1].Source.Key.ID != "source.b" {
				t.Fatal("failure lost independent source keys/cells")
			}
			if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
				t.Fatal("normalization failure changed retained fixture state")
			}
			wantCode := "source_normalization_failed"
			wantKey := "source.a"
			if cancelAfterFirst {
				wantCode = "source_normalization_cancelled"
				wantKey = "source.b"
				if !reflect.DeepEqual(witnessed, []string{"source.a"}) {
					t.Errorf("normalization continued after actual cancellation: %v", witnessed)
				}
			} else if !reflect.DeepEqual(witnessed, []string{"source.a", "source.b"}) {
				t.Errorf("normalization sibling not reached: %v", witnessed)
			}
			found := false
			for _, d := range result.Diagnostics {
				if d.Code == wantCode && d.Key.ID == wantKey && d.Stage == "normalization" && d.Severity == "error" {
					found = true
				}
			}
			if !found || result.Validation.Status != "invalid" {
				t.Errorf("reached failure omitted: code=%s validation=%+v diagnostics=%+v", wantCode, result.Validation, result.Diagnostics)
			}
		})
	}
}

func TestSourceLaneManifestFactCitationOracle(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	original, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		mutate func(*sourceLaneManifest)
	}{
		{"valid", nil},
		{"self consistent wrong responses", func(m *sourceLaneManifest) {
			f := &m.SourceOperations[0].Facts
			f.Groups["responses"] = json.RawMessage(`{"200":{"description":"invented response"}}`)
			r := f.Refs["responses"]
			r.ValueSHA256 = sourceBytesHash(f.Groups["responses"])
			f.Refs["responses"] = r
		}},
		{"wrong copied method", func(m *sourceLaneManifest) { m.SourceOperations[0].Facts.Method = "POST" }},
		{"missing citation", func(m *sourceLaneManifest) { delete(m.SourceOperations[0].Facts.Refs, "responses") }},
		{"omitted fact and citation", func(m *sourceLaneManifest) {
			delete(m.SourceOperations[0].Facts.Groups, "responses")
			delete(m.SourceOperations[0].Facts.Refs, "responses")
		}},
		{"null fact and omitted citation", func(m *sourceLaneManifest) {
			m.SourceOperations[0].Facts.Groups["responses"] = json.RawMessage("null")
			delete(m.SourceOperations[0].Facts.Refs, "responses")
		}},
		{"omitted operation and response", func(m *sourceLaneManifest) {
			for _, name := range []string{"source_operation", "responses"} {
				delete(m.SourceOperations[0].Facts.Groups, name)
				delete(m.SourceOperations[0].Facts.Refs, name)
			}
		}},
		{"absent document", func(m *sourceLaneManifest) {
			r := m.SourceOperations[0].Facts.Refs["responses"]
			r.DocumentID = "other-document"
			m.SourceOperations[0].Facts.Refs["responses"] = r
		}},
		{"wrong cited digest", func(m *sourceLaneManifest) {
			r := m.SourceOperations[0].Facts.Refs["responses"]
			r.ValueSHA256 = sourceBytesHash([]byte("other"))
			m.SourceOperations[0].Facts.Refs["responses"] = r
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var actual sourceLaneManifest
			if err := json.Unmarshal(raw, &actual); err != nil {
				t.Fatal(err)
			}
			if tc.mutate != nil {
				tc.mutate(&actual)
			}
			// Both copied projections agree. Only the independently retained document
			// can disconfirm the bad copied fact or fabricated citation.
			findings := validateSourceLaneManifest(actual, actual)
			if tc.mutate == nil {
				if len(findings) != 0 {
					t.Fatalf("valid retained counterpart rejected: %+v", findings)
				}
				return
			}
			found := false
			for _, d := range findings {
				if d.Key.ID == "source.a" && d.Stage == "source_fact_validation" && d.Severity == "error" {
					found = true
				}
			}
			if !found {
				t.Fatalf("same corrupt projection passed retained-document oracle: %+v", findings)
			}
		})
	}
}

func TestSourceLaneRetainedPointerCoordinates(t *testing.T) {
	var root any
	if err := decodeSourceJSON([]byte(`{"":false,"a/b":{"~key":[{"number":9007199254740993}]}}`), &root); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		pointer string
		want    string
	}{
		{"/", "false"},
		{"/a~1b/~0key/0/number", "9007199254740993"},
		{"/a~1b/~0key/0", `{"number":9007199254740993}`},
	} {
		t.Run(tc.pointer, func(t *testing.T) {
			got, err := sourceLaneRetainedPointer(root, tc.pointer)
			if err != nil || string(got) != tc.want {
				t.Fatalf("retained pointer = %s, %v; want %s", got, err, tc.want)
			}
		})
	}
	whole, err := sourceLaneRetainedPointer(root, "")
	if err != nil || string(whole) != `{"":false,"a/b":{"~key":[{"number":9007199254740993}]}}` {
		t.Fatalf("root pointer = %s, %v", whole, err)
	}
	for _, pointer := range []string{
		"a/b", "/a~2b", "/a~", "/absent", "/a~1b/~0key/00", "/a~1b/~0key/+0",
		"/a~1b/~0key/-", "/a~1b/~0key/1", "/a~1b/~0key/0/number/x",
		strings.Repeat("/x", 257),
	} {
		if _, err := sourceLaneRetainedPointer(root, pointer); err == nil {
			t.Errorf("invalid pointer accepted: %q", pointer)
		}
	}
}

func TestSourceLaneManifestEffectiveParameterOracle(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	path := filepath.Join(root, "source.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	rows := document["rest"].(map[string]any)["operations"].([]any)
	rows[0].(map[string]any)["source_operation"].(map[string]any)["parameters"] = []any{
		map[string]any{"name": "limit", "in": "query", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}},
	}
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	original, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(original.SourceOperations[0].Facts.Parameters) != 1 {
		t.Fatal("retained literal query limit was not normalized")
	}
	for _, tc := range []struct {
		name   string
		mutate func(*sourceFacts)
	}{
		{"valid retained parameter", nil},
		{"requiredness changed", func(f *sourceFacts) { f.Parameters[0].Required = false }},
		{"location changed", func(f *sourceFacts) { f.Parameters[0].In = "header" }},
		{"name changed", func(f *sourceFacts) { f.Parameters[0].Name = "other" }},
		{"resolved schema changed", func(f *sourceFacts) {
			f.Parameters[0].Node = json.RawMessage(`{"name":"limit","in":"query","required":true,"schema":{"type":"string"}}`)
		}},
		{"parameter omitted", func(f *sourceFacts) { f.Parameters = []sourceParameterFact{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			encoded, _ := json.Marshal(original)
			var actual sourceLaneManifest
			if err := json.Unmarshal(encoded, &actual); err != nil {
				t.Fatal(err)
			}
			if tc.mutate != nil {
				tc.mutate(&actual.SourceOperations[0].Facts)
			}
			findings := validateSourceLaneManifest(actual, actual)
			found := false
			for _, d := range findings {
				found = found || (d.Key.ID == "source.a" && d.Stage == "source_fact_validation" && d.Code == "source_parameter_projection_mismatch")
			}
			if tc.mutate == nil && len(findings) != 0 {
				t.Fatalf("valid retained counterpart rejected: %+v", findings)
			}
			if tc.mutate != nil && !found {
				t.Fatalf("self-consistent false parameter projection accepted: %+v", findings)
			}
		})
	}
}

func TestSourceLaneManifestSourceInputBytes(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	path := filepath.Join(root, "source.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Retained file-byte identity includes whitespace; normalized document bytes
	// are not the authority for the original input size.
	raw = append(append([]byte(" \n"), raw...), '\n')
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	result, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, pin := range result.Inputs {
		if pin.Path == "source.json" {
			if pin.Bytes != int64(len(raw)) || pin.SHA256 != sourceBytesHash(raw) {
				t.Fatalf("retained input identity lost: %+v; actual bytes=%d", pin, len(raw))
			}
			return
		}
	}
	t.Fatal("actual retained source input missing")
}

// These are actual collector failures after a known-good loaded/admitted control,
// not injected normalization errors relabeled as later pipeline phases.
func TestSourceLaneManifestCollectorFailurePreservation(t *testing.T) {
	for _, which := range []string{"control", "import", "admission", "execution load"} {
		t.Run(which, func(t *testing.T) {
			root, _, _ := sourceBindingRepositoryFixture(t)
			retained := []byte(`{"schema_version":2,"connector":"acme","counts":{"total":2},"rest":{"operations":[{"id":"source.a","protocol":"rest","method":"GET","path":"/widgets","source_operation":{"summary":"Get widgets","responses":{"200":{"description":"Unknown body"}}}},{"id":"source.b","protocol":"rest","method":"GET","path":"/widgets","source_operation":{"summary":"Get widgets","responses":{"200":{"description":"Unknown body"}}}}]}}`)
			if err := os.WriteFile(filepath.Join(root, "source.json"), retained, 0600); err != nil {
				t.Fatal(err)
			}
			cohort := sourceLaneCohort{SchemaVersion: 1, CohortID: "fixture", Inventories: []sourceInventoryAnchor{{Connector: "acme", Inventory: "primary", Class: "primary", Path: "source.json", SHA256: sourceBytesHash(retained), ExpectedIDs: []string{"source.a", "source.b"}, ExpectedCount: 2}}}
			positive := collectSourceLaneBindings(context.Background(), root, cohort)
			if len(positive.Observations) != 0 || len(positive.Canonical) != 1 || len(positive.Bundles) != 1 {
				t.Fatalf("control did not reach loaded and admitted results: %+v", positive.Observations)
			}
			stage, code := "", ""
			lockPath := filepath.Join(root, "internal/connectors/defs/acme/source.lock.json")
			switch which {
			case "import":
				if err := os.WriteFile(lockPath, []byte(`{"schema_version":4,"operations":`), 0600); err != nil {
					t.Fatal(err)
				}
				stage, code = "canonical_import", "canonical_input_invalid"
			case "admission":
				lock := vNextRequestSchemaLockForSemanticAdmissionTest()
				if _, err := canonicalizeVNextSourceLock(lock); err != nil {
					t.Fatalf("request positive not admitted: %v", err)
				}
				lock.Schemas["schemas/other-request.json"] = json.RawMessage(`{"type":"object","properties":{"count":{"type":"integer"}},"required":["count"],"additionalProperties":false}`)
				lock.Operations[0].SchemaRefs.Request = "schemas/other-request.json"
				raw, err := json.Marshal(lock)
				if err != nil {
					t.Fatal(err)
				}
				decoded, err := decodeVNextSourceLock(raw)
				if err != nil {
					t.Fatalf("admission fixture failed earlier import: %v", err)
				}
				if _, err := canonicalizeVNextSourceLock(decoded); err == nil || !strings.Contains(err.Error(), "/operations/0/schema_refs/request") {
					t.Fatalf("fixture did not reach selected schema admission: %v", err)
				}
				if err := os.WriteFile(lockPath, raw, 0600); err != nil {
					t.Fatal(err)
				}
				stage, code = "canonical_generation", "canonical_generation_unavailable"
			case "execution load":
				if err := os.WriteFile(filepath.Join(root, "internal/connectors/defs/acme/operations.json"), []byte(`{"operations":`), 0600); err != nil {
					t.Fatal(err)
				}
				stage, code = "execution_load", "execution_bundle_invalid"
			}
			before := sourceBindingFixtureSnapshot(t, root)
			observed := collectSourceLaneBindings(context.Background(), root, cohort)
			_, canonicalPresent := observed.Canonical["acme"]
			_, bundlePresent := observed.Bundles["acme"]
			if canonicalPresent != (which == "control" || which == "execution load") || bundlePresent != (which != "execution load") {
				t.Fatalf("wrong reached-stage results: canonical=%v bundle=%v observations=%+v", canonicalPresent, bundlePresent, observed.Observations)
			}
			got, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.SourceOperations) != 2 || got.SourceTotals.Cells != 14 || got.Validation.Status != "valid" {
				t.Fatalf("unclaimed execution failure changed membership or became source invalidity: %+v %+v", got.SourceTotals, got.Validation)
			}
			for i, row := range got.SourceOperations {
				wantID := []string{"source.a", "source.b"}[i]
				if row.Source.Key != (sourceOperationKey{Connector: "acme", Inventory: "primary", ID: wantID}) || !row.Source.Observed || row.Facts.Method != "GET" || row.Facts.Path != "/widgets" || len(row.Lanes) != 7 {
					t.Fatalf("source row changed at %s: %+v", which, row.Source)
				}
				for n, lane := range sourceLaneNames() {
					if row.Lanes[n].Lane != lane || row.Lanes[n].State == "implemented" {
						t.Fatalf("lane lost or falsely promoted: %+v", row.Lanes[n])
					}
				}
				found := false
				for _, d := range got.Diagnostics {
					if d.Key == row.Source.Key && d.Stage == stage && d.Code == code && d.Severity == "deficit" && reflect.DeepEqual(d.Lanes, sourceLaneNames()) {
						found = true
					}
				}
				if found != (which != "control") {
					t.Fatalf("named observation missing or invented for %s: %+v", wantID, got.Diagnostics)
				}
			}
			if after := sourceBindingFixtureSnapshot(t, root); !reflect.DeepEqual(before, after) {
				t.Fatal("collector/manifest altered pre-retained input bytes")
			}
		})
	}
}

func TestSourceLaneManifestSharedFactOmission(t *testing.T) {
	root, cohort := sourceInventoryFixture(t, []string{"source.a", "source.b"}, 2)
	p := filepath.Join(root, "source.json")
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	document["source_contract"] = map[string]any{
		"security":   []any{map[string]any{"bearer": []any{}}},
		"components": map[string]any{"securitySchemes": map[string]any{"bearer": map[string]any{"type": "http", "scheme": "bearer"}}},
		"webhooks":   map[string]any{"changed": map[string]any{"post": map[string]any{"description": "Retained change event"}}},
	}
	rest := document["rest"].(map[string]any)
	rest["path_bridge"] = map[string]any{"source_prefix": "/api/v4", "connector_prefix": ""}
	rest["event_schema_inventory"] = []any{"changed"}
	rest["batch_action_inventory"] = []any{"create"}
	rest["operations"].([]any)[0].(map[string]any)["source_operation"].(map[string]any)["security"] = []any{}
	raw, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
	cohort.Inventories[0].SHA256 = sourceBytesHash(raw)
	original, err := buildSourceLaneManifest(context.Background(), root, cohort, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, group, value, pointer string
		row                         int
	}{
		{"operation override", "security", `[]`, "/rest/operations/0/source_operation/security", 0},
		{"inherited security", "security", `[{"bearer":[]}]`, "/source_contract/security", 1},
		{"schemes", "security_schemes", `{"bearer":{"scheme":"bearer","type":"http"}}`, "/source_contract/components/securitySchemes", 0},
		{"webhooks", "webhooks", `{"changed":{"post":{"description":"Retained change event"}}}`, "/source_contract/webhooks", 0},
		{"bridge", "path_bridge", `{"connector_prefix":"","source_prefix":"/api/v4"}`, "/rest/path_bridge", 0},
		{"events", "event_schema_inventory", `["changed"]`, "/rest/event_schema_inventory", 0},
		{"batch actions", "batch_action_inventory", `["create"]`, "/rest/batch_action_inventory", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, err := canonicalSourceJSON(original.SourceOperations[tc.row].Facts.Groups[tc.group])
			if err != nil || string(value) != tc.value || original.SourceOperations[tc.row].Facts.Refs[tc.group].Pointer != tc.pointer {
				t.Fatalf("literal positive missing: %s (%v)", value, err)
			}
			if findings := validateSourceLaneManifest(original, original); len(findings) != 0 {
				t.Fatalf("positive rejected: %+v", findings)
			}
			encoded, err := json.Marshal(original)
			if err != nil {
				t.Fatal(err)
			}
			var omitted sourceLaneManifest
			if err := json.Unmarshal(encoded, &omitted); err != nil {
				t.Fatal(err)
			}
			delete(omitted.SourceOperations[tc.row].Facts.Groups, tc.group)
			delete(omitted.SourceOperations[tc.row].Facts.Refs, tc.group)
			found := false
			for _, d := range validateSourceLaneManifest(omitted, omitted) {
				found = found || (d.Key == original.SourceOperations[tc.row].Source.Key && d.Stage == "source_fact_validation" && d.Code == "source_retained_fact_missing_or_changed" && d.Pointer == tc.pointer)
			}
			if !found {
				t.Fatal("joint omission of independently retained shared fact and citation accepted")
			}
		})
	}
}

func TestSourceLaneManifestRawSharedFactOmission(t *testing.T) {
	node := json.RawMessage(`{"id":"source.a","protocol":"rest","method":"GET","path":"/items"}`)
	doc := retainedSourceDocument{ID: "fixture:primary", Payload: json.RawMessage(`{"rest":{"operations":[` + string(node) + `]}}`)}
	rawDoc := retainedSourceDocument{ID: "fixture:raw", Payload: json.RawMessage(`{"security":[{"bearer":[]}],"paths":{"/items":{"parameters":[{"name":"limit","in":"query","schema":{"type":"integer"}}],"get":{"responses":{"200":{"description":"Unknown body"}}}}}}`)}
	row := retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "source.a"}, Observed: true, Node: node, DocumentID: doc.ID, RawDocumentID: rawDoc.ID, Pointer: "/rest/operations/0"}
	original := sourceLaneManifestRow{Source: row, Facts: normalizeSourceFacts(row, doc, &rawDoc)}
	documents := map[string]retainedSourceDocument{doc.ID: doc, rawDoc.ID: rawDoc}
	roots := map[string]any{}
	for id, d := range documents {
		var root any
		if err := decodeSourceJSON(d.Payload, &root); err != nil {
			t.Fatal(err)
		}
		roots[id] = root
	}
	if got := validateSourceLaneRequiredFactGroups(original, documents, roots); len(got) != 0 {
		t.Fatalf("raw positive rejected: %+v", got)
	}
	for _, tc := range []struct{ group, pointer, value string }{
		{"security", "/security", `[{"bearer":[]}]`},
		{"path_parameters", "/paths/~1items/parameters", `[{"in":"query","name":"limit","schema":{"type":"integer"}}]`},
	} {
		t.Run(tc.group, func(t *testing.T) {
			value, err := canonicalSourceJSON(original.Facts.Groups[tc.group])
			if err != nil || string(value) != tc.value || original.Facts.Refs[tc.group].DocumentID != rawDoc.ID || original.Facts.Refs[tc.group].Pointer != tc.pointer {
				t.Fatalf("raw positive literal/owner absent: %s %v", value, err)
			}
			encoded, _ := json.Marshal(original)
			var omitted sourceLaneManifestRow
			if err := json.Unmarshal(encoded, &omitted); err != nil {
				t.Fatal(err)
			}
			delete(omitted.Facts.Groups, tc.group)
			delete(omitted.Facts.Refs, tc.group)
			got := validateSourceLaneRequiredFactGroups(omitted, documents, roots)
			if len(got) != 1 || got[0].Pointer != tc.pointer || got[0].Code != "source_retained_fact_missing_or_changed" {
				t.Fatalf("raw source omission not diagnosed: %+v", got)
			}
		})
	}
}

func TestSourceLaneManifestProviderFactCounterexamples(t *testing.T) {
	for _, tc := range []struct {
		name, id, group string
		path            []string
		want            string
	}{
		{"session ID", "vercel.rest.readSessionFile", "parameters", nil, "sessionId"},
		{"required file path", "vercel.rest.readSessionFile", "request_body", []string{"content", "application/json", "schema", "required"}, `["path"]`},
		{"binary response", "vercel.rest.readSessionFile", "responses", []string{"200", "content", "application/octet-stream", "schema", "format"}, `"binary"`},
		{"required events", "vercel.rest.createWebhook", "request_body", []string{"content", "application/json", "schema", "required"}, `["url","events"]`},
		{"minimum events", "vercel.rest.createWebhook", "request_body", []string{"content", "application/json", "schema", "properties", "events", "minItems"}, `1`},
		{"event enum", "vercel.rest.createWebhook", "request_body", []string{"content", "application/json", "schema", "properties", "events", "items", "enum"}, "budget.reached"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			row, doc := retainedFactFixture(t, "vercel", tc.id)
			facts := normalizeSourceFacts(row, doc, nil)
			candidate := sourceLaneManifest{Documents: []retainedSourceDocument{doc}, SourceOperations: []sourceLaneManifestRow{{Source: row, Facts: facts}}}
			if findings := validateSourceLaneFactCitations(candidate); len(findings) != 0 {
				t.Fatalf("retained positive rejected: %+v", findings)
			}
			var value any
			if err := json.Unmarshal(facts.Groups[tc.group], &value); err != nil {
				t.Fatal(err)
			}
			if tc.path == nil {
				params := value.([]any)
				first := params[0].(map[string]any)
				if first["name"] != tc.want || first["in"] != "path" || first["required"] != true {
					t.Fatalf("literal required session ID absent: %+v", first)
				}
				value = params[1:]
			} else {
				parent := value.(map[string]any)
				for _, key := range tc.path[:len(tc.path)-1] {
					parent = parent[key].(map[string]any)
				}
				last := tc.path[len(tc.path)-1]
				if tc.name == "event enum" {
					enums := parent[last].([]any)
					if len(enums) == 0 || enums[0] != tc.want {
						t.Fatalf("literal event enum absent: %+v", enums)
					}
				} else {
					literal, err := json.Marshal(parent[last])
					if err != nil || string(literal) != tc.want {
						t.Fatalf("literal provider fact = %s want %s", literal, tc.want)
					}
				}
				delete(parent, last)
			}
			mutated, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			candidate.SourceOperations[0].Facts.Groups[tc.group] = mutated
			found := false
			for _, d := range validateSourceLaneFactCitations(candidate) {
				found = found || (d.Key == row.Key && d.Code == "source_fact_value_mismatch" && d.Pointer == facts.Refs[tc.group].Pointer)
			}
			if !found {
				t.Fatal("readable omitted provider fact was accepted against retained source")
			}
		})
	}
}
