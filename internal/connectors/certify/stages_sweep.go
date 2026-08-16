package certify

import (
	"fmt"
)

// stageWriteSweepAllPairings runs only when Options.Full is true. It accounts
// every declared write action beyond the first live lifecycle that stages
// 12-17 exercise. An action which has not completed a production mutation,
// independent read-back, and verified cleanup is explicitly not_live. Schema
// preparation is intentionally not evidence of provider execution.
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

	inventory, err := writeActionInventoryFor(rc.opts.Connector)
	if err != nil {
		recordStage(rc, rep, "write_sweep_all_pairings", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("write action inventory: %v", err)
		})
		return nil
	}
	if len(inventory) == 0 {
		skipStage(rc, rep, "write_sweep_all_pairings",
			fmt.Sprintf("skipped: connector %q has no declared write action inventory", rc.opts.Connector))
		return nil
	}

	if rep.Capabilities.WriteActions == nil {
		rep.Capabilities.WriteActions = map[string]WriteActionResult{}
	}

	for _, item := range inventory {
		_, alreadyCoveredLive := rep.Capabilities.WriteActions[item.Action]
		pairing := item.Pairing
		stageName := fmt.Sprintf("write_sweep_%s", item.Action)
		recordStage(rc, rep, stageName, 2, func() (bool, CLIStageInfo, string) {
			if alreadyCoveredLive {
				return true, CLIStageInfo{}, ""
			}
			reason := item.Reason
			if reason == "" {
				reason = "no production certification scenario completed a provider mutation, independent read-back, and verified cleanup"
			}
			reason = "provider mutation was not run: " + reason
			rep.Capabilities.WriteActions[item.Action] = WriteActionResult{
				Result:  "not_live",
				Path:    item.Path,
				Risk:    item.Risk,
				Cleanup: pairing.Cleanup,
				Verify:  pairing.VerifyStream,
				Reason:  reason,
			}
			return false, CLIStageInfo{}, "not_live: " + reason
		})
	}

	return nil
}
