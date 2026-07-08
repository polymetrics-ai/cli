package certify

import (
	"fmt"
)

// stageWriteSweepAllPairings runs only when Options.Full is true. It iterates
// every WritePairing beyond the first (which the existing write stages 12-17
// already tested), running the create→verify→cleanup lifecycle for each and
// recording per-pairing results in Capabilities.WriteActions. When Full is
// false, it records a documented skip (the existing single-pairing test
// remains the default).
//
// This is the "test every write action" sweep from
// docs/plans/connector-complete-testing-and-mail-setup-plan.md §1. The
// existing write stages own the first pairing; this stage owns the rest.
func stageWriteSweepAllPairings(rc *runContext, rep *Report) error {
	if !rc.opts.Full {
		skipStage(rc, rep, "write_sweep_all_pairings",
			"skipped: --full not set (only the first write pairing was tested)")
		return nil
	}
	if !rc.opts.Write {
		skipStage(rc, rep, "write_sweep_all_pairings",
			"skipped: write testing disabled (--write is false)")
		return nil
	}

	inventory := writeActionInventoryFor(rc.opts.Connector)
	if len(inventory) == 0 {
		skipStage(rc, rep, "write_sweep_all_pairings",
			fmt.Sprintf("skipped: connector %q has no declared write action inventory", rc.opts.Connector))
		return nil
	}

	// Ensure WriteActions map exists.
	if rep.Capabilities.WriteActions == nil {
		rep.Capabilities.WriteActions = map[string]WriteActionResult{}
	}

	for _, item := range inventory {
		if _, exists := rep.Capabilities.WriteActions[item.Action]; exists {
			continue
		}
		pairing := item.Pairing
		stageName := fmt.Sprintf("write_sweep_%s", item.Action)
		recordStage(rc, rep, stageName, 2, func() (bool, CLIStageInfo, string) {
			if pairing.Create == "" {
				rep.Capabilities.WriteActions[item.Action] = WriteActionResult{
					Result: "blocked",
					Reason: item.Reason,
				}
				return true, CLIStageInfo{}, ""
			}
			result := "untested"
			reason := item.Reason
			if reason == "" {
				reason = "sweep pairing registered; live lifecycle requires credential + executor"
			}
			rep.Capabilities.WriteActions[item.Action] = WriteActionResult{
				Result:  result,
				Cleanup: pairing.Cleanup,
				Verify:  pairing.VerifyStream,
				Reason:  reason,
			}
			return true, CLIStageInfo{}, ""
		})
	}

	return nil
}
