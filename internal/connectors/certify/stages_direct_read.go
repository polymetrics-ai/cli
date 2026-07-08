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

	candidate, ok := directReadCandidateFor(rc.opts.Connector, rc.opts.Config)
	if !ok {
		reason := fmt.Sprintf("skipped: connector %q has no curated direct-read certification candidate", rc.opts.Connector)
		rep.Capabilities.DirectRead = &CapabilityResult{Result: "skipped", Reason: reason}
		skipStage(rc, rep, "direct_read_sweep", reason)
		return nil
	}

	recordStage(rc, rep, candidate.StageName, 2, func() (bool, CLIStageInfo, string) {
		res := rc.run(candidate.Args...)
		passed, errMsg := assertKind(rc, candidate.StageName, res, "ConnectorCommandDirectRead", 0)
		if !passed {
			rep.Capabilities.DirectRead = &CapabilityResult{Result: "fail", Stream: candidate.Command, Reason: errMsg}
			return false, cliInfoFrom(res), errMsg
		}
		if hits := ScanForSecrets(res.Stdout, secretValuesFromEnv(rc.opts.SecretEnv)); len(hits) != 0 {
			errMsg := fmt.Sprintf("%s: secret value leaked in output: %v", candidate.StageName, hits)
			rep.Capabilities.DirectRead = &CapabilityResult{Result: "fail", Stream: candidate.Command, Reason: errMsg}
			return false, cliInfoFrom(res), errMsg
		}
		rep.Capabilities.DirectRead = &CapabilityResult{Result: "pass", Stream: candidate.Command}
		return true, cliInfoFrom(res), ""
	})
	return nil
}

func directReadCandidateFor(connector string, config map[string]string) (directReadCandidate, bool) {
	switch connector {
	case "github":
		path := "README.md"
		if config != nil && config["direct_read_path"] != "" {
			path = config["direct_read_path"]
		}
		args := []string{
			"github", "repo", "read-file",
			"--credential", sourceCredentialName,
			"--path", path,
		}
		if config != nil && config["direct_read_ref"] != "" {
			args = append(args, "--ref", config["direct_read_ref"])
		}
		args = append(args, "--max-bytes", "1048576", "--json")
		return directReadCandidate{
			StageName: "direct_read_sweep_repo_read_file",
			Command:   "repo read-file",
			Args:      args,
		}, true
	default:
		return directReadCandidate{}, false
	}
}
