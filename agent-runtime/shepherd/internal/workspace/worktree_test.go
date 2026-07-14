package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAttemptWorktreePromoteAndDiscard(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	head := gitOutput(t, repo, "rev-parse", "HEAD")
	manager, err := NewManager(repo, filepath.Join(t.TempDir(), "attempts"))
	if err != nil {
		t.Fatal(err)
	}
	attempt, err := manager.Create(ctx, AttemptIdentity{DeliveryID: "issue-389", Generation: 1, UnitID: "execute-task/M001/S01/T01", Attempt: 1, BaseHead: head})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(attempt.Root, "agent-runtime", "shepherd"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(attempt.Root, "agent-runtime", "shepherd", "proof.txt"), []byte("ok\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	promoted, err := manager.Promote(ctx, attempt, []string{"agent-runtime/shepherd/**"}, "test: promote attempt")
	if err != nil {
		t.Fatal(err)
	}
	if promoted == head {
		t.Fatal("promotion did not advance head")
	}
	if _, err := os.Stat(filepath.Join(repo, "agent-runtime", "shepherd", "proof.txt")); err != nil {
		t.Fatalf("promoted file missing from canonical worktree: %v", err)
	}
	if err := manager.Discard(ctx, attempt); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(attempt.Root); !os.IsNotExist(err) {
		t.Fatalf("attempt root still exists or unexpected error: %v", err)
	}
}

func TestAttemptWorktreeRejectsStaleHeadAndOutOfScopePromotion(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	head := gitOutput(t, repo, "rev-parse", "HEAD")
	manager, err := NewManager(repo, filepath.Join(t.TempDir(), "attempts"))
	if err != nil {
		t.Fatal(err)
	}
	attempt, err := manager.Create(ctx, AttemptIdentity{DeliveryID: "issue-389", Generation: 1, UnitID: "execute-task/M001/S01/T01", Attempt: 1, BaseHead: head})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(attempt.Root, "outside.txt"), []byte("bad\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Promote(ctx, attempt, []string{"agent-runtime/shepherd/**"}, "test: should fail"); err == nil {
		t.Fatal("out-of-scope attempt promotion succeeded")
	}
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitNoOutput(t, repo, "add", "README.md")
	gitNoOutput(t, repo, "commit", "-m", "advance canonical")
	if err := os.Remove(filepath.Join(attempt.Root, "outside.txt")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(attempt.Root, "agent-runtime", "shepherd"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(attempt.Root, "agent-runtime", "shepherd", "proof.txt"), []byte("ok\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Promote(ctx, attempt, []string{"agent-runtime/shepherd/**"}, "test: stale"); err == nil || !strings.Contains(err.Error(), "canonical head changed") {
		t.Fatalf("stale head promotion err=%v", err)
	}
}

func initRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	gitNoOutput(t, repo, "init", "-b", "main")
	gitNoOutput(t, repo, "config", "user.email", "test@example.com")
	gitNoOutput(t, repo, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("initial\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitNoOutput(t, repo, "add", "README.md")
	gitNoOutput(t, repo, "commit", "-m", "initial")
	return repo
}

func gitOutput(t *testing.T, root string, args ...string) string {
	t.Helper()
	raw, err := git(context.Background(), root, args...)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(raw))
}

func gitNoOutput(t *testing.T, root string, args ...string) {
	t.Helper()
	if _, err := git(context.Background(), root, args...); err != nil {
		t.Fatal(err)
	}
}
