//go:build unix

package durable

import (
	"fmt"
	"os"
)

func SyncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open directory: %w", err)
	}
	defer func() { _ = dir.Close() }()
	if err := dir.Sync(); err != nil {
		return fmt.Errorf("sync directory: %w", err)
	}
	return nil
}
