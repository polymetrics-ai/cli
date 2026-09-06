package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type sourceLaneTotals struct {
	ObservedPrimary    int `json:"observed_primary"`
	ObservedSupplement int `json:"observed_supplement"`
	ObservedOperations int `json:"observed_operations"`
	Primary            int `json:"primary"`
	Supplement         int `json:"supplement"`
	Operations         int `json:"operations"`
	Cells              int `json:"cells"`
}
type sourceLaneManifestRow struct {
	Source retainedSourceOperation `json:"source"`
	Facts  sourceFacts             `json:"facts"`
	Lanes  []sourceLaneCell        `json:"lanes"`
}
type sourceLaneSummary struct {
	Lane       string         `json:"lane"`
	Primary    map[string]int `json:"primary"`
	Supplement map[string]int `json:"supplement"`
}
type sourceLaneValidation struct {
	Status   string `json:"status"`
	Errors   int    `json:"errors"`
	Deficits int    `json:"deficits"`
}
type sourceLaneManifest struct {
	SchemaVersion    int                      `json:"schema_version"`
	Kind             string                   `json:"kind"`
	CohortID         string                   `json:"cohort_id"`
	SourceTotals     sourceLaneTotals         `json:"source_totals"`
	Inputs           []sourceArtifactPin      `json:"inputs"`
	Documents        []retainedSourceDocument `json:"documents"`
	SourceOperations []sourceLaneManifestRow  `json:"source_operations"`
	LaneSummary      []sourceLaneSummary      `json:"lane_summary"`
	Diagnostics      []sourceLaneDiagnostic   `json:"diagnostics"`
	Validation       sourceLaneValidation     `json:"validation"`
}

// buildSourceLaneManifest composes source inventory, normalized facts and lane
// rules without running or publishing a connector. Proof is integrated separately.
func buildSourceLaneManifest(ctx context.Context, repo string, cohort sourceLaneCohort, annotations []sourceSemanticAnnotation) (sourceLaneManifest, error) {
	return buildSourceLaneManifestObserved(ctx, repo, cohort, annotations, nil)
}

// The optional observer is a narrow test seam reached after real normalization.
// Production always uses nil and all retained input reads remain unchanged.
func buildSourceLaneManifestObserved(ctx context.Context, repo string, cohort sourceLaneCohort, annotations []sourceSemanticAnnotation, afterNormalize func(sourceOperationKey, sourceFacts) error) (sourceLaneManifest, error) {
	result := sourceLaneManifest{SchemaVersion: 1, Kind: "retained_source_lane_manifest", CohortID: cohort.CohortID, Inputs: []sourceArtifactPin{}, Documents: []retainedSourceDocument{}, SourceOperations: []sourceLaneManifestRow{}, LaneSummary: []sourceLaneSummary{}, Diagnostics: []sourceLaneDiagnostic{}}
	if err := validateSourceLaneCohort(cohort); err != nil {
		return result, fmt.Errorf("cohort anchor invalid")
	}
	inventory := loadRetainedSourceInventory(ctx, repo, cohort)
	result.Diagnostics = append(result.Diagnostics, inventory.Diagnostics...)
	docs := map[string]retainedSourceDocument{}
	for _, document := range inventory.Documents {
		prepared, err := prepareSourceDocument(document)
		if err == nil {
			document = prepared
		}
		docs[document.ID] = document
		result.Documents = append(result.Documents, document)
	}
	for _, anchor := range cohort.Inventories {
		result.Inputs = append(result.Inputs, sourceArtifactPin{Path: anchor.Path, SHA256: anchor.SHA256})
		result.Inputs = append(result.Inputs, anchor.Artifacts...)
	}

	annotationIndex := map[sourceOperationKey]*sourceSemanticAnnotation{}
	expected := map[sourceOperationKey]bool{}
	duplicate := map[sourceOperationKey]bool{}
	for _, source := range inventory.Operations {
		expected[source.Key] = true
	}
	for i := range annotations {
		a := &annotations[i]
		code := ""
		if !expected[a.Key] {
			code = "annotation_source_unknown"
		} else if _, exists := annotationIndex[a.Key]; exists {
			code = "annotation_source_duplicate"
			duplicate[a.Key] = true
		}
		if code != "" {
			result.Diagnostics = append(result.Diagnostics, sourceLaneDiagnostic{Key: a.Key, Lanes: sourceLaneNames(), Stage: "classification", Code: code, Pointer: "data/connector-canon/batch1-source-lane-annotations.json", Owner: a.Key.Connector, Severity: "error"})
		}
		annotationIndex[a.Key] = a
	}
	for key := range duplicate {
		annotationIndex[key] = nil
	}
	proofs := loadSourceLaneProofs(repo, nil)
	result.Diagnostics = append(result.Diagnostics, proofs.Diagnostics...)
	bindings := collectSourceLaneBindings(ctx, repo, cohort)
	for _, source := range inventory.Operations {
		for _, observation := range bindings.Observations {
			if observation.Connector == source.Key.Connector {
				result.Diagnostics = append(result.Diagnostics, sourceLaneDiagnostic{Key: source.Key, Lanes: sourceLaneNames(), Stage: observation.Stage, Code: observation.Code, Pointer: observation.Pointer, Owner: source.Key.Connector, Severity: "deficit"})
			}
		}
		var raw *retainedSourceDocument
		if doc, exists := docs[source.RawDocumentID]; exists {
			raw = &doc
		}
		var facts sourceFacts
		if ctx.Err() != nil {
			facts = sourceFacts{Status: "unavailable", Parameters: []sourceParameterFact{}, Groups: map[string]json.RawMessage{}, Refs: map[string]sourceFactRef{}, Diagnostics: []string{"source_normalization_cancelled"}, CoverageConfidence: "partial", CompletenessLimits: []string{"retained_snapshot_only_not_current_provider_completeness", "source_normalization_cancelled"}}
		} else {
			facts = normalizeSourceFacts(source, docs[source.DocumentID], raw)
			if afterNormalize != nil {
				if err := afterNormalize(source.Key, facts); err != nil {
					facts.Status = "unavailable"
					facts.Diagnostics = append(facts.Diagnostics, "source_normalization_failed")
					facts.CoverageConfidence = "partial"
					facts.CompletenessLimits = append(facts.CompletenessLimits, "source_normalization_failed")
				}
			}
		}

		for _, code := range facts.Diagnostics {
			severity := "deficit"
			if facts.Status == "unavailable" || strings.Contains(code, "invalid") {
				severity = "error"
			}
			result.Diagnostics = append(result.Diagnostics, sourceLaneDiagnostic{Key: source.Key, Lanes: sourceLaneNames(), Stage: "normalization", Code: code, Pointer: source.Pointer, Owner: source.Key.Connector, Severity: severity})
		}
		facts.bindings = &bindings
		annotation := annotationIndex[source.Key]
		cells := classifySourceLanes(source.Key, facts, annotation)
		cells = assessSourceLaneProof(source.Key, cells, proofs)
		// The manifest encodes empty diagnostic collections as arrays.
		for i := range cells {
			if cells[i].Diagnostics == nil {
				cells[i].Diagnostics = []sourceLaneDiagnostic{}
			}
		}
		result.SourceOperations = append(result.SourceOperations, sourceLaneManifestRow{Source: source, Facts: facts, Lanes: cells})
	}
	summarizeSourceLaneManifest(&result)
	return result, nil
}

func summarizeSourceLaneManifest(result *sourceLaneManifest) {
	result.SourceTotals = sourceLaneTotals{}
	result.LaneSummary = []sourceLaneSummary{}
	for _, lane := range sourceLaneNames() {
		states := func() map[string]int {
			return map[string]int{"implemented": 0, "mapped_unproven": 0, "missing_foundation": 0, "not_applicable": 0}
		}
		result.LaneSummary = append(result.LaneSummary, sourceLaneSummary{Lane: lane, Primary: states(), Supplement: states()})
	}
	for _, row := range result.SourceOperations {
		result.SourceTotals.Operations++
		if row.Source.Observed {
			result.SourceTotals.ObservedOperations++
			if row.Source.Class == "primary" {
				result.SourceTotals.ObservedPrimary++
			} else {
				result.SourceTotals.ObservedSupplement++
			}
		}
		if row.Source.Class == "primary" {
			result.SourceTotals.Primary++
		} else {
			result.SourceTotals.Supplement++
		}
		for i, cell := range row.Lanes {
			result.SourceTotals.Cells++
			if row.Source.Class == "primary" {
				result.LaneSummary[i].Primary[cell.State]++
			} else {
				result.LaneSummary[i].Supplement[cell.State]++
			}
			result.Diagnostics = append(result.Diagnostics, cell.Diagnostics...)
		}
	}
	sort.Slice(result.Inputs, func(i, j int) bool { return result.Inputs[i].Path < result.Inputs[j].Path })
	sort.Slice(result.Diagnostics, func(i, j int) bool {
		a, _ := json.Marshal(result.Diagnostics[i])
		b, _ := json.Marshal(result.Diagnostics[j])
		return string(a) < string(b)
	})
	result.Validation = sourceLaneValidation{Status: "valid"}
	for _, d := range result.Diagnostics {
		if d.Severity == "error" {
			result.Validation.Errors++
		} else if d.Severity == "deficit" {
			result.Validation.Deficits++
		}
	}
	if result.Validation.Errors > 0 {
		result.Validation.Status = "invalid"
	}
}

// validateSourceLaneManifest checks a supplied authoring report against the
// independently built report from current retained inputs.
func validateSourceLaneManifest(candidate, expected sourceLaneManifest) []sourceLaneDiagnostic {
	diagnostics := []sourceLaneDiagnostic{}
	add := func(key sourceOperationKey, code, pointer string, lanes []string) {
		diagnostics = append(diagnostics, sourceLaneDiagnostic{Key: key, Lanes: lanes, Stage: "manifest", Code: code, Pointer: pointer, Owner: key.Connector, Severity: "error"})
	}
	if candidate.SchemaVersion != 1 || candidate.Kind != "retained_source_lane_manifest" || candidate.CohortID != expected.CohortID {
		add(sourceOperationKey{}, "manifest_identity_invalid", "", sourceLaneNames())
	}
	if len(candidate.SourceOperations) != expected.SourceTotals.Operations || candidate.SourceTotals != expected.SourceTotals {
		add(sourceOperationKey{}, "manifest_counts_mismatch", "/source_totals", sourceLaneNames())
	}
	expectedRows := map[sourceOperationKey]sourceLaneManifestRow{}
	for _, row := range expected.SourceOperations {
		expectedRows[row.Source.Key] = row
	}
	for _, field := range []struct {
		name                string
		candidate, expected any
	}{
		{"inputs", candidate.Inputs, expected.Inputs}, {"documents", candidate.Documents, expected.Documents}, {"lane_summary", candidate.LaneSummary, expected.LaneSummary}, {"diagnostics", candidate.Diagnostics, expected.Diagnostics}, {"validation", candidate.Validation, expected.Validation},
	} {
		if !sourceLaneJSONEqual(field.candidate, field.expected) {
			add(sourceOperationKey{}, "manifest_"+field.name+"_mismatch", "/"+field.name, sourceLaneNames())
		}
	}

	seen := map[sourceOperationKey]bool{}
	for i, row := range candidate.SourceOperations {
		pointer := fmt.Sprintf("/source_operations/%d", i)
		expectedRow, exists := expectedRows[row.Source.Key]
		if !exists {
			add(row.Source.Key, "manifest_source_unexpected", pointer, sourceLaneNames())
		} else {
			if !sourceLaneJSONEqual(row.Source, expectedRow.Source) {
				add(row.Source.Key, "manifest_source_mismatch", pointer+"/source", sourceLaneNames())
			}
			if !sourceLaneJSONEqual(row.Facts, expectedRow.Facts) {
				add(row.Source.Key, "manifest_facts_mismatch", pointer+"/facts", sourceLaneNames())
			}
			for j, cell := range row.Lanes {
				if j >= len(expectedRow.Lanes) || !sourceLaneJSONEqual(cell, expectedRow.Lanes[j]) {
					add(row.Source.Key, "manifest_lane_mismatch", fmt.Sprintf("%s/lanes/%d", pointer, j), []string{cell.Lane})
				}
			}
		}

		if seen[row.Source.Key] {
			add(row.Source.Key, "manifest_source_duplicate", pointer, sourceLaneNames())
		}
		seen[row.Source.Key] = true
		if len(row.Lanes) != 7 {
			add(row.Source.Key, "manifest_lanes_mismatch", pointer+"/lanes", sourceLaneNames())
			continue
		}
		for j, lane := range sourceLaneNames() {
			if row.Lanes[j].Lane != lane {
				add(row.Source.Key, "manifest_lanes_mismatch", pointer+"/lanes", []string{lane})
			}
		}
	}
	for _, row := range expected.SourceOperations {
		if !seen[row.Source.Key] {
			add(row.Source.Key, "manifest_source_missing", "/source_operations", sourceLaneNames())
		}
	}
	return diagnostics
}

func sourceLaneJSONEqual(a, b any) bool {
	first, err := json.Marshal(a)
	if err != nil {
		return false
	}
	second, err := json.Marshal(b)
	return err == nil && bytes.Equal(first, second)
}
