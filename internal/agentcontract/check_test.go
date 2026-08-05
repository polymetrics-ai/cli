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

	root := t.TempDir()
	if updated, err := SyncProjections(root, contract); err != nil || updated != 2 {
		t.Fatalf("create required Pi projections before block test: updated=%d err=%v", updated, err)
	}
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

func TestOptionalNonPiWaveProjectionMayBeAbsent(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	if updated, err := SyncProjections(root, contract); err != nil || updated != 2 {
		t.Fatalf("create required Pi projections: updated=%d err=%v", updated, err)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("optional non-Pi projections should be absent: %v", err)
	}
}

func TestPiProjectionsAreRequired(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))

	for _, target := range contract.Projections {
		if target.Harness != "pi" {
			continue
		}
		if !target.Required {
			t.Fatalf("Pi projection %s must be required once its wave owns the generated file", target.Path)
		}
		if target.RenderMode != "full" {
			t.Fatalf("Pi projection %s must be a complete generated file", target.Path)
		}
	}
}

func TestSyncCreatesRequiredPiProjections(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatalf("SyncProjections must create required Pi projections: %v", err)
	}
	if updated != 2 {
		t.Fatalf("SyncProjections created %d Pi projections, want 2", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("created Pi projections must pass drift check: %v", err)
	}
	for _, target := range contract.Projections {
		if target.Harness != "pi" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(target.Path)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.HasPrefix(content, []byte("---\nname: ")) || !bytes.Contains(content, []byte("tools:\n")) {
			t.Fatalf("Pi projection %s is not a complete generated Pi agent file", target.Path)
		}
	}
}

func TestPiProjectionRejectsWholeFileDrift(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	if updated, err := SyncProjections(root, contract); err != nil || updated != 2 {
		t.Fatalf("create required Pi projections: updated=%d err=%v", updated, err)
	}

	var target ProjectionTarget
	for _, candidate := range contract.Projections {
		if candidate.Harness == "pi" && candidate.Role == contract.BaseRole.Name {
			target = candidate
			break
		}
	}
	if target.Path == "" {
		t.Fatal("missing delivery-worker Pi projection")
	}
	path := filepath.Join(root, filepath.FromSlash(target.Path))
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(content, []byte("\nhand-written footer\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CheckProjections(root, contract); err == nil {
		t.Fatal("CheckProjections accepted hand-written content in a complete Pi projection")
	}
	updated, err := SyncProjections(root, contract)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 1 {
		t.Fatalf("SyncProjections updated %d drifted Pi projection(s), want 1", updated)
	}
	if err := CheckProjections(root, contract); err != nil {
		t.Fatalf("Pi projection did not pass after sync: %v", err)
	}
}

func TestSyncPiProjectionRejectsSymlinkParent(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, ".pi")); err != nil {
		t.Skipf("cannot create projection ancestor symlink: %v", err)
	}

	if _, err := SyncProjections(root, contract); err == nil {
		t.Fatal("SyncProjections followed a required Pi projection ancestor outside the selected root")
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("SyncProjections created Pi projection content outside the selected root: %#v", entries)
	}
}

func TestProjectionIORejectsSymlinkEscape(t *testing.T) {
	contract := loadRepositoryContract(t, repositoryRoot(t))
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
