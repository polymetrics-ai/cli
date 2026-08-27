package engine

import (
	"fmt"
	"io/fs"
	"net/url"
	"regexp"
	"strings"
)

// CompositeProviderPathIdentity cites one explicitly admitted composite
// provider identity. It is a closed declaration of the sole source binding
// and exact runtime inverse; it is not a template-substitution mechanism.
type CompositeProviderPathIdentity struct {
	Connector    string                         `json:"connector"`
	SourceURL    string                         `json:"source_url"`
	SourceSHA256 string                         `json:"source_sha256"`
	Placeholder  string                         `json:"placeholder"`
	ConfigKeys   []string                       `json:"config_keys"`
	Bindings     []CompositeProviderPathBinding `json:"bindings"`
}

// CompositeProviderPathBinding records one provider operation and the one
// command binding eligible to use the closed composite identity proof.
type CompositeProviderPathBinding struct {
	Order               int    `json:"order"`
	SourceID            string `json:"source_id"`
	ProviderOperationID string `json:"provider_operation_id"`
	SourceLocation      string `json:"source_location"`
	Intent              string `json:"intent"`
	BindingKind         string `json:"binding_kind"`
	BindingID           string `json:"binding_id"`
	Method              string `json:"method"`
	Path                string `json:"path"`
}

const compositeProviderPathIdentityFile = "composite_provider_path_identity.json"

var (
	compositeProviderPathSHA256RE = regexp.MustCompile(`^[a-f0-9]{64}$`)
	compositeProviderPathTokenRE  = regexp.MustCompile(`^[-A-Za-z0-9_]+$`)
)

// validateCompositeProviderPathIdentity validates a declaration-owned proof
// as a closed collection of named source bindings. It does not accept a path
// template or any runtime values from a command; the only inverse it permits
// is constructed later from the identity's listed placeholder and components.
func validateCompositeProviderPathIdentity(connector string, identity *CompositeProviderPathIdentity) error {
	if identity == nil {
		return nil
	}
	if strings.TrimSpace(identity.Connector) == "" || identity.Connector != connector {
		return fmt.Errorf("composite provider path identity connector %q does not match bundle %q", identity.Connector, connector)
	}
	sourceURL, err := url.Parse(identity.SourceURL)
	if err != nil || sourceURL.Scheme != "https" || sourceURL.Host == "" || sourceURL.RawQuery != "" || sourceURL.Fragment != "" {
		return fmt.Errorf("composite provider path identity source_url must be an absolute HTTPS artifact URL")
	}
	if !compositeProviderPathSHA256RE.MatchString(identity.SourceSHA256) {
		return fmt.Errorf("composite provider path identity source_sha256 must be a lowercase SHA-256 digest")
	}
	if !compositeProviderPathIdentityToken(identity.Placeholder) {
		return fmt.Errorf("composite provider path identity placeholder %q is invalid", identity.Placeholder)
	}
	if len(identity.ConfigKeys) < 2 {
		return fmt.Errorf("composite provider path identity requires at least two ordered config_keys")
	}
	configKeys := make(map[string]bool, len(identity.ConfigKeys))
	for _, key := range identity.ConfigKeys {
		if !compositeProviderPathIdentityToken(key) || configKeys[key] {
			return fmt.Errorf("composite provider path identity config_keys must be unique non-empty template keys")
		}
		configKeys[key] = true
	}
	if len(identity.Bindings) == 0 {
		return fmt.Errorf("composite provider path identity requires at least one source binding")
	}
	bindingKeys := make(map[string]bool, len(identity.Bindings))
	for index, binding := range identity.Bindings {
		if binding.Order != index {
			return fmt.Errorf("composite provider path identity binding %d has order %d, want %d", index, binding.Order, index)
		}
		if strings.TrimSpace(binding.SourceID) == "" || strings.TrimSpace(binding.ProviderOperationID) == "" || strings.TrimSpace(binding.SourceLocation) == "" || strings.TrimSpace(binding.BindingID) == "" {
			return fmt.Errorf("composite provider path identity binding %d must retain its source and binding identity", index)
		}
		if !compositeProviderPathBindingLane(binding) {
			return fmt.Errorf("composite provider path identity binding %d is not an ETL stream or reverse-ETL write", index)
		}
		if strings.ToUpper(binding.Method) != binding.Method || strings.TrimSpace(binding.Method) == "" {
			return fmt.Errorf("composite provider path identity binding %d method must be non-empty uppercase", index)
		}
		if !compositeProviderPathCanonicalPath(binding.Path, identity.Placeholder) {
			return fmt.Errorf("composite provider path identity binding %d path must contain exactly one declared identity placeholder in a relative canonical path", index)
		}
		key := strings.Join([]string{binding.Intent, binding.BindingKind, binding.BindingID, binding.Method, binding.Path}, "\x00")
		if bindingKeys[key] {
			return fmt.Errorf("composite provider path identity binding %d duplicates a prior source binding", index)
		}
		bindingKeys[key] = true
	}
	return nil
}

func compositeProviderPathIdentityToken(value string) bool {
	return compositeProviderPathTokenRE.MatchString(value)
}

func compositeProviderPathBindingLane(binding CompositeProviderPathBinding) bool {
	return (binding.Intent == "etl" && binding.BindingKind == "stream") ||
		(binding.Intent == "reverse_etl" && binding.BindingKind == "write")
}

func compositeProviderPathCanonicalPath(path, placeholder string) bool {
	if strings.TrimSpace(path) != path || !strings.HasPrefix(path, "/") || strings.ContainsAny(path, "?#") || strings.Contains(path, "://") {
		return false
	}
	return strings.Count(path, "{"+placeholder+"}") == 1
}

func loadCompositeProviderPathIdentity(sub fs.FS, dirName string) (*CompositeProviderPathIdentity, error) {
	if !fileExists(sub, compositeProviderPathIdentityFile) {
		return nil, nil
	}
	raw, err := readFile(sub, compositeProviderPathIdentityFile)
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.compositeProviderPathIdentity.Validate(mustDecodeAny(raw)); err != nil {
		return nil, fmt.Errorf("load bundle %s: %s: %w", dirName, compositeProviderPathIdentityFile, err)
	}
	var identity CompositeProviderPathIdentity
	if err := strictDecode(raw, &identity); err != nil {
		return nil, fmt.Errorf("load bundle %s: %s: %w", dirName, compositeProviderPathIdentityFile, err)
	}
	if err := validateCompositeProviderPathIdentity(dirName, &identity); err != nil {
		return nil, fmt.Errorf("load bundle %s: %s: %w", dirName, compositeProviderPathIdentityFile, err)
	}
	return &identity, nil
}
