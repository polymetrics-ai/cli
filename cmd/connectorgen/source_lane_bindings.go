package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"math/big"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestidentity"
)

// sourceLaneBindingInputs carries read-only, already observed artifacts into
// source classification. Loading and admission remain separate from source
// membership; an empty collection cannot remove a provider row.
type sourceLaneBindingInputs struct {
	Authoring    map[string][]byte
	Bundles      map[string]engine.Bundle
	Observations []sourceLaneBindingObservation
	Artifacts    map[string][]byte
	Canonical    map[string]vNextCanonicalDescriptor
}

// resolveSourceLaneBindings reports each reference independently. Artifact
// existence never supplies membership or behavioral proof.
func resolveSourceLaneBindings(key sourceOperationKey, facts sourceFacts, annotation sourceSemanticAnnotation, cells []sourceLaneCell) []sourceLaneCell {
	all := append(append([]sourceLaneTargetRef{}, annotation.IntendedBindings...), annotation.MaterializedBindings...)
	conflicts := make([]string, len(all))
	for i := range all {
		for j := 0; j < i; j++ {
			a, b := all[i], all[j]
			a.FieldMappings = nil
			b.FieldMappings = nil
			if !sourceLaneTargetRefEqual(a, b) {
				continue
			}
			code := "target_projection_conflict"
			if sourceLaneTargetRefEqual(all[i], all[j]) {
				code = "target_duplicate_claim"
			}
			conflicts[i] = code
			conflicts[j] = code
		}
	}
	position := -1
	// Explicit source claims are checked even when no target materializes.
	graphqlClaims := sourceLaneGraphQLCitations(facts, annotation.GraphQL)
	for _, issue := range graphqlClaims {
		severity := "error"
		if strings.Contains(issue.Code, "unverified") {
			severity = "deficit"
		}
		for i := range cells {
			cells[i].Diagnostics = append(cells[i].Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{cells[i].Lane}, Stage: "reference", Code: issue.Code, Pointer: issue.Pointer, Owner: key.Connector, Severity: severity})
		}
	}
	for _, group := range []struct {
		refs    []sourceLaneTargetRef
		claimed bool
	}{{annotation.IntendedBindings, false}, {annotation.MaterializedBindings, true}} {
		for _, ref := range group.refs {
			position++
			index := -1
			for i := range cells {
				if cells[i].Lane == ref.Lane {
					index = i
					break
				}
			}
			add := func(code, severity string) {
				d := sourceLaneDiagnostic{Key: key, Lanes: []string{ref.Lane}, Stage: "reference", Code: code, Pointer: ref.Artifact + "#" + ref.Pointer, Owner: key.Connector, Severity: severity}
				if index < 0 {
					for i := range cells {
						cells[i].Diagnostics = append(cells[i].Diagnostics, d)
					}
				} else {
					cells[index].Diagnostics = append(cells[index].Diagnostics, d)
				}
			}
			if index < 0 || ref.Connector != key.Connector {
				add("target_scope_mismatch", "error")
				continue
			}
			if err := sourceLaneTargetRefShape(ref); err != nil {
				add("target_projection_shape_invalid", "error")
				continue
			}
			if conflicts[position] != "" {
				add(conflicts[position], "error")
			}
			if cells[index].Applicability != "applicable" {
				add("target_applicability_conflict", "error")
				continue
			}
			if ref.ID == "" {
				if group.claimed {
					add("materialized_target_unspecified", "error")
				} else {
					add("target_unspecified", "deficit")
				}
				continue
			}
			if ref.Artifact == "" || path.IsAbs(ref.Artifact) || path.Clean(ref.Artifact) != ref.Artifact || strings.HasPrefix(ref.Artifact, "../") || strings.Contains(ref.Artifact, "\\") || !validSourceID(ref.Artifact) || !strings.HasPrefix(ref.Artifact, "internal/connectors/defs/"+key.Connector+"/") {
				add("target_path_invalid", "error")
				continue
			}
			var raw []byte
			if facts.bindings != nil {
				raw = facts.bindings.Artifacts[ref.Artifact]
				if ref.Kind == "canonical_operation" {
					raw = facts.bindings.Authoring[ref.Artifact]
				}
			}
			claims := append(sourceLaneSuppliedClaims(facts, annotation, ref), graphqlClaims...)
			for _, issue := range claims {
				severity := sourceClaimSeverity(group.claimed)
				if !strings.Contains(issue.Code, "unverified") {
					severity = "error"
				}
				cells[index].Diagnostics = append(cells[index].Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{ref.Lane}, Stage: "reference", Code: issue.Code, Pointer: issue.Pointer, Owner: key.Connector, Severity: severity})
			}
			absent := func() {
				if group.claimed {
					add("materialized_target_absent", "error")
				} else {
					add("target_absent", "deficit")
				}
			}
			if len(raw) == 0 {
				absent()
				continue
			}
			if (group.claimed || ref.ArtifactSHA256 != "") && sourceBytesHash(raw) != ref.ArtifactSHA256 {
				add("target_digest_mismatch", "error")
				continue
			}
			expectedPointer, code := sourceLaneTargetPointer(raw, ref)
			if code == "target_absent" {
				absent()
				continue
			}
			if code != "" {
				severity := sourceClaimSeverity(group.claimed)
				if code != "target_contract_unverified" {
					severity = "error"
				}
				add(code, severity)
				continue
			}
			if ref.Pointer == "" && !group.claimed {
				ref.Pointer = expectedPointer
			}
			node, err := sourceJSONPointer(raw, ref.Pointer)
			if err != nil {
				add("target_pointer_mismatch", "error")
				continue
			}
			if ref.Pointer != expectedPointer {
				var identity struct {
					ID string `json:"id"`
				}
				if err := json.Unmarshal(node, &identity); err == nil && identity.ID != "" && identity.ID != ref.ID {
					add("target_identity_mismatch", "error")
				} else {
					add("target_pointer_mismatch", "error")
				}
				continue
			}
			issues := sourceLaneCheckPresent(key, facts, annotation, ref, raw, node)
			for _, issue := range issues {
				severity := sourceClaimSeverity(group.claimed)
				if !strings.Contains(issue.Code, "unverified") && !strings.Contains(issue.Code, "unavailable") {
					severity = "error"
				}
				d := sourceLaneDiagnostic{Key: key, Lanes: []string{ref.Lane}, Stage: "reference", Code: issue.Code, Pointer: issue.Pointer, Owner: key.Connector, Severity: severity}
				cells[index].Diagnostics = append(cells[index].Diagnostics, d)
			}
			if len(issues) == 0 && len(claims) == 0 && len(graphqlClaims) == 0 && conflicts[position] == "" {
				cells[index].References = append(cells[index].References, canonicalSourceLaneTargetRef(ref))
			}
		}
	}
	return sourceLaneCheckCoverage(key, facts, annotation, cells)
}

func sourceLaneSuppliedClaims(facts sourceFacts, a sourceSemanticAnnotation, ref sourceLaneTargetRef) []sourceLaneBindingIssue {
	issues := []sourceLaneBindingIssue{}
	if ref.SourceSchema != nil {
		_, _, code := sourceLaneSchemaAnchor(facts, a, ref)
		if code != "" {
			issues = append(issues, sourceLaneBindingIssue{code, ref.SourceSchema.Pointer})
		}
	}
	for _, m := range ref.FieldMappings {
		_, code := sourceLaneCitedValue(facts, m.Source)
		if code == "" {
			parameter := false
			for _, p := range facts.Parameters {
				parameter = parameter || p.Ref == m.Source
			}
			if !parameter {
				if ref.SourceSchema == nil {
					code = "source_binding_scope_mismatch"
				} else {
					_, code = sourceLaneSchemaProjection(facts, *ref.SourceSchema, m.Source.Pointer)
				}
			}
		}
		if code != "" {
			issues = append(issues, sourceLaneBindingIssue{code, m.Source.Pointer})
		}
	}
	return issues
}

func sourceLaneCheckCoverage(key sourceOperationKey, facts sourceFacts, a sourceSemanticAnnotation, cells []sourceLaneCell) []sourceLaneCell {
	type scope struct {
		ref     sourceFactRef
		request bool
	}
	scopes := []scope{}
	unknownScopes := []string{}
	for _, group := range []string{"request_body", "responses"} {
		owner, ok := facts.Refs[group]
		if !ok {
			continue
		}
		raw := facts.Groups[group]
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}
		root, ok := sourceResolveObject(facts, raw, map[string]bool{}, 0)
		if !ok {
			unknownScopes = append(unknownScopes, owner.Pointer)
			continue
		}
		mediaAt := func(node map[string]json.RawMessage, pointer string, request bool) {
			var media map[string]json.RawMessage
			if json.Unmarshal(node["content"], &media) != nil || len(media) == 0 {
				unknownScopes = append(unknownScopes, pointer)
				return
			}
			for name, value := range media {
				var entry map[string]json.RawMessage
				if json.Unmarshal(value, &entry) != nil || len(entry["schema"]) == 0 {
					unknownScopes = append(unknownScopes, pointer+"/content/"+escapeSourcePointer(name))
					continue
				}
				raw, err := canonicalSourceJSON(entry["schema"])
				if err != nil {
					continue
				}
				scopes = append(scopes, scope{sourceFactRef{DocumentID: owner.DocumentID, Pointer: pointer + "/content/" + escapeSourcePointer(name) + "/schema", ValueSHA256: sourceBytesHash(raw)}, request})
			}
		}
		if group == "request_body" {
			mediaAt(root, owner.Pointer, true)
		} else {
			for status, value := range root {
				if len(status) == 3 && status[0] == '2' {
					node, ok := sourceResolveObject(facts, value, map[string]bool{}, 0)
					if !ok {
						unknownScopes = append(unknownScopes, owner.Pointer+"/"+status)
					} else if status != "204" && status != "205" {
						mediaAt(node, owner.Pointer+"/"+status, false)
					}
				}
			}
		}
	}
	for i := range cells {
		cell := &cells[i]
		hasExecutable := false
		for _, ref := range cell.References {
			if ref.Kind != "schema" {
				hasExecutable = true
			}
		}
		if len(cell.References) > 0 && !hasExecutable {
			cell.Diagnostics = append(cell.Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{cell.Lane}, Stage: "reference", Code: "target_executable_coverage_unverified", Owner: key.Connector, Severity: "deficit"})
		}
		for _, ref := range cell.References {
			if ref.Kind == "schema" || ref.Kind == "sync_transport" {
				continue
			}
			for _, pointer := range unknownScopes {
				cell.Diagnostics = append(cell.Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{cell.Lane}, Stage: "reference", Code: "target_required_scope_unverified", Pointer: pointer, Owner: key.Connector, Severity: "deficit"})
			}
			for _, needed := range scopes {
				covered := false
				for _, other := range cell.References {
					if other.SourceSchema == nil || *other.SourceSchema != needed.ref || other.CanonicalID != ref.CanonicalID || other.Generation != ref.Generation {
						continue
					}
					if needed.request && other.SchemaRole != sourceLaneSchemaRequest || !needed.request && other.SchemaRole != sourceLaneSchemaRecord && other.SchemaRole != sourceLaneSchemaResponse {
						continue
					}
					if sourceLaneSharedConsumer(facts.bindings, ref, other) {
						covered = true
						break
					}
				}
				if !covered {
					cell.Diagnostics = append(cell.Diagnostics, sourceLaneDiagnostic{Key: key, Lanes: []string{cell.Lane}, Stage: "reference", Code: "target_required_scope_unverified", Pointer: needed.ref.Pointer, Owner: key.Connector, Severity: "deficit"})
				}
			}
		}
	}
	return cells
}

func sourceLaneSharedConsumer(inputs *sourceLaneBindingInputs, a, b sourceLaneTargetRef) bool {
	if inputs == nil {
		return false
	}
	consumers := func(ref sourceLaneTargetRef) []string {
		if ref.Kind == "schema" || ref.Kind == "canonical_operation" {
			descriptor := inputs.Canonical[ref.Connector]
			for _, source := range descriptor.Graph.Operations {
				if source.ID != ref.CanonicalID {
					continue
				}
				ids := []string{}
				if source.Stream != nil && (ref.SchemaRole == sourceLaneSchemaRecord || ref.SchemaRole == sourceLaneSchemaResponse || ref.Kind == "canonical_operation") {
					ids = append(ids, "stream:"+source.Stream.Spec.Name)
				}
				if source.Write != nil && (ref.SchemaRole == sourceLaneSchemaRequest || ref.Kind == "canonical_operation") {
					ids = append(ids, "write:"+source.Write.Spec.Name)
				}
				if source.Operation != nil && (ref.SchemaRole == sourceLaneSchemaRequest || ref.Kind == "canonical_operation") {
					ids = append(ids, "operation:"+source.Operation.Spec.ID)
				}
				return ids
			}
			return nil
		}
		if ref.Kind != "command" {
			return []string{ref.Kind + ":" + ref.ID}
		}
		binding, err := engine.ResolveImplementedCommandPath(inputs.Bundles[ref.Connector], ref.ID)
		if err != nil {
			return nil
		}
		return []string{binding.Binding.Kind + ":" + binding.Binding.ID}
	}
	x, y := consumers(a), consumers(b)
	if len(x) == 0 || len(y) == 0 {
		return false
	}
	for _, id := range x {
		if !sourceLaneContains(y, id) {
			return false
		}
	}
	return true
}

func sourceClaimSeverity(claimed bool) string {
	if claimed {
		return "error"
	}
	return "deficit"
}

// sourceLanePathsEqual applies only the retained, accepted bridge contract.
// Neither source facts nor runtime routing are modified by this comparison.
func sourceLanePathsEqual(facts sourceFacts, target string) bool {
	if facts.Path == target {
		return true
	}
	raw := facts.Groups["path_bridge"]
	ref, exists := facts.Refs["path_bridge"]
	if !exists || ref.DocumentID == "" || ref.Pointer != "/rest/path_bridge" {
		return false
	}
	canonical, err := canonicalSourceJSON(raw)
	if err != nil || sourceBytesHash(canonical) != ref.ValueSHA256 {
		return false
	}
	var bridge struct {
		SourcePrefix    string `json:"source_prefix"`
		ConnectorPrefix string `json:"connector_prefix"`
	}
	if err := decodeStrictJSON(json.RawMessage(raw), &bridge); err != nil || bridge.SourcePrefix != "/api/v4" || bridge.ConnectorPrefix != "" {
		return false
	}
	if !strings.HasPrefix(facts.Path, bridge.SourcePrefix+"/") {
		return false
	}
	return strings.TrimPrefix(facts.Path, bridge.SourcePrefix) == target
}

func sourceLaneRESTParameterContract(facts sourceFacts, target engine.RESTOperationSpec) string {
	const mismatch = "target_parameter_mismatch"
	const unknown = "target_parameter_contract_unverified"
	for _, diagnostic := range facts.Diagnostics {
		if diagnostic == "rendered_parameter_contract_unmapped" || strings.HasPrefix(diagnostic, "source_parameter") {
			return unknown
		}
	}
	targets := map[string]engine.OperationParameter{}
	parameters := append(append([]engine.OperationParameter{}, target.Parameters...), target.PaginationParameters...)
	for _, p := range parameters {
		key := p.In + ":" + p.Name
		if _, exists := targets[key]; exists {
			return mismatch
		}
		targets[key] = p
	}
	if len(targets) != len(facts.Parameters) {
		return mismatch
	}
	for _, p := range facts.Parameters {
		actual, exists := targets[p.In+":"+p.Name]
		if !exists || p.Required != actual.Required {
			return mismatch
		}
		var node map[string]json.RawMessage
		if err := decodeSourceJSON(p.Node, &node); err != nil {
			return unknown
		}
		for _, field := range []string{"content", "style", "explode", "allowReserved", "allowEmptyValue"} {
			if _, exists := node[field]; exists {
				return unknown
			}
		}
		schema, ok := sourceResolveObject(facts, node["schema"], map[string]bool{}, 0)
		if !ok {
			return unknown
		}
		for field := range schema {
			switch field {
			case "type", "enum", "minimum", "maximum", "description", "title", "default", "example", "examples":
			default:
				return unknown
			}
		}
		var sourceType string
		if err := json.Unmarshal(schema["type"], &sourceType); err != nil || sourceType == "" {
			return unknown
		}
		if sourceType != actual.Type {
			return mismatch
		}
		for _, bound := range []struct {
			name  string
			value any
		}{{"minimum", actual.Minimum}, {"maximum", actual.Maximum}} {
			raw, err := json.Marshal(bound.value)
			if err != nil {
				return unknown
			}
			source := schema[bound.name]
			if len(source) == 0 {
				source = json.RawMessage("null")
			}
			equal, known := sourceLaneNumericBoundEqual(source, raw)
			if !known {
				return unknown
			}
			if !equal {
				return mismatch
			}
		}
		var values []string
		if enum, exists := schema["enum"]; exists {
			if err := json.Unmarshal(enum, &values); err != nil {
				return unknown
			}
		}
		expected := append([]string{}, values...)
		actualValues := append([]string{}, actual.Values...)
		sort.Strings(expected)
		sort.Strings(actualValues)
		if len(expected) != len(actualValues) {
			return mismatch
		}
		for i := range expected {
			if expected[i] != actualValues[i] {
				return mismatch
			}
		}
		if actual.Repeatable || len(actual.Schema) > 0 || actual.MaxBytes > 0 {
			return unknown
		}
	}
	// The engine sends each fixed entry under its literal query key. Parameter
	// declarations at another wire location and CLI spellings confer no authority.
	fixedNames := make([]string, 0, len(target.Query))
	for name := range target.Query {
		fixedNames = append(fixedNames, name)
	}
	sort.Strings(fixedNames)
	for _, name := range fixedNames {
		value := target.Query[name]
		var source *sourceParameterFact
		for i := range facts.Parameters {
			p := &facts.Parameters[i]
			if p.In == "query" && p.Name == name {
				source = p
				break
			}
		}
		if source == nil {
			return unknown
		}
		var declaration map[string]json.RawMessage
		if decodeSourceJSON(source.Node, &declaration) != nil {
			return unknown
		}
		schema, ok := sourceResolveObject(facts, declaration["schema"], map[string]bool{}, 0)
		if !ok {
			return unknown
		}
		raw, exists := schema["default"]
		if !exists {
			return unknown
		}
		var expected any
		if decodeSourceJSON(raw, &expected) != nil {
			return unknown
		}
		kind := targets["query:"+name].Type
		switch expected := expected.(type) {
		case string:
			if kind != "string" {
				return unknown
			}
			if value != expected {
				return mismatch
			}
		case bool:
			if kind != "boolean" {
				return unknown
			}
			if value != strconv.FormatBool(expected) {
				return mismatch
			}
		case json.Number:
			if kind != "number" && kind != "integer" {
				return unknown
			}
			equal, known := sourceLaneNumericBoundEqual([]byte(expected), []byte(value))
			if !known {
				return unknown
			}
			if !equal {
				return mismatch
			}
		default:
			return unknown
		}
	}

	return ""
}

// sourceLaneNumericBoundEqual compares decimal coefficients and exponents;
// it never expands an exponent into a huge integer or rounds through float64.
func sourceLaneNumericBoundEqual(left, right []byte) (bool, bool) {
	normalize := func(raw []byte) (string, bool) {
		if len(raw) > 4096 {
			return "", false
		}
		var value any
		if err := decodeSourceJSON(raw, &value); err != nil {
			return "", false
		}
		if value == nil {
			return "null", true
		}
		number, ok := value.(json.Number)
		if !ok {
			return "", false
		}
		return sourceLaneNumberKey(string(number)), true
	}
	a, ok := normalize(left)
	if !ok {
		return false, false
	}
	b, ok := normalize(right)
	if !ok {
		return false, false
	}
	return a == b, true
}

// sourceLaneBindingObservation is connector-scoped until manifest assembly
// attaches it to every affected anchored source key and lane.
type sourceLaneBindingObservation struct {
	Connector string
	Stage     string
	Code      string
	Pointer   string
}

func collectSourceLaneBindings(ctx context.Context, repo string, cohort sourceLaneCohort) (result sourceLaneBindingInputs) {
	result = sourceLaneBindingInputs{Authoring: map[string][]byte{}, Artifacts: map[string][]byte{}, Canonical: map[string]vNextCanonicalDescriptor{}, Bundles: map[string]engine.Bundle{}, Observations: []sourceLaneBindingObservation{}}
	connectors := map[string]bool{}
	for _, anchor := range cohort.Inventories {
		connectors[anchor.Connector] = true
	}
	names := make([]string, 0, len(connectors))
	for name := range connectors {
		names = append(names, name)
	}
	sort.Strings(names)
	add := func(connector, stage, code, pointer string) {
		result.Observations = append(result.Observations, sourceLaneBindingObservation{connector, stage, code, pointer})
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		for _, name := range names {
			add(name, "collection", "repository_unavailable", "")
		}
		return result
	}
	defer func() {
		if err := root.Close(); err != nil {
			for _, name := range names {
				add(name, "collection", "repository_close_failed", "")
			}
		}
	}()
	for _, name := range names {
		prefix := "internal/connectors/defs/" + name
		if !validSourceID(name) || path.Base(name) != name || name == "." || name == ".." || strings.Contains(name, "\\") {
			add(name, "collection", "connector_path_invalid", "")
			continue
		}
		if ctx.Err() != nil {
			add(name, "collection", "cancelled", prefix)
			continue
		}
		files := []string{"metadata.json", "spec.json", "streams.json", "writes.json", "operations.json", "cli_surface.json", "rate_limits.json", "changefeed.json", "polling_watermark.json", "sync_transport.json", "database.json"}
		// Walk only the closed execution-schema subtree. WalkDir never follows
		// symlinks; each selected leaf is checked again by the confined reader.
		err := fs.WalkDir(root.FS(), prefix+"/schemas", func(name string, entry fs.DirEntry, walkErr error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() && manifestidentity.IsExecutionJSONFile(strings.TrimPrefix(name, prefix+"/")) {
				files = append(files, strings.TrimPrefix(name, prefix+"/"))
			}
			return nil
		})
		if err != nil && !vNextPublicationPureNotExist(err) {
			// Missing schema directories and unreadable schema inventories are kept
			// as observations; neither can create a canonical or materialized claim.
			add(name, "collection", "schema_inventory_unavailable", prefix+"/schemas")
		}
		sort.Strings(files)
		outputs := map[string][]byte{}
		invalidInputs := false
		for _, file := range files {
			if ctx.Err() != nil {
				add(name, "collection", "cancelled", prefix+"/"+file)
				break
			}
			raw, err := readSourceInput(root, prefix+"/"+file, 64<<20)
			if err != nil {
				if !vNextPublicationPureNotExist(err) {
					add(name, "collection", "artifact_input_invalid", prefix+"/"+file)
					invalidInputs = true
				}
				continue // engine.Load diagnoses purely absent required members.
			}
			outputs[file] = raw
			result.Artifacts[prefix+"/"+file] = raw
		}
		if ctx.Err() != nil {
			continue
		}
		bundle, err := engine.Load(newVNextExecutionFS(name, outputs), name)
		if invalidInputs || err != nil {
			add(name, "execution_load", "execution_bundle_invalid", prefix)
		} else {
			result.Bundles[name] = bundle
		}
		raw, err := readSourceInput(root, prefix+"/source.lock.json", 64<<20)
		if err != nil {
			add(name, "canonical_import", "canonical_input_unavailable", prefix+"/source.lock.json")
			continue
		}
		result.Authoring[prefix+"/source.lock.json"] = raw
		var header struct {
			SchemaVersion int `json:"schema_version"`
		}
		if err := decodeSourceJSON(raw, &header); err != nil {
			add(name, "canonical_import", "canonical_input_invalid", prefix+"/source.lock.json")
			continue
		}
		if header.SchemaVersion != 4 {
			add(name, "canonical_import", "canonical_schema_unavailable", prefix+"/source.lock.json")
			continue
		}
		lock, err := decodeVNextSourceLock(raw)
		if err != nil {
			add(name, "canonical_import", "canonical_input_invalid", prefix+"/source.lock.json")
			continue
		}
		if lock.Connector != name {
			add(name, "canonical_import", "canonical_connector_mismatch", prefix+"/source.lock.json")
			continue
		}
		if ctx.Err() != nil {
			add(name, "canonical_generation", "cancelled", prefix+"/source.lock.json")
			continue
		}
		descriptor, err := canonicalizeVNextSourceLock(lock)
		if err != nil {
			add(name, "canonical_generation", "canonical_generation_unavailable", prefix+"/source.lock.json")
			continue
		}
		result.Canonical[name] = descriptor
	}
	return result
}

func sourceLaneTargetPointer(raw []byte, ref sourceLaneTargetRef) (string, string) {
	collection, identity, file := "", "", ""
	switch ref.Kind {
	case "canonical_operation":
		collection, identity, file = "operations", "id", "source.lock.json"
	case "schema":
		if !validVNextSchemaPath(ref.ID) || ref.Artifact != "internal/connectors/defs/"+ref.Connector+"/"+ref.ID {
			return "", "target_artifact_kind_mismatch"
		}
		var node map[string]json.RawMessage
		if decodeSourceJSON(raw, &node) != nil || node == nil {
			return "", "target_shape_invalid"
		}
		return "", ""
	case "sync_transport":
		if ref.Artifact != "internal/connectors/defs/"+ref.Connector+"/sync_transport.json" {
			return "", "target_artifact_kind_mismatch"
		}
		if ref.Pointer != "/source_transport" && ref.Pointer != "/destination_transport" {
			return "", "target_pointer_mismatch"
		}
		node, err := sourceJSONPointer(raw, ref.Pointer)
		if err != nil {
			return "", "target_absent"
		}
		var role struct {
			Executor connectors.TransportExecutorReference `json:"executor"`
		}
		if json.Unmarshal(node, &role) != nil {
			return "", "target_shape_invalid"
		}
		if role.Executor.ID != ref.ID {
			return "", "target_identity_mismatch"
		}
		return ref.Pointer, ""
	case "operation":
		collection, identity, file = "operations", "id", "operations.json"
	case "write":
		collection, identity, file = "actions", "name", "writes.json"
	case "stream":
		collection, identity, file = "streams", "name", "streams.json"
	case "command":
		collection, identity, file = "commands", "path", "cli_surface.json"
	default:
		return "", "target_contract_unverified"
	}
	if ref.Artifact != "" && ref.Artifact != "internal/connectors/defs/"+ref.Connector+"/"+file {
		return "", "target_artifact_kind_mismatch"
	}
	var root map[string]json.RawMessage
	if err := decodeSourceJSON(raw, &root); err != nil {
		return "", "target_shape_invalid"
	}
	var nodes []json.RawMessage
	if err := json.Unmarshal(root[collection], &nodes); err != nil {
		return "", "target_shape_invalid"
	}
	pointer := ""
	for i, node := range nodes {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(node, &fields); err != nil {
			return "", "target_shape_invalid"
		}
		var id string
		if err := json.Unmarshal(fields[identity], &id); err != nil {
			return "", "target_shape_invalid"
		}
		if id != ref.ID {
			continue
		}
		if pointer != "" {
			return "", "target_identity_ambiguous"
		}
		pointer = "/" + collection + "/" + strconv.Itoa(i)
	}
	if pointer == "" {
		return "", "target_absent"
	}
	return pointer, ""
}

type sourceLaneTypedTarget struct {
	Binding engine.ResolvedCommandBinding
	REST    *engine.RESTOperationSpec
	Write   *engine.WriteAction
	Stream  *engine.StreamSpec
	GraphQL *engine.GraphQLOperationSpec
	RawPath string
}

func sourceLaneObserveTypedTarget(ref sourceLaneTargetRef, node []byte, inputs *sourceLaneBindingInputs) (sourceLaneTypedTarget, string) {
	var result sourceLaneTypedTarget
	if inputs == nil {
		return result, "target_execution_unavailable"
	}
	bundle, exists := inputs.Bundles[ref.Connector]
	if !exists {
		return result, "target_execution_unavailable"
	}
	command := connectors.CommandSurfaceCommand{Path: ref.ID, Intent: ref.Lane}
	switch ref.Kind {
	case "command":
		var declared engine.CLICommand
		if err := decodeStrictJSON(node, &declared); err != nil {
			return result, "target_shape_invalid"
		}
		if declared.Path != ref.ID {
			return result, "target_identity_mismatch"
		}
		if declared.Intent != ref.Lane {
			return result, "target_lane_mismatch"
		}
		if declared.Availability != "implemented" {
			return result, "target_execution_unavailable"
		}
		binding, err := engine.ResolveImplementedCommandPath(bundle, ref.ID)
		if err != nil {
			return result, "target_binding_unresolved"
		}
		file := ""
		switch binding.Binding.Kind {
		case connectors.CommandBindingOperation:
			file = "operations.json"
		case connectors.CommandBindingWrite:
			file = "writes.json"
		case connectors.CommandBindingStream:
			file = "streams.json"
		default:
			return result, "target_contract_unverified"
		}
		target := sourceLaneTargetRef{Kind: binding.Binding.Kind, ID: binding.Binding.ID, Connector: ref.Connector, Lane: ref.Lane, Artifact: "internal/connectors/defs/" + ref.Connector + "/" + file}
		raw := inputs.Artifacts[target.Artifact]
		pointer, code := sourceLaneTargetPointer(raw, target)
		if code != "" {
			return result, code
		}
		selected, err := sourceJSONPointer(raw, pointer)
		if err != nil {
			return result, "target_pointer_mismatch"
		}
		result, code = sourceLaneObserveTypedTarget(target, selected, inputs)
		if code != "" {
			return result, code
		}
		result.Binding = binding
		return result, ""
	case "operation":
		var op engine.OperationSpec
		if err := decodeStrictJSON(node, &op); err != nil {
			return result, "target_shape_invalid"
		}
		if op.ID != ref.ID {
			return result, "target_identity_mismatch"
		}
		if op.GraphQL != nil {
			if (ref.Lane != "direct_read" || op.Kind != "graphql_query") && (ref.Lane != "direct_write" || op.Kind != "graphql_mutation") {
				return result, "target_lane_mismatch"
			}
			command.Operation = ref.ID
			result.GraphQL = op.GraphQL
			result.RawPath = op.GraphQL.Path
			break
		}
		if op.REST == nil {
			return result, "target_contract_unverified"
		}
		if (ref.Lane == "direct_read" && op.Kind != "rest_read") || (ref.Lane == "direct_write" && op.Kind != "rest_write") || (ref.Lane != "direct_read" && ref.Lane != "direct_write") {
			return result, "target_lane_mismatch"
		}
		command.Operation = ref.ID
		result.REST = op.REST
		result.RawPath = op.REST.Path
	case "write":
		var action engine.WriteAction
		if err := decodeStrictJSON(node, &action); err != nil {
			return result, "target_shape_invalid"
		}
		if action.Name != ref.ID {
			return result, "target_identity_mismatch"
		}
		if ref.Lane != "direct_write" && ref.Lane != "reverse_etl" {
			return result, "target_lane_mismatch"
		}
		command.Write = ref.ID
		result.Write = &action
		result.RawPath = action.Path
	case "stream":
		var stream engine.StreamSpec
		if err := decodeStrictJSON(node, &stream); err != nil {
			return result, "target_shape_invalid"
		}
		if stream.Name != ref.ID {
			return result, "target_identity_mismatch"
		}
		if ref.Lane != "direct_read" && ref.Lane != "etl" {
			return result, "target_lane_mismatch"
		}
		command.Stream = ref.ID
		result.Stream = &stream
		result.RawPath = stream.Path
	default:
		return result, "target_contract_unverified"
	}
	binding, err := engine.ResolveImplementedCommandBinding(bundle, command)
	if err != nil {
		return result, "target_binding_unresolved"
	}
	result.Binding = binding
	return result, ""
}

type sourceLaneBindingIssue struct{ Code, Pointer string }

func sourceLaneCitedValue(facts sourceFacts, ref sourceFactRef) (json.RawMessage, string) {
	if !sourceLaneJSONFactRefShape(ref) {
		return nil, "source_binding_citation_mismatch"
	}
	owned := false
	for _, owner := range facts.Refs {
		if owner.DocumentID == ref.DocumentID {
			owned = true
			break
		}
	}
	if !owned {
		return nil, "source_binding_citation_mismatch"
	}
	raw, err := sourceJSONPointer(facts.Document, ref.Pointer)
	if err != nil {
		return nil, "source_binding_citation_mismatch"
	}
	canonical, err := canonicalSourceJSON(raw)
	if err != nil || sourceBytesHash(canonical) != ref.ValueSHA256 {
		return nil, "source_binding_citation_mismatch"
	}
	return raw, ""
}

func sourceLaneGraphQLCitations(facts sourceFacts, refs *sourceLaneGraphQLRefs) []sourceLaneBindingIssue {
	var issues []sourceLaneBindingIssue
	if refs == nil {
		return issues
	}
	owner := facts.Refs["source_operation"]
	for _, ref := range []*sourceFactRef{refs.OperationName, refs.Document, refs.RequestSchema} {
		if ref == nil {
			continue
		}
		_, code := sourceLaneCitedValue(facts, *ref)
		if code == "" && (owner.DocumentID != ref.DocumentID || !strings.HasPrefix(ref.Pointer, owner.Pointer+"/")) {
			if owner.DocumentID != ref.DocumentID {
				code = "source_binding_scope_mismatch"
			} else {
				code = sourceLaneLinkedCitation(facts, owner.Pointer, ref.Pointer)
			}
		}
		if code != "" {
			issues = append(issues, sourceLaneBindingIssue{code, ref.Pointer})
		}
	}
	return issues
}

// sourceLaneLinkedCitation proves one local reference lineage from this
// operation to a physically retained value. It neither searches for equal
// values nor interprets provider-specific linking conventions.
func sourceLaneLinkedCitation(facts sourceFacts, root, wanted string) string {
	visits, matches := 0, 0
	unresolved := false
	var walk func(string, map[string]bool, int)
	walk = func(pointer string, seen map[string]bool, depth int) {
		visits++
		if visits > 4096 || depth > 128 || seen[pointer] {
			unresolved = true
			return
		}
		raw, err := sourceJSONPointer(facts.Document, pointer)
		if err != nil {
			unresolved = true
			return
		}
		if pointer == wanted {
			matches++
			return
		}
		next := make(map[string]bool, len(seen)+1)
		for key, value := range seen {
			next[key] = value
		}
		next[pointer] = true
		var object map[string]json.RawMessage
		if json.Unmarshal(raw, &object) == nil && object != nil {
			if edge, ok := object["$ref"]; ok {
				var ref string
				if json.Unmarshal(edge, &ref) != nil || !strings.HasPrefix(ref, "#/") {
					unresolved = true
				} else {
					walk(facts.RefPrefix+ref[1:], next, depth+1)
				}
			}
			for name := range object {
				if name != "$ref" {
					walk(pointer+"/"+escapeSourcePointer(name), next, depth+1)
				}
			}
			return
		}
		var array []json.RawMessage
		if json.Unmarshal(raw, &array) == nil {
			for i := range array {
				walk(pointer+"/"+strconv.Itoa(i), next, depth+1)
			}
		}
	}
	walk(root, map[string]bool{}, 0)
	if matches > 1 {
		return "source_projection_ambiguous"
	}
	if unresolved {
		return "source_schema_unverified"
	}
	if matches == 1 {
		return ""
	}
	return "source_binding_scope_mismatch"
}

// A projection retains the instance coordinate and each ancestor's presence
// condition. A schema pointer is never treated as an instance path.
type sourceLaneProjection struct {
	Raw        json.RawMessage
	Path       []string
	Required   []bool
	Occurrence sourceLaneOccurrence
}

// Only supported structural edges identify a literal instance occurrence.
// A definition stored below the schema root still needs reference-use search.
func sourceLaneLiteralProjectionOccurrence(root, wanted string) bool {
	if wanted == root {
		return true
	}
	if !strings.HasPrefix(wanted, root+"/") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(wanted, root+"/"), "/")
	for i := 0; i < len(parts); {
		switch parts[i] {
		case "properties":
			if i+1 >= len(parts) {
				return false
			}
			i += 2
		case "items":
			i++
		case "prefixItems":
			if i+1 >= len(parts) {
				return false
			}
			index, err := strconv.Atoi(parts[i+1])
			if err != nil || index < 0 || strconv.Itoa(index) != parts[i+1] {
				return false
			}
			i += 2
		default:
			return false
		}
	}
	return true
}

// Occurrence semantics and physical-search completeness are independent: one
// observed use is not unique until every relevant search branch is accounted for.
type sourceLaneLineageState uint8

const (
	sourceLaneLineageSupported sourceLaneLineageState = iota
	sourceLaneLineageContradictory
	sourceLaneLineageUnverified
)

type sourceLaneProjectionStep struct {
	Coordinate string
	Required   bool
	Coverage   string // property, uniform, fixed, or tail
	MinItems   int
	MaxItems   *int
}

type sourceLaneOccurrence struct {
	Literal  string
	Physical string
	Steps    []sourceLaneProjectionStep
	State    sourceLaneLineageState
}

type sourceLaneProjectionNode struct {
	Properties map[string]json.RawMessage
	Required   []string
	Prefix     []json.RawMessage
	Type       string
	MinItems   int
	MaxItems   *int
}

// Classify edge meaning independently of its JSON encoding. Definition storage
// is not an application; its container must still be a well-formed registry.
func sourceLaneCheckedProjectionNode(node map[string]json.RawMessage) (sourceLaneProjectionNode, sourceLaneLineageState) {
	var result sourceLaneProjectionNode
	state := sourceLaneLineageSupported
	for key, value := range node {
		switch key {
		case "$ref", "properties", "items", "prefixItems", "required", "type":
		case "$defs", "definitions":
			var registry map[string]json.RawMessage
			if json.Unmarshal(value, &registry) != nil || registry == nil {
				state = sourceLaneLineageUnverified
			}
		case "title", "description", "example", "examples", "$comment", "default", "enum", "const", "format", "pattern", "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf", "minLength", "maxLength", "minProperties", "maxProperties":
			// These constraints/annotations do not create instance-use edges. Exact
			// leaf constraint equivalence remains the schema comparison's contract.
		case "additionalProperties":
			var allowed bool
			if json.Unmarshal(value, &allowed) != nil || string(bytes.TrimSpace(value)) == "null" {
				state = sourceLaneLineageUnverified
			}
		case "nullable", "uniqueItems", "deprecated", "readOnly", "writeOnly":
			var flag bool
			if json.Unmarshal(value, &flag) != nil || string(bytes.TrimSpace(value)) == "null" {
				state = sourceLaneLineageUnverified
			}
		case "minItems", "maxItems":
			var count int
			if json.Unmarshal(value, &count) != nil || count < 0 || string(bytes.TrimSpace(value)) == "null" {
				state = sourceLaneLineageUnverified
				continue
			}
			if key == "minItems" {
				result.MinItems = count
			} else {
				result.MaxItems = &count
			}
		default:
			// Includes scalar-valued ref/scope/applicator keywords. None are silently
			// interpreted as harmless annotations or as definitely another use.
			state = sourceLaneLineageUnverified
		}
	}
	local := sourceLocalShapeEvidence(node)
	result.Type, result.Properties, result.Prefix = local.Type, local.Properties, local.Prefix
	if local.TypeState == sourceLocalTypeUnsupported || local.PropertiesState == sourceLocalMemberMalformed || local.PrefixState == sourceLocalPrefixMalformed {
		state = sourceLaneLineageUnverified
	}
	if local.PropertiesState == sourceLocalMemberValid && local.TypeState == sourceLocalTypeSupported && local.Type != "object" && state == sourceLaneLineageSupported {
		state = sourceLaneLineageContradictory
	}
	if raw, exists := node["required"]; exists {
		var entries []json.RawMessage
		if json.Unmarshal(raw, &entries) != nil || entries == nil {
			state = sourceLaneLineageUnverified
		}
		for _, entry := range entries {
			var name string
			if len(bytes.TrimSpace(entry)) == 0 || bytes.TrimSpace(entry)[0] != '"' || json.Unmarshal(entry, &name) != nil {
				state = sourceLaneLineageUnverified
				continue
			}
			result.Required = append(result.Required, name)
		}
		seen := map[string]bool{}
		for _, name := range result.Required {
			if seen[name] {
				state = sourceLaneLineageUnverified
			}
			seen[name] = true
		}
	}
	if local.TypeState == sourceLocalTypeAbsent && (local.HasItems || local.PrefixState != sourceLocalPrefixAbsent) {
		state = sourceLaneLineageUnverified
	}
	if result.MaxItems != nil && *result.MaxItems < result.MinItems {
		state = sourceLaneLineageContradictory
	}
	return result, state
}

func sourceLaneSchemaProjection(facts sourceFacts, root sourceFactRef, wanted string) (sourceLaneProjection, string) {
	var found []sourceLaneProjection
	visits := 0
	incomplete := false
	contradictory := false
	direct := sourceLaneLiteralProjectionOccurrence(root.Pointer, wanted)
	var walk func(string, json.RawMessage, sourceLaneOccurrence, map[string]bool, int)
	walk = func(pointer string, raw json.RawMessage, lineage sourceLaneOccurrence, seen map[string]bool, depth int) {
		if direct && pointer != wanted && !strings.HasPrefix(wanted, pointer+"/") {
			return
		}
		visits++
		if visits > 4096 || depth > 128 {
			incomplete = true
			return
		}
		var node map[string]json.RawMessage
		if decodeSourceJSON(raw, &node) != nil || node == nil {
			incomplete = true
			return
		}
		if seen[pointer] {
			incomplete = true
			return
		}
		next := make(map[string]bool, len(seen)+1)
		for k, v := range seen {
			next[k] = v
		}
		next[pointer] = true
		literalRaw := raw
		// Selecting the literal ref node keeps its identity, but never bypasses
		// validation of the consumed reference, siblings or resolved node semantics.
		if edge, exists := node["$ref"]; exists {
			var ref string
			if !sourceReferenceAnnotationSiblings(node) || json.Unmarshal(edge, &ref) != nil || !strings.HasPrefix(ref, "#/") || !sourceLaneProjectionPointer(ref[1:]) {
				incomplete = true
				return
			}
			physical := facts.RefPrefix + ref[1:]
			value, err := sourceJSONPointer(facts.Document, physical)
			if err != nil {
				incomplete = true
				return
			}
			if pointer != wanted {
				walk(physical, value, lineage, next, depth+1)
				return
			}
			for {
				visits++
				depth++
				if visits > 4096 || depth > 128 || next[physical] {
					incomplete = true
					return
				}
				next[physical] = true
				node = nil
				if decodeSourceJSON(value, &node) != nil || node == nil {
					incomplete = true
					return
				}
				lineage.Physical = physical
				edge, more := node["$ref"]
				if !more {
					break
				}
				if !sourceReferenceAnnotationSiblings(node) || json.Unmarshal(edge, &ref) != nil || !strings.HasPrefix(ref, "#/") || !sourceLaneProjectionPointer(ref[1:]) {
					incomplete = true
					return
				}
				physical = facts.RefPrefix + ref[1:]
				value, err = sourceJSONPointer(facts.Document, physical)
				if err != nil {
					incomplete = true
					return
				}
			}
		}
		shape, state := sourceLaneCheckedProjectionNode(node)
		if state == sourceLaneLineageUnverified {
			incomplete = true
		}
		if state == sourceLaneLineageContradictory {
			contradictory = true
		}
		lineage.State = state
		if pointer == wanted {
			if state != sourceLaneLineageSupported {
				return
			}
			lineage.Literal = pointer
			if lineage.Physical == "" {
				lineage.Physical = pointer
			}
			projection := sourceLaneProjection{Raw: literalRaw, Path: []string{}, Required: []bool{}, Occurrence: lineage}
			for _, step := range lineage.Steps {
				projection.Path = append(projection.Path, step.Coordinate)
				projection.Required = append(projection.Required, step.Required)
			}
			found = append(found, projection)
			return
		}
		// Unknown branches may still establish two distinct known uses during a
		// physical search. They cannot make a selected lineage supported.
		if state != sourceLaneLineageSupported {
			return
		}
		descend := func(child string, value json.RawMessage, step sourceLaneProjectionStep) {
			branch := lineage
			branch.Steps = append(append([]sourceLaneProjectionStep{}, lineage.Steps...), step)
			walk(child, value, branch, next, depth+1)
		}
		names := make([]string, 0, len(shape.Properties))
		for name := range shape.Properties {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			descend(pointer+"/properties/"+escapeSourcePointer(name), shape.Properties[name], sourceLaneProjectionStep{Coordinate: name, Required: sourceLaneContains(shape.Required, name), Coverage: "property"})
		}
		if items, exists := node["items"]; exists && shape.Type == "array" {
			coverage := "uniform"
			if sourceLocalShapeEvidence(node).arrayCoverage() == sourceLocalArrayPrefixOrTail {
				coverage = "tail"
			}
			descend(pointer+"/items", items, sourceLaneProjectionStep{Coordinate: "[]", Coverage: coverage, MinItems: shape.MinItems, MaxItems: shape.MaxItems})
		}
		for i, item := range shape.Prefix {
			if shape.Type != "array" {
				break
			}
			if shape.MaxItems != nil && i >= *shape.MaxItems {
				continue
			}
			descend(pointer+"/prefixItems/"+strconv.Itoa(i), item, sourceLaneProjectionStep{Coordinate: "[" + strconv.Itoa(i) + "]", Required: i < shape.MinItems, Coverage: "fixed", MinItems: shape.MinItems, MaxItems: shape.MaxItems})
		}
	}
	if !sourceLaneProjectionPointer(root.Pointer) || !sourceLaneProjectionPointer(wanted) {
		return sourceLaneProjection{}, "source_binding_scope_mismatch"
	}
	raw, err := sourceJSONPointer(facts.Document, root.Pointer)
	if err != nil {
		return sourceLaneProjection{}, "source_schema_unverified"
	}
	walk(root.Pointer, raw, sourceLaneOccurrence{}, map[string]bool{}, 0)
	if len(found) > 1 {
		return sourceLaneProjection{}, "source_projection_ambiguous"
	}
	if incomplete {
		return sourceLaneProjection{}, "source_schema_unverified"
	}
	if len(found) == 1 && !contradictory {
		return found[0], ""
	}
	return sourceLaneProjection{}, "source_binding_scope_mismatch"
}

func sourceLaneTargetProjection(raw json.RawMessage, pointer string) (sourceLaneProjection, string) {
	if !sourceLaneProjectionPointer(pointer) {
		return sourceLaneProjection{}, "target_schema_pointer_mismatch"
	}
	if _, err := sourceJSONPointer(raw, pointer); err != nil {
		return sourceLaneProjection{}, "target_schema_pointer_mismatch"
	}
	facts := sourceFacts{Document: raw}
	projection, code := sourceLaneSchemaProjection(facts, sourceFactRef{Pointer: ""}, pointer)
	if code == "source_schema_unverified" {
		return projection, "target_schema_unverified"
	}
	if code != "" {
		return projection, "target_schema_pointer_mismatch"
	}
	return projection, ""
}

func sourceLaneContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// Structural comparison deliberately has no implication/union solver. Local
// reference expansion is bounded; lexical numbers are never converted to float.
func sourceLaneSchemaCompare(facts sourceFacts, source, target json.RawMessage) string {
	// A separately unknown composition must not mask an explicit contradiction
	// in a constraint whose meaning is already known on both sides.
	left, leftOK := sourceResolveObject(facts, source, map[string]bool{}, 0)
	right, rightOK := sourceResolveObject(sourceFacts{Document: target}, target, map[string]bool{}, 0)
	if leftOK && rightOK {
		for _, field := range []string{"type", "format", "nullable", "const", "minimum", "maximum", "minLength", "maxLength", "minItems", "maxItems"} {
			x, xOK := left[field]
			y, yOK := right[field]
			if !xOK || !yOK {
				continue
			}
			var xv, yv any
			if decodeSourceJSON(x, &xv) != nil || decodeSourceJSON(y, &yv) != nil {
				continue
			}
			if _, ok := xv.(json.Number); ok {
				if _, ok := yv.(json.Number); ok {
					equal, known := sourceLaneNumericBoundEqual(x, y)
					if known && !equal {
						return "target_schema_mismatch"
					}
					continue
				}
			}
			if !reflect.DeepEqual(xv, yv) {
				return "target_schema_mismatch"
			}
		}
	}
	visits := 0
	var normalize func(sourceFacts, json.RawMessage, int) (any, bool)
	normalize = func(owner sourceFacts, raw json.RawMessage, depth int) (any, bool) {
		visits++
		if visits > 4096 || depth > 128 {
			return nil, false
		}
		node, ok := sourceResolveObject(owner, raw, map[string]bool{}, 0)
		if !ok {
			return nil, false
		}
		out := map[string]any{}
		for k, v := range node {
			switch k {
			case "description", "title", "example", "examples", "$comment":
				continue
			case "properties", "$defs", "definitions":
				if k != "properties" {
					continue
				}
				var props map[string]json.RawMessage
				if decodeSourceJSON(v, &props) != nil {
					return nil, false
				}
				values := map[string]any{}
				for name, child := range props {
					value, ok := normalize(owner, child, depth+1)
					if !ok {
						return nil, false
					}
					values[name] = value
				}
				out[k] = values
			case "items":
				value, ok := normalize(owner, v, depth+1)
				if !ok {
					return nil, false
				}
				out[k] = value
			case "prefixItems":
				var items []json.RawMessage
				if decodeSourceJSON(v, &items) != nil {
					return nil, false
				}
				values := []any{}
				for _, item := range items {
					value, ok := normalize(owner, item, depth+1)
					if !ok {
						return nil, false
					}
					values = append(values, value)
				}
				out[k] = values
			case "required", "enum":
				var values []json.RawMessage
				if decodeSourceJSON(v, &values) != nil {
					return nil, false
				}
				items := []string{}
				for _, value := range values {
					canonical, err := canonicalSourceJSON(value)
					if err != nil {
						return nil, false
					}
					var scalar any
					if decodeSourceJSON(value, &scalar) != nil {
						return nil, false
					}
					if number, ok := scalar.(json.Number); ok {
						items = append(items, "number:"+sourceLaneNumberKey(string(number)))
					} else {
						items = append(items, "json:"+string(canonical))
					}
				}
				sort.Strings(items)
				out[k] = items
			case "type", "format", "nullable", "const", "default", "additionalProperties", "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf", "minLength", "maxLength", "pattern", "minItems", "maxItems", "uniqueItems", "minProperties", "maxProperties":
				var value any
				if decodeSourceJSON(v, &value) != nil {
					return nil, false
				}
				if number, ok := value.(json.Number); ok { // normalized without rounding
					value = struct{ Decimal string }{sourceLaneNumberKey(string(number))}
				}
				out[k] = value
			default:
				return nil, false
			}
		}
		return out, true
	}
	a, ok := normalize(facts, source, 0)
	if !ok {
		return "source_schema_unverified"
	}
	b, ok := normalize(sourceFacts{Document: target}, target, 0)
	if !ok {
		return "target_schema_unverified"
	}
	if !reflect.DeepEqual(a, b) {
		return "target_schema_mismatch"
	}
	return ""
}

func sourceLaneNumberKey(text string) string {
	// Reuse the exact decimal comparison authority for the common integral and
	// exponent spellings by retaining a rational numerator/denominator.
	if len(text) > 4096 {
		return text
	}
	coefficient, exp, hasExp := strings.Cut(strings.ToLower(text), "e")
	exponent := new(big.Int)
	if hasExp {
		if _, ok := exponent.SetString(exp, 10); !ok {
			return text
		}
	}
	if dot := strings.IndexByte(coefficient, '.'); dot >= 0 {
		exponent.Sub(exponent, big.NewInt(int64(len(coefficient)-dot-1)))
		coefficient = coefficient[:dot] + coefficient[dot+1:]
	}
	sign := ""
	if strings.HasPrefix(coefficient, "-") {
		sign = "-"
		coefficient = coefficient[1:]
	}
	coefficient = strings.TrimLeft(coefficient, "0")
	if coefficient == "" {
		return "0e0"
	}
	short := strings.TrimRight(coefficient, "0")
	exponent.Add(exponent, big.NewInt(int64(len(coefficient)-len(short))))
	return sign + short + "e" + exponent.String()
}

func sourceLaneSchemaAnchor(facts sourceFacts, annotation sourceSemanticAnnotation, ref sourceLaneTargetRef) (json.RawMessage, string, string) {
	if ref.SourceSchema == nil {
		return nil, "", "source_schema_unverified"
	}
	raw, code := sourceLaneCitedValue(facts, *ref.SourceSchema)
	if code != "" {
		return nil, "", code
	}
	if annotation.GraphQL != nil && annotation.GraphQL.RequestSchema != nil && *annotation.GraphQL.RequestSchema == *ref.SourceSchema {
		if ref.SchemaRole != sourceLaneSchemaRequest {
			return nil, "", "target_schema_role_mismatch"
		}
		return raw, "application/json", ""
	}
	group := "request_body"
	if ref.SchemaRole == sourceLaneSchemaRecord || ref.SchemaRole == sourceLaneSchemaResponse {
		group = "responses"
	}
	owner := facts.Refs[group]
	if owner.DocumentID != ref.SourceSchema.DocumentID {
		return nil, "", "source_binding_scope_mismatch"
	}
	suffix := strings.TrimPrefix(ref.SourceSchema.Pointer, owner.Pointer)
	parts := strings.Split(suffix, "/")
	media := ""
	if group == "request_body" && len(parts) == 4 && parts[0] == "" && parts[1] == "content" && parts[3] == "schema" {
		media = parts[2]
	}
	if group == "responses" && len(parts) == 5 && parts[0] == "" && len(parts[1]) == 3 && parts[1][0] == '2' && parts[2] == "content" && parts[4] == "schema" {
		media = parts[3]
	}
	if media == "" || !strings.HasPrefix(ref.SourceSchema.Pointer, owner.Pointer+"/") {
		return nil, "", "source_binding_scope_mismatch"
	}
	var node map[string]json.RawMessage
	if decodeSourceJSON(raw, &node) != nil || node == nil {
		return nil, "", "source_binding_scope_mismatch"
	}
	return raw, strings.ReplaceAll(strings.ReplaceAll(media, "~1", "/"), "~0", "~"), ""
}

func sourceLaneEffectiveSchema(observed sourceLaneTypedTarget, bundle engine.Bundle, role sourceLaneSchemaRole) (json.RawMessage, string) {
	switch role {
	case sourceLaneSchemaRequest:
		if observed.Write != nil {
			return observed.Write.RecordSchema, ""
		}
		if observed.REST != nil && len(observed.REST.BodySchema) > 0 {
			return observed.REST.BodySchema, ""
		}
		if observed.GraphQL != nil {
			return observed.GraphQL.VariablesSchema, ""
		}
	case sourceLaneSchemaResponse, sourceLaneSchemaRecord:
		if observed.Stream != nil && observed.Stream.SchemaRef != "" {
			if schema, ok := bundle.Schemas[observed.Stream.Name]; ok {
				return schema.Raw, ""
			}
		}
	}
	if observed.Stream != nil && role == sourceLaneSchemaRequest || observed.Write != nil && role != sourceLaneSchemaRequest {
		return nil, "target_schema_role_mismatch"
	}
	return nil, "target_schema_consumer_unverified"
}

func sourceLaneCheckPresent(key sourceOperationKey, facts sourceFacts, a sourceSemanticAnnotation, ref sourceLaneTargetRef, raw, node []byte, scopeOnly ...bool) []sourceLaneBindingIssue {
	if ref.Kind == "canonical_operation" || ref.Kind == "schema" || ref.Kind == "sync_transport" {
		return sourceLaneCheckAggregate(key, facts, a, ref, raw)
	}
	issues := []sourceLaneBindingIssue{}
	add := func(code, pointer string) {
		if code != "" {
			issues = append(issues, sourceLaneBindingIssue{code, pointer})
		}
	}
	ptr := ref.Artifact + "#" + ref.Pointer
	observed, code := sourceLaneObserveTypedTarget(ref, node, facts.bindings)
	if code != "" {
		add(code, ptr)
		return issues
	}
	bundle := facts.bindings.Bundles[key.Connector]
	if observed.REST != nil {
		add(sourceLaneRESTParameterContract(facts, *observed.REST), ptr)
	} else {
		for _, parameter := range facts.Parameters {
			mapped := false
			for _, m := range ref.FieldMappings {
				mapped = mapped || m.Source == parameter.Ref
			}
			if !mapped {
				if parameter.Required {
					add("target_parameter_mismatch", parameter.Ref.Pointer)
				} else {
					add("target_parameter_contract_unverified", parameter.Ref.Pointer)
				}
			}
		}
		if observed.Write != nil && len(observed.Write.Query) > 0 || observed.Stream != nil && (len(observed.Stream.Query) > 0 || len(observed.Stream.Headers) > 0) {
			add("target_parameter_contract_unverified", ptr)
		}
	}
	add(sourceLaneRouteContract(facts, ref, observed, bundle), ptr)
	issues = append(issues, sourceLaneGraphQLContract(facts, a, ref, observed)...)
	if facts.bindings == nil || ref.CanonicalID == "" || ref.CanonicalPointer == "" || ref.Generation == "" {
		add("target_contract_unverified", ptr)
	} else if descriptor, exists := facts.bindings.Canonical[key.Connector]; !exists {
		add("canonical_binding_absent", ptr)
	} else {
		if descriptor.Connector != key.Connector || descriptor.Staged.Identity.Digest != ref.Generation {
			add("canonical_generation_mismatch", ptr)
		}
		if bundle.Identity.Digest != descriptor.Staged.Identity.Digest {
			add("execution_generation_mismatch", ptr)
		}
		if !bytes.Equal(raw, descriptor.Staged.Outputs[strings.TrimPrefix(ref.Artifact, "internal/connectors/defs/"+key.Connector+"/")]) {
			add("canonical_artifact_mismatch", ptr)
		}
		matches := 0
		for _, p := range descriptor.Staged.Provenance {
			if p.SourceID == ref.CanonicalID && p.FieldPath == ref.CanonicalPointer && p.TargetKind == ref.Kind && p.TargetID == ref.ID {
				matches++
			}
		}
		if matches != 1 {
			add("canonical_provenance_mismatch", ptr)
		}
		if facts.OperationID != "" {
			found := false
			for _, source := range descriptor.Graph.Operations {
				if source.ID != ref.CanonicalID {
					continue
				}
				found = true
				values, err := vNextDecodeSourceFacts(source)
				provider, supplied, e := vNextSourceFact(source, values, "provider_operation")
				if err != nil || e != nil || (supplied && provider != facts.OperationID) {
					add("target_operation_identity_mismatch", ptr)
				} else if !supplied {
					add("target_operation_identity_unverified", ptr)
				}
			}
			if !found {
				add("canonical_provenance_mismatch", ptr)
			}
		}
	}
	if ref.SourceSchema != nil || len(ref.FieldMappings) > 0 {
		issues = append(issues, sourceLaneProjectionContract(facts, a, ref, observed, bundle)...)
	}
	// Missing current provenance remains visible, but cannot hide a known
	// schema/template contradiction encountered above.
	if len(issues) == 0 || ref.SourceSchema != nil || len(ref.FieldMappings) > 0 {
		if observed.Write != nil && observed.Write.GraphQL != nil {
			issues = append(issues, sourceLaneGraphQLVariablesContract(facts, ref, *observed.Write.GraphQL)...)
		} else if observed.Write != nil || observed.REST != nil {
			add(sourceLaneBodyContract(facts, ref, observed), ptr)
		}
		if observed.Stream != nil {
			if ref.SourceSchema == nil {
				add("target_response_contract_unverified", ptr)
			}
		} else if len(scopeOnly) == 0 && !sourceLaneNoResponseBody(facts) {
			add("target_response_contract_unverified", ptr)
		}
	}
	return issues
}

func sourceLaneGraphQLContract(facts sourceFacts, a sourceSemanticAnnotation, ref sourceLaneTargetRef, target sourceLaneTypedTarget) []sourceLaneBindingIssue {
	var document, name string
	if target.GraphQL != nil {
		document = target.GraphQL.Document
		name = target.GraphQL.OperationName
	} else if target.Stream != nil && target.Stream.GraphQL != nil {
		document = target.Stream.GraphQL.Document
		name = target.Stream.GraphQL.OperationName
	} else if target.Write != nil && target.Write.GraphQL != nil {
		document = target.Write.GraphQL.Document
		name = target.Write.GraphQL.OperationName
	} else {
		return nil
	}
	issues := []sourceLaneBindingIssue{}
	add := func(code, pointer string) { issues = append(issues, sourceLaneBindingIssue{code, pointer}) }
	refs := a.GraphQL
	if refs == nil {
		refs = &sourceLaneGraphQLRefs{}
	}
	for _, field := range []struct {
		ref                      *sourceFactRef
		value, missing, mismatch string
	}{{refs.Document, document, "source_graphql_document_unavailable", "target_graphql_document_mismatch"}, {refs.OperationName, name, "source_graphql_operation_unavailable", "target_graphql_operation_mismatch"}} {
		if field.ref == nil {
			add(field.missing, ref.Artifact+"#"+ref.Pointer)
			continue
		}
		raw, code := sourceLaneCitedValue(facts, *field.ref)
		if code != "" {
			add(code, field.ref.Pointer)
			continue
		}
		var value string
		if json.Unmarshal(raw, &value) != nil || value != field.value {
			add(field.mismatch, field.ref.Pointer)
		}
	}
	if refs.RequestSchema == nil {
		add("source_graphql_request_schema_unavailable", ref.Artifact+"#"+ref.Pointer)
	} else if ref.SourceSchema == nil || *ref.SourceSchema != *refs.RequestSchema {
		add("source_binding_scope_mismatch", refs.RequestSchema.Pointer)
	}
	return issues
}

func sourceLaneGraphQLVariablesContract(facts sourceFacts, ref sourceLaneTargetRef, graphql engine.GraphQLRequestSpec) []sourceLaneBindingIssue {
	issues := []sourceLaneBindingIssue{}
	add := func(code string) {
		issues = append(issues, sourceLaneBindingIssue{code, ref.Artifact + "#" + ref.Pointer})
	}
	if ref.SourceSchema == nil {
		add("source_graphql_request_schema_unavailable")
		return issues
	}
	raw, code := sourceLaneCitedValue(facts, *ref.SourceSchema)
	if code != "" {
		add(code)
		return issues
	}
	root, ok := sourceResolveObject(facts, raw, map[string]bool{}, 0)
	if !ok {
		add("source_schema_unverified")
		return issues
	}
	shape, state := sourceLaneCheckedProjectionNode(root)
	var props = shape.Properties
	if state != sourceLaneLineageSupported || sourceLocalShapeEvidence(root).object() != sourceLocalObjectEstablished || props == nil {
		add("source_schema_unverified")
		return issues
	}
	if len(props) != len(graphql.Variables) {
		add("target_graphql_variable_mismatch")
	}
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		value, exists := graphql.Variables[name]
		if !exists {
			add("target_graphql_variable_mismatch")
			continue
		}
		text, ok := value.(string)
		if !ok {
			add("target_graphql_variable_unverified")
			continue
		}
		kind, pointer, code := sourceLaneTemplateCoordinate(text)
		if code != "" {
			add("target_graphql_variable_unverified")
			continue
		}
		matched := false
		unverified := false
		for _, m := range ref.FieldMappings {
			projection, c := sourceLaneSchemaProjection(facts, *ref.SourceSchema, m.Source.Pointer)
			if c == "source_schema_unverified" {
				unverified = true
			}
			if c == "" && len(projection.Path) == 1 && projection.Path[0] == name && m.Target.Pointer != nil && m.Target.Kind == kind && *m.Target.Pointer == pointer {
				matched = true
			}
		}
		if !matched && unverified {
			add("source_schema_unverified")
		} else if !matched {
			add("target_graphql_variable_mismatch")
		}
	}
	return issues
}

func sourceLaneCheckAggregate(key sourceOperationKey, facts sourceFacts, a sourceSemanticAnnotation, ref sourceLaneTargetRef, raw []byte) []sourceLaneBindingIssue {
	issues := []sourceLaneBindingIssue{}
	ptr := ref.Artifact + "#" + ref.Pointer
	add := func(code string) { issues = append(issues, sourceLaneBindingIssue{code, ptr}) }
	if facts.bindings == nil {
		add("target_contract_unverified")
		return issues
	}
	descriptor, ok := facts.bindings.Canonical[key.Connector]
	if !ok {
		add("canonical_binding_absent")
		return issues
	}
	var source *vNextCanonicalOperation
	for i := range descriptor.Graph.Operations {
		if descriptor.Graph.Operations[i].ID == ref.CanonicalID {
			source = &descriptor.Graph.Operations[i]
			break
		}
	}
	if source == nil {
		add("canonical_provenance_mismatch")
		return issues
	}
	bundle, ok := facts.bindings.Bundles[key.Connector]
	if !ok || bundle.Identity.Digest != descriptor.Staged.Identity.Digest {
		add("execution_generation_mismatch")
	}
	if ref.Generation != descriptor.Staged.Identity.Digest {
		add("canonical_generation_mismatch")
	}
	if ref.Kind == "sync_transport" {
		return append(issues, sourceLaneSyncContract(facts, ref, descriptor, *source, raw)...)
	}
	if ref.Kind == "canonical_operation" {
		if ref.ID != ref.CanonicalID || ref.Pointer != "/operations/"+strconv.Itoa(source.Index) || ref.CanonicalPointer != "/operations/"+strconv.Itoa(source.CanonicalIndex) {
			add("canonical_provenance_mismatch")
		}
		selected, err := sourceJSONPointer(raw, ref.Pointer)
		var authored vNextOperationDescriptor
		if err != nil || decodeStrictJSON(selected, &authored) != nil || source.Index < 0 || source.Index >= len(descriptor.Operations) || !vNextJSONEquivalent(authored, descriptor.Operations[source.Index]) {
			add("canonical_authoring_mismatch")
		}
	} else {
		registry := ""
		switch ref.SchemaRole {
		case sourceLaneSchemaRequest:
			registry = source.SchemaRefs.Request
		case sourceLaneSchemaRecord:
			registry = source.SchemaRefs.Record
		case sourceLaneSchemaResponse:
			registry = source.SchemaRefs.Response
		}
		if registry != ref.ID || registry == "" || ref.CanonicalPointer != vNextProvenanceOperationPointer(*source, "schema_refs", string(ref.SchemaRole)) {
			add("canonical_provenance_mismatch")
		}
		if !bytes.Equal(raw, descriptor.Staged.Outputs[ref.ID]) || !bytes.Equal(raw, facts.bindings.Artifacts[ref.Artifact]) {
			add("canonical_artifact_mismatch")
		}
	}
	children := []sourceLaneTargetRef{}
	child := func(kind, id string) {
		r := canonicalSourceLaneTargetRef(ref)
		r.Kind = kind
		r.ID = id
		r.Artifact = "internal/connectors/defs/" + key.Connector + "/" + map[string]string{"operation": "operations.json", "write": "writes.json", "stream": "streams.json"}[kind]
		r.CanonicalPointer = vNextProvenanceOperationPointer(*source, kind)
		childRaw := facts.bindings.Artifacts[r.Artifact]
		r.ArtifactSHA256 = sourceBytesHash(childRaw)
		r.Pointer, _ = sourceLaneTargetPointer(childRaw, r)
		children = append(children, r)
	}
	if source.Stream != nil && (ref.SchemaRole == sourceLaneSchemaRecord || ref.SchemaRole == sourceLaneSchemaResponse || ref.Lane == "etl" || ref.Lane == "direct_read") {
		child("stream", source.Stream.Spec.Name)
	}
	if source.Write != nil && (ref.SchemaRole == sourceLaneSchemaRequest || ref.Lane == "direct_write" || ref.Lane == "reverse_etl") {
		child("write", source.Write.Spec.Name)
	}
	if source.Operation != nil && (ref.SchemaRole == sourceLaneSchemaRequest || ref.Kind == "canonical_operation") {
		child("operation", source.Operation.Spec.ID)
	}
	if len(children) == 0 {
		add("target_schema_consumer_unverified")
	}
	for _, r := range children {
		childRaw := facts.bindings.Artifacts[r.Artifact]
		node, err := sourceJSONPointer(childRaw, r.Pointer)
		if err != nil {
			add("target_pointer_mismatch")
			continue
		}
		if ref.Kind == "schema" {
			observed, code := sourceLaneObserveTypedTarget(r, node, facts.bindings)
			if code != "" {
				add(code)
				continue
			}
			effective, code := sourceLaneEffectiveSchema(observed, bundle, ref.SchemaRole)
			if code != "" {
				add(code)
			} else if !vNextJSONEquivalent(json.RawMessage(raw), effective) {
				add("target_schema_mismatch")
			}
			issues = append(issues, sourceLaneCheckPresent(key, facts, a, r, childRaw, node, true)...)
		} else {
			issues = append(issues, sourceLaneCheckPresent(key, facts, a, r, childRaw, node)...)
		}
	}
	return issues
}

func sourceLaneSyncContract(facts sourceFacts, ref sourceLaneTargetRef, descriptor vNextCanonicalDescriptor, source vNextCanonicalOperation, raw []byte) []sourceLaneBindingIssue {
	issues := []sourceLaneBindingIssue{}
	ptr := ref.Artifact + "#" + ref.Pointer
	add := func(code string) { issues = append(issues, sourceLaneBindingIssue{code, ptr}) }
	if ref.Lane != "sync_transport" {
		add("target_lane_mismatch")
	}
	if ref.CanonicalPointer != "/execution/sync_transport.json"+ref.Pointer {
		add("canonical_provenance_mismatch")
	}
	if !bytes.Equal(raw, descriptor.Staged.Outputs["sync_transport.json"]) || !vNextJSONEquivalent(json.RawMessage(raw), descriptor.Execution["sync_transport.json"]) {
		add("canonical_artifact_mismatch")
	}
	bundle := facts.bindings.Bundles[ref.Connector]
	if bundle.SyncTransport == nil {
		add("target_transport_contract_unverified")
		return issues
	}
	node, err := sourceJSONPointer(raw, ref.Pointer)
	if err != nil {
		add("target_pointer_mismatch")
		return issues
	}
	var childKind, childID string
	switch ref.Pointer {
	case "/source_transport":
		var role connectors.SourceTransportDescriptor
		if decodeStrictJSON(node, &role) != nil || role.Validate() != nil {
			add("target_shape_invalid")
			return issues
		}
		if role.Executor.ID != ref.ID || !reflect.DeepEqual(bundle.SyncTransport.Source, &role) {
			add("target_identity_mismatch")
		}
		if source.Stream == nil || !sourceLaneContains(role.EligibleStreams, source.Stream.Spec.Name) {
			add("target_transport_eligibility_mismatch")
		} else {
			childKind = "stream"
			childID = source.Stream.Spec.Name
		}
		found := false
		for _, sync := range descriptor.Staged.Sync {
			if sync.SourceID != source.ID {
				continue
			}
			found = true
			if sync.FieldPath != "/operations/"+strconv.Itoa(source.Index)+"/stream" {
				add("target_transport_coordinate_unverified")
			}
			if sync.Result.Validate() != nil || sync.Result.Plan == nil {
				add("target_transport_plan_mismatch")
				continue
			}
			plan := sync.Result.Plan
			if source.Stream == nil || plan.Source.ID != source.Stream.Spec.Name || plan.GenerationDigest != descriptor.Staged.Identity.Digest || !vNextPlanUsesManifestSource(*plan, descriptor.Staged.Manifest) {
				add("target_transport_plan_mismatch")
			}
			mode := false
			for _, value := range role.Modes {
				mode = mode || value == plan.Mode
			}
			if !mode {
				add("target_transport_mode_mismatch")
			}
		}
		if !found {
			add("target_transport_plan_unverified")
		}
	case "/destination_transport":
		var role connectors.DestinationTransportDescriptor
		if decodeStrictJSON(node, &role) != nil || role.Validate() != nil {
			add("target_shape_invalid")
			return issues
		}
		if role.Executor.ID != ref.ID || !reflect.DeepEqual(bundle.SyncTransport.Destination, &role) {
			add("target_identity_mismatch")
		}
		if source.Write == nil || !sourceLaneContains(role.EligibleActions, source.Write.Spec.Name) {
			add("target_transport_eligibility_mismatch")
		} else {
			childKind = "write"
			childID = source.Write.Spec.Name
		}
		if role.Acknowledgement != connectors.TransportAcknowledgementDurableWarehouse {
			add("target_transport_acknowledgement_unverified")
		}
		add("target_transport_plan_unverified")
	default:
		add("target_pointer_mismatch")
	}
	if childID != "" {
		matches := 0
		for _, p := range descriptor.Staged.Provenance {
			if p.SourceID == source.ID && p.FieldPath == vNextProvenanceOperationPointer(source, childKind) && p.TargetKind == childKind && p.TargetID == childID {
				matches++
			}
		}
		if matches != 1 {
			add("canonical_provenance_mismatch")
		}
	}
	// Retained callback/event presence cannot prove descriptor delivery/mode/ack
	// guarantees, and a syncplan executor is not a registered transport factory.
	add("source_transport_contract_unverified")
	add("target_transport_executor_unverified")
	return issues
}

func sourceLaneNoResponseBody(facts sourceFacts) bool {
	var responses map[string]json.RawMessage
	if decodeSourceJSON(facts.Groups["responses"], &responses) != nil {
		return false
	}
	success := false
	for status, raw := range responses {
		if len(status) != 3 || status[0] != '2' {
			continue
		}
		success = true
		node, ok := sourceResolveObject(facts, raw, map[string]bool{}, 0)
		if !ok {
			return false
		}
		if status != "204" && status != "205" {
			return false
		}
		if len(node["content"]) > 0 || len(node["schema"]) > 0 {
			return false
		}
	}
	return success
}

func sourceLaneRouteContract(facts sourceFacts, ref sourceLaneTargetRef, target sourceLaneTypedTarget, bundle engine.Bundle) string {
	if facts.Protocol != "rest" && facts.Protocol != "graphql" {
		return "target_contract_unverified"
	}
	if target.Binding.TransportMethod != facts.Method {
		return "target_semantics_mismatch"
	}
	if !strings.Contains(target.RawPath, "{{") {
		if !sourceLanePathsEqual(facts, target.RawPath) {
			return "target_semantics_mismatch"
		}
		return ""
	}
	source := strings.Split(facts.Path, "/")
	actual := strings.Split(target.RawPath, "/")
	if len(source) != len(actual) {
		if sourceLanePathsEqual(facts, strings.TrimPrefix(facts.Path, "/api/v4")) {
			source = strings.Split(strings.TrimPrefix(facts.Path, "/api/v4"), "/")
		}
	}
	if len(source) != len(actual) {
		return "target_semantics_mismatch"
	}
	for i, segment := range actual {
		if !strings.Contains(segment, "{{") {
			if segment != source[i] {
				return "target_semantics_mismatch"
			}
			continue
		}
		kind, coordinate, code := sourceLaneTemplateCoordinate(segment)
		if code != "" {
			return code
		}
		if !strings.HasPrefix(source[i], "{") || !strings.HasSuffix(source[i], "}") {
			return "target_semantics_mismatch"
		}
		name := source[i][1 : len(source[i])-1]
		matched := false
		for _, p := range facts.Parameters {
			if p.In != "path" || p.Name != name {
				continue
			}
			for _, m := range ref.FieldMappings {
				if m.Source == p.Ref && m.Target.Pointer != nil && m.Target.Kind == kind && *m.Target.Pointer == coordinate {
					matched = true
				}
			}
		}
		if !matched {
			if len(ref.FieldMappings) == 0 {
				return "target_path_projection_unverified"
			}
			return "target_path_projection_mismatch"
		}
		if kind == sourceLaneFieldSchema && target.Write == nil {
			return "target_path_projection_mismatch"
		}
	}
	return ""
}

func sourceLaneTemplateCoordinate(value string) (sourceLaneFieldTargetKind, string, string) {
	if !strings.HasPrefix(value, "{{") || !strings.HasSuffix(value, "}}") {
		return "", "", "target_path_projection_unverified"
	}
	inner := strings.TrimSpace(value[2 : len(value)-2])
	parts := strings.Split(inner, "|")
	if len(parts) > 2 || len(parts) == 2 && strings.TrimSpace(parts[1]) != "urlencode" {
		return "", "", "target_path_projection_unverified"
	}
	fields := strings.Split(strings.TrimSpace(parts[0]), ".")
	if len(fields) < 2 || fields[0] != "record" && fields[0] != "config" {
		return "", "", "target_path_projection_unverified"
	}
	kind := sourceLaneFieldSchema
	if fields[0] == "config" {
		kind = sourceLaneFieldConfig
		if len(fields) != 2 {
			return "", "", "target_path_projection_unverified"
		}
	}
	pointer := ""
	for _, field := range fields[1:] {
		if field == "" || strings.ContainsAny(field, " {}[]/\\\t\r\n") {
			return "", "", "target_path_projection_unverified"
		}
		pointer += "/properties/" + escapeSourcePointer(field)
	}
	return kind, pointer, ""
}

func sourceLaneProjectionContract(facts sourceFacts, a sourceSemanticAnnotation, ref sourceLaneTargetRef, target sourceLaneTypedTarget, bundle engine.Bundle) []sourceLaneBindingIssue {
	issues := []sourceLaneBindingIssue{}
	add := func(code, pointer string) {
		if code != "" {
			issues = append(issues, sourceLaneBindingIssue{code, pointer})
		}
	}
	ptr := ref.Artifact + "#" + ref.Pointer
	effective, code := sourceLaneEffectiveSchema(target, bundle, ref.SchemaRole)
	add(code, ptr)
	var anchor json.RawMessage
	if ref.SourceSchema != nil {
		var media string
		anchor, media, code = sourceLaneSchemaAnchor(facts, a, ref)
		add(code, ref.SourceSchema.Pointer)
		if code == "" && media != "application/json" {
			add("target_media_contract_unverified", ref.SourceSchema.Pointer)
		}
	}
	for _, m := range ref.FieldMappings {
		raw, code := sourceLaneCitedValue(facts, m.Source)
		if code != "" {
			add(code, m.Source.Pointer)
			continue
		}
		var parameter *sourceParameterFact
		for i := range facts.Parameters {
			if facts.Parameters[i].Ref == m.Source {
				parameter = &facts.Parameters[i]
				break
			}
		}
		var source sourceLaneProjection
		if parameter != nil {
			node, ok := sourceResolveObject(facts, parameter.Node, map[string]bool{}, 0)
			if !ok {
				add("source_parameter_contract_unverified", m.Source.Pointer)
				continue
			}
			source = sourceLaneProjection{Raw: node["schema"], Required: []bool{parameter.Required}}
		} else {
			if ref.SourceSchema == nil {
				add("source_binding_scope_mismatch", m.Source.Pointer)
				continue
			}
			source, code = sourceLaneSchemaProjection(facts, *ref.SourceSchema, m.Source.Pointer)
			if code != "" {
				add(code, m.Source.Pointer)
				continue
			}
			source.Raw = raw
		}
		if m.Target.Pointer == nil {
			add("target_projection_shape_invalid", m.Source.Pointer)
			continue
		}
		if m.Target.Kind == sourceLaneFieldParameter {
			if parameter == nil || target.REST == nil {
				add("target_parameter_mismatch", m.Source.Pointer)
				continue
			}
			parts := strings.Split(*m.Target.Pointer, "/")
			index, _ := strconv.Atoi(parts[len(parts)-1])
			parameters := target.REST.Parameters
			if len(parts) > 1 && parts[1] == "pagination_parameters" {
				parameters = target.REST.PaginationParameters
			}
			if index < 0 || index >= len(parameters) {
				add("target_parameter_mismatch", m.Source.Pointer)
				continue
			}
			copy := facts
			copy.Parameters = []sourceParameterFact{*parameter}
			add(sourceLaneRESTParameterContract(copy, engine.RESTOperationSpec{Parameters: []engine.OperationParameter{parameters[index]}}), m.Source.Pointer)
			continue
		}
		root := effective
		if m.Target.Kind == sourceLaneFieldConfig {
			root = bundle.RawSpec
		}
		if len(root) == 0 {
			continue
		}
		actual, code := sourceLaneTargetProjection(root, *m.Target.Pointer)
		if code != "" {
			add(code, m.Source.Pointer)
			continue
		}
		add(sourceLaneSchemaCompare(facts, source.Raw, actual.Raw), m.Source.Pointer)
		if parameter != nil {
			if target.Write == nil || parameter.In != "path" {
				add("target_parameter_projection_unverified", m.Source.Pointer)
				continue
			}
			if len(actual.Required) == 0 {
				add("target_parameter_mismatch", m.Source.Pointer)
			} else {
				required := true
				for _, v := range actual.Required {
					required = required && v
				}
				if parameter.Required != required {
					add("target_schema_mismatch", m.Source.Pointer)
				}
			}
			if m.Target.Kind == sourceLaneFieldSchema && (len(actual.Path) != 1 || !sourceLaneContains(target.Write.PathFields, actual.Path[0])) {
				add("target_path_projection_mismatch", m.Source.Pointer)
			}
			continue
		}
		if ref.SchemaRole == sourceLaneSchemaRequest {
			if m.Target.Kind != sourceLaneFieldSchema {
				add("target_body_projection_unverified", m.Source.Pointer)
				continue
			}
			if target.Write != nil && target.Write.BodyType == "json_array" {
				coordinate := []string{}
				coordinate = append(coordinate, strings.Split(target.Write.BodyField, ".")...)
				if len(source.Path) != 0 || !reflect.DeepEqual(actual.Path, coordinate) {
					add("target_body_projection_mismatch", m.Source.Pointer)
				}
			} else if target.Write != nil && target.Write.GraphQL != nil {
				// Variable placement is checked against the actual typed Variables
				// consumer, permitting only an explicitly declared input mapping.
				if !reflect.DeepEqual(source.Required, actual.Required) {
					add("target_schema_mismatch", m.Source.Pointer)
				}
			} else if !reflect.DeepEqual(source.Path, actual.Path) || !reflect.DeepEqual(source.Required, actual.Required) {
				add("target_body_projection_mismatch", m.Source.Pointer)
			}
		} else {
			if target.Stream == nil || m.Target.Kind != sourceLaneFieldSchema {
				add("target_schema_role_mismatch", m.Source.Pointer)
				continue
			}
			coordinate, code := sourceLaneRecordCoordinate(facts, anchor, *target.Stream)
			add(code, m.Source.Pointer)
			if code == "" && (!reflect.DeepEqual(source.Path, coordinate) || len(actual.Path) != 0) {
				add("target_record_projection_mismatch", m.Source.Pointer)
			}
		}
	}
	if len(ref.FieldMappings) == 0 {
		add("target_field_contract_unverified", ptr)
	}
	return issues
}

func sourceLaneRecordCoordinate(facts sourceFacts, anchor json.RawMessage, stream engine.StreamSpec) ([]string, string) {
	if stream.Records.Filter != nil || stream.Records.KeyedObject || stream.Records.WrapField != "" || stream.ArrayZipProjection != nil || len(stream.ComputedFields) > 0 || len(stream.ResponseFields) > 0 || stream.Projection == "passthrough" {
		return nil, "target_record_projection_unverified"
	}
	raw := anchor
	coordinate := []string{}
	fields := []string{}
	if stream.Records.Path != "" && stream.Records.Path != "." {
		fields = strings.Split(stream.Records.Path, ".")
	}
	for index := 0; index <= len(fields); index++ {
		node, ok := sourceResolveObject(facts, raw, map[string]bool{}, 0)
		if !ok {
			return nil, "source_schema_unverified"
		}
		shape, state := sourceLaneCheckedProjectionNode(node)
		if state == sourceLaneLineageUnverified {
			return nil, "source_schema_unverified"
		}
		if state == sourceLaneLineageContradictory {
			return nil, "target_record_projection_mismatch"
		}
		if index < len(fields) {
			if sourceLocalShapeEvidence(node).object() != sourceLocalObjectEstablished || shape.Properties[fields[index]] == nil {
				return nil, "target_record_projection_mismatch"
			}
			raw = shape.Properties[fields[index]]
			coordinate = append(coordinate, fields[index])
			continue
		}
		if shape.Type == "array" {
			if stream.Records.SingleObject {
				return nil, "target_record_projection_mismatch"
			}
			if sourceLocalShapeEvidence(node).arrayCoverage() != sourceLocalArrayUniform {
				return nil, "target_record_projection_unverified"
			}
			coordinate = append(coordinate, "[]")
		} else if sourceLocalShapeEvidence(node).object() != sourceLocalObjectEstablished {
			return nil, "target_record_projection_unverified"
		}
	}
	return coordinate, ""
}

func sourceLaneBodyContract(facts sourceFacts, ref sourceLaneTargetRef, target sourceLaneTypedTarget) string {
	body := facts.Groups["request_body"]
	if len(body) == 0 || string(body) == "null" {
		if target.REST != nil && (len(target.REST.Body) > 0 || len(target.REST.BodySchema) > 0) {
			return "target_request_contract_unverified"
		}
		if target.Write != nil && target.Write.BodyType != "none" {
			return "target_request_contract_unverified"
		}
		return ""
	}
	if ref.SourceSchema == nil || ref.SchemaRole != sourceLaneSchemaRequest {
		return "target_request_contract_unverified"
	}
	anchor, _, code := sourceLaneSchemaAnchor(facts, sourceSemanticAnnotation{}, ref)
	if code != "" {
		return code
	}
	bodyNode, ok := sourceResolveObject(facts, body, map[string]bool{}, 0)
	if !ok {
		return "source_schema_unverified"
	}
	var content map[string]json.RawMessage
	if json.Unmarshal(bodyNode["content"], &content) != nil {
		return "source_schema_unverified"
	}
	if len(content) != 1 {
		return "target_request_scopes_unverified"
	}
	if target.REST != nil {
		media := target.REST.ContentType
		if media == "" {
			media = "application/json"
		}
		if _, exists := content[media]; !exists {
			return "target_media_mismatch"
		}
		if len(target.REST.Body) > 0 {
			return "target_body_projection_unverified"
		}
		return sourceLaneSchemaCompare(facts, anchor, target.REST.BodySchema)
	}
	if target.Write == nil {
		return "target_request_contract_unverified"
	}
	w := target.Write
	if w.DynamicFields != nil {
		return "target_body_projection_unverified"
	}
	if w.BodyType == "json_array" {
		pointer := ""
		for _, field := range strings.Split(w.BodyField, ".") {
			pointer += "/properties/" + escapeSourcePointer(field)
		}
		projection, code := sourceLaneTargetProjection(w.RecordSchema, pointer)
		if code != "" {
			return code
		}
		required := true
		for _, present := range projection.Required {
			required = required && present
		}
		var bodyRequired bool
		_ = json.Unmarshal(bodyNode["required"], &bodyRequired)
		if bodyRequired && !required {
			return "target_request_requiredness_mismatch"
		}
		if code := sourceLaneSchemaCompare(facts, anchor, projection.Raw); code != "" {
			return code
		}
		if len(w.BodySchema) > 0 {
			if code := sourceLaneSchemaCompare(facts, anchor, w.BodySchema); code != "" {
				return code
			}
		}
		covered := false
		for _, m := range ref.FieldMappings {
			covered = covered || m.Source == *ref.SourceSchema && m.Target.Kind == sourceLaneFieldSchema && m.Target.Pointer != nil && *m.Target.Pointer == pointer
		}
		if !covered {
			return "target_body_coverage_unverified"
		}
		return ""
	}
	if w.BodyType != "json" && w.BodyType != "" {
		return "target_body_projection_unverified"
	}
	if _, exists := content["application/json"]; !exists {
		return "target_media_mismatch"
	}
	var required bool
	_ = json.Unmarshal(bodyNode["required"], &required)
	if required != w.BodyRequired {
		return "target_request_requiredness_mismatch"
	}
	var record map[string]json.RawMessage
	if decodeSourceJSON(w.RecordSchema, &record) != nil {
		return "target_schema_unverified"
	}
	var props map[string]json.RawMessage
	_ = json.Unmarshal(record["properties"], &props)
	var requiredFields []string
	_ = json.Unmarshal(record["required"], &requiredFields)
	emitted := map[string]json.RawMessage{}
	keptRequired := []string{}
	for name, value := range props {
		include := !sourceLaneContains(w.PathFields, name)
		if len(w.BodyFields) > 0 {
			include = sourceLaneContains(w.BodyFields, name)
		}
		if include {
			if sourceLaneContains(w.PathFields, name) {
				return "target_body_projection_mismatch"
			}
			emitted[name] = value
			if sourceLaneContains(requiredFields, name) {
				keptRequired = append(keptRequired, name)
			}
		}
	}
	for _, name := range w.BodyFields {
		if _, exists := props[name]; !exists {
			return "target_body_projection_mismatch"
		}
	}
	record["properties"], _ = json.Marshal(emitted)
	record["required"], _ = json.Marshal(keptRequired)
	projected, _ := json.Marshal(record)
	if code := sourceLaneSchemaCompare(facts, anchor, projected); code != "" {
		if code == "target_schema_mismatch" {
			return "target_body_projection_mismatch"
		}
		return code
	}
	// Whole-input or top-level subtree mappings must account for every emitted
	// value. A partial leaf does not silently cover an unvalidated sibling.
	for name := range emitted {
		covered := false
		for _, m := range ref.FieldMappings {
			if m.Target.Kind == sourceLaneFieldSchema && m.Target.Pointer != nil && (*m.Target.Pointer == "" || *m.Target.Pointer == "/properties/"+escapeSourcePointer(name)) {
				covered = true
			}
		}
		if !covered {
			return "target_body_coverage_unverified"
		}
	}
	return ""
}
