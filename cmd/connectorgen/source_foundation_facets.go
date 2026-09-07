package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
)

// Facets retain source/configuration observations for the assessed scope.
// They do not infer implementation absence from a missing join or proof.
type sourceFoundationFacet struct {
	Identity        sourceFoundationCell         `json:"identity"`
	Kind            string                       `json:"kind"`
	SourceRefs      []sourceFactRef              `json:"source_refs"`
	ObservedValues  []string                     `json:"observed_values"`
	DeclarationRefs []sourceLaneTargetRef        `json:"declaration_refs"`
	FitEvidence     []sourceFoundationFacetFit   `json:"fit_evidence"`
	CandidateOwners []sourceFoundationFacetOwner `json:"candidate_owners"`
	Status          string                       `json:"status"`
	Limitation      string                       `json:"limitation"`
}

type sourceFoundationFacetOwner struct {
	AtlasID     string                      `json:"atlas_id"`
	Owner       string                      `json:"owner"`
	Contract    sourceFoundationContractRef `json:"contract"`
	Symbols     []sourceFoundationSymbol    `json:"symbols"`
	Constraints []string                    `json:"constraints"`
}

// These are the finite navigation candidates selected by the140 design,
// never a provider dispatch table or a capability-selection rule.
func sourceFoundationFacetCandidates(kind string) []string {
	switch kind {
	case "mime", "body":
		return []string{"runtime.direct-execution.v1", "runtime.provider-extension-seams.v1"}
	case "auth":
		return []string{"runtime.declared-password-token.v1", "runtime.declared-session.v1", "runtime.declared-aws-sigv4.v1", "runtime.direct-execution.v1", "runtime.provider-extension-seams.v1"}
	case "paging":
		return []string{"runtime.direct-execution.v1", "runtime.offset-count-pagination.v1"}
	default:
		return nil
	}
}

func buildSourceFoundationFacets(ctx context.Context, observed sourceFoundationAssessmentObservations) ([]sourceFoundationFacet, error) {
	rows := map[sourceOperationKey]sourceLaneManifestRow{}
	for _, row := range observed.universe.manifest.SourceOperations {
		rows[row.Source.Key] = row
	}
	result := []sourceFoundationFacet{}
	for _, assessment := range observed.authored {
		source, exists := rows[assessment.Key]
		if !exists {
			return nil, fmt.Errorf("foundation facet source unavailable")
		}
		for _, kind := range []string{"mime", "auth", "body", "paging"} {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			row := sourceFoundationFacet{Identity: sourceFoundationCell{Key: assessment.Key, Lane: assessment.Lane}, Kind: kind,
				SourceRefs: []sourceFactRef{}, ObservedValues: []string{}, DeclarationRefs: []sourceLaneTargetRef{}, CandidateOwners: []sourceFoundationFacetOwner{},
				FitEvidence: []sourceFoundationFacetFit{},
				Status:      "unresolved", Limitation: "Source fields and Atlas candidates are observations; exact mechanism/configuration fit and matching behavioural assertion remain separately required."}
			groups := map[string][]string{"mime": {"request_body", "responses"}, "auth": {"security", "security_schemes"}, "body": {"request_body", "parameters", "path_parameters"}, "paging": {"parameters", "path_parameters", "responses"}}
			for _, group := range groups[kind] {
				if ref, exists := source.Facts.Refs[group]; exists {
					row.SourceRefs = append(row.SourceRefs, ref)
				}
			}
			// A missing normalized field is not evidence of provider absence.
			if len(row.SourceRefs) == 0 {
				if ref, exists := source.Facts.Refs["source_operation"]; exists {
					row.SourceRefs = append(row.SourceRefs, ref)
				}
				row.Limitation = "Required facet fields are not established by the retained normalized groups; preserve source uncertainty and inspect the cited operation before classifying a gap."
			}
			row.ObservedValues = sourceFoundationFacetValues(source.Facts, kind)
			for _, lane := range source.Lanes {
				if lane.Lane == assessment.Lane {
					row.DeclarationRefs = append(row.DeclarationRefs, lane.References...)
					row.FitEvidence = sourceFoundationFacetFits(source.Facts, lane.References, kind)
				}
			}
			for _, id := range sourceFoundationFacetCandidates(kind) {
				entry, exists := observed.atlas.entries[id]
				if !exists {
					return nil, fmt.Errorf("foundation facet Atlas navigation candidate unavailable: %s", id)
				}
				pointer := "/supported_contracts/request_response_shapes"
				value, err := sourceJSONPointer(entry.raw, pointer)
				if err != nil {
					return nil, fmt.Errorf("foundation facet contract unavailable: %w", err)
				}
				canonical, err := canonicalSourceJSON(value)
				if err != nil {
					return nil, err
				}
				row.CandidateOwners = append(row.CandidateOwners, sourceFoundationFacetOwner{AtlasID: id, Owner: entry.Owner.PrimaryPackage,
					Contract: sourceFoundationContractRef{Pointer: pointer, ValueSHA256: sourceBytesHash(canonical)}, Symbols: entry.Owner.Symbols, Constraints: entry.Constraints})
			}
			result = append(result, row)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Identity != result[j].Identity {
			return sourceFoundationCellLess(result[i].Identity, result[j].Identity)
		}
		return result[i].Kind < result[j].Kind
	})
	return result, nil
}

func sourceFoundationFacetValues(facts sourceFacts, kind string) []string {
	values := []string{}
	switch kind {
	case "mime", "body":
		body, ok := sourceResolveObject(facts, facts.Groups["request_body"], map[string]bool{}, 0)
		if ok {
			var content map[string]json.RawMessage
			if json.Unmarshal(body["content"], &content) == nil {
				for media := range content {
					values = append(values, "request_media:"+media)
				}
			}
		}
		if kind == "body" {
			if ok {
				var required bool
				if raw, present := body["required"]; !present {
					values = append(values, "request_required:unspecified")
				} else if json.Unmarshal(raw, &required) != nil || string(raw) == "null" {
					values = append(values, "request_required:unresolved")
				} else {
					values = append(values, fmt.Sprintf("request_required:%t", required))
				}
			}
			for _, parameter := range facts.Parameters {
				values = append(values, "parameter:"+parameter.In+":"+parameter.Name)
			}
		} else {
			var responses map[string]json.RawMessage
			if json.Unmarshal(facts.Groups["responses"], &responses) != nil || len(responses) == 0 {
				values = append(values, "response_contract_unresolved")
			}
			for status, raw := range responses {
				success, known := sourceResponseStatus(status)
				if !known {
					values = append(values, "response_status_unresolved:"+status)
					continue
				}
				if !success {
					continue
				}
				response, resolved := sourceResolveObject(facts, raw, map[string]bool{}, 0)
				var content map[string]json.RawMessage
				if !resolved || json.Unmarshal(response["content"], &content) != nil || len(content) == 0 {
					if status != "204" && status != "205" {
						values = append(values, "response_media_unresolved:"+status)
					}
					continue
				}
				for media := range content {
					values = append(values, "response_media:"+status+":"+media)
				}
			}
		}
	case "auth":
		var alternatives []map[string]json.RawMessage
		if json.Unmarshal(facts.Groups["security"], &alternatives) == nil && alternatives != nil {
			if len(alternatives) == 0 {
				values = append(values, "security_explicit_empty")
			}
			for index, alternative := range alternatives {
				if alternative == nil {
					values = append(values, fmt.Sprintf("security_alternative_%d_unresolved", index))
				} else if len(alternative) == 0 {
					values = append(values, fmt.Sprintf("security_alternative_%d_anonymous", index))
				}
				for name, raw := range alternative {
					prefix := fmt.Sprintf("security_alternative_%d_scheme:%s", index, name)
					values = append(values, prefix)
					var scopes []string
					if json.Unmarshal(raw, &scopes) != nil || scopes == nil {
						values = append(values, prefix+"_scopes_unresolved")
					} else {
						for _, scope := range scopes {
							values = append(values, prefix+"_scope:"+scope)
						}
					}
				}
			}
		} else {
			values = append(values, "security_unspecified_or_unresolved")
		}
	case "paging":
		for _, parameter := range facts.Parameters {
			values = append(values, "parameter:"+parameter.In+":"+parameter.Name)
		}
	}
	sort.Strings(values)
	return slices.Compact(values)
}
