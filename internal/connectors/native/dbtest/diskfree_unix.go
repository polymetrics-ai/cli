//go:build !windows

package dbtest

import (
	"path/filepath"
	"syscall"
)

// diskFreeBytes reports the bytes available to an unprivileged process on the
// filesystem holding the working directory.
func diskFreeBytes() (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(filepath.Clean("."), &stat); err != nil {
		return 0, err
	}
	return uint64(stat.Bavail) * uint64(stat.Bsize), nil
}
