//go:build darwin || linux

package state

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// AdvisoryFileLock holds a kernel-managed exclusive lock on Path. The file is
// deliberately retained between calls: the kernel releases ownership when a
// process exits, including SIGKILL, so a stale marker cannot wedge recovery.
type AdvisoryFileLock struct {
	Path string
}

func (l AdvisoryFileLock) Lock() (func() error, error) {
	if l.Path == "" {
		return nil, errors.New("lock path is required")
	}
	if err := os.MkdirAll(filepath.Dir(l.Path), 0o700); err != nil {
		return nil, fmt.Errorf("create lock directory: %w", err)
	}
	file, err := os.OpenFile(l.Path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open advisory lock file: %w", err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("acquire advisory lock: %w", err)
	}

	released := false
	return func() error {
		if released {
			return nil
		}
		released = true
		var unlockErr error
		if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
			unlockErr = fmt.Errorf("release advisory lock: %w", err)
		}
		if err := file.Close(); unlockErr == nil && err != nil {
			unlockErr = fmt.Errorf("close advisory lock file: %w", err)
		}
		return unlockErr
	}, nil
}
