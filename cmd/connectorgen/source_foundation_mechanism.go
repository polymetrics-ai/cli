package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

type sourceFoundationSelector struct {
	Path           string `json:"path"`
	Pointer        string `json:"pointer"`
	ArtifactSHA256 string `json:"artifact_sha256"`
	ValueSHA256    string `json:"value_sha256"`
}

type sourceFoundationMechanismFit struct {
	ProofID          string                      `json:"proof_id"`
	AtlasID          string                      `json:"atlas_id"`
	Contract         sourceFoundationContractRef `json:"contract"`
	Binding          sourceLaneTargetRef         `json:"binding"`
	DeclarationState string                      `json:"declaration_state"`
	Mechanism        string                      `json:"mechanism"`
	SourceRefs       []sourceFactRef             `json:"source_refs"`
	Selectors        []sourceFoundationSelector  `json:"selectors"`
}

func sourceFoundationKnownMechanism(mechanism string) bool {
	switch mechanism {
	case "", "base_check_exact_status", "rest_write_structured_json_body", "rest_write_static_api_key_header":
		return true
	}
	return false
}

// The source-owned review binds the exact assertion to its subject. A caller
// cannot relabel a current Check or vocabulary observation as an operation fit.
func sourceFoundationMechanismRecord(record sourceFoundationProofRecord, mechanism string) bool {
	if mechanism == "" {
		return true
	}
	if record.AtlasID != "runtime.direct-execution.v1" {
		return false
	}
	switch mechanism {
	case "base_check_exact_status":
		return record.Test.Selected == "TestCheck_ExactSuccessStatusesPreserveDeclaredOutcome/bad_rejects_undeclared_200"
	case "rest_write_structured_json_body":
		return record.Test.Symbol == "TestOperationDirectWriteStructuredRESTBodyIsExactAndPreviewBound" && record.Contract.Pointer == "/supported_contracts/request_response_shapes/12"
	case "rest_write_static_api_key_header":
		return record.Test.Symbol == "TestOperationDirectWriteBindsStaticHTTPMutationsBeforeApproval" && record.Contract.Pointer == "/supported_contracts/guarantees/7"
	}
	return false
}

func sourceFoundationFitTarget(source sourceLaneManifestRow, lane sourceLaneCell, ref sourceLaneTargetRef, observed sourceFoundationAssessmentObservations) (sourceFacts, sourceLaneTypedTarget, string, error) {
	facts := source.Facts
	var target sourceLaneTypedTarget
	if facts.bindings == nil {
		return facts, target, "", fmt.Errorf("exact fit bindings unavailable")
	}
	match := func(candidate sourceLaneTargetRef) bool { return sourceLaneTargetRefEqual(ref, candidate) }
	state := "materialized"
	if !slices.ContainsFunc(lane.References, match) {
		if !slices.ContainsFunc(lane.IntendedBindings, match) {
			return facts, target, "", fmt.Errorf("exact intended binding unavailable")
		}
		if !sourceFoundationAdmissionFit(observed.universe.admission, sourceFoundationCell{Key: source.Source.Key, Lane: lane.Lane}, ref) {
			return facts, target, "", fmt.Errorf("canonical intended fit lacks exact admission witness")
		}
		state = "canonical_intended"
		prefix := "internal/connectors/defs/" + ref.Connector + "/"
		missing := prefix + "cli_surface.json"
		if _, exists := facts.bindings.Artifacts[missing]; exists {
			return facts, target, "", fmt.Errorf("canonical intended fit requires missing CLI artifact")
		}
		descriptor, exists := facts.bindings.Canonical[ref.Connector]
		if !exists || descriptor.Connector != ref.Connector || descriptor.Staged.Identity.Digest != ref.Generation {
			return facts, target, "", fmt.Errorf("intended canonical generation unavailable")
		}
		if _, exists := descriptor.Staged.Outputs["cli_surface.json"]; !exists {
			return facts, target, "", fmt.Errorf("intended canonical CLI absent")
		}
		for _, observation := range facts.bindings.Observations {
			if observation.Connector == ref.Connector {
				return facts, target, "", fmt.Errorf("invalid current canonical input observation")
			}
		}
		for name, raw := range facts.bindings.Artifacts {
			if strings.HasPrefix(name, prefix) && !bytes.Equal(raw, descriptor.Staged.Outputs[strings.TrimPrefix(name, prefix)]) {
				return facts, target, "", fmt.Errorf("present execution artifact differs from canonical generation")
			}
		}
		staged := map[string][]byte{}
		for name, raw := range descriptor.Staged.Outputs {
			if name != "cli_surface.json" {
				current, exists := facts.bindings.Artifacts[prefix+name]
				if !exists || !bytes.Equal(raw, current) {
					return facts, target, "", fmt.Errorf("supporting canonical execution artifact missing or changed")
				}
			}
			staged[prefix+name] = append([]byte{}, raw...)
		}
		bundle, err := engine.Load(newVNextExecutionFS(ref.Connector, descriptor.Staged.Outputs), ref.Connector)
		if err != nil {
			return facts, target, "", fmt.Errorf("canonical intended bundle: %w", err)
		}
		// Fresh maps and a copied facts value cannot promote or repair the
		// actual loaded bundle, artifact collection, identity or source cells.
		authoring := map[string][]byte{}
		for name, raw := range facts.bindings.Authoring {
			authoring[name] = append([]byte{}, raw...)
		}
		facts.bindings = &sourceLaneBindingInputs{Authoring: authoring, Artifacts: staged, Canonical: map[string]vNextCanonicalDescriptor{ref.Connector: descriptor}, Bundles: map[string]engine.Bundle{ref.Connector: bundle}, Observations: []sourceLaneBindingObservation{}}
	}
	raw, exists := facts.bindings.Artifacts[ref.Artifact]
	if !exists || sourceBytesHash(raw) != ref.ArtifactSHA256 {
		return facts, target, "", fmt.Errorf("exact selected artifact pin mismatch")
	}
	node, err := sourceJSONPointer(raw, ref.Pointer)
	if err != nil {
		return facts, target, "", err
	}
	var annotations []sourceSemanticAnnotation
	if json.Unmarshal(observed.universe.manifest.annotationInputs, &annotations) != nil {
		return facts, target, "", fmt.Errorf("retained source annotation unavailable")
	}
	var annotation sourceSemanticAnnotation
	matched := false
	for _, a := range annotations {
		if a.Key == source.Source.Key {
			if matched {
				return facts, target, "", fmt.Errorf("source annotation ambiguous")
			}
			annotation = a
			matched = true
		}
	}
	if !matched {
		return facts, target, "", fmt.Errorf("exact source annotation unavailable")
	}
	if issues := sourceLaneCheckPresent(source.Source.Key, facts, annotation, ref, raw, node); len(issues) != 0 {
		return facts, target, "", fmt.Errorf("exact source/canonical fit refused: %s", issues[0].Code)
	}
	var code string
	target, code = sourceLaneObserveTypedTarget(ref, node, facts.bindings)
	if code != "" {
		return facts, target, "", fmt.Errorf("exact typed target: %s", code)
	}
	return facts, target, state, nil
}

func sourceFoundationSelect(facts sourceFacts, path, pointer string) (sourceFoundationSelector, error) {
	raw, exists := facts.bindings.Artifacts[path]
	if !exists {
		return sourceFoundationSelector{}, fmt.Errorf("supporting selector artifact missing")
	}
	value, err := sourceJSONPointer(raw, pointer)
	if err != nil {
		return sourceFoundationSelector{}, err
	}
	canonical, err := canonicalSourceJSON(value)
	if err != nil {
		return sourceFoundationSelector{}, err
	}
	return sourceFoundationSelector{Path: path, Pointer: pointer, ArtifactSHA256: sourceBytesHash(raw), ValueSHA256: sourceBytesHash(canonical)}, nil
}

var sourceFoundationSecretReference = regexp.MustCompile(`^\{\{\s*secrets\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}$`)

func sourceFoundationMechanismFits(requirement sourceFoundationRequirement, assessment sourceFoundationCellAssessment, source sourceLaneManifestRow, proofs []sourceFoundationAssertionResult, observed sourceFoundationAssessmentObservations) ([]sourceFoundationMechanismFit, error) {
	result := []sourceFoundationMechanismFit{}
	var lane sourceLaneCell
	for _, candidate := range source.Lanes {
		if candidate.Lane == assessment.Lane {
			lane = candidate
		}
	}
	for _, proof := range proofs {
		if proof.mechanism == "" || proof.mechanism == "base_check_exact_status" {
			return nil, fmt.Errorf("reviewed assertion has no applicable operation mechanism")
		}
		for _, ref := range requirement.FitBindings {
			facts, target, state, err := sourceFoundationFitTarget(source, lane, ref, observed)
			if err != nil {
				return nil, err
			}
			if ref.Lane != "direct_write" || target.REST == nil || (ref.Kind != "operation" && ref.Kind != "command") || target.REST.Method == "GET" || target.REST.Method == "HEAD" {
				return nil, fmt.Errorf("reviewed mechanism requires an exact REST write operation")
			}
			fit := sourceFoundationMechanismFit{ProofID: proof.ID, AtlasID: proof.AtlasID, Contract: proof.Contract, Binding: ref, DeclarationState: state, Mechanism: proof.mechanism, SourceRefs: append([]sourceFactRef{}, requirement.SourceRefs...), Selectors: []sourceFoundationSelector{}}
			selectAt := func(path, pointer string) error {
				selector, err := sourceFoundationSelect(facts, path, pointer)
				if err == nil && !slices.Contains(fit.Selectors, selector) {
					fit.Selectors = append(fit.Selectors, selector)
				}
				return err
			}
			if err := selectAt(ref.Artifact, ref.Pointer); err != nil {
				return nil, err
			}
			operationRef := ref
			if ref.Kind == "command" {
				if target.Binding.Binding.Kind != "operation" {
					return nil, fmt.Errorf("command does not select a fixed operation")
				}
				operationRef.Kind, operationRef.ID = "operation", target.Binding.Binding.ID
				operationRef.Artifact = "internal/connectors/defs/" + ref.Connector + "/operations.json"
				pointer, code := sourceLaneTargetPointer(facts.bindings.Artifacts[operationRef.Artifact], operationRef)
				if code != "" {
					return nil, fmt.Errorf("command operation selector unavailable")
				}
				operationRef.Pointer = pointer
				if err := selectAt(operationRef.Artifact, pointer); err != nil {
					return nil, err
				}
			}
			switch proof.mechanism {
			case "rest_write_structured_json_body":
				if target.REST.ContentType != "application/json" || len(target.REST.Body) != 0 || len(target.REST.BodySchema) == 0 || ref.SourceSchema == nil || !slices.Contains(requirement.SourceRefs, *ref.SourceSchema) || !slices.Contains(requirement.SourceRefs, facts.Refs["request_body"]) || sourceLaneBodyContract(facts, ref, target) != "" {
					return nil, fmt.Errorf("exact structured body source/selector fit missing")
				}
				descriptor := facts.bindings.Canonical[ref.Connector]
				schema := ""
				for _, operation := range descriptor.Operations {
					if operation.ID == ref.CanonicalID {
						schema = operation.SchemaRefs.Request
					}
				}
				if schema == "" {
					return nil, fmt.Errorf("canonical request schema selector missing")
				}
				if err := selectAt("internal/connectors/defs/"+ref.Connector+"/"+schema, ""); err != nil {
					return nil, err
				}
			case "rest_write_static_api_key_header":
				bundle := facts.bindings.Bundles[ref.Connector]
				if len(bundle.HTTP.Auth) == 0 || bundle.HTTP.Auth[0].When != "" || bundle.HTTP.Auth[0].Mode != "api_key_header" || bundle.HTTP.Auth[0].Prefix != "" || sourceFoundationAuthFit(facts, bundle.HTTP.Auth) != "auth_scheme_and_placement_match_not_credentials_or_execution_proof" || !slices.Contains(requirement.SourceRefs, facts.Refs["security"]) || !slices.Contains(requirement.SourceRefs, facts.Refs["security_schemes"]) {
					return nil, fmt.Errorf("exact static header auth source/selector fit missing")
				}
				match := sourceFoundationSecretReference.FindStringSubmatch(strings.TrimSpace(bundle.HTTP.Auth[0].Value))
				if len(match) != 2 || bundle.Spec == nil || !slices.Contains(bundle.Spec.SecretKeys(), match[1]) {
					return nil, fmt.Errorf("static auth secret declaration selector missing")
				}
				if err := sourceFoundationAuthSchemeCitation(requirement.SourceRefs, facts, bundle.HTTP.Auth[0]); err != nil {
					return nil, err
				}
				prefix := "internal/connectors/defs/" + ref.Connector + "/"
				if err := selectAt(prefix+"streams.json", "/base/auth/0"); err != nil {
					return nil, err
				}
				if err := selectAt(prefix+"spec.json", "/properties/"+escapeSourcePointer(match[1])); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unknown reviewed fit mechanism")
			}
			if requirement.Assessment == "connector_local_configuration" && (state != "canonical_intended" || ref.Kind != "command" || !slices.Contains(requirement.AffectedArtifacts, "internal/connectors/defs/"+ref.Connector+"/cli_surface.json")) {
				return nil, fmt.Errorf("local CLI artifact deficit not established")
			}
			sort.Slice(fit.Selectors, func(i, j int) bool {
				if fit.Selectors[i].Path != fit.Selectors[j].Path {
					return fit.Selectors[i].Path < fit.Selectors[j].Path
				}
				return fit.Selectors[i].Pointer < fit.Selectors[j].Pointer
			})
			result = append(result, fit)
		}
	}
	return result, nil
}

func sourceFoundationAuthSchemeCitation(refs []sourceFactRef, facts sourceFacts, selected engine.AuthSpec) error {
	var alternatives []map[string]json.RawMessage
	var schemes map[string]json.RawMessage
	if json.Unmarshal(facts.Groups["security"], &alternatives) != nil || json.Unmarshal(facts.Groups["security_schemes"], &schemes) != nil {
		return fmt.Errorf("auth source declarations unavailable")
	}
	for _, alternative := range alternatives {
		if len(alternative) != 1 {
			continue
		}
		for name, rawScopes := range alternative {
			var scopes []string
			if json.Unmarshal(rawScopes, &scopes) != nil || scopes == nil || len(scopes) != 0 {
				continue
			}
			scheme, ok := sourceResolveObject(facts, schemes[name], map[string]bool{}, 0)
			if !ok {
				continue
			}
			var protocol, location, header string
			_ = json.Unmarshal(scheme["type"], &protocol)
			_ = json.Unmarshal(scheme["in"], &location)
			_ = json.Unmarshal(scheme["name"], &header)
			if protocol != "apiKey" || location != "header" || !strings.EqualFold(header, selected.Header) {
				continue
			}
			group := facts.Refs["security_schemes"]
			canonical, err := canonicalSourceJSON(schemes[name])
			if err != nil {
				continue
			}
			want := sourceFactRef{DocumentID: group.DocumentID, Pointer: group.Pointer + "/" + escapeSourcePointer(name), ValueSHA256: sourceBytesHash(canonical)}
			if slices.Contains(refs, want) {
				return nil
			}
		}
	}
	return fmt.Errorf("exact chosen source auth scheme citation missing")
}
