package gsd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalProgrammingLoopIsRenderedAndSurfacedToPi(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(filepath.Join(root, "scripts", "gsd"), "prompt", "programming-loop", "init", "--phase", "issue-372", "--dry-run")
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("programming-loop prompt failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "universal programming loop") {
		t.Fatalf("programming-loop prompt does not contain the repo contract: %s", output)
	}
	raw, err := os.ReadFile(filepath.Join(root, ".gsd", "local-commands.json"))
	if err != nil {
		t.Fatal(err)
	}
	var registry struct {
		Commands []struct {
			Name string `json:"name"`
		} `json:"commands"`
	}
	if err := json.Unmarshal(raw, &registry); err != nil {
		t.Fatal(err)
	}
	if len(registry.Commands) != 1 || registry.Commands[0].Name != "programming-loop" {
		t.Fatalf("unexpected local command registry: %+v", registry.Commands)
	}
	extension, err := os.ReadFile(filepath.Join(root, ".pi", "extensions", "gsd", "index.ts"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(extension), "local-commands.json") {
		t.Fatal("Pi extension does not merge the local command registry")
	}
}
