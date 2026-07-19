package certify

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
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
}

// CredsFile is the parsed shape of certification design §B's creds.yaml:
// environment-variable references only (never secret values) plus
// per-connector sandbox/write/rate_limit/skip flags.
type CredsFile struct {
	Version    int                            `yaml:"version"`
	Defaults   CredsDefaults                  `yaml:"defaults"`
	Connectors map[string]ConnectorCredsEntry `yaml:"connectors"`
}

// LoadCredsFile reads and parses a creds.yaml file at path.
func LoadCredsFile(path string) (CredsFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return CredsFile{}, fmt.Errorf("certify: read creds file %s: %w", path, err)
	}
	var cf CredsFile
	if err := yaml.Unmarshal(raw, &cf); err != nil {
		return CredsFile{}, fmt.Errorf("certify: parse creds file %s: %w", path, err)
	}
	if err := rejectCredentialExec(cf); err != nil {
		return CredsFile{}, err
	}
	return cf, nil
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
