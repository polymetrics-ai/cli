package certify_test

import (
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

// TestParseCredsFileBasicShape covers certification-design §B: version,
// defaults (limit/rate_limit_rps/budget_calls/parallel), and per-connector
// credential (from_env), sandbox/write flags.
func TestParseCredsFileBasicShape(t *testing.T) {
	yaml := `
version: 1
defaults: {limit: 50, rate_limit_rps: 2, budget_calls: 500, parallel: 4}
connectors:
  github:
    credential:
      from_env: {token: PM_CERT_GITHUB_TOKEN}
      config:   {repository: polymetrics-ai/cert-sandbox}
    sandbox: true
    write: true
  stripe:
    credential:
      from_env: {api_key: PM_CERT_STRIPE_API_KEY}
    write: false
    rate_limit_rps: 1
  salesforce:
    skip: true
    reason: "no sandbox tenant yet"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "creds.yaml")
	writeFile(t, path, yaml)

	cf, err := certify.LoadCredsFile(path)
	if err != nil {
		t.Fatalf("LoadCredsFile() error = %v", err)
	}

	if cf.Version != 1 {
		t.Errorf("Version = %d, want 1", cf.Version)
	}
	if cf.Defaults.Limit != 50 {
		t.Errorf("Defaults.Limit = %d, want 50", cf.Defaults.Limit)
	}
	if cf.Defaults.RateLimitRPS != 2 {
		t.Errorf("Defaults.RateLimitRPS = %v, want 2", cf.Defaults.RateLimitRPS)
	}
	if cf.Defaults.BudgetCalls != 500 {
		t.Errorf("Defaults.BudgetCalls = %d, want 500", cf.Defaults.BudgetCalls)
	}
	if cf.Defaults.Parallel != 4 {
		t.Errorf("Defaults.Parallel = %d, want 4", cf.Defaults.Parallel)
	}

	if len(cf.Connectors) != 3 {
		t.Fatalf("len(Connectors) = %d, want 3", len(cf.Connectors))
	}

	gh, ok := cf.Connectors["github"]
	if !ok {
		t.Fatalf("Connectors missing github entry")
	}
	if gh.Credential.FromEnv["token"] != "PM_CERT_GITHUB_TOKEN" {
		t.Errorf("github Credential.FromEnv[token] = %q, want PM_CERT_GITHUB_TOKEN", gh.Credential.FromEnv["token"])
	}
	if gh.Credential.Config["repository"] != "polymetrics-ai/cert-sandbox" {
		t.Errorf("github Credential.Config[repository] = %q", gh.Credential.Config["repository"])
	}
	if !gh.Sandbox {
		t.Errorf("github Sandbox = false, want true")
	}
	if !gh.Write {
		t.Errorf("github Write = false, want true")
	}
	if gh.Skip {
		t.Errorf("github Skip = true, want false")
	}

	stripe, ok := cf.Connectors["stripe"]
	if !ok {
		t.Fatalf("Connectors missing stripe entry")
	}
	if stripe.Credential.FromEnv["api_key"] != "PM_CERT_STRIPE_API_KEY" {
		t.Errorf("stripe Credential.FromEnv[api_key] = %q", stripe.Credential.FromEnv["api_key"])
	}
	if stripe.Write {
		t.Errorf("stripe Write = true, want false")
	}
	if stripe.RateLimitRPS != 1 {
		t.Errorf("stripe RateLimitRPS = %v, want 1", stripe.RateLimitRPS)
	}

	sf, ok := cf.Connectors["salesforce"]
	if !ok {
		t.Fatalf("Connectors missing salesforce entry")
	}
	if !sf.Skip {
		t.Errorf("salesforce Skip = false, want true")
	}
	if sf.Reason != "no sandbox tenant yet" {
		t.Errorf("salesforce Reason = %q", sf.Reason)
	}
}

// TestParseCredsFileMissingFile surfaces a clear error rather than a panic.
func TestParseCredsFileMissingFile(t *testing.T) {
	if _, err := certify.LoadCredsFile(filepath.Join(t.TempDir(), "nope.yaml")); err == nil {
		t.Fatalf("LoadCredsFile() error = nil, want error for missing file")
	}
}

// TestParseCredsFileInvalidYAML surfaces a parse error.
func TestParseCredsFileInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "creds.yaml")
	writeFile(t, path, "connectors: [this is not, valid: yaml structure for our schema")

	if _, err := certify.LoadCredsFile(path); err == nil {
		t.Fatalf("LoadCredsFile() error = nil, want parse error for malformed YAML")
	}
}

func TestParseCredsFileRejectsExecWithoutExecutingIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "creds.yaml")
	writeFile(t, path, `
version: 1
connectors:
  sample:
    credential:
      exec: {token: ["must-not-run"]}
`)

	if _, err := certify.LoadCredsFile(path); err == nil {
		t.Fatal("LoadCredsFile() error = nil, want prohibited exec rejection")
	}
}

// TestCredsFileResolveSecretsFromEnv proves ResolveSecrets reads the
// referenced ENV VARIABLE NAMES (never literal secret values in the file
// itself — certification-design §B: "env/exec references only, safe to
// commit; never secret values").
func TestCredsFileResolveSecretsFromEnv(t *testing.T) {
	t.Setenv("PM_CERT_GITHUB_TOKEN", "ghp_topsecret123")

	entry := certify.ConnectorCredsEntry{
		Credential: certify.CredentialRef{
			FromEnv: map[string]string{"token": "PM_CERT_GITHUB_TOKEN"},
		},
	}

	secrets, err := certify.ResolveSecrets(entry)
	if err != nil {
		t.Fatalf("ResolveSecrets() error = %v", err)
	}
	if secrets["token"] != "ghp_topsecret123" {
		t.Errorf("secrets[token] = %q, want ghp_topsecret123", secrets["token"])
	}
}

// TestCredsFileResolveSecretsFromEnvMissingVarErrors ensures an unset ENV
// reference fails loudly rather than silently certifying with an empty
// secret value.
func TestCredsFileResolveSecretsFromEnvMissingVarErrors(t *testing.T) {
	entry := certify.ConnectorCredsEntry{
		Credential: certify.CredentialRef{
			FromEnv: map[string]string{"token": "PM_CERT_DEFINITELY_UNSET_VAR_XYZ"},
		},
	}

	if _, err := certify.ResolveSecrets(entry); err == nil {
		t.Fatalf("ResolveSecrets() error = nil, want error for unset env var")
	}
}

// TestCredsFileEffectiveOptionsAppliesDefaults proves a connector entry that
// doesn't override rate_limit/budget/limit falls back to CredsFile.Defaults
// (certification-design §B "Worker pool ... per-connector token bucket").
func TestCredsFileEffectiveOptionsAppliesDefaults(t *testing.T) {
	cf := certify.CredsFile{
		Version: 1,
		Defaults: certify.CredsDefaults{
			Limit: 50, RateLimitRPS: 2, BudgetCalls: 500, Parallel: 4,
		},
		Connectors: map[string]certify.ConnectorCredsEntry{
			"github": {Write: true},
			"stripe": {RateLimitRPS: 1},
		},
	}

	gh := cf.EffectiveOptions("github")
	if gh.RateLimitRPS != 2 {
		t.Errorf("github EffectiveOptions.RateLimitRPS = %v, want default 2", gh.RateLimitRPS)
	}
	if gh.BudgetCalls != 500 {
		t.Errorf("github EffectiveOptions.BudgetCalls = %d, want default 500", gh.BudgetCalls)
	}

	stripe := cf.EffectiveOptions("stripe")
	if stripe.RateLimitRPS != 1 {
		t.Errorf("stripe EffectiveOptions.RateLimitRPS = %v, want overridden 1", stripe.RateLimitRPS)
	}
	if stripe.BudgetCalls != 500 {
		t.Errorf("stripe EffectiveOptions.BudgetCalls = %d, want default 500", stripe.BudgetCalls)
	}
}

// TestCredsFileConnectorNamesSorted returns connector keys in stable
// (sorted) order so batch mode iteration and the summary matrix are
// deterministic across runs.
func TestCredsFileConnectorNamesSorted(t *testing.T) {
	cf := certify.CredsFile{
		Connectors: map[string]certify.ConnectorCredsEntry{
			"stripe":     {},
			"github":     {},
			"salesforce": {},
		},
	}

	names := cf.ConnectorNames()
	want := []string{"github", "salesforce", "stripe"}
	if len(names) != len(want) {
		t.Fatalf("ConnectorNames() = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("ConnectorNames()[%d] = %q, want %q", i, names[i], want[i])
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
