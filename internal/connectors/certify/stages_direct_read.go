package certify

import (
	"fmt"
	"reflect"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

type directReadCandidate struct {
	StageName        string
	Command          string
	Args             []string
	OutputAssertions []engine.CertificationOutputAssertion
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
	checkpoint, err := newDirectReadCheckpoint(rc.opts.Connector, candidates, rc.opts.Config)
	if err != nil {
		rep.Capabilities.DirectRead = &CapabilityResult{Result: "fail", StagesChecked: len(candidates), Reason: "certify: prepare direct-read checkpoint"}
		recordStage(rc, rep, "direct_read_checkpoint", 0, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, "certify: prepare direct-read checkpoint"
		})
		return nil
	}
	if rc.opts.Resume && rc.opts.DirectReadCheckpointPath != "" {
		checkpoint, err = loadDirectReadCheckpoint(rc.opts.DirectReadCheckpointPath, rc.opts.Connector, candidates, rc.opts.Config)
		if err != nil {
			rep.Capabilities.DirectRead = &CapabilityResult{Result: "fail", StagesChecked: len(candidates), Reason: "certify: direct-read resume checkpoint is unusable"}
			recordStage(rc, rep, "direct_read_checkpoint", 0, func() (bool, CLIStageInfo, string) {
				return false, CLIStageInfo{}, "certify: direct-read resume checkpoint is unusable"
			})
			return nil
		}
	}

	passedCount := 0
	resumedCount := 0
	for _, candidate := range candidates {
		if rc.opts.Resume && checkpoint.Completed[candidate.StageName] {
			recordResumedDirectReadStage(rep, candidate.StageName)
			passedCount++
			resumedCount++
			continue
		}
		recordStage(rc, rep, candidate.StageName, 2, func() (bool, CLIStageInfo, string) {
			res := rc.run(candidate.Args...)
			passed, errMsg := assertKind(rc, candidate.StageName, res, "ConnectorCommandDirectRead", 0)
			if !passed {
				rep.Capabilities.DirectRead = &CapabilityResult{Result: "fail", Stream: candidate.Command, StagesChecked: len(candidates), Reason: errMsg}
				return false, cliInfoFrom(res), errMsg
			}
			passed, errMsg = assertDirectReadOutputAssertions(candidate.StageName, res, candidate.OutputAssertions)
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
			checkpoint.Completed[candidate.StageName] = true
			if err := saveDirectReadCheckpoint(rc.opts.DirectReadCheckpointPath, checkpoint); err != nil {
				passedCount--
				delete(checkpoint.Completed, candidate.StageName)
				errMsg = candidate.StageName + ": persist direct-read resume checkpoint"
				rep.Capabilities.DirectRead = &CapabilityResult{Result: "fail", Stream: candidate.Command, StagesChecked: len(candidates), Reason: errMsg}
				return false, cliInfoFrom(res), errMsg
			}
			return true, cliInfoFrom(res), ""
		})
	}
	if passedCount == len(candidates) {
		reason := fmt.Sprintf("pass: %d declaration-owned direct-read candidates; no whole connector command or stream surface claim", len(candidates))
		if resumedCount > 0 {
			reason = fmt.Sprintf("%s; %d resumed from a matching prior live checkpoint", reason, resumedCount)
		}
		rep.Capabilities.DirectRead = &CapabilityResult{Result: "pass", StagesChecked: len(candidates), ResumedStages: resumedCount, Reason: reason}
	}
	return nil
}

// recordResumedDirectReadStage records a prior live pass without pretending
// that this invocation made a new provider request. The checkpoint contains
// only candidate names and a declaration/configuration fingerprint.
func recordResumedDirectReadStage(rep *Report, name string) {
	rep.Stages = append(rep.Stages, StageResult{
		Name:    name,
		Tier:    2,
		Passed:  true,
		Resumed: true,
		Error:   "resumed: matching prior live direct-read checkpoint",
		CLI: CLIStageInfo{
			ArgvRedacted: "resumed from direct-read checkpoint",
			ExitCode:     0,
			Kind:         "CertificationCheckpoint",
		},
	})
}

func directReadCandidatesFor(connector string, config map[string]string) []directReadCandidate {
	profile := certificationProfileFor(connector)
	if profile.spec == nil || len(profile.spec.DirectReadCandidates) == 0 {
		return nil
	}
	cohort := strings.TrimSpace(configValue(config, "certification_cohort", ""))
	out := make([]directReadCandidate, 0, len(profile.spec.DirectReadCandidates))
	for _, candidate := range profile.spec.DirectReadCandidates {
		if cohort != "" && candidate.Cohort != cohort {
			continue
		}
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
		StageName:        candidate.StageName,
		Command:          candidate.Command,
		Args:             certificationCommandArgs(connector, config, candidate),
		OutputAssertions: append([]engine.CertificationOutputAssertion(nil), candidate.OutputAssertions...),
	}
}

// assertDirectReadOutputAssertions compares declaration-owned values against
// the parsed, sanitized direct-read envelope. Failure messages name only the
// stage and JSON Pointer: provider values are never rendered into a report.
func assertDirectReadOutputAssertions(stageName string, res CLIResult, assertions []engine.CertificationOutputAssertion) (bool, string) {
	for _, assertion := range assertions {
		actual, found := resolveCertificationJSONPointer(res.Envelope, assertion.JSONPointer)
		if !found {
			return false, fmt.Sprintf("%s: declared output at %s is absent", stageName, assertion.JSONPointer)
		}
		if assertion.ValueType != "" {
			actualType := certificationJSONValueType(actual)
			matches := actualType == assertion.ValueType || (assertion.ValueType == "object_or_array" && (actualType == "object" || actualType == "array"))
			if !matches {
				return false, fmt.Sprintf("%s: declared output at %s has the wrong type", stageName, assertion.JSONPointer)
			}
		}
		if assertion.Equals != nil && !reflect.DeepEqual(actual, assertion.Equals) {
			return false, fmt.Sprintf("%s: declared output at %s does not match", stageName, assertion.JSONPointer)
		}
	}
	return true, ""
}

func resolveCertificationJSONPointer(value any, pointer string) (any, bool) {
	if pointer == "" {
		return value, true
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, false
	}
	current := value
	for _, rawToken := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		token := strings.ReplaceAll(strings.ReplaceAll(rawToken, "~1", "/"), "~0", "~")
		switch node := current.(type) {
		case map[string]any:
			var found bool
			current, found = node[token]
			if !found {
				return nil, false
			}
		case []any:
			if token == "" || strings.HasPrefix(token, "-") {
				return nil, false
			}
			index := 0
			for _, digit := range token {
				if digit < '0' || digit > '9' {
					return nil, false
				}
				index = index*10 + int(digit-'0')
				if index >= len(node) {
					return nil, false
				}
			}
			current = node[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func certificationJSONValueType(value any) string {
	switch value.(type) {
	case nil:
		return "null"
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64:
		return "number"
	default:
		return "unknown"
	}
}

func configValue(config map[string]string, key, fallback string) string {
	if config != nil && config[key] != "" {
		return config[key]
	}
	return fallback
}
