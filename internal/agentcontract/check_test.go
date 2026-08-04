package agentcontract

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestProjectionIORejectsSymlinkEscape(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	contract.Projections[0].Required = true
	target := contract.Projections[0]

	root := t.TempDir()
	outside := t.TempDir()
	escapedPath := filepath.Join(outside, "agents", filepath.Base(target.Path))
	if err := os.MkdirAll(filepath.Dir(escapedPath), 0o755); err != nil {
		t.Fatal(err)
	}
	block, err := RenderBlock(contract, target.Role)
	if err != nil {
		t.Fatal(err)
	}
	original := append([]byte("---\nname: escaped\n---\n\n"), block...)
	original = append(original, []byte("\noutside footer\n")...)
	original = []byte(strings.Replace(string(original), "Receive one assigned job", "Receive escaped work", 1))
	if err := os.WriteFile(escapedPath, original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, ".claude")); err != nil {
		t.Skipf("cannot create projection ancestor symlink: %v", err)
	}

	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections followed a projection ancestor outside the selected root")
	}
	if _, err := SyncProjections(root, contract); err == nil {
		t.Fatal("SyncProjections followed a projection ancestor outside the selected root")
	}
	after, err := os.ReadFile(escapedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, original) {
		t.Fatal("SyncProjections modified a projection outside the selected root")
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
