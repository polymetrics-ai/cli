//go:build unix

package state

import (
	"errors"
	"os"
	"syscall"
)

func acquireFileLock(file *os.File) error {
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return os.ErrExist
		}
		return err
	}
	return nil
}

func releaseFileLock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
