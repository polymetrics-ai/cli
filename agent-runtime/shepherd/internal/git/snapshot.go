package git

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var ErrWriteScopeBreach = errors.New("write_scope_breach")

type Snapshot struct {
	HeadSHA string
	Branch  string
	Dirty   bool
}

type Artifact struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

func Inspect(ctx context.Context, root string) (Snapshot, error) {
	if root == "" {
		return Snapshot{}, errors.New("repository root is required")
	}
	head, err := run(ctx, root, "rev-parse", "HEAD")
	if err != nil {
		return Snapshot{}, err
	}
	branchRaw, err := run(ctx, root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return Snapshot{}, errors.New("issue worktree must be attached to a branch")
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
	branch := strings.TrimSpace(string(branchRaw))
	if len(sha) != 40 {
		return Snapshot{}, errors.New("git returned an invalid head SHA")
	}
	if branch == "" || strings.ContainsAny(branch, "\r\n\x00") {
		return Snapshot{}, errors.New("git returned an invalid branch")
	}
	return Snapshot{HeadSHA: sha, Branch: branch, Dirty: len(tracked) > 0 || len(untracked) > 0}, nil
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
		if !withinAnyScope(path, scopes) && !isMutableGSDProjection(path) {
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

func ArtifactManifest(ctx context.Context, root, startHead, endHead string, scopes []string) ([]Artifact, error) {
	if root == "" || len(startHead) != 40 || len(endHead) != 40 || len(scopes) == 0 {
		return nil, errors.New("repository, exact heads, and write scopes are required")
	}
	raw, err := run(ctx, root, "diff", "--name-only", "-z", startHead, endHead, "--", ".")
	if err != nil {
		return nil, err
	}
	var artifacts []Artifact
	for _, item := range bytes.Split(raw, []byte{0}) {
		path := filepath.ToSlash(strings.TrimSpace(string(item)))
		if path == "" {
			continue
		}
		if filepath.IsAbs(path) || path == ".." || strings.HasPrefix(path, "../") || strings.ContainsRune(path, 0) {
			return nil, errors.New("git returned an unsafe artifact path")
		}
		if !withinAnyScope(path, scopes) && !isMutableGSDProjection(path) {
			return nil, fmt.Errorf("%w: artifact path %q is outside the issue write scope", ErrWriteScopeBreach, path)
		}
		raw, err := run(ctx, root, "show", endHead+":"+path)
		if err != nil {
			// Deleted files are still part of the manifest; bind deletion to a stable marker.
			artifacts = append(artifacts, Artifact{Path: path, Hash: "sha256:" + strings.Repeat("0", 64)})
			continue
		}
		hash := sha256.Sum256(raw)
		artifacts = append(artifacts, Artifact{Path: path, Hash: "sha256:" + fmt.Sprintf("%x", hash[:])})
	}
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Path < artifacts[j].Path })
	return artifacts, nil
}

// ChangedPathsOutsideScopes returns current worker changes that the immutable issue write scope
// does not authorize. It is read-only so callers can use it while a governed unit is running.
func ChangedPathsOutsideScopes(ctx context.Context, root string, scopes []string) ([]string, error) {
	if root == "" || len(scopes) == 0 {
		return nil, errors.New("repository and write scopes are required")
	}
	paths, err := changedPaths(ctx, root)
	if err != nil {
		return nil, err
	}
	outside := make([]string, 0)
	for _, path := range paths {
		if !withinAnyScope(path, scopes) && !isMutableGSDProjection(path) {
			outside = append(outside, path)
		}
	}
	return outside, nil
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
	for _, raw := range bytes.Split(tracked, []byte{0}) {
		path := filepath.ToSlash(strings.TrimSpace(string(raw)))
		if path == "" || path == ".gsd-worktrees" || strings.HasPrefix(path, ".gsd-worktrees/") {
			continue
		}
		if filepath.IsAbs(path) || path == ".." || strings.HasPrefix(path, "../") || strings.ContainsRune(path, 0) {
			return nil, errors.New("git returned an unsafe changed path")
		}
		unique[path] = struct{}{}
	}
	for _, raw := range bytes.Split(untracked, []byte{0}) {
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

func isMutableGSDProjection(path string) bool {
	switch path {
	case ".gsd/REQUIREMENTS.md", ".gsd/ROADMAP.md", ".gsd/DECISIONS.md", ".gsd/KNOWLEDGE.md", ".gsd/QUEUE.md":
		return true
	default:
		return strings.HasPrefix(path, ".gsd/phases/")
	}
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
	cmd.Env = sanitizedGitEnvironment(os.Environ())
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

func sanitizedGitEnvironment(source []string) []string {
	environment := make([]string, 0, len(source)+2)
	for _, entry := range source {
		name, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(strings.ToUpper(name), "GIT_") {
			continue
		}
		environment = append(environment, entry)
	}
	return append(environment, "GIT_TERMINAL_PROMPT=0", "GIT_ASKPASS=")
}
