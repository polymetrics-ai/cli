package safety

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalWriteFSPreservesInRootPathsAndModes(t *testing.T) {
	root := t.TempDir()
	actualDir := filepath.Join(root, "actual")
	if err := os.Mkdir(actualDir, 0o700); err != nil {
		t.Fatalf("create actual directory: %v", err)
	}
	if err := os.Symlink("actual", filepath.Join(root, "alias")); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}
	fs, err := OpenLocalWriteFS(root, false)
	if err != nil {
		t.Fatalf("OpenLocalWriteFS() error = %v", err)
	}
	defer fs.Close()

	dir := filepath.Join(root, "new", "nested")
	if err := fs.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	file, err := fs.OpenFile(filepath.Join(dir, "created.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		t.Fatalf("OpenFile(nonexisting) error = %v", err)
	}
	if _, err := file.Write([]byte("synthetic\n")); err != nil {
		t.Fatalf("write nonexisting file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close nonexisting file: %v", err)
	}
	linked, err := fs.OpenFile(filepath.Join(root, "alias", "linked.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		t.Fatalf("OpenFile(in-root symlink) error = %v", err)
	}
	if err := linked.Close(); err != nil {
		t.Fatalf("close in-root symlink file: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "created.jsonl"))
	if err != nil {
		t.Fatalf("stat created file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("created file mode = %o, want 600", got)
	}
	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat created directory: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("created directory mode = %o, want 700", got)
	}
}

func TestLocalWriteFSRenameReplacesFinalSymlinkWithoutFollowingIt(t *testing.T) {
	root := t.TempDir()
	external := filepath.Join(t.TempDir(), "outside.jsonl")
	before := []byte("outside-sentinel\n")
	if err := os.WriteFile(external, before, 0o600); err != nil {
		t.Fatalf("create external target: %v", err)
	}
	finalPath := filepath.Join(root, "final.jsonl")
	if err := os.Symlink(external, finalPath); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}
	tmpPath := filepath.Join(root, "final.tmp")
	if err := os.WriteFile(tmpPath, []byte("replacement\n"), 0o600); err != nil {
		t.Fatalf("create temporary file: %v", err)
	}
	fs, err := OpenLocalWriteFS(root, false)
	if err != nil {
		t.Fatalf("OpenLocalWriteFS() error = %v", err)
	}
	defer fs.Close()
	if err := fs.Rename(tmpPath, finalPath); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	after, err := os.ReadFile(external)
	if err != nil {
		t.Fatalf("read external target: %v", err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("rooted rename changed the external target")
	}
	info, err := os.Lstat(finalPath)
	if err != nil {
		t.Fatalf("lstat replacement: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatal("rooted rename left the final path as a symlink")
	}
}

func TestLocalWriteFSExplicitExternalPolicyPreservesOrdinaryEffects(t *testing.T) {
	root := t.TempDir()
	external := filepath.Join(t.TempDir(), "outside.jsonl")
	finalPath := filepath.Join(root, "final.jsonl")
	if err := os.Symlink(external, finalPath); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}
	fs, err := OpenLocalWriteFS(root, true)
	if err != nil {
		t.Fatalf("OpenLocalWriteFS() error = %v", err)
	}
	defer fs.Close()
	file, err := fs.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		t.Fatalf("OpenFile(explicit external) error = %v", err)
	}
	if _, err := file.Write([]byte("synthetic\n")); err != nil {
		t.Fatalf("write explicit external target: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close explicit external target: %v", err)
	}
	if _, err := os.Stat(external); err != nil {
		t.Fatalf("explicit external target missing: %v", err)
	}
}
