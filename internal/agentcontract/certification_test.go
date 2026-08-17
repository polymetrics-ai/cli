package agentcontract

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const (
	certificationGeneratedCommand = "go run ./cmd/connectorgen certification-matrix --check"
	certificationCredentialNote   = "Only the credential use documented by this record's protocol exchanges was verified; no broader credential scope is claimed."
)

func TestEvaluateCertificationGateGitHubBaselineAndGreenFixture(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	request := certificationGateRequestFor(contract, "integrate_sub_pr")

	baseline, err := EvaluateCertificationGate(repositoryRoot(t), contract, request)
	if err != nil {
		t.Fatalf("evaluate current generated GitHub baseline: %v", err)
	}
	if baseline.Decision != CertificationGateRetry {
		t.Fatalf("current GitHub decision = %q, want %q; failures=%#v", baseline.Decision, CertificationGateRetry, baseline.Failures)
	}
	assertCertificationFailureID(t, baseline, "capability/github/capability:check/live_evidence")

	contract = loadCertificationTestContract(t)
	request = certificationGateRequest("integrate_sub_pr")
	fixtureRoot := writeGreenCertificationFixture(t)
	green, err := EvaluateCertificationGate(fixtureRoot, contract, request)
	if err != nil {
		t.Fatalf("evaluate all-green generated fixture: %v", err)
	}
	if green.Decision != CertificationGateProceed {
		t.Fatalf("all-green fixture decision = %q, want %q; failures=%#v", green.Decision, CertificationGateProceed, green.Failures)
	}
	if len(green.Failures) != 0 {
		t.Fatalf("all-green fixture returned failures: %#v", green.Failures)
	}
}

func TestCertificationGateBindingCriteriaYieldExactCellIDs(t *testing.T) {
	contract := loadCertificationTestContract(t)
	for _, criterion := range []string{"declared", "implemented", "fixture_tested", "live_tested"} {
		t.Run(criterion, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			cell := certificationFixtureCapabilityCell(t, root)
			cell[criterion] = false
			if criterion == "live_tested" {
				cell["live_evidence"] = []any{}
			}
			writeCertificationFixtureCapabilityCell(t, root, cell)
			rewriteCertificationFixtureDerivedReports(t, root)

			verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateRetry {
				t.Fatalf("decision = %q, want RETRY; failures=%#v", verdict.Decision, verdict.Failures)
			}
			want := "capability/github/capability:check/" + criterion
			assertCertificationFailureID(t, verdict, want)
		})
	}
}

func TestCertificationGateRejectsLiveTestedWithoutEvidenceWithCoordinates(t *testing.T) {
	contract := loadCertificationTestContract(t)
	root := writeGreenCertificationFixture(t)
	cell := certificationFixtureCapabilityCell(t, root)
	cell["live_tested"] = true
	cell["live_evidence"] = []any{}
	writeCertificationFixtureCapabilityCell(t, root, cell)

	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateHalt || len(verdict.Failures) != 1 {
		t.Fatalf("decision = %#v, want one coordinate-bearing HALT", verdict)
	}
	failure := verdict.Failures[0]
	if failure.ID != "capability/github/capability:check/live_evidence" || failure.CellID != "capability/github/capability:check" || failure.EvidenceID != "" {
		t.Fatalf("failure coordinates = %#v, want capability live-evidence coordinate", failure)
	}
}

func TestCertificationGateRejectsMalformedEvidencePointersWithTrustedCoordinates(t *testing.T) {
	contract := loadCertificationTestContract(t)
	tests := []struct {
		name           string
		mutate         func(map[string]any)
		wantEvidenceID string
	}{
		{
			name: "unsafe record",
			mutate: func(pointer map[string]any) {
				pointer["record"] = "../../outside.json"
			},
		},
		{
			name: "safe record with invalid identity",
			mutate: func(pointer map[string]any) {
				pointer["provider"] = "unsafe provider"
			},
			wantEvidenceID: "internal/connectors/certifications/evidence/capability.json",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
			pointer := matrix["connectors"].([]any)[0].(map[string]any)["cells"].([]any)[0].(map[string]any)["live_evidence"].([]any)[0].(map[string]any)
			test.mutate(pointer)
			writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", matrix)

			verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateHalt || len(verdict.Failures) != 1 {
				t.Fatalf("decision = %#v, want one coordinate-bearing HALT", verdict)
			}
			failure := verdict.Failures[0]
			if failure.ID != "evidence/invalid_pointer" || failure.CellID != "capability/github/capability:check" || failure.EvidenceID != test.wantEvidenceID || failure.Message != "invalid_pointer" {
				t.Fatalf("failure = %#v, want fixed invalid-pointer coordinates", failure)
			}
		})
	}
}

func TestCertificationGateRejectsDerivedStatusAndAggregateMismatches(t *testing.T) {
	contract := loadCertificationTestContract(t)
	tests := []struct {
		name   string
		mutate func(*testing.T, string)
	}{
		{
			name: "capability complete",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
				matrix["connectors"].([]any)[0].(map[string]any)["capability_complete"] = false
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", matrix)
			},
		},
		{
			name: "capability baseline",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
				matrix["baseline"].(map[string]any)["capability_complete"] = 0
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", matrix)
			},
		},
		{
			name: "flow baseline",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				matrix["baseline"].(map[string]any)["certified"] = 0
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
		},
		{
			name: "both reported statuses",
			mutate: func(t *testing.T, root string) {
				flow := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				status := flow["connector_statuses"].([]any)[0].(map[string]any)
				status["certified"] = false
				status["label"] = "COMMUNITY BUILD, UNCERTIFIED"
				status["warning"] = "This connector is reachable but is a COMMUNITY BUILD, UNCERTIFIED."
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", flow)

				artifact := readCertificationFixtureObject(t, root, "internal/connectors/certifications/status.json")
				status = artifact["connectors"].([]any)[0].(map[string]any)
				status["certified"] = false
				status["label"] = "COMMUNITY BUILD, UNCERTIFIED"
				status["warning"] = "This connector is reachable but is a COMMUNITY BUILD, UNCERTIFIED."
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/status.json", artifact)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			test.mutate(t, root)

			verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateHalt {
				t.Fatalf("decision = %#v, want HALT for a derived-report mismatch", verdict)
			}
		})
	}
}

func TestCertificationGatePreservesCoordinatesAcrossWorkflowSyncAndFlowCells(t *testing.T) {
	contract := loadCertificationTestContract(t)
	tests := []struct {
		name   string
		mutate func(*testing.T, string)
		want   string
	}{
		{
			name: "workflow",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				cell := matrix["workflows"].([]any)[0].(map[string]any)["cells"].([]any)[0].(map[string]any)
				cell["live_evidence"] = []any{}
				cell["live_tested"] = false
				matrix["workflows"].([]any)[0].(map[string]any)["complete"] = false
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
			want: "workflow/github/etl/live_evidence",
		},
		{
			name: "sync mode",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				cell := matrix["sync_mode_cells"].([]any)[0].(map[string]any)["cells"].([]any)[0].(map[string]any)
				cell["live_evidence"] = []any{}
				cell["live_tested"] = false
				matrix["sync_mode_cells"].([]any)[0].(map[string]any)["complete"] = false
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
			want: "sync_mode/github/full_overwrite/api_read_into_warehouse/live_evidence",
		},
		{
			name: "flow pair",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				cell := matrix["pair_sets"].([]any)[0].(map[string]any)["cell"].(map[string]any)
				cell["live_evidence"] = []any{}
				cell["live_tested"] = false
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
			want: "flow/api_to_api/github/github/live_evidence",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			test.mutate(t, root)
			rewriteCertificationFixtureDerivedReports(t, root)
			verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateRetry {
				t.Fatalf("decision = %#v, want RETRY", verdict)
			}
			assertCertificationFailureID(t, verdict, test.want)
		})
	}
}

func TestCertificationGateDoesNotPromoteReachabilityOrImplementedWithoutLiveEvidence(t *testing.T) {
	contract := loadCertificationTestContract(t)
	root := writeGreenCertificationFixture(t)
	cell := certificationFixtureCapabilityCell(t, root)
	cell["declared"] = true
	cell["implemented"] = true
	cell["fixture_tested"] = true
	cell["live_tested"] = false
	cell["live_evidence"] = []any{}
	writeCertificationFixtureCapabilityCell(t, root, cell)
	rewriteCertificationFixtureDerivedReports(t, root)

	for _, path := range []string{
		"internal/connectors/certifications/capability-matrix.json",
		"internal/connectors/certifications/flow-matrix.json",
		"internal/connectors/certifications/status.json",
	} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			t.Fatalf("fixture input %s is not present: %v", path, err)
		}
	}

	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateRetry {
		t.Fatalf("implemented, reachable fixture decision = %q, want RETRY; failures=%#v", verdict.Decision, verdict.Failures)
	}
	assertCertificationFailureID(t, verdict, "capability/github/capability:check/live_tested")
	assertCertificationFailureID(t, verdict, "capability/github/capability:check/live_evidence")
}

func TestCertificationGateHaltCarriesExactCellAndEvidenceCoordinates(t *testing.T) {
	contract := loadCertificationTestContract(t)
	root := writeGreenCertificationFixture(t)
	record := filepath.Join(root, "internal", "connectors", "certifications", "evidence", "capability.json")
	if err := os.Remove(record); err != nil {
		t.Fatal(err)
	}
	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateHalt || len(verdict.Failures) != 1 {
		t.Fatalf("missing evidence verdict = %#v, want one HALT failure", verdict)
	}
	failure := verdict.Failures[0]
	if failure.ID != "evidence/internal/connectors/certifications/evidence/capability.json/missing" ||
		failure.CellID != "capability/github/capability:check" ||
		failure.EvidenceID != "internal/connectors/certifications/evidence/capability.json" {
		t.Fatalf("failure coordinates = %#v, want exact cell and evidence identifiers", failure)
	}
}

func TestCertificationGateMatchesSemanticallyEquivalentEvidenceProof(t *testing.T) {
	contract := loadCertificationTestContract(t)
	root := writeGreenCertificationFixture(t)
	path := filepath.Join(root, "internal", "connectors", "certifications", "evidence", "capability.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	formatted := bytes.Replace(raw, []byte(`"value":{}`), []byte(`"value": { }`), 1)
	if bytes.Equal(formatted, raw) {
		t.Fatal("fixture did not produce a semantically equivalent proof formatting change")
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		t.Fatal(err)
	}
	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateProceed {
		t.Fatalf("equivalent proof formatting decision = %#v, want PROCEED", verdict)
	}
}

func TestCertificationGateMatchesEvidenceProofsWithReorderedJSONMembers(t *testing.T) {
	contract := loadCertificationTestContract(t)
	root := writeGreenCertificationFixture(t)
	fingerprint := "{{pmcertfp:v1:" + strings.Repeat("a", 64) + "}}"
	value := map[string]any{"a": fingerprint, "b": fingerprint}

	matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
	pointer := matrix["connectors"].([]any)[0].(map[string]any)["cells"].([]any)[0].(map[string]any)["live_evidence"].([]any)[0].(map[string]any)
	pointerProof := pointer["proof"].(map[string]any)
	pointerProof["http_exchanges"].([]any)[0].(map[string]any)["request"].(map[string]any)["body"].(map[string]any)["value"] = value
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", matrix)

	evidence := readCertificationFixtureObject(t, root, "internal/connectors/certifications/evidence/capability.json")
	recordProof := evidence["proof"].(map[string]any)
	recordProof["http_exchanges"].([]any)[0].(map[string]any)["request"].(map[string]any)["body"].(map[string]any)["value"] = value
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/evidence/capability.json", evidence)

	path := filepath.Join(root, "internal", "connectors", "certifications", "evidence", "capability.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ordered, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	reordered, err := json.Marshal(struct {
		B string `json:"b"`
		A string `json:"a"`
	}{B: fingerprint, A: fingerprint})
	if err != nil {
		t.Fatal(err)
	}
	changed := bytes.Replace(raw, ordered, reordered, 1)
	if bytes.Equal(changed, raw) {
		t.Fatal("fixture did not reorder an evidence JSON object")
	}
	if err := os.WriteFile(path, changed, 0o644); err != nil {
		t.Fatal(err)
	}

	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateProceed {
		t.Fatalf("reordered proof decision = %#v, want PROCEED", verdict)
	}
}

func TestCertificationGateRejectsEmptyFingerprintSequences(t *testing.T) {
	contract := loadCertificationTestContract(t)
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "command fingerprint",
			mutate: func(proof map[string]any) {
				proof["pm_command_fingerprint"] = ""
			},
		},
		{
			name: "request target",
			mutate: func(proof map[string]any) {
				proof["http_exchanges"].([]any)[0].(map[string]any)["request"].(map[string]any)["target"] = ""
			},
		},
		{
			name: "credential fingerprint",
			mutate: func(proof map[string]any) {
				proof["credential_fingerprints"] = []any{""}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			mutateCertificationFixtureCapabilityProof(t, root, test.mutate)

			verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateHalt {
				t.Fatalf("empty fingerprint decision = %#v, want HALT", verdict)
			}
		})
	}
}

func TestCertificationGateRejectsIncompleteFlowTopology(t *testing.T) {
	contract := loadCertificationTestContract(t)
	tests := []struct {
		name   string
		mutate func(*testing.T, string)
	}{
		{
			name: "sync mode primitive cross product",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				matrix["sync_mode_kinds"] = append(matrix["sync_mode_kinds"].([]any), map[string]any{"id": "incremental", "discovery_source": "test"})
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
		},
		{
			name: "required warehouse primitive inventory",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				matrix["sync_primitives"] = matrix["sync_primitives"].([]any)[:3]
				set := matrix["sync_mode_cells"].([]any)[0].(map[string]any)
				set["cells"] = set["cells"].([]any)[:3]
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
		},
		{
			name: "derived complete flag",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				matrix["sync_mode_cells"].([]any)[0].(map[string]any)["complete"] = false
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
		},
		{
			name: "flow pair coverage",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				matrix["pair_sets"] = matrix["pair_sets"].([]any)[:3]
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
		},
		{
			name: "connector identity",
			mutate: func(t *testing.T, root string) {
				capabilities := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
				capabilities["connectors"] = append(capabilities["connectors"].([]any), map[string]any{
					"name": "other", "integration_type": "api", "capability_complete": false,
					"cells": []any{map[string]any{
						"function_kind": "capability:check", "applicable": false, "declared": false, "implemented": false,
						"fixture_tested": false, "live_tested": false, "fixture_evidence": []any{}, "live_evidence": []any{},
						"not_applicable": map[string]any{"code": "unsupported_capability", "reason": "fixture connector is outside the supported surface"},
					}},
				})
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", capabilities)

				status := readCertificationFixtureObject(t, root, "internal/connectors/certifications/status.json")
				status["connectors"] = append(status["connectors"].([]any), map[string]any{
					"connector": "other", "certified": false, "label": "COMMUNITY BUILD, UNCERTIFIED", "warning": "This connector is reachable but is a COMMUNITY BUILD, UNCERTIFIED.",
				})
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/status.json", status)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			test.mutate(t, root)

			verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateHalt {
				t.Fatalf("incomplete topology decision = %#v, want HALT", verdict)
			}
		})
	}
}

func TestCertificationGateRejectsFlowKindInventoryDrift(t *testing.T) {
	contract := loadCertificationTestContract(t)
	tests := []struct {
		name   string
		mutate func(*testing.T, string)
	}{
		{
			name: "omitted producer kind",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				matrix["flow_kinds"] = append([]any{}, matrix["flow_kinds"].([]any)[:3]...)
				matrix["pair_sets"] = append([]any{}, matrix["pair_sets"].([]any)[:3]...)
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
		},
		{
			name: "added producer kind",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				matrix["flow_kinds"] = append(matrix["flow_kinds"].([]any), map[string]any{"id": "api_to_api_extra", "source_role": "api_source", "destination_role": "api_destination"})
				pairSet := matrix["pair_sets"].([]any)[0].(map[string]any)
				extra := make(map[string]any, len(pairSet))
				for key, value := range pairSet {
					extra[key] = value
				}
				extra["flow_kind"] = "api_to_api_extra"
				extra["cell"] = certificationTestNotApplicableCell("unsupported_extra_flow", "fixture does not define the added flow kind")
				matrix["pair_sets"] = append(matrix["pair_sets"].([]any), extra)
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
		},
		{
			name: "remapped producer kind",
			mutate: func(t *testing.T, root string) {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
				matrix["flow_kinds"].([]any)[0].(map[string]any)["source_role"] = "database_source"
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			test.mutate(t, root)

			verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateHalt {
				t.Fatalf("decision = %#v, want HALT for flow-kind inventory drift", verdict)
			}
		})
	}
}

func TestCertificationGateRejectsImmutableFlowOverridePromotion(t *testing.T) {
	contract := loadCertificationTestContract(t)
	root := writeGreenCertificationFixture(t)
	matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
	baseCell := matrix["pair_sets"].([]any)[0].(map[string]any)["cell"].(map[string]any)
	overrideCell := cloneCertificationFixtureObject(t, baseCell)
	baseCell["declared"] = false
	baseCell["implemented"] = false
	baseCell["fixture_tested"] = false
	baseCell["fixture_evidence"] = []any{}
	baseCell["live_tested"] = false
	baseCell["live_evidence"] = []any{}
	matrix["pair_overrides"] = []any{map[string]any{
		"flow_kind": "api_to_api", "source": "github", "destination": "github", "mediator": "local_parquet_warehouse", "cell": overrideCell,
	}}
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)
	rewriteCertificationFixtureDerivedReports(t, root)

	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateHalt {
		t.Fatalf("immutable flow override promotion decision = %#v, want HALT", verdict)
	}
}

func TestCertificationGateRejectsFlowPairsThatDisagreeWithEndpointRoles(t *testing.T) {
	contract := loadCertificationTestContract(t)
	tests := []struct {
		name    string
		mutate  func(map[string]any)
		failure string
		cellID  string
	}{
		{
			name: "applicability",
			mutate: func(matrix map[string]any) {
				roles := matrix["connector_roles"].([]any)[0].(map[string]any)["roles"].([]any)
				apiSource := roles[0].(map[string]any)
				apiSource["applicable"] = false
				apiSource["declared"] = false
				apiSource["implemented"] = false
				apiSource["not_applicable"] = map[string]any{
					"code":   "unsupported_api_source",
					"reason": "fixture connector has no api source",
				}
			},
			failure: "flow/api_to_api/github/github/role_invariant",
			cellID:  "flow/api_to_api/github/github",
		},
		{
			name: "declared and implemented conjunction",
			mutate: func(matrix map[string]any) {
				roles := matrix["connector_roles"].([]any)[0].(map[string]any)["roles"].([]any)
				roles[0].(map[string]any)["declared"] = false
				roles[0].(map[string]any)["implemented"] = false
			},
			failure: "flow/api_to_api/github/github/role_invariant",
			cellID:  "flow/api_to_api/github/github",
		},
		{
			name: "not applicable code and reason",
			mutate: func(matrix map[string]any) {
				pair := matrix["pair_sets"].([]any)[1].(map[string]any)
				notApplicable := pair["cell"].(map[string]any)["not_applicable"].(map[string]any)
				notApplicable["code"] = "destination_other"
				notApplicable["reason"] = "destination mismatched reason"
			},
			failure: "flow/api_to_database/github/github/role_invariant",
			cellID:  "flow/api_to_database/github/github",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
			test.mutate(matrix)
			writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)

			verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateHalt || len(verdict.Failures) != 1 {
				t.Fatalf("role/pair invariant verdict = %#v, want one HALT", verdict)
			}
			failure := verdict.Failures[0]
			if failure.ID != test.failure || failure.CellID != test.cellID || failure.EvidenceID != "" {
				t.Fatalf("role/pair invariant coordinates = %#v", failure)
			}
		})
	}
}

func TestCertificationGateBindsRawFlowPairSetEvidenceBeforeOverrides(t *testing.T) {
	contract := loadCertificationTestContract(t)
	const flowCellID = "flow/api_to_api/github/github"
	tests := []struct {
		name       string
		mutate     func(*testing.T, string, map[string]any)
		failureID  string
		evidenceID string
	}{
		{
			name: "safe missing base record",
			mutate: func(_ *testing.T, _ string, pointer map[string]any) {
				pointer["record"] = "internal/connectors/certifications/evidence/missing-flow.json"
			},
			failureID:  "evidence/internal/connectors/certifications/evidence/missing-flow.json/missing",
			evidenceID: "internal/connectors/certifications/evidence/missing-flow.json",
		},
		{
			name: "mismatched base record",
			mutate: func(_ *testing.T, _ string, pointer map[string]any) {
				pointer["run_id"] = "mismatched-run"
			},
			failureID:  "evidence/internal/connectors/certifications/evidence/flow.json/mismatch",
			evidenceID: "internal/connectors/certifications/evidence/flow.json",
		},
		{
			name: "wrong-coordinate base record",
			mutate: func(t *testing.T, root string, pointer map[string]any) {
				record := readCertificationFixtureObject(t, root, "internal/connectors/certifications/evidence/flow.json")
				record["destination"] = "other"
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/evidence/wrong-flow.json", record)
				pointer["record"] = "internal/connectors/certifications/evidence/wrong-flow.json"
			},
			failureID:  "evidence/internal/connectors/certifications/evidence/wrong-flow.json/binding",
			evidenceID: "internal/connectors/certifications/evidence/wrong-flow.json",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
			baseCell := matrix["pair_sets"].([]any)[0].(map[string]any)["cell"].(map[string]any)
			overrideCell := cloneCertificationFixtureObject(t, baseCell)
			matrix["pair_overrides"] = []any{map[string]any{
				"flow_kind": "api_to_api", "source": "github", "destination": "github", "mediator": "local_parquet_warehouse", "cell": overrideCell,
			}}
			pointer := baseCell["live_evidence"].([]any)[0].(map[string]any)
			test.mutate(t, root, pointer)
			writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)

			verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateHalt || len(verdict.Failures) != 1 {
				t.Fatalf("raw base evidence verdict = %#v, want one HALT", verdict)
			}
			failure := verdict.Failures[0]
			if failure.ID != test.failureID || failure.CellID != flowCellID || failure.EvidenceID != test.evidenceID {
				t.Fatalf("failure coordinates = %#v, want id=%q cell=%q evidence=%q", failure, test.failureID, flowCellID, test.evidenceID)
			}
		})
	}
}

func TestCertificationGateRejectsDistinctLargeProofNumbers(t *testing.T) {
	contract := loadCertificationTestContract(t)
	root := writeGreenCertificationFixture(t)
	matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
	pointer := matrix["connectors"].([]any)[0].(map[string]any)["cells"].([]any)[0].(map[string]any)["live_evidence"].([]any)[0].(map[string]any)
	pointerProof := pointer["proof"].(map[string]any)
	pointerBody := pointerProof["http_exchanges"].([]any)[0].(map[string]any)["request"].(map[string]any)["body"].(map[string]any)
	pointerBody["original_bytes"] = int64(9007199254740992)
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", matrix)

	evidence := readCertificationFixtureObject(t, root, "internal/connectors/certifications/evidence/capability.json")
	evidenceProof := evidence["proof"].(map[string]any)
	evidenceBody := evidenceProof["http_exchanges"].([]any)[0].(map[string]any)["request"].(map[string]any)["body"].(map[string]any)
	evidenceBody["original_bytes"] = int64(9007199254740993)
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/evidence/capability.json", evidence)

	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateHalt || len(verdict.Failures) != 1 {
		t.Fatalf("large proof number mismatch verdict = %#v, want one HALT", verdict)
	}
	failure := verdict.Failures[0]
	if failure.ID != "evidence/internal/connectors/certifications/evidence/capability.json/mismatch" || failure.CellID != "capability/github/capability:check" || failure.EvidenceID != "internal/connectors/certifications/evidence/capability.json" {
		t.Fatalf("large proof number mismatch coordinates = %#v", failure)
	}
}

func TestCertificationGateRejectsEscapedOrNonRegularInputs(t *testing.T) {
	contract := loadCertificationTestContract(t)
	tests := []struct {
		name      string
		prepare   func(*testing.T) string
		failureID string
	}{
		{
			name: "artifact ancestor symlink",
			prepare: func(t *testing.T) string {
				outside := writeGreenCertificationFixture(t)
				root := t.TempDir()
				parent := filepath.Join(root, "internal", "connectors")
				if err := os.MkdirAll(parent, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(outside, "internal", "connectors", "certifications"), filepath.Join(parent, "certifications")); err != nil {
					t.Skipf("cannot create certification artifact symlink: %v", err)
				}
				return root
			},
		},
		{
			name: "evidence directory symlink",
			prepare: func(t *testing.T) string {
				root := writeGreenCertificationFixture(t)
				outside := writeGreenCertificationFixture(t)
				directory := filepath.Join(root, "internal", "connectors", "certifications", "evidence")
				if err := os.RemoveAll(directory); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(outside, "internal", "connectors", "certifications", "evidence"), directory); err != nil {
					t.Skipf("cannot create evidence directory symlink: %v", err)
				}
				return root
			},
		},
		{
			name: "non regular evidence record",
			prepare: func(t *testing.T) string {
				root := writeGreenCertificationFixture(t)
				record := filepath.Join(root, "internal", "connectors", "certifications", "evidence", "capability.json")
				if err := os.Remove(record); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(record, 0o755); err != nil {
					t.Fatal(err)
				}
				return root
			},
			failureID: "input/evidence/decode",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := test.prepare(t)
			verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateHalt {
				t.Fatalf("unsafe input decision = %#v, want HALT", verdict)
			}
			if test.failureID != "" {
				assertCertificationFailureID(t, verdict, test.failureID)
			}
		})
	}
}

func TestCertificationGateAcceptsProducerValidDeliveryLimitations(t *testing.T) {
	contract := loadCertificationTestContract(t)
	root := writeGreenCertificationFixture(t)
	mutateCertificationFixtureFlowProof(t, root, func(proof map[string]any) {
		delivery := proof["flow"].(map[string]any)["delivery"].(map[string]any)
		delivery["provider_idempotency_key"] = false
		delivery["limitations"] = []any{map[string]any{
			"guarantee": "provider_idempotency_key",
			"code":      "provider_idempotency_unavailable",
			"reason":    "The fixture provider does not expose an idempotency-key endpoint.",
		}}
	})

	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateProceed {
		t.Fatalf("limited producer-valid flow decision = %#v, want PROCEED", verdict)
	}
}

func TestCertificationGateRejectsUnredactedProofBody(t *testing.T) {
	contract := loadCertificationTestContract(t)
	root := writeGreenCertificationFixture(t)
	evidence := readCertificationFixtureObject(t, root, "internal/connectors/certifications/evidence/capability.json")
	proof := evidence["proof"].(map[string]any)
	exchange := proof["http_exchanges"].([]any)[0].(map[string]any)
	request := exchange["request"].(map[string]any)
	body := request["body"].(map[string]any)
	body["value"] = map[string]any{"token": "redacted-but-not-a-fingerprint"}
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/evidence/capability.json", evidence)

	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateHalt {
		t.Fatalf("unredacted proof decision = %#v, want HALT", verdict)
	}
}

func TestCertificationGateFailsClosedForSchemaAndAdapterInputDrift(t *testing.T) {
	contract := loadCertificationTestContract(t)
	tests := []struct {
		name   string
		mutate func(*testing.T, string) CertificationGateRequest
	}{
		{
			name: "missing capability schema version",
			mutate: func(t *testing.T, root string) CertificationGateRequest {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
				delete(matrix, "schema_version")
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", matrix)
				return certificationGateRequest("integrate_sub_pr")
			},
		},
		{
			name: "unknown capability field",
			mutate: func(t *testing.T, root string) CertificationGateRequest {
				matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
				matrix["adapter_local_escape"] = true
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", matrix)
				return certificationGateRequest("integrate_sub_pr")
			},
		},
		{
			name: "unknown accepted evidence proof schema",
			mutate: func(t *testing.T, root string) CertificationGateRequest {
				evidence := readCertificationFixtureObject(t, root, "internal/connectors/certifications/evidence/capability.json")
				proof := evidence["proof"].(map[string]any)
				proof["redaction_strategy"] = "future_proof_schema_v2"
				writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/evidence/capability.json", evidence)
				return certificationGateRequest("integrate_sub_pr")
			},
		},
		{
			name: "omitted adapter evidence directory",
			mutate: func(_ *testing.T, _ string) CertificationGateRequest {
				request := certificationGateRequest("integrate_sub_pr")
				request.Inputs.EvidenceDirectory = ""
				return request
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			request := test.mutate(t, root)
			verdict, err := EvaluateCertificationGate(root, contract, request)
			if err != nil {
				t.Fatal(err)
			}
			if verdict.Decision != CertificationGateHalt {
				t.Fatalf("decision = %q, want HALT; failures=%#v", verdict.Decision, verdict.Failures)
			}
		})
	}

	verdict, err := EvaluateCertificationGateJSON(t.TempDir(), contract, []byte(`{"schema_version":1,"connector":"github","transition":"integrate_sub_pr","inputs":{"capability_matrix":"internal/connectors/certifications/capability-matrix.json","flow_matrix":"internal/connectors/certifications/flow-matrix.json","status":"internal/connectors/certifications/status.json","evidence_directory":"internal/connectors/certifications/evidence"},"adapter_local_escape":true}`))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateHalt || verdict.Failures[0].ID != "request/decode" {
		t.Fatalf("unknown adapter request field verdict = %#v, want HALT request/decode", verdict)
	}
}

func TestCertificationGateEnforcesEveryProtectedTransition(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	baselineRoot := repositoryRoot(t)
	greenRoot := writeGreenCertificationFixture(t)
	for _, transition := range contract.CertificationGate.EnforcedTransitions {
		t.Run(transition, func(t *testing.T) {
			blocked, err := EnforceCertificationGate(baselineRoot, contract, certificationGateRequestFor(contract, transition))
			if err == nil {
				t.Fatalf("baseline unexpectedly passed protected %s transition", transition)
			}
			var blockedError *CertificationGateBlockedError
			if !errors.As(err, &blockedError) || blocked.Decision != CertificationGateRetry {
				t.Fatalf("baseline enforcement = verdict=%#v err=%v, want RETRY blocked error", blocked, err)
			}
			assertCertificationFailureID(t, blocked, "capability/github/capability:check/live_evidence")

			fixtureContract := loadCertificationTestContract(t)
			proceed, err := EnforceCertificationGate(greenRoot, fixtureContract, certificationGateRequest(transition))
			if err != nil || proceed.Decision != CertificationGateProceed {
				t.Fatalf("all-green enforcement = verdict=%#v err=%v, want PROCEED", proceed, err)
			}
		})
	}
}

func TestCertificationGateCommandIsRenderedForEveryHarness(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	argv, err := marshalArgv(contract.CertificationGate.Command.Argv)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range contract.Projections {
		t.Run(target.Harness+"/"+target.Role, func(t *testing.T) {
			rendered, err := RenderProjection(contract, target)
			if err != nil {
				t.Fatal(err)
			}
			if target.Harness == "codex" {
				rendered = []byte(parseCodexProjection(t, rendered).GetString("developer_instructions"))
			}
			if !bytes.Contains(rendered, []byte(argv)) {
				t.Fatalf("%s projection omits certification gate argv %s", target.Harness, argv)
			}
		})
	}
}

func TestCertificationGateCurrentBaselineRejectsWithoutBreakingStructuralContractCheck(t *testing.T) {
	root := repositoryRoot(t)
	contract := loadRepositoryContract(t, root)
	if err := CheckRoot(context.Background(), root); err != nil {
		t.Fatalf("generic contract/projection check must remain structural with zero certified connectors: %v", err)
	}
	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequestFor(contract, "integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateRetry {
		t.Fatalf("current baseline decision = %q, want RETRY; failures=%#v", verdict.Decision, verdict.Failures)
	}
}

func TestCertificationGateEvaluationIsReadOnly(t *testing.T) {
	contract := loadCertificationTestContract(t)
	root := writeGreenCertificationFixture(t)
	before := certificationFixtureSnapshot(t, root)
	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateProceed {
		t.Fatalf("read-only fixture decision = %#v", verdict)
	}
	after := certificationFixtureSnapshot(t, root)
	if !slices.Equal(before, after) {
		t.Fatalf("read-only evaluator changed generated inputs or evidence\nbefore=%#v\nafter=%#v", before, after)
	}
}

func certificationGateInputsForTest() CertificationGateInputs {
	return CertificationGateInputs{
		CapabilityMatrix:  "internal/connectors/certifications/capability-matrix.json",
		FlowMatrix:        "internal/connectors/certifications/flow-matrix.json",
		Status:            "internal/connectors/certifications/status.json",
		EvidenceDirectory: "internal/connectors/certifications/evidence",
	}
}

func loadCertificationTestContract(t *testing.T) *Contract {
	t.Helper()
	contract := loadRepositoryContract(t, repositoryRoot(t))
	contract.CertificationGate.Inputs = certificationGateInputsForTest()
	contract.CertificationGate.InputFields = []string{"schema_version", "connector", "transition", "inputs.capability_matrix", "inputs.flow_matrix", "inputs.status", "inputs.evidence_directory"}
	return contract
}

func certificationGateRequestFor(contract *Contract, transition string) CertificationGateRequest {
	return CertificationGateRequest{
		SchemaVersion: 1,
		Connector:     "github",
		Transition:    transition,
		Inputs:        contract.CertificationGate.Inputs,
	}
}

func certificationGateRequest(transition string) CertificationGateRequest {
	return CertificationGateRequest{
		SchemaVersion: 1,
		Connector:     "github",
		Transition:    transition,
		Inputs:        certificationGateInputsForTest(),
	}
}

func assertCertificationFailureID(t *testing.T, verdict CertificationGateVerdict, want string) {
	t.Helper()
	for _, failure := range verdict.Failures {
		if failure.ID == want {
			return
		}
	}
	t.Fatalf("verdict failures %#v do not include exact ID %q", verdict.Failures, want)
}

func writeGreenCertificationFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	proof := certificationTestProof(false)
	flowProof := certificationTestProof(true)

	capabilityEvidence := certificationTestEvidence("capability.json", proof, map[string]any{
		"scope":         "capability",
		"connector":     "github",
		"function_kind": "capability:check",
	})
	workflowEvidence := certificationTestEvidence("workflow.json", proof, map[string]any{
		"scope":         "workflow",
		"connector":     "github",
		"workflow_kind": "etl",
	})
	type syncPrimitiveFixture struct {
		id              string
		integrationType string
		capability      string
		direction       string
	}
	syncPrimitives := []syncPrimitiveFixture{
		{id: "api_read_into_warehouse", integrationType: "api", capability: "read", direction: "into_warehouse"},
		{id: "api_write_from_warehouse", integrationType: "api", capability: "write", direction: "from_warehouse"},
		{id: "database_read_into_warehouse", integrationType: "database", capability: "read", direction: "into_warehouse"},
		{id: "database_write_from_warehouse", integrationType: "database", capability: "write", direction: "from_warehouse"},
	}
	syncEvidence := make([]map[string]any, 0, len(syncPrimitives))
	syncCells := make([]any, 0, len(syncPrimitives))
	syncPrimitiveDefinitions := make([]any, 0, len(syncPrimitives))
	for _, primitive := range syncPrimitives {
		evidence := certificationTestEvidence("sync-"+primitive.id+".json", proof, map[string]any{
			"scope":     "sync_mode",
			"connector": "github",
			"sync_mode": "full_overwrite",
			"primitive": primitive.id,
		})
		syncEvidence = append(syncEvidence, evidence)
		syncCells = append(syncCells, certificationTestCell(map[string]any{
			"sync_mode":     "full_overwrite",
			"primitive":     primitive.id,
			"live_evidence": []any{certificationTestEvidencePointer(evidence, proof)},
		}))
		syncPrimitiveDefinitions = append(syncPrimitiveDefinitions, map[string]any{
			"id": primitive.id, "integration_type": primitive.integrationType, "capability": primitive.capability,
			"warehouse_direction": primitive.direction, "discovery_source": "test",
		})
	}
	flowEvidence := certificationTestEvidence("flow.json", flowProof, map[string]any{
		"scope":       "flow",
		"source":      "github",
		"destination": "github",
		"flow_kind":   "api_to_api",
	})

	evidenceRecords := []map[string]any{capabilityEvidence, workflowEvidence}
	evidenceRecords = append(evidenceRecords, syncEvidence...)
	evidenceRecords = append(evidenceRecords, flowEvidence)
	for _, evidence := range evidenceRecords {
		record := evidence["record"].(string)
		sidecar := make(map[string]any, len(evidence)-1)
		for key, value := range evidence {
			if key != "record" {
				sidecar[key] = value
			}
		}
		writeCertificationFixtureJSON(t, root, record, sidecar)
	}

	capabilityCell := certificationTestCell(map[string]any{
		"function_kind": "capability:check",
		"live_evidence": []any{certificationTestEvidencePointer(capabilityEvidence, proof)},
	})
	workflowCell := certificationTestCell(map[string]any{
		"workflow_kind": "etl",
		"live_evidence": []any{certificationTestEvidencePointer(workflowEvidence, proof)},
	})
	flowCell := certificationTestCell(map[string]any{
		"live_evidence": []any{certificationTestEvidencePointer(flowEvidence, flowProof)},
	})
	flowKinds := []any{
		map[string]any{"id": "api_to_api", "source_role": "api_source", "destination_role": "api_destination"},
		map[string]any{"id": "api_to_database", "source_role": "api_source", "destination_role": "database_destination"},
		map[string]any{"id": "database_to_api", "source_role": "database_source", "destination_role": "api_destination"},
		map[string]any{"id": "database_to_database", "source_role": "database_source", "destination_role": "database_destination"},
	}
	pairSets := []any{
		map[string]any{
			"flow_kind": "api_to_api", "mediator": "local_parquet_warehouse", "source_connectors": []any{"github"}, "destination_connectors": []any{"github"}, "cell": flowCell,
		},
		map[string]any{
			"flow_kind": "api_to_database", "mediator": "local_parquet_warehouse", "source_connectors": []any{"github"}, "destination_connectors": []any{"github"},
			"cell": certificationTestNotApplicableCell("destination_unsupported_database_destination", "destination fixture connector has no database destination"),
		},
		map[string]any{
			"flow_kind": "database_to_api", "mediator": "local_parquet_warehouse", "source_connectors": []any{"github"}, "destination_connectors": []any{"github"},
			"cell": certificationTestNotApplicableCell("source_unsupported_database_source", "source fixture connector has no database source"),
		},
		map[string]any{
			"flow_kind": "database_to_database", "mediator": "local_parquet_warehouse", "source_connectors": []any{"github"}, "destination_connectors": []any{"github"},
			"cell": certificationTestNotApplicableCell("source_and_destination_roles_inapplicable", "source database_source and destination database_destination roles are not applicable"),
		},
	}
	capabilityBaseline := map[string]any{
		"connectors":          1,
		"capability_complete": 1,
		"per_kind": []any{map[string]any{
			"function_kind": "capability:check", "connectors": 1, "applicable": 1, "declared": 1, "implemented": 1, "fixture_tested": 1, "live_tested": 1, "complete": 1,
		}},
	}
	flowBaseline := map[string]any{
		"connectors": 1,
		"certified":  1,
		"workflows": []any{map[string]any{
			"workflow_kind": "etl", "connectors": 1, "applicable": 1, "declared": 1, "implemented": 1, "fixture_tested": 1, "live_tested": 1, "complete": 1,
		}},
		"sync_modes": []any{
			map[string]any{"sync_mode": "full_overwrite", "primitive": "api_read_into_warehouse", "connectors": 1, "applicable": 1, "declared": 1, "implemented": 1, "fixture_tested": 1, "live_tested": 1, "complete": 1},
			map[string]any{"sync_mode": "full_overwrite", "primitive": "api_write_from_warehouse", "connectors": 1, "applicable": 1, "declared": 1, "implemented": 1, "fixture_tested": 1, "live_tested": 1, "complete": 1},
			map[string]any{"sync_mode": "full_overwrite", "primitive": "database_read_into_warehouse", "connectors": 1, "applicable": 1, "declared": 1, "implemented": 1, "fixture_tested": 1, "live_tested": 1, "complete": 1},
			map[string]any{"sync_mode": "full_overwrite", "primitive": "database_write_from_warehouse", "connectors": 1, "applicable": 1, "declared": 1, "implemented": 1, "fixture_tested": 1, "live_tested": 1, "complete": 1},
		},
		"per_kind": []any{
			map[string]any{"flow_kind": "api_to_api", "pairs": 1, "applicable": 1, "declared": 1, "implemented": 1, "fixture_tested": 1, "live_tested": 1, "complete": 1},
			map[string]any{"flow_kind": "api_to_database", "pairs": 1, "applicable": 0, "declared": 0, "implemented": 0, "fixture_tested": 0, "live_tested": 0, "complete": 0},
			map[string]any{"flow_kind": "database_to_api", "pairs": 1, "applicable": 0, "declared": 0, "implemented": 0, "fixture_tested": 0, "live_tested": 0, "complete": 0},
			map[string]any{"flow_kind": "database_to_database", "pairs": 1, "applicable": 0, "declared": 0, "implemented": 0, "fixture_tested": 0, "live_tested": 0, "complete": 0},
		},
	}

	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", map[string]any{
		"schema_version":    1,
		"generated_command": certificationGeneratedCommand,
		"function_kinds": []any{map[string]any{
			"id": "capability:check", "category": "capability", "name": "check", "discovery_source": "test",
		}},
		"connectors": []any{map[string]any{
			"name": "github", "integration_type": "api", "capability_complete": true, "cells": []any{capabilityCell},
		}},
		"legacy_certification_inputs": map[string]any{"ignored": true, "files": []any{}},
		"baseline":                    capabilityBaseline,
	})
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", map[string]any{
		"schema_version":    1,
		"generated_command": certificationGeneratedCommand,
		"mediator":          "local_parquet_warehouse",
		"flow_kinds":        flowKinds,
		"workflow_kinds":    []any{map[string]any{"id": "etl", "discovery_source": "test"}},
		"workflows": []any{map[string]any{
			"connector": "github", "complete": true, "cells": []any{workflowCell},
		}},
		"sync_mode_kinds": []any{map[string]any{"id": "full_overwrite", "discovery_source": "test"}},
		"sync_primitives": syncPrimitiveDefinitions,
		"sync_mode_cells": []any{map[string]any{
			"connector": "github", "complete": true, "cells": syncCells,
		}},
		"connector_roles": []any{map[string]any{
			"connector": "github", "roles": []any{
				map[string]any{"role": "api_source", "applicable": true, "declared": true, "implemented": true},
				map[string]any{"role": "api_destination", "applicable": true, "declared": true, "implemented": true},
				map[string]any{"role": "database_source", "applicable": false, "declared": false, "implemented": false, "not_applicable": map[string]any{"code": "unsupported_database_source", "reason": "fixture connector has no database source"}},
				map[string]any{"role": "database_destination", "applicable": false, "declared": false, "implemented": false, "not_applicable": map[string]any{"code": "unsupported_database_destination", "reason": "fixture connector has no database destination"}},
			},
		}},
		"pair_sets":          pairSets,
		"pair_overrides":     []any{},
		"connector_statuses": []any{map[string]any{"connector": "github", "certified": true, "label": "CERTIFIED"}},
		"baseline":           flowBaseline,
	})
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/status.json", map[string]any{
		"schema_version":    1,
		"generated_command": certificationGeneratedCommand,
		"connectors":        []any{map[string]any{"connector": "github", "certified": true, "label": "CERTIFIED"}},
	})
	return root
}

func certificationTestCell(overrides map[string]any) map[string]any {
	cell := map[string]any{
		"applicable": true, "declared": true, "implemented": true, "fixture_tested": true, "live_tested": true,
		"fixture_evidence": []any{"fixture.json"}, "live_evidence": []any{},
	}
	for key, value := range overrides {
		cell[key] = value
	}
	return cell
}

func certificationTestNotApplicableCell(code, reason string) map[string]any {
	return map[string]any{
		"applicable": false, "declared": false, "implemented": false, "fixture_tested": false, "live_tested": false,
		"fixture_evidence": []any{}, "live_evidence": []any{},
		"not_applicable": map[string]any{"code": code, "reason": reason},
	}
}

func cloneCertificationFixtureObject(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var copy map[string]any
	if err := json.Unmarshal(raw, &copy); err != nil {
		t.Fatal(err)
	}
	return copy
}

func certificationTestEvidence(record string, proof map[string]any, scope map[string]any) map[string]any {
	evidence := map[string]any{
		"record":                 "internal/connectors/certifications/evidence/" + record,
		"schema_version":         2,
		"status":                 "passed",
		"credential_scope":       "observed_operations",
		"credential_note":        certificationCredentialNote,
		"credential_scope_proof": "protocol_exchanges",
		"provider":               "test-provider",
		"executed_at":            "2026-08-11T00:00:00Z",
		"run_id":                 "test-run-1",
		"proof":                  proof,
	}
	for key, value := range scope {
		evidence[key] = value
	}
	return evidence
}

func certificationTestEvidencePointer(evidence map[string]any, proof map[string]any) map[string]any {
	return map[string]any{
		"record":                 evidence["record"],
		"provider":               evidence["provider"],
		"executed_at":            evidence["executed_at"],
		"run_id":                 evidence["run_id"],
		"credential_scope":       evidence["credential_scope"],
		"credential_note":        evidence["credential_note"],
		"credential_scope_proof": evidence["credential_scope_proof"],
		"proof":                  proof,
	}
}

func certificationTestProof(flow bool) map[string]any {
	fingerprint := "{{pmcertfp:v1:" + strings.Repeat("a", 64) + "}}"
	body := map[string]any{"encoding": "json", "value": map[string]any{}, "original_bytes": 2, "truncated": false}
	proof := map[string]any{
		"redaction_strategy":      "repository_salted_hmac_sha256_v1",
		"pm_binary_sha256":        strings.Repeat("b", 64),
		"pm_command_fingerprint":  fingerprint,
		"credential_fingerprints": []any{fingerprint},
		"http_exchanges": []any{
			map[string]any{
				"operation": "warehouse_readback", "request": map[string]any{"method": "GET", "target": fingerprint, "query": []any{}, "headers": []any{}, "body": body},
				"response": map[string]any{"status": 200, "headers": []any{}, "body": body},
			},
			map[string]any{
				"operation": "destination_readback", "request": map[string]any{"method": "GET", "target": fingerprint, "query": []any{}, "headers": []any{}, "body": body},
				"response": map[string]any{"status": 200, "headers": []any{}, "body": body},
			},
		},
		"database_exchanges": []any{},
	}
	if flow {
		proof["flow"] = map[string]any{
			"pm_command_fingerprint":         fingerprint,
			"mediator":                       "local_parquet_warehouse",
			"warehouse_readback_operation":   "warehouse_readback",
			"destination_readback_operation": "destination_readback",
			"delivery": map[string]any{
				"resumable": true, "receipt_backed": true, "checkpointed": true, "replay_identity": true, "provider_idempotency_key": true, "limitations": []any{},
			},
		}
	}
	return proof
}

func writeCertificationFixtureJSON(t *testing.T, root, relativePath string, value any) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func certificationFixtureCapabilityCell(t *testing.T, root string) map[string]any {
	t.Helper()
	matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
	connectors := matrix["connectors"].([]any)
	return connectors[0].(map[string]any)["cells"].([]any)[0].(map[string]any)
}

func mutateCertificationFixtureCapabilityProof(t *testing.T, root string, mutate func(map[string]any)) {
	t.Helper()
	matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
	pointer := matrix["connectors"].([]any)[0].(map[string]any)["cells"].([]any)[0].(map[string]any)["live_evidence"].([]any)[0].(map[string]any)
	mutate(pointer["proof"].(map[string]any))
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", matrix)

	evidence := readCertificationFixtureObject(t, root, "internal/connectors/certifications/evidence/capability.json")
	mutate(evidence["proof"].(map[string]any))
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/evidence/capability.json", evidence)
}

func mutateCertificationFixtureFlowProof(t *testing.T, root string, mutate func(map[string]any)) {
	t.Helper()
	matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/flow-matrix.json")
	pointer := matrix["pair_sets"].([]any)[0].(map[string]any)["cell"].(map[string]any)["live_evidence"].([]any)[0].(map[string]any)
	mutate(pointer["proof"].(map[string]any))
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", matrix)

	evidence := readCertificationFixtureObject(t, root, "internal/connectors/certifications/evidence/flow.json")
	mutate(evidence["proof"].(map[string]any))
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/evidence/flow.json", evidence)
}

func writeCertificationFixtureCapabilityCell(t *testing.T, root string, cell map[string]any) {
	t.Helper()
	matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
	matrix["connectors"].([]any)[0].(map[string]any)["cells"].([]any)[0] = cell
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", matrix)
}

func rewriteCertificationFixtureDerivedReports(t *testing.T, root string) {
	t.Helper()
	contract := loadCertificationTestContract(t)
	capabilities, err := loadCapabilityMatrix(root, contract.CertificationGate)
	if err != nil {
		t.Fatal(err)
	}
	flow, err := loadFlowMatrix(root, contract.CertificationGate)
	if err != nil {
		t.Fatal(err)
	}
	status, err := loadStatusArtifact(root, contract.CertificationGate)
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := loadAcceptedCertificationEvidence(root, contract.CertificationGate)
	if err != nil {
		t.Fatal(err)
	}
	artifacts := certificationArtifacts{capabilities: capabilities, flow: flow, status: status, evidence: evidence}
	reports := artifacts.deriveReports()
	for index := range capabilities.Connectors {
		capabilities.Connectors[index].CapabilityComplete = reports.capabilityComplete[capabilities.Connectors[index].Name]
	}
	capabilities.Baseline = reports.capabilityBaseline
	for index := range flow.Workflows {
		flow.Workflows[index].Complete = reports.workflowComplete[flow.Workflows[index].Connector]
	}
	for index := range flow.SyncModeCells {
		flow.SyncModeCells[index].Complete = reports.syncModeComplete[flow.SyncModeCells[index].Connector]
	}
	for index := range flow.ConnectorStatuses {
		flow.ConnectorStatuses[index] = reports.statuses[flow.ConnectorStatuses[index].Connector]
	}
	flow.Baseline = reports.flowBaseline
	for index := range status.Connectors {
		status.Connectors[index] = reports.statuses[status.Connectors[index].Connector]
	}
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", capabilities)
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", flow)
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/status.json", status)
}

func readCertificationFixtureObject(t *testing.T, root, relativePath string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func certificationFixtureSnapshot(t *testing.T, root string) []string {
	t.Helper()
	entries := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		entries = append(entries, filepath.ToSlash(relative)+"\x00"+string(contents))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return entries
}
