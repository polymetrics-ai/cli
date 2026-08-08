package certify

import (
	"strings"
	"testing"
)

func TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints(t *testing.T) {
	result, err := surfaceInventoryFor("github")
	if err != nil {
		t.Fatalf("surfaceInventoryFor(github): %v", err)
	}
	if result.Result != "pass" {
		t.Fatalf("Result = %q reason=%q", result.Result, result.Reason)
	}
	if result.Endpoints != 1224 {
		t.Fatalf("Endpoints = %d, want 1224", result.Endpoints)
	}
	if result.Covered != 1139 {
		t.Fatalf("Covered = %d, want 1139", result.Covered)
	}
	if result.Blocked != 85 {
		t.Fatalf("Blocked = %d, want 85", result.Blocked)
	}
	if result.CoveredBy["stream"] != 37 {
		t.Fatalf("CoveredBy[stream] = %d, want 37", result.CoveredBy["stream"])
	}
	// covered_by.writes is plural for the three PATCH endpoints that back
	// several write contracts each, so this count has to come from
	// WriteTargets(); reading only the singular field leaves those endpoints
	// looking uncovered and fails the whole inventory.
	if result.CoveredBy["write"] != 555 {
		t.Fatalf("CoveredBy[write] = %d, want 555", result.CoveredBy["write"])
	}
	if result.CoveredBy["direct_read"] != 368 {
		t.Fatalf("CoveredBy[direct_read] = %d, want 368", result.CoveredBy["direct_read"])
	}
	if result.CoveredBy["direct_reads"] != 186 {
		t.Fatalf("CoveredBy[direct_reads] = %d, want 186", result.CoveredBy["direct_reads"])
	}
	if result.BlockedByModel["duplicate"] != 67 {
		t.Fatalf("BlockedByModel[duplicate] = %d, want 67", result.BlockedByModel["duplicate"])
	}
	if result.Provenance == nil {
		t.Fatal("Provenance = nil, want legacy provenance evidence")
	}
	if result.Provenance.Status != "legacy_unverified" || result.Provenance.LedgerVersion != 1 {
		t.Fatalf("Provenance = %+v, want v1 legacy_unverified evidence", result.Provenance)
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
