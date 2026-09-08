//go:build !windows

package durability

import "os"

// SyncDirectory flushes directory metadata after an atomic replacement so a
// successful rename is durable across a crash.
func SyncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = directory.Close() }()
	return directory.Sync()
}
