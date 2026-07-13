package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRequireCleanDetectsUntrackedWork(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	commands := [][]string{{"init", "-q"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "Test"}}
	for _, args := range commands {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	file := filepath.Join(root, "tracked.txt")
	if err := os.WriteFile(file, []byte("tracked"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "tracked.txt"}, {"commit", "-qm", "seed"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	snapshot, err := Inspect(context.Background(), root)
	if err != nil || snapshot.Dirty {
		t.Fatalf("clean snapshot=%+v err=%v", snapshot, err)
	}
	if err := os.WriteFile(filepath.Join(root, "untracked.txt"), []byte("dirty"), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err = Inspect(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if err := RequireClean(snapshot); err == nil {
		t.Fatal("expected dirty worktree to block")
	}
}
