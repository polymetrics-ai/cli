package workspace

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReconcileOwnedAttemptRemovesOnlyExactOwnedWorktreeAndBranch(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	head := gitOutput(t, repo, "rev-parse", "HEAD")
	manager, err := NewManager(repo, filepath.Join(t.TempDir(), "attempts"))
	if err != nil {
		t.Fatal(err)
	}
	owned := createReconcileAttempt(t, ctx, manager, head, 1)
	unknown := createReconcileAttempt(t, ctx, manager, head, 2)
	result, err := manager.ReconcileOwnedAttempt(ctx, OwnedAttempt{Attempt: owned, ExpectedHead: head})
	if err != nil || result != ReconcileComplete {
		t.Fatalf("result=%s err=%v", result, err)
	}
	if _, err := os.Stat(owned.Root); !os.IsNotExist(err) {
		t.Fatalf("owned path remains: %v", err)
	}
	if raw := gitOutput(t, repo, "branch", "--list", owned.Branch); raw != "" {
		t.Fatalf("owned branch remains: %q", raw)
	}
	if _, err := os.Stat(unknown.Root); err != nil {
		t.Fatalf("unknown worktree changed: %v", err)
	}
	if raw := gitOutput(t, repo, "branch", "--list", unknown.Branch); raw == "" {
		t.Fatal("unknown branch deleted")
	}
}

func TestReconcileOwnedAttemptPreservesLiveAndMismatchedResources(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	head := gitOutput(t, repo, "rev-parse", "HEAD")
	manager, err := NewManager(repo, filepath.Join(t.TempDir(), "attempts"))
	if err != nil {
		t.Fatal(err)
	}
	live := createReconcileAttempt(t, ctx, manager, head, 1)
	result, err := manager.ReconcileOwnedAttempt(ctx, OwnedAttempt{Attempt: live, ExpectedHead: head, Live: true})
	if err != nil || result != ReconcileSkippedLive {
		t.Fatalf("live result=%s err=%v", result, err)
	}
	if _, err := os.Stat(live.Root); err != nil {
		t.Fatalf("live worktree removed: %v", err)
	}

	mismatched := live
	mismatched.Branch += "-mismatch"
	result, err = manager.ReconcileOwnedAttempt(ctx, OwnedAttempt{Attempt: mismatched, ExpectedHead: head})
	if err == nil || result != ReconcileBlocked {
		t.Fatalf("mismatch result=%s err=%v", result, err)
	}
	if _, err := os.Stat(live.Root); err != nil {
		t.Fatalf("mismatched worktree removed: %v", err)
	}

	escaped := live
	escaped.Root = filepath.Join(t.TempDir(), "not-owned")
	result, err = manager.ReconcileOwnedAttempt(ctx, OwnedAttempt{Attempt: escaped, ExpectedHead: head})
	if err == nil || result != ReconcileBlocked {
		t.Fatalf("escaped result=%s err=%v", result, err)
	}
}

func TestReconcileOwnedAttemptIsIdempotentAndFreshRetryUsesNewResources(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	head := gitOutput(t, repo, "rev-parse", "HEAD")
	manager, err := NewManager(repo, filepath.Join(t.TempDir(), "attempts"))
	if err != nil {
		t.Fatal(err)
	}
	first := createReconcileAttempt(t, ctx, manager, head, 1)
	second := createReconcileAttempt(t, ctx, manager, head, 2)
	if first.Root == second.Root || first.Branch == second.Branch {
		t.Fatalf("retry reused resources: first=%+v second=%+v", first, second)
	}
	owned := OwnedAttempt{Attempt: first, ExpectedHead: head}
	for i := 0; i < 2; i++ {
		result, err := manager.ReconcileOwnedAttempt(ctx, owned)
		if err != nil || result != ReconcileComplete {
			t.Fatalf("pass %d result=%s err=%v", i+1, result, err)
		}
	}
	if _, err := os.Stat(second.Root); err != nil {
		t.Fatalf("fresh retry removed: %v", err)
	}
}

func TestReconcileOwnedAttemptPreservesBranchCheckedOutElsewhere(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	head := gitOutput(t, repo, "rev-parse", "HEAD")
	manager, err := NewManager(repo, filepath.Join(t.TempDir(), "attempts"))
	if err != nil {
		t.Fatal(err)
	}
	attempt := createReconcileAttempt(t, ctx, manager, head, 1)
	if _, err := git(ctx, repo, "worktree", "remove", "--force", attempt.Root); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(t.TempDir(), "other-checkout")
	if _, err := git(ctx, repo, "worktree", "add", other, attempt.Branch); err != nil {
		t.Fatal(err)
	}
	result, err := manager.ReconcileOwnedAttempt(ctx, OwnedAttempt{Attempt: attempt, ExpectedHead: head})
	if err == nil || result != ReconcileBlocked {
		t.Fatalf("checked-out result=%s err=%v", result, err)
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("other checkout removed: %v", err)
	}
	if raw := gitOutput(t, repo, "branch", "--list", attempt.Branch); raw == "" {
		t.Fatal("checked-out branch deleted")
	}
}

func TestReconcileOwnedAttemptBlocksMovedBranchWithoutRemovingIt(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	head := gitOutput(t, repo, "rev-parse", "HEAD")
	manager, err := NewManager(repo, filepath.Join(t.TempDir(), "attempts"))
	if err != nil {
		t.Fatal(err)
	}
	attempt := createReconcileAttempt(t, ctx, manager, head, 1)
	if err := os.WriteFile(filepath.Join(attempt.Root, "moved.txt"), []byte("moved\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitNoOutput(t, attempt.Root, "add", "moved.txt")
	gitNoOutput(t, attempt.Root, "commit", "-m", "move attempt")
	result, err := manager.ReconcileOwnedAttempt(ctx, OwnedAttempt{Attempt: attempt, ExpectedHead: head})
	if err == nil || result != ReconcileBlocked || !strings.Contains(err.Error(), "head") {
		t.Fatalf("result=%s err=%v", result, err)
	}
	if _, err := os.Stat(attempt.Root); err != nil {
		t.Fatalf("moved worktree removed: %v", err)
	}
}

func createReconcileAttempt(t *testing.T, ctx context.Context, manager *Manager, head string, number int64) AttemptWorktree {
	t.Helper()
	attempt, err := manager.Create(ctx, AttemptIdentity{DeliveryID: "issue-389", Generation: 1, UnitID: "execute-task/M001/S01/T01", Attempt: number, BaseHead: head})
	if err != nil {
		t.Fatal(err)
	}
	return attempt
}
