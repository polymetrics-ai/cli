package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"math/big"
	"os"
	"path"
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
	Bundles      map[string]engine.Bundle
	Observations []sourceLaneBindingObservation
	Artifacts    map[string][]byte
	Canonical    map[string]vNextCanonicalDescriptor
}

// resolveSourceLaneBindings reports each reference independently. Artifact
// existence never supplies membership or behavioral proof.
func resolveSourceLaneBindings(key sourceOperationKey, facts sourceFacts, annotation sourceSemanticAnnotation, cells []sourceLaneCell) []sourceLaneCell {
	for _, group := range []struct {
		refs    []sourceLaneTargetRef
		claimed bool
	}{{annotation.IntendedBindings, false}, {annotation.MaterializedBindings, true}} {
		for _, ref := range group.refs {
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
			observed, code := sourceLaneObserveTypedTarget(ref, node, facts.bindings)
			if code != "" {
				severity := sourceClaimSeverity(group.claimed)
				if code == "target_shape_invalid" || code == "target_lane_mismatch" || code == "target_identity_mismatch" {
					severity = "error"
				}
				add(code, severity)
				continue
			}
			if facts.Protocol != "rest" {
				add("target_contract_unverified", sourceClaimSeverity(group.claimed))
				continue
			}
			if strings.Contains(observed.RawPath, "{{") || observed.RawPath != observed.Binding.TransportPath {
				add("target_path_projection_unverified", sourceClaimSeverity(group.claimed))
				continue
			}
			if observed.Binding.TransportMethod != facts.Method || !sourceLanePathsEqual(facts, observed.RawPath) {
				add("target_semantics_mismatch", "error")
				continue
			}
			if facts.bindings == nil || ref.CanonicalID == "" || ref.CanonicalPointer == "" || ref.Generation == "" {
				add("target_contract_unverified", sourceClaimSeverity(group.claimed))
				continue
			}
			descriptor, exists := facts.bindings.Canonical[key.Connector]
			if !exists {
				add("canonical_binding_absent", sourceClaimSeverity(group.claimed))
				continue
			}
			if descriptor.Connector != key.Connector || descriptor.Staged.Identity.Digest != ref.Generation {
				add("canonical_generation_mismatch", "error")
				continue
			}
			actual, exists := facts.bindings.Bundles[key.Connector]
			if !exists || actual.Identity.Digest != descriptor.Staged.Identity.Digest {
				add("execution_generation_mismatch", "error")
				continue
			}
			relative := strings.TrimPrefix(ref.Artifact, "internal/connectors/defs/"+key.Connector+"/")
			if !bytes.Equal(raw, descriptor.Staged.Outputs[relative]) {
				add("canonical_artifact_mismatch", "error")
				continue
			}
			matches := 0
			for _, provenance := range descriptor.Staged.Provenance {
				if provenance.SourceID == ref.CanonicalID && provenance.FieldPath == ref.CanonicalPointer && provenance.TargetKind == ref.Kind && provenance.TargetID == ref.ID {
					matches++
				}
			}
			if matches != 1 {
				add("canonical_provenance_mismatch", "error")
				continue
			}
			// Admission proves the current canonical target identity. The
			// provider-to-target parameter/body/response join remains separate.
			if observed.REST == nil {
				add("target_field_contract_unverified", sourceClaimSeverity(group.claimed))
				continue
			}
			if code := sourceLaneRESTParameterContract(facts, *observed.REST); code != "" {
				severity := sourceClaimSeverity(group.claimed)
				if code == "target_parameter_mismatch" {
					severity = "error"
				}
				add(code, severity)
				continue
			}
			if body := facts.Groups["request_body"]; (len(body) > 0 && string(body) != "null") || len(observed.REST.Body) > 0 || len(observed.REST.BodySchema) > 0 {
				add("target_request_contract_unverified", sourceClaimSeverity(group.claimed))
				continue
			}
			add("target_response_contract_unverified", sourceClaimSeverity(group.claimed))
		}
	}
	return cells
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
		text := string(number)
		negative := strings.HasPrefix(text, "-")
		text = strings.TrimPrefix(text, "-")
		coefficient, exponentText, hasExponent := strings.Cut(strings.ToLower(text), "e")
		exponent := new(big.Int)
		if hasExponent {
			if _, ok := exponent.SetString(exponentText, 10); !ok {
				return "", false
			}
		}
		if dot := strings.IndexByte(coefficient, '.'); dot >= 0 {
			exponent.Sub(exponent, big.NewInt(int64(len(coefficient)-dot-1)))
			coefficient = coefficient[:dot] + coefficient[dot+1:]
		}
		coefficient = strings.TrimLeft(coefficient, "0")
		if coefficient == "" {
			return "0", true
		}
		shortened := strings.TrimRight(coefficient, "0")
		exponent.Add(exponent, big.NewInt(int64(len(coefficient)-len(shortened))))
		sign := ""
		if negative {
			sign = "-"
		}
		return sign + shortened + "e" + exponent.String(), true
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
	result = sourceLaneBindingInputs{Artifacts: map[string][]byte{}, Canonical: map[string]vNextCanonicalDescriptor{}, Bundles: map[string]engine.Bundle{}, Observations: []sourceLaneBindingObservation{}}
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
	collection, identity := "", ""
	switch ref.Kind {
	case "operation":
		collection, identity = "operations", "id"
	case "write":
		collection, identity = "actions", "name"
	case "stream":
		collection, identity = "streams", "name"
	case "command":
		collection, identity = "commands", "path"
	default:
		return "", "target_contract_unverified"
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
		if stream.GraphQL != nil {
			return result, "target_contract_unverified"
		}
		command.Stream = ref.ID
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
