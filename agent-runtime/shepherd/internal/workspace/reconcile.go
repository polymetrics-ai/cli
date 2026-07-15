package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ReconcileResult string

const (
	ReconcileComplete    ReconcileResult = "complete"
	ReconcileSkippedLive ReconcileResult = "skipped_live"
	ReconcileBlocked     ReconcileResult = "blocked"
)

type OwnedAttempt struct {
	Attempt      AttemptWorktree
	ExpectedHead string
	Live         bool
}

type registeredWorktree struct {
	Path   string
	Head   string
	Branch string
}

// VerifyCreatedAttempt positively proves that git worktree add produced the
// deterministic path, branch, head, common-directory, and no-symlink identity.
func (m *Manager) VerifyCreatedAttempt(ctx context.Context, attempt AttemptWorktree) error {
	planned, err := m.Plan(attempt.Identity)
	if err != nil {
		return err
	}
	if filepath.Clean(attempt.Root) != planned.Root || attempt.Branch != planned.Branch {
		return errors.New("created attempt path or branch does not match manager plan")
	}
	worktrees, err := m.registeredWorktrees(ctx)
	if err != nil {
		return err
	}
	registered, found := findRegisteredWorktree(worktrees, planned.Root)
	if !found || registered.Branch != "refs/heads/"+planned.Branch || registered.Head != attempt.Identity.BaseHead {
		return errors.New("created attempt is not exactly registered at its base head")
	}
	return m.verifyManagedWorktree(ctx, planned, attempt.Identity.BaseHead)
}

// ReconcileOwnedAttempt removes only a resource whose path, branch, identity, and head
// match the manager's deterministic plan. Unknown, live, moved, or mismatched resources
// are left untouched.
func (m *Manager) ReconcileOwnedAttempt(ctx context.Context, owned OwnedAttempt) (ReconcileResult, error) {
	planned, err := m.Plan(owned.Attempt.Identity)
	if err != nil {
		return ReconcileBlocked, err
	}
	if filepath.Clean(owned.Attempt.Root) != planned.Root || owned.Attempt.Branch != planned.Branch {
		return ReconcileBlocked, errors.New("durable attempt path or branch does not match manager ownership")
	}
	if len(owned.ExpectedHead) != 40 {
		return ReconcileBlocked, errors.New("exact expected attempt head is required")
	}
	if owned.Live {
		return ReconcileSkippedLive, nil
	}
	worktrees, err := m.registeredWorktrees(ctx)
	if err != nil {
		return ReconcileBlocked, err
	}
	registered, found := findRegisteredWorktree(worktrees, planned.Root)
	if found {
		if registered.Branch != "refs/heads/"+planned.Branch {
			return ReconcileBlocked, errors.New("registered attempt branch does not match durable ownership")
		}
		if registered.Head != owned.ExpectedHead {
			return ReconcileBlocked, errors.New("registered attempt head moved from durable ownership")
		}
		if err := m.verifyManagedWorktree(ctx, planned, owned.ExpectedHead); err != nil {
			return ReconcileBlocked, err
		}
		if _, err := git(ctx, m.RepoRoot, "worktree", "remove", "--force", planned.Root); err != nil {
			return ReconcileBlocked, fmt.Errorf("remove exact owned worktree: %w", err)
		}
	} else if _, statErr := os.Lstat(planned.Root); statErr == nil {
		return ReconcileBlocked, errors.New("unregistered attempt path exists and will not be removed")
	} else if !os.IsNotExist(statErr) {
		return ReconcileBlocked, fmt.Errorf("inspect owned attempt path: %w", statErr)
	}

	worktrees, err = m.registeredWorktrees(ctx)
	if err != nil {
		return ReconcileBlocked, err
	}
	for _, worktree := range worktrees {
		if worktree.Branch == "refs/heads/"+planned.Branch {
			return ReconcileBlocked, errors.New("attempt branch is checked out in another worktree")
		}
	}
	listed, err := git(ctx, m.RepoRoot, "branch", "--list", planned.Branch)
	if err != nil {
		return ReconcileBlocked, err
	}
	if strings.TrimSpace(string(listed)) == "" {
		return ReconcileComplete, nil
	}
	branchHead, err := git(ctx, m.RepoRoot, "rev-parse", "refs/heads/"+planned.Branch)
	if err != nil {
		return ReconcileBlocked, err
	}
	if strings.TrimSpace(string(branchHead)) != owned.ExpectedHead {
		return ReconcileBlocked, errors.New("attempt branch head moved from durable ownership")
	}
	if _, err := git(ctx, m.RepoRoot, "update-ref", "-d", "refs/heads/"+planned.Branch, owned.ExpectedHead); err != nil {
		return ReconcileBlocked, fmt.Errorf("atomically delete exact owned branch: %w", err)
	}
	return ReconcileComplete, nil
}

func (m *Manager) verifyManagedWorktree(ctx context.Context, planned AttemptWorktree, expectedHead string) error {
	rootInfo, err := os.Lstat(m.Root)
	if err != nil || rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return errors.New("managed attempt root is not a real directory")
	}
	if err := rejectSymlinkAncestors(m.Root, filepath.Dir(planned.Root)); err != nil {
		return fmt.Errorf("verify attempt ancestors: %w", err)
	}
	info, err := os.Lstat(planned.Root)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("registered attempt path is not a real owned directory")
	}
	gitMarker, err := os.Lstat(filepath.Join(planned.Root, ".git"))
	if err != nil || gitMarker.Mode()&os.ModeSymlink != 0 || !gitMarker.Mode().IsRegular() {
		return errors.New("registered attempt Git marker is not a real file")
	}
	attemptTop, err := git(ctx, planned.Root, "rev-parse", "--show-toplevel")
	if err != nil || filepath.Clean(strings.TrimSpace(string(attemptTop))) != planned.Root {
		return errors.New("attempt top-level identity does not match durable path")
	}
	canonicalCommon, err := git(ctx, m.RepoRoot, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return err
	}
	attemptCommon, err := git(ctx, planned.Root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil || filepath.Clean(strings.TrimSpace(string(attemptCommon))) != filepath.Clean(strings.TrimSpace(string(canonicalCommon))) {
		return errors.New("attempt Git common directory does not match canonical repository")
	}
	branch, err := git(ctx, planned.Root, "symbolic-ref", "HEAD")
	if err != nil || strings.TrimSpace(string(branch)) != "refs/heads/"+planned.Branch {
		return errors.New("attempt checked-out branch does not match durable ownership")
	}
	head, err := git(ctx, planned.Root, "rev-parse", "HEAD")
	if err != nil || strings.TrimSpace(string(head)) != expectedHead {
		return errors.New("attempt checked-out head moved from durable ownership")
	}
	return nil
}

// ProveOwnedResourcesAbsent verifies that an ambiguous attempt has already been
// cleared by a human/operator. It never removes a path or ref.
func (m *Manager) ProveOwnedResourcesAbsent(ctx context.Context, attempt AttemptWorktree) error {
	planned, err := m.Plan(attempt.Identity)
	if err != nil {
		return err
	}
	if filepath.Clean(attempt.Root) != planned.Root || attempt.Branch != planned.Branch {
		return errors.New("durable attempt path or branch does not match manager ownership")
	}
	if _, err := os.Lstat(planned.Root); err == nil {
		return errors.New("ambiguous attempt path still exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	worktrees, err := m.registeredWorktrees(ctx)
	if err != nil {
		return err
	}
	for _, worktree := range worktrees {
		if worktree.Path == planned.Root || worktree.Branch == "refs/heads/"+planned.Branch {
			return errors.New("ambiguous attempt remains registered or checked out")
		}
	}
	listed, err := git(ctx, m.RepoRoot, "branch", "--list", planned.Branch)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(listed)) != "" {
		return errors.New("ambiguous attempt branch still exists")
	}
	return nil
}

func (m *Manager) registeredWorktrees(ctx context.Context) ([]registeredWorktree, error) {
	raw, err := git(ctx, m.RepoRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var result []registeredWorktree
	var current registeredWorktree
	flush := func() {
		if current.Path != "" {
			current.Path = filepath.Clean(current.Path)
			result = append(result, current)
		}
		current = registeredWorktree{}
	}
	for _, line := range strings.Split(string(raw), "\n") {
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "worktree "):
			if current.Path != "" {
				flush()
			}
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(line, "branch ")
		}
	}
	flush()
	return result, nil
}

func findRegisteredWorktree(worktrees []registeredWorktree, path string) (registeredWorktree, bool) {
	path = filepath.Clean(path)
	for _, worktree := range worktrees {
		if worktree.Path == path {
			return worktree, true
		}
	}
	return registeredWorktree{}, false
}
