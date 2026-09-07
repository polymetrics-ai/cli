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
	UnknownPointer  string
	Envelope        bool
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
		if !sourceReferenceAnnotationSiblings(node) {
			return nil, pointer, lineage, false
		}
		var ref string
		if json.Unmarshal(edge, &ref) != nil || !strings.HasPrefix(ref, "#/") {
			return nil, pointer, lineage, false
		}
		pointer = facts.RefPrefix + ref[1:]
	}
	return nil, pointer, lineage, false
}

// Structural siblings of a reference require conjunction reasoning. These
// descriptive siblings do not change its shape and may retain source context.
func sourceReferenceAnnotationSiblings(node map[string]json.RawMessage) bool {
	for name := range node {
		switch name {
		case "$ref", "description", "title", "deprecated", "example", "examples":
		default:
			return false
		}
	}
	return true
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
				scopes = append(scopes, sourceCollectionScope{ResponsePointer: actual, UnknownPointer: actual + "/content/" + escapeSourcePointer(name)})
				continue
			}
			one, _ := json.Marshal(map[string]json.RawMessage{name: content[name]})
			shape := sourceContentShape(facts, one)
			scope := sourceCollectionScope{ResponsePointer: actual, UnknownPointer: actual + "/content/" + escapeSourcePointer(name), Envelope: shape.Envelope, Cardinality: shape.Cardinality}
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

// sourceLocalShape holds only decoded facts from one already-resolved node.
// It owns no traversal, source read, cache or analysis budget. Consumers retain
// their separate cardinality, diagnostic and checked-lineage obligations.
type sourceLocalTypeState uint8

const (
	sourceLocalTypeAbsent sourceLocalTypeState = iota
	sourceLocalTypeSupported
	sourceLocalTypeUnsupported
)

type sourceLocalMemberState uint8

const (
	sourceLocalMemberAbsent sourceLocalMemberState = iota
	sourceLocalMemberValid
	sourceLocalMemberMalformed
)

type sourceLocalPrefixState uint8

const (
	sourceLocalPrefixAbsent sourceLocalPrefixState = iota
	sourceLocalPrefixEmpty
	sourceLocalPrefixNonempty
	sourceLocalPrefixMalformed
)

type sourceLocalObjectState uint8

const (
	sourceLocalObjectUnverified sourceLocalObjectState = iota
	sourceLocalObjectEstablished
	sourceLocalKnownNonobject
)

type sourceLocalArrayCoverage uint8

const (
	sourceLocalArrayUnverified sourceLocalArrayCoverage = iota
	sourceLocalArrayUniform
	sourceLocalArrayPrefixOrTail
)

type sourceLocalShape struct {
	Type            string
	TypeState       sourceLocalTypeState
	Properties      map[string]json.RawMessage
	PropertiesState sourceLocalMemberState
	Prefix          []json.RawMessage
	PrefixState     sourceLocalPrefixState
	HasItems        bool
	HasComposition  bool
}

func sourceLocalShapeEvidence(node map[string]json.RawMessage) sourceLocalShape {
	var shape sourceLocalShape
	if raw, present := node["type"]; present {
		shape.TypeState = sourceLocalTypeUnsupported
		var typ string
		if json.Unmarshal(raw, &typ) == nil {
			switch typ {
			case "object", "array", "string", "number", "integer", "boolean", "null":
				shape.Type, shape.TypeState = typ, sourceLocalTypeSupported
			}
		}
	}
	if raw, present := node["properties"]; present {
		shape.PropertiesState = sourceLocalMemberMalformed
		if json.Unmarshal(raw, &shape.Properties) == nil && shape.Properties != nil {
			shape.PropertiesState = sourceLocalMemberValid
		}
	}
	if raw, present := node["prefixItems"]; present {
		shape.PrefixState = sourceLocalPrefixMalformed
		if json.Unmarshal(raw, &shape.Prefix) == nil && shape.Prefix != nil {
			shape.PrefixState = sourceLocalPrefixEmpty
			if len(shape.Prefix) > 0 {
				shape.PrefixState = sourceLocalPrefixNonempty
			}
		}
	}
	_, shape.HasItems = node["items"]
	for _, name := range []string{"allOf", "oneOf", "anyOf"} {
		if _, present := node[name]; present {
			shape.HasComposition = true
		}
	}
	return shape
}

func (s sourceLocalShape) object() sourceLocalObjectState {
	if s.HasComposition || s.TypeState == sourceLocalTypeUnsupported || s.PropertiesState == sourceLocalMemberMalformed {
		return sourceLocalObjectUnverified
	}
	if s.TypeState == sourceLocalTypeSupported {
		if s.Type == "object" {
			return sourceLocalObjectEstablished
		}
		return sourceLocalKnownNonobject
	}
	if s.HasItems || s.PrefixState != sourceLocalPrefixAbsent {
		return sourceLocalObjectUnverified
	}
	if s.PropertiesState == sourceLocalMemberValid {
		return sourceLocalObjectEstablished
	}
	return sourceLocalObjectUnverified
}

func (s sourceLocalShape) arrayCoverage() sourceLocalArrayCoverage {
	if s.HasComposition || s.TypeState != sourceLocalTypeSupported || s.Type != "array" || s.PropertiesState == sourceLocalMemberMalformed {
		return sourceLocalArrayUnverified
	}
	switch s.PrefixState {
	case sourceLocalPrefixAbsent, sourceLocalPrefixEmpty:
		return sourceLocalArrayUniform
	case sourceLocalPrefixNonempty:
		return sourceLocalArrayPrefixOrTail
	default:
		return sourceLocalArrayUnverified
	}
}

func sourceCollectionObjectCode(node map[string]json.RawMessage) string {
	switch sourceLocalShapeEvidence(node).object() {
	case sourceLocalObjectEstablished:
		return ""
	case sourceLocalKnownNonobject:
		return "source_collection_interpretation_invalid"
	default:
		return "source_collection_interpretation_unresolved"
	}
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
	if len(node["allOf"]) > 0 || len(node["oneOf"]) > 0 || len(node["anyOf"]) > 0 {
		return sourceCollectionUnknown, nil, unknown
	}
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
			if code := sourceCollectionObjectCode(node); code != "" {
				return sourceCollectionUnknown, nil, code
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
		shape := sourceLocalShapeEvidence(node)
		if shape.HasComposition || shape.TypeState == sourceLocalTypeUnsupported || shape.PropertiesState == sourceLocalMemberMalformed {
			return sourceCollectionUnknown, nil, unknown
		}
		if shape.Type != "array" {
			if shape.object() == sourceLocalObjectEstablished || shape.object() == sourceLocalKnownNonobject {
				return sourceCollectionUnknown, nil, invalid
			}
			return sourceCollectionUnknown, nil, unknown
		}
		if shape.arrayCoverage() != sourceLocalArrayUniform {
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
		if code := sourceCollectionObjectCode(item); code != "" {
			return sourceCollectionUnknown, nil, code
		}
		if ref, ok := sourceCollectionCitationAt(facts, scope.Ref.DocumentID, itemPointer); ok {
			refs = append(refs, ref)
		}
	} else {
		if code := sourceCollectionObjectCode(node); code != "" {
			return sourceCollectionUnknown, nil, code
		}
	}
	owned := a.Citation == facts.Refs["summary"] || a.Citation == facts.Refs["description"]
	owned = owned || (a.Citation.DocumentID == scope.Ref.DocumentID && a.Citation.Pointer == scope.ResponsePointer+"/description")
	if a.Citation.DocumentID == scope.Ref.DocumentID && strings.HasSuffix(a.Citation.Pointer, "/description") {
		for _, base := range lineage {
			owned = owned || a.Citation.Pointer == base+"/description"
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
		if code == "source_collection_interpretation_unresolved" || code == "source_collection_interpretation_missing" || code == "source_collection_scope_unknown" {
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
	// A demonstrated collection does not erase an unresolved successful sibling.
	// These deficits describe individual response occurrences, not aggregate
	// applicability, and never grant a materialized binding or runtime proof.
	if semantics == "read" {
		for _, scope := range scopes {
			if scope.Cardinality != sourceCollectionUnknown {
				continue
			}
			pointer := scope.Ref.Pointer
			if pointer == "" {
				pointer = scope.UnknownPointer
			}
			if pointer == "" {
				pointer = scope.ResponsePointer
			}
			if pointer == "" {
				pointer = facts.Refs["responses"].Pointer
			}
			code := "source_collection_scope_unknown"
			if scope.Envelope && roles[scope.Ref] == "" {
				code = "source_collection_interpretation_missing"
			}
			add(code, pointer)
		}
	}
	sort.Slice(refs, func(i, j int) bool {
		left, _ := json.Marshal(refs[i])
		right, _ := json.Marshal(refs[j])
		return string(left) < string(right)
	})
	unique := refs[:0]
	for _, ref := range refs {
		if len(unique) == 0 || unique[len(unique)-1] != ref {
			unique = append(unique, ref)
		}
	}
	refs = unique
	sort.Slice(issues, func(i, j int) bool {
		left, _ := json.Marshal(issues[i])
		right, _ := json.Marshal(issues[j])
		return string(left) < string(right)
	})
	return result, refs, issues
}

// Recheck emitted interpretation evidence against retained documents and the
// captured authoring input. Agreement between two produced cells is no oracle.
func validateSourceLaneInterpretationEvidence(candidate sourceLaneManifest, annotationInputs json.RawMessage) []sourceLaneDiagnostic {
	var annotations []sourceSemanticAnnotation
	_ = json.Unmarshal(annotationInputs, &annotations)
	documents := map[string]retainedSourceDocument{}
	for _, doc := range candidate.Documents {
		documents[doc.ID] = doc
	}
	issues := []sourceLaneDiagnostic{}
	add := func(key sourceOperationKey, code, pointer string) {
		issues = append(issues, sourceLaneDiagnostic{Key: key, Lanes: []string{"etl"}, Stage: "manifest", Code: code, Pointer: pointer, Owner: key.Connector, Severity: "error"})
	}
	for _, row := range candidate.SourceOperations {
		for _, cell := range row.Lanes {
			if cell.Lane != "etl" || cell.RuleID != "source_response_interpretation" {
				continue
			}
			doc, exists := documents[row.Source.DocumentID]
			if !exists {
				add(row.Source.Key, "source_collection_authority_missing", row.Source.Pointer)
				continue
			}
			source := row.Source
			var err error
			source.Node, err = sourceJSONPointer(doc.Payload, source.Pointer)
			if err != nil {
				add(source.Key, "source_collection_authority_missing", source.Pointer)
				continue
			}
			var raw *retainedSourceDocument
			if d, exists := documents[source.RawDocumentID]; exists {
				raw = &d
			}
			facts := normalizeSourceFacts(source, doc, raw)
			facts.analysis = &sourceShapeAnalysis{Root: facts.referenceRoot, Objects: map[string]map[string]json.RawMessage{}, Shapes: map[string]sourceShape{}}
			allowed := map[sourceFactRef]bool{}
			for _, ref := range facts.Refs {
				allowed[ref] = true
			}
			scopes := sourceCollectionScopes(facts)
			authorized := false
			for _, annotation := range annotations {
				if annotation.Key != source.Key {
					continue
				}
				semantics := sourceOperationSemantics(facts)
				if annotation.Semantics != "" {
					semantics = annotation.Semantics
				}
				for _, interpretation := range annotation.ResponseInterpretations {
					for _, scope := range scopes {
						if scope.Ref != interpretation.ResponseSchema {
							continue
						}
						kind, required, code := validateSourceResponseInterpretation(facts, semantics, scope, interpretation)
						if code != "" {
							continue
						}
						if (cell.Applicability == "applicable" && kind == sourceCollection) || (cell.Applicability == "not_applicable" && kind == sourceNoncollection) {
							authorized = true
						}
						for _, ref := range required {
							allowed[ref] = true
							found := false
							for _, emitted := range cell.FactRefs {
								found = found || emitted == ref
							}
							if !found {
								add(source.Key, "source_collection_citation_missing", ref.Pointer)
							}
						}
					}
				}
			}
			if !authorized {
				add(source.Key, "source_collection_authority_missing", source.Pointer)
			}
			for _, ref := range cell.FactRefs {
				original, exists := documents[ref.DocumentID]
				value, err := sourceJSONPointer(original.Payload, ref.Pointer)
				canonical, canonicalErr := canonicalSourceJSON(value)
				if !exists || err != nil || canonicalErr != nil || sourceBytesHash(canonical) != ref.ValueSHA256 || !allowed[ref] {
					add(source.Key, "source_collection_citation_invalid", ref.Pointer)
				}
			}
		}
	}
	return issues
}
