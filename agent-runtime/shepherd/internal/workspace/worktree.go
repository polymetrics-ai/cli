package workspace

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

type AttemptIdentity struct {
	DeliveryID string
	Generation int64
	UnitID     string
	Attempt    int64
	BaseHead   string
}

type AttemptState string

const (
	AttemptCreated             AttemptState = "created"
	AttemptRunning             AttemptState = "running"
	AttemptValidated           AttemptState = "validated"
	AttemptPromoted            AttemptState = "promoted"
	AttemptDiscarded           AttemptState = "discarded"
	AttemptRetainedForRecovery AttemptState = "retained_for_recovery"
)

type AttemptWorktree struct {
	Root     string
	State    AttemptState
	Identity AttemptIdentity
}

type Manager struct {
	RepoRoot string
	Root     string
}

func NewManager(repoRoot, root string) (*Manager, error) {
	if repoRoot == "" || root == "" || !filepath.IsAbs(repoRoot) || !filepath.IsAbs(root) {
		return nil, errors.New("absolute repository and attempt roots are required")
	}
	canonicalRepo, err := filepath.EvalSymlinks(filepath.Clean(repoRoot))
	if err != nil {
		return nil, fmt.Errorf("resolve repository root: %w", err)
	}
	if info, err := os.Lstat(root); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("attempt root must not be a symlink")
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create attempt root: %w", err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(filepath.Clean(root))
	if err != nil {
		return nil, fmt.Errorf("resolve attempt root: %w", err)
	}
	if rel, err := filepath.Rel(canonicalRepo, canonicalRoot); err != nil || rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))) {
		return nil, errors.New("attempt root must be outside the canonical worktree")
	}
	return &Manager{RepoRoot: canonicalRepo, Root: canonicalRoot}, nil
}

func (m *Manager) Create(ctx context.Context, identity AttemptIdentity) (AttemptWorktree, error) {
	if err := validateIdentity(identity); err != nil {
		return AttemptWorktree{}, err
	}
	if head, err := git(ctx, m.RepoRoot, "rev-parse", "HEAD"); err != nil || strings.TrimSpace(string(head)) != identity.BaseHead {
		if err != nil {
			return AttemptWorktree{}, err
		}
		return AttemptWorktree{}, errors.New("canonical head does not match attempt base")
	}
	path := filepath.Join(m.Root, safePathPart(identity.DeliveryID), fmt.Sprintf("g%d", identity.Generation), safePathPart(identity.UnitID), fmt.Sprintf("attempt-%d", identity.Attempt))
	if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(m.Root)+string(os.PathSeparator)) {
		return AttemptWorktree{}, errors.New("attempt path escapes attempt root")
	}
	if _, err := os.Stat(path); err == nil {
		return AttemptWorktree{}, errors.New("attempt worktree already exists")
	} else if !os.IsNotExist(err) {
		return AttemptWorktree{}, err
	}
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return AttemptWorktree{}, err
	}
	resolvedParent, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return AttemptWorktree{}, err
	}
	if rel, err := filepath.Rel(m.Root, resolvedParent); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return AttemptWorktree{}, errors.New("attempt parent escapes managed root")
	}
	if err := rejectSymlinkAncestors(m.Root, parent); err != nil {
		return AttemptWorktree{}, err
	}
	branch := "shepherd/" + safePathPart(identity.DeliveryID) + "/" + safePathPart(identity.UnitID) + "/" + fmt.Sprintf("g%d-a%d", identity.Generation, identity.Attempt)
	if _, err := git(ctx, m.RepoRoot, "worktree", "add", "-b", branch, path, identity.BaseHead); err != nil {
		return AttemptWorktree{}, fmt.Errorf("create attempt worktree: %w", err)
	}
	return AttemptWorktree{Root: path, State: AttemptCreated, Identity: identity}, nil
}

func (m *Manager) Promote(ctx context.Context, attempt AttemptWorktree, writeScopes []string, message string) (string, error) {
	if err := validateIdentity(attempt.Identity); err != nil {
		return "", err
	}
	if strings.TrimSpace(message) == "" || strings.ContainsAny(message, "\r\n\x00") {
		return "", errors.New("one-line promotion message is required")
	}
	if head, err := git(ctx, m.RepoRoot, "rev-parse", "HEAD"); err != nil || strings.TrimSpace(string(head)) != attempt.Identity.BaseHead {
		if err != nil {
			return "", err
		}
		return "", errors.New("canonical head changed before promotion")
	}
	paths, err := changedPaths(ctx, attempt.Root)
	if err != nil {
		return "", err
	}
	for _, path := range paths {
		if !withinAnyScope(path, writeScopes) {
			return "", fmt.Errorf("changed path %q is outside the issue write scope", path)
		}
	}
	if len(paths) > 0 {
		args := append([]string{"add", "-A", "--"}, paths...)
		if _, err := git(ctx, attempt.Root, args...); err != nil {
			return "", err
		}
		if _, err := git(ctx, attempt.Root, "-c", "core.hooksPath=/dev/null", "commit", "-m", message); err != nil {
			return "", err
		}
	}
	attemptHeadRaw, err := git(ctx, attempt.Root, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	attemptHead := strings.TrimSpace(string(attemptHeadRaw))
	if attemptHead == attempt.Identity.BaseHead {
		return attemptHead, nil
	}
	if _, err := git(ctx, m.RepoRoot, "merge", "--ff-only", attemptHead); err != nil {
		return "", fmt.Errorf("promote attempt head: %w", err)
	}
	return attemptHead, nil
}

func (m *Manager) Discard(ctx context.Context, attempt AttemptWorktree) error {
	if !strings.HasPrefix(filepath.Clean(attempt.Root), filepath.Clean(m.Root)+string(os.PathSeparator)) {
		return errors.New("attempt worktree is not owned by this manager")
	}
	if _, err := git(ctx, m.RepoRoot, "worktree", "remove", "--force", attempt.Root); err != nil {
		return fmt.Errorf("discard attempt worktree: %w", err)
	}
	return nil
}

func rejectSymlinkAncestors(root, path string) error {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	if rel, err := filepath.Rel(root, path); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return errors.New("attempt path escapes attempt root")
	}
	current := root
	relative, _ := filepath.Rel(root, path)
	for _, part := range strings.Split(relative, string(os.PathSeparator)) {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("attempt path contains a symlink ancestor")
		}
	}
	return nil
}

func validateIdentity(identity AttemptIdentity) error {
	if identity.DeliveryID == "" || identity.Generation <= 0 || identity.UnitID == "" || identity.Attempt <= 0 || len(identity.BaseHead) != 40 {
		return errors.New("complete attempt identity is required")
	}
	if strings.ContainsAny(identity.DeliveryID+identity.UnitID, "\r\n\x00") {
		return errors.New("attempt identity contains control characters")
	}
	return nil
}

func safePathPart(value string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			return r
		}
		return '-'
	}, value)
}

func changedPaths(ctx context.Context, root string) ([]string, error) {
	raw, err := git(ctx, root, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, item := range bytes.Split(raw, []byte{0}) {
		if len(item) < 4 {
			continue
		}
		path := filepath.ToSlash(strings.TrimSpace(string(item[3:])))
		if path == "" || filepath.IsAbs(path) || path == ".." || strings.HasPrefix(path, "../") {
			return nil, errors.New("git returned an unsafe changed path")
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func withinAnyScope(path string, scopes []string) bool {
	path = filepath.ToSlash(path)
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

func git(ctx context.Context, root string, args ...string) ([]byte, error) {
	commandArgs := append([]string{"-C", root}, args...)
	cmd := exec.CommandContext(ctx, "git", commandArgs...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s failed: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}
