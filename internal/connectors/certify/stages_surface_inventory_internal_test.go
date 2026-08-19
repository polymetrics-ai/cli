package certify

import (
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints(t *testing.T) {
	result, err := surfaceInventoryFor("github")
	if err != nil {
		t.Fatalf("surfaceInventoryFor(github): %v", err)
	}
	if result.Result != "pass" {
		t.Fatalf("Result = %q reason=%q", result.Result, result.Reason)
	}
	transportEndpoints, transportOperationIDs := githubFixedGraphQLTransportCoverage(t)
	if transportEndpoints != 1 {
		t.Fatalf("Fixed GraphQL transport endpoints = %d, want exactly one shared POST /graphql endpoint", transportEndpoints)
	}
	sourceRootOperations := 0
	supplementalOperations := make([]string, 0)
	for _, operation := range transportOperationIDs {
		if strings.HasPrefix(operation, "github.graphql.query.") || strings.HasPrefix(operation, "github.graphql.mutation.") {
			sourceRootOperations++
			continue
		}
		supplementalOperations = append(supplementalOperations, operation)
	}
	if sourceRootOperations != githubSourceLockedGraphQLRootCount(t) {
		t.Fatalf("Fixed GraphQL source-root operations = %d, want every source-locked GraphQL root %d", sourceRootOperations, githubSourceLockedGraphQLRootCount(t))
	}
	if got, want := strings.Join(supplementalOperations, ","), "github.repo.list"; got != want {
		t.Fatalf("Fixed GraphQL supplemental operations = %q, want %q", got, want)
	}
	transportOperations := len(transportOperationIDs)
	wantEndpoints := githubSourceLockedRESTCount(t) + githubLegacyGraphQLBindingCount(t) + transportEndpoints
	if result.Endpoints != wantEndpoints {
		t.Fatalf("Endpoints = %d, want source-derived REST plus legacy bindings plus fixed GraphQL transport %d", result.Endpoints, wantEndpoints)
	}
	wantCovered := wantEndpoints - 1
	if result.Covered != wantCovered {
		t.Fatalf("Covered = %d, want %d executable endpoints plus one blocked duplicate", result.Covered, wantCovered)
	}
	if result.Blocked != 1 {
		t.Fatalf("Blocked = %d, want the retired user-project draft REST route only", result.Blocked)
	}
	if result.CoveredBy["stream"] != 37 {
		t.Fatalf("CoveredBy[stream] = %d, want 37", result.CoveredBy["stream"])
	}
	// covered_by.writes is plural for the three PATCH endpoints that back
	// several write contracts each, so this count has to come from
	// WriteTargets(); reading only the singular field leaves those endpoints
	// looking uncovered and fails the whole inventory.
	if result.CoveredBy["write"] != 606 {
		t.Fatalf("CoveredBy[write] = %d, want 606 after retiring the invalid user-project draft REST write", result.CoveredBy["write"])
	}
	// Eight formerly singular provider bindings now expose a source CLI alias
	// beside the generated provider command, so surface-sync correctly moves
	// each one from direct_read to direct_reads. The remaining seven promoted
	// reads add aliases to endpoints that were already plural. This is a net
	// change of -8 singular bindings and +23 plural targets from the locked
	// pre-parity inventory (366 singular, 252 plural targets).
	if result.CoveredBy["direct_read"] != 358 {
		t.Fatalf("CoveredBy[direct_read] = %d, want 358 after eight alias-backed bindings became plural", result.CoveredBy["direct_read"])
	}
	if result.CoveredBy["direct_reads"] != 275 {
		t.Fatalf("CoveredBy[direct_reads] = %d, want 275 including all 23 parity alias targets", result.CoveredBy["direct_reads"])
	}
	if result.CoveredBy["operation"] != transportOperations {
		t.Fatalf("CoveredBy[operation] = %d, want all fixed GraphQL bindings %d", result.CoveredBy["operation"], transportOperations)
	}
	if len(result.BlockedByModel) != 1 || result.BlockedByModel["duplicate"] != 1 ||
		len(result.BlockedByStatus) != 1 || result.BlockedByStatus["blocked"] != 1 {
		t.Fatalf("blocked classifications = models=%v status=%v, want one blocked duplicate", result.BlockedByModel, result.BlockedByStatus)
	}
	assertGitHubUserDraftRESTRouteIsBlockedDuplicate(t)
	if result.Provenance == nil {
		t.Fatal("Provenance = nil, want legacy provenance evidence")
	}
	if result.Provenance.Status != "legacy_unverified" || result.Provenance.LedgerVersion != 1 {
		t.Fatalf("Provenance = %+v, want v1 legacy_unverified evidence", result.Provenance)
	}
}

func assertGitHubUserDraftRESTRouteIsBlockedDuplicate(t *testing.T) {
	t.Helper()
	path, err := findAPISurfacePath("github")
	if err != nil {
		t.Fatalf("findAPISurfacePath(github): %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read github api surface: %v", err)
	}
	surface, err := engine.ParseAPISurface(raw)
	if err != nil {
		t.Fatalf("parse github api surface: %v", err)
	}

	const retiredPath = "/user/{user_id}/projectsV2/{project_number}/drafts"
	matches := 0
	for _, endpoint := range surface.Endpoints {
		if endpoint.Method != "POST" || endpoint.Path != retiredPath {
			continue
		}
		matches++
		if endpoint.CoveredBy != nil {
			t.Fatalf("%s still claims executable coverage: %+v", retiredPath, endpoint.CoveredBy)
		}
		operation := endpoint.Operation
		if operation == nil || operation.Model != "duplicate" || operation.Status != "blocked" ||
			!operation.BlockedByDefault || operation.Reason == "" ||
			operation.DuplicateOf != "POST /graphql (github.graphql.mutation.add-project-v2-draft-issue)" {
			t.Fatalf("%s operation = %+v, want the fixed GraphQL command's blocked duplicate", retiredPath, operation)
		}
	}
	if matches != 1 {
		t.Fatalf("%s occurrences = %d, want exactly one source-inventory row", retiredPath, matches)
	}
}

// TestSurfaceInventoryCountsPluralOnlyWriteCoverage pins the two shipped
// bundles whose covered_by rows use ONLY the plural `writes` spelling. github
// cannot catch a regression here: all 231 of its write rows use the singular
// `write`, so hasSurfaceCoverage/addSurfaceCoverageCounts pass identically
// with or without plural support. jira and workday-rest are the real form —
// 292 and 252 endpoints with a `writes` array and no `write` — so reverting
// either function turns every one of those rows into "neither covered nor
// blocked" and fails the stage outright.
func TestSurfaceInventoryCountsPluralOnlyWriteCoverage(t *testing.T) {
	tests := []struct {
		connector       string
		endpoints       int
		covered         int
		blocked         int
		streams         int
		writes          int
		directReads     int
		blockedModel    string
		blockedModelHas int
	}{
		{
			connector: "jira", endpoints: 617, covered: 590, blocked: 27,
			streams: 3, writes: 292, directReads: 295,
			blockedModel: "sensitive_reverse_etl", blockedModelHas: 25,
		},
		{
			connector: "workday-rest", endpoints: 911, covered: 910, blocked: 1,
			streams: 3, writes: 252, directReads: 659,
			blockedModel: "deprecated", blockedModelHas: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.connector, func(t *testing.T) {
			result, err := surfaceInventoryFor(tc.connector)
			if err != nil {
				t.Fatalf("surfaceInventoryFor(%s): %v", tc.connector, err)
			}
			// Classification: every plural-only row must land in Covered, not
			// in the "neither covered nor blocked" fail branch.
			if result.Result != "pass" {
				t.Fatalf("Result = %q reason=%q, want pass", result.Result, result.Reason)
			}
			if result.Endpoints != tc.endpoints {
				t.Errorf("Endpoints = %d, want %d", result.Endpoints, tc.endpoints)
			}
			if result.Covered != tc.covered {
				t.Errorf("Covered = %d, want %d", result.Covered, tc.covered)
			}
			if result.Blocked != tc.blocked {
				t.Errorf("Blocked = %d, want %d", result.Blocked, tc.blocked)
			}
			// Write COUNT: addSurfaceCoverageCounts must total the plural
			// arrays, not just report a non-zero presence.
			if result.CoveredBy["write"] != tc.writes {
				t.Errorf("CoveredBy[write] = %d, want %d", result.CoveredBy["write"], tc.writes)
			}
			if result.CoveredBy["stream"] != tc.streams {
				t.Errorf("CoveredBy[stream] = %d, want %d", result.CoveredBy["stream"], tc.streams)
			}
			if result.CoveredBy["direct_reads"] != tc.directReads {
				t.Errorf("CoveredBy[direct_reads] = %d, want %d", result.CoveredBy["direct_reads"], tc.directReads)
			}
			if result.BlockedByModel[tc.blockedModel] != tc.blockedModelHas {
				t.Errorf("BlockedByModel[%s] = %d, want %d", tc.blockedModel, result.BlockedByModel[tc.blockedModel], tc.blockedModelHas)
			}
		})
	}
}

// TestSurfaceInventoryPluralOnlyBundlesUseNoSingularWrite guards the premise
// of the test above: if a future regeneration rewrote jira/workday-rest to the
// singular spelling, the pinned counts would still pass while no longer
// exercising plural support at all.
func TestSurfaceInventoryPluralOnlyBundlesUseNoSingularWrite(t *testing.T) {
	for connector, wantPlural := range map[string]int{"jira": 292, "workday-rest": 252} {
		t.Run(connector, func(t *testing.T) {
			path, err := findAPISurfacePath(connector)
			if err != nil {
				t.Fatalf("findAPISurfacePath(%s): %v", connector, err)
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s api surface: %v", connector, err)
			}
			surface, err := engine.ParseAPISurface(raw)
			if err != nil {
				t.Fatalf("parse %s api surface: %v", connector, err)
			}
			plural, singular := 0, 0
			for _, ep := range surface.Endpoints {
				if ep.CoveredBy == nil {
					continue
				}
				if ep.CoveredBy.Write != "" {
					singular++
				}
				if len(ep.CoveredBy.Writes) > 0 {
					plural++
				}
			}
			if singular != 0 {
				t.Errorf("%s has %d singular covered_by.write rows, want 0 (plural-only fixture)", connector, singular)
			}
			if plural != wantPlural {
				t.Errorf("%s has %d plural covered_by.writes rows, want %d", connector, plural, wantPlural)
			}
		})
	}
}

// TestSurfaceInventoryCountsPluralWriteCoverage pins the plural spelling of
// covered_by. An endpoint that backs several write contracts names them under
// "writes" and leaves "write" empty; reading only the singular field made it
// read as uncovered, and with no typed operation block to fall back to, the
// whole inventory failed.
func TestSurfaceInventoryCountsPluralWriteCoverage(t *testing.T) {
	raw := `{
		"api": "test API",
		"endpoints": [{
			"method": "PATCH",
			"path": "/widgets/{id}",
			"covered_by": {"writes": ["update_widget", "close_widget", "reopen_widget"]}
		}]
	}`

	result, err := surfaceInventoryFromRaw([]byte(raw))
	if err != nil {
		t.Fatalf("surfaceInventoryFromRaw: %v", err)
	}
	if result.Result != "pass" {
		t.Fatalf("Result = %q reason = %q, want pass", result.Result, result.Reason)
	}
	if result.Covered != 1 {
		t.Fatalf("Covered = %d, want 1", result.Covered)
	}
	if result.CoveredBy["write"] != 3 {
		t.Fatalf("CoveredBy[write] = %d, want all 3 plural targets counted", result.CoveredBy["write"])
	}
}

func TestSurfaceInventoryCountsFixedGraphQLOperationCoverage(t *testing.T) {
	raw := `{
		"api": "test API",
		"endpoints": [{
			"method": "POST",
			"path": "/graphql",
			"covered_by": {"operations": ["acme.graphql.query.viewer", "acme.graphql.mutation.close_widget"]}
		}]
	}`

	result, err := surfaceInventoryFromRaw([]byte(raw))
	if err != nil {
		t.Fatalf("surfaceInventoryFromRaw: %v", err)
	}
	if result.Result != "pass" {
		t.Fatalf("Result = %q reason = %q, want pass", result.Result, result.Reason)
	}
	if result.Covered != 1 {
		t.Fatalf("Covered = %d, want one physical transport endpoint", result.Covered)
	}
	if result.CoveredBy["operation"] != 2 {
		t.Fatalf("CoveredBy[operation] = %d, want both fixed operation IDs", result.CoveredBy["operation"])
	}
}

func TestSurfaceInventoryFromRawReportsProvenanceEvidence(t *testing.T) {
	tests := []struct {
		name             string
		raw              string
		wantResult       string
		wantStatus       string
		wantLedger       int
		wantArtifacts    int
		wantEndpoints    int
		wantCited        int
		wantReasonSubstr string
		wantParseErr     bool
	}{
		{
			name: "complete_v2",
			raw: `{
				"api": "test API",
				"operation_ledger_version": 2,
				"artifacts": [{
					"id": "acme-openapi-2026-08-06",
					"url": "https://docs.acme.test/openapi.yaml",
					"retrieved_at": "2026-08-06"
				}],
				"endpoints": [{
					"method": "GET",
					"path": "/widgets",
					"provenance": {
						"artifact": "acme-openapi-2026-08-06",
						"source_url": "https://docs.acme.test/api/widgets"
					},
					"covered_by": {"stream": "widgets"}
				}]
			}`,
			wantResult:    "pass",
			wantStatus:    "complete",
			wantLedger:    2,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantCited:     1,
		},
		{
			name: "invalid_v2_fails_with_actionable_provenance_diagnostic",
			raw: `{
				"api": "test API",
				"operation_ledger_version": 2,
				"artifacts": [{
					"id": "acme-openapi-2026-08-06",
					"url": "https://docs.acme.test/openapi.yaml",
					"retrieved_at": "2026-08-06"
				}],
				"endpoints": [{
					"method": "GET",
					"path": "/widgets",
					"provenance": {"artifact": "acme-openapi-2026-08-06"},
					"covered_by": {"stream": "widgets"}
				}]
			}`,
			wantResult:       "fail",
			wantStatus:       "invalid",
			wantLedger:       2,
			wantArtifacts:    1,
			wantEndpoints:    1,
			wantReasonSubstr: "provenance.source_url is required",
		},
		{
			name: "pre_ledger_remains_legacy_unverified",
			raw: `{
				"api": "test API",
				"endpoints": [{
					"method": "GET",
					"path": "/widgets",
					"covered_by": {"stream": "widgets"}
				}]
			}`,
			wantResult:    "pass",
			wantStatus:    "legacy_unverified",
			wantEndpoints: 1,
		},
		{
			name: "v1_remains_legacy_unverified",
			raw: `{
				"api": "test API",
				"operation_ledger_version": 1,
				"endpoints": [{
					"method": "GET",
					"path": "/widgets",
					"covered_by": {"stream": "widgets"}
				}]
			}`,
			wantResult:    "pass",
			wantStatus:    "legacy_unverified",
			wantLedger:    1,
			wantEndpoints: 1,
		},
		{
			name: "unsupported_ledger_version_is_rejected",
			raw: `{
				"api": "test API",
				"operation_ledger_version": 3,
				"artifacts": [{
					"id": "acme-openapi-2026-08-06",
					"url": "https://docs.acme.test/openapi.yaml",
					"retrieved_at": "2026-08-06"
				}],
				"endpoints": [{
					"method": "GET",
					"path": "/widgets",
					"provenance": {
						"artifact": "acme-openapi-2026-08-06",
						"source_url": "https://docs.acme.test/api/widgets"
					},
					"covered_by": {"stream": "widgets"}
				}]
			}`,
			wantParseErr: true,
		},
		{
			name: "explicit_zero_ledger_version_is_rejected",
			raw: `{
				"api": "test API",
				"operation_ledger_version": 0,
				"endpoints": [{
					"method": "GET",
					"path": "/widgets",
					"covered_by": {"stream": "widgets"}
				}]
			}`,
			wantParseErr: true,
		},
		{
			name: "null_ledger_version_is_rejected",
			raw: `{
				"api": "test API",
				"operation_ledger_version": null,
				"endpoints": [{
					"method": "GET",
					"path": "/widgets",
					"covered_by": {"stream": "widgets"}
				}]
			}`,
			wantParseErr: true,
		},
		{
			name: "noncanonical_ledger_version_key_is_rejected",
			raw: `{
				"api": "test API",
				"Operation_Ledger_Version": 0,
				"endpoints": [{
					"method": "GET",
					"path": "/widgets",
					"covered_by": {"stream": "widgets"}
				}]
			}`,
			wantParseErr: true,
		},
		{
			name: "trailing_api_surface_is_rejected",
			raw: `{
				"api": "test API",
				"operation_ledger_version": 1,
				"endpoints": [{
					"method": "GET",
					"path": "/widgets",
					"covered_by": {"stream": "widgets"}
				}]
			} {}`,
			wantParseErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := surfaceInventoryFromRaw([]byte(tc.raw))
			if tc.wantParseErr {
				if err == nil {
					t.Fatal("surfaceInventoryFromRaw error = nil, want schema rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("surfaceInventoryFromRaw: %v", err)
			}
			if result.Result != tc.wantResult {
				t.Fatalf("Result = %q reason=%q, want %q", result.Result, result.Reason, tc.wantResult)
			}
			if result.Provenance == nil {
				t.Fatal("Provenance = nil, want evidence")
			}
			got := result.Provenance
			if got.Status != tc.wantStatus || got.LedgerVersion != tc.wantLedger || got.ArtifactCount != tc.wantArtifacts || got.EndpointCount != tc.wantEndpoints || got.CitedEndpoints != tc.wantCited {
				t.Fatalf("Provenance = %+v, want status=%q ledger=%d artifacts=%d endpoints=%d cited=%d", got, tc.wantStatus, tc.wantLedger, tc.wantArtifacts, tc.wantEndpoints, tc.wantCited)
			}
			if tc.wantReasonSubstr != "" && !strings.Contains(result.Reason, tc.wantReasonSubstr) {
				t.Fatalf("Reason = %q, want substring %q", result.Reason, tc.wantReasonSubstr)
			}
		})
	}
}
