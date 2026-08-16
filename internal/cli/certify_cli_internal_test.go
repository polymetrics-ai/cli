package cli

import (
	"bytes"
	"context"
	"io"
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

func TestRenderCertifyReportTextStatesDirectReadBoundary(t *testing.T) {
	report := certify.Report{
		Connector: "github",
		Passed:    true,
		Capabilities: certify.Capabilities{
			DirectRead: &certify.CapabilityResult{
				Result:        "pass",
				StagesChecked: 23,
				Reason:        "pass: 23 declaration-owned direct-read candidates; no whole GitHub command or stream surface claim",
			},
		},
	}

	text := renderCertifyReportText(report)
	for _, want := range []string{
		"direct-read: pass (candidates=23)",
		"23 declaration-owned direct-read candidates",
		"no whole GitHub command or stream surface claim",
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

func TestCertifyOptionsDirectReadOnlyForcesFullDirectSweep(t *testing.T) {
	opts, err := certifyOptionsFromFlags("github", parseFlags([]string{"--direct-read-only", "--from-env", "token=PM_CERT_TOKEN"}))
	if err != nil {
		t.Fatalf("certifyOptionsFromFlags() error = %v", err)
	}
	if !opts.DirectReadOnly || !opts.Full {
		t.Fatalf("direct-read-only options = %+v, want DirectReadOnly and Full", opts)
	}
}

func TestCertifyOptionsDirectReadOnlyRejectsWrite(t *testing.T) {
	_, err := certifyOptionsFromFlags("github", parseFlags([]string{"--direct-read-only", "--write"}))
	if err == nil || !strings.Contains(err.Error(), "cannot be combined with --write") {
		t.Fatalf("certifyOptionsFromFlags() error = %v, want direct-read-only/write refusal", err)
	}
}

func TestRunCertifyDirectReadOnlyRejectsExternalProof(t *testing.T) {
	err := runCertify(context.Background(), t.TempDir(), []string{"github", "--direct-read-only", "--external-proof"}, io.Discard, io.Discard, true)
	if err == nil || !strings.Contains(err.Error(), "cannot be combined with --external-proof") {
		t.Fatalf("runCertify() error = %v, want direct-read-only/external-proof refusal", err)
	}
}
