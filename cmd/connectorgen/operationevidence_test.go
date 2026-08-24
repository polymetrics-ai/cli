package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const githubIssuesSourceID = "github.rest.issues/list-for-repo"

var dockerHubSCIMWriteSourceIDs = []string{
	"dockerhub.rest.post_/v2/scim/2.0/Users",
	"dockerhub.rest.put_/v2/scim/2.0/Users/{id}",
}

func TestOperationEvidenceClassForCommandUsesIntentNotOperationName(t *testing.T) {
	if got := operationEvidenceClassForCommand("binary_upload", "releases_upload_asset"); got != operationEvidenceClassBinaryUpload {
		t.Fatalf("binary_upload classification = %q, want %q", got, operationEvidenceClassBinaryUpload)
	}
	if got := operationEvidenceClassForCommand("direct_write", "releases_upload_asset"); got != operationEvidenceClassDirectWrite {
		t.Fatalf("direct_write classification = %q, want %q; an operation name must not promote an ordinary write", got, operationEvidenceClassDirectWrite)
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

func TestOperationEvidenceFixed100UsesRuntimePreflightForDockerHubSCIMWrites(t *testing.T) {
	artifact, err := buildOperationEvidence(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("build operation evidence: %v", err)
	}

	for _, sourceID := range dockerHubSCIMWriteSourceIDs {
		row, found := artifact.row(sourceID)
		if !found {
			t.Fatalf("operation evidence omitted Docker Hub SCIM row %q", sourceID)
		}
		if row.Runtime.Enabled {
			t.Fatalf("Docker Hub SCIM row %q is runtime-enabled despite commandrunner preflight refusal: %+v", sourceID, row.Runtime)
		}
		if !row.hasGap(operationEvidenceGapRuntimeReachability) {
			t.Fatalf("Docker Hub SCIM row %q gaps = %+v, want %q", sourceID, row.Gaps, operationEvidenceGapRuntimeReachability)
		}
	}

	fixed, err := buildOperationEvidenceFixed100(artifact)
	if err != nil {
		t.Fatalf("build fixed-100 cohort: %v", err)
	}
	if len(fixed.Rows) != 100 {
		t.Fatalf("would-be fixed cohort rows = %d, want 100", len(fixed.Rows))
	}
	for _, row := range fixed.Rows {
		if slices.Contains(dockerHubSCIMWriteSourceIDs, row.SourceID) {
			t.Fatalf("would-be fixed cohort selected preflight-rejected Docker Hub SCIM row %q", row.SourceID)
		}
	}
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
	t.Run("source row removal", func(t *testing.T) {
		removalRoot := operationEvidenceWorkspace(t)
		removeOperationEvidenceSourceOperation(t, removalRoot, githubIssuesSourceID)
		removed, _, _ := runOperationEvidenceForTest(t, removalRoot, "")
		if err := validateOperationEvidenceFixed100(removed, fixed); err == nil || !strings.Contains(err.Error(), githubIssuesSourceID) {
			t.Fatalf("GitHub source-row removal error = %v, want source-specific fixed-cohort failure", err)
		}
	})
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
	removalRoot := operationEvidenceWorkspace(t)
	removalOutput := filepath.Join(removalRoot, "operation-evidence.json")
	removalFixedPath := filepath.Join(removalRoot, "internal", "connectors", "operation-evidence-fixed-100.json")
	removeOperationEvidenceSourceOperation(t, removalRoot, githubIssuesSourceID)
	_, _, _ = runOperationEvidenceForTest(t, removalRoot, "")
	var removalStdout, removalStderr bytes.Buffer
	removalCode := run([]string{"operation-evidence", removalRoot, "--output", removalOutput, "--fixed-100", removalFixedPath, "--check"}, &removalStdout, &removalStderr)
	if removalCode == 0 || !strings.Contains(removalStderr.String(), githubIssuesSourceID) {
		t.Fatalf("GitHub source-row removal check exit=%d stderr=%q, want source-specific fixed-100 failure", removalCode, removalStderr.String())
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
	fixedPath := filepath.Join(root, "internal", "connectors", "operation-evidence-fixed-100.json")
	copyOperationEvidenceFile(t, filepath.Join("..", "..", "internal", "connectors", "operation-evidence-fixed-100.json"), fixedPath)
	copyOperationEvidenceFile(t, filepath.Join("..", "..", "internal", "connectors", "certifications", "current-subject.json"), filepath.Join(root, "internal", "connectors", "certifications", "current-subject.json"))
	connectors := operationEvidenceFixedConnectorNames(t, loadOperationEvidenceFixed100(t, root))
	connectorSet := make(map[string]bool, len(connectors))
	for _, connector := range connectors {
		connectorSet[connector] = true
		copyOperationEvidenceTree(t, filepath.Join("..", "..", "internal", "connectors", "defs", connector), filepath.Join(root, "internal", "connectors", "defs", connector))
	}

	websiteRaw, err := os.ReadFile(filepath.Join("..", "..", "website", "data", "connectors.generated.json"))
	if err != nil {
		t.Fatalf("read generated website data: %v", err)
	}
	var rows []any
	if err := json.Unmarshal(websiteRaw, &rows); err != nil {
		t.Fatalf("decode generated website data: %v", err)
	}
	selected := make([]any, 0, len(connectors))
	for _, item := range rows {
		row := item.(map[string]any)
		if connectorSet[row["slug"].(string)] {
			selected = append(selected, row)
		}
	}
	writeOperationEvidenceJSON(t, filepath.Join(root, "website", "data", "connectors.generated.json"), map[string]any{"rows": selected})
	return root
}

func operationEvidenceFixedConnectorNames(t *testing.T, fixed operationEvidenceFixed100) []string {
	t.Helper()
	seen := make(map[string]bool)
	for _, row := range fixed.Rows {
		connector, _, found := strings.Cut(row.SourceID, ".")
		if !found || connector == "" {
			t.Fatalf("fixed source ID %q has no connector prefix", row.SourceID)
		}
		seen[connector] = true
	}
	connectors := make([]string, 0, len(seen))
	for connector := range seen {
		connectors = append(connectors, connector)
	}
	slices.Sort(connectors)
	return connectors
}

func removeOperationEvidenceSourceOperation(t *testing.T, root, sourceID string) {
	t.Helper()
	connector, _, found := strings.Cut(sourceID, ".")
	if !found || connector == "" {
		t.Fatalf("source ID %q has no connector prefix", sourceID)
	}
	path := filepath.Join(root, "internal", "connectors", "defs", connector, "sources", connector+"-operation-source-lock.json")
	mutateOperationEvidenceJSON(t, path, func(document map[string]any) {
		rest := document["rest"].(map[string]any)
		operations := rest["operations"].([]any)
		filtered := make([]any, 0, len(operations)-1)
		removed := false
		for _, item := range operations {
			operation := item.(map[string]any)
			if operation["id"] == sourceID {
				removed = true
				continue
			}
			filtered = append(filtered, item)
		}
		if !removed {
			t.Fatalf("source operation %q was not present in %s", sourceID, path)
		}
		rest["operations"] = filtered
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
