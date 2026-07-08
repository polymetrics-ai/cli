package certify

import "testing"

func TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints(t *testing.T) {
	result, err := surfaceInventoryFor("github")
	if err != nil {
		t.Fatalf("surfaceInventoryFor(github): %v", err)
	}
	if result.Result != "pass" {
		t.Fatalf("Result = %q reason=%q", result.Result, result.Reason)
	}
	if result.Endpoints != 507 {
		t.Fatalf("Endpoints = %d, want 507", result.Endpoints)
	}
	if result.Covered != 105 {
		t.Fatalf("Covered = %d, want 105", result.Covered)
	}
	if result.Blocked != 402 {
		t.Fatalf("Blocked = %d, want 402", result.Blocked)
	}
	if result.CoveredBy["stream"] != 37 {
		t.Fatalf("CoveredBy[stream] = %d, want 37", result.CoveredBy["stream"])
	}
	if result.CoveredBy["write"] != 67 {
		t.Fatalf("CoveredBy[write] = %d, want 67", result.CoveredBy["write"])
	}
	if result.CoveredBy["direct_reads"] != 2 {
		t.Fatalf("CoveredBy[direct_reads] = %d, want 2", result.CoveredBy["direct_reads"])
	}
	if result.BlockedByModel["direct_read"] != 158 {
		t.Fatalf("BlockedByModel[direct_read] = %d, want 158", result.BlockedByModel["direct_read"])
	}
	if result.BlockedByModel["binary_read"] != 10 {
		t.Fatalf("BlockedByModel[binary_read] = %d, want 10", result.BlockedByModel["binary_read"])
	}
}
