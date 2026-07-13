package gsd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContainerRuntimeHidesHostPlanningAndCredentialSurface(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	auth := filepath.Join(t.TempDir(), "auth.json")
	settings := filepath.Join(t.TempDir(), "settings.json")
	for _, path := range []string{auth, settings} {
		if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	policyDir := filepath.Join(t.TempDir(), ".gsd")
	if err := os.MkdirAll(policyDir, 0o700); err != nil {
		t.Fatal(err)
	}
	gitCommonDir := filepath.Join(t.TempDir(), ".git")
	if err := os.MkdirAll(gitCommonDir, 0o700); err != nil {
		t.Fatal(err)
	}
	config := ContainerConfig{Engine: "podman", Image: "localhost/gsd-pi:1.11.0", GSDStateDir: filepath.Join(t.TempDir(), "gsd"), PlanningDir: filepath.Join(t.TempDir(), "planning"), SessionsDir: filepath.Join(t.TempDir(), "sessions"), BackgroundDir: filepath.Join(t.TempDir(), "bg-shell"), AuthFile: auth, SettingsFile: settings, PolicyDir: policyDir, GitCommonDir: gitCommonDir}
	if err := config.Validate(root); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(config.commandArgs(root, []string{"headless", "query"}), " ")
	for _, required := range []string{root + ":" + root + ":rw", root + "/.gsd", root + "/.planning", root + "/.bg-shell", "/home/shepherd/.pi/agent/sessions:rw", gitCommonDir + ":" + gitCommonDir + ":rw", "GIT_CONFIG_KEY_2=safe.directory", "GIT_CONFIG_VALUE_2=" + root, "auth.json:ro", "settings.json:ro", "--pull=never"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %q in %s", required, joined)
		}
	}
	if strings.Contains(joined, "SSH_AUTH_SOCK") || strings.Contains(joined, "GH_TOKEN") {
		t.Fatal("container inherited publisher credentials")
	}
}

func TestContainerImageReusesBaseNonRootIdentity(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("..", "..", "container", "Containerfile"))
	if err != nil {
		t.Fatal(err)
	}
	containerfile := string(raw)
	if strings.Contains(containerfile, "useradd --create-home --uid 1000") {
		t.Fatal("Node base image already reserves UID 1000; creating another account makes the image unbuildable")
	}
	for _, required := range []string{
		"groupmod --new-name shepherd node",
		"usermod --login shepherd --home /home/shepherd --move-home node",
		"USER shepherd",
	} {
		if !strings.Contains(containerfile, required) {
			t.Fatalf("container image does not establish governed non-root identity: missing %q", required)
		}
	}
}

func TestContainerImageInstallsRequiredGSDRuntimePackages(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("..", "..", "container", "Containerfile"))
	if err != nil {
		t.Fatal(err)
	}
	containerfile := string(raw)
	for _, required := range []string{
		"apt-get update",
		"apt-get install --yes --no-install-recommends",
		"ca-certificates",
		"git",
		"rm -rf /var/lib/apt/lists/*",
	} {
		if !strings.Contains(containerfile, required) {
			t.Fatalf("container image is missing required GSD runtime setup %q", required)
		}
	}
}

func TestContainerImageInstallsGovernedAgentToolchain(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("..", "..", "container", "Containerfile"))
	if err != nil {
		t.Fatal(err)
	}
	containerfile := string(raw)
	for _, required := range []string{
		"golang:1.25.12-bookworm", "make", "jq", "ripgrep",
		"agent-browser@0.31.1", "agent-browser install --with-deps",
		"dpkg --print-architecture", "apt-get install --yes --no-install-recommends chromium",
		"@upstash/context7-mcp@3.2.3", "COPY agent-runtime/shepherd/container/web-fetch",
		"ln -sf /usr/local/bin/web-fetch /usr/local/bin/curl",
		"ln -s /usr/local/go/bin/go /usr/local/bin/go",
	} {
		if !strings.Contains(containerfile, required) {
			t.Errorf("container image is missing governed agent tool %q", required)
		}
	}
	for _, forbidden := range []string{" github-cli", " gh ", "apt-get install curl"} {
		if strings.Contains(containerfile, forbidden) {
			t.Errorf("container image includes forbidden publisher/unrestricted fetch surface %q", forbidden)
		}
	}
	policy, err := os.ReadFile(filepath.Join("..", "..", "container", "agent-browser-policy.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(policy), `"close"`) || strings.Contains(string(policy), `"eval"`) {
		t.Fatal("agent-browser policy must permit cleanup while denying script evaluation")
	}
}

func TestContainerUsesValidatedResearchNetwork(t *testing.T) {
	t.Parallel()
	config, root := validContainerConfig(t)
	config.Network = "shepherd-research"
	if err := config.Validate(root); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(config.commandArgs(root, nil), " ")
	if !strings.Contains(joined, "--network=shepherd-research") {
		t.Fatalf("research network was not applied: %s", joined)
	}
	config.Network = "--privileged"
	if err := config.Validate(root); err == nil {
		t.Fatal("option-like container network must be rejected")
	}
}

func TestProvisionContainerPolicyWritesTrustedContext7MCP(t *testing.T) {
	t.Parallel()
	workDir := t.TempDir()
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(workDir, ".gsd", "agents"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workDir, ".gsd", "PREFERENCES.md"), []byte("policy"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Worker-controlled MCP configuration must never be copied into the protected runtime state.
	if err := os.WriteFile(filepath.Join(workDir, ".gsd", "mcp.json"), []byte(`{"servers":{"evil":{"url":"http://worker.invalid"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := provisionContainerPolicy(filepath.Join(workDir, ".gsd"), stateDir); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(stateDir, "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, "https://mcp.context7.com/mcp") || strings.Contains(text, "worker.invalid") {
		t.Fatalf("unexpected trusted MCP policy: %s", text)
	}
}

func TestContainerPolicyIsProvisionedFromOperatorDirectoryOutsideWorkerTree(t *testing.T) {
	t.Parallel()
	workDir := t.TempDir()
	policyDir := filepath.Join(t.TempDir(), ".gsd")
	stateDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(policyDir, "agents"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(policyDir, "PREFERENCES.md"), []byte("operator policy"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(policyDir, "agents", "reviewer.md"), []byte("operator reviewer"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := provisionContainerPolicy(policyDir, stateDir); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{
		filepath.Join(stateDir, "PREFERENCES.md"):        "operator policy",
		filepath.Join(stateDir, "agents", "reviewer.md"): "operator reviewer",
	} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != want {
			t.Fatalf("%s=%q, want %q", path, raw, want)
		}
	}
	config, _ := validContainerConfig(t)
	config.PolicyDir = filepath.Join(workDir, ".gsd")
	if err := config.Validate(workDir); err == nil {
		t.Fatal("worker-controlled policy directory must be rejected")
	}
}

func TestResearchSidecarIsPrivatePinnedAndJSONEnabled(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..", "research")
	compose, err := os.ReadFile(filepath.Join(root, "compose.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	settings, err := os.ReadFile(filepath.Join(root, "settings.yml"))
	if err != nil {
		t.Fatal(err)
	}
	composeText := string(compose)
	if !strings.Contains(composeText, "searxng/searxng:2026.7.10-799086874") || strings.Contains(composeText, "ports:") {
		t.Fatal("SearXNG must be pinned and must not publish a host port")
	}
	if !strings.Contains(composeText, "secrets.token_urlsafe") || strings.Contains(composeText, "SEARXNG_SECRET=") && !strings.Contains(composeText, "SEARXNG_SECRET=\"$$(python") {
		t.Fatal("SearXNG secret must be generated inside the container at runtime")
	}
	if !strings.Contains(string(settings), "formats:") || !strings.Contains(string(settings), "json") {
		t.Fatal("SearXNG JSON output is not enabled")
	}
}

func validContainerConfig(t *testing.T) (ContainerConfig, string) {
	t.Helper()
	root := t.TempDir()
	auth := filepath.Join(t.TempDir(), "auth.json")
	settings := filepath.Join(t.TempDir(), "settings.json")
	for _, path := range []string{auth, settings} {
		if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	policyDir := filepath.Join(t.TempDir(), ".gsd")
	if err := os.MkdirAll(filepath.Join(policyDir, "agents"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(policyDir, "PREFERENCES.md"), []byte("policy"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitCommonDir := filepath.Join(t.TempDir(), ".git")
	if err := os.MkdirAll(gitCommonDir, 0o700); err != nil {
		t.Fatal(err)
	}
	return ContainerConfig{Engine: "podman", Image: "localhost/gsd-pi:1.11.0", GSDStateDir: filepath.Join(t.TempDir(), "gsd"), PlanningDir: filepath.Join(t.TempDir(), "planning"), SessionsDir: filepath.Join(t.TempDir(), "sessions"), BackgroundDir: filepath.Join(t.TempDir(), "bg-shell"), AuthFile: auth, SettingsFile: settings, PolicyDir: policyDir, GitCommonDir: gitCommonDir}, root
}
