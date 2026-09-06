package main

import (
	"encoding/json"
	"sort"
	"strings"
)

type sourceCollectionKind uint8

const (
	sourceCollectionUnknown sourceCollectionKind = iota
	sourceNoncollection
	sourceCollection
)

// Distinct successful status/media scopes can independently establish source
// applicability; unknown siblings still cannot supply complete target proof.
func mergeSourceCollection(a, b sourceCollectionKind) sourceCollectionKind {
	if a == sourceCollection || b == sourceCollection {
		return sourceCollection
	}
	if a == sourceCollectionUnknown || b == sourceCollectionUnknown {
		return sourceCollectionUnknown
	}
	return sourceNoncollection
}

type sourceResponseInterpretation struct {
	Kind           string        `json:"kind"`
	ResponseSchema sourceFactRef `json:"response_schema"`
	RecordsPointer *string       `json:"records_pointer,omitempty"`
	Citation       sourceFactRef `json:"citation"`
	Clause         string        `json:"clause"`
	recordsPresent bool
}

func (r *sourceResponseInterpretation) UnmarshalJSON(raw []byte) error {
	type plain sourceResponseInterpretation
	var decoded plain
	if err := decodeStrictJSON(raw, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	_, decoded.recordsPresent = fields["records_pointer"]
	*r = sourceResponseInterpretation(decoded)
	return nil
}

type sourceCollectionScope struct {
	Ref             sourceFactRef
	ResponsePointer string
	Cardinality     sourceCollectionKind
}

// Resolve only a bounded local lineage. Non-annotation ref siblings require
// schema conjunction reasoning outside this narrow interpretation contract.
func sourceCollectionObjectAt(facts sourceFacts, pointer string) (map[string]json.RawMessage, string, []string, bool) {
	seen := map[string]bool{}
	lineage := []string{}
	for depth := 0; depth <= 128; depth++ {
		if seen[pointer] {
			return nil, pointer, lineage, false
		}
		seen[pointer] = true
		lineage = append(lineage, pointer)
		if facts.analysis != nil {
			facts.analysis.Visits++
			if facts.analysis.Visits > 100000 {
				facts.analysis.Exhausted = true
				return nil, pointer, lineage, false
			}
		}
		raw, err := sourceAnalysisPointer(facts, pointer)
		if err != nil {
			return nil, pointer, lineage, false
		}
		var node map[string]json.RawMessage
		if json.Unmarshal(raw, &node) != nil || node == nil {
			return nil, pointer, lineage, false
		}
		edge, exists := node["$ref"]
		if !exists {
			return node, pointer, lineage, true
		}
		for name := range node {
			switch name {
			case "$ref", "description", "title", "deprecated", "example", "examples":
			default:
				return nil, pointer, lineage, false
			}
		}
		var ref string
		if json.Unmarshal(edge, &ref) != nil || !strings.HasPrefix(ref, "#/") {
			return nil, pointer, lineage, false
		}
		pointer = facts.RefPrefix + ref[1:]
	}
	return nil, pointer, lineage, false
}

func sourceCollectionScopes(facts sourceFacts) []sourceCollectionScope {
	owner := facts.Refs["responses"]
	var responses map[string]json.RawMessage
	if json.Unmarshal(facts.Groups["responses"], &responses) != nil {
		return []sourceCollectionScope{{Cardinality: sourceCollectionUnknown}}
	}
	statuses := make([]string, 0, len(responses))
	for status := range responses {
		statuses = append(statuses, status)
	}
	sort.Strings(statuses)
	scopes := []sourceCollectionScope{}
	for _, status := range statuses {
		if len(status) != 3 || status[0] != '2' {
			continue
		}
		responsePointer := owner.Pointer + "/" + escapeSourcePointer(status)
		response, actual, _, ok := sourceCollectionObjectAt(facts, responsePointer)
		if !ok {
			scopes = append(scopes, sourceCollectionScope{ResponsePointer: responsePointer})
			continue
		}
		var content map[string]json.RawMessage
		if json.Unmarshal(response["content"], &content) != nil || len(content) == 0 {
			kind := sourceCollectionUnknown
			if (status == "204" || status == "205") && len(response["content"]) == 0 {
				kind = sourceNoncollection
			}
			scopes = append(scopes, sourceCollectionScope{ResponsePointer: actual, Cardinality: kind})
			continue
		}
		media := make([]string, 0, len(content))
		for name := range content {
			media = append(media, name)
		}
		sort.Strings(media)
		for _, name := range media {
			var entry map[string]json.RawMessage
			if json.Unmarshal(content[name], &entry) != nil {
				scopes = append(scopes, sourceCollectionScope{ResponsePointer: actual})
				continue
			}
			one, _ := json.Marshal(map[string]json.RawMessage{name: content[name]})
			shape := sourceContentShape(facts, one)
			scope := sourceCollectionScope{ResponsePointer: actual, Cardinality: shape.Cardinality}
			if schema := entry["schema"]; len(schema) > 0 {
				canonical, err := canonicalSourceJSON(schema)
				if err == nil {
					scope.Ref = sourceFactRef{DocumentID: owner.DocumentID, Pointer: actual + "/content/" + escapeSourcePointer(name) + "/schema", ValueSHA256: sourceBytesHash(canonical)}
				}
			}
			scopes = append(scopes, scope)
		}
	}
	return scopes
}

func sourceCollectionCitationAt(facts sourceFacts, document, pointer string) (sourceFactRef, bool) {
	raw, err := sourceAnalysisPointer(facts, pointer)
	if err != nil {
		return sourceFactRef{}, false
	}
	canonical, err := canonicalSourceJSON(raw)
	if err != nil {
		return sourceFactRef{}, false
	}
	return sourceFactRef{DocumentID: document, Pointer: pointer, ValueSHA256: sourceBytesHash(canonical)}, true
}

func validateSourceResponseInterpretation(facts sourceFacts, semantics string, scope sourceCollectionScope, a sourceResponseInterpretation) (sourceCollectionKind, []sourceFactRef, string) {
	invalid := "source_collection_interpretation_invalid"
	unknown := "source_collection_interpretation_unresolved"
	if a.Kind != "collection" && a.Kind != "single_resource" {
		return sourceCollectionUnknown, nil, invalid
	}
	if a.ResponseSchema != scope.Ref || a.ResponseSchema.Pointer == "" {
		return sourceCollectionUnknown, nil, invalid
	}
	if a.Kind == "collection" && (a.RecordsPointer == nil || semantics == "mutation") {
		return sourceCollectionUnknown, nil, invalid
	}
	if a.Kind == "single_resource" && (a.RecordsPointer != nil || a.recordsPresent) {
		return sourceCollectionUnknown, nil, invalid
	}
	support, code := sourceLaneCitedValue(facts, a.Citation)
	if code != "" {
		return sourceCollectionUnknown, nil, invalid
	}
	var text string
	if json.Unmarshal(support, &text) != nil || strings.TrimSpace(a.Clause) == "" || !strings.Contains(text, a.Clause) {
		return sourceCollectionUnknown, nil, invalid
	}
	node, pointer, lineage, ok := sourceCollectionObjectAt(facts, scope.Ref.Pointer)
	if !ok {
		return sourceCollectionUnknown, nil, unknown
	}
	refs := []sourceFactRef{scope.Ref, a.Citation}
	if a.Kind == "collection" {
		if !sourceLaneProjectionPointer(*a.RecordsPointer) {
			return sourceCollectionUnknown, nil, invalid
		}
		parts := []string{}
		if *a.RecordsPointer != "" {
			parts = strings.Split((*a.RecordsPointer)[1:], "/")
		}
		for _, part := range parts {
			name := strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
			if name == "*" || name == "[]" {
				return sourceCollectionUnknown, nil, invalid
			}
			var props map[string]json.RawMessage
			if json.Unmarshal(node["properties"], &props) != nil || props == nil {
				return sourceCollectionUnknown, nil, invalid
			}
			if _, exists := props[name]; !exists {
				return sourceCollectionUnknown, nil, invalid
			}
			pointer += "/properties/" + escapeSourcePointer(name)
			if ref, ok := sourceCollectionCitationAt(facts, scope.Ref.DocumentID, pointer); ok {
				refs = append(refs, ref)
			}
			var next []string
			node, pointer, next, ok = sourceCollectionObjectAt(facts, pointer)
			lineage = append(lineage, next...)
			if !ok {
				return sourceCollectionUnknown, nil, unknown
			}
		}
		var typ string
		_ = json.Unmarshal(node["type"], &typ)
		if typ != "array" {
			return sourceCollectionUnknown, nil, invalid
		}
		if len(node["allOf"]) > 0 || len(node["oneOf"]) > 0 || len(node["anyOf"]) > 0 || len(node["prefixItems"]) > 0 {
			return sourceCollectionUnknown, nil, unknown
		}
		if ref, ok := sourceCollectionCitationAt(facts, scope.Ref.DocumentID, pointer); ok {
			refs = append(refs, ref)
		}
		item, itemPointer, itemLineage, ok := sourceCollectionObjectAt(facts, pointer+"/items")
		lineage = append(lineage, itemLineage...)
		if !ok {
			return sourceCollectionUnknown, nil, unknown
		}
		var itemType string
		_ = json.Unmarshal(item["type"], &itemType)
		if itemType != "object" && len(item["properties"]) == 0 {
			return sourceCollectionUnknown, nil, invalid
		}
		if len(item["allOf"]) > 0 || len(item["oneOf"]) > 0 || len(item["anyOf"]) > 0 {
			return sourceCollectionUnknown, nil, unknown
		}
		if ref, ok := sourceCollectionCitationAt(facts, scope.Ref.DocumentID, itemPointer); ok {
			refs = append(refs, ref)
		}
	} else {
		var typ string
		_ = json.Unmarshal(node["type"], &typ)
		if typ != "object" && len(node["properties"]) == 0 {
			return sourceCollectionUnknown, nil, invalid
		}
	}
	owned := a.Citation == facts.Refs["summary"] || a.Citation == facts.Refs["description"]
	owned = owned || (a.Citation.DocumentID == scope.Ref.DocumentID && a.Citation.Pointer == scope.ResponsePointer+"/description")
	if a.Citation.DocumentID == scope.Ref.DocumentID && strings.HasSuffix(a.Citation.Pointer, "/description") {
		for _, base := range lineage {
			owned = owned || strings.HasPrefix(a.Citation.Pointer, base+"/")
		}
	}
	if !owned {
		return sourceCollectionUnknown, nil, invalid
	}
	if a.Kind == "single_resource" {
		return sourceNoncollection, refs, ""
	}
	return sourceCollection, refs, ""
}

func applySourceResponseInterpretations(key sourceOperationKey, facts sourceFacts, semantics string, interpretations []sourceResponseInterpretation) (sourceCollectionKind, []sourceFactRef, []sourceLaneDiagnostic) {
	scopes := sourceCollectionScopes(facts)
	issues := []sourceLaneDiagnostic{}
	refs := []sourceFactRef{}
	add := func(code, pointer string) {
		severity := "error"
		if code == "source_collection_interpretation_unresolved" {
			severity = "deficit"
		}
		issues = append(issues, sourceLaneDiagnostic{Key: key, Lanes: []string{"etl"}, Stage: "classification", Code: code, Pointer: pointer, Owner: key.Connector, Severity: severity})
	}
	seen := map[string]bool{}
	roles := map[sourceFactRef]string{}
	for _, a := range interpretations {
		raw, _ := json.Marshal(a)
		identity := string(raw)
		if seen[identity] || (roles[a.ResponseSchema] != "" && roles[a.ResponseSchema] != a.Kind) {
			add("source_collection_interpretation_invalid", a.ResponseSchema.Pointer)
			continue
		}
		seen[identity] = true
		roles[a.ResponseSchema] = a.Kind
		matched := -1
		for i, scope := range scopes {
			if scope.Ref == a.ResponseSchema {
				if matched >= 0 {
					matched = -2
					break
				}
				matched = i
			}
		}
		if matched < 0 {
			add("source_collection_interpretation_invalid", a.ResponseSchema.Pointer)
			continue
		}
		kind, citations, code := validateSourceResponseInterpretation(facts, semantics, scopes[matched], a)
		if code != "" {
			add(code, a.ResponseSchema.Pointer)
			scopes[matched].Cardinality = sourceCollectionUnknown
			continue
		}
		scopes[matched].Cardinality = kind
		refs = append(refs, citations...)
	}
	result := sourceNoncollection
	if len(scopes) == 0 {
		result = sourceCollectionUnknown
	}
	for _, scope := range scopes {
		result = mergeSourceCollection(result, scope.Cardinality)
	}
	if len(issues) > 0 {
		result = sourceCollectionUnknown
	}
	return result, refs, issues
}
