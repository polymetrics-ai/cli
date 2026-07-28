package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/pmbroker"
)

func TestPMBrokerContextCreateUseShowAndList(t *testing.T) {
	isolatePMBrokerUserConfig(t)

	createArgs := append([]string{"--json", "context", "create", "prod"}, pmBrokerContextFlags("production", "remote")...)
	var stdout, stderr bytes.Buffer
	if code := Run(createArgs, &stdout, &stderr); code != 0 {
		t.Fatalf("context create exit = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	var created struct {
		Kind    string `json:"kind"`
		Context struct {
			Name         string `json:"name"`
			Organization struct {
				ID          string `json:"organization_id"`
				DisplayName string `json:"display_name"`
			} `json:"organization"`
			Runtime struct {
				Mode string `json:"mode"`
			} `json:"runtime"`
		} `json:"context"`
	}
	decodeCLIJSON(t, stdout.Bytes(), &created)
	if created.Kind != "PMBrokerContext" || created.Context.Name != "prod" || created.Context.Organization.ID != "org_0123456789abcdef" || created.Context.Runtime.Mode != "remote" {
		t.Fatalf("created context = %#v", created)
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"--json", "context", "use", "prod"}, &stdout, &stderr); code != 0 {
		t.Fatalf("context use exit = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"--json", "context", "show"}, &stdout, &stderr); code != 0 {
		t.Fatalf("context show exit = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	var shown struct {
		Kind    string `json:"kind"`
		Source  string `json:"source"`
		Context struct {
			Name string `json:"name"`
		} `json:"context"`
	}
	decodeCLIJSON(t, stdout.Bytes(), &shown)
	if shown.Kind != "PMBrokerResolvedContext" || shown.Context.Name != "prod" || shown.Source != "active_user" {
		t.Fatalf("shown context = %#v", shown)
	}

	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"--json", "context", "list"}, &stdout, &stderr); code != 0 {
		t.Fatalf("context list exit = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	var listed struct {
		Kind     string `json:"kind"`
		Active   string `json:"active_context"`
		Contexts []struct {
			Name string `json:"name"`
		} `json:"contexts"`
	}
	decodeCLIJSON(t, stdout.Bytes(), &listed)
	if listed.Kind != "PMBrokerContextList" || listed.Active != "prod" || len(listed.Contexts) != 1 || listed.Contexts[0].Name != "prod" {
		t.Fatalf("listed contexts = %#v", listed)
	}
}

func TestPMBrokerContextHybridRequiresPolicy(t *testing.T) {
	isolatePMBrokerUserConfig(t)

	args := append([]string{"--json", "context", "create", "dev"}, pmBrokerContextFlags("development", "hybrid")...)
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	if code != 3 {
		t.Fatalf("context create exit = %d, want validation exit 3\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "validation_error") || !strings.Contains(stdout.String(), "hybrid") {
		t.Fatalf("stdout = %s, want hybrid validation error", stdout.String())
	}
}

func TestPMBrokerMetadataInvalidActionDoesNotReadPoisonedState(t *testing.T) {
	isolatePMBrokerUserConfig(t)
	path, err := pmbroker.DefaultUserStatePath()
	if err != nil {
		t.Fatalf("DefaultUserStatePath: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create user config dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"version":999,"contexts":[]}`), 0o600); err != nil {
		t.Fatalf("write poisoned state: %v", err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"organizations", "bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("organizations bogus exit = %d, want usage exit 2\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stderr.String(), "unsafe") || strings.Contains(stderr.String(), "version") {
		t.Fatalf("stderr = %q, want usage error without state validation details", stderr.String())
	}
}

func TestPMBrokerMetadataCommandsListCachedContextMetadata(t *testing.T) {
	isolatePMBrokerUserConfig(t)

	createArgs := append([]string{"--json", "context", "create", "prod"}, pmBrokerContextFlags("production", "remote")...)
	var stdout, stderr bytes.Buffer
	if code := Run(createArgs, &stdout, &stderr); code != 0 {
		t.Fatalf("context create exit = %d, want 0\nstdout=%s\nstderr=%s", code, stdout.String(), stderr.String())
	}

	checks := []struct {
		args []string
		kind string
		want string
	}{
		{args: []string{"--json", "organizations", "list"}, kind: "OrganizationList", want: "Acme Organization"},
		{args: []string{"--json", "workspaces", "list"}, kind: "WorkspaceList", want: "Analytics Workspace"},
		{args: []string{"--json", "environments", "list"}, kind: "EnvironmentList", want: "Production Environment"},
	}
	for _, check := range checks {
		t.Run(strings.Join(check.args[1:], " "), func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()
			if code := Run(check.args, &stdout, &stderr); code != 0 {
				t.Fatalf("%v exit = %d, want 0\nstdout=%s\nstderr=%s", check.args, code, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), "\"kind\": \""+check.kind+"\"") || !strings.Contains(stdout.String(), check.want) {
				t.Fatalf("stdout = %s, want kind %s and %q", stdout.String(), check.kind, check.want)
			}
		})
	}
}

func TestPMBrokerHelpSurfaces(t *testing.T) {
	isolatePMBrokerUserConfig(t)

	for _, args := range [][]string{
		{"--json", "context"},
		{"--json", "context", "--help"},
		{"--json", "help", "context"},
		{"--json", "organizations"},
		{"--json", "workspaces"},
		{"--json", "environments"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := Run(args, &stdout, &stderr); code != 0 {
				t.Fatalf("%v exit = %d, want 0\nstdout=%s\nstderr=%s", args, code, stdout.String(), stderr.String())
			}
			if !strings.Contains(stdout.String(), "CommandManual") {
				t.Fatalf("stdout = %s, want CommandManual", stdout.String())
			}
		})
	}
}

func TestPMBrokerWebsiteCommandDocsParity(t *testing.T) {
	indexPath := filepath.Join("..", "..", "website", "content", "docs", "index.mdx")
	index, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read website index: %v", err)
	}
	for _, want := range []string{"pm context", "pm organizations", "pm workspaces", "pm environments"} {
		if !strings.Contains(string(index), want) {
			t.Fatalf("%s missing %q", indexPath, want)
		}
	}

	referencePath := filepath.Join("..", "..", "website", "content", "docs", "cli-reference.mdx")
	reference, err := os.ReadFile(referencePath)
	if err != nil {
		t.Fatalf("read website CLI reference: %v", err)
	}
	for _, want := range []string{"`pm context`", "`pm organizations`", "`pm workspaces`", "`pm environments`", "does not enable live provider operations"} {
		if !strings.Contains(string(reference), want) {
			t.Fatalf("%s missing %q", referencePath, want)
		}
	}

	generatedPath := filepath.Join("..", "..", "website", "lib", "docs.generated.ts")
	generated, err := os.ReadFile(generatedPath)
	if err != nil {
		t.Fatalf("read generated website docs data: %v", err)
	}
	for _, want := range []string{"pm context", "pm organizations", "pm workspaces", "pm environments"} {
		if !strings.Contains(string(generated), want) {
			t.Fatalf("%s missing %q", generatedPath, want)
		}
	}
}

func pmBrokerContextFlags(environmentType, runtimeMode string) []string {
	return []string{
		"--organization", "org_0123456789abcdef",
		"--organization-name", "Acme Organization",
		"--workspace", "wks_0123456789abcdef",
		"--workspace-name", "Analytics Workspace",
		"--environment", "env_0123456789abcdef",
		"--environment-name", "Production Environment",
		"--environment-type", environmentType,
		"--broker-profile", "bpf_0123456789abcdef",
		"--broker-profile-name", "Pilot Broker Profile",
		"--runtime-mode", runtimeMode,
	}
}

func isolatePMBrokerUserConfig(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	for _, name := range []string{
		"POLYMETRICS_ROOT", "PM_ROOT",
		"POLYMETRICS_JSON", "PM_JSON",
		"POLYMETRICS_VERSION", "PM_VERSION",
		"POLYMETRICS_PROJECT", "PM_PROJECT",
		"POLYMETRICS_WAREHOUSE_CONNECTOR", "PM_WAREHOUSE_CONNECTOR",
		"POLYMETRICS_WAREHOUSE_PATH", "PM_WAREHOUSE_PATH",
		"POLYMETRICS_POSTGRES_URL", "PM_POSTGRES_URL",
		"POLYMETRICS_DRAGONFLY_ADDR", "PM_DRAGONFLY_ADDR",
		"POLYMETRICS_TEMPORAL_ADDR", "PM_TEMPORAL_ADDR",
		"POLYMETRICS_RLM_IMAGE", "PM_RLM_IMAGE",
		"POLYMETRICS_PODMAN_BIN", "PM_PODMAN_BIN",
		"POLYMETRICS_RLM_FAKE_RUNNER", "PM_RLM_FAKE_RUNNER",
		"POLYMETRICS_RLM_EMBEDDED_WORKER", "PM_RLM_EMBEDDED_WORKER",
		"POLYMETRICS_LLM_PROVIDER", "PM_LLM_PROVIDER",
		"POLYMETRICS_LLM_BASE_URL", "PM_LLM_BASE_URL",
		"POLYMETRICS_LLM_MODEL", "PM_LLM_MODEL",
		"POLYMETRICS_CRONTAB_FILE", "PM_CRONTAB_FILE",
		"POLYMETRICS_BROKER_REQUIRED_CONTEXT", "PM_BROKER_REQUIRED_CONTEXT",
		"POLYMETRICS_BROKER_DEFAULT_CONTEXT", "PM_BROKER_DEFAULT_CONTEXT",
		"POLYMETRICS_BROKER_RUNTIME_MODE", "PM_BROKER_RUNTIME_MODE",
		"POLYMETRICS_BROKER_HYBRID_POLICY", "PM_BROKER_HYBRID_POLICY",
	} {
		t.Setenv(name, "")
	}
}

func decodeCLIJSON(t *testing.T, data []byte, out any) {
	t.Helper()
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("unmarshal CLI JSON: %v\nstdout=%s", err, string(data))
	}
}
