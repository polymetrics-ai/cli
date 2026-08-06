//go:build windows

package durable

import (
	"fmt"
	"syscall"
)

func SyncDirectory(path string) error {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("encode directory path: %w", err)
	}
	handle, err := syscall.CreateFile(
		name,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return fmt.Errorf("open directory: %w", err)
	}
	syncErr := syscall.FlushFileBuffers(handle)
	closeErr := syscall.CloseHandle(handle)
	if syncErr != nil {
		return fmt.Errorf("sync directory: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close directory: %w", closeErr)
	}
	return nil
}

func shouldSyncFilesystemRoot() bool { return false }
