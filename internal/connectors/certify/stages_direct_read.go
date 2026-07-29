package certify

import (
	"fmt"

	"polymetrics.ai/internal/connectors/engine"
)

type directReadCandidate struct {
	StageName string
	Command   string
	Args      []string
}

func stageDirectReadSweep(rc *runContext, rep *Report) error {
	if !rc.opts.Full {
		reason := "skipped: --full not set (direct-read sweep is full-certificate only)"
		rep.Capabilities.DirectRead = &CapabilityResult{Result: "skipped", Reason: reason}
		skipStage(rc, rep, "direct_read_sweep", reason)
		return nil
	}

	candidates := directReadCandidatesFor(rc.opts.Connector, rc.opts.Config)
	if len(candidates) == 0 {
		reason := fmt.Sprintf("skipped: connector %q has no definition-owned direct-read certification candidate", rc.opts.Connector)
		rep.Capabilities.DirectRead = &CapabilityResult{Result: "skipped", Reason: reason}
		skipStage(rc, rep, "direct_read_sweep", reason)
		return nil
	}

	passedCount := 0
	for _, candidate := range candidates {
		recordStage(rc, rep, candidate.StageName, 2, func() (bool, CLIStageInfo, string) {
			res := rc.run(candidate.Args...)
			passed, errMsg := assertKind(rc, candidate.StageName, res, "ConnectorCommandDirectRead", 0)
			if !passed {
				rep.Capabilities.DirectRead = &CapabilityResult{Result: "fail", Stream: candidate.Command, StagesChecked: len(candidates), Reason: errMsg}
				return false, cliInfoFrom(res), errMsg
			}
			if hits := ScanForSecrets(res.Stdout, secretValuesFromEnv(rc.opts.SecretEnv)); len(hits) != 0 {
				errMsg := fmt.Sprintf("%s: secret value leaked in output: %v", candidate.StageName, hits)
				rep.Capabilities.DirectRead = &CapabilityResult{Result: "fail", Stream: candidate.Command, StagesChecked: len(candidates), Reason: errMsg}
				return false, cliInfoFrom(res), errMsg
			}
			passedCount++
			return true, cliInfoFrom(res), ""
		})
	}
	if passedCount == len(candidates) {
		rep.Capabilities.DirectRead = &CapabilityResult{Result: "pass", StagesChecked: len(candidates)}
	}
	return nil
}

func directReadCandidatesFor(connector string, config map[string]string) []directReadCandidate {
	profile := certificationProfileFor(connector)
	if profile.spec == nil || len(profile.spec.DirectReadCandidates) == 0 {
		return nil
	}
	out := make([]directReadCandidate, 0, len(profile.spec.DirectReadCandidates))
	for _, candidate := range profile.spec.DirectReadCandidates {
		out = append(out, commandCandidateFor(connector, config, candidate))
	}
	return out
}

func directReadCandidateFor(connector string, config map[string]string) (directReadCandidate, bool) {
	candidates := directReadCandidatesFor(connector, config)
	if len(candidates) == 0 {
		return directReadCandidate{}, false
	}
	return candidates[0], true
}

func commandCandidateFor(connector string, config map[string]string, candidate engine.CertificationCommandCandidate) directReadCandidate {
	return directReadCandidate{
		StageName: candidate.StageName,
		Command:   candidate.Command,
		Args:      certificationCommandArgs(connector, config, candidate),
	}
}

func configValue(config map[string]string, key, fallback string) string {
	if config != nil && config[key] != "" {
		return config[key]
	}
	return fallback
}
