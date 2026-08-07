package durable

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSyncDirectory(t *testing.T) {
	if err := SyncDirectory(t.TempDir()); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureDirectoryTreeCreatesMissingAncestors(t *testing.T) {
	root := filepath.Join(t.TempDir(), "project")
	directory := filepath.Join(root, "state", "receipts")
	if err := EnsureDirectoryTree(directory, root, 0o700); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(directory)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", directory)
	}
}

func TestEnsureDirectoryTreeSyncsCompleteExistingAncestorChain(t *testing.T) {
	root := filepath.Join(t.TempDir(), "project")
	directory := filepath.Join(root, "state", "receipts")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	synced := map[string]bool{}
	if err := ensureDirectoryTree(directory, root, 0o700, ensureDirectory, func(path string) error {
		synced[path] = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	absoluteDirectory, err := filepath.Abs(directory)
	if err != nil {
		t.Fatal(err)
	}
	for current := absoluteDirectory; ; current = filepath.Dir(current) {
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		if !synced[current] {
			t.Fatalf("directory %s was not synced", current)
		}
	}
}

func TestEnsureDirectoryTreeRejectsPathOutsideRoot(t *testing.T) {
	root := t.TempDir()
	if err := EnsureDirectoryTree(filepath.Join(t.TempDir(), "state"), root, 0o700); err == nil {
		t.Fatal("EnsureDirectoryTree() succeeded for an outside path")
	}
}
