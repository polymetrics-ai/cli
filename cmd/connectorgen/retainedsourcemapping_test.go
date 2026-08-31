package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// TestRetainedSourceMappingCommandIsRegistered starts red: a retained source
// mapping check is an authoring-only source-accounting action. It must not be
// routed through source-import, materialize, or runtime bundle loading.
func TestRetainedSourceMappingCommandIsRegistered(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if exitCode := run([]string{"retained-source-mapping", "bitbucket", "--check"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("retained-source-mapping exit=%d stderr=%q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "mapping-only") {
		t.Fatalf("retained-source-mapping output %q does not state mapping-only boundary", stdout.String())
	}
}

func TestRetainedSourceMappingFrozenV2Cohort(t *testing.T) {
	root := retainedSourceMappingTestRoot(t)
	want := map[string]int{
		"bitbucket": 297,
		"circleci":  111,
		"dockerhub": 54,
		"jira":      617,
		"notion":    49,
		"sentry":    223,
		"stripe":    589,
		"vercel":    400,
	}
	total := 0
	for connector, expected := range want {
		t.Run(connector, func(t *testing.T) {
			lock := retainedSourceMappingTestLock(t, root, connector)
			if sourceImportLockHasCanonicalEvidenceContract(lock) {
				t.Fatalf("%s unexpectedly passes normal canonical_evidence admission", connector)
			}
			result, err := retainedSourceMappingFromRepository(root, connector)
			if err != nil {
				t.Fatalf("build retained source mapping: %v", err)
			}
			if result.Report.SourceOperations != expected || result.Report.VerifiedSourceOperations != expected {
				t.Fatalf("source operations=%d verified=%d, want %d/%d", result.Report.SourceOperations, result.Report.VerifiedSourceOperations, expected, expected)
			}
			if result.Report.ExecutableDeclarations != 0 || !result.Report.MappingOnly {
				t.Fatalf("mapping-only report made an executable claim: %+v", result.Report)
			}
			if err := result.Contract.ValidateRetentionOnly(); err != nil {
				t.Fatalf("validate retention-only contract: %v", err)
			}
			if err := result.Contract.ReconcileSourceOperations(enabledContractSourceOperations(lock)); err != nil {
				t.Fatalf("reconcile exact source IDs: %v", err)
			}
			for _, lane := range result.Contract.Lanes {
				if lane.State == connectors.EnabledLaneImplemented || lane.Source.Implemented != 0 || lane.Source.UnmappedMapping != 0 || lane.Transport != nil || len(lane.Warehouse) != 0 {
					t.Fatalf("%s lane escaped source-only boundary: %+v", lane.Name, lane)
				}
				if len(lane.Artifacts) != 1 || lane.Artifacts[0] != result.Contract.SourceLock.Path {
					t.Fatalf("%s lane retains unexpected artifact binding: %+v", lane.Name, lane.Artifacts)
				}
			}
		})
		total += expected
	}
	if total != 2340 {
		t.Fatalf("frozen retained v2 denominator=%d, want 2340", total)
	}
}

func TestRetainedSourceMappingSidecarBytesAreDeterministicWithoutWriting(t *testing.T) {
	root := retainedSourceMappingTestRoot(t)
	first, err := retainedSourceMappingFromRepository(root, "bitbucket")
	if err != nil {
		t.Fatalf("first source-only build: %v", err)
	}
	second, err := retainedSourceMappingFromRepository(root, "bitbucket")
	if err != nil {
		t.Fatalf("second source-only build: %v", err)
	}
	firstRaw, err := retainedSourceMappingContractJSON(first.Contract)
	if err != nil {
		t.Fatalf("encode first retention-only contract: %v", err)
	}
	secondRaw, err := retainedSourceMappingContractJSON(second.Contract)
	if err != nil {
		t.Fatalf("encode second retention-only contract: %v", err)
	}
	if !bytes.Equal(firstRaw, secondRaw) {
		t.Fatal("retention-only sidecar bytes are not deterministic")
	}
	if !bytes.Contains(firstRaw, []byte(`"retention_only": true`)) || bytes.Contains(firstRaw, []byte(`"implemented": 1`)) {
		t.Fatalf("deterministic bytes make an invalid execution claim: %s", firstRaw)
	}
}

func TestRetainedSourceMappingPreservesCanonicalAdmissionBoundary(t *testing.T) {
	root := retainedSourceMappingTestRoot(t)
	lock := retainedSourceMappingTestLock(t, root, "bitbucket")
	if sourceImportLockHasCanonicalEvidenceContract(lock) {
		t.Fatal("frozen Bitbucket lock unexpectedly passed normal canonical admission")
	}
	if err := retainedSourceMappingEligible(lock); err != nil {
		t.Fatalf("frozen Bitbucket retained mapping should be eligible: %v", err)
	}
	lock.Rest.CanonicalEvidence = true
	if err := retainedSourceMappingEligible(lock); err == nil || !strings.Contains(err.Error(), "canonical_evidence") {
		t.Fatalf("canonical evidence overlap error=%v, want explicit rejection", err)
	}
	if !sourceImportLockHasCanonicalEvidenceContract(lock) {
		t.Fatal("normal canonical admission must stay owned by its existing explicit marker")
	}
}

func TestRetainedSourceMappingRejectsMissingEvidenceAndGraphQL(t *testing.T) {
	root := retainedSourceMappingTestRoot(t)
	base := retainedSourceMappingTestLock(t, root, "bitbucket")
	for _, test := range []struct {
		name   string
		mutate func(*sourceImportLock)
		want   string
	}{
		{
			name: "missing source contract",
			mutate: func(lock *sourceImportLock) {
				lock.SourceContract = nil
			},
			want: "source_contract",
		},
		{
			name: "missing source operation",
			mutate: func(lock *sourceImportLock) {
				lock.Rest.Operations[0].SourceOperation = nil
			},
			want: "source_operation",
		},
		{
			name: "GraphQL inventory",
			mutate: func(lock *sourceImportLock) {
				lock.Counts.GraphQLQuery = 1
			},
			want: "zero GraphQL",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			lock := retainedSourceMappingCloneLock(base)
			test.mutate(&lock)
			if err := retainedSourceMappingEligible(lock); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("eligibility error=%v, want %q", err, test.want)
			}
		})
	}
}

func TestRetainedSourceMappingRejectsConflictingProviderOperationID(t *testing.T) {
	root := retainedSourceMappingTestRoot(t)
	lock := retainedSourceMappingCloneLock(retainedSourceMappingTestLock(t, root, "bitbucket"))
	first := &lock.Rest.Operations[0]
	var operation map[string]any
	if err := decodeSourceJSON(first.SourceOperation.Raw, &operation); err != nil {
		t.Fatalf("decode frozen source operation: %v", err)
	}
	operation["operationId"] = "wrong-provider-operation-id"
	raw, err := json.Marshal(operation)
	if err != nil {
		t.Fatalf("encode conflicting source operation: %v", err)
	}
	first.SourceOperation.Raw = raw
	if _, err := retainedSourceMappingVerifySourceEvidence(lock); err == nil || !strings.Contains(err.Error(), "operationId conflicts") {
		t.Fatalf("retained import error=%v, want retained operation identity conflict", err)
	}
}

func TestRetainedSourceMappingSourceProofRequiresExactIDs(t *testing.T) {
	root := retainedSourceMappingTestRoot(t)
	lock := retainedSourceMappingTestLock(t, root, "bitbucket")
	proof, err := retainedSourceMappingVerifySourceEvidence(lock)
	if err != nil {
		t.Fatalf("verify frozen source evidence: %v", err)
	}
	proof.OperationIDs[1] = proof.OperationIDs[0]
	if err := retainedSourceMappingProofIdentities(lock, proof); err == nil || !strings.Contains(err.Error(), "duplicates source ID") {
		t.Fatalf("duplicate source proof error=%v, want exact-ID rejection", err)
	}
	proof, err = retainedSourceMappingVerifySourceEvidence(lock)
	if err != nil {
		t.Fatalf("reverify frozen source evidence: %v", err)
	}
	proof.OperationIDs[0] = "unknown.source.operation"
	if err := retainedSourceMappingProofIdentities(lock, proof); err == nil || !strings.Contains(err.Error(), "unknown source ID") {
		t.Fatalf("unknown source proof error=%v, want exact-ID rejection", err)
	}
}

func TestRetainedSourceMappingCanValidateEvidenceWithoutSchemaMaterialization(t *testing.T) {
	root := retainedSourceMappingTestRoot(t)
	// Docker Hub's retained provider schema contains a component reference that
	// the descriptor importer cannot resolve today. It is nevertheless a valid,
	// exact source-operation inventory for mapping and must remain visible.
	lock := retainedSourceMappingTestLock(t, root, "dockerhub")
	proof, err := retainedSourceMappingVerifySourceEvidence(lock)
	if err != nil {
		t.Fatalf("mapping-only source evidence rejected Docker Hub: %v", err)
	}
	if len(proof.OperationIDs) != 54 {
		t.Fatalf("Docker Hub verified source IDs=%d, want 54", len(proof.OperationIDs))
	}
}

func TestRetainedSourceMappingMatrixDecoderSupportsBothFormsAndCircleCIAlias(t *testing.T) {
	for _, form := range []string{"source_operations", "operations"} {
		t.Run(form, func(t *testing.T) {
			raw := retainedSourceMappingMatrixFixture(t, form, "example.rest.list")
			matrix, err := decodeRetainedSourceMappingMatrix(raw, "example")
			if err != nil {
				t.Fatalf("decode %s matrix: %v", form, err)
			}
			if got := matrix.Rows["example.rest.list"]["etl"]; got != connectors.EnabledLaneMappedUnproven {
				t.Fatalf("ETL state=%q, want mapped_unproven", got)
			}
		})
	}
}

func TestRetainedSourceMappingMatrixDecoderRejectsAmbiguityAndRuntimeClaims(t *testing.T) {
	base := retainedSourceMappingMatrixFixture(t, "operations", "example.rest.list")
	for _, test := range []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "mixed matrix forms",
			mutate: func(root map[string]any) {
				root["source_operations"] = root["operations"]
			},
			want: "exactly one",
		},
		{
			name: "mixed CircleCI aliases",
			mutate: func(root map[string]any) {
				row := root["operations"].([]any)[0].(map[string]any)
				row["source_id"] = "example.rest.list"
			},
			want: "exactly one",
		},
		{
			name: "missing lane cell",
			mutate: func(root map[string]any) {
				row := root["operations"].([]any)[0].(map[string]any)
				row["cells"] = row["cells"].([]any)[:6]
			},
			want: "exactly seven",
		},
		{
			name: "implemented claim",
			mutate: func(root map[string]any) {
				row := root["operations"].([]any)[0].(map[string]any)
				row["cells"].([]any)[0].(map[string]any)["state"] = "implemented"
			},
			want: "cannot claim",
		},
		{
			name: "duplicate source ID",
			mutate: func(root map[string]any) {
				rows := root["operations"].([]any)
				copyRow := retainedSourceMappingDeepCopy(t, rows[0])
				root["operations"] = append(rows, copyRow)
			},
			want: "duplicates source ID",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var root map[string]any
			if err := json.Unmarshal(base, &root); err != nil {
				t.Fatalf("decode fixture: %v", err)
			}
			test.mutate(root)
			raw, err := json.Marshal(root)
			if err != nil {
				t.Fatalf("encode fixture: %v", err)
			}
			if _, err := decodeRetainedSourceMappingMatrix(raw, "example"); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("decode error=%v, want %q", err, test.want)
			}
		})
	}
}

func retainedSourceMappingTestRoot(t *testing.T) string {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	return root
}

func retainedSourceMappingTestLock(t *testing.T, root, connector string) sourceImportLock {
	t.Helper()
	path := filepath.Join(root, "internal", "connectors", "defs", connector, "sources", connector+"-operation-source-lock.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s source lock: %v", connector, err)
	}
	lock, err := parseSourceImportLock(raw, connector)
	if err != nil {
		t.Fatalf("parse %s source lock: %v", connector, err)
	}
	return lock
}

func retainedSourceMappingCloneLock(lock sourceImportLock) sourceImportLock {
	clone := lock
	clone.Rest.Operations = append([]sourceImportRESTOperation(nil), lock.Rest.Operations...)
	for index := range clone.Rest.Operations {
		if raw := clone.Rest.Operations[index].SourceOperation; raw != nil {
			clone.Rest.Operations[index].SourceOperation = &sourceImportOperationEnrichment{Raw: append([]byte(nil), raw.Raw...)}
		}
	}
	if lock.SourceContract != nil {
		clone.SourceContract = &sourceImportSourceContractEnrichment{Raw: append([]byte(nil), lock.SourceContract.Raw...)}
	}
	return clone
}

func retainedSourceMappingMatrixFixture(t *testing.T, form, sourceID string) []byte {
	t.Helper()
	lanes := []string{"direct_read", "direct_write", "binary_download", "binary_upload", "etl", "reverse_etl", "sync_transport"}
	root := map[string]any{
		"schema_version": 1,
		"connector":      "example",
		"lanes":          lanes,
	}
	if form == "source_operations" {
		rowLanes := make(map[string]any, len(lanes))
		for _, lane := range lanes {
			rowLanes[lane] = map[string]any{"applicability": "not_applicable", "disposition": "not_applicable"}
		}
		rowLanes["etl"] = map[string]any{"applicability": "applicable", "disposition": "mapped_unproven"}
		root["source_operations"] = []any{map[string]any{"source_id": sourceID, "lanes": rowLanes}}
	} else {
		cells := make([]any, 0, len(lanes))
		for _, lane := range lanes {
			state := "not_applicable"
			if lane == "etl" {
				state = "mapped_unproven"
			}
			cells = append(cells, map[string]any{"lane": lane, "state": state})
		}
		root["operations"] = []any{map[string]any{"source_operation_id": sourceID, "cells": cells}}
	}
	raw, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("encode retained source matrix fixture: %v", err)
	}
	return raw
}

func retainedSourceMappingDeepCopy(t *testing.T, value any) any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode fixture copy: %v", err)
	}
	var copied any
	if err := json.Unmarshal(raw, &copied); err != nil {
		t.Fatalf("decode fixture copy: %v", err)
	}
	return copied
}

func TestRetainedSourceMappingOptionErrors(t *testing.T) {
	for _, args := range [][]string{
		{"retained-source-mapping"},
		{"retained-source-mapping", "bitbucket", "--check", "--check"},
		{"retained-source-mapping", "bitbucket", "--out", "x"},
		{"retained-source-mapping", "bitbucket", "stripe"},
	} {
		var stdout, stderr bytes.Buffer
		if exitCode := run(args, &stdout, &stderr); exitCode != 2 {
			t.Fatalf("args=%v exit=%d stderr=%q, want usage error", args, exitCode, stderr.String())
		}
	}
}

func TestRetainedSourceMappingReportHasNoDescriptorPath(t *testing.T) {
	root := retainedSourceMappingTestRoot(t)
	result, err := retainedSourceMappingFromRepository(root, "bitbucket")
	if err != nil {
		t.Fatalf("build retained source mapping: %v", err)
	}
	encoded, err := json.Marshal(result.Report)
	if err != nil {
		t.Fatalf("encode report: %v", err)
	}
	for _, prohibited := range []string{"operation-descriptor", "operations.json", "writes.json", "streams.json", "sync_transport.json"} {
		if strings.Contains(string(encoded), prohibited) {
			t.Fatalf("mapping-only report leaked runtime artifact %q: %s", prohibited, encoded)
		}
	}
	if got := fmt.Sprintf("%d/%d", result.Report.ExecutableDeclarations, len(result.Contract.Lanes)); got != "0/7" {
		t.Fatalf("executable declarations/lanes=%s, want 0/7", got)
	}
}
