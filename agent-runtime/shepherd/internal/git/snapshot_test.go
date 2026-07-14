package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	managedWorktree := filepath.Join(root, ".gsd-worktrees", "M001-test")
	if err := os.MkdirAll(managedWorktree, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(managedWorktree, "managed.txt"), []byte("runtime state"), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err = Inspect(context.Background(), root)
	if err != nil || snapshot.Dirty {
		t.Fatalf("managed GSD worktree dirtied parent snapshot=%+v err=%v", snapshot, err)
	}
	gsdRuntime := filepath.Join(root, ".gsd", "PREFERENCES.md")
	if err := os.MkdirAll(filepath.Dir(gsdRuntime), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gsdRuntime, []byte("local runtime state"), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err = Inspect(context.Background(), root)
	if err != nil || snapshot.Dirty {
		t.Fatalf("untracked GSD runtime state dirtied parent snapshot=%+v err=%v", snapshot, err)
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

func TestInspectRejectsTrackedGSDPolicyChanges(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "Test"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	policy := filepath.Join(root, ".gsd", "PREFERENCES.md")
	if err := os.MkdirAll(filepath.Dir(policy), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(policy, []byte("governed"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", ".gsd/PREFERENCES.md"}, {"commit", "-qm", "seed"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	if err := os.WriteFile(policy, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := Inspect(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if err := RequireClean(snapshot); err == nil {
		t.Fatal("expected tracked GSD policy change to block")
	}
}

func TestRestoreIndexPreservesWorkingTreeChanges(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "Test"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	tracked := filepath.Join(root, "tracked.txt")
	if err := os.WriteFile(tracked, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "tracked.txt"}, {"commit", "-qm", "seed"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	if err := os.WriteFile(tracked, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "-C", root, "add", "tracked.txt")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("stage change: %v: %s", err, output)
	}
	if err := RestoreIndex(context.Background(), root); err != nil {
		t.Fatal(err)
	}
	staged := exec.Command("git", "-C", root, "diff", "--cached", "--quiet")
	if err := staged.Run(); err != nil {
		t.Fatalf("index still differs from HEAD: %v", err)
	}
	working := exec.Command("git", "-C", root, "diff", "--quiet")
	if err := working.Run(); err == nil {
		t.Fatal("working-tree change was lost")
	}
}

func TestCheckpointWithinScopesCommitsOnlyAllowedChanges(t *testing.T) {
	t.Parallel()
	root := initializedRepository(t)
	allowed := filepath.Join(root, "internal", "connectors", "engine", "bundle.go")
	if err := os.MkdirAll(filepath.Dir(allowed), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(allowed, []byte("package engine\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	head, err := CheckpointWithinScopes(context.Background(), root, []string{"internal/connectors/engine/**"}, "chore(gsd): complete M001 S01 T01")
	if err != nil {
		t.Fatal(err)
	}
	if len(head) != 40 {
		t.Fatalf("head=%q", head)
	}
	snapshot, err := Inspect(context.Background(), root)
	if err != nil || snapshot.Dirty {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}
	if second, err := CheckpointWithinScopes(context.Background(), root, []string{"internal/connectors/engine/**"}, "chore(gsd): complete M001 S01 T01"); err != nil || second != head {
		t.Fatalf("idempotent head=%q err=%v", second, err)
	}
}

func TestCheckpointWithinScopesRejectsOutOfScopeChangesWithoutStaging(t *testing.T) {
	t.Parallel()
	root := initializedRepository(t)
	if err := os.WriteFile(filepath.Join(root, "outside.txt"), []byte("no"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckpointWithinScopes(context.Background(), root, []string{"internal/connectors/engine/**"}, "chore(gsd): complete M001 S01 T01"); err == nil {
		t.Fatal("out-of-scope change accepted")
	}
	staged := exec.Command("git", "-C", root, "diff", "--cached", "--quiet")
	if err := staged.Run(); err != nil {
		t.Fatal("out-of-scope change was staged")
	}
}

func TestCheckpointWithinScopesIncludesOnlyMutableTrackedGSDProjections(t *testing.T) {
	t.Parallel()
	root := initializedRepository(t)
	if err := os.MkdirAll(filepath.Join(root, ".gsd", "exec"), 0o700); err != nil {
		t.Fatal(err)
	}
	requirements := filepath.Join(root, ".gsd", "REQUIREMENTS.md")
	preferences := filepath.Join(root, ".gsd", "PREFERENCES.md")
	if err := os.WriteFile(requirements, []byte("active\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(preferences, []byte("governed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	mustGit(t, root, "add", ".gsd/REQUIREMENTS.md", ".gsd/PREFERENCES.md")
	mustGit(t, root, "commit", "-qm", "seed gsd")
	if err := os.WriteFile(requirements, []byte("validated\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".gsd", "exec", "runtime.json"), []byte("runtime"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckpointWithinScopes(context.Background(), root, []string{"internal/connectors/engine/**"}, "chore(gsd): checkpoint M001 S01"); err != nil {
		t.Fatal(err)
	}
	if snapshot, err := Inspect(context.Background(), root); err != nil || snapshot.Dirty {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}
	if err := os.WriteFile(preferences, []byte("weakened\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckpointWithinScopes(context.Background(), root, []string{"internal/connectors/engine/**"}, "chore(gsd): checkpoint M001 S01"); err == nil {
		t.Fatal("tracked GSD policy mutation was checkpointed")
	}
}

func initializedRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "Test"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "seed.txt"), []byte("seed"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "seed.txt"}, {"commit", "-qm", "seed"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	return root
}

func TestFinalizeInitializingMilestoneWorktreeRestoresIndexWithoutReset(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "test@example.invalid"}, {"config", "user.name", "Test"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "tracked.txt"), []byte("tracked"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "tracked.txt"}, {"commit", "-qm", "seed"}, {"branch", "-M", "integration"}} {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, output)
		}
	}
	child := filepath.Join(root, ".gsd-worktrees", "M001-test")
	cmd := exec.Command("git", "-C", root, "worktree", "add", "-qb", "milestone/M001-test", child, "HEAD")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("add worktree: %v: %s", err, output)
	}
	index := gitPath(t, child, "index")
	locked := gitPath(t, child, "locked")
	if err := os.Remove(index); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(index+".lock", nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(locked, []byte("initializing\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	head := strings.TrimSpace(string(mustGit(t, root, "rev-parse", "HEAD")))
	if err := FinalizeInitializingMilestoneWorktree(context.Background(), root, "M001-test", head); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(index); err != nil {
		t.Fatalf("index was not restored: %v", err)
	}
	if _, err := os.Stat(index + ".lock"); !os.IsNotExist(err) {
		t.Fatalf("stale index lock remains: %v", err)
	}
	status := mustGit(t, child, "status", "--porcelain")
	if len(status) != 0 {
		t.Fatalf("repaired worktree is dirty: %q", status)
	}
}

func gitPath(t *testing.T, root, name string) string {
	t.Helper()
	return strings.TrimSpace(string(mustGit(t, root, "rev-parse", "--git-path", name)))
}

func mustGit(t *testing.T, root string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
	return output
}
