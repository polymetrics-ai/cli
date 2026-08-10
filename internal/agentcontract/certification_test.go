package agentcontract

import (
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
	certificationGeneratedCommand = "go run ./cmd/connectorgen certification-matrix"
	certificationCredentialNote   = "Certification used a full-parity credential; a narrower credential exposes a subset of this certified surface."
)

func TestEvaluateCertificationGateGitHubBaselineAndGreenFixture(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	request := certificationGateRequest("integrate_sub_pr")

	baseline, err := EvaluateCertificationGate(repositoryRoot(t), contract, request)
	if err != nil {
		t.Fatalf("evaluate current generated GitHub baseline: %v", err)
	}
	if baseline.Decision != CertificationGateRetry {
		t.Fatalf("current GitHub decision = %q, want %q; failures=%#v", baseline.Decision, CertificationGateRetry, baseline.Failures)
	}
	assertCertificationFailureID(t, baseline, "capability/github/capability:check/live_evidence")

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
	contract := loadRepositoryContract(t, repositoryRoot(t))
	for _, criterion := range []string{"declared", "implemented", "fixture_tested", "live_tested", "live_evidence"} {
		t.Run(criterion, func(t *testing.T) {
			root := writeGreenCertificationFixture(t)
			cell := certificationFixtureCapabilityCell(t, root)
			switch criterion {
			case "live_evidence":
				cell[criterion] = []any{}
			default:
				cell[criterion] = false
			}
			writeCertificationFixtureCapabilityCell(t, root, cell)

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

func TestCertificationGateDoesNotPromoteReachabilityOrImplementedWithoutLiveEvidence(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := writeGreenCertificationFixture(t)
	cell := certificationFixtureCapabilityCell(t, root)
	cell["declared"] = true
	cell["implemented"] = true
	cell["fixture_tested"] = true
	cell["live_tested"] = false
	cell["live_evidence"] = []any{}
	writeCertificationFixtureCapabilityCell(t, root, cell)

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
	contract := loadRepositoryContract(t, repositoryRoot(t))
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

func TestCertificationGateFailsClosedForSchemaAndAdapterInputDrift(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
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
			blocked, err := EnforceCertificationGate(baselineRoot, contract, certificationGateRequest(transition))
			if err == nil {
				t.Fatalf("baseline unexpectedly passed protected %s transition", transition)
			}
			var blockedError *CertificationGateBlockedError
			if !errors.As(err, &blockedError) || blocked.Decision != CertificationGateRetry {
				t.Fatalf("baseline enforcement = verdict=%#v err=%v, want RETRY blocked error", blocked, err)
			}
			assertCertificationFailureID(t, blocked, "capability/github/capability:check/live_evidence")

			proceed, err := EnforceCertificationGate(greenRoot, contract, certificationGateRequest(transition))
			if err != nil || proceed.Decision != CertificationGateProceed {
				t.Fatalf("all-green enforcement = verdict=%#v err=%v, want PROCEED", proceed, err)
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
	verdict, err := EvaluateCertificationGate(root, contract, certificationGateRequest("integrate_sub_pr"))
	if err != nil {
		t.Fatal(err)
	}
	if verdict.Decision != CertificationGateRetry {
		t.Fatalf("current baseline decision = %q, want RETRY; failures=%#v", verdict.Decision, verdict.Failures)
	}
}

func TestCertificationGateEvaluationIsReadOnly(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
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
	syncEvidence := certificationTestEvidence("sync.json", proof, map[string]any{
		"scope":     "sync_mode",
		"connector": "github",
		"sync_mode": "full_overwrite",
		"primitive": "api_read_into_warehouse",
	})
	flowEvidence := certificationTestEvidence("flow.json", flowProof, map[string]any{
		"scope":       "flow",
		"source":      "github",
		"destination": "github",
		"flow_kind":   "api_to_api",
	})

	for _, evidence := range []map[string]any{capabilityEvidence, workflowEvidence, syncEvidence, flowEvidence} {
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
	syncCell := certificationTestCell(map[string]any{
		"sync_mode":     "full_overwrite",
		"primitive":     "api_read_into_warehouse",
		"live_evidence": []any{certificationTestEvidencePointer(syncEvidence, proof)},
	})
	flowCell := certificationTestCell(map[string]any{
		"live_evidence": []any{certificationTestEvidencePointer(flowEvidence, flowProof)},
	})

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
		"baseline":                    map[string]any{"connectors": 1, "capability_complete": 1, "per_kind": []any{}},
	})
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/flow-matrix.json", map[string]any{
		"schema_version":    1,
		"generated_command": certificationGeneratedCommand,
		"mediator":          "local_parquet_warehouse",
		"flow_kinds": []any{map[string]any{
			"id": "api_to_api", "source_role": "api_source", "destination_role": "api_destination",
		}},
		"workflow_kinds": []any{map[string]any{"id": "etl", "discovery_source": "test"}},
		"workflows": []any{map[string]any{
			"connector": "github", "complete": true, "cells": []any{workflowCell},
		}},
		"sync_mode_kinds": []any{map[string]any{"id": "full_overwrite", "discovery_source": "test"}},
		"sync_primitives": []any{map[string]any{
			"id": "api_read_into_warehouse", "integration_type": "api", "capability": "read", "warehouse_direction": "into_warehouse", "discovery_source": "test",
		}},
		"sync_mode_cells": []any{map[string]any{
			"connector": "github", "complete": true, "cells": []any{syncCell},
		}},
		"connector_roles": []any{map[string]any{
			"connector": "github", "roles": []any{
				map[string]any{"role": "api_source", "applicable": true, "declared": true, "implemented": true},
				map[string]any{"role": "api_destination", "applicable": true, "declared": true, "implemented": true},
			},
		}},
		"pair_sets": []any{map[string]any{
			"flow_kind": "api_to_api", "mediator": "local_parquet_warehouse", "source_connectors": []any{"github"}, "destination_connectors": []any{"github"}, "cell": flowCell,
		}},
		"pair_overrides":     []any{},
		"connector_statuses": []any{map[string]any{"connector": "github", "certified": true, "label": "CERTIFIED"}},
		"baseline":           map[string]any{"connectors": 1, "certified": 1, "workflows": []any{}, "sync_modes": []any{}, "per_kind": []any{}},
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

func certificationTestEvidence(record string, proof map[string]any, scope map[string]any) map[string]any {
	evidence := map[string]any{
		"record":           "internal/connectors/certifications/evidence/" + record,
		"schema_version":   1,
		"status":           "passed",
		"credential_scope": "full_parity",
		"credential_note":  certificationCredentialNote,
		"provider":         "test-provider",
		"executed_at":      "2026-08-11T00:00:00Z",
		"run_id":           "test-run-1",
		"proof":            proof,
	}
	for key, value := range scope {
		evidence[key] = value
	}
	return evidence
}

func certificationTestEvidencePointer(evidence map[string]any, proof map[string]any) map[string]any {
	return map[string]any{
		"record":           evidence["record"],
		"provider":         evidence["provider"],
		"executed_at":      evidence["executed_at"],
		"run_id":           evidence["run_id"],
		"credential_scope": evidence["credential_scope"],
		"credential_note":  evidence["credential_note"],
		"proof":            proof,
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

func writeCertificationFixtureCapabilityCell(t *testing.T, root string, cell map[string]any) {
	t.Helper()
	matrix := readCertificationFixtureObject(t, root, "internal/connectors/certifications/capability-matrix.json")
	matrix["connectors"].([]any)[0].(map[string]any)["cells"].([]any)[0] = cell
	writeCertificationFixtureJSON(t, root, "internal/connectors/certifications/capability-matrix.json", matrix)
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
