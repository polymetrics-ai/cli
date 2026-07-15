//go:build !unix

package gsd

import (
	"fmt"
	"os"
)

func runtimePathOwnedByCurrentUser(os.FileInfo) bool {
	return false
}

func openRuntimeFileNoFollow(path string) (*os.File, os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("runtime source must be a regular non-symlink file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		_ = file.Close()
		return nil, nil, fmt.Errorf("runtime source identity changed while opening")
	}
	return file, opened, nil
}
