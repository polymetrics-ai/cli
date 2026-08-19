//go:build windows

package durability

import "syscall"

// SyncDirectory flushes directory metadata after an atomic replacement so a
// successful rename is durable across a crash.
func SyncDirectory(path string) error {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	handle, err := syscall.CreateFile(
		name,
		syscall.GENERIC_WRITE,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return err
	}
	defer func() { _ = syscall.CloseHandle(handle) }()
	return syscall.FlushFileBuffers(handle)
}
