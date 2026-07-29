package certify

import (
	"fmt"
	"strings"
)

type binaryDownloadCandidate struct {
	StageName string
	Command   string
	Args      []string
}

func stageBinaryDownloadSweep(rc *runContext, rep *Report) error {
	if !rc.opts.Full {
		reason := "skipped: --full not set (binary-download sweep is full-certificate only)"
		rep.Capabilities.Binary = &CapabilityResult{Result: "skipped", Reason: reason}
		skipStage(rc, rep, "binary_download_sweep", reason)
		return nil
	}

	candidate, ok := binaryDownloadCandidateFor(rc.opts.Connector)
	if !ok {
		reason := fmt.Sprintf("skipped: connector %q has no definition-owned binary-download certification candidate", rc.opts.Connector)
		rep.Capabilities.Binary = &CapabilityResult{Result: "skipped", Reason: reason}
		skipStage(rc, rep, "binary_download_sweep", reason)
		return nil
	}

	recordStage(rc, rep, candidate.StageName, 2, func() (bool, CLIStageInfo, string) {
		res := rc.run(candidate.Args...)
		if res.ExitCode == 0 {
			errMsg := fmt.Sprintf("%s: binary command unexpectedly ran; operation-backed binary executors must stay blocked until an explicit bounded file policy is implemented", candidate.StageName)
			rep.Capabilities.Binary = &CapabilityResult{Result: "fail", Stream: candidate.Command, Reason: errMsg}
			return false, cliInfoFrom(res), errMsg
		}
		blocked := strings.Contains(res.Stderr, "operation") && strings.Contains(res.Stderr, "executor is not implemented")
		if !blocked {
			errMsg := fmt.Sprintf("%s: binary command failed for an unexpected reason: %s", candidate.StageName, res.Stderr)
			rep.Capabilities.Binary = &CapabilityResult{Result: "fail", Stream: candidate.Command, Reason: errMsg}
			return false, cliInfoFrom(res), errMsg
		}
		if hits := ScanForSecrets(res.Stdout+res.Stderr, secretValuesFromEnv(rc.opts.SecretEnv)); len(hits) != 0 {
			errMsg := fmt.Sprintf("%s: secret value leaked in output: %v", candidate.StageName, hits)
			rep.Capabilities.Binary = &CapabilityResult{Result: "fail", Stream: candidate.Command, Reason: errMsg}
			return false, cliInfoFrom(res), errMsg
		}
		rep.Capabilities.Binary = &CapabilityResult{
			Result: "blocked",
			Stream: candidate.Command,
			Reason: "operation-backed binary download command is declared but safely blocked; bounded binary executor remains a future implementation gate",
		}
		return true, cliInfoFrom(res), ""
	})
	return nil
}

func binaryDownloadCandidateFor(connector string) (binaryDownloadCandidate, bool) {
	profile := certificationProfileFor(connector)
	if profile.spec == nil || len(profile.spec.BinaryCandidates) == 0 {
		return binaryDownloadCandidate{}, false
	}
	candidate := profile.spec.BinaryCandidates[0]
	return binaryDownloadCandidate{
		StageName: candidate.StageName,
		Command:   candidate.Command,
		Args:      certificationCommandArgs(connector, nil, candidate),
	}, true
}
