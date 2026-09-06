package main

import (
	"bytes"
	"path"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

// sourceLaneBindingInputs carries read-only, already observed artifacts into
// source classification. Loading and admission remain separate from source
// membership; an empty collection cannot remove a provider row.
type sourceLaneBindingInputs struct {
	Artifacts map[string][]byte
	Canonical map[string]vNextCanonicalDescriptor
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
			node, err := sourceJSONPointer(raw, ref.Pointer)
			if err != nil {
				absent()
				continue
			}
			if ref.Kind != "operation" {
				add("target_contract_unverified", sourceClaimSeverity(group.claimed))
				continue
			}
			var op engine.OperationSpec
			if err := decodeStrictJSON(node, &op); err != nil {
				add("target_shape_invalid", "error")
				continue
			}
			if op.ID != ref.ID {
				add("target_identity_mismatch", "error")
				continue
			}
			if op.REST == nil || facts.Protocol != "rest" {
				add("target_contract_unverified", sourceClaimSeverity(group.claimed))
				continue
			}
			if op.REST.Method != facts.Method || op.REST.Path != facts.Path {
				add("target_semantics_mismatch", "error")
				continue
			}
			if (ref.Lane == "direct_read" && op.Kind != "rest_read") || (ref.Lane == "direct_write" && op.Kind != "rest_write") || (ref.Lane != "direct_read" && ref.Lane != "direct_write") {
				add("target_lane_mismatch", "error")
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
			if len(facts.Parameters) > 0 || len(op.REST.Parameters) > 0 || len(op.REST.PaginationParameters) > 0 {
				add("target_parameter_contract_unverified", sourceClaimSeverity(group.claimed))
				continue
			}
			if body := facts.Groups["request_body"]; (len(body) > 0 && string(body) != "null") || len(op.REST.Body) > 0 || len(op.REST.BodySchema) > 0 {
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
