package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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
	observed := map[sourceOperationKey]bool{}
	for _, row := range got.SourceOperations {
		if observed[row.Source.Key] || !row.Source.Observed {
			t.Fatalf("source duplicate or unobserved: %+v", row.Source.Key)
		}
		observed[row.Source.Key] = true
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
