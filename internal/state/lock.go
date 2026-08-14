package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type FileLock struct {
	Path string
}

// Lock creates Path with O_EXCL and returns an unlock function that removes it.
func (l FileLock) Lock() (func() error, error) {
	if l.Path == "" {
		return nil, errors.New("lock path is required")
	}
	if err := os.MkdirAll(filepath.Dir(l.Path), 0o700); err != nil {
		return nil, fmt.Errorf("create lock directory: %w", err)
	}
	file, err := os.OpenFile(l.Path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create lock file: %w", err)
	}
	if _, err := fmt.Fprintf(file, "%d\n", os.Getpid()); err != nil {
		_ = file.Close()
		_ = os.Remove(l.Path)
		return nil, fmt.Errorf("write lock file: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(l.Path)
		return nil, fmt.Errorf("close lock file: %w", err)
	}

	released := false
	return func() error {
		if released {
			return nil
		}
		released = true
		if err := os.Remove(l.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove lock file: %w", err)
		}
		return nil
	}, nil
}
