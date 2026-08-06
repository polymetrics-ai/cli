//go:build !unix && !windows

package durable

import "errors"

func SyncDirectory(string) error {
	return errors.New("directory sync is unsupported on this platform")
}

func shouldSyncFilesystemRoot() bool { return false }
