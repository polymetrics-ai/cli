package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
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
	annotationInputs json.RawMessage          `json:"-"`
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
	result.annotationInputs, _ = json.Marshal(annotations)
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
		pin := sourceArtifactPin{Path: anchor.Path, SHA256: anchor.SHA256}
		if document, observed := docs[anchor.Connector+":"+anchor.Inventory]; observed &&
			document.Path == anchor.Path && document.RetainedFileSHA256 == anchor.SHA256 {
			pin.Bytes = document.Bytes
		}
		result.Inputs = append(result.Inputs, pin)
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
	anchoredKeys := make([]sourceOperationKey, 0, len(inventory.Operations))
	for _, source := range inventory.Operations {
		anchoredKeys = append(anchoredKeys, source.Key)
	}
	policy, policyErr := newSourceLaneProofPolicy(anchoredKeys)
	var proofs sourceLaneProofInputs
	if policyErr != nil {
		proofs.Diagnostics = []sourceLaneDiagnostic{{Lanes: sourceLaneNames(), Stage: "proof", Code: "proof_cohort_policy_invalid", Owner: "batch1", Severity: "error"}}
	} else {
		proofs = loadSourceLaneProofs(ctx, repo, policy, sourceLaneProofCatalog{})
	}
	result.Diagnostics = append(result.Diagnostics, proofs.Diagnostics...)
	var demandAtlasOwners map[string]string
	for _, annotation := range annotations {
		if annotation.FoundationGap != "" {
			var pin sourceArtifactPin
			demandAtlasOwners, pin = loadSourceDemandAtlas(ctx, repo)
			if pin.Path != "" {
				result.Inputs = append(result.Inputs, pin)
			}
			break
		}
	}
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
		facts.demandAtlasOwners = demandAtlasOwners
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
	result.Diagnostics = append(result.Diagnostics, validateSourceLaneFactCitations(result)...)
	result.Diagnostics = append(result.Diagnostics, validateSourceLaneInterpretationEvidence(result, result.annotationInputs)...)
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
		switch d.Severity {
		case "error":
			result.Validation.Errors++
		case "deficit":
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
	diagnostics := validateSourceLaneFactCitations(candidate)
	diagnostics = append(diagnostics, validateSourceLaneInterpretationEvidence(candidate, expected.annotationInputs)...)
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

// validateSourceLaneFactCitations verifies copied facts against retained source
// documents independently of a freshly generated report's copied values.
func validateSourceLaneFactCitations(candidate sourceLaneManifest) []sourceLaneDiagnostic {
	diagnostics := []sourceLaneDiagnostic{}
	documents := map[string]retainedSourceDocument{}
	roots := map[string]any{}
	invalid := map[string]bool{}
	for _, document := range candidate.Documents {
		if _, exists := documents[document.ID]; exists || document.ID == "" {
			invalid[document.ID] = true
		}
		documents[document.ID] = document
		if document.ContentType != "text/html" {
			view, err := sourceDocumentViewFor(document)
			if err != nil {
				invalid[document.ID] = true
			} else {
				roots[document.ID] = view.ReferenceRoot
			}
		}
	}
	resolve := func(ref sourceFactRef) (json.RawMessage, error) {
		document, exists := documents[ref.DocumentID]
		if !exists || invalid[ref.DocumentID] {
			return nil, fmt.Errorf("citation document unavailable")
		}
		if document.ContentType == "text/html" {
			return resolveSourceFactValue(document, ref)
		}
		if ref.Section != "" || ref.Part != "" {
			return nil, fmt.Errorf("rendered selector on JSON document")
		}
		return sourceLaneRetainedPointer(roots[ref.DocumentID], ref.Pointer)
	}
	for _, row := range candidate.SourceOperations {
		add := func(code, pointer string) {
			diagnostics = append(diagnostics, sourceLaneDiagnostic{Key: row.Source.Key, Lanes: sourceLaneNames(), Stage: "source_fact_validation", Code: code, Pointer: pointer, Owner: row.Source.Key.Connector, Severity: "error"})
		}
		names := make([]string, 0, len(row.Facts.Groups))
		for name := range row.Facts.Groups {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			copied := row.Facts.Groups[name]
			ref, cited := row.Facts.Refs[name]
			if !cited && bytes.Equal(bytes.TrimSpace(copied), []byte("null")) {
				continue
			}
			if !cited {
				add("source_fact_citation_missing", name)
				continue
			}
			actual, err := resolve(ref)
			canonical, copyErr := canonicalSourceJSON(copied)
			if err != nil || copyErr != nil {
				add("source_fact_citation_invalid", ref.Pointer)
				continue
			}
			actual, err = canonicalSourceJSON(actual)
			if err != nil || sourceBytesHash(actual) != ref.ValueSHA256 || !bytes.Equal(canonical, actual) {
				add("source_fact_value_mismatch", ref.Pointer)
			}
		}
		refNames := make([]string, 0, len(row.Facts.Refs))
		for name := range row.Facts.Refs {
			refNames = append(refNames, name)
		}
		sort.Strings(refNames)
		for _, name := range refNames {
			if _, exists := row.Facts.Groups[name]; !exists {
				add("source_fact_group_missing", name)
			}
		}
		for _, field := range []struct{ name, value string }{
			{"method", row.Facts.Method}, {"path", row.Facts.Path},
			{"protocol", row.Facts.Protocol}, {"operation_id", row.Facts.OperationID},
		} {
			var value string
			raw, exists := row.Facts.Groups[field.name]
			if exists && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				if err := json.Unmarshal(raw, &value); err != nil {
					add("source_fact_scalar_invalid", field.name)
					continue
				}
			}
			if field.name == "method" {
				value = strings.ToUpper(value)
			}
			if field.value != value {
				add("source_fact_scalar_mismatch", field.name)
			}
		}
		if row.Source.Observed {
			diagnostics = append(diagnostics, validateSourceLaneRequiredFactGroups(row, documents, roots)...)
		}
		parameterFacts := row.Facts
		documentID := row.Facts.Refs["source_operation"].DocumentID
		parameterFacts.Document = documents[documentID].Payload
		parameterFacts.referenceRoot = roots[documentID]
		parameterFacts.RefPrefix = ""
		if documentID == row.Source.DocumentID {
			parameterFacts.RefPrefix = "/source_contract"
		}
		parameters, _ := effectiveSourceParameters(parameterFacts, nil)
		if !sourceLaneJSONEqual(parameters, row.Facts.Parameters) {
			add("source_parameter_projection_mismatch", "/facts/effective_parameters")
		}
	}
	return diagnostics
}

// sourceLaneRetainedPointer resolves an RFC 6901 pointer over a prepared JSON
// root. It never follows remote references or reparses the full document.
func sourceLaneRetainedPointer(root any, pointer string) (json.RawMessage, error) {
	if pointer == "" {
		return json.Marshal(root)
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, fmt.Errorf("source pointer must be absolute")
	}
	parts := strings.Split(pointer[1:], "/")
	if len(parts) > 256 {
		return nil, fmt.Errorf("source pointer depth exceeded")
	}
	value := root
	for _, part := range parts {
		for i := 0; i < len(part); i++ {
			if part[i] == '~' {
				if i+1 == len(part) || (part[i+1] != '0' && part[i+1] != '1') {
					return nil, fmt.Errorf("invalid source pointer escape")
				}
				i++
			}
		}
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch node := value.(type) {
		case map[string]any:
			var exists bool
			value, exists = node[part]
			if !exists {
				return nil, fmt.Errorf("source pointer absent")
			}
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(node) || strconv.Itoa(index) != part {
				return nil, fmt.Errorf("source pointer array index invalid")
			}
			value = node[index]
		default:
			return nil, fmt.Errorf("source pointer traverses scalar")
		}
	}
	return json.Marshal(value)
}

// The retained row and operation, rather than the copied group list, determine
// which provider facts must be represented. This detects jointly omitted facts
// and citations without treating an absent provider property as a known value.
func validateSourceLaneRequiredFactGroups(row sourceLaneManifestRow, documents map[string]retainedSourceDocument, roots map[string]any) []sourceLaneDiagnostic {
	diagnostics := []sourceLaneDiagnostic{}
	add := func(code, pointer string) {
		diagnostics = append(diagnostics, sourceLaneDiagnostic{Key: row.Source.Key, Lanes: sourceLaneNames(), Stage: "source_fact_validation", Code: code, Pointer: pointer, Owner: row.Source.Key.Connector, Severity: "error"})
	}
	raw, err := sourceLaneRetainedPointer(roots[row.Source.DocumentID], row.Source.Pointer)
	var node map[string]json.RawMessage
	if err != nil || decodeSourceJSON(raw, &node) != nil || node == nil {
		add("source_row_citation_invalid", row.Source.Pointer)
		return diagnostics
	}
	require := func(name string, value json.RawMessage, documentID, pointer string) {
		if len(value) == 0 {
			return
		}
		ref, cited := row.Facts.Refs[name]
		copied, present := row.Facts.Groups[name]
		canonical, err := canonicalSourceJSON(value)
		actual, copyErr := canonicalSourceJSON(copied)
		if !present || !cited || err != nil || copyErr != nil || !bytes.Equal(canonical, actual) ||
			ref.DocumentID != documentID || ref.Pointer != pointer || ref.Section != "" || ref.Part != "" || ref.ValueSHA256 != sourceBytesHash(canonical) {
			add("source_retained_fact_missing_or_changed", pointer)
		}
	}
	for _, name := range []string{"method", "path", "protocol", "operation_id"} {
		require(name, node[name], row.Source.DocumentID, row.Source.Pointer+"/"+name)
	}
	documentID := row.Source.DocumentID
	pointer := row.Source.Pointer + "/source_operation"
	operationRaw := node["source_operation"]
	if len(operationRaw) == 0 && row.Source.RawDocumentID != "" && documents[row.Source.RawDocumentID].ContentType != "text/html" {
		var method, path string
		if json.Unmarshal(node["method"], &method) != nil || json.Unmarshal(node["path"], &path) != nil {
			add("source_operation_coordinates_invalid", row.Source.Pointer)
			return diagnostics
		}
		documentID = row.Source.RawDocumentID
		pointer = "/paths/" + escapeSourcePointer(path) + "/" + strings.ToLower(method)
		operationRaw, err = sourceLaneRetainedPointer(roots[documentID], pointer)
		if err != nil {
			add("source_operation_citation_invalid", pointer)
			return diagnostics
		}
	}
	if len(operationRaw) == 0 {
		return diagnostics // Rendered/unknown operations have no invented JSON facts.
	}
	require("source_operation", operationRaw, documentID, pointer)
	var operation map[string]json.RawMessage
	if decodeSourceJSON(operationRaw, &operation) != nil {
		add("source_operation_citation_invalid", pointer)
		return diagnostics
	}
	// Inspect retained occurrences independently of effectiveSourceParameters
	// and the claimant's diagnostics/projection. Legal inter-scope precedence
	// cannot make duplicate declarations within either source scope valid.
	checkParameters := func(raw json.RawMessage, sourcePointer string) {
		var parameters []json.RawMessage
		if json.Unmarshal(raw, &parameters) != nil {
			return
		}
		facts := sourceFacts{Document: documents[documentID].Payload, referenceRoot: roots[documentID]}
		if documentID == row.Source.DocumentID {
			facts.RefPrefix = "/source_contract"
		}
		seen := map[[2]string]bool{}
		for i, rawParameter := range parameters {
			parameter, ok := sourceResolveObject(facts, rawParameter, map[string]bool{}, 0)
			if !ok {
				continue
			}
			var key [2]string
			if json.Unmarshal(parameter["in"], &key[0]) != nil || json.Unmarshal(parameter["name"], &key[1]) != nil {
				continue
			}
			if seen[key] {
				add("source_parameter_duplicate", fmt.Sprintf("%s/%d", sourcePointer, i))
			}
			seen[key] = true
		}
	}
	checkParameters(operation["parameters"], pointer+"/parameters")
	for _, field := range []struct{ name, key string }{
		{"parameters", "parameters"}, {"request_body", "requestBody"}, {"responses", "responses"},
		{"summary", "summary"}, {"description", "description"}, {"callbacks", "callbacks"},
		{"deprecated", "deprecated"}, {"external_docs", "externalDocs"},
	} {
		require(field.name, operation[field.key], documentID, pointer+"/"+field.key)
	}
	// Require shared facts from their retained authority even when a claimant
	// removed both the copied group and its citation. Operation security,
	// including an explicit empty array, takes precedence over root security.
	requireAt := func(name, sourceDocument, sourcePointer string) {
		value, err := sourceLaneRetainedPointer(roots[sourceDocument], sourcePointer)
		if err == nil {
			require(name, value, sourceDocument, sourcePointer)
		}
	}
	contractDocument, contractPointer := row.Source.DocumentID, "/source_contract"
	if documentID != row.Source.DocumentID {
		contractDocument, contractPointer = documentID, ""
		if parent, _, ok := strings.Cut(pointer, "/"+strings.ToLower(row.Facts.Method)); ok {
			requireAt("path_parameters", documentID, parent+"/parameters")
			if raw, err := sourceLaneRetainedPointer(roots[documentID], parent+"/parameters"); err == nil {
				checkParameters(raw, parent+"/parameters")
			}
		}
	}
	if security, present := operation["security"]; present {
		require("security", security, documentID, pointer+"/security")
	} else {
		requireAt("security", contractDocument, contractPointer+"/security")
	}
	requireAt("security_schemes", contractDocument, contractPointer+"/components/securitySchemes")
	requireAt("webhooks", contractDocument, contractPointer+"/webhooks")
	for _, name := range []string{"path_bridge", "event_schema_inventory", "batch_action_inventory"} {
		requireAt(name, row.Source.DocumentID, "/rest/"+name)
	}
	return diagnostics
}
