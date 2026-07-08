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
}
