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
