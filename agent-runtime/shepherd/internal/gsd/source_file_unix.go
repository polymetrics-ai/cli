//go:build unix

package gsd

import (
	"fmt"
	"os"
	"syscall"
)

func runtimePathOwnedByCurrentUser(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Uid == uint32(os.Geteuid())
}

func openRuntimeFileNoFollow(path string) (*os.File, os.FileInfo, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("open no-follow runtime file: %w", err)
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		_ = syscall.Close(fd)
		return nil, nil, fmt.Errorf("open no-follow runtime file: invalid descriptor")
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, nil, fmt.Errorf("stat no-follow runtime file: %w", err)
	}
	if !info.Mode().IsRegular() || !runtimePathOwnedByCurrentUser(info) {
		_ = file.Close()
		return nil, nil, fmt.Errorf("runtime source must be an owned regular file")
	}
	return file, info, nil
}
