package main

import (
	"encoding/json"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

type sourceFoundationFacetFit struct {
	Binding sourceLaneTargetRef `json:"binding"`
	Result  string              `json:"result"`
}

func sourceFoundationFacetFits(facts sourceFacts, references []sourceLaneTargetRef, kind string) []sourceFoundationFacetFit {
	result := []sourceFoundationFacetFit{}
	for _, ref := range references {
		fit := sourceFoundationFacetFit{Binding: ref, Result: "exact_declaration_unavailable"}
		if facts.bindings == nil {
			result = append(result, fit)
			continue
		}
		raw := facts.bindings.Artifacts[ref.Artifact]
		if sourceBytesHash(raw) != ref.ArtifactSHA256 {
			fit.Result = "declaration_changed"
			result = append(result, fit)
			continue
		}
		node, err := sourceJSONPointer(raw, ref.Pointer)
		if err != nil {
			result = append(result, fit)
			continue
		}
		target, code := sourceLaneObserveTypedTarget(ref, node, facts.bindings)
		if code != "" {
			fit.Result = code
			result = append(result, fit)
			continue
		}
		switch kind {
		case "body", "mime":
			fit.Result = sourceLaneBodyContract(facts, ref, target)
			if fit.Result == "" {
				fit.Result = "request_body_and_media_match_not_execution_proof"
				if kind == "mime" {
					fit.Result = "request_media_match_response_policy_separately_unproven"
				}
			}
		case "auth":
			fit.Result = sourceFoundationAuthFit(facts, facts.bindings.Bundles[ref.Connector].HTTP.Auth)
		case "paging":
			// Query names, a collection response or a declared paginator alone
			// cannot establish provider cursor/window/termination semantics.
			fit.Result = "paging_source_window_cursor_and_stop_contract_unresolved"
		}
		result = append(result, fit)
	}
	return result
}

// Inspect declared type/placement only. Never resolve conditions, interpolate
// configuration/secrets, construct an authenticator or perform provider I/O.
func sourceFoundationAuthFit(facts sourceFacts, declarations []engine.AuthSpec) string {
	const unresolved = "auth_source_flow_scope_or_selector_join_unresolved"
	if len(declarations) == 0 || strings.TrimSpace(declarations[0].When) != "" {
		return unresolved
	}
	selected := declarations[0] // the first unconditional declaration wins
	var alternatives []map[string]json.RawMessage
	if json.Unmarshal(facts.Groups["security"], &alternatives) != nil || alternatives == nil {
		return unresolved
	}
	if len(alternatives) == 0 && selected.Mode == "none" {
		return "explicit_anonymous_declaration_match_not_execution_proof"
	}
	var schemes map[string]json.RawMessage
	if json.Unmarshal(facts.Groups["security_schemes"], &schemes) != nil {
		return unresolved
	}
	for _, alternative := range alternatives {
		if alternative == nil {
			continue
		}
		if len(alternative) == 0 && selected.Mode == "none" {
			return "explicit_anonymous_declaration_match_not_execution_proof"
		}
		if len(alternative) != 1 {
			continue // no accidental OR interpretation of a conjunctive scheme set
		}
		for name, scopesRaw := range alternative {
			var scopes []string
			if json.Unmarshal(scopesRaw, &scopes) != nil || scopes == nil || len(scopes) != 0 {
				continue // scope-bearing grants need their own exact source join
			}
			scheme, ok := sourceResolveObject(facts, schemes[name], map[string]bool{}, 0)
			if !ok {
				continue
			}
			var protocol, format, location, field string
			if json.Unmarshal(scheme["type"], &protocol) != nil {
				continue
			}
			_ = json.Unmarshal(scheme["scheme"], &format)
			_ = json.Unmarshal(scheme["in"], &location)
			_ = json.Unmarshal(scheme["name"], &field)
			matched := protocol == "http" && (strings.EqualFold(format, "basic") && selected.Mode == "basic" ||
				strings.EqualFold(format, "bearer") && selected.Mode == "bearer")
			if protocol == "apiKey" && field != "" && selected.Prefix == "" {
				matched = location == "header" && selected.Mode == "api_key_header" && strings.EqualFold(field, selected.Header) ||
					location == "query" && selected.Mode == "api_key_query" && field == selected.Param
			}
			if matched {
				return "auth_scheme_and_placement_match_not_credentials_or_execution_proof"
			}
		}
	}
	return unresolved
}
