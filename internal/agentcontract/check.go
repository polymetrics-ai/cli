package agentcontract

import (
	"bytes"
	"context"
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
	if err := CheckGSDCommands(ctx, filepath.Join(root, "scripts", "gsd"), contract.GSD.Commands); err != nil {
		return err
	}
	return CheckProjections(root, contract)
}

// CheckGSDCommands verifies command names through the repository's real GSD adapter.
func CheckGSDCommands(ctx context.Context, script string, commands []string) error {
	for _, command := range commands {
		output, err := exec.CommandContext(ctx, script, "sources", command).CombinedOutput()
		if err != nil {
			return fmt.Errorf("GSD command %q does not resolve: %s: %w", command, strings.TrimSpace(string(output)), err)
		}
	}
	return nil
}

func CheckProjections(root string, contract *Contract) error {
	for _, target := range contract.Projections {
		path := filepath.Join(root, filepath.FromSlash(target.Path))
		content, err := os.ReadFile(path)
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

func SyncProjections(root string, contract *Contract) (int, error) {
	updated := 0
	for _, target := range contract.Projections {
		path := filepath.Join(root, filepath.FromSlash(target.Path))
		content, err := os.ReadFile(path)
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
		info, err := os.Stat(path)
		if err != nil {
			return updated, fmt.Errorf("stat projection %s: %w", target.Path, err)
		}
		if err := writeAtomic(path, next, info.Mode().Perm()); err != nil {
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

func writeAtomic(path string, content []byte, mode os.FileMode) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".agentcontractgen-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	cleanup := func() { _ = os.Remove(temporaryPath) }
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
	return os.Rename(temporaryPath, path)
}
