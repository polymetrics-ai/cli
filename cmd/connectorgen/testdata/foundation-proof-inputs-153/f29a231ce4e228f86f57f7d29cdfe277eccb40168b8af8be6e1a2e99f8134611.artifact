package engine

import (
	"fmt"
	"strings"

	"polymetrics.ai/internal/safety"
)

// ProviderQueryParameterCLIName returns the closed CLI spelling for one
// source-declared REST query key. It is deliberately an opt-in projection
// helper, not a relaxation of safety.ValidateIdentifier: source keys outside
// this narrow grammar remain source-mapped/nonimplemented until their own
// foundation is available.
//
// Plain existing parameter names retain the established underscore-to-hyphen
// spelling. The additional supported form is an ASCII identifier followed by
// one non-empty bracketed identifier segment, for example
// "filter[state]" -> "filter-state". The command declaration retains the
// exact provider key in maps_to, so request encoding remains reversible from
// the declaration rather than reconstructing a provider key from CLI text.
func ProviderQueryParameterCLIName(providerKey string) (string, bool) {
	if err := safety.ValidateIdentifier(providerKey, "provider query parameter"); err == nil {
		return strings.ReplaceAll(providerKey, "_", "-"), true
	}

	segments, ok := providerQueryParameterBracketSegments(providerKey)
	if !ok {
		return "", false
	}
	for index := range segments {
		if err := safety.ValidateIdentifier(segments[index], "provider query parameter segment"); err != nil {
			return "", false
		}
		segments[index] = strings.ReplaceAll(segments[index], "_", "-")
	}
	alias := strings.Join(segments, "-")
	if err := safety.ValidateIdentifier(alias, "provider query parameter CLI alias"); err != nil {
		return "", false
	}
	return alias, true
}

// ValidateProviderQueryParameterCLINames rejects a source declaration whose
// typed query parameters would claim the same CLI flag. It is intentionally
// evaluated per operation, where maps_to retains each original provider key.
func ValidateProviderQueryParameterCLINames(providerKeys []string) error {
	aliases := make(map[string]string, len(providerKeys))
	for _, providerKey := range providerKeys {
		alias, ok := ProviderQueryParameterCLIName(providerKey)
		if !ok {
			return fmt.Errorf("provider query parameter %q has no supported typed CLI alias", providerKey)
		}
		if existing, collision := aliases[alias]; collision && existing != providerKey {
			return fmt.Errorf("provider query parameter alias collision: %q and %q both map to --%s", existing, providerKey, alias)
		}
		aliases[alias] = providerKey
	}
	return nil
}

func providerQueryParameterBracketSegments(providerKey string) ([]string, bool) {
	firstBracket := strings.IndexByte(providerKey, '[')
	if firstBracket <= 0 || !strings.HasSuffix(providerKey, "]") {
		return nil, false
	}
	segment := providerKey[firstBracket+1 : len(providerKey)-1]
	if segment == "" || strings.ContainsAny(segment, "[]") {
		return nil, false
	}
	return []string{providerKey[:firstBracket], segment}, true
}
