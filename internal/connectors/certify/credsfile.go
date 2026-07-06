package certify

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// CredentialRef is the "credential" block of a creds.yaml connector entry
// (certification design §B): env/exec SECRET REFERENCES only — never a
// literal secret value — plus plain (non-secret) config key/values. Safe to
// commit to source control.
type CredentialRef struct {
	// FromEnv maps a credential field name to the ENV VARIABLE NAME that
	// holds its value at run time (e.g. {"token": "PM_CERT_GITHUB_TOKEN"}).
	FromEnv map[string]string `yaml:"from_env"`
	// Exec maps a credential field name to an argv slice; ResolveSecrets
	// runs it and captures trimmed stdout as the field's value (e.g. a
	// password-manager CLI: {"api_key": ["op", "read", "op://..."]}).
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
// env/exec secret REFERENCES only (never secret values) plus per-connector
// sandbox/write/rate_limit/skip flags, safe to commit.
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

// ResolveSecrets resolves every field in entry.Credential (FromEnv + Exec)
// to its actual secret value. Values are NEVER read from the creds file
// itself (certification design §B) — FromEnv reads the referenced
// environment variable, Exec runs the referenced command and captures its
// trimmed stdout. An unset env var or a failing/empty exec command is a
// hard error: certify must never silently proceed with an empty secret.
func ResolveSecrets(entry ConnectorCredsEntry) (map[string]string, error) {
	secrets := make(map[string]string, len(entry.Credential.FromEnv)+len(entry.Credential.Exec))

	for field, envName := range entry.Credential.FromEnv {
		v, ok := os.LookupEnv(envName)
		if !ok || v == "" {
			return nil, fmt.Errorf("certify: credential field %q references unset env var %q", field, envName)
		}
		secrets[field] = v
	}

	for field, argv := range entry.Credential.Exec {
		v, err := runExecSecret(argv)
		if err != nil {
			return nil, fmt.Errorf("certify: credential field %q exec resolution failed: %w", field, err)
		}
		secrets[field] = v
	}

	return secrets, nil
}

// runExecSecret runs argv and returns its trimmed stdout, per certification
// design §B's exec form: {api_key: ["op", "read", "op://..."]}.
func runExecSecret(argv []string) (string, error) {
	if len(argv) == 0 {
		return "", fmt.Errorf("empty exec command")
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("exec %q: %w (stderr: %s)", strings.Join(argv, " "), err, strings.TrimSpace(stderr.String()))
	}
	value := strings.TrimSpace(stdout.String())
	if value == "" {
		return "", fmt.Errorf("exec %q produced empty output", strings.Join(argv, " "))
	}
	return value, nil
}
