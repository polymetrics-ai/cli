//go:build github_fixture_sweep

package cli

import "testing"

// The ordinary internal/cli suite has a fixed 20-minute package deadline.
// These tests each start a fresh pm binary against a faithful GitHub fixture;
// their full behavioral assertions run in the dedicated 45-minute CI job so
// none can consume the shared package budget.
func TestPMBinaryProvesGitHubSharedAdmissionForGeneratedDirectReadFixture(t *testing.T) {
	runPMBinaryProvesGitHubSharedAdmissionForGeneratedDirectReadFixture(t)
}

func TestPMBinaryExecutesGitHubDisputedPartialVerdictsAgainstFixture(t *testing.T) {
	runPMBinaryExecutesGitHubDisputedPartialVerdictsAgainstFixture(t)
}

func TestPMBinaryExecutesGitHubReleasedReadSurfaceAgainstFixture(t *testing.T) {
	runPMBinaryExecutesGitHubReleasedReadSurfaceAgainstFixture(t)
}

func TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle(t *testing.T) {
	runPMBinaryExecutesIssueLabelWarehouseTransportLifecycle(t)
}

func TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip(t *testing.T) {
	runFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip(t)
}
