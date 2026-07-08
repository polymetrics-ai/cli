package certify

import "fmt"

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
		reason := fmt.Sprintf("skipped: connector %q has no curated direct-read certification candidate", rc.opts.Connector)
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
	switch connector {
	case "github":
		filePath := configValue(config, "direct_read_path", "README.md")
		dirPath := configValue(config, "direct_read_dir_path", ".")
		return []directReadCandidate{
			githubDirectReadCandidate("read-file", "direct_read_sweep_repo_read_file", "repo read-file", filePath, config),
			githubDirectReadCandidate("read-dir", "direct_read_sweep_repo_read_dir", "repo read-dir", dirPath, config),
		}
	default:
		return nil
	}
}

func directReadCandidateFor(connector string, config map[string]string) (directReadCandidate, bool) {
	candidates := directReadCandidatesFor(connector, config)
	if len(candidates) == 0 {
		return directReadCandidate{}, false
	}
	return candidates[0], true
}

func githubDirectReadCandidate(action, stageName, command, path string, config map[string]string) directReadCandidate {
	args := []string{
		"github", "repo", action,
		"--credential", sourceCredentialName,
		"--path", path,
	}
	if config != nil && config["direct_read_ref"] != "" {
		args = append(args, "--ref", config["direct_read_ref"])
	}
	args = append(args, "--max-bytes", "1048576", "--json")
	return directReadCandidate{StageName: stageName, Command: command, Args: args}
}

func configValue(config map[string]string, key, fallback string) string {
	if config != nil && config[key] != "" {
		return config[key]
	}
	return fallback
}
