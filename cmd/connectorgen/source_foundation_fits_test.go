package main

import (
	"encoding/json"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestSourceFoundationFitCanonicalBody(t *testing.T) {
	key, facts, annotation := sourceBindingFixture099F(t, "body", nil)
	cell := sourceBindingOutcome099F(t, key, facts, annotation, 1)
	fits := sourceFoundationFacetFits(facts, cell.References, "body")
	if len(fits) != 1 || fits[0].Result != "request_body_and_media_match_not_execution_proof" ||
		!sourceLaneTargetRefEqual(fits[0].Binding, cell.References[0]) {
		t.Fatalf("actual canonical source/body fit: %+v", fits)
	}
}

func TestSourceFoundationFitBindingCustody(t *testing.T) {
	for _, scenario := range []string{"changed artifact", "wrong binding pin"} {
		t.Run(scenario, func(t *testing.T) {
			key, facts, annotation := sourceBindingFixture099F(t, "body", nil)
			cell := sourceBindingOutcome099F(t, key, facts, annotation, 1)
			if fits := sourceFoundationFacetFits(facts, cell.References, "body"); len(fits) != 1 || fits[0].Result != "request_body_and_media_match_not_execution_proof" {
				t.Fatalf("actual current canonical body fit control: %+v", fits)
			}
			if scenario == "changed artifact" {
				path := cell.References[0].Artifact
				// Equivalent JSON still violates the exact admitted artifact pin.
				facts.bindings.Artifacts[path] = append(append([]byte{}, facts.bindings.Artifacts[path]...), '\n')
			} else {
				cell.References[0].ArtifactSHA256 = sourceBytesHash([]byte("unrelated declaration"))
			}
			fits := sourceFoundationFacetFits(facts, cell.References, "body")
			if len(fits) != 1 || fits[0].Result != "declaration_changed" {
				t.Errorf("actual fit consumer reused %s: %+v", scenario, fits)
			}
		})
	}
}

func sourceFoundationAuthFacts(t *testing.T, security string) sourceFacts {
	t.Helper()
	node := json.RawMessage(`{"id":"auth.read","protocol":"rest","method":"get","path":"/items","source_operation":{"security":` + security + `,"responses":{"200":{"description":"ok"}}}}`)
	document := retainedSourceDocument{ID: "auth:fixture", Payload: json.RawMessage(`{"source_contract":{"components":{"securitySchemes":{"basic":{"type":"http","scheme":"basic"},"bearer":{"type":"http","scheme":"bearer"},"header":{"type":"apiKey","in":"header","name":"X-API-Key"},"query":{"type":"apiKey","in":"query","name":"api_key"},"oauth":{"type":"oauth2"}}}},"rest":{"operations":[` + string(node) + `]}}`)}
	return normalizeSourceFacts(retainedSourceOperation{Key: sourceOperationKey{Connector: "fixture", Inventory: "primary", ID: "auth.read"}, Observed: true, Node: node, Pointer: "/rest/operations/0"}, document, nil)
}

func TestSourceFoundationFitAuthTypePlacement(t *testing.T) {
	const matched = "auth_scheme_and_placement_match_not_credentials_or_execution_proof"
	const anonymous = "explicit_anonymous_declaration_match_not_execution_proof"
	const unresolved = "auth_source_flow_scope_or_selector_join_unresolved"
	for _, test := range []struct {
		name, security string
		auth           []engine.AuthSpec
		want           string
	}{
		{"basic", `[{"basic":[]}]`, []engine.AuthSpec{{Mode: "basic"}}, matched},
		{"bearer", `[{"bearer":[]}]`, []engine.AuthSpec{{Mode: "bearer"}}, matched},
		{"header", `[{"header":[]}]`, []engine.AuthSpec{{Mode: "api_key_header", Header: "x-api-key"}}, matched},
		{"query", `[{"query":[]}]`, []engine.AuthSpec{{Mode: "api_key_query", Param: "api_key"}}, matched},
		{"explicit no auth", `[]`, []engine.AuthSpec{{Mode: "none"}}, anonymous},
		{"anonymous alternative", `[{"basic":[]},{}]`, []engine.AuthSpec{{Mode: "none"}}, anonymous},
		{"header mismatch", `[{"header":[]}]`, []engine.AuthSpec{{Mode: "api_key_header", Header: "Other"}}, unresolved},
		{"query case is significant", `[{"query":[]}]`, []engine.AuthSpec{{Mode: "api_key_query", Param: "API_KEY"}}, unresolved},
		{"prefix changes placement", `[{"header":[]}]`, []engine.AuthSpec{{Mode: "api_key_header", Header: "X-API-Key", Prefix: "Bearer "}}, unresolved},
		{"scope not discarded", `[{"basic":["required-scope"]}]`, []engine.AuthSpec{{Mode: "basic"}}, unresolved},
		{"conjunction not alternative", `[{"basic":[],"bearer":[]}]`, []engine.AuthSpec{{Mode: "basic"}}, unresolved},
		{"oauth not bearer", `[{"oauth":[]}]`, []engine.AuthSpec{{Mode: "bearer"}}, unresolved},
		{"no condition evaluation", `[{"basic":[]}]`, []engine.AuthSpec{{Mode: "basic", When: "config.choice == basic"}, {Mode: "basic"}}, unresolved},
		{"first unconditional wins", `[{"basic":[]}]`, []engine.AuthSpec{{Mode: "none"}, {Mode: "basic"}}, unresolved},
		{"null source not anonymous", `null`, []engine.AuthSpec{{Mode: "none"}}, unresolved},
	} {
		t.Run(test.name, func(t *testing.T) {
			facts := sourceFoundationAuthFacts(t, test.security)
			if _, exists := facts.Refs["security"]; !exists {
				t.Fatal("did not reach actual retained auth normalization")
			}
			if got := sourceFoundationAuthFit(facts, test.auth); got != test.want {
				t.Fatalf("auth source/type/placement=%s want%s", got, test.want)
			}
		})
	}
}
