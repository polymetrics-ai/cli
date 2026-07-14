package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttemptWorktreeRejectsSymlinkedAncestor(t *testing.T) {
	repo := initRepo(t)
	root := filepath.Join(t.TempDir(), "attempts")
	manager, err := NewManager(repo, root)
	if err != nil {
		t.Fatal(err)
	}
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "issue-389"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "issue-389", "g1")); err != nil {
		t.Fatal(err)
	}
	head := gitOutput(t, repo, "rev-parse", "HEAD")
	_, err = manager.Create(context.Background(), AttemptIdentity{DeliveryID: "issue-389", Generation: 1, UnitID: "execute-task/M001/S01/T01", Attempt: 1, BaseHead: head})
	if err == nil || !strings.Contains(err.Error(), "escapes") && !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("error=%v, want symlink/escape rejection", err)
	}
}
