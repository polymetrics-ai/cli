package certify

import "testing"

func TestWriteActionsSummaryNeverRollsNonLiveCoverageIntoPass(t *testing.T) {
	got := writeActionsSummary(map[string]WriteActionResult{
		"create_issue": {Result: "pass"},
		"update_issue": {Result: "not_live", Reason: "provider mutation was not run"},
	})
	if got != "not_live" {
		t.Fatalf("writeActionsSummary() = %q, want not_live rather than pass", got)
	}
}
