package durable

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func EnsureDirectoryTree(path, root string, mode os.FileMode) error {
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

	directories := []string{absoluteRoot}
	current := absoluteRoot
	if relativePath != "." {
		for _, component := range strings.Split(relativePath, string(os.PathSeparator)) {
			if component == "" || component == "." {
				continue
			}
			current = filepath.Join(current, component)
			directories = append(directories, current)
		}
	}
	for _, directory := range directories {
		if err := ensureDirectory(directory, mode); err != nil {
			return err
		}
	}
	for _, directory := range directories {
		if filepath.Dir(directory) == directory {
			continue
		}
		if err := SyncDirectory(directory); err != nil {
			return fmt.Errorf("sync durable directory %s: %w", directory, err)
		}
	}
	return nil
}

func ensureDirectory(path string, mode os.FileMode) error {
	parent := filepath.Dir(path)
	err := os.Mkdir(path, mode)
	if err != nil && errors.Is(err, os.ErrNotExist) && parent != path {
		if err := ensureDirectory(parent, mode); err != nil {
			return err
		}
		err = os.Mkdir(path, mode)
	}
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("create durable directory %s: %w", path, err)
	}
	if parent != path {
		if err := SyncDirectory(parent); err != nil {
			return fmt.Errorf("sync durable directory parent %s: %w", parent, err)
		}
		if err := SyncDirectory(path); err != nil {
			return fmt.Errorf("sync durable directory %s: %w", path, err)
		}
	}
	return nil
}
