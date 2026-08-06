package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"polymetrics.ai/internal/durable"
)

type FileLock struct {
	Path          string
	DirectoryRoot string
}

// Lock acquires an exclusive interprocess lock for Path.
func (l FileLock) Lock() (func() error, error) {
	if l.Path == "" {
		return nil, errors.New("lock path is required")
	}
	dir := filepath.Dir(l.Path)
	root := l.DirectoryRoot
	if root == "" {
		root = dir
	}
	if err := durable.EnsureDirectoryTree(dir, root, 0o700); err != nil {
		return nil, fmt.Errorf("create lock directory: %w", err)
	}
	file, err := os.OpenFile(l.Path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create lock file: %w", err)
	}
	if err := acquireFileLock(file); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("acquire lock file: %w", err)
	}

	var releaseOnce sync.Once
	var releaseErr error
	return func() error {
		releaseOnce.Do(func() {
			releaseErr = releaseFileLock(file)
			if closeErr := file.Close(); releaseErr == nil && closeErr != nil {
				releaseErr = fmt.Errorf("close lock file: %w", closeErr)
			}
		})
		return releaseErr
	}, nil
}
