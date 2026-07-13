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
	config := ContainerConfig{Engine: "podman", Image: "localhost/gsd-pi:1.11.0", GSDStateDir: filepath.Join(t.TempDir(), "gsd"), PlanningDir: filepath.Join(t.TempDir(), "planning"), AuthFile: auth, SettingsFile: settings}
	if err := config.Validate(root); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(config.commandArgs(root, []string{"headless", "query"}), " ")
	for _, required := range []string{"/workspace/.gsd", "/workspace/.planning", "auth.json:ro", "settings.json:ro", "--pull=never"} {
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
