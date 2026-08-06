package connectors

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestRateLimitReportCoalescesLongRunsIntoBoundedPolicySummary(t *testing.T) {
	report := NewRateLimitReport()
	report.Declare("fixture", RateLimitDeclarationDeclared)
	for policy := 0; policy < maxRateLimitReportPolicies+4; policy++ {
		report.RecordPolicySelection("fixture", fmt.Sprintf("policy-%d", policy), "account", "endpoint")
	}
	for request := 0; request < 1000; request++ {
		report.RecordPolicySelection("fixture", "policy-0", "account", "endpoint")
		report.RecordPacingWait("fixture", time.Millisecond)
		report.RecordRequestLatency("fixture", 2*time.Millisecond)
	}

	summary := report.Snapshot()
	if len(summary.Connectors) != 1 {
		t.Fatalf("connector count = %d, want 1", len(summary.Connectors))
	}
	connector := summary.Connectors[0]
	if len(connector.Policies) != maxRateLimitReportPolicies || connector.PoliciesOmitted != 4 {
		t.Fatalf("bounded policies = %d with %d omitted, want %d with 4 omitted", len(connector.Policies), connector.PoliciesOmitted, maxRateLimitReportPolicies)
	}
	if connector.PacingWaitMS != 1000 || connector.RequestLatencyMS != 2000 || connector.RequestCount != 1000 {
		t.Fatalf("coalesced totals = %+v", connector)
	}
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if len(encoded) > 4096 {
		t.Fatalf("bounded summary is unexpectedly large (%d bytes): %s", len(encoded), encoded)
	}
}
