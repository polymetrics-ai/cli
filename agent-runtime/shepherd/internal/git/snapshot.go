package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	// Tracked changes are never exempt, including changes to governed policy in
	// .gsd. Official GSD Pi does own untracked project runtime/projection files
	// under .gsd and survivor worktrees under .gsd-worktrees, so inspect source
	// changes separately from those two control-plane paths.
	tracked, err := run(ctx, root, "status", "--porcelain=v1", "-z", "--untracked-files=no", "--", ".")
	if err != nil {
		return Snapshot{}, err
	}
	untracked, err := run(ctx, root, "status", "--porcelain=v1", "-z", "--untracked-files=all", "--", ".", ":(exclude).gsd/**", ":(exclude).gsd-worktrees/**")
	if err != nil {
		return Snapshot{}, err
	}
	sha := strings.TrimSpace(string(head))
	if len(sha) != 40 {
		return Snapshot{}, errors.New("git returned an invalid head SHA")
	}
	return Snapshot{HeadSHA: sha, Dirty: len(tracked) > 0 || len(untracked) > 0}, nil
}

func RequireClean(snapshot Snapshot) error {
	if snapshot.Dirty {
		return errors.New("worktree is dirty; reconcile before dispatch")
	}
	return nil
}

func CheckpointWithinScopes(ctx context.Context, root string, scopes []string, message string) (string, error) {
	if root == "" || len(scopes) == 0 || strings.TrimSpace(message) == "" || strings.ContainsAny(message, "\r\n\x00") {
		return "", errors.New("repository, write scopes, and one-line checkpoint message are required")
	}
	paths, err := changedPaths(ctx, root)
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		head, err := run(ctx, root, "rev-parse", "HEAD")
		return strings.TrimSpace(string(head)), err
	}
	for _, path := range paths {
		if !withinAnyScope(path, scopes) {
			return "", fmt.Errorf("changed path %q is outside the issue write scope", path)
		}
	}
	args := append([]string{"add", "-A", "--"}, paths...)
	if _, err := run(ctx, root, args...); err != nil {
		return "", fmt.Errorf("stage scoped checkpoint: %w", err)
	}
	if _, err := run(ctx, root, "-c", "core.hooksPath=/dev/null", "commit", "-m", message); err != nil {
		return "", fmt.Errorf("commit scoped checkpoint: %w", err)
	}
	snapshot, err := Inspect(ctx, root)
	if err != nil {
		return "", err
	}
	if snapshot.Dirty {
		return "", errors.New("worktree remains dirty after scoped checkpoint")
	}
	return snapshot.HeadSHA, nil
}

func changedPaths(ctx context.Context, root string) ([]string, error) {
	tracked, err := run(ctx, root, "diff", "--name-only", "-z", "HEAD", "--", ".")
	if err != nil {
		return nil, err
	}
	untracked, err := run(ctx, root, "ls-files", "--others", "--exclude-standard", "-z", "--", ".")
	if err != nil {
		return nil, err
	}
	unique := make(map[string]struct{})
	for _, raw := range append(bytes.Split(tracked, []byte{0}), bytes.Split(untracked, []byte{0})...) {
		path := filepath.ToSlash(strings.TrimSpace(string(raw)))
		if path == "" || path == ".gsd" || strings.HasPrefix(path, ".gsd/") || path == ".gsd-worktrees" || strings.HasPrefix(path, ".gsd-worktrees/") {
			continue
		}
		if filepath.IsAbs(path) || path == ".." || strings.HasPrefix(path, "../") || strings.ContainsRune(path, 0) {
			return nil, errors.New("git returned an unsafe changed path")
		}
		unique[path] = struct{}{}
	}
	paths := make([]string, 0, len(unique))
	for path := range unique {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func withinAnyScope(path string, scopes []string) bool {
	for _, scope := range scopes {
		scope = filepath.ToSlash(strings.TrimSpace(scope))
		if prefix, ok := strings.CutSuffix(scope, "/**"); ok {
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				return true
			}
			continue
		}
		if path == scope {
			return true
		}
	}
	return false
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
