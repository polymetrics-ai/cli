package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type sourceLaneReason struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

type sourceLaneSchemaRole string

const (
	sourceLaneSchemaRequest  sourceLaneSchemaRole = "request"
	sourceLaneSchemaResponse sourceLaneSchemaRole = "response"
	sourceLaneSchemaRecord   sourceLaneSchemaRole = "record"
)

type sourceLaneFieldTargetKind string

const (
	sourceLaneFieldSchema    sourceLaneFieldTargetKind = "schema"
	sourceLaneFieldConfig    sourceLaneFieldTargetKind = "config"
	sourceLaneFieldParameter sourceLaneFieldTargetKind = "parameter"
)

type sourceLaneFieldTarget struct {
	Kind    sourceLaneFieldTargetKind `json:"kind"`
	Pointer *string                   `json:"pointer"`
}

type sourceLaneFieldMapping struct {
	Source sourceFactRef         `json:"source"`
	Target sourceLaneFieldTarget `json:"target"`
}

type sourceLaneGraphQLRefs struct {
	OperationName *sourceFactRef `json:"operation_name,omitempty"`
	Document      *sourceFactRef `json:"document,omitempty"`
	RequestSchema *sourceFactRef `json:"request_schema,omitempty"`
}

type sourceLaneTargetRef struct {
	Kind             string                   `json:"kind"`
	Connector        string                   `json:"connector"`
	ID               string                   `json:"id"`
	Lane             string                   `json:"lane"`
	Artifact         string                   `json:"artifact"`
	Pointer          string                   `json:"pointer"`
	ArtifactSHA256   string                   `json:"artifact_sha256"`
	CanonicalID      string                   `json:"canonical_id"`
	CanonicalPointer string                   `json:"canonical_pointer"`
	Generation       string                   `json:"generation"`
	SchemaRole       sourceLaneSchemaRole     `json:"schema_role"`
	SourceSchema     *sourceFactRef           `json:"source_schema,omitempty"`
	FieldMappings    []sourceLaneFieldMapping `json:"field_mappings,omitempty"`
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
	GraphQL              *sourceLaneGraphQLRefs `json:"graphql,omitempty"`
	Key                  sourceOperationKey     `json:"key"`
	Semantics            string                 `json:"semantics"`
	Citation             sourceFactRef          `json:"citation"`
	Clause               string                 `json:"clause"`
	FoundationGap        string                 `json:"foundation_gap"`
	AtlasID              string                 `json:"atlas_id"`
	DecisionRefs         []string               `json:"decision_refs"`
	IntendedBindings     []sourceLaneTargetRef  `json:"intended_bindings"`
	MaterializedBindings []sourceLaneTargetRef  `json:"materialized_bindings"`
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
	request := sourceRequestShape(facts)
	if facts.analysis.Exhausted {
		// Partial traversal cannot establish a shape contract for this operation.
		response, request = sourceShape{}, sourceShape{}
	}
	if response.Binary {
		set(2, true, "successful_binary_response", "responses")
	} else if response.Known {
		set(2, false, "known_nonbinary_response", "responses")
	}
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
	} else if registrationRef, registration := sourceRegistrationDemand(facts); registration {
		// A registration contract neither implements event intake nor excludes the
		// source-backed demand merely because OpenAPI callbacks are absent.
		cells[6].RuleID = "source_registration_demand_unresolved"
		cells[6].FactRefs = append(cells[6].FactRefs, registrationRef)
		cells[6].OwnerRefs = append(cells[6].OwnerRefs, "CP13")
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
	if a.Semantics != "" && sourceActionSemantics(a.Clause) != a.Semantics {
		return fmt.Errorf("semantic interpretation lacks affirmative cited support")
	}
	if asserted := sourceActionSemantics(sourceFactText(facts, "summary")); a.Semantics != "" && asserted != "" && asserted != a.Semantics {
		return fmt.Errorf("semantic interpretation contradicts source action")
	}
	if a.Semantics == "mutation" && (facts.Method == "GET" || facts.Method == "HEAD") {
		return fmt.Errorf("mutation conflicts with safe read contract")
	}
	if a.FoundationGap != "" {
		return validateSourceFoundationDemand(key, facts, a)
	}
	return nil
}

type sourceShape struct{ Known, Binary, Collection bool }

type sourceShapeAnalysis struct {
	Objects   map[string]map[string]json.RawMessage
	Shapes    map[string]sourceShape
	Root      any
	Visits    int
	Exhausted bool
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
	mediaKeys := make([]string, 0, len(content))
	for media := range content {
		mediaKeys = append(mediaKeys, media)
	}
	sort.Strings(mediaKeys)
	for _, media := range mediaKeys {
		rawEntry := content[media]
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
			facts.analysis.Exhausted = true
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
	unresolved := false
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
		unresolved = unresolved || !shape.Known
	}
	var props map[string]json.RawMessage
	if len(node["properties"]) > 0 && json.Unmarshal(node["properties"], &props) == nil {
		propertyKeys := make([]string, 0, len(props))
		for name := range props {
			propertyKeys = append(propertyKeys, name)
		}
		sort.Strings(propertyKeys)
		for _, name := range propertyKeys {
			property := props[name]
			shape := sourceSchemaShape(facts, property, copySourceSeen(seen), depth+1)
			result.Binary = result.Binary || shape.Binary
			unresolved = unresolved || !shape.Known
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
			unresolved = true
			continue
		}
		known := true
		for _, branch := range branches {
			shape := sourceSchemaShape(facts, branch, copySourceSeen(seen), depth+1)
			known = known && shape.Known
			result.Binary = result.Binary || shape.Binary
			result.Collection = result.Collection || shape.Collection
		}
		unresolved = unresolved || !known
		result.Known = result.Known || known
	}
	// A later known composition cannot erase an unresolved sibling contract.
	result.Known = result.Known && !unresolved
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

// sourceRegistrationDemand recognizes affirmative registration descriptions
// from retained prose, never operation IDs or executable artifacts. The result
// only preserves an unresolved demand; it grants no foundation or lane support.
func sourceRegistrationDemand(facts sourceFacts) (sourceFactRef, bool) {
	if facts.Method != "POST" && facts.Method != "PUT" && facts.Method != "PATCH" {
		return sourceFactRef{}, false
	}
	for _, group := range []string{"summary", "description"} {
		ref, exists := facts.Refs[group]
		if !exists {
			continue
		}
		text := strings.ToLower(strings.TrimSpace(sourceFactText(facts, group)))
		affirmative := false
		for _, verb := range []string{"create", "creates", "add", "adds", "register", "registers", "update", "updates"} {
			if strings.HasPrefix(text, verb+" ") {
				affirmative = true
				break
			}
		}
		if !affirmative {
			continue
		}
		// Only the first sentence names the action; later examples cannot turn an
		// unrelated mutation into a registration claim.
		clause := strings.SplitN(text, ".", 2)[0]
		for _, word := range strings.FieldsFunc(clause, func(r rune) bool { return !(r >= 'a' && r <= 'z') }) {
			if word == "webhook" || word == "webhooks" || word == "hook" || word == "hooks" {
				return ref, true
			}
		}
	}
	return sourceFactRef{}, false
}

// sourceLaneProjectionPointer validates syntax only. The binding owner resolves
// each coordinate against the exact selected schema or parameter container.
func sourceLaneProjectionPointer(pointer string) bool {
	if pointer == "" {
		return true
	}
	if !strings.HasPrefix(pointer, "/") {
		return false
	}
	depth := 0
	for i := 0; i < len(pointer); i++ {
		if pointer[i] == '/' {
			depth++
			if depth > 256 {
				return false
			}
		}
		if pointer[i] == '~' {
			if i+1 == len(pointer) || (pointer[i+1] != '0' && pointer[i+1] != '1') {
				return false
			}
			i++
		}
	}
	return true
}

func sourceLaneJSONFactRefShape(ref sourceFactRef) bool {
	return ref.DocumentID != "" && ref.Section == "" && ref.Part == "" &&
		sourceLaneProjectionPointer(ref.Pointer) && sourceLaneDigest(ref.ValueSHA256)
}

func sourceLaneGraphQLRefsShape(refs *sourceLaneGraphQLRefs) bool {
	if refs == nil {
		return true
	}
	for _, ref := range []*sourceFactRef{refs.OperationName, refs.Document, refs.RequestSchema} {
		if ref != nil && !sourceLaneJSONFactRefShape(*ref) {
			return false
		}
	}
	return true
}

// sourceLaneTargetRefShape checks the closed representation, not effective
// source ownership, target existence or executable projection semantics.
func sourceLaneTargetRefShape(ref sourceLaneTargetRef) error {
	switch ref.Kind {
	case "operation", "write", "stream", "command", "schema", "canonical_operation", "sync_transport":
	default:
		return fmt.Errorf("unknown target kind")
	}
	lane := false
	for _, name := range sourceLaneNames() {
		lane = lane || ref.Lane == name
	}
	if !lane {
		return fmt.Errorf("unknown target lane")
	}
	switch ref.SchemaRole {
	case "", sourceLaneSchemaRequest, sourceLaneSchemaResponse, sourceLaneSchemaRecord:
	default:
		return fmt.Errorf("unknown schema role")
	}
	if ref.Kind == "sync_transport" && (ref.SchemaRole != "" || ref.SourceSchema != nil || len(ref.FieldMappings) > 0) {
		return fmt.Errorf("transport descriptor has schema projection")
	}
	if ref.SourceSchema != nil && (!sourceLaneJSONFactRefShape(*ref.SourceSchema) || ref.SchemaRole == "") {
		return fmt.Errorf("invalid source schema citation or role")
	}
	type coordinate struct{ owner, pointer string }
	sources := make([]coordinate, 0, len(ref.FieldMappings))
	targets := make([]coordinate, 0, len(ref.FieldMappings))
	for _, mapping := range ref.FieldMappings {
		if !sourceLaneJSONFactRefShape(mapping.Source) || mapping.Target.Pointer == nil || !sourceLaneProjectionPointer(*mapping.Target.Pointer) {
			return fmt.Errorf("invalid field citation or target pointer")
		}
		pointer := *mapping.Target.Pointer
		switch mapping.Target.Kind {
		case sourceLaneFieldSchema:
			if ref.SchemaRole == "" {
				return fmt.Errorf("schema field has no role")
			}
		case sourceLaneFieldConfig:
			if pointer == "" {
				return fmt.Errorf("config root is not a field binding")
			}
		case sourceLaneFieldParameter:
			parts := strings.Split(pointer, "/")
			if len(parts) != 3 || (parts[1] != "parameters" && parts[1] != "pagination_parameters") {
				return fmt.Errorf("invalid parameter coordinate")
			}
			index := parts[2]
			if index == "" || (len(index) > 1 && index[0] == '0') {
				return fmt.Errorf("invalid parameter index")
			}
			for _, digit := range index {
				if digit < '0' || digit > '9' {
					return fmt.Errorf("invalid parameter index")
				}
			}
		default:
			return fmt.Errorf("unknown field target kind")
		}
		sources = append(sources, coordinate{mapping.Source.DocumentID, mapping.Source.Pointer})
		targets = append(targets, coordinate{string(mapping.Target.Kind), pointer})
	}
	for _, coordinates := range [][]coordinate{sources, targets} {
		sort.Slice(coordinates, func(i, j int) bool {
			if coordinates[i].owner != coordinates[j].owner {
				return coordinates[i].owner < coordinates[j].owner
			}
			return coordinates[i].pointer < coordinates[j].pointer
		})
		seen := map[coordinate]bool{}
		for _, current := range coordinates {
			if seen[current] {
				return fmt.Errorf("duplicate conflicting or overlapping field mappings")
			}
			for i := range current.pointer {
				if current.pointer[i] == '/' && seen[coordinate{current.owner, current.pointer[:i]}] {
					return fmt.Errorf("duplicate conflicting or overlapping field mappings")
				}
			}
			seen[current] = true
		}
	}
	return nil
}

// canonicalSourceLaneTargetRef returns independent pointed values and sorted
// mapping content. Call shape validation first; normalization never hides bad
// duplicate/conflicting claims by deduplicating them.
func canonicalSourceLaneTargetRef(ref sourceLaneTargetRef) sourceLaneTargetRef {
	out := ref
	if ref.SourceSchema != nil {
		value := *ref.SourceSchema
		out.SourceSchema = &value
	}
	out.FieldMappings = nil
	if len(ref.FieldMappings) > 0 {
		out.FieldMappings = append([]sourceLaneFieldMapping(nil), ref.FieldMappings...)
		for i := range out.FieldMappings {
			if ref.FieldMappings[i].Target.Pointer != nil {
				value := *ref.FieldMappings[i].Target.Pointer
				out.FieldMappings[i].Target.Pointer = &value
			}
		}
		sort.Slice(out.FieldMappings, func(i, j int) bool {
			a, b := out.FieldMappings[i], out.FieldMappings[j]
			ap, bp := "", ""
			if a.Target.Pointer != nil {
				ap = *a.Target.Pointer
			}
			if b.Target.Pointer != nil {
				bp = *b.Target.Pointer
			}
			av := []string{a.Source.DocumentID, a.Source.Pointer, a.Source.ValueSHA256, a.Source.Section, a.Source.Part, string(a.Target.Kind), ap}
			bv := []string{b.Source.DocumentID, b.Source.Pointer, b.Source.ValueSHA256, b.Source.Section, b.Source.Part, string(b.Target.Kind), bp}
			for n := range av {
				if av[n] != bv[n] {
					return av[n] < bv[n]
				}
			}
			return false
		})
	}
	return out
}

func sourceLaneTargetRefEqual(a, b sourceLaneTargetRef) bool {
	if sourceLaneTargetRefShape(a) != nil || sourceLaneTargetRefShape(b) != nil {
		return false
	}
	return reflect.DeepEqual(canonicalSourceLaneTargetRef(a), canonicalSourceLaneTargetRef(b))
}

// Citation coordinates cannot collapse JSON null/missing pointers into a
// legitimate root pointer. Optional *sourceFactRef members still accept null.
func (ref *sourceFactRef) UnmarshalJSON(raw []byte) error {
	type plain sourceFactRef
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	for _, name := range []string{"document_id", "pointer", "value_sha256"} {
		value, present := fields[name]
		if !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("missing or null citation coordinate")
		}
	}
	for _, name := range []string{"section", "part"} {
		if value, present := fields[name]; present && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("null rendered citation coordinate")
		}
	}
	var value plain
	if err := decodeStrictJSON(raw, &value); err != nil {
		return err
	}
	*ref = sourceFactRef(value)
	return nil
}

func (role *sourceLaneSchemaRole) UnmarshalJSON(raw []byte) error {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("null schema role")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	*role = sourceLaneSchemaRole(value)
	return nil
}
