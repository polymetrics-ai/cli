package workspace

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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

const AttemptCreated AttemptState = "created"

type AttemptWorktree struct {
	Root     string
	Branch   string
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

func (m *Manager) Plan(identity AttemptIdentity) (AttemptWorktree, error) {
	if err := validateIdentity(identity); err != nil {
		return AttemptWorktree{}, err
	}
	path := filepath.Join(m.Root, safePathPart(identity.DeliveryID), fmt.Sprintf("g%d", identity.Generation), safePathPart(identity.UnitID), fmt.Sprintf("attempt-%d", identity.Attempt))
	if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(m.Root)+string(os.PathSeparator)) {
		return AttemptWorktree{}, errors.New("attempt path escapes attempt root")
	}
	branchHash := sha256.Sum256([]byte(path))
	branch := "shepherd/" + safePathPart(identity.DeliveryID) + "/" + safePathPart(identity.UnitID) + "/" + fmt.Sprintf("g%d-a%d-%s", identity.Generation, identity.Attempt, hex.EncodeToString(branchHash[:4]))
	return AttemptWorktree{Root: path, Branch: branch, State: AttemptCreated, Identity: identity}, nil
}

func (m *Manager) Create(ctx context.Context, identity AttemptIdentity) (AttemptWorktree, error) {
	planned, err := m.Plan(identity)
	if err != nil {
		return AttemptWorktree{}, err
	}
	if head, err := git(ctx, m.RepoRoot, "rev-parse", "HEAD"); err != nil || strings.TrimSpace(string(head)) != identity.BaseHead {
		if err != nil {
			return AttemptWorktree{}, err
		}
		return AttemptWorktree{}, errors.New("canonical head does not match attempt base")
	}
	path := planned.Root
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
	if _, err := git(ctx, m.RepoRoot, "worktree", "add", "-b", planned.Branch, path, identity.BaseHead); err != nil {
		return AttemptWorktree{}, fmt.Errorf("create attempt worktree: %w", err)
	}
	return planned, nil
}

func (m *Manager) PrepareGSDState(ctx context.Context, attempt AttemptWorktree) error {
	if err := validateIdentity(attempt.Identity); err != nil {
		return err
	}
	return copyGSDState(ctx, filepath.Join(m.RepoRoot, ".gsd"), filepath.Join(attempt.Root, ".gsd"))
}

func (m *Manager) AdoptGSDState(ctx context.Context, attempt AttemptWorktree) error {
	if err := validateIdentity(attempt.Identity); err != nil {
		return err
	}
	if !strings.HasPrefix(filepath.Clean(attempt.Root), filepath.Clean(m.Root)+string(os.PathSeparator)) {
		return errors.New("attempt worktree is not owned by this manager")
	}
	return copyGSDState(ctx, filepath.Join(attempt.Root, ".gsd"), filepath.Join(m.RepoRoot, ".gsd"))
}

func (m *Manager) CheckpointCandidate(ctx context.Context, attempt AttemptWorktree, writeScopes []string, message string) (string, error) {
	if err := validateIdentity(attempt.Identity); err != nil {
		return "", err
	}
	if strings.TrimSpace(message) == "" || strings.ContainsAny(message, "\r\n\x00") {
		return "", errors.New("one-line candidate checkpoint message is required")
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
	return strings.TrimSpace(string(attemptHeadRaw)), nil
}

func (m *Manager) PromoteCandidate(ctx context.Context, attempt AttemptWorktree, candidateHead string) error {
	if err := validateIdentity(attempt.Identity); err != nil {
		return err
	}
	if len(candidateHead) != 40 {
		return errors.New("candidate head is required for promotion")
	}
	if head, err := git(ctx, m.RepoRoot, "rev-parse", "HEAD"); err != nil || strings.TrimSpace(string(head)) != attempt.Identity.BaseHead {
		if err != nil {
			return err
		}
		return errors.New("canonical head changed before promotion")
	}
	attemptHeadRaw, err := git(ctx, attempt.Root, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(attemptHeadRaw)) != candidateHead {
		return errors.New("attempt candidate head moved before promotion")
	}
	if candidateHead == attempt.Identity.BaseHead {
		return nil
	}
	if _, err := git(ctx, m.RepoRoot, "merge", "--ff-only", candidateHead); err != nil {
		return fmt.Errorf("promote attempt head: %w", err)
	}
	return nil
}

func (m *Manager) Promote(ctx context.Context, attempt AttemptWorktree, writeScopes []string, message string) (string, error) {
	candidateHead, err := m.CheckpointCandidate(ctx, attempt, writeScopes, message)
	if err != nil {
		return "", err
	}
	if err := m.PromoteCandidate(ctx, attempt, candidateHead); err != nil {
		return "", err
	}
	return candidateHead, nil
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

func copyGSDState(ctx context.Context, src, dst string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("inspect GSD state: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("GSD state root must be a real directory")
	}
	if info, err := os.Lstat(dst); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return errors.New("destination GSD state root must not be a symlink")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("clear destination GSD state: %w", err)
	}
	if err := os.MkdirAll(dst, 0o700); err != nil {
		return fmt.Errorf("create destination GSD state: %w", err)
	}
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return errors.New("GSD state path escapes source root")
		}
		if rel == "." {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("GSD state must not contain symlinks")
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		return copyRegularFile(path, target, info.Mode().Perm())
	})
}

func copyRegularFile(src, dst string, mode os.FileMode) error {
	if mode == 0 {
		mode = 0o600
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
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
		if path == ".gsd" || strings.HasPrefix(path, ".gsd/") || path == ".gsd-worktrees" || strings.HasPrefix(path, ".gsd-worktrees/") {
			continue
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

type boundedGitBuffer struct {
	bytes.Buffer
	limit    int
	exceeded bool
}

func (b *boundedGitBuffer) Write(p []byte) (int, error) {
	original := len(p)
	remaining := b.limit - b.Len()
	if remaining <= 0 {
		b.exceeded = true
		return original, nil
	}
	if len(p) > remaining {
		_, _ = b.Buffer.Write(p[:remaining])
		b.exceeded = true
		return original, nil
	}
	_, _ = b.Buffer.Write(p)
	return original, nil
}

func git(ctx context.Context, root string, args ...string) ([]byte, error) {
	commandArgs := append([]string{"-C", root}, args...)
	cmd := exec.CommandContext(ctx, "git", commandArgs...)
	environment := make([]string, 0, len(os.Environ())+2)
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(strings.ToUpper(name), "GIT_") {
			continue
		}
		environment = append(environment, entry)
	}
	cmd.Env = append(environment, "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=")
	stdout := boundedGitBuffer{limit: 2 * 1024 * 1024}
	stderr := boundedGitBuffer{limit: 64 * 1024}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		diagnostic := strings.TrimSpace(strings.Map(func(r rune) rune {
			if r < 0x20 || r == 0x7f {
				return ' '
			}
			return r
		}, stderr.String()))
		return nil, fmt.Errorf("git %s failed: %w: %s", args[0], err, diagnostic)
	}
	if stdout.exceeded || stderr.exceeded {
		return nil, fmt.Errorf("git %s output exceeded the governed limit", args[0])
	}
	return stdout.Bytes(), nil
}
