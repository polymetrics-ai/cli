//go:build !unix && !windows

package state

import (
	"errors"
	"os"
)

func acquireFileLock(*os.File) error {
	return errors.New("interprocess file locking is unsupported on this platform")
}

func releaseFileLock(*os.File) error { return nil }
