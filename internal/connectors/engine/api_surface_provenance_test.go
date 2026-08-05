package engine

import (
	"strings"
	"testing"
)

func TestValidateSurfaceProvenance(t *testing.T) {
	tests := []struct {
		name          string
		surface       *APISurface
		wantStatus    string
		wantArtifacts int
		wantEndpoints int
		wantCited     int
		wantIssue     string
	}{
		{
			name:          "complete_v2",
			surface:       completeSurfaceV2(),
			wantStatus:    SurfaceProvenanceComplete,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantCited:     1,
		},
		{
			name: "missing_row_provenance",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Endpoints[0].Provenance = nil
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantIssue:     "provenance is required",
		},
		{
			name: "missing_endpoint_citation",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Endpoints[0].Provenance.SourceURL = ""
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantIssue:     "provenance.source_url is required",
		},
		{
			name: "bad_artifact_url",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Artifacts[0].URL = "http://docs.acme.test/openapi.yaml"
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantCited:     1,
			wantIssue:     "url must be an absolute HTTPS URL",
		},
		{
			name: "missing_artifact_url",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Artifacts[0].URL = ""
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantCited:     1,
			wantIssue:     "url must be an absolute HTTPS URL",
		},
		{
			name: "missing_retrieval_date",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Artifacts[0].RetrievedAt = ""
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantCited:     1,
			wantIssue:     "retrieved_at must be an ISO-8601 full-date",
		},
		{
			name: "bad_endpoint_citation_url",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Endpoints[0].Provenance.SourceURL = "http://docs.acme.test/api/widgets"
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantCited:     1,
			wantIssue:     "provenance.source_url must be an absolute HTTPS URL",
		},
		{
			name: "unknown_artifact",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Endpoints[0].Provenance.Artifact = "unknown"
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantCited:     1,
			wantIssue:     `artifact "unknown" resolves to 0 artifacts`,
		},
		{
			name: "duplicate_artifact_id",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Artifacts = append(surface.Artifacts, surface.Artifacts[0])
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 2,
			wantEndpoints: 1,
			wantCited:     1,
			wantIssue:     `artifact "acme-openapi-2026-08-06" resolves to 2 artifacts`,
		},
		{
			name: "invalid_retrieval_date",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Artifacts[0].RetrievedAt = "2026-08-06T12:00:00Z"
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantCited:     1,
			wantIssue:     "retrieved_at must be an ISO-8601 full-date",
		},
		{
			name: "invalid_optional_sha256",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Artifacts[0].SHA256 = "not-a-sha256"
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantCited:     1,
			wantIssue:     "sha256 must be a 64-character hexadecimal digest when present",
		},
		{
			name: "v2_rejects_legacy_operation_source_url",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.Endpoints[0].Operation = &SurfaceOperation{SourceURL: "https://docs.acme.test/api/widgets"}
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantCited:     1,
			wantIssue:     "operation.source_url is v1-only",
		},
		{
			name: "v1_is_accepted_as_legacy_unverified",
			surface: &APISurface{
				OperationLedgerVersion: 1,
				Endpoints: []SurfaceEndpoint{{
					Method: "GET",
					Path:   "/widgets",
				}},
			},
			wantStatus:    SurfaceProvenanceLegacyUnverified,
			wantArtifacts: 0,
			wantEndpoints: 1,
			wantCited:     0,
		},
		{
			name: "pre_ledger_is_accepted_as_legacy_unverified",
			surface: &APISurface{
				Endpoints: []SurfaceEndpoint{{
					Method: "GET",
					Path:   "/widgets",
				}},
			},
			wantStatus:    SurfaceProvenanceLegacyUnverified,
			wantArtifacts: 0,
			wantEndpoints: 1,
			wantCited:     0,
		},
		{
			name: "unsupported_ledger_version",
			surface: func() *APISurface {
				surface := completeSurfaceV2()
				surface.OperationLedgerVersion = 3
				return surface
			}(),
			wantStatus:    SurfaceProvenanceInvalid,
			wantArtifacts: 1,
			wantEndpoints: 1,
			wantIssue:     "operation_ledger_version: 3 is unsupported; expected 1 or 2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateSurfaceProvenance(tc.surface)
			if got.Status != tc.wantStatus {
				t.Fatalf("Status = %q, want %q; issues = %v", got.Status, tc.wantStatus, got.Issues)
			}
			if got.ArtifactCount != tc.wantArtifacts || got.EndpointCount != tc.wantEndpoints || got.CitedEndpoints != tc.wantCited {
				t.Fatalf("counts = artifacts:%d endpoints:%d cited:%d, want artifacts:%d endpoints:%d cited:%d", got.ArtifactCount, got.EndpointCount, got.CitedEndpoints, tc.wantArtifacts, tc.wantEndpoints, tc.wantCited)
			}
			if tc.wantIssue == "" {
				if len(got.Issues) != 0 {
					t.Fatalf("Issues = %v, want none", got.Issues)
				}
				return
			}
			if len(got.Issues) == 0 {
				t.Fatalf("Issues = none, want one containing %q", tc.wantIssue)
			}
			issues := make([]string, 0, len(got.Issues))
			for _, issue := range got.Issues {
				issues = append(issues, issue.Error())
			}
			if !strings.Contains(strings.Join(issues, "\n"), tc.wantIssue) {
				t.Fatalf("Issues = %q, want one containing %q", issues, tc.wantIssue)
			}
		})
	}
}

func completeSurfaceV2() *APISurface {
	return &APISurface{
		OperationLedgerVersion: 2,
		Artifacts: []SurfaceArtifact{{
			ID:          "acme-openapi-2026-08-06",
			URL:         "https://docs.acme.test/openapi.yaml",
			RetrievedAt: "2026-08-06",
			SHA256:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		}},
		Endpoints: []SurfaceEndpoint{{
			Method: "GET",
			Path:   "/widgets",
			Provenance: &SurfaceProvenance{
				Artifact:  "acme-openapi-2026-08-06",
				SourceURL: "https://docs.acme.test/api/widgets",
			},
		}},
	}
}
