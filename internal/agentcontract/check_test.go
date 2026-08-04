package agentcontract

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReferencedGSDCommandsResolve(t *testing.T) {
	root := repositoryRoot(t)
	contract := loadRepositoryContract(t, root)

	if err := CheckGSDCommands(context.Background(), filepath.Join(root, "scripts", "gsd"), contract.GSD.Commands); err != nil {
		t.Fatalf("referenced GSD commands must resolve: %v", err)
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
	contract.Projections[0].Required = true

	root := t.TempDir()
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

func TestOptionalWaveProjectionMayBeAbsent(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	if err := CheckProjections(t.TempDir(), contract); err != nil {
		t.Fatalf("optional Wave 2-4 projections should be absent in Wave 1: %v", err)
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
