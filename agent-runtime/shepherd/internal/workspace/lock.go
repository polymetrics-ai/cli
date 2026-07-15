package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// RepositoryLock serializes Shepherd controllers by canonical Git common
// directory, independent of their configured state directory. The kernel
// releases the advisory lock if the controller exits unexpectedly.
type RepositoryLock struct {
	file *os.File
	path string
}

func (m *Manager) TryAcquireRepositoryLock() (*RepositoryLock, error) {
	commonRaw, err := git(context.Background(), m.RepoRoot, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return nil, err
	}
	common := filepath.Clean(strings.TrimSpace(string(commonRaw)))
	if !filepath.IsAbs(common) {
		return nil, errors.New("canonical Git common directory must be absolute")
	}
	path := filepath.Join(common, "shepherd-controller.lock")
	fd, err := syscall.Open(path, syscall.O_CREAT|syscall.O_RDWR|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0o600)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		_ = syscall.Close(fd)
		return nil, errors.New("open repository lock")
	}
	if err := syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		return nil, errors.New("another Shepherd controller holds the repository lock")
	}
	lock := &RepositoryLock{file: file, path: path}
	if err := lock.Check(); err != nil {
		_ = lock.Close()
		return nil, err
	}
	return lock, nil
}

// Check proves the locked descriptor is still the inode named by the canonical
// lock path. A same-UID worker cannot silently bypass the lock by replacing it.
func (l *RepositoryLock) Check() error {
	if l == nil || l.file == nil {
		return errors.New("repository lock is not held")
	}
	lockedInfo, err := l.file.Stat()
	if err != nil {
		return err
	}
	pathInfo, err := os.Lstat(l.path)
	if err != nil || pathInfo.Mode()&os.ModeSymlink != 0 || !os.SameFile(lockedInfo, pathInfo) {
		return errors.New("repository lock path was replaced")
	}
	return nil
}

func (l *RepositoryLock) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	unlockErr := syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	closeErr := l.file.Close()
	l.file = nil
	return errors.Join(unlockErr, closeErr)
}
