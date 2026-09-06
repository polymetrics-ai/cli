package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type sourceLaneReason struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

type sourceLaneTargetRef struct {
	Kind             string `json:"kind"`
	Connector        string `json:"connector"`
	ID               string `json:"id"`
	Lane             string `json:"lane"`
	Artifact         string `json:"artifact"`
	Pointer          string `json:"pointer"`
	ArtifactSHA256   string `json:"artifact_sha256"`
	CanonicalID      string `json:"canonical_id"`
	CanonicalPointer string `json:"canonical_pointer"`
	Generation       string `json:"generation"`
	SchemaRole       string `json:"schema_role"`
}

type sourceLaneCell struct {
	Lane             string                 `json:"lane"`
	Applicability    string                 `json:"applicability"`
	State            string                 `json:"state"`
	RuleID           string                 `json:"rule_id"`
	FactRefs         []sourceFactRef        `json:"fact_refs"`
	Reason           sourceLaneReason       `json:"reason"`
	IntendedBindings []sourceLaneTargetRef  `json:"intended_bindings"`
	References       []sourceLaneTargetRef  `json:"references"`
	ProofRefs        []string               `json:"proof_refs"`
	OwnerRefs        []string               `json:"owner_refs"`
	GapRefs          []string               `json:"gap_refs"`
	Diagnostics      []sourceLaneDiagnostic `json:"diagnostics"`
}

type sourceSemanticAnnotation struct {
	Key                  sourceOperationKey    `json:"key"`
	Semantics            string                `json:"semantics"`
	Citation             sourceFactRef         `json:"citation"`
	Clause               string                `json:"clause"`
	FoundationGap        string                `json:"foundation_gap"`
	AtlasID              string                `json:"atlas_id"`
	DecisionRefs         []string              `json:"decision_refs"`
	IntendedBindings     []sourceLaneTargetRef `json:"intended_bindings"`
	MaterializedBindings []sourceLaneTargetRef `json:"materialized_bindings"`
}

// classifySourceLanes treats source applicability separately from artifact and
// proof availability. Unresolved semantics retain seven explicit cells.
func classifySourceLanes(key sourceOperationKey, facts sourceFacts, annotation *sourceSemanticAnnotation) []sourceLaneCell {
	facts.analysis = &sourceShapeAnalysis{Root: facts.referenceRoot, Objects: map[string]map[string]json.RawMessage{}, Shapes: map[string]sourceShape{}}
	cells := make([]sourceLaneCell, 0, 7)
	for _, lane := range sourceLaneNames() {
		cells = append(cells, sourceLaneCell{Lane: lane, Applicability: "undetermined", State: "mapped_unproven", RuleID: "facts_unresolved", FactRefs: []sourceFactRef{}, Reason: sourceLaneReason{Code: "facts_unresolved", Text: "Retained facts do not yet establish this lane."}, IntendedBindings: []sourceLaneTargetRef{}, References: []sourceLaneTargetRef{}, ProofRefs: []string{}, OwnerRefs: []string{key.Connector}, GapRefs: []string{}, Diagnostics: []sourceLaneDiagnostic{}})
	}
	if facts.Status == "unavailable" {
		return sourceLaneUnknownDiagnostics(key, cells)
	}
	semantics := sourceOperationSemantics(facts)
	annotationValid := true
	if annotation != nil {
		if err := validateSourceAnnotation(key, facts, *annotation); err != nil {
			annotationValid = false
		} else if annotation.Semantics != "" {
			semantics = annotation.Semantics
		}
	}
	set := func(i int, applicable bool, rule string, groups ...string) {
		cell := &cells[i]
		cell.RuleID = rule
		cell.Applicability = "applicable"
		cell.Reason = sourceLaneReason{Code: "materialization_unproven", Text: "The source establishes this lane; materialization and lane-specific behavior remain unproven."}
		if !applicable {
			cell.Applicability = "not_applicable"
			cell.State = "not_applicable"
			cell.Reason = sourceLaneReason{Code: "source_exclusion", Text: "The retained source contract excludes this lane under " + rule + "."}
		}
		for _, group := range groups {
			if ref, ok := facts.Refs[group]; ok {
				cell.FactRefs = append(cell.FactRefs, ref)
			}
		}
		if annotationValid && annotation != nil && annotation.Semantics != "" {
			cell.FactRefs = append(cell.FactRefs, annotation.Citation)
		}
	}
	switch semantics {
	case "read":
		set(0, true, "source_read", "method", "summary", "responses")
		set(1, false, "source_read", "method", "summary")
		set(5, false, "source_read", "method", "summary")
	case "mutation":
		set(0, false, "source_mutation", "method", "summary")
		set(1, true, "source_mutation", "method", "summary", "request_body")
		set(5, true, "source_destination_mutation", "method", "summary", "request_body")
	}
	response := sourceResponseShape(facts)
	if response.Binary {
		set(2, true, "successful_binary_response", "responses")
	} else if response.Known {
		set(2, false, "known_nonbinary_response", "responses")
	}
	request := sourceRequestShape(facts)
	if request.Binary {
		set(3, true, "binary_request", "request_body")
	} else if request.Known {
		set(3, false, "known_nonbinary_request", "request_body", "source_operation")
	}
	if semantics == "mutation" {
		set(4, false, "source_mutation", "method", "summary")
	} else if semantics == "read" {
		if response.Collection {
			set(4, true, "source_record_collection", "responses")
		} else if response.Known {
			set(4, false, "fixed_noncollection_read", "responses")
		}
	}
	if raw := facts.Groups["callbacks"]; len(raw) > 0 && string(raw) != "null" && string(raw) != "{}" {
		set(6, true, "source_callback_contract", "callbacks")
	} else if _, complete := facts.Refs["source_operation"]; semantics != "" && complete {
		set(6, false, "no_operation_event_contract", "source_operation")
	}
	if annotationValid && annotation != nil {
		if annotation.FoundationGap != "" {
			set(6, true, "existing_receiver_demand", "summary", "request_body")
			cells[6].State = "missing_foundation"
			cells[6].GapRefs = []string{annotation.FoundationGap}
			cells[6].OwnerRefs = append([]string{annotation.AtlasID}, annotation.DecisionRefs...)
			cells[6].Reason = sourceLaneReason{Code: "existing_foundation_demand", Text: "Source-backed registration retains an existing receiver demand; this does not implement or approve a receiver."}
		}
		for _, ref := range annotation.IntendedBindings {
			for i := range cells {
				if cells[i].Lane == ref.Lane {
					cells[i].IntendedBindings = append(cells[i].IntendedBindings, ref)
				}
			}
		}
	}
	if annotationValid && annotation != nil {
		cells = resolveSourceLaneBindings(key, facts, *annotation, cells)
	}
	if !annotationValid {
		for i := range cells {
			cells[i].Diagnostics = append(cells[i].Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{cells[i].Lane}, Stage: "classification", Code: "source_annotation_invalid", Pointer: annotation.Citation.Pointer, Owner: key.Connector, Severity: "error"})
		}
	}
	return sourceLaneUnknownDiagnostics(key, cells)
}

func sourceLaneUnknownDiagnostics(key sourceOperationKey, cells []sourceLaneCell) []sourceLaneCell {
	for i := range cells {
		if cells[i].Applicability != "undetermined" {
			continue
		}
		code := "source_semantics_unresolved"
		switch cells[i].Lane {
		case "binary_download":
			code = "source_response_shape_unresolved"
		case "binary_upload":
			code = "source_request_shape_unresolved"
		case "etl":
			code = "source_record_shape_unresolved"
		case "sync_transport":
			code = "source_event_contract_unresolved"
		}
		cells[i].Diagnostics = append(cells[i].Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{cells[i].Lane}, Stage: "classification", Code: code, Pointer: "", Owner: key.Connector, Severity: "deficit"})
	}
	return cells
}

func sourceOperationSemantics(facts sourceFacts) string {
	if facts.Method == "GET" || facts.Method == "HEAD" {
		return "read"
	}
	summary := sourceFactText(facts, "summary")
	if facts.Method != "POST" && facts.Method != "PUT" && facts.Method != "PATCH" && facts.Method != "DELETE" {
		return ""
	}
	if sourceActionSemantics(summary) == "mutation" {
		return "mutation"
	}
	return ""
}

// sourceActionSemantics recognizes affirmative leading actions only. A word
// inside an object name (such as playlist or read token) is not an action.
// Unsupported prose remains unresolved and requires a supported source clause.
func sourceActionSemantics(clause string) string {
	summary := strings.ToLower(strings.TrimSpace(clause))
	// These are cited source action words, not operation-ID/name heuristics.
	words := []string{"create", "creates", "add", "adds", "update", "updates", "delete", "deletes", "remove", "removes", "approve", "approves", "reject", "rejects", "cancel", "cancels", "set", "sets", "replace", "replaces", "upload", "uploads", "publish", "publishes", "start", "starts", "stop", "stops", "trigger", "triggers", "revoke", "revokes", "restore", "restores", "archive", "archives", "assign", "assigns", "invite", "invites", "enable", "enables", "disable", "disables", "attach", "attaches", "detach", "detaches"}
	for _, word := range words {
		if summary == word || strings.HasPrefix(summary, word+" ") {
			return "mutation"
		}
	}
	for _, word := range []string{"read", "reads", "query", "queries", "retrieve", "retrieves", "search", "searches", "list", "lists", "get", "gets"} {
		if summary == word || strings.HasPrefix(summary, word+" ") {
			return "read"
		}
	}
	return ""
}

func sourceFactText(facts sourceFacts, name string) string {
	var value string
	if err := json.Unmarshal(facts.Groups[name], &value); err != nil {
		return ""
	}
	return value
}

func validateSourceAnnotation(key sourceOperationKey, facts sourceFacts, a sourceSemanticAnnotation) error {
	if a.Key != key || (a.Semantics != "" && a.Semantics != "read" && a.Semantics != "mutation") || strings.TrimSpace(a.Clause) == "" {
		return fmt.Errorf("invalid semantic annotation")
	}
	var cited json.RawMessage
	for name, ref := range facts.Refs {
		if ref == a.Citation {
			cited = facts.Groups[name]
			break
		}
	}
	if len(cited) == 0 {
		return fmt.Errorf("annotation citation is not a retained fact")
	}
	canonical, err := canonicalSourceJSON(cited)
	if err != nil {
		return err
	}
	if sourceBytesHash(canonical) != a.Citation.ValueSHA256 {
		return fmt.Errorf("annotation citation changed")
	}
	var text string
	if err := json.Unmarshal(cited, &text); err != nil || !strings.Contains(text, a.Clause) {
		return fmt.Errorf("annotation clause absent")
	}
	words := strings.ToLower(a.Clause)
	if a.Semantics != "" && sourceActionSemantics(a.Clause) != a.Semantics {
		return fmt.Errorf("semantic interpretation lacks affirmative cited support")
	}
	if asserted := sourceActionSemantics(sourceFactText(facts, "summary")); a.Semantics != "" && asserted != "" && asserted != a.Semantics {
		return fmt.Errorf("semantic interpretation contradicts source action")
	}
	if a.Semantics == "mutation" && (facts.Method == "GET" || facts.Method == "HEAD") {
		return fmt.Errorf("mutation conflicts with safe read contract")
	}
	if a.FoundationGap != "" && (a.FoundationGap != "cli-webhook-event-surface-foundation-r1" || a.AtlasID != "transport.sync-contract.v1" || len(a.DecisionRefs) == 0 || !strings.Contains(words, "webhook")) {
		return fmt.Errorf("unrecognized or uncited foundation demand")
	}
	return nil
}

type sourceShape struct{ Known, Binary, Collection bool }

type sourceShapeAnalysis struct {
	Objects map[string]map[string]json.RawMessage
	Shapes  map[string]sourceShape
	Root    any
	Visits  int
}

func sourceAnalysisPointer(facts sourceFacts, pointer string) (json.RawMessage, error) {
	if facts.analysis == nil {
		return sourceJSONPointer(facts.Document, pointer)
	}
	if facts.analysis.Root == nil {
		if err := decodeSourceJSON(facts.Document, &facts.analysis.Root); err != nil {
			return nil, err
		}
	}
	value := facts.analysis.Root
	for _, part := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		part = strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")
		switch node := value.(type) {
		case map[string]any:
			var ok bool
			value, ok = node[part]
			if !ok {
				return nil, fmt.Errorf("source reference absent")
			}
		case []any:
			index, err := strconv.Atoi(part)
			if err != nil || index < 0 || index >= len(node) {
				return nil, fmt.Errorf("source reference index invalid")
			}
			value = node[index]
		default:
			return nil, fmt.Errorf("source reference traverses scalar")
		}
	}
	return json.Marshal(value)
}

func sourceRequestShape(facts sourceFacts) sourceShape {
	raw := facts.Groups["request_body"]
	if string(raw) == "null" {
		return sourceShape{Known: true}
	}
	body, ok := sourceResolveObject(facts, raw, map[string]bool{}, 0)
	if !ok {
		return sourceShape{}
	}
	return sourceContentShape(facts, body["content"])
}

func sourceResponseShape(facts sourceFacts) sourceShape {
	var responses map[string]json.RawMessage
	if err := json.Unmarshal(facts.Groups["responses"], &responses); err != nil || len(responses) == 0 {
		return sourceShape{}
	}
	result := sourceShape{Known: true}
	found := false
	keys := make([]string, 0, len(responses))
	for status := range responses {
		keys = append(keys, status)
	}
	sort.Strings(keys)
	for _, status := range keys {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		found = true
		response, ok := sourceResolveObject(facts, responses[status], map[string]bool{}, 0)
		if !ok {
			result.Known = false
			continue
		}
		if len(response["content"]) == 0 {
			if status != "204" && status != "205" {
				result.Known = false
			}
			continue
		}
		shape := sourceContentShape(facts, response["content"])
		result.Known = result.Known && shape.Known
		result.Binary = result.Binary || shape.Binary
		result.Collection = result.Collection || shape.Collection
	}
	result.Known = result.Known && found
	return result
}

func sourceContentShape(facts sourceFacts, raw json.RawMessage) sourceShape {
	var content map[string]json.RawMessage
	if err := json.Unmarshal(raw, &content); err != nil || len(content) == 0 {
		return sourceShape{}
	}
	result := sourceShape{Known: true}
	for media, rawEntry := range content {
		var entry map[string]json.RawMessage
		if err := json.Unmarshal(rawEntry, &entry); err != nil {
			result.Known = false
			continue
		}
		lower := strings.ToLower(media)
		concrete := strings.Split(lower, ";")[0]
		switch {
		case concrete == "application/octet-stream" || concrete == "application/pdf" || concrete == "application/zip" || concrete == "application/gzip" || strings.HasPrefix(concrete, "image/") || strings.HasPrefix(concrete, "audio/") || strings.HasPrefix(concrete, "video/"):
			if strings.Contains(concrete, "*") {
				result.Known = false
			} else {
				result.Binary = true
			}
		case strings.Contains(concrete, "*"):
			result.Known = false
		default:
			shape := sourceSchemaShape(facts, entry["schema"], map[string]bool{}, 0)
			result.Known = result.Known && shape.Known
			result.Binary = result.Binary || shape.Binary
			result.Collection = result.Collection || shape.Collection
		}
	}
	return result
}

func sourceResolveObject(facts sourceFacts, raw json.RawMessage, seen map[string]bool, depth int) (map[string]json.RawMessage, bool) {
	if depth > 256 {
		return nil, false
	}
	cacheKey := sourceBytesHash(raw)
	if facts.analysis != nil {
		facts.analysis.Visits++
		if facts.analysis.Visits > 100000 {
			return nil, false
		}
		if cached, exists := facts.analysis.Objects[cacheKey]; exists {
			return cached, true
		}
	}
	var node map[string]json.RawMessage
	if err := json.Unmarshal(raw, &node); err != nil || node == nil {
		return nil, false
	}
	var ref string
	if rawRef, exists := node["$ref"]; exists {
		if err := json.Unmarshal(rawRef, &ref); err != nil || !strings.HasPrefix(ref, "#/") || seen[ref] {
			return nil, false
		}
		seen[ref] = true
		value, err := sourceAnalysisPointer(facts, facts.RefPrefix+strings.TrimPrefix(ref, "#"))
		if err != nil {
			return nil, false
		}
		resolved, ok := sourceResolveObject(facts, value, seen, depth+1)
		if !ok {
			return nil, false
		}
		copy := map[string]json.RawMessage{}
		for k, v := range resolved {
			copy[k] = v
		}
		for k, v := range node {
			if k != "$ref" {
				copy[k] = v
			}
		}
		if facts.analysis != nil {
			facts.analysis.Objects[cacheKey] = copy
		}
		return copy, true
	}
	if facts.analysis != nil {
		facts.analysis.Objects[cacheKey] = node
	}
	return node, true
}

func sourceSchemaShape(facts sourceFacts, raw json.RawMessage, seen map[string]bool, depth int) sourceShape {
	if depth > 128 {
		return sourceShape{}
	}
	cacheKey := sourceBytesHash(raw)
	if facts.analysis != nil {
		if cached, exists := facts.analysis.Shapes[cacheKey]; exists {
			return cached
		}
	}
	node, ok := sourceResolveObject(facts, raw, seen, depth)
	if !ok {
		return sourceShape{}
	}
	var typ, format string
	_ = json.Unmarshal(node["type"], &typ)
	_ = json.Unmarshal(node["format"], &format)
	result := sourceShape{Known: typ != "" || len(node["properties"]) > 0}
	if format == "binary" || format == "byte" {
		result.Known = true
		result.Binary = true
	}
	if typ == "array" {
		item, valid := sourceResolveObject(facts, node["items"], copySourceSeen(seen), depth+1)
		var itemType string
		if valid {
			_ = json.Unmarshal(item["type"], &itemType)
		}
		result.Collection = valid && (itemType == "object" || len(item["properties"]) > 0)
		shape := sourceSchemaShape(facts, node["items"], copySourceSeen(seen), depth+1)
		result.Binary = result.Binary || shape.Binary
		result.Known = result.Known && shape.Known
	}
	var props map[string]json.RawMessage
	if len(node["properties"]) > 0 && json.Unmarshal(node["properties"], &props) == nil {
		for name, property := range props {
			shape := sourceSchemaShape(facts, property, copySourceSeen(seen), depth+1)
			result.Binary = result.Binary || shape.Binary
			result.Known = result.Known && shape.Known
			if name == "data" || name == "items" || name == "results" || name == "values" || name == "records" {
				result.Collection = result.Collection || shape.Collection
			}
		}
	}
	for _, keyword := range []string{"allOf", "oneOf", "anyOf"} {
		var branches []json.RawMessage
		if len(node[keyword]) == 0 {
			continue
		}
		if err := json.Unmarshal(node[keyword], &branches); err != nil || len(branches) == 0 {
			result.Known = false
			continue
		}
		known := true
		for _, branch := range branches {
			shape := sourceSchemaShape(facts, branch, copySourceSeen(seen), depth+1)
			known = known && shape.Known
			result.Binary = result.Binary || shape.Binary
			result.Collection = result.Collection || shape.Collection
		}
		result.Known = known
	}
	if facts.analysis != nil {
		facts.analysis.Shapes[cacheKey] = result
	}
	return result
}

func copySourceSeen(in map[string]bool) map[string]bool {
	out := map[string]bool{}
	for key, value := range in {
		out[key] = value
	}
	return out
}
