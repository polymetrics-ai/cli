package agentcontract

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	certificationGeneratedCommand = "go run ./cmd/connectorgen certification-matrix"
	certificationCredentialNote   = "Certification used a full-parity credential; a narrower credential exposes a subset of this certified surface."
)

func TestEvaluateCertificationGateGitHubBaselineAndGreenFixture(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	request := CertificationGateRequest{
		SchemaVersion: 1,
		Connector:     "github",
		Transition:    "integrate_sub_pr",
	}

	baseline, err := EvaluateCertificationGate(repositoryRoot(t), contract, request)
	if err != nil {
		t.Fatalf("evaluate current generated GitHub baseline: %v", err)
	}
	if baseline.Decision != CertificationGateRetry {
		t.Fatalf("current GitHub decision = %q, want %q", baseline.Decision, CertificationGateRetry)
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
		delete(evidence, "record")
		writeCertificationFixtureJSON(t, root, record, evidence)
	}

	capabilityCell := certificationTestCell(map[string]any{
		"function_kind": "capability:check",
		"live_evidence": []any{certificationEvidencePointer(capabilityEvidence, proof)},
	})
	workflowCell := certificationTestCell(map[string]any{
		"workflow_kind": "etl",
		"live_evidence": []any{certificationEvidencePointer(workflowEvidence, proof)},
	})
	syncCell := certificationTestCell(map[string]any{
		"sync_mode":     "full_overwrite",
		"primitive":     "api_read_into_warehouse",
		"live_evidence": []any{certificationEvidencePointer(syncEvidence, proof)},
	})
	flowCell := certificationTestCell(map[string]any{
		"live_evidence": []any{certificationEvidencePointer(flowEvidence, flowProof)},
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

func certificationEvidencePointer(evidence map[string]any, proof map[string]any) map[string]any {
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
		"http_exchanges": []any{map[string]any{
			"operation": "request", "request": map[string]any{"method": "GET", "target": fingerprint, "query": []any{}, "headers": []any{}, "body": body},
			"response": map[string]any{"status": 200, "headers": []any{}, "body": body},
		}},
		"database_exchanges": []any{},
	}
	if flow {
		proof["flow"] = map[string]any{
			"pm_command_fingerprint":         fingerprint,
			"mediator":                       "local_parquet_warehouse",
			"warehouse_readback_operation":   "request",
			"destination_readback_operation": "request",
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
