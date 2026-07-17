package git

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
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

var ErrWriteScopeBreach = errors.New("write_scope_breach")
var ErrOutputLimit = errors.New("git_output_limit")

const DeletionSentinelHash = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
const maxGitStdoutBytes = 1024 * 1024
const maxGitStderrBytes = 64 * 1024
const maxGitObjectBytes = 8 * 1024 * 1024
const maxGitArtifacts = 128

type Snapshot struct {
	HeadSHA string
	Branch  string
	Dirty   bool
}

type Artifact struct {
	Path    string `json:"path"`
	Hash    string `json:"hash"`
	Deleted bool   `json:"deleted,omitempty"`
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
	raw, err := run(ctx, root, "diff", "--name-status", "-z", "--no-renames", startHead, endHead, "--", ".")
	if err != nil {
		return nil, err
	}
	entries, err := parseNameStatus(raw, scopes)
	if err != nil {
		return nil, err
	}
	var artifacts []Artifact
	for _, entry := range entries {
		switch entry.status {
		case "D":
			artifacts = append(artifacts, Artifact{Path: entry.path, Hash: DeletionSentinelHash, Deleted: true})
		case "A", "M", "T":
			hash, err := hashGitObject(ctx, root, endHead+":"+entry.path)
			if err != nil {
				return nil, err
			}
			if hash == DeletionSentinelHash {
				return nil, errors.New("present artifact unexpectedly hashed to the deletion sentinel")
			}
			artifacts = append(artifacts, Artifact{Path: entry.path, Hash: hash})
		default:
			return nil, fmt.Errorf("git returned unsupported artifact status %q", entry.status)
		}
	}
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Path < artifacts[j].Path })
	return artifacts, nil
}

type nameStatusEntry struct {
	status string
	path   string
}

func parseNameStatus(raw []byte, scopes []string) ([]nameStatusEntry, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if raw[0] == 0 || raw[len(raw)-1] != 0 {
		return nil, errors.New("git returned malformed artifact status")
	}
	records := bytes.Split(raw, []byte{0})
	if len(records) < 2 || len(records[len(records)-1]) != 0 {
		return nil, errors.New("git returned malformed artifact status")
	}
	records = records[:len(records)-1]
	if len(records)%2 != 0 {
		return nil, errors.New("git returned malformed artifact status")
	}
	if len(records)/2 > maxGitArtifacts {
		return nil, errors.New("git artifact set exceeds the governed limit")
	}
	entries := make([]nameStatusEntry, 0, len(records)/2)
	for i := 0; i < len(records); i += 2 {
		if len(records[i]) == 0 || len(records[i+1]) == 0 {
			return nil, errors.New("git returned malformed artifact status")
		}
		status := string(records[i])
		switch status {
		case "A", "M", "T", "D":
		default:
			return nil, fmt.Errorf("git returned unsupported artifact status %q", status)
		}
		path := filepath.ToSlash(string(records[i+1]))
		if err := validateArtifactPath(path); err != nil {
			return nil, err
		}
		if !withinAnyScope(path, scopes) && !isMutableGSDProjection(path) {
			return nil, fmt.Errorf("%w: artifact path %q is outside the issue write scope", ErrWriteScopeBreach, path)
		}
		entries = append(entries, nameStatusEntry{status: status, path: path})
	}
	return entries, nil
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

func validateArtifactPath(path string) error {
	if path == "" || strings.TrimSpace(path) != path || filepath.IsAbs(path) || filepath.Clean(path) != path || path == ".." || strings.HasPrefix(path, "../") || strings.IndexFunc(path, unicode.IsControl) >= 0 {
		return errors.New("git returned an unsafe artifact path")
	}
	return nil
}

func hashGitObject(ctx context.Context, root, object string) (string, error) {
	sizeRaw, err := run(ctx, root, "cat-file", "-s", object)
	if err != nil {
		return "", fmt.Errorf("git cat-file size failed: %w", err)
	}
	declared, err := strconv.ParseInt(strings.TrimSpace(string(sizeRaw)), 10, 64)
	if err != nil || declared < 0 {
		return "", errors.New("git cat-file returned an invalid object size")
	}
	if declared > maxGitObjectBytes {
		return "", fmt.Errorf("hash git object: %w", ErrOutputLimit)
	}
	internalCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	cmd := exec.CommandContext(internalCtx, "git", "-C", root, "cat-file", "blob", object)
	cmd.Env = sanitizedGitEnvironment(os.Environ())
	configureGitProcessTree(cmd)
	stderr := newBoundedBuffer(maxGitStderrBytes, func() { cancel(ErrOutputLimit) })
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	digest := sha256.New()
	written, copyErr := copyHashBounded(digest, stdout, declared, func() { cancel(ErrOutputLimit) })
	waitErr := cmd.Wait()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return "", ctxErr
	}
	if errors.Is(context.Cause(internalCtx), ErrOutputLimit) || written > declared || stderr.Exceeded() {
		return "", fmt.Errorf("hash git object: %w", ErrOutputLimit)
	}
	if copyErr != nil {
		return "", copyErr
	}
	if waitErr != nil {
		return "", fmt.Errorf("git cat-file failed: %w: %s", waitErr, sanitizeDiagnostic(stderr.String()))
	}
	if written != declared {
		return "", errors.New("git cat-file streamed byte count does not match declared size")
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), nil
}

func copyHashBounded(digest io.Writer, source io.Reader, declared int64, onLimit func()) (int64, error) {
	var total int64
	buffer := make([]byte, 32*1024)
	for {
		n, err := source.Read(buffer)
		if n > 0 {
			chunk := buffer[:n]
			if total+int64(n) > declared {
				allowed := declared - total
				if allowed > 0 {
					if _, writeErr := digest.Write(chunk[:allowed]); writeErr != nil {
						return total + int64(n), writeErr
					}
				}
				if onLimit != nil {
					onLimit()
				}
				return total + int64(n), ErrOutputLimit
			}
			if _, writeErr := digest.Write(chunk); writeErr != nil {
				return total, writeErr
			}
			total += int64(n)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return total, nil
			}
			return total, err
		}
	}
}

func run(ctx context.Context, root string, args ...string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("git argv is required")
	}
	commandArgs := append([]string{"-C", root}, args...)
	internalCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	cmd := exec.CommandContext(internalCtx, "git", commandArgs...)
	cmd.Env = sanitizedGitEnvironment(os.Environ())
	configureGitProcessTree(cmd)
	stdout := newBoundedBuffer(maxGitStdoutBytes, func() { cancel(ErrOutputLimit) })
	stderr := newBoundedBuffer(maxGitStderrBytes, func() { cancel(ErrOutputLimit) })
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}
	if errors.Is(context.Cause(internalCtx), ErrOutputLimit) || stdout.Exceeded() || stderr.Exceeded() {
		return nil, fmt.Errorf("git %s output exceeded governed limit: %w", args[0], ErrOutputLimit)
	}
	if err != nil {
		return nil, fmt.Errorf("git %s failed: %w: %s", args[0], err, sanitizeDiagnostic(stderr.String()))
	}
	return stdout.Bytes(), nil
}

func sanitizeDiagnostic(value string) string {
	value = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' || !unicode.IsControl(r) {
			return r
		}
		return ' '
	}, value)
	value = strings.TrimSpace(value)
	if len(value) > 2048 {
		return value[:2048] + "..."
	}
	return value
}

type boundedBuffer struct {
	mu       sync.Mutex
	buffer   bytes.Buffer
	max      int
	exceeded bool
	onLimit  func()
	once     sync.Once
}

func newBoundedBuffer(max int, onLimit func()) *boundedBuffer {
	return &boundedBuffer{max: max, onLimit: onLimit}
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.max <= 0 {
		b.markExceededLocked()
		return len(p), ErrOutputLimit
	}
	remaining := b.max - b.buffer.Len()
	if remaining <= 0 {
		b.markExceededLocked()
		return len(p), ErrOutputLimit
	}
	if len(p) > remaining {
		_, _ = b.buffer.Write(p[:remaining])
		b.markExceededLocked()
		return len(p), ErrOutputLimit
	}
	_, _ = b.buffer.Write(p)
	return len(p), nil
}

func (b *boundedBuffer) markExceededLocked() {
	b.exceeded = true
	b.once.Do(func() {
		if b.onLimit != nil {
			b.onLimit()
		}
	})
}

func (b *boundedBuffer) Exceeded() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.exceeded
}

func (b *boundedBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]byte(nil), b.buffer.Bytes()...)
}

func (b *boundedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buffer.String()
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
