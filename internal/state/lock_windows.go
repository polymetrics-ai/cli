//go:build windows

package state

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	lockFileFailImmediately = 0x00000001
	lockFileExclusive       = 0x00000002
)

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	lockFileExProc    = kernel32.NewProc("LockFileEx")
	unlockFileExProc  = kernel32.NewProc("UnlockFileEx")
	lockViolationCode = syscall.Errno(33)
)

func acquireFileLock(file *os.File) error {
	var overlapped syscall.Overlapped
	result, _, err := lockFileExProc.Call(
		file.Fd(),
		uintptr(lockFileFailImmediately|lockFileExclusive),
		0,
		uintptr(^uint32(0)),
		uintptr(^uint32(0)),
		uintptr(unsafe.Pointer(&overlapped)),
	)
	if result != 0 {
		return nil
	}
	if errors.Is(err, lockViolationCode) {
		return os.ErrExist
	}
	return fmt.Errorf("lock file: %w", err)
}

func releaseFileLock(file *os.File) error {
	var overlapped syscall.Overlapped
	result, _, err := unlockFileExProc.Call(
		file.Fd(),
		0,
		uintptr(^uint32(0)),
		uintptr(^uint32(0)),
		uintptr(unsafe.Pointer(&overlapped)),
	)
	if result != 0 {
		return nil
	}
	return fmt.Errorf("unlock file: %w", err)
}
