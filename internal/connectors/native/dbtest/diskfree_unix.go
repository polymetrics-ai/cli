//go:build !windows

package dbtest

import (
	"path/filepath"
	"syscall"
)

// diskFreeAt reports bytes available to an unprivileged process at path.
func diskFreeAt(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(filepath.Clean(path), &stat); err != nil {
		return 0, err
	}
	return uint64(stat.Bavail) * uint64(stat.Bsize), nil
}
