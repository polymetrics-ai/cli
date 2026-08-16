package warehouse

import (
	"errors"
	"fmt"
	"syscall"
)

// MinimumFastPathFreeBytes is the non-negotiable safety reserve for a
// transformed segment run. The check is made before every segment publish so
// a source stream stops cleanly rather than exhausting the workspace volume.
const MinimumFastPathFreeBytes int64 = 3 << 30

var ErrInsufficientFastPathDisk = errors.New("insufficient free disk for transformed transport")

// InsufficientFastPathDiskError is a typed, safe refusal. It carries only
// aggregate capacity numbers; it never exposes a warehouse path, source
// payload, connector configuration, or segment identity.
type InsufficientFastPathDiskError struct {
	Available int64
	Required  int64
}

func (e *InsufficientFastPathDiskError) Error() string {
	if e == nil {
		return ErrInsufficientFastPathDisk.Error()
	}
	return fmt.Sprintf("%s: available=%d required=%d", ErrInsufficientFastPathDisk, e.Available, e.Required)
}

func (e *InsufficientFastPathDiskError) Unwrap() error { return ErrInsufficientFastPathDisk }

var availableDiskBytes = filesystemAvailableBytes

// RequireMinimumFastPathFreeSpace checks the filesystem that owns the
// warehouse root before a Parquet segment is admitted. It has no connector
// knowledge and does not create a directory or publish an artifact.
func RequireMinimumFastPathFreeSpace(path string) error {
	available, err := availableDiskBytes(path)
	if err != nil {
		return err
	}
	if available < MinimumFastPathFreeBytes {
		return &InsufficientFastPathDiskError{Available: available, Required: MinimumFastPathFreeBytes}
	}
	return nil
}

func filesystemAvailableBytes(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	if stat.Bavail < 0 || stat.Bsize <= 0 {
		return 0, ErrInsufficientFastPathDisk
	}
	available := uint64(stat.Bavail) * uint64(stat.Bsize)
	if available > uint64(^uint64(0)>>1) {
		return int64(^uint64(0) >> 1), nil
	}
	return int64(available), nil
}
