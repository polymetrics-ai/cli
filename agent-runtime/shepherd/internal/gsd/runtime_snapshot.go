package gsd

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	runtimeTreeSchema           = "shepherd.gsd-runtime-tree/v3"
	maxRuntimeTreeEntries       = 100_000
	maxRuntimeTreeBytes   int64 = 1024 * 1024 * 1024
	maxRuntimeFileBytes   int64 = 256 * 1024 * 1024
	maxRuntimeSourceBytes int64 = 2 * 1024 * 1024
)

type hostRuntimeQualification struct {
	NodeSHA256              string
	SourceTreeSHA256        string
	PatchedTreeSHA256       string
	CopiedPatchedTreeSHA256 string
	SnapshotTreeSHA256      string
}

var qualifiedHostRuntimes = map[string]hostRuntimeQualification{
	"darwin/arm64": {
		NodeSHA256:              "d36b3d980963d44bd2c5e844fac4cfeee26a167b744287a4e74a9575af9d0559",
		SourceTreeSHA256:        "3baebd05969d6ed2520d073810e27be42680b8af7bcaa3d57ba658ec7cfb2f2e",
		PatchedTreeSHA256:       "de76fc87a999f65e8420f0f781438fa763c10432991b776545930fe65091753e",
		CopiedPatchedTreeSHA256: "5db688e178ba70cb39749f88d290bdc67c7d4b4cbadbfd28197dd6c5ed3f83b2",
		SnapshotTreeSHA256:      "074a833106a0c5978524300032c4095d0fb52d2c543c5e6b58dbfbfa8fbc50fa",
	},
}

type HostRuntimeGuard struct {
	nodePath            string
	root                string
	options             hostRuntimeSnapshotOptions
	promptCommand       []string
	gsdHome             string
	promptRegistry      UnitRegistry
	modelWorkDir        string
	coordinatorModel    string
	implementationModel string
	expectedThinking    string
}

type hostRuntimeSnapshotOptions struct {
	qualification        hostRuntimeQualification
	originalComposerHash string
	patchedComposerHash  string
	maxEntries           int
	maxTreeBytes         int64
	maxFileBytes         int64
}

func defaultHostRuntimeSnapshotOptions() (hostRuntimeSnapshotOptions, error) {
	qualification, ok := qualifiedHostRuntimes[runtime.GOOS+"/"+runtime.GOARCH]
	if !ok {
		return hostRuntimeSnapshotOptions{}, registryMismatch("host GSD runtime is not qualified for this platform")
	}
	return hostRuntimeSnapshotOptions{
		qualification:        qualification,
		originalComposerHash: officialComposerSHA256,
		patchedComposerHash:  officialPatchedComposerSHA256,
		maxEntries:           maxRuntimeTreeEntries,
		maxTreeBytes:         maxRuntimeTreeBytes,
		maxFileBytes:         maxRuntimeFileBytes,
	}, nil
}

func NewPinnedHostRuntimeGuard(command []string) (*HostRuntimeGuard, error) {
	options, err := defaultHostRuntimeSnapshotOptions()
	if err != nil {
		return nil, err
	}
	if len(command) != 2 || !filepath.IsAbs(command[0]) || !filepath.IsAbs(command[1]) || filepath.Base(command[1]) != "loader.js" || filepath.Base(filepath.Dir(command[1])) != "dist" {
		return nil, registryMismatch("prepared host runtime command is invalid")
	}
	guard := &HostRuntimeGuard{nodePath: command[0], root: filepath.Dir(filepath.Dir(command[1])), options: options}
	if err := guard.Validate(context.Background()); err != nil {
		return nil, err
	}
	return guard, nil
}

func (g *HostRuntimeGuard) BindPromptRuntime(command []string, gsdHome string, registry UnitRegistry) error {
	if g == nil || len(command) != 2 || !filepath.IsAbs(gsdHome) || len(registry.Units) == 0 {
		return registryMismatch("host prompt runtime binding is incomplete")
	}
	g.promptCommand = append([]string(nil), command...)
	g.gsdHome = gsdHome
	g.promptRegistry = registry
	return g.Validate(context.Background())
}

func (g *HostRuntimeGuard) BindModelPolicy(workDir, coordinatorModel, implementationModel, expectedThinking string) error {
	if g == nil || !filepath.IsAbs(workDir) || coordinatorModel == "" || implementationModel == "" || expectedThinking == "" || len(g.promptRegistry.Units) == 0 {
		return registryMismatch("host model policy binding is incomplete")
	}
	g.modelWorkDir = workDir
	g.coordinatorModel = coordinatorModel
	g.implementationModel = implementationModel
	g.expectedThinking = expectedThinking
	return g.Validate(context.Background())
}

func (g *HostRuntimeGuard) Validate(ctx context.Context) error {
	return g.ValidateForWorkDir(ctx, g.modelWorkDir)
}

func (g *HostRuntimeGuard) ValidateForWorkDir(ctx context.Context, workDir string) error {
	if g == nil || ctx == nil {
		return registryMismatch("host runtime guard is required")
	}
	nodeDigest, _, err := boundedRuntimeFileSHA256(g.nodePath, 256*1024*1024)
	if err != nil || nodeDigest != g.options.qualification.NodeSHA256 {
		return registryMismatch("guarded Node executable drifted")
	}
	digest, _, _, err := runtimeTreeDigest(ctx, g.root, g.options)
	if err != nil || digest != g.options.qualification.SnapshotTreeSHA256 {
		return registryMismatch("guarded private GSD runtime drifted")
	}
	if len(g.promptCommand) > 0 {
		if err := validatePinnedPromptToolContractsWithRegistry(g.promptCommand, g.gsdHome, g.promptRegistry); err != nil {
			return err
		}
	}
	if g.coordinatorModel != "" {
		if workDir == "" {
			return registryMismatch("guarded model policy has no work directory")
		}
		if err := ValidateRuntimeSettings(g.gsdHome, workDir, g.coordinatorModel, g.expectedThinking); err != nil {
			return err
		}
		if err := ValidateModelPreferences(g.gsdHome, workDir, g.promptRegistry, g.coordinatorModel, g.implementationModel, g.expectedThinking); err != nil {
			return err
		}
	}
	return nil
}

// PreparePinnedHostRuntime verifies the complete installed package tree and
// copies it into a controller-owned, read-only snapshot before any GSD code is
// executed. The returned command always uses an absolute hash-pinned Node path.
func PreparePinnedHostRuntime(ctx context.Context, command []string, gsdHome, expectedVersion string, untrustedRoots ...string) ([]string, error) {
	options, err := defaultHostRuntimeSnapshotOptions()
	if err != nil {
		return nil, err
	}
	return preparePinnedHostRuntimeWithOptions(ctx, command, gsdHome, expectedVersion, untrustedRoots, options)
}

func preparePinnedHostRuntimeWithOptions(ctx context.Context, command []string, gsdHome, expectedVersion string, untrustedRoots []string, options hostRuntimeSnapshotOptions) ([]string, error) {
	if ctx == nil || expectedVersion != "1.11.0" || !filepath.IsAbs(gsdHome) {
		return nil, registryMismatch("host runtime snapshot requires a context, absolute GSD home, and GSD 1.11.0")
	}
	if options.maxEntries <= 0 || options.maxEntries > maxRuntimeTreeEntries || options.maxTreeBytes <= 0 || options.maxTreeBytes > maxRuntimeTreeBytes || options.maxFileBytes <= 0 || options.maxFileBytes > maxRuntimeFileBytes {
		return nil, registryMismatch("host runtime snapshot bounds are invalid")
	}
	if !validSHA256(options.qualification.NodeSHA256) || !validSHA256(options.qualification.SourceTreeSHA256) || !validSHA256(options.qualification.PatchedTreeSHA256) || !validSHA256(options.qualification.CopiedPatchedTreeSHA256) || !validSHA256(options.qualification.SnapshotTreeSHA256) || !validSHA256(options.originalComposerHash) || !validSHA256(options.patchedComposerHash) {
		return nil, registryMismatch("host runtime qualification is invalid")
	}
	nodePath, err := resolveQualifiedNode(command, options.qualification.NodeSHA256)
	if err != nil {
		return nil, err
	}
	if len(command) != 2 || !filepath.IsAbs(command[1]) || filepath.Clean(command[1]) != command[1] {
		return nil, registryMismatch("GSD loader path must be clean and absolute")
	}
	loader := command[1]
	if filepath.Base(loader) != "loader.js" || filepath.Base(filepath.Dir(loader)) != "dist" {
		return nil, registryMismatch("GSD command does not target the packaged loader")
	}
	packageRoot := filepath.Dir(filepath.Dir(loader))
	for _, path := range []string{packageRoot, filepath.Join(packageRoot, "dist"), loader, gsdHome} {
		if err := requireRuntimePathWithoutSymlink(path); err != nil {
			return nil, err
		}
	}
	if err := validatePackageIdentity(packageRoot, expectedVersion); err != nil {
		return nil, err
	}
	for _, root := range untrustedRoots {
		if strings.TrimSpace(root) == "" {
			continue
		}
		if !filepath.IsAbs(root) {
			return nil, registryMismatch("untrusted runtime boundary must be absolute")
		}
		for _, candidate := range []string{nodePath, packageRoot, gsdHome} {
			within, err := pathWithin(root, candidate)
			if err != nil {
				return nil, registryMismatch("resolve runtime trust boundary")
			}
			if within {
				return nil, registryMismatch("host runtime or controlled snapshot root is inside a worker-controlled worktree")
			}
		}
	}

	snapshotParent, err := ensureOwnedRuntimeDirectoryChain(gsdHome, "agent", "shepherd-runtime", expectedVersion)
	if err != nil {
		return nil, err
	}
	snapshotRoot := filepath.Join(snapshotParent, options.qualification.SnapshotTreeSHA256)
	if info, err := os.Lstat(snapshotRoot); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || info.Mode().Perm()&0o222 != 0 || !runtimePathOwnedByCurrentUser(info) {
			return nil, registryMismatch("private runtime snapshot has an invalid or writable type")
		}
		digest, _, _, err := runtimeTreeDigest(ctx, snapshotRoot, options)
		if err != nil || digest != options.qualification.SnapshotTreeSHA256 {
			return nil, registryMismatch("private runtime snapshot drifted")
		}
		return []string{nodePath, filepath.Join(snapshotRoot, "dist", "loader.js")}, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, registryMismatch("inspect private runtime snapshot")
	}

	sourceDigest, _, _, err := runtimeTreeDigest(ctx, packageRoot, options)
	if err != nil {
		return nil, err
	}
	if sourceDigest != options.qualification.SourceTreeSHA256 && sourceDigest != options.qualification.PatchedTreeSHA256 {
		return nil, registryMismatch("complete installed GSD runtime tree differs from the qualified package")
	}
	temporary, err := os.MkdirTemp(snapshotParent, ".runtime-snapshot-*")
	if err != nil {
		return nil, fmt.Errorf("create private runtime snapshot: %w", err)
	}
	defer func() {
		_ = makeRuntimeTreeWritable(temporary)
		_ = os.RemoveAll(temporary)
	}()
	if err := copyRuntimeTree(ctx, packageRoot, temporary, options); err != nil {
		return nil, err
	}
	temporaryCommand := []string{nodePath, filepath.Join(temporary, "dist", "loader.js")}
	if err := ApplyPinnedHeadlessToolPatch(temporaryCommand, expectedVersion); err != nil {
		return nil, err
	}
	if err := patchPromptContractRootWithHashes(filepath.Join(temporary, "dist", "resources", "extensions", "gsd"), options.originalComposerHash, options.patchedComposerHash); err != nil {
		return nil, err
	}
	patchedDigest, _, _, err := runtimeTreeDigest(ctx, temporary, options)
	if err != nil {
		return nil, err
	}
	if patchedDigest != options.qualification.CopiedPatchedTreeSHA256 {
		return nil, registryMismatch("private canonicalized patched runtime snapshot differs before read-only sealing")
	}
	if err := makeRuntimeTreeReadOnly(temporary); err != nil {
		return nil, err
	}
	sealedDigest, _, _, err := runtimeTreeDigest(ctx, temporary, options)
	if err != nil || sealedDigest != options.qualification.SnapshotTreeSHA256 {
		return nil, registryMismatch("private patched runtime snapshot differs from the qualified canonical snapshot tree")
	}
	// Darwin requires the source directory to be owner-writable for rename.
	// Restore the qualified read-only root before returning any executable path.
	if err := os.Chmod(temporary, 0o700); err != nil {
		return nil, registryMismatch("prepare sealed runtime snapshot publication")
	}
	if err := os.Rename(temporary, snapshotRoot); err != nil {
		if _, statErr := os.Lstat(snapshotRoot); statErr == nil {
			digest, _, _, verifyErr := runtimeTreeDigest(ctx, snapshotRoot, options)
			if verifyErr != nil || digest != options.qualification.SnapshotTreeSHA256 {
				return nil, registryMismatch("concurrent private runtime snapshot differs")
			}
		} else {
			return nil, fmt.Errorf("install private runtime snapshot: %w", err)
		}
	}
	if err := os.Chmod(snapshotRoot, 0o555); err != nil {
		return nil, err
	}
	publishedDigest, _, _, err := runtimeTreeDigest(ctx, snapshotRoot, options)
	if err != nil || publishedDigest != options.qualification.SnapshotTreeSHA256 {
		return nil, registryMismatch("published private runtime snapshot differs after sealing")
	}
	if err := syncDirectory(snapshotParent); err != nil {
		return nil, err
	}
	return []string{nodePath, filepath.Join(snapshotRoot, "dist", "loader.js")}, nil
}

func ensureOwnedRuntimeDirectoryChain(root string, components ...string) (string, error) {
	current := root
	for index := -1; index < len(components); index++ {
		if index >= 0 {
			component := components[index]
			if component == "" || component == "." || component == ".." || strings.ContainsRune(component, filepath.Separator) {
				return "", registryMismatch("runtime directory component is invalid")
			}
			current = filepath.Join(current, component)
			if err := os.Mkdir(current, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
				return "", fmt.Errorf("create controlled runtime directory: %w", err)
			}
		}
		info, err := os.Lstat(current)
		if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || !runtimePathOwnedByCurrentUser(info) {
			return "", registryMismatch("controlled runtime directory chain must be private, owned, and symlink-free")
		}
		if info.Mode().Perm()&0o077 != 0 {
			if err := os.Chmod(current, 0o700); err != nil {
				return "", registryMismatch("secure controlled runtime directory chain")
			}
		}
	}
	return current, nil
}

func resolveQualifiedNode(command []string, expectedHash string) (string, error) {
	if len(command) == 0 || (command[0] != "node" && !filepath.IsAbs(command[0])) {
		return "", registryMismatch("Node executable must be the node command or an absolute path")
	}
	path := command[0]
	if path == "node" {
		resolved, err := exec.LookPath(path)
		if err != nil {
			return "", registryMismatch("resolve Node executable")
		}
		path, err = filepath.Abs(resolved)
		if err != nil {
			return "", registryMismatch("resolve absolute Node executable")
		}
	}
	if filepath.Clean(path) != path {
		return "", registryMismatch("Node executable path must be clean")
	}
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 || !runtimePathOwnedByCurrentUser(info) {
		return "", registryMismatch("Node executable must be an executable non-symlink regular file")
	}
	digest, _, err := boundedRuntimeFileSHA256(path, 256*1024*1024)
	if err != nil || digest != expectedHash {
		return "", registryMismatch("Node executable differs from the qualified runtime")
	}
	return path, nil
}

func validatePackageIdentity(packageRoot, expectedVersion string) error {
	raw, _, err := readBoundedRuntimeFile(filepath.Join(packageRoot, "package.json"), 512*1024)
	if err != nil {
		return registryMismatch("read bounded package metadata")
	}
	if err := rejectDuplicateJSONFields(raw); err != nil {
		return registryMismatch("package metadata has duplicate or malformed fields")
	}
	var document map[string]json.RawMessage
	if err := jsonStrictDecode(raw, &document); err != nil {
		return registryMismatch("decode runtime package metadata")
	}
	var name, version string
	if json.Unmarshal(document["name"], &name) != nil || json.Unmarshal(document["version"], &version) != nil || name != "@opengsd/gsd-pi" || version != expectedVersion {
		return registryMismatch("runtime package name or version differs from the pinned contract")
	}
	return nil
}

func runtimeTreeDigest(ctx context.Context, root string, options hostRuntimeSnapshotOptions) (string, int, int64, error) {
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", 0, 0, registryMismatch("resolve runtime tree root")
	}
	type treeEntry struct {
		relative string
		kind     string
		target   string
		mode     os.FileMode
		size     int64
		digest   string
	}
	paths := make([]string, 1, 4096)
	paths[0] = "."
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return registryMismatch("runtime tree entry escapes its root")
		}
		paths = append(paths, relative)
		if len(paths) > options.maxEntries {
			return registryMismatch("runtime tree exceeds its entry bound")
		}
		return nil
	})
	if err != nil {
		return "", 0, 0, registryMismatch("walk complete runtime tree: " + err.Error())
	}
	sort.Strings(paths)
	entries := make([]treeEntry, len(paths))
	fileIndexes := make([]int, 0, len(paths))
	var total int64
	for index, relative := range paths {
		path := filepath.Join(root, relative)
		info, err := os.Lstat(path)
		if err != nil {
			return "", 0, 0, registryMismatch("inspect runtime tree entry")
		}
		if !runtimePathOwnedByCurrentUser(info) {
			return "", 0, 0, registryMismatch("runtime tree entry is not owned by the controller user")
		}
		entry := treeEntry{relative: filepath.ToSlash(relative), mode: info.Mode().Perm()}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			entry.kind = "l"
			entry.target, err = os.Readlink(path)
			if err != nil || entry.target == "" || strings.ContainsRune(entry.target, '\x00') {
				return "", 0, 0, registryMismatch("runtime symlink target is invalid")
			}
			resolved, resolveErr := filepath.EvalSymlinks(path)
			if resolveErr != nil {
				return "", 0, 0, registryMismatch("resolve runtime symlink")
			}
			within, insideErr := pathInside(canonicalRoot, resolved)
			if insideErr != nil || !within {
				return "", 0, 0, registryMismatch(fmt.Sprintf("runtime symlink %s escapes the package root to %s", entry.relative, resolved))
			}
		case info.IsDir():
			entry.kind = "d"
		case info.Mode().IsRegular():
			entry.kind, entry.size = "f", info.Size()
			if entry.size < 0 || entry.size > options.maxFileBytes {
				return "", 0, 0, registryMismatch("runtime source exceeds its file bound")
			}
			total += entry.size
			if total > options.maxTreeBytes {
				return "", 0, 0, registryMismatch("runtime tree exceeds its byte bound")
			}
			fileIndexes = append(fileIndexes, index)
		default:
			return "", 0, 0, registryMismatch("runtime tree contains an unsupported file type")
		}
		entries[index] = entry
	}
	type hashResult struct {
		index  int
		digest string
		size   int64
		err    error
	}
	workerCount := min(runtime.NumCPU()*2, 16)
	jobs := make(chan int)
	results := make(chan hashResult, len(fileIndexes))
	var workers sync.WaitGroup
	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				digest, size, err := boundedRuntimeFileSHA256(filepath.Join(root, paths[index]), options.maxFileBytes)
				results <- hashResult{index: index, digest: digest, size: size, err: err}
			}
		}()
	}
	for _, index := range fileIndexes {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	close(results)
	for result := range results {
		if result.err != nil || result.size != entries[result.index].size {
			if result.err != nil {
				return "", 0, 0, result.err
			}
			return "", 0, 0, registryMismatch("runtime source size changed while hashing")
		}
		entries[result.index].digest = result.digest
	}
	digest := sha256.New()
	_, _ = io.WriteString(digest, runtimeTreeSchema+"\n")
	for _, entry := range entries {
		switch entry.kind {
		case "l":
			writeManifestFields(digest, "l", entry.relative, fmt.Sprintf("%04o", entry.mode), entry.target)
		case "d":
			writeManifestFields(digest, "d", entry.relative, fmt.Sprintf("%04o", entry.mode), "")
		case "f":
			writeManifestFields(digest, "f", entry.relative, fmt.Sprintf("%04o", entry.mode), strconv.FormatInt(entry.size, 10), entry.digest)
		}
	}
	return hex.EncodeToString(digest.Sum(nil)), len(entries), total, nil
}

func writeManifestFields(digest hash.Hash, fields ...string) {
	for index, field := range fields {
		if index > 0 {
			_, _ = digest.Write([]byte{0})
		}
		_, _ = io.WriteString(digest, field)
	}
	_, _ = digest.Write([]byte{'\n'})
}

func copyRuntimeTree(ctx context.Context, sourceRoot, targetRoot string, options hostRuntimeSnapshotOptions) error {
	count := 0
	var total int64
	return filepath.WalkDir(sourceRoot, func(source string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if source == sourceRoot {
			return nil
		}
		count++
		if count > options.maxEntries {
			return registryMismatch("runtime copy exceeds its entry bound")
		}
		relative, err := filepath.Rel(sourceRoot, source)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return registryMismatch("runtime copy entry escapes its root")
		}
		target := filepath.Join(targetRoot, relative)
		info, err := os.Lstat(source)
		if err != nil {
			return err
		}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			resolved, err := filepath.EvalSymlinks(source)
			if err != nil {
				return err
			}
			within, err := pathInside(sourceRoot, resolved)
			if err != nil || !within {
				return registryMismatch("runtime copy symlink escapes its root")
			}
			resolvedRelative, err := filepath.Rel(sourceRoot, resolved)
			if err != nil {
				return registryMismatch("resolve canonical runtime symlink target")
			}
			canonicalTarget, err := filepath.Rel(filepath.Dir(target), filepath.Join(targetRoot, resolvedRelative))
			if err != nil {
				return registryMismatch("construct canonical runtime symlink target")
			}
			return os.Symlink(canonicalTarget, target)
		case info.IsDir():
			return os.Mkdir(target, info.Mode().Perm())
		case info.Mode().IsRegular():
			total += info.Size()
			if info.Size() > options.maxFileBytes || total > options.maxTreeBytes {
				return registryMismatch("runtime copy exceeds its byte bound")
			}
			return copyRuntimeFileNoFollow(source, target, info.Mode().Perm(), options.maxFileBytes)
		default:
			return registryMismatch("runtime copy contains an unsupported file type")
		}
	})
}

func copyRuntimeFileNoFollow(source, target string, mode os.FileMode, maxBytes int64) error {
	sourceFile, before, err := openRuntimeFileNoFollow(source)
	if err != nil {
		return registryMismatch("open runtime copy source")
	}
	defer func() { _ = sourceFile.Close() }()
	if before.Size() < 0 || before.Size() > maxBytes {
		return registryMismatch("runtime copy source is oversized")
	}
	targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(targetFile, io.LimitReader(sourceFile, maxBytes+1))
	syncErr := targetFile.Sync()
	closeErr := targetFile.Close()
	after, statErr := sourceFile.Stat()
	if copyErr != nil || syncErr != nil || closeErr != nil || statErr != nil {
		return errors.Join(copyErr, syncErr, closeErr, statErr)
	}
	if written != before.Size() || written > maxBytes || !os.SameFile(before, after) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return registryMismatch("runtime copy source changed while reading")
	}
	return nil
}

func boundedRuntimeFileSHA256(path string, maxBytes int64) (string, int64, error) {
	file, before, err := openRuntimeFileNoFollow(path)
	if err != nil {
		return "", 0, registryMismatch("open bounded runtime source")
	}
	defer func() { _ = file.Close() }()
	if before.Size() < 0 || before.Size() > maxBytes {
		return "", 0, registryMismatch("runtime source exceeds its byte bound")
	}
	digest := sha256.New()
	read, err := io.Copy(digest, io.LimitReader(file, maxBytes+1))
	if err != nil {
		return "", 0, registryMismatch("hash bounded runtime source")
	}
	after, err := file.Stat()
	pathAfter, pathErr := os.Lstat(path)
	if err != nil || pathErr != nil || read != before.Size() || read > maxBytes || !os.SameFile(before, after) || !os.SameFile(after, pathAfter) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return "", 0, registryMismatch("runtime source changed while hashing")
	}
	return hex.EncodeToString(digest.Sum(nil)), read, nil
}

func readBoundedRuntimeFile(path string, maxBytes int64) ([]byte, os.FileInfo, error) {
	file, before, err := openRuntimeFileNoFollow(path)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = file.Close() }()
	if before.Size() < 0 || before.Size() > maxBytes {
		return nil, nil, registryMismatch("runtime source exceeds its byte bound")
	}
	reader := bufio.NewReader(io.LimitReader(file, maxBytes+1))
	raw, err := io.ReadAll(reader)
	if err != nil || int64(len(raw)) != before.Size() || int64(len(raw)) > maxBytes {
		return nil, nil, registryMismatch("read bounded runtime source")
	}
	after, err := file.Stat()
	pathAfter, pathErr := os.Lstat(path)
	if err != nil || pathErr != nil || !os.SameFile(before, after) || !os.SameFile(after, pathAfter) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return nil, nil, registryMismatch("runtime source changed while reading")
	}
	return raw, before, nil
}

func makeRuntimeTreeReadOnly(root string) error {
	var directories []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if info.IsDir() {
			directories = append(directories, path)
			return nil
		}
		mode := os.FileMode(0o444)
		if info.Mode().Perm()&0o111 != 0 {
			mode = 0o555
		}
		return os.Chmod(path, mode)
	})
	if err != nil {
		return err
	}
	sort.Slice(directories, func(i, j int) bool { return len(directories[i]) > len(directories[j]) })
	for _, directory := range directories {
		if err := os.Chmod(directory, 0o555); err != nil {
			return err
		}
	}
	return nil
}

func makeRuntimeTreeWritable(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		if info.IsDir() {
			return os.Chmod(path, 0o700)
		}
		return os.Chmod(path, 0o600)
	})
}

func jsonStrictDecode(raw []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return errors.New("trailing JSON")
	}
	return nil
}
