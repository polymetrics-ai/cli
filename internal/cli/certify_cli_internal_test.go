package cli

import (
	"bytes"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

func TestRenderCertifyReportTextIncludesSurfaceProvenanceEvidence(t *testing.T) {
	report := certify.Report{
		Connector: "acme",
		Passed:    true,
		Capabilities: certify.Capabilities{
			Surface: &certify.SurfaceResult{
				Result: "pass",
				Provenance: &certify.SurfaceProvenanceResult{
					Status:         "complete",
					LedgerVersion:  2,
					ArtifactCount:  1,
					EndpointCount:  3,
					CitedEndpoints: 3,
				},
			},
		},
	}
	text := renderCertifyReportText(report)

	for _, want := range []string{
		"surface:  pass",
		"provenance: complete",
		"ledger=2",
		"artifacts=1",
		"endpoints=3",
		"cited=3",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("renderCertifyReportText = %q, want %q", text, want)
		}
	}
}

func TestWriteCertifyReportJSONIncludesSurfaceProvenanceEvidence(t *testing.T) {
	report := certify.Report{
		Connector: "acme",
		Passed:    true,
		Capabilities: certify.Capabilities{
			Surface: &certify.SurfaceResult{
				Result: "pass",
				Provenance: &certify.SurfaceProvenanceResult{
					Status:         "complete",
					LedgerVersion:  2,
					ArtifactCount:  1,
					EndpointCount:  3,
					CitedEndpoints: 3,
				},
			},
		},
	}
	var output bytes.Buffer
	if err := writeCertifyReport(&output, true, report); err != nil {
		t.Fatalf("writeCertifyReport: %v", err)
	}
	for _, want := range []string{
		`"provenance"`,
		`"status": "complete"`,
		`"ledger_version": 2`,
		`"artifact_count": 1`,
		`"endpoint_count": 3`,
		`"cited_endpoints": 3`,
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("JSON output missing %q: %s", want, output.String())
		}
	}
}
