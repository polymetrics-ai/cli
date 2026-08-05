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
	if result.Endpoints != 509 {
		t.Fatalf("Endpoints = %d, want 509", result.Endpoints)
	}
	if result.Covered != 440 {
		t.Fatalf("Covered = %d, want 440", result.Covered)
	}
	if result.Blocked != 69 {
		t.Fatalf("Blocked = %d, want 69", result.Blocked)
	}
	if result.CoveredBy["stream"] != 37 {
		t.Fatalf("CoveredBy[stream] = %d, want 37", result.CoveredBy["stream"])
	}
	if result.CoveredBy["write"] != 231 {
		t.Fatalf("CoveredBy[write] = %d, want 231", result.CoveredBy["write"])
	}
	if result.CoveredBy["direct_reads"] != 173 {
		t.Fatalf("CoveredBy[direct_reads] = %d, want 173", result.CoveredBy["direct_reads"])
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
	}{
		{
			name: "complete_v2",
			raw: `{
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
			name: "v1_remains_legacy_unverified",
			raw: `{
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
			name: "unsupported_ledger_version_fails",
			raw: `{
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
			wantResult:       "fail",
			wantStatus:       "invalid",
			wantLedger:       3,
			wantArtifacts:    1,
			wantEndpoints:    1,
			wantReasonSubstr: "operation_ledger_version: 3 is unsupported; expected 1 or 2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := surfaceInventoryFromRaw([]byte(tc.raw))
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
