package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

type FileLock struct {
	Path string
}

type DirectoryLock struct {
	Path string
}

type directoryLockEntry struct {
	mu   sync.Mutex
	refs int
}

var directoryLockEntries = struct {
	sync.Mutex
	entries map[string]*directoryLockEntry
}{entries: map[string]*directoryLockEntry{}}

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

func (l DirectoryLock) Lock() (func() error, error) {
	if l.Path == "" {
		return nil, errors.New("lock directory is required")
	}
	path := filepath.Clean(l.Path)
	entry := acquireDirectoryLock(path)
	entry.mu.Lock()

	dir, err := os.Open(path)
	if err != nil {
		releaseDirectoryLock(path, entry)
		return nil, fmt.Errorf("open lock directory: %w", err)
	}
	if err := syscall.Flock(int(dir.Fd()), syscall.LOCK_EX); err != nil {
		_ = dir.Close()
		releaseDirectoryLock(path, entry)
		return nil, fmt.Errorf("lock directory: %w", err)
	}

	released := false
	return func() error {
		if released {
			return nil
		}
		released = true
		unlockErr := syscall.Flock(int(dir.Fd()), syscall.LOCK_UN)
		closeErr := dir.Close()
		releaseDirectoryLock(path, entry)
		if unlockErr != nil {
			return fmt.Errorf("unlock directory: %w", unlockErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close lock directory: %w", closeErr)
		}
		return nil
	}, nil
}

func acquireDirectoryLock(path string) *directoryLockEntry {
	directoryLockEntries.Lock()
	defer directoryLockEntries.Unlock()
	entry := directoryLockEntries.entries[path]
	if entry == nil {
		entry = &directoryLockEntry{}
		directoryLockEntries.entries[path] = entry
	}
	entry.refs++
	return entry
}

func releaseDirectoryLock(path string, entry *directoryLockEntry) {
	entry.mu.Unlock()
	directoryLockEntries.Lock()
	defer directoryLockEntries.Unlock()
	entry.refs--
	if entry.refs == 0 {
		delete(directoryLockEntries.entries, path)
	}
}
