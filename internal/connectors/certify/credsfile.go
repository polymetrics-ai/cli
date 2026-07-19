package certify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
)

// CredentialRef is the "credential" block of a creds.yaml connector entry:
// environment-variable references only, never literal secret values, plus
// plain non-secret config values.
type CredentialRef struct {
	// FromEnv maps a credential field name to the ENV VARIABLE NAME that
	// holds its value at run time (e.g. {"token": "PM_CERT_GITHUB_TOKEN"}).
	FromEnv map[string]string `yaml:"from_env"`
	// Exec is retained only so the loader can detect and reject the prohibited
	// credential-file key. Certification never executes external commands.
	Exec map[string][]string `yaml:"exec"`
	// Config carries non-secret connector config (e.g. repository, base_url).
	Config map[string]string `yaml:"config"`
}

// ConnectorCredsEntry is one connector's block under creds.yaml's
// "connectors" map (certification design §B).
type ConnectorCredsEntry struct {
	Credential   CredentialRef `yaml:"credential"`
	Sandbox      bool          `yaml:"sandbox"`
	Write        bool          `yaml:"write"`
	RateLimitRPS float64       `yaml:"rate_limit_rps"`
	BudgetCalls  int           `yaml:"budget_calls"`
	Limit        int           `yaml:"limit"`
	Skip         bool          `yaml:"skip"`
	Reason       string        `yaml:"reason"`
}

// CredsDefaults is the creds.yaml top-level "defaults" block.
type CredsDefaults struct {
	Limit        int     `yaml:"limit"`
	RateLimitRPS float64 `yaml:"rate_limit_rps"`
	BudgetCalls  int     `yaml:"budget_calls"`
	Parallel     int     `yaml:"parallel"`
	parallelSet  bool
}

// CredsFile is the parsed shape of certification design §B's creds.yaml:
// environment-variable references only (never secret values) plus
// per-connector sandbox/write/rate_limit/skip flags.
type CredsFile struct {
	Version    int                            `yaml:"version"`
	Defaults   CredsDefaults                  `yaml:"defaults"`
	Connectors map[string]ConnectorCredsEntry `yaml:"connectors"`
}

const (
	maxCredsFileBytes  = 1 << 20
	maxCredsConnectors = 1000
)

var (
	connectorNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	envNamePattern       = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// LoadCredsFile strictly reads and validates a credential-reference file.
// Literal values are never resolved during parsing.
func LoadCredsFile(path string) (CredsFile, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			file, openErr := os.Open(path)
			if openErr != nil {
				return CredsFile{}, fmt.Errorf("certify: read creds file %s: %w", path, openErr)
			}
			_ = file.Close()
		}
		return CredsFile{}, fmt.Errorf("certify: inspect creds file %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return CredsFile{}, fmt.Errorf("certify: creds file must be a regular non-symlink file")
	}
	if info.Size() > maxCredsFileBytes {
		return CredsFile{}, fmt.Errorf("certify: creds file exceeds %d bytes", maxCredsFileBytes)
	}
	f, err := os.Open(path)
	if err != nil {
		return CredsFile{}, fmt.Errorf("certify: read creds file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	openedInfo, err := f.Stat()
	if err != nil {
		return CredsFile{}, fmt.Errorf("certify: inspect opened creds file %s: %w", path, err)
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		return CredsFile{}, fmt.Errorf("certify: creds file path changed while opening")
	}
	if openedInfo.Size() > maxCredsFileBytes {
		return CredsFile{}, fmt.Errorf("certify: creds file exceeds %d bytes", maxCredsFileBytes)
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxCredsFileBytes+1))
	if err != nil {
		return CredsFile{}, fmt.Errorf("certify: read creds file %s: %w", path, err)
	}
	if len(raw) > maxCredsFileBytes {
		return CredsFile{}, fmt.Errorf("certify: creds file exceeds %d bytes", maxCredsFileBytes)
	}

	var cf CredsFile
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cf); err != nil {
		return CredsFile{}, fmt.Errorf("certify: parse creds file %s: %w", path, err)
	}
	cf.Defaults.parallelSet = yamlMappingHasKey(raw, "defaults", "parallel")
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return CredsFile{}, fmt.Errorf("certify: creds file must contain exactly one YAML document")
		}
		return CredsFile{}, fmt.Errorf("certify: parse trailing creds file content: %w", err)
	}
	if err := validateCredsFile(cf); err != nil {
		return CredsFile{}, err
	}
	return cf, nil
}

func validateCredsFile(cf CredsFile) error {
	if cf.Version != 1 {
		return fmt.Errorf("certify: unsupported credentials file version %d (want 1)", cf.Version)
	}
	if len(cf.Connectors) == 0 {
		return fmt.Errorf("certify: credentials file requires at least one connector")
	}
	if len(cf.Connectors) > maxCredsConnectors {
		return fmt.Errorf("certify: credentials file has %d connectors, maximum is %d", len(cf.Connectors), maxCredsConnectors)
	}
	if cf.Defaults.parallelSet && (cf.Defaults.Parallel < 1 || cf.Defaults.Parallel > MaxParallel) {
		return fmt.Errorf("certify: defaults.parallel must be between 1 and %d when set", MaxParallel)
	}
	if err := rejectCredentialExec(cf); err != nil {
		return err
	}
	registry := bundleregistry.New()
	for _, name := range cf.ConnectorNames() {
		if !connectorNamePattern.MatchString(name) {
			return fmt.Errorf("certify: invalid connector identifier %q", name)
		}
		connector, ok := registry.Get(name)
		if !ok {
			return fmt.Errorf("certify: connector %q is not registered locally", name)
		}
		if err := validateCredentialReferenceForConnector(name, connector, cf.Connectors[name].Credential); err != nil {
			return err
		}
	}
	return nil
}

// ValidateCredentialReference validates one connector's reference-only
// credential block without resolving environment values.
func ValidateCredentialReference(name string, ref CredentialRef) error {
	if !connectorNamePattern.MatchString(name) {
		return fmt.Errorf("certify: invalid connector identifier %q", name)
	}
	connector, ok := bundleregistry.New().Get(name)
	if !ok {
		return fmt.Errorf("certify: connector %q is not registered locally", name)
	}
	return validateCredentialReferenceForConnector(name, connector, ref)
}

func validateCredentialReferenceForConnector(name string, connector connectors.Connector, ref CredentialRef) error {
	manifest := connectors.ManifestOf(connector)
	secretFields := make(map[string]bool, len(manifest.SecretFields))
	for _, field := range manifest.SecretFields {
		secretFields[field.Name] = true
	}
	if def, ok := connectors.DefinitionOf(connector); ok {
		var spec struct {
			Properties map[string]struct {
				Secret bool `json:"x-secret"`
			} `json:"properties"`
		}
		if jsonErr := json.Unmarshal(def.Spec, &spec); jsonErr == nil {
			for field, property := range spec.Properties {
				if property.Secret {
					secretFields[field] = true
				}
			}
		}
	}
	for field, envName := range ref.FromEnv {
		if strings.TrimSpace(field) == "" {
			return fmt.Errorf("certify: connector %q has an empty credential field reference", name)
		}
		if !envNamePattern.MatchString(envName) {
			return fmt.Errorf("certify: connector %q field %q has invalid environment reference", name, field)
		}
	}
	for field := range ref.Config {
		if secretFields[field] || sensitiveFieldName(field) {
			return fmt.Errorf("certify: connector %q config field %q is secret-bearing; use credential.from_env", name, field)
		}
	}
	return nil
}

func yamlMappingHasKey(raw []byte, parent, key string) bool {
	var document yaml.Node
	if err := yaml.Unmarshal(raw, &document); err != nil || len(document.Content) == 0 {
		return false
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return false
	}
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != parent || root.Content[i+1].Kind != yaml.MappingNode {
			continue
		}
		mapping := root.Content[i+1]
		for j := 0; j+1 < len(mapping.Content); j += 2 {
			if mapping.Content[j].Value == key {
				return true
			}
		}
	}
	return false
}

func sensitiveFieldName(name string) bool {
	lower := strings.ToLower(name)
	for _, fragment := range []string{"token", "secret", "password", "private_key", "api_key"} {
		if strings.Contains(lower, fragment) {
			return true
		}
	}
	return false
}

// ConnectorNames returns cf.Connectors' keys in sorted order, so batch mode
// iteration and the summary matrix render deterministically across runs.
func (cf CredsFile) ConnectorNames() []string {
	names := make([]string, 0, len(cf.Connectors))
	for name := range cf.Connectors {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// EffectiveOptions returns the per-connector settings for name, falling back
// to cf.Defaults for any zero-valued override (certification design §B:
// "per-connector token bucket" derived from defaults unless overridden).
func (cf CredsFile) EffectiveOptions(name string) ConnectorCredsEntry {
	entry := cf.Connectors[name]
	if entry.RateLimitRPS == 0 {
		entry.RateLimitRPS = cf.Defaults.RateLimitRPS
	}
	if entry.BudgetCalls == 0 {
		entry.BudgetCalls = cf.Defaults.BudgetCalls
	}
	if entry.Limit == 0 {
		entry.Limit = cf.Defaults.Limit
	}
	return entry
}

// ResolveSecrets resolves environment-variable references to secret values.
// Credential-file exec references are rejected before any environment lookup;
// certification never executes external commands.
func ResolveSecrets(entry ConnectorCredsEntry) (map[string]string, error) {
	if len(entry.Credential.Exec) != 0 {
		return nil, fmt.Errorf("certify: credential-file exec entries are not supported")
	}

	secrets := make(map[string]string, len(entry.Credential.FromEnv))
	for field, envName := range entry.Credential.FromEnv {
		v, ok := os.LookupEnv(envName)
		if !ok || v == "" {
			return nil, fmt.Errorf("certify: credential field %q references unset env var %q", field, envName)
		}
		secrets[field] = v
	}
	return secrets, nil
}

func rejectCredentialExec(cf CredsFile) error {
	for _, name := range cf.ConnectorNames() {
		if len(cf.Connectors[name].Credential.Exec) != 0 {
			return fmt.Errorf("certify: connector %q credential-file exec entries are not supported", name)
		}
	}
	return nil
}
