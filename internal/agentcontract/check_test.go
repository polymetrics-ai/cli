package agentcontract

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestReferencedGSDCommandsResolve(t *testing.T) {
	root := repositoryRoot(t)
	contract := loadRepositoryContract(t, root)

	if err := CheckGSDCommands(context.Background(), root, contract.GSD.Commands); err != nil {
		t.Fatalf("referenced GSD commands must resolve: %v", err)
	}
}

func TestCheckGSDCommandsRunsFromSelectedRoot(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "scripts", "gsd")
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".selected-root"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	contents := []byte("#!/bin/sh\n[ -f .selected-root ]\n")
	if err := os.WriteFile(script, contents, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := CheckGSDCommands(context.Background(), root, []string{"discuss-phase"}); err != nil {
		t.Fatalf("GSD command did not run from selected root: %v", err)
	}
}

func TestCheckProjectionRejectsDivergence(t *testing.T) {
	want := []byte("canonical generated block\n")
	got := []byte("diverged generated block\n")

	if err := CheckProjection(want, got); err == nil {
		t.Fatal("CheckProjection accepted a diverged projection")
	}
}

func TestProjectionDriftCheckAndSync(t *testing.T) {
	repository := repositoryRoot(t)
	contract := loadRepositoryContract(t, repository)

	root := t.TempDir()
	writeCodexProjections(t, root, contract)
	path := filepath.Join(root, filepath.FromSlash(contract.Projections[0].Path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	want, err := RenderBlock(contract, contract.BaseRole.Name)
	if err != nil {
		t.Fatal(err)
	}
	wrapper := append([]byte("---\nname: pm-delivery-worker\n---\n\n"), want...)
	wrapper = append(wrapper, []byte("\nharness-owned footer\n")...)
	if err := os.WriteFile(path, wrapper, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("matching projection failed: %v", err)
	}

	diverged := strings.Replace(string(wrapper), "Receive one assigned job", "Receive many jobs", 1)
	if err := os.WriteFile(path, []byte(diverged), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections accepted a diverged registered projection")
	}

	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatalf("SyncProjections: %v", err)
	}
	if updated != 1 {
		t.Fatalf("SyncProjections updated %d files, want 1", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("projection did not pass after sync: %v", err)
	}
}

func TestOptionalNonCodexProjectionsMayBeAbsent(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections accepted a root without the required Codex projections")
	}
	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 2 {
		t.Fatalf("SyncProjections updated %d files, want the two required Codex projections", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("optional non-Codex projections should remain absent: %v", err)
	}
}

func TestCodexWorkersCannotDelegateToAmbientAgents(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))

	for _, target := range contract.Projections {
		if target.Harness != "codex" {
			continue
		}

		rendered, err := RenderProjection(contract, target)
		if err != nil {
			t.Fatal(err)
		}
		configuration := parseCodexProjection(t, rendered)
		for _, field := range contract.Codex.RequiredFields {
			if !configuration.IsSet(field) || strings.TrimSpace(configuration.GetString(field)) == "" {
				t.Fatalf("%s is missing required Codex field %q", target.Role, field)
			}
		}
		for _, ambientAgent := range []string{"worker", "ambient-user-role"} {
			if codexCanDelegateToAmbientAgent(configuration, ambientAgent) {
				t.Fatalf("%s can delegate to ambient agent %q because agents.enabled is not false", target.Role, ambientAgent)
			}
		}
	}
}

// Codex documents agents.enabled as true by default and false as disabling multi-agent tools.
// This models whether an otherwise reachable built-in or user-defined agent can be delegated to;
// it intentionally does not invoke a live model during a unit test.
func codexCanDelegateToAmbientAgent(configuration *viper.Viper, ambientAgent string) bool {
	return ambientAgent != "" && (!configuration.IsSet("agents.enabled") || configuration.GetBool("agents.enabled"))
}

func TestCodexProjectionDriftRejectsDelegationRegression(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	writeCodexProjections(t, root, contract)

	var target ProjectionTarget
	for _, candidate := range contract.Projections {
		if candidate.Harness == "codex" {
			target = candidate
			break
		}
	}
	if target.Path == "" {
		t.Fatal("canonical contract does not register a Codex projection")
	}
	path := filepath.Join(root, filepath.FromSlash(target.Path))
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, bytes.Replace(content, []byte("agents.enabled = false"), []byte("agents.enabled = true"), 1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections accepted a Codex worker with ambient delegation enabled")
	}

	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 1 {
		t.Fatalf("SyncProjections updated %d files, want 1", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("Codex worker did not pass after sync: %v", err)
	}
}

func parseCodexProjection(t *testing.T, content []byte) *viper.Viper {
	t.Helper()
	configuration := viper.New()
	configuration.SetConfigType("toml")
	if err := configuration.ReadConfig(bytes.NewReader(content)); err != nil {
		t.Fatalf("generated Codex projection is not valid TOML: %v", err)
	}
	return configuration
}

func writeCodexProjections(t *testing.T, root string, contract *Contract) {
	t.Helper()
	for _, target := range contract.Projections {
		if target.Harness != "codex" {
			continue
		}
		content, err := RenderProjection(contract, target)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, filepath.FromSlash(target.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestProjectionIORejectsSymlinkEscape(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))

	root := t.TempDir()
	outside := t.TempDir()
	original := make(map[string][]byte)
	for _, target := range contract.Projections {
		if target.Harness != "codex" {
			continue
		}
		content, err := RenderProjection(contract, target)
		if err != nil {
			t.Fatal(err)
		}
		escapedPath := filepath.Join(outside, "agents", filepath.Base(target.Path))
		if err := os.MkdirAll(filepath.Dir(escapedPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(escapedPath, content, 0o644); err != nil {
			t.Fatal(err)
		}
		original[escapedPath] = content
	}
	if err := os.Symlink(outside, filepath.Join(root, ".codex")); err != nil {
		t.Skipf("cannot create projection ancestor symlink: %v", err)
	}

	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections followed a projection ancestor outside the selected root")
	}
	if _, err := SyncProjections(root, contract); err == nil {
		t.Fatal("SyncProjections followed a projection ancestor outside the selected root")
	}
	for escapedPath, content := range original {
		after, err := os.ReadFile(escapedPath)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(after, content) {
			t.Fatal("SyncProjections modified a projection outside the selected root")
		}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func loadRepositoryContract(t *testing.T, root string) *Contract {
	t.Helper()
	contract, err := Load(filepath.Join(root, SourcePath))
	if err != nil {
		t.Fatal(err)
	}
	return contract
}
