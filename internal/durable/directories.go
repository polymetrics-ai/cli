package durable

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func EnsureDirectoryTree(path, root string, mode os.FileMode) error {
	return ensureDirectoryTree(path, root, mode, SyncDirectory)
}

func ensureDirectoryTree(path, root string, mode os.FileMode, syncDirectory func(string) error) error {
	if path == "" || root == "" {
		return errors.New("durable directory path and root are required")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve durable directory root: %w", err)
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve durable directory path: %w", err)
	}
	relativePath, err := filepath.Rel(absoluteRoot, absolutePath)
	if err != nil {
		return fmt.Errorf("resolve durable directory relative path: %w", err)
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) {
		return errors.New("durable directory path is outside its root")
	}

	directories := directoryAncestors(absolutePath)
	for _, directory := range directories {
		if err := ensureDirectory(directory, mode); err != nil {
			return err
		}
	}
	for _, directory := range directories {
		if err := syncDirectory(directory); err != nil {
			return fmt.Errorf("sync durable directory %s: %w", directory, err)
		}
	}
	return nil
}

func directoryAncestors(path string) []string {
	directories := []string{}
	for current := path; ; current = filepath.Dir(current) {
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		directories = append(directories, current)
	}
	for left, right := 0, len(directories)-1; left < right; left, right = left+1, right-1 {
		directories[left], directories[right] = directories[right], directories[left]
	}
	return directories
}

func ensureDirectory(path string, mode os.FileMode) error {
	err := os.Mkdir(path, mode)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("create durable directory %s: %w", path, err)
	}
	return nil
}
