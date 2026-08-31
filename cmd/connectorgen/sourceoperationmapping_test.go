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
			"record_shape":   map[string]any{"kind": "collection", "citation": binaryCitation},
			"scope":          map[string]any{"values": []any{"workspace"}, "citation": binaryCitation},
			"path_variables": map[string]any{"values": []any{}, "citation": binaryCitation},
			"media":          map[string]any{"request": []any{}, "response": []any{"application/octet-stream"}, "citation": binaryCitation},
			"event_cursor":   map[string]any{"kind": "none", "citation": binaryCitation},
			"mutation":       map[string]any{"kind": "not_mutation", "citation": binaryCitation},
			"applicability": map[string]any{
				"etl":             map[string]any{"kind": "applicable", "citation": binaryCitation},
				"binary_download": map[string]any{"kind": "applicable", "citation": binaryCitation},
				"binary_upload":   map[string]any{"kind": "not_applicable", "citation": binaryCitation},
				"sync_transport":  map[string]any{"kind": "not_applicable", "citation": binaryCitation},
			},
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

func TestBatch1SourceOperationMappingCohortCheckAcceptsTrackedDenominator(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	manifest := filepath.Join(root, "data", "connector-canon", "batch1-source-operation-mapping-cohort.json")
	report, err := sourceOperationMappingCohortPathCheck(root, manifest)
	if err != nil {
		t.Fatalf("check Batch R1 source-operation cohort: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("Batch R1 cohort findings = %+v, want none", report.Findings)
	}
	if report.ConnectorsChecked != 10 || report.SourceOperations != 4341 {
		t.Fatalf("Batch R1 cohort denominator = connectors:%d source_operations:%d, want 10/4341", report.ConnectorsChecked, report.SourceOperations)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"source-operation-mapping-cohort", manifest, "--check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("cohort CLI exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "10 connector(s), 4341 source operation(s), 0 finding(s)") {
		t.Fatalf("cohort CLI output = %q, want tracked denominator summary", got)
	}
}

// TestBatch1SourceOperationMappingCohortRetentionReceipts starts the #4293
// receipt-cohort slice red. A receipt is source-accounting evidence only: it
// must prove all eligible descriptor-free v2 locks reconcile to their owned
// deterministic sidecars without making an executable declaration.
func TestBatch1SourceOperationMappingCohortRetentionReceipts(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	manifest := filepath.Join(root, "data", "connector-canon", "batch1-source-operation-mapping-cohort.json")
	receipts, err := sourceOperationMappingCohortRetentionReceiptCheck(root, manifest)
	if err != nil {
		t.Fatalf("check Batch R1 retention receipts: %v", err)
	}
	if len(receipts.Findings) != 0 {
		t.Fatalf("Batch R1 retention receipt findings = %+v, want none", receipts.Findings)
	}
	if receipts.ConnectorsChecked != 8 || receipts.SourceOperations != 2340 || receipts.ExecutableDeclarations != 0 {
		t.Fatalf("retention receipts = connectors:%d source_operations:%d executable_declarations:%d, want 8/2340/0", receipts.ConnectorsChecked, receipts.SourceOperations, receipts.ExecutableDeclarations)
	}
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
	got := make(map[string]int, len(receipts.Receipts))
	for _, receipt := range receipts.Receipts {
		if receipt.ExecutableDeclarations != 0 {
			t.Fatalf("%s receipt reports executable declarations=%d, want 0", receipt.Connector, receipt.ExecutableDeclarations)
		}
		got[receipt.Connector] = receipt.SourceOperations
	}
	if !sourceOperationMappingReceiptCountsEqual(got, want) {
		t.Fatalf("retention receipt counts = %#v, want %#v", got, want)
	}

	var stdout, stderr bytes.Buffer
	if code := run([]string{"source-operation-mapping-cohort", manifest, "--check", "--check-retention-receipts"}, &stdout, &stderr); code != 0 {
		t.Fatalf("retention receipt CLI exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if got := stdout.String(); !strings.Contains(got, "retention receipts: 8 connector(s), 2340 source operation(s), 0 executable declaration(s), 0 finding(s)") {
		t.Fatalf("retention receipt CLI output = %q, want exact mapping-only receipt summary", got)
	}
}

func TestBatch1SourceOperationMappingCohortRetentionReceiptsRejectMissingAndDriftedSidecars(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	for _, test := range []struct {
		name   string
		mutate func(t *testing.T, root string)
		want   string
	}{
		{
			name: "missing receipt",
			mutate: func(t *testing.T, root string) {
				t.Helper()
				path := filepath.Join(root, "internal", "connectors", "defs", "stripe", "sources", "stripe-retained-mapping-contract.json")
				if err := os.Remove(path); err != nil {
					t.Fatalf("remove fixture receipt: %v", err)
				}
			},
			want: "read retention sidecar",
		},
		{
			name: "deterministic byte drift",
			mutate: func(t *testing.T, root string) {
				t.Helper()
				path := filepath.Join(root, "internal", "connectors", "defs", "stripe", "sources", "stripe-retained-mapping-contract.json")
				raw, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read fixture receipt: %v", err)
				}
				if err := os.WriteFile(path, append(raw, ' '), 0o644); err != nil {
					t.Fatalf("drift fixture receipt: %v", err)
				}
			},
			want: "bytes drifted",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixtureRoot, manifest := sourceOperationMappingRetentionReceiptFixture(t, root)
			test.mutate(t, fixtureRoot)
			receipts, err := sourceOperationMappingCohortRetentionReceiptCheck(fixtureRoot, manifest)
			if err != nil {
				t.Fatalf("check fixture retention receipts: %v", err)
			}
			if got := sourceOperationMappingFindingsText(receipts.Findings); !strings.Contains(got, "stripe:") || !strings.Contains(got, test.want) {
				t.Fatalf("retention receipt findings = %q, want stripe-scoped %q", got, test.want)
			}
		})
	}
}

func TestSourceOperationMappingCohortRetentionReceiptOptionsAndHelp(t *testing.T) {
	for _, args := range [][]string{
		{"source-operation-mapping-cohort", "manifest.json", "--check-retention-receipts"},
		{"source-operation-mapping-cohort", "manifest.json", "--check", "--check-retention-receipts", "--check-retention-receipts"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(args, &stdout, &stderr); code != 2 {
			t.Fatalf("args=%v exit=%d stderr=%q, want usage error", args, code, stderr.String())
		}
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"source-operation-mapping-cohort", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("cohort help exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	for _, want := range []string{"--check-retention-receipts", "retention_only", "does not materialize"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("cohort help = %q, want %q", stdout.String(), want)
		}
	}
}

func TestBatch1SourceOperationMappingCohortCheckRejectsDigestCountAndMembershipDefects(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "source lock digest mismatch",
			mutate: func(document map[string]any) {
				document["source_locks"].([]any)[0].(map[string]any)["sha256"] = strings.Repeat("0", 64)
			},
			want: "source lock SHA-256",
		},
		{
			name: "source ID digest mismatch",
			mutate: func(document map[string]any) {
				document["source_locks"].([]any)[0].(map[string]any)["source_ids_sha256"] = strings.Repeat("0", 64)
			},
			want: "source ID SHA-256",
		},
		{
			name: "aggregate source ID digest mismatch",
			mutate: func(document map[string]any) {
				document["source_operations_sha256"] = strings.Repeat("0", 64)
			},
			want: "does not match tracked cohort",
		},
		{
			name: "source operation count mismatch",
			mutate: func(document map[string]any) {
				document["source_locks"].([]any)[0].(map[string]any)["source_operation_count"] = float64(1)
			},
			want: "source operation count",
		},
		{
			name: "missing fixed connector",
			mutate: func(document map[string]any) {
				document["source_locks"].([]any)[0].(map[string]any)["connector"] = "not-batch1"
			},
			want: "exact Batch R1 connector membership",
		},
		{
			name: "expected connector list mismatch",
			mutate: func(document map[string]any) {
				document["expected_connectors"].([]any)[0] = "not-batch1"
			},
			want: "expected_connectors does not retain exact Batch R1 connector membership",
		},
		{
			name: "source lock path traversal",
			mutate: func(document map[string]any) {
				document["source_locks"].([]any)[0].(map[string]any)["path"] = "../outside-operation-source-lock.json"
			},
			want: "source lock path",
		},
		{
			name: "matrix input path traversal",
			mutate: func(document map[string]any) {
				document["source_locks"].([]any)[0].(map[string]any)["matrix_path"] = "../outside-source-lane-matrix.json"
			},
			want: "connector-local matrix input",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := sourceOperationMappingCohortFixture(t, root)
			document := sourceOperationMappingReadJSON(t, manifest)
			test.mutate(document)
			sourceOperationMappingWriteJSON(t, manifest, document)

			report, err := sourceOperationMappingCohortPathCheck(root, manifest)
			if err != nil {
				t.Fatalf("check cohort: %v", err)
			}
			if len(report.Findings) == 0 {
				t.Fatalf("cohort defect passed")
			}
			if got := sourceOperationMappingFindingsText(report.Findings); !strings.Contains(got, test.want) {
				t.Fatalf("cohort findings = %q, want %q", got, test.want)
			}
		})
	}
}

func TestSourceOperationMappingCheckRequiresSourceBackedMutationWriteCells(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "REST mutation missing direct write",
			mutate: func(document map[string]any) {
				sourceOperationMappingRemoveCell(document["operations"].([]any)[1].(map[string]any), "direct_write")
			},
			want: "mutation source operation \"fixture.rest.post.custom\" has no explicit direct_write disposition",
		},
		{
			name: "REST mutation missing reverse ETL",
			mutate: func(document map[string]any) {
				sourceOperationMappingRemoveCell(document["operations"].([]any)[1].(map[string]any), "reverse_etl")
			},
			want: "mutation source operation \"fixture.rest.post.custom\" has no explicit reverse_etl disposition",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := sourceOperationMappingFixture(t)
			document := sourceOperationMappingReadJSON(t, manifest)
			test.mutate(document)
			sourceOperationMappingWriteJSON(t, manifest, document)
			sourceOperationMappingAssertRejected(t, manifest, test.want)
		})
	}
}

func TestSourceOperationMappingCheckRequiresWriteCellsForLockedDelete(t *testing.T) {
	manifest := sourceOperationMappingFixture(t)
	root := filepath.Dir(manifest)
	lockPath := filepath.Join(root, "fixture", "sources", "fixture-operation-source-lock.json")
	lock := sourceOperationMappingReadJSON(t, lockPath)
	rest := lock["rest"].(map[string]any)
	rest["operation_counts"] = map[string]any{"GET": 1, "DELETE": 1}
	rest["operations"].([]any)[1].(map[string]any)["method"] = "DELETE"
	sourceOperationMappingWriteJSON(t, lockPath, lock)

	document := sourceOperationMappingReadJSON(t, manifest)
	sourceOperationMappingRemoveCell(document["operations"].([]any)[1].(map[string]any), "direct_write")
	sourceOperationMappingWriteJSON(t, manifest, document)
	sourceOperationMappingAssertRejected(t, manifest, "mutation source operation \"fixture.rest.post.custom\" has no explicit direct_write disposition")
}

func TestSourceOperationMappingCheckRetainsSourceCitedNonMutatingPOSTBoundary(t *testing.T) {
	manifest := sourceOperationMappingFixture(t)
	document := sourceOperationMappingReadJSON(t, manifest)
	operation := document["operations"].([]any)[1].(map[string]any)
	operation["facts"].(map[string]any)["mutation"].(map[string]any)["kind"] = "not_mutation"
	sourceOperationMappingRemoveCell(operation, "direct_write")
	sourceOperationMappingRemoveCell(operation, "reverse_etl")
	artifact := document["artifacts"].([]any)[0].(map[string]any)
	links := artifact["cells"].([]any)
	artifact["cells"] = []any{links[0]}
	sourceOperationMappingWriteJSON(t, manifest, document)

	var stdout, stderr bytes.Buffer
	if code := run([]string{"source-operation-mapping", manifest, "--check"}, &stdout, &stderr); code != 0 {
		t.Fatalf("source-cited non-mutating POST exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestSourceOperationMappingCheckRequiresGraphQLMutationWriteCells(t *testing.T) {
	for _, lane := range []string{"direct_write", "reverse_etl"} {
		t.Run(lane, func(t *testing.T) {
			manifest := sourceOperationMappingGraphQLMutationFixture(t)
			document := sourceOperationMappingReadJSON(t, manifest)
			operation := document["operations"].([]any)[0].(map[string]any)
			sourceOperationMappingRemoveCell(operation, lane)
			sourceOperationMappingWriteJSON(t, manifest, document)
			sourceOperationMappingAssertRejected(t, manifest, "mutation source operation \"fixture.graphql.mutation.setWidget\" has no explicit "+lane+" disposition")
		})
	}
}

func TestSourceOperationMappingCheckAcceptsSourceCitedConcreteBinaryMedia(t *testing.T) {
	tests := []struct {
		name  string
		media string
	}{
		{
			name:  "gzip upload",
			media: "application/gzip",
		},
		{
			name:  "parameterized gzip upload",
			media: "application/gzip; charset=binary",
		},
		{
			name:  "provider archive upload",
			media: "application/x-provider-archive",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := sourceOperationMappingFixture(t)
			document := sourceOperationMappingReadJSON(t, manifest)
			operation := document["operations"].([]any)[1].(map[string]any)
			facts := operation["facts"].(map[string]any)
			facts["media"].(map[string]any)["request"] = []any{test.media}
			facts["applicability"].(map[string]any)["binary_upload"].(map[string]any)["kind"] = "applicable"
			sourceOperationMappingSetCellState(t, operation, "binary_upload", "mapped_unproven")
			sourceOperationMappingWriteJSON(t, manifest, document)

			var stdout, stderr bytes.Buffer
			if code := run([]string{"source-operation-mapping", manifest, "--check"}, &stdout, &stderr); code != 0 {
				t.Fatalf("source-cited media %q exit=%d stdout=%q stderr=%q", test.media, code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestSourceOperationMappingCheckRejectsUncitedOrContradictoryBinaryMedia(t *testing.T) {
	tests := []struct {
		name   string
		media  string
		mutate func(map[string]any)
		want   string
	}{
		{
			name:  "media fact cites another source node",
			media: "application/gzip",
			mutate: func(facts map[string]any) {
				facts["media"].(map[string]any)["citation"].(map[string]any)["location"] = `paths["/widgets"].get`
			},
			want: "media fact citation: location",
		},
		{
			name:  "media fact lacks a source citation",
			media: "application/gzip",
			mutate: func(facts map[string]any) {
				delete(facts["media"].(map[string]any), "citation")
			},
			want: "validate manifest shape",
		},
		{
			name:  "JSON request is not binary evidence",
			media: "application/json",
			want:  "binary_upload applicability is not supported by request media evidence",
		},
		{
			name:  "malformed media type is not binary evidence",
			media: "not-a-media-type",
			want:  "binary_upload applicability is not supported by request media evidence",
		},
		{
			name:  "wildcard media type is not binary evidence",
			media: "application/*",
			want:  "binary_upload applicability is not supported by request media evidence",
		},
		{
			name:  "wildcard suffix media type is not binary evidence",
			media: "application/*+xml",
			want:  "binary_upload applicability is not supported by request media evidence",
		},
		{
			name:  "wildcard top-level media type is not binary evidence",
			media: "*/pdf",
			want:  "binary_upload applicability is not supported by request media evidence",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := sourceOperationMappingFixture(t)
			document := sourceOperationMappingReadJSON(t, manifest)
			operation := document["operations"].([]any)[1].(map[string]any)
			facts := operation["facts"].(map[string]any)
			facts["media"].(map[string]any)["request"] = []any{test.media}
			facts["applicability"].(map[string]any)["binary_upload"].(map[string]any)["kind"] = "applicable"
			sourceOperationMappingSetCellState(t, operation, "binary_upload", "mapped_unproven")
			if test.mutate != nil {
				test.mutate(facts)
			}
			sourceOperationMappingWriteJSON(t, manifest, document)
			sourceOperationMappingAssertRejected(t, manifest, test.want)
		})
	}
}

func TestSourceOperationMappingCheckRejectsWildcardBinaryDownloadMedia(t *testing.T) {
	for _, media := range []string{"application/*+xml", "*/pdf"} {
		t.Run(media, func(t *testing.T) {
			manifest := sourceOperationMappingFixture(t)
			document := sourceOperationMappingReadJSON(t, manifest)
			operation := document["operations"].([]any)[0].(map[string]any)
			facts := operation["facts"].(map[string]any)
			facts["media"].(map[string]any)["response"] = []any{media}
			facts["applicability"].(map[string]any)["binary_download"].(map[string]any)["kind"] = "applicable"
			operation["cells"] = append(operation["cells"].([]any), map[string]any{
				"lane":  "binary_download",
				"state": "mapped_unproven",
			})
			sourceOperationMappingWriteJSON(t, manifest, document)

			sourceOperationMappingAssertRejected(t, manifest, "binary_download applicability is not supported by response media evidence")
		})
	}
}

func TestSourceOperationMappingCheckRejectsGraphQLMutationFactMismatch(t *testing.T) {
	manifest := sourceOperationMappingGraphQLMutationFixture(t)
	document := sourceOperationMappingReadJSON(t, manifest)
	document["operations"].([]any)[0].(map[string]any)["facts"].(map[string]any)["mutation"].(map[string]any)["kind"] = "not_mutation"
	sourceOperationMappingWriteJSON(t, manifest, document)
	sourceOperationMappingAssertRejected(t, manifest, "mutation fact contradicts locked GraphQL root Mutation.setWidget identity")
}

func TestSourceOperationMappingCheckRejectsUncitedAndContradictoryApplicabilityFacts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{
			name: "missing record shape",
			mutate: func(document map[string]any) {
				delete(document["operations"].([]any)[0].(map[string]any)["facts"].(map[string]any), "record_shape")
			},
			want: "record_shape",
		},
		{
			name: "fact cites another source node",
			mutate: func(document map[string]any) {
				document["operations"].([]any)[0].(map[string]any)["facts"].(map[string]any)["record_shape"].(map[string]any)["citation"].(map[string]any)["location"] = "custom reference > endpoint"
			},
			want: "record_shape fact citation: location",
		},
		{
			name: "ETL applicability cites another source node",
			mutate: func(document map[string]any) {
				document["operations"].([]any)[0].(map[string]any)["facts"].(map[string]any)["applicability"].(map[string]any)["etl"].(map[string]any)["citation"].(map[string]any)["location"] = "custom reference > endpoint"
			},
			want: "etl applicability fact citation: location",
		},
		{
			name: "ETL applicability contradicts pagination and record shape",
			mutate: func(document map[string]any) {
				document["operations"].([]any)[0].(map[string]any)["facts"].(map[string]any)["applicability"].(map[string]any)["etl"].(map[string]any)["kind"] = "not_applicable"
			},
			want: "etl applicability contradicts collection pagination evidence",
		},
		{
			name: "binary download applicability contradicts response media",
			mutate: func(document map[string]any) {
				facts := document["operations"].([]any)[0].(map[string]any)["facts"].(map[string]any)
				facts["media"].(map[string]any)["response"] = []any{"application/octet-stream"}
			},
			want: "binary_download applicability contradicts response media evidence",
		},
		{
			name: "binary upload applicability contradicts request media",
			mutate: func(document map[string]any) {
				facts := document["operations"].([]any)[1].(map[string]any)["facts"].(map[string]any)
				facts["media"].(map[string]any)["request"] = []any{"application/octet-stream"}
			},
			want: "binary_upload applicability contradicts request media evidence",
		},
		{
			name: "sync applicability contradicts event evidence",
			mutate: func(document map[string]any) {
				document["operations"].([]any)[0].(map[string]any)["facts"].(map[string]any)["applicability"].(map[string]any)["sync_transport"].(map[string]any)["kind"] = "not_applicable"
			},
			want: "sync_transport applicability contradicts event/cursor evidence",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifest := sourceOperationMappingFixture(t)
			document := sourceOperationMappingReadJSON(t, manifest)
			test.mutate(document)
			sourceOperationMappingWriteJSON(t, manifest, document)
			sourceOperationMappingAssertRejected(t, manifest, test.want)
		})
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
					"record_shape":   map[string]any{"kind": "collection", "citation": primaryCitation},
					"scope":          map[string]any{"values": []any{"workspace"}, "citation": primaryCitation},
					"path_variables": map[string]any{"values": []any{}, "citation": primaryCitation},
					"media":          map[string]any{"request": []any{}, "response": []any{"application/json"}, "citation": primaryCitation},
					"event_cursor":   map[string]any{"kind": "cursor", "citation": primaryCitation},
					"mutation":       map[string]any{"kind": "not_mutation", "citation": primaryCitation},
					"applicability": map[string]any{
						"etl":             map[string]any{"kind": "applicable", "citation": primaryCitation},
						"binary_download": map[string]any{"kind": "not_applicable", "citation": primaryCitation},
						"binary_upload":   map[string]any{"kind": "not_applicable", "citation": primaryCitation},
						"sync_transport":  map[string]any{"kind": "applicable", "citation": primaryCitation},
					},
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
					"record_shape":   map[string]any{"kind": "record", "citation": customCitation},
					"scope":          map[string]any{"values": []any{"workspace"}, "citation": customCitation},
					"path_variables": map[string]any{"values": []any{}, "citation": customCitation},
					"media":          map[string]any{"request": []any{"application/json"}, "response": []any{"application/json"}, "citation": customCitation},
					"event_cursor":   map[string]any{"kind": "none", "citation": customCitation},
					"mutation":       map[string]any{"kind": "mutation", "citation": customCitation},
					"applicability": map[string]any{
						"etl":             map[string]any{"kind": "not_applicable", "citation": customCitation},
						"binary_download": map[string]any{"kind": "not_applicable", "citation": customCitation},
						"binary_upload":   map[string]any{"kind": "not_applicable", "citation": customCitation},
						"sync_transport":  map[string]any{"kind": "not_applicable", "citation": customCitation},
					},
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

func sourceOperationMappingCohortFixture(t *testing.T, root string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, "data", "connector-canon", "batch1-source-operation-mapping-cohort.json"))
	if err != nil {
		t.Fatalf("read Batch R1 cohort: %v", err)
	}
	path := filepath.Join(t.TempDir(), "batch1-source-operation-mapping-cohort.json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write cohort fixture: %v", err)
	}
	return path
}

func sourceOperationMappingRetentionReceiptFixture(t *testing.T, root string) (string, string) {
	t.Helper()
	cohortPath := filepath.Join(root, "data", "connector-canon", "batch1-source-operation-mapping-cohort.json")
	raw, err := os.ReadFile(cohortPath)
	if err != nil {
		t.Fatalf("read Batch R1 cohort: %v", err)
	}
	var cohort sourceOperationMappingCohortManifest
	if err := decodeSourceStrictJSON(raw, &cohort); err != nil {
		t.Fatalf("decode Batch R1 cohort: %v", err)
	}
	fixtureRoot := t.TempDir()
	manifest := filepath.Join(fixtureRoot, "data", "connector-canon", "batch1-source-operation-mapping-cohort.json")
	sourceOperationMappingCopyFixtureFile(t, root, fixtureRoot, filepath.ToSlash(filepath.Join("data", "connector-canon", "batch1-source-operation-mapping-cohort.json")))
	for _, sourceLock := range cohort.SourceLocks {
		sourceOperationMappingCopyFixtureFile(t, root, fixtureRoot, sourceLock.Path)
		sourceOperationMappingCopyFixtureFile(t, root, fixtureRoot, sourceLock.MatrixPath)
		sidecar := filepath.ToSlash(filepath.Join("internal", "connectors", "defs", sourceLock.Connector, "sources", sourceLock.Connector+retainedSourceMappingSidecarSuffix))
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(sidecar))); err == nil {
			sourceOperationMappingCopyFixtureFile(t, root, fixtureRoot, sidecar)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", sidecar, err)
		}
	}
	return fixtureRoot, manifest
}

func sourceOperationMappingCopyFixtureFile(t *testing.T, sourceRoot, destinationRoot, relative string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(sourceRoot, filepath.FromSlash(relative)))
	if err != nil {
		t.Fatalf("read fixture source %s: %v", relative, err)
	}
	destination := filepath.Join(destinationRoot, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		t.Fatalf("create fixture directory for %s: %v", relative, err)
	}
	if err := os.WriteFile(destination, raw, 0o644); err != nil {
		t.Fatalf("write fixture source %s: %v", relative, err)
	}
}

func sourceOperationMappingReceiptCountsEqual(left, right map[string]int) bool {
	if len(left) != len(right) {
		return false
	}
	for connector, expected := range right {
		if left[connector] != expected {
			return false
		}
	}
	return true
}

func sourceOperationMappingFindingsText(findings []Finding) string {
	parts := make([]string, 0, len(findings))
	for _, finding := range findings {
		parts = append(parts, finding.Connector+": "+finding.Message)
	}
	return strings.Join(parts, "\n")
}

func sourceOperationMappingRemoveCell(operation map[string]any, lane string) {
	cells := operation["cells"].([]any)
	filtered := make([]any, 0, len(cells))
	for _, rawCell := range cells {
		if rawCell.(map[string]any)["lane"] == lane {
			continue
		}
		filtered = append(filtered, rawCell)
	}
	operation["cells"] = filtered
}

func sourceOperationMappingSetCellState(t *testing.T, operation map[string]any, lane, state string) {
	t.Helper()
	for _, rawCell := range operation["cells"].([]any) {
		cell := rawCell.(map[string]any)
		if cell["lane"] != lane {
			continue
		}
		cell["state"] = state
		delete(cell, "reason")
		return
	}
	t.Fatalf("source operation has no %s cell", lane)
}

func sourceOperationMappingAssertRejected(t *testing.T, manifest, want string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"source-operation-mapping", manifest, "--check"}, &stdout, &stderr); code == 0 {
		t.Fatalf("mapping defect passed: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
	if got := stdout.String() + stderr.String(); !strings.Contains(got, want) {
		t.Fatalf("mapping diagnostic = %q, want %q", got, want)
	}
}

func sourceOperationMappingGraphQLMutationFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	lockPath := filepath.Join(root, "fixture", "sources", "fixture-operation-source-lock.json")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatalf("create GraphQL fixture source directory: %v", err)
	}
	lock := map[string]any{
		"schema_version": 2,
		"connector":      "fixture",
		"rest": map[string]any{
			"source_url": "https://docs.polymetrics.invalid/fixture/openapi",
			"sha256":     "retention-is-ignored-by-mapping-admission",
			"bytes":      -1,
			"openapi":    "3.0.0",
			"operations": []any{},
		},
		"graphql": map[string]any{
			"source_url":      "https://docs.polymetrics.invalid/fixture/graphql",
			"sha256":          "retention-is-ignored-by-mapping-admission",
			"bytes":           -1,
			"query_fields":    []any{},
			"mutation_fields": []any{map[string]any{"root": "Mutation", "name": "setWidget", "line": 1, "signature": "setWidget(id: ID!): Widget"}},
			"type_system":     map[string]any{},
		},
		"counts": map[string]any{"rest": 0, "graphql_query": 0, "graphql_mutation": 1, "total": 1},
	}
	sourceOperationMappingWriteJSON(t, lockPath, lock)

	citation := sourceOperationMappingCitation("https://docs.polymetrics.invalid/fixture/graphql", `graphql.mutation_fields["setWidget"]@line:1`)
	document := map[string]any{
		"schema_version": 1,
		"source_locks": []any{map[string]any{
			"connector": "fixture",
			"path":      "fixture/sources/fixture-operation-source-lock.json",
		}},
		"operations": []any{map[string]any{
			"connector":              "fixture",
			"source_operation_id":    "fixture.graphql.mutation.setWidget",
			"canonical_operation_id": "fixture.graphql.mutation.setWidget",
			"citation":               citation,
			"facts": map[string]any{
				"pagination":     map[string]any{"kind": "none", "citation": citation},
				"record_shape":   map[string]any{"kind": "record", "citation": citation},
				"scope":          map[string]any{"values": []any{}, "citation": citation},
				"path_variables": map[string]any{"values": []any{}, "citation": citation},
				"media":          map[string]any{"request": []any{"application/json"}, "response": []any{"application/json"}, "citation": citation},
				"event_cursor":   map[string]any{"kind": "none", "citation": citation},
				"mutation":       map[string]any{"kind": "mutation", "citation": citation},
				"applicability": map[string]any{
					"etl":             map[string]any{"kind": "not_applicable", "citation": citation},
					"binary_download": map[string]any{"kind": "not_applicable", "citation": citation},
					"binary_upload":   map[string]any{"kind": "not_applicable", "citation": citation},
					"sync_transport":  map[string]any{"kind": "not_applicable", "citation": citation},
				},
			},
			"cells": []any{
				map[string]any{"lane": "direct_write", "state": "mapped_unproven"},
				map[string]any{"lane": "reverse_etl", "state": "mapped_unproven"},
			},
		}},
		"artifacts": []any{},
	}
	manifest := filepath.Join(root, "source-operation-mapping.json")
	sourceOperationMappingWriteJSON(t, manifest, document)
	return manifest
}
