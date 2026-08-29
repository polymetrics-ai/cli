package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

const githubIssuesSourceID = "github.rest.issues/list-for-repo"

func TestOperationEvidenceClassForCommandUsesIntentNotOperationName(t *testing.T) {
	if got := operationEvidenceClassForCommand("binary_upload", "releases_upload_asset"); got != operationEvidenceClassBinaryUpload {
		t.Fatalf("binary_upload classification = %q, want %q", got, operationEvidenceClassBinaryUpload)
	}
	if got := operationEvidenceClassForCommand("direct_write", "releases_upload_asset"); got != operationEvidenceClassDirectWrite {
		t.Fatalf("direct_write classification = %q, want %q; an operation name must not promote an ordinary write", got, operationEvidenceClassDirectWrite)
	}
}

func TestOperationEvidencePreservesSavedAndInteractiveLanes(t *testing.T) {
	root := operationEvidenceWorkspace(t)
	artifact, _, stderr := runOperationEvidenceForTest(t, root, "")
	if stderr != "" {
		t.Fatalf("operation-evidence wrote diagnostics on complete invocation: %s", stderr)
	}

	tests := []struct {
		name          string
		sourceID      string
		enabledLanes  []string
		disabledLanes []string
	}{
		{
			name:         "stream-backed direct read",
			sourceID:     "asana.rest.getCustomFieldsForWorkspace",
			enabledLanes: []string{operationEvidenceClassETL, operationEvidenceClassDirectRead},
		},
		{
			name:         "action-backed direct write",
			sourceID:     "asana.rest.addCustomFieldSettingForGoal",
			enabledLanes: []string{operationEvidenceClassReverseETL, operationEvidenceClassDirectWrite},
		},
		{
			name:          "operation-backed direct read",
			sourceID:      "asana.rest.getAccessRequests",
			enabledLanes:  []string{operationEvidenceClassDirectRead},
			disabledLanes: []string{operationEvidenceClassETL, operationEvidenceClassReverseETL},
		},
		{
			name:          "unsupported batch operation",
			sourceID:      "asana.rest.createBatchRequest",
			disabledLanes: operationEvidenceClasses,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row, found := artifact.row(test.sourceID)
			if !found {
				t.Fatalf("operation evidence omitted source row %q", test.sourceID)
			}
			for _, lane := range test.enabledLanes {
				classification := row.Classifications[lane]
				if !classification.Declared || !classification.Enabled {
					t.Fatalf("%s classification = %+v, want declared and enabled; runtime targets = %v, CLI paths = %v", lane, classification, row.Runtime.Targets, row.CLI.Paths)
				}
			}
			for _, lane := range test.disabledLanes {
				if classification := row.Classifications[lane]; classification.Enabled {
					t.Fatalf("%s classification = %+v, want disabled; runtime targets = %v, CLI paths = %v", lane, classification, row.Runtime.Targets, row.CLI.Paths)
				}
			}
		})
	}
}

const githubReadOnlySourceID = "github.rest.actions/list-artifacts-for-repo"

func (artifact operationEvidenceArtifact) rollupContaining(rollups []operationEvidenceRollup, sourceID string) (operationEvidenceRollup, bool) {
	for _, rollup := range rollups {
		if slices.Contains(rollup.SourceIDs, sourceID) {
			return rollup, true
		}
	}
	return operationEvidenceRollup{}, false
}

func (artifact operationEvidenceArtifact) readOnlyRollup(connector, policy string) (operationEvidenceReadOnlyRollup, bool) {
	for _, rollup := range artifact.IntentionallyReadOnly {
		if rollup.Connector == connector && rollup.Policy == policy {
			return rollup, true
		}
	}
	return operationEvidenceReadOnlyRollup{}, false
}

func TestOperationEvidenceProjectsGitHubAcrossEveryEvidenceSurface(t *testing.T) {
	root := operationEvidenceWorkspace(t)
	artifact, _, stderr := runOperationEvidenceForTest(t, root, "")
	if stderr != "" {
		t.Fatalf("operation-evidence wrote diagnostics on complete invocation: %s", stderr)
	}

	row, found := artifact.row(githubIssuesSourceID)
	if !found {
		t.Fatalf("operation evidence omitted source row %q", githubIssuesSourceID)
	}
	if row.Source.URL == "" || row.Source.SHA256 == "" || row.Source.Location == "" {
		t.Fatalf("source trace = %+v, want provider URL, digest, and source location", row.Source)
	}
	if artifact.Provenance.SourceProjectionSHA256 == "" || artifact.Provenance.RelevantConfigSHA256 == "" {
		t.Fatalf("provenance = %+v, want source projection and relevant configuration digests", artifact.Provenance)
	}
	if row.Canonical.Method != "GET" || row.Canonical.Path != "/repos/{owner}/{repo}/issues" {
		t.Fatalf("canonical mapping = %+v, want GET issues endpoint", row.Canonical)
	}
	if !row.Runtime.Enabled {
		t.Fatalf("runtime reachability = %+v, want enabled", row.Runtime)
	}
	if !slices.Contains(row.CLI.Paths, "issue list") {
		t.Fatalf("CLI evidence = %+v, want generated issue list command", row.CLI)
	}
	if !slices.Contains(row.Website.Paths, "issue list") {
		t.Fatalf("website evidence = %+v, want generated issue list row", row.Website)
	}
	if !slices.Contains(row.Fixtures.Paths, "fixtures/streams/issues/page_1.json") {
		t.Fatalf("fixture evidence = %+v, want issue stream fixture", row.Fixtures)
	}
	if !row.Conformance.Passed || row.Conformance.Proof == "" {
		t.Fatalf("conformance evidence = %+v, want an executed proof", row.Conformance)
	}
	if got := row.Classifications[operationEvidenceClassETL]; !got.Declared || !got.Enabled {
		t.Fatalf("ETL classification = %+v, want declared and enabled", got)
	}
	for _, class := range operationEvidenceClasses {
		if _, present := row.Classifications[class]; !present {
			t.Fatalf("classification surface omitted %q from %+v", class, row.Classifications)
		}
	}
	if len(row.Gaps) != 0 {
		t.Fatalf("complete operation reported gaps: %+v", row.Gaps)
	}
	graphqlRow, found := artifact.row("github.graphql.mutation.createIpAllowListEntry")
	if !found || !slices.Contains(graphqlRow.CLI.Paths, "graphql mutation create-ip-allow-list-entry") || graphqlRow.hasGap(operationEvidenceGapCLICommand) {
		t.Fatalf("GraphQL acronym operation did not retain its exact command evidence: %+v", graphqlRow)
	}
}

func TestOperationEvidenceUsesStrictV3DocumentOperationInventory(t *testing.T) {
	root := operationEvidenceWorkspace(t)
	baseline, _, stderr := runOperationEvidenceForTest(t, root, "")
	if stderr != "" {
		t.Fatalf("baseline operation-evidence diagnostics = %s", stderr)
	}
	operationEvidenceRewriteGitHubLockAsV3(t, root)
	actual, _, stderr := runOperationEvidenceForTest(t, root, "")
	if stderr != "" {
		t.Fatalf("v3 operation-evidence diagnostics = %s", stderr)
	}
	if actual.rowCount() != baseline.rowCount() {
		t.Fatalf("v3 document-owned operation inventory row count = %d, want legacy baseline %d", actual.rowCount(), baseline.rowCount())
	}
	for _, want := range baseline.Rows {
		got, found := actual.row(want.SourceID)
		if !found {
			t.Fatalf("v3 operation-evidence omitted source row %q", want.SourceID)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("v3 operation-evidence row %q changed:\n got %#v\nwant %#v", want.SourceID, got, want)
		}
	}
	declared, found := actual.row(githubIssuesSourceID)
	if !found || !declared.Runtime.Enabled {
		t.Fatalf("v3 declared source row = %#v, want enabled %q", declared, githubIssuesSourceID)
	}
	for _, class := range operationEvidenceClasses {
		if _, found := declared.Classifications[class]; !found {
			t.Fatalf("v3 declared row lost classification lane %q: %#v", class, declared.Classifications)
		}
	}
	var deferred operationEvidenceRow
	for _, row := range baseline.Rows {
		if len(row.Gaps) != 0 || len(row.Foundations) != 0 {
			deferred = row
			break
		}
	}
	if deferred.SourceID == "" {
		t.Fatal("baseline fixture has no deferred source row")
	}
	if got, found := actual.row(deferred.SourceID); !found || !reflect.DeepEqual(got, deferred) {
		t.Fatalf("v3 deferred source row %q = %#v, want %#v", deferred.SourceID, got, deferred)
	}
	if _, found := actual.row("github.graphql.mutation.createIpAllowListEntry"); !found {
		t.Fatal("v3 REST document inventory dropped GraphQL evidence rows")
	}
}

func TestOperationEvidenceRejectsV3InventoryStateClaimingAbsence(t *testing.T) {
	root := operationEvidenceWorkspace(t)
	operationEvidenceRewriteGitHubLockAsV3(t, root)
	path := filepath.Join(root, "internal", "connectors", "defs", "github", "sources", "github-operation-source-lock.json")
	var documentOperations int
	mutateOperationEvidenceJSON(t, path, func(lock map[string]any) {
		rest := lock["rest"].(map[string]any)
		documentOperations = len(rest["source_documents"].([]any)[0].(map[string]any)["operations"].([]any))
		lock["state"] = "dynamic"
		lock["dynamic"] = map[string]any{
			"reason": "contradictory-fixture-state",
			"detail": "A document-owned operation inventory exists and must not be projected as absence.",
		}
	})
	if documentOperations != 1220 {
		t.Fatalf("v3 fixture document operations = %d, want 1220 REST operations", documentOperations)
	}
	input, err := readOperationEvidenceSourceLock(path, "github")
	if err == nil {
		if input.Absence != nil {
			t.Fatalf("v3 source document inventory with %d REST operations was classified as absence: %+v", documentOperations, input.Absence)
		}
		t.Fatal("contradictory v3 state unexpectedly projected without strict source-import rejection")
	}
	if !strings.Contains(err.Error(), "source-import schema") {
		t.Fatalf("v3 contradictory state error = %v, want strict source-import schema rejection", err)
	}
}

func TestOperationEvidenceRejectsDuplicateV3InventoryBeforeAbsenceProjection(t *testing.T) {
	t.Parallel()
	const populated = `[{"id":"retained-document"}]`
	for _, tc := range []struct {
		name  string
		first string
		last  string
	}{
		{name: "populated then empty", first: populated, last: `[]`},
		{name: "empty then populated", first: `[]`, last: populated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "github-operation-source-lock.json")
			raw := []byte(`{"schema_version":3,"connector":"github","state":"dynamic","dynamic":{"reason":"duplicate-fixture","detail":"duplicate source_documents must not become provider absence"},"rest":{"source_documents":` + tc.first + `,"source_documents":` + tc.last + `}}`)
			if err := os.WriteFile(path, raw, 0o644); err != nil {
				t.Fatalf("write duplicate v3 source lock: %v", err)
			}
			input, err := readOperationEvidenceSourceLock(path, "github")
			if err == nil || input.Absence != nil || !strings.Contains(err.Error(), "duplicate JSON object member at /rest/source_documents") {
				t.Fatalf("duplicate v3 source lock input=%+v err=%v, want duplicate rejection before absence projection", input, err)
			}
		})
	}
}

func TestOperationEvidenceReadsAsanaVersion3DocumentOwnedLock(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	input, err := readOperationEvidenceSourceLock(filepath.Join(root, "internal", "connectors", "defs", "asana", "sources", "asana-operation-source-lock.json"), "asana")
	if err != nil {
		t.Fatalf("read Asana version-3 source lock: %v", err)
	}
	if len(input.Operations) != 249 {
		t.Fatalf("Asana source operations = %d, want 249", len(input.Operations))
	}
	for _, operation := range input.Operations {
		if operation.ID != "asana.rest.getAccessRequests" {
			continue
		}
		if operation.Method != "GET" || operation.Path != "/access_requests" || operation.Trace.URL != "https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml" || operation.Trace.Location != `paths["/access_requests"].get` || operation.Trace.SHA256 == "" || operation.Trace.Bytes <= 0 {
			t.Fatalf("Asana access-request evidence = %+v, want document-owned source trace", operation)
		}
		return
	}
	t.Fatal("Asana access-request source operation is absent")
}

func TestOperationEvidenceReportsEachMissingEvidenceKind(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(t *testing.T, root string)
		want   string
	}{
		{
			name: "source trace",
			mutate: func(t *testing.T, root string) {
				mutateOperationEvidenceJSON(t, filepath.Join(root, "internal", "connectors", "defs", "github", "sources", "github-operation-source-lock.json"), func(document map[string]any) {
					rest := document["rest"].(map[string]any)
					delete(rest, "sha256")
				})
			},
			want: operationEvidenceGapSourceTrace,
		},
		{
			name: "canonical mapping",
			mutate: func(t *testing.T, root string) {
				mutateOperationEvidenceJSON(t, filepath.Join(root, "internal", "connectors", "defs", "github", "api_surface.json"), func(document map[string]any) {
					document["endpoints"] = withoutEndpoint(document["endpoints"].([]any), "GET", "/repos/{owner}/{repo}/issues")
				})
			},
			want: operationEvidenceGapCanonicalMapping,
		},
		{
			name: "runtime reachability",
			mutate: func(t *testing.T, root string) {
				mutateOperationEvidenceJSON(t, filepath.Join(root, "internal", "connectors", "defs", "github", "cli_surface.json"), func(document map[string]any) {
					mutateCommand(document["commands"].([]any), "issue list", func(command map[string]any) { command["availability"] = "partial" })
				})
			},
			want: operationEvidenceGapRuntimeReachability,
		},
		{
			name: "CLI command",
			mutate: func(t *testing.T, root string) {
				mutateOperationEvidenceJSON(t, filepath.Join(root, "internal", "connectors", "defs", "github", "cli_surface.json"), func(document map[string]any) {
					document["commands"] = withoutCommand(document["commands"].([]any), "issue list")
				})
			},
			want: operationEvidenceGapCLICommand,
		},
		{
			name: "website row",
			mutate: func(t *testing.T, root string) {
				mutateOperationEvidenceJSON(t, filepath.Join(root, "website", "data", "connectors.generated.json"), func(document map[string]any) {
					for _, item := range document["rows"].([]any) {
						row := item.(map[string]any)
						if row["slug"] != "github" {
							continue
						}
						cli := row["cli_surface"].(map[string]any)
						cli["commands"] = withoutCommand(cli["commands"].([]any), "issue list")
					}
				})
			},
			want: operationEvidenceGapWebsiteRow,
		},
		{
			name: "fixture proof",
			mutate: func(t *testing.T, root string) {
				path := filepath.Join(root, "internal", "connectors", "defs", "github", "fixtures", "streams", "issues")
				if err := os.RemoveAll(path); err != nil {
					t.Fatalf("remove fixture directory %s: %v", path, err)
				}
			},
			want: operationEvidenceGapFixtureProof,
		},
		{
			name: "conformance proof",
			mutate: func(t *testing.T, root string) {
				mutateOperationEvidenceJSON(t, filepath.Join(root, "internal", "connectors", "defs", "github", "metadata.json"), func(document map[string]any) {
					document["conformance"] = map[string]any{"skip_dynamic": true, "reason": "operation-evidence negative fixture"}
				})
			},
			want: operationEvidenceGapConformanceProof,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := operationEvidenceWorkspace(t)
			tc.mutate(t, root)
			artifact, _, _ := runOperationEvidenceForTest(t, root, "")
			row, found := artifact.row(githubIssuesSourceID)
			if !found {
				t.Fatalf("operation was silently omitted after %s loss", tc.name)
			}
			if !row.hasGap(tc.want) {
				t.Fatalf("gaps after %s loss = %+v, want %q", tc.name, row.Gaps, tc.want)
			}
		})
	}
}

func TestOperationEvidenceRecordsProviderEvidencedAbsence(t *testing.T) {
	root := operationEvidenceAbsenceWorkspace(t)
	artifact, _, _ := runOperationEvidenceForTest(t, root, "")
	row, found := artifact.row("testrail.provider-surface")
	if !found {
		t.Fatal("provider-evidenced unavailable surface was omitted")
	}
	if row.Absence == nil || row.Absence.Reason != "no-public-api-description" || row.Absence.Evidence == "" {
		t.Fatalf("absence = %+v, want provider-evidenced absence", row.Absence)
	}
	if len(row.Gaps) != 0 {
		t.Fatalf("provider-evidenced absence was rewritten as a gap: %+v", row.Gaps)
	}
}

func TestOperationEvidenceSeparatesDeclaredReadOnlyFromFoundations(t *testing.T) {
	root := operationEvidenceWorkspace(t)
	mutateOperationEvidenceJSON(t, filepath.Join(root, "internal", "connectors", "defs", "github", "api_surface.json"), func(document map[string]any) {
		for _, item := range document["endpoints"].([]any) {
			endpoint := item.(map[string]any)
			if endpoint["method"] != "GET" || endpoint["path"] != "/repos/{owner}/{repo}/actions/artifacts" {
				continue
			}
			delete(endpoint, "covered_by")
			endpoint["operation"] = map[string]any{
				"model":              "read_only",
				"status":             "blocked",
				"risk":               "low",
				"blocked_by_default": true,
				"reason":             "The connector intentionally does not implement this source-cited read.",
				"notes":              "Named policy: source-cited-read-only-operations-r1",
			}
			return
		}
		t.Fatalf("read-only source endpoint %q was not found", githubReadOnlySourceID)
	})

	artifact, _, stderr := runOperationEvidenceForTest(t, root, "")
	if stderr != "" {
		t.Fatalf("operation-evidence wrote diagnostics for read-only declaration: %s", stderr)
	}
	row, found := artifact.row(githubReadOnlySourceID)
	if !found {
		t.Fatalf("operation evidence omitted declared read-only source row %q", githubReadOnlySourceID)
	}
	if row.ReadOnly == nil || row.ReadOnly.Policy != "source-cited-read-only-operations-r1" || row.ReadOnly.Reason == "" {
		t.Fatalf("read-only evidence = %+v, row = %+v, want explicit policy and reason", row.ReadOnly, row)
	}
	if row.Runtime.Enabled || len(row.Gaps) != 0 || len(row.Foundations) != 0 {
		t.Fatalf("read-only row was rewritten as an execution or foundation gap: %+v", row)
	}
	if _, found := artifact.rollupContaining(artifact.MissingFoundations, githubReadOnlySourceID); found {
		t.Fatalf("read-only source row leaked into missing-foundation rollups: %+v", artifact.MissingFoundations)
	}
	rollup, found := artifact.readOnlyRollup("github", "source-cited-read-only-operations-r1")
	if !found || !slices.Contains(rollup.SourceIDs, githubReadOnlySourceID) {
		t.Fatalf("read-only rollup = %+v, want %q", artifact.IntentionallyReadOnly, githubReadOnlySourceID)
	}
}

func TestOperationEvidenceIsByteStableAndRollsUpDuplicates(t *testing.T) {
	root := operationEvidenceWorkspace(t)
	mutateOperationEvidenceJSON(t, filepath.Join(root, "internal", "connectors", "defs", "github", "sources", "github-operation-source-lock.json"), func(document map[string]any) {
		rest := document["rest"].(map[string]any)
		operations := rest["operations"].([]any)
		for _, item := range operations {
			operation := item.(map[string]any)
			if operation["id"] != githubIssuesSourceID {
				continue
			}
			raw, err := json.Marshal(operation)
			if err != nil {
				t.Fatalf("encode duplicate source operation: %v", err)
			}
			var duplicate map[string]any
			if err := json.Unmarshal(raw, &duplicate); err != nil {
				t.Fatalf("decode duplicate source operation: %v", err)
			}
			rest["operations"] = append(operations, duplicate)
			return
		}
		t.Fatalf("source operation %q not found", githubIssuesSourceID)
	})
	mutateOperationEvidenceJSON(t, filepath.Join(root, "website", "data", "connectors.generated.json"), func(document map[string]any) {
		document["rows"] = withoutWebsiteRow(document["rows"].([]any), "github")
	})
	first, firstBytes, _ := runOperationEvidenceForTest(t, root, "")
	second, secondBytes, _ := runOperationEvidenceForTest(t, root, "")
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatal("same inputs produced non-byte-stable operation evidence artifacts")
	}
	if first.rowCount() != second.rowCount() {
		t.Fatalf("stable artifacts have different row counts %d and %d", first.rowCount(), second.rowCount())
	}
	if first.sourceIDCount(githubIssuesSourceID) != 1 {
		t.Fatalf("duplicate provider source operation was not deterministically deduplicated")
	}
	rollup, found := first.rollup(operationEvidenceGapWebsiteRow)
	if !found {
		t.Fatalf("missing website row did not create a rollup: %+v", first.MissingFoundations)
	}
	if len(rollup.SourceIDs) < 2 || !slices.IsSorted(rollup.SourceIDs) || slices.Contains(rollup.SourceIDs[1:], rollup.SourceIDs[0]) {
		t.Fatalf("website rollup is not a deterministic deduplicated source set: %+v", rollup)
	}
}

func TestOperationEvidenceFixed100RejectsEveryRegression(t *testing.T) {
	root := operationEvidenceWorkspace(t)
	artifact, _, _ := runOperationEvidenceForTest(t, root, "")
	fixed := loadOperationEvidenceFixed100(t, root)
	if len(fixed.Rows) != 100 {
		t.Fatalf("fixed reference rows = %d, want 100", len(fixed.Rows))
	}
	if err := validateOperationEvidenceFixed100(artifact, fixed); err != nil {
		t.Fatalf("unmodified fixed cohort rejected: %v", err)
	}
	for _, expectation := range fixed.Rows {
		t.Run(expectation.SourceID, func(t *testing.T) {
			mutated := artifact.clone()
			row, found := mutated.row(expectation.SourceID)
			if !found {
				t.Fatalf("fixed source row %q absent from test artifact", expectation.SourceID)
			}
			row.Source.SHA256 = ""
			mutated.replace(row)
			if err := validateOperationEvidenceFixed100(mutated, fixed); err == nil || !strings.Contains(err.Error(), expectation.SourceID) {
				t.Fatalf("regression of %q error = %v, want source-specific gate failure", expectation.SourceID, err)
			}
		})
	}
}

func TestOperationEvidenceCheckRunsFixed100Gate(t *testing.T) {
	root := operationEvidenceWorkspace(t)
	_, _, _ = runOperationEvidenceForTest(t, root, "")
	output := filepath.Join(root, "operation-evidence.json")
	fixedPath := filepath.Join(root, "internal", "connectors", "operation-evidence-fixed-100.json")
	runCheck := func() (int, string) {
		t.Helper()
		var stdout, stderr bytes.Buffer
		code := run([]string{"operation-evidence", root, "--output", output, "--fixed-100", fixedPath, "--check"}, &stdout, &stderr)
		return code, stderr.String()
	}
	if code, stderr := runCheck(); code != 0 {
		t.Fatalf("fixed-100 check exit=%d stderr=%q", code, stderr)
	}
	mutateOperationEvidenceJSON(t, fixedPath, func(document map[string]any) {
		rows := document["rows"].([]any)
		rows[0].(map[string]any)["source_sha256"] = "regressed"
	})
	if code, stderr := runCheck(); code == 0 || !strings.Contains(stderr, "fixed-100 validation failed") {
		t.Fatalf("regressed fixed-100 check exit=%d stderr=%q", code, stderr)
	}
}

func operationEvidenceWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	copyOperationEvidenceTree(t, filepath.Join("..", "..", "internal", "connectors", "defs", "asana"), filepath.Join(root, "internal", "connectors", "defs", "asana"))
	copyOperationEvidenceTree(t, filepath.Join("..", "..", "internal", "connectors", "defs", "github"), filepath.Join(root, "internal", "connectors", "defs", "github"))
	copyOperationEvidenceFile(t, filepath.Join("..", "..", "internal", "connectors", "operation-evidence-fixed-100.json"), filepath.Join(root, "internal", "connectors", "operation-evidence-fixed-100.json"))
	copyOperationEvidenceFile(t, filepath.Join("..", "..", "internal", "connectors", "certifications", "current-subject.json"), filepath.Join(root, "internal", "connectors", "certifications", "current-subject.json"))

	websiteRaw, err := os.ReadFile(filepath.Join("..", "..", "website", "data", "connectors.generated.json"))
	if err != nil {
		t.Fatalf("read generated website data: %v", err)
	}
	var rows []any
	if err := json.Unmarshal(websiteRaw, &rows); err != nil {
		t.Fatalf("decode generated website data: %v", err)
	}
	connectorRows := make([]any, 0, 2)
	for _, item := range rows {
		row := item.(map[string]any)
		if row["slug"] == "github" || row["slug"] == "asana" {
			connectorRows = append(connectorRows, row)
		}
	}
	writeOperationEvidenceJSON(t, filepath.Join(root, "website", "data", "connectors.generated.json"), map[string]any{"rows": connectorRows})
	return root
}

func operationEvidenceRewriteGitHubLockAsV3(t *testing.T, root string) {
	t.Helper()
	path := filepath.Join(root, "internal", "connectors", "defs", "github", "sources", "github-operation-source-lock.json")
	mutateOperationEvidenceJSON(t, path, func(lock map[string]any) {
		rest := lock["rest"].(map[string]any)
		artifact := map[string]any{
			"source_url": rest["source_url"],
			"sha256":     rest["sha256"],
			"bytes":      rest["bytes"],
			"openapi":    rest["openapi"],
		}
		lock["schema_version"] = 3
		lock["rest"] = map[string]any{
			"retrieval": "hermetic operation-evidence v3 document inventory fixture",
			"openapi":   []any{rest["openapi"]},
			"source_documents": []any{map[string]any{
				"id":       "github-rest",
				"artifact": artifact,
				"published_source": map[string]any{
					"source_url":  "https://docs.github.com/rest",
					"capture_url": rest["source_url"],
					"sha256":      rest["sha256"],
					"bytes":       rest["bytes"],
					"adapter":     "operation-evidence-v3-fixture-capture",
				},
				"info_version": rest["info_version"],
				"operations":   rest["operations"],
			}},
		}
	})
}

func operationEvidenceAbsenceWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	copyOperationEvidenceFile(t, filepath.Join("..", "..", "internal", "connectors", "certifications", "current-subject.json"), filepath.Join(root, "internal", "connectors", "certifications", "current-subject.json"))
	writeOperationEvidenceJSON(t, filepath.Join(root, "internal", "connectors", "defs", "testrail", "sources", "testrail-operation-source-lock.json"), map[string]any{
		"schema_version": 3,
		"connector":      "testrail",
		"state":          "skipped",
		"skip": map[string]any{
			"attempted_url":    "https://support.testrail.com/hc/en-us/categories/7076541806228",
			"reason":           "no-public-api-description",
			"retrieval_method": "browser",
			"detail":           "The official rendered reference returned a provider security-verification page rather than a public API description.",
		},
		"rest": map[string]any{
			"representation":   "unavailable",
			"retrieval_method": "browser",
			"documents":        []any{},
			"counts": map[string]any{
				"total":     nil,
				"by_method": map[string]any{},
				"by_kind":   map[string]any{"rest": nil},
			},
			"coverage_confidence": map[string]any{
				"level": "unavailable",
				"basis": "no-public-api-description: official provider evidence is unavailable without credentials.",
			},
			"operations": []any{},
		},
	})
	writeOperationEvidenceJSON(t, filepath.Join(root, "website", "data", "connectors.generated.json"), map[string]any{"rows": []any{}})
	return root
}

func runOperationEvidenceForTest(t *testing.T, root, fixedPath string) (operationEvidenceArtifact, []byte, string) {
	t.Helper()
	output := filepath.Join(root, "operation-evidence.json")
	args := []string{"operation-evidence", root, "--output", output}
	if fixedPath != "" {
		args = append(args, "--fixed-100", fixedPath)
	}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("operation-evidence exit=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	raw, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read generated operation evidence: %v", err)
	}
	var artifact operationEvidenceArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatalf("decode generated operation evidence: %v", err)
	}
	return artifact, raw, stderr.String()
}

func loadOperationEvidenceFixed100(t *testing.T, root string) operationEvidenceFixed100 {
	t.Helper()
	fixed, err := readOperationEvidenceFixed100(filepath.Join(root, "internal", "connectors", "operation-evidence-fixed-100.json"))
	if err != nil {
		t.Fatalf("read fixed-100 reference: %v", err)
	}
	return fixed
}

func mutateOperationEvidenceJSON(t *testing.T, path string, mutate func(map[string]any)) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	mutate(document)
	writeOperationEvidenceJSON(t, path, document)
}

func writeOperationEvidenceJSON(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func copyOperationEvidenceTree(t *testing.T, source, destination string) {
	t.Helper()
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatalf("read tree %s: %v", source, err)
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", destination, err)
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())
		if entry.IsDir() {
			copyOperationEvidenceTree(t, sourcePath, destinationPath)
			continue
		}
		raw, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("read %s: %v", sourcePath, err)
		}
		if err := os.WriteFile(destinationPath, raw, 0o644); err != nil {
			t.Fatalf("write %s: %v", destinationPath, err)
		}
	}
}

func copyOperationEvidenceFile(t *testing.T, source, destination string) {
	t.Helper()
	raw, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read %s: %v", source, err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(destination), err)
	}
	if err := os.WriteFile(destination, raw, 0o644); err != nil {
		t.Fatalf("write %s: %v", destination, err)
	}
}

func withoutEndpoint(endpoints []any, method, path string) []any {
	filtered := make([]any, 0, len(endpoints))
	for _, item := range endpoints {
		endpoint := item.(map[string]any)
		if endpoint["method"] == method && endpoint["path"] == path {
			continue
		}
		filtered = append(filtered, endpoint)
	}
	return filtered
}

func withoutCommand(commands []any, path string) []any {
	filtered := make([]any, 0, len(commands))
	for _, item := range commands {
		command := item.(map[string]any)
		if command["path"] == path {
			continue
		}
		filtered = append(filtered, command)
	}
	return filtered
}

func withoutWebsiteRow(rows []any, slug string) []any {
	filtered := make([]any, 0, len(rows))
	for _, item := range rows {
		row := item.(map[string]any)
		if row["slug"] == slug {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func mutateCommand(commands []any, path string, mutate func(map[string]any)) {
	for _, item := range commands {
		command := item.(map[string]any)
		if command["path"] == path {
			mutate(command)
			return
		}
	}
	panic("command " + path + " is absent")
}
