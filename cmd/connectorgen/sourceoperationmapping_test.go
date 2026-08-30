package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceOperationMappingHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"source-operation-mapping", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("source-operation-mapping help exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "source-operation-mapping <manifest> --check") {
		t.Fatalf("source-operation-mapping help = %q, want check-only usage", got)
	}
}

func TestSourceOperationMappingCheckAcceptsCitedMultiLaneManifest(t *testing.T) {
	manifest := sourceOperationMappingFixture(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"source-operation-mapping", manifest, "--check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("source-operation-mapping exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "2 source operation(s), 2 canonical operation(s), 7 cell(s), 0 finding(s)") {
		t.Fatalf("source-operation-mapping output = %q, want source-first clean summary", got)
	}
}

func TestSourceOperationMappingCheckRejectsMembershipAndCellDefects(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "duplicate source operation ID",
			mutate: func(document map[string]any) {
				operations := document["operations"].([]any)
				document["operations"] = append(operations, operations[0])
			},
			want: "duplicate source operation ID",
		},
		{
			name: "locked source row absent from manifest",
			mutate: func(document map[string]any) {
				document["operations"] = document["operations"].([]any)[:1]
			},
			want: "is absent from the mapping manifest",
		},
		{
			name: "pageable source has no ETL disposition",
			mutate: func(document map[string]any) {
				operation := document["operations"].([]any)[0].(map[string]any)
				operation["cells"] = operation["cells"].([]any)[:1]
			},
			want: "pageable source operation \"fixture.rest.get.widgets\" has no explicit etl disposition",
		},
		{
			name: "artifact references nonexistent cell",
			mutate: func(document map[string]any) {
				artifact := document["artifacts"].([]any)[0].(map[string]any)
				artifact["cells"].([]any)[0].(map[string]any)["lane"] = "binary_upload"
			},
			want: "artifact \"fixture/operations.json\" references nonexistent mapping cell fixture.rest.get.widgets/binary_upload",
		},
		{
			name: "artifact path traversal",
			mutate: func(document map[string]any) {
				document["artifacts"].([]any)[0].(map[string]any)["path"] = "../fixture/operations.json"
			},
			want: "artifact path \"../fixture/operations.json\" must be one canonical relative path",
		},
		{
			name: "missing foundation requires stable typed reason",
			mutate: func(document map[string]any) {
				cell := document["operations"].([]any)[0].(map[string]any)["cells"].([]any)[1].(map[string]any)
				delete(cell, "reason")
			},
			want: "missing_foundation requires a stable typed reason",
		},
		{
			name: "not applicable requires source evidence",
			mutate: func(document map[string]any) {
				cell := document["operations"].([]any)[1].(map[string]any)["cells"].([]any)[2].(map[string]any)
				delete(cell, "reason")
			},
			want: "not_applicable cell requires source evidence",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := sourceOperationMappingFixture(t)
			document := sourceOperationMappingReadJSON(t, manifest)
			test.mutate(document)
			sourceOperationMappingWriteJSON(t, manifest, document)

			var stdout, stderr bytes.Buffer
			if code := run([]string{"source-operation-mapping", manifest, "--check"}, &stdout, &stderr); code == 0 {
				t.Fatalf("defect passed: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			if got := stdout.String() + stderr.String(); !strings.Contains(got, test.want) {
				t.Fatalf("diagnostic = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSourceOperationMappingCheckPreservesSupplementalSourceLineageWithoutInflatingCanonicalDenominator(t *testing.T) {
	manifest := sourceOperationMappingFixture(t)
	root := filepath.Dir(manifest)
	binaryLockPath := filepath.Join(root, "fixture", "sources", "fixture-binary-operation-source-lock.json")
	sourceOperationMappingWriteJSON(t, binaryLockPath, declarationAdmissionR2OpenAPILock(t))

	document := sourceOperationMappingReadJSON(t, manifest)
	document["source_locks"] = append(document["source_locks"].([]any), map[string]any{
		"connector": "fixture",
		"path":      "fixture/sources/fixture-binary-operation-source-lock.json",
	})
	binaryCitation := sourceOperationMappingCitation("https://docs.polymetrics.invalid/reference/openapi", `paths["/widgets"].get`)
	document["operations"] = append(document["operations"].([]any), map[string]any{
		"connector":              "fixture",
		"source_operation_id":    "fixture.rest.primary.list-widgets",
		"canonical_operation_id": "fixture.rest.get.widgets",
		"citation":               binaryCitation,
		"facts": map[string]any{
			"pagination":     map[string]any{"kind": "cursor", "citation": binaryCitation},
			"scope":          map[string]any{"values": []any{"workspace"}, "citation": binaryCitation},
			"path_variables": map[string]any{"values": []any{}, "citation": binaryCitation},
			"media":          map[string]any{"request": []any{}, "response": []any{"application/octet-stream"}, "citation": binaryCitation},
			"event_cursor":   map[string]any{"kind": "cursor", "citation": binaryCitation},
		},
		"cells": []any{
			map[string]any{"lane": "binary_download", "state": "mapped_unproven"},
			map[string]any{"lane": "etl", "state": "mapped_unproven"},
		},
	})
	sourceOperationMappingWriteJSON(t, manifest, document)

	report, err := sourceOperationMappingPathCheck(manifest)
	if err != nil {
		t.Fatalf("source-operation-mapping check: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("source-operation-mapping findings = %+v, want none", report.Findings)
	}
	if report.SourceOperations != 3 || report.CanonicalOperations != 2 {
		t.Fatalf("source-operation-mapping denominators = source:%d canonical:%d, want source:3 canonical:2", report.SourceOperations, report.CanonicalOperations)
	}
}

func TestSourceOperationMappingCheckRejectsIncompatibleCanonicalRelation(t *testing.T) {
	manifest := sourceOperationMappingFixture(t)
	document := sourceOperationMappingReadJSON(t, manifest)
	document["operations"].([]any)[1].(map[string]any)["canonical_operation_id"] = "fixture.rest.get.widgets"
	sourceOperationMappingWriteJSON(t, manifest, document)

	var stdout, stderr bytes.Buffer
	if code := run([]string{"source-operation-mapping", manifest, "--check"}, &stdout, &stderr); code == 0 {
		t.Fatalf("incompatible canonical relation passed: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	if got := stdout.String() + stderr.String(); !strings.Contains(got, "does not preserve source-lock operation identity") {
		t.Fatalf("diagnostic = %q, want canonical identity refusal", got)
	}
}

func TestSourceOperationMappingCanonicalIdentityRetainsGraphQLRootField(t *testing.T) {
	message := sourceOperationMappingCanonicalIdentityMismatch(
		declarationAdmissionReviewedOperation{Protocol: "graphql", Method: "GRAPHQL", Path: "widget", ProviderOperationID: "Query.widget"},
		declarationAdmissionReviewedOperation{Protocol: "graphql", Method: "GRAPHQL", Path: "widget", ProviderOperationID: "Mutation.widget"},
	)
	if !strings.Contains(message, "GraphQL provider operation identity") {
		t.Fatalf("GraphQL canonical identity mismatch = %q, want root-field protection", message)
	}
}

func sourceOperationMappingFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	lockPath := filepath.Join(root, "fixture", "sources", "fixture-operation-source-lock.json")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatalf("create source lock directory: %v", err)
	}
	sourceOperationMappingWriteJSON(t, lockPath, declarationAdmissionR2LegacyReferenceLock(t))

	primaryCitation := sourceOperationMappingCitation("https://docs.polymetrics.invalid/fixture/openapi", `paths["/widgets"].get`)
	customCitation := sourceOperationMappingCitation("https://docs.polymetrics.invalid/fixture/custom", "custom reference > endpoint")
	document := map[string]any{
		"schema_version": 1,
		"source_locks": []any{map[string]any{
			"connector": "fixture",
			"path":      "fixture/sources/fixture-operation-source-lock.json",
		}},
		"operations": []any{
			map[string]any{
				"connector":              "fixture",
				"source_operation_id":    "fixture.rest.get.widgets",
				"canonical_operation_id": "fixture.rest.get.widgets",
				"citation":               primaryCitation,
				"facts": map[string]any{
					"pagination":     map[string]any{"kind": "cursor", "citation": primaryCitation},
					"scope":          map[string]any{"values": []any{"workspace"}, "citation": primaryCitation},
					"path_variables": map[string]any{"values": []any{}, "citation": primaryCitation},
					"media":          map[string]any{"request": []any{}, "response": []any{"application/json"}, "citation": primaryCitation},
					"event_cursor":   map[string]any{"kind": "cursor", "citation": primaryCitation},
				},
				"cells": []any{
					map[string]any{"lane": "direct_read", "state": "mapped_unproven"},
					map[string]any{"lane": "etl", "state": "missing_foundation", "reason": sourceOperationMappingReason("missing_foundation", "runtime.fixture-etl.v1", primaryCitation)},
					map[string]any{"lane": "sync_transport", "state": "mapped_unproven"},
				},
			},
			map[string]any{
				"connector":              "fixture",
				"source_operation_id":    "fixture.rest.post.custom",
				"canonical_operation_id": "fixture.rest.post.custom",
				"citation":               customCitation,
				"facts": map[string]any{
					"pagination":     map[string]any{"kind": "none", "citation": customCitation},
					"scope":          map[string]any{"values": []any{"workspace"}, "citation": customCitation},
					"path_variables": map[string]any{"values": []any{}, "citation": customCitation},
					"media":          map[string]any{"request": []any{"application/json"}, "response": []any{"application/json"}, "citation": customCitation},
					"event_cursor":   map[string]any{"kind": "none", "citation": customCitation},
				},
				"cells": []any{
					map[string]any{"lane": "direct_write", "state": "implemented"},
					map[string]any{"lane": "reverse_etl", "state": "mapped_unproven"},
					map[string]any{"lane": "binary_download", "state": "not_applicable", "reason": sourceOperationMappingReason("provider_not_applicable", "provider.no_binary_response", customCitation)},
					map[string]any{"lane": "binary_upload", "state": "not_applicable", "reason": sourceOperationMappingReason("provider_not_applicable", "provider.no_binary_request", customCitation)},
				},
			},
		},
		"artifacts": []any{map[string]any{
			"path": "fixture/operations.json",
			"cells": []any{
				map[string]any{"source_operation_id": "fixture.rest.get.widgets", "lane": "direct_read"},
				map[string]any{"source_operation_id": "fixture.rest.post.custom", "lane": "direct_write"},
			},
		}},
	}
	manifest := filepath.Join(root, "source-operation-mapping.json")
	sourceOperationMappingWriteJSON(t, manifest, document)
	return manifest
}

func sourceOperationMappingCitation(sourceURL, location string) map[string]any {
	return map[string]any{"source_url": sourceURL, "location": location}
}

func sourceOperationMappingReason(kind, id string, citation map[string]any) map[string]any {
	return map[string]any{"kind": kind, "id": id, "citation": citation}
}

func sourceOperationMappingReadJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return document
}

func sourceOperationMappingWriteJSON(t *testing.T, path string, document any) {
	t.Helper()
	raw, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
