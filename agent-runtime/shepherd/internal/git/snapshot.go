package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Snapshot struct {
	HeadSHA string
	Dirty   bool
}

func Inspect(ctx context.Context, root string) (Snapshot, error) {
	if root == "" {
		return Snapshot{}, errors.New("repository root is required")
	}
	head, err := run(ctx, root, "rev-parse", "HEAD")
	if err != nil {
		return Snapshot{}, err
	}
	// GSD owns nested survivor worktrees under this exact control-plane path.
	// Their files are intentionally untracked from the parent worktree and must
	// not make an otherwise clean governed unit fail its postcondition.
	status, err := run(ctx, root, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--", ".", ":(exclude).gsd-worktrees/**")
	if err != nil {
		return Snapshot{}, err
	}
	sha := strings.TrimSpace(string(head))
	if len(sha) != 40 {
		return Snapshot{}, errors.New("git returned an invalid head SHA")
	}
	return Snapshot{HeadSHA: sha, Dirty: len(status) > 0}, nil
}

func RequireClean(snapshot Snapshot) error {
	if snapshot.Dirty {
		return errors.New("worktree is dirty; reconcile before dispatch")
	}
	return nil
}

// RestoreIndex discards only transient staging performed by the governed runtime. It does not
// modify working-tree files, so real source edits remain visible to the post-unit cleanliness gate.
func RestoreIndex(ctx context.Context, root string) error {
	if root == "" {
		return errors.New("repository root is required")
	}
	if _, err := run(ctx, root, "read-tree", "HEAD"); err != nil {
		return fmt.Errorf("restore git index: %w", err)
	}
	return nil
}

// FinalizeInitializingMilestoneWorktree repairs only the narrow interrupted
// `git worktree add` state produced when the container exits after GSD emits
// milestone-ready but before Git publishes its index. It never resets source
// files: read-tree reconstructs the missing index and any real file difference
// remains visible to the final cleanliness check.
func FinalizeInitializingMilestoneWorktree(ctx context.Context, root, milestoneID, expectedHead string) error {
	if !validMilestoneID(milestoneID) || len(expectedHead) != 40 {
		return errors.New("milestone worktree repair requires a valid milestone and expected head")
	}
	child := filepath.Join(root, ".gsd-worktrees", milestoneID)
	branch, err := run(ctx, child, "symbolic-ref", "--short", "HEAD")
	if err != nil || strings.TrimSpace(string(branch)) != "milestone/"+milestoneID {
		return errors.New("managed worktree branch does not match the bound milestone")
	}
	for _, workDir := range []string{root, child} {
		head, headErr := run(ctx, workDir, "rev-parse", "HEAD")
		if headErr != nil || strings.TrimSpace(string(head)) != expectedHead {
			return errors.New("managed worktree head does not match the governed unit baseline")
		}
	}
	indexRaw, err := run(ctx, child, "rev-parse", "--git-path", "index")
	if err != nil {
		return err
	}
	lockedRaw, err := run(ctx, child, "rev-parse", "--git-path", "locked")
	if err != nil {
		return err
	}
	index := strings.TrimSpace(string(indexRaw))
	locked := strings.TrimSpace(string(lockedRaw))
	if indexInfo, indexErr := os.Stat(index); indexErr == nil && indexInfo.Mode().IsRegular() {
		if _, lockedErr := os.Stat(locked); !os.IsNotExist(lockedErr) {
			return errors.New("managed worktree has an index but remains locked")
		}
		status, statusErr := run(ctx, child, "status", "--porcelain=v1", "-z", "--untracked-files=all")
		if statusErr != nil || len(status) != 0 {
			return errors.New("initialized managed worktree is dirty")
		}
		return nil
	} else if !os.IsNotExist(indexErr) {
		return errors.New("managed worktree index is not a regular file")
	}
	lockInfo, err := os.Stat(index + ".lock")
	if err != nil || !lockInfo.Mode().IsRegular() || lockInfo.Size() != 0 {
		return errors.New("managed worktree index lock is not a safe interrupted marker")
	}
	lockedReason, err := os.ReadFile(locked)
	if err != nil || strings.TrimSpace(string(lockedReason)) != "initializing" {
		return errors.New("managed worktree is not locked for initialization")
	}
	if err := os.Remove(index + ".lock"); err != nil {
		return fmt.Errorf("remove stale managed index lock: %w", err)
	}
	if _, err := run(ctx, child, "read-tree", "HEAD"); err != nil {
		return fmt.Errorf("reconstruct managed worktree index: %w", err)
	}
	if _, err := run(ctx, root, "worktree", "unlock", child); err != nil {
		return fmt.Errorf("unlock initialized managed worktree: %w", err)
	}
	status, err := run(ctx, child, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil || len(status) != 0 {
		return errors.New("managed worktree contains changes after index reconstruction")
	}
	return nil
}

func validMilestoneID(value string) bool {
	if !strings.HasPrefix(value, "M") || len(value) < 3 || len(value) > 64 {
		return false
	}
	for _, char := range value[1:] {
		if (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
			return false
		}
	}
	return true
}

func run(ctx context.Context, root string, args ...string) ([]byte, error) {
	commandArgs := append([]string{"-C", root}, args...)
	cmd := exec.CommandContext(ctx, "git", commandArgs...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s failed: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	if stdout.Len() > 1024*1024 {
		return nil, errors.New("git output exceeds safety limit")
	}
	return stdout.Bytes(), nil
}
