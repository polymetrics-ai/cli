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

type binaryUploadCandidate struct {
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
		errMsg := "blocked: operation-backed binary download command is declared but safely blocked; bounded binary executor remains a future implementation gate"
		rep.Capabilities.Binary = &CapabilityResult{Result: "blocked", Stream: candidate.Command, Reason: errMsg}
		return false, cliInfoFrom(res), errMsg
	})
	return nil
}

// stageBinaryUploadSweep never turns a successfully-created reverse plan into
// a binary-upload pass. The provider transfer is only proven after the run has
// recorded byte count/digest and a provider response, followed by independent
// read-back and cleanup. Until a candidate supplies that live proof, this
// stage records the strongest honest non-pass result.
func stageBinaryUploadSweep(rc *runContext, rep *Report) error {
	if !rc.opts.Full {
		reason := "skipped: --full not set (binary-upload sweep is full-certificate only)"
		rep.Capabilities.BinaryUpload = &CapabilityResult{Result: "skipped", Reason: reason}
		skipStage(rc, rep, "binary_upload_sweep", reason)
		return nil
	}

	candidate, ok := binaryUploadCandidateFor(rc.opts.Connector, rc.opts.Config)
	if !ok {
		reason := fmt.Sprintf("skipped: connector %q has no definition-owned binary-upload certification candidate", rc.opts.Connector)
		rep.Capabilities.BinaryUpload = &CapabilityResult{Result: "skipped", Reason: reason}
		skipStage(rc, rep, "binary_upload_sweep", reason)
		return nil
	}

	recordStage(rc, rep, candidate.StageName, 2, func() (bool, CLIStageInfo, string) {
		res := rc.run(candidate.Args...)
		if hits := ScanForSecrets(res.Stdout+res.Stderr, secretValuesFromEnv(rc.opts.SecretEnv)); len(hits) != 0 {
			errMsg := fmt.Sprintf("%s: secret value leaked in output: %v", candidate.StageName, hits)
			rep.Capabilities.BinaryUpload = &CapabilityResult{Result: "fail", Stream: candidate.Command, Reason: errMsg}
			return false, cliInfoFrom(res), errMsg
		}
		if res.ExitCode != 0 {
			errMsg := fmt.Sprintf("blocked: %s: binary upload command did not reach an approval-bound plan: %s", candidate.StageName, res.Stderr)
			rep.Capabilities.BinaryUpload = &CapabilityResult{Result: "blocked", Stream: candidate.Command, Reason: errMsg}
			return false, cliInfoFrom(res), errMsg
		}
		errMsg := "not_live: binary upload command created an approval-bound plan, but no transmitted byte count, digest, provider response, independent read-back, or safe cleanup was proven"
		rep.Capabilities.BinaryUpload = &CapabilityResult{Result: "not_live", Stream: candidate.Command, Reason: errMsg}
		return false, cliInfoFrom(res), errMsg
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

func binaryUploadCandidateFor(connector string, config map[string]string) (binaryUploadCandidate, bool) {
	profile := certificationProfileFor(connector)
	if profile.spec == nil || len(profile.spec.BinaryUploadCandidates) == 0 {
		return binaryUploadCandidate{}, false
	}
	candidate := profile.spec.BinaryUploadCandidates[0]
	return binaryUploadCandidate{
		StageName: candidate.StageName,
		Command:   candidate.Command,
		Args:      certificationCommandArgs(connector, config, candidate),
	}, true
}
