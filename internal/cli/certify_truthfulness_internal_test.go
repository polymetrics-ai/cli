package cli

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

func TestRenderCertifyReportShowsPartialForNonLiveWriteCoverage(t *testing.T) {
	rep := certify.Report{
		Connector: "github",
		Mode:      "partial_live",
		Passed:    true,
		Stages: []certify.StageResult{{
			Name:   "write_sweep_update_issue",
			Passed: false,
			Error:  "not_live: provider mutation was not run: requires a run-owned issue fixture",
		}},
	}
	rendered := renderCertifyReportText(rep)
	for _, want := range []string{"Legacy certification run: github [PARTIAL]", "stage write_sweep_update_issue: not live:"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered certification = %q, want %q", rendered, want)
		}
	}
	if strings.Contains(rendered, "[PASS]") || strings.Contains(rendered, "FAILED") {
		t.Fatalf("rendered certification flattens non-live coverage: %q", rendered)
	}
}
