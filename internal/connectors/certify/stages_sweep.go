package certify

import (
	"fmt"
)

// stageWriteSweepAllPairings runs only when Options.Full is true. It accounts
// every declared write action beyond the first live lifecycle that stages
// 12-17 exercise. Every remaining action reaches the real declarative
// preparation path without records or a provider send. Safe pairings retain
// their separate live lifecycle; unsafe/unpaired actions retain an explicit
// reason explaining why certification did not mutate the provider.
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

	probe := rc.writeActionProbe
	if probe == nil {
		probe, err = certificationWriteActionProbe(rc.opts.Connector)
		if err != nil {
			recordStage(rc, rep, "write_sweep_all_pairings", 2, func() (bool, CLIStageInfo, string) {
				return false, CLIStageInfo{}, err.Error()
			})
			return nil
		}
	}
	for _, item := range inventory {
		_, alreadyCoveredLive := rep.Capabilities.WriteActions[item.Action]
		pairing := item.Pairing
		stageName := fmt.Sprintf("write_sweep_%s", item.Action)
		recordStage(rc, rep, stageName, 2, func() (bool, CLIStageInfo, string) {
			if err := probe(rc.ctx, rc.opts.Connector, item.Action); err != nil {
				rep.Capabilities.WriteActions[item.Action] = WriteActionResult{
					Result: "fail",
					Reason: fmt.Sprintf("write action %q declarative preparation failed: %v", item.Action, err),
				}
				return false, CLIStageInfo{}, rep.Capabilities.WriteActions[item.Action].Reason
			}
			if alreadyCoveredLive {
				return true, CLIStageInfo{}, ""
			}
			result := "pass"
			reason := item.Reason
			if pairing.Create == "" {
				if reason == "" {
					reason = "declarative preparation passed; provider mutation is not run without a certified create/cleanup lifecycle"
				} else {
					reason = "declarative preparation passed; provider mutation not run: " + reason
				}
			} else if reason == "" {
				reason = "declarative preparation passed; live lifecycle requires credential and certified cleanup"
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
