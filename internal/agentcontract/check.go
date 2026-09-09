package agentcontract

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"slices"
	"strings"
)

func CheckRoot(ctx context.Context, root string) error {
	contract, err := Load(filepath.Join(root, SourcePath))
	if err != nil {
		return err
	}
	if err := CheckGSDCommands(ctx, root, contract.GSD.Commands); err != nil {
		return err
	}
	return CheckProjections(root, contract)
}

// CheckGSDCommands verifies command names through the repository's real GSD adapter.
func CheckGSDCommands(ctx context.Context, root string, commands []string) error {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve GSD repository root: %w", err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		return fmt.Errorf("resolve Node executable for GSD adapter: %w", err)
	}
	script := filepath.Join(absoluteRoot, "scripts", "gsd")
	for _, command := range commands {
		invocation := exec.CommandContext(ctx, node, script, "sources", command)
		invocation.Dir = absoluteRoot
		output, err := invocation.CombinedOutput()
		if err != nil {
			return fmt.Errorf("GSD command %q does not resolve: %s: %w", command, strings.TrimSpace(string(output)), err)
		}
	}
	return nil
}

func CheckProjections(root string, contract *Contract) (returnErr error) {
	projectionRoot, err := openProjectionRoot(root, contract)
	if err != nil {
		return err
	}
	defer func() {
		if err := projectionRoot.Close(); err != nil && returnErr == nil {
			returnErr = fmt.Errorf("close projection root: %w", err)
		}
	}()
	if err := checkClaudeAgentInventory(projectionRoot, contract); err != nil {
		return err
	}

	for _, target := range contract.Projections {
		path, err := projectionPath(target.Path)
		if err != nil {
			return err
		}
		content, err := readProjection(projectionRoot, target, path)
		if err != nil {
			if os.IsNotExist(err) && !target.Required {
				continue
			}
			return fmt.Errorf("read %s projection %s: %w", target.Harness, target.Path, err)
		}
		expected, err := RenderProjection(contract, target)
		if err != nil {
			return err
		}
		actual := content
		switch target.RenderMode {
		case "markdown_block":
			actual, _, _, err = extractProjectionBlock(content)
			if err != nil {
				return fmt.Errorf("check projection %s: %w", target.Path, err)
			}
		case claudeMarkdownYAMLFrontmatter:
			expected = normalizeClaudeProjection(expected)
			actual = normalizeClaudeProjection(content)
			policy, ok := contract.ProjectionFor(target.Harness)
			if !ok {
				return fmt.Errorf("check projection %s: canonical %s policy is missing", target.Path, target.Harness)
			}
			frontmatter, err := parseClaudeFrontmatter(actual)
			if err != nil {
				return fmt.Errorf("check projection %s: %w", target.Path, err)
			}
			if err := validateClaudeFrontmatter(frontmatter, target, policy); err != nil {
				return fmt.Errorf("check projection %s: %w", target.Path, err)
			}
		case opencodeMarkdownYAMLFrontmatter:
			expected = normalizeClaudeProjection(expected)
			actual = normalizeClaudeProjection(content)
			frontmatter, err := parseOpenCodeFrontmatter(actual)
			if err != nil {
				return fmt.Errorf("check projection %s: %w", target.Path, err)
			}
			if err := validateOpenCodeFrontmatter(frontmatter, target, contract.OpenCode); err != nil {
				return fmt.Errorf("check projection %s: %w", target.Path, err)
			}
		case "full":
			if err := CheckProjection(expected, content); err != nil {
				return fmt.Errorf("check projection %s: %w", target.Path, err)
			}
			continue
		}
		if err := CheckProjection(expected, actual); err != nil {
			return fmt.Errorf("check projection %s: %w", target.Path, err)
		}
	}
	return nil
}

func checkClaudeAgentInventory(projectionRoot *os.Root, contract *Contract) error {
	const agentDirectory = ".claude/agents"
	expected := make(map[string]string)
	expectedPaths := make([]string, 0, 2)
	for _, target := range contract.Projections {
		if target.Harness != "claude" {
			continue
		}
		expected[target.Path] = target.Role
		expectedPaths = append(expectedPaths, target.Path)
	}
	slices.Sort(expectedPaths)

	for _, directory := range []string{".claude", agentDirectory} {
		info, err := projectionRoot.Lstat(filepath.FromSlash(directory))
		if err != nil {
			return fmt.Errorf("inspect Claude project agent inventory: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("inspect Claude project agent inventory: %s must be a directory, not a symlink or special file", directory)
		}
	}

	found := make(map[string]bool, len(expected))
	seenNames := make(map[string]string, len(expected))
	unexpected := make([]string, 0)
	walkErr := fs.WalkDir(projectionRoot.FS(), agentDirectory, func(agentPath string, entry fs.DirEntry, visitErr error) error {
		if visitErr != nil {
			return visitErr
		}
		if agentPath == agentDirectory {
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("claude project agent inventory contains symlink %s", agentPath)
		}
		if entry.IsDir() || !strings.EqualFold(pathpkg.Ext(agentPath), ".md") {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("claude project agent definition %s is not a regular file", agentPath)
		}
		content, err := fs.ReadFile(projectionRoot.FS(), agentPath)
		if err != nil {
			return fmt.Errorf("read claude project agent definition %s: %w", agentPath, err)
		}
		frontmatter, err := parseClaudeFrontmatter(content)
		if err != nil {
			return fmt.Errorf("parse claude project agent definition %s: %w", agentPath, err)
		}
		if previous, ok := seenNames[frontmatter.Name]; ok {
			return fmt.Errorf("duplicate claude project agent name %q at %s and %s", frontmatter.Name, previous, agentPath)
		}
		seenNames[frontmatter.Name] = agentPath
		role, ok := expected[agentPath]
		if !ok {
			unexpected = append(unexpected, agentPath)
			return nil
		}
		if frontmatter.Name != role {
			return fmt.Errorf("claude project agent definition %s declares name %q, want %q", agentPath, frontmatter.Name, role)
		}
		found[agentPath] = true
		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("inspect Claude project agent inventory: %w", walkErr)
	}
	if len(unexpected) != 0 {
		slices.Sort(unexpected)
		return fmt.Errorf("inspect Claude project agent inventory: unexpected definitions %s; only %s are permitted", strings.Join(unexpected, ", "), strings.Join(expectedPaths, ", "))
	}
	for _, expectedPath := range expectedPaths {
		if !found[expectedPath] {
			return fmt.Errorf("inspect Claude project agent inventory: missing required definition %s", expectedPath)
		}
	}
	return nil
}

func CheckProjection(want, got []byte) error {
	if !bytes.Equal(want, got) {
		return fmt.Errorf("generated projection diverges from canonical source; run go run ./cmd/agentcontractgen sync")
	}
	return nil
}

func SyncProjections(root string, contract *Contract) (updated int, returnErr error) {
	projectionRoot, err := openProjectionRoot(root, contract)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := projectionRoot.Close(); err != nil && returnErr == nil {
			returnErr = fmt.Errorf("close projection root: %w", err)
		}
	}()

	for _, target := range contract.Projections {
		path, err := projectionPath(target.Path)
		if err != nil {
			return updated, err
		}
		expected, err := RenderProjection(contract, target)
		if err != nil {
			return updated, err
		}
		content, err := readProjection(projectionRoot, target, path)
		if err != nil {
			if !os.IsNotExist(err) {
				return updated, fmt.Errorf("read %s projection %s: %w", target.Harness, target.Path, err)
			}
			if !target.Required {
				continue
			}
			if target.RenderMode == claudeMarkdownYAMLFrontmatter || target.RenderMode == opencodeMarkdownYAMLFrontmatter {
				expected = normalizeClaudeProjection(expected)
			}
			if err := ensureProjectionDirectory(projectionRoot, filepath.Dir(path)); err != nil {
				return updated, fmt.Errorf("create projection directory for %s: %w", target.Path, err)
			}
			if err := writeAtomic(projectionRoot, path, expected, 0o644); err != nil {
				return updated, fmt.Errorf("write projection %s: %w", target.Path, err)
			}
			updated++
			continue
		}

		var next []byte
		switch target.RenderMode {
		case "markdown_block":
			block, start, end, err := replacementBlock(content, contract, target)
			if err != nil {
				return updated, fmt.Errorf("sync projection %s: %w", target.Path, err)
			}
			if bytes.Equal(content[start:end], block) {
				continue
			}
			next = append(next, content[:start]...)
			next = append(next, block...)
			next = append(next, content[end:]...)
		case claudeMarkdownYAMLFrontmatter:
			expected = normalizeClaudeProjection(expected)
			if bytes.Equal(normalizeClaudeProjection(content), expected) {
				continue
			}
			next = expected
		case opencodeMarkdownYAMLFrontmatter:
			expected = normalizeClaudeProjection(expected)
			if bytes.Equal(normalizeClaudeProjection(content), expected) {
				continue
			}
			next = expected
		default:
			if bytes.Equal(content, expected) {
				continue
			}
			next = expected
		}
		info, err := projectionRoot.Stat(path)
		if err != nil {
			return updated, fmt.Errorf("stat projection %s: %w", target.Path, err)
		}
		if err := writeAtomic(projectionRoot, path, next, info.Mode().Perm()); err != nil {
			return updated, fmt.Errorf("write projection %s: %w", target.Path, err)
		}
		updated++
	}
	return updated, nil
}

func replacementBlock(content []byte, contract *Contract, target ProjectionTarget) ([]byte, int, int, error) {
	_, start, end, err := extractProjectionBlock(content)
	if err != nil {
		return nil, 0, 0, err
	}
	expected, err := RenderProjection(contract, target)
	if err != nil {
		return nil, 0, 0, err
	}
	return expected, start, end, nil
}

func extractProjectionBlock(content []byte) ([]byte, int, int, error) {
	start := bytes.Index(content, []byte(beginMarkerPrefix))
	if start < 0 {
		return nil, 0, 0, fmt.Errorf("generated block start marker is missing")
	}
	if bytes.Contains(content[start+len(beginMarkerPrefix):], []byte(beginMarkerPrefix)) {
		return nil, 0, 0, fmt.Errorf("multiple generated block start markers found")
	}
	relativeEnd := bytes.Index(content[start:], []byte(endMarker))
	if relativeEnd < 0 {
		return nil, 0, 0, fmt.Errorf("generated block end marker is missing")
	}
	end := start + relativeEnd + len(endMarker)
	if end < len(content) && content[end] == '\n' {
		end++
	}
	return content[start:end], start, end, nil
}

func openProjectionRoot(root string, contract *Contract) (*os.Root, error) {
	if err := contract.Validate(); err != nil {
		return nil, err
	}
	projectionRoot, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("open projection root: %w", err)
	}
	return projectionRoot, nil
}

func projectionPath(path string) (string, error) {
	localPath := filepath.FromSlash(path)
	if !filepath.IsLocal(localPath) || filepath.Clean(localPath) != localPath {
		return "", fmt.Errorf("canonical contract: projection path %q is not local", path)
	}
	return localPath, nil
}

func ensureProjectionDirectory(root *os.Root, directory string) error {
	if directory == "." {
		return nil
	}
	current := ""
	for _, component := range strings.Split(filepath.ToSlash(directory), "/") {
		if component == "" || component == "." || component == ".." {
			return fmt.Errorf("invalid projection directory %q", directory)
		}
		current = filepath.Join(current, component)
		if err := root.Mkdir(current, 0o755); err != nil && !os.IsExist(err) {
			return err
		}
	}
	return nil
}

func readProjection(root *os.Root, target ProjectionTarget, path string) ([]byte, error) {
	if err := validateProjectionReadPath(root, path, target.RenderMode == "full" || target.RenderMode == opencodeMarkdownYAMLFrontmatter); err != nil {
		return nil, err
	}
	return root.ReadFile(path)
}

func validateProjectionReadPath(root *os.Root, path string, requireRegularFile bool) error {
	components := strings.Split(path, string(filepath.Separator))
	current := ""
	for index, component := range components {
		current = filepath.Join(current, component)
		info, err := root.Lstat(current)
		if err != nil {
			return err
		}
		isTarget := index == len(components)-1
		if !isTarget {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("projection ancestor %s is a symbolic link", current)
			}
			if !info.IsDir() {
				return fmt.Errorf("projection ancestor %s is not a directory", current)
			}
			continue
		}
		if requireRegularFile {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("full projection %s is a symbolic link", current)
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("full projection %s is not a regular file", current)
			}
		}
	}
	return nil
}

func writeAtomic(root *os.Root, path string, content []byte, mode os.FileMode) error {
	temporary, temporaryPath, err := createRootTemp(root, filepath.Dir(path))
	if err != nil {
		return err
	}
	cleanup := func() { _ = root.Remove(temporaryPath) }
	defer cleanup()
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return root.Rename(temporaryPath, path)
}

func createRootTemp(root *os.Root, directory string) (*os.File, string, error) {
	for range 10 {
		var random [16]byte
		if _, err := rand.Read(random[:]); err != nil {
			return nil, "", fmt.Errorf("generate temporary projection name: %w", err)
		}
		path := filepath.Join(directory, ".agentcontractgen-"+hex.EncodeToString(random[:]))
		file, err := root.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			return file, path, nil
		}
		if !os.IsExist(err) {
			return nil, "", err
		}
	}
	return nil, "", fmt.Errorf("create temporary projection: name collision limit reached")
}
