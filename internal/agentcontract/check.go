package agentcontract

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	script := filepath.Join(absoluteRoot, "scripts", "gsd")
	for _, command := range commands {
		invocation := exec.CommandContext(ctx, script, "sources", command)
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

	for _, target := range contract.Projections {
		path, err := projectionPath(target.Path)
		if err != nil {
			return err
		}
		content, err := projectionRoot.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) && !target.Required {
				continue
			}
			return fmt.Errorf("read %s projection %s: %w", target.Harness, target.Path, err)
		}
		expected, err := RenderBlock(contract, target.Role)
		if err != nil {
			return err
		}
		actual, _, _, err := extractProjectionBlock(content)
		if err != nil {
			return fmt.Errorf("check projection %s: %w", target.Path, err)
		}
		if err := CheckProjection(expected, actual); err != nil {
			return fmt.Errorf("check projection %s: %w", target.Path, err)
		}
	}
	return nil
}

func CheckProjection(want, got []byte) error {
	if !bytes.Equal(want, got) {
		return fmt.Errorf("generated block diverges from canonical source; run go run ./cmd/agentcontractgen sync")
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
		content, err := projectionRoot.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) && !target.Required {
				continue
			}
			return updated, fmt.Errorf("read %s projection %s: %w", target.Harness, target.Path, err)
		}
		expected, start, end, err := replacementBlock(content, contract, target.Role)
		if err != nil {
			return updated, fmt.Errorf("sync projection %s: %w", target.Path, err)
		}
		if bytes.Equal(content[start:end], expected) {
			continue
		}
		next := make([]byte, 0, len(content)-end+start+len(expected))
		next = append(next, content[:start]...)
		next = append(next, expected...)
		next = append(next, content[end:]...)
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

func replacementBlock(content []byte, contract *Contract, role string) ([]byte, int, int, error) {
	_, start, end, err := extractProjectionBlock(content)
	if err != nil {
		return nil, 0, 0, err
	}
	expected, err := RenderBlock(contract, role)
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
