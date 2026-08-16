package certify

import (
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

type certificationProfile struct {
	spec   *engine.CertificationSpec
	writes []engine.WriteAction
}

func certificationProfileFor(connector string) certificationProfile {
	bundle, err := engine.Load(defs.FS, connector)
	if err != nil || bundle.Certification == nil {
		return certificationProfile{}
	}
	return certificationProfile{spec: bundle.Certification, writes: bundle.Writes}
}

func certificationWriteWaveFor(connector string) (*engine.CertificationWriteWaveSpec, bool) {
	profile := certificationProfileFor(connector)
	if profile.spec == nil || profile.spec.WriteWave == nil {
		return nil, false
	}
	return profile.spec.WriteWave, true
}

func certificationHasWriteWave(connector string) bool {
	_, ok := certificationWriteWaveFor(connector)
	return ok
}

func certificationWriteInventoryClassification(connector, action, path string) (classification, reason string, declared bool, err error) {
	profile := certificationProfileFor(connector)
	if profile.spec == nil {
		return "", "", false, nil
	}
	if wave := profile.spec.WriteWave; wave != nil {
		for _, candidate := range wave.Actions {
			if candidate == action {
				return writeClassificationRepositoryWaveReady,
					"the connector-declared bounded repository scenario is not yet executed for this action", true, nil
			}
		}
	}
	rules := profile.spec.WriteInventory.Rules
	if len(rules) == 0 {
		return "", "", false, nil
	}
	for _, rule := range rules {
		for _, candidate := range rule.Actions {
			if candidate == action {
				return rule.Classification, rule.Reason, true, nil
			}
		}
		for _, candidate := range rule.Paths {
			if candidate == path {
				return rule.Classification, rule.Reason, true, nil
			}
		}
		for _, prefix := range rule.PathPrefixes {
			if strings.HasPrefix(path, prefix) {
				return rule.Classification, rule.Reason, true, nil
			}
		}
	}
	return "", "", true, fmt.Errorf("definition-owned write inventory has no classification for action %q path %q", action, path)
}

func certificationPairingsFor(connector string) []WritePairing {
	profile := certificationProfileFor(connector)
	if profile.spec == nil || len(profile.spec.WritePairings) == 0 {
		return nil
	}
	out := make([]WritePairing, 0, len(profile.spec.WritePairings))
	for _, pairing := range profile.spec.WritePairings {
		out = append(out, writePairingFromDefinition(pairing))
	}
	return out
}

func writePairingFromDefinition(pairing engine.CertificationWritePairing) WritePairing {
	out := WritePairing{
		Create:       pairing.Create,
		Cleanup:      pairing.Cleanup,
		CleanupKind:  pairing.CleanupKind,
		IDField:      pairing.IDField,
		VerifyStream: pairing.VerifyStream,
		VerifyField:  pairing.VerifyField,
	}
	if len(pairing.Overrides) > 0 {
		out.Overrides = make(map[string]any, len(pairing.Overrides))
		for key, value := range pairing.Overrides {
			out.Overrides[key] = value
		}
	}
	return out
}

func certificationDefaultStream(connector string) string {
	profile := certificationProfileFor(connector)
	if profile.spec == nil || strings.TrimSpace(profile.spec.Source.DefaultStream) == "" {
		return "customers"
	}
	return profile.spec.Source.DefaultStream
}

func applyCertificationSourceDefaults(connector string, config map[string]string) map[string]string {
	out := make(map[string]string, len(config)+1)
	for key, value := range config {
		out[key] = value
	}
	profile := certificationProfileFor(connector)
	if profile.spec == nil {
		return out
	}
	for key, value := range profile.spec.Source.SourceCredentialDefaults {
		if out[key] == "" {
			out[key] = value
		}
	}
	// Required config values select a certification-only runtime contract and
	// are declaration-owned rather than caller-controlled. In particular, a
	// caller cannot choose a non-certification tier and silently bypass a
	// provider's require_shared certification policy.
	for key, value := range profile.spec.Source.RequiredCredentialConfig {
		out[key] = value
	}
	return out
}

func certificationLiveStreamUnavailable(connector string, res CLIResult) bool {
	profile := certificationProfileFor(connector)
	if profile.spec == nil || len(profile.spec.Source.LiveUnavailable) == 0 {
		return false
	}
	text := strings.ToLower(liveUnavailableText(res))
	for _, classifier := range profile.spec.Source.LiveUnavailable {
		if classifier.Kind != "" && res.Kind != classifier.Kind {
			continue
		}
		for _, pattern := range classifier.Contains {
			if strings.Contains(text, strings.ToLower(pattern)) {
				return true
			}
		}
	}
	return false
}

func liveUnavailableText(res CLIResult) string {
	text := res.Stdout + "\n" + res.Stderr
	if errObj, _ := res.Envelope["error"].(map[string]any); errObj != nil {
		if msg, _ := errObj["message"].(string); msg != "" {
			text += "\n" + msg
		}
	}
	return text
}

func certificationCommandArgs(connector string, config map[string]string, candidate engine.CertificationCommandCandidate) []string {
	args := make([]string, 0, len(candidate.Args))
	for _, arg := range candidate.Args {
		if arg.WhenConfigKey != "" && configValue(config, arg.WhenConfigKey, "") == "" {
			continue
		}
		switch {
		case arg.Literal != "":
			args = append(args, arg.Literal)
		case arg.Connector:
			args = append(args, connector)
		case arg.SourceCredential:
			args = append(args, sourceCredentialName)
		case arg.ConfigKey != "":
			value := configValue(config, arg.ConfigKey, arg.Default)
			if value == "" && arg.OmitWhenEmpty {
				continue
			}
			args = append(args, value)
		}
	}
	return args
}

func writeActionRecordSchema(connector, actionName string) ([]byte, error) {
	profile := certificationProfileFor(connector)
	if profile.spec == nil {
		return nil, fmt.Errorf("no definition-owned certification metadata available for connector %q", connector)
	}
	for _, action := range profile.writes {
		if action.Name != actionName {
			continue
		}
		if len(action.RecordSchema) == 0 {
			return nil, fmt.Errorf("no record_schema available for %q action %q", connector, actionName)
		}
		return append([]byte(nil), action.RecordSchema...), nil
	}
	return nil, fmt.Errorf("no record_schema available for %q action %q", connector, actionName)
}
